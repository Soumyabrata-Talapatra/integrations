// Package handler provides HTTP handlers for the Customer NFS service.
// Phase: Phase 1 — Address Change
// Handler: AddressChangeHandler — covers CORE-001..004
package handler

import (
	"fmt"
	"strconv"
	"strings"
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

// AddressChangeHandler handles HTTP endpoints for the address change core journey.
// FR-NFS-001 (Aadhaar path), FR-NFS-002 (Manual path), FR-NFS-010 (Multi-channel).
// WF-NFS-001 (Aadhaar workflow), WF-NFS-002 (Manual workflow).
type AddressChangeHandler struct {
	*serverHandler.Base

	srRepo   *repo.ServiceRequestRepository
	addrRepo *repo.AddressChangeRepository
	docRepo  *repo.DocumentUploadRepository
	tc       client.Client
	cfg      *config.Config
}

// NewAddressChangeHandler constructs the handler and wires dependencies via Uber FX.
// Template: serverHandler.New("AddressChange").SetPrefix("/v1").AddPrefix("")
func NewAddressChangeHandler(
	srRepo *repo.ServiceRequestRepository,
	addrRepo *repo.AddressChangeRepository,
	docRepo *repo.DocumentUploadRepository,
	tc client.Client,
	cfg *config.Config,
) *AddressChangeHandler {
	base := serverHandler.New("AddressChange").
		SetPrefix("/v1").
		AddPrefix("")
	return &AddressChangeHandler{
		Base:     base,
		srRepo:   srRepo,
		addrRepo: addrRepo,
		docRepo:  docRepo,
		tc:       tc,
		cfg:      cfg,
	}
}

// Routes registers all address-change HTTP routes.
// CORE-001: POST /nfs/address-change/initiate
// CORE-002: POST /nfs/address-change/:request_id/verify-otp
// CORE-003: POST /nfs/address-change/:request_id/submit
// CORE-004: POST /nfs/address-change/:request_id/approve
func (h *AddressChangeHandler) Routes() []serverRoute.Route {
	return []serverRoute.Route{
		serverRoute.POST("/nfs/address-change/initiate", h.InitiateAddressChange).Name("Initiate Address Change"),
		serverRoute.POST("/nfs/address-change/:request_id/verify-otp", h.VerifyOTP).Name("Verify Address Change OTP"),
		serverRoute.POST("/nfs/address-change/:request_id/submit", h.SubmitAddressChange).Name("Submit Manual Address Change"),
		serverRoute.POST("/nfs/address-change/:request_id/approve", h.ApproveAddressChange).Name("CPC Approve Address Change"),
	}
}

// ─── CORE-001 ─────────────────────────────────────────────────────────────────

// InitiateAddressChange creates a new address change service request and starts the
// appropriate Temporal workflow.
//
// FR-NFS-001 (Aadhaar path), FR-NFS-002 (Manual path).
// BR-NFS-001: AADHAAR → immediate via OTP; BR-NFS-002: MANUAL → SLA-driven CPC approval.
// BR-NFS-005: only INSURED is handled; other roles returned as 422.
// BR-NFS-006: pincode-state validation delegated to ValidateAddressRequest activity.
// BR-NFS-011: ticket number generated as NFS-ANC-{YYYYMMDD}-{SEQ6}.
// BR-NFS-015: AADHAAR only from Portal/Mobile channels.
// VR-NFS-015: duplicate pending request check.
// WF-NFS-001 started for AADHAAR; WF-NFS-002 started for MANUAL.
func (h *AddressChangeHandler) InitiateAddressChange(
	sctx *serverRoute.Context,
	req InitiateAddressChangeRequest,
) (*resp.AddressChangeInitiateResponse, error) {
	// BR-NFS-005: Reject non-INSURED roles immediately — Policy Service handles them.
	if req.AddressUpdateFor != "INSURED" {
		log.Error(sctx.Ctx, "InitiateAddressChange: non-insured role rejected: %s", req.AddressUpdateFor)
		return nil, &apierrors.AppError{Code: 422, Message: fmt.Sprintf("address change for %s is handled by Policy Service", req.AddressUpdateFor)}
	}

	// BR-NFS-015: AADHAAR only allowed from Portal and Mobile channels.
	if req.AuthMethod == "AADHAAR" && req.Channel != "Portal" && req.Channel != "Mobile" {
		log.Error(sctx.Ctx, "InitiateAddressChange: AADHAAR not allowed from channel: %s", req.Channel)
		return nil, &apierrors.AppError{Code: 422, Message: fmt.Sprintf("Aadhaar auto-approval not available from %s. Use MANUAL method", req.Channel)}
	}

	// VR-NFS-015: Check for duplicate pending request.
	isDuplicate, existingTicket, err := h.srRepo.CheckDuplicateRequest(sctx.Ctx, req.CustomerID, "ADDRESS_CHANGE")
	if err != nil {
		log.Error(sctx.Ctx, "InitiateAddressChange: duplicate check failed: %v", err)
		return nil, err
	}
	if isDuplicate {
		log.Error(sctx.Ctx, "InitiateAddressChange: duplicate pending request %s for customer %d", existingTicket, req.CustomerID)
		return nil, &apierrors.AppError{Code: 409, Message: fmt.Sprintf("pending request exists: %s. Complete or withdraw it first", existingTicket)}
	}

	// BR-NFS-011: Generate ticket number.
	ticketNumber, err := h.srRepo.GenerateTicketNumber(sctx.Ctx, "ADDRESS_CHANGE")
	if err != nil {
		log.Error(sctx.Ctx, "InitiateAddressChange: ticket number generation failed: %v", err)
		return nil, err
	}

	// Map request to domain objects.
	sr, addr := req.ToDomain()
	sr.TicketNumber = ticketNumber
	sr.RequestID = uuid.New().String()
	customerIDStr := strconv.FormatInt(req.CustomerID, 10)
	sr.InitiatedBy = req.CustomerID
	sr.CreatedBy = customerIDStr
	addr.DetailID = uuid.New().String()
	addr.RequestID = sr.RequestID
	addr.AddressUpdateFor = req.AddressUpdateFor
	addr.CreatedBy = customerIDStr

	// Prepare initial audit log entry (BR-NFS-016: INSERT-only).
	audit := domain.AuditLog{
		AuditID:       uuid.New().String(),
		RequestID:     sr.RequestID,
		ActionType:    "CREATED",
		NewValueJSON:  fmt.Sprintf(`{"ticket_number":"%s","auth_method":"%s","channel":"%s"}`, ticketNumber, req.AuthMethod, req.Channel),
		PerformedByID: customerIDStr,
		IPAddress:     nil,
		Notes:         strPtr("Request initiated"),
	}

	// BATCH: CreateWithAddressDetail performs 3-insert TX batch.
	err = h.srRepo.CreateWithAddressDetail(sctx.Ctx, &sr, &addr, &audit)
	if err != nil {
		log.Error(sctx.Ctx, "InitiateAddressChange: failed to create service request: %v", err)
		return nil, err
	}

	// Build workflow input.
	slaDays := h.cfg.GetInt("nfs.sladeadlineregional")
	if slaDays == 0 {
		slaDays = 15
	}
	wfInput := workflows.AddressChangeWorkflowInput{
		RequestID:    sr.RequestID,
		TicketNumber: sr.TicketNumber,
		CustomerID:   req.CustomerID,
		PolicyNumber: req.PolicyNumber,
		AuthMethod:   req.AuthMethod,
		AddressType:  req.AddressType,
		Channel:      req.Channel,
		OfficeCode:   strDeref(req.OfficeCode),
		InitiatedBy:  customerIDStr,
		SLADays:      slaDays,
	}

	workflowID := fmt.Sprintf("nfs-address-%s-%s", req.AuthMethod, sr.RequestID)
	options := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: workflows.TaskQueue,
	}

	var otpSent bool
	var slaDeadline *string

	if req.AuthMethod == "AADHAAR" {
		// WF-NFS-001: Aadhaar address change workflow.
		_, err = h.tc.ExecuteWorkflow(sctx.Ctx, options, workflows.AadhaarAddressChangeWorkflow, wfInput)
		if err != nil {
			log.Error(sctx.Ctx, "InitiateAddressChange: failed to start AadhaarAddressChangeWorkflow: %v", err)
			return nil, err
		}
		otpSent = true
		log.Info(sctx.Ctx, "InitiateAddressChange: started WF-NFS-001 workflowID=%s", workflowID)
	} else {
		// WF-NFS-002: Manual address change workflow.
		// BR-NFS-002: Set SLA deadline before workflow starts to satisfy chk_sla_for_manual.
		dl := time.Now().AddDate(0, 0, slaDays)
		if err := h.srRepo.UpdateSLADeadline(sctx.Ctx, sr.RequestID, dl); err != nil {
			log.Error(sctx.Ctx, "InitiateAddressChange: failed to set SLA deadline for %s: %v", sr.RequestID, err)
			return nil, err
		}
		deadline := dl.Format("2006-01-02")
		slaDeadline = &deadline

		_, err = h.tc.ExecuteWorkflow(sctx.Ctx, options, workflows.ManualAddressChangeWorkflow, wfInput)
		if err != nil {
			log.Error(sctx.Ctx, "InitiateAddressChange: failed to start ManualAddressChangeWorkflow: %v", err)
			return nil, err
		}
		log.Info(sctx.Ctx, "InitiateAddressChange: started WF-NFS-002 workflowID=%s slaDeadline=%s", workflowID, deadline)
	}

	r := resp.NewAddressChangeInitiateResponse(&sr, slaDeadline, otpSent)
	return r, nil
}

// ─── CORE-002 ─────────────────────────────────────────────────────────────────

// VerifyOTP accepts the Aadhaar OTP from the customer and signals the running Temporal
// workflow (WF-NFS-001) to proceed with OTP verification and address update.
//
// FR-NFS-001: Aadhaar OTP step.
// BR-NFS-001: successful OTP → status immediately COMPLETED.
// WF-NFS-001: signal "otp_submitted" dispatched to the workflow.
// WORKFLOW STATE: workflow_id/workflow_run_id fetched from DB (stored by StoreWorkflowState activity).
func (h *AddressChangeHandler) VerifyOTP(
	sctx *serverRoute.Context,
	req VerifyOTPRequest,
) (*resp.AddressChangeVerifyOTPResponse, error) {
	// Fetch the current service request to validate existence and current status.
	sr, err := h.srRepo.GetByID(sctx.Ctx, req.RequestID)
	if err != nil {
		if err == pgx.ErrNoRows {
			log.Error(sctx.Ctx, "VerifyOTP: request not found: %s", req.RequestID)
			return nil, &apierrors.AppError{Code: 404, Message: fmt.Sprintf("request not found: %s", req.RequestID)}
		}
		log.Error(sctx.Ctx, "VerifyOTP: failed to fetch request %s: %v", req.RequestID, err)
		return nil, err
	}

	// WORKFLOW STATE: Retrieve workflow_id and workflow_run_id stored by StoreWorkflowState activity.
	wfID, wfRunID, err := h.srRepo.GetWorkflowState(sctx.Ctx, req.RequestID)
	if err != nil {
		log.Error(sctx.Ctx, "VerifyOTP: failed to get workflow state for %s: %v", req.RequestID, err)
		return nil, err
	}

	// Signal WF-NFS-001 with the OTP payload so the workflow can call VerifyAadhaarOTPActivity.
	otpPayload := workflows.OTPSignalPayload{
		OTP:            req.OTPCode,
		OTPReferenceID: req.TxnID,
		CustomerID:     sr.CustomerID,
	}
	err = h.tc.SignalWorkflow(sctx.Ctx, wfID, wfRunID, workflows.SignalOTPSubmitted, otpPayload)
	if err != nil {
		log.Error(sctx.Ctx, "VerifyOTP: failed to signal workflow %s: %v", wfID, err)
		return nil, err
	}

	log.Info(sctx.Ctx, "VerifyOTP: signal sent to workflow %s for request %s", wfID, req.RequestID)

	r := resp.NewAddressVerifyOTPResponse(
		sr.RequestID,
		sr.TicketNumber,
		"IN_PROGRESS",
		"OTP submitted. Address change is being processed.",
	)
	return r, nil
}

// ─── CORE-003 ─────────────────────────────────────────────────────────────────

// SubmitAddressChange finalises the manual address change request after document upload.
// Validates documents are present, signals the running WF-NFS-002 workflow, and transitions
// the request status to PENDING_APPROVAL.
//
// FR-NFS-002: Manual path — submission after document upload.
// BR-NFS-002: status → PENDING_APPROVAL; SLA clock starts.
// BR-NFS-012: valid transition CREATED → PENDING_APPROVAL.
// VR-NFS-009, VR-NFS-010: document validation is handled during DM upload (not repeated here).
// WF-NFS-002: signal "documents_submitted" sent so workflow can move to AssignToCPC phase.
// func (h *AddressChangeHandler) SubmitAddressChange(
// 	sctx *serverRoute.Context,
// 	req SubmitAddressChangeRequest,
// ) (*resp.AddressChangeSubmitResponse, error) {
// 	// Validate request exists and is in CREATED status.
// 	sr, err := h.srRepo.GetByID(sctx.Ctx, req.RequestID)
// 	if err != nil {
// 		if err == pgx.ErrNoRows {
// 			log.Error(sctx.Ctx, "SubmitAddressChange: request not found: %s", req.RequestID)
// 			return nil, &apierrors.AppError{Code: 404, Message: fmt.Sprintf("request not found: %s", req.RequestID)}
// 		}
// 		log.Error(sctx.Ctx, "SubmitAddressChange: failed to fetch request %s: %v", req.RequestID, err)
// 		return nil, err
// 	}

// 	// BR-NFS-012: only CREATED requests can be submitted.
// 	if sr.Status != "CREATED" {
// 		log.Error(sctx.Ctx, "SubmitAddressChange: invalid status %s for request %s", sr.Status, req.RequestID)
// 		return nil, &apierrors.AppError{Code: 422, Message: fmt.Sprintf("request %s is in status %s; only CREATED requests can be submitted", req.RequestID, sr.Status)}
// 	}

// 	// ERR-NFS-ANC-004: At least one document (Address Proof) is required.
// 	if len(req.UploadedDocuments) == 0 {
// 		log.Error(sctx.Ctx, "SubmitAddressChange: no documents uploaded for request %s", req.RequestID)
// 		return nil, &apierrors.AppError{Code: 422, Message: "at least one document is required for manual address change submission"}
// 	}

// 	// Build audit log for status transition (BR-NFS-016: INSERT-only).
// 	audit := domain.AuditLog{
// 		AuditID:       uuid.New().String(), // ← add
// 		RequestID:     req.RequestID,
// 		ActionType:    "PENDING_APPROVAL",
// 		NewValueJSON:  fmt.Sprintf(`{"documents_count":%d,"submitted_by":"%s"}`, len(req.UploadedDocuments), req.SubmittedBy),
// 		PerformedByID: req.SubmittedBy,
// 		Notes:         strPtr("Manual submission with documents"),
// 	}

// 	// BATCH: UpdateStatus performs TX batch: UPDATE service_request + INSERT status_transition + INSERT audit_log.
// 	updatedSR, err := h.srRepo.UpdateStatus(sctx.Ctx, req.RequestID, "PENDING_APPROVAL", strPtr(req.SubmittedBy), nil, nil, &audit)
// 	if err != nil {
// 		log.Error(sctx.Ctx, "SubmitAddressChange: failed to update status for %s: %v", req.RequestID, err)
// 		return nil, err
// 	}

// 	// WORKFLOW STATE: Signal WF-NFS-002 so it can proceed to AssignToCPC.
// 	wfID, wfRunID, err := h.srRepo.GetWorkflowState(sctx.Ctx, req.RequestID)
// 	if err != nil {
// 		log.Error(sctx.Ctx, "SubmitAddressChange: failed to get workflow state for %s: %v", req.RequestID, err)
// 		return nil, err
// 	}

// 	// Signal the workflow with the document submission payload.
// 	submitPayload := activities.DocumentsSubmittedPayload{
// 		RequestID:         req.RequestID,
// 		UploadedDocuments: req.UploadedDocuments,
// 		SubmittedBy:       req.SubmittedBy,
// 	}
// 	err = h.tc.SignalWorkflow(sctx.Ctx, wfID, wfRunID, workflows.SignalDocumentsSubmitted, submitPayload)
// 	if err != nil {
// 		log.Error(sctx.Ctx, "SubmitAddressChange: failed to signal workflow %s: %v", wfID, err)
// 		return nil, err
// 	}

// 	log.Info(sctx.Ctx, "SubmitAddressChange: request %s moved to PENDING_APPROVAL, workflow %s signaled", req.RequestID, wfID)

// 	// BR-NFS-002: calculate SLA deadline from now.
// 	slaDays := h.cfg.GetInt("nfs.sladeadlineregional")
// 	if slaDays == 0 {
// 		slaDays = 15
// 	}
// 	slaDeadline := time.Now().AddDate(0, 0, slaDays).Format("2006-01-02")

//		r := resp.NewAddressSubmitResponse(
//			updatedSR.RequestID,
//			updatedSR.TicketNumber,
//			updatedSR.UpdatedAt.Format(time.RFC3339),
//			slaDeadline,
//		)
//		return r, nil
//	}
func (h *AddressChangeHandler) SubmitAddressChange(
	sctx *serverRoute.Context,
	req SubmitAddressChangeRequest,
) (*resp.AddressChangeSubmitResponse, error) {
	// Validate request exists and is in CREATED status.
	sr, err := h.srRepo.GetByID(sctx.Ctx, req.RequestID)
	if err != nil {
		if err == pgx.ErrNoRows {
			log.Error(sctx.Ctx, "SubmitAddressChange: request not found: %s", req.RequestID)
			return nil, &apierrors.AppError{Code: 404, Message: fmt.Sprintf("request not found: %s", req.RequestID)}
		}
		log.Error(sctx.Ctx, "SubmitAddressChange: failed to fetch request %s: %v", req.RequestID, err)
		return nil, err
	}

	// BR-NFS-012: only CREATED requests can be submitted.
	if sr.Status != "CREATED" && sr.Status != "PENDING_DOCUMENTS" {
		log.Error(sctx.Ctx, "SubmitAddressChange: invalid status %s for request %s", sr.Status, req.RequestID)
		return nil, &apierrors.AppError{Code: 422, Message: fmt.Sprintf("request %s is in status %s; only CREATED requests can be submitted", req.RequestID, sr.Status)}
	}

	// ERR-NFS-ANC-004: At least one document (Address Proof) is required.
	if len(req.UploadedDocuments) == 0 {
		log.Error(sctx.Ctx, "SubmitAddressChange: no documents uploaded for request %s", req.RequestID)
		return nil, &apierrors.AppError{Code: 422, Message: "at least one document is required for manual address change submission"}
	}

	// BR-NFS-002: Set SLA deadline before status transition (required by chk_sla_for_manual).
	slaDays := h.cfg.GetInt("nfs.sladeadlineregional")
	if slaDays == 0 {
		slaDays = 15
	}
	dl := time.Now().AddDate(0, 0, slaDays)
	if sr.Status == "CREATED" {
		if err := h.srRepo.UpdateSLADeadline(sctx.Ctx, req.RequestID, dl); err != nil {
			log.Error(sctx.Ctx, "SubmitAddressChange: failed to set SLA deadline for %s: %v", req.RequestID, err)
			return nil, err
		}
	}

	// Build audit log for status transition (BR-NFS-016: INSERT-only).
	audit := domain.AuditLog{
		AuditID:       uuid.New().String(),
		RequestID:     req.RequestID,
		ActionType:    "STATUS_CHANGE", // ← fix
		NewValueJSON:  fmt.Sprintf(`{"documents_count":%d,"submitted_by":"%s"}`, len(req.UploadedDocuments), req.SubmittedBy),
		PerformedByID: req.SubmittedBy,
		Notes:         strPtr("Manual submission with documents"),
	}

	// BATCH: UpdateStatus performs TX batch: UPDATE service_request + INSERT status_transition + INSERT audit_log.
	updatedSR, err := h.srRepo.UpdateStatus(sctx.Ctx, req.RequestID, "PENDING_APPROVAL", strPtr(req.SubmittedBy), nil, nil, &audit)
	if err != nil {
		log.Error(sctx.Ctx, "SubmitAddressChange: failed to update status for %s: %v", req.RequestID, err)
		return nil, err
	}

	// WORKFLOW STATE: Signal WF-NFS-002 so it can proceed to AssignToCPC.
	wfID, wfRunID, err := h.srRepo.GetWorkflowState(sctx.Ctx, req.RequestID)
	if err != nil {
		log.Error(sctx.Ctx, "SubmitAddressChange: failed to get workflow state for %s: %v", req.RequestID, err)
		return nil, err
	}

	// Signal the workflow with the document submission payload.
	submitPayload := activities.DocumentsSubmittedPayload{
		RequestID:         req.RequestID,
		UploadedDocuments: req.UploadedDocuments,
		SubmittedBy:       req.SubmittedBy,
	}
	err = h.tc.SignalWorkflow(sctx.Ctx, wfID, wfRunID, workflows.SignalDocumentsSubmitted, submitPayload)
	if err != nil {
		if strings.Contains(err.Error(), "workflow execution already completed") {
			return nil, &apierrors.AppError{Code: 422, Message: "workflow has already completed for this request"}
		}
		log.Error(sctx.Ctx, "SubmitAddressChange: failed to signal workflow %s: %v", wfID, err)
		return nil, err
	}

	log.Info(sctx.Ctx, "SubmitAddressChange: request %s moved to PENDING_APPROVAL, workflow %s signaled", req.RequestID, wfID)

	r := resp.NewAddressSubmitResponse(
		updatedSR.RequestID,
		updatedSR.TicketNumber,
		updatedSR.UpdatedAt.Format(time.RFC3339),
		dl.Format("2006-01-02"),
	)
	return r, nil
}

// ─── CORE-004 ─────────────────────────────────────────────────────────────────

// ApproveAddressChange records the CPC decision and signals the running WF-NFS-002 workflow
// to complete the address change, reject it, or request missing documents.
//
// FR-NFS-002, FR-NFS-006, FR-NFS-008, FR-NFS-011.
// BR-NFS-002: CPC approval required for manual path.
// BR-NFS-003: on APPROVE → address propagated to downstream services via workflow.
// BR-NFS-004: address versioning handled inside the workflow activity.
// BR-NFS-012: IN_PROGRESS → COMPLETED / REJECTED / PENDING_DOCUMENTS.
// BR-NFS-016: audit log created (action_type = APPROVED / REJECTED).
// WF-NFS-002: signal "approval_decision" sent to the workflow.
// ERR-NFS-SR-004: access control (CPC role) enforced by middleware.
// ERR-NFS-SR-005: rejection_reason is mandatory when decision=REJECT.
func (h *AddressChangeHandler) ApproveAddressChange(
	sctx *serverRoute.Context,
	req ApproveAddressChangeRequest,
) (*resp.AddressChangeApproveResponse, error) {
	// ERR-NFS-SR-005: rejection_reason is mandatory when decision=REJECT.
	if req.Decision == "REJECT" && (req.RejectionReason == nil || *req.RejectionReason == "") {
		log.Error(sctx.Ctx, "ApproveAddressChange: rejection reason missing for REJECT decision on %s", req.RequestID)
		return nil, &apierrors.AppError{Code: 422, Message: "rejection_reason is required when decision is REJECT"}
	}

	// Validate request exists.
	sr, err := h.srRepo.GetByID(sctx.Ctx, req.RequestID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, &apierrors.AppError{Code: 404, Message: fmt.Sprintf("request not found: %s", req.RequestID)}
		}
		return nil, err
	}
	if sr.Status != "IN_PROGRESS" {
		log.Error(sctx.Ctx, "ApproveNameChange: invalid status %s for request %s", sr.Status, req.RequestID)
		return nil, &apierrors.AppError{Code: 422, Message: fmt.Sprintf("request %s is in status %s; only IN_PROGRESS requests can be approved", req.RequestID, sr.Status)}
	}

	// WORKFLOW STATE: Retrieve workflow_id and workflow_run_id stored by StoreWorkflowState activity.
	wfID, wfRunID, err := h.srRepo.GetWorkflowState(sctx.Ctx, req.RequestID)
	if err != nil {
		log.Error(sctx.Ctx, "ApproveAddressChange: failed to get workflow state for %s: %v", req.RequestID, err)
		return nil, err
	}

	// Map decision to the new status for response construction.
	// BR-NFS-012: IN_PROGRESS → COMPLETED / REJECTED / PENDING_DOCUMENTS.
	newStatus := decisionToStatus(req.Decision)

	// Signal WF-NFS-002 with the approval decision payload.
	approvalPayload := workflows.ApprovalDecisionPayload{
		Decision:         req.Decision,
		ApprovedBy:       strconv.FormatInt(req.ApprovedBy, 10),
		Reason:           strDeref(req.Remarks),
		RejectionReason:  req.RejectionReason,
		MissingDocuments: req.MissingDocuments,
	}
	err = h.tc.SignalWorkflow(sctx.Ctx, wfID, wfRunID, workflows.SignalApprovalDecision, approvalPayload)
	if err != nil {
		log.Error(sctx.Ctx, "ApproveAddressChange: failed to signal workflow %s: %v", wfID, err)
		return nil, err
	}

	log.Info(sctx.Ctx, "ApproveAddressChange: decision %s recorded for request %s, workflow %s signaled", req.Decision, req.RequestID, wfID)

	_ = sr // sr used for future access-control checks (ERR-NFS-SR-004)

	r := resp.NewAddressApproveResponse(
		req.RequestID,
		sr.TicketNumber,
		newStatus,
		req.Decision,
		time.Now().Format(time.RFC3339),
	)
	return r, nil
}

// ─── helpers ─────────────────────────────────────────────────────────────────

// decisionToStatus maps CPC decision strings to expected new request statuses.
// BR-NFS-012: state transition rules.
func decisionToStatus(decision string) string {
	switch decision {
	case "APPROVE":
		return "COMPLETED"
	case "REJECT":
		return "REJECTED"
	case "SEND_BACK":
		return "PENDING_DOCUMENTS"
	default:
		return "UNKNOWN"
	}
}

// strPtr returns a pointer to the given string.
func strPtr(s string) *string { return &s }

// strDeref safely dereferences a string pointer, returning "" if nil.
func strDeref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
