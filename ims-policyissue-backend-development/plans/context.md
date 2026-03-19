# Policy Issue Microservice - Technical Context

## Project Context

**Service**: Policy Issue Microservice  
**Created**: 2026-02-13  
**Status**: Planning Phase  

---

## Key Architectural Decisions

### 1. Database Design Pattern: Phase-Based Tables

**Decision**: Use normalized phase-based tables instead of single monolithic proposals table.

**Rationale**:
- Reduces row size and lock contention
- Allows independent updates at different lifecycle stages
- Cleaner separation of concerns

**Tables**:
- `proposals` - Core shared data (customer_id, status, product_code, SA)
- `proposal_indexing` - Indexing phase data (po_code, dates, channel)
- `proposal_data_entry` - Data entry status tracking
- `proposal_qc_review` - QR assignment and decision
- `proposal_medical` - Medical examination tracking
- `proposal_approval` - Approver routing and decisions
- `proposal_issuance` - Policy number, bond, dispatch info

**Cross-Reference**: See `nbf/dbscripts/policy_issue_schema.sql` lines 277-350

---

### 2. Temporal Workflow Strategy

**Decision**: Use Temporal for ALL proposal processing, not just long-running.

**Workflows**:

| Workflow | Use Case | Duration | Signals |
|----------|----------|----------|---------|
| WF-PI-001 | Standard Policy Issuance | Days-Weeks | QRDecision, ApproverDecision, MedicalResult, PaymentReceived, CPCResubmit |
| WF-PI-002 | Instant Issuance (Aadhaar) | Minutes | None |
| WF-PI-003 | Bulk Proposal Upload | Hours | ProgressUpdate |

**Rationale**:
- Consistent error handling and retry logic
- Built-in audit trail via workflow history
- Easy to add human approval steps
- Saga pattern for compensation

**Cross-Reference**: See `nbf/requirements/policy_issue_requirements.md` lines 1956-2200

---

### 3. Workflow State Optimization

**Decision**: Fetch all required data ONCE in initial activity, store in workflow state.

**Pattern**:
```go
type ProposalWorkflowState struct {
    ProposalID       string
    CustomerID       string
    PolicyData       *domain.Proposal
    CustomerData     *domain.CustomerData
    MedicalData      *domain.MedicalInfo
    CalculationResult *domain.PremiumCalculation
}
```

**Optimization Impact**:
- Without state: 8+ DB calls across activities
- With state: 1 pgx.Batch call, 0 subsequent DB calls

**Reference**: template.md Section 18

---

### 4. pgx.Batch Usage Rules

**Decision**: Use pgx.Batch for ALL multi-query operations.

**Criteria**:
- 2+ queries in same operation → Use Batch
- Single query → Direct execution acceptable

**Example Use Cases**:
- FetchInitialData: proposals + proposal_indexing + proposal_nominees
- CreateProposal: Insert to proposals + proposal_indexing + proposal_status_history
- GetProposalDetail: Core + all phase tables

**Reference**: `nbf/template.md` Section 16, `database-library.md`

---

### 5. SQL Type Decision Matrix

| Query Type | Tool | Example |
|------------|------|---------|
| Simple SELECT/INSERT/UPDATE | Squirrel (dblib.Psql) | GetProposalByID, UpdateStatus |
| INSERT...SELECT | Raw SQL | Create proposal from quote |
| UPDATE...FROM | Raw SQL | Bulk status updates |
| WITH (CTE) | Raw SQL | Recursive hierarchy queries |
| Dynamic filters | Squirrel | Proposal queue with filters |

---

### 6. Repository Pattern

**Decision**: No interfaces for repositories - use concrete structs.

**Pattern**:
```go
type ProposalRepository struct {
    db *dblib.DB
}

func NewProposalRepository(db *dblib.DB) *ProposalRepository {
    return &ProposalRepository{db: db}
}
```

**Rationale**: 
- Simpler dependency injection with FX
- No mocking needed (use test DB)
- Less boilerplate

**Cross-Reference**: `n-api-template.md` Section 6

---

### 7. State Machine Implementation

**Decision**: Database-level state transition validation via trigger.

**Trigger**: `trg_proposal_workflow_transition`

**Valid Transitions**:
```
DRAFT → INDEXED, CANCELLED_DEATH
INDEXED → DATA_ENTRY
DATA_ENTRY → QC_PENDING
QC_PENDING → QC_APPROVED, QC_REJECTED, QC_RETURNED
QC_RETURNED → DATA_ENTRY
QC_APPROVED → PENDING_MEDICAL, APPROVAL_PENDING
PENDING_MEDICAL → MEDICAL_APPROVED, MEDICAL_REJECTED
MEDICAL_APPROVED → APPROVAL_PENDING
APPROVAL_PENDING → APPROVED, REJECTED
APPROVED → ISSUED
ISSUED → DISPATCHED
DISPATCHED → FREE_LOOK_ACTIVE
FREE_LOOK_ACTIVE → ACTIVE, FLC_CANCELLED
```

**Cross-Reference**: `nbf/dbscripts/policy_issue_schema.sql` lines 1048-1073

---

### 8. Approval Routing Strategy

**Decision**: Configurable approval routing by Sum Assured.

**Configuration Table**: `approval_routing_config`

| SA Range | Level | Role |
|----------|-------|------|
| ≤ ₹5,00,000 | L1 | APPROVER_LEVEL_1 |
| ₹5,00,001 - ₹10,00,000 | L2 | APPROVER_LEVEL_2 |
| > ₹10,00,000 | L3 | APPROVER_LEVEL_3 |

**Business Rule**: BR-POL-016

---

### 9. Date Validation Chain

**Decision**: Enforce date sequence at API layer and DB trigger.

**Rule**: BR-POL-018
```
declaration_date ≤ receipt_date ≤ indexing_date ≤ proposal_date
```

**Error Codes**:
- ERR-POL-006: Invalid declaration date
- ERR-POL-007: Invalid receipt date
- ERR-POL-008: Invalid indexing date
- ERR-POL-009: Invalid proposal date

---

### 10. Customer ID Handling

**Decision**: customer_id is MANDATORY at proposal indexing.

**Flow**:
1. Customer identified/created before proposal indexing
2. customer_id passed in ProposalIndexingRequest
3. Proposal service does NOT create customers (Customer Service responsibility)

**Integration**: INT-POL-002 (Customer Service)

---

### 11. Primary Key Type: BIGINT vs UUID

**Decision**: Use BIGINT (auto-increment) for all primary keys instead of UUID.

**Rationale**:
- Better performance for PostgreSQL indexes and joins
- Smaller storage footprint (8 bytes vs 16 bytes)
- Human-readable and sequentially ordered
- Easier for support/debugging (shorter IDs)
- Better compatibility with existing legacy systems

**Implementation**:
```sql
CREATE TABLE quote (
    quote_id BIGINT PRIMARY KEY DEFAULT nextval('policy_issue_seq')
);
```

**Go Code**:
```go
type Quote struct {
    QuoteID int64 `db:"quote_id" json:"quote_id"`
}
```

**API Contract**:
- Path parameters: `type: integer, format: int64`
- JSON response: `quote_id: 12345` (not UUID string)

**Cross-Reference**: 
- Database Schema: `db/migrations/001_policy_issue_schema.sql` (all tables)
- Handler: `handler/quote.go` (QuoteConvertRequest.QuoteID as int64)
- Swagger: `nbf/swagger/policy_issue_swagger.yaml` (quote_id parameter)

---

## Naming Conventions

### Files
- `handler/{resource}.go` - lowercase, singular
- `repo/postgres/{resource}_repository.go` - lowercase, singular
- `core/domain/{resource}.go` - lowercase, singular
- `workflows/{resource}_workflow.go` - lowercase, singular
- `workflows/activities/{resource}_activities.go` - lowercase, plural

### Structs
- Domain: `PascalCase` (e.g., `Proposal`, `Quote`)
- Request DTO: `{Action}{Resource}Request` (e.g., `CreateQuoteRequest`)
- Response DTO: `{Resource}{Action}Response` (e.g., `QuoteCreateResponse`)
- Repository: `{Resource}Repository` (e.g., `ProposalRepository`)
- Handler: `{Resource}Handler` (e.g., `ProposalHandler`)

### Database
- Tables: `snake_case`, plural (e.g., `proposals`, `proposal_nominees`)
- Columns: `snake_case` (e.g., `proposal_id`, `created_at`)
- Enums: `{name}_enum` (e.g., `proposal_status_enum`)

### JSON Tags
- `snake_case` matching DB columns
- Example: `json:"proposal_number" db:"proposal_number"`

---

## Validation Strategy

### Auto-Generated Validators
- Use `govalid` tool: `govalid .` at project root
- Never manually edit `request_*_validator.go` files

### Validation Tags
```go
type QuoteCalculateRequest struct {
    ProductCode string `validate:"required,oneof=SURAKSHA SUVIDHA SANTOSH"`
    SumAssured  float64 `validate:"required,gt=0"`
    PolicyTerm  int `validate:"required,gte=5,lte=50"`
}
```

### Cross-Field Validation
- Use CEL expressions: `validate:"cel=value >= this.OtherField"`

---

## Error Code Mapping

### Validation Errors (400)
| Code | Message | Trigger |
|------|---------|---------|
| ERR-POL-001 | Invalid proposal ID | Path param validation |
| ERR-POL-006 | Invalid declaration date | Date chain validation |
| ERR-POL-010 | Invalid Aadhaar number | Format check |
| ERR-POL-022 | Nominee count exceeded | Max 3 nominees |

### Business Rule Errors (400)
| Code | Message | Trigger |
|------|---------|---------|
| ERR-POL-031 | Age not eligible | Product age limits |
| ERR-POL-033 | Sum assured outside range | Product SA limits |
| ERR-POL-048 | Not eligible for instant issuance | Aadhaar flow criteria |

### Integration Errors (502/503/504)
| Code | Message | Trigger |
|------|---------|---------|
| ERR-POL-061 | Aadhaar service unavailable | UIDAI timeout |
| ERR-POL-064 | Max retries exceeded | External API |

### System Errors (500)
| Code | Message | Trigger |
|------|---------|---------|
| ERR-POL-081 | Database error | DB connection |
| ERR-POL-082 | Temporal workflow error | Workflow execution |

---

## Open Questions / Pending Decisions

| ID | Question | Status | Owner |
|----|----------|--------|-------|
| Q1 | Should we cache product catalog in Redis? | OPEN | Architecture |
| Q2 | What's the SLA for bulk upload processing? | OPEN | Business |
| Q3 | Do we need read replicas for proposal queries? | OPEN | Architecture |
| Q4 | Should medical examination results be encrypted? | OPEN | Security |

---

## Resolved Ambiguities

| Date | Issue | Resolution | Reference |
|------|-------|------------|-----------|
| 2026-02-13 | Customer creation responsibility | Customer Service creates, Policy Issue uses customer_id | INT-POL-002 |
| 2026-02-13 | Quote expiration duration | 30 days default, configurable | FR-POL-001 |
| 2026-02-13 | Medical certificate validity | 60 days from examination | BR-POL-019 |
| 2026-02-13 | FLC period calculation | Based on dispatch_date or delivery_date (channel-based) | BR-POL-028 |

---

### 12. Policy Number Sequence Strategy

**Decision**: Use BIGSERIAL (auto-increment) sequences instead of `policy_number_sequence` table.

**Rationale**:
- BIGSERIAL is simpler and handles concurrent allocation automatically
- No need for complex table-based locking
- PostgreSQL guarantees uniqueness at the sequence level
- Better performance for high-concurrency issuance scenarios

**Implementation**:
```sql
CREATE SEQUENCE policy_number_seq START 1;
```

**Generation Logic**:
```go
func (r *ProposalRepository) GeneratePolicyNumber(ctx context.Context, prefix, stateCode string) (string, error) {
    year := time.Now().Year()
    var seq int64
    err := r.db.QueryRow(ctx, "SELECT nextval('policy_number_seq')").Scan(&seq)
    if err != nil {
        return "", err
    }
    return fmt.Sprintf("%s-%s-%d-%06d", prefix, stateCode, year, seq), nil
}
```

**Format**: `{prefix}-{state_code}-{year}-{sequence:06d}`  
**Example**: `SUR-DL-2025-000042`

**Cross-Reference**: See implementation in `repo/postgres/proposal_repository.go`

---

### 13. First Premium Recording Model

**Decision**: Use `proposal_indexing.first_premium_paid` flag instead of a separate `first_premium` table.

**Rationale**:
- Single-column flag is simpler than separate table
- Payment details are stored in Accounting Service (source of truth)
- Policy Issue service only needs to track whether payment was recorded
- Reduces schema complexity

**Implementation**:
```sql
ALTER TABLE proposal_indexing ADD COLUMN first_premium_paid BOOLEAN NOT NULL DEFAULT FALSE;
```

**Workflow**:
1. Handler calls Accounting Service to record first premium
2. On success, handler sets `first_premium_paid = TRUE`
3. QC submission checks this flag before allowing submission

**Cross-Reference**: 
- Handler: `handler/proposal.go` - `RecordFirstPremium` endpoint
- Database: `db/migrations/001_policy_issue_schema.sql` (proposal_indexing table)

---

### 14. Workflow State Caching Requirement

**Decision**: Fetch customer data ONCE at workflow start via initial activity.

**Rationale**:
- Avoid redundant DB queries across activities
- Customer data (age, gender, state) needed for multiple decisions
- Temporal workflow state persists data between activities
- Reduces external service calls

**Pattern**:
```go
type WorkflowState struct {
    ProposalID     string
    CustomerID     string
    CustomerAge    int      // Fetched once
    CustomerGender string   // Fetched once  
    InsuredState   string   // Fetched once
    // ... other cached data
}

func (w *PolicyIssuanceWorkflow) FetchInitialData(ctx context.Context) error {
    // Single call to Customer Service
    customerData := activities.FetchCustomerData(ctx, w.state.CustomerID)
    w.state.CustomerAge = customerData.Age
    w.state.CustomerGender = customerData.Gender
    w.state.InsuredState = customerData.State
    return nil
}
```

**Activities Using Cached State**:
- Eligibility Check: Uses `CustomerAge`
- Medical Underwriting: Uses `CustomerAge`, `CustomerGender`
- Policy Number Generation: Uses `InsuredState`

**Cross-Reference**: 
- Workflow: `workflows/policy_issuance_workflow.go`
- Activities: `workflows/activities/proposal_activities.go`
- Handler: TODO comment in `handler/proposal.go` SubmitForQC

---

## Performance Considerations

### Database
- Use partial indexes for status-based queries
- Partition `proposal_status_history` by date if > 10M records
- Use `pg_trgm` extension for fuzzy search

### Temporal
- Workflow state size limit: 10MB (monitor and alert)
- Signal channel buffer: 100 signals
- Activity retry: Max 3 for internal, Max 5 for external

### API
- Proposal queue pagination: Default 20, Max 100
- Quote validity: 30 days with cleanup job

---

## Security Considerations

### Data Classification
| Data | Classification | Storage |
|------|----------------|---------|
| Aadhaar Number | PII | Encrypted at rest, masked in logs |
| PAN Number | PII | Encrypted at rest |
| Bank Account | Sensitive | Encrypted at rest |
| Medical History | PHI | Encrypted at rest, audit all access |

### API Security
- All endpoints require authentication (JWT)
- Role-based access control (RBAC)
- Rate limiting: 100 req/min per user

---

## Monitoring & Observability

### Metrics
- Workflow execution duration
- Activity failure rate
- State transition counts
- Queue depth by status

### Alerts
- Workflow stuck > 7 days
- DB connection pool exhausted
- External service errors > 5%

### Logging
- All state transitions logged
- External API calls logged with correlation ID
- Audit log for all approval actions
