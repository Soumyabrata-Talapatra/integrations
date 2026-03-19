package response

import (
	"time"

	"policy-issue-service/core/port"
)

// ============================================
// DOC-POL-001: Document Checklist Response
// ============================================

// DocumentChecklistItem represents a single required document in the checklist
type DocumentChecklistItem struct {
	DocumentType   string     `json:"document_type"`
	SubType        string     `json:"sub_type,omitempty"`          // Differentiator when the same enum type appears more than once (e.g. "PROPOSER_PHOTO_ID", "PROPOSER_ADDRESS_PROOF")
	IsMandatory    bool       `json:"is_mandatory"`
	IsUploaded     bool       `json:"is_uploaded"`
	DocumentID     *string    `json:"document_id,omitempty"`
	UploadDate     *time.Time `json:"upload_date,omitempty"`
	ReasonRequired string     `json:"reason_required"`
}

// DocumentChecklistResponse represents the response for DOC-POL-001
// [DOC-POL-001] Dynamic document checklist
// Components: FR-POL-029
type DocumentChecklistResponse struct {
	port.StatusCodeAndMessage
	ProposalID           int64                   `json:"proposal_id"`
	RequiredDocuments    []DocumentChecklistItem  `json:"required_documents"`
	CompletionPercentage int                      `json:"completion_percentage"`
}

// ============================================
// DOC-POL-002: Document Upload Response
// ============================================

// DocumentUploadResponse represents the response for DOC-POL-002
// [DOC-POL-002] Upload document
// Integration: INT-POL-008 (DMS)
type DocumentUploadResponse struct {
	port.StatusCodeAndMessage
	DocRefID          int64     `json:"document_ref_id"`
	DocumentID        string    `json:"document_id"`
	DocumentType      string    `json:"document_type"`
	FileName          string    `json:"file_name"`
	UploadedAt        time.Time `json:"uploaded_at"`
	DownloadURL       string    `json:"download_url"`
	IsMissingNotation bool      `json:"is_missing_notation"`
}

// ============================================
// DOC-POL-005: Missing Documents List Response
// ============================================

// MissingDocumentItem represents a single missing document notation
type MissingDocumentItem struct {
	MissingDocID        int64      `json:"missing_doc_id"`
	ProposalID          int64      `json:"proposal_id"`
	DocumentType        string     `json:"document_type"`
	DocumentDescription *string    `json:"document_description,omitempty"`
	ReasonMissing       *string    `json:"reason_missing,omitempty"`
	Stage               string     `json:"stage"`
	NotedBy             int64      `json:"noted_by"`
	NotedAt             time.Time  `json:"noted_at"`
	Notes               *string    `json:"notes,omitempty"`
	Status              string     `json:"status"`
	ResolvedBy          *int64     `json:"resolved_by,omitempty"`
	ResolvedAt          *time.Time `json:"resolved_at,omitempty"`
	ResolutionNotes     *string    `json:"resolution_notes,omitempty"`
	UploadedDocumentID  *int64     `json:"uploaded_document_id,omitempty"`
	Waived              bool       `json:"waived"`
	WaivedBy            *int64     `json:"waived_by,omitempty"`
	WaivedAt            *time.Time `json:"waived_at,omitempty"`
	WaiverReason        *string    `json:"waiver_reason,omitempty"`
	FollowUpRequired    bool       `json:"follow_up_required"`
}

// MissingDocumentsListResponse represents the response for DOC-POL-005
// [DOC-POL-005] Get missing documents list
type MissingDocumentsListResponse struct {
	port.StatusCodeAndMessage
	ProposalID       int64                 `json:"proposal_id"`
	ProposalNumber   string                `json:"proposal_number"`
	TotalMissing     int                   `json:"total_missing"`
	PendingCount     int                   `json:"pending_count"`
	UploadedCount    int                   `json:"uploaded_count"`
	WaivedCount      int                   `json:"waived_count"`
	MissingDocuments []MissingDocumentItem `json:"missing_documents"`
}

// ============================================
// DOC-POL-006 / DOC-POL-007: Missing Document Notation Response
// ============================================

// MissingDocumentNotationResponse represents the response for creating/resolving a missing document
// [DOC-POL-006] Record missing document, [DOC-POL-007] Resolve missing document
type MissingDocumentNotationResponse struct {
	port.StatusCodeAndMessage
	MissingDocumentItem
}

// ============================================
// STATUS-POL-001: Proposal Status Response
// ============================================

// ProposalStatusResponse represents the response for STATUS-POL-001
// [STATUS-POL-001] Get proposal status
// Components: BR-POL-015
type ProposalStatusResponse struct {
	port.StatusCodeAndMessage
	ProposalID        int64     `json:"proposal_id"`
	ProposalNumber    string    `json:"proposal_number"`
	Status            string    `json:"status"`
	StatusDescription string    `json:"status_description"`
	LastUpdated       time.Time `json:"last_updated"`
}

// ============================================
// STATUS-POL-002: Proposal Timeline Response
// ============================================

// TimelineEntry represents a single entry in the proposal timeline
type TimelineEntry struct {
	Step       string     `json:"step"`
	Status     string     `json:"status"`
	FromStatus *string    `json:"from_status,omitempty"`
	ToStatus   string     `json:"to_status"`
	Timestamp  *time.Time `json:"timestamp,omitempty"`
	Actor      string     `json:"actor"`
	Comments   *string    `json:"comments,omitempty"`
	Duration   string     `json:"duration,omitempty"`
}

// ProposalTimelineResponse represents the response for STATUS-POL-002
// [STATUS-POL-002] Proposal timeline
// Components: FR-POL-033
type ProposalTimelineResponse struct {
	port.StatusCodeAndMessage
	ProposalID     int64           `json:"proposal_id"`
	ProposalNumber string          `json:"proposal_number"`
	Timeline       []TimelineEntry `json:"timeline"`
}

// ============================================
// STATUS-POL-003: Policy Status Response
// ============================================

// PolicyStatusResponse represents the response for STATUS-POL-003
// [STATUS-POL-003] Policy status
// Components: FR-POL-025
type PolicyStatusResponse struct {
	port.StatusCodeAndMessage
	PolicyID          int64      `json:"policy_id"`
	PolicyNumber      string     `json:"policy_number"`
	Status            string     `json:"status"`
	StatusDescription string     `json:"status_description"`
	EffectiveDate     *time.Time `json:"effective_date,omitempty"`
	LastUpdated       time.Time  `json:"last_updated"`
}

// ============================================
// Generic Delete Response
// ============================================

// DeleteResponse represents the response for DOC-POL-004
// [DOC-POL-004] Remove document
type DeleteResponse struct {
	port.StatusCodeAndMessage
}
