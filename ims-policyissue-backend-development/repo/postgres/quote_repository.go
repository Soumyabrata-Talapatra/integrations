package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"policy-issue-service/core/domain"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	config "gitlab.cept.gov.in/it-2.0-common/api-config"
	log "gitlab.cept.gov.in/it-2.0-common/api-log"
	dblib "gitlab.cept.gov.in/it-2.0-common/n-api-db"
)

// QuoteRepository handles quote-related database operations
type QuoteRepository struct {
	db  *dblib.DB
	cfg *config.Config
}

// NewQuoteRepository creates a new QuoteRepository instance
func NewQuoteRepository(db *dblib.DB, cfg *config.Config) *QuoteRepository {
	return &QuoteRepository{db: db, cfg: cfg}
}

// GetProducts retrieves products from the catalog
func (r *QuoteRepository) GetProducts(ctx context.Context, policyType string, isActive *bool) ([]domain.Product, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
	defer cancel()

	query := dblib.Psql.Select(
		"product_code", "product_name", "product_type", "product_category",
		"min_sum_assured", "max_sum_assured", "min_entry_age", "max_entry_age",
		"max_maturity_age", "min_term", "premium_ceasing_age_options", "available_frequencies",
		"medical_sa_threshold", "is_sa_decrease_allowed", "is_active",
		"effective_from", "effective_to", "description", "created_at", "updated_at", "deleted_at",
	).From("product_catalog")

	if policyType != "" {
		query = query.Where(sq.Eq{"product_type": policyType})
	}

	if isActive != nil {
		query = query.Where(sq.Eq{"is_active": *isActive})
	} else {
		query = query.Where(sq.Eq{"is_active": true})
	}

	query = query.Where(sq.Or{
		sq.Expr("effective_to IS NULL"),
		sq.GtOrEq{"effective_to": time.Now()},
	})

	query = query.OrderBy("product_name")

	return dblib.SelectRows(ctx, r.db, query, pgx.RowToStructByNameLax[domain.Product])
}

// GetProductByCode retrieves a single product by code
func (r *QuoteRepository) GetProductByCode(ctx context.Context, productCode string) (domain.Product, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select(
		"product_code", "product_name", "product_type", "product_category",
		"min_sum_assured", "max_sum_assured", "min_entry_age", "max_entry_age",
		"max_maturity_age", "min_term", "premium_ceasing_age_options",
		"available_frequencies", "medical_sa_threshold", "is_sa_decrease_allowed",
		"is_active", "effective_from", "effective_to", "description", "created_at", "updated_at", "deleted_at",
	).From("product_catalog").Where(sq.Eq{"product_code": productCode})

	return dblib.SelectOne(ctx, r.db, query, pgx.RowToStructByNameLax[domain.Product])
}

// GetPremiumRate retrieves premium rate from Sankalan table
func (r *QuoteRepository) GetPremiumRate(ctx context.Context, productCode string, productCategory string,
	age int, gender string, periodicity string, lookupField string, // "term" or "premium_ceasing_age"
	lookupValue int) (float64, int, error) {
	log.Info(ctx, "age is: %d", age)
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
	defer cancel()

	// -----------------------------
	// SAFETY CHECK (prevent SQL injection)
	// -----------------------------
	if lookupField != "term" && lookupField != "premium_ceasing_age" {
		return 0, 0, fmt.Errorf("invalid lookup field")
	}

	// -----------------------------
	// DYNAMIC QUERY
	// -----------------------------
	query := fmt.Sprintf(`
		SELECT 
			sum_assd::int,
			premium_per_sumassd
		FROM policy_issue.premium_rate_master
		WHERE product_code = $1
		AND product_category = $2
		AND age_at_entry = $3
		AND (gender = $4 OR gender = 'ALL')
		AND periodicity = $5
		AND is_active = true
		AND %s = $6
		ORDER BY created_at DESC
		LIMIT 1
	`, lookupField)

	var sumAssd int
	var rate float64

	err := r.db.QueryRow(
		ctx,
		query,
		productCode,
		productCategory,
		age,
		gender,
		periodicity,
		lookupValue,
	).Scan(&sumAssd, &rate)

	if err != nil {

		if errors.Is(err, pgx.ErrNoRows) {
			return 0, 0, fmt.Errorf(
				"premium rate not found for product=%s category=%s age=%d gender=%s periodicity=%s %s=%d",
				productCode,
				productCategory,
				age,
				gender,
				periodicity,
				lookupField,
				lookupValue,
			)
		}

		return 0, 0, fmt.Errorf("querying premium rate: %w", err)
	}

	// -----------------------------
	// VALIDATE RESULT
	// -----------------------------
	if sumAssd <= 0 {
		return 0, 0, fmt.Errorf("invalid sum_assd configuration in database")
	}

	if rate <= 0 {
		return 0, 0, fmt.Errorf("invalid premium rate configuration in database")
	}

	return rate, sumAssd, nil
}
func (r *QuoteRepository) GetJointLifeAgeAddition(ctx context.Context, ageDiff int) (int, error) {

	var ageAddition int

	query := `
        SELECT age_addition_to_lower_age
        FROM policy_issue.ys_average_age
        WHERE age_difference = $1
          AND status = 'ACTIVE'
        LIMIT 1
    `

	err := r.db.QueryRow(ctx, query, ageDiff).Scan(&ageAddition)
	if err != nil {
		return 0, err
	}

	return ageAddition, nil
}

// func (r *QuoteRepository) GetPremiumRate(ctx context.Context, productCode string, age int, gender domain.Gender, term int) (float64, string, error) {
// 	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
// 	defer cancel()

// 	// Use raw SQL for scalar value retrieval
// 	sql := `SELECT sum_assd,premium_per_sumassd FROM policy_issue.premium_rates
// 		WHERE product_code = $1 AND age_at_entry = $2 AND (gender = $3 OR gender = 'ALL') AND term = $4
// 		AND effective_from <= $5
// 		AND (effective_to IS NULL OR effective_to > $5)
// 		ORDER BY effective_from DESC LIMIT 1`

// 	type rateResult struct {
// 		Rate    float64 `db:"premium_per_sumassd"`
// 		SumAssd string  `db:"sum_assd"`
// 	}

// 	result, err := dblib.ExecReturn(ctx, r.db, sql, []any{productCode, age, gender, term, time.Now()}, pgx.RowToStructByName[rateResult])
// 	if err != nil {
// 		if err == pgx.ErrNoRows {
// 			return 0, "", fmt.Errorf("premium rate not found for product %s, age %d, gender %s, term %d", productCode, age, gender, term)
// 		}
// 		return 0, "", err
// 	}

// 	return result.Rate, result.SumAssd, nil
// }

// CreateQuote inserts a new quote into the database and hydrates the quote with generated fields

type quoteInsertResult struct {
	QuoteID        int64     `db:"quote_id"`
	QuoteRefNumber string    `db:"quote_ref_number"`
	CreatedAt      time.Time `db:"created_at"`
	UpdatedAt      time.Time `db:"updated_at"`
}

func (r *QuoteRepository) CreateQuote(ctx context.Context, quote *domain.Quote) error {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
	defer cancel()

	// Generate quote reference number
	quoteRef, err := r.generateQuoteRefNumber(ctx, quote.PolicyType)
	if err != nil {
		return fmt.Errorf("failed to generate quote reference: %w", err)
	}

	ins := dblib.Psql.Insert("policy_issue.quote").
		Columns(
			"quote_ref_number", "product_code", "policy_type", "customer_id",
			"proposer_name", "proposer_dob", "proposer_gender", "proposer_mobile", "proposer_email",
			"sum_assured", "policy_term", "payment_frequency",
			"base_premium", "gst_amount", "total_payable",
			"maturity_value", "bonus_rate", "channel", "status",
			"created_by", "expires_at", "pdf_document_id", "rebate",
		).
		Values(
			quoteRef, quote.ProductCode, quote.PolicyType, quote.CustomerID,
			quote.ProposerName, quote.ProposerDOB, quote.ProposerGender, quote.ProposerMobile, quote.ProposerEmail,
			quote.SumAssured, quote.PolicyTerm, quote.PaymentFrequency,
			quote.BasePremium, quote.GSTAmount, quote.TotalPayable,
			quote.MaturityValue, quote.BonusRate, quote.Channel, quote.Status,
			quote.CreatedBy, quote.ExpiresAt, quote.PDFDocumentID, quote.Rebate,
		).
		Suffix("RETURNING quote_id, quote_ref_number, created_at, updated_at")
		// Suffix("RETURNING *")

	// Use InsertReturning to get the generated fields back
	result, err := dblib.InsertReturning(ctx, r.db, ins, pgx.RowToStructByName[quoteInsertResult])
	if err != nil {
		return fmt.Errorf("failed to insert quote: %w", err)
	}

	// Hydrate the quote with generated fields
	quote.QuoteID = result.QuoteID
	quote.QuoteRefNumber = result.QuoteRefNumber
	quote.CreatedAt = result.CreatedAt
	quote.UpdatedAt = result.UpdatedAt

	return nil
}

// GetQuoteByID retrieves a quote by its ID
func (r *QuoteRepository) GetQuoteByID(ctx context.Context, quoteID int64) (*domain.Quote, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("quote_id", "quote_ref_number", "product_code", "policy_type", "customer_id",
		"proposer_name", "proposer_dob", "proposer_gender", "proposer_mobile", "proposer_email",
		"sum_assured", "policy_term", "payment_frequency", "base_premium", "gst_amount", "total_payable",
		"maturity_value", "bonus_rate", "channel", "status", "created_by", "created_at", "expires_at",
	).From("quote").Where(sq.Eq{"quote_id": quoteID})

	quote, err := dblib.SelectOne(
		ctx,
		r.db,
		query,
		pgx.RowToStructByNameLax[domain.Quote], // ✅ struct type only
	)
	if err != nil {
		return nil, err
	}

	return &quote, nil
}

// GetQuoteByRefNumber retrieves a quote by reference number
func (r *QuoteRepository) GetQuoteByRefNumber(ctx context.Context, refNumber string) (domain.Quote, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.Select("*").From("quote").Where(sq.Eq{"quote_ref_number": refNumber})

	return dblib.SelectOne(ctx, r.db, query, pgx.RowToStructByName[domain.Quote])
}

// UpdateQuoteStatus updates the status of a quote
func (r *QuoteRepository) UpdateQuoteStatus(ctx context.Context, quoteID int64, status domain.QuoteStatus, proposalID *int64) error {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	updateMap := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}

	if proposalID != nil {
		updateMap["converted_proposal_id"] = *proposalID
	}

	query := dblib.Psql.Update("quote").SetMap(updateMap).Where(sq.Eq{"quote_id": quoteID})

	_, err := dblib.Update(ctx, r.db, query)
	return err
}

// generateQuoteRefNumber generates a unique quote reference number
func (r *QuoteRepository) generateQuoteRefNumber(ctx context.Context, policyType domain.PolicyType) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	// Use raw SQL for sequence generation
	sql := "SELECT nextval('policy_issue.policy_issue_seq') as seq"

	type seqResult struct {
		Seq int64 `db:"seq"`
	}

	result, err := dblib.ExecReturn(ctx, r.db, sql, []any{}, pgx.RowToStructByName[seqResult])
	if err != nil {
		return "", err
	}

	refNumber := fmt.Sprintf("QT-%s-%d-%08d", policyType, time.Now().Year(), result.Seq)
	return refNumber, nil
}

func (r *QuoteRepository) GetRebate(ctx context.Context, productCode string, sumAssured int,
) (float64, error) {

	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutLow"))
	defer cancel()

	query := dblib.Psql.
		Select("min_sum_assd", "multiples_of_sum_assd", "rebate_amount").
		From("policy_issue.rebate_master").
		Where(sq.Eq{
			"product_code": productCode,
			"status":       "ACTIVE",
		}).
		OrderBy("min_sum_assd DESC").
		Limit(1)

	type rebateRow struct {
		MinSumAssd         float64 `db:"min_sum_assd"`
		MultiplesOfSumAssd float64 `db:"multiples_of_sum_assd"`
		RebateAmount       float64 `db:"rebate_amount"`
	}

	row, err := dblib.SelectOne(ctx, r.db, query, pgx.RowToStructByNameLax[rebateRow])
	if err != nil {

		// No rebate config found
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, nil
		}

		return 0, err
	}

	if float64(sumAssured) < row.MinSumAssd {
		return 0, nil
	}

	if productCode == "1005" {

		extraUnits := (float64(sumAssured) - row.MinSumAssd) / row.MultiplesOfSumAssd
		return row.RebateAmount + (extraUnits * row.RebateAmount), nil
	}

	// For other products (₹1 per ₹multiple)
	units := float64(sumAssured) / row.MultiplesOfSumAssd
	return units * row.RebateAmount, nil
}

func (r *QuoteRepository) GetAvailableTerms(ctx context.Context, productCode string, ageAtEntry int,
) ([]int, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed"))
	defer cancel()

	query := sq.Select("DISTINCT term").
		From("policy_issue.premium_rate_master").
		Where(sq.Eq{
			"product_code": productCode,
			"age_at_entry": ageAtEntry,
			"is_active":    true,
		}).
		Where("term IS NOT NULL").
		OrderBy("term ASC").
		PlaceholderFormat(sq.Dollar)

	return dblib.SelectRows(ctx, r.db, query, pgx.RowTo[int])
}
func (r *QuoteRepository) GetAvailablePremiumCeasingAges(ctx context.Context, productCode string, ageAtEntry int,
) ([]int, error) {
	ctx, cancel := context.WithTimeout(ctx, r.cfg.GetDuration("db.QueryTimeoutMed")) // <-- change here
	defer cancel()
	query := sq.Select("DISTINCT premium_ceasing_age").
		From("policy_issue.premium_rate_master").
		Where(sq.Eq{
			"product_code": productCode,
			"age_at_entry": ageAtEntry,
			"is_active":    true,
		}).
		Where("premium_ceasing_age IS NOT NULL").
		OrderBy("premium_ceasing_age ASC").
		PlaceholderFormat(sq.Dollar)

	sql, args, _ := query.ToSql()
	log.Info(ctx, "SQL: %s, args: %v", sql, args)

	results, err := dblib.SelectRows(ctx, r.db, query, pgx.RowTo[int])
	if err != nil {
		log.Error(ctx, "GetAvailablePremiumCeasingAges error: %v", err)
		return nil, err
	}

	log.Info(ctx, "Results: %v", results)
	return results, nil

}

