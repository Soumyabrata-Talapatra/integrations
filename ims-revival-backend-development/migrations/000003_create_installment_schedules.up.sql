-- +migrate Up
-- +migrate Down

-- installment_schedules table
CREATE TABLE revival.installment_schedules (
    schedule_id VARCHAR PRIMARY KEY,
    request_id VARCHAR NOT NULL,
    policy_number VARCHAR NOT NULL,
    installment_number INTEGER NOT NULL,
    installment_amount DECIMAL NOT NULL,
    tax_amount DECIMAL NOT NULL,
    total_amount DECIMAL NOT NULL,
    due_date DATE NOT NULL,
    payment_date TIMESTAMP,
    is_paid BOOLEAN NOT NULL DEFAULT false,
    grace_period_days INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_due_date_positive CHECK (grace_period_days = 0)
);

CREATE INDEX idx_install_schedule_request ON revival.installment_schedules(request_id);
CREATE INDEX idx_install_schedule_policy ON revival.installment_schedules(policy_number);
CREATE INDEX idx_install_schedule_due ON revival.installment_schedules(due_date);

-- +migrate Down
DROP TABLE IF EXISTS revival.installment_schedules;
