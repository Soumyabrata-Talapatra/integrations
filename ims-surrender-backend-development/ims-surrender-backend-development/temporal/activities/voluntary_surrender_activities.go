package activities

import (
	"context"
	"fmt"
	"time"

	"gitlab.cept.gov.in/it-2.0-policy/surrender-service/core/domain"
	repo "gitlab.cept.gov.in/it-2.0-policy/surrender-service/repo/postgres"
)

// Activities for Voluntary Surrender Workflow (TEMP-001)

// voluntarySurrenderActivities holds dependencies for voluntary surrender activities
type voluntarySurrenderActivities struct {
	surrenderRepo *repo.SurrenderRequestRepository
}

// activitiesInstance is the shared instance of activities
var activitiesInstance *voluntarySurrenderActivities

// InitVoluntarySurrenderActivities initializes the activities with repository
func InitVoluntarySurrenderActivities(surrenderRepo *repo.SurrenderRequestRepository) {
	activitiesInstance = &voluntarySurrenderActivities{
		surrenderRepo: surrenderRepo,
	}
}

type ValidateEligibilityInput struct {
	PolicyID           string
	SurrenderRequestID string

	ProductCode  string
	MaturityDate time.Time
}

type ValidateEligibilityResult struct {
	Eligible bool
	Reasons  []string
}

// ValidateEligibilityActivity checks surrender-domain business rules:
//   - Product must not be in the ineligible list (AEA, AEA-10, GY)
//   - Policy must not have passed its maturity date
//
// Note: policy state gate (ACTIVE/VOID_LAPSE/etc.) is already checked by PM
// before dispatching the child workflow, so it is NOT rechecked here.
func ValidateEligibilityActivity(ctx context.Context, input ValidateEligibilityInput) (*ValidateEligibilityResult, error) {
	if activitiesInstance == nil {
		return nil, fmt.Errorf("activities not initialized")
	}

	// policy, err := activitiesInstance.surrenderRepo.FindByPolicyNumber(ctx, input.PolicyID)
	// if err != nil {
	// 	return nil, fmt.Errorf("failed to fetch policy details for %s: %w", input.PolicyID, err)
	// }

	var reasons []string

	ineligibleProducts := []string{"AEA", "AEA-10", "GY"}
	for _, prod := range ineligibleProducts {
		if input.ProductCode == prod {
			reasons = append(reasons, fmt.Sprintf("product '%s' is not eligible for surrender", input.ProductCode))
			break
		}
	}

	if input.MaturityDate.Before(time.Now()) {
		reasons = append(reasons, "policy has reached maturity; process through Maturity Claims")
	}

	return &ValidateEligibilityResult{
		Eligible: len(reasons) == 0,
		Reasons:  reasons,
	}, nil
}

type CalculateSurrenderValueInput struct {
	SurrenderRequestID string
	PolicyID           string
}

type CalculateSurrenderValueResult struct {
	GrossSurrenderValue  float64
	NetSurrenderValue    float64
	PredictedDisposition string
}

func CalculateSurrenderValueActivity(ctx context.Context, input CalculateSurrenderValueInput) (*CalculateSurrenderValueResult, error) {
	// Placeholder - would call calculation service
	return &CalculateSurrenderValueResult{
		GrossSurrenderValue:  50000,
		NetSurrenderValue:    45000,
		PredictedDisposition: "TS",
	}, nil
}

type VerifyDocumentsInput struct {
	SurrenderRequestID string
}

type VerifyDocumentsResult struct {
	AllVerified   bool
	VerifiedCount int
	RequiredCount int
}

func VerifyDocumentsActivity(ctx context.Context, input VerifyDocumentsInput) (*VerifyDocumentsResult, error) {
	// Placeholder - would call document service
	return &VerifyDocumentsResult{
		AllVerified:   true,
		VerifiedCount: 3,
		RequiredCount: 3,
	}, nil
}

type RouteToApprovalInput struct {
	SurrenderRequestID string
	Priority           string
}

type RouteToApprovalResult struct {
	TaskID   string
	Assigned bool
}

func RouteToApprovalActivity(ctx context.Context, input RouteToApprovalInput) (*RouteToApprovalResult, error) {
	// Placeholder - would call approval service
	return &RouteToApprovalResult{
		TaskID:   "task-123",
		Assigned: true,
	}, nil
}

type ProcessPaymentInput struct {
	SurrenderRequestID string
	Amount             float64
	DisbursementMethod string
}

type ProcessPaymentResult struct {
	PaymentReference string
	Success          bool
}

func ProcessPaymentActivity(ctx context.Context, input ProcessPaymentInput) (*ProcessPaymentResult, error) {
	// Placeholder - would call payment service
	return &ProcessPaymentResult{
		PaymentReference: "PAY-" + time.Now().Format("20060102150405"),
		Success:          true,
	}, nil
}

type UpdatePolicyStatusInput struct {
	PolicyID           string
	SurrenderRequestID string
	NewStatus          string
}

type UpdatePolicyStatusResult struct {
	NewStatus string
	Success   bool
}

func UpdatePolicyStatusActivity(ctx context.Context, input UpdatePolicyStatusInput) (*UpdatePolicyStatusResult, error) {
	// Placeholder - would call policy service
	return &UpdatePolicyStatusResult{
		NewStatus: input.NewStatus,
		Success:   true,
	}, nil
}

// New activities for handler operations

type IndexSurrenderInput struct {
	PolicyNumber            string
	SurrenderRequestChannel string
	TemporalWorkflowID      string
	PMServiceRequestID      int64
	PMPolicyDBID            int64
	Stage_name              string
}

type IndexSurrenderResult struct {
	ServiceRequestID string
	Success          bool
}

// IndexSurrenderActivity creates the surrender_request record in the DB and
// stores the Temporal workflow ID so that DE/QC/Approval handlers can later
// look it up and signal the correct workflow instance.
func IndexSurrenderActivity(ctx context.Context, input IndexSurrenderInput) (*IndexSurrenderResult, error) {
	if activitiesInstance == nil {
		return nil, fmt.Errorf("activities not initialized")
	}

	req := domain.IndexSurrenderRequestInput2{
		PolicyNumber:              input.PolicyNumber,
		Surrender_request_channel: input.SurrenderRequestChannel,
		Stage_name:                input.Stage_name,
		TemporalWorkflowID:        input.TemporalWorkflowID,
		PMServiceRequestID:        input.PMServiceRequestID,
		PMPolicyDBID:              input.PMPolicyDBID,
	}

	serviceRequestID, err := activitiesInstance.surrenderRepo.IndexSurrenderRequestRepo(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to index surrender request: %w", err)
	}

	return &IndexSurrenderResult{
		ServiceRequestID: serviceRequestID,
		Success:          true,
	}, nil
}

type SubmitDEInput struct {
	SurrenderRequestID      string
	SurrenderRequestChannel string
	RequestName             string
	CurrentStageName        string
	CreatedBy               int
	Modified_by             int
	Remarks                 string
	Paymentmode             string
	Bankname                string
	Micrcode                string
	Accounttype             string
	Ifsccode                string
	Accountnumber           string
	Accountholdername       string
	Branchname              string
	Banktype                string
	Ismicrvalidated         bool
	Policybond              bool
	Lrrb                    bool
	Prb                     bool
	Pdo_certificate         bool
	Application             bool
	Idproof_insurant        bool
	Addressproof_insurant   bool
	Idproof_messenger       bool
	Addressproof_messenger  bool
	Account_details_proof   bool
	Cpc_office_id           int
	PolicyNumber            string
	Others                  bool
}

type SubmitDEResult struct {
	Success bool
	Message string
}

func SubmitDEActivity(ctx context.Context, input SubmitDEInput) (*SubmitDEResult, error) {
	if activitiesInstance == nil {
		return &SubmitDEResult{
			Success: false,
			Message: "Activities not initialized",
		}, fmt.Errorf("activities not initialized")
	}

	// Convert input to domain.SubmitDERequest
	req := domain.SubmitDERequest{
		Surrender_request_id:      input.SurrenderRequestID,
		Surrender_request_channel: input.SurrenderRequestChannel,
		Request_name:              input.RequestName,
		Current_stage_name:        input.CurrentStageName,
		Created_by:                input.CreatedBy,
		Modified_by:               input.Modified_by, // Using CreatedBy as ModifiedBy
		Remarks:                   input.Remarks,
		Paymentmode:               input.Paymentmode,
		Bankname:                  input.Bankname,
		Micrcode:                  input.Micrcode,
		Accounttype:               input.Accounttype,
		Ifsccode:                  input.Ifsccode,
		Accountnumber:             input.Accountnumber,
		Accountholdername:         input.Accountholdername,
		Branchname:                input.Branchname,
		Banktype:                  input.Banktype,
		Ismicrvalidated:           input.Ismicrvalidated,
		Policybond:                input.Policybond,
		Lrrb:                      input.Lrrb,
		Prb:                       input.Prb,
		Pdo_certificate:           input.Pdo_certificate,
		Application:               input.Application,
		Idproof_insurant:          input.Idproof_insurant,
		Addressproof_insurant:     input.Addressproof_insurant,
		Idproof_messenger:         input.Idproof_messenger,
		Addressproof_messenger:    input.Addressproof_messenger,
		Account_details_proof:     input.Account_details_proof,
		Others:                    input.Others,
		Cpc_office_id:             input.Cpc_office_id,
		PolicyNumber:              input.PolicyNumber,
	}

	// Call the repository
	result, err := activitiesInstance.surrenderRepo.SubmitDERepo(ctx, req)
	if err != nil {
		return &SubmitDEResult{
			Success: false,
			Message: "Failed to submit DE: " + err.Error(),
		}, err
	}

	return &SubmitDEResult{
		Success: true,
		Message: result,
	}, nil
}

type SubmitQCInput struct {
	SurrenderRequestID      string
	SurrenderRequestChannel string
	RequestName             string
	CurrentStageName        string
	CreatedBy               int
	Modified_by             int
	Remarks                 string
	Paymentmode             string
	Bankname                string
	Micrcode                string
	Accounttype             string
	Ifsccode                string
	Accountnumber           string
	Accountholdername       string
	Branchname              string
	Banktype                string
	Ismicrvalidated         bool
	Policybond              bool
	Lrrb                    bool
	Prb                     bool
	Pdo_certificate         bool
	Application             bool
	Idproof_insurant        bool
	Addressproof_insurant   bool
	Idproof_messenger       bool
	Addressproof_messenger  bool
	Account_details_proof   bool
	Cpc_office_id           int
	PolicyNumber            string
	Others                  bool
}

type SubmitQCResult struct {
	Success bool
	Message string
}

func SubmitQCActivity(ctx context.Context, input SubmitDEInput) (*SubmitQCResult, error) {
	if activitiesInstance == nil {
		return &SubmitQCResult{
			Success: false,
			Message: "Activities not initialized",
		}, fmt.Errorf("activities not initialized")
	}

	// Convert input to domain.SubmitDERequest
	// req := domain.SubmitDERequest{
	// 	Surrender_request_id:      input.SurrenderRequestID,
	// 	Surrender_request_channel: input.SurrenderRequestChannel,
	// 	Request_name:              input.RequestName,
	// 	Current_stage_name:        input.CurrentStageName,
	// 	Created_by:                input.CreatedBy,
	// 	Modified_by:               input.CreatedBy, // Using CreatedBy as ModifiedBy
	// 	Remarks:                   " ",
	// }

	req := domain.SubmitQCRequest{
		Surrender_request_id:      input.SurrenderRequestID,
		Surrender_request_channel: input.SurrenderRequestChannel,
		Request_name:              input.RequestName,
		Current_stage_name:        input.CurrentStageName,
		Created_by:                input.CreatedBy,
		Modified_by:               input.Modified_by, // Using CreatedBy as ModifiedBy
		Remarks:                   input.Remarks,
		Paymentmode:               input.Paymentmode,
		Bankname:                  input.Bankname,
		Micrcode:                  input.Micrcode,
		Accounttype:               input.Accounttype,
		Ifsccode:                  input.Ifsccode,
		Accountnumber:             input.Accountnumber,
		Accountholdername:         input.Accountholdername,
		Branchname:                input.Branchname,
		Banktype:                  input.Banktype,
		Ismicrvalidated:           input.Ismicrvalidated,
		Policybond:                input.Policybond,
		Lrrb:                      input.Lrrb,
		Prb:                       input.Prb,
		Pdo_certificate:           input.Pdo_certificate,
		Application:               input.Application,
		Idproof_insurant:          input.Idproof_insurant,
		Addressproof_insurant:     input.Addressproof_insurant,
		Idproof_messenger:         input.Idproof_messenger,
		Addressproof_messenger:    input.Addressproof_messenger,
		Account_details_proof:     input.Account_details_proof,
		Others:                    input.Others,
		Cpc_office_id:             input.Cpc_office_id,
		PolicyNumber:              input.PolicyNumber,
	}

	// Call the repository
	result, err := activitiesInstance.surrenderRepo.SubmitQCRepo(ctx, req)
	if err != nil {
		return &SubmitQCResult{
			Success: false,
			Message: "Failed to submit QC: " + err.Error(),
		}, err
	}

	return &SubmitQCResult{
		Success: true,
		Message: result,
	}, nil
}

// type SubmitApprovalInput struct {
// 	SurrenderRequestID      string
// 	SurrenderRequestChannel string
// 	RequestName             string
// 	CurrentStageName        string
// 	CreatedBy               int
// }

type SubmitApprovalInput struct {
	SurrenderRequestID      string
	SurrenderRequestChannel string
	RequestName             string
	CurrentStageName        string
	CreatedBy               int
	Modified_by             int
	Remarks                 string
	Paymentmode             string
	Bankname                string
	Micrcode                string
	Accounttype             string
	Ifsccode                string
	Accountnumber           string
	Accountholdername       string
	Branchname              string
	Banktype                string
	Ismicrvalidated         bool
	Policybond              bool
	Lrrb                    bool
	Prb                     bool
	Pdo_certificate         bool
	Application             bool
	Idproof_insurant        bool
	Addressproof_insurant   bool
	Idproof_messenger       bool
	Addressproof_messenger  bool
	Account_details_proof   bool
	Cpc_office_id           int
	PolicyNumber            string
	Others                  bool
}

type SubmitApprovalResult struct {
	Success bool
	Message string
	Status  string
}

func SubmitApprovalActivity(ctx context.Context, input SubmitDEInput) (*SubmitApprovalResult, error) {
	if activitiesInstance == nil {
		return &SubmitApprovalResult{
			Success: false,
			Message: "Activities not initialized",
			Status:  "ERROR",
		}, fmt.Errorf("activities not initialized")
	}

	// Convert input to domain.SubmitDERequest
	// req := domain.SubmitDERequest{
	// 	Surrender_request_id:      input.SurrenderRequestID,
	// 	Surrender_request_channel: input.SurrenderRequestChannel,
	// 	Request_name:              input.RequestName,
	// 	Current_stage_name:        input.CurrentStageName,
	// 	Created_by:                input.CreatedBy,
	// 	Modified_by:               input.CreatedBy, // Using CreatedBy as ModifiedBy
	// 	Remarks:                   " ",
	// }

	req := domain.SubmitApprovalRequest{
		Surrender_request_id:      input.SurrenderRequestID,
		Surrender_request_channel: input.SurrenderRequestChannel,
		Request_name:              input.RequestName,
		Current_stage_name:        input.CurrentStageName,
		Created_by:                input.CreatedBy,
		Modified_by:               input.Modified_by, // Using CreatedBy as ModifiedBy
		Remarks:                   input.Remarks,
		Paymentmode:               input.Paymentmode,
		Bankname:                  input.Bankname,
		Micrcode:                  input.Micrcode,
		Accounttype:               input.Accounttype,
		Ifsccode:                  input.Ifsccode,
		Accountnumber:             input.Accountnumber,
		Accountholdername:         input.Accountholdername,
		Branchname:                input.Branchname,
		Banktype:                  input.Banktype,
		Ismicrvalidated:           input.Ismicrvalidated,
		Policybond:                input.Policybond,
		Lrrb:                      input.Lrrb,
		Prb:                       input.Prb,
		Pdo_certificate:           input.Pdo_certificate,
		Application:               input.Application,
		Idproof_insurant:          input.Idproof_insurant,
		Addressproof_insurant:     input.Addressproof_insurant,
		Idproof_messenger:         input.Idproof_messenger,
		Addressproof_messenger:    input.Addressproof_messenger,
		Account_details_proof:     input.Account_details_proof,
		Others:                    input.Others,
		Cpc_office_id:             input.Cpc_office_id,
		PolicyNumber:              input.PolicyNumber,
	}

	// Call the repository
	result, err := activitiesInstance.surrenderRepo.SubmitApprovalRepo(ctx, req)
	if err != nil {
		return &SubmitApprovalResult{
			Success: false,
			Message: "Failed to submit approval: " + err.Error(),
			Status:  "ERROR",
		}, err
	}

	// Determine status based on request name
	status := "APPROVED"
	if input.RequestName == "REJECT" {
		status = "REJECTED"
	}

	return &SubmitApprovalResult{
		Success: true,
		Message: result,
		Status:  status,
	}, nil
}
