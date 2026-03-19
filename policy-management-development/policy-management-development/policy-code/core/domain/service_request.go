package domain

import (
	"encoding/json"
	"time"
)

// ============================================================================
// Request Type Constants (20 — DDL has 3 internal types not in Swagger)
// Source: DDL: request_type enum, Swagger: RequestType schema
// ============================================================================

const (
	// REST-exposed request types (17)
	RequestTypeSurrender           = "SURRENDER"
	RequestTypeLoan                = "LOAN"
	RequestTypeLoanRepayment       = "LOAN_REPAYMENT"
	RequestTypeRevival             = "REVIVAL"
	RequestTypeDeathClaim          = "DEATH_CLAIM"
	RequestTypeMaturityClaim       = "MATURITY_CLAIM"
	RequestTypeSurvivalBenefit     = "SURVIVAL_BENEFIT"
	RequestTypeCommutation         = "COMMUTATION"
	RequestTypeConversion          = "CONVERSION"
	RequestTypeFLC                 = "FLC"
	RequestTypePaidUp              = "PAID_UP"
	RequestTypeNominationChange    = "NOMINATION_CHANGE"
	RequestTypeBillingMethodChange = "BILLING_METHOD_CHANGE"
	RequestTypeAssignment          = "ASSIGNMENT"
	RequestTypeAddressChange       = "ADDRESS_CHANGE"
	RequestTypePremiumRefund       = "PREMIUM_REFUND"
	RequestTypeDuplicateBond       = "DUPLICATE_BOND"

	// Internal types (not in Swagger — triggered internally or by other services)
	RequestTypeForcedSurrender = "FORCED_SURRENDER" // Triggered by Loan Svc batch
	RequestTypeAdminVoid       = "ADMIN_VOID"        // Admin signal — BR-PM-073
	RequestTypeReopen          = "REOPEN"            // Admin reopen after cooling
)

// ============================================================================
// Request Category Constants
// Source: DDL: request_category enum
// ============================================================================

const (
	RequestCategoryFinancial    = "FINANCIAL"
	RequestCategoryNonFinancial = "NON_FINANCIAL"
	RequestCategoryAdmin        = "ADMIN" // Internal — ADMIN_VOID and REOPEN
)

// FinancialRequestTypes are requests that require an exclusive financial lock (BR-PM-030).
var FinancialRequestTypes = map[string]bool{
	RequestTypeSurrender:       true,
	RequestTypeLoan:            true,
	RequestTypeLoanRepayment:   false, // No lock — special case (BR-PM-020)
	RequestTypeRevival:         true,
	RequestTypeDeathClaim:      false, // Preempts, no lock (BR-PM-031)
	RequestTypeMaturityClaim:   true,
	RequestTypeSurvivalBenefit: true,
	RequestTypeCommutation:     true,
	RequestTypeConversion:      true,
	RequestTypeFLC:             true,
	RequestTypePaidUp:          true,
	RequestTypeForcedSurrender: true,
}

// ============================================================================
// Request Status Constants
// Source: DDL: request_status enum, Swagger: RequestStatus
// ============================================================================

const (
	RequestStatusReceived         = "RECEIVED"
	RequestStatusStateGateRejected = "STATE_GATE_REJECTED"
	RequestStatusRouted           = "ROUTED"
	RequestStatusInProgress       = "IN_PROGRESS"
	RequestStatusCompleted        = "COMPLETED"
	RequestStatusCancelled        = "CANCELLED"
	RequestStatusWithdrawn        = "WITHDRAWN"
	RequestStatusTimedOut         = "TIMED_OUT"
	RequestStatusAutoTerminated   = "AUTO_TERMINATED"
)

// ============================================================================
// Request Outcome Constants
// Source: DDL: request_outcome enum, Swagger: RequestOutcome
// ============================================================================

const (
	RequestOutcomeApproved       = "APPROVED"
	RequestOutcomeRejected       = "REJECTED"
	RequestOutcomeWithdrawn      = "WITHDRAWN"
	RequestOutcomeTimeout        = "TIMEOUT"
	RequestOutcomePreempted      = "PREEMPTED"
	RequestOutcomeDomainRejected = "DOMAIN_REJECTED"
)

// ============================================================================
// Source Channel Constants
// Source: DDL: source_channel enum, Swagger: SourceChannel
// ============================================================================

const (
	SourceChannelCustomerPortal = "CUSTOMER_PORTAL"
	SourceChannelCPC            = "CPC"
	SourceChannelMobileApp      = "MOBILE_APP"
	SourceChannelAgentPortal    = "AGENT_PORTAL"
	SourceChannelBatch          = "BATCH"
	SourceChannelSystem         = "SYSTEM"
)

// ============================================================================
// ServiceRequest — Central Request Registry
// Source: §8.3, DDL: policy_mgmt.service_request
// Scale: ~20K new requests/day; partitioned by submitted_at (quarterly)
// ⚠️ Partition key: submitted_at — MUST appear in all WHERE clauses
// ⚠️ PK is composite: (request_id, submitted_at)
// ⚠️ request_id is BIGINT from seq_service_request_id (NOT a UUID)
// ============================================================================

type ServiceRequest struct {
	// Primary key (BIGINT from seq_service_request_id)
	RequestID int64 `json:"request_id" db:"request_id"`

	// Policy reference
	PolicyID     int64  `json:"policy_id"     db:"policy_id"`
	PolicyNumber string `json:"policy_number" db:"policy_number"`

	// Request classification
	RequestType     string `json:"request_type"     db:"request_type"`     // request_type enum
	RequestCategory string `json:"request_category" db:"request_category"` // request_category enum
	Status          string `json:"status"           db:"status"`           // request_status enum
	SourceChannel   string `json:"source_channel"   db:"source_channel"`   // source_channel enum

	// Submission
	SubmittedBy *int64    `json:"submitted_by,omitempty" db:"submitted_by"`
	SubmittedAt time.Time `json:"submitted_at"           db:"submitted_at"` // ⚠️ PARTITION KEY

	// State gate
	StateGateStatus *string `json:"state_gate_status,omitempty" db:"state_gate_status"` // lifecycle_status at check time

	// Routing
	RoutedAt             *time.Time `json:"routed_at,omitempty"              db:"routed_at"`
	DownstreamService    *string    `json:"downstream_service,omitempty"     db:"downstream_service"`
	DownstreamWorkflowID *string    `json:"downstream_workflow_id,omitempty" db:"downstream_workflow_id"`
	DownstreamTaskQueue  *string    `json:"downstream_task_queue,omitempty"  db:"downstream_task_queue"`

	// Completion
	CompletedAt    *time.Time       `json:"completed_at,omitempty"    db:"completed_at"`
	Outcome        *string          `json:"outcome,omitempty"         db:"outcome"`         // request_outcome enum
	OutcomeReason  *string          `json:"outcome_reason,omitempty"  db:"outcome_reason"`
	OutcomePayload json.RawMessage  `json:"outcome_payload,omitempty" db:"outcome_payload"` // JSONB

	// Request data
	RequestPayload json.RawMessage `json:"request_payload" db:"request_payload"` // JSONB — source request body

	// Lifecycle management
	TimeoutAt      *time.Time `json:"timeout_at,omitempty"      db:"timeout_at"`
	IdempotencyKey *string    `json:"idempotency_key,omitempty" db:"idempotency_key"` // X-Idempotency-Key header

	// Audit
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// DownstreamTaskQueueForType returns the Temporal task queue for a given request type.
// Source: Constraint 1 table (FR-PM-009 / §9.1)
func DownstreamTaskQueueForType(requestType string) string {
	switch requestType {
	case RequestTypeSurrender, RequestTypeForcedSurrender:
		return "surrender-tq"
	case RequestTypeLoan, RequestTypeLoanRepayment:
		return "loan-tq"
	case RequestTypeRevival:
		return "revival-tq"
	case RequestTypeDeathClaim, RequestTypeMaturityClaim, RequestTypeSurvivalBenefit:
		return "claims-tq"
	case RequestTypeCommutation:
		return "commutation-tq"
	case RequestTypeConversion:
		return "conversion-tq"
	case RequestTypeFLC:
		return "freelook-tq"
	case RequestTypeNominationChange, RequestTypeBillingMethodChange,
		RequestTypeAssignment, RequestTypeAddressChange, RequestTypeDuplicateBond:
		return "nfs-tq"
	case RequestTypePremiumRefund:
		return "billing-tq"
	default:
		return ""
	}
}
