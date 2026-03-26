// Package activities implements Temporal activity stubs for address change workflows.
//
// WF-NFS-001: Aadhaar-based Address Change
//
//	Sequence: ValidateRequest → CreateServiceRequest → StoreWorkflowState →
//	          RequestAadhaarOTP → [Wait otp_submitted signal, 15 min timeout] →
//	          VerifyAadhaarOTP → UpdateAddressData → CreateAddressVersion →
//	          GenerateAckReceipt → UpdateStatus(COMPLETED) → CreateAuditLog
//
// WF-NFS-002: Manual Address Change (CPC Approval)
//
//	Sequence: ValidateRequest → CreateServiceRequest → StoreWorkflowState →
//	          AssignToCPC → [Wait approval_decision signal, 45-day SLA] →
//	          [If APPROVE] UpdateAddressData → CreateAddressVersion →
//	                       GenerateAckReceipt → UpdateStatus(COMPLETED)
//	          [If REJECT]  UpdateStatus(REJECTED)
//	          [If SEND_BACK] UpdateStatus(PENDING_DOCUMENTS) → notify customer
//	          → CreateAuditLog
//
// BATCH NOTE: CreateServiceRequest activity calls repo.CreateWithAddressDetail
//
//	which batches service_request + address_change_detail + audit_log inserts.
//
// WORKFLOW STATE NOTE: StoreWorkflowState is called immediately after workflow
//
//	start to persist workflow_id + workflow_run_id into nfs.service_request.
//	This enables HTTP signal endpoints (OTP verify, approve) to locate the
//	running workflow via temporal.Client.SignalWorkflow.
package activities

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.temporal.io/sdk/activity"
	temporalclient "go.temporal.io/sdk/client"

	config "gitlab.cept.gov.in/it-2.0-common/api-config"
	log "gitlab.cept.gov.in/it-2.0-common/n-api-log"

	"customer-nfs-service/core/domain"
	"customer-nfs-service/repo/postgres"
)

// AddressChangeActivities holds all activity implementations for address change.
type AddressChangeActivities struct {
	srRepo    *postgres.ServiceRequestRepository
	addrRepo  *postgres.AddressChangeRepository
	auditRepo *postgres.AuditLogRepository
	cfg       *config.Config
	tc        *temporalclient.Client
}

// NewAddressChangeActivities constructs the activity struct.
func NewAddressChangeActivities(
	srRepo *postgres.ServiceRequestRepository,
	addrRepo *postgres.AddressChangeRepository,
	auditRepo *postgres.AuditLogRepository,
	cfg *config.Config,
	tc *temporalclient.Client,
) *AddressChangeActivities {
	return &AddressChangeActivities{
		srRepo:    srRepo,
		addrRepo:  addrRepo,
		auditRepo: auditRepo,
		cfg:       cfg,
		tc:        tc,
	}
}

// ---------------------------------------------------------------------------
// StoreWorkflowState persists the Temporal workflow_id and workflow_run_id
// into nfs.service_request so HTTP signal endpoints can locate the workflow.
//
// WORKFLOW STATE NOTE: Must be called immediately after workflow.GetInfo(ctx)
// at the start of every address change workflow execution.
// WF-NFS-001, WF-NFS-002
// ---------------------------------------------------------------------------
func (a *AddressChangeActivities) StoreWorkflowState(ctx context.Context, input StoreWorkflowStateInput) (*StoreWorkflowStateResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("StoreWorkflowState", "requestID", input.RequestID, "workflowID", input.WorkflowID)

	if err := a.srRepo.UpdateWorkflowState(ctx, input.RequestID, input.WorkflowID, input.WorkflowRunID); err != nil {
		log.Error(ctx, "StoreWorkflowState [%s]: %v", input.RequestID, err)
		return nil, fmt.Errorf("store workflow state: %w", err)
	}
	return &StoreWorkflowStateResult{Stored: true}, nil
}

// ---------------------------------------------------------------------------
// ValidateAddressRequest validates all business rules and duplicate checks
// before creating the service request.
//
// BR-NFS-001..006 (address change rules)
// VR-NFS-001..005 (address field validations)
// VR-NFS-013 (pincode format)
// VR-NFS-015 (duplicate active request check)
// BR-NFS-005 (only INSURED role handled by NFS)
// ---------------------------------------------------------------------------
func (a *AddressChangeActivities) ValidateAddressRequest(ctx context.Context, input ValidateAddressRequestInput) (*ValidateAddressRequestResult, error) {
	result := &ValidateAddressRequestResult{IsValid: true}

	// BR-NFS-005: Only INSURED role is handled by Customer NFS
	if input.AddressUpdateFor != "INSURED" {
		result.IsValid = false
		result.ValidationErrors = append(result.ValidationErrors,
			fmt.Sprintf("BR-NFS-005: address update for %q not handled by Customer NFS; route to Policy Admin Service", input.AddressUpdateFor))
	}

	// VR-NFS-001: address_line1 required and non-empty
	if len(input.NewAddressLine1) == 0 {
		result.IsValid = false
		result.ValidationErrors = append(result.ValidationErrors, "VR-NFS-001: new_address_line1 is required")
	}

	// VR-NFS-002: city required
	if len(input.NewCity) == 0 {
		result.IsValid = false
		result.ValidationErrors = append(result.ValidationErrors, "VR-NFS-002: new_city is required")
	}

	// VR-NFS-003: district required
	if len(input.NewDistrict) == 0 {
		result.IsValid = false
		result.ValidationErrors = append(result.ValidationErrors, "VR-NFS-003: new_district is required")
	}

	// VR-NFS-004: state required
	if len(input.NewState) == 0 {
		result.IsValid = false
		result.ValidationErrors = append(result.ValidationErrors, "VR-NFS-004: new_state is required")
	}

	// VR-NFS-013: pincode must be 6 digits
	if len(input.NewPincode) != 6 {
		result.IsValid = false
		result.ValidationErrors = append(result.ValidationErrors, "VR-NFS-013: pincode must be exactly 6 digits")
	}

	// VR-NFS-015: check for duplicate active request
	hasDup, existingTicket, err := a.srRepo.CheckDuplicateRequest(ctx, input.CustomerID, "ADDRESS_CHANGE")
	if err != nil {
		return nil, fmt.Errorf("duplicate check: %w", err)
	}
	if hasDup {
		result.IsValid = false
		result.HasDuplicate = true
		result.ExistingTicket = existingTicket
		result.ValidationErrors = append(result.ValidationErrors,
			fmt.Sprintf("VR-NFS-015: active request %s already exists for this customer", existingTicket))
	}

	// NOTE: BR-NFS-006 (pincode-state correlation) is validated by DB function
	// nfs.validate_pincode_state called by the VA-001 validation endpoint.

	return result, nil
}

// ---------------------------------------------------------------------------
// CreateAddressServiceRequest creates service_request + address_change_detail
// + initial audit_log in a single batched TX.
//
// BATCH: All 3 INSERTs go in one pgx TX batch (one round-trip).
// FR-NFS-001, BR-NFS-011, BR-NFS-016
// ---------------------------------------------------------------------------
func (a *AddressChangeActivities) CreateAddressServiceRequest(
	ctx context.Context,
	input CreateAddressServiceRequestInput,
) (*CreateAddressServiceRequestResult, error) {
	// Generate ticket number (BR-NFS-011: NFS-ANC-{YYYYMMDD}-{SEQ6})
	ticketNumber, err := a.srRepo.GenerateTicketNumber(ctx, "ADDRESS_CHANGE")
	if err != nil {
		return nil, fmt.Errorf("generate ticket number: %w", err)
	}

	now := time.Now().UTC()
	requestID := newUUID()
	detailID := newUUID()
	auditID := newUUID()

	sr := &domain.ServiceRequest{
		RequestID:    requestID,
		TicketNumber: ticketNumber,
		CustomerID:   input.CustomerID,
		PolicyNumber: input.PolicyNumber,
		RequestType:  "ADDRESS_CHANGE",
		AuthMethod:   input.AuthMethod,
		Status:       "CREATED",
		Channel:      input.Channel,
		OfficeCode:   input.OfficeCode,
		InitiatedBy:  input.InitiatedBy,
		SLADeadline:  input.SLADeadline,
		CreatedBy:    input.InitiatedBy,
		CreatedAt:    now,
	}

	detail := &domain.AddressChangeDetail{
		DetailID:         detailID,
		RequestID:        requestID,
		AddressUpdateFor: input.AddressUpdateFor,
		AddressType:      input.AddressType,
		OldAddressLine1:  input.OldAddressLine1,
		OldCity:          input.OldCity,
		OldDistrict:      input.OldDistrict,
		OldState:         input.OldState,
		OldPincode:       input.OldPincode,
		NewAddressLine1:  input.NewAddressLine1,
		NewAddressLine2:  input.NewAddressLine2,
		NewVillage:       input.NewVillage,
		NewTaluka:        input.NewTaluka,
		NewCity:          input.NewCity,
		NewDistrict:      input.NewDistrict,
		NewState:         input.NewState,
		NewPincode:       input.NewPincode,
		CreatedBy:        input.InitiatedBy,
	}

	newValueJSON := fmt.Sprintf(`{"request_id":"%s","ticket_number":"%s","status":"CREATED"}`, requestID, ticketNumber)
	audit := &domain.AuditLog{
		AuditID:       auditID,
		RequestID:     requestID,
		ActionType:    "CREATED",
		NewValueJSON:  newValueJSON,
		PerformedByID: input.InitiatedBy,
		PerformedAt:   now,
		Channel:       &input.Channel,
		OfficeCode:    input.OfficeCode,
	}

	// BATCH: 3 INSERTs in one TX batch
	if err := a.srRepo.CreateWithAddressDetail(ctx, sr, detail, audit); err != nil {
		log.Error(ctx, "CreateAddressServiceRequest: %v", err)
		return nil, fmt.Errorf("create address service request: %w", err)
	}

	log.Info(ctx, "CreateAddressServiceRequest: created %s (%s)", requestID, ticketNumber)
	return &CreateAddressServiceRequestResult{
		RequestID:    sr.RequestID,
		TicketNumber: sr.TicketNumber,
		Status:       sr.Status,
	}, nil
}

// ---------------------------------------------------------------------------
// RequestAadhaarOTP initiates an Aadhaar OTP request to UIDAI.
// Only allowed for Portal + Mobile channels (VR-NFS-016 — config-driven).
// WF-NFS-001, WF-NFS-003.
// ---------------------------------------------------------------------------
func (a *AddressChangeActivities) RequestAadhaarOTP(ctx context.Context, input AadhaarOTPRequestInput) (*AadhaarOTPRequestResult, error) {
	// VR-NFS-016: check allowed channels from config
	allowedChannels := a.cfg.GetString("nfs.aadhaarallowedchannels") // "Portal,Mobile"
	// TODO: Call UIDAI Aadhaar OTP service (external integration)
	// This is a STUB — replace with actual UIDAI client call
	_ = allowedChannels
	log.Info(ctx, "RequestAadhaarOTP: STUB for requestID=%s customerID=%s channel=%s",
		input.RequestID, input.CustomerID, input.Channel)

	return &AadhaarOTPRequestResult{
		OTPReferenceID: "OTP-REF-STUB-" + input.RequestID,
		ExpiresAt:      time.Now().Add(15 * time.Minute), // 15-min window (WF-NFS-001)
	}, nil
}

// ---------------------------------------------------------------------------
// VerifyAadhaarOTP verifies the submitted OTP and stores UIDAI txn ID.
// WF-NFS-001, WF-NFS-003.
// ---------------------------------------------------------------------------
func (a *AddressChangeActivities) VerifyAadhaarOTP(ctx context.Context, input AadhaarOTPVerifyInput) (*AadhaarOTPVerifyResult, error) {
	// TODO: Call UIDAI OTP verify endpoint (external integration)
	// This is a STUB — replace with actual UIDAI client call
	log.Info(ctx, "VerifyAadhaarOTP: STUB for requestID=%s", input.RequestID)

	aadhaarTxnID := "UIDAI-TXN-STUB-" + input.RequestID

	// Store aadhaar_txn_id in address_change_detail
	if err := a.addrRepo.UpdateAadhaarTxnID(ctx, input.RequestID, aadhaarTxnID); err != nil {
		return nil, fmt.Errorf("store aadhaar_txn_id: %w", err)
	}

	return &AadhaarOTPVerifyResult{
		Verified:     true,
		AadhaarTxnID: aadhaarTxnID,
	}, nil
}

// ---------------------------------------------------------------------------
// UpdateAddressData applies the new address to customer data and records an
// address version history entry. Called after successful Aadhaar OTP verify
// (WF-NFS-001) or CPC approval (WF-NFS-002).
//
// BATCH: Deactivate old version + insert new version in one TX batch
// (CreateAddressVersion is batched internally).
// BR-NFS-001, BR-NFS-002
// ---------------------------------------------------------------------------
func (a *AddressChangeActivities) UpdateAddressData(ctx context.Context, input UpdateAddressDataInput) (*UpdateAddressDataResult, error) {
	// Fetch the current request + detail
	sr, addrDetail, err := a.addrRepo.GetWithServiceRequest(ctx, input.RequestID)
	if err != nil {
		return nil, fmt.Errorf("fetch address details: %w", err)
	}
	log.Info(ctx, "UpdateAddressData: sr.CustomerID=%q sr.RequestID=%q addrDetail.AddressType=%q", sr.CustomerID, sr.RequestID, addrDetail.AddressType)

	// TODO: Call Customer Core Service to update address in customer profile
	// This is the external integration point.
	log.Info(ctx, "UpdateAddressData: STUB calling Customer Core Service for customerID=%s", sr.CustomerID)

	// Record the new address version in address_version_history
	// BATCH: deactivate old version + insert new version in one TX batch
	versionID := newUUID()
	newVersion := &domain.AddressVersionHistory{
		VersionID:     versionID,
		CustomerID:    sr.CustomerID,
		RequestID:     &input.RequestID,
		AddressType:   addrDetail.AddressType,
		AddressLine1:  addrDetail.NewAddressLine1,
		City:          addrDetail.NewCity,
		District:      addrDetail.NewDistrict,
		State:         addrDetail.NewState,
		Pincode:       addrDetail.NewPincode,
		VersionNumber: 1, // Incremented in CreateAddressVersion
		CreatedBy:     input.UpdatedBy,
	}

	if err := a.addrRepo.CreateAddressVersion(ctx, sr.CustomerID, addrDetail.AddressType, newVersion); err != nil {
		return nil, fmt.Errorf("create address version: %w", err)
	}

	return &UpdateAddressDataResult{
		Updated:      true,
		NewVersionID: versionID,
	}, nil
}

// ---------------------------------------------------------------------------
// AssignToCPC assigns the request to the CPC work queue.
// BATCH: UPDATE service_request + INSERT cpc_work_queue + INSERT audit_log.
// FR-NFS-008, BR-NFS-016
// ---------------------------------------------------------------------------
func (a *AddressChangeActivities) AssignToCPC(ctx context.Context, input AssignToCPCInput) (*AssignToCPCResult, error) {
	auditID := newUUID()
	newValueJSON := fmt.Sprintf(`{"request_id":"%s","action":"ASSIGNED_TO_CPC"}`, input.RequestID)
	audit := &domain.AuditLog{
		AuditID:       auditID,
		RequestID:     input.RequestID,
		ActionType:    "ASSIGNED",
		NewValueJSON:  newValueJSON,
		PerformedByID: input.AssignedBy,
		PerformedAt:   time.Now().UTC(),
		OfficeCode:    input.OfficeCode,
	}

	// if err := a.srRepo.AssignToCPC(ctx, input.RequestID, "SYSTEM", input.AssignedBy, input.Priority, input.SLADeadline, audit); err != nil {
	// 	return nil, fmt.Errorf("assign to CPC: %w", err)
	// }
	if err := a.srRepo.AssignToCPC(ctx, input.RequestID, nil, input.AssignedBy, input.Priority, input.SLADeadline, audit); err != nil {
		return nil, fmt.Errorf("assign to CPC: %w", err)
	}
	log.Info(ctx, "AssignToCPC: requestID=%s priority=%d", input.RequestID, input.Priority)
	return &AssignToCPCResult{AssignedTo: "SYSTEM"}, nil
}

// ---------------------------------------------------------------------------
// UpdateStatus transitions the service request status.
// BATCH: UPDATE + status_transition_history + audit_log in one TX batch.
// BR-NFS-012, BR-NFS-016
// ---------------------------------------------------------------------------
func (a *AddressChangeActivities) UpdateStatus(ctx context.Context, input UpdateStatusInput) (*UpdateStatusResult, error) {
	transitionID := newUUID()
	auditID := newUUID()
	now := time.Now().UTC()

	transition := &domain.StatusTransitionHistory{
		TransitionID:     transitionID,
		RequestID:        input.RequestID,
		FromStatus:       &input.FromStatus,
		ToStatus:         input.NewStatus,
		TransitionReason: input.Reason,
		TransitionedBy:   input.UpdatedBy,
		TransitionedAt:   now,
		WorkflowSignalID: input.WorkflowSignalID,
	}

	newValueJSON := fmt.Sprintf(`{"status":"%s"}`, input.NewStatus)
	oldValueJSON := fmt.Sprintf(`{"status":"%s"}`, input.FromStatus)
	audit := &domain.AuditLog{
		AuditID:       auditID,
		RequestID:     input.RequestID,
		ActionType:    "STATUS_CHANGE",
		OldValue:      &oldValueJSON,
		NewValueJSON:  newValueJSON,
		PerformedByID: input.UpdatedBy,
		PerformedAt:   now,
		Channel:       &input.Channel,
		OfficeCode:    input.OfficeCode,
		Notes:         input.Reason,
	}

	if _, err := a.srRepo.UpdateStatus(ctx, input.RequestID, input.NewStatus, &input.UpdatedBy, input.Reason, transition, audit); err != nil {
		log.Error(ctx, "UpdateStatus [%s → %s]: %v", input.FromStatus, input.NewStatus, err)
		return nil, fmt.Errorf("update status: %w", err)
	}

	return &UpdateStatusResult{RequestID: input.RequestID, NewStatus: input.NewStatus}, nil
}

// ---------------------------------------------------------------------------
// GenerateAckReceipt triggers acknowledgement receipt generation via DMS.
// FR-NFS-004: Acknowledgement receipt must be generated after request creation.
// ---------------------------------------------------------------------------
func (a *AddressChangeActivities) GenerateAckReceipt(ctx context.Context, input GenerateAckReceiptInput) (*GenerateAckReceiptResult, error) {
	// TODO: Call DMS service to generate PDF receipt
	// STUB — replace with actual DMS client call
	log.Info(ctx, "GenerateAckReceipt: STUB for requestID=%s ticket=%s", input.RequestID, input.TicketNumber)
	return &GenerateAckReceiptResult{
		ReceiptDMSID:   "DMS-STUB-" + input.RequestID,
		ReceiptFileURL: "/receipts/" + input.TicketNumber + ".pdf",
	}, nil
}

// ---------------------------------------------------------------------------
// Escalate sends an SLA breach notification. FR-NFS-010.
// ---------------------------------------------------------------------------
func (a *AddressChangeActivities) Escalate(ctx context.Context, input EscalateInput) (*EscalateResult, error) {
	// TODO: Call notification service (email/SMS escalation)
	// STUB — replace with actual notification client
	log.Info(ctx, "Escalate: STUB for requestID=%s slaBreached=%v", input.RequestID, input.SLABreached)
	return &EscalateResult{Escalated: true}, nil
}

// ---------------------------------------------------------------------------
// newUUID generates a new UUID string.
// In production, use github.com/google/uuid or uuid_generate_v4() result.
// ---------------------------------------------------------------------------
func newUUID() string {
	return uuid.New().String()
}
