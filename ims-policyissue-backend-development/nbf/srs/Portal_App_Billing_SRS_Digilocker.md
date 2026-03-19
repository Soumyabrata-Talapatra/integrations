> INTERNAL APPROVAL FORM

**Project Name:** Digilocker

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

[4.1 Use Case-1: Generate Policy Bond & Save in Digilocker
[5](#use-case-1-generate-policy-bond-save-in-digilocker)](#use-case-1-generate-policy-bond-save-in-digilocker)

[4.2 Use Case-2: Submit Aadhaar, PAN, and Other Required Documents for
Policy Issue/Claim
[6](#use-case-2-submit-aadhaar-pan-and-other-required-documents-for-policy-issueclaim)](#use-case-2-submit-aadhaar-pan-and-other-required-documents-for-policy-issueclaim)

[**5. Attachments** [7](#attachments)](#attachments)

## **1. Executive Summary**

This document outlines the requirements for integrating the Insurance
Management System (IMS) of India Post PLI with the Digilocker to enable
customers to generate and store e-Bonds digitally.

## **2. Project Scope**

The integration will allow PLI/RPLI customers to:

- Access Digilocker via app or URL.

- Authenticate using Aadhaar-linked Digilocker account.

- Fetch policy details from IMS.

- Generate and store e-Bonds in Digilocker.

- Submit documents for policy procurement or claims through Digilocker.

## **3. Business Requirements**

  -----------------------------------------------------------------------
  **ID**        **Requirements**
  ------------- ---------------------------------------------------------
  FS_DL_001     Enable integration between IMS and Digilocker for e-Bond
                generation. Generate e-Bond in PDF format and store in
                Digilocker.

  FS_DL_002     Provide Aadhaar-based authentication for auto-fetching
                Name and DOB.

  FS_DL_003     Send SMS/email confirmation with e-Bond reference number
                after successful generation.

  FS_DL_004     Provide a "Report an Issue" button linked to PLI helpdesk
                for customer support.

  FS_DL_005     Enable submission of documents for policy procurement or
                claims through Digilocker.
  -----------------------------------------------------------------------

## **4. Functional Requirements Specification**

**Use Case Diagram for Digilocker Integration with IMS:** ![A diagram of
a device AI-generated content may be
incorrect.](media/image1.png){width="6.268055555555556in"
height="4.178472222222222in"}

### 4.1 Use Case-1: Generate Policy Bond & Save in Digilocker

- **Purpose:** To allow an insured customer to generate an electronic
  policy bond (e-Bond) for their PLI/RPLI policy and securely store it
  in their Digilocker account.

- **Main Flow:**

  - Customer opens Digilocker app or URL and logs in using Aadhaar
    credentials.

  - Navigates to Financial Services → PLI.

  - Inputs required details (Policy Number; Name and DOB auto-fetched
    via Aadhaar).

  - Digilocker sends request to IMS to validate policy details.

  - IMS verifies details and returns policy information.

  - IMS generates e-Bond in PDF format and sends it to Digilocker.

  - Digilocker stores e-Bond in customer's account.

  - Customer receives SMS/email confirmation with e-Bond reference
    number.

**Use Case Diagram for Generate Policy Bond & Save in Digilocker Use
Case:**![](media/image2.png){width="6.268055555555556in"
height="4.178472222222222in"}

### 4.2 Use Case-2: Submit Aadhaar, PAN, and Other Required Documents for Policy Issue/Claim

- **Purpose:** To enable customers to submit KYC and other required
  documents (Aadhaar, PAN, income proof, claim documents) through
  Digilocker for policy issuance, servicing, or claim settlement.

- **Main Flow:**

  - Customer logs into Digilocker and navigates to Financial Services →
    PLI.

  - Selects option to submit documents for policy issuance or claim.

  - Chooses required documents (Aadhaar, PAN, etc.) from Digilocker
    repository.

  - Digilocker sends selected documents securely to IMS via API.

  - IMS receives and validates documents for completeness and
    authenticity.

  - IMS updates policy or claim status accordingly.

  - Customer receives confirmation of successful submission.

**Use Case Diagram for Submit Aadhaar, PAN, and Other Required Documents
for Policy Issue/Claim Use Case:**
![](media/image3.png){width="6.268055555555556in"
height="4.178472222222222in"}

## **5. Attachments**

The following documents can be referred.
