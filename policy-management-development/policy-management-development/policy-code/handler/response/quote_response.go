package response

import "policy-management/core/port"

// ============================================================================
// Quote Response DTOs
// Source: Swagger components/schemas: SurrenderQuoteResponse, LoanQuoteResponse,
//         ConversionQuoteResponse
// Used by: GET /policies/{pn}/quotes/surrender, /loan, /conversion
// ============================================================================

// ── GET /policies/{pn}/quotes/surrender ──────────────────────────────────────

// SurrenderBreakdown contains the actuarial breakdown for a surrender quote.
type SurrenderBreakdown struct {
	BaseSVFactor         float64 `json:"base_sv_factor"`
	PremiumsPaidMonths   int     `json:"premiums_paid_months"`
	TotalPremiumsPayable int     `json:"total_premiums_payable"`
}

// SurrenderQuoteData is the payload returned for a surrender quote request.
// Produced by the downstream Surrender Service via a Temporal query activity.
type SurrenderQuoteData struct {
	PolicyNumber        string             `json:"policy_number"`
	QuoteType           string             `json:"quote_type"`            // "SURRENDER"
	GrossSurrenderValue float64            `json:"gross_surrender_value"` // GSV before deductions
	BonusAccumulated    float64            `json:"bonus_accumulated"`
	LoanDeduction       float64            `json:"loan_deduction"`       // Outstanding principal
	InterestDeduction   float64            `json:"interest_deduction"`   // Accrued interest
	NetSurrenderValue   float64            `json:"net_surrender_value"`  // GSV + Bonus - Deductions
	Breakdown           SurrenderBreakdown `json:"breakdown"`
	ValidUntil          string             `json:"valid_until"` // RFC3339 — quote expiry
}

// SurrenderQuoteResponse — GET /api/v1/policies/{pn}/quotes/surrender
type SurrenderQuoteResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	Data                      SurrenderQuoteData `json:"data"`
}

// ── GET /policies/{pn}/quotes/loan ───────────────────────────────────────────

// LoanQuoteData is the payload returned for a loan eligibility and quote request.
// Produced by the downstream Loan Service via a Temporal query activity.
type LoanQuoteData struct {
	PolicyNumber        string  `json:"policy_number"`
	Eligible            bool    `json:"eligible"`
	MaxLoanAmount       float64 `json:"max_loan_amount"`
	InterestRate        float64 `json:"interest_rate"`        // Annual rate (e.g. 0.12 for 12%)
	SurrenderValue      float64 `json:"surrender_value"`      // Current SV used for eligibility
	ExistingLoanBalance float64 `json:"existing_loan_balance"` // Outstanding principal on active loan
	IneligibilityReason *string `json:"ineligibility_reason,omitempty"` // Set if Eligible = false
}

// LoanQuoteResponse — GET /api/v1/policies/{pn}/quotes/loan
type LoanQuoteResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	Data                      LoanQuoteData `json:"data"`
}

// ── GET /policies/{pn}/quotes/conversion ─────────────────────────────────────

// ConversionOption represents a single available product conversion target.
type ConversionOption struct {
	TargetProduct     string  `json:"target_product"`
	EffectiveDate     string  `json:"effective_date"` // Date only: "2006-01-02"
	NewPremium        float64 `json:"new_premium"`
	PremiumDifference float64 `json:"premium_difference"` // Positive = higher than current
	NewSumAssured     float64 `json:"new_sum_assured"`
}

// ConversionQuoteData is the payload returned for a conversion options quote request.
// Produced by the downstream Conversion Service via a Temporal query activity.
type ConversionQuoteData struct {
	PolicyNumber         string             `json:"policy_number"`
	CurrentProduct       string             `json:"current_product"`       // e.g. "WLA", "EA"
	AvailableConversions []ConversionOption `json:"available_conversions"` // Empty if no conversions allowed
}

// ConversionQuoteResponse — GET /api/v1/policies/{pn}/quotes/conversion
type ConversionQuoteResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	Data                      ConversionQuoteData `json:"data"`
}
