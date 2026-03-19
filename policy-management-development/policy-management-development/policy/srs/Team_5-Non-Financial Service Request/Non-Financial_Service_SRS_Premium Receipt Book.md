> *INTERNAL APPROVAL FORM*

**Project Name:** Premium Receipt Book

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
> [4.1 Non-Financial Service Request Indexing page
> 5](#non-financial-service-request-indexing-page)
>
> [4.2 Premium Receipt Book Page 6](#premium-receipt-book-page)
>
> [**5. Attachments** 6](#attachments)

## **1. Executive Summary**

This document defines the requirements for the Premium Receipt Book
(PRB) module in IMS. It supports both digital and physical receipt
workflows, ensuring compliance, traceability, and customer convenience.

## **2. Project Scope**

The PRB Module will:

- Generate digital receipts for all premium payments.

- Deliver receipts via customer portal, mobile app, and SMS.

- Allow optional physical PRB issuance in rural/non-digital areas after
  taking due charges.

## **3. Business Requirements**

  ----------------------------------------------------------------------
  **ID**       **Requirements**
  ------------ ---------------------------------------------------------
  FS_PRB_001   System shall auto-generate a digital premium receipt upon
               successful payment.

  FS_PRB_002   System shall send the receipt via SMS and make it
               available on the portal/app.

  FS_PRB_003   Physical PRB shall be issued only if the customer pays
               required fee.

  FS_PRB_004   lease delete this row

  FS_PRB_005   Each receipt shall be unique .

  FS_PRB_006   System shall maintain a searchable digital archive of all
               receipts.

  FS_PRB_007   System shall allow branch/post office staff to view,
               print, or scan PRBs.

  FS_PRB_008   System shall support reconciliation of physical and
               digital receipts.
  ----------------------------------------------------------------------

## **4. Functional Requirements Specification**

**Flow Diagram for Policy Receipt Book Process:**
![](media/image2.png){width="4.038205380577428in"
height="3.888390201224847in"}

## 4.1 Non-Financial Service Request Indexing page

- **Purpose:** To create Non-Financial Service Request for Premium
  Receipt Book on obtaining due fee.

- **Fields:**

  - Request Type: Option should include service request type like
    Premium Receipt Book, etc.

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

## 4.2 Premium Receipt Book Page

- **Purpose:** To download the Premium Receipt for the Policy.

- Navigate to "Premium Collection" → "Manual Receipt Entry" or similar
  module.

- The user should be able to search the premium receipt using following
  details:

  - Policy Number

  - Customer Name

  - Amount Paid

  - Date of Payment

  - PRB Book Number

  - Receipt Serial Number

  - Payment Mode

- The Premium Receipt should be downloaded in PDF Format.

- The Online Premium Receipt can also be sent to Customer using Email,
  SMS and WhatsApp channel also.

## **5. Attachments**

The following documents can be referred.
