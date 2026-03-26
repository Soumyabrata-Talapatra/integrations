package response

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"customer-nfs-service/core/domain"
)

// TestNewMobileChangeInitiateResponse_ConstructsCorrectly verifies response construction from ServiceRequest.
func TestNewMobileChangeInitiateResponse_ConstructsCorrectly(t *testing.T) {
	sr := &domain.ServiceRequest{
		RequestID:    "req-mob-resp-001",
		TicketNumber: "NFS-MOB-20260325-000001",
		Status:       "CREATED",
	}

	resp := NewMobileChangeInitiateResponse(sr)

	require.NotNil(t, resp)
	assert.Equal(t, "req-mob-resp-001", resp.Data.RequestID)
	assert.Equal(t, "NFS-MOB-20260325-000001", resp.Data.TicketNumber)
	assert.Equal(t, "CREATED", resp.Data.Status)
	assert.True(t, resp.Data.OTPSent, "OTPSent must be true for mobile change initiation")
}

// TestNewMobileChangeInitiateResponse_StatusCodeAndMessage verifies embedded status fields.
func TestNewMobileChangeInitiateResponse_StatusCodeAndMessage(t *testing.T) {
	sr := &domain.ServiceRequest{
		RequestID:    "req-mob-resp-002",
		TicketNumber: "NFS-MOB-20260325-000002",
		Status:       "CREATED",
	}

	resp := NewMobileChangeInitiateResponse(sr)

	require.NotNil(t, resp)
	assert.Equal(t, 201, resp.StatusCode)
	assert.True(t, resp.Success)
	assert.Equal(t, "resource created successfully", resp.Message)
}

// TestNewEmailChangeInitiateResponse_ConstructsCorrectly verifies response construction from ServiceRequest.
func TestNewEmailChangeInitiateResponse_ConstructsCorrectly(t *testing.T) {
	sr := &domain.ServiceRequest{
		RequestID:    "req-email-resp-001",
		TicketNumber: "NFS-EMAIL-20260325-000001",
		Status:       "CREATED",
	}

	resp := NewEmailChangeInitiateResponse(sr)

	require.NotNil(t, resp)
	assert.Equal(t, "req-email-resp-001", resp.Data.RequestID)
	assert.Equal(t, "NFS-EMAIL-20260325-000001", resp.Data.TicketNumber)
	assert.Equal(t, "CREATED", resp.Data.Status)
	assert.True(t, resp.Data.OTPSent, "OTPSent must be true for email change initiation")
}

// TestNewEmailChangeInitiateResponse_StatusCodeAndMessage verifies embedded status fields.
func TestNewEmailChangeInitiateResponse_StatusCodeAndMessage(t *testing.T) {
	sr := &domain.ServiceRequest{
		RequestID:    "req-email-resp-002",
		TicketNumber: "NFS-EMAIL-20260325-000002",
		Status:       "CREATED",
	}

	resp := NewEmailChangeInitiateResponse(sr)

	require.NotNil(t, resp)
	assert.Equal(t, 201, resp.StatusCode)
	assert.True(t, resp.Success)
	assert.Equal(t, "resource created successfully", resp.Message)
}
