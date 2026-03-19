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

// BulkUploadRepository handles bulk upload batch database operations
// Phase 9: Bulk Upload APIs
type BulkUploadRepository struct {
	db  *dblib.DB
	cfg *config.Config
}

// NewBulkUploadRepository creates a new BulkUploadRepository instance
func NewBulkUploadRepository(db *dblib.DB, cfg *config.Config) *BulkUploadRepository {
	return &BulkUploadRepository{db: db, cfg: cfg}
}

// CreateBatch inserts a new bulk upload batch record
// Returns the generated batch_id
func (r *BulkUploadRepository) CreateBatch(ctx context.Context, batch *domain.BulkUploadBatch) error {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
	defer cancel()

	now := time.Now()
	insertSQL := `
		INSERT INTO bulk_upload_batch (
			file_name, total_rows, success_count, failure_count,
			status, uploaded_by, uploaded_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING batch_id, uploaded_at
	`

	type insertResult struct {
		BatchID    int64     `db:"batch_id"`
		UploadedAt time.Time `db:"uploaded_at"`
	}
	result, err := dblib.ExecReturn(ctx, r.db, insertSQL, []any{
		batch.FileName, batch.TotalRows, 0, 0,
		string(domain.BulkUploadStatusProcessing), batch.UploadedBy, now,
	}, pgx.RowToStructByName[insertResult])
	if err != nil {
		return fmt.Errorf("failed to insert bulk upload batch: %w", err)
	}
	batch.BatchID = result.BatchID
	batch.UploadedAt = result.UploadedAt
	batch.Status = domain.BulkUploadStatusProcessing
	return nil
}

// GetBatchByID retrieves a bulk upload batch by ID
func (r *BulkUploadRepository) GetBatchByID(ctx context.Context, batchID int64) (*domain.BulkUploadBatch, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutShort"))
	defer cancel()

	query := dblib.Psql.Select(
		"batch_id", "file_name", "total_rows", "success_count",
		"failure_count", "error_report_doc_id", "status",
		"uploaded_by", "uploaded_at", "completed_at", "metadata",
	).From("bulk_upload_batch").
		Where(sq.Eq{"batch_id": batchID})

	batch, err := dblib.SelectOne(ctx, r.db, query, pgx.RowToStructByName[domain.BulkUploadBatch])
	if err != nil {
		return nil, err
	}
	return &batch, nil
}

// GetProposalNumbersByBatchID retrieves proposal numbers created for a batch
func (r *BulkUploadRepository) GetProposalNumbersByBatchID(ctx context.Context, batchID int64) ([]string, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
	defer cancel()

	query := dblib.Psql.Select("p.proposal_number").
		From("proposals p").
		Join("proposal_indexing pi ON p.proposal_id = pi.proposal_id").
		Where(sq.Eq{"pi.bulk_upload_batch_id": batchID}).
		OrderBy("pi.bulk_upload_row_number ASC")

	type pnRow struct {
		ProposalNumber string `db:"proposal_number"`
	}

	rows, err := dblib.SelectRows(ctx, r.db, query, pgx.RowToStructByName[pnRow])
	if err != nil {
		return nil, err
	}

	numbers := make([]string, len(rows))
	for i, row := range rows {
		numbers[i] = row.ProposalNumber
	}
	return numbers, nil
}
