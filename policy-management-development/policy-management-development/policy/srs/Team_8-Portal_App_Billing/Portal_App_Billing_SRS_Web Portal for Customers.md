> INTERNAL APPROVAL FORM

**Project Name:** Web Portal for Customers

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

[4.1 User Onboarding & Login Authentication
[7](#user-onboarding-login-authentication)](#user-onboarding-login-authentication)

[4.1.1 User Onboarding [8](#user-onboarding)](#user-onboarding)

[4.1.2 Login Authentication
[8](#login-authentication)](#login-authentication)

[4.2 Dashboard Page [8](#dashboard-page)](#dashboard-page)

[4.3 My Policies Page [9](#my-policies-page)](#my-policies-page)

[4.5 Payment Page [11](#payment-page)](#payment-page)

[4.6 Service Request Page
[13](#service-request-page)](#service-request-page)

[4.7 Service Request Tracking Page
[13](#service-request-tracking-page)](#service-request-tracking-page)

[4.8 Notification & Support Page
[15](#notification-support-page)](#notification-support-page)

[4.9 Policy Purchase Page
[17](#policy-purchase-page)](#policy-purchase-page)

[4.10 Tools & Utilities Page
[17](#tools-utilities-page)](#tools-utilities-page)

[4.11 My Profile Page [17](#my-profile-page)](#my-profile-page)

[**5. Appendices** [39](#appendices)](#appendices)

## **1. Executive Summary**

The purpose of this document is to define the requirements for the India
Post PLI Customer Portal. This portal will enable customers to manage
their Postal Life Insurance (PLI) and Rural Postal Life Insurance (RPLI)
policies online, providing a seamless, secure, and user-friendly
experience.

## **2. Project Scope**

The portal will allow customers to:

- Onboard and authenticate securely.

- View and manage policies.

- Make premium payments.

- Initiate and track claims.

- Request non-financial services.

- Purchase new policies.

- Modify existing policy details.

- Access notifications and support.

## **3. Business Requirements**

  -----------------------------------------------------------------------
  Requirement   Business          Description
  ID            Requirement       
  ------------- ----------------- ---------------------------------------
  FS_WPC_001    Customer          The portal shall allow new and existing
                Onboarding        customers to register using policy
                                  details and mobile OTP.

  FS\_ WPC_002  Customer          The system shall validate customer
                Onboarding        identity using Aadhaar or other
                                  government-approved KYC methods.

  FS\_ WPC_003  Secure Login &    The portal shall provide secure login
                Authentication    with username/password and OTP-based
                                  multi-factor authentication.

  FS\_ WPC_004  Secure Login &    The system shall support password reset
                Authentication    and account recovery mechanisms.

  FS\_ WPC_005  Customer          The portal shall display a personalized
                Dashboard         dashboard showing active policies,
                                  premium due dates, claim status, and
                                  notifications.

  FS\_ WPC_006  Customer          The dashboard shall provide quick
                Dashboard         access to key actions like payments and
                                  service requests.

  FS\_ WPC_007  Premium Payment   The portal shall allow customers to pay
                                  premiums online using multiple payment
                                  modes (UPI, Net Banking, Cards).

  FS\_ WPC_008  Premium Payment   The system shall generate and store
                                  digital receipts for each transaction.

  FS\_ WPC_009  Claim Initiation  The portal shall allow customers to
                                  initiate claims (maturity, death,
                                  surrender) by submitting required forms
                                  and documents.

  FS\_ WPC_010  Claim Initiation  The system shall validate claim
                                  eligibility and notify users of next
                                  steps.

  FS\_ WPC_011  Claim Tracking    The portal shall provide real-time
                                  tracking of claim status.

  FS\_ WPC_012  Claim Tracking    Customers shall receive automated
                                  updates via SMS/email at each stage of
                                  claim processing.

  FS\_ WPC_013  Notifications and The portal shall send timely
                Support           notifications for premium due dates,
                                  claim updates, and service request
                                  status.

  FS\_ WPC_014  Notifications and The system shall include a support
                Support           module with chatbot, FAQs, and
                                  ticket-based query resolution.

  FS\_ WPC_015  New Policy        The portal shall allow customers to
                Purchase          explore and initiate purchase of
                Initiation        PLI/RPLI policies.

  FS\_ WPC_016  New Policy        The system shall generate premium
                Purchase          quotes based on user inputs and policy
                Initiation        type.

  FS\_ WPC_017  Policy Purchase   The portal shall facilitate scheduling
                Finalization      of medical tests (if required) and
                                  submission of final documents.

  FS\_ WPC_018  Policy Purchase   The system shall issue policy documents
                Finalization      upon successful verification and
                                  approval.

  FS\_ WPC_019  Non-Financial     The portal shall allow submission of
                Service Requests  service requests that require physical
                -- Type A         verification at a post office.

  FS\_ WPC_020  Non-Financial     The system shall track and update the
                Service Requests  status of such requests.
                -- Type A         

  FS\_ WPC_021  Non-Financial     The portal shall allow submission of
                Service Requests  service requests authenticated via OTP
                -- Type B         (e.g., mobile/email update).

  FS\_ WPC_022  Non-Financial     The system shall process and confirm
                Service Requests  such requests instantly.
                -- Type B         

  FS\_ WPC_023  Policy            The portal shall allow customers to
                Commutation       request changes to premium amount,
                                  premium term, or sum assured.

  FS\_ WPC_024  Policy            The system shall validate requests
                Commutation       against underwriting rules and notify
                                  customers of approval or rejection.

  FS\_ WPC_025  Policy Conversion The portal shall allow eligible
                                  customers to convert their policy type.

  FS\_ WPC_026  Policy Conversion The system shall guide users through
                                  the conversion process and update
                                  policy records accordingly.

  FS\_ WPC_027  Session &         The portal shall implement secure
                Security          session management with automatic
                Management        logout on inactivity.

  FS\_ WPC_028  Session &         The system shall encrypt sensitive data
                Security          and maintain an audit trail of all user
                Management        actions.
  -----------------------------------------------------------------------

## **4. Functional Requirements Specification**

**Flow for the Web Portal for Customers:**

- **Primary Actions before Login:**

  - User Onboarding (User icon)

  - Login Authentication (Lock icon)

- **Dashboard:** Display to Customer After Login

- Main Features in Dashboard:

  - **My Policies Page**

  - **Payment Page**

    i.  Premium Payment

    ii. Repay Full Loan

    iii. Repay Partial Loan

    iv. Payment History

    v.  Income Tax Certificate

  - **Service Requests Page**

    i.  Claim Initiation

    ii. Claim Tracking

    iii. Non-Financial Service Requests Type A (Post Office
         Verification)

    iv. Non-Financial Service Requests Type B (OTP Authentication)

    v.  Policy Commutation (Premium Amount, Term, Sum Assured)

    vi. Policy Conversion

  - **Service Requests Tracking Page**

  - **Notification & Support Page**

  - **Policy Purchase Page**

    i.  Illustration

    ii. Quote

    iii. New Policy Purchase Initiation

    iv. Policy Purchase Medical & Finalization

    v.  Proposal Tracking

    vi. Initial Premium Payment

    vii. Upload Documents

  - **Tools & Utilities Page**

    i.  Request an Agent

    ii. Locate an Agent

    iii. Download Forms and Documents

    iv. Guidelines for Customers

    v.  Customer Grievance

    vi. Escalation Matrix

    vii. Submit Query

    viii. Service Request Documents

    ix. Privacy Policy

  - **My Profile Page**

    i.  Change Password

## 4.1 User Onboarding & Login Authentication

![A diagram of a company AI-generated content may be
incorrect.](media/image1.png){width="6.0625in"
height="9.013823272090988in"}

### 4.1.1 User Onboarding

- **Purpose:** Allow new and existing customers to register.

- **Fields:**

  - Policy Number: Customer's existing policy number for validation.

  - Name: Full name as per policy records.

  - Date of Birth: For identity verification.

  - Mobile Number: Used for OTP verification.

  - Email ID: For communication and notifications.

  - Aadhaar / KYC Details: Government-approved identity verification.

<!-- -->

- **Details:** Integration with KYC APIs for validation.

### 4.1.2 Login Authentication

- **Purpose:** Authenticate customers securely before accessing the
  portal.

- **Fields:**

  - Username / Policy Number: Unique identifier for the customer.

  - Password: Secret key for authentication.

  - OTP: One-time password for multi-factor authentication.

- **Details:** Includes "Forgot Password" and "Register" links.

## 4.2 Dashboard Page

- **Purpose:** This is the page that gets displayed to user after
  successful login.

- **Fields:**

  - My Policies (Document icon)

  - Payment (₹ icon)

  - Service Requests (Gear icon)

  - Notification & Support (Bell icon)

  - Service Requests Tracking (Checklist icon)

  - Policy Purchase (Document icon)

  - Tools & Utilities (Tools icon)

  - My Profile (User icon)

**Wireframe for the Dashboard Page:**

![](media/image2.png){width="4.823588145231846in"
height="3.215546806649169in"}

## 4.3 My Policies Page

- **Purpose:** Show details of all active policies.

- **Fields:**

<!-- -->

- Policy Number: Unique identifier for each policy.

- Policy Type: Type of policy (PLI/RPLI).

- Sum Assured: Coverage amount.

- Premium Amount: Regular payment amount.

- Next Due Date: Upcoming premium payment date.

<!-- -->

- **Details:** Option to download policy documents.

![](media/image3.emf){width="5.177083333333333in"
height="10.183444881889764in"}

## 4.5 Payment Page

- **Purpose:** Enable premium and loan repayments.

- **Fields:**

  - Policy Number: Policy for which payment is being made.

  - Amount Due: Total payable amount.

  - Payment Mode: UPI, Net Banking, Credit/Debit Card.

<!-- -->

- **Details:** Generates digital receipt and updates payment history.

![](media/image4.png){width="5.208333333333333in"
height="9.929835958005249in"}

## 4.6 Service Request Page

- **Purpose:** Submit financial and non-financial requests.

- **Fields:**

  - Request Type: Claim, Policy Commutation, Conversion, etc.

  - Policy Number: Associated policy for the request.

  - Supporting Documents Upload: Required documents for processing.

<!-- -->

- **Details:** OTP authentication for Type B requests.

## 4.7 Service Request Tracking Page

- **Purpose:** Track status of submitted requests.

- **Fields:**

  - Request ID: Unique identifier for each request.

  - Status: Pending, Approved, Rejected.

  - Last Updated Date: Timestamp of last update.

<!-- -->

- **Details:** Real-time updates and SMS/email notifications.

![](media/image5.emf){width="3.5104166666666665in"
height="10.263687664041996in"}

## 4.8 Notification & Support Page

- **Purpose:** Provide alerts and customer assistance.

- **Fields:**

  - Notifications List: Premium due, claim updates, service status.

  - Chatbot / FAQ: Automated help and common queries.

  - Raise Ticket: Submit detailed query for resolution.

<!-- -->

- **Details:** Includes escalation matrix and grievance redressal.

![](media/image6.emf){width="4.416666666666667in"
height="10.200169510061242in"}

## 4.9 Policy Purchase Page

- **Purpose:** Explore and buy new policies.

- **Fields:**

  - Policy Type: Select PLI/RPLI product.

  - Age, Sum Assured, Term: Inputs for premium calculation.

  - Quote & Illustration: Generated based on inputs.

<!-- -->

- **Details:** Upload documents, schedule medical tests, track proposal.

## 4.10 Tools & Utilities Page

- **Purpose:** Provide additional resources.

- **Fields:**

  - Request Agent: Submit request for agent assistance.

  - Locate Agent: Find nearest agent by PIN code.

  - Download Forms: Access policy-related forms.

  - Submit Query: Raise general queries.

<!-- -->

- **Details:** Includes Privacy Policy and Customer Guidelines.

## 4.11 My Profile Page

- **Purpose:** Manage personal details.

- **Fields:**

  - Name: Editable if allowed.

  - Mobile Number: Update with OTP verification.

  - Email: Update for communication.

  - Change Password: Secure password update.

<!-- -->

- **Details:** All changes logged for audit.

**Part 1: New PLI Policy Purchase Flow - Application & Eligibility**

**NEW PLI POLICY PURCHASE FLOW - PART 1/2: APPLICATION & ELIGIBILITY**

![Untitled
diagram-2025-10-30-084228.png](media/image7.png){width="6.315510717410324in"
height="8.847125984251969in"}

**Part 2: New PLI Policy Purchase Flow - Medical & Finalization**

**NEW PLI POLICY PURCHASE FLOW - PART 2/2: MEDICAL & FINALIZATION**

![Untitled
diagram-2025-10-24-141023.png](media/image8.png){width="6.971442475940507in"
height="8.487350174978127in"}

**Medical Requirement Decision Matrix (For Reference)**

**PLI MEDICAL REQUIREMENT MATRIX**

![Untitled
diagram-2025-10-24-141205.png](media/image9.png){width="6.847951662292213in"
height="3.0888943569553806in"}

**\**

**Part 1: New RPLI Policy Purchase Flow - Application & Rural
Validation**

**NEW RPLI POLICY PURCHASE FLOW - PART 1/2: APPLICATION & RURAL
VALIDATION**![Untitled
diagram-2025-10-24-141451.png](media/image10.png){width="6.92137467191601in"
height="6.11613845144357in"}

**\**

**Part 2: New RPLI Policy Purchase Flow - Age Proof & Medical
Requirements**

**NEW RPLI POLICY PURCHASE FLOW - PART 2/2: AGE PROOF & MEDICAL
REQUIREMENTS**

![Untitled
diagram-2025-10-24-141627.png](media/image11.png){width="6.9207130358705164in"
height="7.186797900262468in"}

**\**

PLI Medical Requirement Decision Matrix (For Reference)

**RPLI MEDICAL REQUIREMENT MATRIX**

![Untitled
diagram-2025-10-24-141828.png](media/image12.png){width="6.918163823272091in"
height="7.257709973753281in"}

**1. Non-Financial Service Requests - Type A (Post Office Verification
Required)**

![Untitled
diagram-2025-10-24-103026.png](media/image13.png){width="7.009011373578303in"
height="7.437226596675416in"}

2\. Non-Financial Service Requests - Type B (OTP Authentication Only)

![Untitled
diagram-2025-10-24-103059.png](media/image14.png){width="6.9044991251093615in"
height="8.713514873140857in"}

**Part 1: Claims Initiation & Eligibility Check**

**CLAIMS FLOW - PART 1: INITIATION & ELIGIBILITY**

![Untitled
diagram-2025-10-24-115439.png](media/image15.png){width="6.894143700787402in"
height="6.876896325459318in"}

Part 2: Bank Details Verification & Document Upload

**CLAIMS FLOW - PART 2: BANK VERIFICATION & DOCUMENTS**

![Untitled
diagram-2025-10-24-115655.png](media/image16.png){width="6.586023622047244in"
height="6.22739501312336in"}

Part 3: Notifications & Tracking

**CLAIMS FLOW - PART 3: NOTIFICATIONS & TRACKING**

![Untitled
diagram-2025-10-24-120645.png](media/image17.png){width="6.845699912510936in"
height="6.436860236220473in"}

**Part 1: PLI Policy Commutation - Contract Alteration Eligibility**

PLI POLICY COMMUTATION FLOW - PART 1/4: CONTRACT ALTERATION ELIGIBILITY

![A diagram of a work flow AI-generated content may be
incorrect.](media/image18.png){width="6.515208880139983in"
height="8.867924321959755in"}

Part 2: PLI Policy Commutation - Alteration Type Processing

PLI POLICY COMMUTATION FLOW - PART 2/4: ALTERATION TYPE PROCESSING

![A diagram of a flowchart AI-generated content may be
incorrect.](media/image19.png){width="6.462263779527559in"
height="8.75829615048119in"}

Part 3: PLI Policy Commutation - Financial Calculations & Validation

PLI POLICY COMMUTATION FLOW - PART 3/4: FINANCIAL CALCULATIONS

![A diagram of a flowchart AI-generated content may be
incorrect.](media/image20.png){width="6.216981627296588in"
height="8.840893482064741in"}

Part 4: PLI Policy Commutation - Approval & Contract Update

PLI POLICY COMMUTATION FLOW - PART 4/4: APPROVAL & UPDATES

![A diagram of a flowchart AI-generated content may be
incorrect.](media/image21.png){width="5.179245406824147in"
height="8.762497812773404in"}

PLI Commutation Rules Summary (For Reference)

COMMUTATION TYPES MATRIX

![A diagram of a process AI-generated content may be
incorrect.](media/image22.png){width="6.840820209973753in"
height="2.188561898512686in"}

**Part 1: PLI Policy Conversion - Eligibility & Type Selection**

PLI POLICY CONVERSION FLOW - PART 1/4: ELIGIBILITY & TYPE SELECTION

![A diagram of a diagram AI-generated content may be
incorrect.](media/image23.png){width="6.405660542432196in"
height="8.893330052493438in"}

Part 2: PLI Policy Conversion - Maturity Date & Medical Check

PLI POLICY CONVERSION FLOW - PART 2/4: MATURITY & MEDICAL REQUIREMENTS

![A diagram of a flowchart AI-generated content may be
incorrect.](media/image24.png){width="6.902296587926509in"
height="7.886792432195976in"}

Part 3: PLI Policy Conversion - Fee Calculation & Bonus Treatment

PLI POLICY CONVERSION FLOW - PART 3/4: FEE & BONUS CALCULATIONS

![A diagram of a company AI-generated content may be
incorrect.](media/image25.png){width="4.849056211723535in"
height="8.730300743657043in"}

Part 4: PLI Policy Conversion - Approval & Policy Update

PLI POLICY CONVERSION FLOW - PART 4/4: APPROVAL & IMPLEMENTATION

![A diagram of a company AI-generated content may be
incorrect.](media/image26.png){width="5.650943788276465in"
height="8.773090551181102in"}

PLI Conversion Rules Summary (For Reference)

CONVERSION TYPES MATRIX

![A diagram of a company AI-generated content may be
incorrect.](media/image27.png){width="6.895376202974628in"
height="1.4004385389326335in"}

**Security & Session Management Flow chart:**

![](media/image28.png){width="6.268055555555556in"
height="6.106621828521435in"}

## **5. Appendices**

The Following Documents attached below can be used.

![](media/image29.emf)
