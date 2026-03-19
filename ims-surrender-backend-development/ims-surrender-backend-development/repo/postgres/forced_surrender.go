package repo

import (
	"context"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	config "gitlab.cept.gov.in/it-2.0-common/api-config"
	dblib "gitlab.cept.gov.in/it-2.0-common/n-api-db"

	"gitlab.cept.gov.in/it-2.0-policy/surrender-service/core/domain"
)

// ForcedSurrenderRepository handles all database operations for forced surrender
// Business Rules: BR-FS-001 to BR-FS-018
type ForcedSurrenderRepository struct {
	db  *dblib.DB
	cfg *config.Config
}

// NewForcedSurrenderRepository creates a new forced surrender repository
func NewForcedSurrenderRepository(db *dblib.DB, cfg *config.Config) *ForcedSurrenderRepository {
	return &ForcedSurrenderRepository{
		db:  db,
		cfg: cfg,
	}
}

const (
	reminderTable      = "forced_surrender_reminders"
	paymentWindowTable = "forced_surrender_payment_windows"
)

// CreateReminder creates a forced surrender reminder record
// Business Rules: BR-FS-002, BR-FS-003, BR-FS-004
// Temporal Workflow: TEMP-005
func (r *ForcedSurrenderRepository) CreateReminder(ctx context.Context, data domain.ForcedSurrenderReminder) (domain.ForcedSurrenderReminder, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Insert(reminderTable).
		Columns(
			"policy_id", "reminder_number", "reminder_date",
			"loan_capitalization_ratio", "loan_principal", "loan_interest",
			"gross_surrender_value", "letter_sent", "sms_sent",
			"letter_reference", "sms_reference", "metadata",
		).
		Values(
			data.PolicyID, data.ReminderNumber, data.ReminderDate,
			data.LoanCapitalizationRatio, data.LoanPrincipal, data.LoanInterest,
			data.GrossSurrenderValue, data.LetterSent, data.SMSSent,
			data.LetterReference, data.SMSReference, data.Metadata,
		).
		Suffix("RETURNING *").
		PlaceholderFormat(sq.Dollar)

	result, err := dblib.InsertReturning(ctx, r.db, query, pgx.RowToStructByName[domain.ForcedSurrenderReminder])
	if err != nil {
		return result, err
	}

	return result, nil
}

// FindRemindersByPolicyID retrieves all reminders for a policy
func (r *ForcedSurrenderRepository) FindRemindersByPolicyID(ctx context.Context, policyID string) ([]domain.ForcedSurrenderReminder, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
	defer cancel()

	query := dblib.Psql.Select("*").
		From(reminderTable).
		Where(sq.Eq{"policy_id": policyID}).
		OrderBy("created_at ASC").
		PlaceholderFormat(sq.Dollar)

	results, err := dblib.SelectRows(ctx, r.db, query, pgx.RowToStructByName[domain.ForcedSurrenderReminder])
	if err != nil {
		return results, err
	}

	return results, nil
}

// FindLatestReminderByPolicyID retrieves the latest reminder for a policy
// Business Rule: BR-FS-005 (check if 3rd reminder already sent)
func (r *ForcedSurrenderRepository) FindLatestReminderByPolicyID(ctx context.Context, policyID string) (domain.ForcedSurrenderReminder, bool, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("*").
		From(reminderTable).
		Where(sq.Eq{"policy_id": policyID}).
		OrderBy("created_at DESC").
		Limit(1).
		PlaceholderFormat(sq.Dollar)

	result, found, err := dblib.SelectOneOK(ctx, r.db, query, pgx.RowToStructByName[domain.ForcedSurrenderReminder])
	return result, found, err
}

// FindReminderByID retrieves a reminder by ID
func (r *ForcedSurrenderRepository) FindReminderByID(ctx context.Context, id uuid.UUID) (domain.ForcedSurrenderReminder, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("*").
		From(reminderTable).
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Dollar)

	result, err := dblib.SelectOne(ctx, r.db, query, pgx.RowToStructByName[domain.ForcedSurrenderReminder])
	return result, err
}

// FindPaymentWindowByReminderID retrieves a payment window by reminder ID (via policy lookup)
func (r *ForcedSurrenderRepository) FindPaymentWindowByReminderID(ctx context.Context, policyID uuid.UUID) (domain.ForcedSurrenderPaymentWindow, bool, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("*").
		From(paymentWindowTable).
		Where(sq.Eq{"policy_id": policyID}).
		OrderBy("created_at DESC").
		Limit(1).
		PlaceholderFormat(sq.Dollar)

	result, found, err := dblib.SelectOneOK(ctx, r.db, query, pgx.RowToStructByName[domain.ForcedSurrenderPaymentWindow])
	return result, found, err
}

// CreatePaymentWindow creates a payment window record
// Business Rule: BR-FS-006 (30-day payment window after 3rd reminder)
// Temporal Workflow: TEMP-003
func (r *ForcedSurrenderRepository) CreatePaymentWindow(ctx context.Context, data domain.ForcedSurrenderPaymentWindow) (domain.ForcedSurrenderPaymentWindow, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Insert(paymentWindowTable).
		Columns(
			"surrender_request_id", "policy_id", "window_start_date", "window_end_date",
		).
		Values(
			data.SurrenderRequestID, data.PolicyID, data.WindowStartDate, data.WindowEndDate,
		).
		Suffix("RETURNING *").
		PlaceholderFormat(sq.Dollar)

	result, err := dblib.InsertReturning(ctx, r.db, query, pgx.RowToStructByName[domain.ForcedSurrenderPaymentWindow])
	if err != nil {
		return result, err
	}

	return result, nil
}

// FindPaymentWindowBySurrenderRequestID retrieves payment window by surrender request ID
func (r *ForcedSurrenderRepository) FindPaymentWindowBySurrenderRequestID(ctx context.Context, surrenderRequestID uuid.UUID) (domain.ForcedSurrenderPaymentWindow, bool, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("*").
		From(paymentWindowTable).
		Where(sq.Eq{"surrender_request_id": surrenderRequestID}).
		PlaceholderFormat(sq.Dollar)

	result, found, err := dblib.SelectOneOK(ctx, r.db, query, pgx.RowToStructByName[domain.ForcedSurrenderPaymentWindow])
	return result, found, err
}

// ListExpiredPaymentWindows retrieves payment windows that have expired without payment
// Business Rule: BR-FS-007 (auto-forward to approval queue after expiry)
// Temporal Workflow: TEMP-002
func (r *ForcedSurrenderRepository) ListExpiredPaymentWindows(ctx context.Context) ([]domain.ForcedSurrenderPaymentWindow, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
	defer cancel()

	query := dblib.Psql.Select("*").
		From(paymentWindowTable).
		Where(sq.Eq{
			"payment_received":   false,
			"workflow_forwarded": false,
		}).
		Where(sq.Lt{"window_end_date": time.Now()}).
		PlaceholderFormat(sq.Dollar)

	results, err := dblib.SelectRows(ctx, r.db, query, pgx.RowToStructByName[domain.ForcedSurrenderPaymentWindow])
	if err != nil {
		return results, err
	}

	return results, nil
}

// UpdatePaymentReceived updates payment window when payment is received
// Business Rule: BR-FS-012 (payment creates demand for reversal)
func (r *ForcedSurrenderRepository) UpdatePaymentReceived(ctx context.Context, id uuid.UUID, amount float64, reference string) (domain.ForcedSurrenderPaymentWindow, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	now := time.Now()
	query := dblib.Psql.Update(paymentWindowTable).
		Set("payment_received", true).
		Set("payment_received_at", now).
		Set("payment_amount", amount).
		Set("payment_reference", reference).
		Where(sq.Eq{"id": id}).
		Suffix("RETURNING *").
		PlaceholderFormat(sq.Dollar)

	result, err := dblib.UpdateReturning(ctx, r.db, query, pgx.RowToStructByName[domain.ForcedSurrenderPaymentWindow])
	if err != nil {
		return result, err
	}

	return result, nil
}

// UpdateWorkflowForwarded marks payment window as forwarded to workflow
// Temporal Workflow: TEMP-002
func (r *ForcedSurrenderRepository) UpdateWorkflowForwarded(ctx context.Context, id uuid.UUID) (domain.ForcedSurrenderPaymentWindow, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	now := time.Now()
	query := dblib.Psql.Update(paymentWindowTable).
		Set("workflow_forwarded", true).
		Set("workflow_forwarded_at", now).
		Where(sq.Eq{"id": id}).
		Suffix("RETURNING *").
		PlaceholderFormat(sq.Dollar)

	result, err := dblib.UpdateReturning(ctx, r.db, query, pgx.RowToStructByName[domain.ForcedSurrenderPaymentWindow])
	if err != nil {
		return result, err
	}

	return result, nil
}

// UpdateAutoCompleted marks payment window as auto-completed
// Business Rule: BR-FS-008 (auto-completion disposition)
// Temporal Workflow: TEMP-002
func (r *ForcedSurrenderRepository) UpdateAutoCompleted(ctx context.Context, id uuid.UUID) (domain.ForcedSurrenderPaymentWindow, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	now := time.Now()
	query := dblib.Psql.Update(paymentWindowTable).
		Set("auto_completed", true).
		Set("auto_completed_at", now).
		Where(sq.Eq{"id": id}).
		Suffix("RETURNING *").
		PlaceholderFormat(sq.Dollar)

	result, err := dblib.UpdateReturning(ctx, r.db, query, pgx.RowToStructByName[domain.ForcedSurrenderPaymentWindow])
	if err != nil {
		return result, err
	}

	return result, nil
}
