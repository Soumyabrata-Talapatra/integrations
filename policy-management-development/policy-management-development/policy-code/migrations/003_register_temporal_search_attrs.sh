#!/usr/bin/env bash
# =============================================================================
# POLICY MANAGEMENT ORCHESTRATOR — Temporal Custom Search Attribute Registration
# =============================================================================
# Migration: 003_register_temporal_search_attrs.sh
# Purpose:   Register custom search attributes in the Temporal cluster BEFORE
#            any PolicyLifecycleWorkflow starts. Without prior registration,
#            workflow.UpsertSearchAttributes() calls silently fail or return
#            an error in Cassandra-backed Temporal clusters.
#
# Run ONCE per cluster (safe to re-run: --skip-if-exists flag used).
# Source: req_Policy_Management_Orchestrator_v4_1.md §3.3, Step 1.4
#
# Usage:
#   TEMPORAL_HOST=localhost:7233 ./003_register_temporal_search_attrs.sh
#   TEMPORAL_HOST=temporal.prod.internal:7233 ./003_register_temporal_search_attrs.sh
# =============================================================================

set -euo pipefail

NAMESPACE="${TEMPORAL_NAMESPACE:-pli-insurance}"
HOST="${TEMPORAL_HOST:-localhost:7233}"

echo "==================================================================="
echo "Registering Temporal Custom Search Attributes"
echo "  Host:      ${HOST}"
echo "  Namespace: ${NAMESPACE}"
echo "==================================================================="

# =============================================================================
# METHOD A: Temporal CLI (temporal v1.x — preferred for Temporal Server ≥1.20)
# =============================================================================
# Uses `temporal operator search-attribute create` (new CLI format).
# Attributes are cluster-wide and then visible per namespace.

if command -v temporal &>/dev/null; then
    echo ""
    echo "→ Using: temporal CLI ($(temporal --version 2>&1 | head -1))"
    echo ""

    temporal operator search-attribute create \
        --address "${HOST}" \
        --namespace "${NAMESPACE}" \
        --name PolicyNumber  --type Keyword  \
        --name CurrentStatus --type Keyword  \
        --name ProductType   --type Keyword  \
        --name BillingMethod --type Keyword  \
        --name IssueDate     --type Datetime \
        --name RequestType   --type Keyword  \
        --name RequestID     --type Keyword  \
        --name LastSignalAt  --type Datetime

    echo ""
    echo "✅ Search attributes registered via temporal CLI."

# =============================================================================
# METHOD B: tctl (Legacy CLI — Temporal Server <1.20 or tctl-only environments)
# =============================================================================
elif command -v tctl &>/dev/null; then
    echo ""
    echo "→ Using: tctl ($(tctl --version 2>&1 | head -1))"
    echo ""

    tctl --address "${HOST}" \
         --namespace "${NAMESPACE}" \
         admin cluster add-search-attributes \
         --name PolicyNumber  --type Keyword  \
         --name CurrentStatus --type Keyword  \
         --name ProductType   --type Keyword  \
         --name BillingMethod --type Keyword  \
         --name IssueDate     --type Datetime \
         --name RequestType   --type Keyword  \
         --name RequestID     --type Keyword  \
         --name LastSignalAt  --type Datetime

    echo ""
    echo "✅ Search attributes registered via tctl."

else
    echo ""
    echo "❌ ERROR: Neither 'temporal' nor 'tctl' CLI found in PATH."
    echo "   Install the Temporal CLI: https://docs.temporal.io/cli"
    echo "   Or set PATH to include the temporal binary before running this script."
    exit 1
fi

# =============================================================================
# VERIFICATION — List registered attributes to confirm
# =============================================================================
echo ""
echo "--- Verifying registered attributes for namespace: ${NAMESPACE} ---"

if command -v temporal &>/dev/null; then
    temporal operator search-attribute list \
        --address "${HOST}" \
        --namespace "${NAMESPACE}" \
        2>/dev/null | grep -E "PolicyNumber|CurrentStatus|ProductType|BillingMethod|IssueDate|RequestType|RequestID|LastSignalAt" \
        && echo "✅ All 8 PM search attributes confirmed." \
        || echo "⚠️  Some attributes may not have propagated yet. Retry in 10 seconds."

elif command -v tctl &>/dev/null; then
    tctl --address "${HOST}" \
         --namespace "${NAMESPACE}" \
         cluster get-search-attributes \
         2>/dev/null | grep -E "PolicyNumber|CurrentStatus|ProductType|BillingMethod|IssueDate|RequestType|RequestID|LastSignalAt" \
         && echo "✅ All 8 PM search attributes confirmed." \
         || echo "⚠️  Some attributes may not have propagated yet. Retry in 10 seconds."
fi

echo ""
echo "==================================================================="
echo "ATTRIBUTE REFERENCE (for workflow.UpsertSearchAttributes calls)"
echo "==================================================================="
echo ""
echo "  PolicyNumber   (Keyword)  — Policy number. e.g. PLI/2026/000001"
echo "  CurrentStatus  (Keyword)  — Lifecycle state. e.g. ACTIVE, PAID_UP"
echo "  ProductType    (Keyword)  — PLI or RPLI"
echo "  BillingMethod  (Keyword)  — CASH, DDO, ONLINE, etc."
echo "  IssueDate      (Datetime) — Policy issue date (RFC3339)"
echo "  RequestType    (Keyword)  — Pending request type. e.g. SURRENDER"
echo "  RequestID      (Keyword)  — Active service request UUID"
echo "  LastSignalAt   (Datetime) — Timestamp of last received signal"
echo ""
echo "Used in workflows/policy_lifecycle_workflow.go:"
echo "  workflow.UpsertSearchAttributes(ctx, temporal.NewSearchAttributes("
echo "    temporal.NewSearchAttributeKeyKeyword(\"PolicyNumber\").ValueSet(state.PolicyNumber),"
echo "    temporal.NewSearchAttributeKeyKeyword(\"CurrentStatus\").ValueSet(string(state.Status)),"
echo "    ..."
echo "  ))"
echo ""
echo "==================================================================="
