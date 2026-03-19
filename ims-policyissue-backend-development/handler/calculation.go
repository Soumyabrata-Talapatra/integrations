package handler

import (
	"fmt"
	"math"
	"net/http"
	"policy-issue-service/core/domain"
	"policy-issue-service/core/port"
	resp "policy-issue-service/handler/response"
	repo "policy-issue-service/repo/postgres"

	config "gitlab.cept.gov.in/it-2.0-common/api-config"
	log "gitlab.cept.gov.in/it-2.0-common/api-log"
	apierrors "gitlab.cept.gov.in/it-2.0-common/n-api-errors"
	serverHandler "gitlab.cept.gov.in/it-2.0-common/n-api-server/handler"
	serverRoute "gitlab.cept.gov.in/it-2.0-common/n-api-server/route"
)

// CalculationHandler handles calculation preview HTTP endpoints
// Phase 7: [CALC-POL-001] to [CALC-POL-004]
type CalculationHandler struct {
	*serverHandler.Base
	quoteRepo   *repo.QuoteRepository
	productRepo *repo.ProductRepository
	cfg         *config.Config
}

// NewCalculationHandler creates a new CalculationHandler instance
func NewCalculationHandler(quoteRepo *repo.QuoteRepository, productRepo *repo.ProductRepository, cfg *config.Config) *CalculationHandler {
	base := serverHandler.New("Calculation").SetPrefix("/v1").AddPrefix("")
	return &CalculationHandler{
		Base:        base,
		quoteRepo:   quoteRepo,
		productRepo: productRepo,
		cfg:         cfg,
	}
}

// Routes returns the routes for the CalculationHandler
func (h *CalculationHandler) Routes() []serverRoute.Route {
	return []serverRoute.Route{
		serverRoute.POST("/calculate/premium", h.CalculatePremium).Name("Calculate Premium"),
		serverRoute.POST("/calculate/maturity-value", h.CalculateMaturityValue).Name("Calculate Maturity Value"),
		serverRoute.POST("/calculate/flc-refund", h.CalculateFLCRefund).Name("Calculate FLC Refund"),
		serverRoute.POST("/calculate/gst", h.CalculateGST).Name("Calculate GST"),
		serverRoute.GET("/calculate/term-and-premium-ceasing-age/:product_code", h.GetTermAndPremiumCeasingAge).Name("Get Term and Premium Ceasing Age"),
	}
}

// CalculatePremium performs full premium calculation with base premium, GST, rebates
// [CALC-POL-001] Calculate premium breakdown
// [BR-POL-001] Base Premium = (SA / 1000) x Rate from Sankalan table
// [BR-POL-002] GST = 18% of net premium (CGST 9% + SGST 9% or IGST 18%)
// [BR-POL-003] Rebate based on frequency (Yearly: 2%, Half-Yearly: 1%)
// [BR-POL-004] RPLI Non-Standard Age Proof Loading = base × 1.05
// [BR-POL-010] Total Premium Calculation (End-to-End)
func (h *CalculationHandler) CalculatePremium(sctx *serverRoute.Context, req PremiumCalculationRequest) (*resp.PremiumCalcResponse, error) {

	// ---------------------------------
	// FETCH PRODUCT (ONLY ONCE)
	// ---------------------------------
	product, err := h.productRepo.GetProductByCode(sctx.Ctx, req.ProductCode)
	if err != nil {
		log.Error(sctx, "Invalid product code: %v", err)
		return nil, apierrors.HandleErrorWithStatusCodeAndMessage(
			apierrors.HTTPErrorBadRequest,
			"Invalid product code",
			err,
		)
	}

	// AGE CALCULATION (ANB)
	ageAtEntry, err := domain.CalculateAgeAtEntry(sctx.Ctx, req.ProductCode, req.DateOfBirth, req.SpouseDOB,
		req.DateOfCalculation, h.quoteRepo.GetJointLifeAgeAddition)
	if err != nil {
		log.Error(sctx.Ctx, "Age calculation failed: %v", err)
		return nil, apierrors.HandleErrorWithStatusCodeAndMessage(
			apierrors.HTTPErrorBadRequest,
			"Invalid age calculation",
			err,
		)
	}
	log.Info(
		sctx.Ctx, "AGE FINAL | Product=%s | AgeAtEntry=%d", req.ProductCode, ageAtEntry)
	err = domain.ValidateProductAge(req.ProductCode, ageAtEntry, req.Term)
	if err != nil {
		return nil, apierrors.HandleErrorWithStatusCodeAndMessage(
			apierrors.HTTPErrorBadRequest,
			"Age validation failed",
			err,
		)
	}
	// layout := "2006-01-02"

	// dob, err := time.Parse(layout, req.DateOfBirth)
	// if err != nil {
	// 	return nil, fmt.Errorf("invalid date_of_birth format (YYYY-MM-DD required)")
	// }

	// calcDate, err := time.Parse(layout, req.DateOfCalculation)
	// if err != nil {
	// 	return nil, fmt.Errorf("invalid date_of_calculation format (YYYY-MM-DD required)")
	// }

	// if calcDate.Before(dob) {
	// 	return nil, fmt.Errorf("date_of_calculation cannot be before date_of_birth")
	// }

	// // ---- PROPOSER ANB ----
	// years := calcDate.Year() - dob.Year()
	// if calcDate.Month() < dob.Month() ||
	// 	(calcDate.Month() == dob.Month() && calcDate.Day() < dob.Day()) {
	// 	years--
	// }

	// proposerANB := years + 1
	// ageAtEntry := proposerANB

	// // JOINT LIFE AGE CALCULATION (ONLY FOR 1005)
	// if req.ProductCode == "1005" {

	// 	if req.SpouseDOB == nil {
	// 		return nil, fmt.Errorf("spouse_dob required for product 1005")
	// 	}

	// 	spouseDob, err := time.Parse(layout, *req.SpouseDOB)
	// 	if err != nil {
	// 		return nil, fmt.Errorf("invalid spouse_dob format (YYYY-MM-DD required)")
	// 	}

	// 	if calcDate.Before(spouseDob) {
	// 		return nil, fmt.Errorf("date_of_calculation cannot be before spouse_dob")
	// 	}

	// 	spouseYears := calcDate.Year() - spouseDob.Year()
	// 	if calcDate.Month() < spouseDob.Month() ||
	// 		(calcDate.Month() == spouseDob.Month() && calcDate.Day() < spouseDob.Day()) {
	// 		spouseYears--
	// 	}

	// 	spouseANB := spouseYears + 1

	// 	// Determine lower & higher age
	// 	lowerAge := proposerANB
	// 	higherAge := spouseANB

	// 	if spouseANB < proposerANB {
	// 		lowerAge = spouseANB
	// 		higherAge = proposerANB
	// 	}

	// 	ageDiff := higherAge - lowerAge

	// 	// Fetch addition from table
	// 	ageAddition, err := h.quoteRepo.GetJointLifeAgeAddition(
	// 		sctx.Ctx,
	// 		ageDiff,
	// 	)
	// 	if err != nil {
	// 		return nil, err
	// 	}

	// 	ageAtEntry = lowerAge + ageAddition
	// }

	// DETERMINE LOOKUP FIELD
	var lookupField string
	var lookupValue int

	switch req.ProductCode {

	// TERM PRODUCTS
	case "1002", "1005", "5002", "5003":

		if req.Term <= 0 {
			return nil, apierrors.HandleErrorWithStatusCodeAndMessage(
				apierrors.HTTPErrorBadRequest,
				fmt.Sprintf("term required for product %s", req.ProductCode),
				nil,
			)
		}

		lookupField = "term"
		lookupValue = req.Term

	// PREMIUM CEASING AGE PRODUCTS
	case "1001", "1003", "1004", "1006",
		"5001", "5004", "5005", "5006":

		if req.PremiumCeasingAge <= ageAtEntry {
			return nil, apierrors.HandleErrorWithStatusCodeAndMessage(
				apierrors.HTTPErrorBadRequest,
				fmt.Sprintf("premium_ceasing_age must be greater than age_at_entry for product %s", req.ProductCode),
				nil,
			)
		}

		lookupField = "premium_ceasing_age"
		lookupValue = req.PremiumCeasingAge

	default:
		return nil, apierrors.HandleErrorWithStatusCodeAndMessage(
			apierrors.HTTPErrorBadRequest,
			fmt.Sprintf("product code not found: %s", req.ProductCode),
			nil,
		)
	}

	// FETCH PREMIUM RATE
	rate, baseSumAssd, err := h.quoteRepo.GetPremiumRate(sctx.Ctx, req.ProductCode, string(product.ProductCategory),
		ageAtEntry, req.Gender, req.Periodicity, lookupField, lookupValue)
	if err != nil {
		log.Error(sctx.Ctx, "Premium rate fetch failed: %v", err)
		return nil, err
	}

	if baseSumAssd <= 0 {
		return nil, fmt.Errorf("invalid slab configuration")
	}

	// ---------------------------------
	// SLAB VALIDATION
	// ---------------------------------
	if req.SumAssured%baseSumAssd != 0 {
		return nil, fmt.Errorf("sum assured must be multiple of %d", baseSumAssd)
	}

	units := req.SumAssured / baseSumAssd
	basePremium := float64(units) * rate
	basePremium = math.Round(basePremium)

	// ---------------------------------
	// NON-STANDARD LOADING (5% RPLI)
	// ---------------------------------
	loadedPremium := basePremium

	if req.AgeProofType == "NON_STANDARD" &&
		product.ProductType == domain.PolicyTypeRPLI {

		loadedPremium = basePremium * 1.05
		loadedPremium = math.Round(loadedPremium)
	}

	// ---------------------------------
	// YOUR REBATE LOGIC
	// ---------------------------------
	rebateAmount, err := h.quoteRepo.GetRebate(sctx.Ctx, req.ProductCode, req.SumAssured)
	if err != nil {
		log.Error(sctx.Ctx, "Rebate fetch failed: %v", err)
		return nil, err
	}

	if rebateAmount > loadedPremium {
		rebateAmount = loadedPremium
	}

	netPremium := loadedPremium - rebateAmount
	netPremium = math.Round(netPremium)

	// ---------------------------------
	// GST CALCULATION (18%)
	// ---------------------------------
	gstRate := 0.0

	providerState := req.ProviderState
	if providerState == "" {
		providerState = "MH"
	}

	var cgst, sgst, igst float64

	if req.InsuredState == providerState || req.InsuredState == "" {
		cgst = netPremium * gstRate / 2
		sgst = netPremium * gstRate / 2
	} else {
		igst = netPremium * gstRate
	}

	totalGST := cgst + sgst + igst
	totalGST = math.Round(totalGST*100) / 100

	// ---------------------------------
	// STAMP DUTY (Preview Mode = 0)
	// ---------------------------------
	stampDuty := 0.0

	// ---------------------------------
	// MEDICAL FEE
	// ---------------------------------
	medicalFee := 0.0

	if product.IsMedicalRequired(float64(req.SumAssured)) {
		medicalFee = 500.0 // make configurable later
	}

	// ---------------------------------
	// TOTAL FIRST PAYMENT
	// ---------------------------------
	totalFirstPayment := netPremium + totalGST + stampDuty + medicalFee
	totalFirstPayment = math.Round(totalFirstPayment)

	// Modal premium = installment premium
	modalPremium := netPremium

	// ---------------------------------
	// RESPONSE
	// ---------------------------------
	return &resp.PremiumCalcResponse{
		StatusCodeAndMessage: port.StatusCodeAndMessage{
			StatusCode: http.StatusOK,
			Message:    "Premium calculated successfully",
		},
		BasePremium:       basePremium,
		LoadedPremium:     loadedPremium,
		RebateAmount:      rebateAmount,
		NetPremium:        netPremium,
		CGST:              cgst,
		SGST:              sgst,
		IGST:              igst,
		TotalGST:          totalGST,
		StampDuty:         stampDuty,
		MedicalFee:        medicalFee,
		TotalFirstPayment: totalFirstPayment,
		ModalPremium:      modalPremium,
	}, nil
}

// func (h *CalculationHandler) CalculatePremium(sctx *serverRoute.Context, req PremiumCalculationRequest) (*resp.PremiumCalcResponse, error) {
// 	// Step 1: Get premium rate from Sankalan table [BR-POL-001]
// 	rate, sumAssd, err := h.quoteRepo.GetPremiumRate(
// 		sctx.Ctx,
// 		req.ProductCode,
// 		req.ProductCategory,
// 		req.AgeAtEntry,
// 		req.Gender,
// 		req.Frequency,
// 		"term",         // lookupField
// 		req.PolicyTerm, // lookupValue
// 	)
// 	if err != nil {
// 		log.Error(sctx.Ctx, "[CALC-POL-001] Error fetching premium rate: %v", err)
// 		return nil, err
// 	}

// 	// Step 1: base_premium = (SA / 1000) × Rate [BR-POL-001]
// 	// basePremium := (req.SumAssured / 1000) * rate
// 	sumAssdFloat, err := strconv.ParseFloat(strconv.Itoa(sumAssd), 64)
// 	if err != nil {
// 		return nil, fmt.Errorf("invalid sum_assd value: %v", err)
// 	}

// 	basePremium := (req.SumAssured / sumAssdFloat) * rate
// 	// Step 2: loaded_premium = IF rpli_non_standard THEN base × 1.05 ELSE base [BR-POL-004]
// 	loadedPremium := basePremium
// 	if req.AgeProofType == "NON_STANDARD" {
// 		// Fetch product to check if RPLI
// 		product, err := h.productRepo.GetProductByCode(sctx.Ctx, req.ProductCode)
// 		if err == nil && product.ProductType == domain.PolicyTypeRPLI {
// 			loadedPremium = basePremium * 1.05
// 		}
// 	}

// 	// Step 3: rebate = calculate_rebate(frequency) [BR-POL-003]
// 	rebateAmount := calculatePremiumRebate(loadedPremium, domain.PremiumFrequency(req.Frequency))

// 	// Step 4: net_premium = loaded_premium - rebate
// 	netPremium := loadedPremium - rebateAmount

// 	// Step 5: gst = calculate_gst(net_premium, insured_state, provider_state) [BR-POL-002]
// 	providerState := req.ProviderState
// 	if providerState == "" {
// 		providerState = "MH" // Default provider state: Maharashtra
// 	}

// 	var cgst, sgst, igst, totalGST float64
// 	gstRate := 0.0 // 18% GST

// 	if req.InsuredState == providerState || req.InsuredState == "" {
// 		// Intra-state: CGST + SGST
// 		cgst = netPremium * gstRate / 2
// 		sgst = netPremium * gstRate / 2
// 		igst = 0
// 	} else {
// 		// Inter-state: IGST
// 		cgst = 0
// 		sgst = 0
// 		igst = netPremium * gstRate
// 	}
// 	totalGST = cgst + sgst + igst

// 	// Step 6: stamp_duty (on issuance only - set to 0 for preview)
// 	stampDuty := 0.0

// 	// Step 7: medical_fee (if medical required - check product threshold)
// 	medicalFee := 0.0
// 	product, err := h.productRepo.GetProductByCode(sctx.Ctx, req.ProductCode)
// 	if err == nil && product.IsMedicalRequired(req.SumAssured) {
// 		medicalFee = 500.0 // Default medical examination fee
// 	}

// 	// Step 8: total_first_payment = net_premium + gst + stamp_duty + medical_fee [BR-POL-010]
// 	totalFirstPayment := netPremium + totalGST + stampDuty + medicalFee

// 	// Modal premium = premium per payment period
// 	modalPremium := netPremium

// 	return &resp.PremiumCalcResponse{
// 		StatusCodeAndMessage: port.StatusCodeAndMessage{
// 			StatusCode: http.StatusOK,
// 			Message:    "Premium calculated successfully",
// 		},
// 		BasePremium:       basePremium,
// 		LoadedPremium:     loadedPremium,
// 		RebateAmount:      rebateAmount,
// 		NetPremium:        netPremium,
// 		CGST:              cgst,
// 		SGST:              sgst,
// 		IGST:              igst,
// 		TotalGST:          totalGST,
// 		StampDuty:         stampDuty,
// 		MedicalFee:        medicalFee,
// 		TotalFirstPayment: totalFirstPayment,
// 		ModalPremium:      modalPremium,
// 	}, nil
// }

// CalculateMaturityValue estimates indicative maturity value with guaranteed and bonus components
// [CALC-POL-002] Maturity value estimation
// [FR-POL-001] Quote generation with benefit illustration
func (h *CalculationHandler) CalculateMaturityValue(sctx *serverRoute.Context, req MaturityCalculationRequest) (*resp.MaturityCalcResponse, error) {
	guaranteedSA := req.SumAssured

	// Get indicative bonus rate from config
	bonusRate := h.cfg.GetFloat64("quote.default_bonus_rate")
	if bonusRate <= 0 {
		bonusRate = 4.5 // fallback default: 4.5% per annum
	}

	// Indicative bonus = SA × (bonus_rate / 100) × policy_term
	indicativeBonus := req.SumAssured * (bonusRate / 100) * float64(req.PolicyTerm)

	// Indicative maturity value = guaranteed SA + indicative bonus
	indicativeMaturityValue := guaranteedSA + indicativeBonus

	return &resp.MaturityCalcResponse{
		StatusCodeAndMessage: port.StatusCodeAndMessage{
			StatusCode: http.StatusOK,
			Message:    "Maturity value calculated successfully",
		},
		GuaranteedSumAssured:    guaranteedSA,
		IndicativeBonus:         indicativeBonus,
		IndicativeMaturityValue: indicativeMaturityValue,
		BonusRateUsed:           bonusRate,
		Disclaimer:              "The maturity value shown is indicative only. Actual maturity value depends on bonus declared by PLI Directorate each year. Past performance is not indicative of future results.",
	}, nil
}

// CalculateFLCRefund calculates the Free Look Cancellation refund amount
// [CALC-POL-003] FLC refund preview
// [BR-POL-009] FLC Refund Calculation:
//
//	proportionate_risk_premium = (base_premium / 365) × days_of_coverage
//	flc_refund = initial_premium_paid - proportionate_risk_premium - stamp_duty - medical_fee
//	NOTE: GST is NOT refunded
func (h *CalculationHandler) CalculateFLCRefund(sctx *serverRoute.Context, req FLCRefundCalculationRequest) (*resp.FLCRefundCalcResponse, error) {
	// [BR-POL-009] Calculate proportionate risk premium
	proportionateRiskPremium := (req.BasePremium / 365.0) * float64(req.DaysOfCoverage)

	// Stamp duty deduction
	stampDutyDeduction := req.StampDuty

	// Medical fee deduction
	medicalFeeDeduction := req.MedicalFee

	// Total deductions
	totalDeductions := proportionateRiskPremium + stampDutyDeduction + medicalFeeDeduction

	// Net refund amount
	netRefundAmount := req.PremiumPaid - totalDeductions
	if netRefundAmount < 0 {
		netRefundAmount = 0
	}

	return &resp.FLCRefundCalcResponse{
		StatusCodeAndMessage: port.StatusCodeAndMessage{
			StatusCode: http.StatusOK,
			Message:    "FLC refund calculated successfully",
		},
		PremiumPaid:              req.PremiumPaid,
		ProportionateRiskPremium: proportionateRiskPremium,
		StampDutyDeduction:       stampDutyDeduction,
		MedicalFeeDeduction:      medicalFeeDeduction,
		TotalDeductions:          totalDeductions,
		NetRefundAmount:          netRefundAmount,
	}, nil
}

// CalculateGST calculates GST breakdown based on state of insured vs provider
// [CALC-POL-004] GST breakdown
// [BR-POL-002] GST Calculation:
//
//	Intra-state (same state): CGST 9% + SGST 9%
//	Inter-state (different state): IGST 18%
//	Union Territory: CGST 9% + UTGST 9% (treated same as SGST for simplicity)
func (h *CalculationHandler) CalculateGST(sctx *serverRoute.Context, req GSTCalculationRequest) (*resp.GSTCalcResponse, error) {
	providerState := req.ProviderState
	if providerState == "" {
		providerState = "MH" // Default: Maharashtra (PLI Directorate HQ)
	}

	gstRate := 0.0 // 18% GST rate

	var cgst, sgst, igst, totalGST float64
	var gstType string

	if req.InsuredState == providerState {
		// Intra-state: CGST + SGST (each 9%)
		gstType = "INTRASTATE"
		cgst = req.BaseAmount * 0.00
		sgst = req.BaseAmount * 0.00
		igst = 0
	} else {
		// Inter-state: IGST (18%)
		gstType = "INTERSTATE"
		cgst = 0
		sgst = 0
		igst = req.BaseAmount * 0.18
	}
	totalGST = cgst + sgst + igst

	return &resp.GSTCalcResponse{
		StatusCodeAndMessage: port.StatusCodeAndMessage{
			StatusCode: http.StatusOK,
			Message:    "GST calculated successfully",
		},
		BaseAmount: req.BaseAmount,
		GSTRate:    gstRate,
		CGST:       cgst,
		SGST:       sgst,
		IGST:       igst,
		TotalGST:   totalGST,
		GSTType:    gstType,
	}, nil
}

// calculatePremiumRebate calculates rebate based on payment frequency
// [BR-POL-003] Rebate Calculation
func calculatePremiumRebate(premium float64, frequency domain.PremiumFrequency) float64 {
	switch frequency {
	case domain.FrequencyYearly:
		return premium * 0.02 // 2% rebate for yearly
	case domain.FrequencyHalfYearly:
		return premium * 0.01 // 1% rebate for half-yearly
	default:
		return 0 // No rebate for monthly/quarterly
	}
}

func (h *CalculationHandler) GetTermAndPremiumCeasingAge(
	sctx *serverRoute.Context,
	req GetTermOrPremiumCeasingAgeRequest,
) (*resp.GetTermAndPremiumCeasingAgeResponse, error) {

	ctx := sctx.Ctx

	// Calculate age_at_entry
	ageAtEntry, err := domain.CalculateAgeAtEntry(ctx, req.ProductCode,
		req.DateOfBirth, req.SpouseDOB, req.DateOfCalculation,
		h.quoteRepo.GetJointLifeAgeAddition)
	if err != nil {
		log.Error(ctx, "Error calculating age at entry: %v", err)
		return nil, apierrors.HandleErrorWithStatusCodeAndMessage(
			apierrors.HTTPErrorBadRequest,
			"Invalid age calculation",
			err,
		)
	}
	err = domain.ValidateProductAge(req.ProductCode, ageAtEntry, 0)
	if err != nil {

		log.Error(ctx, "Age validation failed: %v", err)

		return nil, apierrors.HandleErrorWithStatusCodeAndMessage(
			apierrors.HTTPErrorBadRequest,
			"Age validation failed",
			err,
		)
	}

	// Fetch data from repo
	rows, err := h.productRepo.GetTermAndPremiumCeasingAge(ctx, req.ProductCode, ageAtEntry)
	if err != nil {
		log.Error(ctx, "Error fetching term/premium ceasing age: %v", err)
		return nil, apierrors.HandleErrorWithStatusCodeAndMessage(
			apierrors.HTTPErrorServerError,
			"Failed to fetch data",
			err,
		)
	}

	if len(rows) == 0 {
		return nil, apierrors.HandleErrorWithStatusCodeAndMessage(
			apierrors.HTTPErrorNotFound,
			"No term or premium ceasing age found",
			nil,
		)
	}

	var terms []int
	var premiumCeasingAges []int
	var periodicities []string

	termMap := make(map[int]struct{})
	premiumMap := make(map[int]struct{})
	periodicityMap := make(map[string]struct{})

	for _, r := range rows {

		if r.Term != nil {
			if _, exists := termMap[*r.Term]; !exists {
				termMap[*r.Term] = struct{}{}
				terms = append(terms, *r.Term)
			}
		}

		if r.PremiumCeasingAge != nil {
			if _, exists := premiumMap[*r.PremiumCeasingAge]; !exists {
				premiumMap[*r.PremiumCeasingAge] = struct{}{}
				premiumCeasingAges = append(premiumCeasingAges, *r.PremiumCeasingAge)
			}
		}
		if r.Periodicity != nil {
			if _, exists := periodicityMap[*r.Periodicity]; !exists {
				periodicityMap[*r.Periodicity] = struct{}{}
				periodicities = append(periodicities, *r.Periodicity)
			}
		}
	}

	return &resp.GetTermAndPremiumCeasingAgeResponse{
		StatusCodeAndMessage: port.StatusCodeAndMessage{
			StatusCode: http.StatusOK,
			Message:    "Data fetched successfully",
		},
		ProductCode:        req.ProductCode,
		AgeAtEntry:         ageAtEntry,
		Periodicity:        periodicities,
		Terms:              terms,
		PremiumCeasingAges: premiumCeasingAges,
	}, nil
}
