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

// ServiceRequestRepository handles all data access for the service_request table.
//
// ⚠️ service_request is partitioned by submitted_at (quarterly). Including submitted_at
// in WHERE clauses enables partition pruning. Methods that only have request_id will
// scan all partitions (acceptable at ~20K requests/day). Where submitted_at is known,
// it should be passed to enable pruning.
//
// Constraint C7: All SQL uses policy_mgmt. schema prefix.
// [FR-PM-006, FR-PM-007, FR-PM-008]
type ServiceRequestRepository struct {
	db  *dblib.DB
	cfg *config.Config
}

// NewServiceRequestRepository constructs a ServiceRequestRepository.
func NewServiceRequestRepository(db *dblib.DB, cfg *config.Config) *ServiceRequestRepository {
	return &ServiceRequestRepository{db: db, cfg: cfg}
}

// table/view name constants (C7: policy_mgmt. prefix)
const (
	srTable         = "policy_mgmt.service_request"
	mvPendingSummary = "policy_mgmt.mv_pending_summary"
)

// srColumns is the full column projection for service_request SELECT queries.
var srColumns = []string{
	"request_id", "policy_id", "policy_number",
	"request_type", "request_category", "status", "source_channel",
	"submitted_by", "submitted_at",
	"state_gate_status",
	"routed_at", "downstream_service", "downstream_workflow_id", "downstream_task_queue",
	"completed_at", "outcome", "outcome_reason", "outcome_payload",
	"request_payload",
	"timeout_at", "idempotency_key",
	"created_at", "updated_at",
}

// ─────────────────────────────────────────────────────────────────────────────
// Filter structs — used by list/query methods
// ─────────────────────────────────────────────────────────────────────────────

// ListRequestsFilter defines optional filters for listing service requests per policy.
// Used by GET /policies/{pn}/requests. All fields are optional.
type ListRequestsFilter struct {
	RequestType *string // Filter by request_type enum value
	Status      *string // Filter by request_status enum value
	Skip        uint64  // Pagination offset (default 0)
	Limit       uint64  // Page size (default 10, max 100)
	OrderBy     string  // Column to order by (default "submitted_at")
	SortType    string  // "ASC" or "DESC" (default "DESC")
}

// PendingRequestsFilter defines optional filters for the CPC pending inbox.
// Used by GET /requests/pending.
type PendingRequestsFilter struct {
	RequestType   *string // Filter by specific request type
	SourceChannel *string // Filter by source channel
	Skip          uint64
	Limit         uint64
}

// ─────────────────────────────────────────────────────────────────────────────
// CreateServiceRequest — INSERT service_request (status = RECEIVED)
// ─────────────────────────────────────────────────────────────────────────────

// CreateServiceRequest inserts a new service request record with status=RECEIVED.
// Returns the persisted request with request_id (BIGINT from seq_service_request_id)
// and submitted_at (partition key) populated.
// Idempotency: idempotency_key UNIQUE constraint prevents duplicate submissions.
// ⚠️ submitted_at is set to NOW() — acts as the partition key for service_request.
// [FR-PM-006]
func (r *ServiceRequestRepository) CreateServiceRequest(ctx context.Context, sr *domain.ServiceRequest) (*domain.ServiceRequest, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	now := time.Now().UTC()

	query := dblib.Psql.Insert(srTable).
		Columns(
			"policy_id", "policy_number",
			"request_type", "request_category", "status", "source_channel",
			"submitted_by", "submitted_at",
			"state_gate_status",
			"request_payload",
			"timeout_at", "idempotency_key",
			"created_at", "updated_at",
		).
		Values(
			sr.PolicyID, sr.PolicyNumber,
			sr.RequestType, sr.RequestCategory, domain.RequestStatusReceived, sr.SourceChannel,
			sr.SubmittedBy, now,
			sr.StateGateStatus,
			sr.RequestPayload,
			sr.TimeoutAt, sr.IdempotencyKey,
			now, now,
		).
		Suffix("RETURNING " + joinColumns(srColumns))

	result, err := dblib.InsertReturning(ctx, r.db, query, pgx.RowToStructByNameLax[domain.ServiceRequest])
	if err != nil {
		return nil, fmt.Errorf("CreateServiceRequest policy=%s type=%s: %w", sr.PolicyNumber, sr.RequestType, err)
	}
	return &result, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// GetServiceRequest — fetch by request_id
// ─────────────────────────────────────────────────────────────────────────────

// GetServiceRequest retrieves a service request by its BIGINT request_id.
// ⚠️ Does NOT include submitted_at in WHERE → scans all partitions. This is
// acceptable (indexed PK scan) for single-request lookups. For batch updates
// within activities, use the submitted_at variant for partition pruning.
// [FR-PM-006]
func (r *ServiceRequestRepository) GetServiceRequest(ctx context.Context, requestID int64) (*domain.ServiceRequest, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select(srColumns...).
		From(srTable).
		Where(sq.Eq{"request_id": requestID})

	sr, err := dblib.SelectOne(ctx, r.db, query, pgx.RowToStructByNameLax[domain.ServiceRequest])
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, pgx.ErrNoRows // ERR-PM-006: Request not found
		}
		return nil, fmt.Errorf("GetServiceRequest requestID=%d: %w", requestID, err)
	}
	return &sr, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// ListServiceRequestsByPolicy — paginated list with optional filters
// ─────────────────────────────────────────────────────────────────────────────

// ListServiceRequestsByPolicy returns a paginated list of service requests for a policy.
// Uses idx_sr_policy_id index (policy_id, submitted_at DESC) for efficient lookup.
// Returns the list and the total count for pagination metadata.
// [FR-PM-006]
func (r *ServiceRequestRepository) ListServiceRequestsByPolicy(
	ctx context.Context,
	policyID int64,
	f ListRequestsFilter,
) ([]domain.ServiceRequest, int64, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
	defer cancel()

	// Apply defaults.
	if f.Limit == 0 {
		f.Limit = 10
	}
	if f.OrderBy == "" {
		f.OrderBy = "submitted_at"
	}
	if f.SortType == "" {
		f.SortType = "DESC"
	}

	base := dblib.Psql.Select().From(srTable).
		Where(sq.Eq{"policy_id": policyID})

	// Optional filters.
	if f.RequestType != nil {
		base = base.Where(sq.Eq{"request_type": *f.RequestType})
	}
	if f.Status != nil {
		base = base.Where(sq.Eq{"status": *f.Status})
	}

	batch := &pgx.Batch{}

	// Query 1: total count (for pagination).
	countQuery := base.Columns("COUNT(*) AS count")
	var total struct{ Count int64 `db:"count"` }
	dblib.QueueReturnRow(batch, countQuery, pgx.RowToStructByNameLax[struct{ Count int64 `db:"count"` }], &total)

	// Query 2: paginated data.
	dataQuery := base.Columns(srColumns...).
		OrderBy(f.OrderBy + " " + f.SortType).
		Limit(f.Limit).
		Offset(f.Skip)
	var requests []domain.ServiceRequest
	dblib.QueueReturn(batch, dataQuery, pgx.RowToStructByNameLax[domain.ServiceRequest], &requests)

	if err := r.db.SendBatch(ctx, batch).Close(); err != nil {
		return nil, 0, fmt.Errorf("ListServiceRequestsByPolicy policyID=%d: %w", policyID, err)
	}
	return requests, total.Count, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// GetPendingRequests — CPC inbox: RECEIVED, ROUTED, IN_PROGRESS requests
// ─────────────────────────────────────────────────────────────────────────────

// GetPendingRequests returns paginated pending service requests for the CPC inbox.
// Uses idx_sr_pending_cpc partial index (status, request_type, submitted_at DESC)
// WHERE status IN ('RECEIVED', 'ROUTED', 'IN_PROGRESS').
// Returns list and total count for pagination.
// [FR-PM-008]
func (r *ServiceRequestRepository) GetPendingRequests(
	ctx context.Context,
	f PendingRequestsFilter,
) ([]domain.ServiceRequest, int64, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
	defer cancel()

	if f.Limit == 0 {
		f.Limit = 20
	}

	pendingStatuses := []string{
		domain.RequestStatusReceived,
		domain.RequestStatusRouted,
		domain.RequestStatusInProgress,
	}

	base := dblib.Psql.Select().From(srTable).
		Where(sq.Eq{"status": pendingStatuses})

	if f.RequestType != nil {
		base = base.Where(sq.Eq{"request_type": *f.RequestType})
	}
	if f.SourceChannel != nil {
		base = base.Where(sq.Eq{"source_channel": *f.SourceChannel})
	}

	batch := &pgx.Batch{}

	countQuery := base.Columns("COUNT(*) AS count")
	var total struct{ Count int64 `db:"count"` }
	dblib.QueueReturnRow(batch, countQuery, pgx.RowToStructByNameLax[struct{ Count int64 `db:"count"` }], &total)

	dataQuery := base.Columns(srColumns...).
		OrderBy("submitted_at DESC").
		Limit(f.Limit).
		Offset(f.Skip)
	var requests []domain.ServiceRequest
	dblib.QueueReturn(batch, dataQuery, pgx.RowToStructByNameLax[domain.ServiceRequest], &requests)

	if err := r.db.SendBatch(ctx, batch).Close(); err != nil {
		return nil, 0, fmt.Errorf("GetPendingRequests: %w", err)
	}
	return requests, total.Count, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// GetDashboardSummary — CPC dashboard counts from mv_pending_summary MV
// ─────────────────────────────────────────────────────────────────────────────

// pendingSummaryMVRow maps one row from mv_pending_summary materialized view.
type pendingSummaryMVRow struct {
	RequestType    string  `db:"request_type"`
	Status         string  `db:"status"`
	RequestCount   int     `db:"request_count"`
	OldestAgeHours float64 `db:"oldest_age_hours"`
}

// GetDashboardSummary reads from the mv_pending_summary materialized view and
// aggregates the result into the DashboardSummary domain type.
// The view is refreshed every minute via pg_cron.
// [FR-PM-008]
func (r *ServiceRequestRepository) GetDashboardSummary(ctx context.Context) (*domain.DashboardSummary, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select(
		"request_type", "status",
		"SUM(request_count) AS request_count",
		"MAX(oldest_age_hours) AS oldest_age_hours",
	).From(mvPendingSummary).
		GroupBy("request_type", "status").
		OrderBy("request_type, status")

	rows, err := dblib.SelectRows(ctx, r.db, query, pgx.RowToStructByNameLax[pendingSummaryMVRow])
	if err != nil {
		return nil, fmt.Errorf("GetDashboardSummary: %w", err)
	}

	summary := make(map[string]map[string]int)
	var totalPending int
	var oldestHours float64

	for _, row := range rows {
		if _, ok := summary[row.RequestType]; !ok {
			summary[row.RequestType] = make(map[string]int)
		}
		summary[row.RequestType][row.Status] = row.RequestCount
		totalPending += row.RequestCount
		if row.OldestAgeHours > oldestHours {
			oldestHours = row.OldestAgeHours
		}
	}

	return &domain.DashboardSummary{
		Summary:             summary,
		TotalPending:        totalPending,
		OldestRequestAgeHrs: oldestHours,
	}, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Status / outcome update methods — called by workflow activities
// ─────────────────────────────────────────────────────────────────────────────

// UpdateServiceRequestStatus updates the status of a service request.
// Used by the handler after Temporal signal delivery to mark the request as ROUTED.
// Also used by activities to mark STATE_GATE_REJECTED or CANCELLED.
// ⚠️ Gap-7: submittedAt enables partition pruning when supplied by the caller;
//
//	when nil the UPDATE scans all partitions (acceptable for infrequent updates).
//
// [FR-PM-006]
func (r *ServiceRequestRepository) UpdateServiceRequestStatus(
	ctx context.Context,
	requestID int64,
	status string,
	reason *string,
	submittedAt *time.Time, // Gap-7: optional partition key for pruning
) error {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	q := dblib.Psql.Update(srTable).
		Set("status", status).
		Set("updated_at", time.Now().UTC()).
		Where(sq.Eq{"request_id": requestID})

	if submittedAt != nil {
		q = q.Where(sq.Eq{"submitted_at": *submittedAt})
	}
	if reason != nil {
		q = q.Set("outcome_reason", *reason)
	}
	if status == domain.RequestStatusRouted {
		q = q.Set("routed_at", time.Now().UTC())
	}

	// No RETURNING clause — plain UPDATE per database-library.md rules.
	if _, err := dblib.Update(ctx, r.db, q); err != nil {
		return fmt.Errorf("UpdateServiceRequestStatus requestID=%d status=%s: %w", requestID, status, err)
	}
	return nil
}

// UpdateServiceRequestOutcome sets the final outcome on a completed service request.
// Called by UpdateServiceRequestActivity inside the workflow when downstream
// signals completion (approved, rejected, timeout, etc.).
// Sets status=COMPLETED and records the outcome + outcome_payload.
// ⚠️ Gap-7: submittedAt enables partition pruning when supplied by the caller.
// [FR-PM-006]
func (r *ServiceRequestRepository) UpdateServiceRequestOutcome(
	ctx context.Context,
	requestID int64,
	outcome string,
	outcomeReason *string,
	outcomePayload []byte,
	submittedAt *time.Time, // Gap-7: optional partition key for pruning
) error {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	now := time.Now().UTC()
	q := dblib.Psql.Update(srTable).
		Set("status", domain.RequestStatusCompleted).
		Set("outcome", outcome).
		Set("outcome_reason", outcomeReason).
		Set("outcome_payload", outcomePayload).
		Set("completed_at", now).
		Set("updated_at", now).
		Where(sq.Eq{"request_id": requestID})

	if submittedAt != nil {
		q = q.Where(sq.Eq{"submitted_at": *submittedAt})
	}

	// No RETURNING clause — plain UPDATE per database-library.md rules.
	if _, err := dblib.Update(ctx, r.db, q); err != nil {
		return fmt.Errorf("UpdateServiceRequestOutcome requestID=%d outcome=%s: %w", requestID, outcome, err)
	}
	return nil
}

// WithdrawServiceRequest marks a service request as WITHDRAWN.
// Called from the withdrawal handler after the "withdrawal-request" Temporal signal
// is delivered to plw-{policyNumber}. The workflow handles state revert internally.
// ⚠️ Gap-7: submittedAt enables partition pruning when supplied by the caller.
// Only RECEIVED or ROUTED requests can be withdrawn (not IN_PROGRESS) — BR-PM-090.
// [FR-PM-007, BR-PM-090]
func (r *ServiceRequestRepository) WithdrawServiceRequest(
	ctx context.Context,
	requestID int64,
	reason string,
	submittedAt *time.Time, // Gap-7: optional partition key for pruning
) error {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	now := time.Now().UTC()
	q := dblib.Psql.Update(srTable).
		Set("status", domain.RequestStatusWithdrawn).
		Set("outcome", domain.RequestOutcomeWithdrawn).
		Set("outcome_reason", reason).
		Set("completed_at", now).
		Set("updated_at", now).
		Where(sq.Eq{"request_id": requestID}).
		// Only RECEIVED or ROUTED requests can be withdrawn (not IN_PROGRESS).
		Where(sq.Eq{"status": []string{domain.RequestStatusReceived, domain.RequestStatusRouted}})

	if submittedAt != nil {
		q = q.Where(sq.Eq{"submitted_at": *submittedAt})
	}

	// No RETURNING clause — plain UPDATE per database-library.md rules.
	if _, err := dblib.Update(ctx, r.db, q); err != nil {
		return fmt.Errorf("WithdrawServiceRequest requestID=%d: %w", requestID, err)
	}
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// CheckIdempotencyKey — duplicate submission prevention
// ─────────────────────────────────────────────────────────────────────────────

// CheckIdempotencyKey looks up an existing service request by idempotency key.
// Returns nil, nil if no prior request with this key exists (safe to proceed).
// Returns the existing ServiceRequest if found (caller should return 202 with original request_id).
// [FR-PM-006: Idempotency via X-Idempotency-Key header]
func (r *ServiceRequestRepository) CheckIdempotencyKey(ctx context.Context, idempotencyKey string) (*domain.ServiceRequest, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select(srColumns...).
		From(srTable).
		Where(sq.Eq{"idempotency_key": idempotencyKey}).
		Limit(1)

	sr, err := dblib.SelectOne(ctx, r.db, query, pgx.RowToStructByNameLax[domain.ServiceRequest])
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // no duplicate — safe to create
		}
		return nil, fmt.Errorf("CheckIdempotencyKey key=%s: %w", idempotencyKey, err)
	}
	return &sr, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// UpdateDownstreamRouting — set downstream workflow ID after Temporal signal
// ─────────────────────────────────────────────────────────────────────────────

// UpdateDownstreamRouting records the child workflow ID and task queue after the
// Temporal signal has been delivered and the child workflow has started.
// Called by the workflow activity after ExecuteChildWorkflow returns.
// ⚠️ Gap-7: submittedAt enables partition pruning when supplied by the caller.
// [FR-PM-006, Constraint 1]
func (r *ServiceRequestRepository) UpdateDownstreamRouting(
	ctx context.Context,
	requestID int64,
	downstreamWorkflowID string,
	downstreamTaskQueue string,
	downstreamService string,
	submittedAt *time.Time, // Gap-7: optional partition key for pruning
) error {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	now := time.Now().UTC()
	q := dblib.Psql.Update(srTable).
		Set("downstream_workflow_id", downstreamWorkflowID).
		Set("downstream_task_queue", downstreamTaskQueue).
		Set("downstream_service", downstreamService).
		Set("routed_at", now).
		Set("status", domain.RequestStatusRouted).
		Set("updated_at", now).
		Where(sq.Eq{"request_id": requestID})

	if submittedAt != nil {
		q = q.Where(sq.Eq{"submitted_at": *submittedAt})
	}

	// No RETURNING clause — plain UPDATE per database-library.md rules.
	if _, err := dblib.Update(ctx, r.db, q); err != nil {
		return fmt.Errorf("UpdateDownstreamRouting requestID=%d: %w", requestID, err)
	}
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Internal helper
// ─────────────────────────────────────────────────────────────────────────────

// joinColumns joins column names with commas for use in RETURNING clauses.
func joinColumns(cols []string) string {
	result := ""
	for i, c := range cols {
		if i > 0 {
			result += ", "
		}
		result += c
	}
	return result
}

