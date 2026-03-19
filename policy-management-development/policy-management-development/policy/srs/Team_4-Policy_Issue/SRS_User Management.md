> INTERNAL APPROVAL FORM

**Project Name:** User Management

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

[4.1 User Management [5](#user-management)](#user-management)

[4.2 Role Management [6](#role-management)](#role-management)

[4.3 Queue Management [6](#queue-management)](#queue-management)

[4.4 Group Management [6](#group-management)](#group-management)

[4.5 Audit & Logging [6](#audit-logging)](#audit-logging)

[**5. Test Cases** [7](#test-cases)](#test-cases)

[**6. Appendices** [8](#appendices)](#appendices)

## **1. Executive Summary**

This document defines the software and functional requirements for the
User Management Module of the Insurance Management System (IMS) for
India Post PLI. The purpose of this module is to enable administrative
control over users, their roles, access permissions, work queues, and
groups, ensuring proper segregation of duties, operational efficiency,
and compliance with regulatory and internal security policies.

## **2. Project Scope**

The User Management module will:

- Manage creation, modification, deactivation, and reactivation of
  system users.

- Define and assign roles, queues, and groups for controlled access and
  workflow management.

- Ensure secure authentication, authorization, and auditability.

- Integrate with other IMS components (Policy Administration, Claims,
  Premium Collection, etc.) for role-based data access.

## **3. Business Requirements**

  --------------------------------------------------------------------------
  Requirement   Business         Requirements Description
  ID            Requirement      
  ------------- ---------------- -------------------------------------------
  FS_UM_001     Centralized User System shall provide a centralized
                Administration   interface to manage all IMS users across
                                 India Post PLI.

  FS_UM_002     Role-Based       Access to modules, functions, and data
                Access Control   shall be controlled via predefined roles.
                (RBAC)           

  FS_UM_003     Queue Management Work allocation (e.g., policy issuance,
                                 claims, underwriting) shall be handled via
                                 user queues.

  FS_UM_004     Group Management Logical grouping of users (e.g., under a
                                 branch, region, or function) for task
                                 assignment and reporting.

  FS_UM_005     Audit Trail      All user management activities (create,
                                 update, deactivate, role change) shall be
                                 logged with timestamp and actor details.

  FS_UM_006     Authentication & Secure login, password policies, OTP/2FA
                Security         (as per India Post IT policy).

  FS_UM_007     User Lifecycle   Support user creation, activation,
                                 suspension, termination, and reactivation.

  FS_UM_008     Delegation &     Allow temporary assignment of queues or
                Substitution     roles to another user during absence.

  FS_UM_009     Approval         Changes to roles, queues, or group
                Workflow         assignment may require approval by
                                 authorized supervisor.

  FS_UM_010     Integration      Should integrate with centralized India
                                 Post employee directory or HRMS for
                                 validation of user identity.

  FS_UM_011     Reporting        Generate reports (active/inactive users,
                                 roles, login activity, audit logs).
  --------------------------------------------------------------------------

![](media/image1.jpg){width="6.268055555555556in"
height="4.178472222222222in"}

## **4. Functional Requirements Specification**

## 4.1 User Management

**1. Create User**

- Fields: Name, Employee ID, Email, Mobile, Role, Group, Queue, Status
  (Active/Inactive)

- Validations: Unique Employee ID, Valid Email Format

**2. Edit User**

- Editable Fields: Role, Group, Queue, Status

**3. Deactivate User**

- Action: Mark user as inactive without deleting data

**4. View User List**

- Filters: Role, Group, Status

- Actions: Edit, Deactivate, View Details

## 4.2 Role Management

**1. Create Role**

- Fields: Role Name, Description, Permissions (CRUD per module)

**2. Edit Role**

- Editable Fields: Description, Permissions

**3. Assign Role to User**

- Mapping: One user can have multiple roles (if allowed by policy)

## 4.3 Queue Management

Define queues for task allocation (policy issuance, underwriting,
claims, etc.). Assign users or groups to queues. Manage queue ownership
and load distribution.

**1. Create Queue**

- Fields: Queue Name, Description, Associated Tasks, Assigned
  Users/Groups

**2. Assign Queue to User/Group**

- Mapping: One queue can be assigned to multiple users/groups

## 4.4 Group Management

Create and manage user groups (branch, regional, functional). Assign
users and roles to groups. Define reporting structure.

1\. Create Group

- Fields: Group Name, Type (Circle/Division/Branch), Parent Group

2\. Assign Users to Group

- Mapping: One user can belong to one or more groups

## 4.5 Audit & Logging

Every change to user, role, queue, or group shall record:

- Action Type (Create/Edit/Delete/Assign)

- User ID performing the action

- Timestamp

- Old and New values

- IP Address / System Source

## **5. Test Cases**

  -----------------------------------------------------------------------------------------
  **TC ID**   **Requirement   **Test Case**  **Description**   **Expected        **Type**
              ID**                                             Result**          
  ----------- --------------- -------------- ----------------- ----------------- ----------
  TC_UM_001   FS_UM_001       Centralized    Verify            All users across  Positive
                              User View      centralized       branches are      
                                             interface         visible in one    
                                             displays all IMS  interface.        
                                             users.                              

  TC_UM_002   FS_UM_007       Create User    Validate          User is created   Positive
                                             successful        and appears in    
                                             creation of a new user list.        
                                             user with valid                     
                                             details.                            

  TC_UM_003   FS_UM_002       Assign Role    Verify assigning  Role is assigned; Positive
                                             predefined role   user gets correct 
                                             to a user.        permissions.      

  TC_UM_004   FS_UM_003       Queue          Validate adding   User is added to  Positive
                              Assignment     user to a work    queue and tasks   
                                             queue (e.g.,      appear in their   
                                             claims).          dashboard.        

  TC_UM_005   FS_UM_004       Group          Verify grouping   Group is created  Positive
                              Assignment     users under a     and users are     
                                             branch or region. linked correctly. 

  TC_UM_006   FS_UM_005       Audit Trail    Check audit log   Audit log shows   Positive
                              Logging        for user creation correct           
                                             and updates.      timestamp, actor, 
                                                               and action.       

  TC_UM_007   FS_UM_006       Secure Login   Validate login    User logs in      Positive
                                             with correct      successfully with 
                                             credentials and   secure            
                                             OTP/2FA.          authentication.   

  TC_UM_008   FS_UM_006       Password       Verify password   Password accepted Positive
                              Policy         meets complexity  only if it meets  
                                             rules.            policy            
                                                               requirements.     

  TC_UM_009   FS_UM_007       User Lifecycle Test activation,  Status changes    Positive
                                             suspension,       correctly and     
                                             termination, and  reflected in      
                                             reactivation.     system.           

  TC_UM_010   FS_UM_008       Delegation     Validate          Delegation works  Positive
                                             temporary         for specified     
                                             delegation of     period and        
                                             queues to another reverts           
                                             user.             automatically.    

  TC_UM_011   FS_UM_009       Approval       Verify role       Change is pending Positive
                              Workflow       change requires   until approved;   
                                             supervisor        audit logs        
                                             approval.         updated.          

  TC_UM_012   FS_UM_010       HRMS           Validate user     User details      Positive
                              Integration    identity against  match HRMS;       
                                             HRMS directory.   creation allowed. 

  TC_UM_013   FS_UM_011       Reporting      Generate report   Report generated  Positive
                                             of                accurately in     
                                             active/inactive   required format.  
                                             users and login                     
                                             activity.                           

  TC_UM_014   FS_UM_007       Duplicate User Attempt to create System rejects    Negative
                              Creation       user with         with "Duplicate   
                                             existing ID.      User ID" error.   

  TC_UM_015   FS_UM_002       Invalid Role   Assign            System shows      Negative
                              Assignment     non-existent role error "Invalid    
                                             to user.          Role".            

  TC_UM_016   FS_UM_001       Unauthorized   User without      Access denied     Negative
                              Access         admin rights      with proper error 
                                             tries to create   message.          
                                             another user.                       

  TC_UM_017   FS_UM_006       Weak Password  Enter password    System rejects    Negative
                                             not meeting       password with     
                                             complexity rules. policy message.   

  TC_UM_018   FS_UM_006       Invalid OTP    Enter incorrect   Login fails with  Negative
                                             OTP during login. "Invalid OTP"     
                                                               error.            

  TC_UM_019   FS_UM_008       Expired        Try accessing     Access denied;    Negative
                              Delegation     delegated queue   delegation period 
                                             after expiry.     expired.          

  TC_UM_020   FS_UM_009       Missing        Attempt role      Role Change       Negative
                              Approval       change without    should not be     
                                             supervisor        allowed.          
                                             approval.                           
  -----------------------------------------------------------------------------------------

## **6. Appendices**

The Following Documents attached below can be used.
