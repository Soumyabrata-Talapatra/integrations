// Package postgres — document_upload repository.
//
// FR-NFS-002: Document upload (PDF/JPEG/PNG, max 5 MB per VR-NFS-009/010)
// FR-NFS-005: Document retrieval
// FR-NFS-009: Secure link upload for missing documents
// BR-NFS-008: Virus scan on upload before verification
// BR-NFS-014: Max 3 secure link regenerations, max 2 reminders
// BR-NFS-016: Audit every upload and verification event
// VR-NFS-009: Max file size 5 MB
// VR-NFS-010: Allowed MIME types: PDF, JPEG, PNG
//
// BATCH NOTE:
//
//	CreateBatch → multiple document INSERTs in a single pgx.Batch TX.
//	ListByRequestID → COUNT + SELECT in one pgx.Batch (one round-trip).
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

// DocumentUploadRepository handles DB operations for nfs.document_upload (E-4)
// and nfs.missing_document_request (E-5).
type DocumentUploadRepository struct {
	db  *dblib.DB
	cfg *config.Config
}

// NewDocumentUploadRepository constructs the repository.
func NewDocumentUploadRepository(db *dblib.DB, cfg *config.Config) *DocumentUploadRepository {
	return &DocumentUploadRepository{db: db, cfg: cfg}
}

// ---------------------------------------------------------------------------
// CreateBatch inserts multiple documents for a request in a single TX batch.
//
// BATCH: All document INSERTs queued in one pgx.Batch, one round-trip.
// FR-NFS-002, VR-NFS-009/010, BR-NFS-016
// ---------------------------------------------------------------------------
func (r *DocumentUploadRepository) CreateBatch(
	ctx context.Context,
	docs []domain.DocumentUpload,
	audit *domain.AuditLog,
) error {
	if len(docs) == 0 {
		return nil
	}

	timeout := r.cfg.GetDuration("db.QueryTimeoutMed")
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return r.db.WithTx(ctx, func(tx pgx.Tx) error {
		batch := &pgx.Batch{}

		// Queue one INSERT per document (FR-NFS-002)
		for _, doc := range docs {
			docSQL, docArgs, err := dblib.Psql.Insert("nfs.document_upload").
				Columns(
					"document_id", "request_id", "document_type",
					"file_name", "file_url", "file_size_bytes", "mime_type",
					"uploaded_by", "upload_channel", "created_by",
				).
				Values(
					doc.DocumentID, doc.RequestID, doc.DocumentType,
					doc.FileName, doc.FileURL, doc.FileSizeBytes, doc.MimeType,
					doc.UploadedBy, doc.UploadChannel, doc.CreatedBy,
				).
				Suffix("RETURNING document_id, uploaded_at").
				ToSql()
			if err != nil {
				return fmt.Errorf("build document_upload insert: %w", err)
			}
			batch.Queue(docSQL, docArgs...)
		}

		// Final: audit_log INSERT (BR-NFS-016)
		auditSQL, auditArgs, err := dblib.Psql.Insert("nfs.audit_log").
			Columns("audit_id", "request_id", "action_type", "new_value", "performed_by", "channel", "office_code").
			Values(audit.AuditID, audit.RequestID, "DOCUMENT_UPLOAD", audit.NewValueJSON, audit.PerformedByID, audit.Channel, audit.OfficeCode).
			ToSql()
		if err != nil {
			return fmt.Errorf("build audit_log insert: %w", err)
		}
		batch.Queue(auditSQL, auditArgs...)

		br := tx.SendBatch(ctx, batch)
		defer br.Close()

		// Scan RETURNING for each document
		for i := range docs {
			if err := br.QueryRow().Scan(&docs[i].DocumentID, &docs[i].UploadedAt); err != nil {
				return fmt.Errorf("scan document[%d] result: %w", i, err)
			}
		}

		// Audit log exec
		if _, err := br.Exec(); err != nil {
			return fmt.Errorf("exec audit_log insert: %w", err)
		}

		return nil
	})
}

// ---------------------------------------------------------------------------
// GetByID fetches a single document upload by document_id.
// FR-NFS-005
// ---------------------------------------------------------------------------
func (r *DocumentUploadRepository) GetByID(ctx context.Context, documentID string) (*domain.DocumentUpload, error) {
	timeout := r.cfg.GetDuration("db.QueryTimeoutLow")
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	b := dblib.Psql.Select(
		"document_id", "request_id", "document_type",
		"file_name", "file_url", "file_size_bytes", "mime_type",
		"uploaded_by", "upload_channel", "uploaded_at",
		"verification_status", "verified_by", "verified_at", "rejection_reason",
		"dms_document_id", "dms_storage_path",
		"virus_scan_status", "virus_scanned_at",
		"created_at", "updated_at", "version",
	).
		From("nfs.document_upload").
		Where(sq.And{
			sq.Eq{"document_id": documentID},
			sq.Eq{"deleted_at": nil},
		})

	result, err := dblib.SelectOne(ctx, r.db, b, pgx.RowToStructByName[domain.DocumentUpload])
	if err != nil {
		log.Error(ctx, "DocumentUpload.GetByID [%s]: %v", documentID, err)
		return nil, err
	}
	return &result, nil
}

// ---------------------------------------------------------------------------
// ListByRequestID returns paginated documents for a request.
//
// BATCH: COUNT(*) + SELECT rows in one pgx.Batch (1 round-trip).
// FR-NFS-005, DM-003
// ---------------------------------------------------------------------------
func (r *DocumentUploadRepository) ListByRequestID(
	ctx context.Context,
	requestID string,
	skip, limit int,
) ([]domain.DocumentUpload, int64, error) {
	timeout := r.cfg.GetDuration("db.QueryTimeoutMed")
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	baseWhere := sq.And{sq.Eq{"request_id": requestID}, sq.Eq{"deleted_at": nil}}

	countBuilder := dblib.Psql.Select("COUNT(*)").
		From("nfs.document_upload").
		Where(baseWhere)

	listBuilder := dblib.Psql.Select(
		"document_id", "request_id", "document_type",
		"file_name", "file_size_bytes", "mime_type",
		"uploaded_by", "upload_channel", "uploaded_at",
		"verification_status", "virus_scan_status", "version",
	).
		From("nfs.document_upload").
		Where(baseWhere).
		OrderBy("uploaded_at DESC").
		Limit(uint64(limit)).
		Offset(uint64(skip))

	var results []domain.DocumentUpload

	// BATCH: COUNT(*) + list in one round-trip
	countSQL, countArgs, err := countBuilder.ToSql()
	if err != nil {
		return nil, 0, fmt.Errorf("build count query: %w", err)
	}
	listSQL, listArgs, err := listBuilder.ToSql()
	if err != nil {
		return nil, 0, fmt.Errorf("build list query: %w", err)
	}

	batch := &pgx.Batch{}
	batch.Queue(countSQL, countArgs...)
	batch.Queue(listSQL, listArgs...)

	br := r.db.SendBatch(ctx, batch)
	defer br.Close()

	var total int64
	if err := br.QueryRow().Scan(&total); err != nil {
		log.Error(ctx, "document ListByRequestID count: %v", err)
		return nil, 0, fmt.Errorf("count query: %w", err)
	}

	rows, err := br.Query()
	if err != nil {
		log.Error(ctx, "document ListByRequestID list: %v", err)
		return nil, 0, fmt.Errorf("list query: %w", err)
	}
	defer rows.Close()

	results, err = pgx.CollectRows(rows, pgx.RowToStructByName[domain.DocumentUpload])
	if err != nil {
		return nil, 0, fmt.Errorf("collect document rows: %w", err)
	}
	return results, total, nil
}

// ---------------------------------------------------------------------------
// UpdateVerificationStatus updates the CPC document verification result +
// inserts audit_log in a batched TX.
//
// BATCH: UPDATE document + INSERT audit_log in one TX batch.
// FR-NFS-011, BR-NFS-016
// ---------------------------------------------------------------------------
func (r *DocumentUploadRepository) UpdateVerificationStatus(
	ctx context.Context,
	documentID, status, verifiedBy string,
	rejectionReason *string,
	audit *domain.AuditLog,
) error {
	timeout := r.cfg.GetDuration("db.QueryTimeoutMed")
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return r.db.WithTx(ctx, func(tx pgx.Tx) error {
		updateBuilder := dblib.Psql.Update("nfs.document_upload").
			Set("verification_status", status).
			Set("verified_by", verifiedBy).
			Set("verified_at", sq.Expr("NOW()")).
			Set("updated_by", verifiedBy).
			Where(sq.And{sq.Eq{"document_id": documentID}, sq.Eq{"deleted_at": nil}})

		if rejectionReason != nil {
			updateBuilder = updateBuilder.Set("rejection_reason", *rejectionReason)
		}

		auditBuilder := dblib.Psql.Insert("nfs.audit_log").
			Columns("audit_id", "request_id", "action_type", "old_value", "new_value", "performed_by", "channel", "office_code").
			Values(audit.AuditID, audit.RequestID, "DOCUMENT_VERIFIED", audit.OldValue, audit.NewValueJSON, audit.PerformedByID, audit.Channel, audit.OfficeCode)

		batch := &pgx.Batch{}
		if err := dblib.QueueExecRow(batch, updateBuilder); err != nil {
			return fmt.Errorf("queue verification update: %w", err)
		}
		if err := dblib.QueueExecRow(batch, auditBuilder); err != nil {
			return fmt.Errorf("queue audit insert: %w", err)
		}

		br := tx.SendBatch(ctx, batch)
		if err := br.Close(); err != nil {
			return fmt.Errorf("UpdateVerificationStatus batch: %w", err)
		}
		return nil
	})
}

// ---------------------------------------------------------------------------
// UpdateVirusScanStatus updates the virus_scan_status column.
// Called by the virus-scan service callback. BR-NFS-008.
// ---------------------------------------------------------------------------
func (r *DocumentUploadRepository) UpdateVirusScanStatus(ctx context.Context, documentID, scanStatus string) error {
	timeout := r.cfg.GetDuration("db.QueryTimeoutLow")
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	b := dblib.Psql.Update("nfs.document_upload").
		Set("virus_scan_status", scanStatus).
		Set("virus_scanned_at", sq.Expr("NOW()")).
		Where(sq.And{sq.Eq{"document_id": documentID}, sq.Eq{"deleted_at": nil}})

	if _, err := dblib.Update(ctx, r.db, b); err != nil {
		log.Error(ctx, "UpdateVirusScanStatus [%s]: %v", documentID, err)
		return fmt.Errorf("update virus scan status: %w", err)
	}
	return nil
}

// ---------------------------------------------------------------------------
// UpdateDMSReference stores the DMS document ID and storage path after
// successful DMS upload. FR-NFS-002.
// ---------------------------------------------------------------------------
func (r *DocumentUploadRepository) UpdateDMSReference(ctx context.Context, documentID, dmsDocumentID, dmsStoragePath string) error {
	timeout := r.cfg.GetDuration("db.QueryTimeoutLow")
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	b := dblib.Psql.Update("nfs.document_upload").
		Set("dms_document_id", dmsDocumentID).
		Set("dms_storage_path", dmsStoragePath).
		Where(sq.And{sq.Eq{"document_id": documentID}, sq.Eq{"deleted_at": nil}})

	if _, err := dblib.Update(ctx, r.db, b); err != nil {
		log.Error(ctx, "UpdateDMSReference [%s]: %v", documentID, err)
		return fmt.Errorf("update DMS reference: %w", err)
	}
	return nil
}

// ===========================================================================
// Missing Document Request operations (E-5)
// ===========================================================================

// ---------------------------------------------------------------------------
// CreateMissingDocRequest creates a request for a customer to upload a missing
// document via a secure link. BR-NFS-014, FR-NFS-009.
// ---------------------------------------------------------------------------
func (r *DocumentUploadRepository) CreateMissingDocRequest(
	ctx context.Context,
	req *domain.MissingDocumentRequest,
	audit *domain.AuditLog,
) error {
	timeout := r.cfg.GetDuration("db.QueryTimeoutMed")
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return r.db.WithTx(ctx, func(tx pgx.Tx) error {
		batch := &pgx.Batch{}

		// INSERT missing_document_request (BR-NFS-014)
		insertSQL, insertArgs, err := dblib.Psql.Insert("nfs.missing_document_request").
			Columns(
				"missing_doc_id", "request_id", "document_type",
				"secure_link_url", "secure_link_token", "link_expiry",
				"link_generation_count", "status", "requested_by", "message_to_customer", "created_by",
			).
			Values(
				req.MissingDocID, req.RequestID, req.DocumentType,
				req.SecureLinkURL, req.SecureLinkToken, req.LinkExpiry,
				req.LinkGenerationCount, req.Status, req.RequestedBy, req.MessageToCustomer, req.CreatedBy,
			).
			Suffix("RETURNING missing_doc_id, created_at").
			ToSql()
		if err != nil {
			return fmt.Errorf("build missing_doc_request insert: %w", err)
		}
		batch.Queue(insertSQL, insertArgs...)

		// audit_log (BR-NFS-016)
		auditSQL, auditArgs, err := dblib.Psql.Insert("nfs.audit_log").
			Columns("audit_id", "request_id", "action_type", "new_value", "performed_by", "channel", "office_code").
			Values(audit.AuditID, audit.RequestID, "MISSING_DOC_REQUESTED", audit.NewValueJSON, audit.PerformedByID, audit.Channel, audit.OfficeCode).
			ToSql()
		if err != nil {
			return fmt.Errorf("build audit_log insert: %w", err)
		}
		batch.Queue(auditSQL, auditArgs...)

		br := tx.SendBatch(ctx, batch)
		defer br.Close()

		if err := br.QueryRow().Scan(&req.MissingDocID, &req.CreatedAt); err != nil {
			return fmt.Errorf("scan missing_doc_request result: %w", err)
		}
		if _, err := br.Exec(); err != nil {
			return fmt.Errorf("exec audit_log insert: %w", err)
		}
		return nil
	})
}

// ---------------------------------------------------------------------------
// RegenerateSecureLink increments link_generation_count and updates the link
// token/expiry. Max 3 generations enforced at service layer (BR-NFS-014).
// Max link_generation_count = nfs.cfg.GetInt("nfs.maxmissingdocattempts") = 3.
// ---------------------------------------------------------------------------
func (r *DocumentUploadRepository) RegenerateSecureLink(
	ctx context.Context,
	missingDocID, secureLinkURL, secureLinkToken string,
	linkExpiry interface{},
) error {
	timeout := r.cfg.GetDuration("db.QueryTimeoutLow")
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	b := dblib.Psql.Update("nfs.missing_document_request").
		Set("secure_link_url", secureLinkURL).
		Set("secure_link_token", secureLinkToken).
		Set("link_expiry", linkExpiry).
		Set("link_generation_count", sq.Expr("link_generation_count + 1")).
		Where(sq.And{
			sq.Eq{"missing_doc_id": missingDocID},
			sq.Eq{"deleted_at": nil},
			sq.LtOrEq{"link_generation_count": 2}, // max 3 = current+1
		})

	if _, err := dblib.Update(ctx, r.db, b); err != nil {
		log.Error(ctx, "RegenerateSecureLink [%s]: %v", missingDocID, err)
		return fmt.Errorf("regenerate secure link: %w", err)
	}
	return nil
}

// ---------------------------------------------------------------------------
// GetMissingDocByToken looks up a missing doc request by secure link token.
// Used by the secure-link upload endpoint (FR-NFS-009).
// ---------------------------------------------------------------------------
func (r *DocumentUploadRepository) GetMissingDocByToken(ctx context.Context, token string) (*domain.MissingDocumentRequest, error) {
	timeout := r.cfg.GetDuration("db.QueryTimeoutLow")
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	b := dblib.Psql.Select(
		"missing_doc_id", "request_id", "document_type",
		"secure_link_url", "secure_link_token", "link_expiry",
		"link_generation_count", "status", "reminder_count",
		"requested_by", "message_to_customer", "created_at", "version",
	).
		From("nfs.missing_document_request").
		Where(sq.And{
			sq.Eq{"secure_link_token": token},
			sq.Eq{"status": "PENDING"},
			sq.Eq{"deleted_at": nil},
		})

	result, err := dblib.SelectOne(ctx, r.db, b, pgx.RowToStructByName[domain.MissingDocumentRequest])
	if err != nil {
		log.Error(ctx, "GetMissingDocByToken: %v", err)
		return nil, err
	}
	return &result, nil
}

// ---------------------------------------------------------------------------
// MarkMissingDocReceived marks the missing doc as RECEIVED when the customer
// uploads via secure link. BR-NFS-014.
// ---------------------------------------------------------------------------
func (r *DocumentUploadRepository) MarkMissingDocReceived(
	ctx context.Context,
	missingDocID, receivedDocumentID string,
	receivedVia string,
) error {
	timeout := r.cfg.GetDuration("db.QueryTimeoutLow")
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	b := dblib.Psql.Update("nfs.missing_document_request").
		Set("status", "RECEIVED").
		Set("received_at", sq.Expr("NOW()")).
		Set("received_via", receivedVia).
		Set("received_document_id", receivedDocumentID).
		Where(sq.And{
			sq.Eq{"missing_doc_id": missingDocID},
			sq.Eq{"deleted_at": nil},
		})

	if _, err := dblib.Update(ctx, r.db, b); err != nil {
		log.Error(ctx, "MarkMissingDocReceived [%s]: %v", missingDocID, err)
		return fmt.Errorf("mark missing doc received: %w", err)
	}
	return nil
}

// ---------------------------------------------------------------------------
// Create inserts a single document upload record.
// Used by DM-001 (UploadDocument) and DM-005 (SecureUploadDocument).
// VR-NFS-009, VR-NFS-010.
// ---------------------------------------------------------------------------
func (r *DocumentUploadRepository) Create(ctx context.Context, doc *domain.DocumentUpload) (*domain.DocumentUpload, error) {
	timeout := r.cfg.GetDuration("db.QueryTimeoutLow")
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	b := dblib.Psql.Insert("nfs.document_upload").
		Columns(
			"request_id", "document_type", "file_name", "file_url",
			"file_size_bytes", "mime_type", "uploaded_by", "upload_channel",
			"verification_status", "created_by",
		).
		Values(
			doc.RequestID, doc.DocumentType, doc.FileName, doc.FileURL,
			doc.FileSizeBytes, doc.MimeType, doc.UploadedBy, doc.UploadChannel,
			doc.VerificationStatus, doc.CreatedBy,
		).
		Suffix("RETURNING *")

	result, err := dblib.InsertReturning(ctx, r.db, b, pgx.RowToStructByName[domain.DocumentUpload])
	if err != nil {
		log.Error(ctx, "DocumentUpload.Create [%s]: %v", doc.RequestID, err)
		return nil, err
	}
	return &result, nil
}

// ---------------------------------------------------------------------------
// Delete performs a soft-delete on a document upload record.
// DM-004: only callable when parent request is in CREATED status.
// ---------------------------------------------------------------------------
func (r *DocumentUploadRepository) Delete(ctx context.Context, documentID string) error {
	timeout := r.cfg.GetDuration("db.QueryTimeoutLow")
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	b := dblib.Psql.Update("nfs.document_upload").
		Set("deleted_at", sq.Expr("NOW()")).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.And{
			sq.Eq{"document_id": documentID},
			sq.Eq{"deleted_at": nil},
		})

	if _, err := dblib.Update(ctx, r.db, b); err != nil {
		log.Error(ctx, "DocumentUpload.Delete [%s]: %v", documentID, err)
		return err
	}
	return nil
}
