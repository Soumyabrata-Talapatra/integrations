// Package activities implements Temporal activities for mobile change workflows
// in the customer-nfs-service.
//
// WF-NFS-006: Mobile Change (OTP-based verification)
//
//	Sequence: StoreWorkflowState → RequestMobileOTP → [Wait otp_submitted signal, 15 min timeout] →
//	          VerifyMobileOTP → UpdateMobileData → GenerateAckReceipt → UpdateStatus(COMPLETED) →
//	          NotifyPolicyManagement
//
// WORKFLOW STATE NOTE: StoreWorkflowState is called immediately after workflow
//
//	start to persist workflow_id + workflow_run_id into nfs.service_request.
//	This enables HTTP signal endpoints (OTP verify) to locate the
//	running workflow via temporal.Client.SignalWorkflow.
package activities

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.temporal.io/sdk/activity"

	config "gitlab.cept.gov.in/it-2.0-common/api-config"
	log "gitlab.cept.gov.in/it-2.0-common/n-api-log"

	"customer-nfs-service/core/domain"
	"customer-nfs-service/repo/postgres"
)

// MobileChangeActivities holds all activity implementations for mobile change.
type MobileChangeActivities struct {
	srRepo *postgres.ServiceRequestRepository
	mcRepo *postgres.MobileChangeRepository
	cfg    *config.Config
}

// NewMobileChangeActivities creates a new MobileChangeActivities instance.
func NewMobileChangeActivities(
	srRepo *postgres.ServiceRequestRepository,
	mcRepo *postgres.MobileChangeRepository,
	cfg *config.Config,
) *MobileChangeActivities {
	return &MobileChangeActivities{
		srRepo: srRepo,
		mcRepo: mcRepo,
		cfg:    cfg,
	}
}

// ---------------------------------------------------------------------------
// RequestMobileOTP initiates an OTP request to the new mobile number.
// WF-NFS-006 Step 2: OTP dispatched before waiting for otp_submitted signal.
// STUB — calls SMS Gateway Service.
// FR-NFS-006: OTP-based mobile change.
// ---------------------------------------------------------------------------
func (a *MobileChangeActivities) RequestMobileOTP(ctx context.Context, input AadhaarOTPRequestInput) (*AadhaarOTPRequestResult, error) {
	log.Info(ctx, "RequestMobileOTP [STUB]: dispatching OTP to mobile for customer %d", input.CustomerID)
	// TODO: call SMS Gateway Service: POST /sms/otp { customer_id, request_id, mobile_number }
	// Need to fetch new mobile number from mobile_change_detail table

	return &AadhaarOTPRequestResult{
		OTPReferenceID: "stub-mobile-txn-" + input.RequestID,
		ExpiresAt:      time.Now().Add(15 * time.Minute), // 15-min OTP window (WF-NFS-006)
	}, nil
}

// ---------------------------------------------------------------------------
// VerifyMobileOTP verifies the submitted OTP.
// WF-NFS-006 Step 4: Called after otp_submitted signal received.
// ---------------------------------------------------------------------------
func (a *MobileChangeActivities) VerifyMobileOTP(ctx context.Context, input AadhaarOTPVerifyInput) (*AadhaarOTPVerifyResult, error) {
	log.Info(ctx, "VerifyMobileOTP: verifying OTP for request %s", input.RequestID)
	// TODO: call SMS Gateway Service: POST /sms/verify { txn_id, otp }

	// Generate a transaction ID (in real implementation, this would come from SMS Gateway)
	txnID := "SMS-TXN-" + input.RequestID

	// Store transaction ID in mobile_change_detail
	if err := a.mcRepo.UpdateAadhaarTxnID(ctx, input.RequestID, txnID, fmt.Sprintf("%d", input.CustomerID)); err != nil {
		log.Error(ctx, "VerifyMobileOTP: failed to update aadhaar_txn_id: %v", err)
		// Continue even if update fails
	}

	return &AadhaarOTPVerifyResult{
		Verified:     true,
		AadhaarTxnID: txnID,
	}, nil
}

// ---------------------------------------------------------------------------
// MobileUpdateStatus updates nfs.service_request.status for mobile change.
// WF-NFS-006: Used for status transitions (CREATED → DOCUMENTS_EXPIRED, COMPLETED).
// BR-NFS-012: Batched with status_transition_history + audit_log.
// ---------------------------------------------------------------------------
func (a *MobileChangeActivities) MobileUpdateStatus(ctx context.Context, input UpdateStatusInput) error {
	activity.GetLogger(ctx).Info("MobileUpdateStatus", "requestID", input.RequestID, "newStatus", input.NewStatus)

	// Get current status from service_request
	sr, err := a.srRepo.GetByID(ctx, input.RequestID)
	if err != nil {
		return fmt.Errorf("MobileUpdateStatus: get service request: %w", err)
	}

	input.FromStatus = sr.Status

	// Create transition and audit objects
	transitionID := uuid.New().String()
	auditID := uuid.New().String()
	now := time.Now().UTC()

	transition := &domain.StatusTransitionHistory{
		TransitionID:     transitionID,
		RequestID:        input.RequestID,
		FromStatus:       &input.FromStatus,
		ToStatus:         input.NewStatus,
		TransitionedBy:   input.UpdatedBy,
		TransitionedAt:   now,
		TransitionReason: input.Reason,
	}

	oldValue := fmt.Sprintf(`{"status":"%s"}`, input.FromStatus)
	newValue := fmt.Sprintf(`{"status":"%s"}`, input.NewStatus)
	audit := &domain.AuditLog{
		AuditID:       auditID,
		RequestID:     input.RequestID,
		ActionType:    "STATUS_CHANGE",
		OldValue:      &oldValue,
		NewValueJSON:  newValue,
		PerformedByID: input.UpdatedBy,
		PerformedAt:   now,
		Channel:       &input.Channel,
		OfficeCode:    input.OfficeCode,
		Notes:         input.Reason,
	}

	// Update status (batched with status_transition_history + audit_log)
	if _, err := a.srRepo.UpdateStatus(ctx, input.RequestID, input.NewStatus, &input.UpdatedBy, input.Reason, transition, audit); err != nil {
		return fmt.Errorf("MobileUpdateStatus: update status: %w", err)
	}

	return nil
}

// ---------------------------------------------------------------------------
// UpdateMobileData updates the customer's mobile number in the policy system.
// WF-NFS-006 Step 5: Called after successful OTP verification.
// ---------------------------------------------------------------------------
func (a *MobileChangeActivities) UpdateMobileData(ctx context.Context, input UpdateMobileDataInput) (*UpdateMobileDataResult, error) {
	log.Info(ctx, "UpdateMobileData: updating mobile number for customer %d", input.CustomerID)

	// Fetch mobile change detail to get the new mobile number
	detail, err := a.mcRepo.GetByRequestID(ctx, input.RequestID)
	if err != nil {
		return nil, fmt.Errorf("UpdateMobileData: get mobile change detail: %w", err)
	}
	if detail == nil {
		return nil, fmt.Errorf("UpdateMobileData: mobile change detail not found for request %s", input.RequestID)
	}

	// TODO: call Policy Management Service: POST /policy/update-mobile
	// For now, log the actual mobile number
	activity.GetLogger(ctx).Info("UpdateMobileData",
		"requestID", input.RequestID,
		"customerID", input.CustomerID,
		"newMobileNumber", detail.NewMobileNumber,
		"updatedBy", input.UpdatedBy)

	// Create mobile version history
	history := &domain.MobileVersionHistory{
		VersionID:     uuid.New().String(),
		CustomerID:    input.CustomerID,
		RequestID:     &input.RequestID,
		MobileNumber:  detail.NewMobileNumber,
		VersionNumber: 1, // TODO: Get next version number from existing history
		IsActive:      true,
		EffectiveFrom: time.Now().UTC(),
		CreatedBy:     input.UpdatedBy,
	}

	if err := a.mcRepo.CreateMobileVersionHistory(ctx, history); err != nil {
		log.Error(ctx, "UpdateMobileData: failed to create mobile version history: %v", err)
		// Continue even if version history fails
	}

	return &UpdateMobileDataResult{
		Updated:         true,
		NewVersionID:    history.VersionID,
		NewMobileNumber: &detail.NewMobileNumber,
	}, nil
}

// strPtrAct returns a pointer to the given string.
func strPtrAct(s string) *string { return &s }
