package response

import "policy-management/core/port"

// ============================================================================
// CPC & Dashboard Response DTOs
// Source: Swagger components/schemas: PendingRequestsResponse,
//         DashboardSummaryResponse, DashboardMetricsResponse
// Used by: GET /cpc/requests, GET /cpc/dashboard,
//          GET /policies/dashboard/metrics
// ============================================================================

// ── GET /cpc/requests (CPC inbox — pending request queue) ────────────────────

// PendingRequestsData is the CPC inbox payload.
// Contains a page of pending service requests awaiting CPC action.
type PendingRequestsData struct {
	TotalCount int                  `json:"total_count"` // Total matching records (unpaged)
	Page       int                  `json:"page"`        // Current page (1-based)
	PageSize   int                  `json:"page_size"`   // Records per page
	Requests   []RequestSummaryData `json:"requests"`    // RequestSummaryData reused from request_response.go
}

// PendingRequestsResponse — GET /api/v1/cpc/requests
// [FR-PM-006]
type PendingRequestsResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	Data                      PendingRequestsData `json:"data"`
}

// ── GET /cpc/dashboard ───────────────────────────────────────────────────────

// RequestStatusCounts maps request_status → count within a single request type bucket.
// Example: { "RECEIVED": 5, "ROUTED": 32, "IN_PROGRESS": 8 }
type RequestStatusCounts map[string]int

// DashboardSummaryData is the CPC dashboard payload.
// Summary groups active requests by type then by status for at-a-glance triage.
type DashboardSummaryData struct {
	// Summary is a two-level map: request_type → request_status → count.
	// Example: { "SURRENDER": { "RECEIVED": 5, "IN_PROGRESS": 8 }, "REVIVAL": { "ROUTED": 18 } }
	Summary                map[string]RequestStatusCounts `json:"summary"`
	TotalPending           int                            `json:"total_pending"`            // Sum of all non-terminal request counts
	OldestRequestAgeHours  float64                        `json:"oldest_request_age_hours"` // Age of oldest ROUTED/IN_PROGRESS request
}

// DashboardSummaryResponse — GET /api/v1/cpc/dashboard
type DashboardSummaryResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	Data                      DashboardSummaryData `json:"data"`
}

// ── GET /policies/dashboard/metrics ──────────────────────────────────────────

// DashboardMetricsData is the policy management metrics payload.
// Used by operations dashboards to monitor portfolio health.
type DashboardMetricsData struct {
	// PoliciesByStatus maps lifecycle_status → policy count.
	// Example: { "ACTIVE": 2150000, "VOID_LAPSE": 120000 }
	PoliciesByStatus map[string]int `json:"policies_by_status"`

	// PoliciesByProduct maps product_code → policy count.
	// Example: { "WLA": 1200000, "EA": 800000 }
	PoliciesByProduct map[string]int `json:"policies_by_product"`

	// PoliciesByBillingMethod maps billing_method → policy count.
	// Example: { "PAY_RECOVERY": 2100000, "CASH": 900000 }
	PoliciesByBillingMethod map[string]int `json:"policies_by_billing_method"`

	RequestsToday   int `json:"requests_today"`   // New service requests created today
	RequestsPending int `json:"requests_pending"` // Total non-terminal requests (RECEIVED + ROUTED + IN_PROGRESS)
}

// DashboardMetricsResponse — GET /api/v1/policies/dashboard/metrics
type DashboardMetricsResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	Data                      DashboardMetricsData `json:"data"`
}
