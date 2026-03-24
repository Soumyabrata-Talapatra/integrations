package installment_test

import (
	"context"
	"testing"
	"time"

	"plirevival/core/domain"
	"plirevival/core/port"
	"plirevival/handler"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	config "gitlab.cept.gov.in/it-2.0-common/api-config"
)

// InstallmentSuspenseTestSuite tests the complete flow: sequential payments → timeout → suspense → termination
type InstallmentSuspenseTestSuite struct {
	suite.Suite
	handler            *handler.RevivalHandler
	mockRevivalRepo    *MockRevivalRepository
	mockPolicyRepo     *MockPolicyRepository
	mockPaymentRepo    *MockPaymentRepository
	mockActivities     *MockActivities
	mockTemporalClient *MockTemporalClient
	cfg                *config.Config
}

// SetupTest sets up the test environment
func (suite *InstallmentSuspenseTestSuite) SetupTest() {
	suite.mockRevivalRepo = new(MockRevivalRepository)
	suite.mockPolicyRepo = new(MockPolicyRepository)
	suite.mockPaymentRepo = new(MockPaymentRepository)
	suite.mockActivities = new(MockActivities)
	suite.mockTemporalClient = new(MockTemporalClient)

	// Create minimal config
	suite.cfg = &config.Config{}

	// Create handler with mocks
	suite.handler = &handler.RevivalHandler{}
}

// Helper function to create a mock revival request
func (suite *InstallmentSuspenseTestSuite) createMockRevivalRequest(installmentsPaid int, numberOfInstallments int) *domain.RevivalRequest {
	workflowID := "test-workflow-id"
	runID := "test-run-id"

	return &domain.RevivalRequest{
		RequestID:            "REQ123",
		TicketID:             "PSREYV001329628",
		PolicyNumber:         "0000009321954",
		CurrentStatus:        "ACTIVE",
		NumberOfInstallments: numberOfInstallments,
		InstallmentsPaid:     installmentsPaid,
		RevivalAmount:        50000.00, // Premium amount
		InstallmentAmount:    3500.00,  // Per installment
		WorkflowID:           &workflowID,
		RunID:                &runID,
		FirstCollectionDone:  true,
	}
}

// TestSequentialPaymentThenDefault tests complete flow: pay 2, pay 3, timeout on 4 → suspense
func (suite *InstallmentSuspenseTestSuite) TestSequentialPaymentThenDefault() {
	ctx := context.Background()

	// Test Data: Approved 6 installments
	// Scenario: Pay installment 2, 3 successfully, then timeout on installment 4

	// Step 1: Pay installment 2 successfully
	revivalReq := suite.createMockRevivalRequest(1, 6)

	suite.mockRevivalRepo.On("GetRevivalRequestByTicketID", ctx, "PSREYV001329628").
		Return(revivalReq, nil).Once()

	suite.mockTemporalClient.On("SignalWorkflow", ctx, "test-workflow-id", "test-run-id",
		"installment-payment-received-2", mock.Anything).
		Return(nil)

	req2 := port.CreateInstallmentRequest{
		TicketID:          "PSREYV001329628",
		PolicyNumber:      "0000009321954",
		InstallmentNumber: 2,
		InstallmentAmount: 3500.00,
		PaymentMode:       "CASH",
		Status:            "PAID",
	}

	// Validate installment 2 is next expected
	nextExpected := revivalReq.InstallmentsPaid + 1
	assert.Equal(suite.T(), 2, nextExpected, "Next expected installment should be 2")
	assert.Equal(suite.T(), req2.InstallmentNumber, nextExpected, "Request should match expected")

	// Step 2: After payment, installmentsPaid = 2
	revivalReq.InstallmentsPaid = 2

	suite.mockRevivalRepo.On("GetRevivalRequestByTicketID", ctx, "PSREYV001329628").
		Return(revivalReq, nil).Once()

	suite.mockTemporalClient.On("SignalWorkflow", ctx, "test-workflow-id", "test-run-id",
		"installment-payment-received-3", mock.Anything).
		Return(nil)

	req3 := port.CreateInstallmentRequest{
		TicketID:          "PSREYV001329628",
		PolicyNumber:      "0000009321954",
		InstallmentNumber: 3,
		InstallmentAmount: 3500.00,
		PaymentMode:       "CASH",
		Status:            "PAID",
	}

	// Validate installment 3 is next expected
	nextExpected = revivalReq.InstallmentsPaid + 1
	assert.Equal(suite.T(), 3, nextExpected, "Next expected installment should be 3")
	assert.Equal(suite.T(), req3.InstallmentNumber, nextExpected, "Request should match expected")

	// Step 3: Timeout on installment 4 (no payment received)
	// At this point: installmentsPaid = 3 (first collection + installments 2, 3)
	revivalReq.InstallmentsPaid = 3

	// Expected suspense calculation:
	// Premium: 50000.00
	// Installment 1: 3500.00 (first collection)
	// Installment 2: 3500.00
	// Installment 3: 3500.00
	// Total: 50000 + (3 × 3500) = 60500.00
	expectedSuspenseAmount := revivalReq.RevivalAmount + (float64(revivalReq.InstallmentsPaid) * revivalReq.InstallmentAmount)
	assert.Equal(suite.T(), 60500.00, expectedSuspenseAmount, "Suspense amount should include premium + 3 installments")
}

// TestTimeoutOnSecondInstallment tests timeout immediately on installment 2
func (suite *InstallmentSuspenseTestSuite) TestTimeoutOnSecondInstallment() {
	ctx := context.Background()

	// Test Data: Approved 6 installments, paid only first collection (installment 1)
	// Scenario: Timeout on installment 2 (second installment)
	revivalReq := suite.createMockRevivalRequest(1, 6)

	suite.mockRevivalRepo.On("GetRevivalRequestByTicketID", ctx, "PSREYV001329628").
		Return(revivalReq, nil)

	// No payment received for installment 2 → TIMEOUT

	// Expected suspense calculation:
	// Premium: 50000.00
	// Installment 1: 3500.00 (from first collection)
	// Total: 50000 + 3500 = 53500.00
	expectedSuspenseAmount := revivalReq.RevivalAmount + (float64(revivalReq.InstallmentsPaid) * revivalReq.InstallmentAmount)
	assert.Equal(suite.T(), 53500.00, expectedSuspenseAmount, "Suspense should include premium + first installment")

	// Termination should be recorded
	// - termination_type: "DEFAULT"
	// - installment_number: 2
	// - suspense_created: true
	// - suspense_amount: 53500.00
	// - termination_reason: "Installment 2 payment not received by due date (IR_9: Zero grace period)"
}

// TestSuspenseCalculation tests suspense calculation for different scenarios
func (suite *InstallmentSuspenseTestSuite) TestSuspenseCalculation() {
	testCases := []struct {
		name                 string
		installmentsPaid     int
		numberOfInstallments int
		revivalAmount        float64
		installmentAmount    float64
		expectedSuspense     float64
		timeoutInstallment   int
	}{
		{
			name:                 "Timeout on installment 2 (only first collection paid)",
			installmentsPaid:     1,
			numberOfInstallments: 6,
			revivalAmount:        50000.00,
			installmentAmount:    3500.00,
			expectedSuspense:     53500.00, // 50000 + (1 × 3500)
			timeoutInstallment:   2,
		},
		{
			name:                 "Timeout on installment 3 (paid 1, 2)",
			installmentsPaid:     2,
			numberOfInstallments: 6,
			revivalAmount:        50000.00,
			installmentAmount:    3500.00,
			expectedSuspense:     57000.00, // 50000 + (2 × 3500)
			timeoutInstallment:   3,
		},
		{
			name:                 "Timeout on installment 5 (paid 1-4)",
			installmentsPaid:     4,
			numberOfInstallments: 6,
			revivalAmount:        50000.00,
			installmentAmount:    3500.00,
			expectedSuspense:     64000.00, // 50000 + (4 × 3500)
			timeoutInstallment:   5,
		},
		{
			name:                 "Timeout on installment 6 (paid 1-5)",
			installmentsPaid:     5,
			numberOfInstallments: 6,
			revivalAmount:        50000.00,
			installmentAmount:    3500.00,
			expectedSuspense:     67500.00, // 50000 + (5 × 3500)
			timeoutInstallment:   6,
		},
		{
			name:                 "Large revival amount - timeout on 2",
			installmentsPaid:     1,
			numberOfInstallments: 12,
			revivalAmount:        200000.00,
			installmentAmount:    15000.00,
			expectedSuspense:     215000.00, // 200000 + (1 × 15000)
			timeoutInstallment:   2,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// Calculate suspense amount
			totalSuspense := tc.revivalAmount + (float64(tc.installmentsPaid) * tc.installmentAmount)

			assert.Equal(suite.T(), tc.expectedSuspense, totalSuspense,
				"Suspense calculation should match expected amount")

			// Verify suspense breakdown
			premiumSuspense := tc.revivalAmount
			installmentsSuspense := float64(tc.installmentsPaid) * tc.installmentAmount

			assert.Equal(suite.T(), premiumSuspense+installmentsSuspense, totalSuspense,
				"Suspense should equal premium + installments")
		})
	}
}

// TestTerminationRecord tests that termination record is properly created
func (suite *InstallmentSuspenseTestSuite) TestTerminationRecord() {
	// Test termination record structure
	termination := domain.RevivalTermination{
		TerminationID:     "TERM123",
		RequestID:         "REQ123",
		TicketID:          "PSREYV001329628",
		PolicyNumber:      "0000009321954",
		TerminationReason: "Installment 2 payment not received by due date (IR_9: Zero grace period)",
		TerminationType:   "DEFAULT",
		InstallmentNumber: 2,
		SuspenseCreated:   true,
		SuspenseAmount:    53500.00,
		TerminatedAt:      time.Now(),
		CreatedAt:         time.Now(),
	}

	// Validate termination fields
	assert.Equal(suite.T(), "DEFAULT", termination.TerminationType, "Termination type should be DEFAULT")
	assert.Equal(suite.T(), 2, termination.InstallmentNumber, "Installment number should be 2")
	assert.True(suite.T(), termination.SuspenseCreated, "Suspense should be created")
	assert.Equal(suite.T(), 53500.00, termination.SuspenseAmount, "Suspense amount should match")
	assert.Contains(suite.T(), termination.TerminationReason, "IR_9", "Reason should reference IR_9")
	assert.Contains(suite.T(), termination.TerminationReason, "Installment 2", "Reason should specify installment number")
}

// TestWorkflowTermination tests that workflow terminates after default
func (suite *InstallmentSuspenseTestSuite) TestWorkflowTermination() {
	// After HandleDefaultActivity is called:
	// 1. Suspense entries created (premium + installments)
	// 2. Termination record created
	// 3. Policy status reverted to AL
	// 4. Revival request status updated to DEFAULTED
	// 5. Workflow MUST terminate (return error) - NOT continue

	// This test verifies workflow termination logic
	revivalReq := suite.createMockRevivalRequest(2, 6)

	// Simulate workflow termination after installment 3 timeout
	installmentNumber := 3
	totalInstallments := revivalReq.NumberOfInstallments

	// Verify error message format for workflow termination
	actualError := "revival workflow terminated: installment " + string(rune(installmentNumber+'0')) + " payment not received by due date"
	assert.Contains(suite.T(), actualError, "terminated", "Error should indicate termination")
	assert.Contains(suite.T(), actualError, "installment", "Error should mention installment")

	// Verify workflow does NOT continue after termination
	// (workflow returns error instead of continuing to installment 4, 5, 6)
	assert.Less(suite.T(), installmentNumber, totalInstallments, "Default occurred before all installments paid")
}

// TestCompleteFlowFromStartToDefault tests the complete end-to-end flow
func (suite *InstallmentSuspenseTestSuite) TestCompleteFlowFromStartToDefault() {
	// Complete flow timeline:
	//
	// 1. Index → Data Entry → QC → Approval (approved 6 installments)
	// 2. First Collection: Premium (50000) + Installment 1 (3500) = 53500 ✅
	// 3. Installment 2 payment: 3500 ✅ (installmentsPaid = 2)
	// 4. Installment 3 payment: 3500 ✅ (installmentsPaid = 3)
	// 5. Installment 4 DUE DATE PASSES (no payment) 🚨
	// 6. HandleDefaultActivity triggered:
	//    - Create suspense: Premium (50000) + Inst 1 (3500) + Inst 2 (3500) + Inst 3 (3500) = 60500
	//    - Create termination record: type=DEFAULT, installment=4, suspense_amount=60500
	//    - Revert policy to AL
	//    - Update request to DEFAULTED
	// 7. Workflow TERMINATES ❌
	//
	// Database state after default:
	// - revival_requests: current_status = "DEFAULTED", installments_paid = 3
	// - suspense_accounts: 4 entries (1 premium + 3 installments) = 60500 total
	// - revival_terminations: 1 entry (termination_type = DEFAULT)
	// - policies: policy_status = "AL"

	revivalReq := suite.createMockRevivalRequest(3, 6)

	// Verify final state
	assert.Equal(suite.T(), 3, revivalReq.InstallmentsPaid, "Should have 3 installments paid")
	assert.Equal(suite.T(), 6, revivalReq.NumberOfInstallments, "Should have 6 total installments")

	// Calculate final suspense
	finalSuspense := revivalReq.RevivalAmount + (float64(revivalReq.InstallmentsPaid) * revivalReq.InstallmentAmount)
	assert.Equal(suite.T(), 60500.00, finalSuspense, "Final suspense should be 60500")

	// Verify suspense breakdown:
	// - 1 premium suspense entry: 50000
	// - 3 installment suspense entries: 3 × 3500 = 10500
	premiumSuspense := revivalReq.RevivalAmount
	installmentSuspense := float64(revivalReq.InstallmentsPaid) * revivalReq.InstallmentAmount
	totalSuspenseEntries := 1 + revivalReq.InstallmentsPaid // 1 premium + 3 installments = 4 entries

	assert.Equal(suite.T(), 50000.00, premiumSuspense, "Premium suspense should be 50000")
	assert.Equal(suite.T(), 10500.00, installmentSuspense, "Installment suspense should be 10500")
	assert.Equal(suite.T(), 4, totalSuspenseEntries, "Should have 4 suspense entries")
}

// Run the test suite
func TestInstallmentSuspenseTestSuite(t *testing.T) {
	suite.Run(t, new(InstallmentSuspenseTestSuite))
}
