package bootstrap

import (
	"context"
	handler "policy-issue-service/handler"
	repo "policy-issue-service/repo/postgres"
	"policy-issue-service/workflows"
	"policy-issue-service/workflows/activities"

	config "gitlab.cept.gov.in/it-2.0-common/api-config"
	log "gitlab.cept.gov.in/it-2.0-common/n-api-log"
	serverHandler "gitlab.cept.gov.in/it-2.0-common/n-api-server/handler"
	"go.uber.org/fx"

	"go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/worker"
)

// FxRepo module provides all repository implementations
var FxRepo = fx.Module(
	"Repomodule",
	fx.Provide(
		repo.NewQuoteRepository,
		repo.NewProductRepository,
		repo.NewProposalRepository,
		repo.NewAadhaarRepository,
		repo.NewDocumentRepository,
		repo.NewBulkUploadRepository,
	),
)

// FxHandler module provides all HTTP handlers
var FxHandler = fx.Module(
	"Handlermodule",
	fx.Provide(
		fx.Annotate(
			handler.NewQuoteHandler,
			fx.As(new(serverHandler.Handler)),
			fx.ResultTags(serverHandler.ServerControllersGroupTag),
		),
		fx.Annotate(
			handler.NewProposalHandler,
			fx.As(new(serverHandler.Handler)),
			fx.ResultTags(serverHandler.ServerControllersGroupTag),
		),
		fx.Annotate(
			handler.NewAadhaarHandler,
			fx.As(new(serverHandler.Handler)),
			fx.ResultTags(serverHandler.ServerControllersGroupTag),
		),
		fx.Annotate(
			handler.NewApprovalHandler,
			fx.As(new(serverHandler.Handler)),
			fx.ResultTags(serverHandler.ServerControllersGroupTag),
		),
		fx.Annotate(
			handler.NewPolicyHandler,
			fx.As(new(serverHandler.Handler)),
			fx.ResultTags(serverHandler.ServerControllersGroupTag),
		),
		fx.Annotate(
			handler.NewLookupHandler,
			fx.As(new(serverHandler.Handler)),
			fx.ResultTags(serverHandler.ServerControllersGroupTag),
		),
		fx.Annotate(
			handler.NewValidationHandler,
			fx.As(new(serverHandler.Handler)),
			fx.ResultTags(serverHandler.ServerControllersGroupTag),
		),
		fx.Annotate(
			handler.NewCalculationHandler,
			fx.As(new(serverHandler.Handler)),
			fx.ResultTags(serverHandler.ServerControllersGroupTag),
		),
		fx.Annotate(
			handler.NewDocumentHandler,
			fx.As(new(serverHandler.Handler)),
			fx.ResultTags(serverHandler.ServerControllersGroupTag),
		),
		fx.Annotate(
			handler.NewStatusHandler,
			fx.As(new(serverHandler.Handler)),
			fx.ResultTags(serverHandler.ServerControllersGroupTag),
		),
		fx.Annotate(
			handler.NewWorkflowHandler,
			fx.As(new(serverHandler.Handler)),
			fx.ResultTags(serverHandler.ServerControllersGroupTag),
		),
		fx.Annotate(
			handler.NewBulkUploadHandler,
			fx.As(new(serverHandler.Handler)),
			fx.ResultTags(serverHandler.ServerControllersGroupTag),
		),
		fx.Annotate(
			handler.NewCustomerHandler,
			fx.As(new(serverHandler.Handler)),
			fx.ResultTags(serverHandler.ServerControllersGroupTag),
		),
	),
)

// FxTemporal module provides Temporal client and worker
var FxTemporal = fx.Module(
	"Temporalmodule",
	fx.Provide(
		// Provide Temporal client

		func(cfg *config.Config, ctx context.Context) (client.Client, error) {

			temporalHost := cfg.GetString("temporal.host")
			temporalPort := cfg.GetString("temporal.port")
			temporalNamespace := cfg.GetString("temporal.namespace")
			log.Info(ctx, "Connecting temporal at:", temporalHost+":"+temporalPort)
			return client.Dial(client.Options{
				HostPort:  temporalHost + ":" + temporalPort,
				Namespace: temporalNamespace,
			})
		},
		// Provide activity structs
		activities.NewProposalActivities,
		activities.NewAadhaarActivities,
		// Provide PM lifecycle activities; pmTaskQueue is read from config.
		func(cfg *config.Config, proposalRepo *repo.ProposalRepository, c client.Client) *activities.PMLifecycleActivities {
			pmTaskQueue := cfg.GetString("pm.task_queue")
			if pmTaskQueue == "" {
				pmTaskQueue = "policy-manager-queue"
			}
			return activities.NewPMLifecycleActivities(proposalRepo, c, pmTaskQueue)
		},
	),
	fx.Invoke(
		// Register workflows and activities with worker
		func(
			c client.Client,
			cfg *config.Config,
			ctx context.Context,
			proposalActivities *activities.ProposalActivities,
			aadhaarActivities *activities.AadhaarActivities,
			pmActivities *activities.PMLifecycleActivities,
		) error {
			w := worker.New(c, "policy-issue-queue", worker.Options{})

			// Register workflows
			w.RegisterWorkflow(workflows.PolicyIssuanceWorkflow)
			w.RegisterWorkflow(workflows.InstantIssuanceWorkflow)
			w.RegisterWorkflow(workflows.PMSignalReconciliationWorkflow)

			// Register activities
			w.RegisterActivity(proposalActivities.ValidateProposalActivity)
			w.RegisterActivity(proposalActivities.CheckEligibilityActivity)
			w.RegisterActivity(proposalActivities.CalculatePremiumActivity)
			w.RegisterActivity(proposalActivities.SavePremiumToProposalActivity)
			w.RegisterActivity(proposalActivities.UpdateProposalStatusActivity)
			w.RegisterActivity(proposalActivities.SendNotificationActivity)
			w.RegisterActivity(proposalActivities.RequestMedicalReviewActivity)
			w.RegisterActivity(proposalActivities.RouteToApproverActivity)
			w.RegisterActivity(proposalActivities.GeneratePolicyNumberActivity)
			w.RegisterActivity(proposalActivities.GenerateBondActivity)
			w.RegisterActivity(proposalActivities.CreatePolicyIssuanceActivity)
			w.RegisterActivity(proposalActivities.UpdateBondDetailsActivity)
			w.RegisterActivity(aadhaarActivities.ValidateAndCalculatePremiumActivity)
			w.RegisterActivity(aadhaarActivities.CreateAadhaarProposalActivity)
			w.RegisterActivity(aadhaarActivities.CheckInstantIssuanceEligibilityActivity)
			w.RegisterActivity(aadhaarActivities.SendPolicyBondElectronicActivity)
			// PM signal activities
			w.RegisterActivity(pmActivities.StartPMLifecycleActivity)
			w.RegisterActivity(pmActivities.FindUnsignalledPoliciesActivity)

			if err := w.Start(); err != nil {
				return err
			}

			// Create (or update) the Temporal schedule that runs PM reconciliation
			// every 15 minutes. Idempotent — safe to call on every restart.
			_ = ensurePMReconciliationSchedule(ctx, c, cfg)

			return nil
		},
	),
)

// ensurePMReconciliationSchedule creates (or verifies) a Temporal schedule that
// triggers PMSignalReconciliationWorkflow every 15 minutes.
// It is called once on startup and is intentionally best-effort (errors are logged, not fatal).
func ensurePMReconciliationSchedule(ctx context.Context, c client.Client, cfg *config.Config) error {
	scheduleID := "pm-signal-reconciliation-schedule"
	namespace := cfg.GetString("temporal.namespace")
	_ = namespace // namespace is embedded in the client; kept for readability

	_, err := c.ScheduleClient().Create(ctx, client.ScheduleOptions{
		ID: scheduleID,
		Spec: client.ScheduleSpec{
			CronExpressions: []string{"*/15 * * * *"},
		},
		Action: &client.ScheduleWorkflowAction{
			Workflow:  workflows.PMSignalReconciliationWorkflow,
			TaskQueue: workflows.PMReconciliationTaskQueue,
			RetryPolicy: &temporal.RetryPolicy{
				MaximumAttempts: 1, // the schedule itself is the retry loop
			},
		},
		Policies: client.SchedulePolicies{
			Overlap: enums.SCHEDULE_OVERLAP_POLICY_SKIP, // skip if previous run still in progress
		},
	})

	if err != nil {
		// "already exists" is fine — schedule was created on a previous startup.
		// All other errors are logged but non-fatal.
		log.Info(ctx, "ensurePMReconciliationSchedule:", err.Error())
	}

	return nil
}
