package workflows

import (
	"encoding/json"
	"time"
)

// ============================================================
// PM Integration Contract Types
//
// These types MUST exactly match the corresponding types in
// Policy Management's workflows/signals.go so Temporal can
// deserialise the child workflow input and serialise the
// completion signal payload correctly.
// ============================================================

// SurrenderProcessingInput is the input sent by Policy Management's
// PolicyLifecycleWorkflow when it dispatches SurrenderProcessingWorkflow
// as a child workflow via ExecuteChildWorkflow.
//
// Field names and JSON tags must match PM's ChildWorkflowInput exactly.
type SurrenderProcessingInput struct {
	// RequestID is PM's idempotency key (UUID from X-Idempotency-Key header).
	// Used as the dedup key and as the child workflow ID fragment by PM.
	RequestID string `json:"request_id"`

	// PolicyNumber is the canonical policy identifier (e.g. "PLI/2026/000001").
	PolicyNumber string `json:"policy_number"`

	// PolicyDBID is PM's BIGINT policy_id from seq_policy_id.
	PolicyDBID int64 `json:"policy_db_id"`

	// ServiceRequestID is PM's BIGINT PK from the service_request table.
	// Used for cross-referencing back to PM's service_request row.
	ServiceRequestID int64 `json:"service_request_id"`

	// RequestType will always be "SURRENDER" for this workflow.
	RequestType string `json:"request_type"`

	// RequestPayload is the original JSON body submitted to PM's
	// POST /v1/policies/{pn}/requests/surrender endpoint.
	// Contains: disbursement_method, bank_account_id, reason, source_channel.
	RequestPayload json.RawMessage `json:"request_payload"`

	// TimeoutAt is the deadline assigned by PM's routing-timeout config.
	TimeoutAt time.Time `json:"timeout_at"`
}

// PMSurrenderRequestPayload is the shape of RequestPayload sent by PM.
// Matches SubmitSurrenderRequest.Payload in PM's handler/request.go.
type PMSurrenderRequestPayload struct {
	SourceChannel      string `json:"source_channel"`
	DisbursementMethod string `json:"disbursement_method"`
	BankAccountID      int64  `json:"bank_account_id"`
	Reason             string `json:"reason"`
}

// OperationCompletedSignal is the payload sent back to PM's
// PolicyLifecycleWorkflow on the "surrender-completed" signal channel.
//
// Field names and JSON tags must match PM's OperationCompletedSignal exactly.
type OperationCompletedSignal struct {
	// RequestID must equal SurrenderProcessingInput.RequestID for PM to
	// correlate the completion back to the correct pending request.
	RequestID string `json:"request_id"`

	// RequestType is always "SURRENDER".
	RequestType string `json:"request_type"`

	// Outcome is one of: APPROVED, REJECTED, TIMEOUT.
	Outcome string `json:"outcome"`

	// StateTransition describes the PM policy state change, e.g.
	// "PENDING_SURRENDER→SURRENDERED" or "PENDING_SURRENDER→ACTIVE".
	StateTransition string `json:"state_transition,omitempty"`

	// OutcomePayload carries surrender-specific result data (optional).
	OutcomePayload json.RawMessage `json:"outcome_payload,omitempty"`

	// CompletedAt is the UTC time the workflow reached a terminal state.
	CompletedAt time.Time `json:"completed_at"`
}

// Outcome values used in OperationCompletedSignal.
const (
	OutcomeApproved = "APPROVED"
	OutcomeRejected = "REJECTED"
	OutcomeTimeout  = "TIMEOUT"
)

// State transitions sent back to PM.
const (
	StateTransitionSurrendered        = "PENDING_SURRENDER→SURRENDERED"
	StateTransitionSurrenderRejected  = "PENDING_SURRENDER→ACTIVE"
	StateTransitionSurrenderTimeout   = "PENDING_SURRENDER→ACTIVE"
)
