package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"time"

	"plirevival/core/domain"
	"plirevival/core/port"
	repo "plirevival/repo/postgres"
	"plirevival/workflow"

	config "gitlab.cept.gov.in/it-2.0-common/api-config"
	apierrors "gitlab.cept.gov.in/it-2.0-common/n-api-errors"
	log "gitlab.cept.gov.in/it-2.0-common/n-api-log"
	serverHandler "gitlab.cept.gov.in/it-2.0-common/n-api-server/handler"
	serverRoute "gitlab.cept.gov.in/it-2.0-common/n-api-server/route"
	"go.temporal.io/sdk/client"
	tclient "go.temporal.io/sdk/client"
)

type RevivalHandler struct {
	*serverHandler.Base
	revivalRepo    *repo.RevivalRepository
	policyRepo     *repo.PolicyRepository
	paymentRepo    *repo.PaymentRepository
	activities     *workflow.Activities
	temporalClient tclient.Client
	taskQueue      string
}

func NewRevivalHandler(revivalRepo *repo.RevivalRepository, policyRepo *repo.PolicyRepository, paymentRepo *repo.PaymentRepository, activities *workflow.Activities, temporalClient tclient.Client, cfg *config.Config) *RevivalHandler {
	base := serverHandler.New("Revival").SetPrefix("/v1").AddPrefix("/revival")
	return &RevivalHandler{
		Base:           base,
		revivalRepo:    revivalRepo,
		policyRepo:     policyRepo,
		paymentRepo:    paymentRepo,
		activities:     activities,
		temporalClient: temporalClient,
		taskQueue:      cfg.GetString("temporal.taskqueue"),
	}
}

func (h *RevivalHandler) Routes() []serverRoute.Route {
	return []serverRoute.Route{

		// Policy and eligibility endpoints
		serverRoute.GET("/policies/:policy_number", h.GetPolicy).Name("Get Policy"),
		serverRoute.GET("/policies/:policy_number/eligibility", h.GetPolicyEligibility).Name("Get Policy Eligibility"),
		serverRoute.GET("/:policy_number", h.GetRevivalHistory).Name("Get Revival History"),

		// Get all revival requests
		serverRoute.GET("/requests", h.GetAllRequests).Name("Get All Revival Requests"),
		serverRoute.GET("/requests/:ticket_id", h.GetRevivalRequest).Name("Get Revival Request"),

		// Indexing and data entry endpoints
		serverRoute.POST("/requests/index", h.IndexRevivalRequest).Name("Index Revival Request"),
		serverRoute.POST("/requests/:ticket_id/data-entry", h.SubmitDataEntry).Name("Submit Data Entry"),
		serverRoute.POST("/requests/:ticket_id/quality-check", h.SubmitQualityCheck).Name("Submit Quality Check"),

		// Approval and rejection endpoints
		serverRoute.POST("/requests/:ticket_id/approval-decision", h.ApprovalDecision).Name("Approval Decision"),
		serverRoute.POST("/requests/:ticket_id/approval-redirect", h.ApprovalRedirect).Name("Approval Redirect"),

		// Withdrawal endpoint
		serverRoute.POST("/requests/:ticket_id/withdraw", h.WithdrawRequest).Name("Withdraw Request"),

		// Collection endpoints - batch endpoints (no ticket_id in URL)
		serverRoute.POST("/first-collections", h.BatchFirstCollection).Name("Batch First Collection"),
		serverRoute.POST("/installments", h.BatchInstallments).Name("Batch Installments"),

		// Get installment collection details
		serverRoute.GET("/requests/installments", h.GetInstallmentCollection).Name("Get Installment Collection By Policy"),

		// Quotation endpoint
		serverRoute.POST("/policies/:policy_number/quote", h.Quotation).Name("Quotation"),

		// Document endpoints
		serverRoute.POST("/requests/:ticket_id/documents", h.ReceiveDocuments).Name("Receive Documents"),
		serverRoute.POST("/requests/:ticket_id/acceptance-letter", h.GenerateAcceptanceLetter).Name("Generate Acceptance Letter"),
		serverRoute.POST("/requests/:ticket_id/revival-memo", h.GenerateRevivalMemo).Name("Generate Revival Memo"),
	}
}

// GetPolicy retrieves policy details
func (h *RevivalHandler) GetPolicy(sctx *serverRoute.Context, req port.PolicyNumberUri) (*port.PolicyDetailsResponse, error) {
	// Get policy from repository
	policy, err := h.policyRepo.GetPolicyByNumber(sctx.Ctx, req.PolicyNumber)
	if err != nil {
		errMsg := apierrors.HandleErrorWithStatusCodeAndMessage(
			apierrors.DBErrorRecordNotFound,
			"No policy found for the given policy number",
			err,
		)
		return nil, errMsg
	}

	// Construct product name (Product Code - Product Name)
	product := policy.ProductCode
	if policy.ProductName != "" {
		product = fmt.Sprintf("%s - %s", policy.ProductCode, policy.ProductName)
	}

	return &port.PolicyDetailsResponse{
		StatusCodeAndMessage: port.FetchSuccess,
		Data: port.PolicyDetailsData{
			PolicyNumber:       policy.PolicyNumber,
			CustomerName:       policy.CustomerName,
			CustomerID:         policy.CustomerID,
			Product:            product,
			ProductCode:        policy.ProductCode,
			SumAssured:         policy.SumAssured,
			PremiumAmount:      policy.PremiumAmount,
			PremiumFrequency:   policy.PremiumFrequency,
			LapseDate:          policy.PaidToDate, // Lapse date is typically the paid_to_date for lapsed policies
			RevivalCount:       policy.RevivalCount,
			PolicyStatus:       policy.PolicyStatus,
			DateOfCommencement: policy.DateOfCommencement,
			MaturityDate:       policy.MaturityDate,
			PaidToDate:         policy.PaidToDate,
			LastRevivalDate:    policy.LastRevivalDate,
		},
	}, nil
}

// GetPolicyEligibility checks policy eligibility for revival
func (h *RevivalHandler) GetPolicyEligibility(sctx *serverRoute.Context, req port.PolicyNumberUri) (*port.PolicyEligibilityResponse, error) {
	// Get policy from repository
	policy, err := h.policyRepo.GetPolicyByNumber(sctx.Ctx, req.PolicyNumber)
	if err != nil {
		errMsg := apierrors.HandleErrorWithStatusCodeAndMessage(
			apierrors.DBErrorRecordNotFound,
			"No policy found for the given policy number",
			err,
		)
		return nil, errMsg
	}

	// Get max revivals allowed (IR_29)
	maxRevivals, err := h.policyRepo.GetMaxRevivalsAllowed(sctx.Ctx)
	if err != nil {
		maxRevivals = 2 // Default to 2 if config not found
	}

	// Check if policy is in lapsed status (AL)
	eligible := policy.PolicyStatus == "AL"

	// Check if max revivals exceeded (IR_29)
	if policy.RevivalCount >= maxRevivals {
		eligible = false
	}

	// Check for ongoing revival requests
	hasOngoing, err := h.revivalRepo.CheckOngoingRevival(sctx.Ctx, req.PolicyNumber)
	if err == nil && hasOngoing {
		eligible = false
	}

	return &port.PolicyEligibilityResponse{
		StatusCodeAndMessage: port.FetchSuccess,
		Data: port.PolicyEligibilityData{
			PolicyNumber:       policy.PolicyNumber,
			Eligible:           eligible,
			RevivalCount:       policy.RevivalCount,
			MaxRevivalsAllowed: maxRevivals,
			MaxInstallments:    12,
		},
	}, nil
}

// GetRevivalHistory retrieves revival history for policy
func (h *RevivalHandler) GetRevivalHistory(sctx *serverRoute.Context, req port.PolicyNumberUri) (*port.RevivalHistoryResponse, error) {
	// Get revival requests from repository
	revivals, err := h.revivalRepo.GetRevivalRequestsByPolicyNumber(sctx.Ctx, req.PolicyNumber)
	if err != nil {
		errMsg := apierrors.HandleErrorWithStatusCodeAndMessage(
			apierrors.CustomError,
			"Unable to fetch revival history",
			err,
		)
		return nil, errMsg
	}

	// Build revival history entries
	history := make([]port.RevivalHistoryEntry, 0, len(revivals))
	completedCount := 0

	for _, revival := range revivals {
		entry := port.RevivalHistoryEntry{
			TicketID:          revival.TicketID,
			RequestDate:       revival.CreatedAt,
			Status:            revival.CurrentStatus,
			RevivalAmount:     revival.RevivalAmount,
			InstallmentAmount: revival.InstallmentAmount,
			TotalInstallments: revival.NumberOfInstallments,
			InstallmentsPaid:  revival.InstallmentsPaid,
		}

		if revival.CurrentStatus == "COMPLETED" {
			completedCount++
		}

		history = append(history, entry)
	}

	return &port.RevivalHistoryResponse{
		StatusCodeAndMessage: port.FetchSuccess,
		Data: port.RevivalHistoryData{
			PolicyNumber:           req.PolicyNumber,
			RevivalCount:           len(revivals),
			CompletedRevivalsCount: completedCount,
			RevivalHistory:         history,
		},
	}, nil
}

// GetAllRequests retrieves all revival requests with policy details
func (h *RevivalHandler) GetAllRequests(sctx *serverRoute.Context, req port.EmptyRequest) (*port.GetAllRequestsResponse, error) {
	// Get all revival requests with policy details from repository
	requests, err := h.revivalRepo.GetAllRevivalRequests(sctx.Ctx)
	if err != nil {
		errMsg := apierrors.HandleErrorWithStatusCodeAndMessage(
			apierrors.CustomError,
			"Unable to fetch revival requests",
			err,
		)
		return nil, errMsg
	}

	// Build response items
	items := make([]port.RequestListItem, 0, len(requests))
	for _, req := range requests {
		nextAction, nextActor := determineNextAction(req.CurrentStatus)

		item := port.RequestListItem{
			RequestID:     req.RequestID,
			TicketID:      req.TicketID,
			PolicyNumber:  req.PolicyNumber,
			InsuredName:   req.InsuredName,
			CustomerID:    req.CustomerID,
			RequestType:   req.RequestType,
			RequestStatus: req.CurrentStatus,
			RequestedDate: req.RequestedDate,
			NextAction:    nextAction,
			NextActor:     nextActor,
		}
		items = append(items, item)
	}

	return &port.GetAllRequestsResponse{
		StatusCodeAndMessage: port.FetchSuccess,
		Data:                 items,
	}, nil
}

// GetRevivalRequest retrieves a single revival request with progressive details based on workflow stage
func (h *RevivalHandler) GetRevivalRequest(sctx *serverRoute.Context, req port.GetRevivalRequestUri) (*port.GetRevivalRequestResponse, error) {
	log.Debug(sctx.Ctx, "GetRevivalRequest called", "ticket_id", req.TicketID)

	// Get revival request from repository
	revivalReq, err := h.revivalRepo.GetRevivalRequestByTicketID(sctx.Ctx, req.TicketID)
	if err != nil {
		log.Error(sctx.Ctx, "Failed to get revival request", "error", err)
		errMsg := apierrors.HandleErrorWithStatusCodeAndMessage(
			apierrors.DBErrorRecordNotFound,
			"No revival request found for the given ticket ID",
			err,
		)
		return nil, errMsg
	}
	log.Debug(sctx.Ctx, "Revival request fetched", "policy_number", revivalReq.PolicyNumber)

	// Get policy details for customer information
	policy, err := h.policyRepo.GetPolicyByNumber(sctx.Ctx, revivalReq.PolicyNumber)
	var policyValid bool
	if err != nil {
		// Don't fail if policy not found, just leave fields empty
		log.Warn(sctx.Ctx, "Failed to get policy details", "policy_number", revivalReq.PolicyNumber, "error", err)
		policyValid = false
	} else {
		policyValid = true
	}
	log.Debug(sctx.Ctx, "Policy check complete", "valid", policyValid)

	// Build base response with basic information
	response := port.RevivalRequestDetails{
		RequestID:     revivalReq.RequestID,
		TicketID:      revivalReq.TicketID,
		PolicyNumber:  revivalReq.PolicyNumber,
		RequestType:   revivalReq.RequestType,
		CurrentStatus: revivalReq.CurrentStatus,
		CreatedAt:     revivalReq.CreatedAt,
		WorkflowID:    revivalReq.WorkflowID,
		RunID:         revivalReq.RunID,
	}
	log.Debug(sctx.Ctx, "Base response built")

	// Add policy details if available
	if policyValid {
		response.CustomerName = policy.CustomerName
		response.CustomerID = policy.CustomerID
	}
	log.Debug(sctx.Ctx, "Policy details added")

	// Add indexing details (available at all stages)
	if revivalReq.IndexedBy != nil {
		indexingDetails := &port.IndexingDetails{
			IndexedBy: *revivalReq.IndexedBy,
		}

		if revivalReq.IndexedDate != nil {
			indexingDetails.IndexedDate = *revivalReq.IndexedDate
		}

		response.IndexingDetails = indexingDetails
	}
	log.Debug(sctx.Ctx, "Indexing details processed")

	// Add data entry details if available (DATA_ENTRY_COMPLETE and beyond)
	// Check if data entry has been completed based on status or if data entry fields are populated
	if revivalReq.DataEntryBy != nil || revivalReq.NumberOfInstallments > 0 {
		dataEntryDetails := &port.DataEntryDetails{
			NumberOfInstallments: revivalReq.NumberOfInstallments,
			RevivalAmount:        revivalReq.RevivalAmount,
			InstallmentAmount:    revivalReq.InstallmentAmount,
			SGST:                 revivalReq.SGST,
			CGST:                 revivalReq.CGST,
			Interest:             revivalReq.Interest,
			MedicalExaminerCode:  revivalReq.MedicalExaminerCode,
			MedicalExaminerName:  revivalReq.MedicalExaminerName,
		}

		if revivalReq.DataEntryBy != nil {
			dataEntryDetails.DataEnteredBy = *revivalReq.DataEntryBy
		}
		if revivalReq.DataEntryDate != nil {
			dataEntryDetails.DataEntryTimestamp = revivalReq.DataEntryDate
		}

		// Parse documents submitted during data entry if available
		if revivalReq.Documents != nil && *revivalReq.Documents != "" && *revivalReq.Documents != "[]" {
			var docs []port.DocumentSubmission
			if err := json.Unmarshal([]byte(*revivalReq.Documents), &docs); err == nil {
				dataEntryDetails.DocumentsSubmitted = docs
			}
		}

		// Parse missing documents from data entry if available
		if revivalReq.MissingDocumentsList != nil && *revivalReq.MissingDocumentsList != "" && *revivalReq.MissingDocumentsList != "[]" {
			var missingDocs []port.MissingDocument
			if err := json.Unmarshal([]byte(*revivalReq.MissingDocumentsList), &missingDocs); err == nil {
				dataEntryDetails.MissingDocuments = missingDocs
			}
		}

		response.DataEntryDetails = dataEntryDetails
	}
	log.Debug(sctx.Ctx, "Data entry details processed")

	// Add QC details if available (when QC has been performed)
	if revivalReq.QCBy != nil {
		qcDetails := &port.QCDetails{
			QCPassed: revivalReq.CurrentStatus != "DATA_ENTRY_PENDING", // QC passed if not sent back to data entry
		}

		if revivalReq.QCBy != nil {
			qcDetails.QCPerformedBy = *revivalReq.QCBy
		}
		if revivalReq.QCCompleteDate != nil {
			qcDetails.QCTimestamp = revivalReq.QCCompleteDate
		}
		if revivalReq.QCComments != nil {
			qcDetails.QCComments = *revivalReq.QCComments
		}

		// Parse missing documents updated by QC if available
		if revivalReq.MissingDocumentsList != nil && *revivalReq.MissingDocumentsList != "" && *revivalReq.MissingDocumentsList != "[]" {
			var missingDocs []port.MissingDocument
			if err := json.Unmarshal([]byte(*revivalReq.MissingDocumentsList), &missingDocs); err == nil {
				qcDetails.MissingDocuments = missingDocs
			}
		}

		response.QCDetails = qcDetails
	}
	log.Debug(sctx.Ctx, "QC details processed")

	// Add approval details if available (APPROVED, ACTIVE, COMPLETED, etc.)
	if revivalReq.CurrentStatus == "APPROVED" || revivalReq.CurrentStatus == "ACTIVE" || revivalReq.CurrentStatus == "COMPLETED" || revivalReq.CurrentStatus == "DEFAULTED" || revivalReq.CurrentStatus == "REJECTED" {
		approvalDetails := &port.ApprovalDetails{
			Approved: revivalReq.CurrentStatus != "REJECTED",
		}

		if revivalReq.ApprovedBy != nil {
			approvalDetails.ApprovedBy = *revivalReq.ApprovedBy
		}
		if revivalReq.ApprovalDate != nil {
			approvalDetails.ApprovalTimestamp = revivalReq.ApprovalDate
		}
		if revivalReq.ApprovalComments != nil {
			approvalDetails.Comments = *revivalReq.ApprovalComments
		}

		response.ApprovalDetails = approvalDetails
	}
	log.Debug(sctx.Ctx, "Approval details processed")

	// Fetch status history
	statusHistory, err := h.revivalRepo.GetStatusHistoryByRequestID(sctx.Ctx, revivalReq.RequestID)
	if err == nil && len(statusHistory) > 0 {
		historyEntries := make([]port.StatusHistoryEntry, len(statusHistory))
		for i, h := range statusHistory {
			historyEntries[i] = port.StatusHistoryEntry{
				HistoryID:    h.HistoryID,
				FromStatus:   &h.FromStatus,
				ToStatus:     h.ToStatus,
				ChangedAt:    h.ChangedAt,
				ChangedBy:    h.ChangedBy,
				ChangeReason: h.ChangeReason,
			}
		}
		response.StatusHistory = historyEntries
		log.Debug(sctx.Ctx, "Status history populated", "count", len(historyEntries))
	}

	log.Debug(sctx.Ctx, "GetRevivalRequest response built successfully", "ticket_id", req.TicketID, "response", response)

	return &port.GetRevivalRequestResponse{
		StatusCodeAndMessage: port.FetchSuccess,
		Data:                 response,
	}, nil
}

// IndexRevivalRequest creates a new revival request and starts Temporal workflow
func (h *RevivalHandler) IndexRevivalRequest(sctx *serverRoute.Context, req port.IndexRevivalRequest) (*port.IndexRequestResponse, error) {

	//log.Debug(nil,"req",req)
	// Set request date time if not provided
	requestDateTime := req.RequestDateTime
	if requestDateTime.IsZero() {
		requestDateTime = time.Now()
	}

	// Policy validation moved to workflow (ValidatePolicyActivity is the first activity)
	// This ensures all validation runs within the Temporal workflow context

	// Generate ticket ID
	ticketID, err := h.revivalRepo.GenerateTicketID(sctx.Ctx)
	if err != nil {
		log.Error(nil, "failed to generate ticket ID", err)
		errMsg := apierrors.HandleErrorWithStatusCodeAndMessage(
			apierrors.CustomError,
			"Unable to generate ticket ID",
			err,
		)
		return nil, errMsg
	}

	// Serialize documents to JSON string for storage
	// Note: MissingDocumentsList is NOT stored at indexing - only sent by data entry/QC/approver
	var docsJSON string
	if len(req.Documents) > 0 {
		docsBytes, err := json.Marshal(req.Documents)
		if err != nil {
			log.Error(nil, "failed to marshal documents", err)
			errMsg := apierrors.HandleErrorWithStatusCodeAndMessage(
				apierrors.CustomError,
				"Unable to process documents",
				err,
			)
			return nil, errMsg
		}
		docsJSON = string(docsBytes)
	} else {
		docsJSON = "[]" // Empty array
	}

	// Prepare workflow input
	// Note: MaturityDate is fetched by ValidatePolicyActivity (first activity) using batch query
	// This reduces round trips by combining policy fetch + validation in single DB call
	workflowInput := workflow.IndexRevivalInput{
		TicketID:     ticketID,
		PolicyNumber: req.PolicyNumber,
		RequestType:  "installment_revival",
		IndexedBy:    req.IndexedBy,
		IndexedDate:  requestDateTime,
		Documents:    docsJSON,
	}

	// Start Temporal workflow with EAGER EXECUTION
	// Workflow creates DB record as its first activity
	workflowOptions := tclient.StartWorkflowOptions{
		ID:               fmt.Sprintf("revival-workflow-%s", ticketID),
		TaskQueue:        h.taskQueue,
		EnableEagerStart: true, // 🚀 Eager execution for fast response (~20-50ms)
	}

	workflowRun, err := h.temporalClient.ExecuteWorkflow(
		sctx.Ctx,
		workflowOptions,
		workflow.InstallmentRevivalWorkflow,
		workflowInput,
	)
	if err != nil {
		log.Error(nil, "failed to start workflow", err)
		errMsg := apierrors.HandleErrorWithStatusCodeAndMessage(
			apierrors.CustomError,
			"Unable to start revival workflow",
			err,
		)
		return nil, errMsg
	}

	log.Info(nil, "revival workflow started successfully",
		"ticket_id", ticketID,
		"workflow_id", workflowRun.GetID())

	// Return immediately - workflow is already running via eager execution
	// DB record will be created by the workflow's first activity
	return &port.IndexRequestResponse{
		StatusCodeAndMessage: port.CreateSuccess,
		Data: port.IndexRequestData{
			TicketID:        ticketID,
			WorkflowID:      stringPtr(workflowRun.GetID()),
			Status:          "INDEXED",
			RequestDateTime: requestDateTime,
			Message:         "Revival workflow started successfully",
		},
	}, nil
}

// SubmitDataEntry submits data entry for indexed request and signals workflow
func (h *RevivalHandler) SubmitDataEntry(sctx *serverRoute.Context, req port.DataEntryRequest) (*port.DataEntryResponse, error) {
	log.Info(sctx.Ctx, "SubmitDataEntry handler started", "ticket_id", req.TicketID)

	// Validate revival details (including lumpsum/installments constraint)
	if err := port.ValidateRevivalDetails(&req.RevivalDetails); err != nil {
		log.Error(sctx.Ctx, "Revival details validation failed", "ticket_id", req.TicketID, "error", err)
		errMsg := apierrors.HandleErrorWithStatusCodeAndMessage(
			apierrors.CustomError,
			err.Error(),
			err,
		)
		return nil, errMsg
	}

	// Check if revival request exists
	// TODO:check ticketid at temporal side itself
	revivalReq, err := h.revivalRepo.GetRevivalRequestByTicketID(sctx.Ctx, req.TicketID)
	if err != nil {
		log.Error(sctx.Ctx, "Failed to get revival request", "ticket_id", req.TicketID, "error", err)
		errMsg := apierrors.HandleErrorWithStatusCodeAndMessage(
			apierrors.DBErrorRecordNotFound,
			"No revival request found for the given ticket ID",
			err,
		)
		return nil, errMsg
	}
	log.Info(sctx.Ctx, "Revival request found", "ticket_id", req.TicketID, "status", revivalReq.CurrentStatus)

	// Check if request is in INDEXED or DATA_ENTRY_PENDING status (for rework after QC failure)
	if revivalReq.CurrentStatus != "INDEXED" && revivalReq.CurrentStatus != "DATA_ENTRY_PENDING" {
		return &port.DataEntryResponse{
			StatusCodeAndMessage: port.InvalidTicketStatus,
		}, nil
	}

	// Set data entry timestamp if not provided
	dataEntryTimestamp := req.DataEntryTimestamp
	if dataEntryTimestamp.IsZero() {
		dataEntryTimestamp = time.Now()
	}

	// Serialize documents submitted to JSON for storage
	var documentsJSON string
	if len(req.DocumentsSubmitted) > 0 {
		documentsBytes, err := json.Marshal(req.DocumentsSubmitted)
		if err != nil {
			log.Error(sctx.Ctx, "Failed to marshal documents submitted", "ticket_id", req.TicketID, "error", err)
			errMsg := apierrors.HandleErrorWithStatusCodeAndMessage(
				apierrors.CustomError,
				"Unable to process submitted documents",
				err,
			)
			return nil, errMsg
		}
		documentsJSON = string(documentsBytes)
		log.Info(sctx.Ctx, "Serialized documents submitted", "ticket_id", req.TicketID, "count", len(req.DocumentsSubmitted), "json", documentsJSON)
	} else {
		documentsJSON = "[]"
		log.Info(sctx.Ctx, "No documents submitted", "ticket_id", req.TicketID)
	}

	// Serialize missing documents to JSON for storage
	var missingDocsJSON string
	hasMissingDocuments := false
	if len(req.MissingDocuments) > 0 {
		missingDocsBytes, err := json.Marshal(req.MissingDocuments)
		if err != nil {
			log.Error(sctx.Ctx, "Failed to marshal missing documents", "ticket_id", req.TicketID, "error", err)
			errMsg := apierrors.HandleErrorWithStatusCodeAndMessage(
				apierrors.CustomError,
				"Unable to process missing documents",
				err,
			)
			return nil, errMsg
		}
		missingDocsJSON = string(missingDocsBytes)
		hasMissingDocuments = true
		log.Info(sctx.Ctx, "Missing documents identified", "ticket_id", req.TicketID, "count", len(req.MissingDocuments))
	} else {
		missingDocsJSON = "[]"
	}

	// 🔍 CHECK IF DOCUMENTS ARE MISSING
	// If documents are missing, save data but DO NOT proceed to QC and keep status unchanged
	if hasMissingDocuments {
		log.Info(sctx.Ctx, "Data entry has missing documents - saving data but not proceeding to QC, status remains unchanged", "ticket_id", req.TicketID, "current_status", revivalReq.CurrentStatus)

		// Update DB but keep current status unchanged
		err = h.revivalRepo.UpdateRevivalRequestForDataEntryWithPendingDocs(
			sctx.Ctx,
			revivalReq.RequestID,
			revivalReq.CurrentStatus, // Keep status unchanged
			req.DataEnteredBy,
			req.RevivalDetails.RevivalType,
			req.RevivalDetails.Installments,
			req.RevivalDetails.RevivalAmount,
			req.RevivalDetails.InstallmentAmount,
			req.RevivalDetails.TaxBreakdown.SGST,
			req.RevivalDetails.TaxBreakdown.CGST,
			req.RevivalDetails.Interest,
			documentsJSON,
			missingDocsJSON,
			req.MedicalExaminerCode,
			req.MedicalExaminerName,
		)
		if err != nil {
			log.Error(sctx.Ctx, "Failed to update database for data entry with pending documents", "ticket_id", req.TicketID, "error", err)
			errMsg := apierrors.HandleErrorWithStatusCodeAndMessage(
				apierrors.CustomError,
				"Unable to save data entry",
				err,
			)
			return nil, errMsg
		}

		log.Info(sctx.Ctx, "Data saved successfully, status unchanged", "ticket_id", req.TicketID, "status", revivalReq.CurrentStatus)

		return &port.DataEntryResponse{
			StatusCodeAndMessage: port.UpdateSuccess,
			Data: port.DataEntryData{
				TicketID:           req.TicketID,
				Status:             revivalReq.CurrentStatus, // Return current status unchanged
				Message:            fmt.Sprintf("Data entry saved successfully. Awaiting %d missing document(s). Status remains %s. Workflow will proceed to QC once all documents are received.", len(req.MissingDocuments), revivalReq.CurrentStatus),
				DataEntryTimestamp: dataEntryTimestamp,
			},
		}, nil
	}

	// 🚀 NO MISSING DOCUMENTS: Proceed with normal flow
	// SIGNAL-FIRST PATTERN: Send signal to workflow with all data
	// Workflow will update DB via UpdateDataEntryActivity
	// This prevents orphaned DB updates if signal fails
	if revivalReq.WorkflowID == nil || revivalReq.RunID == nil {
		errMsg := apierrors.HandleErrorWithStatusCodeAndMessage(
			apierrors.CustomError,
			"Workflow identifiers are missing for this ticket",
			nil,
		)
		return nil, errMsg
	}

	err = h.temporalClient.SignalWorkflow(
		sctx.Ctx,
		*revivalReq.WorkflowID,
		*revivalReq.RunID,
		"data-entry-complete",
		workflow.DataEntryCompleteSignal{
			EnteredBy:            req.DataEnteredBy,
			EnteredAt:            dataEntryTimestamp,
			RevivalType:          req.RevivalDetails.RevivalType,
			NumberOfInstallments: req.RevivalDetails.Installments,
			RevivalAmount:        req.RevivalDetails.RevivalAmount,
			InstallmentAmount:    req.RevivalDetails.InstallmentAmount,
			MissingDocuments:     missingDocsJSON,
			Documents:            documentsJSON,
			Interest:             req.RevivalDetails.Interest,
			SGST:                 req.RevivalDetails.TaxBreakdown.SGST,
			CGST:                 req.RevivalDetails.TaxBreakdown.CGST,
			MedicalExaminerCode:  req.MedicalExaminerCode,
			MedicalExaminerName:  req.MedicalExaminerName,
		},
	)
	if err != nil {
		log.Error(sctx.Ctx, "Failed to signal workflow for data entry", "ticket_id", req.TicketID, "error", err)
		errMsg := apierrors.HandleErrorWithStatusCodeAndMessage(
			apierrors.CustomError,
			"Unable to update workflow for data entry",
			err,
		)
		return nil, errMsg
	}

	status := "DATA_ENTRY_COMPLETE"
	message := "Data entry submitted successfully"
	stateDetails, stateErr := h.getWorkflowStateDetailsWithRetry(sctx.Ctx, *revivalReq.WorkflowID, *revivalReq.RunID)
	if stateErr != nil {
		log.Warn(sctx.Ctx, "Unable to fetch workflow state details after data entry signal", "ticket_id", req.TicketID, "error", stateErr)
	} else if stateDetails != nil {
		status = normalizeWorkflowStatus(stateDetails.CurrentStatus, status)
		if stateDetails.RecoverableError && stateDetails.LastErrorMessage != "" {
			message = stateDetails.LastErrorMessage
		} else {
			message = "Data entry accepted and processing in workflow"
		}
	}

	log.Info(sctx.Ctx, "SubmitDataEntry returning success response", "ticket_id", req.TicketID)
	response := &port.DataEntryResponse{
		StatusCodeAndMessage: port.UpdateSuccess,
		Data: port.DataEntryData{
			TicketID:           req.TicketID,
			Status:             status,
			Message:            message,
			DataEntryTimestamp: dataEntryTimestamp,
		},
	}
	log.Info(sctx.Ctx, "SubmitDataEntry response created", "response_status", response.StatusCode, "response_success", response.Success)
	return response, nil
}

// SubmitQualityCheck submits quality check for data entry and signals workflow
func (h *RevivalHandler) SubmitQualityCheck(sctx *serverRoute.Context, req port.QualityCheckRequest) (*port.QualityCheckResponse, error) {
	log.Info(sctx.Ctx, "SubmitQualityCheck handler started", "ticket_id", req.TicketID, "qc_passed", req.QCPassed)

	// Check if revival request exists
	revivalReq, err := h.revivalRepo.GetRevivalRequestByTicketID(sctx.Ctx, req.TicketID)
	if err != nil {
		log.Error(sctx.Ctx, "Failed to get revival request for QC", "ticket_id", req.TicketID, "error", err)
		errMsg := apierrors.HandleErrorWithStatusCodeAndMessage(
			apierrors.DBErrorRecordNotFound,
			"No revival request found for the given ticket ID",
			err,
		)
		return nil, errMsg
	}
	log.Info(sctx.Ctx, "Revival request found for QC", "ticket_id", req.TicketID, "status", revivalReq.CurrentStatus)

	// Check if request is in DATA_ENTRY_COMPLETE status
	if revivalReq.CurrentStatus != "DATA_ENTRY_COMPLETE" {
		return &port.QualityCheckResponse{
			StatusCodeAndMessage: port.TicketNotReadyForQC,
		}, nil
	}

	// Set QC timestamp if not provided
	qcTimestamp := req.QCTimestamp
	if qcTimestamp.IsZero() {
		qcTimestamp = time.Now()
	}

	// Serialize missing documents to JSON for storage
	var missingDocsJSON string
	if len(req.MissingDocuments) > 0 {
		missingDocsBytes, err := json.Marshal(req.MissingDocuments)
		if err != nil {
			log.Error(sctx.Ctx, "Failed to marshal missing documents", "ticket_id", req.TicketID, "error", err)
			errMsg := apierrors.HandleErrorWithStatusCodeAndMessage(
				apierrors.CustomError,
				"Unable to process missing documents",
				err,
			)
			return nil, errMsg
		}
		missingDocsJSON = string(missingDocsBytes)
	} else {
		missingDocsJSON = "[]"
	}

	// 🚀 SIGNAL-FIRST PATTERN: Send signal to workflow with all data
	// Workflow will update DB via UpdateQCActivity
	// This prevents orphaned DB updates if signal fails
	nextStatus := "APPROVAL_PENDING"
	if !req.QCPassed {
		nextStatus = "DATA_ENTRY_PENDING"
	}

	if revivalReq.WorkflowID != nil && revivalReq.RunID != nil {
		err = h.temporalClient.SignalWorkflow(
			sctx.Ctx,
			*revivalReq.WorkflowID,
			*revivalReq.RunID,
			"quality-check-complete",
			workflow.QualityCheckCompleteSignal{
				QCPassed:         req.QCPassed,
				QCComments:       req.QCComments,
				PerformedBy:      req.QCPerformedBy,
				PerformedAt:      qcTimestamp,
				MissingDocuments: missingDocsJSON,
			},
		)
		if err != nil {
			errMsg := apierrors.HandleErrorWithStatusCodeAndMessage(
				apierrors.CustomError,
				"Unable to update workflow for quality check",
				err,
			)
			return nil, errMsg
		}

		stateDetails, stateErr := h.getWorkflowStateDetailsWithRetry(sctx.Ctx, *revivalReq.WorkflowID, *revivalReq.RunID)
		if stateErr == nil && stateDetails != nil {
			nextStatus = normalizeWorkflowStatus(stateDetails.CurrentStatus, nextStatus)
			if stateDetails.RecoverableError && stateDetails.LastErrorMessage != "" {
				return &port.QualityCheckResponse{
					StatusCodeAndMessage: port.UpdateSuccess,
					Data: port.QualityCheckData{
						TicketID:    req.TicketID,
						Status:      nextStatus,
						Message:     stateDetails.LastErrorMessage,
						QCTimestamp: qcTimestamp,
					},
				}, nil
			}
		} else if stateErr != nil {
			log.Warn(sctx.Ctx, "Unable to fetch workflow state details after QC signal", "ticket_id", req.TicketID, "error", stateErr)
		}
	}

	log.Info(sctx.Ctx, "SubmitQualityCheck returning success response", "ticket_id", req.TicketID, "next_status", nextStatus)
	response := &port.QualityCheckResponse{
		StatusCodeAndMessage: port.UpdateSuccess,
		Data: port.QualityCheckData{
			TicketID:    req.TicketID,
			Status:      nextStatus,
			Message:     "Quality check submitted and workflow signaled successfully",
			QCTimestamp: qcTimestamp,
		},
	}
	log.Info(sctx.Ctx, "SubmitQualityCheck response created", "response_status", response.StatusCode, "response_success", response.Success)
	return response, nil
}

// ApprovalDecision handles both approval and rejection in a single endpoint
func (h *RevivalHandler) ApprovalDecision(sctx *serverRoute.Context, req port.ApprovalDecisionRequest) (*port.ApprovalDecisionResponse, error) {
	// Check if revival request exists
	revivalReq, err := h.revivalRepo.GetRevivalRequestByTicketID(sctx.Ctx, req.TicketID)
	if err != nil {
		errMsg := apierrors.HandleErrorWithStatusCodeAndMessage(
			apierrors.DBErrorRecordNotFound,
			"No revival request found for the given ticket ID",
			err,
		)
		return nil, errMsg
	}

	// Check if request is in APPROVAL_PENDING status
	if revivalReq.CurrentStatus != "APPROVAL_PENDING" {
		return &port.ApprovalDecisionResponse{
			StatusCodeAndMessage: port.InvalidApprovalStatus,
		}, nil
	}

	// Validate documents (read-only check at approval stage)
	if err := validateDocuments(revivalReq.Documents, revivalReq.MissingDocumentsList); err != nil {
		log.Warn(nil, "Document validation failed at approval", "ticket_id", req.TicketID, "error", err)
		// Return validation error to approver
		return &port.ApprovalDecisionResponse{
			StatusCodeAndMessage: port.InvalidApprovalStatus,
			Data: port.ApprovalDecisionData{
				TicketID:  req.TicketID,
				Approved:  false,
				Status:    "APPROVAL_PENDING",
				Message:   fmt.Sprintf("Document validation failed: %v", err),
				Timestamp: time.Now(),
			},
		}, nil
	}

	// Set timestamp if not provided
	timestamp := req.Timestamp
	if timestamp.IsZero() {
		timestamp = time.Now()
	}

	if revivalReq.WorkflowID == nil || revivalReq.RunID == nil {
		errMsg := apierrors.HandleErrorWithStatusCodeAndMessage(
			apierrors.CustomError,
			"Workflow identifiers are missing for this ticket",
			nil,
		)
		return nil, errMsg
	}

	// Handle approval/rejection - 🚀 SIGNAL-FIRST PATTERN
	// Send signal to workflow with all data
	// Workflow will update DB state via activities
	// This prevents orphaned DB updates if signal fails
	err = h.temporalClient.SignalWorkflow(
		sctx.Ctx,
		*revivalReq.WorkflowID,
		*revivalReq.RunID,
		"approval-decision",
		workflow.ApprovalDecisionSignal{
			Approved:   req.Approved,
			Comments:   req.Comments,
			ApprovedBy: req.PerformedBy,
			ApprovedAt: timestamp,
		},
	)
	if err != nil {
		errMsg := apierrors.HandleErrorWithStatusCodeAndMessage(
			apierrors.CustomError,
			"Unable to update workflow for approval decision",
			err,
		)
		return nil, errMsg
	}

	status := "APPROVAL_PENDING"
	message := "Approval decision submitted successfully"
	recoverableError := false
	stateDetails, stateErr := h.getWorkflowStateDetailsWithRetry(sctx.Ctx, *revivalReq.WorkflowID, *revivalReq.RunID)
	if stateErr != nil {
		log.Warn(sctx.Ctx, "Unable to fetch workflow state details after approval signal", "ticket_id", req.TicketID, "error", stateErr)
	} else if stateDetails != nil {
		status = normalizeWorkflowStatus(stateDetails.CurrentStatus, status)
		recoverableError = stateDetails.RecoverableError
		if stateDetails.RecoverableError && stateDetails.LastErrorMessage != "" {
			message = stateDetails.LastErrorMessage
		} else if req.Approved {
			message = "Revival request approved successfully"
		} else {
			message = "Revival request rejected successfully"
		}
	}

	if !req.Approved {
		letterID := ""
		if status == "REJECTED" && !recoverableError {
			letterID, err = h.activities.GenerateLetterActivity(sctx.Ctx, "REJECTION", revivalReq.RequestID, revivalReq.PolicyNumber)
			if err != nil {
				log.Error(sctx.Ctx, "Failed to generate rejection letter", "ticket_id", req.TicketID, "error", err)
				letterID = ""
			}
		}

		if letterID != "" {
			message = fmt.Sprintf("%s Rejection letter %s generated", message, letterID)
		}

		return &port.ApprovalDecisionResponse{
			StatusCodeAndMessage: port.UpdateSuccess,
			Data: port.ApprovalDecisionData{
				TicketID:  req.TicketID,
				Approved:  false,
				Status:    status,
				Message:   message,
				Timestamp: timestamp,
			},
		}, nil
	}

	// Calculate SLA dates for response
	slaEndDate := timestamp.AddDate(0, 0, 60) // 60-day SLA (IR_11)
	slaRemainingDays := 60

	return &port.ApprovalDecisionResponse{
		StatusCodeAndMessage: port.UpdateSuccess,
		Data: port.ApprovalDecisionData{
			TicketID:         req.TicketID,
			Approved:         !recoverableError,
			Status:           status,
			Message:          message,
			Timestamp:        timestamp,
			SLAEndDate:       &slaEndDate,
			SLARemainingDays: &slaRemainingDays,
		},
	}, nil
}

// ApprovalRedirect allows approver to redirect request to earlier stage for rework
func (h *RevivalHandler) ApprovalRedirect(sctx *serverRoute.Context, req port.ApprovalRedirectRequest) (*port.ApprovalRedirectResponse, error) {
	// Check if revival request exists
	revivalReq, err := h.revivalRepo.GetRevivalRequestByTicketID(sctx.Ctx, req.TicketID)
	if err != nil {
		errMsg := apierrors.HandleErrorWithStatusCodeAndMessage(
			apierrors.DBErrorRecordNotFound,
			"No revival request found for the given ticket ID",
			err,
		)
		return nil, errMsg
	}

	// Check if request is in APPROVAL_PENDING status
	if revivalReq.CurrentStatus != "APPROVAL_PENDING" {
		return &port.ApprovalRedirectResponse{
			StatusCodeAndMessage: port.InvalidApprovalStatus,
			Data: port.ApprovalRedirectData{
				TicketID: req.TicketID,
				Status:   revivalReq.CurrentStatus,
				Message:  "Request must be in APPROVAL_PENDING status to redirect",
			},
		}, nil
	}

	// Set timestamp if not provided
	timestamp := req.Timestamp
	if timestamp.IsZero() {
		timestamp = time.Now()
	}

	// Determine target status based on redirect destination
	var targetStatus string
	var message string
	if req.RedirectTo == "DATA_ENTRY" {
		targetStatus = "DATA_ENTRY_PENDING"
		message = "Request redirected to Data Entry for rework"
	} else if req.RedirectTo == "QC" {
		targetStatus = "DATA_ENTRY_COMPLETE"
		message = "Request redirected to Quality Check for re-verification"
	} else {
		return &port.ApprovalRedirectResponse{
			StatusCodeAndMessage: port.InvalidApprovalStatus,
			Data: port.ApprovalRedirectData{
				TicketID: req.TicketID,
				Message:  "Invalid redirect destination. Must be DATA_ENTRY or QC",
			},
		}, nil
	}

	// 🚀 SIGNAL-FIRST PATTERN: Send signal to workflow to update state
	// Workflow will persist redirected status and update state
	if revivalReq.WorkflowID == nil || revivalReq.RunID == nil {
		errMsg := apierrors.HandleErrorWithStatusCodeAndMessage(
			apierrors.CustomError,
			"Workflow identifiers are missing for this ticket",
			nil,
		)
		return nil, errMsg
	}

	err = h.temporalClient.SignalWorkflow(
		sctx.Ctx,
		*revivalReq.WorkflowID,
		*revivalReq.RunID,
		"approval-decision",
		workflow.ApprovalDecisionSignal{
			Approved:         false, // Not approved, redirecting
			Comments:         req.Comments,
			ApprovedBy:       req.PerformedBy,
			ApprovedAt:       timestamp,
			RedirectToStage:  req.RedirectTo,
			RedirectComments: req.Comments,
		},
	)
	if err != nil {
		errMsg := apierrors.HandleErrorWithStatusCodeAndMessage(
			apierrors.CustomError,
			"Unable to update workflow for redirect",
			err,
		)
		return nil, errMsg
	}

	stateDetails, stateErr := h.getWorkflowStateDetailsWithRetry(sctx.Ctx, *revivalReq.WorkflowID, *revivalReq.RunID)
	if stateErr == nil && stateDetails != nil {
		targetStatus = normalizeWorkflowStatus(stateDetails.CurrentStatus, targetStatus)
		if stateDetails.RecoverableError && stateDetails.LastErrorMessage != "" {
			message = stateDetails.LastErrorMessage
		}
	} else if stateErr != nil {
		log.Warn(sctx.Ctx, "Unable to fetch workflow state details after redirect signal", "ticket_id", req.TicketID, "error", stateErr)
	}

	return &port.ApprovalRedirectResponse{
		StatusCodeAndMessage: port.UpdateSuccess,
		Data: port.ApprovalRedirectData{
			TicketID:     req.TicketID,
			RedirectedTo: targetStatus,
			Status:       targetStatus,
			Message:      message,
			Timestamp:    timestamp,
		},
	}, nil
}

func (h *RevivalHandler) getWorkflowStateDetailsWithRetry(ctx context.Context, workflowID, runID string) (*workflow.RevivalWorkflowState, error) {
	const maxAttempts = 3
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		queryResult, err := h.temporalClient.QueryWorkflow(ctx, workflowID, runID, "getStateDetails")
		if err != nil {
			if attempt == maxAttempts {
				return nil, err
			}
			time.Sleep(120 * time.Millisecond)
			continue
		}

		var state workflow.RevivalWorkflowState
		if err := queryResult.Get(&state); err != nil {
			if attempt == maxAttempts {
				return nil, err
			}
			time.Sleep(120 * time.Millisecond)
			continue
		}

		return &state, nil
	}

	return nil, fmt.Errorf("unable to query workflow state details")
}

func normalizeWorkflowStatus(workflowStatus, fallback string) string {
	switch workflowStatus {
	case "WAITING_FOR_DATA_ENTRY":
		return "DATA_ENTRY_PENDING"
	case "WAITING_FOR_QC":
		return "DATA_ENTRY_COMPLETE"
	case "PROCESSING_DATA_ENTRY":
		return "DATA_ENTRY_PENDING"
	case "PROCESSING_QC", "APPROVAL_PENDING", "PROCESSING_APPROVAL":
		return "APPROVAL_PENDING"
	case "":
		return fallback
	default:
		return workflowStatus
	}
}

// WithdrawRequest withdraws revival request
func (h *RevivalHandler) WithdrawRequest(sctx *serverRoute.Context, req port.WithdrawalRequest) (*port.WithdrawalResponse, error) {
	// Check if revival request exists
	revivalReq, err := h.revivalRepo.GetRevivalRequestByTicketID(sctx.Ctx, req.TicketID)
	if err != nil {
		errMsg := apierrors.HandleErrorWithStatusCodeAndMessage(
			apierrors.DBErrorRecordNotFound,
			"No revival request found for the given ticket ID",
			err,
		)
		return nil, errMsg
	}

	// IR_37: Withdrawal only allowed before first collection
	if revivalReq.FirstCollectionDone {
		return &port.WithdrawalResponse{
			StatusCodeAndMessage: port.InvalidApprovalStatus,
			Data: port.WithdrawalData{
				Message: "Withdrawal not allowed after first collection (IR_37)",
			},
		}, nil
	}

	// Check for suspense entries that need adjustment
	suspenseEntries, err := h.paymentRepo.GetSuspenseByPolicyNumber(sctx.Ctx, revivalReq.PolicyNumber)
	if err != nil {
		suspenseEntries = []domain.SuspenseAccount{}
	}

	suspenseAdjustments := make([]port.SuspenseAdjustment, 0)
	for _, suspense := range suspenseEntries {
		if suspense.RequestID != nil && *suspense.RequestID == revivalReq.RequestID && !suspense.IsReversed {
			suspenseAdjustments = append(suspenseAdjustments, port.SuspenseAdjustment{
				SuspenseID: suspense.SuspenseID,
				Amount:     suspense.Amount,
			})
		}
	}

	// Set withdrawal date if not provided
	withdrawalDate := req.WithdrawalDate
	if withdrawalDate.IsZero() {
		withdrawalDate = time.Now()
	}

	// Update revival request status to WITHDRAWN
	err = h.revivalRepo.UpdateRevivalRequestForWithdrawal(sctx.Ctx, req.TicketID, req.WithdrawalReason)
	if err != nil {
		errMsg := apierrors.HandleErrorWithStatusCodeAndMessage(
			apierrors.CustomError,
			"Unable to withdraw revival request",
			err,
		)
		return nil, errMsg
	}

	message := "Revival request withdrawn successfully"
	if len(suspenseAdjustments) > 0 {
		message = fmt.Sprintf("Revival request withdrawn. %d suspense adjustments pending", len(suspenseAdjustments))
	}

	return &port.WithdrawalResponse{
		StatusCodeAndMessage: port.UpdateSuccess,
		Data: port.WithdrawalData{
			TicketID:            req.TicketID,
			PolicyNumber:        revivalReq.PolicyNumber,
			Status:              "WITHDRAWN",
			WithdrawalDate:      withdrawalDate,
			WithdrawalReason:    req.WithdrawalReason,
			SuspenseAdjustments: suspenseAdjustments,
			Message:             message,
		},
	}, nil
}

// FirstCollection processes first installment collection (dual collection IR_36)
func (h *RevivalHandler) FirstCollection(sctx *serverRoute.Context, req port.FirstCollectionRequest) (*port.FirstCollectionResponse, error) {
	// Set collection date if not provided
	collectionDate := req.CollectionDate
	if collectionDate.IsZero() {
		collectionDate = time.Now()
	}

	// Validate dual collection using activity
	input := workflow.FirstCollectionInput{
		PolicyNumber:      req.PolicyNumber,
		PremiumAmount:     req.PremiumAmount,
		InstallmentAmount: req.InstallmentAmount,
		PaymentMode:       req.PaymentMode,
		TotalAmount:       req.TotalAmount,
		ChequeNumber:      nil,
		BankName:          nil,
		ChequeDate:        nil,
		ChequeAmount:      nil,
		PremiumSGST:       req.PremiumSGST,
		PremiumCGST:       req.PremiumCGST,
		InstallmentSGST:   req.InstallmentSGST,
		InstallmentCGST:   req.InstallmentCGST,
	}

	if req.ChequeDetails != nil {
		input.ChequeNumber = &req.ChequeDetails.ChequeNumber
		input.BankName = &req.ChequeDetails.BankName
		input.ChequeDate = &req.ChequeDetails.ChequeDate
		input.ChequeAmount = &req.ChequeDetails.Amount
	}

	// Get revival request to get request ID
	revivalReq, err := h.revivalRepo.GetRevivalRequestByTicketID(sctx.Ctx, req.TicketID)
	if err != nil {
		log.Error(nil, err)
		return nil, fmt.Errorf("revival request not found: %w", err)
	}

	input.RequestID = revivalReq.RequestID

	// Validate dual collection
	err = h.activities.ValidateDualCollectionActivity(sctx.Ctx, input)
	if err != nil {
		log.Error(nil, err)
		return &port.FirstCollectionResponse{
			StatusCodeAndMessage: port.MissingRequiredFields,
		}, nil
	}

	// Process dual payment
	err = h.activities.ProcessDualPaymentActivity(sctx.Ctx, input, "PAID")
	if err != nil {
		log.Error(nil, err)
		return nil, fmt.Errorf("failed to process dual payment: %w", err)
	}

	// Handle cheque record if needed
	if req.PaymentMode == "CHEQUE" {
		_, err = h.activities.CreateChequeRecordActivity(sctx.Ctx, input)
		if err != nil {
			log.Error(nil, err)
			return nil, fmt.Errorf("failed to create cheque record: %w", err)
		}
	}

	// Send signal to workflow to complete first collection
	if revivalReq.WorkflowID != nil && revivalReq.RunID != nil {
		log.Info(nil, "Sending first-collection-complete signal to revival workflow",
			"workflow_id", *revivalReq.WorkflowID,
			"run_id", *revivalReq.RunID,
			"ticket_id", req.TicketID)

		err = h.temporalClient.SignalWorkflow(
			sctx.Ctx,
			*revivalReq.WorkflowID,
			*revivalReq.RunID,
			"first-collection-complete",
			workflow.FirstCollectionCompleteSignal{
				CollectionDate: collectionDate,
				PaymentMode:    req.PaymentMode,
				TotalAmount:    req.TotalAmount,
			},
		)
		if err != nil {
			log.Error(nil, "Failed to signal revival workflow for first collection",
				"error", err,
				"workflow_id", *revivalReq.WorkflowID,
				"run_id", *revivalReq.RunID)

			// Signal delivery can fail even when workflow has already progressed.
			stateDetails, stateErr := h.getWorkflowStateDetailsWithRetry(sctx.Ctx, *revivalReq.WorkflowID, *revivalReq.RunID)
			if stateErr == nil && stateDetails != nil {
				currentStatus := normalizeWorkflowStatus(stateDetails.CurrentStatus, "APPROVED")
				if currentStatus == "ACTIVE" || currentStatus == "COMPLETED" {
					log.Warn(sctx.Ctx, "First collection signal returned error but workflow is already advanced", "ticket_id", req.TicketID, "workflow_status", currentStatus)
				} else {
					return nil, fmt.Errorf("first collection payment saved but workflow not synced (status=%s): %w", currentStatus, err)
				}
			} else {
				return nil, fmt.Errorf("first collection payment saved but workflow sync could not be verified: %w", err)
			}
		}

		log.Info(nil, "Successfully signaled revival workflow for first collection",
			"workflow_id", *revivalReq.WorkflowID,
			"ticket_id", req.TicketID)
	} else {
		log.Warn(nil, "Cannot signal workflow - WorkflowID or RunID is nil",
			"ticket_id", req.TicketID,
			"has_workflow_id", revivalReq.WorkflowID != nil,
			"has_run_id", revivalReq.RunID != nil)
	}

	// Generate receipt number
	receiptNumber := "RCPT" + collectionDate.Format("20060102150405")

	// Calculate next due date (IR_11: due dates on 1st of each month)
	nextMonth := time.Now().AddDate(0, 1, 0)
	nextDueDate := time.Date(nextMonth.Year(), nextMonth.Month(), 1, 0, 0, 0, time.Now().Nanosecond(), time.Now().Location())

	// Calculate total GST
	totalGST := req.PremiumSGST + req.PremiumCGST + req.InstallmentSGST + req.InstallmentCGST

	collectionStatus := "ACTIVE"
	if revivalReq.WorkflowID != nil && revivalReq.RunID != nil {
		stateDetails, stateErr := h.getWorkflowStateDetailsWithRetry(sctx.Ctx, *revivalReq.WorkflowID, *revivalReq.RunID)
		if stateErr == nil && stateDetails != nil {
			collectionStatus = normalizeWorkflowStatus(stateDetails.CurrentStatus, collectionStatus)
		}
	}

	return &port.FirstCollectionResponse{
		StatusCodeAndMessage: port.UpdateSuccess,
		Data: port.FirstCollectionData{
			ReceiptNumber:     receiptNumber,
			TicketID:          req.TicketID,
			PolicyNumber:      req.PolicyNumber,
			PremiumAmount:     req.PremiumAmount,
			InstallmentAmount: req.InstallmentAmount,
			PremiumSGST:       req.PremiumSGST,
			PremiumCGST:       req.PremiumCGST,
			InstallmentSGST:   req.InstallmentSGST,
			InstallmentCGST:   req.InstallmentCGST,
			TotalGST:          totalGST,
			TotalAmount:       req.TotalAmount,
			ReceiptDate:       collectionDate,
			PaymentMode:       req.PaymentMode,
			Status:            collectionStatus,
			DueDate:           &nextDueDate,
		},
	}, nil
}

// CreateInstallment creates subsequent installment
func (h *RevivalHandler) CreateInstallment(sctx *serverRoute.Context, req port.CreateInstallmentRequest) (*port.InstallmentResponse, error) {
	// Check if revival request exists
	revivalReq, err := h.revivalRepo.GetRevivalRequestByTicketID(sctx.Ctx, req.TicketID)
	if err != nil {
		log.Error(nil, err)
		return nil, fmt.Errorf("revival request not found: %w", err)
	}

	log.Debug(nil, " this after checking revival request")

	expectedAmount := revivalReq.InstallmentAmount

	if req.InstallmentAmount != expectedAmount {
		log.Error(nil, "Invalid installment amount",
			"expected", expectedAmount,
			"received", req.InstallmentAmount,
			"installment_number", req.InstallmentNumber,
		)

		return &port.InstallmentResponse{
			StatusCodeAndMessage: port.InvalidInstallmentAmount,
			Data: port.InstallmentData{
				TicketID:          req.TicketID,
				PolicyNumber:      req.PolicyNumber,
				InstallmentNumber: req.InstallmentNumber,
				Status:            "REJECTED",
			},
		}, nil
	}

	// Check if request is in ACTIVE status
	if revivalReq.CurrentStatus != "ACTIVE" {
		log.Error(nil, "error:", port.PolicyNotActiveStatus)
		return &port.InstallmentResponse{
			StatusCodeAndMessage: port.PolicyNotActiveStatus,
		}, nil
	}
	log.Debug(nil, " this after checking active status")

	// Validate installment number against ACTUAL approved count (not hardcoded 12!)
	if req.InstallmentNumber < 2 {
		log.Error(nil, "Invalid installment number: must be >= 2 (first installment already collected)",
			"installment_number", req.InstallmentNumber)
		return &port.InstallmentResponse{
			StatusCodeAndMessage: port.InvalidApprovalStatus,
			Data: port.InstallmentData{
				TicketID: req.TicketID,
				Status:   "REJECTED",
			},
		}, nil
	}

	// Check against approved installment count
	if req.InstallmentNumber > revivalReq.NumberOfInstallments {
		log.Error(nil, "Installment number exceeds approved count",
			"installment_number", req.InstallmentNumber,
			"approved_installments", revivalReq.NumberOfInstallments)
		return &port.InstallmentResponse{
			StatusCodeAndMessage: port.MaxInstallmentExceeded,
			Data: port.InstallmentData{
				TicketID:          req.TicketID,
				PolicyNumber:      req.PolicyNumber,
				InstallmentNumber: req.InstallmentNumber,
				Status:            "REJECTED",
			},
		}, nil
	}

	// 🎯 SEQUENTIAL VALIDATION: Check if installment is next in sequence
	// installments_paid = number of installments already paid (1 = first collection done)
	// Next expected installment = installments_paid + 1
	nextExpectedInstallment := revivalReq.InstallmentsPaid + 1

	log.Info(nil, "Sequential installment validation",
		"installments_paid", revivalReq.InstallmentsPaid,
		"next_expected", nextExpectedInstallment,
		"requested_installment", req.InstallmentNumber)

	// Check if requested installment matches expected sequence
	if req.InstallmentNumber != nextExpectedInstallment {
		if req.InstallmentNumber < nextExpectedInstallment {
			// Duplicate payment attempt (already paid)
			log.Error(nil, "DUPLICATE PAYMENT: Installment already paid",
				"installment_number", req.InstallmentNumber,
				"next_expected", nextExpectedInstallment,
				"installments_paid", revivalReq.InstallmentsPaid)
			return &port.InstallmentResponse{
				StatusCodeAndMessage: port.InvalidApprovalStatus,
				Data: port.InstallmentData{
					TicketID:          req.TicketID,
					PolicyNumber:      req.PolicyNumber,
					InstallmentNumber: req.InstallmentNumber,
					Status:            "ALREADY_PAID",
				},
			}, nil
		} else {
			// Out-of-order payment attempt (skipping installments)
			log.Error(nil, "OUT-OF-ORDER PAYMENT: Must pay installments sequentially",
				"requested_installment", req.InstallmentNumber,
				"next_expected", nextExpectedInstallment,
				"installments_paid", revivalReq.InstallmentsPaid)
			return &port.InstallmentResponse{
				StatusCodeAndMessage: port.OutOfOrderPayment,
				Data: port.InstallmentData{
					TicketID:          req.TicketID,
					PolicyNumber:      req.PolicyNumber,
					InstallmentNumber: req.InstallmentNumber,
					Status:            "OUT_OF_ORDER",
				},
			}, nil
		}
	}

	// Send signal to InstallmentMonitorWorkflow (child workflow)
	// Signal name format: installment-payment-received-{installmentNumber}
	// The workflow's ProcessInstallmentActivity handles:
	// - Creating payment record in database
	// - Incrementing installments_paid
	// - Checking for completion (all installments paid -> COMPLETED status)

	// Construct child workflow ID using the known pattern: installment-monitor-{requestID}
	monitorWorkflowID := fmt.Sprintf("installment-monitor-%s", revivalReq.RequestID)
	signalName := fmt.Sprintf("installment-payment-received-%d", req.InstallmentNumber)

	log.Info(nil, "Signaling InstallmentMonitorWorkflow",
		"monitor_workflow_id", monitorWorkflowID,
		"signal_name", signalName,
		"installment_number", req.InstallmentNumber)

	// Retry logic for workflow signaling (child workflow might not be fully started yet)
	maxRetries := 3
	retryDelay := 2 * time.Second
	var signalErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		signalErr = h.temporalClient.SignalWorkflow(
			sctx.Ctx,
			monitorWorkflowID,
			"", // Empty run ID signals the latest/current run
			signalName,
			workflow.InstallmentPaymentSignal{
				PaymentDate: req.CollectionDate,
				Amount:      req.InstallmentAmount,
				PaymentMode: req.PaymentMode,
			},
		)

		if signalErr == nil {
			log.Info(nil, "Successfully signaled InstallmentMonitorWorkflow",
				"monitor_workflow_id", monitorWorkflowID,
				"installment_number", req.InstallmentNumber,
				"attempt", attempt)
			break
		}

		if attempt < maxRetries {
			log.Warn(nil, "Failed to signal InstallmentMonitorWorkflow, retrying",
				"monitor_workflow_id", monitorWorkflowID,
				"attempt", attempt,
				"max_retries", maxRetries,
				"error", signalErr)
			time.Sleep(retryDelay)
		}
	}

	if signalErr != nil {
		log.Error(nil, "Failed to signal InstallmentMonitorWorkflow after retries",
			"monitor_workflow_id", monitorWorkflowID,
			"error", signalErr,
			"attempts", maxRetries)
		return nil, fmt.Errorf("failed to signal InstallmentMonitorWorkflow: %w", signalErr)
	}

	// Generate tracking ID for response
	receiptNumber := fmt.Sprintf("INST%s", time.Now().Format("20060102150405"))

	return &port.InstallmentResponse{
		StatusCodeAndMessage: port.CreateSuccess,
		Data: port.InstallmentData{
			ReceiptNumber:     receiptNumber,
			TicketID:          req.TicketID,
			PolicyNumber:      req.PolicyNumber,
			InstallmentNumber: req.InstallmentNumber,
			InstallmentAmount: req.InstallmentAmount,
			Status:            "PROCESSING",
		},
	}, nil
}

// BatchFirstCollection processes multiple first collections
func (h *RevivalHandler) BatchFirstCollection(sctx *serverRoute.Context, req port.BatchFirstCollectionRequest) (*port.BatchFirstCollectionResponse, error) {
	totalSubmitted := len(req.Collections)
	successful := 0
	failed := 0
	results := make([]port.FirstCollectionResultItem, 0, totalSubmitted)

	for _, collection := range req.Collections {
		result := port.FirstCollectionResultItem{
			TicketID:     collection.TicketID,
			PolicyNumber: collection.PolicyNumber,
		}

		// Process single first collection
		response, err := h.FirstCollection(sctx, collection)
		if err != nil {
			// Error occurred
			result.Success = false
			result.Status = "ERROR"
			result.ErrorMessage = err.Error()
			failed++
		} else if response != nil && response.Success {
			// Success
			result.Success = true
			result.Status = response.Data.Status
			result.Message = "First collection processed successfully"
			result.ReceiptNumber = response.Data.ReceiptNumber
			result.PremiumAmount = response.Data.PremiumAmount
			result.InstallmentAmount = response.Data.InstallmentAmount
			result.PremiumSGST = response.Data.PremiumSGST
			result.PremiumCGST = response.Data.PremiumCGST
			result.InstallmentSGST = response.Data.InstallmentSGST
			result.InstallmentCGST = response.Data.InstallmentCGST
			result.TotalGST = response.Data.TotalGST
			result.TotalAmount = response.Data.TotalAmount
			result.ReceiptDate = response.Data.ReceiptDate
			result.PaymentMode = response.Data.PaymentMode
			result.DueDate = response.Data.DueDate
			successful++
		} else if response != nil {
			// Business logic failure
			result.Success = false
			result.Status = "FAILED"
			if response.Message != "" {
				result.ErrorMessage = response.Message
			} else {
				result.ErrorMessage = "First collection validation failed"
			}
			failed++
		} else {
			// Unexpected: nil response without error
			result.Success = false
			result.Status = "ERROR"
			result.ErrorMessage = "Unexpected error: nil response"
			failed++
		}

		results = append(results, result)
	}

	return &port.BatchFirstCollectionResponse{
		StatusCodeAndMessage: port.UpdateSuccess,
		TotalSubmitted:       totalSubmitted,
		Successful:           successful,
		Failed:               failed,
		Results:              results,
	}, nil
}

// BatchInstallments processes multiple installment payments
func (h *RevivalHandler) BatchInstallments(sctx *serverRoute.Context, req port.BatchInstallmentRequest) (*port.BatchInstallmentResponse, error) {
	results := make([]port.InstallmentResultItem, 0)

	for _, installment := range req.Installments {
		// Get revival request for workflow ID
		revivalReq, err := h.revivalRepo.GetRevivalRequestByTicketID(sctx.Ctx, installment.TicketID)
		InstNumber := revivalReq.InstallmentsPaid + 1
		if err != nil {
			result := port.InstallmentResultItem{
				TicketID:          installment.TicketID,
				PolicyNumber:      installment.PolicyNumber,
				InstallmentNumber: InstNumber,
				Success:           false,
				Status:            "ERROR",
				ErrorMessage:      fmt.Sprintf("Revival request not found: %v", err),
			}
			results = append(results, result)
			continue
		}

		// Reject payments when installments are already fully paid
		if revivalReq.InstallmentsPaid >= revivalReq.NumberOfInstallments {
			results = append(results, port.InstallmentResultItem{
				TicketID:          installment.TicketID,
				PolicyNumber:      installment.PolicyNumber,
				InstallmentNumber: installment.InstallmentNumber,
				Success:           false,
				Status:            "REJECTED",
				ErrorMessage:      "Installments already completed",
			})
			continue
		}

		installmentCount := installment.NumberOfInstallments
		if installmentCount < 1 {
			installmentCount = 1
		}
		endInstallment := installment.InstallmentNumber + installmentCount - 1
		if installment.InstallmentNumber > revivalReq.NumberOfInstallments || endInstallment > revivalReq.NumberOfInstallments {
			results = append(results, port.InstallmentResultItem{
				TicketID:          installment.TicketID,
				PolicyNumber:      installment.PolicyNumber,
				InstallmentNumber: installment.InstallmentNumber,
				Success:           false,
				Status:            "REJECTED",
				ErrorMessage:      fmt.Sprintf("Installment number exceeds approved count (%d)", revivalReq.NumberOfInstallments),
			})
			continue
		}

		// Check if number_of_installments is specified and > 1
		if installment.NumberOfInstallments > 1 {
			// Divide installment_amount by number_of_installments
			individualAmount := installment.InstallmentAmount / float64(installment.NumberOfInstallments)

			// Build batch input for workflow
			batchInstallments := make([]workflow.InstallmentPaymentData, installment.NumberOfInstallments)
			for i := 0; i < installment.NumberOfInstallments; i++ {
				batchInstallments[i] = workflow.InstallmentPaymentData{
					InstallmentNumber: installment.InstallmentNumber + i,
					Amount:            individualAmount,
					PaymentMode:       installment.PaymentMode,
				}
			}

			// Execute batch processing workflow synchronously
			workflowOptions := client.StartWorkflowOptions{
				ID:                 fmt.Sprintf("batch-installment-%s-%d-%d", revivalReq.RequestID, installment.InstallmentNumber, time.Now().Unix()),
				TaskQueue:          "revival",
				WorkflowRunTimeout: 5 * time.Minute,
			}

			batchInput := workflow.BatchInstallmentInput{
				RequestID:           revivalReq.RequestID,
				PolicyNumber:        revivalReq.PolicyNumber,
				TicketID:            installment.TicketID,
				Installments:        batchInstallments,
				AtomicMode:          true, // All-or-nothing: revert all if any fails
				MonitorWorkflowID:   fmt.Sprintf("installment-monitor-%s", revivalReq.RequestID),
				SignalMonitorOnPaid: true, // Signal monitor workflow to keep it in sync
			}

			log.Info(nil, "Starting BatchInstallmentProcessingWorkflow",
				"ticket_id", installment.TicketID,
				"num_installments", len(batchInstallments),
				"starting_installment", installment.InstallmentNumber)

			workflowRun, err := h.temporalClient.ExecuteWorkflow(sctx.Ctx, workflowOptions, "BatchInstallmentProcessingWorkflow", batchInput)
			if err != nil {
				log.Error(nil, "Failed to start batch workflow", "error", err)
				// Create ONE consolidated failed result for the entire batch
				results = append(results, port.InstallmentResultItem{
					TicketID:          installment.TicketID,
					PolicyNumber:      installment.PolicyNumber,
					InstallmentNumber: installment.InstallmentNumber,
					Success:           false,
					Status:            "ERROR",
					ErrorMessage:      fmt.Sprintf("Failed to start batch workflow for %d installments: %v", installment.NumberOfInstallments, err),
				})
				continue
			}

			// Wait for workflow to complete (synchronous)
			var batchResult workflow.BatchInstallmentResult
			err = workflowRun.Get(sctx.Ctx, &batchResult)
			if err != nil {
				log.Error(nil, "Batch workflow failed", "error", err)
				// Create ONE consolidated failed result for the entire batch
				results = append(results, port.InstallmentResultItem{
					TicketID:          installment.TicketID,
					PolicyNumber:      installment.PolicyNumber,
					InstallmentNumber: installment.InstallmentNumber,
					Success:           false,
					Status:            "ERROR",
					ErrorMessage:      fmt.Sprintf("Batch workflow error for %d installments: %v", installment.NumberOfInstallments, err),
				})
				continue
			}

			log.Info(nil, "Batch workflow completed",
				"ticket_id", installment.TicketID,
				"successful", batchResult.Successful,
				"failed", batchResult.Failed)

			// Create ONE consolidated result for the entire batch
			result := port.InstallmentResultItem{
				TicketID:          installment.TicketID,
				PolicyNumber:      installment.PolicyNumber,
				InstallmentNumber: installment.InstallmentNumber,
			}

			// In atomic mode: either all succeeded or all failed (reverted)
			if batchResult.Successful == len(batchInstallments) {
				// All succeeded
				result.Success = true
				result.Status = "PROCESSING"
				result.ReceiptNumber = fmt.Sprintf("BATCH%s", time.Now().Format("20060102150405"))
				result.Message = fmt.Sprintf("Batch of %d installments (starting from #%d) processed successfully", installment.NumberOfInstallments, installment.InstallmentNumber)
			} else {
				// At least one failed - all reverted in atomic mode
				result.Success = false
				result.Status = "REVERTED"
				result.ErrorMessage = fmt.Sprintf("Batch processing failed: %d succeeded, %d failed. All %d installments reverted (atomic mode)", batchResult.Successful, batchResult.Failed, installment.NumberOfInstallments)
				// Include failure details if available
				for _, wfResult := range batchResult.Results {
					if !wfResult.Success {
						result.ErrorMessage += fmt.Sprintf(" | Installment #%d: %s", wfResult.InstallmentNumber, wfResult.ErrorMessage)
						break // Just show first failure reason
					}
				}
			}

			results = append(results, result)

		} else {
			// Single installment - process as before
			result := port.InstallmentResultItem{
				TicketID:          installment.TicketID,
				PolicyNumber:      installment.PolicyNumber,
				InstallmentNumber: installment.InstallmentNumber,
			}

			response, err := h.CreateInstallment(sctx, installment)
			if err != nil {
				result.Success = false
				result.Status = "ERROR"
				result.ErrorMessage = err.Error()
			} else if response != nil && response.Success {
				result.Success = true
				result.Status = response.Data.Status
				result.ReceiptNumber = response.Data.ReceiptNumber
				result.InstallmentAmount = response.Data.InstallmentAmount
				result.Message = "Installment processed successfully"
			} else if response != nil {
				result.Success = false
				result.Status = response.Data.Status
				if response.Message != "" {
					result.ErrorMessage = response.Message
				} else {
					result.ErrorMessage = "Installment validation failed"
				}
			} else {
				result.Success = false
				result.Status = "ERROR"
				result.ErrorMessage = "Unexpected error: nil response"
			}

			results = append(results, result)
		}
	}

	// Calculate summary
	totalSubmitted := len(results)
	successful := 0
	failed := 0
	for _, r := range results {
		if r.Success {
			successful++
		} else {
			failed++
		}
	}

	return &port.BatchInstallmentResponse{
		StatusCodeAndMessage: port.UpdateSuccess,
		TotalSubmitted:       totalSubmitted,
		Successful:           successful,
		Failed:               failed,
		Results:              results,
	}, nil
}

// GetInstallmentCollection retrieves installment collection calculation details
func (h *RevivalHandler) GetInstallmentCollection(sctx *serverRoute.Context, req port.GetInstallmentCollectionRequest) (*port.GetInstallmentCollectionResponse, error) {
	// Get revival request to verify it exists and get policy details
	var revivalReq domain.RevivalRequest
	var err error
	if req.PolicyNumber == "" {
		errMsg := apierrors.HandleErrorWithStatusCodeAndMessage(
			apierrors.CustomError,
			"policy_number is required",
			nil,
		)
		return nil, errMsg
	}
	revivalReq, err = h.revivalRepo.GetLatestRevivalRequestByPolicyNumber(sctx.Ctx, req.PolicyNumber)
	if err != nil {
		errMsg := apierrors.HandleErrorWithStatusCodeAndMessage(
			apierrors.DBErrorRecordNotFound,
			"No revival request found for the given policy number",
			err,
		)
		return nil, errMsg
	}

	// Fetch current installment number from DB (installments_paid + 1)
	currentInstallment := revivalReq.InstallmentsPaid + 1
	remainingInstallments := revivalReq.NumberOfInstallments - revivalReq.InstallmentsPaid
	num := remainingInstallments
	if remainingInstallments <= 0 {
		return &port.GetInstallmentCollectionResponse{
			StatusCodeAndMessage: port.InstallmentNotFound,
			Data:                 port.InstallmentCollectionData{},
		}, nil
	}
	// If first installment is not done, always return only the first installment, regardless of number_of_installments
	if currentInstallment == 1 {
		num = 0
	}
	if req.NumberOfInstallments != nil && currentInstallment != 1 {
		if *req.NumberOfInstallments <= 0 {
			num = 1
		} else if *req.NumberOfInstallments < remainingInstallments {
			num = *req.NumberOfInstallments
		}
	}

	var totalAmount float64
	var subtotal float64
	var totalGST float64
	var premiumAmount *float64
	var premiumCGST, premiumSGST *float64
	var installmentCGST, installmentSGST *float64

	fmt.Printf("[DEBUG] TicketID: %s, currentInstallment: %d, num: %d, NumberOfInstallments: %v\n", revivalReq.TicketID, currentInstallment, num, req.NumberOfInstallments)

	// Get policy premium from common.policys table
	policy, err := h.policyRepo.GetPolicyByNumber(sctx.Ctx, req.PolicyNumber)
	if err != nil {
		log.Error(sctx.Ctx, "Failed to get policy", "policy_number", req.PolicyNumber, "error", err)
		return &port.GetInstallmentCollectionResponse{
			StatusCodeAndMessage: port.StatusCodeAndMessage{
				StatusCode: 400,
				Success:    false,
				Message:    fmt.Sprintf("Policy %s not found", req.PolicyNumber),
			},
			Data: port.InstallmentCollectionData{},
		}, nil
	}

	fmt.Printf("[DEBUG] Policy Premium: %.2f\n", policy.PremiumAmount)

	branch := ""
	// Treat num == 0 or num == 1 as single installment calculation
	if num == 0 || num == 1 {
		branch = "single"
		// Calculate for current installment only
		// Determine which amount to use based on RevivalType
		var amountToUse float64
		if revivalReq.RevivalType != nil && *revivalReq.RevivalType == "lumpsum" {
			amountToUse = math.Round(revivalReq.RevivalAmount)
		} else {
			amountToUse = math.Round(revivalReq.InstallmentAmount)
		}

		if currentInstallment == 1 {
			roundedPremium := math.Round(policy.PremiumAmount)
			premiumAmount = &roundedPremium
			premCGST := math.Round(roundedPremium * 0.09)
			premSGST := math.Round(roundedPremium * 0.09)
			premiumCGST = &premCGST
			premiumSGST = &premSGST
			instCGST := math.Round(amountToUse * 0.09)
			instSGST := math.Round(amountToUse * 0.09)
			installmentCGST = &instCGST
			installmentSGST = &instSGST
			subtotal = math.Round(roundedPremium + amountToUse)
			totalGST = math.Round(premCGST + premSGST + instCGST + instSGST)
		} else {
			// From 2nd installment onwards, NO GST is applied
			// installmentCGST and installmentSGST remain nil (excluded from JSON)
			subtotal = math.Round(amountToUse)
			totalGST = 0
		}
		totalAmount = math.Round(subtotal + totalGST)
		fmt.Printf("[DEBUG] Branch: %s, premiumAmount: %v, premiumCGST: %v, premiumSGST: %v\n", branch, premiumAmount, premiumCGST, premiumSGST)
		totalInstallments := revivalReq.NumberOfInstallments
		if revivalReq.RevivalType != nil && *revivalReq.RevivalType == "lumpsum" && totalInstallments == 0 {
			totalInstallments = 1
		}
		return &port.GetInstallmentCollectionResponse{
			StatusCodeAndMessage: port.FetchSuccess,
			Data: port.InstallmentCollectionData{
				TicketID:             revivalReq.TicketID,
				PolicyNumber:         revivalReq.PolicyNumber,
				CustomerName:         policy.CustomerName,
				RevivalType:          revivalReq.RevivalType,
				InstallmentNumber:    currentInstallment,
				NumberOfInstallments: 1,
				TotalInstallments:    totalInstallments,
				PremiumAmount:        premiumAmount,
				InstallmentAmount:    amountToUse,
				PremiumCGST:          premiumCGST,
				PremiumSGST:          premiumSGST,
				InstallmentCGST:      installmentCGST,
				InstallmentSGST:      installmentSGST,
				TotalAmount:          totalAmount,
			},
		}, nil
	}

	// Handle multiple installments (num > 1)
	// GST only applies to first collection, so subsequent installments have NO GST
	branch = "multiple"
	// Determine which amount to use based on RevivalType
	var amountToUse float64
	if revivalReq.RevivalType != nil && *revivalReq.RevivalType == "lumpsum" {
		amountToUse = math.Round(revivalReq.RevivalAmount)
	} else {
		amountToUse = math.Round(revivalReq.InstallmentAmount)
	}
	installmentAmount := math.Round(amountToUse * float64(num))
	subtotal = installmentAmount
	totalGST = 0 // No GST for subsequent installments
	totalAmount = installmentAmount

	fmt.Printf("[DEBUG] Branch: %s, num: %d, installmentAmount: %.2f, totalAmount: %.2f\n", branch, num, installmentAmount, totalAmount)

	totalInstallments := revivalReq.NumberOfInstallments
	if revivalReq.RevivalType != nil && *revivalReq.RevivalType == "lumpsum" && totalInstallments == 0 {
		totalInstallments = 1
	}
	return &port.GetInstallmentCollectionResponse{
		StatusCodeAndMessage: port.FetchSuccess,
		Data: port.InstallmentCollectionData{
			TicketID:             revivalReq.TicketID,
			PolicyNumber:         revivalReq.PolicyNumber,
			CustomerName:         policy.CustomerName,
			RevivalType:          revivalReq.RevivalType,
			InstallmentNumber:    currentInstallment,
			NumberOfInstallments: num,
			PremiumAmount:        nil,
			InstallmentAmount:    installmentAmount,
			PremiumCGST:          nil,
			PremiumSGST:          nil,
			InstallmentCGST:      nil,
			InstallmentSGST:      nil,
			TotalAmount:          totalAmount,
			TotalInstallments:    totalInstallments,
		},
	}, nil
}

// Quotation calculates revival quotation
func (h *RevivalHandler) Quotation(sctx *serverRoute.Context, req port.QuotationRequest) (*port.QuotationResponse, error) {
	installments := 0
	if req.Installments != nil {
		installments = *req.Installments
	}

	// Determine quote calculation date (defaults to today if not provided)
	quoteDate := time.Now()
	if req.QuoteDate != nil {
		quoteDate = *req.QuoteDate
	}

	// Get policy details
	policy, err := h.policyRepo.GetPolicyByNumber(sctx.Ctx, req.PolicyNumber)
	if err != nil {
		errMsg := apierrors.HandleErrorWithStatusCodeAndMessage(
			apierrors.DBErrorRecordNotFound,
			"No policy found for the given policy number",
			err,
		)
		return nil, errMsg
	}

	// Calculate unpaid months (from paid_to_date to quote date)
	var unpaidMonths int
	if policy.PaidToDate != nil {
		unpaidMonths = int(quoteDate.Sub(*policy.PaidToDate).Hours() / 24 / 30)
		if unpaidMonths < 1 {
			unpaidMonths = 1
		}
	} else {
		unpaidMonths = 1
	}

	// Get monthly premium based on frequency
	monthlyPremium := policy.PremiumAmount
	if policy.PremiumFrequency == "YEARLY" {
		monthlyPremium = policy.PremiumAmount / 12
	} else if policy.PremiumFrequency == "QUARTERLY" {
		monthlyPremium = policy.PremiumAmount / 3
	} else if policy.PremiumFrequency == "HALF_YEARLY" {
		monthlyPremium = policy.PremiumAmount / 6
	}

	// IR_5: Revival Amount Formula
	// Revival Amount = [{(1 + Interest)^(Unpaid Months)} - 1] × 101 × Monthly Premium
	// Note: Interest rate would typically come from system configuration
	// Using 8% annual interest rate = 0.08/12 = 0.00667 monthly
	monthlyInterestRate := 0.08 / 12.0
	totalPremiumDue := monthlyPremium * float64(unpaidMonths)

	revivalAmount := (math.Pow(1+monthlyInterestRate, float64(unpaidMonths)) - 1) * 101 * monthlyPremium
	interest := revivalAmount - (monthlyPremium * float64(unpaidMonths))

	// IR_5: Installment Amount Formula
	// Installment Amount = (Revival Amount × Interest) × [{(1+i)^N}/{(1+i)^N - 1}]
	// Where N = number of installments, i = monthly interest rate
	var installmentAmount float64
	if installments > 0 && installments != 1 {
		N := float64(installments)
		i := monthlyInterestRate
		numerator := math.Pow(1+i, N)
		denominator := math.Pow(1+i, N) - 1
		installmentAmount = (revivalAmount * i) * (numerator / denominator)
	} else {
		installmentAmount = 0
	}

	// Calculate tax (18% GST = 9% CGST + 9% SGST)
	var cgst = 0.0
	var sgst = 0.0

	cgst = revivalAmount * 0.09
	sgst = revivalAmount * 0.09
	totalTax := cgst + sgst

	// Round all monetary outputs to nearest rupee
	revivalAmount = math.Round(revivalAmount)
	interest = math.Round(interest)
	installmentAmount = math.Round(installmentAmount)
	cgst = math.Round(cgst)
	sgst = math.Round(sgst)
	totalTax = math.Round(totalTax)
	totalPremiumDue = math.Round(totalPremiumDue)

	// Determine revival type based on installments
	revivalType := "installment"
	if installments == 0 || installments == 1 {
		revivalType = "lumpsum"
	}

	// Calculate total amount due
	var totalAmountDue float64
	if installments == 0 || installments == 1 {
		// Lumpsum: revival amount + total tax
		totalAmountDue = revivalAmount + totalTax
	} else {
		// Installment: (installment amount * number of installments) + total tax
		totalAmountDue = (installmentAmount * float64(installments)) + totalTax
	}
	totalAmountDue = math.Round(totalAmountDue)

	// Quote valid for 30 days from quote date
	quoteValidUntil := quoteDate.AddDate(0, 0, 30)

	return &port.QuotationResponse{
		StatusCodeAndMessage: port.FetchSuccess,
		Data: port.QuotationData{
			PolicyNumber:      req.PolicyNumber,
			RevivalType:       revivalType,
			Installments:      installments,
			RevivalAmount:     revivalAmount,
			Interest:          interest,
			InstallmentAmount: installmentAmount,
			TaxBreakdown: port.TaxBreakdown{
				CGST:     cgst,
				SGST:     sgst,
				TotalTax: totalTax,
			},
			TotalAmountDue:  totalAmountDue,
			QuoteValidUntil: quoteValidUntil,
			TotalPremiumDue: totalPremiumDue,
			FromDueDate:     policy.PaidToDate,
			ToDueDate:       quoteDate,
			Premium:         policy.PremiumAmount,
		},
	}, nil
}

// ReceiveDocuments receives documents for revival request
func (h *RevivalHandler) ReceiveDocuments(sctx *serverRoute.Context, req port.ReceiveDocumentsRequest) (*port.DataEntryResponse, error) {
	// Check if revival request exists
	revivalReq, err := h.revivalRepo.GetRevivalRequestByTicketID(sctx.Ctx, req.TicketID)
	if err != nil {
		errMsg := apierrors.HandleErrorWithStatusCodeAndMessage(
			apierrors.DBErrorRecordNotFound,
			"No revival request found for the given ticket ID",
			err,
		)
		return nil, errMsg
	}

	// Set received date if not provided
	// Update revival request with document details
	status := "DOCUMENTS_PENDING"
	message := fmt.Sprintf("Received %d documents", len(req.DocumentsReceived))

	if req.AllDocumentsReceived {
		status = "DOCUMENTS_COMPLETE"
		message = "All required documents received successfully"

		// Update revival request to mark documents complete
		err = h.revivalRepo.UpdateRevivalRequestStatus(sctx.Ctx, revivalReq.RequestID, "DOCUMENTS_COMPLETE", "SYSTEM")
		if err != nil {
			errMsg := apierrors.HandleErrorWithStatusCodeAndMessage(
				apierrors.CustomError,
				"Unable to update document status",
				err,
			)
			return nil, errMsg
		}
	}

	// Note: In production, would store document details in a documents table
	// For now, we're just updating the revival request status
	_ = revivalReq
	return &port.DataEntryResponse{
		StatusCodeAndMessage: port.UpdateSuccess,
		Data: port.DataEntryData{
			TicketID: req.TicketID,
			Status:   status,
			Message:  message,
		},
	}, nil
}

// GenerateAcceptanceLetter generates acceptance letter
func (h *RevivalHandler) GenerateAcceptanceLetter(sctx *serverRoute.Context, req port.RevivalAcceptanceLetterRequest) (*port.DataEntryResponse, error) {
	// Check if revival request exists
	revivalReq, err := h.revivalRepo.GetRevivalRequestByTicketID(sctx.Ctx, req.TicketID)
	if err != nil {
		errMsg := apierrors.HandleErrorWithStatusCodeAndMessage(
			apierrors.DBErrorRecordNotFound,
			"No revival request found for the given ticket ID",
			err,
		)
		return nil, errMsg
	}

	// Check if request is approved
	if revivalReq.CurrentStatus != "APPROVED" {
		return &port.DataEntryResponse{
			StatusCodeAndMessage: port.InvalidTicketStatus,
			Data: port.DataEntryData{
				Message: "Acceptance letter can only be generated for approved requests",
			},
		}, nil
	}

	// Generate acceptance letter using activity
	letterID, err := h.activities.GenerateLetterActivity(sctx.Ctx, "ACCEPTANCE", revivalReq.RequestID, revivalReq.PolicyNumber)
	if err != nil {
		errMsg := apierrors.HandleErrorWithStatusCodeAndMessage(
			apierrors.CustomError,
			"Unable to generate acceptance letter",
			err,
		)
		return nil, errMsg
	}

	// Send notification to policyholder
	err = h.activities.SendNotificationActivity(sctx.Ctx, revivalReq.PolicyNumber,
		fmt.Sprintf("Your revival request %s has been approved. Letter ID: %s", req.TicketID, letterID),
		"EMAIL")
	if err != nil {
		// Log error but don't fail the request
	}

	return &port.DataEntryResponse{
		StatusCodeAndMessage: port.CreateSuccess,
		Data: port.DataEntryData{
			TicketID: req.TicketID,
			Status:   "LETTER_GENERATED",
			Message:  fmt.Sprintf("Acceptance letter generated successfully. Letter ID: %s", letterID),
		},
	}, nil
}

// GenerateRevivalMemo generates revival memo
func (h *RevivalHandler) GenerateRevivalMemo(sctx *serverRoute.Context, req port.RevivalMemoRequest) (*port.DataEntryResponse, error) {
	// Check if revival request exists
	revivalReq, err := h.revivalRepo.GetRevivalRequestByTicketID(sctx.Ctx, req.TicketID)
	if err != nil {
		errMsg := apierrors.HandleErrorWithStatusCodeAndMessage(
			apierrors.DBErrorRecordNotFound,
			"No revival request found for the given ticket ID",
			err,
		)
		return nil, errMsg
	}

	// Check if first collection is done (memo is generated after first collection)
	if !revivalReq.FirstCollectionDone {
		return &port.DataEntryResponse{
			StatusCodeAndMessage: port.InvalidTicketStatus,
			Data: port.DataEntryData{
				Message: "Revival memo can only be generated after first collection",
			},
		}, nil
	}

	// Generate revival memo using activity
	memoID, err := h.activities.GenerateLetterActivity(sctx.Ctx, "MEMO", revivalReq.RequestID, revivalReq.PolicyNumber)
	if err != nil {
		errMsg := apierrors.HandleErrorWithStatusCodeAndMessage(
			apierrors.CustomError,
			"Unable to generate revival memo",
			err,
		)
		return nil, errMsg
	}

	// Send notification to relevant parties
	err = h.activities.SendNotificationActivity(sctx.Ctx, revivalReq.PolicyNumber,
		fmt.Sprintf("Revival memo generated for policy %s. Memo ID: %s", revivalReq.PolicyNumber, memoID),
		"EMAIL")
	if err != nil {
		// Log error but don't fail the request
	}

	return &port.DataEntryResponse{
		StatusCodeAndMessage: port.CreateSuccess,
		Data: port.DataEntryData{
			TicketID: req.TicketID,
			Status:   "MEMO_GENERATED",
			Message:  fmt.Sprintf("Revival memo generated successfully. Memo ID: %s", memoID),
		},
	}, nil
}

// validateDocuments validates that required documents are present
// Read-only validation used at QC and approval stages
func validateDocuments(documents *string, missingDocsList *string) error {
	// If no documents field in database yet (nil), skip validation
	if documents == nil && missingDocsList == nil {
		return nil
	}

	// Parse documents JSON
	var docs []port.DocumentSubmission
	if documents != nil && *documents != "" && *documents != "[]" {
		if err := json.Unmarshal([]byte(*documents), &docs); err != nil {
			return fmt.Errorf("invalid documents JSON: %w", err)
		}
	}

	// Parse missing documents JSON
	var missingDocs []port.MissingDocument
	if missingDocsList != nil && *missingDocsList != "" && *missingDocsList != "[]" {
		if err := json.Unmarshal([]byte(*missingDocsList), &missingDocs); err != nil {
			return fmt.Errorf("invalid missing_documents_list JSON: %w", err)
		}
	}

	// If there are still missing documents, validation fails
	if len(missingDocs) > 0 {
		docNames := make([]string, len(missingDocs))
		for i, doc := range missingDocs {
			docNames[i] = doc.DocumentName
		}
		return fmt.Errorf("%d documents still missing: %v", len(missingDocs), docNames)
	}

	// All validations passed
	return nil
}

// determineNextAction returns the next action required and the actor based on current status
func determineNextAction(status string) (nextAction string, nextActor string) {
	switch status {
	case "INDEXED", "WAITING_FOR_DATA_ENTRY":
		return "Submit data entry", "Data Entry Operator"
	case "DATA_ENTRY_PENDING":
		return "Rework data entry (failed QC)", "Data Entry Operator"
	case "DATA_ENTRY_COMPLETE", "WAITING_FOR_QC":
		return "Perform quality check", "QC Officer"
	case "APPROVAL_PENDING", "WAITING_FOR_APPROVAL":
		return "Approve or reject request", "Approver"
	case "APPROVED":
		return "Collect first installment (dual payment)", "Collection Officer"
	case "ACTIVE":
		return "Monitor installment payments", "System/Collection Officer"
	case "COMPLETED":
		return "Revival completed successfully", "None"
	case "REJECTED":
		return "Request rejected", "None"
	case "WITHDRAWN":
		return "Request withdrawn", "None"
	case "DEFAULTED":
		return "Handle default and suspense reversal", "Finance Officer"
	case "RETURNED_TO_INDEXER":
		return "Re-index request", "Indexer"
	case "SLA_EXPIRED":
		return "Handle SLA expiry", "Manager"
	case "VALIDATING_POLICY":
		return "Validating policy eligibility", "System"
	case "PROCESSING_DATA_ENTRY", "PROCESSING_QC", "PROCESSING_APPROVAL":
		return "Processing in workflow", "System"
	default:
		return "Unknown status - check workflow", "Unknown"
	}
}

// Helper function to convert string to *string
func stringPtr(s string) *string {
	return &s
}
