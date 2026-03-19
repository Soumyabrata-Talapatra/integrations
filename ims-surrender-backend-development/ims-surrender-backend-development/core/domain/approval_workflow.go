package domain

import (
	"time"

	"github.com/google/uuid"
)

// ApprovalTaskStatus represents the status of an approval task
type ApprovalTaskStatus string

const (
	ApprovalTaskStatusPending    ApprovalTaskStatus = "PENDING"
	ApprovalTaskStatusReserved   ApprovalTaskStatus = "RESERVED"
	ApprovalTaskStatusInProgress ApprovalTaskStatus = "IN_PROGRESS"
	ApprovalTaskStatusCompleted  ApprovalTaskStatus = "COMPLETED"
	ApprovalTaskStatusEscalated  ApprovalTaskStatus = "ESCALATED"
)

// TaskStatus is an alias for ApprovalTaskStatus for convenience
type TaskStatus = ApprovalTaskStatus

// TaskOutcome represents the outcome of a completed approval task
type TaskOutcome string

const (
	TaskOutcomeApproved  TaskOutcome = "APPROVED"
	TaskOutcomeRejected  TaskOutcome = "REJECTED"
	TaskOutcomeEscalated TaskOutcome = "ESCALATED"
	TaskOutcomeWithdrawn TaskOutcome = "WITHDRAWN"
)

// ApprovalTaskPriority represents the priority level of an approval task
type ApprovalTaskPriority string

const (
	ApprovalTaskPriorityLow      ApprovalTaskPriority = "LOW"
	ApprovalTaskPriorityMedium   ApprovalTaskPriority = "MEDIUM"
	ApprovalTaskPriorityHigh     ApprovalTaskPriority = "HIGH"
	ApprovalTaskPriorityCritical ApprovalTaskPriority = "CRITICAL"
)

// ApprovalWorkflowTask represents a task in the approval workflow
// Table: approval_workflow_tasks
// Business Rules: BR-FS-013, BR-FS-016
type ApprovalWorkflowTask struct {
	ID                   uuid.UUID              `json:"id" db:"id"`
	SurrenderRequestID   uuid.UUID              `json:"surrender_request_id" db:"surrender_request_id"`
	TaskNumber           string                 `json:"task_number" db:"task_number"`
	OfficeCode           string                 `json:"office_code" db:"office_code"`
	AssignedTo           *uuid.UUID             `json:"assigned_to" db:"assigned_to"`
	Status               ApprovalTaskStatus     `json:"status" db:"status"`
	Reserved             bool                   `json:"reserved" db:"reserved"`
	ReservedAt           *time.Time             `json:"reserved_at" db:"reserved_at"`
	ReservedBy           *uuid.UUID             `json:"reserved_by" db:"reserved_by"`
	ReservationExpiresAt *time.Time             `json:"reservation_expires_at" db:"reservation_expires_at"`
	Priority             ApprovalTaskPriority   `json:"priority" db:"priority"`
	CreatedAt            time.Time              `json:"created_at" db:"created_at"`
	CompletedAt          *time.Time             `json:"completed_at" db:"completed_at"`
	CompletedBy          *uuid.UUID             `json:"completed_by" db:"completed_by"`
	Escalated            bool                   `json:"escalated" db:"escalated"`
	EscalatedTo          *uuid.UUID             `json:"escalated_to" db:"escalated_to"`
	EscalatedAt          *time.Time             `json:"escalated_at" db:"escalated_at"`
	EscalationReason     *string                `json:"escalation_reason" db:"escalation_reason"`
	Metadata             map[string]interface{} `json:"metadata" db:"metadata"`
}

// SurrenderRequestHistory represents an audit trail entry
// Table: surrender_request_history
// Functional Requirement: FR-SUR-008
type SurrenderRequestHistory struct {
	ID                 uuid.UUID              `json:"id" db:"id"`
	SurrenderRequestID uuid.UUID              `json:"surrender_request_id" db:"surrender_request_id"`
	ChangedBy          uuid.UUID              `json:"changed_by" db:"changed_by"`
	ChangedAt          time.Time              `json:"changed_at" db:"changed_at"`
	OldStatus          *SurrenderStatus       `json:"old_status" db:"old_status"`
	NewStatus          *SurrenderStatus       `json:"new_status" db:"new_status"`
	ChangeType         string                 `json:"change_type" db:"change_type"`
	ChangeDetails      map[string]interface{} `json:"change_details" db:"change_details"`
	Comments           *string                `json:"comments" db:"comments"`
	IPAddress          *string                `json:"ip_address" db:"ip_address"`
	UserAgent          *string                `json:"user_agent" db:"user_agent"`
}

// Change types for history
const (
	ChangeTypeRequestCreated     = "REQUEST_CREATED"
	ChangeTypeStatusChange       = "STATUS_CHANGE"
	ChangeTypeDocumentUploaded   = "DOCUMENT_UPLOADED"
	ChangeTypeDocumentVerified   = "DOCUMENT_VERIFIED"
	ChangeTypeApproved           = "APPROVED"
	ChangeTypeRejected           = "REJECTED"
	ChangeTypeCalculationUpdated = "CALCULATION_UPDATED"
	ChangeTypeEscalated          = "ESCALATED"
)
