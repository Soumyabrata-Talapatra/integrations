package handler

import (
	"math"
	"net/http"
	"time"

	"policy-issue-service/core/domain"
	"policy-issue-service/core/port"
	resp "policy-issue-service/handler/response"
	repo "policy-issue-service/repo/postgres"

	"github.com/jackc/pgx/v5"
	log "gitlab.cept.gov.in/it-2.0-common/api-log"
	serverHandler "gitlab.cept.gov.in/it-2.0-common/n-api-server/handler"
	serverRoute "gitlab.cept.gov.in/it-2.0-common/n-api-server/route"
)

// PolicyHandler handles policy lifecycle HTTP endpoints
type PolicyHandler struct {
	*serverHandler.Base
	proposalRepo *repo.ProposalRepository
}

// NewPolicyHandler creates a new PolicyHandler instance
func NewPolicyHandler(proposalRepo *repo.ProposalRepository) *PolicyHandler {
	base := serverHandler.New("Policies").SetPrefix("/v1").AddPrefix("")
	return &PolicyHandler{
		Base:         base,
		proposalRepo: proposalRepo,
	}
}

// Routes returns the routes for the PolicyHandler
func (h *PolicyHandler) Routes() []serverRoute.Route {
	return []serverRoute.Route{
		serverRoute.GET("/policies/:policy_id", h.GetPolicy).Name("Get Policy"),
		serverRoute.POST("/policies/:policy_id/flc-cancel", h.CancelPolicyFLC).Name("Cancel Policy FLC"),
		serverRoute.GET("/policies/:policy_id/flc-status", h.GetFLCStatus).Name("Get FLC Status"),
	}
}

// GetPolicy retrieves policy details
// [POL-API-023] Get Policy
func (h *PolicyHandler) GetPolicy(sctx *serverRoute.Context, req PolicyIDUri) (*resp.PolicyDetailResponse, error) {
	proposal, err := h.proposalRepo.GetProposalByID(sctx.Ctx, req.PolicyID)
	if err != nil {
		return nil, err
	}

	return &resp.PolicyDetailResponse{
		StatusCodeAndMessage: port.StatusCodeAndMessage{
			StatusCode: http.StatusOK,
			Message:    "Policy details retrieved successfully",
		},
		ProposalID:     proposal.ProposalID,
		ProposalNumber: proposal.ProposalNumber,
		Status:         string(proposal.Status),
		CustomerID:     *proposal.CustomerID,
		SumAssured:     proposal.SumAssured,
	}, nil
}

// PolicyIDUri for policy ID in URI
type PolicyIDUri struct {
	PolicyID int64 `uri:"policy_id" validate:"required"`
}

// FLCCancelRequest for FLC cancellation
// [POL-API-022] Cancel Policy FLC
type FLCCancelRequest struct {
	PolicyID    int64  `uri:"policy_id" validate:"required"`
	Reason      string `json:"reason" validate:"required"`
	RequestedBy string `json:"requested_by" validate:"required"`
	RequestDate string `json:"request_date" validate:"required"`
}

// CancelPolicyFLC cancels policy during free look period
// [POL-API-022] Cancel Policy FLC
// [BR-POL-009] FLC Refund Calculation
// [BR-POL-021] Free Look Period Duration
// [BR-POL-028] FLC Start Date Determination
func (h *PolicyHandler) CancelPolicyFLC(sctx *serverRoute.Context, req FLCCancelRequest) (*resp.FLCCancelResponse, error) {
	// Step 1: Get proposal and validate status
	proposal, err := h.proposalRepo.GetProposalByID(sctx.Ctx, req.PolicyID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return &resp.FLCCancelResponse{
				StatusCodeAndMessage: port.StatusCodeAndMessage{
					StatusCode: http.StatusNotFound,
					Message:    "Policy not found",
				},
			}, nil
		}
		return nil, err
	}

	// Validate that proposal is in FREE_LOOK_ACTIVE status (state machine check)
	if proposal.Status != domain.ProposalStatusFreeLookActive {
		return &resp.FLCCancelResponse{
			StatusCodeAndMessage: port.StatusCodeAndMessage{
				StatusCode: http.StatusBadRequest,
				Message:    "Policy must be in FREE_LOOK_ACTIVE status for FLC cancellation. Current status: " + string(proposal.Status),
			},
		}, nil
	}

	// Step 2: Get issuance data for FLC dates
	issuance, err := h.proposalRepo.GetIssuanceByProposalID(sctx.Ctx, req.PolicyID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return &resp.FLCCancelResponse{
				StatusCodeAndMessage: port.StatusCodeAndMessage{
					StatusCode: http.StatusBadRequest,
					Message:    "No issuance data found for this policy",
				},
			}, nil
		}
		log.Error(sctx.Ctx, "Failed to get issuance data", "policyID", req.PolicyID, "error", err)
		return nil, err
	}

	// Step 3: Determine FLC window using free_look_config
	// [BR-POL-028] FLC Start Date Determination
	flcStartDate, flcEndDate, periodDays, err := h.computeFLCWindow(sctx, proposal, issuance)
	if err != nil {
		log.Error(sctx.Ctx, "Failed to compute FLC window", "policyID", req.PolicyID, "error", err)
		return nil, err
	}

	// Step 4: Validate request date is within FLC window
	now := time.Now()
	if now.After(flcEndDate) {
		daysExpired := int(now.Sub(flcEndDate).Hours() / 24)
		return &resp.FLCCancelResponse{
			StatusCodeAndMessage: port.StatusCodeAndMessage{
				StatusCode: http.StatusBadRequest,
				Message:    "Free look period has expired (ERR-POL-053)",
			},
			PolicyID:           req.PolicyID,
			CancellationStatus: "EXPIRED",
			FLCPeriod: &resp.FLCPeriodInfo{
				StartDate:     flcStartDate.Format("2006-01-02"),
				EndDate:       flcEndDate.Format("2006-01-02"),
				DaysRemaining: -daysExpired,
				PeriodDays:    periodDays,
			},
		}, nil
	}

	// Step 5: Calculate refund
	// [BR-POL-009] Refund = Premium Paid - Proportionate Risk - Stamp Duty - Medical Fee
	premiumPaid, err := h.proposalRepo.GetFirstPremiumAmount(sctx.Ctx, req.PolicyID)
	if err != nil {
		log.Error(sctx.Ctx, "Failed to get premium amount", "policyID", req.PolicyID, "error", err)
		premiumPaid = 0
	}

	refundDetails := h.calculateFLCRefund(premiumPaid, flcStartDate, now, proposal)

	// Step 6: Persist FLC cancellation with refund
	if err := h.proposalRepo.UpdateFLCCancellation(sctx.Ctx, req.PolicyID, req.Reason, refundDetails.RefundAmount); err != nil {
		log.Error(sctx.Ctx, "Failed to update FLC cancellation", "policyID", req.PolicyID, "error", err)
		return nil, err
	}

	// Step 7: Transition proposal status FREE_LOOK_ACTIVE → FLC_CANCELLED
	if err := h.proposalRepo.UpdateProposalStatus(sctx.Ctx, req.PolicyID,
		domain.ProposalStatusFLCCancelled, "FLC cancellation: "+req.Reason, proposal.CreatedBy); err != nil {
		log.Error(sctx.Ctx, "Failed to update proposal status", "policyID", req.PolicyID, "error", err)
		return nil, err
	}

	daysRemaining := int(flcEndDate.Sub(now).Hours() / 24)

	return &resp.FLCCancelResponse{
		StatusCodeAndMessage: port.StatusCodeAndMessage{
			StatusCode: http.StatusOK,
			Message:    "Policy FLC cancellation processed successfully",
		},
		PolicyID:           req.PolicyID,
		CancellationStatus: "FLC_CANCELLED",
		RefundDetails:      refundDetails,
		FLCPeriod: &resp.FLCPeriodInfo{
			StartDate:     flcStartDate.Format("2006-01-02"),
			EndDate:       flcEndDate.Format("2006-01-02"),
			DaysRemaining: daysRemaining,
			PeriodDays:    periodDays,
		},
	}, nil
}

// computeFLCWindow determines the FLC start/end dates based on config
// [BR-POL-021] Free Look Period Duration (15-30 days based on channel)
// [BR-POL-028] FLC Start Date Determination (dispatch, delivery, or email date)
func (h *PolicyHandler) computeFLCWindow(sctx *serverRoute.Context, proposal *domain.Proposal, issuance *repo.IssuanceData) (time.Time, time.Time, int, error) {
	// If dates are already set in issuance record, use them
	if issuance.FLCStartDate != nil && issuance.FLCEndDate != nil {
		periodDays := int(issuance.FLCEndDate.Sub(*issuance.FLCStartDate).Hours() / 24)
		return *issuance.FLCStartDate, *issuance.FLCEndDate, periodDays, nil
	}

	// Get FLC config for this channel
	channel := string(proposal.Channel)
	productType := string(proposal.PolicyType)
	flcConfig, err := h.proposalRepo.GetFLCConfig(sctx.Ctx, channel, &productType)
	if err != nil {
		// Fallback: 15 days from dispatch date (IRDAI standard)
		log.Warn(sctx.Ctx, "FLC config not found, using default 15 days from dispatch", "channel", channel)
		startDate := time.Now()
		if issuance.DispatchDate != nil {
			startDate = *issuance.DispatchDate
		}
		endDate := startDate.AddDate(0, 0, 15)
		return startDate, endDate, 15, nil
	}

	// Determine start date based on start_date_rule
	var startDate time.Time
	switch flcConfig.StartDateRule {
	case "DELIVERY_DATE":
		if issuance.DeliveryDate != nil {
			startDate = *issuance.DeliveryDate
		} else if issuance.DispatchDate != nil {
			startDate = *issuance.DispatchDate
		}
	case "ISSUE_DATE":
		if issuance.PolicyIssueDate != nil {
			startDate = *issuance.PolicyIssueDate
		}
	default: // "DISPATCH_DATE" (default per DDL)
		if issuance.DispatchDate != nil {
			startDate = *issuance.DispatchDate
		}
	}

	if startDate.IsZero() {
		// Fallback to issuance date if no dispatch/delivery date available
		if issuance.PolicyIssueDate != nil {
			startDate = *issuance.PolicyIssueDate
		} else {
			startDate = time.Now()
		}
	}

	endDate := startDate.AddDate(0, 0, flcConfig.PeriodDays)
	return startDate, endDate, flcConfig.PeriodDays, nil
}

// calculateFLCRefund computes the refund per BR-POL-009
// Refund = Premium Paid - Proportionate Risk - Stamp Duty - Medical Fee
func (h *PolicyHandler) calculateFLCRefund(premiumPaid float64, flcStartDate, cancelDate time.Time, proposal *domain.Proposal) *resp.FLCRefundDetails {
	// Calculate proportionate risk premium (days of coverage)
	daysCovered := int(cancelDate.Sub(flcStartDate).Hours()/24) + 1
	if daysCovered < 1 {
		daysCovered = 1
	}

	// Annual premium basis for proportionate calculation
	annualPremium := premiumPaid // simplified: assume premium_paid is the period premium
	dailyRate := annualPremium / 365.0
	proportionateRisk := math.Round(dailyRate*float64(daysCovered)*100) / 100

	// Stamp duty: typically ₹50 for most policies (configurable in future)
	stampDuty := 50.0

	// Medical fee: deduct if medical examination was performed
	var medicalFee float64
	if proposal.IsMedicalRequired {
		medicalFee = 500.0 // standard medical examination fee
	}

	refundAmount := premiumPaid - proportionateRisk - stampDuty - medicalFee
	if refundAmount < 0 {
		refundAmount = 0
	}
	refundAmount = math.Round(refundAmount*100) / 100

	return &resp.FLCRefundDetails{
		PremiumPaid:       premiumPaid,
		ProportionateRisk: proportionateRisk,
		StampDuty:         stampDuty,
		MedicalFee:        medicalFee,
		RefundAmount:      refundAmount,
	}
}

// GetFLCStatus retrieves free look period status
// [POL-API-024] Get FLC Status
func (h *PolicyHandler) GetFLCStatus(sctx *serverRoute.Context, req PolicyIDUri) (*resp.FLCStatusResponse, error) {
	proposal, err := h.proposalRepo.GetProposalByID(sctx.Ctx, req.PolicyID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return &resp.FLCStatusResponse{
				StatusCodeAndMessage: port.StatusCodeAndMessage{
					StatusCode: http.StatusNotFound,
					Message:    "Policy not found",
				},
			}, nil
		}
		return nil, err
	}

	// Get issuance data
	issuance, err := h.proposalRepo.GetIssuanceByProposalID(sctx.Ctx, req.PolicyID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return &resp.FLCStatusResponse{
				StatusCodeAndMessage: port.StatusCodeAndMessage{
					StatusCode: http.StatusOK,
					Message:    "FLC status retrieved",
				},
				PolicyID: req.PolicyID,
				Status:   string(proposal.Status),
				Eligible: false,
			}, nil
		}
		return nil, err
	}

	// Determine FLC status and eligibility
	flcStatus := "NOT_APPLICABLE"
	eligible := false
	var flcPeriod *resp.FLCPeriodInfo

	// if proposal.Status == domain.ProposalStatusFreeLookActive {
	// 	flcStartDate, flcEndDate, periodDays, err := h.computeFLCWindow(sctx, proposal, issuance)
	// 	if err == nil {
	// 		now := time.Now()
	// 		daysRemaining := int(flcEndDate.Sub(now).Hours() / 24)

	// 		if now.Before(flcEndDate) {
	// 			flcStatus = "ACTIVE"
	// 			eligible = true
	// 		} else {
	// 			flcStatus = "EXPIRED"
	// 			eligible = false
	// 		}

	switch proposal.Status {
case domain.ProposalStatusFreeLookActive:
		flcStartDate, flcEndDate, periodDays, err := h.computeFLCWindow(sctx, proposal, issuance)
		if err == nil {
			now := time.Now()
			daysRemaining := int(flcEndDate.Sub(now).Hours() / 24)

			if now.Before(flcEndDate) {
				flcStatus = "ACTIVE"
				eligible = true
			} else {
				flcStatus = "EXPIRED"
				eligible = false
			}

			flcPeriod = &resp.FLCPeriodInfo{
				StartDate:     flcStartDate.Format("2006-01-02"),
				EndDate:       flcEndDate.Format("2006-01-02"),
				DaysRemaining: daysRemaining,
				PeriodDays:    periodDays,
			}
		}
	case domain.ProposalStatusFLCCancelled:
		flcStatus = "CANCELLED"
	case domain.ProposalStatusActive:
		flcStatus = "EXPIRED"
	}

	return &resp.FLCStatusResponse{
		StatusCodeAndMessage: port.StatusCodeAndMessage{
			StatusCode: http.StatusOK,
			Message:    "FLC status retrieved successfully",
		},
		PolicyID:  req.PolicyID,
		Status:    flcStatus,
		FLCPeriod: flcPeriod,
		Eligible:  eligible,
	}, nil
}
