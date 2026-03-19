package repo

import (
	"context"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	config "gitlab.cept.gov.in/it-2.0-common/api-config"
	dblib "gitlab.cept.gov.in/it-2.0-common/n-api-db"

	"policy-management/core/domain"
)

// ErrPolicyVersionConflict is returned by UpdatePolicyStatus when the expected
// version doesn't match the persisted version (optimistic locking failure).
// Callers (Temporal activities) should treat this as a retryable error.
var ErrPolicyVersionConflict = errors.New("policy version conflict: concurrent modification detected [optimistic lock]")

// PolicyRepository handles all data access for the policy lifecycle state,
// financial lock, status history, terminal snapshots, and dashboard metrics.
//
// Constraint C7: All SQL uses the policy_mgmt. schema prefix.
// Constraint 8: Two-Tier Query — handlers call QueryWorkflow first; this repo
//               serves Tier-2 (terminal fallback) and batch scan queries.
// [FR-PM-001, FR-PM-002, FR-PM-011..FR-PM-015]
type PolicyRepository struct {
	db  *dblib.DB
	cfg *config.Config
}

// NewPolicyRepository constructs a PolicyRepository with the injected DB pool and config.
func NewPolicyRepository(db *dblib.DB, cfg *config.Config) *PolicyRepository {
	return &PolicyRepository{db: db, cfg: cfg}
}

// ─────────────────────────────────────────────────────────────────────────────
// Table / view name constants — all with policy_mgmt. schema prefix (C7)
// ─────────────────────────────────────────────────────────────────────────────

const (
	policyTable        = "policy_mgmt.policy"
	policyHistoryTable = "policy_mgmt.policy_status_history"
	policyLockTable    = "policy_mgmt.policy_lock"
	terminalSnapTable  = "policy_mgmt.terminal_state_snapshot"
	mvPolicyDashboard  = "policy_mgmt.mv_policy_dashboard"
)

// policyColumns is the complete column projection for SELECT on policy table.
// Used by all single-policy and batch-policy queries.
var policyColumns = []string{
	"policy_id", "policy_number", "customer_id", "product_code", "product_type",
	"current_status", "previous_status", "previous_status_before_suspension", "effective_from",
	"sum_assured", "current_premium", "premium_mode", "billing_method",
	"issue_date", "policy_inception_date", "maturity_date", "paid_to_date", "next_premium_due_date",
	"agent_id",
	"has_active_loan", "loan_outstanding", "assignment_type", "assignment_status",
	"aml_hold", "dispute_flag", "murder_clause_active",
	"display_status",
	"first_unpaid_premium_date", "remission_expiry_date", "pay_recovery_protection_expiry",
	"paid_up_value", "paid_up_type", "paid_up_date",
	"sb_installments_paid", "sb_total_amount_paid",
	"nomination_status",
	"policyholder_dob",
	"workflow_id", "temporal_run_id",
	"version", "created_at", "updated_at", "created_by", "updated_by",
}

// ─────────────────────────────────────────────────────────────────────────────
// GetPolicyByNumber — Tier-2 DB lookup by human-readable policy number.
// ─────────────────────────────────────────────────────────────────────────────

// GetPolicyByNumber retrieves a policy by its human-readable policy number.
// Returns pgx.ErrNoRows if not found (mapped to ERR-PM-003 at handler layer).
// VR-PM-001: Policy number format validation is done at the handler layer.
// [FR-PM-001]
func (r *PolicyRepository) GetPolicyByNumber(ctx context.Context, policyNumber string) (*domain.Policy, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select(policyColumns...).
		From(policyTable).
		Where(sq.Eq{"policy_number": policyNumber})

	p, err := dblib.SelectOne(ctx, r.db, query, pgx.RowToStructByNameLax[domain.Policy])
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, pgx.ErrNoRows // ERR-PM-003: Policy not found
		}
		return nil, fmt.Errorf("GetPolicyByNumber %q: %w", policyNumber, err)
	}
	return &p, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// GetPolicyByID — internal lookup by BIGINT policy_id (used by activities).
// ─────────────────────────────────────────────────────────────────────────────

// GetPolicyByID retrieves a policy by its internal BIGINT policy_id.
// Used by workflow activities that have the policy_id from the workflow state.
// [FR-PM-001]
func (r *PolicyRepository) GetPolicyByID(ctx context.Context, policyID int64) (*domain.Policy, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select(policyColumns...).
		From(policyTable).
		Where(sq.Eq{"policy_id": policyID})

	p, err := dblib.SelectOne(ctx, r.db, query, pgx.RowToStructByNameLax[domain.Policy])
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, pgx.ErrNoRows
		}
		return nil, fmt.Errorf("GetPolicyByID %d: %w", policyID, err)
	}
	return &p, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// GetPolicyBatchStatus — GET /policies/batch-status (Tier-2 fallback per policy)
// ─────────────────────────────────────────────────────────────────────────────

// GetPolicyBatchStatus retrieves lightweight status data for multiple policies.
// Called by the batch-status handler for policies whose workflows are terminal
// (Tier-2 fallback). Active workflows are queried in parallel via QueryWorkflow.
// [FR-PM-001]
func (r *PolicyRepository) GetPolicyBatchStatus(ctx context.Context, policyNumbers []string) ([]domain.Policy, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
	defer cancel()

	// Minimal projection — only fields needed by the batch-status response DTO.
	query := dblib.Psql.Select(
		"policy_id", "policy_number", "current_status", "display_status",
		"has_active_loan", "assignment_type", "aml_hold", "dispute_flag",
		"workflow_id", "version",
	).From(policyTable).
		Where(sq.Eq{"policy_number": policyNumbers})

	policies, err := dblib.SelectRows(ctx, r.db, query, pgx.RowToStructByNameLax[domain.Policy])
	if err != nil {
		return nil, fmt.Errorf("GetPolicyBatchStatus: %w", err)
	}
	return policies, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// UpdatePolicyStatus — pgx.Batch: UPDATE policy + INSERT policy_status_history
// Called by RecordStateTransitionActivity for all state transitions.
// ─────────────────────────────────────────────────────────────────────────────

// UpdatePolicyStatus records a lifecycle state transition with optimistic locking.
// Steps (two separate round-trips — batch cannot detect 0-rows from version mismatch):
//  1. UPDATE policy WHERE version = expectedVersion — RETURNING policy_id to detect mismatch.
//  2. INSERT policy_status_history audit row (only if step 1 succeeds).
//
// Returns ErrPolicyVersionConflict if expectedVersion doesn't match persisted version.
// Temporal activities must treat ErrPolicyVersionConflict as retryable.
// ⚠️ effective_date is the partition key for policy_status_history — always SET.
// NOTE: Called for SUBSEQUENT transitions only. The initial FREE_LOOK_ACTIVE row
//
//	is inserted by InitializePolicyActivity. [§10.1, §13]
//
// [FR-PM-002, BR-PM-011..023, Gap-4 optimistic lock]
func (r *PolicyRepository) UpdatePolicyStatus(
	ctx context.Context,
	policyID int64,
	fromStatus, toStatus string,
	reason, triggeredBy string,
	triggeredByUserID *int64,
	requestID *int64,
	expectedVersion int64,
) error {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	now := time.Now().UTC()

	// Step 1: UPDATE policy with version guard.
	// WHERE version = expectedVersion ensures no lost updates (Gap-4).
	// RETURNING policy_id — UpdateReturning returns empty slice if version doesn't match
	// (0 rows updated), which we map to ErrPolicyVersionConflict.
	updateQuery := dblib.Psql.Update(policyTable).
		Set("current_status", toStatus).
		Set("previous_status", fromStatus).
		Set("effective_from", now).
		Set("version", sq.Expr("version + 1")).
		Set("updated_at", now).
		Where(sq.Eq{"policy_id": policyID}).
		Where(sq.Eq{"version": expectedVersion}).
		Suffix("RETURNING policy_id")

	type policyIDRow struct {
		PolicyID int64 `db:"policy_id"`
	}
	updatedRows, err := dblib.UpdateReturningBulk(ctx, r.db, updateQuery, pgx.RowToStructByNameLax[policyIDRow])
	if err != nil {
		return fmt.Errorf("UpdatePolicyStatus update policyID=%d %s→%s: %w", policyID, fromStatus, toStatus, err)
	}
	if len(updatedRows) == 0 {
		return fmt.Errorf("UpdatePolicyStatus policyID=%d version=%d: %w",
			policyID, expectedVersion, ErrPolicyVersionConflict)
	}

	// Step 2: INSERT policy_status_history (no RETURNING needed — use dblib.Insert).
	// ⚠️ effective_date = partition key — must be supplied on every INSERT.
	histQuery := dblib.Psql.Insert(policyHistoryTable).
		Columns(
			"policy_id", "from_status", "to_status",
			"transition_reason", "triggered_by_service", "triggered_by_user_id",
			"request_id", "effective_date", "created_at",
		).
		Values(
			policyID, fromStatus, toStatus,
			reason, triggeredBy, triggeredByUserID,
			requestID, now, now,
		)
	if _, err := dblib.Insert(ctx, r.db, histQuery); err != nil {
		return fmt.Errorf("UpdatePolicyStatus history policyID=%d %s→%s: %w", policyID, fromStatus, toStatus, err)
	}
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Financial Lock Management — BR-PM-030
// ─────────────────────────────────────────────────────────────────────────────

// AcquireFinancialLock inserts a row into policy_lock to claim exclusive financial
// processing rights for a policy. PK = policy_id enforces at-most-one lock per policy.
// Returns an error (PostgreSQL unique violation 23505) if a lock already exists.
// The caller (handler or activity) must check if the error is a PK conflict.
// [BR-PM-030]
func (r *PolicyRepository) AcquireFinancialLock(
	ctx context.Context,
	policyID int64,
	requestID int64,
	requestType string,
	timeoutAt time.Time,
) error {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	// INSERT...RETURNING policy_id — InsertReturning returns empty slice if
	// ON CONFLICT DO NOTHING fires (lock already held). Any other error propagates as-is.
	lockQuery := dblib.Psql.Insert(policyLockTable).
		Columns("policy_id", "request_id", "request_type", "locked_at", "timeout_at").
		Values(policyID, requestID, requestType, time.Now().UTC(), timeoutAt).
		Suffix("ON CONFLICT (policy_id) DO NOTHING RETURNING policy_id")

	type lockIDRow struct {
		PolicyID int64 `db:"policy_id"`
	}
	// InsertReturning returns empty slice when ON CONFLICT DO NOTHING fires
	// (meaning the lock already exists — another request owns it).
	lockRows, err := dblib.InsertReturningrows(ctx, r.db, lockQuery, pgx.RowToStructByNameLax[lockIDRow])
	if err != nil {
		return fmt.Errorf("AcquireFinancialLock policyID=%d: %w", policyID, err)
	}
	if len(lockRows) == 0 {
		// ON CONFLICT fired — lock already held by another request.
		return fmt.Errorf("policy %d already has an active financial lock [BR-PM-030]", policyID)
	}
	return nil
}

// ReleaseFinancialLock deletes the financial lock for a policy.
// Called when a downstream service signals completion (any outcome) or on timeout.
// [BR-PM-030]
func (r *PolicyRepository) ReleaseFinancialLock(ctx context.Context, policyID int64) error {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Delete(policyLockTable).
		Where(sq.Eq{"policy_id": policyID})

	if _, err := dblib.Delete(ctx, r.db, query); err != nil {
		return fmt.Errorf("ReleaseFinancialLock policyID=%d: %w", policyID, err)
	}
	return nil
}

// CheckFinancialLock returns the current lock for a policy, or nil if not locked.
// Used by the handler pre-check before submitting a new financial request.
// [BR-PM-030]
func (r *PolicyRepository) CheckFinancialLock(ctx context.Context, policyID int64) (*domain.PolicyLock, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select(
		"policy_id", "request_id", "request_type", "locked_at", "timeout_at",
	).From(policyLockTable).
		Where(sq.Eq{"policy_id": policyID})

	lock, err := dblib.SelectOne(ctx, r.db, query, pgx.RowToStructByNameLax[domain.PolicyLock])
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // no lock held — policy is free
		}
		return nil, fmt.Errorf("CheckFinancialLock policyID=%d: %w", policyID, err)
	}
	return &lock, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Batch Scan Queries — used by BatchStateScanWorkflow activities (Phase 5)
// All methods use partial indexes defined in 001_policy_mgmt_schema.sql.
// ─────────────────────────────────────────────────────────────────────────────

// GetPoliciesDueForLapsation returns ACTIVE policies whose paid_to_date is before
// asOfDate, eligible for VOID_LAPSE / INACTIVE_LAPSE / VOID transition.
// Excludes PAY_RECOVERY policies still within their 12-month active protection window.
// Uses partial index: idx_policy_active_due (paid_to_date, premium_mode WHERE ACTIVE).
// The activity determines the exact target status based on policy age (BR-PM-070).
// Skips pay-recovery policies within active protection window (BR-PM-074).
// [FR-PM-011]
func (r *PolicyRepository) GetPoliciesDueForLapsation(ctx context.Context, asOfDate time.Time) ([]domain.Policy, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutHigh"))
	defer cancel()

	query := dblib.Psql.Select(policyColumns...).
		From(policyTable).
		Where(sq.Eq{"current_status": domain.StatusActive}).
		Where(sq.Lt{"paid_to_date": asOfDate}).
		// BR-PM-074: Exclude PAY_RECOVERY policies still within 12-month protection.
		Where(sq.Or{
			sq.NotEq{"billing_method": domain.BillingMethodPayRecovery},
			sq.Or{
				sq.Eq{"pay_recovery_protection_expiry": nil},
				sq.Lt{"pay_recovery_protection_expiry": asOfDate},
			},
		}).
		OrderBy("paid_to_date ASC").
		Limit(5000) // Safety cap: batch activity pages through if >5K policies

	policies, err := dblib.SelectRows(ctx, r.db, query, pgx.RowToStructByNameLax[domain.Policy])
	if err != nil {
		return nil, fmt.Errorf("GetPoliciesDueForLapsation asOf=%s: %w", asOfDate.Format("2006-01-02"), err)
	}
	return policies, nil
}

// GetPoliciesForPaidUpConversion returns ACTIVE_LAPSE policies eligible for
// automatic paid-up conversion. Eligibility (BR-PM-060,061):
//   - policy_life >= 3 years (issue_date <= now - 3 years)
//   - Activity further filters: paid-up value >= Rs.10,000 (PAID_UP) else VOID
//
// Uses partial index: idx_policy_active_lapse (first_unpaid_premium_date WHERE ACTIVE_LAPSE).
// [FR-PM-013]
func (r *PolicyRepository) GetPoliciesForPaidUpConversion(ctx context.Context) ([]domain.Policy, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutHigh"))
	defer cancel()

	threeYearsAgo := time.Now().UTC().AddDate(-3, 0, 0)

	query := dblib.Psql.Select(policyColumns...).
		From(policyTable).
		Where(sq.Eq{"current_status": domain.StatusActiveLapse}).
		// Policy must have been in force >= 3 years for paid-up eligibility.
		Where(sq.LtOrEq{"issue_date": threeYearsAgo}).
		OrderBy("issue_date ASC").
		Limit(5000)

	policies, err := dblib.SelectRows(ctx, r.db, query, pgx.RowToStructByNameLax[domain.Policy])
	if err != nil {
		return nil, fmt.Errorf("GetPoliciesForPaidUpConversion: %w", err)
	}
	return policies, nil
}

// GetPoliciesForMaturityScan returns ACTIVE policies whose maturity_date falls
// on or before the given horizon date.
// The caller (MaturityScanActivity) computes the horizon using ConfigKeyMaturityNotificationDays
// from policy_state_config (default 90 days), making this method config-agnostic.
// Uses partial index: idx_policy_pending_maturity (maturity_date WHERE ACTIVE + NOT NULL).
// The activity transitions eligible policies to PENDING_MATURITY.
// [FR-PM-015, Additional-1]
func (r *PolicyRepository) GetPoliciesForMaturityScan(ctx context.Context, horizon time.Time) ([]domain.Policy, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutHigh"))
	defer cancel()

	query := dblib.Psql.Select(policyColumns...).
		From(policyTable).
		Where(sq.Eq{"current_status": domain.StatusActive}).
		Where("maturity_date IS NOT NULL").
		// horizon is now().AddDate(0, 0, maturity_notification_days) — supplied by caller.
		Where(sq.LtOrEq{"maturity_date": horizon}).
		OrderBy("maturity_date ASC").
		Limit(5000)

	policies, err := dblib.SelectRows(ctx, r.db, query, pgx.RowToStructByNameLax[domain.Policy])
	if err != nil {
		return nil, fmt.Errorf("GetPoliciesForMaturityScan horizon=%s: %w", horizon.Format("2006-01-02"), err)
	}
	return policies, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Terminal State Snapshot — Two-Tier Query Fallback (Constraint 8, §9.5.1)
// ─────────────────────────────────────────────────────────────────────────────

// UpsertTerminalSnapshot inserts or updates the terminal_state_snapshot row for a policy.
// Called by PersistTerminalStateActivity when a policy reaches a terminal status.
// The final_snapshot JSONB contains the full serialized PolicyLifecycleState,
// enabling Tier-2 REST responses after the Temporal workflow completes cooling.
// [§9.5.1, FR-PM-001]
func (r *PolicyRepository) UpsertTerminalSnapshot(ctx context.Context, snap *domain.TerminalStateSnapshot) error {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	now := time.Now().UTC()

	query := dblib.Psql.Insert(terminalSnapTable).
		Columns(
			"policy_id", "policy_number", "final_status",
			"terminal_at", "cooling_expiry",
			"final_snapshot", "created_at",
		).
		Values(
			snap.PolicyID, snap.PolicyNumber, snap.FinalStatus,
			snap.TerminalAt, snap.CoolingExpiry,
			snap.FinalSnapshot, now,
		).
		Suffix(`ON CONFLICT (policy_id) DO UPDATE SET
			final_status   = EXCLUDED.final_status,
			terminal_at    = EXCLUDED.terminal_at,
			cooling_expiry = EXCLUDED.cooling_expiry,
			final_snapshot = EXCLUDED.final_snapshot`)

	// No RETURNING clause — ON CONFLICT DO UPDATE upserts silently.
	// Use dblib.Insert (no dest) per database-library.md rules.
	if _, err := dblib.Insert(ctx, r.db, query); err != nil {
		return fmt.Errorf("UpsertTerminalSnapshot policyID=%d: %w", snap.PolicyID, err)
	}
	return nil
}

// GetTerminalSnapshot retrieves the terminal state snapshot for a policy by number.
// Returns nil, nil when no terminal snapshot exists (policy is still active).
// [§9.5.1, Constraint 8 Tier-2 fallback]
func (r *PolicyRepository) GetTerminalSnapshot(ctx context.Context, policyNumber string) (*domain.TerminalStateSnapshot, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	// Uses idx_tss_number index on policy_number.
	query := dblib.Psql.Select(
		"policy_id", "policy_number", "final_status",
		"terminal_at", "cooling_expiry",
		"workflow_completed_at", "final_snapshot", "created_at",
	).From(terminalSnapTable).
		Where(sq.Eq{"policy_number": policyNumber})

	snap, err := dblib.SelectOne(ctx, r.db, query, pgx.RowToStructByNameLax[domain.TerminalStateSnapshot])
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // policy not yet terminal
		}
		return nil, fmt.Errorf("GetTerminalSnapshot %q: %w", policyNumber, err)
	}
	return &snap, nil
}

// MarkWorkflowCompleted sets workflow_completed_at on the terminal snapshot.
// Called by MarkWorkflowCompletedActivity after terminal cooling expires.
// [§9.5.1, Constraint 6]
func (r *PolicyRepository) MarkWorkflowCompleted(ctx context.Context, policyID int64) error {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	now := time.Now().UTC()
	query := dblib.Psql.Update(terminalSnapTable).
		Set("workflow_completed_at", now).
		Where(sq.Eq{"policy_id": policyID})

	// No RETURNING clause — plain UPDATE, use dblib.Update without dest.
	if _, err := dblib.Update(ctx, r.db, query); err != nil {
		return fmt.Errorf("MarkWorkflowCompleted policyID=%d: %w", policyID, err)
	}
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// GetPolicyStatusHistory — GET /policies/{pn}/history (DB-only endpoint)
// ─────────────────────────────────────────────────────────────────────────────

// historyColumns is the projection for policy_status_history SELECT queries.
var historyColumns = []string{
	"id", "policy_id", "from_status", "to_status",
	"transition_reason", "triggered_by_service", "triggered_by_user_id",
	"request_id", "effective_date", "created_at",
}

// GetPolicyStatusHistory retrieves paginated state transition history for a policy.
// policy_status_history is partitioned by effective_date; policyID is in WHERE for
// the partial index idx_psh_policy. Returns total count for pagination metadata.
// [FR-PM-002]
func (r *PolicyRepository) GetPolicyStatusHistory(
	ctx context.Context,
	policyID int64,
	skip, limit uint64,
) ([]domain.PolicyStatusHistory, int64, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
	defer cancel()

	if limit == 0 {
		limit = 10
	}

	base := dblib.Psql.Select().From(policyHistoryTable).
		Where(sq.Eq{"policy_id": policyID})

	batch := &pgx.Batch{}

	// Query 1: total count.
	countQuery := base.Columns("COUNT(*) AS count")
	var total struct{ Count int64 `db:"count"` }
	dblib.QueueReturnRow(batch, countQuery, pgx.RowToStructByNameLax[struct{ Count int64 `db:"count"` }], &total)

	// Query 2: paginated history rows, newest first.
	dataQuery := base.Columns(historyColumns...).
		OrderBy("effective_date DESC").
		Limit(limit).
		Offset(skip)
	var rows []domain.PolicyStatusHistory
	dblib.QueueReturn(batch, dataQuery, pgx.RowToStructByNameLax[domain.PolicyStatusHistory], &rows)

	if err := r.db.SendBatch(ctx, batch).Close(); err != nil {
		return nil, 0, fmt.Errorf("GetPolicyStatusHistory policyID=%d: %w", policyID, err)
	}
	return rows, total.Count, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// GetDashboardMetrics — GET /policies/dashboard/metrics
// Reads from mv_policy_dashboard materialized view (refreshed every 15 min).
// Also queries service_request for today's and pending counts.
// ─────────────────────────────────────────────────────────────────────────────

// policyDashboardRow maps a single row from mv_policy_dashboard.
type policyDashboardRow struct {
	CurrentStatus string `db:"current_status"`
	ProductCode   string `db:"product_code"`
	BillingMethod string `db:"billing_method"`
	PolicyCount   int64  `db:"policy_count"`
}

// requestCountRow maps a count query result.
type requestCountRow struct {
	Count int `db:"count"`
}

// GetDashboardMetrics returns aggregated policy and request counts for the admin dashboard.
// Uses mv_policy_dashboard materialized view (pre-aggregated, refreshed every 15 min).
// Uses pgx.Batch to read the MV and compute request counts in a single round-trip.
// [FR-PM-008]
func (r *PolicyRepository) GetDashboardMetrics(ctx context.Context) (*domain.DashboardMetrics, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
	defer cancel()

	batch := &pgx.Batch{}

	// Query 1: Fetch all rows from the materialized view.
	mvQuery := dblib.Psql.Select(
		"current_status", "product_code", "billing_method",
		"SUM(policy_count) AS policy_count",
	).From(mvPolicyDashboard).
		GroupBy("current_status", "product_code", "billing_method")
	var mvRows []policyDashboardRow
	dblib.QueueReturn(batch, mvQuery, pgx.RowToStructByNameLax[policyDashboardRow], &mvRows)

	// Query 2: Count requests submitted today.
	todayQuery := dblib.Psql.Select("COUNT(*) AS count").
		From("policy_mgmt.service_request").
		Where("submitted_at::date = CURRENT_DATE").
		Where(sq.Eq{"status": []string{
			domain.RequestStatusReceived, domain.RequestStatusRouted,
			domain.RequestStatusInProgress, domain.RequestStatusCompleted,
		}})
	var todayCount requestCountRow
	dblib.QueueReturnRow(batch, todayQuery, pgx.RowToStructByNameLax[requestCountRow], &todayCount)

	// Query 3: Count all non-terminal (pending) requests.
	pendingQuery := dblib.Psql.Select("COUNT(*) AS count").
		From("policy_mgmt.service_request").
		Where(sq.Eq{"status": []string{
			domain.RequestStatusReceived,
			domain.RequestStatusRouted,
			domain.RequestStatusInProgress,
		}})
	var pendingCount requestCountRow
	dblib.QueueReturnRow(batch, pendingQuery, pgx.RowToStructByNameLax[requestCountRow], &pendingCount)

	if err := r.db.SendBatch(ctx, batch).Close(); err != nil {
		return nil, fmt.Errorf("GetDashboardMetrics: %w", err)
	}

	// Aggregate mv rows into maps.
	byStatus := make(map[string]int64)
	byProduct := make(map[string]int64)
	byBilling := make(map[string]int64)
	for _, row := range mvRows {
		byStatus[row.CurrentStatus] += row.PolicyCount
		byProduct[row.ProductCode] += row.PolicyCount
		byBilling[row.BillingMethod] += row.PolicyCount
	}

	return &domain.DashboardMetrics{
		PoliciesByStatus:        byStatus,
		PoliciesByProduct:       byProduct,
		PoliciesByBillingMethod: byBilling,
		RequestsToday:           todayCount.Count,
		RequestsPending:         pendingCount.Count,
	}, nil
}
