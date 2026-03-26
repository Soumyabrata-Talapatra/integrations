# Policy Management Orchestrator — Architecture

This document describes the internal architecture, design patterns, and component interactions of the Policy Management Orchestrator microservice (IMS PLI 2.0).

---

## Table of Contents

1. [System Context](#1-system-context)
2. [Module / Package Structure](#2-module--package-structure)
3. [Policy Lifecycle State Machine](#3-policy-lifecycle-state-machine)
4. [Per-Policy Workflow (PLW)](#4-per-policy-workflow-plw)
5. [Request Flow Walkthrough](#5-request-flow-walkthrough)
6. [Batch Scan Workflows](#6-batch-scan-workflows)
7. [Database Schema](#7-database-schema)
8. [Key Design Patterns](#8-key-design-patterns)
9. [Signal Catalog](#9-signal-catalog)
10. [Integration Contracts](#10-integration-contracts)
11. [Downstream Service Routing](#11-downstream-service-routing)

---

## 1. System Context

The Policy Management Orchestrator (PM) sits at the centre of the IMS PLI 2.0 microservice ecosystem.

```
                  ┌─────────────────┐
                  │  Policy Issue   │ ── SignalWithStart ──▶ PM
                  │  Service        │    (policy-created signal)
                  └─────────────────┘

  Customer Portal ──▶ ┌─────────────────────────────┐
  Mobile App      ──▶ │  API Gateway / BFF           │ ──▶ PM REST API
  Agent Portal    ──▶ └─────────────────────────────┘
  CPC System      ──▶

                       PM (this service)
                       ┌──────────────────────────────────────────┐
                       │  Orchestrates child workflows on:         │
                       │  ├── surrender-tq   (Surrender Svc)      │
                       │  ├── loan-tq        (Loan Svc)           │
                       │  ├── claims-tq      (Claims Svc)         │
                       │  ├── revival-tq     (Revival Svc)        │
                       │  ├── commutation-tq (Commutation Svc)    │
                       │  ├── conversion-tq  (Conversion Svc)     │
                       │  ├── freelook-tq    (FLC Svc)            │
                       │  ├── nfs-tq         (NFS — Non-Fin Svc)  │
                       │  └── billing-tq     (Billing Svc)        │
                       └──────────────────────────────────────────┘
```

**PM is the sole writer of policy lifecycle state.** Downstream services process financial operations and signal results back to PM, which then updates the policy status in PostgreSQL and its in-memory workflow state.

---

## 2. Module / Package Structure

```
policy-management/
├── main.go
│     Entry point. Wires the FX dependency-injection container with three
│     modules: FxRepo, FxHandler, FxTemporal.
│
├── bootstrap/bootstrapper.go
│     FX module definitions.
│     • FxRepo    — provides 4 PostgreSQL repository singletons
│     • FxHandler — registers 5 handler structs (33 total endpoints) with
│                   the n-api-server HTTP server
│     • FxTemporal — dials Temporal client, starts worker on
│                    "policy-management-tq", registers all workflows + activities
│     • RegisterBatchSchedules() — creates 6 Temporal Schedules (idempotent)
│
├── configs/
│     YAML configuration. Base file + per-environment overrides loaded
│     by n-api-config based on APP_ENV.
│
├── core/
│   ├── domain/
│   │     All Go constants and entity structs that mirror the DB schema.
│   │     Key files:
│   │       policy.go                — Policy struct, 23 status constants, enums
│   │       service_request.go       — ServiceRequest struct, request type constants
│   │       policy_status_history.go — Audit trail struct
│   │       batch_scan.go            — BatchScanType constants, BatchScanResult
│   └── port/
│         Go interfaces for all external dependencies (repositories,
│         downstream HTTP clients). Keeps domain logic testable.
│
├── handler/
│     Five handler structs — each implements serverHandler.Handler (n-api-server).
│     All financial/NFR handlers share the same 9-step request pattern (see §5).
│       policy_request_handler.go    — 19 endpoints (financial + NFR + admin)
│       request_lifecycle_handler.go — 5 endpoints (list, detail, withdraw, CPC inbox)
│       policy_query_handler.go      — 6 endpoints (status, summary, state-gate, etc.)
│       quote_handler.go             — 3 endpoints (surrender/loan/conversion quotes)
│       cpc_lookup_handler.go        — 6 endpoints (static enum lookups; no DB/Temporal calls)
│
├── repo/postgres/
│     PostgreSQL implementations of the port interfaces.
│     Uses pgx/v5 for I/O and Masterminds/squirrel as a query builder.
│       policy_repository.go         — policy CRUD + terminal snapshot
│       service_request_repository.go — service_request CRUD (partition-aware)
│       signal_repository.go         — signal audit log writes
│       config_repository.go         — policy_state_config reads
│
├── migrations/
│     001_policy_mgmt_schema.sql    — Full DDL (tables, enums, sequences, indexes,
│                                     triggers, materialized views)
│     002_seed_policy_state_config.sql — Config key seed data for workflow config
│     003_register_temporal_search_attrs.sh — Temporal CLI commands for custom
│                                              search attributes
│
└── workflows/
    ├── signals.go
    │     ALL signal channel name constants, query handler name constants,
    │     workflow state structs (PolicyLifecycleState, PolicyMetadata,
    │     PendingRequest, FinancialLock, EncumbranceFlags), integration
    │     contract types, and query result types.
    │     This file is the single source of truth for the PLW wire format.
    │
    ├── policy_lifecycle_workflow.go
    │     The long-running per-policy workflow. See §4.
    │
    ├── batch_scan_workflow.go
    │     Short-lived batch dispatcher. See §6.
    │
    └── activities/
        ├── policy_activities.go   — 12 activities: InitializePolicy,
        │                           RecordStateTransition, UpdateServiceRequest,
        │                           LogSignalReceived, FetchWorkflowConfig,
        │                           FetchAllWorkflowConfigs, ReleaseFinancialLock,
        │                           RefreshStateFromDB, UpdatePolicyMetadata, …
        ├── batch_activities.go    — 7 activities: LapsationScan,
        │                           RemissionExpiryShort/Long, PaidUpConversion,
        │                           MaturityScan, ForcedSurrenderEval,
        │                           RecordBatchScanResult
        └── quote_activities.go    — 3 quote proxy workflows (short-lived):
                                    GetSurrenderQuote, GetLoanQuote,
                                    GetConversionQuote (each calls downstream HTTP)
```

---

## 3. Policy Lifecycle State Machine

### 3.1 The 23 Canonical States

| State | Category | Description |
|-------|----------|-------------|
| `FREE_LOOK_ACTIVE` | Active | Newly issued; 15-day (standard) or 30-day (distance-marketing) free-look period |
| `ACTIVE` | Active | Premium payments up to date |
| `VOID_LAPSE` | Lapse | < 3 years of premiums paid; in remission window |
| `INACTIVE_LAPSE` | Lapse | ≥ 3 years premiums paid; in 12-month remission window |
| `ACTIVE_LAPSE` | Lapse | Beyond remission; eligible for paid-up conversion |
| `REVIVAL_PENDING` | In-flight | Revival request routed to Revival Svc |
| `PAID_UP` | Active | Converted to reduced paid-up sum assured |
| `REDUCED_PAID_UP` | Active | Reduced paid-up (auto batch conversion) |
| `ASSIGNED_TO_PRESIDENT` | Active | Absolute assignment to President of India |
| `PENDING_AUTO_SURRENDER` | In-flight | Forced surrender triggered (loan ≥ 100% GSV) |
| `PENDING_SURRENDER` | In-flight | Voluntary surrender request routed |
| `PENDING_MATURITY` | In-flight | Within maturity window; claim routed |
| `DEATH_CLAIM_INTIMATED` | In-flight | Death notification received |
| `DEATH_UNDER_INVESTIGATION` | In-flight | DCI investigation started |
| `SUSPENDED` | Blocked | AML hold active; no financial requests allowed |
| `VOID` | **Terminal** | Admin void or lapse without remission recovery |
| `SURRENDERED` | **Terminal** | Voluntary surrender completed |
| `TERMINATED_SURRENDER` | **Terminal** | Forced surrender completed |
| `MATURED` | **Terminal** | Maturity claim settled |
| `DEATH_CLAIM_SETTLED` | **Terminal** | Death claim paid |
| `FLC_CANCELLED` | **Terminal** | Cancelled during free-look period |
| `CANCELLED_DEATH` | **Terminal** | Policy cancelled due to death (no claim filed) |
| `CONVERTED` | **Terminal** | Converted to another product |

### 3.2 Key Transition Flows

```
Policy Issue Svc ──SignalWithStart──▶ FREE_LOOK_ACTIVE
                                            │
               FLC cancelled (flc-request) ─┼──▶ FLC_CANCELLED ★
               FLC period expires ──────────┤
                                            ▼
                                         ACTIVE ◀───── AML cleared (restores prev status)
                                            │                    ▲
                          AML flag raised ──┼──────────────────▶ SUSPENDED
                                            │
              Payment dishonored (< 3yr) ───┼──▶ VOID_LAPSE
              Payment dishonored (≥ 3yr) ───┼──▶ INACTIVE_LAPSE ──(remission expires)──▶ ACTIVE_LAPSE
                                            │         │ remission expires                      │
                                            │         ▼                                        │ batch paid-up
                                            │       VOID ★                                    ▼
                                            │                                        PAID_UP or VOID ★
                                            │
              Revival request ─────────────┼──▶ REVIVAL_PENDING ──completed──▶ ACTIVE
                                            │
              Surrender request ────────────┼──▶ PENDING_SURRENDER ──completed──▶ SURRENDERED ★
                                            │
              Death notification ───────────┼──▶ DEATH_CLAIM_INTIMATED
                                            │          │ investigation-started
                                            │          ▼
                                            │    DEATH_UNDER_INVESTIGATION
                                            │          │ investigation-concluded (confirmed)
                                            │          ▼
                                            │    DEATH_CLAIM_SETTLED ★
                                            │
              Maturity scan (90d window) ───┼──▶ PENDING_MATURITY ──settled──▶ MATURED ★
                                            │
              Admin void ───────────────────┼──▶ VOID ★
                                            │
              Forced surrender ─────────────┼──▶ PENDING_AUTO_SURRENDER ──completed──▶ TERMINATED_SURRENDER ★
                                            │
              Voluntary paid-up ────────────┴──▶ PAID_UP or VOID (value < ₹10K threshold) ★

★ = Terminal state; workflow enters cooling period then ends via Continue-As-New
```

### 3.3 Terminal State Handling

When a policy reaches a terminal state, the PLW:

1. Records the transition in DB via `RecordStateTransitionActivity`
2. Cancels any in-flight pending requests (downstream child workflows continue independently due to `ParentClosePolicy: Abandon` — see §8.6)
3. Enters a **cooling period** (configurable via `policy_state_config`) listening only for `reopen-request` signals
4. After cooling, writes a terminal snapshot to `terminal_state_snapshot` then ends via Continue-As-New

---

## 4. Per-Policy Workflow (PLW)

### 4.1 Overview

`PolicyLifecycleWorkflow` in `workflows/policy_lifecycle_workflow.go` is a **long-running Temporal workflow** with ID `plw-{policyNumber}`. It runs for the entire lifetime of a policy — potentially decades.

The workflow's full state is kept in a `PolicyLifecycleState` struct (defined in `workflows/signals.go`) which is serialised and passed across Continue-As-New boundaries.

### 4.2 Continue-As-New (CAN)

Temporal has a hard limit on workflow history size (~50 000 events). PLW uses **Continue-As-New** to reset history while preserving all state:

- An `EventCount` field in `PolicyLifecycleState` increments on every signal received
- When `EventCount` reaches a configurable threshold (default ~500), the PLW serialises its state and calls `workflow.NewContinueAsNewError()`, passing the full state as input to a fresh execution
- `EventCount` is reset to 0 at the start of each new execution

### 4.3 Signal Processing

The PLW runs a main loop using `workflow.Select` with one branch per signal channel. When a signal arrives:

1. The appropriate `handle*` function is called (e.g., `handleFinancialRequest`, `handleDeathNotification`)
2. The handler runs an **idempotency dedup** check via `ProcessedSignalIDs` map (90-day TTL)
3. The handler calls activities to write to DB (e.g., `RecordStateTransitionActivity`)
4. For financial requests: acquires financial lock, routes to child workflow, adds to `PendingRequests`
5. Marks the signal as processed in `ProcessedSignalIDs`

### 4.4 Free-Look Cancellation (FLC) Timer

When a policy is created, `handlePolicyCreated` spawns a goroutine that sleeps until the FLC expiry time (15 or 30 days based on `IsDistanceMarketing`). Because goroutines are lost on Continue-As-New, the expiry time is persisted in `PolicyLifecycleState.FLCExpiryAt`. At the top of every new PLW execution, if `FLCExpiryAt` is non-zero and in the future, the goroutine is respawned with the remaining duration.

### 4.5 Financial Lock

Only one financial request can be in-flight at a time (death claims and NFRs bypass this). The `ActiveLock` field in state tracks the current lock:

- Set in `handleFinancialRequest` when a request is routed
- Cleared in `handleOperationCompleted` (completion), `handleAdminVoid` (pre-clear), or `handleWithdrawal` (cancellation)
- Lock timeout is enforced via the child workflow's `TimeoutAt` field

### 4.6 State Struct Reference

```
PolicyLifecycleState
├── PolicyNumber, PolicyID, PolicyDBID       — Policy identity
├── CurrentStatus, PreviousStatus            — Lifecycle state
├── PreviousStatusBeforeSuspension           — AML revert target
├── Encumbrances (EncumbranceFlags)
│     HasActiveLoan, AssignmentType, AMLHold, DisputeFlag
├── DisplayStatus                            — Computed: status + encumbrance suffixes
│                                              (owned by DB trigger — never set in code)
├── Version                                  — Optimistic lock counter
├── Metadata (PolicyMetadata)                — Premium, dates, product info, distances
├── PendingRequests []PendingRequest         — In-flight routed requests
├── ActiveLock *FinancialLock                — Exclusive financial lock
├── ProcessedSignalIDs map[string]time.Time  — Dedup map (90-day TTL, pruned on CAN)
├── EventCount                               — CAN threshold counter (reset on CAN)
├── CachedConfig map[string]string           — Lazy-loaded workflow config (survives CAN)
└── FLCExpiryAt time.Time                    — Persisted FLC goroutine target (survives CAN)
```

---

## 5. Request Flow Walkthrough

The following traces a financial request (e.g., surrender) end-to-end.

### Step 1 — HTTP Submission

Client `POST /policies/{policyNumber}/surrender` with header `X-Idempotency-Key: <uuid>`.

### Step 2 — Idempotency Check

Handler queries `service_request` by idempotency key. If found → return original `202`. Otherwise continue.

### Step 3 — Policy Lookup

`policyRepo.GetByNumber(policyNumber)` → `404` if not found.

### Step 4 — State Gate Pre-check

Handler calls `client.QueryWorkflow("plw-{policyNumber}", "is-request-eligible", requestType)`.
- Workflow responds `eligible: false` → `422` with reason
- Workflow not found → falls back to DB terminal snapshot → `422` or `404`

### Step 5 — Financial Lock Pre-check

Handler calls `client.QueryWorkflow(..., "get-active-lock")`.
- Lock exists for another request → `409 Conflict`

### Step 6 — Insert `service_request`

`serviceRequestRepo.Create(...)` inserts with `status=RECEIVED`, stores `idempotency_key` and `submitted_at` (partition key).

### Step 7 — Signal PLW

```go
client.SignalWorkflow(ctx, "plw-"+policyNumber, "", "surrender-request", PolicyRequestSignal{
    ServiceRequestID: sr.RequestID,
    IdempotencyKey:   idempotencyKey,
    RequestType:      "SURRENDER",
    RequestCategory:  "FINANCIAL",
    SubmittedAt:      &sr.SubmittedAt,  // partition key carried for DB updates
})
```

### Step 8 — PLW Handles Signal (`handleFinancialRequest`)

Inside the workflow (deterministic, replayed on worker restart):

1. Dedup check (`ProcessedSignalIDs`)
2. `RefreshStateFromDBActivity` — re-reads current status + encumbrances from DB
3. Eligibility check against refreshed in-memory state
4. `LogSignalReceivedActivity` (audit log with PROCESSED or REJECTED outcome)
5. `UpdateServiceRequestActivity` status → `ROUTED` (includes `submitted_at` partition key)
6. Acquire `ActiveLock`
7. `ExecuteChildWorkflow` on `surrender-tq` with `ChildWorkflowInput`
8. Adds `PendingRequest` to state
9. Marks signal as processed

### Step 9 — Handler Returns 202

Handler receives workflow signal ACK and returns `202 Accepted` with `{"request_id": <id>}`.

### Step 10 — Downstream Completes

`SurrenderProcessingWorkflow` on `surrender-tq` processes the surrender, then signals PM:

```go
client.SignalWorkflow(ctx, "plw-"+policyNumber, "", "surrender-completed", OperationCompletedSignal{
    RequestID:       idempotencyKey,
    RequestType:     "SURRENDER",
    Outcome:         "APPROVED",
    StateTransition: "PENDING_SURRENDER→SURRENDERED",
    CompletedAt:     time.Now(),
})
```

### Step 11 — PLW Handles Completion (`handleOperationCompleted`)

1. Dedup check
2. `LogSignalReceivedActivity`
3. Find matching `PendingRequest` by `RequestID`
4. `RecordStateTransitionActivity` — writes `policy_status_history` + updates `policy.current_status`
5. Releases `ActiveLock`
6. Removes from `PendingRequests`
7. `UpdateServiceRequestActivity` status → `COMPLETED`

---

## 6. Batch Scan Workflows

Six Temporal Schedules trigger `BatchStateScanWorkflow` nightly/monthly. All use `SCHEDULE_OVERLAP_POLICY_SKIP` — if a scheduled run is still active when the next fire time arrives, the new run is dropped (prevents double-processing).

| Schedule ID | Scan Type | Cron (IST = UTC+5:30) | Purpose |
|-------------|-----------|----------------------|---------|
| `batch-lapsation-daily` | `LAPSATION` | `30 0 * * *` (00:30) | `ACTIVE` → `VOID_LAPSE` / `INACTIVE_LAPSE` / `VOID` based on payment history and policy age slab |
| `batch-remission-short-daily` | `REMISSION_EXPIRY_SHORT` | `35 0 * * *` (00:35) | `VOID_LAPSE` → `VOID` when remission window expires (policy < 36 months old) |
| `batch-remission-long-daily` | `REMISSION_EXPIRY_LONG` | `40 0 * * *` (00:40) | `INACTIVE_LAPSE` → `ACTIVE_LAPSE` when 12-month remission expires (policy ≥ 36 months) |
| `batch-paidup-monthly` | `PAID_UP_CONVERSION` | `0 1 1 * *` (01:00 1st) | `ACTIVE_LAPSE` → `PAID_UP` (value ≥ ₹10K) or `VOID` (value < ₹10K) |
| `batch-maturity-daily` | `MATURITY_SCAN` | `0 2 * * *` (02:00) | `ACTIVE` → `PENDING_MATURITY` for policies within 90 days of maturity date |
| `batch-forced-surrender-monthly` | `FORCED_SURRENDER_EVAL` | `0 3 1 * *` (03:00 1st) | `ASSIGNED_TO_PRESIDENT` when outstanding loan ≥ 100% of GSV (from surrender-svc) |

Each batch activity:
1. Queries eligible policies from DB in bulk
2. Sends `batch-state-sync` signals to each affected PLW for in-memory sync
3. Writes actual DB state transitions directly from the activity
4. Records a `BatchScanResult` row via `RecordBatchScanResultActivity`

> **Heartbeat requirement:** Batch activities run with `StartToCloseTimeout: 2h` and `HeartbeatTimeout: 5m`. Any new batch activity must call `activity.RecordHeartbeat(ctx, progress)` at least every 5 minutes, otherwise Temporal will cancel it as unresponsive.

---

## 7. Database Schema

All tables live in the `policy_mgmt` schema of the `ims_pli` database.

### 7.1 Key Tables

**`policy`** — Core lifecycle state (3M active / 50M total rows)
- PK: `policy_id BIGINT` (from `seq_policy_id`)
- Natural key: `policy_number TEXT UNIQUE`
- Critical columns: `current_status lifecycle_status`, `display_status TEXT` (trigger-computed), `version BIGINT` (optimistic lock)
- Encumbrance columns: `has_active_loan BOOL`, `assignment_type`, `aml_hold BOOL`, `dispute_flag BOOL`

**`service_request`** — Central request registry (~20K new/day)
- **Partitioned by `submitted_at` (quarterly)**
- PK: composite `(request_id, submitted_at)` — partition key MUST appear in all `WHERE` clauses
- `request_id BIGINT` from `seq_service_request_id`
- `idempotency_key TEXT` — stores `X-Idempotency-Key` header value for dedup

**`policy_status_history`** — Complete audit trail (~10 transitions/policy average)
- **Partitioned by `effective_date` (yearly)**
- PK: composite `(id, effective_date)`
- Records every status transition with `from_status`, `to_status`, `transition_reason`, and `metadata_snapshot JSONB`

**`policy_state_config`** — Workflow configuration key-value store
- Holds routing timeouts, cooling durations, FLC periods, and other workflow-tunable parameters
- Read by `FetchAllWorkflowConfigsActivity` at policy creation; result cached in `PolicyLifecycleState.CachedConfig`

### 7.2 `display_status` — DB Trigger

The `display_status` column on `policy` is **computed by a PostgreSQL trigger** whenever `current_status`, `has_active_loan`, `assignment_type`, `aml_hold`, or `dispute_flag` changes. The format is:

```
{current_status}[_LOAN][_{assignment_type}][_AML_HOLD][_DISPUTED]
```

Example: `ACTIVE_LOAN_CONDITIONAL_AML_HOLD`

**Application code must never `SET display_status = ...` explicitly.** The trigger owns this column.

### 7.3 Optimistic Locking

Every `UPDATE policy` increments `version` and includes `WHERE version = $expected`. A mismatch (0 rows updated) causes the activity to return a conflict error, which Temporal retries.

---

## 8. Key Design Patterns

### 8.1 Two-Tier Policy Query

All policy query endpoints use a two-tier fallback:

```
1. client.QueryWorkflow("plw-{pn}", queryName, ...)
      │
      ├── SUCCESS → return in-memory workflow state (most current, no DB hop)
      │
      └── NOT FOUND / ERROR → fall back to DB
            │
            ├── policyRepo.GetTerminalSnapshot(policyNumber)
            │        → SUCCESS → return snapshot
            └── NOT FOUND → 404 (policy doesn't exist in PM)
```

Active policies get real-time state from the workflow. Terminated policies (whose workflow has ended) are served from the DB snapshot.

### 8.2 Signal Idempotency

Two independent idempotency layers:

| Layer | Mechanism | Scope |
|-------|-----------|-------|
| HTTP handler | `service_request.idempotency_key` DB lookup | Per-request dedup at REST boundary |
| PLW signal handler | `ProcessedSignalIDs map[string]time.Time` | Per-signal dedup inside workflow (90-day TTL) |

The 90-day TTL is enforced by `pruneProcessedSignals(ctx, state)` called on each CAN. It uses `workflow.Now(ctx)` (not `time.Now()`) to remain deterministic during replay.

### 8.3 Financial Lock

Only one financial request can be active per policy at a time:

- `state.ActiveLock *FinancialLock` holds the current exclusive lock
- Set in `handleFinancialRequest`, cleared in `handleOperationCompleted` / `handleAdminVoid` / `handleWithdrawal`
- Death claims and NFRs bypass this lock
- Admin void releases an existing lock before setting state to VOID

### 8.4 Partition-Key Awareness

`service_request` is partitioned quarterly on `submitted_at`. Every `UPDATE service_request WHERE request_id = $1` also includes `AND submitted_at = $2` when the partition key is known (passed via `PendingRequest.SubmittedAt`) to prevent cross-partition sequential scans.

### 8.5 Workflow Config Cache

Configuration (timeouts, cooling durations, FLC periods) is loaded from DB once at policy creation via `FetchAllWorkflowConfigsActivity` and cached in `PolicyLifecycleState.CachedConfig`. The cache survives CAN boundaries (carried in serialised state), so no re-fetch is needed on workflow restart.

### 8.6 Child Workflow Parent-Close Policy

Child workflows are started with `ParentClosePolicy: Abandon`. This means if the PLW undergoes CAN or is terminated, child workflows on downstream task queues continue running independently. PM re-attaches by watching for their per-type completion signals (`surrender-completed`, `loan-completed`, etc.).

### 8.7 Workflow Determinism

The PLW is a Temporal workflow and MUST be deterministic — all replays must produce identical decisions. Rules enforced in this codebase:

- Use `workflow.Now(ctx)` — never `time.Now()`
- Use `workflow.Sleep(ctx, d)` — never `time.Sleep(d)`
- No random values, no goroutine-unsafe maps read across goroutines outside the workflow's single goroutine
- All DB access via Activities (never directly from workflow code)

### 8.8 Audit Logging

Every signal received by PLW is logged via `LogSignalReceivedActivity` with:
- Signal channel name
- Request ID (dedup key)
- Outcome: `PROCESSED`, `REJECTED`, or `DUPLICATE`

This provides a complete audit trail of all inbound signals even when no state transition occurs.

---

## 9. Signal Catalog

### 9.1 Inbound — From REST Handlers to PLW

| Signal Name | Payload Struct | Purpose |
|-------------|---------------|---------|
| `policy-created` | `PolicyCreatedSignal` | Initial lifecycle bootstrap (SignalWithStart from Policy Issue Svc) |
| `surrender-request` | `PolicyRequestSignal` | Route surrender to `surrender-tq` |
| `loan-request` | `PolicyRequestSignal` | Route loan to `loan-tq` |
| `loan-repayment` | `PolicyRequestSignal` | Route loan repayment (no financial lock) |
| `revival-request` | `PolicyRequestSignal` | Route revival to `revival-tq` |
| `death-notification` | `PolicyRequestSignal` | Preemptive — overrides SUSPENDED status |
| `maturity-claim-request` | `PolicyRequestSignal` | Route maturity claim to `claims-tq` |
| `survival-benefit-request` | `PolicyRequestSignal` | Route survival benefit claim |
| `commutation-request` | `PolicyRequestSignal` | Route commutation |
| `conversion-request` | `PolicyRequestSignal` | Route conversion |
| `flc-request` | `PolicyRequestSignal` | Route free-look cancellation to `freelook-tq` |
| `forced-surrender-trigger` | `PolicyRequestSignal` | Triggered by Loan Svc batch |
| `nfr-request` | `PolicyRequestSignal` | All non-financial requests (nomination, address, etc.) |
| `voluntary-paidup-request` | `VoluntaryPaidUpSignal` | Voluntary paid-up (no child workflow) |
| `withdrawal-request` | `WithdrawalRequestSignal` | Cancel active request + release financial lock |
| `admin-void` | `AdminVoidSignal` | Force policy → VOID (BR-PM-073) |
| `reopen-request` | `ReopenRequestSignal` | Exit terminal cooling period |
| `batch-state-sync` | `BatchStateSyncSignal` | In-memory status sync from batch (no DB write) |

### 9.2 Inbound — System / Compliance Signals

| Signal Name | Payload Struct | Purpose |
|-------------|---------------|---------|
| `premium-paid` | `PremiumPaidSignal` | Update `PaidToDate`; may trigger lapse revival |
| `payment-dishonored` | `PaymentDishonoredSignal` | Reverse `PaidToDate`; trigger lapse transition |
| `aml-flag-raised` | `AMLFlagRaisedSignal` | → `SUSPENDED`; saves `PreviousStatusBeforeSuspension` |
| `aml-flag-cleared` | `AMLFlagClearedSignal` | Restore `PreviousStatusBeforeSuspension` |
| `investigation-started` | `InvestigationStartedSignal` | `DEATH_CLAIM_INTIMATED` → `DEATH_UNDER_INVESTIGATION` |
| `investigation-concluded` | `InvestigationConcludedSignal` | → `DEATH_CLAIM_SETTLED` (confirmed) or revert |
| `loan-balance-updated` | `LoanBalanceUpdatedSignal` | Metadata-only update (`LoanOutstanding`) |
| `conversion-reversed` | `ConversionReversedSignal` | `CONVERTED` → `PreviousStatus` (cheque bounce) |
| `customer-id-merge` | `CustomerIDMergeSignal` | Update `customer_id` in policy metadata |
| `dispute-registered` | `DisputeSignal` | Advisory flag — never blocks requests |
| `dispute-resolved` | `DisputeSignal` | Clear dispute advisory flag |

### 9.3 Inbound — Completion Signals (from Downstream Services back to PM)

| Signal Name | From Service | Purpose |
|-------------|-------------|---------|
| `surrender-completed` | Surrender Svc | Voluntary surrender outcome |
| `forced-surrender-completed` | Surrender Svc | Forced surrender outcome |
| `loan-completed` | Loan Svc | Loan processing outcome |
| `loan-repayment-completed` | Loan Svc | Loan repayment outcome |
| `revival-completed` | Revival Svc | Revival processing outcome |
| `claim-settled` | Claims Svc | Death/Maturity/SB claim outcome |
| `commutation-completed` | Commutation Svc | Commutation outcome |
| `conversion-completed` | Conversion Svc | Conversion outcome |
| `flc-completed` | FLC Svc | Free-look cancellation outcome |
| `nfr-completed` | NFS | Non-financial request outcome |
| `operation-completed` | Any | Generic fallback (older integration pattern) |

All completion signals use the `OperationCompletedSignal` payload struct.

### 9.4 Query Handlers (synchronous — no signal, no state change)

| Query Name | Result Struct | Purpose |
|------------|--------------|---------|
| `get-policy-status` | `PolicyStatusQueryResult` | Current + previous status, display status, metadata |
| `get-pending-requests` | `[]PendingRequest` | All in-flight requests |
| `is-request-eligible` | `IsRequestEligibleResult` | State gate check (called by handler before submission) |
| `get-policy-summary` | `PolicyStatusQueryResult` | Same shape as status (summary endpoint) |
| `get-active-lock` | `*FinancialLock` | Current exclusive lock (`nil` if none) |
| `get-status-history` | `[]PolicyStatusHistory` | Recent transitions (read from DB) |
| `get-workflow-health` | `WorkflowHealthResult` | EventCount, CAN time, pending request count |

---

## 10. Integration Contracts

### 10.1 Policy Issue Service → PM (Policy Creation)

Policy Issue Service uses `SignalWithStart` to atomically start the PLW and deliver the first signal:

```go
// WorkflowID: "plw-{policyNumber}"
// Signal channel: "policy-created"
// Workflow input:
StartPMLifecycleInput{
    Signal: PolicyCreatedSignal{
        RequestID:    string        // UUID — idempotency key
        PolicyID:     string        // UUID from Policy Issue (audit cross-ref only)
        PolicyNumber: string        // Human-readable policy number
        Metadata:     PolicyMetadata
    },
    InitialState: PolicyLifecycleState{
        CurrentStatus: "FREE_LOOK_ACTIVE",
        // All other fields zero-valued — PM populates them
    },
}
```

### 10.2 PM → Downstream Services (Request Routing)

PM starts a child workflow on the downstream service's task queue:

```go
ChildWorkflowInput{
    RequestID:        string          // UUID from X-Idempotency-Key
    PolicyNumber:     string
    PolicyDBID:       int64           // BIGINT PM policy_id
    ServiceRequestID: int64           // BIGINT from service_request table
    RequestType:      string          // e.g., "SURRENDER"
    RequestPayload:   json.RawMessage // Original request body (stored as JSONB)
    TimeoutAt:        time.Time       // Deadline for the downstream workflow
}
```

### 10.3 Downstream Services → PM (Completion)

Downstream services signal PM when the operation completes (on signal channel `{type}-completed`):

```go
OperationCompletedSignal{
    RequestID:       string          // UUID — matches ChildWorkflowInput.RequestID
    RequestType:     string          // e.g., "SURRENDER"
    Outcome:         string          // APPROVED | REJECTED | WITHDRAWN | TIMEOUT
    StateTransition: string          // e.g., "PENDING_SURRENDER→SURRENDERED" (optional)
    OutcomePayload:  json.RawMessage // Service-specific result data (optional)
    CompletedAt:     time.Time
}
```

---

## 11. Downstream Service Routing

| Request Type(s) | Task Queue | Child Workflow Type |
|-----------------|------------|---------------------|
| `SURRENDER` | `surrender-tq` | `SurrenderProcessingWorkflow` |
| `FORCED_SURRENDER` | `surrender-tq` | `ForcedSurrenderWorkflow` |
| `LOAN` | `loan-tq` | `LoanProcessingWorkflow` |
| `LOAN_REPAYMENT` | `loan-tq` | `LoanRepaymentWorkflow` |
| `REVIVAL` | `revival-tq` | `InstallmentRevivalWorkflow` |
| `DEATH_CLAIM` | `claims-tq` | `DeathClaimSettlementWorkflow` |
| `MATURITY_CLAIM` | `claims-tq` | `MaturityClaimWorkflow` |
| `SURVIVAL_BENEFIT` | `claims-tq` | `SurvivalBenefitClaimWorkflow` |
| `COMMUTATION` | `commutation-tq` | `CommutationRequestWorkflow` |
| `CONVERSION` | `conversion-tq` | `ConversionMainWorkflow` |
| `FLC` | `freelook-tq` | `FreelookCancellationWorkflow` |
| `NOMINATION_CHANGE`, `BILLING_METHOD_CHANGE`, `ASSIGNMENT`, `ADDRESS_CHANGE`, `DUPLICATE_BOND` | `nfs-tq` | `NFRProcessingWorkflow` |
| `PREMIUM_REFUND` | `billing-tq` | `PremiumRefundWorkflow` |

Child workflow IDs follow the pattern: `{prefix}-{policyNumber}-{idempotencyKey}`.

| Request Type | ID Prefix |
|-------------|-----------|
| `SURRENDER` | `sur` |
| `FORCED_SURRENDER` | `fs` |
| `LOAN` | `loan` |
| `LOAN_REPAYMENT` | `lrp` |
| `REVIVAL` | `rev` |
| `DEATH_CLAIM` | `dc` |
| `MATURITY_CLAIM` | `mc` |
| `SURVIVAL_BENEFIT` | `sb` |
| `COMMUTATION` | `com` |
| `CONVERSION` | `cnv` |
| `FLC` | `flc` |
| All NFR types | `nfr` |
