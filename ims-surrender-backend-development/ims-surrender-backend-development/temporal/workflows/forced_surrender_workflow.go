package workflows

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"gitlab.cept.gov.in/it-2.0-policy/surrender-service/temporal/activities"
)

// ForcedSurrenderWorkflowInput defines the input for forced surrender workflow
// TEMP-002: Forced Surrender Evaluation Workflow
type ForcedSurrenderWorkflowInput struct {
	EvaluationDate string
	BatchSize      int
}

// ForcedSurrenderWorkflowResult defines the result of forced surrender workflow
type ForcedSurrenderWorkflowResult struct {
	EvaluationDate      string
	PoliciesEvaluated   int
	RemindersCreated    int
	SurrendersInitiated int
	CompletedAt         time.Time
}

// ForcedSurrenderWorkflow orchestrates monthly forced surrender evaluation
// TEMP-002: Forced Surrender Evaluation Workflow
// Business Flow:
// 1. Identify eligible policies (6+ months unpaid)
// 2. Create reminders in batches
// 3. Monitor payment windows
// 4. Initiate forced surrender for expired windows
func ForcedSurrenderWorkflow(ctx workflow.Context, input ForcedSurrenderWorkflowInput) (*ForcedSurrenderWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting Forced Surrender Evaluation Workflow", "EvaluationDate", input.EvaluationDate)

	result := &ForcedSurrenderWorkflowResult{
		EvaluationDate: input.EvaluationDate,
	}

	// Configure activity options
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts:    3,
			InitialInterval:    2 * time.Second,
			BackoffCoefficient: 2.0,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	// Step 1: Identify Eligible Policies
	logger.Info("Step 1: Identifying eligible policies")
	var identifyResult activities.IdentifyEligiblePoliciesResult
	err := workflow.ExecuteActivity(ctx, activities.IdentifyEligiblePoliciesActivity, activities.IdentifyEligiblePoliciesInput{
		EvaluationDate:  input.EvaluationDate,
		MinUnpaidMonths: 6,
	}).Get(ctx, &identifyResult)

	if err != nil {
		logger.Error("Failed to identify eligible policies", "error", err)
		return result, err
	}

	result.PoliciesEvaluated = len(identifyResult.EligiblePolicies)
	logger.Info("Found eligible policies", "count", result.PoliciesEvaluated)

	if len(identifyResult.EligiblePolicies) == 0 {
		logger.Info("No eligible policies found")
		result.CompletedAt = workflow.Now(ctx)
		return result, nil
	}

	// Step 2: Create Reminders in Batches
	logger.Info("Step 2: Creating reminders in batches")
	batchSize := 50
	if input.BatchSize > 0 {
		batchSize = input.BatchSize
	}

	var totalRemindersCreated int
	for i := 0; i < len(identifyResult.EligiblePolicies); i += batchSize {
		end := i + batchSize
		if end > len(identifyResult.EligiblePolicies) {
			end = len(identifyResult.EligiblePolicies)
		}

		batch := identifyResult.EligiblePolicies[i:end]
		logger.Info("Processing batch", "batch", i/batchSize+1, "size", len(batch))

		var batchResult activities.CreateRemindersBatchResult
		err := workflow.ExecuteActivity(ctx, activities.CreateRemindersBatchActivity, activities.CreateRemindersBatchInput{
			Policies: batch,
		}).Get(ctx, &batchResult)

		if err != nil {
			logger.Error("Failed to create reminders batch", "error", err, "batch", i/batchSize+1)
			continue
		}

		totalRemindersCreated += batchResult.RemindersCreated
		logger.Info("Batch completed", "reminders_created", batchResult.RemindersCreated)
	}

	result.RemindersCreated = totalRemindersCreated

	// Step 3: Schedule Payment Window Monitoring
	logger.Info("Step 3: Scheduling payment window monitoring")

	// Start child workflow to monitor payment windows
	childWorkflowOptions := workflow.ChildWorkflowOptions{
		WorkflowID:               "payment-window-monitor-" + input.EvaluationDate,
		WorkflowExecutionTimeout: 90 * 24 * time.Hour, // 90 days
	}
	childCtx := workflow.WithChildOptions(ctx, childWorkflowOptions)

	_ = workflow.ExecuteChildWorkflow(childCtx, PaymentWindowMonitorWorkflow, PaymentWindowMonitorInput{
		StartDate: input.EvaluationDate,
	})

	// Don't wait for child workflow - it runs independently
	logger.Info("Payment window monitor workflow started")

	// Step 4: Check for Expired Windows (for current batch)
	logger.Info("Step 4: Checking for expired payment windows")
	var expiredResult activities.CheckExpiredPaymentWindowsResult
	err = workflow.ExecuteActivity(ctx, activities.CheckExpiredPaymentWindowsActivity, activities.CheckExpiredPaymentWindowsInput{
		AsOfDate: input.EvaluationDate,
	}).Get(ctx, &expiredResult)

	if err != nil {
		logger.Error("Failed to check expired windows", "error", err)
		// Don't fail workflow
	} else if len(expiredResult.ExpiredWindows) > 0 {
		logger.Info("Found expired payment windows", "count", len(expiredResult.ExpiredWindows))

		// Step 5: Initiate Forced Surrenders for Expired Windows
		var initiateResult activities.InitiateForcedSurrendersBatchResult
		err = workflow.ExecuteActivity(ctx, activities.InitiateForcedSurrendersBatchActivity, activities.InitiateForcedSurrendersBatchInput{
			ExpiredWindows: expiredResult.ExpiredWindows,
		}).Get(ctx, &initiateResult)

		if err != nil {
			logger.Error("Failed to initiate forced surrenders", "error", err)
		} else {
			result.SurrendersInitiated = initiateResult.SurrendersInitiated
			logger.Info("Forced surrenders initiated", "count", result.SurrendersInitiated)
		}
	}

	logger.Info("Forced Surrender Evaluation Workflow completed")
	result.CompletedAt = workflow.Now(ctx)

	return result, nil
}

// PaymentWindowMonitorInput defines input for payment window monitoring
type PaymentWindowMonitorInput struct {
	StartDate string
}

// PaymentWindowMonitorWorkflow monitors payment windows and initiates forced surrender
// This runs as a long-running child workflow
func PaymentWindowMonitorWorkflow(ctx workflow.Context, input PaymentWindowMonitorInput) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting Payment Window Monitor Workflow")

	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	// Monitor every day for 90 days
	for day := 0; day < 90; day++ {
		// Check for expired windows
		var expiredResult activities.CheckExpiredPaymentWindowsResult
		err := workflow.ExecuteActivity(ctx, activities.CheckExpiredPaymentWindowsActivity, activities.CheckExpiredPaymentWindowsInput{
			AsOfDate: input.StartDate,
		}).Get(ctx, &expiredResult)

		if err != nil {
			logger.Error("Failed to check expired windows", "error", err, "day", day)
		} else if len(expiredResult.ExpiredWindows) > 0 {
			logger.Info("Found expired windows", "count", len(expiredResult.ExpiredWindows), "day", day)

			// Initiate forced surrenders
			var initiateResult activities.InitiateForcedSurrendersBatchResult
			err = workflow.ExecuteActivity(ctx, activities.InitiateForcedSurrendersBatchActivity, activities.InitiateForcedSurrendersBatchInput{
				ExpiredWindows: expiredResult.ExpiredWindows,
			}).Get(ctx, &initiateResult)

			if err != nil {
				logger.Error("Failed to initiate forced surrenders", "error", err)
			} else {
				logger.Info("Initiated forced surrenders", "count", initiateResult.SurrendersInitiated)
			}
		}

		// Sleep for 24 hours
		_ = workflow.Sleep(ctx, 24*time.Hour)
	}

	logger.Info("Payment Window Monitor Workflow completed")
	return nil
}
