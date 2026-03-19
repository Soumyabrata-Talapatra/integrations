package domain

import (
	"encoding/json"
	"time"
)

// ============================================================================
// TerminalStateSnapshot — Persisted Final State for Post-Workflow Queries
// Source: §9.5.1 Terminal Cooling, DDL: policy_mgmt.terminal_state_snapshot
// Scale: 1 row per terminated policy (sparse, grows over years)
// PK: policy_id
// ============================================================================

// TerminalStateSnapshot is written by PersistTerminalStateActivity when a policy
// enters a terminal state. It enables the Two-Tier Query pattern (AD-011):
// - Tier 1: QueryWorkflow (active/cooling workflow)
// - Tier 2 (fallback): Read from this table when workflow has completed
//
// The final_snapshot JSONB field contains the full serialized PolicyLifecycleState
// from the workflow, enabling REST API responses after workflow completion.
type TerminalStateSnapshot struct {
	// PK
	PolicyID     int64  `json:"policy_id"     db:"policy_id"`
	PolicyNumber string `json:"policy_number" db:"policy_number"`

	// Terminal state info
	FinalStatus string    `json:"final_status" db:"final_status"` // lifecycle_status — one of TerminalStatuses
	TerminalAt  time.Time `json:"terminal_at"  db:"terminal_at"`
	CoolingExpiry time.Time `json:"cooling_expiry" db:"cooling_expiry"` // terminal_at + cooling_period_{state}

	// Workflow completion (set by MarkWorkflowCompletedActivity when cooling expires)
	WorkflowCompletedAt *time.Time `json:"workflow_completed_at,omitempty" db:"workflow_completed_at"`

	// Full workflow state snapshot (serialized PolicyLifecycleState from workflow)
	FinalSnapshot json.RawMessage `json:"final_snapshot" db:"final_snapshot"`

	// Audit
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// ============================================================================
// DashboardSummary — Aggregated Request Counts for CPC Dashboard
// Source: FR-PM-008; aggregated from policy_mgmt.service_request
// No DDL table — computed at query time
// ============================================================================

// DashboardSummary provides a breakdown of pending requests by type and status.
// Used by GET /api/v1/requests/pending/summary endpoint.
type DashboardSummary struct {
	// Per request-type breakdown: map[request_type]map[status]count
	Summary             map[string]map[string]int `json:"summary"`
	TotalPending        int                       `json:"total_pending"`
	OldestRequestAgeHrs float64                   `json:"oldest_request_age_hours"`
}

// ============================================================================
// DashboardMetrics — Policy Aggregate Counts for Admin Dashboard
// Source: FR-PM-008; aggregated from mv_policy_dashboard materialized view
// No DDL table — computed from materialized view
// ============================================================================

// DashboardMetrics provides aggregate policy counts for admin dashboard.
// Used by GET /api/v1/policies/dashboard/metrics endpoint.
// Data sourced from mv_policy_dashboard materialized view (refreshed every 15 min).
type DashboardMetrics struct {
	PoliciesByStatus        map[string]int64 `json:"policies_by_status"`
	PoliciesByProduct       map[string]int64 `json:"policies_by_product"`
	PoliciesByBillingMethod map[string]int64 `json:"policies_by_billing_method"`
	RequestsToday           int              `json:"requests_today"`
	RequestsPending         int              `json:"requests_pending"`
}

// ============================================================================
// LookupItem — Generic Lookup Entry
// Used by GET /api/v1/lookups/* endpoints
// ============================================================================

// LookupItem represents a single entry in a static lookup list.
type LookupItem struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
}
