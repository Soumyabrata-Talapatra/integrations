-- Migration for Revival Tables (Test Environment)
-- Creates schemas and tables needed for revival workflow testing

-- Create schemas
CREATE SCHEMA IF NOT EXISTS common;
CREATE SCHEMA IF NOT EXISTS revival;

-- =============================================================================
-- COMMON SCHEMA TABLES
-- =============================================================================

-- Policies table (CORRECTED: without billing_method and office columns)
CREATE TABLE IF NOT EXISTS common.policies (
    policy_number VARCHAR(13) PRIMARY KEY,
    customer_id VARCHAR(50) NOT NULL,
    customer_name VARCHAR(200) NOT NULL,
    product_code VARCHAR(20) NOT NULL,
    product_name VARCHAR(100) NOT NULL,
    policy_status VARCHAR(10) NOT NULL,
    premium_frequency VARCHAR(20) NOT NULL,
    premium_amount NUMERIC(15,2) NOT NULL,
    sum_assured NUMERIC(15,2) NOT NULL,
    paid_to_date DATE,
    maturity_date DATE NOT NULL,
    date_of_commencement DATE NOT NULL,
    revival_count INTEGER NOT NULL DEFAULT 0,
    last_revival_date DATE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_policy_number CHECK (policy_number ~ '^\d{13}$')
);

CREATE INDEX IF NOT EXISTS idx_policy_status ON common.policies(policy_status);
CREATE INDEX IF NOT EXISTS idx_policy_customer ON common.policies(customer_id);

-- System Configuration table
CREATE TABLE IF NOT EXISTS common.system_configuration (
    config_key VARCHAR PRIMARY KEY,
    config_value VARCHAR NOT NULL,
    is_configurable BOOLEAN NOT NULL DEFAULT false,
    description VARCHAR,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_config_key ON common.system_configuration(config_key);

-- Insert default configuration
INSERT INTO common.system_configuration (config_key, config_value, is_configurable, description)
VALUES ('max_revivals_allowed', '2', true, 'Maximum number of revivals allowed per policy (IR_29)')
ON CONFLICT (config_key) DO NOTHING;

-- =============================================================================
-- REVIVAL SCHEMA TABLES
-- =============================================================================

-- Revival Requests table
CREATE TABLE IF NOT EXISTS revival.revival_requests (
    request_id VARCHAR PRIMARY KEY,
    ticket_id VARCHAR UNIQUE NOT NULL,
    policy_number VARCHAR(13) NOT NULL,
    request_type VARCHAR NOT NULL,
    current_status VARCHAR NOT NULL,
    workflow_id VARCHAR,
    run_id VARCHAR,
    indexed_date TIMESTAMP,
    indexed_by VARCHAR,
    data_entry_date TIMESTAMP,
    data_entry_by VARCHAR,
    qc_complete_date TIMESTAMP,
    qc_by VARCHAR,
    approval_date TIMESTAMP,
    approved_by VARCHAR,
    completion_date TIMESTAMP,
    termination_date TIMESTAMP,
    withdrawal_date TIMESTAMP,
    number_of_installments INTEGER NOT NULL DEFAULT 0,
    revival_amount NUMERIC(15,2) NOT NULL DEFAULT 0,
    installment_amount NUMERIC(15,2) NOT NULL DEFAULT 0,
    total_tax_on_unpaid NUMERIC(15,2) NOT NULL DEFAULT 0,
    first_collection_date TIMESTAMP,
    first_collection_done BOOLEAN NOT NULL DEFAULT false,
    blocking_new_collections BOOLEAN NOT NULL DEFAULT true,
    installments_paid INTEGER NOT NULL DEFAULT 0,
    request_owner VARCHAR,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_installments CHECK (number_of_installments >= 0),
    CONSTRAINT chk_ticket_format CHECK (ticket_id ~ '^PSREYV[a-f0-9-]{12}$')
);

CREATE INDEX IF NOT EXISTS idx_revival_policy_number ON revival.revival_requests(policy_number);
CREATE INDEX IF NOT EXISTS idx_revival_current_status ON revival.revival_requests(current_status);
CREATE INDEX IF NOT EXISTS idx_revival_ticket_id ON revival.revival_requests(ticket_id);

-- =============================================================================
-- TEST DATA
-- =============================================================================

-- Test Policy 1: Eligible for revival (Lapsed, no ongoing revival)
INSERT INTO common.policies (
    policy_number, customer_id, customer_name, product_code, product_name,
    policy_status, premium_frequency, premium_amount, sum_assured,
    paid_to_date, maturity_date, date_of_commencement,
    revival_count, last_revival_date
) VALUES (
    '0000000000001',
    'CUST0000000001',
    'John Doe',
    'TERM001',
    'Term Life Insurance',
    'AL',  -- Lapsed status
    'ANNUAL',
    50000.00,
    1000000.00,
    '2024-01-01',
    '2045-01-01',
    '2020-01-01',
    0,  -- No previous revivals
    NULL
) ON CONFLICT (policy_number) DO NOTHING;

-- Test Policy 2: Not eligible - In Force status
INSERT INTO common.policies (
    policy_number, customer_id, customer_name, product_code, product_name,
    policy_status, premium_frequency, premium_amount, sum_assured,
    paid_to_date, maturity_date, date_of_commencement,
    revival_count, last_revival_date
) VALUES (
    '0000000000002',
    'CUST0000000002',
    'Jane Smith',
    'TERM001',
    'Term Life Insurance',
    'IF',  -- In Force - not lapsed
    'MONTHLY',
    5000.00,
    500000.00,
    '2024-12-01',
    '2045-12-01',
    '2020-12-01',
    0,
    NULL
) ON CONFLICT (policy_number) DO NOTHING;

-- Test Policy 3: Not eligible - Max revivals reached
INSERT INTO common.policies (
    policy_number, customer_id, customer_name, product_code, product_name,
    policy_status, premium_frequency, premium_amount, sum_assured,
    paid_to_date, maturity_date, date_of_commencement,
    revival_count, last_revival_date
) VALUES (
    '0000000000003',
    'CUST0000000003',
    'Bob Johnson',
    'ENDOW001',
    'Endowment Policy',
    'AL',  -- Lapsed
    'QUARTERLY',
    15000.00,
    750000.00,
    '2023-01-01',
    '2043-01-01',
    '2018-01-01',
    2,  -- Already revived 2 times (max allowed)
    '2024-06-15'
) ON CONFLICT (policy_number) DO NOTHING;

-- Test Policy 4: Not eligible - Has ongoing revival request
INSERT INTO common.policies (
    policy_number, customer_id, customer_name, product_code, product_name,
    policy_status, premium_frequency, premium_amount, sum_assured,
    paid_to_date, maturity_date, date_of_commencement,
    revival_count, last_revival_date
) VALUES (
    '0000000000004',
    'CUST0000000004',
    'Alice Williams',
    'TERM002',
    'Whole Life Insurance',
    'AL',  -- Lapsed
    'HALF_YEARLY',
    25000.00,
    1500000.00,
    '2023-06-01',
    '2053-06-01',
    '2019-06-01',
    1,
    '2023-12-20'
) ON CONFLICT (policy_number) DO NOTHING;

-- Ongoing revival request for policy 4
INSERT INTO revival.revival_requests (
    request_id, ticket_id, policy_number, request_type, current_status,
    indexed_date, indexed_by, number_of_installments,
    revival_amount, installment_amount, total_tax_on_unpaid
) VALUES (
    'REQ000000000001',
    'PSREYVab12-34cd-56',
    '0000000000004',
    'INSTALLMENT_REVIVAL',
    'INDEXED',  -- Ongoing request
    NOW(),
    'TEST_USER',
    0,
    0.00,
    0.00,
    0.00
) ON CONFLICT (request_id) DO NOTHING;

-- Completed revival request for policy 3 (first revival)
INSERT INTO revival.revival_requests (
    request_id, ticket_id, policy_number, request_type, current_status,
    indexed_date, indexed_by, completion_date,
    number_of_installments, revival_amount, installment_amount, total_tax_on_unpaid
) VALUES (
    'REQ000000000002',
    'PSREYVef78-90ab-12',
    '0000000000003',
    'INSTALLMENT_REVIVAL',
    'COMPLETED',  -- Completed - won't count as ongoing
    '2023-01-15',
    'TEST_USER',
    '2023-03-20',
    6,
    100000.00,
    16666.67,
    5000.00
) ON CONFLICT (request_id) DO NOTHING;
