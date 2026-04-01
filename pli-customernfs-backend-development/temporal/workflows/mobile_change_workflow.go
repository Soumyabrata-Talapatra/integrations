package workflows

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"customer-nfs-service/temporal/activities"
)

// MobileChangeWorkflowInput is the input for WF-NFS-006.
type MobileChangeWorkflowInput struct {
	RequestID    string  `json:"request_id"`
	TicketNumber string  `json:"ticket_number"`
	CustomerID   int64   `json:"customer_id"`
	PolicyNumber *string `json:"policy_number,omitempty"`
	Channel      string  `json:"channel"`
	InitiatedBy  string  `json:"initiated_by"`
}

// ===========================================================================
// WF-NFS-006: MobileChangeWorkflow (OTP-based verification)
// ===========================================================================

// MobileChangeWorkflow handles mobile number change via OTP verification.
// Flow: Initiate -> Send OTP to new mobile -> Signal: otp_submitted -> Verify OTP
//       -> Update mobile -> COMPLETED -> Notify Policy Management
func MobileChangeWorkflow(ctx workflow.Context, input MobileChangeWorkflowInput) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("MobileChangeWorkflow start", "requestID", input.RequestID)

	actOpts := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 3,
			InitialInterval: 2 * time.Second,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, actOpts)

	// Step 1: Store workflow state (enables HTTP signal bridge).
	wfInfo := workflow.GetInfo(ctx)
	if err := workflow.ExecuteActivity(ctx, "StoreWorkflowState", activities.StoreWorkflowStateInput{
		RequestID:     input.RequestID,
		WorkflowID:    wfInfo.WorkflowExecution.ID,
		WorkflowRunID: wfInfo.WorkflowExecution.RunID,
	}).Get(ctx, nil); err != nil {
		return fmt.Errorf("StoreWorkflowState: %w", err)
	}

	// Step 2: Request OTP to new mobile number.
	var otpResult activities.AadhaarOTPRequestResult
	if err := workflow.ExecuteActivity(ctx, "RequestMobileOTP", activities.AadhaarOTPRequestInput{
		RequestID:  input.RequestID,
		CustomerID: input.CustomerID,
		Channel:    input.Channel,
	}).Get(ctx, &otpResult); err != nil {
		return fmt.Errorf("RequestMobileOTP: %w", err)
	}

	// Step 3: Wait for OTP submission signal (15-minute timeout).
	otpCh := workflow.GetSignalChannel(ctx, SignalOTPSubmitted)
	var otpPayload OTPSignalPayload
	timerCtx, cancelTimer := workflow.WithCancel(ctx)
	timer := workflow.NewTimer(timerCtx, 15*time.Minute)

	sel := workflow.NewSelector(ctx)
	otpReceived := false

	sel.AddReceive(otpCh, func(c workflow.ReceiveChannel, more bool) {
		c.Receive(ctx, &otpPayload)
		otpReceived = true
		cancelTimer()
	})
	sel.AddFuture(timer, func(f workflow.Future) {
		// OTP expired — mark as DOCUMENTS_EXPIRED.
		_ = workflow.ExecuteActivity(ctx, "MobileUpdateStatus", activities.UpdateStatusInput{
			RequestID: input.RequestID,
			NewStatus: "DOCUMENTS_EXPIRED",
			UpdatedBy: "SYSTEM",
			Channel:   input.Channel,
		}).Get(ctx, nil)
	})
	sel.Select(ctx)

	if !otpReceived {
		logger.Info("MobileChangeWorkflow OTP timeout", "requestID", input.RequestID)
		return fmt.Errorf("OTP timeout for mobile change request %s", input.RequestID)
	}

	// Step 4: Verify OTP.
	var verifyResult activities.AadhaarOTPVerifyResult
	if err := workflow.ExecuteActivity(ctx, "VerifyMobileOTP", activities.AadhaarOTPVerifyInput{
		RequestID:      input.RequestID,
		OTPSubmitted:   otpPayload.OTP,
		OTPReferenceID: otpPayload.OTPReferenceID,
		CustomerID:     input.CustomerID,
	}).Get(ctx, &verifyResult); err != nil {
		return fmt.Errorf("VerifyMobileOTP: %w", err)
	}

	// Step 5: Update mobile number data.
	var updateResult activities.UpdateMobileDataResult
	if err := workflow.ExecuteActivity(ctx, "UpdateMobileData", activities.UpdateMobileDataInput{
		RequestID:  input.RequestID,
		CustomerID: input.CustomerID,
		UpdatedBy:  fmt.Sprintf("%d", input.CustomerID),
	}).Get(ctx, &updateResult); err != nil {
		return fmt.Errorf("UpdateMobileData: %w", err)
	}

	// Step 6: Generate acknowledgment receipt.
	_ = workflow.ExecuteActivity(ctx, "GenerateAckReceipt", activities.GenerateAckReceiptInput{
		RequestID:    input.RequestID,
		TicketNumber: input.TicketNumber,
		CustomerID:   input.CustomerID,
		RequestType:  "MOBILE_CHANGE",
	}).Get(ctx, nil)

	// Step 7: Status -> COMPLETED.
	if err := workflow.ExecuteActivity(ctx, "MobileUpdateStatus", activities.UpdateStatusInput{
		RequestID: input.RequestID,
		NewStatus: "COMPLETED",
		UpdatedBy: fmt.Sprintf("%d", input.CustomerID),
		Channel:   input.Channel,
	}).Get(ctx, nil); err != nil {
		return fmt.Errorf("UpdateStatus COMPLETED: %w", err)
	}

	logger.Info("MobileChangeWorkflow COMPLETED", "requestID", input.RequestID)

	// Step 8: Notify Policy Management.
	payload := map[string]interface{}{}
	if updateResult.NewMobileNumber != nil {
		payload["mobile_number"] = *updateResult.NewMobileNumber
	}
	
	_ = workflow.ExecuteActivity(ctx, "NotifyPolicyManagement", activities.NotifyPMInput{
		RequestID:   input.RequestID,
		CustomerID:  input.CustomerID,
		RequestType: "MOBILE_CHANGE",
		Outcome:     "APPROVED",
		ChangePayload: mustMarshalJSON(payload),
	}).Get(ctx, nil)

	return nil
}
