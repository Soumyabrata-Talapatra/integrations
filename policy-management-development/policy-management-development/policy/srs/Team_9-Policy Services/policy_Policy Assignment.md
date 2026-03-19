**DEPARTMENT OF POSTS**

**MINISTRY OF COMMUNICATIONS & IT**

**GOVERNMENT OF INDIA**

**System Requirements Specification (SRS)**

Assignment /Reassignment

## Brief Description

The purpose of this document is to detail the policy Assignment process

### 

### **Process Flow**

- Data Indexer (Request Indexing at Post office)

- Document Scanner (for scanning the documents received at the CPC)

- Data Entry Operator (Verifies the details against the pre populated
  details by OCR.

- Non-Financial Transaction Approver (Who approves all Non-financial
  requests)

# Use Case/Activity Diagram

The Use case/Activity Diagram below depicts how the Assignment process
will be carried out

- **Policy Assignment Use Case / Activity Diagram**

**Use case Table -- Policy Assignment:**

  --------------------------------------------------------------------------
  **Use case ID**       02
  --------------------- ----------------------------------------------------
  **Use case title**    Policy Assignment

  **Objective/          Processing a change of assignment request received
  Purpose**             

  **Assumption**        None

  **Initiating Actor**  Customer

  **Initiating Event**  Assignment request Submitted along with the required
                        documents

  **Pre-conditions**    Applicant has filled the required details in the
                        application and attaches the required documents.

  **Post-conditions**   Decision is taken either to accept or decline the
                        assignment request

  **Associated Use      None
  case**                

  **Error/Success       Certain documents are not provided along with the
  conditions**          request.

  **Frequency**         Whenever the requirements are found missing

  **Criticality**       Major

  **Risk**              High
  --------------------------------------------------------------------------

![](media/image1.emf)

**Missing Requirements: Use Case/Activity Diagram:**

**Use Case Table -- Missing Requirements:**

  -----------------------------------------------------------------------
  **Use case ID**       03
  --------------------- -------------------------------------------------
  **Use case Name**     Use Case -- Activity Diagram\_ DoP\_ Missing
                        Requirements

  **Use case title**    Missing Requirements

  **Objective/          Missing requirement or information for assignment
  Purpose**             request

  **Assumption**        There are a few missing requirement identified
                        while processing the request

  **Initiating Actor**  CPC

  **Initiating Event**  Approver identifies if any requirements are
                        missing before a request is approved.

  **Pre-conditions**    There are a few missing requirement identified
                        while processing the request

  **Post-conditions**   Missing requirement have been received

  **Associated Use      None
  case**                

  **Error/Success       Missing requirement received is success
  conditions**          condition.

  **Frequency**         Whenever the requirements are found missing

  **Criticality**       Medium

  **Risk**              High
  -----------------------------------------------------------------------

**Use Case -- Activity Diagram\_ DoP\_ Missing Requirements**

![A diagram of a company AI-generated content may be
incorrect.](media/image2.png){width="6.2672101924759405in"
height="4.892090988626422in"}

# Requirements

+----------+-------------+------------------------------+---------------------+
| **S.No** | **RFP       | **Requirement Description**  | **Solution**        |
|          | Requirement |                              |                     |
|          | No.**       |                              |                     |
+==========+=============+==============================+=====================+
| 1.       | SR-ASG-1    | Must generate a unique       | Unique Customer ID  |
|          |             | customer id for every policy | for Policy owner,   |
|          |             | owner, life insured,         | life insured,       |
|          |             | nominee/ beneficiary/        | nominee, assignee,  |
|          |             | assignee and for every       | appointee would be  |
|          |             | person who submitted a       | generated           |
|          |             | proposal and was entered in  |                     |
|          |             | the IMA                      |                     |
+----------+-------------+------------------------------+---------------------+
| 2.       | SR-ASG-2    | Must have functionality to   | Physical            |
|          |             | enable endorsements on a     | Endorsement is a    |
|          |             | policy like                  | manual process. IMA |
|          |             |                              | should be capable   |
|          |             | Change in address\           | of providing        |
|          |             | Assignment of a policy       | alterations         |
|          |             |                              | mentioned through   |
|          |             |                              | policy alterations, |
|          |             |                              | assignment.         |
+----------+-------------+------------------------------+---------------------+
| 3.       | SR-ASG-3    | System must have ability to  | System sends emails |
|          |             | automatically send email &   | and sms reminder    |
|          |             | SMS reminders (through       | through integration |
|          |             | interface with CIM           | to sent to the user |
|          |             | application) to the Insured/ | who had requested   |
|          |             | for providing the pending    | through integration |
|          |             | documents (if any) or should |                     |
|          |             | have ability to send         |                     |
|          |             | reminders to concerned       |                     |
|          |             | authority to get in touch    |                     |
|          |             | with the Customer            |                     |
+----------+-------------+------------------------------+---------------------+
| 4.       | SR-ASG-5    | System must have the ability | System captures     |
|          |             | to record all                | task re-assignment  |
|          |             | re-assignments/transfers in  | or transfer in      |
|          |             | history.                     | workflow            |
+----------+-------------+------------------------------+---------------------+
| 5.       | SR-ASG-5    | Must validate the            | Field level         |
|          |             | completeness of data.        | validations to      |
|          |             |                              | ensure the data     |
|          |             |                              | entered is within   |
|          |             |                              | the expected range. |
+----------+-------------+------------------------------+---------------------+
| 6.       | SR-ASG-6    | Must provide executive       | Non-Financial       |
|          |             | access to all policy details | transaction         |
|          |             | (e.g. Coverage, terms &      | approver will be    |
|          |             | condition, Nominee details,  | having executive    |
|          |             | Guaranteed surrender value). | access to the       |
|          |             |                              | screens             |
+----------+-------------+------------------------------+---------------------+

# Functionality Description

Policy Assignment is a process that are carried out by a Policyholder
during the Life cycle of a Policy.

**What is Policy Assignment?**

Contrary to Nomination, Assignment of a Policy Transfers rights, title
and interest of the life insurance policy to a person or persons/Trust
or trusts/Company or Companies.

\'Assignor\' is the policyholder who transfers the title, and
\'Assignee\' is the person who gets the title from the assignor.
Therefore in an Assignment process there is a transfer of some or all of
the financial interest from the Assignor to the Assignee.

The assignment can be of two types --- conditional and absolute. An
absolute assignment /conditional assignment are usually done for a
monetary consideration. In absolute assignment the policyholder loses
all his rights over the policy proceeds and the Policy can only be
reassigned to him through a written consent of the Assignee.

However under the Conditional Assignment the Assignment is contingent
upon a Condition and the policy is reassigned to the Policyholder once
the Condition is met.

One can typically come across a conditional assignment where the
policyholder is trying to use the life insurance policy as collateral
against a loan he intends to raise. Once all dues are paid against the
Policy the policy is automatically reassigned to the Assignor.

Assignment has to be in writing and a notice affecting the same has to
be given to Life Insurance Company.

In the event of death of the absolute assignee the rights under the
policy delve on the legal heirs of the assignee. It can only be
reassigned. Assignment Section 38 of insurance act.

# Functional Specification

## Policy Assignment

The Assignment Process begins when a customer decides to assign his
Policy to an Assignee for a material consideration**. At a time there
can be only one Active Assignment for a Policy**.

Upon Screen Landing the user will be presented with the screen with
Policy Summary along with Policy Assignment Section by selecting

For the assignment requests also, Screen Indexing will remain same which
is explained in section 7.1.1

Data Entry Operator will navigate to the Assignment Change screen by
selecting:

Policy Maintenance 🡪 Assignment Change Option

![](media/image3.png){width="6.5in" height="2.2604166666666665in"}

On the Policy user can add an Assignee if there was no prior assignment.
Modification is applicable only on existing assignment.(Reassignment)

Assignee Type drop down includes the following options:

- Individual

- Company

- Trust

Upon selecting the Assignee Type, user will be enabled to enter the
respective details accordingly.

**Add Assignee**

###  **Add Assignee-Individual**

![](media/image4.png){width="6.375in" height="2.25in"}

![](media/image5.png){width="6.40625in" height="3.9583333333333335in"}

The option "Individual" is selected by default. To enter the data on the
Screen the User will double click on the empty row , then all the fields
in the screen will be presented to the user.

User will enter the details by looking at the documents using the **View
Documents** button. The Click will allow the User to access Assignment
form and any other forms which are scanned and saved in the File Server
of Enterprise Content management.

Following fields need to be filled:

- Customer ID

- First name

- Last Name

- Assignment Status

- Type of Assignment

- Consideration Amount(In case the assignment type is Conditional )

- Relationship to Insured

- Date of Birth

- UID Number

- PAN Number

- Assignee Address(Multiple fields in Address are explained in the
  control table below.)

- Assignee Email

- Assignee Phone Number

- Assignment Effective date

**Control Table for Add Assignee Details- Individual:**

+----------------+---------------+---------------+------------------------+---------------+---------------+
| **Field Name** | **Control     | **Data Type** | **Description**        | **Default     | **Default     |
|                | Type**        |               |                        | State**       | Value**       |
+================+===============+===============+========================+===============+===============+
| Customer ID    | Text box      | String        | Gives the Customer ID  | Populated by  | N/A.          |
|                |               |               | number                 | System        |               |
+----------------+---------------+---------------+------------------------+---------------+---------------+
| \*\*First name | Text box      | Varchar       | Gives first name of    | Enabled and   | N/A           |
|                |               |               | the assignee           | blank.        |               |
+----------------+---------------+---------------+------------------------+---------------+---------------+
| \*\* Last Name | Text box      | Varchar       | Gives last name of the | Enabled and   | N/A           |
|                |               |               | assignee               | blank.        |               |
+----------------+---------------+---------------+------------------------+---------------+---------------+
| Assignment     | Label         | Varchar       | Contains the status of | Disabled and  | N/A           |
| Status         |               |               | the assignment on      | Populated by  |               |
|                |               |               | given date             | system.       |               |
|                |               |               |                        |               |               |
|                |               |               | The fields include     |               |               |
|                |               |               | Assigned\              |               |               |
|                |               |               | Reassigned             |               |               |
+----------------+---------------+---------------+------------------------+---------------+---------------+
| \*\*Type of    | Dropdown      | Varchar       | Dropdown with Type of  | Enabled and   | N/A           |
| Assignment     |               |               | assignment             | blank.        |               |
|                |               |               | (Absolute/Conditional) |               |               |
+----------------+---------------+---------------+------------------------+---------------+---------------+
| Consideration  | Text box      | Numeric       | Gives the amount       | Enabled only  | N/A           |
| amount         |               |               | applicable to the      | when Type of  |               |
|                |               |               | assignee               | Assignment    |               |
|                |               |               |                        | selected is   |               |
|                |               |               |                        | Conditional   |               |
+----------------+---------------+---------------+------------------------+---------------+---------------+
| \*\*Date of    | Text box      | Date          | Gives the Birth Date   | Enabled and   | N/A           |
| Birth\*        |               |               | of the Nominee         | blank.        |               |
+----------------+---------------+---------------+------------------------+---------------+---------------+
| UID Number     | Text Box      | String        | Gives the UID Number   | Enabled and   | N/A           |
|                |               |               | of the assignee        | blank.        |               |
+----------------+---------------+---------------+------------------------+---------------+---------------+
| PAN Number     | Text Box      | String        | Givs the PAN number of | Enabled and   | N/A           |
|                |               |               | the Assignee           | Blank         |               |
+----------------+---------------+---------------+------------------------+---------------+---------------+
| **Address**                                                                                             |
+----------------+---------------+---------------+------------------------+---------------+---------------+
| \*\*Address    | Text Box      | Multiple      | Gives the address of   | Enabled and   | N/A           |
| Line 1         |               |               | the assignee.          | Blank.        |               |
+----------------+---------------+---------------+------------------------+---------------+---------------+
| \*\*Address    | Text Box      | Multiple      | Gives the address      | Enabled and   | N/A           |
| Line 2         |               |               |                        | Blank.        |               |
+----------------+---------------+---------------+------------------------+---------------+---------------+
| \*\*Address    | Text Box      | Multiple      | Gives the address      | Enabled and   | N/A           |
| Line 3         |               |               |                        | Blank.        |               |
+----------------+---------------+---------------+------------------------+---------------+---------------+
| Village        | Text Box      | Varchar       | Gives the Village name | Enabled and   | N/A           |
|                |               |               | (If applicable)        | Blank.        |               |
+----------------+---------------+---------------+------------------------+---------------+---------------+
| Taluka         | Text Box      | Varchar       | Gives the Taluka name  | Enabled and   | N/A           |
|                |               |               | (If applicable)        | Blank.        |               |
+----------------+---------------+---------------+------------------------+---------------+---------------+
| **\*\***City   | Text Box      | Varchar       | Gives the City name    | Enabled and   | N/A           |
|                |               |               |                        | Blank.        |               |
+----------------+---------------+---------------+------------------------+---------------+---------------+
| \*\*District   | Textbox       | Varchar       | Gives the district     | Enabled and   | N/A           |
|                |               |               | name                   | Blank.        |               |
+----------------+---------------+---------------+------------------------+---------------+---------------+
| \*\*State      | Dropdown      | Varchar       | Gives the List of      | Enabled and   | N/A           |
|                |               |               | States.                | Blank.        |               |
+----------------+---------------+---------------+------------------------+---------------+---------------+
| \*\*Country    | Dropdown      | Varchar       | Gives the country name | Enabled and   | N/A           |
|                |               |               |                        | blank.        |               |
+----------------+---------------+---------------+------------------------+---------------+---------------+
|                                                                                                         |
+----------------+---------------+---------------+------------------------+---------------+---------------+
| Email          | Textbox       | Multiple      | Gives the email        | Enabled and   | N/A           |
|                |               |               | address of the nominee | blank.        |               |
+----------------+---------------+---------------+------------------------+---------------+---------------+
| \*\*Phone      | Text Box      | Numeric       | Gives the phone number | Enabled and   | N/A           |
| Number         |               |               | of the nominee         | blank.        |               |
|                |               |               | including STD code (if |               |               |
|                |               |               | available)             |               |               |
+----------------+---------------+---------------+------------------------+---------------+---------------+
| \*\*Assignment | Text box      | Date          | Gives the date of      | Enabled and   | N/A           |
| Effective      |               |               | effective date of the  | blank.        |               |
| Date\*         |               |               | assignment             |               |               |
+----------------+---------------+---------------+------------------------+---------------+---------------+

\*Date format will be dd/mm/yyyy

\*\* Mandatory Fields

**[Hyperlinks, Buttons and Options:]{.underline}**

**Search --** Button: Allows to search for any Customer details based on
the search criteria (First Name, Last Name etc) If the Customer record
is present in the system it is tagged to the Customer ID and the same
will be retrieved. If Customer ID is not present system will generate
the Customer ID for the respective Customer will be generated
automatically by system once the Assignment is approved

When user enters any of the details to search for the customer id. He
can enter of the details and click on "Search" button. Upon clicking on
Search button, a pop up opens with available clients list as shown
below.

![](media/image6.png){width="5.927083333333333in"
height="2.7916666666666665in"}

**Save and Submit** -- Button- Allows saving and submitting for
approval.

**View Documents --** Button- Allows to view the documents (assignment
form, the notice of assignment and the Policy Bond.)

**Assignment Approval Screen (Individual):**

Once the Non-financial Approver clicks on the work item, he will be
navigated to the approval screen.

Once the approval request comes to the Non-Financial Transaction
approver, he can take a decision on the request received and can provide
any comments in the Approver Comments Section. The below shown screen is
approval screen for Policy Assignment (Individual).The complete workflow
on how the case is moved to Approver will be dealt in separate SRS. (SRS
\_OTH_001_Non Functional Requirements).

Non-Financial Approver will be navigated to the respective screen
through the Inbox (which is explained above)

![](media/image7.png){width="6.052083333333333in"
height="5.166666666666667in"}

**[Control Table:]{.underline}**

The below control table shows what are the fields available for
non-financial Transaction approver to edit.

+----------------+---------------+---------------+------------------------+---------------+---------------+
| **Field Name** | **Control     | **Data Type** | **Description**        | **Default     | **Default     |
|                | Type**        |               |                        | State**       | Value**       |
+================+===============+===============+========================+===============+===============+
| Customer ID    | Text box      | String        | Gives the Customer ID  | Populated by  | N/A.          |
|                |               |               | number                 | System        |               |
+----------------+---------------+---------------+------------------------+---------------+---------------+
| First name     | Text box      | Varchar       | Gives first name of    | Enabled       | N/A           |
|                |               |               | the assignee           |               |               |
+----------------+---------------+---------------+------------------------+---------------+---------------+
| Last Name      | Text box      | Varchar       | Gives last name of the | Enabled.      | N/A           |
|                |               |               | assignee               |               |               |
+----------------+---------------+---------------+------------------------+---------------+---------------+
| Assignment     | Label         | Varchar       | Contains the status of | Disabled and  | N/A           |
| Status         |               |               | the assignment on      | Populated by  |               |
|                |               |               | given date             | system.       |               |
|                |               |               |                        |               |               |
|                |               |               | The fields include     |               |               |
|                |               |               | Assigned\              |               |               |
|                |               |               | Reassigned             |               |               |
+----------------+---------------+---------------+------------------------+---------------+---------------+
| \*\*Type of    | Dropdown      | Varchar       | Dropdown with Type of  | Enabled       | N/A           |
| Assignment     |               |               | assignment             |               |               |
|                |               |               | (Absolute/Conditional) |               |               |
+----------------+---------------+---------------+------------------------+---------------+---------------+
| Consideration  | Text box      | Numeric       | Gives the amount       | Enabled only  | N/A           |
| amount         |               |               | applicable to the      | when Type of  |               |
|                |               |               | assignee               | Assignment    |               |
|                |               |               |                        | selected is   |               |
|                |               |               |                        | Conditional   |               |
+----------------+---------------+---------------+------------------------+---------------+---------------+
| Date of        | Text box      | Date          | Gives the Birth Date   | Enabled       | N/A           |
| Birth\*        |               |               | of the Nominee         |               |               |
+----------------+---------------+---------------+------------------------+---------------+---------------+
| UID Number     | Text Box      | String        | Gives the UID Number   | Enabled       | N/A           |
|                |               |               | of the assignee        |               |               |
+----------------+---------------+---------------+------------------------+---------------+---------------+
| PAN Number     | Text Box      | String        | Gives the PAN number   | Enabled       | N/A           |
|                |               |               | of the Assignee        |               |               |
+----------------+---------------+---------------+------------------------+---------------+---------------+
| **Address**                                                                                             |
+----------------+---------------+---------------+------------------------+---------------+---------------+
| Address Line 1 | Text Box      | Multiple      | Gives the address of   | Enabled       | N/A           |
|                |               |               | the assignee.          |               |               |
+----------------+---------------+---------------+------------------------+---------------+---------------+
| Address Line 2 | Text Box      | Multiple      | Gives the address      | Enabled       | N/A           |
+----------------+---------------+---------------+------------------------+---------------+---------------+
| Address Line 3 | Text Box      | Multiple      | Gives the address      | Enabled       | N/A           |
+----------------+---------------+---------------+------------------------+---------------+---------------+
| Village        | Text Box      | Varchar       | Gives the Village name | Enabled       | N/A           |
|                |               |               | (If applicable)        |               |               |
+----------------+---------------+---------------+------------------------+---------------+---------------+
| Taluka         | Text Box      | Varchar       | Gives the Taluka name  | Enabled       | N/A           |
|                |               |               | (If applicable)        |               |               |
+----------------+---------------+---------------+------------------------+---------------+---------------+
| **\*\***City   | Text Box      | Varchar       | Gives the City name    | Enabled       | N/A           |
+----------------+---------------+---------------+------------------------+---------------+---------------+
| \*\*District   | Textbox       | Varchar       | Gives the district     | Enabled       | N/A           |
|                |               |               | name                   |               |               |
+----------------+---------------+---------------+------------------------+---------------+---------------+
| \*\*State      | Dropdown      | Varchar       | Gives the List of      | Enabled       | N/A           |
|                |               |               | States.                |               |               |
+----------------+---------------+---------------+------------------------+---------------+---------------+
| \*\*Country    | Dropdown      | Varchar       | Gives the country name | Enabled       | N/A           |
+----------------+---------------+---------------+------------------------+---------------+---------------+
|                                                                                                         |
+----------------+---------------+---------------+------------------------+---------------+---------------+
| Email          | Textbox       | Multiple      | Gives the email        | Enabled       | N/A           |
|                |               |               | address of the nominee |               |               |
+----------------+---------------+---------------+------------------------+---------------+---------------+
| \*\*Phone      | Text Box      | Numeric       | Gives the phone number | Enabled       | N/A           |
| Number         |               |               | of the nominee         |               |               |
|                |               |               | including STD code (if |               |               |
|                |               |               | available)             |               |               |
+----------------+---------------+---------------+------------------------+---------------+---------------+
| \*\*Assignment | Text box      | Date          | Gives the date of      | Disabled      | N/A           |
| Effective      |               |               | effective date of the  |               |               |
| Date\*         |               |               | assignment             |               |               |
+----------------+---------------+---------------+------------------------+---------------+---------------+

**[Hyperlinks, Buttons and Options:]{.underline}**

**Back** -- Button : Allows going back to the Approver dashboard

**Approve** -- Button : Clicking on Approve button, the Assignment
request will be approved.

**Reject --** Button: Allows rejecting the assignment request. In case
of reject button is clicked, then system will give a message "Are you
sure to reject the Assignment Request" with Yes and No Options. If Yes
is clicked, then Assignment Rejection Letter generation trigger will be
initiated.

**Send for Print** - Button: Allows printing the Endorsement letter for
Assignment .It will be enabled when the Assignment is approved or
rejected.

**View Documents**: Allows viewing the documents

**Request for Documents**: If the approver needs more documents, then he
can request for further documents.

### **Add Assignee -Company**

Selecting the option Company from the dropdown "Assignee Type", allows
user to enter the details of the Company. To enter the details user will
click on the **View Documents** button. The Click will allow the User
access to Assignment form which is scanned and saved in the File Server
of Enterprise Content management.

![](media/image8.png){width="6.5in" height="4.09375in"}

When user double clicks on the empty row, he will be allowed to enter
the details in the screen:

Following fields need to be filled:

- Customer ID

- Company Name

- Assignment Status

- Type of Assignment

- Consideration Amount(In case of Conditional Assignment)

- Company Address (Multiple fields in Address are explained in the
  control table below.)

- Phone Number

- Company email

- Type of Company

- Assignment Effective date

**Search --** Button: Allows searching for any Customer details based on
the search criteria (Company Name, etc). If the Customer record is
present in the system it is tagged to the Customer ID and the same will
be retrieved. If Customer ID is not present in the system, in that case
once the assignment request is approved, then system auto assigns a
Customer ID to the client.

**View Documents --** Button- Allows to view the documents (like
Assignment form etc.)

**Save and Submit --**Button- Used to save the details and submit for
approval

**Control Table for Add Assignee -- Company**

+---------------+---------------+---------------+------------------------+---------------+---------------+
| **Field       | **Control     | **Data Type** | **Description**        | **Default     | **Default     |
| Name**        | Type**        |               |                        | State**       | Value**       |
+===============+===============+===============+========================+===============+===============+
| Customer ID   | Text Box      | String        | Gives the Customer ID  | Populated by  | N/A           |
|               |               |               | number                 | system        |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| \*\*Company   | Text Box      | Varchar       | Gives the company      | Enabled.      | N/A           |
| name          |               |               | name.                  |               |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| Assignment    | Dropdown      | Varchar       | Contains the status of | Disabled and  | N/A           |
| Status        |               |               | the assignment on      | Populated by  |               |
|               |               |               | given date. The        | system        |               |
|               |               |               | dropdown includes the  |               |               |
|               |               |               | following options :\   |               |               |
|               |               |               | Assigned\              |               |               |
|               |               |               | Reassigned             |               |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| \*\*Type of   | Dropdown      | Varchar       | Dropdown with Type of  | Enabled.      | N/A           |
| Assignment    |               |               | assignment             |               |               |
|               |               |               | (Absolute/Conditional) |               |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| Consideration | Text Box      | Numeric       | Gives the amount       | Enabled when  | N/A           |
| amount        |               |               | applicable to the      | Type of       |               |
|               |               |               | assignee in case of    | Assignment is |               |
|               |               |               | Conditional            | Conditional.  |               |
|               |               |               | Assignment.            |               |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| **Address**                                                                                            |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| Address Line  | Text Box      | Multiple      | Gives the address of   | Enabled.      | N/A           |
| 1             |               |               | the assignee.          |               |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| Address Line  | Text Box      | Multiple      | Gives the address      | Enabled.      | N/A           |
| 2             |               |               |                        |               |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| \*Address     | Text Box      | Multiple      | Gives the address      | Enabled.      | N/A           |
| Line 3        |               |               |                        |               |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| Village       | Text Box      | Varchar       | Gives the Village name | Enabled and   | N/A           |
|               |               |               | (If available)         | Blank.        |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| Taluka        | Text Box      | Varchar       | Gives the Taluka name  | Enabled and   | N/A           |
|               |               |               | (If available)         | Blank.        |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| \*\*City      | Text Box      | Varchar       | Gives the City name    | Enabled and   | N/A           |
|               |               |               |                        | Blank.        |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| \*\*District  | Textbox       | Varchar       | Gives the district     | Enabled and   | N/A           |
|               |               |               | name                   | Blank.        |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| \*\*State     | Dropdown      | Varchar       | Gives the List of      | Enabled and   | N/A           |
|               |               |               | States.                | Blank.        |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| \*\*Country   | Dropdown      | Varchar       | Gives the country name | Enabled and   | N/A           |
|               |               |               |                        | blank.        |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
|                                                                                                        |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| \*\*Phone     | Text Box      | Numeric       | Gives the phone number | Enabled and   | N/A           |
| Number        |               |               | details                | blank.        |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| \*\*Company   | Text Box      | Multiple      | Gives the Email        | Enabled and   | N/A           |
| email         |               |               | address of the Company | blank.        |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| \*\*Type of   | Dropdown      | Varchar       | Describes the nature   | Enabled and   | N/A           |
| Company       |               |               | of company. Options    | blank.        |               |
|               |               |               | include                |               |               |
|               |               |               |                        |               |               |
|               |               |               | Public Company         |               |               |
|               |               |               |                        |               |               |
|               |               |               | Private Company        |               |               |
|               |               |               |                        |               |               |
|               |               |               | Partnership Firm       |               |               |
|               |               |               |                        |               |               |
|               |               |               | Sole Proprietor ship   |               |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| Assignment    | Text box      | Date          | Gives the effective    | Enabled and   | N/A           |
| Effective     |               |               | date of the assignment | blank.        |               |
| date\*        |               |               |                        |               |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+

\*Date format will be dd/mm/yyyy

\*\* Mandatory fields.

**Approval Screen --Add Assignee -- Company**

Once the Approver clicks on the work item from his dashboard.
Non-Financial Transaction approver will be navigated to the approval
screen where he can take the decision on the request received and
provide comments in the Approver Comments section. The below shown is
approval screen for Policy Assignment (Company). The complete workflow
on how the case is moved to Approver will be dealt in separate SRS. (SRS
\_OTH_001_Non Functional Requirements.) Non-Financial Approver will be
navigated to the approval screen through the Inbox (which is explained
above)

![](media/image9.png){width="6.260416666666667in"
height="5.010416666666667in"}

For the Non-Financial Transaction Approver, the screen will be in
editable mode.

**[Control Table :]{.underline}**

Below shown control table shows what are the fields are editable.

The below control table gives what are the fields available for the
Non-Financial Transaction Approver.

+---------------+---------------+---------------+------------------------+---------------+---------------+
| **Field       | **Control     | **Data Type** | **Description**        | **Default     | **Default     |
| Name**        | Type**        |               |                        | State**       | Value**       |
+===============+===============+===============+========================+===============+===============+
| Customer ID   | Text Box      | String        | Gives the Customer ID  | Populated by  | N/A           |
|               |               |               | number                 | system        |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| \*\*Company   | Text Box      | Varchar       | Gives the company      | Enabled.      | N/A           |
| name          |               |               | name.                  |               |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| Assignment    | Dropdown      | Varchar       | Contains the status of | Disabled and  | N/A           |
| Status        |               |               | the assignment on      | Populated by  |               |
|               |               |               | given date. The        | system        |               |
|               |               |               | dropdown includes the  |               |               |
|               |               |               | following options :\   |               |               |
|               |               |               | Assigned\              |               |               |
|               |               |               | Reassigned             |               |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| \*\*Type of   | Dropdown      | Varchar       | Dropdown with Type of  | Enabled.      | N/A           |
| Assignment    |               |               | assignment             |               |               |
|               |               |               | (Absolute/Conditional) |               |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| Consideration | Text Box      | Numeric       | Gives the amount       | Enabled when  | N/A           |
| amount        |               |               | applicable to the      | Type of       |               |
|               |               |               | assignee in case of    | Assignment is |               |
|               |               |               | Conditional            | Conditional.  |               |
|               |               |               | Assignment.            |               |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| **Address**                                                                                            |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| Address Line  | Text Box      | Multiple      | Gives the address of   | Enabled.      | N/A           |
| 1             |               |               | the assignee.          |               |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| Address Line  | Text Box      | Multiple      | Gives the address      | Enabled.      | N/A           |
| 2             |               |               |                        |               |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| Address Line  | Text Box      | Multiple      | Gives the address      | Enabled.      | N/A           |
| 3             |               |               |                        |               |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| Village       | Text Box      | Varchar       | Gives the Village name | Enabled.      | N/A           |
|               |               |               | (If available)         |               |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| Taluka        | Text Box      | Varchar       | Gives the Taluka name  | Enabled.      | N/A           |
|               |               |               | (If available)         |               |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| City          | Text Box      | Varchar       | Gives the City name    | Enabled.      | N/A           |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| District      | Textbox       | Varchar       | Gives the district     | Enabled.      | N/A           |
|               |               |               | name                   |               |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| State         | Dropdown      | Varchar       | Gives the List of      | Enabled.      | N/A           |
|               |               |               | States.                |               |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| Country       | Dropdown      | Varchar       | Gives the country name | Enabled.      | N/A           |
+---------------+---------------+---------------+------------------------+---------------+---------------+
|                                                                                                        |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| Phone Number  | Text Box      | Numeric       | Gives the phone number | Enabled.      | N/A           |
|               |               |               | details                |               |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| Company email | Text Box      | Multiple      | Gives the Email        | Enabled.      | N/A           |
|               |               |               | address of the Company |               |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| Type of       | Dropdown      | Varchar       | Describes the nature   | Enabled.      | N/A           |
| Company       |               |               | of company. Options    |               |               |
|               |               |               | include                |               |               |
|               |               |               |                        |               |               |
|               |               |               | Public Company         |               |               |
|               |               |               |                        |               |               |
|               |               |               | Private Company        |               |               |
|               |               |               |                        |               |               |
|               |               |               | Partnership Firm       |               |               |
|               |               |               |                        |               |               |
|               |               |               | Sole Proprietor ship   |               |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| Assignment    | Text box      | Date          | Gives the effective    | Disabled.     | N/A           |
| Effective     |               |               | date of the assignment |               |               |
| date\*        |               |               |                        |               |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+

**[Hyperlinks, Buttons and Options :]{.underline}**

**Back** -- Button : Allows going back to the approver dashboard screen.

**Approve** -- Button : Clicking on Approve button, the Assignment
request will be approved.

**Reject --** Button: Allows rejecting the assignment request. In case
of reject button is clicked, then system will give a message "Are you
sure to reject the Assignment Request" with Yes and No Options. If Yes
is clicked, then Assignment Rejection Letter generation trigger will be
initiated.

**Send for Print** - Button: Allows printing the Endorsement letter for
Assignment .It will be enabled in case the Assignment request is
approved or rejected.

**View Documents** -Button: Allows viewing the documents submitted.

**Request for Documents** -- Button: Allows requesting for documents, if
required.

### **Add Assignee-Trust**

Selecting the option "Trust" from the "Assignee Type" drop down list
allows user to enter the trust details. To enter the data on the Screen
the User will click on the **View Documents** button. The Click will
allow the User access to Assignment form which is scanned and saved in
the File Server of Enterprise Content management.

![](media/image10.png){width="6.34375in" height="4.0625in"}

When user double clicks on the empty row, he will be allowed to enter
the details in the screen:

Following fields need to be filled:

- Customer ID

- Trust Name

- Assignment Status

- Type of Assignment

- Consideration Amount(Applicable only when the Type of Assignment is
  Conditional)

- Trust registration Number

- Trust Address(Multiple fields in Address are explained in the control
  table below.)

- Policyholder Relationship to Trust

- Phone Number

- Assignment Effective date

**Search --** Button: Allows searching for any Customer details based on
the search criteria (Trust Name, etc.).. If the Customer record is
present in the system it is tagged to the Customer ID and the same will
be retrieved. If Customer ID is not present system will generate the
Customer ID for the respective Customer by using Generate Customer ID.
If customer ID is not present in the system, in that case once the
Assignment request is approved, then system will auto assign Customer ID
to the client.

**Save and Submit -** Button- Used to save the details and submit for
approval

**View Documents --** Button- Allows to view the documents (like
Assignment form etc.)

**Control Table for Add Assignee -- Trust**

+---------------+---------------+---------------+------------------------+---------------+---------------+
| **Field       | **Control     | **Data Type** | **Description**        | **UI          | **Default     |
| Name**        | Type**        |               |                        | Display**     | Value**       |
+===============+===============+===============+========================+===============+===============+
| Customer ID   | Text Box      | String        | Gives the Customer ID  | Populated by  | N/A           |
|               |               |               | number                 | system        |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| \*\*Trust     | Text Box      | Varchar       | Gives the trust name.  | Enabled and   | N/A           |
| name          |               |               |                        | blank.        |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| Assignment    | Dropdown      | Varchar       | Contains the status of | Disabled and  | N/A           |
| Status        |               |               | the assignment on      | Populated by  |               |
|               |               |               | given date             | system        |               |
|               |               |               |                        |               |               |
|               |               |               | The fields include     |               |               |
|               |               |               | Assigned\              |               |               |
|               |               |               | Reassigned             |               |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| \*\*Type of   | Dropdown      | Varchar       | Dropdown with Type of  | Enabled and   | N/A           |
| Assignment    |               |               | assignment             | blank.        |               |
|               |               |               | (Absolute/Conditional) |               |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| Consideration | Text Box      | Numeric       | Gives the amount       | Enabled only  | N/A           |
| amount        |               |               | applicable to the      | when Type of  |               |
|               |               |               | assignee in case of    | Assignment is |               |
|               |               |               | Conditional Assignment | conditional.  |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| \*\*Trust     | Text Box      | Multiple      | Gives unique           | Enabled and   | N/A           |
| Registration  |               |               | Identification Number  | blank.        |               |
| Number        |               |               | of the Trust           |               |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| **Address**                                                                                            |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| \*\*Address   | Text Box      | Multiple      | Gives the address of   | Enabled and   | N/A           |
| Line 1        |               |               | the assignee.          | Blank.        |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| \*\*Address   | Text Box      | Multiple      | Gives the address      | Enabled and   | N/A           |
| Line 2        |               |               |                        | Blank.        |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| \*\*Address   | Text Box      | Multiple      | Gives the address      | Enabled and   | N/A           |
| Line 3        |               |               |                        | Blank.        |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| Village       | Text Box      | Varchar       | Gives the Village name | Enabled and   | N/A           |
|               |               |               | (If available)         | Blank.        |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| Taluka        | Text Box      | Varchar       | Gives the Taluka name  | Enabled and   | N/A           |
|               |               |               | (If available)         | Blank.        |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| \*\*City      | Text Box      | Varchar       | Gives the City name    | Enabled and   | N/A           |
|               |               |               |                        | Blank.        |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| \*\*District  | Textbox       | Varchar       | Gives the district     | Enabled and   | N/A           |
|               |               |               | name                   | Blank.        |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| \*\*State     | Dropdown      | Varchar       | Gives the List of      | Enabled and   | N/A           |
|               |               |               | States.                | Blank.        |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| \*\*Country   | Dropdown      | Varchar       | Gives the country name | Enabled and   | N/A           |
|               |               |               |                        | blank.        |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| \*\*Policy    | Drop down     | Varchar       | Gives the relationship | Enabled and   | N/A           |
| holder        |               |               | of the policyholder to | blank.        |               |
| relationship  |               |               | the trust. The values  |               |               |
| to trust      |               |               | would be\              |               |               |
|               |               |               | Settler ,Beneficiary   |               |               |
|               |               |               | and Trustee            |               |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| \*\*Phone     | Text Box      | Numeric       | Gives the phone number | Enabled and   | N/A           |
| Number        |               |               | details                | blank.        |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| Assignment    | Text Box      | Date          | Gives the effective    | Enabled and   | N/A           |
| Effective     |               |               | date of the assignment | blank.        |               |
| date\*        |               |               |                        |               |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+

\*Date format will be dd/mm/yyyy

\*\* Mandatory fields

**Approval Screen --Add Assignee -- Trust:**

Once the approval request comes to the Non-Financial Transaction
approver, he can take decision on the request received and provide
comments in the Approver Comments section. In the approval screen the
screen will be in editable mode. The below shown is approval screen for
Policy Assignment (Trust) The complete workflow on how the case is moved
to Approver will be dealt in separate SRS. (SRS \_OTH_001_Non Functional
Requirements). Non-Financial Approver will be navigated to the approval
screen through the Inbox (which is explained above)

![](media/image11.png){width="6.489583333333333in" height="5.125in"}

For non-financial Transaction approver the fields are editable.

**[Control Table:]{.underline}**

Below control table explains what are the fields available for editing:

+---------------+---------------+---------------+------------------------+---------------+---------------+
| **Field       | **Control     | **Data Type** | **Description**        | **UI          | **Default     |
| Name**        | Type**        |               |                        | Display**     | Value**       |
+===============+===============+===============+========================+===============+===============+
| Customer ID   | Text Box      | String        | Gives the Customer ID  | Populated by  | N/A           |
|               |               |               | number                 | system        |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| Trust name    | Text Box      | Varchar       | Gives the trust name.  | Enabled.      | N/A           |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| Assignment    | Dropdown      | Varchar       | Contains the status of | Disabled and  | N/A           |
| Status        |               |               | the assignment on      | Populated by  |               |
|               |               |               | given date             | system        |               |
|               |               |               |                        |               |               |
|               |               |               | The fields include     |               |               |
|               |               |               | Assigned\              |               |               |
|               |               |               | Reassigned             |               |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| \*\*Type of   | Dropdown      | Varchar       | Dropdown with Type of  | Enabled.      | N/A           |
| Assignment    |               |               | assignment             |               |               |
|               |               |               | (Absolute/Conditional) |               |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| Consideration | Text Box      | Numeric       | Gives the amount       | Enabled only  | N/A           |
| amount        |               |               | applicable to the      | when Type of  |               |
|               |               |               | assignee in case of    | Assignment is |               |
|               |               |               | Conditional Assignment | conditional.  |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| Trust         | Text Box      | Multiple      | Gives unique           | Enabled.      | N/A           |
| Registration  |               |               | Identification Number  |               |               |
| Number        |               |               | of the Trust           |               |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| **Address**                                                                                            |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| Address Line  | Text Box      | Multiple      | Gives the address of   | Enabled.      | N/A           |
| 1             |               |               | the assignee.          |               |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| Address Line  | Text Box      | Multiple      | Gives the address      | Enabled.      | N/A           |
| 2             |               |               |                        |               |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| Address Line  | Text Box      | Multiple      | Gives the address      | Enabled.      | N/A           |
| 3             |               |               |                        |               |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| Village       | Text Box      | Varchar       | Gives the Village name | Enabled.      | N/A           |
|               |               |               | (If available)         |               |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| Taluka        | Text Box      | Varchar       | Gives the Taluka name  | Enabled.      | N/A           |
|               |               |               | (If available)         |               |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| City          | Text Box      | Varchar       | Gives the City name    | Enabled.      | N/A           |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| District      | Textbox       | Varchar       | Gives the district     | Enabled.      | N/A           |
|               |               |               | name                   |               |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| State         | Dropdown      | Varchar       | Gives the List of      | Enabled.      | N/A           |
|               |               |               | States.                |               |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| Country       | Dropdown      | Varchar       | Gives the country name | Enabled.      | N/A           |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| \*\*Policy    | Drop down     | Varchar       | Gives the relationship | Enabled.      | N/A           |
| holder        |               |               | of the policyholder to |               |               |
| relationship  |               |               | the trust. The values  |               |               |
| to trust      |               |               | would be\              |               |               |
|               |               |               | Settler ,Beneficiary   |               |               |
|               |               |               | and Trustee            |               |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| Phone Number  | Text Box      | Numeric       | Gives the phone number | Enabled.      | N/A           |
|               |               |               | details                |               |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| Assignment    | Text Box      | Date          | Gives the effective    | Disabled      | N/A           |
| Effective     |               |               | date of the assignment |               |               |
| date\*        |               |               |                        |               |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+

\*Date format would be "dd/mm/yyyy"

**[Hyperlinks, Buttons and Options :]{.underline}**

**Back** -- Button : Allows going back to Approver's Dashboard.

**Approve** -- Button : Clicking on Approve button, the Assignment
request will be approved.

**Reject --** Button: Allows rejecting the assignment request. In case
of reject button is clicked, then system will give a message "Are you
sure to reject the Assignment Request" with Yes and No Options. If Yes
is clicked, then Assignment Rejection Letter generation trigger will be
initiated.

**Send for Print** - Button: Allows printing the Endorsement letter for
Assignment .It will be enabled in case the Assignment request is
approved or rejected.

**View Documents -**Button: Allows viewing the documents submitted.

**Request for Documents --** Button: Allows requesting for further
documents

**Modifying Assignee Details(Reassignment)**

At a time there **will not** be more than one active assignment, Double
clicking on the Assignment section will allow to edit the details.

In case the Assignment type is "Absolute", then a written consent from
the Assignee is essential to perform another assignment.

In case of conditional assignment, the policy will be reassigned back to
the Insured.

When the Assignment status is changed to Reassigned, System enables to
add assignee on the policy

A new Assignment will only be allowed to be made if the policy
assignment status is reassigned.

### **7.2.2a Modifying Assignee Details --Individual(Reassignment)**

Upon double clicking on the existing row present in the Assignment
Section, Data Entry Operator will be allowed to edit the details

![](media/image12.png){width="6.479166666666667in"
height="4.447916666666667in"}

Following fields will be presented to the user to view/update:

- Customer ID

- First name

- Last Name

- Assignment Status

- Type of Assignment

- Consideration Amount

- Relationship to Insured

- Date of Birth

- UID Number

- PAN Number

- Address (Multiple fields in Address are explained in the control table
  below.)

- Email

- Phone Number

- Assignment Effective date

**Control Table for Modify Assignee Details -- Individual**

(Reassignment)

+---------------+---------------+---------------+------------------------+---------------+---------------+
| **Field       | **Control     | **Data Type** | **Description**        | **Default     | **Default     |
| Name**        | Type**        |               |                        | State**       | Value**       |
+===============+===============+===============+========================+===============+===============+
| Customer ID   | Text Box      | String        | Gives the Customer ID  | Populated by  | N/A           |
|               |               |               | number                 | system        |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| First name    | Text Box      | Varchar       | Gives first name of    | Enabled       | N/A           |
|               |               |               | the assignee           |               |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| Last Name     | Text Box      | Varchar       | Gives last name of the | Enabled       | N/A           |
|               |               |               | assignee               |               |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| Assignment    | Dropdown      | Varchar       | Contains the status of | Disabled and  | N/A           |
| Status        |               |               | the assignment on      | populated by  |               |
|               |               |               | given date             | system.       |               |
|               |               |               |                        |               |               |
|               |               |               | The fields include     |               |               |
|               |               |               | Assigned\              |               |               |
|               |               |               | Reassigned             |               |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| Type of       | Dropdown      | Varchar       | Dropdown with Type of  | Enabled.      | N/A           |
| Assignment    |               |               | assignment             |               |               |
|               |               |               | (Absolute/Conditional) |               |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| Consideration | Text Box      | Numeric       | Gives the amount       | Enabled       | N/A           |
| amount        |               |               | applicable to the      |               |               |
|               |               |               | assignee in case of    |               |               |
|               |               |               | Conditional Assignment |               |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| Relationship  | Drop down     | Varchar       | Gives the relationship | Enabled.      | N/A           |
| to Insured    |               |               | of the nominee to the  |               |               |
|               |               |               | policyholder           |               |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
|  Date of      | Text Box      | Date          | Gives the date of      | Enabled       | N/A           |
| Birth\*       |               |               | birth of the assignee  |               |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| UID Number    | Text Box      | String        | Gives the Unique       | Disabled      | N/A           |
|               |               |               | Identification Number  |               |               |
|               |               |               | of the assignee        |               |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| PAN Number    | Text Box      | String        | Gives the PAN Number   | Enabled       | N/A           |
|               |               |               | of the assignee        |               |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| Email         | Text Box      | Multiple      | Gives the email        | Enabled.      | N/A           |
|               |               |               | address of the         |               |               |
|               |               |               | assignee               |               |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| **Address**                                                                                            |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| Address Line  | Text Box      | Multiple      | Gives the address of   | Enabled.      | N/A           |
| 1             |               |               | the assignee.          |               |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| Address Line  | Text Box      | Multiple      | Gives the address      | Enabled.      | N/A           |
| 2             |               |               |                        |               |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| Address Line  | Text Box      | Multiple      | Gives the address      | Enabled.      | N/A           |
| 3             |               |               |                        |               |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| Village       | Text Box      | Varchar       | Gives the Village name | Enabled.      | N/A           |
|               |               |               | (If applicable)        |               |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| Taluka        | Text Box      | Varchar       | Gives the Taluka name  | Enabled.      | N/A           |
|               |               |               | (If applicable)        |               |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| City          | Text Box      | Varchar       | Gives the City name    | Enabled.      | N/A           |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| District      | Textbox       | Varchar       | Gives the district     | Enabled.      | N/A           |
|               |               |               | name                   |               |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| State         | Dropdown      | Varchar       | Gives the List of      | Enabled.      | N/A           |
|               |               |               | States.                |               |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| Country       | Dropdown      | Varchar       | Gives the country name | Enabled.      | N/A           |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| Phone Number  | Text Box      | Numeric       | Gives the phone number | Enabled.      | N/A           |
|               |               |               | details.               |               |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| Assignment    | Text Box      | Date          | Gives the date of      | Disabled      | N/A           |
| Effective     |               |               | expiry of the          |               |               |
| date\*        |               |               | assignment             |               |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+

\*Date format will be dd/mm/yyyy.

**[Hyperlinks, Buttons and Options:]{.underline}**

**Search --disabled**

**\
Save and Submit:** Allows saving the details and send for approval.
(Approval screen as shown above)

**View Documents --** Button- Allows to view the documents

### **Modifying Assignee Details -- Company(Reassignment)**

Upon double clicking on the Assignment Section user will be presented
with the Company details, if Company was added on the policy as
assignee. To update the details user will double click on the row
present in the Assignment Section, To update the details, user will
click on the **View Documents** button (if required). The Click will
allow the User to access documents like Request form which is scanned
and saved in the File Server of Enterprise Content management. User then
will update the details by referring the Request form etc. available.

![](media/image13.png){width="6.447916666666667in" height="4.21875in"}

Following fields will be presented to the user to view/update:

- Customer ID

- Company Name

- Assignment Status

- Type of Assignment

- Consideration Amount(In case of Assignment type is Conditional)

- Company Address (Multiple fields in Address are explained in the
  control table below.)

- Phone Number

- Company email

- Type of Company

- Assignment Effective date

**Control Table for Modify Assignee Details -- Company(Reassignment)**

+---------------+---------------+---------------+------------------------+---------------+---------------+
| **Field       | **Control     | **Data Type** | **Description**        | **Default     | **Default     |
| Name**        | Type**        |               |                        | State**       | Value**       |
+===============+===============+===============+========================+===============+===============+
| Customer ID   | Text Box      | String        | Gives the Customer ID  | Populated by  | N/A           |
|               |               |               | number                 | system        |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| \*\*Company   | Text Box      | Varchar       | Gives the company      | Enabled.      | N/A           |
| name          |               |               | name.                  |               |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| Assignment    | Dropdown      | Varchar       | Contains the status of | Disabled and  | N/A           |
| Status        |               |               | the assignment on      | populated by  |               |
|               |               |               | given date. The        | system        |               |
|               |               |               | dropdown includes the  |               |               |
|               |               |               | following options :\   |               |               |
|               |               |               | Assigned\              |               |               |
|               |               |               | Reassigned             |               |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| \*\*Type of   | Dropdown      | Varchar       | Dropdown with Type of  | Enabled.      | N/A           |
| Assignment    |               |               | assignment             |               |               |
|               |               |               | (Absolute/Conditional) |               |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| Consideration | Text Box      | Numeric       | Gives the amount       | Enabled only  | N/A           |
| amount        |               |               | applicable to the      | if the type   |               |
|               |               |               | assignee in case of    | of assignment |               |
|               |               |               | Conditional            | is            |               |
|               |               |               | Assignment.            | Conditional.  |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| **Address**                                                                                            |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| Address Line  | Text Box      | Multiple      | Gives the address of   | Enabled.      | N/A           |
| 1             |               |               | the assignee.          |               |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| Address Line  | Text Box      | Multiple      | Gives the address      | Enabled.      | N/A           |
| 2             |               |               |                        |               |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| Address Line  | Text Box      | Multiple      | Gives the address      | Enabled.      | N/A           |
| 3             |               |               |                        |               |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| Village       | Text Box      | Varchar       | Gives the Village name | Enabled.      | N/A           |
|               |               |               | (If applicable)        |               |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| Taluka        | Text Box      | Varchar       | Gives the Taluka name  | Enabled.      | N/A           |
|               |               |               | (If applicable)        |               |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| City          | Text Box      | Varchar       | Gives the City name    | Enabled.      | N/A           |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| District      | Textbox       | Varchar       | Gives the district     | Enabled.      | N/A           |
|               |               |               | name                   |               |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| State         | Dropdown      | Varchar       | Gives the List of      | Enabled.      | N/A           |
|               |               |               | States.                |               |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| Country       | Dropdown      | Varchar       | Gives the country name | Enabled.      | N/A           |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| Phone Number  | Text Box      | Numeric       | Gives the phone number | Enabled.      | N/A           |
|               |               |               | details.               |               |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| Company email | Text Box      | Multiple      | Gives the Email        | Enabled.      | N/A           |
|               |               |               | address of the Company |               |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| Type of       | Dropdown      | Varchar       | Describes the nature   | Enabled.      | N/A           |
| Company       |               |               | of company. Options    |               |               |
|               |               |               | include                |               |               |
|               |               |               |                        |               |               |
|               |               |               | Public Company         |               |               |
|               |               |               |                        |               |               |
|               |               |               | Private Company        |               |               |
|               |               |               |                        |               |               |
|               |               |               | Partnership Firm       |               |               |
|               |               |               |                        |               |               |
|               |               |               | Sole Proprietor ship   |               |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| Assignment    | Text box      | Date          | Gives the effective    | Disabled.     | N/A           |
| Effective     |               |               | date of the assignment |               |               |
| date\*        |               |               |                        |               |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+

\*Date format will be dd/mm/yyyy

**[Hyperlinks, Buttons and Options:]{.underline}**

**Search -- Disabled**

**Save and Submit:** Allows saving the details and submitting for
approval. (Approval screen as shown above)

**\
View Documents --** Button- Allows to view the documents.

### **Modify Assignee Details-Trust(Reassignment)**

Upon clicking on "Modify Assignee Details" user will be presented with
the trust details, if trust was added on the policy. To update the
details user will double click on the row present in the Assignment
Section To update the details, user will click on the **View Documents**
button (if required).The Click will allow the User to access documents
like Request form which is scanned and saved in the File Server of
Enterprise Content management. User then will update the details by
referring the Request form etc. available.

![](media/image14.png){width="6.395833333333333in"
height="4.354166666666667in"}

Following fields will be presented to the user to view/update:

- Customer ID

- Trust Name

- Assignment Status

- Type of Assignment

- Consideration Amount(In case of Conditional Assignment)

- Policyholder Relationship to Trust

- Trust registration Number

- Trust Address (Multiple fields in Address are explained in the control
  table below.)

- Trust Beneficiary

- Phone Number

- Assignment Effective date

**[Hyperlinks, Buttons and Options:]{.underline}**

**Search --** Disabled

**Save and Submit:** Allows saving the details and send for approval.
(Approval screen as shown above)

**View Documents --** Button- Allows to view the documents

**Control Table for Modify Assignee Details -- Trust (Reassignment)**

+---------------+---------------+---------------+------------------------+---------------+---------------+
| **Field       | **Control     | **Data Type** | **Description**        | **Default     | **Default     |
| Name**        | Type**        |               |                        | State**       | Value**       |
+===============+===============+===============+========================+===============+===============+
| Customer ID   | Text Box      | String        | Gives the Customer ID  | Disabled.     | N/A           |
|               |               |               | number                 |               |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| Trust name    | Text Box      | Varchar       | Gives the trust name.  | Enabled.      | N/A           |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| Assignment    | Dropdown      | Varchar       | Contains the status of | Disabled and  | N/A           |
| Status        |               |               | the assignment on      | populated by  |               |
|               |               |               | given date             | system        |               |
|               |               |               |                        |               |               |
|               |               |               | The fields include     |               |               |
|               |               |               | Assigned\              |               |               |
|               |               |               | Reassigned             |               |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| Trust         | Text Box      | Multiple      | Gives unique           | Enabled.      | N/A           |
| Registration  |               |               | Identification Number  |               |               |
| Number        |               |               | of the Trust           |               |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| Consideration | Text Box      | Numeric       | Gives the amount       | Enabled.      | N/A           |
| amount        |               |               | applicable to the      |               |               |
|               |               |               | assignee in case of    |               |               |
|               |               |               | Conditional Assignment |               |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| **Address**                                                                                            |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| Address Line  | Text Box      | Multiple      | Gives the address of   | Enabled.      | N/A           |
| 1             |               |               | the Trust.             |               |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| Address Line  | Text Box      | Multiple      | Gives the address      | Enabled.      | N/A           |
| 2             |               |               |                        |               |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| Address Line  | Text Box      | Multiple      | Gives the address      | Enabled.      | N/A           |
| 3             |               |               |                        |               |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| Village       | Text Box      | Varchar       | Gives the Village name | Enabled.      | N/A           |
|               |               |               | (if applicable)        |               |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| Taluka        | Text Box      | Varchar       | Gives the Taluka       | Enabled.      | N/A           |
|               |               |               | name(if applicable)    |               |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| City          | Text Box      | Varchar       | Gives the City name    | Enabled.      | N/A           |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| District      | Textbox       | Varchar       | Gives the district     | Enabled.      | N/A           |
|               |               |               | name                   |               |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| State         | Dropdown      | Varchar       | Gives the List of      | Enabled.      | N/A           |
|               |               |               | States.                |               |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| Country       | Dropdown      | Varchar       | Gives the country name | Enabled.      | N/A           |
+---------------+---------------+---------------+------------------------+---------------+---------------+
|                                                                                                        |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| Policy holder | Drop down     | Varchar       | Gives the relationship | Enabled.      | N/A           |
| relationship  |               |               | of the policyholder to |               |               |
| to trust      |               |               | the trust. The values  |               |               |
|               |               |               | would be\              |               |               |
|               |               |               | Settler ,Beneficiary   |               |               |
|               |               |               | and Trustee            |               |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| Phone Number  | Text Box      | Numeric       | Gives the phone number | Enabled.      | N/A           |
|               |               |               | details.               |               |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| Type of       | Dropdown      | Varchar       | Dropdown with Type of  | Enabled.      | N/A           |
| Assignment    |               |               | assignment             |               |               |
|               |               |               | (Absolute/Conditional) |               |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+
| Assignment    | Text Box      | Date          | Gives the effective    | Disabled.     | N/A           |
| Effective     |               |               | date of the assignment |               |               |
| date\*        |               |               |                        |               |               |
+---------------+---------------+---------------+------------------------+---------------+---------------+

\*Date format will be dd/mm/yyyy

### **View Past Assignments**

User can also have the option to view the past assignments (if made
any). Once user clicks on the button "View Past Assignments", a window
opens which shows all the past assignment details (if present). If there
are no Past Assignments, System will give a message "There are no past
assignments present on this Policy".

Once user clicks on the "View Past Assignments" and if there are any
past Assignee are present the new window shows the Past Assignments
details.

![](media/image15.png){width="6.375in" height="4.21875in"}

**[Hyperlinks, Buttons and Options:]{.underline}**

**Back --** Button: This button is used to navigate back to the main
screen where the Assignee details (added/modified.) can be viewed.

# 

# [Work Flow]{.underline}

At Post Office:

\- Customer submits all the documents like change request forms at the
post office.

-The Data Indexer will input the details in the service request screen
and a request id will be generated

At CPC:

Scanning Executive scans all the documents and sends them to Data Entry
Operator.

Data Entry Operator verifies the details against the pre populated Data
from OCR and modifies any details, if required.

The case is then forwarded to Non-Financial Transaction Approver for
approving .Approver is authorised to approve the new nominee/assignee
and any modifications to be approved.

# 

# 

# 

# Business Rules

1.  **ASG-BR001**-**General Rules- Assignment**

The Post Office Life Insurance Fund/RPOLIF does not prescribe any
particular forms for assignment or for nomination, or for the notice
thereof, and an application on plain paper can also be used for the
purpose.

2.  **ASG-BR002-General Rules- Assignment**

The assignment must be dated and signed by the assignor in the presence
of a witness.

3.  **ASG-BR003-General Rules-Assignment**

Assignment of policy as a whole may be made either in favour of one
person or jointly in favour of two or more persons.

4.  **ASG-BR004-Assignment as security**

No need be in form of separate deed for the case of an assignment in
favour of the President of India: It acts as security for the repayment
of any loan granted out of the Fund, an assignment, otherwise complete,
will be inoperative against the Fund, unless a notice in writing of the
assignment has been delivered to the Postmaster General/ Head of
Division.

5.  **ASG-BR005 --Notice of assignment**

The notice of assignment must be accompanied by the policy duly endorsed
or, where the assignment has been effected by a separate deed, by the
deed or assignment or a copy thereof duly certified to be correct by
both the assigner and the assigns or their duly authorized agents.

6.  **ASG-BR006-Priority of Assignment notices**

The priority of claim under a policy will be governed by the dates on
which the notices of the assignments have been received by the
Postmaster General/ Head of Division at his office.

7.  **ASG-BR007-Rights of assignor**

After the assignment of a policy is once effected, the policy cannot be
dealt with any further by the assignor.

8.  **ASG-BR008-Re-assignment**

In order to enable the policy-holder to deal with the policy again he
should have a re-assignment in writing in his own favour executed by the
assigns, attested by one or more witnesses, and registered in the
records of the Postmaster General/ Head of Division.

9.  **ASG-BR09- Special cases -Murder of the policy holder by the legal
    heir(s)**

If any of the legal heir(s) or the nominee(s)/trustee of a policy holder
has been charged with the murder of the policy holder, the policy money
shall not be paid to him/her unless he/she is honourably acquitted of by
the competent court of law.

10. **ASG-BR010-Transfer of Assignment**

A transfer or assignment of a policy made in accordance with the
provision of section 38 of the Insurance Act, 1938 shall automatically
cancel a nomination (Section 39 (4)), provided that the assignment of a
policy to the insurer who bears the risk on the policy at the time of
the assignment.

11. **ASG-BR011-Registration of Assignment(s)**

An assignment for valuable consideration should be registered in the
office of the Postmaster General/ Head of Division under this rule.

12. **ASG-BR012-Cases where assignees are more than one**

Assignment of policy as a whole may be made either in favour of one
person or jointly in favour of two or more persons.

13. **ASG-BR013-Cases where Assignment is made in favour of President of
    India**

A policy may be assigned to the President of India for the purpose of
paying estate duty payable under the Estate Duty Act, 1953 (34 of 1953)
in the form prescribed in Rule 31 of the Estate Duty Rules, 1953.

In the case of a policy assigned to the President of India for the
purpose of paying estate duty, the assured shall surrender to the
Controller of Estate Duty all former deeds of assignments or
re-assignments, if any, in respect of the policy.

14. **ASG-BR014 --Cases of Material misrepresentation**

Wrong information furnished by a person or suppression of factual
information by a person admitted to the benefits of the PLI Fund/RPLI
Fund will, at the discretion of the Postmaster General, render voidable
the contract concluded with that person and lead to forfeiture of all
payments made by him.

15. **ASG-BR015 Assignment Modification(Reassignment)**

At a time only one Active assignment exists. System Changes the
Assignment status to Reassigned and on clicking the Add Assignment
button or Screen exit the System pushes the expired assignment to the
past Assignment screen.

All past Assignments are listed on the screen in order with the latest
expired assignment on top.

# References

  ----------------------------------------------------------------------------------------------------
  **Serial   **Document**   **Link**
  No.**                     
  ---------- -------------- --------------------------------------------------------------------------
  1          POLI Rules     [ESankalan](http://www.postallifeinsurance.gov.in/static/Rules2011.aspx)
             2011           

  ----------------------------------------------------------------------------------------------------

# Reports/Interfaces

## Reports

Reports for Assignment to be decided by Business Team

## Letters and Documents:

  ---------------------------------------------------------------------------
  **Document   **Letter/Document      **Brief Description**
  Number**     Name**                 
  ------------ ---------------------- ---------------------------------------
  L001         Assignment Letter      Once the assignment is completed,
                                      assignment letter will be sent to the
                                      customer

  L002         Assignment Rejection   In case the assignment is not approved,
               Letter                 Rejection letter will be sent to the
                                      customer

                                      
  ---------------------------------------------------------------------------
