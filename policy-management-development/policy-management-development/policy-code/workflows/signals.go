package workflows

// ============================================================================
// Signal & Workflow Type Definitions
// Source: §9.1 — EXACT signal names (kebab-case), A10.1 integration contract
// [FR-PM-001, FR-PM-002, FR-PM-005, A10.1, §9.1]
// ============================================================================

import (
	"encoding/json"
	"time"
)

// ─────────────────────────────────────────────────────────────────────────────
// Signal Channel Name Constants — EXACT names from §9.1
// [FR-PM-001, BR-PM-110..BR-PM-121]
// ─────────────────────────────────────────────────────────────────────────────

const (
	// Inbound: initial lifecycle signal from Policy Issue Service (SignalWithStart)
	SignalPolicyCreated = "policy-created" // [A10.1.1]

	// Inbound Financial Request Signals — from REST handlers via SignalWorkflow [§9.5.2]
	SignalSurrenderRequest       = "surrender-request"
	SignalLoanRequest            = "loan-request"
	SignalLoanRepayment          = "loan-repayment"
	SignalRevivalRequest         = "revival-request"
	SignalDeathNotification      = "death-notification"       // Preemptive; overrides SUSPENDED [BR-PM-112]
	SignalMaturityClaimRequest   = "maturity-claim-request"
	SignalSurvivalBenefitRequest = "survival-benefit-request"
	SignalCommutationRequest     = "commutation-request"
	SignalConversionRequest      = "conversion-request"
	SignalFLCRequest             = "flc-request"
	SignalForcedSurrenderTrigger = "forced-surrender-trigger" // From Loan Svc batch
	SignalNFRRequest             = "nfr-request"              // All non-financial requests

	// Per-service Completion Signals — from downstream services back to PM
	SignalSurrenderCompleted       = "surrender-completed"
	SignalForcedSurrenderCompleted = "forced-surrender-completed"
	SignalLoanCompleted            = "loan-completed"
	SignalLoanRepaymentCompleted   = "loan-repayment-completed"
	SignalRevivalCompleted         = "revival-completed"
	SignalClaimSettled             = "claim-settled"
	SignalCommutationCompleted     = "commutation-completed"
	SignalConversionCompleted      = "conversion-completed"
	SignalFLCCompleted             = "flc-completed"
	SignalNFRCompleted             = "nfr-completed"
	SignalOperationCompleted       = "operation-completed" // Generic fallback (older pattern)

	// Inbound System / Compliance Signals
	SignalPremiumPaid            = "premium-paid"
	SignalPaymentDishonored      = "payment-dishonored"
	SignalAMLFlagRaised          = "aml-flag-raised"          // → SUSPENDED [BR-PM-110]
	SignalAMLFlagCleared         = "aml-flag-cleared"         // Restore previous [BR-PM-111]
	SignalInvestigationStarted   = "investigation-started"    // DCI → DEATH_UNDER_INVESTIGATION [BR-PM-120]
	SignalInvestigationConcluded = "investigation-concluded"  // [BR-PM-121]
	SignalLoanBalanceUpdated     = "loan-balance-updated"     // Metadata-only update
	SignalConversionReversed     = "conversion-reversed"      // Cheque bounce [BR-CHQ-001]
	SignalAdminVoid              = "admin-void"               // → VOID [BR-PM-073]
	SignalCustomerIDMerge        = "customer-id-merge"        // Metadata update
	SignalVoluntaryPaidUpRequest = "voluntary-paidup-request" // → PAID_UP or VOID [BR-PM-060, BR-PM-061]
	SignalWithdrawalRequest      = "withdrawal-request"       // Cancel active request [BR-PM-090]
	SignalDisputeRegistered      = "dispute-registered"       // Advisory flag only [BR-PM-113, ADR-003]
	SignalDisputeResolved        = "dispute-resolved"         // Advisory flag clear
	SignalBatchStateSync         = "batch-state-sync"         // In-memory only [§9.5.2]
	SignalReopenRequest          = "reopen-request"           // Exit terminal cooling + CAN [§9.5.1]
)

// ─────────────────────────────────────────────────────────────────────────────
// Query Handler Name Constants — EXACT names from §9.1 (kebab-case)
// [FR-PM-004]
// ─────────────────────────────────────────────────────────────────────────────

const (
	QueryGetPolicyStatus    = "get-policy-status"
	QueryGetPendingRequests = "get-pending-requests"
	QueryIsRequestEligible  = "is-request-eligible"
	QueryGetPolicySummary   = "get-policy-summary"
	QueryGetActiveLock      = "get-active-lock"
	QueryGetStatusHistory   = "get-status-history"
	QueryGetWorkflowHealth  = "get-workflow-health"
)

// ─────────────────────────────────────────────────────────────────────────────
// Workflow State Structs — EXACT field names from §9.1
// Serialized across Continue-As-New boundaries [FR-PM-002]
// ─────────────────────────────────────────────────────────────────────────────

// PolicyLifecycleState is the full in-memory state of one PolicyLifecycleWorkflow.
// Passed as the input to the new workflow instance on Continue-As-New. [FR-PM-002]
type PolicyLifecycleState struct {
	PolicyNumber                   string               `json:"policy_number"`
	PolicyID                       string               `json:"policy_id"`      // UUID from Policy Issue (audit cross-ref)
	PolicyDBID                     int64                `json:"policy_db_id"`   // BIGINT from PM seq_policy_id [A13]
	CurrentStatus                  string               `json:"current_status"`
	PreviousStatus                 string               `json:"previous_status"`
	PreviousStatusBeforeSuspension string               `json:"previous_status_before_suspension"` // AML revert [BR-PM-110/111]
	Encumbrances                   EncumbranceFlags     `json:"encumbrances"`
	DisplayStatus                  string               `json:"display_status"` // Computed: status + encumbrances
	Version                        int64                `json:"version"`        // Optimistic locking
	Metadata                       PolicyMetadata       `json:"metadata"`
	PendingRequests                []PendingRequest     `json:"pending_requests"`
	ActiveLock                     *FinancialLock       `json:"active_lock,omitempty"`
	ProcessedSignalIDs             map[string]time.Time `json:"processed_signal_ids"` // Dedup, 90-day TTL
	EventCount                     int                  `json:"event_count"`          // For CAN threshold [FR-PM-002]
	LastCANTime                    time.Time            `json:"last_can_time"`
	LastTransitionAt               time.Time            `json:"last_transition_at"`
	CachedConfig                   map[string]string    `json:"cached_config,omitempty"` // lazy-loaded config cache [Review-Fix-3]
	// FLCExpiryAt is the absolute time when the FLC timer goroutine should fire.
	// Persisted so the goroutine can be respawned after Continue-As-New (goroutines
	// are lost on CAN; without this field the FLC transition would never fire for
	// policies that cross a CAN boundary during the free-look period). [D1]
	FLCExpiryAt                    time.Time            `json:"flc_expiry_at,omitempty"`
}

// PolicyMetadata holds policy-level data needed by the workflow for state gate
// decisions, eligibility checks, and activity calls. [§9.1]
type PolicyMetadata struct {
	CustomerID                  int64     `json:"customer_id"`
	ProductCode                 string    `json:"product_code"`
	ProductType                 string    `json:"product_type"`    // PLI or RPLI
	SumAssured                  float64   `json:"sum_assured"`
	CurrentPremium              float64   `json:"current_premium"`
	PremiumMode                 string    `json:"premium_mode"`    // MONTHLY, QUARTERLY, HALF_YEARLY, YEARLY
	BillingMethod               string    `json:"billing_method"`  // CASH or PAY_RECOVERY [BR-PM-074]
	IssueDate                   time.Time `json:"issue_date"`
	MaturityDate                time.Time `json:"maturity_date"`
	PaidToDate                  time.Time `json:"paid_to_date"`
	AgentID                     *int64     `json:"agent_id,omitempty"`              // Nullable BIGINT [Review-Fix-5]
	LoanOutstanding             float64    `json:"loan_outstanding"`
	AssignmentStatus            string     `json:"assignment_status"`
	PremiumsPaidMonths          int        `json:"premiums_paid_months"`           // For paid-up calc [BR-PM-061]
	TotalPremiumsMonths         int        `json:"total_premiums_months"`
	RemissionExpiryDate         *time.Time `json:"remission_expiry_date,omitempty"` // Nullable [Review-Fix-5]
	PayRecoveryProtectionExpiry *time.Time `json:"pay_recovery_protection_expiry,omitempty"` // Nullable, first_unpaid + 12mo [BR-PM-074, Review-Fix-5]
	SBInstallmentsPaid          int       `json:"sb_installments_paid"`
	NominationStatus            string    `json:"nomination_status"`
	IsDistanceMarketing         bool      `json:"is_distance_marketing"` // 30d FLC for distance-marketing products [Review-Fix-9]
	WorkflowID                  string    `json:"workflow_id"` // plw-{policy_number}
}

// PendingRequest tracks a routed in-flight request waiting for a completion signal. [§9.1]
type PendingRequest struct {
	RequestID          string     `json:"request_id"`           // Dedup key (BIGINT as string or UUID)
	ServiceRequestID   int64      `json:"service_request_id"`   // BIGINT from service_request table
	RequestType        string     `json:"request_type"`
	RequestCategory    string     `json:"request_category"`     // FINANCIAL or NON_FINANCIAL
	DownstreamWorkflow string     `json:"downstream_workflow"`  // Child workflow ID
	RoutedAt           time.Time  `json:"routed_at"`
	TimeoutAt          time.Time  `json:"timeout_at"`
	SubmittedAt        *time.Time `json:"submitted_at,omitempty"` // Partition key for service_request [D4]
}

// FinancialLock represents an exclusive lock held by a financial request. [BR-PM-030]
// Only one active lock at a time; death-notification and NFR bypass this.
type FinancialLock struct {
	RequestID   string    `json:"request_id"`
	RequestType string    `json:"request_type"`
	LockedAt    time.Time `json:"locked_at"`
	TimeoutAt   time.Time `json:"timeout_at"`
}

// EncumbranceFlags groups all encumbrance state; passed to isStateEligible(). [§9.1]
type EncumbranceFlags struct {
	HasActiveLoan  bool   `json:"has_active_loan"`  // Blocks new LOAN requests
	AssignmentType string `json:"assignment_type"`  // NONE, ABSOLUTE, CONDITIONAL
	AMLHold        bool   `json:"aml_hold"`         // true when SUSPENDED [BR-PM-110]
	DisputeFlag    bool   `json:"dispute_flag"`     // Advisory only; never blocks [BR-PM-113, ADR-003]
}

// ─────────────────────────────────────────────────────────────────────────────
// Integration Contract Types — Policy Issue Service → PM
// Source: A10.1
// ─────────────────────────────────────────────────────────────────────────────

// PolicyCreatedSignal is sent by Policy Issue Service on the "policy-created" channel.
// PolicyID is Policy Issue's own UUID — kept for audit trail only.
// PM generates its own BIGINT policy_id via seq_policy_id. [A10.1.4]
type PolicyCreatedSignal struct {
	RequestID    string        `json:"request_id"`    // UUID — idempotency key for dedup
	PolicyID     string        `json:"policy_id"`     // UUID from Policy Issue (audit cross-ref)
	PolicyNumber string        `json:"policy_number"`
	Metadata     PolicyMetadata `json:"metadata"`
}

// StartPMLifecycleInput is the workflow input for SignalWithStart from Policy Issue.
// Policy Issue constructs InitialState with CurrentStatus="FREE_LOOK_ACTIVE". [A10.1.3]
type StartPMLifecycleInput struct {
	Signal       PolicyCreatedSignal  `json:"signal"`
	InitialState PolicyLifecycleState `json:"initial_state"`
}

// ─────────────────────────────────────────────────────────────────────────────
// Child Workflow Input / Output — Standard contract with downstream services
// Source: A10.1B, A10.1C, Constraint 1
// ─────────────────────────────────────────────────────────────────────────────

// ChildWorkflowInput is the input sent to all downstream service workflows via
// ExecuteChildWorkflow. Exact field names required by A10.1B. [Constraint 1]
type ChildWorkflowInput struct {
	RequestID        string          `json:"request_id"`         // Idempotency key
	PolicyNumber     string          `json:"policy_number"`
	PolicyDBID       int64           `json:"policy_db_id"`       // BIGINT PM policy_id
	ServiceRequestID int64           `json:"service_request_id"` // BIGINT from service_request
	RequestType      string          `json:"request_type"`
	RequestPayload   json.RawMessage `json:"request_payload"`    // Original JSONB from handler
	TimeoutAt        time.Time       `json:"timeout_at"`
}

// OperationCompletedSignal is sent by downstream services to PM on completion. [A10.1C]
type OperationCompletedSignal struct {
	RequestID       string          `json:"request_id"`
	RequestType     string          `json:"request_type"`
	Outcome         string          `json:"outcome"`                    // APPROVED, REJECTED, WITHDRAWN, TIMEOUT
	StateTransition string          `json:"state_transition,omitempty"` // e.g. "PENDING_SURRENDER→SURRENDERED"
	OutcomePayload  json.RawMessage `json:"outcome_payload,omitempty"`
	CompletedAt     time.Time       `json:"completed_at"`
}

// ─────────────────────────────────────────────────────────────────────────────
// Inbound Request Signal — from REST handlers after creating service_request row
// Source: §9.5.2
// ─────────────────────────────────────────────────────────────────────────────

// PolicyRequestSignal is sent by REST handlers to the PLW workflow after inserting
// the service_request row. The workflow fetches full request details from DB. [§9.5.2]
//
// IdempotencyKey is the UUID from the X-Idempotency-Key header — used as the dedup
// key in ProcessedSignalIDs and as the child workflow ID fragment per Constraint 1.
// ServiceRequestID is the BIGINT PK from service_request — used for DB updates only.
// SubmittedAt is the service_request.submitted_at partition key; when present it is
// included in WHERE clauses to prevent cross-partition seq-scans. [D4, §8.3]
// [Constraint 1, §9.5.2, Review-Fix-11, Review-Fix-3]
type PolicyRequestSignal struct {
	ServiceRequestID int64      `json:"request_id"`        // BIGINT from service_request table
	IdempotencyKey   string     `json:"idempotency_key"`   // UUID from X-Idempotency-Key header [Review-Fix-11]
	RequestType      string     `json:"request_type"`
	RequestCategory  string     `json:"request_category"`  // FINANCIAL or NON_FINANCIAL
	SourceChannel    string     `json:"source_channel"`
	SubmittedBy      *int64     `json:"submitted_by,omitempty"`
	SubmittedAt      *time.Time `json:"submitted_at,omitempty"` // Partition key for service_request [D4]
}

// ─────────────────────────────────────────────────────────────────────────────
// Inbound System / Compliance Signal Payloads
// ─────────────────────────────────────────────────────────────────────────────

// PremiumPaidSignal — updates PaidToDate; may trigger lapse revival. [§9.1]
type PremiumPaidSignal struct {
	RequestID     string    `json:"request_id"`
	PremiumAmount float64   `json:"premium_amount"`
	PaymentDate   time.Time `json:"payment_date"`
	NewPaidToDate time.Time `json:"new_paid_to_date"`
}

// PaymentDishonoredSignal — reverses PaidToDate; triggers lapse transition. [§9.1]
type PaymentDishonoredSignal struct {
	RequestID      string    `json:"request_id"`
	DishonoredDate time.Time `json:"dishonored_date"`
	Reason         string    `json:"reason"`
}

// AMLFlagRaisedSignal — triggers SUSPENDED; saves PreviousStatusBeforeSuspension. [BR-PM-110]
type AMLFlagRaisedSignal struct {
	RequestID    string `json:"request_id"`
	Reason       string `json:"reason"`
	AuthorityRef string `json:"authority_ref,omitempty"`
}

// AMLFlagClearedSignal — restores PreviousStatusBeforeSuspension. [BR-PM-111]
type AMLFlagClearedSignal struct {
	RequestID string    `json:"request_id"`
	ClearedBy string    `json:"cleared_by"`
	ClearedAt time.Time `json:"cleared_at"`
}

// InvestigationStartedSignal — DEATH_CLAIM_INTIMATED → DEATH_UNDER_INVESTIGATION. [BR-PM-120]
type InvestigationStartedSignal struct {
	RequestID      string `json:"request_id"`
	InvestigatorID string `json:"investigator_id,omitempty"`
}

// InvestigationConcludedSignal — DEATH_UNDER_INVESTIGATION → settled or revert. [BR-PM-121]
type InvestigationConcludedSignal struct {
	RequestID string `json:"request_id"`
	Outcome   string `json:"outcome"` // "CONFIRMED" → DEATH_CLAIM_SETTLED; "REJECTED" → revert
}

// LoanBalanceUpdatedSignal — updates LoanOutstanding in metadata. No state change. [§9.1]
type LoanBalanceUpdatedSignal struct {
	RequestID       string  `json:"request_id"`
	LoanOutstanding float64 `json:"loan_outstanding"`
}

// ConversionReversedSignal — CONVERTED → PreviousStatus (cheque bounce). [BR-CHQ-001]
type ConversionReversedSignal struct {
	RequestID string `json:"request_id"`
	Reason    string `json:"reason"`
}

// AdminVoidSignal — admin force-transition → VOID; cancels all pending. [BR-PM-073]
type AdminVoidSignal struct {
	RequestID    string `json:"request_id"`
	Reason       string `json:"reason"`
	AuthorizedBy int64  `json:"authorized_by"`
}

// CustomerIDMergeSignal — updates customer_id in policy metadata. [§9.1]
type CustomerIDMergeSignal struct {
	RequestID     string `json:"request_id"`
	OldCustomerID int64  `json:"old_customer_id"`
	NewCustomerID int64  `json:"new_customer_id"`
}

// VoluntaryPaidUpSignal — → PAID_UP (value ≥ 10K) or VOID (value < 10K). [BR-PM-060, BR-PM-061]
type VoluntaryPaidUpSignal struct {
	RequestID        string  `json:"request_id"`
	ServiceRequestID int64   `json:"service_request_id"`
	PaidUpValue      float64 `json:"paid_up_value"`
}

// WithdrawalRequestSignal — cancels active downstream workflow + releases lock. [BR-PM-090]
type WithdrawalRequestSignal struct {
	RequestID         string `json:"request_id"`
	TargetRequestID   string `json:"target_request_id"`  // The request being withdrawn
	WithdrawalReason  string `json:"withdrawal_reason"`
}

// DisputeSignal — advisory flag set/clear. Never blocks requests. [BR-PM-113, ADR-003]
type DisputeSignal struct {
	RequestID  string `json:"request_id"`
	DisputeRef string `json:"dispute_ref,omitempty"`
	Reason     string `json:"reason,omitempty"`
}

// BatchStateSyncSignal — in-memory state sync from batch job. NO DB writes by workflow. [§9.5.2]
type BatchStateSyncSignal struct {
	NewStatus     string    `json:"new_status"`
	ScanType      string    `json:"scan_type"`
	ScheduledDate time.Time `json:"scheduled_date"`
}

// ReopenRequestSignal — exits terminal cooling; triggers Continue-As-New back to main loop. [§9.5.1]
type ReopenRequestSignal struct {
	RequestID    string `json:"request_id"`
	ReopenReason string `json:"reopen_reason"`
	AuthorizedBy int64  `json:"authorized_by"`
}

// ─────────────────────────────────────────────────────────────────────────────
// Query Result Types — returned by workflow query handlers
// Source: §9.1 [FR-PM-004]
// ─────────────────────────────────────────────────────────────────────────────

// PolicyStatusQueryResult is returned by "get-policy-status". [FR-PM-004]
type PolicyStatusQueryResult struct {
	CurrentStatus  string         `json:"current_status"`
	PreviousStatus string         `json:"previous_status"`
	DisplayStatus  string         `json:"display_status"`
	EffectiveFrom  time.Time      `json:"effective_from"` // Time of last transition
	Metadata       PolicyMetadata `json:"metadata"`
}

// IsRequestEligibleResult is returned by "is-request-eligible". [FR-PM-004]
type IsRequestEligibleResult struct {
	Eligible     bool   `json:"eligible"`
	Reason       string `json:"reason,omitempty"`
	LockConflict bool   `json:"lock_conflict,omitempty"`
}

// WorkflowHealthResult is returned by "get-workflow-health". [FR-PM-004]
type WorkflowHealthResult struct {
	EventCount          int       `json:"event_count"`
	LastCANTime         time.Time `json:"last_can_time"`
	PendingRequestCount int       `json:"pending_request_count"`
	HasActiveLock       bool      `json:"has_active_lock"`
}

// ─────────────────────────────────────────────────────────────────────────────
// Downstream Workflow Type Mapping
// Source: Constraint 1, FR-PM-009 — EXACT workflow type names. No TODOs.
// ─────────────────────────────────────────────────────────────────────────────

// DownstreamWorkflowTypeForRequest returns the downstream service workflow type
// for the given request type. Constraint 1: EXACT names from FR-PM-009 / §9.1.
func DownstreamWorkflowTypeForRequest(requestType string) string {
	switch requestType {
	case "SURRENDER":
		return "SurrenderProcessingWorkflow"
	case "FORCED_SURRENDER":
		return "ForcedSurrenderWorkflow"
	case "LOAN":
		return "LoanProcessingWorkflow"
	case "LOAN_REPAYMENT":
		return "LoanRepaymentWorkflow"
	case "REVIVAL":
		return "InstallmentRevivalWorkflow"
	case "DEATH_CLAIM":
		return "DeathClaimSettlementWorkflow"
	case "MATURITY_CLAIM":
		return "MaturityClaimWorkflow"
	case "SURVIVAL_BENEFIT":
		return "SurvivalBenefitClaimWorkflow"
	case "COMMUTATION":
		return "CommutationRequestWorkflow"
	case "CONVERSION":
		return "ConversionMainWorkflow"
	case "FLC":
		return "FreelookCancellationWorkflow"
	case "ASSIGNMENT":
		return "AssignmentProcessingWorkflow"
	case "PREMIUM_REFUND":
		return "PremiumRefundWorkflow"
	default: // NOMINATION_CHANGE, ADDRESS_CHANGE, BILLING_METHOD_CHANGE, DUPLICATE_BOND
		return "NFRProcessingWorkflow"
	}
}

// DownstreamChildIDPrefix returns the child workflow ID prefix for the request type.
// Constraint 1: EXACT child ID prefixes from FR-PM-009.
func DownstreamChildIDPrefix(requestType string) string {
	switch requestType {
	case "SURRENDER":
		return "sur"
	case "FORCED_SURRENDER":
		return "fs"
	case "LOAN":
		return "loan"
	case "LOAN_REPAYMENT":
		return "lrp"
	case "REVIVAL":
		return "rev"
	case "DEATH_CLAIM":
		return "dc"
	case "MATURITY_CLAIM":
		return "mc"
	case "SURVIVAL_BENEFIT":
		return "sb"
	case "COMMUTATION":
		return "com"
	case "CONVERSION":
		return "cnv"
	case "FLC":
		return "flc"
	default:
		return "nfr"
	}
}
