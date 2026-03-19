> INTERNAL APPROVAL FORM

**Project Name:** Aadhar

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

[4.1 Request Structure (IMS -\> UIDAI Authentication API)
[5](#request-structure-ims---uidai-authentication-api)](#request-structure-ims---uidai-authentication-api)

[4.2 Response Structure (UIDAI Authentication API -\> IMS)
[6](#response-structure-uidai-authentication-api---ims)](#response-structure-uidai-authentication-api---ims)

[4.3 Error Codes [6](#error-codes)](#error-codes)

[**5. Attachments** [6](#attachments)](#attachments)

## **1. Executive Summary**

This document outlines the requirements to enable Aadhaar-based identity
verification and e-KYC during customer onboarding and servicing to
comply with IRDAI and UIDAI norms, reduce fraud, and streamline KYC
processes.

## **2. Project Scope**

This Integration will:

- Aadhaar Authentication (Biometric, OTP, Demographic)

- Aadhaar e-KYC (Online and Offline XML-based)

- Secure integration with UIDAI via Authentication User Agency (AUA) and
  Authentication Service Agency (ASA)

## **3. Business Requirements**

  -----------------------------------------------------------------------
  **ID**      **Requirements**
  ----------- -----------------------------------------------------------
  FS_AD_001   Implement Aadhaar-based authentication and e-KYC as per
              UIDAI and IRDAI norms, ensuring explicit customer consent
              is captured and stored.

  FS_AD_002   Enable Aadhaar verification during onboarding, policy
              servicing, and updates; block transactions if
              authentication fails or consent is missing.

  FS_AD_003   Encrypt Aadhaar data (PID block, demographic details), use
              UIDAI-approved encryption standards, secure API calls
              (TLS), and store only masked Aadhaar numbers.

  FS_AD_004   Support real-time Aadhaar authentication with quicker
              response time, handle high-volume requests, and implement
              fallback mechanisms for API downtime.

  FS_AD_005   Maintain detailed logs of authentication attempts, consent
              records, and generate compliance reports for audits; detect
              anomalies and prevent identity fraud.
  -----------------------------------------------------------------------

**Flow Diagram for Aadhar Integration:**
![](media/image1.png){width="6.268055555555556in"
height="4.178472222222222in"}

## **4. Functional Requirements Specification**

**Data Fields to be Created in IMS for Integration with Aadhar**

### 4.1 Request Structure (IMS -\> UIDAI Authentication API)

  ------------------------------------------------------------------------
  Field Name      Type          Description
  --------------- ------------- ------------------------------------------
  uid             String        Aadhaar number (12 digits)

  txn             String        Unique transaction ID (UUID)

  ts              DateTime      Timestamp in ISO format

  type            String        Authentication type (OTP / BIO / DEMO)

  pid             Encrypted     Personal Identity Data block (encrypted)

  skey            Encrypted     Session key (encrypted)

  hmac            Encrypted     Hash for integrity check

  meta            Object        Device info (IP, location, device ID)

  consent         Boolean       Customer consent flag
  ------------------------------------------------------------------------

### 4.2 Response Structure (UIDAI Authentication API -\> IMS)

  ------------------------------------------------------------------------
  Field Name       Type          Description
  ---------------- ------------- -----------------------------------------
  txn              String        Same transaction ID for mapping

  ret              String        Authentication result (y = success, n =
                                 failure)

  code             String        Response code

  err              String        Error Code, if any

  info             String        Status message

  ts               DateTime      Response timestamp
  ------------------------------------------------------------------------

### 4.3 Error Codes

- 100: Success -- Authentication successful

- 300: Invalid Aadhaar number

- 310: Aadhaar number does not exist

- 400: Invalid OTP

- 401: OTP expired

- 500: Biometric data mismatch

- 510: Biometric data quality check failed

- 600: Invalid PID block

- 700: Invalid encryption or HMAC

- 800: Invalid AUA/ASA credentials

- 810: Digital signature verification failed

- 900: Technical error at UIDAI server

- 910: Service unavailable

- 999: Unknown error

## **5. Attachments**

The following documents can be referred.
