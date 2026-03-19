package handler

import (
	"errors"
	"net/http"
	"time"

	"policy-issue-service/core/port"
	resp "policy-issue-service/handler/response"
	repo "policy-issue-service/repo/postgres"

	"github.com/google/uuid"
	config "gitlab.cept.gov.in/it-2.0-common/api-config"
	apierrors "gitlab.cept.gov.in/it-2.0-common/n-api-errors"
	log "gitlab.cept.gov.in/it-2.0-common/n-api-log"
	serverHandler "gitlab.cept.gov.in/it-2.0-common/n-api-server/handler"
	serverRoute "gitlab.cept.gov.in/it-2.0-common/n-api-server/route"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
)

// CustomerHandler handles customer-related HTTP endpoints
type CustomerHandler struct {
	*serverHandler.Base
	proposalRepo   *repo.ProposalRepository
	productRepo    *repo.ProductRepository
	cfg            *config.Config
	temporalClient client.Client
}

// NewCustomerHandler creates a new CustomerHandler instance
func NewCustomerHandler(proposalRepo *repo.ProposalRepository, productRepo *repo.ProductRepository, cfg *config.Config, temporalClient client.Client) *CustomerHandler {
	base := serverHandler.New("Customers").SetPrefix("/v1").AddPrefix("/customer")
	return &CustomerHandler{Base: base, proposalRepo: proposalRepo, productRepo: productRepo, cfg: cfg, temporalClient: temporalClient}
}

// Routes returns the routes for the CustomerHandler
func (h *CustomerHandler) Routes() []serverRoute.Route {
	return []serverRoute.Route{
		serverRoute.POST("/get", h.GetCustomer).Name("Get Customer"),
		serverRoute.POST("/create", h.CreateCustomer).Name("Create Customer"),
		serverRoute.POST("/deduplication-check", h.DeduplicateCustomer).Name("Deduplication of Customer"),
		serverRoute.POST("/address", h.CustomerAddress).Name("Customer Address"),
		serverRoute.POST("/contact", h.CustomerContact).Name("Customer Contact"),
	}
}

func HandleWorkflowError(sctx *serverRoute.Context, err error) error {

	var appErr *temporal.ApplicationError

	if errors.As(err, &appErr) {
		return apierrors.HandleErrorWithStatusCodeAndMessage(
			apierrors.HTTPErrorBadRequest,
			appErr.Error(),
			nil,
		)
	}

	return apierrors.HandleErrorWithStatusCodeAndMessage(
		apierrors.HTTPErrorServerError,
		"Workflow execution failed",
		err,
	)
}

func (h *CustomerHandler) DeduplicateCustomer(sctx *serverRoute.Context, input DedupInput) (*resp.CustomerDedupResponse, error) {

	workflowOptions := client.StartWorkflowOptions{
		ID:                       "customer-dedup-" + uuid.NewString(),
		TaskQueue:                "customer-tq",
		WorkflowExecutionTimeout: time.Minute,
	}

	we, err := h.temporalClient.ExecuteWorkflow(
		sctx.Ctx,
		workflowOptions,
		"CustomerDeduplicationWorkflow",
		input,
	)
	if err != nil {
		log.Error(sctx.Ctx, "Failed to start dedupe workflow: %v", err)

		return nil, apierrors.HandleErrorWithStatusCodeAndMessage(
			apierrors.HTTPErrorServerError,
			"Failed to start customer deduplication",
			err,
		)

	}

	var result resp.DedupOutput
	err = we.Get(sctx.Ctx, &result)
	if err != nil {
		return nil, HandleWorkflowError(sctx, err)
	}
	return &resp.CustomerDedupResponse{
		StatusCodeAndMessage: port.StatusCodeAndMessage{
			StatusCode: http.StatusOK,
			Message:    "Customer retrieved successfully",
		},
		DedupOutput: result,
	}, nil
}

func (h *CustomerHandler) GetCustomer(sctx *serverRoute.Context,
	input CustomerGetInput) (*resp.CustomerDetailResponse, error) {

	workflowOptions := client.StartWorkflowOptions{
		ID:                       "customer-get-" + uuid.NewString(),
		TaskQueue:                "customer-tq",
		WorkflowExecutionTimeout: time.Minute,
	}

	we, err := h.temporalClient.ExecuteWorkflow(sctx.Ctx,
		workflowOptions, "CustomerGetWorkflow", input)
	if err != nil {
		log.Error(sctx.Ctx, "Failed to start workflow: %v", err)

		return nil, apierrors.HandleErrorWithStatusCodeAndMessage(
			apierrors.HTTPErrorServerError,
			"Failed to start get customer workflow",
			err,
		)

	}
	var result resp.CustomerGetOutput
	err = we.Get(sctx.Ctx, &result)
	if err != nil {
		return nil, HandleWorkflowError(sctx, err)
	}
	return &resp.CustomerDetailResponse{
		StatusCodeAndMessage: port.StatusCodeAndMessage{
			StatusCode: http.StatusOK,
			Message:    "Customer retrieved successfully",
		},
		CustomerGetOutput: result,
	}, nil
}

func (h *CustomerHandler) CreateCustomer(sctx *serverRoute.Context,
	input CustomerCreateInput) (*resp.CustomerCreateOutput, error) {

	workflowOptions := client.StartWorkflowOptions{
		ID:                       "customer-create-" + input.IdempotencyKey,
		TaskQueue:                "customer-tq",
		WorkflowExecutionTimeout: time.Minute,
	}

	we, err := h.temporalClient.ExecuteWorkflow(sctx.Ctx, workflowOptions,
		"CustomerCreateWorkflowSplit", input)
	if err != nil {
		log.Error(sctx.Ctx, "Failed to start workflow: %v", err)
		return nil, apierrors.HandleErrorWithStatusCodeAndMessage(
			apierrors.HTTPErrorServerError,
			"Failed to start create customer workflow",
			err,
		)
	}

	// Decode into workflow output
	var wfResult resp.CustomerCreateOutput
	err = we.Get(sctx.Ctx, &wfResult)
	if err != nil {
		log.Error(sctx.Ctx, "Workflow execution failed: %v", err)
		return nil, apierrors.HandleErrorWithStatusCodeAndMessage(
			apierrors.HTTPErrorServerError,
			"Create customer workflow failed",
			err,
		)
	}

	// Build HTTP response separately
	response := resp.CustomerCreateOutput{
		CustomerID:    wfResult.CustomerID,
		Status:        wfResult.Status,
		CreatedAt:     wfResult.CreatedAt,
		DedupWarnings: wfResult.DedupWarnings,
	}

	return &response, nil
}

func (h *CustomerHandler) CustomerAddress(sctx *serverRoute.Context,
	input CustomerAddressInput) (*resp.CustomerAddressResponse, error) {

	workflowOptions := client.StartWorkflowOptions{
		ID:                       "customer-address-" + uuid.NewString(),
		TaskQueue:                "customer-tq",
		WorkflowExecutionTimeout: time.Minute,
	}
	if input.Action == "CREATE" && len(input.Address) == 0 {
		return nil, apierrors.HandleErrorWithStatusCodeAndMessage(
			apierrors.HTTPErrorBadRequest,
			"At least one address must be provided",
			nil,
		)
	}
	we, err := h.temporalClient.ExecuteWorkflow(
		sctx.Ctx,
		workflowOptions,
		"CustomerAddressWorkflow",
		input,
	)
	if err != nil {
		log.Error(sctx.Ctx, "Failed to start address workflow: %v", err)
		return nil, apierrors.HandleErrorWithStatusCodeAndMessage(
			apierrors.HTTPErrorServerError,
			"Failed to start address workflow",
			err,
		)
	}

	var result resp.CustomerAddressOutput
	err = we.Get(sctx.Ctx, &result)
	if err != nil {
		return nil, HandleWorkflowError(sctx, err)
	}

	return &resp.CustomerAddressResponse{
		StatusCodeAndMessage: port.StatusCodeAndMessage{
			StatusCode: http.StatusOK,
			Message:    "Customer address processed successfully",
		},
		CustomerAddressOutput: result,
	}, nil
}

func (h *CustomerHandler) CustomerContact(sctx *serverRoute.Context,
	input CustomerContactInput) (*resp.CustomerContactResponse, error) {

	workflowOptions := client.StartWorkflowOptions{
		ID:                       "customer-contact-" + uuid.NewString(),
		TaskQueue:                "customer-tq",
		WorkflowExecutionTimeout: time.Minute,
	}

	// validation
	if input.Action == "CREATE" && len(input.Contacts) == 0 {
		return nil, apierrors.HandleErrorWithStatusCodeAndMessage(
			apierrors.HTTPErrorBadRequest,
			"At least one contact must be provided",
			nil,
		)
	}

	we, err := h.temporalClient.ExecuteWorkflow(sctx.Ctx,
		workflowOptions, "CustomerContactWorkflow", input)

	if err != nil {
		log.Error(sctx.Ctx, "Failed to start contact workflow: %v", err)
		return nil, apierrors.HandleErrorWithStatusCodeAndMessage(
			apierrors.HTTPErrorServerError,
			"Failed to start contact workflow",
			err,
		)
	}

	var result resp.CustomerContactOutput

	err = we.Get(sctx.Ctx, &result)
	if err != nil {
		return nil, HandleWorkflowError(sctx, err)
	}

	return &resp.CustomerContactResponse{
		StatusCodeAndMessage: port.StatusCodeAndMessage{
			StatusCode: http.StatusOK,
			Message:    "Customer contact processed successfully",
		},
		CustomerContactOutput: result,
	}, nil
}

func (h *CustomerHandler) CustomerEmployment(sctx *serverRoute.Context,
	input CustomerEmploymentInput) (*resp.CustomerEmploymentResponse, error) {

	workflowOptions := client.StartWorkflowOptions{
		ID:                       "customer-employment-" + uuid.NewString(),
		TaskQueue:                "customer-tq",
		WorkflowExecutionTimeout: time.Minute,
	}

	// validation
	if input.Action == "CREATE" && input.Employment == nil {
		return nil, apierrors.HandleErrorWithStatusCodeAndMessage(
			apierrors.HTTPErrorBadRequest,
			"employment details required for CREATE",
			nil,
		)
	}

	we, err := h.temporalClient.ExecuteWorkflow(sctx.Ctx,
		workflowOptions, "CustomerEmploymentWorkflow", input)

	if err != nil {
		log.Error(sctx.Ctx, "Failed to start employment workflow: %v", err)
		return nil, apierrors.HandleErrorWithStatusCodeAndMessage(
			apierrors.HTTPErrorServerError,
			"Failed to start employment workflow",
			err,
		)
	}

	var result resp.EmploymentOutput

	err = we.Get(sctx.Ctx, &result)
	if err != nil {
		return nil, HandleWorkflowError(sctx, err)
	}

	return &resp.CustomerEmploymentResponse{
		StatusCodeAndMessage: port.StatusCodeAndMessage{
			StatusCode: http.StatusOK,
			Message:    "Customer employment processed successfully",
		},
		EmploymentOutput: result,
	}, nil
}
