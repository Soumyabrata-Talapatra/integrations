package activities

// ============================================================================
// QuoteActivities — Outbound quote proxy activities + short-lived workflows
//
// Called synchronously from QuoteHandler (NOT from PolicyLifecycleWorkflow).
// Each activity makes a single outbound HTTP call to the downstream service's
// internal calculation API and returns the quote data.
//
// Activity options: StartToCloseTimeout: 10s, RetryPolicy: 3× 2× backoff.
// Workflow execution timeout: 30s (set by QuoteHandler.executeQuoteWorkflow).
//
// Short-lived workflow functions are registered by string name so QuoteHandler
// can invoke them via client.ExecuteWorkflow("GetSurrenderQuoteWorkflow", ...).
//
// [FR-PM-003, §4.2, Constraint 1]
// ============================================================================

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	config "gitlab.cept.gov.in/it-2.0-common/api-config"
)

// ─────────────────────────────────────────────────────────────────────────────
// QuoteActivities struct — injected via FX
// ─────────────────────────────────────────────────────────────────────────────

// QuoteActivities holds dependencies for quote proxy activities.
// No DB dependency — all calls are outbound HTTP to downstream internal APIs. [FR-PM-003]
type QuoteActivities struct {
	cfg        *config.Config
	httpClient *http.Client
}

// NewQuoteActivities constructs a QuoteActivities instance for FX injection.
func NewQuoteActivities(cfg *config.Config) *QuoteActivities {
	return &QuoteActivities{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 8 * time.Second, // Within activity StartToCloseTimeout=10s
		},
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Quote Activity Input / Output Types
// JSON tags must match handler/response quote DTO fields (Temporal JSON decode)
// ─────────────────────────────────────────────────────────────────────────────

// SurrenderBreakdown mirrors resp.SurrenderBreakdown — same JSON tags. [§4.2]
type SurrenderBreakdown struct {
	BaseSVFactor         float64 `json:"base_sv_factor"`
	PremiumsPaidMonths   int     `json:"premiums_paid_months"`
	TotalPremiumsPayable int     `json:"total_premiums_payable"`
}

// SurrenderQuoteResult is the activity return type for GetSurrenderQuoteActivity.
// JSON tags match handler/response.SurrenderQuoteData so wfRun.Get() works. [§4.2]
type SurrenderQuoteResult struct {
	PolicyNumber        string             `json:"policy_number"`
	QuoteType           string             `json:"quote_type"` // "SURRENDER"
	GrossSurrenderValue float64            `json:"gross_surrender_value"`
	BonusAccumulated    float64            `json:"bonus_accumulated"`
	LoanDeduction       float64            `json:"loan_deduction"`
	InterestDeduction   float64            `json:"interest_deduction"`
	NetSurrenderValue   float64            `json:"net_surrender_value"`
	Breakdown           SurrenderBreakdown `json:"breakdown"`
	ValidUntil          string             `json:"valid_until"` // RFC3339
}

// LoanQuoteResult is the activity return type for GetLoanQuoteActivity.
// JSON tags match handler/response.LoanQuoteData. [§4.2]
type LoanQuoteResult struct {
	PolicyNumber        string  `json:"policy_number"`
	Eligible            bool    `json:"eligible"`
	MaxLoanAmount       float64 `json:"max_loan_amount"`
	InterestRate        float64 `json:"interest_rate"`
	SurrenderValue      float64 `json:"surrender_value"`
	ExistingLoanBalance float64 `json:"existing_loan_balance"`
	IneligibilityReason *string `json:"ineligibility_reason,omitempty"`
}

// ConversionOption mirrors resp.ConversionOption — same JSON tags. [§4.2]
type ConversionOption struct {
	TargetProduct     string  `json:"target_product"`
	EffectiveDate     string  `json:"effective_date"`
	NewPremium        float64 `json:"new_premium"`
	PremiumDifference float64 `json:"premium_difference"`
	NewSumAssured     float64 `json:"new_sum_assured"`
}

// ConversionQuoteResult is the activity return type for GetConversionQuoteActivity.
// JSON tags match handler/response.ConversionQuoteData. [§4.2]
type ConversionQuoteResult struct {
	PolicyNumber         string             `json:"policy_number"`
	CurrentProduct       string             `json:"current_product"`
	AvailableConversions []ConversionOption `json:"available_conversions"`
}

// ─────────────────────────────────────────────────────────────────────────────
// Quote Workflow Input types (matching QuoteHandler request types)
// ─────────────────────────────────────────────────────────────────────────────

// QuoteWorkflowInput is the input for surrender and loan quote workflows. [§4.2]
type QuoteWorkflowInput struct {
	PolicyNumber string `json:"policy_number"`
	AsOfDate     string `json:"as_of_date"` // "2006-01-02" — defaults to today if empty
}

// ConversionQuoteWorkflowInput is the input for the conversion quote workflow. [§4.2]
type ConversionQuoteWorkflowInput struct {
	PolicyNumber      string `json:"policy_number"`
	AsOfDate          string `json:"as_of_date"`
	TargetProductCode string `json:"target_product_code"`
}

// ─────────────────────────────────────────────────────────────────────────────
// Activity option helper for quote activities (10s, 3× 2× backoff)
// ─────────────────────────────────────────────────────────────────────────────

func quoteActCtx(ctx workflow.Context) workflow.Context {
	return workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 10 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts:    3,
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    5 * time.Second,
		},
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// Short-lived Quote Workflows
// Registered by function name so QuoteHandler.executeQuoteWorkflow() can call
// them via client.ExecuteWorkflow("GetSurrenderQuoteWorkflow", ...).
// [FR-PM-003, §4.2, Constraint 2 — no REST endpoint; Temporal-only]
// ─────────────────────────────────────────────────────────────────────────────

// GetSurrenderQuoteWorkflow is a short-lived workflow triggered by QuoteHandler.
// Runs a single GetSurrenderQuoteActivity and returns the result. [FR-PM-003]
func GetSurrenderQuoteWorkflow(ctx workflow.Context, input QuoteWorkflowInput) (*SurrenderQuoteResult, error) {
	var quoteActs QuoteActivities
	var result SurrenderQuoteResult
	if err := workflow.ExecuteActivity(quoteActCtx(ctx),
		quoteActs.GetSurrenderQuoteActivity,
		input.PolicyNumber, input.AsOfDate,
	).Get(ctx, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetLoanQuoteWorkflow is a short-lived workflow triggered by QuoteHandler.
// Runs a single GetLoanQuoteActivity and returns the result. [FR-PM-003]
func GetLoanQuoteWorkflow(ctx workflow.Context, input QuoteWorkflowInput) (*LoanQuoteResult, error) {
	var quoteActs QuoteActivities
	var result LoanQuoteResult
	if err := workflow.ExecuteActivity(quoteActCtx(ctx),
		quoteActs.GetLoanQuoteActivity,
		input.PolicyNumber, input.AsOfDate,
	).Get(ctx, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetConversionQuoteWorkflow is a short-lived workflow triggered by QuoteHandler.
// Runs a single GetConversionQuoteActivity and returns the result. [FR-PM-003]
func GetConversionQuoteWorkflow(ctx workflow.Context, input ConversionQuoteWorkflowInput) (*ConversionQuoteResult, error) {
	var quoteActs QuoteActivities
	var result ConversionQuoteResult
	if err := workflow.ExecuteActivity(quoteActCtx(ctx),
		quoteActs.GetConversionQuoteActivity,
		input.PolicyNumber, input.AsOfDate, input.TargetProductCode,
	).Get(ctx, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// GetSurrenderQuoteActivity — calls surrender-svc internal quote endpoint [§4.2]
// StartToCloseTimeout: 10s, RetryPolicy: 3× 2× backoff (fast fail — sync quote)
// ─────────────────────────────────────────────────────────────────────────────

// GetSurrenderQuoteActivity calls the surrender-svc internal calculation API to
// retrieve GSV, bonus, loan deduction, and net surrender value. [FR-PM-003, §4.2]
func (a *QuoteActivities) GetSurrenderQuoteActivity(ctx context.Context, policyNumber, asOfDate string) (*SurrenderQuoteResult, error) {
	baseURL := a.cfg.GetString("services.surrender_svc.internal_url")
	if baseURL == "" {
		return nil, fmt.Errorf("GetSurrenderQuoteActivity: services.surrender_svc.internal_url not configured")
	}

	// [C4] url.PathEscape encodes '/' in policy numbers (e.g. PLI/2026/000001 → PLI%2F2026%2F000001)
	reqURL := fmt.Sprintf("%s/internal/v1/policies/%s/surrender-quote", baseURL, url.PathEscape(policyNumber))
	if asOfDate != "" {
		reqURL = fmt.Sprintf("%s?as_of=%s", reqURL, url.QueryEscape(asOfDate))
	}

	result, err := a.getJSON(ctx, reqURL)
	if err != nil {
		return nil, fmt.Errorf("GetSurrenderQuoteActivity policy=%s: %w", policyNumber, err)
	}

	var quote SurrenderQuoteResult
	if err := json.Unmarshal(result, &quote); err != nil {
		return nil, fmt.Errorf("GetSurrenderQuoteActivity decode policy=%s: %w", policyNumber, err)
	}
	return &quote, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// GetLoanQuoteActivity — calls loan-svc internal eligibility endpoint [§4.2]
// StartToCloseTimeout: 10s, RetryPolicy: 3× 2× backoff
// ─────────────────────────────────────────────────────────────────────────────

// GetLoanQuoteActivity calls the loan-svc internal API to retrieve loan eligibility,
// maximum loan amount, interest rate, and existing loan balance. [FR-PM-003, §4.2]
func (a *QuoteActivities) GetLoanQuoteActivity(ctx context.Context, policyNumber, asOfDate string) (*LoanQuoteResult, error) {
	baseURL := a.cfg.GetString("services.loan_svc.internal_url")
	if baseURL == "" {
		return nil, fmt.Errorf("GetLoanQuoteActivity: services.loan_svc.internal_url not configured")
	}

	// [C4] url.PathEscape handles PLI/YYYY/NNNNNN format
	reqURL := fmt.Sprintf("%s/internal/v1/policies/%s/loan-eligibility", baseURL, url.PathEscape(policyNumber))
	if asOfDate != "" {
		reqURL = fmt.Sprintf("%s?as_of=%s", reqURL, url.QueryEscape(asOfDate))
	}

	result, err := a.getJSON(ctx, reqURL)
	if err != nil {
		return nil, fmt.Errorf("GetLoanQuoteActivity policy=%s: %w", policyNumber, err)
	}

	var quote LoanQuoteResult
	if err := json.Unmarshal(result, &quote); err != nil {
		return nil, fmt.Errorf("GetLoanQuoteActivity decode policy=%s: %w", policyNumber, err)
	}
	return &quote, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// GetConversionQuoteActivity — calls conversion-svc internal options endpoint [§4.2]
// StartToCloseTimeout: 10s, RetryPolicy: 3× 2× backoff
// ─────────────────────────────────────────────────────────────────────────────

// GetConversionQuoteActivity calls the conversion-svc internal API to retrieve
// available conversion options: premium diff, converted SA, conversion terms.
// [FR-PM-003, §4.2]
func (a *QuoteActivities) GetConversionQuoteActivity(ctx context.Context, policyNumber, asOfDate, targetProductCode string) (*ConversionQuoteResult, error) {
	baseURL := a.cfg.GetString("services.conversion_svc.internal_url")
	if baseURL == "" {
		return nil, fmt.Errorf("GetConversionQuoteActivity: services.conversion_svc.internal_url not configured")
	}

	// [C4] url.PathEscape handles PLI/YYYY/NNNNNN format; url.QueryEscape encodes query params
	reqURL := fmt.Sprintf("%s/internal/v1/policies/%s/conversion-options", baseURL, url.PathEscape(policyNumber))
	if asOfDate != "" || targetProductCode != "" {
		sep := "?"
		if asOfDate != "" {
			reqURL = fmt.Sprintf("%s%sas_of=%s", reqURL, sep, url.QueryEscape(asOfDate))
			sep = "&"
		}
		if targetProductCode != "" {
			reqURL = fmt.Sprintf("%s%starget_product=%s", reqURL, sep, url.QueryEscape(targetProductCode))
		}
	}

	result, err := a.getJSON(ctx, reqURL)
	if err != nil {
		return nil, fmt.Errorf("GetConversionQuoteActivity policy=%s: %w", policyNumber, err)
	}

	var quote ConversionQuoteResult
	if err := json.Unmarshal(result, &quote); err != nil {
		return nil, fmt.Errorf("GetConversionQuoteActivity decode policy=%s: %w", policyNumber, err)
	}
	return &quote, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// HTTP helper
// ─────────────────────────────────────────────────────────────────────────────

// getJSON performs a GET request against an internal service URL and returns
// the raw JSON body. Propagates context cancellation for timeout enforcement.
func (a *QuoteActivities) getJSON(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "policy-management/1.0")
	// Propagate workflow run ID as X-Request-ID for downstream correlation tracing [Review-Fix-17]
	if info := activity.GetInfo(ctx); info.WorkflowExecution.RunID != "" {
		req.Header.Set("X-Request-ID", info.WorkflowExecution.RunID)
	}

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("GET %s: %w", url, err)
	}
	defer resp.Body.Close()

	// Limit response to 1 MiB to prevent memory exhaustion from a runaway upstream. [D12]
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return nil, fmt.Errorf("read body from %s: %w", url, err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("GET %s returned HTTP %d: %s", url, resp.StatusCode, string(body))
	}

	return body, nil
}
