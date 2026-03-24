package workflows

import (
	"fmt"
	"policy-issue-service/core/domain"
	"policy-issue-service/workflows/activities"
	"strconv"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

const (
	SignalQRDecision       = "qr-decision"
	SignalApproverDecision = "approver-decision"
	SignalMedicalResult    = "medical-result"
	SignalCPCResubmit      = "cpc-resubmit"
	SignalSubmitForQC      = "submit-for-qc"
)

type SubmitForQCSignal struct {
	DataEntryID int64 `json:"data_entry_id"`
}

// PolicyIssuanceInput contains the input for policy issuance workflow
type PolicyIssuanceInput struct {
	ProposalID     int64  `json:"proposal_id"`
	ProposalNumber string `json:"proposal_number"`
	CustomerID     string `json:"customer_id"`
	ProductCode    string `json:"product_code"`
	// ProductCategory         string                  `json:"product_category"`
	PolicyType              domain.PolicyType       `json:"policy_type"`
	SumAssured              float64                 `json:"sum_assured"`
	PolicyTerm              int                     `json:"policy_term"`
	AgeAtEntry              int                     `json:"age_at_entry,omitempty"`
	Gender                  string                  `json:"gender,omitempty"`
	PremiumPaymentFrequency domain.PremiumFrequency `json:"premium_payment_frequency"`
	AgeProofType            string                  `json:"age_proof_type,omitempty"`
	InsuredState            string                  `json:"insured_state,omitempty"`
	PolicyNumber            string                  `json:"policy_number,omitempty"`
	ProposalDate            time.Time               `json:"proposal_date,omitempty"`
}

// PolicyIssuanceResult contains the result of policy issuance workflow
type PolicyIssuanceResult struct {
	ProposalNumber string `json:"proposal_number"`
	PolicyNumber   string `json:"policy_number,omitempty"`
	Status         string `json:"status"`
	Error          string `json:"error,omitempty"`
}

// QRDecisionSignal represents a QR decision signal
type QRDecisionSignal struct {
	Decision   string `json:"decision"` // APPROVED, REJECTED, RETURNED
	ReviewerID string `json:"reviewer_id"`
	Comments   string `json:"comments"`
}

// ApproverDecisionSignal represents an approver decision signal
type ApproverDecisionSignal struct {
	Decision   string `json:"decision"` // APPROVED, REJECTED
	ApproverID string `json:"approver_id"`
	Comments   string `json:"comments"`
}

// MedicalResultSignal represents a medical examination result signal
type MedicalResultSignal struct {
	Decision        string    `json:"decision"` // APPROVED, REJECTED
	RejectionReason string    `json:"rejection_reason,omitempty"`
	ExaminationDate time.Time `json:"examination_date,omitempty"`
	CertificateID   string    `json:"certificate_id,omitempty"`
}

// CPCResubmitSignal represents a CPC resubmit signal

// PolicyIssuanceWorkflow implements the end-to-end policy issuance workflow
// [WF-PI-001] Standard Policy Issuance Workflow
func PolicyIssuanceWorkflow(ctx workflow.Context, input PolicyIssuanceInput) (*PolicyIssuanceResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting Policy Issuance Workflow", "proposalID", input.ProposalID)

	result := &PolicyIssuanceResult{ProposalNumber: input.ProposalNumber}

	// Register query handler for status checks
	var currentStatus string
	err := workflow.SetQueryHandler(ctx, "QueryProposalStatus", func() (string, error) {
		return currentStatus, nil
	})
	if err != nil {
		return nil, err
	}

	// Activity options
	shortActivityOpts := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    30 * time.Second,
			MaximumAttempts:    3,
		},
	})

	externalCallOpts := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 2 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    2 * time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    5,
		},
	})

	// Suppress unused variable warnings - these will be used when activities are implemented
	_ = shortActivityOpts
	_ = externalCallOpts

	// Step 1: Validate Proposal
	currentStatus = "VALIDATING"
	logger.Info("Step 1: Validating proposal")
	validateInput := activities.ValidateProposalInput{ProposalNumber: input.ProposalNumber}
	var validateResult activities.ValidateProposalResult
	if err := workflow.ExecuteActivity(shortActivityOpts, "ValidateProposalActivity", validateInput).Get(ctx, &validateResult); err != nil {
		logger.Error("Validation failed", "error", err)
		result.Status = "VALIDATION_FAILED"
		result.Error = err.Error()
		return result, nil
	}
	if !validateResult.IsValid {
		result.Status = "VALIDATION_FAILED"
		result.Error = fmt.Sprintf("Validation errors: %v", validateResult.Errors)
		return result, nil
	}

	// Step 2: Check Eligibility
	currentStatus = "CHECKING_ELIGIBILITY"
	logger.Info("Step 2: Checking eligibility")
	eligibilityInput := activities.CheckEligibilityInput{
		ProposalNumber: input.ProposalNumber,
		AgeAtEntry:     input.AgeAtEntry,
	}
	var eligibilityResult activities.CheckEligibilityResult
	if err := workflow.ExecuteActivity(shortActivityOpts, "CheckEligibilityActivity", eligibilityInput).Get(ctx, &eligibilityResult); err != nil {
		logger.Error("Eligibility check failed", "error", err)
		result.Status = "ELIGIBILITY_FAILED"
		result.Error = err.Error()
		return result, nil
	}
	if !eligibilityResult.IsEligible {
		result.Status = "NOT_ELIGIBLE"
		result.Error = eligibilityResult.RejectReason
		return result, nil
	}

	// Step 3: Calculate Premium
	currentStatus = "CALCULATING_PREMIUM"
	logger.Info("Step 3: Calculating premium")
	premiumInput := activities.CalculatePremiumInput{
		ProposalNumber: input.ProposalNumber,
		// ProductCode:     input.ProductCode,
		// ProductCategory: input.ProductCategory,
		AgeAtEntry: input.AgeAtEntry,
		Gender:     input.Gender,
		PolicyTerm: input.PolicyTerm,
		SumAssured: input.SumAssured,
		Frequency:  input.PremiumPaymentFrequency,
	}
	var premiumResult activities.CalculatePremiumResult
	if err := workflow.ExecuteActivity(shortActivityOpts, "CalculatePremiumActivity", premiumInput).Get(ctx, &premiumResult); err != nil {
		logger.Error("Premium calculation failed", "error", err)
		result.Status = "PREMIUM_CALC_FAILED"
		result.Error = err.Error()
		return result, nil
	}

	// Step 3b: Save premium to proposal
	savePremiumInput := activities.SavePremiumInput{
		ProposalNumber: input.ProposalNumber,
		BasePremium:    premiumResult.BasePremium,
		Rebate:         premiumResult.Rebate,
		NetPremium:     premiumResult.NetPremium,
		GSTAmount:      premiumResult.GSTAmount,
		TotalPayable:   premiumResult.TotalPayable,
	}
	if err := workflow.ExecuteActivity(shortActivityOpts, "SavePremiumToProposalActivity", savePremiumInput).Get(ctx, nil); err != nil {
		logger.Error("Failed to save premium", "error", err)
		result.Status = "SAVE_PREMIUM_FAILED"
		result.Error = err.Error()
		return result, nil
	}

	// Step 4: Quality Review (Wait for QR Signal)
	currentStatus = "QC_PENDING"

	qrChan := workflow.GetSignalChannel(ctx, SignalQRDecision)

	qcDone := false

	for !qcDone {
		var qrDecision QRDecisionSignal

		selector := workflow.NewSelector(ctx)
		// qrChan := workflow.GetSignalChannel(ctx, SignalQRDecision)
		selector.AddReceive(qrChan, func(c workflow.ReceiveChannel, more bool) {
			c.Receive(ctx, &qrDecision)
		})

		timeout := workflow.NewTimer(ctx, 30*24*time.Hour)

		selector.AddFuture(timeout, func(f workflow.Future) {
			qrDecision.Decision = "TIMEOUT"
		})

		selector.Select(ctx)

		reviewerID, err := strconv.ParseInt(qrDecision.ReviewerID, 10, 64)
		if err != nil && qrDecision.Decision != "TIMEOUT" {
			return nil, fmt.Errorf("invalid reviewer id")
		}

		switch qrDecision.Decision {

		case "APPROVED":

			currentStatus = "QC_APPROVED"

			update := activities.UpdateStatusInput{
				ProposalNumber: input.ProposalNumber,
				Status:         domain.ProposalStatusQCApproved,
				Comments:       "QC approved: " + qrDecision.Comments,
				ChangedBy:      reviewerID,
			}

			err = workflow.ExecuteActivity(shortActivityOpts,
				"UpdateProposalStatusActivity", update).Get(ctx, nil)
			if err != nil {
				return nil, err
			}
			logger.Info("QC approved")

			qcDone = true  

		case "RETURNED":

			currentStatus = "QC_RETURNED"

			update := activities.UpdateStatusInput{
				ProposalNumber: input.ProposalNumber,
				Status:         domain.ProposalStatusQCReturned,
				Comments:       "QC returned to data entry: " + qrDecision.Comments,
				ChangedBy:      reviewerID,
			}

			err = workflow.ExecuteActivity(shortActivityOpts,
				"UpdateProposalStatusActivity", update).Get(ctx, nil)
			if err != nil {
				return nil, err
			}

			// Move proposal to DATA_ENTRY
			reenter := activities.UpdateStatusInput{
				ProposalNumber: input.ProposalNumber,
				Status:         domain.ProposalStatusDataEntry,
				Comments:       "Returned to data entry for correction",
				ChangedBy:      reviewerID,
			}

			err = workflow.ExecuteActivity(shortActivityOpts,
				"UpdateProposalStatusActivity", reenter).Get(ctx, nil)
			if err != nil {
				return nil, err
			}

			currentStatus = "DATA_ENTRY"

			logger.Info("Waiting for Data Entry to submit again")

			// WAIT for Submit For QC signal
			submitChan := workflow.GetSignalChannel(ctx, SignalSubmitForQC)

			var submit SubmitForQCSignal

			workflow.NewSelector(ctx).
				AddReceive(submitChan, func(c workflow.ReceiveChannel, more bool) {
					c.Receive(ctx, &submit)
				}).Select(ctx)

			logger.Info("Data Entry submitted again")

			// Move back to QC_PENDING
			reqc := activities.UpdateStatusInput{
				ProposalNumber: input.ProposalNumber,
				Status:         domain.ProposalStatusQCPending,
				Comments:       "Resubmitted for QC review",
				ChangedBy:      submit.DataEntryID,
			}

			err = workflow.ExecuteActivity(shortActivityOpts, "UpdateProposalStatusActivity",
				reqc).Get(ctx, nil)
			if err != nil {
				return nil, err
			}

			currentStatus = "QC_PENDING"

		case "REJECTED", "TIMEOUT":

			reject := activities.UpdateStatusInput{
				ProposalNumber: input.ProposalNumber,
				Status:         domain.ProposalStatusQCRejected,
				Comments:       "QC rejected: " + qrDecision.Comments,
				ChangedBy:      reviewerID,
			}

			_ = workflow.ExecuteActivity(shortActivityOpts,
				"UpdateProposalStatusActivity", reject).Get(ctx, nil)

			result.Status = "QC_REJECTED"
			logger.Info("Proposal rejected by QC")

			return result, nil
		}

	}

	// Step 5: Medical Underwriting (if required)
	currentStatus = "PENDING_MEDICAL"
	logger.Info("Step 5: Medical underwriting")

	// Request medical review
	medicalInput := activities.MedicalReviewInput{
		ProposalNumber: input.ProposalNumber,
		CustomerID:     0, // Parse from input.CustomerID
		AgeAtEntry:     input.AgeAtEntry,
		SumAssured:     input.SumAssured,
	}
	var medicalResult activities.MedicalReviewResult
	if err := workflow.ExecuteActivity(shortActivityOpts, "RequestMedicalReviewActivity", medicalInput).Get(ctx, &medicalResult); err != nil {
		logger.Error("Medical review request failed", "error", err)
		result.Status = "MEDICAL_REQUEST_FAILED"
		result.Error = err.Error()
		return result, nil
	}

	// If medical is required, wait for medical result signal
	if medicalResult.MedicalRequired {
		medicalSelector := workflow.NewSelector(ctx)
		medicalChan := workflow.GetSignalChannel(ctx, SignalMedicalResult)
		var medicalSignal MedicalResultSignal
		medicalSelector.AddReceive(medicalChan, func(c workflow.ReceiveChannel, more bool) {
			c.Receive(ctx, &medicalSignal)
		})
		medicalTimer := workflow.NewTimer(ctx, 30*24*time.Hour) // 30 day timeout
		medicalSelector.AddFuture(medicalTimer, func(f workflow.Future) {
			medicalSignal.Decision = "TIMEOUT"
		})
		medicalSelector.Select(ctx)

		if medicalSignal.Decision != "APPROVED" {
			// Persist MEDICAL_REJECTED transition
			medRejectInput := activities.UpdateStatusInput{
				ProposalNumber: input.ProposalNumber,
				Status:         domain.ProposalStatusMedicalRejected,
				Comments:       "Medical examination rejected: " + medicalSignal.RejectionReason,
			}
			_ = workflow.ExecuteActivity(shortActivityOpts, "UpdateProposalStatusActivity", medRejectInput).Get(ctx, nil)

			result.Status = "MEDICAL_REJECTED"
			return result, fmt.Errorf("medical examination rejected: %s", medicalSignal.RejectionReason)
		}

		// Persist MEDICAL_APPROVED transition
		medApproveInput := activities.UpdateStatusInput{
			ProposalNumber: input.ProposalNumber,
			Status:         domain.ProposalStatusMedicalApproved,
			Comments:       "Medical examination approved",
		}
		_ = workflow.ExecuteActivity(shortActivityOpts, "UpdateProposalStatusActivity", medApproveInput).Get(ctx, nil)
		currentStatus = "MEDICAL_APPROVED"
	}

	// Step 6: Approver Routing
	currentStatus = "APPROVAL_PENDING"
	logger.Info("Step 6: Waiting for approver decision")

	// Route to approver based on Sum Assured bands (BR-POL-016)
	routeInput := activities.RouteToApproverInput{
		ProposalNumber: input.ProposalNumber,
		SumAssured:     input.SumAssured,
	}
	var routeResult activities.RouteToApproverResult
	if err := workflow.ExecuteActivity(shortActivityOpts, "RouteToApproverActivity", routeInput).Get(ctx, &routeResult); err != nil {
		logger.Error("Failed to route to approver", "error", err)
		result.Status = "ROUTING_FAILED"
		result.Error = err.Error()
		return result, nil
	}
	logger.Info("Routed to approver", "level", routeResult.ApproverLevel, "role", routeResult.ApproverRole)

	// Wait for approver decision
	approvalSelector := workflow.NewSelector(ctx)
	approvalChan := workflow.GetSignalChannel(ctx, SignalApproverDecision)
	var approverDecision ApproverDecisionSignal
	approvalSelector.AddReceive(approvalChan, func(c workflow.ReceiveChannel, more bool) {
		c.Receive(ctx, &approverDecision)
	})
	approvalTimer := workflow.NewTimer(ctx, 7*24*time.Hour) // 7 day timeout
	approvalSelector.AddFuture(approvalTimer, func(f workflow.Future) {
		approverDecision.Decision = "TIMEOUT"
	})
	approvalSelector.Select(ctx)
	approverID, err := strconv.ParseInt(approverDecision.ApproverID, 10, 64)
	if err != nil {
		return result, fmt.Errorf("invalid approver_id")
	}
	if approverDecision.Decision != "APPROVED" {
		// Persist REJECTED transition (BR-POL-015)
		approverRejectInput := activities.UpdateStatusInput{
			ProposalNumber: input.ProposalNumber,
			Status:         domain.ProposalStatusRejected,
			Comments:       "Rejected by approver: " + approverDecision.Comments,
			ChangedBy:      approverID,
		}
		_ = workflow.ExecuteActivity(shortActivityOpts, "UpdateProposalStatusActivity", approverRejectInput).Get(ctx, nil)

		result.Status = "REJECTED"
		return result, fmt.Errorf("proposal rejected by approver")
	}

	// Persist APPROVED transition (BR-POL-015)
	approverApproveInput := activities.UpdateStatusInput{
		ProposalNumber: input.ProposalNumber,
		Status:         domain.ProposalStatusApproved,
		Comments:       "Approved: " + approverDecision.Comments,
		ChangedBy:      approverID,
	}
	_ = workflow.ExecuteActivity(shortActivityOpts, "UpdateProposalStatusActivity", approverApproveInput).Get(ctx, nil)

	currentStatus = "APPROVED"
	result.Status = "APPROVED"

	// Step 7: Generate Policy Number
	currentStatus = "GENERATING_POLICY_NUMBER"
	logger.Info("Step 7: Generating policy number")
	policyNumberInput := activities.GeneratePolicyNumberInput{
		ProposalNumber: input.ProposalNumber,
		PolicyType:     input.PolicyType,
		StateCode:      input.InsuredState,
	}
	var policyNumberResult activities.GeneratePolicyNumberResult
	if err := workflow.ExecuteActivity(shortActivityOpts, "GeneratePolicyNumberActivity", policyNumberInput).Get(ctx, &policyNumberResult); err != nil {
		logger.Error("Failed to generate policy number", "error", err)
		result.Status = "POLICY_NUMBER_GENERATION_FAILED"
		result.Error = err.Error()
		return result, nil
	}
	result.PolicyNumber = policyNumberResult.PolicyNumber

	// Step 7b: Create proposal issuance record
	currentStatus = "CREATING_ISSUANCE_RECORD"
	logger.Info("Step 7b: Creating proposal issuance record")

	issuanceInput := activities.CreatePolicyIssuanceInput{
		ProposalID:   input.ProposalID, // int64 PK
		PolicyNumber: policyNumberResult.PolicyNumber,
		ProposalDate: input.ProposalDate,
		PolicyTerm:   input.PolicyTerm,
	}

	if err := workflow.ExecuteActivity(shortActivityOpts, "CreatePolicyIssuanceActivity", issuanceInput).Get(ctx, nil); err != nil {

		logger.Error("Failed to create issuance record", "error", err)
		result.Status = "ISSUANCE_FAILED"
		result.Error = err.Error()
		return result, nil
	}
	// Update status to ISSUED
	updateStatusInput := activities.UpdateStatusInput{
		ProposalNumber: input.ProposalNumber,
		Status:         domain.ProposalStatusIssued,
		Comments:       "Policy issued: " + policyNumberResult.PolicyNumber,
		ChangedBy:      approverID,
	}
	if err := workflow.ExecuteActivity(shortActivityOpts, "UpdateProposalStatusActivity", updateStatusInput).Get(ctx, nil); err != nil {
		logger.Error("Failed to update proposal status to ISSUED", "error", err)
	}

	// Step 8: Generate Policy Bond
	currentStatus = "GENERATING_BOND"
	logger.Info("Step 8: Generating policy bond")
	bondInput := activities.GenerateBondInput{
		ProposalNumber: input.ProposalNumber,
		PolicyNumber:   policyNumberResult.PolicyNumber,
	}
	var bondResult activities.GenerateBondResult
	if err := workflow.ExecuteActivity(shortActivityOpts, "GenerateBondActivity", bondInput).Get(ctx, &bondResult); err != nil {
		logger.Error("Failed to generate bond", "error", err)
		// Don't fail the workflow, bond can be regenerated later
	} else {
		logger.Info("Bond generated", "bondID", bondResult.BondDocumentID)

		// Step 8b: Update issuance record with bond details
		updateBondInput := activities.UpdateBondDetailsInput{
			ProposalID:      input.ProposalID,
			BondDocumentID:  bondResult.BondDocumentID,
			BondGeneratedBy: approverID,
		}

		_ = workflow.ExecuteActivity(shortActivityOpts, "UpdateBondDetailsActivity", updateBondInput).Get(ctx, nil)
	}

	// Step 9: Send notification
	sendNotificationInput := activities.SendNotificationInput{
		ProposalNumber:   input.ProposalNumber,
		NotificationType: "POLICY_ISSUED",
		RecipientID:      0, // Parse from input.CustomerID
		Message:          "Your policy has been issued with number: " + policyNumberResult.PolicyNumber,
	}
	if err := workflow.ExecuteActivity(shortActivityOpts, "SendNotificationActivity", sendNotificationInput).Get(ctx, nil); err != nil {
		logger.Error("Failed to send notification", "error", err)
		// Don't fail the workflow for notification failure
	}

	// Step 10: Signal PM service to start lifecycle workflow
	logger.Info("Step 10: Signalling PM lifecycle for issued policy")
	pmSignalInput := activities.StartPMLifecycleInput{
		PolicyNumber: policyNumberResult.PolicyNumber,
		PolicyType:   string(input.PolicyType),
	}
	if err := workflow.ExecuteActivity(externalCallOpts, "StartPMLifecycleActivity", pmSignalInput).Get(ctx, nil); err != nil {
		// PM signal failure must not block the issuance result — the reconciliation
		// workflow will retry. The failure is already persisted in proposal_issuance
		// by the activity itself (pm_signal_status = 'FAILED').
		logger.Error("PM lifecycle signal failed — reconciliation will retry", "error", err)
	}

	result.Status = "ISSUED"
	logger.Info("Policy Issuance Workflow completed successfully", "proposalID", input.ProposalID, "policyNumber", result.PolicyNumber)
	return result, nil
}
