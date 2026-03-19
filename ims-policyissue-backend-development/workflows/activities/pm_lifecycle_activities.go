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
}

// PMCreatedSignal is the signal payload sent to PM's lifecycle workflow.
type PMCreatedSignal struct {
	PolicyNumber string    `json:"policy_number"`
	PolicyType   string    `json:"policy_type"`
	IssuedAt     time.Time `json:"issued_at"`
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

	signal := PMCreatedSignal{
		PolicyNumber: input.PolicyNumber,
		PolicyType:   input.PolicyType,
		IssuedAt:     time.Now().UTC(),
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
		})
	}
	return result, nil
}
