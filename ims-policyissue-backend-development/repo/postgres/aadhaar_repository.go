package postgres

import (
	"context"
	"time"

	"policy-issue-service/core/domain"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	config "gitlab.cept.gov.in/it-2.0-common/api-config"
	dblib "gitlab.cept.gov.in/it-2.0-common/n-api-db"
)

// AadhaarRepository handles Aadhaar session operations
type AadhaarRepository struct {
	db  *dblib.DB
	cfg *config.Config
}

// NewAadhaarRepository creates a new AadhaarRepository instance
func NewAadhaarRepository(db *dblib.DB, cfg *config.Config) *AadhaarRepository {
	return &AadhaarRepository{db: db, cfg: cfg}
}

// StoreSession stores a new Aadhaar session
func (r *AadhaarRepository) StoreSession(ctx context.Context, session domain.AadhaarSession) error {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Insert("aadhaar_sessions").
		Columns("session_id", "transaction_id", "aadhaar_number", "user_data", "otp_verified", "created_at", "expires_at").
		Values(session.SessionID, session.TransactionID, session.AadhaarNumber, session.UserData, session.OTPVerified, session.CreatedAt, session.ExpiresAt)

	_, err := dblib.Insert(ctx, r.db, query)
	return err
}

// GetSessionByID retrieves a session by its ID if it hasn't expired
func (r *AadhaarRepository) GetSessionByID(ctx context.Context, sessionID string) (*domain.AadhaarSession, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("*").
		From("aadhaar_sessions").
		Where(sq.Eq{"session_id": sessionID}).
		Where(sq.Gt{"expires_at": time.Now()})

	session, err := dblib.SelectOne(ctx, r.db, query, pgx.RowToStructByName[domain.AadhaarSession])
	if err != nil {
		return nil, err
	}

	return &session, nil
}

// GetSessionByTransactionID retrieves a session by its transaction ID if it hasn't expired
func (r *AadhaarRepository) GetSessionByTransactionID(ctx context.Context, transactionID string) (*domain.AadhaarSession, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("*").
		From("aadhaar_sessions").
		Where(sq.Eq{"transaction_id": transactionID}).
		Where(sq.Gt{"expires_at": time.Now()})

	session, err := dblib.SelectOne(ctx, r.db, query, pgx.RowToStructByName[domain.AadhaarSession])
	if err != nil {
		return nil, err
	}

	return &session, nil
}

// UpdateSession updates an existing session
func (r *AadhaarRepository) UpdateSession(ctx context.Context, session domain.AadhaarSession) error {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Update("aadhaar_sessions").
		SetMap(map[string]interface{}{
			"user_data":    session.UserData,
			"otp_verified": session.OTPVerified,
			"expires_at":   session.ExpiresAt,
		}).
		Where(sq.Eq{"session_id": session.SessionID})

	_, err := dblib.Update(ctx, r.db, query)
	return err
}
