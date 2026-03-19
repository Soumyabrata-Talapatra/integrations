> INTERNAL APPROVAL FORM

**Project Name:** e-IA Integration

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

[4.1 Policy Linkage API [6](#policy-linkage-api)](#policy-linkage-api)

[4.2 Policy Publish API [6](#policy-publish-api)](#policy-publish-api)

[4.3 Policy Update Event API
[8](#policy-update-event-api)](#policy-update-event-api)

[4.4 Servicing Update API
[8](#servicing-update-api)](#servicing-update-api)

[4.5 Linkage Status Query API
[9](#linkage-status-query-api)](#linkage-status-query-api)

[4.6 Repository Porting API
[10](#repository-porting-api)](#repository-porting-api)

[4.7 Error & Dead‑Letter (DLQ) API
[11](#error-deadletter-dlq-api)](#error-deadletter-dlq-api)

[4.8 Document Access API
[11](#document-access-api)](#document-access-api)

[4.9 Health/Status API [12](#healthstatus-api)](#healthstatus-api)

[4.10 Authentication (Token) API
[12](#authentication-token-api)](#authentication-token-api)

[4.11 Refund/Reconciliation API (Optional for Payment‑related Sync)
[13](#refundreconciliation-api-optional-for-paymentrelated-sync)](#refundreconciliation-api-optional-for-paymentrelated-sync)

[**5. Attachments** [14](#attachments)](#attachments)

## **1. Executive Summary**

This document outlines the requirements for integrating the Insurance
Management System (IMS) of India Post PLI with the National Insurance
Register (e-IA). An e‑Insurance Account (e‑IA) is a digital repository
introduced under IRDAI to store and manage all insurance policies of an
individual (life, health, general, pension) in electronic form through
authorized repositories. Integration will enable the users to:

- Digitize policy issuance & servicing (paperless).

- Offer a unified customer interface across repositories.

- Ensure secure, compliant, real‑time exchange of policy/KYC/service
  data.

## **2. Project Scope**

This Module will:

- Policy Linkage: Link existing PLI policies to a customer's e‑IA (any
  repository).

- New Issuance: Deliver newly issued PLI policies directly to e‑IA
  (e‑policy).

- Updates & Servicing: Synchronize customer KYC, contact data, nominee,
  bank details, endorsements, premium status, loan/lien, revival,
  surrender, claims.

- Status & Notifications: Provide event updates to repositories and
  receive acknowledgements/errors.

- Porting: Support repository migration (e.g., CAMS→NSDL) with
  continuity.

- Consent & Data Privacy: Manage customer consent, repository
  authorization, audit logs.

- Reporting: Regulatory, operational, and reconciliation reports;
  dashboards.

## **3. Business Requirements**

  ----------------------------------------------------------------------
  **ID**        **Requirements**
  ------------- --------------------------------------------------------
  FS_EIA_001    Enable secure linkage of PLI policies to customer
                e‑Insurance Accounts across all IRDAI‑approved
                repositories with consent validation.

  FS_EIA_002    Publish complete policy data and signed e‑policy
                documents to repositories for digital access and
                compliance.

  FS_EIA_003    Synchronize servicing updates and status changes
                bi‑directionally between IMS and repositories with event
                tracking.

  FS_EIA_004    Ensure end-to-end security, privacy, and IRDAI-compliant
                data handling including encryption, consent, and
                India-only storage.

  FS_EIA_005    Provide operational reports, reconciliation mechanisms,
                and refund workflows for payment and data mismatches.

  FS_EIA_006    Maintain high system reliability, performance, and
                observability with support for retries, DLQ, and outage
                resilience.
  ----------------------------------------------------------------------

## **4. Functional Requirements Specification**

**Use Case Diagram for e-IA Integration:**
![](media/image1.png){width="6.268055555555556in"
height="4.178472222222222in"}

**Flow Chart for e-IA and India Post PLI
Integration:**![](media/image2.png){width="2.6180555555555554in"
height="3.9270833333333335in"}

### 4.1 Policy Linkage API

- **Purpose**: To initiate linkage of a PLI policy with a customer's
  e‑Insurance Account (e‑IA) through an IRDAI‑approved repository.

#### 4.1.1 Request Structure (Repository → IMS: Policy Linkage API)

  ---------------------------------------------------------------------------
  **Field Name**   **Type**   **Description**
  ---------------- ---------- -----------------------------------------------
  eiaNumber        String     Unique e‑Insurance Account number

  repositoryCode   String     Code of the repository (NSDL_NIR / CDSL_IR /
                              CAMSREP / KARVY)

  policyNumber     String     Unique PLI policy number

  customerName     String     Full name of the policyholder

  dob              Date       Date of birth of the policyholder (YYYY‑MM‑DD)

  mobile           String     Registered mobile number

  email            String     Registered email ID

  panMasked        String     Masked PAN (e.g., ABCDE1234X)

  aadhaarMasked    String     Masked Aadhaar (e.g., XXXX‑XXXX‑1234)

  consentToken     String     JWT token indicating customer consent and scope
                              (e.g., linkage)

  requestedAt      DateTime   Timestamp of request in ISO 8601 format
  ---------------------------------------------------------------------------

#### 4.1.2 Response Structure (IMS → Repository: Policy Linkage API)

  ----------------------------------------------------------------------------
  **Field Name**        **Type**   **Description**
  --------------------- ---------- -------------------------------------------
  linkageRequestId      String     Unique ID for the linkage request

  status                String     Request status (VALIDATION_IN_PROGRESS /
                                   LINKED / FAILED)

  estimatedCompletion   DateTime   Estimated completion time (if applicable)

  message               String     Status message or error description
  ----------------------------------------------------------------------------

### 4.2 Policy Publish API

- **Purpose**: To publish the full Policy Data Pack (including e‑policy
  PDF reference) from IMS to the repository after successful validation.

#### 4.2.1 Request Structure (IMS → Repository: Policy Publish API)

  ----------------------------------------------------------------------------
  **Field Name**     **Type**   **Description**
  ------------------ ---------- ----------------------------------------------
  linkageRequestId   String     ID of the linkage request that passed
                                validation

  policyNumber       String     Unique PLI policy number

  productCode        String     PLI product code (e.g., PLI_WHOLE_LIFE)

  sumAssured         Number     Sum assured amount

  issueDate          Date       Policy issue date (YYYY‑MM‑DD)

  termYears          Integer    Policy term in years

  premiumMode        String     Premium mode (MONTHLY / QUARTERLY /
                                HALF_YEARLY / YEARLY)

  status             String     Current policy status (IN_FORCE / LAPSED /
                                REVIVED / SURRENDERED / MATURED)

  premiumSchedule    Array      Array of premium items (dueDate, amount,
                                status)

  riders             Array      Rider details, if any

  loanAmount         Number     Current policy loan amount (0 if none)

  nominees           Array      Nominee set (name, relationship, sharePct,
                                minorGuardian)

  contactMobile      String     Registered mobile

  contactEmail       String     Registered email

  ePolicyPdfUrl      String     Time‑bound signed URL to e‑policy PDF

  soaPdfUrl          String     Time‑bound signed URL to Statement of Account
                                PDF

  documentHashes     Object     SHA‑256 hashes for PDFs (ePolicyPdfSha256,
                                soaPdfSha256)

  payloadSignature   String     JWS signature for payload integrity

  publishedAt        DateTime   Timestamp of publish in ISO 8601 format
  ----------------------------------------------------------------------------

#### 4.2.2 Response Structure (Repository → IMS: Policy Publish API)

  -------------------------------------------------------------------
  **Field Name** **Type**   **Description**
  -------------- ---------- -----------------------------------------
  linkageId      String     Unique linkage ID created by repository

  status         String     Repository status (LINKED / DUPLICATE /
                            FAILED)

  dashboardUrl   String     Repository dashboard URL for the policy

  message        String     Status or error description
  -------------------------------------------------------------------

### 4.3 Policy Update Event API

- **Purpose**: To notify the repository of policy servicing changes
  (status updates, endorsements, payments, claims status, etc.).

#### 4.3.1 Request Structure (IMS → Repository: Policy Update Event API)

  ----------------------------------------------------------------------------
  **Field Name**     **Type**   **Description**
  ------------------ ---------- ----------------------------------------------
  eventId            String     Unique event identifier

  eventType          String     Event type (POLICY_UPDATED / PAYMENT_POSTED /
                                CLAIM_STATUS_CHANGED / ENDORSEMENT_ADDED /
                                NOMINEE_CHANGED / PORTED / ERROR)

  policyNumber       String     Unique PLI policy number

  changes            Object     Delta object describing the change (e.g.,
                                status old/new, fields changed)

  occurredAt         DateTime   Event time in ISO 8601 format

  payloadSignature   String     JWS signature for integrity
  ----------------------------------------------------------------------------

#### 4.3.2 Response Structure (Repository → IMS: Policy Update Event API)

  ------------------------------------------------------------
  **Field    **Type**   **Description**
  Name**                
  ---------- ---------- --------------------------------------
  accepted   Boolean    Whether the event was accepted for
                        processing

  eventId    String     Echoed event ID for correlation

  message    String     Acceptance info or reason for
                        rejection
  ------------------------------------------------------------

### 4.4 Servicing Update API

- **Purpose**: To submit customer‑initiated servicing updates (e.g.,
  nominee change, contact change, bank account updates) from repository
  to IMS.

#### 4.4.1 Request Structure (Repository → IMS: Servicing Update API)

  ----------------------------------------------------------------------------
  **Field Name**   **Type**   **Description**
  ---------------- ---------- ------------------------------------------------
  eiaNumber        String     Unique e‑Insurance Account number

  repositoryCode   String     Repository code (NSDL_NIR / CDSL_IR / CAMSREP /
                              KARVY)

  policyNumber     String     Unique PLI policy number

  updateType       String     Type of update (NOMINEE_CHANGED /
                              CONTACT_CHANGED / BANK_CHANGED /
                              ENDORSEMENT_REQUEST)

  payload          Object     Update payload (e.g., nominees array, new
                              contact details)

  consentToken     String     JWT token confirming customer consent for the
                              specific update

  requestedAt      DateTime   Timestamp in ISO 8601 format
  ----------------------------------------------------------------------------

#### 4.4.2 Response Structure (IMS → Repository: Servicing Update API)

  -------------------------------------------------------------------------------
  **Field Name**       **Type**   **Description**
  -------------------- ---------- -----------------------------------------------
  servicingRequestId   String     Unique servicing request ID

  status               String     Status (QUEUED / IN_PROGRESS / COMPLETED /
                                  REJECTED)

  slaSeconds           Integer    Expected processing SLA in seconds

  message              String     Status message or rejection reason
  -------------------------------------------------------------------------------

### 4.5 Linkage Status Query API

- **Purpose**: To query linkage status for a given policy number.

#### 4.5.1 Request Structure (Repository → IMS: Linkage Status Query API)

  -----------------------------------------------
  **Field Name** **Type**   **Description**
  -------------- ---------- ---------------------
  policyNumber   String     Unique PLI policy
                            number

  -----------------------------------------------

#### 4.5.2 Response Structure (IMS → Repository: Linkage Status Query API)

  ----------------------------------------------------------------------------
  **Field Name**   **Type**   **Description**
  ---------------- ---------- ------------------------------------------------
  policyNumber     String     Unique PLI policy number

  eiaNumber        String     Linked e‑IA number (if any)

  repositoryCode   String     Repository code

  status           String     Linkage status (LINKED / NOT_LINKED / PENDING /
                              FAILED)

  linkedAt         DateTime   Linkage timestamp (if LINKED)

  message          String     Additional info
  ----------------------------------------------------------------------------

### 4.6 Repository Porting API

- **Purpose**: To port policy linkages from one repository to another
  with customer consent.

#### 4.6.1 Request Structure (New Repository → IMS: Repository Porting API)

  ----------------------------------------------------------------------
  **Field Name**       **Type**   **Description**
  -------------------- ---------- --------------------------------------
  fromRepositoryCode   String     Source repository code

  toRepositoryCode     String     Destination repository code

  eiaNumber            String     e‑IA number to be ported

  policyNumbers        Array      Array of policy numbers to port

  consentToken         String     JWT token indicating explicit consent
                                  for porting

  requestedAt          DateTime   Timestamp in ISO 8601 format
  ----------------------------------------------------------------------

#### 4.6.2 Response Structure (IMS → New Repository: Repository Porting API)

  ---------------------------------------------------------------------------
  **Field Name**     **Type**   **Description**
  ------------------ ---------- ---------------------------------------------
  portingRequestId   String     Unique porting request ID

  status             String     Status (INITIATED / COMPLETED / PARTIAL /
                                FAILED)

  portedPolicies     Array      Array of objects (policyNumber, status)

  message            String     Status or error description
  ---------------------------------------------------------------------------

### 4.7 Error & Dead‑Letter (DLQ) API

- **Purpose**: To record and fetch unrecoverable messages/events for
  manual reconciliation.

#### 4.7.1 Request Structure (IMS → Repository: DLQ Submit API)

  ---------------------------------------------------------------------------
  **Field Name** **Type**   **Description**
  -------------- ---------- -------------------------------------------------
  dlqId          String     Unique DLQ record ID

  source         String     Source component (EVENT_BUS / POLICY_PUBLISH /
                            SERVICING_UPDATE)

  policyNumber   String     Policy number involved (if any)

  payload        Object     Original message payload (sanitized)

  errorCode      String     Error code (SCHEMA_INVALID / AUTH_FAILURE /
                            CONSENT_EXPIRED / REPO_DOWN / TIMEOUT)

  errorMessage   String     Error description

  failedAt       DateTime   Timestamp in ISO 8601 format
  ---------------------------------------------------------------------------

#### 4.7.2 Response Structure (Repository → IMS: DLQ Submit API)

  ----------------------------------------------
  **Field    **Type**   **Description**
  Name**                
  ---------- ---------- ------------------------
  accepted   Boolean    Whether DLQ record is
                        stored

  dlqId      String     Echoed DLQ ID

  message    String     Storage status info
  ----------------------------------------------

### 4.8 Document Access API

- **Purpose**: To fetch time‑bound, signed URLs to e‑policy and
  statement documents.

#### 4.8.1 Request Structure (Repository → IMS: Document Access API)

  -------------------------------------------------------
  **Field Name** **Type**   **Description**
  -------------- ---------- -----------------------------
  policyNumber   String     Unique PLI policy number

  documentType   String     Type (E_POLICY_PDF / SOA_PDF)

  requestedAt    DateTime   Timestamp in ISO 8601 format
  -------------------------------------------------------

#### 4.8.2 Response Structure (IMS → Repository: Document Access API)

  ------------------------------------------------------------
  **Field     **Type**   **Description**
  Name**                 
  ----------- ---------- -------------------------------------
  url         String     Time‑bound signed URL (validity ≤ 15
                         minutes)

  sha256      String     SHA‑256 hash of the document

  expiresAt   DateTime   URL expiry timestamp

  message     String     Status or error description
  ------------------------------------------------------------

### 4.9 Health/Status API

- **Purpose**: To check the readiness and health of the integration
  endpoints.

#### 4.9.1 Request Structure (Repository → IMS: Health/Status API)

  ------------------------------------------------------
  **Field Name**  **Type**   **Description**
  --------------- ---------- ---------------------------
  correlationId   String     Optional correlation ID for
                             tracing

  ------------------------------------------------------

#### 4.9.2 Response Structure (IMS → Repository: Health/Status API)

  --------------------------------------------------------------------------
  **Field      **Type**   **Description**
  Name**                  
  ------------ ---------- --------------------------------------------------
  status       String     Overall status (UP / DEGRADED / DOWN)

  components   Array      Component status array (API_GATEWAY, DOC_SERVICE,
                          EVENT_BUS)

  timestamp    DateTime   ISO 8601 timestamp

  message      String     Diagnostic info
  --------------------------------------------------------------------------

### 4.10 Authentication (Token) API

- **Purpose**: To obtain OAuth2 access tokens for secured API calls.

#### 4.10.1 Request Structure (Repository → IMS: Token API)

  ----------------------------------------------------------------
  **Field Name** **Type**   **Description**
  -------------- ---------- --------------------------------------
  grantType      String     Must be client_credentials

  clientId       String     Repository client ID

  clientSecret   String     Repository client secret

  scope          String     Requested scopes (e.g., linkage
                            publish events)
  ----------------------------------------------------------------

#### 4.10.2 Response Structure (IMS → Repository: Token API)

  ---------------------------------------------
  **Field       **Type**   **Description**
  Name**                   
  ------------- ---------- --------------------
  accessToken   String     Bearer JWT token

  tokenType     String     Constant Bearer

  expiresIn     Integer    Token expiry in
                           seconds

  scope         String     Granted scopes
  ---------------------------------------------

### 4.11 Refund/Reconciliation API (Optional for Payment‑related Sync)

- **Purpose**: To reconcile payment events or initiate corrections
  (e.g., posted premium mismatches seen in e‑IA vs IMS).

#### 4.11.1 Request Structure (Repository → IMS: Refund/Reconciliation API)

  -----------------------------------------------------------------------------
  **Field Name**  **Type**   **Description**
  --------------- ---------- --------------------------------------------------
  transactionId   String     Unique transaction ID (from repository or IMS)

  reason          String     Reason for reconciliation (POSTING_MISMATCH /
                             DUPLICATE / REVERSAL)

  ts              DateTime   Timestamp in ISO 8601 format
  -----------------------------------------------------------------------------

#### 4.11.2 Response Structure (IMS → Repository: Refund/Reconciliation API)

  -----------------------------------------------------------------------
  **Field Name**       **Type**   **Description**
  -------------------- ---------- ---------------------------------------
  reconId              String     Unique reconciliation ID

  status               String     Status (INITIATED / COMPLETED /
                                  REJECTED)

  expectedCompletion   DateTime   Expected completion date/time

  message              String     Status message
  -----------------------------------------------------------------------

## **5. Attachments**

The following documents can be referred.
