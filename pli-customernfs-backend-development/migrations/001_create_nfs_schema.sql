-- ============================================================================
-- Migration: 001_create_nfs_schema.sql
-- Description: Customer NFS (Non-Financial Service) initial schema
-- Module: customer-nfs-service
-- Version: 1.0.0
-- Entities: service_request (E-1), address_change_detail (E-2),
--           name_change_detail (E-3), document_upload (E-4),
--           missing_document_request (E-5), audit_log (E-6),
--           status_transition_history, cpc_work_queue, ticket_sequence,
--           address_version_history, name_version_history, withdrawal_request
-- ============================================================================

-- ===========================================================================
-- SECTION 1: EXTENSIONS
-- ===========================================================================
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
CREATE EXTENSION IF NOT EXISTS "pg_trgm";
CREATE EXTENSION IF NOT EXISTS "unaccent";

-- ===========================================================================
-- SECTION 2: SCHEMA
-- ===========================================================================
CREATE SCHEMA IF NOT EXISTS nfs;

-- ===========================================================================
-- SECTION 3: ENUMERATED TYPES
-- ===========================================================================

-- Request Type (BR-NFS-001, BR-NFS-007)
CREATE TYPE nfs.request_type_enum AS ENUM (
    'ADDRESS_CHANGE',
    'NAME_CHANGE'
);

-- Authentication Method (BR-NFS-001, BR-NFS-007)
CREATE TYPE nfs.auth_method_enum AS ENUM (
    'AADHAAR',
    'MANUAL'
);

-- Service Request Status (BR-NFS-012)
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

-- Channel (FR-NFS-001)
CREATE TYPE nfs.channel_enum AS ENUM (
    'Portal',
    'Mobile',
    'PostOffice',
    'CallCenter',
    'AgentPortal'
);

-- Address Role (BR-NFS-005)
CREATE TYPE nfs.address_role_enum AS ENUM (
    'INSURED',
    'PROPOSER',
    'ASSIGNEE',
    'TRUSTEE'
);

-- Address Type (FR-NFS-001)
CREATE TYPE nfs.address_type_enum AS ENUM (
    'COMMUNICATION',
    'PERMANENT',
    'OFFICIAL'
);

-- Salutation (VR-NFS-006)
CREATE TYPE nfs.salutation_enum AS ENUM (
    'Mr',
    'Mrs',
    'Ms',
    'Shri',
    'Smt',
    'Dr'
);

-- Document Types (FR-NFS-002)
CREATE TYPE nfs.document_type_enum AS ENUM (
    'ADDRESS_PROOF',
    'RENTAL_AGREEMENT',
    'ADDRESS_CHANGE_APPLICATION_FORM',
    'GAZETTE_NOTIFICATION',
    'NEWSPAPER_NOTIFICATION',
    'NAME_CHANGE_APPLICATION_FORM'
);

-- Allowed MIME Types (VR-NFS-010)
CREATE TYPE nfs.mime_type_enum AS ENUM (
    'application/pdf',
    'image/jpeg',
    'image/png'
);

-- Document Verification Status
CREATE TYPE nfs.verification_status_enum AS ENUM (
    'PENDING',
    'VERIFIED',
    'REJECTED'
);

-- Missing Document Status (BR-NFS-014)
CREATE TYPE nfs.missing_doc_status_enum AS ENUM (
    'PENDING',
    'RECEIVED',
    'EXPIRED',
    'CANCELLED'
);

-- Upload Channel
CREATE TYPE nfs.upload_channel_enum AS ENUM (
    'Portal',
    'Mobile',
    'PostOffice',
    'SecureLink'
);

-- Audit Action Types (BR-NFS-016)
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

-- Approval Decision (FR-NFS-011)
CREATE TYPE nfs.approval_decision_enum AS ENUM (
    'APPROVE',
    'REJECT',
    'SEND_BACK'
);

-- Withdrawal Type (BR-NFS-013)
CREATE TYPE nfs.withdrawal_type_enum AS ENUM (
    'AUTO',
    'MANUAL'
);

-- ===========================================================================
-- SECTION 4: DOMAINS
-- ===========================================================================

CREATE DOMAIN nfs.indian_pincode AS VARCHAR(10)
    CHECK (VALUE ~ '^\d{6}$');

-- Ticket Number Format: NFS-{TYPE}-{YYYYMMDD}-{SEQ6} per BR-NFS-011
CREATE DOMAIN nfs.ticket_number AS VARCHAR(30)
    CHECK (VALUE ~ '^NFS-[A-Z]{3}-\d{8}-\d{6}$');

CREATE DOMAIN nfs.aadhaar_txn_id AS VARCHAR(50);
CREATE DOMAIN nfs.policy_number   AS VARCHAR(20);
CREATE DOMAIN nfs.customer_id     AS BIGINT;
CREATE DOMAIN nfs.office_code     AS VARCHAR(20);

-- ===========================================================================
-- SECTION 5: TABLES
-- ===========================================================================

-- ---------------------------------------------------------------------------
-- TABLE: nfs.service_request  (E-1)
-- Central entity for every customer NFS request lifecycle.
-- workflow_id / workflow_run_id: stored to enable HTTP → Temporal signaling.
-- BATCH NOTE: creation always batched with detail + audit_log inserts.
-- BATCH NOTE: status update batched with status_transition_history + audit_log.
-- ---------------------------------------------------------------------------
CREATE TABLE nfs.service_request (
    request_id               UUID                        PRIMARY KEY DEFAULT uuid_generate_v4(),
    ticket_number            nfs.ticket_number           NOT NULL,
    customer_id              nfs.customer_id             NOT NULL,
    policy_number            nfs.policy_number,
    request_type             nfs.request_type_enum       NOT NULL,
    auth_method              nfs.auth_method_enum        NOT NULL,
    status                   nfs.request_status_enum     NOT NULL DEFAULT 'CREATED',
    previous_status          nfs.request_status_enum,
    channel                  nfs.channel_enum            NOT NULL,
    office_code              nfs.office_code,
    initiated_by             UUID                        NOT NULL,
    assigned_to              UUID,
    approved_by              UUID,
    created_at               TIMESTAMPTZ                 NOT NULL DEFAULT NOW(),
    updated_at               TIMESTAMPTZ                 NOT NULL DEFAULT NOW(),
    completed_at             TIMESTAMPTZ,
    approval_date            TIMESTAMPTZ,
    sla_deadline             TIMESTAMPTZ,
    sla_breached             BOOLEAN                     NOT NULL DEFAULT FALSE,
    rejection_reason         TEXT,
    partial_processing_flag  BOOLEAN                     NOT NULL DEFAULT FALSE,
    guardian_approval_required BOOLEAN                   DEFAULT FALSE,
    -- Temporal workflow state (WF-NFS-001..WF-NFS-005)
    workflow_id              VARCHAR(100),
    workflow_run_id          VARCHAR(100),
    created_by               UUID                        NOT NULL,
    updated_by               UUID,
    deleted_at               TIMESTAMPTZ,
    version                  INTEGER                     NOT NULL DEFAULT 1,
    metadata                 JSONB                       DEFAULT '{}',
    search_vector            TSVECTOR
);

-- ---------------------------------------------------------------------------
-- TABLE: nfs.address_change_detail  (E-2)
-- Old + new address, one row per service_request (UNIQUE on request_id).
-- BATCH NOTE: INSERT batched with service_request and audit_log.
-- ---------------------------------------------------------------------------
CREATE TABLE nfs.address_change_detail (
    detail_id               UUID                        PRIMARY KEY DEFAULT uuid_generate_v4(),
    request_id              UUID                        NOT NULL,
    address_update_for      nfs.address_role_enum       NOT NULL,
    address_type            nfs.address_type_enum       NOT NULL,
    old_address_line1       VARCHAR(200),
    old_address_line2       VARCHAR(200),
    old_village             VARCHAR(100),
    old_taluka              VARCHAR(100),
    old_city                VARCHAR(100),
    old_district            VARCHAR(100),
    old_state               VARCHAR(50),
    old_pincode             nfs.indian_pincode,
    new_address_line1       VARCHAR(200)                NOT NULL,
    new_address_line2       VARCHAR(200),
    new_village             VARCHAR(100),
    new_taluka              VARCHAR(100),
    new_city                VARCHAR(100)                NOT NULL,
    new_district            VARCHAR(100)                NOT NULL,
    new_state               VARCHAR(50)                 NOT NULL,
    new_pincode             nfs.indian_pincode          NOT NULL,
    aadhaar_txn_id          nfs.aadhaar_txn_id,
    old_address_version     INTEGER,
    new_address_version     INTEGER,
    created_at              TIMESTAMPTZ                 NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ                 NOT NULL DEFAULT NOW(),
    created_by              UUID                        NOT NULL,
    updated_by              UUID,
    deleted_at              TIMESTAMPTZ,
    version                 INTEGER                     NOT NULL DEFAULT 1,
    metadata                JSONB                       DEFAULT '{}',
    search_vector           TSVECTOR,
    CONSTRAINT uk_address_change_request UNIQUE (request_id)
);

-- ---------------------------------------------------------------------------
-- TABLE: nfs.name_change_detail  (E-3)
-- Old + new name, one row per service_request (UNIQUE on request_id).
-- BATCH NOTE: INSERT batched with service_request and audit_log.
-- ---------------------------------------------------------------------------
CREATE TABLE nfs.name_change_detail (
    detail_id               UUID                        PRIMARY KEY DEFAULT uuid_generate_v4(),
    request_id              UUID                        NOT NULL,
    old_salutation          nfs.salutation_enum,
    old_first_name          VARCHAR(100),
    old_middle_name         VARCHAR(100),
    old_last_name           VARCHAR(100),
    new_salutation          nfs.salutation_enum         NOT NULL,
    new_first_name          VARCHAR(100)                NOT NULL,
    new_middle_name         VARCHAR(100),
    new_last_name           VARCHAR(100)                NOT NULL,
    policies_affected       INTEGER                     DEFAULT 0,
    aadhaar_txn_id          nfs.aadhaar_txn_id,
    old_name_version        INTEGER,
    new_name_version        INTEGER,
    created_at              TIMESTAMPTZ                 NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ                 NOT NULL DEFAULT NOW(),
    created_by              UUID                        NOT NULL,
    updated_by              UUID,
    deleted_at              TIMESTAMPTZ,
    version                 INTEGER                     NOT NULL DEFAULT 1,
    metadata                JSONB                       DEFAULT '{}',
    search_vector           TSVECTOR,
    CONSTRAINT uk_name_change_request UNIQUE (request_id)
);

-- ---------------------------------------------------------------------------
-- TABLE: nfs.document_upload  (E-4)
-- Documents uploaded for NFS requests. Max 5 MB, PDF/JPEG/PNG (VR-NFS-009/010).
-- BATCH NOTE: Multiple documents per request are inserted in a single batch.
-- ---------------------------------------------------------------------------
CREATE TABLE nfs.document_upload (
    document_id             UUID                        PRIMARY KEY DEFAULT uuid_generate_v4(),
    request_id              UUID                        NOT NULL,
    document_type           nfs.document_type_enum      NOT NULL,
    file_name               VARCHAR(200)                NOT NULL,
    file_url                VARCHAR(500)                NOT NULL,
    file_size_bytes         BIGINT                      NOT NULL,
    mime_type               nfs.mime_type_enum          NOT NULL,
    uploaded_by             UUID                        NOT NULL,
    upload_channel          nfs.upload_channel_enum     NOT NULL,
    uploaded_at             TIMESTAMPTZ                 NOT NULL DEFAULT NOW(),
    verification_status     nfs.verification_status_enum DEFAULT 'PENDING',
    verified_by             UUID,
    verified_at             TIMESTAMPTZ,
    rejection_reason        TEXT,
    dms_document_id         VARCHAR(100),
    dms_storage_path        VARCHAR(500),
    virus_scan_status       VARCHAR(20)                 DEFAULT 'PENDING',
    virus_scanned_at        TIMESTAMPTZ,
    created_at              TIMESTAMPTZ                 NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ                 NOT NULL DEFAULT NOW(),
    created_by              UUID                        NOT NULL,
    updated_by              UUID,
    deleted_at              TIMESTAMPTZ,
    version                 INTEGER                     NOT NULL DEFAULT 1,
    metadata                JSONB                       DEFAULT '{}',
    search_vector           TSVECTOR
);

-- ---------------------------------------------------------------------------
-- TABLE: nfs.missing_document_request  (E-5)
-- CPC-initiated requests for missing customer documents.
-- Max 3 secure link regenerations (BR-NFS-014). Link expires in 7 days.
-- ---------------------------------------------------------------------------
CREATE TABLE nfs.missing_document_request (
    missing_doc_id          UUID                        PRIMARY KEY DEFAULT uuid_generate_v4(),
    request_id              UUID                        NOT NULL,
    document_type           nfs.document_type_enum      NOT NULL,
    secure_link_url         VARCHAR(500),
    secure_link_token       VARCHAR(100),
    link_expiry             TIMESTAMPTZ,
    link_generation_count   INTEGER                     NOT NULL DEFAULT 1,
    status                  nfs.missing_doc_status_enum NOT NULL DEFAULT 'PENDING',
    received_at             TIMESTAMPTZ,
    received_via            nfs.upload_channel_enum,
    received_document_id    UUID,
    reminder_count          INTEGER                     NOT NULL DEFAULT 0,
    last_reminder_at        TIMESTAMPTZ,
    requested_by            UUID                        NOT NULL,
    message_to_customer     TEXT,
    created_at              TIMESTAMPTZ                 NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ                 NOT NULL DEFAULT NOW(),
    created_by              UUID                        NOT NULL,
    updated_by              UUID,
    deleted_at              TIMESTAMPTZ,
    version                 INTEGER                     NOT NULL DEFAULT 1,
    metadata                JSONB                       DEFAULT '{}'
);

-- ---------------------------------------------------------------------------
-- TABLE: nfs.audit_log  (E-6)
-- Immutable audit trail. INSERT-only (no UPDATE/DELETE). Partitioned by year.
-- BATCH NOTE: Every operation batches audit_log INSERT as last step in TX.
-- ---------------------------------------------------------------------------
-- CREATE TABLE nfs.audit_log (
--     audit_id                UUID                            PRIMARY KEY DEFAULT uuid_generate_v4(),
--     request_id              UUID                            NOT NULL,
--     action_type             nfs.audit_action_type_enum      NOT NULL,
--     old_value               JSONB,
--     new_value               JSONB,
--     performed_by            UUID                            NOT NULL,
--     performed_at            TIMESTAMPTZ                     NOT NULL DEFAULT NOW(),
--     ip_address              VARCHAR(50),
--     channel                 nfs.channel_enum,
--     office_code             nfs.office_code,
--     remarks                 TEXT,
--     error_code              VARCHAR(20),
--     error_message           TEXT,
--     metadata                JSONB                           DEFAULT '{}'
-- ) PARTITION BY RANGE (performed_at);
CREATE TABLE nfs.audit_log (
    audit_id                UUID                            NOT NULL DEFAULT uuid_generate_v4(),
    request_id              UUID                            NOT NULL,
    action_type             nfs.audit_action_type_enum      NOT NULL,
    old_value               JSONB,
    new_value               JSONB,
    performed_by            UUID                            NOT NULL,
    performed_at            TIMESTAMPTZ                     NOT NULL DEFAULT NOW(),
    ip_address              VARCHAR(50),
    channel                 nfs.channel_enum,
    office_code             nfs.office_code,
    remarks                 TEXT,
    error_code              VARCHAR(20),
    error_message           TEXT,
    metadata                JSONB                           DEFAULT '{}',
    PRIMARY KEY (audit_id, performed_at)    -- ✅ composite PK includes partition key
) PARTITION BY RANGE (performed_at);

CREATE TABLE nfs.audit_log_2026 PARTITION OF nfs.audit_log FOR VALUES FROM ('2026-01-01') TO ('2027-01-01');
CREATE TABLE nfs.audit_log_2027 PARTITION OF nfs.audit_log FOR VALUES FROM ('2027-01-01') TO ('2028-01-01');
CREATE TABLE nfs.audit_log_2028 PARTITION OF nfs.audit_log FOR VALUES FROM ('2028-01-01') TO ('2029-01-01');
CREATE TABLE nfs.audit_log_2029 PARTITION OF nfs.audit_log FOR VALUES FROM ('2029-01-01') TO ('2030-01-01');
CREATE TABLE nfs.audit_log_2030 PARTITION OF nfs.audit_log FOR VALUES FROM ('2030-01-01') TO ('2031-01-01');
CREATE TABLE nfs.audit_log_default PARTITION OF nfs.audit_log DEFAULT;

-- ---------------------------------------------------------------------------
-- TABLE: nfs.status_transition_history
-- Automatic log of every status change (BR-NFS-012). Driven by trigger
-- tr_service_request_status_transition in production; also inserted via
-- Go batch TX in service layer to avoid trigger latency in tests.
-- ---------------------------------------------------------------------------
CREATE TABLE nfs.status_transition_history (
    transition_id           UUID                        PRIMARY KEY DEFAULT uuid_generate_v4(),
    request_id              UUID                        NOT NULL,
    from_status             nfs.request_status_enum,
    to_status               nfs.request_status_enum     NOT NULL,
    transition_reason       TEXT,
    transitioned_by         UUID                        NOT NULL,
    transitioned_at         TIMESTAMPTZ                 NOT NULL DEFAULT NOW(),
    workflow_signal_id      VARCHAR(100),
    metadata                JSONB                       DEFAULT '{}'
);

-- ---------------------------------------------------------------------------
-- TABLE: nfs.cpc_work_queue
-- CPC manual review queue. One entry per request (UK on request_id).
-- ---------------------------------------------------------------------------
CREATE TABLE nfs.cpc_work_queue (
    queue_id                UUID                        PRIMARY KEY DEFAULT uuid_generate_v4(),
    request_id              UUID                        NOT NULL,
    assigned_to             UUID,
    assigned_at             TIMESTAMPTZ                 NOT NULL DEFAULT NOW(),
    priority                INTEGER                     NOT NULL DEFAULT 5,
    sla_deadline            TIMESTAMPTZ,
    sla_status              VARCHAR(20)                 NOT NULL DEFAULT 'WITHIN_SLA',
    queue_status            VARCHAR(20)                 NOT NULL DEFAULT 'PENDING',
    picked_up_by            UUID,
    picked_up_at            TIMESTAMPTZ,
    created_at              TIMESTAMPTZ                 NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ                 NOT NULL DEFAULT NOW(),
    CONSTRAINT uk_work_queue_request UNIQUE (request_id)
);

-- ---------------------------------------------------------------------------
-- TABLE: nfs.ticket_sequence
-- Per-date, per-type sequence counter for ticket number generation (BR-NFS-011).
-- Go layer uses SELECT FOR UPDATE + upsert to atomically increment.
-- ---------------------------------------------------------------------------
CREATE TABLE nfs.ticket_sequence (
    sequence_date           DATE                        NOT NULL,
    request_type            nfs.request_type_enum       NOT NULL,
    sequence_value          INTEGER                     NOT NULL DEFAULT 0,
    CONSTRAINT pk_ticket_sequence PRIMARY KEY (sequence_date, request_type)
);

-- ---------------------------------------------------------------------------
-- TABLE: nfs.address_version_history
-- Historical address versions per customer for rollback/audit.
-- BATCH NOTE: Insert batched with service_request status update.
-- ---------------------------------------------------------------------------
CREATE TABLE nfs.address_version_history (
    version_id              UUID                        PRIMARY KEY DEFAULT uuid_generate_v4(),
    customer_id             nfs.customer_id             NOT NULL,
    request_id              UUID,
    address_type            nfs.address_type_enum       NOT NULL,
    address_line1           VARCHAR(200)                NOT NULL,
    address_line2           VARCHAR(200),
    village                 VARCHAR(100),
    taluka                  VARCHAR(100),
    city                    VARCHAR(100)                NOT NULL,
    district                VARCHAR(100)                NOT NULL,
    state                   VARCHAR(50)                 NOT NULL,
    pincode                 nfs.indian_pincode          NOT NULL,
    version_number          INTEGER                     NOT NULL,
    is_active               BOOLEAN                     NOT NULL DEFAULT TRUE,
    effective_from          TIMESTAMPTZ                 NOT NULL DEFAULT NOW(),
    effective_to            TIMESTAMPTZ,
    created_at              TIMESTAMPTZ                 NOT NULL DEFAULT NOW(),
    created_by              UUID                        NOT NULL,
    metadata                JSONB                       DEFAULT '{}'
);

-- ---------------------------------------------------------------------------
-- TABLE: nfs.name_version_history
-- Historical name versions per customer for rollback/audit.
-- BATCH NOTE: Insert batched with service_request status update.
-- ---------------------------------------------------------------------------
CREATE TABLE nfs.name_version_history (
    version_id              UUID                        PRIMARY KEY DEFAULT uuid_generate_v4(),
    customer_id             nfs.customer_id             NOT NULL,
    request_id              UUID,
    salutation              nfs.salutation_enum,
    first_name              VARCHAR(100)                NOT NULL,
    middle_name             VARCHAR(100),
    last_name               VARCHAR(100)                NOT NULL,
    version_number          INTEGER                     NOT NULL,
    is_active               BOOLEAN                     NOT NULL DEFAULT TRUE,
    effective_from          TIMESTAMPTZ                 NOT NULL DEFAULT NOW(),
    effective_to            TIMESTAMPTZ,
    created_at              TIMESTAMPTZ                 NOT NULL DEFAULT NOW(),
    created_by              UUID                        NOT NULL,
    metadata                JSONB                       DEFAULT '{}'
);

-- ---------------------------------------------------------------------------
-- TABLE: nfs.withdrawal_request
-- Tracks withdrawal request approval status (BR-NFS-013).
-- ---------------------------------------------------------------------------
CREATE TABLE nfs.withdrawal_request (
    withdrawal_id           UUID                        PRIMARY KEY DEFAULT uuid_generate_v4(),
    request_id              UUID                        NOT NULL,
    withdrawal_reason       TEXT                        NOT NULL,
    withdrawal_type         nfs.withdrawal_type_enum    NOT NULL,
    status                  nfs.request_status_enum     NOT NULL DEFAULT 'PENDING_APPROVAL',
    approved_by             UUID,
    approved_at             TIMESTAMPTZ,
    approval_remarks        TEXT,
    requested_by            UUID                        NOT NULL,
    created_at              TIMESTAMPTZ                 NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMPTZ                 NOT NULL DEFAULT NOW(),
    created_by              UUID                        NOT NULL,
    updated_by              UUID,
    version                 INTEGER                     NOT NULL DEFAULT 1,
    metadata                JSONB                       DEFAULT '{}'
);

-- ===========================================================================
-- SECTION 6: INDEXES
-- ===========================================================================

-- service_request
CREATE INDEX idx_service_request_customer_id      ON nfs.service_request (customer_id);
CREATE INDEX idx_service_request_policy_number     ON nfs.service_request (policy_number);
CREATE INDEX idx_service_request_ticket_number     ON nfs.service_request (ticket_number);
CREATE INDEX idx_service_request_status            ON nfs.service_request (status);
CREATE INDEX idx_service_request_type              ON nfs.service_request (request_type);
CREATE INDEX idx_service_request_auth_method       ON nfs.service_request (auth_method);
CREATE INDEX idx_service_request_channel           ON nfs.service_request (channel);
CREATE INDEX idx_service_request_created_at        ON nfs.service_request (created_at);
CREATE INDEX idx_service_request_assigned_to       ON nfs.service_request (assigned_to);
CREATE INDEX idx_service_request_sla_deadline      ON nfs.service_request (sla_deadline);
CREATE INDEX idx_service_request_workflow_id       ON nfs.service_request (workflow_id);
CREATE INDEX idx_service_request_deleted_at        ON nfs.service_request (deleted_at);
CREATE INDEX idx_service_request_customer_status   ON nfs.service_request (customer_id, status);
CREATE INDEX idx_service_request_status_created    ON nfs.service_request (status, created_at);
CREATE INDEX idx_service_request_search            ON nfs.service_request USING GIN (search_vector);
CREATE INDEX idx_service_request_dashboard         ON nfs.service_request (status, request_type, created_at DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_service_request_active            ON nfs.service_request (customer_id, request_type, status)
    WHERE deleted_at IS NULL AND status NOT IN ('COMPLETED', 'REJECTED', 'WITHDRAWN', 'DOCUMENTS_EXPIRED');

-- address_change_detail
CREATE INDEX idx_address_change_request_id ON nfs.address_change_detail (request_id);
CREATE INDEX idx_address_change_pincode    ON nfs.address_change_detail (new_pincode);
CREATE INDEX idx_address_change_state      ON nfs.address_change_detail (new_state);
CREATE INDEX idx_address_change_search     ON nfs.address_change_detail USING GIN (search_vector);
CREATE INDEX idx_address_change_address_trgm ON nfs.address_change_detail
    USING GIN (new_address_line1 gin_trgm_ops, new_city gin_trgm_ops);

-- name_change_detail
CREATE INDEX idx_name_change_request_id  ON nfs.name_change_detail (request_id);
CREATE INDEX idx_name_change_first_name  ON nfs.name_change_detail (new_first_name);
CREATE INDEX idx_name_change_last_name   ON nfs.name_change_detail (new_last_name);
CREATE INDEX idx_name_change_search      ON nfs.name_change_detail USING GIN (search_vector);
CREATE INDEX idx_name_change_name_trgm   ON nfs.name_change_detail
    USING GIN (new_first_name gin_trgm_ops, new_last_name gin_trgm_ops);

-- document_upload
CREATE INDEX idx_document_upload_request_id  ON nfs.document_upload (request_id);
CREATE INDEX idx_document_upload_type        ON nfs.document_upload (document_type);
CREATE INDEX idx_document_upload_status      ON nfs.document_upload (verification_status);
CREATE INDEX idx_document_upload_uploaded_by ON nfs.document_upload (uploaded_by);
CREATE INDEX idx_document_upload_uploaded_at ON nfs.document_upload (uploaded_at);

-- missing_document_request
CREATE INDEX idx_missing_doc_request_id  ON nfs.missing_document_request (request_id);
CREATE INDEX idx_missing_doc_status      ON nfs.missing_document_request (status);
CREATE INDEX idx_missing_doc_expiry      ON nfs.missing_document_request (link_expiry);
CREATE INDEX idx_missing_doc_link_token  ON nfs.missing_document_request (secure_link_token);

-- audit_log
CREATE INDEX idx_audit_log_request_id   ON nfs.audit_log (request_id);
CREATE INDEX idx_audit_log_performed_at ON nfs.audit_log (performed_at);
CREATE INDEX idx_audit_log_action_type  ON nfs.audit_log (action_type);
CREATE INDEX idx_audit_log_performed_by ON nfs.audit_log (performed_by);

-- status_transition_history
CREATE INDEX idx_status_transition_request_id ON nfs.status_transition_history (request_id);
CREATE INDEX idx_status_transition_at         ON nfs.status_transition_history (transitioned_at);

-- cpc_work_queue
CREATE INDEX idx_cpc_work_queue_assigned_to ON nfs.cpc_work_queue (assigned_to);
CREATE INDEX idx_cpc_work_queue_status      ON nfs.cpc_work_queue (queue_status);
CREATE INDEX idx_cpc_work_queue_sla         ON nfs.cpc_work_queue (sla_deadline, sla_status);
CREATE INDEX idx_cpc_work_queue_priority    ON nfs.cpc_work_queue (priority, created_at);

-- address_version_history / name_version_history
CREATE INDEX idx_address_version_customer ON nfs.address_version_history (customer_id);
CREATE INDEX idx_address_version_active   ON nfs.address_version_history (customer_id, is_active);
CREATE INDEX idx_address_version_request  ON nfs.address_version_history (request_id);
CREATE INDEX idx_name_version_customer    ON nfs.name_version_history (customer_id);
CREATE INDEX idx_name_version_active      ON nfs.name_version_history (customer_id, is_active);
CREATE INDEX idx_name_version_request     ON nfs.name_version_history (request_id);

-- withdrawal_request
CREATE INDEX idx_withdrawal_request_id ON nfs.withdrawal_request (request_id);
CREATE INDEX idx_withdrawal_status     ON nfs.withdrawal_request (status);

-- ===========================================================================
-- SECTION 7: FOREIGN KEY CONSTRAINTS
-- ===========================================================================

ALTER TABLE nfs.address_change_detail
    ADD CONSTRAINT fk_address_change_request
    FOREIGN KEY (request_id) REFERENCES nfs.service_request(request_id)
    ON DELETE RESTRICT ON UPDATE CASCADE;

ALTER TABLE nfs.name_change_detail
    ADD CONSTRAINT fk_name_change_request
    FOREIGN KEY (request_id) REFERENCES nfs.service_request(request_id)
    ON DELETE RESTRICT ON UPDATE CASCADE;

ALTER TABLE nfs.document_upload
    ADD CONSTRAINT fk_document_upload_request
    FOREIGN KEY (request_id) REFERENCES nfs.service_request(request_id)
    ON DELETE RESTRICT ON UPDATE CASCADE;

ALTER TABLE nfs.missing_document_request
    ADD CONSTRAINT fk_missing_doc_request
    FOREIGN KEY (request_id) REFERENCES nfs.service_request(request_id)
    ON DELETE RESTRICT ON UPDATE CASCADE;

ALTER TABLE nfs.audit_log
    ADD CONSTRAINT fk_audit_log_request
    FOREIGN KEY (request_id) REFERENCES nfs.service_request(request_id)
    ON DELETE RESTRICT ON UPDATE CASCADE;

ALTER TABLE nfs.status_transition_history
    ADD CONSTRAINT fk_status_transition_request
    FOREIGN KEY (request_id) REFERENCES nfs.service_request(request_id)
    ON DELETE RESTRICT ON UPDATE CASCADE;

ALTER TABLE nfs.cpc_work_queue
    ADD CONSTRAINT fk_work_queue_request
    FOREIGN KEY (request_id) REFERENCES nfs.service_request(request_id)
    ON DELETE RESTRICT ON UPDATE CASCADE;

ALTER TABLE nfs.withdrawal_request
    ADD CONSTRAINT fk_withdrawal_request
    FOREIGN KEY (request_id) REFERENCES nfs.service_request(request_id)
    ON DELETE RESTRICT ON UPDATE CASCADE;

-- ===========================================================================
-- SECTION 8: CHECK CONSTRAINTS
-- ===========================================================================

-- service_request constraints (BR-NFS-012, BR-NFS-013)
ALTER TABLE nfs.service_request
    ADD CONSTRAINT chk_rejection_reason_required
    CHECK ((status != 'REJECTED') OR (status = 'REJECTED' AND rejection_reason IS NOT NULL AND rejection_reason != ''));

ALTER TABLE nfs.service_request
    ADD CONSTRAINT chk_completed_at_required
    CHECK ((status != 'COMPLETED') OR (status = 'COMPLETED' AND completed_at IS NOT NULL));

ALTER TABLE nfs.service_request
    ADD CONSTRAINT chk_approval_date_required
    CHECK ((approved_by IS NULL) OR (approved_by IS NOT NULL AND approval_date IS NOT NULL));

ALTER TABLE nfs.service_request
    ADD CONSTRAINT chk_office_code_for_postoffice
    CHECK ((channel != 'PostOffice') OR (channel = 'PostOffice' AND office_code IS NOT NULL));

ALTER TABLE nfs.service_request
    ADD CONSTRAINT chk_sla_for_manual
    CHECK ((auth_method != 'MANUAL') OR (auth_method = 'MANUAL' AND sla_deadline IS NOT NULL));

-- address_change_detail constraints (VR-NFS-001..005)
ALTER TABLE nfs.address_change_detail ADD CONSTRAINT chk_address_line1_not_empty CHECK (LENGTH(TRIM(new_address_line1)) > 0);
ALTER TABLE nfs.address_change_detail ADD CONSTRAINT chk_city_not_empty          CHECK (LENGTH(TRIM(new_city)) > 0);
ALTER TABLE nfs.address_change_detail ADD CONSTRAINT chk_district_not_empty      CHECK (LENGTH(TRIM(new_district)) > 0);
ALTER TABLE nfs.address_change_detail ADD CONSTRAINT chk_state_not_empty         CHECK (LENGTH(TRIM(new_state)) > 0);

-- name_change_detail constraints (VR-NFS-006..008)
ALTER TABLE nfs.name_change_detail
    ADD CONSTRAINT chk_first_name_valid
    CHECK (LENGTH(TRIM(new_first_name)) >= 2 AND LENGTH(TRIM(new_first_name)) <= 100 AND new_first_name ~ '^[a-zA-Z\s]+$');

ALTER TABLE nfs.name_change_detail
    ADD CONSTRAINT chk_last_name_valid
    CHECK (LENGTH(TRIM(new_last_name)) >= 2 AND LENGTH(TRIM(new_last_name)) <= 100 AND new_last_name ~ '^[a-zA-Z\s]+$');

ALTER TABLE nfs.name_change_detail
    ADD CONSTRAINT chk_middle_name_valid
    CHECK (new_middle_name IS NULL OR (LENGTH(TRIM(new_middle_name)) >= 2 AND LENGTH(TRIM(new_middle_name)) <= 100 AND new_middle_name ~ '^[a-zA-Z\s]+$'));

-- document_upload constraints (VR-NFS-009, VR-NFS-010)
ALTER TABLE nfs.document_upload ADD CONSTRAINT chk_file_size_limit      CHECK (file_size_bytes > 0 AND file_size_bytes <= 5242880);
ALTER TABLE nfs.document_upload ADD CONSTRAINT chk_file_name_not_empty  CHECK (LENGTH(TRIM(file_name)) > 0);

-- missing_document_request constraints (BR-NFS-014)
ALTER TABLE nfs.missing_document_request ADD CONSTRAINT chk_link_generation_limit CHECK (link_generation_count >= 1 AND link_generation_count <= 3);
ALTER TABLE nfs.missing_document_request ADD CONSTRAINT chk_reminder_limit        CHECK (reminder_count >= 0 AND reminder_count <= 2);

-- cpc_work_queue constraints
ALTER TABLE nfs.cpc_work_queue ADD CONSTRAINT chk_priority_range CHECK (priority >= 1 AND priority <= 10);
ALTER TABLE nfs.cpc_work_queue ADD CONSTRAINT chk_sla_status     CHECK (sla_status IN ('WITHIN_SLA', 'APPROACHING_SLA', 'BREACHED_SLA'));

-- ===========================================================================
-- SECTION 9: UNIQUE CONSTRAINTS
-- ===========================================================================
ALTER TABLE nfs.service_request ADD CONSTRAINT uk_ticket_number UNIQUE (ticket_number);

-- ===========================================================================
-- SECTION 10: FUNCTIONS
-- ===========================================================================

-- generate_ticket_number: NFS-{TYPE}-{YYYYMMDD}-{SEQ6} per BR-NFS-011
CREATE OR REPLACE FUNCTION nfs.generate_ticket_number(
    p_request_type nfs.request_type_enum
) RETURNS VARCHAR(30) AS $$
DECLARE
    v_type_code    VARCHAR(3);
    v_date_part    VARCHAR(8);
    v_sequence     INTEGER;
    v_ticket       VARCHAR(30);
BEGIN
    v_type_code := CASE p_request_type WHEN 'ADDRESS_CHANGE' THEN 'ANC' WHEN 'NAME_CHANGE' THEN 'NMC' ELSE 'UNK' END;
    v_date_part := TO_CHAR(CURRENT_DATE, 'YYYYMMDD');
    INSERT INTO nfs.ticket_sequence (sequence_date, request_type, sequence_value)
    VALUES (CURRENT_DATE, p_request_type, 1)
    ON CONFLICT (sequence_date, request_type)
    DO UPDATE SET sequence_value = ticket_sequence.sequence_value + 1
    RETURNING sequence_value INTO v_sequence;
    v_ticket := 'NFS-' || v_type_code || '-' || v_date_part || '-' || LPAD(v_sequence::TEXT, 6, '0');
    RETURN v_ticket;
END;
$$ LANGUAGE plpgsql;

-- validate_status_transition per BR-NFS-012
CREATE OR REPLACE FUNCTION nfs.validate_status_transition(
    p_current nfs.request_status_enum,
    p_new     nfs.request_status_enum
) RETURNS BOOLEAN AS $$
DECLARE
    v_transitions JSONB := '{
        "CREATED":            ["PENDING_DOCUMENTS","PENDING_APPROVAL","COMPLETED","WITHDRAWN"],
        "PENDING_DOCUMENTS":  ["PENDING_APPROVAL","DOCUMENTS_EXPIRED","WITHDRAWN"],
        "PENDING_APPROVAL":   ["IN_PROGRESS","WITHDRAWN"],
        "IN_PROGRESS":        ["COMPLETED","REJECTED","PENDING_DOCUMENTS"],
        "DOCUMENTS_EXPIRED":  [],
        "COMPLETED":          [],
        "REJECTED":           [],
        "WITHDRAWN":          []
    }';
    v_allowed TEXT[];
BEGIN
    SELECT ARRAY(SELECT jsonb_array_elements_text(v_transitions -> p_current::TEXT)) INTO v_allowed;
    RETURN p_new::TEXT = ANY(v_allowed);
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- check_duplicate_request per VR-NFS-015
CREATE OR REPLACE FUNCTION nfs.check_duplicate_request(
    p_customer_id    nfs.customer_id,
    p_request_type   nfs.request_type_enum,
    p_exclude_id     UUID DEFAULT NULL
) RETURNS TABLE (has_duplicate BOOLEAN, existing_ticket VARCHAR(30)) AS $$
BEGIN
    RETURN QUERY
    SELECT COUNT(*) > 0,
           MAX(sr.ticket_number)::VARCHAR(30)
    FROM nfs.service_request sr
    WHERE sr.customer_id = p_customer_id
      AND sr.request_type = p_request_type
      AND sr.status IN ('CREATED','PENDING_DOCUMENTS','PENDING_APPROVAL','IN_PROGRESS')
      AND sr.deleted_at IS NULL
      AND (p_exclude_id IS NULL OR sr.request_id != p_exclude_id);
END;
$$ LANGUAGE plpgsql;

-- check_withdrawal_eligibility per BR-NFS-013
CREATE OR REPLACE FUNCTION nfs.check_withdrawal_eligibility(
    p_request_id UUID
) RETURNS TABLE (eligible BOOLEAN, approval_type nfs.withdrawal_type_enum, reason TEXT) AS $$
DECLARE
    v_status   nfs.request_status_enum;
    v_partial  BOOLEAN;
BEGIN
    SELECT sr.status, sr.partial_processing_flag INTO v_status, v_partial
    FROM nfs.service_request sr WHERE sr.request_id = p_request_id;

    IF v_status NOT IN ('CREATED','PENDING_DOCUMENTS','PENDING_APPROVAL') THEN
        RETURN QUERY SELECT FALSE, NULL::nfs.withdrawal_type_enum,
            ('Status ' || v_status::TEXT || ' not eligible for withdrawal')::TEXT;
        RETURN;
    END IF;
    IF v_partial THEN
        RETURN QUERY SELECT TRUE, 'MANUAL'::nfs.withdrawal_type_enum, 'Partial processing - CPC approval required'::TEXT;
    ELSE
        RETURN QUERY SELECT TRUE, 'AUTO'::nfs.withdrawal_type_enum, 'No interim changes - auto eligible'::TEXT;
    END IF;
END;
$$ LANGUAGE plpgsql;

-- Auto-update updated_at + version on modifications
CREATE OR REPLACE FUNCTION nfs.update_timestamp() RETURNS TRIGGER AS $$
BEGIN NEW.updated_at := NOW(); NEW.version := OLD.version + 1; RETURN NEW; END;
$$ LANGUAGE plpgsql;

-- Update full-text search vector on service_request
CREATE OR REPLACE FUNCTION nfs.update_service_request_search_vector() RETURNS TRIGGER AS $$
BEGIN
    NEW.search_vector :=
        setweight(to_tsvector('english', COALESCE(NEW.ticket_number, '')), 'A') ||
        setweight(to_tsvector('english', COALESCE(NEW.policy_number, '')), 'B') ||
        setweight(to_tsvector('english', COALESCE(NEW.rejection_reason, '')), 'C');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Update full-text search vector on address_change_detail
CREATE OR REPLACE FUNCTION nfs.update_address_search_vector() RETURNS TRIGGER AS $$
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

-- Update full-text search vector on name_change_detail
CREATE OR REPLACE FUNCTION nfs.update_name_search_vector() RETURNS TRIGGER AS $$
BEGIN
    NEW.search_vector :=
        setweight(to_tsvector('english', COALESCE(NEW.new_first_name, '')), 'A') ||
        setweight(to_tsvector('english', COALESCE(NEW.new_middle_name, '')), 'B') ||
        setweight(to_tsvector('english', COALESCE(NEW.new_last_name, '')), 'A');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Calculate SLA deadline (FR-NFS-003, config-driven: 15/30 days)
CREATE OR REPLACE FUNCTION nfs.calculate_sla_deadline(
    p_auth_method  nfs.auth_method_enum,
    p_created_at   TIMESTAMPTZ,
    p_office_code  nfs.office_code DEFAULT NULL
) RETURNS TIMESTAMPTZ AS $$
DECLARE v_days INTEGER := 15; BEGIN
    IF p_auth_method = 'AADHAAR' THEN RETURN NULL; END IF;
    IF p_office_code IS NOT NULL AND p_office_code LIKE 'CHQ%' THEN v_days := 30; END IF;
    RETURN p_created_at + (v_days || ' days')::INTERVAL;
END;
$$ LANGUAGE plpgsql;

-- Partition creation helper for audit_log
CREATE OR REPLACE FUNCTION nfs.create_audit_partition(p_year INTEGER) RETURNS VOID AS $$
DECLARE
    v_start DATE := make_date(p_year, 1, 1);
    v_end   DATE := make_date(p_year + 1, 1, 1);
    v_name  TEXT := 'nfs.audit_log_' || p_year;
BEGIN
    EXECUTE format('CREATE TABLE IF NOT EXISTS %I PARTITION OF nfs.audit_log FOR VALUES FROM (%L) TO (%L)', v_name, v_start, v_end);
END;
$$ LANGUAGE plpgsql;

-- ===========================================================================
-- SECTION 11: TRIGGERS
-- ===========================================================================

CREATE TRIGGER tr_service_request_timestamp
    BEFORE UPDATE ON nfs.service_request FOR EACH ROW EXECUTE FUNCTION nfs.update_timestamp();

CREATE TRIGGER tr_service_request_search
    BEFORE INSERT OR UPDATE ON nfs.service_request FOR EACH ROW EXECUTE FUNCTION nfs.update_service_request_search_vector();

CREATE TRIGGER tr_address_change_timestamp
    BEFORE UPDATE ON nfs.address_change_detail FOR EACH ROW EXECUTE FUNCTION nfs.update_timestamp();

CREATE TRIGGER tr_address_change_search
    BEFORE INSERT OR UPDATE ON nfs.address_change_detail FOR EACH ROW EXECUTE FUNCTION nfs.update_address_search_vector();

CREATE TRIGGER tr_name_change_timestamp
    BEFORE UPDATE ON nfs.name_change_detail FOR EACH ROW EXECUTE FUNCTION nfs.update_timestamp();

CREATE TRIGGER tr_name_change_search
    BEFORE INSERT OR UPDATE ON nfs.name_change_detail FOR EACH ROW EXECUTE FUNCTION nfs.update_name_search_vector();

CREATE TRIGGER tr_document_upload_timestamp
    BEFORE UPDATE ON nfs.document_upload FOR EACH ROW EXECUTE FUNCTION nfs.update_timestamp();

CREATE TRIGGER tr_missing_document_timestamp
    BEFORE UPDATE ON nfs.missing_document_request FOR EACH ROW EXECUTE FUNCTION nfs.update_timestamp();

CREATE TRIGGER tr_withdrawal_timestamp
    BEFORE UPDATE ON nfs.withdrawal_request FOR EACH ROW EXECUTE FUNCTION nfs.update_timestamp();

CREATE TRIGGER tr_cpc_work_queue_timestamp
    BEFORE UPDATE ON nfs.cpc_work_queue FOR EACH ROW EXECUTE FUNCTION nfs.update_timestamp();

-- ===========================================================================
-- SECTION 12: VIEWS
-- ===========================================================================

CREATE OR REPLACE VIEW nfs.v_active_requests AS
SELECT sr.request_id, sr.ticket_number, sr.customer_id, sr.policy_number,
       sr.request_type, sr.auth_method, sr.status, sr.channel, sr.office_code,
       sr.initiated_by, sr.assigned_to, sr.created_at, sr.sla_deadline,
       sr.sla_breached, sr.partial_processing_flag,
       CASE WHEN sr.sla_deadline IS NULL THEN 'NO_SLA'
            WHEN sr.sla_deadline < NOW() THEN 'BREACHED'
            WHEN sr.sla_deadline < NOW() + INTERVAL '2 days' THEN 'APPROACHING'
            ELSE 'WITHIN_SLA' END AS sla_status,
       EXTRACT(DAY FROM (NOW() - sr.created_at)) AS age_days
FROM nfs.service_request sr
WHERE sr.deleted_at IS NULL
  AND sr.status NOT IN ('COMPLETED','REJECTED','WITHDRAWN','DOCUMENTS_EXPIRED');

CREATE OR REPLACE VIEW nfs.v_pending_approval_requests AS
SELECT sr.request_id, sr.ticket_number, sr.customer_id, sr.policy_number,
       sr.request_type, sr.auth_method, sr.status, sr.channel, sr.office_code,
       sr.assigned_to, sr.created_at, sr.sla_deadline, sr.sla_breached,
       CASE WHEN sr.sla_deadline IS NULL THEN 'NO_SLA'
            WHEN sr.sla_deadline < NOW() THEN 'BREACHED'
            WHEN sr.sla_deadline < NOW() + INTERVAL '2 days' THEN 'APPROACHING'
            ELSE 'WITHIN_SLA' END AS sla_status,
       EXTRACT(DAY FROM (NOW() - sr.created_at)) AS age_days,
       wq.priority, wq.queue_status
FROM nfs.service_request sr
LEFT JOIN nfs.cpc_work_queue wq ON sr.request_id = wq.request_id
WHERE sr.deleted_at IS NULL AND sr.status IN ('PENDING_APPROVAL','IN_PROGRESS');

CREATE OR REPLACE VIEW nfs.v_address_change_summary AS
SELECT sr.request_id, sr.ticket_number, sr.customer_id, sr.policy_number,
       sr.status, sr.auth_method, sr.channel, sr.created_at, sr.completed_at, sr.sla_deadline,
       acd.address_type, acd.address_update_for, acd.old_city AS previous_city,
       acd.old_state AS previous_state, acd.new_city, acd.new_state, acd.new_pincode
FROM nfs.service_request sr
JOIN nfs.address_change_detail acd ON sr.request_id = acd.request_id
WHERE sr.request_type = 'ADDRESS_CHANGE' AND sr.deleted_at IS NULL;

CREATE OR REPLACE VIEW nfs.v_name_change_summary AS
SELECT sr.request_id, sr.ticket_number, sr.customer_id, sr.policy_number,
       sr.status, sr.auth_method, sr.channel, sr.created_at, sr.completed_at, sr.sla_deadline,
       ncd.old_salutation, ncd.old_first_name AS previous_first_name, ncd.old_last_name AS previous_last_name,
       ncd.new_salutation, ncd.new_first_name, ncd.new_last_name, ncd.policies_affected
FROM nfs.service_request sr
JOIN nfs.name_change_detail ncd ON sr.request_id = ncd.request_id
WHERE sr.request_type = 'NAME_CHANGE' AND sr.deleted_at IS NULL;

CREATE OR REPLACE VIEW nfs.v_cpc_dashboard AS
SELECT sr.office_code, sr.request_type, sr.status,
       COUNT(*) AS request_count,
       COUNT(CASE WHEN sr.sla_breached THEN 1 END) AS breached_count,
       AVG(EXTRACT(DAY FROM (NOW() - sr.created_at))) AS avg_age_days,
       MIN(sr.created_at) AS oldest_request,
       MAX(sr.created_at) AS newest_request
FROM nfs.service_request sr
WHERE sr.deleted_at IS NULL
  AND sr.status NOT IN ('COMPLETED','REJECTED','WITHDRAWN','DOCUMENTS_EXPIRED')
GROUP BY sr.office_code, sr.request_type, sr.status;

CREATE OR REPLACE VIEW nfs.v_audit_trail AS
SELECT al.audit_id, al.request_id, sr.ticket_number, al.action_type,
       al.old_value, al.new_value, al.performed_by, al.performed_at,
       al.ip_address, al.channel, al.office_code, al.remarks, al.error_code, al.error_message
FROM nfs.audit_log al JOIN nfs.service_request sr ON al.request_id = sr.request_id;

CREATE OR REPLACE VIEW nfs.v_document_status AS
SELECT sr.request_id, sr.ticket_number, sr.status AS request_status,
       du.document_id, du.document_type, du.file_name, du.verification_status,
       du.uploaded_at, du.verified_at, du.verified_by, du.rejection_reason AS document_rejection_reason
FROM nfs.service_request sr
JOIN nfs.document_upload du ON sr.request_id = du.request_id
WHERE sr.deleted_at IS NULL AND du.deleted_at IS NULL;

-- ===========================================================================
-- SECTION 13: MATERIALIZED VIEWS
-- ===========================================================================

CREATE MATERIALIZED VIEW nfs.mv_daily_statistics AS
SELECT DATE(created_at) AS stat_date, request_type, auth_method, channel, status,
       COUNT(*) AS request_count,
       AVG(EXTRACT(EPOCH FROM (completed_at - created_at)) / 60) AS avg_completion_time_minutes,
       COUNT(CASE WHEN sla_breached THEN 1 END) AS sla_breach_count
FROM nfs.service_request WHERE deleted_at IS NULL
GROUP BY DATE(created_at), request_type, auth_method, channel, status;

CREATE UNIQUE INDEX idx_mv_daily_statistics_date
    ON nfs.mv_daily_statistics (stat_date, request_type, auth_method, channel, status);

-- ===========================================================================
-- SECTION 14: ROLES AND PERMISSIONS
-- ===========================================================================

DO $$ BEGIN
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'nfs_service_role') THEN
        CREATE ROLE nfs_service_role;
    END IF;
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'nfs_cpc_role') THEN
        CREATE ROLE nfs_cpc_role;
    END IF;
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'nfs_admin_role') THEN
        CREATE ROLE nfs_admin_role;
    END IF;
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'nfs_readonly_role') THEN
        CREATE ROLE nfs_readonly_role;
    END IF;
END $$;

GRANT USAGE ON SCHEMA nfs TO nfs_service_role, nfs_cpc_role, nfs_admin_role, nfs_readonly_role;
GRANT SELECT, INSERT, UPDATE ON ALL TABLES IN SCHEMA nfs TO nfs_service_role;
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA nfs TO nfs_service_role;
GRANT EXECUTE ON ALL FUNCTIONS IN SCHEMA nfs TO nfs_service_role;

-- audit_log is INSERT-only for service role (BR-NFS-016)
REVOKE UPDATE, DELETE ON nfs.audit_log FROM nfs_service_role;

GRANT SELECT ON ALL TABLES IN SCHEMA nfs TO nfs_cpc_role;
GRANT UPDATE ON nfs.service_request TO nfs_cpc_role;
GRANT UPDATE ON nfs.cpc_work_queue TO nfs_cpc_role;
GRANT INSERT ON nfs.audit_log TO nfs_cpc_role;

GRANT ALL PRIVILEGES ON SCHEMA nfs TO nfs_admin_role;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA nfs TO nfs_admin_role;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA nfs TO nfs_admin_role;

GRANT SELECT ON ALL TABLES IN SCHEMA nfs TO nfs_readonly_role;

-- ===========================================================================
-- SECTION 15: SEED DATA
-- ===========================================================================

INSERT INTO nfs.ticket_sequence (sequence_date, request_type, sequence_value)
SELECT CURRENT_DATE, 'ADDRESS_CHANGE', 0
WHERE NOT EXISTS (SELECT 1 FROM nfs.ticket_sequence WHERE sequence_date = CURRENT_DATE AND request_type = 'ADDRESS_CHANGE');

INSERT INTO nfs.ticket_sequence (sequence_date, request_type, sequence_value)
SELECT CURRENT_DATE, 'NAME_CHANGE', 0
WHERE NOT EXISTS (SELECT 1 FROM nfs.ticket_sequence WHERE sequence_date = CURRENT_DATE AND request_type = 'NAME_CHANGE');

-- ===========================================================================
-- END OF MIGRATION
-- ===========================================================================
