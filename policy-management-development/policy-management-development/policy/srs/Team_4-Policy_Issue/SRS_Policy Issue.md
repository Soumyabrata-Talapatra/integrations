> INTERNAL APPROVAL FORM

**Project Name: Policy Issue**

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

[3.1 Client Creation [4](#client-creation)](#client-creation)

[3.2 Proposal Management
[6](#proposal-management)](#proposal-management)

[3.3 Policy Issue [7](#policy-issue)](#policy-issue)

[3.4 Document Scanning & Management
[9](#document-scanning-management)](#document-scanning-management)

[3.5 Approval Workflow [11](#approval-workflow)](#approval-workflow)

[3.6 Receipt and Printing Management
[11](#receipt-and-printing-management)](#receipt-and-printing-management)

[3.7 Policy Bond [11](#policy-bond)](#policy-bond)

[**4. Functional Requirements/ Business Rules**
[12](#functional-requirements-business-rules)](#functional-requirements-business-rules)

[1.1 Policy Issue- Without Aadhar Page
[15](#policy-issue--without-aadhar-page)](#policy-issue--without-aadhar-page)

[1.2 Policy Issue- With Aadhar Page
[30](#policy-issue--with-aadhar-page)](#policy-issue--with-aadhar-page)

[1.3 Proposal Creation- Using File Upload
[36](#proposal-creation--using-file-upload)](#proposal-creation--using-file-upload)

[**5. Wireframe** [38](#wireframe)](#wireframe)

[5.1 Policy Issue- without Aadhar Page
[38](#policy-issue--without-aadhar-page-1)](#policy-issue--without-aadhar-page-1)

[5.2 Policy Issue- with Aadhar Page
[40](#policy-issue--with-aadhar-page-1)](#policy-issue--with-aadhar-page-1)

[5.3 Proposal Creation- Using File Upload
[44](#proposal-creation--using-file-upload-1)](#proposal-creation--using-file-upload-1)

[**5. Test Case** [45](#test-case)](#test-case)

[**6. Appendices** [47](#appendices)](#appendices)

## **1. Executive Summary**

To define the requirements for implementing the New Business Policy
Issue process in the Insurance Management System (IMS) for India Post
PLI. This process will enable seamless onboarding of new policyholders,
ensuring compliance, automation, and integration with existing systems.

## **2. Project Scope**

This module will support:

- Client creation

- Proposal management

- Document scanning and management

- Policy issue process (manual, digital via Aadhar, and bulk upload)

- Multi-level approvals (Data Entry at CPC, Quality Reviewer, Approver)

- Receipt and printing management

- Policy bond issuance

## **3. Business Requirements**

### 3.1 Client Creation

+-------------+--------------------------------------------------------+
| Requirement | Requirement                                            |
| ID          |                                                        |
+:============+:=======================================================+
| FS_NB_001   | The system should capture the basic details during     |
|             | customer onboarding through direct workflow-based      |
|             | integration with the sourcing applications such as     |
|             | Customer Portal, mobility solution, and partner        |
|             | channels (including but not limited to).               |
+-------------+--------------------------------------------------------+
| FS_NB_002   | The system should be able to create a unique customer  |
|             | ID applicable to all types of policies in DoP. There   |
|             | should be validation to avoid the creation of          |
|             | duplicates. Customer IDs should be created during      |
|             | policy issuance workflow.                              |
+-------------+--------------------------------------------------------+
| FS_NB_003   | The system should always create a client against any   |
|             | association including but not limited to customers,    |
|             | agents, products, and any other third parties. Some of |
|             | the fields used for creating a client (including but   |
|             | not limited to)                                        |
|             |                                                        |
|             | a\. First, Middle & Last Name                          |
|             |                                                        |
|             | b\. Date of Birth                                      |
|             |                                                        |
|             | c\. Gender                                             |
|             |                                                        |
|             | d\. Email ID                                           |
|             |                                                        |
|             | e\. Phone Number                                       |
|             |                                                        |
|             | f\. Address                                            |
|             |                                                        |
|             | g\. Aadhar Number (Optional)                           |
|             |                                                        |
|             | h\. PAN Number (Mandatory if premium is above 50,000)  |
|             |                                                        |
|             | i\. Nationality                                        |
|             |                                                        |
|             | j\. Country of Residence                               |
|             |                                                        |
|             | k\. Any other fields in the future                     |
|             |                                                        |
|             | l\. eIA ID                                             |
|             |                                                        |
|             | All the above fields should be available in UI & have  |
|             | field-level validations.                               |
+-------------+--------------------------------------------------------+
| FS_NB_004   | The system should be able to deduplicate and flag      |
|             | customers without hampering the customer creation      |
|             | process flow or slowing down the system. The system    |
|             | should facilitate multiple configurations &            |
|             | parameterization through the front end for setting up  |
|             | the deduplication process.                             |
+-------------+--------------------------------------------------------+
| FS_NB_005   | The system should have the capability to integrate     |
|             | with the latest technology such as OCR for capturing   |
|             | the data entered forms.                                |
+-------------+--------------------------------------------------------+
| FS_NB_006   | The system should provide rule-driven validations for  |
|             | the data fields (Example: Pincode & State mapping)     |
+-------------+--------------------------------------------------------+
| FS_NB_007   | The system should highlight any fields for which the   |
|             | customer data entered is outside the boundary          |
|             | conditions set for that specific field which should be |
|             | configurable.                                          |
+-------------+--------------------------------------------------------+
| FS_NB_008   | The system should handle page/module-level validations |
|             | during the proposal quality check.                     |
+-------------+--------------------------------------------------------+
| FS_NB_009   | The system should integrate with different channels    |
|             | such as web portals, mobility solutions, and partner   |
|             | channels such as Common Service Centre (including but  |
|             | not limited to) to capture customer information and    |
|             | undertake validations as per business rules.           |
+-------------+--------------------------------------------------------+
| FS_NB_010   | The system should auto-trigger any additional          |
|             | requirements for onboarding a customer.                |
+-------------+--------------------------------------------------------+
| FS_NB_011   | The system should have the facility to capture and     |
|             | update policyholder details, Life Assured details,     |
|             | Premium payer\'s details, Nominee's details, and       |
|             | additional parameters separately as per DoP\'s         |
|             | requirement.                                           |
+-------------+--------------------------------------------------------+
| FS_NB_012   | System to capture customer bank details as part of     |
|             | account creation (Example: Cancelled cheque can be     |
|             | used to capture bank account information, Account with |
|             | Post Office Savings & IPPBDoP)                         |
+-------------+--------------------------------------------------------+
| FS_NB_013   | The system should maintain the customer ECS mandates,  |
|             | Standing Instruction (SI), and NACH details.           |
+-------------+--------------------------------------------------------+
| FS_NB_014   | The system should validate various key customer KYC    |
|             | documents like Driving License, Birth Certificate,     |
|             | Bank details, PAN, Aadhar, Digilocker, and with other  |
|             | systems through web service integration (including but |
|             | not limited to)                                        |
+-------------+--------------------------------------------------------+
| FS_NB_015   | The system should create separate customers along with |
|             | different customer IDs as two Life Assured in case of  |
|             | Joint life scenarios.                                  |
+-------------+--------------------------------------------------------+
| FS_NB_016   | The system should perform an AML check for blacklisted |
|             | customers and agents before creating a customer entry  |
|             | into the system through third-party integration such   |
|             | as (including but not limited to) IRDAI, NSDL, UIDAI,  |
|             | DoP AML repository, integration, etc.                  |
+-------------+--------------------------------------------------------+
| FS_NB_017   | The system should have the facility to capture and     |
|             | update customer details basis the c-KYC/e-KYC-based    |
|             | authentication (Through Integration with UIDAI, NSDL,  |
|             | Digi locker, integration (including but not limited    |
|             | to) provided by the customer.                          |
+-------------+--------------------------------------------------------+
| FS_NB_018   | The system should have levels of KYC basis low and     |
|             | high-risk customer type, product offered, SA opted and |
|             | other parameters and should be compliant with the      |
|             | latest PMLA rules.                                     |
+-------------+--------------------------------------------------------+
| FS_NB_019   | The system should have a flag for KYC completeness and |
|             | once the KYC process is complete, the flag should      |
|             | automatically get activated.                           |
+-------------+--------------------------------------------------------+
| FS_NB_020   | The solution should be able to update personal data    |
|             | details during policy tenure by integrating real-time  |
|             | with other entities such as UIDAI, NSDL, and Digi      |
|             | locker, integration (including but not limited to).    |
+-------------+--------------------------------------------------------+
| FS_NB_021   | The system should be capable of configuring the        |
|             | customer ID sequencing number as per PLI or RPLI for   |
|             | new business (For Example: PLI Policies will start     |
|             | from 1 and RPLI will start from 5).                    |
+-------------+--------------------------------------------------------+
| FS_NB_022   | The solution must have a provision to consume DoP GCIF |
|             | (Global CIF) once the service to be consumed is        |
|             | available which will be a common Client ID for all DoP |
|             | LOBs - Postal, Banking & Insurance.                    |
+-------------+--------------------------------------------------------+
| FS_NB_023   | The system should create & configure riders as         |
|             | separate products along with creating service requests |
|             | through the front end.                                 |
+-------------+--------------------------------------------------------+
| FS_NB_024   | The system should have the capability to integrate     |
|             | with IRDAI, validate eIA & capture required details as |
|             | per IRDAI guidelines and regulations.                  |
+-------------+--------------------------------------------------------+
| FS_NB_025   | The solution must capture collection & payment         |
|             | channels at the time of Customer creation (Example:    |
|             | Standing Instructions, POSB, eBanking, mobile-Banking, |
|             | IPPBDoP, CSC, etc).                                    |
+-------------+--------------------------------------------------------+

### 3.2 Proposal Management

  -----------------------------------------------------------------------
  Requirement ID Requirement
  -------------- --------------------------------------------------------
  FS_NB_026      The system should be able to source policy from the
                 following channels but not limited to these channels.
                 Direct (Branch Walk-In), Agency, Web Portals, Mobile
                 Application, POS.

  FS_NB_027      The system should capture the basic details entered for
                 BI generation workflow and tag it with Proposal ID for
                 future reference in the PAS.

  FS_NB_028      The system should be able to capture the customer
                 details pre-populated from the quote generation module.

  FS_NB_029      The system should have the capability to integrate with
                 third-party KYC services (Aadhar, NSDL), etc.

  FS_NB_030      The system should be able to validate the data from the
                 images stored in the document management system with
                 proposal data through OCR and ICR functionality.

  FS_NB_031      The system should be able to validate the data entry
                 done by the operations team - all the field level and
                 length validations should be built into the data entry
                 platform.

  FS_NB_032      The system should populate the applicable
                 product-specific quality check questions required for
                 financial & medical u/w review.

  FS_NB_033      The system should perform a basic fraud analytic check
                 at the proposal generation stage basis the customer
                 details provided and previous customer data (if
                 available). The system should trigger additional
                 requirements.

  FS_NB_034      The system should trigger additional requirements in
                 case some documents are attached under Non-standard
                 proofs.

  FS_NB_035      The system should be able to display the proposal
                 summary highlighting the most important fields including
                 but not limited to - Name, DOB, Occupation, Sum Insured,
                 and Premium details.

  FS_NB_036      The system should be integrated with the workflow system
                 and status updates need to be provided to the respective
                 advisor and customer.

  FS_NB_037      The system should be able to provide a return journey
                 for the data entry person. Previously saved information
                 should be retrieved and populated.

  FS_NB_038      The system should maintain workflow details in the
                 workflow master and should maintain versions for
                 auditing purposes.
  -----------------------------------------------------------------------

### 3.3 Policy Issue

  -----------------------------------------------------------------------
  Requirement ID Requirement
  -------------- --------------------------------------------------------
  FS_NB_039      The system should be able to generate data files in any
                 format including but not limited to spool, CSV, etc. for
                 generating the policy kit.

  FS_NB_040      The system should be able to create an electronic policy
                 kit post-issuance through a registered email ID & send
                 it through different modes such as Email, WhatsApp, SMS,
                 etc.

  FS_NB_041      The solution can accept Proposals & issue Policies
                 instantly for selected Product Types from selected
                 Channels with authentication: Through eSign, Digital
                 Sign, Digi Locker, Aadhaar, PAN, e-KYC/c-KYC, eIA for
                 predetermined parameters including but not limited to
                 premium factor, and sum assured.

  FS_NB_042      The system should be able to create a hyperlink to open
                 the policy kit from the link. The hyperlink is available
                 in Channels, Portals, Registered Email ID, etc.

  FS_NB_043      The system should be integrated with NSDL for capturing
                 the e-Insurance Account Number (eIA) of customers used
                 for storing policy documents belonging to various
                 insurers.

  FS_NB_044      The system should update the policy details to EIA
                 through XML, direct API-based updates as per
                 requirements of eIA.

  FS_NB_045      The system should reserve the right of final Policy
                 Issuance with DoP as the final approver in the
                 maker-checker-approver process flow as per business
                 rules.

  FS_NB_046      The system should be able to update the policy and
                 customer-level data during the entire policy lifecycle
                 such as issuance, servicing, claims, etc. system
                 currently & in the future.

  FS_NB_047      The system should have the option to capture the
                 assignee/appointee details, in case the nominee is minor
                 as per business rules.

  FS_NB_048      The system should be able to capture the nominee\'s DOB
                 and auto-calculate the minor status of the nominee.

  FS_NB_049      The system should have a provision to capture the
                 relationship between the nominee and the proposer.

  FS_NB_050      The solution must support the Printing of Policy
                 Documents in any Pre-Designed Format, A4 Sheets, etc. as
                 per business requirements.

  FS_NB_051      System to capture the dispatch date and delivery date of
                 policy document as a reverse feed by integrating with
                 the postal module at DoP for tracking & managing the
                 free look period once the customer acknowledges the
                 policy kit

  FS_NB_052      The solution must populate the Freelook period in the
                 core policy admin System. SMS & Email/inbox notification
                 alerts should be sent to customers & agents. The
                 freelook period should be shown in reports as per
                 business requirements.

  FS_NB_053      System to provide reverse feed to various integrated
                 portals and CRM for the status of the dispatched policy
                 kit

  FS_NB_054      The system should have the capability to provide the
                 Free look period applicable basis with the channel
                 through which the policy was sourced & should be
                 configurable through the front end.

  FS_NB_055      The system should refer to the dispatch management
                 details either through physical post or sent through
                 email for deciding on the Free Look period validity. The
                 solution should be capable of integrating with other DoP
                 & third-party solutions (TPA)

  FS_NB_056      The system should have role-level rights/exceptions to
                 accept free-look cases outside the validity period as
                 per business requirements.

  FS_NB_057      The system should generate a premium and tax refund
                 accounting entry in case of Free Look Cancellation
                 (FLC). GST will not be refunded along with other
                 deductions such as Medical Practitioner fees, Stamp
                 duty, etc. The rules should be parameterized & set
                 through the front end.

  FS_NB_058      The system should generate refund receipts in FLC cases
                 through the front end after deducting all Admin charges,
                 Fees, etc.

  FS_NB_059      The system should provide the capability of bulk upload
                 through the front end of customer data in case of
                 employee policies of a corporation and individual for
                 all business processes such as issuance, claims
                 management, servicing, etc.

  FS_NB_060      The system should provide a bulk upload summary to the
                 processing team to identify any upload failures. The
                 solution must support Template, File Format, Reports as
                 per business requirement.

  FS_NB_061      The system should also perform a Deduplication check for
                 the corporate customer who already has an individual
                 policy.

  FS_NB_062      The solution should alert the user who has booked the
                 policy about the policy bond printing & dispatch status.
                 Such information must be visible through a dashboard &
                 alerts to all stakeholders should be sent.

  FS_NB_063      The solution must be capable of generating separate
                 policy block numbers for each Product, Type, etc as per
                 business requirement & parameterizable through UI
                 (Example: 1 series for PLI; 2 Series for RPLI, etc)

  FS_NB_064      System for policies in institutions like Railway,
                 Defence, etc. should send separate Letters to the DDO
                 for deducting Premium through employee salary. If any
                 change in Taxes (ex. at the 13th month) separate letter
                 to DDOs & Insurant stating a change in GST & Total
                 Premium Intimation Letter to DDO should be generated for
                 customers for any changes, alteration, closure, etc in
                 the Policy. SMS, Email/inbox notifications, and WhatsApp
                 alerts should be sent to DDO & customers for all such
                 events.

  FS_NB_065      The solution should have the option of Bulk Billing
                 Method Change (BMC) option from Cash to pay or vice
                 versa. (If pay recovery mode is discontinued for any
                 organization where their employees are covered, bulk
                 changing of policy status from Pay to Cash should be
                 available).

  FS_NB_066      The system should have OCR functionality (Optical
                 Character Recognition - Application of AI) which will be
                 used for reading information from physical documents
                 such as proposal forms & feeding the data into the
                 system.

  FS_NB_067      The OCR functionality should also help in verifying &
                 validating the ID Proof content with the proposer &
                 insured information captured during proposal fulfilment.
  -----------------------------------------------------------------------

### 3.4 Document Scanning & Management

+-------------+--------------------------------------------------------+
| Requirement | Requirement                                            |
| ID          |                                                        |
+:============+:=======================================================+
| FS_NB_068   | The Document Management System must have the ability   |
|             | to receive the documents described below including but |
|             | not limited to: - Investigation reports (scanned hard  |
|             | copies, electronic copies)                             |
|             |                                                        |
|             | - Photographs (.jpg, .mpg and other formats)           |
|             |                                                        |
|             | - Medical Reports (hard copies of various sizes        |
|             |   including uneven size and nature, Ex: ECG reports,   |
|             |   X-Ray, Scanning reports, and any other reports.) -   |
|             |   Date of Birth proof                                  |
|             |                                                        |
|             | - Address proof                                        |
|             |                                                        |
|             | - Photo ID proof                                       |
|             |                                                        |
|             | - Copy of the payment instrument (like cheque or DD)   |
|             |                                                        |
|             | - Copy of financial documents submitted by the client  |
|             |   (bank statements/ Balance Sheets/ ITRs) - Good       |
|             |   health declarations by the client                    |
|             |                                                        |
|             | - Proposal form                                        |
|             |                                                        |
|             | - Any other scanned documents from other DoP & TPA     |
|             |   solutions                                            |
|             |                                                        |
|             | The documents uploaded will be based on business       |
|             | requirements & acceptable in PDF / JPG format.         |
|             |                                                        |
|             | It should be a common repository for all documents     |
|             | used by all applications in the DoP landscape.         |
+-------------+--------------------------------------------------------+
| FS_NB_069   | The solution must have the facility to Scan, Page      |
|             | Indexing, Numbering & upload into DMS through Online & |
|             | Offline mode as per business requirements.             |
+-------------+--------------------------------------------------------+
| FS_NB_070   | The Document Management System must have the ability   |
|             | to receive claim-related documents described below     |
|             | including but not limited to: Claim intimation         |
|             | document (claim intimation letter or duly filled up    |
|             | claim form) Supporting documents like -                |
|             |                                                        |
|             | - Declarations/discharge vouchers by policy owner/     |
|             |   nominee/ assignee/ beneficiary                       |
|             |                                                        |
|             | - Medical reports, Death certificates, other claim     |
|             |   documents like police reports/ statements, doctor\'s |
|             |   written statements, witness statements Photographs   |
|             |   (hard copies and soft copies)                        |
|             |                                                        |
|             | - Investigation reports (soft copies, scanned copies   |
|             |   of hard copies, audio and video files)               |
|             |                                                        |
|             | - Good health/survival declarations, etc. in case of   |
|             |   maturity/survival benefits                           |
|             |                                                        |
|             | - Summons from the courts in case of legal claims      |
|             |                                                        |
|             | - Communication from the Central Vigilance             |
|             |   Commissioner where applicable                        |
|             |                                                        |
|             | - Communication from the Grievance Cell where          |
|             |   applicable                                           |
|             |                                                        |
|             | - Copies of the communication sent by the insured/     |
|             |   policy owner                                         |
|             |                                                        |
|             | - Letter of Indemnity where applicable (Original       |
|             |   Documents will be tracked in the Physical file.)     |
|             |   Financial Statement of the Life Insured              |
+-------------+--------------------------------------------------------+
| FS_NB_071   | The system must have the ability to store a copy of    |
|             | the Agency License with the agent record.              |
+-------------+--------------------------------------------------------+
| FS_NB_072   | The system should assign a unique identifier for each  |
|             | set of proposal documents & customer proofs            |
|             | submitted/uploaded.                                    |
+-------------+--------------------------------------------------------+
| FS_NB_073   | The system should have integration with the core       |
|             | insurance solution where data entry users can view the |
|             | proposal form for quick TAT.                           |
+-------------+--------------------------------------------------------+
| FS_NB_074   | Documents should be accessible through the core policy |
|             | administration system by clicking the respective       |
|             | document hyperlink.                                    |
+-------------+--------------------------------------------------------+
| FS_NB_075   | The system should have the ability to keep track of    |
|             | changes made to the documents through an audit trail.  |
+-------------+--------------------------------------------------------+
| FS_NB_076   | The system should have the ability to restrict access  |
|             | to the documents to a select group of users.           |
+-------------+--------------------------------------------------------+
| FS_NB_077   | The system should have search functionality based on   |
|             | parameters such as Proposal Number / Policy Number /   |
|             | Service Request Number / Name / Mobile No. / PAN Card  |
|             | Number / Aadhaar Card Number / Email, etc              |
+-------------+--------------------------------------------------------+
| FS_NB_078   | The system should archive documents that have exceeded |
|             | the retention period limit as defined by DoP. The      |
|             | retention period limit should be configurable through  |
|             | the front end.                                         |
+-------------+--------------------------------------------------------+
| FS_NB_079   | The system should provide a UI facility to retain      |
|             | documents beyond the retention period and allow        |
|             | provision for superuser access for deletion/archival.  |
+-------------+--------------------------------------------------------+
| FS_NB_080   | The system should have the ability to allow document   |
|             | upload from different channels but not limited to,     |
|             | customer portal, agent portal, employer portal,        |
|             | Mobility solution, etc.                                |
+-------------+--------------------------------------------------------+
| FS_NB_081   | The system should be integrated for obtaining eKYC     |
|             | from third-party agencies for example Digilocker, etc. |
+-------------+--------------------------------------------------------+
| FS_NB_082   | The system should provide analytics by way of a        |
|             | Dashboard including but not limited to showing the     |
|             | health of the DMS in terms of size consumed,           |
|             | availability, inward and outward document details      |
|             | along with its size, etc.                              |
+-------------+--------------------------------------------------------+
| FS_NB_083   | The system should have a provision to allow users to   |
|             | upload documents through the policy admin system       |
|             | directly.                                              |
+-------------+--------------------------------------------------------+
| FS_NB_084   | The system should have a provision to configure        |
|             | validations for document upload. Some of the           |
|             | validation parameters should include but are not       |
|             | limited to inward file type, size, timestamp, etc.     |
+-------------+--------------------------------------------------------+
| FS_NB_085   | The solution should have provision to store small      |
|             | sizes of data in a Database rather than DMS for easy   |
|             | retrieval & costing.                                   |
+-------------+--------------------------------------------------------+
| FS_NB_086   | The solution must have a Unique ID for all inward Docs |
|             | and separate Block numbers for each integrated         |
|             | Solution (Block Number Example: \'1\' series for       |
|             | documents uploaded through IPPB, DoP; \"2\" series for |
|             | documents uploaded through certain employers).         |
+-------------+--------------------------------------------------------+
| FS_NB_087   | The solution must have the capability of storing all   |
|             | types of Letters generated in PAS & Purged             |
|             | periodically as per business requirements. The purging |
|             | limit & frequency should be configurable through the   |
|             | front end.                                             |
+-------------+--------------------------------------------------------+
| FS_NB_088   | The solution must be capable of replacing/adding to    |
|             | the user-uploaded Docs with the correct set of         |
|             | documents through the maker checker workflow up to a   |
|             | certain period as per business requirements. The       |
|             | period duration should be configurable through UI.     |
+-------------+--------------------------------------------------------+
| FS_NB_089   | The solution must have a Global Archival, Purging, and |
|             | Retrieval Policy in DMS as per business requirements.  |
+-------------+--------------------------------------------------------+
| FS_NB_090   | The solution must trigger SMS & Email Alerts to the IT |
|             | Team for Threshold breaches, Security threats, etc.    |
+-------------+--------------------------------------------------------+

### 3.5 Approval Workflow

  -----------------------------------------------------------------------
  Requirement ID Requirement
  -------------- --------------------------------------------------------
  FS_NB_091      The system shall route policy data to Quality Reviewer
                 after CPC data entry.

  FS_NB_092      The Quality Reviewer shall have the ability to approve,
                 reject, or send back for corrections.

  FS_NB_093      After Quality Reviewer approval, the policy moves to
                 Approver for final authorization.

  FS_NB_094      The Approver shall have the ability to approve or reject
                 the policy issue.

  FS_NB_095      The system shall maintain status and timestamp for each
                 approval step.

  FS_NB_096      Notifications shall be sent to relevant users on status
                 changes.

  FS_NB_097      The system shall allow audit of the entire approval
                 history.
  -----------------------------------------------------------------------

### 3.6 Receipt and Printing Management

  -----------------------------------------------------------------------
  Requirement ID Requirement
  -------------- --------------------------------------------------------
  FS_NB_098      The solution must be capable of printing barcodes as per
                 business rules in the Acceptance Memo & Envelope for
                 Dispatch & Delivery.

  FS_NB_099      The solution must be capable of printing QR Codes, etc.
                 as per business rules in Print Bond, Premium Receipt
                 Book & any other Docs as per business requirement.

  FS_NB_100      The solution must support Centralized / Decentralized
                 Printing of all Documents including Policy bonds in DoP
                 & other identified locations, and vendors.

  FS_NB_101      The solution must integrate with any Printing Solutions
                 to Print various Letters, Sanction Memos, Calculation
                 sheets, Licenses, Quarterly Statements & any other as
                 per business requirements.
  -----------------------------------------------------------------------

### 3.7 Policy Bond

  -----------------------------------------------------------------------
  Requirement ID Requirement
  -------------- --------------------------------------------------------
  FS_NB_102      The system shall generate a policy bond document post
                 final approval.

  FS_NB_103      The bond shall include client details, policy terms,
                 premium, coverage, and signatures.

  FS_NB_104      The system shall allow digital signing of the bond where
                 applicable.

  FS_NB_105      The system shall store a PDF version of the policy bond
                 linked to the client and policy.

  FS_NB_106      The system shall allow retrieval and printing of the
                 policy bond at any time.
  -----------------------------------------------------------------------

## **4. Functional Requirements/ Business Rules**

**Flow Chart for Policy Issue Process:**

![A diagram of a policy issue process AI-generated content may be
incorrect.](media/image1.jpg){width="6.268055555555556in"
height="3.6708333333333334in"}

1.  **Proposal Workflow Diagram Counter without Medical with Digilocker
    and Aadhar KYC:**![](media/image2.png){width="6.095667104111986in"
    height="8.630912073490814in"}

2.  **Proposal workflow Diagram counter with
    Medical:**![](media/image3.png){width="6.268055555555556in"
    height="8.848611111111111in"}

**Policy Issue Cover Page:**

- **Fields:**

  - Policy Issue- Without Aadhar: button

  - Policy Issue- With Aadhar: button

### Policy Issue- Without Aadhar Page

1.  Clicking 'Policy Issue- Without Aadhar' should open this page.

- **Fields:**

  - Creating New Application: button

  - Search for an Existing Application: button

- **Business Rules:**

  - Clicking 'Creating New Application' should move a user to 'New
    Business Indexing Page'.

  - Clicking 'Search for an Existing Application' should move a user to
    'Search Existing Proposal' page.

**Process flow for Policy Issue Process of Customers who do not have
Aadhar:**

![](media/image4.png){width="6.268055555555556in" height="4.6125in"}

1.  New Business Indexing Page.

- **Fields:**

  - Product Type: Dropdown: Contain PLI/RPLI

  - Product Name: Dropdown: Contain list of all products for PLI/RPLI as
    selected above in Product Type dropdown.

  - Product Description: Non-editable Text box with pre-filled details

  - Application Receipt Date: Date Field

  - Date of Proposal: Date Field

  - Date of Declaration: Date Field

  - Date of Indexing: Date Field

  - PO Code: Text box

  - First Name: Text box

  - Middle Name: Text box

  - Last Name: Text box

  - Date of Birth: Date Field

  - Gender: Dropdown with Male, Female, Other as option.

  - Opportunity ID: Text box

  - Issue Circle: Dropdown

  - Issue HO: Dropdown

  - Issue Post Office: Dropdown

  - Sum Assured: Text Box

  - Premium Ceasing Age: Dropdown

  - Premium Payment Frequency: Dropdown

  - Calculate Premium: Button

  - Premium Amount: Non-Editable field to display premium after
    Calculate premium button is pressed.

  - Submit: button

  - Previous: button

- **Business Rules:**

<!-- -->

- After clicking submit, the following page should display:

  - Proposal Number: Non-Editable text box with pre-filled proposal
    Number.

  - Print Customer Acknowledgement Slip.: button: Clicking should
    generate the receipt.

  - Pay Premium: button: Clicking should open Premium Collection page.

  - New Indexing: button: Clicking should Open New Proposal Indexing
    page.

+---------+-----------------+------------------------------------------+
| Serial# | Error Message   | Required Action                          |
+=========+=================+==========================================+
| 1       | Date of         | Users must carefully check the date of   |
|         | declaration     | declaration, the date of receipt of      |
|         | later than Date | application, the date of indexing, and   |
|         | of Indexing     | the date of Proposal. These activities   |
|         |                 | happen in the following sequence:        |
|         |                 |                                          |
|         |                 | - Declaration signed by customer         |
|         |                 |                                          |
|         |                 | - Application received                   |
|         |                 |                                          |
|         |                 | - Indexing performed for an application  |
|         |                 |   Proposal generated                     |
|         |                 |                                          |
|         |                 | Hence, the date of declaration will      |
|         |                 | always be earlier than the date of       |
|         |                 | receipt of application and the date of   |
|         |                 | indexing. Similarly, the date of receipt |
|         |                 | of application will always be earlier    |
|         |                 | than the date of indexing.               |
|         |                 |                                          |
|         |                 | The date of proposal will always be      |
|         |                 | later than all of the above dates.       |
+---------+-----------------+                                          |
| 2       | Date of         |                                          |
|         | declaration     |                                          |
|         | later than      |                                          |
|         | Application     |                                          |
|         | Receipt date    |                                          |
+---------+-----------------+                                          |
| 3       | Date of         |                                          |
|         | Proposal later  |                                          |
|         | than Date of    |                                          |
|         | Indexing        |                                          |
+---------+-----------------+                                          |
| 4       | Date of         |                                          |
|         | Proposal later  |                                          |
|         | than Date of    |                                          |
|         | Declaration     |                                          |
+---------+-----------------+                                          |
| 5       | Date of         |                                          |
|         | Proposal later  |                                          |
|         | than            |                                          |
|         | Application     |                                          |
|         | Receipt date    |                                          |
+---------+-----------------+                                          |
| 6       | Application     |                                          |
|         | Receipt date    |                                          |
|         | later than Date |                                          |
|         | of Indexing     |                                          |
+---------+-----------------+------------------------------------------+

1.  Search Existing Proposal Page.

- **Fields:**

  - Proposal Number: Text box

  - Fetch: button

- **Business Rules:**

  - Clicking 'Fetch' should open the searched proposal.

    1.  'Search Ticket Page' for CPC User to search for
        Proposal/Policies.

- **Fields:**

  - Request Queue: Dropdown:

  - Stage Date Range: To and From Date field

  - Operation Center: Dropdown

  - Status: Dropdown

  - Proposal Number: Text box

  - Product: Dropdown

  - Request Type: Dropdown

  - Policy Number: Text Box

  - Search: button

- **Business Rules:**

  - Clicking 'Search' should Display the following:

    - Ticket ID: Non-Editable text

    - Customer ID: Non-Editable text

    - Policy No./Proposal No.: Non-Editable text

    - Request Type: Non-Editable text

    - Status: Non-Editable text

    - Request Date/Time: Non-Editable text

    - Request Owner: Non-Editable text

    - Indexed By: Non-Editable text

    - Office: Non-Editable text

    - Actions: button: Link to open the proposal

  - The opened proposal should contain the following pages:

    - Insured Details Page

    - Nomination Page

    - Policy Details Page

    - Agent Details Page

    - Medical Information Page

    1.  Insured Details Page.

- **Fields:**

  - Salutation

  - First Name

  - Middle Name

  - Last Name

  - Gender

  - Marital Status

  - Father's Name

  - Husband Name

  - Date of Birth

  - Age Proof

  - Aadhar ID

  - Nationality

  - Communication Address:

    - Address Line1

    - Address Line2

    - Village

    - Taluka

    - City

    - District

    - State

    - Country

    - Pin Code

  - Permanent Address Same as Communication Address: Checkbox: Default
    is Checked. If Unchecked, the following fields should display:

    - Permanent Address:

      - Address Line1

      - Address Line2

      - Village

      - Taluka

      - City

      - District

      - State

      - Country

      - Pin Code

  - Contact Details:

  - Contact Type: Resident/Official

    - Area/STD Code

    - Landline Number

  - Mobile Number

  - Email

  - Insured and Proposer are Same: Checkbox: Default is Checked

  - Employment Details: Checkbox: Default is Unchecked

    - If 'Employment Details' Checkbox is Checked by the user, display
      the following fields:

      - Occupation

      - PAO/DDO Code

      - Organization

      - Designation

      - Date of entry in service

      - Designation of immediate superior

      - PAN Number

      - Monthly Income

      - Employer Address

        - Address Line1

        - Address Line2

        - Village

        - Taluka

        - City

        - District

        - State

        - Country

        - Pin Code

        - Official Phone Number

        - Official Email

      - Qualification

  - Add Spouse: Button: Option present only for Yugal Suraksha to add
    all the above details for Spouse also.

  - Subsequent Payment Method: dropdown: Options are Cash, Online, NACH,
    etc.

  - Bank Account Number for Payout: Text: For Payout during Maturity,
    Claim, etc.

  - IFSC Code: Text: For Payout during Maturity, Claim, etc.

  - If Insured is Married Female, display additional below fields:

    - Number of Children: Textbox

    - Date of Last Delivery: Calendar

    - If pregnant, then expected month of delivery: Dropdown: Options is
      number 1 to 9.

    - Mark of Identification-1: Textbox

    - Mark of Identification -2: Textbox

  - In Case of Children Policy, Add below fields:

    - Mother's Name: Textbox

    - Parent's Policy Number: Option to search the policy and select
      should be present.

- **Business Rules:**

  - 'Insured and Proposer are Same' should be checked by Default.

  - If 'Insured and Proposer are Same' is checked (Default), do not
    display any more fields and use the same data given for insured to
    store in proposer details tables also.

  - If 'Insured and Proposer are Same' is unchecked, display all the
    fields of Insured section again to be inputted for the proposer
    including Name, Gender, Marital Status, Father's Name, Husband Name,
    Date of Birth, Age Proof, Aadhar ID, Nationality, Communication
    Address, Present Address, Contact Details, Mobile and Email.

  - If 'Employment Details' Checkbox is Checked by the user, then only,
    display the employment related fields. By, Default, this field
    should be unchecked.

  -----------------------------------------------------------------------
  Serial#   Error Message     Required Action
  --------- ----------------- -------------------------------------------
  1         Missing Spouse    In Yugal Suraksha proposal, spouse details
            details           (first name/date of birth/gender) are to be
                              provided. Enter appropriate spouse details.

  2         Missing Insured   Provide the name of the husband if the
            Husband's Name    insured is a married female.

  3         Missing Insured   Provide the name of the father if the
            Father\'s Name    insured is a male or an unmarried female.

  4         Missing Insured   This error message is displayed if all the
            Communication     fields or address 1st line from the
            Address           communication address is missing. Users
                              must enter appropriate information in all
                              the fields in the Communication Address
                              section.

  5         Missing Insured   This error message is displayed if all the
            Permanent Address fields or address 1st line from the
                              permanent address is missing. Users must
                              enter appropriate information in all the
                              fields in the Permanent Address section.

  6         Missing Proposer  This message is displayed if the
            Relationship with relationship of the insured and the
            Insured           proposer is not entered. Users must refer
                              to the application form and enter this
                              information.

  7         Missing Proposer  This message is displayed if the proposer
            Husband\'s name   is married female and husband's name is not
                              entered on the screen. Users must enter the
                              Husband's name.

  8         Missing Proposer  This message is displayed if the proposer
            Father\'s Name    is male or unmarried female and the
                              father's name is not entered on the screen.
                              Users must enter the father's name.

  9         Missing City or   Users must enter the name of either the
            Village --        city or the village of the proposer.
            Proposer address  

  10        Missing Date of   This message is displayed if the date on
            Entry             which an insured joined his/her current
                              organization has not been entered. Users
                              must enter the appropriate date of joining.

  11        Missing           If the declaration date is not provided or
            Declaration date  has been deleted.

  12        Missing           If the Application receipt date is not
            Application       provided or has been deleted.
            receipt date      

  13        Missing Policy    Is none of the radio buttons
            Taken under       (HUF/MWPA/Others) has been selected.
  -----------------------------------------------------------------------

![A screenshot of a computer AI-generated content may be
incorrect.](media/image5.png){width="4.987179571303587in"
height="7.48076990376203in"}

1.  Nomination Page.

- **Fields:**

  - Add Nominee: button

    - Clicking of 'Add Nominee' button should lead to display of
      following input fields:

      - Salutation

      - First Name

      - Middle Name

      - Last Name

      - Gender

      - Date of Birth

      - Relationship: Dropdown

      - Share Percentage: Text box

      - Address:

        - Address Line1

        - Address Line2

        - Village

        - Taluka

        - City

        - District

        - State

        - Country

        - Pin Code

      - Phone Number/Mobile

      - Email Address

- **Business Rules:**

  - At max, 3 nominee addition is allowed.

  - Sum of input given in 'Share Percentage' field for all the added
    nominees should be 100%.

  - Nomination is mandatory if HUF, MWPA is not selected.

  -----------------------------------------------------------------------
  Serial#   Error Message     Required Action
  --------- ----------------- -------------------------------------------
  1         Missing Nominee   If the relationship of the nominee with the
            Relationship      insured has not been provided. Users need
                              to specify the relationship.

  2         Missing Nominee   If the share percentage of the sum assured
            share percentage  to be received by various nominees has not
                              been provided. Users need to enter the
                              share percentage.
  -----------------------------------------------------------------------

1.  Policy Details Page.

- **Fields:**

> [Policy Information Section:]{.underline}

- Application Declaration Date

- Date of Acceptance

- Issue Post Office

- Application Receipt date

- Date of Commencement of Risk

- Policy Taken Under: Dropdown: HUF/MWPA/Other: If Other, display
  Textbox.

  - If MWPA is Selected as option for 'Policy Taken Under' dropdown,
    Display the Following additional fields:

    - Do you want to appoint a Trustee for Policy issued under the
      Married Women's Property Act, 1874? : Dropdown: Options are Yes,
      No.

      - If Yes is Selected, Display the following additional fields:

        - Trust Type: Dropdown: Options are Individual, Corporate

        - Trust/Trustee Name: Textbox

        - Trustee Date of Birth: Calendar

        - Trustee Relationship: Textbox

        - Trustee Address: Textbox

  - If HUF is Selected as option for 'Policy Taken Under' dropdown,
    display the following additional fields:

    - Is this Insurance policy financed under HUF Funds?: Dropdown:
      Options are Yes, No

    - Full Name of Karta: Textbox

    - PAN Card Number for HUF: Textbox

    - Is Life assured is different from the Karta, please provide the
      reason as why Karta is not proposing Insurance on his own Life. :
      Dropdown: Options are Yes, No

      - If Yes is selected, display the following:

        - Reason: Textbox: Mandatory if Yes is selected

    - Add HUF Member: button: Option to Details of Co- Parceners/Members
      of HUF.

      - If 'Add HUF Member' button is clicked, display the option to add
        HUF Member by asking the following:

        - Member Name: Textbox

        - Member Relationship: Textbox

        - Member Age: Textbox

      - Option should be given to add at max 7 Members.

> 
>
> [Additional Questions Section:]{.underline}

- Does the proponent presently hold any PLI/RPLI Insurance policy? :
  Dropdown: (Yes/No): If Answer is yes, give option to add policy
  details.

- Does the proponent presently hold any Non PLI/RPLI Insurance policy? :
  Dropdown: (Yes/No): If Answer is yes, give option to add policy
  details.

- Does the proponent presently hold life insurance/health & non-life
  insurance policies of other companies? : Dropdown: (Yes/No): If Answer
  is yes, give option to add policy details

> [Base Coverage Section]{.underline}: Auto-Populate

- Policy Type

- Premium Ceasing Age

- Sum Assured

- Medical

- Coverage Age

> [Premium Information Section:]{.underline}

- Annual Premium Equivalent

- Premium- Initial

- Additional Premium

- Premium Payment Method

- Subsequent Premium Payment Mode

- Premium Payment Frequency

- Modal Premium

- Premium Payable

- Short/Excess Premium

<!-- -->

- **Business Rules:**

  - 1.  Agent Details Page.

- **Fields:**

  - Proposal Without Agent- Checkbox-Default is Checked. Below displays
    will be displayed only if this checkbox is checked.

    - Agent ID: Text box

    - Search Agent: button: To search and select the Agent from existing
      database.

    - Receives Correspondence- Radiobutton: Yes/No

    - Salutation

    - First Name

    - Middle Name

    - Last Name

    - Opportunity ID

    - Area/STD Code

    - Landline Number

    - Mobile Number

    - Email Address

- **Business Rules:**

  - 1.  Medical Information Page- General

- **Fields:**

  - Are you in Sound Health at Present? Dropdowns: Yes/No: Default-Yes

  - Have you ever suffered/suffering from any of the following

    - Tuberculosis: Checkbox: Default- Unchecked

    - Cancer: Checkbox: Default- Unchecked

    - Paralysis: Checkbox: Default- Unchecked

    - Insanity: Checkbox: Default- Unchecked

    - Any disease of heart and lungs: Checkbox: Default- Unchecked

    - Kidney Disease: Checkbox: Default- Unchecked

    - Any Disease of Brain: Checkbox: Default- Unchecked

    - HIV Positive: Checkbox: Default- Unchecked

    - Hepatitis-B: Checkbox: Default- Unchecked

    - Epilepsy: Checkbox: Default- Unchecked

    - Nervous Disorder: Checkbox: Default- Unchecked

    - Liver: Checkbox: Default- Unchecked

    - Leprosy: Checkbox: Default- Unchecked

    - Any Physical Deformity or Handicap: Checkbox: Default- Unchecked

    - Any other serious disease: Checkbox: Default- Unchecked

  - Details of Disease- Text box to be filled if disease Exist

  - Has any of your family members living or dead suffered from any
    hereditary or infectious disease like
    Insanity/Epilepsy/Gout/Asthma/Tuberculosis/Cancer/Leprosy, etc.? :
    Dropdown: Yes/No: Default is No

    - If Yes, Give Details

  - Have you availed any kind of leave on medical ground or hospitalized
    during the last 3 years?: Dropdown: Yes/No: Default: No

    - If Yes, Give Details of:

    - Kind of Leave

    - Period of Leave

    - Ailment

    - Name of Hospital

    - Period of Hospitalization:

    - From

    - To

  - Do you have any physical deformity or congenital by birth defects?:
    Dropdown: Yes/No: Default: No. The below options will display only
    if this dropdown is set to Yes.

    - Congenital: Checkbox: Default- Unchecked

    - Non-Congenital: Checkbox: Default- Unchecked

    - Congenital/Non-Congenital: Checkbox: Default- Unchecked

  - Particulars of Family Doctor, If any

    - Doctor Name

  - Accepted: button

- **Business Rules:**

  - Add two checkboxes in each for Yugal Suraksha Policy with default
    options selected.

    1.  Medical Information Page- Policy Greater than 20 Lakh Sum
        Assured.

- **Fields:**

Additional Health Details Needed if Sum Assured is greater than 20
Lakhs:

(i) Are you currently undergoing/have undergone any tests,
    investigations, awaiting results of any tests, investigation or have
    you ever been advised to undergo any tests, investigations or
    surgery or been hospitalised for general check-up, observations,
    treatment or surgery : Dropdown: Yes/No: Default is No

(ii) Diabetes/ High Blood Sugar : Dropdown: Yes/No: Default is No

(iii) High/ Low Blood Pressure : Dropdown: Yes/No: Default is No

(iv) Have you ever been referred to an : Oncologist or cancer hospital
     for any investigation or treatment : Dropdown: Yes/No: Default is
     No

(v) Did you have any ailment/injury/accident requiring
    treatment//medication for more than a week : Dropdown: Yes/No:
    Default is No

(vi) Have you ever suffered Thyroid disorder or any other disease or
     disorder of the endocrine system : Dropdown: Yes/No: Default is No

(vii) \(vii\) Ave you undergone/have been recommended to undergo
      Angioplasty , bypass surgery, brain surgery, Heart valve surgery
      Aorta surgery or organ transplant : Dropdown: Yes/No: Default is
      No

(viii) Have you ever suffered disorders of eye, ear, nose, throat,
       including defective sight speech or hearing & discharge from ears
       : Dropdown: Yes/No: Default is No

(ix) Have you ever suffered Anaemia, blood or blood related disorders :
     Dropdown: Yes/No: Default is No

(x) Have you ever suffered musculoskeletal disorders such as arthritis,
    recurrent back pain, slipped disc or any other disorder of spine,
    joints, limbs or leprosy : Dropdown: Yes/No: Default is No

**Additional Health Information for Female Proponent (In case of Sum
Assured or Aggregate Sum Assured exceeding 20 lakh)**

i.  Have you ever have any abortion, miscarriage or ectopic pregnancy :
    Dropdown: Yes/No: Default is No

ii. Have you ever undergone any gynaecological investigations, internal
    checkups, breast checkups such as mammogram or biopsy : Dropdown:
    Yes/No: Default is No

iii. Have you ever consulted a Doctor because of an irregularity at the
     breast, vagina, uterus, ovary, fallopian tubes, menstruation, birth
     delivery, complications during pregnancy or child delivery or a
     sexually transmitted diseases? : Dropdown: Yes/No: Default is No

**Personal habits of the proponent impacting health (Required in case of
Sum Assured/ Aggregate Sum Assured is above \`20 lakh)**

If Yes, Whether Frequently or Occasionally

1.  Do you Smoke/ Consume Tobacco? : Dropdown: Yes/No, If Yes, display
    one more dropdown asking whether Frequently or Occasionally. :
    Default is No.

2.  Do you Consume Alcohol? : Dropdown: Yes/No, If Yes, display one more
    dropdown asking whether Frequently or Occasionally. : Default is No.

3.  Do you Consume Drugs? : Dropdown: Yes/No, If Yes, display one more
    dropdown asking whether Frequently or Occasionally. : Default is No.

4.  Do you have any habits, which can adversely impact your health? :
    Dropdown: Yes/No, If Yes, display one more dropdown asking whether
    Frequently or Occasionally. : Default is No.

- **Business Rules:**

  - 'Medical Information Page- Policy Greater than 20 Lakh Sum Assured'
    page should display only if Sum Assured selected by the insured is
    greater than or equal to 20 Lakh.

  - Add two checkboxes in each for Yugal Suraksha Policy with default
    options selected.

    1.  Document Upload Page.

- **Fields:**

  - Add Document: button

  - Attached Existing Document: Table

  - Download: button

  - Delete Document: button

- **Business Rules:**

  - Existing Scanned Document should get uploaded in this page.

  - The page should have option to upload the documents.

  - The page should have option to download the documents.

  - The page should have option to delete the documents.

  - Each document should be versioned and comment note should be added
    for each of the documents.

    1.  Approval Page.

- **Fields:**

  - Approval: Dropdown: Options include Approve, Reject.

  - Submit: button

- **Business Rules:**

  - Quality Reviewer and Approver both should be able to approve the
    Proposal.

  - Policy Number should get generated only after the approval of
    Approver.

  - Upon Rejection, User should input the comments for Rejection.

- **Levels of Approval: There are three levels of Approvers in the
  system:**

  - Approver 1: They can approve proposals with Sum Assured (SA) value
    less than or equal to 500000.

  - Approver 2: They approve proposals with sum assured value more than
    500000 and less than 1000000.

  - Approver 3: They approve proposals with sum assured value more than
    1000000.

### Policy Issue- With Aadhar Page

1.  Clicking 'Policy Issue- With Aadhar' should open this page.

- **Fields:**

  - Aadhar Number: textbox

  - Submit: button

- **Business Rules:**

  - Aadhar Integration should happen in such a way that clicking submit
    button should fetch the following details automatically from Aadhar:

    - Full Name

    - Date of Birth

    - Gender

    - Address

    - Photograph (Base64 image)

    - Email (if linked)

    - Mobile Number

![A screenshot of a contact form AI-generated content may be
incorrect.](media/image6.png){width="2.7604166666666665in"
height="4.140624453193351in"}

1.  User Input page.

- **Fields:**

  - PAN: Textbox

  - Product Type: Dropdown: PLI/RPLI

  - Product Name: Dropdown

  - Premium Ceasing Age/Term

  - Sum Assured

  - Premium Frequency

  - Subsequent Premium Payment Mode

  - Add Nominee: button

    - If Add Nominee button is clicked, display the following to be
      added:

      - Salutation

      - First Name

      - Middle Name

      - Last Name

      - Gender

      - Date of Birth

      - Relationship: Dropdown

      - Share Percentage: Text box

      - Address:

        - Address Line1

        - Address Line2

        - Village

        - Taluka

        - City

        - District

        - State

        - Country

        - Pin Code

      - Phone Number/Mobile

      - Email Address

  - Father\'s/Husband\'s Name: Text box

  - Marital Status

  - Add Employer: button

    - If Add Employer button is clicked, display the following fields
      for adding employer details:

      - Occupation

      - PAO Code

      - Organization

      - Designation

      - Date of entry in service

      - Designation of immediate supervisor

      - PAN Number

      - Monthly Income

      - Employer Address

        - Address Line1

        - Address Line2

        - Village

        - Taluka

        - City

        - District

        - State

        - Country

        - Pin Code

        - Official Phone Number

        - Official Email

      - Qualification

  - Mobile Number: Text box: Input the Aadhar fetched mobile number by
    default

  - Email: Text box: Input the Aadhar fetched Email by default

  - Agent ID: Text box

  - Policy Taken Under: Dropdown: HUF/MWPA/Other: If Other, display
    Textbox.

    - If MWPA is Selected as option for 'Policy Taken Under' dropdown,
      Display the Following additional fields:

      - Do you want to appoint a Trustee for Policy issued under the
        Married Women's Property Act, 1874? : Dropdown: Options are Yes,
        No.

        - If Yes is Selected, Display the following additional fields:

          - Trust Type: Dropdown: Options are Individual, Corporate

          - Trust/Trustee Name: Textbox

          - Trustee Date of Birth: Calendar

          - Trustee Relationship: Textbox

          - Trustee Address: Textbox

    - If HUF is Selected as option for 'Policy Taken Under' dropdown,
      display the following additional fields:

      - Is this Insurance policy financed under HUF Funds?: Dropdown:
        Options are Yes, No

      - Full Name of Karta: Textbox

      - PAN Card Number for HUF: Textbox

      - Is Life assured is different from the Karta, please provide the
        reason as why Karta is not proposing Insurance on his own Life.
        : Dropdown: Options are Yes, No

        - If Yes is selected, display the following:

          - Reason: Textbox: Mandatory if Yes is selected

      - Add HUF Member: button: Option to Details of Co-
        Parceners/Members of HUF.

        - If 'Add HUF Member' button is clicked, display the option to
          add HUF Member by asking the following:

          - Member Name: Textbox

          - Member Relationship: Textbox

          - Member Age: Textbox

        - Option should be given to add at max 7 Members.

  - Medical Information: (Separate Popup for user to confirm with
    default options selected): Popup should contain details as per
    section 4.1.9 and 4.1.10 (Sum Assured Greater than 10 lakh).

  - Add Spouse: Button: Option present only for Yugal Suraksha to add
    all the above details for Spouse also.

  - Subsequent Payment Method: dropdown: Options are Cash, Online, NACH,
    etc.

  - Bank Account Number for Payout: Text: For Payout during Maturity,
    Claim, etc.

  - IFSC Code: Text: For Payout during Maturity, Claim, etc.

  - If Insured is Married Female, display additional below fields:

    - Number of Children: Textbox

    - Date of Last Delivery: Calendar

    - If pregnant, then expected month of delivery: Dropdown: Options is
      number 1 to 9.

    - Mark of Identification-1: Textbox

    - Mark of Identification -2: Textbox

  - In Case of Children Policy, Add below fields:

    - Mother's Name: Textbox

    - Parent's Policy Number: Option to search the policy and select
      should be present.

  - Calculate Premium: button

- **Business Rules:**

  - Clicking 'Calculate Premium' should display the premium details.

  - Option to Pay premium using Cash & Cheque (For Post office)/ Online
    modes should be present.

![A screenshot of a form AI-generated content may be
incorrect.](media/image7.png){width="3.703546587926509in"
height="5.555321522309711in"}

1.  Document Upload page.

- **Fields:**

  - Add Document: button

  - Attached Existing Document: Table

  - Download: button

  - Delete Document: button

- **Business Rules:**

  - User should be able to upload the additional document needed.

    1.  Review and Submit page.

- **Fields:**

  - Policy Details: Table

  - Premium Information: Table

  - Uploaded Document List: Table

  - Submit: button

  - Cancel: button

- **Business Rules:**

  - Submitting the form should display the Proposal number.

  - CPC User should be able to open and Proceed with the proposal as per
    the established process.

  - The Policy Number should get generated automatically for the
    non-Medical Policies and Policy Bond should be sent to
    WhatsApp/Email for the Insured.

![A screenshot of a document list AI-generated content may be
incorrect.](media/image8.png){width="2.8958333333333335in"
height="4.343748906386701in"}

### Proposal Creation- Using File Upload

**Flow Diagram for the Policy Issue Process using File Upload:**

![](media/image9.png){width="3.9068219597550304in"
height="5.860232939632546in"}

1.  Proposal Creation Using File Upload Page.

- **Fields:**

  - Upload File: Excel/CSV file input

  - Download Template: Button to download the standard template

  - Instructions Panel: Guidelines for filling the template (e.g.,
    mandatory fields, format rules)

  - Upload Button: To initiate file upload

  - Validation Summary: Display of success/failure count after upload

  - Error Report Download: Button to download error details for failed
    rows

- **Business Rules:**

  - Successfully Uploading the form should display the list of Proposal
    numbers in the Interface.

  - Option should be there in Interface to download the list of
    successful proposals with Proposal Number.

  - CPC User should be able to open and Proceed with the proposal as per
    the established process.

## **5. Wireframe**

![](media/image10.png){width="2.7407863079615047in"
height="2.448969816272966in"}

### 5.1 Policy Issue- without Aadhar Page

Fig-1: Insured Details Page for Proposal (CPC User View)

![A screenshot of a computer AI-generated content may be
incorrect.](media/image5.png){width="4.987179571303587in"
height="7.48076990376203in"}

Fig-2: Policy Details Page:

![](media/image11.png){width="5.721449037620298in"
height="8.582175196850393in"}

### 5.2 Policy Issue- with Aadhar Page

Fig-1: Input Aadhar and details will be automatically fetched.

![A screenshot of a contact form AI-generated content may be
incorrect.](media/image6.png){width="3.703546587926509in"
height="5.555319335083115in"}

Fig-2: User Input Page:

![A screenshot of a form AI-generated content may be
incorrect.](media/image7.png){width="3.703546587926509in"
height="5.555321522309711in"}

Fig-3: Document Upload Page

![A document upload form AI-generated content may be
incorrect.](media/image12.png){width="2.8471456692913386in"
height="3.126023622047244in"}

Fig-3: Review and Submit Page

![A screenshot of a document list AI-generated content may be
incorrect.](media/image8.png){width="3.7035487751531058in"
height="5.55532261592301in"}

### 5.3 Proposal Creation- Using File Upload

Fig-1: Proposal Creation using file upload. This option is for creation
of Proposals in bulk which can be processed in CPC Later for policy
issue. Payment, if done, using a combined cheque for multiple policy,
the details of initial premium payment can be processed using this
method.

![A screen shot of a computer screen AI-generated content may be
incorrect.](media/image13.png){width="5.647121609798775in"
height="6.402645450568679in"}

## **5. Test Case**

The test case for the policy issue feature is given below:

  ---------------------------------------------------------------------------------------------
  **Functionality**   **TC Number**      **Test Case   **Detail             **Expected Result**
                                         Name**        Description**        
  ------------------- ------------------ ------------- -------------------- -------------------
  Policy Issue at PO  TC_NB_PO_001       Index Docs    Verify successful    Documents indexed
                                                       indexing of proposal and visible to CPC
                                                       documents at PO      

  Policy Issue at PO  TC_NB_PO_002       Retrieve Docs Validate CPC can     CPC retrieves
                                                       retrieve indexed     documents without
                                                       documents            error

  Policy Issue at PO  TC_NB_PO_003       Mandatory     Ensure mandatory     All mandatory
                                         Fields        fields captured      fields saved
                                                       correctly            

  Policy Issue at PO  TC_NB_PO_004       Missing       Missing mandatory    System shows
                                         Fields        fields during CPC    validation error
                                                       entry                

  Policy Issue at PO  TC_NB_PO_005       Invalid DOB   Invalid date format  System rejects
                                                       for DOB              invalid date

  Policy Issue at PO  TC_NB_PO_006       Invalid Sum   Incorrect Sum        System shows error
                                                       Assured below/above  message
                                                       limits               

  Policy Issue at PO  TC_NB_PO_007       Duplicate     Duplicate proposal   System prevents
                                         Proposal      number entry         duplicate entry

  Policy Issue at PO  TC_NB_PO_008       Policy Number System               Policy number
                                         Gen           auto-generates       generated
                                                       Policy Number        successfully

  Policy Issue at PO  TC_NB_PO_009       KYC Missing   Attempt to issue     System blocks
                                                       policy without KYC   issuance

  Policy Issue at PO  TC_NB_PO_010       Full Flow     Complete flow:       Policy issued
                                         Success       Indexing at PO → CPC successfully with
                                                       data entry → Policy  valid Policy Number
                                                       issuance             

  Digital Policy      TC_NB_Aadhar_011   Aadhaar Auth  Verify Aadhaar       Authentication
  Issue for                                            authentication for   successful
  Non-Medical                                          Non-Medical policy   
  Policies                                                                  

  Digital Policy      TC_NB_Aadhar_013   Issue Policy  Validate successful  Policy issued
  Issue for                                            policy issuance      successfully
  Non-Medical                                          after Aadhaar        
  Policies                                             verification         

  Digital Policy      TC_NB_Aadhar_014   Auth Fail     Aadhaar              System shows
  Issue for                                            authentication fails authentication
  Non-Medical                                          (invalid OTP)        error
  Policies                                                                  

  Digital Policy      TC_NB_Aadhar_015   Details       Aadhaar details      System rejects
  Issue for                              Mismatch      mismatch with        proposal
  Non-Medical                                          proposal             
  Policies                                                                  

  Digital Policy      TC_NB_Aadhar_016   Premium Calc  Validate premium     Premium calculated
  Issue for                                            calculation for      correctly
  Non-Medical                                          Non-Medical policy   
  Policies                                                                  

  Digital Policy      TC_NB_Aadhar_017   Nominee       Issue policy without System shows error
  Issue for                              Missing       nominee details      
  Non-Medical                                                               
  Policies                                                                  

  Digital Policy      TC_NB_Aadhar_018   Age Restrict  Restrict Non-Medical System blocks
  Issue for                                            policy for age \> 50 issuance
  Non-Medical                                                               
  Policies                                                                  

  Digital Policy      TC_NB_Aadhar_019   Non-Medical   Aadhaar              Policy issued
  Issue for                              Success       authentication,      successfully
  Non-Medical                                          premium calculation, without medical
  Policies                                             payment, and         certificate
                                                       issuance             

  Digital Policy      TC_NB_Aadhar_020   Aadhaar Auth  Verify Aadhaar       Authentication
  Issue for Medical                      Med           authentication for   successful
  Policies                                             Medical policy       

  Digital Policy      TC_NB_Aadhar_021   Upload Cert   Validate medical     Certificate
  Issue for Medical                                    certificate upload   uploaded
  Policies                                             before issuance      successfully

  Digital Policy      TC_NB_Aadhar_022   Cert Missing  Missing medical      System blocks
  Issue for Medical                                    certificate for      issuance
  Policies                                             required policy      

  Digital Policy      TC_NB_Aadhar_023   Invalid File  Upload invalid file  System rejects file
  Issue for Medical                                    format for           
  Policies                                             certificate          

  Digital Policy      TC_NB_Aadhar_024   Premium Calc  Validate premium     Premium calculated
  Issue for Medical                      Med           calculation based on correctly
  Policies                                             medical status       

  Digital Policy      TC_NB_Aadhar_025   No Approval   Issue policy without System blocks
  Issue for Medical                                    medical approval     issuance
  Policies                                                                  

  Digital Policy      TC_NB_Aadhar_026   Cert Expired  Medical certificate  System rejects
  Issue for Medical                                    expired              certificate
  Policies                                                                  

  Digital Policy      TC_NB_Aadhar_027   Medical       Aadhaar              Policy issued
  Issue for Medical                      Success       authentication,      successfully after
  Policies                                             medical certificate  medical approval
                                                       upload, approval,    
                                                       payment, and         
                                                       issuance             

  Bulk Proposal Issue TC_NB_Bulk_028     Upload File   Verify successful    File uploaded
                                                       upload of bulk       successfully
                                                       proposal file        

  Bulk Proposal Issue TC_NB_Bulk_029     Process       Validate system      All proposals
                                         Proposals     processes all        processed
                                                       proposals            

  Bulk Proposal Issue TC_NB_Bulk_030     Missing       Upload file with     System shows error
                                         Columns       missing mandatory    report
                                                       columns              

  Bulk Proposal Issue TC_NB_Bulk_031     Invalid Data  Upload file with     System rejects
                                                       invalid data types   invalid rows

  Bulk Proposal Issue TC_NB_Bulk_032     Error Report  Validate system      Error report
                                                       generates error      generated
                                                       report for failed    
                                                       records              

  Bulk Proposal Issue TC_NB_Bulk_033     File Too      Upload file          System rejects file
                                         Large         exceeding max size   

  Bulk Proposal Issue TC_NB_Bulk_034     Partial       Validate partial     Valid proposals
                                         Success       success scenario     processed, invalid
                                                                            ones reported

  Bulk Proposal Issue TC_NB_Bulk_035     Bulk Success  Upload valid bulk    All valid proposals
                                                       file, process        converted to
                                                       proposals, generate  policies
                                                       policies             successfully
  ---------------------------------------------------------------------------------------------

## **6. Appendices**

The Following Documents attached below can be used.
