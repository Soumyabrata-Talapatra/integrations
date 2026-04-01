// Package handler provides HTTP handlers for the Customer NFS service.
// Phase: Phase 3 — Supporting Endpoints
// Handler: CPCHandler — covers CPC-001..006
//
// CPC-001: GET  /nfs/cpc/queue                             — work queue with SLA
// CPC-002: GET  /nfs/cpc/queue/stats                       — summary stats
// CPC-003: POST /nfs/requests/:request_id/assign           — assign to officer
// CPC-004: POST /nfs/requests/:request_id/reject           — CPC rejection
// CPC-005: POST /nfs/requests/:request_id/send-back        — alias for request-documents
// CPC-006: GET  /nfs/cpc/sla-dashboard                     — SLA performance
//
// FR-NFS-009: CPC operations support.
// BR-NFS-002: CPC assignment guidelines.
// BR-NFS-012: Status machine enforced by activities.
package handler

import (
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	config "gitlab.cept.gov.in/it-2.0-common/api-config"
	log "gitlab.cept.gov.in/it-2.0-common/n-api-log"
	serverHandler "gitlab.cept.gov.in/it-2.0-common/n-api-server/handler"
	serverRoute "gitlab.cept.gov.in/it-2.0-common/n-api-server/route"

	"customer-nfs-service/core/domain"
	resp "customer-nfs-service/handler/response"
	repo "customer-nfs-service/repo/postgres"
)

// CPCHandler handles CPC operations HTTP endpoints.
// FR-NFS-009: CPC operations. FR-NFS-010: SLA monitoring.
type CPCHandler struct {
	*serverHandler.Base

	srRepo    *repo.ServiceRequestRepository
	auditRepo *repo.AuditLogRepository
	cfg       *config.Config
}

// NewCPCHandler constructs the handler and wires dependencies via Uber FX.
func NewCPCHandler(
	srRepo *repo.ServiceRequestRepository,
	auditRepo *repo.AuditLogRepository,
	cfg *config.Config,
) *CPCHandler {
	base := serverHandler.New("CPC").
		SetPrefix("/v1").
		AddPrefix("")
	return &CPCHandler{
		Base:      base,
		srRepo:    srRepo,
		auditRepo: auditRepo,
		cfg:       cfg,
	}
}

// Routes registers all CPC operations HTTP routes.
func (h *CPCHandler) Routes() []serverRoute.Route {
	return []serverRoute.Route{
		serverRoute.GET("/nfs/cpc/queue", h.GetCPCQueue).Name("CPC Work Queue"),
		serverRoute.GET("/nfs/cpc/queue/stats", h.GetCPCQueueStats).Name("CPC Queue Stats"),
		serverRoute.POST("/nfs/requests/:request_id/assign", h.AssignRequest).Name("Assign Request to CPC Officer"),
		serverRoute.POST("/nfs/requests/:request_id/reject", h.RejectRequest).Name("CPC Reject Request"),
		serverRoute.POST("/nfs/requests/:request_id/send-back", h.SendBack).Name("CPC Send Back for Documents"),
		serverRoute.GET("/nfs/cpc/sla-dashboard", h.GetSLADashboard).Name("CPC SLA Dashboard"),
	}
}

// ─── CPC-001 ──────────────────────────────────────────────────────────────────

// GetCPCQueue returns the CPC work queue with SLA traffic-light indicators.
// GREEN: <50% SLA elapsed. AMBER: 50-80%. RED: >80% or breached.
// BATCH: 2-query batch — count + paginated rows.
// FR-NFS-009, FR-NFS-010.
func (h *CPCHandler) GetCPCQueue(
	sctx *serverRoute.Context,
	req GetCPCQueueRequest,
) (*resp.GetCPCQueueResponse, error) {
	page := 1
	if req.Page != nil && *req.Page > 0 {
		page = *req.Page
	}
	pageSize := 20
	if req.PageSize != nil && *req.PageSize > 0 && *req.PageSize <= 100 {
		pageSize = *req.PageSize
	}

	// BATCH: 2-query count + paginated rows.
	total, requests, err := h.srRepo.ListCPCQueue(sctx.Ctx, req.Status, req.RequestType, req.OfficeCode, req.AssignedToMe, page, pageSize)
	if err != nil {
		log.Error(sctx.Ctx, "GetCPCQueue: failed to query CPC queue: %v", err)
		return nil, err
	}

	items := make([]resp.CPCQueueItem, len(requests))
	for i, sr := range requests {
		slaStatus, daysLeft := computeSLAStatus(sr.SLADeadline)
		slaDL := ""
		if sr.SLADeadline != nil {
			slaDL = sr.SLADeadline.Format("2006-01-02")
		}
		items[i] = resp.CPCQueueItem{
			RequestID:    sr.RequestID,
			TicketNumber: sr.TicketNumber,
			CustomerID:   sr.CustomerID,
			RequestType:  string(sr.RequestType),
			Status:       string(sr.Status),
			AuthMethod:   sr.AuthMethod,
			Channel:      sr.Channel,
			OfficeCode:   *sr.OfficeCode,
			SLADeadline:  slaDL,
			SLAStatus:    slaStatus,
			SLADaysLeft:  daysLeft,
			CreatedAt:    resp.FormatTime(sr.CreatedAt),
			UpdatedAt:    resp.FormatTime(sr.UpdatedAt),
		}
	}

	r := resp.NewGetCPCQueueResponse(total, page, pageSize, items)
	return r, nil
}

// ─── CPC-002 ──────────────────────────────────────────────────────────────────

// GetCPCQueueStats returns summary counts for the CPC queue.
// BATCH: 5-query count batch: by_status, by_type, by_sla, by_channel, overdue.
func (h *CPCHandler) GetCPCQueueStats(
	sctx *serverRoute.Context,
	req GetCPCQueueStatsRequest,
) (*resp.GetCPCQueueStatsResponse, error) {
	stats, err := h.srRepo.GetCPCQueueStats(sctx.Ctx, req.OfficeCode)
	if err != nil {
		log.Error(sctx.Ctx, "GetCPCQueueStats: failed to query stats: %v", err)
		return nil, err
	}

	r := resp.NewGetCPCQueueStatsResponse(req.OfficeCode, resp.CPCQueueStats{
		TotalPending:    stats["total_pending"],
		TotalInProgress: stats["total_in_progress"],
		TotalOverdue:    stats["total_overdue"],
		SLAGreen:        stats["sla_green"],
		SLAAmber:        stats["sla_amber"],
		SLARed:          stats["sla_red"],
		AddressChanges:  stats["address_changes"],
		NameChanges:     stats["name_changes"],
	})
	return r, nil
}

// ─── CPC-003 ──────────────────────────────────────────────────────────────────

// AssignRequest assigns an NFS request to a CPC officer.
// BR-NFS-002: CPC assignment based on customer's servicing office.
// Status transition: PENDING_APPROVAL → IN_PROGRESS.
func (h *CPCHandler) AssignRequest(
	sctx *serverRoute.Context,
	req AssignRequestRequest,
) (*resp.AssignRequestResponse, error) {
	sr, err := h.srRepo.GetByID(sctx.Ctx, req.RequestID)
	if err != nil {
		if err == pgx.ErrNoRows {
			log.Error(sctx.Ctx, "AssignRequest: request not found: %s", req.RequestID)
			return nil, err
		}
		log.Error(sctx.Ctx, "AssignRequest: failed to fetch request %s: %v", req.RequestID, err)
		return nil, err
	}

	if sr.Status != "PENDING_APPROVAL" {
		return nil, fmt.Errorf("request %s must be in PENDING_APPROVAL status to assign (current: %s)", req.RequestID, sr.Status)
	}

	// BATCH: AssignToCPC — 3-op TX batch: UPDATE service_request + INSERT cpc_queue + INSERT audit.
	audit := domain.AuditLog{
		RequestID: req.RequestID,
		// ActionType:    "IN_PROGRESS",
		ActionType:    "STATUS_CHANGE", // was "IN_PROGRESS"
		NewValueJSON:  fmt.Sprintf(`{"assigned_to":"%s"}`, req.AssignedTo),
		PerformedByID: req.AssignedTo,
		Notes:         req.Reason,
	}
	_, err = h.srRepo.UpdateStatus(sctx.Ctx, req.RequestID, "IN_PROGRESS", &req.AssignedTo, nil, nil, &audit)
	if err != nil {
		log.Error(sctx.Ctx, "AssignRequest: failed to update status for %s: %v", req.RequestID, err)
		return nil, err
	}

	log.Info(sctx.Ctx, "AssignRequest: request %s assigned to %s", req.RequestID, req.AssignedTo)
	r := resp.NewAssignRequestResponse(req.RequestID, req.AssignedTo, "IN_PROGRESS")
	return r, nil
}

// ─── CPC-004 ──────────────────────────────────────────────────────────────────

// RejectRequest records a CPC rejection and transitions to REJECTED status.
// BR-NFS-012: IN_PROGRESS → REJECTED.
// ERR-NFS-SR-005: rejection_reason is mandatory.
// BR-NFS-016: audit_log is INSERT-only.
func (h *CPCHandler) RejectRequest(
	sctx *serverRoute.Context,
	req CPCRejectRequest,
) (*resp.CPCRejectResponse, error) {
	sr, err := h.srRepo.GetByID(sctx.Ctx, req.RequestID)
	if err != nil {
		if err == pgx.ErrNoRows {
			log.Error(sctx.Ctx, "RejectRequest: request not found: %s", req.RequestID)
			return nil, err
		}
		log.Error(sctx.Ctx, "RejectRequest: failed to fetch request %s: %v", req.RequestID, err)
		return nil, err
	}

	// BR-NFS-012: only IN_PROGRESS requests can be rejected via CPC path.
	if sr.Status != "IN_PROGRESS" {
		return nil, fmt.Errorf("request %s must be IN_PROGRESS to reject (current: %s)", req.RequestID, sr.Status)
	}

	reasonNote := req.RejectionReason
	audit := domain.AuditLog{
		RequestID:     req.RequestID,
		ActionType:    "REJECTED",
		NewValueJSON:  fmt.Sprintf(`{"rejection_reason":"%s","rejected_by":"%s"}`, req.RejectionReason, req.RejectedBy),
		PerformedByID: req.RejectedBy,
		Notes:         &reasonNote,
	}

	// BATCH: UpdateStatus — TX batch: UPDATE service_request + INSERT status_transition + INSERT audit.
	_, err = h.srRepo.UpdateStatus(sctx.Ctx, req.RequestID, "REJECTED", &req.RejectedBy, &req.RejectionReason, nil, &audit)
	if err != nil {
		log.Error(sctx.Ctx, "RejectRequest: failed to update status for %s: %v", req.RequestID, err)
		return nil, err
	}

	log.Info(sctx.Ctx, "RejectRequest: request %s rejected by %s", req.RequestID, req.RejectedBy)
	r := resp.NewCPCRejectResponse(req.RequestID, req.RejectedBy, req.RejectionReason)
	return r, nil
}

// ─── CPC-005 ──────────────────────────────────────────────────────────────────

// SendBack is an alias for RequestMissingDocuments (CORE-010) via the CPC interface.
// BR-NFS-014: same logic — 7-day expiry, max 3 attempts.
// Status: IN_PROGRESS → PENDING_DOCUMENTS.
func (h *CPCHandler) SendBack(
	sctx *serverRoute.Context,
	req CPCSendBackRequest,
) (*resp.CPCSendBackResponse, error) {
	sr, err := h.srRepo.GetByID(sctx.Ctx, req.RequestID)
	if err != nil {
		if err == pgx.ErrNoRows {
			log.Error(sctx.Ctx, "SendBack: request not found: %s", req.RequestID)
			return nil, err
		}
		log.Error(sctx.Ctx, "SendBack: failed to fetch request %s: %v", req.RequestID, err)
		return nil, err
	}

	// BR-NFS-012: only IN_PROGRESS requests can be sent back.
	if sr.Status != "IN_PROGRESS" {
		return nil, fmt.Errorf("request %s must be IN_PROGRESS to send back (current: %s)", req.RequestID, sr.Status)
	}

	// BR-NFS-014: check attempt count.
	attemptCount, err := h.srRepo.GetDocumentRequestAttemptCount(sctx.Ctx, req.RequestID)
	if err != nil {
		log.Error(sctx.Ctx, "SendBack: failed to get attempt count for %s: %v", req.RequestID, err)
		return nil, err
	}
	if attemptCount >= 3 {
		return nil, fmt.Errorf("maximum document request attempts reached (BR-NFS-014)")
	}

	expiresAt := time.Now().AddDate(0, 0, 7)
	token, err := h.srRepo.CreateSecureUploadLink(sctx.Ctx, req.RequestID, req.MissingDocuments, req.RequestedBy, expiresAt)
	if err != nil {
		log.Error(sctx.Ctx, "SendBack: failed to create secure upload link for %s: %v", req.RequestID, err)
		return nil, err
	}

	audit := domain.AuditLog{
		RequestID: req.RequestID,
		// ActionType:    "PENDING_DOCUMENTS",
		ActionType:    "MISSING_DOC_REQUESTED", // was "PENDING_DOCUMENTS"
		NewValueJSON:  fmt.Sprintf(`{"missing_docs":%d,"requested_by":"%s"}`, len(req.MissingDocuments), req.RequestedBy),
		PerformedByID: req.RequestedBy,
		Notes:         strPtr("CPC sent back for missing documents"),
	}
	_, err = h.srRepo.UpdateStatus(sctx.Ctx, req.RequestID, "PENDING_DOCUMENTS", &req.RequestedBy, nil, nil, &audit)
	if err != nil {
		log.Error(sctx.Ctx, "SendBack: failed to update status for %s: %v", req.RequestID, err)
		return nil, err
	}

	baseURL := h.cfg.GetString("nfs.secureupload.baseurl")
	secureURL := fmt.Sprintf("%s/%s", baseURL, token)
	log.Info(sctx.Ctx, "SendBack: secure link created for request %s attempt %d", req.RequestID, attemptCount+1)

	r := resp.NewCPCSendBackResponse(req.RequestID, secureURL, expiresAt.Format(time.RFC3339), attemptCount+1, req.MissingDocuments)
	return r, nil
}

// ─── CPC-006 ──────────────────────────────────────────────────────────────────

// GetSLADashboard returns aggregated SLA performance metrics for a CPC office.
// Used by supervisors to identify bottlenecks and escalation needs.
// FR-NFS-010: SLA monitoring.
func (h *CPCHandler) GetSLADashboard(
	sctx *serverRoute.Context,
	req GetSLADashboardRequest,
) (*resp.GetSLADashboardResponse, error) {
	stats, err := h.srRepo.GetCPCQueueStats(sctx.Ctx, req.OfficeCode)
	if err != nil {
		log.Error(sctx.Ctx, "GetSLADashboard: failed to query stats: %v", err)
		return nil, err
	}

	// Fetch breached items.
	breachedSRs, err := h.srRepo.ListSLABreached(sctx.Ctx, req.OfficeCode, req.SLAStatus)
	if err != nil {
		log.Error(sctx.Ctx, "GetSLADashboard: failed to query breached: %v", err)
		return nil, err
	}

	breachedItems := make([]resp.SLABreachItem, len(breachedSRs))
	for i, sr := range breachedSRs {
		daysOverdue := 0
		if sr.SLADeadline != nil {
			daysOverdue = int(time.Since(*sr.SLADeadline).Hours() / 24)
		}
		breachedItems[i] = resp.SLABreachItem{
			RequestID:    sr.RequestID,
			TicketNumber: sr.TicketNumber,
			RequestType:  string(sr.RequestType),
			OfficeCode:   *sr.OfficeCode,
			DaysOverdue:  daysOverdue,
		}
	}

	summary := resp.CPCQueueStats{
		TotalPending:    stats["total_pending"],
		TotalInProgress: stats["total_in_progress"],
		TotalOverdue:    stats["total_overdue"],
		SLAGreen:        stats["sla_green"],
		SLAAmber:        stats["sla_amber"],
		SLARed:          stats["sla_red"],
		AddressChanges:  stats["address_changes"],
		NameChanges:     stats["name_changes"],
	}

	r := resp.NewGetSLADashboardResponse(req.OfficeCode, summary, breachedItems)
	return r, nil
}

// ─── helpers ─────────────────────────────────────────────────────────────────

// computeSLAStatus returns the traffic-light SLA status and remaining days.
// GREEN: <50% SLA elapsed, AMBER: 50-80%, RED: >80% or breached.
func computeSLAStatus(deadline *time.Time) (string, int) {
	if deadline == nil {
		return "GREEN", 99
	}
	now := time.Now()
	if now.After(*deadline) {
		daysOverdue := int(now.Sub(*deadline).Hours() / 24)
		return "RED", -daysOverdue
	}
	daysLeft := int(deadline.Sub(now).Hours() / 24)
	if daysLeft == 0 {
		return "RED", 0
	}
	// Determine elapsed percentage using a standard 15-day SLA window.
	slaWindow := 15.0
	elapsed := slaWindow - float64(daysLeft)
	pct := elapsed / slaWindow
	switch {
	case pct < 0.5:
		return "GREEN", daysLeft
	case pct < 0.8:
		return "AMBER", daysLeft
	default:
		return "RED", daysLeft
	}
}
