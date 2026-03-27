// Package handler provides HTTP handlers for the Customer NFS service.
// Phase: Phase 3 — Mobile Change
// Handler: MobileChangeHandler — covers mobile change initiation (WF-NFS-006)
package handler

import (
	"fmt"
	"strconv"

	"github.com/google/uuid"
	"go.temporal.io/sdk/client"

	config "gitlab.cept.gov.in/it-2.0-common/api-config"
	apierrors "gitlab.cept.gov.in/it-2.0-common/n-api-errors"
	log "gitlab.cept.gov.in/it-2.0-common/n-api-log"
	serverHandler "gitlab.cept.gov.in/it-2.0-common/n-api-server/handler"
	serverRoute "gitlab.cept.gov.in/it-2.0-common/n-api-server/route"

	"customer-nfs-service/core/domain"
	resp "customer-nfs-service/handler/response"
	repo "customer-nfs-service/repo/postgres"
	"customer-nfs-service/temporal/workflows"
)

// MobileChangeHandler handles HTTP endpoints for mobile change.
// WF-NFS-006: OTP-based mobile number change.
type MobileChangeHandler struct {
	*serverHandler.Base

	srRepo *repo.ServiceRequestRepository
	tc     client.Client
	cfg    *config.Config
}

// NewMobileChangeHandler constructs the handler and wires dependencies via Uber FX.
func NewMobileChangeHandler(
	srRepo *repo.ServiceRequestRepository,
	tc client.Client,
	cfg *config.Config,
) *MobileChangeHandler {
	base := serverHandler.New("MobileChange").
		SetPrefix("/v1").
		AddPrefix("")
	return &MobileChangeHandler{
		Base:   base,
		srRepo: srRepo,
		tc:     tc,
		cfg:    cfg,
	}
}

// Routes registers all mobile-change HTTP routes.
func (h *MobileChangeHandler) Routes() []serverRoute.Route {
	return []serverRoute.Route{
		serverRoute.POST("/nfs/mobile-change/initiate", h.InitiateMobileChange).Name("Initiate Mobile Change"),
	}
}

// InitiateMobileChange handles POST /nfs/mobile-change/initiate.
// Creates a service request and starts WF-NFS-006.
func (h *MobileChangeHandler) InitiateMobileChange(
	sctx *serverRoute.Context,
	req InitiateMobileChangeRequest,
) (*resp.MobileChangeInitiateResponse, error) {
	// VR-NFS-015: Check for duplicate pending request.
	isDuplicate, existingTicket, err := h.srRepo.CheckDuplicateRequest(sctx.Ctx, req.CustomerID, "MOBILE_CHANGE")
	if err != nil {
		log.Error(sctx.Ctx, "InitiateMobileChange: duplicate check failed: %v", err)
		return nil, err
	}
	if isDuplicate {
		log.Error(sctx.Ctx, "InitiateMobileChange: duplicate pending request %s for customer %d", existingTicket, req.CustomerID)
		return nil, &apierrors.AppError{Code: 409, Message: fmt.Sprintf("pending request exists: %s. Complete or withdraw it first", existingTicket)}
	}

	// Validate office code for PostOffice channel
	if req.Channel == "PostOffice" && (req.OfficeCode == nil || *req.OfficeCode == "") {
		log.Error(sctx.Ctx, "InitiateMobileChange: office code is required for PostOffice channel")
		return nil, &apierrors.AppError{Code: 400, Message: "office_code is required when channel is PostOffice"}
	}

	// Generate ticket number.
	ticketNumber, err := h.srRepo.GenerateTicketNumber(sctx.Ctx, "MOBILE_CHANGE")
	if err != nil {
		log.Error(sctx.Ctx, "InitiateMobileChange: ticket number generation failed: %v", err)
		return nil, err
	}

	customerIDStr := strconv.FormatInt(req.CustomerID, 10)

	sr := domain.ServiceRequest{
		RequestID:    uuid.New().String(),
		CustomerID:   req.CustomerID,
		RequestType:  "MOBILE_CHANGE",
		TicketNumber: ticketNumber,
		AuthMethod:   "OTP",
		Status:       "CREATED",
		Channel:      req.Channel,
		OfficeCode:   req.OfficeCode,
		InitiatedBy:  req.CustomerID,
		CreatedBy:    customerIDStr,
	}

	audit := domain.AuditLog{
		AuditID:       uuid.New().String(),
		RequestID:     sr.RequestID,
		ActionType:    "CREATED",
		NewValueJSON:  fmt.Sprintf(`{"ticket_number":"%s","channel":"%s"}`, ticketNumber, req.Channel),
		PerformedByID: customerIDStr,
		Notes:         strPtr("Mobile change request created"),
	}

	if err := h.srRepo.Create(sctx.Ctx, &sr, &audit); err != nil {
		log.Error(sctx.Ctx, "InitiateMobileChange: failed to create service request: %v", err)
		return nil, err
	}

	wfInput := workflows.MobileChangeWorkflowInput{
		RequestID:    sr.RequestID,
		TicketNumber: sr.TicketNumber,
		CustomerID:   req.CustomerID,
		Channel:      req.Channel,
		InitiatedBy:  customerIDStr,
	}

	workflowID := fmt.Sprintf("nfs-mobile-%s", sr.RequestID)
	options := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: workflows.TaskQueue,
	}

	_, err = h.tc.ExecuteWorkflow(sctx.Ctx, options, workflows.MobileChangeWorkflow, wfInput)
	if err != nil {
		log.Error(sctx.Ctx, "InitiateMobileChange: failed to start MobileChangeWorkflow: %v", err)
		return nil, err
	}

	log.Info(sctx.Ctx, "InitiateMobileChange: started WF-NFS-006 workflowID=%s", workflowID)

	return resp.NewMobileChangeInitiateResponse(&sr), nil
}
