package workflows

// ============================================================================
// PolicyLifecycleWorkflow — per-policy long-running workflow
//
// Trigger:  Temporal SignalWithStart from Policy Issue Service (A10.1)
// Workflow ID: plw-{policy_number}  e.g. plw-PLI/2026/000001
// Namespace:   pli-insurance
// Task Queue:  policy-management-tq
//
// [FR-PM-001, FR-PM-002, Constraint 2, §9.1, §9.5.1, A10.1]
// ============================================================================

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	enumspb "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"policy-management/core/domain"
	acts "policy-management/workflows/activities"
)

// ─────────────────────────────────────────────────────────────────────────────
// Continue-As-New thresholds (EXACT from A23.1 / FR-PM-002)
// ─────────────────────────────────────────────────────────────────────────────

const (
	canEventThreshold   = 40000            // NOT 50000 [A23.1, G2]
	canHistorySizeBytes = 50 * 1024 * 1024 // 50 MB
	canTimeThreshold    = 30 * 24 * time.Hour
)

// policyActs is a zero-value instance of PolicyActivities used ONLY for activity
// function references in workflow.ExecuteActivity calls. Temporal uses reflection
// to extract the registered activity name from the method value — it never calls
// methods on this zero-value instance at runtime (the registered worker instance
// is used). This is the standard Temporal Go SDK pattern for struct activities. [FR-PM-001]
var policyActs acts.PolicyActivities

// ─────────────────────────────────────────────────────────────────────────────
// FLC period helper (Constraint 10)
// ─────────────────────────────────────────────────────────────────────────────

// routingTimeoutForRequest returns the per-type routing timeout duration.
// Values are based on domain/policy_state_config.go seed values. For fully
// config-driven timeouts, call FetchWorkflowConfigActivity with the appropriate
// ConfigKeyRoutingTimeout* key. [Review-Fix-16, domain.ConfigKeyRoutingTimeout*]
func routingTimeoutForRequest(requestType string) time.Duration {
	switch requestType {
	case domain.RequestTypeSurrender, domain.RequestTypeForcedSurrender,
		domain.RequestTypeConversion:
		return 7 * 24 * time.Hour
	case domain.RequestTypeLoan, domain.RequestTypeLoanRepayment:
		return 3 * 24 * time.Hour
	case domain.RequestTypeDeathClaim:
		return 90 * 24 * time.Hour
	case domain.RequestTypeRevival, domain.RequestTypeMaturityClaim,
		domain.RequestTypeSurvivalBenefit, domain.RequestTypeCommutation,
		domain.RequestTypeFLC:
		return 30 * 24 * time.Hour
	default:
		return 15 * 24 * time.Hour // NFR and all others
	}
}

// getFLCPeriod converts a config-supplied FLC days value to a Duration.
// If flcDays <= 0, defaults to 15 days (standard policy; 30 days for distance-marketing
// products — caller should pass the correct value from FetchWorkflowConfigActivity).
// [Review-Fix-8, §10.1.6, ConfigKeyFLCPeriodDays]
func getFLCPeriod(flcDays int) time.Duration {
	if flcDays <= 0 {
		flcDays = 15 // safe default [§10.1.6]
	}
	return time.Duration(flcDays) * 24 * time.Hour
}

// ─────────────────────────────────────────────────────────────────────────────
// Terminal cooling durations (Constraint 6, §9.5.1)
// ─────────────────────────────────────────────────────────────────────────────

func coolingDuration(terminalStatus string) time.Duration {
	switch terminalStatus {
	case domain.StatusDeathClaimSettled:
		return 180 * 24 * time.Hour // 180 days [§9.5.1]
	case domain.StatusMatured, domain.StatusConverted, domain.StatusSurrendered,
		domain.StatusTerminatedSurrender:
		return 90 * 24 * time.Hour // 90 days
	case domain.StatusFLCCancelled, domain.StatusCancelledDeath:
		return 30 * 24 * time.Hour // 30 days
	case domain.StatusVoid:
		return 60 * 24 * time.Hour // 60 days [Review-Fix-14, §9.5.1]
	default:
		return 90 * 24 * time.Hour // safe default
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Config-backed timeout/cooling helpers [B4, Review-Fix-16]
// ─────────────────────────────────────────────────────────────────────────────

// routingTimeoutFromConfig returns the routing timeout for a request type,
// reading from state.CachedConfig (populated at startup) with a hardcoded
// fallback so the workflow is never blocked by a missing DB config row. [B4]
func routingTimeoutFromConfig(cachedConfig map[string]string, requestType string) time.Duration {
	key := routingTimeoutConfigKeyForType(requestType)
	if v, ok := cachedConfig[key]; ok {
		if d, err := time.ParseDuration(v); err == nil && d > 0 {
			return d
		}
	}
	return routingTimeoutForRequest(requestType) // hardcoded fallback
}

// coolingDurationFromConfig returns the terminal cooling period for a status,
// reading from state.CachedConfig with a hardcoded fallback. [B4]
func coolingDurationFromConfig(cachedConfig map[string]string, terminalStatus string) time.Duration {
	key := coolingConfigKeyForStatus(terminalStatus)
	if key != "" {
		if v, ok := cachedConfig[key]; ok {
			if d, err := time.ParseDuration(v); err == nil && d > 0 {
				return d
			}
		}
	}
	return coolingDuration(terminalStatus) // hardcoded fallback
}

// routingTimeoutConfigKeyForType maps request type → ConfigKeyRoutingTimeout* constant.
func routingTimeoutConfigKeyForType(requestType string) string {
	switch requestType {
	case domain.RequestTypeSurrender:
		return domain.ConfigKeyRoutingTimeoutSurrender
	case domain.RequestTypeForcedSurrender:
		return domain.ConfigKeyRoutingTimeoutForcedSurrender
	case domain.RequestTypeLoan:
		return domain.ConfigKeyRoutingTimeoutLoan
	case domain.RequestTypeLoanRepayment:
		return domain.ConfigKeyRoutingTimeoutLoanRepayment
	case domain.RequestTypeRevival:
		return domain.ConfigKeyRoutingTimeoutRevival
	case domain.RequestTypeDeathClaim:
		return domain.ConfigKeyRoutingTimeoutDeathClaim
	case domain.RequestTypeMaturityClaim:
		return domain.ConfigKeyRoutingTimeoutMaturityClaim
	case domain.RequestTypeSurvivalBenefit:
		return domain.ConfigKeyRoutingTimeoutSurvivalBenefit
	case domain.RequestTypeCommutation:
		return domain.ConfigKeyRoutingTimeoutCommutation
	case domain.RequestTypeConversion:
		return domain.ConfigKeyRoutingTimeoutConversion
	case domain.RequestTypeFLC:
		return domain.ConfigKeyRoutingTimeoutFLC
	case domain.RequestTypePremiumRefund:
		return domain.ConfigKeyRoutingTimeoutPremiumRefund
	default:
		return domain.ConfigKeyRoutingTimeoutNFR
	}
}

// coolingConfigKeyForStatus maps terminal status → ConfigKeyCoolingPeriod* constant.
func coolingConfigKeyForStatus(terminalStatus string) string {
	switch terminalStatus {
	case domain.StatusVoid:
		return domain.ConfigKeyCoolingPeriodVoid
	case domain.StatusSurrendered:
		return domain.ConfigKeyCoolingPeriodSurrendered
	case domain.StatusTerminatedSurrender:
		return domain.ConfigKeyCoolingPeriodTerminatedSurrender
	case domain.StatusMatured:
		return domain.ConfigKeyCoolingPeriodMatured
	case domain.StatusDeathClaimSettled:
		return domain.ConfigKeyCoolingPeriodDeathClaimSettled
	case domain.StatusFLCCancelled:
		return domain.ConfigKeyCoolingPeriodFLCCancelled
	case domain.StatusCancelledDeath:
		return domain.ConfigKeyCoolingPeriodCancelledDeath
	case domain.StatusConverted:
		return domain.ConfigKeyCoolingPeriodConverted
	default:
		return ""
	}
}

// requestTypeToSignalName maps domain request types to the kebab-case Temporal signal name.
// Used to produce audit log entries consistent with Temporal signal channel names. [B9]
func requestTypeToSignalName(requestType string) string {
	switch requestType {
	case domain.RequestTypeSurrender:
		return SignalSurrenderRequest
	case domain.RequestTypeLoan:
		return SignalLoanRequest
	case domain.RequestTypeLoanRepayment:
		return SignalLoanRepayment
	case domain.RequestTypeRevival:
		return SignalRevivalRequest
	case domain.RequestTypeDeathClaim:
		return SignalDeathNotification
	case domain.RequestTypeMaturityClaim:
		return SignalMaturityClaimRequest
	case domain.RequestTypeSurvivalBenefit:
		return SignalSurvivalBenefitRequest
	case domain.RequestTypeCommutation:
		return SignalCommutationRequest
	case domain.RequestTypeConversion:
		return SignalConversionRequest
	case domain.RequestTypeFLC:
		return SignalFLCRequest
	case domain.RequestTypeForcedSurrender:
		return SignalForcedSurrenderTrigger
	case domain.RequestTypePaidUp:
		// Voluntary paid-up uses a dedicated signal channel — NOT the forced-surrender trigger. [D10]
		return SignalVoluntaryPaidUpRequest
	default:
		return SignalNFRRequest
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Activity option helpers
// ─────────────────────────────────────────────────────────────────────────────

// shortActCtx wraps ctx with standard short-activity options (DB calls, events).
func shortActCtx(ctx workflow.Context) workflow.Context {
	return workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts:    3,
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    10 * time.Second,
		},
	})
}

// childWFCtx wraps ctx with child workflow options for downstream dispatch.
// ParentClosePolicy is Abandon so child workflows continue independently
// when PLW does ContinueAsNew (default TERMINATE would kill them). [B2]
func childWFCtx(ctx workflow.Context, taskQueue, childID string) workflow.Context {
	return workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
		TaskQueue:         taskQueue,
		WorkflowID:        childID,
		ParentClosePolicy: enumspb.PARENT_CLOSE_POLICY_ABANDON,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 1, // No automatic retry for business workflows
		},
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// shouldContinueAsNew — checks CAN thresholds [FR-PM-002, A23.1]
// ─────────────────────────────────────────────────────────────────────────────

func shouldContinueAsNew(ctx workflow.Context, state PolicyLifecycleState) bool {
	if state.EventCount >= canEventThreshold {
		return true
	}
	info := workflow.GetInfo(ctx)
	// Also check actual Temporal history event count — may diverge from EventCount
	// after non-deterministic retries or timer firings. [Review-Fix-4]
	if int(info.GetCurrentHistoryLength()) >= canEventThreshold {
		return true
	}
	if info.GetCurrentHistorySize() >= canHistorySizeBytes {
		return true
	}
	if workflow.Now(ctx).Sub(state.LastCANTime) >= canTimeThreshold {
		return true
	}
	return false
}

// ─────────────────────────────────────────────────────────────────────────────
// pruneProcessedSignals — evicts dedup entries older than 90 days
// ─────────────────────────────────────────────────────────────────────────────

func pruneProcessedSignals(ctx workflow.Context, state *PolicyLifecycleState) {
	// Use workflow.Now(ctx) — deterministic Temporal clock — instead of time.Now().
	// time.Now() inside workflow code causes replay divergence (ErrWorkflowResultMismatch)
	// because different entries get pruned on replay vs original execution. [B1]
	cutoff := workflow.Now(ctx).Add(-90 * 24 * time.Hour)
	for k, t := range state.ProcessedSignalIDs {
		if t.Before(cutoff) {
			delete(state.ProcessedSignalIDs, k)
		}
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// isStateEligible — pure in-memory state gate check [§9.1, BR-PM-011..BR-PM-023]
// ─────────────────────────────────────────────────────────────────────────────

func isStateEligible(requestType, status string, enc EncumbranceFlags) (bool, string) {
	// SUSPENDED blocks all requests except death-notification and aml-flag-cleared [BR-PM-110]
	if status == domain.StatusSuspended && requestType != domain.RequestTypeDeathClaim {
		return false, "policy is suspended — only death claims accepted [BR-PM-110]"
	}
	switch requestType {
	case domain.RequestTypeSurrender:
		eligible := status == domain.StatusActive || status == domain.StatusVoidLapse ||
			status == domain.StatusInactiveLapse || status == domain.StatusActiveLapse ||
			status == domain.StatusPaidUp
		if !eligible {
			return false, fmt.Sprintf("surrender not allowed in status %s [BR-PM-011]", status)
		}
	case domain.RequestTypeLoan:
		if status != domain.StatusActive {
			return false, fmt.Sprintf("loan not allowed in status %s [BR-PM-012]", status)
		}
		if enc.HasActiveLoan {
			return false, "active loan already exists [BR-PM-012]"
		}
	case domain.RequestTypeLoanRepayment:
		eligible := status == domain.StatusActive || status == domain.StatusAssignedToPresident ||
			status == domain.StatusPendingAutoSurrender
		if !eligible {
			return false, fmt.Sprintf("loan repayment not allowed in status %s [BR-PM-013]", status)
		}
		if !enc.HasActiveLoan {
			return false, "no active loan to repay [BR-PM-013]"
		}
	case domain.RequestTypeRevival:
		eligible := status == domain.StatusVoidLapse || status == domain.StatusInactiveLapse ||
			status == domain.StatusActiveLapse
		if !eligible {
			return false, fmt.Sprintf("revival not allowed in status %s [BR-PM-014]", status)
		}
	case domain.RequestTypeDeathClaim:
		// Accepted in ALL non-terminal states including SUSPENDED [BR-PM-014, BR-PM-112]
		if domain.TerminalStatuses[status] {
			return false, fmt.Sprintf("death claim not allowed in terminal status %s", status)
		}
	case domain.RequestTypeMaturityClaim:
		eligible := status == domain.StatusActive || status == domain.StatusPendingMaturity
		if !eligible {
			return false, fmt.Sprintf("maturity claim not allowed in status %s [BR-PM-015]", status)
		}
	case domain.RequestTypeSurvivalBenefit:
		if status != domain.StatusActive {
			return false, fmt.Sprintf("survival benefit not allowed in status %s [BR-PM-016]", status)
		}
	case domain.RequestTypeCommutation:
		if status != domain.StatusActive {
			return false, fmt.Sprintf("commutation not allowed in status %s [BR-PM-017]", status)
		}
	case domain.RequestTypeConversion:
		if status != domain.StatusActive {
			return false, fmt.Sprintf("conversion not allowed in status %s [BR-PM-018]", status)
		}
	case domain.RequestTypeFLC:
		if status != domain.StatusFreeLookActive {
			return false, fmt.Sprintf("FLC only allowed in FREE_LOOK_ACTIVE status [BR-PM-019]")
		}
	case domain.RequestTypePaidUp:
		eligible := status == domain.StatusActive || status == domain.StatusActiveLapse
		if !eligible {
			return false, fmt.Sprintf("voluntary paid-up not allowed in status %s [BR-PM-060]", status)
		}
	case domain.RequestTypeForcedSurrender:
		eligible := status == domain.StatusAssignedToPresident || status == domain.StatusPendingAutoSurrender
		if !eligible {
			return false, fmt.Sprintf("forced surrender not allowed in status %s", status)
		}
	default:
		// NFR: all non-terminal states (excludes SUSPENDED unless cleared) [BR-PM-023]
		if domain.TerminalStatuses[status] {
			return false, fmt.Sprintf("NFR not allowed in terminal status %s [BR-PM-023]", status)
		}
	}
	return true, ""
}

// requiresFinancialLock returns true if the request type needs an exclusive lock. [BR-PM-030]
func requiresFinancialLock(requestType string) bool {
	// LOAN_REPAYMENT and DEATH_CLAIM are exceptions — no lock required
	switch requestType {
	case domain.RequestTypeLoanRepayment, domain.RequestTypeDeathClaim:
		return false
	}
	_, isFinancial := domain.FinancialRequestTypes[requestType]
	return isFinancial
}

// preRouteStatus returns the pre-route status for request types that require one.
func preRouteStatus(requestType string) string {
	switch requestType {
	case domain.RequestTypeSurrender:
		return domain.StatusPendingSurrender
	case domain.RequestTypeRevival:
		return domain.StatusRevivalPending
	case domain.RequestTypeDeathClaim:
		return domain.StatusDeathClaimIntimated
	case domain.RequestTypeForcedSurrender:
		return domain.StatusPendingAutoSurrender
	}
	return ""
}

// ─────────────────────────────────────────────────────────────────────────────
// PolicyLifecycleWorkflow — main entry point
// [FR-PM-001, Constraint 2]
// ─────────────────────────────────────────────────────────────────────────────

// PolicyLifecycleWorkflow is the per-policy long-running workflow.
// Started by Policy Issue Service via SignalWithStart. Workflow ID: plw-{policyNumber}.
// Accepts the initial state as input; receives all subsequent lifecycle events as signals.
// [FR-PM-001, A10.1, Constraint 2, §9.1]
func PolicyLifecycleWorkflow(ctx workflow.Context, initialState PolicyLifecycleState) error {
	state := initialState
	CurrentStatusKey := temporal.NewSearchAttributeKeyKeyword("CurrentStatus")
	// Initialize maps/times on first run (not on CAN resume)
	if state.ProcessedSignalIDs == nil {
		state.ProcessedSignalIDs = make(map[string]time.Time)
	}
	if state.LastCANTime.IsZero() {
		state.LastCANTime = workflow.Now(ctx)
	}

	// ── Respawn FLC timer goroutine after Continue-As-New [D1] ───────────────
	// workflow.Go goroutines are lost when a workflow returns NewContinueAsNewError.
	// If FLCExpiryAt is set and still in the future, the policy was in FREE_LOOK_ACTIVE
	// when CAN fired. Respawn the goroutine using the remaining duration so the FLC
	// transition fires at the originally scheduled time, not delayed by the CAN round.
	if !state.FLCExpiryAt.IsZero() && workflow.Now(ctx).Before(state.FLCExpiryAt) {
		flcExpiryAt := state.FLCExpiryAt // capture for closure — avoid loop-variable aliasing
		workflow.Go(ctx, func(gCtx workflow.Context) {
			remaining := flcExpiryAt.Sub(workflow.Now(gCtx))
			if remaining > 0 {
				_ = workflow.Sleep(gCtx, remaining)
			}
			if state.CurrentStatus == domain.StatusFreeLookActive {
				doTransition(gCtx, &state, domain.StatusFreeLookActive, domain.StatusActive,
					"FLC period expired without cancellation (respawned after CAN)", "flc-timer", "")
			}
			state.FLCExpiryAt = time.Time{} // clear so subsequent CANs don't re-respawn
		})
	}

	// ── Register query handlers (7, EXACT names from §9.1) ──────────────────

	_ = workflow.SetQueryHandler(ctx, QueryGetPolicySummary, func() (*PolicyLifecycleState, error) {
		// [FR-PM-004] Returns full state for Tier-1 query (REST API two-tier pattern)
		return &state, nil
	})

	_ = workflow.SetQueryHandler(ctx, QueryGetPolicyStatus, func() (PolicyStatusQueryResult, error) {
		// [FR-PM-004]
		return PolicyStatusQueryResult{
			CurrentStatus:  state.CurrentStatus,
			PreviousStatus: state.PreviousStatus,
			DisplayStatus:  state.DisplayStatus,
			EffectiveFrom:  state.LastTransitionAt,
			Metadata:       state.Metadata,
		}, nil
	})

	_ = workflow.SetQueryHandler(ctx, QueryGetPendingRequests, func() ([]PendingRequest, error) {
		// [FR-PM-004]
		return state.PendingRequests, nil
	})

	_ = workflow.SetQueryHandler(ctx, QueryIsRequestEligible, func(requestType string) (IsRequestEligibleResult, error) {
		// Pure in-memory check — no DB call [FR-PM-004, §9.1]
		eligible, reason := isStateEligible(requestType, state.CurrentStatus, state.Encumbrances)
		lockConflict := false
		if eligible && requiresFinancialLock(requestType) && state.ActiveLock != nil {
			eligible = false
			lockConflict = true
			reason = fmt.Sprintf("financial lock held by %s [BR-PM-030]", state.ActiveLock.RequestType)
		}
		return IsRequestEligibleResult{Eligible: eligible, Reason: reason, LockConflict: lockConflict}, nil
	})

	_ = workflow.SetQueryHandler(ctx, QueryGetActiveLock, func() (*FinancialLock, error) {
		// [FR-PM-004]
		return state.ActiveLock, nil
	})

	_ = workflow.SetQueryHandler(ctx, QueryGetStatusHistory, func() (string, error) {
		// [FR-PM-004] History is in DB — redirect message per §9.1
		return "query policy_mgmt.policy_status_history via REST API [FR-PM-004]", nil
	})

	_ = workflow.SetQueryHandler(ctx, QueryGetWorkflowHealth, func() (WorkflowHealthResult, error) {
		// [FR-PM-004]
		return WorkflowHealthResult{
			EventCount:          state.EventCount,
			LastCANTime:         state.LastCANTime,
			PendingRequestCount: len(state.PendingRequests),
			HasActiveLock:       state.ActiveLock != nil,
		}, nil
	})

	// ── Signal channels ──────────────────────────────────────────────────────

	policyCreatedCh := workflow.GetSignalChannel(ctx, SignalPolicyCreated)
	surrenderReqCh := workflow.GetSignalChannel(ctx, SignalSurrenderRequest)
	loanReqCh := workflow.GetSignalChannel(ctx, SignalLoanRequest)
	loanRepaymentCh := workflow.GetSignalChannel(ctx, SignalLoanRepayment)
	revivalReqCh := workflow.GetSignalChannel(ctx, SignalRevivalRequest)
	deathNotifCh := workflow.GetSignalChannel(ctx, SignalDeathNotification)
	maturityClaimCh := workflow.GetSignalChannel(ctx, SignalMaturityClaimRequest)
	survivalBenefitCh := workflow.GetSignalChannel(ctx, SignalSurvivalBenefitRequest)
	commutationReqCh := workflow.GetSignalChannel(ctx, SignalCommutationRequest)
	conversionReqCh := workflow.GetSignalChannel(ctx, SignalConversionRequest)
	flcReqCh := workflow.GetSignalChannel(ctx, SignalFLCRequest)
	forcedSurrenderCh := workflow.GetSignalChannel(ctx, SignalForcedSurrenderTrigger)
	nfrReqCh := workflow.GetSignalChannel(ctx, SignalNFRRequest)
	surrenderCompletedCh := workflow.GetSignalChannel(ctx, SignalSurrenderCompleted)
	forcedSurCompletedCh := workflow.GetSignalChannel(ctx, SignalForcedSurrenderCompleted)
	loanCompletedCh := workflow.GetSignalChannel(ctx, SignalLoanCompleted)
	loanRepayCompletedCh := workflow.GetSignalChannel(ctx, SignalLoanRepaymentCompleted)
	revivalCompletedCh := workflow.GetSignalChannel(ctx, SignalRevivalCompleted)
	claimSettledCh := workflow.GetSignalChannel(ctx, SignalClaimSettled)
	commutationCompletedCh := workflow.GetSignalChannel(ctx, SignalCommutationCompleted)
	conversionCompletedCh := workflow.GetSignalChannel(ctx, SignalConversionCompleted)
	flcCompletedCh := workflow.GetSignalChannel(ctx, SignalFLCCompleted)
	nfrCompletedCh := workflow.GetSignalChannel(ctx, SignalNFRCompleted)
	opCompletedCh := workflow.GetSignalChannel(ctx, SignalOperationCompleted)
	premiumPaidCh := workflow.GetSignalChannel(ctx, SignalPremiumPaid)
	paymentDishonoredCh := workflow.GetSignalChannel(ctx, SignalPaymentDishonored)
	amlRaisedCh := workflow.GetSignalChannel(ctx, SignalAMLFlagRaised)
	amlClearedCh := workflow.GetSignalChannel(ctx, SignalAMLFlagCleared)
	investigationStartCh := workflow.GetSignalChannel(ctx, SignalInvestigationStarted)
	investigationConcludeCh := workflow.GetSignalChannel(ctx, SignalInvestigationConcluded)
	loanBalanceCh := workflow.GetSignalChannel(ctx, SignalLoanBalanceUpdated)
	conversionReversedCh := workflow.GetSignalChannel(ctx, SignalConversionReversed)
	adminVoidCh := workflow.GetSignalChannel(ctx, SignalAdminVoid)
	customerMergeCh := workflow.GetSignalChannel(ctx, SignalCustomerIDMerge)
	voluntaryPUCh := workflow.GetSignalChannel(ctx, SignalVoluntaryPaidUpRequest)
	withdrawalCh := workflow.GetSignalChannel(ctx, SignalWithdrawalRequest)
	disputeRegisteredCh := workflow.GetSignalChannel(ctx, SignalDisputeRegistered)
	disputeResolvedCh := workflow.GetSignalChannel(ctx, SignalDisputeResolved)
	batchSyncCh := workflow.GetSignalChannel(ctx, SignalBatchStateSync)

	// terminal flags controlled by signal handlers
	var reachedTerminal bool

	// ── Main event loop ──────────────────────────────────────────────────────

	for {
		// Check Continue-As-New thresholds [FR-PM-002, A23.1]
		if shouldContinueAsNew(ctx, state) {
			state.LastCANTime = workflow.Now(ctx)
			state.EventCount = 0 // [C1] reset so new run counts from zero; without this the new run
			// immediately re-triggers CAN (tight loop) since EventCount carries over at threshold.
			return workflow.NewContinueAsNewError(ctx, PolicyLifecycleWorkflow, state)
		}

		reachedTerminal = false
		sel := workflow.NewSelector(ctx)

		// ── Signal: policy-created (A10.1.6) ────────────────────────────────
		sel.AddReceive(policyCreatedCh, func(c workflow.ReceiveChannel, _ bool) {
			var sig PolicyCreatedSignal
			c.Receive(ctx, &sig)
			handlePolicyCreated(ctx, &state, sig)
		})

		// ── Signal: financial requests (surrender, loan, revival, etc.) ──────
		sel.AddReceive(surrenderReqCh, func(c workflow.ReceiveChannel, _ bool) {
			var sig PolicyRequestSignal
			c.Receive(ctx, &sig)
			sig.RequestType = domain.RequestTypeSurrender
			handleFinancialRequest(ctx, &state, sig)
		})
		sel.AddReceive(loanReqCh, func(c workflow.ReceiveChannel, _ bool) {
			var sig PolicyRequestSignal
			c.Receive(ctx, &sig)
			sig.RequestType = domain.RequestTypeLoan
			handleFinancialRequest(ctx, &state, sig)
		})
		sel.AddReceive(loanRepaymentCh, func(c workflow.ReceiveChannel, _ bool) {
			var sig PolicyRequestSignal
			c.Receive(ctx, &sig)
			sig.RequestType = domain.RequestTypeLoanRepayment
			handleFinancialRequest(ctx, &state, sig)
		})
		sel.AddReceive(revivalReqCh, func(c workflow.ReceiveChannel, _ bool) {
			var sig PolicyRequestSignal
			c.Receive(ctx, &sig)
			sig.RequestType = domain.RequestTypeRevival
			handleFinancialRequest(ctx, &state, sig)
		})
		sel.AddReceive(deathNotifCh, func(c workflow.ReceiveChannel, _ bool) {
			// HIGHEST PRIORITY: overrides SUSPENDED state [BR-PM-112]
			var sig PolicyRequestSignal
			c.Receive(ctx, &sig)
			sig.RequestType = domain.RequestTypeDeathClaim
			handleDeathNotification(ctx, &state, sig)
		})
		sel.AddReceive(maturityClaimCh, func(c workflow.ReceiveChannel, _ bool) {
			var sig PolicyRequestSignal
			c.Receive(ctx, &sig)
			sig.RequestType = domain.RequestTypeMaturityClaim
			handleFinancialRequest(ctx, &state, sig)
		})
		sel.AddReceive(survivalBenefitCh, func(c workflow.ReceiveChannel, _ bool) {
			var sig PolicyRequestSignal
			c.Receive(ctx, &sig)
			sig.RequestType = domain.RequestTypeSurvivalBenefit
			handleFinancialRequest(ctx, &state, sig)
		})
		sel.AddReceive(commutationReqCh, func(c workflow.ReceiveChannel, _ bool) {
			var sig PolicyRequestSignal
			c.Receive(ctx, &sig)
			sig.RequestType = domain.RequestTypeCommutation
			handleFinancialRequest(ctx, &state, sig)
		})
		sel.AddReceive(conversionReqCh, func(c workflow.ReceiveChannel, _ bool) {
			var sig PolicyRequestSignal
			c.Receive(ctx, &sig)
			sig.RequestType = domain.RequestTypeConversion
			handleFinancialRequest(ctx, &state, sig)
		})
		sel.AddReceive(flcReqCh, func(c workflow.ReceiveChannel, _ bool) {
			var sig PolicyRequestSignal
			c.Receive(ctx, &sig)
			sig.RequestType = domain.RequestTypeFLC
			handleFinancialRequest(ctx, &state, sig)
		})
		sel.AddReceive(forcedSurrenderCh, func(c workflow.ReceiveChannel, _ bool) {
			var sig PolicyRequestSignal
			c.Receive(ctx, &sig)
			sig.RequestType = domain.RequestTypeForcedSurrender
			handleFinancialRequest(ctx, &state, sig)
		})
		sel.AddReceive(nfrReqCh, func(c workflow.ReceiveChannel, _ bool) {
			var sig PolicyRequestSignal
			c.Receive(ctx, &sig)
			handleNFRRequest(ctx, &state, sig)
		})

		// ── Signal: completion signals ────────────────────────────────────────
		sel.AddReceive(surrenderCompletedCh, func(c workflow.ReceiveChannel, _ bool) {
			var sig OperationCompletedSignal
			c.Receive(ctx, &sig)
			sig.RequestType = domain.RequestTypeSurrender
			workflow.GetLogger(ctx).Info("Surrender completion signal received",
				"sig", sig,
			)
			reachedTerminal = handleOperationCompleted(ctx, &state, sig)
		})
		sel.AddReceive(forcedSurCompletedCh, func(c workflow.ReceiveChannel, _ bool) {
			var sig OperationCompletedSignal
			c.Receive(ctx, &sig)
			sig.RequestType = domain.RequestTypeForcedSurrender
			reachedTerminal = handleOperationCompleted(ctx, &state, sig)
		})
		sel.AddReceive(loanCompletedCh, func(c workflow.ReceiveChannel, _ bool) {
			var sig OperationCompletedSignal
			c.Receive(ctx, &sig)
			sig.RequestType = domain.RequestTypeLoan
			reachedTerminal = handleOperationCompleted(ctx, &state, sig)
		})
		sel.AddReceive(loanRepayCompletedCh, func(c workflow.ReceiveChannel, _ bool) {
			var sig OperationCompletedSignal
			c.Receive(ctx, &sig)
			sig.RequestType = domain.RequestTypeLoanRepayment
			reachedTerminal = handleOperationCompleted(ctx, &state, sig)
		})
		sel.AddReceive(revivalCompletedCh, func(c workflow.ReceiveChannel, _ bool) {
			var sig OperationCompletedSignal
			c.Receive(ctx, &sig)
			sig.RequestType = domain.RequestTypeRevival
			reachedTerminal = handleOperationCompleted(ctx, &state, sig)
		})
		sel.AddReceive(claimSettledCh, func(c workflow.ReceiveChannel, _ bool) {
			var sig OperationCompletedSignal
			c.Receive(ctx, &sig)
			reachedTerminal = handleClaimSettled(ctx, &state, sig)
		})
		sel.AddReceive(commutationCompletedCh, func(c workflow.ReceiveChannel, _ bool) {
			var sig OperationCompletedSignal
			c.Receive(ctx, &sig)
			sig.RequestType = domain.RequestTypeCommutation
			reachedTerminal = handleOperationCompleted(ctx, &state, sig)
		})
		sel.AddReceive(conversionCompletedCh, func(c workflow.ReceiveChannel, _ bool) {
			var sig OperationCompletedSignal
			c.Receive(ctx, &sig)
			sig.RequestType = domain.RequestTypeConversion
			reachedTerminal = handleOperationCompleted(ctx, &state, sig)
		})
		sel.AddReceive(flcCompletedCh, func(c workflow.ReceiveChannel, _ bool) {
			var sig OperationCompletedSignal
			c.Receive(ctx, &sig)
			sig.RequestType = domain.RequestTypeFLC
			reachedTerminal = handleOperationCompleted(ctx, &state, sig)
		})
		sel.AddReceive(nfrCompletedCh, func(c workflow.ReceiveChannel, _ bool) {
			var sig OperationCompletedSignal
			c.Receive(ctx, &sig)
			workflow.GetLogger(ctx).Info("NFR Completed signal received",
				"RequestID", sig.RequestID,
				"Outcome", sig.Outcome,
			)

			handleNFRCompleted(ctx, &state, sig)
		})
		sel.AddReceive(opCompletedCh, func(c workflow.ReceiveChannel, _ bool) {
			var sig OperationCompletedSignal
			c.Receive(ctx, &sig)
			reachedTerminal = handleOperationCompleted(ctx, &state, sig)
		})

		// ── Signal: system / compliance ───────────────────────────────────────
		sel.AddReceive(premiumPaidCh, func(c workflow.ReceiveChannel, _ bool) {
			var sig PremiumPaidSignal
			c.Receive(ctx, &sig)
			handlePremiumPaid(ctx, &state, sig)
		})
		sel.AddReceive(paymentDishonoredCh, func(c workflow.ReceiveChannel, _ bool) {
			var sig PaymentDishonoredSignal
			c.Receive(ctx, &sig)
			handlePaymentDishonored(ctx, &state, sig)
		})
		sel.AddReceive(amlRaisedCh, func(c workflow.ReceiveChannel, _ bool) {
			var sig AMLFlagRaisedSignal
			c.Receive(ctx, &sig)
			handleAMLFlagRaised(ctx, &state, sig)
		})
		sel.AddReceive(amlClearedCh, func(c workflow.ReceiveChannel, _ bool) {
			var sig AMLFlagClearedSignal
			c.Receive(ctx, &sig)
			handleAMLFlagCleared(ctx, &state, sig)
		})
		sel.AddReceive(investigationStartCh, func(c workflow.ReceiveChannel, _ bool) {
			var sig InvestigationStartedSignal
			c.Receive(ctx, &sig)
			handleInvestigationStarted(ctx, &state, sig)
		})
		sel.AddReceive(investigationConcludeCh, func(c workflow.ReceiveChannel, _ bool) {
			var sig InvestigationConcludedSignal
			c.Receive(ctx, &sig)
			reachedTerminal = handleInvestigationConcluded(ctx, &state, sig)
		})
		sel.AddReceive(loanBalanceCh, func(c workflow.ReceiveChannel, _ bool) {
			var sig LoanBalanceUpdatedSignal
			c.Receive(ctx, &sig)
			handleLoanBalanceUpdated(ctx, &state, sig)
		})
		sel.AddReceive(conversionReversedCh, func(c workflow.ReceiveChannel, _ bool) {
			var sig ConversionReversedSignal
			c.Receive(ctx, &sig)
			handleConversionReversed(ctx, &state, sig)
		})
		sel.AddReceive(adminVoidCh, func(c workflow.ReceiveChannel, _ bool) {
			var sig AdminVoidSignal
			c.Receive(ctx, &sig)
			reachedTerminal = handleAdminVoid(ctx, &state, sig)
		})
		sel.AddReceive(customerMergeCh, func(c workflow.ReceiveChannel, _ bool) {
			var sig CustomerIDMergeSignal
			c.Receive(ctx, &sig)
			handleCustomerIDMerge(ctx, &state, sig)
		})
		sel.AddReceive(voluntaryPUCh, func(c workflow.ReceiveChannel, _ bool) {
			var sig VoluntaryPaidUpSignal
			c.Receive(ctx, &sig)
			reachedTerminal = handleVoluntaryPaidUp(ctx, &state, sig)
		})
		sel.AddReceive(withdrawalCh, func(c workflow.ReceiveChannel, _ bool) {
			var sig WithdrawalRequestSignal
			c.Receive(ctx, &sig)
			handleWithdrawal(ctx, &state, sig)
		})
		sel.AddReceive(disputeRegisteredCh, func(c workflow.ReceiveChannel, _ bool) {
			var sig DisputeSignal
			c.Receive(ctx, &sig)
			handleDisputeRegistered(ctx, &state, sig)
		})
		sel.AddReceive(disputeResolvedCh, func(c workflow.ReceiveChannel, _ bool) {
			var sig DisputeSignal
			c.Receive(ctx, &sig)
			handleDisputeResolved(ctx, &state, sig)
		})
		sel.AddReceive(batchSyncCh, func(c workflow.ReceiveChannel, _ bool) {
			var sig BatchStateSyncSignal
			c.Receive(ctx, &sig)
			// IN-MEMORY ONLY — batch already wrote DB [Constraint 11, §9.5.2]
			state.PreviousStatus = state.CurrentStatus
			state.CurrentStatus = sig.NewStatus
			state.DisplayStatus = computeDisplayStatus(sig.NewStatus, state.Encumbrances)
			state.LastTransitionAt = workflow.Now(ctx) // [Review-Fix-7]
			state.Version++                            // [Review-Fix-7]: optimistic lock counter
			// No activity call — no DB write here
		})

		sel.Select(ctx)
		state.EventCount++
		pruneProcessedSignals(ctx, &state)

		// Update search attributes after every signal [§9.1]
		err := workflow.UpsertTypedSearchAttributes(
			ctx,
			CurrentStatusKey.ValueSet(state.CurrentStatus),
		)
		if err != nil {
			return err
		}
		if reachedTerminal {
			return handleTerminalCooling(ctx, &state)
		}
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Signal Handler: policy-created (A10.1.6)
// ─────────────────────────────────────────────────────────────────────────────

func handlePolicyCreated(ctx workflow.Context, state *PolicyLifecycleState, sig PolicyCreatedSignal) {
	// STEP 1: Dedup check [A10.1.6]

	PolicyNumberKey := temporal.NewSearchAttributeKeyKeyword("PolicyNumber")
	CurrentStatusKey := temporal.NewSearchAttributeKeyKeyword("CurrentStatus")
	ProductTypeKey := temporal.NewSearchAttributeKeyKeyword("ProductType")
	BillingMethodKey := temporal.NewSearchAttributeKeyKeyword("BillingMethod")
	IssueDateKey := temporal.NewSearchAttributeKeyTime("IssueDate")
	if _, seen := state.ProcessedSignalIDs[sig.RequestID]; seen {
		return
	}

	// STEP 2: InitializePolicyActivity — INSERT policy + initial status history row
	// Returns the BIGINT policy_id from seq_policy_id [A13]
	wfID := workflow.GetInfo(ctx).WorkflowExecution.ID
	var policyDBID int64
	err := workflow.ExecuteActivity(shortActCtx(ctx),
		policyActs.InitializePolicyActivity,
		acts.InitializePolicyParams{
			RequestID:      sig.RequestID,
			PolicyIssueID:  sig.PolicyID,
			PolicyNumber:   sig.PolicyNumber,
			WorkflowID:     wfID,
			CustomerID:     sig.Metadata.CustomerID,
			ProductCode:    sig.Metadata.ProductCode,
			ProductType:    sig.Metadata.ProductType,
			SumAssured:     sig.Metadata.SumAssured,
			CurrentPremium: sig.Metadata.CurrentPremium,
			PremiumMode:    sig.Metadata.PremiumMode,
			BillingMethod:  sig.Metadata.BillingMethod,
			IssueDate:      sig.Metadata.IssueDate,
			MaturityDate:   sig.Metadata.MaturityDate,
			PaidToDate:     sig.Metadata.PaidToDate,
			AgentID:        sig.Metadata.AgentID,
		}).Get(ctx, &policyDBID)
	if err != nil {
		// Activity has 3 retries; if all fail, workflow will be retried by Temporal
		return
	}

	// STEP 3: Store BIGINT policy_id in state
	state.PolicyDBID = policyDBID
	state.PolicyID = sig.PolicyID
	state.PolicyNumber = sig.PolicyNumber
	state.Metadata = sig.Metadata
	state.Metadata.WorkflowID = workflow.GetInfo(ctx).WorkflowExecution.ID

	// STEP 4: UpsertSearchAttributes
	err = workflow.UpsertTypedSearchAttributes(
		ctx,
		PolicyNumberKey.ValueSet(state.PolicyNumber),
		CurrentStatusKey.ValueSet(state.CurrentStatus),
		ProductTypeKey.ValueSet(state.Metadata.ProductType),
		BillingMethodKey.ValueSet(state.Metadata.BillingMethod),
		IssueDateKey.ValueSet(state.Metadata.IssueDate),
	)
	if err != nil {
		return
	}

	// STEP 5: PublishEventActivity
	_ = workflow.ExecuteActivity(shortActCtx(ctx),
		policyActs.PublishEventActivity,
		acts.PolicyEvent{
			PolicyID:     policyDBID,
			PolicyNumber: sig.PolicyNumber,
			EventType:    "POLICY_CREATED",
			OccurredAt:   workflow.Now(ctx),
		}).Get(ctx, nil)

	// STEP 6: Mark signal as processed
	state.ProcessedSignalIDs[sig.RequestID] = workflow.Now(ctx)

	// STEP 7: Batch-load routing timeout + cooling period configs into CachedConfig. [B4, D6]
	// CAN carries state forward, so keys already in the cache are skipped to avoid
	// redundant DB reads on each CAN cycle. A single FetchAllWorkflowConfigsActivity
	// call replaces 21 sequential FetchWorkflowConfigActivity calls. [Review-Fix-16]
	allConfigKeys := []string{
		domain.ConfigKeyRoutingTimeoutSurrender,
		domain.ConfigKeyRoutingTimeoutForcedSurrender,
		domain.ConfigKeyRoutingTimeoutLoan,
		domain.ConfigKeyRoutingTimeoutLoanRepayment,
		domain.ConfigKeyRoutingTimeoutRevival,
		domain.ConfigKeyRoutingTimeoutDeathClaim,
		domain.ConfigKeyRoutingTimeoutMaturityClaim,
		domain.ConfigKeyRoutingTimeoutSurvivalBenefit,
		domain.ConfigKeyRoutingTimeoutCommutation,
		domain.ConfigKeyRoutingTimeoutConversion,
		domain.ConfigKeyRoutingTimeoutFLC,
		domain.ConfigKeyRoutingTimeoutPremiumRefund,
		domain.ConfigKeyRoutingTimeoutNFR,
		domain.ConfigKeyCoolingPeriodVoid,
		domain.ConfigKeyCoolingPeriodSurrendered,
		domain.ConfigKeyCoolingPeriodTerminatedSurrender,
		domain.ConfigKeyCoolingPeriodMatured,
		domain.ConfigKeyCoolingPeriodDeathClaimSettled,
		domain.ConfigKeyCoolingPeriodFLCCancelled,
		domain.ConfigKeyCoolingPeriodCancelledDeath,
		domain.ConfigKeyCoolingPeriodConverted,
	}
	if state.CachedConfig == nil {
		state.CachedConfig = make(map[string]string)
	}
	// Build the list of keys not yet in cache (CAN carry-over guard). [D6]
	keysToFetch := make([]string, 0, len(allConfigKeys))
	for _, k := range allConfigKeys {
		if _, already := state.CachedConfig[k]; !already {
			keysToFetch = append(keysToFetch, k)
		}
	}
	if len(keysToFetch) > 0 {
		// One DB round-trip instead of N sequential activities. [D6]
		var batchedCfg map[string]string
		_ = workflow.ExecuteActivity(shortActCtx(ctx),
			policyActs.FetchAllWorkflowConfigsActivity, keysToFetch).Get(ctx, &batchedCfg)
		for k, v := range batchedCfg {
			state.CachedConfig[k] = v
		}
	}

	// STEP 8: Spawn FLC timer goroutine [Constraint 10, §10.1.6]
	// Distance-marketing products get 30-day FLC window by regulation; standard is 15 days.
	// Select the correct config key based on IsDistanceMarketing flag. [B8, Review-Fix-8]
	flcConfigKey := domain.ConfigKeyFLCPeriodDays
	if sig.Metadata.IsDistanceMarketing {
		flcConfigKey = domain.ConfigKeyFLCPeriodDistanceMarketing
	}
	var flcDaysStr string
	_ = workflow.ExecuteActivity(shortActCtx(ctx),
		policyActs.FetchWorkflowConfigActivity,
		flcConfigKey).Get(ctx, &flcDaysStr)
	flcDays, _ := strconv.Atoi(flcDaysStr)
	flcPeriod := getFLCPeriod(flcDays)
	// Persist the expiry time before spawning so it survives Continue-As-New. [D1]
	// Goroutines launched with workflow.Go are lost when PolicyLifecycleWorkflow
	// returns NewContinueAsNewError; the new workflow run must respawn the goroutine
	// using this field (see respawn block at the top of PolicyLifecycleWorkflow).
	state.FLCExpiryAt = workflow.Now(ctx).Add(flcPeriod)
	workflow.Go(ctx, func(gCtx workflow.Context) {
		_ = workflow.Sleep(gCtx, flcPeriod)
		if state.CurrentStatus == domain.StatusFreeLookActive {
			// FLC period expired without cancellation → transition to ACTIVE
			doTransition(gCtx, state, domain.StatusFreeLookActive, domain.StatusActive,
				"FLC period expired without cancellation", "flc-timer", "")
		}
		// Clear the persisted expiry so future runs don't re-respawn. [D1]
		state.FLCExpiryAt = time.Time{}
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// Signal Handler: financial requests (generic)
// ─────────────────────────────────────────────────────────────────────────────

func handleFinancialRequest(ctx workflow.Context, state *PolicyLifecycleState, sig PolicyRequestSignal) {
	// Use UUID idempotency key as dedup key when available; fall back to BIGINT string
	// for backwards compatibility with signals sent before the idempotency key was added.
	// [Review-Fix-11, Constraint 1]
	dedupKey := sig.IdempotencyKey
	if dedupKey == "" {
		dedupKey = strconv.FormatInt(sig.ServiceRequestID, 10)
	}

	// Dedup check
	if _, seen := state.ProcessedSignalIDs[dedupKey]; seen {
		return
	}

	// [C7] Refresh state from DB before eligibility checks.
	// SignalBatchStateSync may arrive after a DB-first batch operation, leaving
	// in-memory state stale. Refresh ensures isStateEligible sees the latest status. [§9.5.2]
	var refreshed *acts.PolicyRefreshedState
	if err := workflow.ExecuteActivity(shortActCtx(ctx),
		policyActs.RefreshStateFromDBActivity, state.PolicyDBID).Get(ctx, &refreshed); err == nil && refreshed != nil {
		state.CurrentStatus = refreshed.CurrentStatus
		state.Encumbrances.HasActiveLoan = refreshed.HasActiveLoan
		state.Metadata.LoanOutstanding = refreshed.LoanOutstanding // LoanOutstanding is on Metadata [C7]
		state.Encumbrances.AssignmentType = refreshed.AssignmentType
		state.Encumbrances.AMLHold = refreshed.AMLHold
		state.Version = refreshed.Version // Update version to match database
	}
	// If RefreshStateFromDBActivity fails: proceed with in-memory state (best-effort). [C7]

	// [C3] Prepare audit payload — logged after eligibility determination, not before.
	// This ensures the log status matches the actual outcome (PROCESSED vs REJECTED).
	sigPayload, _ := json.Marshal(sig)
	stateBefore := state.CurrentStatus
	logSignal := func(status string) {
		_ = workflow.ExecuteActivity(shortActCtx(ctx),
			policyActs.LogSignalReceivedActivity,
			acts.SignalLogEntry{
				PolicyID:      state.PolicyDBID,
				SignalChannel: requestTypeToSignalName(sig.RequestType), // [B9]
				SignalPayload: sigPayload,
				RequestID:     dedupKey,
				Status:        status,
				StateBefore:   &stateBefore,
			}).Get(ctx, nil)
	}

	// SUSPENDED blocks all financial requests (except death — handled separately) [BR-PM-110]
	if state.CurrentStatus == domain.StatusSuspended {
		logSignal(domain.SignalStatusRejected) // [C3] log REJECTED — not PROCESSED
		_ = workflow.ExecuteActivity(shortActCtx(ctx),
			policyActs.RecordRejectedRequestActivity,
			acts.RejectedRequestParams{
				PolicyID:         state.PolicyDBID,
				SignalChannel:    sig.RequestType,
				ServiceRequestID: &sig.ServiceRequestID,
				Reason:           "policy is SUSPENDED — request blocked [BR-PM-110]",
			}).Get(ctx, nil)
		return
	}

	// In-workflow state gate re-check (race condition guard)
	eligible, reason := isStateEligible(sig.RequestType, state.CurrentStatus, state.Encumbrances)
	if !eligible {
		logSignal(domain.SignalStatusRejected) // [C3] log REJECTED
		_ = workflow.ExecuteActivity(shortActCtx(ctx),
			policyActs.UpdateServiceRequestActivity,
			acts.ServiceRequestUpdate{
				ServiceRequestID: sig.ServiceRequestID,
				SubmittedAt:      sig.SubmittedAt, // [D4] partition key
				Status:           domain.RequestStatusStateGateRejected,
				OutcomeReason:    &reason,
			}).Get(ctx, nil)
		state.ProcessedSignalIDs[dedupKey] = workflow.Now(ctx)
		return
	}

	// Financial lock check [BR-PM-030]
	if requiresFinancialLock(sig.RequestType) && state.ActiveLock != nil {
		lockReason := fmt.Sprintf("financial lock held by %s [BR-PM-030]", state.ActiveLock.RequestType)
		logSignal(domain.SignalStatusRejected) // [C3] log REJECTED
		_ = workflow.ExecuteActivity(shortActCtx(ctx),
			policyActs.UpdateServiceRequestActivity,
			acts.ServiceRequestUpdate{
				ServiceRequestID: sig.ServiceRequestID,
				SubmittedAt:      sig.SubmittedAt, // [D4] partition key
				Status:           domain.RequestStatusStateGateRejected,
				OutcomeReason:    &lockReason,
			}).Get(ctx, nil)
		state.ProcessedSignalIDs[dedupKey] = workflow.Now(ctx)
		return
	}

	// [C3] All checks passed — log PROCESSED once before proceeding with request creation
	logSignal(domain.SignalStatusProcessed)

	// Pre-route state transition (if applicable)
	preRoute := preRouteStatus(sig.RequestType)
	if preRoute != "" {
		doTransition(ctx, state, state.CurrentStatus, preRoute,
			fmt.Sprintf("%s request received", sig.RequestType), sig.RequestType, dedupKey)
	}

	// Set financial lock (if required) [BR-PM-030]
	timeout := workflow.Now(ctx).Add(routingTimeoutFromConfig(state.CachedConfig, sig.RequestType)) // [B4, Review-Fix-16]
	if requiresFinancialLock(sig.RequestType) {
		lockedAt := workflow.Now(ctx)
		state.ActiveLock = &FinancialLock{
			RequestID:   dedupKey,
			RequestType: sig.RequestType,
			LockedAt:    lockedAt,
			TimeoutAt:   timeout,
		}
		// Persist lock to DB so it survives worker restarts [Review-Fix-1, BR-PM-030]
		_ = workflow.ExecuteActivity(shortActCtx(ctx),
			policyActs.AcquireFinancialLockActivity,
			acts.FinancialLockParams{
				PolicyID:         state.PolicyDBID,
				ServiceRequestID: sig.ServiceRequestID,
				RequestType:      sig.RequestType,
				LockedAt:         lockedAt,
				TimeoutAt:        timeout,
			}).Get(ctx, nil)
	}

	// Build child workflow ID — format: {prefix}-{idempotencyKey} per Constraint 1.
	// Using the UUID idempotency key (not BIGINT) ensures global uniqueness and
	// matches downstream correlation expectations. [Constraint 1, Review-Fix-3]
	prefix := DownstreamChildIDPrefix(sig.RequestType)
	childID := fmt.Sprintf("%s-%s", prefix, dedupKey)
	taskQueue := domain.DownstreamTaskQueueForType(sig.RequestType)
	wfType := DownstreamWorkflowTypeForRequest(sig.RequestType)

	// Route to downstream via ExecuteChildWorkflow (fire-and-forget) [Constraint 1]
	childInput := ChildWorkflowInput{
		RequestID:        dedupKey,
		PolicyNumber:     state.PolicyNumber,
		PolicyDBID:       state.PolicyDBID,
		ServiceRequestID: sig.ServiceRequestID,
		RequestType:      sig.RequestType,
		TimeoutAt:        timeout,
	}
	workflow.ExecuteChildWorkflow(childWFCtx(ctx, taskQueue, childID), wfType, childInput)

	// Add to pending requests — carry SubmittedAt for partition-key pruning [D4]
	state.PendingRequests = append(state.PendingRequests, PendingRequest{
		RequestID:          dedupKey,
		ServiceRequestID:   sig.ServiceRequestID,
		RequestType:        sig.RequestType,
		RequestCategory:    sig.RequestCategory,
		DownstreamWorkflow: childID,
		RoutedAt:           workflow.Now(ctx),
		TimeoutAt:          timeout,
		SubmittedAt:        sig.SubmittedAt, // [D4]
	})

	// Update service_request to ROUTED
	routedAt := workflow.Now(ctx)
	downstreamSvc := taskQueue
	_ = workflow.ExecuteActivity(shortActCtx(ctx),
		policyActs.UpdateServiceRequestActivity,
		acts.ServiceRequestUpdate{
			ServiceRequestID:     sig.ServiceRequestID,
			SubmittedAt:          sig.SubmittedAt, // [D4] partition key
			Status:               domain.RequestStatusRouted,
			DownstreamWorkflowID: &childID,
			DownstreamService:    &downstreamSvc,
			DownstreamTaskQueue:  &taskQueue,
			RoutedAt:             &routedAt,
		}).Get(ctx, nil)

	state.ProcessedSignalIDs[dedupKey] = workflow.Now(ctx)
}

// ─────────────────────────────────────────────────────────────────────────────
// Signal Handler: death-notification (PREEMPTIVE — overrides SUSPENDED) [BR-PM-112]
// ─────────────────────────────────────────────────────────────────────────────

func handleDeathNotification(ctx workflow.Context, state *PolicyLifecycleState, sig PolicyRequestSignal) {
	// [D11] Dedup check BEFORE preemption so a replayed/duplicate death signal does
	// not re-cancel child workflows or release the DB lock a second time.
	// Previously the dedup was inside handleFinancialRequest (called at the bottom),
	// but by then the preemption actions had already run. [BR-PM-112]
	dedupKey := sig.IdempotencyKey
	if dedupKey == "" {
		dedupKey = strconv.FormatInt(sig.ServiceRequestID, 10)
	}
	if _, seen := state.ProcessedSignalIDs[dedupKey]; seen {
		return
	}

	// Cancel active child workflow + release financial lock (preemption) [BR-PM-112]
	if state.ActiveLock != nil {
		activeChildID := ""
		for _, pr := range state.PendingRequests {
			if pr.RequestID == state.ActiveLock.RequestID {
				activeChildID = pr.DownstreamWorkflow
				break
			}
		}
		if activeChildID != "" {
			_ = workflow.ExecuteActivity(shortActCtx(ctx),
				policyActs.CancelDownstreamWorkflowActivity,
				acts.CancelWorkflowParams{WorkflowID: activeChildID}).Get(ctx, nil)
		}
		state.ActiveLock = nil
		// Release DB lock — preempted by death notification [Review-Fix-1, BR-PM-030]
		_ = workflow.ExecuteActivity(shortActCtx(ctx),
			policyActs.ReleaseFinancialLockActivity, state.PolicyDBID).Get(ctx, nil)
	}
	// Remove all pending financial requests
	state.PendingRequests = nil

	// Now handle as regular financial request (SUSPENDED state is bypassed for death)
	handleFinancialRequest(ctx, state, sig)
}

// ─────────────────────────────────────────────────────────────────────────────
// Signal Handler: NFR requests
// ─────────────────────────────────────────────────────────────────────────────

func handleNFRRequest(ctx workflow.Context, state *PolicyLifecycleState, sig PolicyRequestSignal) {
	// Use UUID idempotency key as dedup key when available [Review-Fix-11]
	dedupKey := sig.IdempotencyKey
	if dedupKey == "" {
		dedupKey = strconv.FormatInt(sig.ServiceRequestID, 10)
	}
	if _, seen := state.ProcessedSignalIDs[dedupKey]; seen {
		return
	}

	// [C7] Refresh state from DB before eligibility check.
	// SignalBatchStateSync may arrive after a DB-first batch operation,
	// leaving in-memory state stale. [§9.5.2]
	var nfrRefreshed *acts.PolicyRefreshedState
	if err := workflow.ExecuteActivity(shortActCtx(ctx),
		policyActs.RefreshStateFromDBActivity, state.PolicyDBID).Get(ctx, &nfrRefreshed); err == nil && nfrRefreshed != nil {
		state.CurrentStatus = nfrRefreshed.CurrentStatus
		state.Encumbrances.HasActiveLoan = nfrRefreshed.HasActiveLoan
		state.Metadata.LoanOutstanding = nfrRefreshed.LoanOutstanding // LoanOutstanding is on Metadata [C7]
		state.Encumbrances.AssignmentType = nfrRefreshed.AssignmentType
		state.Encumbrances.AMLHold = nfrRefreshed.AMLHold
		state.Version = nfrRefreshed.Version // Update version to match database
	}

	// [C3] Prepare audit payload — logged after eligibility determination.
	nfrPayload, _ := json.Marshal(sig)
	nfrStateBefore := state.CurrentStatus
	logNFRSignal := func(status string) {
		_ = workflow.ExecuteActivity(shortActCtx(ctx),
			policyActs.LogSignalReceivedActivity,
			acts.SignalLogEntry{
				PolicyID:      state.PolicyDBID,
				SignalChannel: SignalNFRRequest,
				SignalPayload: nfrPayload,
				RequestID:     dedupKey,
				Status:        status,
				StateBefore:   &nfrStateBefore,
			}).Get(ctx, nil)
	}

	// NFR: allowed in all non-terminal non-SUSPENDED states [BR-PM-023]
	eligible, reason := isStateEligible(sig.RequestType, state.CurrentStatus, state.Encumbrances)
	if !eligible {
		logNFRSignal(domain.SignalStatusRejected) // [C3] log REJECTED
		_ = workflow.ExecuteActivity(shortActCtx(ctx),
			policyActs.UpdateServiceRequestActivity,
			acts.ServiceRequestUpdate{
				ServiceRequestID: sig.ServiceRequestID,
				SubmittedAt:      sig.SubmittedAt, // [D4] partition key
				Status:           domain.RequestStatusStateGateRejected,
				OutcomeReason:    &reason,
			}).Get(ctx, nil)
		state.ProcessedSignalIDs[dedupKey] = workflow.Now(ctx)
		return
	}

	logNFRSignal(domain.SignalStatusProcessed) // [C3] log PROCESSED — all checks passed

	prefix := DownstreamChildIDPrefix(sig.RequestType)
	childID := fmt.Sprintf("%s-%s", prefix, dedupKey) // [Review-Fix-3/11]: UUID-based child ID, not BIGINT
	taskQueue := domain.DownstreamTaskQueueForType(sig.RequestType)
	wfType := DownstreamWorkflowTypeForRequest(sig.RequestType)
	timeout := workflow.Now(ctx).Add(routingTimeoutFromConfig(state.CachedConfig, sig.RequestType)) // [B4, Review-Fix-16]

	childInput := ChildWorkflowInput{
		RequestID:        dedupKey,
		PolicyNumber:     state.PolicyNumber,
		PolicyDBID:       state.PolicyDBID,
		ServiceRequestID: sig.ServiceRequestID,
		RequestType:      sig.RequestType,
		TimeoutAt:        timeout,
	}
	workflow.ExecuteChildWorkflow(childWFCtx(ctx, taskQueue, childID), wfType, childInput)

	state.PendingRequests = append(state.PendingRequests, PendingRequest{
		RequestID:          dedupKey,
		ServiceRequestID:   sig.ServiceRequestID,
		RequestType:        sig.RequestType,
		RequestCategory:    domain.RequestCategoryNonFinancial,
		DownstreamWorkflow: childID,
		RoutedAt:           workflow.Now(ctx),
		TimeoutAt:          timeout,
		SubmittedAt:        sig.SubmittedAt, // [D4]
	})

	routedAt := workflow.Now(ctx)
	_ = workflow.ExecuteActivity(shortActCtx(ctx),
		policyActs.UpdateServiceRequestActivity,
		acts.ServiceRequestUpdate{
			ServiceRequestID:     sig.ServiceRequestID,
			SubmittedAt:          sig.SubmittedAt, // [D4] partition key
			Status:               domain.RequestStatusRouted,
			DownstreamWorkflowID: &childID,
			DownstreamService:    &taskQueue,
			DownstreamTaskQueue:  &taskQueue,
			RoutedAt:             &routedAt,
		}).Get(ctx, nil)

	state.ProcessedSignalIDs[dedupKey] = workflow.Now(ctx)
}

// ─────────────────────────────────────────────────────────────────────────────
// Signal Handler: operation-completed (generic + per-type)
// ─────────────────────────────────────────────────────────────────────────────

// handleOperationCompleted resolves completion for financial requests.
// Returns true if a terminal state was reached.
func handleOperationCompleted(ctx workflow.Context, state *PolicyLifecycleState, sig OperationCompletedSignal) bool {
	// Dedup check — downstream services may retry operation-completed signals. [B5]
	if _, seen := state.ProcessedSignalIDs[sig.RequestID]; seen {
		return false
	}

	// Audit: log signal receipt in policy_signal_log. [B5, Review-Fix-2]
	sigPayload, _ := json.Marshal(sig)
	stateBefore := state.CurrentStatus
	_ = workflow.ExecuteActivity(shortActCtx(ctx),
		policyActs.LogSignalReceivedActivity,
		acts.SignalLogEntry{
			PolicyID:      state.PolicyDBID,
			SignalChannel: "operation-completed",
			SignalPayload: sigPayload,
			RequestID:     sig.RequestID,
			Status:        domain.SignalStatusProcessed,
			StateBefore:   &stateBefore,
		}).Get(ctx, nil)

	// Find and remove matching pending request
	var matched *PendingRequest
	remaining := state.PendingRequests[:0]
	for i := range state.PendingRequests {
		pr := &state.PendingRequests[i]
		if pr.RequestID == sig.RequestID { // [Review-Fix-10]: strict match — no RequestType fallback
			matched = pr
		} else {
			remaining = append(remaining, *pr)
		}
	}
	state.PendingRequests = remaining

	// [C2] Guard: if no pending request matched, the signal is orphaned (late, duplicate, or misrouted).
	// Return false without attempting to update a service_request with ID=0.
	if matched == nil {
		logger := workflow.GetLogger(ctx)
		logger.Warn("handleOperationCompleted: no pending request matched signal — signal ignored",
			"policyID", state.PolicyDBID,
			"requestID", sig.RequestID,
			"requestType", sig.RequestType,
		)
		state.ProcessedSignalIDs[sig.RequestID] = workflow.Now(ctx) // dedup to prevent re-processing
		return false
	}

	// Release financial lock if this request held it
	if state.ActiveLock != nil && state.ActiveLock.RequestID == sig.RequestID { // [Review-Fix-10]
		state.ActiveLock = nil
		// Release DB lock so subsequent financial requests can proceed [Review-Fix-1, BR-PM-030]
		_ = workflow.ExecuteActivity(shortActCtx(ctx),
			policyActs.ReleaseFinancialLockActivity, state.PolicyDBID).Get(ctx, nil)
	}

	// Update service_request to COMPLETED
	completedAt := workflow.Now(ctx)
	outcome := sig.Outcome
	_ = workflow.ExecuteActivity(shortActCtx(ctx),
		policyActs.UpdateServiceRequestActivity,
		acts.ServiceRequestUpdate{
			ServiceRequestID: func() int64 {
				if matched != nil {
					return matched.ServiceRequestID
				}
				return 0
			}(),
			SubmittedAt:    matched.SubmittedAt, // [D4] partition key from PendingRequest
			Status:         domain.RequestStatusCompleted,
			Outcome:        &outcome,
			CompletedAt:    &completedAt,
			OutcomePayload: sig.OutcomePayload,
		}).Get(ctx, nil)

	// Apply state transition based on request type and outcome
	newStatus, isTerminal := resolveCompletionTransition(state, sig)

	if newStatus != "" && newStatus != state.CurrentStatus {
		doTransition(ctx, state, state.CurrentStatus, newStatus,
			fmt.Sprintf("%s %s", sig.RequestType, sig.Outcome), sig.RequestType, sig.RequestID)
	}

	// Update encumbrances for loan completion
	updateEncumbrancesFromCompletion(ctx, state, sig)

	// Mark signal as processed to prevent double-processing on retried signals. [B5]
	state.ProcessedSignalIDs[sig.RequestID] = workflow.Now(ctx)

	return isTerminal
}

// handleClaimSettled handles the claim-settled completion signal (covers death, maturity, SB).
func handleClaimSettled(ctx workflow.Context, state *PolicyLifecycleState, sig OperationCompletedSignal) bool {
	return handleOperationCompleted(ctx, state, sig)
}

// handleNFRCompleted handles NFR completion; updates metadata. No state change.
func handleNFRCompleted(ctx workflow.Context, state *PolicyLifecycleState, sig OperationCompletedSignal) {
	var matched *PendingRequest
	remaining := state.PendingRequests[:0]
	for i := range state.PendingRequests {
		pr := &state.PendingRequests[i]
		if pr.RequestID == sig.RequestID {
			matched = pr
		} else {
			remaining = append(remaining, *pr)
		}
	}
	state.PendingRequests = remaining

	if matched == nil {
		return
	}

	completedAt := workflow.Now(ctx)
	outcome := sig.Outcome
	_ = workflow.ExecuteActivity(shortActCtx(ctx),
		policyActs.UpdateServiceRequestActivity,
		acts.ServiceRequestUpdate{
			ServiceRequestID: matched.ServiceRequestID,
			SubmittedAt:      matched.SubmittedAt, // [D4] partition key from PendingRequest
			Status:           domain.RequestStatusCompleted,
			Outcome:          &outcome,
			CompletedAt:      &completedAt,
		}).Get(ctx, nil)

	// Metadata update for NFR types that report outcome payload (e.g. assignment).
	// Only allow-listed keys are forwarded to prevent arbitrary metadata corruption. [B7]
	if sig.Outcome == domain.RequestOutcomeApproved && sig.OutcomePayload != nil {
		var payload map[string]interface{}
		if json.Unmarshal(sig.OutcomePayload, &payload) == nil {
			filtered := filterNFRPayload(matched.RequestType, payload)
			if len(filtered) > 0 {
				_ = workflow.ExecuteActivity(shortActCtx(ctx),
					policyActs.UpdatePolicyMetadataActivity,
					acts.MetadataUpdateParams{
						PolicyID: state.PolicyDBID,
						Updates:  filtered,
					}).Get(ctx, nil)
			}
		}
	}
}

// nfrMetadataAllowList defines safe outcome payload keys per NFR type. [B7]
// Only these keys may be written to policy_metadata via UpdatePolicyMetadataActivity;
// all other keys in OutcomePayload are silently discarded to prevent a buggy or rogue
// downstream service from overwriting protected columns (customer_id, sum_assured, etc.)
var nfrMetadataAllowList = map[string]map[string]bool{
	domain.RequestTypeAssignment: {
		"assignment_type":   true,
		"assignment_status": true,
		"assignee_name":     true,
		"assignee_address":  true,
	},
}

// filterNFRPayload returns only allow-listed keys from payload for the given NFR type.
// Returns nil (not an empty map) when the type has no allow-list, suppressing the activity call.
func filterNFRPayload(requestType string, payload map[string]interface{}) map[string]interface{} {
	allowed, ok := nfrMetadataAllowList[requestType]
	if !ok {
		return nil // no metadata updates permitted for this NFR type
	}
	filtered := make(map[string]interface{}, len(allowed))
	for k, v := range payload {
		if allowed[k] {
			filtered[k] = v
		}
	}
	return filtered
}

// ─────────────────────────────────────────────────────────────────────────────
// resolveCompletionTransition — maps outcome → new status [§9.1 completion rules]
// ─────────────────────────────────────────────────────────────────────────────

func resolveCompletionTransition(state *PolicyLifecycleState, sig OperationCompletedSignal) (newStatus string, isTerminal bool) {
	approved := sig.Outcome == domain.RequestOutcomeApproved
	switch sig.RequestType {
	case domain.RequestTypeSurrender:
		if approved {
			return domain.StatusSurrendered, true // → terminal cooling 90d [Constraint 9]
		}
		return state.PreviousStatus, false // revert from PENDING_SURRENDER

	case domain.RequestTypeForcedSurrender:
		if approved {
			// REDUCED_PAID_UP if net ≥ prescribed limit; otherwise TERMINATED_SURRENDER
			if sig.StateTransition == "REDUCED_PAID_UP" {
				return domain.StatusReducedPaidUp, false
			}
			return domain.StatusTerminatedSurrender, true // → terminal cooling 90d
		}
		return state.PreviousStatus, false // revert from PAS

	case domain.RequestTypeLoan:
		// No state change; encumbrances updated separately in updateEncumbrancesFromCompletion
		return "", false

	case domain.RequestTypeLoanRepayment:
		// No state change; encumbrances updated separately
		return "", false

	case domain.RequestTypeRevival:
		if approved {
			return domain.StatusActive, false // + update PaidToDate
		}
		return state.PreviousStatus, false // revert to VL/IL/AL

	case domain.RequestTypeDeathClaim, domain.RequestTypeMaturityClaim, domain.RequestTypeSurvivalBenefit:
		if approved {
			switch sig.RequestType {
			case domain.RequestTypeDeathClaim:
				return domain.StatusDeathClaimSettled, true // → terminal cooling 180d
			case domain.RequestTypeMaturityClaim:
				return domain.StatusMatured, true // → terminal cooling 90d
			case domain.RequestTypeSurvivalBenefit:
				// No state change; update sb_installments_paid
				return "", false
			}
		}
		if sig.Outcome == domain.RequestOutcomeRejected {
			return state.PreviousStatus, false // revert from DCI / pending states
		}
		return "", false

	case domain.RequestTypeCommutation:
		// No state change; update sum_assured + current_premium (metadata)
		return "", false

	case domain.RequestTypeConversion:
		if approved {
			return domain.StatusConverted, true // → terminal cooling 90d
		}
		return "", false

	case domain.RequestTypeFLC:
		if approved {
			return domain.StatusFLCCancelled, true // → terminal cooling 30d
		}
		return "", false
	}
	return "", false
}

// updateEncumbrancesFromCompletion updates loan/assignment encumbrances after
// loan-completed or loan-repayment-completed signals. [§9.1]
func updateEncumbrancesFromCompletion(ctx workflow.Context, state *PolicyLifecycleState, sig OperationCompletedSignal) {
	switch sig.RequestType {
	case domain.RequestTypeLoan:
		if sig.Outcome == domain.RequestOutcomeApproved {
			state.Encumbrances.HasActiveLoan = true
			state.Encumbrances.AssignmentType = "ABSOLUTE" // Loan = absolute assignment to President
			_ = workflow.ExecuteActivity(shortActCtx(ctx),
				policyActs.UpdatePolicyMetadataActivity,
				acts.MetadataUpdateParams{
					PolicyID: state.PolicyDBID,
					Updates: map[string]interface{}{
						"has_active_loan":   true,
						"assignment_type":   "ABSOLUTE",
						"assignment_status": "ASSIGNED_TO_PRESIDENT",
					},
				}).Get(ctx, nil)
		}
	case domain.RequestTypeLoanRepayment:
		if sig.Outcome == domain.RequestOutcomeApproved {
			state.Encumbrances.HasActiveLoan = false
			state.Metadata.LoanOutstanding = 0
			_ = workflow.ExecuteActivity(shortActCtx(ctx),
				policyActs.UpdatePolicyMetadataActivity,
				acts.MetadataUpdateParams{
					PolicyID: state.PolicyDBID,
					Updates: map[string]interface{}{
						"has_active_loan":  false,
						"loan_outstanding": 0,
						"assignment_type":  "NONE",
					},
				}).Get(ctx, nil)
		}
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Signal Handlers: system / compliance signals
// ─────────────────────────────────────────────────────────────────────────────

func handlePremiumPaid(ctx workflow.Context, state *PolicyLifecycleState, sig PremiumPaidSignal) {
	if _, seen := state.ProcessedSignalIDs[sig.RequestID]; seen {
		return
	}
	// Audit: log signal receipt [B6, Review-Fix-2]
	sigPayload, _ := json.Marshal(sig)
	stateBefore := state.CurrentStatus
	_ = workflow.ExecuteActivity(shortActCtx(ctx),
		policyActs.LogSignalReceivedActivity,
		acts.SignalLogEntry{
			PolicyID:      state.PolicyDBID,
			SignalChannel: SignalPremiumPaid,
			SignalPayload: sigPayload,
			RequestID:     sig.RequestID,
			Status:        domain.SignalStatusProcessed,
			StateBefore:   &stateBefore,
		}).Get(ctx, nil)
	state.Metadata.PaidToDate = sig.NewPaidToDate

	// Lapse-revival: if VOID_LAPSE or INACTIVE_LAPSE and within remission window → ACTIVE
	// nil RemissionExpiryDate means no remission window — revival not eligible [Review-Fix-5]
	if (state.CurrentStatus == domain.StatusVoidLapse || state.CurrentStatus == domain.StatusInactiveLapse) &&
		state.Metadata.RemissionExpiryDate != nil &&
		sig.PaymentDate.Before(*state.Metadata.RemissionExpiryDate) {
		doTransition(ctx, state, state.CurrentStatus, domain.StatusActive,
			"premium paid within remission period", SignalPremiumPaid, sig.RequestID)
	}
	// PENDING_AUTO_SURRENDER → ACTIVE if payment within window
	if state.CurrentStatus == domain.StatusPendingAutoSurrender {
		doTransition(ctx, state, state.CurrentStatus, domain.StatusActive,
			"premium paid clears pending auto surrender", SignalPremiumPaid, sig.RequestID)
	}
	_ = workflow.ExecuteActivity(shortActCtx(ctx),
		policyActs.UpdatePolicyMetadataActivity,
		acts.MetadataUpdateParams{
			PolicyID: state.PolicyDBID,
			Updates:  map[string]interface{}{"paid_to_date": sig.NewPaidToDate},
		}).Get(ctx, nil)
	state.ProcessedSignalIDs[sig.RequestID] = workflow.Now(ctx)
}

func handlePaymentDishonored(ctx workflow.Context, state *PolicyLifecycleState, sig PaymentDishonoredSignal) {
	// [v4.1 corrected transitions #65/65a]
	// policy_life < 3 years → VOID_LAPSE; policy_life ≥ 3 years → INACTIVE_LAPSE
	if _, seen := state.ProcessedSignalIDs[sig.RequestID]; seen {
		return
	}
	// Audit: log signal receipt [B6, Review-Fix-2]
	sigPayload, _ := json.Marshal(sig)
	stateBefore := state.CurrentStatus
	_ = workflow.ExecuteActivity(shortActCtx(ctx),
		policyActs.LogSignalReceivedActivity,
		acts.SignalLogEntry{
			PolicyID:      state.PolicyDBID,
			SignalChannel: SignalPaymentDishonored,
			SignalPayload: sigPayload,
			RequestID:     sig.RequestID,
			Status:        domain.SignalStatusProcessed,
			StateBefore:   &stateBefore,
		}).Get(ctx, nil)
	if state.CurrentStatus != domain.StatusActive {
		state.ProcessedSignalIDs[sig.RequestID] = workflow.Now(ctx)
		return
	}
	policyLifeMonths := int(sig.DishonoredDate.Sub(state.Metadata.IssueDate).Hours() / (24 * 30))
	newStatus := domain.StatusVoidLapse
	if policyLifeMonths >= 36 {
		newStatus = domain.StatusInactiveLapse
	}
	// Recalculate remission_expiry_date (12 calendar months from dishonored date). [D5]
	// Use AddDate(0,12,0) instead of 12*30*24h — the latter drifts by up to 5 days
	// versus the DB compute_remission_expiry() function which uses interval '12 months'.
	remissionExpiry := sig.DishonoredDate.AddDate(0, 12, 0)
	state.Metadata.RemissionExpiryDate = &remissionExpiry // [Review-Fix-5]
	doTransition(ctx, state, domain.StatusActive, newStatus,
		fmt.Sprintf("payment dishonored: %s", sig.Reason), SignalPaymentDishonored, sig.RequestID)
	_ = workflow.ExecuteActivity(shortActCtx(ctx),
		policyActs.UpdatePolicyMetadataActivity,
		acts.MetadataUpdateParams{
			PolicyID: state.PolicyDBID,
			Updates:  map[string]interface{}{"remission_expiry_date": remissionExpiry},
		}).Get(ctx, nil)
	state.ProcessedSignalIDs[sig.RequestID] = workflow.Now(ctx)
}

func handleAMLFlagRaised(ctx workflow.Context, state *PolicyLifecycleState, sig AMLFlagRaisedSignal) {
	// Save previous status for restore; transition to SUSPENDED [BR-PM-110]
	if _, seen := state.ProcessedSignalIDs[sig.RequestID]; seen {
		return
	}
	// Audit: log signal receipt [B6, Review-Fix-2]
	sigPayload, _ := json.Marshal(sig)
	stateBefore := state.CurrentStatus
	_ = workflow.ExecuteActivity(shortActCtx(ctx),
		policyActs.LogSignalReceivedActivity,
		acts.SignalLogEntry{
			PolicyID:      state.PolicyDBID,
			SignalChannel: SignalAMLFlagRaised,
			SignalPayload: sigPayload,
			RequestID:     sig.RequestID,
			Status:        domain.SignalStatusProcessed,
			StateBefore:   &stateBefore,
		}).Get(ctx, nil)
	state.PreviousStatusBeforeSuspension = state.CurrentStatus
	state.Encumbrances.AMLHold = true
	doTransition(ctx, state, state.CurrentStatus, domain.StatusSuspended,
		fmt.Sprintf("AML flag raised: %s", sig.Reason), SignalAMLFlagRaised, sig.RequestID)
	state.ProcessedSignalIDs[sig.RequestID] = workflow.Now(ctx)
}

func handleAMLFlagCleared(ctx workflow.Context, state *PolicyLifecycleState, sig AMLFlagClearedSignal) {
	// Restore previous status [BR-PM-111]
	if _, seen := state.ProcessedSignalIDs[sig.RequestID]; seen {
		return
	}
	// Audit: log signal receipt [B6, Review-Fix-2]
	sigPayload, _ := json.Marshal(sig)
	stateBefore := state.CurrentStatus
	_ = workflow.ExecuteActivity(shortActCtx(ctx),
		policyActs.LogSignalReceivedActivity,
		acts.SignalLogEntry{
			PolicyID:      state.PolicyDBID,
			SignalChannel: SignalAMLFlagCleared,
			SignalPayload: sigPayload,
			RequestID:     sig.RequestID,
			Status:        domain.SignalStatusProcessed,
			StateBefore:   &stateBefore,
		}).Get(ctx, nil)
	prevStatus := state.PreviousStatusBeforeSuspension
	if prevStatus == "" {
		prevStatus = domain.StatusActive // fallback
	}
	state.Encumbrances.AMLHold = false
	state.PreviousStatusBeforeSuspension = ""
	doTransition(ctx, state, domain.StatusSuspended, prevStatus,
		"AML flag cleared — restoring previous status", SignalAMLFlagCleared, sig.RequestID)
	state.ProcessedSignalIDs[sig.RequestID] = workflow.Now(ctx)
}

func handleInvestigationStarted(ctx workflow.Context, state *PolicyLifecycleState, sig InvestigationStartedSignal) {
	// DEATH_CLAIM_INTIMATED → DEATH_UNDER_INVESTIGATION [BR-PM-120]
	if _, seen := state.ProcessedSignalIDs[sig.RequestID]; seen {
		return
	}
	// Audit: log signal receipt [B6, Review-Fix-2]
	sigPayload, _ := json.Marshal(sig)
	stateBefore := state.CurrentStatus
	_ = workflow.ExecuteActivity(shortActCtx(ctx),
		policyActs.LogSignalReceivedActivity,
		acts.SignalLogEntry{
			PolicyID:      state.PolicyDBID,
			SignalChannel: SignalInvestigationStarted,
			SignalPayload: sigPayload,
			RequestID:     sig.RequestID,
			Status:        domain.SignalStatusProcessed,
			StateBefore:   &stateBefore,
		}).Get(ctx, nil)
	if state.CurrentStatus == domain.StatusDeathClaimIntimated {
		doTransition(ctx, state, domain.StatusDeathClaimIntimated, domain.StatusDeathUnderInvestigation,
			"death investigation started", SignalInvestigationStarted, sig.RequestID)
	}
	state.ProcessedSignalIDs[sig.RequestID] = workflow.Now(ctx)
}

func handleInvestigationConcluded(ctx workflow.Context, state *PolicyLifecycleState, sig InvestigationConcludedSignal) bool {
	// CONFIRMED → DEATH_CLAIM_SETTLED (terminal); REJECTED → revert [BR-PM-121]
	if _, seen := state.ProcessedSignalIDs[sig.RequestID]; seen {
		return false
	}
	// Audit: log signal receipt [B6, Review-Fix-2]
	sigPayload, _ := json.Marshal(sig)
	stateBefore := state.CurrentStatus
	_ = workflow.ExecuteActivity(shortActCtx(ctx),
		policyActs.LogSignalReceivedActivity,
		acts.SignalLogEntry{
			PolicyID:      state.PolicyDBID,
			SignalChannel: SignalInvestigationConcluded,
			SignalPayload: sigPayload,
			RequestID:     sig.RequestID,
			Status:        domain.SignalStatusProcessed,
			StateBefore:   &stateBefore,
		}).Get(ctx, nil)
	if state.CurrentStatus == domain.StatusDeathUnderInvestigation {
		if sig.Outcome == "CONFIRMED" {
			doTransition(ctx, state, domain.StatusDeathUnderInvestigation, domain.StatusDeathClaimSettled,
				"investigation confirmed — claim settled", SignalInvestigationConcluded, sig.RequestID)
			state.ProcessedSignalIDs[sig.RequestID] = workflow.Now(ctx)
			return true
		}
		// REJECTED → revert to DCI
		doTransition(ctx, state, domain.StatusDeathUnderInvestigation, domain.StatusDeathClaimIntimated,
			"investigation rejected — reverting", SignalInvestigationConcluded, sig.RequestID)
	}
	state.ProcessedSignalIDs[sig.RequestID] = workflow.Now(ctx)
	return false
}

func handleConversionReversed(ctx workflow.Context, state *PolicyLifecycleState, sig ConversionReversedSignal) {
	// CONVERTED → PreviousStatus (cheque bounce) [BR-CHQ-001]
	if _, seen := state.ProcessedSignalIDs[sig.RequestID]; seen {
		return
	}
	// Audit: log signal receipt [B6, Review-Fix-2]
	sigPayload, _ := json.Marshal(sig)
	stateBefore := state.CurrentStatus
	_ = workflow.ExecuteActivity(shortActCtx(ctx),
		policyActs.LogSignalReceivedActivity,
		acts.SignalLogEntry{
			PolicyID:      state.PolicyDBID,
			SignalChannel: SignalConversionReversed,
			SignalPayload: sigPayload,
			RequestID:     sig.RequestID,
			Status:        domain.SignalStatusProcessed,
			StateBefore:   &stateBefore,
		}).Get(ctx, nil)
	if state.CurrentStatus == domain.StatusConverted && state.PreviousStatus != "" {
		doTransition(ctx, state, domain.StatusConverted, state.PreviousStatus,
			fmt.Sprintf("conversion reversed: %s", sig.Reason), SignalConversionReversed, sig.RequestID)
	}
	state.ProcessedSignalIDs[sig.RequestID] = workflow.Now(ctx)
}

func handleLoanBalanceUpdated(ctx workflow.Context, state *PolicyLifecycleState, sig LoanBalanceUpdatedSignal) {
	// Dedup + audit for loan balance update. No state-machine transition — in-memory + DB only. [B6]
	if _, seen := state.ProcessedSignalIDs[sig.RequestID]; seen {
		return
	}
	sigPayload, _ := json.Marshal(sig)
	stateBefore := state.CurrentStatus
	_ = workflow.ExecuteActivity(shortActCtx(ctx),
		policyActs.LogSignalReceivedActivity,
		acts.SignalLogEntry{
			PolicyID:      state.PolicyDBID,
			SignalChannel: SignalLoanBalanceUpdated,
			SignalPayload: sigPayload,
			RequestID:     sig.RequestID,
			Status:        domain.SignalStatusProcessed,
			StateBefore:   &stateBefore,
		}).Get(ctx, nil)
	state.Metadata.LoanOutstanding = sig.LoanOutstanding
	_ = workflow.ExecuteActivity(shortActCtx(ctx),
		policyActs.UpdatePolicyMetadataActivity,
		acts.MetadataUpdateParams{
			PolicyID: state.PolicyDBID,
			Updates:  map[string]interface{}{"loan_outstanding": sig.LoanOutstanding},
		}).Get(ctx, nil)
	state.ProcessedSignalIDs[sig.RequestID] = workflow.Now(ctx)
}

func handleCustomerIDMerge(ctx workflow.Context, state *PolicyLifecycleState, sig CustomerIDMergeSignal) {
	// Dedup + audit for customer ID merge. In-memory + DB only. [B6]
	if _, seen := state.ProcessedSignalIDs[sig.RequestID]; seen {
		return
	}
	sigPayload, _ := json.Marshal(sig)
	stateBefore := state.CurrentStatus
	_ = workflow.ExecuteActivity(shortActCtx(ctx),
		policyActs.LogSignalReceivedActivity,
		acts.SignalLogEntry{
			PolicyID:      state.PolicyDBID,
			SignalChannel: SignalCustomerIDMerge,
			SignalPayload: sigPayload,
			RequestID:     sig.RequestID,
			Status:        domain.SignalStatusProcessed,
			StateBefore:   &stateBefore,
		}).Get(ctx, nil)
	state.Metadata.CustomerID = sig.NewCustomerID
	_ = workflow.ExecuteActivity(shortActCtx(ctx),
		policyActs.UpdatePolicyMetadataActivity,
		acts.MetadataUpdateParams{
			PolicyID: state.PolicyDBID,
			Updates:  map[string]interface{}{"customer_id": sig.NewCustomerID},
		}).Get(ctx, nil)
	state.ProcessedSignalIDs[sig.RequestID] = workflow.Now(ctx)
}

func handleDisputeRegistered(ctx workflow.Context, state *PolicyLifecycleState, sig DisputeSignal) {
	// Advisory-only flag — never blocks requests. [BR-PM-113, ADR-003, B6]
	if _, seen := state.ProcessedSignalIDs[sig.RequestID]; seen {
		return
	}
	sigPayload, _ := json.Marshal(sig)
	stateBefore := state.CurrentStatus
	_ = workflow.ExecuteActivity(shortActCtx(ctx),
		policyActs.LogSignalReceivedActivity,
		acts.SignalLogEntry{
			PolicyID:      state.PolicyDBID,
			SignalChannel: SignalDisputeRegistered,
			SignalPayload: sigPayload,
			RequestID:     sig.RequestID,
			Status:        domain.SignalStatusProcessed,
			StateBefore:   &stateBefore,
		}).Get(ctx, nil)
	state.Encumbrances.DisputeFlag = true
	state.ProcessedSignalIDs[sig.RequestID] = workflow.Now(ctx)
}

func handleDisputeResolved(ctx workflow.Context, state *PolicyLifecycleState, sig DisputeSignal) {
	// Clears advisory dispute flag. [BR-PM-113, B6]
	if _, seen := state.ProcessedSignalIDs[sig.RequestID]; seen {
		return
	}
	sigPayload, _ := json.Marshal(sig)
	stateBefore := state.CurrentStatus
	_ = workflow.ExecuteActivity(shortActCtx(ctx),
		policyActs.LogSignalReceivedActivity,
		acts.SignalLogEntry{
			PolicyID:      state.PolicyDBID,
			SignalChannel: SignalDisputeResolved,
			SignalPayload: sigPayload,
			RequestID:     sig.RequestID,
			Status:        domain.SignalStatusProcessed,
			StateBefore:   &stateBefore,
		}).Get(ctx, nil)
	state.Encumbrances.DisputeFlag = false
	state.ProcessedSignalIDs[sig.RequestID] = workflow.Now(ctx)
}

func handleAdminVoid(ctx workflow.Context, state *PolicyLifecycleState, sig AdminVoidSignal) bool {
	// → VOID; cancel all pending requests [BR-PM-073]
	if _, seen := state.ProcessedSignalIDs[sig.RequestID]; seen {
		return false
	}
	// Audit: log signal receipt [D8]
	sigPayload, _ := json.Marshal(sig)
	stateBefore := state.CurrentStatus
	_ = workflow.ExecuteActivity(shortActCtx(ctx),
		policyActs.LogSignalReceivedActivity,
		acts.SignalLogEntry{
			PolicyID:      state.PolicyDBID,
			SignalChannel: SignalAdminVoid,
			SignalPayload: sigPayload,
			RequestID:     sig.RequestID,
			Status:        domain.SignalStatusProcessed,
			StateBefore:   &stateBefore,
		}).Get(ctx, nil)
	for _, pr := range state.PendingRequests {
		_ = workflow.ExecuteActivity(shortActCtx(ctx),
			policyActs.CancelDownstreamWorkflowActivity,
			acts.CancelWorkflowParams{WorkflowID: pr.DownstreamWorkflow}).Get(ctx, nil)
	}
	state.PendingRequests = nil
	// Release DB financial lock before clearing in-memory pointer. [B3, BR-PM-030]
	// Omitting this call left the DB lock row permanently held — orphaned until
	// manual DB intervention. Mirrors the pattern in handleWithdrawal. [Review-Fix-1]
	if state.ActiveLock != nil {
		_ = workflow.ExecuteActivity(shortActCtx(ctx),
			policyActs.ReleaseFinancialLockActivity, state.PolicyDBID).Get(ctx, nil)
		state.ActiveLock = nil
	}
	doTransition(ctx, state, state.CurrentStatus, domain.StatusVoid,
		fmt.Sprintf("admin void by %d: %s", sig.AuthorizedBy, sig.Reason), SignalAdminVoid, sig.RequestID)
	state.ProcessedSignalIDs[sig.RequestID] = workflow.Now(ctx)
	return true
}

func handleVoluntaryPaidUp(ctx workflow.Context, state *PolicyLifecycleState, sig VoluntaryPaidUpSignal) bool {
	// → PAID_UP (value ≥ 10K) or VOID (value < 10K) [BR-PM-060, BR-PM-061]
	const minPaidUpValue = 10000.0
	if _, seen := state.ProcessedSignalIDs[sig.RequestID]; seen {
		return false
	}
	// Audit: log signal receipt [D8]
	sigPayload, _ := json.Marshal(sig)
	stateBefore := state.CurrentStatus
	_ = workflow.ExecuteActivity(shortActCtx(ctx),
		policyActs.LogSignalReceivedActivity,
		acts.SignalLogEntry{
			PolicyID:      state.PolicyDBID,
			SignalChannel: SignalVoluntaryPaidUpRequest,
			SignalPayload: sigPayload,
			RequestID:     sig.RequestID,
			Status:        domain.SignalStatusProcessed,
			StateBefore:   &stateBefore,
		}).Get(ctx, nil)
	eligible, _ := isStateEligible(domain.RequestTypePaidUp, state.CurrentStatus, state.Encumbrances)
	if !eligible {
		state.ProcessedSignalIDs[sig.RequestID] = workflow.Now(ctx)
		return false
	}
	newStatus := domain.StatusPaidUp
	isTerminal := false
	if sig.PaidUpValue < minPaidUpValue {
		newStatus = domain.StatusVoid // [BR-PM-061]
		isTerminal = true
	}
	doTransition(ctx, state, state.CurrentStatus, newStatus,
		fmt.Sprintf("voluntary paid-up; value=%.2f", sig.PaidUpValue), SignalVoluntaryPaidUpRequest, sig.RequestID)
	state.ProcessedSignalIDs[sig.RequestID] = workflow.Now(ctx)
	return isTerminal
}

func handleWithdrawal(ctx workflow.Context, state *PolicyLifecycleState, sig WithdrawalRequestSignal) {
	// Cancel the active downstream workflow + release lock [BR-PM-090]
	if _, seen := state.ProcessedSignalIDs[sig.RequestID]; seen {
		return
	}
	// Audit: log signal receipt [D8]
	sigPayload, _ := json.Marshal(sig)
	stateBefore := state.CurrentStatus
	_ = workflow.ExecuteActivity(shortActCtx(ctx),
		policyActs.LogSignalReceivedActivity,
		acts.SignalLogEntry{
			PolicyID:      state.PolicyDBID,
			SignalChannel: SignalWithdrawalRequest,
			SignalPayload: sigPayload,
			RequestID:     sig.RequestID,
			Status:        domain.SignalStatusProcessed,
			StateBefore:   &stateBefore,
		}).Get(ctx, nil)
	for i, pr := range state.PendingRequests {
		if pr.RequestID == sig.TargetRequestID {
			_ = workflow.ExecuteActivity(shortActCtx(ctx),
				policyActs.CancelDownstreamWorkflowActivity,
				acts.CancelWorkflowParams{WorkflowID: pr.DownstreamWorkflow}).Get(ctx, nil)
			// Revert pre-route status if applicable
			if preRoute := preRouteStatus(pr.RequestType); preRoute != "" && state.CurrentStatus == preRoute {
				doTransition(ctx, state, preRoute, state.PreviousStatus,
					"request withdrawn", SignalWithdrawalRequest, sig.RequestID)
			}
			state.PendingRequests = append(state.PendingRequests[:i], state.PendingRequests[i+1:]...)
			if state.ActiveLock != nil && state.ActiveLock.RequestID == sig.TargetRequestID {
				state.ActiveLock = nil
				// Release DB lock — request withdrawn [Review-Fix-1, BR-PM-030]
				_ = workflow.ExecuteActivity(shortActCtx(ctx),
					policyActs.ReleaseFinancialLockActivity, state.PolicyDBID).Get(ctx, nil)
			}
			break
		}
	}
	state.ProcessedSignalIDs[sig.RequestID] = workflow.Now(ctx)
}

// ─────────────────────────────────────────────────────────────────────────────
// doTransition — records a state transition (activity + in-memory update)
// ─────────────────────────────────────────────────────────────────────────────

func doTransition(ctx workflow.Context, state *PolicyLifecycleState, fromStatus, toStatus, reason, triggeredBy, requestID string) {
	CurrentStatusKey := temporal.NewSearchAttributeKeyKeyword("CurrentStatus")
	state.PreviousStatus = fromStatus
	state.CurrentStatus = toStatus
	state.DisplayStatus = computeDisplayStatus(toStatus, state.Encumbrances)
	state.Version++
	state.LastTransitionAt = workflow.Now(ctx)

	_ = workflow.ExecuteActivity(shortActCtx(ctx),
		policyActs.RecordStateTransitionActivity,
		acts.StateTransitionParams{
			PolicyID:     state.PolicyDBID,
			PolicyNumber: state.PolicyNumber,
			FromStatus:   fromStatus,
			ToStatus:     toStatus,
			Reason:       reason,
			TriggeredBy:  triggeredBy,
			RequestID:    requestID,
			Version:      state.Version,
		}).Get(ctx, nil)

	_ = workflow.UpsertTypedSearchAttributes(
		ctx,
		CurrentStatusKey.ValueSet(toStatus),
	)
}

// computeDisplayStatus mirrors DB compute_display_status() (migration 001).
// Appends encumbrance suffixes in strict left-to-right order:
//
//	_LOAN → _{AssignmentType} → _AML_HOLD → _DISPUTED
//
// NEVER overrides the lifecycle status — AML hold is a suffix, not a replacement. [C5]
func computeDisplayStatus(status string, enc EncumbranceFlags) string {
	s := status
	if enc.HasActiveLoan {
		s += "_LOAN"
	}
	if enc.AssignmentType != "" && enc.AssignmentType != "NONE" {
		s += "_" + enc.AssignmentType
	}
	if enc.AMLHold {
		s += "_AML_HOLD"
	}
	if enc.DisputeFlag {
		s += "_DISPUTED"
	}
	return s
}

// ─────────────────────────────────────────────────────────────────────────────
// handleTerminalCooling — called on every terminal state transition [Constraint 9, §9.5.1]
// ─────────────────────────────────────────────────────────────────────────────

func handleTerminalCooling(ctx workflow.Context, state *PolicyLifecycleState) error {
	// Persist terminal state snapshot for Tier-2 queries [§9.5.1]
	snapshotJSON, _ := json.Marshal(state)
	cooling := coolingDurationFromConfig(state.CachedConfig, state.CurrentStatus) // [B4, Review-Fix-16]
	coolingExpiry := workflow.Now(ctx).Add(cooling)

	_ = workflow.ExecuteActivity(shortActCtx(ctx),
		policyActs.PersistTerminalStateActivity,
		acts.TerminalStateRecord{
			PolicyID:      state.PolicyDBID,
			PolicyNumber:  state.PolicyNumber,
			FinalStatus:   state.CurrentStatus,
			TerminalAt:    workflow.Now(ctx),
			CoolingExpiry: coolingExpiry,
			FinalSnapshot: snapshotJSON,
		}).Get(ctx, nil)

	// ── Cooling period signal channels ──────────────────────────────────────
	premiumPaidCh := workflow.GetSignalChannel(ctx, SignalPremiumPaid)
	reopenCh := workflow.GetSignalChannel(ctx, SignalReopenRequest)
	loanBalanceCh := workflow.GetSignalChannel(ctx, SignalLoanBalanceUpdated)
	coolingTimer := workflow.NewTimer(ctx, cooling)

	for {
		sel := workflow.NewSelector(ctx)
		sel.AddFuture(coolingTimer, func(f workflow.Future) {
			// Cooling expired — mark workflow completed and return
			_ = workflow.ExecuteActivity(shortActCtx(ctx),
				policyActs.MarkWorkflowCompletedActivity,
				acts.MarkCompletedParams{
					PolicyID:    state.PolicyDBID,
					FinalStatus: state.CurrentStatus,
				}).Get(ctx, nil)
		})
		sel.AddReceive(premiumPaidCh, func(c workflow.ReceiveChannel, _ bool) {
			var sig PremiumPaidSignal
			c.Receive(ctx, &sig)
			// Late premium during cooling → trigger refund [§9.5.1]
			_ = workflow.ExecuteActivity(shortActCtx(ctx),
				policyActs.TriggerPremiumRefundActivity,
				acts.RefundRequest{
					PolicyID:     state.PolicyDBID,
					PolicyNumber: state.PolicyNumber,
					Amount:       sig.PremiumAmount,
					Reason:       "late premium received during terminal cooling period",
					RequestID:    sig.RequestID,
				}).Get(ctx, nil)
		})
		sel.AddReceive(loanBalanceCh, func(c workflow.ReceiveChannel, _ bool) {
			var sig LoanBalanceUpdatedSignal
			c.Receive(ctx, &sig)
			state.Metadata.LoanOutstanding = sig.LoanOutstanding
			// Persist updated loan balance to DB during cooling period [Review-Fix-12]
			_ = workflow.ExecuteActivity(shortActCtx(ctx),
				policyActs.UpdatePolicyMetadataActivity,
				acts.MetadataUpdateParams{
					PolicyID: state.PolicyDBID,
					Updates:  map[string]interface{}{"loan_outstanding": sig.LoanOutstanding},
				}).Get(ctx, nil)
		})
		sel.AddReceive(reopenCh, func(c workflow.ReceiveChannel, _ bool) {
			var sig ReopenRequestSignal
			c.Receive(ctx, &sig)
			// Exit cooling + Continue-As-New back to main loop [§9.5.1]
			// [D2] Use doTransition (not raw assignment) so the reopen is persisted
			// to DB via RecordStateTransitionActivity; previously it was in-memory only
			// meaning the policy remained in the terminal status in the DB indefinitely.
			if _, seen := state.ProcessedSignalIDs[sig.RequestID]; !seen {
				doTransition(ctx, state, state.CurrentStatus, state.PreviousStatus,
					fmt.Sprintf("policy reopened: %s (authorized by %d)", sig.ReopenReason, sig.AuthorizedBy),
					SignalReopenRequest, sig.RequestID)
				state.ProcessedSignalIDs[sig.RequestID] = workflow.Now(ctx)
			}
			// Flag handled by returning CAN error below
		})
		sel.Select(ctx)

		// If cooling timer fired, exit normally
		if coolingTimer.IsReady() {
			return nil
		}

		// If reopened, CAN back to main event loop
		if !domain.TerminalStatuses[state.CurrentStatus] {
			return workflow.NewContinueAsNewError(ctx, PolicyLifecycleWorkflow, *state)
		}
	}
}
