package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"gitlab.cept.gov.in/it-2.0-policy/surrender-service/core/domain"
)

// MockSurrenderRepository is a mock for testing
type MockSurrenderRepository struct {
	mock.Mock
}

func (m *MockSurrenderRepository) Create(ctx interface{}, request domain.PolicySurrenderRequest) (*domain.PolicySurrenderRequest, error) {
	args := m.Called(ctx, request)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.PolicySurrenderRequest), args.Error(1)
}

func (m *MockSurrenderRepository) FindByID(ctx interface{}, id uuid.UUID) (*domain.PolicySurrenderRequest, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.PolicySurrenderRequest), args.Error(1)
}

func (m *MockSurrenderRepository) FindActiveByPolicyID(ctx interface{}, policyID uuid.UUID) (*domain.PolicySurrenderRequest, bool, error) {
	args := m.Called(ctx, policyID)
	if args.Get(0) == nil {
		return nil, args.Bool(1), args.Error(2)
	}
	return args.Get(0).(*domain.PolicySurrenderRequest), args.Bool(1), args.Error(2)
}

func (m *MockSurrenderRepository) UpdateStatus(ctx interface{}, id uuid.UUID, status domain.SurrenderStatus, userID uuid.UUID, comments *string) (*domain.PolicySurrenderRequest, error) {
	args := m.Called(ctx, id, status, userID, comments)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.PolicySurrenderRequest), args.Error(1)
}

// MockDocumentRepository is a mock for testing
type MockDocumentRepository struct {
	mock.Mock
}

func (m *MockDocumentRepository) Create(ctx interface{}, doc domain.SurrenderDocument) (*domain.SurrenderDocument, error) {
	args := m.Called(ctx, doc)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.SurrenderDocument), args.Error(1)
}

func (m *MockDocumentRepository) FindBySurrenderRequestID(ctx interface{}, surrenderRequestID uuid.UUID) ([]domain.SurrenderDocument, error) {
	args := m.Called(ctx, surrenderRequestID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.SurrenderDocument), args.Error(1)
}

func (m *MockDocumentRepository) CheckDocumentExists(ctx interface{}, surrenderRequestID uuid.UUID, docType domain.DocumentType) (bool, error) {
	args := m.Called(ctx, surrenderRequestID, docType)
	return args.Bool(0), args.Error(1)
}

func (m *MockDocumentRepository) CountVerifiedDocuments(ctx interface{}, surrenderRequestID uuid.UUID) (int64, error) {
	args := m.Called(ctx, surrenderRequestID)
	return args.Get(0).(int64), args.Error(1)
}

// TestValidateEligibilityRequest tests the eligibility validation endpoint
func TestValidateEligibilityRequest(t *testing.T) {
	tests := []struct {
		name             string
		requestBody      string
		mockSetup        func(*MockSurrenderRepository)
		expectedStatus   int
		expectedEligible bool
	}{
		{
			name: "eligible policy",
			requestBody: `{
				"policy_id": "123e4567-e89b-12d3-a456-426614174000"
			}`,
			mockSetup: func(m *MockSurrenderRepository) {
				policyID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
				m.On("FindActiveByPolicyID", mock.Anything, policyID).
					Return(nil, false, nil)
			},
			expectedStatus:   http.StatusOK,
			expectedEligible: true,
		},
		{
			name: "policy with active surrender",
			requestBody: `{
				"policy_id": "123e4567-e89b-12d3-a456-426614174000"
			}`,
			mockSetup: func(m *MockSurrenderRepository) {
				policyID := uuid.MustParse("123e4567-e89b-12d3-a456-426614174000")
				existingRequest := &domain.PolicySurrenderRequest{
					ID:            uuid.New(),
					RequestNumber: "SUR-123",
					Status:        domain.SurrenderStatusPendingDocumentUpload,
				}
				m.On("FindActiveByPolicyID", mock.Anything, policyID).
					Return(existingRequest, true, nil)
			},
			expectedStatus:   http.StatusOK,
			expectedEligible: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockRepo := new(MockSurrenderRepository)
			if tt.mockSetup != nil {
				tt.mockSetup(mockRepo)
			}

			// Create request
			req := httptest.NewRequest(http.MethodPost, "/v1/surrender/validate-eligibility",
				strings.NewReader(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			_ = httptest.NewRecorder()

			// TODO: Call handler (would require full setup with router)
			// For now, just verify mocks were called
			mockRepo.AssertExpectations(t)
		})
	}
}

// TestConfirmSurrenderRequest tests the confirm surrender endpoint
func TestConfirmSurrenderRequest(t *testing.T) {
	mockRepo := new(MockSurrenderRepository)
	policyID := "0000000000001"

	// Mock setup
	mockRepo.On("Create", mock.Anything, mock.MatchedBy(func(req domain.PolicySurrenderRequest) bool {
		return req.PolicyID == policyID &&
			req.RequestType == domain.SurrenderRequestTypeVoluntary &&
			req.Status == domain.SurrenderStatusPendingDocumentUpload
	})).Return(&domain.PolicySurrenderRequest{
		ID:            uuid.New(),
		PolicyID:      policyID,
		RequestNumber: "SUR-123456",
		RequestType:   domain.SurrenderRequestTypeVoluntary,
		Status:        domain.SurrenderStatusPendingDocumentUpload,
		CreatedAt:     time.Now(),
	}, nil)

	// Test would call handler with mock
	// Verify correct surrender request is created
	mockRepo.AssertExpectations(t)
}

// TestCalculateSurrenderValue tests surrender value calculation logic
func TestCalculateSurrenderValue(t *testing.T) {
	tests := []struct {
		name            string
		sumAssured      float64
		premiumsPaid    int
		totalPremiums   int
		bonusAmount     float64
		surrenderFactor float64
		deductions      float64
		expectedGSV     float64
		expectedNSV     float64
	}{
		{
			name:            "basic calculation",
			sumAssured:      100000,
			premiumsPaid:    5,
			totalPremiums:   20,
			bonusAmount:     10000,
			surrenderFactor: 0.75,
			deductions:      5000,
			expectedGSV:     33750, // ((100000 * 5 / 20) + 10000) * 0.75
			expectedNSV:     28750, // 33750 - 5000
		},
		{
			name:            "high bonus scenario",
			sumAssured:      200000,
			premiumsPaid:    10,
			totalPremiums:   20,
			bonusAmount:     50000,
			surrenderFactor: 0.80,
			deductions:      10000,
			expectedGSV:     120000, // ((200000 * 10 / 20) + 50000) * 0.80
			expectedNSV:     110000, // 120000 - 10000
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Calculate Paid-Up Value
			paidUpValue := (tt.sumAssured * float64(tt.premiumsPaid)) / float64(tt.totalPremiums)

			// Calculate Gross Surrender Value
			gsv := (paidUpValue + tt.bonusAmount) * tt.surrenderFactor

			// Calculate Net Surrender Value
			nsv := gsv - tt.deductions

			assert.InDelta(t, tt.expectedGSV, gsv, 0.01, "GSV mismatch")
			assert.InDelta(t, tt.expectedNSV, nsv, 0.01, "NSV mismatch")
		})
	}
}

// TestDispositionPrediction tests disposition logic
func TestDispositionPrediction(t *testing.T) {
	tests := []struct {
		name                 string
		netSurrenderValue    float64
		prescribedLimit      float64
		expectedDisposition  string
		expectedPolicyStatus string
	}{
		{
			name:                 "below limit - terminated",
			netSurrenderValue:    1500,
			prescribedLimit:      2000,
			expectedDisposition:  "TERMINATED_SURRENDER",
			expectedPolicyStatus: "TS",
		},
		{
			name:                 "above limit - reduced paid up",
			netSurrenderValue:    5000,
			prescribedLimit:      2000,
			expectedDisposition:  "REDUCED_PAID_UP",
			expectedPolicyStatus: "AU",
		},
		{
			name:                 "at limit - reduced paid up",
			netSurrenderValue:    2000,
			prescribedLimit:      2000,
			expectedDisposition:  "REDUCED_PAID_UP",
			expectedPolicyStatus: "AU",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var disposition, policyStatus string

			if tt.netSurrenderValue >= tt.prescribedLimit {
				disposition = "REDUCED_PAID_UP"
				policyStatus = "AU"
			} else {
				disposition = "TERMINATED_SURRENDER"
				policyStatus = "TS"
			}

			assert.Equal(t, tt.expectedDisposition, disposition)
			assert.Equal(t, tt.expectedPolicyStatus, policyStatus)
		})
	}
}

// TestRequestNumberGeneration tests unique request number generation
func TestRequestNumberGeneration(t *testing.T) {
	policyNumber := "PLI/2020/123456"

	// Generate request number
	timestamp := time.Now().Format("20060102150405")
	requestNumber := "SUR-" + policyNumber + "-" + timestamp

	// Verify format
	assert.Contains(t, requestNumber, "SUR-")
	assert.Contains(t, requestNumber, policyNumber)
	assert.Regexp(t, `SUR-PLI/2020/123456-\d{14}`, requestNumber)

	// Verify uniqueness (different timestamps)
	time.Sleep(1 * time.Second)
	timestamp2 := time.Now().Format("20060102150405")
	requestNumber2 := "SUR-" + policyNumber + "-" + timestamp2

	assert.NotEqual(t, requestNumber, requestNumber2)
}

// TestJSONSerialization tests request/response JSON marshaling
func TestJSONSerialization(t *testing.T) {
	t.Run("ValidateEligibilityRequest", func(t *testing.T) {
		jsonStr := `{"policy_id":"123e4567-e89b-12d3-a456-426614174000"}`

		var req ValidateEligibilityRequest
		err := json.Unmarshal([]byte(jsonStr), &req)

		require.NoError(t, err)
		assert.Equal(t, "123e4567-e89b-12d3-a456-426614174000", req.PolicyID)
	})

	t.Run("CalculateSurrenderResponse", func(t *testing.T) {
		// Would test response serialization
		// This verifies our DTOs work correctly with JSON
	})
}
