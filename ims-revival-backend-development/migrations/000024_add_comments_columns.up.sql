-- +migrate Up
-- +migrate Down

-- Add qc_comments column to revival_requests table (idempotent)
ALTER TABLE revival.revival_requests
ADD COLUMN IF NOT EXISTS qc_comments TEXT;

-- Add approval_comments column to revival_requests table (idempotent)
ALTER TABLE revival.revival_requests
ADD COLUMN IF NOT EXISTS approval_comments TEXT;

-- Add comments for the columns
COMMENT ON COLUMN revival.revival_requests.qc_comments IS 'Comments provided during quality check';
COMMENT ON COLUMN revival.revival_requests.approval_comments IS 'Comments provided during approval/rejection';
