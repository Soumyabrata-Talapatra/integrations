package response

import (
	"time"

	"gitlab.cept.gov.in/it-2.0-policy/surrender-service/core/domain"
	"gitlab.cept.gov.in/it-2.0-policy/surrender-service/core/port"
)

// ============================================
// Eligibility Response DTOs
// ============================================

// EligibilityEligibleResponse represents a successful eligibility check
type EligibilityEligibleResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	Data                      EligibilityEligibleData `json:"data"`
}

type EligibilityEligibleData struct {
	Eligible            bool   `json:"eligible"`
	PolicyID            string `json:"policy_id"`
	PolicyNumber        string `json:"policy_number"`
	ProductCode         string `json:"product_code"`
	ProductName         string `json:"product_name"`
	PremiumsPaid        int    `json:"premiums_paid"`
	MinimumPremiumsPaid int    `json:"minimum_premiums_paid"`
	PolicyStatus        string `json:"policy_status"`
	Message             string `json:"message"`
}

type EligibilityResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	//Data                      EligibilityEligibleData `json:"data"`
}

// EligibilityIneligibleResponse represents an ineligibility response
type EligibilityIneligibleResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	Data                      EligibilityIneligibleData `json:"data"`
}

type EligibilityIneligibleData struct {
	Eligible     bool     `json:"eligible"`
	PolicyID     string   `json:"policy_id"`
	PolicyNumber string   `json:"policy_number"`
	Reasons      []string `json:"reasons"`
	Message      string   `json:"message"`
}

type IndexSurrenderResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	Data                      IndexSurrenderResponseData `json:"data"`
}

type IndexSurrenderResponseData struct {
	ServiceRequestID string `json:"service_id"`
}

type GetDEPendingResponse struct {
	StatusCode int    `json:"statusCode"`
	Success    bool   `json:"success"`
	Message    string `json:"message"`
	Data       any    `json:"data"`
}

// ============================================
// Calculation Response DTOs
// ============================================

// CalculateSurrenderResponse represents surrender value calculation
type CalculateSurrenderResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	Data                      CalculationData `json:"data"`
}

type CalculationData struct {
	CalculationBreakdown  CalculationBreakdownData  `json:"calculation_breakdown"`
	DisbursementOptions   DisbursementOptionsData   `json:"disbursement_options"`
	DispositionPrediction DispositionPredictionData `json:"disposition_prediction"`
}

type CalculationBreakdownData struct {
	PolicyID            string            `json:"policy_id"`
	CalculationDate     string            `json:"calculation_date"`
	SumAssured          float64           `json:"sum_assured"`
	PremiumsPaid        int               `json:"premiums_paid"`
	TotalPremiums       int               `json:"total_premiums"`
	PaidUpValue         float64           `json:"paid_up_value"`
	BonusDetails        []BonusDetailData `json:"bonus_details"`
	TotalBonus          float64           `json:"total_bonus"`
	SurrenderFactor     float64           `json:"surrender_factor"`
	GrossSurrenderValue float64           `json:"gross_surrender_value"`
	Deductions          DeductionsData    `json:"deductions"`
	NetSurrenderValue   float64           `json:"net_surrender_value"`
}

type BonusDetailData struct {
	FinancialYear string  `json:"financial_year"`
	SumAssured    float64 `json:"sum_assured"`
	BonusRate     float64 `json:"bonus_rate"`
	BonusAmount   float64 `json:"bonus_amount"`
}

type DeductionsData struct {
	UnpaidPremiums  float64 `json:"unpaid_premiums"`
	LoanPrincipal   float64 `json:"loan_principal"`
	LoanInterest    float64 `json:"loan_interest"`
	TotalLoan       float64 `json:"total_loan"`
	OtherCharges    float64 `json:"other_charges"`
	TotalDeductions float64 `json:"total_deductions"`
}

type DisbursementOptionsData struct {
	CashAvailable   bool             `json:"cash_available"`
	ChequeAvailable bool             `json:"cheque_available"`
	PayeeDetails    PayeeDetailsData `json:"payee_details"`
}

type PayeeDetailsData struct {
	PayeeName    string `json:"payee_name"`
	PayeeAddress string `json:"payee_address"`
	IsAssigned   bool   `json:"is_assigned"`
	AssigneeName string `json:"assignee_name,omitempty"`
}

type DispositionPredictionData struct {
	PredictedDisposition string  `json:"predicted_disposition"`
	PrescribedLimit      float64 `json:"prescribed_limit"`
	NetAmount            float64 `json:"net_amount"`
	WillCreateReducedPU  bool    `json:"will_create_reduced_paid_up"`
	NewSumAssured        float64 `json:"new_sum_assured,omitempty"`
	NewPolicyStatus      string  `json:"new_policy_status"`
}

// ============================================
// Confirm Surrender Response DTOs
// ============================================

// ConfirmSurrenderResponse represents successful surrender confirmation
type ConfirmSurrenderResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	Data                      ConfirmSurrenderData `json:"data"`
}

type ConfirmSurrenderData struct {
	SurrenderRequestID   string                   `json:"surrender_request_id"`
	RequestNumber        string                   `json:"request_number"`
	PolicyID             string                   `json:"policy_id"`
	PolicyNumber         string                   `json:"policy_number"`
	Status               string                   `json:"status"`
	RequestDate          string                   `json:"request_date"`
	NetSurrenderValue    float64                  `json:"net_surrender_value"`
	DisbursementMethod   string                   `json:"disbursement_method"`
	DocumentRequirements DocumentRequirementsData `json:"document_requirements"`
	WorkflowState        WorkflowStateData        `json:"workflow_state"`
	NextAction           NextActionData           `json:"next_action"`
}

type DocumentRequirementsData struct {
	Required             []DocumentRequirementData `json:"required"`
	TotalRequired        int                       `json:"total_required"`
	TotalUploaded        int                       `json:"total_uploaded"`
	AllDocumentsUploaded bool                      `json:"all_documents_uploaded"`
}

type DocumentRequirementData struct {
	DocumentType string `json:"document_type"`
	DisplayName  string `json:"display_name"`
	Mandatory    bool   `json:"mandatory"`
	Uploaded     bool   `json:"uploaded"`
	Description  string `json:"description"`
}

type WorkflowStateData struct {
	CurrentStage    string   `json:"current_stage"`
	CompletedStages []string `json:"completed_stages"`
	PendingStages   []string `json:"pending_stages"`
	ProgressPercent int      `json:"progress_percent"`
}

type NextActionData struct {
	Action      string `json:"action"`
	Description string `json:"description"`
	URL         string `json:"url"`
}

// ============================================
// Surrender Status Response DTOs
// ============================================

// SurrenderStatusResponse represents surrender request status
type SurrenderStatusResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	Data                      SurrenderStatusData `json:"data"`
}

type SurrenderStatusData struct {
	SurrenderRequestID string                `json:"surrender_request_id"`
	RequestNumber      string                `json:"request_number"`
	PolicyID           string                `json:"policy_id"`
	PolicyNumber       string                `json:"policy_number"`
	RequestType        string                `json:"request_type"`
	Status             string                `json:"status"`
	RequestDate        string                `json:"request_date"`
	NetSurrenderValue  float64               `json:"net_surrender_value"`
	WorkflowState      WorkflowStateData     `json:"workflow_state"`
	History            []RequestHistoryData  `json:"history,omitempty"`
	Details            *SurrenderDetailsData `json:"details,omitempty"`
}

type RequestHistoryData struct {
	Timestamp string `json:"timestamp"`
	Action    string `json:"action"`
	OldStatus string `json:"old_status,omitempty"`
	NewStatus string `json:"new_status,omitempty"`
	ChangedBy string `json:"changed_by"`
	Comments  string `json:"comments,omitempty"`
}

type SurrenderDetailsData struct {
	PolicyDetails      PolicyDetailsData        `json:"policy_details"`
	CalculationDetails CalculationBreakdownData `json:"calculation_details"`
	Documents          []DocumentInfoData       `json:"documents"`
	PaymentDetails     *PaymentDetailsData      `json:"payment_details,omitempty"`
	ApprovalDetails    *ApprovalDetailsData     `json:"approval_details,omitempty"`
}

type PolicyDetailsData struct {
	PolicyID         string  `json:"policy_id"`
	PolicyNumber     string  `json:"policy_number"`
	ProductCode      string  `json:"product_code"`
	ProductName      string  `json:"product_name"`
	SumAssured       float64 `json:"sum_assured"`
	PolicyStatus     string  `json:"policy_status"`
	CommencementDate string  `json:"commencement_date"`
	MaturityDate     string  `json:"maturity_date"`
}

type PaymentDetailsData struct {
	PaymentNumber      string  `json:"payment_number"`
	PaymentDate        string  `json:"payment_date"`
	Amount             float64 `json:"amount"`
	DisbursementMethod string  `json:"disbursement_method"`
	PayeeName          string  `json:"payee_name"`
	Status             string  `json:"status"`
}

type ApprovalDetailsData struct {
	ApprovedBy string `json:"approved_by"`
	ApprovedAt string `json:"approved_at"`
	Comments   string `json:"comments"`
}

// ============================================
// Helper Functions
// ============================================

// NewSurrenderRequest converts domain model to response DTO
func NewSurrenderRequest(d domain.PolicySurrenderRequest) SurrenderRequestResponse {
	return SurrenderRequestResponse{
		ID:                  d.ID.String(),
		PolicyID:            d.PolicyID,
		RequestNumber:       d.RequestNumber,
		RequestType:         string(d.RequestType),
		Status:              string(d.Status),
		RequestDate:         d.RequestDate.Format(time.RFC3339),
		GrossSurrenderValue: d.GrossSurrenderValue,
		NetSurrenderValue:   d.NetSurrenderValue,
		PaidUpValue:         d.PaidUpValue,
		DisbursementMethod:  string(d.DisbursementMethod),
		DisbursementAmount:  d.DisbursementAmount,
		CreatedAt:           d.CreatedAt.Format(time.RFC3339),
		UpdatedAt:           d.UpdatedAt.Format(time.RFC3339),
	}
}

type SurrenderRequestResponse struct {
	ID                  string  `json:"id"`
	PolicyID            string  `json:"policy_id"`
	RequestNumber       string  `json:"request_number"`
	RequestType         string  `json:"request_type"`
	Status              string  `json:"status"`
	RequestDate         string  `json:"request_date"`
	GrossSurrenderValue float64 `json:"gross_surrender_value"`
	NetSurrenderValue   float64 `json:"net_surrender_value"`
	PaidUpValue         float64 `json:"paid_up_value"`
	DisbursementMethod  string  `json:"disbursement_method"`
	DisbursementAmount  float64 `json:"disbursement_amount"`
	CreatedAt           string  `json:"created_at"`
	UpdatedAt           string  `json:"updated_at"`
}
