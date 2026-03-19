# Review_Report

## Summary
- Status: ACTION_REQUIRED
- Scope: Phase 1 Quote module in [plans/plan.md](plans/plan.md:41) reviewed against [plans/spec.md](plans/spec.md:13), [nbf/swagger/policy_issue_swagger.yaml](nbf/swagger/policy_issue_swagger.yaml:181), [nbf/userjourneys/policy_issue_user_journeys.md](nbf/userjourneys/policy_issue_user_journeys.md:238), [nbf/requirements/policy_issue_requirements.md](nbf/requirements/policy_issue_requirements.md:186), and schema [db/migrations/001_policy_issue_schema.sql](db/migrations/001_policy_issue_schema.sql:145).
- Outcome: Core quote paths are not compliant with the API contract and will fail at runtime due to schema mismatches and missing validation/business rules.

## Critical Fixes (Blockers)
1) Quote persistence violates DB constraints and returns invalid IDs. [`QuoteHandler.CreateQuote()`](handler/quote.go:228) builds a [`domain.Quote`](core/domain/quote.go:37) without computed premium amounts and sends it to [`QuoteRepository.CreateQuote()`](repo/postgres/quote_repository.go:103), which inserts into [`quote`](db/migrations/001_policy_issue_schema.sql:178) that enforces positive premium values. This will fail inserts and leaves the ID unset in [`QuoteCreateResponse`](handler/response/quote.go:95). Fix: persist calculation results, populate premium fields, and use insert-returning to hydrate IDs/timestamps.

2) Calculation identifier is non-UUID but the create API enforces UUID. [`QuoteHandler.CalculateQuote()`](handler/quote.go:101) returns a timestamp-based ID via [`generateCalculationID()`](handler/quote.go:339), while [`QuoteCreateRequest`](handler/quote.go:196) enforces UUIDs (also specified in [plans/spec.md](plans/spec.md:130)). Fix: generate UUIDs and persist calculation state.

3) GST computation ignores state-based CGST/SGST vs IGST rules. The required state input in [`QuoteCalculateRequest`](handler/quote.go:86) is unused in [`QuoteHandler.CalculateQuote()`](handler/quote.go:101), which applies a flat 18% split and never produces IGST/UTGST, violating [plans/spec.md](plans/spec.md:101). Fix: branch by intra/inter-state and compute correct tax components.

4) Premium rate lookup references a table not defined in the schema. [`QuoteRepository.GetPremiumRate()`](repo/postgres/quote_repository.go:76) queries a rate table that is absent around the product/quote definitions in [`001_policy_issue_schema.sql`](db/migrations/001_policy_issue_schema.sql:145). Fix: add the table to DDL or align the query to the actual rate source.

5) Quote reference sequence is missing. [`QuoteRepository.generateQuoteRefNumber()`](repo/postgres/quote_repository.go:177) relies on a sequence not present in the current DDL, where only the shared sequence is defined in [`001_policy_issue_schema.sql`](db/migrations/001_policy_issue_schema.sql:23). Fix: add a dedicated sequence or switch to the existing one and document the format.

6) Quote identifier contract mismatch across API/DB/code. The path parameter [`quote_id`](nbf/swagger/policy_issue_swagger.yaml:300) is a UUID, but [`QuoteConvertRequest`](handler/quote.go:283) treats it as a string reference and [`QuoteHandler.ConvertQuoteToProposal()`](handler/quote.go:289) calls [`QuoteRepository.GetQuoteByRefNumber()`](repo/postgres/quote_repository.go:147) while the DB primary key is BIGINT in [`quote`](db/migrations/001_policy_issue_schema.sql:178). Fix: standardize on a single identifier type and update all layers.

7) Non-Phase1 handlers are registered but unimplemented. [`FxHandler`](bootstrap/bootstrapper.go:26) wires handlers whose methods return nil (e.g., [`ProposalHandler.CreateProposalIndexing()`](handler/proposal.go:40), [`AadhaarHandler.InitiateAadhaarAuth()`](handler/aadhaar.go:28), [`ApprovalHandler.QRApproveProposal()`](handler/approval.go:30), [`PolicyHandler.GetPolicy()`](handler/policy.go:28), [`LookupHandler.GetOccupations()`](handler/lookup.go:29)), exposing broken endpoints. Fix: remove these from [`FxHandler`](bootstrap/bootstrapper.go:26) until implemented or complete the methods.

## Major Issues
- Validation is defined but never enforced. Types like [`QuoteCalculateRequest`](handler/quote.go:86) are consumed directly in [`QuoteHandler.CalculateQuote()`](handler/quote.go:101) without a validation hook, so VR checks in [plans/spec.md](plans/spec.md:101) are bypassed.

- Eligibility checks are incomplete. [`QuoteHandler.CalculateQuote()`](handler/quote.go:101) does not verify product type consistency, allowed payment frequency, policy term limits, or maturity-age constraints despite helper [`Product.IsFrequencyAllowed()`](core/domain/product.go:121) and product config in [`Product`](core/domain/product.go:80). This violates VR-PI-012/013/044 in [plans/spec.md](plans/spec.md:101).

- Quote status model diverges from the user journey. [`QuoteStatus`](core/domain/quote.go:7) omits states present in [nbf/userjourneys/policy_issue_user_journeys.md](nbf/userjourneys/policy_issue_user_journeys.md:325), and [`QuoteHandler.CreateQuote()`](handler/quote.go:228) always sets a generated status. Align enums and transitions.

- Deduplication and workflow triggers are missing. [`QuoteHandler.ConvertQuoteToProposal()`](handler/quote.go:289) does not enforce the dedup rule in [plans/spec.md](plans/spec.md:194) and returns hard-coded data instead of starting the workflow described in [plans/spec.md](plans/spec.md:199).

- Quote response is incomplete. [`QuoteCreateResponse`](handler/response/quote.go:95) expects premium breakdown data, but [`QuoteHandler.CreateQuote()`](handler/quote.go:228) does not populate it. [`QuoteHandler.CalculateQuote()`](handler/quote.go:101) also hard-codes the calculation basis rate to 0 instead of the fetched rate.

- Config key-path mismatch. Repositories use config paths in [`QuoteRepository.GetProducts()`](repo/postgres/quote_repository.go:27) that do not align with the database section in [`configs/config.yaml`](configs/config.yaml:8), likely producing zero timeouts.

- SQL mismatch in product repository. [`ProductRepository.GetAllProducts()`](repo/postgres/product_repository.go:25) filters on a non-existent column name relative to [`product_catalog`](db/migrations/001_policy_issue_schema.sql:145), causing runtime SQL errors.

- Temporal workflow and activities are stubs. [`PolicyIssuanceWorkflow()`](workflows/policy_issuance_workflow.go:61) never executes activities, and the activity methods in [`ProposalActivities`](workflows/activities/proposal_activities.go:8) return nil. This violates the workflow state optimization in [plans/context.md](plans/context.md:57).

- Quote creation bypasses required integrations. The journey mandates DMS storage and event emission in [nbf/userjourneys/policy_issue_user_journeys.md](nbf/userjourneys/policy_issue_user_journeys.md:303), but [`QuoteHandler.CreateQuote()`](handler/quote.go:228) only returns a stub URL.

- Security defaults are unsafe. [`configs/config.yaml`](configs/config.yaml:8) and [`configs/config.yaml`](configs/config.yaml:32) ship default DB and JWT secrets; these should be required environment values.

## Minor Issues / Improvements
- Hard-coded quote validity. [`QuoteHandler.CreateQuote()`](handler/quote.go:228) uses a fixed 30-day expiry instead of the configured quote validity in [`configs/config.yaml`](configs/config.yaml:44).

- JSON scan fragility. [`StringArray.Scan()`](core/domain/product.go:51) and [`IntArray.Scan()`](core/domain/product.go:71) assume byte-slice input; add safe type handling to avoid panics.

- Age calculation and bonus assumptions. [`calculateAge()`](handler/quote.go:318) uses the current time without timezone context, and [`QuoteHandler.CalculateQuote()`](handler/quote.go:101) uses a fixed bonus rate; align with business definitions in [nbf/requirements/policy_issue_requirements.md](nbf/requirements/policy_issue_requirements.md:186).

## Context Updates (Suggestions)
- In [plans/context.md](plans/context.md:11), add an explicit decision on quote identifier strategy (UUID vs BIGINT) and reference-number usage; align [nbf/swagger/policy_issue_swagger.yaml](nbf/swagger/policy_issue_swagger.yaml:300) with [`quote`](db/migrations/001_policy_issue_schema.sql:178).

- In [plans/context.md](plans/context.md:81), document the source table/sequence for premium rates and reference-number generation to prevent DDL/code drift (see [`QuoteRepository.GetPremiumRate()`](repo/postgres/quote_repository.go:76) and [`QuoteRepository.generateQuoteRefNumber()`](repo/postgres/quote_repository.go:177)).

- In [plans/context.md](plans/context.md:35), clarify Phase1 scope for handler registration and Temporal usage; if only Quote APIs are active, register only [`QuoteHandler`](handler/quote.go:19) and omit [`FxTemporal`](bootstrap/bootstrapper.go:63).