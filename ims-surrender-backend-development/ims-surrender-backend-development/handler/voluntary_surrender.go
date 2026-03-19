package handler

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"go.temporal.io/sdk/client"

	log "gitlab.cept.gov.in/it-2.0-common/n-api-log"
	serverHandler "gitlab.cept.gov.in/it-2.0-common/n-api-server/handler"
	serverRoute "gitlab.cept.gov.in/it-2.0-common/n-api-server/route"

	//apierrors "gitlab.cept.gov.in/it-2.0-common/n-api-errors"

	"gitlab.cept.gov.in/it-2.0-policy/surrender-service/core/domain"
	"gitlab.cept.gov.in/it-2.0-policy/surrender-service/core/port"
	"gitlab.cept.gov.in/it-2.0-policy/surrender-service/handler/response"
	repo "gitlab.cept.gov.in/it-2.0-policy/surrender-service/repo/postgres"
	"gitlab.cept.gov.in/it-2.0-policy/surrender-service/temporal/activities"
)

// VoluntarySurrenderHandler handles all voluntary surrender operations
// Business Rules: BR-SUR-001 to BR-SUR-018
// Functional Requirements: FR-SUR-001 to FR-SUR-009
type VoluntarySurrenderHandler struct {
	*serverHandler.Base
	surrenderRepo  *repo.SurrenderRequestRepository
	documentRepo   *repo.DocumentRepository
	temporalClient client.Client
	// External service placeholders
	policyService      PolicyServiceInterface
	loanService        LoanServiceInterface
	collectionsService CollectionsServiceInterface
	documentService    DocumentServiceInterface
}

// NewVoluntarySurrenderHandler creates a new voluntary surrender handler
func NewVoluntarySurrenderHandler(
	surrenderRepo *repo.SurrenderRequestRepository,
	documentRepo *repo.DocumentRepository,
	temporalClient client.Client,
) *VoluntarySurrenderHandler {
	base := serverHandler.New("Voluntary Surrender").SetPrefix("/v1").AddPrefix("/surrender")

	return &VoluntarySurrenderHandler{
		Base:           base,
		surrenderRepo:  surrenderRepo,
		documentRepo:   documentRepo,
		temporalClient: temporalClient,
		// Initialize placeholders (in real implementation, these would be injected)
		policyService:      NewMockPolicyService(),
		loanService:        NewMockLoanService(),
		collectionsService: NewMockCollectionsService(),
		documentService:    NewMockDocumentService(),
	}
}

// Routes defines all routes for voluntary surrender
func (h *VoluntarySurrenderHandler) Routes() []serverRoute.Route {
	return []serverRoute.Route{
		serverRoute.GET("/documents/status", h.GetDocumentUploadStatus).Name("Get Document Upload Status"),
		serverRoute.GET("/status", h.GetSurrenderStatus).Name("Get Surrender Status"),

		serverRoute.POST("/validate-eligibility", h.ValidateSurrenderEligibility).Name("Validate Surrender Eligibility"),
		serverRoute.POST("/calculate", h.CalculateSurrenderValue).Name("Calculate Surrender Value"),

		serverRoute.POST("/confirm", h.ConfirmSurrender).Name("Confirm Surrender Request"),
		serverRoute.POST("/documents/upload", h.UploadSurrenderDocument).Name("Upload Surrender Document"),

		serverRoute.POST("/submit-for-verification", h.SubmitForVerification).Name("Submit for Verification"),

		//check eligibility -> index - > DE -> QC -> Approval -> Payment
		serverRoute.GET("/check-eligibility/:policy_number", h.CheckSurrenderEligibility).Name("Check Surrender Eligibility"),
		serverRoute.POST("/index-surrender", h.IndexSurrender).Name("Index Surrender Request"), //1 -> initiate temporal
		serverRoute.GET("/de-pending/:office_id", h.DEPending).Name("Get all Pending DE requests"),
		serverRoute.GET("/qc-pending/:office_id", h.QCPending).Name("Get all Pending QC requests"),
		serverRoute.GET("/approval-pending/:office_id", h.ApprovalPending).Name("Get all Pending approval requests"),

		serverRoute.GET("/all-req-pending/:office_id", h.AllReqPending).Name("Get all Pending approval requests"),

		serverRoute.PUT("/submit-de", h.SubmitDE).Name("Submit Data Entry"),    //2 -> signal
		serverRoute.PUT("/submit-qc", h.SubmitQC).Name("Submit Quality Check"), //3 -> signal
		serverRoute.PUT("/submit-approval", h.SubmitApproval).Name("Submit Approval"),

		serverRoute.GET("/sr-details/:surrender_request_id", h.SRDetails).Name("Get attributes of ServiceRequest"),

		serverRoute.GET("/servicereq-staging-details/:surrender_request_id", h.ServiceReqStagingDetails).Name("Get all Staging details of a service request"),
		serverRoute.GET("/calculate-surrendervalue/:policy_number", h.CalcSurrenderValue).Name("Calculate Surrender Value"),

		// serverRoute.POST("/index-surrender", h.IndexSurrender).Name("Index Surrender Request"), //1 -> initiate temporal
		// serverRoute.PUT("/submit-de", h.SubmitDE).Name("Submit Data Entry"),                    //2 -> signal
		// serverRoute.PUT("/submit-qc", h.SubmitQC).Name("Submit Quality Check"),                 //3 -> signal
		// serverRoute.PUT("/submit-approval", h.SubmitApproval).Name("Submit Approval"),          //4 -> signal
	}

}

type PolicyNoInput struct {
	PolicyNumber string `uri:"policy_number" validate:"required"`
}

// type IndexSurrenderRequest struct {
// 	PolicyNumber string `json:"policy_number" validate:"required"`
// }

func (h *VoluntarySurrenderHandler) CheckSurrenderEligibility(sctx *serverRoute.Context, req PolicyNoInput) (interface{}, error) {

	log.Error(sctx.Ctx, "policy number is : %s", req.PolicyNumber)
	policy, err := h.surrenderRepo.FindByPolicyNumber(sctx.Ctx, req.PolicyNumber)
	if err != nil {
		log.Error(sctx.Ctx, "policy number not found in handler %s", err)
		return h.createIneligibleResponse1(req.PolicyNumber, []string{
			fmt.Sprintf("Policy details not found - Wrong Policy number or Policy number doesnot exist."),
		}), nil
	}
	log.Error(sctx.Ctx, "policy status is : %s", policy.Policy_status)

	validStatuses := []string{"AP", "IL", "AL"}
	isValidStatus := false
	for _, status := range validStatuses {
		if policy.Policy_status == status {
			isValidStatus = true
			break
		}
	}
	if !isValidStatus {
		return h.createIneligibleResponse1(req.PolicyNumber, []string{
			fmt.Sprintf("Policy status '%s' is not eligible for surrender. Only In Force policies (AP, IL, AL) can be surrendered.", policy.Policy_status),
		}), nil
	}

	if policy.Maturity_date.Before(time.Now()) {
		return h.createIneligibleResponse1(req.PolicyNumber, []string{
			"Policy has reached maturity. Please process through Maturity Claims.",
		}), nil
	}

	ineligibleProducts := []string{"AEA", "AEA-10", "GY"}
	for _, prod := range ineligibleProducts {
		if policy.Product_name == prod {
			return h.createIneligibleResponse1(req.PolicyNumber, []string{
				fmt.Sprintf("Product '%s'  is not eligible for surrender.", policy.Product_name),
			}), nil
		}
	}

	r := &response.EligibilityResponse{
		StatusCodeAndMessage: port.StatusCodeAndMessage{
			StatusCode: 200,
			Success:    true,
			Message:    "Policy is eligible for surrender",
		},
	}

	return r, nil

}

func mapToIndexSurrenderInput(req IndexSurrenderRequest) domain.IndexSurrenderRequestInput {
	return domain.IndexSurrenderRequestInput{
		PolicyNumber:              req.PolicyNumber,
		Surrender_request_channel: req.Surrender_request_channel,
		Indexing_office_id:        req.Indexing_office_id,
		Cpc_office_id:             req.Cpc_office_id,
		Created_by:                req.Created_by,
		Modified_by:               req.Modified_by,
		Remarks:                   req.Remarks,
		Paidupvalue:               req.Paidupvalue,
		Bonus:                     req.Bonus,
		Grossamount:               req.Grossamount,
		Loanprincipal:             req.Loanprincipal,
		Loaninterest:              req.Loaninterest,
		Surrenderfactor:           req.Surrenderfactor,
		Othercharges:              req.Othercharges,
		Surrendervalue:            req.Surrendervalue,
		Bonusrate:                 req.Bonusrate,
		Bonusamount:               req.Bonusamount,
		Sumassured:                req.Sumassured,
		Paid_to_date:              req.Paid_to_date,
		Polissdate:                req.Polissdate,
		Maturitydate:              req.Maturitydate,
		Productcode:               req.Productcode,
		Dob:                       req.Dob,
		Unpaidprem:                req.Unpaidprem,
		Def:                       req.Def,
		Stage_name:                req.Stage_name,
	}
}

func (h *VoluntarySurrenderHandler) IndexSurrender(sctx *serverRoute.Context, req IndexSurrenderRequest) (interface{}, error) {
	// Surrender requests must now be initiated through Policy Management.
	// PM dispatches SurrenderProcessingWorkflow as a child workflow, which
	// creates the surrender_request record via IndexSurrenderActivity.
	// This endpoint is no longer the entry point.
	log.Info(sctx.Ctx, "IndexSurrender called — surrender initiation must go through Policy Management")
	return response.GetDEPendingResponse{
		StatusCode: 410,
		Success:    false,
		Message:    "This endpoint is decommissioned. Surrender requests must be initiated through Policy Management (POST /v1/policies/{policy_number}/requests/surrender).",
		Data:       nil,
	}, nil
}

func (h *VoluntarySurrenderHandler) SRDetails(sctx *serverRoute.Context, req SRIDDetailsRequest) (interface{}, error) {

	log.Error(sctx.Ctx, "policy number is : %s", req.Surrender_request_id)
	res, err := h.surrenderRepo.SRDetailsRepo(sctx.Ctx, req.Surrender_request_id)

	log.Error(sctx.Ctx, "servicereqID from handler : %s", res)
	if err != nil {
		log.Error(sctx.Ctx, "SR not found in handler %s", err)
		return h.createIneligibleResponse1(req.Surrender_request_id, []string{
			fmt.Sprintf("Service request ID not found."),
		}), nil
	}

	r := res[0]
	finaldata := domain.SRDetailsOutput1{
		PolicyDetails: domain.PolicyDetails{
			Policy_number:        r.Policy_number,
			Surrender_request_id: r.Surrender_request_id,
			Dob:                  r.Dob,
			Sumassured:           r.Sumassured,
			Polissdate:           r.Polissdate,
			Paid_to_date:         r.Paid_to_date,
			Maturitydate:         r.Maturitydate,
		},
		SurrenderCalculation: domain.SurrenderCalculation{
			Paidupvalue:     r.Paidupvalue,
			Bonus:           r.Bonus,
			Grossamount:     r.Grossamount,
			Loanprincipal:   r.Loanprincipal,
			Loaninterest:    r.Loaninterest,
			Surrenderfactor: r.Surrenderfactor,
			Othercharges:    r.Othercharges,
			Surrendervalue:  r.Surrendervalue,
			Bonusrate:       r.Bonusrate,
			Bonusamount:     r.Bonusamount,
			Unpaidprem:      r.Unpaidprem,
			Def:             r.Def,
		},
		BankDetails: domain.BankDetails{
			Paymentmode:       r.Paymentmode,
			Bankname:          r.Bankname,
			Micrcode:          r.Micrcode,
			Accounttype:       r.Accounttype,
			Ifsccode:          r.Ifsccode,
			Accountnumber:     r.Accountnumber,
			Accountholdername: r.Accountholdername,
			Branchname:        r.Branchname,
			Banktype:          r.Banktype,
			Ismicrvalidated:   r.Ismicrvalidated,
		},
		Documents: domain.Documents{
			Policybond:             r.Policybond,
			Lrrb:                   r.Lrrb,
			Prb:                    r.Prb,
			Pdo_certificate:        r.Pdo_certificate,
			Application:            r.Application,
			Idproof_insurant:       r.Idproof_insurant,
			Addressproof_insurant:  r.Addressproof_insurant,
			Idproof_messenger:      r.Idproof_messenger,
			Addressproof_messenger: r.Addressproof_messenger,
			Account_details_proof:  r.Account_details_proof,
			Others:                 r.Others,
		},
		AdditionalInfo: domain.AdditionalInfo{
			Reason:  r.Reason,
			Remarks: r.Remarks,
		},
	}
	return response.GetDEPendingResponse{
		StatusCode: 200,
		Success:    true,
		Message:    "Request fetched successfully",
		Data:       []domain.SRDetailsOutput1{finaldata},
	}, nil

}

func (h *VoluntarySurrenderHandler) DEPending(sctx *serverRoute.Context, req GetDEPendingRequest) (interface{}, error) {

	log.Error(sctx.Ctx, "office id  is : %s", req.Oid)
	details, err := h.surrenderRepo.DEPendingRepo(sctx.Ctx, req.Oid)

	log.Error(sctx.Ctx, "ofc id  : %s", req.Oid)
	if err != nil {
		log.Error(sctx.Ctx, "SR not found in handler %s", err)
		//return h.createIneligibleResponse1(req.Oid, []string{
		//fmt.Sprintf("Service request ID not found."),
		//}), nil
	}
	return response.GetDEPendingResponse{
		StatusCode: 200,
		Success:    true,
		Message:    "Request fetched successfully",
		Data:       details,
	}, nil

}

func (h *VoluntarySurrenderHandler) QCPending(sctx *serverRoute.Context, req GetDEPendingRequest) (interface{}, error) {

	log.Error(sctx.Ctx, "office id  is : %s", req.Oid)
	details, err := h.surrenderRepo.QCPendingRepo(sctx.Ctx, req.Oid)

	log.Error(sctx.Ctx, "ofc id  : %s", req.Oid)
	if err != nil {
		log.Error(sctx.Ctx, "SR not found in handler %s", err)
		//return h.createIneligibleResponse1(req.Oid, []string{
		//fmt.Sprintf("Service request ID not found."),
		//}), nil
	}
	return response.GetDEPendingResponse{
		StatusCode: 200,
		Success:    true,
		Message:    "Request fetched successfully",
		Data:       details,
	}, nil

}

func (h *VoluntarySurrenderHandler) ApprovalPending(sctx *serverRoute.Context, req GetDEPendingRequest) (interface{}, error) {

	log.Error(sctx.Ctx, "office id  is : %s", req.Oid)
	details, err := h.surrenderRepo.ApprovalPendingRepo(sctx.Ctx, req.Oid)

	log.Error(sctx.Ctx, "ofc id  : %s", req.Oid)
	if err != nil {
		log.Error(sctx.Ctx, "SR not found in handler %s", err)
		//return h.createIneligibleResponse1(req.Oid, []string{
		//fmt.Sprintf("Service request ID not found."),
		//}), nil
	}
	return response.GetDEPendingResponse{
		StatusCode: 200,
		Success:    true,
		Message:    "Request fetched successfully",
		Data:       details,
	}, nil

}

func (h *VoluntarySurrenderHandler) AllReqPending(sctx *serverRoute.Context, req GetDEPendingRequest) (interface{}, error) {

	log.Error(sctx.Ctx, "office id  is : %s", req.Oid)
	details, err := h.surrenderRepo.AllReqPendingRepo(sctx.Ctx, req.Oid)

	log.Error(sctx.Ctx, "ofc id  : %s", req.Oid)
	if err != nil {
		log.Error(sctx.Ctx, "SR not found in handler %s", err)
		//return h.createIneligibleResponse1(req.Oid, []string{
		//fmt.Sprintf("Service request ID not found."),
		//}), nil
	}
	return response.GetDEPendingResponse{
		StatusCode: 200,
		Success:    true,
		Message:    "Request fetched successfully",
		Data:       details,
	}, nil

}

func (h *VoluntarySurrenderHandler) ServiceReqStagingDetails(sctx *serverRoute.Context, req SRIDDetailsRequest) (interface{}, error) {

	log.Error(sctx.Ctx, "office id  is : %s", req.Surrender_request_id)
	details, err := h.surrenderRepo.ServiceReqStagingDetailsRepo(sctx.Ctx, req.Surrender_request_id)

	log.Error(sctx.Ctx, "ofc id  : %s", req.Surrender_request_id)
	if err != nil {
		log.Error(sctx.Ctx, "SR not found in handler %s", err)
		//return h.createIneligibleResponse1(req.Oid, []string{
		//fmt.Sprintf("Service request ID not found."),
		//}), nil
	}
	return response.GetDEPendingResponse{
		StatusCode: 200,
		Success:    true,
		Message:    "Request fetched successfully",
		Data:       details,
	}, nil

}

func (h *VoluntarySurrenderHandler) SubmitDE(sctx *serverRoute.Context, req1 SubmitDERequest) (interface{}, error) {

	log.Error(sctx.Ctx, "Surrender_request_id  is : %s", req1.Surrender_request_id)

	// Look up the Temporal workflow ID stored during IndexSurrenderActivity so we
	// can signal the correct SurrenderProcessingWorkflow instance.
	workflowID, err := h.surrenderRepo.GetWorkflowIDBySurrenderRequestID(sctx.Ctx, req1.Surrender_request_id)
	if err != nil {
		log.Error(sctx.Ctx, "Failed to resolve workflow ID for DE: %s", err)
		return response.GetDEPendingResponse{
			StatusCode: 404,
			Success:    false,
			Message:    fmt.Sprintf("Surrender request not found or workflow not started: %v", err),
			Data:       nil,
		}, nil
	}

	// Create activity input from request
	activityInput := activities.SubmitDEInput{
		SurrenderRequestID:      req1.Surrender_request_id,
		SurrenderRequestChannel: req1.Surrender_request_channel,
		RequestName:             req1.Request_name,
		CurrentStageName:        req1.Current_stage_name,
		CreatedBy:               req1.Created_by,
		Modified_by:             req1.Modified_by,
		Remarks:                 req1.Remarks,
		Paymentmode:             req1.Paymentmode,
		Bankname:                req1.Bankname,
		Micrcode:                req1.Micrcode,
		Accounttype:             req1.Accounttype,
		Ifsccode:                req1.Ifsccode,
		Accountnumber:           req1.Accountnumber,
		Accountholdername:       req1.Accountholdername,
		Branchname:              req1.Branchname,
		Banktype:                req1.Banktype,
		Ismicrvalidated:         req1.Ismicrvalidated,
		Policybond:              req1.Policybond,
		Lrrb:                    req1.Lrrb,
		Prb:                     req1.Prb,
		Pdo_certificate:         req1.Pdo_certificate,
		Application:             req1.Application,
		Idproof_insurant:        req1.Idproof_insurant,
		Addressproof_insurant:   req1.Addressproof_insurant,
		Idproof_messenger:       req1.Idproof_messenger,
		Addressproof_messenger:  req1.Addressproof_messenger,
		Account_details_proof:   req1.Account_details_proof,
		Others:                  req1.Others, //Output:
		Cpc_office_id:           req1.Cpc_office_id,
		PolicyNumber:            req1.PolicyNumber,
	}

	err = h.temporalClient.SignalWorkflow(sctx.Ctx, workflowID, "", "de-completed", activityInput)
	if err != nil {
		log.Error(sctx.Ctx, "Failed to send DE signal: %s", err)
		return response.GetDEPendingResponse{
			StatusCode: 500,
			Success:    false,
			Message:    fmt.Sprintf("Failed to submit DE request: %v", err),
			Data:       nil,
		}, nil
	}

	log.Info(sctx.Ctx, "DE request submitted to workflow", "WorkflowID", workflowID)

	return response.GetDEPendingResponse{
		StatusCode: 200,
		Success:    true,
		Message:    "DE request submitted successfully",
		Data:       nil,
	}, nil

}

func (h *VoluntarySurrenderHandler) SubmitQC(sctx *serverRoute.Context, req1 SubmitDERequest) (interface{}, error) {

	log.Error(sctx.Ctx, "Surrender_request_id  is : %s", req1.Surrender_request_id)

	// Look up the Temporal workflow ID stored during IndexSurrenderActivity so we
	// can signal the correct SurrenderProcessingWorkflow instance.
	workflowID, err := h.surrenderRepo.GetWorkflowIDBySurrenderRequestID(sctx.Ctx, req1.Surrender_request_id)
	if err != nil {
		log.Error(sctx.Ctx, "Failed to resolve workflow ID for QC: %s", err)
		return response.GetDEPendingResponse{
			StatusCode: 404,
			Success:    false,
			Message:    fmt.Sprintf("Surrender request not found or workflow not started: %v", err),
			Data:       nil,
		}, nil
	}

	// Create activity input from request
	// activityInput := activities.SubmitDEInput{
	// 	SurrenderRequestID:      req1.Surrender_request_id,
	// 	SurrenderRequestChannel: req1.Surrender_request_channel,
	// 	RequestName:             req1.Request_name,
	// 	CurrentStageName:        req1.Current_stage_name,
	// 	CreatedBy:               req1.Created_by,
	// }

	activityInput := activities.SubmitQCInput{
		SurrenderRequestID:      req1.Surrender_request_id,
		SurrenderRequestChannel: req1.Surrender_request_channel,
		RequestName:             req1.Request_name,
		CurrentStageName:        req1.Current_stage_name,
		CreatedBy:               req1.Created_by,
		Modified_by:             req1.Modified_by,
		Remarks:                 req1.Remarks,
		Paymentmode:             req1.Paymentmode,
		Bankname:                req1.Bankname,
		Micrcode:                req1.Micrcode,
		Accounttype:             req1.Accounttype,
		Ifsccode:                req1.Ifsccode,
		Accountnumber:           req1.Accountnumber,
		Accountholdername:       req1.Accountholdername,
		Branchname:              req1.Branchname,
		Banktype:                req1.Banktype,
		Ismicrvalidated:         req1.Ismicrvalidated,
		Policybond:              req1.Policybond,
		Lrrb:                    req1.Lrrb,
		Prb:                     req1.Prb,
		Pdo_certificate:         req1.Pdo_certificate,
		Application:             req1.Application,
		Idproof_insurant:        req1.Idproof_insurant,
		Addressproof_insurant:   req1.Addressproof_insurant,
		Idproof_messenger:       req1.Idproof_messenger,
		Addressproof_messenger:  req1.Addressproof_messenger,
		Account_details_proof:   req1.Account_details_proof,
		Others:                  req1.Others, //Output:
		Cpc_office_id:           req1.Cpc_office_id,
		PolicyNumber:            req1.PolicyNumber,
	}

	err = h.temporalClient.SignalWorkflow(sctx.Ctx, workflowID, "", "qc-completed", activityInput)
	if err != nil {
		log.Error(sctx.Ctx, "Failed to send QC signal: %s", err)
		return response.GetDEPendingResponse{
			StatusCode: 500,
			Success:    false,
			Message:    fmt.Sprintf("Failed to submit QC requesttt: %v", err),
			Data:       nil,
		}, nil
	}

	log.Info(sctx.Ctx, "QC request submitted to workflow", "WorkflowID", workflowID)

	return response.GetDEPendingResponse{
		StatusCode: 200,
		Success:    true,
		Message:    "QC request submitted successfully",
		Data:       nil,
	}, nil

}

func (h *VoluntarySurrenderHandler) SubmitApproval(sctx *serverRoute.Context, req1 SubmitDERequest) (interface{}, error) {

	log.Error(sctx.Ctx, "Surrender_request_id  is : %s", req1.Surrender_request_id)

	// Look up the Temporal workflow ID stored during IndexSurrenderActivity so we
	// can signal the correct SurrenderProcessingWorkflow instance.
	workflowID, err := h.surrenderRepo.GetWorkflowIDBySurrenderRequestID(sctx.Ctx, req1.Surrender_request_id)
	if err != nil {
		log.Error(sctx.Ctx, "Failed to resolve workflow ID for Approval: %s", err)
		return response.GetDEPendingResponse{
			StatusCode: 404,
			Success:    false,
			Message:    fmt.Sprintf("Surrender request not found or workflow not started: %v", err),
			Data:       nil,
		}, nil
	}

	// Create activity input from request
	// activityInput := activities.SubmitApprovalInput{
	// 	SurrenderRequestID:      req1.Surrender_request_id,
	// 	SurrenderRequestChannel: req1.Surrender_request_channel,
	// 	RequestName:             req1.Request_name,
	// 	CurrentStageName:        req1.Current_stage_name,
	// 	CreatedBy:               req1.Created_by,
	// }

	activityInput := activities.SubmitApprovalInput{
		SurrenderRequestID:      req1.Surrender_request_id,
		SurrenderRequestChannel: req1.Surrender_request_channel,
		RequestName:             req1.Request_name,
		CurrentStageName:        req1.Current_stage_name,
		CreatedBy:               req1.Created_by,
		Modified_by:             req1.Modified_by,
		Remarks:                 req1.Remarks,
		Paymentmode:             req1.Paymentmode,
		Bankname:                req1.Bankname,
		Micrcode:                req1.Micrcode,
		Accounttype:             req1.Accounttype,
		Ifsccode:                req1.Ifsccode,
		Accountnumber:           req1.Accountnumber,
		Accountholdername:       req1.Accountholdername,
		Branchname:              req1.Branchname,
		Banktype:                req1.Banktype,
		Ismicrvalidated:         req1.Ismicrvalidated,
		Policybond:              req1.Policybond,
		Lrrb:                    req1.Lrrb,
		Prb:                     req1.Prb,
		Pdo_certificate:         req1.Pdo_certificate,
		Application:             req1.Application,
		Idproof_insurant:        req1.Idproof_insurant,
		Addressproof_insurant:   req1.Addressproof_insurant,
		Idproof_messenger:       req1.Idproof_messenger,
		Addressproof_messenger:  req1.Addressproof_messenger,
		Account_details_proof:   req1.Account_details_proof,
		Others:                  req1.Others, //Output:
		Cpc_office_id:           req1.Cpc_office_id,
		PolicyNumber:            req1.PolicyNumber,
	}

	err = h.temporalClient.SignalWorkflow(sctx.Ctx, workflowID, "", "approval-completed", activityInput)
	if err != nil {
		log.Error(sctx.Ctx, "Failed to send Approval signal: %s", err)
		return response.GetDEPendingResponse{
			StatusCode: 500,
			Success:    false,
			Message:    fmt.Sprintf("Failed to submit Approval request: %v", err),
			Data:       nil,
		}, nil
	}

	log.Info(sctx.Ctx, "Approval request submitted to workflow", "WorkflowID", workflowID)

	return response.GetDEPendingResponse{
		StatusCode: 200,
		Success:    true,
		Message:    "Approval request submitted successfully",
		Data:       nil,
	}, nil

}

func (h *VoluntarySurrenderHandler) CalcSurrenderValue(sctx *serverRoute.Context, req1 PolicyNoInput) (interface{}, error) {

	log.Error(sctx.Ctx, "office id  is : %s", req1.PolicyNumber)

	p, err := h.surrenderRepo.FindByPolicyNumber(sctx.Ctx, req1.PolicyNumber)
	if err != nil {
		log.Error(sctx.Ctx, "policy number not found in handler %s", err)
		return h.createIneligibleResponse1(req1.PolicyNumber, []string{
			fmt.Sprintf("Policy details not found - Wrong Policy number or Policy number doesnot exist."),
		}), nil
	}

	log.Error(sctx.Ctx, "policy no is : %s", p.Policy_number)
	log.Error(sctx.Ctx, "sum assured is : %s", p.Sum_assured)
	log.Error(sctx.Ctx, "paid to date is : %s", p.Paid_to_date)
	log.Error(sctx.Ctx, "pol issue date is : %s", p.Polissdate)
	log.Error(sctx.Ctx, "maturity date is : %s", p.Maturity_date)
	log.Error(sctx.Ctx, "bonus is : %s", p.Totalbonus)

	log.Error(sctx.Ctx, "todays date is: %s", time.Now().Format("2006-01-02"))
	log.Error(sctx.Ctx, "product code is : %s", p.Product_code)
	log.Error(sctx.Ctx, "dob is : %s", p.Dob)

	details, err := h.surrenderRepo.CalcSurrenderValuerepo(sctx.Ctx, p.Policy_number, p.Product_code, p.Polissdate, p.Maturity_date, p.Dob)

	bonus, err := h.surrenderRepo.CalcBonusValuerepo(sctx.Ctx, p.Product_code, p.Polissdate)

	//totalBonus := 0
	var cumulative float64
	var result []domain.AccruedBonusOutput

	for _, bonus1 := range bonus {
		// if bonus.Bonus_year >= fromYear && bonus.Bonus_year <= toYear {
		//totalBonus += bonus1.Bonus_rate_per_1000_sa

		bonusValue := float64(bonus1.Bonus_rate_per_1000_sa)

		cumulative += bonusValue

		row := domain.AccruedBonusOutput{

			BonusRate:  float64(bonus1.Bonus_rate_per_1000_sa),
			FromDate:   bonus1.Bonus_from,
			ToDate:     bonus1.Bonus_to,
			BonusValue: bonusValue,
			Cumulative: cumulative,
		}

		result = append(result, row)

		//}
	}

	log.Error(sctx.Ctx, "Bonus is: %f", cumulative)

	// if len(details) == 0 {
	// 	return nil, fmt.Errorf("no surrender factor found")
	// }

	sfObj := details[0]
	sfValue := sfObj.Surrender_factor
	log.Error(sctx.Ctx, "Surrender Factor is: %f", sfValue)

	// // log.Error(sctx.Ctx, "ofc id  : %s", req1.PolicyNumber)
	// if err != nil {
	// 	log.Error(sctx.Ctx, "SR not found in handler %s", err)
	// 	//return h.createIneligibleResponse1(req.Oid, []string{
	// 	//fmt.Sprintf("Service request ID not found."),
	// 	//}), nil
	// }

	var paidupvalue float64
	tot_prem_paid := 0
	tot_prem_payable := 0

	tot_prem_paid = MonthsBetween(p.Polissdate, p.Paid_to_date)
	tot_prem_payable = MonthsBetween(p.Polissdate, p.Maturity_date)

	//paidupvalue = (p.Sum_assured * tot_prem_paid) / tot_prem_payable
	paidupvalue = (float64(p.Sum_assured) * float64(tot_prem_paid)) / float64(tot_prem_payable)

	var surrendervalue float64

	surrendervalue = (float64(paidupvalue) + float64(cumulative)) * sfValue

	log.Error(sctx.Ctx, "Surrender Value is: %f", surrendervalue)

	finalData := domain.SurrenderCalculationResponse{
		PolicyNumber:             p.Policy_number,
		SumAssured:               float64(p.Sum_assured),
		PaidToDate:               p.Paid_to_date,
		Polissdate:               p.Polissdate,
		MaturityDate:             p.Maturity_date,
		ProductCode:              p.Product_code,
		Dob:                      p.Dob,
		Cumulative:               cumulative,
		UnpaidPremiums:           1, // This would be calculated based on the policy details and payment history
		DefOnUnpPrem:             1, // This would be calculated based on the policy details and payment history
		Outstandingloanprinciple: 100,
		Outstandingloaninterest:  10,
		OtherCharges:             0,
		SFValue:                  sfValue,
		PaidUpValue:              paidupvalue,
		SurrenderValue:           surrendervalue,
		BonusDetails:             result,
	}

	return response.GetDEPendingResponse{
		StatusCode: 200,
		Success:    true,
		Message:    "Request fetched successfully",
		Data:       []domain.SurrenderCalculationResponse{finalData},
	}, nil

}

func MonthsBetween(from, to time.Time) int {
	if to.Before(from) {
		return 0
	}

	years := to.Year() - from.Year()
	months := int(to.Month()) - int(from.Month())

	totalMonths := years*12 + months

	// Adjust if the current month day is less than start day
	if to.Day() < from.Day() {
		totalMonths--
	}

	return totalMonths
}

// ValidateSurrenderEligibility validates if a policy is eligible for surrender
// POST /v1/surrender/validate-eligibility
// Business Rules: BR-SUR-001, BR-SUR-002, BR-SUR-003, BR-SUR-004
// Validation Rules: VR-SUR-001, VR-SUR-002, VR-SUR-003, VR-SUR-004
func (h *VoluntarySurrenderHandler) ValidateSurrenderEligibility(sctx *serverRoute.Context, req ValidateEligibilityRequest) (interface{}, error) {
	// Get policy details using PolicyID as a string
	policy, err := h.policyService.GetPolicyByID(sctx.Ctx, req.PolicyID)
	if err != nil {
		log.Error(sctx.Ctx, "Failed to get policy: %v", err)
		return nil, fmt.Errorf("policy not found")
	}

	// BR-SUR-002: Check policy status (must be In Force - AP, IL, or AL)
	// VR-SUR-002: Policy must be in active status
	validStatuses := []string{"AP", "IL", "AL"}
	isValidStatus := false
	for _, status := range validStatuses {
		if policy.Status == status {
			isValidStatus = true
			break
		}
	}
	if !isValidStatus {
		return h.createIneligibleResponse(policy, []string{
			fmt.Sprintf("Policy status '%s' is not eligible for surrender. Only In Force policies (AP, IL, AL) can be surrendered.", policy.Status),
		}), nil
	}

	// BR-SUR-003: Check if policy has reached maturity
	// VR-SUR-003: Matured policies must use maturity claims, not surrender
	if policy.MaturityDate.Before(time.Now()) && policy.ProductCode != "WLA" {
		return h.createIneligibleResponse(policy, []string{
			"Policy has reached maturity. Please process through Maturity Claims.",
		}), nil
	}

	// BR-SUR-004: Check if product type allows surrender
	// VR-SUR-004: AEA and GY products are not eligible
	ineligibleProducts := []string{"AEA", "AEA-10", "GY"}
	for _, prod := range ineligibleProducts {
		if policy.ProductCode == prod {
			return h.createIneligibleResponse(policy, []string{
				fmt.Sprintf("Product '%s' (%s) is not eligible for surrender.", policy.ProductCode, policy.ProductName),
			}), nil
		}
	}

	// BR-SUR-001: Check minimum premium payment period
	// VR-SUR-001: Validate premiums paid against minimum requirement
	minimumPremiums := h.getMinimumPremiumsByProduct(policy.ProductCode)
	premiumsPaid := policy.PremiumsPaid

	if premiumsPaid < minimumPremiums {
		return h.createIneligibleResponse(policy, []string{
			fmt.Sprintf("Insufficient premiums paid. Required: %d years, Paid: %d years.", minimumPremiums, premiumsPaid),
		}), nil
	}

	// BR-SUR-014: Check if there's already an active surrender request
	existingRequest, found, err := h.surrenderRepo.FindActiveByPolicyID(sctx.Ctx, req.PolicyID)
	if err != nil {
		log.Error(sctx.Ctx, "Failed to check existing surrender: %v", err)
		return nil, fmt.Errorf("failed to validate eligibility")
	}
	if found {
		return h.createIneligibleResponse(policy, []string{
			fmt.Sprintf("An active surrender request already exists (Request #%s) with status '%s'.", existingRequest.RequestNumber, existingRequest.Status),
		}), nil
	}

	// All checks passed - policy is eligible
	log.Info(sctx.Ctx, "Policy %s is eligible for surrender", policy.PolicyNumber)

	r := &response.EligibilityEligibleResponse{
		StatusCodeAndMessage: port.StatusCodeAndMessage{
			StatusCode: 200,
			Success:    true,
			Message:    "Policy is eligible for surrender",
		},
		Data: response.EligibilityEligibleData{
			Eligible:            true,
			PolicyID:            policy.ID,
			PolicyNumber:        policy.PolicyNumber,
			ProductCode:         policy.ProductCode,
			ProductName:         policy.ProductName,
			PremiumsPaid:        premiumsPaid,
			MinimumPremiumsPaid: minimumPremiums,
			PolicyStatus:        policy.Status,
			Message:             "Your policy meets all eligibility criteria for surrender. You may proceed with the surrender process.",
		},
	}

	return r, nil
}

// CalculateSurrenderValue calculates the surrender value for a policy
// POST /v1/surrender/calculate
// Business Rules: BR-SUR-006, BR-SUR-007, BR-SUR-008, BR-SUR-009, BR-SUR-010, BR-SUR-011
func (h *VoluntarySurrenderHandler) CalculateSurrenderValue(sctx *serverRoute.Context, req CalculateSurrenderRequest) (*response.CalculateSurrenderResponse, error) {
	// Get policy details using PolicyID as a string
	policy, err := h.policyService.GetPolicyByID(sctx.Ctx, req.PolicyID)
	if err != nil {
		log.Error(sctx.Ctx, "Failed to get policy: %v", err)
		return nil, fmt.Errorf("policy not found")
	}

	// BR-SUR-006: Calculate Paid-Up Value
	// Formula: (Sum Assured × Premiums Paid) / Total Premiums
	paidUpValue := (policy.SumAssured * float64(policy.PremiumsPaid)) / float64(policy.TotalPremiums)

	// BR-SUR-005: Get bonus details and calculate total bonus
	bonusDetails := h.calculateBonusDetails(policy)
	totalBonus := 0.0
	for _, bonus := range bonusDetails {
		totalBonus += bonus.BonusAmount
	}

	// BR-SUR-007: Get surrender factor from actuarial tables
	surrenderFactor := h.getSurrenderFactor(policy.ProductCode, policy.Term, policy.AgeLast)

	// BR-SUR-008: Calculate Gross Surrender Value
	// GSV = (Paid-Up Value + Total Bonus) × Surrender Factor
	grossSurrenderValue := (paidUpValue + totalBonus) * surrenderFactor

	// BR-SUR-009: Calculate deductions
	// Get unpaid premiums
	unpaidPremiums, err := h.collectionsService.GetUnpaidPremiums(sctx.Ctx, req.PolicyID)
	if err != nil {
		log.Error(sctx.Ctx, "Failed to get unpaid premiums: %v", err)
		unpaidPremiums = 0
	}

	// Get loan details
	loanDetails, err := h.loanService.GetLoanDetails(sctx.Ctx, req.PolicyID)
	if err != nil {
		log.Error(sctx.Ctx, "Failed to get loan details: %v", err)
		loanDetails = &LoanDetails{Principal: 0, Interest: 0}
	}

	totalDeductions := unpaidPremiums + loanDetails.Principal + loanDetails.Interest

	// BR-SUR-010: Calculate Net Surrender Value
	// NSV = GSV - Unpaid Premiums - Loan Principal - Loan Interest
	netSurrenderValue := grossSurrenderValue - totalDeductions

	// BR-SUR-011: Predict disposition based on prescribed limit
	prescribedLimit := h.getPrescribedLimit(policy.ProductCode)
	disposition := h.predictDisposition(netSurrenderValue, prescribedLimit, paidUpValue)

	log.Info(sctx.Ctx, "Calculated surrender value for policy %s: GSV=%f, NSV=%f", policy.PolicyNumber, grossSurrenderValue, netSurrenderValue)

	r := &response.CalculateSurrenderResponse{
		StatusCodeAndMessage: port.CustomSuccess,
		Data: response.CalculationData{
			CalculationBreakdown: response.CalculationBreakdownData{
				PolicyID:            policy.ID,
				CalculationDate:     time.Now().Format(time.RFC3339),
				SumAssured:          policy.SumAssured,
				PremiumsPaid:        policy.PremiumsPaid,
				TotalPremiums:       policy.TotalPremiums,
				PaidUpValue:         paidUpValue,
				BonusDetails:        bonusDetails,
				TotalBonus:          totalBonus,
				SurrenderFactor:     surrenderFactor,
				GrossSurrenderValue: grossSurrenderValue,
				Deductions: response.DeductionsData{
					UnpaidPremiums:  unpaidPremiums,
					LoanPrincipal:   loanDetails.Principal,
					LoanInterest:    loanDetails.Interest,
					TotalLoan:       loanDetails.Principal + loanDetails.Interest,
					OtherCharges:    0,
					TotalDeductions: totalDeductions,
				},
				NetSurrenderValue: netSurrenderValue,
			},
			DisbursementOptions: response.DisbursementOptionsData{
				CashAvailable:   true,
				ChequeAvailable: true,
				PayeeDetails: response.PayeeDetailsData{
					PayeeName:    policy.PolicyholderName,
					PayeeAddress: policy.Address,
					IsAssigned:   policy.IsAssigned,
					AssigneeName: policy.AssigneeName,
				},
			},
			DispositionPrediction: disposition,
		},
	}

	return r, nil
}

// ConfirmSurrender creates a new surrender request
// POST /v1/surrender/confirm
// Business Rule: BR-SUR-013
func (h *VoluntarySurrenderHandler) ConfirmSurrender(sctx *serverRoute.Context, req ConfirmSurrenderRequest) (*response.ConfirmSurrenderResponse, error) {
	// Re-run eligibility check using PolicyID as a string
	policy, err := h.policyService.GetPolicyByID(sctx.Ctx, req.PolicyID)
	if err != nil {
		log.Error(sctx.Ctx, "Failed to get policy: %v", err)
		return nil, fmt.Errorf("policy not found")
	}

	// Recalculate surrender value
	calculation, err := h.CalculateSurrenderValue(sctx, CalculateSurrenderRequest{PolicyID: req.PolicyID})
	if err != nil {
		log.Error(sctx.Ctx, "Failed to calculate surrender value: %v", err)
		return nil, err
	}

	// BR-SUR-013: Generate unique request number
	requestNumber := h.generateRequestNumber(policy.PolicyNumber)

	// Create surrender request with mock user ID (in real impl, get from context)
	mockUserID := uuid.New()

	surrenderRequest := domain.PolicySurrenderRequest{
		PolicyID:                     req.PolicyID,
		RequestNumber:                requestNumber,
		RequestType:                  domain.SurrenderRequestTypeVoluntary,
		RequestDate:                  time.Now(),
		SurrenderValueCalculatedDate: time.Now(),
		GrossSurrenderValue:          calculation.Data.CalculationBreakdown.GrossSurrenderValue,
		NetSurrenderValue:            calculation.Data.CalculationBreakdown.NetSurrenderValue,
		PaidUpValue:                  calculation.Data.CalculationBreakdown.PaidUpValue,
		BonusAmount:                  &calculation.Data.CalculationBreakdown.TotalBonus,
		SurrenderFactor:              calculation.Data.CalculationBreakdown.SurrenderFactor,
		UnpaidPremiumsDeduction:      calculation.Data.CalculationBreakdown.Deductions.UnpaidPremiums,
		LoanDeduction:                calculation.Data.CalculationBreakdown.Deductions.TotalLoan,
		DisbursementMethod:           domain.DisbursementMethod(req.DisbursementMethod),
		DisbursementAmount:           calculation.Data.CalculationBreakdown.NetSurrenderValue,
		Reason:                       req.Reason,
		Status:                       domain.SurrenderStatusPendingDocumentUpload,
		Owner:                        domain.RequestOwnerCustomer,
		CreatedBy:                    mockUserID,
		Metadata: map[string]interface{}{
			"policy_number":     policy.PolicyNumber,
			"policyholder_name": policy.PolicyholderName,
			"product_name":      policy.ProductName,
			"product_code":      policy.ProductCode,
		},
	}

	// Create surrender request in database
	created, err := h.surrenderRepo.Create(sctx.Ctx, surrenderRequest)
	if err != nil {
		log.Error(sctx.Ctx, "Failed to create surrender request: %v", err)
		return nil, fmt.Errorf("failed to create surrender request")
	}

	log.Info(sctx.Ctx, "Created surrender request %s for policy %s", created.RequestNumber, policy.PolicyNumber)

	// Get document requirements
	docRequirements := h.getDocumentRequirements(policy, calculation.Data.CalculationBreakdown.Deductions.TotalLoan > 0)

	r := &response.ConfirmSurrenderResponse{
		StatusCodeAndMessage: port.CreateSuccess,
		Data: response.ConfirmSurrenderData{
			SurrenderRequestID: created.ID.String(),
			RequestNumber:      created.RequestNumber,
			PolicyID:           policy.ID,
			PolicyNumber:       policy.PolicyNumber,
			Status:             string(created.Status),
			RequestDate:        created.RequestDate.Format(time.RFC3339),
			NetSurrenderValue:  created.NetSurrenderValue,
			DisbursementMethod: string(created.DisbursementMethod),
			DocumentRequirements: response.DocumentRequirementsData{
				Required:             docRequirements,
				TotalRequired:        len(docRequirements),
				TotalUploaded:        0,
				AllDocumentsUploaded: false,
			},
			WorkflowState: response.WorkflowStateData{
				CurrentStage:    "DOCUMENT_UPLOAD",
				CompletedStages: []string{"ELIGIBILITY_CHECK", "CALCULATION", "REQUEST_CREATED"},
				PendingStages:   []string{"DOCUMENT_UPLOAD", "VERIFICATION", "APPROVAL", "PAYMENT"},
				ProgressPercent: 25,
			},
			NextAction: response.NextActionData{
				Action:      "UPLOAD_DOCUMENTS",
				Description: "Please upload all required documents to proceed",
				URL:         "/v1/surrender/documents/upload",
			},
		},
	}

	return r, nil
}

// UploadSurrenderDocument handles document upload for surrender request
// POST /v1/surrender/documents/upload
// Validation Rules: VR-SUR-007, VR-SUR-008, VR-SUR-009
func (h *VoluntarySurrenderHandler) UploadSurrenderDocument(sctx *serverRoute.Context, req UploadDocumentRequest) (*response.UploadDocumentResponse, error) {
	surrenderRequestID, err := uuid.Parse(req.SurrenderRequestID)
	if err != nil {
		log.Error(sctx.Ctx, "Invalid surrender request ID: %v", err)
		return nil, fmt.Errorf("invalid surrender request ID format")
	}

	// Verify surrender request exists
	surrenderRequest, err := h.surrenderRepo.FindByID(sctx.Ctx, surrenderRequestID)
	if err != nil {
		if err == pgx.ErrNoRows {
			log.Error(sctx.Ctx, "Surrender request not found: %s", req.SurrenderRequestID)
			return nil, fmt.Errorf("surrender request not found")
		}
		log.Error(sctx.Ctx, "Failed to get surrender request: %v", err)
		return nil, err
	}

	// VR-SUR-009: Check if document type already uploaded
	docType := domain.DocumentType(req.DocumentType)
	exists, err := h.documentRepo.CheckDocumentExists(sctx.Ctx, surrenderRequestID, docType)
	if err != nil {
		log.Error(sctx.Ctx, "Failed to check document existence: %v", err)
		return nil, fmt.Errorf("failed to validate document")
	}
	if exists {
		return nil, fmt.Errorf("document of type %s already uploaded", req.DocumentType)
	}

	// VR-SUR-007: Validate file size (max 10MB)
	maxFileSize := int64(10 * 1024 * 1024) // 10MB
	if req.File.Size > maxFileSize {
		return nil, fmt.Errorf("file size exceeds maximum allowed size of 10MB")
	}

	// VR-SUR-008: Validate file type
	allowedTypes := []string{"application/pdf", "image/jpeg", "image/png"}
	fileHeader := req.File.Header.Get("Content-Type")
	isValidType := false
	for _, allowedType := range allowedTypes {
		if fileHeader == allowedType {
			isValidType = true
			break
		}
	}
	if !isValidType {
		return nil, fmt.Errorf("invalid file type. Allowed types: PDF, JPEG, PNG")
	}

	// Open uploaded file
	file, err := req.File.Open()
	if err != nil {
		log.Error(sctx.Ctx, "Failed to open uploaded file: %v", err)
		return nil, fmt.Errorf("failed to process uploaded file")
	}
	defer file.Close()

	// Read file content
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		log.Error(sctx.Ctx, "Failed to read file: %v", err)
		return nil, fmt.Errorf("failed to read file content")
	}

	// Upload to document service (placeholder)
	documentPath, err := h.documentService.UploadDocument(sctx.Ctx, fileBytes, DocumentMetadata{
		PolicyID:           surrenderRequest.PolicyID,
		SurrenderRequestID: surrenderRequestID.String(),
		DocumentType:       req.DocumentType,
		FileName:           req.File.Filename,
	})
	if err != nil {
		log.Error(sctx.Ctx, "Failed to upload document: %v", err)
		return nil, fmt.Errorf("failed to upload document")
	}

	// Save document record in database
	fileSizeInt := int(req.File.Size)
	mimeType := fileHeader
	document := domain.SurrenderDocument{
		SurrenderRequestID: surrenderRequestID,
		DocumentType:       docType,
		DocumentName:       req.File.Filename,
		DocumentPath:       documentPath,
		FileSizeBytes:      &fileSizeInt,
		MimeType:           &mimeType,
		Verified:           false,
		Metadata:           map[string]interface{}{},
	}

	created, err := h.documentRepo.Create(sctx.Ctx, document)
	if err != nil {
		log.Error(sctx.Ctx, "Failed to save document record: %v", err)
		return nil, fmt.Errorf("failed to save document")
	}

	log.Info(sctx.Ctx, "Uploaded document %s for surrender request %s", created.DocumentName, surrenderRequest.RequestNumber)

	r := &response.UploadDocumentResponse{
		StatusCodeAndMessage: port.CreateSuccess,
		Data: response.DocumentUploadData{
			DocumentID:         created.ID.String(),
			SurrenderRequestID: created.SurrenderRequestID.String(),
			DocumentType:       string(created.DocumentType),
			DocumentName:       created.DocumentName,
			FileSizeBytes:      fileSizeInt,
			UploadedDate:       created.UploadedDate.Format(time.RFC3339),
			Verified:           created.Verified,
			UploadStatus:       "SUCCESS",
		},
	}

	return r, nil
}

// GetDocumentUploadStatus retrieves document upload status
// GET /v1/surrender/documents/status
func (h *VoluntarySurrenderHandler) GetDocumentUploadStatus(sctx *serverRoute.Context, req DocumentStatusParams) (*response.DocumentStatusResponse, error) {
	surrenderRequestID, err := uuid.Parse(req.SurrenderRequestID)
	if err != nil {
		log.Error(sctx.Ctx, "Invalid surrender request ID: %v", err)
		return nil, fmt.Errorf("invalid surrender request ID format")
	}

	// Get surrender request
	surrenderRequest, err := h.surrenderRepo.FindByID(sctx.Ctx, surrenderRequestID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("surrender request not found")
		}
		log.Error(sctx.Ctx, "Failed to get surrender request: %v", err)
		return nil, err
	}

	// Get all uploaded documents
	documents, err := h.documentRepo.FindBySurrenderRequestID(sctx.Ctx, surrenderRequestID)
	if err != nil {
		log.Error(sctx.Ctx, "Failed to get documents: %v", err)
		return nil, err
	}

	// Count verified documents
	verifiedCount, err := h.documentRepo.CountVerifiedDocuments(sctx.Ctx, surrenderRequestID)
	if err != nil {
		log.Error(sctx.Ctx, "Failed to count verified documents: %v", err)
		verifiedCount = 0
	}

	// Get policy to determine required documents
	policy, err := h.policyService.GetPolicyByID(sctx.Ctx, surrenderRequest.PolicyID)
	if err != nil {
		log.Error(sctx.Ctx, "Failed to get policy: %v", err)
		return nil, fmt.Errorf("failed to get policy details")
	}

	hasLoan := surrenderRequest.LoanDeduction > 0
	requiredDocs := h.getDocumentRequirements(policy, hasLoan)

	documentsData := response.NewDocumentsResponse(documents)
	allUploaded := len(documents) >= len(requiredDocs)
	allVerified := int(verifiedCount) >= len(requiredDocs)
	canSubmit := allUploaded && allVerified

	r := &response.DocumentStatusResponse{
		StatusCodeAndMessage: port.GetSuccess,
		Data: response.DocumentStatusData{
			SurrenderRequestID:   surrenderRequestID.String(),
			TotalRequired:        len(requiredDocs),
			TotalUploaded:        len(documents),
			TotalVerified:        int(verifiedCount),
			AllDocumentsUploaded: allUploaded,
			AllDocumentsVerified: allVerified,
			CanSubmit:            canSubmit,
			Documents:            documentsData,
		},
	}

	return r, nil
}

// SubmitForVerification submits surrender request for CPC verification
// POST /v1/surrender/submit-for-verification
// Business Rule: BR-SUR-017
// Validation Rule: VR-SUR-010
func (h *VoluntarySurrenderHandler) SubmitForVerification(sctx *serverRoute.Context, req SubmitForVerificationRequest) (*response.SubmitForVerificationResponse, error) {
	surrenderRequestID, err := uuid.Parse(req.SurrenderRequestID)
	if err != nil {
		log.Error(sctx.Ctx, "Invalid surrender request ID: %v", err)
		return nil, fmt.Errorf("invalid surrender request ID format")
	}

	// Get surrender request
	surrenderRequest, err := h.surrenderRepo.FindByID(sctx.Ctx, surrenderRequestID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("surrender request not found")
		}
		log.Error(sctx.Ctx, "Failed to get surrender request: %v", err)
		return nil, err
	}

	// VR-SUR-010: Validate all required documents are uploaded and verified
	policy, err := h.policyService.GetPolicyByID(sctx.Ctx, surrenderRequest.PolicyID)
	if err != nil {
		log.Error(sctx.Ctx, "Failed to get policy: %v", err)
		return nil, fmt.Errorf("failed to get policy details")
	}

	hasLoan := surrenderRequest.LoanDeduction > 0
	requiredDocs := h.getDocumentRequirements(policy, hasLoan)

	documents, err := h.documentRepo.FindBySurrenderRequestID(sctx.Ctx, surrenderRequestID)
	if err != nil {
		log.Error(sctx.Ctx, "Failed to get documents: %v", err)
		return nil, err
	}

	if len(documents) < len(requiredDocs) {
		return nil, fmt.Errorf("not all required documents uploaded. Required: %d, Uploaded: %d", len(requiredDocs), len(documents))
	}

	// Check all documents are verified
	verifiedCount, err := h.documentRepo.CountVerifiedDocuments(sctx.Ctx, surrenderRequestID)
	if err != nil {
		log.Error(sctx.Ctx, "Failed to count verified documents: %v", err)
		return nil, err
	}

	if int(verifiedCount) < len(requiredDocs) {
		return nil, fmt.Errorf("not all documents verified. Required: %d, Verified: %d", len(requiredDocs), verifiedCount)
	}

	// BR-SUR-017: Update status to PENDING_VERIFICATION
	oldStatus := surrenderRequest.Status
	mockUserID := uuid.New()
	updated, err := h.surrenderRepo.UpdateStatus(sctx.Ctx, surrenderRequestID, domain.SurrenderStatusPendingVerification, mockUserID, nil)
	if err != nil {
		log.Error(sctx.Ctx, "Failed to update status: %v", err)
		return nil, fmt.Errorf("failed to submit for verification")
	}

	log.Info(sctx.Ctx, "Submitted surrender request %s for verification", updated.RequestNumber)

	r := &response.SubmitForVerificationResponse{
		StatusCodeAndMessage: port.UpdateSuccess,
		Data: response.SubmitVerificationData{
			SurrenderRequestID: updated.ID.String(),
			RequestNumber:      updated.RequestNumber,
			OldStatus:          string(oldStatus),
			NewStatus:          string(updated.Status),
			SubmittedAt:        time.Now().Format(time.RFC3339),
			WorkflowState: response.WorkflowStateData{
				CurrentStage:    "VERIFICATION",
				CompletedStages: []string{"ELIGIBILITY_CHECK", "CALCULATION", "REQUEST_CREATED", "DOCUMENT_UPLOAD"},
				PendingStages:   []string{"VERIFICATION", "APPROVAL", "PAYMENT"},
				ProgressPercent: 50,
			},
			NextAction: response.NextActionData{
				Action:      "AWAIT_VERIFICATION",
				Description: "Your request is under verification by CPC staff",
				URL:         "/v1/surrender/status?surrender_request_id=" + updated.ID.String(),
			},
		},
	}

	return r, nil
}

// GetSurrenderStatus retrieves surrender request status
// GET /v1/surrender/status
func (h *VoluntarySurrenderHandler) GetSurrenderStatus(sctx *serverRoute.Context, req SurrenderStatusParams) (*response.SurrenderStatusResponse, error) {
	surrenderRequestID, err := uuid.Parse(req.SurrenderRequestID)
	if err != nil {
		log.Error(sctx.Ctx, "Invalid surrender request ID: %v", err)
		return nil, fmt.Errorf("invalid surrender request ID format")
	}

	// Get surrender request
	surrenderRequest, err := h.surrenderRepo.FindByID(sctx.Ctx, surrenderRequestID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("surrender request not found")
		}
		log.Error(sctx.Ctx, "Failed to get surrender request: %v", err)
		return nil, err
	}

	// Build workflow state
	workflowState := h.buildWorkflowState(surrenderRequest.Status)

	// Basic response
	data := response.SurrenderStatusData{
		SurrenderRequestID: surrenderRequest.ID.String(),
		RequestNumber:      surrenderRequest.RequestNumber,
		PolicyID:           surrenderRequest.PolicyID,
		PolicyNumber:       surrenderRequest.Metadata["policy_number"].(string),
		RequestType:        string(surrenderRequest.RequestType),
		Status:             string(surrenderRequest.Status),
		RequestDate:        surrenderRequest.RequestDate.Format(time.RFC3339),
		NetSurrenderValue:  surrenderRequest.NetSurrenderValue,
		WorkflowState:      workflowState,
	}

	// If detailed view requested, add more information
	if req.IncludeDetails {
		// Add detailed information here (history, documents, etc.)
		// This would be populated with actual data in production
		data.Details = &response.SurrenderDetailsData{
			PolicyDetails: response.PolicyDetailsData{
				PolicyID:     surrenderRequest.PolicyID,
				PolicyNumber: surrenderRequest.Metadata["policy_number"].(string),
				ProductCode:  surrenderRequest.Metadata["product_code"].(string),
				ProductName:  surrenderRequest.Metadata["product_name"].(string),
			},
		}
	}

	r := &response.SurrenderStatusResponse{
		StatusCodeAndMessage: port.GetSuccess,
		Data:                 data,
	}

	return r, nil
}

// Helper functions

func (h *VoluntarySurrenderHandler) createIneligibleResponse(policy *PolicyInfo, reasons []string) *response.EligibilityIneligibleResponse {
	return &response.EligibilityIneligibleResponse{
		StatusCodeAndMessage: port.StatusCodeAndMessage{
			StatusCode: 200,
			Success:    false,
			Message:    "Policy is not eligible for surrender",
		},
		Data: response.EligibilityIneligibleData{
			Eligible:     false,
			PolicyID:     policy.ID,
			PolicyNumber: policy.PolicyNumber,
			Reasons:      reasons,
			Message:      "Your policy does not meet the eligibility criteria for surrender.",
		},
	}
}

func (h *VoluntarySurrenderHandler) createIneligibleResponse1(policyNumber string, reasons []string) *response.EligibilityIneligibleResponse {
	return &response.EligibilityIneligibleResponse{
		StatusCodeAndMessage: port.StatusCodeAndMessage{
			StatusCode: 200,
			Success:    false,
			Message:    "Policy is not eligible for surrender",
		},
		Data: response.EligibilityIneligibleData{
			Eligible:     false,
			PolicyID:     "",
			PolicyNumber: policyNumber,
			Reasons:      reasons,
			Message:      "Your policy does not meet the eligibility criteria for surrender.",
		},
	}
}

func (h *VoluntarySurrenderHandler) generateSurReqIDResponse(surservReqID string, reasons []string) *response.IndexSurrenderResponse {
	return &response.IndexSurrenderResponse{
		StatusCodeAndMessage: port.StatusCodeAndMessage{
			StatusCode: 200,
			Success:    true,
			Message:    "Surrender Service Request ID generated successfully",
		},
		Data: response.IndexSurrenderResponseData{
			ServiceRequestID: surservReqID,
		},
	}
}

func (h *VoluntarySurrenderHandler) getMinimumPremiumsByProduct(productCode string) int {
	// BR-SUR-001: Minimum premium payment periods by product
	minimums := map[string]int{
		"WLA":   4, // Whole Life Assurance
		"EA":    3, // Endowment Assurance
		"CWLA":  4, // Convertible Whole Life
		"CHILD": 5, // Child Policy
		"JLA":   3, // Joint Life Assurance
	}
	if min, ok := minimums[productCode]; ok {
		return min
	}
	return 3 // Default
}

func (h *VoluntarySurrenderHandler) calculateBonusDetails(policy *PolicyInfo) []response.BonusDetailData {
	// BR-SUR-005: Calculate bonus for each financial year
	// This is a placeholder - actual implementation would use bonus tables
	bonusDetails := []response.BonusDetailData{}

	// Mock bonus calculation
	for year := 0; year < policy.PremiumsPaid; year++ {
		bonusRate := 50.0 // ₹50 per thousand
		bonusAmount := (policy.SumAssured / 1000) * bonusRate

		bonusDetails = append(bonusDetails, response.BonusDetailData{
			FinancialYear: fmt.Sprintf("%d-%d", 2020+year, 2021+year),
			SumAssured:    policy.SumAssured,
			BonusRate:     bonusRate,
			BonusAmount:   bonusAmount,
		})
	}

	return bonusDetails
}

func (h *VoluntarySurrenderHandler) getSurrenderFactor(productCode string, term int, age int) float64 {
	// BR-SUR-007: Get surrender factor from actuarial tables
	// This is a placeholder - actual implementation would use surrender factor tables
	// Factors typically range from 0.30 to 0.90
	return 0.75
}

func (h *VoluntarySurrenderHandler) getPrescribedLimit(productCode string) float64 {
	// BR-SUR-011: Get prescribed limit from configuration
	// This is a placeholder
	return 2000.0
}

func (h *VoluntarySurrenderHandler) predictDisposition(netAmount float64, prescribedLimit float64, paidUpValue float64) response.DispositionPredictionData {
	// BR-SUR-011, BR-SUR-012: Predict disposition
	if netAmount >= prescribedLimit {
		return response.DispositionPredictionData{
			PredictedDisposition: "REDUCED_PAID_UP",
			PrescribedLimit:      prescribedLimit,
			NetAmount:            netAmount,
			WillCreateReducedPU:  true,
			NewSumAssured:        paidUpValue,
			NewPolicyStatus:      "AU",
		}
	}

	return response.DispositionPredictionData{
		PredictedDisposition: "TERMINATED_SURRENDER",
		PrescribedLimit:      prescribedLimit,
		NetAmount:            netAmount,
		WillCreateReducedPU:  false,
		NewPolicyStatus:      "TS",
	}
}

func (h *VoluntarySurrenderHandler) generateRequestNumber(policyNumber string) string {
	// BR-SUR-013: Generate unique request number
	timestamp := time.Now().Format("20060102150405")
	return fmt.Sprintf("SUR-%s-%s", policyNumber, timestamp)
}

func (h *VoluntarySurrenderHandler) getDocumentRequirements(policy *PolicyInfo, hasLoan bool) []response.DocumentRequirementData {
	// BR-SUR-015: Document requirements
	requirements := []response.DocumentRequirementData{
		{
			DocumentType: "WRITTEN_CONSENT",
			DisplayName:  "Written Consent",
			Mandatory:    true,
			Uploaded:     false,
			Description:  "Written consent for policy surrender",
		},
		{
			DocumentType: "POLICY_BOND",
			DisplayName:  "Policy Bond",
			Mandatory:    true,
			Uploaded:     false,
			Description:  "Original policy bond",
		},
		{
			DocumentType: "PREMIUM_RECEIPT_BOOK",
			DisplayName:  "Premium Receipt Book",
			Mandatory:    true,
			Uploaded:     false,
			Description:  "Premium receipt book",
		},
	}

	if hasLoan {
		requirements = append(requirements, response.DocumentRequirementData{
			DocumentType: "LOAN_BOND",
			DisplayName:  "Loan Bond",
			Mandatory:    true,
			Uploaded:     false,
			Description:  "Loan bond document",
		})
	}

	if policy.IsAssigned {
		requirements = append(requirements, response.DocumentRequirementData{
			DocumentType: "ASSIGNMENT_DEED",
			DisplayName:  "Assignment Deed",
			Mandatory:    true,
			Uploaded:     false,
			Description:  "Assignment deed document",
		})
	}

	return requirements
}

func (h *VoluntarySurrenderHandler) buildWorkflowState(status domain.SurrenderStatus) response.WorkflowStateData {
	stages := map[domain.SurrenderStatus]response.WorkflowStateData{
		domain.SurrenderStatusPendingDocumentUpload: {
			CurrentStage:    "DOCUMENT_UPLOAD",
			CompletedStages: []string{"REQUEST_CREATED"},
			PendingStages:   []string{"DOCUMENT_UPLOAD", "VERIFICATION", "APPROVAL", "PAYMENT"},
			ProgressPercent: 20,
		},
		domain.SurrenderStatusPendingVerification: {
			CurrentStage:    "VERIFICATION",
			CompletedStages: []string{"REQUEST_CREATED", "DOCUMENT_UPLOAD"},
			PendingStages:   []string{"VERIFICATION", "APPROVAL", "PAYMENT"},
			ProgressPercent: 40,
		},
		domain.SurrenderStatusPendingApproval: {
			CurrentStage:    "APPROVAL",
			CompletedStages: []string{"REQUEST_CREATED", "DOCUMENT_UPLOAD", "VERIFICATION"},
			PendingStages:   []string{"APPROVAL", "PAYMENT"},
			ProgressPercent: 60,
		},
		domain.SurrenderStatusApproved: {
			CurrentStage:    "PAYMENT",
			CompletedStages: []string{"REQUEST_CREATED", "DOCUMENT_UPLOAD", "VERIFICATION", "APPROVAL"},
			PendingStages:   []string{"PAYMENT"},
			ProgressPercent: 80,
		},
		domain.SurrenderStatusTerminated: {
			CurrentStage:    "COMPLETED",
			CompletedStages: []string{"REQUEST_CREATED", "DOCUMENT_UPLOAD", "VERIFICATION", "APPROVAL", "PAYMENT"},
			PendingStages:   []string{},
			ProgressPercent: 100,
		},
	}

	if state, ok := stages[status]; ok {
		return state
	}

	return response.WorkflowStateData{
		CurrentStage:    string(status),
		CompletedStages: []string{},
		PendingStages:   []string{},
		ProgressPercent: 0,
	}
}

// ============================================
// External Service Interfaces (Placeholders)
// ============================================

type PolicyServiceInterface interface {
	GetPolicyByID(ctx interface{}, policyID string) (*PolicyInfo, error)
	UpdatePolicyStatus(ctx interface{}, policyID string, status string) error
}

type LoanServiceInterface interface {
	GetLoanDetails(ctx interface{}, policyID string) (*LoanDetails, error)
}

type CollectionsServiceInterface interface {
	GetUnpaidPremiums(ctx interface{}, policyID string) (float64, error)
	GetPoliciesWithUnpaidPremiums(ctx interface{}, cutoffDate time.Time, minUnpaidMonths int) ([]string, error)
}

type DocumentServiceInterface interface {
	UploadDocument(ctx interface{}, file []byte, metadata DocumentMetadata) (string, error)
}

type PolicyInfo struct {
	ID               string
	PolicyNumber     string
	ProductCode      string
	ProductName      string
	Status           string
	SumAssured       float64
	PremiumsPaid     int
	TotalPremiums    int
	MaturityDate     time.Time
	CommencementDate time.Time
	PolicyholderName string
	Address          string
	IsAssigned       bool
	AssigneeName     string
	Term             int
	AgeLast          int
}

type LoanDetails struct {
	Principal float64
	Interest  float64
}

type DocumentMetadata struct {
	PolicyID           string
	SurrenderRequestID string
	DocumentType       string
	FileName           string
}

// Mock implementations (replace with actual services)

func NewMockPolicyService() PolicyServiceInterface {
	return &MockPolicyService{}
}

type MockPolicyService struct{}

func (m *MockPolicyService) GetPolicyByID(ctx interface{}, policyID string) (*PolicyInfo, error) {
	return &PolicyInfo{
		ID:               policyID,
		PolicyNumber:     "PLI/2020/123456111",
		ProductCode:      "EA",
		ProductName:      "Santosh - Endowment Assurance",
		Status:           "AP",
		SumAssured:       100000,
		PremiumsPaid:     5,
		TotalPremiums:    20,
		MaturityDate:     time.Now().AddDate(15, 0, 0),
		CommencementDate: time.Now().AddDate(-5, 0, 0),
		PolicyholderName: "John Doe",
		Address:          "123 Main Street, City",
		IsAssigned:       false,
		Term:             20,
		AgeLast:          35,
	}, nil
}

func (m *MockPolicyService) UpdatePolicyStatus(ctx interface{}, policyID string, status string) error {
	return nil
}

func NewMockLoanService() LoanServiceInterface {
	return &MockLoanService{}
}

type MockLoanService struct{}

func (m *MockLoanService) GetLoanDetails(ctx interface{}, policyID string) (*LoanDetails, error) {
	return &LoanDetails{
		Principal: 5000,
		Interest:  500,
	}, nil
}

func NewMockCollectionsService() CollectionsServiceInterface {
	return &MockCollectionsService{}
}

type MockCollectionsService struct{}

func (m *MockCollectionsService) GetUnpaidPremiums(ctx interface{}, policyID string) (float64, error) {
	// Return 6000 unpaid premiums for the test policy (6 months * 1000)
	if policyID == "33920d68-a5e9-4e7e-8335-131c8347e04d" {
		return 6000, nil
	}
	return 0, nil
}

func (m *MockCollectionsService) GetPoliciesWithUnpaidPremiums(ctx interface{}, cutoffDate time.Time, minUnpaidMonths int) ([]string, error) {
	// Return the test policy as having unpaid premiums for forced surrender evaluation
	return []string{"33920d68-a5e9-4e7e-8335-131c8347e04d"}, nil
}

func NewMockDocumentService() DocumentServiceInterface {
	return &MockDocumentService{}
}

type MockDocumentService struct{}

func (m *MockDocumentService) UploadDocument(ctx interface{}, file []byte, metadata DocumentMetadata) (string, error) {
	// In production, this would upload to actual document storage
	filename := filepath.Join("/documents", "surrender", metadata.SurrenderRequestID, metadata.FileName)

	// Create directory structure
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}

	// Write file (in production, use proper storage service)
	if err := os.WriteFile(filename, file, 0644); err != nil {
		return "", err
	}

	return filename, nil
}
