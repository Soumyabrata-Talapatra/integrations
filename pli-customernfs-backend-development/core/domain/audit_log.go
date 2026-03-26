package domain

import "time"

// AuditLog represents an immutable audit trail record for all NFS operations.
// Entity ID: E-6 (per Customer_NFS_Service_Analysis.md Section 7.6 and Section 20)
//
// Business Rules applied:
//   - BR-NFS-016: INSERT-only — NO UPDATE or DELETE permitted (regulatory compliance)
//   - FR-NFS-008: Every action logged with user_id, timestamp UTC, action_type,
//                 old_value, new_value, IP address, channel
//   - Retention: 10 years minimum (C4 constraint)
//   - Partitioning: yearly partitions for query performance (Section 20.2)
//
// Action Types (nfs.audit_action_type_enum):
//   CREATED, STATUS_CHANGE, DOCUMENT_UPLOAD, DOCUMENT_VERIFIED, ASSIGNED,
//   APPROVED, REJECTED, WITHDRAWN, COMMENT_ADDED, MISSING_DOC_REQUESTED,
//   COMPENSATION_ADDRESS_RESTORE, COMPENSATION_NAME_RESTORE
//
// BATCH NOTE: AuditLog inserts are ALWAYS batched with the primary operation:
//   - Service request creation → batch with service_request INSERT
//   - Status transition → batch with service_request UPDATE + status_transition_history INSERT
//   - Document upload → batch with document_upload INSERT
//   - Approval/Rejection → batch with service_request UPDATE
// This guarantees atomic audit capture — no operation goes unlogged.
// Per BR-NFS-016, audit_log INSERT is part of every multi-step TX batch.
// The audit_log table is partitioned by year (nfs.audit_log_2026 etc.)
// so queries by date range are partition-pruned for performance.
// See: repo/postgres/audit_log.go Create() and all repo methods
type AuditLog struct {
	AuditID    string  `json:"audit_id" db:"audit_id"`
	RequestID  string  `json:"request_id" db:"request_id"`
	ActionType string  `json:"action_type" db:"action_type"`
	OldValue   *string `json:"old_value,omitempty" db:"old_value"`
	// NewValueJSON holds the JSON-encoded new state snapshot (maps to new_value column).
	NewValueJSON string `json:"new_value,omitempty" db:"new_value"`
	// PerformedByID is the user ID of the actor (maps to performed_by column).
	PerformedByID string    `json:"performed_by" db:"performed_by"`
	PerformedAt   time.Time `json:"performed_at" db:"performed_at"`
	IPAddress     *string   `json:"ip_address,omitempty" db:"ip_address"`
	Channel       *string   `json:"channel,omitempty" db:"channel"`
	OfficeCode    *string   `json:"office_code,omitempty" db:"office_code"`
	// Notes maps to the remarks column in the DB.
	Notes        *string `json:"remarks,omitempty" db:"remarks"`
	ErrorCode    *string `json:"error_code,omitempty" db:"error_code"`
	ErrorMessage *string `json:"error_message,omitempty" db:"error_message"`
}

// CPCWorkQueue represents items in the CPC work queue.
// Used by FR-NFS-011 (CPC User Work Queue and Dashboard).
//
// BATCH NOTE: CPC queue item creation is always batched with:
//   1. service_request assignment (UPDATE assigned_to)
//   2. cpc_work_queue INSERT
//   3. audit_log INSERT (ASSIGNED action)
// See: repo/postgres/cpc_work_queue.go AssignToQueue()
type CPCWorkQueue struct {
	QueueID     string     `json:"queue_id" db:"queue_id"`
	RequestID   string     `json:"request_id" db:"request_id"`
	AssignedTo  *string    `json:"assigned_to,omitempty" db:"assigned_to"`
	AssignedAt  time.Time  `json:"assigned_at" db:"assigned_at"`
	Priority    int        `json:"priority" db:"priority"`
	SLADeadline *time.Time `json:"sla_deadline,omitempty" db:"sla_deadline"`
	SLAStatus   string     `json:"sla_status" db:"sla_status"`
	QueueStatus string     `json:"queue_status" db:"queue_status"`
	PickedUpBy  *string    `json:"picked_up_by,omitempty" db:"picked_up_by"`
	PickedUpAt  *time.Time `json:"picked_up_at,omitempty" db:"picked_up_at"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at" db:"updated_at"`
}
