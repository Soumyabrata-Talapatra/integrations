package tests

import (
	"context"
	"testing"
	"time"

	"plirevival/core/domain"
	repo "plirevival/repo/postgres"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	config "gitlab.cept.gov.in/it-2.0-common/api-config"
)

// TestValidatePolicyForRevival_Eligible tests the batched validation query
// with a policy eligible for revival
func TestValidatePolicyForRevival_Eligible(t *testing.T) {
	// Setup
	cfg, err := config.NewDefaultConfigFactory().Create(
		config.WithFileName("config"),
		config.WithFilePaths(".", "../configs"),
	)
	require.NoError(t, err, "Failed to load config")

	database, container := SetUpDB(cfg)
	defer container.Terminate(context.Background())

	policyRepo := repo.NewPolicyRepository(database, cfg)
	ctx := context.Background()

	// Test Policy 1: Eligible for revival
	// - Lapsed status (AL)
	// - No ongoing revival requests
	// - Revival count (0) < max allowed (2)
	result, err := policyRepo.ValidatePolicyForRevival(ctx, "0000000000001")

	// Assertions
	require.NoError(t, err, "Batched validation query should succeed")

	// Verify policy data
	assert.Equal(t, "0000000000001", result.Policy.PolicyNumber)
	assert.Equal(t, "CUST0000000001", result.Policy.CustomerID)
	assert.Equal(t, "John Doe", result.Policy.CustomerName)
	assert.Equal(t, "AL", result.Policy.PolicyStatus, "Policy should be in lapsed status")
	assert.Equal(t, 0, result.Policy.RevivalCount, "Policy should have no previous revivals")

	// Verify validation data
	assert.Equal(t, 2, result.MaxRevivalsAllowed, "Max revivals should be 2 from config")
	assert.Equal(t, 0, result.OngoingRevivalCount, "Should have no ongoing revival requests")

	// Verify eligibility
	assert.True(t, result.Policy.PolicyStatus == "AL", "Policy should be lapsed")
	assert.True(t, result.OngoingRevivalCount == 0, "Should have no ongoing revivals")
	assert.True(t, result.Policy.RevivalCount < result.MaxRevivalsAllowed, "Should be under max revivals limit")

	t.Logf("✅ Test passed: Policy %s is eligible for revival", result.Policy.PolicyNumber)
	t.Logf("   Status: %s, Ongoing: %d, Revival Count: %d/%d",
		result.Policy.PolicyStatus, result.OngoingRevivalCount,
		result.Policy.RevivalCount, result.MaxRevivalsAllowed)
}

// TestValidatePolicyForRevival_NotLapsed tests validation with non-lapsed policy
func TestValidatePolicyForRevival_NotLapsed(t *testing.T) {
	// Setup
	cfg, err := config.NewDefaultConfigFactory().Create(
		config.WithFileName("config"),
		config.WithFilePaths(".", "../configs"),
	)
	require.NoError(t, err, "Failed to load config")

	database, container := SetUpDB(cfg)
	defer container.Terminate(context.Background())

	policyRepo := repo.NewPolicyRepository(database, cfg)
	ctx := context.Background()

	// Test Policy 2: Not lapsed (In Force status)
	result, err := policyRepo.ValidatePolicyForRevival(ctx, "0000000000002")

	// Assertions
	require.NoError(t, err, "Batched validation query should succeed")

	assert.Equal(t, "0000000000002", result.Policy.PolicyNumber)
	assert.Equal(t, "IF", result.Policy.PolicyStatus, "Policy should be In Force")

	// Verify NOT eligible due to status
	assert.False(t, result.Policy.PolicyStatus == "AL", "Policy should NOT be lapsed")

	t.Logf("✅ Test passed: Policy %s is NOT eligible (status: %s, expected: AL)",
		result.Policy.PolicyNumber, result.Policy.PolicyStatus)
}

// TestValidatePolicyForRevival_MaxRevivalsReached tests validation with max revivals reached
func TestValidatePolicyForRevival_MaxRevivalsReached(t *testing.T) {
	// Setup
	cfg, err := config.NewDefaultConfigFactory().Create(
		config.WithFileName("config"),
		config.WithFilePaths(".", "../configs"),
	)
	require.NoError(t, err, "Failed to load config")

	database, container := SetUpDB(cfg)
	defer container.Terminate(context.Background())

	policyRepo := repo.NewPolicyRepository(database, cfg)
	ctx := context.Background()

	// Test Policy 3: Max revivals reached
	result, err := policyRepo.ValidatePolicyForRevival(ctx, "0000000000003")

	// Assertions
	require.NoError(t, err, "Batched validation query should succeed")

	assert.Equal(t, "0000000000003", result.Policy.PolicyNumber)
	assert.Equal(t, "AL", result.Policy.PolicyStatus, "Policy should be lapsed")
	assert.Equal(t, 2, result.Policy.RevivalCount, "Policy should have 2 previous revivals")
	assert.Equal(t, 2, result.MaxRevivalsAllowed, "Max revivals should be 2")

	// Verify NOT eligible due to max revivals
	assert.False(t, result.Policy.RevivalCount < result.MaxRevivalsAllowed,
		"Policy should have reached max revivals limit")

	t.Logf("✅ Test passed: Policy %s is NOT eligible (revivals: %d/%d)",
		result.Policy.PolicyNumber, result.Policy.RevivalCount, result.MaxRevivalsAllowed)
}

// TestValidatePolicyForRevival_OngoingRevival tests validation with ongoing revival request
func TestValidatePolicyForRevival_OngoingRevival(t *testing.T) {
	// Setup
	cfg, err := config.NewDefaultConfigFactory().Create(
		config.WithFileName("config"),
		config.WithFilePaths(".", "../configs"),
	)
	require.NoError(t, err, "Failed to load config")

	database, container := SetUpDB(cfg)
	defer container.Terminate(context.Background())

	policyRepo := repo.NewPolicyRepository(database, cfg)
	ctx := context.Background()

	// Test Policy 4: Has ongoing revival request
	result, err := policyRepo.ValidatePolicyForRevival(ctx, "0000000000004")

	// Assertions
	require.NoError(t, err, "Batched validation query should succeed")

	assert.Equal(t, "0000000000004", result.Policy.PolicyNumber)
	assert.Equal(t, "AL", result.Policy.PolicyStatus, "Policy should be lapsed")
	assert.Equal(t, 1, result.OngoingRevivalCount, "Should have 1 ongoing revival request")

	// Verify NOT eligible due to ongoing revival
	assert.False(t, result.OngoingRevivalCount == 0,
		"Policy should have ongoing revival request")

	t.Logf("✅ Test passed: Policy %s is NOT eligible (ongoing revivals: %d)",
		result.Policy.PolicyNumber, result.OngoingRevivalCount)
}

// TestValidatePolicyForRevival_PolicyNotFound tests validation with non-existent policy
func TestValidatePolicyForRevival_PolicyNotFound(t *testing.T) {
	// Setup
	cfg, err := config.NewDefaultConfigFactory().Create(
		config.WithFileName("config"),
		config.WithFilePaths(".", "../configs"),
	)
	require.NoError(t, err, "Failed to load config")

	database, container := SetUpDB(cfg)
	defer container.Terminate(context.Background())

	policyRepo := repo.NewPolicyRepository(database, cfg)
	ctx := context.Background()

	// Test with non-existent policy
	_, err = policyRepo.ValidatePolicyForRevival(ctx, "9999999999999")

	// Should return error
	require.Error(t, err, "Should return error for non-existent policy")
	assert.Contains(t, err.Error(), "policy not found", "Error should indicate policy not found")

	t.Logf("✅ Test passed: Non-existent policy correctly returns error")
}

// TestValidatePolicyForRevival_PerformanceComparison compares batched vs individual queries
func TestValidatePolicyForRevival_PerformanceComparison(t *testing.T) {
	// Setup
	cfg, err := config.NewDefaultConfigFactory().Create(
		config.WithFileName("config"),
		config.WithFilePaths(".", "../configs"),
	)
	require.NoError(t, err, "Failed to load config")

	database, container := SetUpDB(cfg)
	defer container.Terminate(context.Background())

	policyRepo := repo.NewPolicyRepository(database, cfg)
	revivalRepo := repo.NewRevivalRepository(database, cfg)
	ctx := context.Background()

	policyNumber := "0000000000001"
	iterations := 100

	// =========================================================================
	// OLD APPROACH: 3 separate DB calls
	// =========================================================================
	startOld := time.Now()
	for i := 0; i < iterations; i++ {
		// 1st DB call
		policy, err := policyRepo.GetPolicyByNumber(ctx, policyNumber)
		require.NoError(t, err)

		// 2nd DB call
		hasOngoing, err := revivalRepo.CheckOngoingRevival(ctx, policyNumber)
		require.NoError(t, err)

		// 3rd DB call
		maxRevivals, err := policyRepo.GetMaxRevivalsAllowed(ctx)
		require.NoError(t, err)

		// Use the values to avoid compiler optimization
		_ = policy
		_ = hasOngoing
		_ = maxRevivals
	}
	oldDuration := time.Since(startOld)

	// =========================================================================
	// NEW APPROACH: Single batched query
	// =========================================================================
	startNew := time.Now()
	for i := 0; i < iterations; i++ {
		result, err := policyRepo.ValidatePolicyForRevival(ctx, policyNumber)
		require.NoError(t, err)

		// Use the value to avoid compiler optimization
		_ = result
	}
	newDuration := time.Since(startNew)

	// =========================================================================
	// Performance Analysis
	// =========================================================================
	improvement := float64(oldDuration-newDuration) / float64(oldDuration) * 100
	speedup := float64(oldDuration) / float64(newDuration)

	t.Logf("\n" + "=" + "==========================================================================")
	t.Logf("PERFORMANCE COMPARISON (%d iterations)", iterations)
	t.Logf("=" + "==========================================================================")
	t.Logf("OLD APPROACH (3 separate queries):")
	t.Logf("  Total Time: %v", oldDuration)
	t.Logf("  Avg/Query:  %v", oldDuration/time.Duration(iterations))
	t.Logf("  DB Calls:   %d (%d per iteration)", iterations*3, 3)
	t.Logf("")
	t.Logf("NEW APPROACH (1 batched query):")
	t.Logf("  Total Time: %v", newDuration)
	t.Logf("  Avg/Query:  %v", newDuration/time.Duration(iterations))
	t.Logf("  DB Calls:   %d (%d per iteration)", iterations*1, 1)
	t.Logf("")
	t.Logf("IMPROVEMENT:")
	t.Logf("  Time Saved:  %v (%.2f%% faster)", oldDuration-newDuration, improvement)
	t.Logf("  Speedup:     %.2fx", speedup)
	t.Logf("  DB Savings:  %d fewer queries (%.0f%% reduction)", iterations*2, 66.67)
	t.Logf("=" + "==========================================================================\n")

	// Assert that batched query is faster
	assert.True(t, newDuration < oldDuration,
		"Batched query should be faster than 3 separate queries")

	// At minimum, should see some improvement
	assert.Greater(t, improvement, 0.0,
		"Should see performance improvement with batched query")
}

// TestValidatePolicyForRevival_DataIntegrity verifies batched query returns same data as individual queries
func TestValidatePolicyForRevival_DataIntegrity(t *testing.T) {
	// Setup
	cfg, err := config.NewDefaultConfigFactory().Create(
		config.WithFileName("config"),
		config.WithFilePaths(".", "../configs"),
	)
	require.NoError(t, err, "Failed to load config")

	database, container := SetUpDB(cfg)
	defer container.Terminate(context.Background())

	policyRepo := repo.NewPolicyRepository(database, cfg)
	revivalRepo := repo.NewRevivalRepository(database, cfg)
	ctx := context.Background()

	policyNumber := "0000000000001"

	// Get data using OLD approach (3 queries)
	policy, err := policyRepo.GetPolicyByNumber(ctx, policyNumber)
	require.NoError(t, err)

	hasOngoing, err := revivalRepo.CheckOngoingRevival(ctx, policyNumber)
	require.NoError(t, err)

	maxRevivals, err := policyRepo.GetMaxRevivalsAllowed(ctx)
	require.NoError(t, err)

	// Get data using NEW approach (1 batched query)
	result, err := policyRepo.ValidatePolicyForRevival(ctx, policyNumber)
	require.NoError(t, err)

	// Verify data integrity - both approaches should return identical data
	assert.Equal(t, policy.PolicyNumber, result.Policy.PolicyNumber)
	assert.Equal(t, policy.CustomerID, result.Policy.CustomerID)
	assert.Equal(t, policy.CustomerName, result.Policy.CustomerName)
	assert.Equal(t, policy.PolicyStatus, result.Policy.PolicyStatus)
	assert.Equal(t, policy.RevivalCount, result.Policy.RevivalCount)
	assert.Equal(t, policy.PremiumAmount, result.Policy.PremiumAmount)

	ongoingCount := 0
	if hasOngoing {
		ongoingCount = 1
	}
	assert.Equal(t, ongoingCount, result.OngoingRevivalCount)
	assert.Equal(t, maxRevivals, result.MaxRevivalsAllowed)

	t.Logf("✅ Data integrity verified: Batched query returns identical data to individual queries")
}

// Helper function to compare policy data
func comparePolicies(t *testing.T, expected, actual domain.Policy) {
	assert.Equal(t, expected.PolicyNumber, actual.PolicyNumber, "PolicyNumber mismatch")
	assert.Equal(t, expected.CustomerID, actual.CustomerID, "CustomerID mismatch")
	assert.Equal(t, expected.CustomerName, actual.CustomerName, "CustomerName mismatch")
	assert.Equal(t, expected.ProductCode, actual.ProductCode, "ProductCode mismatch")
	assert.Equal(t, expected.ProductName, actual.ProductName, "ProductName mismatch")
	assert.Equal(t, expected.PolicyStatus, actual.PolicyStatus, "PolicyStatus mismatch")
	assert.Equal(t, expected.PremiumFrequency, actual.PremiumFrequency, "PremiumFrequency mismatch")
	assert.Equal(t, expected.RevivalCount, actual.RevivalCount, "RevivalCount mismatch")
}
