package activities

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.temporal.io/sdk/client"

	"customer-nfs-service/core/port"
)

// PMNotificationActivities handles signaling Policy Management when NFS changes complete.
type PMNotificationActivities struct {
	temporalClient client.Client
	policyRepo     port.PolicyLookupRepository
}

// NewPMNotificationActivities creates a new PMNotificationActivities instance.
func NewPMNotificationActivities(tc client.Client, policyRepo port.PolicyLookupRepository) *PMNotificationActivities {
	return &PMNotificationActivities{
		temporalClient: tc,
		policyRepo:     policyRepo,
	}
}

// PolicyInfo holds minimal policy data needed for PM notification.
type PolicyInfo struct {
	PolicyNumber string `json:"policy_number" db:"policy_number"`
}

// LookupAffectedPoliciesInput is the input for LookupAffectedPolicies.
type LookupAffectedPoliciesInput struct {
	CustomerID int64 `json:"customer_id"`
}

// LookupAffectedPoliciesResult is the result of LookupAffectedPolicies.
type LookupAffectedPoliciesResult struct {
	Policies []PolicyInfo `json:"policies"`
}

// LookupAffectedPolicies fetches all active policy numbers for a given customer.
// This activity queries the policy lookup repository to find all policies
// that need to be notified when a customer's data changes.
func (a *PMNotificationActivities) LookupAffectedPolicies(
	ctx context.Context,
	input LookupAffectedPoliciesInput,
) (*LookupAffectedPoliciesResult, error) {
	policies, err := a.policyRepo.GetActivePoliciesByCustomerID(ctx, input.CustomerID)
	if err != nil {
		return nil, fmt.Errorf("LookupAffectedPolicies: %w", err)
	}
	result := make([]PolicyInfo, len(policies))
	for i, pn := range policies {
		result[i] = PolicyInfo{PolicyNumber: pn}
	}
	return &LookupAffectedPoliciesResult{Policies: result}, nil
}

// NotifyPolicyManagement signals each affected PolicyLifecycleWorkflow in PM
// with the customer NFS change result. Sends "customer-nfr-completed" signal
// to plw-{policy_number} for each affected policy.
//
// This is fire-and-forget from the NFS workflow's perspective — errors are
// logged but the NFS workflow does not fail if PM notification fails.
func (a *PMNotificationActivities) NotifyPolicyManagement(
	ctx context.Context,
	input NotifyPMInput,
) (*NotifyPMResult, error) {
	// Step 1: Look up affected policies for this customer.
	policies, err := a.policyRepo.GetActivePoliciesByCustomerID(ctx, input.CustomerID)
	if err != nil {
		return &NotifyPMResult{PoliciesNotified: 0, Success: false},
			fmt.Errorf("NotifyPolicyManagement: lookup policies: %w", err)
	}

	if len(policies) == 0 {
		return &NotifyPMResult{PoliciesNotified: 0, Success: true}, nil
	}

	// Step 2: Build the OperationCompletedSignal.
	signal := struct {
		RequestID       string          `json:"request_id"`
		RequestType     string          `json:"request_type"`
		Outcome         string          `json:"outcome"`
		StateTransition string          `json:"state_transition,omitempty"`
		OutcomePayload  json.RawMessage `json:"outcome_payload,omitempty"`
		CompletedAt     time.Time       `json:"completed_at"`
	}{
		RequestID:      input.RequestID,
		RequestType:    input.RequestType,
		Outcome:        input.Outcome,
		OutcomePayload: input.ChangePayload,
		CompletedAt:    time.Now().UTC(),
	}

	// Step 3: Signal each policy's PLW workflow.
	const signalName = "customer-nfr-completed"
	notified := 0
	for _, policyNumber := range policies {
		pmWorkflowID := fmt.Sprintf("plw-%s", policyNumber)
		err := a.temporalClient.SignalWorkflow(ctx, pmWorkflowID, "", signalName, signal)
		if err != nil {
			// Log but continue — don't fail the whole notification for one policy.
			fmt.Printf("NotifyPolicyManagement: failed to signal %s: %v\n", pmWorkflowID, err)
			continue
		}
		notified++
	}

	return &NotifyPMResult{
		PoliciesNotified: notified,
		Success:          notified == len(policies),
	}, nil
}
