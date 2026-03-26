package domain

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ─── MobileChangeDetail ──────────────────────────────────────────────────────

// TestMobileChangeDetail_FieldsAndJSONTags verifies struct fields and JSON serialization.
func TestMobileChangeDetail_FieldsAndJSONTags(t *testing.T) {
	oldMobile := "9876543210"
	txnID := "aadhaar-txn-001"
	updatedBy := "system"

	detail := MobileChangeDetail{
		DetailID:        "detail-mob-001",
		RequestID:       "req-mob-001",
		OldMobileNumber: &oldMobile,
		NewMobileNumber: "1234567890",
		AadhaarTxnID:    &txnID,
		CreatedAt:       time.Date(2026, 3, 25, 10, 0, 0, 0, time.UTC),
		UpdatedAt:       time.Date(2026, 3, 25, 10, 5, 0, 0, time.UTC),
		CreatedBy:       "customer-42",
		UpdatedBy:       &updatedBy,
		Version:         1,
	}

	assert.Equal(t, "detail-mob-001", detail.DetailID)
	assert.Equal(t, "req-mob-001", detail.RequestID)
	require.NotNil(t, detail.OldMobileNumber)
	assert.Equal(t, "9876543210", *detail.OldMobileNumber)
	assert.Equal(t, "1234567890", detail.NewMobileNumber)
	require.NotNil(t, detail.AadhaarTxnID)
	assert.Equal(t, "aadhaar-txn-001", *detail.AadhaarTxnID)
	assert.Equal(t, "customer-42", detail.CreatedBy)
	require.NotNil(t, detail.UpdatedBy)
	assert.Equal(t, "system", *detail.UpdatedBy)
	assert.Equal(t, 1, detail.Version)

	// Verify JSON round-trip preserves tags.
	data, err := json.Marshal(detail)
	require.NoError(t, err)

	var decoded MobileChangeDetail
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, detail.DetailID, decoded.DetailID)
	assert.Equal(t, detail.NewMobileNumber, decoded.NewMobileNumber)
	assert.Equal(t, detail.Version, decoded.Version)
}

// ─── EmailChangeDetail ───────────────────────────────────────────────────────

// TestEmailChangeDetail_FieldsAndJSONTags verifies struct fields and JSON serialization.
func TestEmailChangeDetail_FieldsAndJSONTags(t *testing.T) {
	oldEmail := "old@example.com"
	updatedBy := "system"

	detail := EmailChangeDetail{
		DetailID:  "detail-email-001",
		RequestID: "req-email-001",
		OldEmail:  &oldEmail,
		NewEmail:  "new@example.com",
		CreatedAt: time.Date(2026, 3, 25, 11, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 3, 25, 11, 5, 0, 0, time.UTC),
		CreatedBy: "customer-77",
		UpdatedBy: &updatedBy,
		Version:   2,
	}

	assert.Equal(t, "detail-email-001", detail.DetailID)
	assert.Equal(t, "req-email-001", detail.RequestID)
	require.NotNil(t, detail.OldEmail)
	assert.Equal(t, "old@example.com", *detail.OldEmail)
	assert.Equal(t, "new@example.com", detail.NewEmail)
	assert.Equal(t, "customer-77", detail.CreatedBy)
	require.NotNil(t, detail.UpdatedBy)
	assert.Equal(t, "system", *detail.UpdatedBy)
	assert.Equal(t, 2, detail.Version)

	// Verify JSON round-trip preserves tags.
	data, err := json.Marshal(detail)
	require.NoError(t, err)

	var decoded EmailChangeDetail
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, detail.DetailID, decoded.DetailID)
	assert.Equal(t, detail.NewEmail, decoded.NewEmail)
	assert.Equal(t, detail.Version, decoded.Version)
}

// ─── MobileVersionHistory ────────────────────────────────────────────────────

// TestMobileVersionHistory_CustomerIDIsInt64 verifies CustomerID can hold large int64 values.
func TestMobileVersionHistory_CustomerIDIsInt64(t *testing.T) {
	var largeID int64 = 9999999999
	now := time.Now().UTC()

	vh := MobileVersionHistory{
		VersionID:     "vh-mob-001",
		CustomerID:    largeID,
		MobileNumber:  "9876543210",
		VersionNumber: 1,
		IsActive:      true,
		EffectiveFrom: now,
		CreatedAt:     now,
		CreatedBy:     "system",
	}

	assert.Equal(t, int64(9999999999), vh.CustomerID)
	assert.Equal(t, "9876543210", vh.MobileNumber)
	assert.True(t, vh.IsActive)

	// Verify JSON round-trip preserves large CustomerID.
	data, err := json.Marshal(vh)
	require.NoError(t, err)

	var decoded MobileVersionHistory
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, int64(9999999999), decoded.CustomerID)
}

// ─── EmailVersionHistory ─────────────────────────────────────────────────────

// TestEmailVersionHistory_CustomerIDIsInt64 verifies CustomerID can hold large int64 values.
func TestEmailVersionHistory_CustomerIDIsInt64(t *testing.T) {
	var largeID int64 = 9999999999
	now := time.Now().UTC()

	vh := EmailVersionHistory{
		VersionID:     "vh-email-001",
		CustomerID:    largeID,
		Email:         "test@example.com",
		VersionNumber: 1,
		IsActive:      true,
		EffectiveFrom: now,
		CreatedAt:     now,
		CreatedBy:     "system",
	}

	assert.Equal(t, int64(9999999999), vh.CustomerID)
	assert.Equal(t, "test@example.com", vh.Email)
	assert.True(t, vh.IsActive)

	// Verify JSON round-trip preserves large CustomerID.
	data, err := json.Marshal(vh)
	require.NoError(t, err)

	var decoded EmailVersionHistory
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)

	assert.Equal(t, int64(9999999999), decoded.CustomerID)
}
