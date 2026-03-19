package activities

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.temporal.io/sdk/client"
)

// pmSignalActivities holds the Temporal client used to signal PM workflows.
type pmSignalActivities struct {
	temporalClient client.Client
}

var pmSignalInstance *pmSignalActivities

// InitPMSignalActivities initialises the PM signal activity with a Temporal client.
// Must be called during bootstrap before the worker starts.
func InitPMSignalActivities(tc client.Client) {
	pmSignalInstance = &pmSignalActivities{temporalClient: tc}
}

// SignalPMWorkflowInput is the input for SignalPMWorkflowActivity.
// It carries everything needed to send the "surrender-completed" signal
// to PM's PolicyLifecycleWorkflow.
type SignalPMWorkflowInput struct {
	// PMWorkflowID is the target workflow, e.g. "plw-{policyNumber}".
	PMWorkflowID string `json:"pm_workflow_id"`
	// SignalName is the signal channel name on the PM workflow, e.g. "surrender-completed".
	SignalName string `json:"signal_name"`
	// RequestID is PM's idempotency key (must match SurrenderProcessingInput.RequestID).
	RequestID string `json:"request_id"`
	// RequestType is always "SURRENDER".
	RequestType string `json:"request_type"`
	// Outcome is one of APPROVED, REJECTED, TIMEOUT.
	Outcome string `json:"outcome"`
	// StateTransition describes the policy state change, e.g. "PENDING_SURRENDER→SURRENDERED".
	StateTransition string `json:"state_transition,omitempty"`
	// OutcomePayload carries optional surrender-specific result data.
	OutcomePayload json.RawMessage `json:"outcome_payload,omitempty"`
}

// pmOperationCompletedSignal is the payload shape that PM's PolicyLifecycleWorkflow
// expects on the "surrender-completed" signal channel.
// Field names and JSON tags must match PM's OperationCompletedSignal exactly.
type pmOperationCompletedSignal struct {
	RequestID       string          `json:"request_id"`
	RequestType     string          `json:"request_type"`
	Outcome         string          `json:"outcome"`
	StateTransition string          `json:"state_transition,omitempty"`
	OutcomePayload  json.RawMessage `json:"outcome_payload,omitempty"`
	CompletedAt     time.Time       `json:"completed_at"`
}

// SignalPMWorkflowActivity sends a signal to PM's PolicyLifecycleWorkflow
// to report the outcome of the surrender processing.
// This is invoked by SurrenderProcessingWorkflow in every terminal path.
func SignalPMWorkflowActivity(ctx context.Context, input SignalPMWorkflowInput) error {
	if pmSignalInstance == nil {
		return fmt.Errorf("PM signal activities not initialized")
	}

	payload := pmOperationCompletedSignal{
		RequestID:       input.RequestID,
		RequestType:     input.RequestType,
		Outcome:         input.Outcome,
		StateTransition: input.StateTransition,
		OutcomePayload:  input.OutcomePayload,
		CompletedAt:     time.Now().UTC(),
	}

	err := pmSignalInstance.temporalClient.SignalWorkflow(
		ctx,
		input.PMWorkflowID,
		"", // run ID — empty means latest open run
		input.SignalName,
		payload,
	)
	if err != nil {
		return fmt.Errorf("failed to signal PM workflow %s on channel %s: %w",
			input.PMWorkflowID, input.SignalName, err)
	}

	return nil
}
