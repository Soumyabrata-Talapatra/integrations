> INTERNAL APPROVAL FORM

**Project Name:** Post Office Savings Bank (POSB)

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
[6](#functional-requirements-specification)](#functional-requirements-specification)

[4.1 Billing Method Change by In-Person Visit to Post office
[6](#billing-method-change-by-in-person-visit-to-post-office)](#billing-method-change-by-in-person-visit-to-post-office)

[4.2 Billing Method Change through Digital Mode
[7](#billing-method-change-through-digital-mode)](#billing-method-change-through-digital-mode)

[**5. Attachments** [9](#attachments)](#attachments)

## **1. Executive Summary**

This document defines the requirements for integrating the Postal Life
Insurance (PLI) Insurance Management System (IMS) with Post Office
Savings Bank (POSB) to enable automated premium payments and loan
repayments through Standing Instructions (SI), Mobile Banking
(m-Banking), and Online Banking (e-Banking). The integration will allow
customers to change their billing method from existing modes to
POSB-based digital or assisted channels.

## **2. Project Scope**

This Integration aims to:

- Facilitate billing method change and premium collection through POSB
  accounts.

- Support both in-person and digital modes for SI, m-Banking, and
  e-Banking.

- Ensure secure, real-time integration between POSB and IMS.

- Enhance customer experience with notifications, multilingual support,
  and document submission.

## **3. Business Requirements**

+-------------+-----------------------+---------------------------------------+
| **ID**      | **Functionality**     | **Requirements**                      |
+=============+=======================+=======================================+
| FS_POSB_001 | Enable Billing Method | The system must allow customers to    |
|             | Change                | change their existing billing method  |
|             |                       | to:                                   |
|             |                       |                                       |
|             |                       | - Standing Instruction (SI) via POSB  |
|             |                       |   account.                            |
|             |                       |                                       |
|             |                       | - m-Banking option.                   |
|             |                       |                                       |
|             |                       | - e-Banking option.                   |
+-------------+-----------------------+---------------------------------------+
| FS_POSB_002 | Support Multiple      | Customers should be able to initiate  |
|             | Channels              | billing method change:                |
|             |                       |                                       |
|             |                       | - **In-person** at Post Office        |
|             |                       |   (mandatory for SI).                 |
|             |                       |                                       |
|             |                       | - **Digitally** via Mobile Banking    |
|             |                       |   (m-Banking) or Online Banking       |
|             |                       |   (e-Banking).                        |
+-------------+-----------------------+---------------------------------------+
| FS_POSB_003 | Standing Instruction  | The system must support SI            |
|             | Registration          | registration through Post Office      |
|             |                       | counters. SI must be linked to the    |
|             |                       | customer's POSB account and PLI/RPLI  |
|             |                       | policy.                               |
+-------------+-----------------------+---------------------------------------+
| FS_POSB_004 | Digital Payment       | Enable premium payment and loan       |
|             | Processing            | repayment through:                    |
|             |                       |                                       |
|             |                       | - India Post Mobile Banking App.      |
|             |                       |                                       |
|             |                       | - India Post e-Banking Portal.        |
+-------------+-----------------------+---------------------------------------+
| FS_POSB_005 | Real-Time Account     | Integrate POSB API for:               |
|             | verification          |                                       |
|             |                       | - Account number validation.          |
|             |                       |                                       |
|             |                       | - Signature verification.             |
|             |                       |                                       |
|             |                       | - Transaction authentication.         |
+-------------+-----------------------+---------------------------------------+
| FS_POSB_006 | Policy Validation     | IMS must validate policy details      |
|             |                       | (policy number, premium amount, due   |
|             |                       | date) before processing payment.      |
+-------------+-----------------------+---------------------------------------+
| FS_POSB_007 | Transaction           | Display confirmation page with:       |
|             | Confirmation          |                                       |
|             |                       | - Policy details.                     |
|             |                       |                                       |
|             |                       | - Premium amount.                     |
|             |                       |                                       |
|             |                       | - Taxes/rebates.                      |
|             |                       |                                       |
|             |                       | - Total payable amount.               |
+-------------+-----------------------+---------------------------------------+
| FS_POSB_008 | Payment Outcome       | Notify customers of transaction       |
|             | Notification          | status (Success/Failure) via:         |
|             |                       |                                       |
|             |                       | - SMS.                                |
|             |                       |                                       |
|             |                       | - Email.                              |
|             |                       |                                       |
|             |                       | - WhatsApp.                           |
|             |                       |                                       |
|             |                       | - Hard copy (for SI).                 |
+-------------+-----------------------+---------------------------------------+
| FS_POSB_009 | Receipt Generation    | Generate and share digital receipts   |
|             |                       | for successful transactions. Provide  |
|             |                       | duplicate receipt option via IMS and  |
|             |                       | banking channels.                     |
+-------------+-----------------------+---------------------------------------+
| FS_POSB_010 | Refund/Reconciliation | Handle failed transactions with       |
|             |                       | refund and reconciliation by Nodal    |
|             |                       | office.                               |
+-------------+-----------------------+---------------------------------------+
| FS_POSB_011 | Auto-Renewal of SI    | Enable SI auto-renewal post-expiry    |
|             |                       | with customer consent.                |
+-------------+-----------------------+---------------------------------------+
| FS_POSB_012 | Notifications &       | Push notifications for:               |
|             | Alerts                |                                       |
|             |                       | - SI activation.                      |
|             |                       |                                       |
|             |                       | - Upcoming debit reminders.           |
|             |                       |                                       |
|             |                       | - Expiry alerts.                      |
|             |                       |                                       |
|             |                       | - Premium due alerts.                 |
|             |                       |                                       |
|             |                       | - Missed payment notifications.       |
+-------------+-----------------------+---------------------------------------+
| FS_POSB_013 | Dashboard Visibility  | Display summary of active SIs in POSB |
|             |                       | m-Banking and e-Banking dashboards.   |
+-------------+-----------------------+---------------------------------------+
| FS_POSB_014 | Document Submission   | Support DigiLocker-based document     |
|             |                       | upload for claims or service          |
|             |                       | requests.                             |
+-------------+-----------------------+---------------------------------------+
| FS_POSB_015 | Multi-Language        | Provide language options via Bhashini |
|             | Support               | for rural customers.                  |
+-------------+-----------------------+---------------------------------------+
| FS_POSB_016 | Offline Mode          | Enable offline transaction initiation |
|             |                       | with auto-sync when connectivity is   |
|             |                       | restored.                             |
+-------------+-----------------------+---------------------------------------+

## **4. Functional Requirements Specification**

**Use Case Diagram for Integration with POSB:**
![](media/image1.png){width="6.268055555555556in"
height="4.178472222222222in"}

The Customer may choose either of the following options to change the
existing billing method to Standing Instruction (SI), m-Banking, or
e-Banking through their POSB account ---

1.  By visiting the Post Office in person, or

2.  Through digital mode (without visiting the Post Office) using online
    or mobile banking facilities.

\*\* Standing Instructions may be opt by visiting the Post Office in
person only

### 4.1 Billing Method Change by In-Person Visit to Post office

- **Purpose:** To allow customers to change billing method for Renewal
  Payment or Loan Repayment on its policy through in-person visit to
  post office using its POSB Account.

- **Flow Chart for Standing Instructions:**

**Customer Visits Post Office for Billing Method Change**

**Selects Standing Instruction (SI) via POSB Account**

**Submits Form and Documents**

**Official Registers Request in System**

**System Auto-populates and Verifies Account Details**

**POSB API Verification of Account and Signature**

**If Verified → Request Accepted \| If Error → Rejected**

**System Generates Acceptance/Rejection Letter**

**Customer Informed via SMS / Email / WhatsApp / Hard Copy**

### 4.2 Billing Method Change through Digital Mode

#### 4.2.1 Billing Method Change through Mobile Banking

- **Purpose:** To allow customers to change billing method for Renewal
  Payment or Loan Repayment on its policy through M-Banking App of POSB.

- **Flow Chart for M-banking App:**

**Customer logs in to the India Post Mobile Banking App**

**Customer selects "*requests*" tab from the m-banking dashboard**

**Customer chooses "*Pay PLI/RPLI renewal Premium/Loan Repayment*"**

**Customer enter details i.e. Policy Number, No. of Instalments, Debit
Account No., and transaction remarks etc.{then click on submit}**

**Confirmation page**

**{System display details i.e. Policy No., Account No., Policyholder
Name, Premium, Dates, Taxes (if applicable), rebate (if applicable) &
final total premium amount etc.}**

**Verification**

**{Customer verifies all details and enter transaction password for
confirmation}**

**Transaction Processing**

**{System validates payment through POSB integration}**

**Payment Outcome**

**{Success -- Receipt Generated & shared via e-Mail, SMS, Whatsapp
etc.}**

**{Failure- Error message & refund/reconciliation by Nodal office}**

**\* *Duplicate receipt Option:- Customer can generate duplicate receipt
via m-Banking and IMS***

#### 4.2.1 Billing Method Change through e-Banking

- **Purpose:** To allow customers to change billing method for Renewal
  Payment or Loan Repayment on its policy through e-Banking.

- **Flow Chart for e-Banking:**

- **Customer logs in to the India Post e-Banking Portal**

- 

- **Customer selects " General Services" tab from the e-banking
  dashboard**

- **{Customer then selects, "Service requests" then select " New Service
  Request" option from Dropdown and clicks Ok**

- 

- **Customer then clicks on "*Pay PLI/RPLI renewal Premium/Loan
  Repayment*" from available request type**

- 

- **Customer enter details i.e. Policy Number, No. of Instalments, Debit
  Account No. etc. and transaction remarks {then click on submit}**

- 

- **Confirmation page**

- **{System display details i.e. Policy No., Account No., Policyholder
  Name, Premium, Dates, Taxes (if applicable), rebate (if applicable)
  and final total premium amount etc.}**

- 

- **Verification**

- **{Customer verifies all details and enter transaction password for
  confirmation}**

- 

- **Transaction Processing**

- **{System validates payment through POSB integration}**

- 

- **Payment Outcome**

- **{Success -- Receipt Generated & shared via e-Mail, SMS, WhatsApp
  etc.}**

- **{Failure- Error message & refund/reconciliation by Nodal office}**

- 

- **\* *Duplicate receipt Option: - Customer can generate duplicate
  receipt via m-Banking and IMS***

## **5. Attachments**

The following documents can be referred.
