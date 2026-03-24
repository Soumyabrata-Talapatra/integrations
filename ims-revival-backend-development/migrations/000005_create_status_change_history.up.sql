-- +migrate Up
-- +migrate Down

-- status_change_history table
CREATE TABLE revival.status_change_history (
    history_id VARCHAR PRIMARY KEY,
    request_id VARCHAR NOT NULL,
    from_status VARCHAR,
    to_status VARCHAR NOT NULL,
    changed_at TIMESTAMP NOT NULL DEFAULT NOW(),
    changed_by VARCHAR NOT NULL,
    change_reason VARCHAR
);

CREATE INDEX idx_status_history_request ON revival.status_change_history(request_id);
CREATE INDEX idx_status_history_changed_at ON revival.status_change_history(changed_at);

-- +migrate Down
DROP TABLE IF EXISTS revival.status_change_history;
