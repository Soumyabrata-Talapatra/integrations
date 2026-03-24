package domain

import "time"

// RevivalRequest represents a revival request in the system
type RevivalRequest struct {
	RequestID              string     `json:"request_id" db:"request_id"`
	TicketID               string     `json:"ticket_id" db:"ticket_id"`
	PolicyNumber           string     `json:"policy_number" db:"policy_number"`
	RequestType            string     `json:"request_type" db:"request_type"`
	RevivalType            *string    `json:"revival_type" db:"revival_type"`
	CurrentStatus          string     `json:"current_status" db:"current_status"`
	WorkflowID             *string    `json:"workflow_id" db:"workflow_id"`
	RunID                  *string    `json:"run_id" db:"run_id"`
	IndexedDate            *time.Time `json:"indexed_date" db:"indexed_date"`
	IndexedBy              *string    `json:"indexed_by" db:"indexed_by"`
	DataEntryDate          *time.Time `json:"data_entry_date" db:"data_entry_date"`
	DataEntryBy            *string    `json:"data_entry_by" db:"data_entry_by"`
	QCCompleteDate         *time.Time `json:"qc_complete_date" db:"qc_complete_date"`
	QCBy                   *string    `json:"qc_by" db:"qc_by"`
	QCComments             *string    `json:"qc_comments" db:"qc_comments"`
	ApprovalDate           *time.Time `json:"approval_date" db:"approval_date"`
	ApprovedBy             *string    `json:"approved_by" db:"approved_by"`
	ApprovalComments       *string    `json:"approval_comments" db:"approval_comments"`
	CompletionDate         *time.Time `json:"completion_date" db:"completion_date"`
	TerminationDate        *time.Time `json:"termination_date" db:"termination_date"`
	WithdrawalDate         *time.Time `json:"withdrawal_date" db:"withdrawal_date"`
	NumberOfInstallments   int        `json:"number_of_installments" db:"number_of_installments"`
	RevivalAmount          float64    `json:"revival_amount" db:"revival_amount"`
	InstallmentAmount      float64    `json:"installment_amount" db:"installment_amount"`
	TotalTaxOnUnpaid       float64    `json:"total_tax_on_unpaid" db:"total_tax_on_unpaid"`
	FirstCollectionDate    *time.Time `json:"first_collection_date" db:"first_collection_date"`
	FirstCollectionDone    bool       `json:"first_collection_done" db:"first_collection_done"`
	BlockingNewCollections bool       `json:"blocking_new_collections" db:"blocking_new_collections"`
	InstallmentsPaid       int        `json:"installments_paid" db:"installments_paid"`
	// Re-revival suspense adjustment fields
	PreviousSuspenseAmount float64   `json:"previous_suspense_amount" db:"previous_suspense_amount"` // Total suspense from previous revival
	SuspenseAdjusted       bool      `json:"suspense_adjusted" db:"suspense_adjusted"`               // Whether suspense was adjusted
	AdjustedRevivalAmount  float64   `json:"adjusted_revival_amount" db:"adjusted_revival_amount"`   // Revival amount after suspense adjustment
	RequestOwner           *string   `json:"request_owner" db:"request_owner"`
	MedicalExaminerCode    *string   `json:"medical_examiner_code" db:"medical_examiner_code"`
	MedicalExaminerName    *string   `json:"medical_examiner_name" db:"medical_examiner_name"`
	MissingDocumentsList   *string   `json:"missing_documents_list" db:"missing_documents_list"` // JSONB stored as string
	Documents              *string   `json:"documents" db:"documents"`                           // JSONB stored as string
	CreatedAt              time.Time `json:"created_at" db:"created_at"`
	UpdatedAt              time.Time `json:"updated_at" db:"updated_at"`
	SGST                   *float64  `json:"sgst" db:"sgst"`
	CGST                   *float64  `json:"cgst" db:"cgst"`
	Interest               *float64  `json:"interest" db:"interest"`
}

// RevivalRequestWithPolicy represents a revival request with policy details for listing
type RevivalRequestWithPolicy struct {
	RequestID     string     `json:"request_id" db:"request_id"`
	PolicyNumber  string     `json:"policy_number" db:"policy_number"`
	TicketID      string     `json:"ticket_id" db:"ticket_id"`
	InsuredName   string     `json:"insured_name" db:"insured_name"`
	CustomerID    string     `json:"customer_id" db:"customer_id"`
	RequestType   string     `json:"request_type" db:"request_type"`
	CurrentStatus string     `json:"current_status" db:"current_status"`
	RequestedDate *time.Time `json:"requested_date" db:"requested_date"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
}

// InstallmentSchedule represents an installment payment schedule
type InstallmentSchedule struct {
	ScheduleID        string     `json:"schedule_id" db:"schedule_id"`
	RequestID         string     `json:"request_id" db:"request_id"`
	PolicyNumber      string     `json:"policy_number" db:"policy_number"`
	InstallmentNumber int        `json:"installment_number" db:"installment_number"`
	InstallmentAmount float64    `json:"installment_amount" db:"installment_amount"`
	TaxAmount         float64    `json:"tax_amount" db:"tax_amount"`
	TotalAmount       float64    `json:"total_amount" db:"total_amount"`
	DueDate           time.Time  `json:"due_date" db:"due_date"`
	PaymentDate       *time.Time `json:"payment_date" db:"payment_date"`
	IsPaid            bool       `json:"is_paid" db:"is_paid"`
	GracePeriodDays   int        `json:"grace_period_days" db:"grace_period_days"`
	CreatedAt         time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at" db:"updated_at"`
}

// RevivalCalculation represents revival quotation calculations
type RevivalCalculation struct {
	CalculationID             string     `json:"calculation_id" db:"calculation_id"`
	RequestID                 string     `json:"request_id" db:"request_id"`
	PolicyNumber              string     `json:"policy_number" db:"policy_number"`
	DateOfRevival             time.Time  `json:"date_of_revival" db:"date_of_revival"`
	NumberOfInstallments      int        `json:"number_of_installments" db:"number_of_installments"`
	UnpaidPremiumMonths       int        `json:"unpaid_premium_months" db:"unpaid_premium_months"`
	InterestRate              float64    `json:"interest_rate" db:"interest_rate"`
	MonthlyPremium            float64    `json:"monthly_premium" db:"monthly_premium"`
	PremiumAmount             float64    `json:"premium_amount" db:"premium_amount"`
	TaxOnPremium              float64    `json:"tax_on_premium" db:"tax_on_premium"`
	TotalRenewalAmount        float64    `json:"total_renewal_amount" db:"total_renewal_amount"`
	InstallmentAmount         float64    `json:"installment_amount" db:"installment_amount"`
	TaxOnUnpaidPremium        float64    `json:"tax_on_unpaid_premium" db:"tax_on_unpaid_premium"`
	TotalInstallmentAmount    float64    `json:"total_installment_amount" db:"total_installment_amount"`
	GrandTotalFirstCollection float64    `json:"grand_total_first_collection" db:"grand_total_first_collection"`
	ValidUntil                *time.Time `json:"valid_until" db:"valid_until"`
	CalculatedAt              time.Time  `json:"calculated_at" db:"calculated_at"`
}

// Policy represents a policy record
type Policy struct {
	PolicyNumber       string     `json:"policy_number" db:"policy_number"`
	CustomerID         string     `json:"customer_id" db:"customer_id"`
	CustomerName       string     `json:"customer_name" db:"customer_name"`
	ProductCode        string     `json:"product_code" db:"product_code"`
	ProductName        string     `json:"product_name" db:"product_name"`
	PolicyStatus       string     `json:"policy_status" db:"policy_status"`
	PremiumFrequency   string     `json:"premium_frequency" db:"premium_frequency"`
	PremiumAmount      float64    `json:"premium_amount" db:"premium_amount"`
	SumAssured         float64    `json:"sum_assured" db:"sum_assured"`
	PaidToDate         *time.Time `json:"paid_to_date" db:"paid_to_date"`
	MaturityDate       time.Time  `json:"maturity_date" db:"maturity_date"`
	DateOfCommencement time.Time  `json:"date_of_commencement" db:"date_of_commencement"`
	RevivalCount       int        `json:"revival_count" db:"revival_count"`
	LastRevivalDate    *time.Time `json:"last_revival_date" db:"last_revival_date"`
	CreatedAt          time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time  `json:"updated_at" db:"updated_at"`
}

// PaymentTransaction represents a payment transaction
type PaymentTransaction struct {
	PaymentID             string     `json:"payment_id" db:"payment_id"`
	RequestID             string     `json:"request_id" db:"request_id"`
	PolicyNumber          string     `json:"policy_number" db:"policy_number"`
	CollectionBatchID     string     `json:"collection_batch_id" db:"collection_batch_id"`
	LinkedPaymentID       *string    `json:"linked_payment_id" db:"linked_payment_id"`
	ChequeID              *string    `json:"cheque_id" db:"cheque_id"`
	PaymentType           string     `json:"payment_type" db:"payment_type"`
	InstallmentNumber     *int       `json:"installment_number" db:"installment_number"`
	Amount                float64    `json:"amount" db:"amount"`
	TaxAmount             float64    `json:"tax_amount" db:"tax_amount"`
	TotalAmount           float64    `json:"total_amount" db:"total_amount"`
	PaymentMode           string     `json:"payment_mode" db:"payment_mode"`
	PaymentStatus         string     `json:"payment_status" db:"payment_status"`
	CollectionDate        time.Time  `json:"collection_date" db:"collection_date"`
	PaymentDate           *time.Time `json:"payment_date" db:"payment_date"`
	ReceiptID             *string    `json:"receipt_id" db:"receipt_id"`
	TigerBeetleTransferID *string    `json:"tigerbeetle_transfer_id" db:"tigerbeetle_transfer_id"`
	CollectedBy           *string    `json:"collected_by" db:"collected_by"`
	CreatedAt             time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt             time.Time  `json:"updated_at" db:"updated_at"`
}

// CollectionBatchTracking represents collection batch tracking
type CollectionBatchTracking struct {
	BatchID              string     `json:"batch_id" db:"batch_id"`
	RequestID            string     `json:"request_id" db:"request_id"`
	PolicyNumber         string     `json:"policy_number" db:"policy_number"`
	PremiumPaymentID     string     `json:"premium_payment_id" db:"premium_payment_id"`
	InstallmentPaymentID string     `json:"installment_payment_id" db:"installment_payment_id"`
	CollectionComplete   bool       `json:"collection_complete" db:"collection_complete"`
	CollectionDate       time.Time  `json:"collection_date" db:"collection_date"`
	CombinedReceiptID    *string    `json:"combined_receipt_id" db:"combined_receipt_id"`
	CreatedAt            time.Time  `json:"created_at" db:"created_at"`
	CompletedAt          *time.Time `json:"completed_at" db:"completed_at"`
}

// RevivalRequestWorkflowState represents Temporal workflow state
type RevivalRequestWorkflowState struct {
	RequestID           string     `json:"request_id" db:"request_id"`
	WorkflowID          string     `json:"workflow_id" db:"workflow_id"`
	RunID               string     `json:"run_id" db:"run_id"`
	CurrentStatus       string     `json:"current_status" db:"current_status"`
	WorkflowStatus      string     `json:"workflow_status" db:"workflow_status"`
	SLAStartDate        *time.Time `json:"sla_start_date" db:"sla_start_date"`
	SLAEndDate          *time.Time `json:"sla_end_date" db:"sla_end_date"`
	SLAExpired          bool       `json:"sla_expired" db:"sla_expired"`
	FirstCollectionDone bool       `json:"first_collection_done" db:"first_collection_done"`
	TotalInstallments   int        `json:"total_installments" db:"total_installments"`
	InstallmentsPaid    int        `json:"installments_paid" db:"installments_paid"`
	StartedAt           time.Time  `json:"started_at" db:"started_at"`
	CompletedAt         *time.Time `json:"completed_at" db:"completed_at"`
	LastUpdated         time.Time  `json:"last_updated" db:"last_updated"`
}

// StatusChangeHistory represents status change history
type StatusChangeHistory struct {
	HistoryID    string    `json:"history_id" db:"history_id"`
	RequestID    string    `json:"request_id" db:"request_id"`
	FromStatus   string    `json:"from_status" db:"from_status"`
	ToStatus     string    `json:"to_status" db:"to_status"`
	ChangedAt    time.Time `json:"changed_at" db:"changed_at"`
	ChangedBy    string    `json:"changed_by" db:"changed_by"`
	ChangeReason *string   `json:"change_reason" db:"change_reason"`
}

// TigerBeetleAccount represents TigerBeetle ledger account
type TigerBeetleAccount struct {
	AccountID                 string    `json:"account_id" db:"account_id"`
	PolicyNumber              string    `json:"policy_number" db:"policy_number"`
	PremiumAccountID          string    `json:"premium_account_id" db:"premium_account_id"`
	RevivalAccountID          string    `json:"revival_account_id" db:"revival_account_id"`
	LoanAccountID             string    `json:"loan_account_id" db:"loan_account_id"`
	CombinedSuspenseAccountID string    `json:"combined_suspense_account_id" db:"combined_suspense_account_id"`
	RevivalSuspenseAccountID  string    `json:"revival_suspense_account_id" db:"revival_suspense_account_id"`
	CreatedAt                 time.Time `json:"created_at" db:"created_at"`
	UpdatedAt                 time.Time `json:"updated_at" db:"updated_at"`
}

// ChequeClearingStatus represents cheque clearing status
type ChequeClearingStatus struct {
	ChequeID        string     `json:"cheque_id" db:"cheque_id"`
	PaymentID       string     `json:"payment_id" db:"payment_id"`
	RequestID       string     `json:"request_id" db:"request_id"`
	PolicyNumber    string     `json:"policy_number" db:"policy_number"`
	ChequeNumber    string     `json:"cheque_number" db:"cheque_number"`
	BankName        string     `json:"bank_name" db:"bank_name"`
	ChequeDate      time.Time  `json:"cheque_date" db:"cheque_date"`
	Amount          float64    `json:"amount" db:"amount"`
	ClearanceStatus string     `json:"clearance_status" db:"clearance_status"`
	NextDueDate     *time.Time `json:"next_due_date" db:"next_due_date"`
	CreatedAt       time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at" db:"updated_at"`
}

// SuspenseAccount represents suspense account
type SuspenseAccount struct {
	SuspenseID           string     `json:"suspense_id" db:"suspense_id"`
	PolicyNumber         string     `json:"policy_number" db:"policy_number"`
	RequestID            *string    `json:"request_id" db:"request_id"`
	SuspenseType         string     `json:"suspense_type" db:"suspense_type"`
	SuspenseAccountType  string     `json:"suspense_account_type" db:"suspense_account_type"`
	Amount               float64    `json:"amount" db:"amount"`
	IsReversed           bool       `json:"is_reversed" db:"is_reversed"`
	ReversalDate         *time.Time `json:"reversal_date" db:"reversal_date"`
	ReversalAuthorizedBy *string    `json:"reversal_authorized_by" db:"reversal_authorized_by"`
	ReversalReason       *string    `json:"reversal_reason" db:"reversal_reason"`
	CreatedAt            time.Time  `json:"created_at" db:"created_at"`
	CreatedBy            *string    `json:"created_by" db:"created_by"`
	UpdatedAt            time.Time  `json:"updated_at" db:"updated_at"`
	Reason               string     `json:"reason" db:"reason"`
}

// PolicyValidationResult represents batched validation result for revival eligibility
// Combines policy data, config, and ongoing revival check in single DB query
type PolicyValidationResult struct {
	Policy              Policy `json:"policy"`
	MaxRevivalsAllowed  int    `json:"max_revivals_allowed"`
	OngoingRevivalCount int    `json:"ongoing_revival_count"`
}

// RevivalTermination represents the termination record when revival workflow terminates
type RevivalTermination struct {
	TerminationID     string    `json:"termination_id" db:"termination_id"`
	RequestID         string    `json:"request_id" db:"request_id"`
	TicketID          string    `json:"ticket_id" db:"ticket_id"`
	PolicyNumber      string    `json:"policy_number" db:"policy_number"`
	TerminationReason string    `json:"termination_reason" db:"termination_reason"`
	TerminationType   string    `json:"termination_type" db:"termination_type"`     // DEFAULT, MANUAL, SLA_EXPIRED
	InstallmentNumber int       `json:"installment_number" db:"installment_number"` // Which installment caused termination
	SuspenseCreated   bool      `json:"suspense_created" db:"suspense_created"`
	SuspenseAmount    float64   `json:"suspense_amount" db:"suspense_amount"`
	TerminatedAt      time.Time `json:"terminated_at" db:"terminated_at"`
	TerminatedBy      *string   `json:"terminated_by" db:"terminated_by"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
}
