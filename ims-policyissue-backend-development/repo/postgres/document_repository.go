package postgres

import (
	"context"
	"fmt"
	"time"

	"policy-issue-service/core/domain"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	config "gitlab.cept.gov.in/it-2.0-common/api-config"
	dblib "gitlab.cept.gov.in/it-2.0-common/n-api-db"
)

// DocumentRepository handles document-related database operations
// Phase 8: [DOC-POL-001] to [DOC-POL-007], [STATUS-POL-001] to [STATUS-POL-003]
type DocumentRepository struct {
	db  *dblib.DB
	cfg *config.Config
}

// NewDocumentRepository creates a new DocumentRepository instance
func NewDocumentRepository(db *dblib.DB, cfg *config.Config) *DocumentRepository {
	return &DocumentRepository{db: db, cfg: cfg}
}

// ============================================
// Document Reference Operations (E-016)
// ============================================

// GetDocumentsByProposalID retrieves all non-deleted documents for a proposal
// [DOC-POL-001] Used to build document checklist
func (r *DocumentRepository) GetDocumentsByProposalID(ctx context.Context, proposalID int64) ([]domain.ProposalDocumentRef, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
	defer cancel()

	query := dblib.Psql.Select(
		"doc_ref_id", "proposal_id", "document_id", "document_type",
		"file_name", "file_size_bytes", "mime_type", "uploaded_by",
		"uploaded_at", "version", "comments", "document_date",
	).From("proposal_document_ref").
		Where(sq.Eq{"proposal_id": proposalID}).
		Where("deleted_at IS NULL").
		OrderBy("uploaded_at DESC")

	return dblib.SelectRows(ctx, r.db, query, pgx.RowToStructByNameLax[domain.ProposalDocumentRef])
}

// GetDocumentByID retrieves a single document reference by ID
// [DOC-POL-003] Document download
func (r *DocumentRepository) GetDocumentByID(ctx context.Context, proposalID int64, docRefID int64) (*domain.ProposalDocumentRef, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
	defer cancel()

	query := dblib.Psql.Select(
		"doc_ref_id", "proposal_id", "document_id", "document_type",
		"file_name", "file_size_bytes", "mime_type", "uploaded_by",
		"uploaded_at", "version", "comments",
	).From("policy_issue.proposal_document_ref").
		Where(sq.Eq{"proposal_id": proposalID, "doc_ref_id": docRefID}).
		Where("deleted_at IS NULL")

	doc, err := dblib.SelectOne(ctx, r.db, query, pgx.RowToStructByNameLax[domain.ProposalDocumentRef])
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

// CreateDocumentRef inserts a new document reference
// [DOC-POL-002] Document upload
func (r *DocumentRepository) CreateDocumentRef(ctx context.Context, doc *domain.ProposalDocumentRef) error {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
	defer cancel()

	now := time.Now()
	insertSQL := `
		INSERT INTO proposal_document_ref (
			proposal_id, document_id, document_type, file_name,
			file_size_bytes, mime_type, document_date, uploaded_by,
			uploaded_at, version, comments
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING doc_ref_id, uploaded_at
	`

	type insertResult struct {
		DocRefID   int64     `db:"doc_ref_id"`
		UploadedAt time.Time `db:"uploaded_at"`
	}
	result, err := dblib.ExecReturn(ctx, r.db, insertSQL, []any{
		doc.ProposalID, doc.DocumentID, doc.DocumentType, doc.FileName,
		doc.FileSizeBytes, doc.MimeType, doc.DocumentDate, doc.UploadedBy,
		now, 1, doc.Comments,
	}, pgx.RowToStructByName[insertResult])
	if err != nil {
		return fmt.Errorf("failed to insert document ref: %w", err)
	}
	doc.DocRefID = result.DocRefID
	doc.UploadedAt = result.UploadedAt
	return nil
}

// SoftDeleteDocument soft-deletes a document reference
// [DOC-POL-004] Document removal
func (r *DocumentRepository) SoftDeleteDocument(ctx context.Context, proposalID int64, docRefID int64) error {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	updateSQL := `
		UPDATE proposal_document_ref
		SET deleted_at = $1
		WHERE proposal_id = $2 AND doc_ref_id = $3 AND deleted_at IS NULL
	`
	_, err := dblib.Exec(ctx, r.db, updateSQL, []any{time.Now(), proposalID, docRefID})
	return err
}

// ============================================
// Missing Document Operations (E-017)
// ============================================

// GetMissingDocuments retrieves missing documents for a proposal with optional filters
// [DOC-POL-005] Get missing documents
func (r *DocumentRepository) GetMissingDocuments(ctx context.Context, proposalID int64, stage string, status string) ([]domain.ProposalMissingDocument, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
	defer cancel()

	query := dblib.Psql.Select(
		"missing_doc_id", "proposal_id", "document_type", "document_description",
		"stage", "noted_by", "noted_at", "notes", "status",
		"resolved_by", "resolved_at", "resolution_notes",
		"uploaded_document_id", "waived", "waived_by", "waived_at",
		"waiver_reason", "created_at", "updated_at",
	).From("proposal_missing_documents").
		Where(sq.Eq{"proposal_id": proposalID}).
		OrderBy("noted_at DESC")

	if stage != "" {
		query = query.Where(sq.Eq{"stage": stage})
	}
	if status != "" {
		query = query.Where(sq.Eq{"status": status})
	}

	return dblib.SelectRows(ctx, r.db, query, pgx.RowToStructByName[domain.ProposalMissingDocument])
}

// GetMissingDocumentByID retrieves a single missing document record
// [DOC-POL-007] Resolve missing document
func (r *DocumentRepository) GetMissingDocumentByID(ctx context.Context, missingDocID int64) (*domain.ProposalMissingDocument, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutShort"))
	defer cancel()

	query := dblib.Psql.Select(
		"missing_doc_id", "proposal_id", "document_type", "document_description",
		"stage", "noted_by", "noted_at", "notes", "status",
		"resolved_by", "resolved_at", "resolution_notes",
		"uploaded_document_id", "waived", "waived_by", "waived_at",
		"waiver_reason", "created_at", "updated_at",
	).From("proposal_missing_documents").
		Where(sq.Eq{"missing_doc_id": missingDocID})

	doc, err := dblib.SelectOne(ctx, r.db, query, pgx.RowToStructByName[domain.ProposalMissingDocument])
	if err != nil {
		return nil, err
	}
	return &doc, nil
}

// CreateMissingDocument inserts a new missing document record
// [DOC-POL-006] Record missing document
func (r *DocumentRepository) CreateMissingDocument(ctx context.Context, doc *domain.ProposalMissingDocument) error {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
	defer cancel()

	now := time.Now()
	insertSQL := `
		INSERT INTO proposal_missing_documents (
			proposal_id, document_type, document_description,
			stage, noted_by, noted_at, notes, status,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING missing_doc_id, noted_at, created_at, updated_at
	`

	type insertResult struct {
		MissingDocID int64     `db:"missing_doc_id"`
		NotedAt      time.Time `db:"noted_at"`
		CreatedAt    time.Time `db:"created_at"`
		UpdatedAt    time.Time `db:"updated_at"`
	}
	result, err := dblib.ExecReturn(ctx, r.db, insertSQL, []any{
		doc.ProposalID, doc.DocumentType, doc.DocumentDescription,
		doc.Stage, doc.NotedBy, now, doc.Notes, string(domain.MissingDocStatusPending),
		now, now,
	}, pgx.RowToStructByName[insertResult])
	if err != nil {
		return fmt.Errorf("failed to insert missing document: %w", err)
	}
	doc.MissingDocID = result.MissingDocID
	doc.NotedAt = result.NotedAt
	doc.Status = domain.MissingDocStatusPending
	doc.CreatedAt = result.CreatedAt
	doc.UpdatedAt = result.UpdatedAt
	return nil
}

// ResolveMissingDocument updates a missing document as resolved (UPLOADED or WAIVED)
// [DOC-POL-007] Resolve missing document
func (r *DocumentRepository) ResolveMissingDocument(ctx context.Context, missingDocID int64, status domain.MissingDocumentStatus, resolvedBy int64, uploadedDocID *int64, waiverReason *string, resolutionNotes *string) error {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
	defer cancel()

	now := time.Now()

	// Build dynamic update SQL based on resolution type
	if status == domain.MissingDocStatusWaived {
		updateSQL := `
			UPDATE proposal_missing_documents
			SET status = $1, resolved_by = $2, resolved_at = $3, resolution_notes = $4,
				waived = true, waived_by = $5, waived_at = $6, waiver_reason = $7, updated_at = $8
			WHERE missing_doc_id = $9
		`
		_, err := dblib.Exec(ctx, r.db, updateSQL, []any{
			string(status), resolvedBy, now, resolutionNotes,
			resolvedBy, now, waiverReason, now,
			missingDocID,
		})
		return err
	}

	// UPLOADED status
	updateSQL := `
		UPDATE proposal_missing_documents
		SET status = $1, resolved_by = $2, resolved_at = $3, resolution_notes = $4,
			uploaded_document_id = $5, updated_at = $6
		WHERE missing_doc_id = $7
	`
	_, err := dblib.Exec(ctx, r.db, updateSQL, []any{
		string(status), resolvedBy, now, resolutionNotes,
		uploadedDocID, now,
		missingDocID,
	})
	return err
}

// ============================================
// Status History Operations (E-015)
// ============================================

// GetStatusHistory retrieves the complete status history timeline for a proposal
// [STATUS-POL-001] Proposal status, [STATUS-POL-002] Proposal timeline
func (r *DocumentRepository) GetStatusHistory(ctx context.Context, proposalID int64) ([]domain.ProposalStatusHistory, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
	defer cancel()

	query := dblib.Psql.Select(
		"history_id", "proposal_id", "from_status", "to_status",
		"changed_by", "changed_at", "comments", "version", "metadata",
	).From("proposal_status_history").
		Where(sq.Eq{"proposal_id": proposalID}).
		OrderBy("changed_at ASC")

	return dblib.SelectRows(ctx, r.db, query, pgx.RowToStructByName[domain.ProposalStatusHistory])
}

// GetPolicyNumberByProposalID retrieves the policy_number from proposal_issuance table
// [STATUS-POL-003] Policy status lookup
// NOTE: policy_number lives in proposal_issuance (E-007F), NOT in proposals (E-007).
// FR-POL-023: Policy number is generated and stored in issuance phase.
func (r *DocumentRepository) GetPolicyNumberByProposalID(ctx context.Context, proposalID int64) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutShort"))
	defer cancel()

	query := dblib.Psql.Select("COALESCE(policy_number, '') as policy_number").
		From("proposal_issuance").
		Where(sq.Eq{"proposal_id": proposalID})

	type pnResult struct {
		PolicyNumber string `db:"policy_number"`
	}

	result, err := dblib.SelectOne(ctx, r.db, query, pgx.RowToStructByName[pnResult])
	if err != nil {
		// Only ignore "no rows" — proposal_issuance row may not exist if
		// the policy has not been issued yet. All other errors (connection,
		// timeout, etc.) must be surfaced to avoid masking outages.
		if err == pgx.ErrNoRows {
			return "", nil
		}
		return "", fmt.Errorf("failed to query policy number for proposal %d: %w", proposalID, err)
	}
	return result.PolicyNumber, nil
}
