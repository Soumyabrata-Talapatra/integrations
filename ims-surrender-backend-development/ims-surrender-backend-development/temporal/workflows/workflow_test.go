package workflows

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"

	"gitlab.cept.gov.in/it-2.0-policy/surrender-service/temporal/activities"
)

// TestVoluntarySurrenderWorkflow tests the voluntary surrender workflow
func TestVoluntarySurrenderWorkflow(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	// Mock activities
	env.OnActivity(activities.ValidateEligibilityActivity, mock.Anything, mock.Anything).
		Return(&activities.ValidateEligibilityResult{
			Eligible: true,
			Reasons:  []string{},
		}, nil)

	env.OnActivity(activities.CalculateSurrenderValueActivity, mock.Anything, mock.Anything).
		Return(&activities.CalculateSurrenderValueResult{
			GrossSurrenderValue:  50000,
			NetSurrenderValue:    45000,
			PredictedDisposition: "TS",
		}, nil)

	env.OnActivity(activities.VerifyDocumentsActivity, mock.Anything, mock.Anything).
		Return(&activities.VerifyDocumentsResult{
			AllVerified:   true,
			VerifiedCount: 3,
			RequiredCount: 3,
		}, nil)

	env.OnActivity(activities.RouteToApprovalActivity, mock.Anything, mock.Anything).
		Return(&activities.RouteToApprovalResult{
			TaskID:   "task-123",
			Assigned: true,
		}, nil)

	env.OnActivity(activities.ProcessPaymentActivity, mock.Anything, mock.Anything).
		Return(&activities.ProcessPaymentResult{
			PaymentReference: "PAY-123",
			Success:          true,
		}, nil)

	env.OnActivity(activities.UpdatePolicyStatusActivity, mock.Anything, mock.Anything).
		Return(&activities.UpdatePolicyStatusResult{
			NewStatus: "TS",
			Success:   true,
		}, nil)

	// Register signals
	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow("documents-uploaded", true)
	}, time.Second*1)

	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow("approval-decision", "APPROVED")
	}, time.Second*2)

	// Execute workflow
	env.ExecuteWorkflow(VoluntarySurrenderWorkflow, VoluntarySurrenderWorkflowInput{
		SurrenderRequestID: "req-123",
		PolicyID:           "policy-123",
		RequestNumber:      "SUR-123",
		RequestedBy:        "user-123",
	})

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())

	var result VoluntarySurrenderWorkflowResult
	require.NoError(t, env.GetWorkflowResult(&result))

	assert.Equal(t, "COMPLETED", result.Status)
	assert.Equal(t, "PAY-123", result.PaymentReference)
	assert.Equal(t, "TS", result.PolicyStatus)
}

// TestVoluntarySurrenderWorkflow_IneligiblePolicy tests workflow with ineligible policy
func TestVoluntarySurrenderWorkflow_IneligiblePolicy(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	// Mock eligibility check to return false
	env.OnActivity(activities.ValidateEligibilityActivity, mock.Anything, mock.Anything).
		Return(&activities.ValidateEligibilityResult{
			Eligible: false,
			Reasons:  []string{"Policy status not eligible", "Insufficient premiums paid"},
		}, nil)

	// Execute workflow
	env.ExecuteWorkflow(VoluntarySurrenderWorkflow, VoluntarySurrenderWorkflowInput{
		SurrenderRequestID: "req-123",
		PolicyID:           "policy-123",
		RequestNumber:      "SUR-123",
		RequestedBy:        "user-123",
	})

	require.True(t, env.IsWorkflowCompleted())
	require.Error(t, env.GetWorkflowError())

	var result VoluntarySurrenderWorkflowResult
	env.GetWorkflowResult(&result)

	assert.Equal(t, "NOT_ELIGIBLE", result.Status)
}

// TestVoluntarySurrenderWorkflow_DocumentTimeout tests document upload timeout
func TestVoluntarySurrenderWorkflow_DocumentTimeout(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	// Mock activities
	env.OnActivity(activities.ValidateEligibilityActivity, mock.Anything, mock.Anything).
		Return(&activities.ValidateEligibilityResult{Eligible: true}, nil)

	env.OnActivity(activities.CalculateSurrenderValueActivity, mock.Anything, mock.Anything).
		Return(&activities.CalculateSurrenderValueResult{
			NetSurrenderValue: 45000,
		}, nil)

	// Don't signal document upload - let it timeout

	// Execute workflow
	env.ExecuteWorkflow(VoluntarySurrenderWorkflow, VoluntarySurrenderWorkflowInput{
		SurrenderRequestID: "req-123",
		PolicyID:           "policy-123",
	})

	require.True(t, env.IsWorkflowCompleted())

	var result VoluntarySurrenderWorkflowResult
	env.GetWorkflowResult(&result)

	assert.Equal(t, "TIMEOUT_DOCUMENTS", result.Status)
}

// TestForcedSurrenderWorkflow tests the forced surrender evaluation workflow
func TestForcedSurrenderWorkflow(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	// Mock identify eligible policies
	env.OnActivity(activities.IdentifyEligiblePoliciesActivity, mock.Anything, mock.Anything).
		Return(&activities.IdentifyEligiblePoliciesResult{
			EligiblePolicies: []activities.PolicyInfo{
				{PolicyID: "p1", PolicyNumber: "PLI/2020/001", UnpaidMonths: 6},
				{PolicyID: "p2", PolicyNumber: "PLI/2020/002", UnpaidMonths: 9},
			},
		}, nil)

	// Mock create reminders batch
	env.OnActivity(activities.CreateRemindersBatchActivity, mock.Anything, mock.Anything).
		Return(&activities.CreateRemindersBatchResult{
			RemindersCreated: 2,
			Errors:           []string{},
		}, nil)

	// Mock check expired windows
	env.OnActivity(activities.CheckExpiredPaymentWindowsActivity, mock.Anything, mock.Anything).
		Return(&activities.CheckExpiredPaymentWindowsResult{
			ExpiredWindows: []activities.PaymentWindowInfo{},
		}, nil)

	// Execute workflow
	env.ExecuteWorkflow(ForcedSurrenderWorkflow, ForcedSurrenderWorkflowInput{
		EvaluationDate: "2026-01-27",
		BatchSize:      50,
	})

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())

	var result ForcedSurrenderWorkflowResult
	require.NoError(t, env.GetWorkflowResult(&result))

	assert.Equal(t, 2, result.PoliciesEvaluated)
	assert.Equal(t, 2, result.RemindersCreated)
	assert.Equal(t, 0, result.SurrendersInitiated)
}

// TestApprovalWorkflow tests the approval processing workflow
func TestApprovalWorkflow(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	// Mock activities
	env.OnActivity(activities.GetSurrenderRequestDetailsActivity, mock.Anything, mock.Anything).
		Return(&activities.GetSurrenderRequestDetailsResult{
			NetSurrenderValue: 3000, // Below auto-approval limit
		}, nil)

	env.OnActivity(activities.AutoApproveActivity, mock.Anything, mock.Anything).
		Return(&activities.AutoApproveResult{Approved: true}, nil)

	// Execute workflow with auto-approval
	env.ExecuteWorkflow(ApprovalWorkflow, ApprovalWorkflowInput{
		SurrenderRequestID: "req-123",
		Priority:           "NORMAL",
		AutoApprovalLimit:  5000,
	})

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())

	var result ApprovalWorkflowResult
	require.NoError(t, env.GetWorkflowResult(&result))

	assert.Equal(t, "AUTO_APPROVED", result.Decision)
	assert.Equal(t, "SYSTEM", result.ApprovedBy)
}

// TestApprovalWorkflow_ManualApproval tests manual approval flow
func TestApprovalWorkflow_ManualApproval(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	// Mock activities
	env.OnActivity(activities.GetSurrenderRequestDetailsActivity, mock.Anything, mock.Anything).
		Return(&activities.GetSurrenderRequestDetailsResult{
			NetSurrenderValue: 10000, // Above auto-approval limit
		}, nil)

	env.OnActivity(activities.CreateApprovalTaskActivity, mock.Anything, mock.Anything).
		Return(&activities.CreateApprovalTaskResult{TaskID: "task-123"}, nil)

	env.OnActivity(activities.ProcessApprovalDecisionActivity, mock.Anything, mock.Anything).
		Return(&activities.ProcessApprovalDecisionResult{
			ApprovedBy: "user-123",
			Success:    true,
		}, nil)

	// Signal approval decision
	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow("approval-decision", "APPROVED")
	}, time.Second*1)

	// Execute workflow
	env.ExecuteWorkflow(ApprovalWorkflow, ApprovalWorkflowInput{
		SurrenderRequestID: "req-123",
		Priority:           "NORMAL",
		AutoApprovalLimit:  5000,
	})

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())

	var result ApprovalWorkflowResult
	require.NoError(t, env.GetWorkflowResult(&result))

	assert.Equal(t, "APPROVED", result.Decision)
	assert.Equal(t, "user-123", result.ApprovedBy)
}

// TestPaymentWorkflow tests payment processing workflow
func TestPaymentWorkflow(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	// Mock all activities
	env.OnActivity(activities.ValidatePaymentEligibilityActivity, mock.Anything, mock.Anything).
		Return(&activities.ValidatePaymentEligibilityResult{Eligible: true}, nil)

	env.OnActivity(activities.DetermineDispositionActivity, mock.Anything, mock.Anything).
		Return(&activities.DetermineDispositionResult{
			DispositionType: "TERMINATED_SURRENDER",
			NewPolicyStatus: "TS",
		}, nil)

	env.OnActivity(activities.ProcessPaymentActivity, mock.Anything, mock.Anything).
		Return(&activities.ProcessPaymentResult{
			PaymentReference: "PAY-123",
			Success:          true,
		}, nil)

	env.OnActivity(activities.CreateDispositionRecordActivity, mock.Anything, mock.Anything).
		Return(&activities.CreateDispositionRecordResult{DispositionID: "disp-123"}, nil)

	env.OnActivity(activities.UpdatePolicyStatusActivity, mock.Anything, mock.Anything).
		Return(&activities.UpdatePolicyStatusResult{NewStatus: "TS", Success: true}, nil)

	env.OnActivity(activities.SendPaymentNotificationActivity, mock.Anything, mock.Anything).
		Return(&activities.SendPaymentNotificationResult{ChannelsSent: []string{"EMAIL"}}, nil)

	// Execute workflow
	env.ExecuteWorkflow(PaymentWorkflow, PaymentWorkflowInput{
		SurrenderRequestID: "req-123",
		PolicyID:           "policy-123",
		Amount:             45000,
		DisbursementMethod: "CHEQUE",
	})

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())

	var result PaymentWorkflowResult
	require.NoError(t, env.GetWorkflowResult(&result))

	assert.Equal(t, "COMPLETED", result.Status)
	assert.Equal(t, "PAY-123", result.PaymentReference)
	assert.Equal(t, "TERMINATED_SURRENDER", result.DispositionType)
	assert.Equal(t, "TS", result.NewPolicyStatus)
}
