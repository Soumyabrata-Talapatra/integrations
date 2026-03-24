package activities_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"policy-issue-service/repo/postgres"
	"policy-issue-service/workflows/activities"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/testsuite"
)

// ─────────────────────────────────────────────
// Manual mocks
// ─────────────────────────────────────────────

// mockPMSignalStore is a hand-written mock for activities.PMSignalStore.
type mockPMSignalStore struct {
	incrementFn  func(ctx context.Context, policyNumber string) error
	markSentFn   func(ctx context.Context, policyNumber, workflowID string) error
	markFailedFn func(ctx context.Context, policyNumber, errMsg string) error
	findFn       func(ctx context.Context, gracePeriod time.Duration, maxAttempts int) ([]postgres.PMSignalTarget, error)
}

func (m *mockPMSignalStore) IncrementPMSignalAttempts(ctx context.Context, policyNumber string) error {
	if m.incrementFn != nil {
		return m.incrementFn(ctx, policyNumber)
	}
	return nil
}

func (m *mockPMSignalStore) MarkPMSignalSent(ctx context.Context, policyNumber, workflowID string) error {
	if m.markSentFn != nil {
		return m.markSentFn(ctx, policyNumber, workflowID)
	}
	return nil
}

func (m *mockPMSignalStore) MarkPMSignalFailed(ctx context.Context, policyNumber, errMsg string) error {
	if m.markFailedFn != nil {
		return m.markFailedFn(ctx, policyNumber, errMsg)
	}
	return nil
}

func (m *mockPMSignalStore) FindUnsignalledPolicies(ctx context.Context, gracePeriod time.Duration, maxAttempts int) ([]postgres.PMSignalTarget, error) {
	if m.findFn != nil {
		return m.findFn(ctx, gracePeriod, maxAttempts)
	}
	return nil, nil
}

// mockWorkflowRun satisfies client.WorkflowRun (returned by SignalWithStartWorkflow).
type mockWorkflowRun struct{}

func (m *mockWorkflowRun) GetID() string                                        { return "test-workflow-id" }
func (m *mockWorkflowRun) GetRunID() string                                     { return "test-run-id" }
func (m *mockWorkflowRun) Get(ctx context.Context, valuePtr interface{}) error  { return nil }
func (m *mockWorkflowRun) GetWithOptions(ctx context.Context, valuePtr interface{}, opts client.WorkflowRunGetOptions) error {
	return nil
}

// mockTemporalSignaller is a hand-written mock for activities.TemporalSignaller.
type mockTemporalSignaller struct {
	signalFn func(ctx context.Context, workflowID, signalName string, signalArg interface{},
		options client.StartWorkflowOptions, workflow interface{}, workflowArgs ...interface{}) (client.WorkflowRun, error)
}

func (m *mockTemporalSignaller) SignalWithStartWorkflow(
	ctx context.Context,
	workflowID string,
	signalName string,
	signalArg interface{},
	options client.StartWorkflowOptions,
	workflow interface{},
	workflowArgs ...interface{},
) (client.WorkflowRun, error) {
	if m.signalFn != nil {
		return m.signalFn(ctx, workflowID, signalName, signalArg, options, workflow, workflowArgs...)
	}
	return &mockWorkflowRun{}, nil
}

// ─────────────────────────────────────────────
// Helper: build activity with mocks
// ─────────────────────────────────────────────

func newTestActivity(store *mockPMSignalStore, sig *mockTemporalSignaller) *activities.PMLifecycleActivities {
	return activities.NewPMLifecycleActivities(store, sig, "test-pm-queue")
}

// ─────────────────────────────────────────────
// StartPMLifecycleActivity tests
// ─────────────────────────────────────────────

// TestStartPMLifecycleActivity_EmptyPolicyNumber checks that a missing policy
// number returns a NonRetryableApplicationError immediately.
func TestStartPMLifecycleActivity_EmptyPolicyNumber(t *testing.T) {
	var ts testsuite.WorkflowTestSuite
	env := ts.NewTestActivityEnvironment()

	act := newTestActivity(&mockPMSignalStore{}, &mockTemporalSignaller{})
	env.RegisterActivity(act)

	_, err := env.ExecuteActivity(act.StartPMLifecycleActivity, activities.StartPMLifecycleInput{
		PolicyNumber: "",
		PolicyType:   "PLI",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "policy_number is required")
}

// TestStartPMLifecycleActivity_Success verifies the happy path:
// SignalWithStart succeeds → MarkPMSignalSent is called with the correct workflow ID.
func TestStartPMLifecycleActivity_Success(t *testing.T) {
	var ts testsuite.WorkflowTestSuite
	env := ts.NewTestActivityEnvironment()

	var sentWorkflowID string
	store := &mockPMSignalStore{
		markSentFn: func(_ context.Context, policyNumber, wfID string) error {
			sentWorkflowID = wfID
			return nil
		},
	}
	sig := &mockTemporalSignaller{
		signalFn: func(_ context.Context, wfID, _ string, _ interface{},
			_ client.StartWorkflowOptions, _ interface{}, _ ...interface{},
		) (client.WorkflowRun, error) {
			return &mockWorkflowRun{}, nil
		},
	}

	act := newTestActivity(store, sig)
	env.RegisterActivity(act)

	_, err := env.ExecuteActivity(act.StartPMLifecycleActivity, activities.StartPMLifecycleInput{
		PolicyNumber: "PLI/2026/GJ/000001",
		PolicyType:   "PLI",
	})
	require.NoError(t, err)
	assert.Equal(t, "plw-PLI/2026/GJ/000001", sentWorkflowID)
}

// TestStartPMLifecycleActivity_SignalFails verifies that when SignalWithStartWorkflow
// returns an error: MarkPMSignalFailed is called and the activity returns an error
// (so Temporal will retry).
func TestStartPMLifecycleActivity_SignalFails(t *testing.T) {
	var ts testsuite.WorkflowTestSuite
	env := ts.NewTestActivityEnvironment()

	signalErr := errors.New("PM service unavailable")
	var failedPolicyNumber, failedMsg string

	store := &mockPMSignalStore{
		markFailedFn: func(_ context.Context, policyNumber, errMsg string) error {
			failedPolicyNumber = policyNumber
			failedMsg = errMsg
			return nil
		},
	}
	sig := &mockTemporalSignaller{
		signalFn: func(_ context.Context, _ string, _ string, _ interface{},
			_ client.StartWorkflowOptions, _ interface{}, _ ...interface{},
		) (client.WorkflowRun, error) {
			return nil, signalErr
		},
	}

	act := newTestActivity(store, sig)
	env.RegisterActivity(act)

	_, err := env.ExecuteActivity(act.StartPMLifecycleActivity, activities.StartPMLifecycleInput{
		PolicyNumber: "PLI/2026/GJ/000002",
		PolicyType:   "PLI",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "PLI/2026/GJ/000002")
	assert.Equal(t, "PLI/2026/GJ/000002", failedPolicyNumber)
	assert.Contains(t, failedMsg, "PM service unavailable")
}

// TestStartPMLifecycleActivity_SignalSucceedsButDBWriteFails verifies that when
// SignalWithStart succeeds but MarkPMSignalSent fails, the activity returns an
// error (so the reconciliation worker will retry — it's idempotent at PM's side).
func TestStartPMLifecycleActivity_SignalSucceedsButDBWriteFails(t *testing.T) {
	var ts testsuite.WorkflowTestSuite
	env := ts.NewTestActivityEnvironment()

	dbErr := errors.New("db connection lost")
	store := &mockPMSignalStore{
		markSentFn: func(_ context.Context, _, _ string) error {
			return dbErr
		},
	}
	sig := &mockTemporalSignaller{} // always succeeds

	act := newTestActivity(store, sig)
	env.RegisterActivity(act)

	_, err := env.ExecuteActivity(act.StartPMLifecycleActivity, activities.StartPMLifecycleInput{
		PolicyNumber: "PLI/2026/GJ/000003",
		PolicyType:   "PLI",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "signal sent but failed to persist SENT status")
}

// TestStartPMLifecycleActivity_IncrementCalledBeforeSignal verifies that
// IncrementPMSignalAttempts is called on every attempt, regardless of outcome.
func TestStartPMLifecycleActivity_IncrementCalledBeforeSignal(t *testing.T) {
	var ts testsuite.WorkflowTestSuite
	env := ts.NewTestActivityEnvironment()

	incrementCalled := false
	store := &mockPMSignalStore{
		incrementFn: func(_ context.Context, _ string) error {
			incrementCalled = true
			return nil
		},
	}
	sig := &mockTemporalSignaller{}

	act := newTestActivity(store, sig)
	env.RegisterActivity(act)

	_, _ = env.ExecuteActivity(act.StartPMLifecycleActivity, activities.StartPMLifecycleInput{
		PolicyNumber: "PLI/2026/GJ/000004",
		PolicyType:   "PLI",
	})
	assert.True(t, incrementCalled, "IncrementPMSignalAttempts must be called on every attempt")
}

// ─────────────────────────────────────────────
// FindUnsignalledPoliciesActivity tests
// ─────────────────────────────────────────────

// TestFindUnsignalledPoliciesActivity_EmptyResult verifies that an empty DB result
// returns an empty slice without error.
func TestFindUnsignalledPoliciesActivity_EmptyResult(t *testing.T) {
	var ts testsuite.WorkflowTestSuite
	env := ts.NewTestActivityEnvironment()

	store := &mockPMSignalStore{
		findFn: func(_ context.Context, _ time.Duration, _ int) ([]postgres.PMSignalTarget, error) {
			return nil, nil
		},
	}
	act := newTestActivity(store, &mockTemporalSignaller{})
	env.RegisterActivity(act)

	val, err := env.ExecuteActivity(act.FindUnsignalledPoliciesActivity, activities.FindUnsignalledPoliciesInput{
		GracePeriodMinutes: 30,
		MaxAttempts:        20,
	})
	require.NoError(t, err)

	var result []activities.StartPMLifecycleInput
	require.NoError(t, val.Get(&result))
	assert.Empty(t, result)
}

// TestFindUnsignalledPoliciesActivity_MapsTwoTargets verifies that two DB rows are
// correctly mapped to two StartPMLifecycleInput values.
func TestFindUnsignalledPoliciesActivity_MapsTwoTargets(t *testing.T) {
	var ts testsuite.WorkflowTestSuite
	env := ts.NewTestActivityEnvironment()

	dbRows := []postgres.PMSignalTarget{
		{PolicyNumber: "PLI/2026/GJ/000010", PolicyType: "PLI", Attempts: 1},
		{PolicyNumber: "RPLI/2026/MH/000011", PolicyType: "RPLI", Attempts: 3},
	}
	store := &mockPMSignalStore{
		findFn: func(_ context.Context, _ time.Duration, _ int) ([]postgres.PMSignalTarget, error) {
			return dbRows, nil
		},
	}
	act := newTestActivity(store, &mockTemporalSignaller{})
	env.RegisterActivity(act)

	val, err := env.ExecuteActivity(act.FindUnsignalledPoliciesActivity, activities.FindUnsignalledPoliciesInput{
		GracePeriodMinutes: 30,
		MaxAttempts:        20,
	})
	require.NoError(t, err)

	var result []activities.StartPMLifecycleInput
	require.NoError(t, val.Get(&result))

	require.Len(t, result, 2)
	assert.Equal(t, "PLI/2026/GJ/000010", result[0].PolicyNumber)
	assert.Equal(t, "PLI", result[0].PolicyType)
	assert.Equal(t, "RPLI/2026/MH/000011", result[1].PolicyNumber)
	assert.Equal(t, "RPLI", result[1].PolicyType)
}

// TestFindUnsignalledPoliciesActivity_DBError verifies that a DB error is
// propagated as an activity error.
func TestFindUnsignalledPoliciesActivity_DBError(t *testing.T) {
	var ts testsuite.WorkflowTestSuite
	env := ts.NewTestActivityEnvironment()

	store := &mockPMSignalStore{
		findFn: func(_ context.Context, _ time.Duration, _ int) ([]postgres.PMSignalTarget, error) {
			return nil, errors.New("connection refused")
		},
	}
	act := newTestActivity(store, &mockTemporalSignaller{})
	env.RegisterActivity(act)

	_, err := env.ExecuteActivity(act.FindUnsignalledPoliciesActivity, activities.FindUnsignalledPoliciesInput{
		GracePeriodMinutes: 30,
		MaxAttempts:        20,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "connection refused")
}

// TestFindUnsignalledPoliciesActivity_DefaultsApplied verifies that zero-value
// GracePeriodMinutes and MaxAttempts fall back to 30 min / 20 attempts.
func TestFindUnsignalledPoliciesActivity_DefaultsApplied(t *testing.T) {
	var ts testsuite.WorkflowTestSuite
	env := ts.NewTestActivityEnvironment()

	var capturedGrace time.Duration
	var capturedMax int
	store := &mockPMSignalStore{
		findFn: func(_ context.Context, gracePeriod time.Duration, maxAttempts int) ([]postgres.PMSignalTarget, error) {
			capturedGrace = gracePeriod
			capturedMax = maxAttempts
			return nil, nil
		},
	}
	act := newTestActivity(store, &mockTemporalSignaller{})
	env.RegisterActivity(act)

	_, err := env.ExecuteActivity(act.FindUnsignalledPoliciesActivity, activities.FindUnsignalledPoliciesInput{
		GracePeriodMinutes: 0, // should default to 30
		MaxAttempts:        0, // should default to 20
	})
	require.NoError(t, err)
	assert.Equal(t, 30*time.Minute, capturedGrace)
	assert.Equal(t, 20, capturedMax)
}
