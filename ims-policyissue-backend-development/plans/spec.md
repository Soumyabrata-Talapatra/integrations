# Policy Issue Microservice - API Specifications

## Overview

This document contains detailed specifications for the Policy Issue microservice APIs. Each API is specified with:
- Request/Response schemas
- Business rules and validations
- Error codes
- Temporal workflow integration (where applicable)

---

## Quote Module APIs

### POL-API-001: Get Products
**Endpoint**: `GET /products`

**Description**: Retrieves list of available PLI/RPLI products for quote generation.

**Query Parameters**:
| Name | Type | Required | Description |
|------|------|----------|-------------|
| policy_type | string | No | Filter by PLI/RPLI |
| is_active | boolean | No | Default: true |

**Response**: `ProductListResponse`
```json
{
  "products": [
    {
      "product_code": "SURAKSHA",
      "product_name": "Whole Life Assurance",
      "product_type": "PLI",
      "min_sum_assured": 20000,
      "max_sum_assured": null,
      "min_entry_age": 19,
      "max_entry_age": 55,
      "available_frequencies": ["MONTHLY", "QUARTERLY", "HALF_YEARLY", "YEARLY"]
    }
  ]
}
```

**Business Rules**:
- BR-POL-011: PLI product eligibility
- BR-POL-012: RPLI product eligibility

**Repository Pattern**:
```go
func (r *ProductRepository) GetProducts(ctx context.Context, policyType string, isActive bool) ([]domain.Product, error)
```

---

### POL-API-002: Calculate Quote
**Endpoint**: `POST /quotes/calculate`

**Description**: Calculates premium quote based on product, age, sum assured, and term.

**Request**: `QuoteCalculateRequest`
```go
type QuoteCalculateRequest struct {
    PolicyType  string  `json:"policy_type" validate:"required,oneof=PLI RPLI"`
    ProductCode string  `json:"product_code" validate:"required"`
    DateOfBirth string  `json:"date_of_birth" validate:"required,date"`
    Gender      string  `json:"gender" validate:"required,oneof=MALE FEMALE OTHER"`
    SumAssured  float64 `json:"sum_assured" validate:"required,gt=0"`
    PolicyTerm  int     `json:"policy_term" validate:"required,gte=5,lte=50"`
    Frequency   string  `json:"frequency" validate:"required,oneof=MONTHLY QUARTERLY HALF_YEARLY YEARLY"`
    StateCode   string  `json:"state_code"` // For GST calculation
}
```

**Response**: `QuoteCalculateResponse`
```json
{
  "calculation_id": "550e8400-e29b-41d4-a716-446655440000",
  "eligibility": {
    "is_eligible": true,
    "age_at_entry": 35,
    "maturity_age": 70
  },
  "premium_breakdown": {
    "base_premium": 15250.00,
    "rebate": 250.00,
    "net_premium": 15000.00,
    "cgst": 1350.00,
    "sgst": 1350.00,
    "total_gst": 2700.00,
    "total_payable": 17700.00
  },
  "benefit_illustration": {
    "maturity_value_guaranteed": 500000.00,
    "maturity_value_with_bonus": 725000.00,
    "indicative_bonus_rate": 4.5,
    "death_benefit": 500000.00
  }
}
```

**Business Rules**:
- BR-POL-001: Base Premium = (SA / 1000) × Rate from Sankalan table
- BR-POL-002: GST = 18% of net premium (CGST 9% + SGST 9% or IGST 18%)
- BR-POL-003: Rebate based on frequency (Yearly: 2%, Half-Yearly: 1%)
- VR-PI-012: Age must be within product entry age limits
- VR-PI-013: Sum assured must be within product limits
- VR-PI-044: Policy term must be within product limits

**Error Codes**:
- ERR-POL-031: Age not eligible for product
- ERR-POL-033: Sum assured outside range
- ERR-POL-036: Invalid policy term

**Repository Pattern**:
```go
// Get premium rate from Sankalan table
func (r *QuoteRepository) GetPremiumRate(ctx context.Context, age int, gender string, productCode string, term int) (float64, error)

// Get product configuration
func (r *QuoteRepository) GetProductConfig(ctx context.Context, productCode string) (*domain.Product, error)
```

---

### POL-API-003: Create Quote
**Endpoint**: `POST /quotes`

**Description**: Saves a calculated quote and generates quote reference number.

**Request**: `QuoteCreateRequest`
```go
type QuoteCreateRequest struct {
    CalculationID string                `json:"calculation_id" validate:"required,uuid"`
    PolicyType    string                `json:"policy_type" validate:"required"`
    ProductCode   string                `json:"product_code" validate:"required"`
    Proposer      ProposerInfo          `json:"proposer" validate:"required"`
    Coverage      CoverageInfo          `json:"coverage" validate:"required"`
    Channel       string                `json:"channel" validate:"required"`
    CreatedBy     int64                 `json:"created_by" validate:"required"`
    GeneratePDF   bool                  `json:"generate_pdf"`
}
```

**Response**: `QuoteCreateResponse`
```json
{
  "quote_id": "550e8400-e29b-41d4-a716-446655440000",
  "quote_ref_number": "QT-PLI-2026-00012345",
  "status": "GENERATED",
  "premium_breakdown": { /* ... */ },
  "validity": {
    "expires_at": "2026-03-15T10:30:00Z",
    "days_valid": 30
  }
}
```

**Business Logic**:
1. Validate calculation_id exists and is not expired
2. Generate unique quote_ref_number: `QT-{POLICY_TYPE}-{YEAR}-{SEQUENCE:08d}`
3. Save quote to database
4. Generate PDF if requested (async to DMS)
5. Return quote reference

**Repository Pattern**:
```go
func (r *QuoteRepository) CreateQuote(ctx context.Context, quote *domain.Quote) (*domain.Quote, error)
func (r *QuoteRepository) GetQuoteByCalculationID(ctx context.Context, calcID string) (*domain.Quote, error)
```

---

### POL-API-004: Convert Quote to Proposal
**Endpoint**: `POST /quotes/{quote_id}/convert-to-proposal`

**Description**: Converts a saved quote to a draft proposal.

**Path Parameters**:
| Name | Type | Description |
|------|------|-------------|
| quote_id | string | Quote UUID |

**Response**: `QuoteConvertResponse`
```json
{
  "proposal_id": 12345,
  "proposal_number": "PLI-MH-2026-00056789",
  "quote_id": "550e8400-e29b-41d4-a716-446655440000",
  "status": "DATA_ENTRY",
  "redirect_url": "/proposals/12345/insured-details"
}
```

**Business Rules**:
- BR-POL-024: Deduplication check before conversion
- Quote must be in GENERATED or SAVED status
- Quote must not be expired

**Temporal Workflow**: Starts WF-PI-001

**Error Codes**:
- ERR-POL-055: Quote has expired
- ERR-POL-003: Duplicate customer detected

---

## Proposal Core Module APIs

### POL-API-005: Create Proposal Indexing
**Endpoint**: `POST /proposals/indexing`

**Description**: Creates a new proposal through CPC indexing (without Aadhaar).

**Request**: `ProposalIndexingRequest`
```go
type ProposalIndexingRequest struct {
    PolicyType       string    `json:"policy_type" validate:"required,oneof=PLI RPLI"`
    ProductCode      string    `json:"product_code" validate:"required"`
    EntryPath        string    `json:"entry_path" validate:"required"`
    POCode           string    `json:"po_code" validate:"required"`
    Channel          string    `json:"channel" validate:"required"`
    CustomerID       int64     `json:"customer_id" validate:"required"`
    SpouseCustomerID *int64    `json:"spouse_customer_id"`
    Dates            DateInfo  `json:"dates" validate:"required"`
    QuoteRefNumber   string    `json:"quote_ref_number" validate:"required"`
}

type DateInfo struct {
    DeclarationDate string `json:"declaration_date" validate:"required,date"`
    ReceiptDate     string `json:"receipt_date" validate:"required,date"`
    IndexingDate    string `json:"indexing_date" validate:"required,date"`
    ProposalDate    string `json:"proposal_date" validate:"required,date"`
}
```

**Response**: `ProposalIndexingResponse`
```json
{
  "proposal_id": 12345,
  "proposal_number": "PLI-MH-2026-00056789",
  "customer_id": 67890,
  "status": "DATA_ENTRY",
  "workflow_id": "policy-issuance-12345",
  "acknowledgement": {
    "receipt_number": "ACK-PLI-2026-00056789",
    "receipt_date": "2026-02-13",
    "download_url": "/api/v1/documents/ack-12345.pdf"
  },
  "workflow_state": {
    "current_step": "DATA_ENTRY",
    "next_step": "INSURED_DETAILS"
  }
}
```

**Business Rules**:
- BR-POL-015: State machine initialization (DRAFT → DATA_ENTRY)
- BR-POL-018: Date chain validation (declaration ≤ receipt ≤ indexing ≤ proposal)
- BR-POL-024: Deduplication check

**Validation Rules**:
- VR-PI-004 to VR-PI-007: Date sequence validation

**Error Codes**:
- ERR-POL-006: Invalid declaration date
- ERR-POL-007: Invalid receipt date
- ERR-POL-008: Invalid indexing date
- ERR-POL-009: Invalid proposal date

**Repository Pattern (pgx.Batch)**:
```go
func (r *ProposalRepository) CreateProposalIndexing(ctx context.Context, req *domain.ProposalIndexing) (*domain.Proposal, error) {
    batch := &pgx.Batch{}
    
    // 1. Insert into proposals
    // 2. Insert into proposal_indexing
    // 3. Insert into proposal_status_history
    // 4. Insert into proposal_data_entry
    
    // Execute batch
    err := r.db.SendBatch(ctx, batch).Close()
    // ...
}
```

**Temporal Workflow**: Starts WF-PI-001

---

### POL-API-008: Update Insured Details
**Endpoint**: `PUT /proposals/{proposal_id}/sections/insured`

**Description**: Updates insured person details section of proposal.

**Request**: `InsuredDetailsRequest`
```go
type InsuredDetailsRequest struct {
    Salutation                  string              `json:"salutation" validate:"required"`
    FirstName                   string              `json:"first_name" validate:"required,max=100"`
    MiddleName                  string              `json:"middle_name" max="100"`
    LastName                    string              `json:"last_name" validate:"required,max=100"`
    DateOfBirth                 string              `json:"date_of_birth" validate:"required,date"`
    Gender                      string              `json:"gender" validate:"required"`
    MaritalStatus               string              `json:"marital_status"`
    FatherName                  string              `json:"father_name"`
    HusbandName                 string              `json:"husband_name"`
    Mobile                      string              `json:"mobile" validate:"required,pattern=^[0-9]{10}$"`
    Email                       string              `json:"email" validate:"email"`
    AadhaarNumber               string              `json:"aadhaar_number" validate:"pattern=^[0-9]{12}$"`
    PANNumber                   string              `json:"pan_number" validate:"pattern=^[A-Z]{5}[0-9]{4}[A-Z]{1}$"`
    CommunicationAddress        *Address            `json:"communication_address" validate:"required"`
    PermanentAddress            *Address            `json:"permanent_address"`
    IsPermanentSameAsComm       bool                `json:"is_permanent_same_as_communication"`
    Occupation                  string              `json:"occupation"`
    Employment                  *EmploymentInfo     `json:"employment"`
}
```

**Business Rules**:
- VAL-POL-016: Husband name for married female
- VAL-POL-017: Father name for male/unmarried female
- VAL-POL-018: Communication address required
- VR-PI-010: Aadhaar format validation
- VR-PI-011: PAN format validation

**Repository Pattern**:
```go
func (r *ProposalRepository) UpdateInsuredDetails(ctx context.Context, proposalID int64, details *domain.InsuredDetails) error
```

---

### POL-API-009: Update Nominees
**Endpoint**: `PUT /proposals/{proposal_id}/sections/nominees`

**Description**: Updates nominee section of proposal. Maximum 3 nominees allowed.

**Request**: `NomineesRequest`
```go
type NomineesRequest struct {
    Nominees []Nominee `json:"nominees" validate:"required,minitems=1,maxitems=3"`
}

type Nominee struct {
    Salutation      string   `json:"salutation"`
    FirstName       string   `json:"first_name" validate:"required"`
    LastName        string   `json:"last_name" validate:"required"`
    DateOfBirth     string   `json:"date_of_birth" validate:"required,date"`
    Gender          string   `json:"gender" validate:"required"`
    Relationship    string   `json:"relationship" validate:"required"`
    SharePercentage float64  `json:"share_percentage" validate:"required,gte=1,lte=100"`
    IsMinor         bool     `json:"is_minor"`
    Appointee       *Appointee `json:"appointee"`
}
```

**Business Rules**:
- VAL-POL-003: Nominee shares must total 100%
- VAL-POL-004: Maximum 3 nominees
- VAL-POL-005: At least 1 nominee (except HUF/MWPA)
- VAL-POL-006: Appointee required for minor nominee

**Error Codes**:
- ERR-POL-022: Nominee count exceeded
- ERR-POL-023: Nominee share invalid (not 100% total)
- ERR-POL-024: Appointee required for minor

**Validation Logic**:
```go
func (r *ProposalRepository) ValidateNominees(nominees []domain.Nominee) error {
    if len(nominees) > 3 {
        return errors.New("ERR-POL-022: Maximum 3 nominees allowed")
    }
    
    totalShare := 0.0
    for _, n := range nominees {
        totalShare += n.SharePercentage
        if n.IsMinor && n.Appointee == nil {
            return errors.New("ERR-POL-024: Appointee required for minor nominee")
        }
    }
    
    if totalShare != 100.0 {
        return errors.New("ERR-POL-023: Nominee shares must total 100%")
    }
    
    return nil
}
```

---

### POL-API-013: Submit for QC
**Endpoint**: `POST /proposals/{proposal_id}/submit-for-qc`

**Description**: Submits completed proposal for quality review.

**Business Rules**:
- BR-POL-015: State transition DATA_ENTRY → QC_PENDING
- All mandatory sections must be complete
- All mandatory documents must be uploaded

**Temporal Workflow**: Sends signal to WF-PI-001

**Signal**:
```go
type SubmitForQCSignal struct {
    ProposalID string
    SubmittedBy string
    Comments string
}
```

**Error Codes**:
- ERR-POL-029: Mandatory document missing
- ERR-POL-050: Incomplete proposal

---

## Approval Workflow APIs

### POL-API-017: QR Approve Proposal
**Endpoint**: `POST /proposals/{proposal_id}/qr-approve`

**Description**: Quality Reviewer approves proposal for next stage.

**Request**: `QRDecisionRequest`
```go
type QRDecisionRequest struct {
    ReviewerID string `json:"reviewer_id" validate:"required"`
    Comments   string `json:"comments"`
}
```

**Business Rules**:
- BR-POL-015: State transition QC_PENDING → QC_APPROVED
- BR-POL-016: Routes to appropriate approver level based on SA

**State Transitions**:
- If medical required: QC_APPROVED → PENDING_MEDICAL
- If no medical: QC_APPROVED → APPROVAL_PENDING

**Temporal Workflow**: Sends SignalQRDecision with APPROVED

---

### POL-API-018: QR Reject Proposal
**Endpoint**: `POST /proposals/{proposal_id}/qr-reject`

**Request**: `QRRejectRequest`
```go
type QRRejectRequest struct {
    ReviewerID string `json:"reviewer_id" validate:"required"`
    Comments   string `json:"comments" validate:"required"` // BR-POL-017: Rejection requires comments
    ReasonCode string `json:"reason_code" validate:"required"`
}
```

**Business Rules**:
- BR-POL-015: State transition QC_PENDING → QC_REJECTED
- BR-POL-017: Rejection requires comments (mandatory)

**Error Codes**:
- ERR-POL-054: Rejection details required

---

### POL-API-020: Approver Approve Proposal
**Endpoint**: `POST /proposals/{proposal_id}/approve`

**Request**: `ApproverDecisionRequest`
```go
type ApproverDecisionRequest struct {
    ApproverID   string `json:"approver_id" validate:"required"`
    ApproverLevel int   `json:"approver_level" validate:"required,oneof=1 2 3"`
    Comments     string `json:"comments"`
}
```

**Business Rules**:
- BR-POL-015: State transition APPROVAL_PENDING → APPROVED
- BR-POL-016: Approval level must match SA range
- BR-POL-025: Age revalidation at approval time

**Approval Hierarchy**:
- Level 1: SA ≤ ₹5,00,000
- Level 2: SA ≤ ₹10,00,000
- Level 3: SA > ₹10,00,000

**Actions Triggered**:
1. Generate policy number
2. Generate policy bond
3. Trigger commission calculation
4. Send notification to customer

**Error Codes**:
- ERR-POL-049: Incorrect approver level
- ERR-POL-050: Quality review pending

---

## Policy Lifecycle APIs

### POL-API-022: Cancel Policy FLC
**Endpoint**: `POST /policies/{policy_id}/flc-cancel`

**Description**: Processes free look cancellation within FLC period.

**Request**: `FLCCancelRequest`
```go
type FLCCancelRequest struct {
    Reason      string `json:"reason" validate:"required"`
    RequestedBy string `json:"requested_by" validate:"required"`
    RequestDate string `json:"request_date" validate:"required,date"`
}
```

**Response**: `FLCCancelResponse`
```json
{
  "policy_id": "POL-PLI-2026-00056789",
  "cancellation_status": "FLC_CANCELLED",
  "refund_details": {
    "premium_paid": 17700.00,
    "proportionate_risk": 145.00,
    "stamp_duty_deducted": 50.00,
    "medical_fee_deducted": 0.00,
    "refund_amount": 17505.00
  },
  "flc_period": {
    "start_date": "2026-02-15",
    "end_date": "2026-03-02",
    "days_remaining": 5
  }
}
```

**Business Rules**:
- BR-POL-009: FLC Refund Calculation
  - Deduct proportionate risk premium (days of coverage)
  - Deduct stamp duty
  - Deduct medical fee (if applicable)
- BR-POL-021: Free Look Period Duration (15-30 days based on channel)
- BR-POL-028: FLC Start Date Determination (dispatch, delivery, or email date)

**Refund Formula**:
```
Refund = Premium Paid - Proportionate Risk - Stamp Duty - Medical Fee
```

**Temporal Workflow**: Sends SignalFLCCancelRequest

**Error Codes**:
- ERR-POL-053: Free look period expired

---

## Aadhaar Flow APIs

### POL-API-013: Initiate Aadhaar Auth
**Endpoint**: `POST /proposals/aadhaar/initiate`

**Request**: `AadhaarInitiateRequest`
```go
type AadhaarInitiateRequest struct {
    AadhaarNumber string `json:"aadhaar_number" validate:"required,pattern=^[0-9]{12}$"`
    Purpose       string `json:"purpose" validate:"required"`
}
```

**Response**: `AadhaarInitiateResponse`
```json
{
  "transaction_id": "TXN-AAD-20260213-001",
  "uidai_reference": "UID-REF-12345",
  "status": "OTP_SENT",
  "validity_seconds": 600
}
```

**Integration**: INT-POL-003, INT-POL-005 (UIDAI OTP service)

**Error Codes**:
- ERR-POL-010: Invalid Aadhaar number
- ERR-POL-061: Aadhaar service unavailable

---

### POL-API-015: Submit Aadhaar Proposal
**Endpoint**: `POST /proposals/aadhaar/submit`

**Description**: Submits complete proposal with Aadhaar-authenticated data. Supports instant issuance for non-medical cases.

**Instant Issuance Criteria**:
- Age ≤ 50 years
- SA within non-medical limit (< 20 Lakh)
- No adverse medical history
- Payment completed

**Temporal Workflows**:
- WF-PI-002 (InstantIssuanceWorkflow) if eligible
- WF-PI-001 (PolicyIssuanceWorkflow) otherwise

**Error Codes**:
- ERR-POL-048: Not eligible for instant issuance
- ERR-POL-037: Medical examination required

---

## Workflow Management APIs

### WF-POL-001: Query Workflow Status
**Endpoint**: `GET /workflows/{workflow_id}/status`

**Response**:
```json
{
  "workflow_id": "policy-issuance-12345",
  "proposal_id": 12345,
  "current_status": "QC_PENDING",
  "started_at": "2026-02-13T10:30:00Z",
  "last_activity": "2026-02-13T14:45:00Z",
  "history": [
    {"status": "VALIDATING", "timestamp": "2026-02-13T10:30:00Z"},
    {"status": "CHECKING_ELIGIBILITY", "timestamp": "2026-02-13T10:30:05Z"},
    {"status": "CALCULATING_PREMIUM", "timestamp": "2026-02-13T10:30:10Z"},
    {"status": "QC_PENDING", "timestamp": "2026-02-13T10:35:00Z"}
  ]
}
```

**Temporal Query**: Uses QueryProposalStatus query handler

---

## Database Query Patterns

### Pattern 1: Get Proposal Detail (pgx.Batch)
```go
func (r *ProposalRepository) GetProposalDetail(ctx context.Context, proposalID int64) (*domain.ProposalDetail, error) {
    batch := &pgx.Batch{}
    
    var proposal domain.Proposal
    var indexing domain.ProposalIndexing
    var nominees []domain.Nominee
    var medical domain.MedicalInfo
    
    // Queue all queries
    q1 := dblib.Psql.Select("*").From("proposals").Where(sq.Eq{"proposal_id": proposalID})
    dblib.QueueReturnRow(batch, q1, pgx.RowToStructByNameLax[domain.Proposal], &proposal)
    
    q2 := dblib.Psql.Select("*").From("proposal_indexing").Where(sq.Eq{"proposal_id": proposalID})
    dblib.QueueReturnRow(batch, q2, pgx.RowToStructByNameLax[domain.ProposalIndexing], &indexing)
    
    q3 := dblib.Psql.Select("*").From("proposal_nominee").Where(sq.Eq{"proposal_id": proposalID})
    dblib.QueueReturn(batch, q3, pgx.RowToStructByNameLax[domain.Nominee])
    
    q4 := dblib.Psql.Select("*").From("proposal_medical_info").Where(sq.Eq{"proposal_id": proposalID})
    dblib.QueueReturnRow(batch, q4, pgx.RowToStructByNameLax[domain.MedicalInfo], &medical)
    
    // Execute in one round trip
    err := r.db.SendBatch(ctx, batch).Close()
    if err != nil {
        return nil, err
    }
    
    return &domain.ProposalDetail{
        Proposal: proposal,
        Indexing: indexing,
        Nominees: nominees,
        Medical:  medical,
    }, nil
}
```

### Pattern 2: Proposal Queue Query
```go
func (r *ProposalRepository) GetProposalQueue(ctx context.Context, status string, level int, page, limit int) ([]domain.ProposalTicket, error) {
    // Use view v_proposal_ticket_queue
    query := dblib.Psql.Select("*").From("v_proposal_ticket_queue")
    
    if status != "" {
        query = query.Where(sq.Eq{"status": status})
    }
    
    if level > 0 {
        query = query.Where(sq.Eq{"approval_level": level})
    }
    
    query = query.OrderBy("created_at DESC").
        Limit(uint64(limit)).
        Offset(uint64((page - 1) * limit))
    
    return dblib.SelectRows[domain.ProposalTicket](ctx, r.db, query)
}
```

---

## Traceability Matrix

| API ID | Business Rules | Validation Rules | Error Codes | Workflow |
|--------|---------------|------------------|-------------|----------|
| POL-API-001 | BR-POL-011, BR-POL-012 | - | - | - |
| POL-API-002 | BR-POL-001, BR-POL-002, BR-POL-003 | VR-PI-012, VR-PI-013, VR-PI-044 | ERR-POL-031, ERR-POL-033, ERR-POL-036 | - |
| POL-API-004 | BR-POL-024 | - | ERR-POL-055, ERR-POL-003 | WF-PI-001 |
| POL-API-005 | BR-POL-015, BR-POL-018, BR-POL-024 | VR-PI-004 to VR-PI-007 | ERR-POL-006 to ERR-POL-009 | WF-PI-001 |
| POL-API-008 | - | VAL-POL-016 to VAL-POL-019, VR-PI-010, VR-PI-011 | ERR-POL-010, ERR-POL-011 | - |
| POL-API-009 | - | VAL-POL-003 to VAL-POL-006 | ERR-POL-022, ERR-POL-023, ERR-POL-024 | - |
| POL-API-013 | BR-POL-015 | - | ERR-POL-029, ERR-POL-050 | Signal |
| POL-API-017 | BR-POL-015, BR-POL-016 | - | - | Signal |
| POL-API-020 | BR-POL-015, BR-POL-016, BR-POL-025 | - | ERR-POL-049, ERR-POL-050 | Signal |
| POL-API-022 | BR-POL-009, BR-POL-021, BR-POL-028 | VR-PI-024 | ERR-POL-053 | Signal |
