> INTERNAL APPROVAL FORM

**Project Name: Lead Management and Illustration System for Customer
Portal**

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

[**2. Business Objectives**
[4](#business-objectives)](#business-objectives)

[**3. Project Scope** [4](#project-scope)](#project-scope)

[**4. Business Requirements**
[4](#business-requirements)](#business-requirements)

[**5. Functional Requirements Specification**
[5](#functional-requirements-specification)](#functional-requirements-specification)

[**6. Process Flows/ Use Cases/ UML Diagrams**
[11](#process-flows-use-cases-uml-diagrams)](#process-flows-use-cases-uml-diagrams)

[**7. Appendices** [12](#appendices)](#appendices)

## **1. Executive Summary**

This document outlines the business requirements for Lead Management and
Illustration System modules in the Customer Portal of Insurance
Management System (IMS). It aims to streamline customer acquisition,
improve quote accuracy, and ensure compliance with regulatory standards.

## **2. Business Objectives**

The IMS project addresses the need for efficient lead tracking and
compliant benefit illustrations. This system will streamline lead
tracking, customer engagement, and policy illustration generation for
insurance products.

## **3. Project Scope**

The Lead Management and Illustration System will be a module of the
Insurance Admin System used by agents, sales managers, and
administrators. It will:

- Capture and manage leads from various sources.

- Track lead status and interactions.

- Generate personalized insurance illustrations based on customer
  inputs.

- Integrate with the Insurance Management System (IMS) for policy
  creation.

## **4. Business Requirements**

1.  **Lead Management:**

  -----------------------------------------------------------------------
  FS-LMI-001    The system shall capture leads from multiple channels
                (portal, call-centre, walk-ins, referrals, campaigns).
  ------------- ---------------------------------------------------------
  FS-LMI-002    The system shall allow automatic and manual assignment of
                leads to agents based on geographical location (PIN
                code).

  FS-LMI-003    The system shall allow tracking of lead status and
                follow-up actions.

  FS-LMI-004    The system shall allow reallocation of leads based on
                updated PIN code mapping or agent availability.
  -----------------------------------------------------------------------

2.  **Illustration System:**

  -----------------------------------------------------------------------
  FS-LMI-005    The system shall generate benefit illustrations based on
                configurable product rules and customer inputs.
  ------------- ---------------------------------------------------------
  FS-LMI-006    The system shall support generation of new business
                quotes for insurance products.

  FS-LMI-007    The system shall allow configuration of calculation logic
                for illustrations by admin users.
  -----------------------------------------------------------------------

## **5. Functional Requirements Specification**

1.  **Lead Management:**

    1.  **Lead Capture Page**

- **Fields:**

  - Lead Name (\*mandatory)

  - Contact Number (\*mandatory)

  - Email ID

  - Address

  - PIN Code (\*mandatory)

  - Source (Website, Walk-in, Referral, Call Center, Campaign)

  - Product Interested (\*mandatory)

  - Lead Type (New / Existing)

  - Remarks

- **Business Rules:**

  - Mandatory fields: Name, Contact Number, PIN Code, Product
    Interested.

  - Duplicate check on Contact Number.

**Wireframe for the Lead Capture Page:**

![](media/image1.png){width="2.296517935258093in"
height="3.4447779965004375in"}

1.  **Lead Assignment Page**

- **Fields:** Based on the details given by customer in Lead Capture
  Page, the following details should be auto-populated and saved in
  database.

  - Lead ID

  - Assigned Agent

  - Assignment Date

  - Assignment Mode (Manual / Auto)

- **Business Rules:**

  - Auto-assignment based on PIN code mapping to agent's location.

  - Manual override allowed by Sales Manager.

**Wireframe for the Lead Assignment Page:**

![](media/image2.png){width="3.1751793525809275in"
height="4.7627679352580925in"}

1.  **Lead Tracking Page**

- **Fields:**

  - Lead ID

  - Status (New, Contacted, Follow-up, Converted, Dropped)

  - Last Interaction Date

  - Next Follow-up Date

  - Interaction Notes

**Wireframe for Lead Tracking Page:**

![](media/image3.png){width="2.6950885826771653in"
height="4.042635608048994in"}

1.  **Lead Reallocation Page**

- **Fields:**

  - Lead ID

  - Current Agent

  - New Agent (Dropdown based on PIN code proximity)

  - Reason for Reallocation

- **Business Rules:**

  - Reallocation only by Admin.

**Wireframe for Lead Reallocation Page:**

![](media/image4.png){width="1.7549617235345583in"
height="2.6324431321084862in"}

2.  **Illustration System**

    1.  **Illustration Generation Page**

- **Fields:**

  - Customer Name

  - Age / DOB

  - Gender

  - Sum Assured

  - Policy Term

  - Premium Payment Term

  - Product Name

  - Mobile Number

  - Email Address

**Wireframe for the Illustration Generation Page:**

![](media/image5.png){width="3.392268153980752in"
height="5.088404418197725in"}

1.  **Benefit Illustration Page**

- **Fields:** It should display the following output as per the details
  given in Illustration Generation Page:

  - Projected Maturity Value

  - Death Benefit

  - Bonus (if applicable)

  - Premium Frequency (Annual, Half-Yearly, Monthly)

  - Total Amount at Maturity = Sum Assured + Bonus

- **Business Rules:**

  - Total Amount to be Paid at Maturity must be displayed clearly for
    each product.

  - Illustration must be generated as PDF.

  - Option to proceed to policy creation should be available at the end
    of the illustration (button: "Proceed to Policy Creation").

**Wireframe for the Benefit Illustration Page:**

![](media/image6.png){width="4.268338801399825in"
height="6.402508748906387in"}

1.  **Illustration Configuration Page (Admin)**

- **Fields:**

  - Product Name

  - Calculation Formula (editable)

  - Bonus Rate

  - Mortality Charges

  - Admin Charges

- **Business Rules:**

  - Only Admin can edit.

**Wireframe for Illustration Configuration Page:**

![A screenshot of a phone application AI-generated content may be
incorrect.](media/image7.png){width="4.259183070866142in"
height="6.3887740594925635in"}

## **6. Process Flows/ Use Cases/ UML Diagrams**

1.  UML Diagram for Lead Management Process:

![](media/image8.png){width="5.838813429571304in"
height="8.758219597550307in"}

2.  UML Diagram for Illustration System Process:

![](media/image9.png){width="4.82379593175853in"
height="7.235694444444444in"}

## **7. Appendices**

The Following Documents attached below can be used.
