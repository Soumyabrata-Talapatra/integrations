-- ============================================
-- Policy Surrender & Forced Surrender Database Schema
-- Database: surrender_db
-- PostgreSQL Version: 16
-- Optimization: Production-ready for 100K-1M rows
-- ============================================

-- Enable extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ============================================
-- ENUM Types
-- ============================================

CREATE TYPE surrender_request_type AS ENUM (
    'VOLUNTARY',
    'FORCED'
);

CREATE TYPE surrender_status AS ENUM (
    'PENDING_DOCUMENT_UPLOAD',
    'PENDING_VERIFICATION',
    'PENDING_APPROVAL',
    'APPROVED',
    'REJECTED',
    'PENDING_AUTO_COMPLETION',
    'AUTO_COMPLETED',
    'TERMINATED'
);

CREATE TYPE policy_status_surrender AS ENUM (
    'AP',
    'IL',
    'AL',
    'PWS',
    'PAS',
    'TAS',
    'TS',
    'AU'
);

CREATE TYPE previous_policy_status AS ENUM (
    'AP',
    'IL',
    'AL'
);

CREATE TYPE disbursement_method AS ENUM (
    'CASH',
    'CHEQUE'
);

CREATE TYPE document_type AS ENUM (
    'WRITTEN_CONSENT',
    'POLICY_BOND',
    'PREMIUM_RECEIPT_BOOK',
    'PAY_RECOVERY_CERTIFICATE',
    'LOAN_RECEIPT_BOOK',
    'LOAN_BOND',
    'INDEMNITY_BOND',
    'ASSIGNMENT_DEED',
    'DISCHARGE_RECEIPT'
);

CREATE TYPE reminder_level AS ENUM (
    'FIRST',
    'SECOND',
    'THIRD'
);

CREATE TYPE request_owner AS ENUM (
    'CUSTOMER',
    'SYSTEM',
    'POSTMASTER'
);

-- ============================================
-- TABLES
-- ============================================

-- Policy Surrender Request (partitioned by created_at)
CREATE TABLE policy_surrender_requests (
    id UUID DEFAULT uuid_generate_v4() NOT NULL,
    policy_id UUID NOT NULL,
    request_number VARCHAR(50) NOT NULL,
    request_type surrender_request_type NOT NULL,
    previous_policy_status previous_policy_status,
    request_date DATE NOT NULL,
    surrender_value_calculated_date DATE NOT NULL,
    gross_surrender_value NUMERIC(15,2) NOT NULL,
    net_surrender_value NUMERIC(15,2) NOT NULL,
    paid_up_value NUMERIC(15,2) NOT NULL,
    bonus_amount NUMERIC(15,2),
    surrender_factor NUMERIC(8,6) NOT NULL,
    unpaid_premiums_deduction NUMERIC(15,2) NOT NULL DEFAULT 0,
    loan_deduction NUMERIC(15,2) NOT NULL DEFAULT 0,
    other_deductions NUMERIC(15,2) DEFAULT 0,
    disbursement_method disbursement_method NOT NULL,
    disbursement_amount NUMERIC(15,2) NOT NULL,
    reason VARCHAR(500),
    status surrender_status NOT NULL DEFAULT 'PENDING_DOCUMENT_UPLOAD',
    owner request_owner NOT NULL DEFAULT 'CUSTOMER',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_by UUID NOT NULL,
    approved_by UUID,
    approved_at TIMESTAMP WITH TIME ZONE,
    approval_comments TEXT,
    deleted_at TIMESTAMP WITH TIME ZONE,
    version INTEGER NOT NULL DEFAULT 1,
    metadata JSONB DEFAULT '{}',
    search_vector tsvector,

    CONSTRAINT chk_gross_surrender_value_positive CHECK (gross_surrender_value >= 0),
    CONSTRAINT chk_net_surrender_value_positive CHECK (net_surrender_value >= 0),
    CONSTRAINT chk_paid_up_value_positive CHECK (paid_up_value >= 0),
    CONSTRAINT chk_bonus_amount_positive CHECK (bonus_amount IS NULL OR bonus_amount >= 0),
    CONSTRAINT chk_surrender_factor_range CHECK (surrender_factor > 0 AND surrender_factor <= 1),
    CONSTRAINT chk_deductions_positive CHECK (unpaid_premiums_deduction >= 0 AND loan_deduction >= 0 AND other_deductions >= 0),
    CONSTRAINT chk_disbursement_amount_positive CHECK (disbursement_amount >= 0),
    CONSTRAINT chk_request_date_valid CHECK (request_date <= CURRENT_DATE),
    CONSTRAINT chk_surrender_calculated_date_valid CHECK (surrender_value_calculated_date <= CURRENT_DATE),
    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);

-- Partitions for policy_surrender_requests (yearly)
CREATE TABLE policy_surrender_requests_2024 PARTITION OF policy_surrender_requests
    FOR VALUES FROM ('2024-01-01') TO ('2025-01-01');

CREATE TABLE policy_surrender_requests_2025 PARTITION OF policy_surrender_requests
    FOR VALUES FROM ('2025-01-01') TO ('2026-01-01');

CREATE TABLE policy_surrender_requests_2026 PARTITION OF policy_surrender_requests
    FOR VALUES FROM ('2026-01-01') TO ('2027-01-01');

CREATE TABLE policy_surrender_requests_default PARTITION OF policy_surrender_requests
    DEFAULT;

-- Surrender Bonus Details
CREATE TABLE surrender_bonus_details (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    surrender_request_id UUID NOT NULL,
    financial_year VARCHAR(9) NOT NULL,
    sum_assured NUMERIC(15,2) NOT NULL,
    bonus_rate NUMERIC(8,2) NOT NULL,
    bonus_amount NUMERIC(15,2) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_financial_year_format CHECK (financial_year ~ '^\d{4}-\d{4}$'),
    CONSTRAINT chk_sum_assured_positive CHECK (sum_assured > 0),
    CONSTRAINT chk_bonus_rate_positive CHECK (bonus_rate > 0),
    CONSTRAINT chk_bonus_amount_positive CHECK (bonus_amount >= 0)
);

-- Forced Surrender Reminders
CREATE TABLE forced_surrender_reminders (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    policy_id UUID NOT NULL,
    reminder_number reminder_level NOT NULL,
    reminder_date DATE NOT NULL,
    loan_capitalization_ratio NUMERIC(5,4) NOT NULL,
    loan_principal NUMERIC(15,2) NOT NULL,
    loan_interest NUMERIC(15,2) NOT NULL,
    gross_surrender_value NUMERIC(15,2) NOT NULL,
    letter_sent BOOLEAN NOT NULL DEFAULT FALSE,
    sms_sent BOOLEAN NOT NULL DEFAULT FALSE,
    letter_reference VARCHAR(50),
    sms_reference VARCHAR(50),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    metadata JSONB DEFAULT '{}',

    CONSTRAINT chk_loan_capitalization_ratio CHECK (loan_capitalization_ratio >= 0),
    CONSTRAINT chk_loan_principal_positive CHECK (loan_principal >= 0),
    CONSTRAINT chk_loan_interest_positive CHECK (loan_interest >= 0),
    CONSTRAINT chk_gross_surrender_value_positive CHECK (gross_surrender_value >= 0)
);

-- Surrender Documents (partitioned by uploaded_date)
CREATE TABLE surrender_documents (
    id UUID DEFAULT uuid_generate_v4() NOT NULL,
    surrender_request_id UUID NOT NULL,
    document_type document_type NOT NULL,
    document_name VARCHAR(255) NOT NULL,
    document_path TEXT NOT NULL,
    uploaded_date DATE NOT NULL DEFAULT CURRENT_DATE,
    file_size_bytes INTEGER,
    mime_type VARCHAR(100),
    verified BOOLEAN NOT NULL DEFAULT FALSE,
    verified_by UUID,
    verified_at TIMESTAMP WITH TIME ZONE,
    rejection_reason TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    metadata JSONB DEFAULT '{}',

    CONSTRAINT chk_file_size_positive CHECK (file_size_bytes IS NULL OR file_size_bytes > 0),
    CONSTRAINT chk_uploaded_date_valid CHECK (uploaded_date <= CURRENT_DATE),
    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);

-- Partitions for surrender_documents (yearly)
CREATE TABLE surrender_documents_2024 PARTITION OF surrender_documents
    FOR VALUES FROM ('2024-01-01') TO ('2025-01-01');

CREATE TABLE surrender_documents_2025 PARTITION OF surrender_documents
    FOR VALUES FROM ('2025-01-01') TO ('2026-01-01');

CREATE TABLE surrender_documents_2026 PARTITION OF surrender_documents
    FOR VALUES FROM ('2026-01-01') TO ('2027-01-01');

CREATE TABLE surrender_documents_default PARTITION OF surrender_documents
    DEFAULT;

-- Surrender Payments (partitioned by payment_date)
CREATE TABLE surrender_payments (
    id UUID DEFAULT uuid_generate_v4() NOT NULL,
    surrender_request_id UUID NOT NULL,
    payment_number VARCHAR(50) NOT NULL,
    payment_date DATE NOT NULL,
    amount NUMERIC(15,2) NOT NULL,
    disbursement_method disbursement_method NOT NULL,
    cheque_number VARCHAR(50),
    cheque_date DATE,
    bank_name VARCHAR(100),
    branch_name VARCHAR(100),
    payee_name VARCHAR(255) NOT NULL,
    payee_address TEXT,
    transaction_reference VARCHAR(100),
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    processed_at TIMESTAMP WITH TIME ZONE,
    processed_by UUID,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    metadata JSONB DEFAULT '{}',

    CONSTRAINT chk_payment_amount_positive CHECK (amount > 0),
    CONSTRAINT chk_payment_date_valid CHECK (payment_date <= CURRENT_DATE),
    CONSTRAINT chk_cheque_details_valid CHECK (
        (disbursement_method = 'CHEQUE' AND cheque_number IS NOT NULL AND cheque_date IS NOT NULL) OR
        (disbursement_method = 'CASH')
    ),
    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);

-- Partitions for surrender_payments (yearly)
CREATE TABLE surrender_payments_2024 PARTITION OF surrender_payments
    FOR VALUES FROM ('2024-01-01') TO ('2025-01-01');

CREATE TABLE surrender_payments_2025 PARTITION OF surrender_payments
    FOR VALUES FROM ('2025-01-01') TO ('2026-01-01');

CREATE TABLE surrender_payments_2026 PARTITION OF surrender_payments
    FOR VALUES FROM ('2026-01-01') TO ('2027-01-01');

CREATE TABLE surrender_payments_default PARTITION OF surrender_payments
    DEFAULT;

-- Approval Workflow Tasks
CREATE TABLE approval_workflow_tasks (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    surrender_request_id UUID NOT NULL,
    task_number VARCHAR(50) UNIQUE NOT NULL,
    office_code VARCHAR(20) NOT NULL,
    assigned_to UUID,
    status VARCHAR(20) NOT NULL DEFAULT 'PENDING',
    reserved BOOLEAN NOT NULL DEFAULT FALSE,
    reserved_at TIMESTAMP WITH TIME ZONE,
    reserved_by UUID,
    reservation_expires_at TIMESTAMP WITH TIME ZONE,
    priority VARCHAR(20) NOT NULL DEFAULT 'MEDIUM',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,
    completed_by UUID,
    escalated BOOLEAN NOT NULL DEFAULT FALSE,
    escalated_to UUID,
    escalated_at TIMESTAMP WITH TIME ZONE,
    escalation_reason TEXT,
    metadata JSONB DEFAULT '{}',

    CONSTRAINT chk_priority_valid CHECK (priority IN ('LOW', 'MEDIUM', 'HIGH', 'CRITICAL')),
    CONSTRAINT chk_status_valid CHECK (status IN ('PENDING', 'RESERVED', 'IN_PROGRESS', 'COMPLETED', 'ESCALATED'))
);

-- Surrender Request History (Audit Trail)
CREATE TABLE surrender_request_history (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    surrender_request_id UUID NOT NULL,
    changed_by UUID NOT NULL,
    changed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    old_status surrender_status,
    new_status surrender_status,
    change_type VARCHAR(50) NOT NULL,
    change_details JSONB,
    comments TEXT,
    ip_address INET,
    user_agent TEXT
);

-- Forced Surrender Payment Windows
CREATE TABLE forced_surrender_payment_windows (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    surrender_request_id UUID NOT NULL UNIQUE,
    policy_id UUID NOT NULL,
    window_start_date DATE NOT NULL,
    window_end_date DATE NOT NULL,
    payment_received BOOLEAN NOT NULL DEFAULT FALSE,
    payment_received_at TIMESTAMP WITH TIME ZONE,
    payment_amount NUMERIC(15,2),
    payment_reference VARCHAR(100),
    workflow_forwarded BOOLEAN NOT NULL DEFAULT FALSE,
    workflow_forwarded_at TIMESTAMP WITH TIME ZONE,
    auto_completed BOOLEAN NOT NULL DEFAULT FALSE,
    auto_completed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_window_dates_valid CHECK (window_end_date > window_start_date),
    CONSTRAINT chk_payment_amount_positive CHECK (payment_amount IS NULL OR payment_amount > 0)
);

-- Surrender Value Calculations (Audit Trail)
CREATE TABLE surrender_value_calculations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    surrender_request_id UUID NOT NULL,
    calculation_date DATE NOT NULL,
    paid_up_value NUMERIC(15,2) NOT NULL,
    bonus_amount NUMERIC(15,2),
    surrender_factor NUMERIC(8,6) NOT NULL,
    gross_surrender_value NUMERIC(15,2) NOT NULL,
    unpaid_premiums_deduction NUMERIC(15,2) NOT NULL DEFAULT 0,
    loan_principal_deduction NUMERIC(15,2) NOT NULL DEFAULT 0,
    loan_interest_deduction NUMERIC(15,2) NOT NULL DEFAULT 0,
    net_surrender_value NUMERIC(15,2) NOT NULL,
    calculation_breakdown JSONB,
    calculated_by UUID NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_paid_up_value_positive CHECK (paid_up_value >= 0),
    CONSTRAINT chk_gsv_positive CHECK (gross_surrender_value >= 0),
    CONSTRAINT chk_nsv_positive CHECK (net_surrender_value >= 0)
);

-- Policy Surrender Dispositions
CREATE TABLE policy_surrender_dispositions (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    surrender_request_id UUID NOT NULL UNIQUE,
    disposition_type VARCHAR(50) NOT NULL,
    new_policy_status policy_status_surrender,
    new_sum_assured NUMERIC(15,2),
    prescribed_limit NUMERIC(15,2),
    net_surrender_value NUMERIC(15,2) NOT NULL,
    reduced_paid_up_created BOOLEAN NOT NULL DEFAULT FALSE,
    reduced_paid_up_policy_number VARCHAR(50),
    terminated BOOLEAN NOT NULL DEFAULT FALSE,
    termination_reason TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT chk_disposition_type_valid CHECK (disposition_type IN ('REDUCED_PAID_UP', 'TERMINATED_SURRENDER', 'TERMINATED_AUTO_SURRENDER')),
    CONSTRAINT chk_new_sum_assured_positive CHECK (new_sum_assured IS NULL OR new_sum_assured > 0),
    CONSTRAINT chk_prescribed_limit_positive CHECK (prescribed_limit IS NULL OR prescribed_limit > 0)
);

-- ============================================
-- INDEXES
-- ============================================

-- Policy Surrender Requests indexes
-- Note: request_number uniqueness must be enforced at application level for partitioned tables
CREATE INDEX idx_surrender_requests_policy_id ON policy_surrender_requests(policy_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_surrender_requests_request_number ON policy_surrender_requests(request_number);
CREATE INDEX idx_surrender_requests_request_type ON policy_surrender_requests(request_type) WHERE deleted_at IS NULL;
CREATE INDEX idx_surrender_requests_status ON policy_surrender_requests(status) WHERE deleted_at IS NULL;
CREATE INDEX idx_surrender_requests_owner ON policy_surrender_requests(owner) WHERE deleted_at IS NULL;
CREATE INDEX idx_surrender_requests_request_date ON policy_surrender_requests(request_date) WHERE deleted_at IS NULL;
CREATE INDEX idx_surrender_requests_created_at ON policy_surrender_requests(created_at) WHERE deleted_at IS NULL;
CREATE INDEX idx_surrender_requests_created_by ON policy_surrender_requests(created_by) WHERE deleted_at IS NULL;
CREATE INDEX idx_surrender_requests_approved_by ON policy_surrender_requests(approved_by) WHERE deleted_at IS NULL;
CREATE INDEX idx_surrender_requests_composite_pending ON policy_surrender_requests(status, request_date)
    WHERE status IN ('PENDING_DOCUMENT_UPLOAD', 'PENDING_VERIFICATION', 'PENDING_APPROVAL') AND deleted_at IS NULL;
CREATE INDEX idx_surrender_requests_metadata ON policy_surrender_requests USING gin(metadata);
CREATE INDEX idx_surrender_requests_search ON policy_surrender_requests USING gin(search_vector);
CREATE INDEX idx_surrender_requests_forced_pending ON policy_surrender_requests(request_type, status)
    WHERE request_type = 'FORCED' AND status = 'PENDING_AUTO_COMPLETION' AND deleted_at IS NULL;

-- Surrender Bonus Details indexes
CREATE INDEX idx_bonus_details_surrender_request_id ON surrender_bonus_details(surrender_request_id);
CREATE INDEX idx_bonus_details_financial_year ON surrender_bonus_details(financial_year);

-- Forced Surrender Reminders indexes
CREATE INDEX idx_forced_reminders_policy_id ON forced_surrender_reminders(policy_id);
CREATE INDEX idx_forced_reminders_reminder_number ON forced_surrender_reminders(reminder_number);
CREATE INDEX idx_forced_reminders_reminder_date ON forced_surrender_reminders(reminder_date);
CREATE INDEX idx_forced_reminders_loan_ratio ON forced_surrender_reminders(loan_capitalization_ratio);
CREATE INDEX idx_forced_reminders_composite_policy ON forced_surrender_reminders(policy_id, reminder_number);

-- Surrender Documents indexes
CREATE INDEX idx_documents_surrender_request_id ON surrender_documents(surrender_request_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_documents_document_type ON surrender_documents(document_type) WHERE deleted_at IS NULL;
CREATE INDEX idx_documents_uploaded_date ON surrender_documents(uploaded_date) WHERE deleted_at IS NULL;
CREATE INDEX idx_documents_verified ON surrender_documents(verified) WHERE deleted_at IS NULL;
CREATE INDEX idx_documents_composite_request_type ON surrender_documents(surrender_request_id, document_type)
    WHERE deleted_at IS NULL;
CREATE INDEX idx_documents_metadata ON surrender_documents USING gin(metadata);

-- Surrender Payments indexes
-- Note: payment_number uniqueness must be enforced at application level for partitioned tables
CREATE INDEX idx_payments_surrender_request_id ON surrender_payments(surrender_request_id);
CREATE INDEX idx_payments_payment_number ON surrender_payments(payment_number);
CREATE INDEX idx_payments_payment_date ON surrender_payments(payment_date);
CREATE INDEX idx_payments_status ON surrender_payments(status);
CREATE INDEX idx_payments_disbursement_method ON surrender_payments(disbursement_method);
CREATE INDEX idx_payments_cheque_number ON surrender_payments(cheque_number) WHERE cheque_number IS NOT NULL;
CREATE INDEX idx_payments_transaction_reference ON surrender_payments(transaction_reference);
CREATE INDEX idx_payments_composite_date_status ON surrender_payments(payment_date, status);
CREATE INDEX idx_payments_metadata ON surrender_payments USING gin(metadata);

-- Approval Workflow Tasks indexes
CREATE INDEX idx_approval_tasks_surrender_request_id ON approval_workflow_tasks(surrender_request_id);
CREATE INDEX idx_approval_tasks_office_code ON approval_workflow_tasks(office_code);
CREATE INDEX idx_approval_tasks_assigned_to ON approval_workflow_tasks(assigned_to) WHERE status != 'COMPLETED';
CREATE INDEX idx_approval_tasks_status ON approval_workflow_tasks(status);
CREATE INDEX idx_approval_tasks_reserved ON approval_workflow_tasks(reserved) WHERE reserved = TRUE;
CREATE INDEX idx_approval_tasks_priority ON approval_workflow_tasks(priority) WHERE status IN ('PENDING', 'RESERVED');
CREATE INDEX idx_approval_tasks_composite_office_status ON approval_workflow_tasks(office_code, status, priority)
    WHERE status IN ('PENDING', 'RESERVED');
CREATE INDEX idx_approval_tasks_expiration ON approval_workflow_tasks(reservation_expires_at)
    WHERE reserved = TRUE;
CREATE INDEX idx_approval_tasks_metadata ON approval_workflow_tasks USING gin(metadata);

-- Surrender Request History indexes
CREATE INDEX idx_history_surrender_request_id ON surrender_request_history(surrender_request_id);
CREATE INDEX idx_history_changed_by ON surrender_request_history(changed_by);
CREATE INDEX idx_history_changed_at ON surrender_request_history(changed_at);
CREATE INDEX idx_history_change_type ON surrender_request_history(change_type);
CREATE INDEX idx_history_composite_request_date ON surrender_request_history(surrender_request_id, changed_at);

-- Forced Surrender Payment Windows indexes
CREATE INDEX idx_payment_windows_surrender_request_id ON forced_surrender_payment_windows(surrender_request_id);
CREATE INDEX idx_payment_windows_policy_id ON forced_surrender_payment_windows(policy_id);
CREATE INDEX idx_payment_windows_window_dates ON forced_surrender_payment_windows(window_start_date, window_end_date);
CREATE INDEX idx_payment_windows_payment_received ON forced_surrender_payment_windows(payment_received) WHERE NOT payment_received;
CREATE INDEX idx_payment_windows_workflow_forwarded ON forced_surrender_payment_windows(workflow_forwarded) WHERE NOT workflow_forwarded;

-- Surrender Value Calculations indexes
CREATE INDEX idx_calculations_surrender_request_id ON surrender_value_calculations(surrender_request_id);
CREATE INDEX idx_calculations_calculation_date ON surrender_value_calculations(calculation_date);
CREATE INDEX idx_calculations_calculated_by ON surrender_value_calculations(calculated_by);
CREATE INDEX idx_calculations_composite_request_date ON surrender_value_calculations(surrender_request_id, calculation_date);

-- Policy Surrender Dispositions indexes
CREATE INDEX idx_dispositions_surrender_request_id ON policy_surrender_dispositions(surrender_request_id);
CREATE INDEX idx_dispositions_disposition_type ON policy_surrender_dispositions(disposition_type);
CREATE INDEX idx_dispositions_new_policy_status ON policy_surrender_dispositions(new_policy_status);

-- ============================================
-- FUNCTIONS & TRIGGERS
-- ============================================

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    NEW.version = OLD.version + 1;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Function to update search vector for full-text search
CREATE OR REPLACE FUNCTION update_surrender_search_vector()
RETURNS TRIGGER AS $$
BEGIN
    NEW.search_vector := to_tsvector('english',
        COALESCE(NEW.request_number, '') || ' ' ||
        COALESCE(NEW.reason, '') || ' ' ||
        COALESCE(NEW.metadata->>'policy_number', '')
    );
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Function to log surrender status changes
CREATE OR REPLACE FUNCTION log_surrender_status_change()
RETURNS TRIGGER AS $$
BEGIN
    IF OLD.status IS DISTINCT FROM NEW.status OR TG_OP = 'INSERT' THEN
        INSERT INTO surrender_request_history (
            surrender_request_id,
            changed_by,
            old_status,
            new_status,
            change_type,
            change_details,
            comments
        ) VALUES (
            NEW.id,
            NEW.created_by,
            OLD.status,
            NEW.status,
            CASE TG_OP
                WHEN 'INSERT' THEN 'REQUEST_CREATED'
                WHEN 'UPDATE' THEN 'STATUS_CHANGE'
                ELSE 'UNKNOWN'
            END,
            jsonb_build_object(
                'old_status', OLD.status,
                'new_status', NEW.status,
                'timestamp', NOW()
            ),
            CASE
                WHEN NEW.approval_comments IS NOT NULL THEN NEW.approval_comments
                ELSE NULL
            END
        );
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Triggers for policy_surrender_requests
CREATE TRIGGER trg_surrender_requests_updated_at
    BEFORE UPDATE ON policy_surrender_requests
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER trg_surrender_requests_search_vector
    BEFORE INSERT OR UPDATE ON policy_surrender_requests
    FOR EACH ROW EXECUTE FUNCTION update_surrender_search_vector();

CREATE TRIGGER trg_surrender_requests_status_history
    AFTER INSERT OR UPDATE ON policy_surrender_requests
    FOR EACH ROW EXECUTE FUNCTION log_surrender_status_change();
