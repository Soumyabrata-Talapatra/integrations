-- +migrate Up
-- +migrate Down

-- configuration_audit table
CREATE TABLE common.configuration_audit (
    audit_id VARCHAR PRIMARY KEY,
    config_key VARCHAR NOT NULL,
    old_value VARCHAR NOT NULL,
    new_value VARCHAR NOT NULL,
    changed_by VARCHAR NOT NULL,
    changed_reason VARCHAR,
    affected_policies_count INTEGER,
    changed_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_config_key ON common.configuration_audit(config_key);
CREATE INDEX idx_audit_changed_at ON common.configuration_audit(changed_at);

-- +migrate Down
DROP TABLE IF EXISTS common.configuration_audit;
