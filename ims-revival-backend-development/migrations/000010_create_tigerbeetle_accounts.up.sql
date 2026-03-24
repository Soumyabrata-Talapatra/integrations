-- +migrate Up
-- +migrate Down

-- tigerbeetle_accounts table
CREATE TABLE common.tigerbeetle_accounts (
    account_id VARCHAR PRIMARY KEY,
    policy_number VARCHAR UNIQUE NOT NULL,
    premium_account_id VARCHAR NOT NULL,
    revival_account_id VARCHAR NOT NULL,
    loan_account_id VARCHAR NOT NULL,
    combined_suspense_account_id VARCHAR NOT NULL,
    revival_suspense_account_id VARCHAR NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_tb_policy ON common.tigerbeetle_accounts(policy_number);

-- +migrate Down
DROP TABLE IF EXISTS common.tigerbeetle_accounts;
