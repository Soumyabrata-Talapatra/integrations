-- +migrate Up
-- +migrate Down

-- Add revival_type column to revival_requests table
ALTER TABLE revival.revival_requests
ADD COLUMN revival_type VARCHAR;

-- Add comment for the column
COMMENT ON COLUMN revival.revival_requests.revival_type IS 'Type of revival: installment or lumpsum';