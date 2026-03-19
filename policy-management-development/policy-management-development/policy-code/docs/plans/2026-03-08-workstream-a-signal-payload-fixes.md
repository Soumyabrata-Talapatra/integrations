# Workstream A: Signal Payload Mismatch Fixes — Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Fix 5 JSON field name/type mismatches in 3 handler signal payload structs so the PolicyLifecycleWorkflow (PLW) receives correct values instead of silent zero-values.

**Architecture:** The PLW deserialises all inbound Temporal signals via JSON. Handler-side private structs must have JSON tags that exactly match the canonical types in `workflows/signals.go`. No workflow code changes — only handlers.

**Tech Stack:** Go 1.26, Temporal SDK v1.37.0, `github.com/google/uuid v1.6.0` (already in go.mod, promote indirect → direct)

---

## Context for the Engineer

### Signal flow
```
HTTP POST → handler builds payload struct → json.Marshal → tc.SignalWorkflow → PLW receives → json.Unmarshal into workflows.<Type>
```
Any JSON tag mismatch between the handler struct and the workflow struct means the workflow silently gets a zero-value (`""` for string, `0` for int64).

### Canonical types (DO NOT MODIFY these)
File: `workflows/signals.go`

```go
type AdminVoidSignal struct {
    RequestID    string `json:"request_id"`
    Reason       string `json:"reason"`
    AuthorizedBy int64  `json:"authorized_by"`
}

type ReopenRequestSignal struct {
    RequestID    string `json:"request_id"`
    ReopenReason string `json:"reopen_reason"`
    AuthorizedBy int64  `json:"authorized_by"`
}

type WithdrawalRequestSignal struct {
    RequestID        string `json:"request_id"`
    TargetRequestID  string `json:"target_request_id"`
    WithdrawalReason string `json:"withdrawal_reason"`
}
```

### Files to modify
- `handler/policy_request_handler.go` — `adminVoidSignalPayload`, `reopenSignalPayload`
- `handler/request_lifecycle_handler.go` — `withdrawalSignalPayload`
- `go.mod` — promote uuid indirect → direct

---

## Task 1: Add uuid import to go.mod

**Files:**
- Modify: `go.mod`

**Step 1: Add the import in go.mod**

Open `go.mod`. Find the line:
```
github.com/google/uuid v1.6.0 // indirect
```
Remove the `// indirect` comment so it becomes:
```
github.com/google/uuid v1.6.0
```

**Step 2: Run go mod tidy to verify**
```bash
cd /d/policy-manage/policy-management
go mod tidy
```
Expected: exits 0, no changes to `go.sum`.

**Step 3: Verify build still passes**
```bash
go build ./...
```
Expected: no output, exit 0.

**Step 4: Commit**
```bash
git add go.mod go.sum
git commit -m "chore: promote github.com/google/uuid to direct dependency"
```

---

## Task 2: Fix `adminVoidSignalPayload`

**Files:**
- Modify: `handler/policy_request_handler.go` (lines 77-81 and 875-878)
- Create: `handler/signal_payload_test.go`

### Step 1: Write the failing JSON round-trip test

Create `handler/signal_payload_test.go`:

```go
package handler

import (
	"encoding/json"
	"testing"

	"policy-management/workflows"
)

// TestAdminVoidSignalPayload_JSONRoundTrip verifies that adminVoidSignalPayload
// serialises to JSON that workflows.AdminVoidSignal can deserialise correctly.
// This guards against JSON tag mismatches causing silent zero-values in PLW.
func TestAdminVoidSignalPayload_JSONRoundTrip(t *testing.T) {
	payload := adminVoidSignalPayload{
		RequestID:    "test-uuid-1234",
		Reason:       "fraud detected",
		AuthorizedBy: 42,
	}

	b, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var sig workflows.AdminVoidSignal
	if err := json.Unmarshal(b, &sig); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if sig.RequestID != payload.RequestID {
		t.Errorf("RequestID: got %q, want %q", sig.RequestID, payload.RequestID)
	}
	if sig.Reason != payload.Reason {
		t.Errorf("Reason: got %q, want %q", sig.Reason, payload.Reason)
	}
	if sig.AuthorizedBy != payload.AuthorizedBy {
		t.Errorf("AuthorizedBy: got %d, want %d", sig.AuthorizedBy, payload.AuthorizedBy)
	}
}
```

**Step 2: Run test to confirm it FAILS**
```bash
cd /d/policy-manage/policy-management
go test ./handler/... -run TestAdminVoidSignalPayload_JSONRoundTrip -v
```
Expected output: `FAIL` — compile error because `adminVoidSignalPayload` doesn't have `RequestID` or `AuthorizedBy` fields yet.

**Step 3: Fix the struct in `handler/policy_request_handler.go`**

Find lines 77-81:
```go
// adminVoidSignalPayload is the payload for the admin-void signal. [BR-PM-073]
type adminVoidSignalPayload struct {
	Reason   string `json:"reason"`
	VoidedBy int64  `json:"voided_by"`
}
```

Replace with:
```go
// adminVoidSignalPayload is the payload for the admin-void signal. [BR-PM-073]
// JSON tags MUST match workflows.AdminVoidSignal exactly.
type adminVoidSignalPayload struct {
	RequestID    string `json:"request_id"`    // dedup key for PLW ProcessedSignalIDs
	Reason       string `json:"reason"`
	AuthorizedBy int64  `json:"authorized_by"` // was "voided_by" — mismatch fixed
}
```

**Step 4: Fix the call site in `AdminVoidPolicy` (~line 875)**

Find:
```go
	signalPayload := adminVoidSignalPayload{
		Reason:   req.Reason,
		VoidedBy: req.VoidedBy,
	}
```

Replace with:
```go
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

**Step 5: Add uuid import to the file's import block**

In `handler/policy_request_handler.go`, add to the import block:
```go
	"github.com/google/uuid"
```

The full import block should include it among other standard imports, e.g.:
```go
import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	// ... rest of imports unchanged
)
```

**Step 6: Run test to confirm it PASSES**
```bash
go test ./handler/... -run TestAdminVoidSignalPayload_JSONRoundTrip -v
```
Expected: `PASS`

**Step 7: Build check**
```bash
go build ./...
```
Expected: exit 0, no errors.

**Step 8: Commit**
```bash
git add handler/policy_request_handler.go handler/signal_payload_test.go
git commit -m "fix: align adminVoidSignalPayload JSON tags with AdminVoidSignal workflow type

- Add RequestID (request_id) field for PLW dedup key
- Rename VoidedBy json tag from voided_by to authorized_by
- Populate RequestID from X-Idempotency-Key header (uuid fallback)

Without this fix, PLW AdminVoidSignal always gets AuthorizedBy=0 and
dedup key='', breaking idempotency. Fixes BR-PM-073."
```

---

## Task 3: Fix `reopenSignalPayload`

**Files:**
- Modify: `handler/policy_request_handler.go` (lines 83-87 and ~933-936)
- Modify: `handler/signal_payload_test.go` (add test)

**Step 1: Add failing test to `handler/signal_payload_test.go`**

Add after the admin-void test:
```go
// TestReopenSignalPayload_JSONRoundTrip verifies that reopenSignalPayload
// serialises to JSON that workflows.ReopenRequestSignal can deserialise correctly.
func TestReopenSignalPayload_JSONRoundTrip(t *testing.T) {
	payload := reopenSignalPayload{
		RequestID:    "test-uuid-5678",
		ReopenReason: "error correction",
		AuthorizedBy: 99,
	}

	b, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var sig workflows.ReopenRequestSignal
	if err := json.Unmarshal(b, &sig); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if sig.RequestID != payload.RequestID {
		t.Errorf("RequestID: got %q, want %q", sig.RequestID, payload.RequestID)
	}
	if sig.ReopenReason != payload.ReopenReason {
		t.Errorf("ReopenReason: got %q, want %q", sig.ReopenReason, payload.ReopenReason)
	}
	if sig.AuthorizedBy != payload.AuthorizedBy {
		t.Errorf("AuthorizedBy: got %d, want %d", sig.AuthorizedBy, payload.AuthorizedBy)
	}
}
```

**Step 2: Run test to confirm it FAILS**
```bash
go test ./handler/... -run TestReopenSignalPayload_JSONRoundTrip -v
```
Expected: compile error — `reopenSignalPayload` doesn't have `ReopenReason` or `AuthorizedBy` yet.

**Step 3: Fix the struct in `handler/policy_request_handler.go`**

Find lines 83-87:
```go
// reopenSignalPayload is the signal sent to the PLW workflow on reopen. [BR-PM-090+]
type reopenSignalPayload struct {
	Reason     string `json:"reason"`
	ReopenedBy int64  `json:"reopened_by"`
}
```

Replace with:
```go
// reopenSignalPayload is the signal sent to the PLW workflow on reopen. [BR-PM-090+]
// JSON tags MUST match workflows.ReopenRequestSignal exactly.
type reopenSignalPayload struct {
	RequestID    string `json:"request_id"`    // dedup key for PLW ProcessedSignalIDs
	ReopenReason string `json:"reopen_reason"` // was "reason" — mismatch fixed
	AuthorizedBy int64  `json:"authorized_by"` // was "reopened_by" — mismatch fixed
}
```

**Step 4: Fix the call site in `ReopenPolicy` (~line 933)**

Find:
```go
	signalPayload := reopenSignalPayload{
		Reason:     req.Reason,
		ReopenedBy: req.ReopenedBy,
	}
```

Replace with:
```go
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

**Step 5: Run test to confirm it PASSES**
```bash
go test ./handler/... -run TestReopenSignalPayload_JSONRoundTrip -v
```
Expected: `PASS`

**Step 6: Build check**
```bash
go build ./...
```
Expected: exit 0.

**Step 7: Commit**
```bash
git add handler/policy_request_handler.go handler/signal_payload_test.go
git commit -m "fix: align reopenSignalPayload JSON tags with ReopenRequestSignal workflow type

- Add RequestID field for PLW dedup key
- Rename Reason to ReopenReason with json tag reopen_reason
- Rename ReopenedBy to AuthorizedBy with json tag authorized_by

Without this fix, PLW gets ReopenReason='' and AuthorizedBy=0 on every
reopen signal, silently dropping all data. Fixes BR-PM-090+."
```

---

## Task 4: Fix `withdrawalSignalPayload` (most critical fix)

**Files:**
- Modify: `handler/request_lifecycle_handler.go` (lines 37-45 and ~237-242)
- Modify: `handler/signal_payload_test.go` (add test)

**Background:** `TargetRequestID` tells the PLW which pending request to find and cancel.
Without it, `handleWithdrawal()` loops through `PendingRequests` looking for
`pr.RequestID == sig.TargetRequestID` — which is always `""` — and never cancels
the downstream child workflow or releases the financial lock.

**Step 1: Add failing test to `handler/signal_payload_test.go`**

Add after the reopen test:
```go
// TestWithdrawalSignalPayload_JSONRoundTrip verifies that withdrawalSignalPayload
// serialises to JSON that workflows.WithdrawalRequestSignal can deserialise correctly.
func TestWithdrawalSignalPayload_JSONRoundTrip(t *testing.T) {
	payload := withdrawalSignalPayload{
		RequestID:        "withdrawal-12345",
		TargetRequestID:  "idempotency-uuid-abcd",
		WithdrawalReason: "customer changed mind",
	}

	b, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var sig workflows.WithdrawalRequestSignal
	if err := json.Unmarshal(b, &sig); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if sig.RequestID != payload.RequestID {
		t.Errorf("RequestID: got %q, want %q", sig.RequestID, payload.RequestID)
	}
	if sig.TargetRequestID != payload.TargetRequestID {
		t.Errorf("TargetRequestID: got %q, want %q", sig.TargetRequestID, payload.TargetRequestID)
	}
	if sig.WithdrawalReason != payload.WithdrawalReason {
		t.Errorf("WithdrawalReason: got %q, want %q", sig.WithdrawalReason, payload.WithdrawalReason)
	}
}
```

**Step 2: Add workflows import to the test file**

The test file already imports `"policy-management/workflows"`. No additional imports needed.
But `withdrawalSignalPayload` is in `handler/request_lifecycle_handler.go`, same package.
The test file is `package handler` — this is fine since test and source are in the same package.

**Step 3: Run test to confirm it FAILS**
```bash
go test ./handler/... -run TestWithdrawalSignalPayload_JSONRoundTrip -v
```
Expected: compile error — `withdrawalSignalPayload` doesn't have `TargetRequestID` or `WithdrawalReason` yet.

**Step 4: Fix the struct in `handler/request_lifecycle_handler.go`**

Find lines 37-45:
```go
// withdrawalSignalPayload is the payload sent to plw-{policyNumber} on withdrawal.
// The workflow reverts policy state and cancels any pending downstream child workflow.
// [FR-PM-007, BR-PM-090]
type withdrawalSignalPayload struct {
	RequestID   int64  `json:"request_id"`
	RequestType string `json:"request_type"`
	Reason      string `json:"reason"`
	WithdrawnBy *int64 `json:"withdrawn_by,omitempty"`
}
```

Replace with:
```go
// withdrawalSignalPayload is the payload sent to plw-{policyNumber} on withdrawal.
// The workflow uses TargetRequestID to find the PendingRequest and cancel the
// downstream child workflow + release the financial lock. [FR-PM-007, BR-PM-090]
// JSON tags MUST match workflows.WithdrawalRequestSignal exactly.
type withdrawalSignalPayload struct {
	RequestID        string `json:"request_id"`         // stable dedup key: "withdrawal-{requestID}"
	TargetRequestID  string `json:"target_request_id"`  // PendingRequest.RequestID in PLW (UUID or BIGINT string)
	WithdrawalReason string `json:"withdrawal_reason"`  // was "reason" — mismatch fixed
}
```

**Step 5: Fix the call site in `WithdrawRequest` (~line 237)**

Find:
```go
	// Signal plw-{policyNumber} to revert state and cancel downstream child. [BR-PM-090]
	wfID := policyWorkflowID(policyNumber)
	signalPayload := withdrawalSignalPayload{
		RequestID:   requestID,
		RequestType: sr.RequestType,
		Reason:      req.Reason,
		WithdrawnBy: req.WithdrawnBy,
	}
```

Replace with:
```go
	// Signal plw-{policyNumber} to revert state and cancel downstream child. [BR-PM-090]
	// TargetRequestID must match PendingRequest.RequestID stored in PLW state.
	// PLW stores the UUID idempotency key as PendingRequest.RequestID (Constraint 1 / Review-Fix-11).
	// Fall back to BIGINT-as-string for legacy requests submitted without idempotency key.
	wfID := policyWorkflowID(policyNumber)
	targetRequestID := fmt.Sprintf("%d", requestID)
	if sr.IdempotencyKey != nil && *sr.IdempotencyKey != "" {
		targetRequestID = *sr.IdempotencyKey
	}
	signalPayload := withdrawalSignalPayload{
		RequestID:        fmt.Sprintf("withdrawal-%d", requestID), // deterministic dedup key per target
		TargetRequestID:  targetRequestID,
		WithdrawalReason: req.Reason,
	}
```

**Step 6: Add `"fmt"` import if not already present in `request_lifecycle_handler.go`**

Check the imports at the top of `handler/request_lifecycle_handler.go`:
```go
import (
	"errors"
	"fmt"     // ← should already be here
	"net/http"
	"time"
	// ...
)
```
`"fmt"` is already imported (used at line 219 for `fmt.Sprintf`). No import change needed.

**Step 7: Run test to confirm it PASSES**
```bash
go test ./handler/... -run TestWithdrawalSignalPayload_JSONRoundTrip -v
```
Expected: `PASS`

**Step 8: Run all handler tests**
```bash
go test ./handler/... -v
```
Expected: all tests PASS.

**Step 9: Build check**
```bash
go build ./...
```
Expected: exit 0.

**Step 10: Commit**
```bash
git add handler/request_lifecycle_handler.go handler/signal_payload_test.go
git commit -m "fix: align withdrawalSignalPayload with WithdrawalRequestSignal workflow type

CRITICAL: missing target_request_id meant PLW handleWithdrawal() could
never find the pending request — downstream child workflows were never
cancelled and financial locks were never released on withdrawal.

Changes:
- Add TargetRequestID string (target_request_id) from sr.IdempotencyKey
- Change RequestID type from int64 to string; use stable 'withdrawal-N' key
- Rename Reason to WithdrawalReason with json tag withdrawal_reason
- Remove RequestType and WithdrawnBy (not consumed by workflow)

Fixes BR-PM-090."
```

---

## Task 5: Run all tests and final verification

**Step 1: Run all package tests**
```bash
cd /d/policy-manage/policy-management
go test ./... -v 2>&1 | tail -30
```
Expected: all packages report `ok` or `PASS`. No `FAIL`.

**Step 2: Run go vet**
```bash
go vet ./...
```
Expected: no output, exit 0.

**Step 3: Check JSON tags directly (optional sanity check)**
```bash
go test ./handler/... -run "TestAdminVoid|TestReopen|TestWithdrawal" -v
```
Expected: 3 tests, all PASS.

**Step 4: Final commit if any stray changes**
```bash
git status
```
If clean: nothing to do. If any remaining changes:
```bash
git add -p  # review each hunk
git commit -m "chore: cleanup after signal payload mismatch fixes"
```

---

## Verification Checklist

- [ ] `go build ./...` passes
- [ ] `go vet ./...` clean
- [ ] `go test ./handler/... -run TestAdminVoidSignalPayload_JSONRoundTrip` → PASS
- [ ] `go test ./handler/... -run TestReopenSignalPayload_JSONRoundTrip` → PASS
- [ ] `go test ./handler/... -run TestWithdrawalSignalPayload_JSONRoundTrip` → PASS
- [ ] All 3 structs match their canonical `workflows/signals.go` type
- [ ] `withdrawalSignalPayload.TargetRequestID` is populated from `sr.IdempotencyKey`
- [ ] uuid is a direct (not indirect) import in `go.mod`
