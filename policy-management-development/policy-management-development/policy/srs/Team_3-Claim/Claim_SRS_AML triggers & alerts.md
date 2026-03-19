[SRS/FRS on AML Alerts and triggers]{.underline}

Prepared by: Deepak Tripathi (IP Operations, PLI Directorate \| Version:
Initial \| Date: 13/10/2025. This Document has been prepared on the
personal capacity and as per role delegated vide PLI Directorate letter
No 56-14/2025-LI(IMS2.0) dated 07.10.25 / 10.10.25 & is subject to
change.

# **[Anti-Money Laundering (AML) Triggers & Alerts Module]{.underline}**

## 1. Introduction

The Anti-Money Laundering (AML) Triggers & Alerts module has been
developed to automate the process of identifying, flagging, and
escalating potentially suspicious transactions within the insurance
sector. This system is designed to operate strictly in line with
anti-money laundering regulations, with particular adherence to the
guidelines set forth by the Insurance Regulatory and Development
Authority of India (IRDAI) and the Financial Intelligence Unit-India
(FIU-India). Key functionalities of the AML module include:

- Generating alerts in response to activities deemed questionable.

- Assigning risk scores to incidents that are flagged.

- Facilitating the preparation and submission of reports through the
  Finnet interface.

- Maintaining a comprehensive audit trail to ensure traceability and
  accountability at all stages.

**Note 1**: This Software Requirement Specification (SRS)/FRS is linked
to the Archival Policy, which is currently under review. Alert triggers
will be updated in accordance with amendments to the Prevention of Money
Laundering Act (PMLA) or the introduction of new IRDAI norms/Acts.

**Note 2**: A separate SRS/FRS covering Finnet/Fingate functionalities
is also being developed. This document should be read in conjunction
with those additional specifications.

**Note 3**: Integration with NSDL and PAN is nearing completion.
Synchronisation with this SRS will be carried out once that process is
finalised.

Note 4 : As the PLI Directorate is in talks with the PMLA section
regarding **unique identifier** hence, any changes in this shall impact
this entire document accordingly.

## 1.1 Prevention of Money Laundering Act (PMLA) 2002

The Prevention of Money Laundering Act, 2002, which came into effect on
1 July 2005, governs all financial institutions, including insurers.
International regulators place strong emphasis on the implementation of
AML programmes within non-depository financial institutions, with
insurance companies specifically highlighted. The Financial Action Task
Force (FATF) and the International Association of Insurance Supervisors
(IAIS) Insurance Core Principles recommend robust AML programmes. In
alignment with these recommendations, the IRDAI also mandates that
insurers establish and maintain effective AML measures.

## 1.2 Objective

This document aims to:

- Define the AML framework for IMS 2.0.

- Prevent the misuse of insurance products for money laundering
  purposes.

- Ensure full compliance with all applicable AML laws and regulations.

- Safeguard the reputation of the organisation.

- Assist law enforcement agencies in the investigation of money
  laundering activities.

## 2. Money Laundering -- Definition

Money laundering refers to the process through which illegally obtained
funds are disguised to make them appear legitimate. This process
typically consists of three distinct stages, each of which can expose
insurance institutions to criminal activity during the course of regular
business operations:

a.  Placement: The physical introduction of illicit funds into the
    financial system.

b.  Layering: The use of complex transactions to obscure the origins of
    funds.

c.  Integration: The reintroduction of laundered funds into the economy
    in a manner that presents them as legitimate assets.

Section 3 of the PMLA defines money laundering as any direct or indirect
attempt to engage in activities connected with the proceeds of crime,
with the intention of projecting such funds as untainted property.

## 3. Money Laundering Risks

In the absence of a robust AML framework, the organisation is exposed to
several risks, including:

- Reputation Risk: Potential loss arising from diminished stakeholder
  trust.

- Compliance Risk: The possibility of incurring penalties due to
  regulatory failures.

- Operational Risk: Losses caused by inadequate processes or external
  events.

- Legal Risk: Exposure to fines, asset forfeiture, or loss of licence
  resulting from legal violations.

- Financial Risk: Direct monetary losses stemming from the
  aforementioned risks.

## 4. AML/KYC Standards

### 4.1Basic Due Diligence (KYC)

The organisation adheres to IRDAI and other regulatory guidelines by
implementing comprehensive Customer Due Diligence (CDD) procedures to
accurately identify all clients. These procedures include gathering all
required customer information, monitoring complex or unusually large
transactions, and conducting daily screenings against negative lists.
The department/PLI does not enter into any contracts with individuals
linked to criminal or terrorist organisations, and any matches
identified are promptly reported to the appropriate authorities. The
freezing or unfreezing of accounts in accordance with Section 51A of the
Unlawful Activities (Prevention) Act (UAPA) is carried out quickly and
confidentially. Beneficial ownership for non-individual customers is
also verified as part of these protocols.

### 4.2 KYC Timing

- KYC/CDD must be completed prior to commencing any new contract with a
  customer.

- Ongoing KYC reviews are conducted at the time of claim payout and
  whenever transactions are inconsistent or appear unusual.

Note : KYC norms also depends /related to proposed FRS for onboarding of
new customers so KYC timing may alter accordingly

### 4.3 Risk Profile Assessment

Given the scale and value of policies and transactions, the FRS proposes
a Risk-Based Approach (RBA) to customer due diligence. Enhanced scrutiny
is applied to customers, transaction types, and payment methods deemed
to be of higher risk. Regular risk assessments are conducted to guide
monitoring strategies, which are adapted as circumstances and threats
evolve. Customers are classified as either high or low risk based on a
range of factors, including category, occupation, geography, sourcing
channel, product features, payment modes, and reviews against sanctions
lists.

### 4.4 Enhanced Due Diligence

Enhanced Due Diligence involves the under mentioned practices which
needs to be practice as per the existing best insurance practices as
mentioned below:-

**A. Documenting Source of Funds & Net Worth**

- **Purpose:** To ensure that the money used for purchasing insurance
  policies is legitimate.

- **How:** Customers must provide proof of income and disclose their net
  worth. This is supported by collecting documents such as salary slips,
  bank statements, or tax returns. For this PAN integration needs to be
  functional.

**B. Restrictions on Payment Modes & Cash Limits**

- **Purpose:** To reduce the risk of money laundering through cash
  transactions.

- **How:** There are limits on how much cash can be accepted (in line
  with regulatory requirements), and certain payment modes may be
  restricted or require additional checks.

**C. Mandatory PAN & KYC for Third-Party Payments**

- **Purpose:** To prevent misuse of insurance for laundering money
  through third parties.

- **How:** Any payment made by someone other than the policyholder
  requires full PAN (Permanent Account Number) and KYC (Know Your
  Customer) documentation.

**D. Regulatory Reporting to FIU-IND (After integration with
FINNET/FINGATE)**

- **Purpose:** To assist authorities in detecting and investigating
  suspicious activities.

- **How:** The organization submits the following reports to the
  Financial Intelligence Unit -- India (FIU-IND):

  - **Suspicious Transaction Reports (STR):** For transactions suspected
    to involve money laundering.

  - **Cash Transaction Reports (CTR):** For large cash transactions
    above a regulatory threshold.

  - **Counterfeit Currency Reports (CCR):** For any detection of fake
    currency.

  - **Non-Profit Organisation Transactions Reports (NTR):** For
    transactions involving non-profit entities.

**(Note : The above norms/proposed norms are not fully being implemented
but proposed to be implemented in consultation with PMLA Division of
Postal Directorate)**

### 4.5 Record Keeping

Note :(This Section has been left open ended due to Archival Policy in
the process of Approval)

# 5. Functional Scope

  ---------------------------------------------------------------------------
  Functionality           Description
  ----------------------- ---------------------------------------------------
  Trigger                 Implements rule-based logic to systematically
                          identify suspicious transaction patterns within
                          insurance operations.

  Risk Scoring            Assigns a severity or risk level---Low, Medium,
                          High, or Critical---to detected transactions,
                          enabling prioritisation of further actions.

  Alert Dashboard Display Presents all flagged transactions in a
                          user-friendly interface, allowing authorised
                          personnel to review and take appropriate actions.

  Filing                  Prepares data for Suspicious Transaction Reports
  Interface/Integration   (STRs) and Cash Transaction Reports (CTRs),
                          ensuring seamless submission to Finnet or Fingate.

  Audit Trail             Maintains a detailed log of all alerts, user
                          actions, and interactions to support compliance and
                          internal audits.
  ---------------------------------------------------------------------------

# 6. Trigger Logic Definitions

  ------------------------------------------------------------------------------
  Trigger     Description   Condition                 Risk Level Action
  Code                                                           
  ----------- ------------- ------------------------- ---------- ---------------
  AML_001     High Cash     cash_amount \> ₹50,000    High       Alert generated
              Premium       (Premium deposit made in             along with CTR
                            cash more than Rs                    filing
                            50000/-)                             

  AML_002     PAN Mismatch  pan_verified = false      Medium     Alert generated
                                                                 for manual
                                                                 review

  AML_003     Nominee       nominee_change_date \>    Critical   Transaction
              Change Post   death_date                           blocked and STR
              Death                                              filed

  AML_004     Frequent      More than 3 surrenders    Medium     Alert generated
              Surrenders by within 6 months                      for further
              a single                                           investigation
              customer                                           
              profile.                                           

  AML_005     Refund        refund_date \<            High       Alert generated
              Without Bond  bond_dispatch_date                   and event
              Delivery                                           logged in the
                                                                 audit trail
  ------------------------------------------------------------------------------

# 7. Data Inputs & Sources : This table format shall help the system /developer to know what information to look at , where to find it. And what kind of data it is, so it can catch anything that doesnot look appropriate.

  -------------------------------------------------------------------------
  Field (What is it)     Source (From    Format     Remarks/Explaination
                         where this data            
                         can be                     
                         archived)                  
  ---------------------- --------------- ---------- -----------------------
  cash_amount            Payment Table   Numeric    The amount of cash paid
                                                    for an insurance
                                                    policy.

  pan_verified           PAN             Boolean    Whether the customer
                         Verification    (True or   PAN details are
                         API             false)     verified.

  tnominee_change_date   Nominee         Date       The date on which
                         Registry                   change in nominee
                                                    details has been made.

  death_date             Claims Table    Date       \-

  refund_date            Cancellation    Date       \-
                         Table                      

  bond_dispatch_date     CRM Dispatch    Date       \-
                         Table                      
  -------------------------------------------------------------------------

# 8. User Roles & Access Matrix

+-------------------------+-------------------------+-------------------------+
| Role                    | Access                  | Actions                 |
+:========================+:========================+:========================+
| Circle AML Nodal        | Full                    | View, flag, escalate,   |
| Officer/DDM concerned   |                         | and file reports        |
+-------------------------+-------------------------+-------------------------+
| CEPT Reviewer           | Partial                 | View alerts and provide |
|                         |                         | comments                |
+-------------------------+-------------------------+-------------------------+
| Audit Cell/At PLI       | Read-only               | View data and export    |
| Directorate             |                         | logs for audit purposes |
+-------------------------+-------------------------+-------------------------+
| Directorate             | Oversight               | View summary data and   |
|                         |                         | approve filings         |
+-------------------------+-------------------------+-------------------------+
| Note : User role delegation may be changed altered as per the requirement   |
| being made. In any case PLI Directorate be placed at the APEX level to      |
| ensure overall monitoring.                                                  |
+-----------------------------------------------------------------------------+

# 9. System Actions: This is the way system is expected to behave on getting triggered for a particular /defined alert.

+------------------------+---------------------------------------------+
| Trigger                | System Response                             |
+:=======================+:============================================+
| Match Found            | The system generates an alert for the       |
|                        | detected transaction.                       |
+------------------------+---------------------------------------------+
| Risk Score =           | Alert is escalated automatically to the AML |
| High/Critical          | Officer/Nodal Officer for immediate         |
|                        | attention.                                  |
+------------------------+---------------------------------------------+
| Filing Required        | Prepares STR/CTR data in JSON format for    |
|                        | submission.                                 |
|                        |                                             |
|                        | Note : File format be JSON due to           |
|                        | universally expected and computer friendly. |
+------------------------+---------------------------------------------+
| Action Taken           | The system records all actions in the audit |
|                        | trail to support traceability.              |
+------------------------+---------------------------------------------+

# 10. Audit Trail Fields: this means the pieces of information that the system records every time an important action or alert happens. Like detailed log book.

  -----------------------------------------------------------------------
  Field                               Description
  ----------------------------------- -----------------------------------
  alert_id                            A unique identifier assigned to
                                      each alert for tracking purposes.

  Timestamp                           Records the exact date and time
                                      when the trigger occurred.

  trigger_code                        Indicates the specific AML rule
                                      that was matched.

  action_taken                        Reflects the status of the alert
                                      (e.g., flagged, escalated, filed).

  user_id                             Identifies the officer who took
                                      action on the alert.

  filing_status                       Indicates whether the filing is
                                      completed, pending, or rejected.
  -----------------------------------------------------------------------

# 11. Integration Points

+-----------------------------------+-----------------------------------+
| System                            | Purpose                           |
+:==================================+:==================================+
| Finnet/Fingate                    | Facilitates filing of Suspicious  |
|                                   | Transaction Reports and Cash      |
|                                   | Transaction Reports.              |
+-----------------------------------+-----------------------------------+
| PAN Verification API              | Enables validation of customer    |
|                                   | identity through PAN              |
| (once developed/launched as this  | verification.                     |
| is still under process of         |                                   |
| integration in PLI)               |                                   |
+-----------------------------------+-----------------------------------+
| CRM                               | Tracks the dispatch of insurance  |
|                                   | bonds for reference in refund and |
| (There is separate SRS for        | delivery processes.               |
| integration of CRM module, which  |                                   |
| will cater to CRM separately and  |                                   |
| in detailed)                      |                                   |
+-----------------------------------+-----------------------------------+
| Policy DB (Data base)             | Provides access to the            |
|                                   | transaction history associated    |
|                                   | with insurance policies.          |
+-----------------------------------+-----------------------------------+

# 12. Wireframe -- AML Alert Dashboard

The AML Alert Dashboard shall offer a centralised platform for
monitoring, searching, and filtering alerts based on trigger type, risk
level, and date range. The dashboard table displays key fields,
including Alert ID, Policy Number, Trigger, Risk, Status, and available
Actions such as flagging, filing, or blocking transactions.

**Note : (The team leads have been instructed during the briefing that
Wireframe shall be developed by separate dedicated team, hence this
portion has been left out. In case of requirement this shall be handed
over separately as Annexure to SRS**)

# 13. Compliance Notes

This module operates independently of POLI Rules 2011. It adheres
strictly to the **IRDAI AML Master Guidelines issued in 2022, 2023, and
2024**, **the FIU-India STR/CTR schema**, and the **requirements under
PMLA Rule 9(1C) and 9(1D) pertaining to Know Your Customer (KYC)
procedures.**![](media/image1.png){width="1.968503937007874e-2in"
height="0.11811023622047244in"}

# ![](media/image2.png){width="1.968503937007874e-2in" height="0.11811023622047244in"} 
