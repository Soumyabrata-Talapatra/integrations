// Package activities defines input/output types for all Temporal activities
// in the customer-nfs-service.
//
// WF-NFS-001: Aadhaar-based Address Change (15-min OTP window)
// WF-NFS-002: Manual Address Change (45-day CPC approval window)
// WF-NFS-003: Aadhaar-based Name Change (15-min OTP window)
// WF-NFS-004: Manual Name Change (45-day CPC approval window)
// WF-NFS-005: Withdrawal Request
//
// WORKFLOW STATE NOTE:
//
//	Every workflow calls StoreWorkflowStateInput to persist its Temporal
//	workflow_id + workflow_run_id into nfs.service_request. HTTP signal
//	endpoints use these values to send Temporal signals.
//
// SIGNAL TYPES:
//   - "otp_submitted"       → used by WF-NFS-001/003's OTP verify HTTP endpoint
//   - "approval_decision"   → used by WF-NFS-002/004's CPC approve HTTP endpoint
//   - "withdrawal_approved" → used by WF-NFS-005
package activities

import (
	"encoding/json"
	"time"
)

// ===========================================================================
// Common Types
// ===========================================================================

// StoreWorkflowStateInput persists workflow_id + workflow_run_id into
// nfs.service_request so HTTP signal endpoints can locate the workflow.
// Called at the start of every workflow. WF-NFS-001..WF-NFS-005.
type StoreWorkflowStateInput struct {
	RequestID     string `json:"request_id"`
	WorkflowID    string `json:"workflow_id"`
	WorkflowRunID string `json:"workflow_run_id"`
}

type StoreWorkflowStateResult struct {
	Stored bool `json:"stored"`
}

// UpdateStatusInput updates nfs.service_request.status + inserts
// status_transition_history + audit_log (batched in TX).
// BR-NFS-012
type UpdateStatusInput struct {
	RequestID        string  `json:"request_id"`
	NewStatus        string  `json:"new_status"`
	FromStatus       string  `json:"from_status"`
	UpdatedBy        string  `json:"updated_by"`
	Reason           *string `json:"reason,omitempty"`
	WorkflowSignalID *string `json:"workflow_signal_id,omitempty"`
	Channel          string  `json:"channel"`
	OfficeCode       *string `json:"office_code,omitempty"`
}

type UpdateStatusResult struct {
	RequestID string `json:"request_id"`
	NewStatus string `json:"new_status"`
}

// CreateAuditLogInput writes a standalone audit record. BR-NFS-016.
type CreateAuditLogInput struct {
	RequestID   string  `json:"request_id"`
	ActionType  string  `json:"action_type"`
	OldValue    *string `json:"old_value,omitempty"`
	NewValue    *string `json:"new_value,omitempty"`
	PerformedBy string  `json:"performed_by"`
	Channel     string  `json:"channel"`
	OfficeCode  *string `json:"office_code,omitempty"`
	Remarks     *string `json:"remarks,omitempty"`
}

type CreateAuditLogResult struct {
	AuditID string `json:"audit_id"`
}

// GenerateAckReceiptInput triggers DMS receipt generation. FR-NFS-004.
type GenerateAckReceiptInput struct {
	RequestID    string `json:"request_id"`
	TicketNumber string `json:"ticket_number"`
	CustomerID   int64  `json:"customer_id"`
	RequestType  string `json:"request_type"`
}

type GenerateAckReceiptResult struct {
	ReceiptDMSID   string `json:"receipt_dms_id"`
	ReceiptFileURL string `json:"receipt_file_url"`
}

// ===========================================================================
// Address Change Activity Types (WF-NFS-001, WF-NFS-002)
// ===========================================================================

// ValidateAddressRequestInput validates business rules before creating the
// service request. BR-NFS-001..006, VR-NFS-001..005, VR-NFS-013, VR-NFS-015.
type ValidateAddressRequestInput struct {
	CustomerID       int64   `json:"customer_id"`
	PolicyNumber     *string `json:"policy_number,omitempty"`
	AuthMethod       string  `json:"auth_method"`
	AddressUpdateFor string  `json:"address_update_for"`
	AddressType      string  `json:"address_type"`
	// New address fields
	NewAddressLine1 string  `json:"new_address_line1"`
	NewAddressLine2 *string `json:"new_address_line2,omitempty"`
	NewVillage      *string `json:"new_village,omitempty"`
	NewTaluka       *string `json:"new_taluka,omitempty"`
	NewCity         string  `json:"new_city"`
	NewDistrict     string  `json:"new_district"`
	NewState        string  `json:"new_state"`
	NewPincode      string  `json:"new_pincode"`
	Channel         string  `json:"channel"`
	OfficeCode      *string `json:"office_code,omitempty"`
}

type ValidateAddressRequestResult struct {
	IsValid          bool     `json:"is_valid"`
	ValidationErrors []string `json:"validation_errors,omitempty"`
	HasDuplicate     bool     `json:"has_duplicate"`
	ExistingTicket   string   `json:"existing_ticket,omitempty"`
}

// CreateAddressServiceRequestInput creates the service_request + address_change_detail
// + audit_log in a single batched TX. FR-NFS-001, BR-NFS-011, BR-NFS-016.
type CreateAddressServiceRequestInput struct {
	CustomerID       int64      `json:"customer_id"`
	PolicyNumber     *string    `json:"policy_number,omitempty"`
	AuthMethod       string     `json:"auth_method"`
	Channel          string     `json:"channel"`
	OfficeCode       *string    `json:"office_code,omitempty"`
	InitiatedBy      string     `json:"initiated_by"`
	AddressUpdateFor string     `json:"address_update_for"`
	AddressType      string     `json:"address_type"`
	OldAddressLine1  *string    `json:"old_address_line1,omitempty"`
	OldCity          *string    `json:"old_city,omitempty"`
	OldDistrict      *string    `json:"old_district,omitempty"`
	OldState         *string    `json:"old_state,omitempty"`
	OldPincode       *string    `json:"old_pincode,omitempty"`
	NewAddressLine1  string     `json:"new_address_line1"`
	NewAddressLine2  *string    `json:"new_address_line2,omitempty"`
	NewVillage       *string    `json:"new_village,omitempty"`
	NewTaluka        *string    `json:"new_taluka,omitempty"`
	NewCity          string     `json:"new_city"`
	NewDistrict      string     `json:"new_district"`
	NewState         string     `json:"new_state"`
	NewPincode       string     `json:"new_pincode"`
	SLADeadline      *time.Time `json:"sla_deadline,omitempty"`
}

type CreateAddressServiceRequestResult struct {
	RequestID    string `json:"request_id"`
	TicketNumber string `json:"ticket_number"`
	Status       string `json:"status"`
}

// AadhaarOTPRequestInput initiates an Aadhaar OTP request to UIDAI.
// WF-NFS-001, WF-NFS-003. Only allowed via Portal and Mobile (VR-NFS-016).
type AadhaarOTPRequestInput struct {
	RequestID  string `json:"request_id"`
	CustomerID int64  `json:"customer_id"`
	Channel    string `json:"channel"` // Must be Portal or Mobile
}

type AadhaarOTPRequestResult struct {
	OTPReferenceID string    `json:"otp_reference_id"`
	ExpiresAt      time.Time `json:"expires_at"` // 15 minutes
}

// AadhaarOTPVerifyInput verifies the submitted OTP against UIDAI.
// WF-NFS-001, WF-NFS-003.
type AadhaarOTPVerifyInput struct {
	RequestID      string `json:"request_id"`
	OTPReferenceID string `json:"otp_reference_id"`
	OTPSubmitted   string `json:"otp_submitted"`
	CustomerID     int64  `json:"customer_id"`
}

type AadhaarOTPVerifyResult struct {
	Verified     bool   `json:"verified"`
	AadhaarTxnID string `json:"aadhaar_txn_id,omitempty"` // UIDAI transaction ID
	FailReason   string `json:"fail_reason,omitempty"`
}

// UpdateAddressDataInput applies the verified new address to customer data.
// Called after Aadhaar OTP verify (WF-NFS-001) or CPC approval (WF-NFS-002).
// BR-NFS-001, BR-NFS-002
type UpdateAddressDataInput struct {
	RequestID   string `json:"request_id"`
	CustomerID  int64  `json:"customer_id"`
	AddressType string `json:"address_type"`
	UpdatedBy   string `json:"updated_by"`
	// If partial update succeeds, partial_processing_flag is set TRUE
}

type UpdateAddressDataResult struct {
	Updated      bool   `json:"updated"`
	NewVersionID string `json:"new_version_id"`
	// Address change details for Policy Management notification
	NewAddressLine1 *string `json:"new_address_line1,omitempty"`
	NewAddressLine2 *string `json:"new_address_line2,omitempty"`
	NewCity         *string `json:"new_city,omitempty"`
	NewDistrict     *string `json:"new_district,omitempty"`
	NewState        *string `json:"new_state,omitempty"`
	NewPincode      *string `json:"new_pincode,omitempty"`
	AddressType     *string `json:"address_type,omitempty"`
}

// UpdateMobileDataInput updates the customer's mobile number in the policy system.
// WF-NFS-006: Called after successful OTP verification.
type UpdateMobileDataInput struct {
	RequestID  string `json:"request_id"`
	CustomerID int64  `json:"customer_id"`
	UpdatedBy  string `json:"updated_by"`
}

type UpdateMobileDataResult struct {
	Updated      bool   `json:"updated"`
	NewVersionID string `json:"new_version_id"`
	// Mobile change details for Policy Management notification
	NewMobileNumber *string `json:"new_mobile_number,omitempty"`
}

// UpdateEmailDataInput updates the customer's email address in the policy system.
// WF-NFS-007: Called after successful OTP verification.
type UpdateEmailDataInput struct {
	RequestID  string `json:"request_id"`
	CustomerID int64  `json:"customer_id"`
	UpdatedBy  string `json:"updated_by"`
}

type UpdateEmailDataResult struct {
	Updated      bool   `json:"updated"`
	NewVersionID string `json:"new_version_id"`
	// Email change details for Policy Management notification
	NewEmail *string `json:"new_email,omitempty"`
}

// AssignToCPCInput assigns a request to the CPC work queue. FR-NFS-008.
type AssignToCPCInput struct {
	RequestID   string     `json:"request_id"`
	OfficeCode  *string    `json:"office_code,omitempty"`
	Priority    int        `json:"priority"`
	SLADeadline *time.Time `json:"sla_deadline,omitempty"`
	AssignedBy  string     `json:"assigned_by"`
}

type AssignToCPCResult struct {
	QueueID    string `json:"queue_id"`
	AssignedTo string `json:"assigned_to"`
}

// EscalateInput triggers SLA escalation notification. FR-NFS-010.
type EscalateInput struct {
	RequestID    string `json:"request_id"`
	TicketNumber string `json:"ticket_number"`
	SLABreached  bool   `json:"sla_breached"`
	EscalateTo   string `json:"escalate_to"`
}

type EscalateResult struct {
	Escalated bool `json:"escalated"`
}

// ===========================================================================
// Name Change Activity Types (WF-NFS-003, WF-NFS-004)
// ===========================================================================

// ValidateNameRequestInput validates name change request fields.
// BR-NFS-007..010, VR-NFS-006..008, VR-NFS-014, VR-NFS-015.
type ValidateNameRequestInput struct {
	CustomerID   int64   `json:"customer_id"`
	PolicyNumber *string `json:"policy_number,omitempty"`
	AuthMethod   string  `json:"auth_method"`
	// New name
	NewSalutation string  `json:"new_salutation"`
	NewFirstName  string  `json:"new_first_name"`
	NewMiddleName *string `json:"new_middle_name,omitempty"`
	NewLastName   string  `json:"new_last_name"`
	Channel       string  `json:"channel"`
}

type ValidateNameRequestResult struct {
	IsValid          bool     `json:"is_valid"`
	ValidationErrors []string `json:"validation_errors,omitempty"`
	HasDuplicate     bool     `json:"has_duplicate"`
	ExistingTicket   string   `json:"existing_ticket,omitempty"`
}

// CreateNameServiceRequestInput creates service_request + name_change_detail + audit_log.
// FR-NFS-007, BR-NFS-011, BR-NFS-016.
type CreateNameServiceRequestInput struct {
	CustomerID    int64      `json:"customer_id"`
	PolicyNumber  *string    `json:"policy_number,omitempty"`
	AuthMethod    string     `json:"auth_method"`
	Channel       string     `json:"channel"`
	OfficeCode    *string    `json:"office_code,omitempty"`
	InitiatedBy   string     `json:"initiated_by"`
	OldSalutation *string    `json:"old_salutation,omitempty"`
	OldFirstName  *string    `json:"old_first_name,omitempty"`
	OldMiddleName *string    `json:"old_middle_name,omitempty"`
	OldLastName   *string    `json:"old_last_name,omitempty"`
	NewSalutation string     `json:"new_salutation"`
	NewFirstName  string     `json:"new_first_name"`
	NewMiddleName *string    `json:"new_middle_name,omitempty"`
	NewLastName   string     `json:"new_last_name"`
	SLADeadline   *time.Time `json:"sla_deadline,omitempty"`
}

type CreateNameServiceRequestResult struct {
	RequestID    string `json:"request_id"`
	TicketNumber string `json:"ticket_number"`
	Status       string `json:"status"`
}

// UpdateNameDataInput applies the verified new name to customer data.
// Called after Aadhaar OTP verify (WF-NFS-003) or CPC approval (WF-NFS-004).
// BR-NFS-007, BR-NFS-008, BR-NFS-009
type UpdateNameDataInput struct {
	RequestID  string `json:"request_id"`
	CustomerID int64  `json:"customer_id"`
	UpdatedBy  string `json:"updated_by"`
	// Cross-policy name update count returned from Policy Admin Service
}

type UpdateNameDataResult struct {
	Updated          bool   `json:"updated"`
	NewVersionID     string `json:"new_version_id"`
	PoliciesAffected int    `json:"policies_affected"`
	// Name change details for Policy Management notification
	NewSalutation *string `json:"new_salutation,omitempty"`
	NewFirstName  *string `json:"new_first_name,omitempty"`
	NewMiddleName *string `json:"new_middle_name,omitempty"`
	NewLastName   *string `json:"new_last_name,omitempty"`
}

// ===========================================================================
// Withdrawal Activity Types (WF-NFS-005)
// ===========================================================================

// CheckWithdrawalEligibilityInput determines if the request can be withdrawn
// and whether it requires CPC approval (MANUAL) or can be auto-approved.
// BR-NFS-013.
type CheckWithdrawalEligibilityInput struct {
	RequestID string `json:"request_id"`
}

type CheckWithdrawalEligibilityResult struct {
	Eligible     bool   `json:"eligible"`
	ApprovalType string `json:"approval_type"` // "AUTO" or "MANUAL"
	RequestType  string `json:"request_type"`
	AuthMethod   string `json:"auth_method"`
	Reason       string `json:"reason"`
}

// ProcessWithdrawalInput processes a withdrawal request. BR-NFS-013.
type ProcessWithdrawalInput struct {
	RequestID        string  `json:"request_id"`
	WithdrawalID     string  `json:"withdrawal_id"`
	WithdrawalType   string  `json:"withdrawal_type"` // AUTO or MANUAL
	WithdrawalReason string  `json:"withdrawal_reason"`
	RequestedBy      string  `json:"requested_by"`
	ApprovedBy       *int64  `json:"approved_by,omitempty"`
}

type ProcessWithdrawalResult struct {
	Processed bool   `json:"processed"`
	NewStatus string `json:"new_status"`
}

// ─── HTTP → Workflow Signal Payloads ─────────────────────────────────────────
// These types are used by HTTP handlers to signal running Temporal workflows.
// They are defined in the activities package so both handlers and activities can import them
// without creating an import cycle.

// DocumentsSubmittedPayload is the payload sent by CORE-003 to the running WF-NFS-002 workflow.
// Signal name: "documents_submitted" (workflows.SignalDocumentsSubmitted).
// BR-NFS-002: signal transitions the workflow from document-upload wait to AssignToCPC phase.
type DocumentsSubmittedPayload struct {
	RequestID         string   `json:"request_id"`
	UploadedDocuments []string `json:"uploaded_documents"`
	SubmittedBy       string   `json:"submitted_by"`
}

// ===========================================================================
// Policy Management Notification Types
// ===========================================================================

// NotifyPMInput is the input for the NotifyPolicyManagement activity.
// Called at each terminal point (COMPLETED/REJECTED) in NFS workflows
// to signal affected PolicyLifecycleWorkflows in Policy Management.
type NotifyPMInput struct {
	RequestID     string          `json:"request_id"`
	CustomerID    int64           `json:"customer_id"`
	RequestType   string          `json:"request_type"`    // ADDRESS_CHANGE, NAME_CHANGE, MOBILE_CHANGE, EMAIL_CHANGE
	Outcome       string          `json:"outcome"`         // APPROVED, REJECTED
	ChangePayload json.RawMessage `json:"change_payload"`
}

// NotifyPMResult is the result of the NotifyPolicyManagement activity.
type NotifyPMResult struct {
	PoliciesNotified int  `json:"policies_notified"`
	Success          bool `json:"success"`
}
