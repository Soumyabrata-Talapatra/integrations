package domain

import (
	"encoding/json"
	"time"
)

// ============================================================================
// Batch Scan Type Constants
// Source: DDL: batch_scan_type enum, Plan Constraint 5
// ============================================================================

const (
	BatchScanTypeLapsation           = "LAPSATION"
	BatchScanTypeRemissionExpiryShort = "REMISSION_EXPIRY_SHORT"
	BatchScanTypeRemissionExpiryLong  = "REMISSION_EXPIRY_LONG"
	BatchScanTypePaidUpConversion    = "PAID_UP_CONVERSION"
	BatchScanTypeMaturityScan        = "MATURITY_SCAN"
	BatchScanTypeForcedSurrenderEval = "FORCED_SURRENDER_EVAL"
)

// BatchScanScheduleIDs maps scan types to their Temporal Schedule IDs.
// Source: Plan Constraint 5, bootstrap/bootstrapper.go (Phase 5)
var BatchScanScheduleIDs = map[string]string{
	BatchScanTypeLapsation:            "batch-lapsation-daily",
	BatchScanTypeRemissionExpiryShort: "batch-remission-short-daily",
	BatchScanTypeRemissionExpiryLong:  "batch-remission-long-daily",
	BatchScanTypePaidUpConversion:     "batch-paidup-monthly",
	BatchScanTypeMaturityScan:         "batch-maturity-daily",
	BatchScanTypeForcedSurrenderEval:  "batch-forced-surrender-monthly",
}

// ============================================================================
// Batch Scan Status Constants
// Source: DDL: batch_scan_status enum
// ============================================================================

const (
	BatchScanStatusPending   = "PENDING"
	BatchScanStatusRunning   = "RUNNING"
	BatchScanStatusCompleted = "COMPLETED"
	BatchScanStatusFailed    = "FAILED"
)

// ============================================================================
// BatchScanState — Batch Job Execution Tracking
// Source: §8.6, DDL: policy_mgmt.batch_scan_state
// Scale: ~10 rows/day (6 scan types × ~1-2 runs)
// UNIQUE constraint: (scan_type, scheduled_date) — idempotent upsert
// ============================================================================

// BatchScanState tracks one execution of a BatchStateScanWorkflow run.
// Written by RecordBatchScanResultActivity.
// ⚠️ INSERT must use ON CONFLICT DO NOTHING for idempotency.
type BatchScanState struct {
	// PK
	ScanID int64 `json:"scan_id" db:"scan_id"`

	// Job identity (UNIQUE: scan_type + scheduled_date)
	ScanType      string    `json:"scan_type"      db:"scan_type"`      // batch_scan_type enum
	ScheduledDate time.Time `json:"scheduled_date" db:"scheduled_date"` // DATE — the date this job was scheduled for

	// Execution timestamps
	StartedAt   *time.Time `json:"started_at,omitempty"   db:"started_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty" db:"completed_at"`

	// Results
	PoliciesScanned    int `json:"policies_scanned"    db:"policies_scanned"`
	TransitionsApplied int `json:"transitions_applied" db:"transitions_applied"`
	Errors             int `json:"errors"              db:"errors"`

	// Error details (JSONB)
	ErrorDetails json.RawMessage `json:"error_details,omitempty" db:"error_details"`

	// Job status
	Status          string `json:"status"           db:"status"`           // batch_scan_status enum
	DurationSeconds *int   `json:"duration_seconds,omitempty" db:"duration_seconds"`
}
