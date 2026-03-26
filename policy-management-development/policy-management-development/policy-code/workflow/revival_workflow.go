package workflow

// import (
// 	"fmt"
// 	"time"

// 	"go.temporal.io/api/enums/v1"
// 	"go.temporal.io/sdk/temporal"
// 	"go.temporal.io/sdk/workflow"
// )

// // IndexRevivalInput contains all data needed to start the revival workflow
// // Note: MissingDocumentsList is NOT part of indexing - sent by data entry/QC/approver only
// // MaturityDate is fetched via ValidatePolicyActivity (batch query) and cached in workflow state
// type IndexRevivalInput struct {
// 	TicketID     string    `json:"ticket_id"`
// 	PolicyNumber string    `json:"policy_number"`
// 	RequestType  string    `json:"request_type"`
// 	IndexedBy    string    `json:"indexed_by"`
// 	IndexedDate  time.Time `json:"indexed_date"`
// 	Documents    string    `json:"documents"` // JSONB as string
// }

// // InstallmentRevivalWorkflow is the main workflow orchestrating the complete revival process
// // Lifespan: 6-12 months
// func InstallmentRevivalWorkflow(ctx workflow.Context, input IndexRevivalInput) error {
// 	logger := workflow.GetLogger(ctx)

// 	// Get workflow execution info
// 	workflowInfo := workflow.GetInfo(ctx)
// 	workflowID := workflowInfo.WorkflowExecution.ID
// 	runID := workflowInfo.WorkflowExecution.RunID

// 	// Workflow state variable
// 	// MaturityDate will be set after ValidatePolicyActivity (batch query)
// 	state := &RevivalWorkflowState{
// 		RequestID:     "", // Will be set after CreateRevivalRequestActivity
// 		TicketID:      input.TicketID,
// 		PolicyNumber:  input.PolicyNumber,
// 		CurrentStatus: "INITIALIZING",
// 		StartedAt:     workflow.Now(ctx),
// 	}

// 	// 🔍 Register query handler - allows external systems to query state
// 	err := workflow.SetQueryHandler(ctx, "getState", func() (string, error) {
// 		return state.CurrentStatus, nil
// 	})
// 	if err != nil {
// 		logger.Error("Failed to set query handler", "error", err)
// 		setRecoverableWorkflowError(state, "INITIALIZING", "Unable to register workflow query handler", err, workflow.Now(ctx))
// 	}

// 	// 🔍 Register detailed query handler
// 	err = workflow.SetQueryHandler(ctx, "getStateDetails", func() (*RevivalWorkflowState, error) {
// 		return state, nil
// 	})
// 	if err != nil {
// 		logger.Error("Failed to set detailed query handler", "error", err)
// 		setRecoverableWorkflowError(state, "INITIALIZING", "Unable to register detailed workflow query handler", err, workflow.Now(ctx))
// 	}

// 	// Activity options with retry policy
// 	activityOptions := workflow.ActivityOptions{
// 		StartToCloseTimeout: 30 * time.Second,
// 		RetryPolicy: &temporal.RetryPolicy{
// 			InitialInterval:    1 * time.Second,
// 			BackoffCoefficient: 2.0,
// 			MaximumInterval:    30 * time.Second,
// 			MaximumAttempts:    3,
// 		},
// 	}
// 	activityCtx := workflow.WithActivityOptions(ctx, activityOptions)

// 	// 📝 Update state: VALIDATING_POLICY
// 	state.CurrentStatus = "VALIDATING_POLICY"
// 	logger.Info("Validating policy for revival")

// 	// ACTIVITY 0: Validate policy and get maturity date (single batch query)
// 	// This replaces separate GetPolicyByNumber call in handler
// 	var policyValidation PolicyValidationResult
// 	err = workflow.ExecuteActivity(activityCtx, "ValidatePolicyActivity", input.PolicyNumber).Get(ctx, &policyValidation)
// 	if err != nil {
// 		logger.Error("Policy validation failed", "error", err)
// 		state.CurrentStatus = "VALIDATION_FAILED"
// 		setRecoverableWorkflowError(state, "VALIDATING_POLICY", "Policy validation failed", err, workflow.Now(ctx))
// 		return nil
// 	}
// 	clearRecoverableWorkflowError(state)

// 	// Cache maturity date in workflow state for IR_4 validation (no more DB reads needed)
// 	state.MaturityDate = policyValidation.MaturityDate
// 	logger.Info("Policy validated, maturity date cached", "maturity_date", state.MaturityDate)

// 	// 📝 Update state: CREATING_REQUEST
// 	state.CurrentStatus = "CREATING_REQUEST"
// 	logger.Info("Creating revival request in database")

// 	// ACTIVITY 1: Create revival request in database (with workflow IDs)
// 	// Note: MissingDocumentsList not stored at indexing - sent by data entry/QC/approver only
// 	revivalReq := RevivalRequestInput{
// 		TicketID:      input.TicketID,
// 		PolicyNumber:  input.PolicyNumber,
// 		RequestType:   input.RequestType,
// 		CurrentStatus: "INDEXED",
// 		IndexedBy:     input.IndexedBy,
// 		IndexedDate:   input.IndexedDate,
// 		WorkflowID:    workflowID,
// 		RunID:         runID,
// 		Documents:     input.Documents,
// 	}

// 	var createdRequestID string
// 	err = workflow.ExecuteActivity(activityCtx, "CreateRevivalRequestActivity", revivalReq).Get(ctx, &createdRequestID)
// 	if err != nil {
// 		logger.Error("Failed to create revival request", "error", err)
// 		state.CurrentStatus = "FAILED_TO_CREATE"
// 		setRecoverableWorkflowError(state, "CREATING_REQUEST", "Unable to create revival request", err, workflow.Now(ctx))
// 		return nil
// 	}
// 	clearRecoverableWorkflowError(state)

// 	// Update state with created request ID
// 	state.RequestID = createdRequestID
// 	logger.Info("Revival request created successfully", "request_id", createdRequestID)

// 	// 📝 Update state: INDEXED
// 	state.CurrentStatus = "INDEXED"

// 	// 📝 Initialize channels for signals (created once, reused in loop)
// 	dataEntryChannel := workflow.GetSignalChannel(ctx, "data-entry-complete")
// 	qcChannel := workflow.GetSignalChannel(ctx, "quality-check-complete")
// 	approvalChannel := workflow.GetSignalChannel(ctx, "approval-decision")

// 	// 🔄 STATE MACHINE LOOP: Allows natural transitions between stages
// 	// Loop continues until final approval/rejection decision is made
// 	var suspenseResult SuspenseAdjustmentResult // Declared here to be accessible after loop
// 	var dataEntrySignal DataEntryCompleteSignal // Store last data entry signal for suspense calculation

// 	// Start with data entry stage
// 	state.CurrentStatus = "WAITING_FOR_DATA_ENTRY"

// 	// Main workflow loop - processes data entry → QC → approval with natural state transitions
// WorkflowLoop:
// 	for {
// 		switch state.CurrentStatus {
// 		case "WAITING_FOR_DATA_ENTRY":
// 			clearRecoverableWorkflowError(state)
// 			logger.Info("Waiting for data entry completion")

// 			// Wait for data entry signal (blocking call)
// 			dataEntryChannel.Receive(ctx, &dataEntrySignal)

// 			logger.Info("Data entry signal received", "entered_by", dataEntrySignal.EnteredBy)

// 			// 🚫 NEGATIVE PATH: Data entry can reject and return to indexer
// 			if dataEntrySignal.ReturnToIndexer {
// 				logger.Info("Data entry rejected - returning to indexer", "reason", dataEntrySignal.ReturnReason)

// 				terminateErr := workflow.ExecuteActivity(
// 					activityCtx,
// 					"TerminateAndReturnToIndexerActivity",
// 					state.RequestID,
// 					dataEntrySignal.ReturnReason,
// 					"DATA_ENTRY",
// 				).Get(ctx, nil)

// 				if terminateErr != nil {
// 					logger.Error("Failed to return to indexer", "error", terminateErr)
// 					setRecoverableWorkflowError(state, "WAITING_FOR_DATA_ENTRY", "Unable to return request to indexer from data entry", terminateErr, workflow.Now(ctx))
// 					state.CurrentStatus = "WAITING_FOR_DATA_ENTRY"
// 					continue WorkflowLoop
// 				}

// 				state.CurrentStatus = "RETURNED_TO_INDEXER"
// 				logger.Info("Workflow closed - returned to indexer by data entry")
// 				return nil // Workflow terminates
// 			}

// 			// ⏱️ IR_4 Validation: Check maturity date constraint before processing data entry
// 			if maturityErr := validateMaturityDateConstraint(
// 				state.MaturityDate,
// 				dataEntrySignal.NumberOfInstallments,
// 				workflow.Now(ctx),
// 			); maturityErr != nil {
// 				logger.Error("Maturity date validation failed at data entry", "error", maturityErr)

// 				terminateErr := workflow.ExecuteActivity(
// 					activityCtx,
// 					"TerminateAndReturnToIndexerActivity",
// 					state.RequestID,
// 					fmt.Sprintf("Maturity date constraint violated during data entry: %v", maturityErr),
// 					"DATA_ENTRY",
// 				).Get(ctx, nil)

// 				if terminateErr != nil {
// 					logger.Error("Failed to terminate workflow", "error", terminateErr)
// 					setRecoverableWorkflowError(state, "WAITING_FOR_DATA_ENTRY", "Unable to return request to indexer after maturity validation failure", terminateErr, workflow.Now(ctx))
// 					state.CurrentStatus = "WAITING_FOR_DATA_ENTRY"
// 					continue WorkflowLoop
// 				}

// 				state.CurrentStatus = "RETURNED_TO_INDEXER"
// 				logger.Info("Workflow terminated - returned to indexer for maturity date violation")
// 				return nil
// 			}

// 			// 📝 State transition: WAITING_FOR_DATA_ENTRY to PROCESSING_DATA_ENTRY
// 			state.CurrentStatus = "PROCESSING_DATA_ENTRY"

// 			// ACTIVITY: Update revival request with data entry details
// 			dataEntryInput := DataEntryInput{
// 				DataEnteredBy:        dataEntrySignal.EnteredBy,
// 				RevivalType:          dataEntrySignal.RevivalType,
// 				NumberOfInstallments: dataEntrySignal.NumberOfInstallments,
// 				RevivalAmount:        dataEntrySignal.RevivalAmount,
// 				InstallmentAmount:    dataEntrySignal.InstallmentAmount,
// 				DataEntryTimestamp:   dataEntrySignal.EnteredAt,
// 				MissingDocuments:     dataEntrySignal.MissingDocuments,
// 				Documents:            dataEntrySignal.Documents,
// 				SGST:                 dataEntrySignal.SGST,
// 				CGST:                 dataEntrySignal.CGST,
// 				Interest:             dataEntrySignal.Interest,
// 				MedicalExaminerCode:  dataEntrySignal.MedicalExaminerCode,
// 				MedicalExaminerName:  dataEntrySignal.MedicalExaminerName,
// 			}

// 			// Update state with installment info and data entry date for future validations
// 			state.NumberOfInstallments = dataEntrySignal.NumberOfInstallments
// 			state.DataEntryDate = &dataEntrySignal.EnteredAt

// 			err = workflow.ExecuteActivity(activityCtx, "UpdateDataEntryActivity",
// 				state.RequestID,
// 				dataEntryInput,
// 			).Get(ctx, nil)
// 			if err != nil {
// 				logger.Error("Failed to update data entry", "error", err)
// 				setRecoverableWorkflowError(state, "WAITING_FOR_DATA_ENTRY", "Unable to save data entry details", err, workflow.Now(ctx))
// 				state.CurrentStatus = "WAITING_FOR_DATA_ENTRY"
// 				continue WorkflowLoop
// 			}

// 			// 🎯 CHECK FOR PREVIOUS SUSPENSE (Re-revival scenario)
// 			err = workflow.ExecuteActivity(activityCtx, "CheckAndAdjustSuspenseActivity",
// 				state.RequestID,
// 				state.PolicyNumber,
// 				dataEntrySignal.RevivalAmount,
// 			).Get(ctx, &suspenseResult)
// 			if err != nil {
// 				logger.Error("Failed to check and adjust suspense", "error", err)
// 				setRecoverableWorkflowError(state, "WAITING_FOR_DATA_ENTRY", "Unable to process suspense adjustment", err, workflow.Now(ctx))
// 				state.CurrentStatus = "WAITING_FOR_DATA_ENTRY"
// 				continue WorkflowLoop
// 			}
// 			clearRecoverableWorkflowError(state)

// 			// Log suspense adjustment result
// 			if suspenseResult.HasSuspense {
// 				logger.Info("Suspense adjustment applied",
// 					"previous_suspense", suspenseResult.TotalSuspenseAmount,
// 					"adjusted_revival_amount", suspenseResult.AdjustedRevivalAmount)
// 			} else {
// 				logger.Info("No previous suspense found - normal revival flow")
// 			}

// 			// 📝 State transition: Move to DATA_ENTRY_COMPLETE
// 			state.CurrentStatus = "DATA_ENTRY_COMPLETE"
// 			logger.Info("Data entry completed and updated in database", "installments", dataEntrySignal.NumberOfInstallments)
// 			continue WorkflowLoop

// 		case "DATA_ENTRY_COMPLETE":
// 			// Automatic transition to QC stage
// 			state.CurrentStatus = "WAITING_FOR_QC"
// 			logger.Info("Moving to quality check stage")
// 			continue WorkflowLoop

// 		case "WAITING_FOR_QC":
// 			clearRecoverableWorkflowError(state)
// 			logger.Info("Waiting for quality check completion")

// 			// Wait for QC signal (blocking call)
// 			var qcSignal QualityCheckCompleteSignal
// 			qcChannel.Receive(ctx, &qcSignal)

// 			logger.Info("Quality check signal received", "qc_passed", qcSignal.QCPassed, "performed_by", qcSignal.PerformedBy)

// 			// 🚫 NEGATIVE PATH: QC can reject and return to indexer
// 			if qcSignal.ReturnToIndexer {
// 				logger.Info("QC rejected - returning to indexer", "reason", qcSignal.ReturnReason)

// 				terminateErr := workflow.ExecuteActivity(
// 					activityCtx,
// 					"TerminateAndReturnToIndexerActivity",
// 					state.RequestID,
// 					qcSignal.ReturnReason,
// 					"QC",
// 				).Get(ctx, nil)

// 				if terminateErr != nil {
// 					logger.Error("Failed to return to indexer", "error", terminateErr)
// 					setRecoverableWorkflowError(state, "WAITING_FOR_QC", "Unable to return request to indexer from quality check", terminateErr, workflow.Now(ctx))
// 					state.CurrentStatus = "WAITING_FOR_QC"
// 					continue WorkflowLoop
// 				}

// 				state.CurrentStatus = "RETURNED_TO_INDEXER"
// 				logger.Info("Workflow closed - returned to indexer by QC")
// 				return nil
// 			}

// 			// ⏱️ IR_4 Validation: Check maturity date constraint before processing QC
// 			if maturityErr := validateMaturityDateConstraint(
// 				state.MaturityDate,
// 				state.NumberOfInstallments,
// 				workflow.Now(ctx),
// 			); maturityErr != nil {
// 				logger.Error("Maturity date validation failed at QC", "error", maturityErr)

// 				terminateErr := workflow.ExecuteActivity(
// 					activityCtx,
// 					"TerminateAndReturnToIndexerActivity",
// 					state.RequestID,
// 					fmt.Sprintf("Maturity date constraint violated during QC: %v", maturityErr),
// 					"QC",
// 				).Get(ctx, nil)

// 				if terminateErr != nil {
// 					logger.Error("Failed to terminate workflow", "error", terminateErr)
// 					setRecoverableWorkflowError(state, "WAITING_FOR_QC", "Unable to return request to indexer after QC maturity validation failure", terminateErr, workflow.Now(ctx))
// 					state.CurrentStatus = "WAITING_FOR_QC"
// 					continue WorkflowLoop
// 				}

// 				state.CurrentStatus = "RETURNED_TO_INDEXER"
// 				logger.Info("Workflow terminated at QC - returned to indexer for maturity date violation")
// 				return nil
// 			}

// 			// 📝 State transition: WAITING_FOR_QC to PROCESSING_QC
// 			state.CurrentStatus = "PROCESSING_QC"

// 			// ACTIVITY: Update revival request with QC details
// 			err = workflow.ExecuteActivity(activityCtx, "UpdateQCActivity",
// 				state.RequestID,
// 				qcSignal.PerformedBy,
// 				qcSignal.QCComments,
// 				qcSignal.QCPassed,
// 				qcSignal.MissingDocuments,
// 			).Get(ctx, nil)
// 			if err != nil {
// 				logger.Error("Failed to update QC", "error", err)
// 				setRecoverableWorkflowError(state, "WAITING_FOR_QC", "Unable to save quality check details", err, workflow.Now(ctx))
// 				state.CurrentStatus = "WAITING_FOR_QC"
// 				continue WorkflowLoop
// 			}
// 			clearRecoverableWorkflowError(state)

// 			// State transition based on QC result
// 			if !qcSignal.QCPassed {
// 				// 🔄 QC failed - loop back to data entry for rework
// 				state.CurrentStatus = "WAITING_FOR_DATA_ENTRY"
// 				logger.Info("QC failed - returning to data entry for rework")
// 				continue WorkflowLoop
// 			}

// 			// QC passed - move to approval pending
// 			state.CurrentStatus = "APPROVAL_PENDING"
// 			logger.Info("QC passed - moving to approval stage")
// 			continue WorkflowLoop

// 		case "APPROVAL_PENDING":
// 			clearRecoverableWorkflowError(state)
// 			logger.Info("Waiting for approval decision")

// 			// Wait for approval signal (blocking call)
// 			var approvalSignal ApprovalDecisionSignal
// 			approvalChannel.Receive(ctx, &approvalSignal)

// 			logger.Info("Approval decision received", "approved", approvalSignal.Approved, "approved_by", approvalSignal.ApprovedBy)

// 			// 🚫 NEGATIVE PATH: Approver can reject and return to indexer
// 			if approvalSignal.ReturnToIndexer {
// 				logger.Info("Approver rejected - returning to indexer", "reason", approvalSignal.ReturnReason)

// 				terminateErr := workflow.ExecuteActivity(
// 					activityCtx,
// 					"TerminateAndReturnToIndexerActivity",
// 					state.RequestID,
// 					approvalSignal.ReturnReason,
// 					"APPROVAL",
// 				).Get(ctx, nil)

// 				if terminateErr != nil {
// 					logger.Error("Failed to return to indexer", "error", terminateErr)
// 					setRecoverableWorkflowError(state, "APPROVAL_PENDING", "Unable to return request to indexer from approval", terminateErr, workflow.Now(ctx))
// 					state.CurrentStatus = "APPROVAL_PENDING"
// 					continue WorkflowLoop
// 				}

// 				state.CurrentStatus = "RETURNED_TO_INDEXER"
// 				logger.Info("Workflow closed - returned to indexer by approver")
// 				return nil
// 			}

// 			// ⏱️ IR_4 Validation: Check maturity date constraint before processing approval
// 			if maturityErr := validateMaturityDateConstraint(
// 				state.MaturityDate,
// 				state.NumberOfInstallments,
// 				workflow.Now(ctx),
// 			); maturityErr != nil {
// 				logger.Error("Maturity date validation failed at approval", "error", maturityErr)

// 				terminateErr := workflow.ExecuteActivity(
// 					activityCtx,
// 					"TerminateAndReturnToIndexerActivity",
// 					state.RequestID,
// 					fmt.Sprintf("Maturity date constraint violated during approval: %v", maturityErr),
// 					"APPROVAL",
// 				).Get(ctx, nil)

// 				if terminateErr != nil {
// 					logger.Error("Failed to terminate workflow", "error", terminateErr)
// 					setRecoverableWorkflowError(state, "APPROVAL_PENDING", "Unable to return request to indexer after approval maturity validation failure", terminateErr, workflow.Now(ctx))
// 					state.CurrentStatus = "APPROVAL_PENDING"
// 					continue WorkflowLoop
// 				}

// 				state.CurrentStatus = "RETURNED_TO_INDEXER"
// 				logger.Info("Workflow terminated at approval - returned to indexer for maturity date violation")
// 				return nil
// 			}

// 			// 🔄 REDIRECT PATH: Approver can redirect to earlier stage for rework
// 			if approvalSignal.RedirectToStage != "" {
// 				logger.Info("Approval redirected to earlier stage",
// 					"redirect_to", approvalSignal.RedirectToStage,
// 					"performed_by", approvalSignal.ApprovedBy,
// 					"comments", approvalSignal.RedirectComments)

// 				if approvalSignal.RedirectToStage == "DATA_ENTRY" {
// 					err = workflow.ExecuteActivity(
// 						activityCtx,
// 						"UpdateRevivalStatusActivity",
// 						state.RequestID,
// 						"DATA_ENTRY_PENDING",
// 					).Get(ctx, nil)
// 					if err != nil {
// 						logger.Error("Failed to update status for DATA_ENTRY redirect", "error", err)
// 						setRecoverableWorkflowError(state, "APPROVAL_PENDING", "Unable to redirect request to data entry", err, workflow.Now(ctx))
// 						state.CurrentStatus = "APPROVAL_PENDING"
// 						continue WorkflowLoop
// 					}

// 					// 🔄 Redirect to data entry - loop naturally continues
// 					state.CurrentStatus = "WAITING_FOR_DATA_ENTRY"
// 					logger.Info("Redirected to data entry - workflow loops back to data entry stage")
// 					continue WorkflowLoop
// 				} else if approvalSignal.RedirectToStage == "QC" {
// 					err = workflow.ExecuteActivity(
// 						activityCtx,
// 						"UpdateRevivalStatusActivity",
// 						state.RequestID,
// 						"DATA_ENTRY_COMPLETE",
// 					).Get(ctx, nil)
// 					if err != nil {
// 						logger.Error("Failed to update status for QC redirect", "error", err)
// 						setRecoverableWorkflowError(state, "APPROVAL_PENDING", "Unable to redirect request to quality check", err, workflow.Now(ctx))
// 						state.CurrentStatus = "APPROVAL_PENDING"
// 						continue WorkflowLoop
// 					}

// 					// 🔄 Redirect to QC - loop naturally continues
// 					state.CurrentStatus = "WAITING_FOR_QC"
// 					logger.Info("Redirected to QC - workflow loops back to QC stage")
// 					continue WorkflowLoop
// 				}
// 			}

// 			// ✅ FINAL DECISION: Approval or rejection - break loop and continue to collection/termination
// 			if !approvalSignal.Approved {
// 				err = workflow.ExecuteActivity(
// 					activityCtx,
// 					"UpdateRevivalStatusActivity",
// 					state.RequestID,
// 					"REJECTED",
// 				).Get(ctx, nil)
// 				if err != nil {
// 					logger.Error("Failed to persist rejection status", "error", err)
// 					setRecoverableWorkflowError(state, "APPROVAL_PENDING", "Unable to save rejection decision", err, workflow.Now(ctx))
// 					state.CurrentStatus = "APPROVAL_PENDING"
// 					continue WorkflowLoop
// 				}

// 				// Request rejected - workflow terminates
// 				state.CurrentStatus = "REJECTED"
// 				state.CompletedAt = timePtr(workflow.Now(ctx))
// 				logger.Info("Revival request rejected", "rejected_by", approvalSignal.ApprovedBy, "comments", approvalSignal.Comments)
// 				return nil
// 			}

// 			// ✅ APPROVED: Break loop and continue to collection workflow
// 			logger.Info("Revival request approved - proceeding to collection stage")

// 			// �️ MONTH CHANGE VALIDATION: Check if data entry month has changed
// 			if state.DataEntryDate != nil {
// 				currentTime := workflow.Now(ctx)
// 				dataEntryMonth := state.DataEntryDate.Month()
// 				dataEntryYear := state.DataEntryDate.Year()
// 				currentMonth := currentTime.Month()
// 				currentYear := currentTime.Year()

// 				// Check if month or year has changed
// 				if dataEntryMonth != currentMonth || dataEntryYear != currentYear {
// 					logger.Warn("Data entry month has changed - redirecting to data entry",
// 						"data_entry_date", state.DataEntryDate,
// 						"current_date", currentTime,
// 						"data_entry_month", dataEntryMonth,
// 						"current_month", currentMonth)

// 					// Update database status to DATA_ENTRY_PENDING
// 					err = workflow.ExecuteActivity(
// 						activityCtx,
// 						"UpdateRevivalStatusActivity",
// 						state.RequestID,
// 						"DATA_ENTRY_PENDING",
// 					).Get(ctx, nil)
// 					if err != nil {
// 						logger.Error("Failed to update status to DATA_ENTRY_PENDING", "error", err)
// 						setRecoverableWorkflowError(state, "APPROVAL_PENDING", "Unable to redirect request to data entry for month change", err, workflow.Now(ctx))
// 						state.CurrentStatus = "APPROVAL_PENDING"
// 						continue WorkflowLoop
// 					}

// 					// Redirect back to data entry
// 					state.CurrentStatus = "DATA_ENTRY_PENDING"
// 					logger.Info("Redirected to data entry due to month change - revival calculation needs to be updated")
// 					continue WorkflowLoop
// 				}
// 			}

// 			// �📝 State transition: APPROVAL_PENDING to PROCESSING_APPROVAL
// 			state.CurrentStatus = "PROCESSING_APPROVAL"

// 			// Calculate SLA dates
// 			slaStartDate := approvalSignal.ApprovedAt
// 			slaEndDate := slaStartDate.Add(60 * 24 * time.Hour) // 60 days

// 			err = workflow.ExecuteActivity(
// 				activityCtx,
// 				"UpdateWorkflowStateActivity",
// 				state.RequestID,
// 				"APPROVED",
// 				slaStartDate,
// 				slaEndDate,
// 			).Get(ctx, nil)

// 			if err != nil {
// 				logger.Error("Failed to persist SLA dates", "error", err)
// 				setRecoverableWorkflowError(state, "APPROVAL_PENDING", "Unable to persist approval SLA details", err, workflow.Now(ctx))
// 				state.CurrentStatus = "APPROVAL_PENDING"
// 				continue WorkflowLoop
// 			}

// 			// ACTIVITY: Update revival request with approval details
// 			err = workflow.ExecuteActivity(activityCtx, "UpdateApprovalActivity",
// 				state.TicketID,
// 				approvalSignal.ApprovedBy,
// 				approvalSignal.Comments,
// 				slaStartDate,
// 				slaEndDate,
// 			).Get(ctx, nil)
// 			if err != nil {
// 				logger.Error("Failed to update approval", "error", err)
// 				setRecoverableWorkflowError(state, "APPROVAL_PENDING", "Unable to save approval details", err, workflow.Now(ctx))
// 				state.CurrentStatus = "APPROVAL_PENDING"
// 				continue WorkflowLoop
// 			}
// 			clearRecoverableWorkflowError(state)

// 			// State transition: PROCESSING_APPROVAL to APPROVED
// 			state.CurrentStatus = "APPROVED"
// 			state.SLAStartDate = &slaStartDate
// 			state.SLAEndDate = &slaEndDate
// 			logger.Info("Approval updated in database", "sla_end_date", slaEndDate)

// 			// ✅ Approval processing complete - break loop and continue to collection
// 			break WorkflowLoop
// 		}
// 	}

// 	// 📝 POST-APPROVAL: Start collection workflow (no more backwards state transitions allowed)
// 	logger.Info("Starting collection phase")

// 	// Start SLA timer and first collection child workflow
// 	slaTimer := workflow.NewTimer(ctx, 60*24*time.Hour)
// 	firstCollectionChannel := workflow.GetSignalChannel(ctx, "first-collection-complete")

// 	selector := workflow.NewSelector(ctx)

// 	// Track which event occurred
// 	var eventType string
// 	var childWorkflowFuture workflow.ChildWorkflowFuture

// 	// Wait for EITHER first collection OR SLA timeout
// 	selector.AddFuture(slaTimer, func(f workflow.Future) {
// 		// 60-day SLA expired - TERMINATE (IR_10)
// 		eventType = "SLA_TIMEOUT"
// 		state.CurrentStatus = "TERMINATED"
// 		state.SLAExpired = true
// 		state.CompletedAt = timePtr(workflow.Now(ctx))

// 		// Execute termination activity
// 		activityCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
// 			StartToCloseTimeout: time.Minute,
// 		})
// 		workflow.ExecuteActivity(activityCtx, "TerminateRevivalActivity", state.RequestID, "60-day SLA expired").Get(ctx, nil)
// 	})

// 	noPendingInstallments := false

// 	selector.AddReceive(firstCollectionChannel, func(c workflow.ReceiveChannel, more bool) {
// 		var signal FirstCollectionCompleteSignal
// 		c.Receive(ctx, &signal)

// 		logger.Info("📥 RECEIVED first-collection-complete signal",
// 			"request_id", state.RequestID,
// 			"ticket_id", state.TicketID,
// 			"collection_date", signal.CollectionDate,
// 			"payment_mode", signal.PaymentMode,
// 			"total_amount", signal.TotalAmount)

// 		eventType = "FIRST_COLLECTION"

// 		// First collection completed
// 		state.CurrentStatus = "ACTIVE"
// 		state.FirstCollectionDone = true
// 		state.InstallmentsPaid = 1

// 		logger.Info("About to start InstallmentMonitorWorkflow as child",
// 			"request_id", state.RequestID,
// 			"number_of_installments", state.NumberOfInstallments,
// 			"has_suspense", suspenseResult.HasSuspense)

// 		// Start InstallmentMonitorWorkflow as child
// 		// CRITICAL: ParentClosePolicy set to ABANDON so child continues running even after parent completes!
// 		childOptions := workflow.ChildWorkflowOptions{
// 			WorkflowID:        fmt.Sprintf("installment-monitor-%s", state.RequestID),
// 			ParentClosePolicy: enums.PARENT_CLOSE_POLICY_ABANDON, // Child workflow continues independently
// 		}

// 		logger.Info("Child workflow options configured",
// 			"child_workflow_id", fmt.Sprintf("installment-monitor-%s", state.RequestID),
// 			"parent_close_policy", "ABANDON")

// 		//TODO: add 1st day of the month calculation for next due date
// 		childCtx := workflow.WithChildOptions(ctx, childOptions)
// 		// childWorkflowFuture = workflow.ExecuteChildWorkflow(childCtx, InstallmentMonitorWorkflow, InstallmentMonitorInput{
// 		// 	RequestID:            state.RequestID,
// 		// 	NextDueDate:          slaStartDate.Add(30 * 24 * time.Hour), // 1 month from approval
// 		// 	NumberOfInstallments: state.NumberOfInstallments,            // Pass actual installment count
// 		// })

// 		// ---------------- FIX STARTS HERE ----------------

// 		// Default: normal revival flow
// 		pendingInstallments := state.NumberOfInstallments

// 		// If this is a re-revival and suspense adjustment happened,
// 		// reduce installments by what is already covered by suspense
// 		if suspenseResult.HasSuspense {
// 			paidInstallmentsFromSuspense :=
// 				int(suspenseResult.TotalSuspenseAmount / dataEntrySignal.InstallmentAmount)
// 				//assuming that all installments are equal

// 			pendingInstallments = state.NumberOfInstallments - paidInstallmentsFromSuspense

// 			currentPendingInstallments := pendingInstallments - 1

// 			if currentPendingInstallments < 1 {
// 				noPendingInstallments = true
// 				return // DO NOT start InstallmentMonitorWorkflow
// 			}

// 			logger.Info("Re-revival detected – adjusting installment workflow",
// 				"total_installments", state.NumberOfInstallments,
// 				"suspense_amount", suspenseResult.TotalSuspenseAmount,
// 				"installment_amount", dataEntrySignal.InstallmentAmount,
// 				"paid_from_suspense", paidInstallmentsFromSuspense,
// 				"pending_installments", pendingInstallments)
// 		}

// 		// Start InstallmentMonitorWorkflow as child

// 		firstOfNextMonth := time.Date(
// 			state.SLAStartDate.Year(),
// 			state.SLAStartDate.Month()+1,
// 			1,
// 			0, 0, 0, 0,
// 			state.SLAStartDate.Location(),
// 		)

// 		logger.Info("🚀 Executing child workflow now",
// 			"request_id", state.RequestID,
// 			"next_due_date", firstOfNextMonth,
// 			"pending_installments", pendingInstallments)

// 		childWorkflowFuture = workflow.ExecuteChildWorkflow(
// 			childCtx,
// 			InstallmentMonitorWorkflow,
// 			InstallmentMonitorInput{
// 				RequestID: state.RequestID,
// 				// NextDueDate:          slaStartDate.Add(30 * 24 * time.Hour),
// 				NextDueDate:          firstOfNextMonth,
// 				NumberOfInstallments: pendingInstallments, // ✅ FIXED
// 			},
// 		)

// 		logger.Info("✅ Child workflow ExecuteChildWorkflow called",
// 			"request_id", state.RequestID,
// 			"child_workflow_future_nil", childWorkflowFuture == nil)

// 		// ---------------- FIX ENDS HERE ----------------

// 	})

// 	selector.Select(ctx)

// 	logger.Info("Selector completed - processing event",
// 		"event_type", eventType,
// 		"child_workflow_future_nil", childWorkflowFuture == nil,
// 		"no_pending_installments", noPendingInstallments)

// 	if noPendingInstallments {
// 		logger.Info("No pending installments – revival completes after first collection",
// 			"request_id", state.RequestID)

// 		err := workflow.ExecuteActivity(activityCtx,
// 			"FinalizeRevivalAfterFirstCollection",
// 			state.RequestID,
// 		).Get(ctx, nil)
// 		if err != nil {
// 			setRecoverableWorkflowError(state, "ACTIVE", "Unable to finalize revival after first collection", err, workflow.Now(ctx))
// 			logger.Error("Failed to finalize revival after first collection", "error", err, "request_id", state.RequestID)
// 			return nil
// 		}
// 		clearRecoverableWorkflowError(state)
// 	}

// 	// If first collection occurred, wait for child workflow to start before completing parent
// 	if eventType == "FIRST_COLLECTION" && childWorkflowFuture != nil {
// 		logger.Info("Waiting for InstallmentMonitorWorkflow child to start",
// 			"request_id", state.RequestID)

// 		var childExec workflow.Execution
// 		err := childWorkflowFuture.GetChildWorkflowExecution().Get(ctx, &childExec)
// 		if err != nil {
// 			logger.Error("❌ Failed to start InstallmentMonitorWorkflow",
// 				"error", err,
// 				"request_id", state.RequestID)
// 			setRecoverableWorkflowError(state, "ACTIVE", "Unable to start installment monitoring workflow", err, workflow.Now(ctx))
// 			return nil
// 		}
// 		clearRecoverableWorkflowError(state)
// 		logger.Info("✅ InstallmentMonitorWorkflow started successfully",
// 			"child_workflow_id", childExec.ID,
// 			"child_run_id", childExec.RunID,
// 			"request_id", state.RequestID)
// 	} else if eventType == "FIRST_COLLECTION" {
// 		logger.Warn("⚠️ First collection occurred but childWorkflowFuture is nil!",
// 			"request_id", state.RequestID)
// 	} else {
// 		logger.Info("Event was not FIRST_COLLECTION",
// 			"event_type", eventType,
// 			"request_id", state.RequestID)
// 	}

// 	logger.Info("RevivalWorkflow completing",
// 		"request_id", state.RequestID,
// 		"final_status", state.CurrentStatus)

// 	return nil
// }

// // FirstCollectionWorkflow handles the first installment collection (IR_36: Dual Collection)
// // Lifespan: Minutes (cash/online) or Up to 30 days (cheque)
// func FirstCollectionWorkflow(ctx workflow.Context, input FirstCollectionInput) error {
// 	logger := workflow.GetLogger(ctx)

// 	// Set activity options
// 	activityCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
// 		StartToCloseTimeout: 30 * time.Second,
// 		RetryPolicy: &temporal.RetryPolicy{
// 			InitialInterval:    1 * time.Second,
// 			BackoffCoefficient: 2.0,
// 			MaximumInterval:    30 * time.Second,
// 			MaximumAttempts:    3,
// 		},
// 	})

// 	// Validate dual collection: Premium + Installment
// 	err := workflow.ExecuteActivity(activityCtx, "ValidateDualCollectionActivity", input).Get(ctx, nil)
// 	if err != nil {
// 		logger.Error("First collection validation failed", "request_id", input.RequestID, "error", err)
// 		return nil
// 	}

// 	// Process payment based on mode
// 	if input.PaymentMode == "CASH" || input.PaymentMode == "NEFT" || input.PaymentMode == "RTGS" || input.PaymentMode == "UPI" || input.PaymentMode == "CARD" {
// 		// Immediate completion
// 		err = workflow.ExecuteActivity(activityCtx, "ProcessDualPaymentActivity", input, "COMPLETED").Get(ctx, nil)
// 		if err != nil {
// 			logger.Error("Failed to process immediate dual payment", "request_id", input.RequestID, "payment_mode", input.PaymentMode, "error", err)
// 			return nil
// 		}
// 	} else if input.PaymentMode == "CHEQUE" {
// 		// Cheque - start monitoring
// 		var chequeID string
// 		err = workflow.ExecuteActivity(activityCtx, "CreateChequeRecordActivity", input).Get(ctx, &chequeID)
// 		if err != nil {
// 			logger.Error("Failed to create cheque record", "request_id", input.RequestID, "error", err)
// 			return nil
// 		}

// 		// Start ChequeMonitorWorkflow
// 		childOptions := workflow.ChildWorkflowOptions{
// 			WorkflowID: fmt.Sprintf("cheque-monitor-%s", input.RequestID),
// 		}

// 		childCtx := workflow.WithChildOptions(ctx, childOptions)
// 		workflow.ExecuteChildWorkflow(childCtx, ChequeMonitorWorkflow, ChequeMonitorInput{
// 			RequestID: input.RequestID,
// 			ChequeID:  chequeID,
// 			Amount:    input.TotalAmount,
// 		})

// 		// Wait for clearance before returning
// 		chequeChannel := workflow.GetSignalChannel(ctx, "cheque-cleared")
// 		var chequeSignal ChequeClearedSignal

// 		// Wait for cheque clearance signal (blocking call)
// 		chequeChannel.Receive(ctx, &chequeSignal)

// 		// Process cleared cheque
// 		err = workflow.ExecuteActivity(activityCtx, "ProcessDualPaymentActivity", input, "COMPLETED").Get(ctx, nil)
// 		if err != nil {
// 			logger.Error("Failed to process cleared cheque payment", "request_id", input.RequestID, "error", err)
// 			return nil
// 		}
// 	}

// 	return nil
// }

// // ChequeMonitorWorkflow monitors cheque clearance status
// // Lifespan: Up to 30 days
// func ChequeMonitorWorkflow(ctx workflow.Context, input ChequeMonitorInput) error {
// 	// Wait for clearance or dishonor
// 	selector := workflow.NewSelector(ctx)

// 	clearedChannel := workflow.GetSignalChannel(ctx, "cheque-cleared")
// 	dishonoredChannel := workflow.GetSignalChannel(ctx, "cheque-dishonored")

// 	// Set timeout to next due date (typically 30 days)
// 	nextDue := workflow.Now(ctx).Add(30 * 24 * time.Hour)
// 	timeout := workflow.NewTimer(ctx, nextDue.Sub(workflow.Now(ctx)))

// 	selector.AddReceive(clearedChannel, func(c workflow.ReceiveChannel, more bool) {
// 		var signal ChequeClearedSignal
// 		c.Receive(ctx, &signal)
// 		// Cheque cleared - complete collection
// 	})

// 	selector.AddReceive(dishonoredChannel, func(c workflow.ReceiveChannel, more bool) {
// 		var signal ChequeDishonoredSignal
// 		c.Receive(ctx, &signal)
// 		// Cheque dishonored - move to suspense (IR_28: NO first collection suspense reversal)
// 	})

// 	selector.AddFuture(timeout, func(f workflow.Future) {
// 		// Timeout - cheque not cleared by due date
// 	})

// 	selector.Select(ctx)

// 	return nil
// }

// // InstallmentMonitorWorkflow monitors subsequent installment payments
// // Lifespan: Months (until all installments paid or default)
// func InstallmentMonitorWorkflow(ctx workflow.Context, input InstallmentMonitorInput) error {
// 	logger := workflow.GetLogger(ctx)
// 	totalInstallments := input.NumberOfInstallments // Use actual installment count from input

// 	// 🎯 WORKFLOW STATE: Track which installment we're expecting next
// 	// This prevents duplicates and out-of-order payments
// 	currentExpectedInstallment := 2 // Start with installment 2 (first was collected before workflow started)

// 	workflowInfo := workflow.GetInfo(ctx)
// 	logger.Info("InstallmentMonitorWorkflow started",
// 		"workflow_id", workflowInfo.WorkflowExecution.ID,
// 		"run_id", workflowInfo.WorkflowExecution.RunID,
// 		"request_id", input.RequestID,
// 		"total_installments", totalInstallments,
// 		"current_expected_installment", currentExpectedInstallment,
// 		"next_due_date", input.NextDueDate)

// 	// Set activity options
// 	activityCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
// 		StartToCloseTimeout: 30 * time.Second,
// 		RetryPolicy: &temporal.RetryPolicy{
// 			InitialInterval:    1 * time.Second,
// 			BackoffCoefficient: 2.0,
// 			MaximumInterval:    30 * time.Second,
// 			MaximumAttempts:    3,
// 		},
// 	})

// 	// 🎯 Track if default occurred to terminate workflow
// 	defaultOccurred := false

// 	for installmentNumber := 2; installmentNumber <= totalInstallments; installmentNumber++ {
// 		// Calculate next due date (IR_11: 1st of next month)
// 		// For 2nd+ installments, due on 1st of month
// 		var nextDue time.Time
// 		if installmentNumber == 2 {
// 			nextDue = input.NextDueDate
// 		} else {
// 			// Calculate 1st of subsequent month
// 			// Fix: installmentNumber-2 because installment 2 is at baseDate (offset 0)
// 			baseDate := input.NextDueDate
// 			monthsToAdd := installmentNumber - 2 // Installment 3 = +1 month, 4 = +2 months, etc.
// 			nextDue = time.Date(
// 				baseDate.Year(),
// 				baseDate.Month()+time.Month(monthsToAdd),
// 				1,
// 				0, 0, 0, 0,
// 				baseDate.Location(),
// 			)

// 			//for testing *******
// 			//nextDue = workflow.Now(ctx).Add(3 * time.Minute)
// 			//******************
// 		}

// 		logger.Info("Monitoring installment",
// 			"installment_number", installmentNumber,
// 			"due_date", nextDue,
// 			"current_time", workflow.Now(ctx))

// 		// Calculate timer duration with validation to prevent negative durations
// 		timerDuration := nextDue.Add(24 * time.Hour).Sub(workflow.Now(ctx))
// 		//for testing *******
// 		timerDuration = nextDue.Sub(workflow.Now(ctx))

// 		//************************
// 		if timerDuration < 0 {
// 			logger.Warn("Due date is in the past, waiting indefinitely for payment signal (testing mode)",
// 				"installment_number", installmentNumber,
// 				"due_date", nextDue,
// 				"current_time", workflow.Now(ctx))
// 			// Use very long timer (1 year) to effectively wait indefinitely for signal
// 			// This prevents timeout firing immediately during testing with past dates
// 			timerDuration = 365 * 24 * time.Hour // Wait ~1 year (effectively indefinite for testing)
// 		}

// 		// Set timer for due date + 1 day (IR_9: Zero grace period)
// 		dueTimer := workflow.NewTimer(ctx, timerDuration)
// 		paymentChannel := workflow.GetSignalChannel(ctx, fmt.Sprintf("installment-payment-received-%d", installmentNumber))

// 		selector := workflow.NewSelector(ctx)

// 		selector.AddReceive(paymentChannel, func(c workflow.ReceiveChannel, more bool) {
// 			// Payment received - process installment
// 			var signal InstallmentPaymentSignal
// 			c.Receive(ctx, &signal)

// 			logger.Info("Payment signal received",
// 				"expected_installment", currentExpectedInstallment,
// 				"received_installment", installmentNumber,
// 				"amount", signal.Amount,
// 				"payment_mode", signal.PaymentMode)

// 			// 🎯 CRITICAL VALIDATION: Ensure installment is paid in sequence
// 			// This should always match because we're listening on the correct channel,
// 			// but this is an extra safety check
// 			if installmentNumber != currentExpectedInstallment {
// 				logger.Error("SEQUENCE VIOLATION: Received installment out of order",
// 					"expected", currentExpectedInstallment,
// 					"received", installmentNumber,
// 					"request_id", input.RequestID)
// 				// This shouldn't happen due to channel design, but log for debugging
// 			}

// 			err := workflow.ExecuteActivity(activityCtx, "ProcessInstallmentActivity", input.RequestID, installmentNumber, signal.Amount, signal.PaymentMode, "PAID", signal.PaymentDate).Get(ctx, nil)
// 			if err != nil {
// 				logger.Error("Failed to process installment payment",
// 					"installment_number", installmentNumber,
// 					"error", err)
// 			} else {
// 				logger.Info("Installment payment processed successfully",
// 					"installment_number", installmentNumber,
// 					"remaining_installments", totalInstallments-installmentNumber)

// 				// 🎯 INCREMENT EXPECTED INSTALLMENT: Move to next installment
// 				currentExpectedInstallment++
// 				logger.Info("Moved to next expected installment",
// 					"current_expected", currentExpectedInstallment,
// 					"total_installments", totalInstallments)
// 			}

// 			// Check if all installments paid
// 			if installmentNumber == totalInstallments {
// 				logger.Info("All installments paid - workflow completing",
// 					"total_installments", totalInstallments)
// 			}
// 		})

// 		selector.AddFuture(dueTimer, func(f workflow.Future) {
// 			// 🚨 TIMEOUT - No grace period (IR_9)
// 			// Payment not received by due date → DEFAULT
// 			logger.Error("Installment payment timeout - moving to default",
// 				"installment_number", installmentNumber,
// 				"due_date", nextDue,
// 				"request_id", input.RequestID,
// 				"policy_number", "unknown") // Would need to pass from input

// 			// 🎯 CRITICAL: Call HandleDefaultActivity
// 			// This will:
// 			// 1. Create suspense entries for premium + all paid installments
// 			// 2. Create termination record in database
// 			// 3. Revert policy to AL (lapsed)
// 			// 4. Update revival request status to DEFAULTED
// 			err := workflow.ExecuteActivity(activityCtx, "HandleDefaultActivity", input.RequestID, installmentNumber).Get(ctx, nil)
// 			if err != nil {
// 				logger.Error("Failed to handle installment default",
// 					"installment_number", installmentNumber,
// 					"request_id", input.RequestID,
// 					"error", err)
// 				// Even if activity fails, we should terminate workflow
// 				// to prevent further processing
// 			} else {
// 				logger.Info("Default handled successfully - termination record created",
// 					"installment_number", installmentNumber,
// 					"request_id", input.RequestID)
// 			}

// 			// 🎯 SET FLAG: Default occurred - workflow will terminate after selector
// 			defaultOccurred = true
// 		})

// 		logger.Info("Waiting for payment or timeout", "installment_number", installmentNumber)
// 		selector.Select(ctx)
// 		logger.Info("Selector completed for installment", "installment_number", installmentNumber)

// 		// 🎯 CRITICAL: If default occurred, TERMINATE workflow immediately
// 		if defaultOccurred {
// 			logger.Error("Revival workflow terminated due to installment default",
// 				"installment_number", installmentNumber,
// 				"request_id", input.RequestID,
// 				"total_installments", totalInstallments)
// 			return nil
// 		}
// 	}

// 	logger.Info("InstallmentMonitorWorkflow completed",
// 		"request_id", input.RequestID,
// 		"total_installments", totalInstallments)
// 	return nil
// }

// // SLATimerWorkflow manages 60-day SLA countdown
// func SLATimerWorkflow(ctx workflow.Context, slaEnd time.Time) error {
// 	remaining := slaEnd.Sub(workflow.Now(ctx))
// 	if remaining > 0 {
// 		// Wait for SLA expiration
// 		workflow.NewTimer(ctx, remaining).Get(ctx, nil)

// 		// SLA expired - notify parent
// 		// SLA expired - parent workflow will handle termination
// 	}

// 	return nil
// }

// // Workflow state structures
// // Note: PolicyValidationResult is defined in activities.go
// type RevivalWorkflowState struct {
// 	RequestID            string     `json:"request_id"`
// 	TicketID             string     `json:"ticket_id"`
// 	PolicyNumber         string     `json:"policy_number"`
// 	CurrentStatus        string     `json:"current_status"`
// 	LastErrorMessage     string     `json:"last_error_message,omitempty"`
// 	LastErrorStage       string     `json:"last_error_stage,omitempty"`
// 	LastErrorAt          *time.Time `json:"last_error_at,omitempty"`
// 	RecoverableError     bool       `json:"recoverable_error"`
// 	ErrorCount           int        `json:"error_count"`
// 	StartedAt            time.Time  `json:"started_at"`
// 	NumberOfInstallments int        `json:"number_of_installments"` // Added for maturity validation
// 	MaturityDate         time.Time  `json:"maturity_date"`          // Cached from ValidatePolicyActivity - no DB reads for IR_4
// 	DataEntryDate        *time.Time `json:"data_entry_date"`        // Cached from DataEntrySignal - for month change validation
// 	SLAStartDate         *time.Time `json:"sla_start_date"`
// 	SLAEndDate           *time.Time `json:"sla_end_date"`
// 	FirstCollectionDone  bool       `json:"first_collection_done"`
// 	InstallmentsPaid     int        `json:"installments_paid"`
// 	SLAExpired           bool       `json:"sla_expired"`
// 	CompletedAt          *time.Time `json:"completed_at"`
// }

// // RevivalRequestInput for CreateRevivalRequestActivity
// // Note: MissingDocumentsList is NOT part of indexing - sent by data entry/QC/approver only
// type RevivalRequestInput struct {
// 	TicketID      string    `json:"ticket_id"`
// 	PolicyNumber  string    `json:"policy_number"`
// 	RequestType   string    `json:"request_type"`
// 	CurrentStatus string    `json:"current_status"`
// 	IndexedBy     string    `json:"indexed_by"`
// 	IndexedDate   time.Time `json:"indexed_date"`
// 	WorkflowID    string    `json:"workflow_id"`
// 	RunID         string    `json:"run_id"`
// 	Documents     string    `json:"documents"` // JSONB as string
// }

// // Signal structures
// type DataEntryCompleteSignal struct {
// 	EnteredBy            string    `json:"entered_by"`
// 	EnteredAt            time.Time `json:"entered_at"`
// 	RevivalType          string    `json:"revival_type"`
// 	NumberOfInstallments int       `json:"number_of_installments"`
// 	RevivalAmount        float64   `json:"revival_amount"`
// 	InstallmentAmount    float64   `json:"installment_amount"`
// 	MissingDocuments     string    `json:"missing_documents,omitempty"` // JSON string for missing docs saved to DB
// 	Documents            string    `json:"documents,omitempty"`         // JSON string for documents submitted saved to DB
// 	Interest             float64   `json:"interest"`
// 	SGST                 float64   `json:"sgst"`
// 	CGST                 float64   `json:"cgst"`
// 	ReturnToIndexer      bool      `json:"return_to_indexer,omitempty"` // If true, reject and return to indexer
// 	ReturnReason         string    `json:"return_reason,omitempty"`     // Reason for returning to indexer
// 	MedicalExaminerCode  string    `json:"medical_examiner_code,omitempty"`
// 	MedicalExaminerName  string    `json:"medical_examiner_name,omitempty"`
// }

// type QualityCheckCompleteSignal struct {
// 	QCPassed         bool      `json:"qc_passed"`
// 	QCComments       string    `json:"qc_comments"`
// 	PerformedBy      string    `json:"performed_by"`
// 	PerformedAt      time.Time `json:"performed_at"`
// 	MissingDocuments string    `json:"missing_documents,omitempty"` // JSON string for missing docs saved to DB
// 	ReturnToIndexer  bool      `json:"return_to_indexer,omitempty"` // If true, reject and return to indexer
// 	ReturnReason     string    `json:"return_reason,omitempty"`     // Reason for returning to indexer
// }

// type ApprovalDecisionSignal struct {
// 	Approved             bool      `json:"approved"`
// 	Comments             string    `json:"comments"`
// 	ApprovedBy           string    `json:"approved_by"`
// 	ApprovedAt           time.Time `json:"approved_at"`
// 	ReturnToIndexer      bool      `json:"return_to_indexer,omitempty"`      // If true, reject and return to indexer
// 	ReturnReason         string    `json:"return_reason,omitempty"`          // Reason for returning to indexer
// 	MissingDocumentsList []string  `json:"missing_documents_list,omitempty"` // Missing docs when returning
// 	RedirectToStage      string    `json:"redirect_to_stage,omitempty"`      // "DATA_ENTRY" or "QC" for rework
// 	RedirectComments     string    `json:"redirect_comments,omitempty"`      // Comments when redirecting
// }

// type FirstCollectionCompleteSignal struct {
// 	CollectionDate time.Time `json:"collection_date"`
// 	PaymentMode    string    `json:"payment_mode"`
// 	TotalAmount    float64   `json:"total_amount"`
// }

// type ChequeClearedSignal struct {
// 	ClearedAt time.Time `json:"cleared_at"`
// 	BankName  string    `json:"bank_name"`
// }

// type ChequeDishonoredSignal struct {
// 	DishonoredAt time.Time `json:"dishonored_at"`
// 	Reason       string    `json:"reason"`
// }

// type InstallmentPaymentSignal struct {
// 	PaymentDate time.Time `json:"payment_date"`
// 	Amount      float64   `json:"amount"`
// 	PaymentMode string    `json:"payment_mode"`
// }

// // Workflow input structures - using types from activities.go
// type ChequeMonitorInput struct {
// 	RequestID string  `json:"request_id"`
// 	ChequeID  string  `json:"cheque_id"`
// 	Amount    float64 `json:"amount"`
// }

// type InstallmentMonitorInput struct {
// 	RequestID            string    `json:"request_id"`
// 	NextDueDate          time.Time `json:"next_due_date"`
// 	NumberOfInstallments int       `json:"number_of_installments"` // Actual number of installments requested
// 	StartInstallmentNo   int       `json:"start_installment_no"`   // For future use if needed
// }

// // Helper functions
// func timePtr(t time.Time) *time.Time {
// 	return &t
// }

// func setRecoverableWorkflowError(state *RevivalWorkflowState, stage, userMessage string, err error, now time.Time) {
// 	state.LastErrorStage = stage
// 	state.RecoverableError = true
// 	state.LastErrorAt = &now
// 	state.ErrorCount++

// 	if err != nil {
// 		state.LastErrorMessage = fmt.Sprintf("%s: %v", userMessage, err)
// 		return
// 	}

// 	state.LastErrorMessage = userMessage
// }

// func clearRecoverableWorkflowError(state *RevivalWorkflowState) {
// 	state.LastErrorMessage = ""
// 	state.LastErrorStage = ""
// 	state.LastErrorAt = nil
// 	state.RecoverableError = false
// }

// // validateMaturityDateConstraint validates IR_4 compliance inline (no DB read)
// // Returns error if constraint violated
// func validateMaturityDateConstraint(maturityDate time.Time, numberOfInstallments int, firstDueDate time.Time) error {
// 	// Calculate last installment due date
// 	// First installment: firstDueDate
// 	// Subsequent installments: 1st of next months (IR_11)
// 	lastDueDate := firstDueDate.AddDate(0, numberOfInstallments-1, 0)

// 	// For 2+ installments, adjust to 1st of the month
// 	if numberOfInstallments > 1 {
// 		lastDueDate = time.Date(
// 			lastDueDate.Year(),
// 			lastDueDate.Month(),
// 			1, 0, 0, 0, 0,
// 			lastDueDate.Location(),
// 		)
// 	}

// 	// Check IR_4 constraint: last installment must not fall after maturity
// 	if lastDueDate.After(maturityDate) {
// 		return fmt.Errorf(
// 			"IR_4 violation: last installment due (%s) falls after maturity date (%s)",
// 			lastDueDate.Format("2006-01-02"),
// 			maturityDate.Format("2006-01-02"),
// 		)
// 	}

// 	// Check IR_4 constraint: last installment must not fall in maturity month
// 	if lastDueDate.Year() == maturityDate.Year() &&
// 		lastDueDate.Month() == maturityDate.Month() {
// 		return fmt.Errorf(
// 			"IR_4 violation: last installment due (%s) falls in maturity month (%s)",
// 			lastDueDate.Format("2006-01-02"),
// 			maturityDate.Format("2006-01"),
// 		)
// 	}

// 	return nil
// }

// // BatchInstallmentInput contains data for batch installment processing
// type BatchInstallmentInput struct {
// 	RequestID           string                   `json:"request_id"`
// 	PolicyNumber        string                   `json:"policy_number"`
// 	TicketID            string                   `json:"ticket_id"`
// 	Installments        []InstallmentPaymentData `json:"installments"`
// 	AtomicMode          bool                     `json:"atomic_mode"`            // If true, rollback all on any failure
// 	MonitorWorkflowID   string                   `json:"monitor_workflow_id"`    // InstallmentMonitorWorkflow ID to signal
// 	SignalMonitorOnPaid bool                     `json:"signal_monitor_on_paid"` // If true, signal monitor workflow after each payment
// }

// // InstallmentPaymentData represents a single installment payment
// type InstallmentPaymentData struct {
// 	InstallmentNumber int     `json:"installment_number"`
// 	Amount            float64 `json:"amount"`
// 	PaymentMode       string  `json:"payment_mode"`
// }

// // BatchInstallmentResult contains the result of batch processing
// type BatchInstallmentResult struct {
// 	TotalSubmitted int                           `json:"total_submitted"`
// 	Successful     int                           `json:"successful"`
// 	Failed         int                           `json:"failed"`
// 	Results        []InstallmentProcessingResult `json:"results"`
// }

// // InstallmentProcessingResult represents result of a single installment
// type InstallmentProcessingResult struct {
// 	InstallmentNumber int    `json:"installment_number"`
// 	Success           bool   `json:"success"`
// 	PaymentID         string `json:"payment_id,omitempty"`
// 	ErrorMessage      string `json:"error_message,omitempty"`
// }

// // BatchInstallmentProcessingWorkflow processes multiple installments sequentially
// // This ensures proper ordering and atomic batch operations
// func BatchInstallmentProcessingWorkflow(ctx workflow.Context, input BatchInstallmentInput) (*BatchInstallmentResult, error) {
// 	logger := workflow.GetLogger(ctx)

// 	logger.Info("BatchInstallmentProcessingWorkflow started",
// 		"request_id", input.RequestID,
// 		"ticket_id", input.TicketID,
// 		"num_installments", len(input.Installments),
// 		"atomic_mode", input.AtomicMode)

// 	result := &BatchInstallmentResult{
// 		TotalSubmitted: len(input.Installments),
// 		Results:        make([]InstallmentProcessingResult, 0, len(input.Installments)),
// 	}

// 	// 🎯 CRITICAL FIX: Only signal InstallmentMonitorWorkflow, don't directly process payments
// 	// The monitor workflow will execute ProcessInstallmentActivity to avoid duplicates

// 	// Process each installment by signaling the monitor workflow
// 	for i, inst := range input.Installments {
// 		logger.Info("Signaling installment in batch",
// 			"sequence", i+1,
// 			"installment_number", inst.InstallmentNumber,
// 			"amount", inst.Amount)

// 		instResult := InstallmentProcessingResult{
// 			InstallmentNumber: inst.InstallmentNumber,
// 		}

// 		// Signal InstallmentMonitorWorkflow to process this installment
// 		if input.MonitorWorkflowID != "" {
// 			signalName := fmt.Sprintf("installment-payment-received-%d", inst.InstallmentNumber)
// 			signal := InstallmentPaymentSignal{
// 				PaymentDate: workflow.Now(ctx),
// 				Amount:      inst.Amount,
// 				PaymentMode: inst.PaymentMode,
// 			}

// 			logger.Info("Signaling InstallmentMonitorWorkflow from batch",
// 				"monitor_workflow_id", input.MonitorWorkflowID,
// 				"signal_name", signalName,
// 				"installment_number", inst.InstallmentNumber)

// 			// Signal the external workflow
// 			err := workflow.SignalExternalWorkflow(ctx, input.MonitorWorkflowID, "", signalName, signal).Get(ctx, nil)
// 			if err != nil {
// 				logger.Error("Failed to signal InstallmentMonitorWorkflow",
// 					"monitor_workflow_id", input.MonitorWorkflowID,
// 					"installment_number", inst.InstallmentNumber,
// 					"error", err)

// 				instResult.Success = false
// 				instResult.ErrorMessage = fmt.Sprintf("Failed to signal: %v", err)
// 				result.Failed++
// 				result.Results = append(result.Results, instResult)

// 				// In atomic mode, fail entire batch
// 				if input.AtomicMode {
// 					logger.Error("Atomic mode: Failing entire batch due to signaling error",
// 						"failed_installment", inst.InstallmentNumber,
// 						"error", err)
// 					return nil, fmt.Errorf("batch failed to signal installment %d: %w", inst.InstallmentNumber, err)
// 				}

// 				// Non-atomic: continue to next installment
// 				continue
// 			}

// 			instResult.Success = true
// 			result.Successful++
// 			result.Results = append(result.Results, instResult)

// 			logger.Info("Successfully signaled InstallmentMonitorWorkflow from batch",
// 				"monitor_workflow_id", input.MonitorWorkflowID,
// 				"installment_number", inst.InstallmentNumber,
// 				"successful_so_far", result.Successful)

// 			// Add a small delay between signals to allow monitor workflow to process sequentially
// 			workflow.Sleep(ctx, 500*time.Millisecond)
// 		} else {
// 			logger.Error("MonitorWorkflowID is empty, cannot signal",
// 				"installment_number", inst.InstallmentNumber)
// 			instResult.Success = false
// 			instResult.ErrorMessage = "MonitorWorkflowID is empty"
// 			result.Failed++
// 			result.Results = append(result.Results, instResult)

// 			if input.AtomicMode {
// 				return nil, fmt.Errorf("batch failed: MonitorWorkflowID is empty")
// 			}
// 		}
// 	}

// 	logger.Info("BatchInstallmentProcessingWorkflow completed",
// 		"total_submitted", result.TotalSubmitted,
// 		"successful", result.Successful,
// 		"failed", result.Failed)

// 	return result, nil
// }
