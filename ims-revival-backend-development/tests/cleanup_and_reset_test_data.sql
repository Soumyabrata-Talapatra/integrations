-- Cleanup and Reset Test Data for Policy Validation Tests
-- This script removes all test data and resets to a clean state

-- ============================================================================
-- STEP 1: CLEANUP - Remove all test data
-- ============================================================================

-- Remove test revival requests by ticket_id (more reliable)
DELETE FROM revival.revival_requests
WHERE ticket_id IN (
    'PSREYVab12-34cd-56',
    'PSREYVef78-90gh-12'
);

-- Remove any other revival requests for test policies
DELETE FROM revival.revival_requests
WHERE policy_number IN (
    '0000000000001',
    '0000000000002',
    '0000000000003',
    '0000000000004'
);

-- Remove test policies
DELETE FROM common.policies
WHERE policy_number IN (
    '0000000000001',
    '0000000000002',
    '0000000000003',
    '0000000000004'
);

-- Verify cleanup
SELECT 'Cleanup complete - all test data removed' as status;

-- ============================================================================
-- STEP 2: RESET - Insert fresh test data
-- ============================================================================

-- Insert config data
INSERT INTO common.system_configuration (config_key, config_value, is_configurable, description)
VALUES ('max_revivals_allowed', '2', true, 'Maximum number of revivals allowed per policy')
ON CONFLICT (config_key) DO UPDATE SET
    config_value = '2',
    is_configurable = true;

-- ============================================================================
-- TEST POLICY 1: ELIGIBLE (Lapsed, No Ongoing, Under Limit)
-- ============================================================================
INSERT INTO common.policies (
    policy_number, customer_id, customer_name, product_code, product_name,
    policy_status, premium_frequency, premium_amount, sum_assured,
    paid_to_date, maturity_date, date_of_commencement,
    revival_count, last_revival_date
) VALUES (
    '0000000000001',
    'CUST0000000001',
    'John Doe',
    'TERM001',
    'Term Life Insurance',
    'AL',           -- Lapsed status
    'ANNUAL',
    50000.00,
    1000000.00,
    '2024-01-01',
    '2045-01-01',
    '2020-01-01',
    0,              -- No previous revivals
    NULL            -- Never revived
);

-- ============================================================================
-- TEST POLICY 2: NOT ELIGIBLE (In Force Status)
-- ============================================================================
INSERT INTO common.policies (
    policy_number, customer_id, customer_name, product_code, product_name,
    policy_status, premium_frequency, premium_amount, sum_assured,
    paid_to_date, maturity_date, date_of_commencement,
    revival_count, last_revival_date
) VALUES (
    '0000000000002',
    'CUST0000000002',
    'Jane Smith',
    'TERM001',
    'Term Life Insurance',
    'IF',           -- In Force (NOT lapsed) - this is the key difference
    'MONTHLY',
    5000.00,
    500000.00,
    '2024-12-01',
    '2045-12-01',
    '2020-12-01',
    0,              -- No revivals
    NULL
);

-- ============================================================================
-- TEST POLICY 3: NOT ELIGIBLE (Max Revivals Reached)
-- ============================================================================
INSERT INTO common.policies (
    policy_number, customer_id, customer_name, product_code, product_name,
    policy_status, premium_frequency, premium_amount, sum_assured,
    paid_to_date, maturity_date, date_of_commencement,
    revival_count, last_revival_date
) VALUES (
    '0000000000003',
    'CUST0000000003',
    'Bob Johnson',
    'ENDOW001',
    'Endowment Policy',
    'AL',           -- Lapsed
    'QUARTERLY',
    15000.00,
    750000.00,
    '2023-01-01',
    '2043-01-01',
    '2018-01-01',
    2,              -- Already revived 2 times (max allowed)
    '2024-06-15'
);

-- ============================================================================
-- TEST POLICY 4: NOT ELIGIBLE (Has Ongoing Revival)
-- ============================================================================
INSERT INTO common.policies (
    policy_number, customer_id, customer_name, product_code, product_name,
    policy_status, premium_frequency, premium_amount, sum_assured,
    paid_to_date, maturity_date, date_of_commencement,
    revival_count, last_revival_date
) VALUES (
    '0000000000004',
    'CUST0000000004',
    'Alice Williams',
    'TERM002',
    'Whole Life Insurance',
    'AL',           -- Lapsed
    'HALF_YEARLY',
    25000.00,
    1500000.00,
    '2023-06-01',
    '2053-06-01',
    '2019-06-01',
    1,              -- Revived once before
    '2023-12-20'
);

-- Create ongoing revival request for policy 4 ONLY
INSERT INTO revival.revival_requests (
    ticket_id, policy_number, request_type, current_status,
    indexed_date, indexed_by, number_of_installments,
    revival_amount, installment_amount, total_tax_on_unpaid
) VALUES (
    'PSREYVab12-34cd-56',
    '0000000000004',      -- Only for policy 4
    'installment_revival',
    'INDEXED',            -- Ongoing status
    NOW(),
    'TEST_USER',
    0,
    0.00,
    0.00,
    0.00
);

-- ============================================================================
-- VERIFICATION
-- ============================================================================

SELECT 'Test data reset complete!' as status;
SELECT '' as separator;

-- Show all test policies with their eligibility
SELECT
    p.policy_number,
    p.customer_name,
    p.policy_status,
    p.revival_count,
    (SELECT COUNT(*)
     FROM revival.revival_requests r
     WHERE r.policy_number = p.policy_number
       AND r.current_status NOT IN ('COMPLETED', 'WITHDRAWN', 'TERMINATED', 'REJECTED')
    ) as ongoing_revivals,
    (SELECT config_value::int
     FROM common.system_configuration
     WHERE config_key = 'max_revivals_allowed'
    ) as max_allowed,
    CASE
        WHEN p.policy_status = 'AL'
             AND p.revival_count < (SELECT config_value::int FROM common.system_configuration WHERE config_key = 'max_revivals_allowed')
             AND (SELECT COUNT(*) FROM revival.revival_requests r WHERE r.policy_number = p.policy_number AND r.current_status NOT IN ('COMPLETED', 'WITHDRAWN', 'TERMINATED', 'REJECTED')) = 0
        THEN '✅ ELIGIBLE'
        WHEN p.policy_status != 'AL' THEN '❌ NOT LAPSED'
        WHEN p.revival_count >= (SELECT config_value::int FROM common.system_configuration WHERE config_key = 'max_revivals_allowed')
        THEN '❌ MAX REVIVALS'
        WHEN (SELECT COUNT(*) FROM revival.revival_requests r WHERE r.policy_number = p.policy_number AND r.current_status NOT IN ('COMPLETED', 'WITHDRAWN', 'TERMINATED', 'REJECTED')) > 0
        THEN '❌ HAS ONGOING'
        ELSE '❌ OTHER'
    END as eligibility
FROM common.policies p
WHERE p.policy_number IN ('0000000000001', '0000000000002', '0000000000003', '0000000000004')
ORDER BY p.policy_number;

-- Show revival requests count
SELECT '' as separator;
SELECT
    'Total revival requests created: ' || COUNT(*) as revival_requests_status
FROM revival.revival_requests
WHERE policy_number IN ('0000000000001', '0000000000002', '0000000000003', '0000000000004');

-- Expected result:
-- Policy 1: ✅ ELIGIBLE (AL status, 0 revivals, no ongoing)
-- Policy 2: ❌ NOT LAPSED (IF status)
-- Policy 3: ❌ MAX REVIVALS (2 revivals done)
-- Policy 4: ❌ HAS ONGOING (has ongoing revival request)
