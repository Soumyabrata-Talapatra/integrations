// Package workflows implements Temporal workflow stubs for name change and withdrawal.
//
// WF-NFS-003: Aadhaar-based Name Change
//   - OTP window: 15 minutes (configurable)
//   - Signal: "otp_submitted" carries OTP value
//   - On timeout: status → DOCUMENTS_EXPIRED
//   - On verify success: UpdateNameData → COMPLETED (BR-NFS-007)
//   - BR-NFS-010: DOB from UIDAI response DISCARDED
//
// WF-NFS-004: Manual Name Change (CPC Approval Workflow)
//   - BR-NFS-008: At least one legal document required
//   - Signal: "documents_submitted" moves to AssignToCPC
//   - Signal: "approval_decision" — APPROVE → UpdateNameData → COMPLETED
//     — REJECT → REJECTED
//     — SEND_BACK → PENDING_DOCUMENTS
//   - On SLA breach: Escalate activity
//   - BR-NFS-009: On APPROVE, customer.name.updated event propagated to all policies
//
// WF-NFS-005: Customer Withdrawal
//   - BR-NFS-013: Withdrawal only allowed for CREATED/PENDING_DOCUMENTS/PENDING_APPROVAL
//   - Auto-approved if partial_processing_flag=FALSE
//   - Manual CPC approval required if partial_processing_flag=TRUE
//   - On completion: CancelPendingActivities cancels missing-doc links + CPC queue entries
//
// WORKFLOW STATE NOTE:
//
//	StoreWorkflowState is always the FIRST activity. It persists workflow_id + workflow_run_id
//	into nfs.service_request. HTTP handlers use tc.SignalWorkflow with these IDs.
package workflows

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"customer-nfs-service/temporal/activities"
)

// NameChangeWorkflowInput is the input type for WF-NFS-003 and WF-NFS-004.
type NameChangeWorkflowInput struct {
	RequestID    string  `json:"request_id"`
	TicketNumber string  `json:"ticket_number"`
	CustomerID   int64   `json:"customer_id"`
	PolicyNumber *string `json:"policy_number,omitempty"`
	AuthMethod   string  `json:"auth_method"` // AADHAAR or MANUAL
	Channel      string  `json:"channel"`
	OfficeCode   string  `json:"office_code,omitempty"`
	InitiatedBy  string  `json:"initiated_by"`
	SLADays      int     `json:"sla_days"` // 15 or 30
}

// WithdrawalWorkflowInput is the input type for WF-NFS-005.
type WithdrawalWorkflowInput struct {
	RequestID             string `json:"request_id"`
	TicketNumber          string `json:"ticket_number"`
	CustomerID            int64  `json:"customer_id"`
	WithdrawalReason      string `json:"withdrawal_reason"`
	RequestedBy           string `json:"requested_by"`
	PartialProcessingFlag bool   `json:"partial_processing_flag"`
}

// SignalWithdrawalDecision is the signal name for CPC withdrawal approval.
const SignalWithdrawalDecision = "withdrawal_decision"

// withdrawalActivityOpts is the standard activity options for withdrawal operations.
var withdrawalActivityOpts = workflow.ActivityOptions{
	StartToCloseTimeout: 30 * time.Second,
	RetryPolicy: &temporal.RetryPolicy{
		MaximumAttempts: 3,
		InitialInterval: 2 * time.Second,
	},
}

// ===========================================================================
// WF-NFS-003: AadhaarNameChangeWorkflow
// ===========================================================================

// AadhaarNameChangeWorkflow handles Aadhaar-based name change (immediate path).
// Triggered by CORE-005 POST /nfs/name-change/initiate when auth_method=AADHAAR.
//
// FR-NFS-004: Aadhaar name change.
// BR-NFS-007: UIDAI name is authoritative on successful OTP verification.
// BR-NFS-010: DOB from UIDAI response must be discarded.
// BR-NFS-009: customer.name.updated event propagated to all policies on COMPLETE.
//
// State machine: CREATED → [OTP sent] → [Signal: otp_submitted / 15-min timeout]
//
//	→ [VerifyOTP] → UpdateNameData → COMPLETED
//	                                → DOCUMENTS_EXPIRED (on timeout)
func AadhaarNameChangeWorkflow(ctx workflow.Context, input NameChangeWorkflowInput) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("WF-NFS-003 AadhaarNameChangeWorkflow started",
		"requestID", input.RequestID,
		"ticketNumber", input.TicketNumber,
	)

	actCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 3,
			InitialInterval: 2 * time.Second,
		},
	})

	var act *activities.NameChangeActivities

	// Step 1: Store workflow state — enables HTTP signal dispatch.
	// WORKFLOW STATE: workflow_id + workflow_run_id persisted to DB.
	var storeResult activities.StoreWorkflowStateResult
	if err := workflow.ExecuteActivity(actCtx, act.StoreWorkflowStateForName,
		activities.StoreWorkflowStateInput{RequestID: input.RequestID},
	).Get(ctx, &storeResult); err != nil {
		return err
	}

	// Step 2: Send Aadhaar OTP (BR-NFS-007: Aadhaar identity verification).
	var otpResult activities.AadhaarOTPRequestResult
	if err := workflow.ExecuteActivity(actCtx, act.RequestNameOTP,
		activities.AadhaarOTPRequestInput{
			RequestID:  input.RequestID,
			CustomerID: input.CustomerID,
		},
	).Get(ctx, &otpResult); err != nil {
		return err
	}

	// Step 3: Wait for otp_submitted signal (15-minute window).
	otpTimeout := 15 * time.Minute
	otpCh := workflow.GetSignalChannel(ctx, SignalOTPSubmitted)
	var otpPayload OTPSignalPayload

	otpTimer := workflow.NewTimer(ctx, otpTimeout)
	selector := workflow.NewSelector(ctx)
	otpReceived := false

	selector.AddReceive(otpCh, func(c workflow.ReceiveChannel, more bool) {
		c.Receive(ctx, &otpPayload)
		otpReceived = true
	})
	selector.AddFuture(otpTimer, func(f workflow.Future) {})
	selector.Select(ctx)

	if !otpReceived {
		// OTP window expired → DOCUMENTS_EXPIRED (BR-NFS-012).
		logger.Warn("WF-NFS-003: OTP timeout — transitioning to DOCUMENTS_EXPIRED",
			"requestID", input.RequestID)
		_ = workflow.ExecuteActivity(actCtx, act.UpdateNameStatus,
			activities.UpdateStatusInput{
				RequestID: input.RequestID,
				NewStatus: "DOCUMENTS_EXPIRED",
				UpdatedBy: "system",
			},
		).Get(ctx, nil)
		return nil
	}

	// Step 4: Verify OTP with KYC Service (STUB — INT-NFS-003).
	// BR-NFS-010: DOB from UIDAI response DISCARDED inside VerifyNameOTP.
	var verifyResult activities.AadhaarOTPVerifyResult
	if err := workflow.ExecuteActivity(actCtx, act.VerifyNameOTP,
		activities.AadhaarOTPVerifyInput{
			RequestID:      input.RequestID,
			OTPSubmitted:   otpPayload.OTP,
			OTPReferenceID: otpPayload.OTPReferenceID,
			CustomerID:     input.CustomerID,
		},
	).Get(ctx, &verifyResult); err != nil {
		return err
	}

	// Step 5: Apply name change — create new version history entry.
	// BR-NFS-007: UIDAI name is now authoritative.
	// BR-NFS-009: UpdateNameData activity publishes customer.name.updated event.
	var updateResult activities.UpdateNameDataResult
	if err := workflow.ExecuteActivity(actCtx, act.UpdateNameData,
		activities.UpdateNameDataInput{
			RequestID: input.RequestID,
			CustomerID: input.CustomerID,
			UpdatedBy: input.InitiatedBy,
		},
	).Get(ctx, &updateResult); err != nil {
		return err
	}

	// Step 6: Generate acknowledgment receipt (PDF).
	if err := workflow.ExecuteActivity(actCtx, act.GenerateNameAckReceipt,
		activities.GenerateAckReceiptInput{
			RequestID:   input.RequestID,
			RequestType: "NAME_CHANGE",
		},
	).Get(ctx, nil); err != nil {
		logger.Warn("WF-NFS-003: receipt generation failed (non-fatal)", "error", err)
	}

	logger.Info("WF-NFS-003 AadhaarNameChangeWorkflow COMPLETED", "requestID", input.RequestID)

	// Notify Policy Management of completed name change.
	payload := map[string]interface{}{}
	if updateResult.NewSalutation != nil {
		payload["salutation"] = *updateResult.NewSalutation
	}
	if updateResult.NewFirstName != nil {
		payload["first_name"] = *updateResult.NewFirstName
	}
	if updateResult.NewMiddleName != nil {
		payload["middle_name"] = *updateResult.NewMiddleName
	}
	if updateResult.NewLastName != nil {
		payload["last_name"] = *updateResult.NewLastName
	}
	
	_ = workflow.ExecuteActivity(actCtx, "NotifyPolicyManagement", activities.NotifyPMInput{
		RequestID:   input.RequestID,
		CustomerID:  input.CustomerID,
		RequestType: "NAME_CHANGE",
		Outcome:     "APPROVED",
		ChangePayload: mustMarshalJSON(payload),
	}).Get(ctx, nil)

	return nil
}

// ===========================================================================
// WF-NFS-004: ManualNameChangeWorkflow
// ===========================================================================

// ManualNameChangeWorkflow handles manual name change with CPC approval.
// Triggered by CORE-005 POST /nfs/name-change/initiate when auth_method=MANUAL.
//
// FR-NFS-005: Manual name change.
// BR-NFS-008: At least ONE legal document required (Gazette/Newspaper/Application Form).
// BR-NFS-009: customer.name.updated event published on APPROVE.
// BR-NFS-010: DOB remains immutable throughout.
// BR-NFS-012: State machine — CREATED → PENDING_APPROVAL → IN_PROGRESS → COMPLETED/REJECTED/PENDING_DOCUMENTS.
//
// State machine:
//
//	CREATED → [Signal: documents_submitted] → PENDING_APPROVAL
//	        → AssignToCPC → IN_PROGRESS
//	        → [Signal: approval_decision / SLA timeout]
//	        → APPROVE        → UpdateNameData → COMPLETED
//	        → REJECT         → REJECTED
//	        → SEND_BACK      → PENDING_DOCUMENTS (missing-doc sub-flow)
//	        → SLA breach     → Escalate → continue waiting
//
// ManualNameChangeWorkflow handles manual name change with CPC approval.
// Triggered by CORE-005 POST /nfs/name-change/initiate when auth_method=MANUAL.
//
// FR-NFS-005: Manual name change.
// BR-NFS-008: At least ONE legal document required (Gazette/Newspaper/Application Form).
// BR-NFS-009: customer.name.updated event published on APPROVE.
// BR-NFS-010: DOB remains immutable throughout.
// BR-NFS-012: State machine — CREATED → PENDING_APPROVAL → IN_PROGRESS → COMPLETED/REJECTED/PENDING_DOCUMENTS.
//
// State machine:
//
//	CREATED → [Signal: documents_submitted] → PENDING_APPROVAL
//	        → AssignToCPC → IN_PROGRESS
//	        → [Signal: approval_decision / SLA timeout]
//	        → APPROVE        → UpdateNameData → COMPLETED
//	        → REJECT         → REJECTED
//	        → SEND_BACK      → PENDING_DOCUMENTS → [Signal: documents_submitted] → loop
//	        → SLA breach     → Escalate → continue waiting
//	        → Total deadline → DOCUMENTS_EXPIRED
func ManualNameChangeWorkflow(ctx workflow.Context, input NameChangeWorkflowInput) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("WF-NFS-004 ManualNameChangeWorkflow started",
		"requestID", input.RequestID,
		"ticketNumber", input.TicketNumber,
		"slaDays", input.SLADays,
	)

	actCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 3,
			InitialInterval: 2 * time.Second,
		},
	})

	var act *activities.NameChangeActivities

	// Step 1: Store workflow state for HTTP signaling.
	var storeResult activities.StoreWorkflowStateResult
	if err := workflow.ExecuteActivity(actCtx, act.StoreWorkflowStateForName,
		activities.StoreWorkflowStateInput{RequestID: input.RequestID},
	).Get(ctx, &storeResult); err != nil {
		return err
	}

	// Step 2: Wait for documents_submitted signal from CORE-007.
	// This allows the customer to upload documents before CPC assignment.
	docsCh := workflow.GetSignalChannel(ctx, SignalDocumentsSubmitted)
	withdrawCh := workflow.GetSignalChannel(ctx, "withdrawal_requested")

	var docsPayload activities.DocumentsSubmittedPayload
	var withdrawn bool
	sel := workflow.NewSelector(ctx)
	sel.AddReceive(docsCh, func(c workflow.ReceiveChannel, _ bool) {
		c.Receive(ctx, &docsPayload)
	})
	sel.AddReceive(withdrawCh, func(c workflow.ReceiveChannel, _ bool) {
		c.Receive(ctx, nil)
		withdrawn = true
	})
	sel.Select(ctx)
	if withdrawn {
		logger.Info("WF-NFS-004: withdrawal signal received — exiting", "requestID", input.RequestID)
		return nil
	}
	if input.SLADays == 0 {
		input.SLADays = 15
	}
	slaDeadline := workflow.Now(ctx).AddDate(0, 0, input.SLADays)
	gracePeriod := 7 * 24 * time.Hour
	totalWait := slaDeadline.Sub(workflow.Now(ctx)) + gracePeriod

	// Step 3: Assign to CPC pool (BR-NFS-002: assignment based on office).
	// BATCH (TX): AssignNameToCPC — 3-op TX batch.
	var assignResult activities.AssignToCPCResult
	if err := workflow.ExecuteActivity(actCtx, act.AssignNameToCPC,
		activities.AssignToCPCInput{
			RequestID:   input.RequestID,
			OfficeCode:  &input.OfficeCode,
			AssignedBy:  input.InitiatedBy, // ← add
			Priority:    1,                 // ← add
			SLADeadline: &slaDeadline,      // ← add
		},
	).Get(ctx, &assignResult); err != nil {
		return err
	}

	// Step 4: CPC approval loop — handles SEND_BACK re-submissions.
	// BR-NFS-012: SEND_BACK → PENDING_DOCUMENTS → customer re-uploads → back to PENDING_APPROVAL.
	// Loop continues until a terminal decision (APPROVE or REJECT) or total deadline.

	for {
		slaTimer := workflow.NewTimer(ctx, time.Duration(input.SLADays)*24*time.Hour)
		totalTimer := workflow.NewTimer(ctx, totalWait)
		approvalCh := workflow.GetSignalChannel(ctx, SignalApprovalDecision)
		var approvalPayload ApprovalDecisionPayload
		approved := false

		// NOTE: selector.Select fires for exactly ONE case per call (Temporal guarantee).
		// Case 1: approvalCh fires → approved=true → process decision below.
		// Case 2: slaTimer fires → escalate → re-wait via selector2 for approvalCh or totalTimer.
		// Case 3: totalTimer fires → auto-close. slaTimer escalation skipped — moot at this point.
		// Do NOT collapse selector2 into selector1 — re-wait must reuse the same totalTimer
		// so the deadline is not reset after escalation.
		selector := workflow.NewSelector(ctx)

		selector.AddFuture(slaTimer, func(f workflow.Future) {
			logger.Warn("WF-NFS-004: SLA deadline reached — escalating",
				"requestID", input.RequestID,
				"slaDays", input.SLADays,
			)
			_ = workflow.ExecuteActivity(actCtx, act.EscalateNameRequest,
				activities.EscalateInput{RequestID: input.RequestID},
			).Get(ctx, nil)
		})

		selector.AddFuture(totalTimer, func(f workflow.Future) {})

		selector.AddReceive(approvalCh, func(c workflow.ReceiveChannel, more bool) {
			c.Receive(ctx, &approvalPayload)
			approved = true
		})

		selector.Select(ctx)

		// Re-wait after SLA escalation — reuse totalTimer so deadline is not reset.
		if !approved && workflow.Now(ctx).Before(slaDeadline.Add(gracePeriod)) {
			selector2 := workflow.NewSelector(ctx)
			selector2.AddReceive(approvalCh, func(c workflow.ReceiveChannel, more bool) {
				c.Receive(ctx, &approvalPayload)
				approved = true
			})
			selector2.AddFuture(totalTimer, func(f workflow.Future) {})
			selector2.Select(ctx)
		}

		if !approved {
			// Total deadline reached — auto-close (DOCUMENTS_EXPIRED as terminal state).
			logger.Warn("WF-NFS-004: total deadline reached — auto-closing", "requestID", input.RequestID)
			_ = workflow.ExecuteActivity(actCtx, act.UpdateNameStatus,
				activities.UpdateStatusInput{
					RequestID: input.RequestID,
					NewStatus: "DOCUMENTS_EXPIRED",
					UpdatedBy: "system",
				},
			).Get(ctx, nil)
			return nil
		}

		// Step 5: Process approval decision (BR-NFS-012).
		switch approvalPayload.Decision {
		case "APPROVE":
			logger.Info("WF-NFS-004: CPC approved", "requestID", input.RequestID, "by", approvalPayload.ApprovedBy)
			// BR-NFS-009: UpdateNameData → publishes customer.name.updated event.
			// BR-NFS-010: DOB is never modified.
			var updateResult activities.UpdateNameDataResult
			if err := workflow.ExecuteActivity(actCtx, act.UpdateNameData,
				activities.UpdateNameDataInput{
					RequestID: input.RequestID,
					CustomerID: input.CustomerID,
					UpdatedBy: approvalPayload.ApprovedBy,
				},
			).Get(ctx, &updateResult); err != nil {
				return err
			}
			// Generate receipt.
			_ = workflow.ExecuteActivity(actCtx, act.GenerateNameAckReceipt,
				activities.GenerateAckReceiptInput{
					RequestID:   input.RequestID,
					RequestType: "NAME_CHANGE",
				},
			).Get(ctx, nil)
			// Notify Policy Management of completed name change.
			payload := map[string]interface{}{}
			if updateResult.NewSalutation != nil {
				payload["salutation"] = *updateResult.NewSalutation
			}
			if updateResult.NewFirstName != nil {
				payload["first_name"] = *updateResult.NewFirstName
			}
			if updateResult.NewMiddleName != nil {
				payload["middle_name"] = *updateResult.NewMiddleName
			}
			if updateResult.NewLastName != nil {
				payload["last_name"] = *updateResult.NewLastName
			}
			
			_ = workflow.ExecuteActivity(actCtx, "NotifyPolicyManagement", activities.NotifyPMInput{
				RequestID:   input.RequestID,
				CustomerID:  input.CustomerID,
				RequestType: "NAME_CHANGE",
				Outcome:     "APPROVED",
				ChangePayload: mustMarshalJSON(payload),
			}).Get(ctx, nil)

			// Terminal — exit loop.
			logger.Info("WF-NFS-004 ManualNameChangeWorkflow COMPLETED", "requestID", input.RequestID)
			return nil

		case "REJECT":
			// BR-NFS-012: IN_PROGRESS → REJECTED (terminal).
			// BR-NFS-016: audit log created.
			logger.Info("WF-NFS-004: CPC rejected", "requestID", input.RequestID, "by", approvalPayload.ApprovedBy)
			if approvalPayload.RejectionReason == nil || *approvalPayload.RejectionReason == "" {
				return fmt.Errorf("rejection reason is required for REJECT decision")
			}
			rejReason := *approvalPayload.RejectionReason
			_ = workflow.ExecuteActivity(actCtx, act.UpdateNameStatus,
				activities.UpdateStatusInput{
					RequestID: input.RequestID,
					NewStatus: "REJECTED",
					UpdatedBy: approvalPayload.ApprovedBy,
					Reason:    &rejReason,
				},
			).Get(ctx, nil)

			// Notify Policy Management of rejected name change.
			_ = workflow.ExecuteActivity(actCtx, "NotifyPolicyManagement", activities.NotifyPMInput{
				RequestID:   input.RequestID,
				CustomerID:  input.CustomerID,
				RequestType: "NAME_CHANGE",
				Outcome:     "REJECTED",
				ChangePayload: mustMarshalJSON(map[string]interface{}{
					"salutation":  "",
					"first_name":  "",
					"middle_name": "",
					"last_name":   "",
				}),
			}).Get(ctx, nil)

			// Terminal — exit loop.
			return nil

		case "SEND_BACK":
			// BR-NFS-012: IN_PROGRESS → PENDING_DOCUMENTS (non-terminal).
			// BR-NFS-014: customer re-uploads via secure link generated by CORE-010.
			// Re-wait for documents_submitted signal then re-assign to CPC.
			// Loop continues → fresh approval wait at top.
			logger.Info("WF-NFS-004: CPC sent back for documents", "requestID", input.RequestID)
			_ = workflow.ExecuteActivity(actCtx, act.UpdateNameStatus,
				activities.UpdateStatusInput{
					RequestID: input.RequestID,
					NewStatus: "PENDING_DOCUMENTS",
					UpdatedBy: approvalPayload.ApprovedBy,
				},
			).Get(ctx, nil)

			// Wait for customer to re-submit documents before re-entering approval wait.
			resubmitCh := workflow.GetSignalChannel(ctx, SignalDocumentsSubmitted)
			var resubmitPayload activities.DocumentsSubmittedPayload
			var resubmitWithdrawn bool
			sel2 := workflow.NewSelector(ctx)
			sel2.AddReceive(resubmitCh, func(c workflow.ReceiveChannel, _ bool) {
				c.Receive(ctx, &resubmitPayload)
			})
			sel2.AddReceive(withdrawCh, func(c workflow.ReceiveChannel, _ bool) {
				c.Receive(ctx, nil)
				resubmitWithdrawn = true
			})
			sel2.Select(ctx)
			if resubmitWithdrawn {
				logger.Info("WF-NFS-004: withdrawal signal received during SEND_BACK — exiting", "requestID", input.RequestID)
				return nil
			}

			// Re-assign to CPC with fresh context.
			_ = workflow.ExecuteActivity(actCtx, act.AssignNameToCPC,
				activities.AssignToCPCInput{
					RequestID:   input.RequestID,
					OfficeCode:  &input.OfficeCode,
					AssignedBy:  input.InitiatedBy,
					Priority:    1,
					SLADeadline: &slaDeadline,
				},
			).Get(ctx, nil)
			// Continue loop → back to top, new approval wait.
		}
	}
}

// func ManualNameChangeWorkflow(ctx workflow.Context, input NameChangeWorkflowInput) error {
// 	logger := workflow.GetLogger(ctx)
// 	logger.Info("WF-NFS-004 ManualNameChangeWorkflow started",
// 		"requestID", input.RequestID,
// 		"ticketNumber", input.TicketNumber,
// 		"slaDays", input.SLADays,
// 	)

// 	actCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
// 		StartToCloseTimeout: 30 * time.Second,
// 		RetryPolicy: &temporal.RetryPolicy{
// 			MaximumAttempts: 3,
// 			InitialInterval: 2 * time.Second,
// 		},
// 	})

// 	var act *activities.NameChangeActivities

// 	// Step 1: Store workflow state for HTTP signaling.
// 	var storeResult activities.StoreWorkflowStateResult
// 	if err := workflow.ExecuteActivity(actCtx, act.StoreWorkflowStateForName,
// 		activities.StoreWorkflowStateInput{RequestID: input.RequestID},
// 	).Get(ctx, &storeResult); err != nil {
// 		return err
// 	}

// 	// Step 2: Wait for documents_submitted signal from CORE-007.
// 	// This allows the customer to upload documents before CPC assignment.
// 	docsCh := workflow.GetSignalChannel(ctx, SignalDocumentsSubmitted)
// 	var docsPayload activities.DocumentsSubmittedPayload
// 	docsCh.Receive(ctx, &docsPayload)

// 	// Step 3: Assign to CPC pool (BR-NFS-002: assignment based on office).
// 	// BATCH (TX): AssignNameToCPC — 3-op TX batch.
// 	var assignResult activities.AssignToCPCResult
// 	if err := workflow.ExecuteActivity(actCtx, act.AssignNameToCPC,
// 		activities.AssignToCPCInput{
// 			RequestID:  input.RequestID,
// 			OfficeCode: &input.OfficeCode,
// 		},
// 	).Get(ctx, &assignResult); err != nil {
// 		return err
// 	}

// 	// Step 4: Wait for CPC approval_decision signal with SLA deadline + 7-day grace.
// 	if input.SLADays == 0 {
// 		input.SLADays = 15
// 	}
// 	slaDeadline := workflow.Now(ctx).AddDate(0, 0, input.SLADays)
// 	gracePeriod := 7 * 24 * time.Hour
// 	totalWait := slaDeadline.Sub(workflow.Now(ctx)) + gracePeriod

// 	approvalCh := workflow.GetSignalChannel(ctx, SignalApprovalDecision)
// 	var approvalPayload ApprovalDecisionPayload

// 	slaTimer := workflow.NewTimer(ctx, time.Duration(input.SLADays)*24*time.Hour)
// 	totalTimer := workflow.NewTimer(ctx, totalWait)
// 	selector := workflow.NewSelector(ctx)
// 	approved := false

// 	// SLA breach alert (non-terminal — keep waiting with escalation).
// 	selector.AddFuture(slaTimer, func(f workflow.Future) {
// 		logger.Warn("WF-NFS-004: SLA deadline reached — escalating",
// 			"requestID", input.RequestID,
// 			"slaDays", input.SLADays,
// 		)
// 		_ = workflow.ExecuteActivity(actCtx, act.EscalateNameRequest,
// 			activities.EscalateInput{RequestID: input.RequestID},
// 		).Get(ctx, nil)
// 	})

// 	// Total deadline (SLA + 7-day grace) — auto-close.
// 	selector.AddFuture(totalTimer, func(f workflow.Future) {})

// 	// Approval decision signal.
// 	selector.AddReceive(approvalCh, func(c workflow.ReceiveChannel, more bool) {
// 		c.Receive(ctx, &approvalPayload)
// 		approved = true
// 	})

// 	selector.Select(ctx)

// 	// If SLA breach was selected first (not approval), continue waiting for approval or total deadline.
// 	if !approved && workflow.Now(ctx).Before(slaDeadline.Add(gracePeriod)) {
// 		// Re-wait for approval signal after escalation.
// 		selector2 := workflow.NewSelector(ctx)
// 		selector2.AddReceive(approvalCh, func(c workflow.ReceiveChannel, more bool) {
// 			c.Receive(ctx, &approvalPayload)
// 			approved = true
// 		})
// 		// NOTE: selector.Select fires for exactly ONE ready case per call (Temporal guarantee).
// 		// Case 1: approvalCh fires → approved=true → skip re-wait, process decision.
// 		// Case 2: slaTimer fires → escalate → re-wait via selector2 for approvalCh or totalTimer.
// 		// Case 3: totalTimer fires → auto-close. slaTimer escalation is skipped in this case
// 		//         which is acceptable — if total deadline is reached, escalation is moot.
// 		// Do NOT collapse selector2 into selector1 — re-wait must use the same totalTimer
// 		// future so the deadline is not reset after escalation.
// 		selector2.AddFuture(totalTimer, func(f workflow.Future) {})
// 		selector2.Select(ctx)
// 	}

// 	if !approved {
// 		// Total deadline reached — auto-close (DOCUMENTS_EXPIRED as terminal state).
// 		logger.Warn("WF-NFS-004: total deadline reached — auto-closing", "requestID", input.RequestID)
// 		_ = workflow.ExecuteActivity(actCtx, act.UpdateNameStatus,
// 			activities.UpdateStatusInput{
// 				RequestID: input.RequestID,
// 				NewStatus: "DOCUMENTS_EXPIRED",
// 				UpdatedBy: "system",
// 			},
// 		).Get(ctx, nil)
// 		return nil
// 	}

// 	// Step 5: Process approval decision (BR-NFS-012).
// 	switch approvalPayload.Decision {
// 	case "APPROVE":
// 		logger.Info("WF-NFS-004: CPC approved", "requestID", input.RequestID, "by", approvalPayload.ApprovedBy)
// 		// BR-NFS-009: UpdateNameData → publishes customer.name.updated event.
// 		// BR-NFS-010: DOB is never modified.
// 		if err := workflow.ExecuteActivity(actCtx, act.UpdateNameData,
// 			activities.UpdateAddressDataInput{
// 				RequestID: input.RequestID,
// 				UpdatedBy: approvalPayload.ApprovedBy,
// 			},
// 		).Get(ctx, nil); err != nil {
// 			return err
// 		}
// 		// Generate receipt.
// 		_ = workflow.ExecuteActivity(actCtx, act.GenerateNameAckReceipt,
// 			activities.GenerateAckReceiptInput{
// 				RequestID:   input.RequestID,
// 				RequestType: "NAME_CHANGE",
// 			},
// 		).Get(ctx, nil)

// 	case "REJECT":
// 		// BR-NFS-012: IN_PROGRESS → REJECTED (terminal).
// 		// BR-NFS-016: audit log created.
// 		logger.Info("WF-NFS-004: CPC rejected", "requestID", input.RequestID, "by", approvalPayload.ApprovedBy)
// 		rejReason := ""
// 		if approvalPayload.RejectionReason != nil {
// 			rejReason = *approvalPayload.RejectionReason
// 		}
// 		_ = workflow.ExecuteActivity(actCtx, act.UpdateNameStatus,
// 			activities.UpdateStatusInput{
// 				RequestID: input.RequestID,
// 				NewStatus: "REJECTED",
// 				UpdatedBy: approvalPayload.ApprovedBy,
// 				Reason:    &rejReason,
// 			},
// 		).Get(ctx, nil)

// 	case "SEND_BACK":
// 		// BR-NFS-012: IN_PROGRESS → PENDING_DOCUMENTS.
// 		// BR-NFS-014: missing doc link generated, 7-day expiry.
// 		logger.Info("WF-NFS-004: CPC sent back for documents", "requestID", input.RequestID)
// 		_ = workflow.ExecuteActivity(actCtx, act.UpdateNameStatus,
// 			activities.UpdateStatusInput{
// 				RequestID: input.RequestID,
// 				NewStatus: "PENDING_DOCUMENTS",
// 				UpdatedBy: approvalPayload.ApprovedBy,
// 			},
// 		).Get(ctx, nil)
// 	}

// 	logger.Info("WF-NFS-004 ManualNameChangeWorkflow completed", "requestID", input.RequestID, "decision", approvalPayload.Decision)
// 	return nil
// }

// ===========================================================================
// WF-NFS-005: WithdrawalWorkflow
// ===========================================================================

// WithdrawalWorkflow handles customer-initiated withdrawal of NFS requests.
// Triggered by CORE-009 POST /nfs/requests/:request_id/withdraw.
//
// FR-NFS-007: Withdrawal support.
// BR-NFS-013: Only allowed for CREATED / PENDING_DOCUMENTS / PENDING_APPROVAL.
//
//	NOT allowed for IN_PROGRESS, COMPLETED, REJECTED, WITHDRAWN, DOCUMENTS_EXPIRED.
//
// BR-NFS-013: If partial_processing_flag=TRUE → manual CPC approval required.
// BR-NFS-016: Audit log created (action_type=WITHDRAWN).
//
// State machine:
//
//	partial_processing_flag=FALSE → auto-WITHDRAWN (no human approval)
//	partial_processing_flag=TRUE  → PENDING_WITHDRAWAL_APPROVAL
//	                              → [Signal: withdrawal_decision]
//	                              → APPROVED → WITHDRAWN
//	                              → REJECTED → back to original status
func WithdrawalWorkflow(ctx workflow.Context, input WithdrawalWorkflowInput) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("WF-NFS-005 WithdrawalWorkflow started",
		"requestID", input.RequestID,
		"partialProcessing", input.PartialProcessingFlag,
	)

	actCtx := workflow.WithActivityOptions(ctx, withdrawalActivityOpts)

	// Use shared activities (ServiceRequest + AuditLog repos).
	var srAct *activities.WithdrawalActivities
	// Step 1: Check withdrawal eligibility (BR-NFS-013).
	var eligResult activities.CheckWithdrawalEligibilityResult
	if err := workflow.ExecuteActivity(actCtx, srAct.CheckWithdrawalEligibility,
		activities.CheckWithdrawalEligibilityInput{
			RequestID: input.RequestID,
		},
	).Get(ctx, &eligResult); err != nil {
		return err
	}
	if !eligResult.Eligible {
		logger.Error("WF-NFS-005: withdrawal not eligible",
			"requestID", input.RequestID,
			"reason", eligResult.Reason,
		)
		return temporal.NewApplicationError(eligResult.Reason, "ERR-NFS-SR-002")
	}

	if !input.PartialProcessingFlag {
		// Auto-approved withdrawal (BR-NFS-013: no partial processing).
		if err := workflow.ExecuteActivity(actCtx, srAct.ProcessWithdrawal,
			activities.ProcessWithdrawalInput{
				RequestID:        input.RequestID,
				WithdrawalType:   "AUTO",
				WithdrawalReason: input.WithdrawalReason,
				RequestedBy:      input.RequestedBy,
			},
		).Get(ctx, nil); err != nil {
			return err
		}
		logger.Info("WF-NFS-005: auto-withdrawal completed", "requestID", input.RequestID)
		// Signal parent ManualNameChangeWorkflow to exit.
		parentWfID := fmt.Sprintf("nfs-name-MANUAL-%s", input.RequestID)
		if err := workflow.SignalExternalWorkflow(ctx, parentWfID, "", "withdrawal_requested", nil).Get(ctx, nil); err != nil {
			logger.Warn("WF-NFS-005: could not signal parent workflow", "parentWfID", parentWfID, "error", err)
		}
		return nil
	}

	// Manual CPC approval path (partial_processing_flag=TRUE).
	// Move to PENDING_WITHDRAWAL_APPROVAL — signal: withdrawal_decision.
	if err := workflow.ExecuteActivity(actCtx, srAct.UpdateStatus,
		activities.UpdateStatusInput{
			RequestID: input.RequestID,
			NewStatus: "PENDING_WITHDRAWAL_APPROVAL",
			UpdatedBy: input.RequestedBy,
		},
	).Get(ctx, nil); err != nil {
		return err
	}

	// Wait for CPC withdrawal decision signal (30-day max).
	withdrawalCh := workflow.GetSignalChannel(ctx, SignalWithdrawalDecision)
	var withdrawPayload ApprovalDecisionPayload
	timer := workflow.NewTimer(ctx, 30*24*time.Hour)
	selector := workflow.NewSelector(ctx)
	decided := false

	selector.AddReceive(withdrawalCh, func(c workflow.ReceiveChannel, more bool) {
		c.Receive(ctx, &withdrawPayload)
		decided = true
	})
	selector.AddFuture(timer, func(f workflow.Future) {})
	selector.Select(ctx)

	if !decided {
		logger.Warn("WF-NFS-005: withdrawal decision timeout — auto-rejecting", "requestID", input.RequestID)
		return nil
	}

	if withdrawPayload.Decision == "APPROVE" {
		if err := workflow.ExecuteActivity(actCtx, srAct.ProcessWithdrawal,
			activities.ProcessWithdrawalInput{
				RequestID:        input.RequestID,
				WithdrawalType:   "MANUAL",
				WithdrawalReason: input.WithdrawalReason,
				RequestedBy:      input.RequestedBy,
				ApprovedBy:       parseStringToInt64Ptr(withdrawPayload.ApprovedBy),
			},
		).Get(ctx, nil); err != nil {
			return err
		}
		logger.Info("WF-NFS-005: manual withdrawal approved", "requestID", input.RequestID)
	} else {
		logger.Info("WF-NFS-005: withdrawal rejected by CPC", "requestID", input.RequestID)
	}

	return nil
}

// parseStringToInt64Ptr parses a string to *int64, returns nil if empty or parsing fails
func parseStringToInt64Ptr(s string) *int64 {
	if s == "" {
		return nil
	}
	val, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return nil
	}
	return &val
}

// mustMarshalJSON marshals v to JSON, returning nil on error.
func mustMarshalJSON(v interface{}) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	return b
}
