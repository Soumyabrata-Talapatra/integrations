# Module Overview

The Policy Bond Delivery and Freelook Cancellation Management Module for
IMS 2.0, aims to streamline tracking and delivery of policy bonds while
automating regulatory-compliant handling of freelook period
cancellations for Postal Life Insurance (PLI) customers.

Integrating with India Post APIs, CRM, CPGRAMS, and booking interface
(recently developed by CEPT), it will confirm deliveries, monitor
eligibility windows, process cancellations, refunds, generate audit
trails and reports---ensuring compliance and operational efficiency for
PLI in a digital era.

[Flow chart depicting the whole process is illustrated as under:]{.mark}

[Policy Approval]{.mark}

[↓]{.mark}

[Trigger Bond Dispatch (via CEPT Booking Interface)]{.mark}

[↓]{.mark}

[Generate Dispatch ID & Link to Policy Number (SP article No)]{.mark}

[↓]{.mark}

[Send to India Post (Physical) / Upload digitally.]{.mark}

[↓]{.mark}

[Track Delivery Status via India Post API / Digital
Acknowledgment]{.mark}

[↓]{.mark}

[Capture Delivery Date]{.mark}

[↓]{.mark}

[Start Freelook Timer (15 Days from Delivery)]{.mark}

[↓]{.mark}

[Notify Policyholder (SMS/Email/Portal)]{.mark}

[↓]{.mark}

[Cancellation Request Received (within Freelook Window)]{.mark}

[↓]{.mark}

[Validate Eligibility (Timer, Policy Status)]{.mark}

[↓]{.mark}

[Calculate Refund (Premium -- Risk Charges -- Stamp Duty)]{.mark}

[↓]{.mark}

[Trigger Refund Workflow (Accounts Module)]{.mark}

[↓]{.mark}

[Update Policy Status to "Cancelled -- Freelook" (Reasons be ascertained
to improvise)]{.mark}

[↓]{.mark}

[Notify Stakeholders (Policyholder, Circle, Accounts, all stake
holders)]{.mark}

[↓]{.mark}

[Log All Actions in Audit Trail]{.mark}

[↓]{.mark}

[Generate Reports (SLA, Refunds, Cancellations)]{.mark}

[↓]{.mark}

[Complaint Closure & Archival]{.mark}

**[Gap Identified. : Presently in PLI we don't have separate free look
period for cancellation of policies however as per industry practices
policy purchased online/physical from agents varies.]{.mark}**

# Scope & Objectives

## Scope

- [Track the dispatch, delivery, and confirmation of both physical
  and]{.mark} [electronic (ePLI) policy bonds. Both to policy holders
  and agents,]{.mark}

- Automate, monitor, and enforce compliance on the freelook cancellation
  period for all PLI policies.

- Handle policy cancellation requests filed during the freelook window,
  assess documentation, and process refunds in compliance with IRDAI and
  pertinent legal guidelines. In addition to this as per the RFP
  prepared by PLI Directorate the Agent incentive recovery, and
  reinsurance reconciliation

- Integrate with core India Post tracking, CRM, CPGRAMS, and accounts
  modules for real-time data synchronization and workflow.

- Offer continuous audit logging for traceability and regulatory audits
  trails.

- Generate comprehensive and actionable reports on delivery and
  cancellation activities.

## Objectives

- E[nhance customer experience and transparency]{.mark} by automating
  dispatch monitoring and customer communications (SMS, email, and
  CPGRAMS feedback).

- Ensure [operational efficiency]{.mark} and accuracy in recognizing and
  executing valid freelook cancellations.

## Functional Requirements

### 1. Dispatch Tracking

The system must fully automate the tracking of policy bond dispatches
using India Post APIs. Each bond delivery event should be logged with
real-time status fetching, from origin (post office or central
production center) to delivery location, leveraging unique tracking
numbers. ePLI bonds---delivered via DigiLocker---must have their
issuance and download events similarly captured and tracked.

[Key Functionalities:]{.mark}

- Generate and assign India Post tracking numbers to all physical policy
  bond dispatches upon issuance.

- Record recipient confirmation data: signature, OTP verification, photo
  evidence, as applicable.

- For digital bonds (ePLI), track issuance, download, and customer
  acknowledgement events.

- Sync delivery status and events with customer profiles in the CRM.

- Notify policyholders upon dispatch and successful delivery via
  SMS/email.

- [Flag and escalate failed or delayed deliveries (e.g., undelivered
  after threshold days) for manual intervention.]{.mark}

**Elaboration:**

Connecting the system with India Post's tracking service helps make the
process more transparent and reduces manual work. It allows the PLI
system to automatically know when a policy bond has been delivered to
the customer. This helps meet regulatory requirements and provides solid
proof in case of any disputes. The system should use either instant
alerts (web hooks) or regular checks (scheduled API calls) to update the
delivery status in the policy record. For digital policies (ePLI), the
moment a customer downloads the bond should be treated as official
delivery. This can be done by linking with DigiLocker and recording the
download event, making it legally equal to receiving a physical bond.

### 2. Delivery Confirmation

On delivery (physical or digital), the system must capture confirmation
from the recipient:

- [For physical bonds]{.mark}: capture delivery via India Post Proof of
  Delivery (POD), including date, time, recipient acknowledgment, and
  digital signature/photo/OTP as per current delivery practices.

- [For ePLI bonds]{.mark}: record download timestamp, DigiLocker
  authentication, and recipient digital signature/acknowledgment if
  implemented.

**[For illustrative purpose]{.underline}**\
Mr. ABC in Delhi receives his policy bond through Speed Post. The India
Post API confirms delivery on **October 16, 2025**, and the system logs
this date.

- The **freelook timer starts** from that date.

- If Mr. ABC tries to cancel on **November 5**, the system will reject
  it, saying the 15-day window is over.

- If Mr. Sharma disputes this, the system shows:

<!-- -->

- Delivery confirmation from India Post

- Timestamped log in IMS 2.0

- SMS sent to Mr. ABC on delivery day

This becomes the entire process **defensible evidence** in case of
complaint.

.

### 3. Freelook Period Monitoring

Freelook monitoring is legally mandated: the freelook window (typically
15 days) begins from the date the policyholder receives the policy
document. The module must:

- Automatically determine the start and end date of the freelook window
  based on confirmed delivery.

- Calculate the remaining window dynamically.

- Provide customer self-service views (SMS/online) on window expiry.

- Automatically reject or route for manual handling any cancellation
  requests submitted after window expiry (with exception-handling for
  disputed deliveries).

**For illustrative purpose:**

**Mr. ABC policy bond was delivered on October 13. IMS 2.0 logged the
timestamp and started his 15-day freelook window. He got reminders on
day 7 and day 12. Later, he claimed the bond went to the wrong address.
India Post confirmed misdelivery, so the system paused the timer(for
this system should be in sync with delivery app). A fresh bond was
delivered on October 18, and the freelook window restarted. Mr. ABC
cancelled on October 30---well within the revised window---and the
system processed his refund with full audit logs.**

### 4. Cancellation Request Handling

Within the freelook period, the system must:

- [Accept cancellation requests]{.mark} via multiple channels: online
  portal, CRM, Post Office counter, CPGRAMS, or email.

- Validate documentation: cancellation request letter/form, policyholder
  ID proof, original policy bond (physical or digital/ePLI), proof of
  delivery/document receipt, and KYC documentation.

- Allow appointment of an authorized messenger with suitable
  documentation (e.g., medical unfitness certificate).

- Log and timestamp all submission and review events.

- Issue system-generated acknowledgements and progress notifications to
  applicants.

**Business illustration:**

**\
Mr. ABC submits a cancellation request on day 10 of his freelook period.
IMS 2.0 checks the delivery timestamp and confirms he's within the
allowed window. The system asks for required documents---original bond,
ID proof, and bank details for NEFT. Mr. ABC thereafter uploads all
documents, [but his ID looks tampered]{.mark}. The system flags the case
for extra review and logs it in the audit trail. After verification, the
refund is processed securely.**

### 5. Refund Processing

On approval of a valid freelook cancellation, the system must:

- Calculate refund amount after deductions (pro-rata risk premium, stamp
  duty, cost of medicals, etc., as per regulatory and POLI
  Administrative Instructions/rules).

- Support multiple modes of refund: NEFT, POSB account credit, crossed
  cheque, etc.( real time in source account)

- Validate refund account details via required documentation (cancelled
  cheque, POSB details).

- Generate sanction letter and payment voucher for accounts module.

- Integrate with accounts module for transaction posting and
  disbursement.

- Update claim/payment register with all financial and reconciliation.

**Elaboration:**\
**Mr. Sharma's cancellation request is approved within the freelook
window. IMS 2.0 calculates his refund and sends it to the Accts team.
One staff enters the refund details (maker), and another verifies and
approves them (checker). The refund is recorded in the accounts module
with a unique transaction ID. This ID links to the main PLI finance
system, ensuring the refund isn't missed or duplicated. All steps are
logged for audit and DoP reconciliation.**

**[Note: This may be added in NAROs Role / & Re]{.mark}conciliation
Policy**

### 6. Audit Trail Generation

All user/system activities at each step must be comprehensively logged:

- Capture user identity, timestamp, action performed, affected policy
  number, old and new values for sensitive data, and free-text
  comments/reason fields where relevant.

- Maintain immutable logs legal requirements for audit trail compliance.

- Track all package tracking queries, cancellation requests, refund
  calculations, approvals, payments, and manual overrides.

- Provide authorized auditors with secure, read-only access to the
  complete action log for every policy transaction.

**Example**

Mr. Sharma's policy bond was delivered on October 13. IMS 2.0 logs the
delivery timestamp, SMS notification, and freelook activation---all in a
tamper-proof audit trail.

On October 20, he submits a cancellation request. The system records the
request time, uploaded documents, and refund approval steps.

Later, it was revealed that his ID proof was digitally altered. Because
of robust audit trail, every action---from delivery to refund---can be
traced, verified, and flagged for investigation. This protects PLI
legally and helps improve future workflows.

## Integration Points

### India Post Tracking APIs

- Fully utilize India Post's parcel tracking APIs (per reference and
  documentation) for all physical dispatches.

### CRM System Integration

- Leverage CRM used by DoP/PLI for all customer data, communication
  records, and case management.

- Sync policy statuses, delivery updates, and cancellation/refund status
  to CRM customer profiles.

- Trigger outbound communication workflows (SMS/email/WhatsApp alerts)
  and record all messages under the relevant customer record.

### CPGRAMS (Centralized Public Grievance Redressal and Monitoring System)

- Integrate with CPGRAMS APIs for seamless grievance registration,
  status update, and resolution tracking for complaints relating to bond
  delivery, freelook period, or cancellations.

- Provide end-to-end feedback loops: update grievance record with
  automated action logs from IMS 2.0.

### Accounts Module Integration

- Direct integration with the DoP/PLI Accounts module for posting and
  tracking all refund payments, accounting for cancellations, and
  financial reconciliation.

- Validate account code mapping for payments/refunds for correct ledger
  assignments, leveraging DoP's account code scheme.

**Reports and analytics**

The module will further provide user-friendly, configurable analytics
dashboards, as well as operational reports, to support both daily
operations and regulatory compliance. Reports /Description and frequency
period has been mentioned hereinunder:-

+----------------------+-------------------------+--------------------+
| Report/Metric        | Description             | Frequency          |
+:=====================+:========================+:===================+
| Dispatch/Delivery    | List of all dispatched  | Daily/Real-time    |
| Register             | bonds, status, delivery |                    |
|                      | dates, recipient        |                    |
+----------------------+-------------------------+--------------------+
| Freelook Window      | List of active policies | Daily              |
| Compliance           | in freelook window,     |                    |
|                      | days elapsed/remaining  |                    |
|                      |                         |                    |
|                      | Detailed report of sync |                    |
|                      | and non sync data       |                    |
+----------------------+-------------------------+--------------------+
| Cancellation Request | Status of all incoming  | Daily/Weekly       |
| Workflow Report      | cancellation requests,  |                    |
|                      | TAT, ageing             |                    |
+----------------------+-------------------------+--------------------+
| Refund Disbursement  | List of processed       | Daily              |
| Ledger               | refunds, amounts,       |                    |
|                      | payment modes, status   |                    |
+----------------------+-------------------------+--------------------+
| Audit Trail Summary  | Chronological log of    | On demand/monthly  |
|                      | all module events       |                    |
|                      | (creation, updates,     |                    |
|                      | reviews)                |                    |
+----------------------+-------------------------+--------------------+
| CPGRAMS Grievance    | Open, closed, pending,  | Daily/Weekly       |
| Status               | escalated complaints    |                    |
|                      | with action history     |                    |
+----------------------+-------------------------+--------------------+
| Delivery Performance | KPIs such as            | Monthly            |
| Analytics            | first-attempt delivery  |                    |
|                      | rate, on-time delivery  |                    |
|                      | %, exception cases      |                    |
+----------------------+-------------------------+--------------------+
| Exception/Manual     | Details and             | On demand or as    |
| Intervention Log     | justification for all   | per requirement    |
|                      | overrides, delays,      | though backed by   |
|                      | rejected/failed cases   | immediate          |
|                      |                         | superior.          |
+----------------------+-------------------------+--------------------+
