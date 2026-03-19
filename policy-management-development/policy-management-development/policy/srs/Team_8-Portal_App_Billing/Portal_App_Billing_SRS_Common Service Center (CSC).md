> INTERNAL APPROVAL FORM

**Project Name:** Common Service Center (CSC)

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

[4.1 Customer Onboarding API
[5](#customer-onboarding-api)](#customer-onboarding-api)

[4.2 Payment Initiation API
[6](#payment-initiation-api)](#payment-initiation-api)

[4.3 Policy Status API [7](#policy-status-api)](#policy-status-api)

[4.4 Grievance Registration API
[7](#grievance-registration-api)](#grievance-registration-api)

[4.5 Policy Services API
[8](#policy-services-api)](#policy-services-api)

[4.6 Service Request Indexing API
[9](#service-request-indexing-api)](#service-request-indexing-api)

[**5. Attachments** [9](#attachments)](#attachments)

## **1. Executive Summary**

The purpose of this document is to define the requirements for
integrating India Post's PLI Insurance Management System (IMS) with the
Common Service Centre (CSC) platform. This integration aims to enable
customers to access PLI services digitally through CSCs, improving
outreach, reducing dependency on physical post offices, and enhancing
service efficiency.

## **2. Project Scope**

The integration will allow CSC operators to perform:

- Customer Onboarding: New policy purchase, KYC validation, e-mandate
  registration.

- Policy Servicing: Premium payments, policy status checks,
  maturity/claim requests.

- Grievance Management: Complaint registration and resolution via CSC
  portal integrated with PLI CRM.

- Payment Integration: Premium collection through CSC payment gateway.

## **3. Business Requirements**

  ------------------------------------------------------------------------------
  **ID**       **Functionality**   **Requirements**
  ------------ ------------------- ---------------------------------------------
  FS_CSC_001   Customer Onboarding Enable CSC operators to onboard new customers
                                   for PLI policies, including KYC validation
                                   and e-mandate registration.

  FS_CSC_002   Policy Issue        Allow CSC portal to initiate new policy
                                   proposals and send data to IMS for approval
                                   and issuance.

  FS_CSC_003   Premium Collection  Integrate CSC payment gateway with IMS for
                                   real-time premium payment updates and
                                   receipts.

  FS_CSC_004   Policy Servicing    Provide CSC operators access to financial
                                   (premium payment, loan requests) and
                                   non-financial (address change, nominee
                                   update) service request indexing.

  FS_CSC_005   Policy Status Check Enable customers to check policy details and
                                   status through CSC portal using IMS APIs.

  FS_CSC_006   Grievance           Allow CSC portal to register complaints and
               Management          sync with PLI CRM for resolution tracking.

  FS_CSC_007   Monitoring          Provide dashboards for CSC transactions, API
               Dashboard           health, and grievance resolution metrics.
  ------------------------------------------------------------------------------

## **4. Functional Requirements Specification**

**Visual Flow Diagram for Common Service Center (CSC) Integration:**
![](media/image1.png){width="4.064850174978128in"
height="6.097274715660542in"}

### 4.1 Customer Onboarding API

- **Purpose:** To onboard a new customer by capturing personal details,
  validating KYC, and initiating policy proposal.

#### 4.1.1 Request Structure (CSC → IMS: Customer Onboarding API)

  ---------------------------------------------------------
  Field Name        Type       Description
  ----------------- ---------- ----------------------------
  customerName      String     Full name of the customer

  aadhaarNumber     String     Aadhaar number for KYC
                               validation

  mobileNumber      String     Customer's mobile number

  email             String     Customer's email address

  address           String     Residential address

  policyType        String     Type of policy (PLI / RPLI)

  sumAssured        Decimal    Proposed sum assured

  eMandateConsent   Boolean    Consent for e-mandate
                               registration

  ts                DateTime   Timestamp in ISO format
  ---------------------------------------------------------

#### 4.1.2 Response Structure (IMS → CSC: Customer Onboarding API)

  -------------------------------------------------------------
  Field Name       Type       Description
  ---------------- ---------- ---------------------------------
  proposalNumber   String     Unique proposal number

  status           String     Onboarding status (SUCCESS /
                              FAILED)

  message          String     Status message

  ts               DateTime   Response timestamp
  -------------------------------------------------------------

### 4.2 Payment Initiation API

- **Purpose:** To initiate the payment process by sending payment
  details to BBPS and receiving a payment link or status.

#### 4.2.1 Request Structure (IMS → BBPS API: Payment Initiation API)

  ------------------------------------------------
  Field Name     Type       Description
  -------------- ---------- ----------------------
  policyNumber   String     Customer's policy
                            number

  amount         Decimal    Payment amount

  paymentType    String     FULL / ADVANCE

  txn            String     Unique transaction ID

  ts             DateTime   Timestamp in ISO
                            format

  bankPartner    String     Banking partner (e.g.,
                            SBI)

  consent        Boolean    Customer consent flag
  ------------------------------------------------

#### 4.2.2 Response Structure (BBPS → IMS: Payment Initiation API)

  ----------------------------------------------------------------
  Field Name   Type       Description
  ------------ ---------- ----------------------------------------
  txn          String     Same transaction ID

  status       String     Payment status (PENDING / SUCCESS /
                          FAILED)

  paymentUrl   String     URL for payment completion

  message      String     Status message

  ts           DateTime   Response timestamp
  ----------------------------------------------------------------

### 4.3 Policy Status API

- **Purpose:** To fetch the current status of a policy.

#### 4.3.1 Request Structure (CSC → IMS: Policy Status API)

  -----------------------------------------------
  Field Name     Type       Description
  -------------- ---------- ---------------------
  policyNumber   String     Customer's policy
                            number

  ts             DateTime   Timestamp in ISO
                            format
  -----------------------------------------------

#### 4.3.2 Response Structure (IMS → CSC: Policy Status API)

  --------------------------------------------------------------
  Field Name        Type       Description
  ----------------- ---------- ---------------------------------
  policyNumber      String     Policy number

  status            String     Current policy status (ACTIVE /
                               LAPSED)

  nextPremiumDate   Date       Next premium due date

  sumAssured        Decimal    Sum assured

  ts                DateTime   Response timestamp
  --------------------------------------------------------------

### 4.4 Grievance Registration API

- **Purpose:** To register a grievance and sync with PLI CRM.

#### 4.4.1 Request Structure (CSC → IMS: Grievance Registration API)

  ---------------------------------------------------
  Field Name      Type       Description
  --------------- ---------- ------------------------
  customerId      String     Unique customer ID

  complaintType   String     Type of grievance

  description     String     Detailed complaint
                             description

  ts              DateTime   Timestamp in ISO format
  ---------------------------------------------------

#### 4.4.2 Response Structure (IMS → CSC: Grievance Registration API)

  -----------------------------------------------------------------
  Field Name           Type       Description
  -------------------- ---------- ---------------------------------
  grievanceId          String     Unique grievance ID

  status               String     Registration status (SUCCESS /
                                  FAILED)

  expectedResolution   String     Expected resolution time

  ts                   DateTime   Response timestamp
  -----------------------------------------------------------------

### 4.5 Policy Services API

- **Purpose:** To handle financial and non-financial service requests
  related to an existing policy (e.g., address change, nominee update,
  loan request).

#### 4.5.1 Request Structure (CSC → IMS: Policy Services API)

  ---------------------------------------------------------------------------
  Field Name     Type       Description
  -------------- ---------- -------------------------------------------------
  policyNumber   String     Customer's policy number

  serviceType    String     Type of service (ADDRESS_CHANGE / NOMINEE_UPDATE
                            / LOAN_REQUEST)

  details        JSON       Service-specific details (e.g., new address,
                            nominee info)

  customerId     String     Unique customer ID

  ts             DateTime   Timestamp in ISO format
  ---------------------------------------------------------------------------

#### 4.5.2 Response Structure (IMS → CSC: Policy Services API)

  ---------------------------------------------------------------
  Field Name  Type       Description
  ----------- ---------- ----------------------------------------
  requestId   String     Unique service request ID

  status      String     Request status (PENDING / SUCCESS /
                         FAILED)

  message     String     Status message

  ts          DateTime   Response timestamp
  ---------------------------------------------------------------

### 4.6 Service Request Indexing API

- **Purpose:** To log and index all service requests initiated via CSC
  for tracking and audit purposes.

#### 4.6.1 Request Structure (CSC → IMS: Service Request Indexing API)

  ---------------------------------------------------------------------------
  Field Name     Type       Description
  -------------- ---------- -------------------------------------------------
  requestId      String     Unique service request ID

  policyNumber   String     Policy number associated with the request

  serviceType    String     Type of service (ONBOARDING / PAYMENT /
                            POLICY_SERVICE / GRIEVANCE)

  status         String     Current status of the request

  operatorId     String     CSC operator ID

  ts             DateTime   Timestamp in ISO format
  ---------------------------------------------------------------------------

#### 4.6.2 Response Structure (IMS → CSC: Service Request Indexing API)

  -----------------------------------------------------
  Field Name  Type       Description
  ----------- ---------- ------------------------------
  requestId   String     Same request ID

  status      String     Indexing status (SUCCESS /
                         FAILED)

  message     String     Status message

  ts          DateTime   Response timestamp
  -----------------------------------------------------

## **5. Attachments**

The following documents can be referred.
