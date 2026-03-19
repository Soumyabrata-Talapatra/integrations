package domain

import "time"

// ============================================
// Bulk Upload Status Enum (E-006: bulk_upload_batch)
// ============================================

// BulkUploadStatus represents the status of a bulk upload batch
type BulkUploadStatus string

const (
	BulkUploadStatusProcessing BulkUploadStatus = "PROCESSING"
	BulkUploadStatusCompleted  BulkUploadStatus = "COMPLETED"
	BulkUploadStatusFailed     BulkUploadStatus = "FAILED"
)

// ============================================
// E-006: BulkUploadBatch
// ============================================

// BulkUploadBatch represents a bulk proposal upload batch
type BulkUploadBatch struct {
	BatchID          int64            `db:"batch_id" json:"batch_id"`
	FileName         string           `db:"file_name" json:"file_name"`
	TotalRows        int              `db:"total_rows" json:"total_rows"`
	SuccessCount     int              `db:"success_count" json:"success_count"`
	FailureCount     int              `db:"failure_count" json:"failure_count"`
	ErrorReportDocID *string          `db:"error_report_doc_id" json:"error_report_doc_id,omitempty"`
	Status           BulkUploadStatus `db:"status" json:"status"`
	UploadedBy       int64            `db:"uploaded_by" json:"uploaded_by"`
	UploadedAt       time.Time        `db:"uploaded_at" json:"uploaded_at"`
	CompletedAt      *time.Time       `db:"completed_at" json:"completed_at,omitempty"`
	Metadata         *string          `db:"metadata" json:"metadata,omitempty"`
}
