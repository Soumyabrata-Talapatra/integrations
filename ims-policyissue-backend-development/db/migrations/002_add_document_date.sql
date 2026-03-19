-- Migration 002: Add document_date column to proposal_document_ref
-- Reason: VR-PI-023 requires document date validation (not in future).
-- The date must be persisted for audit trail and compliance.

ALTER TABLE proposal_document_ref
    ADD COLUMN document_date DATE;

COMMENT ON COLUMN proposal_document_ref.document_date IS 'Date of the document (VR-PI-023: must not be in the future)';
