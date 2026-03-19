package response

import (
	"policy-issue-service/core/domain"
	"policy-issue-service/core/port"
)

type CustomerDetailResponse struct {
	port.StatusCodeAndMessage
	CustomerGetOutput
}

type CustomerGetInput struct {
	LookupType      string   `json:"lookup_type"`
	CustomerID      string   `json:"customer_id,omitempty"`
	CustomerNumber  string   `json:"customer_number,omitempty"`
	IncludeSections []string `json:"include_sections"`
}

type CustomerGetOutput struct {
	CustomerID     string              `json:"customer_id"`
	CustomerNumber string              `json:"customer_number"`
	Identity       CustomerIdentityOut `json:"identity"`
	// Documents          DocumentRefsOut        `json:"documents"`
	Flags              CustomerFlagsOut       `json:"flags"`
	Addresses          []AddressOut           `json:"addresses,omitempty"`
	Contacts           []ContactOut           `json:"contacts,omitempty"`
	Employment         *EmploymentOut         `json:"employment,omitempty"`
	BankAccounts       []BankAccountOut       `json:"bank_accounts,omitempty"`
	PaymentInstruments []PaymentInstrumentOut `json:"payment_instruments,omitempty"`
	AdditionalInfo     *AdditionalInfoOut     `json:"additional_info,omitempty"`
	Preferences        map[string]string      `json:"preferences,omitempty"`
	Roles              []RoleOut              `json:"roles,omitempty"`
	Status             string                 `json:"status"`
	Version            int                    `json:"version"`
	CreatedAt          string                 `json:"created_at"`
	UpdatedAt          string                 `json:"updated_at"`
}

type CustomerIdentityOut struct {
	FirstName           string `json:"first_name"`
	MiddleName          string `json:"middle_name,omitempty"`
	LastName            string `json:"last_name"`
	DOB                 string `json:"dob"`
	Gender              string `json:"gender"`
	Nationality         string `json:"nationality"`
	CountryOfResidence  string `json:"country_of_residence"`
	MaritalStatus       string `json:"marital_status,omitempty"`
	Salutation          string `json:"salutation,omitempty"`
	FatherName          string `json:"father_name,omitempty"`
	HusbandName         string `json:"husband_name,omitempty"`
	InsuredProposerSame bool   `json:"insured_proposer_same"`
	AadhaarMasked       string `json:"aadhaar_masked,omitempty"`
	PANNumber           string `json:"pan_number,omitempty"`
	EIAID               string `json:"eia_id,omitempty"`
	CKYCNumber          string `json:"ckyc_number,omitempty"`
	GCIFId              string `json:"gcif_id,omitempty"`
}

// type DocumentRefsOut struct {
// 	AadhaarMasked string `json:"aadhaar_masked,omitempty"`
// 	PANNumber     string `json:"pan_number,omitempty"`
// 	EIAID         string `json:"eia_id,omitempty"`
// 	CKYCNumber    string `json:"ckyc_number,omitempty"`
// 	GCIFId        string `json:"gcif_id,omitempty"`
// }

type CustomerFlagsOut struct {
	KYCStatus string `json:"kyc_status"`
	KYCLevel  string `json:"kyc_level,omitempty"`
	AMLStatus string `json:"aml_status"`
	FraudFlag string `json:"fraud_flag"`
}

type AddressOut struct {
	ID            string `json:"address_id"`
	AddressType   string `json:"address_type"`
	Line1         string `json:"address_1"`
	Line2         string `json:"address_2,omitempty"`
	Village       string `json:"village,omitempty"`
	Taluka        string `json:"taluka,omitempty"`
	City          string `json:"city"`
	District      string `json:"district"`
	State         string `json:"state"`
	Country       string `json:"country"`
	PinCode       string `json:"pin_code"`
	Version       int    `json:"version"`
	IsActive      bool   `json:"is_active"`
	EffectiveFrom string `json:"effective_from"`
	EffectiveTo   string `json:"effective_to,omitempty"`
}

type ContactOut struct {
	ID           string `json:"contact_id"`
	ContactType  string `json:"contact_type"`
	ContactValue string `json:"contact_value"`
	IsPrimary    bool   `json:"is_primary"`
	IsVerified   bool   `json:"is_verified"`
	IsActive     bool   `json:"is_active"`
}

type EmploymentOut struct {
	ID                  string  `json:"employment_id"`
	Occupation          string  `json:"occupation"`
	PAODDOCode          string  `json:"pao_ddo_code,omitempty"`
	Organization        string  `json:"organization,omitempty"`
	Designation         string  `json:"designation,omitempty"`
	DateOfEntry         string  `json:"date_of_entry,omitempty"`
	SuperiorDesignation string  `json:"superior_designation,omitempty"`
	MonthlyIncome       float64 `json:"monthly_income,omitempty"`
	Qualification       string  `json:"qualification,omitempty"`
	IsActive            bool    `json:"is_active"`
}

type BankAccountOut struct {
	ID                  string `json:"bank_id"`
	AccountNumberMasked string `json:"account_number_masked"`
	IFSCCode            string `json:"ifsc_code"`
	BankName            string `json:"bank_name"`
	BranchName          string `json:"branch_name,omitempty"`
	AccountType         string `json:"account_type"`
	Purpose             string `json:"purpose"`
	IsPrimary           bool   `json:"is_primary"`
	IsActive            bool   `json:"is_active"`
}

type PaymentInstrumentOut struct {
	ID               string  `json:"payment_id"`
	BankAccountID    string  `json:"bank_account_id"`
	InstrumentType   string  `json:"instrument_type"`
	MandateReference string  `json:"mandate_reference"`
	MaxAmount        float64 `json:"max_amount"`
	Frequency        string  `json:"frequency"`
	StartDate        string  `json:"start_date"`
	EndDate          string  `json:"end_date,omitempty"`
	Status           string  `json:"status"`
}

type AdditionalInfoOut struct {
	MarksOfId1              string `json:"marks_of_id_1,omitempty"`
	MarksOfId2              string `json:"marks_of_id_2,omitempty"`
	NumChildren             *int   `json:"num_children,omitempty"`
	DateLastDelivery        string `json:"date_last_delivery,omitempty"`
	ExpectedMonthOfDelivery string `json:"expected_month_of_delivery,omitempty"`
	MothersName             string `json:"mothers_name,omitempty"`
	ParentsPolicyNumber     string `json:"parents_policy_number,omitempty"`
	AgeProofType            string `json:"age_proof_type,omitempty"`
	PolicyTakenUnder        string `json:"policy_taken_under,omitempty"`
	HUFDetails              string `json:"huf_details,omitempty"`
	MWPADetails             string `json:"mwpa_details,omitempty"`
}

type RoleOut struct {
	ID              string  `json:"id"`
	CustomerID      string  `json:"customer_id"`
	RoleType        string  `json:"role_type"`
	PolicyID        string  `json:"policy_id"`
	EffectiveFrom   string  `json:"effective_from"`
	EffectiveTo     string  `json:"effective_to,omitempty"`
	IsActive        bool    `json:"is_active"`
	Relationship    string  `json:"relationship,omitempty"`
	SharePercentage float64 `json:"share_percentage,omitempty"`
}

type CustomerCreateOutput struct {
	CustomerID int64 `json:"customer_id"`
	// CustomerNumber string       `json:"customer_number"`
	Status        string       `json:"status"`
	DedupWarnings []DedupMatch `json:"dedup_warnings,omitempty"`
	CreatedAt     string       `json:"created_at"`
}

type CustomerDedupResponse struct {
	port.StatusCodeAndMessage
	DedupOutput
}
type DedupMatch struct {
	ExistingCustomerID     string   `json:"existing_customer_id"`
	ExistingCustomerNumber string   `json:"existing_customer_number"`
	Score                  int      `json:"score"`
	MatchedFields          []string `json:"matched_fields"`
	Recommendation         string   `json:"recommendation"`
}

type DedupOutput struct {
	MatchesFound   int          `json:"matches_found"`
	HighestScore   int          `json:"highest_score"`
	Recommendation string       `json:"recommendation"`
	Matches        []DedupMatch `json:"matches"`
}

type CustomerAddressResponseDTO struct {
	AddressID   int64   `json:"address_id"`
	AddressType string  `json:"address_type"`
	Line1       string  `json:"line1"`
	Line2       *string `json:"line2,omitempty"`
	Village     *string `json:"village,omitempty"`
	Taluka      *string `json:"taluka,omitempty"`
	City        string  `json:"city"`
	District    string  `json:"district"`
	State       string  `json:"state"`
	Country     string  `json:"country"`
	PinCode     string  `json:"pin_code"`
	IsActive    bool    `json:"is_active"`
}

func MapAddressToDTO(addr domain.CustomerAddress) CustomerAddressResponseDTO {
	return CustomerAddressResponseDTO{
		AddressID:   addr.AddressID,
		AddressType: addr.AddressType,
		Line1:       addr.Line1,
		Line2:       addr.Line2,
		Village:     addr.Village,
		Taluka:      addr.Taluka,
		City:        addr.City,
		District:    addr.District,
		State:       addr.State,
		Country:     addr.Country,
		PinCode:     addr.PinCode,
		IsActive:    addr.IsActive,
	}
}

type CustomerAddressOutput struct {
	CustomerID string                       `json:"customer_id"`
	AddressID  []int64                      `json:"address_id,omitempty"`
	Address    *CustomerAddressResponseDTO  `json:"address,omitempty"`
	Addresses  []CustomerAddressResponseDTO `json:"addresses,omitempty"`
	Status     string                       `json:"status"`
}

type CustomerAddressResponse struct {
	StatusCodeAndMessage port.StatusCodeAndMessage `json:",inline"`
	CustomerAddressOutput
}

type CustomerContactOutput struct {
	CustomerID string                   `json:"customer_id"`
	Contact    *domain.CustomerContact  `json:"contact,omitempty"`
	Contacts   []domain.CustomerContact `json:"contacts,omitempty"`
	Status     string                   `json:"status"`
}

type CustomerContactResponse struct {
	port.StatusCodeAndMessage
	CustomerContactOutput CustomerContactOutput `json:"data"`
}

type EmploymentOutput struct {
	CustomerID string                    `json:"customer_id"`
	Employment *domain.CustomerEmployment `json:"employment,omitempty"`
	Status     string                    `json:"status"`
}

type CustomerEmploymentResponse struct{
	port.StatusCodeAndMessage
	EmploymentOutput EmploymentOutput `json:"data"`
}