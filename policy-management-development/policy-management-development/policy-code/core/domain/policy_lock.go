package domain

import "time"

// ============================================================================
// PolicyLock — Financial Request Mutual Exclusion
// Source: §8.4, DDL: policy_mgmt.policy_lock, BR-PM-030
// Scale: Sparse — at most 1 row per policy at any time
// PK: policy_id (guarantees at-most-one lock per policy)
// ============================================================================

// PolicyLock enforces that only ONE financial request can be in-flight per policy
// at a time. Acquired when a financial request is ROUTED to a downstream service.
// Released when the downstream service signals completion (any outcome).
// Auto-released if timeout_at elapses without a completion signal.
type PolicyLock struct {
	// PK = policy_id (enforces uniqueness)
	PolicyID    int64  `json:"policy_id"    db:"policy_id"`
	RequestID   int64  `json:"request_id"   db:"request_id"`   // BIGINT from service_request
	RequestType string `json:"request_type" db:"request_type"` // request_type enum

	// Timestamps
	LockedAt  time.Time `json:"locked_at"  db:"locked_at"`
	TimeoutAt time.Time `json:"timeout_at" db:"timeout_at"` // NOW() + routing_timeout_{type} from policy_state_config
}
