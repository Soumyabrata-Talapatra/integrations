# Policy Issue Microservice

**India Post PLI/RPLI Insurance Management System**

A Go-based microservice that manages the end-to-end policy issuance lifecycle for Postal Life Insurance (PLI) and Rural Postal Life Insurance (RPLI). Built on the N-API Template framework with Temporal workflows for long-running business processes.

---

## Table of Contents

- [Architecture Overview](#architecture-overview)
- [Tech Stack](#tech-stack)
- [Project Structure](#project-structure)
- [Getting Started](#getting-started)
- [Configuration](#configuration)
- [Database Setup](#database-setup)
- [API Reference](#api-reference)
- [Workflow Engine (Temporal)](#workflow-engine-temporal)
- [Domain Model & State Machine](#domain-model--state-machine)
- [Key Patterns & Conventions](#key-patterns--conventions)
- [Adding a New Feature](#adding-a-new-feature)
- [Business Rules Reference](#business-rules-reference)
- [Troubleshooting](#troubleshooting)

---

## Architecture Overview

```
                    +-------------------+
                    |   API Gateway     |
                    +--------+----------+
                             |
                    +--------v----------+
                    | Policy Issue      |
                    | Microservice      |
                    | (Go / N-API)      |
                    +----+----+----+----+
                         |    |    |
              +----------+    |    +----------+
              |               |               |
     +--------v-----+  +-----v------+  +-----v--------+
     | PostgreSQL    |  | Temporal   |  | External     |
     | (pgx + dblib)|  | Server     |  | Services     |
     +--------------+  +------------+  | - Customer   |
                                       | - KYC        |
                                       | - Billing    |
                                       | - Document   |
                                       | - Notification|
                                       +--------------+
```

### Design Principles

1. **Phase-based DDL separation** -- The core `proposals` table is intentionally minimal. Phase-specific data (indexing, data entry, QC review, approval, issuance) lives in separate tables joined by `proposal_id`.
2. **Uber FX dependency injection** -- All repositories and handlers are wired via FX modules in `bootstrap/bootstrapper.go`. Adding a new dependency means adding it to the appropriate FX module.
3. **Temporal for long-running workflows** -- QC review, medical underwriting, and approval decisions are modeled as Temporal signals. The workflow can wait days/weeks for human decisions.
4. **Config-driven business rules** -- SA-based approval routing, FLC periods, medical thresholds, and validation limits are read from database config tables or `configs/config.yaml`.

---

## Tech Stack

| Component | Technology |
|-----------|-----------|
| Language | Go 1.25+ |
| HTTP Framework | [N-API Server](https://gitlab.cept.gov.in/it-2.0-common/n-api-server) |
| DI Framework | [Uber FX](https://pkg.go.dev/go.uber.org/fx) |
| Database | PostgreSQL 15+ with [pgx/v5](https://github.com/jackc/pgx) |
| Query Builder | [Squirrel](https://github.com/Masterminds/squirrel) (`dblib.Psql`) |
| DB Library | [N-API DB](https://gitlab.cept.gov.in/it-2.0-common/n-api-db) (`dblib`) |
| Workflow Engine | [Temporal](https://temporal.io/) |
| Config | [API Config](https://gitlab.cept.gov.in/it-2.0-common/api-config) |
| Logging | [API Log](https://gitlab.cept.gov.in/it-2.0-common/api-log) |
| API Spec | OpenAPI 3.0 (Swagger) |

---

## Project Structure

```
policy-issue-service/
|
|-- main.go                          # Application entry point
|-- bootstrap/
|   +-- bootstrapper.go             # Uber FX module definitions (DI wiring)
|
|-- configs/
|   +-- config.yaml                 # Application configuration
|
|-- core/                            # Domain layer (no external dependencies)
|   |-- domain/
|   |   |-- proposal.go             # Proposal entity, status enum, state machine
|   |   |-- quote.go                # Quote entity, product, premium rates
|   |   |-- document.go             # Document types, document refs
|   |   |-- aadhaar.go              # Aadhaar session entity
|   |   |-- product.go              # Product catalog, eligibility helpers
|   |   +-- bulk_upload.go          # Bulk upload batch entity
|   +-- port/
|       |-- request.go              # Base request interfaces
|       +-- response.go             # StatusCodeAndMessage, MetaDataResponse
|
|-- handler/                         # HTTP handlers (presentation layer)
|   |-- proposal.go                 # Proposal CRUD, indexing, sections, QC submit
|   |-- quote.go                    # Quote calculation, conversion to proposal
|   |-- aadhaar.go                  # Aadhaar OTP flow, instant issuance
|   |-- approval.go                 # QR approve/reject/return, approver actions
|   |-- policy.go                   # Policy lifecycle, FLC cancellation
|   |-- document.go                 # Document upload, checklist, missing docs
|   |-- validation.go               # Field validations (pincode, IFSC, PAN, etc.)
|   |-- calculation.go              # Premium, GST, rebate, FLC refund calc
|   |-- lookup.go                   # Static lookup endpoints (products, states, etc.)
|   |-- status.go                   # Proposal/policy status and timeline
|   |-- workflow.go                 # Workflow state query, signal endpoints
|   |-- bulk.go                     # Bulk upload proposal processing
|   |-- flc.go                      # FLC queue and processing
|   |-- request.go                  # All request DTOs (shared across handlers)
|   +-- response/                   # Response DTOs organized by handler
|       |-- proposal.go
|       |-- approval.go
|       |-- policy.go
|       |-- document.go
|       +-- ... (one per handler)
|
|-- repo/postgres/                   # Data access layer (PostgreSQL)
|   |-- proposal_repository.go      # Proposal CRUD, status updates, sections, dedup
|   |-- quote_repository.go         # Quote CRUD, premium rates, product lookup
|   |-- product_repository.go       # Product catalog queries
|   |-- document_repository.go      # Document refs, missing documents
|   |-- aadhaar_repository.go       # Aadhaar session management
|   +-- bulk_upload_repository.go   # Bulk batch tracking
|
|-- workflows/                       # Temporal workflow definitions
|   |-- policy_issuance_workflow.go  # WF-PI-001: Standard policy issuance
|   |-- instant_issuance_workflow.go # WF-PI-002: Aadhaar instant issuance
|   +-- activities/
|       |-- proposal_activities.go   # Validation, premium calc, routing, etc.
|       +-- aadhaar_activities.go    # Aadhaar-specific activities
|
|-- db/migrations/                   # SQL migration files
|   |-- 001_policy_issue_schema.sql  # Full schema (24 tables, enums, triggers)
|   +-- 002_add_document_date.sql    # Document date column addition
|
|-- nbf/swagger/
|   +-- policy_issue_swagger.yaml    # OpenAPI 3.0 specification
|
+-- plans/                           # Design documents
    |-- plan.md                      # Implementation plan with phases
    |-- spec.md                      # API specifications with business rules
    +-- context.md                   # Architecture decisions and rationale
```

---

## Getting Started

### Prerequisites

- **Go 1.25+** -- [Install Go](https://go.dev/dl/)
- **PostgreSQL 15+** -- [Install PostgreSQL](https://www.postgresql.org/download/)
- **Temporal Server** -- [Install Temporal](https://docs.temporal.io/self-hosted-guide/setup)

### 1. Clone the Repository

```bash
git clone https://gitlab.cept.gov.in/it-2.0-common/pli-issue.git
cd pli-issue
```

### 2. Install Dependencies

```bash
go mod download
```

### 3. Set Up the Database

```bash
# Create the database
psql -U postgres -c "CREATE DATABASE policy_issue_db;"

# Run migrations
psql -U postgres -d policy_issue_db -f db/migrations/001_policy_issue_schema.sql
psql -U postgres -d policy_issue_db -f db/migrations/002_add_document_date.sql
```

### 4. Start Temporal Server

```bash
# Using Temporal CLI (development mode)
temporal server start-dev --db-filename temporal.db
```

### 5. Configure Environment

```bash
# Minimum required environment variables
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=postgres
export DB_NAME=policy_issue_db
export TEMPORAL_HOST=localhost:7233
export SERVER_PORT=8080
```

Or edit `configs/config.yaml` directly (supports `${ENV_VAR:-default}` syntax).

### 6. Build and Run

```bash
# Build
go build -o policy-issue-service .

# Run
./policy-issue-service

# Or run directly
go run main.go
```

### 7. Verify

```bash
# Health check (N-API built-in)
curl http://localhost:8080/health

# List products
curl http://localhost:8080/v1/products
```

### Quick Verification Commands

```bash
# Build check (must pass with no errors)
go build ./...

# Static analysis (must pass with no warnings)
go vet ./...
```

---

## Configuration

All configuration lives in `configs/config.yaml`. Values support environment variable substitution with defaults:

```yaml
# Key configuration sections:

server:
  port: ${SERVER_PORT:-8080}        # HTTP server port
  mode: ${SERVER_MODE:-development} # development | production

database:
  host: ${DB_HOST:-localhost}       # PostgreSQL host
  query_timeout_low: 2s             # Simple query timeout
  query_timeout_med: 10s            # Complex query timeout

temporal:
  host: ${TEMPORAL_HOST:-localhost:7233}
  task_queue: policy-issue-queue    # Must match worker registration

proposal:
  max_nominees: 3                   # Max nominees per proposal
  flc_period_days_direct: 15        # FLC period for direct channel
  flc_period_days_distance: 30      # FLC period for distance channel

approval:
  level1_sa_max: 500000             # SA <= 5L -> Level 1 approver
  level2_sa_max: 1000000            # SA <= 10L -> Level 2 approver
  # SA > 10L -> Level 3 approver (from approval_routing_config table)
```

### Config Access in Code

```go
// In handlers/repos that receive *config.Config:
timeout := cfg.GetDuration("db.QueryTimeoutLow")    // 2s
maxSA := cfg.GetFloat64("approval.level1_sa_max")    // 500000.0
mode := cfg.GetString("server.mode")                  // "development"
```

---

## Database Setup

### Schema Overview (24 Tables)

The schema uses phase-based table separation:

| Table | Phase | Description |
|-------|-------|-------------|
| `proposals` | Core | Minimal proposal header (status, customer_id, product, SA) |
| `proposal_indexing` | Indexing | Date chain, PO code, first premium tracking |
| `proposal_data_entry` | Data Entry | Section completion flags (8 sections) |
| `proposal_insured` | Data Entry | Insured person details |
| `proposal_nominee` | Data Entry | Nominee details (max 3 per proposal) |
| `proposal_agent` | Data Entry | Servicing agent details |
| `proposal_medical` | Data Entry | Medical information |
| `proposal_qc_review` | QC | QR decision, reviewer, comments |
| `proposal_approval` | Approval | Approver decision, level, routing |
| `proposal_issuance` | Issuance | Policy number, dates, FLC, bond |
| `proposal_missing_documents` | QC/Approval | Missing document notations |
| `proposal_document_ref` | Documents | Uploaded document references |
| `proposal_status_history` | Audit | Full status transition audit trail |
| `product_catalog` | Reference | Product definitions and limits |
| `premium_rate` | Reference | Sankalan premium rate tables |
| `quote` | Quote | Quote calculations and status |
| `aadhaar_session` | Aadhaar | OTP session management |
| `free_look_config` | Config | FLC period rules by channel |
| `approval_routing_config` | Config | SA-based approver routing |
| `bulk_upload_batch` | Bulk | Batch upload tracking |
| `policy_number_sequence` | Sequence | Policy number generation |

### Key DDL Patterns

```sql
-- All tables use a shared sequence for bigint PKs:
SELECT nextval('policy_issue_seq')

-- Phase tables are joined to proposals via:
CONSTRAINT uq_proposal_<phase> UNIQUE (proposal_id)

-- Soft deletes on proposals:
WHERE deleted_at IS NULL

-- Enum types for type safety:
CREATE TYPE proposal_status_enum AS ENUM ('DRAFT', 'INDEXED', 'DATA_ENTRY', ...);
```

### Running Migrations

```bash
# Full schema (includes all tables, enums, indexes, triggers, views)
psql -d policy_issue_db -f db/migrations/001_policy_issue_schema.sql

# Incremental patches
psql -d policy_issue_db -f db/migrations/002_add_document_date.sql
```

---

## API Reference

The full OpenAPI 3.0 specification is at `nbf/swagger/policy_issue_swagger.yaml`.

### API Groups

| Group | Prefix | Endpoints | Handler File |
|-------|--------|-----------|-------------|
| Quote & Products | `/v1/` | 4 | `quote.go` |
| Proposal Core | `/v1/proposals/` | 12 | `proposal.go` |
| Aadhaar Flow | `/v1/proposals/aadhaar/` | 3 | `aadhaar.go` |
| Approval Workflow | `/v1/proposals/:id/` | 5 | `approval.go` |
| Policy Lifecycle | `/v1/policies/` | 3 | `policy.go` |
| Documents | `/v1/proposals/:id/documents/` | 6 | `document.go` |
| Status & Timeline | `/v1/proposals/:id/status` | 4 | `status.go` |
| Validation | `/v1/validate/` | 8 | `validation.go` |
| Calculation | `/v1/calculate/` | 4 | `calculation.go` |
| Lookup | `/v1/lookup/` | 10 | `lookup.go` |
| Workflow | `/v1/workflows/` | 3 | `workflow.go` |
| Bulk Upload | `/v1/bulk-upload/` | 2 | `bulk.go` |

### Key Endpoints

```
# Quote Flow
POST   /v1/quotes/calculate            # Calculate premium quote
POST   /v1/quotes/:quote_id/convert    # Convert quote to proposal

# Proposal Lifecycle
POST   /v1/proposals                    # Create proposal (CPC indexing)
PUT    /v1/proposals/:id/sections/insured-details
PUT    /v1/proposals/:id/sections/nominees
PUT    /v1/proposals/:id/sections/policy-details
PUT    /v1/proposals/:id/sections/agent
PUT    /v1/proposals/:id/sections/medical
POST   /v1/proposals/:id/submit-for-qc  # Triggers Temporal workflow

# QC & Approval (signal-based)
POST   /v1/proposals/:id/qr-approve
POST   /v1/proposals/:id/qr-reject
POST   /v1/proposals/:id/qr-return
POST   /v1/proposals/:id/approve
POST   /v1/proposals/:id/reject

# Policy
GET    /v1/policies/:id
POST   /v1/policies/:id/flc-cancel      # Free Look Cancellation
GET    /v1/policies/:id/flc-status

# Aadhaar Instant Issuance
POST   /v1/proposals/aadhaar/initiate
POST   /v1/proposals/aadhaar/verify-otp
POST   /v1/proposals/aadhaar/submit
```

---

## Workflow Engine (Temporal)

### Registered Workflows

| Workflow | ID | Duration | Signals |
|----------|----|----------|---------|
| `PolicyIssuanceWorkflow` (WF-PI-001) | `pi-{proposal_number}` | Days to weeks | `qr-decision`, `medical-result`, `approver-decision`, `cpc-resubmit` |
| `InstantIssuanceWorkflow` (WF-PI-002) | `ii-{session_id}` | Minutes | None |

### PolicyIssuanceWorkflow Steps

```
1. ValidateProposalActivity      -- Check required fields
2. CheckEligibilityActivity      -- Age, SA, product limits
3. CalculatePremiumActivity      -- Premium from Sankalan tables
4. SavePremiumToProposalActivity -- Persist premium values
5. [QC Review Loop]              -- Wait for qr-decision signal
   - APPROVED  -> continue
   - RETURNED  -> wait for cpc-resubmit, loop back
   - REJECTED  -> terminate
6. RequestMedicalReviewActivity  -- Check if medical needed
   - If needed -> wait for medical-result signal
7. RouteToApproverActivity       -- Query approval_routing_config by SA
   - Wait for approver-decision signal
8. GeneratePolicyNumberActivity  -- From policy_number_sequence table
9. UpdateProposalStatusActivity  -- Set ISSUED
10. GenerateBondActivity         -- Generate policy bond document
11. SendNotificationActivity     -- Notify customer
```

### Sending Signals to Workflows

```bash
# QR Approval (from handler or external system)
POST /v1/proposals/123/qr-approve
{
  "reviewer_id": "456",
  "comments": "All documents verified"
}

# The handler persists the status change AND signals the workflow
```

### Task Queue

The worker listens on `policy-issue-queue` (configured in `bootstrap/bootstrapper.go`). All activities are registered in the `FxTemporal` module.

---

## Domain Model & State Machine

### Proposal Status Flow (BR-POL-015)

```
DRAFT ──> INDEXED ──> DATA_ENTRY ──> QC_PENDING
                                        |
                          +--------------+--------------+
                          |              |              |
                     QC_APPROVED    QC_RETURNED    QC_REJECTED
                          |              |
                          |        DATA_ENTRY (loop)
                          |
                    +-----+------+
                    |            |
              PENDING_MEDICAL  APPROVAL_PENDING
                    |            |
              +-----+-----+   +-+--------+
              |           |   |          |
        MEDICAL_APPROVED  |  APPROVED  REJECTED
              |           |   |
              +---> APPROVAL_PENDING
                         |
                      APPROVED ──> ISSUED ──> DISPATCHED
                                                  |
                                          FREE_LOOK_ACTIVE
                                            |          |
                                         ACTIVE   FLC_CANCELLED
```

### State Transition Validation

The `CanTransitionTo()` method in `core/domain/proposal.go` enforces all valid transitions. Both handlers and the workflow validate transitions before persisting.

### Section Completion (8 Sections)

Before submitting for QC, ALL sections must be complete in `proposal_data_entry`:

1. `insured_details_complete`
2. `nominee_details_complete`
3. `policy_details_complete`
4. `agent_details_complete`
5. `medical_details_complete`
6. `documents_complete`
7. `declaration_complete`
8. `proposer_details_complete`

Additionally, **first premium must be paid** (tracked in `proposal_indexing.first_premium_paid`).

---

## Key Patterns & Conventions

### Handler Pattern

```go
type MyHandler struct {
    *serverHandler.Base               // Embed base handler
    proposalRepo *repo.ProposalRepository
    cfg          *config.Config
}

func NewMyHandler(proposalRepo *repo.ProposalRepository, cfg *config.Config) *MyHandler {
    base := serverHandler.New("MyHandler").SetPrefix("/v1").AddPrefix("")
    return &MyHandler{Base: base, proposalRepo: proposalRepo, cfg: cfg}
}

func (h *MyHandler) Routes() []serverRoute.Route {
    return []serverRoute.Route{
        serverRoute.GET("/my-endpoint", h.MyMethod).Name("My Endpoint"),
    }
}

func (h *MyHandler) MyMethod(sctx *serverRoute.Context, req MyRequest) (*resp.MyResponse, error) {
    // Implementation
}
```

### Repository Pattern (dblib)

```go
// SELECT with Squirrel builder:
query := dblib.Psql.Select("col1", "col2").From("table").Where(sq.Eq{"id": id})
result, err := dblib.SelectOne(ctx, r.db, query, pgx.RowToStructByName[MyStruct])

// SELECT multiple rows:
rows, err := dblib.SelectRows(ctx, r.db, query, pgx.RowToStructByName[MyStruct])

// SELECT with optional result (no error on zero rows):
result, found, err := dblib.SelectOneOK(ctx, r.db, query, pgx.RowTo[int])

// INSERT:
query := dblib.Psql.Insert("table").SetMap(map[string]interface{}{...})
_, err := dblib.Insert(ctx, r.db, query)

// INSERT ... RETURNING:
result, err := dblib.ExecReturn(ctx, r.db, rawSQL, args, pgx.RowToStructByName[T])

// UPDATE:
query := dblib.Psql.Update("table").SetMap(fields).Where(sq.Eq{"id": id})
_, err := dblib.Update(ctx, r.db, query)

// Raw SQL execution:
_, err := dblib.Exec(ctx, r.db, rawSQL, []any{arg1, arg2})
```

### Request/Response Convention

- **Requests**: Defined in `handler/request.go`, implement `Validate() error`
- **Responses**: Defined in `handler/response/<handler>.go`, embed `port.StatusCodeAndMessage`
- **URI params**: Use `uri:"param_name"` struct tag
- **JSON body**: Use `json:"field_name"` struct tag

```go
type MyRequest struct {
    ProposalID int64  `uri:"proposal_id" validate:"required"`
    Comments   string `json:"comments" validate:"required,maxlength=500"`
}
func (r *MyRequest) Validate() error { return nil }
```

### Error Handling

```go
// For "no rows" -- return business response, not error:
if err == pgx.ErrNoRows {
    return &resp.MyResponse{
        StatusCodeAndMessage: port.StatusCodeAndMessage{
            StatusCode: http.StatusNotFound,
            Message:    "Resource not found",
        },
    }, nil
}

// For real DB errors -- return error (framework handles 500):
return nil, err

// For business validation -- return response with 400:
return &resp.MyResponse{
    StatusCodeAndMessage: port.StatusCodeAndMessage{
        StatusCode: http.StatusBadRequest,
        Message:    "ERR-POL-022: Maximum 3 nominees allowed",
    },
}, nil
```

### Context Timeouts

```go
func (r *MyRepo) MyQuery(ctx context.Context, id int64) (*Result, error) {
    ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
    defer cancel()
    // ... query execution
}
```

Use `QueryTimeoutLow` (2s) for simple queries, `QueryTimeoutMed` (10s) for joins/aggregates.

---

## Adding a New Feature

### Step-by-Step Checklist

#### 1. Add Domain Types (if needed)

Edit `core/domain/` -- add entity structs, enums, or helpers:

```go
// core/domain/my_entity.go
type MyEntity struct {
    ID   int64  `db:"id" json:"id"`
    Name string `db:"name" json:"name"`
}
```

#### 2. Add Repository Methods

Edit `repo/postgres/*_repository.go` or create a new one:

```go
func (r *ProposalRepository) MyNewQuery(ctx context.Context, id int64) (*domain.MyEntity, error) {
    ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
    defer cancel()
    query := dblib.Psql.Select("id", "name").From("my_table").Where(sq.Eq{"id": id})
    result, err := dblib.SelectOne(ctx, r.db, query, pgx.RowToStructByName[domain.MyEntity])
    if err != nil { return nil, err }
    return &result, nil
}
```

If creating a new repository, add its constructor to `bootstrap/bootstrapper.go`:
```go
var FxRepo = fx.Module("Repomodule", fx.Provide(
    // ... existing repos
    repo.NewMyRepository,   // <-- add here
))
```

#### 3. Add Request/Response DTOs

Request DTO in `handler/request.go`:
```go
type MyFeatureRequest struct {
    ProposalID int64  `uri:"proposal_id" validate:"required"`
    Value      string `json:"value" validate:"required"`
}
func (r *MyFeatureRequest) Validate() error { return nil }
```

Response DTO in `handler/response/my_feature.go`:
```go
type MyFeatureResponse struct {
    port.StatusCodeAndMessage
    Data string `json:"data"`
}
```

#### 4. Add Handler Method

In the appropriate handler file:
```go
func (h *ProposalHandler) MyFeature(sctx *serverRoute.Context, req MyFeatureRequest) (*resp.MyFeatureResponse, error) {
    // 1. Validate request
    // 2. Check business rules
    // 3. Call repository
    // 4. Return response
}
```

Register the route in `Routes()`:
```go
serverRoute.POST("/proposals/:proposal_id/my-feature", h.MyFeature).Name("My Feature"),
```

#### 5. Add Migration (if schema changes)

Create `db/migrations/003_my_change.sql`:
```sql
ALTER TABLE proposals ADD COLUMN my_column VARCHAR(50);
```

#### 6. Update Swagger

Edit `nbf/swagger/policy_issue_swagger.yaml` to document the new endpoint.

#### 7. Wire into FX (if new handler)

```go
// bootstrap/bootstrapper.go
fx.Annotate(
    handler.NewMyHandler,
    fx.As(new(serverHandler.Handler)),
    fx.ResultTags(serverHandler.ServerControllersGroupTag),
),
```

#### 8. Build & Verify

```bash
go build ./...   # Must pass
go vet ./...     # Must pass
```

---

## Business Rules Reference

### Approval Routing (BR-POL-016)

Queried from `approval_routing_config` table:

| SA Range | Approver Level | Role |
|----------|---------------|------|
| <= 5,00,000 | Level 1 | APPROVER_LEVEL_1 |
| 5,00,001 - 10,00,000 | Level 2 | APPROVER_LEVEL_2 |
| > 10,00,000 | Level 3 | APPROVER_LEVEL_3 |

### Free Look Cancellation (BR-POL-009 / BR-POL-021 / BR-POL-028)

- **Period**: 15-30 days based on channel (from `free_look_config` table)
- **Start date**: Based on `start_date_rule` -- `DISPATCH_DATE`, `DELIVERY_DATE`, or `ISSUE_DATE`
- **Refund formula**: `Refund = Premium Paid - Proportionate Risk - Stamp Duty - Medical Fee`

### Instant Issuance Eligibility (WF-PI-002)

All criteria must be met:
- Aadhaar verified
- Age <= 50 years
- Sum Assured < 20,00,000
- Non-medical (no medical examination required)
- Premium payment completed

If ineligible, falls back to standard workflow (WF-PI-001).

### Nominee Rules

- Maximum 3 nominees per proposal
- Share percentages must total exactly 100%
- Minor nominees require an appointee (name + relationship)

### Deduplication (BR-POL-024)

Before creating a proposal:
1. Check no active proposal exists for the same `customer_id + product_code`
2. On quote conversion, check the quote hasn't already been converted

### Document Requirements

Conditional based on proposal:
- PAN required if SA >= 50,000
- Employment proof required if SA > 10,00,000
- Medical report required if product threshold exceeded
- Proposer docs required when proposer != insured

### Date Chain (BR-POL-018)

Must satisfy: `declaration_date <= receipt_date <= indexing_date <= proposal_date`

---

## Troubleshooting

### Common Issues

**Build fails with missing imports**
```bash
go mod tidy
go mod download
```

**Temporal connection refused**
```
Ensure Temporal server is running on the configured host:port.
Default: localhost:7233
Check: temporal server start-dev
```

**DB connection errors**
```
Verify PostgreSQL is running and credentials match configs/config.yaml.
Check: psql -U postgres -d policy_issue_db -c "SELECT 1"
```

**"column X does not exist" at runtime**
```
Ensure all migration files have been applied in order.
Check: psql -d policy_issue_db -f db/migrations/001_policy_issue_schema.sql
```

**Handler not registered (404 on endpoint)**
```
Ensure the handler is:
1. Added to FxHandler module in bootstrap/bootstrapper.go
2. Implements Routes() returning the endpoint
3. Constructor accepts correct dependencies (FX injects automatically)
```

### Useful Commands

```bash
# Check all routes registered (grep handler Routes methods)
grep -rn "serverRoute\." handler/*.go

# Find all status transitions in workflow
grep -n "UpdateProposalStatusActivity\|UpdateProposalStatus" workflows/*.go handler/*.go

# Check FX dependency graph (compile-time check)
go build ./...

# View Temporal workflows (requires tctl or Temporal UI)
# Default UI: http://localhost:8233
```

---

## Related Documentation

| Document | Location | Description |
|----------|----------|-------------|
| Implementation Plan | `plans/plan.md` | Phase-by-phase implementation guide |
| API Specifications | `plans/spec.md` | Detailed API specs with business rules |
| Architecture Decisions | `plans/context.md` | ADRs and design rationale |
| OpenAPI Spec | `nbf/swagger/policy_issue_swagger.yaml` | Full Swagger documentation |
| DB Schema | `db/migrations/001_policy_issue_schema.sql` | Complete DDL with comments |
