package domain

import "time"

type Document struct {
	EmployeeID             int64     `json:"employee_id" db:"employee_id"`
	FileNameID             uint64    `json:"file_name_id" db:"file_name_id"`
	ServiceTypeID          int       `json:"service_type_id" db:"service_type_id"`
	DocumentName           string    `json:"document_name" db:"document_name"`
	DocumentType           string    `json:"document_type" db:"document_type"`
	DocumentSize           int64     `json:"document_size" db:"document_size"`
	DocumentApproverPostID string    `json:"document_approver_post_id" db:"document_approver_post_id"`
	DocumentUploadStatus   string    `json:"document_upload_status" db:"document_upload_status"`
	DocumentUploadedBy     string    `json:"document_uploaded_by" db:"document_uploaded_by"`
	DocumentUploadedDate   time.Time `json:"document_uploaded_date" db:"document_uploaded_date"`
	DocumentUpdatedBy      string    `json:"document_updated_by,omitempty" db:"document_updated_by,omitempty"`
	DocumentUpdatedDate    time.Time `json:"document_updated_date,omitempty" db:"document_updated_date,omitempty"`
	DocumentApprovedBy     string    `json:"document_approved_by,omitempty" db:"document_approved_by,omitempty"`
	DocumentApprovedDate   time.Time `json:"document_approved_date,omitempty" db:"document_approved_date,omitempty"`
	Remarks                string    `json:"remarks,omitempty" db:"remarks,omitempty"`
	DocumentID             int       `json:"document_id" db:"document_id"`
	DocumentFilePath       string    `json:"document_file_path" db:"document_file_path"`
}
