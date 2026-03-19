> INTERNAL APPROVAL FORM

**Project Name:** Quote Engine (New Business, Loan, Surrender,
Commutation, Reduce Paid up, Revival, Conversion, Loan Payment, Loan
Repayment & Premium Paid Certificate)

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

#  {#section .TOC-Heading}

#  {#section-1 .TOC-Heading}

# Table of Contents {#table-of-contents .TOC-Heading}

[**1. Executive Summary** [4](#executive-summary)](#executive-summary)

[**2. Project Scope** [4](#project-scope)](#project-scope)

[**3. Business Requirements**
[4](#business-requirements)](#business-requirements)

[**4. Functional Requirements Specification**
[5](#functional-requirements-specification)](#functional-requirements-specification)

[4.1 Quote Generation for New Business
[5](#quote-generation-for-new-business)](#quote-generation-for-new-business)

[4.2 Quote Generation for Other Transactions
[7](#quote-generation-for-other-transactions)](#quote-generation-for-other-transactions)

[**5. Attachments** [8](#attachments)](#attachments)

## **1. Executive Summary**

The purpose of this document is to define the functional and
non-functional requirements for the Quote Engine that will generate
quotes for various transactions under India Post PLI/RPLI, including:

- New Business

- Loan

- Surrender

- Commutation

- Reduce Paid-Up

- Revival

- Conversion

- Loan Payment & Loan Repayment

- Premium Paid Certificate

## **2. Project Scope**

This Module will:

- Provide accurate premium and benefit calculations for new and existing
  policies.

- Generate downloadable and shareable quote documents (PDF).

## **3. Business Requirements**

+------------+---------------------------------------------------------+
| **ID**     | **Requirements**                                        |
+============+=========================================================+
| FS_QE_001  | **Quote Generation for New Business**                   |
|            |                                                         |
|            | - Allow customers, agents, and counter staff to         |
|            |   generate quotes for new PLI/RPLI policies.            |
|            |                                                         |
|            | - Capture proposer details (Name, Gender, DOB, Contact  |
|            |   Info, Category, Location).                            |
|            |                                                         |
|            | - Enable selection of policy type, plan type, sum       |
|            |   assured, term, payment mode, and riders.              |
|            |                                                         |
|            | - Validate eligibility criteria (age, term limits, sum  |
|            |   assured).                                             |
|            |                                                         |
|            | - Calculate premium using rate tables and display GST   |
|            |   and total payable amount.                             |
|            |                                                         |
|            | - Provide maturity value and indicative bonus in quote  |
|            |   summary.                                              |
|            |                                                         |
|            | - Allow modification of inputs and regeneration of      |
|            |   quote.                                                |
|            |                                                         |
|            | - Generate Quote Reference Number and provide           |
|            |   download/print/email options.                         |
+------------+---------------------------------------------------------+
| FS_QE_002  | Support quotes for:                                     |
|            |                                                         |
|            | - **Surrender:** Calculate surrender value, integrate   |
|            |   loan recovery, show net payable.                      |
|            |                                                         |
|            | - **Revival:** Include GST heads, handle installment    |
|            |   failures, show revivals available.                    |
|            |                                                         |
|            | - **Commutation:** Implement commutation logic.         |
|            |                                                         |
|            | - **Conversion:** Correct premium calculation for       |
|            |   age-band conversion, add disclaimers.                 |
|            |                                                         |
|            | - **Loan:** Calculate eligible loan amount as per       |
|            |   product rules, add disclaimers.                       |
|            |                                                         |
|            | - **Reduced Paid-Up:** Calculate surrender value, show  |
|            |   disclaimer about non-revival.                         |
|            |                                                         |
|            | - **Premium Paid Certificate:** Provide downloadable    |
|            |   certificate.                                          |
+------------+---------------------------------------------------------+
| FS_QE_003  | Fetch policy details from IMS database (Policy No,      |
|            | Product Type, Premium, Loan Balance, Bonus, Status).    |
+------------+---------------------------------------------------------+
| FS_QE_004  | Display quote summary with all relevant details         |
|            | (premium, GST, recoveries, net amount).                 |
+------------+---------------------------------------------------------+
| FS_QE_005  | Provide options to download PDF or email to customer.   |
+------------+---------------------------------------------------------+
| FS_QE_006  | Save quote record in IMS database for audit and         |
|            | history.                                                |
+------------+---------------------------------------------------------+
| FS_QE_007  | Display disclaimers for special cases (e.g., forced     |
|            | surrender, conversion requirements).                    |
+------------+---------------------------------------------------------+
| FS_QE_008  | Support multiple channels (Customer Portal, Agent       |
|            | Portal, Counter Staff Interface).                       |
+------------+---------------------------------------------------------+

## **4. Functional Requirements Specification**

### 4.1 Quote Generation for New Business

- **Purpose:** To Generate the Quote for getting the amount premium
  amount that needs to be paid by the customer for buying the PLI/RPLI
  Policies.

- **Flow Steps:**

  - Start: User initiates request for new PLI/RPLI policy.

  - Input Basic Details: Policy Type, Plan Type, Proposer Details (Name,
    Gender, DOB, Mobile, Email, Category, Location).

  - Enter Policy Details: Sum Assured, Policy Term, Payment Mode,
    Riders/Add-ons.

  - System Validations: Eligibility checks (age, term, sum assured),
    duplicate proposal, service area availability.

  - Fetch Rate: Retrieve premium rate from premium table based on Age,
    Gender, Plan Type, Term, Frequency.

  - Calculate Premium: Compute Basic Premium, GST, Total Payable.

  - Display Quote Summary: Show plan, term, sum assured, premium amount,
    maturity value, bonus rate, total payable.

  - Modify Inputs: Allow user to edit and regenerate quote.

  - Generate & Share Quote: Create Quote Reference Number,
    download/print PDF, option to proceed for proposal creation.

  - End: Save quote record in IMS database.

**Wireframe for the New Business Quote Generation Page:**

![A screenshot of a computer AI-generated content may be
incorrect.](media/image1.png){width="6.06548665791776in"
height="9.098230533683289in"}

### 4.2 Quote Generation for Other Transactions

The system shall allow users to generate quotes for:

- **Surrender:** Calculate surrender value, integrate loan recovery,
  display net payable amount. Separate surrender calculation for CWL and
  EA Products.

- **Revival:** Include GST heads, handle instalment failures, implement
  grace period logic, show number of revivals available.

- **Commutation:** Implement logic for commutation quotes.

- **Conversion:** Correct premium calculation for age-band conversion
  (EA/45 → EA/55), resolve zero-premium issue, add disclaimer for
  premium requirement.

- **Loan:** Update eligibility logic per product rules, include
  disclaimer for forced surrender.

- **Reduced Paid-Up:** Calculate surrender value, show disclaimer that
  policy cannot be revived after RPU approval.

- **Premium Paid Certificate:** Provide downloadable certificate for
  policyholder.

**Flow Chart for Generating Quote for Other Transactions:**

![](media/image2.png){width="6.268055555555556in"
height="4.178472222222222in"}

**Wireframe for Quote Generation for Other Transactions Page:**
![](media/image3.png){width="6.268055555555556in"
height="4.178472222222222in"}

## **5. Attachments**

The following documents can be referred.

![](media/image4.emf)
