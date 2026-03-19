package temporal

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	config "gitlab.cept.gov.in/it-2.0-common/api-config"

	log "gitlab.cept.gov.in/it-2.0-common/n-api-log"

	"gitlab.cept.gov.in/it-2.0-policy/surrender-service/temporal/activities"
	"gitlab.cept.gov.in/it-2.0-policy/surrender-service/temporal/workflows"
)

const (
	// Task queue names
	// SurrenderTaskQueue must match the task queue PM uses when dispatching
	// SurrenderProcessingWorkflow as a child workflow ("surrender-tq").
	SurrenderTaskQueue       = "surrender-tq"
	ApprovalTaskQueue        = "approval-task-queue"
	DocumentTaskQueue        = "document-task-queue"
	PaymentTaskQueue         = "payment-task-queue"
	ForcedSurrenderTaskQueue = "forced-surrender-task-queue"
)

// Worker manages Temporal workflow and activity workers
type Worker struct {
	client client.Client
	worker worker.Worker
}

// NewTemporalClient creates a new Temporal client
func NewTemporalClient(cfg *config.Config) (client.Client, error) {
	// Get Temporal server address from config (default to localhost:7233)
	host := cfg.GetString("temporal.host")
	Port := cfg.GetString("temporal.port")
	namespace := cfg.GetString("temporal.namespace")

	// Create the client
	c, err := client.Dial(client.Options{
		HostPort:  host + ":" + Port,
		Namespace: namespace,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Temporal client: %w", err)
	}

	return c, nil
}

// WorkerManager manages the Temporal worker lifecycle
type WorkerManager struct {
	worker *Worker
	ctx    context.Context
}

// NewWorkerManager creates a new worker manager
func NewWorkerManager(temporalClient client.Client) *WorkerManager {
	return &WorkerManager{
		worker: NewWorker(temporalClient),
		ctx:    context.Background(),
	}
}

// RegisterWorkflows starts the Temporal worker
func RegisterWorkflows(wm *WorkerManager) error {
	// Start worker in background
	go func() {
		if err := wm.worker.Start(wm.ctx); err != nil {
			log.Error(wm.ctx, "Temporal worker failed: %v", err)
		}
	}()

	log.Info(wm.ctx, "Temporal workflows registered and worker started")
	return nil
}

// NewWorker creates a new Temporal worker
func NewWorker(temporalClient client.Client) *Worker {
	// Create worker that listens on surrender task queue
	w := worker.New(temporalClient, SurrenderTaskQueue, worker.Options{})

	// Register all workflows
	// SurrenderProcessingWorkflow replaces VoluntarySurrenderWorkflow.
	// It is dispatched by PM's PolicyLifecycleWorkflow as a child workflow.
	w.RegisterWorkflow(workflows.SurrenderProcessingWorkflow)
	w.RegisterWorkflow(workflows.ForcedSurrenderWorkflow)
	w.RegisterWorkflow(workflows.PaymentWindowMonitorWorkflow)
	w.RegisterWorkflow(workflows.ApprovalWorkflow)
	w.RegisterWorkflow(workflows.PaymentWorkflow)
	w.RegisterWorkflow(workflows.DocumentVerificationWorkflow)
	w.RegisterWorkflow(workflows.PolicyUpdateWorkflow)

	// Register all activities - Surrender Processing (PM-driven)
	w.RegisterActivity(activities.IndexSurrenderActivity)
	w.RegisterActivity(activities.ValidateEligibilityActivity)
	w.RegisterActivity(activities.SubmitDEActivity)
	w.RegisterActivity(activities.SubmitQCActivity)
	w.RegisterActivity(activities.SubmitApprovalActivity)
	w.RegisterActivity(activities.CalculateSurrenderValueActivity)
	w.RegisterActivity(activities.ProcessPaymentActivity)
	w.RegisterActivity(activities.UpdatePolicyStatusActivity)
	w.RegisterActivity(activities.SignalPMWorkflowActivity)
	// Legacy activities kept for other workflows
	w.RegisterActivity(activities.VerifyDocumentsActivity)
	w.RegisterActivity(activities.RouteToApprovalActivity)

	// Register activities - Forced Surrender
	w.RegisterActivity(activities.IdentifyEligiblePoliciesActivity)
	w.RegisterActivity(activities.CreateRemindersBatchActivity)
	w.RegisterActivity(activities.CheckExpiredPaymentWindowsActivity)
	w.RegisterActivity(activities.InitiateForcedSurrendersBatchActivity)

	// Register activities - Document Verification
	w.RegisterActivity(activities.ValidateRequiredDocumentsActivity)
	w.RegisterActivity(activities.ExtractDocumentDataActivity)
	w.RegisterActivity(activities.AutoVerifyDocumentActivity)
	w.RegisterActivity(activities.RouteToManualVerificationActivity)
	w.RegisterActivity(activities.UpdateSurrenderStatusActivity)

	// Register activities - Approval
	w.RegisterActivity(activities.GetSurrenderRequestDetailsActivity)
	w.RegisterActivity(activities.AutoApproveActivity)
	w.RegisterActivity(activities.CreateApprovalTaskActivity)
	w.RegisterActivity(activities.EscalateApprovalTaskActivity)
	w.RegisterActivity(activities.ProcessApprovalDecisionActivity)

	// Register activities - Payment
	w.RegisterActivity(activities.ValidatePaymentEligibilityActivity)
	w.RegisterActivity(activities.DetermineDispositionActivity)
	w.RegisterActivity(activities.CreateDispositionRecordActivity)
	w.RegisterActivity(activities.SendPaymentNotificationActivity)

	// Register activities - Policy Update
	w.RegisterActivity(activities.ValidateStatusTransitionActivity)
	w.RegisterActivity(activities.CreatePolicyHistoryActivity)
	w.RegisterActivity(activities.SettlePolicyLoansActivity)
	w.RegisterActivity(activities.StopFutureBonusesActivity)
	w.RegisterActivity(activities.UpdateReducedPaidUpDetailsActivity)
	w.RegisterActivity(activities.SendPolicyUpdateNotificationActivity)
	w.RegisterActivity(activities.ArchivePolicyRecordsActivity)

	return &Worker{
		client: temporalClient,
		worker: w,
	}
}

// Start starts the worker
func (w *Worker) Start(ctx context.Context) error {
	log.Info(ctx, "Starting Temporal worker on task queue: %s", SurrenderTaskQueue)

	err := w.worker.Run(worker.InterruptCh())
	if err != nil {
		log.Error(ctx, "Failed to start Temporal worker: %v", err)
		return err
	}

	return nil
}

// Stop stops the worker
func (w *Worker) Stop() {
	w.worker.Stop()
}
