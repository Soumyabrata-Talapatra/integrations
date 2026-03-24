

-- Insert config data
INSERT INTO common.system_configuration (config_key, config_value, is_configurable, description)
VALUES ('max_revivals_allowed', '2', true, 'Maximum number of revivals allowed per policy')
ON CONFLICT (config_key) DO NOTHING;

-- Insert test policies
-- Policy 1: Eligible (Lapsed, no ongoing, under limit)
INSERT INTO common.policies (
    policy_number, customer_id, customer_name, product_code, product_name,
    policy_status, premium_frequency, premium_amount, sum_assured,
    paid_to_date, maturity_date, date_of_commencement,
    revival_count, last_revival_date,billing_method, office
) VALUES (
    '0000000000001', 'CUST0000000001', 'John Doe', 'TERM001', 'Term Life Insurance',
    'AL', 'ANNUAL', 50000.00, 1000000.00,
    '2024-01-01', '2045-01-01', '2020-01-01',
    0, NULL, 'DIRECT_DEBIT', 'HEAD_OFFICE'
) ON CONFLICT (policy_number) DO UPDATE SET
    policy_status = 'AL',
    revival_count = 0,
    last_revival_date = NULL;

-- Policy 2: Not eligible (In Force)
INSERT INTO common.policies (
    policy_number, customer_id, customer_name, product_code, product_name,
    policy_status, premium_frequency, premium_amount, sum_assured,
    paid_to_date, maturity_date, date_of_commencement,
    revival_count, last_revival_date ,billing_method, office
) VALUES (
    '0000000000002', 'CUST0000000002', 'Jane Smith', 'TERM001', 'Term Life Insurance',
    'IF', 'MONTHLY', 5000.00, 500000.00,
    '2024-12-01', '2045-12-01', '2020-12-01',
    0, NULL, 'CHEQUE', 'BRANCH_OFFICE'
) ON CONFLICT (policy_number) DO UPDATE SET
    policy_status = 'IF',
    revival_count = 0;

-- Policy 3: Not eligible (Max revivals reached)
INSERT INTO common.policies (
    policy_number, customer_id, customer_name, product_code, product_name,
    policy_status, premium_frequency, premium_amount, sum_assured,
    paid_to_date, maturity_date, date_of_commencement,
    revival_count, last_revival_date, billing_method, office
) VALUES (
    '0000000000003', 'CUST0000000003', 'Bob Johnson', 'ENDOW001', 'Endowment Policy',
    'AL', 'QUARTERLY', 15000.00, 750000.00,
    '2023-01-01', '2043-01-01', '2018-01-01',
    2, '2024-06-15',  'STANDING_INSTRUCTION', 'REGIONAL_OFFICE'
) ON CONFLICT (policy_number) DO UPDATE SET
    policy_status = 'AL',
    revival_count = 2,
    last_revival_date = '2024-06-15';

-- Policy 4: Not eligible (Has ongoing revival)
INSERT INTO common.policies (
    policy_number, customer_id, customer_name, product_code, product_name,
    policy_status, premium_frequency, premium_amount, sum_assured,
    paid_to_date, maturity_date, date_of_commencement,
    revival_count, last_revival_date, billing_method, office
) VALUES (
    '0000000000004', 'CUST0000000004', 'Alice Williams', 'TERM002', 'Whole Life Insurance',
    'AL', 'HALF_YEARLY', 25000.00, 1500000.00,
    '2023-06-01', '2053-06-01', '2019-06-01',
    1, '2023-12-20', 'CASH', 'SATELLITE_OFFICE'
) ON CONFLICT (policy_number) DO UPDATE SET
    policy_status = 'AL',
    revival_count = 1,
    last_revival_date = '2023-12-20';

SELECT
    policy_number,
    customer_name,
    policy_status,
    revival_count,
    (SELECT COUNT(*) FROM revival.revival_requests WHERE policy_number = p.policy_number AND current_status NOT IN ('COMPLETED', 'WITHDRAWN', 'TERMINATED', 'REJECTED')) as ongoing_revivals,
    CASE
        WHEN policy_status = 'AL'
             AND revival_count < 2
             AND (SELECT COUNT(*) FROM revival.revival_requests WHERE policy_number = p.policy_number AND current_status NOT IN ('COMPLETED', 'WITHDRAWN', 'TERMINATED', 'REJECTED')) = 0
        THEN 'ELIGIBLE ✓'
        ELSE 'NOT ELIGIBLE ✗'
    END as eligibility_status
FROM common.policies p
WHERE policy_number IN ('0000000000001', '0000000000002', '0000000000003', '0000000000004')
ORDER BY policy_number;
