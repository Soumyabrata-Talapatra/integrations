-- +migrate Up
-- +migrate Down

-- payment_receipts table
CREATE TABLE collection.payment_receipts (
    receipt_id VARCHAR PRIMARY KEY,
    receipt_number VARCHAR UNIQUE NOT NULL,
    request_id VARCHAR NOT NULL,
    policy_number VARCHAR NOT NULL,
    is_combined_receipt BOOLEAN NOT NULL DEFAULT false,
    payment_ids VARCHAR[] NOT NULL,
    total_amount DECIMAL NOT NULL,
    receipt_date DATE NOT NULL,
    payment_mode VARCHAR NOT NULL,
    document_path TEXT,
    generated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    generated_by VARCHAR,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_receipt_request ON collection.payment_receipts(request_id);
CREATE INDEX idx_receipt_number ON collection.payment_receipts(receipt_number);
CREATE INDEX idx_receipt_date ON collection.payment_receipts(receipt_date);

-- +migrate Down
DROP TABLE IF EXISTS collection.payment_receipts;
