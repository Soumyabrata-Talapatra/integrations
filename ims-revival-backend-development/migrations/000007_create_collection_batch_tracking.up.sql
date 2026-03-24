-- +migrate Up
-- +migrate Down

-- collection_batch_tracking table
CREATE TABLE collection.collection_batch_tracking (
    batch_id VARCHAR PRIMARY KEY,
    request_id VARCHAR NOT NULL,
    policy_number VARCHAR NOT NULL,
    premium_payment_id VARCHAR NOT NULL,
    installment_payment_id VARCHAR NOT NULL,
    collection_complete BOOLEAN NOT NULL DEFAULT false,
    collection_date DATE NOT NULL,
    combined_receipt_id VARCHAR,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP
);

CREATE INDEX idx_batch_request ON collection.collection_batch_tracking(request_id);
CREATE INDEX idx_batch_complete ON collection.collection_batch_tracking(collection_complete);

-- +migrate Down
DROP TABLE IF EXISTS collection.collection_batch_tracking;
