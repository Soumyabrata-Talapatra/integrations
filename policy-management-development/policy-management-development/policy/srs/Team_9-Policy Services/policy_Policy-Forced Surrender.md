DEPARTMENT OF POSTS

MINISTRY OF COMMUNICATIONS & IT

GOVERNMENT OF INDIA

System Requirement Specification (SRS)

Policy Forced Surrender

## Brief Description

The purpose of this business specification is to outline the business
requirements necessary to Force Surrender of a Policy

The document covers following important aspects:

- Forced Surrender when Loan O/S amount including Interest is more than
  Gross Surrender value

## Standard Flow

System has to generate a service request as and when the Policy crosses
the value /rule framed by Business term in respect of outstanding loan
cases as Forced Surrender request (or also called as Auto Surrender
Flow) and will be reserved in Approver Queue of the Postmaster.

Approver-\> Either Approve or Reject 🡪 Letter generation

## Sub-Standard Flows

- Forced Surrender

  - system will check when Gross Surrender amount is less than O/S Loan
    Principal + O/s Loan interest -\> System will change the policy
    status as 'Terminated as Auto Surrender' due to loan🡪Policy
    Auto-terminated🡪All pending Financial/Non-Financial requests are
    Auto-terminated

<!-- -->

-  Auto Surrender Flow

<!-- -->

- Approver rejects the request 🡪 Request status is Rejected.

- Approver Queue is not available then request will remain as pending

- Approver Queue is available, but no Approver user is tagged against
  that queue than request will remain as pending

- Approver mapped against the office code from where the Original Loan
  was approved is transferred (Transfer/Relocation) 🡪 Requests will be
  mapped against another user based on the Approver queue available for
  the mentioned Office Code.

# Functionality Description

**Forced Surrender**

In the event of any three defaults (failure to pay) in the payment of
half yearly interest, the Postmaster General/ Head of Division will be
entitled to surrender the policy and to apply the surrender value
thereof in payment of the said loan and interest. The balance, if any,
of such surrender value if adequate for a paid-up value of Rs. 10,000/-
shall be utilized for the issue of such a paid-up policy; otherwise, the
amount will be paid in cash to policy holder entitled thereto.

# Illustration and Understanding Document

+---------+---------------------------+--------------------------+
| Bus     | Illustration              | System Expected Result   |
| Rule no |                           |                          |
+=========+===========================+==========================+
| SR_FS_1 | Force Surrender due to    | (i)On Active policy      |
|         | Loan : Rules\             | status , Force surrender |
|         | \                         | rule will be triggered   |
|         | Loan Disbursement Date :  | when outstanding Loan    |
|         | 15th Dec 2012\            | repayment amount         |
|         | Loan Sanction Amount :    | (principal plus          |
|         | 10,000\                   | interest) is =\>100      |
|         | Gross Surrender Value =   |                          |
|         | 2000\                     | \% of SV. If due amount  |
|         | Outstanding Loan          | remains unpaid within 30 |
|         | Principal Amount = 1500\  | days from the date of    |
|         | Outstanding Loan Interest | 3rd notice/reminder,     |
|         | Amount = 500\             | request will be          |
|         | Policy Status : Active\   | forwarded to Postmaster  |
|         | \                         | queue for further        |
|         | \                         | Approval or Rejection.   |
|         | Forced Surrender Rule to  |                          |
|         | be checked as on First of |                          |
|         | Every Month               |                          |
+---------+---------------------------+--------------------------+
| SR_FS_2 | Force Surrender due to    | Forced surrender rule    |
|         | Loan : Rules\             | triggers when loan O/S + |
|         | \                         | loan int = 2000 is       |
|         | Loan Disbursement Date :  | equals to GSV = 2000,    |
|         | 15th Dec 2012\            | then outstanding Loan    |
|         | Loan Sanction Amount :    | repayment amount         |
|         | 10,000\                   | (principal plus          |
|         | Gross Surrender Value =   | interest) is =\>100      |
|         | 2000\                     |                          |
|         | Outstanding Loan          | \% of SV. 3rd Reminder   |
|         | Principal Amount = 1500\  | will be sent to the      |
|         | Outstanding Loan Interest | user. If due amount      |
|         | Amount = 500\             | remains unpaid within 30 |
|         | Policy Status : Active\   | days from the date of    |
|         | \                         | 3rd notice/reminder,     |
|         | \                         | request will be          |
|         | Forced Surrender Rule to  | forwarded to Postmaster  |
|         | be checked as on First of | queue for further        |
|         | Every Month               | Approval or Rejection.\  |
|         |                           | (iii) If the request is  |
|         |                           | approved by approver     |
|         |                           | than Policy status       |
|         |                           | changed to \'Terminated  |
|         |                           | surrender due to loan\'\ |
|         |                           | /Terminated due to Auto  |
|         |                           | surrender (iv) A Request |
|         |                           | ID has to be generated   |
|         |                           |                          |
|         |                           | \(v\) Reason to be       |
|         |                           | displayed as \' Forced   |
|         |                           | Surrender due to Loan\'\ |
|         |                           | (vi) Request ID, print,  |
|         |                           | view history to be       |
|         |                           | displayed in disable     |
|         |                           | mode.\                   |
|         |                           | (vii) Financial History  |
|         |                           | , event for forced       |
|         |                           | surrender to be          |
|         |                           | generated\               |
|         |                           | (viii) Owner to be       |
|         |                           | displayed as \'System\'\ |
|         |                           | (ix) Status to be        |
|         |                           | displayed as             |
|         |                           | \'Auto-complete\'\       |
|         |                           | (x) Pending Request      |
|         |                           | \'Name Change\' will get |
|         |                           | terminated (xi) On       |
|         |                           | rejection of request by  |
|         |                           | approver, Customer is    |
|         |                           | required to pay the due  |
|         |                           | amount                   |
|         |                           | (principal+interest).    |
|         |                           | Policy status will be    |
|         |                           | updated accordingly.     |
+---------+---------------------------+--------------------------+
| SR_FS_3 | Force Surrender due to    | On policy status PM ,    |
|         | Loan : Pending Request\   | Force surrender rule     |
|         | \                         | will not trigger\        |
|         | Loan Disbursement Date :  | \                        |
|         | 15th Dec 2012\            | Same rule to be followed |
|         | Loan Sanction Amount:     | for                      |
|         | 10,000\                   |                          |
|         | Gross Surrender Value =   | Pending and Terminated   |
|         | 2000\                     | Maturity/Surrender/Death |
|         | Outstanding Loan          |                          |
|         | Principal Amount = 1500\  |                          |
|         | Outstanding Loan Interest |                          |
|         | Amount = 500\             |                          |
|         | Policy Status : Pending   |                          |
|         | Maturity                  |                          |
+---------+---------------------------+--------------------------+
| SR_FS_4 | Force Surrender due to    | \(i\) Any loan           |
|         | Loan :\                   | collection not allowed   |
|         | Impact on post collection | on policy status error   |
|         | & processing\             | message to be displayed  |
|         | \                         | \' Policy status not     |
|         | Loan Disbursement Date :  | eligible\'\              |
|         | 15th December 2012\       | \                        |
|         | Loan Sanction Amount      | (ii)Any further          |
|         | 10,000 Rs.\               | processing not allowed   |
|         | Gross Surrender Value =   | and error message to be  |
|         | 2000\                     | displayed as \'Policy    |
|         | Outstanding Loan          | status not eligible\'\   |
|         | Principal:1500\           | \                        |
|         | Outstanding Loan Interest | (iii) Any collection     |
|         | :500                      | through Bulk/meghdoot    |
|         |                           | not allowed              |
+---------+---------------------------+--------------------------+
| SR_FS_5 | Force Surrender due to    | (i) Request ID to be     |
|         | Loan\                     |     generated            |
|         | \                         |                          |
|         | Loan Disbursement Date :  | (ii) \(iii\) Reason to   |
|         | 15th Dec 2012\            |      be displayed as \'  |
|         | Loan Sanction Amount :    |      Forced Surrender    |
|         | 10,000\                   |      due to Loan\'\      |
|         | Gross Surrender Value =   |      (iv)Request ID,     |
|         | 2000\                     |      print, view history |
|         | Outstanding Loan          |      to be displayed in  |
|         | Principal Amount = 1500\  |      disable mode.\      |
|         | Outstanding Loan Interest |      (v) Financial       |
|         | Amount = 500              |      History , event for |
|         |                           |      forced surrender to |
|         |                           |      be generated\       |
|         |                           |      (vi) Owner to be    |
|         |                           |      displayed           |
|         |                           |      \'System\'\         |
|         |                           |      (vii) Status to be  |
|         |                           |      displayed as        |
|         |                           |      \'Auto-complete\'   |
+---------+---------------------------+--------------------------+

+----------+----------------------------+--------------------------+
| SR_FS_6  | Auto Surrender             | Loan Capitalization      |
|          |                            | (Principle plus          |
|          | Policy No: TN-82048-CC     | Interest) =95% of the    |
|          |                            | Policy's Surrender       |
|          | Policy Holder Name: xxxxx  | Value-1st reminder       |
|          |                            | intimation by            |
|          | Product Name: Santosh      | Postmaster.              |
|          |                            |                          |
|          | Surrender Value as on      |                          |
|          | Date-10 Jan 2025-49400     |                          |
|          |                            |                          |
|          | Loan Principal-50000       |                          |
|          |                            |                          |
|          | Loan Interest-2000         |                          |
|          |                            |                          |
|          | Loan Capitalization is 95% |                          |
|          | of Surrender Value.        |                          |
+==========+============================+==========================+
| SR_FS_7  | Auto Surrender             | Loan Capitalization      |
|          |                            | (Principle plus          |
|          | Policy No: TN-820948-CS    | Interest) \>98% of the   |
|          |                            | Policy's Surrender       |
|          | Policy Holder Name: XXXXX  | Value-Only 2nd reminder  |
|          |                            | intimation by            |
|          | Product Name: Santosh      | Postmaster.              |
|          |                            |                          |
|          | Surrender Value as on      |                          |
|          | Date-10 Jun-26-51480       |                          |
|          |                            |                          |
|          | Loan Principal-50000       |                          |
|          |                            |                          |
|          | Loan Interest-2000         |                          |
|          |                            |                          |
|          | Loan Capitalization is 99% |                          |
|          | of Surrender Value.        |                          |
+----------+----------------------------+--------------------------+
| SR_FS-8  | Auto Surrender             | Loan Capitalization      |
|          |                            | (Principle plus          |
|          | Policy No: TN-820948-CS    | Interest) \>100% of the  |
|          |                            | Policy's Surrender       |
|          | Policy Holder Name: xxxx   | Value-Only 3rd and Final |
|          |                            | reminder intimation by   |
|          | Product Name: Santosh      | Postmaster.              |
|          |                            |                          |
|          | Surrender Value as on      |                          |
|          | Date-10 Jan 26 is 54600    |                          |
|          |                            |                          |
|          | Loan Principal-50000       |                          |
|          |                            |                          |
|          | Loan Interest-2000         |                          |
|          |                            |                          |
|          | Loan Capitalization is     |                          |
|          | 105% of Surrender Value.   |                          |
+----------+----------------------------+--------------------------+
| SR_FS_9  | Auto Surrender             | Loan Capitalization      |
|          |                            | (Principle plus          |
|          | Policy No: TN_82048-CC     | Interest) \>100% of the  |
|          |                            | Policy's Surrender       |
|          | Policy Holder Name: XXXXX  | Value-3rd and Final      |
|          |                            | reminder intimation by   |
|          | Product Name: Santosh      | Postmaster on 10th Jan   |
|          |                            | 2026 alongwith Auto      |
|          | Surrender Value as on      | Surrender process gets   |
|          | Date-10 Jan 2026-54600     | triggered with Auto      |
|          |                            | Surrender request d      |
|          | Loan Principal-50000       | generation and Policy    |
|          |                            | status as PAS (Pending   |
|          | Loan Interest-2000         | Auto Surrender).         |
|          |                            |                          |
|          | Loan Capitalization Six    | If amount remains unpaid |
|          | monthly Cycle-10 Jan 2026  | within the period of 30  |
|          |                            | days post 30 days        |
|          | Loan Capitalization is     | period, Request will     |
|          | 105% of Surrender Value.   | move to workflow         |
|          |                            | mechanism.               |
|          | 10th Jan 2026 plus 30 days |                          |
|          |                            | Post workflow Policy     |
|          |                            | will move to TAS         |
|          |                            | (Terminated Auto         |
|          |                            | Surrender).              |
+----------+----------------------------+--------------------------+
| SR_FS-11 | Auto Surrender             | Amount remains unpaid    |
|          |                            | within the period of 30  |
|          | Policy No: TN-820948-CS    | days post 30 days        |
|          |                            | period.                  |
|          | Policy Holder Name: XXXXX  |                          |
|          |                            | Collection will not be   |
|          | Product Name: Santosh      | allowed at the DE/QC and |
|          |                            | Approver stage.          |
|          | Surrender Value as on      |                          |
|          | Date-10 Jan 2026-54600     |                          |
|          |                            |                          |
|          | Loan Principal-50000       |                          |
|          |                            |                          |
|          | Loan Interest-2000         |                          |
|          |                            |                          |
|          | Loan Capitalization Six    |                          |
|          | monthly Cycle-10 Jan 2026  |                          |
|          |                            |                          |
|          | Loan Capitalization is     |                          |
|          | 105% of Surrender Value.   |                          |
|          |                            |                          |
|          | 10th Jan 2026 plus 30 days |                          |
|          |                            |                          |
|          | Collection made post 30    |                          |
|          | days period.               |                          |
+----------+----------------------------+--------------------------+

## Forced Surrender Description and process flow

## 

![A screenshot of a computer screen AI-generated content may be
incorrect.](media/image1.png){width="6.205555555555556in"
height="3.890972222222222in"}

- Forced Surrender intimation letter should be sent to policy owner in
  the event of any three defaults (failure to pay) in the payment of
  half yearly interest.

The task should be created for workflow queue on 31^st^ day from of the
3rd unpaid loan interest due date and the policy should acquire new
status. (Currently PAS)

> Note: There could be multiple loan transactions on policy taken at
> different intervals; system should identify 3 instances for any
> particular loan transaction.

- If O/S interest is received within 30 days the Forced Surrender task
  should be completed and closed automatically and a letter should go to
  the policy holder.

- In the above situation collection batch should intimate the workflow
  on the payment received.

If O/S interest is not received within or equal to 30 days, System
should check Net surrender value. (i.e. Field: Net Amount)

- If Net Surrender value (Net Amount) is less than prescribed amount,
  then the Surrender Value is paid in Cash and policy should acquire
  "Terminated" status.

Since the Net Amount is more than or equal to prescribed amount, the
policy will be reduced paid up to sum assured \<Net Amount\>

- n the approval, new intimation letter should be sent to the policy
  holder.

  - On surrender screen would display the surrender details as mentioned
    on surrender screen.

  - Surrender request should be same as task created date. (Task created
    date is the date of the 3rd unpaid loan interest due date.)

  - Status transition:

> AP/IL/AL to PAS to TAS (As the case may be).

- If rejected, the policy should acquire status prior to PAS status
  (AP/IL/AL).

Next Forced surrender action on same policy should be initiated only at
next loan capitalisation.

## Forced Surrender Screens

### Approver

![A screenshot of a computer AI-generated content may be
incorrect.](media/image2.png){width="6.205555555555556in"
height="3.3201388888888888in"}

![A screenshot of a computer AI-generated content may be
incorrect.](media/image3.png){width="6.205555555555556in"
height="2.464575678040245in"}

![A blue and white rectangular object AI-generated content may be
incorrect.](media/image4.png){width="6.205555555555556in"
height="1.3423611111111111in"}

![A screenshot of a computer AI-generated content may be
incorrect.](media/image5.png){width="6.205555555555556in"
height="2.845833333333333in"}

List of documents section should not be displayed for these cases.

Approver can take decisions as given in above button.

1.  Approve - Approves, system should check the net surrender value is
    more than or equal to prescribed amount**should be configurable as
    desired by Business rule)**. If it is yes, the existing policy to be
    changed to "reduced paid up policy" with the net surrender value
    amount as the new Sum Assured. Policy status would be changed to AU.

If net surrender value is less than presecribed amount(**should be
configurable as desired by Business rule)**, surrender event should be
processed and policy status should be changed to Terminated auto
surrender from Pending Auto surrender.

2.  Once the approver approves, respective Forced Surrender letter
    should be sent as given in below section.

3.  View documents -- Click on view documents will be enable to view
    policy docket, any other documents received from customer.

4.  Add comments -- enable user to add comments if required

5.  Request history -- allow user to view the history of transaction

6.  Cancel -- enable to user to cancel the page and go back to inbox

- Request type should be marked as Forced Surrender. When approver
  clicks on the item, it automatically reserves to the user. This would
  take user to the surrender screen.

- In below screen -- Reason would be shown automatically as Loan
  interest default, Request date should be the date when workflow
  triggers the policy for Forced Surrender. This date should be allowed
  to change and surrender quote should be modified according to the
  request date.

- Disbursement method should be appearing in dropdown Cash or Cheque.

- If the amount to be refunded above 20,000 (if required), disbursement
  method automatically switched to Cheque.

- Approver can view the net surrender value through the surrender quote
  displayed on screen.

Remove the available loan on policy on the screen

+------+------------+------------+-----------+----------------+----------------+---------------+
| **Sr | **Field    | **Data     | **UI      | **Properties** | **Validation** | **Validation  |
| No** | Name**     | Type/      | Display** |                |                | message**     |
|      |            | Format**   |           |                |                |               |
+======+============+============+===========+================+================+===============+
| 1    | Surrender  | DD/MM/YYYY | Auto      | Data field     | It is          | NA            |
|      | Request    |            | display   | rule           | mandatory      |               |
|      | Date       |            | and not   |                | field.         |               |
|      |            |            | editable  |                |                |               |
|      |            |            |           |                | Should be auto |               |
|      |            |            |           |                | populated as   |               |
|      |            |            |           |                | the date of    |               |
|      |            |            |           |                | the 3rd unpaid |               |
|      |            |            |           |                | loan interest  |               |
|      |            |            |           |                | due date.      |               |
+------+------------+------------+-----------+----------------+----------------+---------------+
| 2    | Reason for | Text       | Auto      | Data field     | Loan interest  | NA            |
|      | surrender  |            | display,  | rule           | default should |               |
|      |            |            | not       |                | be displayed   |               |
|      |            |            | editable  |                | under reason   |               |
|      |            |            |           |                | for surrender  |               |
+------+------------+------------+-----------+----------------+----------------+---------------+
| 3    | Policy     | Text       | Auto      | Data field     | If assignment  | NA            |
|      | assignment |            | display.  | rule           | is yes,        |               |
|      | details    |            | Derived   |                | assignee       |               |
|      | fields     |            | from DB   |                | details --     |               |
|      |            |            |           |                | name of        |               |
|      |            |            |           |                | assignee,      |               |
|      |            |            |           |                | payee name and |               |
|      |            |            |           |                | payee address  |               |
|      |            |            |           |                | should be      |               |
|      |            |            |           |                | displayed      |               |
+------+------------+------------+-----------+----------------+----------------+---------------+
| 4    | Quote      | Numeric    | Auto      | Business Rule  | Net amount     | NA            |
|      |            |            | display   |                | should be      |               |
|      |            |            |           |                | round off.     |               |
|      |            |            | Derived   |                |                |               |
|      |            |            | from DB   |                |                |               |
+------+------------+------------+-----------+----------------+----------------+---------------+
| 5    | Net amount | Numeric    | Auto      | Data field     | Net surrender  | NA            |
|      |            |            | display   | rule           | value should   |               |
|      |            |            |           |                | be displayed   |               |
|      |            |            | Derived   |                |                |               |
+------+------------+------------+-----------+----------------+----------------+---------------+

# Letters

Below are the impacted letters:

### Letter Template Field Details & Rules -- Repayment Schedule Letter

### Auto Surrender Letters

Intimation of Forced Surrender due to Loan-1st and 2nd Intimation

- Letter Type

> New

- Letter Description

> Intimation of Forced Surrender due to Loan is the letter which will be
> triggered at the time of 1st and 2nd Intimations when the Loan
> Capitalization amount =\> 95 % and 98 % of Policy's Surrender value.

- Letter to be sent to:

> Policy holder

- Assumption:

> 'Principal and Interest amount', 'Policy Number' in the letter format
> is considered to be changeable for every letter.

- Letter Trigger Point:

> On 1st and 2nd Intimation

- Letter Format

> Word format as attached below:
>
> Sample letter : Format to be decided by PLI Dte
>
> ![](media/image6.emf)

Intimation of Forced Surrender-3rd Intimation

- Letter Type

> New

- Letter Description

> Intimation of Forced Surrender due to Loan-3rd Intimation is the
> letter which will be triggered at the time of 3rd Intimation when the
> Loan Capitalization amount =\> 100 % of Policy's Surrender value.

- Letter to be sent to:

> Policy holder

- Assumption:

> 'Principal and Interest amount', 'Policy Number' in the letter format
> is considered to be changeable for every letter and 30 days period
> will remain static as shared in the template.

- Letter Trigger Point:

> On 3rd Intimation

- Letter Format To be prescribed by PLI Dte

Forced Surrender Letter

- Letter Type

> New

- Letter Description

> Intimation of Forced Surrender letter when the Loan Capitalization
> amount =\> 100 % of Policy's Surrender value and the Forced surrender
> request is approved by the Postmaster Approver.

- Letter to be sent to:

> Policy holder

- Assumption:

> 'Principal and Interest amount', 'Policy Number' in the letter format
> is considered to be changeable for every letter.

- Letter Trigger Point:

> On Approval of the Forced Surrender request

- Letter format To be prescribed by PLI Dte

> **SMS's**

Formats of SMS (To be decided by Business team) sample template given
below)

## Event Name: Auto Surrender 1st and 2nd Intimation

> This SMS will be [sent]{.mark} at the time of 1st and 2nd Intimations
> when the Loan Capitalization amount =\> 95 % and 98 % of Policy's
> Surrender value.

Alert Periodicity: At the trigger of 1st and 2nd intimation

Alert Message:

To Customer (Policyholder) as an Auto Surrender Intimation: "[Dear
Customer, due to non-repayment of loan interest, your policy no
123456789012345 will be liable for auto-surrender. Please make payment
immediately."]{.mark}

Data Elements:

  -------------------------------------
  Field Name           Data Type/
                       Format
  -------------------- ----------------
  Policy Number        String

  -------------------------------------

## Event Name: Auto Surrender 3rd (Final) Intimation

> This SMS will be [sent]{.mark} at the time of 3rd (Final) Intimation
> when the Loan Capitalization amount =\> 100% of Policy's Surrender
> value.

Alert Periodicity: At the trigger of 3rd intimation

Alert Message:

To Customer (Policyholder) as an Auto Surrender Intimation: "[Dear
Customer, if loan due amount \<total amount\> is not paid in 30 days,
your policy will be processed for auto surrender"]{.mark}.

Data Elements:

  -------------------------------------
  Field Name           Data Type/
                       Format
  -------------------- ----------------
  Policy Number        String

  -------------------------------------

## Event Name: Auto Surrender

> This SMS will be [sent]{.mark} at the time of Approval of Auto
> Surrender request..

Alert Periodicity: At the Approval of Auto Surrender request.

Alert Message:

To Customer (Policyholder) for Policy getting Automatically Surrendered:
"[Dear Customer, due to non-repayment of loan interest, your policy no
123456789012345 has been Forced surrendered]{.mark}".

# Reports

To be decided by Business requirement

![](media/image7.emf)Letter for Intimation of force surrender due to
Loan
