package response

import "policy-issue-service/core/port"

// LookupItem represents a single lookup/reference data item
type LookupItem struct {
	Code        string `json:"code"`
	Label       string `json:"label"`
	Description string `json:"description,omitempty"`
}

// LookupResponse represents the response for all Lookup APIs
// Used by: [LU-POL-001] to [LU-POL-010]
type LookupResponse struct {
	port.StatusCodeAndMessage
	Items []LookupItem `json:"items"`
}
