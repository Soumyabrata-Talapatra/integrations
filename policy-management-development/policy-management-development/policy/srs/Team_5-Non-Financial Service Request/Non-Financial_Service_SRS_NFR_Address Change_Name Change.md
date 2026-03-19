> *INTERNAL APPROVAL FORM*

**Project Name:** Address Change & Name Change

**Version: 1.0**

**Submitted on:**

  ----------------------------------------------------------------------
               **Name**                               **Date**
  ------------ -------------------------------------- ------------------
  **Approved                                          
  By:**                                               

  **Reviewed                                          
  By:**                                               

  **Prepared                                          
  By: **                                              
  ----------------------------------------------------------------------

> *VERSION CONTROL LOG*

  ------------------------------------------------------------------------------
  **Version**   **Date**   **Prepared     **Remarks**
                           By**           
  ------------- ---------- -------------- --------------------------------------
  **1**                                   

                                          

                                          

                                          

                                          
  ------------------------------------------------------------------------------

Table of Contents

> [**1. Executive Summary** 4](#executive-summary)
>
> [**2. Project Scope** 4](#project-scope)
>
> [**3. Business Requirements** 4](#business-requirements)
>
> [**4. Functional Requirements Specification**
> 5](#functional-requirements-specification)
>
> [4.1 Online Name/Address Change Page
> 5](#online-nameaddress-change-page)
>
> [4.2 Non-Financial Service Request Indexing page
> 5](#non-financial-service-request-indexing-page)
>
> [4.3 Name Change Page 6](#name-change-page)
>
> [4.4 Address Change Page 7](#address-change-page)
>
> [**5. Attachments** 9](#attachments)

## **1. Executive Summary**

The purpose of this document is to define requirements for performing
the Address Change and Name Change Transactions in Insurance Management
System (IMS).

## **2. Project Scope**

This scope will include the following modules:

- Address Change Process

- Name Change Process

## **3. Business Requirements**

  ----------------------------------------------------------------------
  **ID**       **Requirements**
  ------------ ---------------------------------------------------------
  FS_ANC_001   Policyholder should be able to request an address change
               online or at a post office.

  FS_ANC_002   System should validate the policy number and customer
               identity before allowing changes.

  FS_ANC_003   System should allow uploading of supporting documents
               (e.g., Aadhaar, Gazette Notification).

  FS_ANC_004   Address Change performed using Aadhar Authentication
               should get completed immediately without any
               interventions.

  FS_ANC_005   Admin users should be able to view pending requests and
               take action.

  FS_ANC_006   Address change should update communication address used
               for correspondence.

  FS_ANC_007   Policyholder should be able to request a name change with
               supporting documents.

  FS_ANC_008   Name Change performed using Aadhar Authentication should
               get completed immediately without any interventions.

  FS_ANC_009   Name change should reflect across all linked policies and
               documents.

  FS_ANC_010   Changes should be reflected in reports and dashboards.

  FS_ANC_011   All changes should be logged with timestamps and user IDs
               for audit trail.

  FS_ANC_012   System should generate acknowledgment receipt for
               customer.

  FS_ANC_013   System should send SMS/email notifications on status
               updates.
  ----------------------------------------------------------------------

## **4. Functional Requirements Specification**

**Flow Diagram for Address Change and Name Change
Process:**![](media/image3.png){width="6.83542760279965in"
height="3.0002395013123357in"}

## 4.1 Online Name/Address Change Page

- To change the Name or Address online using Adhar based integration.

- **Fields & Components:**

  - Aadhar Number (Text field, mandatory)

  - Policy Number: Textbox: Mandatory

  - Change Type: Dropdown: Options are 'Name Change' and 'Address
    Change'.

  - Submit (button)

- **Business Rules:**

  - Clicking on Submit button should fetch the Name or Address details
    as per option selected in 'Change Type' dropdown and should perform
    the Name or Address Change immediately.

  - Customer should immediately be informed via Email/SMS.

  - The Change should get reflected in Policies, Reports and at all
    other places in the system.

  - Audit Trail should be maintained for the Name Change and Address
    Change Transaction with User ID and Timestamp Details.

## 4.2 Non-Financial Service Request Indexing page

- **Purpose:** To create Non-Financial Service Request for Address
  Change and Name Change.

- **Fields:**

  - Request Type: Option should include service request type like
    Address Change, Name Change, etc.

  - Service Request Date: Calendar: Date on which service request is
    indexed.

  - Policy Number: Text

  - Office Code: Text: Facility ID of the office where request is
    getting indexed

  - Service Request Channel: Text: Channel for the service request like
    RICT, CP etc.

  - Username: Text: Emp ID of the user indexing the request

- **Rules:**

  - Request should get created and CPC User should be able to open the
    details.

## 4.3 Name Change Page

- **Purpose:** To perform the Name Change for the User.

- **Field:**

  - Request Type: Auto-Populated

  - Ticket No.: Auto-Populated

  - Salutation: Text: To Update the Salutation Details

  - First Name:

  - Middle Name:

  - Last Name:

  - List of Documents: Checkboxes: Checkboxes present should be 'News
    Paper Notification', 'Name Change Application Form' and 'Gazette
    Notification'

  - Upload Documents: Button: Option should be present to upload the
    documents for the user.

  - Add Comments: Button: Option to Add comments in the request.

  - Request Missing Document Name: Dropdown: Options are 'News Paper
    Notification', 'Name Change Application Form' and 'Gazette
    Notification'

  - Request Missing Document Date: Calendar

  - Request Missing Document Status: Dropdown

  - Request Missing Document Received Date: Calendar

  - Request Missing Document Delete: Icon: Option to delete the table
    row for Missing Document.

  - Request Missing Document Add: link: Option to Add the table row for
    Missing Document.

  - Submit: Button

- **Rules:**

  - Clicking on Submit button should update the Name details of the
    user.

  - If Request Missing Document is filled, the request should not be
    submitted and the notification should be sent to customer for
    submitting the missing documents.

![A screenshot of a computer AI-generated content may be
incorrect.](media/image5.png){width="6.268055555555556in"
height="2.5659722222222223in"}

## 4.4 Address Change Page

- **Purpose:** To perform the Address Change for the User.

- **Field:**

  - Request Type: Auto-Populated

  - Ticket No.: Auto-Populated

  - Address Update For: Dropdown: Options should include Insured,
    Proposer, Assignee, Trustee

  - Address Type: Dropdown: options include official, communication,
    permanent.

  - Address Line1:

  - Address Line2:

  - Village: Text

  - Taluka: Text

  - City: Text

  - District: Text

  - State: Dropdown

  - Pincode: Text

  - Phone Type: Dropdown: options include official, communication,
    permanent.

  - Area Code: Text

  - Landline Number: Text

  - Mobile Number: Text

  - Aadhar Number: Text

  - PAN Number: Text

  - Passport Number: Text

  - Voter ID Number: Text

  - Date of Birth: Calendar

  - List of Documents: Checkboxes: Checkboxes present should be 'Address
    Chang Application Form', 'Rental Agreement' and 'Address Proof'

  - Upload Documents: Button: Option should be present to upload the
    documents for the user.

  - Add Comments: Button: Option to Add comments in the request.

  - Request Missing Document Name: Dropdown: Options should be 'Address
    Chang Application Form', 'Rental Agreement' and 'Address Proof'

  - Request Missing Document Date: Calendar

  - Request Missing Document Status: Dropdown

  - Request Missing Document Received Date: Calendar

  - Request Missing Document Delete: Icon: Option to delete the table
    row for Missing Document.

  - Request Missing Document Add: link: Option to Add the table row for
    Missing Document.

  - Submit: Button

- **Rules:**

  - Clicking on Submit button should update the Address details of the
    user.

  - If Request Missing Document is filled, the request should not be
    submitted and the notification should be sent to customer for
    submitting the missing documents.

![](media/image4.png){width="6.268055555555556in"
height="4.847916666666666in"}

## **5. Attachments**

The following documents can be referred:
