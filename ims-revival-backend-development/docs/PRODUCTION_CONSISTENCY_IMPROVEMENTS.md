# Production Consistency Improvements for IndexRevivalRequest

**Date:** 2025-12-30
**Context:** Preventing inconsistent state when CreateRevivalRequest, ExecuteWorkflow, or UpdateWorkflowIDs fail

---

## Current Problem

The `IndexRevivalRequest` handler has 3 sequential operations that can fail independently:

```go
// Step 1: Create DB record
createdRevival, err := h.revivalRepo.CreateRevivalRequest(ctx, revivalReq)
if err != nil {
    return nil, err  // ✅ Safe - nothing created yet
}

// Step 2: Start Temporal workflow
workflowRun, err := h.temporalClient.ExecuteWorkflow(...)
if err != nil {
    return nil, err  // ❌ DANGER: DB record exists, no workflow
}

// Step 3: Update workflow IDs
err = h.revivalRepo.UpdateWorkflowIDsAndCreateState(ctx, requestID, workflowID, runID, status)
if err != nil {
    return nil, err  // ❌ DANGER: Workflow running, DB doesn't know about it
}
```

**Failure Scenarios:**
1. **Workflow start fails** → Orphaned DB record with no workflow
2. **Update workflow IDs fails** → Workflow running but DB has no reference
3. **Network timeout** → Uncertain state (workflow may or may not have started)

---

## Solution 1: Database Transaction + Compensation (RECOMMENDED) ⭐

### Pattern: Two-Phase with Rollback

Use PostgreSQL transaction with compensation logic:

```go
func (h *RevivalHandler) IndexRevivalRequest(sctx *serverRoute.Context, req port.IndexRevivalRequest) (*port.IndexRequestResponse, error) {
    // Validate policy first (read-only, no transaction needed)
    err := h.activities.ValidatePolicyActivity(sctx.Ctx, "", req.PolicyNumber)
    if err != nil {
        return &port.IndexRequestResponse{
            StatusCodeAndMessage: port.PolicyNotEligible,
            Data: port.IndexRequestData{Message: err.Error()},
        }, nil
    }

    // Generate ticket ID
    ticketID, err := h.revivalRepo.GenerateTicketID(sctx.Ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to generate ticket ID: %w", err)
    }

    // Build revival request
    revivalReq := domain.RevivalRequest{
        TicketID:      ticketID,
        PolicyNumber:  req.PolicyNumber,
        RequestType:   "installment_revival",
        CurrentStatus: "INDEXED",
        IndexedDate:   &req.RequestDateTime,
        IndexedBy:     &req.IndexedBy,
    }

    // START ATOMIC SECTION
    var createdRevival domain.RevivalRequest
    var workflowID, runID string
    var workflowStarted bool

    // Begin database transaction
    tx, err := h.db.Begin(sctx.Ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to begin transaction: %w", err)
    }
    defer tx.Rollback(sctx.Ctx) // Auto-rollback on panic or early return

    // Step 1: Create revival request (within transaction)
    createdRevival, err = h.revivalRepo.CreateRevivalRequestTx(sctx.Ctx, tx, revivalReq)
    if err != nil {
        // Transaction auto-rolls back
        return nil, fmt.Errorf("failed to create revival request: %w", err)
    }

    // Step 2: Start Temporal workflow (BEFORE committing transaction)
    workflowOptions := tclient.StartWorkflowOptions{
        ID:        fmt.Sprintf("revival-workflow-%s", createdRevival.RequestID),
        TaskQueue: h.taskQueue,
    }

    workflowRun, err := h.temporalClient.ExecuteWorkflow(
        sctx.Ctx,
        workflowOptions,
        workflow.InstallmentRevivalWorkflow,
        createdRevival.RequestID,
    )
    if err != nil {
        // Workflow failed to start - transaction will auto-rollback
        log.Error(sctx.Ctx, "failed to start workflow, rolling back DB transaction", err)
        return nil, fmt.Errorf("failed to start workflow: %w", err)
    }

    workflowStarted = true
    workflowID = workflowRun.GetID()
    runID = workflowRun.GetRunID()

    // Step 3: Update workflow IDs (within same transaction)
    err = h.revivalRepo.UpdateWorkflowIDsAndCreateStateTx(sctx.Ctx, tx, createdRevival.RequestID, workflowID, runID, "INDEXED")
    if err != nil {
        // CRITICAL: Workflow started but DB update failed
        // Attempt to terminate workflow before rolling back
        log.Error(sctx.Ctx, "failed to update workflow IDs, attempting workflow cleanup", err)

        if terminateErr := h.temporalClient.TerminateWorkflow(
            sctx.Ctx,
            workflowID,
            runID,
            "Database update failed during initialization",
        ); terminateErr != nil {
            log.Error(sctx.Ctx, "failed to terminate orphaned workflow", terminateErr)
            // TODO: Store in orphaned_workflows table for manual cleanup
        }

        return nil, fmt.Errorf("failed to update workflow IDs: %w", err)
    }

    // Commit transaction - both DB and workflow are consistent
    if err := tx.Commit(sctx.Ctx); err != nil {
        // VERY RARE: Commit failed after workflow started
        log.Error(sctx.Ctx, "transaction commit failed, workflow may be orphaned", err)

        // Attempt workflow termination
        if terminateErr := h.temporalClient.TerminateWorkflow(
            sctx.Ctx,
            workflowID,
            runID,
            "Database transaction commit failed",
        ); terminateErr != nil {
            log.Error(sctx.Ctx, "failed to terminate workflow after commit failure", terminateErr)
        }

        return nil, fmt.Errorf("failed to commit transaction: %w", err)
    }

    log.Info(sctx.Ctx, "revival request indexed successfully",
        "ticket_id", createdRevival.TicketID,
        "workflow_id", workflowID)

    return &port.IndexRequestResponse{
        StatusCodeAndMessage: port.CreateSuccess,
        Data: port.IndexRequestData{
            TicketID:        createdRevival.TicketID,
            WorkflowID:      &workflowID,
            Status:          createdRevival.CurrentStatus,
            RequestDateTime: *createdRevival.IndexedDate,
            Message:         "Revival request indexed and workflow started successfully",
        },
    }, nil
}
```

**Key Benefits:**
- ✅ Atomic DB operations
- ✅ Workflow terminated if DB update fails
- ✅ Transaction rolls back if workflow fails
- ✅ Clear error recovery path

**Repository Changes Needed:**
```go
// Add transaction-aware methods to RevivalRepository

func (r *RevivalRepository) CreateRevivalRequestTx(ctx context.Context, tx pgx.Tx, req domain.RevivalRequest) (domain.RevivalRequest, error) {
    // Same as CreateRevivalRequest but uses tx instead of r.db
    // ...
}

func (r *RevivalRepository) UpdateWorkflowIDsAndCreateStateTx(ctx context.Context, tx pgx.Tx, requestID, workflowID, runID, status string) error {
    // Same as UpdateWorkflowIDsAndCreateState but uses tx
    // ...
}
```

---

## Solution 2: Workflow-First Pattern (Alternative) 🔄

Start workflow BEFORE creating DB record. Workflow creates the record:

```go
func (h *RevivalHandler) IndexRevivalRequest(sctx *serverRoute.Context, req port.IndexRevivalRequest) (*port.IndexRequestResponse, error) {
    // Validate policy
    err := h.activities.ValidatePolicyActivity(sctx.Ctx, "", req.PolicyNumber)
    if err != nil {
        return &port.IndexRequestResponse{
            StatusCodeAndMessage: port.PolicyNotEligible,
            Data: port.IndexRequestData{Message: err.Error()},
        }, nil
    }

    // Generate ticket ID (deterministic, idempotent)
    ticketID := fmt.Sprintf("PSREYV%s-%d", req.PolicyNumber, time.Now().Unix())

    // Build workflow input with ALL data needed
    workflowInput := workflow.IndexRevivalInput{
        TicketID:     ticketID,
        PolicyNumber: req.PolicyNumber,
        RequestType:  "installment_revival",
        IndexedBy:    req.IndexedBy,
        RequestOwner: req.RequestOwner,
        IndexedDate:  req.RequestDateTime,
    }

    // Start workflow FIRST - workflow will create DB record
    workflowOptions := tclient.StartWorkflowOptions{
        ID:        fmt.Sprintf("revival-workflow-%s", ticketID),
        TaskQueue: h.taskQueue,
    }

    workflowRun, err := h.temporalClient.ExecuteWorkflow(
        sctx.Ctx,
        workflowOptions,
        workflow.InstallmentRevivalWorkflow,
        workflowInput,
    )
    if err != nil {
        return nil, fmt.Errorf("failed to start workflow: %w", err)
    }

    // Workflow creates DB record as its FIRST activity
    // Handler just returns workflow ID
    return &port.IndexRequestResponse{
        StatusCodeAndMessage: port.CreateSuccess,
        Data: port.IndexRequestData{
            TicketID:        ticketID,
            WorkflowID:      stringPtr(workflowRun.GetID()),
            Status:          "INDEXED",
            RequestDateTime: req.RequestDateTime,
            Message:         "Revival workflow started successfully",
        },
    }, nil
}
```

**Workflow becomes:**
```go
func InstallmentRevivalWorkflow(ctx workflow.Context, input IndexRevivalInput) error {
    // FIRST ACTIVITY: Create DB record (idempotent)
    revivalReq := domain.RevivalRequest{
        TicketID:      input.TicketID,
        PolicyNumber:  input.PolicyNumber,
        RequestType:   input.RequestType,
        CurrentStatus: "INDEXED",
        IndexedDate:   &input.IndexedDate,
        IndexedBy:     &input.IndexedBy,
        WorkflowID:    stringPtr(workflow.GetInfo(ctx).WorkflowExecution.ID),
        RunID:         stringPtr(workflow.GetInfo(ctx).WorkflowExecution.RunID),
    }

    var createdRevival domain.RevivalRequest
    err := workflow.ExecuteActivity(ctx, activities.CreateRevivalRequestActivity, revivalReq).Get(ctx, &createdRevival)
    if err != nil {
        return fmt.Errorf("failed to create revival request: %w", err)
    }

    // Continue with rest of workflow...
    return nil
}
```

**Benefits:**
- ✅ Temporal guarantees exactly-once execution
- ✅ Workflow ID known before DB record created
- ✅ Automatic retries with exponential backoff
- ✅ No orphaned DB records possible

**Drawbacks:**
- ⚠️ Handler response delayed until first activity completes
- ⚠️ Requires workflow redesign

---

## Solution 3: Idempotent Operations + Retry (Simple) 🔁

Make all operations idempotent and add retry logic:

```go
func (h *RevivalHandler) IndexRevivalRequest(sctx *serverRoute.Context, req port.IndexRevivalRequest) (*port.IndexRequestResponse, error) {
    // ... validation ...

    ticketID, err := h.revivalRepo.GenerateTicketID(sctx.Ctx)
    if err != nil {
        return nil, fmt.Errorf("failed to generate ticket ID: %w", err)
    }

    revivalReq := domain.RevivalRequest{
        TicketID:      ticketID,
        PolicyNumber:  req.PolicyNumber,
        RequestType:   "installment_revival",
        CurrentStatus: "INDEXED",
        IndexedDate:   &req.RequestDateTime,
        IndexedBy:     &req.IndexedBy,
    }

    // Idempotent create (ON CONFLICT DO UPDATE)
    createdRevival, err := h.revivalRepo.UpsertRevivalRequest(sctx.Ctx, revivalReq)
    if err != nil {
        return nil, fmt.Errorf("failed to create revival request: %w", err)
    }

    // Idempotent workflow start (Temporal handles duplicate workflow IDs)
    workflowOptions := tclient.StartWorkflowOptions{
        ID:                       fmt.Sprintf("revival-workflow-%s", createdRevival.RequestID),
        TaskQueue:                h.taskQueue,
        WorkflowIDReusePolicy:    enums.WORKFLOW_ID_REUSE_POLICY_ALLOW_DUPLICATE_FAILED_ONLY,
    }

    // Retry workflow start with exponential backoff
    var workflowRun tclient.WorkflowRun
    err = retry.Do(
        func() error {
            var err error
            workflowRun, err = h.temporalClient.ExecuteWorkflow(sctx.Ctx, workflowOptions, workflow.InstallmentRevivalWorkflow, createdRevival.RequestID)
            return err
        },
        retry.Attempts(3),
        retry.Delay(100*time.Millisecond),
        retry.DelayType(retry.BackOffDelay),
    )
    if err != nil {
        return nil, fmt.Errorf("failed to start workflow after retries: %w", err)
    }

    // Idempotent update (ON CONFLICT DO UPDATE)
    workflowID := workflowRun.GetID()
    runID := workflowRun.GetRunID()
    err = h.revivalRepo.UpsertWorkflowIDs(sctx.Ctx, createdRevival.RequestID, workflowID, runID)
    if err != nil {
        return nil, fmt.Errorf("failed to update workflow IDs: %w", err)
    }

    return &port.IndexRequestResponse{
        StatusCodeAndMessage: port.CreateSuccess,
        Data: port.IndexRequestData{
            TicketID:        createdRevival.TicketID,
            WorkflowID:      &workflowID,
            Status:          createdRevival.CurrentStatus,
            RequestDateTime: *createdRevival.IndexedDate,
            Message:         "Revival request indexed and workflow started successfully",
        },
    }, nil
}
```

**Repository changes:**
```go
func (r *RevivalRepository) UpsertRevivalRequest(ctx context.Context, req domain.RevivalRequest) (domain.RevivalRequest, error) {
    // Use INSERT ... ON CONFLICT (ticket_id) DO UPDATE
    // Returns existing record if duplicate
}

func (r *RevivalRepository) UpsertWorkflowIDs(ctx context.Context, requestID, workflowID, runID string) error {
    // UPDATE with WHERE clause, safe to retry
}
```

**Benefits:**
- ✅ Simple implementation
- ✅ Safe retries
- ✅ No complex transaction logic

**Drawbacks:**
- ⚠️ Doesn't prevent orphaned workflows

---

## Solution 4: Saga Pattern with Compensation 🔄

Implement compensating transactions for each step:

```go
type IndexRevivalSaga struct {
    revivalRepo    *repo.RevivalRepository
    temporalClient tclient.Client
    steps          []SagaStep
}

type SagaStep struct {
    Name        string
    Execute     func(ctx context.Context) error
    Compensate  func(ctx context.Context) error
    Completed   bool
}

func (s *IndexRevivalSaga) Execute(ctx context.Context) error {
    var completedSteps []SagaStep

    for i, step := range s.steps {
        if err := step.Execute(ctx); err != nil {
            // Failure - rollback all completed steps
            log.Error(ctx, "saga step failed, rolling back", "step", step.Name, "error", err)

            for j := len(completedSteps) - 1; j >= 0; j-- {
                if compensateErr := completedSteps[j].Compensate(ctx); compensateErr != nil {
                    log.Error(ctx, "compensation failed", "step", completedSteps[j].Name, "error", compensateErr)
                }
            }

            return fmt.Errorf("saga failed at step %s: %w", step.Name, err)
        }

        s.steps[i].Completed = true
        completedSteps = append(completedSteps, step)
    }

    return nil
}
```

---

## Solution 5: Event Sourcing (Advanced) 📜

Store all state changes as events:

```go
// Event store records every state transition
type RevivalEvent struct {
    EventID      string
    RequestID    string
    EventType    string    // "CREATED", "WORKFLOW_STARTED", "WORKFLOW_LINKED"
    EventData    json.RawMessage
    Timestamp    time.Time
    ProcessedAt  *time.Time
}

// Handler emits events, background worker processes them
func (h *RevivalHandler) IndexRevivalRequest(ctx context.Context, req port.IndexRevivalRequest) (*port.IndexRequestResponse, error) {
    // Emit "REVIVAL_REQUESTED" event
    event := RevivalEvent{
        EventType: "REVIVAL_REQUESTED",
        EventData: marshalJSON(req),
        Timestamp: time.Now(),
    }

    if err := h.eventStore.Append(ctx, event); err != nil {
        return nil, err
    }

    // Background worker processes event and handles retries
    // Returns ticket ID immediately
}
```

---

## Recommendation Summary

| Solution | Complexity | Safety | Performance | Use Case |
|----------|-----------|--------|-------------|----------|
| **1. DB Transaction + Compensation** | Medium | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐ | **BEST for production** |
| **2. Workflow-First** | Medium | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ | Good for async flows |
| **3. Idempotent + Retry** | Low | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ | Quick improvement |
| **4. Saga Pattern** | High | ⭐⭐⭐⭐ | ⭐⭐⭐ | Complex workflows |
| **5. Event Sourcing** | Very High | ⭐⭐⭐⭐⭐ | ⭐⭐⭐ | High-scale systems |

---

## Monitoring & Alerting

Regardless of solution, add monitoring:

```go
// Orphaned workflow detector (run periodically)
func DetectOrphanedWorkflows(ctx context.Context) error {
    // Find workflows without DB records
    workflows := getRunningWorkflows(ctx)

    for _, wf := range workflows {
        requestID := extractRequestID(wf.ID)
        exists, err := revivalRepo.RevivalRequestExists(ctx, requestID)
        if err != nil || !exists {
            // Alert: Orphaned workflow found
            alerting.Send("Orphaned workflow detected", wf.ID)
        }
    }
}

// Orphaned DB record detector
func DetectOrphanedRecords(ctx context.Context) error {
    // Find DB records without workflows
    records := revivalRepo.GetRecordsWithoutWorkflowID(ctx)

    for _, record := range records {
        if time.Since(record.CreatedAt) > 5*time.Minute {
            // Alert: Record created but no workflow
            alerting.Send("Orphaned DB record detected", record.TicketID)
        }
    }
}
```

---

## Implementation Plan

**Phase 1 (Immediate):**
1. Add idempotency to repository methods
2. Add retry logic with exponential backoff
3. Add monitoring for orphaned workflows/records

**Phase 2 (1-2 weeks):**
1. Implement DB transaction support
2. Add workflow termination on DB failure
3. Add compensation logic

**Phase 3 (Future):**
1. Consider workflow-first pattern for async flows
2. Implement event sourcing if scale demands it

---

**Status:** Ready for implementation
**Priority:** HIGH (production consistency issue)
**Estimated Effort:** 2-3 days for Solution 1
