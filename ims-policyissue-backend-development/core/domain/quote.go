package domain

import (
	"time"
)

// QuoteStatus represents the status of a quote
// Aligned with user journey state diagram
type QuoteStatus string

const (
	QuoteStatusCalculating QuoteStatus = "CALCULATING"
	QuoteStatusFailed      QuoteStatus = "FAILED"
	QuoteStatusGenerated   QuoteStatus = "GENERATED"
	QuoteStatusSaved       QuoteStatus = "SAVED"
	QuoteStatusDiscarded   QuoteStatus = "DISCARDED"
	QuoteStatusConverted   QuoteStatus = "CONVERTED"
	QuoteStatusExpired     QuoteStatus = "EXPIRED"
)

// Channel represents the source channel for the quote
type Channel string

const (
	ChannelDirect Channel = "DIRECT"
	ChannelAgency Channel = "AGENCY"
	ChannelWeb    Channel = "WEB"
	ChannelMobile Channel = "MOBILE"
	ChannelPOS    Channel = "POS"
	ChannelCSC    Channel = "CSC"
)

// Gender represents gender options
type Gender string

const (
	GenderMale   Gender = "MALE"
	GenderFemale Gender = "FEMALE"
	GenderOther  Gender = "OTHER"
)

// Quote represents a premium quote for a potential policy
type Quote struct {
	QuoteID             int64                  `db:"quote_id" json:"quote_id"`
	QuoteRefNumber      string                 `db:"quote_ref_number" json:"quote_ref_number"`
	ProductCode         string                 `db:"product_code" json:"product_code"`
	PolicyType          PolicyType             `db:"policy_type" json:"policy_type"`
	CustomerID          *int64                 `db:"customer_id" json:"customer_id,omitempty"`
	ProposerName        *string                `db:"proposer_name" json:"proposer_name,omitempty"`
	ProposerDOB         *time.Time             `db:"proposer_dob" json:"proposer_dob,omitempty"`
	ProposerGender      *Gender                `db:"proposer_gender" json:"proposer_gender,omitempty"`
	ProposerMobile      *string                `db:"proposer_mobile" json:"proposer_mobile,omitempty"`
	ProposerEmail       *string                `db:"proposer_email" json:"proposer_email,omitempty"`
	SumAssured          float64                `db:"sum_assured" json:"sum_assured"`
	PolicyTerm          int                    `db:"policy_term" json:"policy_term"`
	PaymentFrequency    PremiumFrequency       `db:"payment_frequency" json:"payment_frequency"`
	BasePremium         float64                `db:"base_premium" json:"base_premium"`
	GSTAmount           float64                `db:"gst_amount" json:"gst_amount"`
	TotalPayable        float64                `db:"total_payable" json:"total_payable"`
	MaturityValue       *float64               `db:"maturity_value" json:"maturity_value,omitempty"`
	BonusRate           *float64               `db:"bonus_rate" json:"bonus_rate,omitempty"`
	Channel             Channel                `db:"channel" json:"channel"`
	Status              QuoteStatus            `db:"status" json:"status"`
	ConvertedProposalID *int64                 `db:"converted_proposal_id" json:"converted_proposal_id,omitempty"`
	PDFDocumentID       *string                `db:"pdf_document_id" json:"pdf_document_id,omitempty"`
	CreatedBy           int64                  `db:"created_by" json:"created_by"`
	CreatedAt           time.Time              `db:"created_at" json:"created_at"`
	UpdatedAt           time.Time              `db:"updated_at" json:"updated_at"`
	ExpiresAt           *time.Time             `db:"expires_at" json:"expires_at,omitempty"`
	DeletedAt           *time.Time             `db:"deleted_at" json:"deleted_at,omitempty"`
	Version             int                    `db:"version" json:"version"`
	Metadata            map[string]interface{} `db:"metadata" json:"metadata,omitempty"`
	Search_Vector       *string                `db:"search_vector" json:"-"`
	Rebate              *float64               `db:"rebate" json:"rebate,omitempty"`
}

// IsExpired checks if the quote has expired
func (q *Quote) IsExpired() bool {
	if q.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*q.ExpiresAt)
}

// CanBeConverted checks if the quote can be converted to a proposal
func (q *Quote) CanBeConverted() bool {
	if q.Status != QuoteStatusGenerated {
		return false
	}
	if q.IsExpired() {
		return false
	}
	return true
}

// PremiumBreakdown contains detailed premium calculation
type PremiumBreakdown struct {
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

// BenefitIllustration contains projected benefits
type BenefitIllustration struct {
	MaturityValueGuaranteed float64 `json:"maturity_value_guaranteed"`
	MaturityValueWithBonus  float64 `json:"maturity_value_with_bonus"`
	IndicativeBonusRate     float64 `json:"indicative_bonus_rate"`
	DeathBenefit            float64 `json:"death_benefit"`
}

// EligibilityResult contains eligibility check results
type EligibilityResult struct {
	IsEligible   bool   `json:"is_eligible"`
	AgeAtEntry   int    `json:"age_at_entry"`
	MaturityAge  int    `json:"maturity_age"`
	RejectReason string `json:"reject_reason,omitempty"`
}

// CalculationBasis contains the basis for premium calculation
type CalculationBasis struct {
	PremiumTable    string  `json:"premium_table"`
	RatePerThousand float64 `json:"rate_per_thousand"`
	GSTRate         float64 `json:"gst_rate"`
}

// QuoteDetail extends Quote with additional details
type QuoteDetail struct {
	Quote
	Eligibility         EligibilityResult   `json:"eligibility"`
	PremiumBreakdown    PremiumBreakdown    `json:"premium_breakdown"`
	BenefitIllustration BenefitIllustration `json:"benefit_illustration"`
	CalculationBasis    CalculationBasis    `json:"calculation_basis"`
}

// PremiumRate represents a rate from the Sankalan table
type PremiumRate struct {
	ProductCode   string     `db:"product_code" json:"product_code"`
	Age           int        `db:"age" json:"age"`
	Gender        Gender     `db:"gender" json:"gender"`
	Term          int        `db:"term" json:"term"`
	RatePer1000   float64    `db:"rate_per_1000" json:"rate_per_1000"`
	EffectiveFrom time.Time  `db:"effective_from" json:"effective_from"`
	EffectiveTo   *time.Time `db:"effective_to" json:"effective_to,omitempty"`
}

type TermOrPremiumCeasingAge struct {
	ProductCode       string  `db:"product_code"`
	AgeATEntry        *int    `db:"age_at_entry"`
	Periodicity       *string `db:"periodicity"`
	Term              *int    `db:"term"`
	PremiumCeasingAge *int    `db:"premium_ceasing_age"`
}
