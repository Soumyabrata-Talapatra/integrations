> INTERNAL APPROVAL FORM

**Project Name:** Bulk Upload (Pay Recovery and Meghdoot)

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

[4.1 Billing & Collection through Pay Recovery
[5](#billing-collection-through-pay-recovery)](#billing-collection-through-pay-recovery)

[4.2 Billing & Collection through Meghdoot Upload (for Individual
Policy- where premium mode is cash)
[8](#billing-collection-through-meghdoot-upload-for-individual-policy--where-premium-mode-is-cash)](#billing-collection-through-meghdoot-upload-for-individual-policy--where-premium-mode-is-cash)

[**5. Attachments** [9](#attachments)](#attachments)

## **1. Executive Summary**

The Bulk Upload feature in the Payroll Module of IMS enables uploading
of policy-related data for premium recovery in cases where automated
updates via APT 2.0 fail. It supports two sub-pages:

- Pay Recovery Bulk Upload (for APS policies and Non-Postal DDO cases)

- Meghdoot Upload (for individual policies with cash payment mode)

## **2. Project Scope**

This Feature ensures:

- Recovery of failed records for APS and Non-Postal DDO policies.

- Upload of CSV files for policy updates after manual collection
  verification.

- OTP-based confirmation for Meghdoot uploads.

## **3. Business Requirements**

  ------------------------------------------------------------------------
  **ID**       **Feature**   **Requirements**
  ------------ ------------- ---------------------------------------------
  FS_BU_001    Dashboard     Enable Bulk Upload Dashboard to provide
                             navigation to Pay Recovery Bulk Upload and
                             Meghdoot Upload pages.

  FS_BU_002    Pay Recovery  Allow upload of CSV files for APS and
               Bulk Upload   Non-Postal DDO policies after collection
                             verification in APT 2.0.

  FS_BU_003    Pay Recovery  Validate that collection amount exists under
               Bulk Upload   Special Group/DDO Code/Office Code before
                             upload.

  FS_BU_004    Pay Recovery  Reconcile suspense amount in APT 2.0 against
               Bulk Upload   total premiums in the uploaded file.

  FS_BU_005    Pay Recovery  Display validation errors for incorrect or
               Bulk Upload   missing fields in the CSV file.

  FS_BU_006    Pay Recovery  Log all upload attempts and validation
               Bulk Upload   results for audit purposes.

  FS_BU_007    Meghdoot      Provide separate login for Divisional Head
               Upload        for Meghdoot uploads.

  FS_BU_008    Meghdoot      Enable OTP-based confirmation before final
               Upload        submission of Meghdoot upload.

  FS_BU_009    Meghdoot      Validate CSV file against prescribed template
               Upload        before processing.

  FS_BU_010    Meghdoot      Update IMS accurately with uploaded data
               Upload        after successful validation.

  FS_BU_011    Meghdoot      Generate success/failure report for
               Upload        Divisional Head post-upload.

  FS_BU_012    General       Log all actions (uploads, validations,
                             approvals) for audit and compliance purposes.
  ------------------------------------------------------------------------

## **4. Functional Requirements Specification**

### 4.1 Billing & Collection through Pay Recovery

- **Purpose:** In the case of failed records through APT 2.0, recovery
  of premium payments pertaining to APS policies or Non-Postal DDO
  cases---where the DDO/HO has already collected the amount against
  Special Group/DDO Code/Office Code in APT 2.0 but the policy details
  not updated in IMS---should be processed accordingly to enable
  uploading of failed records for APS and Non-Postal DDO policies.

- **Flow Diagram for the Pay Recovery Bulk Upload Process:**

![](media/image1.png){width="6.447222222222222in"
height="3.547222222222222in"}

\* "*Before initiating the collection process, the Counter PA must
verify whether the collection amount for the particular month is already
available under the respective Special Group/DDO Code/Office Code in APT
2.0. If the amount is not available, only then the collection may be
carried out"*

- **Data Upload Process:** Once the collection process for the Special
  Group/DDO Code/Office Code of APT 2.0 is completed, the user may
  initiate the data upload of policy wise against the respective Special
  Group/DDO Code/Office Code of APT 2.0, as illustrated in the flow
  chart below:

![](media/image2.png){width="6.4006944444444445in"
height="3.527083333333333in"}

- **CSV File Template:** After confirmatory screen user may be able to
  upload the CSV file in the prescribed template with below field:

+------------------+--------------------+---------------+-------------+------------------------+
| Field Name       | Mandatory/Optional | Acceptable    | Example     | Error messages         |
|                  |                    | formats       |             |                        |
+==================+====================+===============+=============+========================+
| CIRCLE_CODE      | Mandatory          | alphabets     |  UP, KT     |                        |
+------------------+--------------------+---------------+-------------+------------------------+
| PAO_CODE (DDO    | Conditional        | alpha-numeric |             |                        |
| Code/Office      | Mandatory          |               |             |                        |
| Code)            |                    |               |             |                        |
|                  | Mandatory if the   |               |             |                        |
|                  | file belongs to    |               |             |                        |
|                  | special group      |               |             |                        |
|                  | policies           |               |             |                        |
+------------------+--------------------+---------------+-------------+------------------------+
| POLICY_NO        | Mandatory          | alpha-numeric |             | Policy number cannot   |
|                  |                    |               |             | be blank               |
+------------------+--------------------+---------------+-------------+------------------------+
| NAME             | Optional           | Only          |             |                        |
|                  |                    | alphabets     |             |                        |
+------------------+--------------------+---------------+-------------+------------------------+
| SUM_ASSD         | Optional           | Numeric       |             |                        |
+------------------+--------------------+---------------+-------------+------------------------+
| PREM_AMNT        | Mandatory          | Numeric       |             | Premium amount can't   |
|                  |                    |               |             | be blank               |
+------------------+--------------------+---------------+-------------+------------------------+
| DATE_MATURITY    | Optional           | Numeric       |             |                        |
+------------------+--------------------+---------------+-------------+------------------------+
| DATE_ENTRY       | Optional           | Numeric       |             |                        |
| (Policy Issue    |                    |               |             |                        |
| Date )           |                    |               |             |                        |
+------------------+--------------------+---------------+-------------+------------------------+
| FROM_DATE        | Mandatory          | Numeric       | dd-mm-yyyy\ | From date can't be     |
|                  |                    |               | dd-mmm-yyyy | blank/                 |
|                  |                    |               |             |                        |
|                  |                    |               |             | Invalid date           |
+------------------+--------------------+---------------+-------------+------------------------+
| TO_DATE          | Mandatory          | Numeric       | dd-mm-yyyy\ | TO date can't be       |
|                  |                    |               | dd-mmm-yyyy | blank/                 |
|                  |                    |               |             |                        |
|                  |                    |               |             | Invalid date           |
+------------------+--------------------+---------------+-------------+------------------------+
| TOTAL_RECEIPT    | Mandatory          | Numeric       |             | Total receipt can't be |
| (Premium +       |                    |               |             | blank                  |
| Service Tax/GST) |                    |               |             |                        |
+------------------+--------------------+---------------+-------------+------------------------+
| DATE_TRANSACTION | Mandatory          | Numeric       | dd-mm-yyyy\ | Date_transaction can't |
|                  |                    |               | dd-mmm-yyyy | be blank/              |
|                  |                    |               |             |                        |
|                  |                    |               |             | Invalid date           |
|                  |                    |               |             |                        |
|                  |                    |               |             | Transaction date can't |
|                  |                    |               |             | future date            |
+------------------+--------------------+---------------+-------------+------------------------+
| TRANSACTION_TYPE | Mandatory          | alphabet      |  C/P        | Transaction Type       |
| (Cash/Pay)       |                    |               |             | cannot be blank        |
+------------------+--------------------+---------------+-------------+------------------------+
| REBATE_AMT       | Optional           | Numeric       |             |                        |
+------------------+--------------------+---------------+-------------+------------------------+
| Tax_Type         | Conditional        | Conditional   | 1 for first | 1.If FY tax/Renewal    |
|                  | Mandatory          | Mandatory     | year        | tax is updated and Tax |
|                  |                    |               | tax/GST     | Type is blank, then    |
|                  |                    | (Required if  |             | display error          |
|                  |                    | first year    | 2 for       | message - Tax Type is  |
|                  |                    | tax,          | renewal     | required               |
|                  |                    | GST/renewal   | year        |                        |
|                  |                    | year tax, GST | tax/GST     | 2\. If data type is    |
|                  |                    | is present)   |             | other than numeric --  |
|                  |                    |               |             | Only numeric value     |
|                  |                    |               |             | accepted               |
+------------------+--------------------+---------------+-------------+------------------------+
| First_Year_Tax   | Mandatory          | Numeric       | Can be null | 1\. If transaction     |
|                  |                    |               | or numeric  | date\>= 1 Jan 2015 and |
|                  |                    |               |             | either FY tax or       |
|                  |                    |               |             | Renewal tax is not     |
|                  |                    |               |             | updated, then error    |
|                  |                    |               |             | message to be          |
|                  |                    |               |             | displayed- Service tax |
|                  |                    |               |             | is required            |
|                  |                    |               |             |                        |
|                  |                    |               |             | 2\. If data type is    |
|                  |                    |               |             | other than numeric --  |
|                  |                    |               |             | Only numeric value     |
|                  |                    |               |             | accepted               |
+------------------+--------------------+---------------+-------------+------------------------+
| Renewal_Year_Tax | Mandatory          | Numeric       | Can be null | 1\. If transaction     |
|                  |                    |               | or numeric  | date\>= 1 Jan 2015 and |
|                  |                    |               |             | either FY tax or       |
|                  |                    |               |             | Renewal tax is not     |
|                  |                    |               |             | updated, then error    |
|                  |                    |               |             | message to be          |
|                  |                    |               |             | displayed- Service tax |
|                  |                    |               |             | is required            |
|                  |                    |               |             |                        |
|                  |                    |               |             | 2\. If data type is    |
|                  |                    |               |             | other than numeric --  |
|                  |                    |               |             | Only numeric value     |
|                  |                    |               |             | accepted               |
+------------------+--------------------+---------------+-------------+------------------------+
| First Year CGST  | Mandatory          | Numeric       | Can be null | 1\. If transaction     |
|                  |                    |               | or numeric  | date\>= 1 July 2017    |
|                  |                    |               |             | and either FY GST or   |
|                  |                    |               |             | Renewal GST is not     |
|                  |                    |               |             | updated, then error    |
|                  |                    |               |             | message to be          |
|                  |                    |               |             | displayed- GST is      |
|                  |                    |               |             | required               |
|                  |                    |               |             |                        |
|                  |                    |               |             | 2\. If data type is    |
|                  |                    |               |             | other than numeric --  |
|                  |                    |               |             | Only numeric value     |
|                  |                    |               |             | accepted               |
+------------------+--------------------+---------------+-------------+------------------------+
| First Year       | Mandatory          | Numeric       | Can be null | 1\. If transaction     |
| SGST/UTGST       |                    |               | or numeric  | date\>= 1 July 2017    |
|                  |                    |               |             | and either FY GST or   |
|                  |                    |               |             | Renewal GST is not     |
|                  |                    |               |             | updated, then error    |
|                  |                    |               |             | message to be          |
|                  |                    |               |             | displayed- GST is      |
|                  |                    |               |             | required               |
|                  |                    |               |             |                        |
|                  |                    |               |             | 2\. If data type is    |
|                  |                    |               |             | other than numeric --  |
|                  |                    |               |             | Only numeric value     |
|                  |                    |               |             | accepted               |
+------------------+--------------------+---------------+-------------+------------------------+
| Renewal Year     | Mandatory          | Numeric       | Can be null | 1\. If transaction     |
| CGST             |                    |               | or numeric  | date\>= 1 July 2017    |
|                  |                    |               |             | and either FY GST or   |
|                  |                    |               |             | Renewal GST is not     |
|                  |                    |               |             | updated, then error    |
|                  |                    |               |             | message to be          |
|                  |                    |               |             | displayed- GST is      |
|                  |                    |               |             | required               |
|                  |                    |               |             |                        |
|                  |                    |               |             | 2\. If data type is    |
|                  |                    |               |             | other than numeric --  |
|                  |                    |               |             | Only numeric value     |
|                  |                    |               |             | accepted               |
+------------------+--------------------+---------------+-------------+------------------------+
| Renewal Year     | Mandatory          | Numeric       | Can be null | 1\. If transaction     |
| SGST/UTGST       |                    |               | or numeric  | date\>= 1 July 2017    |
|                  |                    |               |             | and either FY GST or   |
|                  |                    |               |             | Renewal GST is not     |
|                  |                    |               |             | updated, then error    |
|                  |                    |               |             | message to be          |
|                  |                    |               |             | displayed- GST is      |
|                  |                    |               |             | required               |
|                  |                    |               |             |                        |
|                  |                    |               |             | 2\. If data type is    |
|                  |                    |               |             | other than numeric --  |
|                  |                    |               |             | Only numeric value     |
|                  |                    |               |             | accepted               |
+------------------+--------------------+---------------+-------------+------------------------+
| RECEIPT_NUMBER   | Conditional        | Conditional   | Can be null | 1\. If receipt no is   |
|                  | Mandatory          | Mandatory     | or numeric  | not updated for pay    |
|                  |                    |               |             | recovery policy, then  |
|                  |                    | (Required if  |             | error message to be    |
|                  |                    | Pay Recovery  |             | displayed- Receipt no. |
|                  |                    | Policy)       |             | is required against    |
|                  |                    |               |             | DDO Code/Office Code.  |
|                  |                    |               |             |                        |
|                  |                    |               |             | 2\. If data type is    |
|                  |                    |               |             | other than numeric --  |
|                  |                    |               |             | Only numeric value     |
|                  |                    |               |             | accepted               |
+------------------+--------------------+---------------+-------------+------------------------+

- **CSV File Validation:** For pay recovery policies, the system will
  reconcile the total amount available under the DDO Code/Office Code in
  suspense against the total premiums in the file. If the suspense
  amount is less than the total premiums, a validation message will be
  displayed stating that the suspense amount for the DDO Code/Office
  Code is insufficient to cover the total premiums in the file.

### 4.2 Billing & Collection through Meghdoot Upload (for Individual Policy- where premium mode is cash)

- **Purpose:** For policies belonging to individual customers with cash
  payment mode, where premiums for earlier months were not updated in
  IMS, a CSV file in the prescribed format will be prepared by the
  concerned CPC and verified by the Divisional Head. A separate login
  window will be provided to the Divisional Head in IMS to facilitate
  this activity. After thorough verification, the Divisional Head may
  upload the CSV file, and an OTP-based confirmation will be required
  prior to submission. Once submitted, the file will be processed in IMS
  and the output will be made available to the Divisional Head.

- **Flow Diagram for the Meghdoot Bulk Upload Process for Individual
  Policies where payment mode is cash:**

![](media/image3.png){width="6.200694444444444in"
height="3.6805555555555554in"}

- **CSV File Template:** After confirmatory screen user may be able to
  upload the CSV file in the prescribed template (the same which was
  used for Bulk upload, format given in above section 4.1) for policy
  updation in IMS.

## **5. Attachments**

The following documents can be referred.
