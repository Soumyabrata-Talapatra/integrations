> *INTERNAL APPROVAL FORM*

**Project Name:** Agent Change

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
> 4](#functional-requirements-specification)
>
> [4.1 Non-Financial Service Request Indexing page
> 4](#non-financial-service-request-indexing-page)
>
> [4.2 Agent Change Page 5](#agent-change-page)
>
> [**5. Attachments** 5](#attachments)

## **1. Executive Summary**

The Agent Change module enables policyholders to request a change in
their assigned agent for servicing their Postal Life Insurance (PLI)
policy. This may be required due to suspension on group of fraud,
removal and resignation of the agent, or other valid reasons. The system
must ensure proper validation, approval workflow, and audit tracking for
such changes.

## **2. Project Scope**

This scope will include the following modules:

- Agent Change Process

- Automatically update commission mapping and policy.

- Provide real-time notifications and centralized audit logs.

## **3. Business Requirements**

  ---------------------------------------------------------------------
  **ID**      **Requirements**
  ----------- ---------------------------------------------------------
  FS_AC_001   Allow digital initiation of agent change request from IT
              2.0 System, mobile app and customer portal

  FS_AC_002   Validate new agent's eligibility and jurisdiction using
              Agent Master.

  FS_AC_003   Automatically recalculate commission mapping and update
              policy.

  FS_AC_004   Notify stakeholders (customer, old agent, new agent,
              admin) via SMS/Email.

  FS_AC_005   Maintain centralized audit trail for all agent change
              actions.

  FS_AC_006   Restrict agent selection to eligible agents within
              customer's jurisdiction.

  FS_AC_007   Allow uploading of supporting documents (if required).

  FS_AC_008   Provide acknowledgment receipt to customer upon request
              submission.
  ---------------------------------------------------------------------

## **4. Functional Requirements Specification**

**Flow Diagram for Agent Change
Process:**![](media/image2.png){width="6.268055555555556in"
height="4.178472222222222in"}

## 4.1 Non-Financial Service Request Indexing page

- **Purpose:** To create Non-Financial Service Request for Agent Change.

- **Fields:**

  - Request Type: Option should include service request type like Agent
    Change, etc.

  - Service Request Date: Calendar: Date on which service request is
    indexed.

  - Policy Number: Text

  - Current Agent ID: Option to search and Select the Agent should be
    present.

  - New Agent ID: Option to search and Select the Agent should be
    present.

  - Office Code: Text: Facility ID of the office where request is
    getting indexed

  - Service Request Channel: Text: Channel for the service request like
    RICT, CP etc.

  - Username: Text: Emp ID of the user indexing the request

- **Rules:**

  - Request should get created and CPC User should be able to open the
    details.

## 4.2 Agent Change Page

- **Purpose:** To perform the Agent Change for the Policy.

- **Field:**

  - Request Type: Auto-Populated

  - Ticket No.: Auto-Populated

  - Policy Number (auto-validated)

  - Customer Name (auto-filled)

  - Current Agent Code & Name (auto-filled)

  - Reason for Agent Change (dropdown + remarks)

  - Select New Agent (dropdown filtered by jurisdiction & eligibility)

  - Upload Supporting Document (optional)

  - Submit: Button

- **Rules:**

  - Clicking on Submit button should update the Agent details of the
    Policy.

  - It should Automatically recalculate commission mapping and all
    future commission should go to new agent.

  - Upon Successful completion of Agent Change process, notification
    should be sent to Mobile/Email/Whatsapp of both previous and new
    agent along with the policy owner.

## **5. Attachments**

The following documents can be referred.
