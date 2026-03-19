package activities

// ============================================================================
// BatchActivities — Batch state scan activities for BatchStateScanWorkflow
//
// All activities follow the "Database-first" pattern (§9.5.2):
//   1. Bulk UPDATE policy + INSERT policy_status_history (pages of 100-500)
//   2. Rate-limited Temporal signals (50ms between pages) to plw-* workflows
//      with "batch-state-sync" payload (in-memory only — workflow trusts DB)
//
// Activity options for all scans:
//   StartToCloseTimeout: 2h
//   HeartbeatTimeout:    5m (calls activity.RecordHeartbeat every page)
//   RetryPolicy:         3× exponential backoff
//
// [FR-PM-011..FR-PM-015, §9.3, §9.5.2, Constraint 5]
// ============================================================================

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"

	config "gitlab.cept.gov.in/it-2.0-common/api-config"
	dblib "gitlab.cept.gov.in/it-2.0-common/n-api-db"

	"policy-management/core/domain"
)

// ─────────────────────────────────────────────────────────────────────────────
// BatchActivities struct — injected via FX
// ─────────────────────────────────────────────────────────────────────────────

// BatchActivities holds dependencies for all batch scan activities.
// [FR-PM-011..FR-PM-015, §9.3]
type BatchActivities struct {
	db         *dblib.DB
	cfg        *config.Config
	tc         client.Client
	httpClient *http.Client // for GSV lookups from surrender-svc [C8]
}

// NewBatchActivities constructs a BatchActivities instance for FX injection.
func NewBatchActivities(db *dblib.DB, cfg *config.Config, tc client.Client) *BatchActivities {
	return &BatchActivities{
		db:  db,
		cfg: cfg,
		tc:  tc,
		httpClient: &http.Client{
			Timeout: 8 * time.Second, // matches QuoteActivities pattern [C8]
		},
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Batch I/O Types
// ─────────────────────────────────────────────────────────────────────────────

// BatchScanResult is the input to RecordBatchScanResultActivity. [§8.6]
type BatchScanResult struct {
	ScanType           string    `json:"scan_type"`
	ScheduledDate      time.Time `json:"scheduled_date"`
	StartedAt          time.Time `json:"started_at"`
	CompletedAt        time.Time `json:"completed_at"`
	PoliciesScanned    int       `json:"policies_scanned"`
	TransitionsApplied int       `json:"transitions_applied"`
	Errors             int       `json:"errors"`
	Status             string    `json:"status"` // COMPLETED or FAILED
}

// batchPolicyRow is used internally when scanning policies for batch transitions.
// remission_expiry_date is intentionally excluded: after C6, remission is always
// recomputed from issue_date + paid_to_date via computeRemissionExpiry() rather
// than read from the DB column. The DB column is nullable (NULL for policies < 6 mo),
// and scanning NULL into a non-pointer time.Time panics in pgx. [D3]
type batchPolicyRow struct {
	PolicyID     int64     `db:"policy_id"`
	PolicyNumber string    `db:"policy_number"`
	IssueDate    time.Time `db:"issue_date"`
	PaidToDate   time.Time `db:"paid_to_date"`
}

// batchSyncSignal is the in-memory signal payload sent to PLW workflows. [§9.5.2]
type batchSyncSignal struct {
	NewStatus     string    `json:"new_status"`
	ScanType      string    `json:"scan_type"`
	ScheduledDate time.Time `json:"scheduled_date"`
}

// ─────────────────────────────────────────────────────────────────────────────
// Constants
// ─────────────────────────────────────────────────────────────────────────────

const (
	batchPageSize        = 500                   // Policies per page for bulk operations
	batchSignalRateDelay = 50 * time.Millisecond // Delay between signal pages [§9.5.2]

	// Signal and workflow constants (avoid circular import with workflows package)
	batchStateSyncSignal      = "batch-state-sync"       // [signals.go]
	forcedSurrenderSignal     = "forced-surrender-trigger" // [signals.go]
	policyWorkflowIDPrefix    = "plw-"
	policyManagementTaskQueue = "policy-management-tq"    // [Constraint 3]
)

// ─────────────────────────────────────────────────────────────────────────────
// LapsationScanActivity — ACTIVE → VOID_LAPSE / INACTIVE_LAPSE / VOID
// [FR-PM-011, §9.3, BR-PM-074]
// StartToCloseTimeout: 2h, HeartbeatTimeout: 5m
// ─────────────────────────────────────────────────────────────────────────────

// LapsationScanActivity processes daily lapsation for policies whose premiums
// are unpaid. Database-first pattern:
//   - policy_life < 6 months → VOID (no remission)
//   - policy_life 6mo-36mo → VOID_LAPSE (12-month remission)
//   - policy_life ≥ 36 months → INACTIVE_LAPSE (12-month remission)
// Skips pay-recovery policies within 12-month active protection. [BR-PM-074]
// [FR-PM-011, §9.3, §9.5.2]
func (a *BatchActivities) LapsationScanActivity(ctx context.Context, scheduledDate time.Time) (BatchScanResult, error) {
	startedAt := time.Now().UTC()
	result := BatchScanResult{
		ScanType:      domain.BatchScanTypeLapsation,
		ScheduledDate: scheduledDate,
		StartedAt:     startedAt,
	}

	var totalScanned, totalTransitioned int
	offset := 0

	for {
		// Heartbeat every page to prevent HeartbeatTimeout [§9.3]
		activity.RecordHeartbeat(ctx, fmt.Sprintf("lapsation page offset=%d transitioned=%d", offset, totalTransitioned))

		// Fetch page of ACTIVE policies with unpaid premiums eligible for lapsation [FR-PM-011]
		rows, err := a.fetchLapsationCandidates(ctx, scheduledDate, offset)
		if err != nil {
			result.Errors++
			result.Status = domain.BatchScanStatusFailed
			return result, fmt.Errorf("LapsationScanActivity fetchPage offset=%d: %w", offset, err)
		}
		if len(rows) == 0 {
			break // All pages processed
		}

		// [C6] Per-policy remission expiry — mirrors compute_remission_expiry() DB function.
		// paid_to_date maps to first_unpaid_date in the DB function. [§9.3, BR-PM-074]
		voidPolicies := make([]int64, 0)
		voidLapsePairs := make([]policyRemissionPair, 0)
		inactiveLapsePairs := make([]policyRemissionPair, 0)

		for _, row := range rows {
			remissionPtr := computeRemissionExpiry(row.IssueDate, row.PaidToDate, scheduledDate)
			switch {
			case remissionPtr == nil:
				// policy_life < 6 months — VOID immediately, no remission [C6]
				voidPolicies = append(voidPolicies, row.PolicyID)
			case monthsBetween(row.IssueDate, scheduledDate) < 36:
				// VOID_LAPSE: 6–35 months — grace-end + slab days [C6]
				voidLapsePairs = append(voidLapsePairs, policyRemissionPair{
					PolicyID:        row.PolicyID,
					RemissionExpiry: *remissionPtr,
				})
			default:
				// INACTIVE_LAPSE: ≥36 months — paid_to_date + 12 months [C6]
				inactiveLapsePairs = append(inactiveLapsePairs, policyRemissionPair{
					PolicyID:        row.PolicyID,
					RemissionExpiry: *remissionPtr,
				})
			}
		}

		// Bulk update each group + insert history
		now := time.Now().UTC()

		if err := a.bulkTransition(ctx, voidPolicies, domain.StatusActive, domain.StatusVoid,
			"lapsation: policy < 6 months — no remission [C6]", scheduledDate, now, nil); err != nil {
			result.Errors++
		}
		if err := a.bulkTransitionWithRemissions(ctx, voidLapsePairs, domain.StatusActive, domain.StatusVoidLapse,
			"lapsation: policy 6–35mo — grace-end slab remission [C6]", scheduledDate, now); err != nil {
			result.Errors++
		}
		if err := a.bulkTransitionWithRemissions(ctx, inactiveLapsePairs, domain.StatusActive, domain.StatusInactiveLapse,
			"lapsation: policy ≥36mo — first_unpaid+12months remission [C6]", scheduledDate, now); err != nil {
			result.Errors++
		}

		// Send batch-state-sync signals to PLW workflows (rate-limited) [§9.5.2]
		for _, row := range rows {
			remissionPtr := computeRemissionExpiry(row.IssueDate, row.PaidToDate, scheduledDate)
			var newStatus string
			switch {
			case remissionPtr == nil:
				newStatus = domain.StatusVoid
			case monthsBetween(row.IssueDate, scheduledDate) < 36:
				newStatus = domain.StatusVoidLapse
			default:
				newStatus = domain.StatusInactiveLapse
			}
			payload := batchSyncSignal{
				NewStatus:     newStatus,
				ScanType:      domain.BatchScanTypeLapsation,
				ScheduledDate: scheduledDate,
			}
			_ = a.tc.SignalWorkflow(ctx,
				policyWorkflowIDPrefix+row.PolicyNumber,
				"", batchStateSyncSignal, payload)
		}

		totalScanned += len(rows)
		totalTransitioned += len(rows) - result.Errors // approximate
		offset += batchPageSize

		// Rate-limit signals between pages [§9.5.2]
		time.Sleep(batchSignalRateDelay)

		if len(rows) < batchPageSize {
			break // Last page
		}
	}

	result.PoliciesScanned = totalScanned
	result.TransitionsApplied = totalTransitioned
	result.CompletedAt = time.Now().UTC()
	result.Status = domain.BatchScanStatusCompleted
	return result, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// RemissionExpiryShortScanActivity — VOID_LAPSE → VOID (policy_life < 36 months)
// [FR-PM-012, §9.3]
// ─────────────────────────────────────────────────────────────────────────────

// RemissionExpiryShortScanActivity transitions VOID_LAPSE → VOID for policies
// where the 12-month remission period has expired and policy_life < 36 months.
// Database-first pattern. [FR-PM-012, §9.5.2]
func (a *BatchActivities) RemissionExpiryShortScanActivity(ctx context.Context, scheduledDate time.Time) (BatchScanResult, error) {
	return a.runRemissionExpiryScan(ctx, scheduledDate,
		domain.BatchScanTypeRemissionExpiryShort,
		domain.StatusVoidLapse, domain.StatusVoid,
		"remission expired — short policy (<36mo)", false)
}

// ─────────────────────────────────────────────────────────────────────────────
// RemissionExpiryLongScanActivity — INACTIVE_LAPSE → ACTIVE_LAPSE (≥ 36 months)
// [FR-PM-013, §9.3]
// ─────────────────────────────────────────────────────────────────────────────

// RemissionExpiryLongScanActivity transitions INACTIVE_LAPSE → ACTIVE_LAPSE for
// policies where the 12-month remission period has expired and policy_life ≥ 36 months.
// Database-first pattern. [FR-PM-013, §9.5.2]
func (a *BatchActivities) RemissionExpiryLongScanActivity(ctx context.Context, scheduledDate time.Time) (BatchScanResult, error) {
	return a.runRemissionExpiryScan(ctx, scheduledDate,
		domain.BatchScanTypeRemissionExpiryLong,
		domain.StatusInactiveLapse, domain.StatusActiveLapse,
		"remission expired — long policy (≥36mo)", true)
}

// ─────────────────────────────────────────────────────────────────────────────
// PaidUpConversionScanActivity — ACTIVE_LAPSE → PAID_UP or VOID (monthly)
// [FR-PM-014, §9.3, BR-PM-060, BR-PM-061]
// ─────────────────────────────────────────────────────────────────────────────

// PaidUpConversionScanActivity evaluates ACTIVE_LAPSE policies for paid-up conversion.
// Eligibility: premiums_paid_months >= 36 AND policy_life >= 3 years.
// Outcome: if PU value >= Rs.10,000 → PAID_UP; else → VOID. Monthly schedule.
// Database-first pattern. [FR-PM-014, BR-PM-060, BR-PM-061]
func (a *BatchActivities) PaidUpConversionScanActivity(ctx context.Context, scheduledDate time.Time) (BatchScanResult, error) {
	startedAt := time.Now().UTC()
	result := BatchScanResult{
		ScanType:      domain.BatchScanTypePaidUpConversion,
		ScheduledDate: scheduledDate,
		StartedAt:     startedAt,
	}

	type paidUpRow struct {
		PolicyID         int64     `db:"policy_id"`
		PolicyNumber     string    `db:"policy_number"`
		IssueDate        time.Time `db:"issue_date"`
		SumAssured       float64   `db:"sum_assured"`
		PremiumsPaid     int       `db:"premiums_paid_months"`
		TotalPremiums    int       `db:"total_premiums_months"`
		BonusAccumulated float64   `db:"bonus_accumulated"` // [C9] PLI reversionary bonus
	}

	const minPaidUpValue = 10000.0
	offset := 0

	for {
		activity.RecordHeartbeat(ctx, fmt.Sprintf("paid-up offset=%d", offset))

		// Fetch ACTIVE_LAPSE policies with sufficient paid premiums [BR-PM-061]
		q := dblib.Psql.Select("policy_id", "policy_number", "issue_date", "sum_assured",
			"premiums_paid_months", "total_premiums_months", "bonus_accumulated"). // [C9]
			From(actPolicyTable).
			Where(sq.Eq{"current_status": domain.StatusActiveLapse}).
			Where(sq.GtOrEq{"premiums_paid_months": 36}).
			OrderBy("policy_id").
			Limit(batchPageSize).
			Offset(uint64(offset))

		rows, err := dblib.SelectRows(ctx, a.db, q, pgx.RowToStructByNameLax[paidUpRow])
		if err != nil {
			result.Status = domain.BatchScanStatusFailed
			return result, fmt.Errorf("PaidUpConversionScanActivity fetchPage: %w", err)
		}
		if len(rows) == 0 {
			break
		}

		paidUpIDs := make([]int64, 0)
		voidIDs := make([]int64, 0)
		paidUpMap := make(map[int64]string)

		for _, row := range rows {
			policyLifeMonths := monthsBetween(row.IssueDate, scheduledDate)
			if policyLifeMonths < 36 {
				continue // Not yet eligible [BR-PM-061]
			}
			// [C9] PUSA = (premiums_paid / total_premiums) × (sum_assured + bonus_accumulated)
			// PLI Directorate formula: proportional share of (SA + accrued reversionary bonus).
			// [C9, BR-PM-060, BR-PM-061]
			puValue := 0.0
			if row.TotalPremiums > 0 {
				puValue = (float64(row.PremiumsPaid) / float64(row.TotalPremiums)) *
					(row.SumAssured + row.BonusAccumulated)
			}
			if puValue >= minPaidUpValue {
				paidUpIDs = append(paidUpIDs, row.PolicyID)
				paidUpMap[row.PolicyID] = domain.StatusPaidUp
			} else {
				voidIDs = append(voidIDs, row.PolicyID)
				paidUpMap[row.PolicyID] = domain.StatusVoid
			}
		}

		now := time.Now().UTC()
		if err := a.bulkTransition(ctx, paidUpIDs, domain.StatusActiveLapse, domain.StatusPaidUp,
			"paid-up conversion: value ≥ 10K", scheduledDate, now, nil); err != nil {
			result.Errors++
		}
		if err := a.bulkTransition(ctx, voidIDs, domain.StatusActiveLapse, domain.StatusVoid,
			"paid-up conversion: value < 10K — void [BR-PM-061]", scheduledDate, now, nil); err != nil {
			result.Errors++
		}

		// Signals
		for _, row := range rows {
			newStatus, ok := paidUpMap[row.PolicyID]
			if !ok {
				continue
			}
			_ = a.tc.SignalWorkflow(ctx,
				policyWorkflowIDPrefix+row.PolicyNumber,
				"", batchStateSyncSignal, batchSyncSignal{
					NewStatus:     newStatus,
					ScanType:      domain.BatchScanTypePaidUpConversion,
					ScheduledDate: scheduledDate,
				})
		}

		result.PoliciesScanned += len(rows)
		result.TransitionsApplied += len(paidUpIDs) + len(voidIDs)
		offset += batchPageSize
		time.Sleep(batchSignalRateDelay)

		if len(rows) < batchPageSize {
			break
		}
	}

	result.CompletedAt = time.Now().UTC()
	result.Status = domain.BatchScanStatusCompleted
	return result, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// MaturityScanActivity — ACTIVE → PENDING_MATURITY (within 90 days of maturity)
// [FR-PM-015, §9.3]
// ─────────────────────────────────────────────────────────────────────────────

// MaturityScanActivity transitions ACTIVE → PENDING_MATURITY for policies
// within 90 days of their maturity_date. Daily schedule.
// Database-first pattern. [FR-PM-015, §9.3, §9.5.2]
func (a *BatchActivities) MaturityScanActivity(ctx context.Context, scheduledDate time.Time) (BatchScanResult, error) {
	startedAt := time.Now().UTC()
	result := BatchScanResult{
		ScanType:      domain.BatchScanTypeMaturityScan,
		ScheduledDate: scheduledDate,
		StartedAt:     startedAt,
	}

	type maturityRow struct {
		PolicyID     int64  `db:"policy_id"`
		PolicyNumber string `db:"policy_number"`
	}

	maturityWindow := scheduledDate.Add(90 * 24 * time.Hour)
	offset := 0

	for {
		activity.RecordHeartbeat(ctx, fmt.Sprintf("maturity offset=%d", offset))

		q := dblib.Psql.Select("policy_id", "policy_number").
			From(actPolicyTable).
			Where(sq.Eq{"current_status": domain.StatusActive}).
			Where(sq.LtOrEq{"maturity_date": maturityWindow}).
			Where(sq.GtOrEq{"maturity_date": scheduledDate}).
			OrderBy("policy_id").
			Limit(batchPageSize).
			Offset(uint64(offset))

		rows, err := dblib.SelectRows(ctx, a.db, q, pgx.RowToStructByNameLax[maturityRow])
		if err != nil {
			result.Status = domain.BatchScanStatusFailed
			return result, fmt.Errorf("MaturityScanActivity fetchPage: %w", err)
		}
		if len(rows) == 0 {
			break
		}

		ids := make([]int64, len(rows))
		for i, r := range rows {
			ids[i] = r.PolicyID
		}

		now := time.Now().UTC()
		if err := a.bulkTransition(ctx, ids, domain.StatusActive, domain.StatusPendingMaturity,
			"maturity within 90 days", scheduledDate, now, nil); err != nil {
			result.Errors += len(ids)
		} else {
			result.TransitionsApplied += len(ids)
		}

		for _, row := range rows {
			_ = a.tc.SignalWorkflow(ctx,
				policyWorkflowIDPrefix+row.PolicyNumber,
				"", batchStateSyncSignal, batchSyncSignal{
					NewStatus:     domain.StatusPendingMaturity,
					ScanType:      domain.BatchScanTypeMaturityScan,
					ScheduledDate: scheduledDate,
				})
		}

		result.PoliciesScanned += len(rows)
		offset += batchPageSize
		time.Sleep(batchSignalRateDelay)

		if len(rows) < batchPageSize {
			break
		}
	}

	result.CompletedAt = time.Now().UTC()
	result.Status = domain.BatchScanStatusCompleted
	return result, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// ForcedSurrenderEvalActivity — ASSIGNED_TO_PRESIDENT with loan ≥ 100% GSV
// [FR-PM-015b, §9.3, BR-PM-074]
// Monthly schedule — sends forced-surrender-trigger signal (NOT DB transition)
// ─────────────────────────────────────────────────────────────────────────────

// ForcedSurrenderEvalActivity evaluates ASSIGNED_TO_PRESIDENT policies where the
// outstanding loan balance >= 100% of the surrender value. Sends a
// forced-surrender-trigger signal to each qualifying policy's PLW workflow.
// Monthly schedule. [FR-PM-015b, §9.3]
func (a *BatchActivities) ForcedSurrenderEvalActivity(ctx context.Context, scheduledDate time.Time) (BatchScanResult, error) {
	startedAt := time.Now().UTC()
	result := BatchScanResult{
		ScanType:      domain.BatchScanTypeForcedSurrenderEval,
		ScheduledDate: scheduledDate,
		StartedAt:     startedAt,
	}

	type forcedSurrenderRow struct {
		PolicyID        int64   `db:"policy_id"`
		PolicyNumber    string  `db:"policy_number"`
		LoanOutstanding float64 `db:"loan_outstanding"`
		SumAssured      float64 `db:"sum_assured"`
	}

	offset := 0

	// Fetch loan-to-GSV ratio from config [Review-Fix-8]
	var loanRatioFraction float64 = 1.0 // default: 100%
	{
		type cfgRow struct {
			Value string `db:"config_value"`
		}
		cfgQ := dblib.Psql.Select("config_value").
			From("policy_mgmt.policy_state_config").
			Where(sq.Eq{"config_key": domain.ConfigKeyForcedSurrenderLoanRatioPct}).
			Limit(1)
		if cfgRows, err2 := dblib.SelectRows(ctx, a.db, cfgQ, pgx.RowToStructByNameLax[cfgRow]); err2 == nil && len(cfgRows) > 0 {
			if v, err3 := strconv.ParseFloat(cfgRows[0].Value, 64); err3 == nil && v > 0 {
				loanRatioFraction = v / 100.0
			}
		}
	}
	for {
		activity.RecordHeartbeat(ctx, fmt.Sprintf("forced-surrender offset=%d", offset))

		// Select ASSIGNED_TO_PRESIDENT with active loan where loan ≥ 100% sum_assured
		// (sum_assured used as proxy for 100% GSV — actual GSV computation is downstream) [BR-PM-074, Review-Fix-18]
		q := dblib.Psql.Select("policy_id", "policy_number", "loan_outstanding", "sum_assured").
			From(actPolicyTable).
			Where(sq.Eq{"current_status": domain.StatusAssignedToPresident}).
			Where(sq.Eq{"has_active_loan": true}).
			Where(sq.Expr("loan_outstanding >= sum_assured * ?", loanRatioFraction)). // [Review-Fix-8]
			OrderBy("policy_id").
			Limit(batchPageSize).
			Offset(uint64(offset))

		rows, err := dblib.SelectRows(ctx, a.db, q, pgx.RowToStructByNameLax[forcedSurrenderRow])
		if err != nil {
			result.Status = domain.BatchScanStatusFailed
			return result, fmt.Errorf("ForcedSurrenderEvalActivity fetchPage: %w", err)
		}
		if len(rows) == 0 {
			break
		}

		// [C8] For each DB pre-filtered candidate, verify against actual GSV from surrender-svc.
		// The DB pre-filter (sum_assured proxy) is conservative — actual GSV may be lower.
		// Only signal PLW if loan_outstanding >= actual GSV * threshold. [FR-PM-015b, C8]
		actLogger := activity.GetLogger(ctx)
		for _, row := range rows {
			gsv, err := a.fetchGSVFromSurrenderSvc(ctx, row.PolicyNumber)
			if err != nil {
				// Skip — safer to miss a signal than to trigger forced surrender incorrectly. [C8]
				actLogger.Warn("ForcedSurrenderEvalActivity: GSV lookup failed — policy skipped",
					"policyNumber", row.PolicyNumber, "error", err)
				continue
			}

			// Apply actual GSV threshold check [C8, BR-PM-074]
			if row.LoanOutstanding < gsv*loanRatioFraction {
				continue // loan < actual GSV threshold — pre-filter was too aggressive
			}

			payload := map[string]interface{}{
				"trigger_reason":   "loan_balance_exceeds_gsv",
				"loan_outstanding": row.LoanOutstanding,
				"gsv":              gsv, // actual GSV now included [C8]
				"scheduled_date":   scheduledDate.Format("2006-01-02"),
			}
			_ = a.tc.SignalWorkflow(ctx,
				policyWorkflowIDPrefix+row.PolicyNumber,
				"", forcedSurrenderSignal, payload)
		}

		result.PoliciesScanned += len(rows)
		result.TransitionsApplied += len(rows) // signals sent
		offset += batchPageSize
		time.Sleep(batchSignalRateDelay)

		if len(rows) < batchPageSize {
			break
		}
	}

	result.CompletedAt = time.Now().UTC()
	result.Status = domain.BatchScanStatusCompleted
	return result, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// RecordBatchScanResultActivity — INSERT/UPDATE batch_scan_state [§8.6]
// ─────────────────────────────────────────────────────────────────────────────

// RecordBatchScanResultActivity records the outcome of a batch scan job in
// batch_scan_state. ON CONFLICT (scan_type, scheduled_date) DO UPDATE for
// idempotency — Temporal retries will update the same row. [§8.6, §9.3]
func (a *BatchActivities) RecordBatchScanResultActivity(ctx context.Context, r BatchScanResult) error {
	ctx, cancel := context.WithTimeout(ctx, a.cfg.GetDuration("db.QueryTimeoutHigh"))
	defer cancel()

	durSec := int(r.CompletedAt.Sub(r.StartedAt).Seconds())

	q := dblib.Psql.Insert("policy_mgmt.batch_scan_state").
		Columns(
			"scan_type", "scheduled_date",
			"started_at", "completed_at",
			"policies_scanned", "transitions_applied", "errors",
			"status", "duration_seconds",
		).
		Values(
			r.ScanType, r.ScheduledDate,
			r.StartedAt, r.CompletedAt,
			r.PoliciesScanned, r.TransitionsApplied, r.Errors,
			r.Status, durSec,
		).
		Suffix(`ON CONFLICT (scan_type, scheduled_date) DO UPDATE
			SET started_at          = EXCLUDED.started_at,
			    completed_at        = EXCLUDED.completed_at,
			    policies_scanned    = EXCLUDED.policies_scanned,
			    transitions_applied = EXCLUDED.transitions_applied,
			    errors              = EXCLUDED.errors,
			    status              = EXCLUDED.status,
			    duration_seconds    = EXCLUDED.duration_seconds
			RETURNING scan_id`)

	type scanIDRow struct {
		ScanID int64 `db:"scan_id"`
	}
	if _, err := dblib.InsertReturning(ctx, a.db, q, pgx.RowToStructByNameLax[scanIDRow]); err != nil {
		return fmt.Errorf("RecordBatchScanResultActivity scanType=%s date=%s: %w",
			r.ScanType, r.ScheduledDate.Format("2006-01-02"), err)
	}
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Private helpers
// ─────────────────────────────────────────────────────────────────────────────

// runRemissionExpiryScan is the shared implementation for VOID_LAPSE→VOID and
// INACTIVE_LAPSE→ACTIVE_LAPSE remission expiry scans. [FR-PM-012, FR-PM-013]
func (a *BatchActivities) runRemissionExpiryScan(
	ctx context.Context,
	scheduledDate time.Time,
	scanType, fromStatus, toStatus, reason string,
	longPolicy bool, // true for ≥36mo (INACTIVE_LAPSE→ACTIVE_LAPSE), false for short
) (BatchScanResult, error) {
	startedAt := time.Now().UTC()
	result := BatchScanResult{
		ScanType:      scanType,
		ScheduledDate: scheduledDate,
		StartedAt:     startedAt,
	}

	type remissionRow struct {
		PolicyID     int64  `db:"policy_id"`
		PolicyNumber string `db:"policy_number"`
	}

	offset := 0
	for {
		activity.RecordHeartbeat(ctx, fmt.Sprintf("%s offset=%d", scanType, offset))

		// Policies whose remission period has expired [§9.3]
		q := dblib.Psql.Select("policy_id", "policy_number").
			From(actPolicyTable).
			Where(sq.Eq{"current_status": fromStatus}).
			Where(sq.LtOrEq{"remission_expiry_date": scheduledDate}).
			OrderBy("policy_id").
			Limit(batchPageSize).
			Offset(uint64(offset))

		// Filter by policy_life — long scan needs ≥36mo, short needs <36mo
		// policy_life = scheduledDate - issue_date (approximate via months)
		if longPolicy {
			q = q.Where(sq.Expr(
				"EXTRACT(MONTH FROM AGE(?, issue_date)) >= 36",
				scheduledDate,
			))
		} else {
			q = q.Where(sq.Expr(
				"EXTRACT(MONTH FROM AGE(?, issue_date)) < 36",
				scheduledDate,
			))
		}

		rows, err := dblib.SelectRows(ctx, a.db, q, pgx.RowToStructByNameLax[remissionRow])
		if err != nil {
			result.Status = domain.BatchScanStatusFailed
			return result, fmt.Errorf("%s fetchPage: %w", scanType, err)
		}
		if len(rows) == 0 {
			break
		}

		ids := make([]int64, len(rows))
		for i, r := range rows {
			ids[i] = r.PolicyID
		}

		now := time.Now().UTC()
		if err := a.bulkTransition(ctx, ids, fromStatus, toStatus, reason, scheduledDate, now, nil); err != nil {
			result.Errors += len(ids)
		} else {
			result.TransitionsApplied += len(ids)
		}

		for _, row := range rows {
			_ = a.tc.SignalWorkflow(ctx,
				policyWorkflowIDPrefix+row.PolicyNumber,
				"", batchStateSyncSignal, batchSyncSignal{
					NewStatus:     toStatus,
					ScanType:      scanType,
					ScheduledDate: scheduledDate,
				})
		}

		result.PoliciesScanned += len(rows)
		offset += batchPageSize
		time.Sleep(batchSignalRateDelay)

		if len(rows) < batchPageSize {
			break
		}
	}

	result.CompletedAt = time.Now().UTC()
	result.Status = domain.BatchScanStatusCompleted
	return result, nil
}

// bulkTransition performs a bulk UPDATE policy + bulk INSERT policy_status_history
// for the given policy IDs transitioning from→to status.
// Uses pgx.Batch + db.SendBatch for atomic submission. [§9.5.2, Constraint 5]
func (a *BatchActivities) bulkTransition(
	ctx context.Context,
	policyIDs []int64,
	fromStatus, toStatus, reason string,
	scheduledDate, now time.Time,
	remissionExpiry *time.Time,
) error {
	if len(policyIDs) == 0 {
		return nil
	}

	batch := &pgx.Batch{}

	// Bulk UPDATE policy status — WHERE current_status guard ensures idempotency [§9.5.2]
	uq := dblib.Psql.Update(actPolicyTable).
		Set("previous_status", fromStatus).
		Set("current_status", toStatus).
		Set("display_status", toStatus).
		Set("effective_from", now). // [Review-Fix-15]: track transition time
		Set("version", sq.Expr("version + 1")).
		Set("updated_at", now).
		Where(sq.Eq{"policy_id": policyIDs}).
		Where(sq.Eq{"current_status": fromStatus}) // optimistic guard — skips if already transitioned
	if remissionExpiry != nil {
		uq = uq.Set("remission_expiry_date", *remissionExpiry)
	}
	dblib.QueueExecRow(batch, uq)

	// Bulk INSERT policy_status_history — one INSERT per policy for partition key compliance [§8.2]
	for _, id := range policyIDs {
		hq := dblib.Psql.Insert(actPolicyHistTable).
			Columns("policy_id", "from_status", "to_status", "transition_reason",
				"triggered_by_service", "effective_date", "created_at").
			Values(id, fromStatus, toStatus, reason, "batch-scan", now, now).
			Suffix("ON CONFLICT DO NOTHING") // idempotent on retry
		dblib.QueueExecRow(batch, hq)
	}

	if err := a.db.SendBatch(ctx, batch).Close(); err != nil {
		return fmt.Errorf("bulkTransition %s→%s count=%d: %w", fromStatus, toStatus, len(policyIDs), err)
	}
	return nil
}

// monthsBetween returns the number of complete calendar months between from and to.
// More precise than the (hours / 720) approximation — accounts for varying month lengths.
// [Review-Fix-6]
func monthsBetween(from, to time.Time) int {
	years := to.Year() - from.Year()
	months := int(to.Month()) - int(from.Month())
	total := years*12 + months
	// Subtract one month if we haven't reached the anniversary day-of-month yet
	if to.Day() < from.Day() {
		total--
	}
	if total < 0 {
		return 0
	}
	return total
}

// ─────────────────────────────────────────────────────────────────────────────
// Remission slab helpers — C6
// Mirror DB functions: last_day_of_month(), compute_remission_expiry()
// ─────────────────────────────────────────────────────────────────────────────

// lastDayOfMonth returns the last calendar day of the month containing t.
// Mirrors DB: DATE_TRUNC('month', p_date) + INTERVAL '1 month' - INTERVAL '1 day'. [C6]
func lastDayOfMonth(t time.Time) time.Time {
	firstOfNext := time.Date(t.Year(), t.Month()+1, 1, 0, 0, 0, 0, t.Location())
	return firstOfNext.Add(-24 * time.Hour)
}

// computeRemissionExpiry mirrors DB compute_remission_expiry(). [C6]
// paidToDate maps to first_unpaid_date in the DB function.
// Returns nil for policy_life < 6 months (VOID immediately — no remission period).
//
// DB logic:
//
//	v_grace_end := last_day_of_month(p_first_unpaid)
//	< 6mo  → NULL
//	< 12mo → grace_end + 30 days  (VOID_LAPSE)
//	< 24mo → grace_end + 60 days  (VOID_LAPSE)
//	< 36mo → grace_end + 90 days  (VOID_LAPSE)
//	≥ 36mo → first_unpaid + 12 months (INACTIVE_LAPSE)
func computeRemissionExpiry(issueDate, paidToDate, scheduledDate time.Time) *time.Time {
	life := monthsBetween(issueDate, scheduledDate)
	if life < 6 {
		return nil
	}
	graceEnd := lastDayOfMonth(paidToDate)
	var expiry time.Time
	switch {
	case life < 12:
		expiry = graceEnd.AddDate(0, 0, 30)
	case life < 24:
		expiry = graceEnd.AddDate(0, 0, 60)
	case life < 36:
		expiry = graceEnd.AddDate(0, 0, 90)
	default:
		// ≥36mo: first_unpaid_date + 12 months (INACTIVE_LAPSE) [C6]
		expiry = paidToDate.AddDate(0, 12, 0)
	}
	return &expiry
}

// policyRemissionPair pairs a policy ID with its per-policy remission expiry. [C6]
type policyRemissionPair struct {
	PolicyID        int64
	RemissionExpiry time.Time
}

// bulkTransitionWithRemissions performs bulk UPDATE + history INSERT with a per-policy
// remission_expiry_date. Uses pgx.Batch of individual UPDATE statements so each
// row gets its own slab-computed expiry. Same atomic-submission pattern as bulkTransition. [C6]
func (a *BatchActivities) bulkTransitionWithRemissions(
	ctx context.Context,
	pairs []policyRemissionPair,
	fromStatus, toStatus, reason string,
	scheduledDate, now time.Time,
) error {
	if len(pairs) == 0 {
		return nil
	}

	batch := &pgx.Batch{}

	for _, p := range pairs {
		expiry := p.RemissionExpiry
		uq := dblib.Psql.Update(actPolicyTable).
			Set("previous_status", fromStatus).
			Set("current_status", toStatus).
			Set("display_status", toStatus).
			Set("remission_expiry_date", expiry).
			Set("effective_from", now).
			Set("version", sq.Expr("version + 1")).
			Set("updated_at", now).
			Where(sq.Eq{"policy_id": p.PolicyID}).
			Where(sq.Eq{"current_status": fromStatus}) // optimistic guard — idempotent on retry
		dblib.QueueExecRow(batch, uq)

		hq := dblib.Psql.Insert(actPolicyHistTable).
			Columns("policy_id", "from_status", "to_status", "transition_reason",
				"triggered_by_service", "effective_date", "created_at").
			Values(p.PolicyID, fromStatus, toStatus, reason, "batch-scan", now, now).
			Suffix("ON CONFLICT DO NOTHING")
		dblib.QueueExecRow(batch, hq)
	}

	if err := a.db.SendBatch(ctx, batch).Close(); err != nil {
		return fmt.Errorf("bulkTransitionWithRemissions %s→%s count=%d: %w",
			fromStatus, toStatus, len(pairs), err)
	}
	return nil
}

// fetchGSVFromSurrenderSvc calls the surrender-svc internal quote endpoint to get
// the current Gross Surrender Value for the given policy.
// Used by ForcedSurrenderEvalActivity to replace the sum_assured proxy. [C8, FR-PM-015b]
func (a *BatchActivities) fetchGSVFromSurrenderSvc(ctx context.Context, policyNumber string) (float64, error) {
	baseURL := a.cfg.GetString("services.surrender_svc.internal_url")
	if baseURL == "" {
		return 0, fmt.Errorf("fetchGSVFromSurrenderSvc: services.surrender_svc.internal_url not configured")
	}

	reqURL := fmt.Sprintf("%s/internal/v1/policies/%s/surrender-quote",
		baseURL, url.PathEscape(policyNumber))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return 0, fmt.Errorf("fetchGSVFromSurrenderSvc: build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "policy-management/1.0")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return 0, fmt.Errorf("fetchGSVFromSurrenderSvc GET %s: %w", reqURL, err)
	}
	defer resp.Body.Close()

	// Limit to 1 MiB — prevents memory exhaustion from runaway upstream response. [D12]
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return 0, fmt.Errorf("fetchGSVFromSurrenderSvc read body: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return 0, fmt.Errorf("fetchGSVFromSurrenderSvc HTTP %d: %s", resp.StatusCode, string(body))
	}

	// Same JSON shape as QuoteActivities.GetSurrenderQuoteActivity — reuse field [C8]
	var result struct {
		GrossSurrenderValue float64 `json:"gross_surrender_value"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return 0, fmt.Errorf("fetchGSVFromSurrenderSvc decode: %w", err)
	}
	return result.GrossSurrenderValue, nil
}

// fetchLapsationCandidates returns ACTIVE policies eligible for lapsation. [FR-PM-011]
func (a *BatchActivities) fetchLapsationCandidates(ctx context.Context, scheduledDate time.Time, offset int) ([]batchPolicyRow, error) {
	// Paid-to-date < (scheduled_date - 1 month) means premium for current period not paid [FR-PM-011]
	cutoffDate := scheduledDate.Add(-30 * 24 * time.Hour)

	// remission_expiry_date and premiums_paid_months are intentionally excluded:
	// remission is recomputed in-process via computeRemissionExpiry() [C6],
	// and premiums_paid_months is not used in the lapsation scan after C6. [D3]
	q := dblib.Psql.Select(
		"policy_id", "policy_number", "issue_date", "paid_to_date",
	).
		From(actPolicyTable).
		Where(sq.Eq{"current_status": domain.StatusActive}).
		Where(sq.Lt{"paid_to_date": cutoffDate}).
		// Skip pay-recovery policies within 12-month active protection [BR-PM-074]
		Where(sq.Or{
			sq.NotEq{"billing_method": "PAY_RECOVERY"},
			sq.Lt{"pay_recovery_protection_expiry": scheduledDate},
		}).
		OrderBy("policy_id").
		Limit(batchPageSize).
		Offset(uint64(offset))

	rows, err := dblib.SelectRows(ctx, a.db, q, pgx.RowToStructByNameLax[batchPolicyRow])
	if err != nil {
		return nil, err
	}
	return rows, nil
}

