package port

import "context"

// PolicyLookupRepository provides read access to policy data for PM notification.
// This interface is implemented by a repository that queries the policy table
// (or a shared read-only view) to find active policies for a given customer.
type PolicyLookupRepository interface {
	// GetActivePoliciesByCustomerID returns all active policy numbers for a customer.
	// Used by NotifyPolicyManagement activity to signal each policy's PLW workflow.
	GetActivePoliciesByCustomerID(ctx context.Context, customerID int64) ([]string, error)
}
