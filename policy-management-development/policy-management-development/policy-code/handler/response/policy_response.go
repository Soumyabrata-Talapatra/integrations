package response

import (
	"policy-management/core/domain"
	"policy-management/core/port"
)

// ============================================================================
// Policy Response DTOs
// Source: Swagger components/schemas: PolicyStatusResponse, PolicySummaryResponse,
//         PolicyHistoryResponse, StateGateResponse, BatchStatusResponse, Encumbrances
// Used by: GET /policies/{pn}/status, /summary, /history, /state-gate/{type}, /batch-status
// ============================================================================

// Encumbrances represents the policy encumbrance flags in API responses.
// Source: Swagger Encumbrances schema (lines 1554–1577).
// loan_id and assignee_id are nullable BIGINT references to external services.
// They are populated from the workflow's in-memory state (Tier 1 query) and
// will be nil in terminal-snapshot fallback (Tier 2 — DB only tracks flags not IDs).
type Encumbrances struct {
	HasActiveLoan   bool    `json:"has_active_loan"`
	LoanID          *int64  `json:"loan_id,omitempty"`          // BIGINT from loan service; nil if no active loan
	LoanOutstanding float64 `json:"loan_outstanding,omitempty"`
	AssignmentType  string  `json:"assignment_type"`            // NONE|ABSOLUTE|CONDITIONAL
	AssigneeID      *int64  `json:"assignee_id,omitempty"`      // BIGINT from NFS service; nil if not assigned
	AMLHold         bool    `json:"aml_hold"`
	DisputeFlag     bool    `json:"dispute_flag"`
}

// NewEncumbrances builds Encumbrances from a domain.Policy (Tier-2 terminal snapshot path).
// LoanID and AssigneeID are not stored in the policy table — they will be nil here.
// Tier-1 (QueryWorkflow) populates these from the wfEncumbrances struct in phase 5.
func NewEncumbrances(p domain.Policy) Encumbrances {
	return Encumbrances{
		HasActiveLoan:   p.HasActiveLoan,
		LoanOutstanding: p.LoanOutstanding,
		AssignmentType:  p.AssignmentType,
		AMLHold:         p.AMLHold,
		DisputeFlag:     p.DisputeFlag,
	}
}

// ── GET /policies/{pn}/status ────────────────────────────────────────────────

// PolicyStatusData is the data payload for a policy status response.
type PolicyStatusData struct {
	PolicyID         int64        `json:"policy_id"`                  // Gap-2: internal BIGINT for callers
	PolicyNumber     string       `json:"policy_number"`
	LifecycleStatus  string       `json:"lifecycle_status"`
	PreviousStatus   *string      `json:"previous_status,omitempty"`
	Encumbrances     Encumbrances `json:"encumbrances"`
	DisplayStatus    string       `json:"display_status"`
	EffectiveFrom    string       `json:"effective_from"`
	Version          int64        `json:"version"`
	UpdatedAt        string       `json:"updated_at"`
}

// NewPolicyStatusData builds PolicyStatusData from a domain.Policy.
func NewPolicyStatusData(p domain.Policy) PolicyStatusData {
	return PolicyStatusData{
		PolicyID:        p.PolicyID,
		PolicyNumber:    p.PolicyNumber,
		LifecycleStatus: p.CurrentStatus,
		PreviousStatus:  p.PreviousStatus,
		Encumbrances:    NewEncumbrances(p),
		DisplayStatus:   p.DisplayStatus,
		EffectiveFrom:   p.EffectiveFrom.Format("2006-01-02T15:04:05Z07:00"),
		Version:         p.Version,
		UpdatedAt:       p.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// PolicyStatusResponse — GET /api/v1/policies/{pn}/status
// Source: Tier 1 QueryWorkflow("getFullState") → Tier 2 terminal_state_snapshot
// [FR-PM-001] [AD-011]
type PolicyStatusResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	Data                      PolicyStatusData `json:"data"`
}

// ── GET /policies/{pn}/summary ───────────────────────────────────────────────

// PolicySummaryData is the full policy summary payload.
type PolicySummaryData struct {
	PolicyID           int64        `json:"policy_id"`                  // Gap-2: internal BIGINT for callers
	PolicyNumber       string       `json:"policy_number"`
	CustomerID         int64        `json:"customer_id"`
	ProductCode        string       `json:"product_code"`
	ProductType        string       `json:"product_type"`
	LifecycleStatus    string       `json:"lifecycle_status"`
	DisplayStatus      string       `json:"display_status"`
	SumAssured         float64      `json:"sum_assured"`
	CurrentPremium     float64      `json:"current_premium"`
	PremiumMode        string       `json:"premium_mode"`
	BillingMethod      string       `json:"billing_method"`
	IssueDate          string       `json:"issue_date"`
	MaturityDate       *string      `json:"maturity_date,omitempty"`
	PaidToDate         string       `json:"paid_to_date"`
	NextPremiumDueDate *string      `json:"next_premium_due_date,omitempty"`
	Encumbrances       Encumbrances `json:"encumbrances"`
	PaidUpValue        *float64     `json:"paid_up_value,omitempty"`
	AgentID            *int64       `json:"agent_id,omitempty"`
	Version            int64        `json:"version"`
}

// NewPolicySummaryData builds PolicySummaryData from a domain.Policy.
func NewPolicySummaryData(p domain.Policy) PolicySummaryData {
	d := PolicySummaryData{
		PolicyID:        p.PolicyID,
		PolicyNumber:    p.PolicyNumber,
		CustomerID:      p.CustomerID,
		ProductCode:     p.ProductCode,
		ProductType:     p.ProductType,
		LifecycleStatus: p.CurrentStatus,
		DisplayStatus:   p.DisplayStatus,
		SumAssured:      p.SumAssured,
		CurrentPremium:  p.CurrentPremium,
		PremiumMode:     p.PremiumMode,
		BillingMethod:   p.BillingMethod,
		IssueDate:       p.IssueDate.Format("2006-01-02"),
		PaidToDate:      p.PaidToDate.Format("2006-01-02"),
		Encumbrances:    NewEncumbrances(p),
		PaidUpValue:     p.PaidUpValue,
		AgentID:         p.AgentID,
		Version:         p.Version,
	}
	if p.MaturityDate != nil {
		s := p.MaturityDate.Format("2006-01-02")
		d.MaturityDate = &s
	}
	if p.NextPremiumDueDate != nil {
		s := p.NextPremiumDueDate.Format("2006-01-02")
		d.NextPremiumDueDate = &s
	}
	return d
}

// PolicySummaryResponse — GET /api/v1/policies/{pn}/summary
// [FR-PM-001] [AD-011]
type PolicySummaryResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	Data                      PolicySummaryData `json:"data"`
}

// ── GET /policies/{pn}/history ───────────────────────────────────────────────

// PolicyTransition is a single state transition in the history.
type PolicyTransition struct {
	ID                  int64   `json:"id"`
	FromStatus          *string `json:"from_status,omitempty"`
	ToStatus            string  `json:"to_status"`
	TransitionReason    string  `json:"transition_reason"`
	TriggeredByService  string  `json:"triggered_by_service"`
	RequestID           *int64  `json:"request_id,omitempty"`
	EffectiveDate       string  `json:"effective_date"`
}

// NewPolicyTransition converts a domain.PolicyStatusHistory to a response DTO.
func NewPolicyTransition(h domain.PolicyStatusHistory) PolicyTransition {
	return PolicyTransition{
		ID:                 h.ID,
		FromStatus:         h.FromStatus,
		ToStatus:           h.ToStatus,
		TransitionReason:   h.TransitionReason,
		TriggeredByService: h.TriggeredByService,
		RequestID:          h.RequestID,
		EffectiveDate:      h.EffectiveDate.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// PolicyHistoryData is the full history payload.
type PolicyHistoryData struct {
	PolicyNumber     string             `json:"policy_number"`
	TotalTransitions int                `json:"total_transitions"`
	Transitions      []PolicyTransition `json:"transitions"`
}

// PolicyHistoryResponse — GET /api/v1/policies/{pn}/history
// Always DB-only (policy_status_history table). [FR-PM-002]
type PolicyHistoryResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	port.MetaDataResponse     `json:",inline"`
	Data                      PolicyHistoryData `json:"data"`
}

// ── GET /policies/{pn}/state-gate/{type} ─────────────────────────────────────

// StateGateData is the state gate check result payload.
type StateGateData struct {
	PolicyNumber             string       `json:"policy_number"`
	RequestType              string       `json:"request_type"`
	StateGatePassed          bool         `json:"state_gate_passed"`
	CurrentStatus            string       `json:"current_status"`
	AllowedStatuses          []string     `json:"allowed_statuses"`
	Encumbrances             Encumbrances `json:"encumbrances"`
	HasPendingFinancialLock  bool         `json:"has_pending_financial_lock"`
	RejectionReason          *string      `json:"rejection_reason,omitempty"`
}

// StateGateResponse — GET /api/v1/policies/{pn}/state-gate/{type}
// Tier 1: QueryWorkflow("getStateGate") → Tier 2: terminal_state_snapshot [AD-011]
// [BR-PM-011..023]
type StateGateResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	Data                      StateGateData `json:"data"`
}

// ── GET /policies/batch-status ───────────────────────────────────────────────

// BatchPolicyStatus is the status for one policy in a batch query.
type BatchPolicyStatus struct {
	PolicyNumber      string `json:"policy_number"`
	LifecycleStatus   string `json:"lifecycle_status"`
	DisplayStatus     string `json:"display_status"`
	HasPendingRequest bool   `json:"has_pending_request"`
}

// BatchStatusData is the batch status response payload.
type BatchStatusData struct {
	Policies []BatchPolicyStatus `json:"policies"`
}

// BatchStatusResponse — GET /api/v1/policies/batch-status
// Parallel QueryWorkflow per policy. [FR-PM-001] [AD-011]
type BatchStatusResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	Data                      BatchStatusData `json:"data"`
}
