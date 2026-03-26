// Package handler provides HTTP handlers for the Customer NFS service.
// Phase: Phase 2 — Name Change
// Handler: NameChangeHandler — covers CORE-005..010
//
// WF-NFS-003: Aadhaar name change (CORE-005 with auth_method=AADHAAR + CORE-006)
// WF-NFS-004: Manual name change (CORE-005 with auth_method=MANUAL + CORE-007 + CORE-008)
// WF-NFS-005: Withdrawal (CORE-009 — both address and name change requests)
// Sub-workflow: Missing documents (CORE-010 — BR-NFS-014)
package handler

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"go.temporal.io/sdk/client"

	config "gitlab.cept.gov.in/it-2.0-common/api-config"
	apierrors "gitlab.cept.gov.in/it-2.0-common/n-api-errors"
	log "gitlab.cept.gov.in/it-2.0-common/n-api-log"
	serverHandler "gitlab.cept.gov.in/it-2.0-common/n-api-server/handler"
	serverRoute "gitlab.cept.gov.in/it-2.0-common/n-api-server/route"

	"customer-nfs-service/core/domain"
	resp "customer-nfs-service/handler/response"
	repo "customer-nfs-service/repo/postgres"
	"customer-nfs-service/temporal/activities"
	"customer-nfs-service/temporal/workflows"
)

// NameChangeHandler handles HTTP endpoints for the name change core journey.
type NameChangeHandler struct {
	*serverHandler.Base

	srRepo   *repo.ServiceRequestRepository
	nameRepo *repo.NameChangeRepository
	docRepo  *repo.DocumentUploadRepository
	tc       client.Client
	cfg      *config.Config
}

// NewNameChangeHandler constructs the handler and wires dependencies via Uber FX.
func NewNameChangeHandler(
	srRepo *repo.ServiceRequestRepository,
	nameRepo *repo.NameChangeRepository,
	docRepo *repo.DocumentUploadRepository,
	tc client.Client,
	cfg *config.Config,
) *NameChangeHandler {
	base := serverHandler.New("NameChange").
		SetPrefix("/v1").
		AddPrefix("")
	return &NameChangeHandler{
		Base:     base,
		srRepo:   srRepo,
		nameRepo: nameRepo,
		docRepo:  docRepo,
		tc:       tc,
		cfg:      cfg,
	}
}

// Routes registers all name-change HTTP routes.
func (h *NameChangeHandler) Routes() []serverRoute.Route {
	return []serverRoute.Route{
		serverRoute.POST("/nfs/name-change/initiate", h.InitiateNameChange).Name("Initiate Name Change"),
		serverRoute.POST("/nfs/name-change/:request_id/verify-otp", h.VerifyNameOTP).Name("Verify Name Change OTP"),
		serverRoute.POST("/nfs/name-change/:request_id/submit", h.SubmitNameChange).Name("Submit Manual Name Change"),
		serverRoute.POST("/nfs/name-change/:request_id/approve", h.ApproveNameChange).Name("CPC Approve Name Change"),
		serverRoute.POST("/nfs/requests/:request_id/withdraw", h.WithdrawRequest).Name("Withdraw NFS Request"),
		serverRoute.POST("/nfs/requests/:request_id/request-documents", h.RequestMissingDocuments).Name("Request Missing Documents"),
	}
}

// ─── CORE-005 ─────────────────────────────────────────────────────────────────

func (h *NameChangeHandler) InitiateNameChange(
	sctx *serverRoute.Context,
	req InitiateNameChangeRequest,
) (*resp.NameChangeInitiateResponse, error) {
	// BR-NFS-015: AADHAAR only allowed from Portal and Mobile channels.
	if req.AuthMethod == "AADHAAR" && req.Channel != "Portal" && req.Channel != "Mobile" {
		log.Error(sctx.Ctx, "InitiateNameChange: AADHAAR not allowed from channel: %s", req.Channel)
		return nil, &apierrors.AppError{Code: 422, Message: fmt.Sprintf("Aadhaar auto-approval not available from %s. Use MANUAL method", req.Channel)}
	}

	if req.AuthMethod == "MANUAL" && req.NewName == nil {
		log.Error(sctx.Ctx, "InitiateNameChange: new_name required for MANUAL auth_method")
		return nil, &apierrors.AppError{Code: 422, Message: "new_name is required for MANUAL name change"}
	}

	// VR-NFS-015: Duplicate pending request check.
	isDuplicate, existingTicket, err := h.srRepo.CheckDuplicateRequest(sctx.Ctx, req.CustomerID, "NAME_CHANGE")
	if err != nil {
		log.Error(sctx.Ctx, "InitiateNameChange: duplicate check failed: %v", err)
		return nil, err
	}
	if isDuplicate {
		log.Error(sctx.Ctx, "InitiateNameChange: duplicate pending request %s for customer %s", existingTicket, req.CustomerID)
		return nil, &apierrors.AppError{Code: 409, Message: fmt.Sprintf("pending name change request exists: %s. Complete or withdraw it first", existingTicket)}
	}

	// BR-NFS-011: Generate ticket number.
	ticketNumber, err := h.srRepo.GenerateTicketNumber(sctx.Ctx, "NAME_CHANGE")
	if err != nil {
		log.Error(sctx.Ctx, "InitiateNameChange: ticket number generation failed: %v", err)
		return nil, err
	}

	requestID := newUUID()

	sr, nameDetail := req.ToDomain(ticketNumber, domain.SLADeadline{})
	sr.RequestID = requestID
	sr.InitiatedBy = req.CustomerID
	sr.CreatedBy = req.CustomerID
	nameDetail.RequestID = requestID
	nameDetail.DetailID = uuid.New().String()
	nameDetail.CreatedBy = req.CustomerID

	audit := domain.AuditLog{
		AuditID:       uuid.New().String(),
		RequestID:     sr.RequestID,
		ActionType:    "CREATED",
		NewValueJSON:  fmt.Sprintf(`{"ticket_number":"%s","auth_method":"%s","channel":"%s"}`, ticketNumber, req.AuthMethod, req.Channel),
		PerformedByID: req.CustomerID,
		IPAddress:     nil,
		Notes:         strPtr("Name change request initiated"),
	}

	err = h.srRepo.CreateWithNameDetail(sctx.Ctx, sr, nameDetail, &audit)
	if err != nil {
		log.Error(sctx.Ctx, "InitiateNameChange: failed to create service request: %v", err)
		return nil, err
	}

	slaDays := h.cfg.GetInt("nfs.sladeadlineregional")
	if slaDays == 0 {
		slaDays = 15
	}
	wfInput := workflows.NameChangeWorkflowInput{
		RequestID:    sr.RequestID,
		TicketNumber: sr.TicketNumber,
		CustomerID:   req.CustomerID,
		PolicyNumber: req.PolicyNumber,
		AuthMethod:   req.AuthMethod,
		Channel:      req.Channel,
		OfficeCode:   strDeref(req.OfficeCode),
		InitiatedBy:  req.CustomerID,
		SLADays:      slaDays,
	}

	workflowID := fmt.Sprintf("nfs-name-%s-%s", req.AuthMethod, sr.RequestID)
	options := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: workflows.TaskQueue,
	}

	var otpSent bool
	var slaDeadline *time.Time

	if req.AuthMethod == "AADHAAR" {
		_, err = h.tc.ExecuteWorkflow(sctx.Ctx, options, workflows.AadhaarNameChangeWorkflow, wfInput)
		if err != nil {
			log.Error(sctx.Ctx, "InitiateNameChange: failed to start AadhaarNameChangeWorkflow: %v", err)
			return nil, err
		}
		otpSent = true
		log.Info(sctx.Ctx, "InitiateNameChange: started WF-NFS-003 workflowID=%s", workflowID)
	} else {
		_, err = h.tc.ExecuteWorkflow(sctx.Ctx, options, workflows.ManualNameChangeWorkflow, wfInput)
		if err != nil {
			log.Error(sctx.Ctx, "InitiateNameChange: failed to start ManualNameChangeWorkflow: %v", err)
			return nil, err
		}
		dl := time.Now().AddDate(0, 0, slaDays)
		slaDeadline = &dl
		sr.SLADeadline = slaDeadline
		if err := h.srRepo.UpdateSLADeadline(sctx.Ctx, sr.RequestID, dl); err != nil {
			log.Error(sctx.Ctx, "InitiateNameChange: failed to update SLA deadline for %s: %v", sr.RequestID, err)
			return nil, err
		}
		log.Info(sctx.Ctx, "InitiateNameChange: started WF-NFS-004 workflowID=%s slaDeadline=%s", workflowID, dl.Format("2006-01-02"))
	}

	r := resp.NewNameChangeInitiateResponse(sr.RequestID, sr.TicketNumber, req.AuthMethod, "CREATED", otpSent, slaDeadline)
	return r, nil
}

// ─── CORE-006 ─────────────────────────────────────────────────────────────────

func (h *NameChangeHandler) VerifyNameOTP(
	sctx *serverRoute.Context,
	req VerifyNameOTPRequest,
) (*resp.NameVerifyOTPResponse, error) {
	sr, err := h.srRepo.GetByID(sctx.Ctx, req.RequestID)
	if err != nil {
		if err == pgx.ErrNoRows {
			log.Error(sctx.Ctx, "VerifyNameOTP: request not found: %s", req.RequestID)
			return nil, &apierrors.AppError{Code: 404, Message: fmt.Sprintf("request not found: %s", req.RequestID)}
		}
		log.Error(sctx.Ctx, "VerifyNameOTP: failed to fetch request %s: %v", req.RequestID, err)
		return nil, err
	}
	if sr.Status != "CREATED" {
		return nil, &apierrors.AppError{Code: 422, Message: fmt.Sprintf("OTP verification not allowed for request in status %s", sr.Status)}
	}

	wfID, wfRunID, err := h.srRepo.GetWorkflowState(sctx.Ctx, req.RequestID)
	if err != nil {
		log.Error(sctx.Ctx, "VerifyNameOTP: failed to get workflow state for %s: %v", req.RequestID, err)
		return nil, err
	}

	otpPayload := workflows.OTPSignalPayload{
		OTP:            req.OTPCode,
		OTPReferenceID: req.TxnID,
		CustomerID:     sr.CustomerID,
	}
	err = h.tc.SignalWorkflow(sctx.Ctx, wfID, wfRunID, workflows.SignalOTPSubmitted, otpPayload)
	if err != nil {
		log.Error(sctx.Ctx, "VerifyNameOTP: failed to signal workflow %s: %v", wfID, err)
		return nil, err
	}

	log.Info(sctx.Ctx, "VerifyNameOTP: signal sent to WF-NFS-003 %s for request %s", wfID, req.RequestID)
	r := resp.NewNameVerifyOTPResponse(sr.RequestID, "IN_PROGRESS")
	return r, nil
}

// ─── CORE-007 ─────────────────────────────────────────────────────────────────

func (h *NameChangeHandler) SubmitNameChange(
	sctx *serverRoute.Context,
	req SubmitNameChangeRequest,
) (*resp.NameSubmitResponse, error) {
	if err := req.Validate(); err != nil {
		log.Error(sctx.Ctx, "SubmitNameChange: validation failed: %v", err)
		return nil, err
	}

	sr, err := h.srRepo.GetByID(sctx.Ctx, req.RequestID)
	if err != nil {
		if err == pgx.ErrNoRows {
			log.Error(sctx.Ctx, "SubmitNameChange: request not found: %s", req.RequestID)
			return nil, &apierrors.AppError{Code: 404, Message: fmt.Sprintf("request not found: %s", req.RequestID)}
		}
		log.Error(sctx.Ctx, "SubmitNameChange: failed to fetch request %s: %v", req.RequestID, err)
		return nil, err
	}

	if sr.Status != "CREATED" {
		log.Error(sctx.Ctx, "SubmitNameChange: invalid status %s for request %s", sr.Status, req.RequestID)
		return nil, &apierrors.AppError{Code: 422, Message: fmt.Sprintf("request %s is in status %s; only CREATED requests can be submitted", req.RequestID, sr.Status)}
	}

	wfID, wfRunID, err := h.srRepo.GetWorkflowState(sctx.Ctx, req.RequestID)
	if err != nil {
		log.Error(sctx.Ctx, "SubmitNameChange: failed to get workflow state for %s: %v", req.RequestID, err)
		return nil, err
	}

	submitPayload := activities.DocumentsSubmittedPayload{
		RequestID:         req.RequestID,
		UploadedDocuments: req.UploadedDocuments,
		SubmittedBy:       req.SubmittedBy,
	}
	err = h.tc.SignalWorkflow(sctx.Ctx, wfID, wfRunID, workflows.SignalDocumentsSubmitted, submitPayload)
	if err != nil {
		log.Error(sctx.Ctx, "SubmitNameChange: failed to signal workflow %s: %v", wfID, err)
		return nil, err
	}

	log.Info(sctx.Ctx, "SubmitNameChange: request %s → PENDING_APPROVAL, WF-NFS-004 %s signaled", req.RequestID, wfID)
	r := resp.NewNameSubmitResponse(req.RequestID, "PENDING_APPROVAL", len(req.UploadedDocuments))
	return r, nil
}

// ─── CORE-008 ─────────────────────────────────────────────────────────────────

func (h *NameChangeHandler) ApproveNameChange(
	sctx *serverRoute.Context,
	req ApproveNameChangeRequest,
) (*resp.NameApproveResponse, error) {
	if req.Decision == "REJECT" && (req.RejectionReason == nil || *req.RejectionReason == "") {
		log.Error(sctx.Ctx, "ApproveNameChange: rejection reason missing for REJECT decision on %s", req.RequestID)
		return nil, &apierrors.AppError{Code: 422, Message: "rejection_reason is required when decision is REJECT"}
	}
	if req.Decision == "SEND_BACK" && len(req.MissingDocuments) == 0 {
		return nil, &apierrors.AppError{Code: 422, Message: "missing_documents is required when decision is SEND_BACK"}
	}

	sr, err := h.srRepo.GetByID(sctx.Ctx, req.RequestID)
	if err != nil {
		if err == pgx.ErrNoRows {
			log.Error(sctx.Ctx, "ApproveNameChange: request not found: %s", req.RequestID)
			return nil, &apierrors.AppError{Code: 404, Message: fmt.Sprintf("request not found: %s", req.RequestID)}
		}
		log.Error(sctx.Ctx, "ApproveNameChange: failed to fetch request %s: %v", req.RequestID, err)
		return nil, err
	}
	if sr.Status != "IN_PROGRESS" {
		log.Error(sctx.Ctx, "ApproveNameChange: invalid status %s for request %s", sr.Status, req.RequestID)
		return nil, &apierrors.AppError{Code: 422, Message: fmt.Sprintf("request %s is in status %s; only IN_PROGRESS requests can be approved", req.RequestID, sr.Status)}
	}

	wfID, wfRunID, err := h.srRepo.GetWorkflowState(sctx.Ctx, req.RequestID)
	if err != nil {
		log.Error(sctx.Ctx, "ApproveNameChange: failed to get workflow state for %s: %v", req.RequestID, err)
		return nil, err
	}

	newStatus := decisionToStatus(req.Decision)

	approvalPayload := workflows.ApprovalDecisionPayload{
		Decision:         req.Decision,
		ApprovedBy:       req.ApprovedBy,
		Reason:           strDeref(req.Remarks),
		RejectionReason:  req.RejectionReason,
		MissingDocuments: req.MissingDocuments,
	}
	err = h.tc.SignalWorkflow(sctx.Ctx, wfID, wfRunID, workflows.SignalApprovalDecision, approvalPayload)
	if err != nil {
		log.Error(sctx.Ctx, "ApproveNameChange: failed to signal workflow %s: %v", wfID, err)
		return nil, err
	}

	log.Info(sctx.Ctx, "ApproveNameChange: decision %s for request %s, WF-NFS-004 %s signaled", req.Decision, req.RequestID, wfID)

	r := resp.NewNameApproveResponse(req.RequestID, req.Decision, newStatus, 0)
	return r, nil
}

// ─── CORE-009 ─────────────────────────────────────────────────────────────────

func (h *NameChangeHandler) WithdrawRequest(
	sctx *serverRoute.Context,
	req WithdrawRequest,
) (*resp.WithdrawResponse, error) {
	sr, err := h.srRepo.GetByID(sctx.Ctx, req.RequestID)
	if err != nil {
		if err == pgx.ErrNoRows {
			log.Error(sctx.Ctx, "WithdrawRequest: request not found: %s", req.RequestID)
			return nil, &apierrors.AppError{Code: 404, Message: fmt.Sprintf("request not found: %s", req.RequestID)}
		}
		log.Error(sctx.Ctx, "WithdrawRequest: failed to fetch request %s: %v", req.RequestID, err)
		return nil, err
	}

	// BR-NFS-013: Validate withdrawal eligibility by current status.
	eligibleStatuses := map[string]bool{
		"CREATED":           true,
		"PENDING_DOCUMENTS": true,
		"PENDING_APPROVAL":  true,
	}
	if !eligibleStatuses[string(sr.Status)] {
		log.Error(sctx.Ctx, "WithdrawRequest: ineligible status %s for request %s", sr.Status, req.RequestID)
		return nil, &apierrors.AppError{Code: 422, Message: fmt.Sprintf("withdrawal not allowed for request in status %s (BR-NFS-013)", sr.Status)}
	}

	partialProcessing := sr.PartialProcessingFlag

	wfInput := workflows.WithdrawalWorkflowInput{
		RequestID:             req.RequestID,
		TicketNumber:          sr.TicketNumber,
		CustomerID:            sr.CustomerID,
		WithdrawalReason:      req.WithdrawalReason,
		RequestedBy:           req.RequestedBy,
		PartialProcessingFlag: partialProcessing,
	}

	workflowID := fmt.Sprintf("nfs-withdrawal-%s", req.RequestID)
	options := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: workflows.TaskQueue,
	}

	_, err = h.tc.ExecuteWorkflow(sctx.Ctx, options, workflows.WithdrawalWorkflow, wfInput)
	if err != nil {
		log.Error(sctx.Ctx, "WithdrawRequest: failed to start WithdrawalWorkflow: %v", err)
		return nil, err
	}

	log.Info(sctx.Ctx, "WithdrawRequest: started WF-NFS-005 workflowID=%s for request %s", workflowID, req.RequestID)

	newStatus := "WITHDRAWN"
	if partialProcessing {
		newStatus = "PENDING_WITHDRAWAL_APPROVAL"
	}

	r := resp.NewWithdrawResponse(req.RequestID, newStatus, workflowID)
	return r, nil
}

// ─── CORE-010 ─────────────────────────────────────────────────────────────────

func (h *NameChangeHandler) RequestMissingDocuments(
	sctx *serverRoute.Context,
	req RequestMissingDocumentsRequest,
) (*resp.RequestMissingDocumentsResponse, error) {
	sr, err := h.srRepo.GetByID(sctx.Ctx, req.RequestID)
	if err != nil {
		if err == pgx.ErrNoRows {
			log.Error(sctx.Ctx, "RequestMissingDocuments: request not found: %s", req.RequestID)
			return nil, &apierrors.AppError{Code: 404, Message: fmt.Sprintf("request not found: %s", req.RequestID)}
		}
		log.Error(sctx.Ctx, "RequestMissingDocuments: failed to fetch request %s: %v", req.RequestID, err)
		return nil, err
	}

	attemptCount, err := h.srRepo.GetDocumentRequestAttemptCount(sctx.Ctx, req.RequestID)
	if err != nil {
		log.Error(sctx.Ctx, "RequestMissingDocuments: failed to get attempt count for %s: %v", req.RequestID, err)
		return nil, err
	}

	// BR-NFS-014: Max 3 attempts.
	if attemptCount >= 3 {
		log.Error(sctx.Ctx, "RequestMissingDocuments: max attempts reached for %s (attempt %d)", req.RequestID, attemptCount)
		audit := domain.AuditLog{
			RequestID: req.RequestID,
			// ActionType:    "DOCUMENTS_EXPIRED",
			ActionType:    "STATUS_CHANGE", // was "DOCUMENTS_EXPIRED"
			NewValueJSON:  `{"reason":"max_attempts_reached"}`,
			PerformedByID: req.RequestedBy,
			Notes:         strPtr("Maximum document request attempts exceeded (BR-NFS-014)"),
		}
		_, _ = h.srRepo.UpdateStatus(sctx.Ctx, req.RequestID, "DOCUMENTS_EXPIRED", strPtr(req.RequestedBy), nil, nil, &audit)
		return nil, &apierrors.AppError{Code: 422, Message: fmt.Sprintf("maximum document request attempts (%d/3) reached; request expired (BR-NFS-014)", attemptCount)}
	}

	expiresAt := time.Now().AddDate(0, 0, 7)
	token, err := h.srRepo.CreateSecureUploadLink(sctx.Ctx, req.RequestID, req.MissingDocuments, req.RequestedBy, expiresAt)
	if err != nil {
		log.Error(sctx.Ctx, "RequestMissingDocuments: failed to create secure upload link for %s: %v", req.RequestID, err)
		return nil, err
	}

	audit := domain.AuditLog{
		RequestID: req.RequestID,
		// ActionType:    "PENDING_DOCUMENTS",
		ActionType:    "MISSING_DOC_REQUESTED", // was "PENDING_DOCUMENTS"
		NewValueJSON:  fmt.Sprintf(`{"missing_docs":%d,"message":"%s","attempt":%d}`, len(req.MissingDocuments), req.MessageToCustomer, attemptCount+1),
		PerformedByID: req.RequestedBy,
		Notes:         strPtr("Missing documents requested from customer"),
	}
	_, err = h.srRepo.UpdateStatus(sctx.Ctx, req.RequestID, "PENDING_DOCUMENTS", strPtr(req.RequestedBy), nil, nil, &audit)
	if err != nil {
		log.Error(sctx.Ctx, "RequestMissingDocuments: failed to update status for %s: %v", req.RequestID, err)
		return nil, err
	}

	_ = sr
	baseURL := h.cfg.GetString("nfs.secureupload.baseurl")
	secureURL := fmt.Sprintf("%s/%s", baseURL, token)
	log.Info(sctx.Ctx, "RequestMissingDocuments: secure upload link created for request %s attempt %d", req.RequestID, attemptCount+1)

	r := resp.NewRequestMissingDocumentsResponse(
		req.RequestID,
		"PENDING_DOCUMENTS",
		token,
		secureURL,
		expiresAt.Format(time.RFC3339),
		attemptCount+1,
		req.MissingDocuments,
	)
	return r, nil
}

// ✅
func newUUID() string {
	return uuid.New().String()
}
