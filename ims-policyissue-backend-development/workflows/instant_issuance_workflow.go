package workflows

import (
	"policy-issue-service/core/domain"
	"policy-issue-service/workflows/activities"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// InstantIssuanceInput contains input data for the instant issuance workflow
type InstantIssuanceInput struct {
	ProposalNumber     string                  `json:"proposal_number"`
	CustomerID         string                  `json:"customer_id"`
	ProductCode        string                  `json:"product_code"`
	PolicyType         domain.PolicyType       `json:"policy_type"`
	SumAssured         float64                 `json:"sum_assured"`
	PolicyTerm         int                     `json:"policy_term"`
	Frequency          domain.PremiumFrequency `json:"frequency"`
	Channel            domain.Channel          `json:"channel"`
	AadhaarVerified    bool                    `json:"aadhaar_verified"`
	PremiumPaid        bool                    `json:"premium_paid"`
	PaymentRef         string                  `json:"payment_ref"`
	CustomerName       string                  `json:"customer_name"`
	CustomerDOB        string                  `json:"customer_dob"`
	Gender             string                  `json:"gender"`
	Address            string                  `json:"address"`
	MobileNumber       string                  `json:"mobile_number"`
	Email              string                  `json:"email"`
	StateCode          string                  `json:"state_code"`
	EligibleForInstant bool                    `json:"eligible_for_instant"`
}

// InstantIssuanceWorkflow executes the instant issuance flow for Aadhaar-verified proposals
func InstantIssuanceWorkflow(ctx workflow.Context, input InstantIssuanceInput) (*PolicyIssuanceResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting Instant Issuance Workflow", "proposalID", input.ProposalNumber)

	result := &PolicyIssuanceResult{
		ProposalNumber: input.ProposalNumber,
		Status:         "PENDING",
	}

	actCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 60 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    30 * time.Second,
			MaximumAttempts:    3,
		},
	})

	// Step 1: Validate and calculate premium
	var premiumResult activities.PremiumCalculationResult
	err := workflow.ExecuteActivity(actCtx, "ValidateAndCalculatePremiumActivity", activities.PremiumCalculationInput{
		ProductCode: input.ProductCode,
		SumAssured:  input.SumAssured,
		PolicyTerm:  input.PolicyTerm,
		Frequency:   input.Frequency,
		CustomerDOB: input.CustomerDOB,
		Gender:      input.Gender,
	}).Get(ctx, &premiumResult)
	if err != nil {
		return failWorkflow(result, "PREMIUM_ERROR", err.Error()), nil
	}

	// Step 2: Create Aadhaar proposal
	var proposalResult activities.CreateAadhaarProposalResult
	err = workflow.ExecuteActivity(actCtx, "CreateAadhaarProposalActivity", activities.CreateAadhaarProposalInput{
		ProposalNumber:  input.ProposalNumber,
		CustomerID:      input.CustomerID,
		ProductCode:     input.ProductCode,
		PolicyType:      input.PolicyType,
		SumAssured:      input.SumAssured,
		PolicyTerm:      input.PolicyTerm,
		PremiumAmount:   premiumResult.TotalPremium,
		Channel:         input.Channel,
		EntryPath:       domain.EntryPathWithAadhaar,
		CustomerName:    input.CustomerName,
		CustomerDOB:     input.CustomerDOB,
		Gender:          input.Gender,
		Address:         input.Address,
		MobileNumber:    input.MobileNumber,
		AadhaarVerified: input.AadhaarVerified,
	}).Get(ctx, &proposalResult)
	if err != nil {
		return failWorkflow(result, "PROPOSAL_CREATION_ERROR", err.Error()), nil
	}

	// Step 3: Check eligibility for instant issuance
	var eligibilityResult activities.AadhaarEligibilityResult
	err = workflow.ExecuteActivity(actCtx, "CheckInstantIssuanceEligibilityActivity", activities.AadhaarEligibilityInput{
		ProposalNumber: input.ProposalNumber,
		ProductCode:    input.ProductCode,
		SumAssured:     input.SumAssured,
		CustomerDOB:    input.CustomerDOB,
		Gender:         input.Gender,
	}).Get(ctx, &eligibilityResult)
	if err != nil {
		return failWorkflow(result, "ELIGIBILITY_ERROR", err.Error()), nil
	}

	if !eligibilityResult.Eligible {
		return failWorkflow(result, "NOT_ELIGIBLE", eligibilityResult.Reason), nil
	}

	// Step 4: Generate policy number
	var policyNumResult activities.GeneratePolicyNumberResult
	err = workflow.ExecuteActivity(actCtx, "GeneratePolicyNumberActivity", activities.GeneratePolicyNumberInput{
		ProposalNumber: proposalResult.ProposalNumber,
		PolicyType:     input.PolicyType,
		StateCode:      input.StateCode,
	}).Get(ctx, &policyNumResult)
	if err != nil {
		return failWorkflow(result, "POLICY_NUMBER_ERROR", err.Error()), nil
	}

	// Step 5: Generate policy bond
	var bondResult activities.GenerateBondResult
	err = workflow.ExecuteActivity(actCtx, "GenerateBondActivity", activities.GenerateBondInput{
		ProposalNumber: proposalResult.ProposalNumber,
		PolicyNumber:   policyNumResult.PolicyNumber,
	}).Get(ctx, &bondResult)
	if err != nil {
		return failWorkflow(result, "BOND_ERROR", err.Error()), nil
	}

	// Step 6: Send bond via electronic channel
	err = workflow.ExecuteActivity(actCtx, "SendPolicyBondElectronicActivity", activities.SendPolicyBondElectronicInput{
		CustomerID:    input.CustomerID,
		BondDocID:     bondResult.BondDocumentID,
		PolicyNumber:  policyNumResult.PolicyNumber,
		CustomerName:  input.CustomerName,
		MobileNumber:  input.MobileNumber,
		PreferredMode: "WHATSAPP",
	}).Get(ctx, nil)
	if err != nil {
		logger.Warn("Failed to send electronic bond", "error", err)
	}

	// Step 7: Update proposal status to ACTIVE
	err = workflow.ExecuteActivity(actCtx, "UpdateProposalStatusActivity", activities.UpdateStatusInput{
		ProposalNumber: proposalResult.ProposalNumber,
		Status:         domain.ProposalStatusActive,
		ChangedBy:      0,
		Comments:       "Instant issuance via Aadhaar flow",
	}).Get(ctx, nil)
	if err != nil {
		return failWorkflow(result, "STATUS_UPDATE_ERROR", err.Error()), nil
	}

	result.Status = "ACTIVE"
	result.ProposalNumber = proposalResult.ProposalID
	return result, nil
}

// Helper function to create a failed workflow result
func failWorkflow(result *PolicyIssuanceResult, errorCode, errorMessage string) *PolicyIssuanceResult {
	result.Status = "FAILED"
	result.Error = errorCode + ": " + errorMessage
	return result
}
