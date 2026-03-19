![](media/image1.png){width="2.0888976377952755in"
height="1.3891852580927384in"}**Ministry of Communications**

**Department of Posts**

**Insurance**

**Management Solution**

**Software Requirements Specification (SRS)**

**Billing Method Change**

Oct 2025

**Index**

  -------------------------------------------------------------------
   **Sl.**  **Name**                                  **Page No.**
  --------- --------------------------------------- -----------------
      1     Version History                                 3

      2     Executive Summary                               3

      3     Business Objectives                             3

      4     Project Scope                                   3

      5     Business Rules (POLI)                           3

      6     Products                                        3

      7     Prerequisites                                   4

      8     Scenarios                                       4

      9     Functionality Description                       4

     10     Charges                                         4

     11     Channel                                         4

     12     Suggested Enhancements                          4

     13     Flow Chart / Workflow                           5

     14     Business Requirements                           6

     15     Business Logic                                  6

     16     Screen                                          6

     17     Form                                            7

     18     Reports                                         8

     19     Appendices                                      8
  -------------------------------------------------------------------

**01. Version History :**

  ----------------------------------------------------------------------------
    **Date**    **Version**  **Description of Changes**         **Author**
  ------------ ------------- ------------------------------ ------------------
   29.10.2025       1.0      Initial version                R. Venkat Ragavan

                    1.1                                     

                    1.2                                     

                    1.3                                     

                    1.4                                     

                    1.5                                     

                    1.6                                     

                    1.7                                     
  ----------------------------------------------------------------------------

**02. Executive Summary :**

This document outlines the business requirements for Billing Method
Change with Illustrations of Insurance Management System (IMS). It aims
to processing of Billing Method Change of a Policy as per the Rules &
Regulations inforce.

**03. Business Objectives :**

The IMS project addresses the requirements of Billing Method Change of a
Policy by the Insurant and when required.

**04. Project Scope :**

Billing Method Change will be a Service Request of the Insurance Admin
system used by Insurant / Customer as and when required to change the
method.

**05. Business Rules (POLI) :**

49\. Policies held by persons who have left the Government Service. - If
an insured person resigns or retires or is dismissed from the service of
Government, his policy holds good so long as the premium due are
regularly paid by him on the first day of the month or within the period
of grace at the Post office selected by him. As soon as the connection
of the insured person ceases with the Government, he should apply to the
Postmaster General/Head of Division for a Premium Receipt Book informing
him of the name of the Post Office at which the 1st premium, while in
Government service, was deducted and the Post office at which he desires
to pay future premium in cash. A copy of this application should also be
endorsed to the Postmaster of the place at which future payments of
premiums are desired to be made in cash. If the Premium Receipt Book is
not received by the time the next premium after his quitting government
service falls due, he should pay the amount by the due date in cash at
the selected Post office producing a certificate from his last
Disbursing Officer in the form appended at the end of this rule. In such
a case, the concerned Postmaster would grant a receipt for the amount in
from ACG-67. Subsequent premium will be paid in cash on production of
the receipt for the previous month's premium, till the Premium Receipt
Book is received by him. Thereafter, the due premium shall be paid on
production of the Premium Receipt Book and receipt for the amount will
then be given only in the Premium Receipt Book.

**06. Products : All types of Products in PLI & RPLI**

**07. Prerequisites :**

- Policy should be Active Status.

- No pending Premium Payments.

- No pending Service Requests.

- Billing Method Change Form.

- Copy of Policy Bond.

- Premium Receipt book.

- Copy of Pay Recovery Certificate for last 6 months for Existing
  Employees. Copy of Certificate from DDO for New Employees.

- Self Attested Copy of ID & Address Proof.

**08. Scenarios :**

- Cash to Pay

- Pay to Cash

# **09. Functionality Description :**

- BMC means any Policy can be changed from **Cash to Pay and vice
  versa**.

- To raise a new request for Billing Method Change for the policy, user
  should go to 'Home Page' and click on 'Service Request Indexing' on
  the screen, which will navigate the user to 'Request Indexing' screen.

- Request Type selected as "Billing Method Request" for doing Billing
  Method change respectively.

- Insurant / DoP User needs to Scan & Upload the Docs if any Policy from
  Cash to Pay.

- Select the mandatory list of documents and Click Submit.

- The request will be forwarded to the Approver Inbox.

- Approver to approve the request.

- Navigate to the Policy Search Screen and search the policy number. The
  new payment method should be displayed .

**10. Charges : Free**

**11. Channel :** Customer Portal, Mobile App DoP Solution.

**12. Suggested Enhancements :**

  -----------------------------------------------------------------------
  **Sl.**   **Current Process**        **Suggested Process**
  --------- -------------------------- ----------------------------------
  01\.      Manual mode                Digital

  02\.      Portal                     Portal, App

  03\.      Manual : DE \> Approver    Manual : DE \> Approver

  04\.      Digital : Portal \> DE \>  Digital : Approver or Auto
            Approver                   Approver

  05        No Document upload         Document is mandatory if Cash to
                                       Pay
  -----------------------------------------------------------------------

**13. Flow chart / Workflow :**

Insurant Submits Physical Request a/w Documents to PO

Reconciliation by Approver & Approved

Verification of Docs

Move to Approver Inbox

Generation of Service Request ID & Receipt to Insurant, SMS, email

Scanning, Uploading of Docs if any

System shall check eligibility conditions as per Business Rules

Reconciliation, Indexing by DE

SMS, email Alerts to Insurant

Reports to PO & Admin Units

**b. Electronic Request thro Portal / App & Auto / Manual Process by DoP
Users :**

Reconciliation by Approver & Approved

Verification of Docs

Move to Approver Inbox

Generation of Service Request ID & Receipt to Insurant, SMS, email

Scanning, Uploading of Docs if any

System shall check eligibility conditions as per Business Rules

Reconciliation, Indexing by Insurant in Portal, App

SMS, email Alerts to Insurant

Reports to PO & Admin Units

**14. Business Requirements :**

  ------------------------------------------------------------------------
   **RFP Req  **Requirement Description**                        **FRS Ref
     No.**                                                         ID**
  ----------- -------------------------------------------------- ---------
   SR-BMC-1   Policy should be in **Active** status              

   SR-BMC-2   System should allow the Insurant to change the BMC    466
              through Manual, Portal, App, TPA modes.            

   SR-BMC-3   System should allow the DoP Users to process the      466
              BMC through IMS, Portal as per Workflow            

   SR-BMC-4   System should allow the DoP Users to Approve the      466
              BMC requests received from Customer Portal,        
              Customer Mobile App                                

   SR-BMC-5   System should not allow to raise by any Users      
              other than Insurant & DoP Users in any Channel     

   SR-BMC-6   System should not allow if any Premium Payment is  
              due & pending                                      

   SR-BMC-7   System should not allow if any Service Request is  
              pending. If Pending then System should show Error  
              message                                            

   SR-BMC-8   System should store all the Events in History with 
              User IDs, Channel, Timestamp etc till lifetime of  
              the Policy. And also maintains Audit Trails.       

   SR-BMC-9   System should trigger real time SMS, email to        1178
              Insurant                                           

   SR-BMC-10  Solution should have capability to download all    
              the Policies in Category wise (Pay, Cash) in XL,   
              CSV, PDF formats by DoP IT, Business Users         

   SR-BMC-11  Solution should have a mechanism to Reconcile all  
              the Pay Policies as per DoP HR Data. DoP IT,       
              Business Users needs to upload the HR Data into    
              IMS and IMS needs to reconcile.                    

   SR-BMC-12  Solution should not allow the Index, raise Request 
              other than Customer Portal, Customer Mobile App &  
              DoP Solution                                       

   SR-BMC-13  System should allow the DoP IT, Business Team to   
              change BMC on Bulk Upload method either in XL or   
              CSV                                                
  ------------------------------------------------------------------------

**15. Business Logic : as in Sl, 13.**

**16. Screen :**

**a. Insurant in Portal or App :**

**Input :**

Service Request Type : Billing Method Change (drop down)

Policy Number :

Current date :

**System shall collect the Data & Display :**

Policy Number: Customer ID :

Name: Policy Issue Date :

Policy status : Current Product Name:

Paid Till Date : Current Method : Cash or Pay

Convert to : drop down Cash or Pay

Payable Office : (all POs under drop down) -- if Pay to Cash.

Document Upload :

**b. Indexing Screen by DoP User :**

Service Type : Billing Method Change (drop down)

Policy Number :

Current date :

**System shall collect the Data & Display :**

Policy Number: Customer ID :

Name: Policy Issue Date :

Policy status : Current Product Name:

Paid Till Date : Current Method : Cash or Pay

Convert to : drop down Cash or Pay

Payable Office : (all POs under drop down) -- if Pay to Cash.

Document Upload :

**17. Form :**

**Billing Method Change Form**

Government of \_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_

Ministry of \_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_

Department of \_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_

Office of \_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_

No. \_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_ (Name of Station) Dated
the \_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_

**Certificate of the Disbursing Officer regarding premium deduction on
account of PLI**

This is to Certify that Shri / Ms / Smt.
\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_
is working as \_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_ in
the Office of \_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_ at
\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_\_ Station. And he / she
is having the below PLI Policie/s and Premium was / were deducted from
Salary on every month without fail.

  ----------------------------------------------------------------------------------------
  **Sl.**   **Policy   **Name of the     **Premium in **Tax   **Frequency**   **Last
            No.**      Insurant**        Rs.**        Rs.**                   deducted**
  --------- ---------- ----------------- ------------ ------- --------------- ------------
  01\.                                                                        

  02\.                                                                        

  03\.                                                                        

  04\.                                                                        

  05\.                                                                        
  ----------------------------------------------------------------------------------------

Office Seal

> (With full name in Block capitals &
>
> Designation of Disbursing officer)

**18. Report format to Operational & Admin Units :**

**19.** **Appendices :**

POLI Rules

\*\*\* End \*\*\*
