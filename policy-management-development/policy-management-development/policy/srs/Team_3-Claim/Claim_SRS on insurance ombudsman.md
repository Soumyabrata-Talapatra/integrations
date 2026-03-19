Note :This SRS is contingent upon the formal notification of a
PLI-specific Ombudsman framework via Gazette or executive order,
modelled on IRDAI's Insurance Ombudsman Rules, 2017."

Prepared by: Deepak Tripathi (IP Operations, PLI Directorate \| Version:
Initial \| Date: 15/10/2025. This Document has been prepared on the
personal capacity and as per role delegated vide PLI Directorate letter
No 56-14/2025-LI(IMS2.0) dated 07.10.25 / 10.10.25 & is subject to
change. The Ideas/Matrix/Work flow has been copied from the Insurance
Ombudsman Rules 2017 and similar prevalent best practices

# 

# Module Overview

The Insurance Ombudsman Complaints Management module in IMS 2.0 enables
implementation of the updated Insurance Ombudsman Rules, 2017, within a
digital complaints system.

## 2. Limited Scope within the PLI Context

I.  **Not Under IRDAI Jurisdiction:** Postal Life Insurance (PLI) falls
    under the governance of the Department of Posts (DoP) within the
    Ministry of Communications, rather than the Insurance Regulatory and
    Development Authority of India (IRDAI). Consequently, the Insurance
    Ombudsman Rules, 2017---issued under Section 24 of the IRDAI
    Act---are not applicable to PLI.

II. **Absence of Statutory Provision for Ombudsman:** PLI is
    administered via executive orders and departmental regulations, as
    opposed to a dedicated legislative act such as that governing IRDAI.
    As a result, there is currently no statutory provision for
    establishing a quasi-judicial redressal authority, such as an
    Ombudsman.

III. **Administrative Grievance Redressal Channels:** Grievances
     concerning PLI are addressed through administrative measures,
     [including Circle grievance cells]{.mark}, [Directorate
     escalation]{.mark} mechanisms, [CPGRAMS]{.mark}, and the [RTI
     Act.]{.mark} These channels operate on an administrative basis and
     do not provide **adjudicatory or binding resolutions**.

IV. **No Mechanism for Binding Awards:** Unlike the IRDAI Ombudsman
    (Rule 17), PLI lacks an established system for conferring binding
    awards that would be enforceable upon the department.

V.  **Constraints due to Service Rules and Role Creation:** [The
    creation of an Ombudsman position within PLI requires further
    approval from the other ministerial channel, appropriate
    classification, and alignment with existing service rules.]{.mark}

## 3.Functional Scope : 

**[However, in order to comply the SRS demands and requirements , we
shall hypothetically consider PLI to be adhering Ombudsman Rules of its
own (similar in line with the best industry practice]{.underline}**) The
Complaints Management module is designed to comply with all statutory,
procedural, and operational requirements as stipulated by the [Insurance
Ombudsman Rules, 2017]{.mark}. The high-level functional capabilities
required include:

**Complaint Intake**: Multichannel registration (web, mobile, email,
offline, walk-in) supporting digital, written, and assisted modes,
compliant with Rule 14 of insurance ombudsman rules.

**Jurisdiction Mapping**: Dynamic mapping of complaints to territorial
ombudsman centers based on location and insurer/broker data, as per
existing system (to be defined in PLI context)

**Document Management**: Secure upload, retrieval, and management of
supporting documents and evidence.

**Hearing Scheduling**: End-to-end management of ombudsman hearings
(including video hearings), calendar management, automated
notifications, and conflict detection.

**Award Issuance**: Workflow for draft, review, approval, digital
signing, and communication of mediation recommendations and final
awards, with support for compensation calculations and regulatory caps.

**Audit Trail**: Comprehensive, non-tamperable logs of all complaint
actions and system events as per IT Act 2000,

**User & Role Management**: Granular access matrix supporting role-based
permissions (complainant, ombudsman, support staff, insurer, auditor,
admin).

**Communication & Notification**: Automated status updates to parties
via email/SMS/in-app notifications, with configurable templates.

**Escalation & Monitoring**: Automated escalation logic for breaches of
statutory timelines and oversight reporting for regulatory compliance,
analytics, and CEPT review.

**Reporting:** Real-time and scheduled analytics, including SLA
adherence, complaint types, closure rates, satisfaction metrics, and
compliance reports, with export features for regulatory reporting.

**Bilingual Support**: All user-facing fields and regulatory documents
must be available in English and Hindi, with adaptability for regional
language extensions.

**Integration Points**: APIs for connectivity with IRDAI's IGMS/Bima
Bharosa, insurer systems, government portals, and cross-module IMS 2.0
services. (also mentioned separately in last section of this document)

## 5. Complaint Lifecycle Flow

### Overview

The lifecycle of an insurance ombudsman complaint is strictly governed
by statutory provisions under the Insurance Ombudsman Rules, 2017:

**Initiation**: The complainant submits a detailed representation to the
insurer/broker and subsequently approaches the ombudsman if unsatisfied
or after a prescribed period.

**Admissibility & Assignment**: System determines eligibility (Rule 14)
and assigns the case to the relevant jurisdiction center (Rule 11).

**Intake Validation**: Collection and verification of all necessary
personal, policy, and dispute data, and essential supporting documents.

**Triage**: Screening for conflict of interest, duplicate/parallel
litigation, and statutory limitations (e.g., limitation period, claim
value cap).

**Documentation**: Further information and documents may be requested by
ombudsman, with strict tracking and reminders.

**Hearing Scheduling**: System facilitates the scheduling of
physical/video hearings, with logistics and automated notifications.

[Disposition:]{.mark}

**Mediation (Rule 16):** If settled, mediation recommendation is issued;
if accepted, a communication of final settlement is obtained.

**Adjudication (Rule 17):** If unresolved by mediation, the ombudsman
passes a binding written award.

**Award Communication**: System digitally signs and issues awards,
tracks insurer/broker compliance (mandatory within 30 days), and manages
any further escalations.

**Closure**: Once the award is fulfilled (or non-compliance escalates
for regulatory intervention), the complaint is closed and archived, with
audit logs preserved for mandated durations.

**Review & Analytics**: Complaint records feed into periodic statutory
and management review reporting, supporting operational improvement and
regulatory oversight.

**[Complaint Lifecycle Diagram (Textual)]{.underline}**

Flow chart has been illustrated as under for easy understanding:

Complaint Entry [⇒]{.mark} Eligibility Check ⇒ Jurisdiction Mapping ⇒
Intake Review ⇒ Registration & Case ID assignment ⇒ Document Upload ⇒
Assignment to Ombudsman/Staff ⇒ Preliminary Scrutiny ⇒ Requisition for
Info (if any) ⇒ Hearing Schedule Initiation ⇒ (Mediation Track) ⇒
Mediation Outcome Recording ⇒ (If not settled: Award Track) ⇒ Award
Drafting ⇒ Digital Approval & Signing ⇒ Award Dispatch & Compliance
Timer ⇒ Insurer/Broker Action Logging ⇒ Closure OR Escalation (in case
of non-compliance/appeal).

Every major transition is logged in the audit trail with full details
(user ID, timestamp, before/after data, IP address, etc.).

## 6. Data Inputs and Fields

Robust data structures, compliant with IRDAI and statutory norms, are
vital. All primary data inputs and their validation requirements for the
module are outlined below.

A. Complaint Registration:

Complainant Name (as per ID)

Contact Info (address, mobile, email)

Role (policyholder, nominee, legal heir, assignee)

Language Preference (English/Hindi)

Identification (Aadhaar/PAN/Passport/Other)

Policy Number / Claim Number

Agent name/Type & Branch/Office/HO/SO

Type of Policy (PLI/RPLI)

Date(s) of Incident, Representation to Insurer

Description of Issue (reason for complaint, all relevant facts)

Relief Sought/Remedy Sought (compensation, policy servicing, etc.)

B. Jurisdiction Mapping:

Pin Code/Digital pin code/Location of Complainant

Agent Details

Auto-mapped to ombudsman centers using jurisdiction master

C. Attachments:

Scanned Documents (policy, correspondence, denial letters, ID proof,
receipts, bills, any other evidence)

Supported formats: PDF, JPG, PNG, max file size per doc (usually 10 MB)

D. Complaint Category/Type:

Claim Delay cause (Life)

Partial/Full Repudiation

Premium Dispute

Policy Misrepresentation

Non-Issuance of Policy after premium

Policy Servicing Grievance

### E. Hearing and Mediation Fields:

Hearing Dates (requested/scheduled)

Mode (physical/video-conference)

Parties' availability

Consent to Mediation

F. Award & Resolution:

Award Type (mediation/adjudication)

Amount Awarded (with supporting calculation, statutory caps, interest as
applicable)

Award Reasons/Justification

Digital Signature of Ombudsman

Date of Issue

### G. Audit & System Fields:

Complaint Status (Registered, In Review, Hearing Scheduled, Mediation,
Award Issued, Awaiting Compliance, Closed)

SLA Timers (acknowledgement, resolution deadlines)

User/Role Activity Tracker

System-Generated Notifications

### 7.User Roles and Access Matrix

A role-based access control (RBAC) scheme is essential to ensure data
security, confidentiality, and process integrity.

Primary User Roles

+---------------------------+---------------------------------+--------------------------+
| **Role**                  | **Key Activities/Access**       | **Typical Users**        |
+:==========================+:================================+:=========================+
| [Complainant]{.mark}      | Register/view/track complaint,  | Policyholder/public      |
|                           | upload docs                     |                          |
+---------------------------+---------------------------------+--------------------------+
| [Agent/ Rep.]{.mark}      | View/respond to complaint,      | Insurance Agents         |
|                           | upload replies, comply          |                          |
+---------------------------+---------------------------------+--------------------------+
| [Ombudsman]{.mark}        | Review, assign, mediate,        | DPS rank Officers,       |
|                           | adjudicate, issue award         |                          |
+---------------------------+---------------------------------+--------------------------+
| [System Admin]{.mark}     | User mgmt, configuration,       | IMS operations/admins    |
|                           | override, reporting             |                          |
+---------------------------+---------------------------------+--------------------------+
| [Auditor/Reviewer]{.mark} | Audit trails, reports, system   | Internal/external        |
|                           | logs                            | auditors                 |
+---------------------------+---------------------------------+--------------------------+
| **[Note : This RBAC is only imaginary as per the existing best practice and may vary   |
| in the PLI Context]{.underline}**                                                      |
+----------------------------------------------------------------------------------------+

### 8.Access Matrix Illustration

+--------------------+-----------------+----------------+----------------+----------------+----------------+
| **Functionality**  | **Complainant** | **Agent**      | **Ombudsman**  | **Staff**      | **Admin**      |
+====================+=================+================+================+================+================+
| [Lodge             | Yes             | View           | View           | Create         | Override       |
| Complaint]{.mark}  |                 |                |                |                |                |
+--------------------+-----------------+----------------+----------------+----------------+----------------+
| [Upload            | Yes             | Yes            | Yes            | Yes            | Yes            |
| Docs]{.mark}       |                 |                |                |                |                |
+--------------------+-----------------+----------------+----------------+----------------+----------------+
| [Assign/Manage     | \-              | \-             | Yes            | Yes            | Yes            |
| Cases]{.mark}      |                 |                |                |                |                |
+--------------------+-----------------+----------------+----------------+----------------+----------------+
| [Hearing           | View info       | View           | Yes            | Yes            | Yes            |
| Scheduling]{.mark} |                 |                |                |                |                |
+--------------------+-----------------+----------------+----------------+----------------+----------------+
| [Award Draft &     | Notify          | Receive        | Yes            | Assist         | Yes            |
| Issue]{.mark}      |                 |                |                |                |                |
+--------------------+-----------------+----------------+----------------+----------------+----------------+
| [Edit              | Until locked    | Response only  | Yes            | Yes            | Yes            |
| Complaint]{.mark}  |                 |                |                |                |                |
+--------------------+-----------------+----------------+----------------+----------------+----------------+
| [Track             | Yes             | Yes            | Yes            | Yes            | Yes            |
| Status]{.mark}     |                 |                |                |                |                |
+--------------------+-----------------+----------------+----------------+----------------+----------------+
| [Run               | \-              | \-             | Yes            | Yes            | Yes            |
| Reports]{.mark}    |                 |                |                |                |                |
+--------------------+-----------------+----------------+----------------+----------------+----------------+
| [Audit Trail       | \-              | \-             | \-             | \-             | Yes            |
| View]{.mark}       |                 |                |                |                |                |
+--------------------+-----------------+----------------+----------------+----------------+----------------+
| [System Admin /    | \-              | \-             | \-             | \-             | Yes            |
| Config]{.mark}     |                 |                |                |                |                |
+--------------------+-----------------+----------------+----------------+----------------+----------------+
| [Bilingual Field   | Yes             | Yes            | Yes            | Yes            | Yes            |
| Controls]{.mark}   |                 |                |                |                |                |
+--------------------+-----------------+----------------+----------------+----------------+----------------+
| **[Note : This table is only imaginary as per the existing best practice and may vary in the PLI         |
| Context]{.underline}**                                                                                   |
+----------------------------------------------------------------------------------------------------------+

.

### 9.**Integration Note: Complaints Management Module -- PLI Context**

**Current Channels of Complaint Receipt**\
Postal Life Insurance (PLI) currently handles complaints through
multiple intake channels:

- **CPGRAMS** (Centralized Public Grievance Redress and Monitoring
  System)

- **Email / Physical Submissions** at Circle/Divisional Offices

- **RTI Requests** (linked to grievance disclosures)

- **CRM Portal** -- recently adopted by the Department for centralized
  tracking

**CRM Integration (Futuristic Scope)**\
The proposed *Complaints Management Module* under IMS 2.0 shall be
designed to:

- **Integrate with CRM modules**

- **Enable API-based handshake/Integration** with CPGRAMS and CRM
