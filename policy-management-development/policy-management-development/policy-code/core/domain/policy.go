package domain

import "time"

// ============================================================================
// Lifecycle Status Constants (23 canonical states — v4.1)
// Source: §8.1, DDL: lifecycle_status enum
// ============================================================================

const (
	StatusFreeLookActive          = "FREE_LOOK_ACTIVE"
	StatusActive                  = "ACTIVE"
	StatusVoidLapse               = "VOID_LAPSE"
	StatusInactiveLapse           = "INACTIVE_LAPSE"
	StatusActiveLapse             = "ACTIVE_LAPSE"
	StatusPaidUp                  = "PAID_UP"
	StatusReducedPaidUp           = "REDUCED_PAID_UP"
	StatusAssignedToPresident     = "ASSIGNED_TO_PRESIDENT"
	StatusPendingAutoSurrender    = "PENDING_AUTO_SURRENDER"
	StatusPendingSurrender        = "PENDING_SURRENDER"
	StatusRevivalPending          = "REVIVAL_PENDING"
	StatusPendingMaturity         = "PENDING_MATURITY"
	StatusDeathClaimIntimated     = "DEATH_CLAIM_INTIMATED"
	StatusDeathUnderInvestigation = "DEATH_UNDER_INVESTIGATION"
	StatusSuspended               = "SUSPENDED"
	StatusVoid                    = "VOID"
	StatusSurrendered             = "SURRENDERED"
	StatusTerminatedSurrender     = "TERMINATED_SURRENDER"
	StatusMatured                 = "MATURED"
	StatusDeathClaimSettled       = "DEATH_CLAIM_SETTLED"
	StatusFLCCancelled            = "FLC_CANCELLED"
	StatusCancelledDeath          = "CANCELLED_DEATH"
	StatusConverted               = "CONVERTED"
)

// Terminal states — policy workflow enters cooling period then ends.
var TerminalStatuses = map[string]bool{
	StatusVoid:                true,
	StatusSurrendered:         true,
	StatusTerminatedSurrender: true,
	StatusMatured:             true,
	StatusDeathClaimSettled:   true,
	StatusFLCCancelled:        true,
	StatusCancelledDeath:      true,
	StatusConverted:           true,
}

// ============================================================================
// Product Type Constants
// Source: DDL: product_type enum
// ============================================================================

const (
	ProductTypePLI  = "PLI"
	ProductTypeRPLI = "RPLI"
)

// ============================================================================
// Premium Mode Constants
// Source: DDL: premium_mode enum
// ============================================================================

const (
	PremiumModeMonthly    = "MONTHLY"
	PremiumModeQuarterly  = "QUARTERLY"
	PremiumModeHalfYearly = "HALF_YEARLY"
	PremiumModeYearly     = "YEARLY"
)

// ============================================================================
// Billing Method Constants
// Source: DDL: billing_method enum
// ⚠️ NOTE: DDO is NOT in the DDL — only CASH, PAY_RECOVERY, ONLINE
// ============================================================================

const (
	BillingMethodCash        = "CASH"
	BillingMethodPayRecovery = "PAY_RECOVERY"
	BillingMethodOnline      = "ONLINE"
)

// ============================================================================
// Assignment Type Constants
// Source: DDL: assignment_type_enum
// ============================================================================

const (
	AssignmentTypeNone        = "NONE"
	AssignmentTypeAbsolute    = "ABSOLUTE"
	AssignmentTypeConditional = "CONDITIONAL"
)

// ============================================================================
// Paid-Up Type Constants
// Source: DDL: paid_up_type_enum
// ============================================================================

const (
	PaidUpTypeAuto      = "AUTO"
	PaidUpTypeVoluntary = "VOLUNTARY"
	PaidUpTypeReduced   = "REDUCED"
)

// ============================================================================
// Policy — Core Policy State Table
// Source: §8.1, DDL: policy_mgmt.policy
// Scale: 3M active / 50M total
// ============================================================================

// Policy is the core lifecycle state entity. PM is the sole writer.
// All DB column names (db:"") match DDL exactly.
type Policy struct {
	// Primary key (BIGINT from seq_policy_id)
	PolicyID int64 `json:"policy_id" db:"policy_id"`

	// Policy identity
	PolicyNumber string `json:"policy_number" db:"policy_number"`
	CustomerID   int64  `json:"customer_id"   db:"customer_id"`
	ProductCode  string `json:"product_code"  db:"product_code"`
	ProductType  string `json:"product_type"  db:"product_type"` // PLI | RPLI

	// Lifecycle state (PM-owned)
	CurrentStatus                  string     `json:"current_status"                    db:"current_status"`
	PreviousStatus                 *string    `json:"previous_status,omitempty"         db:"previous_status"`
	PreviousStatusBeforeSuspension *string    `json:"previous_status_before_suspension" db:"previous_status_before_suspension"`
	EffectiveFrom                  time.Time  `json:"effective_from"                    db:"effective_from"`

	// Financial data
	SumAssured     float64 `json:"sum_assured"      db:"sum_assured"`
	CurrentPremium float64 `json:"current_premium"  db:"current_premium"`
	PremiumMode    string  `json:"premium_mode"     db:"premium_mode"`    // MONTHLY|QUARTERLY|...
	BillingMethod  string  `json:"billing_method"   db:"billing_method"`  // CASH|PAY_RECOVERY|ONLINE

	// Key dates
	IssueDate           time.Time  `json:"issue_date"             db:"issue_date"`
	PolicyInceptionDate time.Time  `json:"policy_inception_date"  db:"policy_inception_date"`
	MaturityDate        *time.Time `json:"maturity_date,omitempty" db:"maturity_date"`       // NULL for WLA until age 80
	PaidToDate          time.Time  `json:"paid_to_date"           db:"paid_to_date"`
	NextPremiumDueDate  *time.Time `json:"next_premium_due_date,omitempty" db:"next_premium_due_date"`

	// Agent
	AgentID *int64 `json:"agent_id,omitempty" db:"agent_id"`

	// Encumbrance flags (Tier 2 hybrid state model)
	HasActiveLoan      bool    `json:"has_active_loan"     db:"has_active_loan"`
	LoanOutstanding    float64 `json:"loan_outstanding"    db:"loan_outstanding"`
	AssignmentType     string  `json:"assignment_type"     db:"assignment_type"`    // NONE|ABSOLUTE|CONDITIONAL
	AssignmentStatus   string  `json:"assignment_status"   db:"assignment_status"`
	AMLHold            bool    `json:"aml_hold"            db:"aml_hold"`
	DisputeFlag        bool    `json:"dispute_flag"        db:"dispute_flag"`
	MurderClauseActive *bool   `json:"murder_clause_active,omitempty" db:"murder_clause_active"`

	// Display status (computed by DB trigger — never set explicitly in code)
	DisplayStatus string `json:"display_status" db:"display_status"`

	// Lapsation fields (BR-PM-070, BR-PM-074)
	FirstUnpaidPremiumDate       *time.Time `json:"first_unpaid_premium_date,omitempty"       db:"first_unpaid_premium_date"`
	RemissionExpiryDate          *time.Time `json:"remission_expiry_date,omitempty"           db:"remission_expiry_date"`
	PayRecoveryProtectionExpiry  *time.Time `json:"pay_recovery_protection_expiry,omitempty"  db:"pay_recovery_protection_expiry"` // BR-PM-074

	// Paid-up fields (BR-PM-060 to BR-PM-065)
	PaidUpValue *float64   `json:"paid_up_value,omitempty" db:"paid_up_value"`
	PaidUpType  *string    `json:"paid_up_type,omitempty"  db:"paid_up_type"`  // AUTO|VOLUNTARY|REDUCED
	PaidUpDate  *time.Time `json:"paid_up_date,omitempty"  db:"paid_up_date"`

	// Survival benefit tracking
	SBInstallmentsPaid int     `json:"sb_installments_paid" db:"sb_installments_paid"`
	SBTotalAmountPaid  float64 `json:"sb_total_amount_paid" db:"sb_total_amount_paid"`

	// Nomination
	NominationStatus string `json:"nomination_status" db:"nomination_status"` // ABSENT|PRESENT|...

	// WLA-specific (GAP-PM-002: WLA matures at age 80)
	PolicyholderDOB time.Time `json:"policyholder_dob" db:"policyholder_dob"`

	// Temporal workflow references
	WorkflowID    string  `json:"workflow_id"              db:"workflow_id"`    // plw-{policy_number}
	TemporalRunID *string `json:"temporal_run_id,omitempty" db:"temporal_run_id"`

	// Optimistic locking (incremented on every UPDATE)
	Version int64 `json:"version" db:"version"`

	// Audit
	CreatedAt time.Time  `json:"created_at"        db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"        db:"updated_at"`
	CreatedBy *int64     `json:"created_by,omitempty" db:"created_by"`
	UpdatedBy *int64     `json:"updated_by,omitempty" db:"updated_by"`
}
