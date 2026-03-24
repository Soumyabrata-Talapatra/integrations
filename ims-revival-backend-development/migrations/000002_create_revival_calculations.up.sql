-- +migrate Up
-- +migrate Down

-- revival_calculations table
CREATE TABLE revival.revival_calculations (
    calculation_id VARCHAR PRIMARY KEY,
    request_id VARCHAR NOT NULL,
    policy_number VARCHAR NOT NULL,
    date_of_revival DATE NOT NULL,
    number_of_installments INTEGER NOT NULL,
    unpaid_premium_months INTEGER NOT NULL,
    interest_rate DECIMAL,
    monthly_premium DECIMAL,
    premium_amount DECIMAL NOT NULL,
    tax_on_premium DECIMAL NOT NULL,
    total_renewal_amount DECIMAL NOT NULL,
    installment_amount DECIMAL NOT NULL,
    tax_on_unpaid_premium DECIMAL NOT NULL,
    total_installment_amount DECIMAL NOT NULL,
    grand_total_first_collection DECIMAL NOT NULL,
    valid_until TIMESTAMP,
    calculated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_revival_calc_request_id ON revival.revival_calculations(request_id);
CREATE INDEX idx_revival_calc_policy ON revival.revival_calculations(policy_number);

-- +migrate Down
DROP TABLE IF EXISTS revival.revival_calculations;
