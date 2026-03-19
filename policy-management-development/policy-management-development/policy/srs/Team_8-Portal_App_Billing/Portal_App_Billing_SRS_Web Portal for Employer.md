> INTERNAL APPROVAL FORM

**Project Name:** Web Portal for Employer

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

[**2. Scope** [4](#scope)](#scope)

[**3. Business Requirements**
[4](#business-requirements)](#business-requirements)

[**4. Functional Requirements Specification**
[6](#functional-requirements-specification)](#functional-requirements-specification)

[4.1 Login Page [6](#login-page)](#login-page)

[4.2 Employer Registration Page
[6](#employer-registration-page)](#employer-registration-page)

[4.3 Dashboard Page [8](#dashboard-page)](#dashboard-page)

[4.4 Policy List Page [9](#policy-list-page)](#policy-list-page)

[4.5 Add Policy Page [9](#add-policy-page)](#add-policy-page)

[4.6 Bulk Upload Page [10](#bulk-upload-page)](#bulk-upload-page)

[4.7 Payment Processing Page
[12](#payment-processing-page)](#payment-processing-page)

[4.8 Receipt Generation Page
[14](#receipt-generation-page)](#receipt-generation-page)

[4.9 Reports Page [15](#reports-page)](#reports-page)

[4.10 Profile Management Page
[15](#profile-management-page)](#profile-management-page)

[**7. Appendices** [15](#appendices)](#appendices)

## **1. Executive Summary**

The Purpose of this document is to define the requirements for a
web-based portal that enables employers (non-postal Drawing and
Disbursing Officers -- DDOs) to manage Postal Life Insurance (PLI)
policies for their employees. The portal will facilitate premium
collection through salary recovery, bulk policy management, payment
processing, and reporting.

## **2. Scope**

This Portal will allow employers to:

- Register and authenticate securely.

- Upload and manage employee PLI policies (individually or in bulk).

- Upload Premium Payment Information

- Generate receipts and reports.

- Access dashboards for monitoring policy and payment status.

## **3. Business Requirements**

  ------------------------------------------------------------------------
  **Requirement   **Feature**         **Requirement**
  ID**                                
  --------------- ------------------- ------------------------------------
  FS_CP_001       Employer            Employers (non-postal DDOs) must be
                  Registration &      able to register and obtain secure
                  Access              access credentials.

  FS_CP_002       Employer            Registration should capture
                  Registration &      organization details, authorized
                  Access              personnel information, and contact
                                      details.

  FS_CP_003       Secure              Non-Postal DDO user should be able
                  Authentication      to Login using a unique Group Code
                                      and password. Password management
                                      features (reset, lockout after
                                      invalid attempts, history
                                      prevention) should be present.

  FS_CP_004       Policy Management   Employers must be able to add
                                      employee policies individually or
                                      through bulk upload.

  FS_CP_005       Policy Management   Validate policies against PLI
                                      database and enforce rules (e.g.,
                                      policy mode = "Pay", no maturity
                                      within 30 days). Display policy
                                      details in a standardized format.

  FS_CP_006       Premium Collection  Enable employers to process premium
                  via Salary Recovery payments deducted from employee
                                      salaries.

  FS_CP_007       Premium Collection  Ensure salary recovery month matches
                  via Salary Recovery premium month.

  FS_CP_008       Payment Processing  Provide multiple payment options:
                                      Internet Banking, NEFT, RTGS.

  FS_CP_009       Payment Processing  Real-time payment confirmation.

  FS_CP_010       Payment Processing  Validate payment rules (active
                                      policy, no outstanding loans, not
                                      surrendered/ lapsed).

  FS_CP_011       Receipt Generation  Generate receipts with unique
                                      transaction ID, organization
                                      details, policy list, amount
                                      breakdown, and payment timestamp.

  FS_CP_012       Receipt Generation  Allow PDF download and deletion of
                                      wrongly generated receipts.

  FS_CP_013       Dashboard &         Dashboard showing total policies,
                  Reporting           premium due, recent payments,
                                      pending requests, and notifications.

  FS_CP_014       Dashboard &         Export policy list and payment
                  Reporting           history to Excel.
  ------------------------------------------------------------------------

**Flow Diagram for the Employer Portal:**
![](media/image1.png){width="6.268055555555556in"
height="4.178472222222222in"}

## **4. Functional Requirements Specification**

If a non- postal employee chooses the Pay Recovery method for payment of
PLI renewal premiums, the employee must submit the necessary details to
their Drawing and Disbursing Officer (DDO) for premium deduction from
their salary. The DDO will then transfer the deducted amount and update
the policy records.

### 4.1 Login Page

- **Purpose:** Authenticate employer users securely.

- **Fields:**

  - Username: Text: Special Group Code is User Name.

  - Password: Text

  - Login: Button: To Login to Dashboard after Two-Factor Authentication
    Code (OTP via mobile/email).

  - Reset/Forget Password: Link: Option to Reset Password after
    Two-Factor Authentication Code (OTP via mobile/email) should be
    given.

- **Details:**

  - Password reset via registered mobile/email.

  - Maximum 5 invalid attempts before temporary lockout.

  - Password history prevention (last 5 passwords).

### 4.2 Employer Registration Page

- **Purpose:** Allow new employers (non-postal DDOs) to register for
  portal access.

- **Fields:**

  - Organization Name: Full legal name of employer.

  - Group Code: Unique code assigned by India Post.

  - Authorized Person Name: Name of the person managing the portal.

  - Designation: Role of the authorized person.

  - Contact Number: Mobile number for OTP and communication.

  - Email ID: For notifications and password reset.

  - Office Address: Complete postal address.

  - Upload Authorization Letter: PDF or image file validating authority.

- **Details:**

  - Special Group Code should be assigned as username.

  - Password minimum 8 characters with complexity.

  - Mandatory two-factor authentication.

**Flow Chart for Employer Registration:**

![](media/image2.png){width="1.930232939632546in"
height="8.941978346456693in"}

### 4.3 Dashboard Page

- **Purpose:** Provide a summary view of policies and payments.

- **Components (Display Only):**

  - Total Policies: Count of active policies.

  - Premium Due: Amount due for current month.

  - Recent Payments: Last 5 transactions.

  - Pending Requests: Service requests awaiting action.

  - Notifications: System alerts and updates.

  - Provide Navigation tabs for moving to following pages:

    - Policy List Page

    - Add Policy Page

    - Bulk Upload Page

    - Payment Processing Page

    - Receipt Page

    - Reports Page

    - Profile Management Page

- **Details:**

  - Interactive charts for premium trends.

  - Quick links to policy list and payment module.

**Wireframe for the Dashboard Page:**

![](media/image3.png){width="6.268055555555556in"
height="4.178472222222222in"}

### 4.4 Policy List Page

- **Purpose:** Display all managed policies with search and filter
  options.

- **Fields (Display Only Table):**

  - Policy Number

  - Insurant Name

  - Maturity Date

  - Last Premium Date

  - Premium Paid To Date

  - Premium Amount

  - Policy Mode (Pay/Cash)

- **Details:**

  - Filters: Active/Inactive, Premium Due, Policy Mode.

  - Export to Excel option.

  - Option to Add new policies individually or in bulk

  - Option to Remove policies from active list

  - Option to Modify policy details

  - Option to Export policy list to Excel

- **Policy Display Format**\
  The system shall display policy details in the following format:

  -------------------------------------------------------------------------------------
  Policy Number  Insurant   Maturity     Last Premium Premium Paid Premium   Policy
                 Name       Date         Date         To Date                Mode
  -------------- ---------- ------------ ------------ ------------ --------- ----------
  PLI123456789   Employee   DD-MM-YYYY   DD-MM-YYYY   DD-MM-YYYY   Amount    Pay/Cash
                 Name                                                        

  -------------------------------------------------------------------------------------

### 4.5 Add Policy Page

- **Purpose:** Add a new policy manually.

- **Fields:**

  - Policy Number: Must match PLI database.

  - Insurant Name: Employee name.

  - Date of Birth: For validation.

  - Policy Start Date: Original start date.

  - Premium Amount: Monthly premium.

  - Policy Mode: Must be "Pay".

- **Details:**

  - Validate policy against PLI database.

  - Check rules like no maturity within 30 days, no duplicates.

### 4.6 Bulk Upload Page

- **Purpose:** Upload multiple policies via Excel/CSV.

- **Fields:**

  - Upload File: CSV/Excel containing policy details.

  - Template Download: Link to correct format.

- **Details:**

  - Validate file format and data integrity.

  - Show error report for invalid entries.

- **Policy Validation Rules:**

  - Policy must exist in PLI database.

  - Policy mode must be \"Pay\" (not \"Cash\").

  - Premium payment not allowed if maturity within 30 days.

  - Salary recovery month must match premium month.

  - No duplicate policy entries.

**Flow Chart for Bulk Upload of Policies by
Employer:**![](media/image4.png){width="4.043764216972878in"
height="9.290697725284339in"}

### 4.7 Payment Processing Page

- **Purpose:** Process premium payment for selected policies.

- **Fields:**

  - Select Policies: Checkbox list of policies due.

  - Total Premium Amount: Auto-calculated.

  - Payment Method: Internet Banking / NEFT / RTGS.

  - Transaction Reference: Auto-generated after payment.

- **Details:**

  - Real-time payment confirmation.

  - Validate payment rules: active policy, no outstanding loans, not
    surrendered/lapsed.

- **Payment Validation Rules:**

  - Premium must be paid up to current month

  - Policy must be in active status

  - No maturity within 30 days

  - No outstanding loans affecting premium

  - Policy not surrendered or lapsed

- **Payment Gateway Options:**

  - Internet Banking (All public and private sector banks)

  - NEFT Transfer

  - RTGS Transfer

  - Real time Payment confirmation.

- **Receipt Format:**

  - Transaction Type: 'Premium Payment'

  - Unique 16-digit transaction ID

  - Organization details

  - List of all policies paid

  - Amount breakdown

  - Payment date and time

  - Downloadable PDF format

  - Option to delete the wrongly generated receipt

**Flow Chart for Payment Processing by Employer:**
![](media/image5.png){width="3.4534886264216973in"
height="9.266341863517061in"}

### 4.8 Receipt Generation Page

- **Purpose:** Generate and download payment receipts.

- **Fields:**

  - Transaction ID: Unique 16-digit number.

  - Organization Details: Employer name and Group Code.

  - Policy List: Policies included in payment.

  - Amount Breakdown: Premium per policy.

  - Payment Date & Time: Timestamp.

  - Download PDF: Button.

  - Delete Receipt: Option for correction.

- **Details:**

  - PDF Download Options.

**Sample Receipt:**

![](media/image6.png){width="3.7469728783902014in"
height="5.6204604111986in"}

### 4.9 Reports Page

- **Purpose:** Generate and export reports.

- **Fields:**

  - Report Type: Policy List / Payment History.

  - Date Range: Start and end date.

  - Export Format: Excel / PDF.

- **Details:**

  - Include filters for policy status and payment status.

### 4.10 Profile Management Page

- **Purpose:** Allow new employers (non-postal DDOs) to register for
  portal access.

- **Fields:**

  - Organization Details: Editable fields.

  - Contact Information: Mobile and email.

  - Change Password: Old password, new password, confirm password.

  - Two-Factor Settings: Enable/disable OTP.

## **7. Appendices**

The Following Documents attached below can be used.
