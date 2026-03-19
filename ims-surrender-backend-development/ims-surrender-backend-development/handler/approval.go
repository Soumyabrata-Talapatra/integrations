package handler

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	log "gitlab.cept.gov.in/it-2.0-common/n-api-log"
	serverHandler "gitlab.cept.gov.in/it-2.0-common/n-api-server/handler"
	serverRoute "gitlab.cept.gov.in/it-2.0-common/n-api-server/route"

	"gitlab.cept.gov.in/it-2.0-policy/surrender-service/core/domain"
	"gitlab.cept.gov.in/it-2.0-policy/surrender-service/core/port"
	repo "gitlab.cept.gov.in/it-2.0-policy/surrender-service/repo/postgres"
)

// ApprovalHandler handles all approval workflow operations
// Business Rules: BR-APPR-001 to BR-APPR-007
// Functional Requirements: FR-APPR-001 to FR-APPR-007
type ApprovalHandler struct {
	*serverHandler.Base
	surrenderRepo *repo.SurrenderRequestRepository
	approvalRepo  *repo.ApprovalWorkflowRepository
	// External service placeholders
	policyService PolicyServiceInterface
	userService   UserServiceInterface
}

// NewApprovalHandler creates a new approval handler
func NewApprovalHandler(
	surrenderRepo *repo.SurrenderRequestRepository,
	approvalRepo *repo.ApprovalWorkflowRepository,
) *ApprovalHandler {
	base := serverHandler.New("Approval Workflow").SetPrefix("/v1").AddPrefix("/approval")

	return &ApprovalHandler{
		Base:          base,
		surrenderRepo: surrenderRepo,
		approvalRepo:  approvalRepo,
		// Initialize placeholders
		policyService: NewMockPolicyService(),
		userService:   NewMockUserService(),
	}
}

// Routes defines all routes for approval workflow
func (h *ApprovalHandler) Routes() []serverRoute.Route {
	return []serverRoute.Route{
		serverRoute.GET("/queue", h.GetApprovalQueue).Name("Get Approval Queue"),
		serverRoute.POST("/reserve-task", h.ReserveApprovalTask).Name("Reserve Approval Task"),
		serverRoute.POST("/release-task", h.ReleaseApprovalTask).Name("Release Approval Task"),
		serverRoute.POST("/approve", h.ApproveSurrenderRequest).Name("Approve Surrender Request"),
		serverRoute.POST("/reject", h.RejectSurrenderRequest).Name("Reject Surrender Request"),
		serverRoute.POST("/recalculate", h.RecalculateSurrenderValue).Name("Recalculate Surrender Value"),
		serverRoute.POST("/escalate", h.EscalateApprovalTask).Name("Escalate Approval Task"),
		serverRoute.GET("/history", h.GetApprovalHistory).Name("Get Approval History"),
	}
}

// GetApprovalQueue retrieves approval queue with filtering and pagination
// GET /v1/approval/queue
// Business Rule: BR-APPR-001
// Returns paginated list of surrender requests pending approval
func (h *ApprovalHandler) GetApprovalQueue(sctx *serverRoute.Context, req ApprovalQueueParams) (interface{}, error) {
	// Parse filters (currently not used in repository call, but kept for future use)
	var _ *domain.SurrenderStatus
	if req.Status != "" {
		s := domain.SurrenderStatus(req.Status)
		_ = &s
	}

	var _ *domain.SurrenderRequestType
	if req.RequestType != "" {
		rt := domain.SurrenderRequestType(req.RequestType)
		_ = &rt
	}

	var _ *domain.TaskStatus
	if req.TaskStatus != "" {
		ts := domain.TaskStatus(req.TaskStatus)
		_ = &ts
	}

	// Set pagination defaults
	page := 1
	if req.Page > 0 {
		page = req.Page
	}

	limit := 20
	if req.Limit > 0 && req.Limit <= 100 {
		limit = req.Limit
	}

	offset := (page - 1) * limit

	// BR-APPR-001: Get approval queue with filters
	// Use office code from query params, default to "ALL" if not provided
	officeCode := req.OfficeCode
	if officeCode == "" {
		officeCode = "ALL" // Default to ALL if not specified
	}
	tasks, totalCount, err := h.approvalRepo.ListApprovalQueue(sctx.Ctx, officeCode, uint64(offset), uint64(limit))
	if err != nil {
		log.Error(sctx.Ctx, "Failed to get approval queue: %v", err)
		return nil, fmt.Errorf("failed to retrieve approval queue")
	}

	// Build response data
	queueItems := make([]ApprovalQueueItem, 0, len(tasks))
	for _, task := range tasks {
		// Get surrender request details
		surrenderRequest, err := h.surrenderRepo.FindByID(sctx.Ctx, task.SurrenderRequestID)
		if err != nil {
			log.Error(sctx.Ctx, "Failed to get surrender request %s: %v", task.SurrenderRequestID, err)
			continue
		}

		// Build queue item
		item := ApprovalQueueItem{
			TaskID:             task.ID.String(),
			SurrenderRequestID: task.SurrenderRequestID.String(),
			RequestNumber:      surrenderRequest.RequestNumber,
			PolicyID:           surrenderRequest.PolicyID,
			PolicyNumber:       h.getMetadataString(surrenderRequest.Metadata, "policy_number"),
			PolicyholderName:   h.getMetadataString(surrenderRequest.Metadata, "policyholder_name"),
			RequestType:        string(surrenderRequest.RequestType),
			RequestDate:        surrenderRequest.RequestDate.Format(time.RFC3339),
			NetSurrenderValue:  surrenderRequest.NetSurrenderValue,
			SurrenderStatus:    string(surrenderRequest.Status),
			TaskStatus:         string(task.Status),
			Priority:           string(task.Priority),
			AssignedTo:         h.getUserName(task.AssignedTo),
			AssignedDate:       h.formatTime(task.ReservedAt),
			DueDate:            h.formatTime(task.ReservationExpiresAt),
			DaysInQueue:        int(time.Since(task.CreatedAt).Hours() / 24),
		}

		queueItems = append(queueItems, item)
	}

	totalPages := (int(totalCount) + limit - 1) / limit

	log.Info(sctx.Ctx, "Retrieved %d items from approval queue (page %d/%d)", len(queueItems), page, totalPages)

	return &ApprovalQueueResponse{
		StatusCodeAndMessage: port.GetSuccess,
		Data: ApprovalQueueData{
			Items:      queueItems,
			TotalCount: int(totalCount),
			Page:       page,
			Limit:      limit,
			TotalPages: totalPages,
		},
	}, nil
}

// ReserveApprovalTask reserves a task for processing
// POST /v1/approval/reserve-task
// Business Rule: BR-APPR-002
// Assigns task to CPC user and prevents concurrent processing
func (h *ApprovalHandler) ReserveApprovalTask(sctx *serverRoute.Context, req ReserveTaskRequest) (interface{}, error) {
	taskID, err := uuid.Parse(req.TaskID)
	if err != nil {
		log.Error(sctx.Ctx, "Invalid task ID: %v", err)
		return nil, fmt.Errorf("invalid task ID format")
	}

	// Mock user ID (in production, get from auth context)
	userID := uuid.New()
	if req.UserID != "" {
		parsedUserID, err := uuid.Parse(req.UserID)
		if err == nil {
			userID = parsedUserID
		}
	}

	// BR-APPR-002: Reserve task with timeout
	dueDate := time.Now().Add(24 * time.Hour) // 24-hour SLA
	reserved, err := h.approvalRepo.ReserveTask(sctx.Ctx, taskID, userID, dueDate)
	if err != nil {
		if err.Error() == "task already reserved" {
			log.Error(sctx.Ctx, "Task %s already reserved", req.TaskID)
			return nil, fmt.Errorf("task is already reserved by another user")
		}
		if err == pgx.ErrNoRows {
			log.Error(sctx.Ctx, "Task %s not found in approval queue", req.TaskID)
			return nil, fmt.Errorf("task not found. Ensure the task ID exists in the approval queue")
		}
		log.Error(sctx.Ctx, "Failed to reserve task: %v", err)
		return nil, fmt.Errorf("failed to reserve task")
	}

	// Get surrender request details
	surrenderRequest, err := h.surrenderRepo.FindByID(sctx.Ctx, reserved.SurrenderRequestID)
	if err != nil {
		log.Error(sctx.Ctx, "Failed to get surrender request: %v", err)
		return nil, err
	}

	log.Info(sctx.Ctx, "Reserved task %s for user %s", reserved.ID, userID)

	return &ReserveTaskResponse{
		StatusCodeAndMessage: port.UpdateSuccess,
		Data: ReserveTaskData{
			TaskID:             reserved.ID.String(),
			SurrenderRequestID: reserved.SurrenderRequestID.String(),
			RequestNumber:      surrenderRequest.RequestNumber,
			PolicyNumber:       h.getMetadataString(surrenderRequest.Metadata, "policy_number"),
			Status:             string(reserved.Status),
			AssignedTo:         userID.String(),
			AssignedDate:       reserved.ReservedAt.Format(time.RFC3339),
			DueDate:            reserved.ReservationExpiresAt.Format(time.RFC3339),
			Message:            "Task reserved successfully. You have 24 hours to complete this task.",
		},
	}, nil
}

// ReleaseApprovalTask releases a reserved task
// POST /v1/approval/release-task
// Business Rule: BR-APPR-003
// Returns task to queue if user cannot complete it
func (h *ApprovalHandler) ReleaseApprovalTask(sctx *serverRoute.Context, req ReleaseTaskRequest) (interface{}, error) {
	taskID, err := uuid.Parse(req.TaskID)
	if err != nil {
		log.Error(sctx.Ctx, "Invalid task ID: %v", err)
		return nil, fmt.Errorf("invalid task ID format")
	}

	// Mock user ID
	userID := uuid.New()
	if req.UserID != "" {
		parsedUserID, err := uuid.Parse(req.UserID)
		if err == nil {
			userID = parsedUserID
		}
	}

	// BR-APPR-003: Release task back to queue
	released, err := h.approvalRepo.ReleaseTask(sctx.Ctx, taskID)
	if err != nil {
		log.Error(sctx.Ctx, "Failed to release task: %v", err)
		return nil, fmt.Errorf("failed to release task")
	}

	log.Info(sctx.Ctx, "Released task %s by user %s", released.ID, userID)

	return &ReleaseTaskResponse{
		StatusCodeAndMessage: port.UpdateSuccess,
		Data: ReleaseTaskData{
			TaskID:     released.ID.String(),
			Status:     string(released.Status),
			ReleasedBy: userID.String(),
			ReleasedAt: time.Now().Format(time.RFC3339),
			Message:    "Task released successfully and returned to queue",
		},
	}, nil
}

// ApproveSurrenderRequest approves a surrender request
// POST /v1/approval/approve
// Business Rules: BR-APPR-004, BR-APPR-005
// Approves surrender and updates policy status
func (h *ApprovalHandler) ApproveSurrenderRequest(sctx *serverRoute.Context, req ApproveSurrenderRequest) (interface{}, error) {
	surrenderRequestID, err := uuid.Parse(req.SurrenderRequestID)
	if err != nil {
		log.Error(sctx.Ctx, "Invalid surrender request ID: %v", err)
		return nil, fmt.Errorf("invalid surrender request ID format")
	}

	// Mock user ID
	userID := uuid.New()
	if req.ApproverUserID != "" {
		parsedUserID, err := uuid.Parse(req.ApproverUserID)
		if err == nil {
			userID = parsedUserID
		}
	}

	// Get surrender request
	surrenderRequest, err := h.surrenderRepo.FindByID(sctx.Ctx, surrenderRequestID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("surrender request not found")
		}
		log.Error(sctx.Ctx, "Failed to get surrender request: %v", err)
		return nil, err
	}

	// Validate current status
	if surrenderRequest.Status != domain.SurrenderStatusPendingApproval {
		return nil, fmt.Errorf("surrender request is not pending approval (current status: %s)", surrenderRequest.Status)
	}

	// BR-APPR-004: Approve surrender request
	oldStatus := surrenderRequest.Status
	approved, err := h.surrenderRepo.UpdateStatus(sctx.Ctx, surrenderRequestID, domain.SurrenderStatusApproved, userID, &req.ApprovalComments)
	if err != nil {
		log.Error(sctx.Ctx, "Failed to approve surrender request: %v", err)
		return nil, fmt.Errorf("failed to approve surrender request")
	}

	// Complete approval task
	task, err := h.approvalRepo.FindTaskBySurrenderRequestID(sctx.Ctx, surrenderRequestID)
	if err == nil {
		_, err = h.approvalRepo.CompleteTask(sctx.Ctx, task.ID, userID)
		if err != nil {
			log.Error(sctx.Ctx, "Failed to complete approval task: %v", err)
		}
	}

	// BR-APPR-005: Update policy status
	// Determine policy status based on disposition
	var newPolicyStatus string
	if approved.NetSurrenderValue >= h.getPrescribedLimit(surrenderRequest.PolicyID) {
		newPolicyStatus = "AU" // Reduced Paid-Up
	} else {
		newPolicyStatus = "TS" // Terminated Surrender
	}

	err = h.policyService.UpdatePolicyStatus(sctx.Ctx, approved.PolicyID, newPolicyStatus)
	if err != nil {
		log.Error(sctx.Ctx, "Failed to update policy status: %v", err)
		// Continue even if policy update fails - can be retried
	}

	log.Info(sctx.Ctx, "Approved surrender request %s, new status: %s", approved.RequestNumber, approved.Status)

	// Build workflow state
	workflowState := WorkflowStateData{
		CurrentStage:    "APPROVED",
		CompletedStages: []string{"REQUEST_CREATED", "DOCUMENT_UPLOAD", "VERIFICATION", "APPROVAL"},
		PendingStages:   []string{"PAYMENT_PROCESSING"},
		ProgressPercent: 85,
	}

	response := &ApproveSurrenderResponse{
		StatusCodeAndMessage: port.UpdateSuccess,
		Data: ApproveSurrenderData{
			SurrenderRequestID: approved.ID.String(),
			RequestNumber:      approved.RequestNumber,
			PolicyID:           approved.PolicyID,
			PolicyNumber:       h.getMetadataString(approved.Metadata, "policy_number"),
			OldStatus:          string(oldStatus),
			NewStatus:          string(approved.Status),
			NewPolicyStatus:    newPolicyStatus,
			ApprovedBy:         userID.String(),
			ApprovedDate:       time.Now().Format(time.RFC3339),
			ApprovalComments:   req.ApprovalComments,
			NetSurrenderValue:  approved.NetSurrenderValue,
			WorkflowState:      workflowState,
			NextAction: NextActionData{
				Action:      "PROCESS_PAYMENT",
				Description: "Surrender approved. Payment processing will be initiated.",
				URL:         "/v1/surrender/status?surrender_request_id=" + approved.ID.String(),
			},
		},
	}

	return response, nil
}

// RejectSurrenderRequest rejects a surrender request
// POST /v1/approval/reject
// Business Rule: BR-APPR-006
// Rejects surrender with reason and notification
func (h *ApprovalHandler) RejectSurrenderRequest(sctx *serverRoute.Context, req RejectSurrenderRequest) (interface{}, error) {
	surrenderRequestID, err := uuid.Parse(req.SurrenderRequestID)
	if err != nil {
		log.Error(sctx.Ctx, "Invalid surrender request ID: %v", err)
		return nil, fmt.Errorf("invalid surrender request ID format")
	}

	// Validate rejection reason is provided
	if req.RejectionReason == "" {
		return nil, fmt.Errorf("rejection reason is required")
	}

	// Mock user ID
	userID := uuid.New()
	if req.RejectorUserID != "" {
		parsedUserID, err := uuid.Parse(req.RejectorUserID)
		if err == nil {
			userID = parsedUserID
		}
	}

	// Get surrender request
	surrenderRequest, err := h.surrenderRepo.FindByID(sctx.Ctx, surrenderRequestID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("surrender request not found")
		}
		log.Error(sctx.Ctx, "Failed to get surrender request: %v", err)
		return nil, err
	}

	// Validate current status
	if surrenderRequest.Status != domain.SurrenderStatusPendingApproval {
		return nil, fmt.Errorf("surrender request is not pending approval (current status: %s)", surrenderRequest.Status)
	}

	// BR-APPR-006: Reject surrender request
	oldStatus := surrenderRequest.Status
	rejected, err := h.surrenderRepo.UpdateStatus(sctx.Ctx, surrenderRequestID, domain.SurrenderStatusRejected, userID, &req.RejectionReason)
	if err != nil {
		log.Error(sctx.Ctx, "Failed to reject surrender request: %v", err)
		return nil, fmt.Errorf("failed to reject surrender request")
	}

	// Complete approval task
	task, err := h.approvalRepo.FindTaskBySurrenderRequestID(sctx.Ctx, surrenderRequestID)
	if err == nil {
		_, err = h.approvalRepo.CompleteTask(sctx.Ctx, task.ID, userID)
		if err != nil {
			log.Error(sctx.Ctx, "Failed to complete approval task: %v", err)
		}
	}

	log.Info(sctx.Ctx, "Rejected surrender request %s, reason: %s", rejected.RequestNumber, req.RejectionReason)

	return &RejectSurrenderResponse{
		StatusCodeAndMessage: port.UpdateSuccess,
		Data: RejectSurrenderData{
			SurrenderRequestID: rejected.ID.String(),
			RequestNumber:      rejected.RequestNumber,
			PolicyID:           rejected.PolicyID,
			PolicyNumber:       h.getMetadataString(rejected.Metadata, "policy_number"),
			OldStatus:          string(oldStatus),
			NewStatus:          string(rejected.Status),
			RejectedBy:         userID.String(),
			RejectedDate:       time.Now().Format(time.RFC3339),
			RejectionReason:    req.RejectionReason,
			Message:            "Surrender request rejected. Customer will be notified.",
		},
	}, nil
}

// RecalculateSurrenderValue recalculates surrender value during approval
// POST /v1/approval/recalculate
// Business Rule: BR-APPR-007
// Recalculates value with fresh data before approval
func (h *ApprovalHandler) RecalculateSurrenderValue(sctx *serverRoute.Context, req RecalculateRequest) (interface{}, error) {
	surrenderRequestID, err := uuid.Parse(req.SurrenderRequestID)
	if err != nil {
		log.Error(sctx.Ctx, "Invalid surrender request ID: %v", err)
		return nil, fmt.Errorf("invalid surrender request ID format")
	}

	// Mock user ID
	userID := uuid.New()
	if req.UserID != "" {
		parsedUserID, err := uuid.Parse(req.UserID)
		if err == nil {
			userID = parsedUserID
		}
	}

	// Get surrender request
	surrenderRequest, err := h.surrenderRepo.FindByID(sctx.Ctx, surrenderRequestID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("surrender request not found")
		}
		log.Error(sctx.Ctx, "Failed to get surrender request: %v", err)
		return nil, err
	}

	// Get fresh policy data
	policy, err := h.policyService.GetPolicyByID(sctx.Ctx, surrenderRequest.PolicyID)
	if err != nil {
		log.Error(sctx.Ctx, "Failed to get policy: %v", err)
		return nil, fmt.Errorf("failed to retrieve policy data")
	}

	// Store old values
	oldGSV := surrenderRequest.GrossSurrenderValue
	oldNSV := surrenderRequest.NetSurrenderValue

	// Recalculate surrender value (simplified calculation)
	paidUpValue := (policy.SumAssured * float64(policy.PremiumsPaid)) / float64(policy.TotalPremiums)
	totalBonus := (policy.SumAssured / 1000) * 50.0 * float64(policy.PremiumsPaid)
	surrenderFactor := 0.75
	grossValue := (paidUpValue + totalBonus) * surrenderFactor

	// Get fresh deductions
	unpaidPremiums := 0.0 // Placeholder
	loanAmount := 0.0     // Placeholder

	netValue := grossValue - unpaidPremiums - loanAmount

	// BR-APPR-007: Update surrender request with recalculated values
	recalculated, err := h.surrenderRepo.RecalculateSurrenderValue(sctx.Ctx, surrenderRequestID, grossValue, netValue, paidUpValue, nil, loanAmount, unpaidPremiums)
	if err != nil {
		log.Error(sctx.Ctx, "Failed to recalculate surrender value: %v", err)
		return nil, fmt.Errorf("failed to recalculate surrender value")
	}

	log.Info(sctx.Ctx, "Recalculated surrender value for %s: GSV %.2f -> %.2f, NSV %.2f -> %.2f",
		recalculated.RequestNumber, oldGSV, grossValue, oldNSV, netValue)

	return &RecalculateResponse{
		StatusCodeAndMessage: port.UpdateSuccess,
		Data: RecalculateData{
			SurrenderRequestID: recalculated.ID.String(),
			RequestNumber:      recalculated.RequestNumber,
			RecalculationDate:  recalculated.UpdatedAt.Format(time.RFC3339),
			OldValues: ValueComparisonData{
				GrossSurrenderValue: oldGSV,
				NetSurrenderValue:   oldNSV,
			},
			NewValues: ValueComparisonData{
				GrossSurrenderValue: recalculated.GrossSurrenderValue,
				NetSurrenderValue:   recalculated.NetSurrenderValue,
			},
			Difference: DifferenceData{
				GrossDifference: recalculated.GrossSurrenderValue - oldGSV,
				NetDifference:   recalculated.NetSurrenderValue - oldNSV,
				PercentChange:   ((recalculated.NetSurrenderValue - oldNSV) / oldNSV) * 100,
			},
			RecalculatedBy: userID.String(),
			Reason:         req.RecalculationReason,
		},
	}, nil
}

// EscalateApprovalTask escalates a task to higher authority
// POST /v1/approval/escalate
// Business Rule: BR-APPR-008
// Escalates task when SLA breached or complex case
func (h *ApprovalHandler) EscalateApprovalTask(sctx *serverRoute.Context, req EscalateTaskRequest) (interface{}, error) {
	taskID, err := uuid.Parse(req.TaskID)
	if err != nil {
		log.Error(sctx.Ctx, "Invalid task ID: %v", err)
		return nil, fmt.Errorf("invalid task ID format")
	}

	// Mock user IDs
	currentUserID := uuid.New()
	escalatedToUserID := uuid.New()

	if req.EscalatedBy != "" {
		parsedUserID, err := uuid.Parse(req.EscalatedBy)
		if err == nil {
			currentUserID = parsedUserID
		}
	}

	if req.EscalateTo != "" {
		parsedUserID, err := uuid.Parse(req.EscalateTo)
		if err == nil {
			escalatedToUserID = parsedUserID
		}
	}

	// Get task details
	task, err := h.approvalRepo.FindByID(sctx.Ctx, taskID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("task not found")
		}
		log.Error(sctx.Ctx, "Failed to get task: %v", err)
		return nil, err
	}

	// BR-APPR-008: Escalate task with higher priority
	newPriority := domain.ApprovalTaskPriorityHigh
	if task.Priority == domain.ApprovalTaskPriorityHigh {
		newPriority = domain.ApprovalTaskPriorityCritical
	}

	escalated, err := h.approvalRepo.EscalateTask(sctx.Ctx, taskID, escalatedToUserID, req.EscalationReason)
	if err != nil {
		log.Error(sctx.Ctx, "Failed to escalate task: %v", err)
		return nil, fmt.Errorf("failed to escalate task")
	}

	// Get surrender request details
	surrenderRequest, err := h.surrenderRepo.FindByID(sctx.Ctx, task.SurrenderRequestID)
	if err != nil {
		log.Error(sctx.Ctx, "Failed to get surrender request: %v", err)
		return nil, err
	}

	log.Info(sctx.Ctx, "Escalated task %s from %s to %s with priority %s",
		escalated.ID, currentUserID, escalatedToUserID, newPriority)

	return &EscalateTaskResponse{
		StatusCodeAndMessage: port.UpdateSuccess,
		Data: EscalateTaskData{
			TaskID:             escalated.ID.String(),
			SurrenderRequestID: escalated.SurrenderRequestID.String(),
			RequestNumber:      surrenderRequest.RequestNumber,
			OldPriority:        string(task.Priority),
			NewPriority:        string(escalated.Priority),
			EscalatedBy:        currentUserID.String(),
			EscalatedTo:        escalatedToUserID.String(),
			EscalatedDate:      time.Now().Format(time.RFC3339),
			EscalationReason:   req.EscalationReason,
			Message:            fmt.Sprintf("Task escalated to %s with %s priority", h.getUserName(&escalatedToUserID), newPriority),
		},
	}, nil
}

// GetApprovalHistory retrieves approval history for a surrender request
// GET /v1/approval/history
// Returns complete audit trail of approval workflow
func (h *ApprovalHandler) GetApprovalHistory(sctx *serverRoute.Context, req ApprovalHistoryParams) (interface{}, error) {
	surrenderRequestID, err := uuid.Parse(req.SurrenderRequestID)
	if err != nil {
		log.Error(sctx.Ctx, "Invalid surrender request ID: %v", err)
		return nil, fmt.Errorf("invalid surrender request ID format")
	}

	// Get surrender request
	surrenderRequest, err := h.surrenderRepo.FindByID(sctx.Ctx, surrenderRequestID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("surrender request not found")
		}
		log.Error(sctx.Ctx, "Failed to get surrender request: %v", err)
		return nil, err
	}

	// Get approval history
	// TODO: Implement GetApprovalHistory in ApprovalWorkflowRepository
	history := []struct {
		ID               uuid.UUID
		Status           string
		StatusChangedBy  uuid.UUID
		StatusChangeDate time.Time
		Comments         *string
		CreatedAt        time.Time
	}{}
	_ = surrenderRequestID // placeholder to avoid unused variable error

	// Build history items
	historyItems := make([]ApprovalHistoryItem, 0, len(history))
	for _, record := range history {
		item := ApprovalHistoryItem{
			HistoryID:        record.ID.String(),
			Status:           record.Status,
			StatusChangedBy:  h.getUserName(&record.StatusChangedBy),
			StatusChangeDate: record.StatusChangeDate.Format(time.RFC3339),
			Comments:         h.getStringValue(record.Comments),
			Duration:         h.calculateDuration(record.CreatedAt, record.StatusChangeDate),
		}
		historyItems = append(historyItems, item)
	}

	log.Info(sctx.Ctx, "Retrieved %d history records for surrender request %s", len(historyItems), surrenderRequest.RequestNumber)

	return &ApprovalHistoryResponse{
		StatusCodeAndMessage: port.GetSuccess,
		Data: ApprovalHistoryData{
			SurrenderRequestID: surrenderRequest.ID.String(),
			RequestNumber:      surrenderRequest.RequestNumber,
			PolicyNumber:       h.getMetadataString(surrenderRequest.Metadata, "policy_number"),
			CurrentStatus:      string(surrenderRequest.Status),
			TotalRecords:       len(historyItems),
			History:            historyItems,
		},
	}, nil
}

// Helper functions

func (h *ApprovalHandler) getMetadataString(metadata map[string]interface{}, key string) string {
	if val, ok := metadata[key].(string); ok {
		return val
	}
	return ""
}

func (h *ApprovalHandler) getUserName(userID *uuid.UUID) string {
	if userID == nil {
		return "System"
	}
	// In production, fetch from user service
	return userID.String()
}

func (h *ApprovalHandler) formatTime(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(time.RFC3339)
}

func (h *ApprovalHandler) getStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func (h *ApprovalHandler) getPrescribedLimit(policyID string) float64 {
	// Placeholder - get from configuration
	return 2000.0
}

func (h *ApprovalHandler) calculateDuration(start, end time.Time) string {
	duration := end.Sub(start)
	hours := int(duration.Hours())
	days := hours / 24
	remainingHours := hours % 24

	if days > 0 {
		return fmt.Sprintf("%d days, %d hours", days, remainingHours)
	}
	return fmt.Sprintf("%d hours", hours)
}

// ============================================
// Request/Response Types
// ============================================

type ApprovalQueueResponse struct {
	port.StatusCodeAndMessage
	Data ApprovalQueueData `json:"data"`
}

type ApprovalQueueData struct {
	Items      []ApprovalQueueItem `json:"items"`
	TotalCount int                 `json:"total_count"`
	Page       int                 `json:"page"`
	Limit      int                 `json:"limit"`
	TotalPages int                 `json:"total_pages"`
}

type ApprovalQueueItem struct {
	TaskID             string  `json:"task_id"`
	SurrenderRequestID string  `json:"surrender_request_id"`
	RequestNumber      string  `json:"request_number"`
	PolicyID           string  `json:"policy_id"`
	PolicyNumber       string  `json:"policy_number"`
	PolicyholderName   string  `json:"policyholder_name"`
	RequestType        string  `json:"request_type"`
	RequestDate        string  `json:"request_date"`
	NetSurrenderValue  float64 `json:"net_surrender_value"`
	SurrenderStatus    string  `json:"surrender_status"`
	TaskStatus         string  `json:"task_status"`
	Priority           string  `json:"priority"`
	AssignedTo         string  `json:"assigned_to"`
	AssignedDate       string  `json:"assigned_date,omitempty"`
	DueDate            string  `json:"due_date,omitempty"`
	DaysInQueue        int     `json:"days_in_queue"`
}

type ReserveTaskResponse struct {
	port.StatusCodeAndMessage
	Data ReserveTaskData `json:"data"`
}

type ReserveTaskData struct {
	TaskID             string `json:"task_id"`
	SurrenderRequestID string `json:"surrender_request_id"`
	RequestNumber      string `json:"request_number"`
	PolicyNumber       string `json:"policy_number"`
	Status             string `json:"status"`
	AssignedTo         string `json:"assigned_to"`
	AssignedDate       string `json:"assigned_date"`
	DueDate            string `json:"due_date"`
	Message            string `json:"message"`
}

type ReleaseTaskResponse struct {
	port.StatusCodeAndMessage
	Data ReleaseTaskData `json:"data"`
}

type ReleaseTaskData struct {
	TaskID     string `json:"task_id"`
	Status     string `json:"status"`
	ReleasedBy string `json:"released_by"`
	ReleasedAt string `json:"released_at"`
	Message    string `json:"message"`
}

type ApproveSurrenderResponse struct {
	port.StatusCodeAndMessage
	Data ApproveSurrenderData `json:"data"`
}

type ApproveSurrenderData struct {
	SurrenderRequestID string            `json:"surrender_request_id"`
	RequestNumber      string            `json:"request_number"`
	PolicyID           string            `json:"policy_id"`
	PolicyNumber       string            `json:"policy_number"`
	OldStatus          string            `json:"old_status"`
	NewStatus          string            `json:"new_status"`
	NewPolicyStatus    string            `json:"new_policy_status"`
	ApprovedBy         string            `json:"approved_by"`
	ApprovedDate       string            `json:"approved_date"`
	ApprovalComments   string            `json:"approval_comments"`
	NetSurrenderValue  float64           `json:"net_surrender_value"`
	WorkflowState      WorkflowStateData `json:"workflow_state"`
	NextAction         NextActionData    `json:"next_action"`
}

type RejectSurrenderResponse struct {
	port.StatusCodeAndMessage
	Data RejectSurrenderData `json:"data"`
}

type RejectSurrenderData struct {
	SurrenderRequestID string `json:"surrender_request_id"`
	RequestNumber      string `json:"request_number"`
	PolicyID           string `json:"policy_id"`
	PolicyNumber       string `json:"policy_number"`
	OldStatus          string `json:"old_status"`
	NewStatus          string `json:"new_status"`
	RejectedBy         string `json:"rejected_by"`
	RejectedDate       string `json:"rejected_date"`
	RejectionReason    string `json:"rejection_reason"`
	Message            string `json:"message"`
}

type RecalculateResponse struct {
	port.StatusCodeAndMessage
	Data RecalculateData `json:"data"`
}

type RecalculateData struct {
	SurrenderRequestID string              `json:"surrender_request_id"`
	RequestNumber      string              `json:"request_number"`
	RecalculationDate  string              `json:"recalculation_date"`
	OldValues          ValueComparisonData `json:"old_values"`
	NewValues          ValueComparisonData `json:"new_values"`
	Difference         DifferenceData      `json:"difference"`
	RecalculatedBy     string              `json:"recalculated_by"`
	Reason             string              `json:"reason"`
}

type ValueComparisonData struct {
	GrossSurrenderValue float64 `json:"gross_surrender_value"`
	NetSurrenderValue   float64 `json:"net_surrender_value"`
}

type DifferenceData struct {
	GrossDifference float64 `json:"gross_difference"`
	NetDifference   float64 `json:"net_difference"`
	PercentChange   float64 `json:"percent_change"`
}

type EscalateTaskResponse struct {
	port.StatusCodeAndMessage
	Data EscalateTaskData `json:"data"`
}

type EscalateTaskData struct {
	TaskID             string `json:"task_id"`
	SurrenderRequestID string `json:"surrender_request_id"`
	RequestNumber      string `json:"request_number"`
	OldPriority        string `json:"old_priority"`
	NewPriority        string `json:"new_priority"`
	EscalatedBy        string `json:"escalated_by"`
	EscalatedTo        string `json:"escalated_to"`
	EscalatedDate      string `json:"escalated_date"`
	EscalationReason   string `json:"escalation_reason"`
	Message            string `json:"message"`
}

type ApprovalHistoryResponse struct {
	port.StatusCodeAndMessage
	Data ApprovalHistoryData `json:"data"`
}

type ApprovalHistoryData struct {
	SurrenderRequestID string                `json:"surrender_request_id"`
	RequestNumber      string                `json:"request_number"`
	PolicyNumber       string                `json:"policy_number"`
	CurrentStatus      string                `json:"current_status"`
	TotalRecords       int                   `json:"total_records"`
	History            []ApprovalHistoryItem `json:"history"`
}

type ApprovalHistoryItem struct {
	HistoryID        string `json:"history_id"`
	Status           string `json:"status"`
	StatusChangedBy  string `json:"status_changed_by"`
	StatusChangeDate string `json:"status_change_date"`
	Comments         string `json:"comments"`
	Duration         string `json:"duration"`
}

type WorkflowStateData struct {
	CurrentStage    string   `json:"current_stage"`
	CompletedStages []string `json:"completed_stages"`
	PendingStages   []string `json:"pending_stages"`
	ProgressPercent int      `json:"progress_percent"`
}

type NextActionData struct {
	Action      string `json:"action"`
	Description string `json:"description"`
	URL         string `json:"url"`
}

// Query parameter types
type ApprovalQueueParams struct {
	Status      string `form:"status"`
	RequestType string `form:"request_type"`
	TaskStatus  string `form:"task_status"`
	OfficeCode  string `form:"office_code"`
	Page        int    `form:"page"`
	Limit       int    `form:"limit"`
}

type ApprovalHistoryParams struct {
	SurrenderRequestID string `form:"surrender_request_id" validate:"required"`
}

// ============================================
// Additional External Service Interfaces
// ============================================

type UserServiceInterface interface {
	GetUserByID(ctx interface{}, userID string) (*UserInfo, error)
	GetUsersByRole(ctx interface{}, role string) ([]*UserInfo, error)
}

type UserInfo struct {
	ID         string
	Name       string
	Email      string
	Role       string
	Department string
}

// Mock implementations

func NewMockUserService() UserServiceInterface {
	return &MockUserService{}
}

type MockUserService struct{}

func (m *MockUserService) GetUserByID(ctx interface{}, userID string) (*UserInfo, error) {
	return &UserInfo{
		ID:         userID,
		Name:       "CPC Officer",
		Email:      "officer@cept.gov.in",
		Role:       "APPROVER",
		Department: "Claims Processing Center",
	}, nil
}

func (m *MockUserService) GetUsersByRole(ctx interface{}, role string) ([]*UserInfo, error) {
	return []*UserInfo{
		{
			ID:         uuid.New().String(),
			Name:       "Senior Officer 1",
			Email:      "senior1@cept.gov.in",
			Role:       role,
			Department: "Claims Processing Center",
		},
	}, nil
}
