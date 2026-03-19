package domain

import "time"

// ============================================
// Document Type Enum (E-016: proposal_document_ref)
// ============================================

// DocumentType represents the type of document
type DocumentType string

const (
	DocumentTypeProposalForm     DocumentType = "PROPOSAL_FORM"
	DocumentTypeDOBProof         DocumentType = "DOB_PROOF"
	DocumentTypeAddressProof     DocumentType = "ADDRESS_PROOF"
	DocumentTypePhotoID          DocumentType = "PHOTO_ID"
	DocumentTypeMedicalReport    DocumentType = "MEDICAL_REPORT"
	DocumentTypePaymentCopy      DocumentType = "PAYMENT_COPY"
	DocumentTypeHealthDeclaration DocumentType = "HEALTH_DECLARATION"
	DocumentTypePhoto            DocumentType = "PHOTO"
	DocumentTypeIncomeProof      DocumentType = "INCOME_PROOF"
	DocumentTypeEmploymentProof  DocumentType = "EMPLOYMENT_PROOF"
	DocumentTypeOther            DocumentType = "OTHER"
)

// ============================================
// Missing Document Stage Enum
// ============================================

// MissingDocumentStage represents the stage where document was noted as missing
type MissingDocumentStage string

const (
	MissingDocStageQCReview MissingDocumentStage = "QC_REVIEW"
	MissingDocStageApproval MissingDocumentStage = "APPROVAL"
)

// ============================================
// Missing Document Status Enum
// ============================================

// MissingDocumentStatus represents the resolution status of a missing document
type MissingDocumentStatus string

const (
	MissingDocStatusPending  MissingDocumentStatus = "PENDING"
	MissingDocStatusUploaded MissingDocumentStatus = "UPLOADED"
	MissingDocStatusWaived   MissingDocumentStatus = "WAIVED"
)

// ============================================
// E-016: ProposalDocumentRef
// ============================================

// ProposalDocumentRef represents a document reference uploaded for a proposal
type ProposalDocumentRef struct {
	DocRefID      int64        `db:"doc_ref_id" json:"doc_ref_id"`
	ProposalID    int64        `db:"proposal_id" json:"proposal_id"`
	DocumentID    string       `db:"document_id" json:"document_id"`
	DocumentType  DocumentType `db:"document_type" json:"document_type"`
	FileName      *string      `db:"file_name" json:"file_name,omitempty"`
	FileSizeBytes *int64       `db:"file_size_bytes" json:"file_size_bytes,omitempty"`
	MimeType      *string      `db:"mime_type" json:"mime_type,omitempty"`
	DocumentDate  *time.Time   `db:"document_date" json:"document_date,omitempty"` // VR-PI-023: date on the document, must not be in the future
	UploadedBy    int64        `db:"uploaded_by" json:"uploaded_by"`
	UploadedAt    time.Time    `db:"uploaded_at" json:"uploaded_at"`
	Version       int          `db:"version" json:"version"`
	Comments      *string      `db:"comments" json:"comments,omitempty"`
	DeletedAt     *time.Time   `db:"deleted_at" json:"deleted_at,omitempty"`
}

// ============================================
// E-017: ProposalMissingDocument
// ============================================

// ProposalMissingDocument represents a missing document noted during QC/Approval review
type ProposalMissingDocument struct {
	MissingDocID        int64                 `db:"missing_doc_id" json:"missing_doc_id"`
	ProposalID          int64                 `db:"proposal_id" json:"proposal_id"`
	DocumentType        DocumentType          `db:"document_type" json:"document_type"`
	DocumentDescription *string               `db:"document_description" json:"document_description,omitempty"`
	Stage               MissingDocumentStage  `db:"stage" json:"stage"`
	NotedBy             int64                 `db:"noted_by" json:"noted_by"`
	NotedAt             time.Time             `db:"noted_at" json:"noted_at"`
	Notes               *string               `db:"notes" json:"notes,omitempty"`
	Status              MissingDocumentStatus `db:"status" json:"status"`
	ResolvedBy          *int64                `db:"resolved_by" json:"resolved_by,omitempty"`
	ResolvedAt          *time.Time            `db:"resolved_at" json:"resolved_at,omitempty"`
	ResolutionNotes     *string               `db:"resolution_notes" json:"resolution_notes,omitempty"`
	UploadedDocumentID  *int64                `db:"uploaded_document_id" json:"uploaded_document_id,omitempty"`
	Waived              bool                  `db:"waived" json:"waived"`
	WaivedBy            *int64                `db:"waived_by" json:"waived_by,omitempty"`
	WaivedAt            *time.Time            `db:"waived_at" json:"waived_at,omitempty"`
	WaiverReason        *string               `db:"waiver_reason" json:"waiver_reason,omitempty"`
	CreatedAt           time.Time             `db:"created_at" json:"created_at"`
	UpdatedAt           time.Time             `db:"updated_at" json:"updated_at"`
}

// ============================================
// E-015: ProposalStatusHistory
// ============================================

// ProposalStatusHistory represents a status change audit trail entry
type ProposalStatusHistory struct {
	HistoryID  int64          `db:"history_id" json:"history_id"`
	ProposalID int64          `db:"proposal_id" json:"proposal_id"`
	FromStatus *ProposalStatus `db:"from_status" json:"from_status,omitempty"`
	ToStatus   ProposalStatus `db:"to_status" json:"to_status"`
	ChangedBy  int64          `db:"changed_by" json:"changed_by"`
	ChangedAt  time.Time      `db:"changed_at" json:"changed_at"`
	Comments   *string        `db:"comments" json:"comments,omitempty"`
	Version    int            `db:"version" json:"version"`
	Metadata   *string        `db:"metadata" json:"metadata,omitempty"`
}

// ============================================
// Required Documents Checklist Logic
// ============================================

// RequiredDocument represents a single required document in the checklist
type RequiredDocument struct {
	DocumentType   DocumentType `json:"document_type"`
	IsMandatory    bool         `json:"is_mandatory"`
	IsUploaded     bool         `json:"is_uploaded"`
	DocumentID     *string      `json:"document_id,omitempty"`
	UploadDate     *time.Time   `json:"upload_date,omitempty"`
	ReasonRequired string       `json:"reason_required"`
}
