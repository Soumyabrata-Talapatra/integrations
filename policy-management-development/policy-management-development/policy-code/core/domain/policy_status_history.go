package domain

import (
	"encoding/json"
	"time"
)

// ============================================================================
// PolicyStatusHistory — Complete Lifecycle Transition Audit Trail
// Source: §8.2, DDL: policy_mgmt.policy_status_history
// Scale: ~10 transitions avg per policy → 50M+ rows; Retention: 10 years
// Partitioned by effective_date (yearly partitions)
// ⚠️ PARTITION KEY: effective_date — MUST appear in all WHERE and INSERT
// PK: (id, effective_date) — composite due to partitioning
// ============================================================================

// PolicyStatusHistory records every state transition for a policy.
// Written by PM via RecordStateTransitionActivity (pgx.Batch with policy UPDATE).
type PolicyStatusHistory struct {
	// PK (composite with effective_date due to partitioning)
	ID       int64  `json:"id"        db:"id"`
	PolicyID int64  `json:"policy_id" db:"policy_id"`

	// Transition detail
	FromStatus           *string         `json:"from_status,omitempty"            db:"from_status"`            // NULL for initial FREE_LOOK_ACTIVE
	ToStatus             string          `json:"to_status"                        db:"to_status"`
	TransitionReason     string          `json:"transition_reason"                db:"transition_reason"`
	TriggeredByService   string          `json:"triggered_by_service"             db:"triggered_by_service"`
	TriggeredByUserID    *int64          `json:"triggered_by_user_id,omitempty"   db:"triggered_by_user_id"`
	RequestID            *int64          `json:"request_id,omitempty"             db:"request_id"`            // BIGINT FK to service_request
	EffectiveDate        time.Time       `json:"effective_date"                   db:"effective_date"`        // ⚠️ PARTITION KEY
	MetadataSnapshot     json.RawMessage `json:"metadata_snapshot,omitempty"      db:"metadata_snapshot"`     // JSONB snapshot

	// Audit
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
