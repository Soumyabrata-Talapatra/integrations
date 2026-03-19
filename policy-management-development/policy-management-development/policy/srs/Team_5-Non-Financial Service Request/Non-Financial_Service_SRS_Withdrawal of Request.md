> INTERNAL APPROVAL FORM

**Project Name:** Withdrawal of Request

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
[5](#functional-requirements-specification)](#functional-requirements-specification)

[4.1 Non-Financial Service Request Indexing page (At Post Office)
[5](#non-financial-service-request-indexing-page-at-post-office)](#non-financial-service-request-indexing-page-at-post-office)

[**5. Attachments** [6](#attachments)](#attachments)

## **1. Executive Summary**

The purpose of this document is to define the functional and
non-functional requirements for the **Withdrawal of Service Request
(SR)** module in the India Post PLI Insurance Management System (IMS).
This module will enable customers to withdraw service requests
digitally, replacing the current manual process in the McCamish system
with an automated, transparent, and customer-centric workflow.

## **2. Project Scope**

The module applies to all PLI customers who wish to withdraw any service
request before its completion, including:

- Change of address or nomination.

- Policy loan request.

- Policy revival.

- Surrender or maturity claim (pre-processing).

- Policy servicing amendments.

- Channels supported:

  - Online Portal.

  - Mobile App.

  - Post Office counters.

  - Call Centre / Assisted Digital.

## **3. Business Requirements**

  ------------------------------------------------------------------------
  **ID**       **Requirements**
  ------------ -----------------------------------------------------------
  FS_WSR_001   System shall allow customers to initiate withdrawal
               requests via portal, app, Post Office, or call center.

  FS_WSR_002   System shall validate eligibility based on SR status and
               withdrawal window.

  FS_WSR_003   System shall auto-approve eligible withdrawals; route
               exceptions to CPC User.

  FS_WSR_004   System shall update SR status to "Withdrawn by Customer"
               and archive the request.

  FS_WSR_005   System shall notify customers in real-time via SMS, email,
               and app.

  FS_WSR_006   System shall maintain a complete audit trail of withdrawal
               actions.
  ------------------------------------------------------------------------

## **4. Functional Requirements Specification**

**Flow Diagram for Withdrawal of Request Process:**
![](media/image1.png){width="2.983436132983377in"
height="4.475154199475066in"}

## 4.1 Non-Financial Service Request Indexing page (At Post Office)

- **Purpose:** To create Non-Financial Service Request for Withdrawal of
  Request to drop the service request already created in the IMS System.

- **Fields:**

  - Request Type: Service request type like Withdrawal of Request,
    Revival, Surrender, Loan etc.

  - Service Request ID: Service request which needs to be withdrawn.

  - Service Request Date: Date on which service request is indexed.

  - Policy Number: Text

  - Office Code: Facility ID of the office where request is getting
    indexed

  - Service Request Channel: Channel for the service request like RICT,
    CP etc.

  - Username: Emp ID of the user indexing the request

- **Rules:**

  - CPC User should be able to Approve or Reject the request. Upon,
    Approval, the original service request should be reversed
    automatically.

  - All the transactions associated with that service request should
    also be reversed automatically.

  - This transaction will not be applicable if any payment is done by
    customer or if any disbursement happens from the policy.

## **5. Attachments**

The following documents can be referred.
