# Table of Contents {#table-of-contents .TOC-Heading}

[1. Introduction [3](#introduction)](#introduction)

[1.1 Purpose [3](#purpose)](#purpose)

[1.2 Scope [3](#scope)](#scope)

[**1.3 Current Survival Benefit Workflow (McCamish)**
[3](#current-survival-benefit-workflow-mccamish)](#current-survival-benefit-workflow-mccamish)

[Stage 1: Report Generation
[3](#stage-1-report-generation)](#stage-1-report-generation)

[**Stage 2: Intimation to Policyholder**
[3](#stage-2-intimation-to-policyholder)](#stage-2-intimation-to-policyholder)

[Stage 3: Claim Submission
[4](#stage-3-claim-submission)](#stage-3-claim-submission)

[**Stage 4: Initial Scrutiny**
[4](#stage-4-initial-scrutiny)](#stage-4-initial-scrutiny)

[Stage 5: Indexing in McCamish
[4](#stage-5-indexing-in-mccamish)](#stage-5-indexing-in-mccamish)

[Stage 6: Document Scanning
[4](#stage-6-document-scanning)](#stage-6-document-scanning)

[Stage 7: Data Entry [4](#stage-7-data-entry)](#stage-7-data-entry)

[Stage 8: Quality Check (QC)
[5](#stage-8-quality-check-qc)](#stage-8-quality-check-qc)

[Stage 9: Approval Workflow
[5](#stage-9-approval-workflow)](#stage-9-approval-workflow)

[Stage 10: Communication of Decision
[5](#stage-10-communication-of-decision)](#stage-10-communication-of-decision)

[Stage 11: Bank Verification
[5](#stage-11-bank-verification)](#stage-11-bank-verification)

[Stage 12: Disbursement
[6](#stage-12-disbursement)](#stage-12-disbursement)

[Stage 13: Voucher Submission
[6](#stage-13-voucher-submission)](#stage-13-voucher-submission)

[Stage 14: End of Process
[6](#stage-14-end-of-process)](#stage-14-end-of-process)

[1.5 Table on Gap Identification
[7](#table-on-gap-identification)](#table-on-gap-identification)

[1.6 Revised Survival Benefit Workflow -- IMS 2.0
[8](#revised-survival-benefit-workflow-ims-2.0)](#revised-survival-benefit-workflow-ims-2.0)

[1.7 Process Flow Diagram
[9](#process-flow-diagram)](#process-flow-diagram)

[1.8 Rejection Reason Master
[9](#rejection-reason-master)](#rejection-reason-master)

[1.9 Survival Benefit Screens
[10](#survival-benefit-screens)](#survival-benefit-screens)

[1.10 Test Cases [14](#test-cases)](#test-cases)

[1.11 Attachments [16](#attachments)](#attachments)

# 1. Introduction

## 1.1 Purpose

The purpose of this document is to define the detailed **Software
Requirements Specification (SRS)** and **Functional Requirements
Specification (FRS)** for the **Survival Benefit module** in the
upcoming **IMS 2.0** insurance management system.

## 1.2 Scope

The Survival Benefit module in IMS 2.0 will provide [digitalized,
automated, and integrated handling of all end-to-end survival benefit
(SB) processes, including report generation, multi-channel
communication, online claim submission]{.mark}, KYC integration,
approval workflows, audit controls, sanction/disbursement, statutory
compliance, and real-time analytics.

## **1.3 [Current Survival Benefit Workflow (McCamish)]{.underline}**

### Stage 1: Report Generation

a.  On the **first working day of each month**, the CPC generates the
    **HO‑level Detailed Maturity/Survival Benefit Due Report**.

b.  This report lists all policies where Survival benefit is due in the
    **next two months**.

c.  The report is generated manually and requires CPC staff
    intervention.

### 

### **Stage 2: Intimation to Policyholder**

a.  Based on the report, the system generates [an **Intimation
    Letter**]{.mark} pre‑filled [with policy details]{.mark}.

b.  The letter is printed and sent to the insurant **by registered
    post**.

c.  In some cases, SMS/email is also sent, but this is not mandatory or
    consistent.

d.  The intimation letter is accompanied by a **blank Survival Benefit
    Claim Form**.

### Stage 3: Claim Submission

a.  The insurant fills the claim form manually and [attaches required
    documents (policy bond, ID proof, bank details, etc.).]{.mark}

b.  The claim is submitted physically at the **Branch Office (BO), Sub
    Office (SO), or CPC**.

c.  The receiving official (BPM/SPM/Postmaster) acknowledges receipt.

### 

### **Stage 4: Initial Scrutiny**

a.  The receiving official checks whether the claim form is filled
    correctly.

b.  Attached documents [are verified for completeness (policy bond, ID
    proof, bank details).]{.mark}

c.  If documents are missing, the official requests them manually from
    the insurant.

d.  The claim [remains pending until missing]{.underline} documents are
    received.

### Stage 5: Indexing in McCamish

a.  Once documents are complete, the claim is **indexed in McCamish** as
    a **Service Request**.

b.  Indexing is done at the receiving office (BO/SO/CPC).

c.  [A Claim ID is generated for tracking]{.mark}.

### Stage 6: Document Scanning

a.  At CPC, the indexed claim form and all enclosures are **scanned**.

b.  Scanned images are uploaded into the system and linked to the Claim
    ID.

### Stage 7: Data Entry

a.  CPC staff manually enter claim details into McCamish.

b.  Multiple fields are keyed in (policy number, claimant details, bank
    details, etc.).

c.  This **[process is repetitive and requires multiple stages of
    entry]{.underline}**.

### Stage 8: Quality Check (QC)

a.  A CPC supervisor performs **QC verification**.

b.  The supervisor checks correctness of data entry and completeness of
    documents.

c.  If errors are found, the claim is sent back for correction.

### Stage 9: Approval Workflow

a.  The claim is routed to the **Postmaster/Approving Authority**.

b.  The approver reviews all fields, supporting documents, and scanned
    images.

c.  The approver either:

> **Approves the claim,**
>
> or
>
> **Rejects** the claim (with reasons).

### Stage 10: Communication of Decision

a.  If approved:

- A [**Sanction Letter** (3 copies) is generated --- for payment,
  record, and policyholder]{.mark}.

- The sanction letter is printed and dispatched by post.

b.  If rejected:

- A **Rejection Letter** is generated.

- Sent to the insurant by registered post.

### Stage 11: Bank Verification

a.  If bank details are provided, staff manually verify them using the
    **passbook copy** or cancelled cheque.

b.  If details are unclear, the insurant is asked to resubmit.

### Stage 12: Disbursement

a.  Payment is processed manually:

- **EFT/NEFT** (if bank details are valid), or

- **Cheque** (if EFT not possible).

b.  Disbursement details are updated manually in McCamish.

### Stage 13: Voucher Submission

a.  Payment vouchers are prepared manually.

b.  Submitted to the **Accounts Section** for reconciliation and
    closure.

### Stage 14: End of Process

a.  Claim is marked as **"Paid"** or **"Rejected"** in McCamish.

b.  No structured feedback or monitoring is done at this stage.

## 

[\]{.underline}

## [1.5 Table on Gap Identification]{.underline}

  -----------------------------------------------------------------------
  Process Area            Current Practice        Gap Identified
                          (McCamish)              
  ----------------------- ----------------------- -----------------------
  SB Due Report           Manually triggered on   No automation, no
  Generation              1st working day of      real-time visibility
                          month; only monthly     
                          report                  

  Customer Intimation     Intimation letter       Limited channels;
                          generated and sent by   mobile/email not
                          registered post;        mandatory
                          SMS/email optional      

  Claim Form Dispatch     Blank claim form sent   No digital option
                          only by post            

  Claim Submission        Manual submission at    No online submission
                          BO/SO/CPC with physical 
                          documents               

  Initial Scrutiny        Manual check by         Human error chances
                          BPM/SPM/Postmaster      

  Missing Document        Staff manually follow   Time-consuming,
  Handling                up with insurant        

  Indexing in McCamish    Manual indexing at      Error-prone,
                          receiving office        time-consuming

  Document Scanning       Done at CPC after       Delayed digitization
                          receipt                 

  Data Entry              Multiple manual entry   Duplication, errors
                          stages                  

  QC Verification         Manual QC by CPC        Dependent on individual
                          supervisor              diligence

  Approval Workflow       Postmaster reviews      No SLA enforcement
                          manually                

  Sanction/Rejection      Printed and sent by     Slow, costly, no
  Letters                 post                    digital record

  Bank Verification       Manual                  Error-prone, slow
                          passbook/cancelled      
                          cheque check            

  Disbursement            Manual entry for        Delays, reconciliation
                          EFT/cheque              issues

  Voucher Submission      Manual preparation and  Paper dependency
                          submission              

  Customer Tracking       Must visit PO for       No online visibility
                          status                  

  Monitoring by Admin     [No real-time           
  Office                  dashboard]{.mark}       

  Escalation of Pending   [No auto alert for      SLA breaches go
  Cases                   delays]{.mark}          unnoticed

  Customer Feedback       Not collected           No service quality loop

  Integration with Other  McCamish standalone     No API links
  Systems                                         

                                                  
  -----------------------------------------------------------------------

## 1.6 Revised Survival Benefit Workflow -- IMS 2.0

+-----------+--------------------+-------------------------------------------+
| FRS ID    | Functional         | Detailed Specification                    |
|           | Requirement        |                                           |
|           | Description        |                                           |
+===========+====================+===========================================+
| FRS-SB-01 | Auto-Generation of | The system SHALL automatically generate   |
|           | Survival Benefit   | the HO-level SB due report daily/weekly.  |
|           | Due Report         | The system SHALL provide a dashboard view |
|           |                    | of this report, accessible by the         |
|           |                    | CPC/Admin Office.                         |
+-----------+--------------------+-------------------------------------------+
| FRS-SB-02 | Multi-Channel      | The system SHALL automatically send the   |
|           | Intimation to      | intimation notice via SMS, Email, and     |
|           | Policyholder       | WhatsApp. The system SHALL use Registered |
|           |                    | Post as a fallback channel. The system    |
|           |                    | SHALL include a secure link within        |
|           |                    | digital intimations for the insurant to   |
|           |                    | submit the claim                          |
|           |                    |                                           |
|           |                    | online.                                   |
+-----------+--------------------+-------------------------------------------+
| FRS-SB-03 | Online Claim       | The system SHALL allow the insurant to    |
|           | Submission by      | upload the claim form and required        |
|           |                    | documents via the Customer Portal and the |
|           | Insurant           | Mobile App. The system SHALL integrate    |
|           |                    | with DigiLocker for fetching the policy   |
|           |                    | document. The system SHALL generate an    |
|           |                    | Auto-acknowledgment with a unique Claim   |
|           |                    | ID upon successful submission.            |
+-----------+--------------------+-------------------------------------------+
| FRS-SB-04 | Initial Scrutiny   | The system SHALL instantly flag any       |
|           | at                 | missing documents upon submission. The    |
|           |                    | system SHALL send auto-reminders          |
|           | CPC (Digital)      | (SMS/email/WhatsApp) to the customer for  |
|           |                    | outstanding documents. CPC staff SHALL    |
|           |                    | verify the completeness of the claim      |
|           |                    | digitally via the system interface.       |
+-----------+--------------------+-------------------------------------------+
| FRS-SB-05 | Indexing in IMS    | The system SHALL automatically index the  |
|           | 2.0                | claim as a Service Request upon final     |
|           |                    | submission of documents. The system SHALL |
|           | (Automatic)        | link the Service Request to the policy    |
|           |                    | and the Claim ID.                         |
+-----------+--------------------+-------------------------------------------+
| FRS-SB-06 | Document Scanning  | The system SHALL support documents being  |
|           |                    | scanned at the source (BO/SO) or uploaded |
|           | & Upload           | by the customer. The system SHALL         |
|           |                    | auto-tag documents and store them in the  |
|           |                    | ECMS (Electronic Content Management       |
|           |                    | System).                                  |
+-----------+--------------------+-------------------------------------------+
| FRS-SB-07 | Data Entry & QC    | The system SHALL auto-populate claim data |
|           |                    | fields from uploaded documents (using     |
|           | Verification       | OCR/data extraction). The CPC supervisor  |
|           |                    | SHALL perform QC digitally via the        |
|           | (Automated)        | dashboard to verify the accuracy of       |
|           |                    |                                           |
|           |                    | auto-populated data.                      |
+-----------+--------------------+-------------------------------------------+
| FRS-SB-08 | Approval Workflow  | The Postmaster/Approving Authority SHALL  |
|           |                    | review the claim via a dedicated          |
|           |                    | dashboard. The Approver SHALL be able to  |
|           |                    | Approve or Reject the claim using a       |
|           |                    | digital signature. The system SHALL       |
|           |                    | enforce a Service Level Agreement (SLA of |
|           |                    | 7 days) for approval.                     |
+-----------+--------------------+-------------------------------------------+
| FRS-SB-09 | Sanction/Rejection | The system SHALL auto-generate the        |
|           |                    | Sanction or Rejection letter with a       |
|           | Letter Generation  | timestamp (date/time/second). The system  |
|           |                    | SHALL send the letter via email,          |
|           |                    | WhatsApp, and the Customer Portal. If     |
|           |                    | rejected, the letter SHALL include the    |
|           |                    | reason and a link for appeal.             |
+-----------+--------------------+-------------------------------------------+
| FRS-SB-10 | Bank Account       | The system SHALL use API-based validation |
|           |                    | for verifying the policyholder\'s bank    |
|           | Verification       | account details. If validation fails, the |
|           |                    | system SHALL prompt the user/staff for    |
|           |                    | correction.                               |
+-----------+--------------------+-------------------------------------------+
| FRS-SB-11 | Disbursement       | The system SHALL be Integrated with Core  |
|           |                    | Banking or other necessary systems. The   |
|           |                    | system SHALL process payment using Auto   |
|           |                    | NEFT/IMPS. The disbursement status SHALL  |
|           |                    | be updated in IMS 2.0 automatically.      |
+-----------+--------------------+-------------------------------------------+
| FRS-SB-12 | Voucher            | The system SHALL auto-generate the        |
|           |                    | payment voucher. The voucher SHALL be     |
|           | Submission         | submitted digitally to the Accounts       |
|           |                    | section and linked to the disbursement    |
|           |                    | record.                                   |
+-----------+--------------------+-------------------------------------------+
| FRS-SB-13 | Customer Feedback  | The system SHALL send an auto-message to  |
|           |                    | the customer post-claim settlement to     |
|           | Collection         | collect feedback. The feedback SHALL be   |
|           |                    | stored for service quality monitoring.    |
+-----------+--------------------+-------------------------------------------+
| FRS-SB-14 | Monitoring &       | The system SHALL provide a real-time      |
|           |                    | dashboard for the Admin Office. The       |
|           | Escalation         | system SHALL display SLA countdown and    |
|           |                    | color-coded alerts for pending cases. The |
|           |                    | system SHALL trigger auto-escalation      |
|           |                    |                                           |
|           |                    | if a claim is pending beyond the set      |
|           |                    | threshold.                                |
+-----------+--------------------+-------------------------------------------+
| FRS-SB-15 | Customer Claim     | The system SHALL provide online/mobile    |
|           |                    | access for the customer to track their    |
|           | Tracker            | claim status. The tracker SHALL provide   |
|           |                    | updates at each major stage: submission,  |
|           |                    | scrutiny, approval, and payment.          |
+-----------+--------------------+-------------------------------------------+

## 1.7 Process Flow Diagram

![](media/image1.emf)

## 1.8 Rejection Reason Master

#### A. Policy-Related 

  -------------------------------------------------
  Code         Rejection Reason
  ------------ ------------------------------------
  RJ-P-01      Policy number invalid or not found

  RJ-P-02      Policy inactive on Survival Benefit
               due date

  RJ-P-03      Survival Benefit already paid
  -------------------------------------------------

#### B. Claimant & Eligibility 

  -------------------------------------------------
  Code         Rejection Reason
  ------------ ------------------------------------
  RJ-E-01      Claimant details mismatch with
               policy

  RJ-E-02      Policyholder identity mismatch
  -------------------------------------------------

#### 

#### C. Document-Related 

  -------------------------------------------------
  Code         Rejection Reason
  ------------ ------------------------------------
  RJ-D-01      Mandatory documents not submitted

  RJ-D-02      Suspected forged or fraudulent
               documents

  RJ-D-03      Mismatch between physical and
               digital records
  -------------------------------------------------

## 1.9 Survival Benefit Screens

#### HO Level Detailed Survival Benefit Due Report

![A screenshot of a computer AI-generated content may be
incorrect.](media/image2.png){width="6.268055555555556in"
height="1.6479166666666667in"}

#### Intimation Letter

![A letter of a post AI-generated content may be
incorrect.](media/image3.png){width="6.268055555555556in"
height="5.793055555555555in"}

#### Indexing/ Service Request Screen

![A screenshot of a computer AI-generated content may be
incorrect.](media/image4.png){width="6.268055555555556in"
height="1.3708333333333333in"}

#### Data Entry Screen ![A screenshot of a computer AI-generated content may be incorrect.](media/image5.png){width="6.268055555555556in" height="2.535416666666667in"}

![A screenshot of a computer AI-generated content may be
incorrect.](media/image6.png){width="6.268055555555556in"
height="2.515277777777778in"}

![A screenshot of a computer AI-generated content may be
incorrect.](media/image7.png){width="6.268055555555556in"
height="2.5770833333333334in"}

![A screenshot of a computer AI-generated content may be
incorrect.](media/image8.png){width="6.268055555555556in"
height="2.2618055555555556in"}

#### Approval Letter

![A letter of approval to a health insurance policy AI-generated content
may be incorrect.](media/image9.png){width="4.54251968503937in"
height="6.07241469816273in"}

#### Disbursement Update

![](media/image10.png){width="6.268055555555556in"
height="0.7236111111111111in"}

![A screenshot of a computer AI-generated content may be
incorrect.](media/image11.png){width="6.268055555555556in"
height="2.314583333333333in"}

## 1.10 Test Cases

Indicative Test Scenarios for Survival Benefit is as follows:

#### POSITIVE TEST SCENARIOS (User Perspective) 

P-01: Successful receipt of Survival Benefit intimation

- User receives SB due intimation via SMS / Email / WhatsApp.

- Message contains correct policy details and a secure claim submission
  link.

P-02: Secure claim link opens correctly

- User clicks the link and lands on the official Customer Portal /
  Mobile App.

- Policy details are auto-fetched and visible.

P-03: Successful online claim submission

- User fills claim form correctly.

- All required documents are uploaded.

- Claim is submitted successfully.

P-04: Auto-acknowledgement generation

- System generates a unique Claim ID.

- Acknowledgement is shown on screen and sent to user digitally.

P-05: DigiLocker integration works

- User fetches policy document from DigiLocker successfully.

- Document appears correctly in uploaded document list.

P-06: Auto reminders received for pending documents

- If any document is pending, user receives reminder notifications.

- Reminder clearly mentions missing documents.

P-07: Claim status tracking works

- User tracks claim online.

- Status updates are visible at each stage (submitted, under scrutiny,
  approved, paid).

P-08: Sanction letter received after approval

- User receives sanction letter via Email / WhatsApp / Portal.

- Letter contains correct claim details and timestamp.

P-09: Payment credited successfully

- Claim amount is credited to user's bank account through NEFT/IMPS.

- Payment status is updated in the tracker.

P-10: Feedback submission after settlement

- User receives feedback request after claim completion.

- Feedback can be submitted successfully.

#### NEGATIVE TEST SCENARIOS (User Perspective) 

N-01: Secure link does not open

- User clicks the link but it fails due to invalid/expired link.

N-02: Missing mandatory document

- User tries to submit claim without required documents.

N-03: Invalid document upload

- User uploads unsupported file type or corrupt file.

- System can't upload file due to size.

N-04: DigiLocker access failure

- DigiLocker fetch fails due to consent denial or technical issue.

- System allows manual document upload instead.

N-05: Incorrect bank details entered

- User enters wrong account number / IFSC.

- Bank verification fails and correction is requested.

N-06: Claim rejection by authority

- Claim is rejected due to discrepancy or ineligibility.

N-07: Tracker not accessible / Delayed Status update

- User tries to track claim but portal/app is unavailable.

- Claim is processed internally but tracker still shows old status.

N-08: Payment failure

- Disbursement attempt fails due to banking/technical error.

- User is informed that payment is pending or failed.

N-09: Feedback link not opening

- User clicks feedback link but page does not load.

## 1.11 Attachments

The Following Attachments can be referred.

![](media/image12.emf)

![](media/image1.emf)

![](media/image13.emf)
