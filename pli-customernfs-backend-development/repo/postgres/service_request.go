// Package postgres implements the repository layer for customer-nfs-service.
//
// FR-NFS-001: Initiate address/name change service requests
// BR-NFS-011: Ticket number generation (NFS-{TYPE}-{YYYYMMDD}-{SEQ6})
// BR-NFS-012: Status state-machine enforcement
// BR-NFS-013: Withdrawal eligibility (AUTO vs MANUAL)
// VR-NFS-015: Duplicate active request check per customer
//
// BATCH NOTE:
//
//	CreateWithAddressDetail / CreateWithNameDetail → single TX batch:
//	  INSERT service_request + INSERT detail + INSERT audit_log
//	UpdateStatus → single TX batch:
//	  UPDATE service_request + INSERT status_transition_history + INSERT audit_log
//	ListByCustomerID → single pgx.Batch: COUNT(*) + SELECT rows (one round-trip)
//
// WORKFLOW STATE NOTE:
//
//	workflow_id and workflow_run_id are stored at Temporal workflow start.
//	UpdateWorkflowState must be called from the Temporal activity that starts
//	the workflow so subsequent HTTP signal calls can locate the workflow.
package postgres

import (
	"context"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	config "gitlab.cept.gov.in/it-2.0-common/api-config"
	dblib "gitlab.cept.gov.in/it-2.0-common/n-api-db"
	log "gitlab.cept.gov.in/it-2.0-common/n-api-log"

	"customer-nfs-service/core/domain"
)

// psql is the Squirrel placeholder format for PostgreSQL.
// var psql = sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

// newUUID generates a new random UUID string.
func newUUID() string {
	return uuid.New().String()
}

// ServiceRequestRepository handles all DB operations for nfs.service_request (E-1).
type ServiceRequestRepository struct {
	db  *dblib.DB
	cfg *config.Config
}

// NewServiceRequestRepository constructs the repository.
func NewServiceRequestRepository(db *dblib.DB, cfg *config.Config) *ServiceRequestRepository {
	return &ServiceRequestRepository{db: db, cfg: cfg}
}

// ---------------------------------------------------------------------------
// GenerateTicketNumber atomically increments nfs.ticket_sequence and returns
// a formatted ticket string. Uses SELECT FOR UPDATE to avoid race conditions.
// BR-NFS-011
// ---------------------------------------------------------------------------
func (r *ServiceRequestRepository) GenerateTicketNumber(ctx context.Context, requestType string) (string, error) {
	timeout := r.cfg.GetDuration("db.QueryTimeoutLow")
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var ticketNumber string
	err := r.db.WithTx(ctx, func(tx pgx.Tx) error {
		// Lock the current-date row for this request type (BR-NFS-011)
		upsertSQL := `
			INSERT INTO nfs.ticket_sequence (sequence_date, request_type, sequence_value)
			VALUES (CURRENT_DATE, $1, 1)
			ON CONFLICT (sequence_date, request_type)
			DO UPDATE SET sequence_value = ticket_sequence.sequence_value + 1
			RETURNING sequence_value, CURRENT_DATE`

		var seq int
		var today time.Time
		row := tx.QueryRow(ctx, upsertSQL, requestType)
		if err := row.Scan(&seq, &today); err != nil {
			return fmt.Errorf("generate ticket sequence: %w", err)
		}

		typeCode := "ANC"
		if requestType == "NAME_CHANGE" {
			typeCode = "NMC"
		}
		ticketNumber = fmt.Sprintf("NFS-%s-%s-%06d", typeCode, today.Format("20060102"), seq)
		return nil
	})
	if err != nil {
		log.Error(ctx, "GenerateTicketNumber failed: %v", err)
		return "", err
	}
	return ticketNumber, nil
}

// ---------------------------------------------------------------------------
// CheckDuplicateRequest checks whether the customer already has an active
// request of the same type. VR-NFS-015.
// ---------------------------------------------------------------------------
func (r *ServiceRequestRepository) CheckDuplicateRequest(ctx context.Context, customerID int64, requestType string) (bool, string, error) {
	timeout := r.cfg.GetDuration("db.QueryTimeoutLow")
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	query := dblib.Psql.Select("COUNT(*)", "COALESCE(MAX(ticket_number), '')").
		From("nfs.service_request").
		Where(sq.And{
			sq.Eq{"customer_id": customerID},
			sq.Eq{"request_type": requestType},
			sq.Eq{"deleted_at": nil},
			sq.Expr("status IN ('CREATED','PENDING_DOCUMENTS','PENDING_APPROVAL','IN_PROGRESS')"),
		})

	sql, args, err := query.ToSql()
	if err != nil {
		return false, "", fmt.Errorf("build duplicate check query: %w", err)
	}

	var count int
	var existingTicket string
	row := r.db.QueryRow(ctx, sql, args...)
	if err := row.Scan(&count, &existingTicket); err != nil {
		log.Error(ctx, "CheckDuplicateRequest scan: %v", err)
		return false, "", fmt.Errorf("check duplicate request: %w", err)
	}
	return count > 0, existingTicket, nil
}

// ---------------------------------------------------------------------------
// CreateWithAddressDetail creates service_request + address_change_detail +
// audit_log in a single batched transaction.
//
// BATCH: 3 INSERTs in one pgx TX batch → single round-trip.
// FR-NFS-001, BR-NFS-001..006, BR-NFS-016
// ---------------------------------------------------------------------------
func (r *ServiceRequestRepository) CreateWithAddressDetail(
	ctx context.Context,
	sr *domain.ServiceRequest,
	detail *domain.AddressChangeDetail,
	audit *domain.AuditLog,
) error {
	timeout := r.cfg.GetDuration("db.QueryTimeoutMed")
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return r.db.WithTx(ctx, func(tx pgx.Tx) error {
		batch := &pgx.Batch{}

		// 1. INSERT service_request (BR-NFS-001)
		srSQL, srArgs, err := dblib.Psql.Insert("nfs.service_request").
			Columns(
				"request_id", "ticket_number", "customer_id", "policy_number",
				"request_type", "auth_method", "status", "channel", "office_code",
				"initiated_by", "sla_deadline", "created_by", "metadata",
			).
			Values(
				sr.RequestID, sr.TicketNumber, sr.CustomerID, sr.PolicyNumber,
				sr.RequestType, sr.AuthMethod, sr.Status, sr.Channel, sr.OfficeCode,
				sr.InitiatedBy, sr.SLADeadline, sr.CreatedBy, nil,
			).
			Suffix("RETURNING request_id, ticket_number, status, created_at, version").
			ToSql()
		if err != nil {
			return fmt.Errorf("build service_request insert: %w", err)
		}
		batch.Queue(srSQL, srArgs...)

		// 2. INSERT address_change_detail (E-2, BR-NFS-001..006)
		addrSQL, addrArgs, err := dblib.Psql.Insert("nfs.address_change_detail").
			Columns(
				"detail_id", "request_id", "address_update_for", "address_type",
				"old_address_line1", "old_address_line2", "old_village", "old_taluka",
				"old_city", "old_district", "old_state", "old_pincode",
				"new_address_line1", "new_address_line2", "new_village", "new_taluka",
				"new_city", "new_district", "new_state", "new_pincode",
				"aadhaar_txn_id", "created_by",
			).
			Values(
				detail.DetailID, detail.RequestID, detail.AddressUpdateFor, detail.AddressType,
				detail.OldAddressLine1, detail.OldAddressLine2, detail.OldVillage, detail.OldTaluka,
				detail.OldCity, detail.OldDistrict, detail.OldState, detail.OldPincode,
				detail.NewAddressLine1, detail.NewAddressLine2, detail.NewVillage, detail.NewTaluka,
				detail.NewCity, detail.NewDistrict, detail.NewState, detail.NewPincode,
				detail.AadhaarTxnID, detail.CreatedBy,
			).
			Suffix("RETURNING detail_id, created_at").
			ToSql()
		if err != nil {
			return fmt.Errorf("build address_change_detail insert: %w", err)
		}
		batch.Queue(addrSQL, addrArgs...)

		// 3. INSERT audit_log (BR-NFS-016 — every operation must be audited)
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

		// Send all 3 inserts in one network round-trip
		br := tx.SendBatch(ctx, batch)
		// defer br.Close()

		// Result 1: service_request
		if err := br.QueryRow().Scan(
			&sr.RequestID, &sr.TicketNumber, &sr.Status, &sr.CreatedAt, &sr.Version,
		); err != nil {
			br.Close()

			return fmt.Errorf("scan service_request result: %w", err)
		}

		// Result 2: address_change_detail
		if err := br.QueryRow().Scan(&detail.DetailID, &detail.CreatedAt); err != nil {
			br.Close()

			return fmt.Errorf("scan address_change_detail result: %w", err)
		}

		// Result 3: audit_log (exec-only, check RowsAffected)
		if _, err := br.Exec(); err != nil {
			br.Close()

			return fmt.Errorf("exec audit_log insert: %w", err)
		}
		br.Close()

		return nil
	})
}

// ---------------------------------------------------------------------------
// Create creates service_request + audit_log in a single batched transaction.
// Used for mobile/email change requests where detail is added later in workflow.
//
// BATCH: 2 INSERTs in one pgx TX batch.
// FR-NFS-006, FR-NFS-007, BR-NFS-016
// ---------------------------------------------------------------------------
func (r *ServiceRequestRepository) Create(
	ctx context.Context,
	sr *domain.ServiceRequest,
	audit *domain.AuditLog,
) error {
	timeout := r.cfg.GetDuration("db.QueryTimeoutMed")
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return r.db.WithTx(ctx, func(tx pgx.Tx) error {
		batch := &pgx.Batch{}

		// 1. INSERT service_request
		srSQL, srArgs, err := dblib.Psql.Insert("nfs.service_request").
			Columns(
				"request_id", "ticket_number", "customer_id", "policy_number",
				"request_type", "auth_method", "status", "channel", "office_code",
				"initiated_by", "sla_deadline", "created_by", "metadata",
			).
			Values(
				sr.RequestID, sr.TicketNumber, sr.CustomerID, sr.PolicyNumber,
				sr.RequestType, sr.AuthMethod, sr.Status, sr.Channel, sr.OfficeCode,
				sr.InitiatedBy, sr.SLADeadline, sr.CreatedBy, nil,
			).
			Suffix("RETURNING request_id, ticket_number, status, created_at, version").
			ToSql()
		if err != nil {
			return fmt.Errorf("build service_request insert: %w", err)
		}
		batch.Queue(srSQL, srArgs...)

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

		// Result 1: service_request
		if err := br.QueryRow().Scan(
			&sr.RequestID, &sr.TicketNumber, &sr.Status, &sr.CreatedAt, &sr.Version,
		); err != nil {
			br.Close()
			return fmt.Errorf("scan service_request result: %w", err)
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
// CreateWithNameDetail creates service_request + name_change_detail +
// audit_log in a single batched transaction.
//
// BATCH: 3 INSERTs in one pgx TX batch.
// FR-NFS-007, BR-NFS-007..010, BR-NFS-016
// ---------------------------------------------------------------------------
// nullableString returns nil if s is empty, otherwise returns &s.
// Used to avoid inserting empty string into enum columns.
func nullableString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func (r *ServiceRequestRepository) CreateWithNameDetail(
	ctx context.Context,
	sr *domain.ServiceRequest,
	detail *domain.NameChangeDetail,
	audit *domain.AuditLog,
) error {
	timeout := r.cfg.GetDuration("db.QueryTimeoutMed")
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return r.db.WithTx(ctx, func(tx pgx.Tx) error {
		batch := &pgx.Batch{}

		// 1. INSERT service_request
		srSQL, srArgs, err := dblib.Psql.Insert("nfs.service_request").
			Columns(
				"request_id", "ticket_number", "customer_id", "policy_number",
				"request_type", "auth_method", "status", "channel", "office_code",
				"initiated_by", "sla_deadline", "created_by", "metadata",
			).
			Values(
				sr.RequestID, sr.TicketNumber, sr.CustomerID, sr.PolicyNumber,
				sr.RequestType, sr.AuthMethod, sr.Status, sr.Channel, sr.OfficeCode,
				sr.InitiatedBy, sr.SLADeadline, sr.CreatedBy, nil,
			).
			Suffix("RETURNING request_id, ticket_number, status, created_at, version").
			ToSql()
		if err != nil {
			return fmt.Errorf("build service_request insert: %w", err)
		}
		batch.Queue(srSQL, srArgs...)

		// 2. INSERT name_change_detail (E-3, BR-NFS-007..010)
		nameSQL, nameArgs, err := dblib.Psql.Insert("nfs.name_change_detail").
			Columns(
				"detail_id", "request_id",
				"old_salutation", "old_first_name", "old_middle_name", "old_last_name",
				"new_salutation", "new_first_name", "new_middle_name", "new_last_name",
				"aadhaar_txn_id", "created_by",
			).
			Values(
				detail.DetailID, detail.RequestID,
				detail.OldSalutation, detail.OldFirstName, detail.OldMiddleName, detail.OldLastName,
				detail.NewSalutation, detail.NewFirstName, detail.NewMiddleName, detail.NewLastName,
				detail.AadhaarTxnID, detail.CreatedBy,
			).
			Suffix("RETURNING detail_id, created_at").
			ToSql()
		if err != nil {
			return fmt.Errorf("build name_change_detail insert: %w", err)
		}
		batch.Queue(nameSQL, nameArgs...)

		// 3. INSERT audit_log (BR-NFS-016)
		auditSQL, auditArgs, err := dblib.Psql.Insert("nfs.audit_log").
			Columns("audit_id", "request_id", "action_type", "new_value", "performed_by", "channel", "office_code").
			Values(audit.AuditID, audit.RequestID, audit.ActionType, audit.NewValueJSON, audit.PerformedByID, audit.Channel, audit.OfficeCode).
			ToSql()
		if err != nil {
			return fmt.Errorf("build audit_log insert: %w", err)
		}
		batch.Queue(auditSQL, auditArgs...)

		br := tx.SendBatch(ctx, batch)

		if err := br.QueryRow().Scan(&sr.RequestID, &sr.TicketNumber, &sr.Status, &sr.CreatedAt, &sr.Version); err != nil {
			br.Close()
			return fmt.Errorf("scan service_request result: %w", err)
		}
		if err := br.QueryRow().Scan(&detail.DetailID, &detail.CreatedAt); err != nil {
			br.Close()
			return fmt.Errorf("scan name_change_detail result: %w", err)
		}
		if _, err := br.Exec(); err != nil {
			br.Close()
			return fmt.Errorf("exec audit_log insert: %w", err)
		}
		br.Close()
		return nil
	})
}

// func (r *ServiceRequestRepository) CreateWithNameDetail(
// 	ctx context.Context,
// 	sr *domain.ServiceRequest,
// 	detail *domain.NameChangeDetail,
// 	audit *domain.AuditLog,
// ) error {
// 	timeout := r.cfg.GetDuration("db.QueryTimeoutMed")
// 	ctx, cancel := context.WithTimeout(ctx, timeout)
// 	defer cancel()

// 	return r.db.WithTx(ctx, func(tx pgx.Tx) error {
// 		batch := &pgx.Batch{}

// 		// 1. INSERT service_request
// 		srSQL, srArgs, err := dblib.Psql.Insert("nfs.service_request").
// 			Columns(
// 				"request_id", "ticket_number", "customer_id", "policy_number",
// 				"request_type", "auth_method", "status", "channel", "office_code",
// 				"initiated_by", "sla_deadline", "created_by", "metadata",
// 			).
// 			Values(
// 				sr.RequestID, sr.TicketNumber, sr.CustomerID, sr.PolicyNumber,
// 				sr.RequestType, sr.AuthMethod, sr.Status, sr.Channel, sr.OfficeCode,
// 				sr.InitiatedBy, sr.SLADeadline, sr.CreatedBy, nil,
// 			).
// 			Suffix("RETURNING request_id, ticket_number, status, created_at, version").
// 			ToSql()
// 		if err != nil {
// 			return fmt.Errorf("build service_request insert: %w", err)
// 		}
// 		batch.Queue(srSQL, srArgs...)

// 		// 2. INSERT name_change_detail (E-3, BR-NFS-007..010)
// 		nameSQL, nameArgs, err := dblib.Psql.Insert("nfs.name_change_detail").
// 			Columns(
// 				"detail_id", "request_id",
// 				"old_salutation", "old_first_name", "old_middle_name", "old_last_name",
// 				"new_salutation", "new_first_name", "new_middle_name", "new_last_name",
// 				"aadhaar_txn_id", "created_by",
// 			).
// 			Values(
// 				detail.DetailID, detail.RequestID,
// 				detail.OldSalutation, detail.OldFirstName, detail.OldMiddleName, detail.OldLastName,
// 				nullableString(detail.NewSalutation), nullableString(detail.NewFirstName), detail.NewMiddleName, detail.NewLastName,
// 				detail.AadhaarTxnID, detail.CreatedBy,
// 			).
// 			Suffix("RETURNING detail_id, created_at").
// 			ToSql()
// 		if err != nil {
// 			return fmt.Errorf("build name_change_detail insert: %w", err)
// 		}
// 		batch.Queue(nameSQL, nameArgs...)

// 		// 3. INSERT audit_log (BR-NFS-016)
// 		auditSQL, auditArgs, err := dblib.Psql.Insert("nfs.audit_log").
// 			Columns("audit_id", "request_id", "action_type", "new_value", "performed_by", "channel", "office_code").
// 			Values(audit.AuditID, audit.RequestID, audit.ActionType, audit.NewValueJSON, audit.PerformedByID, audit.Channel, audit.OfficeCode).
// 			ToSql()
// 		if err != nil {
// 			return fmt.Errorf("build audit_log insert: %w", err)
// 		}
// 		batch.Queue(auditSQL, auditArgs...)

// 		br := tx.SendBatch(ctx, batch)
// 		defer br.Close()

// 		if err := br.QueryRow().Scan(&sr.RequestID, &sr.TicketNumber, &sr.Status, &sr.CreatedAt, &sr.Version); err != nil {
// 			return fmt.Errorf("scan service_request result: %w", err)
// 		}
// 		if err := br.QueryRow().Scan(&detail.DetailID, &detail.CreatedAt); err != nil {
// 			return fmt.Errorf("scan name_change_detail result: %w", err)
// 		}
// 		if _, err := br.Exec(); err != nil {
// 			return fmt.Errorf("exec audit_log insert: %w", err)
// 		}
// 		return nil
// 	})
// }

// ---------------------------------------------------------------------------
// GetByID fetches a single service request by request_id. BR-NFS-012.
// ---------------------------------------------------------------------------
func (r *ServiceRequestRepository) GetByID(ctx context.Context, requestID string) (*domain.ServiceRequest, error) {
	timeout := r.cfg.GetDuration("db.QueryTimeoutLow")
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	query := dblib.Psql.Select(
		"request_id", "ticket_number", "customer_id", "policy_number",
		"request_type", "auth_method", "status", "previous_status",
		"channel", "office_code", "initiated_by", "assigned_to", "approved_by",
		"created_at", "updated_at", "completed_at", "approval_date",
		"sla_deadline", "sla_breached", "rejection_reason",
		"partial_processing_flag", "guardian_approval_required",
		"workflow_id", "workflow_run_id",
		"created_by", "updated_by", "deleted_at", "version", "metadata",
	).
		From("nfs.service_request").
		Where(sq.And{
			sq.Eq{"request_id": requestID},
			sq.Eq{"deleted_at": nil},
		})

	result, err := dblib.SelectOne(ctx, r.db, query, pgx.RowToStructByName[domain.ServiceRequest])
	if err != nil {
		log.Error(ctx, "GetByID [%s]: %v", requestID, err)
		return nil, err
	}
	return &result, nil
}

// ---------------------------------------------------------------------------
// GetByTicketNumber fetches a service request by its human-readable ticket.
// ST-001 (FR-NFS-012 — status tracking)
// ---------------------------------------------------------------------------
func (r *ServiceRequestRepository) GetByTicketNumber(ctx context.Context, ticketNumber string) (*domain.ServiceRequest, error) {
	timeout := r.cfg.GetDuration("db.QueryTimeoutLow")
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	b := dblib.Psql.Select(
		"request_id", "ticket_number", "customer_id", "policy_number",
		"request_type", "auth_method", "status", "previous_status",
		"channel", "office_code", "initiated_by", "assigned_to", "approved_by",
		"created_at", "updated_at", "completed_at", "approval_date",
		"sla_deadline", "sla_breached", "rejection_reason",
		"partial_processing_flag", "workflow_id", "workflow_run_id",
		"created_by", "updated_by", "version", "metadata",
	).
		From("nfs.service_request").
		Where(sq.And{
			sq.Eq{"ticket_number": ticketNumber},
			sq.Eq{"deleted_at": nil},
		})

	result, err := dblib.SelectOne(ctx, r.db, b, pgx.RowToStructByName[domain.ServiceRequest])
	if err != nil {
		log.Error(ctx, "GetByTicketNumber [%s]: %v", ticketNumber, err)
		return nil, err
	}
	return &result, nil
}

// ---------------------------------------------------------------------------
// ListByCustomerID returns paginated requests for a customer.
//
// BATCH: COUNT(*) + SELECT rows sent as a single pgx.Batch for one round-trip.
// ST-005 (FR-NFS-012 — customer can view own requests)
// ---------------------------------------------------------------------------
func (r *ServiceRequestRepository) ListByCustomerID(
	ctx context.Context,
	customerID int64,
	skip, limit int,
) ([]domain.ServiceRequest, int64, error) {
	timeout := r.cfg.GetDuration("db.QueryTimeoutMed")
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	countSQL, countArgs, err := dblib.Psql.Select("COUNT(*)").
		From("nfs.service_request").
		Where(sq.And{sq.Eq{"customer_id": customerID}, sq.Eq{"deleted_at": nil}}).
		ToSql()
	if err != nil {
		return nil, 0, fmt.Errorf("build count query: %w", err)
	}

	listSQL, listArgs, err := dblib.Psql.Select(
		"request_id", "ticket_number", "customer_id", "policy_number",
		"request_type", "auth_method", "status", "channel", "office_code",
		"initiated_by", "created_at", "updated_at", "sla_deadline", "sla_breached",
		"rejection_reason", "partial_processing_flag", "workflow_id", "workflow_run_id", "version",
	).
		From("nfs.service_request").
		Where(sq.And{sq.Eq{"customer_id": customerID}, sq.Eq{"deleted_at": nil}}).
		OrderBy("created_at DESC").
		Limit(uint64(limit)).
		Offset(uint64(skip)).
		ToSql()
	if err != nil {
		return nil, 0, fmt.Errorf("build list query: %w", err)
	}

	// BATCH: both queries in a single network round-trip
	batch := &pgx.Batch{}
	batch.Queue(countSQL, countArgs...)
	batch.Queue(listSQL, listArgs...)

	br := r.db.SendBatch(ctx, batch)
	defer br.Close()

	var total int64
	if err := br.QueryRow().Scan(&total); err != nil {
		log.Error(ctx, "ListByCustomerID count scan: %v", err)
		return nil, 0, fmt.Errorf("count query: %w", err)
	}

	rows, err := br.Query()
	if err != nil {
		log.Error(ctx, "ListByCustomerID list query: %v", err)
		return nil, 0, fmt.Errorf("list query: %w", err)
	}
	defer rows.Close()

	results, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.ServiceRequest])
	if err != nil {
		return nil, 0, fmt.Errorf("collect ListByCustomerID rows: %w", err)
	}
	return results, total, nil
}

// ---------------------------------------------------------------------------
// UpdateStatus transitions status + inserts status_transition_history +
// inserts audit_log in a single batched TX.
//
// BATCH: 3 operations in one TX batch → single round-trip.
// BR-NFS-012 (state machine), BR-NFS-016 (audit)
// ---------------------------------------------------------------------------
// UpdateStatus transitions a service request status and records audit+history.
// BATCH: UPDATE service_request + INSERT status_transition_history + INSERT audit_log (one TX).
// BR-NFS-012: status machine enforcement.
// BR-NFS-016: audit_log INSERT-only.
//
// updatedBy: *string (nil-safe: defaults to "SYSTEM").
// transition: optional; if nil, auto-generates a minimal transition record.
// Returns the updated ServiceRequest.
func (r *ServiceRequestRepository) UpdateStatus(
	ctx context.Context,
	requestID, newStatus string,
	updatedBy *string,
	rejectionReason *string,
	transition *domain.StatusTransitionHistory,
	audit *domain.AuditLog,
) (*domain.ServiceRequest, error) {
	timeout := r.cfg.GetDuration("db.QueryTimeoutMed")
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	actor := "SYSTEM"
	if updatedBy != nil && *updatedBy != "" {
		actor = *updatedBy
	}

	// Auto-generate transition record if not supplied.
	if transition == nil {
		transition = &domain.StatusTransitionHistory{
			TransitionID:     newUUID(),
			RequestID:        requestID,
			ToStatus:         newStatus,
			TransitionReason: rejectionReason,
			TransitionedBy:   actor,
		}
	}

	var updated domain.ServiceRequest

	dbErr := r.db.WithTx(ctx, func(tx pgx.Tx) error {
		batch := &pgx.Batch{}

		// 1. UPDATE service_request status (BR-NFS-012)
		updateBuilder := dblib.Psql.Update("nfs.service_request").
			Set("status", newStatus).
			Set("previous_status", sq.Expr("status")).
			Set("updated_by", actor).
			Where(sq.And{
				sq.Eq{"request_id": requestID},
				sq.Eq{"deleted_at": nil},
			}).
			Suffix("RETURNING request_id, ticket_number, customer_id, policy_number, request_type, status, updated_at, version")

		if rejectionReason != nil {
			updateBuilder = updateBuilder.Set("rejection_reason", *rejectionReason)
		}
		if newStatus == "COMPLETED" {
			updateBuilder = updateBuilder.Set("completed_at", sq.Expr("NOW()"))
		}

		upSQL, upArgs, err := updateBuilder.ToSql()
		if err != nil {
			return fmt.Errorf("build status update: %w", err)
		}
		batch.Queue(upSQL, upArgs...)

		// 2. INSERT status_transition_history (BR-NFS-012 history)
		thSQL, thArgs, err := dblib.Psql.Insert("nfs.status_transition_history").
			Columns("transition_id", "request_id", "from_status", "to_status", "transition_reason", "transitioned_by", "workflow_signal_id").
			Values(transition.TransitionID, transition.RequestID, transition.FromStatus, transition.ToStatus, transition.TransitionReason, transition.TransitionedBy, transition.WorkflowSignalID).
			ToSql()
		if err != nil {
			return fmt.Errorf("build status_transition_history insert: %w", err)
		}
		batch.Queue(thSQL, thArgs...)

		// 3. INSERT audit_log (BR-NFS-016)
		auditSQL, auditArgs, err := dblib.Psql.Insert("nfs.audit_log").
			Columns("audit_id", "request_id", "action_type", "old_value", "new_value", "performed_by", "channel", "office_code", "remarks").
			Values(audit.AuditID, audit.RequestID, audit.ActionType, audit.OldValue, audit.NewValueJSON, audit.PerformedByID, audit.Channel, audit.OfficeCode, audit.Notes).
			ToSql()
		if err != nil {
			return fmt.Errorf("build audit_log insert: %w", err)
		}
		batch.Queue(auditSQL, auditArgs...)

		br := tx.SendBatch(ctx, batch)

		// Result 1: service_request update
		if err := br.QueryRow().Scan(
			&updated.RequestID, &updated.TicketNumber, &updated.CustomerID, &updated.PolicyNumber,
			&updated.RequestType, &updated.Status, &updated.UpdatedAt, &updated.Version,
		); err != nil {
			br.Close()
			return fmt.Errorf("scan status update result: %w", err)
		}

		// Result 2: status_transition_history insert
		if _, err := br.Exec(); err != nil {
			br.Close()
			return fmt.Errorf("exec status_transition_history insert: %w", err)
		}

		// Result 3: audit_log insert
		if _, err := br.Exec(); err != nil {
			br.Close()
			return fmt.Errorf("exec audit_log insert: %w", err)
		}
		br.Close()
		return nil
	})

	if dbErr != nil {
		return nil, dbErr
	}
	return &updated, nil
}

// ---------------------------------------------------------------------------
// UpdateWorkflowState stores the Temporal workflow_id and workflow_run_id.
//
// WORKFLOW STATE NOTE: Called from CreateServiceRequest Temporal activity after
// workflow.GetInfo(ctx). Stored values are used for HTTP signal calls
// (verify OTP → WF-NFS-001, approve → WF-NFS-002).
// WF-NFS-001, WF-NFS-002, WF-NFS-003, WF-NFS-004
// ---------------------------------------------------------------------------
func (r *ServiceRequestRepository) UpdateWorkflowState(
	ctx context.Context,
	requestID, workflowID, workflowRunID string,
) error {
	timeout := r.cfg.GetDuration("db.QueryTimeoutLow")
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	b := dblib.Psql.Update("nfs.service_request").
		Set("workflow_id", workflowID).
		Set("workflow_run_id", workflowRunID).
		Where(sq.And{
			sq.Eq{"request_id": requestID},
			sq.Eq{"deleted_at": nil},
		})

	if _, err := dblib.Update(ctx, r.db, b); err != nil {
		log.Error(ctx, "UpdateWorkflowState [%s]: %v", requestID, err)
		return fmt.Errorf("update workflow state: %w", err)
	}
	return nil
}

// ---------------------------------------------------------------------------
// GetWorkflowState returns workflow_id + workflow_run_id for a request.
// Used by signal endpoints (OTP verify, approve, withdraw) to locate the
// running Temporal workflow. WF-NFS-001..WF-NFS-005
// ---------------------------------------------------------------------------
func (r *ServiceRequestRepository) GetWorkflowState(ctx context.Context, requestID string) (workflowID, workflowRunID string, err error) {
	timeout := r.cfg.GetDuration("db.QueryTimeoutLow")
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	query, args, err := dblib.Psql.Select("COALESCE(workflow_id,'')", "COALESCE(workflow_run_id,'')").
		From("nfs.service_request").
		Where(sq.And{sq.Eq{"request_id": requestID}, sq.Eq{"deleted_at": nil}}).
		ToSql()
	if err != nil {
		return "", "", fmt.Errorf("build GetWorkflowState query: %w", err)
	}

	row := r.db.QueryRow(ctx, query, args...)
	if err := row.Scan(&workflowID, &workflowRunID); err != nil {
		log.Error(ctx, "GetWorkflowState [%s]: %v", requestID, err)
		return "", "", fmt.Errorf("get workflow state: %w", err)
	}
	return workflowID, workflowRunID, nil
}

// ---------------------------------------------------------------------------
// AssignToCPC assigns the request to a CPC user + inserts queue entry +
// audit_log.
//
// BATCH: UPDATE service_request + INSERT cpc_work_queue + INSERT audit_log.
// FR-NFS-008, BR-NFS-016
// ---------------------------------------------------------------------------
func (r *ServiceRequestRepository) AssignToCPC(
	ctx context.Context,
	requestID string,
	assignedTo *string, assignedBy string,
	priority int,
	slaDeadline *time.Time,
	audit *domain.AuditLog,
) error {
	timeout := r.cfg.GetDuration("db.QueryTimeoutMed")
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return r.db.WithTx(ctx, func(tx pgx.Tx) error {
		batch := &pgx.Batch{}

		// 1. UPDATE service_request.assigned_to
		// srSQL, srArgs, err := dblib.Psql.Update("nfs.service_request").
		// 	Set("assigned_to", assignedTo).
		// 	Set("updated_by", assignedBy).
		// 	Where(sq.And{sq.Eq{"request_id": requestID}, sq.Eq{"deleted_at": nil}}).
		// 	ToSql()
		upd := dblib.Psql.Update("nfs.service_request").
			Set("status", "IN_PROGRESS"). // ← add this

			Set("updated_by", assignedBy).
			Where(sq.And{sq.Eq{"request_id": requestID}, sq.Eq{"deleted_at": nil}})
		if assignedTo != nil && *assignedTo != "" {
			upd = upd.Set("assigned_to", *assignedTo)
		}
		srSQL, srArgs, err := upd.ToSql()
		if err != nil {
			return fmt.Errorf("build assign update: %w", err)
		}
		log.Info(ctx, "AssignToCPC SQL: %s args: %v", srSQL, srArgs) // ← add this

		batch.Queue(srSQL, srArgs...)

		// 2. INSERT cpc_work_queue (FR-NFS-008)
		queueSQL, queueArgs, err := dblib.Psql.Insert("nfs.cpc_work_queue").
			Columns("request_id", "assigned_to", "priority", "sla_deadline").
			Values(requestID, nil, priority, slaDeadline).
			Suffix("ON CONFLICT (request_id) DO UPDATE SET assigned_to=$2, priority=$3, updated_at=NOW()").
			ToSql()
		if err != nil {
			return fmt.Errorf("build cpc_work_queue upsert: %w", err)
		}
		batch.Queue(queueSQL, queueArgs...)

		// 3. INSERT audit_log (BR-NFS-016)
		auditSQL, auditArgs, err := dblib.Psql.Insert("nfs.audit_log").
			Columns("audit_id", "request_id", "action_type", "new_value", "performed_by", "channel", "office_code").
			Values(audit.AuditID, audit.RequestID, "ASSIGNED", audit.NewValueJSON, audit.PerformedByID, audit.Channel, audit.OfficeCode).
			ToSql()
		if err != nil {
			return fmt.Errorf("build audit_log insert: %w", err)
		}
		batch.Queue(auditSQL, auditArgs...)

		br := tx.SendBatch(ctx, batch)

		if _, err := br.Exec(); err != nil {
			br.Close()
			return fmt.Errorf("exec assign update: %w", err)
		}
		if _, err := br.Exec(); err != nil {
			br.Close()
			return fmt.Errorf("exec cpc_work_queue upsert: %w", err)
		}
		if _, err := br.Exec(); err != nil {
			br.Close()
			return fmt.Errorf("exec audit_log insert: %w", err)
		}
		br.Close()
		return nil
	})
}

// ---------------------------------------------------------------------------
// GetStatusHistory returns all status transitions for a request.
// ST-002 (FR-NFS-012 — timeline view)
// ---------------------------------------------------------------------------
func (r *ServiceRequestRepository) GetStatusHistory(ctx context.Context, requestID string) ([]domain.StatusTransitionHistory, error) {
	timeout := r.cfg.GetDuration("db.QueryTimeoutLow")
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	b := dblib.Psql.Select(
		"transition_id", "request_id", "from_status", "to_status",
		"transition_reason", "transitioned_by", "transitioned_at", "workflow_signal_id",
	).
		From("nfs.status_transition_history").
		Where(sq.Eq{"request_id": requestID}).
		OrderBy("transitioned_at ASC")

	results, err := dblib.SelectRows(ctx, r.db, b, pgx.RowToStructByName[domain.StatusTransitionHistory])
	if err != nil {
		log.Error(ctx, "GetStatusHistory [%s]: %v", requestID, err)
		return nil, err
	}
	return results, nil
}

// ---------------------------------------------------------------------------
// GetDocumentRequestAttemptCount returns how many times missing-document links
// have been sent for a request.
// BR-NFS-014: max 3 secure upload link attempts per request.
// ---------------------------------------------------------------------------
func (r *ServiceRequestRepository) GetDocumentRequestAttemptCount(ctx context.Context, requestID string) (int, error) {
	timeout := r.cfg.GetDuration("db.QueryTimeoutLow")
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	sql, args, err := dblib.Psql.Select("COALESCE(link_generation_count, 0)").
		From("nfs.missing_document_request").
		Where(sq.And{
			sq.Eq{"request_id": requestID},
			sq.Eq{"deleted_at": nil},
		}).
		OrderBy("created_at DESC").
		Limit(1).
		ToSql()
	if err != nil {
		return 0, fmt.Errorf("build GetDocumentRequestAttemptCount query: %w", err)
	}

	var count int
	if err := r.db.QueryRow(ctx, sql, args...).Scan(&count); err != nil {
		if err == pgx.ErrNoRows {
			return 0, nil
		}
		log.Error(ctx, "GetDocumentRequestAttemptCount [%s]: %v", requestID, err)
		return 0, err
	}
	return count, nil
}

// ---------------------------------------------------------------------------
// CreateSecureUploadLink creates or regenerates a missing_document_request record
// and returns the secure token.
// BR-NFS-014: 7-day expiry; max 3 attempts.
// BATCH: INSERT missing_document_request + document_types rows in one TX.
// ---------------------------------------------------------------------------
func (r *ServiceRequestRepository) CreateSecureUploadLink(
	ctx context.Context,
	requestID string,
	docTypes []string,
	requestedBy string,
	expiresAt time.Time,
) (string, error) {
	timeout := r.cfg.GetDuration("db.QueryTimeoutMed")
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Generate a unique token (UUID-based).
	token := fmt.Sprintf("slt-%s-%d", requestID, time.Now().UnixNano())

	docsJSON := "["
	for i, d := range docTypes {
		if i > 0 {
			docsJSON += ","
		}
		docsJSON += fmt.Sprintf(`"%s"`, d)
	}
	docsJSON += "]"

	b := dblib.Psql.Insert("nfs.missing_document_request").
		Columns(
			"request_id", "document_type", "secure_link_token", "link_expiry",
			"link_generation_count", "status", "created_by",
		).
		Values(
			requestID, docsJSON, token, expiresAt,
			dblib.Psql.Select("COALESCE(MAX(link_generation_count),0)+1").
				From("nfs.missing_document_request").
				Where(sq.Eq{"request_id": requestID}).
				Prefix("(").Suffix(")"),
			"PENDING", requestedBy,
		).
		Suffix("RETURNING secure_link_token")

	sql, args, err := b.ToSql()
	if err != nil {
		return "", fmt.Errorf("build CreateSecureUploadLink query: %w", err)
	}

	var returnedToken string
	if err := r.db.QueryRow(ctx, sql, args...).Scan(&returnedToken); err != nil {
		log.Error(ctx, "CreateSecureUploadLink [%s]: %v", requestID, err)
		return "", err
	}
	return returnedToken, nil
}

// ---------------------------------------------------------------------------
// GetSecureUploadLink validates a secure upload token and returns the link record.
// Used by DM-005 (SecureUploadDocument) — no auth, token IS the authorization.
// BR-NFS-014: check expiry.
// ---------------------------------------------------------------------------
func (r *ServiceRequestRepository) GetSecureUploadLink(
	ctx context.Context,
	token string,
) (*domain.MissingDocumentRequest, error) {
	timeout := r.cfg.GetDuration("db.QueryTimeoutLow")
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	b := dblib.Psql.Select(
		"missing_doc_id", "request_id", "document_type", "secure_link_token",
		"link_expiry", "link_generation_count", "status",
		"received_at", "received_via", "received_document_id",
		"created_at", "updated_at",
	).
		From("nfs.missing_document_request").
		Where(sq.And{
			sq.Eq{"secure_link_token": token},
			sq.Eq{"status": "PENDING"},
			sq.Eq{"deleted_at": nil},
			sq.Expr("link_expiry > NOW()"),
		})

	result, err := dblib.SelectOne(ctx, r.db, b, pgx.RowToStructByName[domain.MissingDocumentRequest])
	if err != nil {
		log.Error(ctx, "GetSecureUploadLink [%s]: %v", token, err)
		return nil, err
	}
	return &result, nil
}

// ---------------------------------------------------------------------------
// MarkSecureUploadUsed increments the attempt counter and marks the link as RECEIVED.
// Called after successful document upload via secure link.
// ---------------------------------------------------------------------------
func (r *ServiceRequestRepository) MarkSecureUploadUsed(ctx context.Context, token string) error {
	timeout := r.cfg.GetDuration("db.QueryTimeoutLow")
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	b := dblib.Psql.Update("nfs.missing_document_request").
		Set("status", "RECEIVED").
		Set("received_at", sq.Expr("NOW()")).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"secure_link_token": token})

	if _, err := dblib.Update(ctx, r.db, b); err != nil {
		log.Error(ctx, "MarkSecureUploadUsed [%s]: %v", token, err)
		return err
	}
	return nil
}

// ---------------------------------------------------------------------------
// ListCPCQueue returns paginated queue items for CPC officers.
// BATCH: COUNT(*) + SELECT rows in one pgx.Batch (1 round-trip).
// CPC-001, FR-NFS-009.
// ---------------------------------------------------------------------------
func (r *ServiceRequestRepository) ListCPCQueue(
	ctx context.Context,
	status *string,
	requestType *string,
	officeCode *string,
	assignedToMe *string,
	page, pageSize int,
) (int, []domain.ServiceRequest, error) {
	timeout := r.cfg.GetDuration("db.QueryTimeoutMed")
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	conditions := sq.And{sq.Eq{"deleted_at": nil}}
	if status != nil {
		conditions = append(conditions, sq.Eq{"status": *status})
	} else {
		// Default: show PENDING_APPROVAL and IN_PROGRESS items.
		conditions = append(conditions, sq.Eq{"status": []string{"PENDING_APPROVAL", "IN_PROGRESS"}})
	}
	if requestType != nil {
		conditions = append(conditions, sq.Eq{"request_type": *requestType})
	}
	if officeCode != nil {
		conditions = append(conditions, sq.Eq{"office_code": *officeCode})
	}
	if assignedToMe != nil {
		conditions = append(conditions, sq.Eq{"assigned_to": *assignedToMe})
	}

	countSQL, countArgs, err := dblib.Psql.Select("COUNT(*)").
		From("nfs.service_request").
		Where(conditions).
		ToSql()
	if err != nil {
		return 0, nil, fmt.Errorf("build ListCPCQueue count query: %w", err)
	}

	offset := (page - 1) * pageSize
	rowSQL, rowArgs, err := dblib.Psql.Select(
		"request_id", "ticket_number", "customer_id", "policy_number",
		"request_type", "status", "auth_method", "channel",
		"office_code", "assigned_to", "sla_deadline",
		"created_at", "updated_at",
	).
		From("nfs.service_request").
		Where(conditions).
		OrderBy("sla_deadline ASC NULLS LAST", "created_at ASC").
		Limit(uint64(pageSize)).
		Offset(uint64(offset)).
		ToSql()
	if err != nil {
		return 0, nil, fmt.Errorf("build ListCPCQueue row query: %w", err)
	}

	batch := &pgx.Batch{}
	batch.Queue(countSQL, countArgs...)
	batch.Queue(rowSQL, rowArgs...)

	br := r.db.SendBatch(ctx, batch)
	defer br.Close()

	var total int
	if err := br.QueryRow().Scan(&total); err != nil {
		log.Error(ctx, "ListCPCQueue count: %v", err)
		return 0, nil, err
	}

	rows, err := br.Query()
	if err != nil {
		log.Error(ctx, "ListCPCQueue rows: %v", err)
		return 0, nil, err
	}
	defer rows.Close()

	results, err := pgx.CollectRows(rows, pgx.RowToStructByName[domain.ServiceRequest])
	if err != nil {
		return 0, nil, fmt.Errorf("collect ListCPCQueue rows: %w", err)
	}
	return total, results, nil
}

// ---------------------------------------------------------------------------
// GetCPCQueueStats returns aggregated queue statistics for a CPC office.
// BATCH: 5 COUNT queries in one pgx.Batch round-trip.
// CPC-002, FR-NFS-010.
// ---------------------------------------------------------------------------
func (r *ServiceRequestRepository) GetCPCQueueStats(
	ctx context.Context,
	officeCode *string,
) (map[string]int, error) {
	timeout := r.cfg.GetDuration("db.QueryTimeoutMed")
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	conditions := sq.And{sq.Eq{"deleted_at": nil}}
	if officeCode != nil {
		conditions = append(conditions, sq.Eq{"office_code": *officeCode})
	}

	pendingCond := append(conditions, sq.Eq{"status": "PENDING_APPROVAL"})
	inProgressCond := append(conditions, sq.Eq{"status": "IN_PROGRESS"})
	overdueCond := append(conditions, sq.Expr("sla_deadline < NOW()"), sq.Eq{"status": []string{"PENDING_APPROVAL", "IN_PROGRESS"}})
	addrCond := append(conditions, sq.Eq{"request_type": "ADDRESS_CHANGE"}, sq.Eq{"status": []string{"PENDING_APPROVAL", "IN_PROGRESS"}})
	nameCond := append(conditions, sq.Eq{"request_type": "NAME_CHANGE"}, sq.Eq{"status": []string{"PENDING_APPROVAL", "IN_PROGRESS"}})

	batch := &pgx.Batch{}
	for _, cond := range []sq.Sqlizer{pendingCond, inProgressCond, overdueCond, addrCond, nameCond} {
		q, a, _ := dblib.Psql.Select("COUNT(*)").From("nfs.service_request").Where(cond).ToSql()
		batch.Queue(q, a...)
	}

	br := r.db.SendBatch(ctx, batch)
	defer br.Close()

	keys := []string{"total_pending", "total_in_progress", "total_overdue", "address_changes", "name_changes"}
	result := make(map[string]int, len(keys))
	for _, k := range keys {
		var n int
		if err := br.QueryRow().Scan(&n); err != nil {
			log.Error(ctx, "GetCPCQueueStats %s: %v", k, err)
			return nil, err
		}
		result[k] = n
	}
	return result, nil
}

// ---------------------------------------------------------------------------
// ListSLABreached returns requests that have breached their SLA deadline.
// CPC-006, FR-NFS-010.
// ---------------------------------------------------------------------------
func (r *ServiceRequestRepository) ListSLABreached(
	ctx context.Context,
	officeCode *string,
	slaStatus *string,
) ([]domain.ServiceRequest, error) {
	timeout := r.cfg.GetDuration("db.QueryTimeoutMed")
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	conditions := sq.And{
		sq.Eq{"deleted_at": nil},
		sq.Expr("sla_deadline < NOW()"),
		sq.Eq{"status": []string{"PENDING_APPROVAL", "IN_PROGRESS"}},
	}
	if officeCode != nil {
		conditions = append(conditions, sq.Eq{"office_code": *officeCode})
	}

	b := dblib.Psql.Select(
		"request_id", "ticket_number", "customer_id", "policy_number",
		"request_type", "status", "office_code", "sla_deadline",
		"created_at", "updated_at",
	).
		From("nfs.service_request").
		Where(conditions).
		OrderBy("sla_deadline ASC")

	results, err := dblib.SelectRows(ctx, r.db, b, pgx.RowToStructByName[domain.ServiceRequest])
	if err != nil {
		log.Error(ctx, "ListSLABreached: %v", err)
		return nil, err
	}
	return results, nil
}

// ---------------------------------------------------------------------------
// UpdateSLADeadline sets the sla_deadline on a service request after creation.
// Called by InitiateNameChange (MANUAL path) after the workflow has started,
// since SLA is calculated using slaDays config which is only known at runtime.
// BR-NFS-002: SLA deadline = created_at + regional SLA days.
// ---------------------------------------------------------------------------
func (r *ServiceRequestRepository) UpdateSLADeadline(
	ctx context.Context,
	requestID string,
	slaDeadline time.Time,
) error {
	timeout := r.cfg.GetDuration("db.QueryTimeoutLow")
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	b := dblib.Psql.Update("nfs.service_request").
		Set("sla_deadline", slaDeadline).
		Where(sq.And{
			sq.Eq{"request_id": requestID},
			sq.Eq{"deleted_at": nil},
		})

	if _, err := dblib.Update(ctx, r.db, b); err != nil {
		log.Error(ctx, "UpdateSLADeadline [%s]: %v", requestID, err)
		return fmt.Errorf("update SLA deadline: %w", err)
	}
	return nil
}
