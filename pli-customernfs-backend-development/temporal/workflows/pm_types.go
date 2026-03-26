package workflows

import (
	"encoding/json"
	"time"
)

// Signal channel name for Customer NFS → Policy Management completion notification.
const SignalCustomerNFRCompleted = "customer-nfr-completed"

// OperationCompletedSignal matches Policy Management's completion signal contract.
// Sent by Customer NFS workflows to plw-{policy_number} on the
// "customer-nfr-completed" channel when a name/address/mobile/email change completes.
type OperationCompletedSignal struct {
	RequestID       string          `json:"request_id"`
	RequestType     string          `json:"request_type"`       // ADDRESS_CHANGE, NAME_CHANGE, MOBILE_CHANGE, EMAIL_CHANGE
	Outcome         string          `json:"outcome"`            // APPROVED, REJECTED
	StateTransition string          `json:"state_transition,omitempty"`
	OutcomePayload  json.RawMessage `json:"outcome_payload,omitempty"`
	CompletedAt     time.Time       `json:"completed_at"`
}
