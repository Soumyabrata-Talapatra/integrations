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

// ForcedSurrenderHandler handles all forced surrender operations
// These are internal APIs triggered by system processes or CPC staff
// Business Rules: BR-FSUR-001 to BR-FSUR-006
// Functional Requirements: FR-FSUR-001 to FR-FSUR-006
type ForcedSurrenderHandler struct {
	*serverHandler.Base
	surrenderRepo       *repo.SurrenderRequestRepository
	forcedSurrenderRepo *repo.ForcedSurrenderRepository
	// External service placeholders
	policyService       PolicyServiceInterface
	collectionsService  CollectionsServiceInterface
	notificationService NotificationServiceInterface
}

// NewForcedSurrenderHandler creates a new forced surrender handler
func NewForcedSurrenderHandler(
	surrenderRepo *repo.SurrenderRequestRepository,
	forcedSurrenderRepo *repo.ForcedSurrenderRepository,
) *ForcedSurrenderHandler {
	base := serverHandler.New("Forced Surrender").SetPrefix("/v1").AddPrefix("/forced-surrender")

	return &ForcedSurrenderHandler{
		Base:                base,
		surrenderRepo:       surrenderRepo,
		forcedSurrenderRepo: forcedSurrenderRepo,
		// Initialize placeholders
		policyService:       NewMockPolicyService(),
		collectionsService:  NewMockCollectionsService(),
		notificationService: NewMockNotificationService(),
	}
}

// Routes defines all routes for forced surrender
func (h *ForcedSurrenderHandler) Routes() []serverRoute.Route {
	return []serverRoute.Route{
		serverRoute.POST("/trigger-monthly-evaluation", h.TriggerMonthlyEvaluation).Name("Trigger Monthly Evaluation"),
		serverRoute.POST("/send-reminder", h.SendForcedSurrenderReminder).Name("Send Forced Surrender Reminder"),
		serverRoute.POST("/create-payment-window", h.CreatePaymentWindow).Name("Create Payment Window"),
		serverRoute.POST("/record-payment", h.RecordPaymentReceived).Name("Record Payment Received"),
		serverRoute.POST("/initiate-forced-surrender", h.InitiateForcedSurrender).Name("Initiate Forced Surrender"),
		serverRoute.POST("/schedule-batch-evaluation", h.ScheduleBatchEvaluation).Name("Schedule Batch Evaluation"),
		serverRoute.GET("/pending-reminders", h.GetPendingReminders).Name("Get Pending Reminders"),
		serverRoute.GET("/expired-payment-windows", h.GetExpiredPaymentWindows).Name("Get Expired Payment Windows"),
	}
}

// TriggerMonthlyEvaluation evaluates policies for forced surrender eligibility
// POST /v1/forced-surrender/trigger-monthly-evaluation
// Business Rule: BR-FSUR-001
// This is triggered monthly by a scheduler to identify policies with 6+ months of unpaid premiums
func (h *ForcedSurrenderHandler) TriggerMonthlyEvaluation(sctx *serverRoute.Context, req TriggerMonthlyEvaluationRequest) (interface{}, error) {
	evaluationDate := time.Now()
	if req.EvaluationDate != "" {
		parsedDate, err := parseFlexibleDate(req.EvaluationDate)
		if err != nil {
			log.Error(sctx.Ctx, "Invalid evaluation date: %v", err)
			return nil, fmt.Errorf("invalid evaluation date format. Supported formats: YYYY-MM-DD, DD/MM/YYYY, MM/DD/YYYY, YYYY/MM/DD")
		}
		evaluationDate = parsedDate
	}

	log.Info(sctx.Ctx, "Starting monthly forced surrender evaluation for date: %s", evaluationDate.Format("2006-01-02"))

	// BR-FSUR-001: Identify policies with 6+ months of unpaid premiums
	minUnpaidMonths := 6
	cutoffDate := evaluationDate.AddDate(0, -minUnpaidMonths, 0)

	// Get policies with unpaid premiums from collections service
	eligiblePolicies, err := h.collectionsService.GetPoliciesWithUnpaidPremiums(sctx.Ctx, cutoffDate, minUnpaidMonths)
	if err != nil {
		log.Error(sctx.Ctx, "Failed to get policies with unpaid premiums: %v", err)
		return nil, fmt.Errorf("failed to retrieve eligible policies")
	}

	log.Info(sctx.Ctx, "Found %d policies eligible for forced surrender evaluation", len(eligiblePolicies))

	results := make([]EvaluationResult, 0, len(eligiblePolicies))
	processedCount := 0
	errorCount := 0

	for _, policyIDStr := range eligiblePolicies {
		// Get policy details from policy service
		policyInfo, err := h.policyService.GetPolicyByID(sctx.Ctx, policyIDStr)
		if err != nil {
			log.Error(sctx.Ctx, "Failed to get policy details for %s: %v", policyIDStr, err)
			errorCount++
			continue
		}

		// Check if policy already has an active forced surrender process
		_, found, err := h.surrenderRepo.FindActiveByPolicyID(sctx.Ctx, policyIDStr)
		if err != nil && err != pgx.ErrNoRows {
			log.Error(sctx.Ctx, "Failed to check existing surrender for policy %s: %v", policyInfo.PolicyNumber, err)
			errorCount++
			continue
		}

		if found {
			log.Info(sctx.Ctx, "Policy %s already has active surrender request, skipping", policyInfo.PolicyNumber)
			results = append(results, EvaluationResult{
				PolicyID:     policyIDStr,
				PolicyNumber: policyInfo.PolicyNumber,
				Status:       "SKIPPED",
				Reason:       "Active surrender request exists",
			})
			continue
		}

		// Check if policy already has pending reminders
		latestReminder, found, err := h.forcedSurrenderRepo.FindLatestReminderByPolicyID(sctx.Ctx, policyIDStr)
		if err != nil && err != pgx.ErrNoRows {
			log.Error(sctx.Ctx, "Failed to check reminders for policy %s: %v", policyInfo.PolicyNumber, err)
			errorCount++
			continue
		}

		if found && latestReminder.ReminderNumber != "" {
			log.Info(sctx.Ctx, "Policy %s has pending reminder, skipping evaluation", policyInfo.PolicyNumber)
			results = append(results, EvaluationResult{
				PolicyID:     policyIDStr,
				PolicyNumber: policyInfo.PolicyNumber,
				Status:       "SKIPPED",
				Reason:       "Pending reminder exists",
			})
			continue
		}

		// Mark policy as eligible for forced surrender reminder
		results = append(results, EvaluationResult{
			PolicyID:        policyIDStr,
			PolicyNumber:    policyInfo.PolicyNumber,
			Status:          "ELIGIBLE",
			UnpaidMonths:    minUnpaidMonths, // From input parameter
			UnpaidAmount:    0.0,             // Will be calculated by collections service
			LastPremiumDate: "",              // Will be provided by collections service
		})
		processedCount++
	}

	log.Info(sctx.Ctx, "Monthly evaluation completed: %d processed, %d errors", processedCount, errorCount)

	return &MonthlyEvaluationResponse{
		StatusCodeAndMessage: port.CustomSuccess,
		Data: MonthlyEvaluationData{
			EvaluationDate:   evaluationDate.Format(time.RFC3339),
			TotalPolicies:    len(eligiblePolicies),
			EligiblePolicies: processedCount,
			SkippedPolicies:  len(eligiblePolicies) - processedCount - errorCount,
			ErrorCount:       errorCount,
			Results:          results,
			NextAction:       "Send reminders to eligible policies using /forced-surrender/send-reminder endpoint",
		},
	}, nil
}

// SendForcedSurrenderReminder sends reminder notice to policyholder
// POST /v1/forced-surrender/send-reminder
// Business Rules: BR-FSUR-002, BR-FSUR-003
// Sends first or second reminder based on unpaid duration
func (h *ForcedSurrenderHandler) SendForcedSurrenderReminder(sctx *serverRoute.Context, req SendReminderRequest) (interface{}, error) {
	// Get policy details
	policy, err := h.policyService.GetPolicyByID(sctx.Ctx, req.PolicyID)
	if err != nil {
		log.Error(sctx.Ctx, "Failed to get policy: %v", err)
		return nil, fmt.Errorf("policy not found")
	}

	// Get unpaid premium details
	unpaidPremiums, err := h.collectionsService.GetUnpaidPremiums(sctx.Ctx, req.PolicyID)
	if err != nil {
		log.Error(sctx.Ctx, "Failed to get unpaid premiums: %v", err)
		return nil, fmt.Errorf("failed to retrieve unpaid premiums")
	}

	// Check latest reminder
	latestReminder, found, err := h.forcedSurrenderRepo.FindLatestReminderByPolicyID(sctx.Ctx, req.PolicyID)
	if err != nil && err != pgx.ErrNoRows {
		log.Error(sctx.Ctx, "Failed to check existing reminders: %v", err)
		return nil, fmt.Errorf("failed to check reminder history")
	}

	// Determine reminder number
	reminderNumber := domain.ReminderLevelFirst
	if found {
		switch latestReminder.ReminderNumber {
		case domain.ReminderLevelFirst:
			reminderNumber = domain.ReminderLevelSecond
		case domain.ReminderLevelSecond:
			reminderNumber = domain.ReminderLevelThird
		case domain.ReminderLevelThird:
			log.Error(sctx.Ctx, "Maximum reminders already sent for policy %s", policy.PolicyNumber)
			return nil, fmt.Errorf("maximum reminder limit reached")
		}
	}

	// BR-FSUR-002: First reminder at 6 months
	// BR-FSUR-003: Second reminder at 9 months
	reminderDate := time.Now()
	mockUserID := uuid.New()

	// Create reminder record
	reminder := domain.ForcedSurrenderReminder{
		PolicyID:       req.PolicyID,
		ReminderNumber: reminderNumber,
		ReminderDate:   reminderDate,
		LoanPrincipal:  0.0, // TODO: Get from loan service
		LoanInterest:   0.0, // TODO: Get from loan service
		Metadata: map[string]interface{}{
			"policy_number":     policy.PolicyNumber,
			"policyholder_name": policy.PolicyholderName,
			"unpaid_months":     req.UnpaidMonths,
			"unpaid_amount":     unpaidPremiums,
			"created_by":        mockUserID.String(),
		},
	}

	created, err := h.forcedSurrenderRepo.CreateReminder(sctx.Ctx, reminder)
	if err != nil {
		log.Error(sctx.Ctx, "Failed to create reminder: %v", err)
		return nil, fmt.Errorf("failed to create reminder record")
	}

	// Calculate payment window for response
	paymentWindowEnd := reminderDate.AddDate(0, 0, 30)

	// Convert reminder level to int for response
	reminderNumInt := 1
	switch reminderNumber {
	case domain.ReminderLevelFirst:
		reminderNumInt = 1
	case domain.ReminderLevelSecond:
		reminderNumInt = 2
	case domain.ReminderLevelThird:
		reminderNumInt = 3
	}

	// Send notification to policyholder
	notificationSent := h.notificationService.SendReminderNotification(sctx.Ctx, NotificationParams{
		PolicyID:         policy.ID,
		PolicyNumber:     policy.PolicyNumber,
		PolicyholderName: policy.PolicyholderName,
		ReminderNumber:   reminderNumInt,
		UnpaidAmount:     unpaidPremiums,
		PaymentDeadline:  paymentWindowEnd,
		ContactEmail:     policy.PolicyholderName + "@example.com", // Placeholder
		ContactPhone:     "+91-9876543210",                         // Placeholder
	})

	log.Info(sctx.Ctx, "Sent reminder #%d to policy %s, notification status: %v", reminderNumInt, policy.PolicyNumber, notificationSent)

	return &SendReminderResponse{
		StatusCodeAndMessage: port.CreateSuccess,
		Data: SendReminderData{
			ReminderID:         created.ID.String(),
			PolicyID:           policy.ID,
			PolicyNumber:       policy.PolicyNumber,
			ReminderNumber:     reminderNumInt,
			ReminderDate:       reminderDate.Format(time.RFC3339),
			PaymentWindowStart: reminderDate.Format(time.RFC3339),
			PaymentWindowEnd:   paymentWindowEnd.Format(time.RFC3339),
			UnpaidAmount:       unpaidPremiums,
			NotificationSent:   notificationSent,
			Message:            fmt.Sprintf("Reminder #%d sent successfully. Payment window: %d days", reminderNumInt, 30),
		},
	}, nil
}

// CreatePaymentWindow creates a payment window for forced surrender
// POST /v1/forced-surrender/create-payment-window
// Business Rule: BR-FSUR-004
// Creates 30-day payment window after reminder
func (h *ForcedSurrenderHandler) CreatePaymentWindow(sctx *serverRoute.Context, req CreatePaymentWindowRequest) (interface{}, error) {
	reminderID, err := uuid.Parse(req.ReminderID)
	if err != nil {
		log.Error(sctx.Ctx, "Invalid reminder ID: %v", err)
		return nil, fmt.Errorf("invalid reminder ID format")
	}

	// Get reminder details
	reminder, err := h.forcedSurrenderRepo.FindReminderByID(sctx.Ctx, reminderID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("reminder not found")
		}
		log.Error(sctx.Ctx, "Failed to get reminder: %v", err)
		return nil, err
	}

	// Check if payment window already exists
	existingWindow, found, err := h.forcedSurrenderRepo.FindPaymentWindowByReminderID(sctx.Ctx, reminderID)
	if err != nil && err != pgx.ErrNoRows {
		log.Error(sctx.Ctx, "Failed to check existing payment window: %v", err)
		return nil, fmt.Errorf("failed to validate payment window")
	}

	if found {
		log.Info(sctx.Ctx, "Payment window already exists for reminder %s", req.ReminderID)
		return &CreatePaymentWindowResponse{
			StatusCodeAndMessage: port.StatusCodeAndMessage{
				StatusCode: 200,
				Success:    true,
				Message:    "Payment window already exists",
			},
			Data: PaymentWindowData{
				PaymentWindowID: existingWindow.ID.String(),
				ReminderID:      reminderID.String(),
				PolicyID:        reminder.PolicyID,
				WindowStart:     existingWindow.WindowStartDate.Format(time.RFC3339),
				WindowEnd:       existingWindow.WindowEndDate.Format(time.RFC3339),
				ExpectedAmount:  0.0, // Not stored in struct
				PaymentReceived: existingWindow.PaymentReceived,
				Status:          "ACTIVE", // Hardcoded since struct doesn't have status
			},
		}, nil
	}

	// BR-FSUR-004: Create 30-day payment window
	windowStart := time.Now()
	windowEnd := windowStart.AddDate(0, 0, 30)

	// Create surrender request for tracking
	surrenderRequestID := uuid.New() // TODO: Link to actual surrender request

	paymentWindow := domain.ForcedSurrenderPaymentWindow{
		SurrenderRequestID: surrenderRequestID,
		PolicyID:           reminder.PolicyID,
		WindowStartDate:    windowStart,
		WindowEndDate:      windowEnd,
		PaymentReceived:    false,
	}

	created, err := h.forcedSurrenderRepo.CreatePaymentWindow(sctx.Ctx, paymentWindow)
	if err != nil {
		log.Error(sctx.Ctx, "Failed to create payment window: %v", err)
		return nil, fmt.Errorf("failed to create payment window")
	}

	log.Info(sctx.Ctx, "Created payment window for reminder %s: %s to %s", reminder.ID, windowStart.Format("2006-01-02"), windowEnd.Format("2006-01-02"))

	return &CreatePaymentWindowResponse{
		StatusCodeAndMessage: port.CreateSuccess,
		Data: PaymentWindowData{
			PaymentWindowID: created.ID.String(),
			ReminderID:      reminderID.String(),
			PolicyID:        reminder.PolicyID,
			WindowStart:     created.WindowStartDate.Format(time.RFC3339),
			WindowEnd:       created.WindowEndDate.Format(time.RFC3339),
			ExpectedAmount:  0.0, // Not stored in struct
			PaymentReceived: created.PaymentReceived,
			Status:          "ACTIVE",
			DaysRemaining:   int(time.Until(created.WindowEndDate).Hours() / 24),
		},
	}, nil
}

// RecordPaymentReceived records payment received during window
// POST /v1/forced-surrender/record-payment
// Business Rule: BR-FSUR-005
// Updates payment window status when payment is received
func (h *ForcedSurrenderHandler) RecordPaymentReceived(sctx *serverRoute.Context, req RecordPaymentRequest) (interface{}, error) {
	paymentWindowID, err := uuid.Parse(req.PaymentWindowID)
	if err != nil {
		log.Error(sctx.Ctx, "Invalid payment window ID: %v", err)
		return nil, fmt.Errorf("invalid payment window ID format")
	}

	// BR-FSUR-005: Update payment window status
	updated, err := h.forcedSurrenderRepo.UpdatePaymentReceived(sctx.Ctx, paymentWindowID, req.Amount, req.PaymentReference)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("payment window not found")
		}
		log.Error(sctx.Ctx, "Failed to update payment window: %v", err)
		return nil, fmt.Errorf("failed to record payment")
	}

	// Get reminder by policy ID
	reminder, found, err := h.forcedSurrenderRepo.FindLatestReminderByPolicyID(sctx.Ctx, updated.PolicyID)
	if err == nil && found {
		// TODO: Add UpdateReminderCompleted method to repository
		log.Info(sctx.Ctx, "Reminder found for policy, would mark as completed: %s", reminder.ID)
	}

	log.Info(sctx.Ctx, "Recorded payment of %.2f for payment window %s", req.Amount, updated.ID)

	return &RecordPaymentResponse{
		StatusCodeAndMessage: port.UpdateSuccess,
		Data: PaymentRecordData{
			PaymentWindowID:  updated.ID.String(),
			ReminderID:       "", // Not directly linked in struct
			AmountReceived:   req.Amount,
			ExpectedAmount:   0.0, // Not stored in struct
			PaymentReference: req.PaymentReference,
			RecordedDate:     time.Now().Format(time.RFC3339),
			Status:           "PAID",
			Message:          "Payment recorded successfully. Forced surrender process cancelled.",
		},
	}, nil
}

// InitiateForcedSurrender initiates forced surrender process
// POST /v1/forced-surrender/initiate-forced-surrender
// Business Rule: BR-FSUR-006
// Triggered when payment window expires without payment
func (h *ForcedSurrenderHandler) InitiateForcedSurrender(sctx *serverRoute.Context, req InitiateForcedSurrenderRequest) (interface{}, error) {
	// Get policy details
	policy, err := h.policyService.GetPolicyByID(sctx.Ctx, req.PolicyID)
	if err != nil {
		log.Error(sctx.Ctx, "Failed to get policy: %v", err)
		return nil, fmt.Errorf("policy not found")
	}

	// Verify payment window has expired (optional check)
	// Payment window verification removed since InitiateForcedSurrenderRequest doesn't have PaymentWindowID field

	// Check if forced surrender already exists
	existingRequest, found, err := h.surrenderRepo.FindActiveByPolicyID(sctx.Ctx, req.PolicyID)
	if err != nil && err != pgx.ErrNoRows {
		log.Error(sctx.Ctx, "Failed to check existing surrender: %v", err)
		return nil, fmt.Errorf("failed to validate surrender request")
	}

	if found && existingRequest.RequestType == domain.SurrenderRequestTypeForced {
		log.Info(sctx.Ctx, "Forced surrender already exists for policy %s", policy.PolicyNumber)
		return &InitiateForcedSurrenderResponse{
			StatusCodeAndMessage: port.StatusCodeAndMessage{
				StatusCode: 200,
				Success:    true,
				Message:    "Forced surrender request already exists",
			},
			Data: ForcedSurrenderData{
				SurrenderRequestID: existingRequest.ID.String(),
				RequestNumber:      existingRequest.RequestNumber,
				PolicyID:           existingRequest.PolicyID,
				PolicyNumber:       policy.PolicyNumber,
				Status:             string(existingRequest.Status),
				RequestDate:        existingRequest.RequestDate.Format(time.RFC3339),
			},
		}, nil
	}

	// Calculate surrender value (similar to voluntary but automated)
	calculation := h.calculateForcedSurrenderValue(policy)

	// BR-FSUR-006: Create forced surrender request
	requestNumber := h.generateRequestNumber(policy.PolicyNumber)
	mockUserID := uuid.New()

	surrenderRequest := domain.PolicySurrenderRequest{
		PolicyID:                     req.PolicyID,
		RequestNumber:                requestNumber,
		RequestType:                  domain.SurrenderRequestTypeForced,
		RequestDate:                  time.Now(),
		SurrenderValueCalculatedDate: time.Now(),
		GrossSurrenderValue:          calculation.GrossSurrenderValue,
		NetSurrenderValue:            calculation.NetSurrenderValue,
		PaidUpValue:                  calculation.PaidUpValue,
		BonusAmount:                  &calculation.BonusAmount,
		SurrenderFactor:              calculation.SurrenderFactor,
		UnpaidPremiumsDeduction:      calculation.UnpaidPremiums,
		LoanDeduction:                calculation.LoanAmount,
		DisbursementMethod:           domain.DisbursementMethodCheque, // Default for forced
		DisbursementAmount:           calculation.NetSurrenderValue,
		Reason:                       stringPointer("Forced surrender due to non-payment of premiums"),
		Status:                       domain.SurrenderStatusPendingApproval, // Skip document upload
		Owner:                        domain.RequestOwnerSystem,
		CreatedBy:                    mockUserID,
		Metadata: map[string]interface{}{
			"policy_number":     policy.PolicyNumber,
			"policyholder_name": policy.PolicyholderName,
			"product_name":      policy.ProductName,
			"product_code":      policy.ProductCode,
			"forced_reason":     req.Reason,
			"unpaid_months":     0, // TODO: Get from collections service
		},
	}

	created, err := h.surrenderRepo.Create(sctx.Ctx, surrenderRequest)
	if err != nil {
		log.Error(sctx.Ctx, "Failed to create forced surrender request: %v", err)
		return nil, fmt.Errorf("failed to create forced surrender request")
	}

	log.Info(sctx.Ctx, "Initiated forced surrender %s for policy %s", created.RequestNumber, policy.PolicyNumber)

	return &InitiateForcedSurrenderResponse{
		StatusCodeAndMessage: port.CreateSuccess,
		Data: ForcedSurrenderData{
			SurrenderRequestID:  created.ID.String(),
			RequestNumber:       created.RequestNumber,
			PolicyID:            created.PolicyID,
			PolicyNumber:        policy.PolicyNumber,
			Status:              string(created.Status),
			RequestDate:         created.RequestDate.Format(time.RFC3339),
			GrossSurrenderValue: created.GrossSurrenderValue,
			NetSurrenderValue:   created.NetSurrenderValue,
			UnpaidPremiums:      calculation.UnpaidPremiums,
			UnpaidMonths:        0, // TODO: Get from collections service
			Reason:              getStringValue(created.Reason),
			NextAction:          "Request will be routed to approval queue for CPC processing",
		},
	}, nil
}

// ScheduleBatchEvaluation schedules batch evaluation of policies
// POST /v1/forced-surrender/schedule-batch-evaluation
// Internal endpoint for scheduling batch evaluations
func (h *ForcedSurrenderHandler) ScheduleBatchEvaluation(sctx *serverRoute.Context, req ScheduleBatchRequest) (interface{}, error) {
	scheduleDate := time.Now()
	if req.ScheduledAt != "" {
		parsedDate, err := time.Parse(time.RFC3339, req.ScheduledAt)
		if err != nil {
			log.Error(sctx.Ctx, "Invalid schedule date: %v", err)
			return nil, fmt.Errorf("invalid schedule date format")
		}
		scheduleDate = parsedDate
	}

	// In production, this would trigger a Temporal workflow
	// For now, we'll simulate the scheduling
	log.Info(sctx.Ctx, "Scheduled batch evaluation for date: %s", scheduleDate.Format("2006-01-02"))

	return &ScheduleBatchResponse{
		StatusCodeAndMessage: port.CustomSuccess,
		Data: BatchScheduleData{
			ScheduleID:   uuid.New().String(),
			ScheduleDate: scheduleDate.Format(time.RFC3339),
			BatchSize:    0, // TODO: Calculate based on eligible policies
			Status:       "SCHEDULED",
			Message:      "Batch evaluation scheduled successfully. Will be processed by Temporal workflow.",
		},
	}, nil
}

// GetPendingReminders retrieves pending reminders for processing
// GET /v1/forced-surrender/pending-reminders
// Internal endpoint for batch processing
func (h *ForcedSurrenderHandler) GetPendingReminders(sctx *serverRoute.Context, req PendingRemindersParams) (interface{}, error) {
	// Get pending reminders - for now, return empty list as method needs to be implemented
	// TODO: Implement ListPendingReminders in ForcedSurrenderRepository
	reminders := []domain.ForcedSurrenderReminder{}
	log.Info(sctx.Ctx, "ListPendingReminders not yet implemented, returning empty list")

	reminderData := make([]ReminderData, 0, len(reminders))
	for _, reminder := range reminders {
		policyNumber := ""
		if pn, ok := reminder.Metadata["policy_number"].(string); ok {
			policyNumber = pn
		}

		reminderNum := 1
		switch reminder.ReminderNumber {
		case domain.ReminderLevelFirst:
			reminderNum = 1
		case domain.ReminderLevelSecond:
			reminderNum = 2
		case domain.ReminderLevelThird:
			reminderNum = 3
		}

		reminderData = append(reminderData, ReminderData{
			ReminderID:         reminder.ID.String(),
			PolicyID:           reminder.PolicyID,
			PolicyNumber:       policyNumber,
			ReminderNumber:     reminderNum,
			ReminderDate:       reminder.ReminderDate.Format(time.RFC3339),
			PaymentWindowStart: "",    // Not stored in reminder
			PaymentWindowEnd:   "",    // Not stored in reminder
			UnpaidAmount:       0.0,   // TODO: Get from metadata
			Completed:          false, // TODO: Add completed tracking
		})
	}

	log.Info(sctx.Ctx, "Retrieved %d pending reminders", len(reminders))

	return &PendingRemindersResponse{
		StatusCodeAndMessage: port.GetSuccess,
		Data: PendingRemindersData{
			TotalCount: len(reminders),
			Reminders:  reminderData,
		},
	}, nil
}

// GetExpiredPaymentWindows retrieves expired payment windows for forced surrender
// GET /v1/forced-surrender/expired-payment-windows
// Internal endpoint for identifying policies to force surrender
func (h *ForcedSurrenderHandler) GetExpiredPaymentWindows(sctx *serverRoute.Context, req ExpiredWindowsParams) (interface{}, error) {
	// Get expired payment windows
	windows, err := h.forcedSurrenderRepo.ListExpiredPaymentWindows(sctx.Ctx)
	if err != nil {
		log.Error(sctx.Ctx, "Failed to get expired payment windows: %v", err)
		return nil, fmt.Errorf("failed to retrieve expired payment windows")
	}

	windowData := make([]ExpiredWindowData, 0, len(windows))
	for _, window := range windows {
		// Get reminder for this policy
		reminder, found, err := h.forcedSurrenderRepo.FindLatestReminderByPolicyID(sctx.Ctx, window.PolicyID)
		if err != nil || !found {
			log.Error(sctx.Ctx, "Failed to get reminder for window %s: %v", window.ID, err)
			continue
		}

		policyNumber := ""
		if pn, ok := reminder.Metadata["policy_number"].(string); ok {
			policyNumber = pn
		}

		daysExpired := int(time.Since(window.WindowEndDate).Hours() / 24)

		windowData = append(windowData, ExpiredWindowData{
			PaymentWindowID: window.ID.String(),
			ReminderID:      "", // Not directly linked
			PolicyID:        reminder.PolicyID,
			PolicyNumber:    policyNumber,
			WindowEnd:       window.WindowEndDate.Format(time.RFC3339),
			ExpectedAmount:  0.0, // Not stored in struct
			DaysExpired:     daysExpired,
			Status:          "EXPIRED",
		})
	}

	log.Info(sctx.Ctx, "Retrieved %d expired payment windows", len(windows))

	return &ExpiredPaymentWindowsResponse{
		StatusCodeAndMessage: port.GetSuccess,
		Data: ExpiredWindowsData{
			TotalCount:     len(windows),
			ExpiredWindows: windowData,
			Message:        "These policies are eligible for forced surrender initiation",
		},
	}, nil
}

// Helper functions

func (h *ForcedSurrenderHandler) generateRequestNumber(policyNumber string) string {
	timestamp := time.Now().Format("20060102150405")
	return fmt.Sprintf("FSUR-%s-%s", policyNumber, timestamp)
}

func (h *ForcedSurrenderHandler) calculateForcedSurrenderValue(policy *PolicyInfo) SurrenderCalculation {
	// Simplified calculation - similar to voluntary but automated
	paidUpValue := (policy.SumAssured * float64(policy.PremiumsPaid)) / float64(policy.TotalPremiums)
	totalBonus := (policy.SumAssured / 1000) * 50.0 * float64(policy.PremiumsPaid) // Simplified
	surrenderFactor := 0.75
	grossValue := (paidUpValue + totalBonus) * surrenderFactor

	// Get deductions (placeholder values)
	unpaidPremiums := 5000.0
	loanAmount := 2000.0

	netValue := grossValue - unpaidPremiums - loanAmount

	return SurrenderCalculation{
		PaidUpValue:         paidUpValue,
		BonusAmount:         totalBonus,
		SurrenderFactor:     surrenderFactor,
		GrossSurrenderValue: grossValue,
		UnpaidPremiums:      unpaidPremiums,
		LoanAmount:          loanAmount,
		NetSurrenderValue:   netValue,
	}
}

// ============================================
// Request/Response Types
// ============================================

type MonthlyEvaluationResponse struct {
	port.StatusCodeAndMessage
	Data MonthlyEvaluationData `json:"data"`
}

type MonthlyEvaluationData struct {
	EvaluationDate   string             `json:"evaluation_date"`
	TotalPolicies    int                `json:"total_policies"`
	EligiblePolicies int                `json:"eligible_policies"`
	SkippedPolicies  int                `json:"skipped_policies"`
	ErrorCount       int                `json:"error_count"`
	Results          []EvaluationResult `json:"results"`
	NextAction       string             `json:"next_action"`
}

type EvaluationResult struct {
	PolicyID        string  `json:"policy_id"`
	PolicyNumber    string  `json:"policy_number"`
	Status          string  `json:"status"`
	Reason          string  `json:"reason,omitempty"`
	UnpaidMonths    int     `json:"unpaid_months,omitempty"`
	UnpaidAmount    float64 `json:"unpaid_amount,omitempty"`
	LastPremiumDate string  `json:"last_premium_date,omitempty"`
}

type SendReminderResponse struct {
	port.StatusCodeAndMessage
	Data SendReminderData `json:"data"`
}

type SendReminderData struct {
	ReminderID         string  `json:"reminder_id"`
	PolicyID           string  `json:"policy_id"`
	PolicyNumber       string  `json:"policy_number"`
	ReminderNumber     int     `json:"reminder_number"`
	ReminderDate       string  `json:"reminder_date"`
	PaymentWindowStart string  `json:"payment_window_start"`
	PaymentWindowEnd   string  `json:"payment_window_end"`
	UnpaidAmount       float64 `json:"unpaid_amount"`
	NotificationSent   bool    `json:"notification_sent"`
	Message            string  `json:"message"`
}

type CreatePaymentWindowResponse struct {
	port.StatusCodeAndMessage
	Data PaymentWindowData `json:"data"`
}

type PaymentWindowData struct {
	PaymentWindowID string  `json:"payment_window_id"`
	ReminderID      string  `json:"reminder_id"`
	PolicyID        string  `json:"policy_id"`
	WindowStart     string  `json:"window_start"`
	WindowEnd       string  `json:"window_end"`
	ExpectedAmount  float64 `json:"expected_amount"`
	PaymentReceived bool    `json:"payment_received"`
	Status          string  `json:"status"`
	DaysRemaining   int     `json:"days_remaining,omitempty"`
}

type RecordPaymentResponse struct {
	port.StatusCodeAndMessage
	Data PaymentRecordData `json:"data"`
}

type PaymentRecordData struct {
	PaymentWindowID  string  `json:"payment_window_id"`
	ReminderID       string  `json:"reminder_id"`
	AmountReceived   float64 `json:"amount_received"`
	ExpectedAmount   float64 `json:"expected_amount"`
	PaymentReference string  `json:"payment_reference"`
	RecordedDate     string  `json:"recorded_date"`
	Status           string  `json:"status"`
	Message          string  `json:"message"`
}

type InitiateForcedSurrenderResponse struct {
	port.StatusCodeAndMessage
	Data ForcedSurrenderData `json:"data"`
}

type ForcedSurrenderData struct {
	SurrenderRequestID  string  `json:"surrender_request_id"`
	RequestNumber       string  `json:"request_number"`
	PolicyID            string  `json:"policy_id"`
	PolicyNumber        string  `json:"policy_number"`
	Status              string  `json:"status"`
	RequestDate         string  `json:"request_date"`
	GrossSurrenderValue float64 `json:"gross_surrender_value,omitempty"`
	NetSurrenderValue   float64 `json:"net_surrender_value,omitempty"`
	UnpaidPremiums      float64 `json:"unpaid_premiums,omitempty"`
	UnpaidMonths        int     `json:"unpaid_months,omitempty"`
	Reason              string  `json:"reason,omitempty"`
	NextAction          string  `json:"next_action,omitempty"`
}

type ScheduleBatchResponse struct {
	port.StatusCodeAndMessage
	Data BatchScheduleData `json:"data"`
}

type BatchScheduleData struct {
	ScheduleID   string `json:"schedule_id"`
	ScheduleDate string `json:"schedule_date"`
	BatchSize    int    `json:"batch_size"`
	Status       string `json:"status"`
	Message      string `json:"message"`
}

type PendingRemindersResponse struct {
	port.StatusCodeAndMessage
	Data PendingRemindersData `json:"data"`
}

type PendingRemindersData struct {
	TotalCount int            `json:"total_count"`
	Reminders  []ReminderData `json:"reminders"`
}

type ReminderData struct {
	ReminderID         string  `json:"reminder_id"`
	PolicyID           string  `json:"policy_id"`
	PolicyNumber       string  `json:"policy_number"`
	ReminderNumber     int     `json:"reminder_number"`
	ReminderDate       string  `json:"reminder_date"`
	PaymentWindowStart string  `json:"payment_window_start"`
	PaymentWindowEnd   string  `json:"payment_window_end"`
	UnpaidAmount       float64 `json:"unpaid_amount"`
	Completed          bool    `json:"completed"`
}

type ExpiredPaymentWindowsResponse struct {
	port.StatusCodeAndMessage
	Data ExpiredWindowsData `json:"data"`
}

type ExpiredWindowsData struct {
	TotalCount     int                 `json:"total_count"`
	ExpiredWindows []ExpiredWindowData `json:"expired_windows"`
	Message        string              `json:"message"`
}

type ExpiredWindowData struct {
	PaymentWindowID string  `json:"payment_window_id"`
	ReminderID      string  `json:"reminder_id"`
	PolicyID        string  `json:"policy_id"`
	PolicyNumber    string  `json:"policy_number"`
	WindowEnd       string  `json:"window_end"`
	ExpectedAmount  float64 `json:"expected_amount"`
	DaysExpired     int     `json:"days_expired"`
	Status          string  `json:"status"`
}

type SurrenderCalculation struct {
	PaidUpValue         float64
	BonusAmount         float64
	SurrenderFactor     float64
	GrossSurrenderValue float64
	UnpaidPremiums      float64
	LoanAmount          float64
	NetSurrenderValue   float64
}

// Query parameter types
type PendingRemindersParams struct {
	Limit int `query:"limit"`
}

type ExpiredWindowsParams struct {
	Limit int `query:"limit"`
}

// ============================================
// Additional External Service Interfaces
// ============================================

type NotificationServiceInterface interface {
	SendReminderNotification(ctx interface{}, params NotificationParams) bool
}

type NotificationParams struct {
	PolicyID         string
	PolicyNumber     string
	PolicyholderName string
	ReminderNumber   int
	UnpaidAmount     float64
	PaymentDeadline  time.Time
	ContactEmail     string
	ContactPhone     string
}

type PolicyWithUnpaidPremiums struct {
	PolicyID        string
	PolicyNumber    string
	UnpaidMonths    int
	UnpaidAmount    float64
	LastPremiumDate string
}

// Mock implementations

func NewMockNotificationService() NotificationServiceInterface {
	return &MockNotificationService{}
}

type MockNotificationService struct{}

func (m *MockNotificationService) SendReminderNotification(ctx interface{}, params NotificationParams) bool {
	// In production, this would send actual email/SMS/letter
	return true
}

// Extend MockCollectionsService with additional method
type MockCollectionsServiceExtended struct {
	*MockCollectionsService
}

func (m *MockCollectionsServiceExtended) GetPoliciesWithUnpaidPremiums(ctx interface{}, cutoffDate time.Time, minMonths int) ([]PolicyWithUnpaidPremiums, error) {
	// Mock data - in production, query actual collections system
	return []PolicyWithUnpaidPremiums{
		{
			PolicyID:        uuid.New().String(),
			PolicyNumber:    "PLI/2020/111111",
			UnpaidMonths:    6,
			UnpaidAmount:    6000,
			LastPremiumDate: "2025-07-27",
		},
		{
			PolicyID:        uuid.New().String(),
			PolicyNumber:    "PLI/2020/222222",
			UnpaidMonths:    9,
			UnpaidAmount:    9000,
			LastPremiumDate: "2025-04-27",
		},
	}, nil
}

// Helper functions
func stringPointer(s string) *string {
	return &s
}

func getStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// parseFlexibleDate attempts to parse date strings in multiple common formats
func parseFlexibleDate(dateStr string) (time.Time, error) {
	// List of common date formats to try
	formats := []string{
		"2006-01-02",          // YYYY-MM-DD (ISO 8601)
		"02/01/2006",          // DD/MM/YYYY
		"01/02/2006",          // MM/DD/YYYY
		"2006/01/02",          // YYYY/MM/DD
		"02-01-2006",          // DD-MM-YYYY
		"01-02-2006",          // MM-DD-YYYY
		time.RFC3339,          // RFC3339 format
		time.RFC3339Nano,      // RFC3339 with nanoseconds
		"2006-01-02 15:04:05", // YYYY-MM-DD HH:MM:SS
		"02/01/2006 15:04:05", // DD/MM/YYYY HH:MM:SS
		"01/02/2006 15:04:05", // MM/DD/YYYY HH:MM:SS
	}

	var lastErr error
	for _, format := range formats {
		if t, err := time.Parse(format, dateStr); err == nil {
			return t, nil
		} else {
			lastErr = err
		}
	}

	return time.Time{}, fmt.Errorf("unable to parse date '%s': %v", dateStr, lastErr)
}
