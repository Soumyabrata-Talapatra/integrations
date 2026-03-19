package domain

import (
	"encoding/json"
	"time"
)

// ============================================================================
// Signal Processing Status Constants
// Source: DDL: signal_processing_status enum
// ============================================================================

const (
	SignalStatusProcessed = "PROCESSED"
	SignalStatusRejected  = "REJECTED"
	SignalStatusDuplicate = "DUPLICATE"
	SignalStatusFailed    = "FAILED"
)

// ============================================================================
// ProcessedSignalRegistry — Signal Deduplication Table
// Source: §8.8, DDL: policy_mgmt.processed_signal_registry
// Scale: Grows with signal volume; evicted after 90 days (signal_dedup_ttl_days)
// UNIQUE: (request_id, signal_type)
// ⚠️ request_id here is VARCHAR(100) — the UUID string from the caller
//    (different from service_request.request_id which is a BIGINT)
// ============================================================================

// ProcessedSignalRegistry tracks which signals have been processed to prevent
// duplicate processing. Entries are auto-evicted after signal_dedup_ttl_days.
type ProcessedSignalRegistry struct {
	ID         int64     `json:"id"          db:"id"`
	RequestID  string    `json:"request_id"  db:"request_id"`  // VARCHAR(100) — UUID from caller
	SignalType string    `json:"signal_type" db:"signal_type"` // e.g. "policy-created", "surrender-request"
	PolicyID   int64     `json:"policy_id"   db:"policy_id"`
	ReceivedAt time.Time `json:"received_at" db:"received_at"`
	ExpiresAt  time.Time `json:"expires_at"  db:"expires_at"`  // received_at + signal_dedup_ttl_days
}

// ============================================================================
// PolicySignalLog — Full Signal Audit Trail
// Source: §8.9, DDL: policy_mgmt.policy_signal_log
// Scale: Every signal logged — high volume; Retention: 3 years
// Partitioned by received_at (yearly)
// ⚠️ PARTITION KEY: received_at — MUST appear in all WHERE and INSERT
// ============================================================================

// PolicySignalLog records every signal received by a PolicyLifecycleWorkflow,
// including rejected, duplicate, and failed signals.
// Essential for production debugging, compliance auditing, and signal replay.
// Written by LogSignalReceivedActivity (pgx.Batch with ProcessedSignalRegistry).
type PolicySignalLog struct {
	// PK (composite with received_at due to partitioning)
	ID       int64  `json:"id"        db:"id"`
	PolicyID int64  `json:"policy_id" db:"policy_id"`

	// Signal detail
	SignalChannel    string          `json:"signal_channel"              db:"signal_channel"`    // e.g. "surrender-request"
	SignalPayload    json.RawMessage `json:"signal_payload"              db:"signal_payload"`    // JSONB — full signal body
	SourceService    string          `json:"source_service"              db:"source_service"`
	SourceWorkflowID *string         `json:"source_workflow_id,omitempty" db:"source_workflow_id"`
	RequestID        string          `json:"request_id"                  db:"request_id"` // VARCHAR(100) — UUID string

	// Processing
	ReceivedAt      time.Time  `json:"received_at"               db:"received_at"`      // ⚠️ PARTITION KEY
	ProcessedAt     *time.Time `json:"processed_at,omitempty"    db:"processed_at"`
	Status          string     `json:"status"                    db:"status"`          // signal_processing_status enum
	RejectionReason *string    `json:"rejection_reason,omitempty" db:"rejection_reason"`

	// State context
	StateBefore *string `json:"state_before,omitempty" db:"state_before"` // lifecycle_status before
	StateAfter  *string `json:"state_after,omitempty"  db:"state_after"`  // lifecycle_status after
}
