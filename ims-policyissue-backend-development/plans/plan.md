# Policy Issue Microservice - Implementation Plan

## Project Overview

**Service**: Policy Issue Microservice  
**Project**: India Post PLI/RPLI Insurance Management System (IMS)  
**Framework**: N-API Template (Go, Uber FX, Temporal, PostgreSQL)  
**Total APIs**: 62 across 12 categories  

---

## API Categories Summary

| Category | Count | Temporal Workflows | Priority |
|----------|-------|-------------------|----------|
| Quote APIs | 4 | No | HIGH |
| Proposal Core APIs | 12 | Yes (WF-PI-001) | CRITICAL |
| Aadhaar Flow APIs | 3 | Yes (WF-PI-002) | CRITICAL |
| Approval Workflow APIs | 5 | Yes (Signals) | CRITICAL |
| Policy Lifecycle APIs | 3 | Yes (FLC) | HIGH |
| Lookup APIs | 10 | No | MEDIUM |
| Validation APIs | 8 | No | MEDIUM |
| Calculation APIs | 4 | No | HIGH |
| Document APIs | 4 | No | MEDIUM |
| Status & Tracking APIs | 4 | No | MEDIUM |
| Workflow Management APIs | 3 | Yes (Queries) | MEDIUM |
| Bulk Upload APIs | 2 | Yes (WF-PI-003) | MEDIUM |

---

## Implementation Phases

### Phase 0: Foundation & Setup
- [ ] Create project structure following n-api-template
- [ ] Setup go.mod with all dependencies
- [ ] Create bootstrap/bootstrapper.go with FX modules
- [ ] Create configs/config.yaml and environment variants
- [ ] Create core/port/request.go and response.go
- [ ] Create db/migrations from policy_issue_schema.sql

### Phase 1: Quote Module (No Temporal)
**APIs**: POL-API-001 to POL-API-004

- [ ] Core Domain: core/domain/quote.go, core/domain/product.go
- [ ] Handler: handler/quote.go
- [ ] Request DTOs: handler/request.go (QuoteCalculateRequest, QuoteCreateRequest)
- [ ] Response DTOs: handler/response/quote.go
- [ ] Repository: repo/postgres/quote_repository.go
- [ ] Business Logic: Premium calculation (BR-POL-001 to BR-POL-003)
- [ ] Validation: handler/request_quote_validator.go

**Key Business Rules**:
- BR-POL-001: Base Premium Calculation (Sankalan table lookup)
- BR-POL-002: GST Calculation (CGST/SGST or IGST)
- BR-POL-003: Rebate Calculation
- BR-POL-010: Total Premium Pipeline

### Phase 2: Proposal Core Module (With Temporal)
**APIs**: POL-API-005 to POL-API-017

- [ ] Core Domain: core/domain/proposal.go with all phase structs
- [ ] Handler: handler/proposal.go
- [ ] Request DTOs: handler/request.go (ProposalIndexingRequest, InsuredDetailsRequest, NomineesRequest, etc.)
- [ ] Response DTOs: handler/response/proposal.go
- [ ] Repository: repo/postgres/proposal_repository.go (with pgx.Batch)
- [ ] Workflow: workflows/policy_issuance_workflow.go
- [ ] Activities: workflows/activities/proposal_activities.go
- [ ] Workflow State: workflows/state/proposal_workflow_state.go

**Key Endpoints**:
- POST /proposals/indexing (POL-API-005)
- GET /proposals/{proposal_id} (POL-API-006)
- GET /proposals/resolve/{proposal_number} (POL-API-017)
- POST /proposals/{proposal_id}/first-premium (POL-API-007)
- PUT /proposals/{proposal_id}/sections/insured (POL-API-008)
- PUT /proposals/{proposal_id}/sections/nominees (POL-API-009)
- PUT /proposals/{proposal_id}/sections/policy-details (POL-API-010)
- PUT /proposals/{proposal_id}/sections/agent (POL-API-011)
- PUT /proposals/{proposal_id}/sections/medical (POL-API-012)
- POST /proposals/{proposal_id}/submit-for-qc (POL-API-013)
- GET /proposals/{proposal_id}/summary (POL-API-016)
- GET /proposals/queue (WF-POL-003)

**Key Business Rules**:
- BR-POL-015: Proposal State Machine
- BR-POL-016: Approval Routing by SA
- BR-POL-018: Date Validation Chain

### Phase 3: Aadhaar Flow Module (Instant Issuance)
**APIs**: POL-API-013 to POL-API-015

- [ ] Handler: handler/aadhaar.go
- [ ] Workflow: workflows/instant_issuance_workflow.go
- [ ] Activities: workflows/activities/aadhaar_activities.go

**Key Endpoints**:
- POST /proposals/aadhaar/initiate (POL-API-013)
- POST /proposals/aadhaar/verify-otp (POL-API-014)
- POST /proposals/aadhaar/submit (POL-API-015)

### Phase 4: Approval Workflow Module
**APIs**: POL-API-017 to POL-API-021

- [ ] Handler: handler/approval.go
- [ ] Workflow Integration: Signal handlers in policy_issuance_workflow.go

**Key Endpoints**:
- POST /proposals/{proposal_id}/qr-approve (POL-API-017)
- POST /proposals/{proposal_id}/qr-reject (POL-API-018)
- POST /proposals/{proposal_id}/qr-return (POL-API-019)
- POST /proposals/{proposal_id}/approve (POL-API-020)
- POST /proposals/{proposal_id}/reject (POL-API-021)

**Key Business Rules**:
- BR-POL-017: Rejection Requires Comments
- BR-POL-025: Age Revalidation at Approval

### Phase 5: Policy Lifecycle Module
**APIs**: POL-API-022 to POL-API-024

- [ ] Handler: handler/policy.go
- [ ] Workflow: FLC cancellation workflow

**Key Endpoints**:
- GET /policies/{policy_id} (POL-API-023)
- POST /policies/{policy_id}/flc-cancel (POL-API-022)
- GET /policies/{policy_id}/flc-status (POL-API-024)

**Key Business Rules**:
- BR-POL-009: FLC Refund Calculation
- BR-POL-021: Free Look Period Duration
- BR-POL-028: FLC Period Start Date Determination

### Phase 6: Lookup & Validation APIs
**APIs**: 18 endpoints (Lookup: 10, Validation: 8)

- [ ] Handler: handler/lookup.go
- [ ] Handler: handler/validation.go

**Key Endpoints**:
- GET /lookup/products, /lookup/agents, /lookup/occupations, etc.
- POST /validate/aadhaar, /validate/pan, /validate/bank-account, etc.

### Phase 7: Calculation APIs
**APIs**: 4 endpoints

- [ ] Handler: handler/calculation.go

**Key Endpoints**:
- POST /calculate/premium
- POST /calculate/gst
- POST /calculate/flc-refund
- POST /calculate/rebate

### Phase 8: Document & Status APIs
**APIs**: 8 endpoints (Document: 4, Status: 4)

- [ ] Handler: handler/document.go
- [ ] Handler: handler/status.go

### Phase 9: Workflow Management & Bulk Upload
**APIs**: 5 endpoints (Workflow: 3, Bulk: 2)

- [ ] Handler: handler/workflow.go
- [ ] Handler: handler/bulk_upload.go
- [ ] Workflow: workflows/bulk_proposal_upload_workflow.go

### Phase 10: Integration & Testing
- [ ] Integration tests for all workflows
- [ ] End-to-end user journey testing
- [ ] Performance testing for pgx.Batch operations
- [ ] Temporal workflow reliability testing

---

## Database Schema Implementation

### Core Tables
1. proposals (core entity)
2. proposal_indexing (phase table)
3. proposal_data_entry (phase table)
4. proposal_qc_review (phase table)
5. proposal_medical (phase table)
6. proposal_approval (phase table)
7. proposal_issuance (phase table)

### Child Tables
8. proposal_nominee
9. proposal_medical_info
10. proposal_enhanced_medical
11. proposal_agent
12. proposal_mwpa_trustee
13. proposal_huf_member
14. proposal_existing_policy
15. proposal_status_history
16. proposal_document_ref
17. proposal_missing_documents
18. proposal_audit_log

### Reference Tables
19. product_catalog
20. quote
21. policy_number_sequence
22. free_look_config
23. approval_routing_config
24. bulk_upload_batch

---

## Temporal Workflows

### WF-PI-001: PolicyIssuanceWorkflow
**Duration**: Days to weeks (QC, Medical, Approval stages)
**Signals**:
- SignalQRDecision (APPROVED, REJECTED, RETURNED)
- SignalCPCResubmit
- SignalMedicalResult
- SignalApproverDecision
- SignalPaymentReceived

**Activities**:
1. ValidateProposalActivity
2. CheckEligibilityActivity
3. CalculatePremiumActivity
4. SavePremiumToProposalActivity
5. UpdateProposalStatusActivity
6. SendNotificationActivity
7. RequestMedicalReviewActivity
8. RouteToApproverActivity
9. GeneratePolicyNumberActivity
10. GenerateBondActivity
11. DispatchKitActivity
12. TriggerCommissionActivity

### WF-PI-002: InstantIssuanceWorkflow
**Duration**: Minutes
**Criteria**: Aadhaar + Non-medical + SA < 20L + Age ≤ 50

### WF-PI-003: BulkProposalUploadWorkflow
**Duration**: Hours
**Features**: Batch processing with retry logic

---

## Integration Points

| ID | System | Type | Endpoints Affected |
|----|--------|------|-------------------|
| INT-POL-001 | Product Catalog | Internal | Quote, Proposal |
| INT-POL-002 | Customer Service | Internal | All proposal creation |
| INT-POL-003 | KYC Service (UIDAI) | External | Aadhaar Flow |
| INT-POL-004 | NSDL PAN Service | External | Insured Details |
| INT-POL-008 | DMS | Internal | Documents |
| INT-POL-010 | Medical Appointment | Internal | Medical Underwriting |
| INT-POL-011 | Collections | Internal | Premium Payment |
| INT-POL-013 | Notification | Internal | All status changes |

---

## File Structure

```
policy-issue-service/
├── main.go
├── go.mod
├── go.sum
├── configs/
│   ├── config.yaml
│   ├── config.dev.yaml
│   ├── config.sit.yaml
│   ├── config.staging.yaml
│   └── config.prod.yaml
├── bootstrap/
│   └── bootstrapper.go
├── core/
│   ├── domain/
│   │   ├── quote.go
│   │   ├── product.go
│   │   ├── proposal.go
│   │   ├── proposal_indexing.go
│   │   ├── proposal_data_entry.go
│   │   ├── proposal_qc_review.go
│   │   ├── proposal_medical.go
│   │   ├── proposal_approval.go
│   │   ├── proposal_issuance.go
│   │   ├── proposal_nominee.go
│   │   ├── proposal_medical_info.go
│   │   ├── proposal_agent.go
│   │   └── ... (other entities)
│   └── port/
│       ├── request.go
│       └── response.go
├── handler/
│   ├── quote.go
│   ├── proposal.go
│   ├── aadhaar.go
│   ├── approval.go
│   ├── policy.go
│   ├── lookup.go
│   ├── validation.go
│   ├── calculation.go
│   ├── document.go
│   ├── status.go
│   ├── workflow.go
│   ├── bulk_upload.go
│   ├── request.go
│   ├── request_*_validator.go (auto-generated)
│   └── response/
│       ├── quote.go
│       ├── proposal.go
│       ├── aadhaar.go
│       ├── approval.go
│       ├── policy.go
│       └── common.go
├── repo/
│   └── postgres/
│       ├── quote_repository.go
│       ├── proposal_repository.go
│       ├── product_repository.go
│       └── ... (other repositories)
├── workflows/
│   ├── policy_issuance_workflow.go
│   ├── instant_issuance_workflow.go
│   ├── bulk_proposal_upload_workflow.go
│   ├── state/
│   │   └── proposal_workflow_state.go
│   └── activities/
│       ├── proposal_activities.go
│       ├── aadhaar_activities.go
│       └── bulk_upload_activities.go
└── db/
    └── migrations/
        └── policy_issue_schema.sql
```

---

## Development Order

1. **Foundation** (Phase 0)
2. **Quote Module** (Phase 1) - Simple, no Temporal
3. **Proposal Core** (Phase 2) - Main workflow
4. **Aadhaar Flow** (Phase 3) - Instant issuance
5. **Approval Workflow** (Phase 4) - Signal-based
6. **Policy Lifecycle** (Phase 5) - FLC workflow
7. **Supporting APIs** (Phases 6-9) - Can be parallel
8. **Integration & Testing** (Phase 10)

---

## Quality Checklist

- [ ] All Swagger fields mapped
- [ ] All business rules (BR-*) implemented
- [ ] All validation rules (VR-*) implemented
- [ ] pgx.Batch used for 2+ queries
- [ ] Workflow state optimization for data reuse
- [ ] dblib library used (not raw pgx)
- [ ] Squirrel for simple queries, Raw SQL for complex
- [ ] Audit columns populated (created_at, updated_at, deleted_at)
- [ ] Column names match DDL exactly
- [ ] No interfaces for Repositories/Workflows
- [ ] Traceability comments for all rules
