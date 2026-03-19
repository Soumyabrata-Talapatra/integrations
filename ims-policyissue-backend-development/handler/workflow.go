package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"policy-issue-service/core/port"
	resp "policy-issue-service/handler/response"

	log "gitlab.cept.gov.in/it-2.0-common/n-api-log"
	serverHandler "gitlab.cept.gov.in/it-2.0-common/n-api-server/handler"
	serverRoute "gitlab.cept.gov.in/it-2.0-common/n-api-server/route"

	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
)

// WorkflowHandler handles Temporal workflow management HTTP endpoints
// Phase 9: [WF-POL-001] to [WF-POL-002]
type WorkflowHandler struct {
	*serverHandler.Base
	temporalClient client.Client
}

// NewWorkflowHandler creates a new WorkflowHandler instance
func NewWorkflowHandler(temporalClient client.Client) *WorkflowHandler {
	base := serverHandler.New("Workflow").SetPrefix("/v1").AddPrefix("")
	return &WorkflowHandler{
		Base:           base,
		temporalClient: temporalClient,
	}
}

// Routes returns the routes for the WorkflowHandler
func (h *WorkflowHandler) Routes() []serverRoute.Route {
	return []serverRoute.Route{
		serverRoute.GET("/workflows/:workflow_id/state", h.GetWorkflowState).Name("Get Workflow State"),
		serverRoute.POST("/workflows/:workflow_id/signal/:signal_name", h.SendWorkflowSignal).Name("Send Workflow Signal"),
	}
}

// GetWorkflowState retrieves the current state of a Temporal workflow
// [WF-POL-001] Get Temporal workflow state
// [WF-PI-001] Standard Policy Issuance Workflow
// Uses Temporal DescribeWorkflowExecution to get runtime info and
// QueryWorkflow to get the current step from the query handler.
func (h *WorkflowHandler) GetWorkflowState(sctx *serverRoute.Context, req WorkflowIDUri) (*resp.WorkflowStateResponse, error) {
	ctx := sctx.Ctx

	// Step 1: Describe the workflow execution for runtime metadata
	descResp, err := h.temporalClient.DescribeWorkflowExecution(ctx, req.WorkflowID, "")
	if err != nil {
		log.Error(ctx, "[WF-POL-001] Error describing workflow %s: %v", req.WorkflowID, err)
		return nil, fmt.Errorf("workflow %s not found or not accessible: %w", req.WorkflowID, err)
	}

	info := descResp.WorkflowExecutionInfo
	workflowType := ""
	if info.Type != nil {
		workflowType = info.Type.Name
	}

	// Map Temporal workflow status to our status enum
	status := mapTemporalStatus(info.Status)

	// Determine start time and last activity
	var startTime *interface{ UnixMilli() int64 }
	_ = startTime

	// Step 2: Query the workflow for current step (uses SetQueryHandler registered in workflow)
	currentStep := ""
	queryResult, err := h.temporalClient.QueryWorkflow(ctx, req.WorkflowID, "", "QueryProposalStatus")
	if err != nil {
		log.Warn(ctx, "[WF-POL-001] Could not query workflow %s for status: %v", req.WorkflowID, err)
		currentStep = status // Fallback to workflow status
	} else {
		if err := queryResult.Get(&currentStep); err != nil {
			log.Warn(ctx, "[WF-POL-001] Could not decode query result for workflow %s: %v", req.WorkflowID, err)
			currentStep = status
		}
	}

	// Step 3: Determine progress and pending signals based on workflow type and step
	progress, pendingSignals := determineProgress(workflowType, currentStep)

	// Extract timestamps
	response := &resp.WorkflowStateResponse{
		StatusCodeAndMessage: port.StatusCodeAndMessage{
			StatusCode: http.StatusOK,
			Message:    "Workflow state retrieved successfully",
		},
		WorkflowID:     req.WorkflowID,
		WorkflowType:   workflowType,
		Status:         status,
		CurrentStep:    currentStep,
		Progress:       progress,
		PendingSignals: pendingSignals,
	}

	if info.StartTime != nil {
		t := info.StartTime.AsTime()
		response.StartTime = &t
	}
	if info.CloseTime != nil {
		t := info.CloseTime.AsTime()
		response.LastActivity = &t
	} else if info.StartTime != nil {
		// If not closed yet, use execution time as last activity
		t := info.StartTime.AsTime()
		response.LastActivity = &t
	}

	return response, nil
}

// SendWorkflowSignal sends a signal to a running Temporal workflow
// [WF-POL-002] Send signal to workflow
// Available Signals must match constants in workflows/policy_issuance_workflow.go:
//   qr-decision, medical-result, approver-decision, cpc-resubmit,
//   payment-received, flc-cancel-request, death-notification
func (h *WorkflowHandler) SendWorkflowSignal(sctx *serverRoute.Context, req WorkflowSignalRequest) (*resp.WorkflowSignalResponse, error) {
	ctx := sctx.Ctx

	// Step 1: Validate the signal name
	// These MUST match the Signal* constants in workflows/policy_issuance_workflow.go
	validSignals := map[string]bool{
		"qr-decision":        true, // SignalQRDecision
		"medical-result":     true, // SignalMedicalResult
		"approver-decision":  true, // SignalApproverDecision
		"cpc-resubmit":       true, // SignalCPCResubmit
		"payment-received":   true, // future signal
		"flc-cancel-request": true, // future signal
		"death-notification": true, // future signal
	}

	if !validSignals[req.SignalName] {
		return nil, fmt.Errorf("[WF-POL-002] invalid signal name: %s", req.SignalName)
	}

	// Step 2: Serialize the payload
	var signalPayload interface{}
	if req.Payload != nil {
		// Convert map to JSON bytes for the signal
		payloadBytes, err := json.Marshal(req.Payload)
		if err != nil {
			return nil, fmt.Errorf("[WF-POL-002] failed to serialize signal payload: %w", err)
		}
		signalPayload = payloadBytes
	}

	// Step 3: Send the signal to the workflow via Temporal client
	err := h.temporalClient.SignalWorkflow(ctx, req.WorkflowID, "", req.SignalName, signalPayload)
	if err != nil {
		log.Error(ctx, "[WF-POL-002] Error sending signal %s to workflow %s: %v", req.SignalName, req.WorkflowID, err)
		return nil, fmt.Errorf("failed to send signal to workflow %s: %w", req.WorkflowID, err)
	}

	log.Info(ctx, "[WF-POL-002] Signal sent: workflow=%s, signal=%s", req.WorkflowID, req.SignalName)

	return &resp.WorkflowSignalResponse{
		StatusCodeAndMessage: port.StatusCodeAndMessage{
			StatusCode: http.StatusAccepted,
			Message:    "Signal accepted and sent to workflow",
		},
		WorkflowID: req.WorkflowID,
		SignalName: req.SignalName,
		Accepted:   true,
	}, nil
}

// mapTemporalStatus maps Temporal workflow execution status to our string status
func mapTemporalStatus(status enums.WorkflowExecutionStatus) string {
	switch status {
	case enums.WORKFLOW_EXECUTION_STATUS_RUNNING:
		return "RUNNING"
	case enums.WORKFLOW_EXECUTION_STATUS_COMPLETED:
		return "COMPLETED"
	case enums.WORKFLOW_EXECUTION_STATUS_FAILED:
		return "FAILED"
	case enums.WORKFLOW_EXECUTION_STATUS_CANCELED:
		return "TERMINATED"
	case enums.WORKFLOW_EXECUTION_STATUS_TERMINATED:
		return "TERMINATED"
	case enums.WORKFLOW_EXECUTION_STATUS_TIMED_OUT:
		return "FAILED"
	default:
		return "RUNNING"
	}
}

// determineProgress estimates workflow progress based on type and current step
// Step names MUST match the currentStatus values set in the actual workflow code
// (see workflows/policy_issuance_workflow.go for PolicyIssuanceWorkflow)
func determineProgress(workflowType string, currentStep string) (resp.WorkflowProgress, []string) {
	// Define step sequences for each workflow type
	type stepDef struct {
		totalSteps int
		stepMap    map[string]int
		signals    map[string][]string // pending signals at each step
	}

	// PolicyIssuanceWorkflow steps — aligned with currentStatus values in
	// workflows/policy_issuance_workflow.go lines 113-327
	policySteps := stepDef{
		totalSteps: 10,
		stepMap: map[string]int{
			"VALIDATING":                1,  // line 113
			"CHECKING_ELIGIBILITY":      2,  // line 130
			"CALCULATING_PREMIUM":       3,  // line 150
			"QC_PENDING":                4,  // line 185 — waiting for qr-decision signal
			"QC_APPROVED":               5,  // line 205
			"QC_RETURNED":               4,  // line 208 — loops back to QC_PENDING
			"PENDING_MEDICAL":           6,  // line 223 — waiting for medical-result signal
			"MEDICAL_APPROVED":          7,  // line 259
			"APPROVAL_PENDING":          8,  // line 263 — waiting for approver-decision signal
			"APPROVED":                  9,  // line 296
			"GENERATING_POLICY_NUMBER":  9,  // line 300
			"GENERATING_BOND":           10, // line 327
		},
		signals: map[string][]string{
			"QC_PENDING":       {"qr-decision"},        // SignalQRDecision
			"QC_RETURNED":      {"cpc-resubmit"},       // SignalCPCResubmit
			"PENDING_MEDICAL":  {"medical-result"},     // SignalMedicalResult
			"APPROVAL_PENDING": {"approver-decision"},  // SignalApproverDecision
		},
	}

	// InstantIssuanceWorkflow steps
	instantSteps := stepDef{
		totalSteps: 5,
		stepMap: map[string]int{
			"VALIDATING":     1,
			"CALCULATING":    2,
			"AUTO_APPROVING": 3,
			"ISSUING":        4,
			"COMPLETED":      5,
		},
		signals: map[string][]string{},
	}

	// BulkProposalUploadWorkflow steps
	bulkSteps := stepDef{
		totalSteps: 4,
		stepMap: map[string]int{
			"PARSING":    1,
			"VALIDATING": 2,
			"PROCESSING": 3,
			"COMPLETED":  4,
		},
		signals: map[string][]string{},
	}

	var steps stepDef
	switch workflowType {
	case "PolicyIssuanceWorkflow":
		steps = policySteps
	case "InstantIssuanceWorkflow":
		steps = instantSteps
	case "BulkProposalUploadWorkflow":
		steps = bulkSteps
	default:
		// Unknown workflow type, return generic progress
		return resp.WorkflowProgress{
			CompletedSteps: 0,
			TotalSteps:     1,
			Percentage:     0,
		}, nil
	}

	completedSteps := 0
	if step, ok := steps.stepMap[currentStep]; ok {
		completedSteps = step
	}

	percentage := 0
	if steps.totalSteps > 0 {
		percentage = (completedSteps * 100) / steps.totalSteps
	}

	// Determine pending signals for current step
	var pendingSignals []string
	if signals, ok := steps.signals[currentStep]; ok {
		pendingSignals = signals
	}

	return resp.WorkflowProgress{
		CompletedSteps: completedSteps,
		TotalSteps:     steps.totalSteps,
		Percentage:     percentage,
	}, pendingSignals
}

// Ensure context is imported (used in DescribeWorkflowExecution)
var _ context.Context
