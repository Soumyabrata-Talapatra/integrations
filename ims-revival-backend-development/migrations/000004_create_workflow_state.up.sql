-- +migrate Up
-- +migrate Down

-- revival_request_workflow_state table
CREATE TABLE revival.revival_request_workflow_state (
    request_id VARCHAR PRIMARY KEY,
    workflow_id VARCHAR NOT NULL,
    run_id VARCHAR NOT NULL,
    current_status VARCHAR NOT NULL,
    workflow_status VARCHAR NOT NULL DEFAULT 'RUNNING',
    sla_start_date TIMESTAMP,
    sla_end_date TIMESTAMP,
    sla_expired BOOLEAN NOT NULL DEFAULT false,
    first_collection_done BOOLEAN NOT NULL DEFAULT false,
    total_installments INTEGER,
    installments_paid INTEGER NOT NULL DEFAULT 0,
    started_at TIMESTAMP NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP,
    last_updated TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_workflow_state_status ON revival.revival_request_workflow_state(current_status);

-- +migrate Down
DROP TABLE IF EXISTS revival.revival_request_workflow_state;
