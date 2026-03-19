package domain

import (
	"time"

	"github.com/google/uuid"
)

// SurrenderPayment represents a payment made for surrender
// Table: surrender_payments
// Functional Requirement: FR-SUR-007
type SurrenderPayment struct {
	ID                   uuid.UUID              `json:"id" db:"id"`
	SurrenderRequestID   uuid.UUID              `json:"surrender_request_id" db:"surrender_request_id"`
	PaymentNumber        string                 `json:"payment_number" db:"payment_number"`
	PaymentDate          time.Time              `json:"payment_date" db:"payment_date"`
	Amount               float64                `json:"amount" db:"amount"`
	DisbursementMethod   DisbursementMethod     `json:"disbursement_method" db:"disbursement_method"`
	ChequeNumber         *string                `json:"cheque_number" db:"cheque_number"`
	ChequeDate           *time.Time             `json:"cheque_date" db:"cheque_date"`
	BankName             *string                `json:"bank_name" db:"bank_name"`
	BranchName           *string                `json:"branch_name" db:"branch_name"`
	PayeeName            string                 `json:"payee_name" db:"payee_name"`
	PayeeAddress         *string                `json:"payee_address" db:"payee_address"`
	TransactionReference *string                `json:"transaction_reference" db:"transaction_reference"`
	Status               string                 `json:"status" db:"status"`
	ProcessedAt          *time.Time             `json:"processed_at" db:"processed_at"`
	ProcessedBy          *uuid.UUID             `json:"processed_by" db:"processed_by"`
	CreatedAt            time.Time              `json:"created_at" db:"created_at"`
	Metadata             map[string]interface{} `json:"metadata" db:"metadata"`
}

// SurrenderValueCalculation represents a calculation audit trail
// Table: surrender_value_calculations
// Functional Requirement: FR-SUR-002
type SurrenderValueCalculation struct {
	ID                      uuid.UUID              `json:"id" db:"id"`
	SurrenderRequestID      uuid.UUID              `json:"surrender_request_id" db:"surrender_request_id"`
	CalculationDate         time.Time              `json:"calculation_date" db:"calculation_date"`
	PaidUpValue             float64                `json:"paid_up_value" db:"paid_up_value"`
	BonusAmount             *float64               `json:"bonus_amount" db:"bonus_amount"`
	SurrenderFactor         float64                `json:"surrender_factor" db:"surrender_factor"`
	GrossSurrenderValue     float64                `json:"gross_surrender_value" db:"gross_surrender_value"`
	UnpaidPremiumsDeduction float64                `json:"unpaid_premiums_deduction" db:"unpaid_premiums_deduction"`
	LoanPrincipalDeduction  float64                `json:"loan_principal_deduction" db:"loan_principal_deduction"`
	LoanInterestDeduction   float64                `json:"loan_interest_deduction" db:"loan_interest_deduction"`
	NetSurrenderValue       float64                `json:"net_surrender_value" db:"net_surrender_value"`
	CalculationBreakdown    map[string]interface{} `json:"calculation_breakdown" db:"calculation_breakdown"`
	CalculatedBy            uuid.UUID              `json:"calculated_by" db:"calculated_by"`
	CreatedAt               time.Time              `json:"created_at" db:"created_at"`
}

// PolicySurrenderDisposition represents the final disposition after approval
// Table: policy_surrender_dispositions
// Business Rules: BR-SUR-011, BR-SUR-012
type PolicySurrenderDisposition struct {
	ID                    uuid.UUID              `json:"id" db:"id"`
	SurrenderRequestID    uuid.UUID              `json:"surrender_request_id" db:"surrender_request_id"`
	DispositionType       string                 `json:"disposition_type" db:"disposition_type"`
	NewPolicyStatus       *PolicyStatusSurrender `json:"new_policy_status" db:"new_policy_status"`
	NewSumAssured         *float64               `json:"new_sum_assured" db:"new_sum_assured"`
	PrescribedLimit       *float64               `json:"prescribed_limit" db:"prescribed_limit"`
	NetSurrenderValue     float64                `json:"net_surrender_value" db:"net_surrender_value"`
	ReducedPaidUpCreated  bool                   `json:"reduced_paid_up_created" db:"reduced_paid_up_created"`
	ReducedPaidUpPolicyNo *string                `json:"reduced_paid_up_policy_number" db:"reduced_paid_up_policy_number"`
	Terminated            bool                   `json:"terminated" db:"terminated"`
	TerminationReason     *string                `json:"termination_reason" db:"termination_reason"`
	CreatedAt             time.Time              `json:"created_at" db:"created_at"`
}

// Disposition types
const (
	DispositionTypeReducedPaidUp           = "REDUCED_PAID_UP"
	DispositionTypeTerminatedSurrender     = "TERMINATED_SURRENDER"
	DispositionTypeTerminatedAutoSurrender = "TERMINATED_AUTO_SURRENDER"
)
