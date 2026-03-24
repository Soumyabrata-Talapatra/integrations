-- +migrate Down
-- Rollback migration for medical examiner columns

ALTER TABLE revival.revival_requests
DROP COLUMN IF EXISTS medical_examiner_code,
DROP COLUMN IF EXISTS medical_examiner_name;
