> INTERNAL APPROVAL FORM

**Project Name:** EMail

**Version: 1.0**

**Submitted on:**

  ------------------------------------------------------------------------
               **Name**                                 **Date**
  ------------ ---------------------------------------- ------------------
  **Approved                                            
  By:**                                                 

  **Reviewed                                            
  By:**                                                 

  **Prepared                                            
  By: **                                                
  ------------------------------------------------------------------------

> VERSION CONTROL LOG

  -------------------------------------------------------------------------------
  **Version**   **Date**   **Prepared     **Remarks**
                           By**           
  ------------- ---------- -------------- ---------------------------------------
  **1**                                   

                                          

                                          

                                          

                                          
  -------------------------------------------------------------------------------

# Table of Contents {#table-of-contents .TOC-Heading}

[**1. Executive Summary** [4](#executive-summary)](#executive-summary)

[**2. Project Scope** [4](#project-scope)](#project-scope)

[**3. Business Requirements**
[4](#business-requirements)](#business-requirements)

[**4. Functional Requirements Specification**
[4](#functional-requirements-specification)](#functional-requirements-specification)

[4.1 Email Notification Trigger
[4](#email-notification-trigger)](#email-notification-trigger)

[4.2 Email Manual Dispatch
[8](#email-manual-dispatch)](#email-manual-dispatch)

[4.3 Email Template Management
[8](#email-template-management)](#email-template-management)

[4.4 Email Audit Logs & Dashboard
[8](#email-audit-logs-dashboard)](#email-audit-logs-dashboard)

[**5. Attachments** [9](#attachments)](#attachments)

## **1. Executive Summary**

The purpose of this document is to define the functional and
non-functional requirements for the Email Module of the Insurance
Management System (IMS) used by India Post Postal Life Insurance. This
module will facilitate automated and manual email communication with
policyholders, agents, and internal stakeholders.

## **2. Project Scope**

This Module will:

- Automate email dispatch based on policy lifecycle events.

- Support multilingual, standardized templates.

- Track delivery status and retries.

- Maintain audit logs and consent records.

- Integrate with IMS Core, CRM, and approved email servers.

## **3. Business Requirements**

  -----------------------------------------------------------------------
  **ID**      **Requirements**
  ----------- -----------------------------------------------------------
  FS_EM_001   The system shall send automated EMail alerts based on
              predefined insurance lifecycle events.

  FS_EM_002   The system shall support multilingual EMail templates for
              regional language communication.

  FS_EM_003   The system shall track delivery status and enable retries
              for failed EMail messages.

  FS_EM_004   The system shall maintain customer consent records for
              EMail communication preferences.

  FS_EM_005   The system shall provide MIS dashboards to monitor EMail
              volume, success rates, and SLA compliance.
  -----------------------------------------------------------------------

## **4. Functional Requirements Specification**

### 4.1 Email Notification Trigger

- **Purpose:** The System should send automated Emails based on the
  following insurance lifecycle events for the policy and agent.

+--------------------------+--------------------------+--------------------------------------+
| \#                       | EMail Name               | Trigger Condition                    |
+==========================+==========================+======================================+
| 1.                       | Agent Welcome            | This EMail is sent to an Agent just  |
|                          |                          | after the recruitment, informing the |
|                          |                          | Agent of various particulars. For    |
|                          |                          | example, Agent Number, Effective     |
|                          |                          | Date, Agent Portal ID and password   |
|                          |                          | and Sales Support team Contact       |
|                          |                          | Number.                              |
+--------------------------+--------------------------+--------------------------------------+
| 2\.                      | Agent Appointment        | This EMail is sent to an Agent via   |
|                          |                          | the system when the Agent is         |
|                          |                          | appointed. This EMail also informs   |
|                          |                          | the Agent of the following details:  |
|                          |                          |                                      |
|                          |                          | - Agent Number                       |
|                          |                          |                                      |
|                          |                          | - Date of Birth                      |
|                          |                          |                                      |
|                          |                          | - PAN Number details                 |
|                          |                          |                                      |
|                          |                          | - Managing Agent Name                |
+--------------------------+--------------------------+--------------------------------------+
| 3\.                      | Agent License Allotment  | This EMail is generated and sent to  |
|                          |                          | an Agent after license details are   |
|                          |                          | updated on the Agent's profile. This |
|                          |                          | EMail will include details such as   |
|                          |                          | License Number, License Issue Date   |
|                          |                          | and License Renewal Date of the      |
|                          |                          | Agent.                               |
+--------------------------+--------------------------+--------------------------------------+
| 4\.                      | Agent License Suspension | This EMail is generated when the     |
|                          |                          | license of an Agent has not been     |
|                          |                          | renewed and the license has been     |
|                          |                          | moved to the suspended status.       |
+--------------------------+--------------------------+--------------------------------------+
| 5\.                      | Agent Rejection          | This EMail is generated when the     |
|                          |                          | application for a candidate to be an |
|                          |                          | Agent is rejected.                   |
+--------------------------+--------------------------+--------------------------------------+
| 6\.                      | Agent Documents Pending  | This EMail is generated when after   |
|                          |                          | profile creation, the status is      |
|                          |                          | moved to the Pending with the reason |
|                          |                          | as Documents awaited or Discrepancy  |
|                          |                          | found.                               |
+--------------------------+--------------------------+--------------------------------------+
| 7\.                      | Agent Termination        | This EMail is generated whenever an  |
|                          |                          | Agent profile is terminated.         |
+--------------------------+--------------------------+--------------------------------------+
| Policy                                                                                     |
+--------------------------+--------------------------+--------------------------------------+
| 8\.                      | Policy Issue             | Email to be sent on New policy       |
|                          |                          | issue. Digital Policy Bond Link      |
|                          |                          | should be present in Email.          |
+--------------------------+--------------------------+--------------------------------------+
| 9\.                      | Premium Paid             | Email to be sent on successful       |
|                          |                          | Premium Payment. Digital Receipt     |
|                          |                          | should be sent to Policyowner.       |
+--------------------------+--------------------------+--------------------------------------+
| 10\.                     | Duplicate Bond Issued    | Upon Successful issue of Duplicate   |
|                          |                          | Policy Bond. Tracking link (if       |
|                          |                          | available) should be present in      |
|                          |                          | Email.                               |
+--------------------------+--------------------------+--------------------------------------+
| 11\.                     | Service Request Closed   | Upon closure of service request.     |
+--------------------------+--------------------------+--------------------------------------+
| 12\.                     | Intimation/Surrender     | Surrender is processed successfully  |
|                          |                          |                                      |
|                          | Accepted EMail           |                                      |
+--------------------------+--------------------------+--------------------------------------+
| 13\.                     | Policy Quote Value       | User clicks on the Quote icon        |
+--------------------------+--------------------------+--------------------------------------+
| 14\.                     | Surrender Rejection      | Approver rejects the surrender       |
+--------------------------+--------------------------+--------------------------------------+
| 15\.                     | Force Surrender          | Three loan interests are not paid    |
|                          | Intimation               |                                      |
+--------------------------+--------------------------+--------------------------------------+
| 16\.                     | Force Surrender          | Approver approves force surrender    |
|                          | Processed                | and surrender value is \< 10000      |
+--------------------------+--------------------------+--------------------------------------+
| 17\.                     | Force Surrender          | Approver approves force surrender    |
|                          | Processed                | and surrender value is \> 10000      |
+--------------------------+--------------------------+--------------------------------------+
| 18\.                     | Partial Withdrawal EMail | Partial surrender is processed       |
|                          |                          | successfully                         |
+--------------------------+--------------------------+--------------------------------------+
| 19\.                     | Reduced Paid Up          | Reduced Paid up is successfully      |
|                          | Acceptance               | processed                            |
+--------------------------+--------------------------+--------------------------------------+
| 20\.                     | Reduced Paid Up          | Reduced Paid up is rejected by the   |
|                          | Rejection                | approver                             |
+--------------------------+--------------------------+--------------------------------------+
| 21\.                     | Decrease in Sum          | Event is processed successfully      |
|                          | Assured/Premium Accepted |                                      |
|                          | EMail                    |                                      |
+--------------------------+--------------------------+--------------------------------------+
| 22\.                     | Decrease in Sum          | Event is rejected successfully       |
|                          | Assured/Premium          |                                      |
|                          | Rejection EMail          |                                      |
+--------------------------+--------------------------+--------------------------------------+
| 23\.                     | Cover EMail /Loan        | Loan is processed successfully       |
|                          | Sanction EMail           |                                      |
+--------------------------+--------------------------+--------------------------------------+
| 24\.                     | Loan Rejection EMail     | Loan is rejected                     |
+--------------------------+--------------------------+--------------------------------------+
| 25\.                     | Loan Repayment Schedule  | Loan has been sanctioned             |
+--------------------------+--------------------------+--------------------------------------+
| 26\.                     | Default Notice/ Overdue  | Loan interest is not paid and        |
|                          | Intimation- Loan         | capitalization event is triggered    |
|                          | Capitalization EMail     |                                      |
+--------------------------+--------------------------+--------------------------------------+
| 27\.                     | Cheque Dishonor for Loan | Loan repayment cheque is dishonored  |
|                          |                          |                                      |
|                          | Repayment                |                                      |
+--------------------------+--------------------------+--------------------------------------+
| 28\.                     | Loan Closure EMail       | Loan is paid completely and Loan     |
|                          |                          | Account is closed                    |
+--------------------------+--------------------------+--------------------------------------+
| 29\.                     | Loan Quote               | Quote Value to be printed through    |
|                          |                          | the Quote screen                     |
+--------------------------+--------------------------+--------------------------------------+
| 30\.                     | Freelook Cancellation    | Freelook is processed                |
|                          | EMail                    | successfully/approved by the         |
|                          |                          | approver                             |
+--------------------------+--------------------------+--------------------------------------+
| 31\.                     | Freelook Cancellation    | Freelook is rejected by the approver |
|                          | Rejection EMail          |                                      |
+--------------------------+--------------------------+--------------------------------------+
| 32\.                     | Policy Cancellation      | Policy cancellation request is       |
|                          | Acceptance EMail         | approved                             |
+--------------------------+--------------------------+--------------------------------------+
| 33\.                     | Policy Cancellation      | Approver rejects the policy          |
|                          | Rejection EMail          | cancellation request                 |
+--------------------------+--------------------------+--------------------------------------+
| 34\.                     | Conversion Acceptance    | Conversion request is approved       |
|                          | EMail                    |                                      |
+--------------------------+--------------------------+--------------------------------------+
| 35\.                     | Conversion Rejection     | Conversion request is rejected       |
|                          | EMail                    |                                      |
+--------------------------+--------------------------+--------------------------------------+
| 36\.                     | Fund Switch EMail        | Funds are switched successfully      |
|                          | acceptance               |                                      |
+--------------------------+--------------------------+--------------------------------------+
| 37\.                     | Fund Switch Rejection    | Approver rejects Fund Switch         |
|                          | EMail                    |                                      |
+--------------------------+--------------------------+--------------------------------------+
| 38\.                     | Redirection Acceptance   | Future redirection of funds to be    |
|                          | EMail                    | effected                             |
+--------------------------+--------------------------+--------------------------------------+
| 39\.                     | Redirection Rejection    | Approver rejects the redirection     |
|                          | EMail                    | request                              |
+--------------------------+--------------------------+--------------------------------------+
| 40\.                     | Premium Due Notice       | - Renewal reminder EMail will be     |
|                          |                          |   generated based on Frequency,      |
|                          |                          |   specific days before the bill      |
|                          |                          |   to-date.                           |
|                          |                          |                                      |
|                          |                          | Status of policies will not be       |
|                          |                          | generated for monthly frequency      |
|                          |                          | policy                               |
+--------------------------+--------------------------+--------------------------------------+
| 41\.                     | List Bill Due Notice     | - List bill will be populated in the |
|                          |                          |   employer portal by 15th of every   |
|                          |                          |   month                              |
|                          |                          |                                      |
|                          |                          | - Standard EMail to be sent by 10th  |
|                          |                          |   every month                        |
|                          |                          |                                      |
|                          |                          | - The bill details are populated in  |
|                          |                          |   the employer portal that can be    |
|                          |                          |   accessed only by the employer      |
|                          |                          |                                      |
|                          |                          | Reminder must be sent on 7th of the  |
|                          |                          | next month if premium is not         |
|                          |                          | received                             |
+--------------------------+--------------------------+--------------------------------------+
| 42\.                     | List Bill Due Reminder   | Reminder must be sent on 7th of the  |
|                          |                          | next month to the employer if        |
|                          |                          | premium is not received              |
+--------------------------+--------------------------+--------------------------------------+
| 43\.                     | Payment Failure/ Cheque  | Cheque bounce details are received   |
|                          | Bounce EMail             | from the bank                        |
+--------------------------+--------------------------+--------------------------------------+
| 44\.                     | Nomination EMail         | Will be generated when nomination is |
|                          |                          | approved                             |
+--------------------------+--------------------------+--------------------------------------+
| 45\.                     | Absolute Assignment      | Successful processing of the         |
|                          | EMail                    | absolute assignment                  |
+--------------------------+--------------------------+--------------------------------------+
| 46\.                     | Conditional Assignment   | Successful processing of conditional |
|                          | EMail                    | assignment                           |
+--------------------------+--------------------------+--------------------------------------+
| 47\.                     | Assignment Rejection     | Assignment request is rejected       |
|                          | EMail                    |                                      |
+--------------------------+--------------------------+--------------------------------------+
| 48\.                     | Notification EMail       | Claim is notified                    |
+--------------------------+--------------------------+--------------------------------------+
| 49\.                     | Settlement EMail         | Claim is successfully processed      |
|                          |                          |                                      |
|                          |                          | Note: In this case, the EMail will   |
|                          |                          | be sent to the Payee by the user. If |
|                          |                          | the Payee is not mentioned in the    |
|                          |                          | EMail, it needs to pick up the       |
|                          |                          | nominees name and address            |
+--------------------------+--------------------------+--------------------------------------+
| 50\.                     | Discharge Voucher        | Sent as part of claim settlement     |
|                          |                          | EMail                                |
+--------------------------+--------------------------+--------------------------------------+
| 51\.                     | Name Change              | Name in the policy is changed        |
+--------------------------+--------------------------+--------------------------------------+
| 52\.                     | Address Change           | Address in the policy is changed     |
+--------------------------+--------------------------+--------------------------------------+
| 53\.                     | Revival /Reinstatement   | Lapsed policy is revived             |
|                          | EMail                    | successfully and the Approver        |
|                          |                          | approves the revival/reinstatement   |
|                          |                          | request                              |
+--------------------------+--------------------------+--------------------------------------+
| 54\.                     | Reinstatement Rejection  | When revival/reinstatement is        |
|                          | EMail                    | rejected                             |
+--------------------------+--------------------------+--------------------------------------+
| 55\.                     | Void Status EMail        | Policy status is changed to Void     |
+--------------------------+--------------------------+--------------------------------------+
| 56\.                     | EMail to Customer/Agent  | Missing/additional documents are     |
|                          | for Missing Requirement  | required                             |
+--------------------------+--------------------------+--------------------------------------+
| 57\.                     | Non-receipt EMail to     | Non-receipt of missing/additional    |
|                          | customer/Agent for       | documents after 15 days of issue of  |
|                          | Missing Requirement      | first EMail                          |
+--------------------------+--------------------------+--------------------------------------+
| 58\.                     | EMail to Customer for    | Sent before policy issuance in case  |
|                          | Premium Change           | premium changes (counter offer)      |
+--------------------------+--------------------------+--------------------------------------+
| 59\.                     | EMail to Customer for    | Proposal is declined                 |
|                          | Proposal Rejection       |                                      |
+--------------------------+--------------------------+--------------------------------------+
| 60\.                     | Lapse EMail              | Policy is Lapsed                     |
+--------------------------+--------------------------+--------------------------------------+
| 61\.                     | Survival Benefit         | Survival benefit is due for a policy |
|                          | Notification EMail       |                                      |
+--------------------------+--------------------------+--------------------------------------+
| 62\.                     | Survival Benefit         | Survival benefit is approved for a   |
|                          | Settlement EMail         | policy                               |
+--------------------------+--------------------------+--------------------------------------+
| 63\.                     | Survival Benefit         | Survival benefit is rejected for a   |
|                          | Settlement Rejection     | policy                               |
|                          | EMail                    |                                      |
+--------------------------+--------------------------+--------------------------------------+
| 64\.                     | Electronic Fund Transfer | EFT Transaction is not successfully  |
|                          | (EFT) Transaction        | processed due to any reasons         |
|                          | Unsuccessful             |                                      |
+--------------------------+--------------------------+--------------------------------------+
| 65\.                     | Death Claim Rejection    | Death claim has been rejected        |
+--------------------------+--------------------------+--------------------------------------+
| 66\.                     | Premium Payment Method   | Method has been changed successfully |
|                          | Change                   |                                      |
+--------------------------+--------------------------+--------------------------------------+
| 67\.                     | Premium Payment Method   | Approver rejects the request         |
|                          | Change Rejected          |                                      |
+--------------------------+--------------------------+--------------------------------------+
| 68\.                     | Maturity Claim           | Policy reaches its maturity          |
|                          | Notification EMail       |                                      |
+--------------------------+--------------------------+--------------------------------------+
| 69\.                     | Maturity Calculation     | Maturity claim is approved by the    |
|                          | Sheet and Settlement     | approver                             |
|                          | EMail                    |                                      |
+--------------------------+--------------------------+--------------------------------------+
| 70\.                     | Maturity Claim Rejection | Maturity claim is rejected by the    |
|                          | EMail                    | approver                             |
+--------------------------+--------------------------+--------------------------------------+
| 71\.                     | Surrender Withdrawal-    | Approver approves withdrawal         |
|                          | Approval                 |                                      |
+--------------------------+--------------------------+--------------------------------------+
| 72\.                     | Surrender Withdrawal-    | Approver rejects withdrawal          |
|                          | Rejection                |                                      |
+--------------------------+--------------------------+--------------------------------------+

### 4.2 Email Manual Dispatch

- **Purpose:** To provide a user interface for authorized users to
  manually compose and send emails to customers or stakeholders.

- **Fields:**

  - Recipient Email Address -- Text field; supports multiple addresses
    with validation.

  - Subject -- Text field; mandatory.

  - Email Body -- Rich text editor; supports formatting and
    placeholders.

  - Template Selection (Optional) -- Dropdown; pre-defined templates for
    quick use.

  - Attachments -- File upload; supports PDF, DOCX, etc.

  - Send Button -- Action to dispatch the email.

  - Preview Button -- Shows formatted email before sending.

  - Audit Tag (Auto-filled) -- Metadata for tracking sender and
    timestamp.

### 4.3 Email Template Management

- **Purpose:** To allow administrators to create, edit, and manage
  standardized email templates used in automated and manual dispatch.

- **Fields:**

  - Template Name -- Text field; unique identifier.

  - Template Type -- Dropdown; e.g., Policy Issued, Premium Paid.

  - Language -- Dropdown; supports multilingual templates.

  - Subject Line -- Text field; supports dynamic placeholders.

  - Email Body -- Rich text editor; supports HTML and placeholders like
    {CustomerName}.

  - Status -- Toggle; Active/Inactive.

  - Version History -- List; shows previous versions with timestamps.

  - Save/Update Button -- Action to store the template.

### 4.4 Email Audit Logs & Dashboard

- **Purpose:** To provide visibility into email dispatch history,
  delivery status, and performance metrics for compliance and
  operational monitoring.

- **Fields:**

  - Date Range Filter -- Calendar picker; to view logs within a specific
    period.

  - Recipient Email -- Search field; filter by customer email.

  - Event Type -- Dropdown: e.g., Claim Submission, Service Request
    Closed.

  - Delivery Status -- Indicator; Sent, Failed, Bounced.

  - Retry Count -- Numeric field; shows number of retry attempts.

  - Timestamp -- Auto-filled: when the email was sent.

  - Sender ID -- Auto-filled; user/system that triggered the email.

  - Dashboard Widgets -- Graphs showing volume, success rate, SLA
    compliance.

## **5. Attachments**

The following documents can be referred.
