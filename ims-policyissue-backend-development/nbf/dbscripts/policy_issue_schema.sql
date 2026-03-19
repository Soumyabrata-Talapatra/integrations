-- ============================================
-- Policy Issue Service - PostgreSQL Database Schema
-- Database: policy_issue_db
-- PostgreSQL Version: 16
-- Generated: February 2026
-- Source: policy_issue_requirements.md, policy_issue_swagger.yaml
-- ============================================
-- DESIGN NOTES:
-- 1. Normalized into phase-based tables to reduce burden on single proposals table
-- 2. proposal_indexing: Indexing/entry phase data
-- 3. proposal_data_entry: Data entry phase data  
-- 4. proposal_qc_review: QC review phase data
-- 5. proposal_approval: Approval phase data
-- 6. proposal_issuance: Issuance and post-issuance data
-- 7. Added customer_id explicitly to proposals core
-- 8. Added medical_examiner_id for medical examination tracking
-- 9. Added proposal_missing_documents for QC/Approver rejections
-- ============================================

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "pg_trgm";

-- Create sequences for auto-incrementing IDs
CREATE SEQUENCE IF NOT EXISTS policy_issue_seq START 1 INCREMENT 1;

-- ============================================
-- ENUM Types
-- Reference: BR-POL-015 (State Machine), BR-POL-011/012 (Eligibility)
-- ============================================

CREATE TYPE policy_type_enum AS ENUM ('PLI', 'RPLI');

CREATE TYPE proposal_status_enum AS ENUM (
    'DRAFT',
    'INDEXED',
    'DATA_ENTRY',
    'QC_PENDING',
    'QC_APPROVED',
    'QC_REJECTED',
    'QC_RETURNED',
    'PENDING_MEDICAL',
    'MEDICAL_APPROVED',
    'MEDICAL_REJECTED',
    'APPROVAL_PENDING',
    'APPROVED',
    'REJECTED',
    'ISSUED',
    'DISPATCHED',
    'FREE_LOOK_ACTIVE',
    'ACTIVE',
    'FLC_CANCELLED',
    'CANCELLED_DEATH'
);

CREATE TYPE premium_frequency_enum AS ENUM ('MONTHLY', 'QUARTERLY', 'HALF_YEARLY', 'YEARLY');

CREATE TYPE entry_path_enum AS ENUM ('WITHOUT_AADHAAR', 'WITH_AADHAAR', 'BULK_UPLOAD', 'QUOTE_CONVERSION');

CREATE TYPE channel_enum AS ENUM ('DIRECT', 'AGENCY', 'WEB', 'MOBILE', 'POS', 'CSC');

CREATE TYPE policy_taken_under_enum AS ENUM ('HUF', 'MWPA', 'OTHER');

CREATE TYPE medical_status_enum AS ENUM ('NOT_REQUIRED', 'PENDING', 'APPROVED', 'REJECTED', 'EXPIRED');

CREATE TYPE qr_decision_enum AS ENUM ('APPROVED', 'REJECTED', 'RETURNED');

CREATE TYPE approver_decision_enum AS ENUM ('APPROVED', 'REJECTED');

CREATE TYPE flc_status_enum AS ENUM ('NOT_STARTED', 'ACTIVE', 'EXPIRED', 'CANCELLED');

CREATE TYPE payment_method_enum AS ENUM ('CASH', 'CHEQUE', 'DD', 'ONLINE', 'POSB', 'NACH');

CREATE TYPE subsequent_payment_mode_enum AS ENUM ('CASH', 'ONLINE', 'NACH', 'STANDING_INSTRUCTION', 'POSB');

CREATE TYPE premium_payer_type_enum AS ENUM ('SELF', 'EMPLOYER', 'DDO', 'THIRD_PARTY');

CREATE TYPE age_proof_type_enum AS ENUM (
    'AADHAAR',
    'BIRTH_CERTIFICATE',
    'SCHOOL_CERTIFICATE',
    'PASSPORT',
    'VOTER_ID',
    'DRIVING_LICENSE',
    'PAN',
    'OTHER_STANDARD',
    'NON_STANDARD_AFFIDAVIT',
    'NON_STANDARD_DECLARATION'
);

CREATE TYPE gender_enum AS ENUM ('MALE', 'FEMALE', 'OTHER');

CREATE TYPE marital_status_enum AS ENUM ('SINGLE', 'MARRIED', 'DIVORCED', 'WIDOWED');

CREATE TYPE salutation_enum AS ENUM ('MR', 'MRS', 'MS', 'DR', 'SHRI', 'SMT', 'KUM');

CREATE TYPE relationship_enum AS ENUM (
    'FATHER', 'MOTHER', 'SPOUSE', 'SON', 'DAUGHTER', 'BROTHER', 'SISTER',
    'GRANDFATHER', 'GRANDMOTHER', 'UNCLE', 'AUNT', 'NEPHEW', 'NIECE',
    'FRIEND', 'OTHER'
);

CREATE TYPE deformity_type_enum AS ENUM ('CONGENITAL', 'NON_CONGENITAL', 'BOTH');

CREATE TYPE habit_frequency_enum AS ENUM ('NO', 'FREQUENTLY', 'OCCASIONALLY');

CREATE TYPE trust_type_enum AS ENUM ('INDIVIDUAL', 'CORPORATE');

CREATE TYPE proposer_relationship_enum AS ENUM ('PARENT', 'SPOUSE', 'EMPLOYER', 'HUF_KARTA', 'GUARDIAN', 'OTHER');

CREATE TYPE mandate_type_enum AS ENUM ('ECS', 'NACH', 'SI', 'POSB_AUTO_DEBIT');

CREATE TYPE mandate_status_enum AS ENUM ('PENDING', 'ACTIVE', 'SUSPENDED', 'CANCELLED');

CREATE TYPE rider_status_enum AS ENUM ('ACTIVE', 'LAPSED', 'CANCELLED');

CREATE TYPE quote_status_enum AS ENUM ('GENERATED', 'CONVERTED', 'EXPIRED');

CREATE TYPE product_category_enum AS ENUM ('WLA', 'CWLA', 'EA', 'AEA', 'JLA', 'CHILD', 'TEN_YEAR');

CREATE TYPE flc_start_date_rule_enum AS ENUM ('DISPATCH_DATE', 'DELIVERY_DATE', 'EMAIL_SENT_DATE');

CREATE TYPE document_type_enum AS ENUM (
    'PROPOSAL_FORM', 'DOB_PROOF', 'ADDRESS_PROOF', 'PHOTO_ID', 
    'MEDICAL_REPORT', 'PAYMENT_COPY', 'HEALTH_DECLARATION', 'PHOTO',
    'INCOME_PROOF', 'EMPLOYMENT_PROOF', 'OTHER'
);

CREATE TYPE policy_category_enum AS ENUM ('PLI_RPLI', 'NON_PLI_RPLI', 'OTHER_COMPANY');

CREATE TYPE bulk_upload_status_enum AS ENUM ('PROCESSING', 'COMPLETED', 'FAILED');

CREATE TYPE change_type_enum AS ENUM ('INSERT', 'UPDATE', 'DELETE');

CREATE TYPE missing_document_stage_enum AS ENUM ('QC_REVIEW', 'APPROVAL');

CREATE TYPE missing_document_status_enum AS ENUM ('PENDING', 'UPLOADED', 'WAIVED');

-- ============================================
-- Tables
-- ============================================

-- E-001: product_catalog
-- Description: Configurable product catalog for all 12 PLI/RPLI products

CREATE TABLE product_catalog (
    product_code VARCHAR(20) PRIMARY KEY,
    product_name VARCHAR(100) NOT NULL,
    product_type policy_type_enum NOT NULL,
    product_category product_category_enum NOT NULL,
    min_sum_assured DECIMAL(15,2) NOT NULL,
    max_sum_assured DECIMAL(15,2),
    min_entry_age INTEGER NOT NULL,
    max_entry_age INTEGER NOT NULL,
    max_maturity_age INTEGER,
    min_term INTEGER NOT NULL,
    premium_ceasing_age_options JSONB,
    available_frequencies JSONB NOT NULL,
    medical_sa_threshold DECIMAL(15,2),
    is_sa_decrease_allowed BOOLEAN NOT NULL DEFAULT TRUE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    effective_from DATE NOT NULL,
    effective_to DATE,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    CONSTRAINT chk_min_max_sa CHECK (min_sum_assured > 0 AND (max_sum_assured IS NULL OR max_sum_assured >= min_sum_assured)),
    CONSTRAINT chk_entry_age_range CHECK (min_entry_age >= 0 AND max_entry_age >= min_entry_age),
    CONSTRAINT chk_term_positive CHECK (min_term > 0)
);

COMMENT ON TABLE product_catalog IS 'E-001: Product catalog for PLI/RPLI products. BR-POL-011, BR-POL-012';

-- E-002: quote
-- Description: Generated premium quote record

CREATE TABLE quote (
    quote_id BIGINT PRIMARY KEY DEFAULT nextval('policy_issue_seq'),
    quote_ref_number VARCHAR(30) UNIQUE NOT NULL,
    product_code VARCHAR(20) NOT NULL REFERENCES product_catalog(product_code),
    policy_type policy_type_enum NOT NULL,
    customer_id BIGINT,
    proposer_name VARCHAR(200),
    proposer_dob DATE,
    proposer_gender gender_enum,
    proposer_mobile VARCHAR(15),
    proposer_email VARCHAR(100),
    sum_assured DECIMAL(15,2) NOT NULL,
    policy_term INTEGER NOT NULL,
    payment_frequency premium_frequency_enum NOT NULL,
    base_premium DECIMAL(12,2) NOT NULL,
    gst_amount DECIMAL(10,2) NOT NULL,
    total_payable DECIMAL(12,2) NOT NULL,
    maturity_value DECIMAL(15,2),
    bonus_rate DECIMAL(5,2),
    channel channel_enum NOT NULL,
    status quote_status_enum NOT NULL DEFAULT 'GENERATED',
    converted_proposal_id BIGINT,
    pdf_document_id VARCHAR(50),
    created_by BIGINT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE,
    deleted_at TIMESTAMP WITH TIME ZONE,
    version INTEGER NOT NULL DEFAULT 1,
    metadata JSONB,
    search_vector TSVECTOR,
    
    CONSTRAINT chk_quote_sa_positive CHECK (sum_assured > 0),
    CONSTRAINT chk_quote_premium_positive CHECK (base_premium > 0),
    CONSTRAINT chk_quote_term_positive CHECK (policy_term > 0)
);

COMMENT ON TABLE quote IS 'E-002: Premium quote records. BR-POL-001, BR-POL-002, BR-POL-003';
COMMENT ON COLUMN quote.customer_id IS 'Customer ID if quote is linked to existing customer';

-- E-003: policy_number_sequence
CREATE TABLE policy_number_sequence (
    sequence_id BIGINT PRIMARY KEY DEFAULT nextval('policy_issue_seq'),
    product_type policy_type_enum NOT NULL,
    series_prefix VARCHAR(10) NOT NULL,
    next_value BIGINT NOT NULL DEFAULT 1,
    format_pattern VARCHAR(50) NOT NULL DEFAULT '{prefix}-{year}-{value:06d}',
    
    CONSTRAINT uq_product_series UNIQUE (product_type, series_prefix)
);

-- E-004: free_look_config
CREATE TABLE free_look_config (
    config_id BIGINT PRIMARY KEY DEFAULT nextval('policy_issue_seq'),
    channel channel_enum NOT NULL,
    product_type policy_type_enum,
    period_days INTEGER NOT NULL DEFAULT 15,
    start_date_rule flc_start_date_rule_enum NOT NULL DEFAULT 'DISPATCH_DATE',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    CONSTRAINT chk_flc_period_positive CHECK (period_days > 0)
);

-- E-005: approval_routing_config
CREATE TABLE approval_routing_config (
    config_id BIGINT PRIMARY KEY DEFAULT nextval('policy_issue_seq'),
    sa_min DECIMAL(15,2) NOT NULL,
    sa_max DECIMAL(15,2) NOT NULL,
    approver_level INTEGER NOT NULL CHECK (approver_level BETWEEN 1 AND 3),
    approver_role VARCHAR(50) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    CONSTRAINT chk_sa_range CHECK (sa_max >= sa_min)
);

COMMENT ON TABLE approval_routing_config IS 'E-005: Approval routing by SA. BR-POL-016';

-- E-006: bulk_upload_batch
CREATE TABLE bulk_upload_batch (
    batch_id BIGINT PRIMARY KEY DEFAULT nextval('policy_issue_seq'),
    file_name VARCHAR(200) NOT NULL,
    total_rows INTEGER NOT NULL,
    success_count INTEGER NOT NULL DEFAULT 0,
    failure_count INTEGER NOT NULL DEFAULT 0,
    error_report_doc_id VARCHAR(50),
    status bulk_upload_status_enum NOT NULL DEFAULT 'PROCESSING',
    uploaded_by BIGINT NOT NULL,
    uploaded_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,
    metadata JSONB,
    
    CONSTRAINT chk_row_counts CHECK (success_count + failure_count <= total_rows)
);

-- ============================================
-- E-007: proposals (Core Entity - Minimal)
-- Description: Core proposal record with shared data across all phases
-- ============================================

CREATE TABLE proposals (
    proposal_id BIGINT PRIMARY KEY DEFAULT nextval('policy_issue_seq'),
    proposal_number VARCHAR(30) UNIQUE NOT NULL,
    quote_ref_number VARCHAR(30) REFERENCES quote(quote_ref_number),
    
    --Customer References (MANDATORY)
    customer_id BIGINT NOT NULL,
    spouse_customer_id BIGINT,
    proposer_customer_id BIGINT,
    is_proposer_same_as_insured BOOLEAN NOT NULL DEFAULT TRUE,
    premium_payer_type premium_payer_type_enum DEFAULT 'SELF',
    payer_customer_id BIGINT,
    
    -- Product & Coverage
    product_code VARCHAR(20) NOT NULL REFERENCES product_catalog(product_code),
    policy_type policy_type_enum NOT NULL,
    sum_assured DECIMAL(15,2) NOT NULL,
    policy_term INTEGER NOT NULL,
    premium_ceasing_age INTEGER,
    premium_payment_frequency premium_frequency_enum NOT NULL,
    
    -- Premium Summary (calculated values)
    base_premium DECIMAL(12,2),
    gst_amount DECIMAL(10,2),
    total_premium DECIMAL(12,2),
    annual_premium_equivalent DECIMAL(12,2),
    modal_premium DECIMAL(12,2),
    additional_premium DECIMAL(12,2),
    
    -- Workflow Status
    status proposal_status_enum NOT NULL DEFAULT 'DRAFT',
    current_stage VARCHAR(20) NOT NULL DEFAULT 'INDEXING',
    entry_path entry_path_enum NOT NULL,
    channel channel_enum NOT NULL,
    workflow_id VARCHAR(100),
    
    -- Created/Updated metadata
    created_by BIGINT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_by BIGINT,
    deleted_at TIMESTAMP WITH TIME ZONE,
    version INTEGER NOT NULL DEFAULT 1,
    metadata JSONB,
    search_vector TSVECTOR,
    
    CONSTRAINT chk_proposal_sa_positive CHECK (sum_assured > 0),
    CONSTRAINT chk_proposal_term_positive CHECK (policy_term > 0),
    CONSTRAINT chk_proposer_required CHECK (
        (is_proposer_same_as_insured = TRUE AND proposer_customer_id IS NULL) OR
        (is_proposer_same_as_insured = FALSE AND proposer_customer_id IS NOT NULL)
    )
);

COMMENT ON TABLE proposals IS 'E-007: Core proposal record (minimal). Phase data in separate tables to reduce burden';
COMMENT ON COLUMN proposals.customer_id IS 'MANDATORY: Primary insured customer ID from Customer Service';
COMMENT ON COLUMN proposals.current_stage IS 'Current processing stage: INDEXING, DATA_ENTRY, QC_REVIEW, APPROVAL, MEDICAL, ISSUANCE';

-- ============================================
-- E-007A: proposal_indexing (Indexing Phase)
-- Description: Data specific to indexing/entry phase
-- ============================================

CREATE TABLE proposal_indexing (
    indexing_id BIGINT PRIMARY KEY DEFAULT nextval('policy_issue_seq'),
    proposal_id BIGINT NOT NULL REFERENCES proposals(proposal_id) ON DELETE CASCADE,
    
    -- Indexing Dates
    declaration_date DATE NOT NULL,
    receipt_date DATE NOT NULL,
    indexing_date DATE NOT NULL,
    proposal_date DATE NOT NULL,
    
    -- Location/Office
    po_code VARCHAR(20) NOT NULL,
    issue_circle VARCHAR(50),
    issue_ho VARCHAR(50),
    issue_post_office VARCHAR(50),
    
    -- Initial Payment
    first_premium_paid BOOLEAN DEFAULT FALSE,
    first_premium_date DATE,
    first_premium_reference VARCHAR(50),
    first_premium_receipt_number VARCHAR(50),
    premium_payment_method payment_method_enum,
    initial_premium DECIMAL(12,2),
    short_excess_premium DECIMAL(12,2),
    
    -- Bulk Upload reference
    bulk_upload_batch_id BIGINT REFERENCES bulk_upload_batch(batch_id),
    bulk_upload_row_number INTEGER,
    
    -- Indexing metadata
    indexed_by BIGINT NOT NULL,
    indexed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    CONSTRAINT chk_date_chain CHECK (
        declaration_date <= receipt_date AND 
        receipt_date <= indexing_date AND 
        indexing_date <= proposal_date
    ),
    CONSTRAINT uq_proposal_indexing UNIQUE (proposal_id)
);

COMMENT ON TABLE proposal_indexing IS 'E-007A: Indexing phase data. BR-POL-018 (Date Chain), VAL-POL-001';
COMMENT ON COLUMN proposal_indexing.declaration_date IS 'Date of declaration. VAL-POL-024';
COMMENT ON COLUMN proposal_indexing.receipt_date IS 'Application receipt date. VAL-POL-025';

-- ============================================
-- E-007B: proposal_data_entry (Data Entry Phase)
-- Description: Data specific to CPC data entry phase
-- ============================================

CREATE TABLE proposal_data_entry (
    data_entry_id BIGINT PRIMARY KEY DEFAULT nextval('policy_issue_seq'),
    proposal_id BIGINT NOT NULL REFERENCES proposals(proposal_id) ON DELETE CASCADE,
    
    --Special Policy Types
    policy_taken_under policy_taken_under_enum,
    
    -- Age Proof
    age_proof_type age_proof_type_enum,
    aadhaar_photo_document_id VARCHAR(50),
    
    -- Subsequent Payment
    subsequent_payment_mode subsequent_payment_mode_enum,
    
    -- Employment Details (for PLI)
    employment_type VARCHAR(50),
    employer_name VARCHAR(200),
    designation VARCHAR(100),
    pao_code VARCHAR(30),
    ddo_code VARCHAR(30),
    ddo_mobile_no VARCHAR(15),
    ddo_designation VARCHAR(100),
    ddo_email VARCHAR(100),
    ddo_location VARCHAR(200),
    date_of_joining DATE,
    
    -- Data Entry metadata
    data_entry_by BIGINT,
    data_entry_started_at TIMESTAMP WITH TIME ZONE,
    data_entry_completed_at TIMESTAMP WITH TIME ZONE,
    data_entry_status VARCHAR(20) DEFAULT 'IN_PROGRESS',
    
    CONSTRAINT uq_proposal_data_entry UNIQUE (proposal_id)
);

COMMENT ON TABLE proposal_data_entry IS 'E-007B: Data entry phase data captured by CPC operators';

-- ============================================
-- E-007C: proposal_qc_review (QC Review Phase)
-- Description: Quality Check review data
-- ============================================

CREATE TABLE proposal_qc_review (
    qc_review_id BIGINT PRIMARY KEY DEFAULT nextval('policy_issue_seq'),
    proposal_id BIGINT NOT NULL REFERENCES proposals(proposal_id) ON DELETE CASCADE,
    
    -- QC Decision
    qr_decision qr_decision_enum,
    qr_decision_by BIGINT,
    qr_decision_at TIMESTAMP WITH TIME ZONE,
    qr_comments TEXT,
    
    -- QC Review metadata
    qc_assigned_to BIGINT,
    qc_assigned_at TIMESTAMP WITH TIME ZONE,
    qc_review_started_at TIMESTAMP WITH TIME ZONE,
    qc_review_completed_at TIMESTAMP WITH TIME ZONE,
    
    -- Return tracking
    return_count INTEGER DEFAULT 0,
    last_return_reason TEXT,
    last_returned_at TIMESTAMP WITH TIME ZONE,
    
    CONSTRAINT chk_qr_rejection_comments CHECK (
        (qr_decision IN ('REJECTED', 'RETURNED') AND qr_comments IS NOT NULL) OR
        (qr_decision NOT IN ('REJECTED', 'RETURNED'))
    ),
    CONSTRAINT uq_proposal_qc_review UNIQUE (proposal_id)
);

COMMENT ON TABLE proposal_qc_review IS 'E-007C: QC review phase data. BR-POL-017';

-- ============================================
-- E-007D: proposal_medical (Medical Phase)
-- Description: Medical examination data
-- ============================================

CREATE TABLE proposal_medical (
    proposal_medical_id BIGINT PRIMARY KEY DEFAULT nextval('policy_issue_seq'),
    proposal_id BIGINT NOT NULL REFERENCES proposals(proposal_id) ON DELETE CASCADE,
    
    -- Medical Requirement
    is_medical_required BOOLEAN,
    medical_status medical_status_enum,
    medical_sa_threshold DECIMAL(15,2),
    
    -- Medical Appointment
    medical_appointment_id VARCHAR(50),
    medical_appointment_date DATE,
    medical_appointment_status VARCHAR(20),
    medical_facility_id BIGINT,
    medical_facility_name VARCHAR(200),
    
    -- Medical Examiner (NEW - Added per requirement)
    medical_examiner_id BIGINT,
    medical_examiner_name VARCHAR(200),
    medical_examiner_license VARCHAR(50),
    
    -- Medical Certificate
    medical_certificate_date DATE,
    medical_certificate_valid_until DATE,
    medical_report_document_id VARCHAR(50),
    
    -- Medical Result
    medical_decision VARCHAR(20),
    medical_decision_by BIGINT,
    medical_decision_at TIMESTAMP WITH TIME ZONE,
    medical_decision_comments TEXT,
    medical_rejection_reason TEXT,
    
    -- Medical metadata
    medical_requested_at TIMESTAMP WITH TIME ZONE,
    medical_completed_at TIMESTAMP WITH TIME ZONE,
    
    CONSTRAINT chk_medical_cert_validity CHECK (
        medical_certificate_date IS NULL OR 
        medical_certificate_valid_until IS NULL OR 
        medical_certificate_valid_until >= medical_certificate_date
    ),
    CONSTRAINT uq_proposal_medical UNIQUE (proposal_id)
);

COMMENT ON TABLE proposal_medical IS 'E-007D: Medical examination phase data. BR-POL-013, BR-POL-019';
COMMENT ON COLUMN proposal_medical.medical_examiner_id IS 'ID of assigned medical examiner/doctor from Medical Appointment Service';
COMMENT ON COLUMN proposal_medical.medical_certificate_date IS 'Medical exam date. BR-POL-019 (60 day validity)';

-- ============================================
-- E-007E: proposal_approval (Approval Phase)
-- Description: Approval workflow data
-- ============================================

CREATE TABLE proposal_approval (
    approval_id BIGINT PRIMARY KEY DEFAULT nextval('policy_issue_seq'),
    proposal_id BIGINT NOT NULL REFERENCES proposals(proposal_id) ON DELETE CASCADE,
    
    -- Approval Routing
    approval_level INTEGER CHECK (approval_level BETWEEN 1 AND 3),
    approver_role VARCHAR(50),
    approval_routing_rule_id BIGINT,
    
    -- Approver Decision
    approver_decision approver_decision_enum,
    approver_decision_by BIGINT,
    approver_decision_at TIMESTAMP WITH TIME ZONE,
    approver_comments TEXT,
    approver_rejection_reason TEXT,
    
    -- Approval metadata
    assigned_approver_id BIGINT,
    approval_assigned_at TIMESTAMP WITH TIME ZONE,
    approval_due_date DATE,
    approval_reminder_sent BOOLEAN DEFAULT FALSE,
    approval_reminder_count INTEGER DEFAULT 0,
    
    CONSTRAINT chk_approver_rejection_comments CHECK (
        (approver_decision = 'REJECTED' AND approver_comments IS NOT NULL) OR
        (approver_decision IS NULL OR approver_decision != 'REJECTED')
    ),
    CONSTRAINT uq_proposal_approval UNIQUE (proposal_id)
);

COMMENT ON TABLE proposal_approval IS 'E-007E: Approval phase data. BR-POL-016';
COMMENT ON COLUMN proposal_approval.approval_level IS '1, 2, or 3 based on SA bracket. BR-POL-016';

-- ============================================
-- E-007F: proposal_issuance (Issuance Phase)
-- Description: Policy issuance and post-issuance data
-- ============================================

CREATE TABLE proposal_issuance (
    issuance_id BIGINT PRIMARY KEY DEFAULT nextval('policy_issue_seq'),
    proposal_id BIGINT NOT NULL REFERENCES proposals(proposal_id) ON DELETE CASCADE,
    
    -- Policy Number & Dates
    policy_number VARCHAR(30) UNIQUE,
    policy_issue_date DATE,
    acceptance_date DATE,
    policy_commencement_date DATE,
    maturity_date DATE,
    
    -- Bond Generation
    bond_generated BOOLEAN DEFAULT FALSE,
    bond_document_id VARCHAR(50),
    bond_generated_at TIMESTAMP WITH TIME ZONE,
    bond_generated_by BIGINT,
    
    -- Dispatch
    dispatch_date DATE,
    delivery_date DATE,
    dispatch_method VARCHAR(20),
    tracking_number VARCHAR(50),
    
    -- Free Look Period
    flc_start_date DATE,
    flc_end_date DATE,
    flc_status flc_status_enum,
    flc_config_id BIGINT REFERENCES free_look_config(config_id),
    
    -- FLC Cancellation
    flc_cancel_requested_at TIMESTAMP WITH TIME ZONE,
    flc_cancel_reason TEXT,
    flc_cancel_refund_amount DECIMAL(12,2),
    flc_cancel_processed_at TIMESTAMP WITH TIME ZONE,
    
    -- Commission
    commission_triggered BOOLEAN DEFAULT FALSE,
    commission_triggered_at TIMESTAMP WITH TIME ZONE,
    
    CONSTRAINT chk_policy_number_on_issued CHECK (
        (policy_number IS NOT NULL) OR
        (policy_number IS NULL)
    ),
    CONSTRAINT chk_flc_dates CHECK (
        flc_end_date IS NULL OR flc_start_date IS NULL OR flc_end_date > flc_start_date
    ),
    CONSTRAINT uq_proposal_issuance UNIQUE (proposal_id)
);

COMMENT ON TABLE proposal_issuance IS 'E-007F: Issuance and post-issuance phase data. BR-POL-021, BR-POL-023';
COMMENT ON COLUMN proposal_issuance.flc_start_date IS 'Free look period start. BR-POL-021, BR-POL-028';
COMMENT ON COLUMN proposal_issuance.policy_number IS 'Generated policy number. FR-POL-023';

-- ============================================
-- E-008: proposal_nominee
-- ============================================

CREATE TABLE proposal_nominee (
    nominee_id BIGINT PRIMARY KEY DEFAULT nextval('policy_issue_seq'),
    proposal_id BIGINT NOT NULL REFERENCES proposals(proposal_id) ON DELETE CASCADE,
    salutation salutation_enum,
    first_name VARCHAR(100) NOT NULL,
    middle_name VARCHAR(100),
    last_name VARCHAR(100) NOT NULL,
    gender gender_enum NOT NULL,
    date_of_birth DATE NOT NULL,
    is_minor BOOLEAN NOT NULL DEFAULT FALSE,
    relationship relationship_enum NOT NULL,
    share_percentage DECIMAL(5,2) NOT NULL,
    address_line1 VARCHAR(200),
    city VARCHAR(100),
    state VARCHAR(50),
    pin_code VARCHAR(10),
    phone VARCHAR(15),
    email VARCHAR(100),
    appointee_name VARCHAR(200),
    appointee_relationship VARCHAR(30),
    appointee_dob DATE,
    appointee_address TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    CONSTRAINT chk_share_percentage CHECK (share_percentage > 0 AND share_percentage <= 100),
    CONSTRAINT chk_minor_appointee CHECK (
        (is_minor = TRUE AND appointee_name IS NOT NULL AND appointee_relationship IS NOT NULL) OR
        (is_minor = FALSE)
    )
);

COMMENT ON TABLE proposal_nominee IS 'E-008: Nominee details. Max 3 per proposal. VAL-POL-003, VAL-POL-004';

-- ============================================
-- E-009: proposal_medical_info
-- ============================================

CREATE TABLE proposal_medical_info (
    medical_info_id BIGINT PRIMARY KEY DEFAULT nextval('policy_issue_seq'),
    proposal_id BIGINT NOT NULL REFERENCES proposals(proposal_id) ON DELETE CASCADE,
    insured_index INTEGER NOT NULL DEFAULT 1 CHECK (insured_index IN (1, 2)),
    is_sound_health BOOLEAN NOT NULL DEFAULT TRUE,
    disease_tb BOOLEAN NOT NULL DEFAULT FALSE,
    disease_cancer BOOLEAN NOT NULL DEFAULT FALSE,
    disease_paralysis BOOLEAN NOT NULL DEFAULT FALSE,
    disease_insanity BOOLEAN NOT NULL DEFAULT FALSE,
    disease_heart_lungs BOOLEAN NOT NULL DEFAULT FALSE,
    disease_kidney BOOLEAN NOT NULL DEFAULT FALSE,
    disease_brain BOOLEAN NOT NULL DEFAULT FALSE,
    disease_hiv BOOLEAN NOT NULL DEFAULT FALSE,
    disease_hepatitis_b BOOLEAN NOT NULL DEFAULT FALSE,
    disease_epilepsy BOOLEAN NOT NULL DEFAULT FALSE,
    disease_nervous BOOLEAN NOT NULL DEFAULT FALSE,
    disease_liver BOOLEAN NOT NULL DEFAULT FALSE,
    disease_leprosy BOOLEAN NOT NULL DEFAULT FALSE,
    disease_physical_deformity BOOLEAN NOT NULL DEFAULT FALSE,
    disease_other BOOLEAN NOT NULL DEFAULT FALSE,
    disease_details TEXT,
    family_hereditary BOOLEAN NOT NULL DEFAULT FALSE,
    family_hereditary_details TEXT,
    medical_leave_3yr BOOLEAN NOT NULL DEFAULT FALSE,
    leave_kind VARCHAR(50),
    leave_period VARCHAR(50),
    leave_ailment VARCHAR(200),
    hospital_name VARCHAR(200),
    hospitalization_from DATE,
    hospitalization_to DATE,
    physical_deformity BOOLEAN NOT NULL DEFAULT FALSE,
    deformity_type deformity_type_enum,
    family_doctor_name VARCHAR(200),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    CONSTRAINT uq_proposal_insured_medical UNIQUE (proposal_id, insured_index)
);

-- ============================================
-- E-010: proposal_enhanced_medical
-- ============================================

CREATE TABLE proposal_enhanced_medical (
    enhanced_medical_id BIGINT PRIMARY KEY DEFAULT nextval('policy_issue_seq'),
    proposal_id BIGINT NOT NULL REFERENCES proposals(proposal_id) ON DELETE CASCADE,
    insured_index INTEGER NOT NULL DEFAULT 1 CHECK (insured_index IN (1, 2)),
    tests_investigations BOOLEAN NOT NULL DEFAULT FALSE,
    diabetes BOOLEAN NOT NULL DEFAULT FALSE,
    blood_pressure BOOLEAN NOT NULL DEFAULT FALSE,
    oncologist_visit BOOLEAN NOT NULL DEFAULT FALSE,
    ailment_over_1_week BOOLEAN NOT NULL DEFAULT FALSE,
    thyroid BOOLEAN NOT NULL DEFAULT FALSE,
    angioplasty_surgery BOOLEAN NOT NULL DEFAULT FALSE,
    eye_ear_nose BOOLEAN NOT NULL DEFAULT FALSE,
    anaemia_blood BOOLEAN NOT NULL DEFAULT FALSE,
    musculoskeletal BOOLEAN NOT NULL DEFAULT FALSE,
    female_abortion_miscarriage BOOLEAN DEFAULT FALSE,
    female_gynaecological BOOLEAN DEFAULT FALSE,
    female_reproductive BOOLEAN DEFAULT FALSE,
    habit_smoke_tobacco habit_frequency_enum DEFAULT 'NO',
    habit_alcohol habit_frequency_enum DEFAULT 'NO',
    habit_drugs habit_frequency_enum DEFAULT 'NO',
    habit_adverse habit_frequency_enum DEFAULT 'NO',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    CONSTRAINT uq_proposal_enhanced_medical UNIQUE (proposal_id, insured_index)
);

-- ============================================
-- E-011: proposal_agent
-- ============================================

CREATE TABLE proposal_agent (
    proposal_agent_id BIGINT PRIMARY KEY DEFAULT nextval('policy_issue_seq'),
    proposal_id BIGINT NOT NULL REFERENCES proposals(proposal_id) ON DELETE CASCADE,
    agent_id VARCHAR(30) NOT NULL,
    agent_salutation VARCHAR(10),
    agent_name VARCHAR(200),
    agent_mobile VARCHAR(15),
    agent_email VARCHAR(100),
    agent_landline VARCHAR(15),
    agent_std_code VARCHAR(10),
    receives_correspondence BOOLEAN NOT NULL DEFAULT FALSE,
    opportunity_id VARCHAR(50),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    CONSTRAINT uq_proposal_agent UNIQUE (proposal_id)
);

-- ============================================
-- E-012: proposal_mwpa_trustee
-- ============================================

CREATE TABLE proposal_mwpa_trustee (
    trustee_id BIGINT PRIMARY KEY DEFAULT nextval('policy_issue_seq'),
    proposal_id BIGINT NOT NULL REFERENCES proposals(proposal_id) ON DELETE CASCADE,
    trust_type trust_type_enum NOT NULL,
    trustee_name VARCHAR(200) NOT NULL,
    trustee_dob DATE,
    relationship VARCHAR(30),
    address TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    CONSTRAINT uq_proposal_mwpa_trustee UNIQUE (proposal_id)
);

-- ============================================
-- E-013: proposal_huf_member
-- ============================================

CREATE TABLE proposal_huf_member (
    huf_member_id BIGINT PRIMARY KEY DEFAULT nextval('policy_issue_seq'),
    proposal_id BIGINT NOT NULL REFERENCES proposals(proposal_id) ON DELETE CASCADE,
    is_financed_huf BOOLEAN NOT NULL DEFAULT FALSE,
    karta_name VARCHAR(200),
    huf_pan VARCHAR(20),
    life_assured_different_from_karta BOOLEAN DEFAULT FALSE,
    karta_different_reason TEXT,
    member_name VARCHAR(200) NOT NULL,
    member_relationship VARCHAR(30) NOT NULL,
    member_age INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    CONSTRAINT chk_member_age_positive CHECK (member_age >= 0)
);

-- ============================================
-- E-014: proposal_existing_policy
-- ============================================

CREATE TABLE proposal_existing_policy (
    existing_policy_id BIGINT PRIMARY KEY DEFAULT nextval('policy_issue_seq'),
    proposal_id BIGINT NOT NULL REFERENCES proposals(proposal_id) ON DELETE CASCADE,
    policy_category policy_category_enum NOT NULL,
    policy_number VARCHAR(30),
    company_name VARCHAR(200),
    sum_assured DECIMAL(15,2),
    details TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    CONSTRAINT chk_existing_sa_positive CHECK (sum_assured IS NULL OR sum_assured > 0)
);

-- ============================================
-- E-015: proposal_status_history
-- ============================================

CREATE TABLE proposal_status_history (
    history_id BIGINT PRIMARY KEY DEFAULT nextval('policy_issue_seq'),
    proposal_id BIGINT NOT NULL REFERENCES proposals(proposal_id) ON DELETE CASCADE,
    from_status proposal_status_enum,
    to_status proposal_status_enum NOT NULL,
    changed_by BIGINT NOT NULL,
    changed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    comments TEXT,
    version INTEGER NOT NULL,
    metadata JSONB
);

-- ============================================
-- E-016: proposal_document_ref
-- ============================================

CREATE TABLE proposal_document_ref (
    doc_ref_id BIGINT PRIMARY KEY DEFAULT nextval('policy_issue_seq'),
    proposal_id BIGINT NOT NULL REFERENCES proposals(proposal_id) ON DELETE CASCADE,
    document_id VARCHAR(50) NOT NULL,
    document_type document_type_enum NOT NULL,
    file_name VARCHAR(200),
    file_size_bytes BIGINT,
    mime_type VARCHAR(100),
    uploaded_by BIGINT NOT NULL,
    uploaded_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    version INTEGER NOT NULL DEFAULT 1,
    comments TEXT,
    deleted_at TIMESTAMP WITH TIME ZONE
);

COMMENT ON TABLE proposal_document_ref IS 'E-016: Document references uploaded for proposal';

-- ============================================
-- E-017: proposal_missing_documents (NEW)
-- Description: Track missing documents when rejected by QC/Approver
-- ============================================

CREATE TABLE proposal_missing_documents (
    missing_doc_id BIGINT PRIMARY KEY DEFAULT nextval('policy_issue_seq'),
    proposal_id BIGINT NOT NULL REFERENCES proposals(proposal_id) ON DELETE CASCADE,
    
    -- Document Reference
    document_type document_type_enum NOT NULL,
    document_description VARCHAR(200),
    
    -- Stage where missing was noted
    stage missing_document_stage_enum NOT NULL,
    
    -- Who noted the missing document
    noted_by BIGINT NOT NULL,
    noted_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    notes TEXT,
    
    -- Resolution
    status missing_document_status_enum NOT NULL DEFAULT 'PENDING',
    resolved_by BIGINT,
    resolved_at TIMESTAMP WITH TIME ZONE,
    resolution_notes TEXT,
    
    -- Uploaded document reference (when resolved)
    uploaded_document_id BIGINT REFERENCES proposal_document_ref(doc_ref_id),
    
    -- Waiver
    waived BOOLEAN DEFAULT FALSE,
    waived_by BIGINT,
    waived_at TIMESTAMP WITH TIME ZONE,
    waiver_reason TEXT,
    
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

COMMENT ON TABLE proposal_missing_documents IS 'E-017: Missing documents noted during QC/Approval review';
COMMENT ON COLUMN proposal_missing_documents.stage IS 'Stage where missing document was identified: QC_REVIEW or APPROVAL';
COMMENT ON COLUMN proposal_missing_documents.noted_by IS 'User ID (QC or Approver) who noted the missing document';
COMMENT ON COLUMN proposal_missing_documents.status IS 'PENDING, UPLOADED, or WAIVED';

-- ============================================
-- E-018: proposal_proposer
-- ============================================

CREATE TABLE proposal_proposer (
    proposer_id BIGINT PRIMARY KEY DEFAULT nextval('policy_issue_seq'),
    proposal_id BIGINT NOT NULL REFERENCES proposals(proposal_id) ON DELETE CASCADE,
    customer_id BIGINT NOT NULL,
    relationship_to_insured proposer_relationship_enum NOT NULL,
    relationship_details VARCHAR(100),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    CONSTRAINT uq_proposal_proposer UNIQUE (proposal_id)
);

-- ============================================
-- E-019: proposal_payment_mandate
-- ============================================

CREATE TABLE proposal_payment_mandate (
    mandate_id BIGINT PRIMARY KEY DEFAULT nextval('policy_issue_seq'),
    proposal_id BIGINT NOT NULL REFERENCES proposals(proposal_id) ON DELETE CASCADE,
    mandate_type mandate_type_enum NOT NULL,
    bank_account_number VARCHAR(30) NOT NULL,
    bank_ifsc_code VARCHAR(11) NOT NULL,
    bank_name VARCHAR(100),
    mandate_reference VARCHAR(50),
    umrn VARCHAR(30),
    max_amount DECIMAL(12,2) NOT NULL,
    frequency premium_frequency_enum NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE,
    status mandate_status_enum NOT NULL DEFAULT 'PENDING',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    CONSTRAINT uq_proposal_mandate UNIQUE (proposal_id),
    CONSTRAINT chk_mandate_amount_positive CHECK (max_amount > 0),
    CONSTRAINT chk_ifsc_format CHECK (bank_ifsc_code ~ '^[A-Z]{4}0[A-Z0-9]{6}$')
);

-- ============================================
-- E-020: proposal_rider
-- ============================================

CREATE TABLE proposal_rider (
    rider_id BIGINT PRIMARY KEY DEFAULT nextval('policy_issue_seq'),
    proposal_id BIGINT NOT NULL REFERENCES proposals(proposal_id) ON DELETE CASCADE,
    rider_product_code VARCHAR(20) NOT NULL REFERENCES product_catalog(product_code),
    rider_sum_assured DECIMAL(15,2) NOT NULL,
    rider_term INTEGER NOT NULL,
    rider_premium DECIMAL(12,2),
    rider_gst DECIMAL(10,2),
    is_medical_required BOOLEAN DEFAULT FALSE,
    status rider_status_enum NOT NULL DEFAULT 'ACTIVE',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    
    CONSTRAINT chk_rider_sa_positive CHECK (rider_sum_assured > 0),
    CONSTRAINT chk_rider_term_positive CHECK (rider_term > 0)
);

-- ============================================
-- E-021: proposal_audit_log
-- ============================================

CREATE TABLE proposal_audit_log (
    audit_id BIGINT PRIMARY KEY DEFAULT nextval('policy_issue_seq'),
    proposal_id BIGINT NOT NULL REFERENCES proposals(proposal_id) ON DELETE CASCADE,
    entity_type VARCHAR(50) NOT NULL,
    entity_id BIGINT NOT NULL,
    field_name VARCHAR(100) NOT NULL,
    old_value TEXT,
    new_value TEXT,
    change_type change_type_enum NOT NULL,
    changed_by BIGINT NOT NULL,
    changed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    change_reason TEXT,
    metadata JSONB
);

-- ============================================
-- Indexes
-- ============================================

-- Product Catalog Indexes
CREATE INDEX idx_product_catalog_type ON product_catalog(product_type) WHERE deleted_at IS NULL;
CREATE INDEX idx_product_catalog_active ON product_catalog(is_active) WHERE is_active = TRUE AND deleted_at IS NULL;

-- Quote Indexes
CREATE INDEX idx_quote_ref_number ON quote(quote_ref_number);
CREATE INDEX idx_quote_customer ON quote(customer_id) WHERE customer_id IS NOT NULL;
CREATE INDEX idx_quote_status ON quote(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_quote_created_at ON quote(created_at) WHERE deleted_at IS NULL;

-- Proposal Core Indexes
CREATE INDEX idx_proposal_number ON proposals(proposal_number);
CREATE INDEX idx_proposal_customer ON proposals(customer_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_proposal_spouse ON proposals(spouse_customer_id) WHERE spouse_customer_id IS NOT NULL AND deleted_at IS NULL;
CREATE INDEX idx_proposal_proposer ON proposals(proposer_customer_id) WHERE proposer_customer_id IS NOT NULL AND deleted_at IS NULL;
CREATE INDEX idx_proposal_status ON proposals(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_proposal_current_stage ON proposals(current_stage) WHERE deleted_at IS NULL;
CREATE INDEX idx_proposal_product ON proposals(product_code) WHERE deleted_at IS NULL;
CREATE INDEX idx_proposal_created_at ON proposals(created_at) WHERE deleted_at IS NULL;

-- Proposal Composite Indexes
CREATE INDEX idx_proposal_status_created ON proposals(status, created_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_proposal_stage_status ON proposals(current_stage, status) WHERE deleted_at IS NULL;

-- Phase Table Indexes
CREATE INDEX idx_proposal_indexing_proposal ON proposal_indexing(proposal_id);
CREATE INDEX idx_proposal_indexing_po_code ON proposal_indexing(po_code);
CREATE INDEX idx_proposal_indexing_dates ON proposal_indexing(proposal_date);

CREATE INDEX idx_proposal_data_entry_proposal ON proposal_data_entry(proposal_id);
CREATE INDEX idx_proposal_data_entry_status ON proposal_data_entry(data_entry_status);

CREATE INDEX idx_proposal_qc_review_proposal ON proposal_qc_review(proposal_id);
CREATE INDEX idx_proposal_qc_decision ON proposal_qc_review(qr_decision) WHERE qr_decision IS NOT NULL;
CREATE INDEX idx_proposal_qc_assigned ON proposal_qc_review(qc_assigned_to) WHERE qc_assigned_to IS NOT NULL;

CREATE INDEX idx_proposal_medical_proposal ON proposal_medical(proposal_id);
CREATE INDEX idx_proposal_medical_status ON proposal_medical(medical_status) WHERE medical_status IS NOT NULL;
CREATE INDEX idx_proposal_medical_examiner ON proposal_medical(medical_examiner_id) WHERE medical_examiner_id IS NOT NULL;

CREATE INDEX idx_proposal_approval_proposal ON proposal_approval(proposal_id);
CREATE INDEX idx_proposal_approval_level ON proposal_approval(approval_level) WHERE approval_level IS NOT NULL;
CREATE INDEX idx_proposal_approval_assigned ON proposal_approval(assigned_approver_id) WHERE assigned_approver_id IS NOT NULL;

CREATE INDEX idx_proposal_issuance_proposal ON proposal_issuance(proposal_id);
CREATE INDEX idx_proposal_issuance_policy_number ON proposal_issuance(policy_number) WHERE policy_number IS NOT NULL;
CREATE INDEX idx_proposal_issuance_flc ON proposal_issuance(flc_status) WHERE flc_status = 'ACTIVE';

-- Child Table Indexes
CREATE INDEX idx_nominee_proposal ON proposal_nominee(proposal_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_medical_info_proposal ON proposal_medical_info(proposal_id);
CREATE INDEX idx_enhanced_medical_proposal ON proposal_enhanced_medical(proposal_id);
CREATE INDEX idx_agent_proposal ON proposal_agent(proposal_id);
CREATE INDEX idx_huf_proposal ON proposal_huf_member(proposal_id);
CREATE INDEX idx_doc_ref_proposal ON proposal_document_ref(proposal_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_missing_doc_proposal ON proposal_missing_documents(proposal_id);
CREATE INDEX idx_missing_doc_stage ON proposal_missing_documents(stage, status);
CREATE INDEX idx_status_history_proposal ON proposal_status_history(proposal_id);
CREATE INDEX idx_audit_proposal ON proposal_audit_log(proposal_id);

-- ============================================
-- Functions
-- ============================================

CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION validate_workflow_transition()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'UPDATE' AND OLD.status IS DISTINCT FROM NEW.status THEN
        IF NOT (
            (OLD.status = 'DRAFT' AND NEW.status IN ('INDEXED', 'CANCELLED_DEATH')) OR
            (OLD.status = 'INDEXED' AND NEW.status = 'DATA_ENTRY') OR
            (OLD.status = 'DATA_ENTRY' AND NEW.status IN ('QC_PENDING', 'DATA_ENTRY')) OR
            (OLD.status = 'QC_PENDING' AND NEW.status IN ('QC_APPROVED', 'QC_REJECTED', 'QC_RETURNED')) OR
            (OLD.status = 'QC_RETURNED' AND NEW.status = 'DATA_ENTRY') OR
            (OLD.status = 'QC_APPROVED' AND NEW.status IN ('PENDING_MEDICAL', 'APPROVAL_PENDING')) OR
            (OLD.status = 'PENDING_MEDICAL' AND NEW.status IN ('MEDICAL_APPROVED', 'MEDICAL_REJECTED')) OR
            (OLD.status = 'MEDICAL_APPROVED' AND NEW.status = 'APPROVAL_PENDING') OR
            (OLD.status = 'APPROVAL_PENDING' AND NEW.status IN ('APPROVED', 'REJECTED')) OR
            (OLD.status = 'APPROVED' AND NEW.status = 'ISSUED') OR
            (OLD.status = 'ISSUED' AND NEW.status = 'DISPATCHED') OR
            (OLD.status = 'DISPATCHED' AND NEW.status = 'FREE_LOOK_ACTIVE') OR
            (OLD.status = 'FREE_LOOK_ACTIVE' AND NEW.status IN ('ACTIVE', 'FLC_CANCELLED')) OR
            (OLD.status IN ('QC_REJECTED', 'MEDICAL_REJECTED') AND NEW.status = 'REJECTED')
        ) THEN
            RAISE EXCEPTION 'Invalid status transition from % to %. BR-POL-015', OLD.status, NEW.status;
        END IF;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE OR REPLACE FUNCTION update_current_stage()
RETURNS TRIGGER AS $$
BEGIN
    NEW.current_stage := CASE 
        WHEN NEW.status IN ('DRAFT', 'INDEXED') THEN 'INDEXING'
        WHEN NEW.status IN ('DATA_ENTRY', 'QC_RETURNED') THEN 'DATA_ENTRY'
        WHEN NEW.status IN ('QC_PENDING', 'QC_APPROVED', 'QC_REJECTED') THEN 'QC_REVIEW'
        WHEN NEW.status IN ('PENDING_MEDICAL', 'MEDICAL_APPROVED', 'MEDICAL_REJECTED') THEN 'MEDICAL'
        WHEN NEW.status IN ('APPROVAL_PENDING', 'APPROVED', 'REJECTED') THEN 'APPROVAL'
        WHEN NEW.status IN ('ISSUED', 'DISPATCHED', 'FREE_LOOK_ACTIVE', 'ACTIVE', 'FLC_CANCELLED') THEN 'ISSUANCE'
        ELSE 'INDEXING'
    END;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- ============================================
-- Triggers
-- ============================================

CREATE TRIGGER trg_proposals_updated_at BEFORE UPDATE ON proposals
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trg_proposal_workflow_transition BEFORE UPDATE ON proposals
    FOR EACH ROW EXECUTE FUNCTION validate_workflow_transition();

CREATE TRIGGER trg_proposal_current_stage BEFORE INSERT OR UPDATE ON proposals
    FOR EACH ROW EXECUTE FUNCTION update_current_stage();

-- Updated_at triggers for phase tables
CREATE TRIGGER trg_proposal_indexing_updated_at BEFORE UPDATE ON proposal_indexing
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trg_proposal_data_entry_updated_at BEFORE UPDATE ON proposal_data_entry
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trg_proposal_qc_review_updated_at BEFORE UPDATE ON proposal_qc_review
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trg_proposal_medical_updated_at BEFORE UPDATE ON proposal_medical
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trg_proposal_approval_updated_at BEFORE UPDATE ON proposal_approval
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trg_proposal_issuance_updated_at BEFORE UPDATE ON proposal_issuance
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trg_proposal_missing_docs_updated_at BEFORE UPDATE ON proposal_missing_documents
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- ============================================
-- Views
-- ============================================

CREATE VIEW v_proposal_ticket_queue AS
SELECT 
    p.proposal_id,
    p.proposal_number,
    p.policy_type,
    p.product_code,
    p.status,
    p.current_stage,
    p.created_at,
    pi.po_code,
    CASE 
        WHEN p.current_stage = 'INDEXING' THEN 'INDEXING'
        WHEN p.current_stage = 'DATA_ENTRY' THEN 'DATA_ENTRY'
        WHEN p.current_stage = 'QC_REVIEW' THEN 'QC_REVIEW'
        WHEN p.current_stage = 'MEDICAL' THEN 'MEDICAL_PENDING'
        WHEN p.current_stage = 'APPROVAL' THEN 'APPROVAL'
        ELSE 'OTHER'
    END AS request_queue,
    pm.medical_status,
    pa.agent_name,
    pa.agent_id,
    qr.qr_decision,
    ap.approval_level
FROM proposals p
LEFT JOIN proposal_indexing pi ON p.proposal_id = pi.proposal_id
LEFT JOIN proposal_agent pa ON p.proposal_id = pa.proposal_id
LEFT JOIN proposal_medical pm ON p.proposal_id = pm.proposal_id
LEFT JOIN proposal_qc_review qr ON p.proposal_id = qr.proposal_id
LEFT JOIN proposal_approval ap ON p.proposal_id = ap.proposal_id
WHERE p.deleted_at IS NULL
    AND p.status NOT IN ('ACTIVE', 'FLC_CANCELLED', 'REJECTED', 'CANCELLED_DEATH');

COMMENT ON VIEW v_proposal_ticket_queue IS 'Unified ticket queue view';

CREATE VIEW v_proposal_missing_docs_summary AS
SELECT 
    p.proposal_id,
    p.proposal_number,
    p.status,
    COUNT(*) as total_missing_docs,
    COUNT(*) FILTER (WHERE md.status = 'PENDING') as pending_docs,
    COUNT(*) FILTER (WHERE md.status = 'UPLOADED') as uploaded_docs,
    COUNT(*) FILTER (WHERE md.status = 'WAIVED') as waived_docs,
    MAX(md.noted_at) as last_noted_at
FROM proposals p
JOIN proposal_missing_documents md ON p.proposal_id = md.proposal_id
WHERE p.deleted_at IS NULL
GROUP BY p.proposal_id, p.proposal_number, p.status;

COMMENT ON VIEW v_proposal_missing_docs_summary IS 'Summary of missing documents per proposal';

-- ============================================
-- Seed Data
-- ============================================

INSERT INTO free_look_config (channel, period_days, start_date_rule) VALUES
('DIRECT', 15, 'DISPATCH_DATE'),
('AGENCY', 15, 'DISPATCH_DATE'),
('WEB', 15, 'EMAIL_SENT_DATE'),
('MOBILE', 15, 'EMAIL_SENT_DATE'),
('POS', 15, 'DISPATCH_DATE'),
('CSC', 15, 'DISPATCH_DATE');

INSERT INTO approval_routing_config (sa_min, sa_max, approver_level, approver_role) VALUES
(0, 500000, 1, 'APPROVER_LEVEL_1'),
(500000, 2000000, 2, 'APPROVER_LEVEL_2'),
(2000000, 999999999, 3, 'APPROVER_LEVEL_3');

INSERT INTO policy_number_sequence (product_type, series_prefix, next_value, format_pattern) VALUES
('PLI', 'PLI', 1, '{prefix}-{year}-{value:06d}'),
('RPLI', 'RPLI', 1, '{prefix}-{year}-{value:06d}');

-- ============================================
-- End of Schema
-- ============================================
