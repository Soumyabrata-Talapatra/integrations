package response

import (
	"policy-issue-service/core/port"
)

// PolicyDetailResponse represents the response for Get Policy API
type PolicyDetailResponse struct {
	port.StatusCodeAndMessage
	ProposalID     int64   `json:"proposal_id"`
	ProposalNumber string  `json:"proposal_number"`
	PolicyNumber   *string `json:"policy_number"`
	Status         string  `json:"status"`
	CustomerID     int64   `json:"customer_id"`
	SumAssured     float64 `json:"sum_assured"`
}

// FLCRefundDetails contains the refund calculation breakdown
// [BR-POL-009] FLC Refund Calculation
type FLCRefundDetails struct {
	PremiumPaid       float64 `json:"premium_paid"`
	ProportionateRisk float64 `json:"proportionate_risk"`
	StampDuty         float64 `json:"stamp_duty"`
	MedicalFee        float64 `json:"medical_fee_deducted"`
	RefundAmount      float64 `json:"refund_amount"`
}

// FLCPeriodInfo contains the FLC window details
type FLCPeriodInfo struct {
	StartDate     string `json:"start_date"`
	EndDate       string `json:"end_date"`
	DaysRemaining int    `json:"days_remaining"`
	PeriodDays    int    `json:"period_days"`
}

// FLCCancelResponse represents the response for Cancel Policy FLC API
// [POL-API-022] Cancel Policy FLC
type FLCCancelResponse struct {
	port.StatusCodeAndMessage
	PolicyID           int64         `json:"policy_id"`
	CancellationStatus string        `json:"cancellation_status"`
	RefundDetails      *FLCRefundDetails `json:"refund_details,omitempty"`
	FLCPeriod          *FLCPeriodInfo    `json:"flc_period,omitempty"`
}

// FLCStatusResponse represents the response for Get FLC Status API
// [POL-API-024] Get FLC Status
type FLCStatusResponse struct {
	port.StatusCodeAndMessage
	PolicyID  int64          `json:"policy_id"`
	Status    string         `json:"status"`
	FLCPeriod *FLCPeriodInfo `json:"flc_period,omitempty"`
	Eligible  bool           `json:"eligible"`
}
