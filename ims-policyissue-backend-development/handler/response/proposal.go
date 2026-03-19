package response

import (
	"policy-issue-service/core/domain"
	"policy-issue-service/core/port"
	"time"
)

// ProposalIndexingResponse represents the response for Create Proposal Indexing API
// [POL-API-005]
type ProposalIndexingResponse struct {
	port.StatusCodeAndMessage
	ProposalID     int64  `json:"proposal_id"`
	ProposalNumber string `json:"proposal_number"`
	Status         string `json:"status"`
}

// ProposalDetailResponse represents the response for Get Proposal API
// [POL-API-006]
type ProposalDetailResponse struct {
	port.StatusCodeAndMessage
	ProposalID       int64   `json:"proposal_id"`
	ProposalNumber   string  `json:"proposal_number"`
	Status           string  `json:"status"`
	PolicyType       string  `json:"policy_type"`
	ProductCode      string  `json:"product_code"`
	CustomerID       *int64  `json:"customer_id,omitempty"`
	SpouseCustomerID *int64  `json:"spouse_customer_id,omitempty"`
	Channel          string  `json:"channel"`
	SumAssured       float64 `json:"sum_assured"`
	PolicyTerm       int     `json:"policy_term"`
}

// ResolveProposalResponse represents the response for Resolve Proposal Number API
// [POL-API-017]
type ResolveProposalResponse struct {
	port.StatusCodeAndMessage
	ProposalID int64 `json:"proposal_id"`
}

// FirstPremiumResponse represents the response for Record First Premium API
// [POL-API-007]
type FirstPremiumResponse struct {
	port.StatusCodeAndMessage
	Status string `json:"status"`
}

// SectionUpdateResponse represents the response for section update APIs
// [POL-API-008] [POL-API-009] [POL-API-010] [POL-API-011] [POL-API-012]
type SectionUpdateResponse struct {
	port.StatusCodeAndMessage
	Status    string `json:"status"`
	UpdatedAt string `json:"updated_at"`
}

// StartDataEntryResponse represents the response for Start Data Entry API
type StartDataEntryResponse struct {
	port.StatusCodeAndMessage
	ProposalID     int64  `json:"proposal_id"`
	ProposalNumber string `json:"proposal_number"`
	PreviousStatus string `json:"previous_status"`
	NewStatus      string `json:"new_status"`
	AssignedTo     int64  `json:"assigned_to"`
	AssignedAt     string `json:"assigned_at"`
}

// SubmitForQCResponse represents the response for Submit for QC API
// [POL-API-013]
type SubmitForQCResponse struct {
	port.StatusCodeAndMessage
	Status     string `json:"status"`
	WorkflowID string `json:"workflow_id,omitempty"`
	RunID      string `json:"run_id,omitempty"`
}

// ProposalSummaryResponse represents the response for Get Proposal Summary API
// [POL-API-016]
type ProposalSummaryResponse struct {
	port.StatusCodeAndMessage
	ProposalID     int64  `json:"proposal_id"`
	ProposalNumber string `json:"proposal_number"`
	Status         string `json:"status"`
}

// ProposalQueueResponse represents the response for Get Proposal Queue API
// [WF-POL-003]
type ProposalQueueResponse struct {
	port.StatusCodeAndMessage
	port.MetaDataResponse
	Proposals []ProposalSummary `json:"proposals"`
}

// ProposalSummary represents a proposal in the queue
type ProposalSummary struct {
	ProposalID     int64   `json:"proposal_id"`
	ProposalNumber string  `json:"proposal_number"`
	Status         string  `json:"status"`
	CustomerName   string  `json:"customer_name"`
	ProductCode    string  `json:"product_code"`
	SumAssured     float64 `json:"sum_assured"`
	CreatedAt      string  `json:"created_at"`
}

// ProposalSectionResponse represents response for section-wise proposal data
type ProposalSectionResponse struct {
	port.StatusCodeAndMessage
	Section string      `json:"section"`
	Data    interface{} `json:"data"`
}

type ProposerResponse struct {
	ProposerID            int64   `json:"proposer_id"`
	CustomerID            int64   `json:"customer_id"`
	RelationshipToInsured string  `json:"relationship_to_insured"`
	RelationshipDetails   *string `json:"relationship_details,omitempty"`
}

// NomineeResponse hides proposal_id and timestamps
type NomineeResponse struct {
	NomineeID             int64   `json:"nominee_id"`
	Salutation            string  `json:"salutation"`
	FirstName             string  `json:"first_name"`
	MiddleName            *string `json:"middle_name,omitempty"`
	LastName              string  `json:"last_name"`
	Gender                string  `json:"gender"`
	DateOfBirth           *string `json:"date_of_birth"`
	IsMinor               bool    `json:"is_minor"`
	Relationship          string  `json:"relationship"`
	SharePercentage       float64 `json:"share_percentage"`
	AppointeeName         *string `json:"appointee_name,omitempty"`
	AppointeeRelationship *string `json:"appointee_relationship,omitempty"`
}

// QCReviewResponse hides timestamps and internal user IDs
type QCReviewResponse struct {
	QCReviewID  int64   `json:"qc_review_id"`
	QRDecision  *string `json:"qr_decision,omitempty"`
	QRComments  *string `json:"qr_comments,omitempty"`
	ReturnCount int     `json:"return_count"`
}

// DataEntryResponse hides data_entry_by and timestamps
type DataEntryResponse struct {
	DataEntryID             int64  `json:"data_entry_id"`
	PolicyTakenUnder        string `json:"policy_taken_under"`
	AgeProofType            string `json:"age_proof_type"`
	SubsequentPaymentMode   string `json:"subsequent_payment_mode"`
	DataEntryStatus         string `json:"data_entry_status"`
	InsuredDetailsComplete  bool   `json:"insured_details_complete"`
	NomineeDetailsComplete  bool   `json:"nominee_details_complete"`
	PolicyDetailsComplete   bool   `json:"policy_details_complete"`
	AgentDetailsComplete    bool   `json:"agent_details_complete"`
	MedicalDetailsComplete  bool   `json:"medical_details_complete"`
	DeclarationComplete     bool   `json:"declaration_complete"`
	ProposerDetailsComplete bool   `json:"proposer_details_complete"`
	DocumentsComplete       bool   `json:"documents_complete"`
}

func FetchProposerResponse(d *domain.ProposalProposer) *ProposerResponse {

	if d == nil {
		return nil
	}
var relationship string
if d.RelationshipToInsured != "" {
	relationship = string(d.RelationshipToInsured)
}
	return &ProposerResponse{
		ProposerID:            d.ProposerID,
		CustomerID:            d.CustomerID,
		RelationshipToInsured: relationship,
		RelationshipDetails:   d.RelationshipDetails,
	}
}

func FetchNomineeResponse(data []domain.ProposalNominee) []NomineeResponse {

	resp := make([]NomineeResponse, 0, len(data))

	for _, n := range data {
		var dob *string
		if n.DateOfBirth != "" {
			dob = &n.DateOfBirth
		}

		resp = append(resp, NomineeResponse{
			NomineeID:             n.NomineeID,
			Salutation:            n.Salutation,
			FirstName:             n.FirstName,
			MiddleName:            n.MiddleName,
			LastName:              n.LastName,
			Gender:                n.Gender,
			DateOfBirth:           dob,
			IsMinor:               n.IsMinor,
			Relationship:          n.Relationship,
			SharePercentage:       n.SharePercentage,
			AppointeeName:         n.AppointeeName,
			AppointeeRelationship: n.AppointeeRelationship,
		})
	}

	return resp
}
func FetchQCReviewResponse(d *domain.ProposalQCReview) *QCReviewResponse {

	if d == nil {
		return nil
	}

	return &QCReviewResponse{
		QCReviewID:  d.ProposalQCReviewID,
		QRDecision:  d.QRDecision,
		QRComments:  d.QRComments,
		ReturnCount: d.ReturnCount,
	}
}

// Add similar mapping functions for DataEntry, Medical, etc.

func FetchDataEntryResponse(d *domain.ProposalDataEntry) *DataEntryResponse {

	if d == nil {
		return nil
	}

	policyTakenUnder := ""
	if d.PolicyTakenUnder != nil {
		policyTakenUnder = string(*d.PolicyTakenUnder)
	}

	ageProofType := ""
	if d.AgeProofType != nil {
		ageProofType = string(*d.AgeProofType)
	}

	subsequentPaymentMode := ""
	if d.SubsequentPaymentMode != nil {
		subsequentPaymentMode = string(*d.SubsequentPaymentMode)
	}

	return &DataEntryResponse{
		DataEntryID:             d.ProposalDataEntryID,
		PolicyTakenUnder:        policyTakenUnder,
		AgeProofType:            ageProofType,
		SubsequentPaymentMode:   subsequentPaymentMode,
		DataEntryStatus:         string(d.DataEntryStatus),
		InsuredDetailsComplete:  d.InsuredDetailsComplete,
		NomineeDetailsComplete:  d.NomineeDetailsComplete,
		PolicyDetailsComplete:   d.PolicyDetailsComplete,
		AgentDetailsComplete:    d.AgentDetailsComplete,
		MedicalDetailsComplete:  d.MedicalDetailsComplete,
		DeclarationComplete:     d.DeclarationComplete,
		ProposerDetailsComplete: d.ProposerDetailsComplete,
		DocumentsComplete:       d.DocumentsComplete,
	}
}

type InsuredResponse struct {
	InsuredID     int64     `json:"insured_id"`
	Salutation    string    `json:"salutation"`
	FirstName     string    `json:"first_name"`
	MiddleName    *string   `json:"middle_name,omitempty"`
	LastName      string    `json:"last_name"`
	Gender        string    `json:"gender"`
	DateOfBirth   time.Time `json:"date_of_birth"`
	MaritalStatus string    `json:"marital_status"`
	Occupation    string    `json:"occupation"`
	AnnualIncome  float64   `json:"annual_income"`
	AddressLine1  string    `json:"address_line1"`
	AddressLine2  *string   `json:"address_line2,omitempty"`
	AddressLine3  *string   `json:"address_line3,omitempty"`
	City          string    `json:"city"`
	State         string    `json:"state"`
	PinCode       string    `json:"pin_code"`
	Mobile        string    `json:"mobile"`
	Email         string    `json:"email"`
}

func FetchInsuredResponse(d *domain.ProposalInsuredOuptput) *InsuredResponse {

	if d == nil {
		return nil
	}

	maritalStatus := ""
	if d.MaritalStatus != nil {
		maritalStatus = *d.MaritalStatus
	}

	occupation := ""
	if d.Occupation != nil {
		occupation = *d.Occupation
	}

	annualIncome := float64(0)
	if d.AnnualIncome != nil {
		annualIncome = *d.AnnualIncome
	}

	addressLine1 := ""
	if d.AddressLine1 != nil {
		addressLine1 = *d.AddressLine1
	}

	city := ""
	if d.City != nil {
		city = *d.City
	}

	state := ""
	if d.State != nil {
		state = *d.State
	}

	pincode := ""
	if d.PinCode != nil {
		pincode = *d.PinCode
	}

	mobile := ""
	if d.Mobile != nil {
		mobile = *d.Mobile
	}

	email := ""
	if d.Email != nil {
		email = *d.Email
	}

	return &InsuredResponse{
		InsuredID:     d.InsuredID,
		Salutation:    d.Salutation,
		FirstName:     d.FirstName,
		MiddleName:    d.MiddleName,
		LastName:      d.LastName,
		Gender:        d.Gender,
		DateOfBirth:   d.DateOfBirth,
		MaritalStatus: maritalStatus,
		Occupation:    occupation,
		AnnualIncome:  annualIncome,
		AddressLine1:  addressLine1,
		AddressLine2:  d.AddressLine2,
		AddressLine3:  d.AddressLine3,
		City:          city,
		State:         state,
		PinCode:       pincode,
		Mobile:        mobile,
		Email:         email,
	}
}

type MedicalInfoResponse struct {
	InsuredIndex             int    `json:"insured_index"`
	IsSoundHealth            bool   `json:"is_sound_health"`
	DiseaseTB                bool   `json:"disease_tb"`
	DiseaseCancer            bool   `json:"disease_cancer"`
	DiseaseParalysis         bool   `json:"disease_paralysis"`
	DiseaseInsanity          bool   `json:"disease_insanity"`
	DiseaseHeartLungs        bool   `json:"disease_heart_lungs"`
	DiseaseKidney            bool   `json:"disease_kidney"`
	DiseaseBrain             bool   `json:"disease_brain"`
	DiseaseHIV               bool   `json:"disease_hiv"`
	DiseaseHepatitisB        bool   `json:"disease_hepatitis_b"`
	DiseaseEpilepsy          bool   `json:"disease_epilepsy"`
	DiseaseNervous           bool   `json:"disease_nervous"`
	DiseaseLiver             bool   `json:"disease_liver"`
	DiseaseLeprosy           bool   `json:"disease_leprosy"`
	DiseasePhysicalDeformity bool   `json:"disease_physical_deformity"`
	DiseaseOther             bool   `json:"disease_other"`
	DiseaseDetails           string `json:"disease_details"`
	FamilyHereditary         bool   `json:"family_hereditary"`
	FamilyHereditaryDetails  string `json:"family_hereditary_details"`
}

func FetchMedicalInfoResponse(data []domain.ProposalMedicalInfo) []MedicalInfoResponse {

	resp := make([]MedicalInfoResponse, 0, len(data))

	for _, m := range data {

		resp = append(resp, MedicalInfoResponse{
			InsuredIndex:             m.InsuredIndex,
			IsSoundHealth:            m.IsSoundHealth,
			DiseaseTB:                m.DiseaseTB,
			DiseaseCancer:            m.DiseaseCancer,
			DiseaseParalysis:         m.DiseaseParalysis,
			DiseaseInsanity:          m.DiseaseInsanity,
			DiseaseHeartLungs:        m.DiseaseHeartLungs,
			DiseaseKidney:            m.DiseaseKidney,
			DiseaseBrain:             m.DiseaseBrain,
			DiseaseHIV:               m.DiseaseHIV,
			DiseaseHepatitisB:        m.DiseaseHepatitisB,
			DiseaseEpilepsy:          m.DiseaseEpilepsy,
			DiseaseNervous:           m.DiseaseNervous,
			DiseaseLiver:             m.DiseaseLiver,
			DiseaseLeprosy:           m.DiseaseLeprosy,
			DiseasePhysicalDeformity: m.DiseasePhysicalDeformity,
			DiseaseOther:             m.DiseaseOther,
			DiseaseDetails:           m.DiseaseDetails,
			FamilyHereditary:         m.FamilyHereditary,
			FamilyHereditaryDetails:  m.FamilyHereditaryDetails,
		})
	}

	return resp
}

type ProposalResponse struct {
	ProposalID              int64   `json:"proposal_id"`
	ProposalNumber          string  `json:"proposal_number"`
	InsurantName            string  `json:"insurant_name"`
	QuoteRefNumber          string  `json:"quote_ref_number"`
	CustomerID              int64   `json:"customer_id"`
	ProposerCustomerID      int64   `json:"proposer_customer_id"`
	IsProposerSameAsInsured bool    `json:"is_proposer_same_as_insured"`
	PremiumPayerType        string  `json:"premium_payer_type"`
	ProductCode             string  `json:"product_code"`
	PolicyType              string  `json:"policy_type"`
	SumAssured              float64 `json:"sum_assured"`
	PolicyTerm              int     `json:"policy_term"`
	PremiumCeasingAge       int     `json:"premium_ceasing_age"`
	PremiumPaymentFrequency string  `json:"premium_payment_frequency"`
	EntryPath               string  `json:"entry_path"`
	Channel                 string  `json:"channel"`
	Status                  string  `json:"status"`
	CurrentStage            string  `json:"current_stage"`
	BasePremium             float64 `json:"base_premium"`
	TotalPremium            float64 `json:"total_premium"`
	GSTAmount               float64 `json:"gst_amount"`
}

func FetchProposalResponse(d *domain.ProposalOutput) *ProposalResponse {

	if d == nil {
		return nil
	}

	quoteRef := ""
	if d.QuoteRefNumber != nil {
		quoteRef = *d.QuoteRefNumber
	}

	customerID := int64(0)
	if d.CustomerID != nil {
		customerID = *d.CustomerID
	}

	proposerCustomerID := int64(0)
	if d.ProposerCustomerID != nil {
		proposerCustomerID = *d.ProposerCustomerID
	}

	premiumCeasingAge := 0
	if d.PremiumCeasingAge != nil {
		premiumCeasingAge = *d.PremiumCeasingAge
	}

	return &ProposalResponse{
		ProposalID:              d.ProposalID,
		ProposalNumber:          d.ProposalNumber,
		InsurantName:            d.InsurantName,
		QuoteRefNumber:          quoteRef,
		CustomerID:              customerID,
		ProposerCustomerID:      proposerCustomerID,
		IsProposerSameAsInsured: d.IsProposerSameAsInsured,
		PremiumPayerType:        string(d.PremiumPayerType),
		ProductCode:             d.ProductCode,
		PolicyType:              string(d.PolicyType),
		SumAssured:              d.SumAssured,
		PolicyTerm:              d.PolicyTerm,
		PremiumCeasingAge:       premiumCeasingAge,
		PremiumPaymentFrequency: string(d.PremiumPaymentFrequency),
		EntryPath:               string(d.EntryPath),
		Channel:                 string(d.Channel),
		Status:                  string(d.Status),
		CurrentStage:            d.CurrentStage,
		BasePremium:             d.BasePremium,
		TotalPremium:            d.TotalPremium,
		GSTAmount:               d.GSTAmount,
	}
}

// AuditLogsResponse represents the response for audit log retrieval APIs
type AuditLogsResponse struct {
	port.StatusCodeAndMessage
	AuditLogs []AuditLog `json:"audit_logs"`
}

// AuditLog represents a single audit log entry in the response
type AuditLog struct {
	AuditID      int64   `json:"audit_id"`
	ProposalID   int64   `json:"proposal_id"`
	EntityType   string  `json:"entity_type"`
	EntityID     int64   `json:"entity_id"`
	FieldName    string  `json:"field_name"`
	OldValue     *string `json:"old_value,omitempty"`
	NewValue     *string `json:"new_value,omitempty"`
	ChangeType   string  `json:"change_type"`
	ChangedBy    int64   `json:"changed_by"`
	ChangedAt    string  `json:"changed_at"`
	ChangeReason *string `json:"change_reason,omitempty"`
	Metadata     *string `json:"metadata,omitempty"`
}
