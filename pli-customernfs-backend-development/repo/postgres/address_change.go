// Package postgres — address_change repository.
//
// FR-NFS-001: Address change request processing
// BR-NFS-001: Aadhaar-based address change (immediate, via WF-NFS-001)
// BR-NFS-002: Manual address change (CPC approval, via WF-NFS-002)
// BR-NFS-003: Permanent address change requires gazette/newspaper notification
// BR-NFS-004: Address change for joint policies triggers cross-policy update
// BR-NFS-005: Only INSURED role handled by Customer NFS (PROPOSER → Policy Admin)
// BR-NFS-006: Pincode must match selected state
// VR-NFS-001..005: Address field validation
// VR-NFS-013: Pincode format (6 digits)
//
// BATCH NOTE:
//
//	CreateAddressVersion → single TX batch: deactivate old + insert new version.
//	GetWithRequest → single pgx.Batch: service_request + address_change_detail.
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

// AddressChangeRepository handles DB operations for nfs.address_change_detail (E-2)
// and nfs.address_version_history.
type AddressChangeRepository struct {
	db  *dblib.DB
	cfg *config.Config
}

// NewAddressChangeRepository constructs the repository.
func NewAddressChangeRepository(db *dblib.DB, cfg *config.Config) *AddressChangeRepository {
	return &AddressChangeRepository{db: db, cfg: cfg}
}

// ---------------------------------------------------------------------------
// GetByRequestID returns the address change detail for a service request.
// E-2
// ---------------------------------------------------------------------------
func (r *AddressChangeRepository) GetByRequestID(ctx context.Context, requestID string) (*domain.AddressChangeDetail, error) {
	timeout := r.cfg.GetDuration("db.QueryTimeoutLow")
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	builder := dblib.Psql.Select(
		"detail_id", "request_id", "address_update_for", "address_type",
		"old_address_line1", "old_address_line2", "old_village", "old_taluka",
		"old_city", "old_district", "old_state", "old_pincode",
		"new_address_line1", "new_address_line2", "new_village", "new_taluka",
		"new_city", "new_district", "new_state", "new_pincode",
		"aadhaar_txn_id", "old_address_version", "new_address_version",
		"created_at", "updated_at", "version",
	).
		From("nfs.address_change_detail").
		Where(sq.And{
			sq.Eq{"request_id": requestID},
			sq.Eq{"deleted_at": nil},
		})

	result, err := dblib.SelectOne(ctx, r.db, builder, pgx.RowToStructByName[domain.AddressChangeDetail])
	if err != nil {
		log.Error(ctx, "AddressChange.GetByRequestID [%s]: %v", requestID, err)
		return nil, err
	}
	return &result, nil
}

// ---------------------------------------------------------------------------
// GetWithServiceRequest fetches service_request + address_change_detail in a
// single batch (two SELECTs in one round-trip).
//
// BATCH: 2 SELECT queries batched for one network round-trip.
// ST-001 (FR-NFS-012 — fetch full request detail)
// ---------------------------------------------------------------------------
// func (r *AddressChangeRepository) GetWithServiceRequest(
// 	ctx context.Context,
// 	requestID string,
// ) (*domain.ServiceRequest, *domain.AddressChangeDetail, error) {
// 	timeout := r.cfg.GetDuration("db.QueryTimeoutMed")
// 	ctx, cancel := context.WithTimeout(ctx, timeout)
// 	defer cancel()

// 	srBuilder := dblib.Psql.Select(
// 		"request_id", "ticket_number", "customer_id", "policy_number",
// 		"request_type", "auth_method", "status", "channel", "office_code",
// 		"initiated_by", "assigned_to", "approved_by",
// 		"created_at", "updated_at", "completed_at", "approval_date",
// 		"sla_deadline", "sla_breached", "rejection_reason",
// 		"partial_processing_flag", "workflow_id", "workflow_run_id", "version",
// 	).From("nfs.service_request").
// 		Where(sq.And{sq.Eq{"request_id": requestID}, sq.Eq{"deleted_at": nil}})

// 	addrBuilder := dblib.Psql.Select(
// 		"detail_id", "request_id", "address_update_for", "address_type",
// 		"old_address_line1", "old_city", "old_district", "old_state", "old_pincode",
// 		"new_address_line1", "new_address_line2", "new_village", "new_taluka",
// 		"new_city", "new_district", "new_state", "new_pincode",
// 		"aadhaar_txn_id", "old_address_version", "new_address_version", "created_at", "version",
// 	).From("nfs.address_change_detail").
// 		Where(sq.And{sq.Eq{"request_id": requestID}, sq.Eq{"deleted_at": nil}})

// 	// BATCH: both queries in one round-trip
// 	batch := &pgx.Batch{}
// 	var sr domain.ServiceRequest
// 	var addr domain.AddressChangeDetail

// 	if err := dblib.QueueReturnRow(batch, srBuilder, pgx.RowToStructByName[domain.ServiceRequest], &sr); err != nil {
// 		return nil, nil, fmt.Errorf("queue service_request: %w", err)
// 	}
// 	if err := dblib.QueueReturnRow(batch, addrBuilder, pgx.RowToStructByName[domain.AddressChangeDetail], &addr); err != nil {
// 		return nil, nil, fmt.Errorf("queue address_change_detail: %w", err)
// 	}

//		br := r.db.SendBatch(ctx, batch)
//		if err := br.Close(); err != nil {
//			log.Error(ctx, "GetWithServiceRequest [%s]: %v", requestID, err)
//			return nil, nil, err
//		}
//		return &sr, &addr, nil
//	}
func (r *AddressChangeRepository) GetWithServiceRequest(
	ctx context.Context,
	requestID string,
) (*domain.ServiceRequest, *domain.AddressChangeDetail, error) {
	timeout := r.cfg.GetDuration("db.QueryTimeoutMed")
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	srBuilder := dblib.Psql.Select(
		"request_id", "ticket_number", "customer_id", "policy_number",
		"request_type", "auth_method", "status", "previous_status",
		"channel", "office_code", "initiated_by", "assigned_to", "approved_by",
		"created_at", "updated_at", "completed_at", "approval_date",
		"sla_deadline", "sla_breached", "rejection_reason",
		"partial_processing_flag", "guardian_approval_required",
		"workflow_id", "workflow_run_id",
		"created_by", "updated_by", "deleted_at", "version", "metadata",
	).From("nfs.service_request").
		Where(sq.And{sq.Eq{"request_id": requestID}, sq.Eq{"deleted_at": nil}})

	addrBuilder := dblib.Psql.Select(
		"detail_id", "request_id", "address_update_for", "address_type",
		"old_address_line1", "old_address_line2", "old_village", "old_taluka",
		"old_city", "old_district", "old_state", "old_pincode",
		"new_address_line1", "new_address_line2", "new_village", "new_taluka",
		"new_city", "new_district", "new_state", "new_pincode",
		"aadhaar_txn_id", "old_address_version", "new_address_version",
		"created_at", "updated_at", "created_by", "updated_by", "version",
	).From("nfs.address_change_detail").
		Where(sq.And{sq.Eq{"request_id": requestID}, sq.Eq{"deleted_at": nil}})
	sr, err := dblib.SelectOne(ctx, r.db, srBuilder, pgx.RowToStructByName[domain.ServiceRequest])
	if err != nil {
		return nil, nil, fmt.Errorf("fetch service_request: %w", err)
	}

	addr, err := dblib.SelectOne(ctx, r.db, addrBuilder, pgx.RowToStructByName[domain.AddressChangeDetail])
	if err != nil {
		return nil, nil, fmt.Errorf("fetch address_change_detail: %w", err)
	}

	return &sr, &addr, nil
}

// ---------------------------------------------------------------------------
// UpdateAadhaarTxnID stores the UIDAI transaction reference after OTP verify.
// WF-NFS-001 (Aadhaar-based address change)
// ---------------------------------------------------------------------------
func (r *AddressChangeRepository) UpdateAadhaarTxnID(ctx context.Context, requestID, aadhaarTxnID string) error {
	timeout := r.cfg.GetDuration("db.QueryTimeoutLow")
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	builder := dblib.Psql.Update("nfs.address_change_detail").
		Set("aadhaar_txn_id", aadhaarTxnID).
		Where(sq.And{sq.Eq{"request_id": requestID}, sq.Eq{"deleted_at": nil}})

	if _, err := dblib.Update(ctx, r.db, builder); err != nil {
		log.Error(ctx, "UpdateAadhaarTxnID [%s]: %v", requestID, err)
		return fmt.Errorf("update aadhaar_txn_id: %w", err)
	}
	return nil
}

// ---------------------------------------------------------------------------
// CreateAddressVersion records a new address version and deactivates the
// previous active version. Executed inside a batched TX.
//
// BATCH: UPDATE old is_active=false + INSERT new version in one TX batch.
// BR-NFS-001, BR-NFS-002 (address version tracking for rollback)
// ---------------------------------------------------------------------------
func (r *AddressChangeRepository) CreateAddressVersion(
	ctx context.Context,
	customerID string,
	addressType string,
	newVersion *domain.AddressVersionHistory,
) error {
	timeout := r.cfg.GetDuration("db.QueryTimeoutMed")
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return r.db.WithTx(ctx, func(tx pgx.Tx) error {
		batch := &pgx.Batch{}

		// 1. Deactivate current active version for this customer + address type
		// (may affect 0 rows for the very first version)
		deactivateB := dblib.Psql.Update("nfs.address_version_history").
			Set("is_active", false).
			Set("effective_to", sq.Expr("NOW()")).
			Where(sq.And{
				sq.Eq{"customer_id": customerID},
				sq.Eq{"address_type": addressType},
				sq.Eq{"is_active": true},
			})
		if err := dblib.QueueExecRow(batch, deactivateB); err != nil {
			return fmt.Errorf("queue deactivate address version: %w", err)
		}

		// 2. INSERT new active version
		insertB := dblib.Psql.Insert("nfs.address_version_history").
			Columns(
				"version_id", "customer_id", "request_id", "address_type",
				"address_line1", "address_line2", "village", "taluka",
				"city", "district", "state", "pincode",
				"version_number", "is_active", "effective_from", "created_by",
			).
			Values(
				newVersion.VersionID, newVersion.CustomerID, newVersion.RequestID, newVersion.AddressType,
				newVersion.AddressLine1, newVersion.AddressLine2, newVersion.Village, newVersion.Taluka,
				newVersion.City, newVersion.District, newVersion.State, newVersion.Pincode,
				newVersion.VersionNumber, true, sq.Expr("NOW()"), newVersion.CreatedBy,
			).
			Suffix("RETURNING *")
		if err := dblib.QueueReturnRow(batch, insertB, pgx.RowToStructByName[domain.AddressVersionHistory], newVersion); err != nil {
			return fmt.Errorf("queue insert address version: %w", err)
		}

		br := tx.SendBatch(ctx, batch)
		if err := br.Close(); err != nil {
			return fmt.Errorf("exec create address version batch: %w", err)
		}
		return nil
	})
}

// ---------------------------------------------------------------------------
// GetActiveVersion returns the currently active address version for a
// customer and address type.
// BR-NFS-001, BR-NFS-002 (retrieve current address for old_* fields)
// ---------------------------------------------------------------------------
func (r *AddressChangeRepository) GetActiveVersion(
	ctx context.Context,
	customerID, addressType string,
) (*domain.AddressVersionHistory, error) {
	timeout := r.cfg.GetDuration("db.QueryTimeoutLow")
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	builder := dblib.Psql.Select(
		"version_id", "customer_id", "request_id", "address_type",
		"address_line1", "address_line2", "village", "taluka",
		"city", "district", "state", "pincode",
		"version_number", "is_active", "effective_from", "created_at",
	).
		From("nfs.address_version_history").
		Where(sq.And{
			sq.Eq{"customer_id": customerID},
			sq.Eq{"address_type": addressType},
			sq.Eq{"is_active": true},
		}).
		Limit(1)

	result, err := dblib.SelectOne(ctx, r.db, builder, pgx.RowToStructByName[domain.AddressVersionHistory])
	if err != nil {
		log.Error(ctx, "GetActiveVersion [%s/%s]: %v", customerID, addressType, err)
		return nil, err
	}
	return &result, nil
}

// ---------------------------------------------------------------------------
// ListVersionHistory returns all address versions for a customer in
// chronological order.
// ST-002 (FR-NFS-012 — timeline / history display)
// ---------------------------------------------------------------------------
func (r *AddressChangeRepository) ListVersionHistory(
	ctx context.Context,
	customerID, addressType string,
) ([]domain.AddressVersionHistory, error) {
	timeout := r.cfg.GetDuration("db.QueryTimeoutLow")
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	builder := dblib.Psql.Select(
		"version_id", "customer_id", "request_id", "address_type",
		"address_line1", "city", "district", "state", "pincode",
		"version_number", "is_active", "effective_from", "effective_to", "created_at",
	).
		From("nfs.address_version_history").
		Where(sq.And{
			sq.Eq{"customer_id": customerID},
			sq.Eq{"address_type": addressType},
		}).
		OrderBy("version_number ASC")

	results, err := dblib.SelectRows(ctx, r.db, builder, pgx.RowToStructByName[domain.AddressVersionHistory])
	if err != nil {
		log.Error(ctx, "ListVersionHistory [%s/%s]: %v", customerID, addressType, err)
		return nil, err
	}
	return results, nil
}

// ---------------------------------------------------------------------------
// ListStates returns master-data states.
// LU-001 (FR-NFS-001 — address lookup).
// ---------------------------------------------------------------------------
func (r *AddressChangeRepository) ListStates(ctx context.Context, activeOnly bool) ([]domain.State, error) {
	timeout := r.cfg.GetDuration("db.QueryTimeoutLow")
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	builder := dblib.Psql.Select("state_code", "state_name", "circle_code", "is_active", "created_at").
		From("nfs.ref_state").
		OrderBy("state_name ASC")
	if activeOnly {
		builder = builder.Where(sq.Eq{"is_active": true})
	}

	results, err := dblib.SelectRows(ctx, r.db, builder, pgx.RowToStructByName[domain.State])
	if err != nil {
		log.Error(ctx, "ListStates: %v", err)
		return nil, err
	}
	return results, nil
}

// ---------------------------------------------------------------------------
// LookupPincode returns geo data for a pincode.
// LU-002 / VA-001 (BR-NFS-006: pincode ↔ state cross-validation).
// ---------------------------------------------------------------------------
func (r *AddressChangeRepository) LookupPincode(ctx context.Context, pincode string) (*domain.PincodeData, error) {
	timeout := r.cfg.GetDuration("db.QueryTimeoutLow")
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	builder := dblib.Psql.Select(
		"pincode", "state_code", "state_name", "district", "city", "is_active",
	).
		From("nfs.ref_pincode").
		Where(sq.And{
			sq.Eq{"pincode": pincode},
			sq.Eq{"is_active": true},
		})

	result, err := dblib.SelectOne(ctx, r.db, builder, pgx.RowToStructByName[domain.PincodeData])
	if err != nil {
		log.Error(ctx, "LookupPincode [%s]: %v", pincode, err)
		return nil, err
	}
	return &result, nil
}

// ---------------------------------------------------------------------------
// ListDocumentTypes returns the document types required for a request type.
// LU-005 (FR-NFS-002 / FR-NFS-005 — document requirements lookup).
// ---------------------------------------------------------------------------
func (r *AddressChangeRepository) ListDocumentTypes(ctx context.Context, requestType string) ([]domain.DocumentType, error) {
	timeout := r.cfg.GetDuration("db.QueryTimeoutLow")
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	builder := dblib.Psql.Select(
		"doc_type_code", "doc_type_name", "request_type", "is_mandatory", "description", "is_active",
	).
		From("nfs.ref_document_type").
		Where(sq.And{
			sq.Eq{"request_type": requestType},
			sq.Eq{"is_active": true},
		}).
		OrderBy("is_mandatory DESC", "doc_type_name ASC")

	results, err := dblib.SelectRows(ctx, r.db, builder, pgx.RowToStructByName[domain.DocumentType])
	if err != nil {
		log.Error(ctx, "ListDocumentTypes [%s]: %v", requestType, err)
		return nil, err
	}
	return results, nil
}

// ---------------------------------------------------------------------------
// ListOffices returns CPC/RO/HO offices, optionally filtered by type and circle.
// LU-006 (FR-NFS-001 — office selection).
// ---------------------------------------------------------------------------
func (r *AddressChangeRepository) ListOffices(ctx context.Context, officeType, circleCode *string) ([]domain.Office, error) {
	timeout := r.cfg.GetDuration("db.QueryTimeoutLow")
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	builder := dblib.Psql.Select(
		"office_code", "office_name", "office_type", "circle_code",
		"state_code", "address", "pincode", "is_active",
	).
		From("nfs.ref_office").
		Where(sq.Eq{"is_active": true}).
		OrderBy("office_name ASC")

	if officeType != nil {
		builder = builder.Where(sq.Eq{"office_type": *officeType})
	}
	if circleCode != nil {
		builder = builder.Where(sq.Eq{"circle_code": *circleCode})
	}

	results, err := dblib.SelectRows(ctx, r.db, builder, pgx.RowToStructByName[domain.Office])
	if err != nil {
		log.Error(ctx, "ListOffices: %v", err)
		return nil, err
	}
	return results, nil
}

// ---------------------------------------------------------------------------
// ListStatusReasons returns the configurable reason dropdown for a given action.
// LU-008 (FR-NFS-009 — rejection/send-back reasons).
// ---------------------------------------------------------------------------
func (r *AddressChangeRepository) ListStatusReasons(ctx context.Context, actionType *string) ([]domain.StatusReason, error) {
	timeout := r.cfg.GetDuration("db.QueryTimeoutLow")
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	builder := dblib.Psql.Select("reason_code", "reason_text", "action_type", "is_active").
		From("nfs.ref_status_reason").
		Where(sq.Eq{"is_active": true}).
		OrderBy("action_type ASC", "reason_text ASC")

	if actionType != nil {
		builder = builder.Where(sq.Eq{"action_type": *actionType})
	}

	results, err := dblib.SelectRows(ctx, r.db, builder, pgx.RowToStructByName[domain.StatusReason])
	if err != nil {
		log.Error(ctx, "ListStatusReasons: %v", err)
		return nil, err
	}
	return results, nil
}
