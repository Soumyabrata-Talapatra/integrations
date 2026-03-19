# Policy Issue → Policy Management Integration Report

**Last Updated:** March 2026
**Scope:** `ims-policyissue-backend-development` integration with `policy-management-development`
**Requirement Source:** `req_Policy_Management_Orchestrator_v4_1.md` — §10.1, GAP-PM-010
**Status:** PARTIALLY IMPLEMENTED — Signal wired + reliability added; payload contract incomplete

---

## Document Purpose

This report tracks the integration between the Policy Issue (PI) service and the Policy
Management (PM) service. It was originally raised as a critical gap (PLW never spawned). This
version reflects what has been implemented and what still needs to be done before the
integration can be considered production-ready.

---

## Executive Summary

| Area | Before | Now |
|------|--------|-----|
| Step 10 in `PolicyIssuanceWorkflow` | ❌ Missing entirely | ✅ Added |
| PM signal activity (`StartPMLifecycleActivity`) | ❌ Missing | ✅ Added with DB tracking |
| Signal failure recovery | ❌ None | ✅ Reconciliation workflow (every 15 min) |
| DB tracking of signal state | ❌ None | ✅ 5 columns in `proposal_issuance` |
| Unit tests | ❌ None | ✅ 15 test cases |
| Full PM signal payload (`PolicyCreatedSignal` with `PolicyMetadata`) | ❌ Missing | ⚠️ Simplified stub — not complete |
| `pm_contract.go` (shared PM types) | ❌ Missing | ❌ Still missing |
| `InstantIssuanceWorkflow` Step 10 | ❌ Missing | ❌ Still missing |
| `configs/config.yaml` — `pm.task_queue` | ❌ Missing | ⚠️ Config key works, YAML not updated |

**The integration is now connected and won't be silently dropped, but the signal content
PM receives does not yet include the full `PolicyMetadata` it expects. This must be
completed before go-live.**

---

## 1. Why This Integration Is Critical

Every policy operation (surrender, loan, revival, claim, commutation, conversion, FLC) is
gated behind the PM `PolicyLifecycleWorkflow` (PLW). **Without a running PLW, PM will reject
all requests for that policy** — no state gate check can run, no financial lock can be acquired.

The PLW is born exactly once per policy, spawned by Policy Issue Service immediately after
policy activation via a Temporal `SignalWithStart`. This is the **birth event** of the
entire PM lifecycle (req §10.1, GAP-PM-010: *"CRITICAL FIX — PLW won't spawn without this"*).

---

## 2. What PM Expects (The Contract — Unchanged)

### 2.1 Mechanism

| Field | Value |
|-------|-------|
| Temporal mechanism | `SignalWithStart` (atomic: workflow creation + signal) |
| Target task queue | `policy-management-tq` |
| Workflow ID | `plw-{policy_number}` (e.g. `plw-PLI/2026/000001`) |
| Workflow type | `PolicyLifecycleWorkflow` |
| Signal channel | `policy-created` |
| Workflow input type | `PolicyLifecycleState` (initial status = `FREE_LOOK_ACTIVE`) |
| Signal payload type | `PolicyCreatedSignal` |
| Idempotency | `SignalWithStart` is inherently idempotent — safe to retry |

### 2.2 Signal Payload PM expects (`PolicyCreatedSignal`)

```go
type PolicyCreatedSignal struct {
    RequestID    string         `json:"request_id"`    // UUID idempotency key for dedup
    PolicyID     string         `json:"policy_id"`     // PI's UUID (audit cross-ref)
    PolicyNumber string         `json:"policy_number"` // e.g. PLI/2026/000001
    Metadata     PolicyMetadata `json:"metadata"`      // full policy metadata — see §2.3
}
```

### 2.3 `PolicyMetadata` fields required at birth

| Field | Source in PI service | Implemented? |
|-------|---------------------|-------------|
| `CustomerID` (int64) | Parse from `input.CustomerID` (string → int64) | ⚠️ Not in current payload |
| `ProductCode` | `input.ProductCode` | ⚠️ Not in current payload |
| `ProductType` | Derived: `PLI` / `RPLI` from `input.PolicyType` | ⚠️ Partially — `PolicyType` sent |
| `SumAssured` | `input.SumAssured` | ⚠️ Not in current payload |
| `CurrentPremium` | `premiumResult.TotalPayable` (Step 3) | ⚠️ Not in current payload |
| `PremiumMode` | `input.PremiumPaymentFrequency` | ⚠️ Not in current payload |
| `BillingMethod` | From proposal record (`CASH` / `PAY_RECOVERY`) | ⚠️ Not in current payload |
| `IssueDate` | `time.Now().UTC()` at issuance | ✅ `IssuedAt` present |
| `MaturityDate` | `CreatePolicyIssuanceActivity` — `ProposalDate + PolicyTerm` | ⚠️ Not in current payload |
| `PaidToDate` | Same as `IssueDate` at first activation | ⚠️ Not in current payload |
| `AgentID` | From proposal record (nullable) | ⚠️ Not in current payload |
| `NominationStatus` | `"PENDING"` at birth | ⚠️ Not in current payload |
| `IsDistanceMarketing` | Product config flag (30-day vs 15-day FLC) | ⚠️ Not in current payload |
| `WorkflowID` | `plw-{policy_number}` | ⚠️ Not in current payload |

### 2.4 What PI currently sends (`PMCreatedSignal` — stub)

```go
// workflows/activities/pm_lifecycle_activities.go
type PMCreatedSignal struct {
    PolicyNumber string    `json:"policy_number"`
    PolicyType   string    `json:"policy_type"`
    IssuedAt     time.Time `json:"issued_at"`
}
```

This struct is missing `RequestID`, `PolicyID`, and the entire `PolicyMetadata`. PM's
`handlePolicyCreated()` handler at `workflows/policy_lifecycle_workflow.go:816` reads
`Metadata` fields to bootstrap the PLW's state and start the FLC timer. **PM will receive
the signal but will initialise with empty/zero metadata**, which will cause incorrect
behaviour (wrong FLC duration, missing premium data, zero CustomerID).

**This is the primary remaining gap.**

---

## 3. Current State of PI Service

### 3.1 What is now implemented

| Component | File | Status |
|-----------|------|--------|
| Step 10 in `PolicyIssuanceWorkflow` | `workflows/policy_issuance_workflow.go:525` | ✅ Added |
| `PMLifecycleActivities` struct | `workflows/activities/pm_lifecycle_activities.go` | ✅ Added |
| `StartPMLifecycleActivity` | same file | ✅ Added — calls `SignalWithStart` |
| `FindUnsignalledPoliciesActivity` | same file | ✅ Added |
| `PMSignalStore` interface | same file | ✅ Added — makes activities testable |
| `TemporalSignaller` interface | same file | ✅ Added — makes activities testable |
| `PMSignalReconciliationWorkflow` | `workflows/pm_reconciliation_workflow.go` | ✅ Added |
| Reconciliation Temporal schedule (every 15 min) | `bootstrap/bootstrapper.go:186` | ✅ Added |
| `MarkPMSignalSent` / `MarkPMSignalFailed` / `IncrementPMSignalAttempts` | `repo/postgres/proposal_repository.go` | ✅ Added |
| `FindUnsignalledPolicies` | same file | ✅ Added |
| Migration 003: 5 tracking columns + partial index on `proposal_issuance` | `db/migrations/003_pm_signal_tracking.sql` | ✅ Added |
| Activity unit tests (9 cases) | `workflows/activities/pm_lifecycle_activities_test.go` | ✅ Added |
| Workflow unit tests (6 cases) | `workflows/pm_reconciliation_workflow_test.go` | ✅ Added |

### 3.2 What is still missing

| Missing item | File | Impact |
|-------------|------|--------|
| `pm_contract.go` — PM's shared types (`PolicyCreatedSignal`, `PolicyLifecycleState`, `PolicyMetadata`) | `workflows/pm_contract.go` (new) | PM receives empty metadata; PLW bootstraps with zeros |
| Full payload construction in `StartPMLifecycleActivity` | `workflows/activities/pm_lifecycle_activities.go` | Same |
| `constructPMInitialState()` helper | `workflows/policy_issuance_workflow.go` | PLW initial state incomplete |
| `CreatePolicyIssuanceActivity` — return `MaturityDate` + `BillingMethod` | `workflows/activities/proposal_activities.go` | Fields unavailable for payload construction |
| Step 10 in `InstantIssuanceWorkflow` | `workflows/instant_issuance_workflow.go` | Aadhaar-based instant issue path still never signals PM |
| `pm.task_queue` in `configs/config.yaml` | `configs/config.yaml` | Config currently uses code default `"policy-manager-queue"` — must match `"policy-management-tq"` |

---

## 4. What Was Implemented — Detail

### 4.1 Migration 003 — `proposal_issuance` tracking columns

```sql
ALTER TABLE policy_issue.proposal_issuance
    ADD COLUMN pm_signal_status   VARCHAR(20) NOT NULL DEFAULT 'PENDING'
        CHECK (pm_signal_status IN ('PENDING', 'SENT', 'FAILED')),
    ADD COLUMN pm_signal_sent_at  TIMESTAMP WITH TIME ZONE,
    ADD COLUMN pm_signal_attempts INT NOT NULL DEFAULT 0,
    ADD COLUMN pm_signal_last_error TEXT,
    ADD COLUMN pm_plw_workflow_id  VARCHAR(100);

CREATE INDEX idx_issuance_pm_signal_pending
    ON proposal_issuance (policy_issue_date)
    WHERE pm_signal_status IN ('PENDING', 'FAILED');
```

| Status value | Meaning |
|---|---|
| `PENDING` | Default — policy issued, PM signal not yet confirmed |
| `SENT` | `SignalWithStart` returned successfully; PLW is live in PM |
| `FAILED` | Activity exhausted retries; reconciliation worker will retry |

### 4.2 `StartPMLifecycleActivity` — signal + DB tracking flow

```
Call activity
  │
  ├── IncrementPMSignalAttempts()   ← bumps counter regardless of outcome
  │
  ├── SignalWithStart()
  │     ├── SUCCESS → MarkPMSignalSent()   → pm_signal_status = 'SENT'
  │     └── FAILURE → MarkPMSignalFailed() → pm_signal_status = 'FAILED'
  │                   return error         → Temporal retries activity
  │
  └── DB write failure after success → return error (idempotent retry is safe)
```

Step 10 in `PolicyIssuanceWorkflow` treats a PM signal failure as **non-fatal** — the
workflow still returns `ISSUED`. The reconciliation worker is the recovery path. This is
deliberate: policy issuance itself succeeded; PM connection is a separate concern.

### 4.3 PM Signal Reconciliation Workflow

A Temporal schedule (`pm-signal-reconciliation-schedule`, cron `*/15 * * * *`) triggers
`PMSignalReconciliationWorkflow` every 15 minutes:

```
FindUnsignalledPoliciesActivity
    Query: pm_signal_status IN ('PENDING', 'FAILED')
           AND policy_issue_date < NOW() - 30 min  (grace period for in-flight workflows)
           AND pm_signal_attempts < 20
    Returns: up to 100 policies

For each policy → StartPMLifecycleActivity
    Per-policy failures logged; loop continues
    Overlap policy: SKIP (if previous run still in progress)
```

The 30-minute grace period prevents the reconciliation worker from interfering with a
`PolicyIssuanceWorkflow` that is still running and hasn't completed Step 10 yet.

---

## 5. What Still Needs to Be Done — Detailed Instructions

### 5.1 Create `workflows/pm_contract.go` (new file)

Copy the 4 type definitions from PM service. These must be byte-for-byte identical to PM's
structs so that Temporal's JSON serialisation produces matching shapes.

Source in PM: `policy-management-development/workflows/signals.go` (lines ~179–250)

```go
// workflows/pm_contract.go
package workflows

import "time"

// PolicyCreatedSignal is the signal sent to PM's PolicyLifecycleWorkflow on policy birth.
// Must match PM's definition in signals.go exactly.
type PolicyCreatedSignal struct {
    RequestID    string         `json:"request_id"`
    PolicyID     string         `json:"policy_id"`
    PolicyNumber string         `json:"policy_number"`
    Metadata     PolicyMetadata `json:"metadata"`
}

// PolicyLifecycleState is the initial workflow input for PM's PolicyLifecycleWorkflow.
// Must match PM's definition exactly.
type PolicyLifecycleState struct {
    PolicyNumber       string                  `json:"policy_number"`
    PolicyID           string                  `json:"policy_id"`
    CurrentStatus      string                  `json:"current_status"`
    PreviousStatus     string                  `json:"previous_status"`
    Encumbrances       EncumbranceFlags        `json:"encumbrances"`
    DisplayStatus      string                  `json:"display_status"`
    Version            int                     `json:"version"`
    Metadata           PolicyMetadata          `json:"metadata"`
    PendingRequests    []PendingRequest        `json:"pending_requests"`
    ProcessedSignalIDs map[string]time.Time    `json:"processed_signal_ids"`
    EventCount         int                     `json:"event_count"`
}

// PolicyMetadata holds the full policy data PM bootstraps from at birth.
type PolicyMetadata struct {
    CustomerID          int64     `json:"customer_id"`
    ProductCode         string    `json:"product_code"`
    ProductType         string    `json:"product_type"`         // "PLI" or "RPLI"
    SumAssured          float64   `json:"sum_assured"`
    CurrentPremium      float64   `json:"current_premium"`
    PremiumMode         string    `json:"premium_mode"`
    BillingMethod       string    `json:"billing_method"`       // "CASH" or "PAY_RECOVERY"
    IssueDate           time.Time `json:"issue_date"`
    MaturityDate        time.Time `json:"maturity_date"`
    PaidToDate          time.Time `json:"paid_to_date"`
    AgentID             *int64    `json:"agent_id,omitempty"`
    NominationStatus    string    `json:"nomination_status"`    // "PENDING" at birth
    IsDistanceMarketing bool      `json:"is_distance_marketing"` // true → 30-day FLC
    WorkflowID          string    `json:"workflow_id"`
}

// EncumbranceFlags matches PM's type.
type EncumbranceFlags struct {
    AssignmentType string `json:"assignment_type"` // "NONE" at birth
    HasLoan        bool   `json:"has_loan"`
    HasNomination  bool   `json:"has_nomination"`
}

// PendingRequest matches PM's type.
type PendingRequest struct {
    RequestID   string    `json:"request_id"`
    RequestType string    `json:"request_type"`
    ReceivedAt  time.Time `json:"received_at"`
}
```

> **Important:** Verify these struct field names and JSON tags against the actual PM source
> before committing. Any mismatch will cause silent zero-values in PM's state.

### 5.2 Update `StartPMLifecycleInput` and `StartPMLifecycleActivity`

**File:** `workflows/activities/pm_lifecycle_activities.go`

Replace the current `StartPMLifecycleInput` and `PMCreatedSignal` stubs with the full
types from `pm_contract.go`:

```go
// StartPMLifecycleInput carries the full PM contract types needed for SignalWithStart.
type StartPMLifecycleInput struct {
    PolicyNumber string                    `json:"policy_number"`
    InitialState workflows.PolicyLifecycleState `json:"initial_state"`
    Signal       workflows.PolicyCreatedSignal  `json:"signal"`
}
```

In `StartPMLifecycleActivity`, change the `SignalWithStartWorkflow` call to:

```go
_, err := a.signaller.SignalWithStartWorkflow(
    ctx,
    workflowID,
    "policy-created",
    input.Signal,       // full PolicyCreatedSignal with RequestID + PolicyID + Metadata
    startOpts,
    "PolicyLifecycleWorkflow",
    input.InitialState, // full PolicyLifecycleState
)
```

### 5.3 Update `CreatePolicyIssuanceActivity` to return result

**File:** `workflows/activities/proposal_activities.go`

`CreatePolicyIssuanceActivity` currently returns `error` only. It needs to return a result
struct so the workflow can pass `MaturityDate` and `BillingMethod` to Step 10.

```go
// Add this result type:
type CreatePolicyIssuanceResult struct {
    MaturityDate  time.Time `json:"maturity_date"`
    BillingMethod string    `json:"billing_method"` // "CASH" or "PAY_RECOVERY"
}

// Change signature:
func (a *ProposalActivities) CreatePolicyIssuanceActivity(
    ctx context.Context, input CreatePolicyIssuanceInput,
) (*CreatePolicyIssuanceResult, error) {
    // ... existing logic ...
    // Read billing_method from proposal record (already in DB)
    // Return:
    return &CreatePolicyIssuanceResult{
        MaturityDate:  maturityDate,
        BillingMethod: proposal.BillingMethod, // read from proposals table
    }, nil
}
```

Also update Step 7b in `PolicyIssuanceWorkflow` to capture the result:

```go
var issuanceResult activities.CreatePolicyIssuanceResult
if err := workflow.ExecuteActivity(shortActivityOpts,
    "CreatePolicyIssuanceActivity", issuanceInput).Get(ctx, &issuanceResult); err != nil {
    // ... error handling
}
```

### 5.4 Add `constructPMInitialState()` to `PolicyIssuanceWorkflow`

**File:** `workflows/policy_issuance_workflow.go`

Add this helper function (uses data already accumulated in the workflow by Step 10):

```go
func constructPMPayload(
    ctx workflow.Context,
    input        PolicyIssuanceInput,
    pnResult     activities.GeneratePolicyNumberResult,
    premResult   activities.CalculatePremiumResult,
    issuResult   activities.CreatePolicyIssuanceResult,
) (PolicyLifecycleState, PolicyCreatedSignal) {

    workflowID := fmt.Sprintf("plw-%s", pnResult.PolicyNumber)
    now := workflow.Now(ctx).UTC()

    // Parse CustomerID from string to int64
    customerID, _ := strconv.ParseInt(input.CustomerID, 10, 64)

    meta := PolicyMetadata{
        CustomerID:          customerID,
        ProductCode:         input.ProductCode,
        ProductType:         string(input.PolicyType),
        SumAssured:          input.SumAssured,
        CurrentPremium:      premResult.TotalPayable,
        PremiumMode:         string(input.PremiumPaymentFrequency),
        BillingMethod:       issuResult.BillingMethod,
        IssueDate:           now,
        MaturityDate:        issuResult.MaturityDate,
        PaidToDate:          now,
        NominationStatus:    "PENDING",
        IsDistanceMarketing: false, // TODO: read from product config
        WorkflowID:          workflowID,
    }

    state := PolicyLifecycleState{
        PolicyNumber:       pnResult.PolicyNumber,
        PolicyID:           input.ProposalNumber,
        CurrentStatus:      "FREE_LOOK_ACTIVE",
        Encumbrances:       EncumbranceFlags{AssignmentType: "NONE"},
        DisplayStatus:      "FREE_LOOK_ACTIVE",
        Version:            1,
        Metadata:           meta,
        PendingRequests:    []PendingRequest{},
        ProcessedSignalIDs: map[string]time.Time{},
    }

    signal := PolicyCreatedSignal{
        RequestID:    workflow.GetInfo(ctx).WorkflowExecution.ID, // dedup key
        PolicyID:     input.ProposalNumber,
        PolicyNumber: pnResult.PolicyNumber,
        Metadata:     meta,
    }

    return state, signal
}
```

Update Step 10 in the workflow to use it:

```go
// Step 10: Signal PM service to start lifecycle workflow
state, signal := constructPMPayload(ctx, input, policyNumberResult, premiumResult, issuanceResult)
pmSignalInput := activities.StartPMLifecycleInput{
    PolicyNumber: policyNumberResult.PolicyNumber,
    InitialState: state,
    Signal:       signal,
}
if err := workflow.ExecuteActivity(externalCallOpts,
    "StartPMLifecycleActivity", pmSignalInput).Get(ctx, nil); err != nil {
    logger.Error("PM lifecycle signal failed — reconciliation will retry", "error", err)
}
```

### 5.5 Add Step 10 to `InstantIssuanceWorkflow`

**File:** `workflows/instant_issuance_workflow.go`

The Aadhaar-based instant issuance path has the same gap. After the policy is activated in
`InstantIssuanceWorkflow`, add the same Step 10 pattern using the same
`StartPMLifecycleActivity`. The workflow already has access to the relevant fields
(`PolicyNumber`, `PolicyType`, premium result etc.) — assemble the payload the same way.

### 5.6 Update `configs/config.yaml`

**File:** `configs/config.yaml`

Add the PM task queue under the `temporal` section:

```yaml
temporal:
  host: "localhost"
  port: "7233"
  namespace: "default"
  pm_task_queue: "policy-management-tq"   # PM service task queue — MUST match PM worker registration
```

The bootstrapper reads `pm.task_queue` (note: different key path — verify consistency with
the config library's key naming convention used elsewhere in the project).

---

## 6. Shared Contract — Type Strategy

PM's `PolicyLifecycleState`, `PolicyCreatedSignal`, and `PolicyMetadata` types are defined
in `policy-management-development/.../workflows/signals.go`.

| Option | Pros | Cons | Decision |
|--------|------|------|---------|
| **A — Duplicate in PI (`pm_contract.go`)** | Zero cross-service build dependency; follows surrender service pattern | Manual sync on PM type changes | ✅ **Use this** |
| B — Shared Go module | Single source of truth | New module overhead; CI changes | ✗ |

**Process:** When PM changes any field in `PolicyCreatedSignal` or `PolicyMetadata`,
PI team must be notified and `pm_contract.go` must be updated in the same sprint.
Add a comment block in `pm_contract.go` with the PM source file path and last-verified date.

---

## 7. Implementation Checklist

| # | Task | File | Status |
|---|------|------|--------|
| 1 | Migration 003: 5 tracking columns + index | `db/migrations/003_pm_signal_tracking.sql` | ✅ Done |
| 2 | `MarkPMSignalSent` / `MarkPMSignalFailed` / `IncrementPMSignalAttempts` / `FindUnsignalledPolicies` repo methods | `repo/postgres/proposal_repository.go` | ✅ Done |
| 3 | `PMLifecycleActivities` struct + `PMSignalStore` + `TemporalSignaller` interfaces | `workflows/activities/pm_lifecycle_activities.go` | ✅ Done |
| 4 | `StartPMLifecycleActivity` (signal + DB tracking) | same file | ✅ Done — payload stub |
| 5 | `FindUnsignalledPoliciesActivity` | same file | ✅ Done |
| 6 | `PMSignalReconciliationWorkflow` | `workflows/pm_reconciliation_workflow.go` | ✅ Done |
| 7 | Bootstrapper: provide + register + Temporal schedule | `bootstrap/bootstrapper.go` | ✅ Done |
| 8 | Step 10 in `PolicyIssuanceWorkflow` | `workflows/policy_issuance_workflow.go` | ✅ Done — simplified payload |
| 9 | Activity + workflow unit tests (15 cases) | `*_test.go` files | ✅ Done |
| 10 | `pm_contract.go` — copy PM types | `workflows/pm_contract.go` | ❌ TODO |
| 11 | Update `StartPMLifecycleInput` + activity to use full PM types | `workflows/activities/pm_lifecycle_activities.go` | ❌ TODO |
| 12 | `CreatePolicyIssuanceActivity` — return `MaturityDate` + `BillingMethod` | `workflows/activities/proposal_activities.go` | ❌ TODO |
| 13 | `constructPMPayload()` helper in workflow | `workflows/policy_issuance_workflow.go` | ❌ TODO |
| 14 | Step 10 in `InstantIssuanceWorkflow` | `workflows/instant_issuance_workflow.go` | ❌ TODO |
| 15 | `pm_task_queue` in `configs/config.yaml` | `configs/config.yaml` | ❌ TODO |

**Items 10–15 must be completed before production use. Items 1–9 are merged to branch
`claude/setup-dual-projects-9Qi9t`.**

---

## 8. PM Service Readiness Confirmation (Unchanged)

The PM service (`policy-management-tq`) is ready to receive the signal:

| Check | Status |
|-------|--------|
| `PolicyLifecycleWorkflow` registered on `policy-management-tq` | ✅ `bootstrapper.go:145` |
| `SignalPolicyCreated = "policy-created"` constant defined | ✅ `workflows/signals.go:21` |
| `PolicyCreatedSignal` struct defined | ✅ `workflows/signals.go:179` |
| `handlePolicyCreated()` signal handler implemented | ✅ `workflows/policy_lifecycle_workflow.go:816` |
| FLC timer goroutine spawned on receipt | ✅ Per req §10.1.6 |
| `RecordStateTransitionActivity` called on receipt | ✅ Persists `FREE_LOOK_ACTIVE` to DB |
| Duplicate signal detection (dedup by `RequestID`) | ✅ `isDuplicate()` check |

**Nothing to change in PM service. Waiting on PI service to send the correct payload.**

---

## 9. Sequence Diagram — Current State + Reconciliation

```
PolicyIssuanceWorkflow (PI)
├── Step 1:  ValidateProposalActivity
├── Step 2:  CheckEligibilityActivity
├── Step 3:  CalculatePremiumActivity        ← premiumResult accumulated here
├── Step 3b: SavePremiumToProposalActivity
├── Step 4:  [QC Signal loop]
├── Step 5:  RequestMedicalReviewActivity + [Medical Signal]
├── Step 6:  RouteToApproverActivity + [Approver Signal]
├── Step 7:  GeneratePolicyNumberActivity
├── Step 7b: CreatePolicyIssuanceActivity    ← TODO: return MaturityDate + BillingMethod
├── Step 8:  GenerateBondActivity + UpdateBondDetailsActivity
├── Step 9:  SendNotificationActivity
└── Step 10: StartPMLifecycleActivity        ✅ NOW PRESENT
              │  proposal_issuance.pm_signal_status → 'PENDING' (at row creation)
              │  IncrementPMSignalAttempts()
              │
              ├── SignalWithStartWorkflow(
              │     workflowID:  "plw-PLI/2026/000001"
              │     signal:      "policy-created"
              │     payload:     PMCreatedSignal{...}  ⚠️ stub — full payload TODO
              │     task_queue:  "policy-manager-queue" ⚠️ wrong — must be "policy-management-tq"
              │     workflow:    "PolicyLifecycleWorkflow"
              │     input:       PMCreatedSignal{...}   ⚠️ stub — full state TODO
              │   )
              │
              ├── SUCCESS → pm_signal_status = 'SENT'
              └── FAILURE → pm_signal_status = 'FAILED' → Temporal retries → reconciliation


PM Signal Reconciliation (every 15 min via Temporal Schedule)
├── FindUnsignalledPoliciesActivity
│     SELECT WHERE pm_signal_status IN ('PENDING','FAILED')
│       AND policy_issue_date < NOW() - 30min
│       AND pm_signal_attempts < 20
│     LIMIT 100
│
└── For each policy → StartPMLifecycleActivity
      (SignalWithStart is idempotent — safe to retry)


PolicyLifecycleWorkflow (PM — plw-PLI/2026/000001)   ← ready and waiting
├── Receives "policy-created" signal
├── Reads Metadata → TODO: currently receives empty metadata
├── Records FREE_LOOK_ACTIVE → DB
├── Starts FLC timer (15d or 30d based on IsDistanceMarketing)
└── Enters signal-select loop
    (all future policy requests now routable once payload is correct)
```

---

## 10. Risk Assessment

| Risk | Severity | Status |
|------|----------|--------|
| PLW never spawned (no Step 10 at all) | CRITICAL | ✅ Resolved — Step 10 now present |
| Signal silently dropped on transient failure | HIGH | ✅ Resolved — DB tracking + reconciliation |
| PLW bootstraps with empty/zero `PolicyMetadata` | HIGH | ⚠️ Active — payload stub incomplete |
| Wrong FLC duration (0 days because `IsDistanceMarketing` is false/zero) | HIGH | ⚠️ Active — `IsDistanceMarketing` not set |
| PM dedup fails (`RequestID` is empty string) | MEDIUM | ⚠️ Active — `RequestID` not set |
| `InstantIssuanceWorkflow` still has no Step 10 | HIGH | ⚠️ Active — not yet implemented |
| Config key `pm.task_queue` points to wrong queue name | HIGH | ⚠️ Active — YAML not updated |
| Policy stuck in `FREE_LOOK_ACTIVE` forever (if `IsDistanceMarketing` not set) | HIGH | ⚠️ Active — FLC timer depends on this flag |

---

## 11. Onboarding — How to Pick Up the Remaining Work

**Recommended order for the PI team to complete items 10–15:**

1. **Start with §5.1** — create `pm_contract.go`. This is purely copying types with no
   logic. Verify each field against PM's `signals.go` side-by-side.

2. **Then §5.3** — update `CreatePolicyIssuanceActivity` to return `MaturityDate` +
   `BillingMethod`. Read `billing_method` from the `proposals` table (it's already in the
   schema). Update Step 7b in the workflow to capture the result struct.

3. **Then §5.4** — add `constructPMPayload()` and update Step 10. At this point the full
   signal will be sent.

4. **Then §5.2** — update `StartPMLifecycleInput` to use the types from `pm_contract.go`.
   Update the activity tests to use the full struct.

5. **Then §5.5** — add Step 10 to `InstantIssuanceWorkflow` using the same helper.

6. **Finally §5.6** — update `configs/config.yaml` and verify the config key path used
   in the bootstrapper (`pm.task_queue` vs `temporal.pm_task_queue`).

**Integration smoke test** (after all 6 steps):
1. Issue a test policy through the full `PolicyIssuanceWorkflow`
2. Verify `proposal_issuance.pm_signal_status = 'SENT'`
3. Verify `plw-{policy_number}` workflow exists in PM's Temporal namespace
4. Verify PM's `proposal_issuance` table has a `FREE_LOOK_ACTIVE` row for the policy
5. Verify the FLC timer is scheduled (check PM workflow history for `StartTimer`)
6. Repeat steps 1–5 for an instant-issue (Aadhaar) policy
