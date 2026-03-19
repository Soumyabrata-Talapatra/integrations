package domain

import (
	"encoding/json"
	"time"
)

// ============================================================================
// PolicyEvent — Published State Change Events
// Source: §8.5, DDL: policy_mgmt.policy_event, ADR-004
// Scale: 1 event per state transition → high volume; Retention: 7 years
// Partitioned by published_at (yearly partitions)
// ⚠️ PARTITION KEY: published_at — MUST appear in all WHERE and INSERT
// ============================================================================

// PolicyEvent is published by PM on every state change.
// Consumed by: Notification, Accounting, Agent, Audit services.
// Written by PublishEventActivity.
type PolicyEvent struct {
	// PK (composite with published_at due to partitioning)
	EventID     int64  `json:"event_id"   db:"event_id"`
	PolicyID    int64  `json:"policy_id"  db:"policy_id"`

	// Event detail
	EventType    string          `json:"event_type"    db:"event_type"`    // e.g. "PolicyLapsedVoid", "PolicySurrendered"
	EventPayload json.RawMessage `json:"event_payload" db:"event_payload"` // JSONB — typed per event_type

	// Timing
	PublishedAt time.Time `json:"published_at" db:"published_at"` // ⚠️ PARTITION KEY

	// Downstream tracking
	ConsumedBy []string `json:"consumed_by,omitempty" db:"consumed_by"` // TEXT[] — services that consumed the event
}
