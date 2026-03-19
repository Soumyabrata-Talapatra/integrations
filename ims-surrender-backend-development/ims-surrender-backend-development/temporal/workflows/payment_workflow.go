package workflows

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"gitlab.cept.gov.in/it-2.0-policy/surrender-service/temporal/activities"
)

// PaymentWorkflowInput defines the input for payment workflow
// TEMP-005: Payment Disposition Workflow
type PaymentWorkflowInput struct {
	SurrenderRequestID string
	PolicyID           string
	Amount             float64
	DisbursementMethod string
}

// PaymentWorkflowResult defines the result of payment workflow
type PaymentWorkflowResult struct {
	SurrenderRequestID string
	PaymentReference   string
	Status             string
	ProcessedAt        time.Time
	DispositionType    string
	NewPolicyStatus    string
	CompletedAt        time.Time
}

// PaymentWorkflow orchestrates payment processing and disposition
// TEMP-005: Payment Disposition Workflow
// Business Flow:
// 1. Validate payment eligibility
// 2. Determine disposition (Terminated vs Reduced Paid-Up)
// 3. Process payment
// 4. Create disposition record
// 5. Update policy status
// 6. Send notifications
func PaymentWorkflow(ctx workflow.Context, input PaymentWorkflowInput) (*PaymentWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting Payment Workflow", "SurrenderRequestID", input.SurrenderRequestID)

	result := &PaymentWorkflowResult{
		SurrenderRequestID: input.SurrenderRequestID,
	}

	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts:    5,
			InitialInterval:    2 * time.Second,
			BackoffCoefficient: 2.0,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	// Step 1: Validate Payment Eligibility
	logger.Info("Step 1: Validating payment eligibility")
	var validateResult activities.ValidatePaymentEligibilityResult
	err := workflow.ExecuteActivity(ctx, activities.ValidatePaymentEligibilityActivity, activities.ValidatePaymentEligibilityInput{
		SurrenderRequestID: input.SurrenderRequestID,
	}).Get(ctx, &validateResult)

	if err != nil {
		logger.Error("Failed to validate payment eligibility", "error", err)
		result.Status = "FAILED_VALIDATION"
		return result, err
	}

	if !validateResult.Eligible {
		logger.Error("Not eligible for payment", "reason", validateResult.Reason)
		result.Status = "NOT_ELIGIBLE"
		return result, fmt.Errorf("not eligible for payment: %s", validateResult.Reason)
	}

	// Step 2: Determine Disposition Type
	logger.Info("Step 2: Determining disposition type")
	var dispositionResult activities.DetermineDispositionResult
	err = workflow.ExecuteActivity(ctx, activities.DetermineDispositionActivity, activities.DetermineDispositionInput{
		SurrenderRequestID: input.SurrenderRequestID,
		NetSurrenderValue:  input.Amount,
	}).Get(ctx, &dispositionResult)

	if err != nil {
		logger.Error("Failed to determine disposition", "error", err)
		return result, err
	}

	result.DispositionType = dispositionResult.DispositionType
	result.NewPolicyStatus = dispositionResult.NewPolicyStatus
	logger.Info("Disposition determined", "type", result.DispositionType, "policy_status", result.NewPolicyStatus)

	// Step 3: Process Payment
	logger.Info("Step 3: Processing payment")
	var paymentResult activities.ProcessPaymentResult
	err = workflow.ExecuteActivity(ctx, activities.ProcessPaymentActivity, activities.ProcessPaymentInput{
		SurrenderRequestID: input.SurrenderRequestID,
		Amount:             input.Amount,
		DisbursementMethod: input.DisbursementMethod,
	}).Get(ctx, &paymentResult)

	if err != nil {
		logger.Error("Failed to process payment", "error", err)
		result.Status = "FAILED_PAYMENT"
		return result, err
	}

	result.PaymentReference = paymentResult.PaymentReference
	result.ProcessedAt = workflow.Now(ctx)
	logger.Info("Payment processed", "reference", result.PaymentReference)

	// Step 4: Create Disposition Record
	logger.Info("Step 4: Creating disposition record")
	var dispositionRecordResult activities.CreateDispositionRecordResult
	err = workflow.ExecuteActivity(ctx, activities.CreateDispositionRecordActivity, activities.CreateDispositionRecordInput{
		SurrenderRequestID: input.SurrenderRequestID,
		PolicyID:           input.PolicyID,
		DispositionType:    dispositionResult.DispositionType,
		PaymentReference:   result.PaymentReference,
		NetAmount:          input.Amount,
		NewPolicyStatus:    dispositionResult.NewPolicyStatus,
	}).Get(ctx, &dispositionRecordResult)

	if err != nil {
		logger.Error("Failed to create disposition record", "error", err)
		// Don't fail - payment already processed
		result.Status = "PAYMENT_COMPLETED_DISPOSITION_FAILED"
		return result, nil
	}

	// Step 5: Update Policy Status
	logger.Info("Step 5: Updating policy status")
	var policyUpdateResult activities.UpdatePolicyStatusResult
	err = workflow.ExecuteActivity(ctx, activities.UpdatePolicyStatusActivity, activities.UpdatePolicyStatusInput{
		PolicyID:           input.PolicyID,
		SurrenderRequestID: input.SurrenderRequestID,
		NewStatus:          dispositionResult.NewPolicyStatus,
	}).Get(ctx, &policyUpdateResult)

	if err != nil {
		logger.Error("Failed to update policy status", "error", err)
		// Don't fail - payment already processed
		result.Status = "PAYMENT_COMPLETED_POLICY_UPDATE_FAILED"
		return result, nil
	}

	logger.Info("Policy status updated", "new_status", policyUpdateResult.NewStatus)

	// Step 6: Send Notifications
	logger.Info("Step 6: Sending notifications")
	var notificationResult activities.SendPaymentNotificationResult
	err = workflow.ExecuteActivity(ctx, activities.SendPaymentNotificationActivity, activities.SendPaymentNotificationInput{
		SurrenderRequestID: input.SurrenderRequestID,
		PolicyID:           input.PolicyID,
		PaymentReference:   result.PaymentReference,
		Amount:             input.Amount,
		DispositionType:    dispositionResult.DispositionType,
	}).Get(ctx, &notificationResult)

	if err != nil {
		logger.Error("Failed to send notification", "error", err)
		// Don't fail - payment already completed
	} else {
		logger.Info("Notification sent", "channels", notificationResult.ChannelsSent)
	}

	// Workflow completed successfully
	logger.Info("Payment Workflow completed successfully")
	result.Status = "COMPLETED"
	result.CompletedAt = workflow.Now(ctx)

	return result, nil
}
