-- +migrate Up
-- +migrate Down

-- workflow_executions table
CREATE TABLE revival.workflow_executions (
    execution_id VARCHAR PRIMARY KEY,
    request_id VARCHAR NOT NULL,
    workflow_type VARCHAR NOT NULL,
    workflow_id VARCHAR NOT NULL,
    run_id VARCHAR,
    started_at TIMESTAMP NOT NULL,
    completed_at TIMESTAMP,
    status VARCHAR NOT NULL,
    parent_workflow_id VARCHAR,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_execution_request ON revival.workflow_executions(request_id);
CREATE INDEX idx_execution_type ON revival.workflow_executions(workflow_type);
CREATE INDEX idx_execution_status ON revival.workflow_executions(status);
CREATE INDEX idx_execution_started_at ON revival.workflow_executions(started_at);

-- +migrate Down
DROP TABLE IF EXISTS revival.workflow_executions;
