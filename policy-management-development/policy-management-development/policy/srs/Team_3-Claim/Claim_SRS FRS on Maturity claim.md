# Table of Contents {#table-of-contents .TOC-Heading}

[1. Introduction [3](#introduction)](#introduction)

[1.1 Purpose [3](#purpose)](#purpose)

[1.2 Scope [3](#scope)](#scope)

[**1.3 Current Maturity Claim Workflow (McCamish)**
[3](#current-maturity-claim-workflow-mccamish)](#current-maturity-claim-workflow-mccamish)

[Step-by-Step Process [3](#step-by-step-process)](#step-by-step-process)

[1. Pending Maturity Report Generation
[3](#pending-maturity-report-generation)](#pending-maturity-report-generation)

[2. Intimation Letter Dispatch
[3](#intimation-letter-dispatch)](#intimation-letter-dispatch)

[3. Claim Submission [3](#claim-submission)](#claim-submission)

[4. Initial Scrutiny [3](#initial-scrutiny)](#initial-scrutiny)

[5. Indexing in McCamish
[4](#indexing-in-mccamish)](#indexing-in-mccamish)

[6. Document Scanning [4](#document-scanning)](#document-scanning)

[7. Data Entry [4](#data-entry)](#data-entry)

[8. Claim Handler Verification
[4](#claim-handler-verification)](#claim-handler-verification)

[9. Approval Workflow [4](#approval-workflow)](#approval-workflow)

[10. Sanction/Rejection Letter
[4](#sanctionrejection-letter)](#sanctionrejection-letter)

[11. Bank Verification [4](#bank-verification)](#bank-verification)

[12. Disbursement [4](#disbursement)](#disbursement)

[13. Voucher Submission [4](#voucher-submission)](#voucher-submission)

[14. Claim Closure [5](#claim-closure)](#claim-closure)

[15. Customer Communication
[5](#customer-communication)](#customer-communication)

[16. Customer Tracking [5](#customer-tracking)](#customer-tracking)

[17. Monitoring & Escalation
[5](#monitoring-escalation)](#monitoring-escalation)

[18. Feedback Collection
[5](#feedback-collection)](#feedback-collection)

[**Detailed Gap Analysis -- Maturity Claim Workflow**
[6](#detailed-gap-analysis-maturity-claim-workflow)](#detailed-gap-analysis-maturity-claim-workflow)

[Rejection Reasons Master for Maturity Claim
[13](#rejection-reasons-master-for-maturity-claim)](#rejection-reasons-master-for-maturity-claim)

[Workflow for the Maturity Process:
[14](#workflow-for-the-maturity-process)](#workflow-for-the-maturity-process)

[Screen Samples for the Process:
[14](#screen-samples-for-the-process)](#screen-samples-for-the-process)

[Test Cases: [18](#test-cases)](#test-cases)

# 1. Introduction

## 1.1 Purpose

The purpose of this document is to define the detailed **Software
Requirements Specification (SRS)** and **Functional Requirements
Specification (FRS)** for the **Maturity claim module** in the upcoming
**IMS 2.0** insurance management system.

## 1.2 Scope

The Maturity Benefit module in IMS 2.0 will provide [digitalized,
automated, and integrated handling of all end-to-end survival benefit
(SB) processes, including report generation, multi-channel
communication, online claim submission]{.mark}, KYC integration,
approval workflows, audit controls, sanction/disbursement, statutory
compliance, and real-time analytics.

## **1.3 [Current Maturity Claim Workflow (McCamish)]{.underline}**

## [Step-by-Step Process]{.underline}

### Pending Maturity Report Generation

- On the first working day of each month, CPC manually generates a
  report listing policies due for maturity in the next two months.

### Intimation Letter Dispatch

- System generates a prefilled intimation letter.

- Sent via registered post with a blank Maturity Claim Form.

- SMS/email sent only if available.

### Claim Submission

- Insurant submits the filled claim form and original documents
  physically at BO/SO/CPC.

### Initial Scrutiny

- BPM/SPM/Postmaster verifies form completeness and attached documents.

- If documents are missing, manual follow-up is initiated.

### Indexing in McCamish

- Claim indexed manually at receiving office.

- Service request number generated.

### Document Scanning

- Scanning done at CPC after physical receipt.

- Documents uploaded and linked to claim.

### Data Entry

- CPC operator enters all claim details manually.

- Claim moves to Maturity Claim Handler inbox.

### Claim Handler Verification

- Handler re-verifies document checklist.

- May raise missing document request again.

- Once complete, forwarded to Approver inbox.

### Approval Workflow

- Approver checks policy details, claim amount, payee info.

- Approves or redirects for correction.

### Sanction/Rejection Letter

- If approved: Sanction Letter (3 copies) generated and dispatched.

- If rejected: Rejection Letter sent via post.

### Bank Verification

- Manual check of passbook or cancelled cheque.

- If unavailable, cheque issued.

### Disbursement

- EFT/NEFT/Cheque processed manually.

- Disbursement details updated in McCamish.

### Voucher Submission

- Vouchers manually prepared and sent to Accounts.

### Claim Closure

- Claim marked as "Paid" or "Rejected" manually.

- No auto-closure or archiving.

### Customer Communication

- Updates sent only via post.

- No SMS/email/WhatsApp notifications.

### Customer Tracking

- Insurant must visit PO or contact CPC for status.

### Monitoring & Escalation

- Admin Office/cpc monitors manually.

- No SLA alerts or escalation triggers.

### Feedback Collection

- No feedback mechanism post-settlement.

##  **[Detailed Gap Analysis -- Maturity Claim Workflow]{.underline}** 

+--------------------+------------------------+--------------------------+
| Process Area       | Current Practice       | Gap Identified (Detailed |
|                    | (McCamish)             | Description)             |
+:===================+:=======================+:=========================+
| Maturity Due       | CPC manually generates | The system does not      |
| Report Generation  | a monthly report       | support automation or    |
|                    | listing policies due   | real-time scheduling.    |
|                    | for maturity.          |                          |
|                    |                        | Reports are monthly,     |
|                    |                        | limiting proactive       |
|                    |                        | management and           |
|                    |                        | visibility into maturing |
|                    |                        | items.                   |
+--------------------+------------------------+--------------------------+
| Customer           | Printed letters are    | Intimation is primarily  |
| Intimation         | sent by post;          | physical and not         |
|                    | SMS/email used         | system-triggered.        |
|                    | inconsistently.        |                          |
|                    |                        | There is no standardized |
|                    |                        | digital communication    |
|                    |                        | workflow.                |
+--------------------+------------------------+--------------------------+
| Claim Initiation   | Staff initiate claims  | Customers cannot         |
|                    | manually after the     | initiate claims          |
|                    | maturity date.         | independently. The       |
|                    |                        | process is entirely      |
|                    |                        | staff-driven, which      |
|                    |                        | delays initiation and    |
|                    |                        | creates dependency.      |
|                    |                        |                          |
|                    |                        | There is no provision    |
|                    |                        | for customer-side        |
|                    |                        | digital initiation or    |
|                    |                        | pre-claim alerts.        |
+--------------------+------------------------+--------------------------+
| Claim Form         | Claim form is handed   | There is no online       |
| Dispatch           | over manually or sent  | access to downloadable   |
|                    | by post.               | or prefilled claim       |
|                    |                        | forms.                   |
|                    |                        |                          |
|                    |                        | Customers must rely on   |
|                    |                        | physical delivery, which |
|                    |                        | causes delays and        |
|                    |                        | increases paper          |
|                    |                        | dependency.              |
+--------------------+------------------------+--------------------------+
| Claim Submission   | Customers submit forms | The system does not      |
|                    | and documents          | support online claim     |
|                    | physically at          | submission. There is no  |
|                    | BO/SO/CPC.             | secure upload facility   |
|                    |                        | or digital               |
|                    |                        | acknowledgment.          |
|                    |                        |                          |
|                    |                        | Customers must travel    |
|                    |                        | and submit hard copies,  |
|                    |                        | which is inefficient and |
|                    |                        | prone to loss.           |
+--------------------+------------------------+--------------------------+
| Initial Scrutiny   | BPM/SPM/Postmaster     | Scrutiny is entirely     |
|                    | manually checks form   | manual, with no system   |
|                    | and documents.         | validation or checklist  |
|                    |                        | enforcement.             |
|                    |                        |                          |
|                    |                        | Errors in form filling   |
|                    |                        | or missing documents are |
|                    |                        | often overlooked,        |
|                    |                        | leading to rework and    |
|                    |                        | delays.                  |
+--------------------+------------------------+--------------------------+
| Missing Document   | Staff follow up        | There is no automated    |
| Handling           | manually with          | reminder system.         |
|                    | customers.             | Communication is ad hoc  |
|                    |                        | and undocumented.        |
|                    |                        |                          |
|                    |                        | Tracking of pending      |
|                    |                        | documents is manual, and |
|                    |                        | follow-ups are           |
|                    |                        | inconsistent across      |
|                    |                        | offices.                 |
+--------------------+------------------------+--------------------------+
| Indexing in        | CPC operator manually  | Indexing is not          |
| McCamish           | indexes the claim and  | automated , Errors in    |
|                    | generates a Claim ID.  | tagging or               |
|                    |                        | misclassification are    |
|                    |                        | common.                  |
|                    |                        |                          |
|                    |                        | There is no              |
|                    |                        | system-driven validation |
|                    |                        | or auto-indexing based   |
|                    |                        | on uploaded content.     |
+--------------------+------------------------+--------------------------+
| Document Scanning  | Scanning is done only  | Digitization is delayed  |
|                    | after physical receipt | until the claim reaches  |
|                    | at CPC.                | CPC.                     |
|                    |                        |                          |
|                    |                        | Documents are not        |
|                    |                        | available in real-time,  |
|                    |                        | and there is no          |
|                    |                        | provision for scanning   |
|                    |                        | at source (BO/SO) or     |
|                    |                        | customer-side upload.    |
+--------------------+------------------------+--------------------------+
| Data Entry         | CPC staff manually     | The process is           |
|                    | enter claim details in | repetitive and           |
|                    | multiple stages.       | error-prone.             |
|                    |                        |                          |
|                    |                        | There is no              |
|                    |                        | auto-population from     |
|                    |                        | scanned documents .      |
+--------------------+------------------------+--------------------------+
| QC Verification    | CPC supervisor         | QC depends entirely on   |
|                    | performs manual        | individual diligence.    |
|                    | quality check.         | There is no              |
|                    |                        | system-enforced          |
|                    |                        | checklist or dual        |
|                    |                        | authentication. Errors   |
|                    |                        | may pass undetected.     |
+--------------------+------------------------+--------------------------+
| Approval Workflow  | Claims are routed      | There is no SLA          |
|                    | manually to the        | enforcement or digital   |
|                    | approving authority.   | workflow.                |
|                    |                        |                          |
|                    |                        | Approvals are delayed    |
|                    |                        | due to lack of system    |
|                    |                        | alerts or escalation     |
|                    |                        | triggers.                |
|                    |                        |                          |
|                    |                        | The process lacks        |
|                    |                        | transparency and         |
|                    |                        | accountability.          |
+--------------------+------------------------+--------------------------+
| Sanction/Rejection | Printed letters are    | Communication is slow    |
| Letters            | sent by post after     | and not digitally        |
|                    | approval.              | tracked.                 |
|                    |                        |                          |
|                    |                        | There is no provision    |
|                    |                        | for auto-generated       |
|                    |                        | digital letters or       |
|                    |                        | portal-based updates.    |
|                    |                        |                          |
|                    |                        | Customers remain         |
|                    |                        | uninformed until         |
|                    |                        | physical delivery.       |
+--------------------+------------------------+--------------------------+
| Bank Verification  | Staff manually verify  | Verification is slow and |
|                    | passbook or cancelled  | prone to errors.         |
|                    | cheque.                |                          |
|                    |                        | There is no integration  |
|                    |                        | with CBS or bank APIs    |
|                    |                        | for instant validation.  |
|                    |                        |                          |
|                    |                        | Manual checks increase   |
|                    |                        | the risk of payment      |
|                    |                        | rejection.               |
+--------------------+------------------------+--------------------------+
| Disbursement       | Payment is entered     | The process is not       |
|                    | manually for           | integrated with Core     |
|                    | NEFT/ECS/Cheque.       | Banking.                 |
|                    |                        |                          |
|                    |                        | Manual entry causes      |
|                    |                        | delays and               |
|                    |                        | reconciliation issues.   |
|                    |                        |                          |
|                    |                        | There is no real-time    |
|                    |                        | status update or payment |
|                    |                        | tracking.                |
+--------------------+------------------------+--------------------------+
| Voucher Submission | Vouchers are manually  | The process is           |
|                    | prepared and submitted | paper-based and lacks    |
|                    | to Accounts.           | digital linkage to       |
|                    |                        | disbursement records.    |
|                    |                        | There is no              |
|                    |                        | auto-generation or       |
|                    |                        | e-submission of          |
|                    |                        | vouchers.                |
+--------------------+------------------------+--------------------------+
| Claim Closure      | Claims are manually    | Closure is not           |
|                    | marked as "Paid" or    | system-triggered. There  |
|                    | "Rejected."            | is no auto-archiving or  |
|                    |                        | digital finalization.    |
+--------------------+------------------------+--------------------------+
| Customer           | Updates are sent only  | There is no structured   |
| Communication      | via post.              | digital communication.   |
|                    |                        | SMS, email, WhatsApp,    |
|                    |                        | and portal notifications |
|                    |                        | are absent.              |
|                    |                        |                          |
|                    |                        | Customers remain         |
|                    |                        | uninformed about claim   |
|                    |                        | progress.                |
+--------------------+------------------------+--------------------------+
| Customer Tracking  | Customers must visit   | There is no online claim |
|                    | PO or contact CPC for  | tracker.                 |
|                    | status.                |                          |
|                    |                        | Customers cannot check   |
|                    |                        | status independently,    |
|                    |                        | leading to frustration   |
|                    |                        | and increased RTI        |
|                    |                        | queries.                 |
+--------------------+------------------------+--------------------------+
| Monitoring by      | Monitoring is done     | There is no real-time    |
| Admin Office       | manually through       | dashboard or SLA         |
|                    | reports.               | tracking.                |
|                    |                        |                          |
|                    |                        | Admin offices cannot     |
|                    |                        | proactively intervene or |
|                    |                        | assess performance.      |
+--------------------+------------------------+--------------------------+
| Escalation of      | Escalation is handled  | There is no automated    |
| Pending Cases      | manually by higher     | alert system for SLA     |
|                    | offices.               | breaches.                |
|                    |                        |                          |
|                    |                        | Delays go unnoticed      |
|                    |                        | until complaints arise.  |
+--------------------+------------------------+--------------------------+
| Customer Feedback  | Feedback is not        | There is no mechanism to |
|                    | collected after claim  | assess service quality   |
|                    | settlement.            | or identify recurring    |
|                    |                        | issues.                  |
|                    |                        |                          |
|                    |                        | Feedback is not used for |
|                    |                        | process improvement.     |
+--------------------+------------------------+--------------------------+
| Integration with   | McCamish operates      | There is no API          |
| Other Systems      | standalone without     | integration with         |
|                    | external interfaces.   | Aadhaar, CBS,            |
|                    |                        | DigiLocker.              |
|                    |                        |                          |
|                    |                        | Data flow is fragmented, |
|                    |                        | and verification is      |
|                    |                        | manual.                  |
+--------------------+------------------------+--------------------------+

**[Revised Workflow -- Maturity Claim (IMS 2.0)]{.underline}**

+---------------+--------------------+---------------------------------+
| Requirement   | Functional         | Details & Specification         |
| ID            | Requirement        |                                 |
|               | Description        |                                 |
+:==============+:===================+:================================+
| FRS-MAT-01    | Auto-Generation of | The system SHALL automatically  |
|               | Maturity Due       | generate the HO-level Maturity  |
|               | Report             | due report daily/weekly. The    |
|               |                    | system SHALL provide a          |
|               |                    | dashboard view of this report,  |
|               |                    | accessible by the CPC/Admin     |
|               |                    | Office and Approving            |
|               |                    | Authorities.                    |
+---------------+--------------------+---------------------------------+
| FRS-MAT-02    | Multi-Channel      | The system SHALL automatically  |
|               |                    | send the intimation notice via  |
|               | Intimation to      | SMS, Email, WhatsApp, and       |
|               |                    | Portal notification. The system |
|               | Policyholder       | SHALL use Registered Post only  |
|               |                    | as a fallback channel. The      |
|               |                    | system SHALL include a secure   |
|               |                    | link within digital intimations |
|               |                    | for the insurant to submit the  |
|               |                    | claim online or download a      |
|               |                    | prefilled claim form.           |
+---------------+--------------------+---------------------------------+
| FRS-MAT-03    | Customer-Initiated | The system SHALL allow the      |
|               |                    | insurant to initiate the claim  |
|               | Claim Submission   | via the Portal or Mobile app ,  |
|               |                    | and upload documents            |
|               |                    |                                 |
|               |                    | (policy bond, ID proof, bank    |
|               |                    | details). The system SHALL      |
|               |                    | integrate with DigiLocker for   |
|               |                    | fetching documents. The system  |
|               |                    | SHALL generate an               |
|               |                    | Auto-acknowledgment with a      |
|               |                    | unique Claim ID and submission  |
|               |                    | timestamp upon                  |
|               |                    |                                 |
|               |                    | successful submission.          |
+---------------+--------------------+---------------------------------+
| FRS-MAT-04    | System-Assisted    | The system SHALL validate       |
|               |                    | uploaded documents against a    |
|               | Initial            | checklist and instantly flag    |
|               | Scrutiny/Checks    | missing or invalid items. The   |
|               |                    | system SHALL send               |
|               |                    | auto-reminders                  |
|               |                    | (SMS/email/WhatsApp) to the     |
|               |                    | customer for outstanding        |
|               |                    |                                 |
|               |                    | documents. CPC staff SHALL      |
|               |                    | verify the completeness of the  |
|               |                    | claim digitally via the system  |
|               |                    | interface.                      |
+---------------+--------------------+---------------------------------+
| FRS-MAT-05    | Auto-Indexing and  | The system SHALL automatically  |
|               |                    | index the claim with metadata.  |
|               | Document Sync      | Documents SHALL be synced to    |
|               |                    | ECMS (Electronic Content        |
|               |                    | Management System) in real-time |
|               |                    | and linked to the Claim ID and  |
|               |                    | policy record.                  |
+---------------+--------------------+---------------------------------+
| FRS-MAT-06    | Auto-Populated     | The system SHALL extract key    |
|               |                    | fields from uploaded documents  |
|               | Data Entry         | to auto-populate claim data     |
|               |                    | fields. The CPC                 |
|               |                    |                                 |
|               |                    | operator SHALL review and       |
|               |                    | confirm the extracted data.     |
|               |                    | This reduces manual entry and   |
|               |                    | audit risk.                     |
+---------------+--------------------+---------------------------------+
| FRS-MAT-07    | QC Verification    | The Supervisor SHALL perform QC |
|               |                    | digitally using a               |
|               | Checklist          | system-enforced checklist. Dual |
|               |                    | authentication SHALL            |
|               |                    |                                 |
|               |                    | be required for override or     |
|               |                    | waiver. All actions SHALL be    |
|               |                    | logged with a user ID and       |
|               |                    | timestamp.                      |
+---------------+--------------------+---------------------------------+
| FRS-MAT-08    | Approval Workflow  | The Approving Authority SHALL   |
|               |                    | review the claim, documents,    |
|               | and SLA            | and system-calculated maturity  |
|               |                    | amount via                      |
|               | Enforcement        |                                 |
|               |                    | a dedicated digital routing.    |
|               |                    | The system SHALL display an SLA |
|               |                    | countdown (e.g., 7-day window). |
|               |                    | The Approver SHALL be able to   |
|               |                    | Approve or redirect with        |
|               |                    | remarks.                        |
+---------------+--------------------+---------------------------------+
| FRS-MAT-09    | Auto-Generated     | The system SHALL auto-generate  |
|               |                    | the Sanction or Rejection       |
|               | Sanction/Rejection | Letter. Letters SHALL be sent   |
|               |                    | via email, WhatsApp, and the    |
|               | Communication      | Customer Portal. If rejected,   |
|               |                    | the letter SHALL include the    |
|               |                    | reason and appeal link sent     |
|               |                    | digitally. All communications   |
|               |                    | SHALL be timestamped and        |
|               |                    | archived.                       |
+---------------+--------------------+---------------------------------+
| FRS-MAT-10    | Bank Account       | The system SHALL verify bank    |
|               |                    | details via CBS/PFMS API. If    |
|               | Validation         | invalid, the customer SHALL be  |
|               |                    | prompted to re-submit.          |
|               | (API-based)        |                                 |
+---------------+--------------------+---------------------------------+
| FRS-MAT-11    | Disbursement       | The system SHALL be integrated  |
|               |                    | with Core Banking. The system   |
|               | Execution          | SHALL process payment using     |
|               |                    | Auto NEFT/IMPS (preferred). The |
|               |                    | disbursement status SHALL be    |
|               |                    | updated in the system in        |
|               |                    | real-time.                      |
+---------------+--------------------+---------------------------------+
| FRS-MAT-12    | Voucher            | The voucher SHALL be            |
|               |                    | auto-generated                  |
|               | Generation and     | post-disbursement. The voucher  |
|               |                    | SHALL be submitted digitally to |
|               | Submission         | the Accounts section and linked |
|               |                    | to the claim record for audit   |
|               |                    | trail.                          |
+---------------+--------------------+---------------------------------+
| FRS-MAT-13    | Claim Closure and  | The claim SHALL be marked as    |
|               |                    | \"Paid\" or \"Rejected\"        |
|               | Archiving          | automatically. The case file    |
|               |                    | SHALL be archived digitally     |
|               |                    | with a recorded closure         |
|               |                    | timestamp and user ID.          |
+---------------+--------------------+---------------------------------+
| FRS-MAT-14    | Customer           | The system SHALL send an auto   |
|               |                    | message to the customer         |
|               | Feedback           | post-claim settlement to        |
|               |                    | collect feedback. The           |
|               | Collection         |                                 |
|               |                    | feedback SHALL be stored for    |
|               |                    | service quality analytics.      |
+---------------+--------------------+---------------------------------+
| FRS-MAT-15    | Real-Time          | The Admin dashboard SHALL show  |
|               |                    | Pending claims, SLA countdown,  |
|               | Monitoring &       | and Escalated cases.            |
|               |                    | Auto-escalation SHALL be        |
|               | Escalation         | triggered if the SLA is         |
|               |                    | breached.                       |
+---------------+--------------------+---------------------------------+
| FRS-MAT-16    | Customer Claim     | The system SHALL provide        |
|               |                    | online/mobile access for the    |
|               | Tracker            | insurant to track their claim   |
|               |                    | status via the Portal or Mobile |
|               |                    | app. The tracker SHALL provide  |
|               |                    | stage-wise updates with         |
|               |                    | timestamps.                     |
+---------------+--------------------+---------------------------------+
| FRS-MAT-17    | Bank Account       | The system should automatically |
|               | Verification for   | verify the policyholder\'s bank |
|               | Maturity Payout    | account details using an        |
|               |                    | external API so that maturity   |
|               |                    | payouts are securely and        |
|               |                    | accurately credited to the      |
|               |                    | correct, verified account.      |
|               |                    |                                 |
|               |                    | **Functional Requirement        |
|               |                    | (FR):**                         |
|               |                    |                                 |
|               |                    | 1.  The Bank Module SHALL       |
|               |                    |     trigger an API-based        |
|               |                    |     validation check upon       |
|               |                    |     submission of the           |
|               |                    |     policyholder\'s bank        |
|               |                    |     account details (e.g.,      |
|               |                    |     Account Number, IFSC/Sort   |
|               |                    |     Code, etc.).                |
|               |                    |                                 |
|               |                    | 2.  The API response SHALL be   |
|               |                    |     processed to determine the  |
|               |                    |     validity of the bank        |
|               |                    |     account details (e.g.,      |
|               |                    |     account existence, name     |
|               |                    |     matching, active status).   |
|               |                    |                                 |
|               |                    | 3.  If the API validation is    |
|               |                    |     successful, the system      |
|               |                    |     SHALL store the bank        |
|               |                    |     account as Verified and     |
|               |                    |     proceed with the maturity   |
|               |                    |     process.                    |
|               |                    |                                 |
|               |                    | 4.  If the API validation fails |
|               |                    |     (e.g., Invalid Account      |
|               |                    |     Number, Account Not Found,  |
|               |                    |     Name Mismatch), the system  |
|               |                    |     SHALL prevent final         |
|               |                    |     submission and display a    |
|               |                    |     clear, actionable prompt to |
|               |                    |     the user (policyholder or   |
|               |                    |     staff) requesting           |
|               |                    |     correction of the entered   |
|               |                    |     bank account details.       |
+---------------+--------------------+---------------------------------+

### Rejection Reasons Master for Maturity Claim 

**A. Policy-Related Reasons**

  -----------------------------------------------------------------------
  Code              Rejection Reason
  ----------------- -----------------------------------------------------
  RJ-P-01           Policy number is invalid or does not exist

  RJ-P-02           Policy is not active on the date of maturity

  RJ-P-03           Maturity Claim has already been paid for the policy

  RJ-P-04           Policy terminated due to forfeiture/surrender prior
                    to maturity
  -----------------------------------------------------------------------

**B. Claimant & Eligibility-Related Reasons**

  -----------------------------------------------------------------------
  Code              Rejection Reason
  ----------------- -----------------------------------------------------
  RJ-E-01           Claimant details do not match policy records

  RJ-E-02           Identity of policyholder/claimant could not be
                    established

  RJ-E-03           Claim submitted by unauthorized person

  RJ-E-04           Nominee/legal heir details not valid

  RJ-E-05           Multiple claimants with unresolved entitlement
                    dispute
  -----------------------------------------------------------------------

**C. Document-Related Reasons**

  -----------------------------------------------------------------------
  Code              Rejection Reason
  ----------------- -----------------------------------------------------
  RJ-D-01           Mandatory documents not submitted

  RJ-D-02           Submitted documents found forged or suspicious

  RJ-D-03           Mismatch between physical documents and digital
                    records

  RJ-D-04           Policy bond not submitted or invalid

  RJ-D-05           Identity proof/address proof invalid or expired
  -----------------------------------------------------------------------

**D. Bank & Payment-Related Reasons**

  -----------------------------------------------------------------------
  Code              Rejection Reason
  ----------------- -----------------------------------------------------
  RJ-B-01           Bank account details invalid or verification failed

  RJ-B-02           IFSC / account number mismatch

  RJ-B-03           Repeated failure of payment due to incorrect bank
                    details
  -----------------------------------------------------------------------

### Workflow for the Maturity Process:

![](media/image1.emf)

### Screen Samples for the Process:

1.  Pending Policy Maturity Report

![A screenshot of a computer AI-generated content may be
incorrect.](media/image2.png){width="6.268055555555556in"
height="1.4861111111111112in"}

2.  Notification/Intimation Letter

![A letter of insurance with text AI-generated content may be
incorrect.](media/image3.png){width="4.273408792650919in"
height="5.9968110236220475in"}

3.  Maturity Request Indexing:

![A screenshot of a computer AI-generated content may be
incorrect.](media/image4.png){width="6.268055555555556in"
height="1.9979166666666666in"}

4.  Data Entry Screen:

![A screenshot of a computer screen AI-generated content may be
incorrect.](media/image5.png){width="6.268055555555556in"
height="2.6819444444444445in"}

![A screenshot of a computer AI-generated content may be
incorrect.](media/image6.png){width="6.268055555555556in"
height="2.685416666666667in"}

![A screenshot of a computer AI-generated content may be
incorrect.](media/image7.png){width="6.268055555555556in"
height="2.7270833333333333in"}

![A screenshot of a computer AI-generated content may be
incorrect.](media/image8.png){width="6.268055555555556in"
height="2.727777777777778in"}

5.  Missing Documents Page

![A screenshot of a computer screen AI-generated content may be
incorrect.](media/image9.png){width="6.268055555555556in"
height="2.7402777777777776in"}

6.  Approval Letter

> ![A document with a receipt AI-generated content may be
> incorrect.](media/image10.png){width="5.577778871391076in"
> height="9.136362642169729in"}

7.  Disbursement Update

![A screenshot of a computer AI-generated content may be
incorrect.](media/image11.png){width="6.268055555555556in"
height="2.6326388888888888in"}

### Test Cases:

**POSITIVE TEST SCENARIOS (User Perspective)**

P-01: Successful receipt of Maturity due intimation

- User receives Maturity due intimation via SMS / Email / WhatsApp /
  Portal notification.

- Message contains correct policy details and a secure claim submission
  link.

P-02: Secure claim link opens correctly

- User clicks the secure link received in intimation.

- User is redirected to official Customer Portal / Mobile App.

- Policy and customer details are auto-fetched and displayed correctly.

P-03: Successful online maturity claim submission

- User fills maturity claim form correctly.

- All mandatory documents are uploaded.

- Claim is submitted successfully without errors.

P-04: Auto-acknowledgement generation

- System generates a unique Claim ID upon submission.

- Acknowledgement is shown on screen and sent digitally.

P-05: DigiLocker integration works successfully

- User fetches documents from DigiLocker successfully.

- Documents are attached correctly to the claim.

P-06: Auto-reminders received for pending documents

- System sends reminders for missing documents.

- Reminder clearly mentions pending documents.

P-07: Claim status tracking works correctly

- User tracks claim through Portal/App.

- Stage-wise status updates are visible.

P-08: Sanction letter received after approval

- User receives sanction letter digitally.

- Letter contains correct claim details.

P-09: Successful payment of maturity amount

- Amount credited via NEFT/IMPS.

- Status updated as Paid in tracker.

P-10: Feedback submission after claim settlement

- User receives feedback request.

- Feedback submitted successfully.

**NEGATIVE TEST SCENARIOS (User Perspective)**

N-01: Secure claim link does not open

- Link is invalid or expired.

- Error message is displayed.

N-02: Missing mandatory documents

- Submission attempted without required documents.

- System blocks submission.

N-03: Invalid document upload

- Unsupported file format or size.

- Upload error shown.

N-04: DigiLocker fetch failure

- Consent denied or technical issue.

- Manual upload option available.

N-05: Incorrect bank details entered

- Wrong account number or IFSC.

- Bank verification fails.

N-06: Claim rejected by authority

- Claim rejected due to discrepancy or ineligibility.

- Rejection letter sent with reason.

N-07: Tracker not accessible / delayed update

- Portal/App unavailable.

- Status not updated immediately.

N-08: Payment failure

- Disbursement fails due to technical/banking issue.

- User informed payment failed or pending.

N-09: Feedback link not opening

- Feedback page fails to load.

- Retry/error message shown.
