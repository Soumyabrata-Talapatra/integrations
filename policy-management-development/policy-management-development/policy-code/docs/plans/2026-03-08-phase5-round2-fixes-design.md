# Design: Phase 5 Code Review Round-2 Fixes (Workstream C)

**Date:** 2026-03-08
**Author:** Engineering
**Source review:** Phase 5 code review — second pass, 9 confirmed bugs
**Parent plan:** `docs/plans/2026-03-08-phase5-code-review-findings.md`

---

## 1. Scope

Nine bugs found in a second Phase 5 review of `policy_lifecycle_workflow.go`,
`batch_activities.go`, and `quote_activities.go`. All are confirmed functional
defects blocking release.

| ID | Severity | Component | Title |
|----|----------|-----------|-------|
| C1 | Critical  | PLW        | CAN loop — EventCount not reset before ContinueAsNew |
| C2 | Critical  | PLW        | `handleOperationCompleted` proceeds when no pending request matched |
| C3 | High      | PLW        | Double audit log in `handleFinancialRequest` (PROCESSED then REJECTED) |
| C4 | High      | Quote activities | URL path injection — policy number not escaped |
| C5 | High      | PLW        | `computeDisplayStatus` diverges from DB `compute_display_status()` |
| C6 | Critical  | Batch      | `LapsationScanActivity` uses flat 12-month remission; slab logic not implemented |
| C7 | High      | PLW        | `RefreshStateFromDBActivity` never called — stale state race vs batch writes |
| C8 | High      | Batch      | `ForcedSurrenderEvalActivity` uses `sum_assured` proxy instead of actual GSV |
| C9 | High      | Batch      | `PaidUpConversionScanActivity` omits bonus from paid-up value formula |

---

## 2. Technical Context

- **Language / runtime:** Go 1.26, Temporal SDK v1.37.0
- **Key files:**
  - `workflows/policy_lifecycle_workflow.go` — PLW (C1, C2, C3, C5, C7)
  - `workflows/activities/quote_activities.go` — quote proxy activities (C4)
  - `workflows/activities/batch_activities.go` — daily/monthly batch scans (C6, C8, C9)
  - `migrations/001_policy_mgmt_schema.sql` — reference DB functions (`compute_display_status`, `compute_remission_expiry`, `last_day_of_month`)
  - `migrations/002_seed_policy_state_config.sql` — config keys for remission slabs
- **Existing patterns reused:**
  - `bulkTransition` — pgx.Batch bulk UPDATE + history INSERT
  - `LogSignalReceivedActivity` — audit trail for signal receipt
  - `RefreshStateFromDBActivity` — returns `PolicyRefreshedState` (exists but uncalled)

---

## 3. Bug Designs

### C1 — CAN EventCount Reset

**Root cause:** `shouldContinueAsNew()` checks `state.EventCount` and `state.HistorySizeBytes`. When
true, `state.LastCANTime` is updated but `state.EventCount` is NOT reset. On replay after CAN, the
workflow immediately triggers CAN again (tight loop, consuming Temporal quotas).

**Fix:** Add `state.EventCount = 0` and `state.HistorySizeBytes = 0` immediately before
`workflow.NewContinueAsNewError()`.

```go
// Before:
state.LastCANTime = workflow.Now(ctx)
return workflow.NewContinueAsNewError(ctx, PolicyLifecycleWorkflow, state)

// After:
state.LastCANTime = workflow.Now(ctx)
state.EventCount = 0           // ← reset so new run starts fresh
state.HistorySizeBytes = 0     // ← reset size counter
return workflow.NewContinueAsNewError(ctx, PolicyLifecycleWorkflow, state)
```

---

### C2 — `handleOperationCompleted` Nil-Matched Guard

**Root cause:** When a `operation-completed` signal arrives for an unknown request ID, the
loop to find the matching pending request exits with `matched == nil`. The code then proceeds
to update `state.PolicyStatus`, creating a silent zero-value state corruption.

**Fix:** Return early with a warning log if no pending request was matched.

```go
if matched == nil {
    logger.Warn("handleOperationCompleted: no pending request matched signal",
        "signalRequestID", sig.RequestID)
    return
}
// ... existing processing continues
```

---

### C3 — Double Audit Log in `handleFinancialRequest`

**Root cause:** `handleFinancialRequest` calls `LogSignalReceivedActivity` with
`SignalStatus = "PROCESSED"` before performing the eligibility check. If the policy is
SUSPENDED/ineligible, `RecordRejectedRequestActivity` is also called, resulting in two
conflicting audit events for the same signal.

**Fix:** Defer the `LogSignalReceivedActivity` call until after the eligibility check,
using the actual outcome status:

```
1. Decode signal
2. Check isStateEligible()
3a. If ineligible:
    - LogSignalReceivedActivity(status="REJECTED")
    - RecordRejectedRequestActivity(...)
    - return
3b. If eligible:
    - [proceed with request creation]
    - LogSignalReceivedActivity(status="PROCESSED")
```

---

### C4 — URL Path Injection in Quote Activities

**Root cause:** Policy numbers follow the format `PLI/YYYY/NNNNNN` (e.g., `PLI/2026/000001`).
When embedded raw in a URL path via `fmt.Sprintf("%s/.../%s/...", base, policyNumber)`, the
slashes create extra path segments, routing to the wrong endpoint or returning 404/500.

**Fix:** Apply `url.PathEscape(policyNumber)` at URL construction in all three activities:
- `GetSurrenderQuoteActivity` — path segment
- `GetLoanQuoteActivity` — path segment
- `GetConversionQuoteActivity` — path segment

Also escape query parameters with `url.QueryEscape()` for `asOfDate` and `targetProductCode`.

Add `"net/url"` import to `quote_activities.go`.

---

### C5 — `computeDisplayStatus` Diverges from DB

**Root cause:** The Go function `computeDisplayStatus(status string, enc EncumbranceFlags)`:
1. Returns `"SUSPENDED"` (overrides status entirely) when `AMLHold == true` — wrong
2. Appends `"_DISPUTED"` when `DisputeFlag == true` — correct
3. Never appends `"_LOAN"` or `"_{AssignmentType}"` — missing

The DB trigger `compute_display_status()` (in migration 001) produces:
```sql
p_status || (LOAN suffix) || (ASSIGNMENT suffix) || (AML_HOLD suffix) || (DISPUTED suffix)
```
in strict left-to-right priority order.

**Fix:** Rewrite `computeDisplayStatus` to mirror the DB exactly:

```go
func computeDisplayStatus(status string, enc EncumbranceFlags) string {
    s := status
    if enc.HasActiveLoan {
        s += "_LOAN"
    }
    if enc.AssignmentType != "" && enc.AssignmentType != "NONE" {
        s += "_" + enc.AssignmentType
    }
    if enc.AMLHold {
        s += "_AML_HOLD"
    }
    if enc.DisputeFlag {
        s += "_DISPUTED"
    }
    return s
}
```

No call-site changes needed — existing callers pass the same arguments.

---

### C6 — `LapsationScanActivity` Remission Slab Logic

**Root cause:** All lapsing policies (VOID_LAPSE and INACTIVE_LAPSE paths) receive a flat
`remissionExpiry = now + 12×30 days` regardless of policy age. This violates
`compute_remission_expiry()` in the DB and the config keys in migration 002.

**Correct logic** (mirrors DB function):

| Policy life | Lapse status | Remission expiry formula |
|---|---|---|
| < 6 months  | VOID          | `nil` (no remission) |
| 6–12 months | VOID_LAPSE    | `lastDayOfMonth(paidToDate) + 30 days` |
| 12–24 months| VOID_LAPSE    | `lastDayOfMonth(paidToDate) + 60 days` |
| 24–36 months| VOID_LAPSE    | `lastDayOfMonth(paidToDate) + 90 days` |
| ≥ 36 months | INACTIVE_LAPSE| `paidToDate + 12 months` |

Note: `paid_to_date` in the batch row == `first_unpaid_date` parameter in the DB function.
`lastDayOfMonth(paidToDate)` == `last_day_of_month(p_first_unpaid)` in the DB function.

**New helpers to add to `batch_activities.go`:**

```go
// lastDayOfMonth returns the last calendar day of the month containing t.
// Mirrors DB: DATE_TRUNC('month', p_date) + INTERVAL '1 month' - INTERVAL '1 day'
func lastDayOfMonth(t time.Time) time.Time {
    firstOfNext := time.Date(t.Year(), t.Month()+1, 1, 0, 0, 0, 0, t.Location())
    return firstOfNext.Add(-24 * time.Hour)
}

// computeRemissionExpiry mirrors DB compute_remission_expiry().
// Returns nil for < 6 months policy life (no remission — VOID immediately).
func computeRemissionExpiry(issueDate, paidToDate, scheduledDate time.Time) *time.Time {
    life := monthsBetween(issueDate, scheduledDate)
    if life < 6 {
        return nil
    }
    graceEnd := lastDayOfMonth(paidToDate)
    var expiry time.Time
    switch {
    case life < 12:
        expiry = graceEnd.AddDate(0, 0, 30)
    case life < 24:
        expiry = graceEnd.AddDate(0, 0, 60)
    case life < 36:
        expiry = graceEnd.AddDate(0, 0, 90)
    default:
        expiry = paidToDate.AddDate(0, 12, 0) // first_unpaid + 12 months
    }
    return &expiry
}
```

**New `bulkTransitionWithRemissions` helper:**

Since policies in the same page have different `paidToDate` values, their remission expiries
differ. The existing `bulkTransition` accepts a single shared expiry. A new helper is needed:

```go
type policyRemissionPair struct {
    PolicyID        int64
    RemissionExpiry time.Time
}

// bulkTransitionWithRemissions performs per-policy remission expiry updates.
// Uses a pgx.Batch of individual UPDATE statements (one per policy) — same
// atomic-submission pattern as bulkTransition.
func (a *BatchActivities) bulkTransitionWithRemissions(
    ctx context.Context,
    pairs []policyRemissionPair,
    fromStatus, toStatus, reason string,
    scheduledDate, now time.Time,
) error { ... }
```

**Changes to `LapsationScanActivity`:**
- Collect `voidLapsePairs []policyRemissionPair` and `inactiveLapsePairs []policyRemissionPair` (instead of separate ID slices)
- Call `computeRemissionExpiry` per row in the switch
- Call `bulkTransitionWithRemissions` for VOID_LAPSE and INACTIVE_LAPSE groups
- `voidPolicies` (< 6 months) path unchanged — still uses `bulkTransition(..., nil)`

---

### C7 — `RefreshStateFromDBActivity` Never Called

**Root cause:** `RefreshStateFromDBActivity()` exists in `policy_activities.go` and returns
`PolicyRefreshedState` with the current DB status, encumbrance flags, and version. It is never
called from the PLW. When `SignalBatchStateSync` arrives, it updates `state.PolicyStatus` in
memory, but a subsequent `handleFinancialRequest` may call `isStateEligible()` against the
in-memory state that hasn't caught up with a DB-first batch operation.

**Fix:** At the start of `handleFinancialRequest()` and `handleNFRRequest()`, call
`RefreshStateFromDBActivity()` and merge the result into PLW state before eligibility check:

```
1. Call RefreshStateFromDBActivity()
2. Merge: state.PolicyStatus = refreshed.CurrentStatus
          state.Encumbrance.HasActiveLoan = refreshed.HasActiveLoan
          state.Encumbrance.LoanOutstanding = refreshed.LoanOutstanding
          state.Encumbrance.AssignmentType = refreshed.AssignmentType
          state.Encumbrance.AMLHold = refreshed.AMLHold
3. Call isStateEligible() with freshly merged state
```

Activity options: use the existing `shortActCtx(ctx)` (or equivalent 10s timeout context
already used for policy activities), RetryPolicy 3× default.

---

### C8 — `ForcedSurrenderEvalActivity` Uses `sum_assured` Proxy Instead of Actual GSV

**Root cause:** The DB query filter `loan_outstanding >= sum_assured * ratio` uses `sum_assured`
as a proxy for GSV. The comment acknowledges this: "actual GSV computation is downstream".
The PLW signal handler then has no GSV value to use for the forced-surrender threshold check.

**Chosen approach: per-candidate HTTP call to surrender-svc within the activity.**

Rationale:
- Pre-filter (`sum_assured` proxy) limits candidates to a small subset of ASSIGNED_TO_PRESIDENT policies (typically O(10s) per monthly run)
- HTTP call per candidate is acceptable at this cardinality
- Keeps the fix self-contained — no new service endpoint or cross-service coordination needed
- Consistent with the pattern in `QuoteActivities.GetSurrenderQuoteActivity`

**Changes to `BatchActivities`:**
- Add `httpClient *http.Client` field, initialized in `NewBatchActivities`
  ```go
  httpClient: &http.Client{Timeout: 8 * time.Second},
  ```
- Add private helper:
  ```go
  func (a *BatchActivities) fetchGSVFromSurrenderSvc(ctx context.Context, policyNumber string) (float64, error)
  ```
  Calls `<services.surrender_svc.internal_url>/internal/v1/policies/{policyNumber}/surrender-quote`
  (same endpoint as `GetSurrenderQuoteActivity`). Returns `GrossSurrenderValue`.

**Changes to `ForcedSurrenderEvalActivity`:**
- After fetching candidate rows from DB, for each row:
  1. Call `fetchGSVFromSurrenderSvc(ctx, row.PolicyNumber)`
  2. On error: log warning, skip policy (don't signal — safer to miss than to trigger incorrectly)
  3. Check: `row.LoanOutstanding >= gsv * loanRatioFraction`
  4. If yes: add `gsv` to signal payload; send `forced-surrender-trigger` signal
  5. If no: skip (the pre-filter caught it but actual GSV is lower — policy is safe)

---

### C9 — `PaidUpConversionScanActivity` Omits Bonus from Paid-Up Value

**Root cause:** The current formula is:
```go
puValue = (float64(premiumsPaid) / float64(totalPremiums)) * sumAssured
```
The PLI Directorate paid-up sum assured (PUSA) formula requires:
```
PUSA = (premiums_paid / total_premiums) × (sum_assured + bonus_accumulated)
```
The `bonus_accumulated` field exists in the policy table but is not selected in the
`PaidUpConversionScanActivity` query.

**Fix:**
1. Add `BonusAccumulated float64 \`db:"bonus_accumulated"\`` to the internal `paidUpRow` struct
2. Add `"bonus_accumulated"` to the `dblib.Psql.Select(...)` call
3. Change formula:
   ```go
   puValue = (float64(row.PremiumsPaid) / float64(row.TotalPremiums)) * (row.SumAssured + row.BonusAccumulated)
   ```

No other changes needed.

---

## 4. Source Code Structure Changes

| File | Type | Changes |
|------|------|---------|
| `workflows/policy_lifecycle_workflow.go` | Modify | C1 (2 lines), C2 (4 lines), C3 (restructure audit call), C5 (rewrite function), C7 (activity call + merge in 2 handlers) |
| `workflows/activities/quote_activities.go` | Modify | C4 (add `"net/url"` import; 3 url.PathEscape calls + 2 url.QueryEscape calls) |
| `workflows/activities/batch_activities.go` | Modify | C6 (2 new helpers + 1 new bulk helper + lapsation refactor), C8 (httpClient field + fetchGSV helper + eval logic change), C9 (paidUpRow field + select + formula) |

No new files. No schema changes. No migration needed (all config keys already exist in migration 002).

---

## 5. Delivery Phases

### Phase C-1 — Workflow correctness (C1, C2, C3, C5, C7)
Scope: `policy_lifecycle_workflow.go` only. Safe to ship independently.
Verification: `go build ./...`, `go vet ./...`, `go test ./workflows/...`

### Phase C-2 — URL Safety (C4)
Scope: `quote_activities.go`. Zero-risk isolation, no logic change.
Verification: `go build ./...`, `go vet ./...`, `go test ./workflows/...`

### Phase C-3 — Batch Activity Fixes (C6, C8, C9)
Scope: `batch_activities.go`. Highest risk — batch transitions affect policy state in DB.
Verification: `go build ./...`, `go vet ./...`, `go test ./workflows/...`
Manual: Run lapsation scan against staging DB with known test policies covering all 5 slab cases.

---

## 6. Verification Approach

- `go build ./...` — zero compile errors after each fix
- `go vet ./...` — zero vet issues
- `go test ./handler/...` — handler tests must continue to pass (3/3)
- `go test ./workflows/...` — no workflow tests yet; batch tests added as part of C6/C8/C9
- Manual staging: fire an `operation-completed` signal for an unknown request ID → PLW must not crash (C2); lapsation batch scan with 5 test policies covering all slab cases (C6)
