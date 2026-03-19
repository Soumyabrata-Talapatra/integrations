package activities

import (
	"context"
)

// Activities for Forced Surrender Workflow (TEMP-002)

type PolicyInfo struct {
	PolicyID        string
	PolicyNumber    string
	UnpaidMonths    int
	UnpaidAmount    float64
	LastPremiumDate string
}

type IdentifyEligiblePoliciesInput struct {
	EvaluationDate  string
	MinUnpaidMonths int
}

type IdentifyEligiblePoliciesResult struct {
	EligiblePolicies []PolicyInfo
}

func IdentifyEligiblePoliciesActivity(ctx context.Context, input IdentifyEligiblePoliciesInput) (*IdentifyEligiblePoliciesResult, error) {
	// Placeholder - would call collections service
	return &IdentifyEligiblePoliciesResult{
		EligiblePolicies: []PolicyInfo{},
	}, nil
}

type CreateRemindersBatchInput struct {
	Policies []PolicyInfo
}

type CreateRemindersBatchResult struct {
	RemindersCreated int
	Errors           []string
}

func CreateRemindersBatchActivity(ctx context.Context, input CreateRemindersBatchInput) (*CreateRemindersBatchResult, error) {
	// Placeholder - would create reminders and send notifications
	return &CreateRemindersBatchResult{
		RemindersCreated: len(input.Policies),
		Errors:           []string{},
	}, nil
}

type PaymentWindowInfo struct {
	PaymentWindowID string
	PolicyID        string
	PolicyNumber    string
	ExpectedAmount  float64
	WindowEnd       string
}

type CheckExpiredPaymentWindowsInput struct {
	AsOfDate string
}

type CheckExpiredPaymentWindowsResult struct {
	ExpiredWindows []PaymentWindowInfo
}

func CheckExpiredPaymentWindowsActivity(ctx context.Context, input CheckExpiredPaymentWindowsInput) (*CheckExpiredPaymentWindowsResult, error) {
	// Placeholder - would query database
	return &CheckExpiredPaymentWindowsResult{
		ExpiredWindows: []PaymentWindowInfo{},
	}, nil
}

type InitiateForcedSurrendersBatchInput struct {
	ExpiredWindows []PaymentWindowInfo
}

type InitiateForcedSurrendersBatchResult struct {
	SurrendersInitiated int
	Errors              []string
}

func InitiateForcedSurrendersBatchActivity(ctx context.Context, input InitiateForcedSurrendersBatchInput) (*InitiateForcedSurrendersBatchResult, error) {
	// Placeholder - would create forced surrender requests
	return &InitiateForcedSurrendersBatchResult{
		SurrendersInitiated: len(input.ExpiredWindows),
		Errors:              []string{},
	}, nil
}
