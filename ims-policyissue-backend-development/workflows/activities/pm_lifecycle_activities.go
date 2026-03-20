package activities

import (
	"context"
	"fmt"
	"time"

	"policy-issue-service/repo/postgres"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
)

// PMSignalStore is the minimal DB interface needed by PMLifecycleActivities.
// *postgres.ProposalRepository satisfies this interface.
type PMSignalStore interface {
	// MarkPMSignalSent records a successful PM SignalWithStart.
	MarkPMSignalSent(ctx context.Context, policyNumber, plwWorkflowID string) error
	// MarkPMSignalFailed records a failed attempt, incrementing the counter.
	MarkPMSignalFailed(ctx context.Context, policyNumber, errMsg string) error
	// IncrementPMSignalAttempts bumps the attempt counter without changing the status.
	// Called before the SignalWithStart call so each attempt is always counted,
	// regardless of outcome. Separate from MarkPMSignalFailed so that a fresh
	// attempt doesn't immediately write status='FAILED'.
	IncrementPMSignalAttempts(ctx context.Context, policyNumber string) error
	// FindUnsignalledPolicies returns PENDING/FAILED rows older than gracePeriod.
	FindUnsignalledPolicies(ctx context.Context, gracePeriod time.Duration, maxAttempts int) ([]postgres.PMSignalTarget, error)
}

// TemporalSignaller wraps the single Temporal client method used by PMLifecycleActivities.
// client.Client satisfies this interface.
type TemporalSignaller interface {
	SignalWithStartWorkflow(
		ctx context.Context,
		workflowID string,
		signalName string,
		signalArg interface{},
		options client.StartWorkflowOptions,
		workflow interface{},
		workflowArgs ...interface{},
	) (client.WorkflowRun, error)
}

// PMLifecycleActivities contains activities for signalling the Policy Manager (PM) service.
type PMLifecycleActivities struct {
	store       PMSignalStore
	signaller   TemporalSignaller
	pmTaskQueue string
}

// NewPMLifecycleActivities creates a PMLifecycleActivities instance.
// Both *postgres.ProposalRepository and client.Client satisfy the interface parameters.
func NewPMLifecycleActivities(
	store PMSignalStore,
	signaller TemporalSignaller,
	pmTaskQueue string,
) *PMLifecycleActivities {
	return &PMLifecycleActivities{
		store:       store,
		signaller:   signaller,
		pmTaskQueue: pmTaskQueue,
	}
}

// StartPMLifecycleInput is the input for StartPMLifecycleActivity.
type StartPMLifecycleInput struct {
	// PolicyNumber uniquely identifies the issued policy (e.g. "PLI/2026/GJ/000001").
	PolicyNumber string `json:"policy_number"`
	// PolicyType is "PLI" or "RPLI", used to build the PM workflow ID.
	PolicyType string `json:"policy_type"`
	// Additional parameters from PolicyIssuanceWorkflow
	ProposalID              int64  `json:"proposal_id"`
	ProposalNumber          string `json:"proposal_number"`
	CustomerID              string `json:"customer_id"`
	ProductCode             string `json:"product_code"`
	ProductType             string `json:"product_type,omitempty"`
	SumAssured              float64 `json:"sum_assured"`
	PolicyTerm              int    `json:"policy_term"`
	AgeAtEntry              int    `json:"age_at_entry,omitempty"`
	Gender                  string `json:"gender,omitempty"`
	PremiumPaymentFrequency string `json:"premium_payment_frequency"`
	AgeProofType            string `json:"age_proof_type,omitempty"`
	InsuredState            string `json:"insured_state,omitempty"`
	ProposalDate            time.Time `json:"proposal_date,omitempty"`
	// Additional fields from proposals table
	SpouseCustomerID        *int64  `json:"spouse_customer_id,omitempty"`
	ProposerCustomerID      *int64  `json:"proposer_customer_id,omitempty"`
	IsProposerSameAsInsured bool    `json:"is_proposer_same_as_insured"`
	PremiumPayerType        string  `json:"premium_payer_type,omitempty"`
	PayerCustomerID         *int64  `json:"payer_customer_id,omitempty"`
	PremiumCeasingAge       *int    `json:"premium_ceasing_age,omitempty"`
	BasePremium             float64 `json:"base_premium"`
	GSTAmount               float64 `json:"gst_amount"`
	TotalPremium            float64 `json:"total_premium"`
	AnnualPremiumEquivalent float64 `json:"annual_premium_equivalent"`
	ModalPremium            float64 `json:"modal_premium"`
	AdditionalPremium       float64 `json:"additional_premium"`
	EntryPath               string  `json:"entry_path"`
	Channel                 string  `json:"channel"`
	CurrentStage            string  `json:"current_stage"`
	// Dates from proposal_indexing
	DeclarationDate         time.Time `json:"declaration_date,omitempty"`
	ReceiptDate             time.Time `json:"receipt_date,omitempty"`
	IndexingDate            time.Time `json:"indexing_date,omitempty"`
	// Dates from proposal_issuance
	PolicyIssueDate         time.Time `json:"policy_issue_date,omitempty"`
	AcceptanceDate          time.Time `json:"acceptance_date,omitempty"`
	PolicyCommencementDate  time.Time `json:"policy_commencement_date,omitempty"`
	DispatchDate            time.Time `json:"dispatch_date,omitempty"`
	DeliveryDate            time.Time `json:"delivery_date,omitempty"`
	FLCStartDate            time.Time `json:"flc_start_date,omitempty"`
	FLCEndDate              time.Time `json:"flc_end_date,omitempty"`
	// Payment information
	FirstPremiumPaid        bool      `json:"first_premium_paid"`
	FirstPremiumDate        time.Time `json:"first_premium_date,omitempty"`
	FirstPremiumReference   string    `json:"first_premium_reference,omitempty"`
	FirstPremiumReceiptNumber string  `json:"first_premium_receipt_number,omitempty"`
	PremiumPaymentMethod    string    `json:"premium_payment_method,omitempty"`
	// Location information
	POCode                  string    `json:"po_code,omitempty"`
	IssueCircle             string    `json:"issue_circle,omitempty"`
	IssueHO                 string    `json:"issue_ho,omitempty"`
	IssuePostOffice         string    `json:"issue_post_office,omitempty"`
}

// PMCreatedSignal is the signal payload sent to PM's lifecycle workflow.
// This must match the PolicyLifecycleState struct expected by PolicyLifecycleWorkflow.
type PMCreatedSignal struct {
	PolicyNumber                   string               `json:"policy_number"`
	PolicyID                       string               `json:"policy_id"`    // UUID from Policy Issue (audit cross-ref)
	PolicyDBID                     int64                `json:"policy_db_id"` // BIGINT from PM seq_policy_id [A13]
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
	FLCExpiryAt  time.Time `json:"flc_expiry_at,omitempty"`
	ProductCode  string    `json:"product_code"`
	MaturityDate time.Time `json:"maturity_date"`
}

// PolicyMetadata holds policy-level data needed by the workflow for state gate
// decisions, eligibility checks, and activity calls. [§9.1]
type PolicyMetadata struct {
	CustomerID                  int64      `json:"customer_id"`
	ProductCode                 string     `json:"product_code"`
	ProductType                 string     `json:"product_type"` // PLI or RPLI
	SumAssured                  float64    `json:"sum_assured"`
	CurrentPremium              float64    `json:"current_premium"`
	PremiumMode                 string     `json:"premium_mode"`   // MONTHLY, QUARTERLY, HALF_YEARLY, YEARLY
	BillingMethod               string     `json:"billing_method"` // CASH or PAY_RECOVERY [BR-PM-074]
	IssueDate                   time.Time  `json:"issue_date"`
	MaturityDate                time.Time  `json:"maturity_date"`
	PaidToDate                  time.Time  `json:"paid_to_date"`
	AgentID                     *int64     `json:"agent_id,omitempty"` // Nullable BIGINT [Review-Fix-5]
	LoanOutstanding             float64    `json:"loan_outstanding"`
	AssignmentStatus            string     `json:"assignment_status"`
	PremiumsPaidMonths          int        `json:"premiums_paid_months"` // For paid-up calc [BR-PM-061]
	TotalPremiumsMonths         int        `json:"total_premiums_months"`
	RemissionExpiryDate         *time.Time `json:"remission_expiry_date,omitempty"`          // Nullable [Review-Fix-5]
	PayRecoveryProtectionExpiry *time.Time `json:"pay_recovery_protection_expiry,omitempty"` // Nullable, first_unpaid + 12mo [BR-PM-074, Review-Fix-5]
	SBInstallmentsPaid          int        `json:"sb_installments_paid"`
	NominationStatus            string     `json:"nomination_status"`
	IsDistanceMarketing         bool       `json:"is_distance_marketing"` // 30d FLC for distance-marketing products [Review-Fix-9]
	WorkflowID                  string     `json:"workflow_id"`           // plw-{policy_number}
}

// EncumbranceFlags groups all encumbrance state; passed to isStateEligible(). [§9.1]
type EncumbranceFlags struct {
	HasActiveLoan  bool   `json:"has_active_loan"` // Blocks new LOAN requests
	AssignmentType string `json:"assignment_type"` // NONE, ABSOLUTE, CONDITIONAL
	AMLHold        bool   `json:"aml_hold"`        // true when SUSPENDED [BR-PM-110]
	DisputeFlag    bool   `json:"dispute_flag"`    // Advisory only; never blocks [BR-PM-113, ADR-003]
}

// PendingRequest tracks a routed in-flight request waiting for a completion signal. [§9.1]
type PendingRequest struct {
	RequestID          string     `json:"request_id"`         // Dedup key (BIGINT as string or UUID)
	ServiceRequestID   int64      `json:"service_request_id"` // BIGINT from service_request table
	RequestType        string     `json:"request_type"`
	RequestCategory    string     `json:"request_category"`    // FINANCIAL or NON_FINANCIAL
	DownstreamWorkflow string     `json:"downstream_workflow"` // Child workflow ID
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

// StartPMLifecycleActivity sends a SignalWithStart to the PM service for the given policy.
//
// Behaviour:
//   - Increments pm_signal_attempts in proposal_issuance before the call (best-effort).
//   - On success: writes pm_signal_status = 'SENT' and pm_plw_workflow_id.
//   - On failure: writes pm_signal_status = 'FAILED' and pm_signal_last_error,
//     then returns the error so Temporal will retry this activity according to
//     the caller's RetryPolicy.
//
// Idempotency: SignalWithStart is idempotent by workflowID — PM will ignore a
// duplicate start if the workflow is already running.
func (a *PMLifecycleActivities) StartPMLifecycleActivity(ctx context.Context, input StartPMLifecycleInput) error {
	if input.PolicyNumber == "" {
		return temporal.NewNonRetryableApplicationError("policy_number is required", "INVALID_INPUT", nil)
	}

	workflowID := fmt.Sprintf("plw-%s", input.PolicyNumber)

	// Count the attempt before we call PM — best-effort, does not change status.
	_ = a.store.IncrementPMSignalAttempts(ctx, input.PolicyNumber)

	// For a newly issued policy, we need to set initial state
	now := time.Now().UTC()
	
	// Parse customer ID from string to int64
	var customerID int64
	if input.CustomerID != "" {
		_, err := fmt.Sscanf(input.CustomerID, "%d", &customerID)
		if err != nil {
			// If parsing fails, use 0 as default
			customerID = 0
		}
	}
	
	signal := PMCreatedSignal{
		PolicyNumber:                   input.PolicyNumber,
		PolicyID:                       "", // Will be generated by PM service
		PolicyDBID:                     0,  // Will be assigned by PM service
		CurrentStatus:                  "ACTIVE",
		PreviousStatus:                 "",
		PreviousStatusBeforeSuspension: "",
		Encumbrances: EncumbranceFlags{
			HasActiveLoan:  false,
			AssignmentType: "NONE",
			AMLHold:        false,
			DisputeFlag:    false,
		},
		DisplayStatus:      "ACTIVE",
		Version:            1,
		PendingRequests:    []PendingRequest{},
		ActiveLock:         nil,
		ProcessedSignalIDs: make(map[string]time.Time),
		EventCount:         0,
		LastCANTime:        time.Time{},
		LastTransitionAt:   now,
		CachedConfig:       make(map[string]string),
		FLCExpiryAt:        input.FLCEndDate,
		ProductCode:        input.ProductCode,
		MaturityDate:       calculateMaturityDate(input.PolicyCommencementDate, input.PolicyTerm),
		Metadata: PolicyMetadata{
			CustomerID:                  customerID,
			ProductCode:                 input.ProductCode,
			ProductType:                 input.ProductType,
			SumAssured:                  input.SumAssured,
			CurrentPremium:              input.ModalPremium,
			PremiumMode:                 convertPremiumFrequencyToMode(input.PremiumPaymentFrequency),
			BillingMethod:               "CASH", // Default for new policies
			IssueDate:                   input.PolicyIssueDate,
			MaturityDate:                calculateMaturityDate(input.PolicyCommencementDate, input.PolicyTerm),
			PaidToDate:                  input.FirstPremiumDate,
			AgentID:                     nil, // Will be populated from proposal data if available
			LoanOutstanding:             0,
			AssignmentStatus:            "NONE",
			PremiumsPaidMonths:          0,
			TotalPremiumsMonths:         input.PolicyTerm * 12,
			RemissionExpiryDate:         nil,
			PayRecoveryProtectionExpiry: nil,
			SBInstallmentsPaid:          0,
			NominationStatus:            "PENDING",
			IsDistanceMarketing:         input.Channel == "DISTANCE_MARKETING",
			WorkflowID:                  fmt.Sprintf("plw-%s", input.PolicyNumber),
		},
	}

	startOpts := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: a.pmTaskQueue,
	}

	_, err := a.signaller.SignalWithStartWorkflow(
		ctx,
		workflowID,
		"policy-created", // signal name PM's lifecycle workflow listens on
		signal,
		startOpts,
		"PolicyLifecycleWorkflow", // PM's workflow function name (registered on PM worker)
		signal,                    // initial input if the workflow isn't running yet
	)
	if err != nil {
		// Persist failure so the reconciliation worker can find it.
		_ = a.store.MarkPMSignalFailed(ctx, input.PolicyNumber, err.Error())
		return fmt.Errorf("SignalWithStart for %s failed: %w", workflowID, err)
	}

	// Persist success — reconciliation worker will skip this row.
	if dbErr := a.store.MarkPMSignalSent(ctx, input.PolicyNumber, workflowID); dbErr != nil {
		// The signal reached PM; a DB write failure here is non-critical.
		// The reconciliation worker will see status still PENDING/FAILED and
		// re-attempt SignalWithStart, which is idempotent.
		return fmt.Errorf("signal sent but failed to persist SENT status for %s: %w", input.PolicyNumber, dbErr)
	}

	return nil
}

// calculateMaturityDate calculates the maturity date based on policy commencement date and term in years
func calculateMaturityDate(commencementDate time.Time, policyTermYears int) time.Time {
	if commencementDate.IsZero() {
		return time.Time{}
	}
	return commencementDate.AddDate(policyTermYears, 0, 0)
}

// convertPremiumFrequencyToMode converts premium payment frequency string to premium mode
func convertPremiumFrequencyToMode(frequency string) string {
	switch frequency {
	case "MONTHLY":
		return "MONTHLY"
	case "QUARTERLY":
		return "QUARTERLY"
	case "HALF_YEARLY":
		return "HALF_YEARLY"
	case "YEARLY":
		return "YEARLY"
	default:
		return "YEARLY" // Default fallback
	}
}

// FindUnsignalledPoliciesInput is the input for FindUnsignalledPoliciesActivity.
type FindUnsignalledPoliciesInput struct {
	// GracePeriodMinutes is how old a PENDING record must be before reconciliation
	// considers it stuck (prevents interfering with in-progress workflows).
	GracePeriodMinutes int `json:"grace_period_minutes"`
	// MaxAttempts caps how many times a policy will be retried before giving up.
	MaxAttempts int `json:"max_attempts"`
}

// FindUnsignalledPoliciesActivity queries the PI DB for policies that need
// PM signal retry and returns them as StartPMLifecycleInput slices.
func (a *PMLifecycleActivities) FindUnsignalledPoliciesActivity(
	ctx context.Context,
	input FindUnsignalledPoliciesInput,
) ([]StartPMLifecycleInput, error) {
	gracePeriod := time.Duration(input.GracePeriodMinutes) * time.Minute
	maxAttempts := input.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = 20
	}
	if gracePeriod <= 0 {
		gracePeriod = 30 * time.Minute
	}

	targets, err := a.store.FindUnsignalledPolicies(ctx, gracePeriod, maxAttempts)
	if err != nil {
		return nil, fmt.Errorf("FindUnsignalledPoliciesActivity: %w", err)
	}

	result := make([]StartPMLifecycleInput, 0, len(targets))
	for _, t := range targets {
		result = append(result, StartPMLifecycleInput{
			PolicyNumber: t.PolicyNumber,
			PolicyType:   t.PolicyType,
			ProposalID:   t.ProposalID,
			ProposalDate: t.PolicyIssueDate,
		})
	}
	return result, nil
}
