> *INTERNAL APPROVAL FORM*

**Project Name:** Duplicate Policy Bond

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
> [4.2 Duplicate Policy Bond Page 6](#duplicate-policy-bond-page)
>
> [**5. Attachments** 6](#attachments)

## **1. Executive Summary**

This document outlines the functional and non-functional requirements
for implementing the Duplicate Policy Bond feature in Insurance
Management System. The goal is to digitize and streamline the issuance
of duplicate policy bonds, ensuring security, traceability, and
compliance.

## **2. Project Scope**

The feature will allow customers to request duplicate policy bonds
through multiple channels, enable secure identity verification, automate
approval workflows, and ensure secure issuance and dispatch of the
duplicate bond.

## **3. Business Requirements**

  ----------------------------------------------------------------------
  **ID**       **Requirements**
  ------------ ---------------------------------------------------------
  FS_DPB_001   Enable customers to request duplicate bonds via portal,
               mobile app, agent portal and app, or branch, capturing
               essential policy and identity details.

  FS_DPB_002   Verify identity using OTP/eKYC and allow digital upload
               and validation of FIR and indemnity documents.

  FS_DPB_003   Generate duplicate bonds with digital watermark and
               signature, maintaining version history for traceability.

  FS_DPB_004   Provide both digital and physical dispatch modes with
               integrated tracking through India Post systems.

  FS_DPB_005   Ensure complete audit trails and retain records for
               compliance standards.
  ----------------------------------------------------------------------

## **4. Functional Requirements Specification**

**Flow Diagram for Duplicate Policy Bond Process:**![A diagram of a
process flow AI-generated content may be
incorrect.](media/image2.png){width="3.168455818022747in"
height="5.177256124234471in"}

## 4.1 Non-Financial Service Request Indexing page

- **Purpose:** To create Non-Financial Service Request for Duplicate
  Policy Bond Page.

- **Fields:**

  - Request Type: Option should include service request type like
    Duplicate Policy Bond, etc.

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

## 4.2 Duplicate Policy Bond Page

- **Purpose:** To send the Duplicate Policy Bond for the Policy.

- **Field:**

  - Request Type: Auto-Populated

  - Ticket No.: Auto-Populated

  - Policy Number (auto-validated)

  - Request Date: Calendar: Editable

  - Reason: Dropdown: Options for issuing duplicate policy bond like
    Destroyed due to Fire, etc.

  - Fee Received: Checkbox

  - List of Documents: Multiple checkboxes for 'ID and Address Proof',
    'Indemnity Bond', 'Copy of FIR', 'Certificate from fire
    authorities', 'Certificate of Unit Commanding officer', 'Copy of
    Notice of loss published in Newspaper', 'Waive Documents'.

  - Add Documents: button: Option to upload the documents.

  - Add Comments: Button: Option to add comments after clicking this
    button

  - Submit: Button

- **Rules:**

  - Clicking on Submit button should generate a downloadable PDF of
    Policy Bond which can be printed and given to Customer.

## **5. Attachments**

The following documents can be referred.
