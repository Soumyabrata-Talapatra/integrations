// Package postgres provides PostgreSQL repository implementations for the Customer NFS service.
// Phase: Phase 2 — Name Change
//
// BATCH NOTE: All multi-query operations use pgx.Batch for single round-trip efficiency.
// WORKFLOW STATE: workflow_id/workflow_run_id stored in service_request for HTTP signaling.
package postgres

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"

	dblib "gitlab.cept.gov.in/it-2.0-common/n-api-db"

	"customer-nfs-service/core/domain"
)

// NameChangeRepository handles all DB operations for nfs.name_change_detail.
// Phase 2: Name Change (CORE-005..008)
//
// Business Rules:
//
//	BR-NFS-007: Aadhaar path — aadhaar_txn_id stored when OTP verified
//	BR-NFS-008: Manual path — at least one legal document required
//	BR-NFS-009: On completion — policies_affected count updated
//	BR-NFS-010: DOB is immutable — never written to name_change_detail
type NameChangeRepository struct {
	db *dblib.DB
}

// NewNameChangeRepository creates a new NameChangeRepository.
func NewNameChangeRepository(db *dblib.DB) *NameChangeRepository {
	return &NameChangeRepository{db: db}
}

// GetByRequestID returns the name change detail for a given request.
// FR-NFS-004, FR-NFS-005
func (r *NameChangeRepository) GetByRequestID(ctx context.Context, requestID string) (*domain.NameChangeDetail, error) {
	b := dblib.Psql.Select(
		"detail_id", "request_id",
		"old_salutation", "old_first_name", "old_middle_name", "old_last_name",
		"new_salutation", "new_first_name", "new_middle_name", "new_last_name",
		"policies_affected", "aadhaar_txn_id",
		"old_name_version", "new_name_version",
		"created_at", "updated_at", "created_by", "updated_by", "version",
	).From("nfs.name_change_detail").Where(sq.Eq{"request_id": requestID})

	result, err := dblib.SelectOne(ctx, r.db, b, pgx.RowToStructByName[domain.NameChangeDetail])
	if err != nil {
		return nil, fmt.Errorf("NameChangeRepository.GetByRequestID: %w", err)
	}
	return &result, nil
}

// GetWithServiceRequest returns name change detail + service request in a single BATCH round-trip.
// BATCH: 2 SELECT queries — detail + service_request.
// FR-NFS-004, FR-NFS-005: used by workflow activities needing both records.
func (r *NameChangeRepository) GetWithServiceRequest(ctx context.Context, requestID string) (*domain.NameChangeDetail, *domain.ServiceRequest, error) {
	detailBuilder := dblib.Psql.Select(
		"detail_id", "request_id",
		"old_salutation", "old_first_name", "old_middle_name", "old_last_name",
		"new_salutation", "new_first_name", "new_middle_name", "new_last_name",
		"policies_affected", "aadhaar_txn_id",
		"old_name_version", "new_name_version",
		"created_at", "updated_at", "created_by", "updated_by", "version",
	).From("nfs.name_change_detail").Where(sq.Eq{"request_id": requestID})

	srBuilder := dblib.Psql.Select(
		"request_id", "ticket_number", "customer_id", "policy_number",
		"request_type", "auth_method", "status", "previous_status",
		"channel", "office_code", "initiated_by", "assigned_to", "approved_by",
		"created_at", "updated_at", "completed_at", "approval_date",
		"sla_deadline", "sla_breached", "rejection_reason",
		"partial_processing_flag", "guardian_approval_required",
		"workflow_id", "workflow_run_id",
		"created_by", "updated_by", "deleted_at", "version", "metadata",
	).From("nfs.service_request").Where(sq.Eq{"request_id": requestID})

	detail, err := dblib.SelectOne(ctx, r.db, detailBuilder, pgx.RowToStructByName[domain.NameChangeDetail])
	if err != nil {
		return nil, nil, fmt.Errorf("fetch name_change_detail: %w", err)
	}

	sr, err := dblib.SelectOne(ctx, r.db, srBuilder, pgx.RowToStructByName[domain.ServiceRequest])
	if err != nil {
		return nil, nil, fmt.Errorf("fetch service_request: %w", err)
	}

	return &detail, &sr, nil
}

// func (r *NameChangeRepository) GetWithServiceRequest(ctx context.Context, requestID string) (*domain.NameChangeDetail, *domain.ServiceRequest, error) {
// 	detailBuilder := dblib.Psql.Select(
// 		"detail_id", "request_id",
// 		"old_salutation", "old_first_name", "old_middle_name", "old_last_name",
// 		"new_salutation", "new_first_name", "new_middle_name", "new_last_name",
// 		"policies_affected", "aadhaar_txn_id",
// 		"old_name_version", "new_name_version",
// 		"created_at", "updated_at", "created_by", "updated_by", "version",
// 	).From("nfs.name_change_detail").Where(sq.Eq{"request_id": requestID})

// 	srBuilder := dblib.Psql.Select(
// 		"request_id", "customer_id", "policy_number", "request_type",
// 		"ticket_number", "status", "auth_method", "channel", "office_code",
// 		"workflow_id", "workflow_run_id",
// 		"created_at", "updated_at", "created_by", "updated_by", "version",
// 	).From("nfs.service_request").Where(sq.Eq{"request_id": requestID})

// 	batch := &pgx.Batch{}
// 	var detail domain.NameChangeDetail
// 	var sr domain.ServiceRequest

// 	if err := dblib.QueueReturnRow(batch, detailBuilder, pgx.RowToStructByName[domain.NameChangeDetail], &detail); err != nil {
// 		return nil, nil, fmt.Errorf("queue detail query: %w", err)
// 	}
// 	if err := dblib.QueueReturnRow(batch, srBuilder, pgx.RowToStructByName[domain.ServiceRequest], &sr); err != nil {
// 		return nil, nil, fmt.Errorf("queue sr query: %w", err)
// 	}

// 	br := r.db.SendBatch(ctx, batch)
// 	if err := br.Close(); err != nil {
// 		return nil, nil, fmt.Errorf("NameChangeRepository.GetWithServiceRequest: %w", err)
// 	}
// 	return &detail, &sr, nil
// }

// UpdateAadhaarTxnID stores the Aadhaar transaction ID after OTP is sent.
// BR-NFS-007: Aadhaar path — txn_id required for subsequent OTP verification.
func (r *NameChangeRepository) UpdateAadhaarTxnID(ctx context.Context, requestID, txnID string) error {
	b := dblib.Psql.Update("nfs.name_change_detail").
		Set("aadhaar_txn_id", txnID).
		Set("updated_at", sq.Expr("NOW()")).
		Set("version", sq.Expr("version + 1")).
		Where(sq.Eq{"request_id": requestID})

	_, err := dblib.Update(ctx, r.db, b)
	if err != nil {
		return fmt.Errorf("NameChangeRepository.UpdateAadhaarTxnID: %w", err)
	}
	return nil
}

// CreateNameVersion creates a new name version history entry atomically.
// BATCH (TX): 2-op batch — deactivate current + insert new version.
//
// BR-NFS-009: Name versioning — old version deactivated, new version activated.
// BR-NFS-010: DOB is never written here (immutable).
// UpdateIdentityActivity calls this to record the approved name change.
func (r *NameChangeRepository) CreateNameVersion(ctx context.Context, requestID string, version *domain.NameVersionHistory) (*domain.NameVersionHistory, error) {
	var result domain.NameVersionHistory

	err := r.db.WithTx(ctx, func(tx pgx.Tx) error {
		// Step 1: Deactivate current active version.
		deactivateBuilder := dblib.Psql.Update("nfs.name_version_history").
			Set("is_current", false).
			Set("deactivated_at", sq.Expr("NOW()")).
			Where(sq.And{sq.Eq{"request_id": requestID}, sq.Eq{"is_current": true}})

			// Step 2: Insert new version with RETURNING.
			// insertBuilder := dblib.Psql.Insert("nfs.name_version_history").
			// 	Columns(
			// 		"request_id", "salutation", "first_name", "middle_name", "last_name",
			// 		"is_current", "effective_from", "created_by",
			// 	).
			// 	Values(
			// 		requestID, version.Salutation, version.FirstName, version.MiddleName, version.LastName,
			// 		true, sq.Expr("NOW()"), version.CreatedBy,
			// 	).
			// 	Suffix("RETURNING version_id, request_id, salutation, first_name, middle_name, last_name, source_type, aadhaar_txn_id, is_current, effective_from, created_by, created_at")
		insertBuilder := dblib.Psql.Insert("nfs.name_version_history").
			Columns(
				"customer_id", "request_id", "salutation", "first_name", "middle_name", "last_name",
				"version_number", "is_current", "is_active", "effective_from", "created_by",
			).
			Values(
				version.CustomerID, requestID, version.Salutation, version.FirstName, version.MiddleName, version.LastName,
				1, true, true, sq.Expr("NOW()"), version.CreatedBy,
			).
			Suffix("RETURNING version_id, customer_id, request_id, salutation, first_name, middle_name, last_name, version_number, is_active, is_current, effective_from, effective_to, created_by, created_at")
		batch := &pgx.Batch{}
		if err := dblib.QueueExecRow(batch, deactivateBuilder); err != nil {
			return fmt.Errorf("queue deactivate: %w", err)
		}
		if err := dblib.QueueReturnRow(batch, insertBuilder, pgx.RowToStructByName[domain.NameVersionHistory], &result); err != nil {
			return fmt.Errorf("queue insert name version: %w", err)
		}

		br := tx.SendBatch(ctx, batch)
		if err := br.Close(); err != nil {
			return fmt.Errorf("CreateNameVersion batch: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("NameChangeRepository.CreateNameVersion: %w", err)
	}
	return &result, nil
}

// UpdatePoliciesAffected records how many policies were updated after name propagation.
// BR-NFS-009: policies_affected count set after customer.name.updated event consumed.
func (r *NameChangeRepository) UpdatePoliciesAffected(ctx context.Context, requestID string, count int) error {
	b := dblib.Psql.Update("nfs.name_change_detail").
		Set("policies_affected", count).
		Set("updated_at", sq.Expr("NOW()")).
		Set("version", sq.Expr("version + 1")).
		Where(sq.Eq{"request_id": requestID})

	_, err := dblib.Update(ctx, r.db, b)
	if err != nil {
		return fmt.Errorf("NameChangeRepository.UpdatePoliciesAffected: %w", err)
	}
	return nil
}

// GetActiveVersion returns the currently active name version for historical reference.
// Used by compensation workflows (Section 14.2) to restore prior name on rollback.
func (r *NameChangeRepository) GetActiveVersion(ctx context.Context, requestID string) (*domain.NameVersionHistory, error) {
	b := dblib.Psql.Select(
		"version_id", "request_id", "salutation", "first_name", "middle_name", "last_name",
		"source_type", "aadhaar_txn_id", "is_current", "effective_from", "deactivated_at",
		"created_by", "created_at",
	).From("nfs.name_version_history").
		Where(sq.And{sq.Eq{"request_id": requestID}, sq.Eq{"is_current": true}})

	result, err := dblib.SelectOne(ctx, r.db, b, pgx.RowToStructByName[domain.NameVersionHistory])
	if err != nil {
		return nil, fmt.Errorf("NameChangeRepository.GetActiveVersion: %w", err)
	}
	return &result, nil
}

// ListVersionHistory returns all name versions for a request in chronological order.
// Used by ST-001 (GetRequestDetail with include_audit=true) and compensation checks.
func (r *NameChangeRepository) ListVersionHistory(ctx context.Context, requestID string) ([]domain.NameVersionHistory, error) {
	b := dblib.Psql.Select(
		"version_id", "request_id", "salutation", "first_name", "middle_name", "last_name",
		"source_type", "aadhaar_txn_id", "is_current", "effective_from", "deactivated_at",
		"created_by", "created_at",
	).From("nfs.name_version_history").
		Where(sq.Eq{"request_id": requestID}).
		OrderBy("effective_from DESC")

	results, err := dblib.SelectRows(ctx, r.db, b, pgx.RowToStructByName[domain.NameVersionHistory])
	if err != nil {
		return nil, fmt.Errorf("NameChangeRepository.ListVersionHistory: %w", err)
	}
	return results, nil
}
