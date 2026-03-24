-- +migrate Up
-- +migrate Down

-- payment_transactions table
CREATE TABLE collection.payment_transactions (
    payment_id VARCHAR PRIMARY KEY,
    request_id VARCHAR NOT NULL,
    policy_number VARCHAR NOT NULL,
    collection_batch_id VARCHAR NOT NULL,
    linked_payment_id VARCHAR,
    cheque_id VARCHAR,
    payment_type VARCHAR NOT NULL,
    installment_number INTEGER,
    amount DECIMAL NOT NULL,
    tax_amount DECIMAL NOT NULL,
    total_amount DECIMAL NOT NULL,
    payment_mode VARCHAR NOT NULL,
    payment_status VARCHAR NOT NULL,
    collection_date DATE NOT NULL,
    payment_date TIMESTAMP,
    receipt_id VARCHAR,
    tigerbeetle_transfer_id VARCHAR,
    collected_by VARCHAR,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_payment_request ON collection.payment_transactions(request_id);
CREATE INDEX idx_payment_batch ON collection.payment_transactions(collection_batch_id);
CREATE INDEX idx_payment_cheque ON collection.payment_transactions(cheque_id);
CREATE INDEX idx_payment_status ON collection.payment_transactions(payment_status);
CREATE INDEX idx_payment_collection_date ON collection.payment_transactions(collection_date);

-- +migrate Down
DROP TABLE IF EXISTS collection.payment_transactions;
