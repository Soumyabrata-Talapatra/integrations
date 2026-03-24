package repo

import (
	"context"

	"plirevival/core/domain"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	config "gitlab.cept.gov.in/it-2.0-common/api-config"
	dblib "gitlab.cept.gov.in/it-2.0-common/n-api-db"
)

type UserRepository struct {
	db  *dblib.DB
	cfg *config.Config
}

func NewUserRepository(db *dblib.DB, cfg *config.Config) *UserRepository {
	return &UserRepository{db: db, cfg: cfg}
}

const userTable = "users"

func (r *UserRepository) CreateUser(ctx context.Context, name, email string) (domain.User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	ins := dblib.Psql.Insert(userTable).
		Columns("name", "email").
		Values(name, email)

	sql, args, err := ins.ToSql()
	if err != nil {
		return domain.User{}, err
	}
	if _, err := r.db.Exec(ctx, sql, args...); err != nil {
		return domain.User{}, err
	}

	// Fetch inserted row by unique email
	sel := dblib.Psql.Select("id", "name", "email", "created_at", "updated_at").
		From(userTable).Where(sq.Eq{"email": email})
	return dblib.SelectOne(ctx, r.db, sel, pgx.RowToStructByName[domain.User])
}

func (r *UserRepository) GetAllUsers(ctx context.Context, skip, limit uint64) ([]domain.User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
	defer cancel()

	q := dblib.Psql.Select("id", "name", "email", "created_at", "updated_at").
		From(userTable).
		OrderBy("id ASC")
	if limit > 0 {
		q = q.Limit(limit).Offset(skip * limit)
	}
	return dblib.SelectRows(ctx, r.db, q, pgx.RowToStructByName[domain.User])
}

func (r *UserRepository) GetUserByID(ctx context.Context, id int64) (domain.User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	q := dblib.Psql.Select("id", "name", "email", "created_at", "updated_at").
		From(userTable).
		Where(sq.Eq{"id": id})

	return dblib.SelectOne(ctx, r.db, q, pgx.RowToStructByName[domain.User])
}

func (r *UserRepository) UpdateUserByID(ctx context.Context, id int64, name, email *string) (domain.User, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	b := dblib.Psql.Update(userTable)
	if name != nil && *name != "" {
		b = b.Set("name", *name)
	}
	if email != nil && *email != "" {
		b = b.Set("email", *email)
	}
	b = b.Set("updated_at", sq.Expr("NOW()"))
	b = b.Where(sq.Eq{"id": id})

	if _, err := dblib.Update(ctx, r.db, b); err != nil {
		return domain.User{}, err
	}

	sel := dblib.Psql.Select("id", "name", "email", "created_at", "updated_at").
		From(userTable).Where(sq.Eq{"id": id})
	return dblib.SelectOne(ctx, r.db, sel, pgx.RowToStructByName[domain.User])
}
