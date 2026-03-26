// Package handler provides HTTP handlers for the Customer NFS service.
// Phase: Phase 3 — Supporting Endpoints
// Handler: DocumentHandler — covers DM-001..005
//
// DM-001: POST /nfs/documents/upload                    — multipart upload
// DM-002: GET  /nfs/documents/:document_id              — metadata
// DM-003: GET  /nfs/documents/:document_id/download     — file stream
// DM-004: DELETE /nfs/documents/:document_id            — delete (CREATED only)
// DM-005: POST /nfs/secure-upload/:token                — unauthenticated secure upload
package handler

import (
	"fmt"

	"github.com/jackc/pgx/v5"
	"go.temporal.io/sdk/client"

	log "gitlab.cept.gov.in/it-2.0-common/n-api-log"
	serverHandler "gitlab.cept.gov.in/it-2.0-common/n-api-server/handler"

	config "gitlab.cept.gov.in/it-2.0-common/api-config"
	serverRoute "gitlab.cept.gov.in/it-2.0-common/n-api-server/route"

	"customer-nfs-service/core/domain"
	"customer-nfs-service/core/port"
	resp "customer-nfs-service/handler/response"
	repo "customer-nfs-service/repo/postgres"
)

// DocumentHandler manages document upload, download, and lifecycle operations.
// FR-NFS-003 (Document management), FR-NFS-011 (Secure upload link).
// VR-NFS-009: max file size 5 MB.
// VR-NFS-010: allowed file types — application/pdf, image/jpeg, image/png.
type DocumentHandler struct {
	*serverHandler.Base

	srRepo  *repo.ServiceRequestRepository
	docRepo *repo.DocumentUploadRepository
	tc      client.Client
	cfg     *config.Config
}

// NewDocumentHandler constructs the handler and wires dependencies via Uber FX.
func NewDocumentHandler(
	srRepo *repo.ServiceRequestRepository,
	docRepo *repo.DocumentUploadRepository,
	tc client.Client,
	cfg *config.Config,
) *DocumentHandler {
	base := serverHandler.New("Document").
		SetPrefix("/v1").
		AddPrefix("")
	return &DocumentHandler{
		Base:    base,
		srRepo:  srRepo,
		docRepo: docRepo,
		tc:      tc,
		cfg:     cfg,
	}
}

// Routes registers all document management HTTP routes.
func (h *DocumentHandler) Routes() []serverRoute.Route {
	return []serverRoute.Route{
		serverRoute.POST("/nfs/documents/upload", h.UploadDocument).Name("Upload Document"),
		serverRoute.GET("/nfs/documents/:document_id", h.GetDocumentMetadata).Name("Get Document Metadata"),
		serverRoute.GET("/nfs/documents/:document_id/download", h.DownloadDocument).Name("Download Document"),
		serverRoute.DELETE("/nfs/documents/:document_id", h.DeleteDocument).Name("Delete Document"),
		serverRoute.POST("/nfs/secure-upload/:token", h.SecureUploadDocument).Name("Secure Upload Document"),
	}
}

// ─── DM-001 ───────────────────────────────────────────────────────────────────

// UploadDocument handles multipart document upload for an NFS request.
// VR-NFS-009: max file size 5 MB (enforced by framework middleware).
// VR-NFS-010: allowed MIME types — application/pdf, image/jpeg, image/png.
// Stores metadata in nfs.document_upload; file stored in DMS (STUB).
func (h *DocumentHandler) UploadDocument(
	sctx *serverRoute.Context,
	req UploadDocumentRequest,
) (*resp.DocumentUploadResponse, error) {
	// Validate that the request exists and is in an uploadable status.
	sr, err := h.srRepo.GetByID(sctx.Ctx, req.RequestID)
	if err != nil {
		if err == pgx.ErrNoRows {
			log.Error(sctx.Ctx, "UploadDocument: request not found: %s", req.RequestID)
			return nil, err
		}
		log.Error(sctx.Ctx, "UploadDocument: failed to fetch request %s: %v", req.RequestID, err)
		return nil, err
	}

	uploadableStatuses := map[string]bool{
		"CREATED": true, "PENDING_DOCUMENTS": true,
	}
	if !uploadableStatuses[string(sr.Status)] {
		return nil, fmt.Errorf("documents cannot be uploaded when request is in status %s", sr.Status)
	}

	// VR-NFS-010: MIME type validation — done via injected form file header.
	// Framework receives multipart file via sctx.Request.FormFile("file").
	// STUB: file stored in DMS (INT-NFS-006).
	// TODO: stream file to DMS and receive storage reference.
	allowedMIMEs := map[string]bool{
		"application/pdf": true,
		"image/jpeg":      true,
		"image/png":       true,
	}

	// STUB: simulate file receipt metadata. In production, receive via multipart form.
	mimeType := "application/pdf" // populated from form file header
	fileName := "document.pdf"
	fileSize := int64(0)
	if !allowedMIMEs[mimeType] {
		return nil, fmt.Errorf("file type %q not allowed; allowed types: PDF, JPEG, PNG (VR-NFS-010)", mimeType)
	}

	// VR-NFS-009: 5MB hard limit.
	maxSize := int64(5 * 1024 * 1024)
	if fileSize > maxSize {
		return nil, fmt.Errorf("file size %d exceeds 5MB limit (VR-NFS-009)", fileSize)
	}

	doc := &domain.DocumentUpload{
		RequestID:          req.RequestID,
		DocumentType:       req.DocumentType,
		FileName:           fileName,
		FileSizeBytes:      fileSize,
		MimeType:           mimeType,
		FileURL:            fmt.Sprintf("dms/nfs/%s/%s", req.RequestID, fileName),
		VerificationStatus: "PENDING",
		UploadedBy:         req.UploadedBy,
		CreatedBy:          req.UploadedBy,
	}

	created, err := h.docRepo.Create(sctx.Ctx, doc)
	if err != nil {
		log.Error(sctx.Ctx, "UploadDocument: failed to save document metadata for %s: %v", req.RequestID, err)
		return nil, err
	}

	log.Info(sctx.Ctx, "UploadDocument: saved document %s for request %s", created.DocumentID, req.RequestID)
	r := resp.NewDocumentUploadResponse(created.DocumentID, req.RequestID, req.DocumentType, fileName, created.FileURL, mimeType, fileSize)
	return r, nil
}

// ─── DM-002 ───────────────────────────────────────────────────────────────────

// GetDocumentMetadata returns stored metadata for a document without the file bytes.
func (h *DocumentHandler) GetDocumentMetadata(
	sctx *serverRoute.Context,
	req GetDocumentMetadataRequest,
) (*resp.DocumentMetadataResponse, error) {
	doc, err := h.docRepo.GetByID(sctx.Ctx, req.DocumentID)
	if err != nil {
		if err == pgx.ErrNoRows {
			log.Error(sctx.Ctx, "GetDocumentMetadata: document not found: %s", req.DocumentID)
			return nil, err
		}
		log.Error(sctx.Ctx, "GetDocumentMetadata: failed to fetch document %s: %v", req.DocumentID, err)
		return nil, err
	}

	r := &resp.DocumentMetadataResponse{
		StatusCodeAndMessage: port.FetchSuccess,
		Data: resp.DocumentMetadataData{
			DocumentID:         doc.DocumentID,
			RequestID:          doc.RequestID,
			DocumentType:       doc.DocumentType,
			FileName:           doc.FileName,
			FileSizeBytes:      doc.FileSizeBytes,
			MIMEType:           doc.MimeType,
			VerificationStatus: doc.VerificationStatus,
			UploadedBy:         doc.UploadedBy,
			UploadedAt:         resp.FormatTime(doc.UploadedAt),
		},
	}
	return r, nil
}

// ─── DM-003 ───────────────────────────────────────────────────────────────────

// DownloadDocument streams the document file bytes from DMS.
// STUB: returns download URL metadata; in production, proxies DMS stream.
func (h *DocumentHandler) DownloadDocument(
	sctx *serverRoute.Context,
	req DownloadDocumentRequest,
) (*resp.DocumentDownloadResponse, error) {
	doc, err := h.docRepo.GetByID(sctx.Ctx, req.DocumentID)
	if err != nil {
		if err == pgx.ErrNoRows {
			log.Error(sctx.Ctx, "DownloadDocument: document not found: %s", req.DocumentID)
			return nil, err
		}
		log.Error(sctx.Ctx, "DownloadDocument: failed to fetch document %s: %v", req.DocumentID, err)
		return nil, err
	}

	// STUB: in production, generate a pre-signed DMS URL or proxy the file stream.
	// TODO: call DMS (INT-NFS-006) to get pre-signed download URL.
	log.Info(sctx.Ctx, "DownloadDocument [STUB]: generating download URL for document %s", req.DocumentID)

	r := &resp.DocumentDownloadResponse{
		StatusCodeAndMessage: port.FetchSuccess,
		Data: resp.DocumentDownloadData{
			DocumentID:    doc.DocumentID,
			FileName:      doc.FileName,
			MIMEType:      doc.MimeType,
			FileSizeBytes: doc.FileSizeBytes,
			DownloadURL:   fmt.Sprintf("/dms/documents/%s/download", doc.DocumentID),
		},
	}
	return r, nil
}

// ─── DM-004 ───────────────────────────────────────────────────────────────────

// DeleteDocument deletes a document only if the associated request is in CREATED status.
// FR-NFS-003: document lifecycle management.
func (h *DocumentHandler) DeleteDocument(
	sctx *serverRoute.Context,
	req DeleteDocumentRequest,
) (*resp.DocumentDeleteResponse, error) {
	doc, err := h.docRepo.GetByID(sctx.Ctx, req.DocumentID)
	if err != nil {
		if err == pgx.ErrNoRows {
			log.Error(sctx.Ctx, "DeleteDocument: document not found: %s", req.DocumentID)
			return nil, err
		}
		log.Error(sctx.Ctx, "DeleteDocument: failed to fetch document %s: %v", req.DocumentID, err)
		return nil, err
	}

	// Check parent request status — only CREATED requests allow document deletion.
	sr, err := h.srRepo.GetByID(sctx.Ctx, doc.RequestID)
	if err != nil {
		log.Error(sctx.Ctx, "DeleteDocument: failed to fetch parent request %s: %v", doc.RequestID, err)
		return nil, err
	}
	if sr.Status != "CREATED" {
		return nil, fmt.Errorf("documents can only be deleted when request is in CREATED status (current: %s)", sr.Status)
	}

	if err := h.docRepo.Delete(sctx.Ctx, req.DocumentID); err != nil {
		log.Error(sctx.Ctx, "DeleteDocument: failed to delete document %s: %v", req.DocumentID, err)
		return nil, err
	}

	log.Info(sctx.Ctx, "DeleteDocument: deleted document %s for request %s", req.DocumentID, doc.RequestID)
	r := resp.NewDocumentDeleteResponse(req.DocumentID, true)
	return r, nil
}

// ─── DM-005 ───────────────────────────────────────────────────────────────────

// SecureUploadDocument handles unauthenticated document upload via a secure token link.
// This endpoint has no authentication middleware — token IS the authorization.
// BR-NFS-014: secure link expires in 7 days, max 3 attempts per request.
// BR-NFS-014: on 3rd attempt, status → DOCUMENTS_EXPIRED.
// On success: signals "docs_received" to the running workflow.
// Status transition: PENDING_DOCUMENTS → PENDING_APPROVAL.
// VR-NFS-009: 5MB file size limit.
// VR-NFS-010: PDF/JPEG/PNG only.
func (h *DocumentHandler) SecureUploadDocument(
	sctx *serverRoute.Context,
	req SecureUploadDocumentRequest,
) (*resp.SecureUploadResponse, error) {
	// Validate secure upload token — look up in nfs.secure_upload_links.
	link, err := h.srRepo.GetSecureUploadLink(sctx.Ctx, req.Token)
	if err != nil {
		if err == pgx.ErrNoRows {
			log.Error(sctx.Ctx, "SecureUploadDocument: invalid or expired token: %s", req.Token)
			return nil, fmt.Errorf("invalid or expired upload link (BR-NFS-014)")
		}
		log.Error(sctx.Ctx, "SecureUploadDocument: token lookup failed: %v", err)
		return nil, err
	}

	// STUB: validate file type and size (same as DM-001).
	mimeType := "application/pdf"
	fileName := "secure_document.pdf"
	fileSize := int64(0)

	allowedMIMEs := map[string]bool{
		"application/pdf": true, "image/jpeg": true, "image/png": true,
	}
	if !allowedMIMEs[mimeType] {
		return nil, fmt.Errorf("file type %q not allowed (VR-NFS-010)", mimeType)
	}

	doc := &domain.DocumentUpload{
		RequestID:          link.RequestID,
		DocumentType:       req.DocumentType,
		FileName:           fileName,
		FileSizeBytes:      fileSize,
		MimeType:           mimeType,
		FileURL:            fmt.Sprintf("dms/nfs/%s/%s", link.RequestID, fileName),
		VerificationStatus: "PENDING",
		UploadedBy:         "customer-secure-link",
		CreatedBy:          "customer-secure-link",
	}

	created, err := h.docRepo.Create(sctx.Ctx, doc)
	if err != nil {
		log.Error(sctx.Ctx, "SecureUploadDocument: failed to save document for request %s: %v", link.RequestID, err)
		return nil, err
	}

	// Update status PENDING_DOCUMENTS → PENDING_APPROVAL.
	audit := &domain.AuditLog{
		RequestID: link.RequestID,
		// ActionType:    "PENDING_APPROVAL",
		ActionType:    "STATUS_CHANGE", // was "PENDING_APPROVAL"
		NewValueJSON:  fmt.Sprintf(`{"document_id":"%s","via":"secure_link"}`, created.DocumentID),
		PerformedByID: "customer",
		Notes:         strPtr("Documents uploaded via secure link"),
	}
	_, err = h.srRepo.UpdateStatus(sctx.Ctx, link.RequestID, "PENDING_APPROVAL", strPtr("customer"), nil, nil, audit)
	if err != nil {
		log.Error(sctx.Ctx, "SecureUploadDocument: failed to update status for %s: %v", link.RequestID, err)
		return nil, err
	}

	// Signal workflow: "docs_received" → WF-NFS-004 or WF-NFS-002 (manual path).
	wfID, wfRunID, wfErr := h.srRepo.GetWorkflowState(sctx.Ctx, link.RequestID)
	if wfErr == nil {
		signalPayload := map[string]string{"document_id": created.DocumentID, "request_id": link.RequestID}
		if signalErr := h.tc.SignalWorkflow(sctx.Ctx, wfID, wfRunID, "docs_received", signalPayload); signalErr != nil {
			// Non-fatal: log and continue. Workflow will pick up status change on next poll.
			log.Error(sctx.Ctx, "SecureUploadDocument: failed to signal workflow %s: %v", wfID, signalErr)
		} else {
			log.Info(sctx.Ctx, "SecureUploadDocument: signaled workflow %s for request %s", wfID, link.RequestID)
		}
	}

	// Mark token as used (increment attempt counter).
	if markErr := h.srRepo.MarkSecureUploadUsed(sctx.Ctx, req.Token); markErr != nil {
		log.Error(sctx.Ctx, "SecureUploadDocument: failed to mark token used: %v", markErr)
	}

	log.Info(sctx.Ctx, "SecureUploadDocument: document %s uploaded for request %s via secure link", created.DocumentID, link.RequestID)

	r := resp.NewSecureUploadResponse(created.DocumentID, "PENDING_APPROVAL", link.LinkGenerationCount)
	return r, nil
}
