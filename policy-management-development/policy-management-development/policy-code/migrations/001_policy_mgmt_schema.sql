-- ============================================================================
-- POLICY MANAGEMENT ORCHESTRATOR — PostgreSQL Database Schema
-- ============================================================================
-- Project:  IMS PLI 2.0 — India Post Life Insurance Management System
-- Service:  Policy Management Orchestrator Microservice
-- Version:  1.0 (aligned with req_Policy_Management_Orchestrator_v4_1.md)
-- Date:     March 2026
-- ID Type:  BIGINT (BIGSERIAL) — as per project standard
-- Scale:    3M active policies, 50M after legacy migration
-- ============================================================================

 

-- ============================================================================
-- SCHEMA
-- ============================================================================

CREATE SCHEMA IF NOT EXISTS policy_mgmt;
SET search_path TO policy_mgmt;

-- ============================================================================
-- SEQUENCES (for BIGINT PKs)
-- ============================================================================

CREATE SEQUENCE seq_policy_id START 1000000 INCREMENT 1;
CREATE SEQUENCE seq_service_request_id START 1000000 INCREMENT 1;
CREATE SEQUENCE seq_status_history_id START 1 INCREMENT 1;
CREATE SEQUENCE seq_policy_event_id START 1 INCREMENT 1;
CREATE SEQUENCE seq_batch_scan_id START 1 INCREMENT 1;
CREATE SEQUENCE seq_signal_log_id START 1 INCREMENT 1;
CREATE SEQUENCE seq_signal_registry_id START 1 INCREMENT 1;

-- ============================================================================
-- ENUM TYPES
-- ============================================================================

-- 23 canonical lifecycle states (v4.1 — includes VOID_LAPSE)
CREATE TYPE lifecycle_status AS ENUM (
    'FREE_LOOK_ACTIVE',
    'ACTIVE',
    'VOID_LAPSE',              -- v4.1: <3yrs, in remission
    'INACTIVE_LAPSE',          -- v4.1: ≥3yrs, 12-month remission
    'ACTIVE_LAPSE',            -- v4.1: beyond remission, permanent
    'PAID_UP',
    'REDUCED_PAID_UP',
    'ASSIGNED_TO_PRESIDENT',
    'PENDING_AUTO_SURRENDER',
    'PENDING_SURRENDER',
    'REVIVAL_PENDING',
    'PENDING_MATURITY',
    'DEATH_CLAIM_INTIMATED',
    'DEATH_UNDER_INVESTIGATION',
    'SUSPENDED',
    'VOID',
    'SURRENDERED',
    'TERMINATED_SURRENDER',
    'MATURED',
    'DEATH_CLAIM_SETTLED',
    'FLC_CANCELLED',
    'CANCELLED_DEATH',
    'CONVERTED'
);

-- Request types (17 — matches Swagger RequestType enum)
CREATE TYPE request_type AS ENUM (
    'SURRENDER',
    'FORCED_SURRENDER',
    'LOAN',
    'LOAN_REPAYMENT',
    'REVIVAL',
    'DEATH_CLAIM',
    'MATURITY_CLAIM',
    'SURVIVAL_BENEFIT',
    'COMMUTATION',
    'CONVERSION',
    'FLC',
    'PAID_UP',
    'NOMINATION_CHANGE',
    'BILLING_METHOD_CHANGE',
    'ASSIGNMENT',
    'ADDRESS_CHANGE',
    'PREMIUM_REFUND',
    'DUPLICATE_BOND',
    'ADMIN_VOID',
    'REOPEN'
);

CREATE TYPE request_category AS ENUM (
    'FINANCIAL',
    'NON_FINANCIAL',
    'ADMIN'
);

-- Service request status lifecycle
CREATE TYPE request_status AS ENUM (
    'RECEIVED',
    'STATE_GATE_REJECTED',
    'ROUTED',
    'IN_PROGRESS',
    'COMPLETED',
    'CANCELLED',
    'WITHDRAWN',
    'TIMED_OUT',
    'AUTO_TERMINATED'
);

CREATE TYPE request_outcome AS ENUM (
    'APPROVED',
    'REJECTED',
    'WITHDRAWN',
    'TIMEOUT',
    'PREEMPTED',
    'DOMAIN_REJECTED'
);

CREATE TYPE source_channel AS ENUM (
    'CUSTOMER_PORTAL',
    'CPC',
    'MOBILE_APP',
    'AGENT_PORTAL',
    'BATCH',
    'SYSTEM'
);

CREATE TYPE premium_mode AS ENUM (
    'MONTHLY',
    'QUARTERLY',
    'HALF_YEARLY',
    'YEARLY'
);

CREATE TYPE billing_method AS ENUM (
    'CASH',
    'PAY_RECOVERY',
    'ONLINE'
);

CREATE TYPE product_type AS ENUM (
    'PLI',
    'RPLI'
);

CREATE TYPE assignment_type_enum AS ENUM (
    'NONE',
    'ABSOLUTE',
    'CONDITIONAL'
);

CREATE TYPE paid_up_type_enum AS ENUM (
    'AUTO',
    'VOLUNTARY',
    'REDUCED'
);

CREATE TYPE batch_scan_type AS ENUM (
    'LAPSATION',
    'REMISSION_EXPIRY_SHORT',
    'REMISSION_EXPIRY_LONG',
    'PAID_UP_CONVERSION',
    'MATURITY_SCAN',
    'FORCED_SURRENDER_EVAL'
);

CREATE TYPE batch_scan_status AS ENUM (
    'PENDING',
    'RUNNING',
    'COMPLETED',
    'FAILED'
);

CREATE TYPE signal_processing_status AS ENUM (
    'PROCESSED',
    'REJECTED',
    'DUPLICATE',
    'FAILED'
);

-- ============================================================================
-- TABLE 1: policy (Core policy state — source of truth)
-- ============================================================================
-- Source: §8.1 Entity: policy
-- Scale: 3M rows active, 50M total after migration
-- Access pattern: high read (status checks, queries), moderate write (state changes)

CREATE TABLE policy (
    policy_id               BIGINT          NOT NULL DEFAULT nextval('seq_policy_id'),
    policy_number           VARCHAR(30)     NOT NULL,
    customer_id             BIGINT          NOT NULL,
    product_code            VARCHAR(20)     NOT NULL,
    product_type            product_type    NOT NULL,

    -- Lifecycle state (PM-owned)
    current_status          lifecycle_status NOT NULL DEFAULT 'FREE_LOOK_ACTIVE',
    previous_status         lifecycle_status,
    previous_status_before_suspension lifecycle_status,
    effective_from          TIMESTAMPTZ     NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- Financial data
    sum_assured             DECIMAL(15,2)   NOT NULL,
    current_premium         DECIMAL(12,2)   NOT NULL,
    premium_mode            premium_mode    NOT NULL,
    billing_method          billing_method  NOT NULL DEFAULT 'CASH',

    -- Key dates
    issue_date              DATE            NOT NULL,
    policy_inception_date   DATE            NOT NULL,
    maturity_date           DATE,
    paid_to_date            DATE            NOT NULL,
    next_premium_due_date   DATE,

    -- Agent
    agent_id                BIGINT,

    -- Encumbrance flags (Tier 2 of hybrid state model)
    has_active_loan         BOOLEAN         NOT NULL DEFAULT FALSE,
    loan_outstanding        DECIMAL(15,2)   NOT NULL DEFAULT 0,
    assignment_type         assignment_type_enum NOT NULL DEFAULT 'NONE',
    assignment_status       VARCHAR(30)     NOT NULL DEFAULT 'UNASSIGNED',
    aml_hold                BOOLEAN         NOT NULL DEFAULT FALSE,
    dispute_flag            BOOLEAN         NOT NULL DEFAULT FALSE,
    murder_clause_active    BOOLEAN         DEFAULT FALSE,

    -- Display status (computed: lifecycle + encumbrances)
    display_status          VARCHAR(50)     NOT NULL DEFAULT 'FREE_LOOK_ACTIVE',

    -- Lapsation fields
    first_unpaid_premium_date DATE,
    remission_expiry_date     DATE,
    pay_recovery_protection_expiry DATE,

    -- Paid-up fields
    paid_up_value           DECIMAL(15,2),
    paid_up_type            paid_up_type_enum,
    paid_up_date            DATE,

    -- Survival benefit tracking
    sb_installments_paid    INTEGER         DEFAULT 0,
    sb_total_amount_paid    DECIMAL(15,2)   DEFAULT 0,

    -- Nomination
    nomination_status       VARCHAR(20)     NOT NULL DEFAULT 'ABSENT',

    -- WLA-specific
    policyholder_dob        DATE            NOT NULL,

    -- Temporal workflow references
    workflow_id             VARCHAR(100)    NOT NULL,
    temporal_run_id         VARCHAR(100),

    -- Optimistic locking
    version                 BIGINT          NOT NULL DEFAULT 1,

    -- Audit
    created_at              TIMESTAMPTZ     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at              TIMESTAMPTZ     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by              BIGINT,
    updated_by              BIGINT,

    -- Constraints
    CONSTRAINT pk_policy PRIMARY KEY (policy_id),
    CONSTRAINT uk_policy_number UNIQUE (policy_number),
    CONSTRAINT uk_workflow_id UNIQUE (workflow_id),
    CONSTRAINT ck_sum_assured_positive CHECK (sum_assured > 0),
    CONSTRAINT ck_premium_positive CHECK (current_premium > 0),
    CONSTRAINT ck_loan_non_negative CHECK (loan_outstanding >= 0),
    CONSTRAINT ck_version_positive CHECK (version > 0)
);

COMMENT ON TABLE policy IS 'Core policy state table — PM is the sole writer for lifecycle state fields. Source: §8.1';
COMMENT ON COLUMN policy.current_status IS '23 canonical lifecycle states (v4.1)';
COMMENT ON COLUMN policy.display_status IS 'Computed: lifecycle + encumbrances (e.g., ACTIVE_LOAN_LIEN)';
COMMENT ON COLUMN policy.pay_recovery_protection_expiry IS 'BR-PM-074: Pay recovery 12-month active protection. POLI Rules clause (i)/(ii)';
COMMENT ON COLUMN policy.version IS 'Optimistic locking — incremented on every update. Workflow checks version before write';

-- Primary access patterns
CREATE INDEX idx_policy_customer ON policy (customer_id);
CREATE INDEX idx_policy_status ON policy (current_status);
CREATE INDEX idx_policy_product ON policy (product_code, product_type);
CREATE INDEX idx_policy_agent ON policy (agent_id) WHERE agent_id IS NOT NULL;
CREATE INDEX idx_policy_maturity ON policy (maturity_date) WHERE maturity_date IS NOT NULL;
CREATE INDEX idx_policy_billing ON policy (billing_method);

-- Batch job access patterns
CREATE INDEX idx_policy_active_due ON policy (paid_to_date, premium_mode)
    WHERE current_status = 'ACTIVE';
CREATE INDEX idx_policy_void_lapse ON policy (remission_expiry_date)
    WHERE current_status = 'VOID_LAPSE';
CREATE INDEX idx_policy_inactive_lapse ON policy (remission_expiry_date)
    WHERE current_status = 'INACTIVE_LAPSE';
CREATE INDEX idx_policy_active_lapse ON policy (first_unpaid_premium_date)
    WHERE current_status = 'ACTIVE_LAPSE';
CREATE INDEX idx_policy_pay_recovery ON policy (pay_recovery_protection_expiry)
    WHERE billing_method = 'PAY_RECOVERY' AND pay_recovery_protection_expiry IS NOT NULL;
CREATE INDEX idx_policy_pending_maturity ON policy (maturity_date)
    WHERE current_status = 'ACTIVE' AND maturity_date IS NOT NULL;

-- ============================================================================
-- TABLE 2: policy_status_history (Audit trail for all state transitions)
-- ============================================================================
-- Source: §8.2 Entity: policy_status_history
-- Scale: ~10 transitions per policy lifetime avg → 50M+ rows
-- Retention: 10 years
-- Partitioned by effective_date for query performance and archival

CREATE TABLE policy_status_history (
    id                      BIGINT          NOT NULL DEFAULT nextval('seq_status_history_id'),
    policy_id               BIGINT          NOT NULL,
    from_status             lifecycle_status,
    to_status               lifecycle_status NOT NULL,
    transition_reason       VARCHAR(200)    NOT NULL,
    triggered_by_service    VARCHAR(50)     NOT NULL,
    triggered_by_user_id    BIGINT,
    request_id              BIGINT,
    effective_date          TIMESTAMPTZ     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    metadata_snapshot       JSONB,
    created_at              TIMESTAMPTZ     NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT pk_status_history PRIMARY KEY (id, effective_date)
) PARTITION BY RANGE (effective_date);

COMMENT ON TABLE policy_status_history IS 'Complete audit trail of every state transition. Retention: 10 years. Source: §8.2';

-- Yearly partitions
CREATE TABLE policy_status_history_2025 PARTITION OF policy_status_history
    FOR VALUES FROM ('2025-01-01') TO ('2026-01-01');
CREATE TABLE policy_status_history_2026 PARTITION OF policy_status_history
    FOR VALUES FROM ('2026-01-01') TO ('2027-01-01');
CREATE TABLE policy_status_history_2027 PARTITION OF policy_status_history
    FOR VALUES FROM ('2027-01-01') TO ('2028-01-01');
CREATE TABLE policy_status_history_2028 PARTITION OF policy_status_history
    FOR VALUES FROM ('2028-01-01') TO ('2029-01-01');
CREATE TABLE policy_status_history_default PARTITION OF policy_status_history
    DEFAULT;

CREATE INDEX idx_psh_policy_date ON policy_status_history (policy_id, effective_date DESC);
CREATE INDEX idx_psh_request ON policy_status_history (request_id) WHERE request_id IS NOT NULL;
CREATE INDEX idx_psh_service ON policy_status_history (triggered_by_service, effective_date DESC);

-- ============================================================================
-- TABLE 3: service_request (Central Request Registry)
-- ============================================================================
-- Source: §8.3 Entity: service_request
-- Scale: ~20K new requests/day, millions over time
-- Access: CPC inbox queries, request tracking, audit
-- Partitioned by submitted_at for performance

CREATE TABLE service_request (
    request_id              BIGINT          NOT NULL DEFAULT nextval('seq_service_request_id'),
    policy_id               BIGINT          NOT NULL,
    policy_number           VARCHAR(30)     NOT NULL,
    request_type            request_type    NOT NULL,
    request_category        request_category NOT NULL,
    status                  request_status  NOT NULL DEFAULT 'RECEIVED',
    source_channel          source_channel  NOT NULL,
    submitted_by            BIGINT,
    submitted_at            TIMESTAMPTZ     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    state_gate_status       lifecycle_status,
    routed_at               TIMESTAMPTZ,
    downstream_service      VARCHAR(50),
    downstream_workflow_id  VARCHAR(100),
    downstream_task_queue   VARCHAR(50),
    completed_at            TIMESTAMPTZ,
    outcome                 request_outcome,
    outcome_reason          VARCHAR(500),
    outcome_payload         JSONB,
    request_payload         JSONB           NOT NULL,
    timeout_at              TIMESTAMPTZ,
    idempotency_key         VARCHAR(100),
    created_at              TIMESTAMPTZ     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at              TIMESTAMPTZ     NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT pk_service_request PRIMARY KEY (request_id, submitted_at)
) PARTITION BY RANGE (submitted_at);

COMMENT ON TABLE service_request IS 'Central request registry — single source of truth for all policy operations. Source: §8.3';
COMMENT ON COLUMN service_request.request_id IS 'PM-generated ID. Downstream services use as FK for their domain records';
COMMENT ON COLUMN service_request.idempotency_key IS 'Client-provided idempotency key (X-Idempotency-Key header)';

CREATE TABLE IF NOT EXISTS request_idempotency_registry (
    idempotency_key VARCHAR(100) PRIMARY KEY,
    request_id BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_idempotency_created
ON request_idempotency_registry(created_at);

-- Partitions by quarter (high volume)
CREATE TABLE service_request_2026_q1 PARTITION OF service_request
    FOR VALUES FROM ('2026-01-01') TO ('2026-04-01');
CREATE TABLE service_request_2026_q2 PARTITION OF service_request
    FOR VALUES FROM ('2026-04-01') TO ('2026-07-01');
CREATE TABLE service_request_2026_q3 PARTITION OF service_request
    FOR VALUES FROM ('2026-07-01') TO ('2026-10-01');
CREATE TABLE service_request_2026_q4 PARTITION OF service_request
    FOR VALUES FROM ('2026-10-01') TO ('2027-01-01');
CREATE TABLE service_request_default PARTITION OF service_request
    DEFAULT;

-- CPC inbox query (most critical query pattern)
CREATE INDEX idx_sr_pending_cpc ON service_request (status, request_type, submitted_at DESC)
    WHERE status IN ('RECEIVED', 'ROUTED', 'IN_PROGRESS');
CREATE INDEX idx_sr_policy_id ON service_request (policy_id, submitted_at DESC);
CREATE INDEX idx_sr_policy_number ON service_request (policy_number, submitted_at DESC);
CREATE INDEX idx_sr_status_active ON service_request (status)
    WHERE status NOT IN ('COMPLETED', 'STATE_GATE_REJECTED', 'CANCELLED');
CREATE INDEX idx_sr_type_status ON service_request (request_type, status);
CREATE INDEX idx_sr_submitted ON service_request (submitted_at DESC);
CREATE INDEX idx_sr_downstream_wf ON service_request (downstream_workflow_id)
    WHERE downstream_workflow_id IS NOT NULL;
CREATE INDEX idx_sr_timeout ON service_request (timeout_at)
    WHERE status = 'ROUTED' AND timeout_at IS NOT NULL;

-- ============================================================================
-- TABLE 4: policy_lock (Financial request mutual exclusion)
-- ============================================================================
-- Source: §8.4 Entity: policy_lock
-- Scale: At most 1 row per policy at any time (sparse table)
-- Access: Check on every financial request, release on completion

CREATE TABLE policy_lock (
    policy_id               BIGINT          NOT NULL,
    request_id              BIGINT          NOT NULL,
    request_type            request_type    NOT NULL,
    locked_at               TIMESTAMPTZ     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    timeout_at              TIMESTAMPTZ     NOT NULL,

    CONSTRAINT pk_policy_lock PRIMARY KEY (policy_id)
);

COMMENT ON TABLE policy_lock IS 'Financial request mutual exclusion — at most ONE lock per policy. Source: §8.4, BR-PM-030';
COMMENT ON COLUMN policy_lock.timeout_at IS 'Auto-release deadline. If lock expires, PM reverts state and releases.';

CREATE INDEX idx_lock_timeout ON policy_lock (timeout_at)
    WHERE timeout_at IS NOT NULL;
CREATE INDEX idx_lock_request ON policy_lock (request_id);

-- ============================================================================
-- TABLE 5: policy_event (Published state change events)
-- ============================================================================
-- Source: §8.5 Entity: policy_event
-- Scale: 1 event per state transition → high volume
-- Partitioned by published_at

CREATE TABLE policy_event (
    event_id                BIGINT          NOT NULL DEFAULT nextval('seq_policy_event_id'),
    policy_id               BIGINT          NOT NULL,
    event_type              VARCHAR(50)     NOT NULL,
    event_payload           JSONB           NOT NULL,
    published_at            TIMESTAMPTZ     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    consumed_by             TEXT[],

    CONSTRAINT pk_policy_event PRIMARY KEY (event_id, published_at)
) PARTITION BY RANGE (published_at);

COMMENT ON TABLE policy_event IS 'Published state change events for downstream consumption. Retention: 7 years. Source: §8.5';

CREATE TABLE policy_event_2026 PARTITION OF policy_event
    FOR VALUES FROM ('2026-01-01') TO ('2027-01-01');
CREATE TABLE policy_event_2027 PARTITION OF policy_event
    FOR VALUES FROM ('2027-01-01') TO ('2028-01-01');
CREATE TABLE policy_event_default PARTITION OF policy_event
    DEFAULT;

CREATE INDEX idx_pe_policy ON policy_event (policy_id, published_at DESC);
CREATE INDEX idx_pe_type ON policy_event (event_type, published_at DESC);

-- ============================================================================
-- TABLE 6: batch_scan_state (Batch job execution tracking)
-- ============================================================================
-- Source: §8.6 Entity: batch_scan_state
-- Scale: ~10 rows/day (6 scan types × ~1-2 runs)

CREATE TABLE batch_scan_state (
    scan_id                 BIGINT          NOT NULL DEFAULT nextval('seq_batch_scan_id'),
    scan_type               batch_scan_type NOT NULL,
    scheduled_date          DATE            NOT NULL,
    started_at              TIMESTAMPTZ,
    completed_at            TIMESTAMPTZ,
    policies_scanned        INTEGER         NOT NULL DEFAULT 0,
    transitions_applied     INTEGER         NOT NULL DEFAULT 0,
    errors                  INTEGER         NOT NULL DEFAULT 0,
    error_details           JSONB,
    status                  batch_scan_status NOT NULL DEFAULT 'PENDING',
    duration_seconds        INTEGER,

    CONSTRAINT pk_batch_scan PRIMARY KEY (scan_id),
    CONSTRAINT uk_scan_type_date UNIQUE (scan_type, scheduled_date)
);

COMMENT ON TABLE batch_scan_state IS 'Tracks execution of daily/monthly batch jobs. Source: §8.6, FR-PM-011 to FR-PM-015';

CREATE INDEX idx_bs_date ON batch_scan_state (scheduled_date DESC);
CREATE INDEX idx_bs_status ON batch_scan_state (status) WHERE status IN ('PENDING', 'RUNNING');

-- ============================================================================
-- TABLE 7: policy_state_config (Configurable parameters)
-- ============================================================================
-- Source: §8.7 Entity: policy_state_config
-- Scale: ~30 rows (configuration parameters)

CREATE TABLE policy_state_config (
    config_key              VARCHAR(100)    NOT NULL,
    config_value            VARCHAR(500)    NOT NULL,
    description             VARCHAR(500),
    data_type               VARCHAR(20)     NOT NULL DEFAULT 'STRING',
    updated_at              TIMESTAMPTZ     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_by              BIGINT,

    CONSTRAINT pk_config PRIMARY KEY (config_key),
    CONSTRAINT ck_data_type CHECK (data_type IN ('STRING', 'INTEGER', 'DURATION', 'BOOLEAN', 'DECIMAL'))
);

COMMENT ON TABLE policy_state_config IS 'Configurable parameters for PM operations. Source: §8.7';

-- ============================================================================
-- TABLE 8: processed_signal_registry (Dedup tracking)
-- ============================================================================
-- Source: §8.8 Entity: processed_signal_registry
-- Scale: Grows with signal volume, evicted after 90 days

CREATE TABLE processed_signal_registry (
    id                      BIGINT          NOT NULL DEFAULT nextval('seq_signal_registry_id'),
    request_id              VARCHAR(100)    NOT NULL,
    signal_type             VARCHAR(50)     NOT NULL,
    policy_id               BIGINT          NOT NULL,
    received_at             TIMESTAMPTZ     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at              TIMESTAMPTZ     NOT NULL,

    CONSTRAINT pk_signal_registry PRIMARY KEY (id),
    CONSTRAINT uk_signal_dedup UNIQUE (request_id, signal_type)
);

COMMENT ON TABLE processed_signal_registry IS 'Signal deduplication. Entries evicted after 90 days (policy_state_config.signal_dedup_ttl_days). Source: §8.8';

CREATE INDEX idx_psr_policy ON processed_signal_registry (policy_id);
CREATE INDEX idx_psr_expires ON processed_signal_registry (expires_at);

-- ============================================================================
-- TABLE 9: policy_signal_log (Full signal audit trail)
-- ============================================================================
-- Source: §8.9 Entity: policy_signal_log
-- Scale: Every signal logged — high volume
-- Retention: 3 years
-- Partitioned by received_at

CREATE TABLE policy_signal_log (
    id                      BIGINT          NOT NULL DEFAULT nextval('seq_signal_log_id'),
    policy_id               BIGINT          NOT NULL,
    signal_channel          VARCHAR(50)     NOT NULL,
    signal_payload          JSONB           NOT NULL,
    source_service          VARCHAR(50)     NOT NULL,
    source_workflow_id      VARCHAR(100),
    request_id              VARCHAR(100)    NOT NULL,
    received_at             TIMESTAMPTZ     NOT NULL DEFAULT CURRENT_TIMESTAMP,
    processed_at            TIMESTAMPTZ,
    status                  signal_processing_status NOT NULL,
    rejection_reason        VARCHAR(200),
    state_before            lifecycle_status,
    state_after             lifecycle_status,

    CONSTRAINT pk_signal_log PRIMARY KEY (id, received_at)
) PARTITION BY RANGE (received_at);

COMMENT ON TABLE policy_signal_log IS 'Full audit trail of EVERY signal received — including rejected, duplicate, failed. Retention: 3 years. Source: §8.9';

CREATE TABLE policy_signal_log_2026 PARTITION OF policy_signal_log
    FOR VALUES FROM ('2026-01-01') TO ('2027-01-01');
CREATE TABLE policy_signal_log_2027 PARTITION OF policy_signal_log
    FOR VALUES FROM ('2027-01-01') TO ('2028-01-01');
CREATE TABLE policy_signal_log_default PARTITION OF policy_signal_log
    DEFAULT;

CREATE INDEX idx_psl_policy ON policy_signal_log (policy_id, received_at DESC);
CREATE INDEX idx_psl_channel ON policy_signal_log (signal_channel, received_at DESC);
CREATE INDEX idx_psl_status ON policy_signal_log (status, received_at DESC)
    WHERE status != 'PROCESSED';
CREATE INDEX idx_psl_request ON policy_signal_log (request_id);

-- ============================================================================
-- TABLE 10: terminal_state_snapshot (Cooling period final state)
-- ============================================================================
-- Source: §9.5.1 Terminal Cooling
-- When a policy enters terminal state, its full state is persisted here
-- for post-workflow query access (REST API reads from DB after workflow ends)

CREATE TABLE terminal_state_snapshot (
    policy_id               BIGINT          NOT NULL,
    policy_number           VARCHAR(30)     NOT NULL,
    final_status            lifecycle_status NOT NULL,
    terminal_at             TIMESTAMPTZ     NOT NULL,
    cooling_expiry          TIMESTAMPTZ     NOT NULL,
    workflow_completed_at   TIMESTAMPTZ,
    final_snapshot          JSONB           NOT NULL,
    created_at              TIMESTAMPTZ     NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT pk_terminal_snapshot PRIMARY KEY (policy_id)
);

COMMENT ON TABLE terminal_state_snapshot IS 'Persisted when policy enters terminal state. REST APIs read from here after workflow ends. Source: §9.5.1';

CREATE INDEX idx_tss_number ON terminal_state_snapshot (policy_number);
CREATE INDEX idx_tss_cooling ON terminal_state_snapshot (cooling_expiry)
    WHERE workflow_completed_at IS NULL;

-- ============================================================================
-- FUNCTIONS
-- ============================================================================

-- Generate policy number: PLI/YYYY/NNNNNN or RPLI/YYYY/NNNNNN
CREATE OR REPLACE FUNCTION generate_policy_number(p_product_type TEXT)
RETURNS VARCHAR AS $$
DECLARE
    v_prefix VARCHAR(4);
    v_year VARCHAR(4);
    v_next INTEGER;
BEGIN
    v_prefix := CASE p_product_type WHEN 'PLI' THEN 'PLI' ELSE 'RPLI' END;
    v_year := TO_CHAR(CURRENT_DATE, 'YYYY');

    SELECT COALESCE(MAX(
        CAST(SUBSTRING(policy_number FROM v_prefix || '/' || v_year || '/(\d+)') AS INTEGER)
    ), 0) + 1 INTO v_next
    FROM policy
    WHERE policy_number LIKE v_prefix || '/' || v_year || '/%';

    RETURN v_prefix || '/' || v_year || '/' || LPAD(v_next::TEXT, 6, '0');
END;
$$ LANGUAGE plpgsql;

-- Compute display status from lifecycle + encumbrances
CREATE OR REPLACE FUNCTION policy_mgmt.compute_display_status(
    p_status policy_mgmt.lifecycle_status,
    p_has_loan BOOLEAN,
    p_assignment policy_mgmt.assignment_type_enum,
    p_aml_hold BOOLEAN,
    p_dispute BOOLEAN
)
RETURNS VARCHAR
LANGUAGE plpgsql
IMMUTABLE
AS $$
BEGIN
    RETURN p_status::TEXT
        || CASE WHEN p_has_loan THEN '_LOAN' ELSE '' END
        || CASE WHEN p_assignment != 'NONE' THEN '_' || p_assignment::TEXT ELSE '' END
        || CASE WHEN p_aml_hold THEN '_AML_HOLD' ELSE '' END
        || CASE WHEN p_dispute THEN '_DISPUTED' ELSE '' END;
END;
$$;

-- Check if status is terminal
CREATE OR REPLACE FUNCTION is_terminal_status(p_status lifecycle_status)
RETURNS BOOLEAN AS $$
BEGIN
    RETURN p_status IN (
        'VOID', 'SURRENDERED', 'TERMINATED_SURRENDER', 'MATURED',
        'DEATH_CLAIM_SETTLED', 'FLC_CANCELLED', 'CANCELLED_DEATH', 'CONVERTED'
    );
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- Last day of month (for grace period calculation — BR-PM-040)
CREATE OR REPLACE FUNCTION last_day_of_month(p_date DATE)
RETURNS DATE AS $$
BEGIN
    RETURN (DATE_TRUNC('month', p_date) + INTERVAL '1 month' - INTERVAL '1 day')::DATE;
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- Compute remission expiry date (BR-PM-070)
CREATE OR REPLACE FUNCTION compute_remission_expiry(
    p_first_unpaid DATE,
    p_policy_inception DATE,
    p_billing_method billing_method
) RETURNS DATE AS $$
DECLARE
    v_policy_life_months INTEGER;
    v_grace_end DATE;
BEGIN
    v_policy_life_months := EXTRACT(YEAR FROM AGE(p_first_unpaid, p_policy_inception)) * 12
                          + EXTRACT(MONTH FROM AGE(p_first_unpaid, p_policy_inception));
    v_grace_end := last_day_of_month(p_first_unpaid);

    IF v_policy_life_months < 6 THEN
        RETURN NULL;  -- No remission
    ELSIF v_policy_life_months < 12 THEN
        RETURN v_grace_end + INTERVAL '30 days';
    ELSIF v_policy_life_months < 24 THEN
        RETURN v_grace_end + INTERVAL '60 days';
    ELSIF v_policy_life_months < 36 THEN
        RETURN v_grace_end + INTERVAL '90 days';
    ELSE
        -- ≥36 months: 12-month remission from first unpaid
        RETURN p_first_unpaid + INTERVAL '12 months';
    END IF;
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- ============================================================================
-- TRIGGERS
-- ============================================================================

-- Auto-update updated_at timestamp
CREATE OR REPLACE FUNCTION trigger_set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_policy_updated_at
    BEFORE UPDATE ON policy
    FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();

CREATE TRIGGER trg_service_request_updated_at
    BEFORE UPDATE ON service_request
    FOR EACH ROW EXECUTE FUNCTION trigger_set_updated_at();

-- Auto-increment version on policy update
CREATE OR REPLACE FUNCTION trigger_increment_version()
RETURNS TRIGGER AS $$
BEGIN
    NEW.version = OLD.version + 1;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_policy_version
    BEFORE UPDATE ON policy
    FOR EACH ROW EXECUTE FUNCTION trigger_increment_version();

-- Auto-compute display status on policy update
CREATE OR REPLACE FUNCTION trigger_compute_display_status()
RETURNS TRIGGER AS $$
BEGIN
    NEW.display_status = policy_mgmt.compute_display_status(
        NEW.current_status, NEW.has_active_loan,
        NEW.assignment_type, NEW.aml_hold, NEW.dispute_flag
    );
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_policy_display_status
    BEFORE INSERT OR UPDATE ON policy
    FOR EACH ROW EXECUTE FUNCTION trigger_compute_display_status();

-- ============================================================================
-- SEED DATA: policy_state_config
-- ============================================================================

INSERT INTO policy_state_config (config_key, config_value, description, data_type) VALUES
    -- Grace period (v4.1 CORRECTED)
    ('grace_period_rule', 'MONTH_END', 'Grace period extends to last day of premium due month for ALL premium modes', 'STRING'),

    -- Paid-up thresholds
    ('paid_up_min_inforce_months', '36', 'Minimum months policy must be in force for paid-up', 'INTEGER'),
    ('paid_up_min_lapse_months', '12', 'Minimum months since first unpaid for auto paid-up', 'INTEGER'),
    ('paid_up_min_value', '10000', 'Minimum PU value in Rs — below this → VOID', 'DECIMAL'),

    -- Notification windows
    ('maturity_notification_days', '90', 'Days before maturity to send notification / create PENDING_MATURITY', 'INTEGER'),
    ('forced_surrender_window_days', '30', 'Days policyholder has to pay after 3rd reminder before FS approval', 'INTEGER'),

    -- SLA periods
    ('revival_first_installment_sla_days', '60', 'Days allowed for first revival installment', 'INTEGER'),
    ('flc_period_days', '15', 'Free look cancellation period (standard)', 'INTEGER'),
    ('flc_period_distance_marketing_days', '30', 'FLC period for distance marketing policies', 'INTEGER'),

    -- Temporal workflow settings
    ('signal_dedup_ttl_days', '90', 'Days to retain processed signal IDs for dedup', 'INTEGER'),
    ('continue_as_new_max_events', '40000', 'CAN threshold: max event history count', 'INTEGER'),
    ('continue_as_new_max_days', '30', 'CAN threshold: max days between CAN cycles', 'INTEGER'),

    -- v4.1: Configurable routing timeouts
    ('routing_timeout_surrender', '30d', 'Surrender processing timeout', 'DURATION'),
    ('routing_timeout_forced_surrender', '60d', 'Forced surrender processing timeout', 'DURATION'),
    ('routing_timeout_loan', '14d', 'Loan processing timeout', 'DURATION'),
    ('routing_timeout_loan_repayment', '7d', 'Loan repayment processing timeout', 'DURATION'),
    ('routing_timeout_revival', '365d', 'Revival processing timeout (long due to installments)', 'DURATION'),
    ('routing_timeout_death_claim', '90d', 'Death claim processing timeout', 'DURATION'),
    ('routing_timeout_maturity_claim', '30d', 'Maturity claim processing timeout', 'DURATION'),
    ('routing_timeout_survival_benefit', '30d', 'Survival benefit processing timeout', 'DURATION'),
    ('routing_timeout_commutation', '30d', 'Commutation processing timeout', 'DURATION'),
    ('routing_timeout_conversion', '90d', 'Conversion processing timeout', 'DURATION'),
    ('routing_timeout_flc', '15d', 'Freelook cancellation processing timeout', 'DURATION'),
    ('routing_timeout_nfr', '14d', 'Non-financial request processing timeout', 'DURATION'),
    ('routing_timeout_premium_refund', '14d', 'Premium refund processing timeout', 'DURATION'),

    -- Terminal cooling durations (§9.5.1)
    ('cooling_period_void', '60d', 'Cooling period after VOID (late payment refunds)', 'DURATION'),
    ('cooling_period_surrendered', '90d', 'Cooling period after SURRENDERED', 'DURATION'),
    ('cooling_period_terminated_surrender', '90d', 'Cooling period after TERMINATED_SURRENDER', 'DURATION'),
    ('cooling_period_matured', '90d', 'Cooling period after MATURED', 'DURATION'),
    ('cooling_period_death_claim_settled', '180d', 'Cooling period after DEATH_CLAIM_SETTLED (ombudsman reopen)', 'DURATION'),
    ('cooling_period_flc_cancelled', '30d', 'Cooling period after FLC_CANCELLED', 'DURATION'),
    ('cooling_period_cancelled_death', '30d', 'Cooling period after CANCELLED_DEATH', 'DURATION'),
    ('cooling_period_converted', '90d', 'Cooling period after CONVERTED (cheque bounce reversal)', 'DURATION');

-- ============================================================================
-- MATERIALIZED VIEW: Policy Dashboard Metrics
-- ============================================================================
-- For GET /api/v1/policies/dashboard/metrics endpoint
-- Refreshed periodically via pg_cron (not on every write)

CREATE MATERIALIZED VIEW mv_policy_dashboard AS
SELECT
    current_status,
    product_type,
    product_code,
    billing_method,
    COUNT(*)                                     AS policy_count,
    SUM(sum_assured)                             AS total_sum_assured,
    SUM(current_premium)                         AS total_premium,
    SUM(CASE WHEN has_active_loan THEN 1 ELSE 0 END) AS with_active_loan,
    SUM(loan_outstanding)                        AS total_loan_outstanding
FROM policy
GROUP BY current_status, product_type, product_code, billing_method;

CREATE UNIQUE INDEX idx_mv_dashboard
    ON mv_policy_dashboard (current_status, product_type, product_code, billing_method);

COMMENT ON MATERIALIZED VIEW mv_policy_dashboard IS 'Pre-aggregated metrics for admin dashboard. Refresh: every 15 minutes via pg_cron.';

-- ============================================================================
-- MATERIALIZED VIEW: CPC Pending Summary
-- ============================================================================
-- For GET /api/v1/requests/pending/summary endpoint

CREATE MATERIALIZED VIEW mv_pending_summary AS
SELECT
    request_type,
    status,
    source_channel,
    COUNT(*)                                     AS request_count,
    MIN(submitted_at)                            AS oldest_submitted_at,
    EXTRACT(EPOCH FROM (CURRENT_TIMESTAMP - MIN(submitted_at))) / 3600 AS oldest_age_hours
FROM service_request
WHERE status IN ('RECEIVED', 'ROUTED', 'IN_PROGRESS')
GROUP BY request_type, status, source_channel;

CREATE UNIQUE INDEX idx_mv_pending
    ON mv_pending_summary (request_type, status, source_channel);

COMMENT ON MATERIALIZED VIEW mv_pending_summary IS 'Pre-aggregated CPC inbox summary. Refresh: every 1 minute via pg_cron.';

-- ============================================================================
-- pg_cron SCHEDULES (enable pg_cron extension first)
-- ============================================================================
-- Uncomment after enabling pg_cron extension:
--
-- SELECT cron.schedule('refresh-dashboard', '*/15 * * * *',
--     'REFRESH MATERIALIZED VIEW CONCURRENTLY policy_mgmt.mv_policy_dashboard');
--
-- SELECT cron.schedule('refresh-pending', '* * * * *',
--     'REFRESH MATERIALIZED VIEW CONCURRENTLY policy_mgmt.mv_pending_summary');
--
-- SELECT cron.schedule('evict-expired-signals', '0 3 * * *',
--     'DELETE FROM policy_mgmt.processed_signal_registry WHERE expires_at < CURRENT_TIMESTAMP');

-- ============================================================================
-- GRANTS (adjust roles as per deployment)
-- ============================================================================

-- Application role (PM service)
-- GRANT USAGE ON SCHEMA policy_mgmt TO pm_app;
-- GRANT SELECT, INSERT, UPDATE ON ALL TABLES IN SCHEMA policy_mgmt TO pm_app;
-- GRANT USAGE ON ALL SEQUENCES IN SCHEMA policy_mgmt TO pm_app;
-- GRANT SELECT ON ALL TABLES IN SCHEMA policy_mgmt TO pm_readonly;

-- ============================================================================
-- STATISTICS
-- ============================================================================

-- 10 tables, 4 materialized views, 7 sequences, 5 functions, 3 triggers
-- Estimated row counts at steady state:
--   policy:                     3,000,000 active (50M after migration)
--   policy_status_history:      30,000,000+ (10 transitions per policy avg)
--   service_request:            7,300,000/year (~20K/day)
--   policy_lock:                ~5,000 max concurrent
--   policy_event:               30,000,000+ (mirrors status_history)
--   batch_scan_state:           ~3,650/year
--   policy_state_config:        ~35 rows
--   processed_signal_registry:  ~500K active (90-day TTL)
--   policy_signal_log:          ~175M/year (~480K/day)
--   terminal_state_snapshot:    grows as policies terminate

 


CREATE OR REPLACE FUNCTION policy_mgmt.create_service_request(
    p_idempotency_key VARCHAR,
    p_policy_id BIGINT,
    p_policy_number VARCHAR,
    p_request_type policy_mgmt.request_type,
    p_request_category policy_mgmt.request_category,
    p_source_channel policy_mgmt.source_channel,
    p_submitted_by BIGINT,
    p_request_payload JSONB
)
RETURNS BIGINT
LANGUAGE plpgsql
AS $$
DECLARE
    v_request_id BIGINT;
BEGIN

    SELECT request_id
    INTO v_request_id
    FROM policy_mgmt.request_idempotency_registry
    WHERE idempotency_key = p_idempotency_key;

    IF v_request_id IS NOT NULL THEN
        RETURN v_request_id;
    END IF;

    v_request_id := nextval('policy_mgmt.seq_service_request_id');

    INSERT INTO policy_mgmt.request_idempotency_registry(idempotency_key, request_id)
    VALUES (p_idempotency_key, v_request_id)
    ON CONFLICT (idempotency_key) DO NOTHING;

    INSERT INTO policy_mgmt.service_request(
        request_id,
        policy_id,
        policy_number,
        request_type,
        request_category,
        source_channel,
        submitted_by,
        request_payload
    )
    VALUES (
        v_request_id,
        p_policy_id,
        p_policy_number,
        p_request_type,
        p_request_category,
        p_source_channel,
        p_submitted_by,
        p_request_payload
    );

    RETURN v_request_id;

END;
$$;