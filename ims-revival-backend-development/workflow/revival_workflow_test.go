package workflow

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/testsuite"
)

// MockActivities provides mock implementations for workflow activities
type MockActivities struct{}

// ValidatePolicyActivity mock
func (m *MockActivities) ValidatePolicyActivity(ctx context.Context, policyNumber string) (PolicyValidationResult, error) {
	return PolicyValidationResult{}, nil
}

// CreateRevivalRequestActivity mock
func (m *MockActivities) CreateRevivalRequestActivity(ctx context.Context, input RevivalRequestInput) (string, error) {
	return "", nil
}

// UpdateDataEntryActivity mock
func (m *MockActivities) UpdateDataEntryActivity(ctx context.Context, requestID string, input DataEntryInput) error {
	return nil
}

// CheckAndAdjustSuspenseActivity mock
func (m *MockActivities) CheckAndAdjustSuspenseActivity(ctx context.Context, requestID, policyNumber string, revivalAmount float64) (SuspenseAdjustmentResult, error) {
	return SuspenseAdjustmentResult{}, nil
}

// UpdateQCActivity mock
func (m *MockActivities) UpdateQCActivity(ctx context.Context, requestID, qcPerformedBy, qcComments string, qcPassed bool, missingDocuments string) error {
	return nil
}

// TerminateAndReturnToIndexerActivity mock
func (m *MockActivities) TerminateAndReturnToIndexerActivity(ctx context.Context, requestID string, reason string, stage string) error {
	return nil
}

// RevivalWorkflowTestSuite is a test suite for the InstallmentRevivalWorkflow
type RevivalWorkflowTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
	env *testsuite.TestWorkflowEnvironment
}

// SetupTest sets up the test environment before each test
func (s *RevivalWorkflowTestSuite) SetupTest() {
	s.env = s.NewTestWorkflowEnvironment()

	// Register mock activities with explicit names matching workflow's string-based calls
	mockActs := &MockActivities{}
	s.env.RegisterActivityWithOptions(mockActs.ValidatePolicyActivity, activity.RegisterOptions{Name: "ValidatePolicyActivity"})
	s.env.RegisterActivityWithOptions(mockActs.CreateRevivalRequestActivity, activity.RegisterOptions{Name: "CreateRevivalRequestActivity"})
	s.env.RegisterActivityWithOptions(mockActs.UpdateDataEntryActivity, activity.RegisterOptions{Name: "UpdateDataEntryActivity"})
	s.env.RegisterActivityWithOptions(mockActs.CheckAndAdjustSuspenseActivity, activity.RegisterOptions{Name: "CheckAndAdjustSuspenseActivity"})
	s.env.RegisterActivityWithOptions(mockActs.UpdateQCActivity, activity.RegisterOptions{Name: "UpdateQCActivity"})
	s.env.RegisterActivityWithOptions(mockActs.TerminateAndReturnToIndexerActivity, activity.RegisterOptions{Name: "TerminateAndReturnToIndexerActivity"})
}

// AfterTest cleans up after each test
func (s *RevivalWorkflowTestSuite) AfterTest(suiteName, testName string) {
	s.env.AssertExpectations(s.T())
}

// =============================================================================
// TEST: Data Entry ReturnToIndexer Flow
// =============================================================================

// TestDataEntryReturnToIndexer tests that workflow terminates when data entry
// sends ReturnToIndexer=true signal
func (s *RevivalWorkflowTestSuite) TestDataEntryReturnToIndexer() {
	// Setup test input
	input := IndexRevivalInput{
		TicketID:     "TEST-TICKET-001",
		PolicyNumber: "0000000000001",
		RequestType:  "installment_revival",
		IndexedBy:    "test_indexer",
		IndexedDate:  time.Now(),
		Documents:    "[]",
	}

	// Mock activities using string names to match workflow's activity calls
	s.env.OnActivity("ValidatePolicyActivity", mock.Anything, mock.Anything).
		Return(PolicyValidationResult{
			MaturityDate: time.Now().AddDate(5, 0, 0),
		}, nil)

	s.env.OnActivity("CreateRevivalRequestActivity", mock.Anything, mock.Anything).
		Return("REQ-001", nil)

	s.env.OnActivity("TerminateAndReturnToIndexerActivity", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	// Register callback to send signal after workflow starts waiting
	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow("data-entry-complete", DataEntryCompleteSignal{
			EnteredBy:            "test_data_entry",
			EnteredAt:            time.Now(),
			NumberOfInstallments: 12,
			RevivalAmount:        100000.00,
			InstallmentAmount:    8500.00,
			ReturnToIndexer:      true,
			ReturnReason:         "Missing documents: ID proof",
			MissingDocuments:     `[{"document_name":"ID_PROOF","status":"missing"},{"document_name":"ADDRESS_PROOF","status":"missing"}]`,
		})
	}, time.Millisecond*100)

	// Execute workflow
	s.env.ExecuteWorkflow(InstallmentRevivalWorkflow, input)

	// Assert workflow completed successfully
	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
}

// =============================================================================
// TEST: QC ReturnToIndexer Flow
// =============================================================================

// TestQCReturnToIndexer tests that workflow terminates when QC sends ReturnToIndexer=true
func (s *RevivalWorkflowTestSuite) TestQCReturnToIndexer() {
	input := IndexRevivalInput{
		TicketID:     "TEST-TICKET-003",
		PolicyNumber: "0000000000001",
		RequestType:  "installment_revival",
		IndexedBy:    "test_indexer",
		IndexedDate:  time.Now(),
		Documents:    "[]",
	}

	// Mock activities using string names
	s.env.OnActivity("ValidatePolicyActivity", mock.Anything, mock.Anything).
		Return(PolicyValidationResult{
			MaturityDate: time.Now().AddDate(5, 0, 0),
		}, nil)

	s.env.OnActivity("CreateRevivalRequestActivity", mock.Anything, mock.Anything).
		Return("REQ-003", nil)

	s.env.OnActivity("UpdateDataEntryActivity", mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	s.env.OnActivity("CheckAndAdjustSuspenseActivity", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(SuspenseAdjustmentResult{}, nil)

	s.env.OnActivity("TerminateAndReturnToIndexerActivity", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	// Send data entry signal first (normal flow)
	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow("data-entry-complete", DataEntryCompleteSignal{
			EnteredBy:            "test_data_entry",
			EnteredAt:            time.Now(),
			NumberOfInstallments: 12,
			RevivalAmount:        100000.00,
			InstallmentAmount:    8500.00,
			ReturnToIndexer:      false,
		})
	}, time.Millisecond*100)

	// Send QC signal with ReturnToIndexer=true
	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow("quality-check-complete", QualityCheckCompleteSignal{
			QCPassed:         false,
			QCComments:       "Documents not matching policy records",
			PerformedBy:      "test_qc",
			PerformedAt:      time.Now(),
			ReturnToIndexer:  true,
			ReturnReason:     "Documents not matching policy records",
			MissingDocuments: `[{"document_name":"INCOME_PROOF","status":"missing"}]`,
		})
	}, time.Millisecond*200)

	// Execute workflow
	s.env.ExecuteWorkflow(InstallmentRevivalWorkflow, input)

	// Assert workflow completed
	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
}

// =============================================================================
// TEST: Approval ReturnToIndexer Flow
// =============================================================================

// TestApprovalReturnToIndexer tests that workflow terminates when approver sends ReturnToIndexer=true
func (s *RevivalWorkflowTestSuite) TestApprovalReturnToIndexer() {
	input := IndexRevivalInput{
		TicketID:     "TEST-TICKET-005",
		PolicyNumber: "0000000000001",
		RequestType:  "installment_revival",
		IndexedBy:    "test_indexer",
		IndexedDate:  time.Now(),
		Documents:    "[]",
	}

	// Mock activities using string names
	s.env.OnActivity("ValidatePolicyActivity", mock.Anything, mock.Anything).
		Return(PolicyValidationResult{
			MaturityDate: time.Now().AddDate(5, 0, 0),
		}, nil)

	s.env.OnActivity("CreateRevivalRequestActivity", mock.Anything, mock.Anything).
		Return("REQ-005", nil)

	s.env.OnActivity("UpdateDataEntryActivity", mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	s.env.OnActivity("CheckAndAdjustSuspenseActivity", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(SuspenseAdjustmentResult{}, nil)

	s.env.OnActivity("UpdateQCActivity", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	s.env.OnActivity("TerminateAndReturnToIndexerActivity", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	// Send data entry signal
	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow("data-entry-complete", DataEntryCompleteSignal{
			EnteredBy:            "test_data_entry",
			EnteredAt:            time.Now(),
			NumberOfInstallments: 12,
			RevivalAmount:        100000.00,
			InstallmentAmount:    8500.00,
			ReturnToIndexer:      false,
		})
	}, time.Millisecond*100)

	// Send QC passed signal
	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow("quality-check-complete", QualityCheckCompleteSignal{
			QCPassed:        true,
			QCComments:      "All documents verified",
			PerformedBy:     "test_qc",
			PerformedAt:     time.Now(),
			ReturnToIndexer: false,
		})
	}, time.Millisecond*200)

	// Send approval signal with ReturnToIndexer=true
	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow("approval-decision", ApprovalDecisionSignal{
			Approved:        false,
			Comments:        "Policy flagged for fraud review",
			ApprovedBy:      "test_approver",
			ApprovedAt:      time.Now(),
			ReturnToIndexer: true,
			ReturnReason:    "Policy flagged for fraud review",
		})
	}, time.Millisecond*300)

	// Execute workflow
	s.env.ExecuteWorkflow(InstallmentRevivalWorkflow, input)

	// Assert workflow completed
	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
}

// =============================================================================
// TEST: validateMaturityDateConstraint helper function
// =============================================================================

func TestValidateMaturityDateConstraint_Valid(t *testing.T) {
	// Maturity date 2 years from now, 12 installments
	maturityDate := time.Now().AddDate(2, 0, 0)
	numberOfInstallments := 12
	firstDueDate := time.Now()

	err := validateMaturityDateConstraint(maturityDate, numberOfInstallments, firstDueDate)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
}

func TestValidateMaturityDateConstraint_ExceedsMaturity(t *testing.T) {
	// Maturity date 6 months from now, 12 installments (would exceed)
	maturityDate := time.Now().AddDate(0, 6, 0)
	numberOfInstallments := 12
	firstDueDate := time.Now()

	err := validateMaturityDateConstraint(maturityDate, numberOfInstallments, firstDueDate)
	if err == nil {
		t.Error("Expected IR_4 violation error, got nil")
	}
}

func TestValidateMaturityDateConstraint_SameMonth(t *testing.T) {
	// Last installment falls in maturity month (should fail)
	maturityDate := time.Date(2026, 6, 15, 0, 0, 0, 0, time.Local)
	numberOfInstallments := 6
	firstDueDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.Local) // Last due: June 1, 2026

	err := validateMaturityDateConstraint(maturityDate, numberOfInstallments, firstDueDate)
	if err == nil {
		t.Error("Expected IR_4 violation error for same month, got nil")
	}
}

// =============================================================================
// Run Test Suite
// =============================================================================

func TestRevivalWorkflowSuite(t *testing.T) {
	suite.Run(t, new(RevivalWorkflowTestSuite))
}
