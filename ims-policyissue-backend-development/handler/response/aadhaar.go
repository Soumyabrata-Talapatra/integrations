package response

import (
	"policy-issue-service/core/port"
)

// AadhaarInitiateResponse represents the response for Initiate Aadhaar Auth API
type AadhaarInitiateResponse struct {
	port.StatusCodeAndMessage
	TransactionID string `json:"transaction_id"`
	SessionID     string `json:"session_id"`
}

// AadhaarVerifyOTPResponse represents the response for Verify Aadhaar OTP API
type AadhaarVerifyOTPResponse struct {
	port.StatusCodeAndMessage
	SessionID string `json:"session_id"`
	Status    string `json:"status"` // success, failed, already_verified
}

// AadhaarSubmitResponse represents the response for Submit Aadhaar Proposal API
type AadhaarSubmitResponse struct {
	port.StatusCodeAndMessage
	ProposalID   int64  `json:"proposal_id"`
	WorkflowID   string `json:"workflow_id"`
	RunID        string `json:"run_id"`
	IssuanceType string `json:"issuance_type"` // INSTANT or STANDARD
}
