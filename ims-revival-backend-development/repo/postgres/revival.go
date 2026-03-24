package repo

import (
	"context"
	"fmt"
	"time"

	"plirevival/core/domain"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	config "gitlab.cept.gov.in/it-2.0-common/api-config"
	dblib "gitlab.cept.gov.in/it-2.0-common/n-api-db"
)

type RevivalRepository struct {
	db  *dblib.DB
	cfg *config.Config
}

func NewRevivalRepository(db *dblib.DB, cfg *config.Config) *RevivalRepository {
	return &RevivalRepository{db: db, cfg: cfg}
}

const (
	revivalRequestsTable      = "revival.revival_requests"
	installmentSchedulesTable = "revival.installment_schedules"
	revivalCalculationsTable  = "revival.revival_calculations"
	workflowStateTable        = "revival.revival_request_workflow_state"
	statusChangeHistoryTable  = "revival.status_change_history"
	revivalTerminationsTable  = "revival.revival_terminations"
)

// CreateRevivalRequest creates a new revival request using InsertReturning (single DB call)
// Idempotent: Uses ON CONFLICT to handle workflow retries
func (r *RevivalRepository) CreateRevivalRequest(ctx context.Context, req domain.RevivalRequest) (*domain.RevivalRequest, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
	defer cancel()

	if req.RequestID == "" {
		req.RequestID = uuid.New().String()
	}

	ins := dblib.Psql.Insert(revivalRequestsTable).
		Columns(
			"request_id", "ticket_id", "policy_number", "request_type", "current_status",
			"indexed_date", "indexed_by", "workflow_id", "run_id",
			"number_of_installments", "revival_amount", "installment_amount", "total_tax_on_unpaid",
			"blocking_new_collections", "missing_documents_list", "documents",
			"sgst", "cgst", "interest",
		).
		Values(
			req.RequestID, req.TicketID, req.PolicyNumber, req.RequestType, req.CurrentStatus,
			req.IndexedDate, req.IndexedBy, req.WorkflowID, req.RunID,
			req.NumberOfInstallments, req.RevivalAmount, req.InstallmentAmount, req.TotalTaxOnUnpaid,
			req.BlockingNewCollections, req.MissingDocumentsList, req.Documents,
			req.SGST, req.CGST, req.Interest,
		).
		Suffix(`
			ON CONFLICT (ticket_id) DO UPDATE SET
				workflow_id = EXCLUDED.workflow_id,
				run_id = EXCLUDED.run_id,
				updated_at = NOW()
			RETURNING request_id, ticket_id, policy_number, request_type, current_status,
				indexed_date, indexed_by, data_entry_date, data_entry_by,
				qc_complete_date, qc_by, approval_date, approved_by,
				completion_date, termination_date, withdrawal_date,
				number_of_installments, revival_amount, installment_amount, total_tax_on_unpaid,
				first_collection_date, first_collection_done, blocking_new_collections, installments_paid,
				missing_documents_list, documents,
				created_at, updated_at, workflow_id, run_id,
				sgst, cgst, interest, revival_type
		`)

	// Use InsertReturning to avoid second DB call
	return dblib.InsertReturning(ctx, r.db, ins, pgx.RowToAddrOfStructByNameLax[domain.RevivalRequest])
}

// GetRevivalRequestByID retrieves a revival request by ID
func (r *RevivalRepository) GetRevivalRequestByID(ctx context.Context, requestID string) (domain.RevivalRequest, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	q := dblib.Psql.Select(
		"request_id", "ticket_id", "policy_number", "request_type", "current_status",
		"workflow_id", "run_id",
		"indexed_date", "indexed_by", "data_entry_date", "data_entry_by",
		"qc_complete_date", "qc_by", "qc_comments", "approval_date", "approved_by", "approval_comments",
		"completion_date", "termination_date", "withdrawal_date",
		"number_of_installments", "revival_amount", "installment_amount",
		"total_tax_on_unpaid", "first_collection_date", "first_collection_done",
		"blocking_new_collections", "installments_paid", "request_owner",
		"missing_documents_list", "documents",
		"created_at", "updated_at", "previous_suspense_amount", "suspense_adjusted", "adjusted_revival_amount",
		"sgst", "cgst", "interest", "revival_type",
		"medical_examiner_code", "medical_examiner_name",
	).
		From(revivalRequestsTable).
		Where(sq.Eq{"request_id": requestID})

	return dblib.SelectOne(ctx, r.db, q, pgx.RowToStructByName[domain.RevivalRequest])
}

// GetRevivalRequestByTicketID retrieves a revival request by ticket ID
func (r *RevivalRepository) GetRevivalRequestByTicketID(ctx context.Context, ticketID string) (domain.RevivalRequest, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	q := dblib.Psql.Select(
		"request_id", "ticket_id", "policy_number", "request_type", "current_status",
		"workflow_id", "run_id",
		"indexed_date", "indexed_by", "data_entry_date", "data_entry_by",
		"qc_complete_date", "qc_by", "qc_comments", "approval_date", "approved_by", "approval_comments",
		"completion_date", "termination_date", "withdrawal_date",
		"number_of_installments", "revival_amount", "installment_amount",
		"total_tax_on_unpaid", "first_collection_date", "first_collection_done",
		"blocking_new_collections", "installments_paid", "request_owner",
		"missing_documents_list", "documents",
		"created_at", "updated_at", "previous_suspense_amount", "suspense_adjusted", "adjusted_revival_amount",
		"sgst", "cgst", "interest", "revival_type",
		"medical_examiner_code", "medical_examiner_name",
	).
		From(revivalRequestsTable).
		Where(sq.Eq{"ticket_id": ticketID})

	return dblib.SelectOne(ctx, r.db, q, pgx.RowToStructByName[domain.RevivalRequest])
}

// GetRevivalRequestsByPolicyNumber retrieves all revival requests for a policy
func (r *RevivalRepository) GetRevivalRequestsByPolicyNumber(ctx context.Context, policyNumber string) ([]domain.RevivalRequest, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
	defer cancel()

	q := dblib.Psql.Select(
		"request_id", "ticket_id", "policy_number", "request_type", "current_status",
		"workflow_id", "run_id",
		"indexed_date", "indexed_by", "data_entry_date", "data_entry_by",
		"qc_complete_date", "qc_by", "qc_comments", "approval_date", "approved_by", "approval_comments",
		"completion_date", "termination_date", "withdrawal_date",
		"number_of_installments", "revival_amount", "installment_amount",
		"total_tax_on_unpaid", "first_collection_date", "first_collection_done",
		"blocking_new_collections", "installments_paid", "request_owner",
		"missing_documents_list", "documents",
		"created_at", "updated_at",
		"sgst", "cgst", "interest", "revival_type",
		"medical_examiner_code", "medical_examiner_name",
	).
		From(revivalRequestsTable).
		Where(sq.Eq{"policy_number": policyNumber}).
		OrderBy("indexed_date DESC")

	return dblib.SelectRows(ctx, r.db, q, pgx.RowToStructByName[domain.RevivalRequest])
}

// GetLatestRevivalRequestByPolicyNumber retrieves the most recent revival request for a policy
func (r *RevivalRepository) GetLatestRevivalRequestByPolicyNumber(ctx context.Context, policyNumber string) (domain.RevivalRequest, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	q := dblib.Psql.Select(
		"request_id", "ticket_id", "policy_number", "request_type", "current_status",
		"workflow_id", "run_id",
		"indexed_date", "indexed_by", "data_entry_date", "data_entry_by",
		"qc_complete_date", "qc_by", "qc_comments", "approval_date", "approved_by", "approval_comments",
		"completion_date", "termination_date", "withdrawal_date",
		"number_of_installments", "revival_amount", "installment_amount",
		"total_tax_on_unpaid", "first_collection_date", "first_collection_done",
		"blocking_new_collections", "installments_paid", "request_owner",
		"missing_documents_list", "documents",
		"created_at", "updated_at", "previous_suspense_amount", "suspense_adjusted", "adjusted_revival_amount",
		"sgst", "cgst", "interest", "revival_type",
		"medical_examiner_code", "medical_examiner_name",
	).
		From(revivalRequestsTable).
		Where(sq.Eq{"policy_number": policyNumber}).
		OrderBy("indexed_date DESC").
		Limit(1)

	return dblib.SelectOne(ctx, r.db, q, pgx.RowToStructByName[domain.RevivalRequest])
}

// GetAllRevivalRequests retrieves all revival requests with policy details
func (r *RevivalRepository) GetAllRevivalRequests(ctx context.Context) ([]domain.RevivalRequestWithPolicy, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
	defer cancel()

	q := dblib.Psql.Select(
		"rr.request_id",
		"rr.policy_number",
		"rr.ticket_id",
		"p.customer_name as insured_name",
		"p.customer_id",
		"rr.request_type",
		"rr.current_status",
		"rr.indexed_date as requested_date",
		"rr.created_at",
	).
		From(revivalRequestsTable + " rr").
		LeftJoin("common.policies p ON rr.policy_number = p.policy_number").
		OrderBy("rr.indexed_date DESC")

	return dblib.SelectRows(ctx, r.db, q, pgx.RowToStructByName[domain.RevivalRequestWithPolicy])
}

// UpdateRevivalRequestStatus updates revival request status using pgx.batch (single DB round trip)
func (r *RevivalRepository) UpdateRevivalRequestStatus(ctx context.Context, requestID string, status string, changedBy string) error {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	batch := &pgx.Batch{}

	// Queue update query
	upd := dblib.Psql.Update(revivalRequestsTable).
		Set("current_status", status).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"request_id": requestID})

	if err := dblib.QueueExecRow(batch, upd); err != nil {
		return err
	}

	// Queue status change history insert
	ins := dblib.Psql.Insert(statusChangeHistoryTable).
		Columns("history_id", "request_id", "from_status", "to_status", "changed_at", "changed_by", "change_reason").
		Values(uuid.New().String(), requestID, nil, status, time.Now(), changedBy, nil)

	if err := dblib.QueueExecRow(batch, ins); err != nil {
		return err
	}

	// Send batch to database (single round trip, implicit transaction)
	batchResults := r.db.SendBatch(ctx, batch)
	defer batchResults.Close()

	// Check all batched operations succeeded
	for range batch.Len() {
		_, err := batchResults.Exec()
		if err != nil {
			return err
		}
	}

	return batchResults.Close()
}

// UpdateRevivalRequestForApproval updates revival request on approval using pgx.batch (single DB round trip)
func (r *RevivalRepository) UpdateRevivalRequestForApproval(ctx context.Context, ticketID, approvedBy, comments string, slaStart, slaEnd time.Time) error {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	batch := &pgx.Batch{}

	// Queue revival request update
	upd := dblib.Psql.Update(revivalRequestsTable).
		Set("current_status", "APPROVED").
		Set("approval_date", slaStart).
		Set("approved_by", &approvedBy).
		Set("approval_comments", &comments).
		Set("sla_start_date", slaStart).
		Set("sla_end_date", slaEnd).
		Set("blocking_new_collections", true).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"ticket_id": ticketID})

	if err := dblib.QueueExecRow(batch, upd); err != nil {
		return err
	}

	// Queue status change history insert - get actual current_status from the record
	historySQL := `INSERT INTO revival.status_change_history (history_id, request_id, from_status, to_status, changed_at, changed_by, change_reason)
		SELECT $1, request_id, current_status, $2, $3, $4, $5
		FROM revival.revival_requests
		WHERE ticket_id = $6`

	batch.Queue(historySQL, uuid.New().String(), "APPROVED", time.Now(), approvedBy, &comments, ticketID)

	// TODO: workflow_state table update commented out - table may be deprecated
	// If needed, uncomment and use request_id lookup
	// wfUpd := dblib.Psql.Update(workflowStateTable).
	// 	Set("current_status", "APPROVED").
	// 	Set("sla_start_date", slaStart).
	// 	Set("sla_end_date", slaEnd).
	// 	Set("last_updated", sq.Expr("NOW()")).
	// 	Where(sq.Eq{"request_id": requestID})
	// if err := dblib.QueueExecRow(batch, wfUpd); err != nil {
	// 	return err
	// }

	// Send batch (single round trip)
	batchResults := r.db.SendBatch(ctx, batch)
	defer batchResults.Close()

	for range batch.Len() {
		_, err := batchResults.Exec()
		if err != nil {
			return err
		}
	}

	return batchResults.Close()
}

// UpdateRevivalRequestForDataEntry updates revival request for data entry using pgx.batch (single DB round trip)
func (r *RevivalRepository) UpdateRevivalRequestForDataEntry(ctx context.Context, requestID, dataEntryBy, revivalType string, numberOfInstallments int, revivalAmount, installmentAmount, sgst, cgst, interest float64, documents, missingDocuments, medicalExaminerCode, medicalExaminerName string) error {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	batch := &pgx.Batch{}

	// Queue update query
	upd := dblib.Psql.Update(revivalRequestsTable).
		Set("current_status", "DATA_ENTRY_COMPLETE").
		Set("data_entry_date", sq.Expr("NOW()")).
		Set("data_entry_by", &dataEntryBy).
		Set("revival_type", revivalType).
		Set("number_of_installments", numberOfInstallments).
		Set("revival_amount", revivalAmount).
		Set("installment_amount", installmentAmount).
		Set("sgst", sgst).
		Set("cgst", cgst).
		Set("interest", interest).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"request_id": requestID})

	// Add medical examiner details if provided
	if medicalExaminerCode != "" {
		upd = upd.Set("medical_examiner_code", medicalExaminerCode)
	}
	if medicalExaminerName != "" {
		upd = upd.Set("medical_examiner_name", medicalExaminerName)
	}

	// Add documents if provided (including empty array "[]")
	if documents != "" {
		upd = upd.Set("documents", documents)
	}

	// Add missing_documents_list if provided (including empty array "[]")
	if missingDocuments != "" {
		upd = upd.Set("missing_documents_list", missingDocuments)
	}

	if err := dblib.QueueExecRow(batch, upd); err != nil {
		return err
	}

	// Queue status change history insert
	ins := dblib.Psql.Insert(statusChangeHistoryTable).
		Columns("history_id", "request_id", "from_status", "to_status", "changed_at", "changed_by", "change_reason").
		Values(uuid.New().String(), requestID, "INDEXED", "DATA_ENTRY_COMPLETE", time.Now(), dataEntryBy, nil)

	if err := dblib.QueueExecRow(batch, ins); err != nil {
		return err
	}

	// Send batch (single round trip)
	batchResults := r.db.SendBatch(ctx, batch)
	defer batchResults.Close()

	for range batch.Len() {
		_, err := batchResults.Exec()
		if err != nil {
			return err
		}
	}

	return batchResults.Close()
}

// UpdateRevivalRequestForDataEntryWithPendingDocs updates revival request for data entry but keeps current status unchanged
// This is used when data entry is submitted but there are missing documents that need to be received
// The status remains as-is (INDEXED or DATA_ENTRY_PENDING) and workflow does not proceed
func (r *RevivalRepository) UpdateRevivalRequestForDataEntryWithPendingDocs(ctx context.Context, requestID, currentStatus, dataEntryBy, revivalType string, numberOfInstallments int, revivalAmount, installmentAmount, sgst, cgst, interest float64, documents, missingDocuments, medicalExaminerCode, medicalExaminerName string) error {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	batch := &pgx.Batch{}

	// Queue update query - status remains unchanged (current_status not updated)
	upd := dblib.Psql.Update(revivalRequestsTable).
		Set("data_entry_date", sq.Expr("NOW()")).
		Set("data_entry_by", &dataEntryBy).
		Set("revival_type", revivalType).
		Set("number_of_installments", numberOfInstallments).
		Set("revival_amount", revivalAmount).
		Set("installment_amount", installmentAmount).
		Set("sgst", sgst).
		Set("cgst", cgst).
		Set("interest", interest).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"request_id": requestID})

	// Add medical examiner details if provided
	if medicalExaminerCode != "" {
		upd = upd.Set("medical_examiner_code", medicalExaminerCode)
	}
	if medicalExaminerName != "" {
		upd = upd.Set("medical_examiner_name", medicalExaminerName)
	}

	// Add documents if provided (including empty array "[]")
	if documents != "" {
		upd = upd.Set("documents", documents)
	}

	// Add missing_documents_list if provided (including empty array "[]")
	if missingDocuments != "" {
		upd = upd.Set("missing_documents_list", missingDocuments)
	}

	if err := dblib.QueueExecRow(batch, upd); err != nil {
		return err
	}

	// No status change history needed since status is not changing

	// Send batch (single round trip)
	batchResults := r.db.SendBatch(ctx, batch)
	defer batchResults.Close()

	for range batch.Len() {
		_, err := batchResults.Exec()
		if err != nil {
			return err
		}
	}

	return batchResults.Close()
}

// UpdateRevivalRequestForQC updates revival request for quality check using pgx.batch (single DB round trip)
func (r *RevivalRepository) UpdateRevivalRequestForQC(ctx context.Context, requestID, qcBy, qcComments string, qcPassed bool, missingDocuments string) error {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	nextStatus := "APPROVAL_PENDING"
	if !qcPassed {
		nextStatus = "DATA_ENTRY_PENDING"
	}

	batch := &pgx.Batch{}

	// Queue status change history insert FIRST (before update) - captures current_status as from_status
	historySQL := `INSERT INTO revival.status_change_history (history_id, request_id, from_status, to_status, changed_at, changed_by, change_reason)
		SELECT $1, request_id, current_status, $2, $3, $4, $5
		FROM revival.revival_requests
		WHERE request_id = $6`

	batch.Queue(historySQL, uuid.New().String(), nextStatus, time.Now(), qcBy, &qcComments, requestID)

	// Queue update query AFTER history insert
	upd := dblib.Psql.Update(revivalRequestsTable).
		Set("current_status", nextStatus).
		Set("qc_complete_date", sq.Expr("NOW()")).
		Set("qc_by", &qcBy).
		Set("qc_comments", &qcComments).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"request_id": requestID})

	// Add missing_documents_list if provided
	if missingDocuments != "" {
		upd = upd.Set("missing_documents_list", missingDocuments)
	}

	if err := dblib.QueueExecRow(batch, upd); err != nil {
		return err
	}

	// Send batch (single round trip)
	batchResults := r.db.SendBatch(ctx, batch)
	defer batchResults.Close()

	for range batch.Len() {
		_, err := batchResults.Exec()
		if err != nil {
			return err
		}
	}

	return batchResults.Close()
}

// UpdateRevivalRequestForFirstCollection updates revival request after first collection
func (r *RevivalRepository) UpdateRevivalRequestForFirstCollection(ctx context.Context, requestID string) error {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	batch := &pgx.Batch{}

	// Queue revival request update
	upd := dblib.Psql.Update(revivalRequestsTable).
		Set("current_status", "ACTIVE").
		Set("first_collection_date", sq.Expr("NOW()")).
		Set("first_collection_done", true).
		Set("blocking_new_collections", false).
		Set("installments_paid", 1).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"request_id": requestID})

	dblib.QueueExecRow(batch, upd)

	// Queue workflow state update
	wfUpd := dblib.Psql.Update(workflowStateTable).
		Set("first_collection_done", true).
		Set("installments_paid", 1).
		Set("last_updated", sq.Expr("NOW()")).
		Where(sq.Eq{"request_id": requestID})

	dblib.QueueExecRow(batch, wfUpd)

	// Send batch (single round trip, atomic)
	batchResults := r.db.SendBatch(ctx, batch)
	defer batchResults.Close()

	for range batch.Len() {
		_, err := batchResults.Exec()
		if err != nil {
			return err
		}
	}

	return batchResults.Close()
}

// UpdateRevivalRequestForWithdrawal updates revival request for withdrawal
func (r *RevivalRepository) UpdateRevivalRequestForWithdrawal(ctx context.Context, requestID, withdrawalReason string) error {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	batch := &pgx.Batch{}

	// Queue revival request update
	upd := dblib.Psql.Update(revivalRequestsTable).
		Set("current_status", "WITHDRAWN").
		Set("withdrawal_date", sq.Expr("NOW()")).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"request_id": requestID})

	dblib.QueueExecRow(batch, upd)

	// Queue status change history insert
	ins := dblib.Psql.Insert(statusChangeHistoryTable).
		Columns("history_id", "request_id", "from_status", "to_status", "changed_at", "changed_by", "change_reason").
		Values(uuid.New().String(), requestID, "", "WITHDRAWN", time.Now(), "", &withdrawalReason)

	dblib.QueueExecRow(batch, ins)

	// Send batch (single round trip, atomic)
	batchResults := r.db.SendBatch(ctx, batch)
	defer batchResults.Close()

	for range batch.Len() {
		_, err := batchResults.Exec()
		if err != nil {
			return err
		}
	}

	return batchResults.Close()
}

// UpdateRevivalRequestForTermination updates revival request for 60-day SLA termination
func (r *RevivalRepository) UpdateRevivalRequestForTermination(ctx context.Context, requestID string) error {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	batch := &pgx.Batch{}

	// Queue revival request update
	upd := dblib.Psql.Update(revivalRequestsTable).
		Set("current_status", "TERMINATED").
		Set("termination_date", sq.Expr("NOW()")).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"request_id": requestID})

	dblib.QueueExecRow(batch, upd)

	// Queue workflow state update
	wfUpd := dblib.Psql.Update(workflowStateTable).
		Set("current_status", "TERMINATED").
		Set("sla_expired", true).
		Set("workflow_status", "TERMINATED").
		Set("completed_at", sq.Expr("NOW()")).
		Set("last_updated", sq.Expr("NOW()")).
		Where(sq.Eq{"request_id": requestID})

	dblib.QueueExecRow(batch, wfUpd)

	// Send batch (single round trip, atomic)
	batchResults := r.db.SendBatch(ctx, batch)
	defer batchResults.Close()

	for range batch.Len() {
		_, err := batchResults.Exec()
		if err != nil {
			return err
		}
	}

	return batchResults.Close()
}

// RecordStatusChange records a status change in history
func (r *RevivalRepository) RecordStatusChange(ctx context.Context, requestID, fromStatus, toStatus, changedBy string, changeReason *string) error {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	ins := dblib.Psql.Insert(statusChangeHistoryTable).
		Columns("history_id", "request_id", "from_status", "to_status", "changed_at", "changed_by", "change_reason").
		Values(uuid.New().String(), requestID, fromStatus, toStatus, time.Now(), changedBy, changeReason)

	sql, args, err := ins.ToSql()
	if err != nil {
		return err
	}

	_, err = r.db.Exec(ctx, sql, args...)
	return err
}

// UpdateWorkflowIDsAndCreateState updates revival request with workflow IDs and creates workflow state in single batch
func (r *RevivalRepository) UpdateWorkflowIDsAndCreateState(ctx context.Context, requestID, workflowID, runID, currentStatus string) error {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
	defer cancel()

	batch := &pgx.Batch{}

	// Queue revival request update with workflow IDs
	upd := dblib.Psql.Update(revivalRequestsTable).
		Set("workflow_id", workflowID).
		Set("run_id", runID).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"request_id": requestID})

	dblib.QueueExecRow(batch, upd)

	// Queue workflow state insert
	ins := dblib.Psql.Insert(workflowStateTable).
		Columns("request_id", "workflow_id", "run_id", "current_status", "workflow_status", "started_at", "last_updated").
		Values(requestID, workflowID, runID, currentStatus, "RUNNING", time.Now(), time.Now())

	dblib.QueueExecRow(batch, ins)

	// Send batch (single round trip, atomic)
	batchResults := r.db.SendBatch(ctx, batch)
	defer batchResults.Close()

	for range batch.Len() {
		_, err := batchResults.Exec()
		if err != nil {
			return err
		}
	}

	return batchResults.Close()
}

// CreateWorkflowState creates workflow state tracking
func (r *RevivalRepository) CreateWorkflowState(ctx context.Context, requestID, workflowID, runID, currentStatus string) error {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	ins := dblib.Psql.Insert(workflowStateTable).
		Columns("request_id", "workflow_id", "run_id", "current_status", "workflow_status", "started_at", "last_updated").
		Values(requestID, workflowID, runID, currentStatus, "RUNNING", time.Now(), time.Now())

	_, err := dblib.Insert(ctx, r.db, ins)
	return err
}

// UpdateWorkflowState updates workflow state
func (r *RevivalRepository) UpdateWorkflowState(ctx context.Context, requestID string, currentStatus string, slaStart, slaEnd time.Time) error {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	upd := dblib.Psql.Update(workflowStateTable).
		Set("current_status", currentStatus).
		Set("sla_start_date", slaStart).
		Set("sla_end_date", slaEnd).
		Set("last_updated", sq.Expr("NOW()")).
		Where(sq.Eq{"request_id": requestID})

	_, err := dblib.Update(ctx, r.db, upd)
	return err
}

// UpdateWorkflowStateFirstCollection marks first collection complete in workflow state
func (r *RevivalRepository) UpdateWorkflowStateFirstCollection(ctx context.Context, requestID string) error {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	upd := dblib.Psql.Update(workflowStateTable).
		Set("first_collection_done", true).
		Set("installments_paid", 1).
		Set("last_updated", sq.Expr("NOW()")).
		Where(sq.Eq{"request_id": requestID})

	_, err := dblib.Update(ctx, r.db, upd)
	return err
}

// UpdateWorkflowStateForTermination marks workflow as terminated
func (r *RevivalRepository) UpdateWorkflowStateForTermination(ctx context.Context, requestID string) error {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	upd := dblib.Psql.Update(workflowStateTable).
		Set("current_status", "TERMINATED").
		Set("sla_expired", true).
		Set("workflow_status", "TERMINATED").
		Set("completed_at", sq.Expr("NOW()")).
		Set("last_updated", sq.Expr("NOW()")).
		Where(sq.Eq{"request_id": requestID})

	_, err := dblib.Update(ctx, r.db, upd)
	return err
}

// CheckOngoingRevival checks if policy has ongoing revival
func (r *RevivalRepository) CheckOngoingRevival(ctx context.Context, policyNumber string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	q := dblib.Psql.Select("COUNT(*)").
		From(revivalRequestsTable).
		Where(sq.And{
			sq.Eq{"policy_number": policyNumber},
			sq.NotEq{"current_status": []string{"COMPLETED", "WITHDRAWN", "TERMINATED", "REJECTED"}},
		})

	type countResult struct {
		Count int `db:"count"`
	}

	result, err := dblib.SelectOne(ctx, r.db, q, pgx.RowToStructByName[countResult])

	if err != nil {
		return false, err
	}

	return result.Count > 0, nil
}

// GenerateTicketID generates a new ticket ID
func (r *RevivalRepository) GenerateTicketID(ctx context.Context) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	// Use UUID with prefix
	ticketID := fmt.Sprintf("PSREYV%s", uuid.New().String()[:12])
	return ticketID, nil
}

// IncrementInstallmentsPaid increments the installments paid counter
func (r *RevivalRepository) IncrementInstallmentsPaid(ctx context.Context, requestID string) error {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	//concurency issue fix
	upd := dblib.Psql.Update(revivalRequestsTable).
		Set("installments_paid", sq.Expr("installments_paid + 1")).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"request_id": requestID}).
		Where(sq.Expr("installments_paid < number_of_installments"))

	_, err := dblib.Update(ctx, r.db, upd)
	return err
}

// UpdateRevivalRequestForCompletion marks revival request as completed
func (r *RevivalRepository) UpdateRevivalRequestForCompletion(ctx context.Context, requestID string) error {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	now := time.Now()
	upd := dblib.Psql.Update(revivalRequestsTable).
		Set("current_status", "COMPLETED").
		Set("completed_date", &now).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"request_id": requestID})

	_, err := dblib.Update(ctx, r.db, upd)
	return err
}

// GetInstallmentByNumber retrieves an installment by request ID and installment number
// Returns nil if installment doesn't exist (not an error - means not paid yet)
func (r *RevivalRepository) GetInstallmentByNumber(ctx context.Context, requestID string, installmentNumber int) (*domain.InstallmentSchedule, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	q := dblib.Psql.Select(
		"schedule_id", "request_id", "policy_number", "installment_number",
		"installment_amount", "tax_amount", "total_amount",
		"due_date", "payment_date", "is_paid", "grace_period_days",
		"created_at", "updated_at",
	).
		From(installmentSchedulesTable).
		Where(sq.And{
			sq.Eq{"request_id": requestID},
			sq.Eq{"installment_number": installmentNumber},
		})

	installment, err := dblib.SelectOne(ctx, r.db, q, pgx.RowToStructByName[domain.InstallmentSchedule])
	if err != nil {
		if err == pgx.ErrNoRows {
			// Installment doesn't exist yet - not an error, just not paid
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query installment: %w", err)
	}

	return &installment, nil
}

// CreateInstallmentSchedule creates an installment schedule record
func (r *RevivalRepository) CreateInstallmentSchedule(ctx context.Context, scheduleID, requestID, policyNumber string, installmentNumber int, installmentAmount, taxAmount, totalAmount float64, dueDate time.Time, isPaid bool) error {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	var paymentDate *time.Time
	if isPaid {
		now := time.Now()
		paymentDate = &now
	}

	q := dblib.Psql.Insert(installmentSchedulesTable).
		Columns(
			"schedule_id", "request_id", "policy_number", "installment_number",
			"installment_amount", "tax_amount", "total_amount",
			"due_date", "payment_date", "is_paid", "grace_period_days",
			"created_at", "updated_at",
		).
		Values(
			scheduleID, requestID, policyNumber, installmentNumber,
			installmentAmount, taxAmount, totalAmount,
			dueDate, paymentDate, isPaid, 0,
			time.Now(), time.Now(),
		)

	sql, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build installment schedule insert query: %w", err)
	}

	_, err = r.db.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to create installment schedule: %w", err)
	}

	return nil
}

// UpdateRevivalRequestWithSuspenseAdjustment updates revival request with suspense adjustment details
func (r *RevivalRepository) UpdateRevivalRequestWithSuspenseAdjustment(ctx context.Context, requestID string, previousSuspense, adjustedAmount float64) error {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	upd := dblib.Psql.Update(revivalRequestsTable).
		Set("previous_suspense_amount", previousSuspense).
		Set("suspense_adjusted", true).
		Set("adjusted_revival_amount", adjustedAmount).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"request_id": requestID})

	_, err := dblib.Update(ctx, r.db, upd)
	if err != nil {
		return fmt.Errorf("failed to update revival request with suspense adjustment: %w", err)
	}

	return nil
}

// CreateRevivalTermination creates a termination record for revival workflow
func (r *RevivalRepository) CreateRevivalTermination(ctx context.Context, termination *domain.RevivalTermination) error {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	q := dblib.Psql.Insert(revivalTerminationsTable).
		Columns(
			"termination_id", "request_id", "ticket_id", "policy_number",
			"termination_reason", "termination_type", "installment_number",
			"suspense_created", "suspense_amount", "terminated_at", "terminated_by", "created_at",
		).
		Values(
			termination.TerminationID, termination.RequestID, termination.TicketID, termination.PolicyNumber,
			termination.TerminationReason, termination.TerminationType, termination.InstallmentNumber,
			termination.SuspenseCreated, termination.SuspenseAmount, termination.TerminatedAt,
			termination.TerminatedBy, termination.CreatedAt,
		)

	sql, args, err := q.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build termination insert query: %w", err)
	}

	_, err = r.db.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("failed to create revival termination: %w", err)
	}

	return nil
}

// GetStatusHistoryByRequestID retrieves status change history for a request
func (r *RevivalRepository) GetStatusHistoryByRequestID(ctx context.Context, requestID string) ([]domain.StatusChangeHistory, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	sel := dblib.Psql.Select(
		"history_id", "request_id", "from_status", "to_status",
		"changed_at", "changed_by", "change_reason",
	).From(statusChangeHistoryTable).
		Where(sq.Eq{"request_id": requestID}).
		OrderBy("changed_at DESC")

	return dblib.SelectRows(ctx, r.db, sel, pgx.RowToStructByName[domain.StatusChangeHistory])
}
