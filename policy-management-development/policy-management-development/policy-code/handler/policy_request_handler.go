package handler

// PolicyRequestHandler — all policy request submission endpoints (Step 4.1)
//
// Implements 19 endpoints:
//   - 11 financial submission endpoints (Temporal signal → PLW workflow)
//   - 6 non-financial (NFR) submission endpoints (Temporal signal → PLW workflow)
//   - 2 admin/system endpoints (admin-void signal; reopen via SignalWithStart)
//
// Standard handler flow for all submission endpoints [§9.5.2]:
//  1. Bind & Validate (struct tags + custom Validate())
//  2. Check X-Idempotency-Key — return original 202 if duplicate
//  3. Lookup policy by number → 404 if not found
//  4. State gate fast-fail pre-check → 422 if rejected [BR-PM-011..BR-PM-023]
//  5. Financial lock check (financial only) → 409 if locked [BR-PM-030]
//  6. INSERT service_request (status=RECEIVED)
//  7. Send Temporal signal to plw-{policy_number}
//  8. UPDATE service_request status=ROUTED
//  9. Return 202 with RequestID

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	apierrors "gitlab.cept.gov.in/it-2.0-common/n-api-errors"
	log "gitlab.cept.gov.in/it-2.0-common/n-api-log"
	serverHandler "gitlab.cept.gov.in/it-2.0-common/n-api-server/handler"
	serverRoute "gitlab.cept.gov.in/it-2.0-common/n-api-server/route"
	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"

	"policy-management/core/domain"
	"policy-management/core/port"
	resp "policy-management/handler/response"
	repo "policy-management/repo/postgres"
)

// ─────────────────────────────────────────────────────────────────────────────
// Signal channel name constants sent from handlers to plw-{policyNumber}
// These match the constants that Phase 5 workflows/signals.go will define.
// ─────────────────────────────────────────────────────────────────────────────

const (
	signalSurrenderRequest       = "surrender-request"
	signalLoanRequest            = "loan-request"
	signalLoanRepaymentRequest   = "loan-repayment"      // §9.1 canonical name [signals.go]
	signalRevivalRequest         = "revival-request"
	signalDeathClaimRequest      = "death-notification"  // §9.1 canonical name [signals.go, BR-PM-112]
	signalMaturityClaimRequest   = "maturity-claim-request"
	signalSurvivalBenefitRequest = "survival-benefit-request"
	signalCommutationRequest     = "commutation-request"
	signalConversionRequest      = "conversion-request"
	signalFLCRequest             = "flc-request"
	signalVoluntaryPaidUpRequest = "voluntary-paidup-request"
	signalNFRRequest             = "nfr-request" // shared for all NFR types
	signalAdminVoid              = "admin-void"
	signalReopenRequest          = "reopen-request"
)

// policyRequestSignal is the minimal payload from a handler to the PLW workflow
// when a service request is submitted. The workflow fetches full details from DB.
// IdempotencyKey is the UUID idempotency key — used by the workflow as the dedup
// key and as the child workflow ID fragment per Constraint 1. [Review-Fix-11, Review-Fix-3]
type policyRequestSignal struct {
	RequestID       int64  `json:"request_id"`
	IdempotencyKey  string `json:"idempotency_key"` // UUID from X-Idempotency-Key header
	RequestType     string `json:"request_type"`
	RequestCategory string `json:"request_category"`
	SourceChannel   string `json:"source_channel"`
	SubmittedBy     *int64 `json:"submitted_by,omitempty"`
}

// adminVoidSignalPayload is the payload for the admin-void signal. [BR-PM-073]
// JSON tags MUST match workflows.AdminVoidSignal exactly.
type adminVoidSignalPayload struct {
	RequestID    string `json:"request_id"`    // dedup key for PLW ProcessedSignalIDs
	Reason       string `json:"reason"`
	AuthorizedBy int64  `json:"authorized_by"` // was "voided_by" — mismatch fixed
}

// reopenSignalPayload is the signal sent to the PLW workflow on reopen. [BR-PM-090+]
// JSON tags MUST match workflows.ReopenRequestSignal exactly.
type reopenSignalPayload struct {
	RequestID    string `json:"request_id"`    // dedup key for PLW ProcessedSignalIDs
	ReopenReason string `json:"reopen_reason"` // was "reason" — mismatch fixed
	AuthorizedBy int64  `json:"authorized_by"` // was "reopened_by" — mismatch fixed
}

// ─────────────────────────────────────────────────────────────────────────────
// PolicyRequestHandler
// ─────────────────────────────────────────────────────────────────────────────

// PolicyRequestHandler handles all policy request submission and admin endpoints.
// [FR-PM-001, FR-PM-005, FR-PM-006, BR-PM-011..BR-PM-023, BR-PM-030, BR-PM-073]
type PolicyRequestHandler struct {
	*serverHandler.Base
	policyRepo *repo.PolicyRepository
	srRepo     *repo.ServiceRequestRepository
	tc         client.Client
}

// NewPolicyRequestHandler constructs a PolicyRequestHandler with all required dependencies.
func NewPolicyRequestHandler(
	policyRepo *repo.PolicyRepository,
	srRepo *repo.ServiceRequestRepository,
	tc client.Client,
) *PolicyRequestHandler {
	base := serverHandler.New("PolicyRequests").SetPrefix("/v1").AddPrefix("")
	return &PolicyRequestHandler{
		Base:       base,
		policyRepo: policyRepo,
		srRepo:     srRepo,
		tc:         tc,
	}
}

// Routes registers all 19 endpoints for this handler.
func (h *PolicyRequestHandler) Routes() []serverRoute.Route {
	return []serverRoute.Route{
		// ── Financial Submission Endpoints ──────────────────────────────────
		serverRoute.POST("/policies/:policy_number/requests/surrender", h.SubmitSurrenderRequest).
			Name("Submit Surrender Request"),
		serverRoute.POST("/policies/:policy_number/requests/loan", h.SubmitLoanRequest).
			Name("Submit Loan Request"),
		serverRoute.POST("/policies/:policy_number/requests/loan-repayment", h.SubmitLoanRepayment).
			Name("Submit Loan Repayment"),
		serverRoute.POST("/policies/:policy_number/requests/revival", h.SubmitRevivalRequest).
			Name("Submit Revival Request"),
		serverRoute.POST("/policies/:policy_number/requests/death-claim", h.SubmitDeathClaim).
			Name("Submit Death Claim"),
		serverRoute.POST("/policies/:policy_number/requests/maturity-claim", h.SubmitMaturityClaim).
			Name("Submit Maturity Claim"),
		serverRoute.POST("/policies/:policy_number/requests/survival-benefit", h.SubmitSurvivalBenefit).
			Name("Submit Survival Benefit"),
		serverRoute.POST("/policies/:policy_number/requests/commutation", h.SubmitCommutationRequest).
			Name("Submit Commutation Request"),
		serverRoute.POST("/policies/:policy_number/requests/conversion", h.SubmitConversionRequest).
			Name("Submit Conversion Request"),
		serverRoute.POST("/policies/:policy_number/requests/freelook", h.SubmitFreelookCancellation).
			Name("Submit Freelook Cancellation"),
		serverRoute.POST("/policies/:policy_number/requests/paid-up", h.SubmitVoluntaryPaidUp).
			Name("Submit Voluntary Paid-Up"),

		// ── NFR Submission Endpoints ─────────────────────────────────────────
		serverRoute.POST("/policies/:policy_number/requests/nomination-change", h.SubmitNominationChange).
			Name("Submit Nomination Change"),
		serverRoute.POST("/policies/:policy_number/requests/billing-method-change", h.SubmitBillingMethodChange).
			Name("Submit Billing Method Change"),
		serverRoute.POST("/policies/:policy_number/requests/assignment", h.SubmitAssignment).
			Name("Submit Assignment"),
		serverRoute.POST("/policies/:policy_number/requests/address-change", h.SubmitAddressChange).
			Name("Submit Address Change"),
		serverRoute.POST("/policies/:policy_number/requests/premium-refund", h.SubmitPremiumRefund).
			Name("Submit Premium Refund"),
		serverRoute.POST("/policies/:policy_number/requests/duplicate-bond", h.SubmitDuplicateBond).
			Name("Submit Duplicate Bond"),

		// ── Admin / System Endpoints ─────────────────────────────────────────
		serverRoute.POST("/policies/:policy_number/requests/admin-void", h.AdminVoidPolicy).
			Name("Admin Void Policy"),
		serverRoute.POST("/policies/:policy_number/requests/reopen", h.ReopenPolicy).
			Name("Reopen Policy"),
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Combined request types — embed PolicyNumberURI for :policy_number URI binding.
// The framework binds uri: tags via ShouldBindUri AND json: tags via ShouldBind
// on the same struct (route_improved.go buildImproved pattern).
// The Validate() method is promoted from the embedded body struct.
// ─────────────────────────────────────────────────────────────────────────────

type surrenderReq struct {
	PolicyNumberURI
	SubmitSurrenderRequest
}

type loanReq struct {
	PolicyNumberURI
	SubmitLoanRequest
}

type loanRepaymentReq struct {
	PolicyNumberURI
	SubmitLoanRepaymentRequest
}

type revivalReq struct {
	PolicyNumberURI
	SubmitRevivalRequest
}

type deathClaimReq struct {
	PolicyNumberURI
	SubmitDeathClaimRequest
}

type maturityClaimReq struct {
	PolicyNumberURI
	SubmitMaturityClaimRequest
}

type survivalBenefitReq struct {
	PolicyNumberURI
	SubmitSurvivalBenefitRequest
}

type commutationReq struct {
	PolicyNumberURI
	SubmitCommutationRequest
}

type conversionReq struct {
	PolicyNumberURI
	SubmitConversionRequest
}

type freelookReq struct {
	PolicyNumberURI
	SubmitFreelookRequest
}

type paidUpReq struct {
	PolicyNumberURI
	SubmitPaidUpRequest
}

type nominationChangeReq struct {
	PolicyNumberURI
	SubmitNominationChangeRequest
}

type billingMethodChangeReq struct {
	PolicyNumberURI
	SubmitBillingMethodChangeRequest
}

type assignmentReq struct {
	PolicyNumberURI
	SubmitAssignmentRequest
}

type addressChangeReq struct {
	PolicyNumberURI
	SubmitAddressChangeRequest
}

type premiumRefundReq struct {
	PolicyNumberURI
	SubmitPremiumRefundRequest
}

type duplicateBondReq struct {
	PolicyNumberURI
	SubmitDuplicateBondRequest
}

type adminVoidReq struct {
	PolicyNumberURI
	AdminVoidPolicyRequest
}

type reopenReq struct {
	PolicyNumberURI
	ReopenPolicyRequest
}

// ─────────────────────────────────────────────────────────────────────────────
// Private helpers
// ─────────────────────────────────────────────────────────────────────────────

// policyWorkflowID returns the PLW workflow ID for a given policy number.
func policyWorkflowID(policyNumber string) string {
	return "plw-" + policyNumber
}

// newHTTPErr creates a *AppError that the error-handler middleware maps to the
// given HTTP status code. [n-api-errors HandleCommonError]
func newHTTPErr(code int, msg string, cause error) error {
	e := apierrors.NewAppError(msg, code, cause)
	return &e
}

// getIdempotencyKey extracts the X-Idempotency-Key header from the context.
// The framework stores request headers under "RequestHeaders" key.
func getIdempotencyKey(ctx context.Context) string {
	if h, ok := ctx.Value("RequestHeaders").(http.Header); ok {
		return h.Get("X-Idempotency-Key")
	}
	return ""
}

// strPtrIfNonEmpty returns a pointer to s, or nil if s is empty.
func strPtrIfNonEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// checkStateGate returns an error if the policy's current lifecycle status does not
// allow the given request type. This is the fast-fail handler pre-check — the
// authoritative check is inside the PLW workflow after RefreshStateFromDBActivity.
// [BR-PM-011..BR-PM-023, §9.5.2]
func checkStateGate(policy *domain.Policy, requestType string) error {
	status := policy.CurrentStatus

	// Terminal policies block all new service requests.
	if domain.TerminalStatuses[status] {
		return fmt.Errorf("policy %s is in terminal status %s and cannot accept new requests",
			policy.PolicyNumber, status)
	}

	switch requestType {
	case domain.RequestTypeSurrender: // BR-PM-011
		if !inStatusSet(status,
			domain.StatusActive,
			domain.StatusVoidLapse,
			domain.StatusInactiveLapse,
			domain.StatusActiveLapse,
			domain.StatusPaidUp,
		) {
			return fmt.Errorf("surrender request requires ACTIVE, VOID_LAPSE, INACTIVE_LAPSE, ACTIVE_LAPSE or PAID_UP status; current: %s", status)
		}

	case domain.RequestTypeLoan: // BR-PM-012
		if status != domain.StatusActive {
			return fmt.Errorf("loan request requires ACTIVE status; current: %s", status)
		}
		if policy.HasActiveLoan {
			return fmt.Errorf("loan request rejected: policy already has an active loan")
		}

	case domain.RequestTypeLoanRepayment: // BR-PM-020
		if !inStatusSet(status,
			domain.StatusActive,
			domain.StatusAssignedToPresident,
			domain.StatusPendingAutoSurrender,
		) {
			return fmt.Errorf("loan repayment requires ACTIVE, ASSIGNED_TO_PRESIDENT or PENDING_AUTO_SURRENDER status; current: %s", status)
		}
		if !policy.HasActiveLoan {
			return fmt.Errorf("loan repayment requires an active loan on the policy")
		}

	case domain.RequestTypeRevival: // BR-PM-013
		if !inStatusSet(status,
			domain.StatusVoidLapse,
			domain.StatusInactiveLapse,
			domain.StatusActiveLapse,
		) {
			return fmt.Errorf("revival request requires VOID_LAPSE, INACTIVE_LAPSE or ACTIVE_LAPSE status; current: %s", status)
		}

	case domain.RequestTypeDeathClaim: // BR-PM-014: all non-terminal (including SUSPENDED BR-PM-112)
		// Terminal check already done above. All non-terminal statuses allowed.

	case domain.RequestTypeMaturityClaim: // BR-PM-015
		if !inStatusSet(status,
			domain.StatusActive,
			domain.StatusPendingMaturity,
		) {
			return fmt.Errorf("maturity claim requires ACTIVE or PENDING_MATURITY status; current: %s", status)
		}

	case domain.RequestTypeSurvivalBenefit: // BR-PM-016
		if status != domain.StatusActive {
			return fmt.Errorf("survival benefit request requires ACTIVE status; current: %s", status)
		}

	case domain.RequestTypeCommutation: // BR-PM-017
		if status != domain.StatusActive {
			return fmt.Errorf("commutation request requires ACTIVE status; current: %s", status)
		}

	case domain.RequestTypeConversion: // BR-PM-018
		if status != domain.StatusActive {
			return fmt.Errorf("conversion request requires ACTIVE status; current: %s", status)
		}

	case domain.RequestTypeFLC: // BR-PM-019
		if status != domain.StatusFreeLookActive {
			return fmt.Errorf("freelook cancellation requires FREE_LOOK_ACTIVE status; current: %s", status)
		}

	case domain.RequestTypePaidUp: // BR-PM-022
		if !inStatusSet(status,
			domain.StatusActive,
			domain.StatusActiveLapse,
		) {
			return fmt.Errorf("voluntary paid-up requires ACTIVE or ACTIVE_LAPSE status; current: %s", status)
		}

	default:
		// NFR requests (BR-PM-023): all non-terminal statuses are allowed.
		// Terminal check already performed above.
	}

	return nil
}

// inStatusSet returns true if status matches any of the allowed values.
func inStatusSet(status string, allowed ...string) bool {
	for _, a := range allowed {
		if status == a {
			return true
		}
	}
	return false
}

// submitRequest executes the standard 9-step submission flow for all policy request types.
// [FR-PM-001, FR-PM-006]
func (h *PolicyRequestHandler) submitRequest(
	ctx context.Context,
	policyNumber string,
	requestType string,
	requestCategory string,
	isFinancial bool,
	signalName string,
	sourceChannel string,
	submittedBy *int64,
	idempotencyKey string,
	requestPayload json.RawMessage,
) (*resp.RequestAcceptedResponse, error) {
	// Step 2: Idempotency check — return original 202 if duplicate key found.
	// Financial requests require idempotency key to prevent duplicate processing.

	
	if isFinancial && idempotencyKey == "" {
		return nil, newHTTPErr(http.StatusBadRequest,
			"X-Idempotency-Key header required for financial requests", nil)
	}
	
	if idempotencyKey != "" {
		existing, err := h.srRepo.CheckIdempotencyKey(ctx, idempotencyKey)
		if err != nil {
			log.Error(ctx, "CheckIdempotencyKey policy=%s type=%s: %v", policyNumber, requestType, err)
			return nil, err
		}
		if existing != nil {
			// Validate request type matches for idempotency
			if existing.RequestType != requestType {
				return nil, newHTTPErr(http.StatusBadRequest,
					fmt.Sprintf("idempotency key reused for different request type: existing=%s, new=%s",
						existing.RequestType, requestType), nil)
			}
			log.Info(ctx, "idempotent duplicate requestID=%d policy=%s type=%s", existing.RequestID, policyNumber, requestType)
			return &resp.RequestAcceptedResponse{
				StatusCodeAndMessage: port.AcceptedSuccess,
				Data:                 resp.NewRequestAcceptedData(*existing),
			}, nil
		}
	}

	// Step 3: Lookup policy by number.
	policy, err := h.policyRepo.GetPolicyByNumber(ctx, policyNumber)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, newHTTPErr(http.StatusNotFound,
				fmt.Sprintf("policy %s not found", policyNumber), err)
		}
		log.Error(ctx, "GetPolicyByNumber %s: %v", policyNumber, err)
		return nil, err
	}

	// Step 4: State gate fast-fail pre-check. [BR-PM-011..BR-PM-023, §9.5.2]
	if sgErr := checkStateGate(policy, requestType); sgErr != nil {
		return nil, newHTTPErr(http.StatusUnprocessableEntity, sgErr.Error(), sgErr)
	}

	// Step 5: Financial lock check (financial requests only). [BR-PM-030]
	if isFinancial {
		lock, err := h.policyRepo.CheckFinancialLock(ctx, policy.PolicyID)
		if err != nil {
			log.Error(ctx, "CheckFinancialLock policyID=%d: %v", policy.PolicyID, err)
			return nil, err
		}
		if lock != nil {
			return nil, newHTTPErr(http.StatusConflict,
				fmt.Sprintf("financial lock active on policy %s (request_id=%d, type=%s)",
					policyNumber, lock.RequestID, lock.RequestType),
				nil,
			)
		}
	}

	// Step 6: INSERT service_request with status=RECEIVED.
	stateGateStatus := policy.CurrentStatus
	sr := &domain.ServiceRequest{
		PolicyID:        policy.PolicyID,
		PolicyNumber:    policyNumber,
		RequestType:     requestType,
		RequestCategory: requestCategory,
		SourceChannel:   sourceChannel,
		SubmittedBy:     submittedBy,
		StateGateStatus: &stateGateStatus,
		RequestPayload:  requestPayload,
		IdempotencyKey:  strPtrIfNonEmpty(idempotencyKey),
	}
	created, err := h.srRepo.CreateServiceRequest(ctx, sr)
	if err != nil {
		log.Error(ctx, "CreateServiceRequest policy=%s type=%s: %v", policyNumber, requestType, err)
		return nil, err
	}

	// Step 7: Send Temporal signal to plw-{policyNumber}.
	// IdempotencyKey (UUID) is passed alongside the BIGINT request_id so the workflow
	// can use it for dedup and child workflow IDs per Constraint 1. [Review-Fix-11]
	signal := policyRequestSignal{
		RequestID:       created.RequestID,
		IdempotencyKey:  idempotencyKey,
		RequestType:     requestType,
		RequestCategory: requestCategory,
		SourceChannel:   sourceChannel,
		SubmittedBy:     submittedBy,
	}
	wfID := policyWorkflowID(policyNumber)
	if sigErr := h.tc.SignalWorkflow(ctx, wfID, "", signalName, signal); sigErr != nil {
		log.Error(ctx, "SignalWorkflow %s signal=%s requestID=%d: %v",
			wfID, signalName, created.RequestID, sigErr)
		return nil, fmt.Errorf("failed to route request to policy workflow: %w", sigErr)
	}

	// Step 8: UPDATE service_request status=ROUTED.
	if updErr := h.srRepo.UpdateServiceRequestStatus(
		ctx, created.RequestID, domain.RequestStatusRouted, nil, &created.SubmittedAt,
	); updErr != nil {
		// Non-fatal — the workflow will update status via UpdateServiceRequestActivity.
		log.Warn(ctx, "UpdateServiceRequestStatus requestID=%d to ROUTED: %v (non-fatal)", created.RequestID, updErr)
	} else {
		created.Status = domain.RequestStatusRouted
	}

	// Step 9: Return 202.
	return &resp.RequestAcceptedResponse{
		StatusCodeAndMessage: port.AcceptedSuccess,
		Data:                 resp.NewRequestAcceptedData(*created),
	}, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Financial endpoint handlers (11 endpoints)
// ─────────────────────────────────────────────────────────────────────────────

// SubmitSurrenderRequest — POST /v1/policies/:policy_number/requests/surrender
// [FR-PM-001] [BR-PM-011] [BR-PM-030]
func (h *PolicyRequestHandler) SubmitSurrenderRequest(sctx *serverRoute.Context, req surrenderReq) (*resp.RequestAcceptedResponse, error) {
	payload, _ := json.Marshal(req.SubmitSurrenderRequest.Payload)
	return h.submitRequest(
		sctx.Ctx,
		req.PolicyNumber,
		domain.RequestTypeSurrender,
		domain.RequestCategoryFinancial,
		true,
		signalSurrenderRequest,
		req.SourceChannel,
		req.SubmittedBy,
		getIdempotencyKey(sctx.Ctx),
		payload,
	)
}

// SubmitLoanRequest — POST /v1/policies/:policy_number/requests/loan
// [FR-PM-001] [BR-PM-012] [BR-PM-030]
func (h *PolicyRequestHandler) SubmitLoanRequest(sctx *serverRoute.Context, req loanReq) (*resp.RequestAcceptedResponse, error) {
	payload, _ := json.Marshal(req.SubmitLoanRequest.Payload)
	return h.submitRequest(
		sctx.Ctx,
		req.PolicyNumber,
		domain.RequestTypeLoan,
		domain.RequestCategoryFinancial,
		true,
		signalLoanRequest,
		req.SourceChannel,
		req.SubmittedBy,
		getIdempotencyKey(sctx.Ctx),
		payload,
	)
}

// SubmitLoanRepayment — POST /v1/policies/:policy_number/requests/loan-repayment
// No financial lock (BR-PM-020). [FR-PM-001] [BR-PM-020]
func (h *PolicyRequestHandler) SubmitLoanRepayment(sctx *serverRoute.Context, req loanRepaymentReq) (*resp.RequestAcceptedResponse, error) {
	payload, _ := json.Marshal(req.SubmitLoanRepaymentRequest.Payload)
	return h.submitRequest(
		sctx.Ctx,
		req.PolicyNumber,
		domain.RequestTypeLoanRepayment,
		domain.RequestCategoryFinancial,
		false, // no financial lock for loan repayment
		signalLoanRepaymentRequest,
		req.SourceChannel,
		req.SubmittedBy,
		getIdempotencyKey(sctx.Ctx),
		payload,
	)
}

// SubmitRevivalRequest — POST /v1/policies/:policy_number/requests/revival
// [FR-PM-001] [BR-PM-013] [BR-PM-030]
func (h *PolicyRequestHandler) SubmitRevivalRequest(sctx *serverRoute.Context, req revivalReq) (*resp.RequestAcceptedResponse, error) {
	payload, _ := json.Marshal(req.SubmitRevivalRequest.Payload)
	return h.submitRequest(
		sctx.Ctx,
		req.PolicyNumber,
		domain.RequestTypeRevival,
		domain.RequestCategoryFinancial,
		true,
		signalRevivalRequest,
		req.SourceChannel,
		req.SubmittedBy,
		getIdempotencyKey(sctx.Ctx),
		payload,
	)
}

// SubmitDeathClaim — POST /v1/policies/:policy_number/requests/death-claim
// PREEMPTIVE: cancels active financial operation (BR-PM-031). No financial lock.
// [FR-PM-001] [BR-PM-014] [BR-PM-031] [BR-PM-112]
func (h *PolicyRequestHandler) SubmitDeathClaim(sctx *serverRoute.Context, req deathClaimReq) (*resp.RequestAcceptedResponse, error) {
	payload, _ := json.Marshal(req.SubmitDeathClaimRequest.Payload)
	return h.submitRequest(
		sctx.Ctx,
		req.PolicyNumber,
		domain.RequestTypeDeathClaim,
		domain.RequestCategoryFinancial,
		false, // no financial lock — preempts active operations (BR-PM-031)
		signalDeathClaimRequest,
		req.SourceChannel,
		req.SubmittedBy,
		getIdempotencyKey(sctx.Ctx),
		payload,
	)
}

// SubmitMaturityClaim — POST /v1/policies/:policy_number/requests/maturity-claim
// [FR-PM-001] [BR-PM-015] [BR-PM-030]
func (h *PolicyRequestHandler) SubmitMaturityClaim(sctx *serverRoute.Context, req maturityClaimReq) (*resp.RequestAcceptedResponse, error) {
	payload, _ := json.Marshal(req.SubmitMaturityClaimRequest.Payload)
	return h.submitRequest(
		sctx.Ctx,
		req.PolicyNumber,
		domain.RequestTypeMaturityClaim,
		domain.RequestCategoryFinancial,
		true,
		signalMaturityClaimRequest,
		req.SourceChannel,
		req.SubmittedBy,
		getIdempotencyKey(sctx.Ctx),
		payload,
	)
}

// SubmitSurvivalBenefit — POST /v1/policies/:policy_number/requests/survival-benefit
// [FR-PM-001] [BR-PM-016] [BR-PM-030]
func (h *PolicyRequestHandler) SubmitSurvivalBenefit(sctx *serverRoute.Context, req survivalBenefitReq) (*resp.RequestAcceptedResponse, error) {
	payload, _ := json.Marshal(req.SubmitSurvivalBenefitRequest.Payload)
	return h.submitRequest(
		sctx.Ctx,
		req.PolicyNumber,
		domain.RequestTypeSurvivalBenefit,
		domain.RequestCategoryFinancial,
		true,
		signalSurvivalBenefitRequest,
		req.SourceChannel,
		req.SubmittedBy,
		getIdempotencyKey(sctx.Ctx),
		payload,
	)
}

// SubmitCommutationRequest — POST /v1/policies/:policy_number/requests/commutation
// [FR-PM-001] [BR-PM-017] [BR-PM-030]
func (h *PolicyRequestHandler) SubmitCommutationRequest(sctx *serverRoute.Context, req commutationReq) (*resp.RequestAcceptedResponse, error) {
	payload, _ := json.Marshal(req.SubmitCommutationRequest.Payload)
	return h.submitRequest(
		sctx.Ctx,
		req.PolicyNumber,
		domain.RequestTypeCommutation,
		domain.RequestCategoryFinancial,
		true,
		signalCommutationRequest,
		req.SourceChannel,
		req.SubmittedBy,
		getIdempotencyKey(sctx.Ctx),
		payload,
	)
}

// SubmitConversionRequest — POST /v1/policies/:policy_number/requests/conversion
// [FR-PM-001] [BR-PM-018] [BR-PM-030]
func (h *PolicyRequestHandler) SubmitConversionRequest(sctx *serverRoute.Context, req conversionReq) (*resp.RequestAcceptedResponse, error) {
	payload, _ := json.Marshal(req.SubmitConversionRequest.Payload)
	return h.submitRequest(
		sctx.Ctx,
		req.PolicyNumber,
		domain.RequestTypeConversion,
		domain.RequestCategoryFinancial,
		true,
		signalConversionRequest,
		req.SourceChannel,
		req.SubmittedBy,
		getIdempotencyKey(sctx.Ctx),
		payload,
	)
}

// SubmitFreelookCancellation — POST /v1/policies/:policy_number/requests/freelook
// [FR-PM-001] [BR-PM-019] [BR-PM-030]
func (h *PolicyRequestHandler) SubmitFreelookCancellation(sctx *serverRoute.Context, req freelookReq) (*resp.RequestAcceptedResponse, error) {
	payload, _ := json.Marshal(req.SubmitFreelookRequest.Payload)
	return h.submitRequest(
		sctx.Ctx,
		req.PolicyNumber,
		domain.RequestTypeFLC,
		domain.RequestCategoryFinancial,
		true,
		signalFLCRequest,
		req.SourceChannel,
		req.SubmittedBy,
		getIdempotencyKey(sctx.Ctx),
		payload,
	)
}

// SubmitVoluntaryPaidUp — POST /v1/policies/:policy_number/requests/paid-up
// PM-internal: no downstream service; PM calculates PU value and transitions.
// [FR-PM-001] [BR-PM-022] [BR-PM-030] [BR-PM-060] [BR-PM-061]
func (h *PolicyRequestHandler) SubmitVoluntaryPaidUp(sctx *serverRoute.Context, req paidUpReq) (*resp.RequestAcceptedResponse, error) {
	payload, _ := json.Marshal(req.SubmitPaidUpRequest)
	return h.submitRequest(
		sctx.Ctx,
		req.PolicyNumber,
		domain.RequestTypePaidUp,
		domain.RequestCategoryFinancial,
		true,
		signalVoluntaryPaidUpRequest,
		req.SourceChannel,
		req.SubmittedBy,
		getIdempotencyKey(sctx.Ctx),
		payload,
	)
}

// ─────────────────────────────────────────────────────────────────────────────
// NFR endpoint handlers (6 endpoints)
// No financial lock required (BR-PM-023). All non-terminal policies allowed.
// ─────────────────────────────────────────────────────────────────────────────

// SubmitNominationChange — POST /v1/policies/:policy_number/requests/nomination-change
// [FR-PM-001] [BR-PM-023]
func (h *PolicyRequestHandler) SubmitNominationChange(sctx *serverRoute.Context, req nominationChangeReq) (*resp.RequestAcceptedResponse, error) {
	payload, _ := json.Marshal(req.SubmitNominationChangeRequest.Payload)
	return h.submitRequest(
		sctx.Ctx,
		req.PolicyNumber,
		domain.RequestTypeNominationChange,
		domain.RequestCategoryNonFinancial,
		false,
		signalNFRRequest,
		req.SourceChannel,
		req.SubmittedBy,
		getIdempotencyKey(sctx.Ctx),
		payload,
	)
}

// SubmitBillingMethodChange — POST /v1/policies/:policy_number/requests/billing-method-change
// [FR-PM-001] [BR-PM-023]
func (h *PolicyRequestHandler) SubmitBillingMethodChange(sctx *serverRoute.Context, req billingMethodChangeReq) (*resp.RequestAcceptedResponse, error) {
	payload, _ := json.Marshal(req.SubmitBillingMethodChangeRequest.Payload)
	return h.submitRequest(
		sctx.Ctx,
		req.PolicyNumber,
		domain.RequestTypeBillingMethodChange,
		domain.RequestCategoryNonFinancial,
		false,
		signalNFRRequest,
		req.SourceChannel,
		req.SubmittedBy,
		getIdempotencyKey(sctx.Ctx),
		payload,
	)
}

// SubmitAssignment — POST /v1/policies/:policy_number/requests/assignment
// [FR-PM-001] [BR-PM-023]
func (h *PolicyRequestHandler) SubmitAssignment(sctx *serverRoute.Context, req assignmentReq) (*resp.RequestAcceptedResponse, error) {
	payload, _ := json.Marshal(req.SubmitAssignmentRequest.Payload)
	return h.submitRequest(
		sctx.Ctx,
		req.PolicyNumber,
		domain.RequestTypeAssignment,
		domain.RequestCategoryNonFinancial,
		false,
		signalNFRRequest,
		req.SourceChannel,
		req.SubmittedBy,
		getIdempotencyKey(sctx.Ctx),
		payload,
	)
}

// SubmitAddressChange — POST /v1/policies/:policy_number/requests/address-change
// [FR-PM-001] [BR-PM-023]
func (h *PolicyRequestHandler) SubmitAddressChange(sctx *serverRoute.Context, req addressChangeReq) (*resp.RequestAcceptedResponse, error) {
	payload, _ := json.Marshal(req.SubmitAddressChangeRequest.Payload)
	return h.submitRequest(
		sctx.Ctx,
		req.PolicyNumber,
		domain.RequestTypeAddressChange,
		domain.RequestCategoryNonFinancial,
		false,
		signalNFRRequest,
		req.SourceChannel,
		req.SubmittedBy,
		getIdempotencyKey(sctx.Ctx),
		payload,
	)
}

// SubmitPremiumRefund — POST /v1/policies/:policy_number/requests/premium-refund
// [FR-PM-001] [BR-PM-023]
func (h *PolicyRequestHandler) SubmitPremiumRefund(sctx *serverRoute.Context, req premiumRefundReq) (*resp.RequestAcceptedResponse, error) {
	payload, _ := json.Marshal(req.SubmitPremiumRefundRequest.Payload)
	return h.submitRequest(
		sctx.Ctx,
		req.PolicyNumber,
		domain.RequestTypePremiumRefund,
		domain.RequestCategoryNonFinancial,
		false,
		signalNFRRequest,
		req.SourceChannel,
		req.SubmittedBy,
		getIdempotencyKey(sctx.Ctx),
		payload,
	)
}

// SubmitDuplicateBond — POST /v1/policies/:policy_number/requests/duplicate-bond
// [FR-PM-001] [BR-PM-023]
func (h *PolicyRequestHandler) SubmitDuplicateBond(sctx *serverRoute.Context, req duplicateBondReq) (*resp.RequestAcceptedResponse, error) {
	payload, _ := json.Marshal(req.SubmitDuplicateBondRequest)
	return h.submitRequest(
		sctx.Ctx,
		req.PolicyNumber,
		domain.RequestTypeDuplicateBond,
		domain.RequestCategoryNonFinancial,
		false,
		signalNFRRequest,
		req.SourceChannel,
		req.SubmittedBy,
		getIdempotencyKey(sctx.Ctx),
		payload,
	)
}

// ─────────────────────────────────────────────────────────────────────────────
// Admin / System Endpoint handlers (2 endpoints)
// ─────────────────────────────────────────────────────────────────────────────

// AdminVoidPolicy — POST /v1/policies/:policy_number/requests/admin-void
// Sends "admin-void" signal directly to plw-{policy_number}.
// No service_request record is created (admin action). [BR-PM-073]
func (h *PolicyRequestHandler) AdminVoidPolicy(sctx *serverRoute.Context, req adminVoidReq) (*resp.RequestAcceptedResponse, error) {
	policyNumber := req.PolicyNumber

	// Verify policy exists before signaling.
	policy, err := h.policyRepo.GetPolicyByNumber(sctx.Ctx, policyNumber)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, newHTTPErr(http.StatusNotFound,
				fmt.Sprintf("policy %s not found", policyNumber), err)
		}
		log.Error(sctx.Ctx, "AdminVoidPolicy GetPolicyByNumber %s: %v", policyNumber, err)
		return nil, err
	}

	// Reject if already in a terminal state (admin-void is idempotent for active policies).
	if domain.TerminalStatuses[policy.CurrentStatus] {
		return nil, newHTTPErr(http.StatusConflict,
			fmt.Sprintf("policy %s is already in terminal status %s", policyNumber, policy.CurrentStatus),
			nil,
		)
	}

	// Send admin-void signal to PLW workflow — no service_request created.
	idempKey := getIdempotencyKey(sctx.Ctx)
	if idempKey == "" {
		idempKey = uuid.NewString() // generate if caller omitted X-Idempotency-Key
	}
	signalPayload := adminVoidSignalPayload{
		RequestID:    idempKey,
		Reason:       req.Reason,
		AuthorizedBy: req.VoidedBy,
	}
	wfID := policyWorkflowID(policyNumber)
	if err := h.tc.SignalWorkflow(sctx.Ctx, wfID, "", signalAdminVoid, signalPayload); err != nil {
		log.Error(sctx.Ctx, "AdminVoidPolicy SignalWorkflow %s: %v", wfID, err)
		return nil, fmt.Errorf("failed to deliver admin-void signal to policy workflow: %w", err)
	}

	log.Info(sctx.Ctx, "admin-void signal sent policy=%s voidedBy=%d", policyNumber, req.VoidedBy)

	// Return 202 — the workflow handles the actual void transition.
	// A placeholder response is returned since no service_request was created.
	placeholder := domain.ServiceRequest{
		PolicyNumber:    policyNumber,
		PolicyID:        policy.PolicyID,
		RequestType:     domain.RequestTypeAdminVoid,
		RequestCategory: domain.RequestCategoryAdmin,
		Status:          domain.RequestStatusRouted,
		SourceChannel:   domain.SourceChannelSystem,
	}
	return &resp.RequestAcceptedResponse{
		StatusCodeAndMessage: port.AcceptedSuccess,
		Data:                 resp.NewRequestAcceptedData(placeholder),
	}, nil
}

// ReopenPolicy — POST /v1/policies/:policy_number/requests/reopen
// Used AFTER terminal cooling expires to restart a workflow from snapshot state.
// Uses SignalWithStart with WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE to rebirth the workflow.
// [BR-PM-090+, §9.5.1]
func (h *PolicyRequestHandler) ReopenPolicy(sctx *serverRoute.Context, req reopenReq) (*resp.RequestAcceptedResponse, error) {
	policyNumber := req.PolicyNumber

	// Load terminal state snapshot from DB (Tier-2 fallback source).
	snapshot, err := h.policyRepo.GetTerminalSnapshot(sctx.Ctx, policyNumber)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, newHTTPErr(http.StatusNotFound,
				fmt.Sprintf("no terminal snapshot found for policy %s — cannot reopen", policyNumber),
				err,
			)
		}
		log.Error(sctx.Ctx, "ReopenPolicy GetTerminalSnapshot %s: %v", policyNumber, err)
		return nil, err
	}

	// Signal to restart or signal the existing workflow.
	// SignalWithStart atomically signals an existing workflow OR starts a new one
	// with ALLOW_DUPLICATE reuse policy (rebirth after cooling completion).
	wfID := policyWorkflowID(policyNumber)
	startOptions := client.StartWorkflowOptions{
		ID:        wfID,
		TaskQueue: "policy-management-tq",
		// Allow a new execution even after the cooling period completes.
		WorkflowIDReusePolicy: enumspb.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE,
	}
	reopenKey := getIdempotencyKey(sctx.Ctx)
	if reopenKey == "" {
		reopenKey = uuid.NewString()
	}
	signalPayload := reopenSignalPayload{
		RequestID:    reopenKey,
		ReopenReason: req.Reason,
		AuthorizedBy: req.ReopenedBy,
	}
	// FinalSnapshot (JSON) is passed as the initial workflow state arg.
	// PolicyLifecycleWorkflow uses it to restore state from the snapshot.
	if _, err := h.tc.SignalWithStartWorkflow(
		sctx.Ctx,
		wfID,
		signalReopenRequest,
		signalPayload,
		startOptions,
		"PolicyLifecycleWorkflow", // registered workflow type name (Phase 5)
		snapshot.FinalSnapshot,    // initial state from terminal_state_snapshot
	); err != nil {
		log.Error(sctx.Ctx, "ReopenPolicy SignalWithStartWorkflow %s: %v", wfID, err)
		return nil, fmt.Errorf("failed to reopen policy workflow: %w", err)
	}

	log.Info(sctx.Ctx, "policy %s reopen signal sent by user %d", policyNumber, req.ReopenedBy)

	// Return 202 — the workflow handles state restoration.
	placeholder := domain.ServiceRequest{
		PolicyNumber:    policyNumber,
		PolicyID:        snapshot.PolicyID,
		RequestType:     domain.RequestTypeReopen,
		RequestCategory: domain.RequestCategoryAdmin,
		Status:          domain.RequestStatusRouted,
		SourceChannel:   domain.SourceChannelSystem,
	}
	return &resp.RequestAcceptedResponse{
		StatusCodeAndMessage: port.AcceptedSuccess,
		Data:                 resp.NewRequestAcceptedData(placeholder),
	}, nil
}
