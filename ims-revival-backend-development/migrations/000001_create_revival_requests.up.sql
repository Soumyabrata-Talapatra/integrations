-- +migrate Up
-- +migrate Down

-- revival_requests table
CREATE TABLE revival.revival_requests (
    request_id VARCHAR PRIMARY KEY,
    ticket_id VARCHAR UNIQUE NOT NULL,
    policy_number VARCHAR NOT NULL,
    request_type VARCHAR NOT NULL,
    current_status VARCHAR NOT NULL,
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
    number_of_installments INTEGER NOT NULL,
    revival_amount DECIMAL NOT NULL,
    installment_amount DECIMAL NOT NULL,
    total_tax_on_unpaid DECIMAL NOT NULL,
    first_collection_date TIMESTAMP,
    first_collection_done BOOLEAN NOT NULL DEFAULT false,
    blocking_new_collections BOOLEAN NOT NULL DEFAULT true,
    installments_paid INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Indexes for revival_requests
CREATE INDEX idx_revival_policy_number ON revival.revival_requests(policy_number);
CREATE INDEX idx_revival_current_status ON revival.revival_requests(current_status);
CREATE INDEX idx_revival_ticket_id ON revival.revival_requests(ticket_id);

-- +migrate Down
DROP TABLE IF EXISTS revival.revival_requests;
