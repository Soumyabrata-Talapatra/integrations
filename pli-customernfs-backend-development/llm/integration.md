# Customer NFS - Policy Management Integration

## Overview

This integration enables **Customer NFS** (Name/Address/Mobile/Email change service) to notify **Policy Management** whenever a customer change request completes or is rejected. The notification flows one-way: Customer NFS → Policy Management via Temporal signals.

Policy Management's `PolicyLifecycleWorkflow` (`plw-{policy_number}`) receives these signals and updates policy metadata accordingly. This is **not** a policy state change — no financial locks or status transitions are involved.

## Architecture

```
Customer → NFS HTTP API → NFS Workflow (processes change)
                                │
                                │ On COMPLETED / REJECTED:
                                │
                                ├─► Activity: LookupAffectedPolicies(customerID)
                                │   Returns: ["PLI/2026/000001", ...]
                                │
                                └─► Activity: NotifyPolicyManagement
                                    For EACH policy:
                                      SignalExternalWorkflow(
                                          "plw-{policy_number}",
                                          "",
                                          "customer-nfr-completed",
                                          OperationCompletedSignal{...}
                                      )
                                            │
                                            ▼
                              PolicyLifecycleWorkflow (plw-{policy_number})
                                            │
                                            ├─► handleCustomerNFRCompleted()
                                            ├─► Dedup via ProcessedSignalIDs
                                            ├─► Log signal to policy_signal_log
                                            ├─► Update policy metadata (allow-listed keys only)
                                            └─► No state transition
```

## Signal Contract

**Signal Channel**: `customer-nfr-completed`

**Payload** (`OperationCompletedSignal`):

| Field             | Type              | Description                                              |
|-------------------|-------------------|----------------------------------------------------------|
| `request_id`      | `string`          | NFS service request UUID                                 |
| `request_type`    | `string`          | `ADDRESS_CHANGE`, `NAME_CHANGE`, `MOBILE_CHANGE`, `EMAIL_CHANGE` |
| `outcome`         | `string`          | `APPROVED` or `REJECTED`                                 |
| `state_transition`| `string`          | Empty (no policy state change)                           |
| `outcome_payload` | `json.RawMessage` | Changed field values (type-specific)                     |
| `completed_at`    | `time.Time`       | Completion timestamp                                     |

## Changes Made

### Customer NFS Service (`pli-customernfs-backend-development/`)

#### CustomerID Migration (string/UUID → int64/BIGINT)

All occurrences of `CustomerID string` changed to `CustomerID int64` across:
- Domain models: `service_request.go`, `address_change_detail.go`, `name_change_detail.go`
- Activity types: `temporal/activities/types.go` (9 input/output types)
- Workflow inputs: `address_change_workflow.go`, `name_change_workflow.go`
- Request DTOs: `handler/request.go`
- Response DTOs: `handler/response/cpc.go`, `handler/response/status.go`
- DB schema: `migrations/001_create_nfs_schema.sql` (`nfs.customer_id` domain → `BIGINT`)
- Repo functions: `service_request.go`, `address_change.go`
- Handlers: `address_change.go`, `name_change.go`, `status.go`, `lookup.go` (format verbs `%s` → `%d`)

#### New Files

| File | Purpose |
|------|---------|
| `temporal/workflows/pm_types.go` | `OperationCompletedSignal` struct and `SignalCustomerNFRCompleted` constant |
| `temporal/activities/pm_notification_activities.go` | `LookupAffectedPolicies` and `NotifyPolicyManagement` activities |
| `core/port/policy_lookup.go` | `PolicyLookupRepository` interface |
| `repo/postgres/policy_lookup.go` | PostgreSQL implementation for policy lookup |
| `core/domain/mobile_change_detail.go` | `MobileChangeDetail` and `MobileVersionHistory` domain models |
| `core/domain/email_change_detail.go` | `EmailChangeDetail` and `EmailVersionHistory` domain models |
| `temporal/workflows/mobile_change_workflow.go` | WF-NFS-006: Mobile number change workflow |
| `temporal/workflows/email_change_workflow.go` | WF-NFS-007: Email change workflow |
| `handler/mobile_change.go` | HTTP handler for mobile change initiation |
| `handler/email_change.go` | HTTP handler for email change initiation |
| `handler/response/mobile_change.go` | Mobile change response DTOs |
| `handler/response/email_change.go` | Email change response DTOs |
| `migrations/002_add_mobile_email_tables.sql` | DB tables for mobile/email changes + policy_customer_mapping |

#### Modified Workflows (PM Notification Added)

At each terminal point (COMPLETED/REJECTED), a `NotifyPolicyManagement` activity call was added:

- **WF-NFS-001** (Aadhaar Address Change): After COMPLETED status
- **WF-NFS-002** (Manual Address Change): After APPROVED→COMPLETED and REJECTED
- **WF-NFS-003** (Aadhaar Name Change): After COMPLETED status
- **WF-NFS-004** (Manual Name Change): After APPROVED→COMPLETED and REJECTED

### Policy Management Service (`policy-management/`)

#### New Signal Handler

| File | Change |
|------|--------|
| `workflows/signals.go` | Added `SignalCustomerNFRCompleted` constant |
| `workflows/policy_lifecycle_workflow.go` | Added `customer-nfr-completed` signal channel registration |
| `workflows/policy_lifecycle_workflow.go` | Added `sel.AddReceive(customerNfrCompletedCh, ...)` in main select loop |
| `workflows/policy_lifecycle_workflow.go` | Added `handleCustomerNFRCompleted()` function |
| `workflows/policy_lifecycle_workflow.go` | Added `customerNFRMetadataAllowList` (security allow-list) |
| `workflows/policy_lifecycle_workflow.go` | Added `filterCustomerNFRPayload()` function |
| `core/domain/service_request.go` | Added `RequestTypeMobileChange` and `RequestTypeEmailChange` constants |

## New HTTP Endpoints

| Method | Path | Handler | Description |
|--------|------|---------|-------------|
| `POST` | `/nfs/mobile-change/initiate` | `MobileChangeHandler.InitiateMobileChange` | Start mobile number change |
| `POST` | `/nfs/email-change/initiate` | `EmailChangeHandler.InitiateEmailChange` | Start email change |

## Database Migrations

### Migration 001 (Modified)
- `nfs.customer_id` domain changed from `UUID` to `BIGINT`

### Migration 002 (New)
Creates the following tables:
- `nfs.policy_customer_mapping` — Maps customers to their active policies (used by `LookupAffectedPolicies`)
- `nfs.mobile_change_detail` — Stores mobile change request details
- `nfs.mobile_version_history` — Mobile number version history
- `nfs.email_change_detail` — Stores email change request details
- `nfs.email_version_history` — Email version history

## Tests

### Customer NFS Tests

| Test File | What It Tests |
|-----------|---------------|
| `temporal/workflows/pm_types_test.go` | `OperationCompletedSignal` JSON round-trip, signal constant |
| `temporal/activities/pm_notification_test.go` | `NotifyPMInput`/`NotifyPMResult` JSON serialization |
| `temporal/workflows/mobile_change_workflow_test.go` | `MobileChangeWorkflowInput` JSON round-trip |
| `temporal/workflows/email_change_workflow_test.go` | `EmailChangeWorkflowInput` JSON round-trip |
| `handler/request_mobile_email_test.go` | Mobile/email request DTO validation tags |
| `handler/response/mobile_email_response_test.go` | Response constructor functions |
| `core/domain/mobile_email_domain_test.go` | Domain struct field verification |
| `temporal/workflows/address_change_notify_test.go` | `mustMarshalJSON` helper function |

### Policy Management Tests

| Test File | What It Tests |
|-----------|---------------|
| `workflows/customer_nfr_signal_test.go` | Signal constant, `OperationCompletedSignal` JSON serialization |
| `workflows/customer_nfr_handler_test.go` | `customerNFRMetadataAllowList` entries, `filterCustomerNFRPayload` behavior |
| `core/domain/service_request_nfr_test.go` | `MOBILE_CHANGE`/`EMAIL_CHANGE` constants, `DownstreamTaskQueueForType` |

## Key Design Decisions

1. **Separate signal channel** (`customer-nfr-completed`): Keeps externally-initiated (Customer NFS) and internally-initiated (PM) NFR flows isolated with different handlers and trust boundaries.

2. **Activity-based notification**: The notification involves DB lookup + multiple signal calls. An activity provides retry semantics and timeout handling.

3. **Fan-out to all policies**: A customer may hold multiple policies. Each `plw-{policy_number}` is signaled individually.

4. **Fire-and-forget**: NFS workflow completion is not blocked by PM notification failure.

5. **Empty RunID**: Using `""` targets the latest workflow run, handling PM's Continue-As-New transparently.

6. **Metadata allow-list**: `customerNFRMetadataAllowList` prevents rogue payloads from overwriting protected policy columns.

7. **No PendingRequest**: Unlike PM-initiated NFRs, customer NFS changes don't create a `PendingRequest` in PLW.

## What To Do Next

1. **Register new activities**: Add `PMNotificationActivities` to the Temporal worker registration in the NFS service startup code.
2. **Register new workflows**: Register `MobileChangeWorkflow` and `EmailChangeWorkflow` in the NFS Temporal worker.
3. **Wire HTTP routes**: Add routes for `/nfs/mobile-change/initiate` and `/nfs/email-change/initiate` in the router setup.
4. **Run migrations**: Apply migration 002 to create the new tables.
5. **Populate `policy_customer_mapping`**: Set up the data pipeline or initial load for the customer-to-policy mapping table.
6. **Policy lookup decision**: Choose between direct DB read, PM API call, or dedicated mapping service for `LookupAffectedPolicies` (see `docs/integration_approach_customer_nfs_policy_management.md` section 10).
7. **Integration testing**: Test the full flow end-to-end with a running Temporal server.

## Reference

- Full approach document: `docs/integration_approach_customer_nfs_policy_management.md`