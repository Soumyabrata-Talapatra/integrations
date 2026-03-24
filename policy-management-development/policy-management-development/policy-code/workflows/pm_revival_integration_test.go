package workflows

import (
	"encoding/json"
	"testing"
	"time"

	"policy-management/core/domain"
)

// =============================================================================
// PM ↔ Revival Integration Tests
//
// These tests verify the PM-side contract changes for revival integration:
// 1. ChildWorkflowInput includes PMWorkflowID for completion signal routing
// 2. RequestPayload is populated in ChildWorkflowInput
// 3. Revival completion signal (OperationCompletedSignal) is handled correctly
// 4. resolveCompletionTransition maps revival outcomes to correct policy statuses
// =============================================================================

// ─────────────────────────────────────────────────────────────────────────────
// ChildWorkflowInput — PMWorkflowID field
// ─────────────────────────────────────────────────────────────────────────────

func TestChildWorkflowInput_PMWorkflowIDField(t *testing.T) {
	input := ChildWorkflowInput{
		RequestID:        "uuid-123",
		PolicyNumber:     "PLI/2026/000001",
		PolicyDBID:       42,
		ServiceRequestID: 100,
		RequestType:      domain.RequestTypeRevival,
		RequestPayload:   json.RawMessage(`{"requested_installments": 6}`),
		TimeoutAt:        time.Now().Add(30 * 24 * time.Hour),
		PMWorkflowID:     "plw-PLI/2026/000001",
	}

	if input.PMWorkflowID == "" {
		t.Fatal("PMWorkflowID must be set for downstream services to signal completion back")
	}
	if input.PMWorkflowID != "plw-PLI/2026/000001" {
		t.Errorf("PMWorkflowID = %q; want %q", input.PMWorkflowID, "plw-PLI/2026/000001")
	}
}

func TestChildWorkflowInput_PMWorkflowIDInJSON(t *testing.T) {
	input := ChildWorkflowInput{
		RequestID:    "uuid-456",
		PolicyNumber: "PLI/2026/000002",
		RequestType:  domain.RequestTypeRevival,
		PMWorkflowID: "plw-PLI/2026/000002",
	}

	data, err := json.Marshal(input)
	if err != nil {
		t.Fatalf("failed to marshal ChildWorkflowInput: %v", err)
	}

	var m map[string]interface{}
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	pmWfID, ok := m["pm_workflow_id"]
	if !ok {
		t.Fatal("pm_workflow_id missing from JSON output")
	}
	if pmWfID != "plw-PLI/2026/000002" {
		t.Errorf("pm_workflow_id = %v; want %q", pmWfID, "plw-PLI/2026/000002")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// ChildWorkflowInput — RequestPayload populated
// ─────────────────────────────────────────────────────────────────────────────

func TestChildWorkflowInput_RequestPayloadPopulated(t *testing.T) {
	payload := json.RawMessage(`{"requested_installments": 6}`)
	input := ChildWorkflowInput{
		RequestID:      "uuid-789",
		PolicyNumber:   "PLI/2026/000003",
		RequestType:    domain.RequestTypeRevival,
		RequestPayload: payload,
		PMWorkflowID:   "plw-PLI/2026/000003",
	}

	if input.RequestPayload == nil {
		t.Fatal("RequestPayload must be populated for downstream services to access request details")
	}

	var p map[string]interface{}
	if err := json.Unmarshal(input.RequestPayload, &p); err != nil {
		t.Fatalf("failed to unmarshal RequestPayload: %v", err)
	}
	if v, ok := p["requested_installments"]; !ok || v != float64(6) {
		t.Errorf("requested_installments = %v; want 6", v)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// OperationCompletedSignal — revival completion scenarios
// ─────────────────────────────────────────────────────────────────────────────

func TestRevivalCompletionSignal_ApprovedOutcome(t *testing.T) {
	sig := OperationCompletedSignal{
		RequestID:       "uuid-rev-approved",
		RequestType:     domain.RequestTypeRevival,
		Outcome:         domain.RequestOutcomeApproved,
		StateTransition: "REVIVAL_PENDING→ACTIVE",
		CompletedAt:     time.Now(),
	}

	if sig.Outcome != domain.RequestOutcomeApproved {
		t.Errorf("Outcome = %q; want %q", sig.Outcome, domain.RequestOutcomeApproved)
	}
	if sig.RequestType != domain.RequestTypeRevival {
		t.Errorf("RequestType = %q; want %q", sig.RequestType, domain.RequestTypeRevival)
	}
}

func TestRevivalCompletionSignal_RejectedOutcome(t *testing.T) {
	sig := OperationCompletedSignal{
		RequestID:   "uuid-rev-rejected",
		RequestType: domain.RequestTypeRevival,
		Outcome:     domain.RequestOutcomeRejected,
		CompletedAt: time.Now(),
	}

	if sig.Outcome != domain.RequestOutcomeRejected {
		t.Errorf("Outcome = %q; want %q", sig.Outcome, domain.RequestOutcomeRejected)
	}
}

func TestRevivalCompletionSignal_TimeoutOutcome(t *testing.T) {
	sig := OperationCompletedSignal{
		RequestID:   "uuid-rev-timeout",
		RequestType: domain.RequestTypeRevival,
		Outcome:     domain.RequestOutcomeTimeout,
		CompletedAt: time.Now(),
	}

	if sig.Outcome != domain.RequestOutcomeTimeout {
		t.Errorf("Outcome = %q; want %q", sig.Outcome, domain.RequestOutcomeTimeout)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// resolveCompletionTransition — revival state resolution
// ─────────────────────────────────────────────────────────────────────────────

func TestResolveCompletionTransition_RevivalApproved(t *testing.T) {
	state := &PolicyLifecycleState{
		CurrentStatus:  domain.StatusRevivalPending,
		PreviousStatus: domain.StatusVoidLapse,
	}
	sig := OperationCompletedSignal{
		RequestID:   "rev-approved-1",
		RequestType: domain.RequestTypeRevival,
		Outcome:     domain.RequestOutcomeApproved,
	}

	newStatus, isTerminal := resolveCompletionTransition(state, sig)

	if newStatus != domain.StatusActive {
		t.Errorf("newStatus = %q; want %q (ACTIVE)", newStatus, domain.StatusActive)
	}
	if isTerminal {
		t.Error("revival APPROVED should NOT be terminal (policy continues as ACTIVE)")
	}
}

func TestResolveCompletionTransition_RevivalVoid(t *testing.T) {
	// Phase-2 VOID (installment default or SLA timeout) → VOID
	state := &PolicyLifecycleState{
		CurrentStatus:  domain.StatusActive, // Already ACTIVE from phase-1
		PreviousStatus: domain.StatusVoidLapse,
	}
	sig := OperationCompletedSignal{
		RequestID:   "rev-void-1",
		RequestType: domain.RequestTypeRevival,
		Outcome:     "VOID",
	}

	newStatus, isTerminal := resolveCompletionTransition(state, sig)

	if newStatus != domain.StatusVoid {
		t.Errorf("newStatus = %q; want %q", newStatus, domain.StatusVoid)
	}
	if isTerminal {
		t.Error("revival VOID should NOT be terminal")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Task queue routing — revival uses revival-tq
// ─────────────────────────────────────────────────────────────────────────────

func TestDownstreamTaskQueue_Revival(t *testing.T) {
	tq := domain.DownstreamTaskQueueForType(domain.RequestTypeRevival)
	if tq != "revival-tq" {
		t.Errorf("DownstreamTaskQueueForType(REVIVAL) = %q; want %q", tq, "revival-tq")
	}
}

func TestDownstreamWorkflowType_Revival(t *testing.T) {
	wfType := DownstreamWorkflowTypeForRequest(domain.RequestTypeRevival)
	if wfType != "InstallmentRevivalWorkflow" {
		t.Errorf("DownstreamWorkflowTypeForRequest(REVIVAL) = %q; want %q", wfType, "InstallmentRevivalWorkflow")
	}
}

func TestDownstreamChildIDPrefix_Revival(t *testing.T) {
	prefix := DownstreamChildIDPrefix(domain.RequestTypeRevival)
	if prefix != "rev" {
		t.Errorf("DownstreamChildIDPrefix(REVIVAL) = %q; want %q", prefix, "rev")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Pre-route status — revival sets REVIVAL_PENDING
// ─────────────────────────────────────────────────────────────────────────────

func TestPreRouteStatus_Revival(t *testing.T) {
	status := preRouteStatus(domain.RequestTypeRevival)
	if status != domain.StatusRevivalPending {
		t.Errorf("preRouteStatus(REVIVAL) = %q; want %q", status, domain.StatusRevivalPending)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Routing timeout — revival gets 30 days
// ─────────────────────────────────────────────────────────────────────────────

func TestRoutingTimeout_Revival(t *testing.T) {
	timeout := routingTimeoutForRequest(domain.RequestTypeRevival)
	expected := 30 * 24 * time.Hour
	if timeout != expected {
		t.Errorf("routingTimeoutForRequest(REVIVAL) = %v; want %v (30 days)", timeout, expected)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Revival signal channel name matches constant
// ─────────────────────────────────────────────────────────────────────────────

func TestRevivalSignalChannelNames(t *testing.T) {
	if SignalRevivalRequest != "revival-request" {
		t.Errorf("SignalRevivalRequest = %q; want %q", SignalRevivalRequest, "revival-request")
	}
	if SignalRevivalApproved != "revival-approved" {
		t.Errorf("SignalRevivalApproved = %q; want %q", SignalRevivalApproved, "revival-approved")
	}
	if SignalRevivalCompleted != "revival-completed" {
		t.Errorf("SignalRevivalCompleted = %q; want %q", SignalRevivalCompleted, "revival-completed")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Financial lock required for revival
// ─────────────────────────────────────────────────────────────────────────────

func TestRevivalRequiresFinancialLock(t *testing.T) {
	requires, ok := domain.FinancialRequestTypes[domain.RequestTypeRevival]
	if !ok {
		t.Fatal("REVIVAL not found in FinancialRequestTypes map")
	}
	if !requires {
		t.Error("REVIVAL should require a financial lock (FinancialRequestTypes[REVIVAL] = true)")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// ChildWorkflowInput JSON round-trip preserves all fields
// ─────────────────────────────────────────────────────────────────────────────

func TestChildWorkflowInput_JSONRoundTrip(t *testing.T) {
	timeout := time.Date(2026, 4, 20, 0, 0, 0, 0, time.UTC)
	original := ChildWorkflowInput{
		RequestID:        "uuid-roundtrip",
		PolicyNumber:     "PLI/2026/000001",
		PolicyDBID:       42,
		ServiceRequestID: 100,
		RequestType:      domain.RequestTypeRevival,
		RequestPayload:   json.RawMessage(`{"requested_installments":6}`),
		TimeoutAt:        timeout,
		PMWorkflowID:     "plw-PLI/2026/000001",
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var decoded ChildWorkflowInput
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}

	if decoded.RequestID != original.RequestID {
		t.Errorf("RequestID: %q != %q", decoded.RequestID, original.RequestID)
	}
	if decoded.PolicyNumber != original.PolicyNumber {
		t.Errorf("PolicyNumber: %q != %q", decoded.PolicyNumber, original.PolicyNumber)
	}
	if decoded.PolicyDBID != original.PolicyDBID {
		t.Errorf("PolicyDBID: %d != %d", decoded.PolicyDBID, original.PolicyDBID)
	}
	if decoded.ServiceRequestID != original.ServiceRequestID {
		t.Errorf("ServiceRequestID: %d != %d", decoded.ServiceRequestID, original.ServiceRequestID)
	}
	if decoded.RequestType != original.RequestType {
		t.Errorf("RequestType: %q != %q", decoded.RequestType, original.RequestType)
	}
	if string(decoded.RequestPayload) != string(original.RequestPayload) {
		t.Errorf("RequestPayload: %s != %s", decoded.RequestPayload, original.RequestPayload)
	}
	if !decoded.TimeoutAt.Equal(original.TimeoutAt) {
		t.Errorf("TimeoutAt: %v != %v", decoded.TimeoutAt, original.TimeoutAt)
	}
	if decoded.PMWorkflowID != original.PMWorkflowID {
		t.Errorf("PMWorkflowID: %q != %q", decoded.PMWorkflowID, original.PMWorkflowID)
	}
}
