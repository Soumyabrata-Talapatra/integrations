package workflows

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMustMarshalJSON_ValidMap verifies mustMarshalJSON returns non-nil JSON for a valid map.
func TestMustMarshalJSON_ValidMap(t *testing.T) {
	input := map[string]interface{}{
		"mobile_number": "9876543210",
		"old_mobile":    "1234567890",
	}

	result := mustMarshalJSON(input)

	require.NotNil(t, result, "mustMarshalJSON should return non-nil for a valid map")
	assert.True(t, json.Valid(result), "result must be valid JSON")
}

// TestMustMarshalJSON_RoundTrip verifies mustMarshalJSON result can be unmarshaled back to a map.
func TestMustMarshalJSON_RoundTrip(t *testing.T) {
	input := map[string]interface{}{
		"email":     "new@example.com",
		"old_email": "old@example.com",
	}

	result := mustMarshalJSON(input)
	require.NotNil(t, result)

	var decoded map[string]interface{}
	err := json.Unmarshal(result, &decoded)
	require.NoError(t, err)

	assert.Equal(t, "new@example.com", decoded["email"])
	assert.Equal(t, "old@example.com", decoded["old_email"])
}

// TestMustMarshalJSON_EmptyMap verifies mustMarshalJSON with empty map returns valid JSON "{}".
func TestMustMarshalJSON_EmptyMap(t *testing.T) {
	input := map[string]interface{}{}

	result := mustMarshalJSON(input)

	require.NotNil(t, result, "mustMarshalJSON should return non-nil for an empty map")
	assert.JSONEq(t, `{}`, string(result))
}
