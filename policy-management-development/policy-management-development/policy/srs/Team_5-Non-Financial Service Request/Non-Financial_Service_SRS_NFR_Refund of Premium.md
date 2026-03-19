> *INTERNAL APPROVAL FORM*

**Project Name:** Refund of Premium

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
> [4.2 Duplicate Policy Bond Page 6](#_bvhbe5iq7zad)
>
> [**5. Attachments** 6](#attachments)

## **1. Executive Summary**

This document outlines the Software Requirements Specification (SRS) for
the Refund of Premium module in the India Post PLI Insurance Management
System (IMS 2.0). The objective is to digitize and streamline the refund
process, replacing the current manual and paper-based system with an
automated, transparent, and efficient digital workflow.

## **2. Project Scope**

The Refund of Premium module will enable customers to request refunds
for premium payments made erroneously. The system will automate
validation, computation, and disbursement of refunds, integrating with
financial systems such as NEFT, POSB, and IPPB.

## **3. Business Requirements**

  ----------------------------------------------------------------------
  **ID**       **Requirements**
  ------------ ---------------------------------------------------------
  FS_ROP_001   System shall allow customers to submit refund requests
               digitally.

  FS_ROP_002   System shall compute refund amount with applicable
               GST/TDS rules.

  FS_ROP_003   System shall disburse refunds via NEFT, POSB, or IPPB.

  FS_ROP_004   System shall notify customers of refund status at each
               stage.

  FS_ROP_005   System shall maintain a complete audit trail of all
               refund transactions.
  ----------------------------------------------------------------------

**Refund of premium process:**

![](media/image2.png){width="4.441984908136483in"
height="2.961159230096238in"}

## **4. Functional Requirements Specification**

## 4.1 Non-Financial Service Request Indexing page

- **Purpose:** To create Service Request for Refund of Premium.

- **Fields:**

  - Service Request Date: Calendar: Date on which service request is
    indexed.

  - Policy Number: Text

  - customerId: Unique ID of the customer (e.g., CIF or Aadhaar-linked
    ID).

  - refundReason: Reason for requesting the refund (e.g., duplicate
    payment, wrong deposit).

  - paymentDate: Date on which the premium payment was made.

  - paymentAmount: Amount paid by the customer.

  - paymentMode: Mode of payment used (e.g., Online, POSB, IPPB, Cash).

  - receiptNumber: Receipt or transaction ID associated with the
    payment.

  - bankAccountNumber: Customer's bank account number for receiving the
    refund.

  - ifscCode: IFSC code of the customer's bank branch.

  - bankName: Name of the bank where the refund will be credited.

  - accountHolderName: Name of the account holder as per bank records.

  - supportingDocuments: Optional documents such as payment proof or
    cancelled cheque (usually in Base64 or as a file URL).

  - remarks: Any additional comments or notes related to the refund
    request.

  - Office Code: Text: Facility ID of the office where request is
    getting indexed

  - Service Request Channel: Text: Channel for the service request like
    RICT, CP etc.

  - Username: Text: Emp ID of the user indexing the request

- **Rules:**

  - Request should get created and CPC User should be able to open the
    details.

  - Automatic Payment should be processed once the request got approved
    by CPC User in the bank account details provided by customer.

## **5. Attachments**

The following documents can be referred.
