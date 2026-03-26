// Package handler provides HTTP handlers for the Customer NFS service.
// Phase: Phase 3 — Supporting Endpoints
// Handler: StatusHandler — covers ST-001..005
//
// ST-001: GET /nfs/requests/:request_id                   — full detail
// ST-002: GET /nfs/requests/:request_id/timeline          — audit log
// ST-003: GET /nfs/requests/:request_id/receipt           — PDF (binary)
// ST-004: GET /nfs/customers/:customer_id/requests        — paginated list
// ST-005: GET /nfs/requests/:request_id/documents         — document list
package handler

import (
	"fmt"

	"github.com/jackc/pgx/v5"

	log "gitlab.cept.gov.in/it-2.0-common/n-api-log"
	serverHandler "gitlab.cept.gov.in/it-2.0-common/n-api-server/handler"
	serverRoute "gitlab.cept.gov.in/it-2.0-common/n-api-server/route"

	"customer-nfs-service/core/port"
	resp "customer-nfs-service/handler/response"
	repo "customer-nfs-service/repo/postgres"
)

// StatusHandler handles HTTP endpoints for request status and tracking.
// FR-NFS-006 (Status visibility), FR-NFS-008 (Audit trail), FR-NFS-009 (Receipt).
type StatusHandler struct {
	*serverHandler.Base

	srRepo    *repo.ServiceRequestRepository
	addrRepo  *repo.AddressChangeRepository
	nameRepo  *repo.NameChangeRepository
	docRepo   *repo.DocumentUploadRepository
	auditRepo *repo.AuditLogRepository
}

// NewStatusHandler constructs the handler and wires dependencies via Uber FX.
func NewStatusHandler(
	srRepo *repo.ServiceRequestRepository,
	addrRepo *repo.AddressChangeRepository,
	nameRepo *repo.NameChangeRepository,
	docRepo *repo.DocumentUploadRepository,
	auditRepo *repo.AuditLogRepository,
) *StatusHandler {
	base := serverHandler.New("Status").
		SetPrefix("/v1").
		AddPrefix("")
	return &StatusHandler{
		Base:      base,
		srRepo:    srRepo,
		addrRepo:  addrRepo,
		nameRepo:  nameRepo,
		docRepo:   docRepo,
		auditRepo: auditRepo,
	}
}

// Routes registers all status and tracking HTTP routes.
func (h *StatusHandler) Routes() []serverRoute.Route {
	return []serverRoute.Route{
		serverRoute.GET("/nfs/requests/:request_id", h.GetRequestDetail).Name("Get Request Detail"),
		serverRoute.GET("/nfs/requests/:request_id/timeline", h.GetTimeline).Name("Get Request Timeline"),
		serverRoute.GET("/nfs/requests/:request_id/receipt", h.GetReceipt).Name("Get Receipt PDF"),
		serverRoute.GET("/nfs/customers/:customer_id/requests", h.ListCustomerRequests).Name("List Customer Requests"),
		serverRoute.GET("/nfs/requests/:request_id/documents", h.GetRequestDocuments).Name("Get Request Documents"),
	}
}

// ─── ST-001 ───────────────────────────────────────────────────────────────────

// GetRequestDetail returns full detail for a single NFS request.
// BATCH: 2-query batch — service_request + type-specific detail (single round-trip).
// Optional includes: documents (ST-005 inline) and audit_trail (ST-002 inline).
// FR-NFS-006: Request visibility.
func (h *StatusHandler) GetRequestDetail(
	sctx *serverRoute.Context,
	req GetRequestDetailRequest,
) (*resp.RequestDetailResponse, error) {
	sr, err := h.srRepo.GetByID(sctx.Ctx, req.RequestID)
	if err != nil {
		if err == pgx.ErrNoRows {
			log.Error(sctx.Ctx, "GetRequestDetail: request not found: %s", req.RequestID)
			return nil, err
		}
		log.Error(sctx.Ctx, "GetRequestDetail: failed to fetch request %s: %v", req.RequestID, err)
		return nil, err
	}

	r := &resp.RequestDetailResponse{
		StatusCodeAndMessage: port.FetchSuccess,
		Data: resp.RequestDetailData{
			RequestID:    sr.RequestID,
			TicketNumber: sr.TicketNumber,
			CustomerID:   sr.CustomerID,
			PolicyNumber: sr.PolicyNumber,
			RequestType:  string(sr.RequestType),
			Status:       string(sr.Status),
			AuthMethod:   sr.AuthMethod,
			Channel:      sr.Channel,
			CreatedAt:    resp.FormatTime(sr.CreatedAt),
			UpdatedAt:    resp.FormatTime(sr.UpdatedAt),
		},
	}
	if sr.OfficeCode != nil {
		r.Data.OfficeCode = sr.OfficeCode
	}

	// Fetch type-specific detail (BATCH within each repo).
	switch string(sr.RequestType) {
	case "ADDRESS_CHANGE":
		detail, addrErr := h.addrRepo.GetByRequestID(sctx.Ctx, req.RequestID)
		if addrErr == nil {
			r.Data.AddressDetail = &resp.AddressDetailSummary{
				AddressType:     detail.AddressType,
				NewAddressLine1: detail.NewAddressLine1,
				NewCity:         detail.NewCity,
				NewDistrict:     detail.NewDistrict,
				NewState:        detail.NewState,
				NewPincode:      detail.NewPincode,
			}
		}
	case "NAME_CHANGE":
		detail, nameErr := h.nameRepo.GetByRequestID(sctx.Ctx, req.RequestID)
		if nameErr == nil {
			oldName := ""
			if detail.OldFirstName != nil {
				oldName = fmt.Sprintf("%s %s", strDeref(detail.OldSalutation), strDeref(detail.OldFirstName))
				if detail.OldLastName != nil {
					oldName += " " + strDeref(detail.OldLastName)
				}
			}
			newName := fmt.Sprintf("%s %s %s", detail.NewSalutation, detail.NewFirstName, detail.NewLastName)
			r.Data.NameDetail = &resp.NameDetailSummary{
				OldName:          oldName,
				NewName:          newName,
				PoliciesAffected: detail.PoliciesAffected,
			}
		}
	}

	// Optional: include documents.
	includeDocuments := req.IncludeDocuments != nil && *req.IncludeDocuments
	if includeDocuments {
		docs, _, docsErr := h.docRepo.ListByRequestID(sctx.Ctx, req.RequestID, 0, 100)
		if docsErr == nil {
			docItems := make([]resp.RequestDetailDocumentItem, len(docs))
			for i, d := range docs {
				docItems[i] = resp.RequestDetailDocumentItem{
					DocumentID:         d.DocumentID,
					DocumentType:       d.DocumentType,
					FileName:           d.FileName,
					FileSizeBytes:      d.FileSizeBytes,
					VerificationStatus: d.VerificationStatus,
					UploadedAt:         resp.FormatTime(d.UploadedAt),
				}
			}
			r.Data.Documents = docItems
		}
	}

	// Optional: include audit trail.
	includeAudit := req.IncludeAudit != nil && *req.IncludeAudit
	if includeAudit {
		logs, _, auditErr := h.auditRepo.ListByRequestID(sctx.Ctx, req.RequestID, 0, 100)
		if auditErr == nil {
			trailItems := make([]resp.AuditTrailItem, len(logs))
			for i, l := range logs {
				trailItems[i] = resp.AuditTrailItem{
					ActionType:  l.ActionType,
					PerformedBy: l.PerformedByID,
					PerformedAt: resp.FormatTime(l.PerformedAt),
					OldValue:    l.OldValue,
					NewValue:    strPtr(l.NewValueJSON),
					Notes:       l.Notes,
				}
			}
			r.Data.AuditTrail = trailItems
		}
	}

	return r, nil
}

// ─── ST-002 ───────────────────────────────────────────────────────────────────

// GetTimeline returns the audit log history for a request in chronological order.
// FR-NFS-008: Audit trail. BR-NFS-016: audit_log is INSERT-only.
func (h *StatusHandler) GetTimeline(
	sctx *serverRoute.Context,
	req GetTimelineRequest,
) (*resp.TimelineResponse, error) {
	sr, err := h.srRepo.GetByID(sctx.Ctx, req.RequestID)
	if err != nil {
		if err == pgx.ErrNoRows {
			log.Error(sctx.Ctx, "GetTimeline: request not found: %s", req.RequestID)
			return nil, err
		}
		log.Error(sctx.Ctx, "GetTimeline: failed to fetch request %s: %v", req.RequestID, err)
		return nil, err
	}

	logs, _, err := h.auditRepo.ListByRequestID(sctx.Ctx, req.RequestID, 0, 500)
	if err != nil {
		log.Error(sctx.Ctx, "GetTimeline: failed to fetch audit logs for %s: %v", req.RequestID, err)
		return nil, err
	}

	trailItems := make([]resp.AuditTrailItem, len(logs))
	for i, l := range logs {
		trailItems[i] = resp.AuditTrailItem{
			ActionType:  l.ActionType,
			PerformedBy: l.PerformedByID,
			PerformedAt: resp.FormatTime(l.PerformedAt),
			OldValue:    l.OldValue,
			NewValue:    strPtr(l.NewValueJSON),
			Notes:       l.Notes,
		}
	}

	r := &resp.TimelineResponse{
		StatusCodeAndMessage: port.FetchSuccess,
		Data: resp.TimelineData{
			RequestID:    sr.RequestID,
			TicketNumber: sr.TicketNumber,
			Timeline:     trailItems,
		},
	}
	return r, nil
}

// ─── ST-003 ───────────────────────────────────────────────────────────────────

// GetReceipt returns the PDF acknowledgment receipt for a completed request.
// STUB: in production, fetches from DMS (Document Management Service).
// Response: Content-Type: application/pdf (binary stream).
// FR-NFS-004/005 acknowledgment receipt.
func (h *StatusHandler) GetReceipt(
	sctx *serverRoute.Context,
	req GetReceiptRequest,
) (*resp.ReceiptResponse, error) {
	sr, err := h.srRepo.GetByID(sctx.Ctx, req.RequestID)
	if err != nil {
		if err == pgx.ErrNoRows {
			log.Error(sctx.Ctx, "GetReceipt: request not found: %s", req.RequestID)
			return nil, err
		}
		log.Error(sctx.Ctx, "GetReceipt: failed to fetch request %s: %v", req.RequestID, err)
		return nil, err
	}

	// STUB: in production, generate/fetch PDF from DMS and stream bytes.
	// TODO: call Receipt Service (INT-NFS-008) for the actual PDF.
	log.Info(sctx.Ctx, "GetReceipt [STUB]: request %s ticket %s", sr.RequestID, sr.TicketNumber)

	r := &resp.ReceiptResponse{
		StatusCodeAndMessage: port.FetchSuccess,
		Data: resp.ReceiptData{
			RequestID:    sr.RequestID,
			TicketNumber: sr.TicketNumber,
			ReceiptURL:   fmt.Sprintf("/dms/receipts/nfs-%s.pdf", sr.RequestID),
			GeneratedAt:  resp.FormatTime(sr.UpdatedAt),
		},
	}
	return r, nil
}

// ─── ST-004 ───────────────────────────────────────────────────────────────────

// ListCustomerRequests returns a paginated list of NFS requests for a customer.
// BATCH: 2-query batch — total count + paginated rows (single round-trip).
// FR-NFS-006: Customer self-service visibility.
func (h *StatusHandler) ListCustomerRequests(
	sctx *serverRoute.Context,
	req ListCustomerRequestsRequest,
) (*resp.ListCustomerRequestsResponse, error) {
	page := 1
	if req.Page != nil && *req.Page > 0 {
		page = *req.Page
	}
	pageSize := 20
	if req.PageSize != nil && *req.PageSize > 0 && *req.PageSize <= 100 {
		pageSize = *req.PageSize
	}

	// BATCH: ListByCustomerID — 2-query batch: count + rows.
	requests, total, err := h.srRepo.ListByCustomerID(sctx.Ctx, req.CustomerID, page, pageSize)
	if err != nil {
		log.Error(sctx.Ctx, "ListCustomerRequests: failed for customer %d: %v", req.CustomerID, err)
		return nil, err
	}

	items := make([]resp.CustomerRequestSummary, len(requests))
	for i, sr := range requests {
		items[i] = resp.CustomerRequestSummary{
			RequestID:    sr.RequestID,
			TicketNumber: sr.TicketNumber,
			RequestType:  string(sr.RequestType),
			Status:       string(sr.Status),
			AuthMethod:   sr.AuthMethod,
			Channel:      sr.Channel,
			CreatedAt:    resp.FormatTime(sr.CreatedAt),
			UpdatedAt:    resp.FormatTime(sr.UpdatedAt),
		}
	}

	r := resp.NewListCustomerRequestsResponse(req.CustomerID, int(total), page, pageSize, items)
	return r, nil
}

// ─── ST-005 ───────────────────────────────────────────────────────────────────

// GetRequestDocuments returns the list of documents attached to a request.
// Includes verification status for each document (PENDING/VERIFIED/REJECTED).
func (h *StatusHandler) GetRequestDocuments(
	sctx *serverRoute.Context,
	req GetRequestDocumentsRequest,
) (*resp.RequestDocumentsResponse, error) {
	docs, _, err := h.docRepo.ListByRequestID(sctx.Ctx, req.RequestID, 0, 100)
	if err != nil {
		log.Error(sctx.Ctx, "GetRequestDocuments: failed for request %s: %v", req.RequestID, err)
		return nil, err
	}

	items := make([]resp.RequestDetailDocumentItem, len(docs))
	for i, d := range docs {
		items[i] = resp.RequestDetailDocumentItem{
			DocumentID:         d.DocumentID,
			DocumentType:       d.DocumentType,
			FileName:           d.FileName,
			FileSizeBytes:      d.FileSizeBytes,
			VerificationStatus: d.VerificationStatus,
			UploadedAt:         resp.FormatTime(d.UploadedAt),
		}
	}

	r := resp.NewRequestDocumentsResponse(req.RequestID, items)
	return r, nil
}
