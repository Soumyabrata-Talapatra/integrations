package handler

// PolicyQueryHandler — real-time policy state query endpoints (Step 4.4)
//
// Implements 6 endpoints:
//   - GET /policies/{pn}/status          — Two-Tier: QueryWorkflow("getFullState") → snapshot
//   - GET /policies/{pn}/summary         — Two-Tier: QueryWorkflow("getFullState") → snapshot
//   - GET /policies/{pn}/state-gate/{t}  — Two-Tier: QueryWorkflow("getStateGate") → snapshot
//   - GET /policies/{pn}/history         — DB only (policy_status_history table)
//   - GET /policies/batch-status         — Parallel QueryWorkflow per policy → snapshot
//   - GET /policies/dashboard/metrics    — DB only (mv_policy_dashboard MV)
//
// Two-Tier Query Pattern (Constraint 8, §9.5.1, AD-011):
//   Step 1: client.QueryWorkflow("plw-{policyNumber}", queryName, args...)
//           → Success: return workflow in-memory state (most current)
//           → Workflow not found or query error: fall through to Step 2
//   Step 2: Read from terminal_state_snapshot via policyRepo.GetTerminalSnapshot()
//           → Not found in snapshot: return 404 (policy does not exist in PM)

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sync"

	pgx "github.com/jackc/pgx/v5"
	log "gitlab.cept.gov.in/it-2.0-common/n-api-log"
	serverHandler "gitlab.cept.gov.in/it-2.0-common/n-api-server/handler"
	serverRoute "gitlab.cept.gov.in/it-2.0-common/n-api-server/route"
	"go.temporal.io/sdk/client"

	"policy-management/core/domain"
	"policy-management/core/port"
	resp "policy-management/handler/response"
	repo "policy-management/repo/postgres"
)

// ─────────────────────────────────────────────────────────────────────────────
// Workflow query handler names (Phase 5 registers these on PolicyLifecycleWorkflow)
// ─────────────────────────────────────────────────────────────────────────────

const (
	queryGetFullState = "getFullState"
	queryGetStateGate = "getStateGate"
)

// ─────────────────────────────────────────────────────────────────────────────
// Workflow query result types
//
// These structs define the JSON shape that Phase 5 workflow query handlers MUST
// return. PolicyLifecycleWorkflow will serialize its internal state into this
// exact shape when "getFullState" / "getStateGate" is called.
// ─────────────────────────────────────────────────────────────────────────────

// wfEncumbrances mirrors the Encumbrances shape from the workflow's PolicyLifecycleState.
// Phase 5 PolicyLifecycleWorkflow MUST return this exact shape from "getFullState" query.
// loan_id and assignee_id are BIGINT IDs from external services (loan-svc, nfs-svc).
type wfEncumbrances struct {
	HasActiveLoan   bool    `json:"has_active_loan"`
	LoanID          *int64  `json:"loan_id,omitempty"`         // BIGINT from loan service; nil if no active loan
	LoanOutstanding float64 `json:"loan_outstanding,omitempty"`
	AssignmentType  string  `json:"assignment_type"`           // NONE | ABSOLUTE | CONDITIONAL
	AssigneeID      *int64  `json:"assignee_id,omitempty"`     // BIGINT from NFS service; nil if not assigned
	AMLHold         bool    `json:"aml_hold"`
	DisputeFlag     bool    `json:"dispute_flag"`
}

// wfPolicyState is the result shape of the "getFullState" workflow query.
// Phase 5 PolicyLifecycleWorkflow maps its PolicyLifecycleState + PolicyMetadata
// into this struct for the REST layer to consume.
type wfPolicyState struct {
	// Identity & status
	PolicyID       int64          `json:"policy_id"`
	PolicyNumber   string         `json:"policy_number"`
	CurrentStatus  string         `json:"current_status"`
	PreviousStatus *string        `json:"previous_status,omitempty"`
	Encumbrances   wfEncumbrances `json:"encumbrances"`
	DisplayStatus  string         `json:"display_status"`
	Version        int64          `json:"version"`
	EffectiveFrom  string         `json:"effective_from"` // RFC3339
	UpdatedAt      string         `json:"updated_at"`     // RFC3339

	// Metadata (for summary endpoint)
	CustomerID         int64    `json:"customer_id"`
	ProductCode        string   `json:"product_code"`
	ProductType        string   `json:"product_type"`
	SumAssured         float64  `json:"sum_assured"`
	CurrentPremium     float64  `json:"current_premium"`
	PremiumMode        string   `json:"premium_mode"`
	BillingMethod      string   `json:"billing_method"`
	IssueDate          string   `json:"issue_date"`   // "2006-01-02"
	MaturityDate       *string  `json:"maturity_date,omitempty"`
	PaidToDate         string   `json:"paid_to_date"` // "2006-01-02"
	NextPremiumDueDate *string  `json:"next_premium_due_date,omitempty"`
	PaidUpValue        *float64 `json:"paid_up_value,omitempty"`
	AgentID            *int64   `json:"agent_id,omitempty"`
}

// wfStateGateResult is the result shape of the "getStateGate" workflow query.
// Phase 5 returns this after calling its internal isStateEligible().
type wfStateGateResult struct {
	CurrentStatus   string         `json:"current_status"`
	IsEligible      bool           `json:"is_eligible"`
	RejectionReason *string        `json:"rejection_reason,omitempty"`
	HasActiveLock   bool           `json:"has_active_lock"`
	Encumbrances    wfEncumbrances `json:"encumbrances"`
}

// ─────────────────────────────────────────────────────────────────────────────
// PolicyQueryHandler
// ─────────────────────────────────────────────────────────────────────────────

// PolicyQueryHandler handles all real-time policy state query endpoints.
// Depends on both tc (Temporal client — Tier 1) and policyRepo (DB — Tier 2 fallback).
// [FR-PM-001, AD-011, Constraint 8]
type PolicyQueryHandler struct {
	*serverHandler.Base
	policyRepo *repo.PolicyRepository
	tc         client.Client
}

// NewPolicyQueryHandler constructs a PolicyQueryHandler with required dependencies.
func NewPolicyQueryHandler(
	policyRepo *repo.PolicyRepository,
	tc client.Client,
) *PolicyQueryHandler {
	base := serverHandler.New("PolicyQuery").SetPrefix("/v1").AddPrefix("")
	return &PolicyQueryHandler{
		Base:       base,
		policyRepo: policyRepo,
		tc:         tc,
	}
}

// Routes registers all 6 policy query endpoints.
func (h *PolicyQueryHandler) Routes() []serverRoute.Route {
	return []serverRoute.Route{
		serverRoute.GET("/policies/:policy_number/status", h.GetPolicyStatus).
			Name("Get Policy Status"),
		serverRoute.GET("/policies/:policy_number/summary", h.GetPolicySummary).
			Name("Get Policy Summary"),
		serverRoute.GET("/policies/:policy_number/state-gate/:request_type", h.GetStateGate).
			Name("Get State Gate"),
		serverRoute.GET("/policies/:policy_number/history", h.GetPolicyHistory).
			Name("Get Policy History"),
		serverRoute.GET("/policies/batch-status", h.GetBatchStatus).
			Name("Get Batch Status"),
		serverRoute.GET("/policies/dashboard/metrics", h.GetDashboardMetrics).
			Name("Get Dashboard Metrics"),
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Combined request types
// ─────────────────────────────────────────────────────────────────────────────

type policyStatusReq struct {
	PolicyNumberURI
}

type policySummaryReq struct {
	PolicyNumberURI
}

type stateGateReq struct {
	PolicyNumberURI
	StateGateTypeURI
}

type policyHistoryReq struct {
	PolicyNumberURI
	port.MetadataRequest
}

// ─────────────────────────────────────────────────────────────────────────────
// Status (Two-Tier)
// ─────────────────────────────────────────────────────────────────────────────

// GetPolicyStatus — GET /v1/policies/:policy_number/status
// Returns current lifecycle status, previous status, encumbrances, display status.
// Two-Tier: QueryWorkflow("getFullState") → terminal_state_snapshot fallback. [AD-011]
func (h *PolicyQueryHandler) GetPolicyStatus(sctx *serverRoute.Context, req policyStatusReq) (*resp.PolicyStatusResponse, error) {
	state, err := h.twoTierFullState(sctx.Ctx, req.PolicyNumber)
	if err != nil {
		return nil, err
	}

	return &resp.PolicyStatusResponse{
		StatusCodeAndMessage: port.FetchSuccess,
		Data: resp.PolicyStatusData{
			PolicyID:        state.PolicyID,
			PolicyNumber:    state.PolicyNumber,
			LifecycleStatus: state.CurrentStatus,
			PreviousStatus:  state.PreviousStatus,
			Encumbrances: resp.Encumbrances{
				HasActiveLoan:   state.Encumbrances.HasActiveLoan,
				LoanID:          state.Encumbrances.LoanID,
				LoanOutstanding: state.Encumbrances.LoanOutstanding,
				AssignmentType:  state.Encumbrances.AssignmentType,
				AssigneeID:      state.Encumbrances.AssigneeID,
				AMLHold:         state.Encumbrances.AMLHold,
				DisputeFlag:     state.Encumbrances.DisputeFlag,
			},
			DisplayStatus: state.DisplayStatus,
			EffectiveFrom: state.EffectiveFrom,
			Version:       state.Version,
			UpdatedAt:     state.UpdatedAt,
		},
	}, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Summary (Two-Tier)
// ─────────────────────────────────────────────────────────────────────────────

// GetPolicySummary — GET /v1/policies/:policy_number/summary
// Returns full policy metadata including product, premium, dates, encumbrances.
// Two-Tier: QueryWorkflow("getFullState") → terminal_state_snapshot fallback. [AD-011]
func (h *PolicyQueryHandler) GetPolicySummary(sctx *serverRoute.Context, req policySummaryReq) (*resp.PolicySummaryResponse, error) {
	state, err := h.twoTierFullState(sctx.Ctx, req.PolicyNumber)
	if err != nil {
		return nil, err
	}

	return &resp.PolicySummaryResponse{
		StatusCodeAndMessage: port.FetchSuccess,
		Data: resp.PolicySummaryData{
			PolicyID:           state.PolicyID,
			PolicyNumber:       state.PolicyNumber,
			CustomerID:         state.CustomerID,
			ProductCode:        state.ProductCode,
			ProductType:        state.ProductType,
			LifecycleStatus:    state.CurrentStatus,
			DisplayStatus:      state.DisplayStatus,
			SumAssured:         state.SumAssured,
			CurrentPremium:     state.CurrentPremium,
			PremiumMode:        state.PremiumMode,
			BillingMethod:      state.BillingMethod,
			IssueDate:          state.IssueDate,
			MaturityDate:       state.MaturityDate,
			PaidToDate:         state.PaidToDate,
			NextPremiumDueDate: state.NextPremiumDueDate,
			Encumbrances: resp.Encumbrances{
				HasActiveLoan:   state.Encumbrances.HasActiveLoan,
				LoanID:          state.Encumbrances.LoanID,
				LoanOutstanding: state.Encumbrances.LoanOutstanding,
				AssignmentType:  state.Encumbrances.AssignmentType,
				AssigneeID:      state.Encumbrances.AssigneeID,
				AMLHold:         state.Encumbrances.AMLHold,
				DisputeFlag:     state.Encumbrances.DisputeFlag,
			},
			PaidUpValue: state.PaidUpValue,
			AgentID:     state.AgentID,
			Version:     state.Version,
		},
	}, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// State Gate (Two-Tier)
// ─────────────────────────────────────────────────────────────────────────────

// GetStateGate — GET /v1/policies/:policy_number/state-gate/:request_type
// Checks eligibility of a specific request type against current policy state.
// Tier 1: QueryWorkflow("getStateGate", requestType) → in-memory isStateEligible()
// Tier 2: terminal_state_snapshot + static checkStateGate(). [BR-PM-011..023]
func (h *PolicyQueryHandler) GetStateGate(sctx *serverRoute.Context, req stateGateReq) (*resp.StateGateResponse, error) {
	ctx := sctx.Ctx
	policyNumber := req.PolicyNumber
	requestType := req.RequestType
	wfID := "plw-" + policyNumber

	// Tier 1: QueryWorkflow("getStateGate", requestType).
	qResp, queryErr := h.tc.QueryWorkflow(ctx, wfID, "", queryGetStateGate, requestType)
	if queryErr == nil {
		var sgResult wfStateGateResult
		if decodeErr := qResp.Get(&sgResult); decodeErr == nil {
			return h.buildStateGateResp(policyNumber, requestType, sgResult), nil
		}
		log.Warn(ctx, "getStateGate decode %s: falling through to Tier 2", policyNumber)
	}

	// Tier 2: Check active policy table.
	policy, err := h.policyRepo.GetPolicyByNumber(ctx, policyNumber)
	if err == nil {
		// Active policy found — run static state gate check.
		sgErr := checkStateGate(policy, requestType)
		passed := sgErr == nil
		var rejReason *string
		if sgErr != nil {
			s := sgErr.Error()
			rejReason = &s
		}

		return &resp.StateGateResponse{
			StatusCodeAndMessage: port.FetchSuccess,
			Data: resp.StateGateData{
				PolicyNumber:    policyNumber,
				RequestType:     requestType,
				StateGatePassed: passed,
				CurrentStatus:   policy.CurrentStatus,
				AllowedStatuses: allowedStatusesForType(requestType),
				Encumbrances: resp.Encumbrances{
					HasActiveLoan:   policy.HasActiveLoan,
					LoanOutstanding: policy.LoanOutstanding,
					AssignmentType:  policy.AssignmentType,
					AMLHold:         policy.AMLHold,
					DisputeFlag:     policy.DisputeFlag,
				},
				RejectionReason: rejReason,
			},
		}, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		log.Error(ctx, "GetPolicyByNumber %s: %v", policyNumber, err)
		return nil, err
	}

	// Tier 3: terminal_state_snapshot fallback.
	snap, snapErr := h.policyRepo.GetTerminalSnapshot(ctx, policyNumber)
	if snapErr != nil {
		log.Error(ctx, "GetTerminalSnapshot %s: %v", policyNumber, snapErr)
		return nil, snapErr
	}
	if snap == nil {
		return nil, newHTTPErr(http.StatusNotFound,
			fmt.Sprintf("policy %s not found", policyNumber), nil)
	}

	// Parse snapshot to get current status and encumbrances.
	var snapState wfPolicyState
	if err := json.Unmarshal(snap.FinalSnapshot, &snapState); err != nil {
		snapState.CurrentStatus = snap.FinalStatus
		snapState.PolicyNumber = snap.PolicyNumber
	}

	// Run static state gate check — terminal policies always fail. [TerminalStatuses]
	sgErr := checkStateGate(&domain.Policy{
		PolicyNumber:  policyNumber,
		CurrentStatus: snapState.CurrentStatus,
	}, requestType)
	passed := sgErr == nil
	var rejReason *string
	if sgErr != nil {
		s := sgErr.Error()
		rejReason = &s
	}

	return &resp.StateGateResponse{
		StatusCodeAndMessage: port.FetchSuccess,
		Data: resp.StateGateData{
			PolicyNumber:    policyNumber,
			RequestType:     requestType,
			StateGatePassed: passed,
			CurrentStatus:   snapState.CurrentStatus,
			AllowedStatuses: allowedStatusesForType(requestType),
			Encumbrances: resp.Encumbrances{
				HasActiveLoan:   snapState.Encumbrances.HasActiveLoan,
				LoanID:          snapState.Encumbrances.LoanID,
				LoanOutstanding: snapState.Encumbrances.LoanOutstanding,
				AssignmentType:  snapState.Encumbrances.AssignmentType,
				AssigneeID:      snapState.Encumbrances.AssigneeID,
				AMLHold:         snapState.Encumbrances.AMLHold,
				DisputeFlag:     snapState.Encumbrances.DisputeFlag,
			},
			RejectionReason: rejReason,
		},
	}, nil
}

func (h *PolicyQueryHandler) buildStateGateResp(policyNumber, requestType string, r wfStateGateResult) *resp.StateGateResponse {
	return &resp.StateGateResponse{
		StatusCodeAndMessage: port.FetchSuccess,
		Data: resp.StateGateData{
			PolicyNumber:    policyNumber,
			RequestType:     requestType,
			StateGatePassed: r.IsEligible,
			CurrentStatus:   r.CurrentStatus,
			AllowedStatuses: allowedStatusesForType(requestType),
			Encumbrances: resp.Encumbrances{
				HasActiveLoan:   r.Encumbrances.HasActiveLoan,
				LoanID:          r.Encumbrances.LoanID,
				LoanOutstanding: r.Encumbrances.LoanOutstanding,
				AssignmentType:  r.Encumbrances.AssignmentType,
				AssigneeID:      r.Encumbrances.AssigneeID,
				AMLHold:         r.Encumbrances.AMLHold,
				DisputeFlag:     r.Encumbrances.DisputeFlag,
			},
			HasPendingFinancialLock: r.HasActiveLock,
			RejectionReason:         r.RejectionReason,
		},
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// History (DB-only)
// ─────────────────────────────────────────────────────────────────────────────

// GetPolicyHistory — GET /v1/policies/:policy_number/history
// DB-only: always reads policy_status_history table. Workflow is not queried.
// History is not stored in workflow in-memory state — always in DB. [FR-PM-002]
func (h *PolicyQueryHandler) GetPolicyHistory(sctx *serverRoute.Context, req policyHistoryReq) (*resp.PolicyHistoryResponse, error) {
	ctx := sctx.Ctx
	policyNumber := req.PolicyNumber
	limit := req.Limit
	if limit == 0 {
		limit = 10
	}

	// Lookup policy to get policyID (active policy path).
	policy, err := h.policyRepo.GetPolicyByNumber(ctx, policyNumber)
	if err != nil {
		if !errors.Is(err, pgx.ErrNoRows) {
			log.Error(ctx, "GetPolicyByNumber %s: %v", policyNumber, err)
			return nil, err
		}
		// Policy not in active table → check terminal_state_snapshot for policyID.
		snap, snapErr := h.policyRepo.GetTerminalSnapshot(ctx, policyNumber)
		if snapErr != nil || snap == nil {
			return nil, newHTTPErr(http.StatusNotFound,
				fmt.Sprintf("policy %s not found", policyNumber), err)
		}
		return h.fetchAndBuildHistory(ctx, policyNumber, snap.PolicyID, req.Skip, limit)
	}

	return h.fetchAndBuildHistory(ctx, policyNumber, policy.PolicyID, req.Skip, limit)
}

func (h *PolicyQueryHandler) fetchAndBuildHistory(
	ctx context.Context,
	policyNumber string,
	policyID int64,
	skip, limit uint64,
) (*resp.PolicyHistoryResponse, error) {
	history, total, err := h.policyRepo.GetPolicyStatusHistory(ctx, policyID, skip, limit)
	if err != nil {
		log.Error(ctx, "GetPolicyStatusHistory policyID=%d: %v", policyID, err)
		return nil, err
	}

	transitions := make([]resp.PolicyTransition, 0, len(history))
	for _, h := range history {
		transitions = append(transitions, resp.NewPolicyTransition(h))
	}

	return &resp.PolicyHistoryResponse{
		StatusCodeAndMessage: port.ListSuccess,
		MetaDataResponse:     port.NewMetaDataResponse(skip, limit, uint64(total)),
		Data: resp.PolicyHistoryData{
			PolicyNumber:     policyNumber,
			TotalTransitions: int(total),
			Transitions:      transitions,
		},
	}, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Batch Status (Parallel Two-Tier)
// ─────────────────────────────────────────────────────────────────────────────

// GetBatchStatus — GET /v1/policies/batch-status
// Returns lightweight status for up to 50 policies.
// Runs parallel Two-Tier queries per policy, collects results. [FR-PM-001, AD-011]
func (h *PolicyQueryHandler) GetBatchStatus(sctx *serverRoute.Context, req BatchStatusRequest) (*resp.BatchStatusResponse, error) {
	ctx := sctx.Ctx
	policyNumbers := req.PolicyNumbers

	if len(policyNumbers) == 0 {
		return &resp.BatchStatusResponse{
			StatusCodeAndMessage: port.FetchSuccess,
			Data:                 resp.BatchStatusData{Policies: []resp.BatchPolicyStatus{}},
		}, nil
	}

	type batchResult struct {
		idx    int
		status resp.BatchPolicyStatus
		err    error
	}

	results := make([]batchResult, len(policyNumbers))
	var wg sync.WaitGroup

	for i, pn := range policyNumbers {
		wg.Add(1)
		go func(idx int, policyNumber string) {
			defer wg.Done()
			batchStatus, bErr := h.singleBatchStatus(ctx, policyNumber)
			results[idx] = batchResult{idx: idx, status: batchStatus, err: bErr}
		}(i, pn)
	}
	wg.Wait()

	policies := make([]resp.BatchPolicyStatus, 0, len(policyNumbers))
	for _, r := range results {
		if r.err != nil {
			log.Warn(ctx, "batch-status %s: %v (skipped in response)", policyNumbers[r.idx], r.err)
			continue
		}
		policies = append(policies, r.status)
	}

	return &resp.BatchStatusResponse{
		StatusCodeAndMessage: port.FetchSuccess,
		Data:                 resp.BatchStatusData{Policies: policies},
	}, nil
}

// singleBatchStatus returns the lightweight status for one policy.
// Uses Two-Tier pattern: QueryWorkflow first, terminal snapshot fallback.
func (h *PolicyQueryHandler) singleBatchStatus(ctx context.Context, policyNumber string) (resp.BatchPolicyStatus, error) {
	wfID := "plw-" + policyNumber

	// Tier 1: Try QueryWorkflow.
	qResp, err := h.tc.QueryWorkflow(ctx, wfID, "", queryGetFullState)
	if err == nil {
		var state wfPolicyState
		if decodeErr := qResp.Get(&state); decodeErr == nil {
			return resp.BatchPolicyStatus{
				PolicyNumber:    policyNumber,
				LifecycleStatus: state.CurrentStatus,
				DisplayStatus:   state.DisplayStatus,
			}, nil
		}
	}

	// Tier 2: Check active policy table.
	policy, err := h.policyRepo.GetPolicyByNumber(ctx, policyNumber)
	if err == nil {
		// Active policy found.
		return resp.BatchPolicyStatus{
			PolicyNumber:    policyNumber,
			LifecycleStatus: policy.CurrentStatus,
			DisplayStatus:   policy.DisplayStatus,
		}, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return resp.BatchPolicyStatus{}, fmt.Errorf("GetPolicyByNumber %s: %w", policyNumber, err)
	}

	// Tier 3: terminal_state_snapshot fallback.
	snap, snapErr := h.policyRepo.GetTerminalSnapshot(ctx, policyNumber)
	if snapErr != nil {
		return resp.BatchPolicyStatus{}, fmt.Errorf("GetTerminalSnapshot %s: %w", policyNumber, snapErr)
	}
	if snap == nil {
		return resp.BatchPolicyStatus{}, fmt.Errorf("policy %s not found", policyNumber)
	}

	var state wfPolicyState
	if unmarshalErr := json.Unmarshal(snap.FinalSnapshot, &state); unmarshalErr != nil {
		// Minimum fields from snapshot header.
		state.CurrentStatus = snap.FinalStatus
		state.DisplayStatus = snap.FinalStatus
	}

	return resp.BatchPolicyStatus{
		PolicyNumber:    policyNumber,
		LifecycleStatus: state.CurrentStatus,
		DisplayStatus:   state.DisplayStatus,
	}, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Dashboard Metrics (DB-only)
// ─────────────────────────────────────────────────────────────────────────────

// GetDashboardMetrics — GET /v1/policies/dashboard/metrics
// DB-only: reads mv_policy_dashboard (refreshed every 15 min) + request counts.
// No per-workflow query — aggregation across 3M+ workflows is not feasible. [FR-PM-008]
func (h *PolicyQueryHandler) GetDashboardMetrics(sctx *serverRoute.Context, req struct{}) (*resp.DashboardMetricsResponse, error) {
	ctx := sctx.Ctx

	metrics, err := h.policyRepo.GetDashboardMetrics(ctx)
	if err != nil {
		log.Error(ctx, "GetDashboardMetrics: %v", err)
		return nil, err
	}

	// Convert domain map[string]int64 → response map[string]int.
	byStatus := make(map[string]int, len(metrics.PoliciesByStatus))
	for k, v := range metrics.PoliciesByStatus {
		byStatus[k] = int(v)
	}
	byProduct := make(map[string]int, len(metrics.PoliciesByProduct))
	for k, v := range metrics.PoliciesByProduct {
		byProduct[k] = int(v)
	}
	byBilling := make(map[string]int, len(metrics.PoliciesByBillingMethod))
	for k, v := range metrics.PoliciesByBillingMethod {
		byBilling[k] = int(v)
	}

	return &resp.DashboardMetricsResponse{
		StatusCodeAndMessage: port.FetchSuccess,
		Data: resp.DashboardMetricsData{
			PoliciesByStatus:        byStatus,
			PoliciesByProduct:       byProduct,
			PoliciesByBillingMethod: byBilling,
			RequestsToday:           metrics.RequestsToday,
			RequestsPending:         metrics.RequestsPending,
		},
	}, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Two-Tier core helper
// ─────────────────────────────────────────────────────────────────────────────

// twoTierFullState implements the Two-Tier Query Pattern for the full policy state.
// Tier 1: QueryWorkflow("getFullState") → workflow in-memory state (most current)
// Tier 2: GetTerminalSnapshot → terminal snapshot (post-cooling fallback)
// Returns 404 AppError if neither tier finds the policy.
func (h *PolicyQueryHandler) twoTierFullState(ctx context.Context, policyNumber string) (*wfPolicyState, error) {
	wfID := "plw-" + policyNumber

	// Tier 1: Try QueryWorkflow.
	qResp, queryErr := h.tc.QueryWorkflow(ctx, wfID, "", queryGetFullState)
	if queryErr == nil {
		var state wfPolicyState
		if decodeErr := qResp.Get(&state); decodeErr == nil {
			return &state, nil
		}
		log.Warn(ctx, "getFullState decode %s: falling through to Tier 2", policyNumber)
	}

	// Tier 2: Check active policy table.
	policy, err := h.policyRepo.GetPolicyByNumber(ctx, policyNumber)
	if err == nil {
		// Active policy found — convert to wfPolicyState.
		return h.convertPolicyToWFState(policy), nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		log.Error(ctx, "GetPolicyByNumber %s: %v", policyNumber, err)
		return nil, err
	}

	// Tier 3: terminal_state_snapshot fallback.
	snap, snapErr := h.policyRepo.GetTerminalSnapshot(ctx, policyNumber)
	if snapErr != nil {
		log.Error(ctx, "GetTerminalSnapshot %s: %v", policyNumber, snapErr)
		return nil, snapErr
	}
	if snap == nil {
		return nil, newHTTPErr(http.StatusNotFound,
			fmt.Sprintf("policy %s not found", policyNumber), nil)
	}

	var state wfPolicyState
	if unmarshalErr := json.Unmarshal(snap.FinalSnapshot, &state); unmarshalErr != nil {
		// Fallback: populate minimum fields from snapshot header.
		log.Warn(ctx, "parse terminal snapshot %s: %v; using header fields only", policyNumber, unmarshalErr)
		state.PolicyNumber = snap.PolicyNumber
		state.PolicyID = snap.PolicyID
		state.CurrentStatus = snap.FinalStatus
		state.DisplayStatus = snap.FinalStatus
		state.EffectiveFrom = snap.TerminalAt.Format("2006-01-02T15:04:05Z07:00")
		state.UpdatedAt = snap.TerminalAt.Format("2006-01-02T15:04:05Z07:00")
	}
	return &state, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Static helpers
// ─────────────────────────────────────────────────────────────────────────────

// allowedStatusesForType returns the lifecycle statuses that allow the given
// request type. Used to populate StateGateData.AllowedStatuses. [BR-PM-011..023]
func allowedStatusesForType(requestType string) []string {
	switch requestType {
	case domain.RequestTypeSurrender:
		return []string{domain.StatusActive, domain.StatusVoidLapse, domain.StatusInactiveLapse, domain.StatusActiveLapse, domain.StatusPaidUp}
	case domain.RequestTypeLoan:
		return []string{domain.StatusActive}
	case domain.RequestTypeLoanRepayment:
		return []string{domain.StatusActive, domain.StatusAssignedToPresident, domain.StatusPendingAutoSurrender}
	case domain.RequestTypeRevival:
		return []string{domain.StatusVoidLapse, domain.StatusInactiveLapse, domain.StatusActiveLapse}
	case domain.RequestTypeDeathClaim:
		// BR-PM-014, BR-PM-112: Death claim accepted in any non-terminal state including SUSPENDED
		return []string{
			domain.StatusFreeLookActive,
			domain.StatusActive,
			domain.StatusVoidLapse,
			domain.StatusInactiveLapse,
			domain.StatusActiveLapse,
			domain.StatusPaidUp,
			domain.StatusReducedPaidUp,
			domain.StatusAssignedToPresident,
			domain.StatusPendingAutoSurrender,
			domain.StatusPendingSurrender,
			domain.StatusRevivalPending,
			domain.StatusPendingMaturity,
			domain.StatusDeathClaimIntimated,
			domain.StatusDeathUnderInvestigation,
			domain.StatusSuspended,
		}
	case domain.RequestTypeMaturityClaim:
		return []string{domain.StatusActive, domain.StatusPendingMaturity}
	case domain.RequestTypeSurvivalBenefit:
		return []string{domain.StatusActive}
	case domain.RequestTypeCommutation:
		return []string{domain.StatusActive}
	case domain.RequestTypeConversion:
		return []string{domain.StatusActive}
	case domain.RequestTypeFLC:
		return []string{domain.StatusFreeLookActive}
	case domain.RequestTypePaidUp:
		return []string{domain.StatusActive, domain.StatusActiveLapse}
	default:
		// BR-PM-023: NFR and other request types accepted in any non-terminal state (excludes SUSPENDED)
		return []string{
			domain.StatusFreeLookActive,
			domain.StatusActive,
			domain.StatusVoidLapse,
			domain.StatusInactiveLapse,
			domain.StatusActiveLapse,
			domain.StatusPaidUp,
			domain.StatusReducedPaidUp,
			domain.StatusAssignedToPresident,
			domain.StatusPendingAutoSurrender,
			domain.StatusPendingSurrender,
			domain.StatusRevivalPending,
			domain.StatusPendingMaturity,
			domain.StatusDeathClaimIntimated,
			domain.StatusDeathUnderInvestigation,
		}
	}
}

// convertPolicyToWFState converts a domain.Policy to wfPolicyState for query responses.
func (h *PolicyQueryHandler) convertPolicyToWFState(policy *domain.Policy) *wfPolicyState {
	var previousStatus *string
	if policy.PreviousStatus != nil {
		prevStatus := *policy.PreviousStatus
		previousStatus = &prevStatus
	}

	// LoanID and AssigneeID are not stored in domain.Policy - set to nil
	var loanID *int64
	var assigneeID *int64

	return &wfPolicyState{
		PolicyID:       policy.PolicyID,
		PolicyNumber:   policy.PolicyNumber,
		CurrentStatus:  policy.CurrentStatus,
		PreviousStatus: previousStatus,
		Encumbrances: wfEncumbrances{
			HasActiveLoan:   policy.HasActiveLoan,
			LoanID:          loanID, // Not stored in domain.Policy
			LoanOutstanding: policy.LoanOutstanding,
			AssignmentType:  policy.AssignmentType,
			AssigneeID:      assigneeID, // Not stored in domain.Policy
			AMLHold:         policy.AMLHold,
			DisputeFlag:     policy.DisputeFlag,
		},
		DisplayStatus:  policy.DisplayStatus,
		Version:        policy.Version,
		EffectiveFrom:  policy.EffectiveFrom.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:      policy.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		CustomerID:     policy.CustomerID,
		ProductCode:    policy.ProductCode,
		ProductType:    policy.ProductType,
		SumAssured:     policy.SumAssured,
		CurrentPremium: policy.CurrentPremium,
		PremiumMode:    policy.PremiumMode,
		BillingMethod:  policy.BillingMethod,
		AgentID:        policy.AgentID,
		PaidUpValue:    policy.PaidUpValue,
	}
}
