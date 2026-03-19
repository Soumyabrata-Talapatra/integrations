> INTERNAL APPROVAL FORM

**Project Name:** Letters & Reports- Policy

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

[Existing Letters: [4](#existing-letters)](#existing-letters)

[Existing Reports [8](#existing-reports)](#existing-reports)

[**4. Functional Requirements Specification**
[11](#functional-requirements-specification)](#functional-requirements-specification)

[4.1 Create Reports Page
[11](#create-reports-page)](#create-reports-page)

[4.2 Letters Page [11](#letters-page)](#letters-page)

[4.3 Historical Reports Page
[12](#historical-reports-page)](#historical-reports-page)

[**5. Samples:** [12](#samples)](#samples)

## **1. Executive Summary**

This module is designed to automate the generation of letters and
reports related to policy in the Postal Life Insurance system. It
supports both event-triggered letters and manually or periodically
generated reports, with standardized formatting and archival
capabilities.

## **2. Project Scope**

This module will help in understanding the requirements for letters and
reports that need to be generated for the policies.

## **3. Business Requirements**

  -----------------------------------------------------------------------
  Requirement ID Requirement
  -------------- --------------------------------------------------------
  FS_LR_001      The system must generate specific letters based on
                 policy lifecycle events (e.g., policy issue, premium
                 payment, renewal, surrender, etc.).

  FS_LR_002      All letters must follow a consistent format including
                 header, footer, font, salutation, and date/amount
                 formatting.

  FS_LR_003      Reports should be available in both PDF and Excel
                 formats and support adhoc generation.

  FS_LR_004      Users should be able to filter and generate reports
                 based on different parameters like policy number,
                 Customer ID, PAN Number, etc.

  FS_LR_005      Historical letters and reports must be stored and
                 retrievable by users.

  FS_LR_006      The data present in reports should contain data from all
                 sources where transaction is happening like IT 2.0,
                 BBPS, DARPAN, etc. It should be a Unified real-time data
                 warehouse.
  -----------------------------------------------------------------------

### Existing Letters:

List of Existing Letters:

+------+---------------+----------------------+-------------------------------+
| \#   | Letter Name   | Letter Title         | Trigger Condition             |
+:====:+===============+======================+===============================+
| 1\.  | Surrender     | Intimation/Surrender | Surrender is processed        |
|      |               |                      | successfully                  |
|      |               | Accepted Letter      |                               |
+------+---------------+----------------------+-------------------------------+
| 2\.  | Surrender     | Policy Quote Value   | User clicks on the Quote icon |
+------+---------------+----------------------+-------------------------------+
| 3\.  | Surrender     | Surrender Rejection  | Approver rejects the          |
|      |               |                      | surrender                     |
+------+---------------+----------------------+-------------------------------+
| 4\.  | Surrender     | Force Surrender      | Three loan interests are not  |
|      |               | Intimation           | paid                          |
+------+---------------+----------------------+-------------------------------+
| 5\.  | Surrender     | Force Surrender      | Approver approves force       |
|      |               | Processed            | surrender and surrender value |
|      |               |                      | is \< 10000                   |
+------+---------------+----------------------+-------------------------------+
| 6\.  | Surrender     | Force Surrender      | Approver approves force       |
|      |               | Processed            | surrender and surrender value |
|      |               |                      | is \> 10000                   |
+------+---------------+----------------------+-------------------------------+
| 7\.  | Surrender     | Partial Withdrawal   | Partial surrender is          |
|      |               | Letter               | processed successfully        |
+------+---------------+----------------------+-------------------------------+
| 8\.  | Non           | Reduced Paid Up      | Reduced Paid up is            |
|      | Forfeiture    | Acceptance           | successfully processed        |
|      | Option (NFO)  |                      |                               |
+------+---------------+----------------------+-------------------------------+
| 9\.  | Non           | Reduced Paid Up      | Reduced Paid up is rejected   |
|      | Forfeiture    | Rejection            | by the approver               |
|      | Option (NFO)  |                      |                               |
+------+---------------+----------------------+-------------------------------+
| 10\. | Commutation   | Decrease in Sum      | Event is processed            |
|      |               | Assured/Premium      | successfully                  |
|      |               | Accepted Letter      |                               |
+------+---------------+----------------------+-------------------------------+
| 11\. | Commutation   | Decrease in Sum      | Event is rejected             |
|      |               | Assured/Premium      | successfully                  |
|      |               | Rejection Letter     |                               |
+------+---------------+----------------------+-------------------------------+
| 12\. | Loan          | Cover Letter /Loan   | Loan is processed             |
|      |               | Sanction Letter      | successfully                  |
+------+---------------+----------------------+-------------------------------+
| 13\. | Loan          | Loan Rejection       | Loan is rejected              |
|      |               | Letter               |                               |
+------+---------------+----------------------+-------------------------------+
| 14\. | Loan          | Loan Repayment       | Loan has been sanctioned      |
|      |               | Schedule             |                               |
+------+---------------+----------------------+-------------------------------+
| 15\. | Loan          | Default Notice/      | Loan interest is not paid and |
|      |               | Overdue Intimation-  | capitalization event is       |
|      |               | Loan Capitalization  | triggered                     |
|      |               | Letter               |                               |
+------+---------------+----------------------+-------------------------------+
| 16\. | Loan          | Cheque Dishonor for  | Loan repayment cheque is      |
|      |               | Loan                 | dishonored                    |
|      |               |                      |                               |
|      |               | Repayment            |                               |
+------+---------------+----------------------+-------------------------------+
| 17\. | Loan          | Loan Closure Letter  | Loan is paid completely and   |
|      |               |                      | Loan Account is closed        |
+------+---------------+----------------------+-------------------------------+
| 18\. | Loan          | Loan Quote           | Quote Value to be printed     |
|      |               |                      | through the Quote screen      |
+------+---------------+----------------------+-------------------------------+
| 19\. | Freelook      | Freelook             | Freelook is processed         |
|      |               | Cancellation Letter  | successfully/approved by the  |
|      |               |                      | approver                      |
+------+---------------+----------------------+-------------------------------+
| 20\. | Freelook      | Freelook             | Freelook is rejected by the   |
|      |               | Cancellation         | approver                      |
|      |               | Rejection Letter     |                               |
+------+---------------+----------------------+-------------------------------+
| 21\. | Cancellation  | Policy Cancellation  | Policy cancellation request   |
|      |               | Acceptance Letter    | is approved                   |
+------+---------------+----------------------+-------------------------------+
| 22\. | Cancellation  | Policy Cancellation  | Approver rejects the policy   |
|      |               | Rejection Letter     | cancellation request          |
+------+---------------+----------------------+-------------------------------+
| 23\. | Conversion    | Conversion           | Conversion request is         |
|      |               | Acceptance Letter    | approved                      |
+------+---------------+----------------------+-------------------------------+
| 24\. | Conversion    | Conversion Rejection | Conversion request is         |
|      |               | Letter               | rejected                      |
+------+---------------+----------------------+-------------------------------+
| 25\. | Fund Switch   | Fund Switch Letter   | Funds are switched            |
|      |               | acceptance           | successfully                  |
+------+---------------+----------------------+-------------------------------+
| 26\. | Fund Switch   | Fund Switch          | Approver rejects Fund Switch  |
|      |               | Rejection Letter     |                               |
+------+---------------+----------------------+-------------------------------+
| 27\. | Redirection   | Redirection          | Future redirection of funds   |
|      |               | Acceptance Letter    | to be effected                |
+------+---------------+----------------------+-------------------------------+
| 28\. | Redirection   | Redirection          | Approver rejects the          |
|      |               | Rejection Letter     | redirection request           |
+------+---------------+----------------------+-------------------------------+
| 29\. | Billing and   | Premium Due Notice   | - Renewal reminder letter     |
|      | Collection    |                      |   will be generated based on  |
|      |               |                      |   Frequency, specific days    |
|      |               |                      |   before the bill to-date.    |
|      |               |                      |                               |
|      |               |                      | - Status of policies will not |
|      |               |                      |   be generated for monthly    |
|      |               |                      |   frequency policy            |
+------+---------------+----------------------+-------------------------------+
| 30\. | Billing and   | List Bill Due Notice | - List bill will be populated |
|      | Collection    |                      |   in the employer portal by   |
|      |               |                      |   15th of every month         |
|      |               |                      |                               |
|      |               |                      | - Standard letter to be sent  |
|      |               |                      |   by 10th every month         |
|      |               |                      |                               |
|      |               |                      | - The bill details are        |
|      |               |                      |   populated in the employer   |
|      |               |                      |   portal that can be accessed |
|      |               |                      |   only by the employer        |
|      |               |                      |                               |
|      |               |                      | - Reminder must be sent on    |
|      |               |                      |   7th of the next month if    |
|      |               |                      |   premium is not received     |
+------+---------------+----------------------+-------------------------------+
| 31\. | Billing and   | List Bill Due        | Reminder must be sent on 7th  |
|      | Collection    | Reminder             | of the next month to the      |
|      |               |                      | employer if premium is not    |
|      |               |                      | received                      |
+------+---------------+----------------------+-------------------------------+
| 32\. | Billing and   | Payment Failure/     | Cheque bounce details are     |
|      | Collection    | Cheque Bounce Letter | received from the bank        |
+------+---------------+----------------------+-------------------------------+
| 33\. | Nomination    | Nomination Letter    | Will be generated when        |
|      | Letter        |                      | nomination is approved        |
+------+---------------+----------------------+-------------------------------+
| 34\. | Assignment    | Absolute Assignment  | Successful processing of the  |
|      | Letter        | Letter               | absolute assignment           |
+------+---------------+----------------------+-------------------------------+
| 35\. | Assignment    | Conditional          | Successful processing of      |
|      | Letter        | Assignment Letter    | conditional assignment        |
+------+---------------+----------------------+-------------------------------+
| 36\. | Assignment    | Assignment Rejection | Assignment request is         |
|      | Letter        | Letter               | rejected                      |
+------+---------------+----------------------+-------------------------------+
| 37\. | Death Claim   | Notification Letter  | Claim is notified             |
+------+---------------+----------------------+-------------------------------+
| 38\. | Death Claim   | Settlement Letter    | Claim is successfully         |
|      |               |                      | processed                     |
|      |               |                      |                               |
|      |               |                      | Note: In this case, the       |
|      |               |                      | letter will be sent to the    |
|      |               |                      | Payee by the user. If the     |
|      |               |                      | Payee is not mentioned in the |
|      |               |                      | letter, it needs to pick up   |
|      |               |                      | the nominees name and address |
+------+---------------+----------------------+-------------------------------+
| 39\. | Death Claim   | Discharge Voucher    | Sent as part of claim         |
|      |               |                      | settlement letter             |
+------+---------------+----------------------+-------------------------------+
| 40\. | Policy        | Name Change          | Name in the policy is changed |
|      | Servicing     |                      |                               |
+------+---------------+----------------------+-------------------------------+
| 41\. | Policy        | Address Change       | Address in the policy is      |
|      | Servicing     |                      | changed                       |
+------+---------------+----------------------+-------------------------------+
| 42\. | Reinstatement | Revival              | Lapsed policy is revived      |
|      |               | /Reinstatement       | successfully and the Approver |
|      |               | Letter               | approves the                  |
|      |               |                      | revival/reinstatement request |
+------+---------------+----------------------+-------------------------------+
| 43\. | Reinstatement | Reinstatement        | When revival/reinstatement is |
|      |               | Rejection Letter     | rejected                      |
+------+---------------+----------------------+-------------------------------+
| 44\. | Reinstatement | Void Status Letter   | Policy status is changed to   |
|      |               |                      | Void                          |
+------+---------------+----------------------+-------------------------------+
| 45\. | NBF and       | Letter to            | Missing/additional documents  |
|      | Policy        | Customer/Agent for   | are required                  |
|      | Service       | Missing Requirement  |                               |
+------+---------------+----------------------+-------------------------------+
| 46\. | NBF and       | Non-receipt Letter   | Non-receipt of                |
|      | Policy        | to customer/Agent    | missing/additional documents  |
|      | Service       | for Missing          | after 15 days of issue of     |
|      |               | Requirement          | first letter                  |
+------+---------------+----------------------+-------------------------------+
| 47\. | NBF           | Letter to Customer   | Sent before policy issuance   |
|      |               | for Premium Change   | in case premium changes       |
|      |               |                      | (counter offer)               |
+------+---------------+----------------------+-------------------------------+
| 48\. | NBF           | Letter to Customer   | Proposal is declined          |
|      |               | for Proposal         |                               |
|      |               | Rejection            |                               |
+------+---------------+----------------------+-------------------------------+
| 49\. | Reinstatement | Lapse Letter         | Policy is Lapsed              |
+------+---------------+----------------------+-------------------------------+
| 50\. | Survival      | Survival Benefit     | Survival benefit is due for a |
|      | Benefit       | Notification Letter  | policy                        |
+------+---------------+----------------------+-------------------------------+
| 51\. | Survival      | Survival Benefit     | Survival benefit is approved  |
|      | Benefit       | Settlement Letter    | for a policy                  |
+------+---------------+----------------------+-------------------------------+
| 52\. | Survival      | Survival Benefit     | Survival benefit is rejected  |
|      | Benefit       | Settlement Rejection | for a policy                  |
|      |               | Letter               |                               |
+------+---------------+----------------------+-------------------------------+
| 53\. | Billing and   | Electronic Fund      | EFT Transaction is not        |
|      | Collection    | Transfer (EFT)       | successfully processed due to |
|      |               | Transaction          | any reasons                   |
|      |               | Unsuccessful         |                               |
+------+---------------+----------------------+-------------------------------+
| 54\. | Death Claim   | Death Claim          | Death claim has been rejected |
|      |               | Rejection            |                               |
+------+---------------+----------------------+-------------------------------+
| 55\. | Policy        | Premium Payment      | Method has been changed       |
|      | Servicing     | Method Change        | successfully                  |
+------+---------------+----------------------+-------------------------------+
| 56\. | Policy        | Premium Payment      | Approver rejects the request  |
|      | Servicing     | Method Change        |                               |
|      |               | Rejected             |                               |
+------+---------------+----------------------+-------------------------------+
| 57\. | Maturity      | Maturity Claim       | Policy reaches its maturity   |
|      | Claim         | Notification Letter  |                               |
+------+---------------+----------------------+-------------------------------+
| 58\. | Maturity      | Maturity Calculation | Maturity claim is approved by |
|      | Claim         | Sheet and Settlement | the approver                  |
|      |               | Letter               |                               |
+------+---------------+----------------------+-------------------------------+
| 59\. | Maturity      | Maturity Claim       | Maturity claim is rejected by |
|      | Claim         | Rejection Letter     | the approver                  |
+------+---------------+----------------------+-------------------------------+
| 60\. | Surrender     | Surrender            | Approver approves withdrawal  |
|      |               | Withdrawal- Approval |                               |
+------+---------------+----------------------+-------------------------------+
| 61\. | Surrender     | Surrender            | Approver rejects withdrawal   |
|      |               | Withdrawal-          |                               |
|      |               | Rejection            |                               |
+------+---------------+----------------------+-------------------------------+

Elements Common across Letters

+------+-----------------+--------------------------------------------+
| \#   | Type of         | Information                                |
|      | Information     |                                            |
+======+=================+============================================+
| 1    | Header          | This section displays the logo and the     |
|      |                 | return address.                            |
+------+-----------------+--------------------------------------------+
| 2    | Footer          | This section displays the respective       |
|      |                 | Central Processing Center (CPC) address    |
|      |                 | and contact details.                       |
+------+-----------------+--------------------------------------------+
| 3    | Date Format     | In top left corner, the date lines up with |
|      |                 | the company name under the logo and return |
|      |                 | address. Format of date is DD/MM/YYYY. For |
|      |                 | example, 18/03/2012. The same date format  |
|      |                 | is used in the body of letter.             |
+------+-----------------+--------------------------------------------+
| 4    | Amount Format   | Amount starts with the Rs. Symbol and a    |
|      |                 | comma is added after thousand. The amount  |
|      |                 | is always up to two decimal places. For    |
|      |                 | example, Rs.2,54,000.00                    |
+------+-----------------+--------------------------------------------+
| 5    | Closing         | The following standard closing paragraph   |
|      | Paragraph       | is used wherever applicable: If you have   |
|      |                 | any questions, please contact your         |
|      |                 | Representative or our Head Office Member   |
|      |                 | Services Department at \<CSR Contact       |
|      |                 | Number\>.                                  |
|      |                 |                                            |
|      |                 | - If you have any questions, please        |
|      |                 |   contact your Representative or our Head  |
|      |                 |   Office Member Services Department at     |
|      |                 |   1-800-000-0000.                          |
|      |                 |                                            |
|      |                 | Enclosures: If any Signature 1: For the    |
|      |                 | post issue letters -Yours faithfully,      |
|      |                 | Member Services India Post                 |
+------+-----------------+--------------------------------------------+
| 6    | Font            | All letters are in Arial font style with   |
|      |                 | font size 10.                              |
+------+-----------------+--------------------------------------------+
| 7    | Heading         | This displays the Letter Title.            |
+------+-----------------+--------------------------------------------+
| 8    | Salutation      | This includes name of the Person to whom   |
|      |                 | letter is being sent. For example,         |
|      |                 | salutation implies Dear Mr./Ms./Mrs.       |
+------+-----------------+--------------------------------------------+

### Existing Reports

Report Types:

1.  Periodic Reports: These reports are automatically generated daily,
    weekly, monthly, half-yearly or yearly.

2.  Adhoc Reports: These reports are generated manually as and when
    required.

3.  MIS Reports: The application allows users to process various types
    of reports. Some of these reports are used by the higher management
    to take important decisions. These reports are called the Management
    Information System (MIS) reports.

**List of Existing Reports**

+--------------------+--------------------+--------------------+----------------------------------------+
| \#                 | Report Category    | Report Title       | Frequency                              |
+====================+====================+====================+========================================+
| 1                  | New Business       | New Policies       | Daily                                  |
|                    |                    | Procured           |                                        |
+--------------------+--------------------+--------------------+----------------------------------------+
| 2                  | Cancellation       | Cancelled          | Monthly                                |
|                    |                    | Transactions       |                                        |
+--------------------+--------------------+--------------------+----------------------------------------+
| 3                  | Billing and        | NACH-billing       | Monthly HO wise                        |
|                    | Collection         | method change      |                                        |
+--------------------+--------------------+--------------------+----------------------------------------+
| 4                  | DARPAN             | DARPAN Category    | Daily                                  |
|                    |                    | wise Txns report   |                                        |
+--------------------+--------------------+--------------------+----------------------------------------+
| 5                  | Policy             | Circle wise Active | Monthly                                |
|                    |                    | and Inactive       |                                        |
|                    |                    | policies as on     |                                        |
|                    |                    | 1^st^ date of      |                                        |
|                    |                    | Month              |                                        |
+--------------------+--------------------+--------------------+----------------------------------------+
| 6                  | Policy             | Division wise      | Monthly                                |
|                    |                    | Active and Lapsed  |                                        |
|                    |                    | Policies as on     |                                        |
|                    |                    | 1^st^ date of the  |                                        |
|                    |                    | month              |                                        |
+--------------------+--------------------+--------------------+----------------------------------------+
| 7                  | Policy             | HO wise No. of     | Monthly                                |
|                    |                    | policies with      |                                        |
|                    |                    | Aadhar/            |                                        |
|                    |                    | Mobile/e-mail      |                                        |
|                    |                    | updation Monthly   |                                        |
+--------------------+--------------------+--------------------+----------------------------------------+
| 8                  | Customer Portal    | Customer Portal    | Daily                                  |
|                    |                    | Day End collection |                                        |
|                    |                    | report             |                                        |
+--------------------+--------------------+--------------------+----------------------------------------+
| 9                  | Billing and        | Pending Renewal    | Daily and Monthly                      |
|                    | Collection         | Report             |                                        |
+--------------------+--------------------+--------------------+----------------------------------------+
| 10                 | Loan               | Loan installment   | Monthly/Quarterly/Semi-annually/Yearly |
|                    |                    | arrears Report     |                                        |
+--------------------+--------------------+--------------------+----------------------------------------+
| 11                 | Loan               | Loan Interest      | Monthly/Quarterly/Semi-annually/Yearly |
|                    |                    | Recovered Report   |                                        |
+--------------------+--------------------+--------------------+----------------------------------------+
| 12                 | Claim              | Death Claim Report | Monthly/Quarterly/Semi-annually/Yearly |
+--------------------+--------------------+--------------------+----------------------------------------+
| 13                 | Claim, Surrender   | Full Surrender     | Monthly/Quarterly/Semi-annually/Yearly |
|                    |                    | Report, Death      |                                        |
|                    |                    | Claim Report,      |                                        |
|                    |                    | Maturity Claim     |                                        |
|                    |                    | Report             |                                        |
+--------------------+--------------------+--------------------+----------------------------------------+
| 14                 | Claim              | Product Wise Claim | Monthly/Quarterly/Semi-annually/Yearly |
|                    |                    | Paid Report        |                                        |
+--------------------+--------------------+--------------------+----------------------------------------+
| 15                 | Accounts and       | Calculation of     | Adhoc                                  |
|                    | Actuary            |                    |                                        |
|                    |                    | Incurred But Not   |                                        |
|                    |                    |                    |                                        |
|                    |                    | Reported (IBNR)    |                                        |
|                    |                    |                    |                                        |
|                    |                    | for Accounts       |                                        |
|                    |                    |                    |                                        |
|                    |                    | Department and     |                                        |
|                    |                    |                    |                                        |
|                    |                    | Appointed Actuary  |                                        |
+--------------------+--------------------+--------------------+----------------------------------------+
| 16                 | Claim              | Claim Ratio Report | Monthly                                |
+--------------------+--------------------+--------------------+----------------------------------------+
| 17                 | Claim              | Claim Report on    | Daily                                  |
|                    |                    | earned premium     |                                        |
+--------------------+--------------------+--------------------+----------------------------------------+
| 18                 | Claim              | Death Claim        | As frequency could be varied, date     |
|                    |                    | Pending Report     | range would help user to pull the      |
|                    |                    |                    | report as per the requirement          |
+--------------------+--------------------+--------------------+----------------------------------------+
| 19                 | Claim              | Average Claims per | Adhoc                                  |
|                    |                    | Policy Report      |                                        |
+--------------------+--------------------+--------------------+----------------------------------------+
| 20                 | Disbursement       | Disbursement       | Adhoc                                  |
|                    |                    | through NEFT       |                                        |
|                    |                    | Report             |                                        |
+--------------------+--------------------+--------------------+----------------------------------------+
| 21                 | Claim              | Claims             | Monthly/Quarterly/Semi-annually/Yearly |
|                    |                    | Distribution       |                                        |
|                    |                    | Channel Wise       |                                        |
+--------------------+--------------------+--------------------+----------------------------------------+
| 22                 | Profitability      | Product Portfolio  | Adhoc                                  |
|                    |                    | Report             |                                        |
+--------------------+--------------------+--------------------+----------------------------------------+
| 23                 | Profitability      | Profitability      | Yearly                                 |
|                    |                    | Analysis Report    |                                        |
+--------------------+--------------------+--------------------+----------------------------------------+
| 24                 | Profitability      | Risk Analysis      | Monthly                                |
|                    |                    | Report             |                                        |
+--------------------+--------------------+--------------------+----------------------------------------+
| 25                 | New Product        | Key Features of    | Anytime /Adhoc                         |
|                    |                    | product for filing |                                        |
|                    |                    | new product        |                                        |
+--------------------+--------------------+--------------------+----------------------------------------+
| 26                 | Profitability      | Business           | Adhoc                                  |
|                    |                    | Projection for     |                                        |
|                    |                    | Next Five          |                                        |
|                    |                    | Financial Years    |                                        |
+--------------------+--------------------+--------------------+----------------------------------------+
| 27                 | Profitability      | Investment         | Daily                                  |
|                    |                    | Performance Report |                                        |
+--------------------+--------------------+--------------------+----------------------------------------+
| 28                 | Billing and        | Daily Premium      | Daily                                  |
|                    | Collection         | Realization Report |                                        |
+--------------------+--------------------+--------------------+----------------------------------------+
| 29                 | Billing and        | Cheque Bounce      | Daily                                  |
|                    | Collection         | Report             |                                        |
+--------------------+--------------------+--------------------+----------------------------------------+
| 30                 | Billing and        | Premium            | Daily, Date Range Given                |
|                    | Collection         | Realization Report |                                        |
+--------------------+--------------------+--------------------+----------------------------------------+
| 31                 | Billing and        | Advance Premium    | Daily, Date Range Given                |
|                    | Collection         |                    |                                        |
|                    |                    | Realization Report |                                        |
+--------------------+--------------------+--------------------+----------------------------------------+
| 32                 | Profitability      | Branch Product     | Adhoc                                  |
|                    |                    | Portfolio Report   |                                        |
+--------------------+--------------------+--------------------+----------------------------------------+
| 33                 | Profitability      | Key Parameters     | Adhoc                                  |
|                    |                    | Report             |                                        |
+--------------------+--------------------+--------------------+----------------------------------------+
| 34                 | Turn Around        | Grievance Turn     | Daily                                  |
|                    |                    | Around Time Report |                                        |
|                    | Time (TAT)         |                    |                                        |
+--------------------+--------------------+--------------------+----------------------------------------+
| 35                 | Turn Around Time   | Grievance Analysis | Daily                                  |
|                    | (TAT)              | Report             |                                        |
+--------------------+--------------------+--------------------+----------------------------------------+
| 36                 | Profitability      | Portfolio          | Yearly                                 |
|                    |                    | Profitability      |                                        |
|                    |                    | Report             |                                        |
+--------------------+--------------------+--------------------+----------------------------------------+
| 37                 | Claim              | Claim Register     | Daily/Monthly/Quarterly/Semi           |
|                    |                    | Report             | annually/Yearly                        |
+--------------------+--------------------+--------------------+----------------------------------------+
| MIS Reports                                                                                           |
+--------------------+--------------------+--------------------+----------------------------------------+
| 38                 | New Business       | New Business in    | Daily/ Weekly/ Monthly/ Quarterly/     |
|                    |                    | Forced Report      | Semi-annually/Yearly                   |
+--------------------+--------------------+--------------------+----------------------------------------+
| 39                 | DARPAN             | DARPAN MIS- upto   | Daily                                  |
|                    |                    | BO level           |                                        |
+--------------------+--------------------+--------------------+----------------------------------------+
| 40                 | Billing and        | Daily Premium      | Daily                                  |
|                    | Collection         | Collection Report  |                                        |
|                    |                    | Daily Cheque       |                                        |
|                    |                    | Realization Report |                                        |
+--------------------+--------------------+--------------------+----------------------------------------+
| 41                 | New Business       | Policy pending     | Daily                                  |
|                    |                    | Enforcement Report |                                        |
+--------------------+--------------------+--------------------+----------------------------------------+
| 42                 | Claim              | Report - Claims    | Daily                                  |
|                    |                    | Logged             |                                        |
+--------------------+--------------------+--------------------+----------------------------------------+
| 43                 | Claim              | Report - Claims    | Daily/ Weekly/ Monthly/ Quarterly/     |
|                    |                    | Settled Death      | Semi-annually/Yearly                   |
|                    |                    | Report - Claims    |                                        |
|                    |                    | Settled Maturity   |                                        |
+--------------------+--------------------+--------------------+----------------------------------------+
| 44                 | Loan               | Report - Loan      | Daily/ Weekly/ Monthly/ Quarterly/     |
|                    |                    | Disbursed          | Semi-annually/Yearly                   |
+--------------------+--------------------+--------------------+----------------------------------------+
| 45                 | Assignment         | Report - Policy    | Monthly/ Quarterly/                    |
|                    |                    | Assignment-        | Semi-annually/Yearly                   |
|                    |                    | Absolute Report -  |                                        |
|                    |                    | Policy Assignment- |                                        |
|                    |                    | Conditional        |                                        |
+--------------------+--------------------+--------------------+----------------------------------------+
| 46                 | Billing and        | Premium            | Daily/ Weekly/ Monthly/ Quarterly/     |
|                    | Collection         | Accountability     | Semi-annually/Yearly                   |
+--------------------+--------------------+--------------------+----------------------------------------+
| 47                 | Lapsation          | Report - Lapse     | Daily/ Weekly/ Monthly/ Quarterly/     |
|                    |                    | Events             | Semi-annually/Yearly                   |
+--------------------+--------------------+--------------------+----------------------------------------+
| 48                 | Surrender          | Report - Full      | Daily/ Weekly/ Monthly/ Quarterly/     |
|                    |                    | Surrender Report - |                                        |
|                    |                    | Partial Surrender  | Semi-annually/Yearly                   |
+--------------------+--------------------+--------------------+----------------------------------------+
| 49                 | Reinsurance        | Report - Policy    | Daily/ Weekly/ Monthly/ Quarterly/     |
|                    |                    | Reinsurance        | Semi-annually/Yearly                   |
+--------------------+--------------------+--------------------+----------------------------------------+
| 50                 | Unit Linked        | Report -Fund       | Daily                                  |
|                    | Insurance Policy   | Details            |                                        |
|                    | (ULIP)             |                    |                                        |
+--------------------+--------------------+--------------------+----------------------------------------+
| 51                 | Billing and        | Premium and        | Daily                                  |
|                    | Collection         | Documentation      |                                        |
|                    |                    | Received Report    |                                        |
+--------------------+--------------------+--------------------+----------------------------------------+
| 52                 | Conversion         | Report - Policy    | Monthly                                |
|                    |                    | Conversion         |                                        |
+--------------------+--------------------+--------------------+----------------------------------------+
| 53                 | TAT                | Report -TAT (Turn  | Monthly                                |
|                    |                    |                    |                                        |
|                    |                    | Around time)       |                                        |
+--------------------+--------------------+--------------------+----------------------------------------+
| 54                 | New Business       | BO Level           | Adhoc                                  |
|                    |                    | Classification of  |                                        |
|                    |                    | New Business       |                                        |
+--------------------+--------------------+--------------------+----------------------------------------+
| 55                 | New Business       | Reason Wise        | Adhoc                                  |
|                    |                    | Freelook           |                                        |
|                    |                    | Cancellation       |                                        |
|                    |                    | Report             |                                        |
+--------------------+--------------------+--------------------+----------------------------------------+
| 56                 | New Business       | Policy             | Adhoc                                  |
|                    |                    | Cancellation       |                                        |
|                    |                    | Report             |                                        |
+--------------------+--------------------+--------------------+----------------------------------------+

Elements Common across Reports

+------+-----------------+--------------------------------------------+
| \#   | Type of         | Information                                |
|      | Information     |                                            |
+======+=================+============================================+
| 1    | Company Name    | This section displays the text Postal Life |
|      |                 | Insurance and logo.                        |
+------+-----------------+--------------------------------------------+
| 2    | Header          | This section contains the following        |
|      |                 | information:                               |
|      |                 |                                            |
|      |                 | - Report Name                              |
|      |                 |                                            |
|      |                 | - Process Date: - \<report start date\> -  |
|      |                 |   \<report end date\>                      |
+------+-----------------+--------------------------------------------+
| 3    | Footer          | This section contains the following        |
|      |                 | elements:                                  |
|      |                 |                                            |
|      |                 | - \< Current Date/time\> left justified    |
|      |                 |   along the bottom                         |
|      |                 |                                            |
|      |                 | - Page \<page number\> of \< number of     |
|      |                 |   pages\> right justified along the bottom |
+------+-----------------+--------------------------------------------+

## **4. Functional Requirements Specification**

### 4.1 Create Reports Page

> Clicking 'Create Reports' button should open this page. This page
> allows users to generate reports using filters.

- **Fields:**

  - Report Type: Dropdown (PDF, Excel).

  - Policy Number: Text: Option to search & select policy should be
    present.

  - Customer ID: Text: Option to search & select Customer should be
    present.

  - Report Start Date / End Date: Calendar fields.

  - Office Type / Location: Dropdowns.

  - Generate Report: Button to initiate report generation.

### 4.2 Letters Page

> This page stores and displays all letters generated by the agent.

- **Fields:**

  - Search Bar: Text input to search by Agent Name or Agent ID

  - Letter Type Filter: Dropdown (e.g., Welcome, Appointment,
    Termination)

  - Date Range: Two calendar fields (Start Date, End Date)

  - Policy Number: Text: Option to search & select policy should be
    present.

  - Customer ID: Text: Option to search & select Customer should be
    present.

  - Letter Type: Display column

  - Generated Date: Display column

  - View Letter: Button/icon to open/download the letter

  - Download Format: Option to download as PDF

  - Pagination Controls: For navigating through records

### 4.3 Historical Reports Page

> This page stores and displays all reports generated by the user. It
> should be limited to last 10 generated Reports by the respective user.

- **Fields:**

  - Search Bar: Text input to search by Report Title

  - Report Category Filter: Dropdown

  - Date Range: Two calendar fields (Start Date, End Date)

  - Report Title: Display column

  - Report Category: Display column

  - Generated Date: Display column

  - Format: Display column (PDF/Excel)

  - View Report: Button/icon to open/download the report

  - Pagination Controls: For navigating through records

## **5. Samples:**

**[Existing Core Insurance Solution reports available at utilities
portal of PLI:-]{.underline}**

1.  **Daily Report -New Policies Procured :**

![A screenshot of a computer AI-generated content may be
incorrect.](media/image1.png){width="7.020833333333333in"
height="3.6041666666666665in"}

2.  **Cancelled Transactions-Monthly report :** ![A white background
    with many small black and orange letters AI-generated content may be
    incorrect.](media/image2.png){width="6.251850393700788in"
    height="3.5635542432195977in"}

3.  ![A screenshot of a computer AI-generated content may be
    incorrect.](media/image3.png){width="6.552083333333333in"
    height="3.4375in"}**NACH-billing method change-Monthly HO wise :**

4.  **DARPAN MIS- upto BO level :**

5.  **DARPAN Category wise Txns report :**

![A screenshot of a computer AI-generated content may be
incorrect.](media/image4.png){width="7.197916666666667in"
height="2.36875in"}

6.  ![A close-up of a document AI-generated content may be
    incorrect.](media/image5.png){width="7.614583333333333in"
    height="4.041666666666667in"}**Circle wise Active and Inactive
    policies as on 1^st^ date of Month:**

7.  **Division wise Active and Lapsed Policies as on 1^st^ date of the
    month:**

![](media/image6.png){width="6.59375in" height="3.2291666666666665in"}

8.  ![A screenshot of a spreadsheet AI-generated content may be
    incorrect.](media/image7.png){width="7.104166666666667in"
    height="3.1354166666666665in"}**HO wise No. of policies with Aadhar/
    Mobile/e-mail updation Monthly:**

- **Reports available through front end McCamish:-**

<!-- -->

- The existing reports available at McCamish in two types i.e. detailed
  and consolidated

- In PDF and Excel format

- Access of report given hierarchy wise Circle-Region-Division

- Screenshot of Reports module available in McCamish is attached
  herewith:

![](media/image8.png)![](media/image9.png)![](media/image10.png)![A
screenshot of a computer AI-generated content may be
incorrect.](media/image8.jpeg){width="7.0in"
height="3.8958333333333335in"}![](media/image12.png)

![](media/image13.png)![](media/image14.png) ![A screenshot of a
computer AI-generated content may be
incorrect.](media/image9.jpeg){width="5.833333333333333in"
height="3.2121456692913384in"}

[Customer Portal Day End collection report:]{.underline}

![A screenshot of a document AI-generated content may be
incorrect.](media/image11.png){width="5.7799136045494315in"
height="3.640702099737533in"}

> ![](media/image17.png)![](media/image18.png)![](media/image18.png)![](media/image19.png)![](media/image20.png)![A
> screenshot of a computer AI-generated content may be
> incorrect.](media/image12.jpeg){width="6.791666666666667in"
> height="3.7708333333333335in"}

![](media/image22.png)![](media/image17.png)![](media/image17.png)![](media/image23.png)![](media/image24.png)![A
screenshot of a computer AI-generated content may be
incorrect.](media/image13.jpeg){width="7.15625in"
height="3.875in"}Average Claims per Policy report : ![A close-up of a
document AI-generated content may be
incorrect.](media/image15.png){width="6.5in"
height="4.076388888888889in"}

> ![A screenshot of a computer AI-generated content may be
> incorrect.](media/image16.png){width="6.263888888888889in"
> height="3.4791666666666665in"}**[Consolidated Disbursement through
> NEFT Monthly Circle wise]{.underline}**:

![A screenshot of a computer AI-generated content may be
incorrect.](media/image17.jpeg){width="7.041666666666667in"
height="3.5416666666666665in"}

> ![](media/image29.png)![](media/image30.png)
>
> ![A screenshot of a computer AI-generated content may be
> incorrect.](media/image18.jpeg){width="7.239583333333333in"
> height="3.5729166666666665in"}![](media/image32.png)![](media/image33.png)![](media/image34.png)![](media/image17.png)
>
> ![A screenshot of a computer AI-generated content may be
> incorrect.](media/image19.jpeg){width="7.041666666666667in"
> height="4.020833333333333in"}

![](media/image36.png)![](media/image37.png)

![](media/image38.png)![](media/image39.png)![A screenshot of a computer
AI-generated content may be
incorrect.](media/image20.jpeg){width="6.854166666666667in"
height="3.46875in"}

![](media/image41.png)![](media/image42.png)![A screenshot of a computer
AI-generated content may be
incorrect.](media/image21.jpeg){width="6.916666666666667in"
height="3.6875in"}

Reports available-Adhoc Scheduling:

![A screenshot of a computer AI-generated content may be
incorrect.](media/image22.jpeg){width="6.916666666666667in"
height="3.0208333333333335in"}

![A screenshot of a computer AI-generated content may be
incorrect.](media/image23.jpeg){width="6.59375in"
height="3.4583333333333335in"}
