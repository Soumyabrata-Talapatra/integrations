-- +migrate Up
-- +migrate Down

-- generated_letters table
CREATE TABLE revival.generated_letters (
    letter_id VARCHAR PRIMARY KEY,
    request_id VARCHAR NOT NULL,
    policy_number VARCHAR NOT NULL,
    letter_type VARCHAR NOT NULL,
    document_path TEXT,
    document_hash VARCHAR,
    generated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    generated_by VARCHAR,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_letter_request ON revival.generated_letters(request_id);
CREATE INDEX idx_letter_type ON revival.generated_letters(letter_type);

-- +migrate Down
DROP TABLE IF EXISTS revival.generated_letters;
