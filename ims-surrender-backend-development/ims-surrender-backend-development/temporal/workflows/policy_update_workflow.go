package workflows

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"gitlab.cept.gov.in/it-2.0-policy/surrender-service/temporal/activities"
)

// PolicyUpdateWorkflowInput defines the input for policy update workflow
// TEMP-006: Policy Status Update Workflow
type PolicyUpdateWorkflowInput struct {
	PolicyID           string
	SurrenderRequestID string
	NewStatus          string
	DispositionType    string
}

// PolicyUpdateWorkflowResult defines the result of policy update workflow
type PolicyUpdateWorkflowResult struct {
	PolicyID       string
	OldStatus      string
	NewStatus      string
	UpdatedAt      time.Time
	RelatedUpdates int
	CompletedAt    time.Time
}

// PolicyUpdateWorkflow orchestrates policy status update process
// TEMP-006: Policy Status Update Workflow
// Business Flow:
// 1. Validate policy exists and status transition is valid
// 2. Update policy status
// 3. Create policy history record
// 4. Update related entities (loans, bonuses)
// 5. Send notifications
// 6. Archive old records if terminated
func PolicyUpdateWorkflow(ctx workflow.Context, input PolicyUpdateWorkflowInput) (*PolicyUpdateWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting Policy Update Workflow", "PolicyID", input.PolicyID, "NewStatus", input.NewStatus)

	result := &PolicyUpdateWorkflowResult{
		PolicyID:  input.PolicyID,
		NewStatus: input.NewStatus,
	}

	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts:    3,
			InitialInterval:    1 * time.Second,
			BackoffCoefficient: 2.0,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	// Step 1: Validate Status Transition
	logger.Info("Step 1: Validating status transition")
	var validateResult activities.ValidateStatusTransitionResult
	err := workflow.ExecuteActivity(ctx, activities.ValidateStatusTransitionActivity, activities.ValidateStatusTransitionInput{
		PolicyID:  input.PolicyID,
		NewStatus: input.NewStatus,
	}).Get(ctx, &validateResult)

	if err != nil {
		logger.Error("Failed to validate status transition", "error", err)
		return result, err
	}

	if !validateResult.Valid {
		logger.Error("Invalid status transition", "reason", validateResult.Reason)
		return result, err
	}

	result.OldStatus = validateResult.CurrentStatus

	// Step 2: Update Policy Status
	logger.Info("Step 2: Updating policy status")
	var updateResult activities.UpdatePolicyStatusResult
	err = workflow.ExecuteActivity(ctx, activities.UpdatePolicyStatusActivity, activities.UpdatePolicyStatusInput{
		PolicyID:           input.PolicyID,
		SurrenderRequestID: input.SurrenderRequestID,
		NewStatus:          input.NewStatus,
	}).Get(ctx, &updateResult)

	if err != nil {
		logger.Error("Failed to update policy status", "error", err)
		return result, err
	}

	result.UpdatedAt = workflow.Now(ctx)
	logger.Info("Policy status updated", "old_status", result.OldStatus, "new_status", result.NewStatus)

	// Step 3: Create Policy History Record
	logger.Info("Step 3: Creating policy history record")
	var historyResult activities.CreatePolicyHistoryResult
	err = workflow.ExecuteActivity(ctx, activities.CreatePolicyHistoryActivity, activities.CreatePolicyHistoryInput{
		PolicyID:           input.PolicyID,
		SurrenderRequestID: input.SurrenderRequestID,
		OldStatus:          result.OldStatus,
		NewStatus:          result.NewStatus,
		ChangeReason:       "Surrender processing - " + input.DispositionType,
	}).Get(ctx, &historyResult)

	if err != nil {
		logger.Error("Failed to create policy history", "error", err)
		// Don't fail workflow
	}

	// Step 4: Update Related Entities
	logger.Info("Step 4: Updating related entities")
	relatedUpdates := 0

	// Step 4a: Settle Loans (if any)
	if input.NewStatus == "TS" || input.NewStatus == "AU" {
		var loanResult activities.SettlePolicyLoansResult
		err = workflow.ExecuteActivity(ctx, activities.SettlePolicyLoansActivity, activities.SettlePolicyLoansInput{
			PolicyID:           input.PolicyID,
			SurrenderRequestID: input.SurrenderRequestID,
		}).Get(ctx, &loanResult)

		if err != nil {
			logger.Error("Failed to settle loans", "error", err)
		} else if loanResult.LoansSettled > 0 {
			logger.Info("Loans settled", "count", loanResult.LoansSettled)
			relatedUpdates++
		}
	}

	// Step 4b: Stop Future Bonuses (if terminated)
	if input.NewStatus == "TS" {
		var bonusResult activities.StopFutureBonusesResult
		err = workflow.ExecuteActivity(ctx, activities.StopFutureBonusesActivity, activities.StopFutureBonusesInput{
			PolicyID: input.PolicyID,
		}).Get(ctx, &bonusResult)

		if err != nil {
			logger.Error("Failed to stop future bonuses", "error", err)
		} else {
			logger.Info("Future bonuses stopped")
			relatedUpdates++
		}
	}

	// Step 4c: Update Reduced Paid-Up Details (if AU status)
	if input.NewStatus == "AU" {
		var reducedPUResult activities.UpdateReducedPaidUpDetailsResult
		err = workflow.ExecuteActivity(ctx, activities.UpdateReducedPaidUpDetailsActivity, activities.UpdateReducedPaidUpDetailsInput{
			PolicyID:           input.PolicyID,
			SurrenderRequestID: input.SurrenderRequestID,
		}).Get(ctx, &reducedPUResult)

		if err != nil {
			logger.Error("Failed to update reduced paid-up details", "error", err)
		} else {
			logger.Info("Reduced paid-up details updated", "new_sum_assured", reducedPUResult.NewSumAssured)
			relatedUpdates++
		}
	}

	result.RelatedUpdates = relatedUpdates

	// Step 5: Send Notifications
	logger.Info("Step 5: Sending notifications")
	var notificationResult activities.SendPolicyUpdateNotificationResult
	err = workflow.ExecuteActivity(ctx, activities.SendPolicyUpdateNotificationActivity, activities.SendPolicyUpdateNotificationInput{
		PolicyID:        input.PolicyID,
		OldStatus:       result.OldStatus,
		NewStatus:       result.NewStatus,
		DispositionType: input.DispositionType,
	}).Get(ctx, &notificationResult)

	if err != nil {
		logger.Error("Failed to send notification", "error", err)
		// Don't fail workflow
	} else {
		logger.Info("Notification sent", "channels", notificationResult.ChannelsSent)
	}

	// Step 6: Archive Records (if terminated)
	if input.NewStatus == "TS" {
		logger.Info("Step 6: Archiving policy records")
		var archiveResult activities.ArchivePolicyRecordsResult
		err = workflow.ExecuteActivity(ctx, activities.ArchivePolicyRecordsActivity, activities.ArchivePolicyRecordsInput{
			PolicyID: input.PolicyID,
		}).Get(ctx, &archiveResult)

		if err != nil {
			logger.Error("Failed to archive records", "error", err)
			// Don't fail workflow
		} else {
			logger.Info("Policy records archived", "records_archived", archiveResult.RecordsArchived)
		}
	}

	logger.Info("Policy Update Workflow completed successfully")
	result.CompletedAt = workflow.Now(ctx)

	return result, nil
}
