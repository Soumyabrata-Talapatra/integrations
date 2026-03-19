package bootstrap

import (
	"context"
	"fmt"
	"time"

	enums "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/client"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"go.temporal.io/sdk/worker"
	"go.uber.org/fx"

	config "gitlab.cept.gov.in/it-2.0-common/api-config"
	serverHandler "gitlab.cept.gov.in/it-2.0-common/n-api-server/handler"

	// Domain constants — batch scan type names for Temporal Schedule args [Constraint 5]
	"policy-management/core/domain"

	// Handlers — Phase 4 complete
	handler "policy-management/handler"

	// Repositories — Phase 3 complete
	repo "policy-management/repo/postgres"

	// Workflows — Phase 5 [FR-PM-001, FR-PM-011..FR-PM-015]
	"policy-management/workflows"

	// Activities — Phase 5 [FR-PM-001, FR-PM-003, FR-PM-011..FR-PM-015]
	acts "policy-management/workflows/activities"
)

// ---------------------------------------------------------------------------
// FxRepo — all PostgreSQL repositories
// ---------------------------------------------------------------------------
// FR-PM-001: Policy lifecycle state persistence
// FR-PM-006: Service request persistence
// Populated progressively in Phase 3.
var FxRepo = fx.Module(
	"Repomodule",
	fx.Provide(
		repo.NewPolicyRepository,
		repo.NewServiceRequestRepository,
		repo.NewSignalRepository,
		repo.NewConfigRepository,
	),
)

// ---------------------------------------------------------------------------
// FxHandler — all HTTP handlers (33 endpoints, Phase 4)
// ---------------------------------------------------------------------------
// FR-PM-001..FR-PM-008: All REST endpoints.
// Handlers MUST use fx.Annotate + ServerControllersGroupTag.
// Populated progressively in Phase 4.
// Additional-6: nil serverHandler.Handler provider removed — it caused a runtime
// panic when n-api-server iterated over the group. Fx provides an empty slice
// when no providers are tagged with ServerControllersGroupTag.
var FxHandler = fx.Module(
	"Handlermodule",
	fx.Provide(
		// Phase 4.1 — Request Submission Handlers (19 endpoints)
		// FR-PM-002..FR-PM-010: All policy service request endpoints.
		fx.Annotate(handler.NewPolicyRequestHandler,
			fx.As(new(serverHandler.Handler)),
			fx.ResultTags(serverHandler.ServerControllersGroupTag)),

		// Phase 4.2 — Quote Proxy Handlers (3 endpoints)
		// FR-PM-003: Synchronous quote proxy via short-lived workflows.
		fx.Annotate(handler.NewQuoteHandler,
			fx.As(new(serverHandler.Handler)),
			fx.ResultTags(serverHandler.ServerControllersGroupTag)),

		// Phase 4.3 — Request Lifecycle Handlers (5 endpoints)
		// FR-PM-006: List, detail, withdraw service requests; CPC pending inbox.
		fx.Annotate(handler.NewRequestLifecycleHandler,
			fx.As(new(serverHandler.Handler)),
			fx.ResultTags(serverHandler.ServerControllersGroupTag)),

		// Phase 4.4 — Policy Query Handlers (6 endpoints)
		// FR-PM-004: Policy status, summary, state-gate, history, batch, dashboard.
		fx.Annotate(handler.NewPolicyQueryHandler,
			fx.As(new(serverHandler.Handler)),
			fx.ResultTags(serverHandler.ServerControllersGroupTag)),

		// Phase 4.5 — CPC & Static Lookup Handlers (6 endpoints)
		// FR-PM-004, FR-PM-008: Static enum lookup endpoints (no DB/Temporal calls).
		fx.Annotate(handler.NewCPCLookupHandler,
			fx.As(new(serverHandler.Handler)),
			fx.ResultTags(serverHandler.ServerControllersGroupTag)),
	),
)

// ---------------------------------------------------------------------------
// FxTemporal — Temporal client + worker on policy-management-tq
// ---------------------------------------------------------------------------
// Constraint 3: Single task queue "policy-management-tq" for ALL PM workflows + activities.
// Constraint 4: Namespace "pli-insurance".
// FR-PM-001: PolicyLifecycleWorkflow (per-policy long-running).
// FR-PM-011..FR-PM-015: BatchStateScanWorkflow (6 Temporal Schedules).
var FxTemporal = fx.Module(
	"Temporalmodule",
	fx.Provide(
		// Temporal client — namespace and host read from config (Additional-5).
		// Config keys: temporal.hostport (default localhost:7233), temporal.namespace (default pli-insurance).
		// Overridden per environment via configs/config.{env}.yaml.
		func(cfg *config.Config) (client.Client, error) {
			hostPort := cfg.GetString("temporal.hostport")
			if hostPort == "" {
				hostPort = "localhost:7233" // safe fallback for local dev
			}
			namespace := cfg.GetString("temporal.namespace")
			if namespace == "" {
				namespace = "pli-insurance" // safe fallback for local dev
			}
			c, err := client.Dial(client.Options{
				HostPort:  hostPort,
				Namespace: namespace,
			})
			if err != nil {
				return nil, fmt.Errorf("temporal client dial hostPort=%s namespace=%s: %w", hostPort, namespace, err)
			}
			return c, nil
		},

		// Activity struct providers — Phase 5.3
		// FX resolves constructor args (db, cfg, tc) automatically from the container.
		acts.NewPolicyActivities, // PolicyActivities — 12 policy lifecycle activities [FR-PM-001]
		acts.NewBatchActivities,  // BatchActivities  — 7 batch scan activities [FR-PM-011..FR-PM-015]
		acts.NewQuoteActivities,  // QuoteActivities  — 3 quote proxy activities [FR-PM-003]
	),

	fx.Invoke(
		// Register all workflows and activities with the single PM worker on policy-management-tq.
		// Constraint 3: ONE task queue for all PM workflows + activities.
		// Constraint 4: namespace "pli-insurance" (set via client.Dial options above).
		// [FR-PM-001, FR-PM-003, FR-PM-011..FR-PM-015, Constraint 3, Constraint 4]
		func(
			lc fx.Lifecycle,
			c client.Client,
			policyActs *acts.PolicyActivities,
			batchActs *acts.BatchActivities,
			quoteActs *acts.QuoteActivities,
		) error {
			w := worker.New(c, "policy-management-tq", worker.Options{
				// FR-PM-001: Concurrency limits for 3M active workflows
				MaxConcurrentWorkflowTaskExecutionSize: 100,
				MaxConcurrentActivityExecutionSize:     200,
			})

			// ── Workflow registrations ──────────────────────────────────────────
			// PolicyLifecycleWorkflow — long-running per-policy CAN workflow [FR-PM-001]
			w.RegisterWorkflow(workflows.PolicyLifecycleWorkflow)
			// BatchStateScanWorkflow — per-scan-type batch workflow [FR-PM-011..FR-PM-015]
			w.RegisterWorkflow(workflows.BatchStateScanWorkflow)
			// Short-lived quote proxy workflows — must be registered by exact function
			// name so QuoteHandler can invoke via client.ExecuteWorkflow(wfType, ...) [FR-PM-003]
			w.RegisterWorkflow(acts.GetSurrenderQuoteWorkflow)
			w.RegisterWorkflow(acts.GetLoanQuoteWorkflow)
			w.RegisterWorkflow(acts.GetConversionQuoteWorkflow)

			// ── Activity registrations (struct = all methods registered at once) ─
			// Registering a struct pointer registers every exported method as an
			// activity. FX provides instances with all dependencies already injected.
			w.RegisterActivity(policyActs) // 12 policy lifecycle activities [FR-PM-001]
			w.RegisterActivity(batchActs)  // 7 batch scan activities       [FR-PM-011..FR-PM-015]
			w.RegisterActivity(quoteActs)  // 3 quote proxy activities      [FR-PM-003]

			lc.Append(fx.Hook{
				OnStart: func(ctx context.Context) error {
					return w.Start()
				},
				OnStop: func(ctx context.Context) error {
					w.Stop()
					return nil
				},
			})
			return nil
		},
	),
)

// ---------------------------------------------------------------------------
// RegisterBatchSchedules — create 6 Temporal Schedules for BatchStateScanWorkflow
// ---------------------------------------------------------------------------
// Idempotent — skips any schedule that already exists, so safe to call on every
// deploy or to re-run after a failed partial registration.
//
// Intended to be called ONCE during environment setup (NOT on every app start).
// See migrations/004_create_temporal_schedules.sh for the invocation script.
//
// Constraint 5: Uses ScheduleClient().Create() — NOT the workflow CronSchedule property.
// Schedule IDs and cron expressions (IST = UTC+5:30):
//
//	batch-lapsation-daily          → LAPSATION             "30 0 * * *"   (00:30 IST)
//	batch-remission-short-daily    → REMISSION_EXPIRY_SHORT "35 0 * * *"   (00:35 IST)
//	batch-remission-long-daily     → REMISSION_EXPIRY_LONG  "40 0 * * *"   (00:40 IST)
//	batch-paidup-monthly           → PAID_UP_CONVERSION     "0 1 1 * *"    (01:00 1st monthly)
//	batch-maturity-daily           → MATURITY_SCAN          "0 2 * * *"    (02:00 IST)
//	batch-forced-surrender-monthly → FORCED_SURRENDER_EVAL  "0 3 1 * *"    (03:00 1st monthly)
//
// [FR-PM-011..FR-PM-015, Constraint 5, §9.3]
func RegisterBatchSchedules(ctx context.Context, c client.Client) error {
	type batchSchedule struct {
		id       string
		scanType string
		cron     string
	}

	schedules := []batchSchedule{
		{
			id:       "batch-lapsation-daily",
			scanType: domain.BatchScanTypeLapsation,
			cron:     "30 0 * * *",
		},
		{
			id:       "batch-remission-short-daily",
			scanType: domain.BatchScanTypeRemissionExpiryShort,
			cron:     "35 0 * * *",
		},
		{
			id:       "batch-remission-long-daily",
			scanType: domain.BatchScanTypeRemissionExpiryLong,
			cron:     "40 0 * * *",
		},
		{
			id:       "batch-paidup-monthly",
			scanType: domain.BatchScanTypePaidUpConversion,
			cron:     "0 1 1 * *",
		},
		{
			id:       "batch-maturity-daily",
			scanType: domain.BatchScanTypeMaturityScan,
			cron:     "0 2 * * *",
		},
		{
			id:       "batch-forced-surrender-monthly",
			scanType: domain.BatchScanTypeForcedSurrenderEval,
			cron:     "0 3 1 * *",
		},
	}

	sc := c.ScheduleClient()

	for _, s := range schedules {
		_, err := sc.Create(ctx, client.ScheduleOptions{
			ID: s.id,
			Spec: client.ScheduleSpec{
				CronExpressions: []string{s.cron},
			},
			Action: &client.ScheduleWorkflowAction{
				// WorkflowID is omitted — Temporal auto-generates a unique ID per run
				// so concurrent overlap detection works correctly. [§9.3]
				Workflow: workflows.BatchStateScanWorkflow,
				// scheduledDate = time.Time{} → BatchStateScanWorkflow defaults to
				// workflow.Now(ctx) so the scan processes the day it actually runs.
				// Explicit dates can be passed for backfill via manual trigger. [§9.3]
				Args:                     []interface{}{s.scanType, time.Time{}},
				TaskQueue:                "policy-management-tq",
				WorkflowExecutionTimeout: 4 * time.Hour, // covers 2h activity + overhead [§9.3]
			},
			// Skip overlap: if a scheduled run is still active when the next fire time
			// arrives, the new run is dropped — prevents double-processing. [§9.5.2]
			Overlap: enums.SCHEDULE_OVERLAP_POLICY_SKIP,
		})
		if err != nil {
			if scheduleAlreadyExists(err) {
				// Idempotent — schedule was already registered; nothing to do.
				continue
			}
			return fmt.Errorf("RegisterBatchSchedules: create schedule %q: %w", s.id, err)
		}
	}

	return nil
}

// scheduleAlreadyExists returns true when Temporal's ScheduleClient.Create() fails
// because a schedule with the same ID is already registered.
// Uses the typed gRPC status code check instead of brittle string matching so
// it works regardless of how Temporal formats the error message. [D13, Constraint 5]
func scheduleAlreadyExists(err error) bool {
	if err == nil {
		return false
	}
	return status.Code(err) == codes.AlreadyExists
}
