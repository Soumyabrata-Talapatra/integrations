// Package response contains HTTP response DTOs for name change endpoints.
//
// Phase 2: Name Change (CORE-005..010)
// Mirrors handler/response/address_change.go pattern.
package response

import (
	"customer-nfs-service/core/port"
	"time"
)

// ─── CORE-005: POST /nfs/name-change/initiate ─────────────────────────────────

// NameChangeInitiateData is the payload returned on a successful CORE-005 call.
// Workflow started: WF-NFS-003 (AADHAAR) or WF-NFS-004 (MANUAL).
type NameChangeInitiateData struct {
	RequestID    string `json:"request_id"`    // ← add
	TicketNumber string `json:"ticket_number"` // ← add
	// Status reflects the initial state: always "CREATED" at this stage.
	Status string `json:"status"`
	// AuthMethod echoes the auth method used so clients can branch their UX.
	AuthMethod string `json:"auth_method"`
	// OTPSent is true when the Aadhaar OTP has been dispatched (AADHAAR path only).
	OTPSent bool `json:"otp_sent,omitempty"`
	// SLADeadline is the ISO-8601 date by which the request must be processed (Manual path).
	SLADeadline *time.Time `json:"sla_deadline,omitempty"`
	// Message provides a human-readable outcome.
	Message string `json:"message"`
}

// NameChangeInitiateResponse is the top-level response for CORE-005.
//
//	type NameChangeInitiateResponse struct {
//		port.StatusCodeAndMessage `json:",inline"`
//		Data                      NameChangeInitiateData `json:"data"`
//	}
type NameChangeInitiateResponse struct {
	port.StatusCodeAndMessage
	Data NameChangeInitiateData `json:"data"`
}

// NewNameChangeInitiateResponse constructs the CORE-005 success response.
//
//	func NewNameChangeInitiateResponse(authMethod, status string, otpSent bool, slaDeadline *time.Time) *NameChangeInitiateResponse {
//		msg := "Name change request created successfully. Please verify your Aadhaar OTP."
//		if authMethod == "MANUAL" {
//			msg = "Name change request created. Please upload supporting documents."
//			otpSent = false
//		}
//		return &NameChangeInitiateResponse{
//			StatusCodeAndMessage: port.CreateSuccess,
//			Data: NameChangeInitiateData{
//				Status:      status,
//				AuthMethod:  authMethod,
//				OTPSent:     otpSent,
//				SLADeadline: slaDeadline,
//				Message:     msg,
//			},
//		}
//	}
func NewNameChangeInitiateResponse(requestID, ticketNumber, authMethod, status string, otpSent bool, slaDeadline *time.Time) *NameChangeInitiateResponse {
	msg := "Name change request created successfully. Please verify your Aadhaar OTP."
	if authMethod == "MANUAL" {
		msg = "Name change request created. Please upload supporting documents."
		otpSent = false
	}
	return &NameChangeInitiateResponse{
		StatusCodeAndMessage: port.CreateSuccess,
		Data: NameChangeInitiateData{
			RequestID:    requestID,
			TicketNumber: ticketNumber,
			Status:       status,
			AuthMethod:   authMethod,
			OTPSent:      otpSent,
			SLADeadline:  slaDeadline,
			Message:      msg,
		},
	}
}

// ─── CORE-006: POST /nfs/name-change/:request_id/verify-otp ──────────────────

// NameVerifyOTPData is the payload returned on a successful CORE-006 call.
// BR-NFS-007: Successful OTP verification → status moves to COMPLETED (Aadhaar path).
// BR-NFS-009: customer.name.updated event published by workflow on COMPLETED.
type NameVerifyOTPData struct {
	// RequestID identifies the request that was just completed.
	RequestID string `json:"request_id"`
	// Status will be "COMPLETED" on success (BR-NFS-007).
	Status string `json:"status"`
	// VerifiedAt is the ISO-8601 timestamp of OTP verification.
	VerifiedAt string `json:"verified_at"`
	// Message provides a human-readable outcome.
	Message string `json:"message"`
}

// NameVerifyOTPResponse is the top-level response for CORE-006.
type NameVerifyOTPResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	Data                      NameVerifyOTPData `json:"data"`
}

// NewNameVerifyOTPResponse constructs the CORE-006 success response.
func NewNameVerifyOTPResponse(requestID, status string) *NameVerifyOTPResponse {
	return &NameVerifyOTPResponse{
		StatusCodeAndMessage: port.OTPSuccess,
		Data: NameVerifyOTPData{
			RequestID:  requestID,
			Status:     status,
			VerifiedAt: time.Now().UTC().Format(time.RFC3339),
			Message:    "Aadhaar OTP verified. Name change will be processed shortly.",
		},
	}
}

// ─── CORE-007: POST /nfs/name-change/:request_id/submit ──────────────────────

// NameSubmitData is the payload returned on a successful CORE-007 call.
// BR-NFS-008: Document submission triggers workflow to advance to CPC review.
// Documents signal sent → status transitions to PENDING_APPROVAL.
type NameSubmitData struct {
	// RequestID identifies the submitted request.
	RequestID string `json:"request_id"`
	// Status will be "PENDING_APPROVAL".
	Status string `json:"status"`
	// DocumentsReceived is the count of uploaded documents.
	DocumentsReceived int `json:"documents_received"`
	// SubmittedAt is the ISO-8601 timestamp of submission.
	SubmittedAt string `json:"submitted_at"`
	// Message provides a human-readable outcome.
	Message string `json:"message"`
}

// NameSubmitResponse is the top-level response for CORE-007.
type NameSubmitResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	Data                      NameSubmitData `json:"data"`
}

// NewNameSubmitResponse constructs the CORE-007 success response.
func NewNameSubmitResponse(requestID, status string, docCount int) *NameSubmitResponse {
	return &NameSubmitResponse{
		StatusCodeAndMessage: port.UpdateSuccess,
		Data: NameSubmitData{
			RequestID:         requestID,
			Status:            status,
			DocumentsReceived: docCount,
			SubmittedAt:       time.Now().UTC().Format(time.RFC3339),
			Message:           "Documents submitted. Your request has been forwarded for CPC review.",
		},
	}
}

// ─── CORE-008: POST /nfs/name-change/:request_id/approve ─────────────────────

// NameApproveData is the payload returned on a successful CORE-008 call.
// BR-NFS-009: On APPROVE, customer.name.updated event published + all policies updated.
// BR-NFS-012: Decision triggers status transition in running WF-NFS-004.
type NameApproveData struct {
	// RequestID identifies the request the decision was made on.
	RequestID string `json:"request_id"`
	// Decision echoes the CPC decision (APPROVE / REJECT / SEND_BACK).
	Decision string `json:"decision"`
	// NewStatus reflects the resulting state: COMPLETED / REJECTED / PENDING_DOCUMENTS.
	NewStatus string `json:"new_status"`
	// PoliciesAffected is the count of policies updated on APPROVE (BR-NFS-009).
	PoliciesAffected int `json:"policies_affected,omitempty"`
	// Message provides a human-readable outcome.
	Message string `json:"message"`
}

// NameApproveResponse is the top-level response for CORE-008.
type NameApproveResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	Data                      NameApproveData `json:"data"`
}

// NewNameApproveResponse constructs the CORE-008 success response.
func NewNameApproveResponse(requestID, decision, newStatus string, policiesAffected int) *NameApproveResponse {
	var msg string
	switch decision {
	case "APPROVE":
		msg = "Name change approved. Customer profile and all associated policies will be updated."
	case "REJECT":
		msg = "Name change rejected."
	case "SEND_BACK":
		msg = "Request sent back for additional documents."
	default:
		msg = "Decision processed."
	}
	return &NameApproveResponse{
		StatusCodeAndMessage: port.UpdateSuccess,
		Data: NameApproveData{
			RequestID:        requestID,
			Decision:         decision,
			NewStatus:        newStatus,
			PoliciesAffected: policiesAffected,
			Message:          msg,
		},
	}
}

// ─── CORE-009: POST /nfs/requests/:request_id/withdraw ───────────────────────

// WithdrawData is the payload returned on a successful CORE-009 call.
// BR-NFS-013: Withdrawal eligibility confirmed before workflow starts.
// partial_processing_flag=FALSE → auto-WITHDRAWN immediately.
// partial_processing_flag=TRUE  → PENDING_WITHDRAWAL_APPROVAL (CPC manual).
type WithdrawData struct {
	// RequestID identifies the withdrawn request.
	RequestID string `json:"request_id"`
	// Status is WITHDRAWN or PENDING_WITHDRAWAL_APPROVAL.
	Status string `json:"status"`
	// WithdrawalID is the UUID of the withdrawal record (if created).
	WithdrawalID string `json:"withdrawal_id,omitempty"`
	// Message provides a human-readable outcome.
	Message string `json:"message"`
}

// WithdrawResponse is the top-level response for CORE-009.
type WithdrawResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	Data                      WithdrawData `json:"data"`
}

// NewWithdrawResponse constructs the CORE-009 success response.
func NewWithdrawResponse(requestID, status, withdrawalID string) *WithdrawResponse {
	var msg string
	switch status {
	case "WITHDRAWN":
		msg = "Your request has been withdrawn successfully."
	case "PENDING_WITHDRAWAL_APPROVAL":
		msg = "Withdrawal request submitted. CPC approval is required due to partial processing."
	default:
		msg = "Withdrawal request submitted."
	}
	return &WithdrawResponse{
		StatusCodeAndMessage: port.UpdateSuccess,
		Data: WithdrawData{
			RequestID:    requestID,
			Status:       status,
			WithdrawalID: withdrawalID,
			Message:      msg,
		},
	}
}

// ─── CORE-010: POST /nfs/requests/:request_id/request-documents ──────────────

// RequestMissingDocumentsData is the payload returned on a successful CORE-010 call.
// BR-NFS-014: Secure link generated for customer (7-day expiry, max 3 attempts).
// BR-NFS-014: After 3rd link generation → status DOCUMENTS_EXPIRED.
type RequestMissingDocumentsData struct {
	// RequestID identifies the request.
	RequestID string `json:"request_id"`
	// Status will be PENDING_DOCUMENTS.
	Status string `json:"status"`
	// SecureUploadToken is the token component of the secure upload link.
	SecureUploadToken string `json:"secure_upload_token,omitempty"`
	// SecureUploadURL is the full URL sent to the customer.
	SecureUploadURL string `json:"secure_upload_url,omitempty"`
	// ExpiresAt is the ISO-8601 expiry of the secure link (7 days).
	ExpiresAt string `json:"expires_at,omitempty"`
	// AttemptNumber tracks the link generation attempt (1, 2, or 3 per BR-NFS-014).
	AttemptNumber int `json:"attempt_number"`
	// MissingDocuments is the list of document types that must be uploaded.
	MissingDocuments []string `json:"missing_documents"`
	// Message provides a human-readable outcome.
	Message string `json:"message"`
}

// RequestMissingDocumentsResponse is the top-level response for CORE-010.
type RequestMissingDocumentsResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	Data                      RequestMissingDocumentsData `json:"data"`
}

// NewRequestMissingDocumentsResponse constructs the CORE-010 success response.
func NewRequestMissingDocumentsResponse(requestID, status, token, url, expiresAt string, attempt int, docs []string) *RequestMissingDocumentsResponse {
	return &RequestMissingDocumentsResponse{
		StatusCodeAndMessage: port.UpdateSuccess,
		Data: RequestMissingDocumentsData{
			RequestID:         requestID,
			Status:            status,
			SecureUploadToken: token,
			SecureUploadURL:   url,
			ExpiresAt:         expiresAt,
			AttemptNumber:     attempt,
			MissingDocuments:  docs,
			Message:           "Secure upload link generated. Customer has been notified.",
		},
	}
}
