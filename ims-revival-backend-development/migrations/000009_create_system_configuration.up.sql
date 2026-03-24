-- +migrate Up
-- +migrate Down

-- system_configuration table
CREATE TABLE common.system_configuration (
    config_key VARCHAR PRIMARY KEY,
    config_value VARCHAR NOT NULL,
    is_configurable BOOLEAN NOT NULL DEFAULT false,
    description VARCHAR,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Insert default configuration
INSERT INTO common.system_configuration (config_key, config_value, is_configurable, description) VALUES
 
    ('max_revivals_allowed', '2', true, 'Maximum number of revivals allowed per policy (IR_29)');

CREATE INDEX idx_config_key ON common.system_configuration(config_key);

-- +migrate Down
DROP TABLE IF EXISTS common.system_configuration;
