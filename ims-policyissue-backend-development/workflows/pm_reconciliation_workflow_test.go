package workflows_test

import (
	"errors"
	"testing"

	"policy-issue-service/workflows"
	"policy-issue-service/workflows/activities"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.temporal.io/sdk/testsuite"
)

// ─────────────────────────────────────────────
// Test suite setup
// ─────────────────────────────────────────────

type PMReconciliationWorkflowTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
	env *testsuite.TestWorkflowEnvironment
}

func (s *PMReconciliationWorkflowTestSuite) SetupTest() {
	s.env = s.NewTestWorkflowEnvironment()
}

func (s *PMReconciliationWorkflowTestSuite) AfterTest(_, _ string) {
	s.env.AssertExpectations(s.T())
}

func TestPMReconciliationWorkflowSuite(t *testing.T) {
	suite.Run(t, new(PMReconciliationWorkflowTestSuite))
}

// ─────────────────────────────────────────────
// Test cases
// ─────────────────────────────────────────────

// TestNothingToReconcile verifies that when FindUnsignalledPoliciesActivity returns
// an empty list the workflow completes successfully without calling StartPMLifecycleActivity.
func (s *PMReconciliationWorkflowTestSuite) TestNothingToReconcile() {
	emptyTargets := []activities.StartPMLifecycleInput{}

	s.env.OnActivity("FindUnsignalledPoliciesActivity", mock.Anything, mock.Anything).
		Return(emptyTargets, nil).Once()

	// StartPMLifecycleActivity must NOT be called
	s.env.OnActivity("StartPMLifecycleActivity", mock.Anything, mock.Anything).
		Return(nil).Times(0)

	s.env.ExecuteWorkflow(workflows.PMSignalReconciliationWorkflow)

	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
}

// TestAllSignalsSucceed verifies that when two policies are returned, both
// StartPMLifecycleActivity calls succeed and the workflow completes without error.
func (s *PMReconciliationWorkflowTestSuite) TestAllSignalsSucceed() {
	targets := []activities.StartPMLifecycleInput{
		{PolicyNumber: "PLI/2026/GJ/000001", PolicyType: "PLI"},
		{PolicyNumber: "RPLI/2026/MH/000002", PolicyType: "RPLI"},
	}

	s.env.OnActivity("FindUnsignalledPoliciesActivity", mock.Anything, mock.Anything).
		Return(targets, nil).Once()

	s.env.OnActivity("StartPMLifecycleActivity", mock.Anything,
		activities.StartPMLifecycleInput{PolicyNumber: "PLI/2026/GJ/000001", PolicyType: "PLI"}).
		Return(nil).Once()

	s.env.OnActivity("StartPMLifecycleActivity", mock.Anything,
		activities.StartPMLifecycleInput{PolicyNumber: "RPLI/2026/MH/000002", PolicyType: "RPLI"}).
		Return(nil).Once()

	s.env.ExecuteWorkflow(workflows.PMSignalReconciliationWorkflow)

	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
}

// TestPartialFailure verifies that when some signals fail the workflow still
// completes without error — each policy is independent and failures are logged.
func (s *PMReconciliationWorkflowTestSuite) TestPartialFailure() {
	targets := []activities.StartPMLifecycleInput{
		{PolicyNumber: "PLI/2026/GJ/000010", PolicyType: "PLI"},
		{PolicyNumber: "PLI/2026/GJ/000011", PolicyType: "PLI"},
		{PolicyNumber: "PLI/2026/GJ/000012", PolicyType: "PLI"},
	}

	s.env.OnActivity("FindUnsignalledPoliciesActivity", mock.Anything, mock.Anything).
		Return(targets, nil).Once()

	// First policy succeeds
	s.env.OnActivity("StartPMLifecycleActivity", mock.Anything,
		activities.StartPMLifecycleInput{PolicyNumber: "PLI/2026/GJ/000010", PolicyType: "PLI"}).
		Return(nil).Once()

	// Second policy fails
	s.env.OnActivity("StartPMLifecycleActivity", mock.Anything,
		activities.StartPMLifecycleInput{PolicyNumber: "PLI/2026/GJ/000011", PolicyType: "PLI"}).
		Return(errors.New("PM down")).Once()

	// Third policy succeeds
	s.env.OnActivity("StartPMLifecycleActivity", mock.Anything,
		activities.StartPMLifecycleInput{PolicyNumber: "PLI/2026/GJ/000012", PolicyType: "PLI"}).
		Return(nil).Once()

	s.env.ExecuteWorkflow(workflows.PMSignalReconciliationWorkflow)

	s.True(s.env.IsWorkflowCompleted())
	// Workflow must NOT fail even though one signal failed
	s.NoError(s.env.GetWorkflowError())
}

// TestAllSignalsFail verifies that even when every signal fails, the workflow
// itself completes without returning an error (failures are per-policy).
func (s *PMReconciliationWorkflowTestSuite) TestAllSignalsFail() {
	targets := []activities.StartPMLifecycleInput{
		{PolicyNumber: "PLI/2026/GJ/000020", PolicyType: "PLI"},
		{PolicyNumber: "PLI/2026/GJ/000021", PolicyType: "PLI"},
	}

	s.env.OnActivity("FindUnsignalledPoliciesActivity", mock.Anything, mock.Anything).
		Return(targets, nil).Once()

	s.env.OnActivity("StartPMLifecycleActivity", mock.Anything, mock.Anything).
		Return(errors.New("PM unreachable")).Times(2)

	s.env.ExecuteWorkflow(workflows.PMSignalReconciliationWorkflow)

	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
}

// TestFindActivityFails verifies that when FindUnsignalledPoliciesActivity itself
// fails the workflow returns an error.
func (s *PMReconciliationWorkflowTestSuite) TestFindActivityFails() {
	s.env.OnActivity("FindUnsignalledPoliciesActivity", mock.Anything, mock.Anything).
		Return([]activities.StartPMLifecycleInput(nil), errors.New("db unavailable")).Once()

	s.env.ExecuteWorkflow(workflows.PMSignalReconciliationWorkflow)

	s.True(s.env.IsWorkflowCompleted())
	err := s.env.GetWorkflowError()
	s.Error(err)
	s.Contains(err.Error(), "db unavailable")
}

// TestFindActivityPassesCorrectDefaults verifies that the workflow passes the
// constants pmGracePeriodMinutes=30 and pmMaxAttempts=20 to the query activity.
func (s *PMReconciliationWorkflowTestSuite) TestFindActivityPassesCorrectDefaults() {
	expectedInput := activities.FindUnsignalledPoliciesInput{
		GracePeriodMinutes: 30,
		MaxAttempts:        20,
	}

	s.env.OnActivity("FindUnsignalledPoliciesActivity", mock.Anything, expectedInput).
		Return([]activities.StartPMLifecycleInput{}, nil).Once()

	s.env.ExecuteWorkflow(workflows.PMSignalReconciliationWorkflow)

	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
}
