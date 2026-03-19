-- ============================================================
-- Migration 003: PM Integration — store Temporal workflow ID
-- and PM cross-reference fields on surrender_requests
-- ============================================================
--
-- When Policy Management dispatches SurrenderProcessingWorkflow
-- as a child workflow, the workflow ID is assigned by PM (e.g.
-- "sur-{idempotency_key}") and is not predictable from the
-- surrender_request_id alone. We store it here so DE/QC/Approval
-- handlers can look it up and signal the correct workflow.
--
-- pm_service_request_id : PM's service_request.request_id (BIGINT)
-- pm_policy_db_id       : PM's policy.policy_id (BIGINT)
-- temporal_workflow_id  : running Temporal workflow ID for this request
-- ============================================================

ALTER TABLE finservicemgmt.surrender_requests
    ADD COLUMN IF NOT EXISTS temporal_workflow_id  TEXT,
    ADD COLUMN IF NOT EXISTS pm_service_request_id BIGINT,
    ADD COLUMN IF NOT EXISTS pm_policy_db_id       BIGINT;

-- Index so handlers can resolve workflow ID in O(1)
CREATE INDEX IF NOT EXISTS idx_surrender_requests_temporal_workflow_id
    ON finservicemgmt.surrender_requests (temporal_workflow_id)
    WHERE temporal_workflow_id IS NOT NULL;

COMMENT ON COLUMN finservicemgmt.surrender_requests.temporal_workflow_id
    IS 'Temporal workflow ID for this surrender request. Used by DE/QC/Approval handlers to signal the correct running workflow.';

COMMENT ON COLUMN finservicemgmt.surrender_requests.pm_service_request_id
    IS 'Cross-reference to Policy Management service_request.request_id (BIGINT).';

COMMENT ON COLUMN finservicemgmt.surrender_requests.pm_policy_db_id
    IS 'Cross-reference to Policy Management policy.policy_id (BIGINT).';
