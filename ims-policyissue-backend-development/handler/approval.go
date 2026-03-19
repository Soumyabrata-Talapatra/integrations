package handler

import (
	"fmt"
	"net/http"
	"policy-issue-service/core/domain"
	"policy-issue-service/core/port"
	resp "policy-issue-service/handler/response"
	repo "policy-issue-service/repo/postgres"
	"policy-issue-service/workflows"

	log "gitlab.cept.gov.in/it-2.0-common/api-log"
	serverHandler "gitlab.cept.gov.in/it-2.0-common/n-api-server/handler"
	serverRoute "gitlab.cept.gov.in/it-2.0-common/n-api-server/route"

	"go.temporal.io/sdk/client"
)

// ApprovalHandler handles approval-related HTTP endpoints
type ApprovalHandler struct {
	*serverHandler.Base

	temporalClient client.Client
	proposalRepo   *repo.ProposalRepository
}

// NewApprovalHandler creates a new ApprovalHandler instance
func NewApprovalHandler(temporalClient client.Client, proposalRepo *repo.ProposalRepository) *ApprovalHandler {
	base := serverHandler.New("Approval").SetPrefix("/v1").AddPrefix("")
	return &ApprovalHandler{
		Base:           base,
		temporalClient: temporalClient,
		proposalRepo:   proposalRepo,
	}
}

// Routes returns the routes for the ApprovalHandler
func (h *ApprovalHandler) Routes() []serverRoute.Route {
	return []serverRoute.Route{
		serverRoute.POST("/proposals/:proposal_id/qr-approve", h.QRApprove).Name("QR Approve"),
		serverRoute.POST("/proposals/:proposal_id/qr-reject", h.QRReject).Name("QR Reject"),
		serverRoute.POST("/proposals/:proposal_id/qr-return", h.QRReturn).Name("QR Return"),
		serverRoute.POST("/proposals/:proposal_id/approve", h.ApproverApprove).Name("Approver Approve"),
		serverRoute.POST("/proposals/:proposal_id/reject", h.ApproverReject).Name("Approver Reject"),
	}
}

// QRApprove approves a proposal at the QR level
// [POL-API-017] QR Approve
// State transition: QC_PENDING → QC_APPROVED (BR-POL-015)
func (h *ApprovalHandler) QRApprove(sctx *serverRoute.Context, req QRApproveRequest) (*resp.QRApproveResponse, error) {

	proposal, err := h.proposalRepo.GetProposalByID(sctx.Ctx, req.ProposalID)
	if err != nil {
		return nil, err
	}

	// Validate proposal is in QC_PENDING status
	if proposal.Status != domain.ProposalStatusQCPending {
		return &resp.QRApproveResponse{
			StatusCodeAndMessage: port.StatusCodeAndMessage{
				StatusCode: http.StatusBadRequest,
				Message:    fmt.Sprintf("Proposal must be in QC_PENDING status for QR approval. Current: %s", proposal.Status),
			},
		}, nil
	}

	// Record QC review decision
	if err := h.proposalRepo.RecordQCReview(sctx.Ctx, req.ProposalID, "APPROVED", req.Comments, req.ReviewerID); err != nil {
		return nil, err
	}

	// Persist status transition: QC_PENDING → QC_APPROVED
	// changedBy, _ := strconv.ParseInt(req.ReviewerID, 10, 64)
	// if err := h.proposalRepo.UpdateProposalStatus(sctx.Ctx, req.ProposalID,
	// 	domain.ProposalStatusQCApproved, "QC approved: "+req.Comments, changedBy); err != nil {
	// 	log.Error(sctx.Ctx, "Failed to update proposal status to QC_APPROVED", "proposalID", req.ProposalID, "error", err)
	// 	return nil, err
	// }
	// log.Info(sctx.Ctx, "WorkflowID value", "proposalID", req.ProposalID, "workflowID", proposal.WorkflowID)
	// Signal workflow
	if proposal.WorkflowID != nil {
		if err := h.temporalClient.SignalWorkflow(sctx.Ctx, *proposal.WorkflowID, "", workflows.SignalQRDecision, workflows.QRDecisionSignal{
			Decision:   "APPROVED",
			Comments:   req.Comments,
			ReviewerID: req.ReviewerID,
		}); err != nil {
			log.Error(sctx.Ctx, "Failed to signal workflow for QC approval", "proposalID", req.ProposalID, "error", err)
		}
	}

	return &resp.QRApproveResponse{
		StatusCodeAndMessage: port.StatusCodeAndMessage{
			StatusCode: http.StatusOK,
			Message:    "QR approval processed successfully",
		},
		Status: "success",
	}, nil
}

// QRReject rejects a proposal at the QR level
// [POL-API-018] QR Reject
// State transition: QC_PENDING → QC_REJECTED (BR-POL-015)
func (h *ApprovalHandler) QRReject(sctx *serverRoute.Context, req QRRejectRequest,
) (*resp.QRRejectResponse, error) {

	proposal, err := h.proposalRepo.GetProposalByID(sctx.Ctx, req.ProposalID)
	if err != nil {
		return nil, err
	}

	// Validate status
	if proposal.Status != domain.ProposalStatusQCPending {
		return &resp.QRRejectResponse{
			StatusCodeAndMessage: port.StatusCodeAndMessage{
				StatusCode: http.StatusBadRequest,
				Message: fmt.Sprintf(
					"Proposal must be in QC_PENDING status for QR rejection. Current: %s",
					proposal.Status,
				),
			},
		}, nil
	}

	// Record QC review
	err = h.proposalRepo.RecordQCReview(sctx.Ctx, req.ProposalID, "REJECTED", req.Comments,
		req.ReviewerID)

	if err != nil {
		log.Error(sctx.Ctx, "Failed to record QC review", "proposalID", req.ProposalID,
			"error", err)
		return nil, err
	}

	// Workflow must exist
	if proposal.WorkflowID == nil {
		return nil, fmt.Errorf("workflow not found for proposal")
	}

	// Signal workflow
	err = h.temporalClient.SignalWorkflow(
		sctx.Ctx,
		*proposal.WorkflowID,
		"",
		workflows.SignalQRDecision,
		workflows.QRDecisionSignal{
			Decision:   "REJECTED",
			Comments:   req.Comments,
			ReviewerID: req.ReviewerID,
		},
	)

	if err != nil {
		log.Error(sctx.Ctx, "Failed to signal workflow", "proposalID", req.ProposalID,
			"error", err)
		return nil, err
	}

	return &resp.QRRejectResponse{
		StatusCodeAndMessage: port.StatusCodeAndMessage{
			StatusCode: http.StatusOK,
			Message:    "QR rejection processed successfully",
		},
		Status: "success",
	}, nil
}

// func (h *ApprovalHandler) QRReject(sctx *serverRoute.Context, req QRRejectRequest) (*resp.QRRejectResponse, error) {

// 	proposal, err := h.proposalRepo.GetProposalByID(sctx.Ctx, req.ProposalID)
// 	if err != nil {
// 		return nil, err
// 	}

// 	// Validate proposal is in QC_PENDING status
// 	if proposal.Status != domain.ProposalStatusQCPending {
// 		return &resp.QRRejectResponse{
// 			StatusCodeAndMessage: port.StatusCodeAndMessage{
// 				StatusCode: http.StatusBadRequest,
// 				Message:    fmt.Sprintf("Proposal must be in QC_PENDING status for QR rejection. Current: %s", proposal.Status),
// 			},
// 		}, nil
// 	}

// 	// Record QC review decision
// 	if err := h.proposalRepo.RecordQCReview(sctx.Ctx, req.ProposalID, "REJECTED", req.Comments, req.ReviewerID); err != nil {
// 		return nil, err
// 	}

// 	// Persist status transition: QC_PENDING → QC_REJECTED
// 	// changedBy, _ := strconv.ParseInt(req.ReviewerID, 10, 64)
// 	// if err := h.proposalRepo.UpdateProposalStatus(sctx.Ctx, req.ProposalID,
// 	// 	domain.ProposalStatusQCRejected, "QC rejected: "+req.Comments, changedBy); err != nil {
// 	// 	log.Error(sctx.Ctx, "Failed to update proposal status to QC_REJECTED", "proposalID", req.ProposalID, "error", err)
// 	// 	return nil, err
// 	// }

// 	// Signal workflow
// 	if proposal.WorkflowID != nil {
// 		err := h.temporalClient.SignalWorkflow(sctx.Ctx, *proposal.WorkflowID, "", workflows.SignalQRDecision, workflows.QRDecisionSignal{
// 			Decision:   "REJECTED",
// 			Comments:   req.Comments,
// 			ReviewerID: req.ReviewerID,
// 		},
// 		)
// 		if err != nil {
// 			log.Error(sctx.Ctx, "Failed to signal workflow",
// 				"proposalID", req.ProposalID,
// 				"error", err)

// 			return nil, err
// 		}
// 	}

// 	return &resp.QRRejectResponse{
// 		StatusCodeAndMessage: port.StatusCodeAndMessage{
// 			StatusCode: http.StatusOK,
// 			Message:    "QR rejection processed successfully",
// 		},
// 		Status: "success",
// 	}, nil
// }

// QRReturn returns a proposal to data entry at the QR level
// [POL-API-019] QR Return
// State transition: QC_PENDING → QC_RETURNED (BR-POL-015)
func (h *ApprovalHandler) QRReturn(sctx *serverRoute.Context, req QRReturnRequest) (*resp.QRReturnResponse, error) {

	proposal, err := h.proposalRepo.GetProposalByID(sctx.Ctx, req.ProposalID)
	if err != nil {
		return nil, err
	}

	// Validate proposal is in QC_PENDING status
	if proposal.Status != domain.ProposalStatusQCPending {
		return &resp.QRReturnResponse{
			StatusCodeAndMessage: port.StatusCodeAndMessage{
				StatusCode: http.StatusBadRequest,
				Message:    fmt.Sprintf("Proposal must be in QC_PENDING status for QR return. Current: %s", proposal.Status),
			},
		}, nil
	}

	// Record QC review decision
	if err := h.proposalRepo.RecordQCReview(sctx.Ctx, req.ProposalID, "RETURNED", req.Comments, req.ReviewerID); err != nil {
		return nil, err
	}

	// Persist status transition: QC_PENDING → QC_RETURNED
	// changedBy, _ := strconv.ParseInt(req.ReviewerID, 10, 64)
	// if err := h.proposalRepo.UpdateProposalStatus(sctx.Ctx, req.ProposalID,
	// 	domain.ProposalStatusQCReturned, "QC returned to data entry: "+req.Comments, changedBy); err != nil {
	// 	log.Error(sctx.Ctx, "Failed to update proposal status to QC_RETURNED", "proposalID", req.ProposalID, "error", err)
	// 	return nil, err
	// }

	// Signal workflow
	if proposal.WorkflowID != nil {
		if err := h.temporalClient.SignalWorkflow(sctx.Ctx, *proposal.WorkflowID, "", workflows.SignalQRDecision, workflows.QRDecisionSignal{
			Decision:   "RETURNED",
			Comments:   req.Comments,
			ReviewerID: req.ReviewerID,
		}); err != nil {
			log.Error(sctx.Ctx, "Failed to signal workflow for QC return", "proposalID", req.ProposalID, "error", err)
		}
	}

	return &resp.QRReturnResponse{
		StatusCodeAndMessage: port.StatusCodeAndMessage{
			StatusCode: http.StatusOK,
			Message:    "QR return processed successfully",
		},
		Status: "success",
	}, nil
}

// ApproverApprove approves a proposal at the approver level
// [POL-API-020] Approver Approve
// State transition: APPROVAL_PENDING → APPROVED (BR-POL-015)
func (h *ApprovalHandler) ApproverApprove(sctx *serverRoute.Context, req ApproverApproveRequest) (*resp.ApproverApproveResponse, error) {

	proposal, err := h.proposalRepo.GetProposalByID(sctx.Ctx, req.ProposalID)
	if err != nil {
		return nil, err
	}

	// Validate proposal is in APPROVAL_PENDING status
	if proposal.Status != domain.ProposalStatusApprovalPending {
		return &resp.ApproverApproveResponse{
			StatusCodeAndMessage: port.StatusCodeAndMessage{
				StatusCode: http.StatusBadRequest,
				Message:    fmt.Sprintf("Proposal must be in APPROVAL_PENDING status for approval. Current: %s", proposal.Status),
			},
		}, nil
	}

	// Record approval decision
	if err := h.proposalRepo.RecordApproval(sctx.Ctx, req.ProposalID, "APPROVED", req.Comments, req.ApproverID); err != nil {
		return nil, err
	}

	// Persist status transition: APPROVAL_PENDING → APPROVED
	// changedBy, _ := strconv.ParseInt(req.ApproverID, 10, 64)
	// if err := h.proposalRepo.UpdateProposalStatus(sctx.Ctx, req.ProposalID,
	// 	domain.ProposalStatusApproved, "Approved: "+req.Comments, changedBy); err != nil {
	// 	log.Error(sctx.Ctx, "Failed to update proposal status to APPROVED", "proposalID", req.ProposalID, "error", err)
	// 	return nil, err
	// }

	// Signal workflow
	if proposal.WorkflowID != nil {
		if err := h.temporalClient.SignalWorkflow(sctx.Ctx, *proposal.WorkflowID, "", workflows.SignalApproverDecision, workflows.ApproverDecisionSignal{
			Decision:   "APPROVED",
			Comments:   req.Comments,
			ApproverID: req.ApproverID,
		}); err != nil {
			log.Error(sctx.Ctx, "Failed to signal workflow for approval", "proposalID", req.ProposalID, "error", err)
		}
	}

	return &resp.ApproverApproveResponse{
		StatusCodeAndMessage: port.StatusCodeAndMessage{
			StatusCode: http.StatusOK,
			Message:    "Approver approval processed successfully",
		},
		Status: "success",
	}, nil
}

// ApproverReject rejects a proposal at the approver level
// [POL-API-021] Approver Reject
// State transition: APPROVAL_PENDING → REJECTED (BR-POL-015)
func (h *ApprovalHandler) ApproverReject(sctx *serverRoute.Context, req ApproverRejectRequest) (*resp.ApproverRejectResponse, error) {

	proposal, err := h.proposalRepo.GetProposalByID(sctx.Ctx, req.ProposalID)
	if err != nil {
		return nil, err
	}

	// Validate proposal is in APPROVAL_PENDING status
	if proposal.Status != domain.ProposalStatusApprovalPending {
		return &resp.ApproverRejectResponse{
			StatusCodeAndMessage: port.StatusCodeAndMessage{
				StatusCode: http.StatusBadRequest,
				Message:    fmt.Sprintf("Proposal must be in APPROVAL_PENDING status for rejection. Current: %s", proposal.Status),
			},
		}, nil
	}

	// Record approval decision
	if err := h.proposalRepo.RecordApproval(sctx.Ctx, req.ProposalID, "REJECTED", req.Comments, req.ApproverID); err != nil {
		return nil, err
	}

	// // Persist status transition: APPROVAL_PENDING → REJECTED
	// changedBy, _ := strconv.ParseInt(req.ApproverID, 10, 64)
	// if err := h.proposalRepo.UpdateProposalStatus(sctx.Ctx, req.ProposalID,
	// 	domain.ProposalStatusRejected, "Rejected: "+req.Comments, changedBy); err != nil {
	// 	log.Error(sctx.Ctx, "Failed to update proposal status to REJECTED", "proposalID", req.ProposalID, "error", err)
	// 	return nil, err
	// }

	// Signal workflow
	if proposal.WorkflowID != nil {
		if err := h.temporalClient.SignalWorkflow(sctx.Ctx, *proposal.WorkflowID, "", workflows.SignalApproverDecision, workflows.ApproverDecisionSignal{
			Decision:   "REJECTED",
			Comments:   req.Comments,
			ApproverID: req.ApproverID,
		}); err != nil {
			log.Error(sctx.Ctx, "Failed to signal workflow for rejection", "proposalID", req.ProposalID, "error", err)
		}
	}

	return &resp.ApproverRejectResponse{
		StatusCodeAndMessage: port.StatusCodeAndMessage{
			StatusCode: http.StatusOK,
			Message:    "Approver rejection processed successfully",
		},
		Status: "success",
	}, nil
}
