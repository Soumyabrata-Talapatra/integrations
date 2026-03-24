-- +migrate Up
-- +migrate Down

-- request_stage_history table
CREATE TABLE revival.request_stage_history (
    history_id VARCHAR PRIMARY KEY,
    request_id VARCHAR NOT NULL,
    stage_name VARCHAR NOT NULL,
    stage_action VARCHAR NOT NULL,
    comments TEXT,
    action_at TIMESTAMP NOT NULL DEFAULT NOW(),
    action_by VARCHAR NOT NULL,
    missing_documents TEXT[]
);

CREATE INDEX idx_stage_history_request ON revival.request_stage_history(request_id);
CREATE INDEX idx_stage_history_stage ON revival.request_stage_history(stage_name);
CREATE INDEX idx_stage_history_action_at ON revival.request_stage_history(action_at);

-- +migrate Down
DROP TABLE IF EXISTS revival.request_stage_history;
