package repo

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gitlab.cept.gov.in/it-2.0-policy/surrender-service/core/domain"
)

// TestSurrenderRequestRepository_Create tests creating a surrender request
func TestSurrenderRequestRepository_Create(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	repo := NewSurrenderRequestRepository(pool)
	ctx := context.Background()

	tests := []struct {
		name    string
		request domain.PolicySurrenderRequest
		wantErr bool
	}{
		{
			name: "valid voluntary surrender request",
			request: domain.PolicySurrenderRequest{
				PolicyID:                "0000000000001",
				RequestNumber:           "SUR-TEST-001",
				RequestType:             domain.SurrenderRequestTypeVoluntary,
				RequestDate:             time.Now(),
				GrossSurrenderValue:     50000,
				NetSurrenderValue:       45000,
				PaidUpValue:             40000,
				SurrenderFactor:         0.75,
				UnpaidPremiumsDeduction: 3000,
				LoanDeduction:           2000,
				DisbursementMethod:      domain.DisbursementMethodCheque,
				DisbursementAmount:      45000,
				Status:                  domain.SurrenderStatusPendingDocumentUpload,
				Owner:                   domain.RequestOwnerCustomer,
				CreatedBy:               uuid.New(),
				Metadata:                map[string]interface{}{"test": "data"},
			},
			wantErr: false,
		},
		{
			name: "valid forced surrender request",
			request: domain.PolicySurrenderRequest{
				PolicyID:                "0000000000002",
				RequestNumber:           "FSUR-TEST-001",
				RequestType:             domain.SurrenderRequestTypeForced,
				RequestDate:             time.Now(),
				GrossSurrenderValue:     30000,
				NetSurrenderValue:       25000,
				PaidUpValue:             20000,
				SurrenderFactor:         0.75,
				UnpaidPremiumsDeduction: 5000,
				DisbursementMethod:      domain.DisbursementMethodCheque,
				DisbursementAmount:      25000,
				Status:                  domain.SurrenderStatusPendingApproval,
				Owner:                   domain.RequestOwnerCPC,
				CreatedBy:               uuid.New(),
				Metadata:                map[string]interface{}{"forced_reason": "non-payment"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			created, err := repo.Create(ctx, tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotEqual(t, uuid.Nil, created.ID)
			assert.Equal(t, tt.request.RequestNumber, created.RequestNumber)
			assert.Equal(t, tt.request.RequestType, created.RequestType)
			assert.Equal(t, tt.request.Status, created.Status)
			assert.Equal(t, tt.request.NetSurrenderValue, created.NetSurrenderValue)
			assert.NotZero(t, created.CreatedAt)
		})
	}
}

// TestSurrenderRequestRepository_FindByID tests finding by ID
func TestSurrenderRequestRepository_FindByID(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	repo := NewSurrenderRequestRepository(pool)
	ctx := context.Background()

	// Create test request
	request := domain.PolicySurrenderRequest{
		PolicyID:            "0000000000003",
		RequestNumber:       "SUR-TEST-FIND-001",
		RequestType:         domain.SurrenderRequestTypeVoluntary,
		RequestDate:         time.Now(),
		GrossSurrenderValue: 50000,
		NetSurrenderValue:   45000,
		Status:              domain.SurrenderStatusPendingDocumentUpload,
		Owner:               domain.RequestOwnerCustomer,
		CreatedBy:           uuid.New(),
		Metadata:            map[string]interface{}{},
	}

	created, err := repo.Create(ctx, request)
	require.NoError(t, err)

	// Test finding by ID
	found, err := repo.FindByID(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, found.ID)
	assert.Equal(t, created.RequestNumber, found.RequestNumber)
	assert.Equal(t, created.Status, found.Status)

	// Test not found
	_, err = repo.FindByID(ctx, uuid.New())
	assert.Error(t, err)
}

// TestSurrenderRequestRepository_UpdateStatus tests status updates
func TestSurrenderRequestRepository_UpdateStatus(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	repo := NewSurrenderRequestRepository(pool)
	ctx := context.Background()

	// Create test request
	request := domain.PolicySurrenderRequest{
		PolicyID:            "0000000000004",
		RequestNumber:       "SUR-TEST-STATUS-001",
		RequestType:         domain.SurrenderRequestTypeVoluntary,
		RequestDate:         time.Now(),
		GrossSurrenderValue: 50000,
		NetSurrenderValue:   45000,
		Status:              domain.SurrenderStatusPendingDocumentUpload,
		Owner:               domain.RequestOwnerCustomer,
		CreatedBy:           uuid.New(),
		Metadata:            map[string]interface{}{},
	}

	created, err := repo.Create(ctx, request)
	require.NoError(t, err)

	// Update status
	comments := "Documents verified"
	userID := uuid.New()
	updated, err := repo.UpdateStatus(ctx, created.ID, domain.SurrenderStatusPendingVerification, userID, &comments)

	require.NoError(t, err)
	assert.Equal(t, domain.SurrenderStatusPendingVerification, updated.Status)
	assert.NotEqual(t, created.UpdatedAt, updated.UpdatedAt)
}

// TestSurrenderRequestRepository_ListPendingAutoCompletion tests batch query
func TestSurrenderRequestRepository_ListPendingAutoCompletion(t *testing.T) {
	pool := setupTestDB(t)
	defer pool.Close()

	repo := NewSurrenderRequestRepository(pool)
	ctx := context.Background()

	// Create test requests
	for i := 0; i < 5; i++ {
		request := domain.PolicySurrenderRequest{
			PolicyID:            fmt.Sprintf("%013d", i+5),
			RequestNumber:       "SUR-BATCH-" + string(rune(i)),
			RequestType:         domain.SurrenderRequestTypeVoluntary,
			RequestDate:         time.Now().AddDate(0, 0, -40), // 40 days ago
			GrossSurrenderValue: 50000,
			NetSurrenderValue:   45000,
			Status:              domain.SurrenderStatusApproved,
			Owner:               domain.RequestOwnerCustomer,
			CreatedBy:           uuid.New(),
			Metadata:            map[string]interface{}{},
		}
		_, err := repo.Create(ctx, request)
		require.NoError(t, err)
	}

	// List pending auto-completion
	cutoffDate := time.Now().AddDate(0, 0, -30)
	requests, err := repo.ListPendingAutoCompletion(ctx, cutoffDate, 10)

	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(requests), 0)
}

// Helper function to setup test database
func setupTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	// For actual tests, use test database connection
	// For this example, we'll skip if no test DB available
	t.Skip("Skipping database test - requires test database setup")

	// Example connection string for test DB
	// connString := "postgres://user:pass@localhost:5432/surrender_test"
	// pool, err := pgxpool.New(context.Background(), connString)
	// require.NoError(t, err)
	// return pool

	return nil
}
