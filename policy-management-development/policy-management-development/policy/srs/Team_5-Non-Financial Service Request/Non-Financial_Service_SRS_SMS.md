> *INTERNAL APPROVAL FORM*

**Project Name:** SMS

**Version: 1.0**

**Submitted on:**

  ----------------------------------------------------------------------
               **Name**                               **Date**
  ------------ -------------------------------------- ------------------
  **Approved                                          
  By:**                                               

  **Reviewed                                          
  By:**                                               

  **Prepared                                          
  By: **                                              
  ----------------------------------------------------------------------

> *VERSION CONTROL LOG*

  ------------------------------------------------------------------------------
  **Version**   **Date**   **Prepared     **Remarks**
                           By**           
  ------------- ---------- -------------- --------------------------------------
  **1**                                   

                                          

                                          

                                          

                                          
  ------------------------------------------------------------------------------

Table of Contents

> [**1. Executive Summary** 4](#executive-summary)
>
> [**2. Project Scope** 4](#project-scope)
>
> [**3. Business Requirements** 4](#business-requirements)
>
> [**4. Functional Requirements Specification**
> 4](#functional-requirements-specification)
>
> [4.1 SMS Notification Trigger 4](#sms-notification-trigger)
>
> [4.2 SMS Manual Dispatch 8](#sms-manual-dispatch)
>
> [4.3 SMS Template Management 8](#sms-template-management)
>
> [4.4 SMS Audit Logs & Dashboard 8](#sms-audit-logs-dashboard)
>
> [**5. Attachments** 9](#attachments)

## **1. Executive Summary**

The purpose of this document is to define the functional and
non-functional requirements for the SMS Module of the Insurance
Management System (IMS) used by India Post Postal Life Insurance. This
module will facilitate automated and manual SMS communication with
policyholders, agents, and internal stakeholders.

## **2. Project Scope**

This Module will:

- Automate SMS dispatch based on policy lifecycle events.

- Support multilingual, standardized templates.

- Track delivery status and retries.

- Maintain audit logs and consent records.

- Integrate with IMS Core, CRM, and approved SMS servers.

## **3. Business Requirements**

  ---------------------------------------------------------------------
  **ID**      **Requirements**
  ----------- ---------------------------------------------------------
  FS_EM_001   The system shall send automated SMS alerts based on
              predefined insurance lifecycle events.

  FS_EM_002   The system shall support multilingual SMS templates for
              regional language communication.

  FS_EM_003   The system shall track delivery status and enable retries
              for failed SMS messages.

  FS_EM_004   The system shall maintain customer consent records for
              SMS communication preferences.

  FS_EM_005   The system shall provide MIS dashboards to monitor SMS
              volume, success rates, and SLA compliance.
  ---------------------------------------------------------------------

## **4. Functional Requirements Specification**

### 4.1 SMS Notification Trigger

- **Purpose:** The System should send automated SMSs based on the
  following insurance lifecycle events for the policy and agent.

+------------------------+------------------------+------------------------------------+
| \#                     | SMS Name               | Trigger Condition                  |
+------------------------+------------------------+------------------------------------+
| 1.                     | Agent Welcome          | This SMS is sent to an Agent just  |
|                        |                        | after the recruitment, informing   |
|                        |                        | the Agent of various particulars.  |
|                        |                        | For example, Agent Number,         |
|                        |                        | Effective Date, Agent Portal ID    |
|                        |                        | and password and Sales Support     |
|                        |                        | team Contact Number.               |
+------------------------+------------------------+------------------------------------+
| 2\.                    | Agent Appointment      | This SMS is sent to an Agent via   |
|                        |                        | the system when the Agent is       |
|                        |                        | appointed. This SMS also informs   |
|                        |                        | the Agent of the following         |
|                        |                        | details:                           |
|                        |                        |                                    |
|                        |                        | - Agent Number                     |
|                        |                        |                                    |
|                        |                        | - Date of Birth                    |
|                        |                        |                                    |
|                        |                        | - PAN Number details               |
|                        |                        |                                    |
|                        |                        | - Managing Agent Name              |
+------------------------+------------------------+------------------------------------+
| 3\.                    | Agent License          | This SMS is generated and sent to  |
|                        | Allotment              | an Agent after license details are |
|                        |                        | updated on the Agent's profile.    |
|                        |                        | This SMS will include details such |
|                        |                        | as License Number, License Issue   |
|                        |                        | Date and License Renewal Date of   |
|                        |                        | the Agent.                         |
+------------------------+------------------------+------------------------------------+
| 4\.                    | Agent License          | This SMS is generated when the     |
|                        | Suspension             | license of an Agent has not been   |
|                        |                        | renewed and the license has been   |
|                        |                        | moved to the suspended status.     |
+------------------------+------------------------+------------------------------------+
| 5\.                    | Agent Rejection        | This SMS is generated when the     |
|                        |                        | application for a candidate to be  |
|                        |                        | an Agent is rejected.              |
+------------------------+------------------------+------------------------------------+
| 6\.                    | Agent Documents        | This SMS is generated when after   |
|                        | Pending                | profile creation, the status is    |
|                        |                        | moved to the Pending with the      |
|                        |                        | reason as Documents awaited or     |
|                        |                        | Discrepancy found.                 |
+------------------------+------------------------+------------------------------------+
| 7\.                    | Agent Termination      | This SMS is generated whenever an  |
|                        |                        | Agent profile is terminated.       |
+------------------------+------------------------+------------------------------------+
| Policy                                                                               |
+------------------------+------------------------+------------------------------------+
| 8\.                    | Policy Issue           | SMS to be sent on New policy       |
|                        |                        | issue. Digital Policy Bond Link    |
|                        |                        | should be present in SMS.          |
+------------------------+------------------------+------------------------------------+
| 9\.                    | Premium Paid           | SMS to be sent on successful       |
|                        |                        | Premium Payment. Digital Receipt   |
|                        |                        | should be sent to Policyowner.     |
+------------------------+------------------------+------------------------------------+
| 10\.                   | Duplicate Bond Issued  | Upon Successful issue of Duplicate |
|                        |                        | Policy Bond. Tracking link (if     |
|                        |                        | available) should be present in    |
|                        |                        | SMS.                               |
+------------------------+------------------------+------------------------------------+
| 11\.                   | Service Request Closed | Upon closure of service request.   |
+------------------------+------------------------+------------------------------------+
| 12\.                   | Intimation/Surrender   | Surrender is processed             |
|                        |                        | successfully                       |
|                        | Accepted SMS           |                                    |
+------------------------+------------------------+------------------------------------+
| 13\.                   | Policy Quote Value     | User clicks on the Quote icon      |
+------------------------+------------------------+------------------------------------+
| 14\.                   | Surrender Rejection    | Approver rejects the surrender     |
+------------------------+------------------------+------------------------------------+
| 15\.                   | Force Surrender        | Three loan interests are not paid  |
|                        | Intimation             |                                    |
+------------------------+------------------------+------------------------------------+
| 16\.                   | Force Surrender        | Approver approves force surrender  |
|                        | Processed              | and surrender value is \< 10000    |
+------------------------+------------------------+------------------------------------+
| 17\.                   | Force Surrender        | Approver approves force surrender  |
|                        | Processed              | and surrender value is \> 10000    |
+------------------------+------------------------+------------------------------------+
| 18\.                   | Partial Withdrawal SMS | Partial surrender is processed     |
|                        |                        | successfully                       |
+------------------------+------------------------+------------------------------------+
| 19\.                   | Reduced Paid Up        | Reduced Paid up is successfully    |
|                        | Acceptance             | processed                          |
+------------------------+------------------------+------------------------------------+
| 20\.                   | Reduced Paid Up        | Reduced Paid up is rejected by the |
|                        | Rejection              | approver                           |
+------------------------+------------------------+------------------------------------+
| 21\.                   | Decrease in Sum        | Event is processed successfully    |
|                        | Assured/Premium        |                                    |
|                        | Accepted SMS           |                                    |
+------------------------+------------------------+------------------------------------+
| 22\.                   | Decrease in Sum        | Event is rejected successfully     |
|                        | Assured/Premium        |                                    |
|                        | Rejection SMS          |                                    |
+------------------------+------------------------+------------------------------------+
| 23\.                   | Cover SMS /Loan        | Loan is processed successfully     |
|                        | Sanction SMS           |                                    |
+------------------------+------------------------+------------------------------------+
| 24\.                   | Loan Rejection SMS     | Loan is rejected                   |
+------------------------+------------------------+------------------------------------+
| 25\.                   | Loan Repayment         | Loan has been sanctioned           |
|                        | Schedule               |                                    |
+------------------------+------------------------+------------------------------------+
| 26\.                   | Default Notice/        | Loan interest is not paid and      |
|                        | Overdue Intimation-    | capitalization event is triggered  |
|                        | Loan Capitalization    |                                    |
|                        | SMS                    |                                    |
+------------------------+------------------------+------------------------------------+
| 27\.                   | Cheque Dishonor for    | Loan repayment cheque is           |
|                        | Loan                   | dishonored                         |
|                        |                        |                                    |
|                        | Repayment              |                                    |
+------------------------+------------------------+------------------------------------+
| 28\.                   | Loan Closure SMS       | Loan is paid completely and Loan   |
|                        |                        | Account is closed                  |
+------------------------+------------------------+------------------------------------+
| 29\.                   | Loan Quote             | Quote Value to be printed through  |
|                        |                        | the Quote screen                   |
+------------------------+------------------------+------------------------------------+
| 30\.                   | Freelook Cancellation  | Freelook is processed              |
|                        | SMS                    | successfully/approved by the       |
|                        |                        | approver                           |
+------------------------+------------------------+------------------------------------+
| 31\.                   | Freelook Cancellation  | Freelook is rejected by the        |
|                        | Rejection SMS          | approver                           |
+------------------------+------------------------+------------------------------------+
| 32\.                   | Policy Cancellation    | Policy cancellation request is     |
|                        | Acceptance SMS         | approved                           |
+------------------------+------------------------+------------------------------------+
| 33\.                   | Policy Cancellation    | Approver rejects the policy        |
|                        | Rejection SMS          | cancellation request               |
+------------------------+------------------------+------------------------------------+
| 34\.                   | Conversion Acceptance  | Conversion request is approved     |
|                        | SMS                    |                                    |
+------------------------+------------------------+------------------------------------+
| 35\.                   | Conversion Rejection   | Conversion request is rejected     |
|                        | SMS                    |                                    |
+------------------------+------------------------+------------------------------------+
| 36\.                   | Fund Switch SMS        | Funds are switched successfully    |
|                        | acceptance             |                                    |
+------------------------+------------------------+------------------------------------+
| 37\.                   | Fund Switch Rejection  | Approver rejects Fund Switch       |
|                        | SMS                    |                                    |
+------------------------+------------------------+------------------------------------+
| 38\.                   | Redirection Acceptance | Future redirection of funds to be  |
|                        | SMS                    | effected                           |
+------------------------+------------------------+------------------------------------+
| 39\.                   | Redirection Rejection  | Approver rejects the redirection   |
|                        | SMS                    | request                            |
+------------------------+------------------------+------------------------------------+
| 40\.                   | Premium Due Notice     | - Renewal reminder SMS will be     |
|                        |                        |   generated based on Frequency,    |
|                        |                        |   specific days before the bill    |
|                        |                        |   to-date.                         |
|                        |                        |                                    |
|                        |                        | Status of policies will not be     |
|                        |                        | generated for monthly frequency    |
|                        |                        | policy                             |
+------------------------+------------------------+------------------------------------+
| 41\.                   | List Bill Due Notice   | - List bill will be populated in   |
|                        |                        |   the employer portal by 15th of   |
|                        |                        |   every month                      |
|                        |                        |                                    |
|                        |                        | - Standard SMS to be sent by 10th  |
|                        |                        |   every month                      |
|                        |                        |                                    |
|                        |                        | - The bill details are populated   |
|                        |                        |   in the employer portal that can  |
|                        |                        |   be accessed only by the employer |
|                        |                        |                                    |
|                        |                        | Reminder must be sent on 7th of    |
|                        |                        | the next month if premium is not   |
|                        |                        | received                           |
+------------------------+------------------------+------------------------------------+
| 42\.                   | List Bill Due Reminder | Reminder must be sent on 7th of    |
|                        |                        | the next month to the employer if  |
|                        |                        | premium is not received            |
+------------------------+------------------------+------------------------------------+
| 43\.                   | Payment Failure/       | Cheque bounce details are received |
|                        | Cheque Bounce SMS      | from the bank                      |
+------------------------+------------------------+------------------------------------+
| 44\.                   | Nomination SMS         | Will be generated when nomination  |
|                        |                        | is approved                        |
+------------------------+------------------------+------------------------------------+
| 45\.                   | Absolute Assignment    | Successful processing of the       |
|                        | SMS                    | absolute assignment                |
+------------------------+------------------------+------------------------------------+
| 46\.                   | Conditional Assignment | Successful processing of           |
|                        | SMS                    | conditional assignment             |
+------------------------+------------------------+------------------------------------+
| 47\.                   | Assignment Rejection   | Assignment request is rejected     |
|                        | SMS                    |                                    |
+------------------------+------------------------+------------------------------------+
| 48\.                   | Notification SMS       | Claim is notified                  |
+------------------------+------------------------+------------------------------------+
| 49\.                   | Settlement SMS         | Claim is successfully processed    |
|                        |                        |                                    |
|                        |                        | Note: In this case, the SMS will   |
|                        |                        | be sent to the Payee by the user.  |
|                        |                        | If the Payee is not mentioned in   |
|                        |                        | the SMS, it needs to pick up the   |
|                        |                        | nominees name and address          |
+------------------------+------------------------+------------------------------------+
| 50\.                   | Discharge Voucher      | Sent as part of claim settlement   |
|                        |                        | SMS                                |
+------------------------+------------------------+------------------------------------+
| 51\.                   | Name Change            | Name in the policy is changed      |
+------------------------+------------------------+------------------------------------+
| 52\.                   | Address Change         | Address in the policy is changed   |
+------------------------+------------------------+------------------------------------+
| 53\.                   | Revival /Reinstatement | Lapsed policy is revived           |
|                        | SMS                    | successfully and the Approver      |
|                        |                        | approves the revival/reinstatement |
|                        |                        | request                            |
+------------------------+------------------------+------------------------------------+
| 54\.                   | Reinstatement          | When revival/reinstatement is      |
|                        | Rejection SMS          | rejected                           |
+------------------------+------------------------+------------------------------------+
| 55\.                   | Void Status SMS        | Policy status is changed to Void   |
+------------------------+------------------------+------------------------------------+
| 56\.                   | SMS to Customer/Agent  | Missing/additional documents are   |
|                        | for Missing            | required                           |
|                        | Requirement            |                                    |
+------------------------+------------------------+------------------------------------+
| 57\.                   | Non-receipt SMS to     | Non-receipt of missing/additional  |
|                        | customer/Agent for     | documents after 15 days of issue   |
|                        | Missing Requirement    | of first SMS                       |
+------------------------+------------------------+------------------------------------+
| 58\.                   | SMS to Customer for    | Sent before policy issuance in     |
|                        | Premium Change         | case premium changes (counter      |
|                        |                        | offer)                             |
+------------------------+------------------------+------------------------------------+
| 59\.                   | SMS to Customer for    | Proposal is declined               |
|                        | Proposal Rejection     |                                    |
+------------------------+------------------------+------------------------------------+
| 60\.                   | SMS to customer for    | Policy is Lapsed                   |
|                        | Policy Lapseation      |                                    |
+------------------------+------------------------+------------------------------------+
| 61\.                   | Survival Benefit       | Survival benefit is due for a      |
|                        | Notification SMS       | policy                             |
+------------------------+------------------------+------------------------------------+
| 62\.                   | Survival Benefit       | Survival benefit is approved for a |
|                        | Settlement SMS         | policy                             |
+------------------------+------------------------+------------------------------------+
| 63\.                   | Survival Benefit       | Survival benefit is rejected for a |
|                        | Settlement Rejection   | policy                             |
|                        | SMS                    |                                    |
+------------------------+------------------------+------------------------------------+
| 64\.                   | Electronic Fund        | EFT Transaction is not             |
|                        | Transfer (EFT)         | successfully processed due to any  |
|                        | Transaction            | reasons                            |
|                        | Unsuccessful           |                                    |
+------------------------+------------------------+------------------------------------+
| 65\.                   | Death Claim Rejection  | Death claim has been rejected      |
+------------------------+------------------------+------------------------------------+
| 66\.                   | Premium Payment Method | Method has been changed            |
|                        | Change                 | successfully                       |
+------------------------+------------------------+------------------------------------+
| 67\.                   | Premium Payment Method | Approver rejects the request       |
|                        | Change Rejected        |                                    |
+------------------------+------------------------+------------------------------------+
| 68\.                   | Maturity Claim         | Policy reaches its maturity        |
|                        | Notification SMS       |                                    |
+------------------------+------------------------+------------------------------------+
| 69\.                   | Maturity Calculation   | Maturity claim is approved by the  |
|                        | Sheet and Settlement   | approver                           |
|                        | SMS                    |                                    |
+------------------------+------------------------+------------------------------------+
| 70\.                   | Maturity Claim         | Maturity claim is rejected by the  |
|                        | Rejection SMS          | approver                           |
+------------------------+------------------------+------------------------------------+
| 71\.                   | Surrender Withdrawal-  | Approver approves withdrawal       |
|                        | Approval               |                                    |
+------------------------+------------------------+------------------------------------+
| 72\.                   | Surrender Withdrawal-  | Approver rejects withdrawal        |
|                        | Rejection              |                                    |
+========================+========================+====================================+

### 4.2 SMS Manual Dispatch

- **Purpose:** To provide a user interface for authorized users to
  manually compose and send SMSs to customers or stakeholders.

- **Fields:**

  - Recipient Mobile Number -- Text field; supports multiple Mobile
    Number with validation.

  - Subject -- Text field; mandatory.

  - SMS Body -- Rich text editor; supports formatting and placeholders.

  - Template Selection (Optional) -- Dropdown; pre-defined templates for
    quick use.

  - Attachments -- File upload; supports PDF, DOCX, etc.

  - Send Button -- Action to dispatch the SMS.

  - Preview Button -- Shows formatted SMS before sending.

  - Audit Tag (Auto-filled) -- Metadata for tracking sender and
    timestamp.

### 4.3 SMS Template Management

- **Purpose:** To allow administrators to create, edit, and manage
  standardized SMS templates used in automated and manual dispatch.

- **Fields:**

  - Template Name -- Text field; unique identifier.

  - Template Type -- Dropdown; e.g., Policy Issued, Premium Paid.

  - Language -- Dropdown; supports multilingual templates.

  - Subject Line -- Text field; supports dynamic placeholders.

  - SMS Body -- Rich text editor; supports HTML and placeholders like
    {CustomerName}.

  - Status -- Toggle; Active/Inactive.

  - Version History -- List; shows previous versions with timestamps.

  - Save/Update Button -- Action to store the template.

### 4.4 SMS Audit Logs & Dashboard

- **Purpose:** To provide visibility into SMS dispatch history, delivery
  status, and performance metrics for compliance and operational
  monitoring.

- **Fields:**

  - Date Range Filter -- Calendar picker; to view logs within a specific
    period.

  - Recipient Mobile Number- Search Field; filter by Customer Mobile
    Number

  - Recipient SMS Type -- Search field; filter by customer SMS Type.

  - Event Type -- Dropdown: e.g., Claim Submission, Service Request
    Closed.

  - Delivery Status -- Indicator; Sent, Failed, Bounced.

  - Retry Count -- Numeric field; shows number of retry attempts.

  - Timestamp -- Auto-filled: when the SMS was sent.

  - Sender ID -- Auto-filled; user/system that triggered the SMS.

  - Dashboard Widgets -- Graphs showing volume, success rate, SLA
    compliance.

## **5. Attachments**

The following documents can be referred.
