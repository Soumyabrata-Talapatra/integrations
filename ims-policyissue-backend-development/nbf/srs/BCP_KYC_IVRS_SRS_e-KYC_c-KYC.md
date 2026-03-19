> INTERNAL APPROVAL FORM

**Project Name:** e-KYC/c-KYC

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

[4.1 e-KYC [5](#e-kyc)](#e-kyc)

[4.1.1 Request Structure for (IMS -\> UIDAI e-KYC API)
[5](#request-structure-for-ims---uidai-e-kyc-api)](#request-structure-for-ims---uidai-e-kyc-api)

[4.1.2 Response Structure (UIDAI e-KYC API -\> IMS)
[5](#response-structure-uidai-e-kyc-api---ims)](#response-structure-uidai-e-kyc-api---ims)

[4.1.3 Error Codes [6](#error-codes)](#error-codes)

[4.2 c-KYC [6](#c-kyc)](#c-kyc)

[4.2.1 Request Structure for (IMS -\> CKYC API)
[6](#request-structure-for-ims---ckyc-api)](#request-structure-for-ims---ckyc-api)

[4.2.2 Response Structure (CKYC API -\> IMS)
[6](#response-structure-ckyc-api---ims)](#response-structure-ckyc-api---ims)

[4.2.3 Error Codes [6](#error-codes-1)](#error-codes-1)

[**5. Attachments** [6](#attachments)](#attachments)

## **1. Executive Summary**

This document outlines the requirements to enable digital KYC
verification using Aadhaar-based e-KYC and Central KYC Registry (CKYC)
for customer onboarding, servicing, and compliance in IMS.

## **2. Project Scope**

This Integration will:

- e-KYC: Aadhaar-based authentication for identity verification.

- CKYC: Fetch and update customer KYC records from Central KYC Registry
  as per IRDAI AML/CFT guidelines.

## **3. Business Requirements**

  ------------------------------------------------------------------------
  **ID**       **Requirements**
  ------------ -----------------------------------------------------------
  FS_KYC_001   Implement e-KYC (Aadhaar-based) and CKYC verification as
               per UIDAI, IRDAI, and PMLA guidelines, ensuring explicit
               customer consent is captured and stored.

  FS_KYC_002   Enable KYC verification during onboarding, policy
               servicing, and claims; block transactions if KYC fails or
               consent is missing.

  FS_KYC_003   Encrypt Aadhaar and CKYC data, use secure API communication
               (TLS, digital signatures), and store only masked
               identifiers; comply with UIDAI and CKYC security norms.

  FS_KYC_004   Support real-time KYC verification with quicker response
               time, handle bulk requests, and implement fallback
               mechanisms for API downtime.

  FS_KYC_005   Maintain logs of all KYC verification attempts, consent
               records, and generate compliance reports; detect duplicate
               or fraudulent identities and flag anomalies.
  ------------------------------------------------------------------------

**Flow Diagram for e-KYC/c-KYC Integration:**
![](media/image1.png){width="6.268055555555556in"
height="4.178472222222222in"}

## **4. Functional Requirements Specification**

## 4.1 e-KYC

### 4.1.1 Request Structure for (IMS -\> UIDAI e-KYC API)

  ------------------------------------------------------------------------
  Field Name      Type          Description
  --------------- ------------- ------------------------------------------
  aadhar          String        12-digit Aadhaar number

  otp             String        One-Time Password for authentication

  txn             String        Unique transaction ID

  consent         Boolean       Customer consent flag
  ------------------------------------------------------------------------

### 4.1.2 Response Structure (UIDAI e-KYC API -\> IMS)

  ------------------------------------------------------------------------
  Field Name       Type          Description
  ---------------- ------------- -----------------------------------------
  status           String        Success / Failure

  name             String        Name as per Aadhaar

  dob              Date          Date of Birth

  gender           String        Gender

  address          String        Full address

  photo            Base64        Photo from Aadhaar

  mobile           String        Registered mobile number

  email            String        Registered email address
  ------------------------------------------------------------------------

### 4.1.3 Error Codes

- 100: Success

- 300: Invalid Aadhaar

- 400: Invalid OTP

- 500: Consent missing

## 4.2 c-KYC

### 4.2.1 Request Structure for (IMS -\> CKYC API)

  ------------------------------------------------------------------------
  Field Name      Type          Description
  --------------- ------------- ------------------------------------------
  fi_code         String        Financial Institution code

  request_id      String        Unique request identifier

  id_type         String        Type of ID (CKYC)

  id_no           String        CKYC number (14 digits)

  date_time       DateTime      Request timestamp
  ------------------------------------------------------------------------

### 4.2.2 Response Structure (CKYC API -\> IMS)

  ------------------------------------------------------------------------
  Field Name       Type          Description
  ---------------- ------------- -----------------------------------------
  status           String        Success / Failure

  ckyc_no          String        CKYC number

  name             String        Customer name

  dob              Date          Date of Birth

  address          String        Customer Address

  Kyc_type         String        KYC type (Normal / Simplified)
  ------------------------------------------------------------------------

### 4.2.3 Error Codes

- 200: Success

- 310: CKYC record not found

- 700: Invalid encryption

- 900: Technical error

## **5. Attachments**

The following documents can be referred.

![](media/image2.emf) ![](media/image3.emf) ![](media/image4.emf)
