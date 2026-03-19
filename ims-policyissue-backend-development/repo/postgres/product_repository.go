package postgres

import (
	"context"

	"policy-issue-service/core/domain"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	config "gitlab.cept.gov.in/it-2.0-common/api-config"
	dblib "gitlab.cept.gov.in/it-2.0-common/n-api-db"
)

// ProductRepository handles product catalog operations
type ProductRepository struct {
	db  *dblib.DB
	cfg *config.Config
}

// NewProductRepository creates a new ProductRepository instance
func NewProductRepository(db *dblib.DB, cfg *config.Config) *ProductRepository {
	return &ProductRepository{db: db, cfg: cfg}
}

// GetAllProducts retrieves all active, non-deleted products.
// Applies deleted_at IS NULL filter to exclude soft-deleted entries.
func (r *ProductRepository) GetAllProducts(ctx context.Context, policyType string) ([]domain.Product, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
	defer cancel()

	query := dblib.Psql.Select("*").From("product_catalog").
		Where(sq.Eq{"is_active": true}).
		Where(sq.Eq{"deleted_at": nil}) // exclude soft-deleted products

	if policyType != "" {
		query = query.Where(sq.Eq{"product_type": policyType})
	}

	return dblib.SelectRows(ctx, r.db, query, pgx.RowToStructByName[domain.Product])
}

// GetProductByCode retrieves a product by its code
func (r *ProductRepository) GetProductByCode(ctx context.Context, productCode string) (*domain.Product, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("*").From("product_catalog").Where(sq.Eq{"product_code": productCode})

	product, err := dblib.SelectOne(ctx, r.db, query, pgx.RowToStructByName[domain.Product])
	if err != nil {
		return nil, err
	}

	return &product, nil
}

// ValidateProductEligibility checks if product is eligible for given parameters
func (r *ProductRepository) ValidateProductEligibility(ctx context.Context, productCode string, age int, sa float64, term int) (bool, string, error) {
	product, err := r.GetProductByCode(ctx, productCode)
	if err != nil {
		return false, "", err
	}

	if !product.IsEligibleAge(age) {
		return false, "Age not eligible for this product", nil
	}

	if !product.IsEligibleSA(sa) {
		return false, "Sum assured outside product limits", nil
	}

	// Additional validation for term can be added here

	return true, "", nil
}

func (r *ProductRepository) GetTermAndPremiumCeasingAge(ctx context.Context,
	productCode string, ageAtEntry int) ([]domain.TermOrPremiumCeasingAge, error) {

	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.
		Select("term", "premium_ceasing_age","periodicity").
		From("policy_issue.premium_rate_master").
		Distinct().
		Where(sq.Eq{
			"product_code": productCode,
			"age_at_entry": ageAtEntry,
			"is_active":    true,
		}).
		OrderBy("term", "premium_ceasing_age")

	data, err := dblib.SelectRows(ctx, r.db, query,
		pgx.RowToStructByNameLax[domain.TermOrPremiumCeasingAge])
	if err != nil {
		return nil, err
	}

	return data, nil
}

