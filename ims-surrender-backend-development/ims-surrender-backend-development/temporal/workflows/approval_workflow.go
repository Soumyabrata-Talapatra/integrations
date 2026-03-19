package workflows

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"gitlab.cept.gov.in/it-2.0-policy/surrender-service/temporal/activities"
)

// ApprovalWorkflowInput defines the input for approval workflow
// TEMP-004: Approval Processing Workflow
type ApprovalWorkflowInput struct {
	SurrenderRequestID string
	Priority           string
	AutoApprovalLimit  float64
}

// ApprovalWorkflowResult defines the result of approval workflow
type ApprovalWorkflowResult struct {
	SurrenderRequestID string
	Decision           string
	ApprovedBy         string
	ApprovedAt         time.Time
	EscalationCount    int
	CompletedAt        time.Time
}

// ApprovalWorkflow orchestrates the approval process
// TEMP-004: Approval Processing Workflow
// Business Flow:
// 1. Check auto-approval eligibility
// 2. Create approval task
// 3. Wait for assignment
// 4. Monitor SLA
// 5. Auto-escalate if needed
// 6. Process decision
func ApprovalWorkflow(ctx workflow.Context, input ApprovalWorkflowInput) (*ApprovalWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting Approval Workflow", "SurrenderRequestID", input.SurrenderRequestID)

	result := &ApprovalWorkflowResult{
		SurrenderRequestID: input.SurrenderRequestID,
	}

	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	// Step 1: Get Surrender Request Details
	logger.Info("Step 1: Getting surrender request details")
	var detailsResult activities.GetSurrenderRequestDetailsResult
	err := workflow.ExecuteActivity(ctx, activities.GetSurrenderRequestDetailsActivity, activities.GetSurrenderRequestDetailsInput{
		SurrenderRequestID: input.SurrenderRequestID,
	}).Get(ctx, &detailsResult)

	if err != nil {
		logger.Error("Failed to get surrender details", "error", err)
		return result, err
	}

	// Step 2: Check Auto-Approval Eligibility
	logger.Info("Step 2: Checking auto-approval eligibility")
	autoApprovalLimit := input.AutoApprovalLimit
	if autoApprovalLimit == 0 {
		autoApprovalLimit = 5000.0 // Default threshold
	}

	if detailsResult.NetSurrenderValue <= autoApprovalLimit {
		logger.Info("Amount eligible for auto-approval", "amount", detailsResult.NetSurrenderValue)

		var autoApproveResult activities.AutoApproveResult
		err = workflow.ExecuteActivity(ctx, activities.AutoApproveActivity, activities.AutoApproveInput{
			SurrenderRequestID: input.SurrenderRequestID,
		}).Get(ctx, &autoApproveResult)

		if err != nil {
			logger.Error("Auto-approval failed", "error", err)
			// Continue to manual approval
		} else {
			logger.Info("Auto-approved successfully")
			result.Decision = "AUTO_APPROVED"
			result.ApprovedBy = "SYSTEM"
			result.ApprovedAt = workflow.Now(ctx)
			result.CompletedAt = workflow.Now(ctx)
			return result, nil
		}
	}

	// Step 3: Create Approval Task
	logger.Info("Step 3: Creating approval task")
	var createTaskResult activities.CreateApprovalTaskResult
	err = workflow.ExecuteActivity(ctx, activities.CreateApprovalTaskActivity, activities.CreateApprovalTaskInput{
		SurrenderRequestID: input.SurrenderRequestID,
		Priority:           input.Priority,
	}).Get(ctx, &createTaskResult)

	if err != nil {
		logger.Error("Failed to create approval task", "error", err)
		return result, err
	}

	taskID := createTaskResult.TaskID
	logger.Info("Approval task created", "TaskID", taskID)

	// Step 4: Wait for Approval Decision with SLA Monitoring
	logger.Info("Step 4: Waiting for approval decision")

	var decision string
	escalationCount := 0
	slaHours := 24 // 24-hour SLA
	maxEscalations := 3

	for {
		// Wait for approval decision signal or SLA timeout
		selector := workflow.NewSelector(ctx)

		decisionChannel := workflow.GetSignalChannel(ctx, "approval-decision")
		selector.AddReceive(decisionChannel, func(c workflow.ReceiveChannel, more bool) {
			c.Receive(ctx, &decision)
			logger.Info("Received approval decision", "decision", decision)
		})

		// SLA timer
		slaTimer := workflow.NewTimer(ctx, time.Duration(slaHours)*time.Hour)
		selector.AddFuture(slaTimer, func(f workflow.Future) {
			logger.Warn("SLA timeout reached", "hours", slaHours)
		})

		selector.Select(ctx)

		// If decision received, break
		if decision != "" {
			break
		}

		// SLA breached - escalate
		escalationCount++
		logger.Info("Escalating task", "escalation_count", escalationCount)

		if escalationCount >= maxEscalations {
			logger.Error("Maximum escalations reached")
			result.Decision = "TIMEOUT"
			result.EscalationCount = escalationCount
			result.CompletedAt = workflow.Now(ctx)
			return result, fmt.Errorf("approval timeout after %d escalations", escalationCount)
		}

		var escalateResult activities.EscalateApprovalTaskResult
		err = workflow.ExecuteActivity(ctx, activities.EscalateApprovalTaskActivity, activities.EscalateApprovalTaskInput{
			TaskID:           taskID,
			EscalationLevel:  escalationCount,
			EscalationReason: fmt.Sprintf("SLA breach - %d hours elapsed", slaHours*escalationCount),
		}).Get(ctx, &escalateResult)

		if err != nil {
			logger.Error("Failed to escalate task", "error", err)
		} else {
			logger.Info("Task escalated successfully", "new_priority", escalateResult.NewPriority)
		}

		// Reduce SLA for escalated task
		slaHours = 12 // Escalated tasks have 12-hour SLA
	}

	result.EscalationCount = escalationCount

	// Step 5: Process Approval Decision
	logger.Info("Step 5: Processing approval decision", "decision", decision)

	if decision == "APPROVED" {
		var processResult activities.ProcessApprovalDecisionResult
		err = workflow.ExecuteActivity(ctx, activities.ProcessApprovalDecisionActivity, activities.ProcessApprovalDecisionInput{
			SurrenderRequestID: input.SurrenderRequestID,
			Decision:           decision,
			TaskID:             taskID,
		}).Get(ctx, &processResult)

		if err != nil {
			logger.Error("Failed to process approval", "error", err)
			return result, err
		}

		result.Decision = "APPROVED"
		result.ApprovedBy = processResult.ApprovedBy
		result.ApprovedAt = workflow.Now(ctx)
	} else {
		result.Decision = "REJECTED"
	}

	logger.Info("Approval Workflow completed", "decision", result.Decision)
	result.CompletedAt = workflow.Now(ctx)

	return result, nil
}
