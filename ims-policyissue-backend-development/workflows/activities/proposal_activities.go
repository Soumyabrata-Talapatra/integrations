package activities

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"policy-issue-service/core/domain"
	"policy-issue-service/repo/postgres"
)

// ProposalActivities contains activities for proposal processing
type ProposalActivities struct {
	proposalRepo *postgres.ProposalRepository
	quoteRepo    *postgres.QuoteRepository
	productRepo  *postgres.ProductRepository
}

// NewProposalActivities creates a new ProposalActivities instance
func NewProposalActivities(
	proposalRepo *postgres.ProposalRepository,
	quoteRepo *postgres.QuoteRepository,
	productRepo *postgres.ProductRepository,
) *ProposalActivities {
	return &ProposalActivities{
		proposalRepo: proposalRepo,
		quoteRepo:    quoteRepo,
		productRepo:  productRepo,
	}
}

// ValidateProposalInput contains input for ValidateProposalActivity
type ValidateProposalInput struct {
	ProposalNumber string `json:"proposal_number"`
}

// ValidateProposalResult contains result of ValidateProposalActivity
type ValidateProposalResult struct {
	IsValid  bool                  `json:"is_valid"`
	Errors   []string              `json:"errors,omitempty"`
	Proposal domain.ProposalOutput `json:"proposal"`
}

// ValidateProposalActivity validates proposal data
func (a *ProposalActivities) ValidateProposalActivity(ctx context.Context, input ValidateProposalInput) (*ValidateProposalResult, error) {
	// Get proposal by number
	proposal, err := a.proposalRepo.GetProposalByNumber(ctx, input.ProposalNumber)
	if err != nil {
		return &ValidateProposalResult{
			IsValid: false,
			Errors:  []string{"Failed to retrieve proposal: " + err.Error()},
		}, nil
	}

	// Validate proposal completeness
	var errors []string

	// Check required fields
	if proposal.ProductCode == "" {
		errors = append(errors, "Product code is required")
	}
	// if proposal.CustomerID == 0 {
	// 	errors = append(errors, "Customer ID is required")
	// }
	if proposal.CustomerID == nil || *proposal.CustomerID == 0 {
		errors = append(errors, "Customer ID is required")
	}
	if proposal.SumAssured <= 0 {
		errors = append(errors, "Sum assured must be greater than 0")
	}
	if proposal.PolicyTerm <= 0 {
		errors = append(errors, "Policy term is required")
	}

	return &ValidateProposalResult{
		IsValid:  len(errors) == 0,
		Errors:   errors,
		Proposal: *proposal,
	}, nil
}

// CheckEligibilityInput contains input for CheckEligibilityActivity
type CheckEligibilityInput struct {
	ProposalNumber string `json:"proposal_number"`
	AgeAtEntry     int    `json:"age_at_entry"`
}

// CheckEligibilityResult contains result of CheckEligibilityActivity
type CheckEligibilityResult struct {
	IsEligible   bool   `json:"is_eligible"`
	RejectReason string `json:"reject_reason,omitempty"`
}

// CheckEligibilityActivity checks customer eligibility
func (a *ProposalActivities) CheckEligibilityActivity(ctx context.Context, input CheckEligibilityInput) (*CheckEligibilityResult, error) {
	// Get proposal
	proposal, err := a.proposalRepo.GetProposalByNumber(ctx, input.ProposalNumber)
	if err != nil {
		return &CheckEligibilityResult{
			IsEligible:   false,
			RejectReason: "Failed to retrieve proposal: " + err.Error(),
		}, nil
	}

	// Get product configuration
	product, err := a.quoteRepo.GetProductByCode(ctx, proposal.ProductCode)
	if err != nil {
		return &CheckEligibilityResult{
			IsEligible:   false,
			RejectReason: "Failed to retrieve product: " + err.Error(),
		}, nil
	}

	// Check age eligibility
	if !product.IsEligibleAge(input.AgeAtEntry) {
		return &CheckEligibilityResult{
			IsEligible: false,
			RejectReason: fmt.Sprintf("Age %d is not eligible for product %s (range: %d-%d)",
				input.AgeAtEntry, proposal.ProductCode, product.MinEntryAge, product.MaxEntryAge),
		}, nil
	}

	// Check sum assured eligibility
	if !product.IsEligibleSA(proposal.SumAssured) {
		maxSA := "unlimited"
		if product.MaxSumAssured != nil {
			maxSA = fmt.Sprintf("%.2f", *product.MaxSumAssured)
		}
		return &CheckEligibilityResult{
			IsEligible: false,
			RejectReason: fmt.Sprintf("Sum assured %.2f is outside product limits (%.2f-%s)",
				proposal.SumAssured, product.MinSumAssured, maxSA),
		}, nil
	}

	return &CheckEligibilityResult{
		IsEligible: true,
	}, nil
}

// CalculatePremiumInput contains input for CalculatePremiumActivity
// type CalculatePremiumInput struct {
// 	ProposalNumber  string                  `json:"proposal_number"`
// 	AgeAtEntry      int                     `json:"age_at_entry"`
// 	Gender          string                  `json:"gender"`
// 	PolicyTerm      int                     `json:"policy_term"`
// 	SumAssured      float64                 `json:"sum_assured"`
// 	Frequency       domain.PremiumFrequency `json:"frequency"`
// 	ProductCategory string                  `json:"product_category"`
// }

type CalculatePremiumInput struct {
	ProposalNumber string `json:"proposal_number"`
	// ProductCode     string                  `json:"product_code"`
	// ProductCategory string                  `json:"product_category"`
	AgeAtEntry int                     `json:"age_at_entry"`
	Gender     string                  `json:"gender"`
	PolicyTerm int                     `json:"policy_term"`
	SumAssured float64                 `json:"sum_assured"`
	Frequency  domain.PremiumFrequency `json:"frequency"`
}

// CalculatePremiumResult contains result of CalculatePremiumActivity
type CalculatePremiumResult struct {
	BasePremium  float64 `json:"base_premium"`
	Rebate       float64 `json:"rebate"`
	NetPremium   float64 `json:"net_premium"`
	GSTAmount    float64 `json:"gst_amount"`
	TotalPayable float64 `json:"total_payable"`
}

// CalculatePremiumActivity calculates premium
func (a *ProposalActivities) CalculatePremiumActivity(ctx context.Context, input CalculatePremiumInput) (*CalculatePremiumResult, error) {
	// Get product configuration
	proposal, err := a.proposalRepo.GetProposalByNumber(ctx, input.ProposalNumber)
	if err != nil {
		return nil, err
	}
	if input.Frequency == "" {
		return nil, fmt.Errorf("premium frequency is required")
	}

	product, err := a.productRepo.GetProductByCode(ctx, proposal.ProductCode)
	if err != nil {
		return nil, fmt.Errorf("invalid product code")
	}
	err = domain.ValidateProductAge(proposal.ProductCode, input.AgeAtEntry, input.PolicyTerm)
	if err != nil {
		return nil, err
	}
	productCategory := string(product.ProductCategory)
	var lookupField string
	var lookupValue int

	switch proposal.ProductCode {

	// TERM PRODUCTS
	case "1002", "1005", "5002", "5003":
		lookupField = "term"
		lookupValue = input.PolicyTerm

	// PREMIUM CEASING AGE PRODUCTS
	case "1001", "1003", "1004", "1006",
		"5001", "5004", "5005", "5006":

		if proposal.PremiumCeasingAge == nil {
			return nil, fmt.Errorf(
				"premium_ceasing_age not set for proposal: %s",
				proposal.ProposalNumber,
			)
		}

		lookupField = "premium_ceasing_age"
		lookupValue = *proposal.PremiumCeasingAge

	default:
		return nil, fmt.Errorf("invalid product code: %s", proposal.ProductCode)
	}

	var sumAssdFloat float64
	// Get premium rate from Sankalan table
	// rate, sumAssd, err := a.quoteRepo.GetPremiumRate(ctx, proposal.ProductCode, input.ProductCategory, input.AgeAtEntry, (input.Gender), string(input.Frequency), "term", input.PolicyTerm)
	// if err != nil {
	// 	return nil, err
	// }
	rate, sumAssd, err := a.quoteRepo.GetPremiumRate(ctx, proposal.ProductCode, productCategory,
		input.AgeAtEntry, input.Gender, string(input.Frequency), lookupField, lookupValue)
	if err != nil {
		return nil, err
	}
	// // Calculate base premium: (SA / 1000) * Rate
	// basePremium := (input.SumAssured / 1000) * rate
	// Convert sum_assd string to float64
	sumAssdFloat, err = strconv.ParseFloat(strconv.Itoa(sumAssd), 64)
	if err != nil {
		return nil, fmt.Errorf("invalid sum_assd value: %v", err)
	}

	basePremium := (input.SumAssured / sumAssdFloat) * rate
	// Calculate rebate based on frequency
	// var rebate float64
	// switch input.Frequency {
	// case domain.FrequencyHalfYearly:
	// 	rebate = basePremium * 0.015
	// case domain.FrequencyYearly:
	// 	rebate = basePremium * 0.03
	// }
	// netPremium := basePremium - rebate
	rebateAmount, err := a.quoteRepo.GetRebate(ctx, proposal.ProductCode, int(input.SumAssured))

	if err != nil {
		return nil, err
	}

	// Safety: rebate should not exceed base premium
	if rebateAmount > basePremium {
		rebateAmount = basePremium
	}

	netPremium := basePremium - rebateAmount

	// Calculate GST (18%)
	gstAmount := netPremium * 0.0
	totalPayable := netPremium + gstAmount

	return &CalculatePremiumResult{
		BasePremium:  basePremium,
		Rebate:       rebateAmount,
		NetPremium:   netPremium,
		GSTAmount:    gstAmount,
		TotalPayable: totalPayable,
	}, nil
}

// SavePremiumInput contains input for SavePremiumToProposalActivity
type SavePremiumInput struct {
	ProposalNumber string  `json:"proposal_number"`
	BasePremium    float64 `json:"base_premium"`
	Rebate         float64 `json:"rebate"`
	NetPremium     float64 `json:"net_premium"`
	GSTAmount      float64 `json:"gst_amount"`
	TotalPayable   float64 `json:"total_payable"`
}

// SavePremiumToProposalActivity saves calculated premium
func (a *ProposalActivities) SavePremiumToProposalActivity(ctx context.Context, input SavePremiumInput) error {
	// Get proposal ID
	proposal, err := a.proposalRepo.GetProposalByNumber(ctx, input.ProposalNumber)
	if err != nil {
		return err
	}

	// Update proposal with premium values using correct column name
	fields := map[string]interface{}{
		"annual_premium_equivalent": input.TotalPayable,
		"base_premium":              input.BasePremium,
		"gst_amount":                input.GSTAmount,
		"total_premium":             input.TotalPayable,
		"updated_at":                time.Now(),
	}

	return a.proposalRepo.UpdateProposalFields(ctx, proposal.ProposalID, fields)
}

// UpdateStatusInput contains input for UpdateProposalStatusActivity
type UpdateStatusInput struct {
	ProposalNumber string                `json:"proposal_number"`
	Status         domain.ProposalStatus `json:"status"`
	Comments       string                `json:"comments,omitempty"`
	ChangedBy      int64                 `json:"changed_by"`
}

// UpdateProposalStatusActivity updates proposal status
func (a *ProposalActivities) UpdateProposalStatusActivity(ctx context.Context, input UpdateStatusInput) error {
	proposal, err := a.proposalRepo.GetProposalByNumber(ctx, input.ProposalNumber)
	if err != nil {
		return err
	}

	// Use ChangedBy if provided, otherwise use proposal creator
	// changedBy := input.ChangedBy
	// if changedBy == 0 {
	// 	changedBy = proposal.CreatedBy
	// }
	if input.ChangedBy == 0 {
		return fmt.Errorf("changed_by is mandatory")
	}
	return a.proposalRepo.UpdateProposalStatus(ctx, proposal.ProposalID, input.Status, input.Comments, input.ChangedBy)
}

// SendNotificationInput contains input for SendNotificationActivity
type SendNotificationInput struct {
	ProposalNumber   string `json:"proposal_number"`
	NotificationType string `json:"notification_type"`
	RecipientID      int64  `json:"recipient_id"`
	Message          string `json:"message"`
}

// SendNotificationActivity sends notifications
func (a *ProposalActivities) SendNotificationActivity(ctx context.Context, input SendNotificationInput) error {
	// TODO: Implement notification sending via notification service
	// This is a placeholder that would integrate with the notification service
	return nil
}

// MedicalReviewInput contains input for RequestMedicalReviewActivity
type MedicalReviewInput struct {
	ProposalNumber string  `json:"proposal_number"`
	CustomerID     int64   `json:"customer_id"`
	AgeAtEntry     int     `json:"age_at_entry"`
	SumAssured     float64 `json:"sum_assured"`
}

// MedicalReviewResult contains result of RequestMedicalReviewActivity
type MedicalReviewResult struct {
	MedicalRequired bool   `json:"medical_required"`
	ExaminationType string `json:"examination_type,omitempty"`
	Instructions    string `json:"instructions,omitempty"`
}

// RequestMedicalReviewActivity requests medical examination
func (a *ProposalActivities) RequestMedicalReviewActivity(ctx context.Context, input MedicalReviewInput) (*MedicalReviewResult, error) {
	// Determine if medical is required based on age and sum assured
	// [BR-POL-028] Medical Requirements
	medicalRequired := false
	var examinationType string

	if input.AgeAtEntry >= 50 {
		medicalRequired = true
		examinationType = "FULL_MEDICAL"
	} else if input.SumAssured > 500000 {
		medicalRequired = true
		examinationType = "PARAMEDICAL"
	}

	if medicalRequired {
		// Update proposal status to PENDING_MEDICAL
		proposal, err := a.proposalRepo.GetProposalByNumber(ctx, input.ProposalNumber)
		if err != nil {
			return nil, err
		}

		if err := a.proposalRepo.UpdateProposalStatus(ctx, proposal.ProposalID,
			domain.ProposalStatusPendingMedical, "Medical examination required", proposal.CreatedBy); err != nil {
			return nil, err
		}
	}

	return &MedicalReviewResult{
		MedicalRequired: medicalRequired,
		ExaminationType: examinationType,
		Instructions:    "Please visit nearest post office for medical examination",
	}, nil
}

// RouteToApproverInput contains input for RouteToApproverActivity
type RouteToApproverInput struct {
	ProposalNumber string  `json:"proposal_number"`
	SumAssured     float64 `json:"sum_assured"`
}

// RouteToApproverResult contains the routing decision
type RouteToApproverResult struct {
	ApproverLevel int    `json:"approver_level"`
	ApproverRole  string `json:"approver_role"`
}

// RouteToApproverActivity routes proposal to the correct approver level based on SA bands
// [BR-POL-016] Approval Routing by Sum Assured
// Queries approval_routing_config to determine the correct approver level:
//   - SA ≤ ₹5,00,000         → Level 1 (APPROVER_LEVEL_1)
//   - ₹5,00,001 - ₹10,00,000 → Level 2 (APPROVER_LEVEL_2)
//   - SA > ₹10,00,000        → Level 3 (APPROVER_LEVEL_3)
func (a *ProposalActivities) RouteToApproverActivity(ctx context.Context, input RouteToApproverInput) (*RouteToApproverResult, error) {
	proposal, err := a.proposalRepo.GetProposalByNumber(ctx, input.ProposalNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to get proposal %s: %w", input.ProposalNumber, err)
	}

	// Use the SA from input (passed from workflow) to query routing config
	sa := input.SumAssured
	if sa <= 0 {
		sa = proposal.SumAssured
	}

	routingConfig, err := a.proposalRepo.GetApprovalRoutingConfig(ctx, sa)
	if err != nil {
		return nil, fmt.Errorf("failed to get approval routing for SA %.2f: %w", sa, err)
	}

	comment := fmt.Sprintf("Routed to approver: Level %d (%s) for SA %.2f",
		routingConfig.ApproverLevel, routingConfig.ApproverRole, sa)

	if err := a.proposalRepo.UpdateProposalStatus(ctx, proposal.ProposalID,
		domain.ProposalStatusApprovalPending, comment, proposal.CreatedBy); err != nil {
		return nil, err
	}

	return &RouteToApproverResult{
		ApproverLevel: routingConfig.ApproverLevel,
		ApproverRole:  routingConfig.ApproverRole,
	}, nil
}

// GeneratePolicyNumberInput contains input for GeneratePolicyNumberActivity
type GeneratePolicyNumberInput struct {
	ProposalNumber string            `json:"proposal_number"`
	PolicyType     domain.PolicyType `json:"policy_type"`
	StateCode      string            `json:"state_code"`
}

// GeneratePolicyNumberResult contains result of GeneratePolicyNumberActivity
type GeneratePolicyNumberResult struct {
	PolicyNumber string `json:"policy_number"`
}

// GeneratePolicyNumberActivity generates policy number
func (a *ProposalActivities) GeneratePolicyNumberActivity(ctx context.Context, input GeneratePolicyNumberInput) (*GeneratePolicyNumberResult, error) {
	policyNumber, err := a.proposalRepo.GeneratePolicyNumber(ctx, input.PolicyType, input.StateCode)
	if err != nil {
		return nil, err
	}

	return &GeneratePolicyNumberResult{
		PolicyNumber: policyNumber,
	}, nil
}

// GenerateBondInput contains input for GenerateBondActivity
type GenerateBondInput struct {
	ProposalNumber string `json:"proposal_number"`
	PolicyNumber   string `json:"policy_number"`
}

// GenerateBondResult contains result of GenerateBondActivity
type GenerateBondResult struct {
	BondDocumentID string `json:"bond_document_id"`
	BondURL        string `json:"bond_url"`
}

// GenerateBondActivity generates policy bond
func (a *ProposalActivities) GenerateBondActivity(ctx context.Context, input GenerateBondInput) (*GenerateBondResult, error) {
	// TODO: Implement bond generation via document service
	// This is a placeholder that would integrate with the document service

	// Generate bond document ID
	bondDocumentID := fmt.Sprintf("BOND-%s-%d", input.PolicyNumber, time.Now().Unix())

	return &GenerateBondResult{
		BondDocumentID: bondDocumentID,
		BondURL:        fmt.Sprintf("/documents/%s", bondDocumentID),
	}, nil
}

type CreatePolicyIssuanceInput struct {
	ProposalID   int64     `json:"proposal_id"`
	PolicyNumber string    `json:"policy_number"`
	ProposalDate time.Time `json:"proposal_date"`
	PolicyTerm   int       `json:"policy_term"`
}

func (a *ProposalActivities) CreatePolicyIssuanceActivity(ctx context.Context,
	input CreatePolicyIssuanceInput,
) error {

	issueDate := time.Now().UTC()
	commencementDate := input.ProposalDate
	maturityDate := commencementDate.AddDate(input.PolicyTerm, 0, 0)

	err := a.proposalRepo.InsertProposalIssuance(
		ctx,
		input.ProposalID,
		input.PolicyNumber,
		issueDate,
		commencementDate,
		maturityDate,
	)
	if err != nil {
		return err
	}
	// Update policy number in proposals and nominees
	err = a.proposalRepo.UpdatePolicyNumber(ctx,input.ProposalID,input.PolicyNumber)
	if err != nil {
		return err
	}
	return nil
}

type UpdateBondDetailsInput struct {
	ProposalID      int64  `json:"proposal_id"`
	BondDocumentID  string `json:"bond_document_id"`
	BondGeneratedBy int64  `json:"bond_generated_by"`
}

func (a *ProposalActivities) UpdateBondDetailsActivity(
	ctx context.Context,
	input UpdateBondDetailsInput,
) error {

	if input.BondDocumentID == "" {
		return fmt.Errorf("bond_document_id is required")
	}

	if input.BondGeneratedBy == 0 {
		return fmt.Errorf("bond_generated_by is required")
	}

	return a.proposalRepo.UpdateBondDetails(
		ctx,
		input.ProposalID,
		input.BondDocumentID,
		input.BondGeneratedBy,
	)
}
