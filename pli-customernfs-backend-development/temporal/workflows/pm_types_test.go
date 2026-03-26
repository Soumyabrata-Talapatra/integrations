package workflows

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSignalCustomerNFRCompleted_ConstantValue verifies the signal channel name constant.
func TestSignalCustomerNFRCompleted_ConstantValue(t *testing.T) {
	assert.Equal(t, "customer-nfr-completed", SignalCustomerNFRCompleted)
}

// TestOperationCompletedSignal_JSONRoundTrip verifies marshal -> unmarshal preserves all fields.
func TestOperationCompletedSignal_JSONRoundTrip(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	payload := json.RawMessage(`{"key":"value"}`)

	original := OperationCompletedSignal{
		RequestID:       "req-001",
		RequestType:     "ADDRESS_CHANGE",
		Outcome:         "APPROVED",
		StateTransition: "COMPLETED",
		OutcomePayload:  payload,
		CompletedAt:     now,
	}

	data, err := json.Marshal(original)
	require.NoError(t, err)

	var decoded OperationCompletedSignal
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, original.RequestID, decoded.RequestID)
	assert.Equal(t, original.RequestType, decoded.RequestType)
	assert.Equal(t, original.Outcome, decoded.Outcome)
	assert.Equal(t, original.StateTransition, decoded.StateTransition)
	assert.JSONEq(t, string(original.OutcomePayload), string(decoded.OutcomePayload))
	assert.True(t, original.CompletedAt.Equal(decoded.CompletedAt))
}

// TestOperationCompletedSignal_NilOutcomePayload verifies serialization with nil OutcomePayload.
func TestOperationCompletedSignal_NilOutcomePayload(t *testing.T) {
	signal := OperationCompletedSignal{
		RequestID:   "req-002",
		RequestType: "MOBILE_CHANGE",
		Outcome:     "REJECTED",
		CompletedAt: time.Now().UTC(),
	}

	data, err := json.Marshal(signal)
	require.NoError(t, err)

	var decoded OperationCompletedSignal
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, "req-002", decoded.RequestID)
	assert.Equal(t, "MOBILE_CHANGE", decoded.RequestType)
	assert.Equal(t, "REJECTED", decoded.Outcome)
	assert.Empty(t, decoded.StateTransition, "omitempty field should be empty when not set")
	assert.Nil(t, decoded.OutcomePayload, "nil OutcomePayload should remain nil after round-trip")
}

// TestOperationCompletedSignal_PopulatedOutcomePayload verifies round-trip with a populated payload.
func TestOperationCompletedSignal_PopulatedOutcomePayload(t *testing.T) {
	payload := json.RawMessage(`{"mobile_number":"9876543210","old_mobile":"1234567890"}`)
	now := time.Now().UTC().Truncate(time.Second)

	signal := OperationCompletedSignal{
		RequestID:       "req-003",
		RequestType:     "MOBILE_CHANGE",
		Outcome:         "APPROVED",
		StateTransition: "PENDING_APPROVAL->COMPLETED",
		OutcomePayload:  payload,
		CompletedAt:     now,
	}

	data, err := json.Marshal(signal)
	require.NoError(t, err)

	var decoded OperationCompletedSignal
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, signal.RequestID, decoded.RequestID)
	assert.JSONEq(t, `{"mobile_number":"9876543210","old_mobile":"1234567890"}`, string(decoded.OutcomePayload))
	assert.True(t, now.Equal(decoded.CompletedAt))
}
