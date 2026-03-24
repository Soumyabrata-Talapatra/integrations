-- +migrate Up
-- +migrate Down

-- workflow_signals_log table
CREATE TABLE revival.workflow_signals_log (
    log_id VARCHAR PRIMARY KEY,
    request_id VARCHAR NOT NULL,
    workflow_id VARCHAR NOT NULL,
    signal_name VARCHAR NOT NULL,
    signal_payload JSONB,
    sent_at TIMESTAMP NOT NULL DEFAULT NOW(),
    sent_by VARCHAR NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_signal_request ON revival.workflow_signals_log(request_id);
CREATE INDEX idx_signal_name ON revival.workflow_signals_log(signal_name);
CREATE INDEX idx_signal_sent_at ON revival.workflow_signals_log(sent_at);

-- +migrate Down
DROP TABLE IF EXISTS revival.workflow_signals_log;
