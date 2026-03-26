// Package response contains HTTP response DTOs for document management endpoints.
//
// Phase 3: Supporting Endpoints
// DM-001..005 response types.
package response

import (
	"customer-nfs-service/core/port"
	"time"
)

// ─── DM-001: POST /nfs/documents/upload ──────────────────────────────────────

// DocumentUploadData is the payload returned on a successful DM-001 call.
// VR-NFS-009: file size validated before this response is generated.
// VR-NFS-010: MIME type validated.
type DocumentUploadData struct {
	// DocumentID is the UUID of the newly stored document.
	DocumentID string `json:"document_id"`
	// RequestID links the document to its service request.
	RequestID string `json:"request_id"`
	// DocumentType is the classification of the uploaded document.
	DocumentType string `json:"document_type"`
	// FileName is the original file name as supplied by the uploader.
	FileName string `json:"file_name"`
	// FileSizeBytes is the size of the stored file in bytes.
	FileSizeBytes int64 `json:"file_size_bytes"`
	// MIMEType is the detected MIME type of the file.
	MIMEType string `json:"mime_type"`
	// StoragePath is the internal DMS path or object storage key.
	StoragePath string `json:"storage_path"`
	// UploadedAt is the ISO-8601 timestamp of the upload.
	UploadedAt string `json:"uploaded_at"`
	// VerificationStatus is always "PENDING" at upload time.
	VerificationStatus string `json:"verification_status"`
}

// DocumentUploadResponse is the top-level success response for DM-001.
type DocumentUploadResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	Data                      DocumentUploadData `json:"data"`
}

// NewDocumentUploadResponse constructs the DM-001 success response.
func NewDocumentUploadResponse(documentID, requestID, docType, fileName, storagePath, mimeType string, fileSize int64) *DocumentUploadResponse {
	return &DocumentUploadResponse{
		StatusCodeAndMessage: port.CreateSuccess,
		Data: DocumentUploadData{
			DocumentID:         documentID,
			RequestID:          requestID,
			DocumentType:       docType,
			FileName:           fileName,
			FileSizeBytes:      fileSize,
			MIMEType:           mimeType,
			StoragePath:        storagePath,
			UploadedAt:         time.Now().UTC().Format(time.RFC3339),
			VerificationStatus: "PENDING",
		},
	}
}

// ─── DM-002: GET /nfs/documents/:document_id ──────────────────────────────────

// DocumentMetadataData is the payload returned on a successful DM-002 call.
type DocumentMetadataData struct {
	// DocumentID is the UUID of the document.
	DocumentID string `json:"document_id"`
	// RequestID links the document to its service request.
	RequestID string `json:"request_id"`
	// DocumentType is the classification of the document.
	DocumentType string `json:"document_type"`
	// FileName is the original file name.
	FileName string `json:"file_name"`
	// FileSizeBytes is the size of the stored file.
	FileSizeBytes int64 `json:"file_size_bytes"`
	// MIMEType is the MIME type of the document.
	MIMEType string `json:"mime_type"`
	// VerificationStatus is PENDING, VERIFIED, or REJECTED.
	VerificationStatus string `json:"verification_status"`
	// UploadedBy is the user ID who uploaded the document.
	UploadedBy string `json:"uploaded_by"`
	// UploadedAt is the ISO-8601 timestamp of the upload.
	UploadedAt string `json:"uploaded_at"`
	// VerifiedBy is the CPC officer who verified the document (if verified).
	VerifiedBy *string `json:"verified_by,omitempty"`
	// VerifiedAt is the ISO-8601 timestamp of verification (if verified).
	VerifiedAt *string `json:"verified_at,omitempty"`
	// RejectionReason explains why the document was rejected (if rejected).
	RejectionReason *string `json:"rejection_reason,omitempty"`
}

// DocumentMetadataResponse is the top-level success response for DM-002.
type DocumentMetadataResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	Data                      DocumentMetadataData `json:"data"`
}

// ─── DM-003: GET /nfs/documents/:document_id/download ────────────────────────

// DocumentDownloadData returns file metadata; actual bytes are written to the response stream.
// In practice, the handler writes PDF/image bytes with the appropriate Content-Type header.
type DocumentDownloadData struct {
	// DocumentID is the UUID of the document.
	DocumentID string `json:"document_id"`
	// FileName is the original file name.
	FileName string `json:"file_name"`
	// MIMEType is the MIME type of the document.
	MIMEType string `json:"mime_type"`
	// FileSizeBytes is the size of the file.
	FileSizeBytes int64 `json:"file_size_bytes"`
	// DownloadURL is a pre-signed DMS URL for retrieving the file (if object storage).
	DownloadURL string `json:"download_url"`
}

// DocumentDownloadResponse is the top-level success response for DM-003.
type DocumentDownloadResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	Data                      DocumentDownloadData `json:"data"`
}

// ─── DM-004: DELETE /nfs/documents/:document_id ──────────────────────────────

// DocumentDeleteData is the payload returned on a successful DM-004 call.
// DM-004: only CREATED status documents can be deleted.
type DocumentDeleteData struct {
	// DocumentID is the UUID of the deleted document.
	DocumentID string `json:"document_id"`
	// Deleted indicates whether the deletion was successful.
	Deleted bool `json:"deleted"`
	// Message provides a human-readable outcome.
	Message string `json:"message"`
}

// DocumentDeleteResponse is the top-level success response for DM-004.
type DocumentDeleteResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	Data                      DocumentDeleteData `json:"data"`
}

// NewDocumentDeleteResponse constructs the DM-004 response.
func NewDocumentDeleteResponse(documentID string, deleted bool) *DocumentDeleteResponse {
	msg := "Document deleted successfully."
	if !deleted {
		msg = "Document could not be deleted."
	}
	return &DocumentDeleteResponse{
		StatusCodeAndMessage: port.DeleteSuccess,
		Data:                 DocumentDeleteData{DocumentID: documentID, Deleted: deleted, Message: msg},
	}
}

// ─── DM-005: POST /nfs/secure-upload/:token ──────────────────────────────────

// SecureUploadData is the payload returned on a successful DM-005 call.
// BR-NFS-014: acknowledge upload received; status → PENDING_APPROVAL after signal.
type SecureUploadData struct {
	// DocumentID is the UUID of the uploaded document.
	DocumentID string `json:"document_id"`
	// Message provides a human-readable outcome.
	Message string `json:"message"`
	// NewStatus reflects the request status after the upload signal (PENDING_APPROVAL).
	NewStatus string `json:"new_status"`
	// AttemptNumber tracks which upload attempt this was (max 3 per BR-NFS-014).
	AttemptNumber int `json:"attempt_number"`
}

// SecureUploadResponse is the top-level success response for DM-005.
type SecureUploadResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	Data                      SecureUploadData `json:"data"`
}

// NewSecureUploadResponse constructs the DM-005 response.
func NewSecureUploadResponse(documentID, newStatus string, attempt int) *SecureUploadResponse {
	return &SecureUploadResponse{
		StatusCodeAndMessage: port.CreateSuccess,
		Data: SecureUploadData{
			DocumentID:    documentID,
			Message:       "Document uploaded successfully. Your request has been forwarded for review.",
			NewStatus:     newStatus,
			AttemptNumber: attempt,
		},
	}
}
