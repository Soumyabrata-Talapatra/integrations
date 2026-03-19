> INTERNAL APPROVAL FORM

**Project Name:** Agent Incentive, Commission and Producer Management

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
[6](#functional-requirements-specification)](#functional-requirements-specification)

[4.1 Agent Onboarding [6](#agent-onboarding)](#agent-onboarding)

[4.1.1 New Profile Options
[6](#new-profile-options)](#new-profile-options)

[4.1.2 Enter Profile Details Page
[6](#enter-profile-details-page)](#enter-profile-details-page)

[4.1.3 Select New Advisor Coordinator Page
[8](#select-new-advisor-coordinator-page)](#select-new-advisor-coordinator-page)

[4.2 Agent Profile Management
[8](#agent-profile-management)](#agent-profile-management)

[4.2.1 Agent Search [8](#agent-search)](#agent-search)

[4.2.2 Agent Profile Maintenance Page
[9](#agent-profile-maintenance-page)](#agent-profile-maintenance-page)

[4.2.3 License Management Page
[10](#license-management-page)](#license-management-page)

[4.2.3 Agent Termination Page
[10](#agent-termination-page)](#agent-termination-page)

[4.3 Agent Commission Management
[11](#agent-commission-management)](#agent-commission-management)

[4.3.1 Commission Rate Table View Page
[11](#commission-rate-table-view-page)](#commission-rate-table-view-page)

[4.3.2 Commission History Search Page
[11](#commission-history-search-page)](#commission-history-search-page)

[4.3.3 Commission Processing Steps:
[11](#commission-processing-steps)](#commission-processing-steps)

[**5. Appendices** [15](#appendices)](#appendices)

## **1. Executive Summary**

To provide the functional and non-functional requirements for developing
an Agent Onboarding & Commission Management System similar to PMACS for
India Post PLI. The system will support onboarding and administration of
Agents, Advisor Coordinators, Field Officers, and Departmental
Employees; manage licensing and status lifecycle; and calculate,
approve, and disburse commissions.

## **2. Project Scope**

This system will support:

- Agent onboarding (Advisor, Coordinator, Field Officer, Departmental
  Employee)

- Agent profile management

- Commission rate setup and processing

- Trail and Final Incentive statement generation and disbursement

- Licensing and termination workflows

## **3. Business Requirements**

+-------------+---------------+-----------------------------------------------+
| Requirement | Functionality | Requirement                                   |
| ID          |               |                                               |
+:============+===============+:==============================================+
| FS_IC_001   | Agent         | The system shall allow creation of a new      |
|             | Onboarding    | Advisor profile, which must be linked to an   |
|             |               | existing Advisor Coordinator already present  |
|             |               | in the system.                                |
+-------------+               +-----------------------------------------------+
| FS_IC_002   |               | The system shall support creation of a new    |
|             |               | Advisor Coordinator profile, including        |
|             |               | assignment to a specific circle and division. |
+-------------+               +-----------------------------------------------+
| FS_IC_003   |               | The system shall enable onboarding of         |
|             |               | Departmental Employees by auto-populating     |
|             |               | profile data using a valid Employee ID from   |
|             |               | the HRMS system.                              |
+-------------+               +-----------------------------------------------+
| FS_IC_004   |               | The system shall allow onboarding of Field    |
|             |               | Officers either by auto-fetching data using   |
|             |               | Employee ID or through manual entry.          |
+-------------+---------------+-----------------------------------------------+
| FS_IC_005   | Agent Profile | The system shall provide a search interface   |
|             | Management    | to locate agent profiles using Agent ID,      |
|             |               | Name, PAN, or Mobile Number.                  |
+-------------+               +-----------------------------------------------+
| FS_IC_006   |               | The system shall display a dashboard view of  |
|             |               | the agent profile with editable sections for  |
|             |               | each data category.                           |
+-------------+               +-----------------------------------------------+
| FS_IC_007   |               | The system shall allow authorized users to    |
|             |               | update the agent's name with proper           |
|             |               | validation and audit logging.                 |
+-------------+               +-----------------------------------------------+
| FS_IC_008   |               | The system shall allow updating of PAN        |
|             |               | information with format validation and        |
|             |               | uniqueness checks.                            |
+-------------+               +-----------------------------------------------+
| FS_IC_009   |               | The system shall support status updates for   |
|             |               | agents (e.g., Active, Suspended, Terminated)  |
|             |               | along with mandatory reason entry.            |
+-------------+               +-----------------------------------------------+
| FS_IC_010   |               | The system shall allow updating of personal   |
|             |               | information including date of birth, gender,  |
|             |               | and marital status.                           |
+-------------+               +-----------------------------------------------+
| FS_IC_011   |               | The system shall support addition and         |
|             |               | modification of distribution channel details  |
|             |               | with effective dates.                         |
+-------------+               +-----------------------------------------------+
| FS_IC_012   |               | The system shall allow entry and update of    |
|             |               | external identification numbers and their     |
|             |               | sources.                                      |
+-------------+               +-----------------------------------------------+
| FS_IC_013   |               | The system shall support assignment and       |
|             |               | modification of product class information     |
|             |               | linked to the agent.                          |
+-------------+               +-----------------------------------------------+
| FS_IC_014   |               | The system shall allow entry and update of    |
|             |               | multiple address types: Official, Permanent,  |
|             |               | and Communication.                            |
+-------------+               +-----------------------------------------------+
| FS_IC_015   |               | The system shall support entry and update of  |
|             |               | phone numbers including official/resident     |
|             |               | landline and mobile numbers.                  |
+-------------+               +-----------------------------------------------+
| FS_IC_016   |               | The system shall allow entry and update of    |
|             |               | email addresses categorized as official,      |
|             |               | permanent, and communication.                 |
+-------------+               +-----------------------------------------------+
| FS_IC_017   |               | The system shall support assignment of        |
|             |               | authority types and validity periods for      |
|             |               | agents.                                       |
+-------------+               +-----------------------------------------------+
| FS_IC_018   |               | The system shall allow entry and update of    |
|             |               | insurance licensing details and generate      |
|             |               | automated reminders at 1 month, 15 days, 7    |
|             |               | days, and on the day of expiry.               |
+-------------+               +-----------------------------------------------+
| FS_IC_019   |               | The system shall enforce license renewal      |
|             |               | rules: first renewal after 1 year, subsequent |
|             |               | renewals every 3 years.                       |
+-------------+               +-----------------------------------------------+
| FS_IC_020   |               | The system shall automatically deactivate an  |
|             |               | advisor code if the license renewal date has  |
|             |               | elapsed, or allow manual cancellation via the |
|             |               | License Update Entry page.                    |
+-------------+---------------+-----------------------------------------------+
| FS_IC_021   | Advisor       | The system shall allow termination of an      |
|             | Termination   | advisor profile with mandatory entry of       |
|             |               | termination reason and effective date.        |
+-------------+---------------+-----------------------------------------------+
| FS_IC_022   | Commission    | The system shall provide a Commission Rate    |
|             | Processing    | Table setup interface with fields for Rate,   |
|             |               | Policy Duration (Months), Product Type,       |
|             |               | Product Plan Code, Agent Type, and Policy     |
|             |               | Term (Years).                                 |
+-------------+               +-----------------------------------------------+
| FS_IC_023   |               | The system shall allow searching of           |
|             |               | commission history by policy number and agent |
|             |               | ID.                                           |
+-------------+               +-----------------------------------------------+
| FS_IC_024   |               | The system shall support execution of monthly |
|             |               | Commission Calculation Batch jobs to compute  |
|             |               | agent commissions.                            |
+-------------+               +-----------------------------------------------+
| FS_IC_025   |               | The system shall support automatic generation |
|             |               | of Trial Statements via batch job based on    |
|             |               | policies sold.                                |
+-------------+               +-----------------------------------------------+
| FS_IC_026   |               | The system shall provide a page to view       |
|             |               | generated Trial Statements with agent-wise    |
|             |               | commission details.                           |
+-------------+               +-----------------------------------------------+
| FS_IC_027   |               | The system shall provide a Manual Trial       |
|             |               | Statement Generation page with fields for     |
|             |               | Processing Unit, Statement Format, Max        |
|             |               | Statement Due Date, Max Transaction Effective |
|             |               | Date, Max Process Date, Statement Date,       |
|             |               | Contract Holder, Advisor Coordinator, and     |
|             |               | Carrier.                                      |
+-------------+               +-----------------------------------------------+
| FS_IC_028   |               | The system shall provide an Approving Trial   |
|             |               | Statement page that displays commission       |
|             |               | amounts and allows full or partial            |
|             |               | disbursement approval.                        |
+-------------+               +-----------------------------------------------+
| FS_IC_029   |               | The system shall support execution of Final   |
|             |               | Incentive Statement Generation batch job      |
|             |               | after trial statement approval.               |
+-------------+               +-----------------------------------------------+
| FS_IC_030   |               | The system shall provide a Final Statements   |
|             |               | page displaying final commission amounts for  |
|             |               | agents.                                       |
+-------------+               +-----------------------------------------------+
| FS_IC_031   |               | The system shall provide a Disbursement       |
|             |               | Details page to input cheque or EFT           |
|             |               | information.                                  |
+-------------+               +-----------------------------------------------+
| FS_IC_032   |               | The system shall support automatic            |
|             |               | disbursement of commission amounts based on   |
|             |               | final statements, with immediate processing   |
|             |               | for cheque and queued processing for EFT.     |
+-------------+---------------+-----------------------------------------------+

## **4. Functional Requirements Specification**

## 4.1 Agent Onboarding

### 4.1.1 New Profile Options

- **Purpose:** To select the type of Agent that needs to be onboarded.

- **Fields & Rules:**

  - Agent Type: dropdown: Options are Advisor, Advisor Coordinator,
    Departmental Employee, Field Officer

  - Employee Number: Textbox: Mandatory for Departmental Employee,
    optional for Field Officer, Not Applicable for others.

  - Person Type: Dropdown: Options are Individual, Corporate/Group.

  - Advisor Undergoing Training: Checkbox: Default Unchecked

  - Continue: button

### 4.1.2 Enter Profile Details Page

- **Purpose:** This page will get displayed after selecting new profile
  options and clicking continue button.

- **Fields & Rules:**

  - Profile Type: Dropdown

  - Office Type: Dropdown

  - Office Code: Textbox

  - Advisor Sub-Type: Dropdown

  - Effective Date: Calendar

  - Distribution Channel: Multiselect Dropdown: Options are India Post.

  - Product Class: Multiselect Dropdown: Options are PLI, RPLI

  - Title: Dropdown

  - First Name: Textbox

  - Middle Name: Textbox

  - Last Name: Textbox

  - Gender: Dropdown: Options are Male, Female, Other

  - Date of Birth: Calendar

  - Category: Dropdown

  - Marital Status: Dropdown

  - Aadhar Number: Textbox

  - PAN: Textbox

  - Designation/Rank: Dropdown

  - Service Number: Textbox

  - Professional Title: Dropdown

  - Address:

    - Address Type: Dropdown: Options as Official, Permanent,
      Communication.

    - Address Line1-

    - Address Line2-

    - Village-

    - Taluka-

    - City-

    - District-

    - State-

    - Country-

    - Pin Code-

  - Phone:

  - Email:

  - Bank Account#:

  - Bank IFSC Code:

  - Superior Advisor: This will open Select New Advisor Coordinator
    Page, and the user will return to this page after selection of
    advisors.

  - Office Affiliation: Textbox: Input Affiliated Office Code

  ------------------------------------------------------------------------
  Serial     Condition           Error Message           Required Action
  Number                                                 
  ---------- ------------------- ----------------------- -----------------
  1          If the **Profile    Please select a Profile Users must select
             type** is not       Type.                   the Profile Type
             selected and                                from the
             **Continue** button                         drop-down list.
             is pressed.                                 

  2          If the PAN number   PAN number entered      Users must enter
             entered already     already exists for      the PAN number
             exists for some     another advisor's       which does not
             other profile.      profile and cannot be   exist for another
                                 for this profile.       advisor's
                                                         profile.

  3          PAN number should   Please enter a 10 digit Users must enter
             be of 10            Permanent Account       the 10 digit PAN
             characters. If the  Number (PAN).           number.
             PAN number length                           
             doesn't match.                              

  4          PAN number should   Please enter correct    Users must enter
             be entered in the   PAN.                    the PAN number as
             standard format as                          per the standard
             shown above. If the                         format.
             PAN doesn't match                           
             with the format as                          
             already defined in                          
             the system.                                 

  5          If the Last name is Please enter a Last     Users must enter
             not entered and     name.                   the last name of
             **Continue** button                         the Advisor.
             is pressed.                                 

  6          If the First name   Please enter a First    Users must enter
             is not entered and  name.                   the first name of
             **Continue** button                         the Advisor.
             is pressed.                                 

  7          If the **Date of    Please enter a valid    Users must enter
             Birth** is not      Date of Birth           the valid Date Of
             entered and                                 Birth.
             **Continue** button                         
             is pressed.                                 
  ------------------------------------------------------------------------

### 4.1.3 Select New Advisor Coordinator Page

- **Purpose:** If the new profile reports to an Advisor Coordinator
  already present in system, then select Agent type as 'Advisor' and
  click the Continue button to move the user to the Select New Advisor
  Coordinator page.

- **Fields & Rules:**

  - AC Profile#: text box

  - AC Profile Name: text box

  - Search: button

- Clicking the search button should display the following table with
  auto-populated agent details:

  - AC Profile#

  - AC Profile Name

  - Profile Type

  - Status

  - Person: Individual or Corporate/Group

  - Action: Link for selecting the Advisor Coordinator

  -----------------------------------------------------------------------
  Serial Number Error Message                     Required Action
  ------------- --------------------------------- -----------------------
  1             Your selected criteria did not    Users must change the
                return any rows. Please change    selection
                your selections and try again.    

  -----------------------------------------------------------------------

## 4.2 Agent Profile Management

### 4.2.1 Agent Search

- **Purpose:** To search an Existing Agent for performing some action
  like View / Edit / Terminate / Suspend / viewing Commission History.

- **Fields:**

  - Agent ID: Textbox

  - Last Name: Textbox

  - First Name: Textbox

  - PAN: Textbox

  - Mobile Number: Textbox

  - Status: Textbox

  - Superior Advisor Code: Textbox

  - Office ID: Textbox

  - Advisor Undergoing Training: Checkbox

  - Search: button

- **Business Rules**:

  - Results displayed in a table with clickable rows.

  - Clicking on any Agent in table will open that Agent Profile
    Maintenance Page for that agent.

  - 'Export to Excel' option should be present for the table details to
    be exported in excel format.

### 4.2.2 Agent Profile Maintenance Page

- **Purpose:** Displays agent details with option to edit each section.

- **Fields & Rules:** It should display all the information about the
  agent with 'Update' link in each section to update the details of that
  section. The sections that need to be displayed are:

  - Advisor Name Section

  - PAN Number Information

  - Status Information: Add additional 'Change Status To' dropdown for
    updating the Agent Status to Expired/Suspended/ Terminated.

  - Personal Information

  - Distribution Channel

  - External Number

  - Product Class

![A screenshot of a computer AI-generated content may be
incorrect.](media/image1.png){width="6.268055555555556in"
height="3.451388888888889in"}

### 4.2.3 License Management Page

- **Purpose:** This page will display the list of licenses the current
  agent have along with the option to Add or Delete License.

- **Fields:**

  - License Line: Dropdown: Option as Life

  - License Type: Dropdown

  - License Number: Textbox

  - Resident Status: Dropdown: Option as Resident, Non-Resident

  - License Date: Calendar

  - Renewal Date: Calendar

  - Authority Date: Calendar

  - Submit: Button

  - Update Renewal Date: Button

  - Delete License: Button

- **Business Rules:**

  - First License Renewal notice will be generated for the license
    renewal 1 month before the License expiry date.

  - Second License Renewal notice will be generated for the license
    renewal 15 days before the License expiry date, if not renewed
    already.

  - Third License Renewal notice will be generated for the license
    renewal 7 days before the License expiry date, if not renewed
    already.

  - Final License Renewal notice will be generated for the license
    renewal on the License expiry date, if not renewed already.

### 4.2.3 Agent Termination Page

- Search an Agent.

- Move to 'Agent Profile Maintenance Page' for the Agent.

- For terminating a profile, user needs to select the 'Change Status To'
  Dropdown in status section as Terminated and then click on Update.
  Agent Termination Page will open. Input the details and click submit
  to terminate the agent.

- **Fields:**

  - Status: dropdown

  - Status Reason: dropdown

  - Status Date: Calendar

  - Effective Date: Calendar

  - Termination Date: Calendar

  - Submit: Button

## 4.3 Agent Goal Setting

- **Purpose:** To set the performance goals for the Agents. The set
  goals should reflect in the agent profile in Agent Portal when the
  agent logins from his ID.

- **Fields:**

  - Agent ID: Text: Option should be given to search & select the agent.

  - Agent Name: Text: Auto-populated

  - Goal Period: From & To Dates from Calendar: To specify timeframe for
    the goal.

  - Target Number of Policies: Textbox: The number of new policies the
    agent aims to sell.

  - Target Premium Collection (₹): Textbox: The total premium amount
    expected to be collected.

  - Product-wise Targets: Textbox: Goals broken down by PLI/RPLI product
    types (e.g., Endowment, Whole Life).

  - Comments: Textbox

  - Submit: button

## 4.4 Agent Commission Management

### 4.4.1 Commission Rate Table View Page

- **Purpose:** To view the Defined commission rates based on multiple
  parameters.

- **Fields:** The page should display the following Commission table:

  - Rate (%): Decimal: Commission percentage

  - Policy Duration (Months): Integer: Duration of policy in months

  - Product Type: Dropdown: PLI, RPLI

  - Product Plan Code: Text: Unique code for the plan

  - Agent Type: Dropdown: e.g., Direct Agent, Field Officer

  - Policy Term (Years): Integer: Total term of the policy

- **Action:**

  - User should only view the table.

### 4.4.2 Commission History Search Page

- **Purpose:** View historical commission data.

- **Search Filters:**

  - Agent ID

  - Policy Number

  - Date Range

  - Product Type

  - Commission Type (First Year, Renewal, Bonus)

- **Result Table:**

  - \| Agent ID \| Policy Number \| Product Type \| Commission Type \|
    Amount \| Status \| Date Processed \|

- **Actions:**

  - Export to Excel/PDF

  - View Detailed Statement

### 4.4.3 Commission Processing Steps:

- **Purpose:** To generate and pay the commission for the Agents.

![](media/image2.png){width="3.638888888888889in"
height="5.458333333333333in"}

#### 4.4.3.1 Run Commission Calculation Batch Jobs

- **Trigger:** Scheduled or Manual

- **Function:** Calculates Commission based on active policies and rate
  table.

#### 4.4.3.2 Trial Statement Generation Batch Job

- **Function:** Generates trial commission statements for review.

- **Output:** Trial Statement per agent and policy.

#### 4.4.3.3 View Trial Statement Page

- **Fields:**

  - Agent ID

  - Policy Number

  - Commission Type

  - Calculated Amount

  - Status (Pending/Approved)

  - Remarks

- **Action:**

  - Filter by Agent, Policy, Circle

  - Export to Excel/PDF

  - Raise Correction

#### 4.4.3.4 Manual Trial Statement Generation Page

- **Fields:**

  - Processing Unit: Dropdown: e.g., IT2.0

  - Statement Format: Dropdown: e.g., Standard

  - Max Statement Due Date: Date: Cut-off for statement

  - Max Transaction Effective Date: Date: Latest transaction date

  - Max Process Date: Date: Latest processing date

  - Statement Date: Date: Date of statement

  - Contract Holder: Text: Name of policyholder

  - Advisor Coordinator: Text: Agent's supervisor

  - Carrier: Text: Insurance carrier

  - Tax Deduction (TDS %): Decimal: Applicable tax deduction

- **Action:**

  - Generate Statement

  - Save Draft

  - Submit for Approval

#### 4.4.3.5 Approving Trial Statement Page

- **Fields:**

  - Agent ID

  - Policy Number

  - Commission Amount

  - Status

  - Remarks

  - Disbursement Option (Full / Partial)

  - Part Disbursement (%) -- Enabled only if selected

- **Actions:**

  - Apply Part Disbursement

  - Approve All Rows in P/U

  - Submit

#### 4.4.3.6 Final Incentive Statement Generation Batch Job

- **Trigger:** Scheduled or Manual

- **Function:** Locks approved trial data and generates final commission
  statements.

#### 4.4.3.7 Final Statements Page

- **Fields:**

  - Agent ID

  - Policy Number

  - Final Commission Amount

  - TDS Deducted

  - Net Payable

  - Payment Status

- **Actions:**

  - View Statement

  - Export PDF/Excel

  - Send to Disbursement

#### 4.4.3.8 Disbursement Details Page

- **Fields:**

  - Agent ID

  - Payment Mode (Cheque / EFT)

  - Cheque Number (if applicable)

  - Bank Name

  - IFSC Code

  - Account Number

  - Payment Date

  - Amount Paid

  - Remarks

- **Actions:**

  - Save

  - Submit

  - Generate Payment File

#### 4.4.3.9 Automatic Disbursement Option

- **Functionality:**

  - If Cheque is selected:

    - Disbursement is marked as complete immediately.

  - If EFT is selected:

    - Payment file is generated and sent to PFMS/Bank.

    - Status updated upon confirmation.

## **5. Test Case**

  --------------------------------------------------------------------------------------------
  **TC     **Functionality**   **Test Case       **Input Data**     **Expected      **Type**
  ID**                         Description**                        Result**        
  -------- ------------------- ----------------- ------------------ --------------- ----------
  TC_001   Agent Onboarding    Create new        Advisor details +  Advisor profile Positive
                               Advisor linked to valid Coordinator  created         
                               existing          ID                 successfully    
                               Coordinator                                          

  TC_002   Agent Onboarding    Create Advisor    Advisor details    Error:          Negative
                               without linking   only               "Coordinator ID 
                               Coordinator                          is mandatory"   

  TC_003   Coordinator         Create new        Coordinator        Coordinator     Positive
           Onboarding          Coordinator with  details + Circle + profile created 
                               valid Circle &    Division           successfully    
                               Division                                             

  TC_004   Coordinator         Create            Coordinator        Error: "Circle  Negative
           Onboarding          Coordinator       details only       assignment      
                               without Circle                       required"       
                               assignment                                           

  TC_005   Dept Employee       Auto-populate     Employee ID from   Profile         Positive
           Onboarding          using valid       HRMS               populated       
                               Employee ID                          correctly       

  TC_006   Dept Employee       Auto-populate     Invalid ID         Error:          Negative
           Onboarding          using invalid                        "Employee ID    
                               Employee ID                          not found"      

  TC_007   Field Officer       Manual entry of   Valid details      Profile created Positive
           Onboarding          details           entered manually   successfully    

  TC_008   Field Officer       Manual entry with Missing Name or ID Error:          Negative
           Onboarding          missing mandatory                    "Mandatory      
                               fields                               fields missing" 

  TC_009   Agent Search        Search by valid   Agent ID = 12345   Agent profile   Positive
                               Agent ID                             displayed       

  TC_010   Agent Search        Search by invalid Agent ID = 99999   Error: "No      Negative
                               Agent ID                             records found"  

  TC_011   PAN Update          Update PAN with   PAN = ABCDE1234F   PAN updated     Positive
                               valid format                         successfully    

  TC_012   PAN Update          Update PAN with   PAN = 12345ABCDE   Error: "Invalid Negative
                               invalid format                       PAN format"     

  TC_013   Status Update       Change status to  Status =           Status updated  Positive
                               Suspended with    Suspended, Reason  successfully    
                               reason            = "Non-compliance"                 

  TC_014   Status Update       Change status     Status =           Error: "Reason  Negative
                               without reason    Suspended, Reason  is mandatory"   
                                                 = blank                            

  TC_015   License Reminder    Generate reminder License expiry =   Reminder        Positive
                               15 days before    15 days ahead      generated       
                               expiry                                               

  TC_016   License Reminder    Generate reminder License expired    No reminder,    Negative
                               after expiry                         license         
                                                                    deactivated     

  TC_017   Commission Rate     Add valid         Rate = 5%, Product Rate saved      Positive
           Setup               commission rate   Type = Endowment   successfully    

  TC_018   Commission Rate     Add rate without  Rate = 5%, Product Error: "Product Negative
           Setup               Product Type      Type = blank       Type required"  

  TC_019   Commission Batch    Execute monthly   Policies exist for Commission      Positive
                               batch with valid  month              calculated      
                               policies                                             

  TC_020   Commission Batch    Execute batch     No policies for    Error: "No data Negative
                               with no policies  month              to process"     

  TC_021   Approve Trial       Approve full      Valid trial        Approval        Positive
           Statement           disbursement      statement          successful      

  TC_022   Approve Trial       Approve without   No selection       Error:          Negative
           Statement           selecting                            "Disbursement   
                               disbursement type                    type required"  

  TC_023   Commission          Auto disbursement Valid EFT details  Commission      Positive
           Disbursement        via EFT                              queued for EFT  

  TC_024   Commission          Disbursement      Missing cheque/EFT Error: "Payment Negative
           Disbursement        without payment   info               details         
                               details                              required"       
  --------------------------------------------------------------------------------------------

## **6. Appendices**

The Following Documents attached below can be used.

![](media/image3.emf)
