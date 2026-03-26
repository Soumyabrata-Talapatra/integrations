// Package postgres — audit_log repository.
//
// BR-NFS-016: Every operation must produce an immutable audit record.
// E-6: audit_log is an INSERT-ONLY table (no UPDATE, no DELETE, partitioned by year).
//
// BATCH NOTE:
//
//	This repository is INSERT-only. All inserts are expected to be batched
//	with the primary operation in the caller's TX batch (service_request.go,
//	address_change.go, document_upload.go). These standalone Create/CreateBatch
//	methods are provided for cases where a standalone audit insert is needed
//	(e.g. read-only operations, lookup calls, error captures).
//
//	ListByRequestID → COUNT(*) + SELECT rows in a single pgx.Batch.
package postgres

import (
	"context"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"

	config "gitlab.cept.gov.in/it-2.0-common/api-config"
	dblib "gitlab.cept.gov.in/it-2.0-common/n-api-db"
	log "gitlab.cept.gov.in/it-2.0-common/n-api-log"

	"customer-nfs-service/core/domain"
)

// AuditLogRepository handles INSERT-only operations for nfs.audit_log (E-6).
// The table is partitioned by year; all queries must include performed_at for
// partition pruning.
type AuditLogRepository struct {
	db  *dblib.DB
	cfg *config.Config
}

// NewAuditLogRepository constructs the repository.
func NewAuditLogRepository(db *dblib.DB, cfg *config.Config) *AuditLogRepository {
	return &AuditLogRepository{db: db, cfg: cfg}
}

// ---------------------------------------------------------------------------
// Create inserts a single audit record.
// BR-NFS-016. Use this only for standalone audit inserts; prefer batching
// with the primary operation TX whenever possible.
// ---------------------------------------------------------------------------
func (r *AuditLogRepository) Create(ctx context.Context, a *domain.AuditLog) error {
	timeout := r.cfg.GetDuration("db.QueryTimeoutLow")
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	query := dblib.Psql.Insert("nfs.audit_log").
		Columns(
			"audit_id", "request_id", "action_type",
			"old_value", "new_value",
			"performed_by", "performed_at",
			"ip_address", "channel", "office_code",
			"remarks", "error_code", "error_message",
		).
		Values(
			a.AuditID, a.RequestID, a.ActionType,
			a.OldValue, a.NewValueJSON,
			a.PerformedByID, a.PerformedAt,
			a.IPAddress, a.Channel, a.OfficeCode,
			a.Notes, a.ErrorCode, a.ErrorMessage,
		)

	// Use dblib.Insert with RETURNING audit_id so dblib can scan result.
	// Append RETURNING to query returned by squirrel.

	_, err := dblib.Insert(ctx, r.db, query)
	if err != nil {
		log.Error(ctx, "AuditLog.Create [%s]: %v", a.RequestID, err)
		return fmt.Errorf("insert audit_log: %w", err)
	}
	return nil
}

// ---------------------------------------------------------------------------
// CreateBatch inserts multiple audit records in a single pgx.Batch.
// Use when multiple audit events occur in the same logical operation
// (e.g. bulk document uploads trigger one audit per doc).
//
// BATCH: All INSERTs sent in one pgx.Batch round-trip.
// BR-NFS-016
// ---------------------------------------------------------------------------
func (r *AuditLogRepository) CreateBatch(ctx context.Context, records []domain.AuditLog) error {
	if len(records) == 0 {
		return nil
	}

	timeout := r.cfg.GetDuration("db.QueryTimeoutMed")
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	batch := &pgx.Batch{}
	for _, a := range records {
		q, args, err := dblib.Psql.Insert("nfs.audit_log").
			Columns(
				"audit_id", "request_id", "action_type",
				"old_value", "new_value",
				"performed_by", "performed_at",
				"ip_address", "channel", "office_code",
				"remarks", "error_code", "error_message",
			).
			Values(
				a.AuditID, a.RequestID, a.ActionType,
				a.OldValue, a.NewValueJSON,
				a.PerformedByID, a.PerformedAt,
				a.IPAddress, a.Channel, a.OfficeCode,
				a.Notes, a.ErrorCode, a.ErrorMessage,
			).
			ToSql()
		if err != nil {
			return fmt.Errorf("build audit_log batch insert: %w", err)
		}
		batch.Queue(q, args...)
	}

	br := r.db.SendBatch(ctx, batch)
	defer br.Close()

	for i := range records {
		if _, err := br.Exec(); err != nil {
			log.Error(ctx, "AuditLog.CreateBatch [%d]: %v", i, err)
			return fmt.Errorf("batch audit_log insert[%d]: %w", i, err)
		}
	}
	return nil
}

// ---------------------------------------------------------------------------
// ListByRequestID returns paginated audit records for a request.
//
// BATCH: COUNT(*) + SELECT rows in a single pgx.Batch (one round-trip).
// DM-004, ST-002 (FR-NFS-012 — timeline/audit trail)
// ---------------------------------------------------------------------------
func (r *AuditLogRepository) ListByRequestID(
	ctx context.Context,
	requestID string,
	skip, limit int,
) ([]domain.AuditLog, int64, error) {
	timeout := r.cfg.GetDuration("db.QueryTimeoutMed")
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	baseWhere := sq.Eq{"request_id": requestID}

	countSQL, countArgs, err := dblib.Psql.Select("COUNT(*)").
		From("nfs.audit_log").
		Where(baseWhere).
		ToSql()
	if err != nil {
		return nil, 0, fmt.Errorf("build audit count query: %w", err)
	}

	listSQL, listArgs, err := dblib.Psql.Select(
		"audit_id", "request_id", "action_type",
		"old_value", "new_value",
		"performed_by", "performed_at",
		"ip_address", "channel", "office_code",
		"remarks", "error_code", "error_message",
	).
		From("nfs.audit_log").
		Where(baseWhere).
		OrderBy("performed_at ASC").
		Limit(uint64(limit)).
		Offset(uint64(skip)).
		ToSql()
	if err != nil {
		return nil, 0, fmt.Errorf("build audit list query: %w", err)
	}

	// BATCH: count + list in one round-trip
	batch := &pgx.Batch{}
	batch.Queue(countSQL, countArgs...)
	batch.Queue(listSQL, listArgs...)

	br := r.db.SendBatch(ctx, batch)
	defer br.Close()

	var total int64
	if err := br.QueryRow().Scan(&total); err != nil {
		log.Error(ctx, "AuditLog.ListByRequestID count: %v", err)
		return nil, 0, fmt.Errorf("audit count query: %w", err)
	}

	rows, err := br.Query()
	if err != nil {
		log.Error(ctx, "AuditLog.ListByRequestID list: %v", err)
		return nil, 0, fmt.Errorf("audit list query: %w", err)
	}
	defer rows.Close()

	var results []domain.AuditLog
	for rows.Next() {
		var a domain.AuditLog
		if err := rows.Scan(
			&a.AuditID, &a.RequestID, &a.ActionType,
			&a.OldValue, &a.NewValueJSON,
			&a.PerformedByID, &a.PerformedAt,
			&a.IPAddress, &a.Channel, &a.OfficeCode,
			&a.Notes, &a.ErrorCode, &a.ErrorMessage,
		); err != nil {
			return nil, 0, fmt.Errorf("scan audit_log row: %w", err)
		}
		results = append(results, a)
	}
	return results, total, nil
}

// ---------------------------------------------------------------------------
// ListByDateRange returns audit records for reporting/SLA analysis.
// Used by the CPC dashboard (CPC-006).
// BATCH: COUNT + SELECT rows.
// ---------------------------------------------------------------------------
func (r *AuditLogRepository) ListByDateRange(
	ctx context.Context,
	from, to time.Time,
	officeCode string,
	skip, limit int,
) ([]domain.AuditLog, int64, error) {
	timeout := r.cfg.GetDuration("db.QueryTimeoutMed")
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	baseWhere := sq.And{
		sq.GtOrEq{"performed_at": from},
		sq.LtOrEq{"performed_at": to},
	}
	if officeCode != "" {
		baseWhere = append(baseWhere, sq.Eq{"office_code": officeCode})
	}

	countSQL, countArgs, err := dblib.Psql.Select("COUNT(*)").
		From("nfs.audit_log").
		Where(baseWhere).
		ToSql()
	if err != nil {
		return nil, 0, fmt.Errorf("build date-range count query: %w", err)
	}

	listSQL, listArgs, err := dblib.Psql.Select(
		"audit_id", "request_id", "action_type",
		"performed_by", "performed_at", "channel", "office_code", "remarks",
	).
		From("nfs.audit_log").
		Where(baseWhere).
		OrderBy("performed_at DESC").
		Limit(uint64(limit)).
		Offset(uint64(skip)).
		ToSql()
	if err != nil {
		return nil, 0, fmt.Errorf("build date-range list query: %w", err)
	}

	// BATCH: count + list in one round-trip
	batch := &pgx.Batch{}
	batch.Queue(countSQL, countArgs...)
	batch.Queue(listSQL, listArgs...)

	br := r.db.SendBatch(ctx, batch)
	defer br.Close()

	var total int64
	if err := br.QueryRow().Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("date-range count query: %w", err)
	}

	rows, err := br.Query()
	if err != nil {
		return nil, 0, fmt.Errorf("date-range list query: %w", err)
	}
	defer rows.Close()

	var results []domain.AuditLog
	for rows.Next() {
		var a domain.AuditLog
		if err := rows.Scan(
			&a.AuditID, &a.RequestID, &a.ActionType,
			&a.PerformedByID, &a.PerformedAt, &a.Channel, &a.OfficeCode, &a.Notes,
		); err != nil {
			return nil, 0, fmt.Errorf("scan audit date-range row: %w", err)
		}
		results = append(results, a)
	}
	return results, total, nil
}
