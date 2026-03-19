package handler

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"policy-issue-service/core/domain"
	"policy-issue-service/core/port"
	resp "policy-issue-service/handler/response"
	repo "policy-issue-service/repo/postgres"

	log "gitlab.cept.gov.in/it-2.0-common/n-api-log"
	serverHandler "gitlab.cept.gov.in/it-2.0-common/n-api-server/handler"
	serverRoute "gitlab.cept.gov.in/it-2.0-common/n-api-server/route"

	"go.temporal.io/sdk/client"
)

// BulkUploadHandler handles bulk proposal upload HTTP endpoints
// Phase 9: Bulk Upload APIs
type BulkUploadHandler struct {
	*serverHandler.Base
	bulkUploadRepo *repo.BulkUploadRepository
	temporalClient client.Client
}

// NewBulkUploadHandler creates a new BulkUploadHandler instance
func NewBulkUploadHandler(bulkUploadRepo *repo.BulkUploadRepository, temporalClient client.Client) *BulkUploadHandler {
	base := serverHandler.New("BulkUpload").SetPrefix("/v1").AddPrefix("")
	return &BulkUploadHandler{
		Base:           base,
		bulkUploadRepo: bulkUploadRepo,
		temporalClient: temporalClient,
	}
}

// Routes returns the routes for the BulkUploadHandler
func (h *BulkUploadHandler) Routes() []serverRoute.Route {
	return []serverRoute.Route{
		serverRoute.POST("/bulk-upload/proposals", h.UploadBulkProposals).Name("Upload Bulk Proposals"),
		serverRoute.GET("/bulk-upload/batches/:batch_id", h.GetBulkUploadStatus).Name("Get Bulk Upload Status"),
	}
}

// UploadBulkProposals accepts a bulk proposal file for processing
// [FR-POL-021] Bulk Proposal Upload
// [BR-POL-029] Combined cheque reconciliation
// Temporal Workflow: Starts WF-PI-003 (BulkProposalUploadWorkflow)
// File Formats: CSV, Excel
// Max Rows: 1000 per batch
func (h *BulkUploadHandler) UploadBulkProposals(sctx *serverRoute.Context, req BulkUploadRequest) (*resp.BulkUploadResponse, error) {
	// Step 1: Validate file type — only CSV and Excel formats accepted
	validBulkMimeTypes := map[string]bool{
		"text/csv":                 true, // .csv
		"application/vnd.ms-excel": true, // .xls
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": true, // .xlsx
	}
	if !validBulkMimeTypes[req.MimeType] {
		return nil, fmt.Errorf("[FR-POL-021] invalid file type: %s. Allowed: CSV (.csv), Excel (.xls, .xlsx)", req.MimeType)
	}

	// Step 2: Validate file extension matches expected formats
	validExtensions := map[string]bool{".csv": true, ".xls": true, ".xlsx": true}
	fileExt := strings.ToLower(filepath.Ext(req.FileName))
	if fileExt == "" || !validExtensions[fileExt] {
		return nil, fmt.Errorf("[FR-POL-021] file must have .csv, .xls, or .xlsx extension, got: %s", req.FileName)
	}

	// Step 3: Validate file size — max 10 MB
	maxBulkFileSize := int64(10 * 1024 * 1024) // 10 MB
	if req.FileSize > maxBulkFileSize {
		return nil, fmt.Errorf("[FR-POL-021] file size %d bytes exceeds maximum allowed size of 10 MB", req.FileSize)
	}

	// Step 4: Validate max rows
	if req.TotalRows > 1000 {
		return nil, fmt.Errorf("[FR-POL-021] maximum 1000 rows per batch, got %d", req.TotalRows)
	}

	// Step 5: Validate combined cheque fields
	if req.PaymentType == "COMBINED_CHEQUE" && req.ChequeAmount <= 0 {
		return nil, fmt.Errorf("[BR-POL-029] cheque_amount is required when payment_type is COMBINED_CHEQUE")
	}

	// Step 6: Create batch record in database
	batch := &domain.BulkUploadBatch{
		FileName:   req.FileName,
		TotalRows:  req.TotalRows,
		UploadedBy: req.UploadedBy,
	}

	if err := h.bulkUploadRepo.CreateBatch(sctx.Ctx, batch); err != nil {
		log.Error(sctx.Ctx, "[FR-POL-021] Error creating bulk upload batch: %v", err)
		return nil, err
	}

	// Step 7: Start BulkProposalUploadWorkflow via Temporal
	workflowID := fmt.Sprintf("bu-%d", batch.BatchID)

	// Workflow input - matches WF-PI-003 BulkUploadInput
	workflowInput := map[string]interface{}{
		"batch_id":      batch.BatchID,
		"file_name":     req.FileName,
		"total_rows":    req.TotalRows,
		"payment_type":  req.PaymentType,
		"cheque_amount": req.ChequeAmount,
		"uploaded_by":   req.UploadedBy,
	}

	workflowOptions := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: "policy-issue-queue",
	}

	// Start the workflow (fire-and-forget)
	_, err := h.temporalClient.ExecuteWorkflow(sctx.Ctx, workflowOptions, "BulkProposalUploadWorkflow", workflowInput)
	if err != nil {
		log.Error(sctx.Ctx, "[FR-POL-021] Error starting bulk upload workflow: %v", err)
		// Batch was created but workflow failed to start - still return batch ID
		// The batch can be retried later
		log.Warn(sctx.Ctx, "[FR-POL-021] Batch %d created but workflow not started. Manual retry needed.", batch.BatchID)
	}

	// Step 8: Estimate completion time
	estimatedMinutes := (req.TotalRows / 100) + 1 // Rough estimate: ~100 rows per minute
	estimatedCompletion := time.Now().Add(time.Duration(estimatedMinutes) * time.Minute)

	log.Info(sctx.Ctx, "[FR-POL-021] Bulk upload batch accepted: batch_id=%d, file=%s, rows=%d, workflow=%s",
		batch.BatchID, req.FileName, req.TotalRows, workflowID)

	return &resp.BulkUploadResponse{
		StatusCodeAndMessage: port.StatusCodeAndMessage{
			StatusCode: http.StatusAccepted,
			Message:    "Bulk upload accepted for processing",
		},
		BatchID:             batch.BatchID,
		FileName:            req.FileName,
		Status:              "ACCEPTED",
		WorkflowID:          workflowID,
		EstimatedCompletion: &estimatedCompletion,
	}, nil
}

// GetBulkUploadStatus retrieves the status and results of a bulk upload batch
func (h *BulkUploadHandler) GetBulkUploadStatus(sctx *serverRoute.Context, req BatchIDUri) (*resp.BulkUploadStatusResponse, error) {
	// Step 1: Get batch record
	batch, err := h.bulkUploadRepo.GetBatchByID(sctx.Ctx, req.BatchID)
	if err != nil {
		log.Error(sctx.Ctx, "[FR-POL-021] Batch %d not found: %v", req.BatchID, err)
		return nil, err
	}

	// Step 2: Get proposal numbers if batch is completed
	var proposalNumbers []string
	if batch.Status == domain.BulkUploadStatusCompleted && batch.SuccessCount > 0 {
		proposalNumbers, err = h.bulkUploadRepo.GetProposalNumbersByBatchID(sctx.Ctx, req.BatchID)
		if err != nil {
			log.Warn(sctx.Ctx, "[FR-POL-021] Could not fetch proposal numbers for batch %d: %v", req.BatchID, err)
			// Non-fatal, continue without proposal numbers
		}
	}

	// Step 3: Build error report info if failures exist
	var errorReport *resp.ErrorReport
	if batch.ErrorReportDocID != nil && *batch.ErrorReportDocID != "" {
		errorReport = &resp.ErrorReport{
			DocumentID:  *batch.ErrorReportDocID,
			DownloadURL: fmt.Sprintf("/api/v1/documents/%s", *batch.ErrorReportDocID),
		}
	}

	return &resp.BulkUploadStatusResponse{
		StatusCodeAndMessage: port.StatusCodeAndMessage{
			StatusCode: http.StatusOK,
			Message:    "Bulk upload status retrieved successfully",
		},
		BatchID:         batch.BatchID,
		FileName:        batch.FileName,
		Status:          string(batch.Status),
		TotalRows:       batch.TotalRows,
		SuccessCount:    batch.SuccessCount,
		FailureCount:    batch.FailureCount,
		ProposalNumbers: proposalNumbers,
		ErrorReportInfo: errorReport,
		StartedAt:       batch.UploadedAt,
		CompletedAt:     batch.CompletedAt,
	}, nil
}
