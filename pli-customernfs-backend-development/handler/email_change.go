// Package handler provides HTTP handlers for the Customer NFS service.
// Phase: Phase 3 — Email Change
// Handler: EmailChangeHandler — covers email change initiation (WF-NFS-007)
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

// EmailChangeHandler handles HTTP endpoints for email change.
// WF-NFS-007: OTP-based email change.
type EmailChangeHandler struct {
	*serverHandler.Base

	srRepo *repo.ServiceRequestRepository
	tc     client.Client
	cfg    *config.Config
}

// NewEmailChangeHandler constructs the handler and wires dependencies via Uber FX.
func NewEmailChangeHandler(
	srRepo *repo.ServiceRequestRepository,
	tc client.Client,
	cfg *config.Config,
) *EmailChangeHandler {
	base := serverHandler.New("EmailChange").
		SetPrefix("/v1").
		AddPrefix("")
	return &EmailChangeHandler{
		Base:   base,
		srRepo: srRepo,
		tc:     tc,
		cfg:    cfg,
	}
}

// Routes registers all email-change HTTP routes.
func (h *EmailChangeHandler) Routes() []serverRoute.Route {
	return []serverRoute.Route{
		serverRoute.POST("/nfs/email-change/initiate", h.InitiateEmailChange).Name("Initiate Email Change"),
	}
}

// InitiateEmailChange handles POST /nfs/email-change/initiate.
// Creates a service request and starts WF-NFS-007.
func (h *EmailChangeHandler) InitiateEmailChange(
	sctx *serverRoute.Context,
	req InitiateEmailChangeRequest,
) (*resp.EmailChangeInitiateResponse, error) {
	// VR-NFS-015: Check for duplicate pending request.
	isDuplicate, existingTicket, err := h.srRepo.CheckDuplicateRequest(sctx.Ctx, req.CustomerID, "EMAIL_CHANGE")
	if err != nil {
		log.Error(sctx.Ctx, "InitiateEmailChange: duplicate check failed: %v", err)
		return nil, err
	}
	if isDuplicate {
		log.Error(sctx.Ctx, "InitiateEmailChange: duplicate pending request %s for customer %d", existingTicket, req.CustomerID)
		return nil, &apierrors.AppError{Code: 409, Message: fmt.Sprintf("pending request exists: %s. Complete or withdraw it first", existingTicket)}
	}

	// Validate office code for PostOffice channel
	if req.Channel == "PostOffice" && (req.OfficeCode == nil || *req.OfficeCode == "") {
		log.Error(sctx.Ctx, "InitiateEmailChange: office code is required for PostOffice channel")
		return nil, &apierrors.AppError{Code: 400, Message: "office_code is required when channel is PostOffice"}
	}

	// Generate ticket number.
	ticketNumber, err := h.srRepo.GenerateTicketNumber(sctx.Ctx, "EMAIL_CHANGE")
	if err != nil {
		log.Error(sctx.Ctx, "InitiateEmailChange: ticket number generation failed: %v", err)
		return nil, err
	}

	customerIDStr := strconv.FormatInt(req.CustomerID, 10)

	sr := domain.ServiceRequest{
		RequestID:    uuid.New().String(),
		CustomerID:   req.CustomerID,
		RequestType:  "EMAIL_CHANGE",
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
		Notes:         strPtr("Email change request created"),
	}

	if err := h.srRepo.Create(sctx.Ctx, &sr, &audit); err != nil {
		log.Error(sctx.Ctx, "InitiateEmailChange: failed to create service request: %v", err)
		return nil, err
	}

	wfInput := workflows.EmailChangeWorkflowInput{
		RequestID:    sr.RequestID,
		TicketNumber: sr.TicketNumber,
		CustomerID:   req.CustomerID,
		Channel:      req.Channel,
		InitiatedBy:  customerIDStr,
	}

	workflowID := fmt.Sprintf("nfs-email-%s", sr.RequestID)
	options := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: workflows.TaskQueue,
	}

	_, err = h.tc.ExecuteWorkflow(sctx.Ctx, options, workflows.EmailChangeWorkflow, wfInput)
	if err != nil {
		log.Error(sctx.Ctx, "InitiateEmailChange: failed to start EmailChangeWorkflow: %v", err)
		return nil, err
	}

	log.Info(sctx.Ctx, "InitiateEmailChange: started WF-NFS-007 workflowID=%s", workflowID)

	return resp.NewEmailChangeInitiateResponse(&sr), nil
}
