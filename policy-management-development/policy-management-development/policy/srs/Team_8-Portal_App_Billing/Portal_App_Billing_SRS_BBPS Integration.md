> INTERNAL APPROVAL FORM

**Project Name:** BBPS Integration

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

[4.1 Policy Validation API
[6](#policy-validation-api)](#policy-validation-api)

[4.2 Payment Initiation API
[6](#payment-initiation-api)](#payment-initiation-api)

[4.3 Payment Status API [7](#payment-status-api)](#payment-status-api)

[4.4 Refund/Reconciliation API
[7](#refundreconciliation-api)](#refundreconciliation-api)

[4.5 Error Codes [8](#error-codes)](#error-codes)

[**5. Attachments** [8](#attachments)](#attachments)

## **1. Executive Summary**

This document outlines the requirements for integrating the Insurance
Management System (IMS) of India Post PLI with the Bharat Bill Payment
System (BBPS) via Bharat e-Connect. The integration will enable
policyholders to pay premiums using UPI apps and other digital payment
modes.

## **2. Project Scope**

This Module will:

- Allow policyholders to pay premiums (due or advance) via BBPS.

- Enable real-time validation of policy details.

- Generate receipts instantly.

- Ensure secure and transparent transaction flow.

- Automate reconciliation through the Nodal Office.

## **3. Business Requirements**

  -------------------------------------------------------------------------------
  **ID**        **Functionality**   **Requirements**
  ------------- ------------------- ---------------------------------------------
  FS_BBPS_001   BBPS Integration    IMS must integrate with NPCI BBPS via Bharat
                                    e-Connect for premium payments.

  FS_BBPS_002   Policy Validation   Customer should enter Policy Number and DOB;
                                    IMS validates details in real-time.

  FS_BBPS_003   Premium Fetch       IMS should fetch due premium or allow advance
                                    payment (up to 12 months).

  FS_BBPS_004   Payment Options     Customer should be able to pay via UPI,
                                    Cards, Net Banking, e-Wallets.

  FS_BBPS_005   Banking Partner     SBI will act as the BBPS banking partner for
                                    transaction processing.

  FS_BBPS_006   Receipt Generation  IMS must generate and share digital receipts
                                    post successful payment.

  FS_BBPS_007   Error Handling      Display clear error messages for invalid
                                    policy or failed transactions.

  FS_BBPS_008   Refund &            Nodal Office should handle refunds and
                Reconciliation      reconciliation for failed payments.

  FS_BBPS_009   Security            All data exchanges must be encrypted and
                                    comply with RBI/NPCI norms.

  FS_BBPS_010   Compliance          Integration must adhere to BBPS and RBI
                                    guidelines for digital payments.
  -------------------------------------------------------------------------------

## **4. Functional Requirements Specification**

**Visual Flow Diagram for BBPS Integration:**
![](media/image1.png){width="5.627127077865267in"
height="8.44069116360455in"}

### 4.1 Policy Validation API

- **Purpose:** To validate the customer's policy details and fetch
  premium information before initiating payment.

#### 4.1.1 Request Structure (IMS → BBPS API: Policy Validation API)

  ---------------------------------------------------------------------
  Field Name      Type           Description
  --------------- -------------- --------------------------------------
  policyNumber    String         Unique policy number of the customer

  dateOfBirth     Date           Customer's date of birth in YYYY-MM-DD
                                 format

  txn             String         Unique transaction ID (UUID)

  ts              DateTime       Timestamp in ISO format

  channel         String         Payment channel (BBPS)

  consent         Boolean        Customer consent flag
  ---------------------------------------------------------------------

#### 

#### 4.1.2 Response Structure (BBPS → IMS: Policy Validation)

  -----------------------------------------------------------------------
  Field Name             Type          Description
  ---------------------- ------------- ----------------------------------
  txn                    String        Same transaction ID for mapping

  status                 String        Validation result (SUCCESS /
                                       ERROR)

  premiumDue             Decimal       Premium amount due

  advanceAllowedMonths   Integer       Maximum months allowed for advance
                                       payment

  message                String        Status message

  ts                     DateTime      Response timestamp
  -----------------------------------------------------------------------

### 4.2 Payment Initiation API

- **Purpose:** To initiate the payment process by sending payment
  details to BBPS and receiving a payment link or status.

#### 4.2.1 Request Structure (IMS → BBPS API: Payment Initiation API)

  ---------------------------------------------------------------------
  Field Name      Type           Description
  --------------- -------------- --------------------------------------
  policyNumber    String         Customer's policy number

  amount          Decimal        Payment amount

  paymentType     String         FULL / ADVANCE

  txn             String         Unique transaction ID

  ts              DateTime       Timestamp in ISO format

  bankPartner     String         Banking partner (SBI)

  consent         Boolean        Customer consent flag
  ---------------------------------------------------------------------

#### 4.2.2 Response Structure (BBPS → IMS: Payment Initiation API)

  ---------------------------------------------------------------------
  Field Name      Type           Description
  --------------- -------------- --------------------------------------
  txn             String         Same transaction ID

  status          String         Payment status (PENDING / SUCCESS /
                                 FAILED)

  paymentUrl      String         URL for payment completion

  message         String         Status message

  ts              DateTime       Response timestamp
  ---------------------------------------------------------------------

### 4.3 Payment Status API

- **Purpose:** To check the current status of a payment transaction.

#### 4.3.1 Request Structure (IMS → BBPS API: Payment Status API)

  ---------------------------------------------------------------------
  Field Name      Type           Description
  --------------- -------------- --------------------------------------
  transactionId   String         Unique transaction ID

  ts              DateTime       Timestamp in ISO format
  ---------------------------------------------------------------------

#### 4.3.2 Response Structure (BBPS → IMS: Payment Status API)

  ---------------------------------------------------------------------
  Field Name      Type           Description
  --------------- -------------- --------------------------------------
  txn             String         Same transaction ID

  status          String         Payment status (SUCCESS / FAILED)

  receiptUrl      String         URL for digital receipt

  message         String         Status message

  ts              DateTime       Response timestamp
  ---------------------------------------------------------------------

### 4.4 Refund/Reconciliation API

- **Purpose:** To initiate refund or reconciliation for failed or
  disputed transactions.

#### 4.4.1 Request Structure (IMS → BBPS API: Refund/Reconciliation API)

  ---------------------------------------------------------------------
  Field Name      Type           Description
  --------------- -------------- --------------------------------------
  transactionId   String         Unique transaction ID

  reason          String         Reason for refund

  ts              DateTime       Timestamp in ISO format
  ---------------------------------------------------------------------

#### 4.4.2 Response Structure (BBPS → IMS: Refund/Reconciliation API)

  ------------------------------------------------------------------------
  Field Name           Type           Description
  -------------------- -------------- ------------------------------------
  refundId             String         Unique refund ID

  status               String         Refund status (INITIATED /
                                      COMPLETED)

  expectedCompletion   DateTime       Expected completion date

  message              String         Status message
  ------------------------------------------------------------------------

### 4.5 Error Codes

  ------------------------------------------------------
  Error Code        Description
  ----------------- ------------------------------------
  100               Success -- Transaction successful

  200               Invalid Policy Number

  210               Policy details mismatch

  300               Invalid Amount

  310               Amount exceeds allowed limit

  400               Payment declined by bank

  410               Payment timeout

  500               Technical error at BBPS

  510               Service unavailable

  999               Unknown error
  ------------------------------------------------------

## **5. Attachments**

The following documents can be referred.
