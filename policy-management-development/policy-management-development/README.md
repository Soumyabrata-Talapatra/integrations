# README & Architecture Documentation — Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Create a concise root `README.md` and a detailed `docs/ARCHITECTURE.md` so any developer can understand the service's architecture, setup, and request flow without reading source code first.

**Architecture:** Option C dual-structure — short README.md as entry-point (~250 lines) linking to a deep-dive `docs/ARCHITECTURE.md` (~500 lines). Local dev uses manual prerequisites (PostgreSQL + Temporal installed separately, config override via `configs/config.local.yaml`).

**Tech Stack:** Go 1.26, Temporal SDK v1.37, PostgreSQL 15+, pgx/v5, Uber FX, n-api-\* internal framework, squirrel query builder.

---

## Task 1: Create `README.md`

**Files:**
- Create: `README.md` (project root: `d:/policy-manage/policy-management/README.md`)

**Step 1: Write the file with the content below**

Write the following content exactly to `README.md`:

```markdown
# Policy Management Orchestrator

> **IMS PLI 2.0** — Per-policy lifecycle orchestration microservice for India Post Life Insurance

## What This Service Does

The Policy Management Orchestrator (PM) is the central lifecycle authority for all PLI/RPLI insurance policies in the IMS 2.0 system. It owns every state transition across **23 canonical lifecycle states** (from `FREE_LOOK_ACTIVE` through to terminal states like `MATURED`, `SURRENDERED`, and `DEATH_CLAIM_SETTLED`), exposes **33 REST endpoints** for request submission and policy queries, and orchestrates **20 request types** through downstream specialist services via Temporal child workflows.

Each active policy is backed by a long-running Temporal workflow (`PolicyLifecycleWorkflow`) that runs for the entire policy lifetime — often decades — using Temporal's Continue-As-New (CAN) pattern to manage history size. Six scheduled batch workflows handle nightly/monthly automated state transitions (lapsation, maturity scan, paid-up conversion, forced surrender evaluation).

## Architecture at a Glance

```
  REST Clients
       │
       ▼
 ┌─────────────────────────────────────────────┐
 │         Handler Layer (33 endpoints)         │
 │  policy_request_handler.go     (19 ep)       │
 │  request_lifecycle_handler.go   (5 ep)       │
 │  policy_query_handler.go        (6 ep)       │
 │  quote_handler.go               (3 ep)       │
 │  cpc_lookup_handler.go          (6 ep — static lookup) │
 └────────────────┬────────────────────────────┘
                  │  SignalWorkflow / QueryWorkflow
                  ▼
 ┌─────────────────────────────────────────────┐
 │   Temporal (namespace: pli-insurance)        │
 │   Task Queue: policy-management-tq           │
 │                                              │
 │   PolicyLifecycleWorkflow                   │◀─── 6 Temporal Schedules
 │     per-policy, long-running CAN             │     (lapsation, remission,
 │                                              │      maturity, paid-up,
 │     ──── child workflows ──────────────────▶ │      forced surrender)
 │     (surrender-tq, loan-tq, claims-tq …)    │
 └────────────────┬────────────────────────────┘
                  │  Activities (pgx/squirrel)
                  ▼
 ┌─────────────────────────────────────────────┐
 │   PostgreSQL (schema: policy_mgmt)           │
 │   policy · service_request                  │
 │   policy_status_history · policy_state_config│
 └─────────────────────────────────────────────┘
```

## Prerequisites

| Dependency | Version | Notes |
|------------|---------|-------|
| Go | 1.26+ | |
| PostgreSQL | 15+ | Database: `ims_pli`, schema: `policy_mgmt` |
| Temporal Server | latest | Namespace: `pli-insurance` |

**Local dev tip:** Use [Temporal CLI](https://docs.temporal.io/cli) `temporal server start-dev` for a zero-config local Temporal instance. It starts on `localhost:7233` by default.

## Quick Start

**1. Clone and download dependencies**

```bash
git clone <repo-url>
cd policy-management
go mod download
```

**2. Create a local config override**

```bash
cp configs/config.yaml configs/config.local.yaml
# Edit config.local.yaml — at minimum set db.password
```

**3. Start a local Temporal server**

```bash
temporal server start-dev --namespace pli-insurance --ui-port 8088
# Temporal UI available at http://localhost:8088
```

**4. Create the database and run migrations**

```bash
createdb ims_pli
psql ims_pli -f migrations/001_policy_mgmt_schema.sql
psql ims_pli -f migrations/002_seed_policy_state_config.sql
```

**5. Register Temporal Search Attributes (once per namespace)**

```bash
bash migrations/003_register_temporal_search_attrs.sh
```

**6. Run the service**

```bash
APP_ENV=local go run .
# Service starts on :8080 (configurable in n-api-server)
```

> **Batch schedules** are registered separately via `bootstrap.RegisterBatchSchedules()`. This is intended to be called once per environment (not on every startup). See `migrations/` for the invocation script.

## Running Tests

```bash
go build ./...          # Verify no compile errors
go vet ./...            # Static analysis
go test ./...           # All tests (37 total: handler + workflow + activities)

# Targeted test runs
go test ./workflows/activities/... -v   # Activity tests only
go test -run TestLapsation ./workflows/activities/... -v  # Specific test
```

## Configuration Reference

All defaults are in `configs/config.yaml`. Override per environment in `configs/config.{env}.yaml`.
Set `APP_ENV=<name>` to load `configs/config.<name>.yaml` on top of the base config.

| Key | Default | Description |
|-----|---------|-------------|
| `db.host` | `localhost` | PostgreSQL host |
| `db.port` | `5432` | PostgreSQL port |
| `db.database` | `ims_pli` | Database name |
| `db.schema` | `policy_mgmt` | Schema name |
| `db.username` | `postgres` | DB user |
| `db.password` | `${DB_PASSWORD}` | DB password — use env var |
| `db.maxconns` | `20` | Connection pool max |
| `db.QueryTimeoutLow` | `2s` | Simple lookups |
| `db.QueryTimeoutMed` | `5s` | Batch inserts, aggregations |
| `db.QueryTimeoutHigh` | `30s` | Full-table batch scans |
| `temporal.hostport` | `localhost:7233` | Temporal gRPC address |
| `temporal.namespace` | `pli-insurance` | Temporal namespace |
| `temporal.taskqueue` | `policy-management-tq` | Worker task queue |
| `trace.enabled` | `false` | OpenTelemetry tracing on/off |

## Project Layout

```
policy-management/
├── main.go                          # Entry point — FX app bootstrap
├── bootstrap/
│   └── bootstrapper.go              # FX modules: FxRepo, FxHandler, FxTemporal
│                                    # + RegisterBatchSchedules (6 Temporal Schedules)
├── configs/
│   ├── config.yaml                  # Base config (all environments)
│   └── config.{env}.yaml            # Per-environment overrides (local, dev, staging, prod)
├── core/
│   ├── domain/                      # Constants, lifecycle status enums, DB entity structs
│   └── port/                        # Repository + downstream service interfaces
├── handler/                         # 33 REST endpoints across 5 handler structs
│   └── response/                    # Shared HTTP response types
├── migrations/                      # SQL DDL, seed data, Temporal setup scripts
├── repo/postgres/                   # PostgreSQL repository implementations (pgx/v5 + squirrel)
├── workflows/
│   ├── signals.go                   # Signal/query name constants + all state structs
│   ├── policy_lifecycle_workflow.go # Long-running per-policy Temporal workflow (CAN pattern)
│   ├── batch_scan_workflow.go       # Short-lived batch state transition workflow
│   └── activities/
│       ├── policy_activities.go     # 12 lifecycle activities (DB reads/writes)
│       ├── batch_activities.go      # 7 batch scan activities (lapsation, maturity, etc.)
│       └── quote_activities.go      # 3 quote proxy activities (short-lived workflows)
└── docs/
    ├── ARCHITECTURE.md              # Detailed architecture documentation ← start here
    └── plans/                       # Per-feature design docs and implementation plans
```

## Further Reading

- **[Architecture Deep-Dive](docs/ARCHITECTURE.md)** — State machine, workflow internals, DB schema, all signals, integration contracts, key design patterns
- **[Design Docs](docs/plans/)** — Per-feature design decisions and implementation plans recorded during development
```

**Step 2: Verify the file renders correctly**

Open `README.md` in a Markdown previewer (VS Code: Ctrl+Shift+V). Check:
- [ ] ASCII art diagram aligns and is readable
- [ ] All code blocks close properly (no unclosed triple-backtick)
- [ ] Link to `docs/ARCHITECTURE.md` exists (will resolve once Task 2 is done)
- [ ] Link to `docs/plans/` resolves (directory exists)

**Step 3: Commit**

```bash
git add README.md
git commit -m "docs: add concise entry-point README.md

Covers: service purpose, architecture ASCII diagram, prerequisites,
quick start (6 steps), config reference, project layout, and links
to docs/ARCHITECTURE.md for deep-dive detail.

Co-Authored-By: Claude Sonnet 4.6 <noreply@anthropic.com>"
```

---

## Task 2: Create `docs/ARCHITECTURE.md`

**Files:**
- Create: `docs/ARCHITECTURE.md` (`d:/policy-manage/policy-management/docs/ARCHITECTURE.md`)

**Step 1: Write the file with the content below**

Write the following content exactly to `docs/ARCHITECTURE.md`:

````markdown
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
│     • FxRepo   — provides 4 PostgreSQL repository singletons
│     • FxHandler — registers 5 handler structs (33 total endpoints) with
│                   the n-api-server HTTP server
│     • FxTemporal — dials Temporal client, starts worker on
│                    "policy-management-tq", registers all workflows + activities
│     • RegisterBatchSchedules() — creates 6 Temporal Schedules (idempotent)
│
├── configs/
│     YAML configuration. Base file + per-environment overrides loaded
│     by the n-api-config library based on APP_ENV.
│
├── core/
│   ├── domain/
│   │     All Go constants and entity structs that mirror the DB schema.
│   │     Key files:
│   │       policy.go            — Policy struct, 23 status constants, enums
│   │       service_request.go   — ServiceRequest struct, request type constants
│   │       policy_status_history.go — Audit trail struct
│   │       batch_scan.go        — BatchScanType constants, BatchScanResult
│   └── port/
│         Go interfaces for all external dependencies (repositories,
│         downstream HTTP clients). Keeps domain logic testable.
│
├── handler/
│     Five handler structs — each implements serverHandler.Handler (n-api-server).
│     All handlers share the same 9-step request pattern (see §5).
│       policy_request_handler.go   — 19 endpoints (financial + NFR + admin)
│       request_lifecycle_handler.go — 5 endpoints (list, detail, withdraw, CPC)
│       policy_query_handler.go     — 6 endpoints (status, summary, state-gate, etc.)
│       quote_handler.go            — 3 endpoints (surrender/loan/conversion quotes)
│       cpc_lookup_handler.go       — 6 endpoints (static enum lookups, no DB calls)
│
├── repo/postgres/
│     PostgreSQL implementations of the port interfaces.
│     Uses pgx/v5 for I/O and Masterminds/squirrel as a query builder.
│       policy_repository.go        — policy CRUD + terminal snapshot
│       service_request_repository.go — service_request CRUD (partition-aware)
│       signal_repository.go        — signal audit log writes
│       config_repository.go        — policy_state_config reads
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
                    FLC period expires ─────┤
                                            ▼
                                         ACTIVE ◀──────── AML cleared (restores prev)
                                            │                    ▲
                          AML flag raised ──┼──────────▶ SUSPENDED
                                            │
                 Payment dishonored (< 3yr)─┼──▶ VOID_LAPSE ──(remission expires)──▶ VOID
                 Payment dishonored (≥ 3yr)─┼──▶ INACTIVE_LAPSE ──(expiry)──▶ ACTIVE_LAPSE
                                            │                                        │
                                            │                        batch paid-up ──┤
                                            │                                        ▼
                        Surrender request ──┼──▶ PENDING_SURRENDER          PAID_UP / VOID
                                            │          │ completed
                                            │          ▼
                                            │    SURRENDERED ★
                                            │
                        Death notification ─┼──▶ DEATH_CLAIM_INTIMATED
                                            │          │ investigation starts
                                            │          ▼
                                            │    DEATH_UNDER_INVESTIGATION
                                            │          │ confirmed
                                            │          ▼
                                            │    DEATH_CLAIM_SETTLED ★
                                            │
                      Maturity scan (90d) ──┼──▶ PENDING_MATURITY
                                            │          │ settled
                                            │          ▼
                                            │      MATURED ★
                                            │
                          Admin void ───────┼──▶ VOID ★
                          FLC request ──────┼──▶ FLC_CANCELLED ★
                                            │
                    Forced surrender ───────┼──▶ PENDING_AUTO_SURRENDER
                                            │          │ completed
                                            │          ▼
                                            │    TERMINATED_SURRENDER ★
                                            │
                  Voluntary paid-up ────────┴──▶ PAID_UP or VOID (value < 10K threshold)

★ = Terminal state; workflow enters cooling period then ends via Continue-As-New
```

### 3.3 Terminal State Handling

When a policy reaches a terminal state (`VOID`, `SURRENDERED`, `TERMINATED_SURRENDER`, `MATURED`, `DEATH_CLAIM_SETTLED`, `FLC_CANCELLED`, `CANCELLED_DEATH`, `CONVERTED`), the PLW:

1. Records the transition in DB via `RecordStateTransitionActivity`
2. Cancels any pending child workflows (except if reopen is expected)
3. Enters a **cooling period** (configurable, default 30 days) listening only for `reopen-request` signals
4. After cooling, writes a terminal snapshot to `terminal_state_snapshot` via Continue-As-New, then the workflow ends

---

## 4. Per-Policy Workflow (PLW)

### 4.1 Overview

`PolicyLifecycleWorkflow` in `workflows/policy_lifecycle_workflow.go` is a **long-running Temporal workflow** with ID `plw-{policyNumber}`. It runs for the entire lifetime of a policy — potentially decades.

The workflow's full state is kept in a `PolicyLifecycleState` struct (defined in `workflows/signals.go`) which is serialised and passed across Continue-As-New boundaries.

### 4.2 Continue-As-New (CAN)

Temporal has a hard limit on workflow history size (~50 000 events). PLW uses **Continue-As-New** to reset history while preserving all state:

- An `EventCount` field in `PolicyLifecycleState` increments on every signal received
- When `EventCount` reaches a threshold (configurable, default ~500), the PLW serialises its state and calls `workflow.NewContinueAsNewError()`, passing the full state as input to a fresh execution
- `EventCount` is reset to 0 at the start of each new execution

### 4.3 Signal Processing

The PLW runs a main loop that blocks on multiple `workflow.Go` goroutines and `workflow.Select` branches, one per signal channel. When a signal arrives:

1. The appropriate `handle*` function is called (e.g., `handleFinancialRequest`, `handleDeathNotification`)
2. The handler runs **idempotency dedup** check via `ProcessedSignalIDs` map (90-day TTL)
3. The handler calls activities to write to DB (e.g., `RecordStateTransitionActivity`)
4. For financial requests: acquires financial lock, routes to child workflow, adds to `PendingRequests`
5. Marks the signal as processed in `ProcessedSignalIDs`

### 4.4 Free-Look Cancellation (FLC) Timer

When a policy is created, `handlePolicyCreated` spawns a goroutine that sleeps until the FLC expiry time (15 or 30 days based on `IsDistanceMarketing`). Because goroutines are lost on Continue-As-New, the expiry time is persisted in `PolicyLifecycleState.FLCExpiryAt`. At the top of every new PLW execution, if `FLCExpiryAt` is set and in the future, the goroutine is respawned with the remaining duration.

### 4.5 Financial Lock

Only one financial request can be in-flight at a time (except death claims, which preempt). The `ActiveLock` field in state tracks the current lock. On every new financial request, the handler checks for an existing lock and returns a `409 Conflict` response to the caller if one is active.

### 4.6 State Struct

```
PolicyLifecycleState
├── PolicyNumber, PolicyID, PolicyDBID       — Policy identity
├── CurrentStatus, PreviousStatus            — Lifecycle state
├── PreviousStatusBeforeSuspension           — AML revert target
├── Encumbrances (EncumbranceFlags)          — Loan, Assignment, AML, Dispute flags
├── DisplayStatus                            — Computed: status + encumbrance suffixes
├── Version                                  — Optimistic lock counter
├── Metadata (PolicyMetadata)                — Premium, dates, product info
├── PendingRequests []PendingRequest         — In-flight routed requests
├── ActiveLock *FinancialLock                — Exclusive financial lock
├── ProcessedSignalIDs map[string]time.Time  — Dedup map (90-day TTL)
├── EventCount                               — CAN threshold counter
├── CachedConfig map[string]string           — Lazy-loaded workflow config
└── FLCExpiryAt time.Time                    — Persisted FLC goroutine target
```

---

## 5. Request Flow Walkthrough

The following traces a financial request (e.g., surrender) end-to-end.

### Step 1 — HTTP Submission

Client `POST /policies/{policyNumber}/surrender` with header `X-Idempotency-Key: <uuid>`.

### Step 2 — Idempotency Check

Handler queries `service_request` by idempotency key. If found → return original 202. Otherwise continue.

### Step 3 — Policy Lookup

`policyRepo.GetByNumber(policyNumber)` → 404 if not found.

### Step 4 — State Gate Pre-check

Handler calls `client.QueryWorkflow("plw-{policyNumber}", "is-request-eligible", requestType)`.
- If workflow responds `eligible: false` → 422 with reason
- If workflow not found → falls back to DB terminal snapshot → 422 or 404

### Step 5 — Financial Lock Pre-check

Handler calls `client.QueryWorkflow(..., "get-active-lock")`.
- If lock exists for another request → 409 Conflict

### Step 6 — Insert service_request

`serviceRequestRepo.Create(...)` inserts with `status=RECEIVED`, stores `idempotency_key` and `submitted_at` (partition key).

### Step 7 — Signal PLW

```go
client.SignalWorkflow(ctx, "plw-"+policyNumber, "", "surrender-request", PolicyRequestSignal{
    ServiceRequestID: sr.RequestID,
    IdempotencyKey:   idempotencyKey,
    RequestType:      "SURRENDER",
    RequestCategory:  "FINANCIAL",
    SubmittedAt:      &sr.SubmittedAt,
})
```

### Step 8 — PLW Handles Signal (`handleFinancialRequest`)

Inside the workflow:
1. Dedup check (ProcessedSignalIDs)
2. `RefreshStateFromDBActivity` — re-reads current status + encumbrances from DB
3. Eligibility check against in-memory state
4. `LogSignalReceivedActivity` (audit log)
5. `UpdateServiceRequestActivity` status → ROUTED (includes `submitted_at` partition key)
6. Acquire `ActiveLock`
7. `ExecuteChildWorkflow` on `surrender-tq` with `ChildWorkflowInput`
8. Adds to `PendingRequests`
9. Marks signal as processed

### Step 9 — Handler Returns 202

Handler receives workflow signal ACK and returns `202 Accepted` with `{"request_id": <id>}`.

### Step 10 — Downstream Completes

The `SurrenderProcessingWorkflow` on `surrender-tq` processes the surrender, then signals PM back:

```
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
3. Find matching `PendingRequest`
4. `RecordStateTransitionActivity` — writes policy_status_history + updates policy.current_status
5. Releases `ActiveLock`
6. Removes from `PendingRequests`
7. `UpdateServiceRequestActivity` status → COMPLETED

---

## 6. Batch Scan Workflows

Six Temporal Schedules trigger `BatchStateScanWorkflow` nightly/monthly. All schedules use `SCHEDULE_OVERLAP_POLICY_SKIP` so concurrent runs are dropped rather than queued.

| Schedule ID | Scan Type | Cron (IST) | Purpose |
|-------------|-----------|------------|---------|
| `batch-lapsation-daily` | `LAPSATION` | `30 0 * * *` (00:30) | ACTIVE → VOID_LAPSE / INACTIVE_LAPSE / VOID based on payment history and slab |
| `batch-remission-short-daily` | `REMISSION_EXPIRY_SHORT` | `35 0 * * *` (00:35) | VOID_LAPSE → VOID when remission window expires (< 36 months) |
| `batch-remission-long-daily` | `REMISSION_EXPIRY_LONG` | `40 0 * * *` (00:40) | INACTIVE_LAPSE → ACTIVE_LAPSE when 12-month remission expires (≥ 36 months) |
| `batch-paidup-monthly` | `PAID_UP_CONVERSION` | `0 1 1 * *` (01:00 1st) | ACTIVE_LAPSE → PAID_UP (value ≥ ₹10K) or VOID (value < ₹10K) |
| `batch-maturity-daily` | `MATURITY_SCAN` | `0 2 * * *` (02:00) | ACTIVE → PENDING_MATURITY for policies within 90 days of maturity date |
| `batch-forced-surrender-monthly` | `FORCED_SURRENDER_EVAL` | `0 3 1 * *` (03:00 1st) | ASSIGNED_TO_PRESIDENT when outstanding loan ≥ 100% of GSV |

Each batch activity:
- Queries eligible policies from DB in bulk
- Sends `batch-state-sync` signals to each affected PLW (in-memory sync only — no DB write from signal handler)
- Writes the actual DB state transition directly from the activity (not through the workflow)
- Records a `BatchScanResult` row via `RecordBatchScanResultActivity`

---

## 7. Database Schema

All tables live in the `policy_mgmt` schema of the `ims_pli` database.

### 7.1 Key Tables

**`policy`** — Core lifecycle state (3M active / 50M total rows)
- PK: `policy_id BIGINT` (from `seq_policy_id`)
- Natural key: `policy_number TEXT UNIQUE`
- Critical columns: `current_status`, `display_status` (trigger-computed), `version` (optimistic lock)
- Encumbrance columns: `has_active_loan`, `assignment_type`, `aml_hold`, `dispute_flag`

**`service_request`** — Central request registry (~20K new/day)
- **Partitioned by `submitted_at` (quarterly)**
- PK: composite `(request_id, submitted_at)` — partition key MUST appear in all WHERE clauses
- `request_id BIGINT` from `seq_service_request_id`
- `idempotency_key` — stores X-Idempotency-Key header for dedup

**`policy_status_history`** — Complete audit trail (~10 transitions/policy avg)
- **Partitioned by `effective_date` (yearly)**
- PK: composite `(id, effective_date)`
- Records every status transition with from/to status, reason, and metadata snapshot

**`policy_state_config`** — Workflow configuration key-value store
- Holds routing timeouts, cooling durations, FLC periods, and other workflow-tunable parameters
- Read by `FetchAllWorkflowConfigsActivity` at policy creation and cached in workflow state

### 7.2 Display Status — DB Trigger

The `display_status` column on `policy` is **computed by a PostgreSQL trigger** whenever `current_status`, `has_active_loan`, `assignment_type`, `aml_hold`, or `dispute_flag` changes. The format is:

```
{current_status}[_LOAN][_{assignment_type}][_AML_HOLD][_DISPUTED]
```

For example: `ACTIVE_LOAN_CONDITIONAL_AML_HOLD`

**Important:** Application code must never `SET display_status = ...` explicitly. The trigger owns this column.

### 7.3 Optimistic Locking

Every UPDATE to the `policy` table increments `version` and includes `WHERE version = $expected` in the WHERE clause. A mismatch returns 0 rows updated, which the activity treats as a conflict error.

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

This ensures active policies get real-time state from the workflow while terminated policies (whose workflow has ended) are served from the DB snapshot.

### 8.2 Idempotency

Two independent idempotency layers:

| Layer | Mechanism | Scope |
|-------|-----------|-------|
| HTTP handler | `service_request.idempotency_key` DB lookup | Per-request dedup at REST boundary |
| PLW signal handler | `ProcessedSignalIDs map[string]time.Time` | Per-signal dedup inside workflow (90-day TTL) |

The 90-day TTL is enforced by `pruneProcessedSignals(ctx, state)` called on each CAN.

### 8.3 Financial Lock

Only one financial request can be active per policy at a time:

- `state.ActiveLock *FinancialLock` holds the current exclusive lock
- Set in `handleFinancialRequest`, cleared in `handleOperationCompleted` / `handleAdminVoid` / `handleWithdrawal`
- Death claims and NFRs bypass this lock (they don't set `ActiveLock`)
- Lock timeout enforced by the child workflow's `TimeoutAt` field

### 8.4 Partition-Key Awareness

`service_request` is partitioned quarterly on `submitted_at`. Every `UPDATE service_request WHERE request_id = $1` also includes `AND submitted_at = $2` when the partition key is known (passed via `PendingRequest.SubmittedAt`) to prevent cross-partition sequential scans.

### 8.5 Config Cache

Workflow configuration (timeouts, cooling durations, FLC periods) is loaded from the DB once at policy creation via `FetchAllWorkflowConfigsActivity` and cached in `PolicyLifecycleState.CachedConfig`. On CAN, the cached map is carried across in the serialised state, so no re-fetch is needed.

### 8.6 Child Workflow Lifecycle

Child workflows are started with `ParentClosePolicy: Abandon`. This means if the PLW undergoes CAN or is terminated, child workflows (on `surrender-tq`, `loan-tq`, etc.) continue running independently. PM re-attaches to them by watching for their completion signals.

---

## 9. Signal Catalog

### Inbound — From REST Handlers to PLW

| Signal Name | Payload Struct | Direction | Purpose |
|-------------|---------------|-----------|---------|
| `policy-created` | `PolicyCreatedSignal` | Policy Issue Svc → PM | Initial lifecycle bootstrap (SignalWithStart) |
| `surrender-request` | `PolicyRequestSignal` | Handler → PLW | Route surrender to surrender-tq |
| `loan-request` | `PolicyRequestSignal` | Handler → PLW | Route loan to loan-tq |
| `loan-repayment` | `PolicyRequestSignal` | Handler → PLW | Route loan repayment (no financial lock) |
| `revival-request` | `PolicyRequestSignal` | Handler → PLW | Route revival to revival-tq |
| `death-notification` | `PolicyRequestSignal` | Handler → PLW | Preemptive — overrides SUSPENDED |
| `maturity-claim-request` | `PolicyRequestSignal` | Handler → PLW | Route maturity claim |
| `survival-benefit-request` | `PolicyRequestSignal` | Handler → PLW | Route survival benefit claim |
| `commutation-request` | `PolicyRequestSignal` | Handler → PLW | Route commutation |
| `conversion-request` | `PolicyRequestSignal` | Handler → PLW | Route conversion |
| `flc-request` | `PolicyRequestSignal` | Handler → PLW | Route free-look cancellation |
| `forced-surrender-trigger` | `PolicyRequestSignal` | Loan Svc batch → PLW | Trigger forced surrender |
| `nfr-request` | `PolicyRequestSignal` | Handler → PLW | All non-financial requests |
| `voluntary-paidup-request` | `VoluntaryPaidUpSignal` | Handler → PLW | Voluntary paid-up (no child WF) |
| `withdrawal-request` | `WithdrawalRequestSignal` | Handler → PLW | Cancel active request + release lock |
| `admin-void` | `AdminVoidSignal` | Handler → PLW | Force → VOID (BR-PM-073) |
| `reopen-request` | `ReopenRequestSignal` | Handler → PLW | Exit terminal cooling |
| `batch-state-sync` | `BatchStateSyncSignal` | Batch activity → PLW | In-memory status sync (no DB write) |

### Inbound — System / Compliance Signals

| Signal Name | Payload Struct | Purpose |
|-------------|---------------|---------|
| `premium-paid` | `PremiumPaidSignal` | Update PaidToDate; may trigger lapse revival |
| `payment-dishonored` | `PaymentDishonoredSignal` | Reverse PaidToDate; trigger lapse transition |
| `aml-flag-raised` | `AMLFlagRaisedSignal` | → SUSPENDED; save previous status |
| `aml-flag-cleared` | `AMLFlagClearedSignal` | Restore PreviousStatusBeforeSuspension |
| `investigation-started` | `InvestigationStartedSignal` | DEATH_CLAIM_INTIMATED → DEATH_UNDER_INVESTIGATION |
| `investigation-concluded` | `InvestigationConcludedSignal` | → DEATH_CLAIM_SETTLED (confirmed) or revert |
| `loan-balance-updated` | `LoanBalanceUpdatedSignal` | Metadata-only update (LoanOutstanding) |
| `conversion-reversed` | `ConversionReversedSignal` | CONVERTED → PreviousStatus (cheque bounce) |
| `customer-id-merge` | `CustomerIDMergeSignal` | Update customer_id in metadata |
| `dispute-registered` | `DisputeSignal` | Advisory flag — never blocks requests |
| `dispute-resolved` | `DisputeSignal` | Clear dispute advisory flag |

### Inbound — Completion Signals (from Downstream Services back to PM)

| Signal Name | Payload Struct | From |
|-------------|---------------|------|
| `surrender-completed` | `OperationCompletedSignal` | Surrender Svc |
| `forced-surrender-completed` | `OperationCompletedSignal` | Surrender Svc |
| `loan-completed` | `OperationCompletedSignal` | Loan Svc |
| `loan-repayment-completed` | `OperationCompletedSignal` | Loan Svc |
| `revival-completed` | `OperationCompletedSignal` | Revival Svc |
| `claim-settled` | `OperationCompletedSignal` | Claims Svc |
| `commutation-completed` | `OperationCompletedSignal` | Commutation Svc |
| `conversion-completed` | `OperationCompletedSignal` | Conversion Svc |
| `flc-completed` | `OperationCompletedSignal` | FLC Svc |
| `nfr-completed` | `OperationCompletedSignal` | NFS |
| `operation-completed` | `OperationCompletedSignal` | Generic fallback (older pattern) |

### Query Handlers (synchronous — no signal)

| Query Name | Result Struct | Purpose |
|------------|--------------|---------|
| `get-policy-status` | `PolicyStatusQueryResult` | Current + previous status, metadata |
| `get-pending-requests` | `[]PendingRequest` | All in-flight requests |
| `is-request-eligible` | `IsRequestEligibleResult` | State gate check before submission |
| `get-policy-summary` | `PolicyStatusQueryResult` | Alias for status (used by summary endpoint) |
| `get-active-lock` | `*FinancialLock` | Current exclusive lock (nil if none) |
| `get-status-history` | `[]PolicyStatusHistory` | Recent transitions (from DB) |
| `get-workflow-health` | `WorkflowHealthResult` | EventCount, CAN time, pending count |

---

## 10. Integration Contracts

### 10.1 Policy Issue Service → PM (Policy Creation)

Policy Issue Service uses `SignalWithStart` to atomically start the PLW and deliver the first signal:

```go
// Input to SignalWithStart
StartPMLifecycleInput {
    Signal: PolicyCreatedSignal {
        RequestID:    string    // UUID — idempotency key
        PolicyID:     string    // UUID from Policy Issue (audit cross-ref only)
        PolicyNumber: string    // Human-readable policy number
        Metadata:     PolicyMetadata
    }
    InitialState: PolicyLifecycleState {
        CurrentStatus: "FREE_LOOK_ACTIVE"
        // All other fields zero-valued — PM populates them
    }
}
```

### 10.2 PM → Downstream Services (Request Routing)

PM starts a child workflow on the downstream service's task queue:

```go
ChildWorkflowInput {
    RequestID:        string          // UUID from X-Idempotency-Key
    PolicyNumber:     string
    PolicyDBID:       int64           // BIGINT PM policy_id
    ServiceRequestID: int64           // BIGINT from service_request table
    RequestType:      string          // e.g., "SURRENDER"
    RequestPayload:   json.RawMessage // Original request body (JSONB)
    TimeoutAt:        time.Time       // Deadline for the downstream workflow
}
```

### 10.3 Downstream Services → PM (Completion)

Downstream services signal PM when the operation completes:

```go
OperationCompletedSignal {
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

| Request Type(s) | Task Queue | Workflow Type |
|-----------------|------------|---------------|
| `SURRENDER`, `FORCED_SURRENDER` | `surrender-tq` | `SurrenderProcessingWorkflow` / `ForcedSurrenderWorkflow` |
| `LOAN`, `LOAN_REPAYMENT` | `loan-tq` | `LoanProcessingWorkflow` / `LoanRepaymentWorkflow` |
| `REVIVAL` | `revival-tq` | `InstallmentRevivalWorkflow` |
| `DEATH_CLAIM`, `MATURITY_CLAIM`, `SURVIVAL_BENEFIT` | `claims-tq` | `DeathClaimSettlementWorkflow` / `MaturityClaimWorkflow` / `SurvivalBenefitClaimWorkflow` |
| `COMMUTATION` | `commutation-tq` | `CommutationRequestWorkflow` |
| `CONVERSION` | `conversion-tq` | `ConversionMainWorkflow` |
| `FLC` | `freelook-tq` | `FreelookCancellationWorkflow` |
| `NOMINATION_CHANGE`, `BILLING_METHOD_CHANGE`, `ASSIGNMENT`, `ADDRESS_CHANGE`, `DUPLICATE_BOND` | `nfs-tq` | `NFRProcessingWorkflow` |
| `PREMIUM_REFUND` | `billing-tq` | `PremiumRefundWorkflow` |

Child workflow IDs follow the pattern: `{prefix}-{policyNumber}-{idempotencyKey}`.
Prefixes: `sur`, `fs`, `loan`, `lrp`, `rev`, `dc`, `mc`, `sb`, `com`, `cnv`, `flc`, `nfr`.
````

**Step 2: Verify accuracy**

Cross-check the following against source code:
- [ ] All 23 status constants match `core/domain/policy.go` exactly
- [ ] All signal name strings match `workflows/signals.go` constants
- [ ] All 6 batch schedule IDs and cron expressions match `bootstrap/bootstrapper.go`
- [ ] Task queue names match `core/domain/service_request.go` `DownstreamTaskQueueForType()`
- [ ] Workflow type names match `workflows/signals.go` `DownstreamWorkflowTypeForRequest()`
- [ ] Child ID prefixes match `workflows/signals.go` `DownstreamChildIDPrefix()`

**Step 3: Commit**

```bash
git add docs/ARCHITECTURE.md
git commit -m "docs: add comprehensive ARCHITECTURE.md

Covers: system context, package structure, 23-state machine (with
ASCII diagram), PLW internals (CAN, FLC timer, financial lock),
request flow walkthrough (11 steps), batch schedules, DB schema
(partitioned tables, trigger-computed display_status), 8 design
patterns, full signal catalog (40 signals), integration contracts,
and downstream routing table.

Co-Authored-By: Claude Sonnet 4.6 <noreply@anthropic.com>"
```

---

## Task 3: Update plan.md

**Files:**
- Modify: `d:\policy-manage\.zencoder\chats\c4b9a45b-ccf6-471d-929f-1924d4d43fbf\plan.md`

**Step 1: Add new step to plan.md**

Append the following new step to plan.md (after the last Workstream D section):

```markdown
### [ ] Step: Implementation — README & Architecture Documentation

#### [ ] E1: Create root README.md
- File: `README.md`
- Concise entry-point: purpose, ASCII architecture diagram, prerequisites, quick start (6 steps), config reference, project layout, further reading links

#### [ ] E2: Create docs/ARCHITECTURE.md
- File: `docs/ARCHITECTURE.md`
- Deep-dive: system context, package structure, 23-state machine, PLW internals (CAN/FLC/lock), request flow walkthrough, batch schedules, DB schema, 8 design patterns, signal catalog, integration contracts, downstream routing

### [ ] Step: Verification (README & Architecture)

- [ ] README.md renders without broken Markdown
- [ ] ARCHITECTURE.md all state names match `core/domain/policy.go`
- [ ] ARCHITECTURE.md all signal names match `workflows/signals.go`
- [ ] ARCHITECTURE.md all task queues match `core/domain/service_request.go`
- [ ] `go build ./...` — still clean (no code changes)
```

**Step 2: Commit plan.md**

```bash
git add "d:\policy-manage\.zencoder\chats\c4b9a45b-ccf6-471d-929f-1924d4d43fbf\plan.md"
git commit -m "chore: add README docs workstream to plan.md

Co-Authored-By: Claude Sonnet 4.6 <noreply@anthropic.com>"
```

---

## Task 4: Final Verification

**Step 1: Confirm both files exist**

```bash
ls docs/ARCHITECTURE.md README.md
```

Expected output: both files listed.

**Step 2: Confirm build still clean (no code was changed)**

```bash
go build ./...
go vet ./...
```

Expected: no output (clean).

**Step 3: Word count sanity check**

```bash
wc -l README.md docs/ARCHITECTURE.md
```

Expected: README ~200-260 lines, ARCHITECTURE ~480-560 lines.
