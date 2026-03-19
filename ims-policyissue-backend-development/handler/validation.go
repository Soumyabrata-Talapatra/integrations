package handler

import (
	"fmt"
	"math"
	"net/http"
	"regexp"
	"time"

	"policy-issue-service/core/domain"
	"policy-issue-service/core/port"
	resp "policy-issue-service/handler/response"
	repo "policy-issue-service/repo/postgres"

	config "gitlab.cept.gov.in/it-2.0-common/api-config"
	log "gitlab.cept.gov.in/it-2.0-common/n-api-log"
	serverHandler "gitlab.cept.gov.in/it-2.0-common/n-api-server/handler"
	serverRoute "gitlab.cept.gov.in/it-2.0-common/n-api-server/route"
)

// Regex patterns for format validation
var (
	aadhaarPattern = regexp.MustCompile(`^[0-9]{12}$`)
	panPattern     = regexp.MustCompile(`^[A-Z]{5}[0-9]{4}[A-Z]{1}$`)
	pincodePattern = regexp.MustCompile(`^[0-9]{6}$`)
	ifscPattern    = regexp.MustCompile(`^[A-Z]{4}0[A-Z0-9]{6}$`)
)

// ValidationHandler handles pre-validation HTTP endpoints
// Phase 6: [VAL-POL-001] to [VAL-POL-008]
type ValidationHandler struct {
	*serverHandler.Base
	productRepo  *repo.ProductRepository
	proposalRepo *repo.ProposalRepository
	cfg          *config.Config
}

// NewValidationHandler creates a new ValidationHandler instance
func NewValidationHandler(productRepo *repo.ProductRepository, proposalRepo *repo.ProposalRepository, cfg *config.Config) *ValidationHandler {
	base := serverHandler.New("Validation").SetPrefix("/v1").AddPrefix("")
	return &ValidationHandler{
		Base:         base,
		productRepo:  productRepo,
		proposalRepo: proposalRepo,
		cfg:          cfg,
	}
}

// Routes returns the routes for the ValidationHandler
func (h *ValidationHandler) Routes() []serverRoute.Route {
	return []serverRoute.Route{
		serverRoute.POST("/validate/eligibility", h.ValidateEligibility).Name("Validate Eligibility"),
		serverRoute.POST("/validate/aadhaar-format", h.ValidateAadhaarFormat).Name("Validate Aadhaar Format"),
		serverRoute.POST("/validate/pan-format", h.ValidatePANFormat).Name("Validate PAN Format"),
		serverRoute.POST("/validate/pincode", h.ValidatePincode).Name("Validate Pincode"),
		serverRoute.POST("/validate/bank-ifsc", h.ValidateBankIFSC).Name("Validate Bank IFSC"),
		serverRoute.POST("/validate/nominee-shares", h.ValidateNomineeShares).Name("Validate Nominee Shares"),
		serverRoute.POST("/validate/date-chain", h.ValidateDateChain).Name("Validate Date Chain"),
		serverRoute.POST("/validate/aggregate-sa", h.ValidateAggregateSA).Name("Validate Aggregate SA"),
	}
}

// ValidateEligibility performs real-time eligibility check for product
// [VAL-POL-001] Real-time eligibility check
// Components: BR-POL-011, BR-POL-012
// Used In: UJ-POL-001
func (h *ValidationHandler) ValidateEligibility(sctx *serverRoute.Context, req EligibilityCheckRequest) (*resp.EligibilityCheckResponse, error) {
	// Fetch product configuration from database
	product, err := h.productRepo.GetProductByCode(sctx.Ctx, req.ProductCode)
	if err != nil {
		log.Error(sctx.Ctx, "[VAL-POL-001] Error fetching product %s: %v", req.ProductCode, err)
		return nil, fmt.Errorf("product not found: %s", req.ProductCode)
	}

	// Calculate age from DOB
	dob, err := time.Parse("2006-01-02", req.DateOfBirth)
	if err != nil {
		return nil, fmt.Errorf("invalid date_of_birth format, expected YYYY-MM-DD")
	}
	ageAtEntry := domain.CalculateAge(req.DateOfBirth)
	_ = dob // used for age calculation via domain.CalculateAge

	isEligible := true
	var checks []resp.EligibilityCheck

	// [VR-PI-012] Age eligibility check
	ageCheck := resp.EligibilityCheck{
		Check:  "age_eligibility",
		Passed: product.IsEligibleAge(ageAtEntry),
	}
	if ageCheck.Passed {
		ageCheck.Message = fmt.Sprintf("Age %d is within eligible range (%d-%d)", ageAtEntry, product.MinEntryAge, product.MaxEntryAge)
	} else {
		ageCheck.Message = fmt.Sprintf("Age %d is outside eligible range (%d-%d) for product %s", ageAtEntry, product.MinEntryAge, product.MaxEntryAge, req.ProductCode)
		isEligible = false
	}
	checks = append(checks, ageCheck)

	// [VR-PI-013] Sum assured eligibility check
	saCheck := resp.EligibilityCheck{
		Check:  "sum_assured_eligibility",
		Passed: product.IsEligibleSA(req.SumAssured),
	}
	if saCheck.Passed {
		saCheck.Message = "Sum assured is within product limits"
	} else {
		maxSA := "unlimited"
		if product.MaxSumAssured != nil {
			maxSA = fmt.Sprintf("%.2f", *product.MaxSumAssured)
		}
		saCheck.Message = fmt.Sprintf("Sum assured %.2f is outside product limits (min: %.2f, max: %s)", req.SumAssured, product.MinSumAssured, maxSA)
		isEligible = false
	}
	checks = append(checks, saCheck)

	// [VR-PI-044] Policy term validation (if provided)
	if req.PolicyTerm > 0 {
		termCheck := resp.EligibilityCheck{
			Check:  "policy_term_eligibility",
			Passed: req.PolicyTerm >= product.MinTerm,
		}
		if termCheck.Passed {
			termCheck.Message = fmt.Sprintf("Policy term %d meets minimum requirement (%d)", req.PolicyTerm, product.MinTerm)
		} else {
			termCheck.Message = fmt.Sprintf("Policy term %d is below minimum (%d) for product %s", req.PolicyTerm, product.MinTerm, req.ProductCode)
			isEligible = false
		}
		checks = append(checks, termCheck)

		// Maturity age check
		maturityAge := ageAtEntry + req.PolicyTerm
		maturityCheck := resp.EligibilityCheck{
			Check:  "maturity_age_check",
			Passed: true,
		}
		if product.MaxMaturityAge != nil && maturityAge > *product.MaxMaturityAge {
			maturityCheck.Passed = false
			maturityCheck.Message = fmt.Sprintf("Maturity age %d exceeds maximum allowed %d", maturityAge, *product.MaxMaturityAge)
			isEligible = false
		} else {
			maturityCheck.Message = fmt.Sprintf("Maturity age %d is within limits", maturityAge)
		}
		checks = append(checks, maturityCheck)
	}

	// Medical requirement check (informational)
	medicalCheck := resp.EligibilityCheck{
		Check:  "medical_requirement",
		Passed: true, // This is informational, doesn't affect eligibility
	}
	if product.IsMedicalRequired(req.SumAssured) {
		medicalCheck.Message = "Medical examination is required for this sum assured"
	} else {
		medicalCheck.Message = "No medical examination required"
	}
	checks = append(checks, medicalCheck)

	// Aggregate SA check (if existing SA provided)
	var aggregateSA *resp.AggregateSAInfo
	if req.ExistingAggregateSA > 0 {
		// TODO: [INT-POL-002] In future, fetch from Customer Service
		proposedTotal := req.ExistingAggregateSA + req.SumAssured

		// Use product-specific MaxSumAssured as the aggregate limit.
		// Fall back to configurable default (validation.default_max_aggregate_sa).
		maxAllowed := h.cfg.GetFloat64("validation.default_max_aggregate_sa")
		if maxAllowed <= 0 {
			maxAllowed = 5000000.0 // ultimate fallback
		}
		if product.MaxSumAssured != nil {
			maxAllowed = *product.MaxSumAssured
		}
		withinLimit := proposedTotal <= maxAllowed

		aggregateSA = &resp.AggregateSAInfo{
			Current:     req.ExistingAggregateSA,
			Proposed:    proposedTotal,
			MaxAllowed:  maxAllowed,
			WithinLimit: withinLimit,
		}

		aggregateCheck := resp.EligibilityCheck{
			Check:  "aggregate_sa_check",
			Passed: withinLimit,
		}
		if withinLimit {
			aggregateCheck.Message = fmt.Sprintf("Aggregate SA %.2f is within limit %.2f", proposedTotal, maxAllowed)
		} else {
			aggregateCheck.Message = fmt.Sprintf("Aggregate SA %.2f exceeds maximum allowed %.2f", proposedTotal, maxAllowed)
			isEligible = false
		}
		checks = append(checks, aggregateCheck)
	}

	return &resp.EligibilityCheckResponse{
		StatusCodeAndMessage: port.StatusCodeAndMessage{
			StatusCode: http.StatusOK,
			Message:    "Eligibility check completed",
		},
		IsEligible:        isEligible,
		AgeAtEntry:        ageAtEntry,
		EligibilityChecks: checks,
		AggregateSA:       aggregateSA,
	}, nil
}

// ValidateAadhaarFormat validates Aadhaar number format
// [VAL-POL-002] Aadhaar format validation
// Components: VR-PI-008
// Error: ERR-POL-010
func (h *ValidationHandler) ValidateAadhaarFormat(sctx *serverRoute.Context, req AadhaarFormatRequest) (*resp.FormatValidationResponse, error) {
	isValid := aadhaarPattern.MatchString(req.AadhaarNumber)

	message := "Aadhaar number format is valid"
	if !isValid {
		message = "Invalid Aadhaar number format. Must be exactly 12 digits"
	}

	return &resp.FormatValidationResponse{
		StatusCodeAndMessage: port.StatusCodeAndMessage{
			StatusCode: http.StatusOK,
			Message:    "Aadhaar format validation completed",
		},
		IsValid: isValid,
		Message: message,
	}, nil
}

// ValidatePANFormat validates PAN number format
// [VAL-POL-003] PAN format validation
// Components: VR-PI-009
// Error: ERR-POL-011
func (h *ValidationHandler) ValidatePANFormat(sctx *serverRoute.Context, req PANFormatRequest) (*resp.FormatValidationResponse, error) {
	isValid := panPattern.MatchString(req.PANNumber)

	message := "PAN number format is valid"
	if !isValid {
		message = "Invalid PAN number format. Must match pattern AAAAA9999A (5 uppercase letters, 4 digits, 1 uppercase letter)"
	}

	return &resp.FormatValidationResponse{
		StatusCodeAndMessage: port.StatusCodeAndMessage{
			StatusCode: http.StatusOK,
			Message:    "PAN format validation completed",
		},
		IsValid: isValid,
		Message: message,
	}, nil
}

// pincodeFirstDigitToStates maps the first digit of an Indian pincode to the
// postal circle / states it covers. India Post assigns the first digit based on
// postal regions. This provides a basic cross-check for pincode-state consistency.
// Source: India Post Pincode Directory (https://www.indiapost.gov.in/)
var pincodeFirstDigitToStates = map[byte][]string{
	'1': {"DL", "HR", "HP", "JK", "PB", "CH", "LA"},                         // Northern Region
	'2': {"UP", "UT"},                                                         // Uttar Pradesh / Uttarakhand
	'3': {"RJ", "GJ", "DN"},                                                  // Rajasthan / Gujarat
	'4': {"MH", "GA"},                                                         // Maharashtra / Goa
	'5': {"AP", "TG", "KA"},                                                   // Andhra Pradesh / Telangana / Karnataka
	'6': {"KL", "TN", "PY", "LD"},                                             // Kerala / Tamil Nadu / Puducherry
	'7': {"OR", "WB", "AN", "SK"},                                             // Odisha / West Bengal / Andaman
	'8': {"BR", "JH", "AS", "MN", "ML", "MZ", "NL", "AR", "TR", "SK", "MP"}, // Bihar / North-East / MP
	'9': {"CT"},                                                                // Army Post Offices / Chhattisgarh
}

// ValidatePincode validates pincode format and optionally validates pincode-state match
// [VAL-POL-004] Pincode-state validation
// Components: VR-PI-034, VR-PI-035
// VR-PI-035: Pincode first-digit must be consistent with the supplied state code
func (h *ValidationHandler) ValidatePincode(sctx *serverRoute.Context, req PincodeValidationRequest) (*resp.PincodeValidationResponse, error) {
	isValid := pincodePattern.MatchString(req.Pincode)

	message := "Pincode format is valid"
	if !isValid {
		message = "Invalid pincode format. Must be exactly 6 digits"
		return &resp.PincodeValidationResponse{
			StatusCodeAndMessage: port.StatusCodeAndMessage{
				StatusCode: http.StatusOK,
				Message:    "Pincode validation completed",
			},
			IsValid: false,
			Pincode: req.Pincode,
			State:   req.State,
			Message: message,
		}, nil
	}

	// [VR-PI-035] Validate pincode-state consistency using first-digit postal circle mapping
	if req.State != "" {
		firstDigit := req.Pincode[0]
		validStates, knownDigit := pincodeFirstDigitToStates[firstDigit]
		if knownDigit {
			stateMatch := false
			for _, s := range validStates {
				if s == req.State {
					stateMatch = true
					break
				}
			}
			if !stateMatch {
				isValid = false
				message = fmt.Sprintf("Pincode %s (region %c) does not match state %s", req.Pincode, firstDigit, req.State)
			}
		}
		// Unknown first digit is treated as valid (future postal circles may be added)
	}

	return &resp.PincodeValidationResponse{
		StatusCodeAndMessage: port.StatusCodeAndMessage{
			StatusCode: http.StatusOK,
			Message:    "Pincode validation completed",
		},
		IsValid: isValid,
		Pincode: req.Pincode,
		State:   req.State,
		Message: message,
	}, nil
}

// knownBankCodes maps common IFSC bank codes to bank names.
// This provides basic bank identification without a full RBI directory integration.
// Source: RBI IFSC bank code registry (commonly used codes for India Post / PLI context)
var knownBankCodes = map[string]string{
	"SBIN": "State Bank of India",
	"PUNB": "Punjab National Bank",
	"BARB": "Bank of Baroda",
	"CNRB": "Canara Bank",
	"UBIN": "Union Bank of India",
	"IOBA": "Indian Overseas Bank",
	"BKID": "Bank of India",
	"CBIN": "Central Bank of India",
	"UCBA": "UCO Bank",
	"PSIB": "Punjab & Sind Bank",
	"IDIB": "Indian Bank",
	"MAHB": "Bank of Maharashtra",
	"ALLA": "Allahabad Bank",
	"CORP": "Union Bank (erstwhile Corporation Bank)",
	"HDFC": "HDFC Bank",
	"ICIC": "ICICI Bank",
	"UTIB": "Axis Bank",
	"KKBK": "Kotak Mahindra Bank",
	"YESB": "Yes Bank",
	"IDFB": "IDFC First Bank",
}

// ValidateBankIFSC validates IFSC code format and performs bank code verification
// [VAL-POL-005] IFSC code validation
// Components: VR-PI-030
// Checks:
//  1. Format: 4 uppercase letters + '0' + 6 alphanumeric characters
//  2. Bank code (first 4 chars) lookup against known bank codes
//  3. 5th character must always be '0' (reserved by RBI)
func (h *ValidationHandler) ValidateBankIFSC(sctx *serverRoute.Context, req IFSCValidationRequest) (*resp.IFSCValidationResponse, error) {
	isValid := ifscPattern.MatchString(req.IFSCCode)

	response := &resp.IFSCValidationResponse{
		StatusCodeAndMessage: port.StatusCodeAndMessage{
			StatusCode: http.StatusOK,
			Message:    "IFSC validation completed",
		},
		IsValid:  isValid,
		IFSCCode: req.IFSCCode,
	}

	if !isValid {
		response.StatusCodeAndMessage.Message = "Invalid IFSC code format. Must be 4 uppercase letters, followed by 0, followed by 6 alphanumeric characters"
		return response, nil
	}

	// Extract bank code (first 4 characters) and perform bank lookup
	bankCode := req.IFSCCode[:4]
	if bankName, known := knownBankCodes[bankCode]; known {
		response.BankName = bankName
	} else {
		// Bank code not in our registry — flag as warning but don't invalidate,
		// since we don't have the full RBI directory yet.
		log.Warn(sctx.Ctx, "[VAL-POL-005] IFSC bank code %s not found in known bank codes registry", bankCode)
	}

	// Extract branch code (last 6 characters) for logging/audit
	branchCode := req.IFSCCode[5:]
	_ = branchCode // available for future branch directory lookup

	return response, nil
}

// maxNomineesPerProposal is the maximum number of nominees allowed per proposal.
// Source: E-008 proposal_nominee DDL comment: "Max 3 per proposal"
const maxNomineesPerProposal = 3

// ValidateNomineeShares validates nominee share percentages, count, and appointee rules
// [VAL-POL-006] Nominee share % validation
// Components: VR-PI-018
// Rules:
//   - Shares must total exactly 100%
//   - Maximum 3 nominees per proposal (E-008 constraint)
//   - Minor nominees must have appointee_name and appointee_relationship (chk_minor_appointee)
//
// Error: ERR-POL-023
func (h *ValidationHandler) ValidateNomineeShares(sctx *serverRoute.Context, req NomineeSharesRequest) (*resp.ShareValidationResponse, error) {
	shareResult := func(isValid bool, total float64, msg string) *resp.ShareValidationResponse {
		return &resp.ShareValidationResponse{
			StatusCodeAndMessage: port.StatusCodeAndMessage{
				StatusCode: http.StatusOK,
				Message:    "Nominee share validation completed",
			},
			IsValid:         isValid,
			TotalPercentage: total,
			Message:         msg,
		}
	}

	// Determine which input format was used. Prefer `Nominees` (richer validation);
	// fall back to `Shares` (backward compatibility).
	if len(req.Nominees) > 0 {
		// ── Rich validation path ──

		// Rule 1: Max nominee count
		if len(req.Nominees) > maxNomineesPerProposal {
			return shareResult(false, 0, fmt.Sprintf(
				"Maximum %d nominees allowed per proposal. Received: %d",
				maxNomineesPerProposal, len(req.Nominees))), nil
		}

		totalPercentage := 0.0
		for i, n := range req.Nominees {
			// Rule 2: Share range
			if n.Share <= 0 || n.Share > 100 {
				return shareResult(false, totalPercentage,
					fmt.Sprintf("Nominee %d: share must be between 0 and 100. Found: %.2f", i+1, n.Share)), nil
			}
			totalPercentage += n.Share

			// Rule 3: Appointee rules for minor nominees
			// Mirrors DDL constraint chk_minor_appointee:
			//   (is_minor = TRUE AND appointee_name IS NOT NULL AND appointee_relationship IS NOT NULL) OR (is_minor = FALSE)
			if n.IsMinor {
				if n.AppointeeName == "" {
					return shareResult(false, totalPercentage,
						fmt.Sprintf("Nominee %d is a minor: appointee_name is required", i+1)), nil
				}
				if n.AppointeeRelationship == "" {
					return shareResult(false, totalPercentage,
						fmt.Sprintf("Nominee %d is a minor: appointee_relationship is required", i+1)), nil
				}
			}
		}

		// Rule 4: Total must equal 100%
		isValid := math.Abs(totalPercentage-100.0) < 0.01
		message := "Nominee shares total 100%"
		if !isValid {
			message = fmt.Sprintf("Nominee shares must total exactly 100%%. Current total: %.2f%%", totalPercentage)
		}

		return shareResult(isValid, totalPercentage, message), nil
	}

	// ── Legacy path: simple shares array ──
	if len(req.Shares) == 0 {
		return shareResult(false, 0, "At least one nominee share is required"), nil
	}

	// Rule 1: Max nominee count
	if len(req.Shares) > maxNomineesPerProposal {
		return shareResult(false, 0, fmt.Sprintf(
			"Maximum %d nominees allowed per proposal. Received: %d",
			maxNomineesPerProposal, len(req.Shares))), nil
	}

	totalPercentage := 0.0
	for _, share := range req.Shares {
		if share <= 0 || share > 100 {
			return shareResult(false, totalPercentage,
				fmt.Sprintf("Each nominee share must be between 0 and 100. Found: %.2f", share)), nil
		}
		totalPercentage += share
	}

	// Use tolerance for floating point comparison
	isValid := math.Abs(totalPercentage-100.0) < 0.01

	message := "Nominee shares total 100%"
	if !isValid {
		message = fmt.Sprintf("Nominee shares must total exactly 100%%. Current total: %.2f%%", totalPercentage)
	}

	return shareResult(isValid, totalPercentage, message), nil
}

// ValidateDateChain validates the proposal date sequence
// [VAL-POL-007] Date chain validation
// Components: BR-POL-018
// Rule: declaration_date <= receipt_date <= indexing_date <= proposal_date
// Errors: ERR-POL-006, ERR-POL-007, ERR-POL-008, ERR-POL-009
func (h *ValidationHandler) ValidateDateChain(sctx *serverRoute.Context, req DateChainValidationRequest) (*resp.DateChainValidationResponse, error) {
	var errors []resp.DateChainError

	// Parse all dates
	declarationDate, err := time.Parse("2006-01-02", req.DeclarationDate)
	if err != nil {
		errors = append(errors, resp.DateChainError{
			Field:   "declaration_date",
			Message: "Invalid date format. Expected YYYY-MM-DD",
		})
	}

	receiptDate, err := time.Parse("2006-01-02", req.ReceiptDate)
	if err != nil {
		errors = append(errors, resp.DateChainError{
			Field:   "receipt_date",
			Message: "Invalid date format. Expected YYYY-MM-DD",
		})
	}

	indexingDate, err := time.Parse("2006-01-02", req.IndexingDate)
	if err != nil {
		errors = append(errors, resp.DateChainError{
			Field:   "indexing_date",
			Message: "Invalid date format. Expected YYYY-MM-DD",
		})
	}

	proposalDate, err := time.Parse("2006-01-02", req.ProposalDate)
	if err != nil {
		errors = append(errors, resp.DateChainError{
			Field:   "proposal_date",
			Message: "Invalid date format. Expected YYYY-MM-DD",
		})
	}

	// If any date failed to parse, return errors immediately
	if len(errors) > 0 {
		return &resp.DateChainValidationResponse{
			StatusCodeAndMessage: port.StatusCodeAndMessage{
				StatusCode: http.StatusOK,
				Message:    "Date chain validation completed with errors",
			},
			IsValid: false,
			Errors:  errors,
		}, nil
	}

	// [BR-POL-018] Validate date chain: declaration_date <= receipt_date <= indexing_date <= proposal_date

	// [ERR-POL-007] receipt_date must be >= declaration_date
	if receiptDate.Before(declarationDate) {
		errors = append(errors, resp.DateChainError{
			Field:   "receipt_date",
			Message: fmt.Sprintf("Receipt date (%s) must be on or after declaration date (%s)", req.ReceiptDate, req.DeclarationDate),
		})
	}

	// [ERR-POL-008] indexing_date must be >= receipt_date
	if indexingDate.Before(receiptDate) {
		errors = append(errors, resp.DateChainError{
			Field:   "indexing_date",
			Message: fmt.Sprintf("Indexing date (%s) must be on or after receipt date (%s)", req.IndexingDate, req.ReceiptDate),
		})
	}

	// [ERR-POL-009] proposal_date must be >= indexing_date
	if proposalDate.Before(indexingDate) {
		errors = append(errors, resp.DateChainError{
			Field:   "proposal_date",
			Message: fmt.Sprintf("Proposal date (%s) must be on or after indexing date (%s)", req.ProposalDate, req.IndexingDate),
		})
	}

	// No dates should be in the future
	today := time.Now().Truncate(24 * time.Hour)
	if declarationDate.After(today) {
		errors = append(errors, resp.DateChainError{
			Field:   "declaration_date",
			Message: "Declaration date cannot be in the future",
		})
	}
	if receiptDate.After(today) {
		errors = append(errors, resp.DateChainError{
			Field:   "receipt_date",
			Message: "Receipt date cannot be in the future",
		})
	}
	if indexingDate.After(today) {
		errors = append(errors, resp.DateChainError{
			Field:   "indexing_date",
			Message: "Indexing date cannot be in the future",
		})
	}

	isValid := len(errors) == 0
	message := "Date chain validation passed"
	if !isValid {
		message = "Date chain validation completed with errors"
	}

	return &resp.DateChainValidationResponse{
		StatusCodeAndMessage: port.StatusCodeAndMessage{
			StatusCode: http.StatusOK,
			Message:    message,
		},
		IsValid: isValid,
		Errors:  errors,
	}, nil
}

// ValidateAggregateSA checks aggregate sum assured against limits
// [VAL-POL-008] Aggregate SA check
// Components: INT-POL-002
// Integration: Customer Service (Portfolio check)
func (h *ValidationHandler) ValidateAggregateSA(sctx *serverRoute.Context, req AggregateSACheckRequest) (*resp.AggregateSACheckResponse, error) {
	// TODO: [INT-POL-002] Integrate with Customer Service to fetch existing aggregate SA
	// For now, calculate from local proposal data

	// Fetch existing aggregate SA for customer from proposals
	existingSA, err := h.proposalRepo.GetCustomerAggregateSA(sctx.Ctx, req.CustomerID, req.PolicyType)
	if err != nil {
		log.Error(sctx.Ctx, "[VAL-POL-008] Error fetching aggregate SA for customer %d: %v", req.CustomerID, err)
		// If error, proceed with 0 existing SA
		existingSA = 0
	}

	proposedTotal := existingSA + req.ProposedSA

	// Determine max allowed SA — source from product catalog first,
	// then from configuration, with a hard-coded ultimate fallback.
	maxAllowed := h.cfg.GetFloat64("validation.default_max_aggregate_sa")
	if maxAllowed <= 0 {
		maxAllowed = 5000000.0 // ultimate fallback
	}

	// If product code is provided, use product-specific limits
	if req.ProductCode != "" {
		product, err := h.productRepo.GetProductByCode(sctx.Ctx, req.ProductCode)
		if err == nil && product.MaxSumAssured != nil {
			maxAllowed = *product.MaxSumAssured
		}
	}

	isEligible := proposedTotal <= maxAllowed

	reason := ""
	if !isEligible {
		reason = fmt.Sprintf("Proposed aggregate SA (%.2f) exceeds maximum allowed (%.2f) for %s policies", proposedTotal, maxAllowed, req.PolicyType)
	}

	return &resp.AggregateSACheckResponse{
		StatusCodeAndMessage: port.StatusCodeAndMessage{
			StatusCode: http.StatusOK,
			Message:    "Aggregate SA check completed",
		},
		IsEligible:          isEligible,
		CurrentAggregateSA:  existingSA,
		ProposedAggregateSA: proposedTotal,
		MaxAllowedSA:        maxAllowed,
		Reason:              reason,
	}, nil
}
