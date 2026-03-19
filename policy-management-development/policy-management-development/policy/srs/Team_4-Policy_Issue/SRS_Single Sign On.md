> INTERNAL APPROVAL FORM

**Project Name:** Single Sign-On

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

[4.1 Authentication Flow
[5](#authentication-flow)](#authentication-flow)

[4.2 Login page [5](#login-page)](#login-page)

[4.3 Forget/Reset Password
[5](#forgetreset-password)](#forgetreset-password)

[4.4 SSO Admin Console and Session Management
[5](#sso-admin-console-and-session-management)](#sso-admin-console-and-session-management)

[**5. Appendices** [6](#appendices)](#appendices)

## **1. Executive Summary**

The purpose of this SRS document is to define the requirements for the
Single Sign-On (SSO) module within the Insurance Management System (IMS)
of India Post PLI.

This module will enable secure and seamless authentication for users
across various IMS modules (Policy Admin, Claims, HR, Reports, etc.)
using one unified login. It will ensure centralized user identity
management, session control, and audit logging.

## **2. Project Scope**

This module will:

- Provide centralized authentication for all IMS modules.

- Integrate with India Post's existing identity provider (IdP) or Active
  Directory.

- Support role-based access control.

- Ensure secure session management and token-based authentication.

- Log all login/logout activities for audit purposes.

## **3. Business Requirements**

  -----------------------------------------------------------------------
  Requirement ID Requirements
  -------------- --------------------------------------------------------
  FS_SSO_001     Users should log in once and access all IMS modules
                 without re-authentication.

  FS_SSO_002     SSO should integrate with India Post's existing identity
                 provider (e.g., LDAP, AD, or OAuth2).

  FS_SSO_003     Role-based access should be enforced
                 post-authentication.

  FS_SSO_004     Session timeout and auto-logout should be configurable.

  FS_SSO_005     Login/logout activities should be logged for audit and
                 compliance.

  FS_SSO_006     SSO should support multi-factor authentication (MFA) if
                 enabled by the IdP.

  FS_SSO_007     SSO should work across web and mobile platforms.

  FS_SSO_008     Users should be redirected to their respective
                 dashboards based on role and office hierarchy.

  FS_SSO_009     Admins should be able to revoke access centrally.

  FS_SSO_010     System should support token refresh and revocation.
  -----------------------------------------------------------------------

![](media/image1.jpg){width="3.84375in" height="5.765625546806649in"}

## **4. Functional Requirements Specification**

## 4.1 Authentication Flow

  -----------------------------------------------------------------------
  Function          Description
  ----------------- -----------------------------------------------------
  Login             Authenticate user via centralized IdP

  Token Generation  Generate secure access token post-authentication

  Token Validation  Validate token for each module access

  Role Mapping      Map user to roles and permissions

  Session           Manage session lifecycle and timeout
  Management        

  Logout            Invalidate session and token

  Audit Logging     Log login/logout and token activities

  MFA Support       Trigger MFA if configured in IdP
  -----------------------------------------------------------------------

## 4.2 Login page

User authentication entry point for IMS.

**Features:**

- Supports username/password or 2FA

- Error messages for invalid credentials

- Auto redirect to IMS dashboard upon success

## 4.3 Forget/Reset Password

Enables users to recover or reset their password securely.

## 4.4 SSO Admin Console and Session Management

Administrative management of SSO configuration. Track and manage active
user sessions.

**Fields:**

- Active Sessions: List of currently logged-in users

- Force Logout: Button to terminate selected user session

- Login History: Table of login/logouts with timestamps

- Audit Log: Full trail of SSO activities

## **5. Appendices**

The Following Documents attached below can be used.
