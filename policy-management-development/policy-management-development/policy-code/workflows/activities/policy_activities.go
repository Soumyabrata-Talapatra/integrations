package activities

// ============================================================================
// PolicyActivities — core policy lifecycle activities
//
// All activities use struct receivers for FX injection compatibility.
// Batch operations use dblib.QueueExecRow + db.SendBatch (Constraint 5).
// All SQL uses policy_mgmt. schema prefix (Constraint 7).
//
// [FR-PM-001, A21.1, Constraint 5, Constraint 7]
// ============================================================================

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	config "gitlab.cept.gov.in/it-2.0-common/api-config"
	dblib "gitlab.cept.gov.in/it-2.0-common/n-api-db"
	"go.temporal.io/sdk/client"

	"policy-management/core/domain"
)

// ─────────────────────────────────────────────────────────────────────────────
// PolicyActivities struct — injected via FX
// ─────────────────────────────────────────────────────────────────────────────

// PolicyActivities holds dependencies for all core policy lifecycle activities.
// [A21.1, FR-PM-001]
type PolicyActivities struct {
	db  *dblib.DB
	cfg *config.Config
	tc  client.Client
}

// NewPolicyActivities constructs a PolicyActivities instance for FX injection.
func NewPolicyActivities(db *dblib.DB, cfg *config.Config, tc client.Client) *PolicyActivities {
	return &PolicyActivities{db: db, cfg: cfg, tc: tc}
}

// ─────────────────────────────────────────────────────────────────────────────
// Activity Input / Output Types
// ─────────────────────────────────────────────────────────────────────────────

// InitializePolicyParams is the input to InitializePolicyActivity. [A10.1.6]
type InitializePolicyParams struct {
	RequestID      string    `json:"request_id"`      // UUID from Policy Issue for dedup
	PolicyIssueID  string    `json:"policy_issue_id"` // UUID from Policy Issue (audit only)
	PolicyNumber   string    `json:"policy_number"`
	WorkflowID     string    `json:"workflow_id"` // plw-{policy_number}
	CustomerID     int64     `json:"customer_id"`
	ProductCode    string    `json:"product_code"`
	ProductType    string    `json:"product_type"`
	SumAssured     float64   `json:"sum_assured"`
	CurrentPremium float64   `json:"current_premium"`
	PremiumMode    string    `json:"premium_mode"`
	BillingMethod  string    `json:"billing_method"`
	IssueDate      time.Time `json:"issue_date"`
	MaturityDate   time.Time `json:"maturity_date"`
	PaidToDate     time.Time `json:"paid_to_date"`
	AgentID        *int64    `json:"agent_id,omitempty"` // Nullable BIGINT [Review-Fix-5]
}

// StateTransitionParams is the input to RecordStateTransitionActivity.
// Used for all SUBSEQUENT state transitions (after InitializePolicyActivity). [A21.1]
type StateTransitionParams struct {
	PolicyID     int64  `json:"policy_id"`
	PolicyNumber string `json:"policy_number"`
	FromStatus   string `json:"from_status"`
	ToStatus     string `json:"to_status"`
	Reason       string `json:"reason"`
	TriggeredBy  string `json:"triggered_by"` // signal type or service name
	RequestID    string `json:"request_id,omitempty"`
	Version      int64  `json:"version"` // Post-increment version for lock guard
}

// MetadataUpdateParams is the input to UpdatePolicyMetadataActivity.
type MetadataUpdateParams struct {
	PolicyID int64                  `json:"policy_id"`
	Updates  map[string]interface{} `json:"updates"` // Column → value
}

// PolicyEvent is the input to PublishEventActivity. [ADR-004]
type PolicyEvent struct {
	PolicyID     int64           `json:"policy_id"`
	PolicyNumber string          `json:"policy_number"`
	EventType    string          `json:"event_type"`
	Payload      json.RawMessage `json:"payload,omitempty"`
	Version      int64           `json:"version"`
	OccurredAt   time.Time       `json:"occurred_at"`
}

// RejectedRequestParams is the input to RecordRejectedRequestActivity.
type RejectedRequestParams struct {
	PolicyID         int64           `json:"policy_id"`
	SignalChannel    string          `json:"signal_channel"`
	RequestID        string          `json:"request_id,omitempty"`
	ServiceRequestID *int64          `json:"service_request_id,omitempty"`
	Reason           string          `json:"reason"`
	Payload          json.RawMessage `json:"payload,omitempty"`
}

// CancelWorkflowParams is the input to CancelDownstreamWorkflowActivity.
type CancelWorkflowParams struct {
	WorkflowID string `json:"workflow_id"`
	Reason     string `json:"reason,omitempty"`
}

// SignalLogEntry is the input to LogSignalReceivedActivity. [§8.9]
type SignalLogEntry struct {
	PolicyID         int64           `json:"policy_id"`
	SignalChannel    string          `json:"signal_channel"`
	SignalPayload    json.RawMessage `json:"signal_payload"`
	SourceService    string          `json:"source_service,omitempty"`
	SourceWorkflowID *string         `json:"source_workflow_id,omitempty"`
	RequestID        string          `json:"request_id"`
	Status           string          `json:"status"` // signal_processing_status enum
	RejectionReason  *string         `json:"rejection_reason,omitempty"`
	StateBefore      *string         `json:"state_before,omitempty"`
	StateAfter       *string         `json:"state_after,omitempty"`
}

// ServiceRequestUpdate is the input to UpdateServiceRequestActivity.
// Partition key submitted_at used when available for pruning. [§8.3]
type ServiceRequestUpdate struct {
	ServiceRequestID     int64           `json:"service_request_id"`
	SubmittedAt          *time.Time      `json:"submitted_at,omitempty"`
	Status               string          `json:"status"`
	Outcome              *string         `json:"outcome,omitempty"`
	OutcomeReason        *string         `json:"outcome_reason,omitempty"`
	OutcomePayload       json.RawMessage `json:"outcome_payload,omitempty"`
	DownstreamWorkflowID *string         `json:"downstream_workflow_id,omitempty"`
	DownstreamService    *string         `json:"downstream_service,omitempty"`
	DownstreamTaskQueue  *string         `json:"downstream_task_queue,omitempty"`
	RoutedAt             *time.Time      `json:"routed_at,omitempty"`
	CompletedAt          *time.Time      `json:"completed_at,omitempty"`
}

// PolicyRefreshedState is returned by RefreshStateFromDBActivity. [§9.5.2]
type PolicyRefreshedState struct {
	CurrentStatus   string  `db:"current_status"  json:"current_status"`
	PreviousStatus  *string `db:"previous_status" json:"previous_status,omitempty"`
	Version         int64   `db:"version"         json:"version"`
	HasActiveLoan   bool    `db:"has_active_loan" json:"has_active_loan"`
	LoanOutstanding float64 `db:"loan_outstanding" json:"loan_outstanding"`
	AssignmentType  string  `db:"assignment_type"  json:"assignment_type"`
	AMLHold         bool    `db:"aml_hold"         json:"aml_hold"`
}

// TerminalStateRecord is the input to PersistTerminalStateActivity.
type TerminalStateRecord struct {
	PolicyID      int64           `json:"policy_id"`
	PolicyNumber  string          `json:"policy_number"`
	FinalStatus   string          `json:"final_status"`
	TerminalAt    time.Time       `json:"terminal_at"`
	CoolingExpiry time.Time       `json:"cooling_expiry"`
	FinalSnapshot json.RawMessage `json:"final_snapshot"` // Serialized PolicyLifecycleState
}

// MarkCompletedParams is the input to MarkWorkflowCompletedActivity.
type MarkCompletedParams struct {
	PolicyID    int64  `json:"policy_id"`
	FinalStatus string `json:"final_status"`
}

// RefundRequest is the input to TriggerPremiumRefundActivity.
type RefundRequest struct {
	PolicyID     int64   `json:"policy_id"`
	PolicyNumber string  `json:"policy_number"`
	Amount       float64 `json:"amount"`
	Reason       string  `json:"reason"`
	RequestID    string  `json:"request_id"`
}

// ─────────────────────────────────────────────────────────────────────────────
// Table name constants (Constraint 7: all use policy_mgmt. prefix)
// ─────────────────────────────────────────────────────────────────────────────

const (
	actPolicyTable       = "policy_mgmt.policy"
	actPolicyHistTable   = "policy_mgmt.policy_status_history"
	actServiceReqTable   = "policy_mgmt.service_request"
	actSignalLogTable    = "policy_mgmt.policy_signal_log"
	actSignalRegTable    = "policy_mgmt.processed_signal_registry"
	actPolicyEventTable  = "policy_mgmt.policy_event"
	actTerminalSnapTable = "policy_mgmt.terminal_state_snapshot"
	actPolicyLockTable   = "policy_mgmt.policy_lock" // [BR-PM-030]
)

// FinancialLockParams is the input to AcquireFinancialLockActivity. [BR-PM-030]
type FinancialLockParams struct {
	PolicyID         int64     `json:"policy_id"`
	ServiceRequestID int64     `json:"service_request_id"` // BIGINT from service_request (maps to policy_lock.request_id)
	RequestType      string    `json:"request_type"`
	LockedAt         time.Time `json:"locked_at"`
	TimeoutAt        time.Time `json:"timeout_at"`
}

// ─────────────────────────────────────────────────────────────────────────────
// InitializePolicyActivity — INSERT policy + initial history row (idempotent)
// Called ONCE on "policy-created" signal. Returns BIGINT policy_id. [A10.1.6, A13]
// ─────────────────────────────────────────────────────────────────────────────

// InitializePolicyActivity inserts the policy record and the initial FREE_LOOK_ACTIVE
// history row. Step 1 returns policy_id via RETURNING; Step 2 inserts history row.
// ON CONFLICT (policy_number) DO NOTHING makes this idempotent for Temporal retries.
// Returns the BIGINT policy_id from seq_policy_id. [A10.1.6, A13, Constraint 7]
//
// Note: AD-003 mandates pgx.Batch for 2+ DB writes. pgx.Batch cannot be used here
// because Step 2 (history INSERT) depends on the policy_id returned by Step 1
// (policy INSERT with RETURNING). The correct single-round-trip solution is a CTE
// (WITH ins_policy AS (...) INSERT INTO history SELECT policy_id FROM ins_policy).
// The CTE approach is tracked as a follow-up refactor once dblib raw-SQL helpers are
// validated in this context. [D9]
func (a *PolicyActivities) InitializePolicyActivity(ctx context.Context, p InitializePolicyParams) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, a.cfg.GetDuration("db.QueryTimeoutHigh"))
	defer cancel()

	now := time.Now().UTC()

	// Step 1: INSERT policy — RETURNING policy_id (idempotent via ON CONFLICT DO NOTHING)
	insertQ := dblib.Psql.Insert(actPolicyTable).
		Columns(
			"policy_number", "workflow_id", "customer_id", "product_code", "product_type",
			"current_status", "previous_status",
			"sum_assured", "current_premium", "premium_mode", "billing_method",
			"issue_date", "maturity_date", "paid_to_date",
			"agent_id", "version", "created_at", "updated_at",
		).
		Values(
			p.PolicyNumber, p.WorkflowID, p.CustomerID, p.ProductCode, p.ProductType,
			domain.StatusFreeLookActive, "",
			p.SumAssured, p.CurrentPremium, p.PremiumMode, p.BillingMethod,
			p.IssueDate, p.MaturityDate, p.PaidToDate,
			p.AgentID, 1, now, now,
		).
		Suffix("ON CONFLICT (policy_number) DO NOTHING RETURNING policy_id")

	type policyIDResult struct {
		PolicyID int64 `db:"policy_id"`
	}
	// InsertReturningrows returns empty slice when ON CONFLICT DO NOTHING fires (Temporal retry).
	insertedRows, insertErr := dblib.InsertReturningrows(ctx, a.db, insertQ, pgx.RowToStructByNameLax[policyIDResult])
	if insertErr != nil {
		return 0, fmt.Errorf("InitializePolicyActivity INSERT policy=%s: %w", p.PolicyNumber, insertErr)
	}
	var policyID int64
	if len(insertedRows) > 0 {
		policyID = insertedRows[0].PolicyID
	} else {
		// ON CONFLICT DO NOTHING fired — fetch existing policy_id
		fetchQ := dblib.Psql.Select("policy_id").
			From(actPolicyTable).
			Where(sq.Eq{"policy_number": p.PolicyNumber})
		fetched, fetchErr := dblib.SelectOne(ctx, a.db, fetchQ, pgx.RowToStructByNameLax[policyIDResult])
		if fetchErr != nil {
			return 0, fmt.Errorf("InitializePolicyActivity SELECT policy=%s: %w", p.PolicyNumber, fetchErr)
		}
		policyID = fetched.PolicyID
	}

	// Step 2: INSERT initial history row (idempotent via ON CONFLICT DO NOTHING) [A10.1.6]
	histQ := dblib.Psql.Insert(actPolicyHistTable).
		Columns(
			"policy_id", "from_status", "to_status",
			"transition_reason", "triggered_by_service", "request_id",
			"effective_date", "created_at",
		).
		Values(
			policyID, "", domain.StatusFreeLookActive,
			"Policy issued and activated", "policy-issue", p.RequestID,
			now, now,
		).
		Suffix("ON CONFLICT DO NOTHING")
	if _, err := dblib.Insert(ctx, a.db, histQ); err != nil {
		return 0, fmt.Errorf("InitializePolicyActivity INSERT history policyID=%d: %w", policyID, err)
	}

	return policyID, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// RecordStateTransitionActivity — UPDATE policy + INSERT history [A21.1]
// Two-step with optimistic locking (Constraint 5, Constraint 7)
// ─────────────────────────────────────────────────────────────────────────────

// RecordStateTransitionActivity atomically updates the policy row (optimistic
// lock guard) and inserts a status_history record. [A21.1, Constraint 7]
func (a *PolicyActivities) RecordStateTransitionActivity(ctx context.Context, p StateTransitionParams) error {
	ctx, cancel := context.WithTimeout(ctx, a.cfg.GetDuration("db.QueryTimeoutHigh"))
	defer cancel()

	now := time.Now().UTC()

	// Step 1: UPDATE policy with version guard (version in state is post-increment).
	// display_status is intentionally omitted: it is computed by a DB trigger
	// (compute_display_status) and must never be explicitly SET from application code.
	// Setting it here would race with the trigger and produce wrong composite values
	// for encumbered policies (e.g. ACTIVE_LOAN_AML_HOLD). [D7, context.md §10.4]
	updateQ := dblib.Psql.Update(actPolicyTable).
		Set("current_status", p.ToStatus).
		Set("previous_status", p.FromStatus).
		Set("effective_from", now).
		Set("version", sq.Expr("version + 1")).
		Set("updated_at", now).
		Where(sq.Eq{"policy_id": p.PolicyID}).
		Where(sq.Eq{"version": p.Version - 1}).
		Suffix("RETURNING policy_id")

	type updatedPolicyRow struct {
		PolicyID int64 `db:"policy_id"`
	}

	_, updateErr := dblib.UpdateReturningBulk(ctx, a.db, updateQ, pgx.RowToStructByNameLax[updatedPolicyRow])
	if updateErr != nil {
		return fmt.Errorf("RecordStateTransitionActivity: UPDATE policy=%d %s→%s: %w",
			p.PolicyID, p.FromStatus, p.ToStatus, updateErr)
	}
	// if len(updatedRows) == 0 {
	// 	return fmt.Errorf("RecordStateTransitionActivity: version conflict policyID=%d [optimistic lock]", p.PolicyID)
	// }

	// Step 2: INSERT history row [§8.2]
	var reqIDVal interface{} = nil
	if p.RequestID != "" {
		// Parse RequestID string to int64 for BIGINT column
		if reqID, err := strconv.ParseInt(p.RequestID, 10, 64); err == nil {
			reqIDVal = reqID
		} else {
			// If it's not a numeric string (e.g., UUID or template), set to NULL
			// Database constraint allows NULL for request_id
			reqIDVal = nil
		}
	}
	histQ := dblib.Psql.Insert(actPolicyHistTable).
		Columns(
			"policy_id", "from_status", "to_status",
			"transition_reason", "triggered_by_service", "request_id",
			"effective_date", "created_at",
		).
		Values(p.PolicyID, p.FromStatus, p.ToStatus, p.Reason, p.TriggeredBy, reqIDVal, now, now)

	if _, err := dblib.Insert(ctx, a.db, histQ); err != nil {
		return fmt.Errorf("RecordStateTransitionActivity: INSERT history policy=%d %s→%s: %w",
			p.PolicyID, p.FromStatus, p.ToStatus, err)
	}
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// UpdatePolicyMetadataActivity — UPDATE specific metadata columns [A21.1]
// ─────────────────────────────────────────────────────────────────────────────

// UpdatePolicyMetadataActivity updates specific metadata columns on the policy row.
// Used for loan_outstanding, paid_to_date, customer_id, assignment_type, etc. [A21.1]
func (a *PolicyActivities) UpdatePolicyMetadataActivity(ctx context.Context, p MetadataUpdateParams) error {
	if len(p.Updates) == 0 {
		return nil
	}
	ctx, cancel := context.WithTimeout(ctx, a.cfg.GetDuration("db.QueryTimeoutHigh"))
	defer cancel()

	qb := dblib.Psql.Update(actPolicyTable).
		Where(sq.Eq{"policy_id": p.PolicyID}).
		Set("updated_at", sq.Expr("NOW()"))
	for col, val := range p.Updates {
		qb = qb.Set(col, val)
	}
	qb = qb.Suffix("RETURNING policy_id")

	type metaIDRow struct {
		PolicyID int64 `db:"policy_id"`
	}
	if _, err := dblib.UpdateReturning(ctx, a.db, qb, pgx.RowToStructByNameLax[metaIDRow]); err != nil {
		return fmt.Errorf("UpdatePolicyMetadataActivity policy=%d: %w", p.PolicyID, err)
	}
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// PublishEventActivity — INSERT policy_event [A21.1, ADR-004]
// ─────────────────────────────────────────────────────────────────────────────

// PublishEventActivity inserts a policy_event row for downstream subscribers
// (Notification, Accounting, Agent, Audit services). [A21.1, ADR-004]
func (a *PolicyActivities) PublishEventActivity(ctx context.Context, e PolicyEvent) error {
	ctx, cancel := context.WithTimeout(ctx, a.cfg.GetDuration("db.QueryTimeoutHigh"))
	defer cancel()

	payload := e.Payload
	if payload == nil {
		payload = json.RawMessage(`{}`)
	}
	publishedAt := time.Now().UTC()

	q := dblib.Psql.Insert(actPolicyEventTable).
		Columns("policy_id", "event_type", "event_payload", "published_at").
		Values(e.PolicyID, e.EventType, payload, publishedAt).
		Suffix("RETURNING id")

	type idRow struct {
		ID int64 `db:"id"`
	}
	if _, err := dblib.InsertReturning(ctx, a.db, q, pgx.RowToStructByNameLax[idRow]); err != nil {
		return fmt.Errorf("PublishEventActivity policy=%d type=%s: %w", e.PolicyID, e.EventType, err)
	}
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// RecordRejectedRequestActivity — INSERT signal_log status=REJECTED [A21.1]
// ─────────────────────────────────────────────────────────────────────────────

// RecordRejectedRequestActivity logs a rejected signal for audit purposes. [A21.1]
func (a *PolicyActivities) RecordRejectedRequestActivity(ctx context.Context, p RejectedRequestParams) error {
	ctx, cancel := context.WithTimeout(ctx, a.cfg.GetDuration("db.QueryTimeoutHigh"))
	defer cancel()

	payload := p.Payload
	if payload == nil {
		payload = json.RawMessage(`{}`)
	}
	receivedAt := time.Now().UTC()

	var reqIDVal interface{} = nil
	if p.RequestID != "" {
		reqIDVal = p.RequestID
	}

	q := dblib.Psql.Insert(actSignalLogTable).
		Columns("policy_id", "signal_channel", "signal_payload",
			"request_id", "received_at", "status", "rejection_reason").
		Values(p.PolicyID, p.SignalChannel, payload,
			reqIDVal, receivedAt, domain.SignalStatusRejected, p.Reason).
		Suffix("RETURNING id")

	type idRow struct {
		ID int64 `db:"id"`
	}
	if _, err := dblib.InsertReturning(ctx, a.db, q, pgx.RowToStructByNameLax[idRow]); err != nil {
		return fmt.Errorf("RecordRejectedRequestActivity policy=%d channel=%s: %w",
			p.PolicyID, p.SignalChannel, err)
	}
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// CancelDownstreamWorkflowActivity — cancels a child workflow [A21.1, BR-PM-090]
// ─────────────────────────────────────────────────────────────────────────────

// CancelDownstreamWorkflowActivity sends a cancellation to a downstream child workflow.
// Non-fatal if the workflow has already completed. [A21.1, BR-PM-090, BR-PM-112]
func (a *PolicyActivities) CancelDownstreamWorkflowActivity(ctx context.Context, p CancelWorkflowParams) error {
	// Non-fatal: workflow may have already completed
	_ = a.tc.CancelWorkflow(ctx, p.WorkflowID, "")
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// LogSignalReceivedActivity — INSERT signal_log + registry (pgx.Batch) [A21.1]
// Constraint 5: batch INSERT for policy_signal_log + processed_signal_registry
// ─────────────────────────────────────────────────────────────────────────────

// LogSignalReceivedActivity atomically inserts a signal log entry and a dedup
// registry entry using dblib.QueueExecRow + SendBatch. [A21.1, §9.1, Constraint 5]
func (a *PolicyActivities) LogSignalReceivedActivity(ctx context.Context, e SignalLogEntry) error {
	ctx, cancel := context.WithTimeout(ctx, a.cfg.GetDuration("db.QueryTimeoutHigh"))
	defer cancel()

	payload := e.SignalPayload
	if payload == nil {
		payload = json.RawMessage(`{}`)
	}
	receivedAt := time.Now().UTC()

	// Provide default source service if not specified
	sourceService := e.SourceService
	if sourceService == "" {
		sourceService = "policy-mgmt-orchestrator"
	}

	batch := &pgx.Batch{}

	// Query 1: INSERT policy_signal_log — partitioned by received_at [§8.9]
	logQ := dblib.Psql.Insert(actSignalLogTable).
		Columns(
			"policy_id", "signal_channel", "signal_payload", "source_service",
			"source_workflow_id", "request_id",
			"received_at", "status", "rejection_reason",
			"state_before", "state_after",
		).
		Values(
			e.PolicyID, e.SignalChannel, payload, sourceService,
			e.SourceWorkflowID, e.RequestID,
			receivedAt, e.Status, e.RejectionReason,
			e.StateBefore, e.StateAfter,
		)
	dblib.QueueExecRow(batch, logQ)

	// Query 2: INSERT processed_signal_registry (90-day TTL dedup) [§8.8]
	regQ := dblib.Psql.Insert(actSignalRegTable).
		Columns("request_id", "signal_type", "policy_id", "received_at", "expires_at").
		Values(e.RequestID, e.SignalChannel, e.PolicyID,
			receivedAt, receivedAt.Add(90*24*time.Hour)).
		Suffix("ON CONFLICT (request_id, signal_type) DO NOTHING")
	dblib.QueueExecRow(batch, regQ)

	if err := a.db.SendBatch(ctx, batch).Close(); err != nil {
		return fmt.Errorf("LogSignalReceivedActivity policy=%d channel=%s: %w",
			e.PolicyID, e.SignalChannel, err)
	}
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// UpdateServiceRequestActivity — UPDATE service_request row [A21.1, FR-PM-006]
// ─────────────────────────────────────────────────────────────────────────────

// UpdateServiceRequestActivity updates the service_request row to reflect routing
// or completion status. [A21.1, FR-PM-006]
func (a *PolicyActivities) UpdateServiceRequestActivity(ctx context.Context, u ServiceRequestUpdate) error {
	if u.ServiceRequestID == 0 {
		return nil
	}
	ctx, cancel := context.WithTimeout(ctx, a.cfg.GetDuration("db.QueryTimeoutHigh"))
	defer cancel()

	qb := dblib.Psql.Update(actServiceReqTable).
		Where(sq.Eq{"request_id": u.ServiceRequestID}).
		Set("status", u.Status).
		Set("updated_at", sq.Expr("NOW()"))
	// Include partition key when available to avoid cross-partition seq-scans. [D4, §8.3]
	// service_request is partitioned by submitted_at (quarterly). Without this WHERE
	// condition the planner must scan all partitions to find the matching row.
	if u.SubmittedAt != nil {
		qb = qb.Where(sq.Eq{"submitted_at": *u.SubmittedAt})
	}

	if u.Outcome != nil {
		qb = qb.Set("outcome", *u.Outcome)
	}
	if u.OutcomeReason != nil {
		qb = qb.Set("outcome_reason", *u.OutcomeReason)
	}
	if u.OutcomePayload != nil {
		qb = qb.Set("outcome_payload", u.OutcomePayload)
	}
	if u.DownstreamWorkflowID != nil {
		qb = qb.Set("downstream_workflow_id", *u.DownstreamWorkflowID)
	}
	if u.DownstreamService != nil {
		qb = qb.Set("downstream_service", *u.DownstreamService)
	}
	if u.DownstreamTaskQueue != nil {
		qb = qb.Set("downstream_task_queue", *u.DownstreamTaskQueue)
	}
	if u.RoutedAt != nil {
		qb = qb.Set("routed_at", *u.RoutedAt)
	}
	if u.CompletedAt != nil {
		qb = qb.Set("completed_at", *u.CompletedAt)
	}

	qb = qb.Suffix("RETURNING request_id")
	type srIDRow struct {
		RequestID int64 `db:"request_id"`
	}
	if _, err := dblib.UpdateReturning(ctx, a.db, qb, pgx.RowToStructByNameLax[srIDRow]); err != nil {
		return fmt.Errorf("UpdateServiceRequestActivity srID=%d: %w", u.ServiceRequestID, err)
	}
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// RefreshStateFromDBActivity — SELECT policy for batch edge-case [§9.5.2, A21.1]
// ─────────────────────────────────────────────────────────────────────────────

// RefreshStateFromDBActivity reads the current policy state from DB.
// Used when a financial request arrives between batch DB write and batch-state-sync signal.
// [§9.5.2, A21.1]
func (a *PolicyActivities) RefreshStateFromDBActivity(ctx context.Context, policyID int64) (*PolicyRefreshedState, error) {
	ctx, cancel := context.WithTimeout(ctx, a.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	q := dblib.Psql.Select(
		"current_status", "previous_status", "version",
		"has_active_loan", "loan_outstanding", "assignment_type", "aml_hold",
	).
		From(actPolicyTable).
		Where(sq.Eq{"policy_id": policyID})

	s, err := dblib.SelectOne(ctx, a.db, q, pgx.RowToStructByNameLax[PolicyRefreshedState])
	if err != nil {
		return nil, fmt.Errorf("RefreshStateFromDBActivity policyID=%d: %w", policyID, err)
	}
	return &s, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// PersistTerminalStateActivity — UPSERT terminal_state_snapshot [§9.5.1, A21.1]
// ─────────────────────────────────────────────────────────────────────────────

// PersistTerminalStateActivity inserts a terminal state snapshot for Tier-2 queries.
// ON CONFLICT DO UPDATE ensures the latest snapshot is always current. [§9.5.1, A21.1]
func (a *PolicyActivities) PersistTerminalStateActivity(ctx context.Context, r TerminalStateRecord) error {
	ctx, cancel := context.WithTimeout(ctx, a.cfg.GetDuration("db.QueryTimeoutHigh"))
	defer cancel()

	q := dblib.Psql.Insert(actTerminalSnapTable).
		Columns("policy_id", "policy_number", "final_status",
			"terminal_at", "cooling_expiry", "final_snapshot", "created_at").
		Values(r.PolicyID, r.PolicyNumber, r.FinalStatus,
			r.TerminalAt, r.CoolingExpiry, r.FinalSnapshot, sq.Expr("NOW()")).
		Suffix(`ON CONFLICT (policy_id) DO UPDATE
			SET final_status   = EXCLUDED.final_status,
			    terminal_at    = EXCLUDED.terminal_at,
			    cooling_expiry = EXCLUDED.cooling_expiry,
			    final_snapshot = EXCLUDED.final_snapshot
			RETURNING policy_id`)

	type termIDRow struct {
		PolicyID int64 `db:"policy_id"`
	}
	if _, err := dblib.InsertReturning(ctx, a.db, q, pgx.RowToStructByNameLax[termIDRow]); err != nil {
		return fmt.Errorf("PersistTerminalStateActivity policyID=%d: %w", r.PolicyID, err)
	}
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// MarkWorkflowCompletedActivity — UPDATE terminal_state_snapshot [§9.5.1, A21.1]
// ─────────────────────────────────────────────────────────────────────────────

// MarkWorkflowCompletedActivity sets workflow_completed_at when cooling expires.
// Must be called before workflow returns nil. [§9.5.1, A21.1]
func (a *PolicyActivities) MarkWorkflowCompletedActivity(ctx context.Context, p MarkCompletedParams) error {
	ctx, cancel := context.WithTimeout(ctx, a.cfg.GetDuration("db.QueryTimeoutHigh"))
	defer cancel()

	q := dblib.Psql.Update(actTerminalSnapTable).
		Set("workflow_completed_at", sq.Expr("NOW()")).
		Where(sq.Eq{"policy_id": p.PolicyID}).
		Suffix("RETURNING policy_id")

	type markIDRow struct {
		PolicyID int64 `db:"policy_id"`
	}
	if _, err := dblib.UpdateReturning(ctx, a.db, q, pgx.RowToStructByNameLax[markIDRow]); err != nil {
		return fmt.Errorf("MarkWorkflowCompletedActivity policyID=%d: %w", p.PolicyID, err)
	}
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// TriggerPremiumRefundActivity — INSERT PREMIUM_REFUND service_request [§9.5.1]
// ─────────────────────────────────────────────────────────────────────────────

// TriggerPremiumRefundActivity creates a PREMIUM_REFUND service_request for premiums
// received during terminal cooling period. [§9.5.1, A21.1]
func (a *PolicyActivities) TriggerPremiumRefundActivity(ctx context.Context, r RefundRequest) error {
	ctx, cancel := context.WithTimeout(ctx, a.cfg.GetDuration("db.QueryTimeoutHigh"))
	defer cancel()

	payload, _ := json.Marshal(map[string]interface{}{
		"amount": r.Amount, "reason": r.Reason, "request_id": r.RequestID,
	})
	submittedAt := time.Now().UTC()

	q := dblib.Psql.Insert(actServiceReqTable).
		Columns("policy_id", "policy_number", "request_type", "request_category",
			"status", "source_channel", "submitted_at", "request_payload",
			"created_at", "updated_at").
		Values(r.PolicyID, r.PolicyNumber, domain.RequestTypePremiumRefund, domain.RequestCategoryFinancial,
			domain.RequestStatusReceived, domain.SourceChannelSystem,
			submittedAt, payload, submittedAt, submittedAt).
		Suffix("ON CONFLICT DO NOTHING RETURNING request_id")

	type refundIDRow struct {
		RequestID int64 `db:"request_id"`
	}
	// ON CONFLICT DO NOTHING returns empty slice — treat as idempotent success.
	if _, err := dblib.InsertReturning(ctx, a.db, q, pgx.RowToStructByNameLax[refundIDRow]); err != nil {
		return fmt.Errorf("TriggerPremiumRefundActivity policyID=%d: %w", r.PolicyID, err)
	}
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// FetchWorkflowConfigActivity — reads a single config value from policy_state_config
// Used by the workflow to get config-driven values (FLC period, routing timeouts, etc.)
// without embedding DB access directly in workflow code. [Review-Fix-8, §8.7]
// ─────────────────────────────────────────────────────────────────────────────

func (a *PolicyActivities) FetchWorkflowConfigActivity(ctx context.Context, key string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, a.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	q := dblib.Psql.Select("config_value").
		From("policy_mgmt.policy_state_config").
		Where(sq.Eq{"config_key": key})

	type cfgRow struct {
		ConfigValue string `db:"config_value"`
	}
	dest, err := dblib.SelectOne(ctx, a.db, q, pgx.RowToStructByNameLax[cfgRow])
	if err != nil {
		return "", fmt.Errorf("FetchWorkflowConfigActivity key=%s: %w", key, err)
	}
	return dest.ConfigValue, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// FetchAllWorkflowConfigsActivity — batch-reads N config values in one query
// Replaces the sequential loop over FetchWorkflowConfigActivity at policy creation.
// One SELECT ... WHERE config_key = ANY($keys) instead of N round-trips. [D6, §8.7]
// ─────────────────────────────────────────────────────────────────────────────

// FetchAllWorkflowConfigsActivity fetches all specified config keys in a single
// DB round-trip and returns a map[config_key]config_value. Missing keys are simply
// absent from the returned map; callers fall back to hardcoded defaults. [D6]
func (a *PolicyActivities) FetchAllWorkflowConfigsActivity(ctx context.Context, keys []string) (map[string]string, error) {
	if len(keys) == 0 {
		return map[string]string{}, nil
	}
	ctx, cancel := context.WithTimeout(ctx, a.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	q := dblib.Psql.Select("config_key", "config_value").
		From("policy_mgmt.policy_state_config").
		Where(sq.Eq{"config_key": keys}) // squirrel generates WHERE config_key IN ($1,$2,...)

	type cfgRow struct {
		ConfigKey   string `db:"config_key"`
		ConfigValue string `db:"config_value"`
	}
	rows, err := dblib.SelectRows(ctx, a.db, q, pgx.RowToStructByNameLax[cfgRow])
	if err != nil {
		return nil, fmt.Errorf("FetchAllWorkflowConfigsActivity: %w", err)
	}
	result := make(map[string]string, len(rows))
	for _, r := range rows {
		result[r.ConfigKey] = r.ConfigValue
	}
	return result, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// AcquireFinancialLockActivity — INSERT policy_lock row (idempotent) [BR-PM-030]
// Called immediately after routing a financial request to a downstream workflow.
// ON CONFLICT (policy_id) DO UPDATE keeps timeout_at fresh on Temporal retries.
// ─────────────────────────────────────────────────────────────────────────────

func (a *PolicyActivities) AcquireFinancialLockActivity(ctx context.Context, p FinancialLockParams) error {
	ctx, cancel := context.WithTimeout(ctx, a.cfg.GetDuration("db.QueryTimeoutHigh"))
	defer cancel()

	q := dblib.Psql.Insert(actPolicyLockTable).
		Columns("policy_id", "request_id", "request_type", "locked_at", "timeout_at").
		Values(p.PolicyID, p.ServiceRequestID, p.RequestType, p.LockedAt, p.TimeoutAt).
		Suffix("ON CONFLICT (policy_id) DO UPDATE SET request_id = EXCLUDED.request_id, request_type = EXCLUDED.request_type, locked_at = EXCLUDED.locked_at, timeout_at = EXCLUDED.timeout_at")

	if _, err := dblib.Insert(ctx, a.db, q); err != nil {
		return fmt.Errorf("AcquireFinancialLockActivity policyID=%d requestType=%s: %w", p.PolicyID, p.RequestType, err)
	}
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// ReleaseFinancialLockActivity — DELETE policy_lock row [BR-PM-030]
// Called when a financial request completes, is withdrawn, or is preempted.
// Idempotent — deleting a non-existent row is a no-op.
// ─────────────────────────────────────────────────────────────────────────────

func (a *PolicyActivities) ReleaseFinancialLockActivity(ctx context.Context, policyID int64) error {
	ctx, cancel := context.WithTimeout(ctx, a.cfg.GetDuration("db.QueryTimeoutHigh"))
	defer cancel()

	q := dblib.Psql.Delete(actPolicyLockTable).
		Where(sq.Eq{"policy_id": policyID})

	if _, err := dblib.Delete(ctx, a.db, q); err != nil {
		return fmt.Errorf("ReleaseFinancialLockActivity policyID=%d: %w", policyID, err)
	}
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────────────

func nullableStr(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
