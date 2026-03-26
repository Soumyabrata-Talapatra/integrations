// Package response contains HTTP response DTOs for status and tracking endpoints.
//
// Phase 3: Supporting Endpoints
// ST-001..005 response types.
package response

import (
	"customer-nfs-service/core/port"
	"time"
)

// ─── ST-001: GET /nfs/requests/:request_id ───────────────────────────────────

// RequestDetailDocumentItem is a document attached to a request.
type RequestDetailDocumentItem struct {
	DocumentID         string `json:"document_id"`
	DocumentType       string `json:"document_type"`
	FileName           string `json:"file_name"`
	FileSizeBytes      int64  `json:"file_size_bytes"`
	VerificationStatus string `json:"verification_status"` // PENDING, VERIFIED, REJECTED
	UploadedAt         string `json:"uploaded_at"`
}

// AuditTrailItem is a single audit log entry.
type AuditTrailItem struct {
	ActionType  string  `json:"action_type"`
	PerformedBy string  `json:"performed_by"`
	PerformedAt string  `json:"performed_at"`
	OldValue    *string `json:"old_value,omitempty"`
	NewValue    *string `json:"new_value,omitempty"`
	Notes       *string `json:"notes,omitempty"`
}

// RequestDetailData is the payload for ST-001.
// BATCH: service_request + type-specific detail (address or name) in single round-trip.
type RequestDetailData struct {
	RequestID    string  `json:"request_id"`
	TicketNumber string  `json:"ticket_number"`
	CustomerID   string  `json:"customer_id"`
	PolicyNumber *string `json:"policy_number,omitempty"`
	RequestType  string  `json:"request_type"`
	Status       string  `json:"status"`
	AuthMethod   string  `json:"auth_method"`
	Channel      string  `json:"channel"`
	OfficeCode   *string `json:"office_code,omitempty"`
	SLADeadline  *string `json:"sla_deadline,omitempty"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
	// Type-specific sub-objects (one will be populated based on request_type)
	AddressDetail *AddressDetailSummary `json:"address_detail,omitempty"`
	NameDetail    *NameDetailSummary    `json:"name_detail,omitempty"`
	// Optional includes
	Documents  []RequestDetailDocumentItem `json:"documents,omitempty"`
	AuditTrail []AuditTrailItem            `json:"audit_trail,omitempty"`
}

// RequestDetailResponse is the top-level success response for ST-001.
type RequestDetailResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	Data                      RequestDetailData `json:"data"`
}

// AddressDetailSummary is a summary of address change detail fields.
type AddressDetailSummary struct {
	AddressType     string `json:"address_type"`
	NewAddressLine1 string `json:"new_address_line1"`
	NewCity         string `json:"new_city"`
	NewDistrict     string `json:"new_district"`
	NewState        string `json:"new_state"`
	NewPincode      string `json:"new_pincode"`
}

// NameDetailSummary is a summary of name change detail fields.
// BR-NFS-010: DOB fields intentionally absent.
type NameDetailSummary struct {
	OldName          string `json:"old_name,omitempty"` // "old_salutation old_first_name old_last_name"
	NewName          string `json:"new_name"`           // "new_salutation new_first_name new_last_name"
	PoliciesAffected int    `json:"policies_affected"`
}

// ─── ST-002: GET /nfs/requests/:request_id/timeline ──────────────────────────

// TimelineData is the payload for ST-002.
// FR-NFS-008: Audit trail history in chronological order.
type TimelineData struct {
	RequestID    string           `json:"request_id"`
	TicketNumber string           `json:"ticket_number"`
	Timeline     []AuditTrailItem `json:"timeline"`
}

// TimelineResponse is the top-level success response for ST-002.
type TimelineResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	Data                      TimelineData `json:"data"`
}

// ─── ST-003: GET /nfs/requests/:request_id/receipt ───────────────────────────

// ReceiptData is the payload for ST-003.
// Binary PDF returned with Content-Type: application/pdf.
// This struct conveys metadata if JSON is requested, but typically the handler
// writes PDF bytes directly to the response writer.
type ReceiptData struct {
	RequestID    string `json:"request_id"`
	TicketNumber string `json:"ticket_number"`
	ReceiptURL   string `json:"receipt_url"`
	GeneratedAt  string `json:"generated_at"`
}

// ReceiptResponse is the top-level success response for ST-003.
type ReceiptResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	Data                      ReceiptData `json:"data"`
}

// ─── ST-004: GET /nfs/customers/:customer_id/requests ────────────────────────

// CustomerRequestSummary is a single row in the paginated customer request list.
type CustomerRequestSummary struct {
	RequestID    string  `json:"request_id"`
	TicketNumber string  `json:"ticket_number"`
	RequestType  string  `json:"request_type"`
	Status       string  `json:"status"`
	AuthMethod   string  `json:"auth_method"`
	Channel      string  `json:"channel"`
	SLADeadline  *string `json:"sla_deadline,omitempty"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
}

// ListCustomerRequestsData is the payload for ST-004.
// BATCH: count + rows in single round-trip.
type ListCustomerRequestsData struct {
	CustomerID string                   `json:"customer_id"`
	Total      int                      `json:"total"`
	Page       int                      `json:"page"`
	PageSize   int                      `json:"page_size"`
	Requests   []CustomerRequestSummary `json:"requests"`
}

// ListCustomerRequestsResponse is the top-level success response for ST-004.
type ListCustomerRequestsResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	Data                      ListCustomerRequestsData `json:"data"`
}

// NewListCustomerRequestsResponse constructs the ST-004 response.
func NewListCustomerRequestsResponse(customerID string, total, page, pageSize int, requests []CustomerRequestSummary) *ListCustomerRequestsResponse {
	return &ListCustomerRequestsResponse{
		StatusCodeAndMessage: port.ListSuccess,
		Data: ListCustomerRequestsData{
			CustomerID: customerID,
			Total:      total,
			Page:       page,
			PageSize:   pageSize,
			Requests:   requests,
		},
	}
}

// ─── ST-005: GET /nfs/requests/:request_id/documents ─────────────────────────

// RequestDocumentsData is the payload for ST-005.
type RequestDocumentsData struct {
	RequestID string                      `json:"request_id"`
	Documents []RequestDetailDocumentItem `json:"documents"`
	Total     int                         `json:"total"`
}

// RequestDocumentsResponse is the top-level success response for ST-005.
type RequestDocumentsResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	Data                      RequestDocumentsData `json:"data"`
}

// NewRequestDocumentsResponse constructs the ST-005 response.
func NewRequestDocumentsResponse(requestID string, docs []RequestDetailDocumentItem) *RequestDocumentsResponse {
	return &RequestDocumentsResponse{
		StatusCodeAndMessage: port.ListSuccess,
		Data: RequestDocumentsData{
			RequestID: requestID,
			Documents: docs,
			Total:     len(docs),
		},
	}
}

// strPtr is a local helper (shared between response helpers).
func strPtr(s string) *string { return &s }

// FormatTime formats a time.Time as RFC3339.
func FormatTime(t time.Time) string { return t.UTC().Format(time.RFC3339) }
