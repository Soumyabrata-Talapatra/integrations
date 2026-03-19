package response

import "policy-issue-service/core/port"

// --- VAL-POL-001: Eligibility Check ---

// EligibilityCheck represents a single eligibility check result
type EligibilityCheck struct {
	Check   string `json:"check"`
	Passed  bool   `json:"passed"`
	Message string `json:"message,omitempty"`
}

// AggregateSAInfo represents aggregate sum assured details
type AggregateSAInfo struct {
	Current     float64 `json:"current"`
	Proposed    float64 `json:"proposed"`
	MaxAllowed  float64 `json:"max_allowed"`
	WithinLimit bool    `json:"within_limit"`
}

// EligibilityCheckResponse represents the response for eligibility validation
// [VAL-POL-001] Real-time eligibility check
// Components: BR-POL-011, BR-POL-012
type EligibilityCheckResponse struct {
	port.StatusCodeAndMessage
	IsEligible       bool               `json:"is_eligible"`
	AgeAtEntry       int                `json:"age_at_entry"`
	EligibilityChecks []EligibilityCheck `json:"eligibility_checks"`
	AggregateSA      *AggregateSAInfo   `json:"aggregate_sa,omitempty"`
}

// --- VAL-POL-002, VAL-POL-003: Format Validation ---

// FormatValidationResponse represents the response for format validation APIs
// [VAL-POL-002] Aadhaar format validation (VR-PI-008)
// [VAL-POL-003] PAN format validation (VR-PI-009)
type FormatValidationResponse struct {
	port.StatusCodeAndMessage
	IsValid bool   `json:"is_valid"`
	Message string `json:"message"`
}

// --- VAL-POL-004: Pincode Validation ---

// PincodeValidationResponse represents the response for pincode validation
// [VAL-POL-004] Pincode-state validation (VR-PI-034, VR-PI-035)
type PincodeValidationResponse struct {
	port.StatusCodeAndMessage
	IsValid    bool   `json:"is_valid"`
	Pincode    string `json:"pincode"`
	State      string `json:"state,omitempty"`
	District   string `json:"district,omitempty"`
	PostOffice string `json:"post_office,omitempty"`
	Message    string `json:"message"`
}

// --- VAL-POL-005: IFSC Validation ---

// IFSCValidationResponse represents the response for IFSC validation
// [VAL-POL-005] IFSC code validation (VR-PI-030)
type IFSCValidationResponse struct {
	port.StatusCodeAndMessage
	IsValid    bool   `json:"is_valid"`
	IFSCCode   string `json:"ifsc_code"`
	BankName   string `json:"bank_name,omitempty"`
	BranchName string `json:"branch_name,omitempty"`
	Address    string `json:"address,omitempty"`
}

// --- VAL-POL-006: Nominee Shares Validation ---

// ShareValidationResponse represents the response for nominee share validation
// [VAL-POL-006] Nominee share % validation (VR-PI-018)
type ShareValidationResponse struct {
	port.StatusCodeAndMessage
	IsValid         bool    `json:"is_valid"`
	TotalPercentage float64 `json:"total_percentage"`
	Message         string  `json:"message"`
}

// --- VAL-POL-007: Date Chain Validation ---

// DateChainError represents a single date chain validation error
type DateChainError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// DateChainValidationResponse represents the response for date chain validation
// [VAL-POL-007] Date chain validation (BR-POL-018)
type DateChainValidationResponse struct {
	port.StatusCodeAndMessage
	IsValid bool             `json:"is_valid"`
	Errors  []DateChainError `json:"errors,omitempty"`
}

// --- VAL-POL-008: Aggregate SA Check ---

// AggregateSACheckResponse represents the response for aggregate SA validation
// [VAL-POL-008] Aggregate SA check (INT-POL-002)
type AggregateSACheckResponse struct {
	port.StatusCodeAndMessage
	IsEligible         bool    `json:"is_eligible"`
	CurrentAggregateSA float64 `json:"current_aggregate_sa"`
	ProposedAggregateSA float64 `json:"proposed_aggregate_sa"`
	MaxAllowedSA       float64 `json:"max_allowed_sa"`
	Reason             string  `json:"reason,omitempty"`
}
