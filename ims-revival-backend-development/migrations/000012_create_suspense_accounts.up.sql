-- +migrate Up
-- +migrate Down

-- suspense_accounts table
CREATE TABLE collection.suspense_accounts (
    suspense_id VARCHAR PRIMARY KEY,
    policy_number VARCHAR NOT NULL,
    request_id VARCHAR,
    suspense_type VARCHAR NOT NULL,
    amount DECIMAL NOT NULL,
    is_reversed BOOLEAN NOT NULL DEFAULT false,
    reversal_date TIMESTAMP,
    reversal_authorized_by VARCHAR,
    reversal_reason VARCHAR,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    created_by VARCHAR,
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_suspense_policy ON collection.suspense_accounts(policy_number);
CREATE INDEX idx_suspense_request ON collection.suspense_accounts(request_id);
CREATE INDEX idx_suspense_type ON collection.suspense_accounts(suspense_type);
CREATE INDEX idx_suspense_reversed ON collection.suspense_accounts(is_reversed);

-- +migrate Down
DROP TABLE IF EXISTS collection.suspense_accounts;
