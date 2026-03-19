> INTERNAL APPROVAL FORM

**Project Name: Collection**

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

[**2. Scope** [4](#scope)](#scope)

[**4. Business Requirements**
[4](#business-requirements)](#business-requirements)

[**5. Functional Requirements Specification**
[5](#functional-requirements-specification)](#functional-requirements-specification)

[5.1 Initial Premium Collection
[5](#initial-premium-collection)](#initial-premium-collection)

[5.2 Other Individual Collection
[7](#other-individual-collection)](#other-individual-collection)

[5.3 Special Group Collection
[12](#special-group-collection)](#special-group-collection)

[**6. Test Cases** [12](#test-cases)](#test-cases)

[**6. Appendices** [16](#appendices)](#appendices)

## **1. Executive Summary**

Define functional requirements for Collections lifecycle across PLI
products, enabling premium and related financial transactions, receipt
lifecycle, reversals, and reporting, while ensuring compliance with
India Post and insurance accounting controls.

## **2. Scope**

**The following is in scope:**

- Initial Collection

- Renewal Collection

- Reinstatement Collection

- Revival Lumpsum Collection

- Revival Installment Collection

- Full Loan Repayment

- Partial Loan Repayment

- Policy Miscellaneous Collection

- Agent Miscellaneous Collection

- Special Group Collection

**Other features that should also be included:**

- (F1) Same-day Receipt Cancellation

- (F2) Duplicate Receipt Generation

## **4. Business Requirements**

  -----------------------------------------------------------------------
  Requirement   Requirement
  ID            
  ------------- ---------------------------------------------------------
  FS-COL-001    Enable Initial Premium Collection based on Proposal No or
                Mobile No.

  FS-COL-002    Enable Renewal Premium Collections with interest/arrears
                calculation.

  FS-COL-003    Support Reinstatement & Revival (Lump Sum / Installment).

  FS-COL-004    Support Full & Partial Loan Repayments with real-time
                loan balance updates.

  FS-COL-005    Support Policy & Agent Miscellaneous Collections.

  FS-COL-006    Enable Special Group / Bulk Collections for payroll
                societies.

  FS-COL-007    Support multiple payment modes (cash, UPI, POS, cheque,
                DD, online).

  FS-COL-008    Generate unique digital receipt for all collections.

  FS-COL-009    Allow Same Day Receipt Cancellation with Supervisor
                approval.

  FS-COL-010    Support Duplicate Receipt Generation.

  FS-COL-011    Validate policy status, loan status, premium due rules.

  FS-COL-012    Provide audit trail and transaction logs.

  FS-COL-013    Enable search by Proposal No, Policy No, or Mobile
                Number.

  FS-COL-014    Integrate with future digital payment platforms (BBPS,
                IPPB, UPI autopay).
  -----------------------------------------------------------------------

## **5. Functional Requirements Specification**

### 5.1 Initial Premium Collection

Collect the first premium against a proposal pending first premium and
generate the receipt. Payment option for new proposal will get displayed
once the user input other details in new business page and proceed.
However, User should also get the option to search the proposal and
perform the payment.

**Workflow for Initial Premium Payment:**

![](media/image1.png){width="6.268055555555556in"
height="4.178472222222222in"}

#### Search Section:

- Search By: Radio (Proposal/Mobile): Mandatory

- Proposal No: Text: Required if Proposal selected

- Mobile No: Numeric: 10-digit validation

#### Proposal details:

- Proposal No: Auto-populated after search

- Customer Name: Auto

- Mobile Number: Auto

- Sum Assured: Auto

- Plan Type: Auto

- First Premium Amount: Auto

#### Payment Details

- Payment Mode: mandatory

- Instrument Number: Mandatory for cheque/DD

- Bank Name: Mandatory for cheque/DD

- Amount: Auto-Populated

- Date: Auto (Transaction Date)

#### Receipt Actions

- Generate Receipt

- Preview Receipt

- Print/Download Receipt

- Send SMS/Email

**Wireframe for the Initial Collection:**

![A screenshot of a web page AI-generated content may be
incorrect.](media/image2.png){width="3.5625in" height="5.34375in"}

### 5.2 Other Individual Collection

A single unified page to perform Renewal, Reinstatement, Revival (Lump
Sum & Installment), Loan Repayment (Full/Partial), Policy Miscellaneous,
and Agent Miscellaneous collections. The page uses policy validation to
render type-specific panels dynamically while keeping one header, one
payment section, and one receipt action area.

**Page Layout (Top to Bottom)**

1.  **Header -- Policy Context & Validation**

2.  **Collection Type Selector** (Dropdown)

3.  **Dynamic Collection Panel** (fields depend on selected type)

4.  **Payment Details** (common component)

5.  **Receipt Preview & Actions** (generate/print/email/SMS)

6.  **Audit & Notes** (common component)

**Wireframe for the Collection Screen:**

![](media/image3.png){width="3.2031988188976377in"
height="4.804796587926509in"}

#### 5.2.1 Policy Context and Validation

Provides the essential policy context required for any collection. The
Validate action performs real-time checks with the Insurance Management
System (IMS) to load eligibility, schedules, and agent mapping needed
for downstream collection processing.

**Fields:**

1.  Policy: Text (Searchable): Option should also be present to search
    and select a policy.

2.  Validate (Button): Calls policy validation API. It loads the
    following information for the policy:

    a.  Customer Name

    b.  Product Plan

    c.  Policy Status

    d.  Premium Mode

    e.  Next Premium Due Date

    f.  Loan Outstanding

    g.  Agent Code

    h.  View Ledger / History (Button): Opens modal: last 10
        transactions, loan ledger entries.

#### 5.2.2 Collection Type Selector

A single dropdown to choose the collection type---Renewal,
Reinstatement, Revival Lump Sum, Revival Installment, Full Loan
Repayment, Partial Loan Repayment, Policy Miscellaneous, or Agent
Miscellaneous. Selecting a type should dynamically configures the page
with the correct fields, validations, and rules, ensuring a streamlined
workflow without navigating to different pages.

**Fields:**

- Collection Type Selector: Dropdown:

  - Values are as follows:

    - RENEWAL -- Renewal Collection

    - REINSTATEMENT -- Reinstatement Collection

    - REVIVAL_LUMP_SUM -- Revival Lump Sum

    - REVIVAL_INSTALLMENT -- Revival Installment

    - LOAN_FULL -- Full Loan Repayment

    - LOAN_PARTIAL -- Partial Loan Repayment

    - POLICY_MISC -- Policy Miscellaneous Collection

    - AGENT_MISC -- Agent Miscellaneous Collection

- Selecting the Dropdown Renders the corresponding Dynamic Collection
  Panel below.

#### 

#### 5.2.3 Dynamic Collection Panel

Context-sensitive panel that displays the fields and computed amounts
for the chosen collection type.

A.  Renewal Collection

> **Fields:**

a.  Current Installment Amount: Read-only text

b.  Late Fee: Read-Only text

c.  GST: Read-Only text

d.  Arrears: Read-Only text

e.  Advance Amount: Text

f.  Total Payable: Read-Only text: Total Amount that needs to be paid.

<!-- -->

B.  Reinstatement Collection (REINSTATEMENT)

**Fields:**

a.  Reinstatement Fee: Read-Only text

b.  Interest: Read-Only text

c.  Total Payable: Read-Only text

<!-- -->

C.  Revival Lumpsum (REVIVAL_LUMP_SUM)

**Fields:**

a.  Arrears (Premiums): Read-Only text

b.  Interest on Arrears: Read-Only text

c.  Revival fee: Read-Only text

d.  Total Payable Amount: Read-Only text

<!-- -->

D.  Revival Installment (REVIVAL_INSTALLMENT)

**Fields:**

a.  Installment Plan ID: Read-Only text

b.  Current Installment No.: Read-Only text

c.  Installment Amount: Text

d.  Remaining Installment Amount: Read-Only text

e.  Total Payable Amount: Read-Only text

<!-- -->

E.  Full Loan Repayment (LOAN_FULL)

**Fields:**

a.  Loan Principal Outstanding

b.  Accrued Interest

c.  Additional Fee Amount

d.  Assignment Info: Text

e.  Total Closure Amount: Read-only Text

<!-- -->

F.  Partial Loan Repayment (LOAN_PARTIAL)

**Fields**

a.  Outstanding Balance: Read-only Text

b.  Repayment Amount: Text

c.  New Loan Balance: Read-only Text: Shows expected post allocation
    balances.

<!-- -->

G.  Policy Miscellaneous (POLICY_MISC)

**Fields:**

a.  Charge Code: Dropdown: Configured codes (e.g., Endorsement Fee,
    Duplicate Document Fee, etc.).

b.  Charge Description: Read-only text

c.  Amount: Text

d.  GST: Amount

e.  Total Payable Amount: Read-Only Text

f.  Remarks: Text: Optional

<!-- -->

H.  Agent Miscellaneous (AGENT_MISC)

**Fields:**

a.  Agent Code: Text with an option to search Agent code using Agent
    Name, etc.

b.  Agent Name: Auto-populated

c.  Charge Code: Values: Registration Fee, Renewal Fee, Training Fee,
    Penalty, etc.

d.  Amount: Text

e.  GST: text

f.  Total Amount Payable: Read-Only Text

g.  Remarks: Text

#### 5.2.4 Payment Details

Reusable payment capture area supporting Cash, Cheque/DD, UPI, Card
(POS), NetBanking. It exposes contextual fields like instrument number,
bank name, instrument date, UPI ID, POS reference, and payer contact.
Centralized validation ensures correct payment references and readiness
for receipting.

**Fields:**

- Payment Method: Dropdown: Cash, Cheque, DD, UPI, Card (POS), Net
  Banking

- Cheque/DD No.: Text

- Bank Name: Text

- Instrument Date: Date

- UPI ID: Text

- Payer Mobile: Text: Optional: 10 digit only

- Payer Email: Text: Optional: Email Format

**Payment Actions:**

- Authorize Payment -- validates totals & instrument fields; creates
  transaction.

- Generate Receipt -- produces receipt with QR/barcode; renders preview.

#### 5.2.5 Receipt Preview & Actions

Shows the Receipt Number, Issue Date/Time, Itemized Amounts, and
QR/Barcode once payment is authorized. Users can Print, Download PDF, or
Send via Email/SMS. The format adheres to organizational receipting
standards, enabling auditability and customer communication.

**Fields:**

- Receipt No.: Text

- Issue Date/Time: Date

- Operator/Branch Code: Read-Only Text

- Amount Breakdown: Read-Only (itemized lines: base, fees, taxes)

- Payment Ref: Read-Only (Cash, UPI, Card, Cheque, DD)

- QR/Barcode: Image

**Actions:**

- Print

- Download PDF

- Send Email/SMS

#### 5.2.6 Audit & Notes

Captures transaction remarks, supervisor approvals (for overrides or
high-value scenarios), and links to the audit trail. Every action is
logged with user, timestamp, and reason codes, fulfilling governance
requirements and supporting post-transaction reviews.

**Fields:**

- Remarks: Text

- User: User ID for the person performing the Transaction

### 5.3 Special Group Collection

The system shall provide bulk upload functionality on the Special Group
Collection (SGC) page in the PLI application to facilitate creation of
SGC entries for premium remittances received from non-postal DDOs or
approved organizations. The bulk upload shall allow users to upload
policy-wise premium data along with DDO and remittance details, perform
validations on file structure, mandatory fields, policy status, and
total amount matching the remittance, and upon successful validation,
create the SGC record under the appropriate PLI head of account as a
temporary holding entry for subsequent adjustment and reconciliation by
authorized users.

## **6. Test Cases**

  ----------------------------------------------------------------------------------------
  **Scenario**     **Preconditions**       **Steps**               **Expected Result**
  ---------------- ----------------------- ----------------------- -----------------------
  Search policy by Policy exists &         Enter Policy No → Click Header fields
  Policy No --     active/grace            **Search** → Click      auto-populate
  valid                                    **Validate**            (Customer, Product,
                                                                   Status, Mode, Next Due,
                                                                   Loan O/S, Agent).
                                                                   Alerts bar shows any
                                                                   grace/late info.

  Search policy -- Policy doesn't exist    Enter Policy No →       Inline error: "Policy
  invalid policy                           **Search**              not found." No data
                                                                   populated.

  Validate without ---                     Click **Validate** with Validation blocked;
  searching                                empty Policy field      error: "Enter policy to
                                                                   validate."

  View             Validated policy        Click **View            Modal shows last 10
  Ledger/History                           Ledger/History**        transactions & loan
  opens                                                            entries; view-only.

  Policy in Lapse  Policy status = Lapse   Enter Policy No →       Status badge shows
                                           Search → Validate       Lapse; **Renewal**
                                                                   panel disabled; info
                                                                   message: "Renewal not
                                                                   allowed for Lapse."

  Collection Type  Validated policy        Open **Collection       Dropdown visible with
  dropdown shows                           Type**                  only **Renewal**
  only Renewal                                                     selected; no other
                                                                   options.

  Renewal amounts  Policy Active/Grace     Validate → Ensure panel Current Installment,
  load             with next due           shows amounts           Late Fee, GST, Arrears,
                                                                   Total Payable show
                                                                   correct computed values
                                                                   (RO).

  Advance Amount   Policy Active; advance  Input **Advance         Total Payable updates
  entry optional   allowed                 Amount**                correctly = current +
                                                                   arrears + late + GST +
                                                                   advance.

  Advance Amount   Config max advance = 12 Enter advance exceeding Error: "Advance exceeds
  over limit       installments            limit                   allowed limit." No
                                                                   calculation commit.

  Non-numeric      ---                     Enter alphanumeric in   Field validation error:
  advance                                  Advance                 "Enter numeric amount."

  Zero due amount  Policy paid-up scenario Validate                Total Payable shows ₹0;
                   for current cycle                               **Authorize Payment**
                                                                   disabled; info: "No
                                                                   amount due."

  Cash payment --  Validated policy; Total Payment Method=Cash →   Authorization succeeds;
  simple           \> 0                    Authorize               transaction created;
                                                                   **Generate Receipt**
                                                                   enabled.

  Cheque -- all    Validated policy        Payment Method=Cheque → Authorization succeeds;
  mandatory fields                         Fill Instrument Number, instrument captured in
                                           Bank, Date (not future) txn; **Generate
                                           → Authorize             Receipt** enabled.

  Cheque --        ---                     Payment Method=Cheque → Error: "Instrument
  missing                                  Leave InstNo blank →    number is required."
  instrument                               Authorize               Authorization blocked.
  number                                                           

  Cheque -- future ---                     Payment Method=Cheque → Error: "Instrument date
  date                                     Date \> today →         cannot be in the
                                           Authorize               future." Blocked.

  DD -- bank name  ---                     Payment Method=DD →     Error: "Bank name is
  missing                                  Leave Bank blank →      required." Blocked.
                                           Authorize               

  UPI -- valid VPA ---                     Payment Method=UPI →    Authorization succeeds;
                                           Enter VPA → Authorize   UPI reference captured
                                                                   post-gateway;
                                                                   **Generate Receipt**
                                                                   enabled.

  UPI -- invalid   ---                     Payment Method=UPI →    Inline error: "Enter a
  VPA format                               Enter invalid VPA →     valid UPI ID
                                           Authorize               (name@psp)." Blocked.

  Card (POS) --    POS available           Payment Method=Card →   Error: "POS reference
  reference                                Leave POS Ref blank →   is required." Blocked.
  required                                 Authorize               

  NetBanking --    NB gateway reachable    Payment                 Redirect/flow
  success                                  Method=NetBanking →     completes;
                                           Authorize               authorization success;
                                                                   **Generate Receipt**
                                                                   enabled.

  Gateway down     Simulate outage         Payment                 Error banner: "Payment
  (UPI/NB)                                 Method=UPI/NetBanking → service unavailable.
                                           Authorize               Try later." Transaction
                                                                   not created.

  Payer Mobile     ---                     Enter Mobile with       Error: "Enter 10-digit
  validation                               non-10 digits           mobile."

  Payer Email      ---                     Enter invalid email     Error: "Enter valid
  format                                                           email address."

  Generate Receipt Authorization done      Click **Generate        Receipt No generated;
  after                                    Receipt**               Issue Date/Time
  authorization                                                    populated; QR/Barcode
                                                                   shown; Payment Ref
                                                                   captured.

  Print & Download Receipt generated       Click **Print** &       Print dialog opens; PDF
                                           **Download PDF**        downloads; file name
                                                                   pattern
                                                                   RCP-YYYYMM-#####.pdf.

  Email/SMS send   Receipt generated;      Click **Send            Notifications sent;
                   contact present         Email/SMS**             success toast; audit
                                                                   logged.

  Generate without No authorization        Click **Generate        Button disabled;
  auth                                     Receipt**               message: "Authorize
                                                                   payment before
                                                                   generating receipt."

  Duplicate        Receipt already issued  Click **Generate        Idempotency guard; no
  Generate attempt                         Receipt** again         second receipt; info:
                                                                   "Receipt already
                                                                   generated for this
                                                                   transaction."

  Remarks optional Authorization/receipt   Enter Remarks → Save    Remarks saved; appears
  saved            done                                            in audit trail with
                                                                   timestamp & user.

  User             Logged-in user session  Perform transaction     Audit shows UserID,
  auto-capture                                                     time, action, reason
                                                                   codes (where
                                                                   applicable).

  Operator cannot  Operator role           Attempt to edit RO      RO fields locked; no
  override amounts                         fields                  edit allowed; audit
                                                                   logs any attempt (if
                                                                   tracked).

  Supervisor       Threshold set (e.g.,    Payment Method=Cash;    Supervisor approval
  required for     ₹50,000)                Amount ≥ threshold →    prompt; without
  high cash                                Authorize               approval → blocked;
                                                                   with approval →
                                                                   proceed.

  Late fee         Overdue beyond grace    Validate policy         Late Fee matches rules;
  calculation                                                      Total Payable = sum of
  correct                                                          all components.

  GST component    GST applicable          Validate policy         GST field shows correct
  displayed                                                        tax; included in total.

  Negative total   Data anomaly            Validate policy with    System safeguards:
  prevented                                malformed amounts       totals cannot be
                                                                   negative; error:
                                                                   "Invalid amount
                                                                   calculation."

  Switch policy    Have unsaved data       Enter some values →     Prompt: "Switching
  after partial                            change Policy →         policy will clear
  entry                                    Validate                entered data. Proceed?"
                                                                   On confirm, panel
                                                                   resets to new policy.

  Global error     Simulate service error  Trigger validation when Red banner with
  banner styling                           IMS down                actionable message &
                                                                   retry option; no crash.
  ----------------------------------------------------------------------------------------

## **6. Appendices**

The Following Documents attached below can be used.
