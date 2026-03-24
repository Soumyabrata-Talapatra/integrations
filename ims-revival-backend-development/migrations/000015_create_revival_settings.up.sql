-- +migrate Up
-- +migrate Down

-- revival_settings table (IR_29 configuration)
CREATE TABLE revival.revival_settings (
    setting_key VARCHAR PRIMARY KEY,
    setting_value VARCHAR NOT NULL,
    is_configurable BOOLEAN NOT NULL DEFAULT false,
    description VARCHAR,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Insert default setting
INSERT INTO revival.revival_settings (setting_key, setting_value, is_configurable, description) VALUES 
    ('max_revivals_allowed', '2', true, 'Maximum number of revivals allowed per policy (IR_29)');

-- Insert fixed (non-configurable) settings for reference
INSERT INTO revival.revival_settings (setting_key, setting_value, is_configurable, description) VALUES
    ('max_installments', '12', false, 'Maximum number of installments allowed (IR_4)'),
    ('grace_period_days', '0', false, 'Zero grace period for installments (IR_9)'),
    ('sla_days', '60', false, '60-day SLA from approval date (IR_10)'),
    ('due_date_rule', 'first_of_next_month', false, 'Subsequent installments due on 1st of month (IR_11)');

CREATE INDEX idx_settings_key ON revival.revival_settings(setting_key);

-- +migrate Down
DROP TABLE IF EXISTS revival.revival_settings;
