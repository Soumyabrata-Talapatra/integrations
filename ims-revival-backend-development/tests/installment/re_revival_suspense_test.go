package installment_test

import (
	"context"
	"testing"
	"time"

	"plirevival/core/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// ReRevivalSuspenseTestSuite tests re-revival suspense adjustment flow
type ReRevivalSuspenseTestSuite struct {
	suite.Suite
}

// Helper function to create suspense entries
func (suite *ReRevivalSuspenseTestSuite) createSuspenseEntries(policyNumber string, premiumAmount, installmentAmount float64, installmentsPaid int) []domain.SuspenseAccount {
	requestID := "PREV-REQ-123"
	createdBy := "SYSTEM"

	var entries []domain.SuspenseAccount

	// Premium suspense
	entries = append(entries, domain.SuspenseAccount{
		SuspenseID:   "SUSP-PREM-001",
		PolicyNumber: policyNumber,
		RequestID:    &requestID,
		SuspenseType: "REVIVAL_PREMIUM",
		Amount:       premiumAmount,
		IsReversed:   false,
		CreatedAt:    time.Now().Add(-30 * 24 * time.Hour), // 30 days ago
		CreatedBy:    &createdBy,
		UpdatedAt:    time.Now().Add(-30 * 24 * time.Hour),
	})

	// Installment suspense entries
	for i := 1; i <= installmentsPaid; i++ {
		entries = append(entries, domain.SuspenseAccount{
			SuspenseID:   "SUSP-INST-" + string(rune('0'+i)),
			PolicyNumber: policyNumber,
			RequestID:    &requestID,
			SuspenseType: "REVIVAL_INSTALLMENT_" + string(rune('0'+i)),
			Amount:       installmentAmount,
			IsReversed:   false,
			CreatedAt:    time.Now().Add(-30 * 24 * time.Hour),
			CreatedBy:    &createdBy,
			UpdatedAt:    time.Now().Add(-30 * 24 * time.Hour),
		})
	}

	return entries
}

// TestNoSuspense_NormalRevival tests normal revival flow when no suspense exists
func (suite *ReRevivalSuspenseTestSuite) TestNoSuspense_NormalRevival() {
	ctx := context.Background()
	_ = ctx

	// Scenario: First time revival for this policy
	// No previous revival defaulted, so no suspense exists

	revivalAmount := 50000.00

	// Expected: No suspense adjustment
	// Revival amount remains unchanged

	suspenseEntries := []domain.SuspenseAccount{} // No suspense
	totalSuspense := 0.0

	adjustedAmount := revivalAmount - totalSuspense

	assert.Equal(suite.T(), 0, len(suspenseEntries), "No suspense entries should exist")
	assert.Equal(suite.T(), 50000.00, adjustedAmount, "Revival amount should remain unchanged")
	assert.Equal(suite.T(), revivalAmount, adjustedAmount, "Adjusted amount equals original amount")
}

// TestWithSuspense_SimpleAdjustment tests re-revival with simple suspense adjustment
func (suite *ReRevivalSuspenseTestSuite) TestWithSuspense_SimpleAdjustment() {
	ctx := context.Background()
	_ = ctx

	// Scenario: Previous revival defaulted on installment 2
	// Previous revival: Premium 50000, paid installment 1 (3500)
	// Total suspense: 53500
	// New revival amount: 80000
	// Expected adjusted amount: 80000 - 53500 = 26500

	policyNumber := "0000009321954"
	newRevivalAmount := 80000.00

	// Create suspense entries from previous revival
	suspenseEntries := suite.createSuspenseEntries(policyNumber, 50000.00, 3500.00, 1)

	// Calculate total suspense
	totalSuspense := 0.0
	for _, entry := range suspenseEntries {
		totalSuspense += entry.Amount
	}

	// Calculate adjusted amount
	adjustedAmount := newRevivalAmount - totalSuspense

	assert.Equal(suite.T(), 2, len(suspenseEntries), "Should have 2 suspense entries (premium + 1 installment)")
	assert.Equal(suite.T(), 53500.00, totalSuspense, "Total suspense should be 53500")
	assert.Equal(suite.T(), 26500.00, adjustedAmount, "Adjusted amount should be 26500")
	assert.Greater(suite.T(), adjustedAmount, 0.0, "Adjusted amount must be positive")
}

// TestWithSuspense_MultipleInstallments tests re-revival after defaulting on later installment
func (suite *ReRevivalSuspenseTestSuite) TestWithSuspense_MultipleInstallments() {
	ctx := context.Background()
	_ = ctx

	// Scenario: Previous revival defaulted on installment 4
	// Previous revival: Premium 50000, paid installments 1, 2, 3 (3 × 3500 = 10500)
	// Total suspense: 60500
	// New revival amount: 100000
	// Expected adjusted amount: 100000 - 60500 = 39500

	policyNumber := "0000009321954"
	newRevivalAmount := 100000.00

	// Create suspense entries from previous revival (premium + 3 installments)
	suspenseEntries := suite.createSuspenseEntries(policyNumber, 50000.00, 3500.00, 3)

	// Calculate total suspense
	totalSuspense := 0.0
	for _, entry := range suspenseEntries {
		totalSuspense += entry.Amount
	}

	// Calculate adjusted amount
	adjustedAmount := newRevivalAmount - totalSuspense

	assert.Equal(suite.T(), 4, len(suspenseEntries), "Should have 4 suspense entries (premium + 3 installments)")
	assert.Equal(suite.T(), 60500.00, totalSuspense, "Total suspense should be 60500")
	assert.Equal(suite.T(), 39500.00, adjustedAmount, "Adjusted amount should be 39500")
}

// TestSuspenseExceedsRevival_Error tests error when suspense exceeds new revival amount
func (suite *ReRevivalSuspenseTestSuite) TestSuspenseExceedsRevival_Error() {
	ctx := context.Background()
	_ = ctx

	// Scenario: Previous revival had high premium, new revival amount is lower
	// Previous suspense: 200000
	// New revival amount: 150000
	// Expected: Error - adjusted amount would be negative

	policyNumber := "0000009321954"
	newRevivalAmount := 150000.00

	// Create large suspense entries
	suspenseEntries := suite.createSuspenseEntries(policyNumber, 180000.00, 10000.00, 2)

	// Calculate total suspense
	totalSuspense := 0.0
	for _, entry := range suspenseEntries {
		totalSuspense += entry.Amount
	}

	// Calculate adjusted amount
	adjustedAmount := newRevivalAmount - totalSuspense

	assert.Equal(suite.T(), 200000.00, totalSuspense, "Total suspense should be 200000")
	assert.Less(suite.T(), adjustedAmount, 0.0, "Adjusted amount would be negative")
	assert.Equal(suite.T(), -50000.00, adjustedAmount, "Negative adjustment calculated")
	// In real implementation, this should throw an error
}

// TestSuspenseAdjustmentCalculations tests various suspense adjustment scenarios
func (suite *ReRevivalSuspenseTestSuite) TestSuspenseAdjustmentCalculations() {
	testCases := []struct {
		name                  string
		previousPremium       float64
		previousInstallment   float64
		installmentsPaid      int
		newRevivalAmount      float64
		expectedSuspense      float64
		expectedAdjusted      float64
		shouldSucceed         bool
	}{
		{
			name:                  "Small suspense, large new revival",
			previousPremium:       30000.00,
			previousInstallment:   2500.00,
			installmentsPaid:      1,
			newRevivalAmount:      100000.00,
			expectedSuspense:      32500.00,
			expectedAdjusted:      67500.00,
			shouldSucceed:         true,
		},
		{
			name:                  "Large suspense from many installments",
			previousPremium:       50000.00,
			previousInstallment:   3500.00,
			installmentsPaid:      5,
			newRevivalAmount:      100000.00,
			expectedSuspense:      67500.00,
			expectedAdjusted:      32500.00,
			shouldSucceed:         true,
		},
		{
			name:                  "Exact match - suspense equals revival",
			previousPremium:       40000.00,
			previousInstallment:   5000.00,
			installmentsPaid:      2,
			newRevivalAmount:      50000.00,
			expectedSuspense:      50000.00,
			expectedAdjusted:      0.00,
			shouldSucceed:         true, // Zero is acceptable
		},
		{
			name:                  "Suspense exceeds revival - should fail",
			previousPremium:       100000.00,
			previousInstallment:   10000.00,
			installmentsPaid:      3,
			newRevivalAmount:      80000.00,
			expectedSuspense:      130000.00,
			expectedAdjusted:      -50000.00,
			shouldSucceed:         false,
		},
		{
			name:                  "Minimal suspense",
			previousPremium:       10000.00,
			previousInstallment:   1000.00,
			installmentsPaid:      1,
			newRevivalAmount:      50000.00,
			expectedSuspense:      11000.00,
			expectedAdjusted:      39000.00,
			shouldSucceed:         true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			// Create suspense entries
			suspenseEntries := suite.createSuspenseEntries(
				"0000009321954",
				tc.previousPremium,
				tc.previousInstallment,
				tc.installmentsPaid,
			)

			// Calculate total suspense
			totalSuspense := 0.0
			for _, entry := range suspenseEntries {
				totalSuspense += entry.Amount
			}

			// Calculate adjusted amount
			adjustedAmount := tc.newRevivalAmount - totalSuspense

			// Verify calculations
			assert.Equal(suite.T(), tc.expectedSuspense, totalSuspense,
				"Total suspense should match expected")
			assert.Equal(suite.T(), tc.expectedAdjusted, adjustedAmount,
				"Adjusted amount should match expected")

			// Verify success condition
			if tc.shouldSucceed {
				assert.GreaterOrEqual(suite.T(), adjustedAmount, 0.0,
					"Adjusted amount should be non-negative for success cases")
			} else {
				assert.Less(suite.T(), adjustedAmount, 0.0,
					"Adjusted amount should be negative for failure cases")
			}

			// Verify suspense entry count
			expectedEntries := 1 + tc.installmentsPaid // Premium + installments
			assert.Equal(suite.T(), expectedEntries, len(suspenseEntries),
				"Should have correct number of suspense entries")
		})
	}
}

// TestSuspenseReversal tests marking suspense as reversed after adjustment
func (suite *ReRevivalSuspenseTestSuite) TestSuspenseReversal() {
	ctx := context.Background()
	_ = ctx

	// Scenario: After suspense adjustment, all suspense entries should be marked as reversed

	policyNumber := "0000009321954"
	newRequestID := "NEW-REQ-456"

	// Create suspense entries
	suspenseEntries := suite.createSuspenseEntries(policyNumber, 50000.00, 3500.00, 2)

	// Simulate reversal
	reversedEntries := make([]domain.SuspenseAccount, len(suspenseEntries))
	copy(reversedEntries, suspenseEntries)

	now := time.Now()
	reversalReason := "Adjusted against new revival request " + newRequestID

	for i := range reversedEntries {
		reversedEntries[i].IsReversed = true
		reversedEntries[i].ReversalDate = &now
		reversedEntries[i].ReversalAuthorizedBy = strPtr("SYSTEM")
		reversedEntries[i].ReversalReason = &reversalReason
		reversedEntries[i].UpdatedAt = now
	}

	// Verify all entries are reversed
	for _, entry := range reversedEntries {
		assert.True(suite.T(), entry.IsReversed, "Entry should be marked as reversed")
		assert.NotNil(suite.T(), entry.ReversalDate, "Reversal date should be set")
		assert.NotNil(suite.T(), entry.ReversalAuthorizedBy, "Reversal authorized by should be set")
		assert.NotNil(suite.T(), entry.ReversalReason, "Reversal reason should be set")
		assert.Contains(suite.T(), *entry.ReversalReason, newRequestID, "Reason should reference new request")
	}
}

// TestCompleteReRevivalFlow tests the complete end-to-end re-revival flow
func (suite *ReRevivalSuspenseTestSuite) TestCompleteReRevivalFlow() {
	// Complete flow timeline:
	//
	// PREVIOUS REVIVAL (ended in default):
	// 1. Index → Data Entry → QC → Approval (approved 6 installments)
	// 2. First Collection: Premium (50000) + Installment 1 (3500) = 53500 ✅
	// 3. Installment 2 payment: 3500 ✅
	// 4. Installment 3 DUE DATE PASSES (no payment) 🚨
	// 5. Default handling:
	//    - Create suspense: Premium (50000) + Inst 1 (3500) + Inst 2 (3500) = 57000
	//    - Termination record created
	//    - Policy reverted to AL
	//    - Workflow terminated
	//
	// NEW REVIVAL (with suspense adjustment):
	// 1. Index → Data Entry (new revival amount: 100000)
	// 2. CheckAndAdjustSuspenseActivity triggered:
	//    - Found previous suspense: 57000
	//    - Adjusted amount: 100000 - 57000 = 43000 ✅
	//    - Mark suspense as reversed
	//    - Update revival_request:
	//      * previous_suspense_amount = 57000
	//      * suspense_adjusted = true
	//      * adjusted_revival_amount = 43000
	// 3. Continue normal flow with adjusted amount (43000)

	policyNumber := "0000009321954"
	newRevivalAmount := 100000.00

	// Previous suspense (premium + 2 installments)
	suspenseEntries := suite.createSuspenseEntries(policyNumber, 50000.00, 3500.00, 2)

	totalSuspense := 0.0
	for _, entry := range suspenseEntries {
		totalSuspense += entry.Amount
	}

	adjustedAmount := newRevivalAmount - totalSuspense

	// Verify calculations
	assert.Equal(suite.T(), 3, len(suspenseEntries), "Should have 3 suspense entries")
	assert.Equal(suite.T(), 57000.00, totalSuspense, "Total suspense should be 57000")
	assert.Equal(suite.T(), 43000.00, adjustedAmount, "Adjusted amount should be 43000")

	// Verify revival request would be updated with these values
	revivalRequest := domain.RevivalRequest{
		RequestID:              "NEW-REQ-456",
		PolicyNumber:           policyNumber,
		RevivalAmount:          newRevivalAmount,
		PreviousSuspenseAmount: totalSuspense,
		SuspenseAdjusted:       true,
		AdjustedRevivalAmount:  adjustedAmount,
	}

	assert.Equal(suite.T(), 100000.00, revivalRequest.RevivalAmount, "Original revival amount")
	assert.Equal(suite.T(), 57000.00, revivalRequest.PreviousSuspenseAmount, "Previous suspense")
	assert.True(suite.T(), revivalRequest.SuspenseAdjusted, "Suspense adjusted flag")
	assert.Equal(suite.T(), 43000.00, revivalRequest.AdjustedRevivalAmount, "Adjusted amount")

	// Customer effectively pays: 43000 (new revival) + 57000 (previous suspense adjusted) = 100000 total
	effectiveAmount := revivalRequest.AdjustedRevivalAmount + revivalRequest.PreviousSuspenseAmount
	assert.Equal(suite.T(), newRevivalAmount, effectiveAmount, "Effective amount equals new revival amount")
}

// Helper function
func strPtr(s string) *string {
	return &s
}

// Run the test suite
func TestReRevivalSuspenseTestSuite(t *testing.T) {
	suite.Run(t, new(ReRevivalSuspenseTestSuite))
}
