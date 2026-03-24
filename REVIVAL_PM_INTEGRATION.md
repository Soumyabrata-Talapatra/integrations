# Revival ↔ Policy Management Integration

## Overview

This document describes the integration between the **Revival microservice** (`ims-revival-backend`) and the **Policy Management orchestrator** (`policy-management`). Both services use Temporal.io workflows. PM acts as the central lifecycle orchestrator; Revival is a downstream service that processes revival requests and signals completion back to PM.

---

## Architecture

```
                          REST Client
                              │
                              ▼
                ┌──────────────────────────┐
                │   PM REST Handler        │
                │  POST /policies/{pn}/    │
                │       requests/revival   │
                └────────────┬─────────────┘
                             │ 1. INSERT service_request (status=RECEIVED)
                             │ 2. Signal PLW ("revival-request")
                             │ 3. UPDATE service_request (status=ROUTED)
                             ▼
                ┌──────────────────────────┐
                │  PolicyLifecycleWorkflow │  (plw-{policyNumber})
                │  Task Queue:             │
                │    policy-management-tq  │
                │                          │
                │  handleFinancialRequest: │
                │   • State gate check     │
                │   • Acquire fin. lock    │
                │   • Fetch RequestPayload │
                │   • Spawn child workflow │
                └────────────┬─────────────┘
                             │ ExecuteChildWorkflow
                             │   Type: "InstallmentRevivalWorkflow"
                             │   Queue: "revival-tq"
                             │   ID: "rev-{idempotencyKey}"
                             │   Input: ChildWorkflowInput
                             ▼
                ┌──────────────────────────┐
                │ InstallmentRevivalWF     │  (revival-tq)
                │                          │
                │  Stages:                 │
                │   Validate Policy        │
                │   → Data Entry           │
                │   → QC                   │
                │   → Approval             │
                │   → First Collection     │
                │   → Installment Monitor  │
                │                          │
                │  On terminal state:      │
                │   • Write DB status      │
                │   • NotifyPMActivity     │
                └────────────┬─────────────┘
                             │ SignalWorkflow(pmWorkflowID,
                             │   "revival-completed",
                             │   PMCompletionSignal)
                             ▼
                ┌──────────────────────────┐
                │  PolicyLifecycleWorkflow │
                │                          │
                │  handleOperationCompleted│
                │   • Release fin. lock    │
                │   • Update service_req   │
                │   • State transition:    │
                │     APPROVED → ACTIVE    │
                │     REJECTED → revert    │
                └──────────────────────────┘
```

---

## Changed Files

### Policy Management (PM)

| File | Change |
|------|--------|
| `workflows/signals.go` | Added `PMWorkflowID` field to `ChildWorkflowInput` |
| `workflows/activities/policy_activities.go` | Added `FetchRequestPayloadActivity` |
| `workflows/policy_lifecycle_workflow.go` | Updated `handleFinancialRequest` to fetch payload and set `PMWorkflowID` |
| `workflows/pm_revival_integration_test.go` | **New** — 14 integration tests |

### Revival Service

| File | Change |
|------|--------|
| `configs/config.yaml` | Task queue: `"revival"` → `"revival-tq"` |
| `configs/config.dev.yaml` | Task queue: `"revival"` → `"revival-tq"` |
| `workflow/revival_workflow.go` | `IndexRevivalInput` accepts PM contract fields; `notifyPolicyManagement()` helper; PM notification at all terminal points |
| `workflow/activities.go` | Added `PMCompletionSignal`, `NotifyPolicyManagementActivity`, `temporalClient` to Activities struct |
| `bootstrap/temporal.go` | Registered `NotifyPolicyManagementActivity` |
| `handler/revival.go` | Removed `ValidatePolicyActivity` call (moved to workflow) |
| `workflow/pm_integration_test.go` | **New** — 11 workflow tests + 8 unit tests |

---

## Data Contracts

### 1. PM → Revival: `ChildWorkflowInput`

Sent via `ExecuteChildWorkflow` on the `revival-tq` task queue.

```go
type ChildWorkflowInput struct {
    RequestID        string          `json:"request_id"`          // UUID idempotency key
    PolicyNumber     string          `json:"policy_number"`
    PolicyDBID       int64           `json:"policy_db_id"`        // PM's BIGINT policy_id
    ServiceRequestID int64           `json:"service_request_id"`  // PM's service_request PK
    RequestType      string          `json:"request_type"`        // "REVIVAL"
    RequestPayload   json.RawMessage `json:"request_payload"`     // Original request body
    TimeoutAt        time.Time       `json:"timeout_at"`          // Routing timeout
    PMWorkflowID     string          `json:"pm_workflow_id"`      // "plw-{policyNumber}"
}
```

Revival deserializes this into `IndexRevivalInput`, which has matching JSON tags for all PM fields plus standalone-mode fields (`ticket_id`, `indexed_by`, etc.) that remain empty when started by PM.

### 2. Revival → PM: `PMCompletionSignal`

Sent via `SignalWorkflow` on the `"revival-completed"` channel to `plw-{policyNumber}`.

```go
// Revival sends this:
type PMCompletionSignal struct {
    RequestID       string    `json:"request_id"`
    RequestType     string    `json:"request_type"`            // "REVIVAL"
    Outcome         string    `json:"outcome"`                 // APPROVED | REJECTED | TIMEOUT
    StateTransition string    `json:"state_transition,omitempty"`
    CompletedAt     time.Time `json:"completed_at"`
}

// PM deserializes it as:
type OperationCompletedSignal struct {
    RequestID       string          `json:"request_id"`
    RequestType     string          `json:"request_type"`
    Outcome         string          `json:"outcome"`
    StateTransition string          `json:"state_transition,omitempty"`
    OutcomePayload  json.RawMessage `json:"outcome_payload,omitempty"`  // Optional
    CompletedAt     time.Time       `json:"completed_at"`
}
```

These are JSON-compatible. PM's extra `OutcomePayload` field is `omitempty` and stays `nil` when Revival doesn't send it.

---

## Signal Names & Task Queues

| Constant | Value | Direction |
|----------|-------|-----------|
| `SignalRevivalRequest` | `"revival-request"` | REST Handler → PM PLW |
| `SignalRevivalApproved` | `"revival-approved"` | Revival → PM PLW (phase-1) |
| `SignalRevivalCompleted` | `"revival-completed"` | Revival → PM PLW (phase-2) |
| PM Task Queue | `"policy-management-tq"` | PM workflows + activities |
| Revival Task Queue | `"revival-tq"` | Revival workflows + activities |
| Child Workflow ID | `"rev-{idempotencyKey}"` | PM-assigned, unique per request |
| PM Workflow ID | `"plw-{policyNumber}"` | Passed to Revival as `PMWorkflowID` |

---

## Two-Phase PM Notification

Revival uses a **two-phase signal design** to communicate with PM:

### Phase 1: `"revival-approved"` — Immediate Approval

Sent **immediately when the approver approves**, regardless of pending installments.

PM's `handleRevivalApproved` handler:
1. Releases the financial lock
2. Transitions policy to `ACTIVE`
3. **Keeps the PendingRequest** (so phase-2 can still match it)
4. Dedups with key `{RequestID}-approved` (does not block phase-2)

### Phase 2: `"revival-completed"` — Final Outcome

Sent when the revival reaches its final state (success or failure).

PM's existing `handleOperationCompleted` handler:
1. Matches and removes the PendingRequest
2. Updates `service_request` to COMPLETED
3. Applies state transition via `resolveCompletionTransition`:

| Outcome | New Policy Status | When |
|---------|-------------------|------|
| `APPROVED` | `ACTIVE` (no-op, already ACTIVE) | Not currently sent (all-paid path has no signal) |
| `VOID` | `VOID` | Installment default or 60-day SLA expired |

### Pre-Approval Failures (Single Phase)

If the request fails **before approval** (validation failure, approver rejects), only a single `"revival-completed"` signal is sent. PM handles it normally — the PendingRequest is still intact and the financial lock hasn't been released yet.

---

## Revival → PM Notification Points

| Event | Signal Channel | Outcome | State Transition |
|-------|---------------|---------|-----------------|
| Approver approves | `revival-approved` | `APPROVED` | `REVIVAL_PENDING→ACTIVE` |
| Validation fails | `revival-completed` | `REJECTED` | `REVIVAL_PENDING→VALIDATION_FAILED` |
| Approver rejects | `revival-completed` | `REJECTED` | `REVIVAL_PENDING→REJECTED` |
| 60-day SLA expires | `revival-completed` | `VOID` | `ACTIVE→VOID` |
| Installment default | `revival-completed` | `VOID` | `ACTIVE→VOID` |

When `PMWorkflowID` is empty (standalone mode), notifications are silently skipped.

---

## Dual-Mode Operation

Revival supports two modes:

### Standalone Mode (REST handler starts workflow directly)
- Revival's handler creates `IndexRevivalInput` with `TicketID`, `IndexedBy`, `Documents`, etc.
- `PMWorkflowID` is empty → `IsPMIntegrated()` returns `false`
- No PM notification on completion

### PM-Integrated Mode (PM starts workflow as child)
- PM sends `ChildWorkflowInput` → deserialized into `IndexRevivalInput`
- `PMWorkflowID` is set → `IsPMIntegrated()` returns `true`
- `RequestID` (PM's idempotency key) maps to `TicketID` and `PMRequestID`
- On completion, Revival signals PM via `NotifyPolicyManagementActivity`

---

## Validation Change

Policy validation (`ValidatePolicyActivity`) was **moved from the REST handler to the workflow**:

- **Before**: Handler called `ValidatePolicyActivity` → failure returned HTTP error, no workflow started
- **After**: Workflow calls `ValidatePolicyActivity` as its first activity → failure results in `VALIDATION_FAILED` state with PM notification

This ensures:
- PM always gets a completion signal (even for validation failures)
- The workflow has full control over the validation lifecycle
- DB audit trail captures validation outcomes

---

## New PM Activity

### `FetchRequestPayloadActivity`

```go
func (a *PolicyActivities) FetchRequestPayloadActivity(
    ctx context.Context,
    serviceRequestID int64,
    submittedAt *time.Time,
) (json.RawMessage, error)
```

- Fetches `request_payload` from `policy_mgmt.service_request` by `request_id`
- Uses `submitted_at` partition key when available (avoids cross-partition scans)
- Called in `handleFinancialRequest` before dispatching child workflows
- Auto-registered via struct-based registration (`w.RegisterActivity(policyActs)`)

---

## Test Coverage

### PM Tests (`pm_revival_integration_test.go`)

| Test | What it verifies |
|------|-----------------|
| `TestChildWorkflowInput_PMWorkflowIDField` | PMWorkflowID is present in struct |
| `TestChildWorkflowInput_PMWorkflowIDInJSON` | `pm_workflow_id` appears in JSON output |
| `TestChildWorkflowInput_RequestPayloadPopulated` | RequestPayload carries original request body |
| `TestRevivalCompletionSignal_*` | APPROVED / REJECTED / TIMEOUT outcomes |
| `TestResolveCompletionTransition_RevivalApproved` | REVIVAL + APPROVED → ACTIVE (non-terminal) |
| `TestResolveCompletionTransition_RevivalVoid` | REVIVAL + VOID → VOID (installment default or SLA timeout) |
| `TestDownstreamTaskQueue_Revival` | Routes to `"revival-tq"` |
| `TestDownstreamWorkflowType_Revival` | Maps to `"InstallmentRevivalWorkflow"` |
| `TestDownstreamChildIDPrefix_Revival` | Prefix is `"rev"` |
| `TestPreRouteStatus_Revival` | Pre-route status is `REVIVAL_PENDING` |
| `TestRoutingTimeout_Revival` | Timeout is 30 days |
| `TestRevivalSignalChannelNames` | All 3 signal names match constants (`revival-request`, `revival-approved`, `revival-completed`) |
| `TestRevivalRequiresFinancialLock` | Revival requires exclusive lock |
| `TestChildWorkflowInput_JSONRoundTrip` | Full JSON marshal/unmarshal preserves all fields |

### Revival Tests (`pm_integration_test.go`)

| Test | What it verifies |
|------|-----------------|
| `TestPMFieldsPropagatedToWorkflowState` | PM fields reach workflow state and are queryable |
| `TestPMNotificationOnRejection` | PM notified with REJECTED on approval denial |
| `TestPMNotificationOnValidationFailed` | PM notified with REJECTED on validation failure |
| `TestPMNotificationOnSLATimeout` | Phase-1 APPROVED + Phase-2 VOID both sent on SLA expiry |
| `TestPMNotificationOnCompletedNoPending` | Phase-1 APPROVED sent via `revival-approved` channel |
| `TestNoPMNotificationWhenStandalone` | No notification when PMWorkflowID is empty |
| `TestInstallmentMonitorCompleteNoPMNotification` | Child workflow does NOT send APPROVED (PM already notified at approval) |
| `TestPMNotificationOnInstallmentDefault` | Child workflow sends VOID via `revival-completed` → VOID |
| `TestNoNotificationFromChildWhenStandalone` | Child skips notification in standalone mode |
| `TestValidationRunsInWorkflow` | ValidatePolicyActivity called inside workflow |
| `TestValidationFailurePreventsRequestCreation` | Failed validation stops before DB insert |
| `TestIsPMIntegrated` | IsPMIntegrated() helper works for both modes |
| `TestPMContractFieldMapping` | PM contract fields deserialize correctly |
| Unit tests | PMCompletionSignal, IndexRevivalInput, InstallmentMonitorInput, RevivalWorkflowState field verification |
