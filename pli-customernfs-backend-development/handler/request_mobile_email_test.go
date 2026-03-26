package handler_test

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	handler "customer-nfs-service/handler"
)

// ─── InitiateMobileChangeRequest ─────────────────────────────────────────────

// TestInitiateMobileChangeRequest_CustomerIDValidationTag verifies the validate tag for CustomerID.
func TestInitiateMobileChangeRequest_CustomerIDValidationTag(t *testing.T) {
	rt := reflect.TypeOf(handler.InitiateMobileChangeRequest{})
	field, ok := rt.FieldByName("CustomerID")
	require.True(t, ok, "CustomerID field must exist")

	tag := field.Tag.Get("validate")
	assert.Contains(t, tag, "required,gt=0", "CustomerID must have required,gt=0 validation")
}

// TestInitiateMobileChangeRequest_NewMobileNumberValidationTag verifies the validate tag for NewMobileNumber.
func TestInitiateMobileChangeRequest_NewMobileNumberValidationTag(t *testing.T) {
	rt := reflect.TypeOf(handler.InitiateMobileChangeRequest{})
	field, ok := rt.FieldByName("NewMobileNumber")
	require.True(t, ok, "NewMobileNumber field must exist")

	tag := field.Tag.Get("validate")
	assert.Contains(t, tag, "required,len=10,numeric", "NewMobileNumber must have required,len=10,numeric validation")
}

// TestInitiateMobileChangeRequest_ChannelValidationTag verifies the validate tag for Channel.
func TestInitiateMobileChangeRequest_ChannelValidationTag(t *testing.T) {
	rt := reflect.TypeOf(handler.InitiateMobileChangeRequest{})
	field, ok := rt.FieldByName("Channel")
	require.True(t, ok, "Channel field must exist")

	tag := field.Tag.Get("validate")
	assert.Contains(t, tag, "required", "Channel must have required validation")
}

// ─── InitiateEmailChangeRequest ──────────────────────────────────────────────

// TestInitiateEmailChangeRequest_NewEmailValidationTag verifies the validate tag for NewEmail.
func TestInitiateEmailChangeRequest_NewEmailValidationTag(t *testing.T) {
	rt := reflect.TypeOf(handler.InitiateEmailChangeRequest{})
	field, ok := rt.FieldByName("NewEmail")
	require.True(t, ok, "NewEmail field must exist")

	tag := field.Tag.Get("validate")
	assert.Contains(t, tag, "required,email,max=255", "NewEmail must have required,email,max=255 validation")
}

// TestInitiateEmailChangeRequest_CustomerIDValidationTag verifies the validate tag for CustomerID.
func TestInitiateEmailChangeRequest_CustomerIDValidationTag(t *testing.T) {
	rt := reflect.TypeOf(handler.InitiateEmailChangeRequest{})
	field, ok := rt.FieldByName("CustomerID")
	require.True(t, ok, "CustomerID field must exist")

	tag := field.Tag.Get("validate")
	assert.Contains(t, tag, "required,gt=0", "CustomerID must have required,gt=0 validation")
}

// ─── VerifyMobileOTPRequest ──────────────────────────────────────────────────

// TestVerifyMobileOTPRequest_Fields verifies field mapping of VerifyMobileOTPRequest.
func TestVerifyMobileOTPRequest_Fields(t *testing.T) {
	req := handler.VerifyMobileOTPRequest{
		RequestID: "req-mob-otp-001",
		OTPCode:   "123456",
		TxnID:     "txn-abc",
	}

	assert.Equal(t, "req-mob-otp-001", req.RequestID)
	assert.Equal(t, "123456", req.OTPCode)
	assert.Equal(t, "txn-abc", req.TxnID)

	// Verify struct tags exist.
	rt := reflect.TypeOf(handler.VerifyMobileOTPRequest{})

	ridField, ok := rt.FieldByName("RequestID")
	require.True(t, ok)
	assert.Contains(t, ridField.Tag.Get("validate"), "required")

	otpField, ok := rt.FieldByName("OTPCode")
	require.True(t, ok)
	assert.Contains(t, otpField.Tag.Get("validate"), "required,len=6,numeric")

	txnField, ok := rt.FieldByName("TxnID")
	require.True(t, ok)
	assert.Contains(t, txnField.Tag.Get("validate"), "required")
}

// ─── VerifyEmailOTPRequest ───────────────────────────────────────────────────

// TestVerifyEmailOTPRequest_Fields verifies field mapping of VerifyEmailOTPRequest.
func TestVerifyEmailOTPRequest_Fields(t *testing.T) {
	req := handler.VerifyEmailOTPRequest{
		RequestID: "req-email-otp-001",
		OTPCode:   "654321",
		TxnID:     "txn-xyz",
	}

	assert.Equal(t, "req-email-otp-001", req.RequestID)
	assert.Equal(t, "654321", req.OTPCode)
	assert.Equal(t, "txn-xyz", req.TxnID)

	// Verify struct tags exist.
	rt := reflect.TypeOf(handler.VerifyEmailOTPRequest{})

	ridField, ok := rt.FieldByName("RequestID")
	require.True(t, ok)
	assert.Contains(t, ridField.Tag.Get("validate"), "required")

	otpField, ok := rt.FieldByName("OTPCode")
	require.True(t, ok)
	assert.Contains(t, otpField.Tag.Get("validate"), "required,len=6,numeric")

	txnField, ok := rt.FieldByName("TxnID")
	require.True(t, ok)
	assert.Contains(t, txnField.Tag.Get("validate"), "required")
}
