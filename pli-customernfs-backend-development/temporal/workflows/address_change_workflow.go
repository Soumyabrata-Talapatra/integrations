// // Package workflows implements Temporal workflow stubs for address change.
// //
// // WF-NFS-001: Aadhaar-based Address Change
// //   - OTP window: 15 minutes (configurable)
// //   - Signal: "otp_submitted" carries OTP value
// //   - On timeout: status → DOCUMENTS_EXPIRED
// //   - On verify success: UpdateAddressData → COMPLETED
// //
// // WF-NFS-002: Manual Address Change (CPC Approval Workflow)
// //   - SLA deadline: 15 days (Regional) / 30 days (Circle HQ) from config
// //   - Signal: "approval_decision" carries { decision, reason, approved_by }
// //   - On approve: UpdateAddressData → COMPLETED
// //   - On reject:  status → REJECTED
// //   - On send-back: status → PENDING_DOCUMENTS (customer re-upload sub-flow)
// //   - On SLA breach: Escalate → sla_breached = TRUE
// //
// // WORKFLOW STATE NOTE:
// //
// //	Both workflows call StoreWorkflowState activity immediately after start.
// //	This persists workflow_id + workflow_run_id into nfs.service_request.
// //	HTTP handlers use these values with temporal.Client.SignalWorkflow.
// //
// // BATCH NOTE:
// //
// //	CreateAddressServiceRequest activity uses a single batched TX:
// //	service_request INSERT + address_change_detail INSERT + audit_log INSERT.
// package workflows

// import (
// 	"fmt"
// 	"time"

// 	"go.temporal.io/sdk/temporal"
// 	"go.temporal.io/sdk/workflow"

// 	"customer-nfs-service/temporal/activities"
// )

// // Signal names
// const (
// 	SignalOTPSubmitted       = "otp_submitted"
// 	SignalApprovalDecision   = "approval_decision"
// 	SignalWithdrawalApproved = "withdrawal_approved"
// 	SignalDocumentsSubmitted = "documents_submitted"
// )

// // Task queue
// const TaskQueue = "customer-nfs-tq"

// // OTPSignalPayload is the payload for the otp_submitted signal.
// type OTPSignalPayload struct {
// 	OTP            string `json:"otp"`
// 	OTPReferenceID string `json:"otp_reference_id"`
// 	CustomerID     string `json:"customer_id"`
// }

// // ApprovalDecisionPayload is the payload for the approval_decision signal.
// type ApprovalDecisionPayload struct {
// 	Decision         string   `json:"decision"`
// 	ApprovedBy       string   `json:"approved_by"`
// 	Reason           string   `json:"reason,omitempty"`
// 	RejectionReason  *string  `json:"rejection_reason,omitempty"`
// 	MissingDocuments []string `json:"missing_documents,omitempty"`
// }

// // AddressChangeWorkflowInput is the WF input type for both WF-NFS-001 and WF-NFS-002.
// type AddressChangeWorkflowInput struct {
// 	RequestID    string  `json:"request_id"`
// 	TicketNumber string  `json:"ticket_number"`
// 	CustomerID   string  `json:"customer_id"`
// 	PolicyNumber *string `json:"policy_number,omitempty"`
// 	AuthMethod   string  `json:"auth_method"`
// 	AddressType  string  `json:"address_type"`
// 	Channel      string  `json:"channel"`
// 	OfficeCode   string  `json:"office_code,omitempty"`
// 	InitiatedBy  string  `json:"initiated_by"`
// 	SLADays      int     `json:"sla_days"`
// }

// // defaultActivityOptions is the standard RetryPolicy for NFS activities.
// var defaultActivityOptions = workflow.ActivityOptions{
// 	StartToCloseTimeout: 30 * time.Second,
// 	RetryPolicy: &temporal.RetryPolicy{
// 		MaximumAttempts: 3,
// 		InitialInterval: 2 * time.Second,
// 	},
// }

// // ===========================================================================
// // WF-NFS-001: AadhaarAddressChangeWorkflow
// // ===========================================================================

// func AadhaarAddressChangeWorkflow(ctx workflow.Context, input AddressChangeWorkflowInput) error {
// 	logger := workflow.GetLogger(ctx)
// 	logger.Info("AadhaarAddressChangeWorkflow start", "requestID", input.RequestID)

// 	ctx = workflow.WithActivityOptions(ctx, defaultActivityOptions)

// 	// Step 1: Store workflow state
// 	wfInfo := workflow.GetInfo(ctx)
// 	if err := workflow.ExecuteActivity(ctx, "StoreWorkflowState", activities.StoreWorkflowStateInput{
// 		RequestID:     input.RequestID,
// 		WorkflowID:    wfInfo.WorkflowExecution.ID,
// 		WorkflowRunID: wfInfo.WorkflowExecution.RunID,
// 	}).Get(ctx, nil); err != nil {
// 		return fmt.Errorf("StoreWorkflowState: %w", err)
// 	}

// 	// Step 2: Request Aadhaar OTP
// 	var otpResult activities.AadhaarOTPRequestResult
// 	if err := workflow.ExecuteActivity(ctx, "RequestAadhaarOTP", activities.AadhaarOTPRequestInput{
// 		RequestID:  input.RequestID,
// 		CustomerID: input.CustomerID,
// 		Channel:    input.Channel,
// 	}).Get(ctx, &otpResult); err != nil {
// 		return fmt.Errorf("RequestAadhaarOTP: %w", err)
// 	}
// 	logger.Info("OTP requested", "referenceID", otpResult.OTPReferenceID)

// 	// Step 3: Wait for otp_submitted signal (15-min timeout)
// 	otpCh := workflow.GetSignalChannel(ctx, SignalOTPSubmitted)
// 	var otpPayload OTPSignalPayload
// 	otpTimeout := 15 * time.Minute

// 	sel := workflow.NewSelector(ctx)
// 	var timedOut bool
// 	var gotOTP bool
// 	timer := workflow.NewTimer(ctx, otpTimeout)
// 	sel.AddFuture(timer, func(f workflow.Future) { timedOut = true })
// 	sel.AddReceive(otpCh, func(c workflow.ReceiveChannel, more bool) {
// 		c.Receive(ctx, &otpPayload)
// 		gotOTP = true
// 	})
// 	sel.Select(ctx)

// 	if timedOut || !gotOTP {
// 		logger.Info("OTP timeout", "requestID", input.RequestID)
// 		_ = workflow.ExecuteActivity(ctx, "AddressUpdateStatus", activities.UpdateStatusInput{
// 			RequestID:  input.RequestID,
// 			NewStatus:  "DOCUMENTS_EXPIRED",
// 			FromStatus: "CREATED",
// 			UpdatedBy:  input.InitiatedBy,
// 			Reason:     strPtr("OTP window expired after 15 minutes"),
// 			Channel:    input.Channel,
// 		}).Get(ctx, nil)
// 		return nil
// 	}

// 	// Step 4: Verify Aadhaar OTP
// 	var verifyResult activities.AadhaarOTPVerifyResult
// 	if err := workflow.ExecuteActivity(ctx, "VerifyAadhaarOTP", activities.AadhaarOTPVerifyInput{
// 		RequestID:      input.RequestID,
// 		OTPReferenceID: otpPayload.OTPReferenceID,
// 		OTPSubmitted:   otpPayload.OTP,
// 		CustomerID:     input.CustomerID,
// 	}).Get(ctx, &verifyResult); err != nil {
// 		return fmt.Errorf("VerifyAadhaarOTP: %w", err)
// 	}

// 	if !verifyResult.Verified {
// 		logger.Info("OTP verification failed", "reason", verifyResult.FailReason)
// 		_ = workflow.ExecuteActivity(ctx, "AddressUpdateStatus", activities.UpdateStatusInput{
// 			RequestID:  input.RequestID,
// 			NewStatus:  "REJECTED",
// 			FromStatus: "CREATED",
// 			UpdatedBy:  input.InitiatedBy,
// 			Reason:     &verifyResult.FailReason,
// 			Channel:    input.Channel,
// 		}).Get(ctx, nil)
// 		return nil
// 	}

// 	// Step 5: Apply new address data
// 	if err := workflow.ExecuteActivity(ctx, "UpdateAddressData", activities.UpdateAddressDataInput{
// 		RequestID:   input.RequestID,
// 		CustomerID:  input.CustomerID,
// 		AddressType: "COMMUNICATION",
// 		UpdatedBy:   input.InitiatedBy,
// 	}).Get(ctx, nil); err != nil {
// 		logger.Error("UpdateAddressData failed (partial)", "error", err)
// 		return fmt.Errorf("UpdateAddressData: %w", err)
// 	}

// 	// Step 6: Generate acknowledgement receipt
// 	if err := workflow.ExecuteActivity(ctx, "GenerateAckReceipt", activities.GenerateAckReceiptInput{
// 		RequestID:    input.RequestID,
// 		TicketNumber: input.TicketNumber,
// 		CustomerID:   input.CustomerID,
// 		RequestType:  "ADDRESS_CHANGE",
// 	}).Get(ctx, nil); err != nil {
// 		logger.Error("GenerateAckReceipt failed (non-fatal)", "error", err)
// 	}

// 	// Step 7: Status → COMPLETED
// 	if err := workflow.ExecuteActivity(ctx, "AddressUpdateStatus", activities.UpdateStatusInput{
// 		RequestID:  input.RequestID,
// 		NewStatus:  "COMPLETED",
// 		FromStatus: "CREATED",
// 		UpdatedBy:  input.InitiatedBy,
// 		Channel:    input.Channel,
// 	}).Get(ctx, nil); err != nil {
// 		return fmt.Errorf("UpdateStatus COMPLETED: %w", err)
// 	}

// 	logger.Info("AadhaarAddressChangeWorkflow COMPLETED", "requestID", input.RequestID)
// 	return nil
// }

// // ===========================================================================
// // WF-NFS-002: ManualAddressChangeWorkflow
// // ===========================================================================

// func ManualAddressChangeWorkflow(ctx workflow.Context, input AddressChangeWorkflowInput) error {
// 	logger := workflow.GetLogger(ctx)
// 	logger.Info("ManualAddressChangeWorkflow start", "requestID", input.RequestID)

// 	ctx = workflow.WithActivityOptions(ctx, defaultActivityOptions)

// 	// Step 1: Store workflow state
// 	wfInfo := workflow.GetInfo(ctx)
// 	if err := workflow.ExecuteActivity(ctx, "StoreWorkflowState", activities.StoreWorkflowStateInput{
// 		RequestID:     input.RequestID,
// 		WorkflowID:    wfInfo.WorkflowExecution.ID,
// 		WorkflowRunID: wfInfo.WorkflowExecution.RunID,
// 	}).Get(ctx, nil); err != nil {
// 		return fmt.Errorf("StoreWorkflowState: %w", err)
// 	}

// 	// Step 2: Status → PENDING_APPROVAL
// 	if err := workflow.ExecuteActivity(ctx, "AddressUpdateStatus", activities.UpdateStatusInput{
// 		RequestID:  input.RequestID,
// 		NewStatus:  "PENDING_APPROVAL",
// 		FromStatus: "CREATED",
// 		UpdatedBy:  input.InitiatedBy,
// 		Channel:    input.Channel,
// 		OfficeCode: &input.OfficeCode,
// 	}).Get(ctx, nil); err != nil {
// 		return fmt.Errorf("UpdateStatus PENDING_APPROVAL: %w", err)
// 	}

// 	// Step 3: Assign to CPC queue
// 	slaDays := input.SLADays
// 	if slaDays == 0 {
// 		slaDays = 15
// 	}
// 	slaDeadline := workflow.Now(ctx).Add(time.Duration(slaDays) * 24 * time.Hour)
// 	if err := workflow.ExecuteActivity(ctx, "AssignToCPC", activities.AssignToCPCInput{
// 		RequestID:   input.RequestID,
// 		OfficeCode:  &input.OfficeCode,
// 		Priority:    5,
// 		SLADeadline: &slaDeadline,
// 		AssignedBy:  input.InitiatedBy,
// 	}).Get(ctx, nil); err != nil {
// 		return fmt.Errorf("AssignToCPC: %w", err)
// 	}

// 	// Step 4+5: CPC approval loop — handles SEND_BACK re-submissions.
// 	for {
// 		approvalCh := workflow.GetSignalChannel(ctx, SignalApprovalDecision)
// 		var approvalPayload ApprovalDecisionPayload
// 		var gotDecision bool

// 		slaTimer := workflow.NewTimer(ctx, time.Duration(slaDays)*24*time.Hour)
// 		sel := workflow.NewSelector(ctx)
// 		var slaBreached bool

// 		sel.AddFuture(slaTimer, func(f workflow.Future) { slaBreached = true })
// 		sel.AddReceive(approvalCh, func(c workflow.ReceiveChannel, more bool) {
// 			c.Receive(ctx, &approvalPayload)
// 			gotDecision = true
// 		})
// 		sel.Select(ctx)

// 		if slaBreached && !gotDecision {
// 			logger.Warn("SLA breached", "requestID", input.RequestID)
// 			_ = workflow.ExecuteActivity(ctx, "Escalate", activities.EscalateInput{
// 				RequestID:    input.RequestID,
// 				TicketNumber: input.TicketNumber,
// 				SLABreached:  true,
// 				EscalateTo:   "CIRCLE_HQ",
// 			}).Get(ctx, nil)

// 			graceTimer := workflow.NewTimer(ctx, 7*24*time.Hour)
// 			sel2 := workflow.NewSelector(ctx)
// 			sel2.AddFuture(graceTimer, func(f workflow.Future) {})
// 			sel2.AddReceive(approvalCh, func(c workflow.ReceiveChannel, more bool) {
// 				c.Receive(ctx, &approvalPayload)
// 				gotDecision = true
// 			})
// 			sel2.Select(ctx)
// 		}

// 		if !gotDecision {
// 			_ = workflow.ExecuteActivity(ctx, "AddressUpdateStatus", activities.UpdateStatusInput{
// 				RequestID:  input.RequestID,
// 				NewStatus:  "REJECTED",
// 				FromStatus: "PENDING_APPROVAL",
// 				UpdatedBy:  input.InitiatedBy,
// 				Reason:     strPtr("No CPC decision within SLA + grace period"),
// 				Channel:    input.Channel,
// 			}).Get(ctx, nil)
// 			return nil
// 		}

// 		switch approvalPayload.Decision {
// 		case "APPROVE":
// 			logger.Info("CPC approved", "requestID", input.RequestID, "by", approvalPayload.ApprovedBy)
// 			if err := workflow.ExecuteActivity(ctx, "AddressUpdateStatus", activities.UpdateStatusInput{
// 				RequestID:  input.RequestID,
// 				NewStatus:  "IN_PROGRESS",
// 				FromStatus: "PENDING_APPROVAL",
// 				UpdatedBy:  approvalPayload.ApprovedBy,
// 				Channel:    input.Channel,
// 			}).Get(ctx, nil); err != nil {
// 				return fmt.Errorf("UpdateStatus IN_PROGRESS: %w", err)
// 			}
// 			if err := workflow.ExecuteActivity(ctx, "UpdateAddressData", activities.UpdateAddressDataInput{
// 				RequestID:   input.RequestID,
// 				CustomerID:  input.CustomerID,
// 				AddressType: "COMMUNICATION",
// 				UpdatedBy:   approvalPayload.ApprovedBy,
// 			}).Get(ctx, nil); err != nil {
// 				return fmt.Errorf("UpdateAddressData: %w", err)
// 			}
// 			_ = workflow.ExecuteActivity(ctx, "GenerateAckReceipt", activities.GenerateAckReceiptInput{
// 				RequestID:    input.RequestID,
// 				TicketNumber: input.TicketNumber,
// 				CustomerID:   input.CustomerID,
// 				RequestType:  "ADDRESS_CHANGE",
// 			}).Get(ctx, nil)
// 			if err := workflow.ExecuteActivity(ctx, "AddressUpdateStatus", activities.UpdateStatusInput{
// 				RequestID:  input.RequestID,
// 				NewStatus:  "COMPLETED",
// 				FromStatus: "IN_PROGRESS",
// 				UpdatedBy:  approvalPayload.ApprovedBy,
// 				Channel:    input.Channel,
// 			}).Get(ctx, nil); err != nil {
// 				return fmt.Errorf("UpdateStatus COMPLETED: %w", err)
// 			}
// 			logger.Info("ManualAddressChangeWorkflow COMPLETED", "requestID", input.RequestID)
// 			return nil

// 		case "REJECT":
// 			if err := workflow.ExecuteActivity(ctx, "AddressUpdateStatus", activities.UpdateStatusInput{
// 				RequestID:  input.RequestID,
// 				NewStatus:  "REJECTED",
// 				FromStatus: "PENDING_APPROVAL",
// 				UpdatedBy:  approvalPayload.ApprovedBy,
// 				Reason:     &approvalPayload.Reason,
// 				Channel:    input.Channel,
// 			}).Get(ctx, nil); err != nil {
// 				return fmt.Errorf("UpdateStatus REJECTED: %w", err)
// 			}
// 			logger.Info("ManualAddressChangeWorkflow REJECTED", "requestID", input.RequestID)
// 			return nil

// 		case "SEND_BACK":
// 			if err := workflow.ExecuteActivity(ctx, "AddressUpdateStatus", activities.UpdateStatusInput{
// 				RequestID:  input.RequestID,
// 				NewStatus:  "PENDING_DOCUMENTS",
// 				FromStatus: "PENDING_APPROVAL",
// 				UpdatedBy:  approvalPayload.ApprovedBy,
// 				Reason:     &approvalPayload.Reason,
// 				Channel:    input.Channel,
// 			}).Get(ctx, nil); err != nil {
// 				return fmt.Errorf("UpdateStatus PENDING_DOCUMENTS: %w", err)
// 			}
// 			// Wait for customer to re-submit documents.
// 			resubmitCh := workflow.GetSignalChannel(ctx, SignalDocumentsSubmitted)
// 			var resubmitPayload activities.DocumentsSubmittedPayload
// 			resubmitCh.Receive(ctx, &resubmitPayload)
// 			// Re-assign to CPC.
// 			if err := workflow.ExecuteActivity(ctx, "AssignToCPC", activities.AssignToCPCInput{
// 				RequestID:   input.RequestID,
// 				OfficeCode:  &input.OfficeCode,
// 				Priority:    5,
// 				SLADeadline: &slaDeadline,
// 				AssignedBy:  input.InitiatedBy,
// 			}).Get(ctx, nil); err != nil {
// 				return fmt.Errorf("AssignToCPC re-assign: %w", err)
// 			}
// 			// Continue loop → back to top for new approval wait.

// 		default:
// 			return fmt.Errorf("unknown approval decision: %s", approvalPayload.Decision)
// 		}
// 	}
// }

// // strPtr is a helper to convert string literal to *string.
// func strPtr(s string) *string { return &s }

// Package workflows implements Temporal workflow stubs for address change.
//
// WF-NFS-001: Aadhaar-based Address Change
//   - OTP window: 15 minutes (configurable)
//   - Signal: "otp_submitted" carries OTP value
//   - On timeout: status → DOCUMENTS_EXPIRED
//   - On verify success: UpdateAddressData → COMPLETED
//
// WF-NFS-002: Manual Address Change (CPC Approval Workflow)
//   - SLA deadline: 15 days (Regional) / 30 days (Circle HQ) from config
//   - Signal: "approval_decision" carries { decision, reason, approved_by }
//   - On approve: UpdateAddressData → COMPLETED
//   - On reject:  status → REJECTED
//   - On send-back: status → PENDING_DOCUMENTS (customer re-upload sub-flow)
//   - On SLA breach: Escalate → sla_breached = TRUE
//
// WORKFLOW STATE NOTE:
//
//	Both workflows call StoreWorkflowState activity immediately after start.
//	This persists workflow_id + workflow_run_id into nfs.service_request.
//	HTTP handlers use these values with temporal.Client.SignalWorkflow.
//
// BATCH NOTE:
//
//	CreateAddressServiceRequest activity uses a single batched TX:
//	service_request INSERT + address_change_detail INSERT + audit_log INSERT.
package workflows

import (
	"encoding/json"
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"customer-nfs-service/temporal/activities"
)

// Signal names
const (
	SignalOTPSubmitted       = "otp_submitted"
	SignalApprovalDecision   = "approval_decision"
	SignalWithdrawalApproved = "withdrawal_approved"
	SignalDocumentsSubmitted = "documents_submitted"
)

// Task queue
const TaskQueue = "customer-nfs-tq"

// OTPSignalPayload is the payload for the otp_submitted signal.
type OTPSignalPayload struct {
	OTP            string `json:"otp"`
	OTPReferenceID string `json:"otp_reference_id"`
	CustomerID     int64  `json:"customer_id"`
}

// ApprovalDecisionPayload is the payload for the approval_decision signal.
type ApprovalDecisionPayload struct {
	Decision         string   `json:"decision"`
	ApprovedBy       string   `json:"approved_by"`
	Reason           string   `json:"reason,omitempty"`
	RejectionReason  *string  `json:"rejection_reason,omitempty"`
	MissingDocuments []string `json:"missing_documents,omitempty"`
}

// AddressChangeWorkflowInput is the WF input type for both WF-NFS-001 and WF-NFS-002.
type AddressChangeWorkflowInput struct {
	RequestID    string  `json:"request_id"`
	TicketNumber string  `json:"ticket_number"`
	CustomerID   int64   `json:"customer_id"`
	PolicyNumber *string `json:"policy_number,omitempty"`
	AuthMethod   string  `json:"auth_method"`
	AddressType  string  `json:"address_type"`
	Channel      string  `json:"channel"`
	OfficeCode   string  `json:"office_code,omitempty"`
	InitiatedBy  string  `json:"initiated_by"`
	SLADays      int     `json:"sla_days"`
}

// defaultActivityOptions is the standard RetryPolicy for NFS activities.
var defaultActivityOptions = workflow.ActivityOptions{
	StartToCloseTimeout: 30 * time.Second,
	RetryPolicy: &temporal.RetryPolicy{
		MaximumAttempts: 3,
		InitialInterval: 2 * time.Second,
	},
}

// ===========================================================================
// WF-NFS-001: AadhaarAddressChangeWorkflow
// ===========================================================================

func AadhaarAddressChangeWorkflow(ctx workflow.Context, input AddressChangeWorkflowInput) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("AadhaarAddressChangeWorkflow start", "requestID", input.RequestID)

	ctx = workflow.WithActivityOptions(ctx, defaultActivityOptions)

	// Step 1: Store workflow state
	wfInfo := workflow.GetInfo(ctx)
	if err := workflow.ExecuteActivity(ctx, "StoreWorkflowState", activities.StoreWorkflowStateInput{
		RequestID:     input.RequestID,
		WorkflowID:    wfInfo.WorkflowExecution.ID,
		WorkflowRunID: wfInfo.WorkflowExecution.RunID,
	}).Get(ctx, nil); err != nil {
		return fmt.Errorf("StoreWorkflowState: %w", err)
	}

	// Step 2: Request Aadhaar OTP
	var otpResult activities.AadhaarOTPRequestResult
	if err := workflow.ExecuteActivity(ctx, "RequestAadhaarOTP", activities.AadhaarOTPRequestInput{
		RequestID:  input.RequestID,
		CustomerID: input.CustomerID,
		Channel:    input.Channel,
	}).Get(ctx, &otpResult); err != nil {
		return fmt.Errorf("RequestAadhaarOTP: %w", err)
	}
	logger.Info("OTP requested", "referenceID", otpResult.OTPReferenceID)

	// Step 3: Wait for otp_submitted signal (15-min timeout)
	otpCh := workflow.GetSignalChannel(ctx, SignalOTPSubmitted)
	var otpPayload OTPSignalPayload
	otpTimeout := 15 * time.Minute

	sel := workflow.NewSelector(ctx)
	var timedOut bool
	var gotOTP bool
	timer := workflow.NewTimer(ctx, otpTimeout)
	sel.AddFuture(timer, func(f workflow.Future) { timedOut = true })
	sel.AddReceive(otpCh, func(c workflow.ReceiveChannel, more bool) {
		c.Receive(ctx, &otpPayload)
		gotOTP = true
	})
	sel.Select(ctx)

	if timedOut || !gotOTP {
		logger.Info("OTP timeout", "requestID", input.RequestID)
		_ = workflow.ExecuteActivity(ctx, "AddressUpdateStatus", activities.UpdateStatusInput{
			RequestID:  input.RequestID,
			NewStatus:  "DOCUMENTS_EXPIRED",
			FromStatus: "CREATED",
			UpdatedBy:  input.InitiatedBy,
			Reason:     strPtr("OTP window expired after 15 minutes"),
			Channel:    input.Channel,
		}).Get(ctx, nil)
		return nil
	}

	// Step 4: Verify Aadhaar OTP
	var verifyResult activities.AadhaarOTPVerifyResult
	if err := workflow.ExecuteActivity(ctx, "VerifyAadhaarOTP", activities.AadhaarOTPVerifyInput{
		RequestID:      input.RequestID,
		OTPReferenceID: otpPayload.OTPReferenceID,
		OTPSubmitted:   otpPayload.OTP,
		CustomerID:     input.CustomerID,
	}).Get(ctx, &verifyResult); err != nil {
		return fmt.Errorf("VerifyAadhaarOTP: %w", err)
	}

	if !verifyResult.Verified {
		logger.Info("OTP verification failed", "reason", verifyResult.FailReason)
		_ = workflow.ExecuteActivity(ctx, "AddressUpdateStatus", activities.UpdateStatusInput{
			RequestID:  input.RequestID,
			NewStatus:  "REJECTED",
			FromStatus: "CREATED",
			UpdatedBy:  input.InitiatedBy,
			Reason:     &verifyResult.FailReason,
			Channel:    input.Channel,
		}).Get(ctx, nil)
		return nil
	}

	// Step 5: Apply new address data
	if err := workflow.ExecuteActivity(ctx, "UpdateAddressData", activities.UpdateAddressDataInput{
		RequestID:   input.RequestID,
		CustomerID:  input.CustomerID,
		AddressType: "COMMUNICATION",
		UpdatedBy:   input.InitiatedBy,
	}).Get(ctx, nil); err != nil {
		logger.Error("UpdateAddressData failed (partial)", "error", err)
		return fmt.Errorf("UpdateAddressData: %w", err)
	}

	// Step 6: Generate acknowledgement receipt
	if err := workflow.ExecuteActivity(ctx, "GenerateAckReceipt", activities.GenerateAckReceiptInput{
		RequestID:    input.RequestID,
		TicketNumber: input.TicketNumber,
		CustomerID:   input.CustomerID,
		RequestType:  "ADDRESS_CHANGE",
	}).Get(ctx, nil); err != nil {
		logger.Error("GenerateAckReceipt failed (non-fatal)", "error", err)
	}

	// Step 7: Status → COMPLETED
	if err := workflow.ExecuteActivity(ctx, "AddressUpdateStatus", activities.UpdateStatusInput{
		RequestID:  input.RequestID,
		NewStatus:  "COMPLETED",
		FromStatus: "CREATED",
		UpdatedBy:  input.InitiatedBy,
		Channel:    input.Channel,
	}).Get(ctx, nil); err != nil {
		return fmt.Errorf("UpdateStatus COMPLETED: %w", err)
	}

	logger.Info("AadhaarAddressChangeWorkflow COMPLETED", "requestID", input.RequestID)

	// Notify Policy Management of completed address change.
	_ = workflow.ExecuteActivity(ctx, "NotifyPolicyManagement", activities.NotifyPMInput{
		RequestID:   input.RequestID,
		CustomerID:  input.CustomerID,
		RequestType: "ADDRESS_CHANGE",
		Outcome:     "APPROVED",
		ChangePayload: mustMarshalJSON(map[string]interface{}{
			"address_type": input.AddressType,
		}),
	}).Get(ctx, nil)

	return nil
}

// ===========================================================================
// WF-NFS-002: ManualAddressChangeWorkflow
// ===========================================================================

func ManualAddressChangeWorkflow(ctx workflow.Context, input AddressChangeWorkflowInput) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("ManualAddressChangeWorkflow start", "requestID", input.RequestID)

	ctx = workflow.WithActivityOptions(ctx, defaultActivityOptions)

	// Step 1: Store workflow state
	wfInfo := workflow.GetInfo(ctx)
	if err := workflow.ExecuteActivity(ctx, "StoreWorkflowState", activities.StoreWorkflowStateInput{
		RequestID:     input.RequestID,
		WorkflowID:    wfInfo.WorkflowExecution.ID,
		WorkflowRunID: wfInfo.WorkflowExecution.RunID,
	}).Get(ctx, nil); err != nil {
		return fmt.Errorf("StoreWorkflowState: %w", err)
	}

	// Step 2: Status → PENDING_APPROVAL
	if err := workflow.ExecuteActivity(ctx, "AddressUpdateStatus", activities.UpdateStatusInput{
		RequestID:  input.RequestID,
		NewStatus:  "PENDING_APPROVAL",
		FromStatus: "CREATED",
		UpdatedBy:  input.InitiatedBy,
		Channel:    input.Channel,
		OfficeCode: &input.OfficeCode,
	}).Get(ctx, nil); err != nil {
		return fmt.Errorf("UpdateStatus PENDING_APPROVAL: %w", err)
	}

	// Step 3: Assign to CPC queue
	slaDays := input.SLADays
	if slaDays == 0 {
		slaDays = 15
	}
	slaDeadline := workflow.Now(ctx).Add(time.Duration(slaDays) * 24 * time.Hour)
	if err := workflow.ExecuteActivity(ctx, "AssignToCPC", activities.AssignToCPCInput{
		RequestID:   input.RequestID,
		OfficeCode:  &input.OfficeCode,
		Priority:    5,
		SLADeadline: &slaDeadline,
		AssignedBy:  input.InitiatedBy,
	}).Get(ctx, nil); err != nil {
		return fmt.Errorf("AssignToCPC: %w", err)
	}

	// Step 4+5: CPC approval loop — handles SEND_BACK re-submissions.
	for {
		approvalCh := workflow.GetSignalChannel(ctx, SignalApprovalDecision)
		var approvalPayload ApprovalDecisionPayload
		var gotDecision bool

		slaTimer := workflow.NewTimer(ctx, time.Duration(slaDays)*24*time.Hour)
		sel := workflow.NewSelector(ctx)
		var slaBreached bool

		sel.AddFuture(slaTimer, func(f workflow.Future) { slaBreached = true })
		sel.AddReceive(approvalCh, func(c workflow.ReceiveChannel, more bool) {
			c.Receive(ctx, &approvalPayload)
			gotDecision = true
		})
		sel.Select(ctx)

		if slaBreached && !gotDecision {
			logger.Warn("SLA breached", "requestID", input.RequestID)
			_ = workflow.ExecuteActivity(ctx, "Escalate", activities.EscalateInput{
				RequestID:    input.RequestID,
				TicketNumber: input.TicketNumber,
				SLABreached:  true,
				EscalateTo:   "CIRCLE_HQ",
			}).Get(ctx, nil)

			graceTimer := workflow.NewTimer(ctx, 7*24*time.Hour)
			sel2 := workflow.NewSelector(ctx)
			sel2.AddFuture(graceTimer, func(f workflow.Future) {})
			sel2.AddReceive(approvalCh, func(c workflow.ReceiveChannel, more bool) {
				c.Receive(ctx, &approvalPayload)
				gotDecision = true
			})
			sel2.Select(ctx)
		}

		if !gotDecision {
			_ = workflow.ExecuteActivity(ctx, "AddressUpdateStatus", activities.UpdateStatusInput{
				RequestID:  input.RequestID,
				NewStatus:  "REJECTED",
				FromStatus: "PENDING_APPROVAL",
				UpdatedBy:  input.InitiatedBy,
				Reason:     strPtr("No CPC decision within SLA + grace period"),
				Channel:    input.Channel,
			}).Get(ctx, nil)
			return nil
		}

		switch approvalPayload.Decision {
		case "APPROVE":
			logger.Info("CPC approved", "requestID", input.RequestID, "by", approvalPayload.ApprovedBy)
			if err := workflow.ExecuteActivity(ctx, "AddressUpdateStatus", activities.UpdateStatusInput{
				RequestID:  input.RequestID,
				NewStatus:  "IN_PROGRESS",
				FromStatus: "PENDING_APPROVAL",
				UpdatedBy:  approvalPayload.ApprovedBy,
				Channel:    input.Channel,
			}).Get(ctx, nil); err != nil {
				return fmt.Errorf("UpdateStatus IN_PROGRESS: %w", err)
			}
			if err := workflow.ExecuteActivity(ctx, "UpdateAddressData", activities.UpdateAddressDataInput{
				RequestID:   input.RequestID,
				CustomerID:  input.CustomerID,
				AddressType: "COMMUNICATION",
				UpdatedBy:   approvalPayload.ApprovedBy,
			}).Get(ctx, nil); err != nil {
				return fmt.Errorf("UpdateAddressData: %w", err)
			}
			_ = workflow.ExecuteActivity(ctx, "GenerateAckReceipt", activities.GenerateAckReceiptInput{
				RequestID:    input.RequestID,
				TicketNumber: input.TicketNumber,
				CustomerID:   input.CustomerID,
				RequestType:  "ADDRESS_CHANGE",
			}).Get(ctx, nil)
			if err := workflow.ExecuteActivity(ctx, "AddressUpdateStatus", activities.UpdateStatusInput{
				RequestID:  input.RequestID,
				NewStatus:  "COMPLETED",
				FromStatus: "IN_PROGRESS",
				UpdatedBy:  approvalPayload.ApprovedBy,
				Channel:    input.Channel,
			}).Get(ctx, nil); err != nil {
				return fmt.Errorf("UpdateStatus COMPLETED: %w", err)
			}
			logger.Info("ManualAddressChangeWorkflow COMPLETED", "requestID", input.RequestID)

			// Notify Policy Management of completed address change.
			_ = workflow.ExecuteActivity(ctx, "NotifyPolicyManagement", activities.NotifyPMInput{
				RequestID:   input.RequestID,
				CustomerID:  input.CustomerID,
				RequestType: "ADDRESS_CHANGE",
				Outcome:     "APPROVED",
				ChangePayload: mustMarshalJSON(map[string]interface{}{
					"address_type": input.AddressType,
				}),
			}).Get(ctx, nil)

			return nil

		case "REJECT":
			if err := workflow.ExecuteActivity(ctx, "AddressUpdateStatus", activities.UpdateStatusInput{
				RequestID:  input.RequestID,
				NewStatus:  "REJECTED",
				FromStatus: "PENDING_APPROVAL",
				UpdatedBy:  approvalPayload.ApprovedBy,
				Reason:     &approvalPayload.Reason,
				Channel:    input.Channel,
			}).Get(ctx, nil); err != nil {
				return fmt.Errorf("UpdateStatus REJECTED: %w", err)
			}
			logger.Info("ManualAddressChangeWorkflow REJECTED", "requestID", input.RequestID)

			// Notify Policy Management of rejected address change.
			_ = workflow.ExecuteActivity(ctx, "NotifyPolicyManagement", activities.NotifyPMInput{
				RequestID:   input.RequestID,
				CustomerID:  input.CustomerID,
				RequestType: "ADDRESS_CHANGE",
				Outcome:     "REJECTED",
			}).Get(ctx, nil)

			return nil

		case "SEND_BACK":
			if err := workflow.ExecuteActivity(ctx, "AddressUpdateStatus", activities.UpdateStatusInput{
				RequestID:  input.RequestID,
				NewStatus:  "PENDING_DOCUMENTS",
				FromStatus: "PENDING_APPROVAL",
				UpdatedBy:  approvalPayload.ApprovedBy,
				Reason:     &approvalPayload.Reason,
				Channel:    input.Channel,
			}).Get(ctx, nil); err != nil {
				return fmt.Errorf("UpdateStatus PENDING_DOCUMENTS: %w", err)
			}
			// Wait for customer to re-submit documents.
			resubmitCh := workflow.GetSignalChannel(ctx, SignalDocumentsSubmitted)
			var resubmitPayload activities.DocumentsSubmittedPayload
			resubmitCh.Receive(ctx, &resubmitPayload)
			// Re-assign to CPC.
			if err := workflow.ExecuteActivity(ctx, "AssignToCPC", activities.AssignToCPCInput{
				RequestID:   input.RequestID,
				OfficeCode:  &input.OfficeCode,
				Priority:    5,
				SLADeadline: &slaDeadline,
				AssignedBy:  input.InitiatedBy,
			}).Get(ctx, nil); err != nil {
				return fmt.Errorf("AssignToCPC re-assign: %w", err)
			}
			// Continue loop → back to top for new approval wait.

		default:
			return fmt.Errorf("unknown approval decision: %s", approvalPayload.Decision)
		}
	}
}

// strPtr is a helper to convert string literal to *string.
func strPtr(s string) *string { return &s }

// mustMarshalJSON marshals v to JSON, returning nil on error.
func mustMarshalJSON(v interface{}) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	return b
}
