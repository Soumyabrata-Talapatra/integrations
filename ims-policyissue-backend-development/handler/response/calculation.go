package response

import "policy-issue-service/core/port"

// --- CALC-POL-001: Premium Calculation ---

// PremiumCalcResponse represents the response for premium calculation preview
// [CALC-POL-001] Full premium calculation
// Components: BR-POL-001, BR-POL-002, BR-POL-010
type PremiumCalcResponse struct {
	port.StatusCodeAndMessage
	BasePremium       float64 `json:"base_premium"`
	LoadedPremium     float64 `json:"loaded_premium"`
	RebateAmount      float64 `json:"rebate_amount"`
	NetPremium        float64 `json:"net_premium"`
	CGST              float64 `json:"cgst"`
	SGST              float64 `json:"sgst"`
	IGST              float64 `json:"igst"`
	TotalGST          float64 `json:"total_gst"`
	StampDuty         float64 `json:"stamp_duty"`
	MedicalFee        float64 `json:"medical_fee"`
	TotalFirstPayment float64 `json:"total_first_payment"`
	ModalPremium      float64 `json:"modal_premium"`
}

// --- CALC-POL-002: Maturity Value Calculation ---

// MaturityCalcResponse represents the response for maturity value estimation
// [CALC-POL-002] Maturity value estimation
// Components: FR-POL-001
type MaturityCalcResponse struct {
	port.StatusCodeAndMessage
	GuaranteedSumAssured    float64 `json:"guaranteed_sum_assured"`
	IndicativeBonus         float64 `json:"indicative_bonus"`
	IndicativeMaturityValue float64 `json:"indicative_maturity_value"`
	BonusRateUsed           float64 `json:"bonus_rate_used"`
	Disclaimer              string  `json:"disclaimer"`
}

// --- CALC-POL-003: FLC Refund Calculation ---

// FLCRefundCalcResponse represents the response for FLC refund preview
// [CALC-POL-003] FLC refund preview
// Components: BR-POL-009
type FLCRefundCalcResponse struct {
	port.StatusCodeAndMessage
	PremiumPaid              float64 `json:"premium_paid"`
	ProportionateRiskPremium float64 `json:"proportionate_risk_premium"`
	StampDutyDeduction       float64 `json:"stamp_duty_deduction"`
	MedicalFeeDeduction      float64 `json:"medical_fee_deduction"`
	TotalDeductions          float64 `json:"total_deductions"`
	NetRefundAmount          float64 `json:"net_refund_amount"`
}

// --- CALC-POL-004: GST Calculation ---

// GSTCalcResponse represents the response for GST breakdown
// [CALC-POL-004] GST breakdown
// Components: BR-POL-002
type GSTCalcResponse struct {
	port.StatusCodeAndMessage
	BaseAmount float64 `json:"base_amount"`
	GSTRate    float64 `json:"gst_rate"`
	CGST       float64 `json:"cgst"`
	SGST       float64 `json:"sgst"`
	IGST       float64 `json:"igst"`
	TotalGST   float64 `json:"total_gst"`
	GSTType    string  `json:"gst_type"`
}

type GetTermAndPremiumCeasingAgeResponse struct {
	StatusCodeAndMessage port.StatusCodeAndMessage `json:",inline"`
	ProductCode          string                    `json:"product_code"`
	AgeAtEntry           int                       `json:"age_at_entry"`
	Periodicity          []string                  `json:"periodicity"`
	Terms                []int                     `json:"terms,omitempty"`
	PremiumCeasingAges   []int                     `json:"premium_ceasing_ages,omitempty"`
}
