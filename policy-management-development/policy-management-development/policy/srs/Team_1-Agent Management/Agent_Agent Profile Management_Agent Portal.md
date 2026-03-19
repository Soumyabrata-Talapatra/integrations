> INTERNAL APPROVAL FORM

**Project Name:** Agent Management (Web Portal for Agent)

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

[4.1 Login Page [6](#login-page)](#login-page)

[4.2 Home Page [6](#home-page)](#home-page)

[4.3 Agent Diary & Tour Management Page
[7](#agent-diary-tour-management-page)](#agent-diary-tour-management-page)

[4.3.1 View Assigned Leads Page
[7](#view-assigned-leads-page)](#view-assigned-leads-page)

[4.3.2 Update Lead Status Page
[7](#update-lead-status-page)](#update-lead-status-page)

[4.3.3 Plan Your Daily Tour Page
[8](#plan-your-daily-tour-page)](#plan-your-daily-tour-page)

[4.3.4 Edit Tour Information Page
[8](#edit-tour-information-page)](#edit-tour-information-page)

[4.3.5 Tour Reports Page [8](#tour-reports-page)](#tour-reports-page)

[4.4 Meeting Planner & Calendar Management page
[9](#meeting-planner-calendar-management-page)](#meeting-planner-calendar-management-page)

[4.4.1 Schedule Meeting Page
[9](#schedule-meeting-page)](#schedule-meeting-page)

[4.4.2 Update Meeting Outcome Page
[9](#update-meeting-outcome-page)](#update-meeting-outcome-page)

[4.5 Doorstep Customer Services page
[9](#doorstep-customer-services-page)](#doorstep-customer-services-page)

[4.6 My Business page [9](#my-business-page)](#my-business-page)

[4.7 New Business page [10](#new-business-page)](#new-business-page)

[**5. Appendices** [10](#appendices)](#appendices)

## **1. Executive Summary**

To develop a comprehensive Agent Portal for India Post PLI agents that
enhances self-service capabilities, streamlines business processes, and
improves agent productivity.

## **2. Project Scope**

The portal will provide secure access to agent-specific data, policy
management tools, commission tracking, customer engagement features, and
support services.

## **3. Business Requirements**

  -----------------------------------------------------------------------
  Requirement   Business          Description
  ID            Requirement       
  ------------- ----------------- ---------------------------------------
  FS_AM_001     User Access &     The portal should provide a secure
                Authentication    OTP-based login system that verifies
                                  the identity of registered agents
                                  before granting access. It should also
                                  support session timeout and maintain
                                  detailed login logs to ensure data
                                  protection and auditability.

  FS_AM_002     Home Page &       The home page should display a
                Dashboard         personalized dashboard showing business
                                  performance, key statistics like
                                  policies sold and commissions earned,
                                  and alerts for renewals or pending
                                  proposals. It should offer quick links
                                  to essential modules such as My
                                  Business, New Business, Reports, and
                                  Tools, ensuring a user-friendly
                                  experience across devices.

  FS_AM_003     Agent Diary &     Agents should be able to plan and
                Tour Management   record their daily field visits,
                                  customer meetings, and follow-ups
                                  through an integrated diary and tour
                                  planner. The system should help them
                                  monitor completed activities, track
                                  expenses, and generate structured
                                  reports for supervisory review.

  FS_AM_004     Meeting Planner & The meeting planner should allow agents
                Calendar          to schedule customer or internal
                Management        meetings with reminders and recurring
                                  options. It should also offer a
                                  calendar view for better time
                                  management and support synchronization
                                  with popular calendar tools like Google
                                  Calendar or Outlook.

  FS_AM_005     Doorstep Customer Agents should be able to record service
                Services          requests such as premium collection,
                                  policy updates, or document pickups
                                  during doorstep visits. The system
                                  should capture timestamps and service
                                  outcomes to ensure transparency and
                                  accountability in customer
                                  interactions.

  FS_AM_006     My Business       This module should allow agents to view
                                  their existing customer portfolios and
                                  policies, generate policy quotes, and
                                  track claims or renewals. It should
                                  also support commission tracking with
                                  detailed incentive statements,
                                  Disbursement History and provide lead
                                  management tools for capturing and
                                  converting potential customers.

  FS_AM_007     New Business      Agents should be able to generate
                                  premium quotes and benefit
                                  illustrations and submit new policy
                                  proposals digitally with e-KYC and
                                  e-sign features. The system should help
                                  track the proposal status and link with
                                  medical appointment scheduling where
                                  required.

  FS_AM_008     Performance       Agents should be able to monitor their
                Tracking          progress against business targets
                                  through a dedicated performance
                                  dashboard. Supervisors should have
                                  visibility into overall performance
                                  metrics across agents or regions to
                                  assess productivity and identify
                                  improvement areas.

  FS_AM_009     Tools & Downloads The portal should provide downloadable
                                  resources such as policy servicing
                                  forms, brochures, and FAQs to support
                                  agents' day-to-day operations. It
                                  should also offer calculators and
                                  product estimators to assist in
                                  customer advisory and policy planning.

  FS_AM_010     Reports           Agents should be able to generate and
                                  download a wide range of reports,
                                  including business, commission,
                                  incentive, and policy-related reports.
                                  These reports should be easy to filter,
                                  export, and interpret, providing a
                                  clear view of business performance and
                                  compliance requirements.

  FS_AM_011     User Profile      The user profile section should allow
                Management        agents to update their personal
                                  details, upload KYC and bank
                                  information, and manage communication
                                  preferences. It should also notify them
                                  of license renewal deadlines and
                                  maintain their verified identification
                                  data securely.

  FS_AM_012     Learning &        A learning section should offer digital
                Development       training modules, product tutorials,
                                  and certification programs to enhance
                                  agent skills. It should track learning
                                  progress and completion while keeping
                                  agents informed about new product
                                  updates and mandatory training
                                  sessions.

  FS_AM_013     Help & Support    The help section should include FAQs,
                                  guides, and tutorials to assist agents
                                  in resolving technical or operational
                                  queries. It should also enable them to
                                  raise support tickets, contact IT
                                  helpdesks, and receive timely responses
                                  through a centralized grievance
                                  management system.

  FS_AM_014     Document          The portal should allow agents to
                Management        securely upload and manage
                                  customer-related documents such as KYC
                                  proofs and policy forms. It should
                                  maintain version control, provide easy
                                  retrieval, and integrate with digital
                                  repositories for seamless document
                                  exchange.
  -----------------------------------------------------------------------

## **4. Functional Requirements Specification**

**Flow Diagram for the Agent Portal:**

![](media/image1.png){width="6.268055555555556in"
height="3.3805555555555555in"}

## Login Page

> Opening the Agent Portal should open the Login Page.

- **Fields:**

  - Agent ID: Text

  - Password: Text

  - Login: Button

  - Forget Password: Link

- **After Clicking Login:**

  - OTP will be sent to the registered mobile number. The following
    fields will be displayed:

    - OTP: Text

    - Submit: Button

- **Post Login**

  - After successful login, the user should be navigated to the Agent
    Home Page.

## Home Page

> The Home Page will serve as the primary dashboard, showing key
> metrics, notifications, and quick navigation links to core modules.
> This page shows notifications for renewals, appointments, meetings,
> and new leads.

- **Provide navigation tiles/links for all modules:**

  - Agent Diary & Tour Management

  - Meeting Planner & Calendar Management

  - My Business

  - New Business

  - Doorstep Customer Services

  - Medical Appointment

  - Document Upload & Management

  - Tools & Downloads

  - Reports

  - User Profile

  - Learning

  - Help

  - Notifications

  - Logout link

## Agent Diary & Tour Management Page

> To record, plan, and manage daily tours and field activities of agents
> including lead management activities received.

- **Fields:**

  - View Assigned Leads: button

  - Create New Tour: button

  - Tour Report: button: To generate tour report

- **Table:**

  - The page should list all the previously created Tours in table
    format with 'Edit' button present for each row. Clicking on 'Edit
    button\' will move the user to 'Edit Tour Information' Page for that
    Tour.

### View Assigned Leads Page

To allow agents to view leads assigned by supervisors or generated
through system lead allocation.

- **Table (non-editable) contain the following details of leads:**

  - Lead Name

  - Contact Number

  - Email ID

  - Address

  - PIN Code

  - Source

  - Product Interested: Dropdown

  - Lead Type

  - Remarks

  - Status: Dropdown: Options are Not-Attended, In-Progress, Customer
    Not Interested, Policy Taken

  - Update Lead: link: Clicking this link should move the user to Update
    Lead status page.

### Update Lead Status Page

To allow agents to update status of leads with comments.

- **Fields:**

  - Lead Name: Text

  - Contact Number: Text

  - Status: Text

  - Comments: Text

  - Submit: Button

### Plan Your Daily Tour Page

To enable the agent to create a planned route for customer visits on a
specific date.

- **Fields:**

  - Tour Date: Calendar

  - Location: Text

  - Purpose of Tour: Text

  - Customer Name: Text

  - Customer Mobile: Text

  - Customer Email: Text

  - Customer Address: Text

  - Product Interested: Dropdown

  - Status: Dropdown: Started, In-progress, Complete

  - Submit: button

### Edit Tour Information Page

To update the details for any tour.

- **Fields:**

  - Tour Date: Calendar

  - Location: Text

  - Purpose of Tour: Text

  - Customer Name: Text

  - Customer Mobile: Text

  - Customer Email: Text

  - Customer Address: Text

  - Product Interested: Dropdown

  - Submit: button

### Tour Reports Page

To view the reports for the tour.

- **Fields:**

  - Tour Start Date: Calendar

  - Tour End Date: Calendar

  - Status: Dropdown: Options are All, Started, In-Progress, Completed.

  - Generate Report: Button: Clicking this should generate the report as
    per filter given above in Excel format.

## Meeting Planner & Calendar Management page

> To plan and schedule customer meetings, team meetings, and training
> sessions.

- **Fields:**

  - Scheduled Meetings & Training Sessions: Table: Display the currently
    scheduled customer meetings, team meetings, and training sessions
    with 'Update' link beside each meeting to Update Meeting Outcomes.

  - Schedule Meeting: button

### Schedule Meeting Page

- **Fields:**

  - Meeting Type: Dropdown: Options are 'Customer Meeting', 'Team
    Meeting' and 'Training Session'.

  - Automatic Reminder Date: Calendar

  - Automatic Reminder Time: Timestamp

  - Meeting Details: Text

### Update Meeting Outcome Page

- **Fields:**

  - List of Attendees: Auto-populate: It should auto-populate the list
    of attendees who have attended the meeting.

  - Meeting Details: Text

  - Meeting Outcome: Text

  - Submit: Button

## Doorstep Customer Services page

> The page is used to view the lead management report and service
> request

- **Fields:**

  - New Service Request: button

  - Lead Management: button

  - Service Request Information: Table: There should be a table with
    list of previously raised service requests along with status and
    option to 'Update' beside each of the service requests.

  - Generate Report: button: To download the list of service requests
    along with status in excel format.

- **Rules:**

  - Clicking 'New Service Request' should move the user to 'New Service
    Request' page for creating new service request for the client.

  - Clicking \'Lead Management' should move the user to 'View Assigned
    Lead Page'.

### New Service Request Page

The page should be used to create new service requests for the
customers.

- **Fields:**

  - Request Type: Mandatory

  - Policy Number: Option to search & select the policy should be
    present

  - Paid to Date: Mandatory

  - Intimator First Name: Mandatory for Death Claim Only

  - Intimator Middle Name: Mandatory for Death Claim Only

  - Intimator Last Name: Mandatory for Death Claim Only

  - Intimator Rel to Insured: Mandatory for Death Claim Only

  - Intimator Phone Number: Mandatory for Death Claim Only

  - Date Of Death: Mandatory for Death Claim Only

  - Disputed Reason: Optional

  - Primary Insured: Optional

  - Secondary Insured: Optional

  - Mandatory Document: Option to Upload Document should be there.

  - Service Request Date: Mandatory

  - Policy Number: Mandatory

  - Office Code: Mandatory

  - Service Request Channel: Mandatory

  - Username: Mandatory

  - Submit: button

## My Business page

> My business page will display the links for current business related
> aspects for the agent.

- **Fields:**

  - Claims Management & Quotes Generation: button

  - Commission & Incentive Tracking: button

  - Business Performance Tracking: button

### Claims Management Dashboard Page

- **Purpose:** To generate the quote for Policy Loan, Surrender, Revival
  and Reduced Paid-up for the policy.

![Untitled
diagram-2025-10-30-062532.png](media/image2.png){width="6.268055555555556in"
height="4.517720909886264in"}

### 4.6.2 Commission Tracking Page

- **Purpose:** To view the commission for the agent and to generate
  commission and tax statements.

![Untitled
diagram-2025-10-24-094619.png](media/image3.png){width="2.982988845144357in"
height="5.965974409448819in"}

### Business Performance Tracking Page

- **Purpose:** To view the Agent Goals and to check if the target is
  getting achieved by the agent.

![Untitled
diagram-2025-10-30-062904.png](media/image4.png){width="6.268055555555556in"
height="5.0887773403324585in"}

### Customer Management & Policy Alerts

- **Purpose:** To view the status of policies for the agent and to send
  the revival offers for the agents.

![Untitled
diagram-2025-10-24-094906.png](media/image5.png){width="2.2847222222222223in"
height="7.075269028871391in"}

## New Business page

### New Business Issuance

- **Purpose:** To perform the process of policy issue by the agent.

![Untitled
diagram-2025-10-30-062306.png](media/image6.png){width="5.652299868766404in"
height="7.216589020122485in"}

### Medical Examination Management

![Untitled
diagram-2025-10-24-095218.png](media/image7.png){width="3.826388888888889in"
height="6.930153105861767in"}

### Pending Policies Tracking

- **Purpose:** To track the status of pending applications.

![Untitled
diagram-2025-10-24-095359.png](media/image8.png){width="4.729166666666667in"
height="6.996574803149606in"}

## Digital Learning & Product Information

To Access the Information related to PLI/RPLI Products and for other
learning activities.

![Untitled
diagram-2025-10-24-095005.png](media/image9.png){width="6.268055555555556in"
height="3.2468547681539808in"}

## Document Upload page

> The page is used to upload the documents to be sent to CPC and PO for
> processing.![Untitled
> diagram-2025-10-24-095523.png](media/image10.png){width="3.724286964129484in"
> height="4.572794181977253in"}

## **5. Test Case**

  -------------------------------------------------------------------------
  **TC ID**  **Test Case     **Description**        **Expected Result**
             Name**                                 
  ---------- --------------- ---------------------- -----------------------
  TC_001     Successful OTP  Verify that a          Agent is logged in
             Login           registered agent can   successfully
                             log in using OTP       

  TC_002     Invalid OTP     Verify login fails     Error message
             Login           with incorrect OTP     displayed: "Invalid
                                                    OTP"

  TC_003     Session Timeout Verify session expires Session ends; user
                             after inactivity       redirected to login
                                                    page

  TC_004     Multiple Failed Verify account lock    Account locked; error
             Login Attempts  after 5 failed OTP     message displayed
                             attempts               

  TC_005     Dashboard Loads Verify dashboard shows Dashboard displays
             Correctly       policies sold,         accurate data
                             commissions, alerts    

  TC_006     Dashboard on    Verify responsive      Dashboard elements
             Mobile          design on mobile       adjust properly
                             devices                

  TC_007     Quick Links     Verify quick links     Redirects to correct
             Navigation      navigate to correct    page
                             modules                

  TC_008     Add New Tour    Verify agent can add a Tour plan saved
             Plan            new tour plan          successfully

  TC_009     Invalid Date    Verify system rejects  Error message displayed
             Entry           past dates for new     
                             tour                   

  TC_010     Sync with       Verify meeting sync    Meeting appears in
             Google Calendar works with Google      Google Calendar
                             Calendar               

  TC_011     Record Premium  Verify agent can       Timestamp and outcome
             Collection      record premium         saved
                             collection during      
                             visit                  

  TC_012     View Customer   Verify agent can view  Customer portfolio
             Portfolio       customer policies      displayed

  TC_013     Submit New      Verify proposal        Proposal submitted
             Proposal with   submission works with  successfully
             e-KYC           e-KYC                  

  TC_014     View            Verify agent sees      Dashboard shows
             Performance     correct targets and    accurate metrics
             Dashboard       metrics                

  TC_015     Download        Verify agent can       File downloads
             Brochure        download resources     successfully

  TC_016     Generate        Verify report          Report generated and
             Commission      generation and export  downloaded
             Report          works                  

  TC_017     Update Bank     Verify agent can       Details updated
             Details         update bank info       successfully
                             securely               

  TC_018     Complete        Verify progress        Status updated to
             Training Module tracking for training  "Completed"
                             modules                

  TC_019     Raise Support   Verify ticket          Ticket created with
             Ticket          submission works       reference number

  TC_020     Upload KYC      Verify secure upload   Document stored
             Document        of KYC documents       securely
  -------------------------------------------------------------------------

## **6. Appendices**

The Following Documents attached below can be used.
