# 

# 

# 

# 

# **[SRS/FRS On Death Claim Settlement]{.underline}**

# 

# 

# 

Current Death claim work process in McCamish

**Introduction**

The process for settling death claims in Postal Life Insurance (PLI) and
Rural Postal Life Insurance (RPLI) has been made easier and more
organised with computer systems like McCamish and ECMS. Several people
are involved in the process, including the [person making the
claim]{.mark}, [staff at the (CPC]{.mark}), [document indexers]{.mark},
[approvers]{.mark}, [investigators,]{.mark} and those who handle
payments.

**1. Claim Initiation & Registration**

The process starts when a nominee, legal heir, or assignee submits a
death claim at any post office---Branch Office, Sub Office, or CPC. The
staff checks the original documents and enters the claim into the
system, creating a Claim ID. An acknowledgement is given to the
claimant, and the claim is tracked in the system. If there is no nominee
or the policy details are unclear, extra documents may be needed and a
supervisor will review the claim.

  ----------------------
  Steps Invovled
  ----------------------
  Claim Submission

  Initial Document Check

  Claim Entry

  Acknowledgement

  Claim Tracking

  Additional Documents

  Supervisor Review
  ----------------------

**2. Document Capture & Indexing**

Once registered, the claim enters the document capture phase. The
indexer scans and uploads all supporting documents into ECMS, tagging
them against the Claim ID. Mandatory documents include the [death
certificate]{.mark}, [claim form]{.mark}, [policy bond]{.mark} or
[indemnity, claimant ID/address proof, and bank mandate]{.mark}. For
unnatural deaths, [FIR and postmortem reports are required]{.mark}.

If nomination is absent, succession certificates or legal heir
affidavits are mandatory. The system flags missing documents and places
the claim in a "Pending for Documents" state. Automated notifications
are sent to the claimant, and if documents are not received within 15
days (plus a 7-day grace period), the claim is returned via registered
post.

**3. Investigation & Verification (Conditional)**

[Investigation is triggered under specific conditions]{.mark}:

if death occurs within three years of policy acceptance or revival, The
approving authority nominates an [Inquiry Officer---typically an IP,
ASP, or PRI(P)---]{.mark}to conduct the investigation. The officer
verifies the cause of death, checks hospital and police records, and
ensures no material facts were suppressed during policy issuance. The
investigation report must be submitted within 21 days.

Based on the findings, the claim status is updated to "Clear,"
"Suspect," or "Fraud." Suspect or fraudulent claims are escalated for
manual review and may be rejected with documented evidence.

**4. Claim Calculation & Benefit Computation**

Once documentation and investigation (if applicable) are complete,
[McCamish computes the claim amount]{.mark}. The calculation includes
the base sum assured, accrued bonuses, and any excess premiums.

Deductions include outstanding loans, unpaid premiums, and applicable
taxes. The system performs automated computation but allows manual
override in exceptional cases, such as disputed policy data or
court-directed adjustments.

All calculations are logged with user ID, timestamp, and before/after
values for audit traceability.

**5.Approval Workflow & Decision Points**

The claim then [moves to the approval stage]{.mark}. The approver
reviews all indexed documents, investigation reports, and
system-calculated benefits. Based on financial limits and policy type,
the system routes the claim to the appropriate approver. [For death
claims without investigation, the approval must be completed within 15
days]{.mark}; for those requiring investigation, [the timeline extends
to 45 days. The approver may approve, reject, or send the claim for
reinvestigation]{.mark}. Rejected claims trigger automated rejection
letters with reasons and inform the claimant of appellate rights.

**6 Disbursement & Payment Execution**

Upon approval, the disbursement [officer initiates payment via NEFT,
POSB EFT, or cheque. McCamish integrates with Finacle or IT 2.0]{.mark}
to verify account details and execute payments.. Payment details are
updated in McCamish, and the claim status is marked as "Paid." The
sanction memo, payment acknowledgment, and updated status are logged and
reconciled across systems.

**7 Reopen & Exception Handling**

Claims may be reopened under valid circumstances---court orders, new
evidence, administrative lapses, or claimant appeals. CPC users or
supervisors initiate reopening via the Service Request History screen. A
new request ID is generated, and the claim re-enters the workflow.

.

**8 Communication & Notifications**

Transparent communication is maintained throughout the claim lifecycle.
McCamish sends automated SMS, email, or portal notifications at key
milestones---registration, document status, investigation,
approval/rejection, and payment. All communications are logged and
attached to the claim file. Communication failures are flagged, and
corrective actions are documented for audit and grievance redressal.

**9 Appeal Mechanism**

If a claim is rejected, the claimant may file an appeal [within 90
days]{.mark}. The appeal is submitted in writing---via post, email, or
in person---and may be [routed through CPC]{.mark}. [The appellate
authority is the next higher officer in the approval hierarchy.]{.mark}
The authority may request additional documents or reports and must issue
a reasoned order within 45 days. The decision includes detailed
justification, rulings, and supporting documents. Delays in filing may
be condoned if justified. All appeal actions are logged and linked to
the original claim for traceability.

.

# [Gaps identified in the current system]{.mark}

**1. Fragmented Accountability Across multiple parties.**

- **Issue**: Multiple parties (BO, SO, CPC, Approver, Inquiry Officer,
  Postmaster) are involved, but **no single point of ownership** is
  defined for end-to-end claim resolution.

- **Impact**: Diffused responsibility leads to delays, finger-pointing,

- **Fix**: Introduce a "Claim Case Owner" role at CPC level to track and
  drive each case from registration to closure.

**2. Manual Handling and Paper Dependency**

- **Issue**: Physical document verification, manual indexing, and
  reliance on hard copies (e.g., PRB, indemnity bonds) increase
  turnaround time and risk of loss.

- **Impact**: Breach of SLA, and potential for fraud or misplacement.

- **Fix**: Mandate full digitization of claim documents at source
  (BO/SO) with real-time ECMS sync; eliminate paper-based routing.

**3. Inquiry Process is Loosely Defined**

- **Issue**: Inquiry officer nomination, scope, and timelines are
  vaguely enforced. No system-based tracking of inquiry status or
  escalation.

- **Impact**: Inquiry delays are the biggest contributor to SLA breaches
  in Category II claims.

- **Fix**: Integrate inquiry module into McCamish with auto-assignment,
  TAT monitoring, and escalation triggers after 7/14/21 days.

**4. No Real-Time SLA Monitoring or Alerts**

- **Issue**: SLA timelines (14/47 days) are documented but **not
  enforced by system alerts or dashboards**.

- **Impact**: Delays go unnoticed until complaints arise; no proactive
  governance.

- **Fix**: Implement SLA dashboards with color-coded alerts,
  auto-escalation to higher authorities, and [daily CPC performance
  reports]{.mark}.

**5. Rejection and Appeal Handling is Opaque**

- **Issue**: Rejection letters are sent, but **no structured tracking of
  appeals**, no TAT enforcement, and no audit trail of appellate
  decisions.

- **Impact**: [Claimants lose trust;]{.mark} audit teams can't verify
  fairness or timeliness of appeal disposal.

- **Fix**: Build appeal module in McCamish with:

  - Auto-routing to appellate authority

  - 45-day TAT enforcement

  - Mandatory reasoned order upload

  - Appeal dashboard for monitoring

**6. Disbursement Reconciliation is Not Fully Integrated**

- **Issue**: [Disbursement via NEFT/POSB EFT is not always
  auto-reconciled with McCamish]{.mark}; fallback to cheque is manual.

- **Impact**[: Risk of double payment, payment delay]{.mark}s

- **Fix**: Ensure full integration with Finacle/IT 2.0 for:

  - Real-time status updates

  - Daily reconciliation reports

**7. No Penal Interest Auto-Trigger**

- **Issue**: Penal interest (8% p.a.) is mentioned but **not
  system-enforced**.

- **Impact**: Claimants may be denied rightful compensation for delays;
  audit flags likely.

- **Fix**: Configure McCamish to:

  - Auto-calculate interest post-SLA breach

  - Require approver justification if interest is waived

**8. No Root Cause Analysis for Rejections or Delays**

- **Issue**: Rejections are communicated, but **no structured RCA is
  captured or reviewed**.

- **Impact**: Repeated errors go unaddressed.

- **Fix**: Mandate RCA entry for every rejection/delay, review monthly
  at Divisional/Regional level.

**9. Lack of Real-Time Claim Status Visibility for Claimants**

- **Issue**: [Claimants rely on physical letters or manual follow-up to
  know status.]{.mark}

- **Impact**: Frustration, complaints, and RTI overload.

- **Fix**: Enable claim status tracking via:

  - SMS/Email updates at every stage

  - Online portal/mobile app access with Claim ID

**10. No Audit Trail for Manual Overrides**

- **Issue**: [Manual interventions (e.g., calculation override, document
  waivers) are not always logged with justification]{.mark}.

- **Impact**: High audit risk; potential for misuse or untraceable
  decisions.

- **Fix**: Enforce mandatory remarks and [digital signature for all
  overrides; include in audit logs.]{.mark}

# 

# **✅** **[Required changes in the IMS 2.0]{.underline}** 

**1. Digitize Document Intake at Source**

**Current Pain Point**: Manual verification and physical routing from
BO/SO to CPC\
[**Refinement**:]{.mark}

- [Equip BO/SO counters with scanning devices or mobile apps]{.mark}

- [Enable direct ECMS upload at point of receipt]{.mark}

- [Auto-generate Claim ID and acknowledgment digitally]{.mark}

**2. Introduce a Unified Claim Dashboard**

**Current Pain Point**: Fragmented tracking across McCamish, ECMS, and
manual logs\
[**Refinement**:]{.mark}

- [Build a centralized dashboard showing:]{.mark}

  - [Claim status]{.mark}

  - [SLA countdown]{.mark}

  - [Pending actions]{.mark}

  - [Assigned officers]{.mark}

- [Accessible to CPC, Approvers, and Audit teams]{.mark}

**3. Automate SLA Monitoring and Escalation**

**Current Pain Point**: SLA breaches go unnoticed until complaints
arise\
[**Refinement**:]{.mark}

- [Configure system alerts for each stage (e.g., indexing, inquiry,
  approval)]{.mark}

- [Auto-escalate to next authority if TAT breached]{.mark}

- [Daily SLA compliance report to Divisional Head]{.mark}

**4. Embed Inquiry Workflow in McCamish**

**Current Pain Point**: Inquiry is offline, untracked, and delay-prone\
[**Refinement**:]{.mark}

- [Auto-assign Inquiry Officer based on jurisdiction]{.mark}

- [Digital inquiry form with upload option for FIR, PM report, hospital
  notes]{.mark}

- [Inquiry status visible to CPC and Approver]{.mark}

**5. Enable Claimant Self-Service Portal**

**Current Pain Point**: Claimants rely on manual follow-up or RTI\
[**Refinement**:]{.mark}

- [Launch a web/mobile portal for:]{.mark}

  - [Claim status tracking]{.mark}

  - [Document upload]{.mark}

  - [Appeal submission]{.mark}

  - [SMS/email alerts at each milestone]{.mark}

**6. Simplify Document Requirements with Dynamic Checklists**

**Current Pain Point**: Uniform checklist applied to all cases\
[**Refinement**:]{.mark}

- [System-generated checklist based on:]{.mark}

  - [Type of death (natural/unnatural)]{.mark}

  - [Nomination status]{.mark}

  - [Policy type (MWPA/HUF)]{.mark}

- [Only relevant documents requested]{.mark}

**7. Auto-Trigger Penal Interest Calculation**

**Current Pain Point**: Penal interest is discretionary and often
missed\
[**Refinement**:]{.mark}

- [System calculates interest @ 8% p.a. post SLA breach]{.mark}

- [Approver must justify if interest is waived]{.mark}

- [Interest added to disbursement voucher automatically]{.mark}

**8. Streamline Reopen and Appeal Handling**

**Current Pain Point**: Reopen and appeal are manual, opaque, and
untracked\
[**Refinement**:]{.mark}

- [Reopen button with reason capture and document upload]{.mark}

- [Appeal module with:]{.mark}

  - [Auto-routing to appellate authority]{.mark}

  - [45-day TAT enforcement]{.mark}

  - [Reasoned order upload]{.mark}

  - [Linked audit trail to original claim]{.mark}

**9. Integrate Disbursement with Finacle/IT 2.0**

**Current Pain Point**: EFT failures and cheque fallbacks are manually
reconciled\
[**Refinement**:]{.mark}

- [Real-time payment status sync with banking system]{.mark}

- [Auto-cancellation of failed vouchers]{.mark}

- [Daily reconciliation report for audit and finance teams]{.mark}

**10. Assign Case Owner for End-to-End Tracking**

**Current Pain Point**: No single accountable officer per claim\
[**Refinement**:]{.mark}

- [Assign "Claim Case Owner" at CPC level]{.mark}

- [Responsible for:]{.mark}

  - [SLA adherence]{.mark}

  - [Document completeness]{.mark}

  - [Inquiry coordination]{.mark}

  - [Closure and communication]{.mark}
