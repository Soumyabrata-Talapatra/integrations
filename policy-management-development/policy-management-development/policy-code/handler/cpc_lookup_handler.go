package handler

// CPCLookupHandler — static enum lookup endpoints (Step 4.5)
//
// Implements 6 lookup endpoints — all return static data from domain constants:
//   - GET /lookups/request-types         — all request type codes + labels
//   - GET /lookups/source-channels       — all valid source channel values
//   - GET /lookups/disbursement-methods  — valid disbursement methods
//   - GET /lookups/nfr-types             — non-financial request type codes
//   - GET /lookups/lifecycle-states      — 23-state lifecycle catalogue
//   - GET /lookups/products              — PLI/RPLI product code catalogue
//
// All endpoints require no query parameters and are GET-only.
// Static data is derived from domain constants (no DB or Temporal calls).

import (
	serverHandler "gitlab.cept.gov.in/it-2.0-common/n-api-server/handler"
	serverRoute "gitlab.cept.gov.in/it-2.0-common/n-api-server/route"

	"policy-management/core/domain"
	"policy-management/core/port"
	resp "policy-management/handler/response"
)

// ─────────────────────────────────────────────────────────────────────────────
// CPCLookupHandler
// ─────────────────────────────────────────────────────────────────────────────

// CPCLookupHandler handles all static lookup endpoints.
// No DB or Temporal dependencies — all data is derived from domain constants.
// [FR-PM-004, FR-PM-008]
type CPCLookupHandler struct {
	*serverHandler.Base
}

// NewCPCLookupHandler constructs a CPCLookupHandler (no external dependencies).
func NewCPCLookupHandler() *CPCLookupHandler {
	base := serverHandler.New("Lookups").SetPrefix("/v1").AddPrefix("")
	return &CPCLookupHandler{Base: base}
}

// Routes registers all 6 lookup endpoints.
func (h *CPCLookupHandler) Routes() []serverRoute.Route {
	return []serverRoute.Route{
		serverRoute.GET("/lookups/request-types", h.GetRequestTypes).
			Name("Get Request Types"),
		serverRoute.GET("/lookups/source-channels", h.GetSourceChannels).
			Name("Get Source Channels"),
		serverRoute.GET("/lookups/disbursement-methods", h.GetDisbursementMethods).
			Name("Get Disbursement Methods"),
		serverRoute.GET("/lookups/nfr-types", h.GetNFRTypes).
			Name("Get NFR Types"),
		serverRoute.GET("/lookups/lifecycle-states", h.GetLifecycleStates).
			Name("Get Lifecycle States"),
		serverRoute.GET("/lookups/products", h.GetProducts).
			Name("Get Products"),
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Lookup endpoint handlers
// ─────────────────────────────────────────────────────────────────────────────

// GetRequestTypes — GET /v1/lookups/request-types
// Returns all 17 REST-exposed request type codes and their human-readable labels.
// [FR-PM-004]
func (h *CPCLookupHandler) GetRequestTypes(sctx *serverRoute.Context, req struct{}) (*resp.LookupListResponse, error) {
	items := []resp.LookupItemData{
		resp.NewLookupItemData(domain.RequestTypeSurrender, "Surrender", strPtr("Policy surrender request")),
		resp.NewLookupItemData(domain.RequestTypeLoan, "Policy Loan", strPtr("New loan against policy surrender value")),
		resp.NewLookupItemData(domain.RequestTypeLoanRepayment, "Loan Repayment", strPtr("Repayment of outstanding policy loan")),
		resp.NewLookupItemData(domain.RequestTypeRevival, "Revival", strPtr("Revival of lapsed policy")),
		resp.NewLookupItemData(domain.RequestTypeDeathClaim, "Death Claim", strPtr("Death claim settlement")),
		resp.NewLookupItemData(domain.RequestTypeMaturityClaim, "Maturity Claim", strPtr("Maturity benefit claim")),
		resp.NewLookupItemData(domain.RequestTypeSurvivalBenefit, "Survival Benefit", strPtr("Survival benefit installment claim")),
		resp.NewLookupItemData(domain.RequestTypeCommutation, "Commutation", strPtr("Commutation of pension installments")),
		resp.NewLookupItemData(domain.RequestTypeConversion, "Conversion", strPtr("Product conversion request")),
		resp.NewLookupItemData(domain.RequestTypeFLC, "Freelook Cancellation", strPtr("Cancellation within free-look period")),
		resp.NewLookupItemData(domain.RequestTypePaidUp, "Voluntary Paid-Up", strPtr("Voluntary conversion to paid-up policy")),
		resp.NewLookupItemData(domain.RequestTypeNominationChange, "Nomination Change", strPtr("Update policy nominee details")),
		resp.NewLookupItemData(domain.RequestTypeBillingMethodChange, "Billing Method Change", strPtr("Change premium collection method")),
		resp.NewLookupItemData(domain.RequestTypeAssignment, "Assignment", strPtr("Policy assignment to a third party")),
		resp.NewLookupItemData(domain.RequestTypeAddressChange, "Address Change", strPtr("Update policyholder address and personal details")),
		resp.NewLookupItemData(domain.RequestTypePremiumRefund, "Premium Refund", strPtr("Refund of erroneously collected premium")),
		resp.NewLookupItemData(domain.RequestTypeDuplicateBond, "Duplicate Bond", strPtr("Issuance of duplicate policy bond")),
	}
	return &resp.LookupListResponse{
		StatusCodeAndMessage: port.ListSuccess,
		Items:                items,
	}, nil
}

// GetSourceChannels — GET /v1/lookups/source-channels
// Returns all valid source channels for request submissions.
// [FR-PM-004]
func (h *CPCLookupHandler) GetSourceChannels(sctx *serverRoute.Context, req struct{}) (*resp.LookupListResponse, error) {
	items := []resp.LookupItemData{
		resp.NewLookupItemData("CUSTOMER_PORTAL", "Customer Portal", strPtr("Online self-service portal for policyholders")),
		resp.NewLookupItemData("CPC", "CPC Counter", strPtr("Central Processing Centre (manual/counter submissions)")),
		resp.NewLookupItemData("MOBILE_APP", "Mobile App", strPtr("PLI/RPLI mobile application")),
		resp.NewLookupItemData("AGENT_PORTAL", "Agent Portal", strPtr("Field agent-assisted submission portal")),
		resp.NewLookupItemData("BATCH", "Batch Process", strPtr("Automated system batch job")),
		resp.NewLookupItemData("SYSTEM", "System Internal", strPtr("Internal system-generated requests")),
	}
	return &resp.LookupListResponse{
		StatusCodeAndMessage: port.ListSuccess,
		Items:                items,
	}, nil
}

// GetDisbursementMethods — GET /v1/lookups/disbursement-methods
// Returns all valid disbursement methods for financial claim payouts.
// [FR-PM-004]
func (h *CPCLookupHandler) GetDisbursementMethods(sctx *serverRoute.Context, req struct{}) (*resp.LookupListResponse, error) {
	items := []resp.LookupItemData{
		resp.NewLookupItemData("NEFT", "NEFT", strPtr("National Electronic Funds Transfer — bank account credit")),
		resp.NewLookupItemData("CHEQUE", "Cheque", strPtr("Physical cheque payable to policyholder")),
		resp.NewLookupItemData("CASH", "Cash", strPtr("Cash disbursement at post office counter")),
		resp.NewLookupItemData("MONEY_ORDER", "Money Order", strPtr("Postal money order (applicable for RPLI surrender only)")),
	}
	return &resp.LookupListResponse{
		StatusCodeAndMessage: port.ListSuccess,
		Items:                items,
	}, nil
}

// GetNFRTypes — GET /v1/lookups/nfr-types
// Returns the 6 non-financial request types (no financial lock, can run concurrently).
// [FR-PM-004, BR-PM-023]
func (h *CPCLookupHandler) GetNFRTypes(sctx *serverRoute.Context, req struct{}) (*resp.LookupListResponse, error) {
	items := []resp.LookupItemData{
		resp.NewLookupItemData(domain.RequestTypeNominationChange, "Nomination Change", strPtr("Update nominee name, relationship, and share percentage")),
		resp.NewLookupItemData(domain.RequestTypeBillingMethodChange, "Billing Method Change", strPtr("Switch between CASH, PAY_RECOVERY, and ONLINE collection")),
		resp.NewLookupItemData(domain.RequestTypeAssignment, "Assignment", strPtr("Assign policy to a third party (ABSOLUTE or CONDITIONAL)")),
		resp.NewLookupItemData(domain.RequestTypeAddressChange, "Address Change", strPtr("Update policyholder registered address and name correction")),
		resp.NewLookupItemData(domain.RequestTypePremiumRefund, "Premium Refund", strPtr("Refund of erroneously collected premium installment")),
		resp.NewLookupItemData(domain.RequestTypeDuplicateBond, "Duplicate Bond", strPtr("Issuance of duplicate policy bond document")),
	}
	return &resp.LookupListResponse{
		StatusCodeAndMessage: port.ListSuccess,
		Items:                items,
	}, nil
}

// GetLifecycleStates — GET /v1/lookups/lifecycle-states
// Returns the full 23-state lifecycle catalogue with categories and terminal flags.
// [FR-PM-004]
func (h *CPCLookupHandler) GetLifecycleStates(sctx *serverRoute.Context, req struct{}) (*resp.LifecycleStatesResponse, error) {
	states := []resp.LifecycleStateData{
		// Active / in-force states
		{Code: domain.StatusFreeLookActive, ShortCode: "FLA", Category: "Active", Description: "Policy in free-look period (15/30 days post-issuance)", IsTerminal: false},
		{Code: domain.StatusActive, ShortCode: "ACT", Category: "Active", Description: "Policy in-force with premiums paid up-to-date", IsTerminal: false},
		{Code: domain.StatusAssignedToPresident, ShortCode: "ATP", Category: "Active", Description: "Policy assigned to President of India (pay-recovery active)", IsTerminal: false},

		// Lapsed states
		{Code: domain.StatusVoidLapse, ShortCode: "VL", Category: "Lapsed", Description: "Lapsed within 3 years — revival within remission period allowed", IsTerminal: false},
		{Code: domain.StatusInactiveLapse, ShortCode: "IL", Category: "Lapsed", Description: "Lapsed after 3 years — revival within extended period allowed", IsTerminal: false},
		{Code: domain.StatusActiveLapse, ShortCode: "AL", Category: "Lapsed", Description: "Lapsed but eligible for paid-up conversion (policy_life ≥ 3 yr)", IsTerminal: false},

		// Paid-up states
		{Code: domain.StatusPaidUp, ShortCode: "PU", Category: "Paid-Up", Description: "Voluntary paid-up conversion completed", IsTerminal: false},
		{Code: domain.StatusReducedPaidUp, ShortCode: "RPU", Category: "Paid-Up", Description: "Automatic reduced paid-up after extended lapse", IsTerminal: false},

		// Pending / transitional states
		{Code: domain.StatusPendingSurrender, ShortCode: "PSR", Category: "Pending", Description: "Surrender request accepted; processing in progress", IsTerminal: false},
		{Code: domain.StatusRevivalPending, ShortCode: "RVP", Category: "Pending", Description: "Revival request accepted; processing in progress", IsTerminal: false},
		{Code: domain.StatusPendingMaturity, ShortCode: "PMT", Category: "Pending", Description: "Maturity date within notification window; awaiting claim", IsTerminal: false},
		{Code: domain.StatusDeathClaimIntimated, ShortCode: "DCI", Category: "Pending", Description: "Death notified; claim verification in progress", IsTerminal: false},
		{Code: domain.StatusDeathUnderInvestigation, ShortCode: "DUI", Category: "Pending", Description: "Death claim under investigation (suspicious circumstances)", IsTerminal: false},
		{Code: domain.StatusPendingAutoSurrender, ShortCode: "PAS", Category: "Pending", Description: "Auto-surrender evaluation triggered by batch scan", IsTerminal: false},

		// Compliance / hold states
		{Code: domain.StatusSuspended, ShortCode: "SUS", Category: "Compliance", Description: "AML/fraud hold; accepts death claims only", IsTerminal: false},

		// Terminal states
		{Code: domain.StatusVoid, ShortCode: "VD", Category: "Terminal", Description: "Policy voided (admin void or FLC cancellation)", IsTerminal: true},
		{Code: domain.StatusSurrendered, ShortCode: "SRD", Category: "Terminal", Description: "Policy surrendered; SVF paid to policyholder", IsTerminal: true},
		{Code: domain.StatusTerminatedSurrender, ShortCode: "TS", Category: "Terminal", Description: "Forced surrender after extended lapse (batch-triggered)", IsTerminal: true},
		{Code: domain.StatusMatured, ShortCode: "MTD", Category: "Terminal", Description: "Policy matured; maturity benefit paid", IsTerminal: true},
		{Code: domain.StatusDeathClaimSettled, ShortCode: "DCS", Category: "Terminal", Description: "Death claim settled; SA + bonuses paid to claimant", IsTerminal: true},
		{Code: domain.StatusFLCCancelled, ShortCode: "FLC", Category: "Terminal", Description: "Policy cancelled within free-look period; full refund issued", IsTerminal: true},
		{Code: domain.StatusCancelledDeath, ShortCode: "CD", Category: "Terminal", Description: "Policy cancelled due to death during free-look period", IsTerminal: true},
		{Code: domain.StatusConverted, ShortCode: "CNV", Category: "Terminal", Description: "Policy converted to new product; original policy closed", IsTerminal: true},
	}
	return &resp.LifecycleStatesResponse{
		StatusCodeAndMessage: port.ListSuccess,
		States:               states,
	}, nil
}

// GetProducts — GET /v1/lookups/products
// Returns the PLI/RPLI product code catalogue.
// Product metadata is static (sourced from DDL seed data). [FR-PM-004]
func (h *CPCLookupHandler) GetProducts(sctx *serverRoute.Context, req struct{}) (*resp.LookupListResponse, error) {
	items := []resp.LookupItemData{
		// PLI products
		resp.NewLookupItemData("WLA", "Whole Life Assurance (PLI)", strPtr("Conventional whole life policy with profit sharing; PLI product")),
		resp.NewLookupItemData("EA", "Endowment Assurance (PLI)", strPtr("Term endowment policy with maturity benefit; PLI product")),
		resp.NewLookupItemData("AEA", "Anticipated Endowment Assurance (PLI)", strPtr("Endowment with periodic survival benefits; PLI product")),
		resp.NewLookupItemData("JEA", "Joint Endowment Assurance (PLI)", strPtr("Joint life endowment for couples; PLI product")),
		resp.NewLookupItemData("CWP", "Children's Welfare Plan (PLI)", strPtr("Endowment for children; PLI product")),
		resp.NewLookupItemData("YRP", "Yugal Raksha (RPLI)", strPtr("Joint life rural endowment policy; RPLI product")),

		// RPLI products
		resp.NewLookupItemData("GS", "Gram Santosh (RPLI)", strPtr("Whole life rural policy; RPLI product")),
		resp.NewLookupItemData("GS_EA", "Gram Suraksha (RPLI)", strPtr("Rural endowment policy; RPLI product")),
		resp.NewLookupItemData("GS_AEA", "Gram Sumangal (RPLI)", strPtr("Rural anticipated endowment with survival benefits; RPLI product")),
		resp.NewLookupItemData("BAL_JEEVAN", "Bal Jeevan Bima (RPLI)", strPtr("Children's rural policy; RPLI product")),
	}
	return &resp.LookupListResponse{
		StatusCodeAndMessage: port.ListSuccess,
		Items:                items,
	}, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Package-level helpers
// ─────────────────────────────────────────────────────────────────────────────

// strPtr returns a pointer to the given string (used for optional description fields).
func strPtr(s string) *string {
	return &s
}
