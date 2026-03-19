package domain

import (
	"time"

	"github.com/google/uuid"
	"github.com/volatiletech/null/v9"
)

// SurrenderRequestType represents the type of surrender request
// Business Rule: BR-SUR-001, BR-FS-001
type SurrenderRequestType string

const (
	SurrenderRequestTypeVoluntary SurrenderRequestType = "VOLUNTARY"
	SurrenderRequestTypeForced    SurrenderRequestType = "FORCED"
)

// SurrenderStatus represents the current status of a surrender request
// Workflow: WF-SUR-001, WF-FS-001
type SurrenderStatus string

const (
	SurrenderStatusPendingDocumentUpload SurrenderStatus = "PENDING_DOCUMENT_UPLOAD"
	SurrenderStatusPendingVerification   SurrenderStatus = "PENDING_VERIFICATION"
	SurrenderStatusPendingApproval       SurrenderStatus = "PENDING_APPROVAL"
	SurrenderStatusApproved              SurrenderStatus = "APPROVED"
	SurrenderStatusRejected              SurrenderStatus = "REJECTED"
	SurrenderStatusPendingAutoCompletion SurrenderStatus = "PENDING_AUTO_COMPLETION"
	SurrenderStatusAutoCompleted         SurrenderStatus = "AUTO_COMPLETED"
	SurrenderStatusTerminated            SurrenderStatus = "TERMINATED"
)

// PolicyStatusSurrender represents policy status after surrender processing
// Business Rule: BR-SUR-011, BR-FS-008
type PolicyStatusSurrender string

const (
	PolicyStatusAP  PolicyStatusSurrender = "AP"  // Active Premium
	PolicyStatusIL  PolicyStatusSurrender = "IL"  // In Lapse
	PolicyStatusAL  PolicyStatusSurrender = "AL"  // Automatic Lapse
	PolicyStatusPWS PolicyStatusSurrender = "PWS" // Pending With Staff
	PolicyStatusPAS PolicyStatusSurrender = "PAS" // Pending Auto Surrender
	PolicyStatusTAS PolicyStatusSurrender = "TAS" // Terminated Auto Surrender
	PolicyStatusTS  PolicyStatusSurrender = "TS"  // Terminated Surrender
	PolicyStatusAU  PolicyStatusSurrender = "AU"  // Automatic (Reduced Paid-Up)
)

// PreviousPolicyStatus stores policy status before surrender processing for reversion
// Business Rule: BR-FS-018 (CRITICAL)
type PreviousPolicyStatus string

const (
	PreviousPolicyStatusAP PreviousPolicyStatus = "AP"
	PreviousPolicyStatusIL PreviousPolicyStatus = "IL"
	PreviousPolicyStatusAL PreviousPolicyStatus = "AL"
)

// DisbursementMethod represents the method of payment
// Business Rule: BR-SUR-009
type DisbursementMethod string

const (
	DisbursementMethodCash   DisbursementMethod = "CASH"
	DisbursementMethodCheque DisbursementMethod = "CHEQUE"
)

// RequestOwner represents who initiated the surrender request
// Business Rule: BR-FS-001
type RequestOwner string

const (
	RequestOwnerCustomer   RequestOwner = "CUSTOMER"
	RequestOwnerSystem     RequestOwner = "SYSTEM"
	RequestOwnerPostmaster RequestOwner = "POSTMASTER"
	RequestOwnerCPC        RequestOwner = "CPC" // Central Processing Center for forced surrender
)

// PolicySurrenderRequest represents a policy surrender request (voluntary or forced)
// Table: policy_surrender_requests
// Business Rules: BR-SUR-001 to BR-SUR-018, BR-FS-001 to BR-FS-018
type PolicySurrenderRequest struct {
	ID                           uuid.UUID              `json:"id" db:"id"`
	PolicyID                     string                 `json:"policy_id" db:"policy_id"`
	RequestNumber                string                 `json:"request_number" db:"request_number"`
	RequestType                  SurrenderRequestType   `json:"request_type" db:"request_type"`
	PreviousPolicyStatus         *PreviousPolicyStatus  `json:"previous_policy_status" db:"previous_policy_status"` // BR-FS-018
	RequestDate                  time.Time              `json:"request_date" db:"request_date"`
	SurrenderValueCalculatedDate time.Time              `json:"surrender_value_calculated_date" db:"surrender_value_calculated_date"`
	GrossSurrenderValue          float64                `json:"gross_surrender_value" db:"gross_surrender_value"`
	NetSurrenderValue            float64                `json:"net_surrender_value" db:"net_surrender_value"`
	PaidUpValue                  float64                `json:"paid_up_value" db:"paid_up_value"`
	BonusAmount                  *float64               `json:"bonus_amount" db:"bonus_amount"`
	SurrenderFactor              float64                `json:"surrender_factor" db:"surrender_factor"`
	UnpaidPremiumsDeduction      float64                `json:"unpaid_premiums_deduction" db:"unpaid_premiums_deduction"`
	LoanDeduction                float64                `json:"loan_deduction" db:"loan_deduction"`
	OtherDeductions              *float64               `json:"other_deductions" db:"other_deductions"`
	DisbursementMethod           DisbursementMethod     `json:"disbursement_method" db:"disbursement_method"`
	DisbursementAmount           float64                `json:"disbursement_amount" db:"disbursement_amount"`
	Reason                       *string                `json:"reason" db:"reason"`
	Status                       SurrenderStatus        `json:"status" db:"status"`
	Owner                        RequestOwner           `json:"owner" db:"owner"`
	CreatedAt                    time.Time              `json:"created_at" db:"created_at"`
	UpdatedAt                    time.Time              `json:"updated_at" db:"updated_at"`
	CreatedBy                    uuid.UUID              `json:"created_by" db:"created_by"`
	ApprovedBy                   *uuid.UUID             `json:"approved_by" db:"approved_by"`
	ApprovedAt                   *time.Time             `json:"approved_at" db:"approved_at"`
	ApprovalComments             *string                `json:"approval_comments" db:"approval_comments"`
	DeletedAt                    *time.Time             `json:"deleted_at" db:"deleted_at"`
	Version                      int                    `json:"version" db:"version"`
	Metadata                     map[string]interface{} `json:"metadata" db:"metadata"`
	SearchVector                 *string                `json:"-" db:"search_vector"` // PostgreSQL tsvector for full-text search
}

type PolicyDetailsOutput struct {
	Policy_number            string       `json:"Policy_number" select:"policy_number"`
	Customer_id              string       `json:"customer_id" select:"customer_id"`
	Customer_name            string       `json:"customer_name" select:"customer_name"`
	Product_code             string       `json:"product_code" select:"product_code"`
	Product_name             string       `json:"product_name" select:"product_name"`
	Policy_status            string       `json:"policy_status" select:"policy_status"`
	Premium_frequency        null.String  `json:"premium_frequency" select:"premium_frequency"`
	Premium_amount           null.Float64 `json:"premium_amount" select:"premium_amount"`
	Sum_assured              int          `json:"sum_assured" select:"sum_assured"`
	Revival_count            null.Float64 `json:"revival_count" select:"revival_count"`
	Paid_to_date             time.Time    `json:"paid_to_date" select:"paid_to_date"`
	Maturity_date            time.Time    `json:"maturity_date" select:"maturity_date"`
	Date_of_commencement     time.Time    `json:"date_of_commencement" select:"date_of_commencement"`
	Last_revival_date        time.Time    `json:"last_revival_date" select:"last_revival_date"`
	Polissdate               time.Time    `json:"polissdate" select:"polissdate"`
	Outstandingloanprinciple null.Float64 `json:"outstandingloanprinciple" select:"outstandingloanprinciple"`
	Outstandingloaninterest  null.Float64 `json:"outstandingloaninterest" select:"outstandingloaninterest"`
	Totalbonus               null.Float64 `json:"totalbonus" select:"totalbonus"`
	Dob                      time.Time    `json:"dob" select:"dob"`
}

type SurrenderCalculationResponse struct {
	PolicyNumber             string               `json:"policy_number"`
	SumAssured               float64              `json:"sum_assured"`
	PaidToDate               time.Time            `json:"paid_to_date"`
	Polissdate               time.Time            `json:"polissdate"`
	MaturityDate             time.Time            `json:"maturity_date"`
	ProductCode              string               `json:"product_code"`
	Dob                      time.Time            `json:"dob"`
	Cumulative               float64              `json:"cumulative"`
	UnpaidPremiums           float64              `json:"unpaid_premiums"`
	DefOnUnpPrem             float64              `json:"def_on_unp_prem"`
	Outstandingloanprinciple float64              `json:"outstandingloanprinciple"`
	Outstandingloaninterest  float64              `json:"outstandingloaninterest"`
	SFValue                  float64              `json:"sf_value"`
	PaidUpValue              float64              `json:"paidupvalue"`
	OtherCharges             float64              `json:"other_charges"`
	SurrenderValue           float64              `json:"surrender_value"`
	BonusDetails             []AccruedBonusOutput `json:"bonus_details"`
}
type SubmitDERequest struct {
	Surrender_request_id      string `json:"surrender_request_id" validate:"required"`
	Surrender_request_channel string `json:"surrender_request_channel"`
	Request_name              string `json:"request_name"`
	Current_stage_name        string `json:"current_stage_name"`
	Created_by                int    `json:"created_by"`
	Modified_by               int    `json:"modified_by"`
	Remarks                   string `json:"remarks"`
	Paymentmode               string `json:"paymentmode"`
	Bankname                  string `json:"bankname"`
	Micrcode                  string `json:"micrcode"`
	Accounttype               string `json:"accounttype"`
	Ifsccode                  string `json:"ifsccode"`
	Accountnumber             string `json:"accountnumber"`
	Accountholdername         string `json:"accountholdername"`
	Branchname                string `json:"branchname"`
	Banktype                  string `json:"banktype"`
	Ismicrvalidated           bool   `json:"ismicrvalidated" select:"ismicrvalidated"`
	Policybond                bool   `json:"Policybond" select:"policybond"`
	Lrrb                      bool   `json:"Lrrb" select:"lrrb"`
	Prb                       bool   `json:"Prb" select:"prb"`
	Pdo_certificate           bool   `json:"Pdo_certificate" select:"pdo_certificate"`
	Application               bool   `json:"Application" select:"application"`
	Idproof_insurant          bool   `json:"Idproof_insurant" select:"idproof_insurant"`
	Addressproof_insurant     bool   `json:"Addressproof_insurant" select:"addressproof_insurant"`
	Idproof_messenger         bool   `json:"Idproof_messenger" select:"idproof_messenger"`
	Addressproof_messenger    bool   `json:"Addressproof_messenger" select:"addressproof_messenger"`
	Account_details_proof     bool   `json:"Account_details_proof" select:"account_details_proof"`
	Others                    bool   `json:"Others" select:"others"`
	Cpc_office_id             int    `json:"cpc_office_id"`
	PolicyNumber              string `json:"policy_number"`
}

type SubmitQCRequest struct {
	Surrender_request_id      string `json:"surrender_request_id" validate:"required"`
	Surrender_request_channel string `json:"surrender_request_channel"`
	Request_name              string `json:"request_name"`
	Current_stage_name        string `json:"current_stage_name"`
	Created_by                int    `json:"created_by"`
	Modified_by               int    `json:"modified_by"`
	Remarks                   string `json:"remarks"`
	Paymentmode               string `json:"paymentmode"`
	Bankname                  string `json:"bankname"`
	Micrcode                  string `json:"micrcode"`
	Accounttype               string `json:"accounttype"`
	Ifsccode                  string `json:"ifsccode"`
	Accountnumber             string `json:"accountnumber"`
	Accountholdername         string `json:"accountholdername"`
	Branchname                string `json:"branchname"`
	Banktype                  string `json:"banktype"`
	Ismicrvalidated           bool   `json:"ismicrvalidated" select:"ismicrvalidated"`
	Policybond                bool   `json:"Policybond" select:"policybond"`
	Lrrb                      bool   `json:"Lrrb" select:"lrrb"`
	Prb                       bool   `json:"Prb" select:"prb"`
	Pdo_certificate           bool   `json:"Pdo_certificate" select:"pdo_certificate"`
	Application               bool   `json:"Application" select:"application"`
	Idproof_insurant          bool   `json:"Idproof_insurant" select:"idproof_insurant"`
	Addressproof_insurant     bool   `json:"Addressproof_insurant" select:"addressproof_insurant"`
	Idproof_messenger         bool   `json:"Idproof_messenger" select:"idproof_messenger"`
	Addressproof_messenger    bool   `json:"Addressproof_messenger" select:"addressproof_messenger"`
	Account_details_proof     bool   `json:"Account_details_proof" select:"account_details_proof"`
	Others                    bool   `json:"Others" select:"others"`
	Cpc_office_id             int    `json:"cpc_office_id"`
	PolicyNumber              string `json:"policy_number"`
}

type SubmitApprovalRequest struct {
	Surrender_request_id      string `json:"surrender_request_id" validate:"required"`
	Surrender_request_channel string `json:"surrender_request_channel"`
	Request_name              string `json:"request_name"`
	Current_stage_name        string `json:"current_stage_name"`
	Created_by                int    `json:"created_by"`
	Modified_by               int    `json:"modified_by"`
	Remarks                   string `json:"remarks"`
	Paymentmode               string `json:"paymentmode"`
	Bankname                  string `json:"bankname"`
	Micrcode                  string `json:"micrcode"`
	Accounttype               string `json:"accounttype"`
	Ifsccode                  string `json:"ifsccode"`
	Accountnumber             string `json:"accountnumber"`
	Accountholdername         string `json:"accountholdername"`
	Branchname                string `json:"branchname"`
	Banktype                  string `json:"banktype"`
	Ismicrvalidated           bool   `json:"ismicrvalidated" select:"ismicrvalidated"`
	Policybond                bool   `json:"Policybond" select:"policybond"`
	Lrrb                      bool   `json:"Lrrb" select:"lrrb"`
	Prb                       bool   `json:"Prb" select:"prb"`
	Pdo_certificate           bool   `json:"Pdo_certificate" select:"pdo_certificate"`
	Application               bool   `json:"Application" select:"application"`
	Idproof_insurant          bool   `json:"Idproof_insurant" select:"idproof_insurant"`
	Addressproof_insurant     bool   `json:"Addressproof_insurant" select:"addressproof_insurant"`
	Idproof_messenger         bool   `json:"Idproof_messenger" select:"idproof_messenger"`
	Addressproof_messenger    bool   `json:"Addressproof_messenger" select:"addressproof_messenger"`
	Account_details_proof     bool   `json:"Account_details_proof" select:"account_details_proof"`
	Others                    bool   `json:"Others" select:"others"`
	Cpc_office_id             int    `json:"cpc_office_id"`
	PolicyNumber              string `json:"policy_number"`
}

type SurrenderFactorOutput struct {
	Product_code     string  `json:"product_code"`
	Age_at_entry     int     `json:"age_at_entry"`
	Age_at_maturity  int     `json:"age_at_maturity"`
	Type_of_policy   string  `json:"type_of_policy"`
	Surrender_factor float64 `json:"surrender_factor"`
	Status           string  `json:"status"`
}

type BonusOutput struct {
	Bonus_rate_id          int    `json:"bonus_rate_id"`
	Product_code           string `json:"product_code"`
	Bonus_declaration_year int    `json:"bonus_declaration_year"`
	Bonus_year             int    `json:"bonus_year"`
	Bonus_rate_per_1000_sa int    `json:"bonus_rate_per_1000_sa"`
	Bonus_from             string `json:"bonus_from" select:"bonus_from"`
	Bonus_to               string `json:"bonus_to" select:"bonus_to"`
	Bonus_type             string `json:"bonus_type"`
	Status                 string `json:"status"`
}

type AccruedBonusOutput struct {
	// SrNo       int       `json:"sr_no"`
	BonusRate  float64 `json:"bonus_rate"`
	FromDate   string  `json:"from_date"`
	ToDate     string  `json:"to_date"`
	BonusValue float64 `json:"bonus_value"`
	Cumulative float64 `json:"cumulative"`
}

type SRDetailsOutput struct {
	Policy_number          null.String  `json:"Policy_number" select:"policy_number"`
	Surrender_request_id   null.String  `json:"Surrender_request_id" select:"surrender_request_id"`
	Paidupvalue            null.Float64 `json:"Paidupvalue" select:"paidupvalue"`
	Bonus                  null.Float64 `json:"Bonus" select:"bonus"`
	Grossamount            null.Float64 `json:"Grossamount" select:"grossamount"`
	Loanprincipal          null.Float64 `json:"Loanprincipal" select:"loanprincipal"`
	Loaninterest           null.Float64 `json:"Loaninterest" select:"loaninterest"`
	Surrenderfactor        null.Float64 `json:"surrenderfactor" select:"surrenderfactor"`
	Othercharges           null.Float64 `json:"othercharges" select:"othercharges"`
	Surrendervalue         null.Float64 `json:"surrendervalue" select:"surrendervalue"`
	Bonusrate              null.Float64 `json:"Bonusrate" select:"bonusrate"`
	Bonusamount            null.Float64 `json:"Bonusamount" select:"bonusamount"`
	Paymentmode            null.String  `json:"Paymentmode" select:"paymentmode"`
	Bankname               null.String  `json:"Bankname" select:"bankname"`
	Micrcode               null.String  `json:"Micrcode" select:"micrcode"`
	Accounttype            null.String  `json:"Accounttype" select:"accounttype"`
	Ifsccode               null.String  `json:"Ifsccode" select:"ifsccode"`
	Accountnumber          null.String  `json:"Accountnumber" select:"accountnumber"`
	Accountholdername      null.String  `json:"Accountholdername" select:"accountholdername"`
	Branchname             null.String  `json:"Branchname" select:"branchname"`
	Banktype               null.String  `json:"Banktype" select:"banktype"`
	Ismicrvalidated        bool         `json:"Ismicrvalidated" select:"ismicrvalidated"`
	Policybond             bool         `json:"Policybond" select:"policybond"`
	Lrrb                   bool         `json:"Lrrb" select:"lrrb"`
	Prb                    bool         `json:"Prb" select:"prb"`
	Pdo_certificate        bool         `json:"Pdo_certificate" select:"pdo_certificate"`
	Application            bool         `json:"Application" select:"application"`
	Idproof_insurant       bool         `json:"Idproof_insurant" select:"idproof_insurant"`
	Addressproof_insurant  bool         `json:"Addressproof_insurant" select:"addressproof_insurant"`
	Idproof_messenger      bool         `json:"Idproof_messenger" select:"idproof_messenger"`
	Addressproof_messenger bool         `json:"Addressproof_messenger" select:"addressproof_messenger"`
	Account_details_proof  bool         `json:"Account_details_proof" select:"account_details_proof"`
	Reason                 null.String  `json:"reason" select:"reason"`
	Remarks                null.String  `json:"remarks" select:"remarks"`
	Sumassured             null.Float64 `json:"sumassured" select:"sumassured"`
	Paid_to_date           time.Time    `json:"paid_to_date"`
	Polissdate             time.Time    `json:"polissdate"`
	Maturitydate           time.Time    `json:"maturitydate"`
	Productcode            string       `json:"productcode" select:"productcode"`
	Dob                    time.Time    `json:"dob"`
	Unpaidprem             null.Float64 `json:"unpaidprem" select:"unpaidprem"`
	Def                    null.Float64 `json:"def" select:"def"`
	Others                 bool         `json:"Others" select:"others"`
}

type PolicyDetails struct {
	Policy_number        null.String  `json:"Policy_number" select:"policy_number"`
	Surrender_request_id null.String  `json:"Surrender_request_id" select:"surrender_request_id"`
	Productcode          string       `json:"productcode" select:"productcode"`
	Dob                  time.Time    `json:"dob"`
	Sumassured           null.Float64 `json:"sumassured" select:"sumassured"`
	Paid_to_date         time.Time    `json:"paid_to_date"`
	Polissdate           time.Time    `json:"polissdate"`
	Maturitydate         time.Time    `json:"maturitydate"`
}

type SurrenderCalculation struct {
	Paidupvalue     null.Float64 `json:"Paidupvalue" select:"paidupvalue"`
	Bonus           null.Float64 `json:"Bonus" select:"bonus"`
	Grossamount     null.Float64 `json:"Grossamount" select:"grossamount"`
	Loanprincipal   null.Float64 `json:"Loanprincipal" select:"loanprincipal"`
	Loaninterest    null.Float64 `json:"Loaninterest" select:"loaninterest"`
	Surrenderfactor null.Float64 `json:"surrenderfactor" select:"surrenderfactor"`
	Othercharges    null.Float64 `json:"othercharges" select:"othercharges"`
	Surrendervalue  null.Float64 `json:"surrendervalue" select:"surrendervalue"`
	Bonusrate       null.Float64 `json:"Bonusrate" select:"bonusrate"`
	Bonusamount     null.Float64 `json:"Bonusamount" select:"bonusamount"`
	Unpaidprem      null.Float64 `json:"unpaidprem" select:"unpaidprem"`
	Def             null.Float64 `json:"def" select:"def"`
}

type BankDetails struct {
	Paymentmode       null.String `json:"Paymentmode" select:"paymentmode"`
	Bankname          null.String `json:"Bankname" select:"bankname"`
	Micrcode          null.String `json:"Micrcode" select:"micrcode"`
	Accounttype       null.String `json:"Accounttype" select:"accounttype"`
	Ifsccode          null.String `json:"Ifsccode" select:"ifsccode"`
	Accountnumber     null.String `json:"Accountnumber" select:"accountnumber"`
	Accountholdername null.String `json:"Accountholdername" select:"accountholdername"`
	Branchname        null.String `json:"Branchname" select:"branchname"`
	Banktype          null.String `json:"Banktype" select:"banktype"`
	Ismicrvalidated   bool        `json:"Ismicrvalidated" select:"ismicrvalidated"`
}

type Documents struct {
	Policybond             bool `json:"Policybond" select:"policybond"`
	Lrrb                   bool `json:"Lrrb" select:"lrrb"`
	Prb                    bool `json:"Prb" select:"prb"`
	Pdo_certificate        bool `json:"Pdo_certificate" select:"pdo_certificate"`
	Application            bool `json:"Application" select:"application"`
	Idproof_insurant       bool `json:"Idproof_insurant" select:"idproof_insurant"`
	Addressproof_insurant  bool `json:"Addressproof_insurant" select:"addressproof_insurant"`
	Idproof_messenger      bool `json:"Idproof_messenger" select:"idproof_messenger"`
	Addressproof_messenger bool `json:"Addressproof_messenger" select:"addressproof_messenger"`
	Account_details_proof  bool `json:"Account_details_proof" select:"account_details_proof"`
	Others                 bool `json:"Others" select:"others"`
}

type AdditionalInfo struct {
	Reason  null.String `json:"reason" select:"reason"`
	Remarks null.String `json:"remarks" select:"remarks"`
}

type SRDetailsOutput1 struct {
	PolicyDetails        PolicyDetails        `json:"policyDetails"`
	SurrenderCalculation SurrenderCalculation `json:"surrenderCalculation"`
	BankDetails          BankDetails          `json:"bankDetails"`
	Documents            Documents            `json:"documents"`
	AdditionalInfo       AdditionalInfo       `json:"additionalInfo"`
}

type IndexSurrenderRequestInput struct {
	PolicyNumber              string  `json:"policy_number" validate:"required"`
	Surrender_request_channel string  `json:"surrender_request_channel"`
	Indexing_office_id        int     `json:"indexing_office_id"`
	Cpc_office_id             int     `json:"cpc_office_id"`
	Created_by                int     `json:"created_by"`
	Modified_by               int     `json:"modified_by"`
	Remarks                   string  `json:"remarks"`
	// PM integration fields — populated when the workflow is dispatched by Policy Management
	TemporalWorkflowID  string `json:"temporal_workflow_id"`
	PMServiceRequestID  int64  `json:"pm_service_request_id"`
	PMPolicyDBID        int64  `json:"pm_policy_db_id"`
	Paidupvalue               float64 `json:"paidupvalue"`
	Bonus                     float64 `json:"bonusvalue"`
	Grossamount               float64 `json:"grossamount"`
	Loanprincipal             float64 `json:"loanprincipal"`
	Loaninterest              float64 `json:"loaninterest"`
	Surrenderfactor           float64 `json:"surrenderfactor"`
	Othercharges              float64 `json:"othercharges"`
	Surrendervalue            float64 `json:"surrendervalue"`
	Bonusrate                 float64 `json:"bonusrate"`
	Bonusamount               float64 `json:"bonusamount"`
	Sumassured                float64 `json:"sumassured"`
	Paid_to_date              string  `json:"paid_to_date"`
	Polissdate                string  `json:"polissdate"`
	Maturitydate              string  `json:"maturitydate"`
	Productcode               string  `json:"productcode"`
	Dob                       string  `json:"dob"`
	Unpaidprem                float64 `json:"unpaidprem"`
	Def                       float64 `json:"def"`
	Stage_name                string  `json:"stage_name"`
}

type IndexSurrenderRequestInput2 struct {
	PolicyNumber              string  `json:"policy_number" validate:"required"`
	Surrender_request_channel string  `json:"surrender_request_channel"`
	Indexing_office_id        int     `json:"indexing_office_id"`
	Cpc_office_id             int     `json:"cpc_office_id"`
	Created_by                int     `json:"created_by"`
	Modified_by               int     `json:"modified_by"`
	Remarks                   string  `json:"remarks"`
	// PM integration fields — populated when the workflow is dispatched by Policy Management
	TemporalWorkflowID  string `json:"temporal_workflow_id"`
	PMServiceRequestID  int64  `json:"pm_service_request_id"`
	PMPolicyDBID        int64  `json:"pm_policy_db_id"`
	Paidupvalue               float64 `json:"paidupvalue"`
	Bonus                     float64 `json:"bonusvalue"`
	Grossamount               float64 `json:"grossamount"`
	Loanprincipal             float64 `json:"loanprincipal"`
	Loaninterest              float64 `json:"loaninterest"`
	Surrenderfactor           float64 `json:"surrenderfactor"`
	Othercharges              float64 `json:"othercharges"`
	Surrendervalue            float64 `json:"surrendervalue"`
	Bonusrate                 float64 `json:"bonusrate"`
	Bonusamount               float64 `json:"bonusamount"`
	Sumassured                float64 `json:"sumassured"`
	Paid_to_date              *time.Time  `json:"paid_to_date"`
	Polissdate                *time.Time  `json:"polissdate"`
	Maturitydate              *time.Time  `json:"maturitydate"`
	Productcode               string  `json:"productcode"`
	Dob                       *time.Time  `json:"dob"`
	Unpaidprem                float64 `json:"unpaidprem"`
	Def                       float64 `json:"def"`
	Stage_name                string  `json:"stage_name"`
}
type SRStagingDetailsOutput struct {
	Surrender_request_id      string    `json:"Surrender_request_id" select:"surrender_request_id"`
	Surrender_request_channel string    `json:"Surrender_request_channel" select:"surrender_request_channel"`
	Request_name              string    `json:"Request_name" select:"request_name"`
	Policy_number             string    `json:"Policy_number" select:"policy_number"`
	Current_stage_name        string    `json:"Current_stage_name" select:"current_stage_name"`
	Created_by                int       `json:"Created_by" select:"created_by"`
	Created_date              time.Time `json:"Created_date" select:"created_date"`
	Cpc_office_id             int       `json:"Cpc_office_id" select:"cpc_office_id"`
	Remarks                   string    `json:"Remarks" select:"remarks"`
}

type PendingRequestOutput struct {
	Surrender_request_id      null.String `json:"Surrender_request_id" select:"surrender_request_id"`
	Surrender_request_channel null.String `json:"Surrender_request_channel" select:"surrender_request_channel"`
	Request_name              null.String `json:"Request_name" select:"request_name"`
	Policy_number             null.String `json:"Policy_number" select:"policy_number"`
	Stage_name                null.String `json:"Stage_name" select:"stage_name"`
	Indexing_office_id        null.Int    `json:"Indexing_office_id" select:"indexing_office_id"`
	Cpc_office_id             null.Int    `json:"Cpc_office_id" select:"cpc_office_id"`
	Created_by                null.Int    `json:"Created_by" select:"created_by"`
	Created_date              time.Time   `json:"Created_date" select:"created_date"`
	Remarks                   null.String `json:"Remarks" select:"remarks"`
}

// SurrenderBonusDetail represents bonus details for a surrender request
// Table: surrender_bonus_details
// Business Rule: BR-SUR-005
type SurrenderBonusDetail struct {
	ID                 uuid.UUID `json:"id" db:"id"`
	SurrenderRequestID uuid.UUID `json:"surrender_request_id" db:"surrender_request_id"`
	FinancialYear      string    `json:"financial_year" db:"financial_year"`
	SumAssured         float64   `json:"sum_assured" db:"sum_assured"`
	BonusRate          float64   `json:"bonus_rate" db:"bonus_rate"`
	BonusAmount        float64   `json:"bonus_amount" db:"bonus_amount"`
	CreatedAt          time.Time `json:"created_at" db:"created_at"`
}
