package response

import "policy-management/core/port"

// ============================================================================
// Lookup Response DTOs
// Source: Swagger components/schemas: LookupListResponse, LifecycleStatesResponse
// Used by: GET /lookups/request-types, /lifecycle-states, /source-channels,
//          /product-types, /premium-modes, /billing-methods
// ============================================================================

// ── GET /lookups/{type} (generic enum lookups) ───────────────────────────────

// LookupItemData is a single code/label pair returned for enum lookups.
// Maps to domain.LookupItem when converted through constructors.
type LookupItemData struct {
	Code        string  `json:"code"`
	Label       string  `json:"label"`
	Description *string `json:"description,omitempty"` // Optional long description
}

// NewLookupItemData builds a LookupItemData from code, label, and optional description.
func NewLookupItemData(code, label string, description *string) LookupItemData {
	return LookupItemData{
		Code:        code,
		Label:       label,
		Description: description,
	}
}

// LookupListResponse — GET /api/v1/lookups/{type}
// Used for: request-types, source-channels, product-types,
//           premium-modes, billing-methods.
type LookupListResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	Items                     []LookupItemData `json:"items"`
}

// ── GET /lookups/lifecycle-states (extended lifecycle state detail) ───────────

// LifecycleStateData extends LookupItemData with lifecycle-specific metadata.
// The `category` field groups states into operational buckets for display.
type LifecycleStateData struct {
	Code        string `json:"code"`
	ShortCode   string `json:"short_code,omitempty"` // Abbreviated code, e.g. "VL" for VOID_LAPSE
	Category    string `json:"category"`             // "Active" | "Lapsed" | "Pending" | "Terminal"
	Description string `json:"description"`
	IsTerminal  bool   `json:"is_terminal"` // true for states in domain.TerminalStatuses
}

// LifecycleStatesResponse — GET /api/v1/lookups/lifecycle-states
// Returns the full 23-state catalogue with category groupings and terminal flags.
type LifecycleStatesResponse struct {
	port.StatusCodeAndMessage `json:",inline"`
	States                    []LifecycleStateData `json:"states"`
}
