package handler

import (
	"encoding/json"
	"policy-issue-service/core/domain"
	"policy-issue-service/core/port"
)

// AadhaarInitiateRequest for initiating Aadhaar authentication
type AadhaarInitiateRequest struct {
	AadhaarNumber string `json:"aadhaar_number" validate:"required,pattern=^[0-9]{12}$"`
	Purpose       string `json:"purpose" validate:"required"`
}

// func (r *AadhaarInitiateRequest) Validate() error { return nil }

// // AadhaarVerifyOTPRequest for verifying Aadhaar OTP
// type AadhaarVerifyOTPRequest struct {
// 	TransactionID string `json:"transaction_id" validate:"required"`
// 	OTP           string `json:"otp" validate:"required,len=6"`
// }

// func (r *AadhaarVerifyOTPRequest) Validate() error { return nil }

// AadhaarSubmitRequest for submitting Aadhaar proposal
type AadhaarSubmitRequest struct {
	SessionID    string                  `json:"session_id" validate:"required"`
	PolicyType   domain.PolicyType       `json:"policy_type" validate:"required,oneof=PLI RPLI"`
	ProductCode  string                  `json:"product_code" validate:"required"`
	SumAssured   float64                 `json:"sum_assured" validate:"required,gt=0"`
	PolicyTerm   int                     `json:"policy_term" validate:"required,gte=5,lte=50"`
	Frequency    domain.PremiumFrequency `json:"frequency" validate:"required,oneof=MONTHLY QUARTERLY HALF_YEARLY YEARLY"`
	Channel      domain.Channel          `json:"channel" validate:"required,oneof=DIRECT AGENCY WEB MOBILE POS CSC"`
	MobileNumber string                  `json:"mobile_number" validate:"required,pattern=^\\d{10}$"`
	Email        string                  `json:"email" validate:"required,email"`
	PremiumPaid  bool                    `json:"premium_paid"`
	PaymentRef   string                  `json:"payment_ref"`
}

// func (r *AadhaarSubmitRequest) Validate() error { return nil }

// QRApproveRequest for QR approval decision
type QRApproveRequest struct {
	ProposalID int64  `uri:"proposal_id" validate:"required"`
	ReviewerID string `json:"reviewer_id" validate:"required"`
	Comments   string `json:"comments" validate:"omitempty,maxlength=500"`
}

// func (r *QRApproveRequest) Validate() error { return nil }

// QRRejectRequest for QR rejection decision
type QRRejectRequest struct {
	ProposalID int64  `uri:"proposal_id" validate:"required"`
	ReviewerID string `json:"reviewer_id" validate:"required"`
	Comments   string `json:"comments" validate:"required,maxlength=500"`
}

// func (r *QRRejectRequest) Validate() error { return nil }

// QRReturnRequest for QR return decision
type QRReturnRequest struct {
	ProposalID int64  `uri:"proposal_id" validate:"required"`
	ReviewerID string `json:"reviewer_id" validate:"required"`
	Comments   string `json:"comments" validate:"required,maxlength=500"`
}

// func (r *QRReturnRequest) Validate() error { return nil }

// ApproverApproveRequest for Approver approval decision
type ApproverApproveRequest struct {
	ProposalID int64  `uri:"proposal_id" validate:"required"`
	ApproverID string `json:"approver_id" validate:"required"`
	Comments   string `json:"comments" validate:"omitempty,maxlength=500"`
}

//func (r *ApproverApproveRequest) Validate() error { return nil }

// ApproverRejectRequest for Approver rejection decision
type ApproverRejectRequest struct {
	ProposalID int64  `uri:"proposal_id" validate:"required"`
	ApproverID string `json:"approver_id" validate:"required"`
	Comments   string `json:"comments" validate:"required,maxlength=500"`
}

//func (r *ApproverRejectRequest) Validate() error { return nil }

// FLCInitiateRequest for initiating Free Look Cancellation
type FLCInitiateRequest struct {
	ProposalID    int64  `uri:"proposal_id" validate:"required"`
	CustomerID    string `json:"customer_id" validate:"required"`
	RequestReason string `json:"request_reason" validate:"required,maxlength=500"`
	Comments      string `json:"comments" validate:"omitempty,maxlength=1000"`
}

//func (r *FLCInitiateRequest) Validate() error { return nil }

// FLCApproveRequest for approving Free Look Cancellation
type FLCApproveRequest struct {
	ProposalID int64  `uri:"proposal_id" validate:"required"`
	ApproverID string `json:"approver_id" validate:"required"`
	Comments   string `json:"comments" validate:"omitempty,maxlength=1000"`
}

//func (r *FLCApproveRequest) Validate() error { return nil }

// FLCRejectRequest for rejecting Free Look Cancellation
type FLCRejectRequest struct {
	ProposalID   int64  `uri:"proposal_id" validate:"required"`
	ApproverID   string `json:"approver_id" validate:"required"`
	RejectReason string `json:"reject_reason" validate:"required,maxlength=500"`
	Comments     string `json:"comments" validate:"omitempty,maxlength=1000"`
}

//func (r *FLCRejectRequest) Validate() error { return nil }

// FLCQueueRequest for FLC queue query
type FLCQueueRequest struct {
	port.MetadataRequest
	Status string `query:"status" validate:"omitempty,oneof=pending approved rejected"`
}

// func (r *FLCQueueRequest) Validate() error { return nil }

// ProposalIndexingRequest for creating a new proposal
type ProposalIndexingRequest struct {
	PolicyType      string `json:"policy_type" validate:"required,oneof=PLI RPLI"`
	ProductCode     string `json:"product_code" validate:"required"`
	InsurantName    string `json:"insurant_name" validate:"required"`
	EntryPath       string `json:"entry_path" validate:"required,oneof=WITHOUT_AADHAAR WITH_AADHAAR BULK_UPLOAD QUOTE_CONVERSION"`
	POCode          string `json:"po_code" validate:"required"`
	IssueCircle     string `json:"issue_circle" validate:"required"`
	IssueHO         string `json:"issue_ho" validate:"required"`
	IssuePostOffice string `json:"issue_post_office" validate:"required"`
	Channel         string `json:"channel" validate:"required,oneof=DIRECT AGENCY WEB MOBILE POS CSC"`
	// CustomerID              int64    `json:"customer_id" validate:"required"`
	SpouseCustomerID        *int64   `json:"spouse_customer_id,omitempty"`
	Dates                   DateInfo `json:"dates" validate:"required"`
	QuoteRefNumber          string   `json:"quote_ref_number" validate:"omitempty"`
	SumAssured              float64  `json:"sum_assured" validate:"required,gt=0"`
	PolicyTerm              int      `json:"policy_term" validate:"required,gt=0"`
	PremiumPaymentFrequency string   `json:"premium_payment_frequency" validate:"required,oneof=MONTHLY QUARTERLY HALF_YEARLY YEARLY SINGLE"`
	BasePremium             float64  `json:"base_premium" validate:"required,gt=0"`
	TotalPremium            float64  `json:"total_premium" validate:"required,gt=0"`
	GSTAmount               float64  `json:"gst_amount" validate:"omitempty"`
	CreatedBy               int64    `json:"created_by" validate:"required"`
}

// DateInfo contains proposal dates
type DateInfo struct {
	DeclarationDate string `json:"declaration_date" validate:"required,date"`
	ReceiptDate     string `json:"receipt_date" validate:"required,date"`
	IndexingDate    string `json:"indexing_date" validate:"required,date"`
	ProposalDate    string `json:"proposal_date" validate:"required,date"`
}

// func (r *ProposalIndexingRequest) Validate() error { return nil }

// ProposalIDUri for proposal ID in URI
type ProposalIDUri struct {
	ProposalID int64 `uri:"proposal_id" validate:"required"`
}

// func (r *ProposalIDUri) Validate() error { return nil }

// ProposalNumberUri for proposal number in URI
type ProposalNumberUri struct {
	ProposalNumber string `uri:"proposal_number" validate:"required"`
}

// func (r *ProposalNumberUri) Validate() error { return nil }

// FirstPremiumRequest for recording first premium
type FirstPremiumRequest struct {
	ProposalID       int64   `uri:"proposal_id" validate:"required"`
	PaymentMethod    string  `json:"payment_method" validate:"required,oneof=CASH CHEQUE DD ONLINE POSB NACH"`
	Amount           float64 `json:"amount" validate:"required,gt=0"`
	ReceiptNumber    string  `json:"receipt_number" validate:"required"`
	PaymentDate      string  `json:"payment_date" validate:"required,datetime"`
	PaymentReference string  `json:"payment_reference"`
	CollectedBy      int64   `json:"collected_by"`
}

// func (r *FirstPremiumRequest) Validate() error { return nil }

// InsuredDetailsRequest for updating insured details
type InsuredDetailsRequest struct {
	ProposalID    int64   `uri:"proposal_id" validate:"required"`
	CustomerID    int64   `json:"customer_id" validate:"required"`
	Salutation    string  `json:"salutation" validate:"required"`
	FirstName     string  `json:"first_name" validate:"required"`
	MiddleName    string  `json:"middle_name"`
	LastName      string  `json:"last_name" validate:"required"`
	Gender        string  `json:"gender" validate:"required,oneof=MALE FEMALE OTHER ALL"`
	DateOfBirth   string  `json:"date_of_birth" validate:"required,datetime"`
	MaritalStatus string  `json:"marital_status"`
	Occupation    string  `json:"occupation"`
	AnnualIncome  float64 `json:"annual_income"`
	AddressLine1  string  `json:"address_line1" validate:"required"`
	AddressLine2  string  `json:"address_line2"`
	AddressLine3  string  `json:"address_line3"`
	City          string  `json:"city" validate:"required"`
	State         string  `json:"state" validate:"required"`
	PinCode       string  `json:"pin_code" validate:"required,pattern=^[0-9]{6}$"`
	Mobile        string  `json:"mobile" validate:"required,pattern=^[0-9]{10}$"`
	Email         string  `json:"email" validate:"omitempty,email"`
	DataEntryBy   int64   `json:"data_entry_by"`
}

// func (r *InsuredDetailsRequest) Validate() error { return nil }

// NomineesRequest for updating nominees
type NomineesRequest struct {
	ProposalID int64        `uri:"proposal_id" validate:"required"`
	Nominees   []NomineeDTO `json:"nominees" validate:"required,min=1"`
}

// func (r *NomineesRequest) Validate() error { return nil }

// NomineeDTO represents a nominee in request
type NomineeDTO struct {
	Salutation            string  `json:"salutation" validate:"required"`
	FirstName             string  `json:"first_name" validate:"required"`
	MiddleName            string  `json:"middle_name"`
	LastName              string  `json:"last_name" validate:"required"`
	Gender                string  `json:"gender" validate:"required,oneof=M F O"`
	DateOfBirth           string  `json:"date_of_birth" validate:"required,date"`
	IsMinor               bool    `json:"is_minor"`
	Relationship          string  `json:"relationship" validate:"required"`
	SharePercentage       float64 `json:"share_percentage" validate:"required,gt=0,lte=100"`
	AppointeeName         string  `json:"appointee_name"`
	AppointeeRelationship string  `json:"appointee_relationship"`
	NomineeCustomerID     *int64  `json:"nominee_customer_id" validate:"omitempty"`
}

// PolicyDetailsRequest for updating policy details
type PolicyDetailsRequest struct {
	ProposalID            int64               `uri:"proposal_id" validate:"required"`
	SumAssured            float64             `json:"sum_assured" validate:"required,gt=0"`
	PolicyTerm            int                 `json:"policy_term" validate:"required,gt=0"`
	PremiumCeasingAge     int                 `json:"premium_ceasing_age"`
	PremiumFrequency      string              `json:"premium_frequency" validate:"required"`
	SubsequentPaymentMode string              `json:"subsequent_payment_mode" validate:"required"`
	PolicyTakenUnder      string              `json:"policy_taken_under"`
	AgeProofType          string              `json:"age_proof_type" validate:"required"`
	HUFMembers            []HUFMemberRequest  `json:"huf_members,omitempty"`
	MWPATrustee           *MWPATrusteeRequest `json:"mwpa_trustee,omitempty"`
}

type HUFMemberRequest struct {
	IsFinancedHUF                 bool    `json:"is_financed_huf"`
	KartaName                     *string `json:"karta_name,omitempty"`
	HUFPan                        *string `json:"huf_pan,omitempty"`
	LifeAssuredDifferentFromKarta bool    `json:"life_assured_different_from_karta"`
	KartaDifferentReason          *string `json:"karta_different_reason,omitempty"`
	MemberName                    string  `json:"member_name" validate:"required"`
	MemberRelationship            string  `json:"member_relationship" validate:"required"`
	MemberAge                     int     `json:"member_age" validate:"omitempty,gt=0"`
}

type MWPATrusteeRequest struct {
	TrustType    string  `json:"trust_type" validate:"required"`
	TrusteeName  string  `json:"trustee_name" validate:"required"`
	TrusteeDOB   *string `json:"trustee_dob,omitempty"`
	Relationship *string `json:"relationship,omitempty"`
	Address      *string `json:"address,omitempty"`
}

// func (r *PolicyDetailsRequest) Validate() error { return nil }

// AgentDetailsRequest for updating agent details
//
//	type AgentDetailsRequest struct {
//		ProposalID int64  `uri:"proposal_id" validate:"required"`
//		AgentID    string `json:"agent_id" validate:"required"`
//		AgentName  string `json:"agent_name"`
//		AgentCode  string `json:"agent_code"`
//		AgentType  string `json:"agent_type"`
//	}
type AgentDetailsRequest struct {
	ProposalID             int64  `uri:"proposal_id" validate:"required"`
	AgentID                string `json:"agent_id" validate:"required"`
	AgentSalutation        string `json:"agent_salutation"`
	AgentName              string `json:"agent_name"`
	AgentMobile            string `json:"agent_mobile"`
	AgentEmail             string `json:"agent_email"`
	AgentLandline          string `json:"agent_landline"`
	AgentStdCode           string `json:"agent_std_code"`
	ReceivesCorrespondence bool   `json:"receives_correspondence"`
	OpportunityID          string `json:"opportunity_id"`
}

// func (r *AgentDetailsRequest) Validate() error { return nil }

// MedicalInfoDTO represents medical info in request
//
//	type MedicalInfoDTO struct {
//		InsuredIndex      int    `json:"insured_index"`
//		IsSoundHealth     bool   `json:"is_sound_health"`
//		DiseaseTB         bool   `json:"disease_tb"`
//		DiseaseCancer     bool   `json:"disease_cancer"`
//		DiseaseParalysis  bool   `json:"disease_paralysis"`
//		DiseaseInsanity   bool   `json:"disease_insanity"`
//		DiseaseHeartLungs bool   `json:"disease_heart_lungs"`
//		DiseaseKidney     bool   `json:"disease_kidney"`
//		DiseaseBrain      bool   `json:"disease_brain"`
//		DiseaseHIV        bool   `json:"disease_hiv"`
//		DiseaseHepatitisB bool   `json:"disease_hepatitis_b"`
//		DiseaseEpilepsy   bool   `json:"disease_epilepsy"`
//		DiseaseNervous    bool   `json:"disease_nervous"`
//		DiseaseLiver      bool   `json:"disease_liver"`
//		DiseaseLeprosy    bool   `json:"disease_leprosy"`
//		OtherDiseases     bool   `json:"other_diseases"`
//		DiseaseDetails    string `json:"disease_details"`
//	}
type MedicalInfoDTO struct {
	InsuredIndex             int     `json:"insured_index"`
	IsSoundHealth            bool    `json:"is_sound_health"`
	DiseaseTB                bool    `json:"disease_tb"`
	DiseaseCancer            bool    `json:"disease_cancer"`
	DiseaseParalysis         bool    `json:"disease_paralysis"`
	DiseaseInsanity          bool    `json:"disease_insanity"`
	DiseaseHeartLungs        bool    `json:"disease_heart_lungs"`
	DiseaseKidney            bool    `json:"disease_kidney"`
	DiseaseBrain             bool    `json:"disease_brain"`
	DiseaseHIV               bool    `json:"disease_hiv"`
	DiseaseHepatitisB        bool    `json:"disease_hepatitis_b"`
	DiseaseEpilepsy          bool    `json:"disease_epilepsy"`
	DiseaseNervous           bool    `json:"disease_nervous"`
	DiseaseLiver             bool    `json:"disease_liver"`
	DiseaseLeprosy           bool    `json:"disease_leprosy"`
	DiseasePhysicalDeformity bool    `json:"disease_physical_deformity"`
	DiseaseOther             bool    `json:"disease_other"`
	DiseaseDetails           string  `json:"disease_details"`
	FamilyHereditary         bool    `json:"family_hereditary"`
	FamilyHereditaryDetails  string  `json:"family_hereditary_details"`
	MedicalLeave3yr          bool    `json:"medical_leave_3yr"`
	LeaveKind                string  `json:"leave_kind"`
	LeavePeriod              string  `json:"leave_period"`
	LeaveAilment             string  `json:"leave_ailment"`
	HospitalName             string  `json:"hospital_name"`
	HospitalizationFrom      *string `json:"hospitalization_from"`
	HospitalizationTo        *string `json:"hospitalization_to"`
	PhysicalDeformity        bool    `json:"physical_deformity"`
	DeformityType            *string `json:"deformity_type"`
	FamilyDoctorName         string  `json:"family_doctor_name"`
}

// MedicalInfoRequest for updating medical info
type MedicalInfoRequest struct {
	ProposalID  int64            `uri:"proposal_id" validate:"required"`
	MedicalInfo []MedicalInfoDTO `json:"medical_info" validate:"required,min=1"`
}

// func (r *MedicalInfoRequest) Validate() error { return nil }

// DeclarationRequest for updating declaration
type DeclarationRequest struct {
	ProposalID int64 `uri:"proposal_id" validate:"required"`
	IsAgreed   bool  `json:"is_agreed" validate:"required"`
}

// func (r *DeclarationRequest) Validate() error { return nil }

// ProposerDetailsRequest for updating proposer details
type ProposerDetailsRequest struct {
	ProposalID          int64  `uri:"proposal_id" validate:"required"`
	Salutation          string `json:"salutation" validate:"omitempty"`
	FirstName           string `json:"first_name" validate:"omitempty"`
	MiddleName          string `json:"middle_name"`
	LastName            string `json:"last_name" validate:"omitempty"`
	Gender              string `json:"gender" validate:"omitempty,oneof=MALE FEMALE OTHER"`
	DateOfBirth         string `json:"date_of_birth" validate:"omitempty,datetime"`
	MaritalStatus       string `json:"marital_status"`
	Occupation          string `json:"occupation"`
	AnnualIncome        string `json:"annual_income"`
	AddressLine1        string `json:"address_line1"`
	AddressLine2        string `json:"address_line2"`
	AddressLine3        string `json:"address_line3"`
	City                string `json:"city"`
	State               string `json:"state"`
	PinCode             string `json:"pin_code"`
	Mobile              string `json:"mobile" validate:"omitempty,pattern=^[0-9]{10}$"`
	Email               string `json:"email" validate:"omitempty,email"`
	Relationship        string `json:"relationship" validate:"omitempty"`
	RelationshipDetails string `json:"relationship_details"` // Required when relationship is "OTHER"
	IsSameAsInsured     bool   `json:"is_same_as_insured"`   // If true, proposer details same as insured
	DataEntryBy         int64  `json:"data_entry_by"`        // User ID of data entry operator
	CustomerID          string `json:"customer_id,omitempty"`
	CustomerNumber      string `json:"customer_number,omitempty"`
}

// func (r *ProposerDetailsRequest) Validate() error { return nil }

// SubmitForQCRequest for QC submission
type SubmitForQCRequest struct {
	ProposalID  int64 `uri:"proposal_id" validate:"required"`
	DataEntryBy int64 `json:"data_entry_by" validate:"required"`
}

// func (r *SubmitForQCRequest) Validate() error { return nil }

// GetProposalQueueRequest for filtering proposals in queue
type GetProposalQueueRequest struct {
	Status string `query:"status,omitempty" validate:"omitempty"`
	port.MetadataRequest
}
type MetadataRequest struct {
	Page  int `json:"page" query:"page" validate:"gte=1"`
	Limit int `json:"limit" query:"limit" validate:"gte=1,lte=100"`
}

// func (r *GetProposalQueueRequest) Validate() error { return nil }

// FLCStatusUri for FLC status URI params
type FLCStatusUri struct {
	ProposalID int64 `uri:"proposal_id" validate:"required"`
}

// func (r *FLCStatusUri) Validate() error { return nil }

// ========================================================================
// Phase 6: Validation Request DTOs
// ========================================================================

// EligibilityCheckRequest for real-time eligibility check
// [VAL-POL-001] Components: BR-POL-011, BR-POL-012
type EligibilityCheckRequest struct {
	CustomerID          *int64  `json:"customer_id,omitempty"`
	ProductCode         string  `json:"product_code" validate:"required"`
	PolicyType          string  `json:"policy_type" validate:"omitempty,oneof=PLI RPLI"`
	DateOfBirth         string  `json:"date_of_birth" validate:"required,date"`
	SumAssured          float64 `json:"sum_assured" validate:"required,gt=0"`
	PolicyTerm          int     `json:"policy_term" validate:"omitempty,gt=0"`
	AccruedSumAssured   float64 `json:"accrued_sum_assured,omitempty"`
	ExistingAggregateSA float64 `json:"existing_aggregate_sa,omitempty"`
}

// func (r *EligibilityCheckRequest) Validate() error { return nil }

// AadhaarFormatRequest for Aadhaar format validation
// [VAL-POL-002] Components: VR-PI-008
type AadhaarFormatRequest struct {
	AadhaarNumber string `json:"aadhaar_number" validate:"required"`
}

// func (r *AadhaarFormatRequest) Validate() error { return nil }

// PANFormatRequest for PAN format validation
// [VAL-POL-003] Components: VR-PI-009
type PANFormatRequest struct {
	PANNumber string `json:"pan_number" validate:"required"`
}

// func (r *PANFormatRequest) Validate() error { return nil }

// PincodeValidationRequest for pincode-state validation
// [VAL-POL-004] Components: VR-PI-034, VR-PI-035
type PincodeValidationRequest struct {
	Pincode string `json:"pincode" validate:"required"`
	State   string `json:"state,omitempty"`
}

// func (r *PincodeValidationRequest) Validate() error { return nil }

// IFSCValidationRequest for IFSC code validation
// [VAL-POL-005] Components: VR-PI-030
type IFSCValidationRequest struct {
	IFSCCode string `json:"ifsc_code" validate:"required"`
}

// func (r *IFSCValidationRequest) Validate() error { return nil }

// NomineeShareEntry represents a single nominee for share validation
type NomineeShareEntry struct {
	Share                 float64 `json:"share" validate:"required,gt=0,lte=100"`
	IsMinor               bool    `json:"is_minor"`
	AppointeeName         string  `json:"appointee_name,omitempty"`
	AppointeeRelationship string  `json:"appointee_relationship,omitempty"`
}

// NomineeSharesRequest for nominee share validation
// [VAL-POL-006] Components: VR-PI-018
// Validates: share totals 100%, max 3 nominees, appointee rules for minors
type NomineeSharesRequest struct {
	// Shares is kept for backward compatibility (simple share-only validation)
	Shares []float64 `json:"shares,omitempty"`
	// Nominees provides richer validation (share + minor/appointee rules)
	Nominees []NomineeShareEntry `json:"nominees,omitempty"`
}

// func (r *NomineeSharesRequest) Validate() error { return nil }

// DateChainValidationRequest for date sequence validation
// [VAL-POL-007] Components: BR-POL-018
type DateChainValidationRequest struct {
	DeclarationDate string `json:"declaration_date" validate:"required,date"`
	ReceiptDate     string `json:"receipt_date" validate:"required,date"`
	IndexingDate    string `json:"indexing_date" validate:"required,date"`
	ProposalDate    string `json:"proposal_date" validate:"required,date"`
}

// func (r *DateChainValidationRequest) Validate() error { return nil }

// AggregateSACheckRequest for aggregate SA validation
// [VAL-POL-008] Components: INT-POL-002
type AggregateSACheckRequest struct {
	CustomerID  int64   `json:"customer_id" validate:"required"`
	ProductCode string  `json:"product_code,omitempty"`
	PolicyType  string  `json:"policy_type" validate:"omitempty,oneof=PLI RPLI"`
	ProposedSA  float64 `json:"proposed_sa" validate:"required,gt=0"`
}

// func (r *AggregateSACheckRequest) Validate() error { return nil }

// ========================================================================
// Phase 7: Calculation Request DTOs
// ========================================================================

// PremiumCalculationRequest for full premium calculation preview
// [CALC-POL-001] Components: BR-POL-001, BR-POL-002, BR-POL-010
//
//	type PremiumCalculationRequest struct {
//		ProductCode     string  `json:"product_code" validate:"required"`
//		ProductCategory string  `json:"product_category" validate:"required"`
//		AgeAtEntry      int     `json:"age_at_entry" validate:"required,gt=0"`
//		Gender          string  `json:"gender" validate:"required,oneof=MALE FEMALE OTHER"`
//		SumAssured      float64 `json:"sum_assured" validate:"required,gt=0"`
//		PolicyTerm      int     `json:"policy_term" validate:"required,gt=0"`
//		Frequency       string  `json:"frequency" validate:"required,oneof=MONTHLY QUARTERLY HALF_YEARLY YEARLY"`
//		AgeProofType    string  `json:"age_proof_type" validate:"omitempty,oneof=STANDARD NON_STANDARD"`
//		InsuredState    string  `json:"insured_state,omitempty"`
//		ProviderState   string  `json:"provider_state,omitempty"`
//	}
type PremiumCalculationRequest struct {
	ProductCode string `json:"product_code" validate:"required"`
	// ProductCategory   string  `json:"product_category" validate:"required"`
	DateOfBirth       string  `json:"date_of_birth" validate:"required"`
	DateOfCalculation string  `json:"date_of_calculation" validate:"required"`
	Gender            string  `json:"gender" validate:"required"`
	SumAssured        int     `json:"sum_assured" validate:"required,gt=0"`
	Term              int     `json:"term"`
	PremiumCeasingAge int     `json:"premium_ceasing_age"`
	Periodicity       string  `json:"periodicity" validate:"required"`
	AgeProofType      string  `json:"age_proof_type"`
	InsuredState      string  `json:"insured_state"`
	ProviderState     string  `json:"provider_state"`
	ParentDOB         *string `json:"parent_dob,omitempty"` // Required for child plans, optional otherwise
	SpouseDOB         *string `json:"spouse_dob,omitempty"` // Required for spouse cover, optional otherwise
}

// func (r *PremiumCalculationRequest) Validate() error { return nil }

// MaturityCalculationRequest for maturity value estimation
// [CALC-POL-002] Components: FR-POL-001
type MaturityCalculationRequest struct {
	ProductCode string  `json:"product_code" validate:"required"`
	SumAssured  float64 `json:"sum_assured" validate:"required,gt=0"`
	PolicyTerm  int     `json:"policy_term" validate:"required,gt=0"`
}

// func (r *MaturityCalculationRequest) Validate() error { return nil }

// FLCRefundCalculationRequest for FLC refund preview
// [CALC-POL-003] Components: BR-POL-009
type FLCRefundCalculationRequest struct {
	PremiumPaid    float64 `json:"premium_paid" validate:"required,gt=0"`
	DaysOfCoverage int     `json:"days_of_coverage" validate:"required,gt=0"`
	BasePremium    float64 `json:"base_premium" validate:"required,gt=0"`
	StampDuty      float64 `json:"stamp_duty"`
	MedicalFee     float64 `json:"medical_fee"`
}

// func (r *FLCRefundCalculationRequest) Validate() error { return nil }

// GSTCalculationRequest for GST breakdown
// [CALC-POL-004] Components: BR-POL-002
type GSTCalculationRequest struct {
	BaseAmount    float64 `json:"base_amount" validate:"required,gt=0"`
	InsuredState  string  `json:"insured_state" validate:"required"`
	ProviderState string  `json:"provider_state,omitempty"`
}

// func (r *GSTCalculationRequest) Validate() error { return nil }

// ========================================================================
// Phase 8: Document & Status Request DTOs
// ========================================================================

// DocumentUploadRequest for uploading a document to a proposal
// [DOC-POL-002] Components: INT-POL-008 (DMS), VR-PI-021, VR-PI-022, VR-PI-023
type DocumentUploadRequest struct {
	ProposalID   int64  `uri:"proposal_id" validate:"required"`
	DocumentType string `json:"document_type" validate:"required,oneof=PROPOSAL_FORM DOB_PROOF ADDRESS_PROOF PHOTO_ID MEDICAL_REPORT PAYMENT_COPY HEALTH_DECLARATION PHOTO INCOME_PROOF EMPLOYMENT_PROOF OTHER"`
	FileName     string `json:"file_name" validate:"required"`
	MimeType     string `json:"mime_type" validate:"required,oneof=application/pdf image/jpeg image/png image/jpg"`
	FileSize     int64  `json:"file_size" validate:"required,gt=0"`
	DocumentDate string `json:"document_date" validate:"required"` // [VR-PI-023] YYYY-MM-DD, must not be in the future
	Comments     string `json:"comments,omitempty"`
	UploadedBy   int64  `json:"uploaded_by" validate:"required"`
}

// func (r *DocumentUploadRequest) Validate() error { return nil }

// DocumentIDUri for document ID in URI
type DocumentIDUri struct {
	ProposalID int64 `uri:"proposal_id" validate:"required"`
	DocumentID int64 `uri:"document_id" validate:"required"`
}

// func (r *DocumentIDUri) Validate() error { return nil }

// MissingDocumentsQuery for filtering missing documents
// [DOC-POL-005]
type MissingDocumentsQuery struct {
	ProposalID int64  `uri:"proposal_id" validate:"required"`
	Stage      string `query:"stage,omitempty" validate:"omitempty,oneof=QC_REVIEW APPROVAL"`
	Status     string `query:"status,omitempty" validate:"omitempty,oneof=PENDING UPLOADED WAIVED"`
}

// func (r *MissingDocumentsQuery) Validate() error { return nil }

// MissingDocumentCreateRequest for recording a missing document
// [DOC-POL-006]
type MissingDocumentCreateRequest struct {
	ProposalID          int64  `uri:"proposal_id" validate:"required"`
	DocumentType        string `json:"document_type" validate:"required,oneof=PROPOSAL_FORM DOB_PROOF ADDRESS_PROOF PHOTO_ID MEDICAL_REPORT PAYMENT_COPY HEALTH_DECLARATION PHOTO INCOME_PROOF EMPLOYMENT_PROOF OTHER"`
	DocumentDescription string `json:"document_description,omitempty"`
	ReasonMissing       string `json:"reason_missing,omitempty"`
	Stage               string `json:"stage" validate:"required,oneof=QC_REVIEW APPROVAL"`
	Notes               string `json:"notes,omitempty"`
	FollowUpRequired    *bool  `json:"follow_up_required,omitempty"`
	NotedBy             int64  `json:"noted_by" validate:"required"`
}

// func (r *MissingDocumentCreateRequest) Validate() error { return nil }

// MissingDocumentResolveRequest for resolving a missing document
// [DOC-POL-007]
type MissingDocumentResolveRequest struct {
	ProposalID         int64  `uri:"proposal_id" validate:"required"`
	MissingDocID       int64  `uri:"missing_doc_id" validate:"required"`
	Status             string `json:"status" validate:"required,oneof=UPLOADED WAIVED"`
	UploadedDocumentID *int64 `json:"uploaded_document_id,omitempty"`
	WaiverReason       string `json:"waiver_reason,omitempty"`
	ResolutionNotes    string `json:"resolution_notes,omitempty"`
	ResolvedBy         int64  `json:"resolved_by" validate:"required"`
}

// func (r *MissingDocumentResolveRequest) Validate() error { return nil }

// NOTE: PolicyIDUri is already defined in handler/policy.go
// Reused for [STATUS-POL-003] endpoint

// ========================================================================
// Phase 9: Workflow & Bulk Upload Request DTOs
// ========================================================================

// WorkflowIDUri for workflow ID in URI
// [WF-POL-001]
type WorkflowIDUri struct {
	WorkflowID string `uri:"workflow_id" validate:"required"`
}

// func (r *WorkflowIDUri) Validate() error { return nil }

// WorkflowSignalRequest for sending a signal to a workflow
// [WF-POL-002]
// Signal names MUST match constants in workflows/policy_issuance_workflow.go:
//
//	SignalQRDecision="qr-decision", SignalMedicalResult="medical-result",
//	SignalApproverDecision="approver-decision", SignalCPCResubmit="cpc-resubmit"
type WorkflowSignalRequest struct {
	WorkflowID string                 `uri:"workflow_id" validate:"required"`
	SignalName string                 `uri:"signal_name" validate:"required,oneof=qr-decision medical-result approver-decision payment-received cpc-resubmit flc-cancel-request death-notification"`
	Payload    map[string]interface{} `json:"payload"`
}

// func (r *WorkflowSignalRequest) Validate() error { return nil }

// BulkUploadRequest for uploading a bulk proposal file
// [FR-POL-021] Bulk Proposal Upload
// File Formats: CSV (.csv), Excel (.xlsx, .xls)
// Max File Size: 10 MB
type BulkUploadRequest struct {
	FileName     string  `json:"file_name" validate:"required"`
	MimeType     string  `json:"mime_type" validate:"required"`      // file content type for format enforcement
	FileSize     int64   `json:"file_size" validate:"required,gt=0"` // file size in bytes
	TotalRows    int     `json:"total_rows" validate:"required,gt=0,lte=1000"`
	PaymentType  string  `json:"payment_type" validate:"required,oneof=INDIVIDUAL COMBINED_CHEQUE"`
	ChequeAmount float64 `json:"cheque_amount,omitempty"`
	UploadedBy   int64   `json:"uploaded_by" validate:"required"`
}

// func (r *BulkUploadRequest) Validate() error { return nil }

// BatchIDUri for batch ID in URI
type BatchIDUri struct {
	BatchID int64 `uri:"batch_id" validate:"required"`
}

// func (r *BatchIDUri) Validate() error { return nil }
type AadhaarVerifyOTPRequest struct {
	TransactionID string `json:"transaction_id" validate:"required"`
	OTP           string `json:"otp" validate:"required,len=6"`
}

// GetProductsRequest for filtering products
type GetProductsRequest struct {
	PolicyType string `form:"policy_type" validate:"omitempty,oneof=PLI RPLI"`
	IsActive   *bool  `form:"is_active,omitempty" query:"is_active"`
}

// QuoteCalculateRequest for calculating premium quote
// [BR-POL-001] Base Premium Calculation
// [BR-POL-002] GST Calculation
// [BR-POL-003] Rebate Calculation
type QuoteCalculateRequest struct {
	// PolicyType      string  `json:"policy_type" validate:"required,oneof=PLI RPLI"`
	ProductCode string `uri:"product_code" validate:"required"`
	// ProductCategory   string  `json:"product_category" validate:"required"` // Optional, can be used for more granular pricing rules
	DateOfBirth       string  `json:"date_of_birth" validate:"required"`
	DateOfCalculation string  `json:"date_of_calculation" validate:"required"`
	Gender            string  `json:"gender" validate:"required"`
	Periodicity       string  `json:"periodicity" validate:"required"`
	Term              int     `json:"term" validate:"omitempty"`
	PremiumCeasingAge int     `json:"premium_ceasing_age" validate:"omitempty"`
	SumAssured        int     `json:"sum_assured" validate:"required,gt=0"`
	ChildDOB          *string `json:"child_dob,omitempty"`  // Required for child plans, optional otherwise
	SpouseDOB         *string `json:"spouse_dob,omitempty"` // Required for spouse cover, optional otherwise

}

// type QuoteCalculateRequest struct {
// 	PolicyType  string  `json:"policy_type" validate:"required,oneof=PLI RPLI"`
// 	ProductCode string  `json:"product_code" validate:"required"`
// 	DateOfBirth string  `json:"date_of_birth" validate:"required,datetime"`
// 	Gender      string  `json:"gender" validate:"required,oneof=MALE FEMALE OTHER ALL"`
// 	SumAssured  float64 `json:"sum_assured" validate:"required,gt=0"`
// 	PolicyTerm  int     `json:"policy_term" validate:"required,gte=5,lte=50"`
// 	Frequency   string  `json:"frequency" validate:"required,oneof=MONTHLY QUARTERLY HALF_YEARLY YEARLY Monthly"`
// 	StateCode   string  `json:"state_code" validate:"required"`
// }

// StartDataEntryRequest for starting data entry on an indexed proposal
type StartDataEntryRequest struct {
	ProposalID int64  `uri:"proposal_id" validate:"required"`
	AssignedTo int64  `json:"assigned_to" validate:"required"` // CPC user ID
	Comments   string `json:"comments" validate:"omitempty,maxlength=500"`
}

// QuoteCreateRequest for saving a calculated quote
// [FR-POL-001] New Business Quote Generation
type QuoteCreateRequest struct {
	CalculationID string       `json:"calculation_id" validate:"required,uuid"`
	PolicyType    string       `json:"policy_type" validate:"required,oneof=PLI RPLI"`
	ProductCode   string       `json:"product_code" validate:"required"`
	Proposer      ProposerInfo `json:"proposer" validate:"required"`
	Coverage      CoverageInfo `json:"coverage" validate:"required"`
	Premium       PremiumInfo  `json:"premium" validate:"required"`
	Channel       string       `json:"channel" validate:"required,oneof=DIRECT AGENCY WEB MOBILE POS CSC"`
	CreatedBy     int64        `json:"created_by" validate:"required"`
	GeneratePDF   bool         `json:"generate_pdf"`
	SendEmail     bool         `json:"send_email"`
}
type QuoteGenerateRequest struct {
	ProductCode string `uri:"product_code" validate:"required"`
	// ProductCategory   string       `json:"product_category" validate:"required"`
	DateOfCalculation string       `json:"date_of_calculation" validate:"required"`
	Periodicity       string       `json:"periodicity" validate:"required"`
	Term              int          `json:"term,omitempty"`
	PremiumCeasingAge int          `json:"premium_ceasing_age,omitempty"`
	SumAssured        int          `json:"sum_assured" validate:"required,gt=0"`
	Proposer          ProposerInfo `json:"proposer" validate:"required"`
	ChildDOB          *string      `json:"child_dob,omitempty"`
	SpouseDOB         *string      `json:"spouse_dob,omitempty"`
	PolicyType        string       `json:"policy_type" validate:"required,oneof=PLI RPLI"`
	Channel           string       `json:"channel" validate:"required,oneof=DIRECT AGENCY WEB MOBILE POS CSC"`
	CreatedBy         int64        `json:"created_by" validate:"required"`
	GeneratePDF       bool         `json:"generate_pdf"`
	SendEmail         bool         `json:"send_email"`
}
type GetQuoteRequestParams struct {
	QuoteID int64 `uri:"quote_id" validate:"required"`
}

// PremiumInfo contains calculated premium details
type PremiumInfo struct {
	BasePremium  float64 `json:"base_premium" validate:"required,gt=0"`
	Rebate       float64 `json:"rebate"`
	NetPremium   float64 `json:"net_premium" validate:"required,gt=0"`
	CGST         float64 `json:"cgst"`
	SGST         float64 `json:"sgst"`
	TotalGST     float64 `json:"total_gst"`
	TotalPayable float64 `json:"total_payable" validate:"required,gt=0"`
}

// ProposerInfo contains proposer personal details
type ProposerInfo struct {
	Name      string `json:"name" validate:"required,max=200"`
	DOB       string `json:"date_of_birth" validate:"required,date"`
	Gender    string `json:"gender" validate:"required,oneof=MALE FEMALE OTHER"`
	Mobile    string `json:"mobile" validate:"required,pattern=^[0-9]{10}$"`
	Email     string `json:"email" validate:"omitempty,email,max=100"`
	Category  string `json:"category,omitempty"`
	Location  string `json:"location,omitempty"`
	StateCode string `json:"state_code"`
}

// CoverageInfo contains policy coverage details
type CoverageInfo struct {
	SumAssured       float64 `json:"sum_assured" validate:"required,gt=0"`
	PolicyTerm       int     `json:"policy_term" validate:"required,gte=5,lte=50"`
	PaymentFrequency string  `json:"payment_frequency" validate:"required,oneof=MONTHLY QUARTERLY HALF_YEARLY YEARLY"`
}

// QuoteConvertRequest for converting quote to proposal
type QuoteConvertRequest struct {
	QuoteRefNumber string `uri:"quote_ref_number" validate:"required"`
	CustomerID     int64  `json:"customer_id" validate:"required"`
	CreatedBy      int64  `json:"created_by" validate:"required"`
}

// GetCustomerRequest for fetching customer details
type CustomerGetInput struct {
	LookupType      string   `json:"lookup_type"`
	CustomerID      string   `json:"customer_id" validate:"required"`
	CustomerNumber  string   `json:"customer_number,omitempty"`
	IncludeSections []string `json:"include_sections"`
}

type CustomerCreateBatchInput struct {
	CustomerID int64               `json:"customer_id"`
	Input      CustomerCreateInput `json:"input"`
}

type CustomerCreateInput struct {
	IdempotencyKey string                `json:"idempotency_key" validate:"required,min=1,max=100"`
	ProductType    string                `json:"product_type" validate:"required,oneof=PLI RPLI"`
	Identity       CustomerIdentityInput `json:"identity" validate:"required"`
	Documents      *DocumentRefsInput    `json:"documents,omitempty"`
	Addresses      []AddressInput        `json:"addresses,omitempty"`
	Contacts       []ContactInput        `json:"contacts,omitempty"`
	Employment     *EmploymentInput      `json:"employment,omitempty"`
	BankAccounts   []BankAccountInput    `json:"bank_accounts,omitempty" validate:"dive"`
	AdditionalInfo *AdditionalInfoInput  `json:"additional_info,omitempty"`
	Preferences    []PreferenceInput     `json:"preferences,omitempty" validate:"dive"`
}

type CustomerIdentityInput struct {
	FirstName           string `json:"first_name" validate:"required,min=1,max=100"`
	MiddleName          string `json:"middle_name,omitempty" validate:"omitempty,max=100"`
	LastName            string `json:"last_name" validate:"required,min=1,max=100"`
	DOB                 string `json:"dob" validate:"required,datetime=2006-01-02"`
	Gender              string `json:"gender" validate:"required,oneof=MALE FEMALE OTHER"`
	Nationality         string `json:"nationality,omitempty" validate:"omitempty,min=2,max=50"`
	CountryOfResidence  string `json:"country_of_residence,omitempty" validate:"omitempty,min=2,max=50"`
	MaritalStatus       string `json:"marital_status,omitempty" validate:"omitempty,oneof=SINGLE MARRIED DIVORCED WIDOWED"`
	Salutation          string `json:"salutation,omitempty" validate:"omitempty,oneof=MR MRS MS DR"`
	FatherName          string `json:"father_name,omitempty" validate:"omitempty,max=200"`
	HusbandName         string `json:"husband_name,omitempty" validate:"omitempty,max=200"`
	InsuredProposerSame bool   `json:"insured_proposer_same" validate:"required"`
}

type DocumentRefsInput struct {
	AadhaarMasked string `json:"aadhaar_masked,omitempty" validate:"omitempty,len=12"`
	PANNumber     string `json:"pan_number,omitempty" validate:"omitempty,len=10"`
	EIAID         string `json:"eia_id,omitempty" validate:"omitempty,min=1,max=50"`
	CKYCNumber    string `json:"ckyc_number,omitempty" validate:"omitempty,min=1,max=50"`
	GCIFId        string `json:"gcif_id,omitempty" validate:"omitempty,min=1,max=50"`
}

type AddressInput struct {
	AddressType  string `json:"address_type" validate:"required,oneof=COMMUNICATION PERMANENT EMPLOYER"`
	Line1        string `json:"address_1" validate:"required,min=1,max=200"`
	Line2        string `json:"address_2,omitempty" validate:"omitempty,max=200"`
	Village      string `json:"village,omitempty" validate:"omitempty,max=100"`
	Taluka       string `json:"taluka,omitempty" validate:"omitempty,max=100"`
	Mandal       string `json:"mandal,omitempty" validate:"omitempty,max=100"`
	City         string `json:"city" validate:"required,min=1,max=100"`
	District     string `json:"district" validate:"required,min=1,max=100"`
	State        string `json:"state" validate:"required,min=1,max=50"`
	Country      string `json:"country" validate:"required,min=2,max=50"`
	PinCode      string `json:"pin_code" validate:"required,len=6"`
	VersionID    int    `json:"version_id,omitempty"`
	ChangeReason string `json:"change_reason,omitempty" validate:"omitempty,max=500"`
	ApprovedBy   string `json:"approved_by,omitempty" validate:"omitempty,max=100"`
}

type ContactInput struct {
	ContactType  string `json:"contact_type" validate:"required,oneof=MOBILE EMAIL LANDLINE"`
	ContactValue string `json:"contact_value" validate:"required,min=1,max=100"`
	IsPrimary    bool   `json:"is_primary"`
}

type EmploymentInput struct {
	Occupation          string        `json:"occupation" validate:"required,min=1,max=100"`
	PAODDOCode          string        `json:"pao_ddo_code,omitempty" validate:"omitempty,min=1,max=50"`
	Organization        string        `json:"organization,omitempty" validate:"omitempty,max=200"`
	Designation         string        `json:"designation,omitempty" validate:"omitempty,max=100"`
	DateOfEntry         string        `json:"date_of_entry,omitempty" validate:"omitempty,datetime=2006-01-02"`
	SuperiorDesignation string        `json:"superior_designation,omitempty" validate:"omitempty,max=100"`
	MonthlyIncome       float64       `json:"monthly_income,omitempty" validate:"omitempty,min=0"`
	Qualification       string        `json:"qualification,omitempty" validate:"omitempty,max=100"`
	EmployerAddress     *AddressInput `json:"employer_address,omitempty"`
}

type BankAccountInput struct {
	AccountNumber string `json:"account_number" validate:"required,min=9,max=18"`
	IFSCCode      string `json:"ifsc_code" validate:"required,len=11"`
	BankName      string `json:"bank_name" validate:"required,min=1,max=100"`
	BranchName    string `json:"branch_name,omitempty" validate:"omitempty,max=100"`
	AccountType   string `json:"account_type" validate:"required,oneof=SAVINGS CURRENT SALARY NRE NRO"`
	Purpose       string `json:"purpose" validate:"required,oneof=PREMIUM_PAYMENT CLAIM_SETTLEMENT REFUND"`
	IsPrimary     bool   `json:"is_primary" validate:"required"`
}

type AdditionalInfoInput struct {
	MarksOfId1              string          `json:"marks_of_id_1,omitempty" validate:"omitempty,max=100"`
	MarksOfId2              string          `json:"marks_of_id_2,omitempty" validate:"omitempty,max=100"`
	NumChildren             *int            `json:"num_children,omitempty" validate:"omitempty,min=0,max=20"`
	DateLastDelivery        string          `json:"date_last_delivery,omitempty" validate:"omitempty,datetime=2006-01-02"`
	ExpectedMonthOfDelivery string          `json:"expected_month_of_delivery,omitempty" validate:"omitempty,max=20"`
	MothersName             string          `json:"mothers_name,omitempty" validate:"omitempty,max=100"`
	ParentsPolicyNumber     string          `json:"parents_policy_number,omitempty" validate:"omitempty,max=50"`
	AgeProofType            string          `json:"age_proof_type,omitempty" validate:"omitempty,oneof=BIRTH_CERTIFICATE SCHOOL_CERTIFICATE PASSPORT AADHAAR"`
	PolicyTakenUnder        string          `json:"policy_taken_under,omitempty" validate:"omitempty,oneof=HUF MWPA"`
	HUFDetails              json.RawMessage `db:"huf_details"`
	MWPADetails             json.RawMessage `db:"mwpa_details"`
}

type PreferenceInput struct {
	Key   string `json:"key" validate:"required,min=1,max=50"`
	Value string `json:"value" validate:"required,min=1,max=200"`
}

type GetTermOrPremiumCeasingAgeRequest struct {
	ProductCode       string  `uri:"product_code" validate:"required"`
	DateOfBirth       string  `form:"date_of_birth" validate:"required,datetime"`
	DateOfCalculation string  `form:"date_of_calculation" validate:"required"`
	SpouseDOB         *string `form:"spouse_dob,omitempty"`
}

type GetQuoteRefRequestParams struct {
	QuoteRefNumber string `uri:"quote_ref_number" validate:"required"`
}

type GetProposalSectionRequest struct {
	Section        string `form:"section" validate:"required"`
	ProposalNumber string `form:"proposal_number,omitempty"`
	QuoteRefNumber string `form:"quote_ref_number,omitempty"`
}

type DedupInput struct {
	FirstName     string `json:"first_name" validate:"omitempty"`
	LastName      string `json:"last_name" validate:"omitempty"`
	DOB           string `json:"dob" validate:"omitempty,datetime=2006-01-02"`
	AadhaarMasked string `json:"aadhaar_masked,omitempty"`
	PANNumber     string `json:"pan_number,omitempty"`
}

type CustomerAddressInput struct {
	Action       string         `json:"action" validate:"required"`
	CustomerID   string         `json:"customer_id" validate:"required"`
	Address      []AddressInput `json:"address,omitempty"`
	AddressID    int64          `json:"address_id,omitempty"`
	AddressType  string         `json:"address_type,omitempty"`
	PinCode      string         `json:"pin_code,omitempty"`
	State        string         `json:"state,omitempty"`
	Reason       string         `json:"reason,omitempty"`
	ChangeReason string         `json:"change_reason,omitempty"`
	ApprovedBy   string         `json:"approved_by,omitempty"`
	ActiveOnly   bool           `json:"active_only,omitempty"`
}

type CustomerContactInput struct {
	Action      string         `json:"action" validate:"required"`
	CustomerID  string         `json:"customer_id" validate:"required"`
	Contacts    []ContactInput `json:"contacts,omitempty"`
	ContactID   int64          `json:"contact_id,omitempty"`
	Reason      string         `json:"reason,omitempty"`
	ApprovedBy  string         `json:"approved_by,omitempty"`
	ActiveOnly  bool           `json:"active_only,omitempty"`
	ContactType string         `json:"contact_type,omitempty"`
}

type CustomerEmploymentInput struct {
	Action     string           `json:"action"`
	CustomerID string           `json:"customer_id"`
	Employment *EmploymentInput `json:"employment,omitempty"`
	Reason     string           `json:"reason,omitempty"`
	ApprovedBy string           `json:"approved_by,omitempty"`
}
