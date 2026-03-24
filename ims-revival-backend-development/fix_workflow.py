with open('workflow/revival_workflow.go', 'w', encoding='utf-8') as f:
    content = '''package workflow

import (
\t"fmt"
\t"time"

\t"go.temporal.io/sdk/workflow"
)

// InstallmentRevivalWorkflow is the main workflow orchestrating the complete revival process
// Lifespan: 6-12 months
func InstallmentRevivalWorkflow(ctx workflow.Context, requestID string) error {
\t// Workflow state
\tstate := &RevivalWorkflowState{
\t\tRequestID:     requestID,
\t\tCurrentStatus: "INDEXED",
\t\tStartedAt:     workflow.Now(ctx),
\t}

\t// Stage 1: Wait for data entry completion
\tdataEntryChannel := workflow.GetSignalChannel(ctx, "data-entry-complete")
\tvar dataEntrySignal DataEntryCompleteSignal

\t// Note: Activity registration happens in the worker
\t_ = state

\t// Wait for data entry
\tselector := workflow.NewSelector(ctx)

\tdataEntryChannel.Receive(ctx, &dataEntrySignal)

\tselector.Select(ctx)

\t// State transition: INDEXED to DATA_ENTRY_COMPLETE
\tstate.CurrentStatus = "DATA_ENTRY_COMPLETE"

\t// Stage 2: Wait for quality check
\tqcChannel := workflow.GetSignalChannel(ctx, "quality-check-complete")
\tvar qcSignal QualityCheckCompleteSignal

\tselector = workflow.NewSelector(ctx)
\tqcChannel.Receive(ctx, &qcSignal)
\tselector.Select(ctx)

\t// State transition: DATA_ENTRY_COMPLETE to QC_COMPLETE
\tif !qcSignal.QCPassed {
\t\t// Return to data entry
\t\tstate.CurrentStatus = "DATA_ENTRY_PENDING"
\t\treturn nil
\t}

\tstate.CurrentStatus = "QC_COMPLETE"

\t// Stage 3: Wait for approval decision
\tapprovalChannel := workflow.GetSignalChannel(ctx, "approval-decision")
\tvar approvalSignal ApprovalDecisionSignal

\tselector = workflow.NewSelector(ctx)
\tapprovalChannel.Receive(ctx, &approvalSignal)
\tselector.Select(ctx)

\tif !approvalSignal.Approved {
\t\t// Request rejected
\t\tstate.CurrentStatus = "WITHDRAWN"
\t\tstate.CompletedAt = timePtr(workflow.Now(ctx))
\t\treturn nil
\t}

\t// State transition: QC_COMPLETE to APPROVED
\tstate.CurrentStatus = "APPROVED"
\tslaStartDate := workflow.Now(ctx)
\tslaEndDate := slaStartDate.Add(60 * 24 * time.Hour) // 60 days
\tstate.SLAStartDate = &slaStartDate
\tstate.SLAEndDate = &slaEndDate

\t// Start SLA timer and first collection child workflow
\tslaTimer := workflow.NewTimer(ctx, 60*24*time.Hour)
\tfirstCollectionChannel := workflow.GetSignalChannel(ctx, "first-collection-complete")

\tselector = workflow.NewSelector(ctx)

\t// Wait for EITHER first collection OR SLA timeout
\tselector.AddFuture(slaTimer, func(f workflow.Future) {
\t\t// 60-day SLA expired - TERMINATE (IR_10)
\t\tstate.CurrentStatus = "TERMINATED"
\t\tstate.SLAExpired = true
\t\tstate.CompletedAt = timePtr(workflow.Now(ctx))

\t\t// Execute termination activity
\t\tactivityCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
\t\t\tStartToCloseTimeout: time.Minute,
\t\t})
\t\tworkflow.ExecuteActivity(activityCtx, "TerminateRevivalActivity", requestID, "60-day SLA expired").Get(ctx, nil)
\t})

\tselector.AddReceive(firstCollectionChannel, func(c workflow.ReceiveChannel, more bool) {
\t\tvar signal FirstCollectionCompleteSignal
\t\tc.Receive(ctx, &signal)

\t\t// First collection completed
\t\tstate.CurrentStatus = "ACTIVE"
\t\tstate.FirstCollectionDone = true
\t\tstate.InstallmentsPaid = 1

\t\t// Start InstallmentMonitorWorkflow as child
\t\tchildOptions := workflow.ChildWorkflowOptions{
\t\t\tWorkflowID: fmt.Sprintf("installment-monitor-%s", requestID),
\t\t}

\t\tchildCtx := workflow.WithChildOptions(ctx, childOptions)
\t\tworkflow.ExecuteChildWorkflow(childCtx, InstallmentMonitorWorkflow, InstallmentMonitorInput{
\t\t\tRequestID:   requestID,
\t\t\tNextDueDate: slaStartDate.Add(30 * 24 * time.Hour), // 1 month from approval
\t\t})
\t})

\tselector.Select(ctx)

\treturn nil
}

// FirstCollectionWorkflow handles the first installment collection (IR_36: Dual Collection)
// Lifespan: Minutes (cash/online) or Up to 30 days (cheque)
func FirstCollectionWorkflow(ctx workflow.Context, input FirstCollectionInput) error {
\t// Set activity options
\tactivityCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
\t\tStartToCloseTimeout: time.Minute * 5,
\t})

\t// Validate dual collection: Premium + Installment
\terr := workflow.ExecuteActivity(activityCtx, "ValidateDualCollectionActivity", input).Get(ctx, nil)
\tif err != nil {
\t\treturn err
\t}

\t// Process payment based on mode
\tif input.PaymentMode == "CASH" || input.PaymentMode == "NEFT" || input.PaymentMode == "RTGS" || input.PaymentMode == "UPI" || input.PaymentMode == "CARD" {
\t\t// Immediate completion
\t\terr = workflow.ExecuteActivity(activityCtx, "ProcessDualPaymentActivity", input, "COMPLETED").Get(ctx, nil)
\t\tif err != nil {
\t\t\treturn err
\t\t}
\t} else if input.PaymentMode == "CHEQUE" {
\t\t// Cheque - start monitoring
\t\tvar chequeID string
\t\terr = workflow.ExecuteActivity(activityCtx, "CreateChequeRecordActivity", input).Get(ctx, &chequeID)
\t\tif err != nil {
\t\t\treturn err
\t\t}

\t\t// Start ChequeMonitorWorkflow
\t\tchildOptions := workflow.ChildWorkflowOptions{
\t\t\tWorkflowID: fmt.Sprintf("cheque-monitor-%s", input.RequestID),
\t\t}

\t\tchildCtx := workflow.WithChildOptions(ctx, childOptions)
\t\tworkflow.ExecuteChildWorkflow(childCtx, ChequeMonitorWorkflow, ChequeMonitorInput{
\t\t\tRequestID: input.RequestID,
\t\t\tChequeID:  chequeID,
\t\t\tAmount:    input.TotalAmount,
\t\t})

\t\t// Wait for clearance before returning
\t\tchequeChannel := workflow.GetSignalChannel(ctx, "cheque-cleared")
\t\tvar chequeSignal ChequeClearedSignal

\t\tselector := workflow.NewSelector(ctx)
\t\tchequeChannel.Receive(ctx, &chequeSignal)
\t\tselector.Select(ctx)

\t\t// Process cleared cheque
\t\terr = workflow.ExecuteActivity(activityCtx, "ProcessDualPaymentActivity", input, "COMPLETED").Get(ctx, nil)
\t\tif err != nil {
\t\t\treturn err
\t\t}
\t}

\treturn nil
}

// ChequeMonitorWorkflow monitors cheque clearance status
// Lifespan: Up to 30 days
func ChequeMonitorWorkflow(ctx workflow.Context, input ChequeMonitorInput) error {
\t// Wait for clearance or dishonor
\tselector := workflow.NewSelector(ctx)

\tclearedChannel := workflow.GetSignalChannel(ctx, "cheque-cleared")
\tdishonoredChannel := workflow.GetSignalChannel(ctx, "cheque-dishonored")

\t// Set timeout to next due date (typically 30 days)
\tnextDue := workflow.Now(ctx).Add(30 * 24 * time.Hour)
\ttimeout := workflow.NewTimer(ctx, nextDue.Sub(workflow.Now(ctx)))

\tselector.AddReceive(clearedChannel, func(c workflow.ReceiveChannel, more bool) {
\t\tvar signal ChequeClearedSignal
\t\tc.Receive(ctx, &signal)
\t\t// Cheque cleared - complete collection
\t})

\tselector.AddReceive(dishonoredChannel, func(c workflow.ReceiveChannel, more bool) {
\t\tvar signal ChequeDishonoredSignal
\t\tc.Receive(ctx, &signal)
\t\t// Cheque dishonored - move to suspense (IR_28: NO first collection suspense reversal)
\t})

\tselector.AddFuture(timeout, func(f workflow.Future) {
\t\t// Timeout - cheque not cleared by due date
\t})

\tselector.Select(ctx)

\treturn nil
}

// InstallmentMonitorWorkflow monitors subsequent installment payments
// Lifespan: Months (until all installments paid or default)
func InstallmentMonitorWorkflow(ctx workflow.Context, input InstallmentMonitorInput) error {
\ttotalInstallments := 12 // Default; would query from revival request in production

\t// Set activity options
\tactivityCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
\t\tStartToCloseTimeout: time.Minute * 5,
\t})

\tfor installmentNumber := 2; installmentNumber <= totalInstallments; installmentNumber++ {
\t\t// Calculate next due date (IR_11: 1st of next month)
\t\t// For 2nd+ installments, due on 1st of month
\t\tvar nextDue time.Time
\t\tif installmentNumber == 2 {
\t\t\tnextDue = input.NextDueDate
\t\t} else {
\t\t\t// Calculate 1st of subsequent month
\t\t\tbaseDate := input.NextDueDate
\t\t\tnextDue = time.Date(
\t\t\t\tbaseDate.Year(),
\t\t\t\tbaseDate.Month()+time.Month(installmentNumber-1),
\t\t\t\t1,
\t\t\t\t0, 0, 0, 0,
\t\t\t\tbaseDate.Location(),
\t\t\t)
\t\t}

\t\t// Set timer for due date + 1 day (IR_9: Zero grace period)
\t\tdueTimer := workflow.NewTimer(ctx, nextDue.Add(24*time.Hour).Sub(workflow.Now(ctx)))
\t\tpaymentChannel := workflow.GetSignalChannel(ctx, fmt.Sprintf("installment-payment-received-%d", installmentNumber))

\t\tselector := workflow.NewSelector(ctx)

\t\tselector.AddReceive(paymentChannel, func(c workflow.ReceiveChannel, more bool) {
\t\t\t// Payment received - process installment
\t\t\tvar signal InstallmentPaymentSignal
\t\t\tc.Receive(ctx, &signal)

\t\t\tworkflow.ExecuteActivity(activityCtx, "ProcessInstallmentActivity", input.RequestID, installmentNumber, signal.Amount, signal.PaymentMode, "PAID").Get(ctx, nil)

\t\t\t// Check if all installments paid
\t\t\tif installmentNumber == totalInstallments {
\t\t\t\t// All installments paid - workflow completion handled by parent
\t\t\t}
\t\t})

\t\tselector.AddFuture(dueTimer, func(f workflow.Future) {
\t\t\t// TIMEOUT - No grace period (IR_9)
\t\t\t// Move to suspense + revert to AL
\t\t\tworkflow.ExecuteActivity(activityCtx, "HandleDefaultActivity", input.RequestID, installmentNumber).Get(ctx, nil)

\t\t\t// Default handled - workflow continues to monitor remaining installments
\t\t})

\t\tselector.Select(ctx)
\t}

\treturn nil
}

// SLATimerWorkflow manages 60-day SLA countdown
func SLATimerWorkflow(ctx workflow.Context, slaEnd time.Time) error {
\tremaining := slaEnd.Sub(workflow.Now(ctx))
\tif remaining > 0 {
\t\t// Wait for SLA expiration
\t\tworkflow.NewTimer(ctx, remaining).Get(ctx, nil)

\t\t// SLA expired - notify parent
\t\t// SLA expired - parent workflow will handle termination
\t}

\treturn nil
}

// Workflow state structures
type RevivalWorkflowState struct {
\tRequestID           string     `json:"request_id"`
\tCurrentStatus       string     `json:"current_status"`
\tStartedAt           time.Time  `json:"started_at"`
\tSLAStartDate        *time.Time `json:"sla_start_date"`
\tSLAEndDate          *time.Time `json:"sla_end_date"`
\tFirstCollectionDone bool       `json:"first_collection_done"`
\tInstallmentsPaid    int        `json:"installments_paid"`
\tSLAExpired          bool       `json:"sla_expired"`
\tCompletedAt         *time.Time `json:"completed_at"`
}

// Signal structures
type DataEntryCompleteSignal struct {
\tEnteredBy   string    `json:"entered_by"`
\tEnteredAt   time.Time `json:"entered_at"`
\tMissingDocs []string  `json:"missing_docs,omitempty"`
}

type QualityCheckCompleteSignal struct {
\tQCPassed    bool      `json:"qc_passed"`
\tQCComments  string    `json:"qc_comments"`
\tPerformedBy string    `json:"performed_by"`
\tPerformedAt time.Time `json:"performed_at"`
\tMissingDocs []string  `json:"missing_docs,omitempty"`
}

type ApprovalDecisionSignal struct {
\tApproved    bool      `json:"approved"`
\tComments    string    `json:"comments"`
\tApprovedBy  string    `json:"approved_by"`
\tApprovedAt  time.Time `json:"approved_at"`
\tMissingDocs []string  `json:"missing_docs,omitempty"`
}

type FirstCollectionCompleteSignal struct {
\tCollectionDate time.Time `json:"collection_date"`
\tPaymentMode    string    `json:"payment_mode"`
\tTotalAmount    float64   `json:"total_amount"`
}

type ChequeClearedSignal struct {
\tClearedAt time.Time `json:"cleared_at"`
\tBankName  string    `json:"bank_name"`
}

type ChequeDishonoredSignal struct {
\tDishonoredAt time.Time `json:"dishonored_at"`
\tReason       string    `json:"reason"`
}

type InstallmentPaymentSignal struct {
\tPaymentDate time.Time `json:"payment_date"`
\tAmount      float64   `json:"amount"`
\tPaymentMode string    `json:"payment_mode"`
}

// Workflow input structures - using types from activities.go
type ChequeMonitorInput struct {
\tRequestID string  `json:"request_id"`
\tChequeID  string  `json:"cheque_id"`
\tAmount    float64 `json:"amount"`
}

type InstallmentMonitorInput struct {
\tRequestID   string    `json:"request_id"`
\tNextDueDate time.Time `json:"next_due_date"`
}

// Helper functions
func timePtr(t time.Time) *time.Time {
\treturn &t
}
'''
    f.write(content)
    print('File written successfully')
