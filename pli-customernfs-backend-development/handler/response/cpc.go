// Package response contains HTTP response DTOs for CPC operations endpoints.
//
// Phase 3: Supporting Endpoints
// CPC-001..006 response types.
package response

import (
	"customer-nfs-service/core/port"
	"time"
)

// ─── CPC-001: GET /nfs/cpc/queue ─────────────────────────────────────────────

// CPCQueueItem is a single work queue entry with SLA indicators.
// SLA status: GREEN (<50% elapsed), AMBER (50-80% elapsed), RED (>80% elapsed).
type CPCQueueItem struct {
	RequestID    string  `json:"request_id"`
	TicketNumber string  `json:"ticket_number"`
	CustomerID   int64   `json:"customer_id"`
	RequestType  string  `json:"request_type"`
	Status       string  `json:"status"`
	AuthMethod   string  `json:"auth_method"`
	Channel      string  `json:"channel"`
	OfficeCode   string  `json:"office_code"`
	AssignedTo   *string `json:"assigned_to,omitempty"`
	SLADeadline  string  `json:"sla_deadline"`
	SLAStatus    string  `json:"sla_status"` // GREEN / AMBER / RED
	SLADaysLeft  int     `json:"sla_days_left"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
}

// GetCPCQueueData is the payload returned on a successful CPC-001 call.
type GetCPCQueueData struct {
	// Total is the total number of items matching the filter (before pagination).
	Total int `json:"total"`
	// Page is the current page number.
	Page int `json:"page"`
	// PageSize is the number of items per page.
	PageSize int `json:"page_size"`
	// Queue contains the paginated list of work queue items.
	Queue []CPCQueueItem `json:"queue"`
}

// GetCPCQueueResponse is the top-level success response for CPC-001.
type GetCPCQueueResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	Data                      GetCPCQueueData `json:"data"`
}

// NewGetCPCQueueResponse constructs the CPC-001 response.
func NewGetCPCQueueResponse(total, page, pageSize int, queue []CPCQueueItem) *GetCPCQueueResponse {
	return &GetCPCQueueResponse{
		StatusCodeAndMessage: port.ListSuccess,
		Data:                 GetCPCQueueData{Total: total, Page: page, PageSize: pageSize, Queue: queue},
	}
}

// ─── CPC-002: GET /nfs/cpc/queue/stats ───────────────────────────────────────

// CPCQueueStats is a breakdown of work queue counts.
// BATCH: 5-query batch: by_status, by_type, by_sla, by_channel, overdue.
type CPCQueueStats struct {
	TotalPending    int `json:"total_pending"`
	TotalInProgress int `json:"total_in_progress"`
	TotalOverdue    int `json:"total_overdue"`
	// SLA breakdown
	SLAGreen int `json:"sla_green"`
	SLAAmber int `json:"sla_amber"`
	SLARed   int `json:"sla_red"`
	// Type breakdown
	AddressChanges int `json:"address_changes"`
	NameChanges    int `json:"name_changes"`
}

// GetCPCQueueStatsData is the payload returned on a successful CPC-002 call.
type GetCPCQueueStatsData struct {
	// OfficeCode filters stats to a particular office; nil means system-wide.
	OfficeCode *string `json:"office_code,omitempty"`
	// Stats contains the queue breakdown.
	Stats CPCQueueStats `json:"stats"`
	// GeneratedAt is the ISO-8601 timestamp when the stats snapshot was taken.
	GeneratedAt string `json:"generated_at"`
}

// GetCPCQueueStatsResponse is the top-level success response for CPC-002.
type GetCPCQueueStatsResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	Data                      GetCPCQueueStatsData `json:"data"`
}

// NewGetCPCQueueStatsResponse constructs the CPC-002 response.
func NewGetCPCQueueStatsResponse(officeCode *string, stats CPCQueueStats) *GetCPCQueueStatsResponse {
	return &GetCPCQueueStatsResponse{
		StatusCodeAndMessage: port.FetchSuccess,
		Data: GetCPCQueueStatsData{
			OfficeCode:  officeCode,
			Stats:       stats,
			GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		},
	}
}

// ─── CPC-003: POST /nfs/requests/:request_id/assign ──────────────────────────

// AssignRequestData is the payload returned on a successful CPC-003 call.
// BR-NFS-002: assignment puts request IN_PROGRESS.
type AssignRequestData struct {
	// RequestID identifies the assigned request.
	RequestID string `json:"request_id"`
	// AssignedTo is the CPC officer user ID.
	AssignedTo string `json:"assigned_to"`
	// AssignedAt is the ISO-8601 timestamp of assignment.
	AssignedAt string `json:"assigned_at"`
	// Status reflects the new status: IN_PROGRESS after assignment.
	Status string `json:"status"`
	// Message provides a human-readable outcome.
	Message string `json:"message"`
}

// AssignRequestResponse is the top-level success response for CPC-003.
type AssignRequestResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	Data                      AssignRequestData `json:"data"`
}

// NewAssignRequestResponse constructs the CPC-003 response.
func NewAssignRequestResponse(requestID, assignedTo, status string) *AssignRequestResponse {
	return &AssignRequestResponse{
		StatusCodeAndMessage: port.UpdateSuccess,
		Data: AssignRequestData{
			RequestID:  requestID,
			AssignedTo: assignedTo,
			AssignedAt: time.Now().UTC().Format(time.RFC3339),
			Status:     status,
			Message:    "Request assigned successfully.",
		},
	}
}

// ─── CPC-004: POST /nfs/requests/:request_id/reject ──────────────────────────

// CPCRejectData is the payload returned on a successful CPC-004 call.
// BR-NFS-012: IN_PROGRESS → REJECTED.
type CPCRejectData struct {
	// RequestID identifies the rejected request.
	RequestID string `json:"request_id"`
	// NewStatus reflects the resulting state: REJECTED.
	NewStatus string `json:"new_status"`
	// RejectedBy is the CPC officer who rejected the request.
	RejectedBy string `json:"rejected_by"`
	// RejectedAt is the ISO-8601 timestamp of rejection.
	RejectedAt string `json:"rejected_at"`
	// RejectionReason is the mandatory reason for rejection (BR-NFS-012).
	RejectionReason string `json:"rejection_reason"`
	// Message provides a human-readable outcome.
	Message string `json:"message"`
}

// CPCRejectResponse is the top-level success response for CPC-004.
type CPCRejectResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	Data                      CPCRejectData `json:"data"`
}

// NewCPCRejectResponse constructs the CPC-004 response.
func NewCPCRejectResponse(requestID, rejectedBy, reason string) *CPCRejectResponse {
	return &CPCRejectResponse{
		StatusCodeAndMessage: port.UpdateSuccess,
		Data: CPCRejectData{
			RequestID:       requestID,
			NewStatus:       "REJECTED",
			RejectedBy:      rejectedBy,
			RejectedAt:      time.Now().UTC().Format(time.RFC3339),
			RejectionReason: reason,
			Message:         "Request rejected.",
		},
	}
}

// ─── CPC-005: POST /nfs/requests/:request_id/send-back ───────────────────────

// CPCSendBackData is the payload returned on a successful CPC-005 call.
// Same as RequestMissingDocumentsResponse — alias endpoint.
// BR-NFS-014: status → PENDING_DOCUMENTS; secure link generated.
type CPCSendBackData struct {
	// RequestID identifies the request.
	RequestID string `json:"request_id"`
	// Status reflects the new state: PENDING_DOCUMENTS.
	Status string `json:"status"`
	// SecureUploadURL is the link sent to the customer for document re-upload.
	SecureUploadURL string `json:"secure_upload_url,omitempty"`
	// ExpiresAt is the link expiry timestamp.
	ExpiresAt string `json:"expires_at,omitempty"`
	// AttemptNumber tracks the send-back attempt.
	AttemptNumber int `json:"attempt_number"`
	// MissingDocuments is the list of document types the customer must upload.
	MissingDocuments []string `json:"missing_documents"`
	// Message provides a human-readable outcome.
	Message string `json:"message"`
}

// CPCSendBackResponse is the top-level success response for CPC-005.
type CPCSendBackResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	Data                      CPCSendBackData `json:"data"`
}

// NewCPCSendBackResponse constructs the CPC-005 response.
func NewCPCSendBackResponse(requestID, url, expiresAt string, attempt int, docs []string) *CPCSendBackResponse {
	return &CPCSendBackResponse{
		StatusCodeAndMessage: port.UpdateSuccess,
		Data: CPCSendBackData{
			RequestID:        requestID,
			Status:           "PENDING_DOCUMENTS",
			SecureUploadURL:  url,
			ExpiresAt:        expiresAt,
			AttemptNumber:    attempt,
			MissingDocuments: docs,
			Message:          "Request sent back for additional documents. Customer notified.",
		},
	}
}

// ─── CPC-006: GET /nfs/cpc/sla-dashboard ─────────────────────────────────────

// SLABreachItem represents a single SLA breach record.
type SLABreachItem struct {
	RequestID    string `json:"request_id"`
	TicketNumber string `json:"ticket_number"`
	RequestType  string `json:"request_type"`
	OfficeCode   string `json:"office_code"`
	SLADeadline  string `json:"sla_deadline"`
	DaysOverdue  int    `json:"days_overdue"`
	AssignedTo   string `json:"assigned_to,omitempty"`
}

// GetSLADashboardData is the payload returned on a successful CPC-006 call.
// CPC-006: Lists breached items + aggregate stats.
type GetSLADashboardData struct {
	// OfficeCode filters the dashboard to a particular office; nil means system-wide.
	OfficeCode *string `json:"office_code,omitempty"`
	// Summary contains aggregate SLA statistics.
	Summary CPCQueueStats `json:"summary"`
	// BreachedItems lists requests that have exceeded their SLA deadline.
	BreachedItems []SLABreachItem `json:"breached_items"`
	// GeneratedAt is the ISO-8601 timestamp when the dashboard was generated.
	GeneratedAt string `json:"generated_at"`
}

// GetSLADashboardResponse is the top-level success response for CPC-006.
type GetSLADashboardResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	Data                      GetSLADashboardData `json:"data"`
}

// NewGetSLADashboardResponse constructs the CPC-006 response.
func NewGetSLADashboardResponse(officeCode *string, stats CPCQueueStats, breached []SLABreachItem) *GetSLADashboardResponse {
	return &GetSLADashboardResponse{
		StatusCodeAndMessage: port.FetchSuccess,
		Data: GetSLADashboardData{
			OfficeCode:    officeCode,
			Summary:       stats,
			BreachedItems: breached,
			GeneratedAt:   time.Now().UTC().Format(time.RFC3339),
		},
	}
}
