-- +migrate Up
-- +migrate Down

-- document_references table (links revival_requests to documents)
CREATE TABLE revival.document_references (
    reference_id VARCHAR PRIMARY KEY,
    request_id VARCHAR NOT NULL,
    document_type VARCHAR NOT NULL,
    document_id VARCHAR NOT NULL,
    document_name VARCHAR,
    document_path TEXT,
    received_date TIMESTAMP,
    received_by VARCHAR,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ref_request ON revival.document_references(request_id);
CREATE INDEX idx_ref_type ON revival.document_references(document_type);

-- +migrate Down
DROP TABLE IF EXISTS revival.document_references;
