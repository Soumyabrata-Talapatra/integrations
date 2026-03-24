package port

import "time"

// EmptyRequest represents a request with no parameters
type EmptyRequest struct{}

// PolicyNumberUri represents policy number in URI
type PolicyNumberUri struct {
	PolicyNumber string `uri:"policy_number" validate:"required,len=13"`
}

// TicketIdUri represents ticket ID in URI
type TicketIdUri struct {
	TicketID string `uri:"ticket_id" validate:"required"`
}

// IndexRevivalRequest represents request to index revival
// Note: MissingDocumentsList is NOT part of indexing - it's sent by data entry/QC/approver only when not approved
type IndexRevivalRequest struct {
	PolicyNumber       string                 `json:"policy_number" validate:"required,len=13"`
	CustomerID         string                 `json:"customer_id" validate:"required"`
	CustomerName       string                 `json:"customer_name" validate:"required"`
	RequestDateTime    time.Time              `json:"request_date_time"`
	RequestOwner       string                 `json:"request_owner" validate:"required"`
	IndexedBy          string                 `json:"indexed_by" validate:"required"`
	OfficeID           string                 `json:"office_id" validate:"required"`
	RevivalEligibility RevivalEligibilityInfo `json:"revival_eligibility"`
	Documents          []DocumentSubmission   `json:"documents"`
}

type RevivalEligibilityInfo struct {
	Eligible           bool `json:"eligible"`
	RevivalCount       int  `json:"revival_count"`
	MaxRevivalsAllowed int  `json:"max_revivals_allowed"`
	MaxInstallments    int  `json:"max_installments"`
}

// DataEntryRequest represents data entry request
type DataEntryRequest struct {
	TicketID            string               `json:"ticket_id" validate:"required"`
	PolicyNumber        string               `json:"policy_number" validate:"required"`
	CustomerID          string               `json:"customer_id" validate:"required"`
	CustomerName        string               `json:"customer_name" validate:"required"`
	Product             string               `json:"product" validate:"required"`
	SumAssured          float64              `json:"sum_assured" validate:"required"`
	PremiumAmount       float64              `json:"premium_amount" validate:"required"`
	LapseDate           time.Time            `json:"lapse_date" validate:"required"`
	OfficeID            string               `json:"office_id" validate:"required"`
	RevivalDetails      RevivalDetails       `json:"revival_details" validate:"required"`
	MedicalExaminerCode string               `json:"medical_examiner_code"`
	MedicalExaminerName string               `json:"medical_examiner_name"`
	DocumentsSubmitted  []DocumentSubmission `json:"documents_submitted"`
	MissingDocuments    []MissingDocument    `json:"missing_documents"`
	DataEnteredBy       string               `json:"data_entered_by" validate:"required"`
	DataEntryTimestamp  time.Time            `json:"data_entry_timestamp"`
}

type RevivalDetails struct {
	RevivalType         string       `json:"revival_type" validate:"required,oneof=installment lumpsum"`
	Installments        int          `json:"installments" validate:"required,min=1,max=12"`
	RevivalAmount       float64      `json:"revival_amount" validate:"required"`
	InstallmentAmount   float64      `json:"installment_amount" validate:"required"`
	TotalAmountDue      float64      `json:"total_amount_due" validate:"required"`
	TaxBreakdown        TaxBreakdown `json:"tax_breakdown"`
	Interest            float64      `json:"interest"`
	MedicalExaminerCode string       `json:"medical_examiner_code"`
	MedicalExaminerName string       `json:"medical_examiner_name"`
}

type DocumentSubmission struct {
	DocumentType string `json:"document_type"`
	DocumentID   string `json:"document_id"`
}

// MissingDocument represents a document that is missing/not received
type MissingDocument struct {
	DocumentName string `json:"document_name"` // Name of the missing document (e.g., "ID_PROOF", "ADDRESS_PROOF")
	Remarks      string `json:"remarks"`       // Optional remarks about why it's missing or when expected
}

// QualityCheckRequest represents quality check request
type QualityCheckRequest struct {
	TicketID         string            `json:"ticket_id" validate:"required"`
	QCPassed         bool              `json:"qc_passed"`
	QCComments       string            `json:"qc_comments" validate:"required"`
	QCPerformedBy    string            `json:"qc_performed_by" validate:"required"`
	QCTimestamp      time.Time         `json:"qc_timestamp"`
	OfficeID         string            `json:"office_id" validate:"required"`
	MissingDocuments []MissingDocument `json:"missing_documents"`
}

// ApprovalRequest represents approval request
// ApprovalDecisionRequest represents unified approval/rejection decision
type ApprovalDecisionRequest struct {
	TicketID    string    `json:"ticket_id" validate:"required"`
	Approved    bool      `json:"approved"` // true = approve, false = reject
	Comments    string    `json:"comments" validate:"required"`
	PerformedBy string    `json:"performed_by" validate:"required"`
	Timestamp   time.Time `json:"timestamp"`
	OfficeID    string    `json:"office_id" validate:"required"`
}

// ApprovalRedirectRequest represents approver redirecting request to earlier stage
type ApprovalRedirectRequest struct {
	TicketID    string    `json:"ticket_id" validate:"required"`
	RedirectTo  string    `json:"redirect_to" validate:"required,oneof=DATA_ENTRY QC"` // DATA_ENTRY or QC
	Comments    string    `json:"comments"`                                            // Optional comments
	PerformedBy string    `json:"performed_by" validate:"required"`
	Timestamp   time.Time `json:"timestamp"`
	OfficeID    string    `json:"office_id" validate:"required"`
}

// GetRevivalRequestUri is the URI parameter for getting a single revival request
type GetRevivalRequestUri struct {
	TicketID string `uri:"ticket_id" validate:"required"`
}

// WithdrawalRequest represents withdrawal request
type WithdrawalRequest struct {
	TicketID         string    `json:"ticket_id" validate:"required"`
	PolicyNumber     string    `json:"policy_number" validate:"required,len=13"`
	WithdrawalReason string    `json:"withdrawal_reason" validate:"required"`
	WithdrawalDate   time.Time `json:"withdrawal_date"`
	OfficeID         string    `json:"office_id" validate:"required"`
}

// FirstCollectionRequest represents first installment collection request (single record)
type FirstCollectionRequest struct {
	TicketID          string         `json:"ticket_id" validate:"required"`
	PolicyNumber      string         `json:"policy_number" validate:"required,len=13"`
	CollectionDate    time.Time      `json:"collection_date" validate:"required"`
	PaymentMode       string         `json:"payment_mode" validate:"required,oneof=CASH CHEQUE NEFT RTGS UPI CARD"`
	PremiumAmount     float64        `json:"premium_amount" validate:"required"`
	InstallmentAmount float64        `json:"installment_amount" validate:"required"`
	TotalAmount       float64        `json:"total_amount" validate:"required"`
	ChequeDetails     *ChequeDetails `json:"cheque_details"`
	Interest          float64        `json:"interest"`
	PremiumSGST       float64        `json:"premium_sgst"`
	PremiumCGST       float64        `json:"premium_cgst"`
	InstallmentSGST   float64        `json:"installment_sgst"`
	InstallmentCGST   float64        `json:"installment_cgst"`
	Rebate            float64        `json:"rebate"`
	OfficeID          string         `json:"office_id"`
	UserID            string         `json:"user_id"`
	NEFTDetails       *NEFTDetails   `json:"neft_details"`
}

// BatchFirstCollectionRequest represents batch first collection request
type BatchFirstCollectionRequest struct {
	Collections []FirstCollectionRequest `json:"collections" validate:"required,min=1,dive"`
}

type ChequeDetails struct {
	ChequeNumber string    `json:"cheque_number" validate:"required"`
	BankName     string    `json:"bank_name" validate:"required"`
	ChequeDate   time.Time `json:"cheque_date" validate:"required"`
	Amount       float64   `json:"amount" validate:"required"`
}

// NEFTDetails holds NEFT payment information
type NEFTDetails struct {
	URTNumber string    `json:"urt_number"`
	Date      time.Time `json:"date"`
	Bank      string    `json:"bank"`
	IFSCCode  string    `json:"ifsc_code"`
}

// QuotationRequest represents quotation request
type QuotationRequest struct {
	PolicyNumber string     `json:"policy_number" validate:"required,len=13"`
	Installments *int       `json:"installments" validate:"required,min=0,max=12"`
	QuoteDate    *time.Time `json:"quote_date"` // Optional: date for which quote should be calculated (defaults to today)
}

// CreateInstallmentRequest represents installment creation request (single record)
type CreateInstallmentRequest struct {
	TicketID             string       `json:"ticket_id" validate:"required"`
	PolicyNumber         string       `json:"policy_number" validate:"required"`
	InstallmentNumber    int          `json:"installment_number" validate:"required,min=2,max=12"`
	InstallmentAmount    float64      `json:"installment_amount" validate:"required"`
	CollectionDate       time.Time    `json:"collection_date" validate:"required"`
	PaymentMode          string       `json:"payment_mode" validate:"required,oneof=CASH CHEQUE NEFT RTGS UPI CARD"`
	Status               string       `json:"status" validate:"required,oneof=PENDING PAID DEFAULTED"`
	Interest             float64      `json:"interest"`
	SGST                 float64      `json:"sgst"`
	CGST                 float64      `json:"cgst"`
	Rebate               float64      `json:"rebate"`
	OfficeID             string       `json:"office_id"`
	UserID               string       `json:"user_id"`
	NEFTDetails          *NEFTDetails `json:"neft_details"`
	NumberOfInstallments int          `json:"number_of_installments"`
}

// BatchInstallmentRequest represents batch installment creation request
type BatchInstallmentRequest struct {
	Installments []CreateInstallmentRequest `json:"installments" validate:"required,min=1,dive"`
}

// ReceiveDocumentsRequest represents document receipt request
type ReceiveDocumentsRequest struct {
	TicketID             string            `json:"ticket_id" validate:"required"`
	DocumentsReceived    []DocumentReceipt `json:"documents_received" validate:"required"`
	AllDocumentsReceived bool              `json:"all_documents_received" validate:"required"`
	OfficeID             string            `json:"office_id" validate:"required"`
}

type DocumentReceipt struct {
	DocumentType string    `json:"document_type" validate:"required"`
	DocumentName string    `json:"document_name" validate:"required"`
	ReceivedDate time.Time `json:"received_date" validate:"required"`
	ReceivedBy   string    `json:"received_by" validate:"required"`
}

// RevivalAcceptanceLetterRequest represents acceptance letter generation request
type RevivalAcceptanceLetterRequest struct {
	TicketID         string         `json:"ticket_id" validate:"required"`
	PolicyNumber     string         `json:"policy_number" validate:"required,len=13"`
	PolicyholderName string         `json:"policyholder_name" validate:"required"`
	Address          string         `json:"address" validate:"required"`
	ApprovalDate     time.Time      `json:"approval_date" validate:"required"`
	RevivalDetails   RevivalDetails `json:"revival_details" validate:"required"`
	OfficeID         string         `json:"office_id" validate:"required"`
}

// RevivalMemoRequest represents revival memo generation request
type RevivalMemoRequest struct {
	TicketID       string         `json:"ticket_id" validate:"required"`
	PolicyNumber   string         `json:"policy_number" validate:"required,len=13"`
	OfficeID       string         `json:"office_id" validate:"required"`
	ApprovalDate   time.Time      `json:"approval_date" validate:"required"`
	RevivalSummary RevivalSummary `json:"revival_summary" validate:"required"`
}

type RevivalSummary struct {
	RevivalAmount      float64   `json:"revival_amount" validate:"required"`
	TotalInstallments  int       `json:"total_installments" validate:"required"`
	InstallmentAmount  float64   `json:"installment_amount" validate:"required"`
	DualCollectionDate time.Time `json:"dual_collection_date" validate:"required"`
}

// GetInstallmentCollectionRequest represents request to get installment collection details
type GetInstallmentCollectionRequest struct {
	NumberOfInstallments *int   `form:"number_of_installments"` // Optional, can be nil
	PolicyNumber         string `form:"policy_number" validate:"required"`
	OfficeID             string `form:"office_id"`
}
