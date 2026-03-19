> INTERNAL APPROVAL FORM

**Project Name:** Letters & Reports- Agent

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

[Existing Letters: [4](#existing-letters)](#existing-letters)

[Existing Reports: [5](#existing-reports)](#existing-reports)

[**4. Functional Requirements Specification**
[6](#functional-requirements-specification)](#functional-requirements-specification)

[4.1 Create Reports Page
[6](#create-reports-page)](#create-reports-page)

[4.2 Letters Page [6](#letters-page)](#letters-page)

[4.3 Historical Reports Page
[7](#historical-reports-page)](#historical-reports-page)

[**5. Report Samples** [7](#report-samples)](#report-samples)

## **1. Executive Summary**

This module is designed to automate the generation of letters and
reports related to agents in the Postal Life Insurance system. It
supports both event-triggered letters and manually or periodically
generated reports, with standardized formatting and archival
capabilities.

## **2. Project Scope**

This module will help in understanding the requirements for letters and
reports that need to be generated for the agents.

## **3. Business Requirements**

  -----------------------------------------------------------------------
  Requirement ID Requirement
  -------------- --------------------------------------------------------
  FS_AL_001      The system must generate specific letters based on agent
                 lifecycle events (e.g., recruitment, appointment,
                 license updates, termination).

  FS_AL_002      All letters must follow a consistent format including
                 header, footer, font, salutation, and date/amount
                 formatting.

  FS_AL_003      Reports should be available in both PDF and Excel
                 formats and support adhoc generation.

  FS_AL_004      Users should be able to filter and generate reports
                 based on advisor type, office location, profile type,
                 and date range.

  FS_AL_005      Historical letters and reports must be stored and
                 retrievable by users.
  -----------------------------------------------------------------------

### Existing Letters:

List of Existing Letters:

+------+--------------+-----------------------------------------------+
| \#   | Letter Name  | Trigger Condition                             |
+======+==============+===============================================+
| 1.   | Agent        | This letter is sent to an Agent just after    |
|      | Welcome      | the recruitment, informing the Agent of       |
|      | Letter       | various particulars. For example, Agent       |
|      |              | Number, Effective Date, Agent Portal ID and   |
|      |              | password and Sales Support team Contact       |
|      |              | Number.                                       |
+------+--------------+-----------------------------------------------+
| 2\.  | Agent        | This letter is sent to an Agent via the       |
|      | Appointment  | system when the Agent is appointed. This      |
|      | Letter       | letter also informs the Agent of the          |
|      |              | following details:                            |
|      |              |                                               |
|      |              | - Agent Number                                |
|      |              |                                               |
|      |              | - Date of Birth                               |
|      |              |                                               |
|      |              | - PAN Number details                          |
|      |              |                                               |
|      |              | - Managing Agent Name                         |
+------+--------------+-----------------------------------------------+
| 3\.  | License      | This letter is generated and sent to an Agent |
|      | Allotment    | after license details are updated on the      |
|      | Letter       | Agent's profile. This letter will include     |
|      |              | details such as License Number, License Issue |
|      |              | Date and License Renewal Date of the Agent.   |
+------+--------------+-----------------------------------------------+
| 4\.  | License      | This letter is generated when the license of  |
|      | Suspension   | an Agent has not been renewed and the license |
|      | Letter       | has been moved to the suspended status.       |
+------+--------------+-----------------------------------------------+
| 5\.  | Rejection    | This letter is generated when the application |
|      | Letter       | for a candidate to be an Agent is rejected.   |
+------+--------------+-----------------------------------------------+
| 6\.  | Documents    | This letter is generated when after profile   |
|      | Pending      | creation, the status is moved to the Pending  |
|      | Letter       | with the reason as Documents awaited or       |
|      |              | Discrepancy found.                            |
+------+--------------+-----------------------------------------------+
| 7\.  | Agent        | This letter is generated whenever an Agent    |
|      | Termination  | profile is terminated.                        |
|      | Letter       |                                               |
+------+--------------+-----------------------------------------------+

Elements Common across Letters

  -----------------------------------------------------------------------
  \#      Type of           Information
          Information       
  ------- ----------------- ---------------------------------------------
  1       Header            This section displays the logo and the return
                            address.

  2       Footer            This section displays the DoP Postal Life
                            Insurance.

  3       Date Format       In top left corner, the date lines up with
                            the company name under the logo and address.
                            Format of date is \<Month\>\<Date\>,\<Year\>.
                            For example, July 24, 2013. The same date
                            format is used in the body of letter.

  4       Amount Format     Amount starts with the Rs. symbol and a comma
                            is added after thousand. The amount is always
                            up to two decimal places. For example,
                            Rs.2,54,000.00.

  5       Closing Paragraph The closing paragraph is added in each
                            letter. This paragraph depends on the type of
                            letter.

  6       Font              All letters are in Arial font style with font
                            size 10.

  7       Heading           This displays the Letter Title.

  8       Salutation        This includes name of the Agent and
                            salutation. For example, salutation implies
                            Dear Mr./Ms./Mrs.
  -----------------------------------------------------------------------

### Existing Reports:

Report Types:

1.  Periodic Reports: These reports are automatically generated daily,
    weekly, monthly, half-yearly or yearly.

2.  Adhoc Reports: These reports are generated manually as and when
    required.

List of Existing Reports

  -----------------------------------------------------------------------
  \#      Report Category  Report Title      Frequency
  ------- ---------------- ----------------- ----------------------------
  1       Commission       Disbursement      Adhoc
                           Report            

  2       License          Agents Licensed   Adhoc
                           Report            

  3       License          License Renewal   Adhoc
                           Follow-up Report  

  4       Producer         Agent Termination Adhoc
                           Report            

  5       Producer         Performance       Adhoc
                           Report            

  6       Commission       Incentive Report  Adhoc

  7       Commission       Premium Report    Adhoc

  8       Commission       Commission        Adhoc
                           Statement         

  9       Producer         Profitability     Adhoc
                           Report            
  -----------------------------------------------------------------------

Elements Common across Reports

+------+-----------------+--------------------------------------------+
| \#   | Type of         | Information                                |
|      | Information     |                                            |
+======+=================+============================================+
| 1    | Company Name    | This section displays the text Postal Life |
|      |                 | Insurance and logo.                        |
+------+-----------------+--------------------------------------------+
| 2    | Header          | This section contains the following        |
|      |                 | information:                               |
|      |                 |                                            |
|      |                 | - Report Name                              |
|      |                 |                                            |
|      |                 | - Process Date: - \<report start date\> -  |
|      |                 |   \<report end date\>                      |
+------+-----------------+--------------------------------------------+
| 3    | Footer          | This section contains the following        |
|      |                 | elements:                                  |
|      |                 |                                            |
|      |                 | - \< Current Date/time\> left justified    |
|      |                 |   along the bottom                         |
|      |                 |                                            |
|      |                 | - Page \<page number\> of \< number of     |
|      |                 |   pages\> right justified along the bottom |
+------+-----------------+--------------------------------------------+

## **4. Functional Requirements Specification**

### 4.1 Create Reports Page

> Clicking 'Create Reports' button should open this page. This page
> allows users to generate reports using filters.

- **Fields:**

  - Advisor Type: Multi-select checkboxes (Departmental Employee
    Advisor, Field Officer Advisor, Advisor Coordinator, Advisor).

  - Report Type: Dropdown (PDF, Excel).

  - Select Advisor: Button to search and select specific advisor.

  - Statement Start Date / End Date: Calendar fields.

  - Office Type / Location: Dropdowns.

  - Profile Type: Multi-select checkboxes (CPMG, Departmental Employee,
    Development Officer, Direct Agent, etc.).

  - Generate Report: Button to initiate report generation.

### 4.2 Letters Page

> This page stores and displays all letters generated by the agent.

- **Fields:**

  - Search Bar: Text input to search by Agent Name or Agent ID

  - Letter Type Filter: Dropdown (e.g., Welcome, Appointment,
    Termination)

  - Date Range: Two calendar fields (Start Date, End Date)

  - Agent ID: Display column

  - Agent Name: Display column

  - Letter Type: Display column

  - Generated Date: Display column

  - View Letter: Button/icon to open/download the letter

  - Download Format: Option to download as PDF

  - Pagination Controls: For navigating through records

### 4.3 Historical Reports Page

> This page stores and displays all reports generated by the agent. It
> should be limited to last 10 generated Reports by the respective user.

- **Fields:**

  - Search Bar: Text input to search by Report Title

  - Report Category Filter: Dropdown (Commission, License, Producer)

  - Date Range: Two calendar fields (Start Date, End Date)

  - Report Title: Display column

  - Report Category: Display column

  - Generated Date: Display column

  - Format: Display column (PDF/Excel)

  - View Report: Button/icon to open/download the report

  - Pagination Controls: For navigating through records

## **5. Report Samples**
