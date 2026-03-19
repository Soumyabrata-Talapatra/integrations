package domain

import (
	"context"
	"errors"
	"time"

	log "gitlab.cept.gov.in/it-2.0-common/n-api-log"
)

// Common errors
var (
	ErrInvalidDateSequence = errors.New("invalid date sequence: declaration_date <= receipt_date <= indexing_date <= proposal_date")
)

// IsValidStatusTransition checks if a status transition is valid
// This is a helper function for cases where you don't have a full Proposal object
// [BR-POL-015] Proposal State Machine
func IsValidStatusTransition(fromStatus ProposalStatus, toStatus ProposalStatus) bool {
	// Create a temporary Proposal to use the existing validation logic
	tempProposal := &Proposal{Status: fromStatus}
	return tempProposal.CanTransitionTo(toStatus)
}

// ProposalStatus represents the status of a proposal
type ProposalStatus string

const (
	ProposalStatusDraft           ProposalStatus = "DRAFT"
	ProposalStatusIndexed         ProposalStatus = "INDEXED"
	ProposalStatusDataEntry       ProposalStatus = "DATA_ENTRY"
	ProposalStatusQCPending       ProposalStatus = "QC_PENDING"
	ProposalStatusQCApproved      ProposalStatus = "QC_APPROVED"
	ProposalStatusQCRejected      ProposalStatus = "QC_REJECTED"
	ProposalStatusQCReturned      ProposalStatus = "QC_RETURNED"
	ProposalStatusPendingMedical  ProposalStatus = "PENDING_MEDICAL"
	ProposalStatusMedicalApproved ProposalStatus = "MEDICAL_APPROVED"
	ProposalStatusMedicalRejected ProposalStatus = "MEDICAL_REJECTED"
	ProposalStatusApprovalPending ProposalStatus = "APPROVAL_PENDING"
	ProposalStatusApproved        ProposalStatus = "APPROVED"
	ProposalStatusRejected        ProposalStatus = "REJECTED"
	ProposalStatusIssued          ProposalStatus = "ISSUED"
	ProposalStatusDispatched      ProposalStatus = "DISPATCHED"
	ProposalStatusFreeLookActive  ProposalStatus = "FREE_LOOK_ACTIVE"
	ProposalStatusActive          ProposalStatus = "ACTIVE"
	ProposalStatusFLCCancelled    ProposalStatus = "FLC_CANCELLED"
	ProposalStatusCancelledDeath  ProposalStatus = "CANCELLED_DEATH"
)

// EntryPath represents the entry path for proposal creation
type EntryPath string

const (
	EntryPathWithoutAadhaar  EntryPath = "WITHOUT_AADHAAR"
	EntryPathWithAadhaar     EntryPath = "WITH_AADHAAR"
	EntryPathBulkUpload      EntryPath = "BULK_UPLOAD"
	EntryPathQuoteConversion EntryPath = "QUOTE_CONVERSION"
)

// PremiumPayerType represents who pays the premium
type PremiumPayerType string

const (
	PremiumPayerSelf       PremiumPayerType = "SELF"
	PremiumPayerEmployer   PremiumPayerType = "EMPLOYER"
	PremiumPayerDDO        PremiumPayerType = "DDO"
	PremiumPayerThirdParty PremiumPayerType = "THIRD_PARTY"
)

// AgeProofType represents the type of age proof provided
type AgeProofType string

const (
	AgeProofAadhaar              AgeProofType = "AADHAAR"
	AgeProofBirthCertificate     AgeProofType = "BIRTH_CERTIFICATE"
	AgeProofSchoolCertificate    AgeProofType = "SCHOOL_CERTIFICATE"
	AgeProofPassport             AgeProofType = "PASSPORT"
	AgeProofVoterID              AgeProofType = "VOTER_ID"
	AgeProofDrivingLicense       AgeProofType = "DRIVING_LICENSE"
	AgeProofPAN                  AgeProofType = "PAN"
	AgeProofOtherStandard        AgeProofType = "OTHER_STANDARD"
	AgeProofNonStandardAffidavit AgeProofType = "NON_STANDARD_AFFIDAVIT"
	AgeProofNonStandardDecl      AgeProofType = "NON_STANDARD_DECLARATION"
)

// PolicyTakenUnder represents how the policy is taken
type PolicyTakenUnder string

const (
	PolicyTakenUnderHUF   PolicyTakenUnder = "HUF"
	PolicyTakenUnderMWPA  PolicyTakenUnder = "MWPA"
	PolicyTakenUnderOther PolicyTakenUnder = "OTHER"
)

// Proposal represents the core proposal entity
type Proposal struct {
	ProposalID              int64            `db:"proposal_id" json:"proposal_id"`
	InsurantName            string           `db:"insurant_name" json:"insurant_name"`
	ProposalNumber          string           `db:"proposal_number" json:"proposal_number"`
	QuoteRefNumber          *string          `db:"quote_ref_number" json:"quote_ref_number,omitempty"`
	CustomerID              *int64           `db:"customer_id" json:"customer_id,omitempty"`
	SpouseCustomerID        *int64           `db:"spouse_customer_id" json:"spouse_customer_id,omitempty"`
	ProposerCustomerID      *int64           `db:"proposer_customer_id" json:"proposer_customer_id,omitempty"`
	IsProposerSameAsInsured bool             `db:"is_proposer_same_as_insured" json:"is_proposer_same_as_insured"`
	PremiumPayerType        PremiumPayerType `db:"premium_payer_type" json:"premium_payer_type"`
	PayerCustomerID         *int64           `db:"payer_customer_id" json:"payer_customer_id,omitempty"`
	ProductCode             string           `db:"product_code" json:"product_code"`
	PolicyType              PolicyType       `db:"policy_type" json:"policy_type"`
	SumAssured              float64          `db:"sum_assured" json:"sum_assured"`
	PolicyTerm              int              `db:"policy_term" json:"policy_term"`
	PremiumCeasingAge       *int             `db:"premium_ceasing_age" json:"premium_ceasing_age,omitempty"`
	PremiumPaymentFrequency PremiumFrequency `db:"premium_payment_frequency" json:"premium_payment_frequency"`
	EntryPath               EntryPath        `db:"entry_path" json:"entry_path"`
	Channel                 Channel          `db:"channel" json:"channel"`
	Status                  ProposalStatus   `db:"status" json:"status"`
	CurrentStage            string           `db:"current_stage" json:"current_stage"`
	IsMedicalRequired       bool             `db:"is_medical_required" json:"is_medical_required"`
	IsPANRequired           bool             `db:"is_pan_required" json:"is_pan_required"`
	WorkflowID              *string          `db:"workflow_id" json:"workflow_id,omitempty"`
	CreatedBy               int64            `db:"created_by" json:"created_by"`
	CreatedAt               time.Time        `db:"created_at" json:"created_at"`
	UpdatedAt               time.Time        `db:"updated_at" json:"updated_at"`
	DeletedAt               *time.Time       `db:"deleted_at" json:"deleted_at,omitempty"`
	Version                 int              `db:"version" json:"version"`
	BasePremium             float64          `db:"base_premium" json:"base_premium"`
	TotalPremium            float64          `db:"total_premium" json:"total_premium"`
	GSTAmount               float64          `db:"gst_amount" json:"gst_amount"`
}

type ProposalOutput struct {
	ProposalID              int64            `db:"proposal_id" json:"proposal_id"`
	ProposalNumber          string           `db:"proposal_number" json:"proposal_number"`
	InsurantName            string           `db:"insurant_name" json:"insurant_name"`
	QuoteRefNumber          *string          `db:"quote_ref_number" json:"quote_ref_number,omitempty"`
	CustomerID              *int64           `db:"customer_id" json:"customer_id,omitempty"`
	SpouseCustomerID        *int64           `db:"spouse_customer_id" json:"spouse_customer_id,omitempty"`
	ProposerCustomerID      *int64           `db:"proposer_customer_id" json:"proposer_customer_id,omitempty"`
	IsProposerSameAsInsured bool             `db:"is_proposer_same_as_insured" json:"is_proposer_same_as_insured"`
	PremiumPayerType        PremiumPayerType `db:"premium_payer_type" json:"premium_payer_type"`
	PayerCustomerID         *int64           `db:"payer_customer_id" json:"payer_customer_id,omitempty"`
	ProductCode             string           `db:"product_code" json:"product_code"`
	PolicyType              PolicyType       `db:"policy_type" json:"policy_type"`
	SumAssured              float64          `db:"sum_assured" json:"sum_assured"`
	PolicyTerm              int              `db:"policy_term" json:"policy_term"`
	PremiumCeasingAge       *int             `db:"premium_ceasing_age" json:"premium_ceasing_age,omitempty"`
	PremiumPaymentFrequency PremiumFrequency `db:"premium_payment_frequency" json:"premium_payment_frequency"`
	EntryPath               EntryPath        `db:"entry_path" json:"entry_path"`
	Channel                 Channel          `db:"channel" json:"channel"`
	Status                  ProposalStatus   `db:"status" json:"status"`
	CurrentStage            string           `db:"current_stage" json:"current_stage"`
	// IsMedicalRequired       bool             `db:"is_medical_required" json:"is_medical_required"`
	// IsPANRequired           bool             `db:"is_pan_required" json:"is_pan_required"`
	WorkflowID   *string    `db:"workflow_id" json:"workflow_id,omitempty"`
	CreatedBy    int64      `db:"created_by" json:"created_by"`
	CreatedAt    time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time  `db:"updated_at" json:"updated_at"`
	DeletedAt    *time.Time `db:"deleted_at" json:"deleted_at,omitempty"`
	Version      int        `db:"version" json:"version"`
	BasePremium  float64    `db:"base_premium" json:"base_premium"`
	TotalPremium float64    `db:"total_premium" json:"total_premium"`
	GSTAmount    float64    `db:"gst_amount" json:"gst_amount"`
}

// CanTransitionTo checks if the proposal can transition to the given status
// [BR-POL-015] Proposal State Machine
func (p *Proposal) CanTransitionTo(newStatus ProposalStatus) bool {
	validTransitions := map[ProposalStatus][]ProposalStatus{
		ProposalStatusDraft:           {ProposalStatusIndexed, ProposalStatusCancelledDeath},
		ProposalStatusIndexed:         {ProposalStatusDataEntry},
		ProposalStatusDataEntry:       {ProposalStatusQCPending},
		ProposalStatusQCPending:       {ProposalStatusQCApproved, ProposalStatusQCRejected, ProposalStatusQCReturned},
		ProposalStatusQCReturned:      {ProposalStatusDataEntry},
		ProposalStatusQCApproved:      {ProposalStatusPendingMedical, ProposalStatusApprovalPending},
		ProposalStatusPendingMedical:  {ProposalStatusMedicalApproved, ProposalStatusMedicalRejected},
		ProposalStatusMedicalApproved: {ProposalStatusApprovalPending},
		ProposalStatusApprovalPending: {ProposalStatusApproved, ProposalStatusRejected},
		ProposalStatusApproved:        {ProposalStatusIssued},
		ProposalStatusIssued:          {ProposalStatusDispatched},
		ProposalStatusDispatched:      {ProposalStatusFreeLookActive},
		ProposalStatusFreeLookActive:  {ProposalStatusActive, ProposalStatusFLCCancelled},
	}

	allowedStatuses, exists := validTransitions[p.Status]
	if !exists {
		return false
	}

	for _, allowed := range allowedStatuses {
		if allowed == newStatus {
			return true
		}
	}
	return false
}

// ProposalIndexing represents the indexing phase data
type ProposalIndexing struct {
	ProposalIndexingID int64     `db:"proposal_indexing_id" json:"proposal_indexing_id"`
	ProposalID         int64     `db:"proposal_id" json:"proposal_id"`
	POCode             string    `db:"po_code" json:"po_code"`
	IssueCircle        string    `db:"issue_circle" json:"issue_circle"`
	IssueHO            string    `db:"issue_ho" json:"issue_ho"`
	IssuePostOffice    string    `db:"issue_post_office" json:"issue_post_office"`
	CircleCode         *string   `db:"circle_code" json:"circle_code,omitempty"`
	DivisionCode       *string   `db:"division_code" json:"division_code,omitempty"`
	DeclarationDate    time.Time `db:"declaration_date" json:"declaration_date"`
	ReceiptDate        time.Time `db:"receipt_date" json:"receipt_date"`
	IndexingDate       time.Time `db:"indexing_date" json:"indexing_date"`
	ProposalDate       time.Time `db:"proposal_date" json:"proposal_date"`
	OpportunityID      *string   `db:"opportunity_id" json:"opportunity_id,omitempty"`
	ReceiptNumber      *string   `db:"receipt_number" json:"receipt_number,omitempty"`
	CreatedAt          time.Time `db:"created_at" json:"created_at"`
	UpdatedAt          time.Time `db:"updated_at" json:"updated_at"`
}

// SubsequentPaymentMode represents the mode for subsequent premium payments
type SubsequentPaymentMode string

const (
	SubsequentPaymentCash                SubsequentPaymentMode = "CASH"
	SubsequentPaymentOnline              SubsequentPaymentMode = "ONLINE"
	SubsequentPaymentNACH                SubsequentPaymentMode = "NACH"
	SubsequentPaymentStandingInstruction SubsequentPaymentMode = "STANDING_INSTRUCTION"
	SubsequentPaymentPOSB                SubsequentPaymentMode = "POSB"
)

// ProposalDataEntry represents the data entry phase
type ProposalDataEntry struct {
	ProposalDataEntryID     int64                  `db:"data_entry_id" json:"data_entry_id"`
	ProposalID              int64                  `db:"proposal_id" json:"proposal_id"`
	PolicyTakenUnder        *PolicyTakenUnder      `db:"policy_taken_under" json:"policy_taken_under,omitempty"`
	AadharPhotoDocumentID   *string                `db:"aadhaar_photo_document_id" json:"aadhar_photo_document_id,omitempty"`
	AgeProofType            *AgeProofType          `db:"age_proof_type" json:"age_proof_type,omitempty"`
	SubsequentPaymentMode   *SubsequentPaymentMode `db:"subsequent_payment_mode" json:"subsequent_payment_mode,omitempty"`
	DataEntryStatus         string                 `db:"data_entry_status" json:"data_entry_status"`
	InsuredDetailsComplete  bool                   `db:"insured_details_complete" json:"insured_details_complete"`
	NomineeDetailsComplete  bool                   `db:"nominee_details_complete" json:"nominee_details_complete"`
	PolicyDetailsComplete   bool                   `db:"policy_details_complete" json:"policy_details_complete"`
	AgentDetailsComplete    bool                   `db:"agent_details_complete" json:"agent_details_complete"`
	MedicalDetailsComplete  bool                   `db:"medical_details_complete" json:"medical_details_complete"`
	DeclarationComplete     bool                   `db:"declaration_complete" json:"declaration_complete"`
	ProposerDetailsComplete bool                   `db:"proposer_details_complete" json:"proposer_details_complete"`
	DocumentsComplete       bool                   `db:"documents_complete" json:"documents_complete"`
	DataEntryBy             *int64                 `db:"data_entry_by" json:"data_entry_by,omitempty"`
	CreatedAt               time.Time              `db:"created_at"`
	UpdatedAt               time.Time              `db:"updated_at"`
}

// ProposalQCReview represents the QC review phase
type ProposalQCReview struct {
	ProposalQCReviewID int64      `db:"qc_review_id" json:"qc_review_id"`
	ProposalID         int64      `db:"proposal_id" json:"proposal_id"`
	QCAssignedTo       *int64     `db:"qc_assigned_to" json:"qc_assigned_to,omitempty"`
	QCAssignedAt       *time.Time `db:"qc_assigned_at" json:"qc_assigned_at,omitempty"`
	QRDecision         *string    `db:"qr_decision" json:"qr_decision,omitempty"`
	QRDecisionAt       *time.Time `db:"qr_decision_at" json:"qr_decision_at,omitempty"`
	QRComments         *string    `db:"qr_comments" json:"qr_comments,omitempty"`
	QRDecisionBy       *int64     `db:"qr_decision_by" json:"qr_decision_by,omitempty"`
	CreatedAt          time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt          time.Time  `db:"updated_at" json:"updated_at"`
	ReturnCount        int        `db:"return_count" json:"return_count"`
	LastReturnReason   *string    `db:"last_return_reason" json:"last_return_reason,omitempty"`
}

// ProposalApproval represents the approval phase
type ProposalApproval struct {
	ProposalApprovalID      int64      `db:"approval_id" json:"approval_id"`
	ProposalID              int64      `db:"proposal_id" json:"proposal_id"`
	ApprovalLevel           int        `db:"approval_level" json:"approval_level"`
	ApproverRole            *string    `db:"approver_role" json:"approver_role"`
	AssignedApproverID      *int64     `db:"assigned_approver_id" json:"assigned_approver_id,omitempty"`
	ApproverDecision        *string    `db:"approver_decision" json:"approver_decision,omitempty"`
	ApproverComments        *string    `db:"approver_comments" json:"approver_comments,omitempty"`
	ApproverDecisionBy      *int64     `db:"approver_decision_by" json:"approver_decision_by,omitempty"`
	ApproverDecisionAt      *time.Time `db:"approver_decision_at" json:"approver_decision_at,omitempty"`
	ApprovalDueDate         *time.Time `db:"approval_due_date" json:"approval_duedate,omitempty"`
	ApprovalReminderSent    *bool      `db:"approval_reminder_sent" json:"approval_reminder_sent,omitempty"`
	ApproverRejectionReason *string    `db:"approver_rejection_reason" json:"approver_rejection_reason,omitempty"`
	CreatedAt               time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt               time.Time  `db:"updated_at" json:"updated_at"`
}

// ProposalIssuance represents the issuance phase
type ProposalIssuance struct {
	ProposalIssuanceID int64      `db:"issuance_id" json:"issuance_id"`
	ProposalID         int64      `db:"proposal_id" json:"proposal_id"`
	PolicyNumber       *string    `db:"policy_number" json:"policy_number,omitempty"`
	IssuedAt           *time.Time `db:"issued_at" json:"issued_at,omitempty"`
	IssuedBy           *int64     `db:"issued_by" json:"issued_by,omitempty"`
	BondGenerated      bool       `db:"bond_generated" json:"bond_generated"`
	BondDocumentID     *string    `db:"bond_document_id" json:"bond_document_id,omitempty"`
	DispatchDate       *time.Time `db:"dispatch_date" json:"dispatch_date,omitempty"`
	DispatchMode       *string    `db:"dispatch_mode" json:"dispatch_mode,omitempty"`
	TrackingNumber     *string    `db:"tracking_number" json:"tracking_number,omitempty"`
	FLCStatus          string     `db:"flc_status" json:"flc_status"`
	FLCStartDate       *time.Time `db:"flc_start_date" json:"flc_start_date,omitempty"`
	FLCEndDate         *time.Time `db:"flc_end_date" json:"flc_end_date,omitempty"`
	CreatedAt          time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt          time.Time  `db:"updated_at" json:"updated_at"`
}

// ProposalDetail combines all proposal data for API responses
type ProposalDetail struct {
	Proposal  Proposal           `json:"proposal"`
	Indexing  ProposalIndexing   `json:"indexing,omitempty"`
	DataEntry *ProposalDataEntry `json:"data_entry,omitempty"`
	QCReview  *ProposalQCReview  `json:"qc_review,omitempty"`
	Approval  *ProposalApproval  `json:"approval,omitempty"`
	Issuance  *ProposalIssuance  `json:"issuance,omitempty"`
	Insured   *ProposalInsured   `json:"insured,omitempty"`
}

// ProposalInsured represents the insured person detailed information
type ProposalInsured struct {
	InsuredID     int64     `db:"insured_id" json:"insured_id"`
	ProposalID    int64     `db:"proposal_id" json:"proposal_id"`
	Salutation    string    `db:"salutation" json:"salutation"`
	FirstName     string    `db:"first_name" json:"first_name"`
	MiddleName    *string   `db:"middle_name" json:"middle_name,omitempty"`
	LastName      string    `db:"last_name" json:"last_name"`
	Gender        string    `db:"gender" json:"gender"`
	DateOfBirth   string    `db:"date_of_birth" json:"date_of_birth"`
	MaritalStatus *string   `db:"marital_status" json:"marital_status,omitempty"`
	Occupation    *string   `db:"occupation" json:"occupation,omitempty"`
	AnnualIncome  *float64  `db:"annual_income" json:"annual_income,omitempty"`
	AddressLine1  *string   `db:"address_line1" json:"address_line1,omitempty"`
	AddressLine2  *string   `db:"address_line2" json:"address_line2,omitempty"`
	AddressLine3  *string   `db:"address_line3" json:"address_line3,omitempty"`
	City          *string   `db:"city" json:"city,omitempty"`
	State         *string   `db:"state" json:"state,omitempty"`
	PinCode       *string   `db:"pin_code" json:"pin_code,omitempty"`
	Mobile        *string   `db:"mobile" json:"mobile,omitempty"`
	Email         *string   `db:"email" json:"email,omitempty"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time `db:"updated_at" json:"updated_at"`
}

type ProposalInsuredOuptput struct {
	InsuredID     int64     `db:"insured_id" json:"insured_id"`
	ProposalID    int64     `db:"proposal_id" json:"proposal_id"`
	Salutation    string    `db:"salutation" json:"salutation"`
	FirstName     string    `db:"first_name" json:"first_name"`
	MiddleName    *string   `db:"middle_name" json:"middle_name,omitempty"`
	LastName      string    `db:"last_name" json:"last_name"`
	Gender        string    `db:"gender" json:"gender"`
	DateOfBirth   time.Time `db:"date_of_birth" json:"date_of_birth"`
	MaritalStatus *string   `db:"marital_status" json:"marital_status,omitempty"`
	Occupation    *string   `db:"occupation" json:"occupation,omitempty"`
	AnnualIncome  *float64  `db:"annual_income" json:"annual_income,omitempty"`
	AddressLine1  *string   `db:"address_line1" json:"address_line1,omitempty"`
	AddressLine2  *string   `db:"address_line2" json:"address_line2,omitempty"`
	AddressLine3  *string   `db:"address_line3" json:"address_line3,omitempty"`
	City          *string   `db:"city" json:"city,omitempty"`
	State         *string   `db:"state" json:"state,omitempty"`
	PinCode       *string   `db:"pin_code" json:"pin_code,omitempty"`
	Mobile        *string   `db:"mobile" json:"mobile,omitempty"`
	Email         *string   `db:"email" json:"email,omitempty"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time `db:"updated_at" json:"updated_at"`
}

// CalculateAge calculates age from date of birth string (YYYY-MM-DD)
func CalculateAge(dobStr string) int {
	dob, err := time.Parse("2006-01-02", dobStr)
	if err != nil {
		log.Error(context.TODO(), err)
		return 0
	}
	now := time.Now()
	years := now.Year() - dob.Year()
	if now.YearDay() < dob.YearDay() {
		years--
	}
	return years
}

func CalculateAgeTime(dob time.Time) int {
	now := time.Now()
	years := now.Year() - dob.Year()
	if now.YearDay() < dob.YearDay() {
		years--
	}
	return years
}
func CalculateANB(dob time.Time, proposalDate time.Time) int {
	years := proposalDate.Year() - dob.Year()

	if proposalDate.Month() < dob.Month() ||
		(proposalDate.Month() == dob.Month() &&
			proposalDate.Day() < dob.Day()) {
		years--
	}

	return years + 1
}

// ProposalNominee represents a nominee for the proposal
type ProposalNominee struct {
	NomineeID             int64     `db:"nominee_id" json:"nominee_id"`
	ProposalID            int64     `db:"proposal_id" json:"proposal_id"`
	Salutation            string    `db:"salutation" json:"salutation"`
	FirstName             string    `db:"first_name" json:"first_name"`
	MiddleName            *string   `db:"middle_name" json:"middle_name,omitempty"`
	LastName              string    `db:"last_name" json:"last_name"`
	Gender                string    `db:"gender" json:"gender"`
	DateOfBirth           string    `db:"date_of_birth" json:"date_of_birth"`
	IsMinor               bool      `db:"is_minor" json:"is_minor"`
	Relationship          string    `db:"relationship" json:"relationship"`
	SharePercentage       float64   `db:"share_percentage" json:"share_percentage"`
	AppointeeName         *string   `db:"appointee_name" json:"appointee_name,omitempty"`
	AppointeeRelationship *string   `db:"appointee_relationship" json:"appointee_relationship,omitempty"`
	CreatedAt             time.Time `db:"created_at" json:"created_at"`
	UpdatedAt             time.Time `db:"updated_at" json:"updated_at"`
	NomineeCustomerID     *int64    `db:"nominee_customer_id" json:"nominee_customer_id"`
}

// ProposalMedicalInfo represents medical questionnaire data
//
//	type ProposalMedicalInfo struct {
//		MedicalInfoID     int64  `db:"medical_info_id" json:"medical_info_id"`
//		ProposalID        int64  `db:"proposal_id" json:"proposal_id"`
//		InsuredIndex      int    `db:"insured_index" json:"insured_index"`
//		IsSoundHealth     bool   `db:"is_sound_health" json:"is_sound_health"`
//		DiseaseTB         bool   `db:"disease_tb" json:"disease_tb"`
//		DiseaseCancer     bool   `db:"disease_cancer" json:"disease_cancer"`
//		DiseaseParalysis  bool   `db:"disease_paralysis" json:"disease_paralysis"`
//		DiseaseInsanity   bool   `db:"disease_insanity" json:"disease_insanity"`
//		DiseaseHeartLungs bool   `db:"disease_heart_lungs" json:"disease_heart_lungs"`
//		DiseaseKidney     bool   `db:"disease_kidney" json:"disease_kidney"`
//		DiseaseBrain      bool   `db:"disease_brain" json:"disease_brain"`
//		DiseaseHIV        bool   `db:"disease_hiv" json:"disease_hiv"`
//		DiseaseHepatitisB bool   `db:"disease_hepatitis_b" json:"disease_hepatitis_b"`
//		DiseaseEpilepsy   bool   `db:"disease_epilepsy" json:"disease_epilepsy"`
//		DiseaseNervous    bool   `db:"disease_nervous" json:"disease_nervous"`
//		DiseaseLiver      bool   `db:"disease_liver" json:"disease_liver"`
//		DiseaseLeprosy    bool   `db:"disease_leprosy" json:"disease_leprosy"`
//		OtherDiseases     bool   `db:"other_diseases" json:"other_diseases"`
//		DiseaseDetails    string `db:"disease_details" json:"disease_details"`
//	}
type ProposalMedicalInfo struct {
	MedicalInfoID            int64      `db:"medical_info_id" json:"medical_info_id"`
	ProposalID               int64      `db:"proposal_id" json:"proposal_id"`
	InsuredIndex             int        `db:"insured_index" json:"insured_index"`
	IsSoundHealth            bool       `db:"is_sound_health" json:"is_sound_health"`
	DiseaseTB                bool       `db:"disease_tb" json:"disease_tb"`
	DiseaseCancer            bool       `db:"disease_cancer" json:"disease_cancer"`
	DiseaseParalysis         bool       `db:"disease_paralysis" json:"disease_paralysis"`
	DiseaseInsanity          bool       `db:"disease_insanity" json:"disease_insanity"`
	DiseaseHeartLungs        bool       `db:"disease_heart_lungs" json:"disease_heart_lungs"`
	DiseaseKidney            bool       `db:"disease_kidney" json:"disease_kidney"`
	DiseaseBrain             bool       `db:"disease_brain" json:"disease_brain"`
	DiseaseHIV               bool       `db:"disease_hiv" json:"disease_hiv"`
	DiseaseHepatitisB        bool       `db:"disease_hepatitis_b" json:"disease_hepatitis_b"`
	DiseaseEpilepsy          bool       `db:"disease_epilepsy" json:"disease_epilepsy"`
	DiseaseNervous           bool       `db:"disease_nervous" json:"disease_nervous"`
	DiseaseLiver             bool       `db:"disease_liver" json:"disease_liver"`
	DiseaseLeprosy           bool       `db:"disease_leprosy" json:"disease_leprosy"`
	DiseasePhysicalDeformity bool       `db:"disease_physical_deformity" json:"disease_physical_deformity"`
	DiseaseOther             bool       `db:"disease_other" json:"disease_other"`
	DiseaseDetails           string     `db:"disease_details" json:"disease_details"`
	FamilyHereditary         bool       `db:"family_hereditary" json:"family_hereditary"`
	FamilyHereditaryDetails  string     `db:"family_hereditary_details" json:"family_hereditary_details"`
	MedicalLeave3yr          bool       `db:"medical_leave_3yr" json:"medical_leave_3yr"`
	LeaveKind                string     `db:"leave_kind" json:"leave_kind"`
	LeavePeriod              string     `db:"leave_period" json:"leave_period"`
	LeaveAilment             string     `db:"leave_ailment" json:"leave_ailment"`
	HospitalName             string     `db:"hospital_name" json:"hospital_name"`
	HospitalizationFrom      *time.Time `db:"hospitalization_from" json:"hospitalization_from,omitempty"`
	HospitalizationTo        *time.Time `db:"hospitalization_to" json:"hospitalization_to,omitempty"`
	PhysicalDeformity        bool       `db:"physical_deformity" json:"physical_deformity"`
	DeformityType            *string    `db:"deformity_type" json:"deformity_type,omitempty"`
	FamilyDoctorName         string     `db:"family_doctor_name" json:"family_doctor_name"`
	CreatedAt                time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt                time.Time  `db:"updated_at" json:"updated_at"`
}

// ProposerRelationship represents the relationship between proposer and insured
type ProposerRelationship string

const (
	ProposerRelationshipParent   ProposerRelationship = "PARENT"
	ProposerRelationshipSpouse   ProposerRelationship = "SPOUSE"
	ProposerRelationshipEmployer ProposerRelationship = "EMPLOYER"
	ProposerRelationshipHUFKarta ProposerRelationship = "HUF_KARTA"
	ProposerRelationshipGuardian ProposerRelationship = "GUARDIAN"
	ProposerRelationshipOther    ProposerRelationship = "OTHER"
	ProposerRelationshipSelf     ProposerRelationship = "SELF"
)

// ProposalProposer represents proposer details when proposer ≠ insured
type ProposalProposer struct {
	ProposerID            int64                `db:"proposer_id" json:"proposer_id"`
	ProposalID            int64                `db:"proposal_id" json:"proposal_id"`
	CustomerID            int64                `db:"customer_id" json:"customer_id"`
	RelationshipToInsured ProposerRelationship `db:"relationship_to_insured" json:"relationship_to_insured"`
	RelationshipDetails   *string              `db:"relationship_details" json:"relationship_details,omitempty"`
	CreatedAt             time.Time            `db:"created_at" json:"created_at"`
	UpdatedAt             time.Time            `db:"updated_at" json:"updated_at"`
}

type ProposalIndexingSection struct {
	DeclarationDate time.Time `db:"declaration_date" json:"declaration_date"`
	ReceiptDate     time.Time `db:"receipt_date" json:"receipt_date"`
	IndexingDate    time.Time `db:"indexing_date" json:"indexing_date"`
	ProposalDate    time.Time `db:"proposal_date" json:"proposal_date"`
	POCode          string    `db:"po_code" json:"po_code"`
	IssueCircle     *string   `db:"issue_circle" json:"issue_circle,omitempty"`
	IssueHO         *string   `db:"issue_ho" json:"issue_ho,omitempty"`
	IssuePostOffice *string   `db:"issue_post_office" json:"issue_post_office,omitempty"`
}

type ProposalFirstPremium struct {
	FirstPremiumPaid          bool       `db:"first_premium_paid" json:"first_premium_paid"`
	FirstPremiumDate          *time.Time `db:"first_premium_date" json:"first_premium_date,omitempty"`
	FirstPremiumReference     *string    `db:"first_premium_reference" json:"first_premium_reference,omitempty"`
	FirstPremiumReceiptNumber *string    `db:"first_premium_receipt_number" json:"first_premium_receipt_number,omitempty"`
	PremiumPaymentMethod      *string    `db:"premium_payment_method" json:"premium_payment_method,omitempty"`
	InitialPremium            *float64   `db:"initial_premium" json:"initial_premium,omitempty"`
	ShortExcessPremium        *float64   `db:"short_excess_premium" json:"short_excess_premium,omitempty"`
}

type ProposalAgentOutput struct {
	AgentID                string  `db:"agent_id" json:"agent_id"`
	AgentSalutation        *string `db:"agent_salutation" json:"agent_salutation,omitempty"`
	AgentName              *string `db:"agent_name" json:"agent_name,omitempty"`
	AgentMobile            *string `db:"agent_mobile" json:"agent_mobile,omitempty"`
	AgentEmail             *string `db:"agent_email" json:"agent_email,omitempty"`
	AgentLandline          *string `db:"agent_landline" json:"agent_landline,omitempty"`
	AgentSTDCode           *string `db:"agent_std_code" json:"agent_std_code,omitempty"`
	ReceivesCorrespondence bool    `db:"receives_correspondence" json:"receives_correspondence"`
	OpportunityID          *string `db:"opportunity_id" json:"opportunity_id,omitempty"`
}

// ChangeType represents the type of change in audit log
type ChangeType string

const (
	ChangeTypeInsert ChangeType = "INSERT"
	ChangeTypeUpdate ChangeType = "UPDATE"
	ChangeTypeDelete ChangeType = "DELETE"
)

// ProposalAuditLog represents field-level audit trail for proposal changes
type ProposalAuditLog struct {
	AuditID      int64      `db:"audit_id" json:"audit_id"`
	ProposalID   int64      `db:"proposal_id" json:"proposal_id"`
	EntityType   string     `db:"entity_type" json:"entity_type"`
	EntityID     int64      `db:"entity_id" json:"entity_id"`
	FieldName    string     `db:"field_name" json:"field_name"`
	OldValue     *string    `db:"old_value" json:"old_value,omitempty"`
	NewValue     *string    `db:"new_value" json:"new_value,omitempty"`
	ChangeType   ChangeType `db:"change_type" json:"change_type"`
	ChangedBy    int64      `db:"changed_by" json:"changed_by"`
	ChangedAt    time.Time  `db:"changed_at" json:"changed_at"`
	ChangeReason *string    `db:"change_reason" json:"change_reason,omitempty"`
	Metadata     *string    `db:"metadata" json:"metadata,omitempty"`
}
