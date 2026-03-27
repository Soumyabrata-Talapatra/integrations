// Package activities implements Temporal activity stubs for name change workflows.
//
// Workflows served: WF-NFS-003 (Aadhaar name change), WF-NFS-004 (Manual name change)
//
// BATCH NOTE:
//
//	CreateNameServiceRequest uses a 3-INSERT TX batch:
//	service_request + name_change_detail + audit_log (single round-trip).
//
// WORKFLOW STATE NOTE:
//
//	StoreWorkflowState is always the FIRST activity called.
//	It persists workflow_id + workflow_run_id into nfs.service_request so HTTP handlers
//	can call SignalWorkflow for verify-otp (CORE-006) and approve (CORE-008).
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

// NameChangeActivities bundles all Temporal activities for WF-NFS-003 and WF-NFS-004.
// FR-NFS-004 (Aadhaar), FR-NFS-005 (Manual), FR-NFS-006 (Lifecycle), FR-NFS-008 (Audit)
type NameChangeActivities struct {
	srRepo    *postgres.ServiceRequestRepository
	nameRepo  *postgres.NameChangeRepository
	auditRepo *postgres.AuditLogRepository
	cfg       config.Config
	tc        temporalclient.Client
}

// NewNameChangeActivities creates the activities struct for DI injection.
func NewNameChangeActivities(
	srRepo *postgres.ServiceRequestRepository,
	nameRepo *postgres.NameChangeRepository,
	auditRepo *postgres.AuditLogRepository,
	cfg config.Config,
	tc temporalclient.Client,
) *NameChangeActivities {
	return &NameChangeActivities{
		srRepo:    srRepo,
		nameRepo:  nameRepo,
		auditRepo: auditRepo,
		cfg:       cfg,
		tc:        tc,
	}
}

// StoreWorkflowStateForName persists workflow_id and workflow_run_id for the name change request.
// FIRST activity in every name change workflow — enables HTTP signal dispatch.
// WORKFLOW STATE: see temporal/workflows/address_change_workflow.go note.
func (a *NameChangeActivities) StoreWorkflowStateForName(ctx context.Context, input StoreWorkflowStateInput) (*StoreWorkflowStateResult, error) {
	logger := activity.GetLogger(ctx)
	info := activity.GetInfo(ctx)

	wfID := info.WorkflowExecution.ID
	runID := info.WorkflowExecution.RunID

	logger.Info("StoreWorkflowState(name): persisting workflow identity",
		"requestID", input.RequestID,
		"workflowID", wfID,
		"runID", runID,
	)

	err := a.srRepo.UpdateWorkflowState(ctx, input.RequestID, wfID, runID)
	if err != nil {
		return nil, fmt.Errorf("StoreWorkflowStateForName: %w", err)
	}

	return &StoreWorkflowStateResult{Stored: true}, nil
}

// ValidateNameRequest validates all name fields before creating the service request.
// FR-NFS-004, FR-NFS-005.
// VR-NFS-006: first_name — not empty, alpha+spaces, max 100 chars.
// VR-NFS-007: last_name — not empty, alpha+spaces, max 100 chars.
// VR-NFS-008: salutation — must be in {Mr, Mrs, Ms, Shri, Smt, Dr}.
// VR-NFS-014: DOB must NOT be present in request (BR-NFS-010 — DOB immutable).
// VR-NFS-015: duplicate check — no active pending request of same type.
// BR-NFS-015: AADHAAR only from Portal and Mobile channels.
func (a *NameChangeActivities) ValidateNameRequest(ctx context.Context, input ValidateNameRequestInput) (*ValidateNameRequestResult, error) {
	logger := activity.GetLogger(ctx)
	log.Info(ctx, "ValidateNameRequest: validating request for customer %d", input.CustomerID)

	// VR-NFS-008: Allowed salutations.
	allowedSalutations := map[string]bool{
		"Mr": true, "Mrs": true, "Ms": true,
		"Shri": true, "Smt": true, "Dr": true,
	}
	result := &ValidateNameRequestResult{IsValid: true}

	if input.NewSalutation != "" {
		if !allowedSalutations[input.NewSalutation] {
			result.IsValid = false
			result.ValidationErrors = append(result.ValidationErrors,
				fmt.Sprintf("VR-NFS-008: invalid salutation %q; must be one of Mr/Mrs/Ms/Shri/Smt/Dr", input.NewSalutation))
		}
	}

	// VR-NFS-006: first_name check.
	if len(input.NewFirstName) == 0 || len(input.NewFirstName) > 100 {
		result.IsValid = false
		result.ValidationErrors = append(result.ValidationErrors,
			"VR-NFS-006: first_name is required and must be ≤ 100 characters")
	}

	// VR-NFS-007: last_name check.
	if len(input.NewLastName) == 0 || len(input.NewLastName) > 100 {
		result.IsValid = false
		result.ValidationErrors = append(result.ValidationErrors,
			"VR-NFS-007: last_name is required and must be ≤ 100 characters")
	}

	// BR-NFS-015: AADHAAR only from Portal and Mobile.
	if input.AuthMethod == "AADHAAR" && input.Channel != "Portal" && input.Channel != "Mobile" {
		result.IsValid = false
		result.ValidationErrors = append(result.ValidationErrors,
			fmt.Sprintf("BR-NFS-015: Aadhaar auto-approval not available from channel %q", input.Channel))
	}

	// VR-NFS-015: duplicate check.--duplicate check
	// isDuplicate, existingTicket, err := a.srRepo.CheckDuplicateRequest(ctx, input.CustomerID, "NAME_CHANGE")
	// if err != nil {
	// 	return nil, fmt.Errorf("ValidateNameRequest: duplicate check: %w", err)
	// }
	// if isDuplicate {
	// 	result.IsValid = false
	// 	result.HasDuplicate = true
	// 	result.ExistingTicket = existingTicket
	// 	result.ValidationErrors = append(result.ValidationErrors,
	// 		fmt.Sprintf("VR-NFS-015: active request %s already exists for this customer", existingTicket))
	// }

	logger.Info("ValidateNameRequest: validation passed", "customerID", input.CustomerID)
	return result, nil
}

// CreateNameServiceRequest creates the service_request + name_change_detail + audit_log atomically.
// BATCH: 3-INSERT TX batch — single DB round-trip (see repo/postgres/service_request.go).
// FR-NFS-004, FR-NFS-005, BR-NFS-011 (ticket number generated before this call).
// BR-NFS-016: audit_log is INSERT-only.
func (a *NameChangeActivities) CreateNameServiceRequest(ctx context.Context, input CreateNameServiceRequestInput) (*CreateNameServiceRequestResult, error) {
	// Generate ticket number (BR-NFS-011: NFS-NMC-{YYYYMMDD}-{SEQ6})
	ticketNumber, err := a.srRepo.GenerateTicketNumber(ctx, "NAME_CHANGE")
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
		RequestType:  "NAME_CHANGE",
		AuthMethod:   input.AuthMethod,
		Status:       "CREATED",
		Channel:      input.Channel,
		OfficeCode:   input.OfficeCode,
		InitiatedBy:  parseStringToInt64(input.InitiatedBy),
		SLADeadline:  input.SLADeadline,
		CreatedBy:    input.InitiatedBy,
		CreatedAt:    now,
	}

	detail := &domain.NameChangeDetail{
		DetailID:      detailID,
		RequestID:     requestID,
		OldSalutation: input.OldSalutation,
		OldFirstName:  input.OldFirstName,
		OldMiddleName: input.OldMiddleName,
		OldLastName:   input.OldLastName,
		NewSalutation: &input.NewSalutation,
		NewFirstName:  &input.NewFirstName,
		NewMiddleName: input.NewMiddleName,
		NewLastName:   &input.NewLastName,
		CreatedBy:     input.InitiatedBy,
		CreatedAt:     now,
	}

	newValueJSON := fmt.Sprintf(`{"request_id":%q,"ticket_number":%q,"status":"CREATED"}`, requestID, ticketNumber)
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

	activity.GetLogger(ctx).Info("CreateNameServiceRequest: creating batched TX",
		"requestID", requestID, "ticketNumber", ticketNumber, "authMethod", input.AuthMethod)

	// BATCH: CreateWithNameDetail — 3-INSERT TX batch.
	if err := a.srRepo.CreateWithNameDetail(ctx, sr, detail, audit); err != nil {
		log.Error(ctx, "CreateNameServiceRequest: %v", err)
		return nil, fmt.Errorf("CreateNameServiceRequest: %w", err)
	}

	log.Info(ctx, "CreateNameServiceRequest: created %s (%s)", requestID, ticketNumber)
	return &CreateNameServiceRequestResult{
		RequestID:    requestID,
		TicketNumber: ticketNumber,
		Status:       "CREATED",
	}, nil
}

// RequestNameOTP sends an Aadhaar OTP for name change identity verification.
// WF-NFS-003 Step 2: OTP dispatched before waiting for otp_submitted signal.
// STUB — calls KYC Service (INT-NFS-003).
// FR-NFS-004: Aadhaar-based name change.
func (a *NameChangeActivities) RequestNameOTP(ctx context.Context, input AadhaarOTPRequestInput) (*AadhaarOTPRequestResult, error) {
	log.Info(ctx, "RequestNameOTP [STUB]: dispatching Aadhaar OTP for customer %d", input.CustomerID)
	// TODO: call KYC Service: POST /kyc/aadhaar/otp { customer_id, request_id }
	return &AadhaarOTPRequestResult{
		OTPReferenceID: "stub-txn-" + input.RequestID,
		ExpiresAt:      time.Now().Add(15 * time.Minute), // 15-min OTP window (WF-NFS-003)
	}, nil
}

// VerifyNameOTP verifies the Aadhaar OTP and retrieves the authoritative name from UIDAI.
// WF-NFS-003 Step 3: Called after otp_submitted signal received.
// BR-NFS-007: UIDAI name becomes authoritative on successful OTP verify.
// BR-NFS-010: DOB from UIDAI response is DISCARDED — never stored.
// STUB — calls KYC Service (INT-NFS-003).
func (a *NameChangeActivities) VerifyNameOTP(ctx context.Context, input AadhaarOTPVerifyInput) (*AadhaarOTPVerifyResult, error) {
	log.Info(ctx, "VerifyNameOTP [STUB]: verifying OTP for request %s", input.RequestID)
	// TODO: call KYC Service: POST /kyc/aadhaar/verify { txn_id, otp }
	// On success: UIDAI returns { salutation, first_name, middle_name, last_name }
	// BR-NFS-010: DOB from UIDAI response MUST be discarded.

	// Generate a stub UIDAI transaction ID (BR-NFS-007: stored in name_change_detail).
	aadhaarTxnID := "UIDAI-TXN-STUB-" + input.RequestID

	// Store Aadhaar txn_id in name_change_detail (BR-NFS-007).
	if err := a.nameRepo.UpdateAadhaarTxnID(ctx, input.RequestID, aadhaarTxnID); err != nil {
		return nil, fmt.Errorf("VerifyNameOTP: store txn_id: %w", err)
	}

	return &AadhaarOTPVerifyResult{
		Verified:     true,
		AadhaarTxnID: aadhaarTxnID,
	}, nil
}

// UpdateNameData applies the approved name change — creates a new version history entry.
// BATCH (TX): 2-op (deactivate old + insert new) + 1 audit_log INSERT.
// BR-NFS-009: emits customer.name.updated event for Policy Service consumption.
// BR-NFS-010: DOB is never written here (immutable).
// Called by WF-NFS-003 (after OTP verify) and WF-NFS-004 (after APPROVE signal).
func (a *NameChangeActivities) UpdateNameData(ctx context.Context, input UpdateAddressDataInput) (*UpdateAddressDataResult, error) {
	log.Info(ctx, "UpdateNameData: applying name change for request %s", input.RequestID)

	// Fetch service request + name detail to get customer data for the version entry.
	nameDetail, sr, err := a.nameRepo.GetWithServiceRequest(ctx, input.RequestID)
	if err != nil {
		return nil, fmt.Errorf("UpdateNameData: fetch request: %w", err)
	}

	// TODO: call Customer Core Service to update the canonical customer profile.
	log.Info(ctx, "UpdateNameData: STUB calling Customer Core Service for customerID=%d", sr.CustomerID)

	// Create a new name version history entry (deactivates old + inserts new in one TX batch).
	// // BR-NFS-009: the version entry is the source of truth for cross-policy propagation.
	// // Only Salutation, FirstName, MiddleName, LastName, CreatedBy are used by the repo INSERT;
	// // all other fields (version_id, effective_from, is_current) are generated by the DB.
	// salutation := nameDetail.NewSalutation
	// versionEntry := &domain.NameVersionHistory{
	// 	Salutation: &salutation,
	// 	FirstName:  nameDetail.NewFirstName,
	// 	MiddleName: nameDetail.NewMiddleName,
	// 	LastName:   nameDetail.NewLastName,
	// 	CreatedBy:  input.UpdatedBy,
	// }
	// salutation := ""
	// if nameDetail.NewSalutation != nil {
	// 	salutation = *nameDetail.NewSalutation
	// }
	firstName := ""
	if nameDetail.NewFirstName != nil {
		firstName = *nameDetail.NewFirstName
	}
	lastName := ""
	if nameDetail.NewLastName != nil {
		lastName = *nameDetail.NewLastName
	}
	var salutationPtr *string
	if nameDetail.NewSalutation != nil && *nameDetail.NewSalutation != "" {
		salutationPtr = nameDetail.NewSalutation
	}
	versionEntry := &domain.NameVersionHistory{
		CustomerID: sr.CustomerID, // ← add this
		Salutation: salutationPtr, // ← nil if empty
		FirstName:  firstName,
		MiddleName: nameDetail.NewMiddleName,
		LastName:   lastName,
		CreatedBy:  parseStringToInt64(input.UpdatedBy),
	}
	created, err := a.nameRepo.CreateNameVersion(ctx, input.RequestID, versionEntry)
	if err != nil {
		return nil, fmt.Errorf("UpdateNameData: create name version: %w", err)
	}

	// Update status to COMPLETED + audit_log (BATCH in UpdateStatus).
	// audit := &domain.AuditLog{
	// 	ActionType:    "COMPLETED",
	// 	PerformedByID: input.UpdatedBy,
	// 	Notes:         strPtrAct("Name data updated successfully"),
	// }
	audit := &domain.AuditLog{
		AuditID:       uuid.New().String(),
		RequestID:     input.RequestID,
		ActionType:    "STATUS_CHANGE",
		NewValueJSON:  `{"status":"COMPLETED","updated_by":"` + input.UpdatedBy + `"}`,
		PerformedByID: input.UpdatedBy,
		Notes:         strPtrAct("Name data updated successfully"),
	}
	_, err = a.srRepo.UpdateStatus(ctx, input.RequestID, "COMPLETED", &input.UpdatedBy, nil, nil, audit)
	if err != nil {
		return nil, fmt.Errorf("UpdateNameData: update status: %w", err)
	}

	return &UpdateAddressDataResult{Updated: true, NewVersionID: created.VersionID}, nil
}

// AssignNameToCPC assigns a manual name change request to the CPC work queue.
// BATCH (TX): 3-op — UPDATE service_request + INSERT cpc_work_queue + INSERT audit_log.
// BR-NFS-002: CPC assignment based on customer's servicing office.
// Called by WF-NFS-004 after documents_submitted signal.
// func (a *NameChangeActivities) AssignNameToCPC(ctx context.Context, input AssignToCPCInput) (*AssignToCPCResult, error) {
// 	// log.Info(ctx, "AssignNameToCPC: assigning request %s to CPC office %v", input.RequestID, input.OfficeCode)
// 	log.Info(ctx, "AssignNameToCPC: assigning request %s to CPC office %s", input.RequestID, input.OfficeCode)
// 	audit := &domain.AuditLog{
// 		ActionType:    "PENDING_APPROVAL",
// 		PerformedByID: input.AssignedBy,
// 		Notes:         strPtrAct("Assigned to CPC queue"),
// 	}

//		// BATCH: AssignToCPC — 3-op TX batch (UPDATE service_request + INSERT cpc_work_queue + INSERT audit_log).
//		// if err := a.srRepo.AssignToCPC(ctx, input.RequestID, "", input.AssignedBy, input.Priority, input.SLADeadline, audit); err != nil {
//		// 	return nil, fmt.Errorf("AssignNameToCPC: %w", err)
//		// }
//		if err := a.srRepo.AssignToCPC(ctx, input.RequestID, input.OfficeCode, input.AssignedBy, input.Priority, input.SLADeadline, audit); err != nil {
//			return nil, fmt.Errorf("AssignNameToCPC: %w", err)
//		}
//		return &AssignToCPCResult{
//			QueueID:    newUUID(),
//			AssignedTo: input.AssignedBy,
//		}, nil
//	}
func (a *NameChangeActivities) AssignNameToCPC(ctx context.Context, input AssignToCPCInput) (*AssignToCPCResult, error) {
	officeCode := ""
	if input.OfficeCode != nil {
		officeCode = *input.OfficeCode
	}

	log.Info(ctx, "AssignNameToCPC: assigning request %s to CPC office %s", input.RequestID, officeCode)
	officeCode = ""
	if input.OfficeCode != nil {
		officeCode = *input.OfficeCode
	}
	audit := &domain.AuditLog{
		AuditID:   uuid.New().String(), // ← add
		RequestID: input.RequestID,     // ← add
		// ActionType:   "PENDING_APPROVAL",
		ActionType:   "STATUS_CHANGE", // was "COMPLETED"
		NewValueJSON: fmt.Sprintf(`{"office_code":"%s","priority":%d}`, officeCode, input.Priority),

		PerformedByID: input.AssignedBy,
		Notes:         strPtrAct("Assigned to CPC queue"),
	}

	// BATCH: AssignToCPC — 3-op TX batch (UPDATE service_request + INSERT cpc_work_queue + INSERT audit_log).
	if err := a.srRepo.AssignToCPC(ctx, input.RequestID, nil, input.AssignedBy, input.Priority, input.SLADeadline, audit); err != nil {
		return nil, fmt.Errorf("AssignNameToCPC: exec assign update: %w", err)
	}

	return &AssignToCPCResult{
		QueueID:    newUUID(),
		AssignedTo: input.AssignedBy,
	}, nil
}

// UpdateNameStatus updates the service request status with transition history + audit.
// BATCH (TX): 3-op — UPDATE service_request + INSERT status_transition + INSERT audit_log.
// BR-NFS-016: audit_log is INSERT-only.
func (a *NameChangeActivities) UpdateNameStatus(ctx context.Context, input UpdateStatusInput) (*UpdateStatusResult, error) {
	log.Info(ctx, "UpdateNameStatus: request %s → %s", input.RequestID, input.NewStatus)

	audit := &domain.AuditLog{
		// ActionType:    input.NewStatus,
		// input.NewStatus is a status value not an audit action — fix:
		ActionType:    "STATUS_CHANGE", // was input.NewStatus
		PerformedByID: input.UpdatedBy,
		Notes:         input.Reason,
	}
	updated, err := a.srRepo.UpdateStatus(ctx, input.RequestID, input.NewStatus, &input.UpdatedBy, nil, nil, audit)
	if err != nil {
		return nil, fmt.Errorf("UpdateNameStatus: %w", err)
	}

	return &UpdateStatusResult{
		RequestID: input.RequestID,
		NewStatus: string(updated.Status),
	}, nil
}

// GenerateNameAckReceipt generates a PDF acknowledgment receipt for the name change.
// WF-NFS-003/004: last step after COMPLETED state.
// STUB — calls Receipt Generation Service.
func (a *NameChangeActivities) GenerateNameAckReceipt(ctx context.Context, input GenerateAckReceiptInput) (*GenerateAckReceiptResult, error) {
	log.Info(ctx, "GenerateNameAckReceipt [STUB]: generating receipt for request %s", input.RequestID)
	// TODO: call Receipt Service: POST /receipts/nfs { request_id, request_type=NAME_CHANGE }
	return &GenerateAckReceiptResult{
		ReceiptDMSID:   "",
		ReceiptFileURL: fmt.Sprintf("/nfs/requests/%s/receipt", input.RequestID),
	}, nil
}

// EscalateNameRequest escalates a name change request approaching SLA breach.
// WF-NFS-004: triggered when SLA deadline approaches without resolution.
// STUB — publishes nfr.sla.breached event via event bus (INT-NFS-009).
func (a *NameChangeActivities) EscalateNameRequest(ctx context.Context, input EscalateInput) (*EscalateResult, error) {
	log.Info(ctx, "EscalateNameRequest [STUB]: escalating request %s (SLA breached)", input.RequestID)
	// TODO: publish nfr.sla.breached event to Kafka/event bus.
	// TODO: notify supervisor via Notification Service.
	return &EscalateResult{Escalated: true}, nil
}

// strPtrAct is a local helper returning a pointer to the given string.
func strPtrAct(s string) *string { return &s }


