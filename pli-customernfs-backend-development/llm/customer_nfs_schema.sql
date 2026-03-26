-- ============================================================================
-- Customer NFS (Non-Financial Service) - PostgreSQL 16 Database Schema
-- Module: Customer NFS Service
-- Version: 1.0.0
-- Date: March 2026
-- Description: Database schema for customer non-financial service requests
--              including address change and name change workflows
-- ============================================================================

-- ============================================================================
-- SECTION 1: EXTENSIONS
-- ============================================================================

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";
CREATE EXTENSION IF NOT EXISTS "unaccent";

-- ============================================================================
-- SECTION 2: SCHEMAS
-- ============================================================================

CREATE SCHEMA IF NOT EXISTS nfs;
CREATE SCHEMA IF NOT EXISTS audit;

-- ============================================================================
-- SECTION 3: ENUMERATED TYPES
-- ============================================================================

-- Request Type Enumeration
CREATE TYPE nfs.request_type_enum AS ENUM (
    'ADDRESS_CHANGE',
    'NAME_CHANGE'
);

-- Authentication Method Enumeration
CREATE TYPE nfs.auth_method_enum AS ENUM (
    'AADHAAR',
    'MANUAL'
);

-- Service Request Status Enumeration
CREATE TYPE nfs.request_status_enum AS ENUM (
    'CREATED',
    'PENDING_DOCUMENTS',
    'PENDING_APPROVAL',
    'IN_PROGRESS',
    'COMPLETED',
    'REJECTED',
    'WITHDRAWN',
    'DOCUMENTS_EXPIRED'
);

-- Channel Enumeration
CREATE TYPE nfs.channel_enum AS ENUM (
    'Portal',
    'Mobile',
    'PostOffice',
    'CallCenter',
    'AgentPortal'
);

-- Address Update For Enumeration (Role)
CREATE TYPE nfs.address_role_enum AS ENUM (
    'INSURED',
    'PROPOSER',
    'ASSIGNEE',
    'TRUSTEE'
);

-- Address Type Enumeration
CREATE TYPE nfs.address_type_enum AS ENUM (
    'COMMUNICATION',
    'PERMANENT',
    'OFFICIAL'
);

-- Salutation Enumeration
CREATE TYPE nfs.salutation_enum AS ENUM (
    'Mr',
    'Mrs',
    'Ms',
    'Shri',
    'Smt',
    'Dr'
);

-- Document Type Enumeration
CREATE TYPE nfs.document_type_enum AS ENUM (
    'ADDRESS_PROOF',
    'RENTAL_AGREEMENT',
    'ADDRESS_CHANGE_APPLICATION_FORM',
    'GAZETTE_NOTIFICATION',
    'NEWSPAPER_NOTIFICATION',
    'NAME_CHANGE_APPLICATION_FORM'
);

-- MIME Type Enumeration
CREATE TYPE nfs.mime_type_enum AS ENUM (
    'application/pdf',
    'image/jpeg',
    'image/png'
);

-- Document Verification Status Enumeration
CREATE TYPE nfs.verification_status_enum AS ENUM (
    'PENDING',
    'VERIFIED',
    'REJECTED'
);

-- Missing Document Status Enumeration
CREATE TYPE nfs.missing_doc_status_enum AS ENUM (
    'PENDING',
    'RECEIVED',
    'EXPIRED',
    'CANCELLED'
);

-- Upload Channel Enumeration
CREATE TYPE nfs.upload_channel_enum AS ENUM (
    'Portal',
    'Mobile',
    'PostOffice',
    'SecureLink'
);

-- Audit Action Type Enumeration
CREATE TYPE nfs.audit_action_type_enum AS ENUM (
    'CREATED',
    'STATUS_CHANGE',
    'DOCUMENT_UPLOAD',
    'DOCUMENT_VERIFIED',
    'ASSIGNED',
    'APPROVED',
    'REJECTED',
    'WITHDRAWN',
    'COMMENT_ADDED',
    'MISSING_DOC_REQUESTED',
    'COMPENSATION_ADDRESS_RESTORE',
    'COMPENSATION_NAME_RESTORE'
);

-- Approval Decision Enumeration
CREATE TYPE nfs.approval_decision_enum AS ENUM (
    'APPROVE',
    'REJECT',
    'SEND_BACK'
);

-- Withdrawal Type Enumeration
CREATE TYPE nfs.withdrawal_type_enum AS ENUM (
    'AUTO',
    'MANUAL'
);

-- ============================================================================
-- SECTION 4: DOMAINS
-- ============================================================================

-- Indian Pincode Domain (6 digits)
CREATE DOMAIN nfs.indian_pincode AS VARCHAR(10)
    CHECK (VALUE ~ '^\d{6}$');

-- Ticket Number Domain
CREATE DOMAIN nfs.ticket_number AS VARCHAR(30)
    CHECK (VALUE ~ '^NFS-[A-Z]{3}-\d{8}-\d{6}$');

-- Aadhaar Transaction ID Domain
CREATE DOMAIN nfs.aadhaar_txn_id AS VARCHAR(50);

-- Policy Number Domain
CREATE DOMAIN nfs.policy_number AS VARCHAR(20);

-- Customer ID Domain
CREATE DOMAIN nfs.customer_id AS UUID;

-- Office Code Domain
CREATE DOMAIN nfs.office_code AS VARCHAR(20);

-- ============================================================================
-- SECTION 5: TABLES
-- ============================================================================

-- ----------------------------------------------------------------------------
-- TABLE: nfs.service_request
-- Description: Central entity tracking every customer NFS request
-- Entity ID: E-1
-- ----------------------------------------------------------------------------
CREATE TABLE nfs.service_request (
    -- Primary Key
    request_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- Ticket Number (Human-readable ID)
    ticket_number nfs.ticket_number NOT NULL,
    
    -- Foreign Keys
    customer_id nfs.customer_id NOT NULL,
    policy_number nfs.policy_number,
    
    -- Request Classification
    request_type nfs.request_type_enum NOT NULL,
    auth_method nfs.auth_method_enum NOT NULL,
    
    -- Status Management
    status nfs.request_status_enum NOT NULL DEFAULT 'CREATED',
    previous_status nfs.request_status_enum,
    
    -- Channel Information
    channel nfs.channel_enum NOT NULL,
    office_code nfs.office_code,
    
    -- User References
    initiated_by UUID NOT NULL,
    assigned_to UUID,
    approved_by UUID,
    
    -- Timestamps
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    approval_date TIMESTAMPTZ,
    
    -- SLA Management
    sla_deadline TIMESTAMPTZ,
    sla_breached BOOLEAN NOT NULL DEFAULT FALSE,
    
    -- Rejection Details
    rejection_reason TEXT,
    
    -- Processing Flags
    partial_processing_flag BOOLEAN NOT NULL DEFAULT FALSE,
    guardian_approval_required BOOLEAN DEFAULT FALSE,
    
    -- Workflow State
    workflow_id VARCHAR(100),
    workflow_run_id VARCHAR(100),
    
    -- Audit Fields
    created_by UUID NOT NULL,
    updated_by UUID,
    deleted_at TIMESTAMPTZ,
    version INTEGER NOT NULL DEFAULT 1,
    
    -- Metadata
    metadata JSONB DEFAULT '{}',
    
    -- Full-text Search
    search_vector TSVECTOR
);

-- ----------------------------------------------------------------------------
-- TABLE: nfs.address_change_detail
-- Description: Stores old and new address data for address change requests
-- Entity ID: E-2
-- ----------------------------------------------------------------------------
CREATE TABLE nfs.address_change_detail (
    -- Primary Key
    detail_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- Foreign Key
    request_id UUID NOT NULL,
    
    -- Address Role and Type
    address_update_for nfs.address_role_enum NOT NULL,
    address_type nfs.address_type_enum NOT NULL,
    
    -- Old Address (Previous Values)
    old_address_line1 VARCHAR(200),
    old_address_line2 VARCHAR(200),
    old_village VARCHAR(100),
    old_taluka VARCHAR(100),
    old_city VARCHAR(100),
    old_district VARCHAR(100),
    old_state VARCHAR(50),
    old_pincode nfs.indian_pincode,
    
    -- New Address (Proposed Values)
    new_address_line1 VARCHAR(200) NOT NULL,
    new_address_line2 VARCHAR(200),
    new_village VARCHAR(100),
    new_taluka VARCHAR(100),
    new_city VARCHAR(100) NOT NULL,
    new_district VARCHAR(100) NOT NULL,
    new_state VARCHAR(50) NOT NULL,
    new_pincode nfs.indian_pincode NOT NULL,
    
    -- Aadhaar Transaction Reference
    aadhaar_txn_id nfs.aadhaar_txn_id,
    
    -- Address Versioning
    old_address_version INTEGER,
    new_address_version INTEGER,
    
    -- Audit Fields
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID NOT NULL,
    updated_by UUID,
    deleted_at TIMESTAMPTZ,
    version INTEGER NOT NULL DEFAULT 1,
    
    -- Metadata
    metadata JSONB DEFAULT '{}',
    
    -- Full-text Search
    search_vector TSVECTOR,
    
    -- Constraint: One detail per request
    CONSTRAINT uk_address_change_request UNIQUE (request_id)
);

-- ----------------------------------------------------------------------------
-- TABLE: nfs.name_change_detail
-- Description: Stores old and new name data for name change requests
-- Entity ID: E-3
-- ----------------------------------------------------------------------------
CREATE TABLE nfs.name_change_detail (
    -- Primary Key
    detail_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- Foreign Key
    request_id UUID NOT NULL,
    
    -- Old Name (Previous Values)
    old_salutation nfs.salutation_enum,
    old_first_name VARCHAR(100),
    old_middle_name VARCHAR(100),
    old_last_name VARCHAR(100),
    
    -- New Name (Proposed Values)
    new_salutation nfs.salutation_enum NOT NULL,
    new_first_name VARCHAR(100) NOT NULL,
    new_middle_name VARCHAR(100),
    new_last_name VARCHAR(100) NOT NULL,
    
    -- Cross-Policy Impact
    policies_affected INTEGER DEFAULT 0,
    
    -- Aadhaar Transaction Reference
    aadhaar_txn_id nfs.aadhaar_txn_id,
    
    -- Name Versioning
    old_name_version INTEGER,
    new_name_version INTEGER,
    
    -- Audit Fields
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID NOT NULL,
    updated_by UUID,
    deleted_at TIMESTAMPTZ,
    version INTEGER NOT NULL DEFAULT 1,
    
    -- Metadata
    metadata JSONB DEFAULT '{}',
    
    -- Full-text Search
    search_vector TSVECTOR,
    
    -- Constraint: One detail per request
    CONSTRAINT uk_name_change_request UNIQUE (request_id)
);

-- ----------------------------------------------------------------------------
-- TABLE: nfs.document_upload
-- Description: Documents uploaded as part of NFS request processing
-- Entity ID: E-4
-- ----------------------------------------------------------------------------
CREATE TABLE nfs.document_upload (
    -- Primary Key
    document_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- Foreign Key
    request_id UUID NOT NULL,
    
    -- Document Classification
    document_type nfs.document_type_enum NOT NULL,
    
    -- File Information
    file_name VARCHAR(200) NOT NULL,
    file_url VARCHAR(500) NOT NULL,
    file_size_bytes BIGINT NOT NULL,
    mime_type nfs.mime_type_enum NOT NULL,
    
    -- Upload Information
    uploaded_by UUID NOT NULL,
    upload_channel nfs.upload_channel_enum NOT NULL,
    uploaded_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- Verification Status
    verification_status nfs.verification_status_enum DEFAULT 'PENDING',
    verified_by UUID,
    verified_at TIMESTAMPTZ,
    rejection_reason TEXT,
    
    -- DMS Reference
    dms_document_id VARCHAR(100),
    dms_storage_path VARCHAR(500),
    
    -- Virus Scan Status
    virus_scan_status VARCHAR(20) DEFAULT 'PENDING',
    virus_scanned_at TIMESTAMPTZ,
    
    -- Audit Fields
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID NOT NULL,
    updated_by UUID,
    deleted_at TIMESTAMPTZ,
    version INTEGER NOT NULL DEFAULT 1,
    
    -- Metadata
    metadata JSONB DEFAULT '{}',
    
    -- Full-text Search
    search_vector TSVECTOR
);

-- ----------------------------------------------------------------------------
-- TABLE: nfs.missing_document_request
-- Description: Tracks CPC-initiated requests for missing documents
-- Entity ID: E-5
-- ----------------------------------------------------------------------------
CREATE TABLE nfs.missing_document_request (
    -- Primary Key
    missing_doc_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- Foreign Key
    request_id UUID NOT NULL,
    
    -- Document Type Requested
    document_type nfs.document_type_enum NOT NULL,
    
    -- Secure Link Information
    secure_link_url VARCHAR(500),
    secure_link_token VARCHAR(100),
    link_expiry TIMESTAMPTZ,
    link_generation_count INTEGER NOT NULL DEFAULT 1,
    
    -- Status
    status nfs.missing_doc_status_enum NOT NULL DEFAULT 'PENDING',
    
    -- Receipt Information
    received_at TIMESTAMPTZ,
    received_via nfs.upload_channel_enum,
    received_document_id UUID,
    
    -- Reminder Tracking
    reminder_count INTEGER NOT NULL DEFAULT 0,
    last_reminder_at TIMESTAMPTZ,
    
    -- CPC User Reference
    requested_by UUID NOT NULL,
    
    -- Customer Message
    message_to_customer TEXT,
    
    -- Audit Fields
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID NOT NULL,
    updated_by UUID,
    deleted_at TIMESTAMPTZ,
    version INTEGER NOT NULL DEFAULT 1,
    
    -- Metadata
    metadata JSONB DEFAULT '{}'
);

-- ----------------------------------------------------------------------------
-- TABLE: nfs.audit_log
-- Description: Immutable audit trail for all NFS operations
-- Entity ID: E-6
-- Partitioned by year for performance
-- ----------------------------------------------------------------------------
CREATE TABLE nfs.audit_log (
    -- Primary Key
    audit_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- Foreign Key
    request_id UUID NOT NULL,
    
    -- Action Information
    action_type nfs.audit_action_type_enum NOT NULL,
    
    -- State Snapshots
    old_value JSONB,
    new_value JSONB,
    
    -- User Information
    performed_by UUID NOT NULL,
    performed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- Context Information
    ip_address VARCHAR(50),
    channel nfs.channel_enum,
    office_code nfs.office_code,
    
    -- Additional Information
    remarks TEXT,
    error_code VARCHAR(20),
    error_message TEXT,
    
    -- Metadata
    metadata JSONB DEFAULT '{}'
) PARTITION BY RANGE (performed_at);

-- Create partitions for audit_log (Yearly partitions)
CREATE TABLE nfs.audit_log_2026 PARTITION OF nfs.audit_log
    FOR VALUES FROM ('2026-01-01') TO ('2027-01-01');

CREATE TABLE nfs.audit_log_2027 PARTITION OF nfs.audit_log
    FOR VALUES FROM ('2027-01-01') TO ('2028-01-01');

CREATE TABLE nfs.audit_log_2028 PARTITION OF nfs.audit_log
    FOR VALUES FROM ('2028-01-01') TO ('2029-01-01');

CREATE TABLE nfs.audit_log_2029 PARTITION OF nfs.audit_log
    FOR VALUES FROM ('2029-01-01') TO ('2030-01-01');

CREATE TABLE nfs.audit_log_2030 PARTITION OF nfs.audit_log
    FOR VALUES FROM ('2030-01-01') TO ('2031-01-01');

-- Default partition for any out-of-range data
CREATE TABLE nfs.audit_log_default PARTITION OF nfs.audit_log
    DEFAULT;

-- ----------------------------------------------------------------------------
-- TABLE: nfs.status_transition_history
-- Description: Tracks all status transitions for a service request
-- ----------------------------------------------------------------------------
CREATE TABLE nfs.status_transition_history (
    -- Primary Key
    transition_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- Foreign Key
    request_id UUID NOT NULL,
    
    -- Transition Details
    from_status nfs.request_status_enum,
    to_status nfs.request_status_enum NOT NULL,
    
    -- Transition Metadata
    transition_reason TEXT,
    transitioned_by UUID NOT NULL,
    transitioned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- Workflow Context
    workflow_signal_id VARCHAR(100),
    
    -- Metadata
    metadata JSONB DEFAULT '{}'
);

-- ----------------------------------------------------------------------------
-- TABLE: nfs.cpc_work_queue
-- Description: CPC work queue for pending requests
-- ----------------------------------------------------------------------------
CREATE TABLE nfs.cpc_work_queue (
    -- Primary Key
    queue_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- Foreign Key
    request_id UUID NOT NULL,
    
    -- Assignment
    assigned_to UUID,
    assigned_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- Priority
    priority INTEGER NOT NULL DEFAULT 5,
    
    -- SLA Information
    sla_deadline TIMESTAMPTZ,
    sla_status VARCHAR(20) NOT NULL DEFAULT 'WITHIN_SLA',
    
    -- Queue Status
    queue_status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    
    -- Pickup Information
    picked_up_by UUID,
    picked_up_at TIMESTAMPTZ,
    
    -- Audit Fields
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    
    -- Constraint: One queue entry per request
    CONSTRAINT uk_work_queue_request UNIQUE (request_id)
);

-- ----------------------------------------------------------------------------
-- TABLE: nfs.ticket_sequence
-- Description: Sequence table for ticket number generation
-- ----------------------------------------------------------------------------
CREATE TABLE nfs.ticket_sequence (
    -- Primary Key (Composite)
    sequence_date DATE NOT NULL,
    request_type nfs.request_type_enum NOT NULL,
    
    -- Sequence Value
    sequence_value INTEGER NOT NULL DEFAULT 0,
    
    -- Constraint
    CONSTRAINT pk_ticket_sequence PRIMARY KEY (sequence_date, request_type)
);

-- ----------------------------------------------------------------------------
-- TABLE: nfs.address_version_history
-- Description: Address version history for audit and rollback
-- ----------------------------------------------------------------------------
CREATE TABLE nfs.address_version_history (
    -- Primary Key
    version_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- Customer Reference
    customer_id nfs.customer_id NOT NULL,
    
    -- Request Reference
    request_id UUID,
    
    -- Address Data
    address_type nfs.address_type_enum NOT NULL,
    address_line1 VARCHAR(200) NOT NULL,
    address_line2 VARCHAR(200),
    village VARCHAR(100),
    taluka VARCHAR(100),
    city VARCHAR(100) NOT NULL,
    district VARCHAR(100) NOT NULL,
    state VARCHAR(50) NOT NULL,
    pincode nfs.indian_pincode NOT NULL,
    
    -- Version Information
    version_number INTEGER NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    
    -- Effective Period
    effective_from TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    effective_to TIMESTAMPTZ,
    
    -- Audit Fields
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID NOT NULL,
    
    -- Metadata
    metadata JSONB DEFAULT '{}'
);

-- ----------------------------------------------------------------------------
-- TABLE: nfs.name_version_history
-- Description: Name version history for audit and rollback
-- ----------------------------------------------------------------------------
CREATE TABLE nfs.name_version_history (
    -- Primary Key
    version_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- Customer Reference
    customer_id nfs.customer_id NOT NULL,
    
    -- Request Reference
    request_id UUID,
    
    -- Name Data
    salutation nfs.salutation_enum,
    first_name VARCHAR(100) NOT NULL,
    middle_name VARCHAR(100),
    last_name VARCHAR(100) NOT NULL,
    
    -- Version Information
    version_number INTEGER NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    
    -- Effective Period
    effective_from TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    effective_to TIMESTAMPTZ,
    
    -- Audit Fields
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID NOT NULL,
    
    -- Metadata
    metadata JSONB DEFAULT '{}'
);

-- ----------------------------------------------------------------------------
-- TABLE: nfs.withdrawal_request
-- Description: Tracks withdrawal requests and their approval status
-- ----------------------------------------------------------------------------
CREATE TABLE nfs.withdrawal_request (
    -- Primary Key
    withdrawal_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    
    -- Foreign Key
    request_id UUID NOT NULL,
    
    -- Withdrawal Details
    withdrawal_reason TEXT NOT NULL,
    withdrawal_type nfs.withdrawal_type_enum NOT NULL,
    
    -- Status
    status nfs.request_status_enum NOT NULL DEFAULT 'PENDING_APPROVAL',
    
    -- Approval Information
    approved_by UUID,
    approved_at TIMESTAMPTZ,
    approval_remarks TEXT,
    
    -- User References
    requested_by UUID NOT NULL,
    
    -- Audit Fields
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by UUID NOT NULL,
    updated_by UUID,
    version INTEGER NOT NULL DEFAULT 1,
    
    -- Metadata
    metadata JSONB DEFAULT '{}'
);

-- ============================================================================
-- SECTION 6: INDEXES
-- ============================================================================

-- Service Request Indexes
CREATE INDEX idx_service_request_customer_id ON nfs.service_request (customer_id);
CREATE INDEX idx_service_request_policy_number ON nfs.service_request (policy_number);
CREATE INDEX idx_service_request_ticket_number ON nfs.service_request (ticket_number);
CREATE INDEX idx_service_request_status ON nfs.service_request (status);
CREATE INDEX idx_service_request_type ON nfs.service_request (request_type);
CREATE INDEX idx_service_request_auth_method ON nfs.service_request (auth_method);
CREATE INDEX idx_service_request_channel ON nfs.service_request (channel);
CREATE INDEX idx_service_request_created_at ON nfs.service_request (created_at);
CREATE INDEX idx_service_request_assigned_to ON nfs.service_request (assigned_to);
CREATE INDEX idx_service_request_sla_deadline ON nfs.service_request (sla_deadline);
CREATE INDEX idx_service_request_workflow_id ON nfs.service_request (workflow_id);
CREATE INDEX idx_service_request_deleted_at ON nfs.service_request (deleted_at);
CREATE INDEX idx_service_request_customer_status ON nfs.service_request (customer_id, status);
CREATE INDEX idx_service_request_status_created ON nfs.service_request (status, created_at);
CREATE INDEX idx_service_request_search ON nfs.service_request USING GIN (search_vector);

-- Composite Index for Dashboard Queries
CREATE INDEX idx_service_request_dashboard ON nfs.service_request (status, request_type, created_at DESC)
    WHERE deleted_at IS NULL;

-- Partial Index for Active Requests
CREATE INDEX idx_service_request_active ON nfs.service_request (customer_id, request_type, status)
    WHERE deleted_at IS NULL AND status NOT IN ('COMPLETED', 'REJECTED', 'WITHDRAWN', 'DOCUMENTS_EXPIRED');

-- Address Change Detail Indexes
CREATE INDEX idx_address_change_request_id ON nfs.address_change_detail (request_id);
CREATE INDEX idx_address_change_pincode ON nfs.address_change_detail (new_pincode);
CREATE INDEX idx_address_change_state ON nfs.address_change_detail (new_state);
CREATE INDEX idx_address_change_search ON nfs.address_change_detail USING GIN (search_vector);

-- Name Change Detail Indexes
CREATE INDEX idx_name_change_request_id ON nfs.name_change_detail (request_id);
CREATE INDEX idx_name_change_first_name ON nfs.name_change_detail (new_first_name);
CREATE INDEX idx_name_change_last_name ON nfs.name_change_detail (new_last_name);
CREATE INDEX idx_name_change_search ON nfs.name_change_detail USING GIN (search_vector);

-- Document Upload Indexes
CREATE INDEX idx_document_upload_request_id ON nfs.document_upload (request_id);
CREATE INDEX idx_document_upload_type ON nfs.document_upload (document_type);
CREATE INDEX idx_document_upload_status ON nfs.document_upload (verification_status);
CREATE INDEX idx_document_upload_uploaded_by ON nfs.document_upload (uploaded_by);
CREATE INDEX idx_document_upload_uploaded_at ON nfs.document_upload (uploaded_at);
CREATE INDEX idx_document_upload_search ON nfs.document_upload USING GIN (search_vector);

-- Missing Document Request Indexes
CREATE INDEX idx_missing_doc_request_id ON nfs.missing_document_request (request_id);
CREATE INDEX idx_missing_doc_status ON nfs.missing_document_request (status);
CREATE INDEX idx_missing_doc_expiry ON nfs.missing_document_request (link_expiry);
CREATE INDEX idx_missing_doc_link_token ON nfs.missing_document_request (secure_link_token);

-- Audit Log Indexes
CREATE INDEX idx_audit_log_request_id ON nfs.audit_log (request_id);
CREATE INDEX idx_audit_log_performed_at ON nfs.audit_log (performed_at);
CREATE INDEX idx_audit_log_action_type ON nfs.audit_log (action_type);
CREATE INDEX idx_audit_log_performed_by ON nfs.audit_log (performed_by);

-- Status Transition History Indexes
CREATE INDEX idx_status_transition_request_id ON nfs.status_transition_history (request_id);
CREATE INDEX idx_status_transition_at ON nfs.status_transition_history (transitioned_at);

-- CPC Work Queue Indexes
CREATE INDEX idx_cpc_work_queue_assigned_to ON nfs.cpc_work_queue (assigned_to);
CREATE INDEX idx_cpc_work_queue_status ON nfs.cpc_work_queue (queue_status);
CREATE INDEX idx_cpc_work_queue_sla ON nfs.cpc_work_queue (sla_deadline, sla_status);
CREATE INDEX idx_cpc_work_queue_priority ON nfs.cpc_work_queue (priority, created_at);

-- Address Version History Indexes
CREATE INDEX idx_address_version_customer ON nfs.address_version_history (customer_id);
CREATE INDEX idx_address_version_active ON nfs.address_version_history (customer_id, is_active);
CREATE INDEX idx_address_version_request ON nfs.address_version_history (request_id);

-- Name Version History Indexes
CREATE INDEX idx_name_version_customer ON nfs.name_version_history (customer_id);
CREATE INDEX idx_name_version_active ON nfs.name_version_history (customer_id, is_active);
CREATE INDEX idx_name_version_request ON nfs.name_version_history (request_id);

-- Withdrawal Request Indexes
CREATE INDEX idx_withdrawal_request_id ON nfs.withdrawal_request (request_id);
CREATE INDEX idx_withdrawal_status ON nfs.withdrawal_request (status);

-- Trigram Indexes for Text Search
CREATE INDEX idx_address_change_address_trgm ON nfs.address_change_detail 
    USING GIN (new_address_line1 gin_trgm_ops, new_city gin_trgm_ops);

CREATE INDEX idx_name_change_name_trgm ON nfs.name_change_detail 
    USING GIN (new_first_name gin_trgm_ops, new_last_name gin_trgm_ops);

-- ============================================================================
-- SECTION 7: FOREIGN KEY CONSTRAINTS
-- ============================================================================

-- Address Change Detail Foreign Keys
ALTER TABLE nfs.address_change_detail
    ADD CONSTRAINT fk_address_change_request
    FOREIGN KEY (request_id) REFERENCES nfs.service_request(request_id)
    ON DELETE RESTRICT ON UPDATE CASCADE;

-- Name Change Detail Foreign Keys
ALTER TABLE nfs.name_change_detail
    ADD CONSTRAINT fk_name_change_request
    FOREIGN KEY (request_id) REFERENCES nfs.service_request(request_id)
    ON DELETE RESTRICT ON UPDATE CASCADE;

-- Document Upload Foreign Keys
ALTER TABLE nfs.document_upload
    ADD CONSTRAINT fk_document_upload_request
    FOREIGN KEY (request_id) REFERENCES nfs.service_request(request_id)
    ON DELETE RESTRICT ON UPDATE CASCADE;

-- Missing Document Request Foreign Keys
ALTER TABLE nfs.missing_document_request
    ADD CONSTRAINT fk_missing_doc_request
    FOREIGN KEY (request_id) REFERENCES nfs.service_request(request_id)
    ON DELETE RESTRICT ON UPDATE CASCADE;

-- Audit Log Foreign Keys
ALTER TABLE nfs.audit_log
    ADD CONSTRAINT fk_audit_log_request
    FOREIGN KEY (request_id) REFERENCES nfs.service_request(request_id)
    ON DELETE RESTRICT ON UPDATE CASCADE;

-- Status Transition History Foreign Keys
ALTER TABLE nfs.status_transition_history
    ADD CONSTRAINT fk_status_transition_request
    FOREIGN KEY (request_id) REFERENCES nfs.service_request(request_id)
    ON DELETE RESTRICT ON UPDATE CASCADE;

-- CPC Work Queue Foreign Keys
ALTER TABLE nfs.cpc_work_queue
    ADD CONSTRAINT fk_work_queue_request
    FOREIGN KEY (request_id) REFERENCES nfs.service_request(request_id)
    ON DELETE RESTRICT ON UPDATE CASCADE;

-- Withdrawal Request Foreign Keys
ALTER TABLE nfs.withdrawal_request
    ADD CONSTRAINT fk_withdrawal_request
    FOREIGN KEY (request_id) REFERENCES nfs.service_request(request_id)
    ON DELETE RESTRICT ON UPDATE CASCADE;

-- ============================================================================
-- SECTION 8: CHECK CONSTRAINTS
-- ============================================================================

-- Service Request Constraints

-- Constraint: Rejection reason required when status is REJECTED
ALTER TABLE nfs.service_request
    ADD CONSTRAINT chk_rejection_reason_required
    CHECK (
        (status != 'REJECTED') OR 
        (status = 'REJECTED' AND rejection_reason IS NOT NULL AND rejection_reason != '')
    );

-- Constraint: Completed at required when status is COMPLETED
ALTER TABLE nfs.service_request
    ADD CONSTRAINT chk_completed_at_required
    CHECK (
        (status != 'COMPLETED') OR 
        (status = 'COMPLETED' AND completed_at IS NOT NULL)
    );

-- Constraint: Approval date required when approved_by is set
ALTER TABLE nfs.service_request
    ADD CONSTRAINT chk_approval_date_required
    CHECK (
        (approved_by IS NULL) OR 
        (approved_by IS NOT NULL AND approval_date IS NOT NULL)
    );

-- Constraint: Office code required for PostOffice channel
ALTER TABLE nfs.service_request
    ADD CONSTRAINT chk_office_code_for_postoffice
    CHECK (
        (channel != 'PostOffice') OR 
        (channel = 'PostOffice' AND office_code IS NOT NULL)
    );

-- Constraint: SLA deadline required for MANUAL auth method
ALTER TABLE nfs.service_request
    ADD CONSTRAINT chk_sla_for_manual
    CHECK (
        (auth_method != 'MANUAL') OR 
        (auth_method = 'MANUAL' AND sla_deadline IS NOT NULL)
    );

-- Address Change Detail Constraints

-- Constraint: New address line1 not empty
ALTER TABLE nfs.address_change_detail
    ADD CONSTRAINT chk_address_line1_not_empty
    CHECK (LENGTH(TRIM(new_address_line1)) > 0);

-- Constraint: New city not empty
ALTER TABLE nfs.address_change_detail
    ADD CONSTRAINT chk_city_not_empty
    CHECK (LENGTH(TRIM(new_city)) > 0);

-- Constraint: New district not empty
ALTER TABLE nfs.address_change_detail
    ADD CONSTRAINT chk_district_not_empty
    CHECK (LENGTH(TRIM(new_district)) > 0);

-- Constraint: New state not empty
ALTER TABLE nfs.address_change_detail
    ADD CONSTRAINT chk_state_not_empty
    CHECK (LENGTH(TRIM(new_state)) > 0);

-- Name Change Detail Constraints

-- Constraint: New first name not empty and valid format
ALTER TABLE nfs.name_change_detail
    ADD CONSTRAINT chk_first_name_valid
    CHECK (
        LENGTH(TRIM(new_first_name)) >= 2 AND 
        LENGTH(TRIM(new_first_name)) <= 100 AND
        new_first_name ~ '^[a-zA-Z\s]+$'
    );

-- Constraint: New last name not empty and valid format
ALTER TABLE nfs.name_change_detail
    ADD CONSTRAINT chk_last_name_valid
    CHECK (
        LENGTH(TRIM(new_last_name)) >= 2 AND 
        LENGTH(TRIM(new_last_name)) <= 100 AND
        new_last_name ~ '^[a-zA-Z\s]+$'
    );

-- Constraint: Middle name valid format if provided
ALTER TABLE nfs.name_change_detail
    ADD CONSTRAINT chk_middle_name_valid
    CHECK (
        new_middle_name IS NULL OR 
        (LENGTH(TRIM(new_middle_name)) >= 2 AND 
         LENGTH(TRIM(new_middle_name)) <= 100 AND
         new_middle_name ~ '^[a-zA-Z\s]+$')
    );

-- Document Upload Constraints

-- Constraint: File size within limit (5 MB)
ALTER TABLE nfs.document_upload
    ADD CONSTRAINT chk_file_size_limit
    CHECK (file_size_bytes > 0 AND file_size_bytes <= 5242880);

-- Constraint: File name not empty
ALTER TABLE nfs.document_upload
    ADD CONSTRAINT chk_file_name_not_empty
    CHECK (LENGTH(TRIM(file_name)) > 0);

-- Missing Document Request Constraints

-- Constraint: Link generation count within limit
ALTER TABLE nfs.missing_document_request
    ADD CONSTRAINT chk_link_generation_limit
    CHECK (link_generation_count >= 1 AND link_generation_count <= 3);

-- Constraint: Reminder count within limit
ALTER TABLE nfs.missing_document_request
    ADD CONSTRAINT chk_reminder_limit
    CHECK (reminder_count >= 0 AND reminder_count <= 2);

-- CPC Work Queue Constraints

-- Constraint: Priority within valid range
ALTER TABLE nfs.cpc_work_queue
    ADD CONSTRAINT chk_priority_range
    CHECK (priority >= 1 AND priority <= 10);

-- Constraint: SLA status valid
ALTER TABLE nfs.cpc_work_queue
    ADD CONSTRAINT chk_sla_status
    CHECK (sla_status IN ('WITHIN_SLA', 'APPROACHING_SLA', 'BREACHED_SLA'));

-- ============================================================================
-- SECTION 9: UNIQUE CONSTRAINTS
-- ============================================================================

-- Unique ticket number
ALTER TABLE nfs.service_request
    ADD CONSTRAINT uk_ticket_number UNIQUE (ticket_number);

-- ============================================================================
-- SECTION 10: FUNCTIONS
-- ============================================================================

-- ----------------------------------------------------------------------------
-- FUNCTION: nfs.generate_ticket_number
-- Description: Generates unique ticket number per BR-NFS-011
-- Format: NFS-{TYPE}-{YYYYMMDD}-{SEQ6}
-- ----------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION nfs.generate_ticket_number(
    p_request_type nfs.request_type_enum
) RETURNS VARCHAR(30) AS $$
DECLARE
    v_type_code VARCHAR(3);
    v_date_part VARCHAR(8);
    v_sequence INTEGER;
    v_ticket_number VARCHAR(30);
BEGIN
    -- Determine type code
    CASE p_request_type
        WHEN 'ADDRESS_CHANGE' THEN v_type_code := 'ANC';
        WHEN 'NAME_CHANGE' THEN v_type_code := 'ANC';
        ELSE v_type_code := 'UNK';
    END CASE;
    
    -- Get date part
    v_date_part := TO_CHAR(CURRENT_DATE, 'YYYYMMDD');
    
    -- Get or create sequence
    INSERT INTO nfs.ticket_sequence (sequence_date, request_type, sequence_value)
    VALUES (CURRENT_DATE, p_request_type, 1)
    ON CONFLICT (sequence_date, request_type) 
    DO UPDATE SET sequence_value = ticket_sequence.sequence_value + 1
    RETURNING sequence_value INTO v_sequence;
    
    -- Build ticket number
    v_ticket_number := 'NFS-' || v_type_code || '-' || v_date_part || '-' || LPAD(v_sequence::TEXT, 6, '0');
    
    RETURN v_ticket_number;
END;
$$ LANGUAGE plpgsql;

-- ----------------------------------------------------------------------------
-- FUNCTION: nfs.validate_status_transition
-- Description: Validates status transitions per BR-NFS-012
-- ----------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION nfs.validate_status_transition(
    p_current_status nfs.request_status_enum,
    p_new_status nfs.request_status_enum
) RETURNS BOOLEAN AS $$
DECLARE
    v_valid_transitions JSONB := '{
        "CREATED": ["PENDING_DOCUMENTS", "PENDING_APPROVAL", "COMPLETED", "WITHDRAWN"],
        "PENDING_DOCUMENTS": ["PENDING_APPROVAL", "DOCUMENTS_EXPIRED", "WITHDRAWN"],
        "PENDING_APPROVAL": ["IN_PROGRESS", "WITHDRAWN"],
        "IN_PROGRESS": ["COMPLETED", "REJECTED", "PENDING_DOCUMENTS"],
        "DOCUMENTS_EXPIRED": [],
        "COMPLETED": [],
        "REJECTED": [],
        "WITHDRAWN": []
    }';
    v_allowed_statuses TEXT[];
BEGIN
    -- Get allowed transitions for current status
    SELECT ARRAY(SELECT jsonb_array_elements_text(v_valid_transitions -> p_current_status::TEXT))
    INTO v_allowed_statuses;
    
    -- Check if new status is in allowed list
    IF p_new_status::TEXT = ANY(v_allowed_statuses) THEN
        RETURN TRUE;
    END IF;
    
    RETURN FALSE;
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- ----------------------------------------------------------------------------
-- FUNCTION: nfs.check_withdrawal_eligibility
-- Description: Checks withdrawal eligibility per BR-NFS-013
-- ----------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION nfs.check_withdrawal_eligibility(
    p_request_id UUID
) RETURNS TABLE (
    eligible BOOLEAN,
    approval_type nfs.withdrawal_type_enum,
    reason TEXT
) AS $$
DECLARE
    v_status nfs.request_status_enum;
    v_partial_processing BOOLEAN;
BEGIN
    -- Get request status and partial processing flag
    SELECT sr.status, sr.partial_processing_flag
    INTO v_status, v_partial_processing
    FROM nfs.service_request sr
    WHERE sr.request_id = p_request_id;
    
    -- Check if status is eligible
    IF v_status NOT IN ('CREATED', 'PENDING_DOCUMENTS', 'PENDING_APPROVAL') THEN
        RETURN QUERY SELECT FALSE, NULL::nfs.withdrawal_type_enum, 
            'Status ' || v_status::TEXT || ' not eligible for withdrawal'::TEXT;
        RETURN;
    END IF;
    
    -- Check for partial processing
    IF v_partial_processing THEN
        RETURN QUERY SELECT TRUE, 'MANUAL'::nfs.withdrawal_type_enum, 
            'Partial processing detected - requires CPC approval'::TEXT;
    ELSE
        RETURN QUERY SELECT TRUE, 'AUTO'::nfs.withdrawal_type_enum, 
            'No interim changes - auto-approval eligible'::TEXT;
    END IF;
END;
$$ LANGUAGE plpgsql;

-- ----------------------------------------------------------------------------
-- FUNCTION: nfs.update_sla_status
-- Description: Updates SLA status based on deadline
-- ----------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION nfs.update_sla_status()
RETURNS TRIGGER AS $$
BEGIN
    -- Update SLA breach flag
    IF NEW.sla_deadline IS NOT NULL THEN
        IF NEW.status NOT IN ('COMPLETED', 'REJECTED', 'WITHDRAWN', 'DOCUMENTS_EXPIRED') THEN
            IF NOW() > NEW.sla_deadline THEN
                NEW.sla_breached := TRUE;
            END IF;
        END IF;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- ----------------------------------------------------------------------------
-- FUNCTION: nfs.update_timestamp
-- Description: Updates the updated_at timestamp on row modification
-- ----------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION nfs.update_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at := NOW();
    NEW.version := OLD.version + 1;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- ----------------------------------------------------------------------------
-- FUNCTION: nfs.update_search_vector
-- Description: Updates the search vector for full-text search
-- ----------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION nfs.update_service_request_search_vector()
RETURNS TRIGGER AS $$
BEGIN
    NEW.search_vector :=
        setweight(to_tsvector('english', COALESCE(NEW.ticket_number, '')), 'A') ||
        setweight(to_tsvector('english', COALESCE(NEW.policy_number, '')), 'B') ||
        setweight(to_tsvector('english', COALESCE(NEW.rejection_reason, '')), 'C');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- ----------------------------------------------------------------------------
-- FUNCTION: nfs.update_address_search_vector
-- Description: Updates the search vector for address change detail
-- ----------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION nfs.update_address_search_vector()
RETURNS TRIGGER AS $$
BEGIN
    NEW.search_vector :=
        setweight(to_tsvector('english', COALESCE(NEW.new_address_line1, '')), 'A') ||
        setweight(to_tsvector('english', COALESCE(NEW.new_city, '')), 'A') ||
        setweight(to_tsvector('english', COALESCE(NEW.new_district, '')), 'B') ||
        setweight(to_tsvector('english', COALESCE(NEW.new_state, '')), 'B') ||
        setweight(to_tsvector('english', COALESCE(NEW.new_pincode, '')), 'A');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- ----------------------------------------------------------------------------
-- FUNCTION: nfs.update_name_search_vector
-- Description: Updates the search vector for name change detail
-- ----------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION nfs.update_name_search_vector()
RETURNS TRIGGER AS $$
BEGIN
    NEW.search_vector :=
        setweight(to_tsvector('english', COALESCE(NEW.new_first_name, '')), 'A') ||
        setweight(to_tsvector('english', COALESCE(NEW.new_middle_name, '')), 'B') ||
        setweight(to_tsvector('english', COALESCE(NEW.new_last_name, '')), 'A');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- ----------------------------------------------------------------------------
-- FUNCTION: nfs.log_status_transition
-- Description: Logs status transitions to history table
-- ----------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION nfs.log_status_transition()
RETURNS TRIGGER AS $$
BEGIN
    IF OLD.status IS DISTINCT FROM NEW.status THEN
        INSERT INTO nfs.status_transition_history (
            request_id,
            from_status,
            to_status,
            transition_reason,
            transitioned_by,
            workflow_signal_id
        ) VALUES (
            NEW.request_id,
            OLD.status,
            NEW.status,
            NEW.rejection_reason,
            NEW.updated_by,
            NEW.workflow_run_id
        );
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- ----------------------------------------------------------------------------
-- FUNCTION: nfs.calculate_sla_deadline
-- Description: Calculates SLA deadline based on auth method and office type
-- ----------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION nfs.calculate_sla_deadline(
    p_auth_method nfs.auth_method_enum,
    p_created_at TIMESTAMPTZ,
    p_office_code nfs.office_code DEFAULT NULL
) RETURNS TIMESTAMPTZ AS $$
DECLARE
    v_sla_days INTEGER;
BEGIN
    -- Aadhaar-based requests have no SLA (immediate)
    IF p_auth_method = 'AADHAAR' THEN
        RETURN NULL;
    END IF;
    
    -- Manual requests: 15 business days for Regional, 30 for Circle HQ
    -- Simplified: using calendar days (15 or 30)
    -- TODO: Implement business day calculation
    
    v_sla_days := 15; -- Default to Regional Office SLA
    
    -- Check if Circle HQ (based on office code pattern)
    IF p_office_code IS NOT NULL AND p_office_code LIKE 'CHQ%' THEN
        v_sla_days := 30;
    END IF;
    
    RETURN p_created_at + (v_sla_days || ' days')::INTERVAL;
END;
$$ LANGUAGE plpgsql;

-- ----------------------------------------------------------------------------
-- FUNCTION: nfs.check_duplicate_request
-- Description: Checks for duplicate pending requests per VR-NFS-015
-- ----------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION nfs.check_duplicate_request(
    p_customer_id nfs.customer_id,
    p_request_type nfs.request_type_enum,
    p_exclude_request_id UUID DEFAULT NULL
) RETURNS TABLE (
    has_duplicate BOOLEAN,
    existing_ticket VARCHAR(30)
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        COUNT(*) > 0 AS has_duplicate,
        MAX(sr.ticket_number)::VARCHAR(30) AS existing_ticket
    FROM nfs.service_request sr
    WHERE sr.customer_id = p_customer_id
      AND sr.request_type = p_request_type
      AND sr.status IN ('CREATED', 'PENDING_DOCUMENTS', 'PENDING_APPROVAL', 'IN_PROGRESS')
      AND sr.deleted_at IS NULL
      AND (p_exclude_request_id IS NULL OR sr.request_id != p_exclude_request_id);
END;
$$ LANGUAGE plpgsql;

-- ----------------------------------------------------------------------------
-- FUNCTION: nfs.validate_pincode_state
-- Description: Validates pincode-state correlation per BR-NFS-006
-- ----------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION nfs.validate_pincode_state(
    p_pincode nfs.indian_pincode,
    p_state VARCHAR(50)
) RETURNS BOOLEAN AS $$
DECLARE
    v_pincode_prefix VARCHAR(2);
    v_state_prefixes JSONB := '{
        "01": ["Delhi"],
        "02": ["Haryana", "Himachal Pradesh"],
        "03": ["Punjab", "Chandigarh"],
        "04": ["Rajasthan"],
        "05": ["Uttar Pradesh", "Uttarakhand"],
        "06": ["Haryana"],
        "07": ["Bihar", "Jharkhand"],
        "08": ["West Bengal", "Sikkim"],
        "09": ["Arunachal Pradesh", "Assam", "Manipur", "Meghalaya", "Mizoram", "Nagaland", "Tripura"],
        "10": ["Tamil Nadu"],
        "11": ["Kerala"],
        "12": ["Karnataka"],
        "13": ["Andhra Pradesh"],
        "14": ["Maharashtra"],
        "15": ["Gujarat"],
        "16": ["Goa"],
        "17": ["Madhya Pradesh"],
        "18": ["Chhattisgarh"],
        "19": ["Odisha"],
        "20": ["Telangana"],
        "21": ["Jammu and Kashmir"],
        "22": ["Ladakh"],
        "23": ["Puducherry"],
        "24": ["Daman and Diu", "Dadra and Nagar Haveli"],
        "25": ["Andaman and Nicobar Islands"],
        "26": ["Lakshadweep"]
    }';
    v_valid_states TEXT[];
BEGIN
    -- Extract pincode prefix
    v_pincode_prefix := SUBSTRING(p_pincode, 1, 2);
    
    -- Check if prefix exists
    IF NOT v_state_prefixes ? v_pincode_prefix THEN
        RETURN FALSE;
    END IF;
    
    -- Get valid states for prefix
    SELECT ARRAY(SELECT jsonb_array_elements_text(v_state_prefixes -> v_pincode_prefix))
    INTO v_valid_states;
    
    -- Check if state is valid for prefix
    IF p_state = ANY(v_valid_states) THEN
        RETURN TRUE;
    END IF;
    
    RETURN FALSE;
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- ============================================================================
-- SECTION 11: TRIGGERS
-- ============================================================================

-- Trigger: Update timestamp on service_request modification
CREATE TRIGGER tr_service_request_timestamp
    BEFORE UPDATE ON nfs.service_request
    FOR EACH ROW
    EXECUTE FUNCTION nfs.update_timestamp();

-- Trigger: Update search vector on service_request
CREATE TRIGGER tr_service_request_search_vector
    BEFORE INSERT OR UPDATE ON nfs.service_request
    FOR EACH ROW
    EXECUTE FUNCTION nfs.update_service_request_search_vector();

-- Trigger: Log status transitions
CREATE TRIGGER tr_service_request_status_transition
    AFTER UPDATE ON nfs.service_request
    FOR EACH ROW
    EXECUTE FUNCTION nfs.log_status_transition();

-- Trigger: Update SLA status
CREATE TRIGGER tr_service_request_sla_status
    BEFORE INSERT OR UPDATE ON nfs.service_request
    FOR EACH ROW
    EXECUTE FUNCTION nfs.update_sla_status();

-- Trigger: Update timestamp on address_change_detail modification
CREATE TRIGGER tr_address_change_timestamp
    BEFORE UPDATE ON nfs.address_change_detail
    FOR EACH ROW
    EXECUTE FUNCTION nfs.update_timestamp();

-- Trigger: Update search vector on address_change_detail
CREATE TRIGGER tr_address_change_search_vector
    BEFORE INSERT OR UPDATE ON nfs.address_change_detail
    FOR EACH ROW
    EXECUTE FUNCTION nfs.update_address_search_vector();

-- Trigger: Update timestamp on name_change_detail modification
CREATE TRIGGER tr_name_change_timestamp
    BEFORE UPDATE ON nfs.name_change_detail
    FOR EACH ROW
    EXECUTE FUNCTION nfs.update_timestamp();

-- Trigger: Update search vector on name_change_detail
CREATE TRIGGER tr_name_change_search_vector
    BEFORE INSERT OR UPDATE ON nfs.name_change_detail
    FOR EACH ROW
    EXECUTE FUNCTION nfs.update_name_search_vector();

-- Trigger: Update timestamp on document_upload modification
CREATE TRIGGER tr_document_upload_timestamp
    BEFORE UPDATE ON nfs.document_upload
    FOR EACH ROW
    EXECUTE FUNCTION nfs.update_timestamp();

-- Trigger: Update timestamp on missing_document_request modification
CREATE TRIGGER tr_missing_document_timestamp
    BEFORE UPDATE ON nfs.missing_document_request
    FOR EACH ROW
    EXECUTE FUNCTION nfs.update_timestamp();

-- Trigger: Update timestamp on withdrawal_request modification
CREATE TRIGGER tr_withdrawal_request_timestamp
    BEFORE UPDATE ON nfs.withdrawal_request
    FOR EACH ROW
    EXECUTE FUNCTION nfs.update_timestamp();

-- Trigger: Update timestamp on cpc_work_queue modification
CREATE TRIGGER tr_cpc_work_queue_timestamp
    BEFORE UPDATE ON nfs.cpc_work_queue
    FOR EACH ROW
    EXECUTE FUNCTION nfs.update_timestamp();

-- ============================================================================
-- SECTION 12: VIEWS
-- ============================================================================

-- ----------------------------------------------------------------------------
-- VIEW: nfs.v_active_requests
-- Description: View of active (non-terminal) service requests
-- ----------------------------------------------------------------------------
CREATE OR REPLACE VIEW nfs.v_active_requests AS
SELECT 
    sr.request_id,
    sr.ticket_number,
    sr.customer_id,
    sr.policy_number,
    sr.request_type,
    sr.auth_method,
    sr.status,
    sr.channel,
    sr.office_code,
    sr.initiated_by,
    sr.assigned_to,
    sr.created_at,
    sr.sla_deadline,
    sr.sla_breached,
    CASE 
        WHEN sr.sla_deadline IS NULL THEN 'NO_SLA'
        WHEN sr.sla_deadline < NOW() THEN 'BREACHED'
        WHEN sr.sla_deadline < NOW() + INTERVAL '2 days' THEN 'APPROACHING'
        ELSE 'WITHIN_SLA'
    END AS sla_status,
    EXTRACT(DAY FROM (NOW() - sr.created_at)) AS age_days,
    sr.partial_processing_flag
FROM nfs.service_request sr
WHERE sr.deleted_at IS NULL
  AND sr.status NOT IN ('COMPLETED', 'REJECTED', 'WITHDRAWN', 'DOCUMENTS_EXPIRED');

-- ----------------------------------------------------------------------------
-- VIEW: nfs.v_pending_approval_requests
-- Description: View of requests pending CPC approval
-- ----------------------------------------------------------------------------
CREATE OR REPLACE VIEW nfs.v_pending_approval_requests AS
SELECT 
    sr.request_id,
    sr.ticket_number,
    sr.customer_id,
    sr.policy_number,
    sr.request_type,
    sr.auth_method,
    sr.status,
    sr.channel,
    sr.office_code,
    sr.assigned_to,
    sr.created_at,
    sr.sla_deadline,
    sr.sla_breached,
    CASE 
        WHEN sr.sla_deadline IS NULL THEN 'NO_SLA'
        WHEN sr.sla_deadline < NOW() THEN 'BREACHED'
        WHEN sr.sla_deadline < NOW() + INTERVAL '2 days' THEN 'APPROACHING'
        ELSE 'WITHIN_SLA'
    END AS sla_status,
    EXTRACT(DAY FROM (NOW() - sr.created_at)) AS age_days,
    wq.priority,
    wq.queue_status
FROM nfs.service_request sr
LEFT JOIN nfs.cpc_work_queue wq ON sr.request_id = wq.request_id
WHERE sr.deleted_at IS NULL
  AND sr.status IN ('PENDING_APPROVAL', 'IN_PROGRESS');

-- ----------------------------------------------------------------------------
-- VIEW: nfs.v_address_change_summary
-- Description: Summary view of address change requests
-- ----------------------------------------------------------------------------
CREATE OR REPLACE VIEW nfs.v_address_change_summary AS
SELECT 
    sr.request_id,
    sr.ticket_number,
    sr.customer_id,
    sr.policy_number,
    sr.status,
    sr.auth_method,
    sr.channel,
    sr.created_at,
    sr.completed_at,
    sr.sla_deadline,
    acd.address_type,
    acd.address_update_for,
    acd.old_city AS previous_city,
    acd.old_state AS previous_state,
    acd.new_city,
    acd.new_state,
    acd.new_pincode
FROM nfs.service_request sr
JOIN nfs.address_change_detail acd ON sr.request_id = acd.request_id
WHERE sr.request_type = 'ADDRESS_CHANGE'
  AND sr.deleted_at IS NULL;

-- ----------------------------------------------------------------------------
-- VIEW: nfs.v_name_change_summary
-- Description: Summary view of name change requests
-- ----------------------------------------------------------------------------
CREATE OR REPLACE VIEW nfs.v_name_change_summary AS
SELECT 
    sr.request_id,
    sr.ticket_number,
    sr.customer_id,
    sr.policy_number,
    sr.status,
    sr.auth_method,
    sr.channel,
    sr.created_at,
    sr.completed_at,
    sr.sla_deadline,
    ncd.old_salutation,
    ncd.old_first_name AS previous_first_name,
    ncd.old_last_name AS previous_last_name,
    ncd.new_salutation,
    ncd.new_first_name,
    ncd.new_last_name,
    ncd.policies_affected
FROM nfs.service_request sr
JOIN nfs.name_change_detail ncd ON sr.request_id = ncd.request_id
WHERE sr.request_type = 'NAME_CHANGE'
  AND sr.deleted_at IS NULL;

-- ----------------------------------------------------------------------------
-- VIEW: nfs.v_cpc_dashboard
-- Description: CPC dashboard view with aggregated statistics
-- ----------------------------------------------------------------------------
CREATE OR REPLACE VIEW nfs.v_cpc_dashboard AS
SELECT 
    sr.office_code,
    sr.request_type,
    sr.status,
    COUNT(*) AS request_count,
    COUNT(CASE WHEN sr.sla_breached THEN 1 END) AS breached_count,
    AVG(EXTRACT(DAY FROM (NOW() - sr.created_at))) AS avg_age_days,
    MIN(sr.created_at) AS oldest_request,
    MAX(sr.created_at) AS newest_request
FROM nfs.service_request sr
WHERE sr.deleted_at IS NULL
  AND sr.status NOT IN ('COMPLETED', 'REJECTED', 'WITHDRAWN', 'DOCUMENTS_EXPIRED')
GROUP BY sr.office_code, sr.request_type, sr.status;

-- ----------------------------------------------------------------------------
-- VIEW: nfs.v_request_timeline
-- Description: Combined timeline view for request tracking
-- ----------------------------------------------------------------------------
CREATE OR REPLACE VIEW nfs.v_request_timeline AS
SELECT 
    sr.request_id,
    sr.ticket_number,
    sr.customer_id,
    sr.request_type,
    sr.status,
    sr.created_at,
    sr.completed_at,
    EXTRACT(EPOCH FROM (sr.completed_at - sr.created_at)) / 60 AS completion_time_minutes,
    CASE 
        WHEN sr.auth_method = 'AADHAAR' THEN 'IMMEDIATE'
        ELSE 'MANUAL_APPROVAL'
    END AS processing_type,
    sr.sla_breached,
    sr.assigned_to,
    sr.approved_by,
    sr.rejection_reason
FROM nfs.service_request sr
WHERE sr.deleted_at IS NULL;

-- ----------------------------------------------------------------------------
-- VIEW: nfs.v_document_status
-- Description: Document verification status view
-- ----------------------------------------------------------------------------
CREATE OR REPLACE VIEW nfs.v_document_status AS
SELECT 
    sr.request_id,
    sr.ticket_number,
    sr.status AS request_status,
    du.document_id,
    du.document_type,
    du.file_name,
    du.verification_status,
    du.uploaded_at,
    du.verified_at,
    du.verified_by,
    du.rejection_reason AS document_rejection_reason
FROM nfs.service_request sr
JOIN nfs.document_upload du ON sr.request_id = du.request_id
WHERE sr.deleted_at IS NULL
  AND du.deleted_at IS NULL;

-- ----------------------------------------------------------------------------
-- VIEW: nfs.v_missing_document_tracking
-- Description: Missing document request tracking view
-- ----------------------------------------------------------------------------
CREATE OR REPLACE VIEW nfs.v_missing_document_tracking AS
SELECT 
    sr.request_id,
    sr.ticket_number,
    sr.status AS request_status,
    mdr.missing_doc_id,
    mdr.document_type,
    mdr.status AS doc_request_status,
    mdr.link_generation_count,
    mdr.link_expiry,
    mdr.reminder_count,
    mdr.created_at AS requested_at,
    mdr.received_at,
    CASE 
        WHEN mdr.status = 'PENDING' AND mdr.link_expiry < NOW() THEN 'EXPIRED'
        WHEN mdr.status = 'PENDING' AND mdr.link_expiry < NOW() + INTERVAL '2 days' THEN 'EXPIRING_SOON'
        ELSE mdr.status::TEXT
    END AS effective_status
FROM nfs.service_request sr
JOIN nfs.missing_document_request mdr ON sr.request_id = mdr.request_id
WHERE sr.deleted_at IS NULL;

-- ----------------------------------------------------------------------------
-- VIEW: nfs.v_audit_trail
-- Description: Comprehensive audit trail view
-- ----------------------------------------------------------------------------
CREATE OR REPLACE VIEW nfs.v_audit_trail AS
SELECT 
    al.audit_id,
    al.request_id,
    sr.ticket_number,
    al.action_type,
    al.old_value,
    al.new_value,
    al.performed_by,
    al.performed_at,
    al.ip_address,
    al.channel,
    al.office_code,
    al.remarks,
    al.error_code,
    al.error_message
FROM nfs.audit_log al
JOIN nfs.service_request sr ON al.request_id = sr.request_id;

-- ============================================================================
-- SECTION 13: MATERIALIZED VIEWS
-- ============================================================================

-- ----------------------------------------------------------------------------
-- MATERIALIZED VIEW: nfs.mv_daily_statistics
-- Description: Daily statistics for reporting
-- ----------------------------------------------------------------------------
CREATE MATERIALIZED VIEW nfs.mv_daily_statistics AS
SELECT 
    DATE(created_at) AS stat_date,
    request_type,
    auth_method,
    channel,
    status,
    COUNT(*) AS request_count,
    AVG(EXTRACT(EPOCH FROM (completed_at - created_at)) / 60) AS avg_completion_time_minutes,
    COUNT(CASE WHEN sla_breached THEN 1 END) AS sla_breach_count
FROM nfs.service_request
WHERE deleted_at IS NULL
GROUP BY DATE(created_at), request_type, auth_method, channel, status;

CREATE UNIQUE INDEX idx_mv_daily_statistics_date ON nfs.mv_daily_statistics (stat_date, request_type, auth_method, channel, status);

-- ----------------------------------------------------------------------------
-- MATERIALIZED VIEW: nfs.mv_cpc_performance
-- Description: CPC user performance metrics
-- ----------------------------------------------------------------------------
CREATE MATERIALIZED VIEW nfs.mv_cpc_performance AS
SELECT 
    assigned_to AS cpc_user_id,
    request_type,
    COUNT(*) AS total_assigned,
    COUNT(CASE WHEN status = 'COMPLETED' THEN 1 END) AS completed_count,
    COUNT(CASE WHEN status = 'REJECTED' THEN 1 END) AS rejected_count,
    COUNT(CASE WHEN sla_breached THEN 1 END) AS breached_count,
    AVG(EXTRACT(DAY FROM (completed_at - created_at))) AS avg_completion_days
FROM nfs.service_request
WHERE deleted_at IS NULL
  AND assigned_to IS NOT NULL
GROUP BY assigned_to, request_type;

CREATE UNIQUE INDEX idx_mv_cpc_performance_user ON nfs.mv_cpc_performance (cpc_user_id, request_type);

-- ============================================================================
-- SECTION 14: ROW LEVEL SECURITY
-- ============================================================================

-- Enable RLS on all tables
ALTER TABLE nfs.service_request ENABLE ROW LEVEL SECURITY;
ALTER TABLE nfs.address_change_detail ENABLE ROW LEVEL SECURITY;
ALTER TABLE nfs.name_change_detail ENABLE ROW LEVEL SECURITY;
ALTER TABLE nfs.document_upload ENABLE ROW LEVEL SECURITY;
ALTER TABLE nfs.missing_document_request ENABLE ROW LEVEL SECURITY;
ALTER TABLE nfs.audit_log ENABLE ROW LEVEL SECURITY;
ALTER TABLE nfs.status_transition_history ENABLE ROW LEVEL SECURITY;
ALTER TABLE nfs.cpc_work_queue ENABLE ROW LEVEL SECURITY;
ALTER TABLE nfs.withdrawal_request ENABLE ROW LEVEL SECURITY;

-- RLS Policy: Service Request - Customers can view their own requests
CREATE POLICY policy_service_request_customer_select ON nfs.service_request
    FOR SELECT
    USING (customer_id = current_setting('app.current_customer_id', TRUE)::UUID);

-- RLS Policy: Service Request - CPC users can view assigned requests
CREATE POLICY policy_service_request_cpc_select ON nfs.service_request
    FOR SELECT
    USING (assigned_to = current_setting('app.current_user_id', TRUE)::UUID);

-- RLS Policy: Service Request - Office-based access
CREATE POLICY policy_service_request_office_select ON nfs.service_request
    FOR SELECT
    USING (office_code = current_setting('app.current_office_code', TRUE));

-- RLS Policy: Audit Log - Read-only for all authenticated users
CREATE POLICY policy_audit_log_select ON nfs.audit_log
    FOR SELECT
    USING (TRUE);

-- RLS Policy: Audit Log - Insert only (no update/delete)
CREATE POLICY policy_audit_log_insert ON nfs.audit_log
    FOR INSERT
    WITH CHECK (TRUE);

-- ============================================================================
-- SECTION 15: ROLES AND PERMISSIONS
-- ============================================================================

-- Create roles
CREATE ROLE IF NOT EXISTS nfs_service_role;
CREATE ROLE IF NOT EXISTS nfs_cpc_role;
CREATE ROLE IF NOT EXISTS nfs_admin_role;
CREATE ROLE IF NOT EXISTS nfs_readonly_role;

-- Grant permissions to nfs_service_role (application service)
GRANT USAGE ON SCHEMA nfs TO nfs_service_role;
GRANT SELECT, INSERT, UPDATE ON ALL TABLES IN SCHEMA nfs TO nfs_service_role;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA nfs TO nfs_service_role;
GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA nfs TO nfs_service_role;

-- Grant permissions to nfs_cpc_role (CPC users)
GRANT USAGE ON SCHEMA nfs TO nfs_cpc_role;
GRANT SELECT ON ALL TABLES IN SCHEMA nfs TO nfs_cpc_role;
GRANT UPDATE ON nfs.service_request TO nfs_cpc_role;
GRANT UPDATE ON nfs.cpc_work_queue TO nfs_cpc_role;
GRANT INSERT ON nfs.audit_log TO nfs_cpc_role;

-- Grant permissions to nfs_admin_role (administrators)
GRANT ALL PRIVILEGES ON SCHEMA nfs TO nfs_admin_role;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA nfs TO nfs_admin_role;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA nfs TO nfs_admin_role;
GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA nfs TO nfs_admin_role;

-- Grant permissions to nfs_readonly_role (reporting)
GRANT USAGE ON SCHEMA nfs TO nfs_readonly_role;
GRANT SELECT ON ALL TABLES IN SCHEMA nfs TO nfs_readonly_role;
GRANT SELECT ON ALL SEQUENCES IN SCHEMA nfs TO nfs_readonly_role;

-- Revoke UPDATE and DELETE from audit log (INSERT-only)
REVOKE UPDATE, DELETE ON nfs.audit_log FROM nfs_service_role;
REVOKE UPDATE, DELETE ON nfs.audit_log FROM nfs_cpc_role;
REVOKE UPDATE, DELETE ON nfs.audit_log FROM nfs_admin_role;

-- ============================================================================
-- SECTION 16: COMMENTS
-- ============================================================================

-- Schema Comments
COMMENT ON SCHEMA nfs IS 'Customer NFS (Non-Financial Service) schema for address and name change workflows';

-- Table Comments
COMMENT ON TABLE nfs.service_request IS 'Central entity tracking every customer NFS request through its lifecycle. Entity E-1.';
COMMENT ON TABLE nfs.address_change_detail IS 'Stores old and new address data for address change requests. Entity E-2.';
COMMENT ON TABLE nfs.name_change_detail IS 'Stores old and new name data for name change requests. Entity E-3.';
COMMENT ON TABLE nfs.document_upload IS 'Documents uploaded as part of NFS request processing. Entity E-4.';
COMMENT ON TABLE nfs.missing_document_request IS 'Tracks CPC-initiated requests for missing documents from customer. Entity E-5.';
COMMENT ON TABLE nfs.audit_log IS 'Immutable audit trail for all NFS operations. Entity E-6. INSERT-only table.';
COMMENT ON TABLE nfs.status_transition_history IS 'Tracks all status transitions for a service request.';
COMMENT ON TABLE nfs.cpc_work_queue IS 'CPC work queue for pending requests.';
COMMENT ON TABLE nfs.ticket_sequence IS 'Sequence table for ticket number generation per BR-NFS-11.';
COMMENT ON TABLE nfs.address_version_history IS 'Address version history for audit and rollback.';
COMMENT ON TABLE nfs.name_version_history IS 'Name version history for audit and rollback.';
COMMENT ON TABLE nfs.withdrawal_request IS 'Tracks withdrawal requests and their approval status.';

-- Column Comments for service_request
COMMENT ON COLUMN nfs.service_request.request_id IS 'Primary key - UUID generated by system';
COMMENT ON COLUMN nfs.service_request.ticket_number IS 'Human-readable ticket number. Format: NFS-{TYPE}-{YYYYMMDD}-{SEQ6} per BR-NFS-011';
COMMENT ON COLUMN nfs.service_request.customer_id IS 'Foreign key to customer in Customer Core Service';
COMMENT ON COLUMN nfs.service_request.request_type IS 'NFS type: ADDRESS_CHANGE or NAME_CHANGE';
COMMENT ON COLUMN nfs.service_request.auth_method IS 'Verification method: AADHAAR (immediate) or MANUAL (approval workflow)';
COMMENT ON COLUMN nfs.service_request.status IS 'Current status per BR-NFS-012 state machine';
COMMENT ON COLUMN nfs.service_request.channel IS 'Initiation channel: Portal, Mobile, PostOffice, CallCenter, AgentPortal';
COMMENT ON COLUMN nfs.service_request.partial_processing_flag IS 'TRUE if any data update activity succeeded, preventing auto-withdrawal';
COMMENT ON COLUMN nfs.service_request.sla_deadline IS 'SLA expiry timestamp. Calculated: created_at + SLA_days';
COMMENT ON COLUMN nfs.service_request.workflow_id IS 'Temporal workflow ID for this request';
COMMENT ON COLUMN nfs.service_request.workflow_run_id IS 'Temporal workflow run ID';

-- Column Comments for address_change_detail
COMMENT ON COLUMN nfs.address_change_detail.address_update_for IS 'Role: INSURED, PROPOSER, ASSIGNEE, TRUSTEE. Only INSURED handled by Customer NFS per BR-NFS-005';
COMMENT ON COLUMN nfs.address_change_detail.address_type IS 'Address category: COMMUNICATION, PERMANENT, OFFICIAL';
COMMENT ON COLUMN nfs.address_change_detail.new_pincode IS '6-digit Indian pincode. Must match state per BR-NFS-006';

-- Column Comments for name_change_detail
COMMENT ON COLUMN nfs.name_change_detail.policies_affected IS 'Count of linked policies updated on name change completion';
COMMENT ON COLUMN nfs.name_change_detail.aadhaar_txn_id IS 'UIDAI transaction ID if auth_method=AADHAAR';

-- Column Comments for document_upload
COMMENT ON COLUMN nfs.document_upload.file_size_bytes IS 'File size in bytes. Max 5 MB per VR-NFS-009';
COMMENT ON COLUMN nfs.document_upload.mime_type IS 'File type: application/pdf, image/jpeg, image/png per VR-NFS-010';

-- Column Comments for missing_document_request
COMMENT ON COLUMN nfs.missing_document_request.link_generation_count IS 'Number of times link generated. Max 3 per BR-NFS-014';
COMMENT ON COLUMN nfs.missing_document_request.link_expiry IS 'Link expiry timestamp. Default: created_at + 7 days';

-- Column Comments for audit_log
COMMENT ON COLUMN nfs.audit_log.action_type IS 'Action performed: CREATED, STATUS_CHANGE, DOCUMENT_UPLOAD, etc.';
COMMENT ON COLUMN nfs.audit_log.old_value IS 'Previous state as JSON snapshot';
COMMENT ON COLUMN nfs.audit_log.new_value IS 'New state as JSON snapshot';

-- Function Comments
COMMENT ON FUNCTION nfs.generate_ticket_number(nfs.request_type_enum) IS 'Generates unique ticket number per BR-NFS-11. Format: NFS-{TYPE}-{YYYYMMDD}-{SEQ6}';
COMMENT ON FUNCTION nfs.validate_status_transition(nfs.request_status_enum, nfs.request_status_enum) IS 'Validates status transitions per BR-NFS-012 state machine';
COMMENT ON FUNCTION nfs.check_withdrawal_eligibility(UUID) IS 'Checks withdrawal eligibility per BR-NFS-013';
COMMENT ON FUNCTION nfs.validate_pincode_state(nfs.indian_pincode, VARCHAR) IS 'Validates pincode-state correlation per BR-NFS-006';

-- View Comments
COMMENT ON VIEW nfs.v_active_requests IS 'View of active (non-terminal) service requests';
COMMENT ON VIEW nfs.v_pending_approval_requests IS 'View of requests pending CPC approval';
COMMENT ON VIEW nfs.v_address_change_summary IS 'Summary view of address change requests';
COMMENT ON VIEW nfs.v_name_change_summary IS 'Summary view of name change requests';
COMMENT ON VIEW nfs.v_cpc_dashboard IS 'CPC dashboard view with aggregated statistics';
COMMENT ON VIEW nfs.v_audit_trail IS 'Comprehensive audit trail view';

-- ============================================================================
-- SECTION 17: REFERENCE DATA
-- ============================================================================

-- Insert initial ticket sequence for current date (if not exists)
INSERT INTO nfs.ticket_sequence (sequence_date, request_type, sequence_value)
SELECT CURRENT_DATE, 'ADDRESS_CHANGE', 0
WHERE NOT EXISTS (
    SELECT 1 FROM nfs.ticket_sequence 
    WHERE sequence_date = CURRENT_DATE AND request_type = 'ADDRESS_CHANGE'
);

INSERT INTO nfs.ticket_sequence (sequence_date, request_type, sequence_value)
SELECT CURRENT_DATE, 'NAME_CHANGE', 0
WHERE NOT EXISTS (
    SELECT 1 FROM nfs.ticket_sequence 
    WHERE sequence_date = CURRENT_DATE AND request_type = 'NAME_CHANGE'
);

-- ============================================================================
-- SECTION 18: PARTITION MANAGEMENT FUNCTIONS
-- ============================================================================

-- ----------------------------------------------------------------------------
-- FUNCTION: nfs.create_audit_partition
-- Description: Creates a new partition for the audit_log table
-- ----------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION nfs.create_audit_partition(p_year INTEGER)
RETURNS VOID AS $$
DECLARE
    v_start_date DATE;
    v_end_date DATE;
    v_partition_name TEXT;
BEGIN
    v_start_date := make_date(p_year, 1, 1);
    v_end_date := make_date(p_year + 1, 1, 1);
    v_partition_name := 'nfs.audit_log_' || p_year;
    
    EXECUTE format(
        'CREATE TABLE IF NOT EXISTS %I PARTITION OF nfs.audit_log FOR VALUES FROM (%L) TO (%L)',
        v_partition_name, v_start_date, v_end_date
    );
END;
$$ LANGUAGE plpgsql;

-- ----------------------------------------------------------------------------
-- FUNCTION: nfs.refresh_materialized_views
-- Description: Refreshes all materialized views
-- ----------------------------------------------------------------------------
CREATE OR REPLACE FUNCTION nfs.refresh_materialized_views()
RETURNS VOID AS $$
BEGIN
    REFRESH MATERIALIZED VIEW CONCURRENTLY nfs.mv_daily_statistics;
    REFRESH MATERIALIZED VIEW CONCURRENTLY nfs.mv_cpc_performance;
END;
$$ LANGUAGE plpgsql;

-- ============================================================================
-- END OF SCHEMA
-- ============================================================================
