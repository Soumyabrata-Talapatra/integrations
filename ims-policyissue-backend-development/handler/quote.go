package handler

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
	"time"

	"policy-issue-service/core/domain"
	"policy-issue-service/core/port"
	resp "policy-issue-service/handler/response"
	repo "policy-issue-service/repo/postgres"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	config "gitlab.cept.gov.in/it-2.0-common/api-config"
	apierrors "gitlab.cept.gov.in/it-2.0-common/n-api-errors"
	log "gitlab.cept.gov.in/it-2.0-common/n-api-log"
	serverHandler "gitlab.cept.gov.in/it-2.0-common/n-api-server/handler"
	serverRoute "gitlab.cept.gov.in/it-2.0-common/n-api-server/route"
)

// QuoteHandler handles quote-related HTTP endpoints
type QuoteHandler struct {
	*serverHandler.Base
	quoteRepo    *repo.QuoteRepository
	productRepo  *repo.ProductRepository
	proposalRepo *repo.ProposalRepository
	cfg          *config.Config
}

// NewQuoteHandler creates a new QuoteHandler instance
func NewQuoteHandler(quoteRepo *repo.QuoteRepository, productRepo *repo.ProductRepository, proposalRepo *repo.ProposalRepository, cfg *config.Config) *QuoteHandler {
	base := serverHandler.New("Quotes").SetPrefix("/v1").AddPrefix("")
	return &QuoteHandler{Base: base, quoteRepo: quoteRepo, productRepo: productRepo, proposalRepo: proposalRepo, cfg: cfg}
}

// Routes returns the routes for the QuoteHandler
func (h *QuoteHandler) Routes() []serverRoute.Route {
	return []serverRoute.Route{
		serverRoute.GET("/products", h.GetProducts).Name("Get Products"),
		serverRoute.POST("/quotes/calculate/:product_code", h.CalculateQuote).Name("Calculate Quote"),
		serverRoute.POST("/quotes", h.CreateQuote).Name("Create Quote"),
		serverRoute.GET("/quotes/:quote_id", h.GetQuoteByID).Name("Get Quote by ID"), // Optional: for retrieving quote details
		serverRoute.POST("/quotes/:quote_ref_number/convert-to-proposal", h.ConvertQuoteToProposal).Name("Convert Quote to Proposal"),
		serverRoute.POST("/quotes/generate-quote/:product_code", h.GenerateQuote).Name("Generate Quote with PDF"),
		serverRoute.GET("/quotes/ref/:quote_ref_number", h.GetQuoteByNumber).Name("Resolve Quote Number"),
	}
}

// GetProducts retrieves available products
// [POL-API-001] Get Products
// [BR-POL-011] [BR-POL-012] Product Eligibility
func (h *QuoteHandler) GetProducts(sctx *serverRoute.Context, req GetProductsRequest) (*resp.ProductListResponse, error) {
	products, err := h.quoteRepo.GetProducts(sctx.Ctx, req.PolicyType, req.IsActive)
	if err != nil {
		log.Error(sctx.Ctx, "Error fetching products: %v", err)
		return nil, apierrors.HandleErrorWithStatusCodeAndMessage(
			apierrors.HTTPErrorServerError,
			"Failed to fetch products",
			err,
		)
	}
	// If no products found
	if len(products) == 0 {
		return nil, apierrors.HandleErrorWithStatusCodeAndMessage(
			apierrors.HTTPErrorNotFound,
			"No products found",
			nil,
		)
	}
	// Map domain to response
	productResponses := make([]resp.ProductResponse, len(products))
	for i, p := range products {
		productResponses[i] = resp.ProductResponse{
			ProductCode:              p.ProductCode,
			ProductName:              p.ProductName,
			ProductType:              string(p.ProductType),
			ProductCategory:          string(p.ProductCategory),
			MinSumAssured:            p.MinSumAssured,
			MaxSumAssured:            p.MaxSumAssured,
			MinEntryAge:              p.MinEntryAge,
			MaxEntryAge:              p.MaxEntryAge,
			MaxMaturityAge:           p.MaxMaturityAge,
			MinTerm:                  p.MinTerm,
			PremiumCeasingAgeOptions: []int(p.PremiumCeasingAgeOptions),
			AvailableFrequencies:     []string(p.AvailableFrequencies),
			Description:              p.Description,
		}
	}

	return &resp.ProductListResponse{
		StatusCodeAndMessage: port.StatusCodeAndMessage{
			StatusCode: http.StatusOK,
			Message:    "Products retrieved successfully",
		},
		Products: productResponses,
	}, nil
}

// CalculateQuote calculates premium quote
// [POL-API-002] Calculate Quote
func (h *QuoteHandler) CalculateQuote(sctx *serverRoute.Context,
	req QuoteCalculateRequest) (*resp.QuoteCalculateResponse, error) {

	product, err := h.productRepo.GetProductByCode(sctx.Ctx, req.ProductCode)
	if err != nil {
		log.Error(sctx.Ctx, "Error fetching product: %v", err)
		return nil, err
	}

	ageAtEntry, err := domain.CalculateAgeAtEntry(sctx.Ctx, req.ProductCode, req.DateOfBirth, req.SpouseDOB,
		req.DateOfCalculation,
		h.quoteRepo.GetJointLifeAgeAddition)
	if err != nil {
		log.Error(sctx.Ctx, "Age calculation failed: %v", err)

		return nil, apierrors.HandleErrorWithStatusCodeAndMessage(
			apierrors.HTTPErrorBadRequest,
			err.Error(),
			err,
		)
	}
	log.Info(sctx.Ctx, "AGE FINAL | Product=%s | AgeAtEntry=%d", req.ProductCode, ageAtEntry)

	err = domain.ValidateProductAge(req.ProductCode, ageAtEntry, req.Term)
	if err != nil {
		return nil, apierrors.HandleErrorWithStatusCodeAndMessage(
			apierrors.HTTPErrorBadRequest,
			err.Error(),
			err,
		)
	}

	// baseEligibility := domain.EligibilityResult{
	// 	IsEligible:  true,
	// 	AgeAtEntry:  ageAtEntry,
	// 	MaturityAge: ageAtEntry + req.Term,
	// }

	// if !product.IsEligibleAge(ageAtEntry) {
	// 	baseEligibility.IsEligible = false
	// 	baseEligibility.RejectReason = "Age not eligible for this product"
	// }

	// if !product.IsEligibleSA(float64(req.SumAssured)) {
	// 	baseEligibility.IsEligible = false
	// 	baseEligibility.RejectReason = fmt.Sprintf(
	// 		"Sum assured outside product limits: %d",
	// 		req.SumAssured,
	// 	)
	// }

	// -----------------------------
	// PREMIUM RATE FETCH
	// -----------------------------
	var lookupField string
	var lookupValue int
	var effectiveTerm int
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
		effectiveTerm = req.Term
		lookupField = "term"
		lookupValue = req.Term

	// PREMIUM CEASING AGE PRODUCTS
	case "1001", "1003", "1004", "1006",
		"5001", "5004", "5005", "5006":

		if req.PremiumCeasingAge <= ageAtEntry {
			return nil, fmt.Errorf(
				"premium_ceasing_age must be greater than age_at_entry for product %s",
				req.ProductCode,
			)
		}
		effectiveTerm = req.PremiumCeasingAge - ageAtEntry
		lookupField = "premium_ceasing_age"
		lookupValue = req.PremiumCeasingAge

	default:
		return nil, fmt.Errorf("invalid product code: %s", req.ProductCode)
	}

	// ELIGIBILITY CHECK (BASE)
	baseEligibility := domain.EligibilityResult{
		IsEligible:  true,
		AgeAtEntry:  ageAtEntry,
		MaturityAge: ageAtEntry + effectiveTerm,
	}

	if !product.IsEligibleAge(ageAtEntry) {
		baseEligibility.IsEligible = false
		baseEligibility.RejectReason = "Age not eligible for this product"
	}

	if !product.IsEligibleSA(float64(req.SumAssured)) {
		baseEligibility.IsEligible = false
		baseEligibility.RejectReason = fmt.Sprintf(
			"Sum assured outside product limits: %d",
			req.SumAssured,
		)
	}

	rate, baseSumAssd, err := h.quoteRepo.GetPremiumRate(sctx.Ctx, req.ProductCode, string(product.ProductCategory),
		ageAtEntry, req.Gender, req.Periodicity, lookupField, lookupValue)
	if err != nil {
		return nil, err
	}

	if baseSumAssd <= 0 {
		return nil, fmt.Errorf("invalid slab configuration")
	}

	//Bonus rate from config
	indicativeBonusRate := h.cfg.GetFloat64("quote.default_bonus_rate")
	if indicativeBonusRate <= 0 {
		indicativeBonusRate = 4.5
	}

	// PROJECTIONS LOOP (5 SEQUENCE)

	var calculations []resp.QuoteCalculationItem

	displayIncrement := 100000          // 1 lakh step
	calculationSlab := int(baseSumAssd) // 5000 (for validation)

	minSA := int(product.MinSumAssured)

	maxSA := 0
	if product.MaxSumAssured != nil {
		maxSA = int(*product.MaxSumAssured)
	}

	for i := 0; i < 5; i++ {

		projectedSA := req.SumAssured + (i * displayIncrement)

		// Do not exceed max sum assured
		if maxSA > 0 && projectedSA > maxSA {
			break
		}

		// Ensure above min
		if projectedSA < minSA {
			continue
		}

		// validate slab
		if projectedSA%calculationSlab != 0 {
			continue
		}

		// process projectedSA
		units := projectedSA / int(baseSumAssd)
		basePremium := float64(units) * rate
		basePremium = math.Round(basePremium*100) / 100

		// rebate := calculateRebate(basePremium, domain.PremiumFrequency(req.Periodicity))
		// netPremium := basePremium - rebate
		// Calculate Large SA rebate
		// --- Rebate (same as CalculatePremium)
		rebate, err := h.quoteRepo.GetRebate(sctx.Ctx, req.ProductCode, projectedSA)
		if err != nil {
			return nil, err
		}

		// Safety: rebate should not exceed premium
		if rebate > basePremium {
			rebate = basePremium
		}

		netPremium := basePremium - rebate
		gstRate := 0.0
		gstAmount := netPremium * gstRate
		totalPayable := netPremium + gstAmount

		maturityWithBonus :=
			float64(projectedSA) *
				(1 + (indicativeBonusRate/100)*float64(req.Term))

		calculations = append(calculations, resp.QuoteCalculationItem{
			SumAssured:  int64(projectedSA),
			Eligibility: resp.MapDomainToEligibility(baseEligibility),
			PremiumBreakdown: resp.PremiumBreakdownResponse{
				BasePremium:  basePremium,
				Rebate:       rebate,
				NetPremium:   netPremium,
				CGST:         gstAmount / 2,
				SGST:         gstAmount / 2,
				IGST:         0,
				TotalGST:     gstAmount,
				TotalPayable: totalPayable,
			},
			BenefitIllustration: resp.BenefitIllustrationResponse{
				MaturityValueGuaranteed: float64(projectedSA),
				MaturityValueWithBonus:  maturityWithBonus,
				IndicativeBonusRate:     indicativeBonusRate,
				DeathBenefit:            float64(projectedSA),
			},
		})
	}

	// -----------------------------
	// GENERATE CALCULATION ID
	// -----------------------------
	calculationID := uuid.New().String()

	// -----------------------------
	// FINAL RESPONSE
	// -----------------------------
	return &resp.QuoteCalculateResponse{
		StatusCodeAndMessage: port.StatusCodeAndMessage{
			StatusCode: http.StatusOK,
			Message:    "Premium calculated successfully",
		},
		CalculationID: calculationID,
		CalculationBasis: resp.CalculationBasisResponse{
			PremiumTable: "Sankalan",
			SumAssd:      float64(baseSumAssd),
			GSTRate:      0,
		},
		Calculations: calculations,
		WorkflowState: port.WorkflowStateResponse{
			CurrentStep: "CALCULATE_QUOTE",
			NextStep:    "CREATE_QUOTE",
			Status:      "IN_PROGRESS",
		},
	}, nil
}

// func (h *QuoteHandler) CalculateQuote(sctx *serverRoute.Context, req QuoteCalculateRequest) (*resp.QuoteCalculateResponse, error) {
// 	// Get product configuration
// 	product, err := h.productRepo.GetProductByCode(sctx.Ctx, req.ProductCode)
// 	if err != nil {
// 		log.Error(sctx.Ctx, "Error fetching product: %v", err)
// 		return nil, err
// 	}

// 	// Calculate age from DOB
// 	dob, err := time.Parse("2006-01-02", req.DateOfBirth)
// 	if err != nil {
// 		return nil, err
// 	}
// 	ageAtEntry := calculateAge(dob)

// 	// Check eligibility
// 	eligibility := domain.EligibilityResult{
// 		IsEligible:  true,
// 		AgeAtEntry:  ageAtEntry,
// 		MaturityAge: ageAtEntry + req.PolicyTerm,
// 	}

// 	if !product.IsEligibleAge(ageAtEntry) {
// 		eligibility.IsEligible = false
// 		eligibility.RejectReason = "Age not eligible for this product"
// 	}

// 	if !product.IsEligibleSA(req.SumAssured) {
// 		eligibility.IsEligible = false
// 		eligibility.RejectReason = "Sum assured outside product limits: " + fmt.Sprintf("%.2f", req.SumAssured)
// 	}

// 	// Calculate premium
// 	var basePremium, gstAmount, totalPayable, rebate float64
// 	var rate float64
// 	var sumAssdFloat float64
// 	if eligibility.IsEligible {
// 		// Get rate from Sankalan table
// 		rateValue, sumAssd, err := h.quoteRepo.GetPremiumRate(sctx.Ctx, req.ProductCode, ageAtEntry, domain.Gender(req.Gender), req.PolicyTerm)
// 		if err != nil {
// 			log.Error(sctx.Ctx, "Error fetching premium rate: %v", err)
// 			return nil, err
// 		}
// 		// Calculate base premium: (SA / 1000) * Rate
// 		sumAssdFloat, err = strconv.ParseFloat(sumAssd, 64)
// 		if err != nil {
// 			return nil, fmt.Errorf("invalid sum_assd value: %v", err)
// 		}
// 		rate = rateValue
// 		basePremium = (req.SumAssured / sumAssdFloat) * rate

// 		// Calculate rebate based on frequency
// 		rebate = calculateRebate(basePremium, domain.PremiumFrequency(req.Frequency))
// 		netPremium := basePremium - rebate

// 		// Calculate GST (18%)
// 		gstAmount = netPremium * 0.0
// 		totalPayable = netPremium + gstAmount
// 	}

// 	// Calculate benefit illustration with configured bonus rate
// 	maturityValueGuaranteed := req.SumAssured
// 	indicativeBonusRate := h.cfg.GetFloat64("quote.default_bonus_rate")
// 	if indicativeBonusRate <= 0 {
// 		indicativeBonusRate = 4.5 // fallback default
// 	}
// 	maturityValueWithBonus := req.SumAssured * (1 + (indicativeBonusRate/100)*float64(req.PolicyTerm))

// 	// Generate UUID for calculation ID
// 	calculationID := uuid.New().String()

// 	return &resp.QuoteCalculateResponse{
// 		StatusCodeAndMessage: port.StatusCodeAndMessage{
// 			StatusCode: http.StatusOK,
// 			Message:    "Premium calculated successfully",
// 		},
// 		CalculationID: calculationID,
// 		Eligibility:   resp.MapDomainToEligibility(eligibility),
// 		PremiumBreakdown: resp.PremiumBreakdownResponse{
// 			BasePremium:  basePremium,
// 			Rebate:       rebate,
// 			NetPremium:   basePremium - rebate,
// 			CGST:         gstAmount / 2,
// 			SGST:         gstAmount / 2,
// 			IGST:         0,
// 			TotalGST:     gstAmount,
// 			TotalPayable: totalPayable,
// 		},
// 		BenefitIllustration: resp.BenefitIllustrationResponse{
// 			MaturityValueGuaranteed: maturityValueGuaranteed,
// 			MaturityValueWithBonus:  maturityValueWithBonus,
// 			IndicativeBonusRate:     indicativeBonusRate,
// 			DeathBenefit:            req.SumAssured,
// 		},
// 		CalculationBasis: resp.CalculationBasisResponse{
// 			PremiumTable: "Sankalan",
// 			SumAssd:      sumAssdFloat,
// 			// PremiumPerSumAssd: rate,
// 			GSTRate: 0.0,
// 		},
// 		WorkflowState: port.WorkflowStateResponse{
// 			CurrentStep: "CALCULATE_QUOTE",
// 			NextStep:    "CREATE_QUOTE",
// 			Status:      "IN_PROGRESS",
// 		},
// 	}, nil
// }

// CreateQuote saves a calculated quote
// [POL-API-003] Create Quote
func (h *QuoteHandler) CreateQuote(sctx *serverRoute.Context, req QuoteCreateRequest) (*resp.QuoteCreateResponse, error) {
	// Calculate premium expiry using configured validity days
	validityDays := h.cfg.GetInt("quote.validity_days")
	if validityDays <= 0 {
		validityDays = 30 // fallback default
	}
	expiresAt := time.Now().Add(time.Duration(validityDays) * 24 * time.Hour)
	dob, _ := time.Parse("2006-01-02", req.Proposer.DOB)
	gender := domain.Gender(req.Proposer.Gender)

	quote := domain.Quote{
		ProductCode:      req.ProductCode,
		PolicyType:       domain.PolicyType(req.PolicyType),
		ProposerName:     &req.Proposer.Name,
		ProposerDOB:      &dob,
		ProposerGender:   &gender,
		ProposerMobile:   &req.Proposer.Mobile,
		ProposerEmail:    &req.Proposer.Email,
		SumAssured:       req.Coverage.SumAssured,
		PolicyTerm:       req.Coverage.PolicyTerm,
		PaymentFrequency: domain.PremiumFrequency(req.Coverage.PaymentFrequency),
		BasePremium:      req.Premium.BasePremium,
		GSTAmount:        req.Premium.TotalGST,
		TotalPayable:     req.Premium.TotalPayable,
		Channel:          domain.Channel(req.Channel),
		Status:           domain.QuoteStatusGenerated,
		CreatedBy:        req.CreatedBy,
		ExpiresAt:        &expiresAt,
	}

	// Save to database - this will populate QuoteID, QuoteRefNumber, CreatedAt, UpdatedAt
	if err := h.quoteRepo.CreateQuote(sctx.Ctx, &quote); err != nil {
		log.Error(sctx.Ctx, "Error creating quote: %v", err)
		return nil, apierrors.HandleErrorWithStatusCodeAndMessage(
			apierrors.HTTPErrorServerError,
			"Failed to create quote",
			err,
		)
	}

	// TODO: [FR-POL-001] DMS Integration - Store quote document in Document Management System
	// Reference: nbf/userjourneys/policy_issue_user_journeys.md:303
	// dmsDocID, err := h.dmsClient.StoreQuoteDocument(sctx.Ctx, quote)

	// TODO: [FR-POL-001] Event Emission - Publish QuoteCreated event to message bus
	// Reference: nbf/userjourneys/policy_issue_user_journeys.md:303
	// err := h.eventBus.PublishQuoteCreated(sctx.Ctx, quote)

	var pdfInfo *resp.QuotePDFInfo
	if req.GeneratePDF {
		pdfInfo = &resp.QuotePDFInfo{
			DocumentID:  "DOC-" + quote.QuoteRefNumber,
			DownloadURL: "/api/v1/documents/" + quote.QuoteRefNumber + ".pdf",
		}
	}

	// return &resp.QuoteCreateResponse{
	// 	StatusCodeAndMessage: port.StatusCodeAndMessage{
	// 		StatusCode: http.StatusCreated,
	// 		Message:    "Quote created successfully",
	// 	},
	// 	QuoteID:        strconv.FormatInt(quote.QuoteID, 10),
	// 	QuoteRefNumber: quote.QuoteRefNumber,
	// 	Status:         string(quote.Status),
	// 	Validity: resp.QuoteValidityInfo{
	// 		ExpiresAt: expiresAt,
	// 		DaysValid: 30,
	// 	},
	// 	PDFDocument: pdfInfo,
	// }, nil
	return &resp.QuoteCreateResponse{
		StatusCodeAndMessage: port.StatusCodeAndMessage{
			StatusCode: http.StatusCreated,
			Message:    "Quote created successfully",
		},
		QuoteID:        strconv.FormatInt(quote.QuoteID, 10),
		QuoteRefNumber: quote.QuoteRefNumber,
		Status:         string(quote.Status),

		PremiumBreakdown: resp.PremiumBreakdownResponse{
			BasePremium:  quote.BasePremium,
			Rebate:       req.Premium.Rebate,
			NetPremium:   req.Premium.NetPremium,
			CGST:         req.Premium.CGST,
			SGST:         req.Premium.SGST,
			IGST:         0,
			TotalGST:     quote.GSTAmount,
			StampDuty:    0,
			TotalPayable: quote.TotalPayable,
		},

		Validity: resp.QuoteValidityInfo{
			ExpiresAt: expiresAt,
			DaysValid: validityDays,
		},

		PDFDocument: pdfInfo,

		WorkflowState: port.WorkflowStateResponse{
			CurrentStep: "QUOTE_CREATED",
			NextStep:    "CREATE_PROPOSAL",
			Status:      "IN_PROGRESS",
		},
	}, nil

}

// ConvertQuoteToProposal converts a quote to a proposal
// [POL-API-004] Convert Quote to Proposal
// [FR-POL-003] Quote-to-Proposal Conversion
// [BR-POL-024] Deduplication check
func (h *QuoteHandler) ConvertQuoteToProposal(sctx *serverRoute.Context, req QuoteConvertRequest) (*resp.QuoteConvertResponse, error) {
	quote, err := h.quoteRepo.GetQuoteByRefNumber(sctx.Ctx, req.QuoteRefNumber)
	if err != nil {
		log.Error(sctx.Ctx, "Error fetching quote: %v", err)
		return nil, handleRepoError(err, "Quote not found", "Failed to fetch quote")
	}

	// Check if quote can be converted
	if !quote.CanBeConverted() {
		return &resp.QuoteConvertResponse{
			StatusCodeAndMessage: port.StatusCodeAndMessage{
				StatusCode: http.StatusBadRequest,
				Message:    "ERR-POL-003: Quote cannot be converted — expired or already converted",
			},
		}, nil
	}

	// [BR-POL-024] Deduplication check — prevent converting same quote twice
	existingFromQuote, err := h.proposalRepo.CheckDuplicateQuoteConversion(sctx.Ctx, quote.QuoteRefNumber)
	if err != nil {
		log.Error(sctx.Ctx, "Failed to check for duplicate quote conversion", "quoteRef", quote.QuoteRefNumber, "error", err)
		return nil, err
	}
	if existingFromQuote != nil {
		return &resp.QuoteConvertResponse{
			StatusCodeAndMessage: port.StatusCodeAndMessage{
				StatusCode: http.StatusConflict,
				// Message:    fmt.Sprintf("ERR-POL-055: Quote %s already converted to proposal %s (status: %s)", quote.QuoteRefNumber, existingFromQuote.ProposalNumber, existingFromQuote.Status),
				Message: fmt.Sprintf("Quote already converted. Redirecting to proposal %s", existingFromQuote.ProposalNumber),
			},
			QuoteRefNumber: quote.QuoteRefNumber,
			ProposalID:     existingFromQuote.ProposalID,
			ProposalNumber: existingFromQuote.ProposalNumber,
			Status:         existingFromQuote.Status,
			RedirectURL:    fmt.Sprintf("/proposals/%d", existingFromQuote.ProposalID),
		}, nil
	}

	// [BR-POL-024] Deduplication check — prevent duplicate proposals for same customer + product
	// existingProposal, err := h.proposalRepo.CheckDuplicateProposal(sctx.Ctx,  quote.ProductCode)
	// if err != nil {
	// 	log.Error(sctx.Ctx, "Failed to check for duplicate proposals", "customerID", req.CustomerID, "error", err)
	// 	return nil, err
	// }
	// if existingProposal != nil {
	// 	return &resp.QuoteConvertResponse{
	// 		StatusCodeAndMessage: port.StatusCodeAndMessage{
	// 			StatusCode: http.StatusConflict,
	// 			// Message:    fmt.Sprintf("ERR-POL-055: Duplicate proposal exists — proposal %s (status: %s) for customer %d with product %s is already in progress", existingProposal.ProposalNumber, existingProposal.Status, req.CustomerID, quote.ProductCode),
	// 			Message: fmt.Sprintf("Quote already converted. Redirecting to proposal %s", existingProposal.ProposalNumber),
	// 		},
	// 		QuoteID:        req.QuoteID,
	// 		ProposalID:     existingProposal.ProposalID,
	// 		ProposalNumber: existingProposal.ProposalNumber,
	// 		Status:         existingProposal.Status,
	// 		RedirectURL:    fmt.Sprintf("/proposals/%d", existingProposal.ProposalID),
	// 	}, nil
	// }

	// // TODO: Actual proposal creation from quote data
	// return &resp.QuoteConvertResponse{
	// 	StatusCodeAndMessage: port.StatusCodeAndMessage{
	// 		StatusCode: http.StatusCreated,
	// 		Message:    "Quote converted to proposal successfully",
	// 	},
	// 	QuoteID:        req.QuoteID,
	// 	ProposalID:     12345,
	// 	ProposalNumber: "PLI-MH-2026-00056789",
	// 	Status:         "DATA_ENTRY",
	// 	RedirectURL:    "/proposals/12345/insured-details",
	// }, nil
	// --- Build Proposal domain object from Quote ---
	// -----------------------------
	// Build Proposal from Quote
	// -----------------------------

	quoteRef := quote.QuoteRefNumber
	now := time.Now()

	proposal := &domain.Proposal{
		PolicyType:              quote.PolicyType, // must exist in Quote
		ProductCode:             quote.ProductCode,
		CustomerID:              &req.CustomerID,
		SumAssured:              quote.SumAssured,
		PolicyTerm:              quote.PolicyTerm,
		PremiumPaymentFrequency: domain.PremiumFrequency("MONTHLY"), // default if not in quote
		EntryPath:               domain.EntryPath("QUOTE_CONVERSION"),
		Channel:                 domain.Channel("WEB"),
		Status:                  domain.ProposalStatusIndexed,
		QuoteRefNumber:          &quoteRef,
		CreatedBy:               req.CreatedBy,
		BasePremium:             0,
		GSTAmount:               0,
		TotalPremium:            0,
	}

	// -----------------------------
	// Minimal Indexing Data
	// -----------------------------
	indexing := &domain.ProposalIndexing{
		POCode:          "WEB",
		IssueCircle:     "WEB",
		IssueHO:         "WEB",
		IssuePostOffice: "WEB",
		DeclarationDate: now,
		ReceiptDate:     now,
		IndexingDate:    now,
		ProposalDate:    now,
	}

	// -----------------------------
	// Create Proposal in DB
	// -----------------------------
	if err := h.proposalRepo.CreateProposalWithIndexing(sctx.Ctx, proposal, indexing); err != nil {
		log.Error(sctx.Ctx, "Error converting quote to proposal: %v", err)
		return nil, err
	}

	// -----------------------------
	// Return Real Response
	// -----------------------------
	return &resp.QuoteConvertResponse{
		StatusCodeAndMessage: port.StatusCodeAndMessage{
			StatusCode: http.StatusCreated,
			Message:    "Quote converted to proposal successfully",
		},
		QuoteRefNumber: quote.QuoteRefNumber,
		ProposalID:     proposal.ProposalID,
		ProposalNumber: proposal.ProposalNumber,
		Status:         string(proposal.Status),
		RedirectURL:    fmt.Sprintf("/v1/proposals/%d", proposal.ProposalID),
	}, nil

}

// Helper functions
// calculateAge calculates age in completed years using UTC for consistency
func calculateAge(dob time.Time) int {
	now := time.Now().UTC()
	// Normalize dob to UTC for comparison
	dob = dob.UTC()
	years := now.Year() - dob.Year()
	// Check if birthday has occurred this year
	if now.Month() < dob.Month() || (now.Month() == dob.Month() && now.Day() < dob.Day()) {
		years--
	}
	return years
}

// func calculateRebate(basePremium float64, frequency domain.PremiumFrequency) float64 {
// 	switch frequency {
// 	case domain.FrequencyYearly:
// 		return basePremium * 0.02
// 	case domain.FrequencyHalfYearly:
// 		return basePremium * 0.01
// 	default:
// 		return 0
// 	}
// }

func calculateLargeSARebate(productCode string, sumAssured int) float64 {

	// Yugal Suraksha (Product Code: 1005)
	if productCode == "1005" {

		// Rebate starts from ₹40,000
		if sumAssured < 40000 {
			return 0
		}

		// ₹1 at 40,000 + ₹1 per 10,000 above 40,000
		return float64(1 + (sumAssured-40000)/10000)
	}
	// Other PLI products
	// Other PLI products
	if sumAssured < 20000 {
		return 0
	}

	// ₹1 per ₹20,000
	return float64(sumAssured / 20000)
}
func (h *QuoteHandler) GetQuoteByID(sctx *serverRoute.Context, req GetQuoteRequestParams,
) (*resp.GetQuoteResponse, error) {

	quote, err := h.quoteRepo.GetQuoteByID(sctx.Ctx, req.QuoteID)
	if err != nil {
		log.Error(sctx.Ctx, "Error fetching quote: %v", err)
		return nil, apierrors.HandleErrorWithStatusCodeAndMessage(
			apierrors.HTTPErrorServerError,
			"Failed to fetch quote",
			err,
		)
	}

	// If no quote found
	if quote == nil {
		return nil, apierrors.HandleErrorWithStatusCodeAndMessage(
			apierrors.HTTPErrorNotFound,
			"Quote not found",
			nil,
		)
	}
	var gender *string
	if quote.ProposerGender != nil {
		g := string(*quote.ProposerGender)
		gender = &g
	}
	respData := resp.QuoteDetailResponse{
		QuoteID:             quote.QuoteID,
		QuoteRefNumber:      quote.QuoteRefNumber,
		ProductCode:         quote.ProductCode,
		PolicyType:          string(quote.PolicyType),
		CustomerID:          quote.CustomerID,
		ProposerName:        quote.ProposerName,
		ProposerDOB:         quote.ProposerDOB,
		ProposerGender:      gender,
		ProposerMobile:      quote.ProposerMobile,
		ProposerEmail:       quote.ProposerEmail,
		SumAssured:          quote.SumAssured,
		PolicyTerm:          quote.PolicyTerm,
		PaymentFrequency:    string(quote.PaymentFrequency),
		BasePremium:         quote.BasePremium,
		GSTAmount:           quote.GSTAmount,
		TotalPayable:        quote.TotalPayable,
		MaturityValue:       quote.MaturityValue,
		BonusRate:           quote.BonusRate,
		Channel:             string(quote.Channel),
		Status:              string(quote.Status),
		ConvertedProposalID: quote.ConvertedProposalID,
		PDFDocumentID:       quote.PDFDocumentID,
		// CreatedBy:           quote.CreatedBy,
		// CreatedAt:           quote.CreatedAt,
		ExpiresAt: quote.ExpiresAt,
	}

	return &resp.GetQuoteResponse{
		StatusCodeAndMessage: port.StatusCodeAndMessage{
			StatusCode: http.StatusOK,
			Message:    "Quote retrieved successfully",
		},
		Quote: respData,
	}, nil
}

func (h *QuoteHandler) GenerateQuote(sctx *serverRoute.Context, req QuoteGenerateRequest) (*resp.QuoteGenerateResponse, error) {

	layout := "2006-01-02"

	// ---------------------------------
	// FETCH PRODUCT FIRST
	// ---------------------------------
	product, err := h.productRepo.GetProductByCode(sctx.Ctx, req.ProductCode)
	if err != nil {
		return nil, fmt.Errorf("invalid product code")
	}

	// // ---------------------------------
	// // VALIDATE PRODUCT CATEGORY
	// // ---------------------------------
	// if req.ProductCategory == "" {
	// 	return nil, fmt.Errorf(
	// 		"product_category is required for product %s",
	// 		req.ProductCode,
	// 	)
	// }

	// ---------------------------------
	// AGE CALCULATION (USE DOMAIN LOGIC)
	// ---------------------------------
	ageAtEntry, err := domain.CalculateAgeAtEntry(sctx.Ctx, req.ProductCode,
		req.Proposer.DOB, req.SpouseDOB, req.DateOfCalculation,
		h.quoteRepo.GetJointLifeAgeAddition,
	)
	if err != nil {
		log.Error(sctx.Ctx, "Age calculation failed: %v", err)
		return nil, apierrors.HandleErrorWithStatusCodeAndMessage(
			apierrors.HTTPErrorBadRequest,
			err.Error(),
			err,
		)
	}
	err = domain.ValidateProductAge(req.ProductCode, ageAtEntry, req.Term)
	if err != nil {
		return nil, apierrors.HandleErrorWithStatusCodeAndMessage(
			apierrors.HTTPErrorBadRequest,
			err.Error(),
			err,
		)
	}
	// LOOKUP FIELD BASED ON PRODUCT CODE
	var lookupField string
	var lookupValue int
	var effectiveTerm int

	switch req.ProductCode {
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
		effectiveTerm = req.Term
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
		effectiveTerm = req.PremiumCeasingAge - ageAtEntry

	default:
		return nil, apierrors.HandleErrorWithStatusCodeAndMessage(
			apierrors.HTTPErrorBadRequest,
			fmt.Sprintf("product code not found: %s", req.ProductCode),
			nil,
		)
	}

	rate, baseSumAssd, err := h.quoteRepo.GetPremiumRate(sctx.Ctx, req.ProductCode, string(product.ProductCategory),
		ageAtEntry, "ALL", req.Periodicity, lookupField, lookupValue)
	if err != nil {
		log.Error(sctx.Ctx, "GetPremiumRate failed: %v", err)
		return nil, err
	}

	indicativeBonusRate := h.cfg.GetFloat64("quote.default_bonus_rate")
	if indicativeBonusRate <= 0 {
		indicativeBonusRate = 4.5
	}

	gstRate := h.cfg.GetFloat64("quote.gst_rate")
	if gstRate <= 0 {
		gstRate = 0
	}

	var calculations []resp.QuoteCalculationItem

	displayIncrement := 100000
	calculationSlab := int(baseSumAssd)

	baseEligibility := domain.EligibilityResult{
		IsEligible:  true,
		AgeAtEntry:  ageAtEntry,
		MaturityAge: ageAtEntry + effectiveTerm,
	}

	if !product.IsEligibleAge(ageAtEntry) {
		baseEligibility.IsEligible = false
		baseEligibility.RejectReason = "Age not eligible"
	}

	if !product.IsEligibleSA(float64(req.SumAssured)) {
		baseEligibility.IsEligible = false
		baseEligibility.RejectReason = "Sum assured outside limits"
	}

	minSA := int(product.MinSumAssured)

	maxSA := 0
	if product.MaxSumAssured != nil {
		maxSA = int(*product.MaxSumAssured)
	}

	for i := 0; i < 5; i++ {

		projectedSA := req.SumAssured + (i * displayIncrement)

		// Do not exceed max sum assured
		if maxSA > 0 && projectedSA > maxSA {
			break
		}

		// Ensure above min
		if projectedSA < minSA {
			continue
		}

		if projectedSA%calculationSlab != 0 {
			continue
		}

		units := projectedSA / int(baseSumAssd)
		basePremium := math.Round(float64(units)*rate*100) / 100

		rebate, err := h.quoteRepo.GetRebate(sctx.Ctx, req.ProductCode, projectedSA)
		if err != nil {
			return nil, err
		}

		netPremium := basePremium - rebate
		gstAmount := netPremium * gstRate
		totalPayable := netPremium + gstAmount

		maturityWithBonus :=
			float64(projectedSA) *
				(1 + (indicativeBonusRate/100)*float64(effectiveTerm))

		calculations = append(calculations, resp.QuoteCalculationItem{
			SumAssured:  int64(projectedSA),
			Eligibility: resp.MapDomainToEligibility(baseEligibility),
			PremiumBreakdown: resp.PremiumBreakdownResponse{
				BasePremium:  basePremium,
				Rebate:       rebate,
				NetPremium:   netPremium,
				CGST:         gstAmount / 2,
				SGST:         gstAmount / 2,
				IGST:         0,
				TotalGST:     gstAmount,
				TotalPayable: totalPayable,
			},
			BenefitIllustration: resp.BenefitIllustrationResponse{
				MaturityValueGuaranteed: float64(projectedSA),
				MaturityValueWithBonus:  maturityWithBonus,
				IndicativeBonusRate:     indicativeBonusRate,
				DeathBenefit:            float64(projectedSA),
			},
		})
	}

	if len(calculations) == 0 {
		return nil, fmt.Errorf("no valid slab found")
	}

	selected := calculations[0]

	//INSERT QUOTE

	validityDays := h.cfg.GetInt("quote.validity_days")
	if validityDays <= 0 {
		validityDays = 30
	}

	expiresAt := time.Now().Add(time.Duration(validityDays) * 24 * time.Hour)
	pdfDocID := "DOC-" + uuid.New().String()

	proposerDOB, _ := time.Parse(layout, req.Proposer.DOB)
	proposerGender := domain.Gender(req.Proposer.Gender)
	maturityValue := selected.BenefitIllustration.MaturityValueWithBonus
	bonusRate := selected.BenefitIllustration.IndicativeBonusRate
	rebateValue := selected.PremiumBreakdown.Rebate

	quote := domain.Quote{
		ProductCode:      req.ProductCode,
		PolicyType:       domain.PolicyType(req.PolicyType),
		ProposerName:     &req.Proposer.Name,
		ProposerDOB:      &proposerDOB,
		ProposerGender:   &proposerGender,
		SumAssured:       float64(req.SumAssured),
		PolicyTerm:       effectiveTerm,
		PaymentFrequency: domain.PremiumFrequency(req.Periodicity),
		BasePremium:      selected.PremiumBreakdown.BasePremium,
		GSTAmount:        selected.PremiumBreakdown.TotalGST,
		TotalPayable:     selected.PremiumBreakdown.TotalPayable,
		MaturityValue:    &maturityValue,
		BonusRate:        &bonusRate,
		Status:           domain.QuoteStatusGenerated,
		Channel:          domain.Channel(req.Channel),
		CreatedBy:        req.CreatedBy,
		ExpiresAt:        &expiresAt,
		PDFDocumentID:    &pdfDocID,
		Rebate:           &rebateValue,
		ProposerMobile:   &req.Proposer.Mobile,
		ProposerEmail:    &req.Proposer.Email,
	}

	if err := h.quoteRepo.CreateQuote(sctx.Ctx, &quote); err != nil {
		return nil, err
	}

	// FINAL RESPONSE (COMBINED)
	return &resp.QuoteGenerateResponse{
		StatusCodeAndMessage: port.StatusCodeAndMessage{
			StatusCode: http.StatusCreated,
			Message:    "Quote generated successfully",
		},
		ProductCategory: string(product.ProductCategory),
		CalculationBasis: resp.CalculationBasisResponse{
			PremiumTable: "Sankalan",
			SumAssd:      float64(baseSumAssd),
			GSTRate:      gstRate,
		},
		Calculations:     calculations,
		QuoteID:          strconv.FormatInt(quote.QuoteID, 10),
		QuoteRefNumber:   quote.QuoteRefNumber,
		Status:           string(quote.Status),
		PremiumBreakdown: selected.PremiumBreakdown,
		PDFDocument: &resp.QuotePDFInfo{
			DocumentID:  pdfDocID,
			DownloadURL: "/api/v1/documents/" + quote.QuoteRefNumber + ".pdf",
		},
		Validity: resp.QuoteValidityInfo{
			ExpiresAt: expiresAt,
			DaysValid: validityDays,
		},
		WorkflowState: port.WorkflowStateResponse{
			CurrentStep: "QUOTE_CREATED",
			NextStep:    "CREATE_PROPOSAL",
			Status:      "IN_PROGRESS",
		},
	}, nil
}

func (h *QuoteHandler) GetQuoteByNumber(sctx *serverRoute.Context, req GetQuoteRefRequestParams,
) (*resp.GetQuoteResponse, error) {

	ctx := sctx.Ctx

	quote, err := h.quoteRepo.GetQuoteByRefNumber(ctx, req.QuoteRefNumber)
	if err != nil {

		if err == pgx.ErrNoRows {
			return nil, apierrors.HandleErrorWithStatusCodeAndMessage(
				apierrors.HTTPErrorNotFound,
				"Quote not found",
				nil,
			)
		}

		log.Error(ctx, "Error fetching quote: %v", err)
		return nil, apierrors.HandleErrorWithStatusCodeAndMessage(
			apierrors.HTTPErrorServerError,
			"Failed to fetch quote",
			err,
		)
	}

	var gender *string
	if quote.ProposerGender != nil {
		g := string(*quote.ProposerGender)
		gender = &g
	}

	respData := resp.QuoteDetailResponse{
		QuoteID:             quote.QuoteID,
		QuoteRefNumber:      quote.QuoteRefNumber,
		ProductCode:         quote.ProductCode,
		PolicyType:          string(quote.PolicyType),
		CustomerID:          quote.CustomerID,
		ProposerName:        quote.ProposerName,
		ProposerDOB:         quote.ProposerDOB,
		ProposerGender:      gender,
		ProposerMobile:      quote.ProposerMobile,
		ProposerEmail:       quote.ProposerEmail,
		SumAssured:          quote.SumAssured,
		PolicyTerm:          quote.PolicyTerm,
		PaymentFrequency:    string(quote.PaymentFrequency),
		BasePremium:         quote.BasePremium,
		GSTAmount:           quote.GSTAmount,
		TotalPayable:        quote.TotalPayable,
		MaturityValue:       quote.MaturityValue,
		BonusRate:           quote.BonusRate,
		Channel:             string(quote.Channel),
		Status:              string(quote.Status),
		ConvertedProposalID: quote.ConvertedProposalID,
		PDFDocumentID:       quote.PDFDocumentID,
		// CreatedBy:           quote.CreatedBy,
		// CreatedAt:           quote.CreatedAt,
		ExpiresAt: quote.ExpiresAt,
	}

	return &resp.GetQuoteResponse{
		StatusCodeAndMessage: port.StatusCodeAndMessage{
			StatusCode: http.StatusOK,
			Message:    "Quote retrieved successfully",
		},
		Quote: respData,
	}, nil
}
