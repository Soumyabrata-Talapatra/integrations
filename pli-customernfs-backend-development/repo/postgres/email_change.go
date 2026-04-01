// Package postgres implements the repository layer for customer-nfs-service.
//
// FR-NFS-007: Email Change (OTP-based verification)
// BR-NFS-016: Audit trail for all operations
//
// BATCH NOTE:
//
//	CreateEmailChangeDetail → single TX batch:
//	  INSERT email_change_detail + INSERT audit_log
package postgres

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"

	config "gitlab.cept.gov.in/it-2.0-common/api-config"
	dblib "gitlab.cept.gov.in/it-2.0-common/n-api-db"
	log "gitlab.cept.gov.in/it-2.0-common/n-api-log"

	"customer-nfs-service/core/domain"
)

// EmailChangeRepository handles all DB operations for nfs.email_change_detail.
type EmailChangeRepository struct {
	db  *dblib.DB
	cfg *config.Config
}

// NewEmailChangeRepository constructs the repository.
func NewEmailChangeRepository(db *dblib.DB, cfg *config.Config) *EmailChangeRepository {
	return &EmailChangeRepository{db: db, cfg: cfg}
}

// ---------------------------------------------------------------------------
// CreateEmailChangeDetail creates email_change_detail + audit_log in a single batched transaction.
//
// BATCH: 2 INSERTs in one pgx TX batch.
// FR-NFS-007, BR-NFS-016
// ---------------------------------------------------------------------------
func (r *EmailChangeRepository) CreateEmailChangeDetail(
	ctx context.Context,
	detail *domain.EmailChangeDetail,
	audit *domain.AuditLog,
) error {
	timeout := r.cfg.GetDuration("db.QueryTimeoutMed")
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return r.db.WithTx(ctx, func(tx pgx.Tx) error {
		batch := &pgx.Batch{}

		// 1. INSERT email_change_detail
		emailSQL, emailArgs, err := dblib.Psql.Insert("nfs.email_change_detail").
			Columns(
				"detail_id", "request_id", "old_email", "new_email",
				"created_by",
			).
			Values(
				detail.DetailID, detail.RequestID, detail.OldEmail, detail.NewEmail,
				detail.CreatedBy,
			).
			Suffix("RETURNING detail_id, created_at").
			ToSql()
		if err != nil {
			return fmt.Errorf("build email_change_detail insert: %w", err)
		}
		batch.Queue(emailSQL, emailArgs...)

		// 2. INSERT audit_log (BR-NFS-016 — every operation must be audited)
		auditSQL, auditArgs, err := dblib.Psql.Insert("nfs.audit_log").
			Columns(
				"audit_id", "request_id", "action_type", "new_value",
				"performed_by", "channel", "office_code",
			).
			Values(
				audit.AuditID, audit.RequestID, audit.ActionType, audit.NewValueJSON,
				audit.PerformedByID, audit.Channel, audit.OfficeCode,
			).
			ToSql()
		if err != nil {
			return fmt.Errorf("build audit_log insert: %w", err)
		}
		batch.Queue(auditSQL, auditArgs...)

		// Send both inserts in one network round-trip
		br := tx.SendBatch(ctx, batch)

		// Result 1: email_change_detail
		if err := br.QueryRow().Scan(&detail.DetailID, &detail.CreatedAt); err != nil {
			br.Close()
			return fmt.Errorf("scan email_change_detail result: %w", err)
		}

		// Result 2: audit_log (exec-only, check RowsAffected)
		if _, err := br.Exec(); err != nil {
			br.Close()
			return fmt.Errorf("exec audit_log insert: %w", err)
		}
		br.Close()

		return nil
	})
}

// ---------------------------------------------------------------------------
// GetByRequestID retrieves email_change_detail for a given request_id.
// Used by UpdateEmailData activity to fetch new email.
// ---------------------------------------------------------------------------
func (r *EmailChangeRepository) GetByRequestID(ctx context.Context, requestID string) (*domain.EmailChangeDetail, error) {
	timeout := r.cfg.GetDuration("db.QueryTimeoutLow")
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	query := dblib.Psql.Select(
		"detail_id", "request_id", "old_email", "new_email",
		"created_at", "updated_at", "created_by", "updated_by", "version",
	).
		From("nfs.email_change_detail").
		Where(sq.Eq{"request_id": requestID}).
		Limit(1)

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build email_change_detail select: %w", err)
	}

	var detail domain.EmailChangeDetail
	row := r.db.QueryRow(ctx, sql, args...)
	if err := row.Scan(
		&detail.DetailID, &detail.RequestID, &detail.OldEmail, &detail.NewEmail,
		&detail.CreatedAt, &detail.UpdatedAt, &detail.CreatedBy, &detail.UpdatedBy, &detail.Version,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		log.Error(ctx, "GetByRequestID scan: %v", err)
		return nil, fmt.Errorf("get email_change_detail: %w", err)
	}

	return &detail, nil
}

// ---------------------------------------------------------------------------
// CreateEmailVersionHistory creates a new version in email_version_history.
// Called after successful email update to maintain version history.
// ---------------------------------------------------------------------------
func (r *EmailChangeRepository) CreateEmailVersionHistory(
	ctx context.Context,
	history *domain.EmailVersionHistory,
) error {
	timeout := r.cfg.GetDuration("db.QueryTimeoutLow")
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	query := dblib.Psql.Insert("nfs.email_version_history").
		Columns(
			"version_id", "customer_id", "request_id", "email",
			"version_number", "is_active", "effective_from", "effective_to", "created_by",
		).
		Values(
			history.VersionID, history.CustomerID, history.RequestID, history.Email,
			history.VersionNumber, history.IsActive, history.EffectiveFrom, history.EffectiveTo, history.CreatedBy,
		)

	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("build email_version_history insert: %w", err)
	}

	_, err = r.db.Exec(ctx, sql, args...)
	if err != nil {
		log.Error(ctx, "CreateEmailVersionHistory exec: %v", err)
		return fmt.Errorf("create email_version_history: %w", err)
	}

	return nil
}
