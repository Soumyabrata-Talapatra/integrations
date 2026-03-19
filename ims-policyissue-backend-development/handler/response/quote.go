package response

import (
	"time"

	"policy-issue-service/core/domain"
	"policy-issue-service/core/port"
)

// ProductResponse represents a product in the catalog
type ProductResponse struct {
	ProductCode              string   `json:"product_code"`
	ProductName              string   `json:"product_name"`
	ProductType              string   `json:"product_type"`
	ProductCategory          string   `json:"product_category"`
	MinSumAssured            float64  `json:"min_sum_assured"`
	MaxSumAssured            *float64 `json:"max_sum_assured,omitempty"`
	MinEntryAge              int      `json:"min_entry_age"`
	MaxEntryAge              int      `json:"max_entry_age"`
	MaxMaturityAge           *int     `json:"max_maturity_age,omitempty"`
	MinTerm                  int      `json:"min_term"`
	PremiumCeasingAgeOptions []int    `json:"premium_ceasing_age_options,omitempty"`
	AvailableFrequencies     []string `json:"available_frequencies"`
	Description              *string  `json:"description,omitempty"`
}

// ProductListResponse represents the response for GetProducts API
// [POL-API-001]
type ProductListResponse struct {
	port.StatusCodeAndMessage
	Products []ProductResponse `json:"products"`
}

// EligibilityResponse represents eligibility check result
// [VR-PI-012] [VR-PI-013] [VR-PI-044]
type EligibilityResponse struct {
	IsEligible   bool   `json:"is_eligible"`
	AgeAtEntry   int    `json:"age_at_entry"`
	MaturityAge  int    `json:"maturity_age"`
	RejectReason string `json:"reject_reason,omitempty"`
}

// BenefitIllustrationResponse represents projected benefits
type BenefitIllustrationResponse struct {
	MaturityValueGuaranteed float64 `json:"maturity_value_guaranteed"`
	MaturityValueWithBonus  float64 `json:"maturity_value_with_bonus"`
	IndicativeBonusRate     float64 `json:"indicative_bonus_rate"`
	DeathBenefit            float64 `json:"death_benefit"`
}

// CalculationBasisResponse represents the basis for calculation
type CalculationBasisResponse struct {
	PremiumTable string  `json:"premium_table"`
	SumAssd      float64 `json:"sum_assd"`
	// PremiumPerSumAssd float64 `json:"premium_per_sum_assd"`
	GSTRate float64 `json:"gst_rate"`
}

// QuoteCalculateResponse represents the response for CalculateQuote API
// [POL-API-002]
type QuoteCalculationItem struct {
	SumAssured          int64                       `json:"sum_assured"`
	Eligibility         EligibilityResponse         `json:"eligibility"`
	PremiumBreakdown    PremiumBreakdownResponse    `json:"premium_breakdown"`
	BenefitIllustration BenefitIllustrationResponse `json:"benefit_illustration"`
}

//	type QuoteCalculateResponse struct {
//		port.StatusCodeAndMessage
//		CalculationID       string                      `json:"calculation_id"`
//		Eligibility         EligibilityResponse         `json:"eligibility"`
//		PremiumBreakdown    PremiumBreakdownResponse    `json:"premium_breakdown"`
//		BenefitIllustration BenefitIllustrationResponse `json:"benefit_illustration"`
//		CalculationBasis    CalculationBasisResponse    `json:"calculation_basis"`
//		WorkflowState       port.WorkflowStateResponse `json:"workflow_state,omitempty"`
//	}
type QuoteCalculateResponse struct {
	port.StatusCodeAndMessage
	CalculationID    string                     `json:"calculation_id"`
	CalculationBasis CalculationBasisResponse   `json:"calculation_basis"`
	Calculations     []QuoteCalculationItem     `json:"calculations"`
	WorkflowState    port.WorkflowStateResponse `json:"workflow_state,omitempty"`
}

// QuotePDFInfo represents PDF document information
type QuotePDFInfo struct {
	DocumentID  string `json:"document_id"`
	DownloadURL string `json:"download_url"`
}

// QuoteValidityInfo represents quote validity information
type QuoteValidityInfo struct {
	ExpiresAt time.Time `json:"expires_at"`
	DaysValid int       `json:"days_valid"`
}

// QuoteCreateResponse represents the response for CreateQuote API
// [POL-API-003]
type QuoteCreateResponse struct {
	port.StatusCodeAndMessage
	QuoteID          string                     `json:"quote_id"`
	QuoteRefNumber   string                     `json:"quote_ref_number"`
	Status           string                     `json:"status"`
	PremiumBreakdown PremiumBreakdownResponse   `json:"premium_breakdown"`
	PDFDocument      *QuotePDFInfo              `json:"pdf_document,omitempty"`
	Validity         QuoteValidityInfo          `json:"validity"`
	WorkflowState    port.WorkflowStateResponse `json:"workflow_state,omitempty"`
}

// PremiumBreakdownResponse represents premium calculation breakdown
// [BR-POL-001] [BR-POL-002] [BR-POL-003]
type PremiumBreakdownResponse struct {
	BasePremium  float64 `json:"base_premium"`
	Rebate       float64 `json:"rebate"`
	NetPremium   float64 `json:"net_premium"`
	CGST         float64 `json:"cgst"`
	SGST         float64 `json:"sgst"`
	IGST         float64 `json:"igst"`
	TotalGST     float64 `json:"total_gst"`
	StampDuty    float64 `json:"stamp_duty,omitempty"`
	TotalPayable float64 `json:"total_payable"`
}

// QuoteConvertResponse represents the response for ConvertQuoteToProposal API
// [POL-API-004]
type QuoteConvertResponse struct {
	port.StatusCodeAndMessage
	ProposalID     int64  `json:"proposal_id"`
	ProposalNumber string `json:"proposal_number"`
	QuoteRefNumber string `json:"quote_ref_number"` 
	Status         string `json:"status"`
	RedirectURL    string `json:"redirect_url"`
}

// QuoteSummaryResponse represents a quote summary for listing
type QuoteSummaryResponse struct {
	QuoteID        string    `json:"quote_id"`
	QuoteRefNumber string    `json:"quote_ref_number"`
	ProductCode    string    `json:"product_code"`
	PolicyType     string    `json:"policy_type"`
	SumAssured     float64   `json:"sum_assured"`
	TotalPayable   float64   `json:"total_payable"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	ExpiresAt      time.Time `json:"expires_at"`
}

// QuoteListResponse represents the response for listing quotes
type QuoteListResponse struct {
	port.StatusCodeAndMessage
	port.MetaDataResponse
	Quotes []QuoteSummaryResponse `json:"quotes"`
}

// MapDomainToPremiumBreakdown maps domain PremiumBreakdown to response
func MapDomainToPremiumBreakdown(pb domain.PremiumBreakdown) PremiumBreakdownResponse {
	return PremiumBreakdownResponse{
		BasePremium:  pb.BasePremium,
		Rebate:       pb.Rebate,
		NetPremium:   pb.NetPremium,
		CGST:         pb.CGST,
		SGST:         pb.SGST,
		IGST:         pb.IGST,
		TotalGST:     pb.TotalGST,
		StampDuty:    pb.StampDuty,
		TotalPayable: pb.TotalPayable,
	}
}

// MapDomainToEligibility maps domain EligibilityResult to response
func MapDomainToEligibility(e domain.EligibilityResult) EligibilityResponse {
	return EligibilityResponse{
		IsEligible:   e.IsEligible,
		AgeAtEntry:   e.AgeAtEntry,
		MaturityAge:  e.MaturityAge,
		RejectReason: e.RejectReason,
	}
}

// MapDomainToBenefitIllustration maps domain BenefitIllustration to response
func MapDomainToBenefitIllustration(b domain.BenefitIllustration) BenefitIllustrationResponse {
	return BenefitIllustrationResponse{
		MaturityValueGuaranteed: b.MaturityValueGuaranteed,
		MaturityValueWithBonus:  b.MaturityValueWithBonus,
		IndicativeBonusRate:     b.IndicativeBonusRate,
		DeathBenefit:            b.DeathBenefit,
	}
}

type QuoteDetailResponse struct {
	QuoteID             int64                  `json:"quote_id"`
	QuoteRefNumber      string                 `json:"quote_ref_number"`
	ProductCode         string                 `json:"product_code"`
	PolicyType          string                 `json:"policy_type"`
	CustomerID          *int64                 `json:"customer_id,omitempty"`
	ProposerName        *string                `json:"proposer_name,omitempty"`
	ProposerDOB         *time.Time             `json:"proposer_dob,omitempty"`
	ProposerGender      *string                `json:"proposer_gender,omitempty"`
	ProposerMobile      *string                `json:"proposer_mobile,omitempty"`
	ProposerEmail       *string                `json:"proposer_email,omitempty"`
	SumAssured          float64                `json:"sum_assured"`
	PolicyTerm          int                    `json:"policy_term"`
	PaymentFrequency    string                 `json:"payment_frequency"`
	BasePremium         float64                `json:"base_premium"`
	GSTAmount           float64                `json:"gst_amount"`
	TotalPayable        float64                `json:"total_payable"`
	MaturityValue       *float64               `json:"maturity_value,omitempty"`
	BonusRate           *float64               `json:"bonus_rate,omitempty"`
	Channel             string                 `json:"channel"`
	Status              string                 `json:"status"`
	ConvertedProposalID *int64                 `json:"converted_proposal_id,omitempty"`
	PDFDocumentID       *string                `json:"pdf_document_id,omitempty"`
	// CreatedBy           int64                  `json:"created_by"`
	// CreatedAt           time.Time              `json:"created_at"`
	// UpdatedAt           time.Time              `json:"updated_at"`
	ExpiresAt           *time.Time             `json:"expires_at,omitempty"`
	Version             int                    `json:"version"`
	Metadata            map[string]interface{} `json:"metadata,omitempty"`
}
type GetQuoteResponse struct {
	port.StatusCodeAndMessage
	Quote QuoteDetailResponse `json:"quote"`
}

type QuoteGenerateResponse struct {
	port.StatusCodeAndMessage
	QuoteID          string                     `json:"quote_id"`
	QuoteRefNumber   string                     `json:"quote_ref_number"`
	ProductCategory  string                     `json:"product_category"`
	Status           string                     `json:"status"`
	PremiumBreakdown PremiumBreakdownResponse   `json:"premium_breakdown"`
	PDFDocument      *QuotePDFInfo              `json:"pdf_document,omitempty"`
	Validity         QuoteValidityInfo          `json:"validity"`
	CalculationBasis CalculationBasisResponse   `json:"calculation_basis"`
	Calculations     []QuoteCalculationItem     `json:"calculations"`
	WorkflowState    port.WorkflowStateResponse `json:"workflow_state"`
}
