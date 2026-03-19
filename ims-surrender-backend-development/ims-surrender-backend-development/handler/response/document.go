package response

import (
	"time"

	"gitlab.cept.gov.in/it-2.0-policy/surrender-service/core/domain"
	"gitlab.cept.gov.in/it-2.0-policy/surrender-service/core/port"
)

// ============================================
// Document Upload Response DTOs
// ============================================

// UploadDocumentResponse represents document upload result
type UploadDocumentResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	Data                      DocumentUploadData `json:"data"`
}

type DocumentUploadData struct {
	DocumentID         string `json:"document_id"`
	SurrenderRequestID string `json:"surrender_request_id"`
	DocumentType       string `json:"document_type"`
	DocumentName       string `json:"document_name"`
	FileSizeBytes      int    `json:"file_size_bytes"`
	UploadedDate       string `json:"uploaded_date"`
	Verified           bool   `json:"verified"`
	UploadStatus       string `json:"upload_status"`
}

// ============================================
// Document Status Response DTOs
// ============================================

// DocumentStatusResponse represents document upload status
type DocumentStatusResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	Data                      DocumentStatusData `json:"data"`
}

type DocumentStatusData struct {
	SurrenderRequestID   string             `json:"surrender_request_id"`
	TotalRequired        int                `json:"total_required"`
	TotalUploaded        int                `json:"total_uploaded"`
	TotalVerified        int                `json:"total_verified"`
	AllDocumentsUploaded bool               `json:"all_documents_uploaded"`
	AllDocumentsVerified bool               `json:"all_documents_verified"`
	CanSubmit            bool               `json:"can_submit"`
	Documents            []DocumentInfoData `json:"documents"`
}

type DocumentInfoData struct {
	DocumentID      string  `json:"document_id"`
	DocumentType    string  `json:"document_type"`
	DisplayName     string  `json:"display_name"`
	DocumentName    string  `json:"document_name"`
	FileSizeBytes   int     `json:"file_size_bytes"`
	UploadedDate    string  `json:"uploaded_date"`
	Verified        bool    `json:"verified"`
	VerifiedBy      *string `json:"verified_by,omitempty"`
	VerifiedAt      *string `json:"verified_at,omitempty"`
	RejectionReason *string `json:"rejection_reason,omitempty"`
	Status          string  `json:"status"`
}

// ============================================
// Submit for Verification Response DTOs
// ============================================

// SubmitForVerificationResponse represents submission result
type SubmitForVerificationResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	Data                      SubmitVerificationData `json:"data"`
}

type SubmitVerificationData struct {
	SurrenderRequestID string            `json:"surrender_request_id"`
	RequestNumber      string            `json:"request_number"`
	OldStatus          string            `json:"old_status"`
	NewStatus          string            `json:"new_status"`
	SubmittedAt        string            `json:"submitted_at"`
	WorkflowState      WorkflowStateData `json:"workflow_state"`
	NextAction         NextActionData    `json:"next_action"`
}

// ============================================
// Helper Functions
// ============================================

// NewDocumentResponse converts domain document to response DTO
func NewDocumentResponse(d domain.SurrenderDocument) DocumentInfoData {
	status := "UPLOADED"
	if d.Verified {
		status = "VERIFIED"
	} else if d.RejectionReason != nil {
		status = "REJECTED"
	}

	var verifiedBy *string
	var verifiedAt *string
	if d.VerifiedBy != nil {
		vb := d.VerifiedBy.String()
		verifiedBy = &vb
	}
	if d.VerifiedAt != nil {
		va := d.VerifiedAt.Format(time.RFC3339)
		verifiedAt = &va
	}

	var fileSizeBytes int
	if d.FileSizeBytes != nil {
		fileSizeBytes = *d.FileSizeBytes
	}

	return DocumentInfoData{
		DocumentID:      d.ID.String(),
		DocumentType:    string(d.DocumentType),
		DisplayName:     getDocumentDisplayName(d.DocumentType),
		DocumentName:    d.DocumentName,
		FileSizeBytes:   fileSizeBytes,
		UploadedDate:    d.UploadedDate.Format(time.RFC3339),
		Verified:        d.Verified,
		VerifiedBy:      verifiedBy,
		VerifiedAt:      verifiedAt,
		RejectionReason: d.RejectionReason,
		Status:          status,
	}
}

// NewDocumentsResponse converts slice of domain documents to response DTOs
func NewDocumentsResponse(data []domain.SurrenderDocument) []DocumentInfoData {
	res := make([]DocumentInfoData, 0, len(data))
	for _, d := range data {
		res = append(res, NewDocumentResponse(d))
	}
	return res
}

// getDocumentDisplayName returns user-friendly document name
func getDocumentDisplayName(docType domain.DocumentType) string {
	displayNames := map[domain.DocumentType]string{
		domain.DocumentTypeWrittenConsent:     "Written Consent",
		domain.DocumentTypePolicyBond:         "Policy Bond",
		domain.DocumentTypePremiumReceiptBook: "Premium Receipt Book",
		domain.DocumentTypePayRecoveryCert:    "Pay Recovery Certificate",
		domain.DocumentTypeLoanReceiptBook:    "Loan Receipt Book",
		domain.DocumentTypeLoanBond:           "Loan Bond",
		domain.DocumentTypeIndemnityBond:      "Indemnity Bond",
		domain.DocumentTypeAssignmentDeed:     "Assignment Deed",
		domain.DocumentTypeDischargeReceipt:   "Discharge Receipt",
	}

	if name, ok := displayNames[docType]; ok {
		return name
	}
	return string(docType)
}
