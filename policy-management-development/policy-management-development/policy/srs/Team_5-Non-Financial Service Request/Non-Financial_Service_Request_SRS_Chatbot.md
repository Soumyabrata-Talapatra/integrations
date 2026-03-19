> INTERNAL APPROVAL FORM

**Project Name:** Chatbot

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
[4](#functional-requirements-specification)](#functional-requirements-specification)

[4.1 Use Cases 1: Policy Status Inquiry
[5](#use-cases-1-policy-status-inquiry)](#use-cases-1-policy-status-inquiry)

[4.2 Use Cases 2: Premium Reminder
[5](#use-cases-2-premium-reminder)](#use-cases-2-premium-reminder)

[4.3 Use Cases 3: Premium Payment Links
[6](#use-cases-3-premium-payment-links)](#use-cases-3-premium-payment-links)

[4.4 Use Cases 4: Service Request Tracking
[6](#use-cases-4-service-request-tracking)](#use-cases-4-service-request-tracking)

[4.5 Use Cases 5: Loan/Surrender Value Checks
[6](#use-cases-5-loansurrender-value-checks)](#use-cases-5-loansurrender-value-checks)

[4.6 Use Cases 6: Registration and onboarding assistance
[6](#use-cases-6-registration-and-onboarding-assistance)](#use-cases-6-registration-and-onboarding-assistance)

[4.7 Use Cases 7: Guided claim submission and nomination updates
[7](#use-cases-7-guided-claim-submission-and-nomination-updates)](#use-cases-7-guided-claim-submission-and-nomination-updates)

[4.8 Use Cases 8: FAQ Handling
[7](#use-cases-8-faq-handling)](#use-cases-8-faq-handling)

[4.9 Use Cases 9: Escalation to Live Agent
[7](#use-cases-9-escalation-to-live-agent)](#use-cases-9-escalation-to-live-agent)

[**5. Attachments** [7](#attachments)](#attachments)

## **1. Executive Summary**

The purpose of this document is to define the functional and
non-functional requirements for the Chatbot Module in IMS. The chatbot
will serve as a virtual assistant to help policyholders and stakeholders
interact with the system for queries, service requests, and information
retrieval.

## **2. Project Scope**

This Module will:

- Provide 24x7 automated assistance via web and mobile platforms.

- Handle FAQs, policy status queries, premium due, and service request
  tracking.

- Integrate with IMS Core, CRM, and Notification Engine.

- Support multilingual interactions and escalation to human agents.

## **3. Business Requirements**

  -----------------------------------------------------------------------
  **ID**      **Requirements**
  ----------- -----------------------------------------------------------
  FS_CB_001   The chatbot shall provide instant responses to customer
              queries using a predefined knowledge base and FAQs.

  FS_CB_002   The chatbot shall retrieve and display real-time policy and
              premium information from IMS Core.

  FS_CB_003   The chatbot shall accept and track service requests
              initiated by users.

  FS_CB_004   The chatbot shall escalate unresolved or complex queries to
              live customer service agents.
  -----------------------------------------------------------------------

## **4. Functional Requirements Specification**

![](media/image1.png){width="6.268055555555556in"
height="4.178472222222222in"}

### 4.1 Use Cases 1: Policy Status Inquiry

- **Actor:** Policyholder

- **Trigger:** User asks "What is my policy status?"

- **Flow:**

  - Chatbot authenticates user via OTP.

  - Retrieves policy status from IMS Core.

  - Displays status and next premium due date.

### 4.2 Use Cases 2: Premium Reminder

- **Actor:** Policyholder

- **Trigger:** User asks "When is my next premium due?"

- **Flow:**

  - Chatbot verifies identity.

  - Fetches premium schedule.

  - Sends SMS/Email reminder if requested.

### 4.3 Use Cases 3: Premium Payment Links

- **Actor:** Policyholder

- **Trigger:** User asks "How should I pay the premium?"

- **Flow:**

  - Chatbot displays the premium payment links and disclose all the
    information including BBPS and other digital modes of payments with
    payment link in the chat.

### 4.4 Use Cases 4: Service Request Tracking

- **Actor:** Policyholder

- **Trigger:** User asks "What's the status of my service request?"

- **Flow:**

  - Chatbot asks for request ID.

  - Retrieves status from service module.

  - Displays current status and expected resolution time.

### 4.5 Use Cases 5: Loan/Surrender Value Checks

- **Actor:** Policyholder

- **Trigger:** User asks "What's the status of my service request?"

- **Flow:**

  - Chatbot asks for request ID.

  - Retrieves status from service module.

  - Displays current status and expected resolution time.

### 4.6 Use Cases 6: Registration and onboarding assistance

- **Actor:** Policyholder

- **Trigger:** User asks "How to buy Policy?"

- **Flow:**

  - Chatbot asks for Name, Mobile Number and EMail.

  - Send the data to Lead Management System.

  - Displays the Information and Link regarding Digital Onboarding and
    policy issue.

### 4.7 Use Cases 7: Guided claim submission and nomination updates

- **Actor:** Policyholder

- **Trigger:** User asks "How to process Claim?" or "How to add
  Beneficiary or Nominee"

- **Flow:**

  - Chatbot asks for Policy Number.

  - Displays steps for the Claim submission (Both Digital and Offline at
    Post Office) or for adding Nomination.

### 4.8 Use Cases 8: FAQ Handling

- **Actor:** Policyholder

- **Trigger:** User asks question similar to "How can I update my
  address?"

- **Flow:**

  - Chatbot matches query to FAQ database.

  - Provides step-by-step instructions or link to service form.

### 4.9 Use Cases 9: Escalation to Live Agent

- **Actor:** Policyholder

- **Trigger:** User asks "I want to talk to a person."

- **Flow:**

  - Chatbot checks agent availability.

  - Transfers session to live chat.

  - Logs escalation event.

## **5. Attachments**

The following documents can be referred.
