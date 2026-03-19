package handler

// RequestLifecycleHandler — request listing, detail, withdrawal and CPC inbox (Step 4.3)
//
// Implements 5 endpoints:
//   - GET  /policies/{pn}/requests                        — list with filters
//   - GET  /policies/{pn}/requests/{request_id}           — detail
//   - PUT  /policies/{pn}/requests/{request_id}/withdraw  — withdraw [BR-PM-090]
//   - GET  /requests/pending                              — CPC inbox
//   - GET  /requests/pending/summary                      — CPC dashboard
//
// Withdraw flow [FR-PM-007, BR-PM-090]:
//  1. Lookup service request by request_id
//  2. Verify request belongs to the given policy_number (trust DB — no extra lookup)
//  3. WithdrawServiceRequest → marks status=WITHDRAWN (only if RECEIVED or ROUTED)
//  4. Signal plw-{policyNumber} with "withdrawal-request" so PLW reverts state
//  5. Return 200 WithdrawalResponse

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
	log "gitlab.cept.gov.in/it-2.0-common/n-api-log"
	serverHandler "gitlab.cept.gov.in/it-2.0-common/n-api-server/handler"
	serverRoute "gitlab.cept.gov.in/it-2.0-common/n-api-server/route"
	"go.temporal.io/sdk/client"

	"policy-management/core/domain"
	"policy-management/core/port"
	resp "policy-management/handler/response"
	repo "policy-management/repo/postgres"
)

// withdrawalSignalPayload is the payload sent to plw-{policyNumber} on withdrawal.
// The workflow uses TargetRequestID to find the PendingRequest and cancel the
// downstream child workflow + release the financial lock. [FR-PM-007, BR-PM-090]
// JSON tags MUST match workflows.WithdrawalRequestSignal exactly.
type withdrawalSignalPayload struct {
	RequestID        string `json:"request_id"`        // stable dedup key: "withdrawal-{requestID}"
	TargetRequestID  string `json:"target_request_id"` // PendingRequest.RequestID in PLW (UUID or BIGINT string)
	WithdrawalReason string `json:"withdrawal_reason"` // was "reason" — mismatch fixed
}

// ─────────────────────────────────────────────────────────────────────────────
// RequestLifecycleHandler
// ─────────────────────────────────────────────────────────────────────────────

// RequestLifecycleHandler handles request listing, detail, withdrawal, and CPC inbox.
// [FR-PM-006, FR-PM-007, FR-PM-008]
type RequestLifecycleHandler struct {
	*serverHandler.Base
	policyRepo *repo.PolicyRepository
	srRepo     *repo.ServiceRequestRepository
	tc         client.Client
}

// NewRequestLifecycleHandler constructs a RequestLifecycleHandler with dependencies.
func NewRequestLifecycleHandler(
	policyRepo *repo.PolicyRepository,
	srRepo *repo.ServiceRequestRepository,
	tc client.Client,
) *RequestLifecycleHandler {
	base := serverHandler.New("RequestLifecycle").SetPrefix("/v1").AddPrefix("")
	return &RequestLifecycleHandler{
		Base:       base,
		policyRepo: policyRepo,
		srRepo:     srRepo,
		tc:         tc,
	}
}

// Routes registers all 5 request lifecycle endpoints.
func (h *RequestLifecycleHandler) Routes() []serverRoute.Route {
	return []serverRoute.Route{
		// Per-policy request management
		serverRoute.GET("/policies/:policy_number/requests", h.ListRequests).
			Name("List Policy Requests"),
		serverRoute.GET("/policies/:policy_number/requests/:request_id", h.GetRequest).
			Name("Get Request Detail"),
		serverRoute.PUT("/policies/:policy_number/requests/:request_id/withdraw", h.WithdrawRequest).
			Name("Withdraw Request"),

		// CPC inbox (cross-policy)
		serverRoute.GET("/requests/pending", h.GetPendingRequests).
			Name("Get Pending Requests"),
		serverRoute.GET("/requests/pending/summary", h.GetPendingSummary).
			Name("Get Pending Summary"),
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Combined request types
// ─────────────────────────────────────────────────────────────────────────────

type listRequestsReq struct {
	PolicyNumberURI
	ListRequestsParams
}

type getRequestReq struct {
	PolicyNumberURI
	RequestIDURI
}

type withdrawReq struct {
	PolicyNumberURI
	RequestIDURI
	WithdrawRequestRequest
}

// ─────────────────────────────────────────────────────────────────────────────
// Endpoint handlers
// ─────────────────────────────────────────────────────────────────────────────

// ListRequests — GET /v1/policies/:policy_number/requests
// Returns a paginated list of service requests for the given policy.
// Supports filtering by status, request_type, source_channel, date range. [FR-PM-006]
func (h *RequestLifecycleHandler) ListRequests(sctx *serverRoute.Context, req listRequestsReq) (*resp.RequestListResponse, error) {
	ctx := sctx.Ctx
	policyNumber := req.PolicyNumber

	// Lookup policy to get policyID (needed by ListServiceRequestsByPolicy index).
	policy, err := h.policyRepo.GetPolicyByNumber(ctx, policyNumber)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, newHTTPErr(http.StatusNotFound,
				fmt.Sprintf("policy %s not found", policyNumber), err)
		}
		log.Error(ctx, "GetPolicyByNumber %s: %v", policyNumber, err)
		return nil, err
	}

	// Build filters from query params.
	f := repo.ListRequestsFilter{
		Skip:    req.Skip,
		Limit:   req.Limit,
		OrderBy: req.SortBy,
		SortType: func() string {
			// Default DESC — newest first.
			return "DESC"
		}(),
	}
	if req.RequestType != "" {
		f.RequestType = &req.RequestType
	}
	if req.Status != "" {
		f.Status = &req.Status
	}
	if f.Limit == 0 {
		f.Limit = 10
	}

	requests, total, err := h.srRepo.ListServiceRequestsByPolicy(ctx, policy.PolicyID, f)
	if err != nil {
		log.Error(ctx, "ListServiceRequestsByPolicy policyID=%d: %v", policy.PolicyID, err)
		return nil, err
	}

	return &resp.RequestListResponse{
		StatusCodeAndMessage: port.ListSuccess,
		MetaDataResponse:     port.NewMetaDataResponse(f.Skip, f.Limit, uint64(total)),
		Data:                 resp.NewRequestSummaryDataList(requests),
	}, nil
}

// GetRequest — GET /v1/policies/:policy_number/requests/:request_id
// Returns the full detail of a single service request. [FR-PM-006]
func (h *RequestLifecycleHandler) GetRequest(sctx *serverRoute.Context, req getRequestReq) (*resp.RequestDetailResponse, error) {
	ctx := sctx.Ctx

	sr, err := h.srRepo.GetServiceRequest(ctx, req.RequestID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, newHTTPErr(http.StatusNotFound,
				fmt.Sprintf("request %d not found", req.RequestID), err)
		}
		log.Error(ctx, "GetServiceRequest requestID=%d: %v", req.RequestID, err)
		return nil, err
	}

	// Verify request belongs to the policy in the URL. [Security check]
	if sr.PolicyNumber != req.PolicyNumber {
		return nil, newHTTPErr(http.StatusNotFound,
			fmt.Sprintf("request %d not found for policy %s", req.RequestID, req.PolicyNumber), nil)
	}

	return &resp.RequestDetailResponse{
		StatusCodeAndMessage: port.FetchSuccess,
		Data:                 resp.NewRequestDetailData(*sr),
	}, nil
}

// WithdrawRequest — PUT /v1/policies/:policy_number/requests/:request_id/withdraw
// Withdraws a pending service request. Only RECEIVED or ROUTED requests can be withdrawn.
// Signals plw-{policyNumber} to revert policy state and cancel downstream child workflow.
// [FR-PM-007, BR-PM-090]
func (h *RequestLifecycleHandler) WithdrawRequest(sctx *serverRoute.Context, req withdrawReq) (*resp.WithdrawalResponse, error) {
	ctx := sctx.Ctx
	policyNumber := req.PolicyNumber
	requestID := req.RequestID

	// Fetch the service request to validate it belongs to this policy.
	sr, err := h.srRepo.GetServiceRequest(ctx, requestID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, newHTTPErr(http.StatusNotFound,
				fmt.Sprintf("request %d not found", requestID), err)
		}
		log.Error(ctx, "GetServiceRequest requestID=%d: %v", requestID, err)
		return nil, err
	}

	// Verify request belongs to the policy in the URL.
	if sr.PolicyNumber != policyNumber {
		return nil, newHTTPErr(http.StatusNotFound,
			fmt.Sprintf("request %d not found for policy %s", requestID, policyNumber), nil)
	}

	// Verify request is in a withdrawable status. [BR-PM-090]
	if sr.Status != domain.RequestStatusReceived && sr.Status != domain.RequestStatusRouted {
		return nil, newHTTPErr(http.StatusConflict,
			fmt.Sprintf("request %d has status %s and cannot be withdrawn; only RECEIVED or ROUTED requests can be withdrawn",
				requestID, sr.Status), nil)
	}

	// Withdraw in DB — marks status=WITHDRAWN (WHERE status IN RECEIVED,ROUTED).
	if wErr := h.srRepo.WithdrawServiceRequest(ctx, requestID, req.Reason, &sr.SubmittedAt); wErr != nil {
		log.Error(ctx, "WithdrawServiceRequest requestID=%d: %v", requestID, wErr)
		return nil, wErr
	}

	// Signal plw-{policyNumber} to revert state and cancel downstream child. [BR-PM-090]
	// TargetRequestID must match PendingRequest.RequestID stored in PLW state.
	// PLW stores the UUID idempotency key as PendingRequest.RequestID (Constraint 1 / Review-Fix-11).
	// Fall back to BIGINT-as-string for legacy requests submitted without idempotency key.
	wfID := policyWorkflowID(policyNumber)
	targetRequestID := fmt.Sprintf("%d", requestID)
	if sr.IdempotencyKey != nil && *sr.IdempotencyKey != "" {
		targetRequestID = *sr.IdempotencyKey
	}
	signalPayload := withdrawalSignalPayload{
		RequestID:        fmt.Sprintf("withdrawal-%d", requestID), // deterministic dedup key per target
		TargetRequestID:  targetRequestID,
		WithdrawalReason: req.Reason,
	}
	if sigErr := h.tc.SignalWorkflow(ctx, wfID, "", "withdrawal-request", signalPayload); sigErr != nil {
		// Non-fatal — DB update succeeded. Workflow may be completed or not exist.
		// Log warning and return success; PLW will reconcile via DB on next refresh.
		log.Warn(ctx, "SignalWorkflow withdrawal-request wfID=%s requestID=%d: %v (non-fatal)", wfID, requestID, sigErr)
	}

	now := time.Now().UTC()
	return &resp.WithdrawalResponse{
		StatusCodeAndMessage: port.WithdrawnSuccess,
		Data: resp.WithdrawalData{
			RequestID:   requestID,
			Status:      domain.RequestStatusWithdrawn,
			WithdrawnAt: now.Format("2006-01-02T15:04:05Z07:00"),
		},
	}, nil
}

// GetPendingRequests — GET /v1/requests/pending
// CPC inbox: lists all RECEIVED, ROUTED, IN_PROGRESS requests across all policies.
// Supports filtering by request_type. [FR-PM-008]
func (h *RequestLifecycleHandler) GetPendingRequests(sctx *serverRoute.Context, req ListPendingRequestsParams) (*resp.PendingRequestsResponse, error) {
	ctx := sctx.Ctx

	f := repo.PendingRequestsFilter{
		Skip:  req.Skip,
		Limit: req.Limit,
	}
	if f.Limit == 0 {
		f.Limit = 20
	}
	if req.RequestType != "" {
		f.RequestType = &req.RequestType
	}

	requests, total, err := h.srRepo.GetPendingRequests(ctx, f)
	if err != nil {
		log.Error(ctx, "GetPendingRequests: %v", err)
		return nil, err
	}

	page := 1
	if f.Skip > 0 && f.Limit > 0 {
		page = int(f.Skip/f.Limit) + 1
	}

	return &resp.PendingRequestsResponse{
		StatusCodeAndMessage: port.ListSuccess,
		Data: resp.PendingRequestsData{
			TotalCount: int(total),
			Page:       page,
			PageSize:   int(f.Limit),
			Requests:   resp.NewRequestSummaryDataList(requests),
		},
	}, nil
}

// GetPendingSummary — GET /v1/requests/pending/summary
// CPC dashboard: aggregated pending request counts by type and status. [FR-PM-008]
func (h *RequestLifecycleHandler) GetPendingSummary(sctx *serverRoute.Context, req struct{}) (*resp.DashboardSummaryResponse, error) {
	ctx := sctx.Ctx

	summary, err := h.srRepo.GetDashboardSummary(ctx)
	if err != nil {
		log.Error(ctx, "GetDashboardSummary: %v", err)
		return nil, err
	}

	// Convert domain.DashboardSummary → response DTO map types.
	summaryMap := make(map[string]resp.RequestStatusCounts)
	for reqType, statusCounts := range summary.Summary {
		summaryMap[reqType] = resp.RequestStatusCounts(statusCounts)
	}

	return &resp.DashboardSummaryResponse{
		StatusCodeAndMessage: port.FetchSuccess,
		Data: resp.DashboardSummaryData{
			Summary:               summaryMap,
			TotalPending:          summary.TotalPending,
			OldestRequestAgeHours: summary.OldestRequestAgeHrs,
		},
	}, nil
}
