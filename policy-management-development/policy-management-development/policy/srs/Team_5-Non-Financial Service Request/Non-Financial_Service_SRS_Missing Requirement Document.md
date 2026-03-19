> *INTERNAL APPROVAL FORM*

**Project Name:** Missing Requirement Document

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
> [4.1 Missing Documents Request Page (CPC User Only)
> 5](#missing-documents-request-page-cpc-user-only)
>
> [4.2 Non-Financial Service Request Indexing page (At Post Office)
> 6](#non-financial-service-request-indexing-page-at-post-office)
>
> [4.2 Missing Document Upload Page (Link sent to Customer)
> 7](#missing-document-upload-page-link-sent-to-customer)
>
> [**5. Attachments** 7](#attachments)

## **1. Executive Summary**

This document outlines the software requirements for implementing a
Missing Requirement Document Management feature for CPC users in India
Post PLI. The system will enable CPC users to identify missing documents
in customer applications and facilitate communication with customers to
either upload the documents online or submit them at the nearest post
office.

## **2. Project Scope**

This Module will:

- Allow CPC users to mark missing documents.

- Generate communication (email/SMS/letter) to customers.

- Provide a secure upload link or instructions for physical submission.

- Track document submission status.

## **3. Business Requirements**

  ----------------------------------------------------------------------
  **ID**       **Requirements**
  ------------ ---------------------------------------------------------
  FS_MRD_001   CPC users must be able to raise a Missing Requirement
               request directly from the Insurance Management System,
               selecting specific documents required from the customer.

  FS_MRD_002   Customers should be able to submit missing documents via
               secure digital platforms mobile application, customer
               portal, including self-service (portal upload) and
               assisted service (at post office counters).

  FS_MRD_003   The system must send real-time alerts to customers via
               SMS, email, and portal notifications, detailing the
               missing documents and submission options.

  FS_MRD_004   The system must provide dashboards for CPC users to
               monitor document submission status

  FS_MRD_005   Uploaded documents must be automatically tagged to the
               relevant SR, ensuring data integrity and reducing manual
               intervention.
  ----------------------------------------------------------------------

## **4. Functional Requirements Specification**

**Flow Diagram for Missing Requirement Document Process:** ![A diagram
of a post office AI-generated content may be
incorrect.](media/image2.png){width="6.268055555555556in"
height="4.178472222222222in"}

## 4.1 Missing Documents Request Page (CPC User Only)

- **Purpose:** CPC User should be able to raise the request for missing
  documents to be collected from customer. Raising of request by CPC
  should initiate a link to be sent digitally to customer in their
  Email, SMS, WhatsApp, etc. in which user can upload the documents, or
  they can submit the document physically at Post Office.

- **Fields:**

  - Missing Document Request ID -- Auto-generated, read-only unique
    identifier.

  - Policy Number -- Text input, mandatory, validated against active
    policies.

  - Customer Name -- Auto-populated from policy details, read-only.

  - Customer Mobile Number -- Auto-populated, editable if needed.

  - Customer Email ID -- Auto-populated, editable if needed.

  - Document Type -- Dropdown, multi-select (e.g., Identity Proof,
    Address Proof).

  - Document Description / Remarks -- Text area for additional notes.

  - Preferred Submission Mode -- Radio buttons (Upload Online / Submit
    at Post Office).

  - Secure Link Expiry Date -- Date picker, default 7 days from request
    creation.

  - Notification Channels -- Checkboxes (SMS, Email, WhatsApp), default
    all selected.

  - Message Preview -- Read-only, shows notification text to be sent.

  - Attach to Service Request (SR) -- Auto-tagging enabled upon
    submission.

  - Dashboard Visibility -- Status tracking (Pending / Completed /
    Expired).

  - Buttons -- Submit Request, Cancel, Preview Notification.

  - Previous Request Table with Status: A table should display with
    previously raised requests along with the status of documents
    submitted.

- **Rules:**

  - Successful initiation of request should initiate link to be sent
    digitally to customer in their Email, SMS, WhatsApp, etc. in which
    user can upload the documents, or they can submit the document
    physically at Post Office or through customer portal, mobile
    application and with help of agent through agent portal.

## 4.2 Non-Financial Service Request Indexing page (At Post Office)

- **Purpose:** To create Non-Financial Service Request for Missing
  Requirements Documents.

- **Fields:**

  - Missing Document Request ID: Textbox: Input the Missing Document
    Request ID against which customer is submitting the documents.

  - Request Type: Option should include service request type like
    Missing Requirement Document, etc.

  - Document Name: Textbox

  - Service Request Date: Calendar: Date on which service request is
    indexed.

  - Policy Number: Text

  - Office Code: Text: Facility ID of the office where request is
    getting indexed

  - Service Request Channel: Text: Channel for the service request like
    RICT, CP etc.

  - Username: Text: Emp ID of the user indexing the request

- **Rules:**

  - Upon Successful submission, the Missing Documents Request created by
    CPC user should get updated along with the scanned copy of submitted
    document.

  - CPC User should be able to Approve or Reject the Document. If the
    Documents are rejected, link for uploading document will be shared
    again to the customer.

## 4.2 Missing Document Upload Page (Link sent to Customer)

- **Purpose:** To Upload the Missing Document from the auto link sent to
  customer when the Missing Requirement Document request raised by CPC.

- **Field:**

  - Policy Number -- Displayed, read-only (auto-populated from link).

  - Customer Name -- Displayed, read-only (auto-populated).

  - Service Request ID -- Displayed, read-only (auto-tagged to SR).

  - List of Required Documents -- Dynamic list showing document types
    requested (e.g., Identity Proof, Address Proof).

  - Upload Document -- File upload control for each required document.

    - Features:

      - Accept formats: PDF, JPEG, PNG

      - Max size: e.g., 5 MB per file

      - Drag-and-drop or browse option

  - Remarks (Optional) -- Text area for customer notes.

  - Submit Button -- To upload and confirm submission.

  - Cancel Button -- To exit without uploading.

  - Progress Indicator -- Shows upload status (e.g., percentage
    completed).

  - Help/Instructions Section -- Guidelines for acceptable file formats
    and size.

  - Confirmation Message -- Displayed after successful upload with
    reference number.

## **5. Attachments**

The following documents can be referred.
