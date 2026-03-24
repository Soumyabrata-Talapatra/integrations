package repo

import (
	"context"
	"fmt"
	"time"

	"plirevival/core/domain"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	config "gitlab.cept.gov.in/it-2.0-common/api-config"
	dblib "gitlab.cept.gov.in/it-2.0-common/n-api-db"
)

type PolicyRepository struct {
	db  *dblib.DB
	cfg *config.Config
}

func NewPolicyRepository(db *dblib.DB, cfg *config.Config) *PolicyRepository {
	return &PolicyRepository{db: db, cfg: cfg}
}

const (
	policiesTable            = "common.policies"
	systemConfigurationTable = "common.system_configuration"
	tigerbeetleAccountsTable = "common.tigerbeetle_accounts"
)

// GetPolicyByNumber retrieves a policy by policy number
func (r *PolicyRepository) GetPolicyByNumber(ctx context.Context, policyNumber string) (domain.Policy, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	q := dblib.Psql.Select(
		"policy_number", "customer_id", "customer_name", "product_code", "product_name",
		"policy_status", "premium_frequency", "premium_amount", "sum_assured",
		"paid_to_date", "maturity_date", "date_of_commencement",
		"revival_count", "last_revival_date",
		"created_at", "updated_at",
	).
		From(policiesTable).
		Where(sq.Eq{"policy_number": policyNumber})

	return dblib.SelectOne(ctx, r.db, q, pgx.RowToStructByName[domain.Policy])
}

// UpdatePolicyStatus updates policy status
func (r *PolicyRepository) UpdatePolicyStatus(ctx context.Context, policyNumber string, status string) error {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	upd := dblib.Psql.Update(policiesTable).
		Set("policy_status", status).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"policy_number": policyNumber})

	_, err := dblib.Update(ctx, r.db, upd)
	return err
}

// UpdatePolicyPaidToDate updates policy paid_to_date
func (r *PolicyRepository) UpdatePolicyPaidToDate(ctx context.Context, policyNumber string, paidToDate *time.Time) error {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	upd := dblib.Psql.Update(policiesTable).
		Set("paid_to_date", paidToDate).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"policy_number": policyNumber})

	_, err := dblib.Update(ctx, r.db, upd)
	return err
}

// IncrementRevivalCount increments revival count for a policy
func (r *PolicyRepository) IncrementRevivalCount(ctx context.Context, policyNumber string, lastRevivalDate *time.Time) error {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	upd := dblib.Psql.Update(policiesTable).
		Set("revival_count", sq.Expr("revival_count + 1")).
		Set("last_revival_date", lastRevivalDate).
		Set("updated_at", sq.Expr("NOW()")).
		Where(sq.Eq{"policy_number": policyNumber})

	_, err := dblib.Update(ctx, r.db, upd)
	return err
}

// GetMaxRevivalsAllowed retrieves max revivals allowed configuration
func (r *PolicyRepository) GetMaxRevivalsAllowed(ctx context.Context) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	q := dblib.Psql.Select("config_value").
		From(systemConfigurationTable).
		Where(sq.Eq{"config_key": "max_revivals_allowed"})

	type configResult struct {
		ConfigValue string `db:"config_value"`
	}

	result, err := dblib.SelectOne(ctx, r.db, q, pgx.RowToStructByName[configResult])
	if err != nil {
		return 0, err
	}

	// Parse integer from config_value (stored as string)
	var maxRevivals int
	if _, err := fmt.Sscanf(result.ConfigValue, "%d", &maxRevivals); err != nil {
		return 2, nil // Default to 2 if parsing fails
	}

	return maxRevivals, nil
}

// CheckPolicyStatus checks if policy is in required status
func (r *PolicyRepository) CheckPolicyStatus(ctx context.Context, policyNumber string, status string) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	q := dblib.Psql.Select("COUNT(*)").
		From(policiesTable).
		Where(sq.And{
			sq.Eq{"policy_number": policyNumber},
			sq.Eq{"policy_status": status},
		})

	type countResult struct {
		Count int `db:"count"`
	}

	result, err := dblib.SelectOne(ctx, r.db, q, pgx.RowToStructByName[countResult])
	if err != nil {
		return false, err
	}

	return result.Count > 0, nil
}

// ValidatePolicyForRevival performs batched policy validation for revival eligibility
// Combines 3 queries into 1 DB round trip: GetPolicyByNumber + CheckOngoingRevival + GetMaxRevivalsAllowed
func (r *PolicyRepository) ValidatePolicyForRevival(ctx context.Context, policyNumber string) (domain.PolicyValidationResult, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	// Single query combining policy data, config value, and ongoing revival count
	// Using CROSS JOIN for config and LEFT JOIN LATERAL for ongoing revival count
	q := dblib.Psql.Select(
		"p.policy_number", "p.customer_id", "p.customer_name", "p.product_code", "p.product_name",
		"p.policy_status", "p.premium_frequency", "p.premium_amount", "p.sum_assured",
		"p.paid_to_date", "p.maturity_date", "p.date_of_commencement",
		"p.revival_count", "p.last_revival_date", "p.created_at", "p.updated_at",
		"c.config_value as max_revivals_config",
		"COALESCE(r.ongoing_count, 0) as ongoing_revival_count",
	).
		From(policiesTable + " p").
		// CROSS JOIN for config value
		JoinClause("CROSS JOIN (SELECT config_value FROM common.system_configuration WHERE config_key = 'max_revivals_allowed') c").
		// LEFT JOIN LATERAL for ongoing revival count
		JoinClause("LEFT JOIN LATERAL (SELECT COUNT(*)::int as ongoing_count FROM revival.revival_requests WHERE policy_number = p.policy_number AND current_status NOT IN ('COMPLETED', 'WITHDRAWN', 'TERMINATED', 'REJECTED', 'DEFAULTED')) r ON true").
		Where(sq.Eq{"p.policy_number": policyNumber})

	type validationRow struct {
		// Policy fields
		PolicyNumber       string     `db:"policy_number"`
		CustomerID         string     `db:"customer_id"`
		CustomerName       string     `db:"customer_name"`
		ProductCode        string     `db:"product_code"`
		ProductName        string     `db:"product_name"`
		PolicyStatus       string     `db:"policy_status"`
		PremiumFrequency   string     `db:"premium_frequency"`
		PremiumAmount      float64    `db:"premium_amount"`
		SumAssured         float64    `db:"sum_assured"`
		PaidToDate         *time.Time `db:"paid_to_date"`
		MaturityDate       time.Time  `db:"maturity_date"`
		DateOfCommencement time.Time  `db:"date_of_commencement"`
		RevivalCount       int        `db:"revival_count"`
		LastRevivalDate    *time.Time `db:"last_revival_date"`
		CreatedAt          time.Time  `db:"created_at"`
		UpdatedAt          time.Time  `db:"updated_at"`
		// Config and validation fields
		MaxRevivalsConfig   string `db:"max_revivals_config"`
		OngoingRevivalCount int    `db:"ongoing_revival_count"`
	}

	row, err := dblib.SelectOne(ctx, r.db, q, pgx.RowToStructByName[validationRow])
	if err != nil {
		return domain.PolicyValidationResult{}, fmt.Errorf("failed to query policy validation: %w", err)
	}

	// Parse max revivals from config
	var maxRevivals int
	if _, err := fmt.Sscanf(row.MaxRevivalsConfig, "%d", &maxRevivals); err != nil {
		maxRevivals = 2 // Default to 2 if parsing fails
	}

	// Build result
	result := domain.PolicyValidationResult{
		Policy: domain.Policy{
			PolicyNumber:       row.PolicyNumber,
			CustomerID:         row.CustomerID,
			CustomerName:       row.CustomerName,
			ProductCode:        row.ProductCode,
			ProductName:        row.ProductName,
			PolicyStatus:       row.PolicyStatus,
			PremiumFrequency:   row.PremiumFrequency,
			PremiumAmount:      row.PremiumAmount,
			SumAssured:         row.SumAssured,
			PaidToDate:         row.PaidToDate,
			MaturityDate:       row.MaturityDate,
			DateOfCommencement: row.DateOfCommencement,
			RevivalCount:       row.RevivalCount,
			LastRevivalDate:    row.LastRevivalDate,
			CreatedAt:          row.CreatedAt,
			UpdatedAt:          row.UpdatedAt,
		},
		MaxRevivalsAllowed:  maxRevivals,
		OngoingRevivalCount: row.OngoingRevivalCount,
	}

	return result, nil
}
