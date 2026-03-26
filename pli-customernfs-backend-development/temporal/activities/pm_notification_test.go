package activities

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNotifyPMInput_JSONRoundTrip verifies marshal -> unmarshal preserves all fields.
func TestNotifyPMInput_JSONRoundTrip(t *testing.T) {
	payload := json.RawMessage(`{"address":"updated"}`)
	original := NotifyPMInput{
		RequestID:     "req-100",
		CustomerID:    12345,
		RequestType:   "ADDRESS_CHANGE",
		Outcome:       "APPROVED",
		ChangePayload: payload,
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded NotifyPMInput
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, original.RequestID, decoded.RequestID)
	assert.Equal(t, original.CustomerID, decoded.CustomerID)
	assert.Equal(t, original.RequestType, decoded.RequestType)
	assert.Equal(t, original.Outcome, decoded.Outcome)
	assert.JSONEq(t, string(original.ChangePayload), string(decoded.ChangePayload))
}

// TestNotifyPMResult_JSONRoundTrip verifies marshal -> unmarshal preserves all fields.
func TestNotifyPMResult_JSONRoundTrip(t *testing.T) {
	original := NotifyPMResult{
		PoliciesNotified: 3,
		Success:          true,
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded NotifyPMResult
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, 3, decoded.PoliciesNotified)
	assert.True(t, decoded.Success)
}

// TestNotifyPMInput_NilChangePayload verifies serialization with nil ChangePayload.
func TestNotifyPMInput_NilChangePayload(t *testing.T) {
	input := NotifyPMInput{
		RequestID:     "req-200",
		CustomerID:    999,
		RequestType:   "MOBILE_CHANGE",
		Outcome:       "REJECTED",
		ChangePayload: nil,
	}

	data, err := json.Marshal(input)
	require.NoError(t, err)

	var decoded NotifyPMInput
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, "req-200", decoded.RequestID)
	assert.Nil(t, decoded.ChangePayload, "nil ChangePayload should remain nil after round-trip")
}

// TestNotifyPMInput_PopulatedChangePayload verifies round-trip with populated ChangePayload.
func TestNotifyPMInput_PopulatedChangePayload(t *testing.T) {
	payload := json.RawMessage(`{"email":"new@example.com","old_email":"old@example.com"}`)
	input := NotifyPMInput{
		RequestID:     "req-300",
		CustomerID:    55555,
		RequestType:   "EMAIL_CHANGE",
		Outcome:       "APPROVED",
		ChangePayload: payload,
	}

	data, err := json.Marshal(input)
	require.NoError(t, err)

	var decoded NotifyPMInput
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, "EMAIL_CHANGE", decoded.RequestType)
	assert.JSONEq(t, `{"email":"new@example.com","old_email":"old@example.com"}`, string(decoded.ChangePayload))
}

// TestNotifyPMInput_CustomerIDIsInt64 verifies CustomerID can hold large int64 values.
func TestNotifyPMInput_CustomerIDIsInt64(t *testing.T) {
	var largeID int64 = 9999999999
	input := NotifyPMInput{
		RequestID:   "req-400",
		CustomerID:  largeID,
		RequestType: "NAME_CHANGE",
		Outcome:     "APPROVED",
	}

	assert.Equal(t, int64(9999999999), input.CustomerID)

	data, err := json.Marshal(input)
	require.NoError(t, err)

	var decoded NotifyPMInput
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, int64(9999999999), decoded.CustomerID)
}
