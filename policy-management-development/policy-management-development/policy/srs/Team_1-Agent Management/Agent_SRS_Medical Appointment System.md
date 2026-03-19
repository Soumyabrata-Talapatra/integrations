> INTERNAL APPROVAL FORM

**Project Name:** Medical Appointment System

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

[4.1 Doctor Onboarding Page
[5](#doctor-onboarding-page)](#doctor-onboarding-page)

[4.2 Doctor Login Page [7](#doctor-login-page)](#doctor-login-page)

[4.3 Doctor Dashboard Page
[7](#doctor-dashboard-page)](#doctor-dashboard-page)

[4.4 Customer Profile Page
[8](#customer-profile-page)](#customer-profile-page)

[4.5 Appointment Scheduling Page
[10](#appointment-scheduling-page)](#appointment-scheduling-page)

[4.6 Medical Report Page
[11](#medical-report-page)](#medical-report-page)

[4.7 Policy Approval Page
[12](#policy-approval-page)](#policy-approval-page)

[**5. Workflow** [13](#workflow)](#workflow)

## **1. Executive Summary**

The purpose of this document is to define requirements to enable India
Post to schedule and manage medical reviews for customers applying for
new PLI and RPLI policies. The system ensures:

- Doctors review applicants' health status.

- Medical reports are captured and linked to policy proposals.

- Insurance admins approve/reject policy issuance based on medical
  review.

## **2. Project Scope**

This scope will include the following modules:

- Doctor Onboarding (panel doctors for medical review)

- Customer Profile Access (linked to policy proposal)

- Appointment Scheduling for Medical Review

- Medical Report Upload & Validation

- Policy Approval Workflow (based on medical outcome)

- Billing & Insurance Processing (if medical tests are chargeable)

## **3. Business Requirements**

  -----------------------------------------------------------------------
  **ID**      **Requirements**
  ----------- -----------------------------------------------------------
  FS_MA_001   The system must allow onboarding of doctors authorized for
              medical reviews.

  FS_MA_002   Each doctor must have verified credentials (license number,
              specialization, location).

  FS_MA_003   Only approved doctors should be available for appointment
              scheduling.

  FS_MA_004   The system must display customer details linked to policy
              proposals (Name, Age, Proposal ID).

  FS_MA_005   Insurance admins must have access to customer health
              history (if available).

  FS_MA_006   Customers cannot edit insurance or proposal details; only
              contact info or non-health related info can be updated.

  FS_MA_007   Insurance admin must be able to schedule medical review
              appointments for customers.

  FS_MA_008   Appointment must be linked to a valid policy proposal.

  FS_MA_009   Notifications must be sent to both customer and doctor upon
              scheduling.

  FS_MA_010   Doctors may upload medical reports after the review.

  FS_MA_011   Doctors Reports must include mandatory health parameters
              (BP, Sugar, ECG, etc.).

  FS_MA_012   Doctors Reports must be digitally signed for authenticity.

  FS_MA_013   Doctors Reports must be stored securely and linked to the
              proposal ID.

  FS_MA_014   Insurance admin must review medical reports before
              approving/rejecting policy issuance.

  FS_MA_015   Rejection must include a reason and comments.

  FS_MA_016   The system must update proposal status automatically after
              approval/rejection.

  FS_MA_017   All medical data must be encrypted and comply with HIPAA or
              health data regulations of India.

  FS_MA_018   Access control must ensure only authorized users can view
              sensitive data.

  FS_MA_019   Audit logs must be maintained for all actions (appointment,
              report upload, approval).

  FS_MA_020   The System must allow deleting of Health Information if
              requested by customer.
  -----------------------------------------------------------------------

## **4. Functional Requirements Specification**

### 4.1 Doctor Onboarding Page

- To register and manage doctors who will conduct medical reviews for
  insurance policy applicants.

- **Fields & Components:**

  - Full Name (Text field, mandatory)

  - License Number (Text field, mandatory, validated against regulatory
    database)

  - Specialization (Dropdown: General Physician, Cardiologist, etc.)

  - Location (Text field or dropdown for city/state)

  - Contact Number (Text field, numeric validation)

  - Email Address (Text field, email format validation)

  - Medical Officer Type (Dropdown with 9 options listed)

    - Options:

      1.  Assistant Civil Surgeon or above/Medical Officer in PHC, RMPs.

      2.  Medical Officer equivalent to Assistant Civil Surgeon or above
          employed in Central/State/Municipal District Board/Local
          Board/Cantonment Board/Union Board Hospital and Medical
          Officers of Public Sector Undertakings

      3.  Medical Officers (Gr-II)

      4.  Dy Civil surgeon or above

      5.  Medical Officer equivalent to Deputy Civil Surgeon or above
          employed in Central/State/Municipal District Board/Local
          Board/Cantonment Board/Union Board Hospital and Medical
          Officers of Public Sector Undertakings with at least 10 Years
          of experience

      6.  Retired Medical Officers (Gr-I)

      7.  Civil surgeon, Medical Officer not lower than that of Civil
          Surgeon or Chief Medical Officer, CMO Grade-I, Specialist
          Grade-II

      8.  Medical Officer (Allopathic) equivalent to Civil Surgeon
          employed in Central/State/Municipal District Board/Local
          Board/Cantonment Board/Union Board Hospital and Medical
          Officers of Public Sector Undertakings with at least 15 Years
          of experience

      9.  Retired Civil Surgeon, CMO Gr-I and Specialist Grade-II

  - Submit (button)

- **Business Rules:**

  - Medical Officer Type Options 1, 2, 3 → Eligible for approving up to
    ₹5 Lakh Sum Assured

  - Medical Officer Type Options 4, 5, 6 → Eligible for approving above
    ₹5 Lakh up to ₹10 Lakh

  - Medical Officer Type Options 7, 8, 9 → Eligible for approving above
    ₹10 Lakh

![](media/image1.png){width="3.3614752843394577in"
height="5.042214566929134in"}

### 4.2 Doctor Login Page

- **Fields & Components:**

  - Login ID (Text field, mandatory)

  - Password (Text field, mandatory)

  - Sing-In (button)

  - Forget Password (Link)

  - Sign Up Link (for new doctors)

- **Business Rules:**

  - After Successful Login, Doctors should move to its Dashboard page.

  - Clicking Signup Link should open the Doctor Onboarding Page.

  - Clicking Forget Password should let the user to reset their password
    using OTP given on their Mobile Number or other Aadhar based
    Authentications.

![](media/image2.png){width="4.219338363954506in"
height="4.354774715660542in"}

### 4.3 Doctor Dashboard Page

- To provide doctors with an overview of pending medical review
  proposals and allow them to upload reports after consultation.

- **Fields & Components:**

  - Doctor Name (Display Text)

  - Logout (button)

  - Pending Medical Reviews: button

  - Approved Proposals: button

  - Proposal List Table

    - Proposal ID (Clickable link)

    - Customer Name

    - Appointment Date & Time

    - Coverage Amount

    - Status (Pending / Completed)

    - Action (View Details / Upload Report)

- **Business Rules:**

  - Clicking the Proposal ID should open Customer Profile page for the
    respective proposal.

  - Clicking Pending Medical Reviews should give the list of all the
    proposals pending for medical review.

  - Clicking Approved Proposals should give the list of approved
    proposals.

![A screenshot of a doctor dashboard AI-generated content may be
incorrect.](media/image3.png){width="5.399156824146981in"
height="3.599238845144357in"}

### 4.4 Customer Profile Page

- To display customer details linked to policy proposals and allow
  scheduling of medical review appointments.

- **Fields & Components:**

> **Customer Details**: Text (To Separate Section)

- Full Name (Text field, read-only)

- Age (Text field, read-only)

- Gender (Optional dropdown)

- Contact Number (Editable text field)

- Email Address (Editable text field)

> **Insurance Details**: Text (To Separate Section)

- Policy Proposal ID (Read-only)

- Coverage Amount (Read-only)

- Proposal Status (Read-only: Pending Medical Review / Approved /
  Rejected)

> **Medical Review Status**: Text (To Separate Section)

- Appointment Date & Time (Read-only if scheduled)

- Doctor Assigned (Read-only)

- Medical Report Status (Pending / Uploaded / Reviewed)

> **Appointment History**: Text (To Separate Section)

- Table with columns: Date, Doctor, Status, Comments

- **Schedule Appointment** (Button)

- **View Medical Report** (Button -- visible after upload)

- **Policy Approval (**button)

<!-- -->

- **Business Rules:**

  - Appointment can only be scheduled if proposal status = "Pending
    Medical Review".

  - Appointment history must show all past medical reviews linked to
    proposals.

  - Clicking Policy Approval should lead the Doctor to Policy Approval
    Page for Approving/Rejecting the Proposal.

![](media/image4.png){width="4.018921697287839in"
height="6.028383639545057in"}

### 4.5 Appointment Scheduling Page

- This page is used to confirm the Appointment.

- **Fields & Components:**

  - Customer Full Name (Text field, mandatory): Auto-filled as per
    proposal.

  - Doctor Name: Auto-filled (current doctor login) and can be changed.

  - Date: Format: DD: -MM-YYYY

  - Time: Timestamp

  - Status: Dropdown

  - Schedule: Button

![](media/image5.png){width="3.5028860454943134in"
height="3.7118055555555554in"}

### 4.6 Medical Report Page

- To allow doctors to upload the medical review report (if needed) for a
  customer after the appointment.

- **Fields & Components:**

  - Customer (Read-only field)

  - Doctor (Read-only field)

  - Report Details (Text field for summary)

  - Upload File Section (Drag & drop or click to upload PDF/JPG)

  - Submit (Button)

![A screenshot of a medical report AI-generated content may be
incorrect.](media/image6.png){width="3.210326990376203in"
height="3.418882327209099in"}

### 4.7 Policy Approval Page

- To allow doctors to approve or reject the policy proposal based on
  guidelines.

- **Fields & Components:**

  - Customer Name: Auto-filled from Proposal

  - Proposal ID: Auto-filled from Proposal

  - Coverage Amount: Auto-filled from Proposal

  - Appointment Date: Auto-filled

  - Doctor Name: Auto-filled

  - Medical Report: Option to Upload

  - Approve (Button)

  - Reject (Button)

  - Comments Box (mandatory if rejecting)

  - Download Report (Button)

![](media/image7.png){width="3.8532884951881017in"
height="5.7799343832021in"}

## **5. Workflow**

The workflow for the Medical Appointment System is as follows:

![](media/image8.png){width="5.76820428696413in"
height="8.652310804899388in"}

## 6. Test Case

  ------------------------------------------------------------------------------------------
  **TC     **Functionality**   **Test Case       **Input Data**   **Expected      **Type**
  ID**                         Description**                      Result**        
  -------- ------------------- ----------------- ---------------- --------------- ----------
  TC_001   Doctor Onboarding   Onboard doctor    Name, License    Doctor profile  Positive
                               with valid        No,              created         
                               details           Specialization   successfully    

  TC_002   Doctor Onboarding   Onboard doctor    Name,            Error: "License Negative
                               without license   Specialization   number          
                               number            only             required"       

  TC_003   Credential          Verify doctor     License No =     Credentials     Positive
           Verification        with valid        valid            verified        
                               license                                            

  TC_004   Credential          Verify doctor     License No =     Error: "Invalid Negative
           Verification        with invalid      invalid          license number" 
                               license                                            

  TC_005   Approved Doctor     Display only      Approved doctor  List shows only Positive
           List                approved doctors  profiles exist   approved        
                                                                  doctors         

  TC_006   Approved Doctor     Display list when No approved      Message: "No    Negative
           List                no approved       doctors          doctors         
                               doctors                            available"      

  TC_007   Customer Details    Display linked    Proposal ID =    Customer        Positive
                               customer details  valid            details         
                                                                  displayed       

  TC_008   Customer Details    Display details   Proposal ID =    Error:          Negative
                               for invalid       invalid          "Proposal not   
                               proposal ID                        found"          

  TC_009   Customer Edit       Update contact    Mobile No        Update          Positive
                               info              updated          successful      

  TC_010   Customer Edit       Attempt to edit   Proposal amount  Error:          Negative
                               proposal details  changed          "Proposal       
                                                                  details cannot  
                                                                  be edited"      

  TC_011   Appointment         Schedule          Doctor ID +      Appointment     Positive
           Scheduling          appointment with  Proposal ID      scheduled       
                               valid doctor &                                     
                               proposal                                           

  TC_012   Appointment         Schedule          Doctor ID =      Error: "Doctor  Negative
           Scheduling          appointment with  unapproved       not authorized" 
                               unapproved doctor                                  

  TC_013   Notifications       Send              Valid            Notifications   Positive
                               notifications on  appointment      sent to doctor  
                               appointment       scheduled        & customer      

  TC_014   Notifications       Fail notification Invalid email    Error logged,   Negative
                               due to invalid    for customer     retry mechanism 
                               email                              triggered       

  TC_015   Report Upload       Upload report     BP, Sugar, ECG   Report uploaded Positive
                               with mandatory    included         successfully    
                               fields                                             

  TC_016   Report Validation   Upload report     BP, Sugar only   Error: "ECG     Negative
                               missing ECG                        mandatory"      

  TC_017   Digital Signature   Upload report     Signed report    Accepted        Positive
                               with valid                                         
                               signature                                          

  TC_018   Digital Signature   Upload report     Unsigned report  Error: "Digital Negative
                               without signature                  signature       
                                                                  required"       

  TC_019   Admin Review        Approve policy    Valid report     Proposal status Positive
                               after report      reviewed         updated to      
                               review                             Approved        

  TC_020   Rejection Reason    Reject policy     Reason field     Error: "Reason  Negative
                               without reason    blank            mandatory"      

  TC_021   Data Encryption     Verify encryption Access DB        Data encrypted  Positive
                               of stored reports                                  

  TC_022   Access Control      Unauthorized user User role =      Error: "Access  Negative
                               tries to view     Customer         denied"         
                               report                                             

  TC_023   Audit Logs          Check logs after  Appointment      Log entry       Positive
                               appointment       created          present         
                               scheduling                                         

  TC_024   Delete Health Info  Delete health     Customer request Health info     Positive
                               info on customer  valid            deleted         
                               request                            securely        

  TC_025   Delete Health Info  Attempt delete    Unauthorized     Error: "Not     Negative
                               without           request          permitted"      
                               authorization                                      
  ------------------------------------------------------------------------------------------
