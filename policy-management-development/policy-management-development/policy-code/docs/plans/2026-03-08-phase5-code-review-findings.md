# Phase 5 Code Review — Findings Status
**Date:** 2026-03-08
**Source:** External code review of `workflows/policy_lifecycle_workflow.go`

---

## Summary

| # | Severity | Finding | Status |
|---|----------|---------|--------|
| 1 | **CRITICAL** | Non-deterministic clock in `pruneProcessedSignals()` | ❌ OPEN |
| 2 | **CRITICAL** | Child workflows terminated on parent CAN/close | ❌ OPEN |
| 3 | **CRITICAL** | Admin-void leaves DB financial lock row | ❌ OPEN |
| 4 | **HIGH** | Config-driven timeouts/cooling hardcoded | ❌ OPEN |
| 5 | **HIGH** | `handleOperationCompleted()` lacks dedup + audit log | ❌ OPEN |
| 6 | **HIGH** | System signals missing `LogSignalReceivedActivity` audit | ❌ OPEN |
| 7 | **HIGH** | Unbounded metadata forwarded from NFR completion | ❌ OPEN |
| 8 | **MEDIUM** | Distance-marketing FLC period not applied | ❌ OPEN |
| 9 | **MEDIUM** | Signal audit channel logs request type not signal name | ❌ OPEN |
| 10 | **LOW** | Forced-surrender uses sum_assured proxy for GSV | ⚠️ ACKNOWLEDGED |

---

## Detailed Findings

### Finding 1 — Non-deterministic clock in `pruneProcessedSignals()` ❌ CRITICAL

**Location:** `workflows/policy_lifecycle_workflow.go:156-163`

**Evidence:**
```go
func pruneProcessedSignals(state *PolicyLifecycleState) {
    cutoff := time.Now().Add(-90 * 24 * time.Hour) // ← Go wall-clock inside workflow!
    for k, t := range state.ProcessedSignalIDs {
        if t.Before(cutoff) {
            delete(state.ProcessedSignalIDs, k)
        }
    }
}
```

**Impact:** `time.Now()` returns the real wall-clock time on each replay. During
Temporal workflow replay (after worker restart, CAN, or history replay for debugging),
this produces a different `cutoff` than the original execution, potentially deleting
different dedup entries. This is a **replay divergence** — the workflow's state
diverges from its history, violating Temporal's determinism contract. Non-determinism
errors surface as `go.temporal.io/sdk/internal.ErrWorkflowResultMismatch` and can
corrupt workflow state.

**Fix:** Pass `ctx workflow.Context` to `pruneProcessedSignals` and use
`workflow.Now(ctx)` instead of `time.Now()`.

```go
// After
func pruneProcessedSignals(ctx workflow.Context, state *PolicyLifecycleState) {
    cutoff := workflow.Now(ctx).Add(-90 * 24 * time.Hour)
    for k, t := range state.ProcessedSignalIDs {
        if t.Before(cutoff) {
            delete(state.ProcessedSignalIDs, k)
        }
    }
}
// Call site (line 650):
pruneProcessedSignals(ctx, &state)
```

---

### Finding 2 — Child workflows terminated on parent CAN/close ❌ CRITICAL

**Location:** `workflows/policy_lifecycle_workflow.go:119-127`

**Evidence:**
```go
func childWFCtx(ctx workflow.Context, taskQueue, childID string) workflow.Context {
    return workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
        TaskQueue:  taskQueue,
        WorkflowID: childID,
        RetryPolicy: &temporal.RetryPolicy{
            MaximumAttempts: 1,
        },
        // ← ParentClosePolicy missing! Defaults to TERMINATE
    })
}
```

**Impact:** When the PLW does Continue-As-New (CAN), Temporal sends a `TERMINATE`
signal to all child workflows using the default `ParentClosePolicy`. This means
active downstream workflows (surrender, loan, etc.) are killed mid-execution when PLW
crosses the event threshold. The downstream services never complete, financial locks
are never released, and `service_request` rows remain stuck in `IN_PROGRESS`.

**Fix:** Set `ParentClosePolicy: temporal.ParentClosePolicyAbandon` so child workflows
continue independently after CAN:

```go
func childWFCtx(ctx workflow.Context, taskQueue, childID string) workflow.Context {
    return workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
        TaskQueue:         taskQueue,
        WorkflowID:        childID,
        ParentClosePolicy: temporal.ParentClosePolicyAbandon, // NEW
        RetryPolicy: &temporal.RetryPolicy{
            MaximumAttempts: 1,
        },
    })
}
```

---

### Finding 3 — Admin-void leaves DB financial lock row ❌ CRITICAL

**Location:** `workflows/policy_lifecycle_workflow.go:1365-1381`

**Evidence:**
```go
func handleAdminVoid(ctx workflow.Context, state *PolicyLifecycleState, sig AdminVoidSignal) bool {
    // ...
    state.PendingRequests = nil
    state.ActiveLock = nil   // ← clears in-memory lock only!
    doTransition(...)
    // ← No ReleaseFinancialLockActivity call here!
}
```

**Compare with `handleWithdrawal()` (correct pattern):**
```go
if state.ActiveLock != nil && state.ActiveLock.RequestID == sig.TargetRequestID {
    state.ActiveLock = nil
    _ = workflow.ExecuteActivity(shortActCtx(ctx),
        policyActs.ReleaseFinancialLockActivity, state.PolicyDBID).Get(ctx, nil) // ← present
}
```

**Impact:** After admin-void, the `policy_financial_lock` DB row remains. If the policy
is later reopened, the stale lock row prevents any new financial requests [BR-PM-030].

**Fix:** Add `ReleaseFinancialLockActivity` call in `handleAdminVoid()` when a lock exists:

```go
// After cancelling pending requests, release DB lock if one was held
if state.ActiveLock != nil {
    _ = workflow.ExecuteActivity(shortActCtx(ctx),
        policyActs.ReleaseFinancialLockActivity, state.PolicyDBID).Get(ctx, nil)
}
state.PendingRequests = nil
state.ActiveLock = nil
```

---

### Finding 4 — Config-driven timeouts/cooling hardcoded ❌ HIGH

**Location:** `workflows/policy_lifecycle_workflow.go:52-68` (routing), `:85-99` (cooling)

**Evidence:**
```go
func routingTimeoutForRequest(requestType string) time.Duration {
    switch requestType {
    case domain.RequestTypeSurrender: return 7 * 24 * time.Hour // hardcoded!
    case domain.RequestTypeLoan:      return 3 * 24 * time.Hour // hardcoded!
    // ...
    }
}

func coolingDuration(terminalStatus string) time.Duration {
    switch terminalStatus {
    case domain.StatusDeathClaimSettled: return 180 * 24 * time.Hour // hardcoded!
    // ...
    }
}
```

Both functions have comments acknowledging the issue (`[Review-Fix-16]`), and all config
keys exist in `domain/policy_state_config.go` (e.g. `ConfigKeyRoutingTimeoutSurrender`,
`ConfigKeyCoolingPeriodVoid`, etc.).

**Impact:** Admins cannot adjust timeouts without a code deploy. For cooling periods,
regulatory changes (e.g. death claim settlement window) cannot be applied without
re-deploying.

**Fix:** Fetch config values via `FetchWorkflowConfigActivity` at workflow startup and
cache in `state.CachedConfig`. Routing timeout config should be fetched per-request
when routing. Use hardcoded values as in-code fallback.

```go
// Routing timeout — read from cached config
func routingTimeoutForRequestFromConfig(cachedConfig map[string]string, requestType string) time.Duration {
    configKey := routingTimeoutConfigKey(requestType) // maps type → ConfigKeyRoutingTimeout*
    if v, ok := cachedConfig[configKey]; ok {
        if d, err := time.ParseDuration(v); err == nil {
            return d
        }
    }
    return routingTimeoutForRequest(requestType) // fallback to hardcoded
}
```

---

### Finding 5 — `handleOperationCompleted()` lacks dedup + audit log ❌ HIGH

**Location:** `workflows/policy_lifecycle_workflow.go:1018-1070`

**Evidence:**
```go
func handleOperationCompleted(ctx workflow.Context, state *PolicyLifecycleState, sig OperationCompletedSignal) bool {
    // ← No ProcessedSignalIDs dedup check!
    // ← No LogSignalReceivedActivity call!
    var matched *PendingRequest
    // ... proceeds to update service_request and state transition
}
```

**Compare with `handlePremiumPaid()` (correct pattern):**
```go
func handlePremiumPaid(...) {
    if _, seen := state.ProcessedSignalIDs[sig.RequestID]; seen { return } // dedup ✓
    // ... state update
    state.ProcessedSignalIDs[sig.RequestID] = workflow.Now(ctx) // mark ✓
}
```

**Impact:** If a downstream service retries the `operation-completed` signal (e.g. due
to gRPC timeout), the PLW processes the completion twice: updates `service_request` to
COMPLETED again, potentially fires a second state transition, releases the financial
lock twice.

**Fix:** Add dedup check + `LogSignalReceivedActivity` at the start of
`handleOperationCompleted()`, and mark as processed at the end.

---

### Finding 6 — System signals missing `LogSignalReceivedActivity` ❌ HIGH

**Location:** `handlePremiumPaid()`, `handlePaymentDishonored()`, `handleAMLFlagRaised()`,
`handleAMLFlagCleared()`, `handleInvestigationStarted()`, `handleInvestigationConcluded()`,
`handleLoanBalanceUpdated()`, `handleConversionReversed()`, `handleCustomerIDMerge()`,
`handleDisputeRegistered()`, `handleDisputeResolved()`

**Evidence:** All system signal handlers have dedup via `ProcessedSignalIDs` but none
call `policyActs.LogSignalReceivedActivity`. Only `handleFinancialRequest()` logs.

**Impact:** System/compliance signals (AML, payment, investigation) leave no audit trail
in `policy_signal_log` or `processed_signal_registry`. Compliance audits cannot trace
when and why a policy was suspended, lapsed, etc.

**Fix:** Add `LogSignalReceivedActivity` call at the start of each system signal handler,
after the dedup check.

---

### Finding 7 — Unbounded metadata forwarded from NFR completion ❌ HIGH

**Location:** `workflows/policy_lifecycle_workflow.go:1107-1119`

**Evidence:**
```go
if matched.RequestType == domain.RequestTypeAssignment && sig.Outcome == domain.RequestOutcomeApproved {
    if sig.OutcomePayload != nil {
        var payload map[string]interface{}
        if json.Unmarshal(sig.OutcomePayload, &payload) == nil {
            _ = workflow.ExecuteActivity(shortActCtx(ctx),
                policyActs.UpdatePolicyMetadataActivity,
                acts.MetadataUpdateParams{
                    PolicyID: state.PolicyDBID,
                    Updates:  payload, // ← arbitrary keys from downstream!
                }).Get(ctx, nil)
        }
    }
}
```

**Impact:** A rogue or buggy assignment service could send arbitrary keys in
`OutcomePayload` — potentially overwriting unrelated columns like `customer_id`,
`sum_assured`, or `product_type` via `UpdatePolicyMetadataActivity`.

**Fix:** Allow-list the permitted update keys for each NFR type:

```go
// Allow-list for assignment NFR outcome payload
var assignmentAllowedKeys = map[string]bool{
    "assignment_type":   true,
    "assignment_status": true,
    "assignee_name":     true,
}

func filterPayload(payload map[string]interface{}, allowed map[string]bool) map[string]interface{} {
    filtered := make(map[string]interface{}, len(allowed))
    for k, v := range payload {
        if allowed[k] {
            filtered[k] = v
        }
    }
    return filtered
}
```

---

### Finding 8 — Distance-marketing FLC period not applied ❌ MEDIUM

**Location:** `workflows/policy_lifecycle_workflow.go:731-737`

**Evidence:**
```go
var flcDaysStr string
_ = workflow.ExecuteActivity(shortActCtx(ctx),
    policyActs.FetchWorkflowConfigActivity,
    domain.ConfigKeyFLCPeriodDays).Get(ctx, &flcDaysStr) // ← always standard 15d
// IsDistanceMarketing flag never checked!
flcDays, _ := strconv.Atoi(flcDaysStr)
flcPeriod := getFLCPeriod(flcDays)
```

`domain.ConfigKeyFLCPeriodDistanceMarketing` exists but is not fetched. The
`sig.Metadata.IsDistanceMarketing` flag is available but ignored.

**Impact:** Distance-marketing policy holders are entitled to a 30-day FLC window by
regulation. Assigning only 15 days is a compliance failure.

**Fix:**
```go
configKey := domain.ConfigKeyFLCPeriodDays
if sig.Metadata.IsDistanceMarketing {
    configKey = domain.ConfigKeyFLCPeriodDistanceMarketing
}
_ = workflow.ExecuteActivity(..., policyActs.FetchWorkflowConfigActivity, configKey).Get(ctx, &flcDaysStr)
```

---

### Finding 9 — Signal audit channel logs request type not signal name ❌ MEDIUM

**Location:** `workflows/policy_lifecycle_workflow.go:773-779`

**Evidence:**
```go
_ = workflow.ExecuteActivity(shortActCtx(ctx),
    policyActs.LogSignalReceivedActivity,
    acts.SignalLogEntry{
        SignalChannel: sig.RequestType, // ← "SURRENDER" not "surrender-request"
```

**Impact:** The `policy_signal_log` table stores `"SURRENDER"` in `signal_channel`,
but the actual Temporal signal name is `"surrender-request"`. Cross-referencing audit
logs against Temporal history or other signal handlers is inconsistent.

**Fix:** Map `RequestType` → kebab-case signal name:
```go
SignalChannel: requestTypeToSignalName(sig.RequestType),
// where requestTypeToSignalName("SURRENDER") = "surrender-request" etc.
```

---

### Finding 10 — Forced-surrender uses sum_assured proxy for GSV ⚠️ LOW / ACKNOWLEDGED

**Location:** `workflows/activities/batch_activities.go:488-493`

**Evidence:**
```go
// (sum_assured used as proxy for 100% GSV — actual GSV computation is downstream)
// [BR-PM-074, Review-Fix-18]
Where(sq.Expr("loan_outstanding >= sum_assured * ?", loanRatioFraction))
```

The code already has a `Review-Fix-18` comment acknowledging this approximation. A real
GSV computation would require calling the loan service — this is an architectural
decision that requires downstream service contract work.

**Status:** ACKNOWLEDGED IN CODE. Not a silent bug. This is an explicit engineering
trade-off with a tracking tag. Recommend adding a TODO with a JIRA/issue reference and
leaving for the loan service integration phase.

---

## Action Required

Findings 1–9 are open bugs. Suggested grouping for implementation:

### Group A — Critical (fix immediately, before next deploy)
- Finding 1: Non-deterministic clock
- Finding 2: Missing `ParentClosePolicy: Abandon`
- Finding 3: Admin-void DB lock not released

### Group B — High (fix in next sprint)
- Finding 4: Config-driven timeouts/cooling
- Finding 5: `handleOperationCompleted` dedup + audit
- Finding 6: System signal audit logging
- Finding 7: Allow-list NFR metadata updates

### Group C — Medium (schedule)
- Finding 8: Distance-marketing FLC period
- Finding 9: Signal audit channel name

### Group D — Acknowledged
- Finding 10: Forced-surrender GSV proxy (tracked, not urgent)
