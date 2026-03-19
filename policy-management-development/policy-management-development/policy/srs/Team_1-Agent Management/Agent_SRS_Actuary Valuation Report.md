> INTERNAL APPROVAL FORM

**Project Name:** Actuary Valuation Report

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

[3.1 Actuary Valuation Reports
[4](#actuary-valuation-reports)](#actuary-valuation-reports)

[**4. Reports Format** [4](#reports-format)](#reports-format)

[4.1 Report 1: Policy Data Format
[4](#report-1-policy-data-format)](#report-1-policy-data-format)

[4.2 Report 2: Exit By Report
[7](#report-2-exit-by-report)](#report-2-exit-by-report)

[**5. Functional Requirements Specification**
[8](#functional-requirements-specification)](#functional-requirements-specification)

[5.1 Generate Actuarial Report Page
[8](#generate-actuarial-report-page)](#generate-actuarial-report-page)

## **1. Executive Summary**

The purpose of this module is to generate actuarial valuation reports
for insurance policies as per regulatory and internal actuarial
requirements. Two reports are required:

- Policy Data Format Report

- Exit By Report

## **2. Project Scope**

The reports will be generated from IMS policy database and will include
all active and exited policies as of the valuation date. These reports
will be used by the actuarial team for liability calculation, bonus
allocation, and risk assessment.

## **3. Business Requirements**

### 3.1 Actuary Valuation Reports

  -----------------------------------------------------------------------
  Requirement ID Requirement
  -------------- --------------------------------------------------------
  FS_AV_001      Generate a Policy Data Format Report including all
                 active policies as of the valuation date, with accrued
                 bonus split into Current and Vested Bonus, and details
                 on premium payments and policyholder demographics.

  FS_AV_002      Generate an Exit By Report capturing all policies exited
                 during the valuation period (death, surrender,
                 maturity), including claim details, cause of exit, and
                 payment status.

  FS_AV_003      Reports should be exportable in Excel (.xlsx) and CSV
                 formats and Column headers must match actuarial
                 specifications.

  FS_AV_004      The report should be for the period of last 1 year and
                 should be live and synchronized with transactions
                 processed in real time.

  FS_AV_005      The system shall provide a summary screen showing the
                 number of records extracted.

  FS_AV_006      The system should display the last downloaded report
                 information for both the report types and use should be
                 able to re-download that report.
  -----------------------------------------------------------------------

## **4. Reports Format**

### 4.1 Report 1: Policy Data Format

**Description:** Extract policy-level details for valuation.

**Fields & Specifications:**

+-------+--------------------------------+--------------+------------------------------------------+----------------+
| **Sr. | **Field Name**                 | **Field      | **Description**                          | **Comment**    |
| No.** |                                | Type**       |                                          |                |
+:=====:+================================+==============+==========================================+================+
| 1     | POLICY_NO                      | Alpha        | Policy number                            |                |
|       |                                | Numeric      |                                          |                |
+-------+--------------------------------+--------------+------------------------------------------+----------------+
| 2     | INSURED_NAME                   | Text         | Name of the primary life assured         |                |
+-------+--------------------------------+--------------+------------------------------------------+----------------+
| 3     | POLICY_STATUS                  | Text         | Policy status as on date of valuation    |                |
+-------+--------------------------------+--------------+------------------------------------------+----------------+
| 4     | TYPE_OF_POLICIES               | Text         | Product/Plan code                        |                |
+-------+--------------------------------+--------------+------------------------------------------+----------------+
| 5     | DATE_OF_BIRTH                  | Date         | Date of birth of primary life assured    |                |
|       |                                | (DD-MM-YYYY) |                                          |                |
+-------+--------------------------------+--------------+------------------------------------------+----------------+
| 6     | DATE_OF_ACCEPTANCE             | Date         | Date of commencement of policy           |                |
|       |                                | (DD-MM-YYYY) |                                          |                |
+-------+--------------------------------+--------------+------------------------------------------+----------------+
| 7     | DATE_OF_MATURITY               | Date         | Date of maturity of the policy           |                |
|       |                                | (DD-MM-YYYY) |                                          |                |
+-------+--------------------------------+--------------+------------------------------------------+----------------+
| 8     | MATURITY_AGE                   | Integer      | Age last birthday of primary life        |                |
|       |                                |              | assured as on date of maturity           |                |
+-------+--------------------------------+--------------+------------------------------------------+----------------+
| 9     | PREMIUM_AMOUNT                 | Numeric      | Instalment Premium payable as on premium |                |
|       |                                |              | due date                                 |                |
+-------+--------------------------------+--------------+------------------------------------------+----------------+
| 10    | ANNUALISED_PREMIUM             | Numeric      | Annual Premium (This should be equal to  |                |
|       |                                |              | PREMIUM_AMOUNT\*MODE_OF_PREMIUM_PAYMENT) |                |
+-------+--------------------------------+--------------+------------------------------------------+----------------+
| 11    | MODE_OF_PREMIUM_PAYMENT        | Text         | Premium payment frequency                |                |
+-------+--------------------------------+--------------+------------------------------------------+----------------+
| 12    | CATEGORY                       | Text         | Medical Indicator                        |                |
+-------+--------------------------------+--------------+------------------------------------------+----------------+
| 13    | SUM_ASSURED                    | Numeric      | Original sum assured                     |                |
+-------+--------------------------------+--------------+------------------------------------------+----------------+
| 14    | ACCRUED_BONUS                  | Numeric      | Bonus attached to policy since inception | We understand  |
|       |                                |              | of the policy as on date of valuation.   | that the       |
|       |                                |              |                                          | \"Accrued      |
|       |                                |              |                                          | Bonus\" column |
|       |                                |              |                                          | in the         |
|       |                                |              |                                          | database       |
|       |                                |              |                                          | includes the   |
|       |                                |              |                                          | current        |
|       |                                |              |                                          | year\'s bonus. |
|       |                                |              |                                          | Please         |
|       |                                |              |                                          | confirm. Given |
|       |                                |              |                                          | our            |
|       |                                |              |                                          | understandig   |
|       |                                |              |                                          | is correct,    |
|       |                                |              |                                          | Kindly split   |
|       |                                |              |                                          | \"Accrued      |
|       |                                |              |                                          | Bonus\" column |
|       |                                |              |                                          | into two       |
|       |                                |              |                                          | separate       |
|       |                                |              |                                          | columns (as    |
|       |                                |              |                                          | under)         |
+-------+--------------------------------+--------------+------------------------------------------+----------------+
|       | a.  CURRENT_BONUS              | Numeric      | For the valuation of March 2025, the     |                |
|       |                                |              | current bonus represents the bonus for   |                |
|       |                                |              | the period March 2024 to March 2025      |                |
+-------+--------------------------------+--------------+------------------------------------------+----------------+
|       | b.  VESTED_BONUS               | Numeric      | For the valuation of March 2025, the     |                |
|       |                                |              | vested bonus represents the bonus uptil  |                |
|       |                                |              | March 2024                               |                |
+-------+--------------------------------+--------------+------------------------------------------+----------------+
| 15    | OFFICECODE                     | Alpha        | Branch/Office Code                       |                |
|       |                                | Numeric      |                                          |                |
+-------+--------------------------------+--------------+------------------------------------------+----------------+
| 16    | VYOB                           | Numeric      | valuation year of birth                  |                |
+-------+--------------------------------+--------------+------------------------------------------+----------------+
| 17    | PAID_TO_DATE                   | Date         | Date till which premium has been paid by | As per our     |
|       |                                | (DD-MM-YYYY) | the policyholder                         | understanding, |
|       |                                |              |                                          | If policy is   |
|       |                                |              |                                          | issued on 28   |
|       |                                |              |                                          | Feb 2022.\     |
|       |                                |              |                                          | First premium  |
|       |                                |              |                                          | will be        |
|       |                                |              |                                          | received on 28 |
|       |                                |              |                                          | Feb 2022 and   |
|       |                                |              |                                          | 2nd premium    |
|       |                                |              |                                          | will be        |
|       |                                |              |                                          | received on    |
|       |                                |              |                                          | 1st March      |
|       |                                |              |                                          | 2022.\         |
|       |                                |              |                                          | The            |
|       |                                |              |                                          | PAID_TO_DATE   |
|       |                                |              |                                          | for this       |
|       |                                |              |                                          | policy will be |
|       |                                |              |                                          | 31st March     |
|       |                                |              |                                          | 2022 after     |
|       |                                |              |                                          | receiving 2nd  |
|       |                                |              |                                          | premium.\      |
|       |                                |              |                                          | **Request you  |
|       |                                |              |                                          | to confirm if  |
|       |                                |              |                                          | our            |
|       |                                |              |                                          | understanding  |
|       |                                |              |                                          | is correct**   |
+-------+--------------------------------+--------------+------------------------------------------+----------------+
| 18    | DATE_OF_LAST_PREMIUMS          | Date         |                                          | What does this |
|       |                                | (DD-MM-YYYY) |                                          | field          |
|       |                                |              |                                          | currently      |
|       |                                |              |                                          | contains       |
+-------+--------------------------------+--------------+------------------------------------------+----------------+
| 19    | TOTAL_AMOUNT_OF_UNPAID_PREMIUM | Numeric      |                                          | What does this |
|       |                                |              |                                          | field          |
|       |                                |              |                                          | currently      |
|       |                                |              |                                          | contains       |
+-------+--------------------------------+--------------+------------------------------------------+----------------+
| 20    | DATE_OF_SANCTION               | Date         |                                          | What does this |
|       |                                | (DD-MM-YYYY) |                                          | field          |
|       |                                |              |                                          | currently      |
|       |                                |              |                                          | contains       |
+-------+--------------------------------+--------------+------------------------------------------+----------------+
| 21    | AMOUNT_PAID_UP                 | Numeric      |                                          | What does this |
|       |                                |              |                                          | field          |
|       |                                |              |                                          | currently      |
|       |                                |              |                                          | contains       |
+-------+--------------------------------+--------------+------------------------------------------+----------------+
| 22    | LA_GENDER                      | Text         | Gender of primary life assured (M or F)  | It should take |
|       |                                |              |                                          | values M or F  |
+-------+--------------------------------+--------------+------------------------------------------+----------------+
| 23    | PPT                            | Integer      | Premium Paying Term of the contract      |                |
+-------+--------------------------------+--------------+------------------------------------------+----------------+
| 24    | PT                             | Integer      | Policy Term of the contract              |                |
+-------+--------------------------------+--------------+------------------------------------------+----------------+
| 25    | SECOND_LIFE_DOB                | Date         | Date of birth of\                        |                |
|       |                                | (DD-MM-YYYY) | - second life in case of Joint Life      |                |
|       |                                |              | Assurance\                               |                |
|       |                                |              | - Proposer in case of Children Policy\   |                |
|       |                                |              | - For others this field should be        |                |
|       |                                |              | **01-01-1900**                           |                |
+-------+--------------------------------+--------------+------------------------------------------+----------------+
| 26    | SECOND_LIFE_GENDER             | Text         | Gender of\                               | It should take |
|       |                                |              | - second life in case of Joint Life      | values M, F or |
|       |                                |              | Assurance\                               | Blank          |
|       |                                |              | - Proposer in case of Children Policy\   |                |
|       |                                |              | - For others this field should be        |                |
|       |                                |              | **\'BLANK\'**                            |                |
+-------+--------------------------------+--------------+------------------------------------------+----------------+
| 27    | PREMIUM_CEASING_AGE            | Numeric      | Age of the policyholder at which the     |                |
|       |                                |              | policyholder will stop paying premium    |                |
+-------+--------------------------------+--------------+------------------------------------------+----------------+
| 28    | AGE_AT_ENTRY                   | Integer      | Age last birthday of primary life        |                |
|       |                                |              | assured as on date of acceptance         |                |
+-------+--------------------------------+--------------+------------------------------------------+----------------+
| 29    | SECOND_LIFE_AGE_AT_ENTRY       | Integer      | Age last birthday of secondary life      |                |
|       |                                |              | assured (proposer in case of child       |                |
|       |                                |              | policy) as on date of acceptance         |                |
+-------+--------------------------------+--------------+------------------------------------------+----------------+

### 4.2 Report 2: Exit By Report

**Description:** Extract details of policies exited during valuation
period.

**Fields & Specifications:**

  -------------------------------------------------------------------------------------------
    **Sr. **Field Name**            **Field Type** **Description**
    No.**                                          
  ------- ------------------------- -------------- ------------------------------------------
        1 POLICY_NO                 Alpha Numeric  Policy number

        2 INSURED_NAME              Text           Name of the primary life assured

        3 POLICY_STATUS             Text           Policy status as on date of valuation

        4 TYPE_OF_POLICIES          Text           Product/Plan code

        5 DATE_OF_BIRTH             Date           Date of birth of primary life assured
                                    (DD-MM-YYYY)   

        6 DATE_OF_ACCEPTANCE        Date           Date of commencement of policy
                                    (DD-MM-YYYY)   

        7 DATE_OF_MATURITY          Date           Date of maturity of the policy
                                    (DD-MM-YYYY)   

        8 MATURITY_AGE              Integer        Age last birthday of primary life assured
                                                   as on date of maturity

        9 PREMIUM_AMOUNT            Numeric        Instalment Premium payable as on premium
                                                   due date

       10 ANNUALISED_PREMIUM        Numeric        Annual Premium (This should be equal to
                                                   PREMIUM_AMOUNT\*MODE_OF_PREMIUM_PAYMENT)

       11 MODE_OF_PREMIUM_PAYMENT   Text           Premium payment frequency

       12 Category                  Text           Medical Indicator

       13 SUM_ASSURED               Numeric        Original sum assured

       14 ACCRUED_BONUS             Numeric        Bonus attached to policy since inception
                                                   of the policy as on date of valuation.

       15 DATE_OF_CESSATION         Date            
                                    (DD-MM-YYYY)   

       16 CAUSE_OF_EXIT             Text            

       17 DATE_OF_DEATH             Date            
                                    (DD-MM-YYYY)   

       18 AMOUNT                    Numeric         

       19 DATE_OF_APPLICATION       Date            
                                    (DD-MM-YYYY)   

       20 DATE_OF_SANCTION          Date            
                                    (DD-MM-YYYY)   

       21 OFFICECODE                Alpha Numeric  Branch/Office Code

       22 VYOB                      Numeric        valuation year of birth

       23 PAID_TO_DATE              Date           Date till which premium has been paid by
                                    (DD-MM-YYYY)   the policyholder

       24 DATE_OF_CLAIM INTIMATION  Date           Date when the claim was first intimated in
                                    (DD-MM-YYYY)   Postal Life Insurance

       25 DATE_OF_INCIDENT          Date           Date when the event has reported to be
                                    (DD-MM-YYYY)   occurred

       26 CLAIM AMOUNT PAID         Numeric        Actual claim amount paid

       27 CLAIM STATUS                             One of the following : - Open, Closed with
                                                   payment, Rejected (i.e. closed without
                                                   payment)

       28 DATE_OF_PAYMENT           Date           Date on which the death/ surrender/
                                    (DD-MM-YYYY)   maturity payment was made
  -------------------------------------------------------------------------------------------

## **5. Functional Requirements Specification**

### 5.1 Generate Actuarial Report Page

> 5.1.1 Generate Report Page

- **Fields:**

  - **Report Type** (Dropdown)

    - Options: Policy Data Report, Exit Data Report

    - Mandatory: Yes

    - Validation: Must select one option.

  - **From Date** (Date Picker)

    - Format: DD-MM-YYYY

    - Mandatory: Yes

  - **To Date** (Date Picker)

    - Format: DD-MM-YYYY

    - Validation: Must be ≥ From Date

    - Mandatory: Yes

  - **Export Format** (dropdown)

    - Options: Excel (.xlsx), CSV

    - Mandatory: Yes

  - **Generate Report** (button)

    - Enabled only when all mandatory fields are valid

  - **Last Policy Data Report download** (link)

    - Display Timestamp when it was generated.

  - **Last Exit By Report download** (link)

    - Display Timestamp when it was generated.

- **Business Rules:**

  - For Policy Data Report, the system will use only the To Date as the
    valuation date.

  - For Exit By Report, both From Date and To Date define the valuation
    period.

  - Role-Based Access: **Restricted to Actuarial and Admin users.**

![](media/image1.png){width="3.776434820647419in"
height="5.66465113735783in"}
