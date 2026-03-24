# Revival Service <-> Policy Management Orchestrator Integration Report

**Date:** 2026-03-20
**Scope:** Integration analysis between `ims-revival-backend-development` (Revival Service) and `policy-management-development` (Policy Management Orchestrator)

---

## 1. Executive Summary

The Policy Management Orchestrator (PM) is designed as the **central coordinator** for all policy operations. Revival Service currently operates as a **standalone microservice** with its own REST endpoints, Temporal workflows, and direct policy status updates. Integration requires Revival to become a **child workflow** of PM, relinquishing control of policy state transitions and entry-point REST endpoints.

### Key Finding: The gap is significant but well-defined

PM already has **complete scaffolding** for Revival integration (signal channels, routing config, state transitions, eligibility checks). Revival needs to **restructure** its workflow entry point and replace direct policy status updates with PM completion signals.

---

## 2. Current Architecture (As-Is)

### 2.1 Revival Service (Standalone)

| Aspect | Current Implementation |
|--------|----------------------|
| **Task Queue** | `revival` (config.yaml) |
| **Entry Point** | REST handler `POST /v1/revival/requests/index` starts `InstallmentRevivalWorkflow` directly |
| **Workflow ID** | Self-generated: `revival-{ticketID}` pattern |
| **Request ID** | Self-generated UUID in `CreateRevivalRequestActivity` |
| **Policy Status Updates** | **Direct DB writes** via `policyRepo.UpdatePolicyStatus()` in activities (e.g., `"IF"`, `"AL"`) |
| **Policy Validation** | Own `ValidatePolicyActivity` checks `PolicyStatus == "AL"` against Revival's local `policies` table |
| **Internal Statuses** | INDEXED -> WAITING_FOR_DATA_ENTRY -> DATA_ENTRY_COMPLETE -> WAITING_FOR_QC -> APPROVAL_PENDING -> APPROVED -> ACTIVE -> COMPLETED/DEFAULTED/TERMINATED |

**Key Workflows:**
- `InstallmentRevivalWorkflow` — Main orchestrator (Indexing -> DE -> QC -> Approval -> Collection)
- `FirstCollectionWorkflow` — First premium+installment dual collection
- `ChequeMonitorWorkflow` — Cheque clearance tracking
- `InstallmentMonitorWorkflow` — Subsequent installment tracking (child, `ParentClosePolicy: ABANDON`)
- `BatchInstallmentProcessingWorkflow` — Batch installment signal relay

### 2.2 Policy Management Orchestrator (PM)

| Aspect | Implementation |
|--------|---------------|
| **Task Queue** | `policy-management-tq` |
| **Workflow** | `PolicyLifecycleWorkflow` (`plw-{policyNumber}`) — one per policy, runs for decades |
| **Revival Signal** | `revival-request` channel, routed as child workflow on `revival-tq` |
| **Completion Signal** | `revival-completed` channel, expects `OperationCompletedSignal` |
| **State Gate** | Revival allowed from: `VOID_LAPSE`, `INACTIVE_LAPSE`, `ACTIVE_LAPSE` |
| **Pre-Route Status** | Transitions policy to `REVIVAL_PENDING` before routing |
| **Routing Timeout** | 365 days (configurable via `routing_timeout_revival`) |
| **Child Workflow Type** | Expects `InstallmentRevivalWorkflow` on `revival-tq` |
| **Child WF ID Pattern** | `rev-{policyNumber}-{idempotencyKey}` |
| **Financial Lock** | Acquires exclusive lock; no other financial ops during revival |

---

## 3. Integration Gap Analysis

### 3.1 CRITICAL: Workflow Input Mismatch

**PM sends:** `ChildWorkflowInput`
```go
type ChildWorkflowInput struct {
    RequestID        string          // UUID from X-Idempotency-Key
    PolicyNumber     string
    PolicyDBID       int64           // PM's policy_id
    ServiceRequestID int64           // PM's service_request table ID
    RequestType      string          // "REVIVAL"
    RequestPayload   json.RawMessage // Original request body
    TimeoutAt        time.Time       // Deadline
}
```

**Revival expects:** `IndexRevivalInput`
```go
type IndexRevivalInput struct {
    TicketID     string
    PolicyNumber string
    RequestType  string
    IndexedBy    string
    IndexedDate  time.Time
    Documents    string
}
```

**Action Required:** Revival must accept `ChildWorkflowInput` as its workflow entry point OR create an adapter workflow.

### 3.2 CRITICAL: Direct Policy Status Updates Must Be Removed

Revival currently updates policy status directly in multiple locations:

| File | Activity | Direct Update | Must Change To |
|------|----------|---------------|----------------|
| `workflow/activities.go:392` | `ProcessInstallmentActivity` | `policyRepo.UpdatePolicyStatus(ctx, pn, "IF")` | Signal PM `revival-completed` with outcome `COMPLETED` |
| `workflow/activities.go:504` | `HandleDefaultActivity` | `policyRepo.UpdatePolicyStatus(ctx, pn, "AL")` | Signal PM `revival-completed` with outcome `DEFAULT` |
| `workflow/activities.go:803` | `FinalizeRevivalAfterFirstCollection` | `policyRepo.UpdatePolicyStatus(ctx, pn, "IF")` | Signal PM `revival-completed` with outcome `COMPLETED` |

**PM is the sole writer of policy lifecycle state.** Revival must never directly update `policy.current_status`.

### 3.3 CRITICAL: Request ID Ownership

- **Current:** Revival generates its own `request_id` (UUID) in `CreateRevivalRequestActivity`
- **Required:** PM generates `request_id` in its `service_request` table. Revival's `revival_requests.request_id` should be an FK to PM's `service_request.request_id`

### 3.4 CRITICAL: REST Entry Point Migration

- **Current:** Revival exposes `POST /v1/revival/requests/index` — portals call this directly
- **Required:** PM exposes `POST /api/v1/policies/{pn}/requests/revival` — portals call PM, PM routes to Revival as child workflow
- Revival must **remove** the indexing endpoint as a portal-facing entry point (keep internal workflow processing)

### 3.5 HIGH: Task Queue Name Mismatch

- **Revival config:** `temporal.taskqueue: "revival"`
- **PM expects:** `revival-tq` (in `financialSignalConfigs()`)

**Action:** Either change Revival's task queue to `revival-tq` or update PM's routing config.

### 3.6 HIGH: Completion Signal Not Implemented

Revival currently does **not** signal PM on completion. It needs a new activity:

```go
// SignalPMCompletionActivity signals Policy Management on revival outcome
func (a *Activities) SignalPMCompletionActivity(ctx context.Context, policyNumber string, signal OperationCompletedSignal) error {
    return a.temporalClient.SignalWorkflow(ctx,
        "plw-"+policyNumber, "",
        "revival-completed",
        signal,
    )
}
```

Must be called with the correct `OperationCompletedSignal`:
```go
OperationCompletedSignal{
    RequestID:       input.RequestID,
    RequestType:     "REVIVAL",
    Outcome:         "COMPLETED|DEFAULT|CANCELLED",
    StateTransition: "REVIVAL_PENDING->ACTIVE",  // or "REVIVAL_PENDING->ACTIVE_LAPSE"
    OutcomePayload:  json.RawMessage(`{
        "new_paid_to_date": "...",
        "revival_date": "...",
        "installments_paid": N,
        "total_collected": M
    }`),
    CompletedAt:     time.Now(),
}
```

### 3.7 HIGH: Policy Validation Duplication

- Revival's `ValidatePolicyActivity` checks `PolicyStatus == "AL"` and `OngoingRevivalCount` against Revival's own policies table
- PM's `isStateEligible()` checks status is one of `VOID_LAPSE`, `INACTIVE_LAPSE`, `ACTIVE_LAPSE`

**Issue:** PM uses different status codes (`VOID_LAPSE`/`INACTIVE_LAPSE`/`ACTIVE_LAPSE`) vs Revival (`AL`). Revival uses a **local copy** of the policies table with its own status codes.

**Resolution:** Revival should trust PM's state gate check (already done before routing). Revival's domain validation (5yr window, medical, max revivals) should remain, but status check should be removed or aligned.

### 3.8 MEDIUM: Withdrawal Signal Not Connected

- PM has `withdrawal-request` signal that cancels active requests and releases financial locks
- Revival has `POST /requests/:ticket_id/withdraw` endpoint that terminates its workflow
- These are not connected — PM withdrawal should propagate to Revival's child workflow

### 3.9 MEDIUM: Database Schema Independence

Revival has its own `policies` table (migration `000008_create_policies.up.sql`) with different columns and status codes than PM's `policy` table. Post-integration:
- Revival should **not maintain its own policy table**
- Policy data should come from PM via `ChildWorkflowInput.RequestPayload` or from PM's database via API
- Revival keeps its domain-specific tables (`revival_requests`, `installment_schedules`, `payment_transactions`, `suspense_accounts`, etc.)

### 3.10 LOW: Status Code Mapping

| PM Status | Revival Status | Notes |
|-----------|---------------|-------|
| `VOID_LAPSE` | `AL` | PM splits lapse into 3 states |
| `INACTIVE_LAPSE` | `AL` | Revival treats all lapse as single "AL" |
| `ACTIVE_LAPSE` | `AL` | Need mapping or unified codes |
| `REVIVAL_PENDING` | (not used) | PM sets this; Revival tracks internally |
| `ACTIVE` | `IF` ("In Force") | On completion, PM sets ACTIVE |

---

## 4. Integration Design

### 4.1 Target Architecture

```
Portal/CPC
    │
    ▼  POST /api/v1/policies/{pn}/requests/revival
┌──────────────────────────────────────────────┐
│  POLICY MANAGEMENT ORCHESTRATOR (PM)         │
│                                              │
│  PolicyLifecycleWorkflow (plw-{pn})          │
│    1. State gate: VL/IL/AL allowed           │
│    2. Financial lock acquired                │
│    3. Status → REVIVAL_PENDING               │
│    4. Insert service_request (RECEIVED→ROUTED)│
│    5. Start child workflow on revival-tq     │
│    6. Return 202 Accepted                    │
│                                              │
│  ... waits for revival-completed signal ...   │
│                                              │
│    7. Receive OperationCompletedSignal       │
│    8. Status → ACTIVE (or revert to AL)      │
│    9. Release financial lock                 │
│   10. Update service_request → COMPLETED     │
└──────────────────────────────────────────────┘
         │ child workflow
         ▼
┌──────────────────────────────────────────────┐
│  REVIVAL SERVICE (revival-tq)                │
│                                              │
│  InstallmentRevivalWorkflow(ChildWorkflowInput)│
│    1. Save revival_detail (FK → PM request_id)│
│    2. Domain validation (5yr, medical, etc.) │
│    3. Indexing → Data Entry → QC → Approval  │
│    4. First Collection (dual payment)        │
│    5. Installment Monitor (child, ABANDON)   │
│    6. On completion: Signal PM               │
│       "revival-completed" with outcome       │
│    7. On default: Signal PM                  │
│       "revival-completed" with DEFAULT       │
└──────────────────────────────────────────────┘
```

### 4.2 What Revival Keeps (Domain Logic)

- Domain validation: 5yr window, medical exam requirements, max revival count
- Approval workflow: Indexing -> Data Entry -> QC -> Approver (with rework loops)
- Installment calculation, re-revival suspense adjustment
- First collection (dual payment) processing
- Installment monitoring and default handling
- Cheque clearance tracking
- Payment transaction management
- Suspense account management
- Letter generation and notifications
- All domain-specific DB tables (revival_requests, installment_schedules, payment_transactions, etc.)

### 4.3 What Revival Removes/Changes

| Change | Scope | Details |
|--------|-------|---------|
| Remove portal REST entry point | `handler/revival.go` | Remove `IndexRevivalRequest` as portal-facing; PM now handles intake |
| Accept `ChildWorkflowInput` | `workflow/revival_workflow.go` | Change workflow signature or add adapter |
| Remove direct policy status updates | `workflow/activities.go` | Remove all `policyRepo.UpdatePolicyStatus()` calls |
| Add PM completion signal | New activity | Signal `plw-{pn}` on `revival-completed` channel |
| Use PM's request_id as FK | `repo/postgres/revival.go` | `revival_requests.request_id` FK to PM's `service_request.request_id` |
| Task queue rename | `configs/config.yaml` | `revival` -> `revival-tq` |
| Remove local policies table | Migration + repo | Use PM's policy data from `ChildWorkflowInput` or cross-service query |

---

## 5. Implementation Plan (Step-by-Step)

### Phase 1: Contract Alignment (No behavior change)

1. **Define shared types package** — Create a shared Go package (or copy) with:
   - `ChildWorkflowInput` struct
   - `OperationCompletedSignal` struct
   - Signal channel constants

2. **Rename task queue** — Change `configs/config.yaml` from `revival` to `revival-tq`

3. **Add `request_id` FK column** — Migration to add PM's `request_id` reference to `revival_requests`

### Phase 2: Workflow Adapter

4. **Create wrapper workflow** — Adapt `ChildWorkflowInput` to `IndexRevivalInput`:
   ```go
   func InstallmentRevivalWorkflow(ctx workflow.Context, input ChildWorkflowInput) error {
       // Extract revival-specific fields from input.RequestPayload
       var revivalPayload RevivalPayloadFromPM
       json.Unmarshal(input.RequestPayload, &revivalPayload)

       // Map to internal processing
       internalInput := InternalRevivalInput{
           PMRequestID:  input.RequestID,
           PolicyNumber: input.PolicyNumber,
           PolicyDBID:   input.PolicyDBID,
           // ... map fields from revivalPayload
       }
       return processRevival(ctx, internalInput)
   }
   ```

5. **Preserve internal workflow logic** — The Indexing->DE->QC->Approval->Collection pipeline stays identical

### Phase 3: Signal PM on Completion

6. **Add Temporal client to Activities** — Revival's `Activities` struct needs access to Temporal client for signaling PM

7. **Add `SignalPMCompletionActivity`** — New activity that sends `OperationCompletedSignal` to `plw-{policyNumber}`

8. **Replace policy status updates** — In `ProcessInstallmentActivity`, `HandleDefaultActivity`, `FinalizeRevivalAfterFirstCollection`: remove `policyRepo.UpdatePolicyStatus()`, add `SignalPMCompletionActivity` call

### Phase 4: Entry Point Migration

9. **Remove portal-facing REST endpoints** — Or re-route them to call PM instead
10. **Keep internal management endpoints** — Revival's data-entry/QC/approval signal endpoints remain (these are internal CPC operations on an already-routed request)
11. **Keep query endpoints** — `GET /revival/requests/:ticket_id`, installment details, etc.

### Phase 5: Database Cleanup

12. **Remove local policies table dependency** — Policy data comes from PM's `ChildWorkflowInput`
13. **Add FK constraint** — `revival_requests.pm_request_id REFERENCES pm_service_request(request_id)`
14. **Align status codes** — Map Revival's internal statuses to a consistent vocabulary or keep them as Revival-internal

---

## 6. Risk Assessment

| Risk | Impact | Mitigation |
|------|--------|------------|
| 365-day workflow timeout | PM may time out the child before Revival completes installment collection | Configure `routing_timeout_revival` to match Revival's max lifecycle (12+ months) |
| Continue-As-New in PM | PM's PLW may CAN while Revival child is running | Already handled: `ParentClosePolicy: ABANDON` + empty RunID in signal target |
| Financial lock held for months | Revival can take 6-12 months; locks out surrender/loan | This is by design (BR-PM-013) — only one financial op at a time |
| Double status writes during migration | Both PM and Revival may try to update policy status | Must be atomic: Revival removes direct writes before PM starts routing |
| Existing in-flight workflows | Active Revival workflows started before integration | Run old and new in parallel; old workflows complete independently |
| Death notification during revival | Death preempts all; PM cancels pending requests | Revival's child workflow continues (`ABANDON` policy) but PM ignores completion signal |

---

## 7. Files Requiring Changes

### Revival Service (`ims-revival-backend-development`)

| File | Change Type | Description |
|------|-------------|-------------|
| `workflow/revival_workflow.go` | **Major** | Change `InstallmentRevivalWorkflow` input from `IndexRevivalInput` to `ChildWorkflowInput` |
| `workflow/activities.go:392,504,803` | **Major** | Remove `policyRepo.UpdatePolicyStatus()` calls, add PM signal |
| `workflow/activities.go` (new) | **New** | Add `SignalPMCompletionActivity` |
| `bootstrap/temporal.go:59` | **Minor** | Register new activity; update workflow registration |
| `configs/config.yaml` | **Minor** | Task queue: `revival` -> `revival-tq` |
| `handler/revival.go:60` | **Major** | Remove/refactor `IndexRevivalRequest` endpoint |
| `core/domain/revival.go` | **Minor** | Add `PMRequestID` field to `RevivalRequest` |
| `migrations/` (new) | **New** | Add `pm_request_id` column, drop policies table dependency |

### Policy Management (`policy-management-development`)

| File | Change Type | Description |
|------|-------------|-------------|
| `handler/policy_request_handler.go` | **Verify** | Ensure revival request endpoint is wired |
| `workflows/policy_lifecycle_workflow.go` | **Verify** | Revival signal handling already implemented |
| `workflows/signals.go` | **Verify** | `SignalRevivalRequest`/`SignalRevivalCompleted` already defined |

PM side is **mostly ready** — the scaffolding for Revival routing is complete in the workflow and handler code.

---

## 8. Conclusion

The integration path is **well-defined and tractable**. PM has already built the complete scaffolding for Revival as a child workflow — signal channels, state transitions, eligibility checks, and routing configuration are all in place. The primary work falls on the Revival Service side:

1. Accept PM's `ChildWorkflowInput` instead of self-starting
2. Signal PM on completion/default instead of updating policy status directly
3. Use PM's `request_id` as a foreign key
4. Rename task queue to `revival-tq`

The Revival Service's core domain logic (approval pipeline, installment tracking, suspense management) remains **unchanged**. The integration is essentially rewiring the entry point and exit point while keeping the business processing pipeline intact.
