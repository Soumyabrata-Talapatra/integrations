package workflows

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"gitlab.cept.gov.in/it-2.0-policy/surrender-service/temporal/activities"
)

// DocumentVerificationWorkflowInput defines the input for document verification workflow
// TEMP-003: Document Verification Workflow
type DocumentVerificationWorkflowInput struct {
	SurrenderRequestID string
	DocumentIDs        []string
}

// DocumentVerificationWorkflowResult defines the result of document verification workflow
type DocumentVerificationWorkflowResult struct {
	SurrenderRequestID string
	TotalDocuments     int
	VerifiedDocuments  int
	RejectedDocuments  int
	PendingDocuments   int
	AllVerified        bool
	CompletedAt        time.Time
}

// DocumentVerificationWorkflow orchestrates document verification process
// TEMP-003: Document Verification Workflow
// Business Flow:
// 1. Validate all required documents uploaded
// 2. Perform OCR/extraction (if needed)
// 3. Auto-verify documents where possible
// 4. Route remaining to manual verification
// 5. Track verification status
// 6. Update surrender request status
func DocumentVerificationWorkflow(ctx workflow.Context, input DocumentVerificationWorkflowInput) (*DocumentVerificationWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting Document Verification Workflow", "SurrenderRequestID", input.SurrenderRequestID)

	result := &DocumentVerificationWorkflowResult{
		SurrenderRequestID: input.SurrenderRequestID,
		TotalDocuments:     len(input.DocumentIDs),
	}

	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	// Step 1: Validate Required Documents
	logger.Info("Step 1: Validating required documents")
	var validateResult activities.ValidateRequiredDocumentsResult
	err := workflow.ExecuteActivity(ctx, activities.ValidateRequiredDocumentsActivity, activities.ValidateRequiredDocumentsInput{
		SurrenderRequestID: input.SurrenderRequestID,
	}).Get(ctx, &validateResult)

	if err != nil {
		logger.Error("Failed to validate required documents", "error", err)
		return result, err
	}

	if !validateResult.AllUploaded {
		logger.Warn("Not all required documents uploaded", "uploaded", validateResult.UploadedCount, "required", validateResult.RequiredCount)
		result.PendingDocuments = validateResult.RequiredCount - validateResult.UploadedCount
		result.CompletedAt = workflow.Now(ctx)
		return result, nil
	}

	// Step 2: Process Each Document
	logger.Info("Step 2: Processing documents", "count", len(input.DocumentIDs))

	verifiedCount := 0
	rejectedCount := 0
	pendingCount := 0

	for _, documentID := range input.DocumentIDs {
		logger.Info("Processing document", "DocumentID", documentID)

		// Step 2a: Extract Document Data (OCR if needed)
		var extractResult activities.ExtractDocumentDataResult
		err := workflow.ExecuteActivity(ctx, activities.ExtractDocumentDataActivity, activities.ExtractDocumentDataInput{
			DocumentID: documentID,
		}).Get(ctx, &extractResult)

		if err != nil {
			logger.Error("Failed to extract document data", "error", err, "DocumentID", documentID)
			rejectedCount++
			continue
		}

		// Step 2b: Auto-Verify Document
		var autoVerifyResult activities.AutoVerifyDocumentResult
		err = workflow.ExecuteActivity(ctx, activities.AutoVerifyDocumentActivity, activities.AutoVerifyDocumentInput{
			DocumentID:      documentID,
			ExtractedData:   extractResult.ExtractedData,
			ConfidenceScore: extractResult.ConfidenceScore,
		}).Get(ctx, &autoVerifyResult)

		if err != nil {
			logger.Error("Failed to auto-verify document", "error", err, "DocumentID", documentID)
			pendingCount++
			continue
		}

		if autoVerifyResult.Verified {
			logger.Info("Document auto-verified", "DocumentID", documentID)
			verifiedCount++
		} else if autoVerifyResult.RequiresManualReview {
			logger.Info("Document requires manual review", "DocumentID", documentID, "reason", autoVerifyResult.Reason)
			pendingCount++

			// Step 2c: Route to Manual Verification
			var manualVerifyResult activities.RouteToManualVerificationResult
			err = workflow.ExecuteActivity(ctx, activities.RouteToManualVerificationActivity, activities.RouteToManualVerificationInput{
				DocumentID: documentID,
				Reason:     autoVerifyResult.Reason,
			}).Get(ctx, &manualVerifyResult)

			if err != nil {
				logger.Error("Failed to route to manual verification", "error", err, "DocumentID", documentID)
			}
		} else {
			logger.Warn("Document rejected", "DocumentID", documentID, "reason", autoVerifyResult.Reason)
			rejectedCount++
		}
	}

	result.VerifiedDocuments = verifiedCount
	result.RejectedDocuments = rejectedCount
	result.PendingDocuments = pendingCount
	result.AllVerified = (verifiedCount == len(input.DocumentIDs))

	logger.Info("Document processing summary",
		"verified", verifiedCount,
		"rejected", rejectedCount,
		"pending", pendingCount)

	// Step 3: Update Surrender Request Status
	logger.Info("Step 3: Updating surrender request status")

	var newStatus string
	if result.AllVerified {
		newStatus = "PENDING_APPROVAL"
	} else if rejectedCount > 0 {
		newStatus = "DOCUMENT_REJECTED"
	} else if pendingCount > 0 {
		newStatus = "PENDING_DOCUMENT_VERIFICATION"
	}

	var updateResult activities.UpdateSurrenderStatusResult
	err = workflow.ExecuteActivity(ctx, activities.UpdateSurrenderStatusActivity, activities.UpdateSurrenderStatusInput{
		SurrenderRequestID: input.SurrenderRequestID,
		NewStatus:          newStatus,
	}).Get(ctx, &updateResult)

	if err != nil {
		logger.Error("Failed to update surrender status", "error", err)
		// Don't fail workflow
	} else {
		logger.Info("Surrender status updated", "new_status", newStatus)
	}

	// If all verified, route to approval
	if result.AllVerified {
		logger.Info("All documents verified, routing to approval")

		var approvalResult activities.RouteToApprovalResult
		err = workflow.ExecuteActivity(ctx, activities.RouteToApprovalActivity, activities.RouteToApprovalInput{
			SurrenderRequestID: input.SurrenderRequestID,
			Priority:           "NORMAL",
		}).Get(ctx, &approvalResult)

		if err != nil {
			logger.Error("Failed to route to approval", "error", err)
		} else {
			logger.Info("Routed to approval queue", "TaskID", approvalResult.TaskID)
		}
	}

	logger.Info("Document Verification Workflow completed")
	result.CompletedAt = workflow.Now(ctx)

	return result, nil
}
