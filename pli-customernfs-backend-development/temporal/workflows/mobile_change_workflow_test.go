package workflows

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMobileChangeWorkflowInput_JSONRoundTrip verifies marshal -> unmarshal preserves all fields.
func TestMobileChangeWorkflowInput_JSONRoundTrip(t *testing.T) {
	policyNum := "POL-12345"
	original := MobileChangeWorkflowInput{
		RequestID:    "req-mob-001",
		TicketNumber: "NFS-MOB-20260325-000001",
		CustomerID:   42,
		PolicyNumber: &policyNum,
		Channel:      "Portal",
		InitiatedBy:  "42",
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded MobileChangeWorkflowInput
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, original.RequestID, decoded.RequestID)
	assert.Equal(t, original.TicketNumber, decoded.TicketNumber)
	assert.Equal(t, original.CustomerID, decoded.CustomerID)
	require.NotNil(t, decoded.PolicyNumber)
	assert.Equal(t, "POL-12345", *decoded.PolicyNumber)
	assert.Equal(t, original.Channel, decoded.Channel)
	assert.Equal(t, original.InitiatedBy, decoded.InitiatedBy)
}

// TestMobileChangeWorkflowInput_FieldMapping verifies all fields are populated correctly.
func TestMobileChangeWorkflowInput_FieldMapping(t *testing.T) {
	policyNum := "POL-99999"
	input := MobileChangeWorkflowInput{
		RequestID:    "req-mob-002",
		TicketNumber: "NFS-MOB-20260325-000002",
		CustomerID:   100200300,
		PolicyNumber: &policyNum,
		Channel:      "Mobile",
		InitiatedBy:  "100200300",
	}

	assert.Equal(t, "req-mob-002", input.RequestID)
	assert.Equal(t, "NFS-MOB-20260325-000002", input.TicketNumber)
	assert.Equal(t, int64(100200300), input.CustomerID)
	require.NotNil(t, input.PolicyNumber)
	assert.Equal(t, "POL-99999", *input.PolicyNumber)
	assert.Equal(t, "Mobile", input.Channel)
	assert.Equal(t, "100200300", input.InitiatedBy)
}

// TestMobileChangeWorkflowInput_NilPolicyNumber verifies optional PolicyNumber field.
func TestMobileChangeWorkflowInput_NilPolicyNumber(t *testing.T) {
	input := MobileChangeWorkflowInput{
		RequestID:    "req-mob-003",
		TicketNumber: "NFS-MOB-20260325-000003",
		CustomerID:   555,
		PolicyNumber: nil,
		Channel:      "CallCenter",
		InitiatedBy:  "555",
	}

	assert.Nil(t, input.PolicyNumber)

	data, err := json.Marshal(input)
	require.NoError(t, err)

	var decoded MobileChangeWorkflowInput
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Nil(t, decoded.PolicyNumber, "nil PolicyNumber should remain nil after round-trip")
	assert.Equal(t, "req-mob-003", decoded.RequestID)
}
