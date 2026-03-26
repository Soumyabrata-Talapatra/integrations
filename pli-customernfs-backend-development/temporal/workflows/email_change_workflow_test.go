package workflows

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEmailChangeWorkflowInput_JSONRoundTrip verifies marshal -> unmarshal preserves all fields.
func TestEmailChangeWorkflowInput_JSONRoundTrip(t *testing.T) {
	policyNum := "POL-EMAIL-001"
	original := EmailChangeWorkflowInput{
		RequestID:    "req-email-001",
		TicketNumber: "NFS-EMAIL-20260325-000001",
		CustomerID:   7777,
		PolicyNumber: &policyNum,
		Channel:      "Portal",
		InitiatedBy:  "7777",
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded EmailChangeWorkflowInput
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, original.RequestID, decoded.RequestID)
	assert.Equal(t, original.TicketNumber, decoded.TicketNumber)
	assert.Equal(t, original.CustomerID, decoded.CustomerID)
	require.NotNil(t, decoded.PolicyNumber)
	assert.Equal(t, "POL-EMAIL-001", *decoded.PolicyNumber)
	assert.Equal(t, original.Channel, decoded.Channel)
	assert.Equal(t, original.InitiatedBy, decoded.InitiatedBy)
}

// TestEmailChangeWorkflowInput_FieldMapping verifies all fields are populated correctly.
func TestEmailChangeWorkflowInput_FieldMapping(t *testing.T) {
	policyNum := "POL-EMAIL-999"
	input := EmailChangeWorkflowInput{
		RequestID:    "req-email-002",
		TicketNumber: "NFS-EMAIL-20260325-000002",
		CustomerID:   88888888,
		PolicyNumber: &policyNum,
		Channel:      "AgentPortal",
		InitiatedBy:  "88888888",
	}

	assert.Equal(t, "req-email-002", input.RequestID)
	assert.Equal(t, "NFS-EMAIL-20260325-000002", input.TicketNumber)
	assert.Equal(t, int64(88888888), input.CustomerID)
	require.NotNil(t, input.PolicyNumber)
	assert.Equal(t, "POL-EMAIL-999", *input.PolicyNumber)
	assert.Equal(t, "AgentPortal", input.Channel)
	assert.Equal(t, "88888888", input.InitiatedBy)
}

// TestEmailChangeWorkflowInput_NilPolicyNumber verifies optional PolicyNumber field.
func TestEmailChangeWorkflowInput_NilPolicyNumber(t *testing.T) {
	input := EmailChangeWorkflowInput{
		RequestID:    "req-email-003",
		TicketNumber: "NFS-EMAIL-20260325-000003",
		CustomerID:   123,
		PolicyNumber: nil,
		Channel:      "PostOffice",
		InitiatedBy:  "123",
	}

	assert.Nil(t, input.PolicyNumber)

	data, err := json.Marshal(input)
	require.NoError(t, err)

	var decoded EmailChangeWorkflowInput
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Nil(t, decoded.PolicyNumber, "nil PolicyNumber should remain nil after round-trip")
	assert.Equal(t, "req-email-003", decoded.RequestID)
}
