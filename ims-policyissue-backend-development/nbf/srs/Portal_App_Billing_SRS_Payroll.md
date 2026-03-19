> INTERNAL APPROVAL FORM

**Project Name:** Payroll

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

[4.1 Dashboard Page [5](#dashboard-page)](#dashboard-page)

[4.2 Opt-in for Pay Recovery Page
[6](#opt-in-for-pay-recovery-page)](#opt-in-for-pay-recovery-page)

[4.3 Request Verification Page (Postal DDO Only)
[7](#request-verification-page-postal-ddo-only)](#request-verification-page-postal-ddo-only)

[4.4 Success/Failure Page
[7](#successfailure-page)](#successfailure-page)

[4.5 Demand file Management Page
[8](#demand-file-management-page)](#demand-file-management-page)

[4.6 Bulk Upload Page [8](#bulk-upload-page)](#bulk-upload-page)

[**5. Attachments** [8](#attachments)](#attachments)

## **1. Executive Summary**

The Payroll Module will enable premium collection for PLI policies
through salary deductions for Department of Posts (DoP) employees who
opt for the Pay Recovery method. It will integrate with the Employee
Self Service (ESS) Portal (APT 2.0) and the Insurance Management System
(IMS) to ensure seamless data flow, verification, and transaction
processing.

## **2. Project Scope**

The scope of the module is:

- Employees submit policy details via ESS Portal/App.

- Postal DDO verifies and configures payroll changes.

- Payroll system deducts premiums based on IMS demand file.

- API-based integration between Payroll Module and IMS for real-time
  updates.

- Notifications to employees for success/failure cases.

## **3. Business Requirements**

  -----------------------------------------------------------------------
  **ID**        **Requirements**
  ------------- ---------------------------------------------------------
  FS_PR_001     DoP employees can opt for PLI premium payment via salary
                deduction by submitting policy details on the ESS Portal
                (APT 2.0) for Pay Recovery configuration.

  FS_PR_002     Postal DDOs should verify employee requests and configure
                payroll settings for premium deduction.

  FS_PR_003     IMS should generate a monthly demand file specifying
                premium amounts for employees who opted for Pay Recovery.

  FS_PR_004     Payroll system must deduct premiums from employee
                salaries during payroll run based on IMS demand file.

  FS_PR_005     Employees should receive success/failure notifications
                via ESS Portal, SMS, and/or email.

  FS_PR_006     The module should maintain logs of all transactions,
                failures, and refunds for audit purposes.
  -----------------------------------------------------------------------

## **4. Functional Requirements Specification**

**Flow Chart for Payroll module:**

**IMS Generates Demand File on Defined Date Each Month**

**⬇**

**Data Available for Policy Premium Deduction from Employees' Salaries**

**⬇**

**Payroll Run Executed → Premium Amount Deducted as per Demand File**

**⬇**

**Transaction Details Transmitted via APT 2.0 to IMS**

**⬇**

**APT 2.0 Response Received by IMS**

**⬇**

**IMS Captures and Processes Response**

**⬇**

**Success Case → Premium Details Successfully Updated in IMS**

**⬇**

**Failure Case → IMS Shares Failure Reason with APT 2.0**

**⬇**

**DDO Reviews Failure:**

**• If Update Not Possible → Refund Initiated via APT 2.0\
• If Update Possible → Bulk Upload Method Used for Premium Updation**

**⬇**

**Notifications Triggered to Insurant (Success/Failure)**

**⬇**

**Status Updated on Employee Self-Service Portal**

### 4.1 Dashboard Page

- **Purpose:** Provides a centralized view of all payroll-related
  actions and navigation to sub-pages.

- **Fields:**

  - Navigation Tiles for the Following Pages:

    - Opt-in for Pay Recovery

    - Request Verification

    - Success/Failure

    - Demand File Management

    - Bulk Upload

  - Status Indicators: Pending Verifications, Failed Recoveries

**Dashboard Page Wireframe:**

![](media/image1.png){width="6.261290463692038in"
height="3.510906605424322in"}

### 4.2 Opt-in for Pay Recovery Page

- **Purpose:** Allows employees to opt-in for premium recovery from
  salary.

- **Fields:**

  - Employee ID (Text, Mandatory)

  - Policy Number (Text, Mandatory)

  - Opt-in Checkbox (Boolean)

  - Effective Date (Date Picker)

  - Submit Button

### 4.3 Request Verification Page (Postal DDO Only)

- **Purpose:** Enables DDO to verify employee requests for payroll
  deduction.

- **Fields:**

  - Employee ID

  - Policy Number

  - Request Status (Dropdown: Pending/Approved/Rejected)

  - Verification Date

  - Remarks (Text Area)

  - Approve/Reject Buttons

- **Rules:**

  - Postal DDO should only be able to access Request Verification Page.

  - Postal DDO Should be able to approve the requests.

### 4.4 Success/Failure Page

- **Purpose:** Displays results of payroll recovery transactions.

- **Fields:**

  - **Successful Recovery Table:** All the Successful Recoveries should
    get displayed with only 10 display in first page and then pagination
    option given. This Table should contain the following details:

    - Employee ID

    - Policy Number

    - Deducted Amount

    - Transaction Date

    - Status (Success)

  - **Recovery Failure Table:** All the Recoveries that got failed
    should get displayed with only 10 display in first page and then
    pagination option given. This Table should contain the following
    details:

    - Employee ID

    - Policy Number

    - Deducted Amount

    - Failure Reason

    - Retry Button: Button: Clicking this button should ask for the
      following additional information and upon submit should create
      request for retry of recovery. The additional fields are:

      - Next Retry Date: Calendar Icon: To choose the next date on which
        Pay Recovery will be retried.

      - Submit: Button: To complete the Recovery Retry setup.

### 4.5 Demand file Management Page

- **Purpose:** Generates monthly demand file for payroll deduction.

- **Fields:**

  - Month/Year (Dropdown)

  - Generate File Button

  - Download File Link

  - Status (Generated/Processing)

![](media/image2.png){width="5.291666666666667in"
height="3.527582020997375in"}

### 4.6 Bulk Upload Page

- **Purpose:** Used to Perform Bulk Upload for Pay Recovery or Meghdoot
  File Upload Process.

Note: Details regarding Bulk Upload is given in separate SRS document.

## **5. Attachments**

The following documents can be referred
