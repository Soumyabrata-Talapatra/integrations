package bootstrap

import (
	"context"
	// ← add this
	"go.temporal.io/sdk/activity"
	temporalclient "go.temporal.io/sdk/client"
	temporalworker "go.temporal.io/sdk/worker"
	"go.uber.org/fx"

	config "gitlab.cept.gov.in/it-2.0-common/api-config"
	serverHandler "gitlab.cept.gov.in/it-2.0-common/n-api-server/handler"

	handler "customer-nfs-service/handler"
	repo "customer-nfs-service/repo/postgres"
	"customer-nfs-service/temporal/activities"
	"customer-nfs-service/temporal/workflows"
)

// FxRepo provides all repository implementations via Uber FX.
var FxRepo = fx.Module(
	"Repomodule",
	fx.Provide(
		// Phase 1 — Address Change
		repo.NewServiceRequestRepository,
		repo.NewAddressChangeRepository,
		repo.NewDocumentUploadRepository,
		repo.NewAuditLogRepository,
		// Phase 2 — Name Change
		repo.NewNameChangeRepository,
	),
)

// FxHandler provides all HTTP handler implementations via Uber FX.
var FxHandler = fx.Module(
	"Handlermodule",
	fx.Provide(
		// Phase 1 — Address Change (CORE-001..004)
		fx.Annotate(
			handler.NewAddressChangeHandler,
			fx.As(new(serverHandler.Handler)),
			fx.ResultTags(serverHandler.ServerControllersGroupTag),
		),
		// Phase 2 — Name Change (CORE-005..010)
		fx.Annotate(
			handler.NewNameChangeHandler,
			fx.As(new(serverHandler.Handler)),
			fx.ResultTags(serverHandler.ServerControllersGroupTag),
		),
		// Phase 3 — Lookup + Validation (LU-001..008, VA-001..003)
		fx.Annotate(
			handler.NewLookupHandler,
			fx.As(new(serverHandler.Handler)),
			fx.ResultTags(serverHandler.ServerControllersGroupTag),
		),
		// Phase 3 — Status (ST-001..005)
		fx.Annotate(
			handler.NewStatusHandler,
			fx.As(new(serverHandler.Handler)),
			fx.ResultTags(serverHandler.ServerControllersGroupTag),
		),
		// Phase 3 — Document Management (DM-001..005)
		fx.Annotate(
			handler.NewDocumentHandler,
			fx.As(new(serverHandler.Handler)),
			fx.ResultTags(serverHandler.ServerControllersGroupTag),
		),
		// Phase 3 — CPC Operations (CPC-001..006)
		fx.Annotate(
			handler.NewCPCHandler,
			fx.As(new(serverHandler.Handler)),
			fx.ResultTags(serverHandler.ServerControllersGroupTag),
		),
	),
)

// FxTemporal provides the Temporal workflow client configured from application config.
var FxTemporal = fx.Module(
	"Temporalmodule",
	fx.Provide(
		NewTemporalClient,
	),
)

// NewTemporalClient constructs a Temporal client from application configuration.
func NewTemporalClient(cfg *config.Config) (temporalclient.Client, error) {
	host := cfg.GetString("temporal.host")
	if host == "" {
		host = "localhost"
	}
	port := cfg.GetString("temporal.port")
	if port == "" {
		port = "7233"
	}
	hostPort := host + ":" + port
	namespace := cfg.GetString("temporal.namespace")
	if namespace == "" {
		namespace = "default"
	}

	tc, err := temporalclient.Dial(temporalclient.Options{
		HostPort:  hostPort,
		Namespace: namespace,
	})
	if err != nil {
		return nil, err
	}

	return tc, nil
}

// FxWorker provides the Temporal worker that polls customer-nfs-tq and executes
// workflow and activity implementations. Started via fx.Lifecycle OnStart/OnStop.
var FxWorker = fx.Module(
	"Workermodule",
	fx.Provide(NewTemporalWorker),
	fx.Invoke(StartTemporalWorker),
)

// NewTemporalWorker constructs the Temporal worker with all workflows and activities registered.
func NewTemporalWorker(
	tc temporalclient.Client,
	srRepo *repo.ServiceRequestRepository,
	nameRepo *repo.NameChangeRepository,
	addrRepo *repo.AddressChangeRepository, // ← add

	auditRepo *repo.AuditLogRepository,
	cfg *config.Config,
) temporalworker.Worker {
	w := temporalworker.New(tc, workflows.TaskQueue, temporalworker.Options{})

	// Register workflows
	w.RegisterWorkflow(workflows.AadhaarNameChangeWorkflow)
	w.RegisterWorkflow(workflows.ManualNameChangeWorkflow)
	w.RegisterWorkflow(workflows.WithdrawalWorkflow)
	w.RegisterWorkflow(workflows.AadhaarAddressChangeWorkflow)
	w.RegisterWorkflow(workflows.ManualAddressChangeWorkflow)

	// Register activities
	nameAct := activities.NewNameChangeActivities(srRepo, nameRepo, auditRepo, *cfg, tc)
	w.RegisterActivity(nameAct)
	addrAct := activities.NewAddressChangeActivities(srRepo, addrRepo, auditRepo, cfg, &tc)
	w.RegisterActivityWithOptions(addrAct.StoreWorkflowState, activity.RegisterOptions{Name: "StoreWorkflowState"})
	w.RegisterActivityWithOptions(addrAct.ValidateAddressRequest, activity.RegisterOptions{Name: "ValidateAddressRequest"})
	w.RegisterActivityWithOptions(addrAct.CreateAddressServiceRequest, activity.RegisterOptions{Name: "CreateAddressServiceRequest"})
	w.RegisterActivityWithOptions(addrAct.RequestAadhaarOTP, activity.RegisterOptions{Name: "RequestAadhaarOTP"})
	w.RegisterActivityWithOptions(addrAct.VerifyAadhaarOTP, activity.RegisterOptions{Name: "VerifyAadhaarOTP"})
	w.RegisterActivityWithOptions(addrAct.UpdateAddressData, activity.RegisterOptions{Name: "UpdateAddressData"})
	w.RegisterActivityWithOptions(addrAct.AssignToCPC, activity.RegisterOptions{Name: "AssignToCPC"})
	// w.RegisterActivityWithOptions(addrAct.UpdateStatus, activity.RegisterOptions{Name: "UpdateStatus"})
	w.RegisterActivityWithOptions(addrAct.UpdateStatus, activity.RegisterOptions{Name: "AddressUpdateStatus"})
	w.RegisterActivityWithOptions(addrAct.GenerateAckReceipt, activity.RegisterOptions{Name: "GenerateAckReceipt"})
	w.RegisterActivityWithOptions(addrAct.Escalate, activity.RegisterOptions{Name: "Escalate"})
	withdrawAct := activities.NewWithdrawalActivities(srRepo, auditRepo, *cfg)
	w.RegisterActivityWithOptions(withdrawAct.CheckWithdrawalEligibility, activity.RegisterOptions{Name: "CheckWithdrawalEligibility"})
	w.RegisterActivityWithOptions(withdrawAct.ProcessWithdrawal, activity.RegisterOptions{Name: "ProcessWithdrawal"})
	w.RegisterActivityWithOptions(withdrawAct.UpdateStatus, activity.RegisterOptions{Name: "WithdrawalUpdateStatus"})
	return w
}

// StartTemporalWorker wires the worker into the FX lifecycle so it starts and
// stops cleanly with the application.
func StartTemporalWorker(lc fx.Lifecycle, w temporalworker.Worker) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return w.Start()
		},
		OnStop: func(ctx context.Context) error {
			w.Stop()
			return nil
		},
	})
}
