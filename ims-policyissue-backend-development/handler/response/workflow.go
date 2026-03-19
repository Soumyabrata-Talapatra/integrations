package response

import (
	"time"

	"policy-issue-service/core/port"
)

// ============================================
// WF-POL-001: Workflow State Response
// ============================================

// WorkflowProgress represents the workflow step progress
type WorkflowProgress struct {
	CompletedSteps int `json:"completed_steps"`
	TotalSteps     int `json:"total_steps"`
	Percentage     int `json:"percentage"`
}

// WorkflowStateResponse represents the response for WF-POL-001
// [WF-POL-001] Get Temporal workflow state
// Components: WF-PI-001
type WorkflowStateResponse struct {
	port.StatusCodeAndMessage
	WorkflowID     string           `json:"workflow_id"`
	WorkflowType   string           `json:"workflow_type"`
	Status         string           `json:"status"`
	CurrentStep    string           `json:"current_step"`
	Progress       WorkflowProgress `json:"progress"`
	PendingSignals []string         `json:"pending_signals"`
	StartTime      *time.Time       `json:"start_time,omitempty"`
	LastActivity   *time.Time       `json:"last_activity,omitempty"`
}

// ============================================
// WF-POL-002: Workflow Signal Response
// ============================================

// WorkflowSignalResponse represents the response for WF-POL-002
// [WF-POL-002] Send signal to workflow
type WorkflowSignalResponse struct {
	port.StatusCodeAndMessage
	WorkflowID string `json:"workflow_id"`
	SignalName string `json:"signal_name"`
	Accepted   bool   `json:"accepted"`
}

// ============================================
// Bulk Upload Responses
// ============================================

// BulkUploadResponse represents the response for uploading a bulk proposal file
// [FR-POL-021] Bulk Proposal Upload
// Temporal Workflow: WF-PI-003 (BulkProposalUploadWorkflow)
type BulkUploadResponse struct {
	port.StatusCodeAndMessage
	BatchID             int64      `json:"batch_id"`
	FileName            string     `json:"file_name"`
	Status              string     `json:"status"`
	WorkflowID          string     `json:"workflow_id"`
	EstimatedCompletion *time.Time `json:"estimated_completion,omitempty"`
}

// ErrorReport represents the error report for a batch
type ErrorReport struct {
	DocumentID  string `json:"document_id,omitempty"`
	DownloadURL string `json:"download_url,omitempty"`
}

// BulkUploadStatusResponse represents the response for getting bulk upload status
type BulkUploadStatusResponse struct {
	port.StatusCodeAndMessage
	BatchID         int64        `json:"batch_id"`
	FileName        string       `json:"file_name"`
	Status          string       `json:"status"`
	TotalRows       int          `json:"total_rows"`
	SuccessCount    int          `json:"success_count"`
	FailureCount    int          `json:"failure_count"`
	ProposalNumbers []string     `json:"proposal_numbers,omitempty"`
	ErrorReportInfo *ErrorReport `json:"error_report,omitempty"`
	StartedAt       time.Time    `json:"started_at"`
	CompletedAt     *time.Time   `json:"completed_at,omitempty"`
}
