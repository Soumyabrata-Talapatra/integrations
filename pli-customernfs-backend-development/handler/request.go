package handler

import (
	"customer-nfs-service/core/domain"
)

// ─────────────────────────────────────────────────────────────────────────────
// Address Change Request DTOs
// Phase 1 — CORE-001..004
// ─────────────────────────────────────────────────────────────────────────────

// InitiateAddressChangeRequest is the request body for CORE-001.
// FR-NFS-001 (Aadhaar path), FR-NFS-002 (Manual path), FR-NFS-010 (Multi-channel)
// BR-NFS-001 (AADHAAR → immediate), BR-NFS-002 (MANUAL → PENDING_APPROVAL)
// BR-NFS-005 (only INSURED role handled here), BR-NFS-011 (ticket number generation)
// BR-NFS-015 (AADHAAR only from Portal/Mobile)
// VR-NFS-001..005 (address field validations), VR-NFS-011..013, VR-NFS-015 (duplicate check)
// WF-NFS-001 (Aadhaar), WF-NFS-002 (Manual)
type InitiateAddressChangeRequest struct {
	// CustomerID is the unique identifier of the customer initiating the address change.
	// VR-NFS-015: used for duplicate check.
	CustomerID int64 `json:"customer_id" validate:"required,gt=0"`
	// PolicyNumber optionally restricts the address change to one policy.
	PolicyNumber *string `json:"policy_number,omitempty" validate:"omitempty,min=1,max=30"`

	// AuthMethod determines the verification path.
	// BR-NFS-001: AADHAAR → immediate; BR-NFS-002: MANUAL → CPC approval.
	// VR-NFS-011: must be AADHAAR or MANUAL.
	AuthMethod string `json:"auth_method" validate:"required,oneof=AADHAAR MANUAL"`

	// AddressUpdateFor indicates whose address is being changed.
	// BR-NFS-005: only INSURED is handled; others are routed to Policy Service.
	// VR-NFS-013: must be one of the allowed roles.
	AddressUpdateFor string `json:"address_update_for" validate:"required,oneof=INSURED PROPOSER ASSIGNEE TRUSTEE"`

	// AddressType specifies which address record to update.
	AddressType string `json:"address_type" validate:"required,oneof=COMMUNICATION PERMANENT OFFICIAL"`

	// New address fields — VR-NFS-001..005
	NewAddressLine1 string  `json:"new_address_line1" validate:"required,min=1,max=200"`
	NewAddressLine2 *string `json:"new_address_line2,omitempty" validate:"omitempty,max=200"`
	NewVillage      *string `json:"new_village,omitempty" validate:"omitempty,max=100"`
	NewTaluka       *string `json:"new_taluka,omitempty" validate:"omitempty,max=100"`
	// VR-NFS-004: city is required.
	NewCity string `json:"new_city" validate:"required,min=1,max=100"`
	// VR-NFS-005: district is required.
	NewDistrict string `json:"new_district" validate:"required,min=1,max=100"`
	// VR-NFS-002: state must match pincode.
	NewState string `json:"new_state" validate:"required,min=1,max=50"`
	// VR-NFS-001: pincode must be 6 numeric digits; BR-NFS-006: must match state.
	NewPincode string `json:"new_pincode" validate:"required,len=6,numeric"`

	// Channel is the originating channel.
	// VR-NFS-012: must be one of the configured channels.
	// BR-NFS-015: AADHAAR only allowed from Portal and Mobile.
	Channel string `json:"channel" validate:"required,oneof=Portal Mobile PostOffice CallCenter AgentPortal"`

	// OfficeCode is the office code for the servicing post office/circle (optional).
	OfficeCode *string `json:"office_code,omitempty" validate:"omitempty,max=20"`
}

// ToDomain converts the HTTP request into the domain ServiceRequest + AddressChangeDetail pair.
func (r InitiateAddressChangeRequest) ToDomain() (domain.ServiceRequest, domain.AddressChangeDetail) {
	sr := domain.ServiceRequest{
		CustomerID:   r.CustomerID,
		PolicyNumber: r.PolicyNumber,
		RequestType:  "ADDRESS_CHANGE",
		AuthMethod:   r.AuthMethod,
		Channel:      r.Channel,
		OfficeCode:   r.OfficeCode,
		Status:       "CREATED",
	}

	addr := domain.AddressChangeDetail{
		AddressUpdateFor: r.AddressUpdateFor, // ← add
		AddressType:      r.AddressType,
		NewAddressLine1:  r.NewAddressLine1, // add this
		NewAddressLine2:  r.NewAddressLine2, // add this too
		NewVillage:       r.NewVillage,
		NewTaluka:        r.NewTaluka,
		NewCity:          r.NewCity,
		NewDistrict:      r.NewDistrict,
		NewState:         r.NewState,
		NewPincode:       r.NewPincode,
	}

	return sr, addr
}

// VerifyOTPRequest is the request body + path param for CORE-002.
// FR-NFS-001: Aadhaar OTP verification step.
// BR-NFS-001: successful OTP → status immediately COMPLETED.
// WF-NFS-001: signal "otp_submitted" sent to the running Temporal workflow.
type VerifyOTPRequest struct {
	// RequestID is the path parameter identifying the service request.
	RequestID string `uri:"request_id" validate:"required,min=1,max=50"`

	// OTPCode is the 6-digit numeric OTP entered by the customer.
	OTPCode string `json:"otp_code" validate:"required,len=6,numeric"`

	// TxnID is the Aadhaar transaction ID returned by KYC Service when OTP was sent.
	TxnID string `json:"txn_id" validate:"required,min=1,max=100"`
}

// SubmitAddressChangeRequest is the request body + path param for CORE-003.
// FR-NFS-002: Manual path — finalise submission after document upload.
// BR-NFS-002: moves request to PENDING_APPROVAL, SLA clock starts.
// BR-NFS-012: valid transition CREATED → PENDING_APPROVAL.
// VR-NFS-009 (file size), VR-NFS-010 (mime type) — validated during DM upload phase.
// WF-NFS-002: workflow transitions to wait for approval_decision signal.
type SubmitAddressChangeRequest struct {
	// RequestID is the path parameter identifying the service request.
	RequestID string `uri:"request_id" validate:"required,min=1,max=50"`

	// UploadedDocuments lists document IDs already uploaded via the DM API.
	// At least one document is required (Address Proof is mandatory per ERR-NFS-ANC-004).
	UploadedDocuments []string `json:"uploaded_documents" validate:"required,min=1,dive,min=1,max=50"`

	// Declaration confirms the customer has acknowledged the declaration statement.
	Declaration bool `json:"declaration" validate:"required"`

	// SubmittedBy is the user/agent ID submitting the request.
	// SubmittedBy string `json:"submitted_by,omitempty" validate:"omitempty,min=1,max=50"`
	SubmittedBy string `json:"submitted_by" validate:"required,min=1,max=50"`
}

// ApproveAddressChangeRequest is the request body + path param for CORE-004.
// FR-NFS-002, FR-NFS-006, FR-NFS-008, FR-NFS-011.
// BR-NFS-002: CPC approval required for manual path.
// BR-NFS-003: on APPROVE → address propagated to downstream services.
// BR-NFS-004: address versioning — new version created on APPROVE.
// BR-NFS-012: IN_PROGRESS → COMPLETED / REJECTED / PENDING_DOCUMENTS.
// BR-NFS-016: audit log created (action_type = APPROVED / REJECTED).
// WF-NFS-002: signal "approval_decision" sent to the running workflow.
// ERR-NFS-SR-004: only assigned CPC or supervisory CPC can act.
// ERR-NFS-SR-005: rejection_reason is mandatory when decision=REJECT.
type ApproveAddressChangeRequest struct {
	// RequestID is the path parameter identifying the service request.
	RequestID string `uri:"request_id" validate:"required,min=1,max=50"`

	// Decision is the CPC decision on the address change request.
	// BR-NFS-012: must be APPROVE, REJECT, or SEND_BACK.
	Decision string `json:"decision" validate:"required,oneof=APPROVE REJECT SEND_BACK"`

	// ApprovedBy is the CPC user ID submitting the decision.
	ApprovedBy string `json:"approved_by" validate:"required,min=1,max=50"`

	// Remarks are optional general remarks from the CPC officer.
	Remarks *string `json:"remarks,omitempty" validate:"omitempty,max=1000"`

	// RejectionReason is mandatory when Decision=REJECT (ERR-NFS-SR-005).
	RejectionReason *string `json:"rejection_reason,omitempty" validate:"omitempty,max=1000"`

	// MissingDocuments lists the document types the CPC is requesting when Decision=SEND_BACK.
	MissingDocuments []string `json:"missing_documents,omitempty" validate:"omitempty,dive,min=1,max=100"`
}

// ─────────────────────────────────────────────────────────────────────────────
// Name Change Request DTOs
// Phase 2 — CORE-005..010
// ─────────────────────────────────────────────────────────────────────────────

// NewNamePayload carries the desired new name for a manual name change request.
// VR-NFS-006: first_name max 100 chars.
// VR-NFS-007: last_name max 100 chars.
// VR-NFS-008: salutation ∈ {Mr, Mrs, Ms, Shri, Smt, Dr}.
// BR-NFS-010 / VR-NFS-014: date_of_birth field intentionally absent.
type NewNamePayload struct {
	Salutation string  `json:"salutation" validate:"required,oneof=Mr Mrs Ms Shri Smt Dr"` // VR-NFS-008
	FirstName  string  `json:"first_name" validate:"required,max=100"`                     // VR-NFS-006
	MiddleName *string `json:"middle_name,omitempty"`
	LastName   string  `json:"last_name" validate:"required,max=100"` // VR-NFS-007
}

// InitiateNameChangeRequest is the request payload for CORE-005.
// BR-NFS-010 / VR-NFS-014: date_of_birth is intentionally absent.
// VR-NFS-015: Duplicate check for pending NAME_CHANGE request (handled in activity).
// BR-NFS-015: AADHAAR only from Portal and Mobile channels.
// WF-NFS-003 (AADHAAR) / WF-NFS-004 (MANUAL)
type InitiateNameChangeRequest struct {
	CustomerID   int64           `json:"customer_id" validate:"required,gt=0"`
	PolicyNumber *string         `json:"policy_number,omitempty"`
	AuthMethod   string          `json:"auth_method" validate:"required,oneof=AADHAAR MANUAL"`
	NewName      *NewNamePayload `json:"new_name,omitempty"` // required for MANUAL; ignored for AADHAAR
	// Channel      string          `json:"channel" validate:"required"`
	Channel string `json:"channel" validate:"required,oneof=Portal Mobile PostOffice CallCenter AgentPortal"`

	OfficeCode *string `json:"office_code,omitempty"`
}

// ToDomain converts the request to domain ServiceRequest + NameChangeDetail.
// BR-NFS-010: DOB is never set here.
func (r *InitiateNameChangeRequest) ToDomain(ticketNumber string, slaDeadline domain.SLADeadline) (*domain.ServiceRequest, *domain.NameChangeDetail) {
	sr := domain.ServiceRequest{
		CustomerID:   r.CustomerID,
		PolicyNumber: r.PolicyNumber,
		RequestType:  "NAME_CHANGE",
		TicketNumber: ticketNumber,
		Status:       "CREATED",
		AuthMethod:   r.AuthMethod,
		Channel:      r.Channel,
	}
	if r.OfficeCode != nil {
		sr.OfficeCode = r.OfficeCode
	}

	detail := domain.NameChangeDetail{
		RequestID:        sr.RequestID, // ❌ sr.RequestID is empty here — UUID not yet generated
		PoliciesAffected: 0,
	}
	// TODO Fix: Set RequestID after the DB insert returns the generated ID, not in ToDomain().

	if r.AuthMethod == "MANUAL" && r.NewName != nil {
		detail.NewSalutation = &r.NewName.Salutation
		detail.NewFirstName = &r.NewName.FirstName
		detail.NewMiddleName = r.NewName.MiddleName
		detail.NewLastName = &r.NewName.LastName
	}

	return &sr, &detail
}

// VerifyNameOTPRequest is the request payload for CORE-006.
// Signals WF-NFS-003 with "otp_submitted" signal.
// FR-NFS-004: Aadhaar-based name change — OTP step.
type VerifyNameOTPRequest struct {
	RequestID string `uri:"request_id" validate:"required,uuid"`
	OTPCode   string `json:"otp_code" validate:"required,len=6,numeric"`
	TxnID     string `json:"txn_id" validate:"required"` // from UIDAI OTP send response
}

// SubmitNameChangeRequest is the request payload for CORE-007.
// Signals WF-NFS-004 with "documents_submitted" signal.
// BR-NFS-008: At least ONE of the following document types is required.
// FR-NFS-005: Manual name change — document submission step.
type SubmitNameChangeRequest struct {
	RequestID         string   `uri:"request_id" validate:"required,uuid"`
	UploadedDocuments []string `json:"uploaded_documents" validate:"required,min=1,dive,uuid"` // document_upload IDs BR-NFS-008
	Declaration       bool     `json:"declaration" validate:"required"`
	SubmittedBy       string   `json:"submitted_by" validate:"required"`
}

// Validate checks BR-NFS-008 — at least ONE required document type must be present.--duplicate
// func (r *SubmitNameChangeRequest) Validate() error {
// 	if len(r.UploadedDocuments) == 0 {
// 		return fmt.Errorf("at least one document required (BR-NFS-008)")
// 	}
// 	if !r.Declaration {
// 		return fmt.Errorf("declaration must be accepted to submit documents")
// 	}
// 	return nil
// }

// ApproveNameChangeRequest is the request payload for CORE-008.
// Signals WF-NFS-004 with "approval_decision" signal.
// BR-NFS-009: On APPROVE, UpdateNameData activity publishes customer.name.updated event.
type ApproveNameChangeRequest struct {
	RequestID        string   `uri:"request_id" validate:"required,uuid"`
	Decision         string   `json:"decision" validate:"required,oneof=APPROVE REJECT SEND_BACK"`
	ApprovedBy       string   `json:"approved_by" validate:"required"`
	Remarks          *string  `json:"remarks,omitempty"`
	RejectionReason  *string  `json:"rejection_reason,omitempty"`  // required when Decision=REJECT
	MissingDocuments []string `json:"missing_documents,omitempty"` // required when Decision=SEND_BACK
}

// WithdrawRequest is the request payload for CORE-009.
// Starts WF-NFS-005 (Withdrawal Workflow).
// BR-NFS-013: Eligible statuses — CREATED, PENDING_DOCUMENTS, PENDING_APPROVAL only.
type WithdrawRequest struct {
	RequestID        string `uri:"request_id" validate:"required,uuid"`
	WithdrawalReason string `json:"withdrawal_reason" validate:"required"`
	RequestedBy      string `json:"requested_by" validate:"required"`
}

// RequestMissingDocumentsRequest is the request payload for CORE-010.
// BR-NFS-014: Generates a secure upload link (7-day expiry, max 3 attempts).
// Status transition: IN_PROGRESS → PENDING_DOCUMENTS.
type RequestMissingDocumentsRequest struct {
	RequestID         string   `uri:"request_id" validate:"required,uuid"`
	MissingDocuments  []string `json:"missing_documents" validate:"required,min=1"`
	MessageToCustomer string   `json:"message_to_customer" validate:"required"`
	RequestedBy       string   `json:"requested_by" validate:"required"`
}

// ─────────────────────────────────────────────────────────────────────────────
// Document Management Request DTOs
// Phase 3 — DM-001..005
// ─────────────────────────────────────────────────────────────────────────────

// UploadDocumentRequest carries metadata for a document upload.
// VR-NFS-009: max file size 5 MB.
// VR-NFS-010: allowed MIME types — application/pdf, image/jpeg, image/png.
type UploadDocumentRequest struct {
	RequestID    string  `form:"request_id" validate:"required,uuid"`
	DocumentType string  `form:"document_type" validate:"required"`
	Description  *string `form:"description,omitempty"`
	UploadedBy   string  `form:"uploaded_by" validate:"required"`
}

// GetDocumentMetadataRequest retrieves metadata for a specific uploaded document.
type GetDocumentMetadataRequest struct {
	DocumentID string `uri:"document_id" validate:"required,uuid"`
}

// DownloadDocumentRequest returns the actual file bytes for a document.
// Access control: customer can download their own; CPC can download any for review.
type DownloadDocumentRequest struct {
	DocumentID string `uri:"document_id" validate:"required,uuid"`
}

// DeleteDocumentRequest deletes a document upload.
// Only allowed when associated request is in CREATED status.
type DeleteDocumentRequest struct {
	DocumentID string `uri:"document_id" validate:"required,uuid"`
}

// SecureUploadDocumentRequest handles the unauthenticated secure upload via token.
// BR-NFS-014: secure link expires in 7 days, max 3 attempts.
// Status transition: PENDING_DOCUMENTS → PENDING_APPROVAL.
type SecureUploadDocumentRequest struct {
	Token        string  `uri:"token" validate:"required"`
	DocumentType string  `form:"document_type" validate:"required"`
	Description  *string `form:"description,omitempty"`
}

// ─────────────────────────────────────────────────────────────────────────────
// CPC Operations Request DTOs
// Phase 3 — CPC-001..006
// ─────────────────────────────────────────────────────────────────────────────

// GetCPCQueueRequest returns the work queue for CPC staff with SLA color indicators.
// SLA status: GREEN (<50% elapsed), AMBER (50-80% elapsed), RED (>80% elapsed / breached).
type GetCPCQueueRequest struct {
	Status       *string `form:"status"`
	RequestType  *string `form:"request_type"`
	SLAStatus    *string `form:"sla_status"`
	AssignedToMe *string `form:"assigned_to"`
	OfficeCode   *string `form:"office_code"`
	Page         *int    `form:"page"`
	PageSize     *int    `form:"page_size"`
}

// GetCPCQueueStatsRequest returns summary stats for a CPC office queue.
type GetCPCQueueStatsRequest struct {
	OfficeCode *string `form:"office_code"`
}

// AssignRequestRequest assigns a request to a specific CPC officer.
// BR-NFS-002: CPC assignment based on customer's servicing office.
type AssignRequestRequest struct {
	RequestID  string  `uri:"request_id" validate:"required,uuid"`
	AssignedTo string  `json:"assigned_to" validate:"required"`
	Reason     *string `json:"reason,omitempty"`
}

// CPCRejectRequest rejects an NFS request through the CPC rejection path.
// BR-NFS-012: IN_PROGRESS → REJECTED.
// ERR-NFS-SR-005: rejection_reason is mandatory.
type CPCRejectRequest struct {
	RequestID           string  `uri:"request_id" validate:"required,uuid"`
	RejectionReason     string  `json:"rejection_reason" validate:"required"` // mandatory — ERR-NFS-SR-005
	RejectionReasonCode *string `json:"rejection_reason_code,omitempty"`
	RejectedBy          string  `json:"rejected_by" validate:"required"`
}

// CPCSendBackRequest is an alias for requesting missing documents via CPC.
// BR-NFS-014: 7-day link expiry, max 3 attempts.
// Status: IN_PROGRESS → PENDING_DOCUMENTS.
type CPCSendBackRequest struct {
	RequestID         string   `uri:"request_id" validate:"required,uuid"`
	MissingDocuments  []string `json:"missing_documents" validate:"required,min=1"`
	MessageToCustomer string   `json:"message_to_customer" validate:"required"`
	RequestedBy       string   `json:"requested_by" validate:"required"`
}

// GetSLADashboardRequest returns SLA performance metrics for a given office/period.
type GetSLADashboardRequest struct {
	OfficeCode *string `form:"office_code"`
	SLAStatus  *string `form:"sla_status"`
}

// ─────────────────────────────────────────────────────────────────────────────
// Lookup & Validation Request DTOs
// Phase 3 — LU-001..008, VA-001..003
// ─────────────────────────────────────────────────────────────────────────────

// ListStatesRequest contains query params for LU-001.
type ListStatesRequest struct {
	ActiveOnly *bool `form:"active_only"` // default: true
}

// ValidatePincodeRequest validates a 6-digit pincode and optionally checks the claimed state.
// VR-NFS-001: pincode must be 6 numeric digits.
// BR-NFS-006: pincode-state mapping validated against India Post master table.
type ValidatePincodeRequest struct {
	Pincode string  `uri:"pincode" validate:"required,len=6,numeric"`
	State   *string `form:"state"`
}

// ListAddressTypesRequest has no fields (static list).
// Returns: PERMANENT, COMMUNICATION, BOTH
type ListAddressTypesRequest struct{}

// ListSalutationsRequest has no fields (static list).
// Returns: {Mr, Mrs, Ms, Shri, Smt, Dr} — VR-NFS-008
type ListSalutationsRequest struct{}

// ListDocumentTypesRequest returns document types for a given NFS request type.
// request_type: ADDRESS_CHANGE or NAME_CHANGE
type ListDocumentTypesRequest struct {
	RequestType string `uri:"request_type" validate:"required,oneof=ADDRESS_CHANGE NAME_CHANGE"`
}

// ListOfficesRequest returns CPC offices filtered by type and circle code.
type ListOfficesRequest struct {
	Type       *string `form:"type"`
	CircleCode *string `form:"circle_code"`
}

// ListChannelsRequest has no fields (static list).
// Returns: Portal, Mobile, Branch, Satellite (VR-NFS-016)
type ListChannelsRequest struct{}

// ListStatusReasonsRequest returns pre-defined reason codes for a given action type.
type ListStatusReasonsRequest struct {
	ActionType *string `form:"action_type"` // REJECT, SEND_BACK, WITHDRAW
}

// ValidateAddressRequest checks address fields for compliance.
// VR-NFS-001: valid Indian pincode.
// VR-NFS-002: state not empty, matches India state list.
// VR-NFS-003: address_line_1 not empty, max 150 chars.
// VR-NFS-004: city not empty, max 100 chars.
// VR-NFS-005: district not empty, max 100 chars.
// BR-NFS-006: pincode-state mismatch → validation error.
type ValidateAddressRequest struct {
	AddressLine1 string  `json:"address_line1" validate:"required,max=150"` // VR-NFS-003
	AddressLine2 *string `json:"address_line2,omitempty"`
	Village      *string `json:"village,omitempty"`
	Taluka       *string `json:"taluka,omitempty"`
	City         string  `json:"city" validate:"required,max=100"`          // VR-NFS-004
	District     string  `json:"district" validate:"required,max=100"`      // VR-NFS-005
	State        string  `json:"state" validate:"required"`                 // VR-NFS-002
	Pincode      string  `json:"pincode" validate:"required,len=6,numeric"` // VR-NFS-001
}

// ValidateNameFieldsRequest validates name fields for compliance.
// VR-NFS-006: first_name — alphabets and spaces only, max 100 chars.
// VR-NFS-007: last_name — alphabets and spaces only, max 100 chars.
// VR-NFS-008: salutation — {Mr, Mrs, Ms, Shri, Smt, Dr}.
// VR-NFS-014: date_of_birth MUST NOT be present (BR-NFS-010 — immutable).
type ValidateNameFieldsRequest struct {
	Salutation *string `json:"salutation,omitempty" validate:"omitempty,oneof=Mr Mrs Ms Shri Smt Dr"` // VR-NFS-008
	FirstName  *string `json:"first_name,omitempty" validate:"omitempty,max=100"`                     // VR-NFS-006
	MiddleName *string `json:"middle_name,omitempty"`
	LastName   *string `json:"last_name,omitempty" validate:"omitempty,max=100"` // VR-NFS-007
	// NOTE: date_of_birth field intentionally absent — VR-NFS-014 / BR-NFS-010
}

// DuplicateCheckRequest checks for a pending active request.
// VR-NFS-015: only one active request of a given type per customer.
type DuplicateCheckRequest struct {
	CustomerID  int64  `uri:"customer_id" validate:"required,gt=0"`
	RequestType string `uri:"request_type" validate:"required,oneof=ADDRESS_CHANGE NAME_CHANGE"`
}

// ─────────────────────────────────────────────────────────────────────────────
// Status & Tracking Request DTOs
// Phase 3 — ST-001..005
// ─────────────────────────────────────────────────────────────────────────────

// GetRequestDetailRequest contains path + query params for ST-001.
// include_documents: appends document list to response.
// include_audit: appends audit trail to response (for CPC/admin use).
type GetRequestDetailRequest struct {
	RequestID        string `uri:"request_id" validate:"required,uuid"`
	IncludeDocuments *bool  `form:"include_documents"`
	IncludeAudit     *bool  `form:"include_audit"`
}

// GetTimelineRequest returns audit log history for a request.
// FR-NFS-008: Audit trail.
type GetTimelineRequest struct {
	RequestID string `uri:"request_id" validate:"required,uuid"`
}

// GetReceiptRequest returns a PDF receipt for a completed/acknowledged request.
// FR-NFS-004/005: acknowledgment receipt generation.
type GetReceiptRequest struct {
	RequestID string `uri:"request_id" validate:"required,uuid"`
}

// ListCustomerRequestsRequest returns paginated list of requests for a customer.
// BATCH: 2-query BATCH — count + paginated rows.
type ListCustomerRequestsRequest struct {
	CustomerID  int64   `uri:"customer_id" validate:"required,gt=0"`
	RequestType *string `form:"request_type"` // ADDRESS_CHANGE | NAME_CHANGE
	Status      *string `form:"status"`
	Page        *int    `form:"page"`
	PageSize    *int    `form:"page_size"`
}

// GetRequestDocumentsRequest returns documents for a specific request.
type GetRequestDocumentsRequest struct {
	RequestID string `uri:"request_id" validate:"required,uuid"`
}

// ─────────────────────────────────────────────────────────────────────────────
// Mobile Change Request DTOs
// ─────────────────────────────────────────────────────────────────────────────

// InitiateMobileChangeRequest is the request payload for mobile change initiation.
// WF-NFS-006: OTP-based mobile number change.
type InitiateMobileChangeRequest struct {
	CustomerID      int64  `json:"customer_id" validate:"required,gt=0"`
	NewMobileNumber string `json:"new_mobile_number" validate:"required,len=10,numeric"`
	Channel         string `json:"channel" validate:"required,oneof=Portal Mobile PostOffice CallCenter AgentPortal"`
}

// VerifyMobileOTPRequest is the request payload for mobile change OTP verification.
type VerifyMobileOTPRequest struct {
	RequestID string `uri:"request_id" validate:"required"`
	OTPCode   string `json:"otp_code" validate:"required,len=6,numeric"`
	TxnID     string `json:"txn_id" validate:"required"`
}

// ─────────────────────────────────────────────────────────────────────────────
// Email Change Request DTOs
// ─────────────────────────────────────────────────────────────────────────────

// InitiateEmailChangeRequest is the request payload for email change initiation.
// WF-NFS-007: OTP-based email change.
type InitiateEmailChangeRequest struct {
	CustomerID int64  `json:"customer_id" validate:"required,gt=0"`
	NewEmail   string `json:"new_email" validate:"required,email,max=255"`
	Channel    string `json:"channel" validate:"required,oneof=Portal Mobile PostOffice CallCenter AgentPortal"`
}

// VerifyEmailOTPRequest is the request payload for email change OTP verification.
type VerifyEmailOTPRequest struct {
	RequestID string `uri:"request_id" validate:"required"`
	OTPCode   string `json:"otp_code" validate:"required,len=6,numeric"`
	TxnID     string `json:"txn_id" validate:"required"`
}
