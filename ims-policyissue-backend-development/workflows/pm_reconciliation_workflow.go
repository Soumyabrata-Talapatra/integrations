package workflows

import (
	"fmt"
	"time"

	"policy-issue-service/workflows/activities"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

const (
	// PMReconciliationTaskQueue is the queue this workflow runs on (same as the main worker).
	PMReconciliationTaskQueue = "policy-issue-queue"

	// PMReconciliationWorkflowID is the stable ID used when scheduling this workflow.
	PMReconciliationWorkflowID = "pm-signal-reconciliation"

	// pmGracePeriodMinutes — PENDING rows older than this are considered stuck.
	pmGracePeriodMinutes = 30

	// pmMaxAttempts — stop retrying a policy after this many total attempts.
	pmMaxAttempts = 20
)

// PMSignalReconciliationWorkflow is intended to run on a Temporal schedule
// (every 15 minutes). It finds policies whose PM signal was never confirmed
// and re-issues SignalWithStart for each one.
//
// Register this workflow in the worker and create a Temporal schedule pointing
// at PMReconciliationWorkflowID with spec "*/15 * * * *".
func PMSignalReconciliationWorkflow(ctx workflow.Context) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("PMSignalReconciliationWorkflow: starting run")

	queryOpts := workflow.ActivityOptions{
		StartToCloseTimeout: 2 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 3,
		},
	}
	queryCtx := workflow.WithActivityOptions(ctx, queryOpts)

	// Step 1: Discover policies that need retry.
	findInput := activities.FindUnsignalledPoliciesInput{
		GracePeriodMinutes: pmGracePeriodMinutes,
		MaxAttempts:        pmMaxAttempts,
	}
	var targets []activities.StartPMLifecycleInput
	if err := workflow.ExecuteActivity(queryCtx,
		"FindUnsignalledPoliciesActivity", findInput,
	).Get(ctx, &targets); err != nil {
		return fmt.Errorf("PMSignalReconciliationWorkflow: failed to find targets: %w", err)
	}

	if len(targets) == 0 {
		logger.Info("PMSignalReconciliationWorkflow: nothing to reconcile")
		return nil
	}

	logger.Info("PMSignalReconciliationWorkflow: reconciling policies", "count", len(targets))

	signalOpts := workflow.ActivityOptions{
		StartToCloseTimeout: 1 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    2 * time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    30 * time.Second,
			// Only 1 attempt here — the outer reconciliation schedule is the retry loop.
			MaximumAttempts: 1,
		},
	}
	signalCtx := workflow.WithActivityOptions(ctx, signalOpts)

	// Step 2: Fire SignalWithStart for each policy.
	// Failures are logged but do not abort the loop — each policy is independent.
	var failed int
	for _, target := range targets {
		err := workflow.ExecuteActivity(signalCtx,
			"StartPMLifecycleActivity", target,
		).Get(ctx, nil)
		if err != nil {
			logger.Error("PMSignalReconciliationWorkflow: signal failed",
				"policy_number", target.PolicyNumber,
				"error", err,
			)
			failed++
		}
	}

	logger.Info("PMSignalReconciliationWorkflow: run complete",
		"total", len(targets),
		"failed", failed,
		"succeeded", len(targets)-failed,
	)
	return nil
}
