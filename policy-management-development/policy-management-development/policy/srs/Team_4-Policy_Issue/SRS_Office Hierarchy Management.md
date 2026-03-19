> INTERNAL APPROVAL FORM

**Project Name:** Office Hierarchy Management

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

[4.1 Office Management Dashboard
[5](#office-management-dashboard)](#office-management-dashboard)

[4.2 Add / Edit Office [6](#add-edit-office)](#add-edit-office)

[4.3 Upgradation/Downgradation
[6](#upgradationdowngradation)](#upgradationdowngradation)

[4.4 Office Closure [6](#office-closure)](#office-closure)

[4.5 Audit Trail [7](#audit-trail)](#audit-trail)

[**5. Appendices** [7](#test-cases)](#test-cases)

## **1. Executive Summary**

The purpose of this document is to define the requirements for the
Office Hierarchy Management Module within the Insurance Management
System (IMS) of India Post PLI. This module will enable the management
of the hierarchical structure of post offices, administrative units, and
their associated users for efficient insurance operations.

The module supports the addition, upgradation, downgrading, closure, and
modification of offices and user roles mapped to them.

## **2. Project Scope**

The Office Hierarchy Management module shall allow authorized
administrators to:

- Create and manage offices (HO, DO, BO, CPC, Circle Office, etc.).

- Perform hierarchy operations: addition, upgradation, downgrading,
  merging, closure.

- Manage mapping of users to offices.

- Ensure role-based access and permissions.

- Maintain audit trail and history of structural changes.

- Integrate with other IMS modules (Policy Admin, Claims, Reports)
  through APIs.

## **3. Business Requirements**

  -----------------------------------------------------------------------
  Requirement ID Requirements
  -------------- --------------------------------------------------------
  FS_OH_001      The system shall maintain a centralized hierarchical
                 master of all India Post PLI offices.

  FS_OH_002      The system shall allow addition of new offices (HO, DO,
                 BO, CPC, Circle).

  FS_OH_003      The system shall support upgradation and downgrading of
                 offices based on operational changes.

  FS_OH_004      The system shall allow closure of an office after
                 necessary approval.

  FS_OH_005      The system shall maintain parent-child relationships
                 between offices.

  FS_OH_006      The system shall capture effective dates for all
                 structural changes.

  FS_OH_007      The system shall maintain complete audit trails of all
                 changes with user IDs and timestamps.

  FS_OH_008      The system shall restrict unauthorized access using
                 role-based control.

  FS_OH_009      The system shall integrate with the User Management
                 module to remap users automatically after office
                 structural changes.

  FS_OH_010      The system shall generate reports of hierarchy status
                 and historical changes.

  FS_OH_011      The system shall validate office codes and prevent
                 duplication.

  FS_OH_012      The system shall allow searching, filtering, and
                 exporting of office hierarchy data.
  -----------------------------------------------------------------------

**Sample Office Hierarchy Tree:**

Circle Office (Delhi)

├── Head Office (Delhi HO)

│ ├── Divisional Office (South Delhi DO)

│ │ ├── Branch Office (Lajpat Nagar BO)

│ │ └── Branch Office (Malviya Nagar BO)

│ └── CPC (Delhi CPC)

![](media/image1.jpg){width="6.268055555555556in"
height="4.178472222222222in"}

## **4. Functional Requirements Specification**

## 4.1 Office Management Dashboard

Central console for viewing and managing the office hierarchy.

**Fields:**

- Office Name

- Office Code

- Office Type (Circle, Region, Division, Branch, Sub-Branch)

- Parent Office

- Status (Active, Closed)

- Date of Establishment

- Date of Closure (if applicable)

- Reason for Closure/Upgrade/Downgrade

- Number of Users

- Number of Policies

**Actions:**

- Add Office

- Edit Office

- Upgrade/Downgrade

- Close Office

- View Hierarchy

- Bulk Upload

- Export to Excel/PDF

## 4.2 Add / Edit Office

Create or modify an office record.

**Fields:**

- Office Name (Text, Required)

- Office Code (Alphanumeric, Unique, Required)

- Office Type (Dropdown: Circle, Region, Division, Branch, Sub-Branch)

- Parent Office (Dropdown based on hierarchy)

- Address (Text Area)

- Contact Number (Numeric)

- Email (Email Format)

- Date of Establishment (Date Picker)

**Actions:**

- Save

- Cancel

- Submit for Approval

## 4.3 Upgradation/Downgradation

Change the level/type of an existing office.

**Fields:**

- Current Office Type (Read-only)

- New Office Type (Dropdown)

- Reason (Text Area)

- Effective Date (Date Picker)

## 4.4 Office Closure

Mark an office as closed after approval.

**Fields:**

- Office Name (Read-only)

- Closure Date (Date Picker)

- Reason (Text Area)

- Reassign Users To (Dropdown of active offices)

- Reassign Policies To (Dropdown of active offices)

**Validations:**

- Mandatory reassignment of users and policies

- Closure date must be \>= current date

## 4.5 Audit Trail

Track all changes made in office hierarchy.

**Fields:**

Change ID: Auto-generated

Office Code: Affected office

Change Type: Addition/Upgrade/Closure etc.

Changed By: User ID

Changed Date: Timestamp

Old Value: text

New Value: text

## **5. Test Cases**

The Following Documents attached below can be used.

  -----------------------------------------------------------------------------
  **TC ID**   **Short Name** **Description**          **Expected Result**
  ----------- -------------- ------------------------ -------------------------
  TC_OH_001   Centralized    Verify that the system   All offices (HO, DO, BO,
              Hierarchy      maintains a centralized  CPC, Circle) are visible
              Master         hierarchical master of   in a single hierarchy
                             all offices.             view.

  TC_OH_002   Add New Office Validate addition of a   New office is
                             new office type (HO, DO, successfully added with
                             BO, CPC, Circle).        correct details and
                                                      appears in hierarchy.

  TC_OH_003   Upgrade Office Test upgradation of an   Office type changes
                             office (e.g., BO → DO).  successfully; hierarchy
                                                      updates accordingly.

  TC_OH_004   Downgrade      Test downgrading of an   Office type changes
              Office         office (e.g., DO → BO).  successfully; hierarchy
                                                      updates accordingly.

  TC_OH_005   Close Office   Validate closure of an   Office status changes to
                             office after approval    "Closed"; hierarchy
                             workflow.                reflects closure.

  TC_OH_006   Parent-Child   Verify parent-child      Correct parent-child
              Relationship   linkage between offices. relationships maintained
                                                      after changes.

  TC_OH_007   Effective      Validate capturing of    Effective date is stored
              Dates          effective dates for      and displayed in
                             structural changes.      audit/history.

  TC_OH_008   Audit Trail    Check audit trail for    Audit log shows correct
                             all changes with user ID user, timestamp, and
                             and timestamp.           change details.

  TC_OH_009   Role-Based     Validate that            Unauthorized users
              Access         unauthorized users       receive "Access Denied"
                             cannot modify hierarchy. message.

  TC_OH_010   User Remapping Verify integration with  Users are reassigned to
                             User Management for      correct office
                             automatic user remapping automatically.
                             after office changes.    

  TC_OH_011   Hierarchy      Validate generation of   Reports are generated
              Reports        hierarchy status and     accurately with current
                             historical change        and past data.
                             reports.                 

  TC_OH_012   Office Code    Check validation of      Duplicate office codes
              Validation     office codes and         are rejected with error
                             prevention of            message.
                             duplicates.              

  TC_OH_013   Search &       Validate searching and   Search and filter return
              Filter         filtering of office      correct results based on
                             hierarchy data.          criteria.

  TC_OH_014   Export Data    Verify export            Data exports successfully
                             functionality for        in selected format.
                             hierarchy data           
                             (CSV/Excel).             
  -----------------------------------------------------------------------------

## **6. Appendices**

The Following Documents attached below can be used.
