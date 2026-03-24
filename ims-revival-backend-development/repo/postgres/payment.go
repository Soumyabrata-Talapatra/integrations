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
	log "gitlab.cept.gov.in/it-2.0-common/n-api-log"
)

type PaymentRepository struct {
	db  *dblib.DB
	cfg *config.Config
}

func NewPaymentRepository(db *dblib.DB, cfg *config.Config) *PaymentRepository {
	return &PaymentRepository{db: db, cfg: cfg}
}

const (
	paymentTransactionsTable  = "collection.payment_transactions"
	collectionBatchTable      = "collection.collection_batch_tracking"
	chequeClearingStatusTable = "collection.cheque_clearing_status"
	suspenseAccountsTable     = "collection.suspense_accounts"
)

// CreateDualPaymentRecords creates premium + installment payments using batch operations (IR_36)
// Batch is implicit transactional and single trip to database
func (r *PaymentRepository) CreateDualPaymentRecords(ctx context.Context, premiumPayment, installmentPayment domain.PaymentTransaction, batchID string) error {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
	defer cancel()

	batch := &pgx.Batch{}

	// Generate IDs and set premium payment fields
	premiumPayment.PaymentID = uuid.New().String()
	premiumPayment.PaymentType = "PREMIUM"
	premiumPayment.CollectionBatchID = batchID
	premiumPayment.CollectionDate = time.Now()
	premiumPayment.InstallmentNumber = nil
	premiumPayment.LinkedPaymentID = &installmentPayment.PaymentID

	// Generate IDs and set installment payment fields
	installmentPayment.PaymentID = uuid.New().String()
	installmentPayment.PaymentType = "INSTALLMENT"
	installmentPayment.CollectionBatchID = batchID
	installmentPayment.CollectionDate = time.Now()
	installmentPayment.InstallmentNumber = &[]int{1}[0]
	installmentPayment.LinkedPaymentID = &premiumPayment.PaymentID

	// Queue premium payment insert
	premiumIns := dblib.Psql.Insert(paymentTransactionsTable).
		Columns(
			"payment_id", "request_id", "policy_number", "collection_batch_id", "linked_payment_id",
			"payment_type", "amount", "tax_amount", "total_amount", "payment_mode",
			"payment_status", "collection_date", "collected_by",
		).
		Values(
			premiumPayment.PaymentID, premiumPayment.RequestID, premiumPayment.PolicyNumber,
			batchID, &installmentPayment.PaymentID,
			premiumPayment.PaymentType, premiumPayment.Amount, premiumPayment.TaxAmount, premiumPayment.TotalAmount,
			premiumPayment.PaymentMode, premiumPayment.PaymentStatus, premiumPayment.CollectionDate, premiumPayment.CollectedBy,
		)

	if err := dblib.QueueExecRow(batch, premiumIns); err != nil {
		return err
	}

	// Queue installment payment insert
	installmentIns := dblib.Psql.Insert(paymentTransactionsTable).
		Columns(
			"payment_id", "request_id", "policy_number", "collection_batch_id", "linked_payment_id",
			"payment_type", "installment_number", "amount", "tax_amount", "total_amount",
			"payment_mode", "payment_status", "collection_date", "collected_by",
		).
		Values(
			installmentPayment.PaymentID, installmentPayment.RequestID, installmentPayment.PolicyNumber,
			batchID, &premiumPayment.PaymentID,
			installmentPayment.PaymentType, installmentPayment.InstallmentNumber, installmentPayment.Amount,
			installmentPayment.TaxAmount, installmentPayment.TotalAmount, installmentPayment.PaymentMode,
			installmentPayment.PaymentStatus, installmentPayment.CollectionDate, installmentPayment.CollectedBy,
		)

	if err := dblib.QueueExecRow(batch, installmentIns); err != nil {
		return err
	}

	// Create collection batch tracking
	collectionBatch := domain.CollectionBatchTracking{
		BatchID:              batchID,
		RequestID:            premiumPayment.RequestID,
		PolicyNumber:         premiumPayment.PolicyNumber,
		PremiumPaymentID:     premiumPayment.PaymentID,
		InstallmentPaymentID: installmentPayment.PaymentID,
		CollectionComplete:   false,
		CollectionDate:       time.Now(),
	}

	insBatch := dblib.Psql.Insert(collectionBatchTable).
		Columns(
			"batch_id", "request_id", "policy_number", "premium_payment_id",
			"installment_payment_id", "collection_complete", "collection_date",
		).
		Values(
			collectionBatch.BatchID, collectionBatch.RequestID, collectionBatch.PolicyNumber,
			collectionBatch.PremiumPaymentID, collectionBatch.InstallmentPaymentID,
			collectionBatch.CollectionComplete, collectionBatch.CollectionDate,
		)

	if err := dblib.QueueExecRow(batch, insBatch); err != nil {
		return err
	}

	// Send batch to database (implicit transaction, single trip)
	batchResults := r.db.SendBatch(ctx, batch)
	defer batchResults.Close()

	// Check all batched operations succeeded
	for range batch.Len() {
		commandTag, err := batchResults.Exec()
		if err != nil {
			return err
		}
		if commandTag.RowsAffected() == 0 {
			return fmt.Errorf("no rows affected during batch operation")
		}
	}

	return batchResults.Close()
}
func (r *PaymentRepository) CreateInstallmentPayment(
	ctx context.Context,
	payment domain.PaymentTransaction,
) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
	defer cancel()

	// 🎯 DUPLICATE PREVENTION: Check if payment already exists for this installment
	if payment.InstallmentNumber != nil {
		var existingPaymentID string
		checkSQL := `
			SELECT payment_id 
			FROM collection.payment_transactions 
			WHERE request_id = $1 AND installment_number = $2
			LIMIT 1
		`
		err := r.db.QueryRow(ctx, checkSQL, payment.RequestID, *payment.InstallmentNumber).Scan(&existingPaymentID)
		if err == nil {
			// Payment already exists - return existing payment_id instead of creating duplicate
			log.Info(nil, "⚠️ Installment payment already exists, skipping duplicate",
				"request_id", payment.RequestID,
				"installment_number", *payment.InstallmentNumber,
				"existing_payment_id", existingPaymentID)
			return existingPaymentID, nil
		}
		// If error is pgx.ErrNoRows, continue with insertion (no duplicate found)
		// Any other error will be caught later
	}

	// Generate payment ID and batch ID
	payment.PaymentID = uuid.New().String()
	batchID := uuid.New().String()

	// Enforce installment payment type
	payment.PaymentType = "INSTALLMENT"
	payment.CollectionDate = time.Now()
	payment.CollectionBatchID = batchID

	log.Info(nil, "💳 Inserting payment record using dblib.Insert",
		"payment_id", payment.PaymentID,
		"batch_id", batchID,
		"installment_number", payment.InstallmentNumber,
		"amount", payment.Amount)

	ins := dblib.Psql.Insert(paymentTransactionsTable).
		Columns(
			"payment_id",
			"request_id",
			"policy_number",
			"collection_batch_id",
			"payment_type",
			"installment_number",
			"amount",
			"tax_amount",
			"total_amount",
			"payment_mode",
			"payment_status",
			"collection_date",
			"collected_by",
		).
		Values(
			payment.PaymentID,
			payment.RequestID,
			payment.PolicyNumber,
			batchID,
			payment.PaymentType,
			payment.InstallmentNumber,
			payment.Amount,
			payment.TaxAmount,
			payment.TotalAmount,
			payment.PaymentMode,
			payment.PaymentStatus,
			payment.CollectionDate,
			payment.CollectedBy,
		)

	// Use dblib.Insert for proper transaction handling (same as CreateWorkflowState)
	_, err := dblib.Insert(ctx, r.db, ins)
	if err != nil {
		log.Error(nil, "❌ dblib.Insert failed", "payment_id", payment.PaymentID, "error", err)
		return "", fmt.Errorf("failed to insert payment: %w", err)
	}

	log.Info(nil, "✅ Payment inserted successfully via dblib.Insert", "payment_id", payment.PaymentID)

	return payment.PaymentID, nil
}

// UpdatePaymentStatus updates payment status
func (r *PaymentRepository) UpdatePaymentStatus(ctx context.Context, paymentID string, status string) error {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	upd := dblib.Psql.Update(paymentTransactionsTable).
		Set("payment_status", status).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"payment_id": paymentID})

	_, err := dblib.Update(ctx, r.db, upd)
	return err
}

// UpdateCollectionBatchComplete marks collection batch as complete
func (r *PaymentRepository) UpdateCollectionBatchComplete(ctx context.Context, batchID string, combinedReceiptID string) error {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	upd := dblib.Psql.Update(collectionBatchTable).
		Set("collection_complete", true).
		Set("combined_receipt_id", &combinedReceiptID).
		Set("completed_at", sq.Expr("NOW()")).
		Where(sq.Eq{"batch_id": batchID})

	_, err := dblib.Update(ctx, r.db, upd)
	return err
}

// GetPaymentByID retrieves a payment by ID
func (r *PaymentRepository) GetPaymentByID(ctx context.Context, paymentID string) (domain.PaymentTransaction, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	q := dblib.Psql.Select(
		"payment_id", "request_id", "policy_number", "collection_batch_id", "linked_payment_id",
		"cheque_id", "payment_type", "installment_number", "amount", "tax_amount",
		"total_amount", "payment_mode", "payment_status", "collection_date",
		"payment_date", "receipt_id", "tigerbeetle_transfer_id", "collected_by",
		"created_at", "updated_at",
	).
		From(paymentTransactionsTable).
		Where(sq.Eq{"payment_id": paymentID})

	return dblib.SelectOne(ctx, r.db, q, pgx.RowToStructByName[domain.PaymentTransaction])
}

// GetPaymentsByBatchID retrieves all payments in a collection batch
func (r *PaymentRepository) GetPaymentsByBatchID(ctx context.Context, batchID string) ([]domain.PaymentTransaction, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
	defer cancel()

	q := dblib.Psql.Select(
		"payment_id", "request_id", "policy_number", "collection_batch_id", "linked_payment_id",
		"payment_type", "amount", "total_amount", "payment_status",
		"collection_date", "payment_date", "receipt_id",
	).
		From(paymentTransactionsTable).
		Where(sq.Eq{"collection_batch_id": batchID}).
		OrderBy("created_at")

	return dblib.SelectRows(ctx, r.db, q, pgx.RowToStructByName[domain.PaymentTransaction])
}

// CreateChequeClearingRecord creates a cheque clearing status record
func (r *PaymentRepository) CreateChequeClearingRecord(ctx context.Context, cheque domain.ChequeClearingStatus) error {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	if cheque.ChequeID == "" {
		cheque.ChequeID = uuid.New().String()
	}

	ins := dblib.Psql.Insert(chequeClearingStatusTable).
		Columns(
			"cheque_id", "payment_id", "request_id", "policy_number",
			"cheque_number", "bank_name", "cheque_date", "amount",
			"clearance_status", "next_due_date",
		).
		Values(
			cheque.ChequeID, cheque.PaymentID, cheque.RequestID, cheque.PolicyNumber,
			cheque.ChequeNumber, cheque.BankName, cheque.ChequeDate, cheque.Amount,
			cheque.ClearanceStatus, cheque.NextDueDate,
		)

	sql, args, err := ins.ToSql()
	if err != nil {
		return err
	}

	_, err = r.db.Exec(ctx, sql, args...)
	return err
}

// UpdateChequeClearanceStatus updates cheque clearance status
func (r *PaymentRepository) UpdateChequeClearanceStatus(ctx context.Context, chequeID, status string) error {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	upd := dblib.Psql.Update(chequeClearingStatusTable).
		Set("clearance_status", status).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"cheque_id": chequeID})

	_, err := dblib.Update(ctx, r.db, upd)
	return err
}

// CreateSuspenseEntry creates a suspense account entry
func (r *PaymentRepository) CreateSuspenseEntry(ctx context.Context, suspense domain.SuspenseAccount) error {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	if suspense.SuspenseID == "" {
		suspense.SuspenseID = uuid.New().String()
	}

	ins := dblib.Psql.Insert(suspenseAccountsTable).
		Columns(
			"suspense_id", "policy_number", "request_id", "suspense_type",
			"amount", "is_reversed", "created_by", "reason", "suspense_account_type",
		).
		Values(
			suspense.SuspenseID, suspense.PolicyNumber, suspense.RequestID,
			suspense.SuspenseType, suspense.Amount, suspense.IsReversed, suspense.CreatedBy, suspense.Reason,
			suspense.SuspenseAccountType,
		)

	sql, args, err := ins.ToSql()
	if err != nil {
		return err
	}

	_, err = r.db.Exec(ctx, sql, args...)
	return err
}

// ReverseSuspenseEntry marks suspense as reversed (IR_28 check applies at business layer)
func (r *PaymentRepository) ReverseSuspenseEntry(ctx context.Context, suspenseID string, reversalAuthorizedBy, reversalReason string) error {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	upd := dblib.Psql.Update(suspenseAccountsTable).
		Set("is_reversed", true).
		Set("reversal_date", sq.Expr("NOW()")).
		Set("reversal_authorized_by", reversalAuthorizedBy).
		Set("reversal_reason", reversalReason).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"suspense_id": suspenseID})

	_, err := dblib.Update(ctx, r.db, upd)
	return err
}

// GetSuspenseByPolicyNumber retrieves suspense entries for a policy
func (r *PaymentRepository) GetSuspenseByPolicyNumber(ctx context.Context, policyNumber string) ([]domain.SuspenseAccount, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
	defer cancel()

	q := dblib.Psql.Select(
		"suspense_id", "policy_number", "request_id", "suspense_type",
		"amount", "is_reversed", "reversal_date", "reversal_authorized_by",
		"reversal_reason", "created_at", "created_by", "updated_at",
	).
		From(suspenseAccountsTable).
		Where(sq.Eq{"policy_number": policyNumber}).
		OrderBy("created_at DESC")

	return dblib.SelectRows(ctx, r.db, q, pgx.RowToStructByName[domain.SuspenseAccount])
}

// GetActiveSuspenseByPolicy retrieves non-reversed suspense entries for a policy
// Used for re-revival suspense adjustment
func (r *PaymentRepository) GetActiveSuspenseByPolicy(ctx context.Context, policyNumber string) ([]domain.SuspenseAccount, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
	defer cancel()

	q := dblib.Psql.Select(
		"suspense_id", "policy_number", "request_id", "suspense_type",
		"amount", "is_reversed", "reversal_date", "reversal_authorized_by",
		"reversal_reason", "created_at", "created_by", "updated_at", "suspense_account_type",
		"reason",
	).
		From(suspenseAccountsTable).
		Where(sq.And{
			sq.Eq{"policy_number": policyNumber},
			sq.Eq{"is_reversed": false},
		}).
		OrderBy("created_at ASC")

	return dblib.SelectRows(ctx, r.db, q, pgx.RowToStructByName[domain.SuspenseAccount])
}

// ReverseSuspenseForAdjustment reverses multiple suspense entries for re-revival adjustment
// Marks all suspense entries as reversed with reason "Adjusted against new revival"
func (r *PaymentRepository) ReverseSuspenseForAdjustment(ctx context.Context, policyNumber, newRequestID, authorizedBy string) error {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
	defer cancel()

	reason := "Adjusted against new revival request " + newRequestID

	upd := dblib.Psql.Update(suspenseAccountsTable).
		Set("is_reversed", true).
		Set("reversal_date", sq.Expr("NOW()")).
		Set("reversal_authorized_by", authorizedBy).
		Set("reversal_reason", reason).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.And{
			sq.Eq{"policy_number": policyNumber},
			sq.Eq{"is_reversed": false},
		})

	_, err := dblib.Update(ctx, r.db, upd)
	return err
}

// GetCollectionBatchByRequestID retrieves collection batch by request ID
func (r *PaymentRepository) GetCollectionBatchByRequestID(ctx context.Context, requestID string) (domain.CollectionBatchTracking, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	q := dblib.Psql.Select(
		"batch_id", "request_id", "policy_number", "premium_payment_id",
		"installment_payment_id", "collection_complete", "collection_date",
		"combined_receipt_id", "created_at", "completed_at",
	).
		From(collectionBatchTable).
		Where(sq.Eq{"request_id": requestID})

	return dblib.SelectOne(ctx, r.db, q, pgx.RowToStructByName[domain.CollectionBatchTracking])
}

func (r *PaymentRepository) GetTotalInstallmentPaid(
	ctx context.Context,
	requestID string,
) (float64, error) {

	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	q := dblib.Psql.
		Select("COALESCE(SUM(amount), 0)").
		From(paymentTransactionsTable).
		Where(sq.Eq{
			"request_id":     requestID,
			"payment_type":   "INSTALLMENT",
			"payment_status": "COMPLETED",
		})

	sqlStr, args, err := q.ToSql()
	if err != nil {
		return 0, err
	}

	var totalPaid float64
	err = r.db.QueryRow(ctx, sqlStr, args...).Scan(&totalPaid)
	if err != nil {
		return 0, err
	}

	return totalPaid, nil
}

func (r *PaymentRepository) GetTotalPremiumPaid(
	ctx context.Context,
	requestID string,
) (float64, error) {

	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	q := dblib.Psql.
		Select("COALESCE(SUM(amount), 0)").
		From(paymentTransactionsTable).
		Where(sq.Eq{
			"request_id":     requestID,
			"payment_type":   "PREMIUM",
			"payment_status": "COMPLETED",
		})

	sqlStr, args, err := q.ToSql()
	if err != nil {
		return 0, err
	}

	var totalPaid float64
	err = r.db.QueryRow(ctx, sqlStr, args...).Scan(&totalPaid)
	if err != nil {
		return 0, err
	}

	return totalPaid, nil
}
