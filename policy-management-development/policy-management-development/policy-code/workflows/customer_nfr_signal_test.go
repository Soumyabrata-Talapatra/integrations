package workflows

import (
	"encoding/json"
	"testing"
	"time"
)

// TestSignalCustomerNFRCompleted_Constant verifies the signal channel name
// matches the exact kebab-case string expected by Customer NFS callers.
func TestSignalCustomerNFRCompleted_Constant(t *testing.T) {
	const want = "customer-nfr-completed"
	if SignalCustomerNFRCompleted != want {
		t.Errorf("SignalCustomerNFRCompleted = %q, want %q", SignalCustomerNFRCompleted, want)
	}
}

// TestOperationCompletedSignal_JSONRoundTrip_CustomerNFR verifies that an
// OperationCompletedSignal with a customer NFR request type survives a
// JSON marshal/unmarshal round-trip without data loss.
func TestOperationCompletedSignal_JSONRoundTrip_CustomerNFR(t *testing.T) {
	now := time.Date(2026, 3, 25, 10, 0, 0, 0, time.UTC)
	original := OperationCompletedSignal{
		RequestID:   "req-addr-001",
		RequestType: "ADDRESS_CHANGE",
		Outcome:     "APPROVED",
		CompletedAt: now,
	}

	b, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var decoded OperationCompletedSignal
	if err := json.Unmarshal(b, &decoded); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if decoded.RequestID != original.RequestID {
		t.Errorf("RequestID: got %q, want %q", decoded.RequestID, original.RequestID)
	}
	if decoded.RequestType != original.RequestType {
		t.Errorf("RequestType: got %q, want %q", decoded.RequestType, original.RequestType)
	}
	if decoded.Outcome != original.Outcome {
		t.Errorf("Outcome: got %q, want %q", decoded.Outcome, original.Outcome)
	}
	if !decoded.CompletedAt.Equal(original.CompletedAt) {
		t.Errorf("CompletedAt: got %v, want %v", decoded.CompletedAt, original.CompletedAt)
	}
}

// TestOperationCompletedSignal_AllNFRTypes ensures all four customer NFR
// request types serialise and deserialise correctly (table-driven).
func TestOperationCompletedSignal_AllNFRTypes(t *testing.T) {
	types := []struct {
		name        string
		requestType string
	}{
		{"address change", "ADDRESS_CHANGE"},
		{"name change", "NAME_CHANGE"},
		{"mobile change", "MOBILE_CHANGE"},
		{"email change", "EMAIL_CHANGE"},
	}

	for _, tc := range types {
		t.Run(tc.name, func(t *testing.T) {
			sig := OperationCompletedSignal{
				RequestID:   "req-" + tc.requestType,
				RequestType: tc.requestType,
				Outcome:     "APPROVED",
				CompletedAt: time.Now().UTC(),
			}

			b, err := json.Marshal(sig)
			if err != nil {
				t.Fatalf("marshal error: %v", err)
			}

			var decoded OperationCompletedSignal
			if err := json.Unmarshal(b, &decoded); err != nil {
				t.Fatalf("unmarshal error: %v", err)
			}

			if decoded.RequestType != tc.requestType {
				t.Errorf("RequestType: got %q, want %q", decoded.RequestType, tc.requestType)
			}
			if decoded.Outcome != "APPROVED" {
				t.Errorf("Outcome: got %q, want %q", decoded.Outcome, "APPROVED")
			}
		})
	}
}

// TestOperationCompletedSignal_OutcomePayload_CustomerNFR verifies that
// OutcomePayload carries domain-specific data for each NFR type.
func TestOperationCompletedSignal_OutcomePayload_CustomerNFR(t *testing.T) {
	cases := []struct {
		name        string
		requestType string
		payload     map[string]interface{}
	}{
		{
			name:        "address payload",
			requestType: "ADDRESS_CHANGE",
			payload: map[string]interface{}{
				"address_type": "PERMANENT",
				"pin_code":     "110001",
				"state":        "Delhi",
				"district":     "Central Delhi",
				"city":         "New Delhi",
			},
		},
		{
			name:        "name payload",
			requestType: "NAME_CHANGE",
			payload: map[string]interface{}{
				"salutation":  "Mr",
				"first_name":  "Rahul",
				"middle_name": "",
				"last_name":   "Sharma",
			},
		},
		{
			name:        "mobile payload",
			requestType: "MOBILE_CHANGE",
			payload: map[string]interface{}{
				"mobile_number": "9876543210",
			},
		},
		{
			name:        "email payload",
			requestType: "EMAIL_CHANGE",
			payload: map[string]interface{}{
				"email": "user@example.com",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			raw, err := json.Marshal(tc.payload)
			if err != nil {
				t.Fatalf("marshal payload: %v", err)
			}

			sig := OperationCompletedSignal{
				RequestID:      "req-" + tc.requestType,
				RequestType:    tc.requestType,
				Outcome:        "APPROVED",
				OutcomePayload: raw,
				CompletedAt:    time.Now().UTC(),
			}

			b, err := json.Marshal(sig)
			if err != nil {
				t.Fatalf("marshal signal: %v", err)
			}

			var decoded OperationCompletedSignal
			if err := json.Unmarshal(b, &decoded); err != nil {
				t.Fatalf("unmarshal signal: %v", err)
			}

			if decoded.OutcomePayload == nil {
				t.Fatal("OutcomePayload is nil after round-trip")
			}

			var got map[string]interface{}
			if err := json.Unmarshal(decoded.OutcomePayload, &got); err != nil {
				t.Fatalf("unmarshal outcome payload: %v", err)
			}

			for k, want := range tc.payload {
				v, ok := got[k]
				if !ok {
					t.Errorf("missing key %q in OutcomePayload", k)
					continue
				}
				// json.Unmarshal decodes numbers as float64; compare as strings.
				gotStr, wantStr := stringify(v), stringify(want)
				if gotStr != wantStr {
					t.Errorf("key %q: got %q, want %q", k, gotStr, wantStr)
				}
			}
		})
	}
}

// stringify converts an interface value to a comparable string representation.
func stringify(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	default:
		b, _ := json.Marshal(val)
		return string(b)
	}
}
