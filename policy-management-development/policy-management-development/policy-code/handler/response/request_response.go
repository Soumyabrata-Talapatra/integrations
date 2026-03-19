package response

import (
	"time"

	"policy-management/core/domain"
	"policy-management/core/port"
)

// ============================================================================
// Request Lifecycle Response DTOs
// Source: Swagger components/schemas: RequestAcceptedResponse, RequestDetailResponse,
//         RequestListResponse, RequestSummary, WithdrawalResponse
// Used by: POST /requests/*, GET /requests/*, PUT /requests/*/withdraw
// ============================================================================

// ── POST /policies/{pn}/requests/* (all submission endpoints) ────────────────

// RequestAcceptedData is the 202 Accepted payload for all request submissions.
// Contains enough information for the caller to track the request.
type RequestAcceptedData struct {
	RequestID         int64  `json:"request_id"`         // BIGINT from service_request
	PolicyNumber      string `json:"policy_number"`
	RequestType       string `json:"request_type"`
	RequestCategory   string `json:"request_category"`
	Status            string `json:"status"`             // "ROUTED"
	StateGateStatus   string `json:"state_gate_status"`  // Policy status at check time
	DownstreamService string `json:"downstream_service,omitempty"`
	SubmittedAt       string `json:"submitted_at"`
	TimeoutAt         string `json:"timeout_at,omitempty"`
	TrackingURL       string `json:"tracking_url,omitempty"`
}

// NewRequestAcceptedData builds the 202 response from a service request.
func NewRequestAcceptedData(sr domain.ServiceRequest) RequestAcceptedData {
	d := RequestAcceptedData{
		RequestID:       sr.RequestID,
		PolicyNumber:    sr.PolicyNumber,
		RequestType:     sr.RequestType,
		RequestCategory: sr.RequestCategory,
		Status:          sr.Status,
		SubmittedAt:     sr.SubmittedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
	if sr.StateGateStatus != nil {
		d.StateGateStatus = *sr.StateGateStatus
	}
	if sr.DownstreamService != nil {
		d.DownstreamService = *sr.DownstreamService
	}
	if sr.TimeoutAt != nil {
		d.TimeoutAt = sr.TimeoutAt.Format("2006-01-02T15:04:05Z07:00")
	}
	return d
}

// RequestAcceptedResponse — 202 response for all request submission endpoints.
type RequestAcceptedResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	Data                      RequestAcceptedData `json:"data"`
}

// ── GET /policies/{pn}/requests/{request_id} ────────────────────────────────

// TimelineEvent is a single step in the request processing timeline.
type TimelineEvent struct {
	Event  string  `json:"event"`
	At     string  `json:"at"`
	Detail *string `json:"detail,omitempty"`
}

// RequestDetailData is the full request detail payload.
type RequestDetailData struct {
	RequestID            int64           `json:"request_id"`
	PolicyNumber         string          `json:"policy_number"`
	RequestType          string          `json:"request_type"`
	RequestCategory      string          `json:"request_category"`
	Status               string          `json:"status"`
	SourceChannel        string          `json:"source_channel"`
	SubmittedBy          *int64          `json:"submitted_by,omitempty"`
	SubmittedAt          string          `json:"submitted_at"`
	StateGateStatus      *string         `json:"state_gate_status,omitempty"`
	RoutedAt             *string         `json:"routed_at,omitempty"`
	DownstreamService    *string         `json:"downstream_service,omitempty"`
	DownstreamWorkflowID *string         `json:"downstream_workflow_id,omitempty"`
	TimeoutAt            *string         `json:"timeout_at,omitempty"`
	CompletedAt          *string         `json:"completed_at,omitempty"`
	Outcome              *string         `json:"outcome,omitempty"`
	OutcomeReason        *string         `json:"outcome_reason,omitempty"`
	Timeline             []TimelineEvent `json:"timeline,omitempty"`
}

// NewRequestDetailData builds RequestDetailData from a domain.ServiceRequest.
func NewRequestDetailData(sr domain.ServiceRequest) RequestDetailData {
	d := RequestDetailData{
		RequestID:            sr.RequestID,
		PolicyNumber:         sr.PolicyNumber,
		RequestType:          sr.RequestType,
		RequestCategory:      sr.RequestCategory,
		Status:               sr.Status,
		SourceChannel:        sr.SourceChannel,
		SubmittedBy:          sr.SubmittedBy,
		SubmittedAt:          sr.SubmittedAt.Format("2006-01-02T15:04:05Z07:00"),
		StateGateStatus:      sr.StateGateStatus,
		DownstreamService:    sr.DownstreamService,
		DownstreamWorkflowID: sr.DownstreamWorkflowID,
		Outcome:              sr.Outcome,
		OutcomeReason:        sr.OutcomeReason,
	}
	if sr.RoutedAt != nil {
		s := sr.RoutedAt.Format("2006-01-02T15:04:05Z07:00")
		d.RoutedAt = &s
	}
	if sr.TimeoutAt != nil {
		s := sr.TimeoutAt.Format("2006-01-02T15:04:05Z07:00")
		d.TimeoutAt = &s
	}
	if sr.CompletedAt != nil {
		s := sr.CompletedAt.Format("2006-01-02T15:04:05Z07:00")
		d.CompletedAt = &s
	}
	return d
}

// RequestDetailResponse — GET /api/v1/policies/{pn}/requests/{request_id}
// [FR-PM-006]
type RequestDetailResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	Data                      RequestDetailData `json:"data"`
}

// ── GET /policies/{pn}/requests ──────────────────────────────────────────────

// RequestSummaryData is a condensed request entry for list views.
type RequestSummaryData struct {
	RequestID       int64   `json:"request_id"`
	PolicyNumber    string  `json:"policy_number"`
	RequestType     string  `json:"request_type"`
	RequestCategory string  `json:"request_category"`
	Status          string  `json:"status"`
	SourceChannel   string  `json:"source_channel"`
	SubmittedAt     string  `json:"submitted_at"`
	Outcome         *string `json:"outcome,omitempty"`
	AgeHours        float64 `json:"age_hours"`
}

// NewRequestSummaryData builds RequestSummaryData from a domain.ServiceRequest.
// Enhancement-2: AgeHours is computed from SubmittedAt to now, giving CPC agents
// a real-time measure of how long a request has been waiting.
func NewRequestSummaryData(sr domain.ServiceRequest) RequestSummaryData {
	return RequestSummaryData{
		RequestID:       sr.RequestID,
		PolicyNumber:    sr.PolicyNumber,
		RequestType:     sr.RequestType,
		RequestCategory: sr.RequestCategory,
		Status:          sr.Status,
		SourceChannel:   sr.SourceChannel,
		SubmittedAt:     sr.SubmittedAt.Format("2006-01-02T15:04:05Z07:00"),
		Outcome:         sr.Outcome,
		AgeHours:        time.Since(sr.SubmittedAt).Hours(), // Enhancement-2: elapsed since submission
	}
}

// NewRequestSummaryDataList converts a slice of service requests to response DTOs.
func NewRequestSummaryDataList(srs []domain.ServiceRequest) []RequestSummaryData {
	result := make([]RequestSummaryData, 0, len(srs))
	for _, sr := range srs {
		result = append(result, NewRequestSummaryData(sr))
	}
	return result
}

// RequestListResponse — GET /api/v1/policies/{pn}/requests
// [FR-PM-006]
type RequestListResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	port.MetaDataResponse     `json:",inline"`
	Data                      []RequestSummaryData `json:"data"`
}

// ── PUT /policies/{pn}/requests/{request_id}/withdraw ────────────────────────

// WithdrawalData is the response payload for a successful withdrawal.
type WithdrawalData struct {
	RequestID            int64   `json:"request_id"`
	Status               string  `json:"status"`               // "WITHDRAWN"
	WithdrawnAt          string  `json:"withdrawn_at"`
	PreviousPolicyStatus *string `json:"previous_policy_status,omitempty"` // Policy status before withdrawal
	RevertedPolicyStatus *string `json:"reverted_policy_status,omitempty"` // Policy status after revert
}

// WithdrawalResponse — PUT /api/v1/policies/{pn}/requests/{request_id}/withdraw
// [FR-PM-007] [BR-PM-090]
type WithdrawalResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	Data                      WithdrawalData `json:"data"`
}
