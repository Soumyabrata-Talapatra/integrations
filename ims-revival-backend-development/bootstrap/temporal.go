package bootstrap

import (
	"context"
	"time"

	"plirevival/workflow"

	config "gitlab.cept.gov.in/it-2.0-common/api-config"
	log "gitlab.cept.gov.in/it-2.0-common/n-api-log"
	tclient "go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.uber.org/fx"
)

func temporalclient(ctx context.Context, c *config.Config) (temporalclient tclient.Client, err error) {
	TemporalHost := c.GetString("temporal.host")
	TemporalPort := c.GetString("temporal.port")
	//cert, _ := tls.LoadX509KeyPair("../cept.cer", "")
	// TemporalHost := "localhost"
	// TemporalPort := "7233"
	logger := TemporalLoggerAdapter(ctx)
	hostPort := TemporalHost + ":" + TemporalPort
	temporalClient, err := tclient.Dial(tclient.Options{
		HostPort: hostPort,
		//Namespace: "default",
		Logger:    logger,
		Namespace: c.GetString("temporal.namespace"),
		// ConnectionOptions: tclient.ConnectionOptions{
		// 	TLS: &tls.Config{
		// 		Certificates: []tls.Certificate{cert},
		// 	},
		// },
	})
	if err != nil {
		return nil, err
	}

	return temporalClient, nil

}

func ProvideTemporalWorker(config *config.Config, c tclient.Client, activities *workflow.Activities) worker.Worker {

	// Set up Temporal Worker
	w := worker.New(c, config.GetString("temporal.taskqueue"), worker.Options{

		MaxConcurrentActivityExecutionSize:     100,
		MaxConcurrentWorkflowTaskExecutionSize: 100,
		MaxConcurrentActivityTaskPollers:       5,
		MaxConcurrentWorkflowTaskPollers:       5,

		WorkerActivitiesPerSecond:    50, // Allows decent throughput
		TaskQueueActivitiesPerSecond: 40,
		StickyScheduleToStartTimeout: time.Minute * 5,
		DeadlockDetectionTimeout:     time.Second * 5,
	})

	w.RegisterWorkflow(workflow.InstallmentRevivalWorkflow)
	w.RegisterWorkflow(workflow.FirstCollectionWorkflow)
	w.RegisterWorkflow(workflow.ChequeMonitorWorkflow)
	w.RegisterWorkflow(workflow.InstallmentMonitorWorkflow)
	w.RegisterWorkflow(workflow.SLATimerWorkflow)
	w.RegisterWorkflow(workflow.BatchInstallmentProcessingWorkflow)

	// Register Workflows
	//w.RegisterWorkflow(service.ParentWorkflow)

	// Register Activities
	//w.RegisterActivity(service.PreWorkflowDetails)

	// Revival request lifecycle activities
	w.RegisterActivity(activities.CreateRevivalRequestActivity)
	w.RegisterActivity(activities.UpdateRevivalStatusActivity)
	w.RegisterActivity(activities.UpdateDataEntryActivity)
	w.RegisterActivity(activities.UpdateQCActivity)
	w.RegisterActivity(activities.UpdateApprovalActivity)
	w.RegisterActivity(activities.CheckAndAdjustSuspenseActivity)
	w.RegisterActivity(activities.FinalizeRevivalAfterFirstCollection)

	// Policy validation activities
	w.RegisterActivity(activities.ValidatePolicyActivity)
	w.RegisterActivity(activities.ValidateMaturityDateConstraintActivity)

	// Workflow termination activities
	w.RegisterActivity(activities.TerminateAndReturnToIndexerActivity)

	// Collection activities
	w.RegisterActivity(activities.ValidateDualCollectionActivity)
	w.RegisterActivity(activities.ProcessDualPaymentActivity)
	w.RegisterActivity(activities.CreateChequeRecordActivity)

	// Installment activities
	w.RegisterActivity(activities.ProcessInstallmentActivity)
	w.RegisterActivity(activities.HandleDefaultActivity)

	// Workflow management activities
	w.RegisterActivity(activities.TerminateRevivalActivity)

	// Communication activities
	w.RegisterActivity(activities.GenerateLetterActivity)
	w.RegisterActivity(activities.SendNotificationActivity)

	//Sla update activity
	w.RegisterActivity(activities.UpdateWorkflowStateActivity)

	// PM integration activities
	w.RegisterActivity(activities.NotifyPolicyManagementActivity)

	return w
}

func temporallifecycle(lc fx.Lifecycle, temporalclient tclient.Client) {
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			return nil
		},
		OnStop: func(ctx context.Context) error {
			temporalclient.Close()
			return nil
		},
	})

}

func RunWorker(lc fx.Lifecycle, w worker.Worker) {
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			go func() {
				err := w.Run(worker.InterruptCh())
				if err != nil {
					log.Fatal(nil, "Unable to start Worker", err)
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			w.Stop()
			return nil
		},
	})
}

var Fxtemporal = fx.Module(
	"temporal",
	fx.Provide(
		temporalclient,
		ProvideTemporalWorker,
	),
	fx.Invoke(temporallifecycle, RunWorker),
	// Temporal Client Initialization

)
