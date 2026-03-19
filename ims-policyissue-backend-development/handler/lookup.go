package handler

import (
	"net/http"

	"policy-issue-service/core/port"
	resp "policy-issue-service/handler/response"
	repo "policy-issue-service/repo/postgres"

	log "gitlab.cept.gov.in/it-2.0-common/n-api-log"
	serverHandler "gitlab.cept.gov.in/it-2.0-common/n-api-server/handler"
	serverRoute "gitlab.cept.gov.in/it-2.0-common/n-api-server/route"
)

// LookupHandler handles lookup/reference data HTTP endpoints
// Phase 6: [LU-POL-001] to [LU-POL-010]
type LookupHandler struct {
	*serverHandler.Base
	productRepo *repo.ProductRepository
}

// NewLookupHandler creates a new LookupHandler instance
func NewLookupHandler(productRepo *repo.ProductRepository) *LookupHandler {
	base := serverHandler.New("Lookup").SetPrefix("/v1").AddPrefix("")
	return &LookupHandler{Base: base, productRepo: productRepo}
}

// Routes returns the routes for the LookupHandler
func (h *LookupHandler) Routes() []serverRoute.Route {
	return []serverRoute.Route{
		serverRoute.GET("/lookup/products", h.GetProducts).Name("Lookup Products"),
		serverRoute.GET("/lookup/relationships", h.GetRelationships).Name("Lookup Relationships"),
		serverRoute.GET("/lookup/salutations", h.GetSalutations).Name("Lookup Salutations"),
		serverRoute.GET("/lookup/states", h.GetStates).Name("Lookup States"),
		serverRoute.GET("/lookup/occupations", h.GetOccupations).Name("Lookup Occupations"),
		serverRoute.GET("/lookup/age-proof-types", h.GetAgeProofTypes).Name("Lookup Age Proof Types"),
		serverRoute.GET("/lookup/payment-modes", h.GetPaymentModes).Name("Lookup Payment Modes"),
		serverRoute.GET("/lookup/document-types", h.GetDocumentTypes).Name("Lookup Document Types"),
		serverRoute.GET("/lookup/rejection-reasons", h.GetRejectionReasons).Name("Lookup Rejection Reasons"),
		serverRoute.GET("/lookup/cancellation-reasons", h.GetCancellationReasons).Name("Lookup Cancellation Reasons"),
	}
}

// newLookupResponse is a helper to build a standard LookupResponse
func newLookupResponse(items []resp.LookupItem) *resp.LookupResponse {
	return &resp.LookupResponse{
		StatusCodeAndMessage: port.StatusCodeAndMessage{
			StatusCode: http.StatusOK,
			Message:    "Lookup data retrieved successfully",
		},
		Items: items,
	}
}

// GetProducts retrieves product dropdown list from the database
// [LU-POL-001] Product dropdown
// Used In: UJ-POL-001 Step 1
func (h *LookupHandler) GetProducts(sctx *serverRoute.Context, _ struct{}) (*resp.LookupResponse, error) {
	products, err := h.productRepo.GetAllProducts(sctx.Ctx, "")
	if err != nil {
		log.Error(sctx.Ctx, "Error fetching products for lookup: %v", err)
		return nil, err
	}

	items := make([]resp.LookupItem, len(products))
	for i, p := range products {
		items[i] = resp.LookupItem{
			Code:        p.ProductCode,
			Label:       p.ProductName,
			Description: string(p.ProductType) + " - " + string(p.ProductCategory),
		}
	}

	return newLookupResponse(items), nil
}

// GetRelationships retrieves nominee relationship dropdown
// [LU-POL-002] Nominee relationship dropdown
// Used In: UJ-POL-002 Step 4
// Source: relationship_enum from DDL (001_policy_issue_schema.sql)
// NOTE: Values must stay in sync with the DDL enum definition.
// If the enum changes, update this list accordingly.
func (h *LookupHandler) GetRelationships(sctx *serverRoute.Context, _ struct{}) (*resp.LookupResponse, error) {
	items := []resp.LookupItem{
		{Code: "FATHER", Label: "Father"},
		{Code: "MOTHER", Label: "Mother"},
		{Code: "SPOUSE", Label: "Spouse"},
		{Code: "SON", Label: "Son"},
		{Code: "DAUGHTER", Label: "Daughter"},
		{Code: "BROTHER", Label: "Brother"},
		{Code: "SISTER", Label: "Sister"},
		{Code: "GRANDFATHER", Label: "Grandfather"},
		{Code: "GRANDMOTHER", Label: "Grandmother"},
		{Code: "UNCLE", Label: "Uncle"},
		{Code: "AUNT", Label: "Aunt"},
		{Code: "NEPHEW", Label: "Nephew"},
		{Code: "NIECE", Label: "Niece"},
		{Code: "FRIEND", Label: "Friend"},
		{Code: "OTHER", Label: "Other"},
	}
	return newLookupResponse(items), nil
}

// GetSalutations retrieves salutation dropdown
// [LU-POL-003] Salutation dropdown
// Used In: All insured details screens
// Source: salutation_enum from DDL (001_policy_issue_schema.sql)
// NOTE: Values must stay in sync with the DDL enum definition.
func (h *LookupHandler) GetSalutations(sctx *serverRoute.Context, _ struct{}) (*resp.LookupResponse, error) {
	items := []resp.LookupItem{
		{Code: "MR", Label: "Mr."},
		{Code: "MRS", Label: "Mrs."},
		{Code: "MS", Label: "Ms."},
		{Code: "DR", Label: "Dr."},
		{Code: "SHRI", Label: "Shri"},
		{Code: "SMT", Label: "Smt."},
		{Code: "KUM", Label: "Kumari"},
	}
	return newLookupResponse(items), nil
}

// GetStates retrieves Indian states and union territories dropdown
// [LU-POL-004] State dropdown
// Used In: Address capture
// NOTE: Aligned with Indian postal circle codes. Keep in sync with master data.
func (h *LookupHandler) GetStates(sctx *serverRoute.Context, _ struct{}) (*resp.LookupResponse, error) {
	items := []resp.LookupItem{
		{Code: "AN", Label: "Andaman and Nicobar Islands"},
		{Code: "AP", Label: "Andhra Pradesh"},
		{Code: "AR", Label: "Arunachal Pradesh"},
		{Code: "AS", Label: "Assam"},
		{Code: "BR", Label: "Bihar"},
		{Code: "CH", Label: "Chandigarh"},
		{Code: "CT", Label: "Chhattisgarh"},
		{Code: "DN", Label: "Dadra and Nagar Haveli and Daman and Diu"},
		{Code: "DL", Label: "Delhi"},
		{Code: "GA", Label: "Goa"},
		{Code: "GJ", Label: "Gujarat"},
		{Code: "HR", Label: "Haryana"},
		{Code: "HP", Label: "Himachal Pradesh"},
		{Code: "JK", Label: "Jammu and Kashmir"},
		{Code: "JH", Label: "Jharkhand"},
		{Code: "KA", Label: "Karnataka"},
		{Code: "KL", Label: "Kerala"},
		{Code: "LA", Label: "Ladakh"},
		{Code: "LD", Label: "Lakshadweep"},
		{Code: "MP", Label: "Madhya Pradesh"},
		{Code: "MH", Label: "Maharashtra"},
		{Code: "MN", Label: "Manipur"},
		{Code: "ML", Label: "Meghalaya"},
		{Code: "MZ", Label: "Mizoram"},
		{Code: "NL", Label: "Nagaland"},
		{Code: "OR", Label: "Odisha"},
		{Code: "PY", Label: "Puducherry"},
		{Code: "PB", Label: "Punjab"},
		{Code: "RJ", Label: "Rajasthan"},
		{Code: "SK", Label: "Sikkim"},
		{Code: "TN", Label: "Tamil Nadu"},
		{Code: "TG", Label: "Telangana"},
		{Code: "TR", Label: "Tripura"},
		{Code: "UP", Label: "Uttar Pradesh"},
		{Code: "UT", Label: "Uttarakhand"},
		{Code: "WB", Label: "West Bengal"},
	}
	return newLookupResponse(items), nil
}

// GetOccupations retrieves occupation dropdown
// [LU-POL-005] Occupation dropdown
// Used In: Employment details
func (h *LookupHandler) GetOccupations(sctx *serverRoute.Context, _ struct{}) (*resp.LookupResponse, error) {
	items := []resp.LookupItem{
		{Code: "GOVT_EMPLOYEE", Label: "Government Employee", Description: "Central/State government employee"},
		{Code: "PSU_EMPLOYEE", Label: "PSU Employee", Description: "Public sector undertaking employee"},
		{Code: "DEFENCE", Label: "Defence Personnel", Description: "Armed forces personnel"},
		{Code: "RAILWAY", Label: "Railway Employee", Description: "Indian Railways employee"},
		{Code: "POSTAL", Label: "Postal Employee", Description: "Department of Posts employee"},
		{Code: "TEACHER", Label: "Teacher", Description: "Government/aided school teacher"},
		{Code: "FARMER", Label: "Farmer", Description: "Agricultural worker"},
		{Code: "LABOURER", Label: "Labourer", Description: "Manual labourer / daily wage worker"},
		{Code: "SELF_EMPLOYED", Label: "Self Employed", Description: "Self-employed / business owner"},
		{Code: "PRIVATE_SECTOR", Label: "Private Sector Employee", Description: "Private company employee"},
		{Code: "PROFESSIONAL", Label: "Professional", Description: "Doctor, lawyer, engineer, etc."},
		{Code: "STUDENT", Label: "Student", Description: "Full-time student"},
		{Code: "HOMEMAKER", Label: "Homemaker", Description: "Homemaker / housewife"},
		{Code: "RETIRED", Label: "Retired", Description: "Retired from service"},
		{Code: "OTHER", Label: "Other", Description: "Other occupation"},
	}
	return newLookupResponse(items), nil
}

// GetAgeProofTypes retrieves age proof types dropdown
// [LU-POL-006] Age proof types
// Used In: Document upload
// Source: age_proof_type_enum from DDL (001_policy_issue_schema.sql)
// NOTE: Values must stay in sync with the DDL enum definition.
func (h *LookupHandler) GetAgeProofTypes(sctx *serverRoute.Context, _ struct{}) (*resp.LookupResponse, error) {
	items := []resp.LookupItem{
		{Code: "AADHAAR", Label: "Aadhaar Card", Description: "UIDAI Aadhaar card"},
		{Code: "BIRTH_CERTIFICATE", Label: "Birth Certificate", Description: "Certified birth certificate"},
		{Code: "SCHOOL_CERTIFICATE", Label: "School Certificate", Description: "School leaving / matriculation certificate"},
		{Code: "PASSPORT", Label: "Passport", Description: "Indian passport"},
		{Code: "VOTER_ID", Label: "Voter ID", Description: "Election Commission voter ID card"},
		{Code: "DRIVING_LICENSE", Label: "Driving License", Description: "Motor vehicle driving license"},
		{Code: "PAN", Label: "PAN Card", Description: "Income Tax PAN card"},
		{Code: "OTHER_STANDARD", Label: "Other Standard Proof", Description: "Other acceptable standard age proof"},
		{Code: "NON_STANDARD_AFFIDAVIT", Label: "Non-Standard (Affidavit)", Description: "Court affidavit as age proof"},
		{Code: "NON_STANDARD_DECLARATION", Label: "Non-Standard (Declaration)", Description: "Self-declaration as age proof"},
	}
	return newLookupResponse(items), nil
}

// GetPaymentModes retrieves payment modes dropdown
// [LU-POL-007] Payment mode dropdown
// Used In: Premium payment
// Source: payment_method_enum from DDL (001_policy_issue_schema.sql) + additional modes from Swagger
// NOTE: Values must stay in sync with the DDL enum definition.
func (h *LookupHandler) GetPaymentModes(sctx *serverRoute.Context, _ struct{}) (*resp.LookupResponse, error) {
	items := []resp.LookupItem{
		{Code: "CASH", Label: "Cash", Description: "Cash payment at post office"},
		{Code: "CHEQUE", Label: "Cheque", Description: "Payment by cheque"},
		{Code: "DD", Label: "Demand Draft", Description: "Payment by demand draft"},
		{Code: "ONLINE", Label: "Online", Description: "Online payment via portal"},
		{Code: "POSB", Label: "POSB", Description: "Deduction from Post Office Savings Bank account"},
		{Code: "NACH", Label: "NACH", Description: "National Automated Clearing House mandate"},
		{Code: "ECS", Label: "ECS", Description: "Electronic Clearing Service"},
		{Code: "STANDING_INSTRUCTION", Label: "Standing Instruction", Description: "Auto-debit standing instruction"},
		{Code: "UPI", Label: "UPI", Description: "Unified Payments Interface"},
		{Code: "NEFT", Label: "NEFT", Description: "National Electronic Fund Transfer"},
		{Code: "RTGS", Label: "RTGS", Description: "Real Time Gross Settlement"},
	}
	return newLookupResponse(items), nil
}

// GetDocumentTypes retrieves document types dropdown
// [LU-POL-008] Document type dropdown
// Used In: Document upload
// Source: document_type_enum from DDL (001_policy_issue_schema.sql)
// NOTE: Values must stay in sync with the DDL enum and domain.DocumentType constants.
func (h *LookupHandler) GetDocumentTypes(sctx *serverRoute.Context, _ struct{}) (*resp.LookupResponse, error) {
	items := []resp.LookupItem{
		{Code: "PROPOSAL_FORM", Label: "Proposal Form", Description: "Signed proposal form"},
		{Code: "DOB_PROOF", Label: "Date of Birth Proof", Description: "Age/DOB proof document"},
		{Code: "ADDRESS_PROOF", Label: "Address Proof", Description: "Communication/permanent address proof"},
		{Code: "PHOTO_ID", Label: "Photo ID", Description: "Government-issued photo identification"},
		{Code: "MEDICAL_REPORT", Label: "Medical Report", Description: "Medical examination report"},
		{Code: "PAYMENT_COPY", Label: "Payment Copy", Description: "Premium payment receipt copy"},
		{Code: "HEALTH_DECLARATION", Label: "Health Declaration", Description: "Self-declaration of health status"},
		{Code: "PHOTO", Label: "Photograph", Description: "Passport-size photograph"},
		{Code: "INCOME_PROOF", Label: "Income Proof", Description: "Salary slip / income certificate"},
		{Code: "EMPLOYMENT_PROOF", Label: "Employment Proof", Description: "Employment certificate / ID card"},
		{Code: "OTHER", Label: "Other", Description: "Other supporting document"},
	}
	return newLookupResponse(items), nil
}

// GetRejectionReasons retrieves rejection reasons dropdown
// [LU-POL-009] Rejection reason dropdown
// Used In: UJ-POL-006 (QR/Approver reject)
func (h *LookupHandler) GetRejectionReasons(sctx *serverRoute.Context, _ struct{}) (*resp.LookupResponse, error) {
	items := []resp.LookupItem{
		{Code: "INCOMPLETE_DOCS", Label: "Incomplete Documents", Description: "Required documents are missing or incomplete"},
		{Code: "INVALID_DETAILS", Label: "Invalid Details", Description: "Personal or policy details are incorrect"},
		{Code: "AGE_INELIGIBLE", Label: "Age Not Eligible", Description: "Insured age is outside eligible range"},
		{Code: "SA_EXCEEDS_LIMIT", Label: "SA Exceeds Limit", Description: "Sum assured exceeds maximum allowed limit"},
		{Code: "MEDICAL_ADVERSE", Label: "Adverse Medical Report", Description: "Medical examination results are unfavorable"},
		{Code: "KYC_MISMATCH", Label: "KYC Mismatch", Description: "KYC details do not match submitted documents"},
		{Code: "DUPLICATE_PROPOSAL", Label: "Duplicate Proposal", Description: "Duplicate proposal exists for same insured"},
		{Code: "FRAUDULENT", Label: "Suspected Fraud", Description: "Proposal suspected of fraudulent activity"},
		{Code: "INCOME_INSUFFICIENT", Label: "Insufficient Income", Description: "Income does not justify the sum assured"},
		{Code: "OCCUPATION_RISK", Label: "Occupation Risk", Description: "Occupation carries high risk as per underwriting guidelines"},
		{Code: "OTHER", Label: "Other", Description: "Other reason for rejection"},
	}
	return newLookupResponse(items), nil
}

// GetCancellationReasons retrieves FLC cancellation reasons dropdown
// [LU-POL-010] FLC cancellation reasons
// Used In: UJ-POL-009 (FLC cancellation)
func (h *LookupHandler) GetCancellationReasons(sctx *serverRoute.Context, _ struct{}) (*resp.LookupResponse, error) {
	items := []resp.LookupItem{
		{Code: "NOT_SATISFIED_TERMS", Label: "Not Satisfied with Terms", Description: "Policyholder not satisfied with policy terms and conditions"},
		{Code: "FOUND_BETTER_OPTION", Label: "Found Better Option", Description: "Found better insurance product from another provider"},
		{Code: "FINANCIAL_DIFFICULTY", Label: "Financial Difficulty", Description: "Unable to continue premium payments due to financial hardship"},
		{Code: "MISREPRESENTATION", Label: "Misrepresentation by Agent", Description: "Policy benefits were misrepresented by agent"},
		{Code: "WRONG_PRODUCT", Label: "Wrong Product Sold", Description: "Product does not match customer's requirement"},
		{Code: "PREMIUM_TOO_HIGH", Label: "Premium Too High", Description: "Premium amount is higher than expected"},
		{Code: "CHANGE_OF_MIND", Label: "Change of Mind", Description: "Customer changed decision about the policy"},
		{Code: "OTHER", Label: "Other", Description: "Other reason for free look cancellation"},
	}
	return newLookupResponse(items), nil
}
