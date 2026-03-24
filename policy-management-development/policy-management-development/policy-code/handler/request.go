package handler

import (
	"fmt"

	"policy-management/core/domain"
	"policy-management/core/port"
)

// ============================================================================
// Handler Request DTOs
// Source: Swagger pm_swagger.yaml (components/schemas/*)
// Validation tags follow n-api-template pattern.
// All request structs bind to JSON body unless noted (form/uri tags).
// ============================================================================

// ============================================================================
// URI parameter structs (for path params)
// ============================================================================

// PolicyNumberURI binds the {policy_number} path parameter.
// VR-PM-001: Pattern PLI/YYYY/NNNNNN (10-16 chars) or RPLI/YYYY/NNNNNN (11-17 chars).
// Min/max bounds enforce basic format; full regex validation is registered as a
// custom validator "policy_number" in the bootstrapper (VR-PM-001).
type PolicyNumberURI struct {
	PolicyNumber string `uri:"policy_number" validate:"required,min=10,max=20"`
}

// RequestIDURI binds the {request_id} path parameter.
// request_id is a BIGINT from seq_service_request_id — NOT a UUID. [Gap-1]
type RequestIDURI struct {
	RequestID int64 `uri:"request_id" validate:"required,min=1"`
}

// StateGateTypeURI binds the {request_type} path parameter for state-gate endpoint.
// Only REST-exposed request types are valid. [Enhancement-1]
type StateGateTypeURI struct {
	RequestType string `uri:"request_type" validate:"required,oneof=SURRENDER LOAN LOAN_REPAYMENT REVIVAL DEATH_CLAIM MATURITY_CLAIM SURVIVAL_BENEFIT COMMUTATION CONVERSION FLC PAID_UP NOMINATION_CHANGE BILLING_METHOD_CHANGE ASSIGNMENT ADDRESS_CHANGE PREMIUM_REFUND DUPLICATE_BOND"`
}

// ============================================================================
// STEP 1: Generic Request Payload base (embedded in all submission requests)
// Source: Swagger GenericRequestPayload
// ============================================================================

// GenericRequestPayload is embedded in all request submission structs.
type GenericRequestPayload struct {
	SourceChannel string `json:"source_channel" validate:"required,oneof=CUSTOMER_PORTAL CPC MOBILE_APP AGENT_PORTAL BATCH SYSTEM"`
	SubmittedBy   *int64 `json:"submitted_by"   validate:"omitempty,min=1"` // BIGINT user ID; nil for batch/system [Gap-1]
	Notes         string `json:"notes"          validate:"omitempty,max=1000"`
}

// ============================================================================
// STEP 2: Financial Request Submission DTOs (11 endpoints)
// All go to PolicyLifecycleWorkflow via Temporal signal
// ============================================================================

// SubmitSurrenderRequest — POST /policies/{pn}/requests/surrender
// State Gate: BR-PM-011 — Allowed: ACTIVE, VL, IL, AL, PAID_UP
// [FR-PM-001] [BR-PM-011] [BR-PM-030]
type SubmitSurrenderRequest struct {
	GenericRequestPayload
	Payload struct {
		DisbursementMethod string `json:"disbursement_method" validate:"omitempty,oneof=NEFT CHEQUE CASH MONEY_ORDER"`
		BankAccountID      int64  `json:"bank_account_id"     validate:"omitempty,min=1"` // Required for NEFT; BIGINT from billing service
		Reason             string `json:"reason"              validate:"omitempty"`
	} `json:"payload" validate:"omitempty"`
}

// SubmitLoanRequest — POST /policies/{pn}/requests/loan
// State Gate: BR-PM-012 — Allowed: ACTIVE (+ has_active_loan must be false)
// [FR-PM-001] [BR-PM-012] [BR-PM-030]
type SubmitLoanRequest struct {
	GenericRequestPayload
	Payload struct {
		RequestedAmount    float64 `json:"requested_amount"    validate:"omitempty,gt=0"` // Optional — max eligible if omitted
		DisbursementMethod string  `json:"disbursement_method" validate:"omitempty,oneof=NEFT CHEQUE"`
	} `json:"payload" validate:"omitempty"`
}

// SubmitLoanRepaymentRequest — POST /policies/{pn}/requests/loan-repayment
// State Gate: BR-PM-020 — Allowed: ACTIVE, ATP, PAS (+ has_active_loan must be true)
// No financial lock for loan repayment.
// [FR-PM-001] [BR-PM-020]
type SubmitLoanRepaymentRequest struct {
	GenericRequestPayload
	Payload struct {
		LoanID      int64   `json:"loan_id"      validate:"required,min=1"` // BIGINT from loan service
		Amount      float64 `json:"amount"       validate:"required,gt=0"`
		PaymentMode string  `json:"payment_mode" validate:"required,oneof=CASH CHEQUE ONLINE PAY_RECOVERY"`
	} `json:"payload" validate:"required"`
}

// SubmitRevivalRequest — POST /policies/{pn}/requests/revival
// State Gate: BR-PM-013 — Allowed: VL, IL, AL
// [FR-PM-001] [BR-PM-013] [BR-PM-030]
type SubmitRevivalRequest struct {
	GenericRequestPayload
	Payload struct {
		RequestedInstallments int `json:"requested_installments" validate:"omitempty,min=2,max=12"`
	} `json:"payload" validate:"omitempty"`
}

// SubmitDeathClaimRequest — POST /policies/{pn}/requests/death-claim
// State Gate: BR-PM-014 — Allowed: ALL non-terminal (including SUSPENDED — BR-PM-112)
// PREEMPTIVE: cancels active financial operation (BR-PM-031). No financial lock.
// [FR-PM-001] [BR-PM-014] [BR-PM-031] [BR-PM-112]
type SubmitDeathClaimRequest struct {
	GenericRequestPayload
	Payload struct {
		CustomerID           int64  `json:"customer_id"           validate:"required,min=1"` // BIGINT from customer service
		DateOfDeath          string `json:"date_of_death"         validate:"required"`        // RFC3339 date
		CauseOfDeath         string `json:"cause_of_death"        validate:"omitempty,oneof=NATURAL ACCIDENTAL SUICIDE UNKNOWN"`
		ReportedBy           string `json:"reported_by"           validate:"required,oneof=NOMINEE LEGAL_HEIR ASSIGNEE POST_OFFICE"`
		ClaimantID           int64  `json:"claimant_id"           validate:"omitempty,min=1"` // BIGINT from claims service
		ClaimantRelationship string `json:"claimant_relationship" validate:"required,oneof=SPOUSE CHILD PARENT SIBLING LEGAL_HEIR ASSIGNEE OTHER"`
	} `json:"payload" validate:"required"`
}

// SubmitMaturityClaimRequest — POST /policies/{pn}/requests/maturity-claim
// State Gate: BR-PM-015 — Allowed: ACTIVE, PENDING_MATURITY
// [FR-PM-001] [BR-PM-015] [BR-PM-030]
type SubmitMaturityClaimRequest struct {
	GenericRequestPayload
	Payload struct {
		DisbursementMethod string `json:"disbursement_method" validate:"omitempty,oneof=NEFT CHEQUE CASH"`
		BankAccountID      int64  `json:"bank_account_id"     validate:"omitempty,min=1"` // Required for NEFT; BIGINT from billing service
	} `json:"payload" validate:"omitempty"`
}

// SubmitSurvivalBenefitRequest — POST /policies/{pn}/requests/survival-benefit
// State Gate: BR-PM-016 — Allowed: ACTIVE
// [FR-PM-001] [BR-PM-016] [BR-PM-030]
type SubmitSurvivalBenefitRequest struct {
	GenericRequestPayload
	Payload struct {
		SBInstallmentNumber int    `json:"sb_installment_number" validate:"required,min=1"`
		DisbursementMethod  string `json:"disbursement_method"   validate:"omitempty,oneof=NEFT CHEQUE CASH"`
		BankAccountID       int64  `json:"bank_account_id"       validate:"omitempty,min=1"` // BIGINT from billing service
	} `json:"payload" validate:"required"`
}

// SubmitCommutationRequest — POST /policies/{pn}/requests/commutation
// State Gate: BR-PM-017 — Allowed: ACTIVE
// [FR-PM-001] [BR-PM-017] [BR-PM-030]
type SubmitCommutationRequest struct {
	GenericRequestPayload
	Payload struct {
		CommutationPercentage float64 `json:"commutation_percentage" validate:"omitempty,min=0,max=100"`
	} `json:"payload" validate:"omitempty"`
}

// SubmitConversionRequest — POST /policies/{pn}/requests/conversion
// State Gate: BR-PM-018 — Allowed: ACTIVE
// [FR-PM-001] [BR-PM-018] [BR-PM-030]
type SubmitConversionRequest struct {
	GenericRequestPayload
	Payload struct {
		TargetProductCode string `json:"target_product_code" validate:"required"`
	} `json:"payload" validate:"required"`
}

// SubmitFreelookRequest — POST /policies/{pn}/requests/freelook
// State Gate: BR-PM-019 — Allowed: FREE_LOOK_ACTIVE only (15/30-day timer)
// [FR-PM-001] [BR-PM-019] [BR-PM-030]
type SubmitFreelookRequest struct {
	GenericRequestPayload
	Payload struct {
		Reason                  string `json:"reason"                   validate:"required"`
		OriginalBondSurrendered bool   `json:"original_bond_surrendered" validate:"omitempty"`
	} `json:"payload" validate:"required"`
}

// SubmitPaidUpRequest — POST /policies/{pn}/requests/paid-up
// State Gate: BR-PM-022 — Allowed: ACTIVE, ACTIVE_LAPSE
// PM-internal: no downstream service; PM calculates PU value and transitions.
// [FR-PM-001] [BR-PM-022] [BR-PM-030] [BR-PM-060] [BR-PM-061]
type SubmitPaidUpRequest struct {
	GenericRequestPayload
	// Paid-up uses GenericRequestPayload only — no specific payload fields
}

// ============================================================================
// STEP 3: NFR Request Submission DTOs (6 endpoints)
// No financial lock (BR-PM-023). Run concurrently with financial requests.
// ============================================================================

// SubmitNominationChangeRequest — POST /policies/{pn}/requests/nomination-change
// [FR-PM-001] [BR-PM-023]
type SubmitNominationChangeRequest struct {
	GenericRequestPayload
	Payload struct {
		Nominees []NomineeDetail `json:"nominees" validate:"required,min=1,dive"`
	} `json:"payload" validate:"required"`
}

// NomineeDetail represents a single nominee entry.
type NomineeDetail struct {
	Name             string  `json:"name"              validate:"required"`
	Relationship     string  `json:"relationship"      validate:"required"`
	SharePercentage  float64 `json:"share_percentage"  validate:"required,gt=0,lte=100"`
	DateOfBirth      string  `json:"date_of_birth"     validate:"omitempty"` // RFC3339 date
}

// SubmitBillingMethodChangeRequest — POST /policies/{pn}/requests/billing-method-change
// [FR-PM-001] [BR-PM-023]
type SubmitBillingMethodChangeRequest struct {
	GenericRequestPayload
	Payload struct {
		NewBillingMethod string `json:"new_billing_method" validate:"required,oneof=CASH PAY_RECOVERY ONLINE"`
		EmployerCode     string `json:"employer_code"      validate:"omitempty"` // Required when changing to PAY_RECOVERY
	} `json:"payload" validate:"required"`
}

// SubmitAssignmentRequest — POST /policies/{pn}/requests/assignment
// [FR-PM-001] [BR-PM-023]
type SubmitAssignmentRequest struct {
	GenericRequestPayload
	Payload struct {
		AssignmentType  string `json:"assignment_type"  validate:"required,oneof=ABSOLUTE CONDITIONAL"`
		AssigneeName    string `json:"assignee_name"    validate:"required"`
		AssigneeAddress string `json:"assignee_address" validate:"omitempty"`
	} `json:"payload" validate:"required"`
}

// SubmitAddressChangeRequest — POST /policies/{pn}/requests/address-change
// [FR-PM-001] [BR-PM-023]
type SubmitAddressChangeRequest struct {
	GenericRequestPayload
	Payload struct {
		NewAddress *AddressDetail `json:"new_address" validate:"omitempty"`
		NewName    string         `json:"new_name"    validate:"omitempty"` // If name change also requested
	} `json:"payload" validate:"omitempty"`
}

// AddressDetail represents an Indian address.
type AddressDetail struct {
	Line1   string `json:"line1"   validate:"omitempty"`
	Line2   string `json:"line2"   validate:"omitempty"`
	City    string `json:"city"    validate:"omitempty"`
	State   string `json:"state"   validate:"omitempty"`
	Pincode string `json:"pincode" validate:"omitempty,len=6"` // Pattern: ^\d{6}$
}

// SubmitPremiumRefundRequest — POST /policies/{pn}/requests/premium-refund
// [FR-PM-001] [BR-PM-023]
type SubmitPremiumRefundRequest struct {
	GenericRequestPayload
	Payload struct {
		ReceiptNumber string  `json:"receipt_number" validate:"required"`
		RefundAmount  float64 `json:"refund_amount"  validate:"required,gt=0"`
		RefundReason  string  `json:"refund_reason"  validate:"omitempty"`
	} `json:"payload" validate:"required"`
}

// SubmitDuplicateBondRequest — POST /policies/{pn}/requests/duplicate-bond
// [FR-PM-001] [BR-PM-023]
type SubmitDuplicateBondRequest struct {
	GenericRequestPayload
	// Uses GenericRequestPayload only
}

// ============================================================================
// STEP 4: Admin/System Endpoints (2 endpoints — no idempotency key required)
// ============================================================================

// AdminVoidPolicyRequest — POST /policies/{pn}/requests/admin-void
// Sends "admin-void" signal to plw-{pn}. No service_request record. [BR-PM-073]
type AdminVoidPolicyRequest struct {
	Reason   string `json:"reason"    validate:"required"`
	VoidedBy int64  `json:"voided_by" validate:"required,min=1"` // CPC admin BIGINT user ID [Gap-1]
}

// ReopenPolicyRequest — POST /policies/{pn}/requests/reopen
// Used AFTER terminal cooling expires to restart a workflow from snapshot state.
// Calls SignalWithStart with WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE.
type ReopenPolicyRequest struct {
	ReopenedBy int64  `json:"reopened_by" validate:"required,min=1"` // BIGINT user ID [Gap-1]
	Reason     string `json:"reason"      validate:"required"`
}

// ============================================================================
// STEP 5: Quote Requests (3 endpoints — synchronous, no Temporal workflow started)
// ============================================================================

// QuoteRequest — POST /policies/{pn}/quotes/surrender or /quotes/loan
// Handler → GetSurrenderQuoteActivity or GetLoanQuoteActivity (synchronous)
type QuoteRequest struct {
	AsOfDate string `json:"as_of_date" validate:"omitempty"` // RFC3339 date; defaults to today
}

// ConversionQuoteRequest — POST /policies/{pn}/quotes/conversion
// Handler → GetConversionQuoteActivity (synchronous)
type ConversionQuoteRequest struct {
	TargetProductCode string `json:"target_product_code" validate:"required"`
	AsOfDate          string `json:"as_of_date"          validate:"omitempty"`
}

// ============================================================================
// STEP 6: Request Lifecycle DTOs
// ============================================================================

// WithdrawRequestRequest — PUT /policies/{pn}/requests/{request_id}/withdraw
// Sends "withdrawal-request" signal to plw-{pn}; cancels downstream child workflow.
// [FR-PM-007] [BR-PM-090]
type WithdrawRequestRequest struct {
	Reason      string `json:"reason"       validate:"required"`
	WithdrawnBy *int64 `json:"withdrawn_by" validate:"omitempty,min=1"` // BIGINT user ID; nil for system/batch [Gap-1]
}

// ListRequestsParams — GET /policies/{pn}/requests (query params)
// [FR-PM-006]
type ListRequestsParams struct {
	port.MetadataRequest
	Status        string `form:"status"         validate:"omitempty"`
	RequestType   string `form:"request_type"   validate:"omitempty"`
	SourceChannel string `form:"source_channel" validate:"omitempty"`
	DateFrom      string `form:"date_from"      validate:"omitempty"`
	DateTo        string `form:"date_to"        validate:"omitempty"`
	SortBy        string `form:"sort_by"        validate:"omitempty,oneof=submitted_at completed_at request_type"`
}

// ListPendingRequestsParams — GET /requests/pending (query params)
// CPC inbox. [FR-PM-008]
type ListPendingRequestsParams struct {
	port.MetadataRequest
	Status      string `form:"status"       validate:"omitempty,oneof=RECEIVED ROUTED IN_PROGRESS"`
	RequestType string `form:"request_type" validate:"omitempty"`
	SortBy      string `form:"sort_by"      validate:"omitempty,oneof=submitted_at age_hours request_type"`
}

// BatchStatusRequest — GET /policies/batch-status (query params)
// [FR-PM-001]
type BatchStatusRequest struct {
	PolicyNumbers []string `form:"policy_numbers" validate:"required,min=1,max=100"`
}

// ============================================================================
// STEP 7: ToSignalPayload helpers
// Convert request DTOs to the map[string]interface{} payload stored in
// service_request.request_payload (JSONB) and sent as Temporal signal payload.
// ============================================================================

// ToSignalPayload converts SubmitSurrenderRequest to the signal payload map.
func (r SubmitSurrenderRequest) ToSignalPayload() map[string]interface{} {
	return map[string]interface{}{
		"source_channel":      r.SourceChannel,
		"submitted_by":        r.SubmittedBy,
		"notes":               r.Notes,
		"disbursement_method": r.Payload.DisbursementMethod,
		"bank_account_id":     r.Payload.BankAccountID,
		"reason":              r.Payload.Reason,
	}
}

// ============================================================================
// STEP 8: Validate() stubs — buildImproved Validator interface compliance
// Non-GET handlers require Validate() to be implemented on the req struct.
// Phase 6 (govalid) will replace these stubs with struct-tag-based validators.
// ============================================================================

// Validate is a stub — govalid will generate struct-tag validation in Phase 6.
func (r SubmitLoanRequest) Validate() error { return nil }

// Validate is a stub — govalid will generate struct-tag validation in Phase 6.
func (r SubmitLoanRepaymentRequest) Validate() error { return nil }

// Validate is a stub — govalid will generate struct-tag validation in Phase 6.
func (r SubmitRevivalRequest) Validate() error { return nil }

// Validate is a stub — govalid will generate struct-tag validation in Phase 6.
func (r SubmitDeathClaimRequest) Validate() error { return nil }

// Validate is a stub — govalid will generate struct-tag validation in Phase 6.
func (r SubmitCommutationRequest) Validate() error { return nil }

// Validate is a stub — govalid will generate struct-tag validation in Phase 6.
func (r SubmitConversionRequest) Validate() error { return nil }

// Validate is a stub — govalid will generate struct-tag validation in Phase 6.
func (r SubmitFreelookRequest) Validate() error { return nil }

// Validate is a stub — govalid will generate struct-tag validation in Phase 6.
func (r SubmitPaidUpRequest) Validate() error { return nil }

// Validate is a stub — govalid will generate struct-tag validation in Phase 6.
func (r SubmitNominationChangeRequest) Validate() error { return nil }

// Validate is a stub — govalid will generate struct-tag validation in Phase 6.
func (r SubmitAssignmentRequest) Validate() error { return nil }

// Validate is a stub — govalid will generate struct-tag validation in Phase 6.
func (r SubmitAddressChangeRequest) Validate() error { return nil }

// Validate is a stub — govalid will generate struct-tag validation in Phase 6.
func (r SubmitPremiumRefundRequest) Validate() error { return nil }

// Validate is a stub — govalid will generate struct-tag validation in Phase 6.
func (r SubmitDuplicateBondRequest) Validate() error { return nil }

// Validate is a stub — govalid will generate struct-tag validation in Phase 6.
func (r AdminVoidPolicyRequest) Validate() error { return nil }

// Validate is a stub — govalid will generate struct-tag validation in Phase 6.
func (r ReopenPolicyRequest) Validate() error { return nil }

// Validate is a stub — govalid will generate struct-tag validation in Phase 6.
func (r QuoteRequest) Validate() error { return nil }

// Validate is a stub — govalid will generate struct-tag validation in Phase 6.
func (r ConversionQuoteRequest) Validate() error { return nil }

// Validate is a stub — govalid will generate struct-tag validation in Phase 6.
func (r WithdrawRequestRequest) Validate() error { return nil }

// ============================================================================
// STEP 9: Cross-Field Validation (Gap-3)
// Struct-tag validators cannot express conditional rules.
// These Validate() methods are called by handlers AFTER struct-tag binding.
// ============================================================================

// Validate enforces Gap-3 rule: employer_code is required when
// new_billing_method is PAY_RECOVERY (post office payroll deduction).
func (r SubmitBillingMethodChangeRequest) Validate() error {
	if r.Payload.NewBillingMethod == "PAY_RECOVERY" && r.Payload.EmployerCode == "" {
		return fmt.Errorf("employer_code is required when new_billing_method is PAY_RECOVERY")
	}
	return nil
}

// Validate enforces Gap-3 rule: bank_account_id is required when
// disbursement_method is NEFT (electronic bank transfer).
func (r SubmitSurrenderRequest) Validate() error {
	if r.Payload.DisbursementMethod == "NEFT" && r.Payload.BankAccountID == 0 {
		return fmt.Errorf("bank_account_id is required when disbursement_method is NEFT")
	}
	return nil
}

// Validate enforces Gap-3 rule: bank_account_id is required when
// disbursement_method is NEFT.
func (r SubmitMaturityClaimRequest) Validate() error {
	if r.Payload.DisbursementMethod == "NEFT" && r.Payload.BankAccountID == 0 {
		return fmt.Errorf("bank_account_id is required when disbursement_method is NEFT")
	}
	return nil
}

// Validate enforces Gap-3 rule: bank_account_id is required when
// disbursement_method is NEFT.
func (r SubmitSurvivalBenefitRequest) Validate() error {
	if r.Payload.DisbursementMethod == "NEFT" && r.Payload.BankAccountID == 0 {
		return fmt.Errorf("bank_account_id is required when disbursement_method is NEFT")
	}
	return nil
}

// ToRequestCategory returns the request category for a given request type.
func ToRequestCategory(requestType string) string {
	switch requestType {
	case domain.RequestTypeSurrender,
		domain.RequestTypeLoan,
		domain.RequestTypeLoanRepayment,
		domain.RequestTypeRevival,
		domain.RequestTypeDeathClaim,
		domain.RequestTypeMaturityClaim,
		domain.RequestTypeSurvivalBenefit,
		domain.RequestTypeCommutation,
		domain.RequestTypeConversion,
		domain.RequestTypeFLC,
		domain.RequestTypePaidUp,
		domain.RequestTypeForcedSurrender:
		return domain.RequestCategoryFinancial
	case domain.RequestTypeNominationChange,
		domain.RequestTypeBillingMethodChange,
		domain.RequestTypeAssignment,
		domain.RequestTypeAddressChange,
		domain.RequestTypePremiumRefund,
		domain.RequestTypeDuplicateBond:
		return domain.RequestCategoryNonFinancial
	case domain.RequestTypeAdminVoid,
		domain.RequestTypeReopen:
		return domain.RequestCategoryAdmin
	default:
		return domain.RequestCategoryNonFinancial
	}
}
