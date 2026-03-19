package handler

import (
	"fmt"
	"net/http"
	"time"

	"policy-issue-service/core/domain"
	"policy-issue-service/core/port"
	resp "policy-issue-service/handler/response"
	repo "policy-issue-service/repo/postgres"

	"github.com/jackc/pgx/v5"
	apierrors "gitlab.cept.gov.in/it-2.0-common/n-api-errors"
	log "gitlab.cept.gov.in/it-2.0-common/n-api-log"
	serverHandler "gitlab.cept.gov.in/it-2.0-common/n-api-server/handler"
	serverRoute "gitlab.cept.gov.in/it-2.0-common/n-api-server/route"
)

// StatusHandler handles status and tracking HTTP endpoints
// Phase 8: [STATUS-POL-001] to [STATUS-POL-003]
type StatusHandler struct {
	*serverHandler.Base
	proposalRepo *repo.ProposalRepository
	documentRepo *repo.DocumentRepository
}

// NewStatusHandler creates a new StatusHandler instance
func NewStatusHandler(proposalRepo *repo.ProposalRepository, documentRepo *repo.DocumentRepository) *StatusHandler {
	base := serverHandler.New("Status").SetPrefix("/v1").AddPrefix("")
	return &StatusHandler{
		Base:         base,
		proposalRepo: proposalRepo,
		documentRepo: documentRepo,
	}
}

// Routes returns the routes for the StatusHandler
func (h *StatusHandler) Routes() []serverRoute.Route {
	return []serverRoute.Route{
		serverRoute.GET("/proposals/:proposal_id/status", h.GetProposalStatus).Name("Get Proposal Status"),
		serverRoute.GET("/proposals/:proposal_id/timeline", h.GetProposalTimeline).Name("Get Proposal Timeline"),
		serverRoute.GET("/policies/:policy_id/status", h.GetPolicyStatus).Name("Get Policy Status"),
		
	}
}

// statusDescriptions provides human-readable descriptions for proposal statuses
// [BR-POL-015] Proposal State Machine
var statusDescriptions = map[domain.ProposalStatus]string{
	domain.ProposalStatusDraft:           "Proposal created, pending indexing",
	domain.ProposalStatusIndexed:         "Proposal indexed at counter, awaiting data entry",
	domain.ProposalStatusDataEntry:       "Data entry in progress at CPC",
	domain.ProposalStatusQCPending:       "Submitted for Quality Check review",
	domain.ProposalStatusQCApproved:      "Quality Check passed, proceeding to next stage",
	domain.ProposalStatusQCRejected:      "Rejected during Quality Check review",
	domain.ProposalStatusQCReturned:      "Returned from QC for corrections",
	domain.ProposalStatusPendingMedical:  "Awaiting medical examination results",
	domain.ProposalStatusMedicalApproved: "Medical examination approved",
	domain.ProposalStatusMedicalRejected: "Medical examination rejected",
	domain.ProposalStatusApprovalPending: "Pending approval by authorized approver",
	domain.ProposalStatusApproved:        "Proposal approved for policy issuance",
	domain.ProposalStatusRejected:        "Proposal rejected by approver",
	domain.ProposalStatusIssued:          "Policy issued, bond generated",
	domain.ProposalStatusDispatched:      "Policy bond dispatched to customer",
	domain.ProposalStatusFreeLookActive:  "Free Look Period active — customer may cancel",
	domain.ProposalStatusActive:          "Policy is active",
	domain.ProposalStatusFLCCancelled:    "Policy cancelled during Free Look Period",
	domain.ProposalStatusCancelledDeath:  "Proposal cancelled due to death of insured",
}

// GetProposalStatus retrieves the current status of a proposal
// [STATUS-POL-001] Get proposal status
// [BR-POL-015] Proposal State Machine
func (h *StatusHandler) GetProposalStatus(sctx *serverRoute.Context, req ProposalIDUri) (*resp.ProposalStatusResponse, error) {
	proposal, err := h.proposalRepo.GetProposalByID(sctx.Ctx, req.ProposalID)
	if err != nil {
		log.Error(sctx.Ctx, "[STATUS-POL-001] Proposal %d not found: %v", req.ProposalID, err)
		return nil, err
	}

	description := statusDescriptions[proposal.Status]
	if description == "" {
		description = string(proposal.Status)
	}

	return &resp.ProposalStatusResponse{
		StatusCodeAndMessage: port.StatusCodeAndMessage{
			StatusCode: http.StatusOK,
			Message:    "Proposal status retrieved successfully",
		},
		ProposalID:        proposal.ProposalID,
		ProposalNumber:    proposal.ProposalNumber,
		Status:            string(proposal.Status),
		StatusDescription: description,
		LastUpdated:       proposal.UpdatedAt,
	}, nil
}

// GetProposalTimeline retrieves the complete timeline of proposal status changes
// [STATUS-POL-002] Get proposal timeline
// [FR-POL-033] Proposal Summary View
// Returns complete timeline with timestamps, actors, and duration between steps
func (h *StatusHandler) GetProposalTimeline(sctx *serverRoute.Context, req ProposalIDUri) (*resp.ProposalTimelineResponse, error) {
	// Step 1: Get proposal for metadata
	proposal, err := h.proposalRepo.GetProposalByID(sctx.Ctx, req.ProposalID)
	if err != nil {
		log.Error(sctx.Ctx, "[STATUS-POL-002] Proposal %d not found: %v", req.ProposalID, err)
		return nil, err
	}

	// Step 2: Get status history from DB
	history, err := h.documentRepo.GetStatusHistory(sctx.Ctx, req.ProposalID)
	if err != nil {
		log.Error(sctx.Ctx, "[STATUS-POL-002] Error fetching status history for proposal %d: %v", req.ProposalID, err)
		return nil, err
	}

	// Step 3: Build timeline entries
	timeline := make([]resp.TimelineEntry, len(history))
	for i, h := range history {
		var fromStatus *string
		if h.FromStatus != nil {
			fs := string(*h.FromStatus)
			fromStatus = &fs
		}

		// Determine step label based on to_status
		step := mapStatusToStep(h.ToStatus)

		// Determine timeline status (COMPLETED for past, CURRENT for latest)
		timelineStatus := "COMPLETED"
		if i == len(history)-1 {
			timelineStatus = "CURRENT"
		}

		// Calculate duration from previous step
		var duration string
		if i > 0 {
			dur := h.ChangedAt.Sub(history[i-1].ChangedAt)
			duration = formatDuration(dur)
		}

		timestamp := h.ChangedAt
		timeline[i] = resp.TimelineEntry{
			Step:       step,
			Status:     timelineStatus,
			FromStatus: fromStatus,
			ToStatus:   string(h.ToStatus),
			Timestamp:  &timestamp,
			Actor:      fmt.Sprintf("User-%d", h.ChangedBy),
			Comments:   h.Comments,
			Duration:   duration,
		}
	}

	return &resp.ProposalTimelineResponse{
		StatusCodeAndMessage: port.StatusCodeAndMessage{
			StatusCode: http.StatusOK,
			Message:    "Proposal timeline retrieved successfully",
		},
		ProposalID:     proposal.ProposalID,
		ProposalNumber: proposal.ProposalNumber,
		Timeline:       timeline,
	}, nil
}

// GetPolicyStatus retrieves the status of an issued policy
// [STATUS-POL-003] Get policy status
// [FR-POL-025] Policy Kit & Dispatch
// NOTE: policy_id is the proposal_id for proposals that have reached ISSUED stage
func (h *StatusHandler) GetPolicyStatus(sctx *serverRoute.Context, req PolicyIDUri) (*resp.PolicyStatusResponse, error) {
	// In this system, policy_id maps to the proposal_id for issued proposals.
	// Policy is a lifecycle stage of a proposal, not a separate entity.
	proposal, err := h.proposalRepo.GetProposalByID(sctx.Ctx, req.PolicyID)
	// if err != nil {
	// 	log.Error(sctx.Ctx, "[STATUS-POL-003] Policy/Proposal %d not found: %v", req.PolicyID, err)
	// 	return nil, err
	// }
	if err != nil {

		if err == pgx.ErrNoRows {
			return nil, apierrors.HandleErrorWithStatusCodeAndMessage(
				apierrors.HTTPErrorNotFound,
				fmt.Sprintf("Policy %d not found", req.PolicyID),
				err,
			)
		}

		log.Error(sctx.Ctx, "[STATUS-POL-003] Failed fetching proposal %d: %v", req.PolicyID, err)
		return nil, apierrors.HandleErrorWithStatusCodeAndMessage(
			apierrors.HTTPErrorServerError,
			"Failed to fetch policy details",
			err,
		)
	}
	// Validate that this proposal has reached policy stage
	validPolicyStatuses := map[domain.ProposalStatus]bool{
		domain.ProposalStatusIssued:         true,
		domain.ProposalStatusDispatched:     true,
		domain.ProposalStatusFreeLookActive: true,
		domain.ProposalStatusActive:         true,
		domain.ProposalStatusFLCCancelled:   true,
	}

	if !validPolicyStatuses[proposal.Status] {
		// return nil, fmt.Errorf("[STATUS-POL-003] proposal %d has not been issued as a policy yet (status: %s)", req.PolicyID, proposal.Status)
	return nil, apierrors.HandleErrorWithStatusCodeAndMessage(
			apierrors.HTTPErrorBadRequest,
			fmt.Sprintf(
				"[STATUS-POL-003] Proposal %d has not been issued as a policy yet (status: %s)",
				req.PolicyID,
				proposal.Status,
			),
			nil,
		)
	}

	// Build policy status description
	description := statusDescriptions[proposal.Status]
	if description == "" {
		description = string(proposal.Status)
	}

	// Policy number (set after approval via BR-POL-015)
	// Fetched separately since Proposal struct doesn't include policy_number field
	policyNumber, err := h.documentRepo.GetPolicyNumberByProposalID(sctx.Ctx, req.PolicyID)
	if err != nil {
		log.Warn(sctx.Ctx, "[STATUS-POL-003] Could not fetch policy number for proposal %d: %v", req.PolicyID, err)
		policyNumber = fmt.Sprintf("POL-%d", proposal.ProposalID)
	}
	if policyNumber == "" {
		policyNumber = fmt.Sprintf("POL-%d", proposal.ProposalID)
	}

	return &resp.PolicyStatusResponse{
		StatusCodeAndMessage: port.StatusCodeAndMessage{
			StatusCode: http.StatusOK,
			Message:    "Policy status retrieved successfully",
		},
		PolicyID:          proposal.ProposalID,
		PolicyNumber:      policyNumber,
		Status:            string(proposal.Status),
		StatusDescription: description,
		LastUpdated:       proposal.UpdatedAt,
	}, nil
}

// mapStatusToStep maps a proposal status to a human-readable workflow step label
func mapStatusToStep(status domain.ProposalStatus) string {
	switch status {
	case domain.ProposalStatusDraft:
		return "Proposal Created"
	case domain.ProposalStatusIndexed:
		return "Counter Indexing"
	case domain.ProposalStatusDataEntry:
		return "CPC Data Entry"
	case domain.ProposalStatusQCPending:
		return "Submitted for QC"
	case domain.ProposalStatusQCApproved:
		return "QC Approved"
	case domain.ProposalStatusQCRejected:
		return "QC Rejected"
	case domain.ProposalStatusQCReturned:
		return "Returned for Corrections"
	case domain.ProposalStatusPendingMedical:
		return "Medical Examination"
	case domain.ProposalStatusMedicalApproved:
		return "Medical Approved"
	case domain.ProposalStatusMedicalRejected:
		return "Medical Rejected"
	case domain.ProposalStatusApprovalPending:
		return "Approval Review"
	case domain.ProposalStatusApproved:
		return "Approved"
	case domain.ProposalStatusRejected:
		return "Rejected"
	case domain.ProposalStatusIssued:
		return "Policy Issued"
	case domain.ProposalStatusDispatched:
		return "Policy Dispatched"
	case domain.ProposalStatusFreeLookActive:
		return "Free Look Period"
	case domain.ProposalStatusActive:
		return "Policy Active"
	case domain.ProposalStatusFLCCancelled:
		return "FLC Cancelled"
	case domain.ProposalStatusCancelledDeath:
		return "Cancelled (Death)"
	default:
		return string(status)
	}
}

// formatDuration formats a time.Duration into a human-readable string
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh %dm", int(d.Hours()), int(d.Minutes())%60)
	}
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	return fmt.Sprintf("%dd %dh", days, hours)
}
