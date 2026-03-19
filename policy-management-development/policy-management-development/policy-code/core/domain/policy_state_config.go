package domain

import (
	"strconv"
	"time"
)

// ============================================================================
// PolicyStateConfig — Configurable Parameters for PM Operations
// Source: §8.7, DDL: policy_mgmt.policy_state_config
// Scale: ~43 rows after migrations 001 + 002
// PK: config_key
// ⚠️ data_type constraint: STRING | INTEGER | DURATION | BOOLEAN | DECIMAL
//    JSON is NOT a valid data_type — store JSON as STRING and parse in app layer.
// ============================================================================

// PolicyStateConfig stores a single configurable key-value parameter.
type PolicyStateConfig struct {
	ConfigKey   string     `json:"config_key"            db:"config_key"`
	ConfigValue string     `json:"config_value"          db:"config_value"`
	Description *string    `json:"description,omitempty" db:"description"`
	DataType    string     `json:"data_type"             db:"data_type"` // STRING|INTEGER|DURATION|BOOLEAN|DECIMAL
	UpdatedAt   time.Time  `json:"updated_at"            db:"updated_at"`
	UpdatedBy   *int64     `json:"updated_by,omitempty"  db:"updated_by"`
}

// ============================================================================
// Well-known config key constants — matches seeds in migrations 001 + 002
// ============================================================================

const (
	ConfigKeyGracePeriodRule        = "grace_period_rule"
	ConfigKeyPaidUpMinInforceMonths = "paid_up_min_inforce_months"
	ConfigKeyPaidUpMinLapseMonths   = "paid_up_min_lapse_months"
	ConfigKeyPaidUpMinValue         = "paid_up_min_value"

	ConfigKeyMaturityNotificationDays  = "maturity_notification_days"
	ConfigKeyForcedSurrenderWindowDays = "forced_surrender_window_days"

	ConfigKeyRevivalFirstInstallmentSLADays = "revival_first_installment_sla_days"
	ConfigKeyFLCPeriodDays                  = "flc_period_days"
	ConfigKeyFLCPeriodDistanceMarketing     = "flc_period_distance_marketing_days"

	ConfigKeySignalDedupTTLDays       = "signal_dedup_ttl_days"
	ConfigKeyCANMaxEvents             = "continue_as_new_max_events"
	ConfigKeyCANMaxDays               = "continue_as_new_max_days"
	ConfigKeyCANMaxHistoryMB          = "continue_as_new_max_history_mb"

	ConfigKeyRoutingTimeoutSurrender       = "routing_timeout_surrender"
	ConfigKeyRoutingTimeoutForcedSurrender = "routing_timeout_forced_surrender"
	ConfigKeyRoutingTimeoutLoan            = "routing_timeout_loan"
	ConfigKeyRoutingTimeoutLoanRepayment   = "routing_timeout_loan_repayment"
	ConfigKeyRoutingTimeoutRevival         = "routing_timeout_revival"
	ConfigKeyRoutingTimeoutDeathClaim      = "routing_timeout_death_claim"
	ConfigKeyRoutingTimeoutMaturityClaim   = "routing_timeout_maturity_claim"
	ConfigKeyRoutingTimeoutSurvivalBenefit = "routing_timeout_survival_benefit"
	ConfigKeyRoutingTimeoutCommutation     = "routing_timeout_commutation"
	ConfigKeyRoutingTimeoutConversion      = "routing_timeout_conversion"
	ConfigKeyRoutingTimeoutFLC             = "routing_timeout_flc"
	ConfigKeyRoutingTimeoutNFR             = "routing_timeout_nfr"
	ConfigKeyRoutingTimeoutPremiumRefund   = "routing_timeout_premium_refund"

	ConfigKeyCoolingPeriodVoid               = "cooling_period_void"
	ConfigKeyCoolingPeriodSurrendered        = "cooling_period_surrendered"
	ConfigKeyCoolingPeriodTerminatedSurrender = "cooling_period_terminated_surrender"
	ConfigKeyCoolingPeriodMatured            = "cooling_period_matured"
	ConfigKeyCoolingPeriodDeathClaimSettled  = "cooling_period_death_claim_settled"
	ConfigKeyCoolingPeriodFLCCancelled       = "cooling_period_flc_cancelled"
	ConfigKeyCoolingPeriodCancelledDeath     = "cooling_period_cancelled_death"
	ConfigKeyCoolingPeriodConverted          = "cooling_period_converted"

	ConfigKeyRemissionMinPolicyLifeMonths = "remission_min_policy_life_months"
	ConfigKeyLapseAgeThresholdMonths      = "lapse_age_threshold_months"
	ConfigKeyRemissionSlab6MODays         = "remission_slab_6mo_days"
	ConfigKeyRemissionSlab12MODays        = "remission_slab_12mo_days"
	ConfigKeyRemissionSlab24MODays        = "remission_slab_24mo_days"
	ConfigKeyRemissionSlab36MOMonths      = "remission_slab_36mo_months"

	ConfigKeyPayRecoveryProtectionMonths   = "pay_recovery_protection_months"
	ConfigKeyForcedSurrenderReminderCount  = "forced_surrender_reminder_count"
	ConfigKeyForcedSurrenderLoanRatioPct   = "forced_surrender_loan_ratio_threshold_pct"
	ConfigKeyWLAMaturityAgeYears           = "wla_maturity_age_years"
	ConfigKeyDashboardMVRefreshInterval    = "dashboard_mv_refresh_interval_minutes"

	ConfigKeyProductCatalogPLI  = "product_catalog.PLI"
	ConfigKeyProductCatalogRPLI = "product_catalog.RPLI"
)

// AsInt converts the config value to int. Returns 0 if conversion fails.
func (c PolicyStateConfig) AsInt() int {
	v, _ := strconv.Atoi(c.ConfigValue)
	return v
}

// AsFloat converts the config value to float64. Returns 0 if conversion fails.
func (c PolicyStateConfig) AsFloat() float64 {
	v, _ := strconv.ParseFloat(c.ConfigValue, 64)
	return v
}
