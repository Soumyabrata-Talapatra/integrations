package activities

import (
	"context"
)

// Activities for Document Verification Workflow (TEMP-003)

type ValidateRequiredDocumentsInput struct {
	SurrenderRequestID string
}

type ValidateRequiredDocumentsResult struct {
	AllUploaded   bool
	UploadedCount int
	RequiredCount int
}

func ValidateRequiredDocumentsActivity(ctx context.Context, input ValidateRequiredDocumentsInput) (*ValidateRequiredDocumentsResult, error) {
	// Placeholder - would check document repository
	return &ValidateRequiredDocumentsResult{
		AllUploaded:   true,
		UploadedCount: 3,
		RequiredCount: 3,
	}, nil
}

type ExtractDocumentDataInput struct {
	DocumentID string
}

type ExtractDocumentDataResult struct {
	ExtractedData   map[string]interface{}
	ConfidenceScore float64
}

func ExtractDocumentDataActivity(ctx context.Context, input ExtractDocumentDataInput) (*ExtractDocumentDataResult, error) {
	// Placeholder - would use OCR service
	return &ExtractDocumentDataResult{
		ExtractedData: map[string]interface{}{
			"policy_number": "PLI/2020/123456",
			"name":          "John Doe",
		},
		ConfidenceScore: 0.95,
	}, nil
}

type AutoVerifyDocumentInput struct {
	DocumentID      string
	ExtractedData   map[string]interface{}
	ConfidenceScore float64
}

type AutoVerifyDocumentResult struct {
	Verified             bool
	RequiresManualReview bool
	Reason               string
}

func AutoVerifyDocumentActivity(ctx context.Context, input AutoVerifyDocumentInput) (*AutoVerifyDocumentResult, error) {
	// Placeholder - would apply verification rules
	if input.ConfidenceScore >= 0.90 {
		return &AutoVerifyDocumentResult{
			Verified:             true,
			RequiresManualReview: false,
		}, nil
	}

	return &AutoVerifyDocumentResult{
		Verified:             false,
		RequiresManualReview: true,
		Reason:               "Low confidence score",
	}, nil
}

type RouteToManualVerificationInput struct {
	DocumentID string
	Reason     string
}

type RouteToManualVerificationResult struct {
	TaskID string
}

func RouteToManualVerificationActivity(ctx context.Context, input RouteToManualVerificationInput) (*RouteToManualVerificationResult, error) {
	// Placeholder - would create verification task
	return &RouteToManualVerificationResult{
		TaskID: "verify-task-123",
	}, nil
}

type UpdateSurrenderStatusInput struct {
	SurrenderRequestID string
	NewStatus          string
}

type UpdateSurrenderStatusResult struct {
	Success bool
}

func UpdateSurrenderStatusActivity(ctx context.Context, input UpdateSurrenderStatusInput) (*UpdateSurrenderStatusResult, error) {
	// Placeholder - would update surrender repository
	return &UpdateSurrenderStatusResult{
		Success: true,
	}, nil
}
