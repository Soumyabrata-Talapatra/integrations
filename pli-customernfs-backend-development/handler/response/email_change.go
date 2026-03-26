package response

import (
	"customer-nfs-service/core/domain"
	"customer-nfs-service/core/port"
)

// EmailChangeInitiateData is the payload for email change initiation response.
type EmailChangeInitiateData struct {
	RequestID    string `json:"request_id"`
	TicketNumber string `json:"ticket_number"`
	Status       string `json:"status"`
	OTPSent      bool   `json:"otp_sent"`
}

// EmailChangeInitiateResponse is the top-level response for email change initiation.
type EmailChangeInitiateResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	Data                      EmailChangeInitiateData `json:"data"`
}

// NewEmailChangeInitiateResponse constructs the email change initiation response.
func NewEmailChangeInitiateResponse(sr *domain.ServiceRequest) *EmailChangeInitiateResponse {
	return &EmailChangeInitiateResponse{
		StatusCodeAndMessage: port.CreateSuccess,
		Data: EmailChangeInitiateData{
			RequestID:    sr.RequestID,
			TicketNumber: sr.TicketNumber,
			Status:       sr.Status,
			OTPSent:      true,
		},
	}
}
