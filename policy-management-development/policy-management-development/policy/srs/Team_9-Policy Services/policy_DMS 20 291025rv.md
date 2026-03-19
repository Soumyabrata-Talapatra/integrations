![](media/image1.png){width="2.0888976377952755in"
height="1.3891852580927384in"}**Ministry of Communications**

**Department of Posts**

**Insurance**

**Management Solution**

**Software Requirements Specification (SRS)**

**Document Management System**

Oct 2025

**Index**

  -------------------------------------------------------------------
   **Sl.**  **Name**                                  **Page No.**
  --------- --------------------------------------- -----------------
      1     Version History                                 3

      2     Executive Summary                               3

      3     Business Objectives                             3

      4     Stakeholders                                    3

      5     Type of Documents                               3

      6     Functionality Description                       4

      7     Business Rules (POLI)                           4

      8     Prerequisites                                   4

      9     File Format / Nomenclature                      4

     10     Suggested Enhancements                          5

     11     Business Requirements                           5

     12     Business Logic                                  7

     13     Screen                                          7

     14     Error Messages                                  8

     15     Report format to Operational & Admin           10
            Units                                   

     16     Appendices                                     10
  -------------------------------------------------------------------

**01. Version History :**

  ----------------------------------------------------------------------------
    **Date**    **Version**  **Description of Changes**         **Author**
  ------------ ------------- ------------------------------ ------------------
   20.10.2025       1.0      Initial version                R. Venkat Ragavan

                    1.1                                     

                    1.2                                     

                    1.3                                     

                    1.4                                     

                    1.5                                     

                    1.6                                     

                    1.7                                     
  ----------------------------------------------------------------------------

**02. Executive Summary :**

This document outlines the business requirements to Scan & upload the
required Documents in the Document Management System for the usage of
Users of Insurance Management System (IMS) to View Retrieve, Download,
Print as and when requires.

**03. Business Objectives :**

The IMS project addresses the requirements of Scanning & Uploading the
Documents of a Policy by the Insurant or DoP User and when required.

**04. Stakeholders :**

- Insurant.

- DoP Users like DE, Approver

- DoP License Issuing Authority (DH, RH).

- All types of Sales Force

- Medical Examiners

- Under Writers

- DoP Business Team.

- DoP IT Team.

- DoP Accounts Team.

- Actuary.

- DPLI Kolkatta.

- Investment Dn, Mumbai.

- TPA.

**05. Types of Documents :**

- Proposal Forms.

- Any other Documents related to Proposals.

- Photograph

- Specimen Signature

- Cheque.

- KYC.

- Specimen Signature.

- Consent, Declaration etc.

- Policy Bond.

- All types of Service Request forms.

- Medical Certificates, Life Certificate, Divorce Documents.

- Documents related to Maturity, Death, Surrender like Legal Heir,
  Legal, Death Certificate etc.

- Documents related to all types of Sales Force like Educational
  Qualifications, KYC, Income etc.

- Documents related to all types of Medical Examiner like License, Place
  etc.

- Documents related to all types of Under Writers like Educational
  Qualifications, License etc.

- Documents related to Actuary.

- Documents related to DPLI Kolkatta, Investment Dn Mumbai etc.

- Documents related to other Stakeholders like IRDA, RBI, MoF etc.

- Documents related to Third Party Service Providers.

- Documents related to Payment Gateway.

- Digital Data related to DoP Admin.

- All types of Rules, Regulations, Circulars, SOPs, User Manuals etc
  issued by PLI Dte.

- Legal Documents.

# **06. Functionality Description :**

- Before processing the Proposals, Service Requests the required Forms,
  Documents etc must be Scanned & readily available for Upload as
  Uploading is a part of Workflow.

- **Proposal** : Insurant or DoP Users need to Scan all the required
  documents & upload into CMS.

- **Service Request :** Insurant or DoP Users need to Scan all the
  required documents & upload into CMS.

- Soft copy of the documents available in DMS shall be downloaded by DoP
  Users, Insurant as and when required for processing in IMS as per the
  Workflow.

**07. Business Rules (POLI) :**

To amend.

**08. Prerequisites :**

- Scanner.

- Convert into pdf format.

- Integration with IMS to upload the Documents.

**09. File Format / Nomenclature : in pdf, JPEG, PNG.**

  ---------------------------------------------------------------------------------------
  **Sl.**   **Name of Service     **Algorithm in PLI**
            Request**             
  --------- --------------------- -------------------------------------------------------
  01\.      Proposal Form         PLI\_\<customerId\>\_\<ProposalNo\>\_\<ddmmyy01\>.pdf

  02\.      Proposal Form KYC     PLI\_\<customerId\>\_\<ProposalNo\>\_\< ddmmyy 02\>.pdf

  03\.      Policy Servicing      PS\_\<customerId\>\_\<PolicyNo\>\_\<SerReqNo\>\_\<
                                  ddmmyy 01\>.pdf

  04\.      Policy Servicing Addl PS\_\<customerId\>\_\<PolicyNo\>\_\<SerReqNo\>\_\<
            Docs                  ddmmyy 02\>.pdf
  ---------------------------------------------------------------------------------------

  -----------------------------------------------------------------------------
  **Sl.**   **Name of Service     **Algorithm in RPLI**
            Request**             
  --------- --------------------- ---------------------------------------------
  01\.      Proposal Form         RPLI\_\<customerId\>\_\<ProposalNo\>\_\<
                                  ddmmyy 01\>.pdf

  02\.      Proposal Form KYC     RPLI\_\<customerId\>\_\<ProposalNo\>\_\<
                                  ddmmyy 02\>.pdf
  -----------------------------------------------------------------------------

**10. Suggested Enhancements :**

  -----------------------------------------------------------------------------
  **Sl.**   **Current Process**       **Suggested Process**
  --------- ------------------------- -----------------------------------------
  01\.      DataCap & FileNet Utility May be discarded

  02\.      ECMS Alternate Solution   May be discarded

  03\.      ECMS Optional             Part of Workflow in IMS

  04\.      New                       Users needs to upload Docs as per the IMS
                                      Workflow in pdf, JPEG, PNG format. IMS
                                      must provide facility to Upload in the
                                      relevant Screen, Card.
  -----------------------------------------------------------------------------

**11. Business Requirements :**

  -----------------------------------------------------------------------
  **RFP   **Requirement Description**                           **FRS Ref
  Req                                                             \#**
  No.**                                                         
  ------- ----------------------------------------------------- ---------
  CMS-1   File Formats shall be pdf, JPEG, PNG and other           874
          Formats are not permitted.                            

  CMS-2   System should restrict FILE SIZE limits as per the    808, 890
          Business Standards, usually 10mb per file as per      
          Workflow. Other than Workflow Users, no limit         

  CMS-3   All types of Uploads are part of Workflow as per the     875
          respective Users like Insurant, DE, Approver, License 
          Issuing Authority, Business Team, IT Team etc.        

  CMS-4   System should have a Unique Identifier to all the        878
          Docs for easy retrieval                               

  CMS-5   System should allow the Insurant to Upload all the      1225,
          required Forms during Proposal stage.                 1259, 30

  CMS-6   System should allow the Insurant to Retrieve, View,   
          Download & Print all the Documents related to his /   
          her Policy only but not limited to Proposal Form,     
          Policy Bond, Service Request Forms etc.               

  CMS-7   System should allow the DoP Users like DE, Approver   816, 889
          etc to Upload all the required Forms during Proposal  
          stage.                                                

  CMS-8   System should allow the DoP Users like DE, Approver   816, 882
          etc to Retrieve, View, Download & Print all the       
          Documents related to any Policy.                      

  CMS -9  System should allow the Insurant to Upload all the     50, 149
          required Forms, Docs in all types of Service Requests 
          stage related to his / her Policy only.               

  CMS -10 System should allow the DoP User like DE, Approver     50, 149
          etc to Upload all the required Forms, Docs in all     
          types of Service Requests stage.                      

  CMS -11 System should allow all the Sales Force to Upload the   877,
          required Forms, Docs.                                 1005, 30

  CMS -12 System should allow the License Issuing Authorities     1034
          to Upload the required Forms, Docs.                   

  CMS -13 System should allow the Sales Force to Retrieve,      
          View, Download & Print all the Forms, Docs related to 
          Policies under his / her Agency. And also Docs        
          related to his / her Agency like License etc.         

  CMS -14 System should allow all the Medical Examiners to        1090
          Upload the required Forms, Docs.                      

  CMS -15 System should allow the Medical Examiners to          
          Retrieve, View, Download & Print all the Forms, Docs  
          related to his / her Medical Documents only as per    
          Business Rules.                                       

  CMS -17 System should allow all the Under Writers to Upload   547, 620,
          the required Forms, Docs.                             663, 825,
                                                                   828

  CMS -18 System should allow the Under Writers to View all the   674,
          Forms, Docs related to the Policies assigned to them  
          as per Business Rules.                                

  CMS -19 System should allow the Actuary to View the Forms,    
          Docs, Data of all the Policies as per Business Rules. 

  CMS -20 System should allow the DPLI Kolkatta Users to View   
          the Forms, Docs, Data of all the Policies as per      
          Business Rules.                                       

  CMS -21 System should allow the Investment Dn to View the     593, 597
          Forms, Documents, Data of all the Policies as per     
          Business Rules.                                       

  CMS -22 System should allow the DoP IT Team to Upload the        754
          Data, Docs, Forms.                                    

  CMS -23 System should allow the DoP Business Team to Upload   597, 754
          the Data, Docs, Forms.                                

  CMS -24 System should allow all the DoP Users to View,        
          Download the Forms, Documents, Data hosted by DoP IT  
          Team & Business Team.                                 

  CMS -25 System should allow the DoP Admin to Upload, View,    
          Download the Data, Forms, Documents available in CMS. 

  CMS -26 System should allow the authorized Users to View,     
          Download the Digital Data in CSV, Excel, PDF formats. 

  CMS -27 System should allow the restricted Users to **Search     881
          (exclusive Docs Search Card)** the available          
          Documents, Forms related to any Policy as per         
          Business Rules.                                       

  CMS -28 System should display the successful upload window    
          "Uploaded successfully" with Green Button.            

  CMS -29 System should allow some other Users also like IT        192
          Team, Business Team etc apart from the IMS Workflow.  

  CMS -30 If any Document uploading is **Optional** as per      
          Business Rules then System should allow on par with   
          Workflow.                                             

  CMS -31 System should allow the Bulk Customer to Upload the      460
          Data, Docs, Forms.                                    

  CMS -32 System should allow the Fund Managers to Upload, View    606
          the Data, Forms, Documents available in CMS. Further  
          allow to Download the Digital Data in XL, CSV, PDF    
          Formats.                                              

  CMS -33 System should have a provision, integration to fetch, 792, 887
          View, Validate & download the Data from DigiLocker,   
          UID, eKYC, PAN, NSDL, PennyDrop, eIA, DoP Solutions   
          etc.                                                  

  CMS -34 System should be capable to Auto Storing of any Data  844, 893
          from integral Systems, Modules etc and retrievable by 
          the Users like Payroll, Group Upload, Cash Upload     
          (Meghdoot), Incentive Statement, Commission           
          Statement, License, MER Data, UW Data, various        
          Reports, Recon, Digital payments Data, Claims, Manual 
          & Auto generated Letters, BCP, KYC, etc.              

  CMS -35 System should be capable to Scan & Upload Docs in OCR   1204
          & ICR                                                 

  CMS -36 System should be capable to Scan & Upload Cheque        1220
          Leaf, Cheque Returned details etc                     

  CMS -37 System should be capable to manage the Data in DMS as  58, 884
          per Data Retention Policy of the Dept which includes  
          Archival, Soft & Hard Deletion.                       

  CMS -38 System should allow the IT, Business Users to         893, 895
          configure Archival Policy with Timeline for all types 
          of Data through Front end.                            

  CMS -39 System should have a provision to retrieve the           60
          Archived Data, Soft Deleted Data to a restricted User 
          Groups                                                

  CMS -40 System should allow the Users to upload the Docs         875
          through Online & Offline                              

  CMS -41 System should allow the Insurant, Users etc to Upload    886
          Docs, Forms from any Channels like Portal, App, TPA   
          etc                                                   

  CMS -42 **Configuration :** System should allow the IT,          895
          Business Users to Archie the Data through Front end.  

  CMS -43 **Configuration :** System should allow the IT,          890
          Business Team to Configure the FILE FORMAT, FILE      
          SIZE, Algorithm through Front end.                    

  CMS -44 System must trigger SMS, email, critical Alerts to IT    896
          Team due to Threshold breach, Security Threat etc.    
  -----------------------------------------------------------------------

**12. Business Logic :**

  ------------------------------------------------------------------------
  **Activity**       **Docs**                          **By User**
  ------------------ --------------------------------- -------------------
  Proposal Indexing  Proposal Forms, KYC, Docs         Insurant

  Proposal Indexing  Proposal Forms, KYC, Docs         DE, Approver

  Policy Servicing   Relevant Forms, KYC, Docs         Insurant

  Policy Servicing   Relevant Forms, KYC, Docs         DE, Approver

  Digital Data                                         DoP IT Team

  Forms, SOPs, Rules                                   DoP Business Team

  Documents          Related to Sales Force, MER, UW,  Relevant Teams
                     Licensing Authority, Actuary,     
                     Investment Dn, Business Team, IT  
                     Team.                             
  ------------------------------------------------------------------------

**13. Screen :**

- "Upload" Button at the bottom of the Screens of a. Proposal ie)
  Indexing, DE, Approver, b. Service Request, c. Sales Force, d. Medical
  Examiner, e. Under Writer.

- "View Documents" at the **Right Top** of the Screens of a. Proposal
  ie) Indexing, DE, Approver, b. Service Request, c. Sales Force, d.
  Medical Examiner, e. Under Writer.

![](media/image2.png){width="3.8402788713910763in"
height="2.2131397637795276in"}

![](media/image3.png){width="3.201771653543307in"
height="0.5296751968503937in"}

![](media/image4.png){width="6.284202755905512in"
height="2.051783683289589in"}

**14. Error Messages :**

**Indexing Level by Insurant & DoP User :**

  --------- --------- ------------------- ------------------ --------------------
  **Sl.**   **Field   **Condition**       **Error Message**  **Required Action**
            Name**                                           

  1         Policy No If numeric field is Policy Number      Users must enter a
                      entered with        should be in the   numeric value.
                      non-numeric value.  numeric format     

  2         Policy No If Policy Number    NA                 System should not
                      field is entered                       allow entering more
                      with more than                         than 13 digits.
                      allowed digits.                        

  3         Policy No If left blank.      Missing Policy No  Users must enter the
                                                             Policy No.

  4         Date      Date field filled   Date should be in  Users must enter the
                      with non-date       the date format    Date in the date
                      value, special                         format.
                      characters other                       
                      than \'/\'                             

  5         Date      Date entered in     Date should be in  Users must enter the
                      MMDDYYYY format or  DDMMYYYY format    Date in DDMMYYYY
                      YYYYMMDD format                        format.

  6         Date      Date entered is     Date entered is    Users must enter a
                      invalid             invalid            valid date.

  7         Date      If Date field is    Date should be     No back date is
                      mentioned in the    more than or equal allowed in case of
                      past where the      to today\'s date   quote. Users must
                      system does not                        enter a date which
                      expect so                              is more than or
                                                             equal to today's
                                                             date.

  8         Date      Entered date is     Date should not be Users must enter a
                      lesser than policy  less than Policy   date which is not
                      issue               issue date         less than Policy
                                                             issue date.

  9         Date      If left blank       Missing Date       Users must enter a
                                                             date
  --------- --------- ------------------- ------------------ --------------------

**Processing level :**

**Data Entry Stage :**

  --------------------------------------------------------------------------------
   **Sl.**  **Field     **Condition**        **Error Message**    **Required
            Name**                                                Action**
  --------- ----------- -------------------- -------------------- ----------------
      1     Product     If \'None\'          Please select any     
                                             product              

      2     Service     If \'None\'          Please select any    Insurant, DoP
            Type                             Service Type         User needs to
                                                                  select any one
                                                                  of the Service
                                                                  Request

      3     Request     If document name is  Missing document     On selection of
            Missing     selected and date    request date and     the document
            Documents   and status are left  status               from the
                        blank                                     dropdown request
                                                                  date and status
                                                                  should also be
                                                                  selected

      4     Document    If left blank        Select at least one   
            Name                             document             

      5     Document    Date field filled    Date should be in     
            Request     with non-date value, the date format      
            Date        special characters                        
                        other than \'/\'                          

      6     Document    Date entered in      Date should be in     
            Request     MMDDYYYY\            DDMMYYYY format      
            Date        format or YYYYMMDD                        
                        format                                    

      7     Document    Date entered is      Document Request\     
            Request     invalid              date entered is      
            Date                             invalid              

      8     Document    Entered date is      Date should not be    
            Request     lesser than policy   less than Policy     
            Date        issue date           issue date           

      9     Document    If left blank        Missing document      
            Request                          request date         
            Date                                                  

     10     Users       User trying to       Not an Authorized    Pl try to Upload
                        Upload must be an    User                 with Authorized
                        Authorized User as                        User ID.
                        per Workflow                              

     11     Users       User trying to View, Not an Authorized    Pl try to View,
                        Download must be an  User                 Download with
                        Authorized User as                        Authorized User
                        per Workflow                              ID.

     12     Format      Should be PDF, JPEG, Supports PDF, JPEG,   
                        PNG                  PNG Formats only.    

     13     Size        File size should be  Size of the file is  Pl reduce the
                        below 10mb           above 10mb           size & try to
                                                                  upload.

     14     Digital     Upload any Data by   Not an Authorized    Pl try to Upload
            Data        selected Users       User                 with Authorized
                                                                  User ID.
  --------------------------------------------------------------------------------

**Approver Stage :**

  -------------------------------------------------------------------------------
   **Sl.**  **Field     **Condition**      **Error        **Required Action**
            Name**                         Message**      
  --------- ----------- ------------------ -------------- -----------------------
      1     Reason for  If not entered.    Missing Reason The user must enter the
            Reject                         for Reject     reason for rejection.

      2     Reason for  If only special    Reason for     The user must enter the
            Reject      characters or      reject should  valid reason for
                        numbers or         be valid       rejection.
                        repeated values                   
                        are entered.                      

      3     List of     If left blank.     Missing list   The user must select at
            documents                      of documents   least one of the
                                                          documents.

      4     Request     If document name   Missing        On selection of the
            Missing     is selected and    document       document from the
            Documents   date and status    request date   drop-down, the user
                        are left blank     and status     must also select the
                                                          request date and
                                                          status.
  -------------------------------------------------------------------------------

**15. Report format to Operational & Admin Units :**

**16.** **Appendices :**

POLI Rules

ECMS SOP.

\*\*\* End \*\*\*
