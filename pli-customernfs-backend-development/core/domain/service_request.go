package domain

import "time"

// ServiceRequest is the central entity tracking every customer NFS request.
// Entity ID: E-1 (per Customer_NFS_Service_Analysis.md Section 7.1)
//
// Business Rules applied:
//   - BR-NFS-011: Ticket number format NFS-{TYPE}-{YYYYMMDD}-{SEQ6}
//   - BR-NFS-012: 8 status states, 17 valid state transitions
//   - BR-NFS-013: Withdrawal eligibility via status check
//   - BR-NFS-015: Aadhaar auto-approval restricted to Portal/Mobile
//   - BR-NFS-016: All mutations logged in audit_log (INSERT-only)
//
// WORKFLOW STATE: workflow_id and workflow_run_id store the Temporal workflow
// references so HTTP handlers can signal running workflows (e.g., approval
// decisions, OTP submission). This enables the long-running (15-30 day)
// approval workflows to receive signals from CPC API calls.
// See: Section 9 (Temporal Workflows), WF-NFS-001 to WF-NFS-005
//
// BATCH NOTE: ServiceRequest creation always batches 3 inserts in one TX:
//   1. nfs.service_request (this entity)
//   2. nfs.address_change_detail OR nfs.name_change_detail (type detail)
//   3. nfs.audit_log (CREATED action, BR-NFS-016)
// See: repo/postgres/service_request.go CreateWithDetail()
type ServiceRequest struct {
	RequestID                string     `json:"request_id" db:"request_id"`
	TicketNumber             string     `json:"ticket_number" db:"ticket_number"`
	CustomerID               int64      `json:"customer_id" db:"customer_id"`
	PolicyNumber             *string    `json:"policy_number,omitempty" db:"policy_number"`
	RequestType              string     `json:"request_type" db:"request_type"`
	AuthMethod               string     `json:"auth_method" db:"auth_method"`
	Status                   string     `json:"status" db:"status"`
	PreviousStatus           *string    `json:"previous_status,omitempty" db:"previous_status"`
	Channel                  string     `json:"channel" db:"channel"`
	OfficeCode               *string    `json:"office_code,omitempty" db:"office_code"`
	InitiatedBy              string     `json:"initiated_by" db:"initiated_by"`
	AssignedTo               *string    `json:"assigned_to,omitempty" db:"assigned_to"`
	ApprovedBy               *string    `json:"approved_by,omitempty" db:"approved_by"`
	CreatedAt                time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt                time.Time  `json:"updated_at" db:"updated_at"`
	CompletedAt              *time.Time `json:"completed_at,omitempty" db:"completed_at"`
	ApprovalDate             *time.Time `json:"approval_date,omitempty" db:"approval_date"`
	SLADeadline              *time.Time `json:"sla_deadline,omitempty" db:"sla_deadline"`
	SLABreached              bool       `json:"sla_breached" db:"sla_breached"`
	RejectionReason          *string    `json:"rejection_reason,omitempty" db:"rejection_reason"`
	PartialProcessingFlag    bool       `json:"partial_processing_flag" db:"partial_processing_flag"`
	GuardianApprovalRequired *bool      `json:"guardian_approval_required,omitempty" db:"guardian_approval_required"`
	// WorkflowID and WorkflowRunID store Temporal workflow references.
	// These are set when the Temporal workflow is started and used by
	// HTTP API handlers to signal (approve/reject/withdraw) the workflow.
	// Critical for: BR-NFS-001 (OTP signal), CPC approve/reject signals
	WorkflowID    *string    `json:"workflow_id,omitempty" db:"workflow_id"`
	WorkflowRunID *string    `json:"workflow_run_id,omitempty" db:"workflow_run_id"`
	CreatedBy     string     `json:"created_by" db:"created_by"`
	UpdatedBy     *string    `json:"updated_by,omitempty" db:"updated_by"`
	DeletedAt     *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
	Version       int        `json:"version" db:"version"`
	Metadata      *string    `json:"metadata,omitempty" db:"metadata"`
}

// StatusTransitionHistory tracks all status transitions for audit purposes.
// Entity from Section 7 (Service Request Status Machine, BR-NFS-012)
//
// BATCH NOTE: Status transitions are always written in batch with audit_log:
//   1. nfs.service_request (UPDATE status)
//   2. nfs.status_transition_history (INSERT transition record)
//   3. nfs.audit_log (INSERT STATUS_CHANGE action)
// See: repo/postgres/service_request.go UpdateStatus()
type StatusTransitionHistory struct {
	TransitionID     string    `json:"transition_id" db:"transition_id"`
	RequestID        string    `json:"request_id" db:"request_id"`
	FromStatus       *string   `json:"from_status,omitempty" db:"from_status"`
	ToStatus         string    `json:"to_status" db:"to_status"`
	TransitionReason *string   `json:"transition_reason,omitempty" db:"transition_reason"`
	TransitionedBy   string    `json:"transitioned_by" db:"transitioned_by"`
	TransitionedAt   time.Time `json:"transitioned_at" db:"transitioned_at"`
	WorkflowSignalID *string   `json:"workflow_signal_id,omitempty" db:"workflow_signal_id"`
}

// TicketSequence tracks daily sequence numbers for ticket ID generation.
// Used by BR-NFS-011: NFS-{TYPE}-{YYYYMMDD}-{SEQ6}
//
// BATCH NOTE: Ticket generation uses a batch select-then-increment to
// prevent race conditions under concurrent request creation.
// See: repo/postgres/service_request.go GenerateTicketNumber()
type TicketSequence struct {
	SequenceDate  time.Time `json:"sequence_date" db:"sequence_date"`
	RequestType   string    `json:"request_type" db:"request_type"`
	SequenceValue int       `json:"sequence_value" db:"sequence_value"`
}

// WithdrawalRequest tracks withdrawal requests and their approval status.
// Entity from FR-NFS-007, BR-NFS-013 (Withdrawal Eligibility)
//
// BATCH NOTE: Withdrawal creation batches with service_request status update +
// audit_log INSERT in a single transaction.
// See: repo/postgres/withdrawal.go Create()
type WithdrawalRequest struct {
	WithdrawalID     string     `json:"withdrawal_id" db:"withdrawal_id"`
	RequestID        string     `json:"request_id" db:"request_id"`
	WithdrawalReason string     `json:"withdrawal_reason" db:"withdrawal_reason"`
	WithdrawalType   string     `json:"withdrawal_type" db:"withdrawal_type"`
	Status           string     `json:"status" db:"status"`
	ApprovedBy       *string    `json:"approved_by,omitempty" db:"approved_by"`
	ApprovedAt       *time.Time `json:"approved_at,omitempty" db:"approved_at"`
	ApprovalRemarks  *string    `json:"approval_remarks,omitempty" db:"approval_remarks"`
	RequestedBy      string     `json:"requested_by" db:"requested_by"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`
	CreatedBy        string     `json:"created_by" db:"created_by"`
	UpdatedBy        *string    `json:"updated_by,omitempty" db:"updated_by"`
	Version          int        `json:"version" db:"version"`
}
