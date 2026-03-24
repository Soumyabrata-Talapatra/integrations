-- +migrate Up
-- +migrate Down

-- policies table (common schema)
CREATE TABLE common.policies (
    policy_number VARCHAR PRIMARY KEY,
    customer_id VARCHAR NOT NULL,
    customer_name VARCHAR NOT NULL,
    product_code VARCHAR NOT NULL,
    product_name VARCHAR NOT NULL,
    policy_status VARCHAR NOT NULL,
    premium_frequency VARCHAR NOT NULL,
    premium_amount DECIMAL NOT NULL,
    sum_assured DECIMAL NOT NULL,
    paid_to_date DATE,
    maturity_date DATE NOT NULL,
    date_of_commencement DATE NOT NULL,
    billing_method VARCHAR NOT NULL,
    revival_count INTEGER NOT NULL DEFAULT 0,
    last_revival_date DATE,
    office VARCHAR NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_policy_status ON common.policies(policy_status);
CREATE INDEX idx_policy_customer ON common.policies(customer_id);

-- +migrate Down
DROP TABLE IF EXISTS common.policies;
