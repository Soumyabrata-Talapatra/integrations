package handler

import (
	"mime/multipart"
)

// ============================================
// Voluntary Surrender Request DTOs
// ============================================

// ValidateEligibilityRequest represents the request to validate surrender eligibility
// Business Rules: BR-SUR-001, BR-SUR-002, BR-SUR-003, BR-SUR-004
// Validation Rules: VR-SUR-001, VR-SUR-002, VR-SUR-003, VR-SUR-004
type ValidateEligibilityRequest struct {
	PolicyID string `json:"policy_id" validate:"required"`
}

// CalculateSurrenderRequest represents the request to calculate surrender value
// Business Rules: BR-SUR-006, BR-SUR-007, BR-SUR-008, BR-SUR-009, BR-SUR-010, BR-SUR-011
type CalculateSurrenderRequest struct {
	PolicyID string `json:"policy_id" validate:"required"`
}

type IndexSurrenderRequest struct {
	PolicyNumber              string  `json:"policy_number" validate:"required"`
	Surrender_request_channel string  `json:"surrender_request_channel"`
	Indexing_office_id        int     `json:"indexing_office_id"`
	Cpc_office_id             int     `json:"cpc_office_id"`
	Created_by                int     `json:"created_by"`
	Modified_by               int     `json:"modified_by"`
	Remarks                   string  `json:"remarks"`
	Paidupvalue               float64 `json:"paidupvalue"`
	Bonus                     float64 `json:"bonusvalue"`
	Grossamount               float64 `json:"grossamount"`
	Loanprincipal             float64 `json:"loanprincipal"`
	Loaninterest              float64 `json:"loaninterest"`
	Surrenderfactor           float64 `json:"surrenderfactor"`
	Othercharges              float64 `json:"othercharges"`
	Surrendervalue            float64 `json:"surrendervalue"`
	Bonusrate                 float64 `json:"bonusrate"`
	Bonusamount               float64 `json:"bonusamount"`
	Sumassured                float64 `json:"sumassured"`
	Paid_to_date              string  `json:"paid_to_date"`
	Polissdate                string  `json:"polissdate"`
	Maturitydate              string  `json:"maturitydate"`
	Productcode               string  `json:"productcode"`
	Dob                       string  `json:"dob"`
	Unpaidprem                float64 `json:"unpaidprem"`
	Def                       float64 `json:"def"`
	Stage_name                string  `json:"stage_name"`
}

type SubmitDERequest struct {
	Surrender_request_id      string `json:"surrender_request_id" validate:"required"`
	Surrender_request_channel string `json:"surrender_request_channel"`
	Request_name              string `json:"request_name"`
	Current_stage_name        string `json:"current_stage_name"`
	Created_by                int    `json:"created_by"`
	Modified_by               int    `json:"modified_by"`
	Remarks                   string `json:"remarks"`
	Paymentmode               string `json:"paymentmode"`
	Bankname                  string `json:"bankname"`
	Micrcode                  string `json:"micrcode"`
	Accounttype               string `json:"accounttype"`
	Ifsccode                  string `json:"ifsccode"`
	Accountnumber             string `json:"accountnumber"`
	Accountholdername         string `json:"accountholdername"`
	Branchname                string `json:"branchname"`
	Banktype                  string `json:"banktype"`
	Ismicrvalidated           bool   `json:"ismicrvalidated" select:"ismicrvalidated"`
	Policybond                bool   `json:"Policybond" select:"policybond"`
	Lrrb                      bool   `json:"Lrrb" select:"lrrb"`
	Prb                       bool   `json:"Prb" select:"prb"`
	Pdo_certificate           bool   `json:"Pdo_certificate" select:"pdo_certificate"`
	Application               bool   `json:"Application" select:"application"`
	Idproof_insurant          bool   `json:"Idproof_insurant" select:"idproof_insurant"`
	Addressproof_insurant     bool   `json:"Addressproof_insurant" select:"addressproof_insurant"`
	Idproof_messenger         bool   `json:"Idproof_messenger" select:"idproof_messenger"`
	Addressproof_messenger    bool   `json:"Addressproof_messenger" select:"addressproof_messenger"`
	Account_details_proof     bool   `json:"Account_details_proof" select:"account_details_proof"`
	Cpc_office_id             int    `json:"cpc_office_id"`
	PolicyNumber              string `json:"policy_number"`
	Others                    bool   `json:"Others" select:"others"`
}

type SRIDDetailsRequest struct {
	Surrender_request_id string `uri:"surrender_request_id" validate:"required"`
}

type GetDEPendingRequest struct {
	Oid int `uri:"office_id" validate:"required"`
}

// ConfirmSurrenderRequest represents the request to confirm surrender
// Business Rule: BR-SUR-013
type ConfirmSurrenderRequest struct {
	PolicyID           string  `json:"policy_id" validate:"required"`
	DisbursementMethod string  `json:"disbursement_method" validate:"required,oneof=CASH CHEQUE"`
	Reason             *string `json:"reason" validate:"omitempty,max=500"`
}

// UploadDocumentRequest represents the multipart form data for document upload
// Validation Rules: VR-SUR-007, VR-SUR-008, VR-SUR-009
type UploadDocumentRequest struct {
	SurrenderRequestID string                `form:"surrender_request_id" validate:"required"`
	DocumentType       string                `form:"document_type" validate:"required,oneof=WRITTEN_CONSENT POLICY_BOND PREMIUM_RECEIPT_BOOK PAY_RECOVERY_CERTIFICATE LOAN_RECEIPT_BOOK LOAN_BOND INDEMNITY_BOND ASSIGNMENT_DEED DISCHARGE_RECEIPT"`
	File               *multipart.FileHeader `form:"file" validate:"required"`
}

// DocumentStatusParams represents query parameters for document status
type DocumentStatusParams struct {
	SurrenderRequestID string `form:"surrender_request_id" validate:"required"`
}

// SubmitForVerificationRequest represents the request to submit for verification
// Business Rule: BR-SUR-017
// Validation Rule: VR-SUR-010
type SubmitForVerificationRequest struct {
	SurrenderRequestID string `json:"surrender_request_id" validate:"required"`
}

// SurrenderStatusParams represents query parameters for surrender status
type SurrenderStatusParams struct {
	SurrenderRequestID string `form:"surrender_request_id" validate:"required"`
	IncludeDetails     bool   `form:"include_details"`
}

// ============================================
// Forced Surrender Request DTOs
// ============================================

// TriggerMonthlyEvaluationRequest triggers monthly evaluation
// Temporal Workflow: TEMP-001
type TriggerMonthlyEvaluationRequest struct {
	EvaluationMonth string `json:"evaluation_month" validate:"required"`
	EvaluationDate  string `json:"evaluation_date" validate:"required"`
}

// EvaluatePolicyRequest evaluates a specific policy
// Business Rules: BR-FS-002, BR-FS-003, BR-FS-004
// Temporal Workflow: TEMP-001
type EvaluatePolicyRequest struct {
	PolicyID string `json:"policy_id" validate:"required"`
}

// TriggerReminderRequest triggers a reminder
// Temporal Workflow: TEMP-005
type TriggerReminderRequest struct {
	PolicyID      string `json:"policy_id" validate:"required"`
	ReminderLevel string `json:"reminder_level" validate:"required,oneof=FIRST SECOND THIRD"`
}

// CreateForcedSurrenderRequest creates forced surrender request
// Business Rules: BR-FS-004, BR-FS-005, BR-FS-006
// Temporal Workflow: TEMP-003
type CreateForcedSurrenderRequest struct {
	PolicyID string `json:"policy_id" validate:"required"`
}

// AutoCompleteRequest auto-completes forced surrender
// Business Rules: BR-FS-007, BR-FS-008
// Temporal Workflow: TEMP-002
type AutoCompleteRequest struct {
	SurrenderRequestID string `json:"surrender_request_id" validate:"required"`
}

// ForwardToApprovalRequest forwards to approval queue
// Temporal Workflow: TEMP-002
type ForwardToApprovalRequest struct {
	SurrenderRequestID string `json:"surrender_request_id" validate:"required"`
}

// RevertStatusRequest reverts policy status
// Business Rule: BR-FS-018 (CRITICAL)
// Temporal Workflow: TEMP-004
type RevertStatusRequest struct {
	SurrenderRequestID string `json:"surrender_request_id" validate:"required"`
}

// ============================================
// Approval Workflow Request DTOs
// ============================================

// ApprovalQueueParams represents query parameters for approval queue
// Business Rules: BR-FS-013, BR-FS-016
// type ApprovalQueueParams struct {
// 	port.MetadataRequest
// 	OfficeCode string `form:"office_code" validate:"required"`
// 	Status     string `form:"status" validate:"omitempty,oneof=PENDING RESERVED IN_PROGRESS"`
// }

// ReserveRequestRequest reserves a surrender request
// Business Rule: BR-FS-016
type ReserveRequestRequest struct {
	TaskID string `json:"task_id" validate:"required"`
}

// RequestDetailParams represents query parameters for request details
type RequestDetailParams struct {
	SurrenderRequestID string `form:"surrender_request_id" validate:"required"`
}

// ApproveSurrenderRequest represents approval action
// Functional Requirement: FR-FS-007
// Temporal Workflow: TEMP-006
type ApproveSurrenderRequest struct {
	SurrenderRequestID string `json:"surrender_request_id" validate:"required"`
	Comments           string `json:"comments" validate:"required,min=10,max=500"`
	ApproverID         string `json:"approver_id" validate:"required"`
	ApproverUserID     string `json:"approver_user_id" validate:"required"`
	ApprovalComments   string `json:"approval_comments" validate:"required,min=10,max=500"`
}

// RejectSurrenderRequest represents rejection action
// Business Rules: BR-FS-007, BR-FS-012, BR-FS-018
type RejectSurrenderRequest struct {
	SurrenderRequestID string `json:"surrender_request_id" validate:"required"`
	ApproverUserID     string `json:"approver_user_id" validate:"required"`
	RejectionReason    string `json:"rejection_reason" validate:"required,min=10,max=500"`
	ApproverID         string `json:"approver_id" validate:"required"`
	RejectorUserID     string `json:"rejector_user_id" validate:"required"`
}

// ReleaseRequestRequest releases a reserved request
// Functional Requirement: FR-FS-009
type ReleaseRequestRequest struct {
	TaskID string `json:"task_id" validate:"required"`
}

// EscalateRequestRequest escalates a request
// Business Rule: BR-FS-013
type EscalateRequestRequest struct {
	TaskID           string `json:"task_id" validate:"required"`
	EscalateTo       string `json:"escalate_to" validate:"required"`
	EscalationReason string `json:"escalation_reason" validate:"required,min=10,max=500"`
}

// RecalculateRequest recalculates surrender value
// Business Rule: BR-FS-010
type RecalculateRequest struct {
	SurrenderRequestID  string `json:"surrender_request_id" validate:"required"`
	UserID              string `json:"user_id" validate:"required"`
	RecalculationReason string `json:"recalculation_reason" validate:"required,min=10,max=500"`
}

// ReserveTaskRequest reserves an approval task for processing
type ReserveTaskRequest struct {
	TaskID string `json:"task_id" validate:"required"`
	UserID string `json:"user_id" validate:"required"`
}

// ReleaseTaskRequest releases a reserved task
type ReleaseTaskRequest struct {
	TaskID string `json:"task_id" validate:"required"`
	UserID string `json:"user_id" validate:"required"`
}

// EscalateTaskRequest escalates an approval task
type EscalateTaskRequest struct {
	TaskID           string `json:"task_id" validate:"required"`
	EscalationReason string `json:"escalation_reason" validate:"required,min=10,max=500"`
	UserID           string `json:"user_id" validate:"required"`
	EscalatedBy      string `json:"escalated_by" validate:"required"`
	EscalateTo       string `json:"escalate_to" validate:"required"`
}

// SendReminderRequest sends forced surrender reminder
// Business Rule: BR-FSUR-002, BR-FSUR-003
type SendReminderRequest struct {
	PolicyID     string `json:"policy_id" validate:"required"`
	UnpaidMonths int    `json:"unpaid_months" validate:"required,min=6"`
}

// CreatePaymentWindowRequest creates forced surrender payment window
// Business Rule: BR-FS-008
type CreatePaymentWindowRequest struct {
	SurrenderRequestID string `json:"surrender_request_id" validate:"required"`
	ReminderID         string `json:"reminder_id" validate:"required"`
	WindowStartDate    string `json:"window_start_date" validate:"required"`
	WindowEndDate      string `json:"window_end_date" validate:"required"`
}

// RecordPaymentRequest records payment for forced surrender
// Business Rule: BR-FS-007
type RecordPaymentRequest struct {
	PaymentWindowID  string  `json:"payment_window_id" validate:"required"`
	Amount           float64 `json:"amount" validate:"required,gt=0"`
	PaymentReference string  `json:"payment_reference" validate:"required"`
	PaymentDate      string  `json:"payment_date" validate:"required"`
}

// InitiateForcedSurrenderRequest initiates forced surrender process
// Business Rule: BR-FS-004
type InitiateForcedSurrenderRequest struct {
	PolicyID string `json:"policy_id" validate:"required"`
	Reason   string `json:"reason" validate:"omitempty,max=500"`
}

// ScheduleBatchRequest schedules batch processing
// Temporal Workflow: TEMP-001
type ScheduleBatchRequest struct {
	ProcessType string `json:"process_type" validate:"required,oneof=MONTHLY_EVALUATION REMINDER_BATCH AUTO_COMPLETION"`
	ScheduledAt string `json:"scheduled_at" validate:"required"`
}

// SearchSurrenderRequest searches surrender requests with filters
type SearchSurrenderRequest struct {
	RequestType string `form:"request_type" validate:"omitempty,oneof=VOLUNTARY FORCED"`
	Status      string `form:"status" validate:"omitempty"`
	TaskStatus  string `form:"task_status" validate:"omitempty"`
	OfficeCode  string `form:"office_code" validate:"omitempty"`
	PolicyID    string `form:"policy_id" validate:"omitempty,uuid"`
	FromDate    string `form:"from_date" validate:"omitempty"`
	ToDate      string `form:"to_date" validate:"omitempty"`
	Page        int    `form:"page" validate:"omitempty,min=1"`
	Limit       int    `form:"limit" validate:"omitempty,min=1,max=100"`
}
