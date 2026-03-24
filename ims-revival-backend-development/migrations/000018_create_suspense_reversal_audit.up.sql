-- +migrate Up
-- +migrate Down

-- suspense_reversal_audit table (IR_28 enforcement)
CREATE TABLE collection.suspense_reversal_audit (
    suspense_id VARCHAR NOT NULL,
    reversed_at TIMESTAMP NOT NULL DEFAULT NOW(),
    authorized_by VARCHAR NOT NULL,
    reversal_reason VARCHAR NOT NULL,
    amount_reversed DECIMAL NOT NULL,
    is_first_collection BOOLEAN NOT NULL,
    reversal_allowed BOOLEAN NOT NULL,
    rejection_reason VARCHAR,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_suspense ON collection.suspense_reversal_audit(suspense_id);
CREATE INDEX idx_audit_reversed_at ON collection.suspense_reversal_audit(reversed_at);

-- +migrate Down
DROP TABLE IF EXISTS collection.suspense_reversal_audit;
