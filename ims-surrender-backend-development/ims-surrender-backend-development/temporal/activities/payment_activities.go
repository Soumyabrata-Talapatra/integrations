package activities

import (
	"context"
)

// Activities for Payment Workflow (TEMP-005)

type ValidatePaymentEligibilityInput struct {
	SurrenderRequestID string
}

type ValidatePaymentEligibilityResult struct {
	Eligible bool
	Reason   string
}

func ValidatePaymentEligibilityActivity(ctx context.Context, input ValidatePaymentEligibilityInput) (*ValidatePaymentEligibilityResult, error) {
	// Placeholder - would validate payment eligibility
	return &ValidatePaymentEligibilityResult{
		Eligible: true,
		Reason:   "",
	}, nil
}

type DetermineDispositionInput struct {
	SurrenderRequestID string
	NetSurrenderValue  float64
}

type DetermineDispositionResult struct {
	DispositionType string
	NewPolicyStatus string
	NewSumAssured   float64
}

func DetermineDispositionActivity(ctx context.Context, input DetermineDispositionInput) (*DetermineDispositionResult, error) {
	// Placeholder - would determine disposition based on prescribed limit
	prescribedLimit := 2000.0

	if input.NetSurrenderValue >= prescribedLimit {
		return &DetermineDispositionResult{
			DispositionType: "REDUCED_PAID_UP",
			NewPolicyStatus: "AU",
			NewSumAssured:   input.NetSurrenderValue,
		}, nil
	}

	return &DetermineDispositionResult{
		DispositionType: "TERMINATED_SURRENDER",
		NewPolicyStatus: "TS",
		NewSumAssured:   0,
	}, nil
}

type CreateDispositionRecordInput struct {
	SurrenderRequestID string
	PolicyID           string
	DispositionType    string
	PaymentReference   string
	NetAmount          float64
	NewPolicyStatus    string
}

type CreateDispositionRecordResult struct {
	DispositionID string
}

func CreateDispositionRecordActivity(ctx context.Context, input CreateDispositionRecordInput) (*CreateDispositionRecordResult, error) {
	// Placeholder - would create disposition record
	return &CreateDispositionRecordResult{
		DispositionID: "disp-123",
	}, nil
}

type SendPaymentNotificationInput struct {
	SurrenderRequestID string
	PolicyID           string
	PaymentReference   string
	Amount             float64
	DispositionType    string
}

type SendPaymentNotificationResult struct {
	ChannelsSent []string
}

func SendPaymentNotificationActivity(ctx context.Context, input SendPaymentNotificationInput) (*SendPaymentNotificationResult, error) {
	// Placeholder - would send notifications
	return &SendPaymentNotificationResult{
		ChannelsSent: []string{"EMAIL", "SMS"},
	}, nil
}
