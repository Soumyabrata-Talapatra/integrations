**DEPARTMENT OF POSTS**

**MINISTRY OF COMMUNICATIONS & IT**

**GOVERNMENT OF INDIA**

**[Customer ID Merging]{.underline}**

**1. Overview**

> This document defines the complete process for **Customer ID Merging**
> (combining multiple CIDs of the same person into one Primary/Global
> CID) and **Customer ID Unmerging** (reversing a previous merge by
> restoring secondary CIDs). Both functionalities apply to PLI and RPLI
> products.

**2. Brief Description**

- **Merging**: Consolidates multiple CIDs → one Primary CID. All
  policies move to Primary CID.

- **Unmerging**: Restores previously merged secondary CIDs as
  independent active CIDs again. Policies are moved back to their
  original CIDs.

**3. Customer ID Merging -- Eligibility & Business Rules (unchanged
summary)**

- Only insured roles eligible for Primary CID

- Policies must be active (no terminated status, no pending financial
  requests)

- No HUF/MWPA policies

- Mandatory document: Customer ID Merging Form

- All attributes of Primary CID override secondary CIDs

4\. Customer ID Merging -- Detailed Process Flow & UI Wireframes

4.1 Indexing Screen (Merge Request) -- Wireframe

+\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\--+

\| Service Request Indexing \|

+\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\--+

\| Request Type : \[Merge Customer ID ▼\] \|

\| Policy Number: \[\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\] Validate
Button \|

\| Date : 18-Nov-2025 (auto) \|

\| \[ \] For Joint Life → Radio: ( ) Insured 1 ( ) Insured 2 \|

\| \|

\| → Next Button \|

+\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\--+

After Next →

+\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\--+

\| Policy Summary \| Pending Requests \|

\| Insured : Primary Insured (or Secondary) \|

\| Name : XXXXXXXX \|

\| CID : C12345678 \|

\| \|

\| \[\*\] I confirm all attributes of Primary CID will override \|

\| secondary CIDs (DOB, Name, Address, Bank, etc.) \|

\| \|

\| Submit Cancel \|

+\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\--+

On Submit → Ticket ID: PSMCI000012345

#### 

#### 4.2 Data Entry / QC Screen (Merge) -- Wireframe

+\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\--+

\| Ticket ID: PSMCI000012345 \| View Documents Button \|

+\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\--+

\| Primary Customer Details (Read-only grid) \|

\| CID \| Name \| DOB \| Role \|

\| C12345678 \| RAHUL KUMAR \| 01/01/1990\| Primary Insured \|

+\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\--+

\| Search Customer to Merge \|

\| ( ) Customer ID \[\_\_\_\_\_\_\_\_\_\_\_\_\] ( ) Policy Number
\[\_\_\_\_\_\_\]\|

\| Continue Button \|

+\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\--+

\| Search Results → Select Customers to be Merged (grid) \|

\| \[Add\] CID \| Name \| DOB \| Matching Score \|

\| C87654321 \| RAHUL KUMAR\| 01/01/1990\| 100% (Aadhaar) \|

+\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\--+

\| Merged Customers (grid) \|

\| CID \| Name \| Policies \| \[Delete Icon\] \|

\| C87654321 \| RAHUL KUMAR \| P001,P002\| \[X\] \|

\| C11223344 \| RAHUL KUMAR \| P003 \| \[X\] \|

+\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\--+

\| Add Comments \| Request Missing Docs \| Submit \| Cancel \|

+\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\--+

#### 4.3 Approver Screen (Merge) -- Wireframe

Same layout as DE/QC but all fields read-only except Approve / Reject /
Redirect / Request Missing Docs buttons.

**5. Customer ID Unmerging (De-merging) -- New Functionality**

- **Purpose**

> To reverse an erroneous merge. Only possible if:

- Merge request was approved within the last 180 days (configurable)

- No new policies issued on Primary CID after merge

- No financial transactions (loan, partial maturity, etc.) performed on
  merged policies after merge

- No change in critical attributes (DOB, Name via CSI, etc.) after merge

- **Request Type**

> New service request type: **"Unmerge Customer ID"** (or "Reverse
> Customer ID Merge")

- **Eligibility Checks (system enforced)**

1.  Original merge request must be in Approved status

2.  Time since approval ≤ 180 days

3.  No new policy linked to Primary CID post merge

4.  No financial requests processed post merge

5.  Primary CID has not been merged again into another CID

6.  No death/maturity claim processed on any policy

> If any check fails → Error: "Unmerging not possible due to \[reason\]"

- **Roles Allowed to Index Unmerge**

> Only Supervisory/Administrative roles (CPC Supervisor, Divisional
> Head, etc.)

- **Documents Required for Unmerge**

<!-- -->

- Unmerge Request Form (new template) signed by customer or with valid
  reason by office

- Original Merge Request Form copy (auto-attached by system)

**6. Customer ID Unmerging -- Detailed Process Flow & UI Wireframes**

**6.1 Indexing Screen (Unmerge Request)**

> text
>
> +\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\--+
>
> \| Service Request Indexing \|
>
> +\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\--+
>
> \| Request Type : \[Unmerge Customer ID ▼\] \|
>
> \| Primary CID : \[C12345678\_\_\_\_\_\_\_\_\_\_\_\] Validate \|
>
> \| OR Policy No : \[\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\] \|
>
> \| Date : 18-Nov-2025 \|
>
> \| \|
>
> \| → Next \|
>
> +\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\--+
>
> After validation →
>
> \| Merged History (auto-displayed) \|
>
> \| Original Merge Ticket : PSMCI000012345 (18-May-2025) \|
>
> \| Secondary CIDs : C87654321, C11223344 \|
>
> \| Policies moved : P001,P002,P003 \|
>
> \| \|
>
> \| \[ \] I confirm unmerge is required and eligibility met \|
>
> \| \|
>
> \| Submit → Ticket ID: PSUMCI0000009876 \|
>
> +\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\--+

- **7.2 Data Entry / QC Screen (Unmerge)**

> text
>
> +\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\--+
>
> \| Ticket ID: PSUMCI0000009876 \|
>
> +\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\--+
>
> \| Primary CID : C12345678 Name: RAHUL KUMAR \|
>
> \| Merge Date : 18-May-2025 Merge Ticket: PSMCI000012345 \|
>
> +\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\--+
>
> \| Secondary CIDs to Restore (system auto-populates) \|
>
> \| \[ \] C87654321 → Policies: P001,P002 → Restore as Primary
> Insured\|
>
> \| \[ \] C11223344 → Policy : P003 → Restore as Primary Insured\|
>
> \| \|
>
> \| Reason for Unmerge : \[Dropdown + Free Text\] \|
>
> \| - Data Entry Error - Customer Request - Other
> \_\_\_\_\_\_\_\_\_\_\_\_\|
>
> \| \|
>
> \| View Original Merge Documents \| View Unmerge Form \|
>
> +\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\--+
>
> \| Submit \| Cancel \| Request Missing Docs \|
>
> +\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\--+

**6.3 Approver Screen (Unmerge)**

> Read-only view + buttons: **Approve Unmerge** → All secondary CIDs
> restored as active, policies moved back, Primary CID retains only its
> original policies. **Reject** → Merge remains intact.
>
> On successful unmerge:

- Secondary CIDs become active again with their original attributes
  restored from audit log

- Policies re-tagged to original CIDs

- Audit trail updated: "Unmerged via ticket PSUMCI0000009876"

**7. Workflow Diagrams**

- **Merge Workflow**

> Indexer → Scan Docs → DE (Search & Add Secondary CIDs)
>
> ↓
>
> QC (Verify & Modify if needed)
>
> ↓
>
> Approver → Approve → Merge Executed
>
> → Reject/Redirect

- **Unmerge Workflow**

> Supervisor → Index Unmerge Request (Primary CID)
>
> ↓
>
> System Auto-check Eligibility
>
> ↓ (if passed)
>
> DE/QC → Verify Reason & Documents
>
> ↓
>
> Approver → Approve Unmerge → CIDs & Policies Restored
>
> → Reject

- **9. Changes After Merging / Unmerging**

+--------------+--------------+---------------+--------------------+
| > **Action** | > **Primary  | > **Secondary | > **Policies**     |
|              | > CID**      | > CID(s)**    |                    |
+:=============+:=============+:==============+:===================+
| > Merge      | > Remains    | > Deactivated | > All moved to     |
|              | > active     |               | > Primary          |
+--------------+--------------+---------------+--------------------+
| > Unmerge    | > Keeps      | > Restored as | > Moved back to    |
|              | > original   | > active      | > original CIDs    |
|              | > only       |               |                    |
+--------------+--------------+---------------+--------------------+

**8. Customer Portal Behaviour**

- After merge: Login with old CID → "Your CID has been merged to Global
  CID XXXXXXXX"

- After unmerge: Old CIDs become valid for login again.

9\. **Extension of CID Merge/Unmerge facility to:**

1.  **New Insurance Solution (McCamish/NIS)** -- During New Proposal
    Onboarding (Department users)

2.  **New Self-Service Web Portal / Customer Portal** --
    Customer-initiated Merge/Unmerge

3.  **1. Merge/Unmerge during New Proposal Onboarding (Department User
    -- CPC/PO Staff)**

4.  **Scenario**

Customer comes for a new PLI/RPLI policy. During proposal data entry,
system detects that the prospect already has one or more existing CIDs
(based on Aadhaar/PAN/DOB/Name match).

**10. Process Flow**

1.  User enters prospect details (Name, DOB, Aadhaar, PAN, Mobile, etc.)

2.  System performs real-time duplicate check

3.  If existing CID(s) found → "Possible Duplicate Customer" pop-up with
    matching CIDs

4.  User can choose:

    - Create New CID (default)

    - Merge with one of the existing CIDs (Primary CID selected by user)

    - Unmerge (only if a previous erroneous merge is detected)

**11. Wireframe -- Proposal Data Entry Screen (New Insurance Solution)**

+\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\--+

\| New Proposal Entry - PLI/RPLI \|

+\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\--+

\| Name : \[RAHUL KUMAR \] \|

\| DOB : \[01/01/1990 \] \|

\| Aadhaar : \[1234 5678 9012 \] \|

\| PAN : \[ABCDE1234F \] \|

\| Mobile : \[9876543210 \] \|

\| \... \|

\| \[Save & Continue\] \|

+\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\--+

→ System Duplicate Check Triggered

+\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\--+

\| ⚠ Possible Duplicate Customer Found \|

+\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\--+

\| Matching Existing Customer IDs: \|

\| \[ \] CID: C12345678 \| Name: RAHUL KUMAR \| DOB: 01/01/1990 \|
Match: Aadhaar + PAN \|

\| \[ \] CID: C87654321 \| Name: RAHUL K \| DOB: 01/01/1990 \| Match:
Aadhaar \|

\| \|

\| Options: \|

\| ( ) Create New Customer ID \|

\| ( ) Merge into Existing CID → \[Select from above ▼\] \|

\| ( ) Unmerge previous merge (only if merged \<180 days) \|

\| \|

\| Reason (mandatory if Merge/Unmerge):
\[\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\] \|

\| \|

\| \[Proceed\] \[Cancel\] \|

+\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\--+

If Merge selected → Mini-merge workflow (same validations as regular
merge)

→ On approval, new policy tagged to selected Primary CID

If Unmerge selected → System checks eligibility → restores secondary CID
→ new policy tagged to restored CID

**11. Approval Requirement**

- Merge/Unmerge during proposal → Auto-approved if match score = 100%
  (Aadhaar/PAN + DOB)

- Otherwise → Goes to Supervisor inbox for approval (same as regular
  merge ticket)

12\. **Customer Self-Service Portal -- CID Merge/Unmerge (Customer
Initiated)**

**Allowed Only When**

- Customer logs in with registered mobile/email

- Exact match on at least **two** of the following:

  - Aadhaar / PAN

  - DOB

  - Father's Name / Spouse Name

  - Registered Mobile

- OTP validation mandatory on registered mobile/email

**13. Restrictions**

- Customer can only merge their own CIDs (system detects ownership via
  KYC/mobile)

- Unmerge allowed only within 90 days of merge

- No financial transactions post-merge

- Maximum 3 CIDs can be merged in one request

**14. Wireframe -- Customer Portal (Self-Service)**

**Login → My Profile → Manage Customer IDs**

text

+\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\--+

\| My Profile → Manage Customer IDs \|

+\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\--+

\| Your Active Customer IDs \|

\| • Global CID: C12345678 (3 policies) \|

\| • Merged CID: C87654321 (inactive - merged on 10-Oct-2025) \|

+\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\--+

\| Merge Another Customer ID \|

\| Enter CID or Policy Number: \[\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\] \|

\| \[Search\] \|

+\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\--+

After Search (if match found)

\| Found Possible Match \|

\| CID: C87654321 \| Name: RAHUL KUMAR \| DOB: 01/01/1990 \|

\| Matching Criteria: ✓ Aadhaar ✓ DOB ✓ Mobile \|

\| \|

\| \[Merge into my Primary CID (C12345678)\] \|

\| \|

\| OTP sent to +91-9876543210 \|

\| Enter OTP: \[\_\_\_\_\_\_\] \|

\| \|

\| \[Submit Merge Request\] \|

+\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\--+

Success → "Merge request submitted. Will be processed within 2 working
days."

**15. Unmerge Option (Only if merged \<90 days)**

+\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\--+

\| Merged Customer IDs \|

\| CID: C87654321 → Merged into C12345678 on 10-Oct-2025 \|

\| \[Request Unmerge\] \|

+\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\--+

\| Reason: ( ) Mistake in merging ( ) Other \[\_\_\_\_\_\_\_\_\_\] \|

\| OTP sent to registered mobile \|

\| Enter OTP: \[\_\_\_\_\_\_\] \|

\| \[Submit Unmerge Request\] \|

+\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\-\--+

**16. Backend Processing (Self-Service Requests)**

- All customer-initiated merge/unmerge requests → Create ticket (PSMCI /
  PSUMCI)

- Auto-approved if:

  - 100% KYC match + OTP validated

  - No financial activity post-merge

- Otherwise → Routed to CPC Supervisor for manual approval

- Customer notified via SMS/Email/Portal on completion

**17. Updated Workflow Summary (Including New Channels)**

1\. Regular Office Request → Indexer → DE → QC → Approver

2\. During New Proposal → Proposal Entry Staff → Auto-detect →
Mini-merge → Auto/Supervisor Approve

3\. Customer Portal → Customer → OTP + Matching Criteria → Auto-backend
ticket → Auto or Supervisor Approve

**18. Updated Eligibility Matrix**

  -------------------------------------------------------------------------------------
  **Channel**   **Merge     **Unmerge   **OTP        **Min Matching        **Approval
                Allowed**   Allowed**   Required**   Criteria**            Needed**
  ------------- ----------- ----------- ------------ --------------------- ------------
  Service       Yes         Yes (\<180  No           Full system rules     Always
  Request                   days)                                          (Approver)
  (Regular)                                                                

  New Proposal  Yes         Yes (\<180  No           Aadhaar/PAN +         Always
  Onboarding                days)                    DOB+Father name       (Approver)

  Customer      Yes         Yes (\<90   Yes          3 out of              Auto
  Portal                    days)                    (Aadhaar/PAN, DOB,    
                                                     Mobile, Father Name)  
  -------------------------------------------------------------------------------------
