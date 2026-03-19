> *INTERNAL APPROVAL FORM*

**Project Name:** WhatsApp

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
> [4.1 WhatsApp Notification Trigger 4](#whatsapp-notification-trigger)
>
> [4.2 WhatsApp Manual Dispatch 8](#whatsapp-message-manual-dispatch)
>
> [4.3 WhatsApp Template Management 8](#whatsapp-template-management)
>
> [4.4 WhatsApp Audit Logs & Dashboard
> 8](#whatsapp-audit-logs-dashboard)
>
> [**5. Attachments** 9](#attachments)

## **1. Executive Summary**

The purpose of this document is to define the functional and
non-functional requirements for the WhatsApp Module of the Insurance
Management System (IMS) used by India Post Postal Life Insurance. This
module will facilitate automated and manual WhatsApp communication with
policyholders, agents, and internal stakeholders.

## **2. Project Scope**

This Module will:

- Automate WhatsApp dispatch based on policy lifecycle events.

- Support multilingual, standardized templates.

- Track delivery status and retries.

- Maintain audit logs and consent records.

- Integrate with IMS Core, CRM, and approved WhatsApp servers.

## **3. Business Requirements**

  ---------------------------------------------------------------------
  **ID**      **Requirements**
  ----------- ---------------------------------------------------------
  FS_EM_001   The system shall send automated WhatsApp alerts based on
              predefined insurance lifecycle events.

  FS_EM_002   The system shall support multilingual WhatsApp templates
              for regional language communication.

  FS_EM_003   The system shall track delivery status and enable retries
              for failed WhatsApp messages.

  FS_EM_004   The system shall maintain customer consent records for
              WhatsApp communication preferences.

  FS_EM_005   The system shall provide MIS dashboards to monitor
              WhatsApp volume, success rates, and SLA compliance.
  ---------------------------------------------------------------------

## **4. Functional Requirements Specification**

### 4.1 WhatsApp Notification Trigger

- **Purpose:** The System should send automated WhatsApps based on the
  following insurance lifecycle events for the policy and agent.

+------------------------+------------------------+------------------------------------+
| \#                     | WhatsApp Name          | Trigger Condition                  |
+------------------------+------------------------+------------------------------------+
| 1.                     | Agent Welcome          | This WhatsApp Message is sent to   |
|                        |                        | an Agent just after the            |
|                        |                        | recruitment, informing the Agent   |
|                        |                        | of various particulars. For        |
|                        |                        | example, Agent Number, Effective   |
|                        |                        | Date, Agent Portal ID and password |
|                        |                        | and Sales Support team Contact     |
|                        |                        | Number.                            |
+------------------------+------------------------+------------------------------------+
| 2\.                    | Agent Appointment      | This WhatsApp Message is sent to   |
|                        |                        | an Agent via the system when the   |
|                        |                        | Agent is appointed. This WhatsApp  |
|                        |                        | Message also informs the Agent of  |
|                        |                        | the following details:             |
|                        |                        |                                    |
|                        |                        | - Agent Number                     |
|                        |                        |                                    |
|                        |                        | - Date of Birth                    |
|                        |                        |                                    |
|                        |                        | - PAN Number details               |
|                        |                        |                                    |
|                        |                        | - Managing Agent Name              |
+------------------------+------------------------+------------------------------------+
| 3\.                    | Agent License          | This WhatsApp Message is generated |
|                        | Allotment              | and sent to an Agent after license |
|                        |                        | details are updated on the Agent's |
|                        |                        | profile. This WhatsApp Message     |
|                        |                        | will include details such as       |
|                        |                        | License Number, License Issue Date |
|                        |                        | and License Renewal Date of the    |
|                        |                        | Agent.                             |
+------------------------+------------------------+------------------------------------+
| 4\.                    | Agent License          | This WhatsApp Message is generated |
|                        | Suspension             | when the license of an Agent has   |
|                        |                        | not been renewed and the license   |
|                        |                        | has been moved to the suspended    |
|                        |                        | status.                            |
+------------------------+------------------------+------------------------------------+
| 5\.                    | Rejection of Agency    | This WhatsApp Message is generated |
|                        | request                | when the application for a         |
|                        |                        | candidate to be an Agent is        |
|                        |                        | rejected.                          |
+------------------------+------------------------+------------------------------------+
| 6\.                    | Agent Documents        | This WhatsApp Message is generated |
|                        | Pending                | when after profile creation, the   |
|                        |                        | status is moved to the Pending     |
|                        |                        | with the reason as Documents       |
|                        |                        | awaited or Discrepancy found.      |
+------------------------+------------------------+------------------------------------+
| 7\.                    | Agent Termination      | This WhatsApp Message is generated |
|                        |                        | whenever an Agent profile is       |
|                        |                        | terminated.                        |
+------------------------+------------------------+------------------------------------+
| Policy                                                                               |
+------------------------+------------------------+------------------------------------+
| 8\.                    | Policy Issue           | WhatsApp Message to be sent on New |
|                        |                        | policy issue. Digital Policy Bond  |
|                        |                        | Link should be present in WhatsApp |
|                        |                        | Message.                           |
+------------------------+------------------------+------------------------------------+
| 9\.                    | Premium Paid           | WhatsApp Message to be sent on     |
|                        |                        | successful Premium Payment.        |
|                        |                        | Digital Receipt should be sent to  |
|                        |                        | Policyowner.                       |
+------------------------+------------------------+------------------------------------+
| 10\.                   | Duplicate Bond Issued  | Upon Successful issue of Duplicate |
|                        |                        | Policy Bond. Tracking link (if     |
|                        |                        | available) should be present in    |
|                        |                        | WhatsApp Message.                  |
+------------------------+------------------------+------------------------------------+
| 11\.                   | Service Request Closed | Upon closure of service request.   |
+------------------------+------------------------+------------------------------------+
| 12\.                   | Intimation/Surrender   | Surrender is processed             |
|                        |                        | successfully                       |
|                        | Accepted WhatsApp      |                                    |
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
| 18\.                   | Partial Withdrawal     | Partial surrender is processed     |
|                        | WhatsApp Message       | successfully                       |
+------------------------+------------------------+------------------------------------+
| 19\.                   | Reduced Paid Up        | Reduced Paid up is successfully    |
|                        | Acceptance             | processed                          |
+------------------------+------------------------+------------------------------------+
| 20\.                   | Reduced Paid Up        | Reduced Paid up is rejected by the |
|                        | Rejection              | approver                           |
+------------------------+------------------------+------------------------------------+
| 21\.                   | Decrease in Sum        | Event is processed successfully    |
|                        | Assured/Premium        |                                    |
|                        | Accepted WhatsApp      |                                    |
|                        | Message                |                                    |
+------------------------+------------------------+------------------------------------+
| 22\.                   | Decrease in Sum        | Event is rejected successfully     |
|                        | Assured/Premium        |                                    |
|                        | Rejection WhatsApp     |                                    |
|                        | Message                |                                    |
+------------------------+------------------------+------------------------------------+
| 23\.                   | Cover WhatsApp Message | Loan is processed successfully     |
|                        | /Loan Sanction         |                                    |
|                        | WhatsApp Message       |                                    |
+------------------------+------------------------+------------------------------------+
| 24\.                   | Loan Rejection         | Loan is rejected                   |
|                        | WhatsApp Message       |                                    |
+------------------------+------------------------+------------------------------------+
| 25\.                   | Loan Repayment         | Loan has been sanctioned           |
|                        | Schedule               |                                    |
+------------------------+------------------------+------------------------------------+
| 26\.                   | Default Notice/        | Loan interest is not paid and      |
|                        | Overdue Intimation-    | capitalization event is triggered  |
|                        | Loan Capitalization    |                                    |
|                        | WhatsApp Message       |                                    |
+------------------------+------------------------+------------------------------------+
| 27\.                   | Cheque Dishonor for    | Loan repayment cheque is           |
|                        | Loan                   | dishonored                         |
|                        |                        |                                    |
|                        | Repayment              |                                    |
+------------------------+------------------------+------------------------------------+
| 28\.                   | Loan Closure WhatsApp  | Loan is paid completely and Loan   |
|                        | Message                | Account is closed                  |
+------------------------+------------------------+------------------------------------+
| 29\.                   | Loan Quote             | Quote Value to be printed through  |
|                        |                        | the Quote screen                   |
+------------------------+------------------------+------------------------------------+
| 30\.                   | Freelook Cancellation  | Freelook is processed              |
|                        | WhatsApp Message       | successfully/approved by the       |
|                        |                        | approver                           |
+------------------------+------------------------+------------------------------------+
| 31\.                   | Freelook Cancellation  | Freelook is rejected by the        |
|                        | Rejection WhatsApp     | approver                           |
|                        | Message                |                                    |
+------------------------+------------------------+------------------------------------+
| 32\.                   | Policy Cancellation    | Policy cancellation request is     |
|                        | Acceptance WhatsApp    | approved                           |
|                        | Message                |                                    |
+------------------------+------------------------+------------------------------------+
| 33\.                   | Policy Cancellation    | Approver rejects the policy        |
|                        | Rejection WhatsApp     | cancellation request               |
|                        | Message                |                                    |
+------------------------+------------------------+------------------------------------+
| 34\.                   | Conversion Acceptance  | Conversion request is approved     |
|                        | WhatsApp Message       |                                    |
+------------------------+------------------------+------------------------------------+
| 35\.                   | Conversion Rejection   | Conversion request is rejected     |
|                        | WhatsApp Message       |                                    |
+------------------------+------------------------+------------------------------------+
| 36\.                   | Fund Switch WhatsApp   | Funds are switched successfully    |
|                        | Message acceptance     |                                    |
+------------------------+------------------------+------------------------------------+
| 37\.                   | Fund Switch Rejection  | Approver rejects Fund Switch       |
|                        | WhatsApp Message       |                                    |
+------------------------+------------------------+------------------------------------+
| 38\.                   | Redirection Acceptance | Future redirection of funds to be  |
|                        | WhatsApp Message       | effected                           |
+------------------------+------------------------+------------------------------------+
| 39\.                   | Redirection Rejection  | Approver rejects the redirection   |
|                        | WhatsApp Message       | request                            |
+------------------------+------------------------+------------------------------------+
| 40\.                   | Premium Due Notice     | - Renewal reminder WhatsApp        |
|                        |                        |   Message will be generated based  |
|                        |                        |   on Frequency, specific days      |
|                        |                        |   before the bill to-date.         |
|                        |                        |                                    |
|                        |                        | Status of policies will not be     |
|                        |                        | generated for monthly frequency    |
|                        |                        | policy                             |
+------------------------+------------------------+------------------------------------+
| 41\.                   | List Bill Due Notice   | - List bill will be populated in   |
|                        |                        |   the employer portal by 15th of   |
|                        |                        |   every month                      |
|                        |                        |                                    |
|                        |                        | - Standard WhatsApp Message to be  |
|                        |                        |   sent by 10th every month         |
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
|                        | Cheque Bounce WhatsApp | from the bank                      |
|                        | Message                |                                    |
+------------------------+------------------------+------------------------------------+
| 44\.                   | Nomination WhatsApp    | Will be generated when nomination  |
|                        | Message                | is approved                        |
+------------------------+------------------------+------------------------------------+
| 45\.                   | Absolute Assignment    | Successful processing of the       |
|                        | WhatsApp Message       | absolute assignment                |
+------------------------+------------------------+------------------------------------+
| 46\.                   | Conditional Assignment | Successful processing of           |
|                        | WhatsApp Message       | conditional assignment             |
+------------------------+------------------------+------------------------------------+
| 47\.                   | Assignment Rejection   | Assignment request is rejected     |
|                        | WhatsApp Message       |                                    |
+------------------------+------------------------+------------------------------------+
| 48\.                   | Notification WhatsApp  | Claim is notified                  |
|                        | Message                |                                    |
+------------------------+------------------------+------------------------------------+
| 49\.                   | Settlement WhatsApp    | Claim is successfully processed    |
|                        | Message                |                                    |
|                        |                        | Note: In this case, the WhatsApp   |
|                        |                        | Message will be sent to the Payee  |
|                        |                        | by the user. If the Payee is not   |
|                        |                        | mentioned in the WhatsApp Message, |
|                        |                        | it needs to pick up the nominees   |
|                        |                        | name and address                   |
+------------------------+------------------------+------------------------------------+
| 50\.                   | Discharge Voucher      | Sent as part of claim settlement   |
|                        |                        | WhatsApp Message                   |
+------------------------+------------------------+------------------------------------+
| 51\.                   | Name Change            | Name in the policy is changed      |
+------------------------+------------------------+------------------------------------+
| 52\.                   | Address Change         | Address in the policy is changed   |
+------------------------+------------------------+------------------------------------+
| 53\.                   | Revival /Reinstatement | Lapsed policy is revived           |
|                        | WhatsApp Message       | successfully and the Approver      |
|                        |                        | approves the revival/reinstatement |
|                        |                        | request                            |
+------------------------+------------------------+------------------------------------+
| 54\.                   | Reinstatement          | When revival/reinstatement is      |
|                        | Rejection WhatsApp     | rejected                           |
|                        | Message                |                                    |
+------------------------+------------------------+------------------------------------+
| 55\.                   | Void Status WhatsApp   | Policy status is changed to Void   |
|                        | Message                |                                    |
+------------------------+------------------------+------------------------------------+
| 56\.                   | WhatsApp Message to    | Missing/additional documents are   |
|                        | Customer/Agent for     | required                           |
|                        | Missing Requirement    |                                    |
+------------------------+------------------------+------------------------------------+
| 57\.                   | Non-receipt WhatsApp   | Non-receipt of missing/additional  |
|                        | Message to             | documents after 15 days of issue   |
|                        | customer/Agent for     | of first WhatsApp Message          |
|                        | Missing Requirement    |                                    |
+------------------------+------------------------+------------------------------------+
| 58\.                   | WhatsApp Message to    | Sent before policy issuance in     |
|                        | Customer for Premium   | case premium changes (counter      |
|                        | Change                 | offer)                             |
+------------------------+------------------------+------------------------------------+
| 59\.                   | WhatsApp Message to    | Proposal is declined               |
|                        | Customer for Proposal  |                                    |
|                        | Rejection              |                                    |
+------------------------+------------------------+------------------------------------+
| 60\.                   | Policy Lapse WhatsApp  | Policy is Lapsed                   |
|                        | Message                |                                    |
+------------------------+------------------------+------------------------------------+
| 61\.                   | Survival Benefit       | Survival benefit is due for a      |
|                        | Notification WhatsApp  | policy                             |
|                        | Message                |                                    |
+------------------------+------------------------+------------------------------------+
| 62\.                   | Survival Benefit       | Survival benefit is approved for a |
|                        | Settlement WhatsApp    | policy                             |
|                        | Message                |                                    |
+------------------------+------------------------+------------------------------------+
| 63\.                   | Survival Benefit       | Survival benefit is rejected for a |
|                        | Settlement Rejection   | policy                             |
|                        | WhatsApp Message       |                                    |
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
|                        | Notification WhatsApp  |                                    |
|                        | Message                |                                    |
+------------------------+------------------------+------------------------------------+
| 69\.                   | Maturity Calculation   | Maturity claim is approved by the  |
|                        | Sheet and Settlement   | approver                           |
|                        | WhatsApp Message       |                                    |
+------------------------+------------------------+------------------------------------+
| 70\.                   | Maturity Claim         | Maturity claim is rejected by the  |
|                        | Rejection WhatsApp     | approver                           |
|                        | Message                |                                    |
+------------------------+------------------------+------------------------------------+
| 71\.                   | Surrender Withdrawal-  | Approver approves withdrawal       |
|                        | Approval               |                                    |
+------------------------+------------------------+------------------------------------+
| 72\.                   | Surrender Withdrawal-  | Approver rejects withdrawal        |
|                        | Rejection              |                                    |
+========================+========================+====================================+

### 4.2 WhatsApp Message Manual Dispatch

- **Purpose:** To provide a user interface for authorized users to
  manually compose and send WhatsApp Messages to customers or
  stakeholders. Option to send Quote via WhatsApp should be present.

- **Fields:**

  - Recipient Mobile Number -- Text field; supports multiple addresses
    with validation.

  - Subject -- Text field; mandatory.

  - WhatsApp Message Body -- Rich text editor; supports formatting and
    placeholders.

  - Template Selection (Optional) -- Dropdown; pre-defined templates for
    quick use.

  - Quote: Dropdown: Option to select Quote generated by user should be
    present in dropdown. It should give an option to send Quote as
    WhatsApp message to customers.

  - Attachments -- File upload; supports PDF, DOCX, etc.

  - Send Button -- Action to dispatch the WhatsApp Message.

  - Preview Button -- Shows formatted WhatsApp Message before sending.

  - Audit Tag (Auto-filled) -- Metadata for tracking sender and
    timestamp.

### 4.3 WhatsApp Template Management

- **Purpose:** To allow administrators to create, edit, and manage
  standardized WhatsApp Message templates used in automated and manual
  dispatch.

- **Fields:**

  - Template Name -- Text field; unique identifier.

  - Template Type -- Dropdown; e.g., Policy Issued, Premium Paid.

  - Language -- Dropdown; supports multilingual templates.

  - Subject Line -- Text field; supports dynamic placeholders.

  - WhatsApp Message Body -- Rich text editor; supports HTML and
    placeholders like {CustomerName}.

  - Status -- Toggle; Active/Inactive.

  - Version History -- List; shows previous versions with timestamps.

  - Save/Update Button -- Action to store the template.

### 4.4 WhatsApp Audit Logs & Dashboard

- **Purpose:** To provide visibility into WhatsApp Message dispatch
  history, delivery status, and performance metrics for compliance and
  operational monitoring.

- **Fields:**

  - Date Range Filter -- Calendar picker; to view logs within a specific
    period.

  - Mobile Number -- Search field; filter by customer WhatsApp Message.

  - Event Type -- Dropdown: e.g., Claim Submission, Service Request
    Closed.

  - Delivery Status -- Indicator; Sent, Failed, Bounced.

  - Retry Count -- Numeric field; shows number of retry attempts.

  - Timestamp -- Auto-filled: when the WhatsApp Message was sent.

  - Sender ID -- Auto-filled; user/system that triggered the WhatsApp
    Message.

  - Dashboard Widgets -- Graphs showing volume, success rate, SLA
    compliance.

## **5. Attachments**

The following documents can be referred.
