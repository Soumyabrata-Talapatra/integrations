-- +migrate Up
-- +migrate Down

-- revival_settings_audit table
CREATE TABLE revival.revival_settings_audit (
    audit_id VARCHAR PRIMARY KEY,
    setting_key VARCHAR NOT NULL,
    old_value VARCHAR NOT NULL,
    new_value VARCHAR NOT NULL,
    changed_by VARCHAR NOT NULL,
    changed_reason VARCHAR,
    changed_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_audit_key ON revival.revival_settings_audit(setting_key);
CREATE INDEX idx_audit_changed_at ON revival.revival_settings_audit(changed_at);

-- +migrate Down
DROP TABLE IF EXISTS revival.revival_settings_audit;
