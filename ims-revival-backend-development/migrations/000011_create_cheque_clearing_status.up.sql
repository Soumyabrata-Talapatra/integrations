-- +migrate Up
-- +migrate Down

-- cheque_clearing_status table
CREATE TABLE collection.cheque_clearing_status (
    cheque_id VARCHAR PRIMARY KEY,
    payment_id VARCHAR NOT NULL,
    request_id VARCHAR NOT NULL,
    policy_number VARCHAR NOT NULL,
    cheque_number VARCHAR NOT NULL,
    bank_name VARCHAR NOT NULL,
    cheque_date DATE NOT NULL,
    amount DECIMAL NOT NULL,
    clearance_status VARCHAR NOT NULL,
    next_due_date DATE,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_cheque_payment ON collection.cheque_clearing_status(payment_id);
CREATE INDEX idx_cheque_status ON collection.cheque_clearing_status(clearance_status);
CREATE INDEX idx_cheque_policy ON collection.cheque_clearing_status(policy_number);

-- +migrate Down
DROP TABLE IF EXISTS collection.cheque_clearing_status;
