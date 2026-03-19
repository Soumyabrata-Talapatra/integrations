package handler

// QuoteHandler — synchronous quote proxy endpoints (Step 4.2)
//
// Implements 3 endpoints:
//   - POST /policies/{pn}/quotes/surrender — state gate check → GetSurrenderQuoteWorkflow
//   - POST /policies/{pn}/quotes/loan      — state gate check → GetLoanQuoteWorkflow
//   - POST /policies/{pn}/quotes/conversion — state gate check → GetConversionQuoteWorkflow
//
// Quote Proxy Mechanism (§4.2):
//   Handlers are synchronous — they call a short-lived Temporal workflow that
//   runs a single GetXxxQuoteActivity on policy-management-tq. The activity
//   calls the downstream service's internal calculation API and returns the
//   quote result to the handler via workflowRun.Get().
//   Workflow execution timeout: 30s (covers 2× 10s activity + network latency).
//   Activity StartToCloseTimeout: 10s with 2× backoff retry (3 max attempts).

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
	log "gitlab.cept.gov.in/it-2.0-common/n-api-log"
	serverHandler "gitlab.cept.gov.in/it-2.0-common/n-api-server/handler"
	serverRoute "gitlab.cept.gov.in/it-2.0-common/n-api-server/route"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"

	"policy-management/core/domain"
	"policy-management/core/port"
	resp "policy-management/handler/response"
	repo "policy-management/repo/postgres"
)

// ─────────────────────────────────────────────────────────────────────────────
// Temporal workflow names for quote proxies (Phase 5 will register these)
// ─────────────────────────────────────────────────────────────────────────────

const (
	quoteWFSurrender  = "GetSurrenderQuoteWorkflow"
	quoteWFLoan       = "GetLoanQuoteWorkflow"
	quoteWFConversion = "GetConversionQuoteWorkflow"

	quoteWorkflowTimeout = 30 * time.Second
)

// ─────────────────────────────────────────────────────────────────────────────
// Quote workflow input types
// ─────────────────────────────────────────────────────────────────────────────

// QuoteWorkflowInput is the standard input for surrender and loan quote workflows.
// Phase 5 GetSurrenderQuoteActivity / GetLoanQuoteActivity use AsOfDate to
// compute quotes as-of a specific valuation date.
type QuoteWorkflowInput struct {
	PolicyNumber string `json:"policy_number"`
	AsOfDate     string `json:"as_of_date"` // "2006-01-02" — defaults to today if empty
}

// ConversionQuoteWorkflowInput extends QuoteWorkflowInput with the target product.
// Phase 5 GetConversionQuoteActivity uses TargetProductCode to fetch conversion options.
type ConversionQuoteWorkflowInput struct {
	PolicyNumber      string `json:"policy_number"`
	AsOfDate          string `json:"as_of_date"`
	TargetProductCode string `json:"target_product_code"`
}

// ─────────────────────────────────────────────────────────────────────────────
// QuoteHandler
// ─────────────────────────────────────────────────────────────────────────────

// QuoteHandler handles all synchronous quote proxy endpoints.
// [FR-PM-003, §4.2 Quote Proxy Mechanism]
type QuoteHandler struct {
	*serverHandler.Base
	policyRepo *repo.PolicyRepository
	tc         client.Client
}

// NewQuoteHandler constructs a QuoteHandler with required dependencies.
func NewQuoteHandler(
	policyRepo *repo.PolicyRepository,
	tc client.Client,
) *QuoteHandler {
	base := serverHandler.New("Quotes").SetPrefix("/v1").AddPrefix("")
	return &QuoteHandler{
		Base:       base,
		policyRepo: policyRepo,
		tc:         tc,
	}
}

// Routes registers all 3 quote proxy endpoints.
func (h *QuoteHandler) Routes() []serverRoute.Route {
	return []serverRoute.Route{
		serverRoute.POST("/policies/:policy_number/quotes/surrender", h.GetSurrenderQuote).
			Name("Get Surrender Quote"),
		serverRoute.POST("/policies/:policy_number/quotes/loan", h.GetLoanQuote).
			Name("Get Loan Quote"),
		serverRoute.POST("/policies/:policy_number/quotes/conversion", h.GetConversionQuote).
			Name("Get Conversion Quote"),
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Combined request types (uri: + json: binding on same struct)
// ─────────────────────────────────────────────────────────────────────────────

type surrenderQuoteReq struct {
	PolicyNumberURI
	QuoteRequest
}

type loanQuoteReq struct {
	PolicyNumberURI
	QuoteRequest
}

type conversionQuoteReq struct {
	PolicyNumberURI
	ConversionQuoteRequest
}

// ─────────────────────────────────────────────────────────────────────────────
// Quote Proxy Endpoints
// ─────────────────────────────────────────────────────────────────────────────

// GetSurrenderQuote — POST /v1/policies/:policy_number/quotes/surrender
// State gate: ACTIVE, VOID_LAPSE, INACTIVE_LAPSE, ACTIVE_LAPSE, PAID_UP (BR-PM-011 logic).
// Calls GetSurrenderQuoteWorkflow synchronously via ExecuteWorkflow + Get.
// [FR-PM-003]
func (h *QuoteHandler) GetSurrenderQuote(sctx *serverRoute.Context, req surrenderQuoteReq) (*resp.SurrenderQuoteResponse, error) {
	policyNumber := req.PolicyNumber
	ctx := sctx.Ctx

	policy, err := h.lookupPolicy(ctx, policyNumber)
	if err != nil {
		return nil, err
	}

	// State gate: surrender eligibility (mirrors BR-PM-011)
	if sgErr := checkStateGate(policy, domain.RequestTypeSurrender); sgErr != nil {
		return nil, newHTTPErr(http.StatusUnprocessableEntity,
			fmt.Sprintf("surrender quote not available: %s", sgErr.Error()), sgErr)
	}

	asOfDate := normaliseAsOfDate(req.AsOfDate)
	input := QuoteWorkflowInput{
		PolicyNumber: policyNumber,
		AsOfDate:     asOfDate,
	}

	var quoteData resp.SurrenderQuoteData
	if err := h.executeQuoteWorkflow(ctx, quoteWFSurrender, policyNumber, input, &quoteData); err != nil {
		return nil, err
	}

	return &resp.SurrenderQuoteResponse{
		StatusCodeAndMessage: port.FetchSuccess,
		Data:                 quoteData,
	}, nil
}

// GetLoanQuote — POST /v1/policies/:policy_number/quotes/loan
// State gate: ACTIVE, no active loan (BR-PM-012 logic).
// Calls GetLoanQuoteWorkflow synchronously via ExecuteWorkflow + Get.
// [FR-PM-003]
func (h *QuoteHandler) GetLoanQuote(sctx *serverRoute.Context, req loanQuoteReq) (*resp.LoanQuoteResponse, error) {
	policyNumber := req.PolicyNumber
	ctx := sctx.Ctx

	policy, err := h.lookupPolicy(ctx, policyNumber)
	if err != nil {
		return nil, err
	}

	// State gate: loan eligibility (mirrors BR-PM-012)
	if sgErr := checkStateGate(policy, domain.RequestTypeLoan); sgErr != nil {
		return nil, newHTTPErr(http.StatusUnprocessableEntity,
			fmt.Sprintf("loan quote not available: %s", sgErr.Error()), sgErr)
	}

	asOfDate := normaliseAsOfDate(req.AsOfDate)
	input := QuoteWorkflowInput{
		PolicyNumber: policyNumber,
		AsOfDate:     asOfDate,
	}

	var quoteData resp.LoanQuoteData
	if err := h.executeQuoteWorkflow(ctx, quoteWFLoan, policyNumber, input, &quoteData); err != nil {
		return nil, err
	}

	return &resp.LoanQuoteResponse{
		StatusCodeAndMessage: port.FetchSuccess,
		Data:                 quoteData,
	}, nil
}

// GetConversionQuote — POST /v1/policies/:policy_number/quotes/conversion
// State gate: ACTIVE (BR-PM-018 logic).
// Calls GetConversionQuoteWorkflow synchronously via ExecuteWorkflow + Get.
// [FR-PM-003]
func (h *QuoteHandler) GetConversionQuote(sctx *serverRoute.Context, req conversionQuoteReq) (*resp.ConversionQuoteResponse, error) {
	policyNumber := req.PolicyNumber
	ctx := sctx.Ctx

	policy, err := h.lookupPolicy(ctx, policyNumber)
	if err != nil {
		return nil, err
	}

	// State gate: conversion eligibility (mirrors BR-PM-018)
	if sgErr := checkStateGate(policy, domain.RequestTypeConversion); sgErr != nil {
		return nil, newHTTPErr(http.StatusUnprocessableEntity,
			fmt.Sprintf("conversion quote not available: %s", sgErr.Error()), sgErr)
	}

	asOfDate := normaliseAsOfDate(req.AsOfDate)
	input := ConversionQuoteWorkflowInput{
		PolicyNumber:      policyNumber,
		AsOfDate:          asOfDate,
		TargetProductCode: req.TargetProductCode,
	}

	var quoteData resp.ConversionQuoteData
	if err := h.executeQuoteWorkflow(ctx, quoteWFConversion, policyNumber, input, &quoteData); err != nil {
		return nil, err
	}

	return &resp.ConversionQuoteResponse{
		StatusCodeAndMessage: port.FetchSuccess,
		Data:                 quoteData,
	}, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Private helpers
// ─────────────────────────────────────────────────────────────────────────────

// lookupPolicy fetches a policy by number, returning a 404 AppError if not found.
func (h *QuoteHandler) lookupPolicy(ctx context.Context, policyNumber string) (*domain.Policy, error) {
	policy, err := h.policyRepo.GetPolicyByNumber(ctx, policyNumber)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, newHTTPErr(http.StatusNotFound,
				fmt.Sprintf("policy %s not found", policyNumber), err)
		}
		log.Error(ctx, "GetPolicyByNumber %s: %v", policyNumber, err)
		return nil, err
	}
	return policy, nil
}

// executeQuoteWorkflow starts a short-lived quote workflow, waits for the result,
// and unmarshals it into resultPtr. The workflow runs a single GetXxxQuoteActivity
// on policy-management-tq that calls the downstream service's calculation API.
//
// Execution timeout: 30s. Activity retry policy: 3 attempts, 2× backoff (§4.2).
func (h *QuoteHandler) executeQuoteWorkflow(
	ctx context.Context,
	workflowType string,
	policyNumber string,
	input interface{},
	resultPtr interface{},
) error {
	quoteCtx, cancel := context.WithTimeout(ctx, quoteWorkflowTimeout)
	defer cancel()

	opts := client.StartWorkflowOptions{
		TaskQueue:                "policy-management-tq",
		WorkflowExecutionTimeout: quoteWorkflowTimeout,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts:    3,
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    10 * time.Second,
		},
	}

	wfRun, err := h.tc.ExecuteWorkflow(quoteCtx, opts, workflowType, input)
	if err != nil {
		log.Error(ctx, "%s start failed policy=%s: %v", workflowType, policyNumber, err)
		return newHTTPErr(http.StatusServiceUnavailable,
			"quote service unavailable — please retry", err)
	}

	if err := wfRun.Get(quoteCtx, resultPtr); err != nil {
		log.Error(ctx, "%s result failed policy=%s: %v", workflowType, policyNumber, err)
		return newHTTPErr(http.StatusServiceUnavailable,
			"quote retrieval failed — please retry", err)
	}

	return nil
}

// normaliseAsOfDate returns asOfDate if non-empty, otherwise today's date in "2006-01-02" format.
func normaliseAsOfDate(asOfDate string) string {
	if asOfDate != "" {
		return asOfDate
	}
	return time.Now().Format("2006-01-02")
}
