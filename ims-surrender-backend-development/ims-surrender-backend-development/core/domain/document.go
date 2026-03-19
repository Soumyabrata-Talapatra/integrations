package domain

import (
	"time"

	"github.com/google/uuid"
)

// DocumentType represents the type of document required for surrender
// Validation Rule: VR-SUR-003
type DocumentType string

const (
	DocumentTypeWrittenConsent     DocumentType = "WRITTEN_CONSENT"
	DocumentTypePolicyBond         DocumentType = "POLICY_BOND"
	DocumentTypePremiumReceiptBook DocumentType = "PREMIUM_RECEIPT_BOOK"
	DocumentTypePayRecoveryCert    DocumentType = "PAY_RECOVERY_CERTIFICATE"
	DocumentTypeLoanReceiptBook    DocumentType = "LOAN_RECEIPT_BOOK"
	DocumentTypeLoanBond           DocumentType = "LOAN_BOND"
	DocumentTypeIndemnityBond      DocumentType = "INDEMNITY_BOND"
	DocumentTypeAssignmentDeed     DocumentType = "ASSIGNMENT_DEED"
	DocumentTypeDischargeReceipt   DocumentType = "DISCHARGE_RECEIPT"
)

// SurrenderDocument represents an uploaded document for surrender request
// Table: surrender_documents
// Functional Requirement: FR-SUR-004
type SurrenderDocument struct {
	ID                 uuid.UUID              `json:"id" db:"id"`
	SurrenderRequestID uuid.UUID              `json:"surrender_request_id" db:"surrender_request_id"`
	DocumentType       DocumentType           `json:"document_type" db:"document_type"`
	DocumentName       string                 `json:"document_name" db:"document_name"`
	DocumentPath       string                 `json:"document_path" db:"document_path"`
	UploadedDate       time.Time              `json:"uploaded_date" db:"uploaded_date"`
	FileSizeBytes      *int                   `json:"file_size_bytes" db:"file_size_bytes"`
	MimeType           *string                `json:"mime_type" db:"mime_type"`
	Verified           bool                   `json:"verified" db:"verified"`
	VerifiedBy         *uuid.UUID             `json:"verified_by" db:"verified_by"`
	VerifiedAt         *time.Time             `json:"verified_at" db:"verified_at"`
	RejectionReason    *string                `json:"rejection_reason" db:"rejection_reason"`
	CreatedAt          time.Time              `json:"created_at" db:"created_at"`
	DeletedAt          *time.Time             `json:"deleted_at" db:"deleted_at"`
	Metadata           map[string]interface{} `json:"metadata" db:"metadata"`
}
