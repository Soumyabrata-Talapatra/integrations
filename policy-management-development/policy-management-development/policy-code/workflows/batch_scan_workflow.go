package workflows

// ============================================================================
// BatchStateScanWorkflow — Batch state transition workflow
//
// Trigger: 6 separate Temporal Schedules via ScheduleClient().Create()
//          (NOT CronSchedule workflow property — Constraint 5)
// Task Queue: policy-management-tq (Constraint 3)
// Namespace:  pli-insurance (Constraint 4)
//
// Schedule IDs and cron expressions (IST = UTC+5:30):
//   batch-lapsation-daily          → LAPSATION             "30 0 * * *"
//   batch-remission-short-daily    → REMISSION_EXPIRY_SHORT "35 0 * * *"
//   batch-remission-long-daily     → REMISSION_EXPIRY_LONG  "40 0 * * *"
//   batch-paidup-monthly           → PAID_UP_CONVERSION     "0 1 1 * *"
//   batch-maturity-daily           → MATURITY_SCAN          "0 2 * * *"
//   batch-forced-surrender-monthly → FORCED_SURRENDER_EVAL  "0 3 1 * *"
//
// [FR-PM-011..FR-PM-015, Constraint 3, Constraint 4, Constraint 5, §9.3]
// ============================================================================

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"policy-management/core/domain"
	acts "policy-management/workflows/activities"
)

// batchActCtx wraps ctx with activity options for long-running batch scan activities.
// StartToCloseTimeout: 2h, HeartbeatTimeout: 5m, RetryPolicy: 3× 2× backoff. [§9.3]
func batchActCtx(ctx workflow.Context) workflow.Context {
	return workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 2 * time.Hour,
		HeartbeatTimeout:    5 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts:    3,
			InitialInterval:    30 * time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    5 * time.Minute,
		},
	})
}

// batchActs is a zero-value instance of BatchActivities used for activity
// function references in workflow.ExecuteActivity calls. [§9.3]
var batchActs acts.BatchActivities

// BatchStateScanWorkflow is the per-scan-type batch state transition workflow.
// It receives a scanType string and scheduledDate, dispatches to the appropriate
// scan activity, then records the result. [FR-PM-011..FR-PM-015, Constraint 5, §9.3]
//
// One schedule per scan type — 6 schedules total — all on policy-management-tq.
// The workflow is short-lived: runs one activity and records the result.
func BatchStateScanWorkflow(ctx workflow.Context, scanType string, scheduledDate time.Time) error {
	var result acts.BatchScanResult
	var err error

	// Capture dispatch time at workflow entry — used in failure path to report
	// accurate StartedAt rather than the (much later) failure time. [D14]
	workflowStartedAt := workflow.Now(ctx)

	// When triggered by a Temporal Schedule the Args slot holds time.Time{} so the
	// workflow derives the scan date from its own execution time. Manual/backfill
	// triggers can pass an explicit non-zero date. [Constraint 5, §9.3]
	if scheduledDate.IsZero() {
		scheduledDate = workflowStartedAt.UTC()
	}

	// Dispatch to the appropriate scan activity based on scanType [Constraint 5]
	switch scanType {
	case domain.BatchScanTypeLapsation:
		// Daily 00:30 IST — ACTIVE → VOID_LAPSE / INACTIVE_LAPSE / VOID [FR-PM-011]
		err = workflow.ExecuteActivity(batchActCtx(ctx),
			batchActs.LapsationScanActivity,
			scheduledDate,
		).Get(ctx, &result)

	case domain.BatchScanTypeRemissionExpiryShort:
		// Daily 00:35 IST — VOID_LAPSE → VOID (policy < 36 months) [FR-PM-012]
		err = workflow.ExecuteActivity(batchActCtx(ctx),
			batchActs.RemissionExpiryShortScanActivity,
			scheduledDate,
		).Get(ctx, &result)

	case domain.BatchScanTypeRemissionExpiryLong:
		// Daily 00:40 IST — INACTIVE_LAPSE → ACTIVE_LAPSE (policy ≥ 36 months) [FR-PM-013]
		err = workflow.ExecuteActivity(batchActCtx(ctx),
			batchActs.RemissionExpiryLongScanActivity,
			scheduledDate,
		).Get(ctx, &result)

	case domain.BatchScanTypePaidUpConversion:
		// Monthly 01:00 IST 1st — ACTIVE_LAPSE → PAID_UP or VOID [FR-PM-014]
		err = workflow.ExecuteActivity(batchActCtx(ctx),
			batchActs.PaidUpConversionScanActivity,
			scheduledDate,
		).Get(ctx, &result)

	case domain.BatchScanTypeMaturityScan:
		// Daily 02:00 IST — ACTIVE → PENDING_MATURITY (within 90 days of maturity) [FR-PM-015]
		err = workflow.ExecuteActivity(batchActCtx(ctx),
			batchActs.MaturityScanActivity,
			scheduledDate,
		).Get(ctx, &result)

	case domain.BatchScanTypeForcedSurrenderEval:
		// Monthly 03:00 IST 1st — ASSIGNED_TO_PRESIDENT with loan ≥ 100% GSV [FR-PM-015b]
		err = workflow.ExecuteActivity(batchActCtx(ctx),
			batchActs.ForcedSurrenderEvalActivity,
			scheduledDate,
		).Get(ctx, &result)

	default:
		return fmt.Errorf("BatchStateScanWorkflow: unknown scanType=%q [Constraint 5]", scanType)
	}

	if err != nil {
		// Record FAILED status before returning error.
		// StartedAt uses workflowStartedAt (captured at entry) not workflow.Now(ctx) here,
		// because Now() at this point is the failure time — not the dispatch time. [D14]
		failResult := acts.BatchScanResult{
			ScanType:      scanType,
			ScheduledDate: scheduledDate,
			StartedAt:     workflowStartedAt,
			CompletedAt:   workflow.Now(ctx),
			Status:        domain.BatchScanStatusFailed,
			Errors:        1,
		}
		// Best-effort record — ignore error
		_ = workflow.ExecuteActivity(shortActCtx(ctx),
			batchActs.RecordBatchScanResultActivity,
			failResult,
		).Get(ctx, nil)
		return fmt.Errorf("BatchStateScanWorkflow scanType=%s: %w", scanType, err)
	}

	// Record successful scan result [§8.6]
	if recordErr := workflow.ExecuteActivity(shortActCtx(ctx),
		batchActs.RecordBatchScanResultActivity,
		result,
	).Get(ctx, nil); recordErr != nil {
		// Non-fatal: scan data is lost but transitions already committed to DB
		return fmt.Errorf("BatchStateScanWorkflow RecordResult scanType=%s: %w", scanType, recordErr)
	}

	return nil
}
