package port

import "time"

// PolicyEligibilityResponse represents policy eligibility check response
type PolicyEligibilityResponse struct {
	StatusCodeAndMessage
	Data PolicyEligibilityData `json:"data"`
}

type PolicyEligibilityData struct {
	PolicyNumber             string  `json:"policy_number"`
	Eligible                 bool    `json:"eligible"`
	RevivalCount             int     `json:"revival_count"`
	MaxRevivalsAllowed       int     `json:"max_revivals_allowed"`
	MaxInstallments          int     `json:"max_installments"`
	RevivalEligibilityReason *string `json:"revival_eligibility_reason"`
}

// PolicyDetailsResponse represents policy details response
type PolicyDetailsResponse struct {
	StatusCodeAndMessage
	Data PolicyDetailsData `json:"data"`
}

type PolicyDetailsData struct {
	PolicyNumber       string     `json:"policy_number"`
	CustomerName       string     `json:"customer_name"`
	CustomerID         string     `json:"customer_id"`
	Product            string     `json:"product"`
	ProductCode        string     `json:"product_code"`
	SumAssured         float64    `json:"sum_assured"`
	PremiumAmount      float64    `json:"premium_amount"`
	PremiumFrequency   string     `json:"premium_frequency"`
	LapseDate          *time.Time `json:"lapse_date"`
	RevivalCount       int        `json:"revival_count"`
	PolicyStatus       string     `json:"policy_status"`
	DateOfCommencement time.Time  `json:"date_of_commencement"`
	MaturityDate       time.Time  `json:"maturity_date"`
	PaidToDate         *time.Time `json:"paid_to_date"`
	LastRevivalDate    *time.Time `json:"last_revival_date"`
}

// RevivalHistoryResponse represents revival history response
type RevivalHistoryResponse struct {
	StatusCodeAndMessage
	Data RevivalHistoryData `json:"data"`
}

type RevivalHistoryData struct {
	PolicyNumber           string                `json:"policy_number"`
	RevivalCount           int                   `json:"revival_count"`
	CompletedRevivalsCount int                   `json:"completed_revivals_count"`
	RevivalHistory         []RevivalHistoryEntry `json:"revival_history"`
}

type RevivalHistoryEntry struct {
	TicketID          string    `json:"ticket_id"`
	RequestDate       time.Time `json:"request_date"`
	Status            string    `json:"status"`
	InstallmentsPaid  int       `json:"installments_paid"`
	TotalInstallments int       `json:"total_installments"`
	RevivalAmount     float64   `json:"revival_amount"`
	InstallmentAmount float64   `json:"installment_amount"`
}

// IndexRequestResponse represents indexing response
type IndexRequestResponse struct {
	StatusCodeAndMessage
	Data IndexRequestData `json:"data"`
}

type IndexRequestData struct {
	TicketID        string    `json:"ticket_id"`
	WorkflowID      *string   `json:"workflow_id,omitempty"`
	Status          string    `json:"status"`
	RequestDateTime time.Time `json:"request_date_time"`
	Message         string    `json:"message"`
}

// QuotationResponse represents quotation response
type QuotationResponse struct {
	StatusCodeAndMessage
	Data QuotationData `json:"data"`
}

type QuotationData struct {
	PolicyNumber      string       `json:"policy_number"`
	RevivalType       string       `json:"revival_type"`
	Installments      int          `json:"installments"`
	RevivalAmount     float64      `json:"revival_amount"`
	Interest          float64      `json:"interest"`
	InstallmentAmount float64      `json:"installment_amount"`
	TaxBreakdown      TaxBreakdown `json:"tax_breakdown"`
	TotalAmountDue    float64      `json:"total_amount_due"`
	QuoteValidUntil   time.Time    `json:"quote_valid_until"`
	TotalPremiumDue   float64      `json:"total_premium_due"`
	FromDueDate       *time.Time   `json:"from_due_date,omitempty"`
	ToDueDate         time.Time    `json:"to_due_date"`
	Premium           float64      `json:"premium"`
}

type TaxBreakdown struct {
	CGST     float64 `json:"cgst"`
	SGST     float64 `json:"sgst"`
	TotalTax float64 `json:"total_tax"`
}

// DataEntryResponse represents data entry response
type DataEntryResponse struct {
	StatusCodeAndMessage
	Data DataEntryData `json:"data"`
}

type DataEntryData struct {
	TicketID           string    `json:"ticket_id"`
	Status             string    `json:"status"`
	Message            string    `json:"message"`
	DataEntryTimestamp time.Time `json:"data_entry_timestamp"`
}

// QualityCheckResponse represents quality check response
type QualityCheckResponse struct {
	StatusCodeAndMessage
	Data QualityCheckData `json:"data"`
}

type QualityCheckData struct {
	TicketID    string    `json:"ticket_id"`
	Status      string    `json:"status"`
	Message     string    `json:"message"`
	QCTimestamp time.Time `json:"qc_timestamp"`
}

// ApprovalDecisionResponse represents unified approval/rejection response
type ApprovalDecisionResponse struct {
	StatusCodeAndMessage
	Data ApprovalDecisionData `json:"data"`
}

type ApprovalDecisionData struct {
	TicketID         string     `json:"ticket_id"`
	Approved         bool       `json:"approved"`
	Status           string     `json:"status"` // "APPROVED" or "REJECTED"
	Message          string     `json:"message"`
	Timestamp        time.Time  `json:"timestamp"`
	SLAEndDate       *time.Time `json:"sla_end_date,omitempty"`       // Only for approvals
	SLARemainingDays *int       `json:"sla_remaining_days,omitempty"` // Only for approvals
}

// ApprovalRedirectResponse represents response when approver redirects to earlier stage
type ApprovalRedirectResponse struct {
	StatusCodeAndMessage
	Data ApprovalRedirectData `json:"data"`
}

type ApprovalRedirectData struct {
	TicketID     string    `json:"ticket_id"`
	RedirectedTo string    `json:"redirected_to"` // "DATA_ENTRY_PENDING" or "DATA_ENTRY_COMPLETE"
	Status       string    `json:"status"`
	Message      string    `json:"message"`
	Timestamp    time.Time `json:"timestamp"`
}

// GetRevivalRequestResponse returns detailed revival request based on workflow stage
type GetRevivalRequestResponse struct {
	StatusCodeAndMessage
	Data RevivalRequestDetails `json:"data"`
}

type RevivalRequestDetails struct {
	// Basic Information (available at all stages)
	RequestID     string    `json:"request_id"`
	TicketID      string    `json:"ticket_id"`
	PolicyNumber  string    `json:"policy_number"`
	CustomerName  string    `json:"customer_name,omitempty"`
	CustomerID    string    `json:"customer_id,omitempty"`
	RequestType   string    `json:"request_type"`
	CurrentStatus string    `json:"current_status"`
	CreatedAt     time.Time `json:"created_at"`
	WorkflowID    *string   `json:"workflow_id,omitempty"`
	RunID         *string   `json:"run_id,omitempty"`

	// Indexing Information (available from indexing stage)
	IndexingDetails *IndexingDetails `json:"indexing_details,omitempty"`

	// Data Entry Information (available from data entry stage)
	DataEntryDetails *DataEntryDetails `json:"data_entry_details,omitempty"`

	// QC Information (available from QC stage)
	QCDetails *QCDetails `json:"qc_details,omitempty"`

	// Approval Information (available from approval stage)
	ApprovalDetails *ApprovalDetails `json:"approval_details,omitempty"`

	// Status History (available at all stages)
	StatusHistory []StatusHistoryEntry `json:"status_history,omitempty"`
}

type IndexingDetails struct {
	IndexedBy   string               `json:"indexed_by"`
	IndexedDate time.Time            `json:"indexed_date"`
	Documents   []DocumentSubmission `json:"documents,omitempty"`
}

type DataEntryDetails struct {
	DataEnteredBy        string               `json:"data_entered_by,omitempty"`
	DataEntryTimestamp   *time.Time           `json:"data_entry_timestamp,omitempty"`
	NumberOfInstallments int                  `json:"number_of_installments,omitempty"`
	RevivalAmount        float64              `json:"revival_amount,omitempty"`
	InstallmentAmount    float64              `json:"installment_amount,omitempty"`
	SGST                 *float64             `json:"sgst,omitempty"`
	CGST                 *float64             `json:"cgst,omitempty"`
	Interest             *float64             `json:"interest,omitempty"`
	DocumentsSubmitted   []DocumentSubmission `json:"documents_submitted,omitempty"`
	MissingDocuments     []MissingDocument    `json:"missing_documents,omitempty"`
	MedicalExaminerCode  *string              `json:"medical_examiner_code,omitempty"`
	MedicalExaminerName  *string              `json:"medical_examiner_name,omitempty"`
}

type QCDetails struct {
	QCPassed         bool              `json:"qc_passed"`
	QCComments       string            `json:"qc_comments,omitempty"`
	QCPerformedBy    string            `json:"qc_performed_by,omitempty"`
	QCTimestamp      *time.Time        `json:"qc_timestamp,omitempty"`
	MissingDocuments []MissingDocument `json:"missing_documents,omitempty"`
}

type ApprovalDetails struct {
	Approved          bool       `json:"approved"`
	ApprovedBy        string     `json:"approved_by,omitempty"`
	ApprovalTimestamp *time.Time `json:"approval_timestamp,omitempty"`
	Comments          string     `json:"comments,omitempty"`
	SLAStartDate      *time.Time `json:"sla_start_date,omitempty"`
	SLAEndDate        *time.Time `json:"sla_end_date,omitempty"`
}

type StatusHistoryEntry struct {
	HistoryID    string    `json:"history_id"`
	FromStatus   *string   `json:"from_status,omitempty"`
	ToStatus     string    `json:"to_status"`
	ChangedAt    time.Time `json:"changed_at"`
	ChangedBy    string    `json:"changed_by"`
	ChangeReason *string   `json:"change_reason,omitempty"`
}

// FirstCollectionResponse represents first collection response
type FirstCollectionResponse struct {
	StatusCodeAndMessage
	Data FirstCollectionData `json:"data"`
}

type FirstCollectionData struct {
	ReceiptNumber     string     `json:"receipt_number"`
	TicketID          string     `json:"ticket_id"`
	PolicyNumber      string     `json:"policy_number"`
	PremiumAmount     float64    `json:"premium_amount"`
	InstallmentAmount float64    `json:"installment_amount"`
	PremiumSGST       float64    `json:"premium_sgst"`
	PremiumCGST       float64    `json:"premium_cgst"`
	InstallmentSGST   float64    `json:"installment_sgst"`
	InstallmentCGST   float64    `json:"installment_cgst"`
	TotalGST          float64    `json:"total_gst"`
	TotalAmount       float64    `json:"total_amount"`
	ReceiptDate       time.Time  `json:"receipt_date"`
	PaymentMode       string     `json:"payment_mode"`
	Status            string     `json:"status"`
	DueDate           *time.Time `json:"due_date,omitempty"`
}

// BatchFirstCollectionResponse represents batch first collection response
type BatchFirstCollectionResponse struct {
	StatusCodeAndMessage
	TotalSubmitted int                         `json:"total_submitted"`
	Successful     int                         `json:"successful"`
	Failed         int                         `json:"failed"`
	Results        []FirstCollectionResultItem `json:"results"`
}

type BatchFirstCollectionData struct {
	TotalSubmitted int                         `json:"total_submitted"`
	Successful     int                         `json:"successful"`
	Failed         int                         `json:"failed"`
	Results        []FirstCollectionResultItem `json:"results"`
}

type FirstCollectionResultItem struct {
	TicketID          string     `json:"ticket_id"`
	PolicyNumber      string     `json:"policy_number"`
	ReceiptNumber     string     `json:"receipt_number,omitempty"`
	PremiumAmount     float64    `json:"premium_amount,omitempty"`
	InstallmentAmount float64    `json:"installment_amount,omitempty"`
	PremiumSGST       float64    `json:"premium_sgst,omitempty"`
	PremiumCGST       float64    `json:"premium_cgst,omitempty"`
	InstallmentSGST   float64    `json:"installment_sgst,omitempty"`
	InstallmentCGST   float64    `json:"installment_cgst,omitempty"`
	TotalGST          float64    `json:"total_gst,omitempty"`
	TotalAmount       float64    `json:"total_amount,omitempty"`
	ReceiptDate       time.Time  `json:"receipt_date,omitempty"`
	PaymentMode       string     `json:"payment_mode,omitempty"`
	DueDate           *time.Time `json:"due_date,omitempty"`
	Success           bool       `json:"success"`
	Status            string     `json:"status"`
	Message           string     `json:"message,omitempty"`
	ErrorMessage      string     `json:"error_message,omitempty"`
}

// WithdrawalResponse represents withdrawal response
type WithdrawalResponse struct {
	StatusCodeAndMessage
	Data WithdrawalData `json:"data"`
}

type WithdrawalData struct {
	TicketID            string               `json:"ticket_id"`
	PolicyNumber        string               `json:"policy_number"`
	Status              string               `json:"status"`
	WithdrawalDate      time.Time            `json:"withdrawal_date"`
	WithdrawalReason    string               `json:"withdrawal_reason"`
	SuspenseAdjustments []SuspenseAdjustment `json:"suspense_adjustments"`
	Message             string               `json:"message"`
}

type SuspenseAdjustment struct {
	SuspenseID        string  `json:"suspense_id"`
	Amount            float64 `json:"amount"`
	ReversedToPremium bool    `json:"reversed_to_premium"`
}

// InstallmentResponse represents installment response
type InstallmentResponse struct {
	StatusCodeAndMessage
	Data InstallmentData `json:"data"`
}

type InstallmentData struct {
	ReceiptNumber     string  `json:"receipt_number"`
	TicketID          string  `json:"ticket_id"`
	PolicyNumber      string  `json:"policy_number"`
	InstallmentNumber int     `json:"installment_number"`
	InstallmentAmount float64 `json:"installment_amount"`
	Status            string  `json:"status"`
}

// BatchInstallmentResponse represents batch installment creation response
type BatchInstallmentResponse struct {
	StatusCodeAndMessage
	TotalSubmitted int                     `json:"total_submitted"`
	Successful     int                     `json:"successful"`
	Failed         int                     `json:"failed"`
	Results        []InstallmentResultItem `json:"results"`
}

type BatchInstallmentData struct {
	TotalSubmitted int                     `json:"total_submitted"`
	Successful     int                     `json:"successful"`
	Failed         int                     `json:"failed"`
	Results        []InstallmentResultItem `json:"results"`
}

type InstallmentResultItem struct {
	TicketID          string  `json:"ticket_id"`
	PolicyNumber      string  `json:"policy_number"`
	InstallmentNumber int     `json:"installment_number"`
	ReceiptNumber     string  `json:"receipt_number,omitempty"`
	InstallmentAmount float64 `json:"installment_amount,omitempty"`
	Success           bool    `json:"success"`
	Status            string  `json:"status"`
	Message           string  `json:"message,omitempty"`
	ErrorMessage      string  `json:"error_message,omitempty"`
}

// InstallmentRevivalHistoryResponse represents full installment revival history
type InstallmentRevivalHistoryResponse struct {
	StatusCodeAndMessage
	Data InstallmentRevivalHistoryData `json:"data"`
}

type InstallmentRevivalHistoryData struct {
	PolicyNumber      string                           `json:"policy_number"`
	TotalRevivals     int                              `json:"total_revivals"`
	CompletedRevivals int                              `json:"completed_revivals"`
	DefaultedRevivals int                              `json:"defaulted_revivals"`
	RevivalHistory    []InstallmentRevivalHistoryEntry `json:"revival_history"`
}

type InstallmentRevivalHistoryEntry struct {
	TicketID          string            `json:"ticket_id"`
	IndexDate         time.Time         `json:"index_date"`
	ApprovalDate      *time.Time        `json:"approval_date"`
	CompletionDate    *time.Time        `json:"completion_date"`
	Status            string            `json:"status"`
	InstallmentPlan   InstallmentPlan   `json:"installment_plan"`
	CollectionSummary CollectionSummary `json:"collection_summary"`
}

type InstallmentPlan struct {
	TotalInstallments int     `json:"total_installments"`
	InstallmentAmount float64 `json:"installment_amount"`
	TotalAmount       float64 `json:"total_amount"`
}

type CollectionSummary struct {
	FirstCollectionDate time.Time `json:"first_collection_date"`
	InstallmentsPaid    int       `json:"installments_paid"`
	TotalCollected      float64   `json:"total_collected"`
}

// GetAllRequestsResponse represents response for getting all revival requests
type GetAllRequestsResponse struct {
	StatusCodeAndMessage
	Data []RequestListItem `json:"data"`
}

type RequestListItem struct {
	RequestID     string     `json:"request_id"`
	TicketID      string     `json:"ticket_id"`
	PolicyNumber  string     `json:"policy_number"`
	InsuredName   string     `json:"insured_name"`
	CustomerID    string     `json:"customer_id"`
	RequestType   string     `json:"request_type"`
	RequestStatus string     `json:"request_status"`
	RequestedDate *time.Time `json:"requested_date"`
	NextAction    string     `json:"next_action"`
	NextActor     string     `json:"next_actor"`
	WorkflowState string     `json:"workflow_state,omitempty"`
}

// Common error status codes for revival
var (
	PolicyNotFound           StatusCodeAndMessage = StatusCodeAndMessage{StatusCode: 404, Success: false, Message: "Policy not found"}
	PolicyNotEligible        StatusCodeAndMessage = StatusCodeAndMessage{StatusCode: 422, Success: false, Message: "Policy not eligible for revival"}
	OngoingRevivalExists     StatusCodeAndMessage = StatusCodeAndMessage{StatusCode: 409, Success: false, Message: "Ongoing revival request exists"}
	InvalidTicketStatus      StatusCodeAndMessage = StatusCodeAndMessage{StatusCode: 422, Success: false, Message: "Invalid ticket status"}
	MissingRequiredFields    StatusCodeAndMessage = StatusCodeAndMessage{StatusCode: 422, Success: false, Message: "Missing required fields"}
	TicketNotReadyForQC      StatusCodeAndMessage = StatusCodeAndMessage{StatusCode: 422, Success: false, Message: "Ticket not ready for Quality Check"}
	InvalidApprovalStatus    StatusCodeAndMessage = StatusCodeAndMessage{StatusCode: 422, Success: false, Message: "Invalid ticket status for approval"}
	InvalidRejectionStatus   StatusCodeAndMessage = StatusCodeAndMessage{StatusCode: 422, Success: false, Message: "Cannot reject request in current status"}
	MaxRevivalsExceeded      StatusCodeAndMessage = StatusCodeAndMessage{StatusCode: 422, Success: false, Message: "Maximum revivals allowed exceeded"}
	MaxInstallmentExceeded   StatusCodeAndMessage = StatusCodeAndMessage{StatusCode: 422, Success: false, Message: "Installment number exceeds approved count"}
	OutOfOrderPayment        StatusCodeAndMessage = StatusCodeAndMessage{StatusCode: 422, Success: false, Message: "OUT-OF-ORDER PAYMENT: Must pay installments sequentially"}
	InvalidInstallmentNumber StatusCodeAndMessage = StatusCodeAndMessage{StatusCode: 422, Success: false, Message: "Invalid installment number: must be >= 2"}
	InvalidInstallmentAmount StatusCodeAndMessage = StatusCodeAndMessage{StatusCode: 422, Success: false, Message: "Invalid installment should match the configured amount"}
	PolicyNotActiveStatus    StatusCodeAndMessage = StatusCodeAndMessage{StatusCode: 422, Success: false, Message: "Please ensure the policy is in Active status before revival."}
	InstallmentNotFound      StatusCodeAndMessage = StatusCodeAndMessage{StatusCode: 404, Success: false, Message: "Installment collection not found"}
)

// GetInstallmentCollectionResponse represents installment collection details response
type GetInstallmentCollectionResponse struct {
	StatusCodeAndMessage
	Data InstallmentCollectionData `json:"data"`
}

type InstallmentCollectionData struct {
	TicketID             string  `json:"ticket_id"`
	PolicyNumber         string  `json:"policy_number"`
	CustomerName         string  `json:"customer_name"`
	RevivalType          *string `json:"revival_type,omitempty"`
	InstallmentNumber    int     `json:"installment_number"`
	NumberOfInstallments int     `json:"number_of_installments"`
	TotalInstallments    int     `json:"total_installments"`

	// First collection specific (installment 1)
	PremiumAmount     *float64 `json:"premium_amount,omitempty"`
	InstallmentAmount float64  `json:"installment_amount"`

	// Tax breakdown for first installment (separate GST for premium and installment)
	PremiumCGST     *float64 `json:"premium_cgst,omitempty"`     // 9% on premium (first installment only)
	PremiumSGST     *float64 `json:"premium_sgst,omitempty"`     // 9% on premium (first installment only)
	InstallmentCGST *float64 `json:"installment_cgst,omitempty"` // 9% on installment (first installment only)
	InstallmentSGST *float64 `json:"installment_sgst,omitempty"` // 9% on installment (first installment only)

	// Total amounts
	TotalAmount float64 `json:"total_amount"` // Premium + Installment + GST
}
