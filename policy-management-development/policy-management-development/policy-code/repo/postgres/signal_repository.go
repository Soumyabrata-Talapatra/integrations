package repo

import (
	"context"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	config "gitlab.cept.gov.in/it-2.0-common/api-config"
	dblib "gitlab.cept.gov.in/it-2.0-common/n-api-db"

	"policy-management/core/domain"
)

// SignalRepository handles all data access for signal deduplication, audit logging,
// and policy event publishing.
//
// Tables managed:
//   - policy_mgmt.processed_signal_registry — dedup table (90-day TTL)
//   - policy_mgmt.policy_signal_log         — full audit trail (partitioned by received_at)
//   - policy_mgmt.policy_event              — outbound event log (partitioned by published_at)
//
// Constraint C7: All SQL uses policy_mgmt. schema prefix.
// [FR-PM-004, §9.1 signal deduplication, §9.5 audit trail]
type SignalRepository struct {
	db  *dblib.DB
	cfg *config.Config
}

// NewSignalRepository constructs a SignalRepository.
func NewSignalRepository(db *dblib.DB, cfg *config.Config) *SignalRepository {
	return &SignalRepository{db: db, cfg: cfg}
}

// table name constants (C7: policy_mgmt. prefix)
const (
	signalRegistryTable = "policy_mgmt.processed_signal_registry"
	signalLogTable      = "policy_mgmt.policy_signal_log"
	policyEventTable    = "policy_mgmt.policy_event"
)

// ─────────────────────────────────────────────────────────────────────────────
// CheckDuplicateSignal — Signal Dedup (idempotency before processing)
// ─────────────────────────────────────────────────────────────────────────────

// CheckDuplicateSignal returns true if the (requestID, signalType) pair has already
// been processed (i.e., it exists in processed_signal_registry).
// Called by PolicyLifecycleWorkflow at the top of each signal handler.
// requestID is the UUID string from the caller (VARCHAR(100)), NOT the BIGINT service_request.request_id.
// [§9.1 signal deduplication, FR-PM-004]
func (r *SignalRepository) CheckDuplicateSignal(ctx context.Context, requestID, signalType string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	// UNIQUE constraint: (request_id, signal_type) — single row if exists.
	query := dblib.Psql.Select("COUNT(*) AS count").
		From(signalRegistryTable).
		Where(sq.Eq{
			"request_id":  requestID,
			"signal_type": signalType,
		})

	type countRow struct{ Count int `db:"count"` }
	result, err := dblib.SelectOne(ctx, r.db, query, pgx.RowToStructByNameLax[countRow])
	if err != nil {
		return false, fmt.Errorf("CheckDuplicateSignal requestID=%s type=%s: %w", requestID, signalType, err)
	}
	return result.Count > 0, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// RegisterSignal — Insert into processed_signal_registry (mark as processed)
// ─────────────────────────────────────────────────────────────────────────────

// RegisterSignal inserts a row into processed_signal_registry to mark this
// (requestID, signalType) as processed. Entries expire after signal_dedup_ttl_days (90 days).
// ⚠️ request_id here is VARCHAR(100) — UUID string from caller.
//
//	(Different from service_request.request_id which is BIGINT)
//
// [§9.1 signal deduplication]
func (r *SignalRepository) RegisterSignal(ctx context.Context, reg *domain.ProcessedSignalRegistry) error {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Insert(signalRegistryTable).
		Columns("request_id", "signal_type", "policy_id", "received_at", "expires_at").
		Values(reg.RequestID, reg.SignalType, reg.PolicyID, reg.ReceivedAt, reg.ExpiresAt).
		// Idempotent: ON CONFLICT DO NOTHING protects against activity retries.
		Suffix("ON CONFLICT (request_id, signal_type) DO NOTHING RETURNING id")

	type idRow struct{ ID int64 `db:"id"` }
	// ON CONFLICT DO NOTHING returns empty slice — treat as success (idempotent).
	if _, err := dblib.InsertReturning(ctx, r.db, query, pgx.RowToStructByNameLax[idRow]); err != nil {
		return fmt.Errorf("RegisterSignal requestID=%s type=%s: %w", reg.RequestID, reg.SignalType, err)
	}
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// LogSignal — Insert into policy_signal_log (audit trail)
// ─────────────────────────────────────────────────────────────────────────────

// LogSignal inserts a single audit record into policy_signal_log.
// ⚠️ received_at is the partition key for policy_signal_log — must be supplied.
// Every signal (PROCESSED, REJECTED, DUPLICATE, FAILED) is logged for compliance.
// [§9.5 audit trail, compliance retention 3 years]
func (r *SignalRepository) LogSignal(ctx context.Context, log *domain.PolicySignalLog) error {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Insert(signalLogTable).
		Columns(
			"policy_id", "signal_channel", "signal_payload",
			"source_service", "source_workflow_id", "request_id",
			"received_at", "processed_at", "status", "rejection_reason",
			"state_before", "state_after",
		).
		Values(
			log.PolicyID, log.SignalChannel, log.SignalPayload,
			log.SourceService, log.SourceWorkflowID, log.RequestID,
			log.ReceivedAt, log.ProcessedAt, log.Status, log.RejectionReason,
			log.StateBefore, log.StateAfter,
		).
		// ⚠️ received_at must be supplied — it is the partition key.
		Suffix("RETURNING id")

	type idRow struct{ ID int64 `db:"id"` }
	if _, err := dblib.InsertReturning(ctx, r.db, query, pgx.RowToStructByNameLax[idRow]); err != nil {
		return fmt.Errorf("LogSignal policyID=%d channel=%s: %w", log.PolicyID, log.SignalChannel, err)
	}
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// LogAndRegisterSignal — pgx.Batch: INSERT log + INSERT registry atomically
// Called by LogSignalReceivedActivity (§21.1)
// ─────────────────────────────────────────────────────────────────────────────

// LogAndRegisterSignal is the primary method called by LogSignalReceivedActivity.
// It atomically:
//  1. INSERT INTO policy_signal_log  (audit trail)
//  2. INSERT INTO processed_signal_registry (dedup marker)
//
// Uses pgx.Batch for a single network round-trip.
// ⚠️ log.ReceivedAt is the partition key for policy_signal_log — always set.
// ⚠️ reg.RequestID is VARCHAR(100) UUID — not service_request.request_id (BIGINT).
// [§21.1 LogSignalReceivedActivity, §9.1 dedup]
func (r *SignalRepository) LogAndRegisterSignal(
	ctx context.Context,
	log *domain.PolicySignalLog,
	reg *domain.ProcessedSignalRegistry,
) error {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	batch := &pgx.Batch{}

	// Query 1: INSERT policy_signal_log (audit record).
	// ⚠️ received_at = partition key — must be in every INSERT.
	logQuery := dblib.Psql.Insert(signalLogTable).
		Columns(
			"policy_id", "signal_channel", "signal_payload",
			"source_service", "source_workflow_id", "request_id",
			"received_at", "processed_at", "status", "rejection_reason",
			"state_before", "state_after",
		).
		Values(
			log.PolicyID, log.SignalChannel, log.SignalPayload,
			log.SourceService, log.SourceWorkflowID, log.RequestID,
			log.ReceivedAt, log.ProcessedAt, log.Status, log.RejectionReason,
			log.StateBefore, log.StateAfter,
		)
	dblib.QueueExecRow(batch, logQuery)

	// Query 2: INSERT processed_signal_registry (dedup marker, 90-day TTL).
	// ON CONFLICT DO NOTHING — idempotent for Temporal activity retries.
	regQuery := dblib.Psql.Insert(signalRegistryTable).
		Columns("request_id", "signal_type", "policy_id", "received_at", "expires_at").
		Values(reg.RequestID, reg.SignalType, reg.PolicyID, reg.ReceivedAt, reg.ExpiresAt).
		Suffix("ON CONFLICT (request_id, signal_type) DO NOTHING")
	dblib.QueueExecRow(batch, regQuery)

	if err := r.db.SendBatch(ctx, batch).Close(); err != nil {
		return fmt.Errorf("LogAndRegisterSignal policyID=%d channel=%s requestID=%s: %w",
			log.PolicyID, log.SignalChannel, reg.RequestID, err)
	}
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// PublishEvent — INSERT into policy_event (outbound event log)
// ─────────────────────────────────────────────────────────────────────────────

// PublishEvent inserts an outbound policy event into policy_event.
// Called by PublishEventActivity to record events for downstream consumers
// (Notification, Accounting, Agent, Audit services — ADR-004).
// ⚠️ published_at is the partition key for policy_event — always supplied.
// [ADR-004, §21.1 PublishEventActivity]
func (r *SignalRepository) PublishEvent(ctx context.Context, event *domain.PolicyEvent) error {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	publishedAt := time.Now().UTC()
	if !event.PublishedAt.IsZero() {
		publishedAt = event.PublishedAt
	}

	query := dblib.Psql.Insert(policyEventTable).
		Columns(
			"policy_id", "event_type", "event_payload",
			"published_at", "consumed_by",
		).
		Values(
			event.PolicyID, event.EventType, event.EventPayload,
			publishedAt, event.ConsumedBy,
		).
		// ⚠️ published_at must be set — partition key.
		Suffix("RETURNING id")

	type idRow struct{ ID int64 `db:"id"` }
	if _, err := dblib.InsertReturning(ctx, r.db, query, pgx.RowToStructByNameLax[idRow]); err != nil {
		return fmt.Errorf("PublishEvent policyID=%d type=%s: %w", event.PolicyID, event.EventType, err)
	}
	return nil
}
