package bootstrap

import (
	"go.uber.org/fx"

	handler "gitlab.cept.gov.in/it-2.0-policy/surrender-service/handler"
	repo "gitlab.cept.gov.in/it-2.0-policy/surrender-service/repo/postgres"
	"gitlab.cept.gov.in/it-2.0-policy/surrender-service/temporal"
	"gitlab.cept.gov.in/it-2.0-policy/surrender-service/temporal/activities"

	serverHandler "gitlab.cept.gov.in/it-2.0-common/n-api-server/handler"
)

// FxRepo module provides all repository implementations
var FxRepo = fx.Module(
	"Repomodule",
	fx.Provide(
		// Add repository providers here
		repo.NewSurrenderRequestRepository,
		repo.NewDocumentRepository,
		repo.NewApprovalWorkflowRepository,
		repo.NewForcedSurrenderRepository,
	),
	fx.Invoke(
		// Initialize activities with repository
		activities.InitVoluntarySurrenderActivities,
	),
)

// FxHandler module provides all HTTP handlers
var FxHandler = fx.Module(
	"Handlermodule",
	fx.Provide(
		// Voluntary Surrender Handlers
		fx.Annotate(
			handler.NewVoluntarySurrenderHandler,
			fx.As(new(serverHandler.Handler)),
			fx.ResultTags(serverHandler.ServerControllersGroupTag),
		),

		// Forced Surrender Internal Handlers
		fx.Annotate(
			handler.NewForcedSurrenderHandler,
			fx.As(new(serverHandler.Handler)),
			fx.ResultTags(serverHandler.ServerControllersGroupTag),
		),

		// Approval Workflow Handlers
		fx.Annotate(
			handler.NewApprovalHandler,
			fx.As(new(serverHandler.Handler)),
			fx.ResultTags(serverHandler.ServerControllersGroupTag),
		),
	),
)

// FxTemporal module provides Temporal workflow and activity workers
var FxTemporal = fx.Module(
	"Temporalmodule",
	fx.Provide(
		temporal.NewTemporalClient,
		temporal.NewWorkerManager,
	),
	fx.Invoke(
		// InitPMSignalActivities must be called before RegisterWorkflows so the
		// Temporal client is available when SignalPMWorkflowActivity executes.
		activities.InitPMSignalActivities,
		temporal.RegisterWorkflows,
	),
)
