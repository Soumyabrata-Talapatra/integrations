# Design: Signal Payload Mismatch Fixes (Option A)
**Date:** 2026-03-08
**Status:** Approved
**Refs:** BR-PM-073, BR-PM-090, BR-PM-090+, §9.5.1, §9.5.2

---

## 1. Problem Statement

Three handler-side signal payload structs have JSON field names and/or types that differ
from the canonical workflow types defined in `workflows/signals.go`. Because Temporal
serialises signals via JSON, every mismatch silently yields a zero-value field in the
workflow — no panic, no error, just silent data loss.

### 1.1 Mismatch Table

| Struct | Field | Handler sends | Workflow expects | Runtime Impact |
|--------|-------|---------------|-----------------|----------------|
| `adminVoidSignalPayload` | dedup key | *(missing)* | `request_id` string | PLW dedup key = `""` every call |
| `adminVoidSignalPayload` | actor id | `voided_by` int64 | `authorized_by` int64 | `AuthorizedBy` always `0` in PLW |
| `reopenSignalPayload` | dedup key | *(missing)* | `request_id` string | PLW dedup key = `""` every call |
| `reopenSignalPayload` | reason | `reason` string | `reopen_reason` string | `ReopenReason` always `""` in PLW |
| `reopenSignalPayload` | actor id | `reopened_by` int64 | `authorized_by` int64 | `AuthorizedBy` always `0` in PLW |
| `withdrawalSignalPayload` | request id | `request_id` **int64** | `request_id` **string** | `RequestID` always `""` in PLW |
| `withdrawalSignalPayload` | target | *(missing)* | `target_request_id` string | PLW cannot find pending request to cancel |
| `withdrawalSignalPayload` | reason | `reason` string | `withdrawal_reason` string | `WithdrawalReason` always `""` in PLW |

**Most critical:** `withdrawalSignalPayload` missing `target_request_id` means
`handleWithdrawal()` in the PLW can never match the pending request — the downstream
child workflow is never cancelled and the financial lock is never released [BR-PM-090].

---

## 2. Scope

**In scope (this fix):**
- Fix 3 handler signal payload structs
- Update 3 call sites that construct those structs
- Add `github.com/google/uuid` as a direct import (currently indirect in `go.mod`)

**Out of scope (Option B — separate PR):**
- 11 missing system/compliance signal HTTP endpoints

---

## 3. Files Affected

| File | Change |
|------|--------|
| `handler/policy_request_handler.go` | Fix `adminVoidSignalPayload`, `reopenSignalPayload`; update 2 call sites |
| `handler/request_lifecycle_handler.go` | Fix `withdrawalSignalPayload`; update 1 call site |
| `go.mod` | Promote `github.com/google/uuid` indirect → direct |

No workflow code changes. `workflows/signals.go` is already correct.

---

## 4. Detailed Design

### 4.1 Fix `adminVoidSignalPayload`

**File:** `handler/policy_request_handler.go`

```go
// Before
type adminVoidSignalPayload struct {
    Reason   string `json:"reason"`
    VoidedBy int64  `json:"voided_by"`
}

// After
type adminVoidSignalPayload struct {
    RequestID    string `json:"request_id"`    // dedup key for PLW ProcessedSignalIDs
    Reason       string `json:"reason"`
    AuthorizedBy int64  `json:"authorized_by"` // was "voided_by"
}
```

**Call site in `AdminVoidPolicy`:**
```go
// Before
signalPayload := adminVoidSignalPayload{
    Reason:   req.Reason,
    VoidedBy: req.VoidedBy,
}

// After
idempKey := getIdempotencyKey(sctx.Ctx)
if idempKey == "" {
    idempKey = uuid.NewString() // generate if caller omitted X-Idempotency-Key
}
signalPayload := adminVoidSignalPayload{
    RequestID:    idempKey,
    Reason:       req.Reason,
    AuthorizedBy: req.VoidedBy,
}
```

**Rationale for `RequestID`:** Admin-void can legitimately be retried (e.g. on network
failure). Using the caller-provided `X-Idempotency-Key` header gives the PLW a stable
dedup key for HTTP-level retries. When the header is absent a fresh UUID is generated —
each distinct admin-void invocation gets a unique key, so two intentional admin-void
calls both reach the PLW.

---

### 4.2 Fix `reopenSignalPayload`

**File:** `handler/policy_request_handler.go`

```go
// Before
type reopenSignalPayload struct {
    Reason     string `json:"reason"`
    ReopenedBy int64  `json:"reopened_by"`
}

// After
type reopenSignalPayload struct {
    RequestID    string `json:"request_id"`    // dedup key
    ReopenReason string `json:"reopen_reason"` // was "reason"
    AuthorizedBy int64  `json:"authorized_by"` // was "reopened_by"
}
```

**Call site in `ReopenPolicy`:**
```go
// Before
signalPayload := reopenSignalPayload{
    Reason:     req.Reason,
    ReopenedBy: req.ReopenedBy,
}

// After
idempKey := getIdempotencyKey(sctx.Ctx)
if idempKey == "" {
    idempKey = uuid.NewString()
}
signalPayload := reopenSignalPayload{
    RequestID:    idempKey,
    ReopenReason: req.Reason,
    AuthorizedBy: req.ReopenedBy,
}
```

---

### 4.3 Fix `withdrawalSignalPayload`

**File:** `handler/request_lifecycle_handler.go`

```go
// Before
type withdrawalSignalPayload struct {
    RequestID   int64  `json:"request_id"`            // wrong type
    RequestType string `json:"request_type"`           // unused by workflow
    Reason      string `json:"reason"`                  // wrong tag
    WithdrawnBy *int64 `json:"withdrawn_by,omitempty"` // unused by workflow
}

// After
type withdrawalSignalPayload struct {
    RequestID        string `json:"request_id"`         // type: string; stable dedup key
    TargetRequestID  string `json:"target_request_id"`  // NEW — identifies pending request in PLW
    WithdrawalReason string `json:"withdrawal_reason"`  // was "reason"
}
```

**Call site in `WithdrawRequest` (after DB withdrawal succeeds):**
```go
// After
// TargetRequestID must match PendingRequest.RequestID stored in PLW.
// PLW stores the UUID idempotency key as PendingRequest.RequestID (Constraint 1).
// Fall back to BIGINT-as-string if idempotency key is absent (legacy requests).
targetRequestID := fmt.Sprintf("%d", requestID)
if sr.IdempotencyKey != nil && *sr.IdempotencyKey != "" {
    targetRequestID = *sr.IdempotencyKey
}

signalPayload := withdrawalSignalPayload{
    RequestID:        fmt.Sprintf("withdrawal-%d", requestID), // stable idempotent dedup key
    TargetRequestID:  targetRequestID,
    WithdrawalReason: req.Reason,
}
```

**Rationale for `TargetRequestID` resolution:**
- PLW stores `PendingRequest.RequestID = IdempotencyKey` (the UUID from the original
  submission's `X-Idempotency-Key` header per Constraint 1 / Review-Fix-11).
- The withdrawal handler already fetches `sr` (the `ServiceRequest` record) from DB,
  which has the `IdempotencyKey *string` field.
- Fallback to BIGINT string handles requests submitted before idempotency key was added.

**Rationale for `RequestID`:** `fmt.Sprintf("withdrawal-%d", requestID)` is deterministic
per withdrawal target — retrying the same withdrawal POST yields the same dedup key,
so the PLW processes it exactly once even on HTTP retries.

---

## 5. Dependency Change

`github.com/google/uuid v1.6.0` is already in `go.mod` as an indirect dependency
(pulled in by `go.temporal.io/sdk`). Promoting it to direct:

```
// go.mod — change indirect to direct
github.com/google/uuid v1.6.0
```

No version bump required; the existing version is already present in `go.sum`.

---

## 6. Testing Plan

1. **Unit tests (new):**
   - `TestAdminVoidSignalPayload_JSONRoundTrip` — marshal `adminVoidSignalPayload`,
     unmarshal into `workflows.AdminVoidSignal`, assert all fields match.
   - `TestReopenSignalPayload_JSONRoundTrip` — same pattern for reopen.
   - `TestWithdrawalSignalPayload_JSONRoundTrip` — same pattern; assert
     `TargetRequestID` is preserved.

2. **Handler integration tests (existing):**
   - Run `go test ./handler/...` — no regressions expected.

3. **Manual verification:** Fire admin-void/reopen/withdraw via HTTP and confirm the
   Temporal workflow receives correct field values using `tctl workflow showid`.

---

## 7. Risk

**Low.** Changes are purely in handler-side private structs. The workflow types
(`workflows/signals.go`) are not modified. The fixes make the handler JSON _match_
what the workflow already expects — no workflow behavioural change.

The only net-new behaviour is that the PLW now actually receives:
- A valid dedup key (prevents silent collision on `""`)
- A valid `AuthorizedBy` actor (instead of `0`)
- A valid `TargetRequestID` (withdrawal now works end-to-end)
