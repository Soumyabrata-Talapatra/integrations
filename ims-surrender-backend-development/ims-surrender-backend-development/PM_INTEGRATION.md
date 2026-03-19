# PM Integration — Surrender Service

This document explains the changes made to integrate the Surrender Service with
Policy Management (PM) as the central orchestrator for surrender initiation.

---

## Background

### Old Flow

```
Staff/Customer
      |
      v
POST /v1/surrender/index-surrender   [Surrender Service]
      |
      |-- 1. Eligibility check (handler layer)
      |-- 2. Create DB record (IndexSurrenderRequestRepo)
      |-- 3. Start VoluntarySurrenderWorkflow
      |       TaskQueue : surrender-task-queue
      |       WorkflowID: "voluntary-surrender-{surrenderRequestID}"
      v
VoluntarySurrenderWorkflow
      |-- Wait DE signal   → SubmitDEActivity
      |-- Wait QC signal   → SubmitQCActivity
      |-- Wait Approval    → SubmitApprovalActivity
      |-- Calculate
      |-- Payment
      `-- UpdatePolicyStatus
          (no signal back to PM)
```

### New Flow

```
Staff/Customer
      |
      v
POST /v1/policies/{policy_number}/requests/surrender   [Policy Management]
      |
      |-- Policy state gate check (PM)
      |-- Financial lock acquisition (PM)
      |-- Signals PLW (plw-{policyNumber})
      v
PolicyLifecycleWorkflow   [PM — Temporal]
      |
      | ExecuteChildWorkflow("SurrenderProcessingWorkflow")
      | TaskQueue : surrender-tq
      | WorkflowID: "sur-{idempotency-key}"
      v
SurrenderProcessingWorkflow   [Surrender Service — Temporal]
      |
      |-- Step 1 : IndexSurrenderActivity
      |            Creates DB record, stores temporal_workflow_id
      |
      |-- Step 2 : ValidateEligibilityActivity
      |            Product ineligibility + maturity date check
      |
      |-- Step 3 : Wait "de-completed" signal
      |            PUT /v1/surrender/submit-de   [Surrender Service]
      |            SubmitDEActivity
      |
      |-- Step 4 : Wait "qc-completed" signal
      |            PUT /v1/surrender/submit-qc   [Surrender Service]
      |            SubmitQCActivity
      |
      |-- Step 5 : Wait "approval-completed" signal
      |            PUT /v1/surrender/submit-approval   [Surrender Service]
      |            SubmitApprovalActivity
      |
      |-- Step 6 : CalculateSurrenderValueActivity
      |-- Step 7 : ProcessPaymentActivity
      |-- Step 8 : UpdatePolicyStatusActivity
      |
      `-- Step 9 : SignalPMWorkflowActivity
                   Sends "surrender-completed" → plw-{policyNumber}
                   PM releases financial lock and transitions policy state
```

---

## Problem Statements and How Each Was Solved

### 1. Wrong task queue

**Problem:** PM dispatches the child workflow to task queue `surrender-tq`, but
the surrender service worker was listening on `surrender-task-queue`.

**Fix:** `temporal/worker.go`
```go
// Before
SurrenderTaskQueue = "surrender-task-queue"

// After
SurrenderTaskQueue = "surrender-tq"
```

---

### 2. Wrong workflow name

**Problem:** PM invokes `SurrenderProcessingWorkflow` by name (resolved via
`DownstreamWorkflowTypeForRequest("SURRENDER")`), but the surrender service
registered `VoluntarySurrenderWorkflow`.

**Fix:** Renamed/rewrote the workflow in `temporal/workflows/voluntary_surrender_workflow.go`.

```go
// Before
func VoluntarySurrenderWorkflow(ctx workflow.Context, input VoluntarySurrenderWorkflowInput) error

// After
func SurrenderProcessingWorkflow(ctx workflow.Context, input SurrenderProcessingInput) error
```

Registration in `temporal/worker.go`:
```go
// Before
w.RegisterWorkflow(workflows.VoluntarySurrenderWorkflow)

// After
w.RegisterWorkflow(workflows.SurrenderProcessingWorkflow)
```

---

### 3. Input contract mismatch

**Problem:** The old workflow accepted `VoluntarySurrenderWorkflowInput`
(`SurrenderRequestID`, `PolicyID`, `RequestNumber`, `RequestedBy`). PM sends
`ChildWorkflowInput` with a completely different shape.

**Fix:** Created `temporal/workflows/pm_contract.go` which mirrors PM's
`ChildWorkflowInput` exactly:

```go
type SurrenderProcessingInput struct {
    RequestID        string          `json:"request_id"`
    PolicyNumber     string          `json:"policy_number"`
    PolicyDBID       int64           `json:"policy_db_id"`
    ServiceRequestID int64           `json:"service_request_id"`
    RequestType      string          `json:"request_type"`
    RequestPayload   json.RawMessage `json:"request_payload"`
    TimeoutAt        time.Time       `json:"timeout_at"`
}
```

`RequestPayload` is the original JSON body from the PM handler:
```json
{
  "source_channel": "CPC",
  "disbursement_method": "CHEQUE",
  "bank_account_id": 123,
  "reason": "financial need"
}
```

---

### 4. DB record creation and eligibility checks were in the handler

**Problem:** In the old flow, `POST /v1/surrender/index-surrender` created the
DB record and ran eligibility checks before the workflow started. In the new
flow, PM calls the workflow directly — the handler is bypassed entirely.

**Fix:** Both operations moved into the workflow itself.

**Step 1 — `IndexSurrenderActivity`** creates the DB record:

```go
// temporal/activities/voluntary_surrender_activities.go

func IndexSurrenderActivity(ctx context.Context, input IndexSurrenderInput) (*IndexSurrenderResult, error) {
    req := domain.IndexSurrenderRequestInput{
        PolicyNumber:              input.PolicyNumber,
        Surrender_request_channel: input.SurrenderRequestChannel,
        Stage_name:                input.Stage_name,
        TemporalWorkflowID:        input.TemporalWorkflowID,
        PMServiceRequestID:        input.PMServiceRequestID,
        PMPolicyDBID:              input.PMPolicyDBID,
    }
    serviceRequestID, err := activitiesInstance.surrenderRepo.IndexSurrenderRequestRepo(ctx, req)
    ...
}
```

**Step 2 — `ValidateEligibilityActivity`** checks business rules (policy state
gate is already checked by PM before dispatch, so it is not rechecked here):

```go
func ValidateEligibilityActivity(ctx context.Context, input ValidateEligibilityInput) (*ValidateEligibilityResult, error) {
    policy, err := activitiesInstance.surrenderRepo.FindByPolicyNumber(ctx, input.PolicyID)

    // Rule 1: ineligible products
    ineligibleProducts := []string{"AEA", "AEA-10", "GY"}
    for _, prod := range ineligibleProducts {
        if policy.Product_name == prod {
            reasons = append(reasons, ...)
        }
    }

    // Rule 2: maturity date check
    if policy.Maturity_date.Before(time.Now()) {
        reasons = append(reasons, "policy has reached maturity...")
    }

    return &ValidateEligibilityResult{Eligible: len(reasons) == 0, Reasons: reasons}, nil
}
```

---

### 5. Signal routing broken after workflow ID change

**Problem:** `SubmitDE`, `SubmitQC`, and `SubmitApproval` handlers constructed
the workflow ID by string concatenation:
```go
workflowID := "voluntary-surrender-" + req1.Surrender_request_id
```
In the new flow, the workflow ID is `"sur-{PM-idempotency-key}"` — not
derivable from `surrender_request_id`.

**Fix (two parts):**

**Part A — Store the workflow ID in the DB during Step 1.**

New columns added to `finservicemgmt.surrender_requests` via
`migrations/003_add_pm_integration.sql`:

```sql
ALTER TABLE finservicemgmt.surrender_requests
    ADD COLUMN IF NOT EXISTS temporal_workflow_id  TEXT,
    ADD COLUMN IF NOT EXISTS pm_service_request_id BIGINT,
    ADD COLUMN IF NOT EXISTS pm_policy_db_id       BIGINT;

CREATE INDEX IF NOT EXISTS idx_surrender_requests_temporal_workflow_id
    ON finservicemgmt.surrender_requests (temporal_workflow_id)
    WHERE temporal_workflow_id IS NOT NULL;
```

`IndexSurrenderActivity` passes `TemporalWorkflowID` (the running workflow's
own ID from `workflow.GetInfo(ctx).WorkflowExecution.ID`) to the repo INSERT.

**Part B — Handlers look up the workflow ID from DB.**

`repo/postgres/surrender_request.go` — new method:
```go
func (r *SurrenderRequestRepository) GetWorkflowIDBySurrenderRequestID(
    ctx context.Context, srID string,
) (string, error) {
    query := dblib.Psql.Select("temporal_workflow_id").
        From("finservicemgmt.surrender_requests").
        Where(sq.Eq{"surrender_request_id": srID})
    ...
}
```

`handler/voluntary_surrender.go` — each of SubmitDE / SubmitQC / SubmitApproval:
```go
// Before
workflowID := "voluntary-surrender-" + req1.Surrender_request_id

// After
workflowID, err := h.surrenderRepo.GetWorkflowIDBySurrenderRequestID(
    sctx.Ctx, req1.Surrender_request_id,
)
if err != nil {
    return 404 response
}
```

---

### 6. No outcome reporting back to PM

**Problem:** The old workflow had no mechanism to notify PM of success or
failure. PM's `PolicyLifecycleWorkflow` would wait indefinitely.

**Fix:** Created `temporal/activities/pm_signal_activity.go`.

`SignalPMWorkflowActivity` uses the Temporal client to signal
`"surrender-completed"` on `plw-{policyNumber}`:

```go
func SignalPMWorkflowActivity(ctx context.Context, input SignalPMWorkflowInput) error {
    payload := pmOperationCompletedSignal{
        RequestID:       input.RequestID,
        RequestType:     input.RequestType,
        Outcome:         input.Outcome,         // APPROVED | REJECTED | TIMEOUT
        StateTransition: input.StateTransition,
        OutcomePayload:  input.OutcomePayload,
        CompletedAt:     time.Now().UTC(),
    }
    return pmSignalInstance.temporalClient.SignalWorkflow(
        ctx, input.PMWorkflowID, "", input.SignalName, payload,
    )
}
```

The `signalPMBack` helper in the workflow calls this in **every terminal path**:

| Terminal path | Outcome | StateTransition |
|--------------|---------|----------------|
| Payment processed | `APPROVED` | `PENDING_SURRENDER→SURRENDERED` |
| Any step error | `REJECTED` | `PENDING_SURRENDER→ACTIVE` |
| DE/QC/Approval timeout | `TIMEOUT` | `PENDING_SURRENDER→ACTIVE` |

Timeouts: DE = 7 days, QC = 7 days, Approval = 30 days.

The activity is initialised with the Temporal client via `InitPMSignalActivities`
called from `bootstrap/bootstrapper.go`:
```go
fx.Invoke(
    activities.InitPMSignalActivities,  // must run before RegisterWorkflows
    temporal.RegisterWorkflows,
)
```

---

### 7. IndexSurrender handler decommissioned

**Problem:** `POST /v1/surrender/index-surrender` would conflict with the new
flow — calling it would create a duplicate DB record without an associated
running workflow.

**Fix:** The handler now returns HTTP 410 Gone:

```go
func (h *VoluntarySurrenderHandler) IndexSurrender(...) (interface{}, error) {
    return response.GetDEPendingResponse{
        StatusCode: 410,
        Success:    false,
        Message:    "This endpoint is decommissioned. Surrender requests must be " +
                    "initiated through Policy Management.",
    }, nil
}
```

---

## Summary of All File Changes

| File | Type | Change |
|------|------|--------|
| `migrations/003_add_pm_integration.sql` | New | Adds `temporal_workflow_id`, `pm_service_request_id`, `pm_policy_db_id` columns |
| `temporal/workflows/pm_contract.go` | New | `SurrenderProcessingInput`, `OperationCompletedSignal`, outcome/transition constants |
| `temporal/activities/pm_signal_activity.go` | New | `SignalPMWorkflowActivity` — signals PM workflow with surrender outcome |
| `temporal/workflows/voluntary_surrender_workflow.go` | Rewritten | `VoluntarySurrenderWorkflow` → `SurrenderProcessingWorkflow` with 9-step PM-driven flow |
| `temporal/worker.go` | Modified | Task queue name fixed; new workflow and activity registered |
| `temporal/activities/voluntary_surrender_activities.go` | Modified | Real `IndexSurrenderActivity` and `ValidateEligibilityActivity` implemented |
| `core/domain/surrender_request.go` | Modified | `IndexSurrenderRequestInput` — added 3 PM fields |
| `repo/postgres/surrender_request.go` | Modified | INSERT stores PM fields; new `GetWorkflowIDBySurrenderRequestID` method |
| `handler/voluntary_surrender.go` | Modified | SubmitDE/QC/Approval use DB lookup for workflow ID; IndexSurrender returns 410 |
| `bootstrap/bootstrapper.go` | Modified | Added `InitPMSignalActivities` invocation |
