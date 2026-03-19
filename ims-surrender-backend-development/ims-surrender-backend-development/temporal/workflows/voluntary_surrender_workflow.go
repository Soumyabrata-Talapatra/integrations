package workflows

import (
	"encoding/json"
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"gitlab.cept.gov.in/it-2.0-policy/surrender-service/temporal/activities"
)

// ============================================================
// SurrenderProcessingWorkflow
//
// Entry point invoked by Policy Management's
// PolicyLifecycleWorkflow via ExecuteChildWorkflow.
//
// Workflow ID pattern : sur-{PM-idempotency-key}
// Task Queue          : surrender-tq
// Input               : SurrenderProcessingInput (PM contract)
//
// Flow:
//  1. IndexSurrenderActivity      — create surrender_request in DB,
//                                   store temporal_workflow_id so
//                                   DE/QC/Approval handlers can signal here
//  2. ValidateEligibilityActivity — business-level rules (product
//                                   eligibility, maturity check).
//                                   Policy state gate was already
//                                   checked by PM before dispatch.
//  3. Wait "de-completed" signal  → SubmitDEActivity
//  4. Wait "qc-completed" signal  → SubmitQCActivity
//  5. Wait "approval-completed"   → SubmitApprovalActivity
//  6. CalculateSurrenderValueActivity
//  7. ProcessPaymentActivity
//  8. UpdatePolicyStatusActivity
//  9. SignalPMWorkflowActivity    — send "surrender-completed" back to
//                                   PM's PolicyLifecycleWorkflow
// ============================================================

func SurrenderProcessingWorkflow(ctx workflow.Context, input SurrenderProcessingInput) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("SurrenderProcessingWorkflow started",
		"RequestID", input.RequestID,
		"PolicyNumber", input.PolicyNumber,
		"PMServiceRequestID", input.ServiceRequestID,
	)

	// Current Temporal workflow ID assigned by PM (e.g. "sur-{uuid}").
	// Passed to IndexSurrenderActivity so it can be stored in DB,
	// enabling DE/QC/Approval handlers to signal us by surrender_request_id.
	workflowID := workflow.GetInfo(ctx).WorkflowExecution.ID

	// Parse the original PM request payload for surrender-specific fields.
	var pmPayload PMSurrenderRequestPayload
	if len(input.RequestPayload) > 0 {
		_ = json.Unmarshal(input.RequestPayload, &pmPayload)
	}

	// Standard short-activity options used for DB operations.
	stdOpts := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts:    3,
			InitialInterval:    1 * time.Second,
			BackoffCoefficient: 2.0,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, stdOpts)

	// ── Step 1: Index ────────────────────────────────────────────────────────
	// Create the surrender_request record in the surrender service DB.
	// Stores temporal_workflow_id so signal-based handlers can find us.
	logger.Info("Step 1: Indexing surrender request")
	var indexResult activities.IndexSurrenderResult
	err := workflow.ExecuteActivity(ctx, activities.IndexSurrenderActivity, activities.IndexSurrenderInput{
		PolicyNumber:            input.PolicyNumber,
		SurrenderRequestChannel: pmPayload.SourceChannel,
		TemporalWorkflowID:      workflowID,
		PMServiceRequestID:      input.ServiceRequestID,
		PMPolicyDBID:            input.PolicyDBID,
		Stage_name:              "Indexed",
	}).Get(ctx, &indexResult)
	if err != nil {
		logger.Error("Step 1 failed: could not index surrender request", "error", err)
		signalPMBack(ctx, input, OutcomeRejected, StateTransitionSurrenderRejected,
			fmt.Sprintf("failed to create surrender request record: %v", err))
		return err
	}
	surrenderRequestID := indexResult.ServiceRequestID
	logger.Info("Step 1 complete", "surrender_request_id", surrenderRequestID)

	// ── Step 2: Business-level eligibility validation ─────────────────────
	// PM already verified the policy state gate (ACTIVE/VOID_LAPSE/etc.).
	// This step checks surrender-domain specific rules:
	//   - Product must not be in the ineligible list (AEA, AEA-10, GY)
	//   - Policy must not have passed its maturity date
	logger.Info("Step 2: Validating surrender eligibility")
	var eligibility activities.ValidateEligibilityResult
	err = workflow.ExecuteActivity(ctx, activities.ValidateEligibilityActivity, activities.ValidateEligibilityInput{
		PolicyID:           input.PolicyNumber,
		SurrenderRequestID: surrenderRequestID,
		ProductCode:        input.ProductCode,
		MaturityDate:       input.MaturityDate,
	}).Get(ctx, &eligibility)
	if err != nil {
		logger.Error("Step 2 failed: eligibility check error", "error", err)
		signalPMBack(ctx, input, OutcomeRejected, StateTransitionSurrenderRejected, err.Error())
		return err
	}
	if !eligibility.Eligible {
		logger.Info("Step 2: policy not eligible for surrender", "reasons", eligibility.Reasons)
		signalPMBack(ctx, input, OutcomeRejected, StateTransitionSurrenderRejected,
			fmt.Sprintf("not eligible: %v", eligibility.Reasons))
		return fmt.Errorf("policy not eligible for surrender: %v", eligibility.Reasons)
	}
	logger.Info("Step 2 complete: policy is eligible")

	// ── Step 3: Data Entry ───────────────────────────────────────────────────
	// Wait for the CPC officer to complete Data Entry via
	// PUT /v1/surrender/submit-de (handler signals "de-completed" here).
	logger.Info("Step 3: Waiting for Data Entry completion signal")
	var deInput activities.SubmitDEInput
	deSelector := workflow.NewSelector(ctx)

	deChannel := workflow.GetSignalChannel(ctx, "de-completed")
	deSelector.AddReceive(deChannel, func(c workflow.ReceiveChannel, more bool) {
		c.Receive(ctx, &deInput)
		logger.Info("Received de-completed signal", "surrender_request_id", deInput.SurrenderRequestID)
	})

	deTimeout := workflow.NewTimer(ctx, 7*24*time.Hour)
	deSelector.AddFuture(deTimeout, func(f workflow.Future) {
		logger.Warn("Data Entry timeout reached (7 days)")
	})
	deSelector.Select(ctx)

	if deInput.SurrenderRequestID == "" {
		logger.Error("Step 3: Data Entry timeout — no signal received within 7 days")
		signalPMBack(ctx, input, OutcomeTimeout, StateTransitionSurrenderTimeout,
			"Data Entry not completed within 7 days")
		return fmt.Errorf("data entry timeout")
	}

	logger.Info("Step 3: Executing SubmitDEActivity")
	var deResult activities.SubmitDEResult
	err = workflow.ExecuteActivity(ctx, activities.SubmitDEActivity, deInput).Get(ctx, &deResult)
	if err != nil {
		logger.Error("Step 3: SubmitDEActivity failed", "error", err)
		signalPMBack(ctx, input, OutcomeRejected, StateTransitionSurrenderRejected, err.Error())
		return err
	}
	if !deResult.Success {
		logger.Error("Step 3: DE submission rejected", "message", deResult.Message)
		signalPMBack(ctx, input, OutcomeRejected, StateTransitionSurrenderRejected, deResult.Message)
		return fmt.Errorf("DE submission failed: %s", deResult.Message)
	}
	logger.Info("Step 3 complete")

	// ── Step 4: Quality Check ─────────────────────────────────────────────
	// Wait for QC officer via PUT /v1/surrender/submit-qc.
	logger.Info("Step 4: Waiting for Quality Check completion signal")
	var qcInput activities.SubmitDEInput
	qcSelector := workflow.NewSelector(ctx)

	qcChannel := workflow.GetSignalChannel(ctx, "qc-completed")
	qcSelector.AddReceive(qcChannel, func(c workflow.ReceiveChannel, more bool) {
		c.Receive(ctx, &qcInput)
		logger.Info("Received qc-completed signal", "surrender_request_id", qcInput.SurrenderRequestID)
	})

	qcTimeout := workflow.NewTimer(ctx, 7*24*time.Hour)
	qcSelector.AddFuture(qcTimeout, func(f workflow.Future) {
		logger.Warn("Quality Check timeout reached (7 days)")
	})
	qcSelector.Select(ctx)

	if qcInput.SurrenderRequestID == "" {
		logger.Error("Step 4: Quality Check timeout — no signal received within 7 days")
		signalPMBack(ctx, input, OutcomeTimeout, StateTransitionSurrenderTimeout,
			"Quality Check not completed within 7 days")
		return fmt.Errorf("quality check timeout")
	}

	logger.Info("Step 4: Executing SubmitQCActivity")
	var qcResult activities.SubmitQCResult
	err = workflow.ExecuteActivity(ctx, activities.SubmitQCActivity, qcInput).Get(ctx, &qcResult)
	if err != nil {
		logger.Error("Step 4: SubmitQCActivity failed", "error", err)
		signalPMBack(ctx, input, OutcomeRejected, StateTransitionSurrenderRejected, err.Error())
		return err
	}
	if !qcResult.Success {
		logger.Error("Step 4: QC submission rejected", "message", qcResult.Message)
		signalPMBack(ctx, input, OutcomeRejected, StateTransitionSurrenderRejected, qcResult.Message)
		return fmt.Errorf("QC submission failed: %s", qcResult.Message)
	}
	logger.Info("Step 4 complete")

	// ── Step 5: Approval ──────────────────────────────────────────────────
	// Wait for approver via PUT /v1/surrender/submit-approval.
	logger.Info("Step 5: Waiting for Approval completion signal")
	var approvalInput activities.SubmitDEInput
	approvalSelector := workflow.NewSelector(ctx)

	approvalChannel := workflow.GetSignalChannel(ctx, "approval-completed")
	approvalSelector.AddReceive(approvalChannel, func(c workflow.ReceiveChannel, more bool) {
		c.Receive(ctx, &approvalInput)
		logger.Info("Received approval-completed signal", "surrender_request_id", approvalInput.SurrenderRequestID)
	})

	approvalTimeout := workflow.NewTimer(ctx, 30*24*time.Hour)
	approvalSelector.AddFuture(approvalTimeout, func(f workflow.Future) {
		logger.Warn("Approval timeout reached (30 days)")
	})
	approvalSelector.Select(ctx)

	if approvalInput.SurrenderRequestID == "" {
		logger.Error("Step 5: Approval timeout — no signal received within 30 days")
		signalPMBack(ctx, input, OutcomeTimeout, StateTransitionSurrenderTimeout,
			"Approval not completed within 30 days")
		return fmt.Errorf("approval timeout")
	}

	logger.Info("Step 5: Executing SubmitApprovalActivity")
	var approvalResult activities.SubmitApprovalResult
	err = workflow.ExecuteActivity(ctx, activities.SubmitApprovalActivity, approvalInput).Get(ctx, &approvalResult)
	if err != nil {
		logger.Error("Step 5: SubmitApprovalActivity failed", "error", err)
		signalPMBack(ctx, input, OutcomeRejected, StateTransitionSurrenderRejected, err.Error())
		return err
	}
	if !approvalResult.Success {
		logger.Error("Step 5: Approval submission rejected", "message", approvalResult.Message)
		signalPMBack(ctx, input, OutcomeRejected, StateTransitionSurrenderRejected, approvalResult.Message)
		return fmt.Errorf("approval submission failed: %s", approvalResult.Message)
	}
	if approvalResult.Status == "REJECTED" {
		logger.Info("Step 5: Surrender request rejected by approver")
		signalPMBack(ctx, input, OutcomeRejected, StateTransitionSurrenderRejected,
			"surrender rejected during approval")
		return nil
	}
	logger.Info("Step 5 complete", "approval_status", approvalResult.Status)

	// ── Step 6: Calculate Surrender Value ────────────────────────────────
	logger.Info("Step 6: Calculating surrender value")
	var calcResult activities.CalculateSurrenderValueResult
	err = workflow.ExecuteActivity(ctx, activities.CalculateSurrenderValueActivity,
		activities.CalculateSurrenderValueInput{
			SurrenderRequestID: surrenderRequestID,
			PolicyID:           input.PolicyNumber,
		},
	).Get(ctx, &calcResult)
	if err != nil {
		logger.Error("Step 6: calculation failed", "error", err)
		signalPMBack(ctx, input, OutcomeRejected, StateTransitionSurrenderRejected, err.Error())
		return err
	}
	logger.Info("Step 6 complete", "net_surrender_value", calcResult.NetSurrenderValue)

	// ── Step 7: Process Payment ──────────────────────────────────────────
	logger.Info("Step 7: Processing payment")
	paymentOpts := workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts:    5,
			InitialInterval:    2 * time.Second,
			BackoffCoefficient: 2.0,
		},
	}
	paymentCtx := workflow.WithActivityOptions(ctx, paymentOpts)
	var paymentResult activities.ProcessPaymentResult
	err = workflow.ExecuteActivity(paymentCtx, activities.ProcessPaymentActivity,
		activities.ProcessPaymentInput{
			SurrenderRequestID: surrenderRequestID,
			Amount:             calcResult.NetSurrenderValue,
		},
	).Get(paymentCtx, &paymentResult)
	if err != nil {
		logger.Error("Step 7: payment failed", "error", err)
		signalPMBack(ctx, input, OutcomeRejected, StateTransitionSurrenderRejected, err.Error())
		return err
	}
	logger.Info("Step 7 complete", "payment_reference", paymentResult.PaymentReference)

	// ── Step 8: Update Policy Status ─────────────────────────────────────
	logger.Info("Step 8: Updating policy status")
	var policyUpdateResult activities.UpdatePolicyStatusResult
	err = workflow.ExecuteActivity(ctx, activities.UpdatePolicyStatusActivity,
		activities.UpdatePolicyStatusInput{
			PolicyID:           input.PolicyNumber,
			SurrenderRequestID: surrenderRequestID,
			NewStatus:          calcResult.PredictedDisposition,
		},
	).Get(ctx, &policyUpdateResult)
	if err != nil {
		// Payment already processed — log but do not fail the workflow.
		// PM will still receive APPROVED so it can close the financial lock.
		logger.Error("Step 8: policy status update failed (non-fatal, payment already processed)",
			"error", err)
	} else {
		logger.Info("Step 8 complete", "new_policy_status", policyUpdateResult.NewStatus)
	}

	// ── Step 9: Signal PM back ───────────────────────────────────────────
	// Notify PM's PolicyLifecycleWorkflow that surrender completed.
	// PM uses this to release the financial lock and transition:
	//   PENDING_SURRENDER → SURRENDERED
	logger.Info("Step 9: Signalling PM with surrender-completed (APPROVED)")
	signalPMBack(ctx, input, OutcomeApproved, StateTransitionSurrendered, "")

	logger.Info("SurrenderProcessingWorkflow completed successfully",
		"surrender_request_id", surrenderRequestID,
		"payment_reference", paymentResult.PaymentReference,
	)
	return nil
}

// signalPMBack sends the "surrender-completed" signal to PM's PLW workflow
// (plw-{policyNumber}) with the given outcome via SignalPMWorkflowActivity.
// It is called in every terminal path (success, rejection, timeout).
func signalPMBack(
	ctx workflow.Context,
	input SurrenderProcessingInput,
	outcome string,
	stateTransition string,
	reason string,
) {
	logger := workflow.GetLogger(ctx)
	plwWorkflowID := "plw-" + input.PolicyNumber

	// Build optional outcome payload with failure reason when applicable.
	var outcomePayload json.RawMessage
	if reason != "" {
		raw, _ := json.Marshal(map[string]string{"reason": reason})
		outcomePayload = raw
	}

	signalOpts := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts:    5,
			InitialInterval:    2 * time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    30 * time.Second,
		},
	}
	signalCtx := workflow.WithActivityOptions(ctx, signalOpts)

	err := workflow.ExecuteActivity(signalCtx, activities.SignalPMWorkflowActivity,
		activities.SignalPMWorkflowInput{
			PMWorkflowID:    plwWorkflowID,
			SignalName:      "surrender-completed",
			RequestID:       input.RequestID,
			RequestType:     "SURRENDER",
			Outcome:         outcome,
			StateTransition: stateTransition,
			OutcomePayload:  outcomePayload,
		},
	).Get(signalCtx, nil)

	if err != nil {
		// Log but do not propagate — the workflow itself reached a terminal state.
		// PM's timeout mechanism will eventually clean up if the signal never arrives.
		logger.Error("signalPMBack: failed to signal PM workflow",
			"plw_workflow_id", plwWorkflowID,
			"outcome", outcome,
			"error", err,
		)
	} else {
		logger.Info("signalPMBack: sent surrender-completed signal",
			"plw_workflow_id", plwWorkflowID,
			"outcome", outcome,
			"state_transition", stateTransition,
		)
	}
}
