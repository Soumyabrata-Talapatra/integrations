> INTERNAL APPROVAL FORM

**Project Name:** UMANG Mobile App

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
[5](#functional-requirements-specification)](#functional-requirements-specification)

[4.1 Policy Dashboard page
[5](#policy-dashboard-page)](#policy-dashboard-page)

[4.2 Premium Payment Page
[6](#premium-payment-page)](#premium-payment-page)

[4.3 Service Request page
[6](#service-request-page)](#service-request-page)

[4.4 Policy Purchase page
[6](#policy-purchase-page)](#policy-purchase-page)

[4.5 Language Settings [7](#language-settings)](#language-settings)

[4.6 DigiLocker Upload Page
[7](#digilocker-upload-page)](#digilocker-upload-page)

[**5. Attachments** [7](#attachments)](#attachments)

## **1. Executive Summary**

This document outlines the requirement for integrating the UMANG mobile
application with the Insurance Management System (IMS) of India Post
PLI/RPLI. The integration aims to provide seamless digital access to
insurance services for customers via UMANG.

## **2. Project Scope**

This scope will include the integration of Insurance Management System
with UMANG Application.

## **3. Business Requirements**

  -----------------------------------------------------------------------
  **ID**      **Requirements**
  ----------- -----------------------------------------------------------
  FS_UM_001   Enable PLI/RPLI customers to view policy details via UMANG.

  FS_UM_002   Allow digital premium payments through UMANG using UPI,
              cards, net banking.

  FS_UM_003   Provide service request and grievance submission and
              tracking.

  FS_UM_004   Facilitate new policy purchase and product comparison.

  FS_UM_005   Support multilingual interface via Bhashini for rural
              accessibility.

  FS_UM_006   Integrate DigiLocker for document submission during policy
              purchase or claim.

  FS_UM_007   Enable e-Bond generation for new policies.

  FS_UM_008   Provide reminders for premium due dates and default alerts.

  FS_UM_009   Ensure secure, real-time data exchange between UMANG and
              IMS.

  FS_UM_010   Allow download of policy documents and certificates.
  -----------------------------------------------------------------------

## **4. Functional Requirements Specification**

**Use Case Diagram for UMANG Integration with IMS:**
![](media/image1.png){width="6.268055555555556in"
height="4.178472222222222in"}

### 4.1 Policy Dashboard page

- **Purpose:** Policy Dashboard in UMANG App will display policy details
  from IMS via secure API and it also allow document downloads in PDF.
  Customers can view active/inactive policies with full details (sum
  assured, premium, term, bonus) including accrued bonus, download
  policy document, certificates including income tax certificates, and
  track proposals, policies, and claims.

- **Fields:**

  - Policy Number: Text: Unique identifier of the policy

  - Policy Status: Dropdown: Active / Inactive

  - Sum Assured: Numeric: Total coverage amount

  - Premium Amount: Numeric: Monthly/Quarterly/Annual premium

  - Term: Numeric: Policy duration in years

  - Accrued Bonus: Numeric: Bonus accumulated till date

  - Download Documents: Button: Downloads policy certificate and IT
    certificate

### 4.2 Premium Payment Page

- **Purpose:** Premium Payment page will Integrate with Bharat BillPay
  or payment gateway for performing premium payment for PLI/RPLI
  Policies. It should Trigger payment confirmation to IMS and Update
  payment status in IMS. The insured may also set up an auto-debit
  facility for future premium payments. Additionally, the UMANG mobile
  app can send timely reminders for upcoming premium due dates and
  alerts for any applicable default fees. Pay premiums via UPI,
  debit/credit cards, or net banking etc.

- **Fields:**

  - Policy Number: Text: Auto-filled from dashboard

  - Due Amount: Numeric: Premium due

  - Payment Mode: Dropdown: UPI / Debit Card / Credit Card / Net Banking

  - Auto-debit Setup: Checkbox: Enable auto-debit for future payments

  - Reminder Notification: Toggle: Enable SMS/Push alerts for due dates

### 4.3 Service Request page

- **Purpose:** Service Request Page will allow the user to Submit
  request to Insurance Management System and Generate ticket ID and
  allow status tracking. Users can submit service requests and
  grievances through UMANG and track their status via IT 2.0 and CRM
  integration with UMANG.

- **Fields:**

  - Request Type: Dropdown: Address Change / Nominee Update / etc.

  - Description: Text Area: Details of the request

  - Upload Supporting Docs: File Upload: Optional document upload

  - Track Status: Button: View current status of request

### 4.4 Policy Purchase page

- **Purpose:** Policy Purchase page should allow a user to issue the
  policy digitally. It should Suggest plans based on age/income, allow
  policy procurement, Integrate with DigiLocker for document
  verification and Generate e-Bond and send to Insurance Management
  System.

- **Fields:**

  - Customer Name: Text: Full name of applicant

  - Age: Numeric: Age of applicant

  - Income: Numeric: Annual income of applicant

  - Preferred Plan: Dropdown: Suggested based on input

  - Compare Plans: Button: Show comparison of PLI/RPLI plans

  - Upload KYC Docs: File Upload: Via DigiLocker

  - Generate e-Bond: Button: Create e-Bond document for policy

### 4.5 Language Settings

- **Purpose:** The Content should get displayed in other languages also
  if selected by the user in the UMANG Application.

- **Fields:**

  - Preferred Language: Dropdown: Hindi / Tamil / Bengali / etc.

  - Accessibility Mode: Toggle: Enable simplified UI for rural users

### 4.6 DigiLocker Upload Page

- **Purpose:** User should be able to Authenticate via DigiLocker, Fetch
  verified documents and Submit to Insurance Management System for
  policy or claim processing.

- **Fields:**

  - Document Type: Dropdown: Aadhaar / PAN / Income Proof / etc.

  - Upload via DigiLocker: Button: Launch DigiLocker picker for document
    upload

  - Status: Text: Uploaded / Pending

## **5. Attachments**

The following documents can be referred.
