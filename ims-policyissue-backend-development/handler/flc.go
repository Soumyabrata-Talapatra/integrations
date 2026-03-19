package handler

import (
	"fmt"
	"net/http"

	"policy-issue-service/core/port"
	resp "policy-issue-service/handler/response"
	repo "policy-issue-service/repo/postgres"

	log "gitlab.cept.gov.in/it-2.0-common/api-log"
	serverHandler "gitlab.cept.gov.in/it-2.0-common/n-api-server/handler"
	serverRoute "gitlab.cept.gov.in/it-2.0-common/n-api-server/route"

	"go.temporal.io/sdk/client"
)

// FLCManagementHandler handles Free Look Cancellation related HTTP endpoints
type FLCManagementHandler struct {
	*serverHandler.Base

	temporalClient client.Client
	proposalRepo   *repo.ProposalRepository
}

// NewFLCManagementHandler creates a new FLCManagementHandler instance
func NewFLCManagementHandler(temporalClient client.Client, proposalRepo *repo.ProposalRepository) *FLCManagementHandler {
	base := serverHandler.New("FLCManagement").SetPrefix("/v1").AddPrefix("")
	return &FLCManagementHandler{
		Base:           base,
		temporalClient: temporalClient,
		proposalRepo:   proposalRepo,
	}
}

// Routes returns the routes for the FLCManagementHandler
func (h *FLCManagementHandler) Routes() []serverRoute.Route {
	return []serverRoute.Route{
		serverRoute.POST("/proposals/:proposal_id/initiate-flc", h.InitiateFLC).Name("Initiate FLC"),
		serverRoute.POST("/proposals/:proposal_id/approve-flc", h.ApproveFLC).Name("Approve FLC"),
		serverRoute.POST("/proposals/:proposal_id/reject-flc", h.RejectFLC).Name("Reject FLC"),
		serverRoute.GET("/proposals/:proposal_id/flc-status", h.GetFLCStatus).Name("Get FLC Status"),
		serverRoute.GET("/proposals/flc-queue", h.GetFLCQueue).Name("Get FLC Queue"),
	}
}

// InitiateFLC initiates a Free Look Cancellation request
// [POL-API-022] Initiate FLC
func (h *FLCManagementHandler) InitiateFLC(sctx *serverRoute.Context, req FLCInitiateRequest) (*resp.FLCInitiateResponse, error) {
	 

	if err := h.proposalRepo.RecordFLCRequest(sctx.Ctx, req.ProposalID, req.RequestReason, req.Comments); err != nil {
		log.Error(sctx.Ctx, "Failed to record FLC request", "proposalID", req.ProposalID, "error", err)
		return nil, err
	}

	return &resp.FLCInitiateResponse{
		StatusCodeAndMessage: port.StatusCodeAndMessage{
			StatusCode: http.StatusOK,
			Message:    "Free Look Cancellation initiated successfully",
		},
		FLCRequestID: fmt.Sprintf("FLC_%d", req.ProposalID),
		Status:       "initiated",
	}, nil
}

// ApproveFLC approves a Free Look Cancellation request
// [POL-API-023] Approve FLC
func (h *FLCManagementHandler) ApproveFLC(sctx *serverRoute.Context, req FLCApproveRequest) (*resp.FLCApproveResponse, error) {
	 

	// Update status to FLC_CANCELLED
	if err := h.proposalRepo.UpdateProposalStatus(sctx.Ctx, req.ProposalID, "FLC_CANCELLED", req.Comments, 0); err != nil {
		return nil, err
	}

	return &resp.FLCApproveResponse{
		StatusCodeAndMessage: port.StatusCodeAndMessage{
			StatusCode: http.StatusOK,
			Message:    "Free Look Cancellation approved successfully",
		},
		Status: "approved",
	}, nil
}

// RejectFLC rejects a Free Look Cancellation request
// [POL-API-024] Reject FLC
func (h *FLCManagementHandler) RejectFLC(sctx *serverRoute.Context, req FLCRejectRequest) (*resp.FLCRejectResponse, error) {
	 

	// Revert or update status
	if err := h.proposalRepo.UpdateProposalStatus(sctx.Ctx, req.ProposalID, "ACTIVE", req.RejectReason, 0); err != nil {
		return nil, err
	}

	return &resp.FLCRejectResponse{
		StatusCodeAndMessage: port.StatusCodeAndMessage{
			StatusCode: http.StatusOK,
			Message:    "Free Look Cancellation rejected successfully",
		},
		Status: "rejected",
	}, nil
}

// GetFLCStatus gets the status of a Free Look Cancellation request
// [POL-API-025] Get FLC Status
func (h *FLCManagementHandler) GetFLCStatus(sctx *serverRoute.Context, req FLCStatusUri) (*resp.GetFLCStatusResponse, error) {
	proposal, err := h.proposalRepo.GetProposalByID(sctx.Ctx, req.ProposalID)
	if err != nil {
		return nil, err
	}

	return &resp.GetFLCStatusResponse{
		StatusCodeAndMessage: port.StatusCodeAndMessage{
			StatusCode: http.StatusOK,
			Message:    "FLC status retrieved successfully",
		},
		Data: map[string]interface{}{
			"status": proposal.Status,
		},
	}, nil
}

// GetFLCQueue gets the Free Look Cancellation queue
// [POL-API-026] Get FLC Queue
func (h *FLCManagementHandler) GetFLCQueue(sctx *serverRoute.Context, req FLCQueueRequest) (*resp.FLCQueueResponse, error) {
	 

	req.MetadataRequest.SetDefaults()

	// For now returning empty list as we need to implement specific FLC queue query in repo
	return &resp.FLCQueueResponse{
		StatusCodeAndMessage: port.StatusCodeAndMessage{
			StatusCode: http.StatusOK,
			Message:    "FLC queue retrieved successfully",
		},
		MetaDataResponse: port.NewMetaDataResponse(0, int(req.Skip), int(req.Limit)),
		Data:             []interface{}{},
	}, nil
}
