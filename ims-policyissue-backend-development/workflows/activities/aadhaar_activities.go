package activities

import (
	"context"
	"fmt"
	"policy-issue-service/core/domain"
	"policy-issue-service/repo/postgres"
	"strconv"
	"strings"
	"time"

	"go.temporal.io/sdk/activity"
)

// AadhaarActivities contains all activities related to Aadhaar flow and instant issuance
type AadhaarActivities struct {
	proposalRepo *postgres.ProposalRepository
	productRepo  *postgres.ProductRepository
	quoteRepo    *postgres.QuoteRepository
}

// NewAadhaarActivities creates a new AadhaarActivities instance
func NewAadhaarActivities(proposalRepo *postgres.ProposalRepository, productRepo *postgres.ProductRepository, quoteRepo *postgres.QuoteRepository) *AadhaarActivities {
	return &AadhaarActivities{
		proposalRepo: proposalRepo,
		productRepo:  productRepo,
		quoteRepo:    quoteRepo,
	}
}

// PremiumCalculationInput represents input for premium calculation
type PremiumCalculationInput struct {
	ProductCode string                  `json:"product_code"`
	SumAssured  float64                 `json:"sum_assured"`
	PolicyTerm  int                     `json:"policy_term"`
	Frequency   domain.PremiumFrequency `json:"frequency"`
	CustomerDOB string                  `json:"customer_dob"`
	Gender      string                  `json:"gender"`

	ProductCategory string `json:"product_category"`
	AgeAtEntry      int    `json:"age_at_entry"`
	AgeProofType    string `json:"age_proof_type"`
	InsuredState    string `json:"insured_state"`
	ProviderState   string `json:"provider_state"`
}

// PremiumCalculationResult represents result of premium calculation
type PremiumCalculationResult struct {
	BasePremium             float64                `json:"base_premium"`
	ServiceTax              float64                `json:"service_tax"`
	GST                     float64                `json:"gst"`
	RebateAmount            float64                `json:"rebate_amount"`
	AnnualPremium           float64                `json:"annual_premium"`
	AnnualPremiumEquivalent float64                `json:"annual_premium_equivalent"`
	TotalPremium            float64                `json:"total_premium"`
	CalculationDetails      map[string]interface{} `json:"calculation_details"`
}

// ValidateAndCalculatePremiumActivity validates eligibility and calculates premium for Aadhaar proposals
func (a *AadhaarActivities) ValidateAndCalculatePremiumActivity(ctx context.Context, input PremiumCalculationInput) (PremiumCalculationResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting premium calculation for Aadhaar flow",
		"productCode", input.ProductCode,
		"sumAssured", input.SumAssured)

	// Calculate age from DOB
	age := calculateAgeFromDOB(input.CustomerDOB)

	var sumAssdFloat float64
	// Get premium rate from Sankalan table
	rate, sumAssd, err := a.quoteRepo.GetPremiumRate(
		ctx,
		input.ProductCode,
		input.ProductCategory,
		input.AgeAtEntry,
		input.Gender,
		string(input.Frequency),
		"term",           // lookupField
		input.PolicyTerm, // lookupValue
	)
	if err != nil {
		logger.Error("Failed to get premium rate", "error", err)
		return PremiumCalculationResult{}, fmt.Errorf("failed to get premium rate: %w", err)
	}

	// // Calculate base premium: (SA / 1000) * Rate
	// basePremium := (input.SumAssured / 1000) * rate
	sumAssdFloat, err = strconv.ParseFloat(strconv.Itoa(sumAssd), 64)
	if err != nil {
		return PremiumCalculationResult{}, fmt.Errorf("invalid sum_assd value: %w", err)
	}

	basePremium := (input.SumAssured / sumAssdFloat) * rate
	// Calculate rebate based on frequency
	var rebate float64
	switch input.Frequency {
	case domain.FrequencyHalfYearly:
		rebate = basePremium * 0.015
	case domain.FrequencyYearly:
		rebate = basePremium * 0.03
	}

	netPremium := basePremium - rebate

	// Calculate GST (18%)
	gstAmount := netPremium * 0.0
	totalPayable := netPremium + gstAmount

	calculationResult := PremiumCalculationResult{
		BasePremium:             basePremium,
		RebateAmount:            rebate,
		GST:                     gstAmount,
		AnnualPremium:           netPremium, // Simplified
		AnnualPremiumEquivalent: netPremium, // Simplified
		TotalPremium:            totalPayable,
		CalculationDetails: map[string]interface{}{
			"method":       "sankalan_table",
			"product_code": input.ProductCode,
			"sum_assured":  input.SumAssured,
			"age":          age,
			"rate":         rate,
		},
	}

	logger.Info("Premium calculation completed", "totalPremium", calculationResult.TotalPremium)
	return calculationResult, nil
}

// CreateAadhaarProposalInput represents input for creating Aadhaar proposal
type CreateAadhaarProposalInput struct {
	ProposalNumber  string                  `json:"proposal_number"`
	CustomerID      string                  `json:"customer_id"`
	ProductCode     string                  `json:"product_code"`
	PolicyType      domain.PolicyType       `json:"policy_type"`
	SumAssured      float64                 `json:"sum_assured"`
	PolicyTerm      int                     `json:"policy_term"`
	Frequency       domain.PremiumFrequency `json:"frequency"`
	PremiumAmount   float64                 `json:"premium_amount"`
	Channel         domain.Channel          `json:"channel"`
	EntryPath       domain.EntryPath        `json:"entry_path"`
	AadhaarNumber   string                  `json:"aadhaar_number"`
	CustomerName    string                  `json:"customer_name"`
	CustomerDOB     string                  `json:"customer_dob"`
	Gender          string                  `json:"gender"`
	Address         string                  `json:"address"`
	MobileNumber    string                  `json:"mobile_number"`
	AadhaarVerified bool                    `json:"aadhaar_verified"`
}

// CreateAadhaarProposalResult represents result of creating Aadhaar proposal
type CreateAadhaarProposalResult struct {
	ProposalID     string `json:"proposal_id"`
	ProposalNumber string `json:"proposal_number"`
	Status         string `json:"status"`
	Message        string `json:"message"`
}

// CreateAadhaarProposalActivity creates a proposal using Aadhaar-verified data
func (a *AadhaarActivities) CreateAadhaarProposalActivity(ctx context.Context, input CreateAadhaarProposalInput) (CreateAadhaarProposalResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Creating Aadhaar-based proposal",
		"customerID", input.CustomerID,
		"productCode", input.ProductCode,
		"sumAssured", input.SumAssured)

	// Validate that Aadhaar is verified
	if !input.AadhaarVerified {
		return CreateAadhaarProposalResult{}, fmt.Errorf("Aadhaar must be verified to create Aadhaar-based proposal")
	}

	custID, _ := strconv.ParseInt(input.CustomerID, 10, 64)

	proposal := &domain.Proposal{
		CustomerID:              &custID,
		ProductCode:             input.ProductCode,
		PolicyType:              input.PolicyType,
		Channel:                 input.Channel,
		Status:                  domain.ProposalStatusDataEntry,
		EntryPath:               input.EntryPath,
		SumAssured:              input.SumAssured,
		PolicyTerm:              input.PolicyTerm,
		PremiumPaymentFrequency: input.Frequency,
		CreatedBy:               0, // System created
	}

	// Basic name splitting for mock
	nameParts := strings.Fields(input.CustomerName)
	firstName := ""
	lastName := ""
	if len(nameParts) > 0 {
		firstName = nameParts[0]
	}
	if len(nameParts) > 1 {
		lastName = nameParts[len(nameParts)-1]
	}

	insured := &domain.ProposalInsured{
		FirstName:    firstName,
		LastName:     lastName,
		Gender:       input.Gender,
		DateOfBirth:  input.CustomerDOB,
		AddressLine1: &input.Address,
		Mobile:       &input.MobileNumber,
	}

	if err := a.proposalRepo.CreateProposalWithAadhaar(ctx, proposal, insured); err != nil {
		logger.Error("Failed to create Aadhaar proposal", "error", err)
		return CreateAadhaarProposalResult{}, fmt.Errorf("failed to create proposal: %w", err)
	}

	return CreateAadhaarProposalResult{
		ProposalID:     fmt.Sprintf("%d", proposal.ProposalID),
		ProposalNumber: proposal.ProposalNumber,
		Status:         "SUCCESS",
		Message:        "Aadhaar-based proposal created successfully",
	}, nil
}

// AadhaarEligibilityInput represents input for Aadhaar eligibility check
type AadhaarEligibilityInput struct {
	ProposalNumber string  `json:"proposal_number"`
	ProductCode    string  `json:"product_code"`
	SumAssured     float64 `json:"sum_assured"`
	CustomerDOB    string  `json:"customer_dob"`
	Gender         string  `json:"gender"`
	// Additional parameters as needed
}

// AadhaarEligibilityResult represents result of Aadhaar eligibility check
type AadhaarEligibilityResult struct {
	Eligible bool                   `json:"eligible"`
	Reason   string                 `json:"reason"`
	Details  map[string]interface{} `json:"details"`
}

// CheckInstantIssuanceEligibilityActivity checks if the proposal is eligible for instant issuance
func (a *AadhaarActivities) CheckInstantIssuanceEligibilityActivity(ctx context.Context, input AadhaarEligibilityInput) (AadhaarEligibilityResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Checking instant issuance eligibility",
		"proposalID", input.ProposalNumber,
		"productCode", input.ProductCode,
		"sumAssured", input.SumAssured)

	// Calculate age from DOB
	// In a real implementation, this would parse the DOB string and calculate the age
	// For now, we'll simulate the age check
	age := calculateAgeFromDOB(input.CustomerDOB)

	// Get product configuration to determine non-medical limit
	productConfig, err := a.getProductConfig(ctx, input.ProductCode)
	if err != nil {
		logger.Error("Failed to get product configuration", "error", err)
		return AadhaarEligibilityResult{
			Eligible: false,
			Reason:   fmt.Sprintf("Could not retrieve product configuration: %v", err),
		}, nil
	}

	// Check eligibility criteria based on requirements:
	// 1. Age <= 50 years
	// 2. Sum Assured <= NonMedicalLimit
	// 3. Product supports instant issuance
	// 4. No medical requirements for the specific case

	isEligible := true
	reason := ""

	if age > 50 {
		isEligible = false
		reason = fmt.Sprintf("Age %d exceeds maximum age limit of 50 for instant issuance", age)
	} else if input.SumAssured > productConfig.NonMedicalLimit {
		isEligible = false
		reason = fmt.Sprintf("Sum assured %.2f exceeds non-medical limit %.2f", input.SumAssured, productConfig.NonMedicalLimit)
	} else if !productConfig.SupportsInstantIssuance {
		isEligible = false
		reason = fmt.Sprintf("Product %s does not support instant issuance", input.ProductCode)
	}

	result := AadhaarEligibilityResult{
		Eligible: isEligible,
		Reason:   reason,
		Details: map[string]interface{}{
			"age":             age,
			"ageLimit":        50,
			"sumAssured":      input.SumAssured,
			"nonMedicalLimit": productConfig.NonMedicalLimit,
			"supportsInstant": productConfig.SupportsInstantIssuance,
			"productCode":     input.ProductCode,
		},
	}

	if isEligible {
		logger.Info("Proposal is eligible for instant issuance", "proposalID", input.ProposalNumber)
	} else {
		logger.Info("Proposal is NOT eligible for instant issuance", "proposalID", input.ProposalNumber, "reason", reason)
	}

	return result, nil
}

// SendPolicyBondElectronicInput represents input for SendPolicyBondElectronicActivity
type SendPolicyBondElectronicInput struct {
	CustomerID    string `json:"customer_id"`
	BondDocID     string `json:"bond_doc_id"`
	PolicyNumber  string `json:"policy_number"`
	CustomerName  string `json:"customer_name"`
	MobileNumber  string `json:"mobile_number"`
	PreferredMode string `json:"preferred_mode"`
}

// SendPolicyBondElectronicActivity sends the policy bond via electronic channels
func (a *AadhaarActivities) SendPolicyBondElectronicActivity(ctx context.Context, input SendPolicyBondElectronicInput) error {
	logger := activity.GetLogger(ctx)
	logger.Info("Sending electronic policy bond",
		"customerID", input.CustomerID,
		"policyNumber", input.PolicyNumber,
		"mode", input.PreferredMode)

	// Mock implementation
	return nil
}

// ProductConfig holds product-specific configuration
type ProductConfig struct {
	ProductCode             string
	NonMedicalLimit         float64
	SupportsInstantIssuance bool
	MinAge                  int
	MaxAge                  int
}

// getProductConfig retrieves product configuration from the repository
func (a *AadhaarActivities) getProductConfig(ctx context.Context, productCode string) (ProductConfig, error) {
	product, err := a.productRepo.GetProductByCode(ctx, productCode)
	if err != nil {
		return ProductConfig{}, err
	}

	nonMedicalLimit := 0.0
	if product.MedicalSAThreshold != nil {
		nonMedicalLimit = *product.MedicalSAThreshold
	}

	config := ProductConfig{
		ProductCode:             product.ProductCode,
		NonMedicalLimit:         nonMedicalLimit, // Use real threshold
		SupportsInstantIssuance: true,            // Still defaulting as no column exists
		MinAge:                  product.MinEntryAge,
		MaxAge:                  product.MaxEntryAge,
	}
	return config, nil
}

// calculateAgeFromDOB calculates age from date of birth string (YYYY-MM-DD)
func calculateAgeFromDOB(dob string) int {
	birthDate, err := time.Parse("2006-01-02", dob)
	if err != nil {
		return 0
	}
	now := time.Now()
	age := now.Year() - birthDate.Year()
	if now.YearDay() < birthDate.YearDay() {
		age--
	}
	return age
}
