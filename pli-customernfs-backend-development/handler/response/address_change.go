// Package response defines HTTP response DTOs for the Address Change core journey.
// Phase: Phase 1 — Address Change
// APIs: CORE-001, CORE-002, CORE-003, CORE-004
package response

import (
	"customer-nfs-service/core/domain"
	"customer-nfs-service/core/port"
)

// ─── CORE-001: Initiate Address Change ───────────────────────────────────────

// AddressInitiateData is the payload returned on a successful CORE-001 call.
// FR-NFS-001 (Aadhaar path), FR-NFS-002 (Manual path), BR-NFS-011 (ticket number).
type AddressInitiateData struct {
	// RequestID is the UUID of the newly created service request (E-1).
	RequestID string `json:"request_id"`

	// TicketNumber is the human-readable ticket number (BR-NFS-011: NFS-ANC-YYYYMMDD-SEQNO).
	TicketNumber string `json:"ticket_number"`

	// Status reflects the initial state: always "CREATED" at this stage.
	Status string `json:"status"`

	// AuthMethod echoes the auth method used so clients can branch their UX.
	AuthMethod string `json:"auth_method"`

	// OTPSent is true when the Aadhaar OTP has been dispatched (AADHAAR path only).
	// FR-NFS-001: client should redirect to OTP entry screen when true.
	OTPSent bool `json:"otp_sent,omitempty"`

	// SLADeadline is the ISO-8601 date by which the request must be processed (Manual path).
	// BR-NFS-002: 15 days Regional / 30 days Circle HQ.
	SLADeadline *string `json:"sla_deadline,omitempty"`
}

// AddressChangeInitiateResponse is the top-level response for CORE-001.
type AddressChangeInitiateResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	Data                      AddressInitiateData `json:"data"`
}

// NewAddressChangeInitiateResponse builds the CORE-001 response from the created service request.
func NewAddressChangeInitiateResponse(sr *domain.ServiceRequest, slaDeadline *string, otpSent bool) *AddressChangeInitiateResponse {
	data := AddressInitiateData{
		RequestID:    sr.RequestID,
		TicketNumber: sr.TicketNumber,
		Status:       string(sr.Status),
		AuthMethod:   sr.AuthMethod,
		OTPSent:      otpSent,
		SLADeadline:  slaDeadline,
	}
	return &AddressChangeInitiateResponse{
		StatusCodeAndMessage: port.CreateSuccess,
		Data:                 data,
	}
}

// ─── CORE-002: Verify OTP ────────────────────────────────────────────────────

// AddressVerifyOTPData is the payload returned on a successful CORE-002 call.
// BR-NFS-001: OTP verified → status immediately COMPLETED.
// WF-NFS-001: workflow signal "otp_submitted" acknowledged.
type AddressVerifyOTPData struct {
	// RequestID identifies the request that was just completed.
	RequestID string `json:"request_id"`

	// TicketNumber is the human-readable reference.
	TicketNumber string `json:"ticket_number"`

	// Status will be "COMPLETED" on success or "FAILED" on OTP error.
	Status string `json:"status"`

	// Message provides a human-readable outcome.
	Message string `json:"message"`
}

// AddressChangeVerifyOTPResponse is the top-level response for CORE-002.
type AddressChangeVerifyOTPResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	Data                      AddressVerifyOTPData `json:"data"`
}

// NewAddressVerifyOTPResponse builds the CORE-002 response.
func NewAddressVerifyOTPResponse(requestID, ticketNumber, status, message string) *AddressChangeVerifyOTPResponse {
	return &AddressChangeVerifyOTPResponse{
		StatusCodeAndMessage: port.OTPSuccess,
		Data: AddressVerifyOTPData{
			RequestID:    requestID,
			TicketNumber: ticketNumber,
			Status:       status,
			Message:      message,
		},
	}
}

// ─── CORE-003: Submit Manual Address Change ───────────────────────────────────

// AddressSubmitData is the payload returned on a successful CORE-003 call.
// BR-NFS-002: status moved to PENDING_APPROVAL; SLA clock starts.
// WF-NFS-002: workflow transitions to wait for approval_decision signal.
type AddressSubmitData struct {
	// RequestID identifies the submitted request.
	RequestID string `json:"request_id"`

	// TicketNumber is the human-readable reference.
	TicketNumber string `json:"ticket_number"`

	// Status will be "PENDING_APPROVAL".
	Status string `json:"status"`

	// SubmittedAt is the ISO-8601 timestamp of submission.
	SubmittedAt string `json:"submitted_at"`

	// SLADeadline is the date by which CPC must process the request (BR-NFS-002).
	SLADeadline string `json:"sla_deadline"`
}

// AddressChangeSubmitResponse is the top-level response for CORE-003.
type AddressChangeSubmitResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	Data                      AddressSubmitData `json:"data"`
}

// NewAddressSubmitResponse builds the CORE-003 response.
func NewAddressSubmitResponse(requestID, ticketNumber, submittedAt, slaDeadline string) *AddressChangeSubmitResponse {
	return &AddressChangeSubmitResponse{
		StatusCodeAndMessage: port.UpdateSuccess,
		Data: AddressSubmitData{
			RequestID:    requestID,
			TicketNumber: ticketNumber,
			Status:       "PENDING_APPROVAL",
			SubmittedAt:  submittedAt,
			SLADeadline:  slaDeadline,
		},
	}
}

// ─── CORE-004: Approve / Reject Address Change ───────────────────────────────

// AddressApproveData is the payload returned on a successful CORE-004 call.
// BR-NFS-002: CPC decision recorded; WF-NFS-002: approval_decision signal sent.
// BR-NFS-012: state transitions depend on decision.
type AddressApproveData struct {
	// RequestID identifies the request the decision was made on.
	RequestID string `json:"request_id"`

	// TicketNumber is the human-readable reference.
	TicketNumber string `json:"ticket_number"`

	// Status reflects the new state: COMPLETED / REJECTED / PENDING_DOCUMENTS.
	Status string `json:"status"`

	// Decision echoes the CPC decision (APPROVE / REJECT / SEND_BACK).
	Decision string `json:"decision"`

	// ProcessedAt is the ISO-8601 timestamp of the CPC action.
	ProcessedAt string `json:"processed_at"`
}

// AddressChangeApproveResponse is the top-level response for CORE-004.
type AddressChangeApproveResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	Data                      AddressApproveData `json:"data"`
}

// NewAddressApproveResponse builds the CORE-004 response.
func NewAddressApproveResponse(requestID, ticketNumber, status, decision, processedAt string) *AddressChangeApproveResponse {
	return &AddressChangeApproveResponse{
		StatusCodeAndMessage: port.UpdateSuccess,
		Data: AddressApproveData{
			RequestID:    requestID,
			TicketNumber: ticketNumber,
			Status:       status,
			Decision:     decision,
			ProcessedAt:  processedAt,
		},
	}
}
