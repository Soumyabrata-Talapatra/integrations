# Workstream B: Phase 5 Workflow Fixes — Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Fix 9 bugs found in `workflows/policy_lifecycle_workflow.go` during Phase 5 code review. Bugs are grouped by severity: 3 critical (fix first), 4 high, 2 medium.

**Architecture:** All fixes are in the PLW workflow and its activity helpers. The PLW is a long-running Temporal workflow — non-determinism errors cause replay failures. Some fixes require passing `ctx workflow.Context` through helper functions. No DB schema changes. No API changes.

**Tech Stack:** Go 1.26, Temporal SDK v1.37.0, `go.temporal.io/sdk/workflow`, `go.temporal.io/sdk/temporal`

---

## Context for the Engineer

### Temporal determinism rule
All workflow code must be **deterministic** — same input (signals, timers) must produce same output every replay. Never use:
- `time.Now()` — use `workflow.Now(ctx)` instead
- `math/rand` — use `workflow.Now(ctx).UnixNano()` as seed if needed
- `os`, `net/http`, any I/O — use activities instead

### Key types
```go
// workflow.Context  — Temporal's deterministic context (not standard context.Context)
// workflow.Now(ctx) — deterministic clock; use instead of time.Now()
// temporal.ParentClosePolicyAbandon — keeps child workflows alive after parent CAN
```

### File being modified
`workflows/policy_lifecycle_workflow.go`

### Activity available for releasing financial lock
```go
// policyActs.ReleaseFinancialLockActivity(ctx, policyDBID int64)
// Defined in: workflows/activities/policy_activities.go:700
```

---

## Group A — Critical (fix in this order)

---

## Task 1: Fix non-deterministic clock in `pruneProcessedSignals()`

**File:** `workflows/policy_lifecycle_workflow.go`

**The bug (line 156-163):**
```go
func pruneProcessedSignals(state *PolicyLifecycleState) {
    cutoff := time.Now().Add(-90 * 24 * time.Hour) // ← WRONG: real wall-clock
    for k, t := range state.ProcessedSignalIDs {
        if t.Before(cutoff) {
            delete(state.ProcessedSignalIDs, k)
        }
    }
}
// Called at line 650: pruneProcessedSignals(&state)
```

**Why this breaks replay:** On workflow replay after a worker restart, `time.Now()` returns a different value than the original execution. Entries get deleted from `ProcessedSignalIDs` at different times on replay vs. original, causing `ErrWorkflowResultMismatch`.

**Step 1: Change the function signature and body**

Find:
```go
func pruneProcessedSignals(state *PolicyLifecycleState) {
	cutoff := time.Now().Add(-90 * 24 * time.Hour)
```

Replace with:
```go
func pruneProcessedSignals(ctx workflow.Context, state *PolicyLifecycleState) {
	cutoff := workflow.Now(ctx).Add(-90 * 24 * time.Hour)
```

**Step 2: Update the call site (line 650)**

Find:
```go
		pruneProcessedSignals(&state)
```

Replace with:
```go
		pruneProcessedSignals(ctx, &state)
```

**Step 3: Build check**
```bash
cd /d/policy-manage/policy-management
go build ./...
```
Expected: exit 0. If it complains `time` package unused, check if `time` is still used elsewhere in the file (it is — in `canTimeThreshold`, etc.). No import changes needed.

**Step 4: Commit**
```bash
git add workflows/policy_lifecycle_workflow.go
git commit -m "fix: replace time.Now() with workflow.Now(ctx) in pruneProcessedSignals

Using Go wall-clock inside workflow code causes replay divergence:
on worker restart, different entries get pruned from ProcessedSignalIDs,
violating Temporal's determinism contract (ErrWorkflowResultMismatch).

workflow.Now(ctx) returns the deterministic time from workflow history."
```

---

## Task 2: Add `ParentClosePolicy: Abandon` to child workflow options

**File:** `workflows/policy_lifecycle_workflow.go`

**The bug (line 119-127):**
```go
func childWFCtx(ctx workflow.Context, taskQueue, childID string) workflow.Context {
	return workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
		TaskQueue:  taskQueue,
		WorkflowID: childID,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 1,
		},
		// ← ParentClosePolicy missing; defaults to TERMINATE
	})
}
```

**Why this breaks:** When PLW reaches Continue-As-New threshold (~40k events), Temporal
creates a new workflow execution and sends TERMINATE to all children using default policy.
Active surrender/loan/claim child workflows are killed mid-execution.

**Step 1: Add the ParentClosePolicy field**

Find the `childWFCtx` function and replace the return statement:
```go
// Before
	return workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
		TaskQueue:  taskQueue,
		WorkflowID: childID,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 1, // No automatic retry for business workflows
		},
	})

// After
	return workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
		TaskQueue:         taskQueue,
		WorkflowID:        childID,
		ParentClosePolicy: temporal.ParentClosePolicyAbandon, // keep children alive on CAN
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 1, // No automatic retry for business workflows
		},
	})
```

**Step 2: Verify `temporal` package is already imported**

Check the import block at the top of the file:
```go
import (
	...
	"go.temporal.io/sdk/temporal"   // ← should already be here
	"go.temporal.io/sdk/workflow"
	...
)
```
If missing, add `"go.temporal.io/sdk/temporal"` to the import block.

**Step 3: Build check**
```bash
go build ./...
```
Expected: exit 0.

**Step 4: Commit**
```bash
git add workflows/policy_lifecycle_workflow.go
git commit -m "fix: set ParentClosePolicy=Abandon for all child workflows

Default ParentClosePolicy is TERMINATE, which kills active downstream
workflows (surrender, loan, claim) when PLW does Continue-As-New.
With Abandon, children continue to completion independently after CAN."
```

---

## Task 3: Fix admin-void DB lock not released

**File:** `workflows/policy_lifecycle_workflow.go`

**The bug (line 1365-1381):**
```go
func handleAdminVoid(ctx workflow.Context, state *PolicyLifecycleState, sig AdminVoidSignal) bool {
	// ...
	for _, pr := range state.PendingRequests {
		_ = workflow.ExecuteActivity(shortActCtx(ctx),
			policyActs.CancelDownstreamWorkflowActivity, ...).Get(ctx, nil)
	}
	state.PendingRequests = nil
	state.ActiveLock = nil   // ← clears in-memory only! DB row remains.
	doTransition(...)
```

**Why this breaks:** After admin-void, the `policy_financial_lock` DB row persists.
When the policy is reopened, any new financial request hits `CheckFinancialLock()`,
finds the stale row, and returns HTTP 409 forever.

**Step 1: Add `ReleaseFinancialLockActivity` call in `handleAdminVoid`**

Find the body of `handleAdminVoid`. After the loop that cancels pending workflows, before `state.PendingRequests = nil`, insert:

```go
	// Release DB financial lock if one was held — in-memory clear alone leaves a stale DB row
	// which blocks all future financial requests after reopen [BR-PM-030, BR-PM-073]
	if state.ActiveLock != nil {
		_ = workflow.ExecuteActivity(shortActCtx(ctx),
			policyActs.ReleaseFinancialLockActivity, state.PolicyDBID).Get(ctx, nil)
	}
	state.PendingRequests = nil
	state.ActiveLock = nil
```

The full corrected function body looks like:
```go
func handleAdminVoid(ctx workflow.Context, state *PolicyLifecycleState, sig AdminVoidSignal) bool {
	// → VOID; cancel all pending requests [BR-PM-073]
	if _, seen := state.ProcessedSignalIDs[sig.RequestID]; seen {
		return false
	}
	for _, pr := range state.PendingRequests {
		_ = workflow.ExecuteActivity(shortActCtx(ctx),
			policyActs.CancelDownstreamWorkflowActivity,
			acts.CancelWorkflowParams{WorkflowID: pr.DownstreamWorkflow}).Get(ctx, nil)
	}
	// Release DB financial lock before clearing in-memory state [Review-Fix-1, BR-PM-030]
	if state.ActiveLock != nil {
		_ = workflow.ExecuteActivity(shortActCtx(ctx),
			policyActs.ReleaseFinancialLockActivity, state.PolicyDBID).Get(ctx, nil)
	}
	state.PendingRequests = nil
	state.ActiveLock = nil
	doTransition(ctx, state, state.CurrentStatus, domain.StatusVoid,
		fmt.Sprintf("admin void by %d: %s", sig.AuthorizedBy, sig.Reason), SignalAdminVoid, sig.RequestID)
	state.ProcessedSignalIDs[sig.RequestID] = workflow.Now(ctx)
	return true
}
```

**Step 2: Build check**
```bash
go build ./...
```
Expected: exit 0.

**Step 3: Commit**
```bash
git add workflows/policy_lifecycle_workflow.go
git commit -m "fix: release DB financial lock in handleAdminVoid

state.ActiveLock = nil clears in-memory lock but leaves the
policy_financial_lock DB row. After reopen, CheckFinancialLock()
finds the stale row and blocks all financial requests (HTTP 409 forever).

Added ReleaseFinancialLockActivity call when ActiveLock is set,
consistent with handleWithdrawal(). Fixes BR-PM-073 + BR-PM-030."
```

---

## Group B — High

---

## Task 4: Config-driven routing timeouts + cooling durations

**File:** `workflows/policy_lifecycle_workflow.go`

**The bug:** `routingTimeoutForRequest()` (line 52) and `coolingDuration()` (line 85)
hardcode durations that should be fetched from `policy_state_config` via
`FetchWorkflowConfigActivity`. Config keys already exist in `domain/policy_state_config.go`.

**Strategy:** Add two helper functions that look up from `state.CachedConfig` with
hardcoded fallback. Populate `state.CachedConfig` at workflow startup.

**Step 1: Add config-lookup helpers**

After the existing `coolingDuration()` function, add:

```go
// routingTimeoutFromConfig returns the routing timeout for a request type,
// reading from the workflow's CachedConfig with hardcoded fallback. [Review-Fix-16]
func routingTimeoutFromConfig(cachedConfig map[string]string, requestType string) time.Duration {
	key := routingTimeoutConfigKeyForType(requestType)
	if v, ok := cachedConfig[key]; ok {
		if d, err := time.ParseDuration(v); err == nil && d > 0 {
			return d
		}
	}
	return routingTimeoutForRequest(requestType) // hardcoded fallback
}

// coolingDurationFromConfig returns the terminal cooling period,
// reading from CachedConfig with hardcoded fallback. [Review-Fix-16]
func coolingDurationFromConfig(cachedConfig map[string]string, terminalStatus string) time.Duration {
	key := coolingConfigKeyForStatus(terminalStatus)
	if v, ok := cachedConfig[key]; ok {
		if d, err := time.ParseDuration(v); err == nil && d > 0 {
			return d
		}
	}
	return coolingDuration(terminalStatus) // hardcoded fallback
}

// routingTimeoutConfigKeyForType maps request type → ConfigKeyRoutingTimeout* constant.
func routingTimeoutConfigKeyForType(requestType string) string {
	switch requestType {
	case domain.RequestTypeSurrender:
		return domain.ConfigKeyRoutingTimeoutSurrender
	case domain.RequestTypeForcedSurrender:
		return domain.ConfigKeyRoutingTimeoutForcedSurrender
	case domain.RequestTypeLoan:
		return domain.ConfigKeyRoutingTimeoutLoan
	case domain.RequestTypeLoanRepayment:
		return domain.ConfigKeyRoutingTimeoutLoanRepayment
	case domain.RequestTypeRevival:
		return domain.ConfigKeyRoutingTimeoutRevival
	case domain.RequestTypeDeathClaim:
		return domain.ConfigKeyRoutingTimeoutDeathClaim
	case domain.RequestTypeMaturityClaim:
		return domain.ConfigKeyRoutingTimeoutMaturityClaim
	case domain.RequestTypeSurvivalBenefit:
		return domain.ConfigKeyRoutingTimeoutSurvivalBenefit
	case domain.RequestTypeCommutation:
		return domain.ConfigKeyRoutingTimeoutCommutation
	case domain.RequestTypeConversion:
		return domain.ConfigKeyRoutingTimeoutConversion
	case domain.RequestTypeFLC:
		return domain.ConfigKeyRoutingTimeoutFLC
	case domain.RequestTypePremiumRefund:
		return domain.ConfigKeyRoutingTimeoutPremiumRefund
	default:
		return domain.ConfigKeyRoutingTimeoutNFR
	}
}

// coolingConfigKeyForStatus maps terminal status → ConfigKeyCoolingPeriod* constant.
func coolingConfigKeyForStatus(terminalStatus string) string {
	switch terminalStatus {
	case domain.StatusVoid:
		return domain.ConfigKeyCoolingPeriodVoid
	case domain.StatusSurrendered:
		return domain.ConfigKeyCoolingPeriodSurrendered
	case domain.StatusTerminatedSurrender:
		return domain.ConfigKeyCoolingPeriodTerminatedSurrender
	case domain.StatusMatured:
		return domain.ConfigKeyCoolingPeriodMatured
	case domain.StatusDeathClaimSettled:
		return domain.ConfigKeyCoolingPeriodDeathClaimSettled
	case domain.StatusFLCCancelled:
		return domain.ConfigKeyCoolingPeriodFLCCancelled
	case domain.StatusCancelledDeath:
		return domain.ConfigKeyCoolingPeriodCancelledDeath
	case domain.StatusConverted:
		return domain.ConfigKeyCoolingPeriodConverted
	default:
		return ""
	}
}
```

**Step 2: Load all timeout/cooling configs at workflow startup**

Find the main workflow function (look for `func PolicyLifecycleWorkflow`). Find where
`state.CachedConfig` is populated (or where the main loop starts). Add a batch config
fetch right after the policy-created signal is processed and `state.PolicyDBID` is set.

Look for where `FetchWorkflowConfigActivity` is already called (for FLC period). Near
that area, also batch-load routing/cooling configs:

```go
// Batch-load routing timeout + cooling period configs into CachedConfig [Review-Fix-16]
allTimeoutKeys := []string{
    domain.ConfigKeyRoutingTimeoutSurrender,
    domain.ConfigKeyRoutingTimeoutForcedSurrender,
    domain.ConfigKeyRoutingTimeoutLoan,
    domain.ConfigKeyRoutingTimeoutLoanRepayment,
    domain.ConfigKeyRoutingTimeoutRevival,
    domain.ConfigKeyRoutingTimeoutDeathClaim,
    domain.ConfigKeyRoutingTimeoutMaturityClaim,
    domain.ConfigKeyRoutingTimeoutSurvivalBenefit,
    domain.ConfigKeyRoutingTimeoutCommutation,
    domain.ConfigKeyRoutingTimeoutConversion,
    domain.ConfigKeyRoutingTimeoutFLC,
    domain.ConfigKeyRoutingTimeoutPremiumRefund,
    domain.ConfigKeyRoutingTimeoutNFR,
    domain.ConfigKeyCoolingPeriodVoid,
    domain.ConfigKeyCoolingPeriodSurrendered,
    domain.ConfigKeyCoolingPeriodTerminatedSurrender,
    domain.ConfigKeyCoolingPeriodMatured,
    domain.ConfigKeyCoolingPeriodDeathClaimSettled,
    domain.ConfigKeyCoolingPeriodFLCCancelled,
    domain.ConfigKeyCoolingPeriodCancelledDeath,
    domain.ConfigKeyCoolingPeriodConverted,
}
if state.CachedConfig == nil {
    state.CachedConfig = make(map[string]string)
}
for _, configKey := range allTimeoutKeys {
    if _, already := state.CachedConfig[configKey]; already {
        continue // skip already-loaded (CAN carry-over)
    }
    var val string
    _ = workflow.ExecuteActivity(shortActCtx(ctx),
        policyActs.FetchWorkflowConfigActivity, configKey).Get(ctx, &val)
    if val != "" {
        state.CachedConfig[configKey] = val
    }
}
```

**Step 3: Update call sites to use config-backed helpers**

Search for `routingTimeoutForRequest(` and `coolingDuration(` in the file. Replace each:

```bash
# To find all usages:
grep -n "routingTimeoutForRequest\|coolingDuration(" workflows/policy_lifecycle_workflow.go
```

Replace `routingTimeoutForRequest(sig.RequestType)` with:
```go
routingTimeoutFromConfig(state.CachedConfig, sig.RequestType)
```

Replace `coolingDuration(state.CurrentStatus)` with:
```go
coolingDurationFromConfig(state.CachedConfig, state.CurrentStatus)
```

**Step 4: Build check**
```bash
go build ./...
```
Expected: exit 0. Fix any compilation errors.

**Step 5: Commit**
```bash
git add workflows/policy_lifecycle_workflow.go
git commit -m "feat: use config-driven routing timeouts and cooling durations

routingTimeoutForRequest() and coolingDuration() previously hardcoded
durations. Now reads from state.CachedConfig populated at workflow
startup via FetchWorkflowConfigActivity, with hardcoded values as
fallback. All 13 routing timeout and 8 cooling period config keys are
loaded. Fixes Review-Fix-16."
```

---

## Task 5: Add dedup + audit to `handleOperationCompleted()`

**File:** `workflows/policy_lifecycle_workflow.go`

**The bug (line 1018):** `handleOperationCompleted()` has no dedup check and no audit log.
A retried `operation-completed` signal processes twice: double status update + double
state transition.

**Step 1: Find the function and add dedup + logging at the top**

Find:
```go
func handleOperationCompleted(ctx workflow.Context, state *PolicyLifecycleState, sig OperationCompletedSignal) bool {
	// Find and remove matching pending request
	var matched *PendingRequest
```

Replace the start of the function with:
```go
func handleOperationCompleted(ctx workflow.Context, state *PolicyLifecycleState, sig OperationCompletedSignal) bool {
	// Dedup check — downstream services may retry completion signals
	if _, seen := state.ProcessedSignalIDs[sig.RequestID]; seen {
		return false
	}

	// Audit: log signal receipt in policy_signal_log [Review-Fix-2]
	sigPayload, _ := json.Marshal(sig)
	stateBefore := state.CurrentStatus
	_ = workflow.ExecuteActivity(shortActCtx(ctx),
		policyActs.LogSignalReceivedActivity,
		acts.SignalLogEntry{
			PolicyID:      state.PolicyDBID,
			SignalChannel: "operation-completed",
			SignalPayload: sigPayload,
			RequestID:     sig.RequestID,
			Status:        domain.SignalStatusProcessed,
			StateBefore:   &stateBefore,
		}).Get(ctx, nil)

	// Find and remove matching pending request
	var matched *PendingRequest
```

**Step 2: Mark as processed at the end of the function**

Find the `return isTerminal` at the end of `handleOperationCompleted`. Before it, add:
```go
	state.ProcessedSignalIDs[sig.RequestID] = workflow.Now(ctx)
	return isTerminal
```

**Step 3: Verify `acts` and `json` are imported** (they should already be; check imports).

**Step 4: Build check**
```bash
go build ./...
```
Expected: exit 0.

**Step 5: Commit**
```bash
git add workflows/policy_lifecycle_workflow.go
git commit -m "fix: add dedup and audit logging to handleOperationCompleted

Downstream services may retry operation-completed signals (e.g. on gRPC
timeout). Without dedup, PLW processed completions multiple times: double
state transitions and double service_request status updates.

Added ProcessedSignalIDs check and LogSignalReceivedActivity call
consistent with handleFinancialRequest(). Fixes Phase 5 finding 5."
```

---

## Task 6: Add `LogSignalReceivedActivity` to system signal handlers

**File:** `workflows/policy_lifecycle_workflow.go`

**The bug:** 11 system/compliance signal handlers have dedup via `ProcessedSignalIDs`
but none call `LogSignalReceivedActivity`. Compliance signals (AML suspension, payment
dishonor, investigation) leave no audit trail.

**Handlers to update:**
1. `handlePremiumPaid`
2. `handlePaymentDishonored`
3. `handleAMLFlagRaised`
4. `handleAMLFlagCleared`
5. `handleInvestigationStarted`
6. `handleInvestigationConcluded`
7. `handleLoanBalanceUpdated`
8. `handleConversionReversed`
9. `handleCustomerIDMerge`
10. `handleDisputeRegistered` / `handleDisputeResolved`

**Pattern to apply to each handler** (example for `handlePremiumPaid`):

Find — after the dedup check, before state updates:
```go
func handlePremiumPaid(ctx workflow.Context, state *PolicyLifecycleState, sig PremiumPaidSignal) {
	if _, seen := state.ProcessedSignalIDs[sig.RequestID]; seen {
		return
	}
	state.Metadata.PaidToDate = sig.NewPaidToDate
```

Replace with:
```go
func handlePremiumPaid(ctx workflow.Context, state *PolicyLifecycleState, sig PremiumPaidSignal) {
	if _, seen := state.ProcessedSignalIDs[sig.RequestID]; seen {
		return
	}
	// Audit: log signal receipt [Review-Fix-2]
	sigPayload, _ := json.Marshal(sig)
	stateBefore := state.CurrentStatus
	_ = workflow.ExecuteActivity(shortActCtx(ctx),
		policyActs.LogSignalReceivedActivity,
		acts.SignalLogEntry{
			PolicyID:      state.PolicyDBID,
			SignalChannel: SignalPremiumPaid,
			SignalPayload: sigPayload,
			RequestID:     sig.RequestID,
			Status:        domain.SignalStatusProcessed,
			StateBefore:   &stateBefore,
		}).Get(ctx, nil)
	state.Metadata.PaidToDate = sig.NewPaidToDate
```

Apply the same pattern to each of the 11 handlers. Use the correct signal channel
constant for each:
- `handlePremiumPaid` → `SignalPremiumPaid`
- `handlePaymentDishonored` → `SignalPaymentDishonored`
- `handleAMLFlagRaised` → `SignalAMLFlagRaised`
- `handleAMLFlagCleared` → `SignalAMLFlagCleared`
- `handleInvestigationStarted` → `SignalInvestigationStarted`
- `handleInvestigationConcluded` → `SignalInvestigationConcluded`
- `handleLoanBalanceUpdated` → `SignalLoanBalanceUpdated`
- `handleConversionReversed` → `SignalConversionReversed`
- `handleCustomerIDMerge` → `SignalCustomerIDMerge`
- `handleDisputeRegistered` (inside dispute handler) → `SignalDisputeRegistered`
- `handleDisputeResolved` → `SignalDisputeResolved`

**Note for dispute handlers:** The dispute signals in the main selector loop may be
inline (not separate functions). Search for `disputeRegisteredCh` and `disputeResolvedCh`
in the main loop. If they are inline closures, extract them into named functions first,
then add logging.

**Step 1: Apply pattern to all 11 handlers**

**Step 2: Build check**
```bash
go build ./...
```
Expected: exit 0.

**Step 3: Commit**
```bash
git add workflows/policy_lifecycle_workflow.go
git commit -m "feat: add LogSignalReceivedActivity to all system/compliance signal handlers

11 handlers (premium-paid, payment-dishonored, AML, investigation,
loan-balance, conversion-reversed, customer-id-merge, dispute) had
dedup but no audit trail in policy_signal_log.

Added LogSignalReceivedActivity after dedup check in each handler,
using the kebab-case signal channel constant from signals.go.
Fixes Phase 5 finding 6."
```

---

## Task 7: Allow-list NFR metadata update keys

**File:** `workflows/policy_lifecycle_workflow.go`

**The bug (line 1107-1119):** `handleNFRCompleted()` forwards arbitrary
`OutcomePayload` keys to `UpdatePolicyMetadataActivity` for ASSIGNMENT completions.
A rogue service could corrupt `customer_id`, `sum_assured`, etc.

**Step 1: Add `filterPayload` helper function**

After `handleNFRCompleted`, add:
```go
// nfrMetadataAllowList defines safe outcome payload keys per NFR type.
// Only these keys may be written to policy_metadata via UpdatePolicyMetadataActivity.
var nfrMetadataAllowList = map[string]map[string]bool{
	domain.RequestTypeAssignment: {
		"assignment_type":   true,
		"assignment_status": true,
		"assignee_name":     true,
		"assignee_address":  true,
	},
}

// filterNFRPayload returns only allow-listed keys from payload for the given NFR type.
func filterNFRPayload(requestType string, payload map[string]interface{}) map[string]interface{} {
	allowed, ok := nfrMetadataAllowList[requestType]
	if !ok {
		return nil // no metadata updates for this NFR type
	}
	filtered := make(map[string]interface{}, len(allowed))
	for k, v := range payload {
		if allowed[k] {
			filtered[k] = v
		}
	}
	return filtered
}
```

**Step 2: Update `handleNFRCompleted` to use the filter**

Find:
```go
	// Metadata update for assignment NFR
	if matched.RequestType == domain.RequestTypeAssignment && sig.Outcome == domain.RequestOutcomeApproved {
		if sig.OutcomePayload != nil {
			var payload map[string]interface{}
			if json.Unmarshal(sig.OutcomePayload, &payload) == nil {
				_ = workflow.ExecuteActivity(shortActCtx(ctx),
					policyActs.UpdatePolicyMetadataActivity,
					acts.MetadataUpdateParams{
						PolicyID: state.PolicyDBID,
						Updates:  payload,
					}).Get(ctx, nil)
			}
		}
	}
```

Replace with:
```go
	// Metadata update for NFR types that report outcome payload (e.g. assignment).
	// Only allow-listed keys are forwarded to prevent arbitrary metadata corruption.
	if sig.Outcome == domain.RequestOutcomeApproved && sig.OutcomePayload != nil {
		var payload map[string]interface{}
		if json.Unmarshal(sig.OutcomePayload, &payload) == nil {
			filtered := filterNFRPayload(matched.RequestType, payload)
			if len(filtered) > 0 {
				_ = workflow.ExecuteActivity(shortActCtx(ctx),
					policyActs.UpdatePolicyMetadataActivity,
					acts.MetadataUpdateParams{
						PolicyID: state.PolicyDBID,
						Updates:  filtered,
					}).Get(ctx, nil)
			}
		}
	}
```

**Step 3: Build check**
```bash
go build ./...
```
Expected: exit 0.

**Step 4: Commit**
```bash
git add workflows/policy_lifecycle_workflow.go
git commit -m "fix: allow-list NFR metadata keys forwarded to UpdatePolicyMetadataActivity

handleNFRCompleted previously forwarded arbitrary OutcomePayload keys,
allowing a rogue or buggy downstream service to overwrite protected
columns (customer_id, sum_assured, etc.).

Added nfrMetadataAllowList and filterNFRPayload helper. Assignment NFR
may only update assignment_type, assignment_status, assignee_name,
assignee_address. Fixes Phase 5 finding 7."
```

---

## Group C — Medium

---

## Task 8: Distance-marketing FLC period

**File:** `workflows/policy_lifecycle_workflow.go`

**The bug (line 731-737):** `handlePolicyCreated()` always fetches `ConfigKeyFLCPeriodDays`
(standard 15-day FLC window). Distance-marketing policies are entitled to 30 days by
regulation, controlled by `ConfigKeyFLCPeriodDistanceMarketing`.

**Step 1: Update config key selection in `handlePolicyCreated`**

Find:
```go
	// STEP 7: Spawn FLC timer goroutine [Constraint 10, §10.1.6]
	// Fetch FLC period from config; default to 15 days if config not found [Review-Fix-8]
	var flcDaysStr string
	_ = workflow.ExecuteActivity(shortActCtx(ctx),
		policyActs.FetchWorkflowConfigActivity,
		domain.ConfigKeyFLCPeriodDays).Get(ctx, &flcDaysStr)
```

Replace with:
```go
	// STEP 7: Spawn FLC timer goroutine [Constraint 10, §10.1.6]
	// Distance-marketing products get 30-day FLC window; standard is 15 days. [Review-Fix-8]
	flcConfigKey := domain.ConfigKeyFLCPeriodDays
	if sig.Metadata.IsDistanceMarketing {
		flcConfigKey = domain.ConfigKeyFLCPeriodDistanceMarketing
	}
	var flcDaysStr string
	_ = workflow.ExecuteActivity(shortActCtx(ctx),
		policyActs.FetchWorkflowConfigActivity,
		flcConfigKey).Get(ctx, &flcDaysStr)
```

**Step 2: Build check**
```bash
go build ./...
```
Expected: exit 0.

**Step 3: Commit**
```bash
git add workflows/policy_lifecycle_workflow.go
git commit -m "fix: apply distance-marketing FLC period (30d) when IsDistanceMarketing=true

Standard FLC is 15 days; distance-marketing policies are entitled to 30
days by regulation. handlePolicyCreated() previously ignored the
IsDistanceMarketing metadata flag and always applied the standard period.

Now selects ConfigKeyFLCPeriodDistanceMarketing config key when flag is
set. Fixes Phase 5 finding 8 / compliance requirement."
```

---

## Task 9: Signal audit channel name fix

**File:** `workflows/policy_lifecycle_workflow.go`

**The bug (line ~775):** `handleFinancialRequest()` logs `SignalChannel: sig.RequestType`
which is `"SURRENDER"`, `"LOAN"` etc. The actual Temporal signal name is the
kebab-case constant (e.g. `"surrender-request"`). Audit logs are inconsistent with
Temporal history.

**Step 1: Add a request type → signal name mapping helper**

After the existing `routingTimeoutConfigKeyForType` function (added in Task 4), add:
```go
// requestTypeToSignalName maps domain request types to the kebab-case Temporal signal name.
// Used for audit log consistency with Temporal signal channel names. [signals.go §9.1]
func requestTypeToSignalName(requestType string) string {
	switch requestType {
	case domain.RequestTypeSurrender:
		return SignalSurrenderRequest
	case domain.RequestTypeLoan:
		return SignalLoanRequest
	case domain.RequestTypeLoanRepayment:
		return SignalLoanRepayment
	case domain.RequestTypeRevival:
		return SignalRevivalRequest
	case domain.RequestTypeDeathClaim:
		return SignalDeathNotification
	case domain.RequestTypeMaturityClaim:
		return SignalMaturityClaimRequest
	case domain.RequestTypeSurvivalBenefit:
		return SignalSurvivalBenefitRequest
	case domain.RequestTypeCommutation:
		return SignalCommutationRequest
	case domain.RequestTypeConversion:
		return SignalConversionRequest
	case domain.RequestTypeFLC:
		return SignalFLCRequest
	case domain.RequestTypePaidUp, domain.RequestTypeForcedSurrender:
		return SignalForcedSurrenderTrigger // forced surrender uses this for batch
	default:
		return SignalNFRRequest
	}
}
```

**Step 2: Update `handleFinancialRequest` to use signal name**

Find (around line 773-779):
```go
	_ = workflow.ExecuteActivity(shortActCtx(ctx),
		policyActs.LogSignalReceivedActivity,
		acts.SignalLogEntry{
			PolicyID:      state.PolicyDBID,
			SignalChannel: sig.RequestType,
```

Replace `sig.RequestType` with the helper call:
```go
			SignalChannel: requestTypeToSignalName(sig.RequestType),
```

**Step 3: Build check**
```bash
go build ./...
```
Expected: exit 0.

**Step 4: Commit**
```bash
git add workflows/policy_lifecycle_workflow.go
git commit -m "fix: log kebab-case signal channel name in handleFinancialRequest audit

Previously logged sig.RequestType ('SURRENDER') instead of the actual
Temporal signal channel name ('surrender-request'). Audit log entries
are now consistent with signals.go constants and Temporal history.
Fixes Phase 5 finding 9."
```

---

## Task 10: Final build and verification

**Step 1: Full build**
```bash
cd /d/policy-manage/policy-management
go build ./...
```
Expected: exit 0.

**Step 2: Vet**
```bash
go vet ./...
```
Expected: no output, exit 0.

**Step 3: Run any existing tests**
```bash
go test ./...
```
Expected: all packages pass (no existing tests should regress).

**Step 4: Verify no uses of bare `time.Now()` in workflow package**
```bash
grep -n "time\.Now()" /d/policy-manage/policy-management/workflows/policy_lifecycle_workflow.go
```
Expected: no output (all `time.Now()` replaced with `workflow.Now(ctx)`).

**Step 5: Verify `ParentClosePolicy` is set**
```bash
grep -n "ParentClosePolicy" /d/policy-manage/policy-management/workflows/policy_lifecycle_workflow.go
```
Expected: one line containing `ParentClosePolicyAbandon`.

**Step 6: Final summary commit (if any cleanup needed)**
```bash
git status
```
If all changes committed: nothing to do.

---

## Verification Checklist

### Critical (Group A)
- [ ] `pruneProcessedSignals` accepts `workflow.Context` and uses `workflow.Now(ctx)`
- [ ] No `time.Now()` calls in `policy_lifecycle_workflow.go`
- [ ] `childWFCtx()` has `ParentClosePolicy: temporal.ParentClosePolicyAbandon`
- [ ] `handleAdminVoid()` calls `ReleaseFinancialLockActivity` when `ActiveLock != nil`

### High (Group B)
- [ ] `routingTimeoutFromConfig()` and `coolingDurationFromConfig()` helpers added
- [ ] Config keys batch-loaded at workflow startup into `state.CachedConfig`
- [ ] `handleOperationCompleted()` has dedup check + LogSignalReceivedActivity
- [ ] All 11 system signal handlers call `LogSignalReceivedActivity`
- [ ] `filterNFRPayload()` helper added with `nfrMetadataAllowList`
- [ ] `handleNFRCompleted()` uses `filterNFRPayload()` before `UpdatePolicyMetadataActivity`

### Medium (Group C)
- [ ] `handlePolicyCreated()` uses `ConfigKeyFLCPeriodDistanceMarketing` when `IsDistanceMarketing=true`
- [ ] `handleFinancialRequest()` uses `requestTypeToSignalName()` for `SignalChannel`

### Build
- [ ] `go build ./...` passes
- [ ] `go vet ./...` clean
- [ ] `go test ./...` no regressions
