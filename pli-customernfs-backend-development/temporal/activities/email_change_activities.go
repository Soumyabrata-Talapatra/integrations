// Package activities implements Temporal activities for email change workflows
// in the customer-nfs-service.
//
// WF-NFS-007: Email Change (OTP/verification-based)
//
//	Sequence: StoreWorkflowState → RequestEmailOTP → [Wait otp_submitted signal, 15 min timeout] →
//	          VerifyEmailOTP → UpdateEmailData → GenerateAckReceipt → UpdateStatus(COMPLETED) →
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

// EmailChangeActivities holds all activity implementations for email change.
type EmailChangeActivities struct {
	srRepo *postgres.ServiceRequestRepository
	cfg    *config.Config
}

// NewEmailChangeActivities creates a new EmailChangeActivities instance.
func NewEmailChangeActivities(
	srRepo *postgres.ServiceRequestRepository,
	cfg *config.Config,
) *EmailChangeActivities {
	return &EmailChangeActivities{
		srRepo: srRepo,
		cfg:    cfg,
	}
}

// ---------------------------------------------------------------------------
// RequestEmailOTP initiates an OTP request to the new email address.
// WF-NFS-007 Step 2: OTP dispatched before waiting for otp_submitted signal.
// STUB — calls Email Gateway Service.
// FR-NFS-007: OTP-based email change.
// ---------------------------------------------------------------------------
func (a *EmailChangeActivities) RequestEmailOTP(ctx context.Context, input AadhaarOTPRequestInput) (*AadhaarOTPRequestResult, error) {
	log.Info(ctx, "RequestEmailOTP [STUB]: dispatching OTP to email for customer %d", input.CustomerID)
	// TODO: call Email Gateway Service: POST /email/otp { customer_id, request_id, email_address }
	// Need to fetch new email from email_change_detail table
	
	return &AadhaarOTPRequestResult{
		OTPReferenceID: "stub-email-txn-" + input.RequestID,
		ExpiresAt:      time.Now().Add(15 * time.Minute), // 15-min OTP window (WF-NFS-007)
	}, nil
}

// ---------------------------------------------------------------------------
// VerifyEmailOTP verifies the submitted OTP.
// WF-NFS-007 Step 4: Called after otp_submitted signal received.
// STUB — calls Email Gateway Service.
// ---------------------------------------------------------------------------
func (a *EmailChangeActivities) VerifyEmailOTP(ctx context.Context, input AadhaarOTPVerifyInput) (*AadhaarOTPVerifyResult, error) {
	log.Info(ctx, "VerifyEmailOTP [STUB]: verifying OTP for request %s", input.RequestID)
	// TODO: call Email Gateway Service: POST /email/verify { txn_id, otp }
	
	// Generate a stub transaction ID
	txnID := "EMAIL-TXN-STUB-" + input.RequestID
	
	// TODO: Store transaction ID in email_change_detail
	// Need to create EmailChangeRepository first
	
	return &AadhaarOTPVerifyResult{
		Verified:     true,
		AadhaarTxnID: txnID,
	}, nil
}

// ---------------------------------------------------------------------------
// EmailUpdateStatus updates nfs.service_request.status for email change.
// WF-NFS-007: Used for status transitions (CREATED → DOCUMENTS_EXPIRED, COMPLETED).
// BR-NFS-012: Batched with status_transition_history + audit_log.
// ---------------------------------------------------------------------------
func (a *EmailChangeActivities) EmailUpdateStatus(ctx context.Context, input UpdateStatusInput) error {
	activity.GetLogger(ctx).Info("EmailUpdateStatus", "requestID", input.RequestID, "newStatus", input.NewStatus)
	
	// Get current status from service_request
	sr, err := a.srRepo.GetByID(ctx, input.RequestID)
	if err != nil {
		return fmt.Errorf("EmailUpdateStatus: get service request: %w", err)
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
		return fmt.Errorf("EmailUpdateStatus: update status: %w", err)
	}
	
	return nil
}

// ---------------------------------------------------------------------------
// UpdateEmailData updates the customer's email address in the policy system.
// WF-NFS-007 Step 5: Called after successful OTP verification.
// STUB — calls Policy Management Service.
// ---------------------------------------------------------------------------
func (a *EmailChangeActivities) UpdateEmailData(ctx context.Context, input UpdateEmailDataInput) error {
	log.Info(ctx, "UpdateEmailData [STUB]: updating email address for customer %d", input.CustomerID)
	
	// TODO: call Policy Management Service: POST /policy/update-email
	// Need to fetch new email from email_change_detail table
	
	// For now, just log the operation
	activity.GetLogger(ctx).Info("UpdateEmailData stub", 
		"requestID", input.RequestID, 
		"customerID", input.CustomerID,
		"updatedBy", input.UpdatedBy)
	
	return nil
}