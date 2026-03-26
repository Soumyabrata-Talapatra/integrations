package response

import (
	"customer-nfs-service/core/domain"
	"customer-nfs-service/core/port"
)

// MobileChangeInitiateData is the payload for mobile change initiation response.
type MobileChangeInitiateData struct {
	RequestID    string `json:"request_id"`
	TicketNumber string `json:"ticket_number"`
	Status       string `json:"status"`
	OTPSent      bool   `json:"otp_sent"`
}

// MobileChangeInitiateResponse is the top-level response for mobile change initiation.
type MobileChangeInitiateResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	Data                      MobileChangeInitiateData `json:"data"`
}

// NewMobileChangeInitiateResponse constructs the mobile change initiation response.
func NewMobileChangeInitiateResponse(sr *domain.ServiceRequest) *MobileChangeInitiateResponse {
	return &MobileChangeInitiateResponse{
		StatusCodeAndMessage: port.CreateSuccess,
		Data: MobileChangeInitiateData{
			RequestID:    sr.RequestID,
			TicketNumber: sr.TicketNumber,
			Status:       sr.Status,
			OTPSent:      true,
		},
	}
}
