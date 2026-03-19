> INTERNAL APPROVAL FORM

**Project Name:** Product Setup for Existing products (Rules, business
logic, Calculation, Product Tables, validations)

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

[**3. Product Categories**
[4](#product-categories)](#product-categories)

[3.1 PLI Products [5](#pli-products)](#pli-products)

[3.2 RPLI Products [5](#rpli-products)](#rpli-products)

[**4. Functional Requirements Specification**
[6](#functional-requirements-specification)](#functional-requirements-specification)

[4.1 Product Configuration
[6](#product-configuration)](#product-configuration)

[4.2 Bonus Configuration
[6](#bonus-configuration)](#bonus-configuration)

[4.3 Premium Calculation
[6](#premium-calculation)](#premium-calculation)

[4.4 Maturity & Survival Benefits
[6](#maturity-survival-benefits)](#maturity-survival-benefits)

[4.5 Loan & Surrender Rules
[6](#loan-surrender-rules)](#loan-surrender-rules)

[4.6 Policy Conversion [6](#policy-conversion)](#policy-conversion)

[**5. Wireframe** [6](#wireframe)](#wireframe)

[**6. Appendices** [6](#appendices)](#appendices)

## **1. Executive Summary**

This document describes the functional and nonfunctional requirements
for the Product Setup component of the PLI / RPLI (Postal Life Insurance
/ Rural Postal Life Insurance). The Product Setup module is responsible
for defining insurance product variants, premium rates, bonus rates,
term structures, formulae, tables, and related system behaviour to
support quoting, issuance, premium collection, maturity, surrender, etc.

## **2. Project Scope**

This module will support:

- Definition and management of various PLI / RPLI product types (e.g.
  Endowment Assurance, Whole Life, Anticipated Endowment, Convertible
  Whole Life, Children's policies)

- Storage and retrieval of premium rate tables (first year, renewal)

- Storage and retrieval of bonus / reversionary bonus rates and terminal
  bonus rules

- Computation engine for premium, maturity value, bonus accrual,
  surrender value, paid‑up value, etc.

- Versioning of rate tables (yearly / financial year basis)

- Integration with the quoting / proposal / issuance / premium
  collection subsystems.

- Validation of inputs (age, sum assured, term etc.)

- APIs / interface to read / update product setup data

## **3. Product Categories**

Postal Life Insurance has following categories of Products.

**PLI Products**:

- Whole Life Assurance (Suraksha)

- Endowment Assurance (Santosh)

- Convertible Whole Life Assurance (Suvidha)

- Anticipated Endowment Assurance (Sumangal)

- Joint Life Assurance (Yugal Suraksha)

- Children Policy (Bal Jeevan Bima)

**RPLI Products**:

- Whole Life Assurance (Gram Suraksha)

- Endowment Assurance (Gram Santosh)

- Convertible Whole Life Assurance (Gram Suvidha)

- Anticipated Endowment Assurance (Gram Sumangal)

- 10-Year RPLI (Gram Priya)

- Children Policy (Gram Bal Jeevan Bima)

### 3.1 PLI Products

#### 3.1.1 Whole Life Assurance (Suraksha)

  ------------------------------------------------------------------------
  Requirement ID  Requirement
  --------------- --------------------------------------------------------
  PR_PLI_WL_001   Minimum Age for entry is 19 Years

  PR_PLI_WL_002   Maximum Age for entry is 55 Years

  PR_PLI_WL_003   Minimum Sum Assured is ₹ 20,000

  PR_PLI_WL_004   Maximum Sum Assured is ₹ 50,00,000 (50 Lakh)

  PR_PLI_WL_005   Loan Facility is allowed after 4 Years

  PR_PLI_WL_006   Surrender Facility is allowed after 3 Years

  PR_PLI_WL_007   Policy is not eligible for bonus if surrendered before 5
                  years

  PR_PLI_WL_008   Accrued bonus is payable to the insured either on
                  attaining the age of 80 years, or to his/her legal
                  representatives or assignees on death of the insured,
                  whichever occurs earlier, provided the policy is in
                  force on the date of claim.

  PR_PLI_WL_009   Policy can be converted into Endowment Assurance Policy
                  up to 59 years of age of the insurant provided the date
                  of conversion does not fall within one year of the date
                  of cessation of premium payment or date of maturity.

  PR_PLI_WL_010   Premium paying age can be opted for as 55,58 or 60
                  years.

  PR_PLI_WL_011   Proportionate bonus on reduced sum assured as per
                  table-1 is paid if policy is surrendered.
  ------------------------------------------------------------------------

**[Table-1: Bonus Paid on Reduced Sum-Assured on
Surrender]{.underline}**

#### 3.1.2 Endowment Assurance (Santosh)

  ------------------------------------------------------------------------
  Requirement ID  Requirement
  --------------- --------------------------------------------------------
  PR_PLI_EA_001   Minimum Age for entry is 19 Years

  PR_PLI_EA_002   Maximum Age for entry is 55 Years

  PR_PLI_EA_003   Minimum Sum Assured is ₹ 20,000

  PR_PLI_EA_004   Maximum Sum Assured is ₹ 50,00,000 (50 Lakh)

  PR_PLI_EA_005   Loan Facility is allowed after 3 Years

  PR_PLI_EA_006   Surrender Facility is allowed after 3 Years

  PR_PLI_EA_007   Policy is not eligible for bonus if surrendered before 5
                  years

  PR_PLI_EA_008   Accrued bonus is paid on sum assured till attainment of
                  maturity age or death of insured, whichever is earlier.

  PR_PLI_EA_009   Maturity date for the policy can be 35,40,45,50,55,58 &
                  60 years of age.

  PR_PLI_EA_010   Proportionate bonus on reduced sum assured as per
                  table-1 is paid if policy is surrendered.
  ------------------------------------------------------------------------

#### 3.1.3 Convertible Whole Life Assurance (Suvidha)

  -------------------------------------------------------------------------
  Requirement ID   Requirement
  ---------------- --------------------------------------------------------
  PR_PLI_CWL_001   Minimum Age for entry is 19 Years

  PR_PLI_CWL_002   Maximum Age for entry is 50 Years

  PR_PLI_CWL_003   Minimum Sum Assured is ₹ 20,000

  PR_PLI_CWL_004   Maximum Sum Assured is ₹ 50,00,000 (50 Lakh)

  PR_PLI_CWL_005   Loan Facility is allowed after 4 Years

  PR_PLI_CWL_006   Surrender Facility is allowed after 3 Years

  PR_PLI_CWL_007   Policy is not eligible for bonus if surrendered before 5
                   years

  PR_PLI_CWL_008   Policy can be converted into Endowment Assurance after 5
                   years not later than 6 years of taking policy. If not
                   converted, policy will be treated as Whole Life
                   Assurance

  PR_PLI_CWL_009   Accrued bonus is paid on sum assured till attainment of
                   maturity age or death of insured, whichever is earlier.

  PR_PLI_CWL_010   On conversion, bonus of Endowment Assurance will be
                   payable.

  PR_PLI_CWL_011   Proportionate bonus on reduced sum assured as per
                   table-1 is paid if policy is surrendered.
  -------------------------------------------------------------------------

#### 3.1.4 Anticipated Endowment Assurance (Sumangal)

  -------------------------------------------------------------------------
  Requirement ID   Requirement
  ---------------- --------------------------------------------------------
  PR_PLI_AEA_001   Minimum Age for entry is 19 Years

  PR_PLI_AEA_002   Maximum Age for entry is 40 years for 20 years\' term
                   policy & 45 years for 15 years\' term policy

  PR_PLI_AEA_003   Minimum Sum Assured is ₹ 20,000

  PR_PLI_AEA_004   Maximum Sum Assured is ₹ 50,00,000 (50 Lakh)

  PR_PLI_AEA_005   Policy Term can be 15 Years and 20 Years.

  PR_PLI_AEA_006   Loan Facility is not allowed for this policy.

  PR_PLI_AEA_007   Surrender Facility is not allowed for this policy.

  PR_PLI_AEA_008   Accrued bonus is paid on sum assured till attainment of
                   maturity age or death of insured, whichever is earlier.

  PR_PLI_AEA_009   Survival Benefit paid periodically for the 15-year term
                   policy is: 20% each on completion of 6 years, 9 years &
                   12 years and 40% with accrued bonus on maturity.

  PR_PLI_AEA_010   Survival Benefit paid periodically for the 20-year term
                   policy is: 20% each on completion of 8 years, 12 years &
                   16 years and 40% with accrued bonus on maturity.
  -------------------------------------------------------------------------

#### 3.1.5 Joint Life Assurance (Yugal Suraksha)

  ------------------------------------------------------------------------
  Requirement ID  Requirement
  --------------- --------------------------------------------------------
  PR_PLI_JL_001   This policy type gives Life cover to both spouses to the
                  extent of sum assured with accrued bonus with a single
                  premium. Person having more than one spouse living,
                  cover will be available in respect of eldest spouse.

  PR_PLI_JL_002   Minimum Age for entry is 21 Years

  PR_PLI_JL_003   Policy Term is Minimum 5 Years and Maximum 20 Years.

  PR_PLI_JL_004   Maximum Age for entry is 45 Years

  PR_PLI_JL_005   Minimum Sum Assured is ₹ 20,000

  PR_PLI_JL_006   Maximum Sum Assured is ₹ 50,00,000 (50 Lakh)

  PR_PLI_JL_007   Loan Facility is allowed after 3 Years

  PR_PLI_JL_008   Surrender Facility is allowed after 3 Years

  PR_PLI_JL_009   Policy is not eligible for bonus if surrendered before 5
                  years

  PR_PLI_JL_010   Accrued bonus is payable to the insured either on
                  attaining the maturity years, or to his/her legal
                  representatives or assignees on death of the insured,
                  whichever occurs earlier, provided the policy is in
                  force on the date of claim.

  PR_PLI_JL_011   Proportionate bonus on reduced sum assured as per
                  table-1 is paid if policy is surrendered.
  ------------------------------------------------------------------------

#### 3.1.6 Children Policy (Bal Jeevan Bima)

  ------------------------------------------------------------------------
  Requirement ID  Requirement
  --------------- --------------------------------------------------------
  PR_PLI_CP_001   The scheme provides life insurance cover to children of
                  policy holders. No premium is payable, in case of death
                  of Main Policy Holder.

  PR_PLI_CP_002   Minimum Age for entry is 5 Years

  PR_PLI_CP_003   Maximum Age for entry is 20 Years

  PR_PLI_CP_004   Minimum Sum Assured is ₹ 20,000

  PR_PLI_CP_005   Maximum Sum Assured is ₹ 3 lakh or equal to the sum
                  assured of the parent, whichever is less. Maximum 2
                  children are eligible. Only one Policy for each child.

  PR_PLI_CP_006   Loan Facility is not available for this policy.

  PR_PLI_CP_007   Surrender Facility is not available for this policy.

  PR_PLI_CP_008   Has facility for making it paid up, provided premiums
                  are paid continuously for 5 years.

  PR_PLI_CP_009   Accrued bonus is paid on sum assured till attainment of
                  maturity age or death of insured, whichever is earlier.

  PR_PLI_CP_010   No premium to be paid on the Children Policy, on the
                  death of policy holder (parent). Full sum assured and
                  bonus accrued shall be paid on completion of term
  ------------------------------------------------------------------------

### 3.2 RPLI Products

#### 3.2.1 Whole Life Assurance (Gram Suraksha)

  -------------------------------------------------------------------------
  Requirement ID   Requirement
  ---------------- --------------------------------------------------------
  PR_RPLI_WL_001   Minimum Age for entry is 19 Years

  PR_RPLI_WL_002   Maximum Age for entry is 55 Years

  PR_RPLI_WL_003   Minimum Sum Assured is ₹ 10,000

  PR_RPLI_WL_004   Maximum Sum Assured is ₹ 10,00,000 (10 Lakh)

  PR_RPLI_WL_005   Loan Facility is allowed after 4 Years

  PR_RPLI_WL_006   Surrender Facility is allowed after 3 Years

  PR_RPLI_WL_007   Policy is not eligible for bonus if surrendered before 5
                   years

  PR_RPLI_WL_008   Accrued bonus is payable to the insured either on
                   attaining the age of 80 years, or to his/her legal
                   representatives or assignees on death of the insured,
                   whichever occurs earlier, provided the policy is in
                   force on the date of claim.

  PR_RPLI_WL_009   Policy can be converted into Endowment Assurance Policy
                   up to 59 years of age of the insurant provided the date
                   of conversion does not fall within one year of the date
                   of cessation of premium payment or date of maturity.

  PR_RPLI_WL_010   Premium paying age can be opted for as 55,58 or 60
                   years.

  PR_RPLI_WL_011   Proportionate bonus on reduced sum assured as per
                   table-1 is paid if policy is surrendered.
  -------------------------------------------------------------------------

#### 3.2.2 Endowment Assurance (Gram Santosh)

  -------------------------------------------------------------------------
  Requirement ID   Requirement
  ---------------- --------------------------------------------------------
  PR_RPLI_EA_001   Minimum Age for entry is 19 Years

  PR_RPLI_EA_002   Maximum Age for entry is 55 Years

  PR_RPLI_EA_003   Minimum Sum Assured is ₹ 10,000

  PR_RPLI_EA_004   Maximum Sum Assured is ₹ 10,00,000 (10 Lakh)

  PR_RPLI_EA_005   Loan Facility is allowed after 3 Years

  PR_RPLI_EA_006   Surrender Facility is allowed after 3 Years

  PR_RPLI_EA_007   Policy is not eligible for bonus if surrendered before 5
                   years

  PR_RPLI_EA_008   Accrued bonus is paid on sum assured till attainment of
                   maturity age or death of insured, whichever is earlier.

  PR_RPLI_EA_009   Maturity date for the policy can be 35,40,45,50,55,58 &
                   60 years of age.

  PR_RPLI_EA_010   Proportionate bonus on reduced sum assured as per
                   table-1 is paid if policy is surrendered.
  -------------------------------------------------------------------------

#### 3.2.3 Convertible Whole Life Assurance (Gram Suvidha)

  -------------------------------------------------------------------------
  Requirement ID    Requirement
  ----------------- -------------------------------------------------------
  PR_RPLI_CWL_001   Minimum Age for entry is 19 Years

  PR_RPLI_CWL_002   Maximum Age for entry is 45 Years

  PR_RPLI_CWL_003   Minimum Sum Assured is ₹ 10,000

  PR_RPLI_CWL_004   Maximum Sum Assured is ₹ 10,00,000 (10 Lakh)

  PR_RPLI_CWL_005   Loan Facility is allowed after 4 Years

  PR_RPLI_CWL_006   Surrender Facility is allowed after 3 Years

  PR_RPLI_CWL_007   Policy is not eligible for bonus if surrendered before
                    5 years

  PR_RPLI_CWL_008   Policy can be converted into Endowment Assurance after
                    5 years not later than 6 years of taking policy. If not
                    converted, policy will be treated as Whole Life
                    Assurance

  PR_RPLI_CWL_009   Accrued bonus is paid on sum assured till attainment of
                    maturity age or death of insured, whichever is earlier.

  PR_RPLI_CWL_010   On conversion, bonus of Endowment Assurance will be
                    payable.

  PR_RPLI_CWL_011   Proportionate bonus on reduced sum assured as per
                    table-1 is paid if policy is surrendered.
  -------------------------------------------------------------------------

#### 3.2.4 Anticipated Endowment Assurance (Gram Sumangal)

  -------------------------------------------------------------------------
  Requirement ID    Requirement
  ----------------- -------------------------------------------------------
  PR_RPLI_AEA_001   Minimum Age for entry is 19 Years

  PR_RPLI_AEA_002   Maximum Age for entry is 40 years.

  PR_RPLI_AEA_003   Minimum Sum Assured is ₹ 10,000

  PR_RPLI_AEA_004   Maximum Sum Assured is ₹ 10,00,000 (10 Lakh)

  PR_RPLI_AEA_005   Loan Facility is not allowed for this policy.

  PR_RPLI_AEA_006   Policy Term can be 15 Years and 20 Years.

  PR_RPLI_AEA_007   Surrender Facility is allowed after 3 Years

  PR_RPLI_AEA_008   Policy is not eligible for bonus if surrendered before
                    5 years

  PR_RPLI_AEA_009   Accrued bonus is paid on sum assured till attainment of
                    maturity age or death of insured, whichever is earlier.

  PR_RPLI_AEA_010   Survival Benefit paid periodically for the 15-year term
                    policy is: 20% each on completion of 6 years, 9 years &
                    12 years and 40% with accrued bonus on maturity.

  PR_RPLI_AEA_011   Survival Benefit paid periodically for the 20-year term
                    policy is: 20% each on completion of 8 years, 12 years
                    & 16 years and 40% with accrued bonus on maturity.

  PR_RPLI_AEA_012   Proportionate bonus on reduced sum assured as per
                    table-1 is paid if policy is surrendered.
  -------------------------------------------------------------------------

#### 3.2.5 10-Year Rural PLI (Gram Priya)

  -------------------------------------------------------------------------
  Requirement ID   Requirement
  ---------------- --------------------------------------------------------
  PR_RPLI_TY_001   Minimum Age for entry is 20 Years

  PR_RPLI_TY_002   Maximum Age for entry is 45 Years.

  PR_RPLI_TY_003   Policy Term is 10 Years.

  PR_RPLI_TY_004   Minimum Sum Assured is ₹ 10,000

  PR_RPLI_TY_005   Maximum Sum Assured is ₹ 10,00,000 (10 Lakh)

  PR_RPLI_TY_006   Loan Facility is not available for this policy.

  PR_RPLI_TY_007   Surrender Facility is not available for this policy.

  PR_RPLI_TY_008   No interest is charged up to one year as arrears of
                   premia in case of natural calamities like flood,
                   drought, earthquake, cyclone etc.

  PR_RPLI_TY_009   Accrued bonus is payable to the insured either on
                   attaining the maturity years, or to his/her legal
                   representatives or assignees on death of the insured,
                   whichever occurs earlier, provided the policy is in
                   force on the date of claim.

  PR_RPLI_TY_010   Survival benefits are paid after 4 years- 20% after 7
                   years- 20% and after 10 years -- 60% with accrued bonus.
  -------------------------------------------------------------------------

#### 3.2.6 Children Policy (Gram Bal Jeevan Bima)

  -------------------------------------------------------------------------
  Requirement ID   Requirement
  ---------------- --------------------------------------------------------
  PR_RPLI_CP_001   The scheme provides life insurance cover to children of
                   policy holders. No premium is payable, in case of death
                   of Main Policy Holder.

  PR_RPLI_CP_002   Minimum Age for entry is 5 Years

  PR_RPLI_CP_003   Maximum Age for entry is 20 Years. Policy holder
                   (parent) should not be over 45 years of age.

  PR_RPLI_CP_004   Minimum Sum Assured is ₹ 10,000

  PR_RPLI_CP_005   Maximum Sum Assured is ₹ 1 lakh or equal to the sum
                   assured of the parent, whichever is less. Maximum 2
                   children are eligible. Only one Policy for each child.

  PR_RPLI_CP_006   Loan Facility is not available for this policy.

  PR_RPLI_CP_007   Surrender Facility is not available for this policy.

  PR_RPLI_CP_008   Has facility for making it paid up, provided premiums
                   are paid continuously for 5 years.

  PR_RPLI_CP_009   Accrued bonus is paid on sum assured till attainment of
                   maturity age or death of insured, whichever is earlier.

  PR_RPLI_CP_010   No premium to be paid on the Children Policy, on the
                   death of policy holder (parent). Full sum assured and
                   bonus accrued shall be paid on completion of term
  -------------------------------------------------------------------------

## **4. Functional Requirements Specification**

## 4.1 Product Configuration

+-------------+--------------------------------------------------------+
| Requirement | Requirement                                            |
| ID          |                                                        |
+=============+========================================================+
| FS_PC_001   | Admin can create new insurance products with unique    |
|             | codes.                                                 |
+-------------+--------------------------------------------------------+
| FS_PC_002   | Admin can define product type (Endowment, Whole Life,  |
|             | etc.).                                                 |
+-------------+--------------------------------------------------------+
| FS_PC_003   | Admin can set eligibility criteria (age, sum assured   |
|             | range, etc.).                                          |
+-------------+--------------------------------------------------------+
| FS_PC_004   | Admin can define premium tables based on:              |
|             |                                                        |
|             | - Age                                                  |
|             |                                                        |
|             | - Sum Assured                                          |
|             |                                                        |
|             | - Policy Term                                          |
|             |                                                        |
|             | - Frequency (Monthly, Quarterly, Half-Yearly, Yearly)  |
+-------------+--------------------------------------------------------+

## 4.2 Bonus Configuration

+-------------+--------------------------------------------------------+
| Requirement | Requirement                                            |
| ID          |                                                        |
+=============+========================================================+
| FS_PC_005   | Admin can define annual bonus rates (e.g., ₹50 or ₹65  |
|             | per ₹1000 sum assured).                                |
+-------------+--------------------------------------------------------+
| FS_PC_006   | Bonus Applicability as per criteria defined for the    |
|             | respective products in product setup (e.g. After 3     |
|             | years)                                                 |
+-------------+--------------------------------------------------------+
| FS_PC_007   | Bonus Calculation Logic:                               |
|             |                                                        |
|             | Annual Bonus = (Sum Assured / 1000) × Bonus Rate       |
+-------------+--------------------------------------------------------+

## 4.3 Premium Calculation

+-------------+--------------------------------------------------------+
| Requirement | Requirement                                            |
| ID          |                                                        |
+=============+========================================================+
| FS_PC_008   | System calculates premium as given in premium table    |
|             | based on:                                              |
|             |                                                        |
|             | - Entry Age                                            |
|             |                                                        |
|             | - Sum Assured                                          |
|             |                                                        |
|             | - Policy Term                                          |
|             |                                                        |
|             | - Frequency                                            |
+-------------+--------------------------------------------------------+
| FS_PC_009   | Premium calculation formula:                           |
|             |                                                        |
|             | Premium = (Base Rate × Sum Assured)                    |
+-------------+--------------------------------------------------------+

## 4.4 Maturity & Survival Benefits

+-------------+--------------------------------------------------------+
| Requirement | Requirement                                            |
| ID          |                                                        |
+=============+========================================================+
| FS_PC_010   | Define maturity age and payout rules as per the        |
|             | formula given.                                         |
+-------------+--------------------------------------------------------+
| FS_PC_011   | Configure survival benefits (e.g., 20% payout at 4, 7, |
|             | 10 years for Sumangal).                                |
+-------------+--------------------------------------------------------+
| FS_PC_012   | Maturity calculation:                                  |
|             |                                                        |
|             | Maturity Amount = Sum Assured + Accrued Bonus -        |
|             | Survival Benefits (if any)                             |
+-------------+--------------------------------------------------------+

## 4.5 Loan & Surrender Rules

  -----------------------------------------------------------------------
  Requirement ID Requirement
  -------------- --------------------------------------------------------
  FS_PC_013      Define loan eligibility (e.g., after 3 years).

  FS_PC_014      Define surrender value rules and penalties.
  -----------------------------------------------------------------------

## 4.6 Policy Conversion

  -----------------------------------------------------------------------
  Requirement ID Requirement
  -------------- --------------------------------------------------------
  FS_PC_015      Enable conversion of Whole Life to Endowment (e.g.,
                 Suvidha to Santosh) after 5 years.

  -----------------------------------------------------------------------

## **5. Wireframe**

## **6. Appendices**

The Following Documents attached below can be used.
