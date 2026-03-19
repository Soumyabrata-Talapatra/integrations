-- ============================================
-- Test Data Seed Script
-- Database: surrender_db
-- Purpose: Populate tables with sample data for API testing
-- ============================================

-- ============================================
-- 1. Policy Surrender Requests (VOLUNTARY)
-- ============================================
INSERT INTO policy_surrender_requests (
    id, policy_id, request_number, request_type, previous_policy_status,
    request_date, surrender_value_calculated_date, gross_surrender_value,
    net_surrender_value, paid_up_value, bonus_amount, surrender_factor,
    unpaid_premiums_deduction, loan_deduction, other_deductions,
    disbursement_method, disbursement_amount, reason, status, owner,
    created_at, created_by, metadata
) VALUES
(
    '550e8400-e29b-41d4-a716-446655440001'::uuid,
    '650e8400-e29b-41d4-a716-446655440001'::uuid,
    'VSR-2026-001',
    'VOLUNTARY',
    'AP',
    '2026-01-15',
    '2026-01-15',
    150000.00,
    135000.00,
    140000.00,
    10000.00,
    0.925000,
    5000.00,
    0.00,
    1000.00,
    'CHEQUE',
    135000.00,
    'Financial hardship',
    'PENDING_DOCUMENT_UPLOAD',
    'CUSTOMER',
    NOW(),
    '750e8400-e29b-41d4-a716-446655440001'::uuid,
    '{"policy_number": "POL-2026-001", "customer_name": "John Doe", "phone": "+91-9876543210"}'::jsonb
),
(
    '550e8400-e29b-41d4-a716-446655440002'::uuid,
    '650e8400-e29b-41d4-a716-446655440002'::uuid,
    'VSR-2026-002',
    'VOLUNTARY',
    'AL',
    '2026-01-20',
    '2026-01-20',
    200000.00,
    185000.00,
    190000.00,
    15000.00,
    0.950000,
    8000.00,
    2000.00,
    500.00,
    'CASH',
    185000.00,
    'Policy switch',
    'PENDING_VERIFICATION',
    'CUSTOMER',
    NOW(),
    '750e8400-e29b-41d4-a716-446655440001'::uuid,
    '{"policy_number": "POL-2026-002", "customer_name": "Jane Smith", "phone": "+91-9876543211"}'::jsonb
),
(
    '550e8400-e29b-41d4-a716-446655440003'::uuid,
    '650e8400-e29b-41d4-a716-446655440003'::uuid,
    'VSR-2026-003',
    'VOLUNTARY',
    'AP',
    '2026-01-10',
    '2026-01-10',
    180000.00,
    165000.00,
    170000.00,
    12000.00,
    0.941176,
    6000.00,
    5000.00,
    2000.00,
    'CHEQUE',
    165000.00,
    'Need funds',
    'PENDING_APPROVAL',
    'SYSTEM',
    NOW() - INTERVAL '3 days',
    '750e8400-e29b-41d4-a716-446655440001'::uuid,
    '{"policy_number": "POL-2026-003", "customer_name": "Robert Johnson", "phone": "+91-9876543212"}'::jsonb
);

-- ============================================
-- 2. Policy Surrender Requests (FORCED)
-- ============================================
INSERT INTO policy_surrender_requests (
    id, policy_id, request_number, request_type, previous_policy_status,
    request_date, surrender_value_calculated_date, gross_surrender_value,
    net_surrender_value, paid_up_value, bonus_amount, surrender_factor,
    unpaid_premiums_deduction, loan_deduction, other_deductions,
    disbursement_method, disbursement_amount, reason, status, owner,
    created_at, created_by, metadata
) VALUES
(
    '550e8400-e29b-41d4-a716-446655440004'::uuid,
    '650e8400-e29b-41d4-a716-446655440004'::uuid,
    'FSR-2026-001',
    'FORCED',
    'AP',
    '2026-01-01',
    '2026-01-01',
    120000.00,
    110000.00,
    115000.00,
    8000.00,
    0.956522,
    3000.00,
    4000.00,
    1000.00,
    'CHEQUE',
    110000.00,
    'Premium lapsed - automatic',
    'PENDING_AUTO_COMPLETION',
    'SYSTEM',
    NOW() - INTERVAL '10 days',
    '750e8400-e29b-41d4-a716-446655440001'::uuid,
    '{"policy_number": "POL-2026-004", "customer_name": "Michael Chen", "phone": "+91-9876543213", "loan_amount": 50000}'::jsonb
),
(
    '550e8400-e29b-41d4-a716-446655440005'::uuid,
    '650e8400-e29b-41d4-a716-446655440005'::uuid,
    'FSR-2026-002',
    'FORCED',
    'AL',
    '2026-01-05',
    '2026-01-05',
    95000.00,
    85000.00,
    90000.00,
    6000.00,
    0.944444,
    4000.00,
    3000.00,
    500.00,
    'CHEQUE',
    85000.00,
    'Premium lapsed - automatic',
    'PENDING_AUTO_COMPLETION',
    'SYSTEM',
    NOW() - INTERVAL '5 days',
    '750e8400-e29b-41d4-a716-446655440001'::uuid,
    '{"policy_number": "POL-2026-005", "customer_name": "Sarah Williams", "phone": "+91-9876543214", "loan_amount": 35000}'::jsonb
);

-- ============================================
-- 3. Surrender Bonus Details
-- ============================================
INSERT INTO surrender_bonus_details (
    id, surrender_request_id, financial_year, sum_assured, bonus_rate, bonus_amount, created_at
) VALUES
('550e8400-e29b-41d4-a716-446655550001'::uuid, '550e8400-e29b-41d4-a716-446655440001'::uuid, '2023-2024', 100000.00, 5.00, 5000.00, NOW()),
('550e8400-e29b-41d4-a716-446655550002'::uuid, '550e8400-e29b-41d4-a716-446655440001'::uuid, '2024-2025', 100000.00, 5.00, 5000.00, NOW()),
('550e8400-e29b-41d4-a716-446655550003'::uuid, '550e8400-e29b-41d4-a716-446655440002'::uuid, '2023-2024', 150000.00, 5.50, 8250.00, NOW()),
('550e8400-e29b-41d4-a716-446655550004'::uuid, '550e8400-e29b-41d4-a716-446655440002'::uuid, '2024-2025', 150000.00, 5.50, 8250.00, NOW()),
('550e8400-e29b-41d4-a716-446655550005'::uuid, '550e8400-e29b-41d4-a716-446655440004'::uuid, '2023-2024', 100000.00, 4.50, 4500.00, NOW()),
('550e8400-e29b-41d4-a716-446655550006'::uuid, '550e8400-e29b-41d4-a716-446655440005'::uuid, '2023-2024', 80000.00, 4.50, 3600.00, NOW());

-- ============================================
-- 4. Forced Surrender Reminders
-- ============================================
INSERT INTO forced_surrender_reminders (
    id, policy_id, reminder_number, reminder_date, loan_capitalization_ratio,
    loan_principal, loan_interest, gross_surrender_value,
    letter_sent, sms_sent, letter_reference, sms_reference, created_at, metadata
) VALUES
('550e8400-e29b-41d4-a716-446655660001'::uuid, '650e8400-e29b-41d4-a716-446655440004'::uuid, 'FIRST', '2025-11-01', 0.4500, 45000.00, 5000.00, 120000.00, TRUE, TRUE, 'LTR-2025-11-001', 'SMS-2025-11-001', NOW() - INTERVAL '60 days', '{}'),
('550e8400-e29b-41d4-a716-446655660002'::uuid, '650e8400-e29b-41d4-a716-446655440004'::uuid, 'SECOND', '2025-12-01', 0.4650, 46500.00, 5500.00, 120000.00, TRUE, TRUE, 'LTR-2025-12-001', 'SMS-2025-12-001', NOW() - INTERVAL '30 days', '{}'),
('550e8400-e29b-41d4-a716-446655660003'::uuid, '650e8400-e29b-41d4-a716-446655440005'::uuid, 'FIRST', '2025-12-15', 0.3800, 32000.00, 3500.00, 95000.00, TRUE, FALSE, 'LTR-2025-12-002', NULL, NOW() - INTERVAL '45 days', '{}');

-- ============================================
-- 5. Forced Surrender Payment Windows
-- ============================================
INSERT INTO forced_surrender_payment_windows (
    id, surrender_request_id, policy_id, window_start_date, window_end_date,
    payment_received, payment_received_at, payment_amount, payment_reference,
    workflow_forwarded, workflow_forwarded_at, auto_completed, auto_completed_at, created_at
) VALUES
('550e8400-e29b-41d4-a716-446655770001'::uuid, '550e8400-e29b-41d4-a716-446655440004'::uuid, '650e8400-e29b-41d4-a716-446655440004'::uuid, '2026-01-01', '2026-01-31', FALSE, NULL, NULL, NULL, FALSE, NULL, FALSE, NULL, NOW()),
('550e8400-e29b-41d4-a716-446655770002'::uuid, '550e8400-e29b-41d4-a716-446655440005'::uuid, '650e8400-e29b-41d4-a716-446655440005'::uuid, '2026-01-05', '2026-02-05', FALSE, NULL, NULL, NULL, FALSE, NULL, FALSE, NULL, NOW());

-- ============================================
-- 6. Surrender Documents
-- ============================================
INSERT INTO surrender_documents (
    id, surrender_request_id, document_type, document_name, document_path,
    uploaded_date, file_size_bytes, mime_type, verified, verified_by, verified_at,
    rejection_reason, created_at, deleted_at, metadata
) VALUES
('550e8400-e29b-41d4-a716-446655880001'::uuid, '550e8400-e29b-41d4-a716-446655440001'::uuid, 'WRITTEN_CONSENT', 'Consent_Letter_Doe.pdf', '/documents/VSR-2026-001/Consent_Letter_Doe.pdf', '2026-01-16', 125000, 'application/pdf', FALSE, NULL, NULL, NULL, NOW(), NULL, '{"page_count": 2, "size_mb": 0.12}'),
('550e8400-e29b-41d4-a716-446655880002'::uuid, '550e8400-e29b-41d4-a716-446655440001'::uuid, 'POLICY_BOND', 'Policy_Bond_Doe.pdf', '/documents/VSR-2026-001/Policy_Bond_Doe.pdf', '2026-01-16', 250000, 'application/pdf', FALSE, NULL, NULL, NULL, NOW(), NULL, '{"page_count": 5, "size_mb": 0.24}'),
('550e8400-e29b-41d4-a716-446655880003'::uuid, '550e8400-e29b-41d4-a716-446655440002'::uuid, 'WRITTEN_CONSENT', 'Consent_Smith.pdf', '/documents/VSR-2026-002/Consent_Smith.pdf', '2026-01-21', 128000, 'application/pdf', TRUE, '850e8400-e29b-41d4-a716-446655440001'::uuid, NOW() - INTERVAL '2 days', NULL, NOW() - INTERVAL '2 days', NULL, '{"page_count": 2, "size_mb": 0.13, "verified_by": "Verification Officer"}'),
('550e8400-e29b-41d4-a716-446655880004'::uuid, '550e8400-e29b-41d4-a716-446655440002'::uuid, 'PREMIUM_RECEIPT_BOOK', 'Premium_Receipt_Smith.pdf', '/documents/VSR-2026-002/Premium_Receipt_Smith.pdf', '2026-01-21', 300000, 'application/pdf', TRUE, '850e8400-e29b-41d4-a716-446655440001'::uuid, NOW() - INTERVAL '2 days', NULL, NOW() - INTERVAL '2 days', NULL, '{"page_count": 6, "size_mb": 0.29}'),
('550e8400-e29b-41d4-a716-446655880005'::uuid, '550e8400-e29b-41d4-a716-446655440003'::uuid, 'WRITTEN_CONSENT', 'Consent_Johnson.pdf', '/documents/VSR-2026-003/Consent_Johnson.pdf', '2026-01-13', 130000, 'application/pdf', TRUE, '850e8400-e29b-41d4-a716-446655440001'::uuid, NOW() - INTERVAL '4 days', NULL, NOW() - INTERVAL '4 days', NULL, '{"page_count": 2, "size_mb": 0.13}');

-- ============================================
-- 7. Approval Workflow Tasks
-- ============================================
INSERT INTO approval_workflow_tasks (
    id, surrender_request_id, task_number, office_code, assigned_to, status,
    reserved, reserved_at, reserved_by, reservation_expires_at, priority,
    created_at, completed_at, completed_by, escalated, escalated_to, escalated_at,
    escalation_reason, metadata
) VALUES
('550e8400-e29b-41d4-a716-446655990001'::uuid, '550e8400-e29b-41d4-a716-446655440003'::uuid, 'APR-2026-001', 'PUNE', '950e8400-e29b-41d4-a716-446655440001'::uuid, 'PENDING', FALSE, NULL, NULL, NULL, 'MEDIUM', NOW() - INTERVAL '3 days', NULL, NULL, FALSE, NULL, NULL, NULL, '{"office_address": "PUNE-001", "reviewed_by": "Approval Team"}'),
('550e8400-e29b-41d4-a716-446655990002'::uuid, '550e8400-e29b-41d4-a716-446655440002'::uuid, 'APR-2026-002', 'MUMBAI', '950e8400-e29b-41d4-a716-446655440002'::uuid, 'IN_PROGRESS', TRUE, NOW() - INTERVAL '1 day', '950e8400-e29b-41d4-a716-446655440002'::uuid, NOW() + INTERVAL '6 days', 'HIGH', NOW() - INTERVAL '2 days', NULL, NULL, FALSE, NULL, NULL, NULL, '{"office_address": "MUMBAI-001", "notes": "Pending final verification"}'),
('550e8400-e29b-41d4-a716-446655990003'::uuid, '550e8400-e29b-41d4-a716-446655440001'::uuid, 'APR-2026-003', 'DELHI', NULL, 'PENDING', FALSE, NULL, NULL, NULL, 'LOW', NOW() - INTERVAL '1 day', NULL, NULL, FALSE, NULL, NULL, NULL, '{"office_address": "DELHI-001", "priority_reason": "Standard processing"}');

-- ============================================
-- 8. Surrender Payments
-- ============================================
INSERT INTO surrender_payments (
    id, surrender_request_id, payment_number, payment_date, amount,
    disbursement_method, cheque_number, cheque_date, bank_name, branch_name,
    payee_name, payee_address, transaction_reference, status,
    processed_at, processed_by, created_at, metadata
) VALUES
('550e8400-e29b-41d4-a716-446655110001'::uuid, '550e8400-e29b-41d4-a716-446655440001'::uuid, 'PAY-2026-001', '2026-01-18', 135000.00, 'CHEQUE', 'CHQ-000123', '2026-01-18', 'HDFC Bank', 'Pune Branch', 'John Doe', '123 Main St, Pune', 'REF-2026-001', 'PENDING', NULL, NULL, NOW(), '{"cheque_status": "issued", "bank_remarks": ""}'),
('550e8400-e29b-41d4-a716-446655110002'::uuid, '550e8400-e29b-41d4-a716-446655440002'::uuid, 'PAY-2026-002', '2026-01-22', 185000.00, 'CASH', NULL, NULL, NULL, NULL, 'Jane Smith', '456 Oak Ave, Mumbai', 'REF-2026-002', 'PROCESSED', NOW() - INTERVAL '2 days', '850e8400-e29b-41d4-a716-446655440001'::uuid, NOW() - INTERVAL '2 days', '{"disbursement_method": "cash_handover", "received_by": "Customer_Self"}');

-- ============================================
-- 9. Surrender Value Calculations (Audit Trail)
-- ============================================
INSERT INTO surrender_value_calculations (
    id, surrender_request_id, calculation_date, paid_up_value, bonus_amount,
    surrender_factor, gross_surrender_value, unpaid_premiums_deduction,
    loan_principal_deduction, loan_interest_deduction, net_surrender_value,
    calculation_breakdown, calculated_by, created_at
) VALUES
('550e8400-e29b-41d4-a716-446655220001'::uuid, '550e8400-e29b-41d4-a716-446655440001'::uuid, '2026-01-15', 140000.00, 10000.00, 0.925000, 150000.00, 5000.00, 0.00, 0.00, 135000.00, '{"method": "standard_calculation", "factors": ["paid_up_value", "bonus", "surrender_factor"]}', '750e8400-e29b-41d4-a716-446655440001'::uuid, NOW()),
('550e8400-e29b-41d4-a716-446655220002'::uuid, '550e8400-e29b-41d4-a716-446655440002'::uuid, '2026-01-20', 190000.00, 15000.00, 0.950000, 200000.00, 8000.00, 2000.00, 0.00, 185000.00, '{"method": "standard_calculation", "factors": ["paid_up_value", "bonus", "surrender_factor", "loan_deduction"]}', '750e8400-e29b-41d4-a716-446655440001'::uuid, NOW());

-- ============================================
-- 10. Policy Surrender Dispositions
-- ============================================
INSERT INTO policy_surrender_dispositions (
    id, surrender_request_id, disposition_type, new_policy_status, new_sum_assured,
    prescribed_limit, net_surrender_value, reduced_paid_up_created,
    reduced_paid_up_policy_number, terminated, termination_reason, created_at
) VALUES
('550e8400-e29b-41d4-a716-446655330001'::uuid, '550e8400-e29b-41d4-a716-446655440003'::uuid, 'TERMINATED_SURRENDER', 'TS', NULL, NULL, 165000.00, FALSE, NULL, TRUE, 'Voluntary surrender request approved', NOW());

-- ============================================
-- Summary of Test Data Created
-- ============================================
-- Total Surrender Requests: 5 (3 Voluntary, 2 Forced)
-- Approval Tasks: 3 (1 Pending, 1 In Progress, 1 Assigned)
-- Documents: 5 (3 Verified, 2 Pending verification)
-- Payments: 2 (1 Pending, 1 Processed)
-- Bonus Details: 6 records
-- Forced Reminders: 3 records
-- Payment Windows: 2 records
-- Value Calculations: 2 records
-- Dispositions: 1 record
