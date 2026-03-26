-- ============================================================================
-- Migration: 003_alter_columns_to_bigint.sql
-- Description: Change UUID/VARCHAR columns to BIGINT for user ID fields
-- Module: customer-nfs-service
-- Version: 1.0.0
-- ============================================================================

-- ===========================================================================
-- SECTION 1: VIEW DEFINITIONS (to be recreated after column changes)
-- ===========================================================================

-- Store the view definitions that depend on columns being altered
CREATE OR REPLACE FUNCTION nfs.recreate_views_after_alter() RETURNS VOID AS $$
BEGIN
    -- Drop views that depend on columns being altered
    DROP VIEW IF EXISTS nfs.v_document_status CASCADE;
    DROP VIEW IF EXISTS nfs.v_audit_trail CASCADE;
    
    -- Recreate v_document_status with BIGINT columns
    CREATE OR REPLACE VIEW nfs.v_document_status AS
    SELECT sr.request_id, sr.ticket_number, sr.status AS request_status,
           du.document_id, du.document_type, du.file_name, du.verification_status,
           du.uploaded_at, du.verified_at, du.verified_by, du.rejection_reason AS document_rejection_reason
    FROM nfs.service_request sr
    JOIN nfs.document_upload du ON sr.request_id = du.request_id
    WHERE sr.deleted_at IS NULL AND du.deleted_at IS NULL;
    
    -- Recreate v_audit_trail with BIGINT columns
    CREATE OR REPLACE VIEW nfs.v_audit_trail AS
    SELECT al.audit_id, al.request_id, sr.ticket_number, al.action_type,
           al.old_value, al.new_value, al.performed_by, al.performed_at,
           al.ip_address, al.channel, al.office_code, al.remarks, al.error_code, al.error_message
    FROM nfs.audit_log al JOIN nfs.service_request sr ON al.request_id = sr.request_id;
    
    RAISE NOTICE 'Recreated dependent views with BIGINT columns';
END;
$$ LANGUAGE plpgsql;

-- ===========================================================================
-- SECTION 2: CONDITIONAL ALTER COMMANDS
-- ===========================================================================

-- Helper function to check if column needs alteration
CREATE OR REPLACE FUNCTION nfs.alter_column_if_needed(
    p_table_name TEXT,
    p_column_name TEXT,
    p_target_type TEXT
) RETURNS VOID AS $$
DECLARE
    v_current_type TEXT;
    v_alter_sql TEXT;
    v_using_clause TEXT;
BEGIN
    -- Get current column type
    SELECT data_type 
    INTO v_current_type
    FROM information_schema.columns 
    WHERE table_schema = 'nfs' 
      AND table_name = p_table_name 
      AND column_name = p_column_name;
    
    -- Only alter if current type is not already target type
    IF v_current_type IS NOT NULL AND v_current_type != p_target_type THEN
        -- Determine the USING clause based on current type
        IF v_current_type = 'uuid' THEN
            -- For UUID columns, use hash-based conversion
            v_using_clause := format('USING ((''x'' || substr(md5(%I::text), 1, 16))::bit(64)::bigint)', p_column_name);
        ELSIF v_current_type = 'character varying' THEN
            -- For VARCHAR columns, try direct cast
            v_using_clause := format('USING %I::bigint', p_column_name);
        ELSIF v_current_type = 'bigint' THEN
            -- Already BIGINT, no change needed
            RAISE NOTICE 'Column nfs.%.% already has type % (no change needed)', p_table_name, p_column_name, p_target_type;
            RETURN;
        ELSE
            -- For other types, use default cast
            v_using_clause := format('USING %I::%s', p_column_name, p_target_type);
        END IF;
        
        v_alter_sql := format(
            'ALTER TABLE nfs.%I ALTER COLUMN %I TYPE %s %s',
            p_table_name, p_column_name, p_target_type, v_using_clause
        );
        EXECUTE v_alter_sql;
        RAISE NOTICE 'Altered nfs.%.% from % to %', p_table_name, p_column_name, v_current_type, p_target_type;
    ELSE
        RAISE NOTICE 'Column nfs.%.% already has type % (no change needed)', p_table_name, p_column_name, p_target_type;
    END IF;
END;
$$ LANGUAGE plpgsql;

-- ===========================================================================
-- SECTION 2: DROP DEPENDENT VIEWS
-- ===========================================================================

DO $$ 
BEGIN
    -- Drop views that depend on columns being altered
    DROP VIEW IF EXISTS nfs.v_document_status CASCADE;
    DROP VIEW IF EXISTS nfs.v_audit_trail CASCADE;
    RAISE NOTICE 'Dropped dependent views';
END $$;

-- ===========================================================================
-- SECTION 3: ALTER ALL COLUMNS
-- ===========================================================================

-- 1. nfs.service_request
SELECT nfs.alter_column_if_needed('service_request', 'created_by', 'bigint');
SELECT nfs.alter_column_if_needed('service_request', 'updated_by', 'bigint');

-- 2. nfs.address_change_detail
SELECT nfs.alter_column_if_needed('address_change_detail', 'created_by', 'bigint');
SELECT nfs.alter_column_if_needed('address_change_detail', 'updated_by', 'bigint');

-- 3. nfs.name_change_detail
SELECT nfs.alter_column_if_needed('name_change_detail', 'created_by', 'bigint');
SELECT nfs.alter_column_if_needed('name_change_detail', 'updated_by', 'bigint');

-- 4. nfs.document_upload
SELECT nfs.alter_column_if_needed('document_upload', 'uploaded_by', 'bigint');
SELECT nfs.alter_column_if_needed('document_upload', 'verified_by', 'bigint');
SELECT nfs.alter_column_if_needed('document_upload', 'created_by', 'bigint');
SELECT nfs.alter_column_if_needed('document_upload', 'updated_by', 'bigint');

-- 5. nfs.missing_document_request
SELECT nfs.alter_column_if_needed('missing_document_request', 'requested_by', 'bigint');
SELECT nfs.alter_column_if_needed('missing_document_request', 'created_by', 'bigint');
SELECT nfs.alter_column_if_needed('missing_document_request', 'updated_by', 'bigint');

-- 6. nfs.audit_log
SELECT nfs.alter_column_if_needed('audit_log', 'performed_by', 'bigint');

-- 7. nfs.status_transition_history
SELECT nfs.alter_column_if_needed('status_transition_history', 'transitioned_by', 'bigint');

-- 8. nfs.cpc_work_queue
SELECT nfs.alter_column_if_needed('cpc_work_queue', 'picked_up_by', 'bigint');

-- 9. nfs.address_version_history
SELECT nfs.alter_column_if_needed('address_version_history', 'created_by', 'bigint');

-- 10. nfs.name_version_history
SELECT nfs.alter_column_if_needed('name_version_history', 'created_by', 'bigint');

-- 11. nfs.withdrawal_request
SELECT nfs.alter_column_if_needed('withdrawal_request', 'requested_by', 'bigint');
SELECT nfs.alter_column_if_needed('withdrawal_request', 'created_by', 'bigint');
SELECT nfs.alter_column_if_needed('withdrawal_request', 'updated_by', 'bigint');

-- 12. nfs.mobile_change_detail
SELECT nfs.alter_column_if_needed('mobile_change_detail', 'created_by', 'bigint');
SELECT nfs.alter_column_if_needed('mobile_change_detail', 'updated_by', 'bigint');

-- 13. nfs.mobile_version_history
SELECT nfs.alter_column_if_needed('mobile_version_history', 'created_by', 'bigint');

-- 14. nfs.email_change_detail
SELECT nfs.alter_column_if_needed('email_change_detail', 'created_by', 'bigint');
SELECT nfs.alter_column_if_needed('email_change_detail', 'updated_by', 'bigint');

-- 15. nfs.email_version_history
SELECT nfs.alter_column_if_needed('email_version_history', 'created_by', 'bigint');

-- ===========================================================================
-- SECTION 4: RECREATE VIEWS
-- ===========================================================================

-- Recreate v_document_status with BIGINT columns
CREATE OR REPLACE VIEW nfs.v_document_status AS
SELECT sr.request_id, sr.ticket_number, sr.status AS request_status,
       du.document_id, du.document_type, du.file_name, du.verification_status,
       du.uploaded_at, du.verified_at, du.verified_by, du.rejection_reason AS document_rejection_reason
FROM nfs.service_request sr
JOIN nfs.document_upload du ON sr.request_id = du.request_id
WHERE sr.deleted_at IS NULL AND du.deleted_at IS NULL;

-- Recreate v_audit_trail with BIGINT columns
CREATE OR REPLACE VIEW nfs.v_audit_trail AS
SELECT al.audit_id, al.request_id, sr.ticket_number, al.action_type,
       al.old_value, al.new_value, al.performed_by, al.performed_at,
       al.ip_address, al.channel, al.office_code, al.remarks, al.error_code, al.error_message
FROM nfs.audit_log al JOIN nfs.service_request sr ON al.request_id = sr.request_id;

DO $$ 
BEGIN
    RAISE NOTICE 'Recreated dependent views with BIGINT columns';
END $$;

-- ===========================================================================
-- SECTION 5: CLEANUP
-- ===========================================================================

-- Drop the helper functions
DROP FUNCTION nfs.recreate_views_after_alter();
DROP FUNCTION nfs.alter_column_if_needed(TEXT, TEXT, TEXT);

-- ===========================================================================
-- SECTION 4: MIGRATION NOTES
-- ===========================================================================

-- IMPORTANT NOTES:
-- 1. This migration uses conditional logic to only alter columns that are not already BIGINT
-- 2. For UUID columns, the conversion will fail if data exists. You may need to:
--    - Clear existing data first, OR
--    - Use a custom conversion function for UUID to BIGINT
-- 3. For VARCHAR columns containing numeric values, conversion will work
-- 4. If you have UUID data that needs preservation, create a custom conversion:
--    Example: USING (('x' || substr(md5(uuid_column::text), 1, 16))::bit(64)::bigint)