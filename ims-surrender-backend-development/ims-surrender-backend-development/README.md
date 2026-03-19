# IMS Surrender Backend

Surrender processing microservice for PLI/RPLI policies. Handles the full voluntary surrender lifecycle — Data Entry, Quality Check, Approval, calculation, and payment — via Temporal workflows.

---

## Architecture Overview

The surrender service operates as a **downstream child service** under Policy Management (PM). PM is the sole entry point for initiating a surrender; the surrender service handles all subsequent processing steps and signals the outcome back to PM.

```
Customer/Staff
      |
      v
POST /v1/policies/{policy_number}/requests/surrender   [Policy Management]
      |
      | Signals PLW (plw-{policy_number})
      v
PolicyLifecycleWorkflow  [Policy Management — Temporal]
      |
      | ExecuteChildWorkflow("SurrenderProcessingWorkflow")
      | TaskQueue: surrender-tq
      v
SurrenderProcessingWorkflow  [Surrender Service — Temporal]
      |
      |-- Step 1: IndexSurrenderActivity        (creates DB record, stores workflow ID)
      |-- Step 2: ValidateEligibilityActivity   (product + maturity business rules)
      |
      |-- Step 3: Wait "de-completed" signal    <-- PUT /v1/surrender/submit-de
      |           SubmitDEActivity
      |
      |-- Step 4: Wait "qc-completed" signal    <-- PUT /v1/surrender/submit-qc
      |           SubmitQCActivity
      |
      |-- Step 5: Wait "approval-completed"     <-- PUT /v1/surrender/submit-approval
      |           SubmitApprovalActivity
      |
      |-- Step 6: CalculateSurrenderValueActivity
      |-- Step 7: ProcessPaymentActivity
      |-- Step 8: UpdatePolicyStatusActivity
      |
      `-- Step 9: SignalPMWorkflowActivity
                  Sends "surrender-completed" → plw-{policy_number}
```

---

## PM Integration Changes (March 2026)

### What Changed

Previously, the surrender service was self-contained: a direct API call to `POST /v1/surrender/index-surrender` created the DB record and started the Temporal workflow. This endpoint is now **decommissioned (returns HTTP 410)**. Policy Management is the only entry point.

### Flow Differences

| Aspect | Old Flow | New Flow |
|--------|----------|----------|
| Entry point | `POST /v1/surrender/index-surrender` | `POST /v1/policies/{pn}/requests/surrender` (PM) |
| Workflow start | Handler starts `VoluntarySurrenderWorkflow` | PM dispatches `SurrenderProcessingWorkflow` as child |
| Task queue | `surrender-task-queue` | `surrender-tq` (matches PM config) |
| Workflow name | `VoluntarySurrenderWorkflow` | `SurrenderProcessingWorkflow` |
| DB record creation | Handler before workflow start | Step 1 inside workflow (`IndexSurrenderActivity`) |
| Eligibility validation | Handler layer (before workflow) | Step 2 inside workflow (`ValidateEligibilityActivity`) |
| Signal routing | `workflowID = "voluntary-surrender-" + surrenderRequestID` | Look up `temporal_workflow_id` from DB by `surrender_request_id` |
| Outcome reporting | None | `SignalPMWorkflowActivity` → "surrender-completed" on `plw-{policyNumber}` |

### DE / QC / Approval Endpoints (unchanged URLs, updated internals)

These endpoints remain at the surrender service and work the same way from a caller's perspective:

- `PUT /v1/surrender/submit-de`
- `PUT /v1/surrender/submit-qc`
- `PUT /v1/surrender/submit-approval`

**Internal change:** They previously constructed the workflow ID by string concatenation (`"voluntary-surrender-" + surrender_request_id`). They now look up `temporal_workflow_id` from `finservicemgmt.surrender_requests` using the `surrender_request_id` field, then signal that workflow directly.

This lookup works because `IndexSurrenderActivity` (step 1 of the workflow) stores the Temporal workflow ID in the DB as soon as a surrender is initiated by PM.

---

## Files Changed

### New Files

| File | Purpose |
|------|---------|
| `migrations/003_add_pm_integration.sql` | Adds `temporal_workflow_id`, `pm_service_request_id`, `pm_policy_db_id` columns to `finservicemgmt.surrender_requests` |
| `temporal/workflows/pm_contract.go` | Input/output contract types shared with PM: `SurrenderProcessingInput`, `OperationCompletedSignal`, outcome constants |
| `temporal/activities/pm_signal_activity.go` | `SignalPMWorkflowActivity` — signals `surrender-completed` back to PM's `plw-{policyNumber}` workflow |

### Modified Files

| File | What Changed |
|------|-------------|
| `temporal/workflows/voluntary_surrender_workflow.go` | Replaced `VoluntarySurrenderWorkflow` with `SurrenderProcessingWorkflow`; added `signalPMBack` helper; all 9 steps |
| `temporal/worker.go` | Task queue `surrender-task-queue` → `surrender-tq`; registers `SurrenderProcessingWorkflow` and `SignalPMWorkflowActivity` |
| `temporal/activities/voluntary_surrender_activities.go` | Implemented real `IndexSurrenderActivity` (calls repo); implemented real `ValidateEligibilityActivity` (product/maturity checks); updated `IndexSurrenderInput` struct with PM fields |
| `core/domain/surrender_request.go` | Added `TemporalWorkflowID`, `PMServiceRequestID`, `PMPolicyDBID` fields to `IndexSurrenderRequestInput` |
| `repo/postgres/surrender_request.go` | `IndexSurrenderRequestRepo` now stores the 3 PM fields; new `GetWorkflowIDBySurrenderRequestID` method for signal routing |
| `handler/voluntary_surrender.go` | `SubmitDE/QC/Approval` look up `temporal_workflow_id` from DB; `IndexSurrender` returns HTTP 410 |
| `bootstrap/bootstrapper.go` | Added `activities.InitPMSignalActivities` invocation so the Temporal client is injected before the worker starts |

---

## Database Migration

Run `migrations/003_add_pm_integration.sql` before deploying this version:

```sql
ALTER TABLE finservicemgmt.surrender_requests
    ADD COLUMN IF NOT EXISTS temporal_workflow_id  TEXT,
    ADD COLUMN IF NOT EXISTS pm_service_request_id BIGINT,
    ADD COLUMN IF NOT EXISTS pm_policy_db_id       BIGINT;

CREATE INDEX IF NOT EXISTS idx_surrender_requests_temporal_workflow_id
    ON finservicemgmt.surrender_requests (temporal_workflow_id)
    WHERE temporal_workflow_id IS NOT NULL;
```

| Column | Type | Purpose |
|--------|------|---------|
| `temporal_workflow_id` | TEXT | Temporal workflow ID (`sur-{uuid}`) — used by DE/QC/Approval handlers to signal the correct workflow |
| `pm_service_request_id` | BIGINT | Cross-reference to PM's `service_request.request_id` |
| `pm_policy_db_id` | BIGINT | Cross-reference to PM's `policy.policy_id` |

---

## PM Contract

### Input (PM → Surrender)

PM sends this as the child workflow input. Field names must match exactly.

```go
type SurrenderProcessingInput struct {
    RequestID        string          `json:"request_id"`          // PM idempotency key (UUID)
    PolicyNumber     string          `json:"policy_number"`       // e.g. "PLI/2026/000001"
    PolicyDBID       int64           `json:"policy_db_id"`        // PM's BIGINT policy_id
    ServiceRequestID int64           `json:"service_request_id"`  // PM's service_request PK
    RequestType      string          `json:"request_type"`        // always "SURRENDER"
    RequestPayload   json.RawMessage `json:"request_payload"`     // original surrender JSON body
    TimeoutAt        time.Time       `json:"timeout_at"`
}
```

`RequestPayload` contains:
```json
{
  "source_channel": "CPC",
  "disbursement_method": "CHEQUE",
  "bank_account_id": 123,
  "reason": "..."
}
```

### Output (Surrender → PM)

The `surrender-completed` signal is sent on the `plw-{policyNumber}` workflow with:

```go
type OperationCompletedSignal struct {
    RequestID       string          `json:"request_id"`
    RequestType     string          `json:"request_type"`        // "SURRENDER"
    Outcome         string          `json:"outcome"`             // APPROVED | REJECTED | TIMEOUT
    StateTransition string          `json:"state_transition"`    // e.g. "PENDING_SURRENDER→SURRENDERED"
    OutcomePayload  json.RawMessage `json:"outcome_payload"`     // optional failure reason
    CompletedAt     time.Time       `json:"completed_at"`
}
```

| Outcome | StateTransition | When |
|---------|----------------|------|
| `APPROVED` | `PENDING_SURRENDER→SURRENDERED` | Payment processed successfully |
| `REJECTED` | `PENDING_SURRENDER→ACTIVE` | Any step fails (index, eligibility, DE, QC, approval, payment) |
| `TIMEOUT` | `PENDING_SURRENDER→ACTIVE` | DE not completed in 7 days, QC not completed in 7 days, or Approval not completed in 30 days |

---

## Eligibility Validation (in Workflow)

`ValidateEligibilityActivity` (step 2) enforces surrender-domain rules. Policy state gate (ACTIVE / VOID_LAPSE / etc.) is checked by PM before dispatch and is **not** rechecked here.

Rules checked:
- **Ineligible products:** AEA, AEA-10, GY — surrender not allowed
- **Maturity date:** if `maturity_date` is in the past, redirect to Maturity Claims

---

## Signal Timeouts

| Step | Signal | Timeout |
|------|--------|---------|
| Data Entry | `de-completed` | 7 days |
| Quality Check | `qc-completed` | 7 days |
| Approval | `approval-completed` | 30 days |

On timeout the workflow sends `TIMEOUT` / `PENDING_SURRENDER→ACTIVE` back to PM and returns an error.
