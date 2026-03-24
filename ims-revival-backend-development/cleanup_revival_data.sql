-- Cleanup script to reset revival workflow data
-- This allows restarting tests from the beginning

-- Delete installment schedules
DELETE FROM revival.installment_schedules 
WHERE request_id IN (
  SELECT request_id FROM revival.revival_requests 
  WHERE policy_number = '0000000000001'
);

-- Delete payment transactions
DELETE FROM collection.payment_transactions 
WHERE request_id IN (
  SELECT request_id FROM revival.revival_requests 
  WHERE policy_number = '0000000000001'
);

-- Delete workflow state
DELETE FROM revival.revival_request_workflow_state 
WHERE request_id IN (
  SELECT request_id FROM revival.revival_requests 
  WHERE policy_number = '0000000000001'
);

-- Delete status change history
DELETE FROM revival.status_change_history 
WHERE request_id IN (
  SELECT request_id FROM revival.revival_requests 
  WHERE policy_number = '0000000000001'
);

-- Delete revival requests
DELETE FROM revival.revival_requests 
WHERE policy_number = '0000000000001';

-- Update policy status back to AL (LAPSED) for testing
UPDATE common.policies 
SET policy_status = 'AL' 
WHERE policy_number = '0000000000001';

SELECT 'Cleanup complete! Policy 0000000000001 is ready for revival testing.' as status;
