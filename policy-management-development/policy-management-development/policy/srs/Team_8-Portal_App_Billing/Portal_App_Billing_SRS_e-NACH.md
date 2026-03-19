> INTERNAL APPROVAL FORM

**Project Name:** e-NACH

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

[4.1 Customer Initiation Page
[6](#customer-initiation-page)](#customer-initiation-page)

[4.2 Mandate Registration Page
[6](#mandate-registration-page)](#mandate-registration-page)

[4.3 Mandate Status Page
[6](#mandate-status-page)](#mandate-status-page)

[4.4 Premium Deduction Dashboard
[7](#premium-deduction-dashboard)](#premium-deduction-dashboard)

[4.5 Notifications & Alerts Page
[7](#notifications-alerts-page)](#notifications-alerts-page)

[4.6 Exception Handling Page
[7](#exception-handling-page)](#exception-handling-page)

[4.7 Admin Dashboard [7](#admin-dashboard)](#admin-dashboard)

[**5. Attachments** [8](#attachments)](#attachments)

## **1. Executive Summary**

The purpose of the e-NACH module is to enable automated, secure, and
paperless premium collection for PLI policies via NPCI's e-NACH
platform. This module will streamline recurring payments, reduce manual
errors, and enhance customer convenience.

## **2. Project Scope**

This module will integrate with the IMS (APT 2.0 / McCamish) and NPCI's
e-NACH platform to facilitate mandate registration, authentication,
recurring premium deductions, and exception handling.

## **3. Business Requirements**

  --------------------------------------------------------------------------------
  **ID**        **Functionality**   **Requirements**
  ------------- ------------------- ----------------------------------------------
  FS_NACH_001   Mandate             During policy purchase or renewal, customers
                Registration        should be able to opt for e-NACH, provide bank
                                    details, and authenticate the mandate using
                                    Net Banking, Debit Card, or Aadhaar, with the
                                    status updated in the Insurance Management
                                    System (IMS).

  FS_NACH_002   Premium Collection  Premiums should be auto-debited on the due
                                    date for active mandates, followed by
                                    confirmation updates in the policy record and
                                    generation of premium receipts.

  FS_NACH_003   Notifications       Customers must receive SMS or email
                                    notifications regarding mandate registration
                                    status and premium deduction outcomes.

  FS_NACH_004   Exception Handling  The system should detect and log failed
                                    transactions, notify customers with reasons,
                                    and allow re-initiation of mandates or
                                    payments.

  FS_NACH_005   Admin & Reporting   Admin users should have access to dashboards
                                    for mandate tracking, reconciliation reports,
                                    and audit logs to ensure compliance.
  --------------------------------------------------------------------------------

## **4. Functional Requirements Specification**

**Visual Flowchart for the e-NACH Process:** ![A screenshot of a
computer screen AI-generated content may be
incorrect.](media/image1.png){width="6.024798775153106in"
height="9.037199256342957in"}

### 4.1 Customer Initiation Page

- **Purpose:** Allow customer to opt for e-NACH during policy purchase
  or renewal.

- **Fields:**

  - Policy Number (Auto-populated from IMS)

  - Customer Name (Auto-populated)

  - Mobile Number (Editable for verification)

  - Email ID (Editable for notifications)

  - Opt-in Checkbox ("I agree to register for e-NACH")

  - Preferred Channel (Dropdown: Portal / Mobile App / Counter)

  - Submit Button (Triggers mandate registration process)

### 4.2 Mandate Registration Page

- **Purpose:** Capture bank details and initiate mandate request.

- **Fields:**

  - Policy Number (Read-only)

  - Customer Name (Read-only)

  - Bank Name (Dropdown with search)

  - IFSC Code (Auto-validated)

  - Account Number (Masked input for security)

  - Account Type (Radio: Savings / Current)

  - Authentication Mode (Radio: Net Banking / Debit Card / Aadhaar)

  - NPCI Consent Checkbox (Mandatory before proceeding)

  - Proceed to Authentication Button (Redirects to NPCI page)

### 4.3 Mandate Status Page

- **Purpose:** Display current status of mandate.

- **Fields:**

  - Policy Number

  - Mandate Reference Number (MRN)

  - Status (Active / Pending / Rejected)

  - Date of Registration

  - Bank Response Details (Reason for rejection if any)

  - Download Mandate Confirmation (PDF)

### 4.4 Premium Deduction Dashboard

- **Purpose:** Show upcoming and completed premium deductions.

- **Fields:**

  - Policy Number

  - Due Date

  - Premium Amount

  - Transaction Status (Success / Failed)

  - NPCI Transaction ID

  - Receipt Download Link

### 4.5 Notifications & Alerts Page

- **Purpose:** Manage SMS/Email notifications.

- **Fields:**

  - Policy Number

  - Customer Contact Details

  - Notification Type (Mandate Confirmation / Premium Deduction /
    Failure Alert)

  - Status (Sent / Pending)

  - Resend Option

### 4.6 Exception Handling Page

- **Purpose:** Handle failed transactions and re-initiation.

- **Fields:**

  - Policy Number

  - Failure Reason (Insufficient Balance / Expired Mandate / Bank
    Rejection)

  - Retry Option (Button to re-initiate)

  - Alternative Payment Option (UPI / Net Banking)

  - Customer Acknowledgement Checkbox

### 4.7 Admin Dashboard

- **Purpose:** For PLI staff to monitor and reconcile.

- **Fields:**

  - Total Mandates (Active / Pending / Rejected)

  - Total Premium Collections

  - Failed Transactions Count

  - Reconciliation Report Download

  - Search by Policy Number / MRN

## **5. Attachments**

The following documents can be referred.
