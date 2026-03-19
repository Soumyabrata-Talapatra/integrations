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

// DocumentRepository handles all database operations for surrender documents
// Functional Requirement: FR-SUR-004
type DocumentRepository struct {
	db  *dblib.DB
	cfg *config.Config
}

// NewDocumentRepository creates a new document repository
func NewDocumentRepository(db *dblib.DB, cfg *config.Config) *DocumentRepository {
	return &DocumentRepository{
		db:  db,
		cfg: cfg,
	}
}

const documentTable = "surrender_documents"

// Create inserts a new document
// Validation Rule: VR-SUR-007, VR-SUR-008
func (r *DocumentRepository) Create(ctx context.Context, data domain.SurrenderDocument) (domain.SurrenderDocument, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Insert(documentTable).
		Columns(
			"surrender_request_id", "document_type", "document_name",
			"document_path", "file_size_bytes", "mime_type", "metadata",
		).
		Values(
			data.SurrenderRequestID, data.DocumentType, data.DocumentName,
			data.DocumentPath, data.FileSizeBytes, data.MimeType, data.Metadata,
		).
		Suffix("RETURNING *").
		PlaceholderFormat(sq.Dollar)

	result, err := dblib.InsertReturning(ctx, r.db, query, pgx.RowToStructByName[domain.SurrenderDocument])
	if err != nil {
		return result, err
	}

	return result, nil
}

// FindByID retrieves a document by ID
func (r *DocumentRepository) FindByID(ctx context.Context, id uuid.UUID) (domain.SurrenderDocument, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("*").
		From(documentTable).
		Where(sq.Eq{"id": id}).
		Where(sq.Eq{"deleted_at": nil}).
		PlaceholderFormat(sq.Dollar)

	result, err := dblib.SelectOne(ctx, r.db, query, pgx.RowToStructByName[domain.SurrenderDocument])
	if err != nil {
		return result, err
	}

	return result, nil
}

// FindBySurrenderRequestID retrieves all documents for a surrender request
// Business Rule: BR-SUR-015 (document requirements)
func (r *DocumentRepository) FindBySurrenderRequestID(ctx context.Context, surrenderRequestID uuid.UUID) ([]domain.SurrenderDocument, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
	defer cancel()

	query := dblib.Psql.Select("*").
		From(documentTable).
		Where(sq.Eq{"surrender_request_id": surrenderRequestID}).
		Where(sq.Eq{"deleted_at": nil}).
		OrderBy("created_at ASC").
		PlaceholderFormat(sq.Dollar)

	results, err := dblib.SelectRows(ctx, r.db, query, pgx.RowToStructByName[domain.SurrenderDocument])
	if err != nil {
		return results, err
	}

	return results, nil
}

// CheckDocumentExists checks if a document type exists for a surrender request
// Validation Rule: VR-SUR-009 (duplicate document prevention)
func (r *DocumentRepository) CheckDocumentExists(ctx context.Context, surrenderRequestID uuid.UUID, docType domain.DocumentType) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("COUNT(*)").
		From(documentTable).
		Where(sq.Eq{
			"surrender_request_id": surrenderRequestID,
			"document_type":        docType,
			"deleted_at":           nil,
		}).
		PlaceholderFormat(sq.Dollar)

	var count int64
	sql, args, err := query.ToSql()
	if err != nil {
		return false, err
	}

	err = r.db.QueryRow(ctx, sql, args...).Scan(&count)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// VerifyDocument marks a document as verified
// Business Rule: BR-FS-015 (CPC document verification)
func (r *DocumentRepository) VerifyDocument(ctx context.Context, id uuid.UUID, verifiedBy uuid.UUID) (domain.SurrenderDocument, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	now := time.Now()
	query := dblib.Psql.Update(documentTable).
		Set("verified", true).
		Set("verified_by", verifiedBy).
		Set("verified_at", now).
		Where(sq.Eq{"id": id}).
		Suffix("RETURNING *").
		PlaceholderFormat(sq.Dollar)

	result, err := dblib.UpdateReturning(ctx, r.db, query, pgx.RowToStructByName[domain.SurrenderDocument])
	if err != nil {
		return result, err
	}

	return result, nil
}

// RejectDocument marks a document as rejected
func (r *DocumentRepository) RejectDocument(ctx context.Context, id uuid.UUID, reason string) (domain.SurrenderDocument, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Update(documentTable).
		Set("verified", false).
		Set("rejection_reason", reason).
		Where(sq.Eq{"id": id}).
		Suffix("RETURNING *").
		PlaceholderFormat(sq.Dollar)

	result, err := dblib.UpdateReturning(ctx, r.db, query, pgx.RowToStructByName[domain.SurrenderDocument])
	if err != nil {
		return result, err
	}

	return result, nil
}

// CountVerifiedDocuments counts verified documents for a surrender request
// Validation Rule: VR-SUR-010 (all documents must be verified before submission)
func (r *DocumentRepository) CountVerifiedDocuments(ctx context.Context, surrenderRequestID uuid.UUID) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("COUNT(*)").
		From(documentTable).
		Where(sq.Eq{
			"surrender_request_id": surrenderRequestID,
			"verified":             true,
			"deleted_at":           nil,
		}).
		PlaceholderFormat(sq.Dollar)

	var count int64
	sql, args, err := query.ToSql()
	if err != nil {
		return 0, err
	}

	err = r.db.QueryRow(ctx, sql, args...).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}

// SoftDelete soft deletes a document
func (r *DocumentRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Update(documentTable).
		Set("deleted_at", time.Now()).
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Dollar)

	_, err := dblib.Update(ctx, r.db, query)
	return err
}
