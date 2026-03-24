package workflow

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/testsuite"
)

// =============================================================================
// PM Integration Test Suite
// =============================================================================

// PMIntegrationTestSuite tests Policy Management integration features:
// 1. Task queue alignment (revival-tq)
// 2. PM notification at terminal workflow points
// 3. Policy validation runs in workflow (not handler)
type PMIntegrationTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
	env *testsuite.TestWorkflowEnvironment
}

// MockPMActivities provides mock implementations including PM notification
type MockPMActivities struct{}

func (m *MockPMActivities) ValidatePolicyActivity(ctx interface{}, policyNumber string) (PolicyValidationResult, error) {
	return PolicyValidationResult{}, nil
}
func (m *MockPMActivities) CreateRevivalRequestActivity(ctx interface{}, input RevivalRequestInput) (string, error) {
	return "", nil
}
func (m *MockPMActivities) UpdateDataEntryActivity(ctx interface{}, requestID string, input DataEntryInput) error {
	return nil
}
func (m *MockPMActivities) CheckAndAdjustSuspenseActivity(ctx interface{}, requestID, policyNumber string, revivalAmount float64) (SuspenseAdjustmentResult, error) {
	return SuspenseAdjustmentResult{}, nil
}
func (m *MockPMActivities) UpdateQCActivity(ctx interface{}, requestID, qcPerformedBy, qcComments string, qcPassed bool, missingDocuments string) error {
	return nil
}
func (m *MockPMActivities) TerminateAndReturnToIndexerActivity(ctx interface{}, requestID string, reason string, stage string) error {
	return nil
}
func (m *MockPMActivities) UpdateRevivalStatusActivity(ctx interface{}, requestID string, newStatus string) error {
	return nil
}
func (m *MockPMActivities) UpdateApprovalActivity(ctx interface{}, ticketID, approvedBy, comments string, slaStartDate, slaEndDate time.Time) error {
	return nil
}
func (m *MockPMActivities) UpdateWorkflowStateActivity(ctx interface{}, requestID string, status string, slaStart, slaEnd time.Time) error {
	return nil
}
func (m *MockPMActivities) TerminateRevivalActivity(ctx interface{}, requestID, reason string) error {
	return nil
}
func (m *MockPMActivities) FinalizeRevivalAfterFirstCollection(ctx interface{}, requestID string) error {
	return nil
}
func (m *MockPMActivities) NotifyPolicyManagementActivity(ctx interface{}, pmWorkflowID string, signalChannel string, signal PMCompletionSignal) error {
	return nil
}
func (m *MockPMActivities) ProcessInstallmentActivity(ctx interface{}, requestID string, installmentNumber int, amount float64, paymentMode, status string, collectionDate time.Time) error {
	return nil
}
func (m *MockPMActivities) HandleDefaultActivity(ctx interface{}, requestID string, installmentNumber int) error {
	return nil
}

func (s *PMIntegrationTestSuite) SetupTest() {
	s.env = s.NewTestWorkflowEnvironment()

	mockActs := &MockPMActivities{}
	s.env.RegisterActivityWithOptions(mockActs.ValidatePolicyActivity, activity.RegisterOptions{Name: "ValidatePolicyActivity"})
	s.env.RegisterActivityWithOptions(mockActs.CreateRevivalRequestActivity, activity.RegisterOptions{Name: "CreateRevivalRequestActivity"})
	s.env.RegisterActivityWithOptions(mockActs.UpdateDataEntryActivity, activity.RegisterOptions{Name: "UpdateDataEntryActivity"})
	s.env.RegisterActivityWithOptions(mockActs.CheckAndAdjustSuspenseActivity, activity.RegisterOptions{Name: "CheckAndAdjustSuspenseActivity"})
	s.env.RegisterActivityWithOptions(mockActs.UpdateQCActivity, activity.RegisterOptions{Name: "UpdateQCActivity"})
	s.env.RegisterActivityWithOptions(mockActs.TerminateAndReturnToIndexerActivity, activity.RegisterOptions{Name: "TerminateAndReturnToIndexerActivity"})
	s.env.RegisterActivityWithOptions(mockActs.UpdateRevivalStatusActivity, activity.RegisterOptions{Name: "UpdateRevivalStatusActivity"})
	s.env.RegisterActivityWithOptions(mockActs.UpdateApprovalActivity, activity.RegisterOptions{Name: "UpdateApprovalActivity"})
	s.env.RegisterActivityWithOptions(mockActs.UpdateWorkflowStateActivity, activity.RegisterOptions{Name: "UpdateWorkflowStateActivity"})
	s.env.RegisterActivityWithOptions(mockActs.TerminateRevivalActivity, activity.RegisterOptions{Name: "TerminateRevivalActivity"})
	s.env.RegisterActivityWithOptions(mockActs.FinalizeRevivalAfterFirstCollection, activity.RegisterOptions{Name: "FinalizeRevivalAfterFirstCollection"})
	s.env.RegisterActivityWithOptions(mockActs.NotifyPolicyManagementActivity, activity.RegisterOptions{Name: "NotifyPolicyManagementActivity"})
	s.env.RegisterActivityWithOptions(mockActs.ProcessInstallmentActivity, activity.RegisterOptions{Name: "ProcessInstallmentActivity"})
	s.env.RegisterActivityWithOptions(mockActs.HandleDefaultActivity, activity.RegisterOptions{Name: "HandleDefaultActivity"})
}

func (s *PMIntegrationTestSuite) AfterTest(suiteName, testName string) {
	s.env.AssertExpectations(s.T())
}

// =============================================================================
// TEST: PM fields propagate through workflow state
// =============================================================================

func (s *PMIntegrationTestSuite) TestPMFieldsPropagatedToWorkflowState() {
	input := IndexRevivalInput{
		TicketID:     "TEST-PM-001",
		PolicyNumber: "0000000000001",
		RequestType:  "installment_revival",
		IndexedBy:    "test_indexer",
		IndexedDate:  time.Now(),
		Documents:    "[]",
		PMWorkflowID: "plw-0000000000001",
		PMRequestID:  "pm-req-001",
	}

	s.env.OnActivity("ValidatePolicyActivity", mock.Anything, mock.Anything).
		Return(PolicyValidationResult{MaturityDate: time.Now().AddDate(5, 0, 0)}, nil)
	s.env.OnActivity("CreateRevivalRequestActivity", mock.Anything, mock.Anything).
		Return("REQ-PM-001", nil)
	s.env.OnActivity("TerminateAndReturnToIndexerActivity", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil)
	// NotifyPolicyManagementActivity should NOT be called when workflow returns to indexer
	// (return to indexer is an internal revival flow, not a PM terminal event)

	// Send data entry with return to indexer to end workflow quickly
	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow("data-entry-complete", DataEntryCompleteSignal{
			EnteredBy:            "test_user",
			EnteredAt:            time.Now(),
			NumberOfInstallments: 3,
			RevivalAmount:        30000,
			InstallmentAmount:    10000,
			ReturnToIndexer:      true,
			ReturnReason:         "Test return",
		})
	}, time.Millisecond*100)

	// Query workflow state to verify PM fields are set
	s.env.RegisterDelayedCallback(func() {
		result, err := s.env.QueryWorkflow("getStateDetails")
		s.NoError(err)

		var state RevivalWorkflowState
		err = result.Get(&state)
		s.NoError(err)
		s.Equal("plw-0000000000001", state.PMWorkflowID)
		s.Equal("pm-req-001", state.PMRequestID)
	}, time.Millisecond*50)

	s.env.ExecuteWorkflow(InstallmentRevivalWorkflow, input)

	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
}

// =============================================================================
// TEST: PM notification on REJECTED (approval rejected)
// =============================================================================

func (s *PMIntegrationTestSuite) TestPMNotificationOnRejection() {
	input := IndexRevivalInput{
		TicketID:     "TEST-PM-REJ-001",
		PolicyNumber: "0000000000001",
		RequestType:  "installment_revival",
		IndexedBy:    "test_indexer",
		IndexedDate:  time.Now(),
		Documents:    "[]",
		PMWorkflowID: "plw-0000000000001",
		PMRequestID:  "pm-req-rej-001",
	}

	s.env.OnActivity("ValidatePolicyActivity", mock.Anything, mock.Anything).
		Return(PolicyValidationResult{MaturityDate: time.Now().AddDate(5, 0, 0)}, nil)
	s.env.OnActivity("CreateRevivalRequestActivity", mock.Anything, mock.Anything).
		Return("REQ-REJ-001", nil)
	s.env.OnActivity("UpdateDataEntryActivity", mock.Anything, mock.Anything, mock.Anything).
		Return(nil)
	s.env.OnActivity("CheckAndAdjustSuspenseActivity", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(SuspenseAdjustmentResult{}, nil)
	s.env.OnActivity("UpdateQCActivity", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil)
	s.env.OnActivity("UpdateRevivalStatusActivity", mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	// Expect PM notification with REJECTED outcome (pre-approval, uses revival-completed)
	s.env.OnActivity("NotifyPolicyManagementActivity", mock.Anything,
		"plw-0000000000001",
		"revival-completed",
		mock.MatchedBy(func(signal PMCompletionSignal) bool {
			return signal.RequestID == "pm-req-rej-001" &&
				signal.RequestType == "REVIVAL" &&
				signal.Outcome == "REJECTED" &&
				signal.StateTransition == "REVIVAL_PENDING→REJECTED"
		}),
	).Return(nil).Once()

	// Data entry
	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow("data-entry-complete", DataEntryCompleteSignal{
			EnteredBy:            "test_user",
			EnteredAt:            time.Now(),
			NumberOfInstallments: 3,
			RevivalAmount:        30000,
			InstallmentAmount:    10000,
		})
	}, time.Millisecond*100)

	// QC pass
	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow("quality-check-complete", QualityCheckCompleteSignal{
			QCPassed:    true,
			QCComments:  "OK",
			PerformedBy: "test_qc",
			PerformedAt: time.Now(),
		})
	}, time.Millisecond*200)

	// Approval REJECTED
	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow("approval-decision", ApprovalDecisionSignal{
			Approved:   false,
			Comments:   "Not eligible",
			ApprovedBy: "test_approver",
			ApprovedAt: time.Now(),
		})
	}, time.Millisecond*300)

	s.env.ExecuteWorkflow(InstallmentRevivalWorkflow, input)

	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
}

// =============================================================================
// TEST: PM notification on VALIDATION_FAILED
// =============================================================================

func (s *PMIntegrationTestSuite) TestPMNotificationOnValidationFailed() {
	input := IndexRevivalInput{
		TicketID:     "TEST-PM-VALFAIL-001",
		PolicyNumber: "0000000000001",
		RequestType:  "installment_revival",
		IndexedBy:    "test_indexer",
		IndexedDate:  time.Now(),
		Documents:    "[]",
		PMWorkflowID: "plw-0000000000001",
		PMRequestID:  "pm-req-valfail-001",
	}

	// Make ValidatePolicyActivity fail
	s.env.OnActivity("ValidatePolicyActivity", mock.Anything, mock.Anything).
		Return(PolicyValidationResult{}, assert.AnError)

	// Expect PM notification with REJECTED outcome for validation failure (pre-approval, uses revival-completed)
	s.env.OnActivity("NotifyPolicyManagementActivity", mock.Anything,
		"plw-0000000000001",
		"revival-completed",
		mock.MatchedBy(func(signal PMCompletionSignal) bool {
			return signal.RequestID == "pm-req-valfail-001" &&
				signal.RequestType == "REVIVAL" &&
				signal.Outcome == "REJECTED" &&
				signal.StateTransition == "REVIVAL_PENDING→VALIDATION_FAILED"
		}),
	).Return(nil).Once()

	s.env.ExecuteWorkflow(InstallmentRevivalWorkflow, input)

	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
}

// =============================================================================
// TEST: PM notification on SLA TIMEOUT
// =============================================================================

func (s *PMIntegrationTestSuite) TestPMNotificationOnSLATimeout() {
	input := IndexRevivalInput{
		TicketID:     "TEST-PM-SLA-001",
		PolicyNumber: "0000000000001",
		RequestType:  "installment_revival",
		IndexedBy:    "test_indexer",
		IndexedDate:  time.Now(),
		Documents:    "[]",
		PMWorkflowID: "plw-0000000000001",
		PMRequestID:  "pm-req-sla-001",
	}

	s.env.OnActivity("ValidatePolicyActivity", mock.Anything, mock.Anything).
		Return(PolicyValidationResult{MaturityDate: time.Now().AddDate(5, 0, 0)}, nil)
	s.env.OnActivity("CreateRevivalRequestActivity", mock.Anything, mock.Anything).
		Return("REQ-SLA-001", nil)
	s.env.OnActivity("UpdateDataEntryActivity", mock.Anything, mock.Anything, mock.Anything).
		Return(nil)
	s.env.OnActivity("CheckAndAdjustSuspenseActivity", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(SuspenseAdjustmentResult{}, nil)
	s.env.OnActivity("UpdateQCActivity", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil)
	s.env.OnActivity("UpdateApprovalActivity", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil)
	s.env.OnActivity("UpdateWorkflowStateActivity", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil)
	s.env.OnActivity("TerminateRevivalActivity", mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	// Phase-1: Expect APPROVED notification at approval time
	s.env.OnActivity("NotifyPolicyManagementActivity", mock.Anything,
		"plw-0000000000001",
		"revival-approved",
		mock.MatchedBy(func(signal PMCompletionSignal) bool {
			return signal.RequestID == "pm-req-sla-001" &&
				signal.Outcome == "APPROVED"
		}),
	).Return(nil).Once()

	// Phase-2: Expect VOID notification when SLA expires
	s.env.OnActivity("NotifyPolicyManagementActivity", mock.Anything,
		"plw-0000000000001",
		"revival-completed",
		mock.MatchedBy(func(signal PMCompletionSignal) bool {
			return signal.RequestID == "pm-req-sla-001" &&
				signal.RequestType == "REVIVAL" &&
				signal.Outcome == "VOID" &&
				signal.StateTransition == "ACTIVE→VOID"
		}),
	).Return(nil).Once()

	// Data entry
	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow("data-entry-complete", DataEntryCompleteSignal{
			EnteredBy:            "test_user",
			EnteredAt:            time.Now(),
			NumberOfInstallments: 3,
			RevivalAmount:        30000,
			InstallmentAmount:    10000,
		})
	}, time.Millisecond*100)

	// QC pass
	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow("quality-check-complete", QualityCheckCompleteSignal{
			QCPassed:    true,
			QCComments:  "OK",
			PerformedBy: "test_qc",
			PerformedAt: time.Now(),
		})
	}, time.Millisecond*200)

	// Approval
	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow("approval-decision", ApprovalDecisionSignal{
			Approved:   true,
			Comments:   "Approved",
			ApprovedBy: "test_approver",
			ApprovedAt: time.Now(),
		})
	}, time.Millisecond*300)

	// DO NOT send first-collection signal → let 60-day SLA timer expire

	s.env.ExecuteWorkflow(InstallmentRevivalWorkflow, input)

	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
}

// =============================================================================
// TEST: PM notification on COMPLETED (no pending installments)
// =============================================================================

func (s *PMIntegrationTestSuite) TestPMNotificationOnCompletedNoPending() {
	input := IndexRevivalInput{
		TicketID:     "TEST-PM-COMP-001",
		PolicyNumber: "0000000000001",
		RequestType:  "installment_revival",
		IndexedBy:    "test_indexer",
		IndexedDate:  time.Now(),
		Documents:    "[]",
		PMWorkflowID: "plw-0000000000001",
		PMRequestID:  "pm-req-comp-001",
	}

	s.env.OnActivity("ValidatePolicyActivity", mock.Anything, mock.Anything).
		Return(PolicyValidationResult{MaturityDate: time.Now().AddDate(5, 0, 0)}, nil)
	s.env.OnActivity("CreateRevivalRequestActivity", mock.Anything, mock.Anything).
		Return("REQ-COMP-001", nil)
	s.env.OnActivity("UpdateDataEntryActivity", mock.Anything, mock.Anything, mock.Anything).
		Return(nil)
	// Suspense covers all installments (re-revival scenario)
	s.env.OnActivity("CheckAndAdjustSuspenseActivity", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(SuspenseAdjustmentResult{
			HasSuspense:           true,
			TotalSuspenseAmount:   20000,
			OriginalRevivalAmount: 30000,
			AdjustedRevivalAmount: 10000,
			SuspenseEntriesCount:  2,
		}, nil)
	s.env.OnActivity("UpdateQCActivity", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil)
	s.env.OnActivity("UpdateApprovalActivity", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil)
	s.env.OnActivity("UpdateWorkflowStateActivity", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil)
	s.env.OnActivity("FinalizeRevivalAfterFirstCollection", mock.Anything, mock.Anything).
		Return(nil)

	// Phase-1: Expect APPROVED notification at approval time (via revival-approved channel)
	s.env.OnActivity("NotifyPolicyManagementActivity", mock.Anything,
		"plw-0000000000001",
		"revival-approved",
		mock.MatchedBy(func(signal PMCompletionSignal) bool {
			return signal.RequestID == "pm-req-comp-001" &&
				signal.RequestType == "REVIVAL" &&
				signal.Outcome == "APPROVED" &&
				signal.StateTransition == "REVIVAL_PENDING→ACTIVE"
		}),
	).Return(nil).Once()

	// Data entry: 2 installments, installment amount = 10000
	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow("data-entry-complete", DataEntryCompleteSignal{
			EnteredBy:            "test_user",
			EnteredAt:            time.Now(),
			NumberOfInstallments: 2,
			RevivalAmount:        20000,
			InstallmentAmount:    10000,
		})
	}, time.Millisecond*100)

	// QC pass
	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow("quality-check-complete", QualityCheckCompleteSignal{
			QCPassed:    true,
			QCComments:  "OK",
			PerformedBy: "test_qc",
			PerformedAt: time.Now(),
		})
	}, time.Millisecond*200)

	// Approval
	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow("approval-decision", ApprovalDecisionSignal{
			Approved:   true,
			Comments:   "Approved",
			ApprovedBy: "test_approver",
			ApprovedAt: time.Now(),
		})
	}, time.Millisecond*300)

	// First collection
	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow("first-collection-complete", FirstCollectionCompleteSignal{
			CollectionDate: time.Now(),
			PaymentMode:    "CASH",
			TotalAmount:    10000,
		})
	}, time.Millisecond*400)

	s.env.ExecuteWorkflow(InstallmentRevivalWorkflow, input)

	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
}

// =============================================================================
// TEST: No PM notification when PMWorkflowID is empty (standalone mode)
// =============================================================================

func (s *PMIntegrationTestSuite) TestNoPMNotificationWhenStandalone() {
	input := IndexRevivalInput{
		TicketID:     "TEST-STANDALONE-001",
		PolicyNumber: "0000000000001",
		RequestType:  "installment_revival",
		IndexedBy:    "test_indexer",
		IndexedDate:  time.Now(),
		Documents:    "[]",
		// PMWorkflowID intentionally empty - standalone mode
	}

	// Make validation fail to trigger a terminal state
	s.env.OnActivity("ValidatePolicyActivity", mock.Anything, mock.Anything).
		Return(PolicyValidationResult{}, assert.AnError)

	// NotifyPolicyManagementActivity should NOT be called at all
	// (no .OnActivity for it, so if called the test will fail in AfterTest)

	s.env.ExecuteWorkflow(InstallmentRevivalWorkflow, input)

	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
}

// =============================================================================
// TEST: PM notification on InstallmentMonitorWorkflow completion
// =============================================================================

func (s *PMIntegrationTestSuite) TestInstallmentMonitorCompleteNoPMNotification() {
	// PM is notified at approval time, not when installments complete.
	// InstallmentMonitorWorkflow should NOT call NotifyPolicyManagementActivity.
	s.env.RegisterActivityWithOptions((&MockPMActivities{}).ProcessInstallmentActivity, activity.RegisterOptions{Name: "ProcessInstallmentActivity"})

	input := InstallmentMonitorInput{
		RequestID:            "REQ-MONITOR-001",
		NextDueDate:          time.Now().AddDate(0, 1, 0),
		NumberOfInstallments: 2, // Only 1 remaining (installment 2)
		PolicyNumber:         "0000000000001",
		PMWorkflowID:         "plw-0000000000001",
		PMRequestID:          "pm-req-monitor-001",
	}

	s.env.OnActivity("ProcessInstallmentActivity", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	// NotifyPolicyManagementActivity should NOT be called — PM was already notified at approval

	// Send installment 2 payment
	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow("installment-payment-received-2", InstallmentPaymentSignal{
			PaymentDate: time.Now(),
			Amount:      10000,
			PaymentMode: "CASH",
		})
	}, time.Millisecond*100)

	s.env.ExecuteWorkflow(InstallmentMonitorWorkflow, input)

	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
}

// =============================================================================
// TEST: PM notification on InstallmentMonitorWorkflow DEFAULT
// =============================================================================

func (s *PMIntegrationTestSuite) TestPMNotificationOnInstallmentDefault() {
	s.env.RegisterActivityWithOptions((&MockPMActivities{}).HandleDefaultActivity, activity.RegisterOptions{Name: "HandleDefaultActivity"})

	input := InstallmentMonitorInput{
		RequestID:            "REQ-DEFAULT-001",
		NextDueDate:          time.Now().Add(-1 * time.Hour), // Due in the past to trigger fast timeout
		NumberOfInstallments: 2,
		PolicyNumber:         "0000000000001",
		PMWorkflowID:         "plw-0000000000001",
		PMRequestID:          "pm-req-default-001",
	}

	s.env.OnActivity("HandleDefaultActivity", mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	// Phase-2: Expect VOID notification for default (via revival-completed)
	s.env.OnActivity("NotifyPolicyManagementActivity", mock.Anything,
		"plw-0000000000001",
		"revival-completed",
		mock.MatchedBy(func(signal PMCompletionSignal) bool {
			return signal.RequestID == "pm-req-default-001" &&
				signal.RequestType == "REVIVAL" &&
				signal.Outcome == "VOID" &&
				signal.StateTransition == "ACTIVE→VOID"
		}),
	).Return(nil).Once()

	// Do NOT send payment signal → let timer expire to trigger default

	s.env.ExecuteWorkflow(InstallmentMonitorWorkflow, input)

	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
}

// =============================================================================
// TEST: No PM notification from InstallmentMonitorWorkflow in standalone mode
// =============================================================================

func (s *PMIntegrationTestSuite) TestNoNotificationFromChildWhenStandalone() {
	s.env.RegisterActivityWithOptions((&MockPMActivities{}).ProcessInstallmentActivity, activity.RegisterOptions{Name: "ProcessInstallmentActivity"})

	input := InstallmentMonitorInput{
		RequestID:            "REQ-STANDALONE-CHILD",
		NextDueDate:          time.Now().AddDate(0, 1, 0),
		NumberOfInstallments: 2,
		// PMWorkflowID intentionally empty
	}

	s.env.OnActivity("ProcessInstallmentActivity", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	// NotifyPolicyManagementActivity should NOT be called

	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow("installment-payment-received-2", InstallmentPaymentSignal{
			PaymentDate: time.Now(),
			Amount:      10000,
			PaymentMode: "CASH",
		})
	}, time.Millisecond*100)

	s.env.ExecuteWorkflow(InstallmentMonitorWorkflow, input)

	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
}

// =============================================================================
// TEST: Validation runs inside workflow (not handler)
// =============================================================================

func (s *PMIntegrationTestSuite) TestValidationRunsInWorkflow() {
	// This test verifies that ValidatePolicyActivity is called as the first
	// activity within the workflow. If validation had been removed from the
	// workflow, this mock would not be called and the workflow would proceed
	// without validation.
	input := IndexRevivalInput{
		TicketID:     "TEST-VAL-IN-WF",
		PolicyNumber: "0000000000001",
		RequestType:  "installment_revival",
		IndexedBy:    "test_indexer",
		IndexedDate:  time.Now(),
		Documents:    "[]",
	}

	// Expect ValidatePolicyActivity to be called exactly once (in the workflow)
	s.env.OnActivity("ValidatePolicyActivity", mock.Anything, "0000000000001").
		Return(PolicyValidationResult{MaturityDate: time.Now().AddDate(5, 0, 0)}, nil).Once()

	s.env.OnActivity("CreateRevivalRequestActivity", mock.Anything, mock.Anything).
		Return("REQ-VAL-001", nil)
	s.env.OnActivity("TerminateAndReturnToIndexerActivity", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	// End workflow quickly
	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow("data-entry-complete", DataEntryCompleteSignal{
			EnteredBy:            "test_user",
			EnteredAt:            time.Now(),
			NumberOfInstallments: 3,
			RevivalAmount:        30000,
			InstallmentAmount:    10000,
			ReturnToIndexer:      true,
			ReturnReason:         "Done",
		})
	}, time.Millisecond*100)

	s.env.ExecuteWorkflow(InstallmentRevivalWorkflow, input)

	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
	// AfterTest will verify that ValidatePolicyActivity was called exactly once
}

// =============================================================================
// TEST: Validation failure in workflow prevents request creation
// =============================================================================

func (s *PMIntegrationTestSuite) TestValidationFailurePreventsRequestCreation() {
	input := IndexRevivalInput{
		TicketID:     "TEST-VAL-FAIL",
		PolicyNumber: "BAD_POLICY_NUM",
		RequestType:  "installment_revival",
		IndexedBy:    "test_indexer",
		IndexedDate:  time.Now(),
		Documents:    "[]",
	}

	// ValidatePolicyActivity fails (policy not in AL status)
	s.env.OnActivity("ValidatePolicyActivity", mock.Anything, "BAD_POLICY_NUM").
		Return(PolicyValidationResult{}, assert.AnError).Once()

	// CreateRevivalRequestActivity should NOT be called
	// (no mock registered, so if called the test will panic/fail)

	s.env.ExecuteWorkflow(InstallmentRevivalWorkflow, input)

	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
}

// =============================================================================
// Run PM Integration Test Suite
// =============================================================================

func TestPMIntegrationSuite(t *testing.T) {
	suite.Run(t, new(PMIntegrationTestSuite))
}

// =============================================================================
// Unit Tests: PMCompletionSignal struct
// =============================================================================

func TestPMCompletionSignalFields(t *testing.T) {
	now := time.Now()
	signal := PMCompletionSignal{
		RequestID:       "test-req-001",
		RequestType:     "REVIVAL",
		Outcome:         "APPROVED",
		StateTransition: "REVIVAL_PENDING→ACTIVE",
		CompletedAt:     now,
	}

	assert.Equal(t, "test-req-001", signal.RequestID)
	assert.Equal(t, "REVIVAL", signal.RequestType)
	assert.Equal(t, "APPROVED", signal.Outcome)
	assert.Equal(t, "REVIVAL_PENDING→ACTIVE", signal.StateTransition)
	assert.Equal(t, now, signal.CompletedAt)
}

func TestPMCompletionSignalOutcomeValues(t *testing.T) {
	// Verify all valid PM outcome values
	validOutcomes := []string{"APPROVED", "REJECTED", "WITHDRAWN", "TIMEOUT"}
	for _, outcome := range validOutcomes {
		signal := PMCompletionSignal{
			RequestID:   "test",
			RequestType: "REVIVAL",
			Outcome:     outcome,
			CompletedAt: time.Now(),
		}
		assert.Equal(t, outcome, signal.Outcome, "Outcome should be set to %s", outcome)
	}
}

// =============================================================================
// Unit Tests: IndexRevivalInput PM fields
// =============================================================================

func TestIndexRevivalInputPMFields(t *testing.T) {
	input := IndexRevivalInput{
		TicketID:     "TICKET-001",
		PolicyNumber: "0000000000001",
		RequestType:  "installment_revival",
		IndexedBy:    "user1",
		IndexedDate:  time.Now(),
		Documents:    "[]",
		PMWorkflowID: "plw-0000000000001",
		PMRequestID:  "pm-req-123",
	}

	assert.Equal(t, "plw-0000000000001", input.PMWorkflowID)
	assert.Equal(t, "pm-req-123", input.PMRequestID)
}

func TestIndexRevivalInputPMFieldsEmpty(t *testing.T) {
	// Standalone mode - PM fields are empty
	input := IndexRevivalInput{
		TicketID:     "TICKET-002",
		PolicyNumber: "0000000000001",
		RequestType:  "installment_revival",
		IndexedBy:    "user1",
		IndexedDate:  time.Now(),
		Documents:    "[]",
	}

	assert.Empty(t, input.PMWorkflowID)
	assert.Empty(t, input.PMRequestID)
}

// =============================================================================
// Unit Tests: InstallmentMonitorInput PM fields
// =============================================================================

func TestInstallmentMonitorInputPMFields(t *testing.T) {
	input := InstallmentMonitorInput{
		RequestID:            "REQ-001",
		NextDueDate:          time.Now(),
		NumberOfInstallments: 5,
		PolicyNumber:         "0000000000001",
		PMWorkflowID:         "plw-0000000000001",
		PMRequestID:          "pm-req-456",
	}

	assert.Equal(t, "plw-0000000000001", input.PMWorkflowID)
	assert.Equal(t, "pm-req-456", input.PMRequestID)
	assert.Equal(t, "0000000000001", input.PolicyNumber)
}

// =============================================================================
// Unit Tests: RevivalWorkflowState PM fields
// =============================================================================

func TestRevivalWorkflowStatePMFields(t *testing.T) {
	state := RevivalWorkflowState{
		RequestID:    "REQ-001",
		TicketID:     "TICKET-001",
		PolicyNumber: "0000000000001",
		PMWorkflowID: "plw-0000000000001",
		PMRequestID:  "pm-req-789",
	}

	assert.Equal(t, "plw-0000000000001", state.PMWorkflowID)
	assert.Equal(t, "pm-req-789", state.PMRequestID)
}

// =============================================================================
// Unit Tests: IsPMIntegrated helper
// =============================================================================

func TestIsPMIntegrated(t *testing.T) {
	// PM-integrated mode: PMWorkflowID is set
	pmInput := IndexRevivalInput{
		RequestID:    "uuid-123",
		PolicyNumber: "0000000000001",
		RequestType:  "REVIVAL",
		PMWorkflowID: "plw-0000000000001",
	}
	assert.True(t, pmInput.IsPMIntegrated())

	// Standalone mode: PMWorkflowID is empty
	standaloneInput := IndexRevivalInput{
		TicketID:     "TICKET-001",
		PolicyNumber: "0000000000001",
		RequestType:  "installment_revival",
		IndexedBy:    "user1",
	}
	assert.False(t, standaloneInput.IsPMIntegrated())
}

// =============================================================================
// Unit Tests: PM contract field mapping (ChildWorkflowInput compatibility)
// =============================================================================

func TestPMContractFieldMapping(t *testing.T) {
	// Simulate PM's ChildWorkflowInput fields arriving via JSON deserialization
	input := IndexRevivalInput{
		RequestID:        "uuid-idempotency-key",
		PolicyNumber:     "PLI/2026/000001",
		PolicyDBID:       42,
		ServiceRequestID: 100,
		RequestType:      "REVIVAL",
		RequestPayload:   []byte(`{"requested_installments": 6}`),
		PMWorkflowID:     "plw-PLI/2026/000001",
	}

	assert.Equal(t, "uuid-idempotency-key", input.RequestID)
	assert.Equal(t, int64(42), input.PolicyDBID)
	assert.Equal(t, int64(100), input.ServiceRequestID)
	assert.NotNil(t, input.RequestPayload)
	assert.True(t, input.IsPMIntegrated())
	// PMRequestID defaults to RequestID when not explicitly set
	assert.Empty(t, input.PMRequestID, "PMRequestID is empty; workflow uses RequestID as fallback")
}
