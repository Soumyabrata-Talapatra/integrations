package handler

import (
	"encoding/json"
	"testing"

	"policy-management/workflows"
)

// TestAdminVoidSignalPayload_JSONRoundTrip verifies that adminVoidSignalPayload
// serialises to JSON that workflows.AdminVoidSignal can deserialise correctly.
// Guards against JSON tag mismatches causing silent zero-values in PLW.
func TestAdminVoidSignalPayload_JSONRoundTrip(t *testing.T) {
	payload := adminVoidSignalPayload{
		RequestID:    "test-uuid-1234",
		Reason:       "fraud detected",
		AuthorizedBy: 42,
	}

	b, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var sig workflows.AdminVoidSignal
	if err := json.Unmarshal(b, &sig); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if sig.RequestID != payload.RequestID {
		t.Errorf("RequestID: got %q, want %q", sig.RequestID, payload.RequestID)
	}
	if sig.Reason != payload.Reason {
		t.Errorf("Reason: got %q, want %q", sig.Reason, payload.Reason)
	}
	if sig.AuthorizedBy != payload.AuthorizedBy {
		t.Errorf("AuthorizedBy: got %d, want %d", sig.AuthorizedBy, payload.AuthorizedBy)
	}
}

// TestReopenSignalPayload_JSONRoundTrip verifies that reopenSignalPayload
// serialises to JSON that workflows.ReopenRequestSignal can deserialise correctly.
func TestReopenSignalPayload_JSONRoundTrip(t *testing.T) {
	payload := reopenSignalPayload{
		RequestID:    "test-uuid-5678",
		ReopenReason: "error correction",
		AuthorizedBy: 99,
	}

	b, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var sig workflows.ReopenRequestSignal
	if err := json.Unmarshal(b, &sig); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if sig.RequestID != payload.RequestID {
		t.Errorf("RequestID: got %q, want %q", sig.RequestID, payload.RequestID)
	}
	if sig.ReopenReason != payload.ReopenReason {
		t.Errorf("ReopenReason: got %q, want %q", sig.ReopenReason, payload.ReopenReason)
	}
	if sig.AuthorizedBy != payload.AuthorizedBy {
		t.Errorf("AuthorizedBy: got %d, want %d", sig.AuthorizedBy, payload.AuthorizedBy)
	}
}

// TestWithdrawalSignalPayload_JSONRoundTrip verifies that withdrawalSignalPayload
// serialises to JSON that workflows.WithdrawalRequestSignal can deserialise correctly.
// CRITICAL: missing TargetRequestID means PLW handleWithdrawal() can never cancel
// the downstream child workflow or release the financial lock.
func TestWithdrawalSignalPayload_JSONRoundTrip(t *testing.T) {
	payload := withdrawalSignalPayload{
		RequestID:        "withdrawal-12345",
		TargetRequestID:  "idempotency-uuid-abcd",
		WithdrawalReason: "customer changed mind",
	}

	b, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var sig workflows.WithdrawalRequestSignal
	if err := json.Unmarshal(b, &sig); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if sig.RequestID != payload.RequestID {
		t.Errorf("RequestID: got %q, want %q", sig.RequestID, payload.RequestID)
	}
	if sig.TargetRequestID != payload.TargetRequestID {
		t.Errorf("TargetRequestID: got %q, want %q", sig.TargetRequestID, payload.TargetRequestID)
	}
	if sig.WithdrawalReason != payload.WithdrawalReason {
		t.Errorf("WithdrawalReason: got %q, want %q", sig.WithdrawalReason, payload.WithdrawalReason)
	}
}
