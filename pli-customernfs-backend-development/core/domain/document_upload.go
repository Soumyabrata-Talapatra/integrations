package domain

import "time"

// DocumentUpload represents documents uploaded as part of NFS request processing.
// Entity ID: E-4 (per Customer_NFS_Service_Analysis.md Section 7.4)
//
// Business Rules applied:
//   - FR-NFS-002: Address change manual path — Address Proof is mandatory
//   - FR-NFS-005: Name change manual path — at least 1 of Gazette/Newspaper required
//   - FR-NFS-009: Missing document sub-workflow — secure upload link documents
//   - BR-NFS-008: At least one legal document required for manual name change
//
// Validation Rules:
//   - VR-NFS-009: max 5 MB per file
//   - VR-NFS-010: MIME type must be application/pdf, image/jpeg, or image/png
//
// BATCH NOTE: Multiple document uploads in one request use batch insert:
//   batch := pgx.Batch{}
//   FOR each doc:
//     → INSERT nfs.document_upload
//   END
//   → INSERT nfs.audit_log (DOCUMENT_UPLOAD action per doc)
// For list queries (count + data), uses batch SELECT:
//   batch → SELECT COUNT(*) + SELECT rows (single round-trip)
// See: repo/postgres/document_upload.go CreateBatch() and ListByRequestID()
type DocumentUpload struct {
	DocumentID         string     `json:"document_id" db:"document_id"`
	RequestID          string     `json:"request_id" db:"request_id"`
	DocumentType       string     `json:"document_type" db:"document_type"`
	FileName           string     `json:"file_name" db:"file_name"`
	FileURL            string     `json:"file_url" db:"file_url"`
	FileSizeBytes      int64      `json:"file_size_bytes" db:"file_size_bytes"`
	MimeType           string     `json:"mime_type" db:"mime_type"`
	UploadedBy         string     `json:"uploaded_by" db:"uploaded_by"`
	UploadChannel      string     `json:"upload_channel" db:"upload_channel"`
	UploadedAt         time.Time  `json:"uploaded_at" db:"uploaded_at"`
	VerificationStatus string     `json:"verification_status" db:"verification_status"`
	VerifiedBy         *string    `json:"verified_by,omitempty" db:"verified_by"`
	VerifiedAt         *time.Time `json:"verified_at,omitempty" db:"verified_at"`
	RejectionReason    *string    `json:"rejection_reason,omitempty" db:"rejection_reason"`
	DMSDocumentID      *string    `json:"dms_document_id,omitempty" db:"dms_document_id"`
	DMSStoragePath     *string    `json:"dms_storage_path,omitempty" db:"dms_storage_path"`
	VirusScanStatus    string     `json:"virus_scan_status" db:"virus_scan_status"`
	VirusScannedAt     *time.Time `json:"virus_scanned_at,omitempty" db:"virus_scanned_at"`
	CreatedAt          time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at" db:"updated_at"`
	CreatedBy          string     `json:"created_by" db:"created_by"`
	UpdatedBy          *string    `json:"updated_by,omitempty" db:"updated_by"`
	Version            int        `json:"version" db:"version"`
}

// MissingDocumentRequest tracks CPC-initiated requests for missing documents.
// Entity ID: E-5 (per Customer_NFS_Service_Analysis.md Section 7.5)
//
// Business Rules applied:
//   - BR-NFS-014: Max 3 link generation attempts, 7-day expiry each
//   - FR-NFS-009: CPC sends secure upload link, customer uploads within expiry
//   - After 3 expired links: status → DOCUMENTS_EXPIRED (terminal state)
//
// BATCH NOTE: Missing doc request creation batches:
//   batch → INSERT nfs.missing_document_request
//   batch → UPDATE nfs.service_request SET status='PENDING_DOCUMENTS'
//   batch → INSERT nfs.status_transition_history
//   batch → INSERT nfs.audit_log (MISSING_DOC_REQUESTED action)
// All in one TX to ensure consistency per BR-NFS-012 state machine.
// See: repo/postgres/missing_document.go Create()
type MissingDocumentRequest struct {
	MissingDocID    string     `json:"missing_doc_id" db:"missing_doc_id"`
	RequestID       string     `json:"request_id" db:"request_id"`
	DocumentType    string     `json:"document_type" db:"document_type"`
	SecureLinkURL   *string    `json:"secure_link_url,omitempty" db:"secure_link_url"`
	SecureLinkToken *string    `json:"secure_link_token,omitempty" db:"secure_link_token"`
	LinkExpiry      *time.Time `json:"link_expiry,omitempty" db:"link_expiry"`
	// LinkGenerationCount: max 3 (BR-NFS-014)
	LinkGenerationCount int        `json:"link_generation_count" db:"link_generation_count"`
	Status              string     `json:"status" db:"status"`
	ReceivedAt          *time.Time `json:"received_at,omitempty" db:"received_at"`
	ReceivedVia         *string    `json:"received_via,omitempty" db:"received_via"`
	ReceivedDocumentID  *string    `json:"received_document_id,omitempty" db:"received_document_id"`
	ReminderCount       int        `json:"reminder_count" db:"reminder_count"`
	LastReminderAt      *time.Time `json:"last_reminder_at,omitempty" db:"last_reminder_at"`
	RequestedBy         string     `json:"requested_by" db:"requested_by"`
	MessageToCustomer   *string    `json:"message_to_customer,omitempty" db:"message_to_customer"`
	CreatedAt           time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at" db:"updated_at"`
	CreatedBy           string     `json:"created_by" db:"created_by"`
	UpdatedBy           *string    `json:"updated_by,omitempty" db:"updated_by"`
	Version             int        `json:"version" db:"version"`
}
