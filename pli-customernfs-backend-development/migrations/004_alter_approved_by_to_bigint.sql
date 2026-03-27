-- ============================================================================
-- Migration: 004_alter_approved_by_to_bigint.sql
-- Description: Change approved_by columns from UUID to BIGINT to align with other user ID columns
-- Module: customer-nfs-service
-- Version: 1.0.0
-- ============================================================================

-- ===========================================================================
-- SECTION 1: ALTER approved_by COLUMNS
-- ===========================================================================

-- Helper function to check if column needs alteration (similar to 003 migration)
CREATE OR REPLACE FUNCTION nfs.alter_column_if_needed_004(
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
-- SECTION 2: ALTER ALL approved_by COLUMNS
-- ===========================================================================

-- 1. nfs.service_request - initiated_by, assigned_to, approved_by columns
SELECT nfs.alter_column_if_needed_004('service_request', 'initiated_by', 'bigint');
SELECT nfs.alter_column_if_needed_004('service_request', 'assigned_to', 'bigint');
SELECT nfs.alter_column_if_needed_004('service_request', 'approved_by', 'bigint');

-- 2. nfs.withdrawal_request - approved_by column  
SELECT nfs.alter_column_if_needed_004('withdrawal_request', 'approved_by', 'bigint');

-- 3. nfs.name_version_history - created_by column
SELECT nfs.alter_column_if_needed_004('name_version_history', 'created_by', 'bigint');

-- 4. nfs.address_version_history - created_by column
SELECT nfs.alter_column_if_needed_004('address_version_history', 'created_by', 'bigint');

-- ===========================================================================
-- SECTION 3: CLEANUP
-- ===========================================================================

-- Drop the helper function
DROP FUNCTION nfs.alter_column_if_needed_004(TEXT, TEXT, TEXT);

-- ===========================================================================
-- SECTION 4: MIGRATION NOTES
-- ===========================================================================

-- IMPORTANT NOTES:
-- 1. This migration changes approved_by columns from UUID to BIGINT to align with
--    other user ID columns that were changed in migration 003.
-- 2. For UUID columns, the conversion uses a hash function to convert UUID to BIGINT.
-- 3. If the column contains actual UUID values, they will be converted to numeric
--    values via MD5 hash.
-- 4. If the column contains numeric values as strings, they will be cast directly.
-- 5. This migration should be run after migration 003.