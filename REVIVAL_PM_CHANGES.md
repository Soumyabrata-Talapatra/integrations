# Revival ↔ Policy Management — Change Log for Team Review

**Branch:** `claude/temporal-policy-integration-cSA9D`
**Date:** 2026-03-23

---

## What Changed and Why

We integrated the Revival microservice (`ims-revival-backend`) with the Policy Management orchestrator (`policy-management`) using Temporal signals. The key design decision is a **two-phase signal approach** that releases the financial lock immediately on approval while still tracking the installment lifecycle.

---

## Commits (oldest → newest)

| # | Commit | Summary |
|---|--------|---------|
| 1 | `a1cd4a0` | Core integration — Revival accepts PM contract fields, sends completion signals |
| 2 | `ea84ca9` | Integration tests for both Revival and PM sides |
| 3 | `7df729f` | PM-side changes — child workflow routing, signal handlers, state transitions |
| 4 | `d33b410` | Integration README (`REVIVAL_PM_INTEGRATION.md`) |
| 5 | `382af8a` | Move APPROVED notification to approval time (not installment completion) |
| 6 | `46a19dd` | Two-phase signal design (`revival-approved` + `revival-completed`) |
| 7 | `9f559f4` | Standardize post-approval failure outcome to `VOID` |

---

## Files Changed — Revival Service (`ims-revival-backend-development`)

### `workflow/revival_workflow.go`
- **`IndexRevivalInput`** — Added PM contract fields: `PMWorkflowID`, `PMRequestID`, `PolicyDBID`, `ServiceRequestID`, `RequestPayload`, `TimeoutAt`
- **`IsPMIntegrated()`** — Helper: returns `true` when `PMWorkflowID != ""`
- **`notifyPolicyManagement()`** — New helper, accepts `signalChannel` param (`"revival-approved"` or `"revival-completed"`)
- **`notifyPMFromChild()`** — Same as above, for use from `InstallmentMonitorWorkflow`
- **Approval stage (line ~628)** — Sends `"revival-approved"` with outcome `APPROVED` immediately
- **SLA timeout (line ~663)** — Sends `"revival-completed"` with outcome `VOID`
- **Installment default (line ~1119)** — Sends `"revival-completed"` with outcome `VOID`
- **Validation failure (line ~124)** — Sends `"revival-completed"` with outcome `REJECTED`
- **Approver rejects (line ~538)** — Sends `"revival-completed"` with outcome `REJECTED`
- **Validation moved** from REST handler into workflow (ensures PM always gets a signal)

### `workflow/activities.go`
- **`PMCompletionSignal`** — New struct with `RequestID`, `RequestType`, `Outcome`, `StateTransition`, `CompletedAt`
- **`NotifyPolicyManagementActivity`** — New activity, accepts `pmWorkflowID`, `signalChannel`, `signal`; calls `temporalClient.SignalWorkflow()`
- **`Activities` struct** — Added `temporalClient` field

### `bootstrap/temporal.go`
- Registered `NotifyPolicyManagementActivity`

### `configs/config.yaml` and `configs/config.dev.yaml`
- Task queue: `"revival"` → `"revival-tq"` (matches PM's routing table)

### `handler/revival.go`
- Removed `ValidatePolicyActivity` call from handler (moved to workflow)

### `workflow/pm_integration_test.go` *(new file — 835 lines)*
- 11 workflow-level tests + 8 unit tests covering all signal paths

---

## Files Changed — Policy Management (`policy-management-development`)

### `workflows/signals.go`
- **`SignalRevivalApproved = "revival-approved"`** — New constant (phase-1)
- **`SignalRevivalCompleted = "revival-completed"`** — Existing (now phase-2)
- **`ChildWorkflowInput`** — Added `PMWorkflowID` field

### `workflows/policy_lifecycle_workflow.go`
- **`revivalApprovedCh`** — New signal channel registered in main selector
- **`handleRevivalApproved()`** — New handler:
  - Releases financial lock
  - Transitions policy → `ACTIVE`
  - **Keeps PendingRequest** (so phase-2 can match it)
  - Dedups with key `{RequestID}-approved` (does not block phase-2)
- **`handleOperationCompleted()`** — Unchanged; handles phase-2 `"revival-completed"`
- **`resolveCompletionTransition()`** — Updated REVIVAL case:
  - `APPROVED` → `StatusActive` (no-op if already ACTIVE)
  - Any other outcome (`VOID`, etc.) → `StatusVoid`
- **`handleFinancialRequest()`** — Fetches `RequestPayload` via `FetchRequestPayloadActivity`, populates `PMWorkflowID` in `ChildWorkflowInput`

### `workflows/activities/policy_activities.go`
- **`FetchRequestPayloadActivity`** — New activity, fetches `request_payload` JSONB from `service_request` table

### `core/domain/service_request.go`
- **`DownstreamTaskQueueForType("REVIVAL")`** → returns `"revival-tq"`

### `workflows/pm_revival_integration_test.go` *(new file)*
- 15 tests covering struct fields, JSON round-trip, signal names, state transitions, routing

---

## Two-Phase Signal Design

```
                    APPROVAL
                       │
          ┌────────────┴────────────┐
          │  "revival-approved"     │  ← Phase 1
          │  Outcome: APPROVED      │
          │  PM Action:             │
          │   • Release fin. lock   │
          │   • Policy → ACTIVE     │
          │   • Keep PendingRequest │
          └────────────┬────────────┘
                       │
         ┌─────────────┼─────────────┐
         │             │             │
    ALL PAID      SLA TIMEOUT   INSTALLMENT
   (no signal)        │          DEFAULT
                      │             │
          ┌───────────┴─┐  ┌───────┴───────┐
          │ "revival-   │  │ "revival-     │  ← Phase 2
          │  completed" │  │  completed"   │
          │ Outcome:    │  │ Outcome:      │
          │  VOID       │  │  VOID         │
          │ PM Action:  │  │ PM Action:    │
          │  → VOID     │  │  → VOID       │
          │  Remove PR  │  │  Remove PR    │
          └─────────────┘  └───────────────┘
```

### Pre-Approval Failures (Single Phase)

If the request fails **before** approval (validation failure or approver rejects), only `"revival-completed"` is sent with outcome `REJECTED`. PM handles it normally — the PendingRequest is still intact and the financial lock hasn't been released yet.

---

## Signal Summary

| When | Signal Channel | Outcome | PM Transition |
|------|---------------|---------|--------------|
| Approver approves | `revival-approved` | `APPROVED` | → `ACTIVE`, release lock |
| Validation fails | `revival-completed` | `REJECTED` | → revert, release lock |
| Approver rejects | `revival-completed` | `REJECTED` | → revert, release lock |
| 60-day SLA expires | `revival-completed` | `VOID` | → `VOID`, remove PendingRequest |
| Installment not paid by 1st | `revival-completed` | `VOID` | → `VOID`, remove PendingRequest |
| All installments paid | *(no signal)* | — | Policy already `ACTIVE` |

---

## Things to Verify

1. **PM `handleRevivalApproved`** — Confirm the dedup key `{RequestID}-approved` doesn't conflict with other signal types
2. **`NotifyPolicyManagementActivity`** signature change — Now takes 3 params (`pmWorkflowID`, `signalChannel`, `signal`) instead of 2. Ensure `bootstrap/temporal.go` re-registration picks up the new signature
3. **Task queue alignment** — Both PM (`DownstreamTaskQueueForType`) and Revival (`config.yaml`) must use `"revival-tq"`
4. **Validation in workflow** — `ValidatePolicyActivity` now runs inside the workflow, not the REST handler. Verify the handler no longer calls it
5. **PendingRequest cleanup** — When all installments are paid (no phase-2 signal), the PendingRequest stays in PM state until TTL cleanup. Confirm this is acceptable or add a cleanup signal
6. **Standalone mode** — When `PMWorkflowID` is empty, all PM notifications are skipped. Verify existing standalone revival flows are unaffected
7. **`PMCompletionSignal` ↔ `OperationCompletedSignal` JSON compatibility** — PM has an extra `OutcomePayload` field (`omitempty`); Revival doesn't send it. Confirm deserialization works
