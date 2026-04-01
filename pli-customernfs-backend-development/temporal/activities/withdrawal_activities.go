// Package activities — withdrawal activities for WF-NFS-005.
//
// WF-NFS-005: Customer NFS Request Withdrawal
//
// BR-NFS-013: Withdrawal only allowed for CREATED / PENDING_DOCUMENTS / PENDING_APPROVAL.
//
//	NOT allowed for IN_PROGRESS, COMPLETED, REJECTED, WITHDRAWN, DOCUMENTS_EXPIRED.
//	If partial_processing_flag=TRUE → manual CPC approval required.
//	If partial_processing_flag=FALSE → auto-approved immediately.
//
// BR-NFS-016: All withdrawal transitions are captured in audit_log (INSERT-only).
package activities

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"go.temporal.io/sdk/activity"

	config "gitlab.cept.gov.in/it-2.0-common/api-config"
	log "gitlab.cept.gov.in/it-2.0-common/n-api-log"

	"customer-nfs-service/core/domain"
	"customer-nfs-service/repo/postgres"
)

// WithdrawalActivities bundles all Temporal activities for WF-NFS-005.
// FR-NFS-007 (Withdrawal), BR-NFS-013 (Eligibility), BR-NFS-016 (Audit)
type WithdrawalActivities struct {
	srRepo    *postgres.ServiceRequestRepository
	auditRepo *postgres.AuditLogRepository
	cfg       config.Config
}

// NewWithdrawalActivities creates the activities struct for DI injection.
func NewWithdrawalActivities(
	srRepo *postgres.ServiceRequestRepository,
	auditRepo *postgres.AuditLogRepository,
	cfg config.Config,
) *WithdrawalActivities {
	return &WithdrawalActivities{
		srRepo:    srRepo,
		auditRepo: auditRepo,
		cfg:       cfg,
	}
}

// strPtrWithdraw is a local helper returning a pointer to the given string.
// (cannot reuse strPtrAct from name_change_activities.go as it would be a duplicate symbol)
var strPtrWithdraw = func(s string) *string { return &s }

// withdrawableStatuses contains the statuses from which withdrawal is permitted.
// BR-NFS-013: Terminal and in-progress states are excluded.
var withdrawableStatuses = map[string]bool{
	"CREATED":           true,
	"PENDING_DOCUMENTS": true,
	"PENDING_APPROVAL":  true,
}

// CheckWithdrawalEligibility verifies that the service request is in a
// withdrawable status and determines whether CPC approval is required.
// BR-NFS-013: withdrawal rejected for IN_PROGRESS / COMPLETED / REJECTED /
// WITHDRAWN / DOCUMENTS_EXPIRED states.
// Returns ApprovalType="AUTO" when partial_processing_flag=FALSE,
// "MANUAL" when partial_processing_flag=TRUE.
func (a *WithdrawalActivities) CheckWithdrawalEligibility(
	ctx context.Context,
	input CheckWithdrawalEligibilityInput,
) (*CheckWithdrawalEligibilityResult, error) {
	logger := activity.GetLogger(ctx)
	log.Info(ctx, "CheckWithdrawalEligibility: checking request %s", input.RequestID)

	sr, err := a.srRepo.GetByID(ctx, input.RequestID)
	if err != nil {
		return nil, fmt.Errorf("CheckWithdrawalEligibility: fetch request: %w", err)
	}

	status := string(sr.Status)
	if !withdrawableStatuses[status] {
		logger.Warn("CheckWithdrawalEligibility: not eligible",
			"requestID", input.RequestID,
			"status", status,
		)
		return &CheckWithdrawalEligibilityResult{
			Eligible: false,
			Reason:   fmt.Sprintf("withdrawal not allowed for status %q (BR-NFS-013)", status),
		}, nil
	}

	// BR-NFS-013: partial processing requires manual CPC approval path.
	approvalType := "AUTO"
	if sr.PartialProcessingFlag {
		approvalType = "MANUAL"
	}

	logger.Info("CheckWithdrawalEligibility: eligible",
		"requestID", input.RequestID,
		"approvalType", approvalType,
	)
	return &CheckWithdrawalEligibilityResult{
		Eligible:     true,
		ApprovalType: approvalType,
	}, nil
}

// ProcessWithdrawal transitions the service request to WITHDRAWN status and
// records the action in the audit log.
// BR-NFS-013: handles both AUTO and MANUAL (CPC-approved) withdrawal paths.
// BR-NFS-016: audit log is INSERT-only — records action_type=WITHDRAWN.
// BATCH (TX): UpdateStatus internally does UPDATE service_request +
// INSERT status_transition_history + INSERT audit_log in a single round-trip.
func (a *WithdrawalActivities) ProcessWithdrawal(
	ctx context.Context,
	input ProcessWithdrawalInput,
) (*ProcessWithdrawalResult, error) {
	log.Info(ctx, "ProcessWithdrawal: withdrawing request %s (type=%s)", input.RequestID, input.WithdrawalType)

	// Determine the actor: approver if MANUAL, requester if AUTO.
	performedBy := input.RequestedBy
	if input.ApprovedBy != nil {
		performedBy = strconv.FormatInt(*input.ApprovedBy, 10)
	}

	audit := &domain.AuditLog{
		AuditID:       uuid.New().String(), // ← add
		RequestID:     input.RequestID,     // ← add
		ActionType:    "WITHDRAWN",
		NewValueJSON:  fmt.Sprintf(`{"withdrawal_type":%q,"reason":%q}`, input.WithdrawalType, input.WithdrawalReason),
		PerformedByID: performedBy,
		Notes:         strPtrWithdraw(fmt.Sprintf("Request withdrawn (%s): %s", input.WithdrawalType, input.WithdrawalReason)),
	}

	updated, err := a.srRepo.UpdateStatus(ctx, input.RequestID, "WITHDRAWN", &performedBy, nil, nil, audit)
	if err != nil {
		return nil, fmt.Errorf("ProcessWithdrawal: update status: %w", err)
	}

	log.Info(ctx, "ProcessWithdrawal: request %s withdrawn successfully", input.RequestID)
	return &ProcessWithdrawalResult{
		Processed: true,
		NewStatus: string(updated.Status),
	}, nil
}

// UpdateStatus updates the service request status with a status_transition_history
// entry and an audit_log record (INSERT-only per BR-NFS-016).
// Used in WF-NFS-005 to move to PENDING_WITHDRAWAL_APPROVAL on the manual path.
// BATCH (TX): UPDATE service_request + INSERT status_transition + INSERT audit_log.
func (a *WithdrawalActivities) UpdateStatus(
	ctx context.Context,
	input UpdateStatusInput,
) (*UpdateStatusResult, error) {
	log.Info(ctx, "WithdrawalActivities.UpdateStatus: request %s → %s", input.RequestID, input.NewStatus)

	notes := input.Reason
	now := time.Now().UTC()
	newValueJSON := fmt.Sprintf(`{"status":"%s"}`, input.NewStatus)
	audit := &domain.AuditLog{
		AuditID:       uuid.New().String(),
		RequestID:     input.RequestID,
		ActionType:    "STATUS_CHANGE",
		NewValueJSON:  newValueJSON,
		PerformedByID: input.UpdatedBy,
		PerformedAt:   now,
		Notes:         notes,
	}

	updated, err := a.srRepo.UpdateStatus(ctx, input.RequestID, input.NewStatus, &input.UpdatedBy, input.Reason, nil, audit)
	if err != nil {
		return nil, fmt.Errorf("WithdrawalActivities.UpdateStatus: %w", err)
	}

	return &UpdateStatusResult{
		RequestID: input.RequestID,
		NewStatus: string(updated.Status),
	}, nil
}
