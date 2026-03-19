package activities

import (
	"context"
)

// Activities for Policy Update Workflow (TEMP-006)

type ValidateStatusTransitionInput struct {
	PolicyID  string
	NewStatus string
}

type ValidateStatusTransitionResult struct {
	Valid         bool
	CurrentStatus string
	Reason        string
}

func ValidateStatusTransitionActivity(ctx context.Context, input ValidateStatusTransitionInput) (*ValidateStatusTransitionResult, error) {
	// Placeholder - would validate status transition
	return &ValidateStatusTransitionResult{
		Valid:         true,
		CurrentStatus: "AP",
		Reason:        "",
	}, nil
}

type CreatePolicyHistoryInput struct {
	PolicyID           string
	SurrenderRequestID string
	OldStatus          string
	NewStatus          string
	ChangeReason       string
}

type CreatePolicyHistoryResult struct {
	HistoryID string
}

func CreatePolicyHistoryActivity(ctx context.Context, input CreatePolicyHistoryInput) (*CreatePolicyHistoryResult, error) {
	// Placeholder - would create policy history record
	return &CreatePolicyHistoryResult{
		HistoryID: "hist-123",
	}, nil
}

type SettlePolicyLoansInput struct {
	PolicyID           string
	SurrenderRequestID string
}

type SettlePolicyLoansResult struct {
	LoansSettled int
	TotalAmount  float64
}

func SettlePolicyLoansActivity(ctx context.Context, input SettlePolicyLoansInput) (*SettlePolicyLoansResult, error) {
	// Placeholder - would settle loans
	return &SettlePolicyLoansResult{
		LoansSettled: 0,
		TotalAmount:  0,
	}, nil
}

type StopFutureBonusesInput struct {
	PolicyID string
}

type StopFutureBonusesResult struct {
	Success bool
}

func StopFutureBonusesActivity(ctx context.Context, input StopFutureBonusesInput) (*StopFutureBonusesResult, error) {
	// Placeholder - would stop future bonuses
	return &StopFutureBonusesResult{
		Success: true,
	}, nil
}

type UpdateReducedPaidUpDetailsInput struct {
	PolicyID           string
	SurrenderRequestID string
}

type UpdateReducedPaidUpDetailsResult struct {
	NewSumAssured float64
	Success       bool
}

func UpdateReducedPaidUpDetailsActivity(ctx context.Context, input UpdateReducedPaidUpDetailsInput) (*UpdateReducedPaidUpDetailsResult, error) {
	// Placeholder - would update reduced paid-up details
	return &UpdateReducedPaidUpDetailsResult{
		NewSumAssured: 50000,
		Success:       true,
	}, nil
}

type SendPolicyUpdateNotificationInput struct {
	PolicyID        string
	OldStatus       string
	NewStatus       string
	DispositionType string
}

type SendPolicyUpdateNotificationResult struct {
	ChannelsSent []string
}

func SendPolicyUpdateNotificationActivity(ctx context.Context, input SendPolicyUpdateNotificationInput) (*SendPolicyUpdateNotificationResult, error) {
	// Placeholder - would send notifications
	return &SendPolicyUpdateNotificationResult{
		ChannelsSent: []string{"EMAIL", "SMS", "LETTER"},
	}, nil
}

type ArchivePolicyRecordsInput struct {
	PolicyID string
}

type ArchivePolicyRecordsResult struct {
	RecordsArchived int
}

func ArchivePolicyRecordsActivity(ctx context.Context, input ArchivePolicyRecordsInput) (*ArchivePolicyRecordsResult, error) {
	// Placeholder - would archive policy records
	return &ArchivePolicyRecordsResult{
		RecordsArchived: 0,
	}, nil
}
