-- +migrate Up
-- Add medical examiner columns to revival_requests table

ALTER TABLE revival.revival_requests
ADD COLUMN IF NOT EXISTS medical_examiner_code VARCHAR DEFAULT NULL,
ADD COLUMN IF NOT EXISTS medical_examiner_name VARCHAR DEFAULT NULL;

-- Add comments for the new columns
COMMENT ON COLUMN revival.revival_requests.medical_examiner_code IS 'Code of the medical examiner conducting the examination';
COMMENT ON COLUMN revival.revival_requests.medical_examiner_name IS 'Name of the medical examiner conducting the examination';

-- +migrate Down
-- Drop medical examiner columns from revival_requests table

ALTER TABLE revival.revival_requests
DROP COLUMN IF EXISTS medical_examiner_code,
DROP COLUMN IF EXISTS medical_examiner_name;
