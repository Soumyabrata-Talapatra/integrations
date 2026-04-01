// Package postgres implements the repository layer for customer-nfs-service.
//
// FR-NFS-006: Mobile Change (OTP-based verification)
// BR-NFS-016: Audit trail for all operations
//
// BATCH NOTE:
//
//	CreateMobileChangeDetail → single TX batch:
//	  INSERT mobile_change_detail + INSERT audit_log
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

// MobileChangeRepository handles all DB operations for nfs.mobile_change_detail.
type MobileChangeRepository struct {
	db  *dblib.DB
	cfg *config.Config
}

// NewMobileChangeRepository constructs the repository.
func NewMobileChangeRepository(db *dblib.DB, cfg *config.Config) *MobileChangeRepository {
	return &MobileChangeRepository{db: db, cfg: cfg}
}

// ---------------------------------------------------------------------------
// CreateMobileChangeDetail creates mobile_change_detail + audit_log in a single batched transaction.
//
// BATCH: 2 INSERTs in one pgx TX batch.
// FR-NFS-006, BR-NFS-016
// ---------------------------------------------------------------------------
func (r *MobileChangeRepository) CreateMobileChangeDetail(
	ctx context.Context,
	detail *domain.MobileChangeDetail,
	audit *domain.AuditLog,
) error {
	timeout := r.cfg.GetDuration("db.QueryTimeoutMed")
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return r.db.WithTx(ctx, func(tx pgx.Tx) error {
		batch := &pgx.Batch{}

		// 1. INSERT mobile_change_detail
		mobileSQL, mobileArgs, err := dblib.Psql.Insert("nfs.mobile_change_detail").
			Columns(
				"detail_id", "request_id", "old_mobile_number", "new_mobile_number",
				"aadhaar_txn_id", "created_by",
			).
			Values(
				detail.DetailID, detail.RequestID, detail.OldMobileNumber, detail.NewMobileNumber,
				detail.AadhaarTxnID, detail.CreatedBy,
			).
			Suffix("RETURNING detail_id, created_at").
			ToSql()
		if err != nil {
			return fmt.Errorf("build mobile_change_detail insert: %w", err)
		}
		batch.Queue(mobileSQL, mobileArgs...)

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

		// Result 1: mobile_change_detail
		if err := br.QueryRow().Scan(&detail.DetailID, &detail.CreatedAt); err != nil {
			br.Close()
			return fmt.Errorf("scan mobile_change_detail result: %w", err)
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
// GetByRequestID retrieves mobile_change_detail for a given request_id.
// Used by UpdateMobileData activity to fetch new mobile number.
// ---------------------------------------------------------------------------
func (r *MobileChangeRepository) GetByRequestID(ctx context.Context, requestID string) (*domain.MobileChangeDetail, error) {
	timeout := r.cfg.GetDuration("db.QueryTimeoutLow")
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	query := dblib.Psql.Select(
		"detail_id", "request_id", "old_mobile_number", "new_mobile_number",
		"aadhaar_txn_id", "created_at", "updated_at", "created_by", "updated_by", "version",
	).
		From("nfs.mobile_change_detail").
		Where(sq.Eq{"request_id": requestID}).
		Limit(1)

	sql, args, err := query.ToSql()
	if err != nil {
		return nil, fmt.Errorf("build mobile_change_detail select: %w", err)
	}

	var detail domain.MobileChangeDetail
	row := r.db.QueryRow(ctx, sql, args...)
	if err := row.Scan(
		&detail.DetailID, &detail.RequestID, &detail.OldMobileNumber, &detail.NewMobileNumber,
		&detail.AadhaarTxnID, &detail.CreatedAt, &detail.UpdatedAt, &detail.CreatedBy, &detail.UpdatedBy, &detail.Version,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		log.Error(ctx, "GetByRequestID scan: %v", err)
		return nil, fmt.Errorf("get mobile_change_detail: %w", err)
	}

	return &detail, nil
}

// ---------------------------------------------------------------------------
// UpdateAadhaarTxnID updates the aadhaar_txn_id for a mobile change detail.
// Called after OTP verification to store the transaction ID.
// ---------------------------------------------------------------------------
func (r *MobileChangeRepository) UpdateAadhaarTxnID(
	ctx context.Context,
	requestID string,
	aadhaarTxnID string,
	updatedBy string,
) error {
	timeout := r.cfg.GetDuration("db.QueryTimeoutLow")
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	query := dblib.Psql.Update("nfs.mobile_change_detail").
		Set("aadhaar_txn_id", aadhaarTxnID).
		Set("updated_by", updatedBy).
		Set("updated_at", "NOW()").
		Where(sq.Eq{"request_id": requestID})

	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("build mobile_change_detail update: %w", err)
	}

	_, err = r.db.Exec(ctx, sql, args...)
	if err != nil {
		log.Error(ctx, "UpdateAadhaarTxnID exec: %v", err)
		return fmt.Errorf("update mobile_change_detail aadhaar_txn_id: %w", err)
	}

	return nil
}

// ---------------------------------------------------------------------------
// CreateMobileVersionHistory creates a new version in mobile_version_history.
// Called after successful mobile update to maintain version history.
// ---------------------------------------------------------------------------
func (r *MobileChangeRepository) CreateMobileVersionHistory(
	ctx context.Context,
	history *domain.MobileVersionHistory,
) error {
	timeout := r.cfg.GetDuration("db.QueryTimeoutLow")
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	query := dblib.Psql.Insert("nfs.mobile_version_history").
		Columns(
			"version_id", "customer_id", "request_id", "mobile_number",
			"version_number", "is_active", "effective_from", "effective_to", "created_by",
		).
		Values(
			history.VersionID, history.CustomerID, history.RequestID, history.MobileNumber,
			history.VersionNumber, history.IsActive, history.EffectiveFrom, history.EffectiveTo, history.CreatedBy,
		)

	sql, args, err := query.ToSql()
	if err != nil {
		return fmt.Errorf("build mobile_version_history insert: %w", err)
	}

	_, err = r.db.Exec(ctx, sql, args...)
	if err != nil {
		log.Error(ctx, "CreateMobileVersionHistory exec: %v", err)
		return fmt.Errorf("create mobile_version_history: %w", err)
	}

	return nil
}
