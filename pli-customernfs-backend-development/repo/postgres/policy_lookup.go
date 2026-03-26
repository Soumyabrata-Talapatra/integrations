package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// PolicyLookupRepository implements port.PolicyLookupRepository.
// Queries a shared read-only view or the policy table to find active policies by customer.
type PolicyLookupRepository struct {
	db *pgxpool.Pool
}

// NewPolicyLookupRepository creates a new PolicyLookupRepository.
func NewPolicyLookupRepository(db *pgxpool.Pool) *PolicyLookupRepository {
	return &PolicyLookupRepository{db: db}
}

// GetActivePoliciesByCustomerID returns all active (non-terminal) policy numbers for a customer.
// Queries the policy_customer_mapping table or policy table.
// Terminal statuses excluded: VOID, SURRENDERED, TERMINATED_SURRENDER, MATURED,
// DEATH_CLAIM_SETTLED, FLC_CANCELLED, CANCELLED_DEATH, CONVERTED.
func (r *PolicyLookupRepository) GetActivePoliciesByCustomerID(
	ctx context.Context,
	customerID int64,
) ([]string, error) {
	query := `
		SELECT policy_number
		FROM nfs.policy_customer_mapping
		WHERE customer_id = $1
		  AND is_active = true
		ORDER BY policy_number
	`
	rows, err := r.db.Query(ctx, query, customerID)
	if err != nil {
		return nil, fmt.Errorf("GetActivePoliciesByCustomerID: %w", err)
	}
	defer rows.Close()

	var policies []string
	for rows.Next() {
		var pn string
		if err := rows.Scan(&pn); err != nil {
			return nil, fmt.Errorf("GetActivePoliciesByCustomerID scan: %w", err)
		}
		policies = append(policies, pn)
	}
	return policies, rows.Err()
}
