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

// ApprovalWorkflowRepository handles all database operations for approval workflow
// Business Rules: BR-FS-013, BR-FS-016
type ApprovalWorkflowRepository struct {
	db  *dblib.DB
	cfg *config.Config
}

// NewApprovalWorkflowRepository creates a new approval workflow repository
func NewApprovalWorkflowRepository(db *dblib.DB, cfg *config.Config) *ApprovalWorkflowRepository {
	return &ApprovalWorkflowRepository{
		db:  db,
		cfg: cfg,
	}
}

const approvalTaskTable = "approval_workflow_tasks"

// CreateTask creates a new approval task
// Functional Requirement: FR-FS-006
func (r *ApprovalWorkflowRepository) CreateTask(ctx context.Context, data domain.ApprovalWorkflowTask) (domain.ApprovalWorkflowTask, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Insert(approvalTaskTable).
		Columns(
			"surrender_request_id", "task_number", "office_code",
			"assigned_to", "priority", "metadata",
		).
		Values(
			data.SurrenderRequestID, data.TaskNumber, data.OfficeCode,
			data.AssignedTo, data.Priority, data.Metadata,
		).
		Suffix("RETURNING *").
		PlaceholderFormat(sq.Dollar)

	result, err := dblib.InsertReturning(ctx, r.db, query, pgx.RowToStructByName[domain.ApprovalWorkflowTask])
	if err != nil {
		return result, err
	}

	return result, nil
}

// FindByID retrieves a task by ID
func (r *ApprovalWorkflowRepository) FindByID(ctx context.Context, id uuid.UUID) (domain.ApprovalWorkflowTask, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("*").
		From(approvalTaskTable).
		Where(sq.Eq{"id": id}).
		PlaceholderFormat(sq.Dollar)

	result, err := dblib.SelectOne(ctx, r.db, query, pgx.RowToStructByName[domain.ApprovalWorkflowTask])
	if err != nil {
		return result, err
	}

	return result, nil
}

// FindBySurrenderRequestID retrieves task by surrender request ID
func (r *ApprovalWorkflowRepository) FindBySurrenderRequestID(ctx context.Context, surrenderRequestID uuid.UUID) (domain.ApprovalWorkflowTask, bool, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("*").
		From(approvalTaskTable).
		Where(sq.Eq{"surrender_request_id": surrenderRequestID}).
		OrderBy("created_at DESC").
		Limit(1).
		PlaceholderFormat(sq.Dollar)

	result, found, err := dblib.SelectOneOK(ctx, r.db, query, pgx.RowToStructByName[domain.ApprovalWorkflowTask])
	return result, found, err
}

// ListApprovalQueue retrieves approval queue for an office
// Business Rule: BR-FS-013 (approver queue filtering)
// Functional Requirement: FR-FS-008
func (r *ApprovalWorkflowRepository) ListApprovalQueue(ctx context.Context, officeCode string, skip, limit uint64) ([]domain.ApprovalWorkflowTask, uint64, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
	defer cancel()

	// Build where clause
	whereClause := sq.And{
		sq.Eq{"status": []string{string(domain.ApprovalTaskStatusPending), string(domain.ApprovalTaskStatusReserved)}},
	}

	// Add office_code filter only if not "ALL"
	// "ALL" means show tasks from all offices
	if officeCode != "ALL" {
		whereClause = append(whereClause, sq.Eq{"office_code": officeCode})
	}

	// Build queries
	countQuery := dblib.Psql.Select("COUNT(*)").
		From(approvalTaskTable).
		Where(whereClause)

	selectQuery := dblib.Psql.Select("*").
		From(approvalTaskTable).
		Where(whereClause).
		OrderBy("priority DESC", "created_at ASC").
		Limit(limit).
		Offset(skip)

	// Execute count query
	var totalCount uint64
	countSQL, countArgs, _ := countQuery.PlaceholderFormat(sq.Dollar).ToSql()
	err := r.db.QueryRow(ctx, countSQL, countArgs...).Scan(&totalCount)
	if err != nil {
		return nil, 0, err
	}

	// Execute select query
	results, err := dblib.SelectRows(ctx, r.db, selectQuery.PlaceholderFormat(sq.Dollar), pgx.RowToStructByName[domain.ApprovalWorkflowTask])
	if err != nil {
		return nil, 0, err
	}

	return results, totalCount, nil
}

// ReserveTask reserves a task for an approver
// Business Rule: BR-FS-016 (auto-reservation)
func (r *ApprovalWorkflowRepository) ReserveTask(ctx context.Context, id uuid.UUID, reservedBy uuid.UUID, expiresAt time.Time) (domain.ApprovalWorkflowTask, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	now := time.Now()
	query := dblib.Psql.Update(approvalTaskTable).
		Set("reserved", true).
		Set("reserved_by", reservedBy).
		Set("reserved_at", now).
		Set("reservation_expires_at", expiresAt).
		Set("status", domain.ApprovalTaskStatusReserved).
		Where(sq.Eq{"id": id}).
		Where(sq.Eq{"reserved": false}).
		Suffix("RETURNING *").
		PlaceholderFormat(sq.Dollar)

	result, err := dblib.UpdateReturning(ctx, r.db, query, pgx.RowToStructByName[domain.ApprovalWorkflowTask])
	if err != nil {
		return result, err
	}

	return result, nil
}

// ReleaseTask releases a reserved task
// Functional Requirement: FR-FS-009
func (r *ApprovalWorkflowRepository) ReleaseTask(ctx context.Context, id uuid.UUID) (domain.ApprovalWorkflowTask, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Update(approvalTaskTable).
		Set("reserved", false).
		Set("reserved_by", nil).
		Set("reserved_at", nil).
		Set("reservation_expires_at", nil).
		Set("status", domain.ApprovalTaskStatusPending).
		Where(sq.Eq{"id": id}).
		Suffix("RETURNING *").
		PlaceholderFormat(sq.Dollar)

	result, err := dblib.UpdateReturning(ctx, r.db, query, pgx.RowToStructByName[domain.ApprovalWorkflowTask])
	if err != nil {
		return result, err
	}

	return result, nil
}

// CompleteTask marks a task as completed
func (r *ApprovalWorkflowRepository) CompleteTask(ctx context.Context, id uuid.UUID, completedBy uuid.UUID) (domain.ApprovalWorkflowTask, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	now := time.Now()
	query := dblib.Psql.Update(approvalTaskTable).
		Set("status", domain.ApprovalTaskStatusCompleted).
		Set("completed_at", now).
		Set("completed_by", completedBy).
		Where(sq.Eq{"id": id}).
		Suffix("RETURNING *").
		PlaceholderFormat(sq.Dollar)

	result, err := dblib.UpdateReturning(ctx, r.db, query, pgx.RowToStructByName[domain.ApprovalWorkflowTask])
	if err != nil {
		return result, err
	}

	return result, nil
}

// EscalateTask escalates a task to higher authority
// Business Rule: BR-FS-013 (escalation hierarchy)
func (r *ApprovalWorkflowRepository) EscalateTask(ctx context.Context, id uuid.UUID, escalatedTo uuid.UUID, reason string) (domain.ApprovalWorkflowTask, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	now := time.Now()
	query := dblib.Psql.Update(approvalTaskTable).
		Set("escalated", true).
		Set("escalated_to", escalatedTo).
		Set("escalated_at", now).
		Set("escalation_reason", reason).
		Set("status", domain.ApprovalTaskStatusEscalated).
		Where(sq.Eq{"id": id}).
		Suffix("RETURNING *").
		PlaceholderFormat(sq.Dollar)

	result, err := dblib.UpdateReturning(ctx, r.db, query, pgx.RowToStructByName[domain.ApprovalWorkflowTask])
	if err != nil {
		return result, err
	}

	return result, nil
}

// ListExpiredReservations retrieves tasks with expired reservations
func (r *ApprovalWorkflowRepository) ListExpiredReservations(ctx context.Context) ([]domain.ApprovalWorkflowTask, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
	defer cancel()

	query := dblib.Psql.Select("*").
		From(approvalTaskTable).
		Where(sq.Eq{"reserved": true}).
		Where(sq.Lt{"reservation_expires_at": time.Now()}).
		PlaceholderFormat(sq.Dollar)

	results, err := dblib.SelectRows(ctx, r.db, query, pgx.RowToStructByName[domain.ApprovalWorkflowTask])
	if err != nil {
		return results, err
	}

	return results, nil
}

// FindTaskBySurrenderRequestID finds an approval task by surrender request ID
func (r *ApprovalWorkflowRepository) FindTaskBySurrenderRequestID(ctx context.Context, surrenderRequestID uuid.UUID) (domain.ApprovalWorkflowTask, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
	defer cancel()

	query := dblib.Psql.Select("*").
		From(approvalTaskTable).
		Where(sq.Eq{"surrender_request_id": surrenderRequestID}).
		OrderBy("created_at DESC").
		Limit(1).
		PlaceholderFormat(sq.Dollar)

	result, found, err := dblib.SelectOneOK(ctx, r.db, query, pgx.RowToStructByName[domain.ApprovalWorkflowTask])
	if err != nil {
		return result, err
	}
	if !found {
		return result, pgx.ErrNoRows
	}
	return result, nil
}
