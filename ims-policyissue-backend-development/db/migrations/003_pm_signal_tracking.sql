-- ============================================
-- Migration 003: PM Signal Tracking
-- Purpose: Track Policy Manager (PM) SignalWithStart status per issued policy.
--          Enables the reconciliation worker to detect and retry orphaned signals.
-- ============================================

ALTER TABLE policy_issue.proposal_issuance
    ADD COLUMN IF NOT EXISTS pm_signal_status   VARCHAR(20)  NOT NULL DEFAULT 'PENDING'
        CONSTRAINT chk_pm_signal_status CHECK (pm_signal_status IN ('PENDING', 'SENT', 'FAILED')),
    ADD COLUMN IF NOT EXISTS pm_signal_sent_at  TIMESTAMP WITH TIME ZONE,
    ADD COLUMN IF NOT EXISTS pm_signal_attempts INT          NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS pm_signal_last_error TEXT,
    ADD COLUMN IF NOT EXISTS pm_plw_workflow_id  VARCHAR(100);

COMMENT ON COLUMN policy_issue.proposal_issuance.pm_signal_status IS
    'Tracks whether the PM SignalWithStart for this policy has been delivered. '
    'PENDING = not yet confirmed; SENT = SignalWithStart succeeded; FAILED = retries exhausted.';

COMMENT ON COLUMN policy_issue.proposal_issuance.pm_signal_sent_at IS
    'Timestamp when SignalWithStart returned successfully.';

COMMENT ON COLUMN policy_issue.proposal_issuance.pm_signal_attempts IS
    'Total number of SignalWithStart attempts made (incremented each time the activity runs).';

COMMENT ON COLUMN policy_issue.proposal_issuance.pm_signal_last_error IS
    'Last error string from a failed SignalWithStart attempt. Cleared on success.';

COMMENT ON COLUMN policy_issue.proposal_issuance.pm_plw_workflow_id IS
    'Temporal workflow ID of the PM lifecycle workflow (e.g. plw-PLI/2026/000001). Set on success.';

-- Partial index: only rows that need reconciliation attention
CREATE INDEX IF NOT EXISTS idx_issuance_pm_signal_pending
    ON policy_issue.proposal_issuance (policy_issue_date)
    WHERE pm_signal_status IN ('PENDING', 'FAILED');
