**DEPARTMENT OF POSTS**

**MINISTRY OF COMMUNICATIONS & IT**

**GOVERNMENT OF INDIA**

**System Requirement Specification (SRS)**

**Installment Revival**

# Table of Contents {#table-of-contents .TOC-Heading}

[1. Overview [4](#overview)](#overview)

[1.1 Brief Description [4](#brief-description)](#brief-description)

[1.2 Users of this Document [4](#_Toc522193762)](#_Toc522193762)

[1.3 Acronym/Abbreviation
[4](#acronymabbreviation)](#acronymabbreviation)

[2. System Flows [4](#_Toc521516511)](#_Toc521516511)

[2.1 Standard Flow [4](#standard-flow)](#standard-flow)

[2.2 Sub-Standard Flows [4](#_Toc472433580)](#_Toc472433580)

[3. Assumptions [6](#_Toc469677935)](#_Toc469677935)

[4. Scope [6](#_Toc522193768)](#_Toc522193768)

[4.1 In scope [6](#_Toc522193769)](#_Toc522193769)

[4.2 Out of scope [6](#_Toc522193770)](#_Toc522193770)

[5. Business Rules and Logics
[9](#business-rules-and-logics)](#business-rules-and-logics)

[6. Illustration and Understanding Document
[18](#illustration-and-understanding-scenarios)](#illustration-and-understanding-scenarios)

[7. Screens and Navigation
[48](#screens-and-navigation)](#screens-and-navigation)

[8. Events [52](#_Toc522193774)](#_Toc522193774)

[8.1 Changes in Policy Status [52](#_Toc522193775)](#_Toc522193775)

[9. Accounting [52](#_Toc522193776)](#_Toc522193776)

[10. Reference [52](#_Toc522193777)](#_Toc522193777)

[11. Documents/Letters [53](#documentsletters)](#documentsletters)

[12. Application /Functionality Impact
[53](#_Toc522193779)](#_Toc522193779)

[13. Appendix [53](#_Toc522193780)](#_Toc522193780)

[14. Issues / Constraints / Risk [53](#_Toc522193781)](#_Toc522193781)

# Overview

## Brief Description {#brief-description .ABC}

> This document is prepared to detail the functionality of instalment
> revival and its impact on other functionalities.

## Acronym/Abbreviation

  ----------------------- ---------------------------------------------
  **Acronym/Term**        **Abbreviation /Description**

  PLI                     Postal Life Insurance

  RPLI                    Rural Postal Life Insurance

  PO                      Post Office

  DoP PLI                 Department of Posts, Postal Life Insurance

  PAS                     Policy Administrative System

  IL                      Inactive Lapse

  AL                      Active Lapse

  PM                      Pending Maturity
  ----------------------- ---------------------------------------------

[]{#_Toc521516511 .anchor}

# Existing System Flows

## Standard Flow

> Indexing of Revival Request → Data Entry → Quality Checker → Approver
> → Letter Generation (Instalment revival acceptance) → Payment of first
> Installment amount → Letter Generation (Revival memo) → Policy Status
> changed to AP → Request status changed to approved
>
> [Payment of subsequent instalments]{.underline}
>
> Go to Collection module →Select Transaction Type as "Instalment
> Revival" → Enter Policy Number→ From date Auto Populated → To Date is
> Auto Populated → Enter Actual Amount Paid → Enter Payment Mode →
> Submit→ Receipt Number is generated→ Print receipt
>
> [Advance payment of instalments]{.underline}
>
> []{#_Toc472433580 .anchor}Go to Collection module →Select Transaction
> Type as "Instalment Revival" → Enter Policy Number→ From date Auto
> Populated → To Date is Auto Populated → Enter Amount Paid (can be more
> than required amount) → Enter Payment Mode → Submit→ Receipt Number is
> generated→ Print receipt

## Sub-Standard Flows

> [Rejected by Approver]{.underline}
>
> Indexing of Revival Request → Data Entry → Quality Checker → Approver
> rejected the request → Letter Generation (Reinstatement rejection
> letter) → Policy Status remains AL → Request status changed to
> Rejected.
>
> [First instalment not paid within 60 days of approval]{.underline}
>
> Indexing of Revival Request → Data Entry → Quality Checker → Approver
> → Letter Generation (Instalment revival acceptance) → First
> Installment amount not paid → Policy Status remains AL → Revival
> installment amount re-calculated

[Default of second instalment]{.underline}

> Indexing of Revival Request → Data Entry → Quality Checker → Approver
> → Letter Generation (Instalment revival acceptance) → payment of first
> Installment amount → Policy Status changed to AP → Letter Generation
> (Revival memo) → Request status changed to Approved → Default of 2^nd^
> installment amount → Policy status changed to Lapsed.
>
> [Default of Premium Amount when all the instalments are not
> paid]{.underline}
>
> Indexing of Revival Request → Data Entry → Quality Checker → Approver
> → Letter Generation (Instalment revival acceptance) → payment of first
> Installment amount → Policy Status changed to AP → Letter Generation
> (Reinstatement rejection letter) → Request status changed to Approved
> → Default of Premium amount → Policy status changed to Lapsed.
>
> [Redirected by Approver]{.underline}
>
> Indexing of Revival Request → Data Entry → Quality Checker → Approver
> redirected the request → Data Entry → Quality Checker → Approver →
> Letter Generation (Instalment revival acceptance) → Payment of first
> Installment amount → Letter Generation (Revival memo) → Policy Status
> changed to AP → Request status changed to approved
>
> [Missing document by Data Entry]{.underline}
>
> Indexing of Revival Request → Data Entry → DE raises missing document
> → Missing document received → DE → QC→ Approver → Letter Generation
> (Instalment revival acceptance) → Payment of first Installment amount
> → Letter Generation (Revival memo) → Policy Status changed to AP →
> Request status changed to approved
>
> [Missing document by Quality Checker]{.underline}
>
> Indexing of Revival Request → Data Entry → Quality Checker raises
> missing document → Missing document received → Quality checker →
> Approver → Letter Generation (Instalment revival acceptance) → Payment
> of first Installment amount → Letter Generation (Revival memo) →
> Policy Status changed to AP → Request status changed to approved
>
> [Missing document by Approver]{.underline}
>
> Indexing of Revival Request → Data Entry → Quality Checker → Approver
> raises missing document → missing document received → Approver →
> Letter Generation (Instalment revival acceptance) → Payment of first
> Installment amount → Letter Generation (Revival memo) → Policy Status
> changed to AP → Request status changed to approved
>
> [Withdrawal of Request]{.underline}
>
> Indexing of Revival Request → Data Entry → Quality Checker → Approver
> → withdrawal of request (after approval but before collection) →
> request status changed to withdrawn.[]{#_Toc469677935 .anchor}

# Business Rules and Logics

+----------+------------------------+-----------------------------+----------------------------------+
| Business | Rule                   | Description                 | Logic                            |
| Rule No  |                        |                             |                                  |
+==========+========================+=============================+==================================+
| IR_1     | Product Eligibility    | Product eligible for        | Installment revival is allowed   |
|          |                        | Installment Revival         | on all the products.             |
|          |                        | request.                    |                                  |
+----------+------------------------+-----------------------------+----------------------------------+
| IR_2     | Policy Status          | Policy status at the time   | Installment Revival can be       |
|          |                        | of indexing the request     | indexed only on policies with AL |
|          |                        |                             | status.                          |
+----------+------------------------+-----------------------------+----------------------------------+
| IR_3     | Minimum number of      | Minimum number of           | Minimum number of instalments    |
|          | instalments            | instalments                 | are 2.                           |
+----------+------------------------+-----------------------------+----------------------------------+
| IR_4     | Maximum Number of      | Maximum Number of           | - Maximum number of installments |
|          | Installments           | Installments                |   are 12.                        |
|          |                        |                             |                                  |
|          |                        |                             | - Error will populate on Revival |
|          |                        |                             |   processing screen in case due  |
|          |                        |                             |   date of any instalment falls   |
|          |                        |                             |   after or in the month of       |
|          |                        |                             |   maturity.                      |
|          |                        |                             |                                  |
|          |                        |                             | > *Error: "Due date of           |
|          |                        |                             | > instalment is in or after      |
|          |                        |                             | > maturity month. Please index   |
|          |                        |                             | > new request".*                 |
+----------+------------------------+-----------------------------+----------------------------------+
| IR_5     | Calculations           | Formula                     | Revival calculation using        |
|          |                        |                             | formula: = \[{(1+Interest) \^    |
|          |                        |                             | No. of unpaid premium months) -- |
|          |                        |                             | 1} X 101\] X Monthly Premium     |
|          |                        |                             |                                  |
|          |                        |                             | Revival interest to be           |
|          |                        |                             | calculated up to current months  |
|          |                        |                             | (Date of Indexing) -- 1 month.   |
|          |                        |                             |                                  |
|          |                        |                             | Unpaid premiums to be calculated |
|          |                        |                             | up to frequency before the       |
|          |                        |                             | current billing frequency paid   |
|          |                        |                             | to date. [(monthly               |
|          |                        |                             | frequency)]{.mark}               |
|          |                        |                             |                                  |
|          |                        |                             | [Unpaid premiums to be           |
|          |                        |                             | calculated up to the current     |
|          |                        |                             | billing frequency paid to date.  |
|          |                        |                             | (other than monthly              |
|          |                        |                             | frequency)]{.mark}               |
|          |                        |                             |                                  |
|          |                        |                             | [Example:]{.underline}           |
|          |                        |                             |                                  |
|          |                        |                             | Paid to Date = 31 March 2018     |
|          |                        |                             |                                  |
|          |                        |                             | Revival Indexing Done on 15 Dec  |
|          |                        |                             | 2018                             |
|          |                        |                             |                                  |
|          |                        |                             | Premium Frequency = Quarterly    |
|          |                        |                             |                                  |
|          |                        |                             | Unpaid premiums will be          |
|          |                        |                             | calculated from 1 April 2018     |
|          |                        |                             | till Dec [2018 (09               |
|          |                        |                             | months)]{.mark}                  |
|          |                        |                             |                                  |
|          |                        |                             | For a policy with frequency      |
|          |                        |                             | other than monthly, Monthly      |
|          |                        |                             | premium as required for the      |
|          |                        |                             | above formula is to be           |
|          |                        |                             | determined by a dividing Total   |
|          |                        |                             | Unpaid Premium amount with       |
|          |                        |                             | number of months in current      |
|          |                        |                             | frequency.                       |
|          |                        |                             |                                  |
|          |                        |                             | Instalment calculation using     |
|          |                        |                             | formula = (Revival Amount X      |
|          |                        |                             | Interest) X \[{(1+Interest) \^   |
|          |                        |                             | No. of Installments opted} /     |
|          |                        |                             | {(1+Interest)\^ No. of           |
|          |                        |                             | Installments opted -1}\]         |
|          |                        |                             |                                  |
|          |                        |                             | **[First Due Amount after        |
|          |                        |                             | Approval of Revival              |
|          |                        |                             | request]{.underline}**           |
|          |                        |                             |                                  |
|          |                        |                             | [In case of monthly              |
|          |                        |                             | premium:]{.underline}            |
|          |                        |                             |                                  |
|          |                        |                             | Instalment Amount + Taxes on     |
|          |                        |                             | total unpaid premiums + Current  |
|          |                        |                             | month premium + taxes on modal   |
|          |                        |                             | premium                          |
+----------+------------------------+-----------------------------+----------------------------------+
| IR_6     | Tax calculation        | Tax calculation and         | Tax on Total unpaid premiums     |
|          |                        | collection                  | will be paid at the time of      |
|          |                        |                             | first instalment.                |
|          |                        |                             |                                  |
|          |                        |                             | Tax Rate at the time of          |
|          |                        |                             | Collection will be considered    |
|          |                        |                             | for tax calculation (either      |
|          |                        |                             | service tax/GST).                |
|          |                        |                             |                                  |
|          |                        |                             | Office code where transaction is |
|          |                        |                             | happening and that CGST, SGST or |
|          |                        |                             | UTGST rate will be applicable.   |
|          |                        |                             | This office code will be of the  |
|          |                        |                             | user office code who is login in |
|          |                        |                             | the application.                 |
|          |                        |                             |                                  |
|          |                        |                             | No taxes will be paid on the     |
|          |                        |                             | subsequent instalment.           |
+----------+------------------------+-----------------------------+----------------------------------+
| IR_7     | Mode of Instalment     | Mode of installment         | - Instalments can be paid by     |
|          | Collection             | collection                  |   cash, cheque, online           |
|          |                        |                             |                                  |
|          |                        |                             | - Collection via Meghdoot/bulk   |
|          |                        |                             |   upload is not allowed.         |
|          |                        |                             |                                  |
|          |                        |                             | > Error "Instalment payment not  |
|          |                        |                             | > allowed through Meghdoot/bulk  |
|          |                        |                             | > upload" will be displayed on   |
|          |                        |                             | > the bulk upload screen on the  |
|          |                        |                             | > click of the upload button.    |
|          |                        |                             | >                                |
|          |                        |                             | > The same error will be         |
|          |                        |                             | > displayed in the csv file.     |
|          |                        |                             |                                  |
|          |                        |                             | - Part payment of installment    |
|          |                        |                             |   amount is not allowed.         |
|          |                        |                             |                                  |
|          |                        |                             | - If any instalment is paid by   |
|          |                        |                             |   cheque and the cheque gets     |
|          |                        |                             |   dishonored, then the policy    |
|          |                        |                             |   status will be changed back to |
|          |                        |                             |   Lapse and any amount paid as a |
|          |                        |                             |   part of Revival will be sent   |
|          |                        |                             |   to suspense for adjustment on  |
|          |                        |                             |   subsequent revival or          |
|          |                        |                             |   settlement of claim whichever  |
|          |                        |                             |   is earlier.                    |
|          |                        |                             |                                  |
|          |                        |                             | - If any instalment is paid via  |
|          |                        |                             |   cheque and cheque is not       |
|          |                        |                             |   cleared until next due date,   |
|          |                        |                             |   then next instalment is not    |
|          |                        |                             |   allowed and policy will lapse  |
|          |                        |                             |   and paid to date will be as    |
|          |                        |                             |   before revival request.        |
+----------+------------------------+-----------------------------+----------------------------------+
| IR_8     | Installment Frequency  | Instalment Frequency        | - Installment revival will be    |
|          |                        |                             |   only in monthly installments   |
|          |                        |                             |   irrespective of whether the    |
|          |                        |                             |   policy has the premium         |
|          |                        |                             |   frequency monthly, quarterly,  |
|          |                        |                             |   half yearly or yearly.         |
+----------+------------------------+-----------------------------+----------------------------------+
| IR_9     | Grace period           | Grace period for            | - There is no grace period for   |
|          |                        | installment revival         |   making installment revival     |
|          |                        |                             |   payments. If installment is    |
|          |                        |                             |   not paid on the due date then  |
|          |                        |                             |   policy will lapse.             |
|          |                        |                             |                                  |
|          |                        |                             | - For renewal collection there   |
|          |                        |                             |   will be no change. It will be  |
|          |                        |                             |   as existing.                   |
+----------+------------------------+-----------------------------+----------------------------------+
| IR_10    | SLA for First          | SLA for First Instalment    | First Instalment need to be paid |
|          | Instalment             |                             | within 60 days from the date of  |
|          |                        |                             | Approval of the request. If it   |
|          |                        |                             | is not paid then:                |
|          |                        |                             |                                  |
|          |                        |                             | - Policy status remains AL.      |
|          |                        |                             |                                  |
|          |                        |                             | - Installment Revival amount     |
|          |                        |                             |   will be re-calculated.         |
|          |                        |                             |                                  |
|          |                        |                             | - Paid to date will not change.  |
+----------+------------------------+-----------------------------+----------------------------------+
| IR_11    | Due date of            |                             | Due date of the 2nd and          |
|          | instalments            |                             | subsequent instalment will be    |
|          |                        |                             | done on the 1st day of the month |
|          |                        |                             | following the month in which     |
|          |                        |                             | first instalment was paid.       |
|          |                        |                             |                                  |
|          |                        |                             | [Example:]{.underline}           |
|          |                        |                             |                                  |
|          |                        |                             | Policyholder opted for 3         |
|          |                        |                             | instalments                      |
|          |                        |                             |                                  |
|          |                        |                             | First instalment collected on 25 |
|          |                        |                             | Jan 2018                         |
|          |                        |                             |                                  |
|          |                        |                             | Due date for 2^nd^ instalment    |
|          |                        |                             | will be 1 Feb 2018.              |
|          |                        |                             |                                  |
|          |                        |                             | Due date for 3^rd^ instalment    |
|          |                        |                             | will be 1 March 2018.            |
|          |                        |                             |                                  |
|          |                        |                             | In case of system failure if     |
|          |                        |                             | customer is unable to make       |
|          |                        |                             | payment of dues on the due date, |
|          |                        |                             | then customer is allowed to make |
|          |                        |                             | pending installments once issue  |
|          |                        |                             | is resolved.                     |
+----------+------------------------+-----------------------------+----------------------------------+
| IR_12    | Payment of Premium     | Payment of Premium          | Premiums will be paid on every   |
|          |                        |                             | due date as per billing          |
|          |                        |                             | frequency.                       |
+----------+------------------------+-----------------------------+----------------------------------+
| IR_13    | First Instalment       | Change in Policy status     | When customer makes first        |
|          | payment                | after collection -- First   | Payment:                         |
|          |                        | Instalment                  |                                  |
|          |                        |                             | - Policy status will change to   |
|          |                        |                             |   AP                             |
|          |                        |                             |                                  |
|          |                        |                             | - Paid to date will be updated   |
|          |                        |                             |   as per current billing         |
|          |                        |                             |   frequency paid to date.        |
+----------+------------------------+-----------------------------+----------------------------------+
| IR_14    | Advance Premiums       | Advance Premiums            | Customer is allowed to make      |
|          |                        |                             | advance premium payments.        |
|          |                        |                             |                                  |
|          |                        |                             | Customer will have to change the |
|          |                        |                             | 'From Date' in accordance to the |
|          |                        |                             | advance collection.              |
+----------+------------------------+-----------------------------+----------------------------------+
| IR_15    | Advance Installments   | Advance Installments        | - Customer has an option to make |
|          |                        |                             |   advance instalment payment.    |
|          |                        |                             |   Amount paid can be more than   |
|          |                        |                             |   total instalments.             |
|          |                        |                             |                                  |
|          |                        |                             | - If amount paid is sufficient   |
|          |                        |                             |   to accommodate certain number  |
|          |                        |                             |   of instalments then user has   |
|          |                        |                             |   an option to change the 'To    |
|          |                        |                             |   Date'.                         |
|          |                        |                             |                                  |
|          |                        |                             | - Any amount paid that is not    |
|          |                        |                             |   sufficient to accommodate any  |
|          |                        |                             |   instalment will be kept in     |
|          |                        |                             |   suspense and can be adjusted   |
|          |                        |                             |   against future instalments.    |
|          |                        |                             |                                  |
|          |                        |                             | - Total amount paid cannot be    |
|          |                        |                             |   more than all the instalments. |
|          |                        |                             |   If user pays more than total   |
|          |                        |                             |   instalments error "Amount more |
|          |                        |                             |   than total instalments" will   |
|          |                        |                             |   be displayed on the collection |
|          |                        |                             |   screen.                        |
+----------+------------------------+-----------------------------+----------------------------------+
| IR_16    | Default of subsequent  | If any of the installment   | If customer defaults any         |
|          | Installment            | after 1st installment is    | subsequent installment amount,   |
|          |                        | not collected.              | then:                            |
|          |                        |                             |                                  |
|          |                        |                             | - Policy status changes back to  |
|          |                        |                             |   AL.                            |
|          |                        |                             |                                  |
|          |                        |                             | - Any Instalment or Premium      |
|          |                        |                             |   collected after approval of    |
|          |                        |                             |   instalment revival request     |
|          |                        |                             |   will move to suspense.         |
|          |                        |                             |                                  |
|          |                        |                             | - Tax component will not move to |
|          |                        |                             |   suspense.                      |
|          |                        |                             |                                  |
|          |                        |                             | - Paid to date will as of before |
|          |                        |                             |   Revival.                       |
+----------+------------------------+-----------------------------+----------------------------------+
| IR_17    | Existing request       | Already existing            | New Installment revival request  |
|          |                        | Installment revival request | cannot be indexed when there is  |
|          |                        |                             | pending installment revival      |
|          |                        |                             | request or all the instalments   |
|          |                        |                             | are not paid.                    |
|          |                        |                             |                                  |
|          |                        |                             | Error will be displayed "Revival |
|          |                        |                             | request already pending"         |
+----------+------------------------+-----------------------------+----------------------------------+
| IR_18    | Death Claim/Maturity   | If /Death Claim/Maturity    | - Death/Maturity Claim request   |
|          | Claim                  | Claim request is indexed.   |   is allowed to be indexed even  |
|          |                        |                             |   if Revival request is in       |
|          |                        |                             |   progress and not yet approved. |
|          |                        |                             |   In this case instalment        |
|          |                        |                             |   revival request will be auto   |
|          |                        |                             |   terminated.                    |
|          |                        |                             |                                  |
|          |                        |                             | - [In the event of death of the  |
|          |                        |                             |   insurant, who has been         |
|          |                        |                             |   depositing the installment of  |
|          |                        |                             |   arrears as directed by the     |
|          |                        |                             |   Postmaster General/ Head of    |
|          |                        |                             |   Division besides the normal    |
|          |                        |                             |   monthly premia regularly as    |
|          |                        |                             |   and when due notwithstanding   |
|          |                        |                             |   the fact that some arrears of  |
|          |                        |                             |   premia remain unpaid at the    |
|          |                        |                             |   time of death, the claim       |
|          |                        |                             |   against the said policy shall  |
|          |                        |                             |   be accepted subject to the     |
|          |                        |                             |   deduction of such arrears of   |
|          |                        |                             |   premia and interest thereon    |
|          |                        |                             |   besides loan amount and        |
|          |                        |                             |   interest thereon, if any, from |
|          |                        |                             |   the claim amount.]{.mark}      |
|          |                        |                             |                                  |
|          |                        |                             | - Death/Maturity Claim request   |
|          |                        |                             |   is allowed to be indexed even  |
|          |                        |                             |   if all the instalments are not |
|          |                        |                             |   paid and claim amount will be  |
|          |                        |                             |   paid based on AL policy        |
|          |                        |                             |   status.                        |
|          |                        |                             |                                  |
|          |                        |                             | Any amount collected after       |
|          |                        |                             | approval of installment revival  |
|          |                        |                             | request will move to suspense    |
|          |                        |                             | and will be treated as excess    |
|          |                        |                             | amount while disbursing the      |
|          |                        |                             | claim amount.                    |
+----------+------------------------+-----------------------------+----------------------------------+
| IR_19    | Conversion/Commutation | Conversion/Commutation      | - Conversion/Commutation request |
|          | request                | request                     |   is not allowed to be           |
|          |                        | indexed/Approval/Effective. |   indexed/Approved and Effective |
|          |                        |                             |   until all the instalments are  |
|          |                        |                             |   paid.                          |
|          |                        |                             |                                  |
|          |                        |                             | - Conversion/Commutation cannot  |
|          |                        |                             |   be indexed/Approved and        |
|          |                        |                             |   Effective if instalment        |
|          |                        |                             |   revival request is in progress |
|          |                        |                             |   and not yet approved.          |
+----------+------------------------+-----------------------------+----------------------------------+
| IR_20    | Billing Method change  | Billing Method change       | - Billing method change is       |
|          |                        |                             |   allowed to change irrespective |
|          |                        |                             |   of all the instalments paid or |
|          |                        |                             |   not. Only exception to this is |
|          |                        |                             |   Cash to pay recovery or vice   |
|          |                        |                             |   versa                          |
+----------+------------------------+-----------------------------+----------------------------------+
| IR_21    | Survival Claim         | Survival Claim payment      | - Survival Claim payment should  |
|          |                        |                             |   not be allowed to be paid if   |
|          |                        |                             |   Instalment revival is not      |
|          |                        |                             |   completed and all installment  |
|          |                        |                             |   are not paid.                  |
|          |                        |                             |                                  |
|          |                        |                             | - Once all the instalments are   |
|          |                        |                             |   paid then survival claim is    |
|          |                        |                             |   allowed to be indexed for the  |
|          |                        |                             |   previous due.                  |
+----------+------------------------+-----------------------------+----------------------------------+
| IR_22    | Loan                   | Loan Request Indexing       | - Loan request                   |
|          |                        |                             |   Indexing/Approval/Disbursement |
|          |                        |                             |   will not be allowed until all  |
|          |                        |                             |   the instalments are paid.      |
|          |                        |                             |                                  |
|          |                        |                             | - Loan repayments are allowed    |
|          |                        |                             |   even if all the revival        |
|          |                        |                             |   instalments are paid or not.   |
+----------+------------------------+-----------------------------+----------------------------------+
| IR_23    | Policy                 | Policy Cancellation/Free    | - If Instalment revival request  |
|          | Cancellation/Free look | Look Cancellation           |   is in progress and not yet     |
|          | cancellation           |                             |   approved, Policy Cancellation  |
|          |                        |                             |   will be allowed and instalment |
|          |                        |                             |   revival request will           |
|          |                        |                             |   terminate.                     |
|          |                        |                             |                                  |
|          |                        |                             | - If all the instalments are not |
|          |                        |                             |   paid and Policy Cancellation,  |
|          |                        |                             |   then policy cancellation is    |
|          |                        |                             |   allowed, policy status will    |
|          |                        |                             |   change to AL and all the       |
|          |                        |                             |   instalments and premium        |
|          |                        |                             |   collected after the approval   |
|          |                        |                             |   of instalment revival request  |
|          |                        |                             |   will move to suspense. This    |
|          |                        |                             |   amount will be disbursed with  |
|          |                        |                             |   the claim amount.              |
+----------+------------------------+-----------------------------+----------------------------------+
| IR_24    | Suspense               | Suspense Account            | - Any advance instalment amount  |
|          |                        |                             |   will be kept in policy         |
|          |                        |                             |   suspense and should be         |
|          |                        |                             |   identified as 'IR' in order to |
|          |                        |                             |   signify extra amount collected |
|          |                        |                             |   against instalment.            |
|          |                        |                             |                                  |
|          |                        |                             | - Suspense entry should also be  |
|          |                        |                             |   created for interest component |
|          |                        |                             |   included in instalment amount. |
|          |                        |                             |                                  |
|          |                        |                             | - Any advance instalment paid    |
|          |                        |                             |   will be adjusted only against  |
|          |                        |                             |   instalments.                   |
+----------+------------------------+-----------------------------+----------------------------------+
| IR_25    | Letters                | Same as Existing Instalment | [In case of Approval of          |
|          |                        | Revival Letters             | request:]{.underline}            |
|          |                        |                             |                                  |
|          |                        |                             | - Instalment revival acceptance  |
|          |                        |                             |   letter                         |
|          |                        |                             |                                  |
|          |                        |                             | - Revival memo                   |
|          |                        |                             |                                  |
|          |                        |                             | [In case of Rejection of         |
|          |                        |                             | request:]{.underline}            |
|          |                        |                             |                                  |
|          |                        |                             | - Reinstatement Rejection Letter |
+----------+------------------------+-----------------------------+----------------------------------+
| IR_26    | Rebate                 | Rebate on Premium and       | - Rebate is given only in case   |
|          |                        | Instalment                  |   of advance Premium collection. |
|          |                        |                             |   Rebate rule is as existing.    |
|          |                        |                             |                                  |
|          |                        |                             | - Rebate is not applicable for   |
|          |                        |                             |   instalment amount.             |
+----------+------------------------+-----------------------------+----------------------------------+
| IR_27    | Receipt Cancellation   | Instalment Revival receipt  | - If receipt for subsequent      |
|          |                        | cancellation                |   instalment amount is           |
|          |                        |                             |   cancelled, Policy status will  |
|          |                        |                             |   remain AL.                     |
|          |                        |                             |                                  |
|          |                        |                             | - In case of first instalment:   |
|          |                        |                             |                                  |
|          |                        |                             | <!-- -->                         |
|          |                        |                             |                                  |
|          |                        |                             | - [If collection is done and     |
|          |                        |                             |   request is completed and in    |
|          |                        |                             |   future any receipt with        |
|          |                        |                             |   respect to that request is     |
|          |                        |                             |   canceled, it will not have any |
|          |                        |                             |   impact on the request and no   |
|          |                        |                             |   SLA will be applicable. No     |
|          |                        |                             |   collection of first instalment |
|          |                        |                             |   will be allowed w.r.t that     |
|          |                        |                             |   request.]{.mark}               |
+----------+------------------------+-----------------------------+----------------------------------+
| IR_28    | Suspense Reversal      | Suspense Reversal           | - Suspense reversal for the      |
|          |                        |                             |   first collection should be     |
|          |                        |                             |   restricted. Error "Suspense    |
|          |                        |                             |   reversal not allowed for first |
|          |                        |                             |   collection".                   |
+----------+------------------------+-----------------------------+----------------------------------+
| IR_29    | Maximum Number of      | Maximum Number of Revivals  | [Maximum number of revivals can  |
|          | Revivals               |                             | be made Configurable.]{.mark}    |
|          |                        |                             | Policyholder can revive the      |
|          |                        |                             | policy at any number of times.   |
+----------+------------------------+-----------------------------+----------------------------------+
| IR_30    | Receipt                | Receipt generated for       | Receipt will be generated for    |
|          |                        | instalment revival          | instalments paid. Receipt format |
|          |                        |                             | will be shared. [Duplicate       |
|          |                        |                             | receipt will be generated as per |
|          |                        |                             | the existing rules.]{.mark}      |
+----------+------------------------+-----------------------------+----------------------------------+
| IR_31    | SMS/Email              | SMS/Email                   | Template for SMS and Email will  |
|          |                        |                             | be covered under SMS/Email SRS.  |
+----------+------------------------+-----------------------------+----------------------------------+
| IR_32    | Non-Financial          | Non-Financial               | Non-Financial request will take  |
|          |                        |                             | priority over Instalment         |
|          |                        |                             | Revival.                         |
|          |                        |                             |                                  |
|          |                        |                             | Example: If there is address     |
|          |                        |                             | change request, then it will be  |
|          |                        |                             | approved irrespective of         |
|          |                        |                             | instalment revival request is    |
|          |                        |                             | approved or not or all           |
|          |                        |                             | instalments are paid or not.     |
+----------+------------------------+-----------------------------+----------------------------------+
| IR_33    | Reports                | Impact on reports           | Format of the reports will be    |
|          |                        |                             | covered under Reports SRS        |
+----------+------------------------+-----------------------------+----------------------------------+
| IR_34    | Billing Frequency      | Billing frequency Change    | User will not be allowed to      |
|          | Change                 |                             | change billing frequency if      |
|          |                        |                             |                                  |
|          |                        |                             | - Instalment revival Request is  |
|          |                        |                             |   pending for approval           |
+----------+------------------------+-----------------------------+----------------------------------+
| IR_35    | Suspense Transfer      | Suspense transfer of        | Suspense transfer of instalment  |
|          |                        | instalment suspense amount  | suspense amount will not be      |
|          |                        |                             | allowed.                         |
+----------+------------------------+-----------------------------+----------------------------------+
| IR_36    | Collection             | Collection screen           | First collection amount consist  |
|          |                        |                             | of Premium + Instalment amount + |
|          |                        |                             | Taxes on unpaid Premium + Taxes  |
|          |                        |                             | on Premium.                      |
|          |                        |                             |                                  |
|          |                        |                             | Premium + Taxes on premium to be |
|          |                        |                             | collected through 'Renewal'      |
|          |                        |                             | screen.                          |
|          |                        |                             |                                  |
|          |                        |                             | Instalment + taxes on unpaid     |
|          |                        |                             | premiums to be collected through |
|          |                        |                             | 'Instalment Revival' screen.     |
|          |                        |                             |                                  |
|          |                        |                             | Only one transaction for Renewal |
|          |                        |                             | and Instalment can be processed  |
|          |                        |                             | until Revival request collection |
|          |                        |                             | stage is completed.              |
|          |                        |                             |                                  |
|          |                        |                             | In case user tries to execute    |
|          |                        |                             | multiple transaction for         |
|          |                        |                             | Renewal/Instalments, error       |
|          |                        |                             | "Revival request not completed"  |
|          |                        |                             | will be thrown.                  |
|          |                        |                             |                                  |
|          |                        |                             | Collection stage should be       |
|          |                        |                             | marked as completed when both    |
|          |                        |                             | first instalment amount and      |
|          |                        |                             | required premium + taxes are     |
|          |                        |                             | collected.                       |
|          |                        |                             |                                  |
|          |                        |                             | User should be given a message   |
|          |                        |                             | on collection screen that        |
|          |                        |                             | instalment is to be paid to      |
|          |                        |                             | avoid policy lapsation at the    |
|          |                        |                             | time of renewal collection, vice |
|          |                        |                             | versa, at the time of instalment |
|          |                        |                             | collection, a message should be  |
|          |                        |                             | given to collect renewal premium |
|          |                        |                             | for current month as both        |
|          |                        |                             | instalment and monthly renewal   |
|          |                        |                             | premium are mandatory to avoid   |
|          |                        |                             | policy lapsation.                |
+----------+------------------------+-----------------------------+----------------------------------+
| IR_37    | Withdrawal of          | Withdrawal of Instalment    | Instalment revival request can   |
|          | Instalment Revival     | Revival request             | be withdrawn before collection   |
|          | request                |                             | of first instalment + required   |
|          |                        |                             | premium.                         |
|          |                        |                             |                                  |
|          |                        |                             | Otherwise error "Revival request |
|          |                        |                             | cannot be withdrawn as           |
|          |                        |                             | collection is ~~not~~ done" will |
|          |                        |                             | be thrown.                       |
|          |                        |                             |                                  |
|          |                        |                             | Request status will change to    |
|          |                        |                             | 'withdrawn'                      |
+----------+------------------------+-----------------------------+----------------------------------+

# Illustration and Understanding Scenarios

+----------------+---------------------------------------+--------------------------------------------------------------+
| **Business     | **Illustration as per Policy Life     | **Expected System Response**                                 |
| Rule - No**    | Cycle**                               |                                                              |
+:===============+=======================================+:=============================================================+
| IR_1           | **[Product Eligibility]{.underline}** | Instalment Revival request is allowed to be indexed.         |
|                |                                       |                                                              |
|                | Policy number XXXX00011,              |                                                              |
|                |                                       |                                                              |
|                | Policy issuance date is 15-Mar-1993   |                                                              |
|                |                                       |                                                              |
|                | Paid to Date = 31 March 2017          |                                                              |
|                |                                       |                                                              |
|                | Product: CWLA                         |                                                              |
|                |                                       |                                                              |
|                | Policy Status = AL                    |                                                              |
|                |                                       |                                                              |
|                | Current Date = 10 August 2018         |                                                              |
+----------------+---------------------------------------+--------------------------------------------------------------+
| IR_1           | **[Product Eligibility]{.underline}** | Instalment Revival request is allowed to be indexed.         |
|                |                                       |                                                              |
|                | Policy number XXXX00011,              |                                                              |
|                |                                       |                                                              |
|                | Policy issuance date is 15-Mar-1993   |                                                              |
|                |                                       |                                                              |
|                | Paid to Date = 31 March 2017          |                                                              |
|                |                                       |                                                              |
|                | Product: EA                           |                                                              |
|                |                                       |                                                              |
|                | Policy Status = AL                    |                                                              |
|                |                                       |                                                              |
|                | Current Date = 10 August 2018         |                                                              |
+----------------+---------------------------------------+--------------------------------------------------------------+
| IR_2           | **[Policy Status]{.underline}**       | Instalment Revival request is allowed to be indexed.         |
|                |                                       |                                                              |
|                | Policy number XXXX00011,              |                                                              |
|                |                                       |                                                              |
|                | Policy issuance date is 15-Mar-1993   |                                                              |
|                |                                       |                                                              |
|                | Paid to Date = 31 March 2017          |                                                              |
|                |                                       |                                                              |
|                | Current Date = 06 August 2018         |                                                              |
|                |                                       |                                                              |
|                | Product: WLA                          |                                                              |
|                |                                       |                                                              |
|                | Policy Status = AL                    |                                                              |
+----------------+---------------------------------------+--------------------------------------------------------------+
| IR_2           | **[Policy Status]{.underline}**       | Instalment Revival request is not allowed to be indexed as   |
|                |                                       | policy status is AP.                                         |
|                | Policy number XXXX00011,              |                                                              |
|                |                                       |                                                              |
|                | Policy issuance date is 15-Mar-1993   |                                                              |
|                |                                       |                                                              |
|                | Paid to Date = 31 August 2018         |                                                              |
|                |                                       |                                                              |
|                | Current Date = 06 August 2018         |                                                              |
|                |                                       |                                                              |
|                | Product: EA                           |                                                              |
|                |                                       |                                                              |
|                | Policy Status = AP                    |                                                              |
+----------------+---------------------------------------+--------------------------------------------------------------+
| IR_2           | **[Policy Status]{.underline}**       | Instalment Revival request is not allowed to be indexed as   |
|                |                                       | policy status is not AL.                                     |
|                | Policy number XXXX00011,              |                                                              |
|                |                                       |                                                              |
|                | Policy issuance date is 15-Mar-1993   |                                                              |
|                |                                       |                                                              |
|                | Paid to Date = 31 March 2017          |                                                              |
|                |                                       |                                                              |
|                | Current Date = 06 August 2018         |                                                              |
|                |                                       |                                                              |
|                | Product: CWLA                         |                                                              |
|                |                                       |                                                              |
|                | Policy Status = Pending Surrender     |                                                              |
+----------------+---------------------------------------+--------------------------------------------------------------+
| IR_3           | **[Minimum number of                  | Minimum Number of instalments shown in the dropdown on the   |
|                | instalments]{.underline}**            | Revival screen is 2.                                         |
|                |                                       |                                                              |
|                | Policy number XXXX00011,              |                                                              |
|                |                                       |                                                              |
|                | Policy issuance date is 15-Mar-1993   |                                                              |
|                |                                       |                                                              |
|                | Paid to Date = 31 March 2017          |                                                              |
|                |                                       |                                                              |
|                | Current Date = 06 August 2018         |                                                              |
|                |                                       |                                                              |
|                | Product: CWLA                         |                                                              |
|                |                                       |                                                              |
|                | Policy Status = AL                    |                                                              |
|                |                                       |                                                              |
|                | Customer opts for instalments         |                                                              |
+----------------+---------------------------------------+--------------------------------------------------------------+
| IR_4           | **[Maximum Number of                  | Maximum Number of instalments shown in the dropdown on the   |
|                | Installments]{.underline}**           | Revival screen is 12.                                        |
|                |                                       |                                                              |
|                | Policy number XXXX00011,              |                                                              |
|                |                                       |                                                              |
|                | Policy issuance date is 15-Mar-1993   |                                                              |
|                |                                       |                                                              |
|                | Paid to Date = 31 March 2017          |                                                              |
|                |                                       |                                                              |
|                | Current Date = 06 August 2018         |                                                              |
|                |                                       |                                                              |
|                | Product: CWLA                         |                                                              |
|                |                                       |                                                              |
|                | Policy Status = AL                    |                                                              |
|                |                                       |                                                              |
|                | Customer opts for instalments         |                                                              |
+----------------+---------------------------------------+--------------------------------------------------------------+
| IR_4           | **[Maximum Number of                  | User will not allowed to pay the first instalment because    |
|                | Installments]{.underline}**           | per calculation the due date of 10^th^ Instalment will fall  |
|                |                                       | on 1^st^ June 2019 if the customer makes payment of 15 Sep   |
|                | Policy number XXXX00011,              | 2019.                                                        |
|                |                                       |                                                              |
|                | Policy issuance date is 15-Mar-1993   | As per rule no instalment due date should fall in or after   |
|                |                                       | maturity month.                                              |
|                | Paid to Date = 31 March 2017          |                                                              |
|                |                                       | Error *"Due date of instalment is in or after maturity       |
|                | Product: CWLA                         | month. Please index new request"* will be thrown on the      |
|                |                                       | instalment collection screen.                                |
|                | Policy Status = AL                    |                                                              |
|                |                                       | User need to withdraw and Index new request.                 |
|                | Instalment revival request Approved   |                                                              |
|                | on 06 August 2018                     |                                                              |
|                |                                       |                                                              |
|                | Customer opts 10 instalments          |                                                              |
|                |                                       |                                                              |
|                | Maturity Date = 15 June 2019          |                                                              |
|                |                                       |                                                              |
|                | Customer tries to make payment of     |                                                              |
|                | first instalment on 15 Sep 2018       |                                                              |
+----------------+---------------------------------------+--------------------------------------------------------------+
| IR_5           | **[Calculations]{.underline}**        |  Revival Amount =   { (1.01\^11 -- 1) \* 101} \* 500 =       |
|                |                                       | 5841.25                                                      |
|                |   ---------------------------------   |                                                              |
|                |   Policy Issue   30/Jan/2014          | Instalment Revival Formula: (Revival Amount X Interest) X    |
|                |   Date                                | \[{(1+Interest) No. of Instalments opted} / {(1+Interest)    |
|                |   -------------- ------------------   | No. of Instalments opted -1}\]                               |
|                |   Paid to Date   30/November/2017     |                                                              |
|                |                                       | [Instalment Amount]{.underline} =  (5841.25 \*0.01) \* \[ {  |
|                |   Revival        30/November/2018     | (1.01)\^5}  /   {(1.01)\^5 -- 1}\] = 1203.53                 |
|                |   Indexing Date                       |                                                              |
|                |   (Current Date)                      | [Tax on Total Unpaid Premiums:]{.underline}                  |
|                |                                       |                                                              |
|                |   Modal Premium  500                  | GST Renewal Year @ [2.25]{.mark}% (on total unpaid Premiums) |
|                |                                       | = 5500 \* 2.25% = 123.75                                     |
|                |   Number of      11                   |                                                              |
|                |   unpaid Premium                      | [Tax on Modal Premiums:]{.underline}                         |
|                |   Months                              |                                                              |
|                |                                       | GST Renewal Year @ 2.25% (on Modal Premiums) = 500 \* 2.25%  |
|                |   Total Unpaid   500\*11 = 5500       | = 11.25                                                      |
|                |   Premiums                            |                                                              |
|                |                                       | [First Instalment:]{.underline}                              |
|                |   Number of      5                    |                                                              |
|                |   installment                         |  Amount to be paid at the time of first Instalment =         |
|                |   ---------------------------------   | Instalment Amount + Modal Premium for current month + GST on |
|                |                                       | Total Unpaid Premium Amount + GST on Modal                   |
|                |                                       | Premium.                                                     |
|                |                                       |                                                              |
|                |                                       | Amount to be paid at the time of first Instalment =          |
|                |                                       | 1203.53 + 500 + 123.75 + 11.25 = 1838.53                     |
|                |                                       |                                                              |
|                |                                       | [Subsequent Instalments:]{.underline}                        |
|                |                                       |                                                              |
|                |                                       | Amount to be paid as subsequent instalment = Instalment      |
|                |                                       | Amount + Modal Premium + Service Tax on Modal Premium.       |
|                |                                       |                                                              |
|                |                                       | Amount to be paid as subsequent instalment = 1203.53 + 500 + |
|                |                                       | 11.25 = 1714.78                                              |
+----------------+---------------------------------------+--------------------------------------------------------------+
| IR_5           | **[Calculation]{.underline}**         | [Revival quotation date = 31 July 2018]{.mark}               |
|                |                                       |                                                              |
|                |   -----------------------------       | [Interest will be calculated up to current month --          |
|                |   Policy Issue   30/Jan/2014          | 1]{.mark}                                                    |
|                |   Date                                |                                                              |
|                |   -------------- --------------       | [Therefore, Number of months for interest calculation = Sep  |
|                |   Paid to Date   31/August            | 2015 till June 2018 = 34 Months]{.mark}                      |
|                |                  /2015                |                                                              |
|                |                                       | **[[Total unpaid Premiums]{.underline}]{.mark}**             |
|                |   Revival        31/July/2018         |                                                              |
|                |   Indexing Date                       |   ---------------------------------------------------------- |
|                |   (Current Date)                      |   [Cycle]{.mark}          [   Year]{.mark} [Modal            |
|                |                                       |                                            Premium]{.mark}   |
|                |   Number of      5                    |   --------------------- ------------------ ----------------- |
|                |   installment                         |   [Sep-Feb]{.mark}        [2015-16]{.mark} [5852]{.mark}     |
|                |   -----------------------------       |                                                              |
|                |                                       |   [Mar-August]{.mark}     [2016-16]{.mark} [5852]{.mark}     |
|                | Billing Frequency change effective    |                                                              |
|                | from Dec 2014 = Quarterly             |   [Sep-Feb]{.mark}        [2016-17]{.mark} [5852]{.mark}     |
|                |                                       |                                                              |
|                | Billing Frequency change effective    |   [Mar-August]{.mark}     [2017-17]{.mark} [5852]{.mark}     |
|                | from March 2015 = Semi Annual         |                                                              |
|                |                                       |   [Sep-Feb]{.mark}        [2017-18]{.mark} [5852]{.mark}     |
|                | Modal Premium changed from 2966 to    |                                                              |
|                | 5852                                  |   [Mar-Aug]{.mark}        [2018-18]{.mark} [5852]{.mark}     |
|                |                                       |                                                              |
|                |                                       |   [Total Unpaid                [36]{.mark} [35,112]{.mark}   |
|                |                                       |   Premium]{.mark}                                            |
|                |                                       |   ---------------------------------------------------------- |
|                |                                       |                                                              |
|                |                                       | [Monthly Premium = Unpaid Premiums/Number of unpaid          |
|                |                                       | months]{.mark}                                               |
|                |                                       |                                                              |
|                |                                       | [Monthly Premium = 35,112/36 = 975.33]{.mark}                |
|                |                                       |                                                              |
|                |                                       | [Revival Amount = {(1.01)\^34 -- 1} \*101 \* 975.33 =        |
|                |                                       | 39657.18]{.mark}                                             |
|                |                                       |                                                              |
|                |                                       | [Interest (for 34 months) = Monthly Premium x No. of Months  |
|                |                                       | for which interest to be charged - Revival Amount.]{.mark}   |
|                |                                       |                                                              |
|                |                                       | [975.33\*34-39657.18=6495.96]{.mark}                         |
|                |                                       |                                                              |
|                |                                       | [Total Revival Amount to be paid= Unpaid Premium (for 36     |
|                |                                       | months)+ Interest (for 34 months).]{.mark}                   |
|                |                                       |                                                              |
|                |                                       | [35112+6495.96=41607.97]{.mark}                              |
|                |                                       |                                                              |
|                |                                       | [[Instalment Amount]{.underline} =  (41607.97\*0.01) \* \[ { |
|                |                                       | (1.01)\^5}  /   {(1.01)\^5 -- 1}\] =]{.mark}                 |
|                |                                       |                                                              |
|                |                                       | [8572.89]{.mark}                                             |
|                |                                       |                                                              |
|                |                                       | [[Tax on Total Unpaid Premiums:]{.underline}]{.mark}         |
|                |                                       |                                                              |
|                |                                       | [GST Renewal Year @ 2.25% (on total unpaid Premiums) =       |
|                |                                       | 35112\* 2.25% = 790.02]{.mark}                               |
|                |                                       |                                                              |
|                |                                       | [[First Instalment:]{.underline}]{.mark}                     |
|                |                                       |                                                              |
|                |                                       | [ Amount to be paid at the time of first Instalment =        |
|                |                                       | Instalment Amount + GST on Total Unpaid Premium              |
|                |                                       | Amount]{.mark}                                               |
|                |                                       |                                                              |
|                |                                       | [Amount to be paid at the time of first Instalment =         |
|                |                                       | 8572.89 + 790.02 = 9363.00]{.mark}                           |
|                |                                       |                                                              |
|                |                                       | [\*Premium as and when due as per frequency will have to be  |
|                |                                       | paid by the insurant.]{.mark}                                |
|                |                                       |                                                              |
|                |                                       | [\*This revival instalment calculation is valid till the end |
|                |                                       | of current month i.e. 31^st^ July 2018.]{.mark}              |
+----------------+---------------------------------------+--------------------------------------------------------------+
| IR_6           | **[Tax calculation]{.underline}**     | Number of unpaid months = 15                                 |
|                |                                       |                                                              |
|                | Policy number XXXX00011,              | Total unpaid Premiums = 1175 \* 15 = 17625                   |
|                |                                       |                                                              |
|                | Policy issuance date is 15-Mar-1993   | Taxes = 17625 \* 2.25/100 = 396.56                           |
|                |                                       |                                                              |
|                | Paid to Date = 31 April 2017          | Taxes for July Premium = 1175 \* 2.25/100 = 26.43            |
|                |                                       |                                                              |
|                | Product: CWLA                         | Rs.422 (396 + 26) will be paid as taxes at the time of first |
|                |                                       | instalment.                                                  |
|                | Policy Status = AL                    |                                                              |
|                |                                       |                                                              |
|                | Revival Date = 01 July 2017           |                                                              |
|                |                                       |                                                              |
|                | Premium Amount = 1175                 |                                                              |
+----------------+---------------------------------------+--------------------------------------------------------------+
| IR_6           | **[Tax calculation]{.underline}**     | 2^nd^ instalment is paid. Taxes on 2^nd^ instalment will not |
|                |                                       | be paid.                                                     |
|                | Policy number XXXX00011,              |                                                              |
|                |                                       |                                                              |
|                | Policy issuance date is 15-Mar-1993   |                                                              |
|                |                                       |                                                              |
|                | Paid to Date = 30 April 2017          |                                                              |
|                |                                       |                                                              |
|                | Product: CWLA                         |                                                              |
|                |                                       |                                                              |
|                | Policy Status = AL                    |                                                              |
|                |                                       |                                                              |
|                | Due date = 01 May 2017                |                                                              |
|                |                                       |                                                              |
|                | Revival Date = 01 June 2018           |                                                              |
|                |                                       |                                                              |
|                | Premium Amount = 1175                 |                                                              |
|                |                                       |                                                              |
|                | Total unpaid Premiums = 1175 \* 13 =  |                                                              |
|                | 15275                                 |                                                              |
|                |                                       |                                                              |
|                | Taxes = 15275 \* 2.25/100 = 343.687   |                                                              |
|                |                                       |                                                              |
|                | On July 2018 2^nd^ instalment is paid |                                                              |
+----------------+---------------------------------------+--------------------------------------------------------------+
| IR_7           | **[Mode of Instalment                 | Collection through Cash is allowed.                          |
|                | Collection]{.underline}**             |                                                              |
|                |                                       |                                                              |
|                | Policy number XXXX00011,              |                                                              |
|                |                                       |                                                              |
|                | Policy issuance date is 15-Mar-1993   |                                                              |
|                |                                       |                                                              |
|                | Paid to Date = 01 April 2017          |                                                              |
|                |                                       |                                                              |
|                | Product: CWLA                         |                                                              |
|                |                                       |                                                              |
|                | Policy Status = AL                    |                                                              |
|                |                                       |                                                              |
|                | Due date = 01 May 2017                |                                                              |
|                |                                       |                                                              |
|                | Revival Date = 01 June 2018           |                                                              |
|                |                                       |                                                              |
|                | Premium Amount = 1175                 |                                                              |
|                |                                       |                                                              |
|                | Instalment revival request approved   |                                                              |
|                | on 25 June 2018                       |                                                              |
|                |                                       |                                                              |
|                | Mode selected Cash                    |                                                              |
|                |                                       |                                                              |
|                | 1^st^ Instalment collected on 11 July |                                                              |
|                | 2018                                  |                                                              |
+----------------+---------------------------------------+--------------------------------------------------------------+
| IR_7           | **[Mode of Instalment                 | Request status will be 'Approved'                            |
|                | Collection]{.underline}**             |                                                              |
|                |                                       | Policy status will change to AL                              |
|                | Policy number XXXX00011,              |                                                              |
|                |                                       | Paid to Date will be 01 April 2017                           |
|                | Policy issuance date is 15-Mar-1993   |                                                              |
|                |                                       | Rs.1000 will move to suspense                                |
|                | Paid to Date = 01 April 2017          |                                                              |
|                |                                       |                                                              |
|                | Product: CWLA                         |                                                              |
|                |                                       |                                                              |
|                | Policy Status = AL                    |                                                              |
|                |                                       |                                                              |
|                | Due date = 01 May 2017                |                                                              |
|                |                                       |                                                              |
|                | Revival Date = 01 June 2018           |                                                              |
|                |                                       |                                                              |
|                | Premium Amount = 1175                 |                                                              |
|                |                                       |                                                              |
|                | Instalment revival request approved   |                                                              |
|                | on 25 June 2018                       |                                                              |
|                |                                       |                                                              |
|                | As on 26 June 2018, customer made the |                                                              |
|                | payment of first instalment through   |                                                              |
|                | cheque and cash.                      |                                                              |
|                |                                       |                                                              |
|                | Cash = 1000                           |                                                              |
|                |                                       |                                                              |
|                | Cheque = 175                          |                                                              |
|                |                                       |                                                              |
|                | Cheque dishonored on 1 July 2018      |                                                              |
+----------------+---------------------------------------+--------------------------------------------------------------+
| IR_7           | **[Mode of Instalment                 | 2^nd^ instalment payment is not allowed.                     |
|                | Collection]{.underline}**             |                                                              |
|                |                                       | Policy status will change to AL                              |
|                | Policy number XXXX00011,              |                                                              |
|                |                                       | Paid to Date will be 01 April 2017                           |
|                | Policy issuance date is 15-Mar-1993   |                                                              |
|                |                                       | Rs.1000 will move to suspense                                |
|                | Paid to Date = 01 April 2017          |                                                              |
|                |                                       |                                                              |
|                | Product: CWLA                         |                                                              |
|                |                                       |                                                              |
|                | Policy Status = AL                    |                                                              |
|                |                                       |                                                              |
|                | Due date = 01 May 2017                |                                                              |
|                |                                       |                                                              |
|                | Revival Date = 01 June 2018           |                                                              |
|                |                                       |                                                              |
|                | Premium Amount = 1175                 |                                                              |
|                |                                       |                                                              |
|                | Instalment revival request approved   |                                                              |
|                | on 25 June 2018                       |                                                              |
|                |                                       |                                                              |
|                | As on 26 June 2018, customer made the |                                                              |
|                | payment of first instalment through   |                                                              |
|                | cheque and cash.                      |                                                              |
|                |                                       |                                                              |
|                | Cash = 1000                           |                                                              |
|                |                                       |                                                              |
|                | Cheque = 175                          |                                                              |
|                |                                       |                                                              |
|                | Cheque not cleared till 1 July 2018   |                                                              |
+----------------+---------------------------------------+--------------------------------------------------------------+
| IR_7           | **[Mode of Instalment                 | Error will be thrown "Instalment payment not allowed through |
|                | Collection]{.underline}**             | Meghdoot upload/Employer portal"                             |
|                |                                       |                                                              |
|                | Policy number XXXX00011,              |                                                              |
|                |                                       |                                                              |
|                | Policy issuance date is 15-Mar-1993   |                                                              |
|                |                                       |                                                              |
|                | Paid to Date = 01 April 2017          |                                                              |
|                |                                       |                                                              |
|                | Product: CWLA                         |                                                              |
|                |                                       |                                                              |
|                | Policy Status = AL                    |                                                              |
|                |                                       |                                                              |
|                | Due date = 01 May 2017                |                                                              |
|                |                                       |                                                              |
|                | Revival Date = 01 June 2018           |                                                              |
|                |                                       |                                                              |
|                | Premium Amount = 1175                 |                                                              |
|                |                                       |                                                              |
|                | Instalment revival request approved   |                                                              |
|                | on 25 June 2018                       |                                                              |
|                |                                       |                                                              |
|                | As on 26 June -- User tries to make   |                                                              |
|                | payment of instalment through         |                                                              |
|                | Meghdoot upload.                      |                                                              |
+----------------+---------------------------------------+--------------------------------------------------------------+
| IR_8           | **[Installment                        | 2^nd^ Instalment will be due on 1^st^ September 2018.        |
|                | Frequency]{.underline}**              |                                                              |
|                |                                       | 3^rd^ Instalment will be due on 1^st^ October 2018.          |
|                | Policy number XXXX00011,              |                                                              |
|                |                                       |                                                              |
|                | Policy issuance date is 15-Mar-1993   |                                                              |
|                |                                       |                                                              |
|                | Paid to Date = 31 March 2017          |                                                              |
|                |                                       |                                                              |
|                | Product: CWLA                         |                                                              |
|                |                                       |                                                              |
|                | Policy Status = AL                    |                                                              |
|                |                                       |                                                              |
|                | Premium Frequency = Semi Annual       |                                                              |
|                |                                       |                                                              |
|                | Instalment Amount = 1000              |                                                              |
|                |                                       |                                                              |
|                | Taxes on unpaid premiums = 300        |                                                              |
|                |                                       |                                                              |
|                | Customer opted 3 instalments          |                                                              |
|                |                                       |                                                              |
|                | Instalment Revival request Approved   |                                                              |
|                | on 06 August 2018.                    |                                                              |
|                |                                       |                                                              |
|                | Collection(Rs. 1300) done on 10       |                                                              |
|                | August 2018                           |                                                              |
+----------------+---------------------------------------+--------------------------------------------------------------+
| IR_9           | **[Grace Period]{.underline}**        | There is no Grace period.(Not permissible under POLI rules)  |
|                |                                       |                                                              |
|                | Policy number XXXX00011,              | 2^nd^ Instalment will be due on 1^st^ September 2018.        |
|                |                                       |                                                              |
|                | Policy issuance date is 15-Mar-1993   | If Instalment is not paid on 1^st^ September 2018, then      |
|                |                                       | policy status will be AL and paid to date will be 31 March   |
|                | Paid to Date = 31 March 2017          | 2017.                                                        |
|                |                                       |                                                              |
|                | Product: CWLA                         | Rs.1300 will move to suspense                                |
|                |                                       |                                                              |
|                | Policy Status = AL                    |                                                              |
|                |                                       |                                                              |
|                | Premium Frequency = Semi Annual       |                                                              |
|                |                                       |                                                              |
|                | Instalment Amount = 1000              |                                                              |
|                |                                       |                                                              |
|                | Taxes on unpaid premiums = 300        |                                                              |
|                |                                       |                                                              |
|                | Customer opted 3 instalments          |                                                              |
|                |                                       |                                                              |
|                | Instalment Revival request Approved   |                                                              |
|                | on 06 August 2018.                    |                                                              |
|                |                                       |                                                              |
|                | Collection(Rs. 1300) done on 10       |                                                              |
|                | August 2018                           |                                                              |
+----------------+---------------------------------------+--------------------------------------------------------------+
| IR_10          | **[SLA for First                      | As collection is not done, system will recalculate revival   |
|                | Instalment]{.underline}**             | instalment amount and policy status and paid to date will    |
|                |                                       | remain as AL and 31 March 2017 respectively.                 |
|                | Policy number XXXX00011,              |                                                              |
|                |                                       |                                                              |
|                | Policy issuance date is 15-Mar-1993   |                                                              |
|                |                                       |                                                              |
|                | Paid to Date = 31 March 2017          |                                                              |
|                |                                       |                                                              |
|                | Product: WLA                          |                                                              |
|                |                                       |                                                              |
|                | Policy Status = AL                    |                                                              |
|                |                                       |                                                              |
|                | Frequency = Monthly                   |                                                              |
|                |                                       |                                                              |
|                | Instalment Revival request indexed on |                                                              |
|                | 06 August 2018.                       |                                                              |
|                |                                       |                                                              |
|                | Instalment revival Request Approved   |                                                              |
|                | on 10 August 2018.                    |                                                              |
|                |                                       |                                                              |
|                | Collection not done till 11 Oct 2018  |                                                              |
+----------------+---------------------------------------+--------------------------------------------------------------+
| IR_11          | **[Due Date]{.underline}**            | 2^nd^ Instalment will be due on 1^st^ September 2018.        |
|                |                                       |                                                              |
|                | Policy number XXXX00011,              | 3^rd^ Instalment will be due on 1^st^ October 2018.          |
|                |                                       |                                                              |
|                | Policy issuance date is 15-Mar-1993   |                                                              |
|                |                                       |                                                              |
|                | Paid to Date = 31 March 2017          |                                                              |
|                |                                       |                                                              |
|                | Product: CWLA                         |                                                              |
|                |                                       |                                                              |
|                | Policy Status = AL                    |                                                              |
|                |                                       |                                                              |
|                | Premium Frequency = Semi Annual       |                                                              |
|                |                                       |                                                              |
|                | Instalment Amount = 1000              |                                                              |
|                |                                       |                                                              |
|                | Taxes on unpaid premiums = 300        |                                                              |
|                |                                       |                                                              |
|                | Customer opted 3 instalments          |                                                              |
|                |                                       |                                                              |
|                | Instalment Revival request Approved   |                                                              |
|                | on 06 August 2018.                    |                                                              |
|                |                                       |                                                              |
|                | Collection(Rs. 1300) done on 10       |                                                              |
|                | August 2018                           |                                                              |
+----------------+---------------------------------------+--------------------------------------------------------------+
| IR_12          | **[Payment of Premium]{.underline}**  | Premium of Rs.500 will be due on 1 Oct 2018.                 |
|                |                                       |                                                              |
|                | Policy number XXXX00011,              |                                                              |
|                |                                       |                                                              |
|                | Policy issuance date is 15-Mar-1993   |                                                              |
|                |                                       |                                                              |
|                | Paid to Date = 31 March 2017          |                                                              |
|                |                                       |                                                              |
|                | Product: CWLA                         |                                                              |
|                |                                       |                                                              |
|                | Policy Status = AL                    |                                                              |
|                |                                       |                                                              |
|                | Premium Frequency = quarterly         |                                                              |
|                |                                       |                                                              |
|                | Instalment Amount = 1000              |                                                              |
|                |                                       |                                                              |
|                | Taxes on unpaid premiums = 300        |                                                              |
|                |                                       |                                                              |
|                | Premium Amount = 500 per quarter      |                                                              |
|                |                                       |                                                              |
|                | Customer opted 3 instalments          |                                                              |
|                |                                       |                                                              |
|                | Instalment Revival Indexed on 1       |                                                              |
|                | August 2018                           |                                                              |
|                |                                       |                                                              |
|                | Instalment Revival request Approved   |                                                              |
|                | on 06 August 2018.                    |                                                              |
|                |                                       |                                                              |
|                | Collection(Rs.1800) done on 10 August |                                                              |
|                | 2018                                  |                                                              |
+----------------+---------------------------------------+--------------------------------------------------------------+
| IR_12          | **[Payment of Premium]{.underline}**  | Premium of Rs.500 will be due for September 2018 along with  |
|                |                                       | the first instalment amount.                                 |
|                | Policy number XXXX00011,              |                                                              |
|                |                                       | Next Premium will be due on 1^st^ Oct 2018                   |
|                | Policy issuance date is 15-Mar-1993   |                                                              |
|                |                                       |                                                              |
|                | Paid to Date = 31 March 2017          |                                                              |
|                |                                       |                                                              |
|                | Product: CWLA                         |                                                              |
|                |                                       |                                                              |
|                | Policy Status = AL                    |                                                              |
|                |                                       |                                                              |
|                | Premium Frequency = Monthly           |                                                              |
|                |                                       |                                                              |
|                | Instalment Amount = 1000              |                                                              |
|                |                                       |                                                              |
|                | Taxes on unpaid premiums = 300        |                                                              |
|                |                                       |                                                              |
|                | Premium Amount = 500                  |                                                              |
|                |                                       |                                                              |
|                | Customer opted 3 instalments          |                                                              |
|                |                                       |                                                              |
|                | Instalment Revival Indexed on 1       |                                                              |
|                | August 2018                           |                                                              |
|                |                                       |                                                              |
|                | Instalment Revival request Approved   |                                                              |
|                | on 06 August 2018.                    |                                                              |
|                |                                       |                                                              |
|                | Collection(Rs.1300) done on 10        |                                                              |
|                | September 2018                        |                                                              |
+----------------+---------------------------------------+--------------------------------------------------------------+
| IR_13          | **[First Instalment                   | Policy status will remain AL and Paid to Date will be 31     |
|                | payment]{.underline}**                | March 2017                                                   |
|                |                                       |                                                              |
|                | Policy number XXXX00011,              |                                                              |
|                |                                       |                                                              |
|                | Policy issuance date is 15-Mar-1993   |                                                              |
|                |                                       |                                                              |
|                | Paid to Date = 31 March 2017          |                                                              |
|                |                                       |                                                              |
|                | Product: WLA                          |                                                              |
|                |                                       |                                                              |
|                | Policy Status = AL                    |                                                              |
|                |                                       |                                                              |
|                | Frequency = Monthly                   |                                                              |
|                |                                       |                                                              |
|                | Instalment Revival request indexed on |                                                              |
|                | 06 August 2018.                       |                                                              |
|                |                                       |                                                              |
|                | Instalment revival Request Approved   |                                                              |
|                | on 10 August 2018.                    |                                                              |
|                |                                       |                                                              |
|                | Collection not yet done.              |                                                              |
|                |                                       |                                                              |
|                | Current Date = 11 August 2018         |                                                              |
+----------------+---------------------------------------+--------------------------------------------------------------+
| IR_13          | **[First Instalment                   | Policy status will change to AP and Paid to Date will be 31  |
|                | payment]{.underline}**                | August 2018                                                  |
|                |                                       |                                                              |
|                | Policy number XXXX00011,              |                                                              |
|                |                                       |                                                              |
|                | Policy issuance date is 15-Mar-1993   |                                                              |
|                |                                       |                                                              |
|                | Paid to Date = 31 March 2017          |                                                              |
|                |                                       |                                                              |
|                | Product: WLA                          |                                                              |
|                |                                       |                                                              |
|                | Policy Status = AL                    |                                                              |
|                |                                       |                                                              |
|                | Frequency = Monthly                   |                                                              |
|                |                                       |                                                              |
|                | Instalment Revival request indexed on |                                                              |
|                | 06 August 2018.                       |                                                              |
|                |                                       |                                                              |
|                | Instalment revival Request Approved   |                                                              |
|                | on 10 August 2018.                    |                                                              |
|                |                                       |                                                              |
|                | Collection done on 11 August          |                                                              |
|                |                                       |                                                              |
|                | Current Date = 11 August 2018         |                                                              |
+----------------+---------------------------------------+--------------------------------------------------------------+
| IR_14          | **[Advance Premiums]{.underline}**    | Advance premium payment is allowed. User can opt to have 'To |
|                |                                       | Date' as 30 October 2018.                                    |
|                | Policy number XXXX00011,              |                                                              |
|                |                                       |                                                              |
|                | Policy issuance date is 15-Mar-1993   |                                                              |
|                |                                       |                                                              |
|                | Paid to Date = 31 March 2017          |                                                              |
|                |                                       |                                                              |
|                | Product: CWLA                         |                                                              |
|                |                                       |                                                              |
|                | Policy Status = AL                    |                                                              |
|                |                                       |                                                              |
|                | Frequency = Monthly                   |                                                              |
|                |                                       |                                                              |
|                | Modal Premium = 200                   |                                                              |
|                |                                       |                                                              |
|                | Revival Request Indexed on 01 August  |                                                              |
|                | 2018.                                 |                                                              |
|                |                                       |                                                              |
|                | Instalment Revival request Approved   |                                                              |
|                | on 06 August 2018.                    |                                                              |
|                |                                       |                                                              |
|                | Collection done on 10 August 2018     |                                                              |
|                |                                       |                                                              |
|                | Paid to Date is updated as 31 August  |                                                              |
|                | 2018.                                 |                                                              |
|                |                                       |                                                              |
|                | Premium paid on 01 September = 600.   |                                                              |
+----------------+---------------------------------------+--------------------------------------------------------------+
| IR_15          | **[Advance Instalments]{.underline}** | Policyholder is allowed to pay advance instalment.           |
|                |                                       |                                                              |
|                | Policy number XXXX00011,              | 1000 will be applied to 2^nd^ instalment and 500 will be     |
|                |                                       | kept in suspense which will be adjusted against 3^rd^        |
|                | Policy issuance date is 15-Mar-1993   | instalment.                                                  |
|                |                                       |                                                              |
|                | Paid to Date = 31 March 2017          | At the time of 3^rd^ instalment user need to pay only        |
|                |                                       |                                                              |
|                | Product: CWLA                         |                                                              |
|                |                                       |                                                              |
|                | Policy Status = AL                    |                                                              |
|                |                                       |                                                              |
|                | Frequency = Monthly                   |                                                              |
|                |                                       |                                                              |
|                | Instalment Amount = 1000              |                                                              |
|                |                                       |                                                              |
|                | Taxes on unpaid premiums = 300        |                                                              |
|                |                                       |                                                              |
|                | Customer opted 3 instalments          |                                                              |
|                |                                       |                                                              |
|                | Instalment Revival request Approved   |                                                              |
|                | on 06 August 2018.                    |                                                              |
|                |                                       |                                                              |
|                | Collection(Rs.1300) done on 10 August |                                                              |
|                | 2018                                  |                                                              |
|                |                                       |                                                              |
|                | At the time of 2^nd^ instalment on 01 |                                                              |
|                | September, policyholder paid = 1500   |                                                              |
+----------------+---------------------------------------+--------------------------------------------------------------+
| IR_15          | **[Advance Instalments]{.underline}** | Policyholder is allowed to pay all the instalments at the    |
|                |                                       | time of first instalment only and needs to adjust 'To date'  |
|                | Policy number XXXX00011,              | accordingly..                                                |
|                |                                       |                                                              |
|                | Policy issuance date is 15-Mar-1993   |                                                              |
|                |                                       |                                                              |
|                | Paid to Date = 31 March 2017          |                                                              |
|                |                                       |                                                              |
|                | Product: CWLA                         |                                                              |
|                |                                       |                                                              |
|                | Policy Status = AL                    |                                                              |
|                |                                       |                                                              |
|                | Frequency = Monthly                   |                                                              |
|                |                                       |                                                              |
|                | Instalment Amount = 1000              |                                                              |
|                |                                       |                                                              |
|                | Taxes on unpaid premiums = 300        |                                                              |
|                |                                       |                                                              |
|                | Customer opted 3 instalments          |                                                              |
|                |                                       |                                                              |
|                | Instalment Revival request Approved   |                                                              |
|                | on 06 August 2018.                    |                                                              |
|                |                                       |                                                              |
|                | Collection(Rs.3300) done on 10 August |                                                              |
|                | 2018                                  |                                                              |
+----------------+---------------------------------------+--------------------------------------------------------------+
| IR_15          | **[Advance Instalments]{.underline}** | Policyholder is allowed to pay all the instalments at the    |
|                |                                       | time of first instalment only.                               |
|                | Policy number XXXX00011,              |                                                              |
|                |                                       | This Extra Rs.300 will be kept in suspense and adjusted      |
|                | Policy issuance date is 15-Mar-1993   | towards subsequent revival instalment. ~~will be disbursed   |
|                |                                       | as excess amount at the time of claim.~~                     |
|                | Paid to Date = 31 March 2017          |                                                              |
|                |                                       |                                                              |
|                | Product: CWLA                         |                                                              |
|                |                                       |                                                              |
|                | Policy Status = AL                    |                                                              |
|                |                                       |                                                              |
|                | Frequency = Monthly                   |                                                              |
|                |                                       |                                                              |
|                | Instalment Amount = 1000              |                                                              |
|                |                                       |                                                              |
|                | Taxes on unpaid premiums = 300        |                                                              |
|                |                                       |                                                              |
|                | Customer opted 3 instalments          |                                                              |
|                |                                       |                                                              |
|                | Instalment Revival request Approved   |                                                              |
|                | on 06 August 2018.                    |                                                              |
|                |                                       |                                                              |
|                | Collection(Rs.3300) done on 10 August |                                                              |
|                | 2018                                  |                                                              |
|                |                                       |                                                              |
|                | As on 1^st^ September customer paid   |                                                              |
|                | 100 Rs.                               |                                                              |
|                |                                       |                                                              |
|                | As on 1^st^ Oct 2018, customer paid   |                                                              |
|                | 200 Rs.                               |                                                              |
+----------------+---------------------------------------+--------------------------------------------------------------+
| IR_15          | **[Advance Instalments]{.underline}** | First instalment revival request completed.                  |
|                |                                       |                                                              |
|                | Policy number XXXX00011,              | Policyholder is allowed to pay all the instalments at the    |
|                |                                       | time of first instalment only.                               |
|                | Policy issuance date is 15-Mar-1993   |                                                              |
|                |                                       | .                                                            |
|                | Paid to Date = 31 March 2017          |                                                              |
|                |                                       | On the Collection screen amount of Rs.300 will be displayed  |
|                | Product: CWLA                         | under 'Amount received' field. This amount customer can use. |
|                |                                       |                                                              |
|                | Policy Status = AL                    | Customer need to pay only Rs.250 at the time of 1^st^        |
|                |                                       | instlament revival collection as on 15 September 2020.       |
|                | Frequency = Monthly                   |                                                              |
|                |                                       |                                                              |
|                | Instalment Amount = 1000              |                                                              |
|                |                                       |                                                              |
|                | Taxes on unpaid premiums = 300        |                                                              |
|                |                                       |                                                              |
|                | Customer opted 3 instalments          |                                                              |
|                |                                       |                                                              |
|                | Instalment Revival request Approved   |                                                              |
|                | on 06 August 2018.                    |                                                              |
|                |                                       |                                                              |
|                | Collection(Rs.3300) done on 10 August |                                                              |
|                | 2018                                  |                                                              |
|                |                                       |                                                              |
|                | As on 1^st^ September customer paid   |                                                              |
|                | 100 Rs.                               |                                                              |
|                |                                       |                                                              |
|                | As on 1^st^ Oct 2018, customer paid   |                                                              |
|                | 200 Rs.                               |                                                              |
|                |                                       |                                                              |
|                | 2^nd^ Instalment Revival request      |                                                              |
|                | indexed on 5 Sep 2020.                |                                                              |
|                |                                       |                                                              |
|                | Instalment revival request approved   |                                                              |
|                | on 10 Sep 2020.                       |                                                              |
|                |                                       |                                                              |
|                | Instalment amount = 500               |                                                              |
|                |                                       |                                                              |
|                | Taxes on unpaid premiums = 50         |                                                              |
|                |                                       |                                                              |
|                | Customer makes 1^st^ instalment as on |                                                              |
|                | 15 September 2020.                    |                                                              |
+----------------+---------------------------------------+--------------------------------------------------------------+
| IR_15          | **[Advance Instalments]{.underline}** | Policyholder is allowed to pay all the instalments at the    |
|                |                                       | time of first instalment only.                               |
|                | Policy number XXXX00011,              |                                                              |
|                |                                       | On the Collection screen amount of Rs.300 will be displayed  |
|                | Policy issuance date is 15-Mar-1993   | under 'Amount received' field. This amount customer can use. |
|                |                                       |                                                              |
|                | Paid to Date = 31 March 2017          | Customer need to pay only Rs.350 at the time of 1^st^        |
|                |                                       | instlament revival collection as on 15 September 2020..      |
|                | Product: CWLA                         |                                                              |
|                |                                       |                                                              |
|                | Policy Status = AL                    |                                                              |
|                |                                       |                                                              |
|                | Frequency = Monthly                   |                                                              |
|                |                                       |                                                              |
|                | Instalment Amount = 1000              |                                                              |
|                |                                       |                                                              |
|                | Taxes on unpaid premiums = 300        |                                                              |
|                |                                       |                                                              |
|                | Customer opted 3 instalments          |                                                              |
|                |                                       |                                                              |
|                | Instalment Revival request Approved   |                                                              |
|                | on 06 August 2018.                    |                                                              |
|                |                                       |                                                              |
|                | Collection(Rs.3300) done on 10 August |                                                              |
|                | 2018                                  |                                                              |
|                |                                       |                                                              |
|                | As on 1^st^ September customer paid   |                                                              |
|                | 100 Rs.                               |                                                              |
|                |                                       |                                                              |
|                | As on 1^st^ Oct 2018, customer paid   |                                                              |
|                | 200 Rs.                               |                                                              |
|                |                                       |                                                              |
|                | 2^nd^ Revival request indexed on 5    |                                                              |
|                | Sep 2020.                             |                                                              |
|                |                                       |                                                              |
|                | Revival request approved on 10 Sep    |                                                              |
|                | 2020.                                 |                                                              |
|                |                                       |                                                              |
|                | Revival amount including taxes = 650  |                                                              |
|                |                                       |                                                              |
|                | Collection done on 15 Sep 2020        |                                                              |
+----------------+---------------------------------------+--------------------------------------------------------------+
| IR_16          | **[Default of subsequent              | Policy Status will change to AL.                             |
|                | Installment]{.underline}**            |                                                              |
|                |                                       | Paid to Date will change to 31 March 2017.                   |
|                | Policy number XXXX00011,              |                                                              |
|                |                                       | Rs.1000 will move to policy suspense account.                |
|                | Policy issuance date is 15-Mar-1993   |                                                              |
|                |                                       | Taxes will not move to suspense.                             |
|                | Paid to Date = 31 March 2017          |                                                              |
|                |                                       |                                                              |
|                | Product: CWLA                         |                                                              |
|                |                                       |                                                              |
|                | Policy Status = AL                    |                                                              |
|                |                                       |                                                              |
|                | Frequency = Monthly                   |                                                              |
|                |                                       |                                                              |
|                | Instalment Revival request indexed on |                                                              |
|                | 06 August 2018.                       |                                                              |
|                |                                       |                                                              |
|                | Instalment revival Request Approved   |                                                              |
|                | on 10 August 2018.                    |                                                              |
|                |                                       |                                                              |
|                | Instalment amount = 1000              |                                                              |
|                |                                       |                                                              |
|                | Taxes on unpaid premiums = 300        |                                                              |
|                |                                       |                                                              |
|                | Collection of amount = 1300, done on  |                                                              |
|                | 11 August 2018                        |                                                              |
|                |                                       |                                                              |
|                | 2^nd^ instalment due on 01 September  |                                                              |
|                | 2018                                  |                                                              |
|                |                                       |                                                              |
|                | Amount not paid for 2^nd^ instalment  |                                                              |
|                |                                       |                                                              |
|                | Current Date = 2 Sep 2018             |                                                              |
+----------------+---------------------------------------+--------------------------------------------------------------+
| IR_17          | **[Existing request]{.underline}**    | New Revival instalment request will not be indexed as there  |
|                |                                       | is already pending revival request.                          |
|                | Policy number XXXX00011,              |                                                              |
|                |                                       |                                                              |
|                | Policy issuance date is 15-Mar-1993   |                                                              |
|                |                                       |                                                              |
|                | Paid to Date = 31 March 2017          |                                                              |
|                |                                       |                                                              |
|                | Product: CWLA                         |                                                              |
|                |                                       |                                                              |
|                | Policy Status = AL                    |                                                              |
|                |                                       |                                                              |
|                | Frequency = Monthly                   |                                                              |
|                |                                       |                                                              |
|                | Instalment Revival request indexed on |                                                              |
|                | 06 August 2018.                       |                                                              |
|                |                                       |                                                              |
|                | Instalment revival Request not yet    |                                                              |
|                | approved.                             |                                                              |
|                |                                       |                                                              |
|                | User tried to index new instalment    |                                                              |
|                | revival request.                      |                                                              |
+----------------+---------------------------------------+--------------------------------------------------------------+
| IR_18          | **[Surrender/Death Claim/Maturity     | Surrender indexing is not allowed till all the instalments   |
|                | Claim]{.underline}**                  | are paid.                                                    |
|                |                                       |                                                              |
|                | Policy number XXXX00011,              |                                                              |
|                |                                       |                                                              |
|                | Policy issuance date is 15-Mar-1993   |                                                              |
|                |                                       |                                                              |
|                | Paid to Date = 31 March 2017          |                                                              |
|                |                                       |                                                              |
|                | Product: CWLA                         |                                                              |
|                |                                       |                                                              |
|                | Policy Status = AL                    |                                                              |
|                |                                       |                                                              |
|                | Frequency = Monthly                   |                                                              |
|                |                                       |                                                              |
|                | Instalment Revival request indexed on |                                                              |
|                | 06 August 2018.                       |                                                              |
|                |                                       |                                                              |
|                | Instalment revival Request Approved   |                                                              |
|                | on 10 August 2018.                    |                                                              |
|                |                                       |                                                              |
|                | Instalment amount = 1000              |                                                              |
|                |                                       |                                                              |
|                | Taxes on unpaid premiums = 300        |                                                              |
|                |                                       |                                                              |
|                | Collection done on 11 August =        |                                                              |
|                | Rs.1300                               |                                                              |
|                |                                       |                                                              |
|                | Number of instalments = 5             |                                                              |
|                |                                       |                                                              |
|                | Surrender Request Indexed on 10       |                                                              |
|                | October 2018                          |                                                              |
+----------------+---------------------------------------+--------------------------------------------------------------+
| IR_18          | **[Surrender/Death Claim/Maturity     | Surrender indexing is not allowed till all the instalments   |
|                | Claim]{.underline}**                  | are paid.                                                    |
|                |                                       |                                                              |
|                | Policy number XXXX00011,              |                                                              |
|                |                                       |                                                              |
|                | Policy issuance date is 15-Mar-1993   |                                                              |
|                |                                       |                                                              |
|                | Paid to Date = 31 March 2017          |                                                              |
|                |                                       |                                                              |
|                | Product: CWLA                         |                                                              |
|                |                                       |                                                              |
|                | Policy Status = AL                    |                                                              |
|                |                                       |                                                              |
|                | Frequency = Monthly                   |                                                              |
|                |                                       |                                                              |
|                | Instalment Revival request indexed on |                                                              |
|                | 06 August 2018.                       |                                                              |
|                |                                       |                                                              |
|                | Instalment revival Request Approved   |                                                              |
|                | on 10 August 2018.                    |                                                              |
|                |                                       |                                                              |
|                | Collection not done yet               |                                                              |
|                |                                       |                                                              |
|                | Number of instalments = 5             |                                                              |
|                |                                       |                                                              |
|                | Surrender Request Indexed on 15       |                                                              |
|                | August 2018                           |                                                              |
+----------------+---------------------------------------+--------------------------------------------------------------+
| IR_19          | **[Conversion/Commutation             | User will not be allowed to index Conversion request. Error  |
|                | request]{.underline}**                | "Revival Instalments pending"                                |
|                |                                       |                                                              |
|                | Policy number XXXX00011,              |                                                              |
|                |                                       |                                                              |
|                | Policy issuance date is 15-Mar-1993   |                                                              |
|                |                                       |                                                              |
|                | Paid to Date = 31 March 2017          |                                                              |
|                |                                       |                                                              |
|                | Product: CWLA                         |                                                              |
|                |                                       |                                                              |
|                | Policy Status = AL                    |                                                              |
|                |                                       |                                                              |
|                | Frequency = Monthly                   |                                                              |
|                |                                       |                                                              |
|                | Instalment Revival request indexed on |                                                              |
|                | 06 August 2018.                       |                                                              |
|                |                                       |                                                              |
|                | Instalment revival Request Approved   |                                                              |
|                | on 10 August 2018.                    |                                                              |
|                |                                       |                                                              |
|                | Collection done on 11 August 2018     |                                                              |
|                |                                       |                                                              |
|                | Number of instalments = 5             |                                                              |
|                |                                       |                                                              |
|                | User tries to Index Conversion        |                                                              |
|                | request as on 15 August 2018          |                                                              |
+----------------+---------------------------------------+--------------------------------------------------------------+
| IR_19          | **[Conversion/Commutation             | Conversion request will not get effective as instalment      |
|                | request]{.underline}**                | revival request is pending.                                  |
|                |                                       |                                                              |
|                | Policy number XXXX00011,              |                                                              |
|                |                                       |                                                              |
|                | Policy issuance date is 15-Mar-1993   |                                                              |
|                |                                       |                                                              |
|                | Conversion request Approved on 10     |                                                              |
|                | April 2016.                           |                                                              |
|                |                                       |                                                              |
|                | Paid to Date = 30 April 2016          |                                                              |
|                |                                       |                                                              |
|                | Policy Status = AL                    |                                                              |
|                |                                       |                                                              |
|                | Instalment Revival request Indexed on |                                                              |
|                | 10 Feb 2017.                          |                                                              |
|                |                                       |                                                              |
|                | Instalment revival request not yet    |                                                              |
|                | approved                              |                                                              |
|                |                                       |                                                              |
|                | Conversion effective date = 15 March  |                                                              |
|                | 2017                                  |                                                              |
+----------------+---------------------------------------+--------------------------------------------------------------+
| IR_19          | **[Conversion/Commutation             | Conversion request will not get effective as all the         |
|                | request]{.underline}**                | instalments are not paid.                                    |
|                |                                       |                                                              |
|                | Policy number XXXX00011,              |                                                              |
|                |                                       |                                                              |
|                | Policy issuance date is 15-Mar-1993   |                                                              |
|                |                                       |                                                              |
|                | Conversion request Approved on 10     |                                                              |
|                | April 2016.                           |                                                              |
|                |                                       |                                                              |
|                | Paid to Date = 30 April 2016          |                                                              |
|                |                                       |                                                              |
|                | Policy Status = AL                    |                                                              |
|                |                                       |                                                              |
|                | Instalment Revival request Approved   |                                                              |
|                | on 10 Jan 2017.                       |                                                              |
|                |                                       |                                                              |
|                | Policyholder opted for 5 instalments. |                                                              |
|                |                                       |                                                              |
|                | Conversion effective date = 15 March  |                                                              |
|                | 2017                                  |                                                              |
|                |                                       |                                                              |
|                | Only 3 instalments are paid and 2 are |                                                              |
|                | pending                               |                                                              |
+----------------+---------------------------------------+--------------------------------------------------------------+
| IR_20          | **[Billing Method                     | Billing Method change request will allowed to be indexed.    |
|                | change]{.underline}**                 |                                                              |
|                |                                       |                                                              |
|                | Policy number XXXX00011               |                                                              |
|                |                                       |                                                              |
|                | Policy issuance date is 15-Mar-1993   |                                                              |
|                |                                       |                                                              |
|                | Paid to Date = 31 March 2017          |                                                              |
|                |                                       |                                                              |
|                | Product: CWLA                         |                                                              |
|                |                                       |                                                              |
|                | Policy Status = AL                    |                                                              |
|                |                                       |                                                              |
|                | Frequency = Monthly                   |                                                              |
|                |                                       |                                                              |
|                | Instalment Revival request indexed on |                                                              |
|                | 06 August 2018.                       |                                                              |
|                |                                       |                                                              |
|                | Instalment revival Request Approved   |                                                              |
|                | on 10 August 2018.                    |                                                              |
|                |                                       |                                                              |
|                | Collection done on 11 August 2018     |                                                              |
|                |                                       |                                                              |
|                | Number of instalments = 5             |                                                              |
|                |                                       |                                                              |
|                | 2 instalments are paid                |                                                              |
|                |                                       |                                                              |
|                | Billing method change request indexed |                                                              |
|                | on 15 September 2018.                 |                                                              |
+----------------+---------------------------------------+--------------------------------------------------------------+
| IR_20          | **[Billing Method                     | Billing Method change request will allowed to be approved    |
|                | change]{.underline}**                 | and effective as on 15 Sep 2018.                             |
|                |                                       |                                                              |
|                | Policy number XXXX00011               |                                                              |
|                |                                       |                                                              |
|                | Policy issuance date is 15-Mar-1993   |                                                              |
|                |                                       |                                                              |
|                | Paid to Date = 31 March 2017          |                                                              |
|                |                                       |                                                              |
|                | Product: CWLA                         |                                                              |
|                |                                       |                                                              |
|                | Policy Status = AL                    |                                                              |
|                |                                       |                                                              |
|                | Frequency = Monthly                   |                                                              |
|                |                                       |                                                              |
|                | Instalment Revival request indexed on |                                                              |
|                | 06 August 2018.                       |                                                              |
|                |                                       |                                                              |
|                | Instalment revival Request Approved   |                                                              |
|                | on 10 August 2018.                    |                                                              |
|                |                                       |                                                              |
|                | Collection done on 11 August 2018     |                                                              |
|                |                                       |                                                              |
|                | Number of instalments = 5             |                                                              |
|                |                                       |                                                              |
|                | 2 instalments are paid                |                                                              |
|                |                                       |                                                              |
|                | Billing method change request         |                                                              |
|                | Approved on 15 September 2018.        |                                                              |
+----------------+---------------------------------------+--------------------------------------------------------------+
| IR_21          | **[Survival Claim]{.underline}**      | User is not allowed to index Survival Claim request.         |
|                |                                       |                                                              |
|                | Policy number XXXX00011               |                                                              |
|                |                                       |                                                              |
|                | Policy issuance date is 15-Mar-1993   |                                                              |
|                |                                       |                                                              |
|                | Paid to Date = 31 March 2017          |                                                              |
|                |                                       |                                                              |
|                | Product: CWLA                         |                                                              |
|                |                                       |                                                              |
|                | Policy Status = AL                    |                                                              |
|                |                                       |                                                              |
|                | Frequency = Monthly                   |                                                              |
|                |                                       |                                                              |
|                | Instalment Revival request indexed on |                                                              |
|                | 06 August 2018.                       |                                                              |
|                |                                       |                                                              |
|                | Instalment revival Request Approved   |                                                              |
|                | on 10 August 2018.                    |                                                              |
|                |                                       |                                                              |
|                | Collection done on 11 August 2018     |                                                              |
|                |                                       |                                                              |
|                | Number of instalments = 5             |                                                              |
|                |                                       |                                                              |
|                | 4 instalments are pending             |                                                              |
|                |                                       |                                                              |
|                | 2^nd^ Survival claim payment due date |                                                              |
|                | Sep 2018                              |                                                              |
+----------------+---------------------------------------+--------------------------------------------------------------+
| IR_21          | **[Survival Claim]{.underline}**      | User is allowed to index Survival Claim request on any date  |
|                |                                       | after Dec 2018 for the due of Sep 2018.                      |
|                | Policy number XXXX00011               |                                                              |
|                |                                       |                                                              |
|                | Policy issuance date is 15-Mar-1993   |                                                              |
|                |                                       |                                                              |
|                | Paid to Date = 31 March 2017          |                                                              |
|                |                                       |                                                              |
|                | Product: CWLA                         |                                                              |
|                |                                       |                                                              |
|                | Policy Status = AL                    |                                                              |
|                |                                       |                                                              |
|                | Frequency = Monthly                   |                                                              |
|                |                                       |                                                              |
|                | Instalment Revival request indexed on |                                                              |
|                | 06 August 2018.                       |                                                              |
|                |                                       |                                                              |
|                | Instalment revival Request Approved   |                                                              |
|                | on 10 August 2018.                    |                                                              |
|                |                                       |                                                              |
|                | Collection done on 11 August 2018     |                                                              |
|                |                                       |                                                              |
|                | Number of instalments = 5             |                                                              |
|                |                                       |                                                              |
|                | 4 instalments are pending             |                                                              |
|                |                                       |                                                              |
|                | 2^nd^ Survival claim payment due date |                                                              |
|                | Sep 2018                              |                                                              |
|                |                                       |                                                              |
|                | All the instalments are paid by 1 Dec |                                                              |
|                | 2018                                  |                                                              |
+----------------+---------------------------------------+--------------------------------------------------------------+
| IR_22          | **[Loan]{.underline}**                | Indexing of Loan Request will not be allowed as all the      |
|                |                                       | instalments are not paid. Error "Revival Instalments are     |
|                | Policy number XXXX00011,              | pending" will populate.                                      |
|                |                                       |                                                              |
|                | Policy issuance date is 15-Mar-1993   |                                                              |
|                |                                       |                                                              |
|                | Paid to Date = 31 March 2017          |                                                              |
|                |                                       |                                                              |
|                | Product: CWLA                         |                                                              |
|                |                                       |                                                              |
|                | Policy Status = AL                    |                                                              |
|                |                                       |                                                              |
|                | Frequency = Monthly                   |                                                              |
|                |                                       |                                                              |
|                | Instalment Revival request indexed on |                                                              |
|                | 06 August 2018.                       |                                                              |
|                |                                       |                                                              |
|                | Instalment revival Request Approved   |                                                              |
|                | on 10 August 2018.                    |                                                              |
|                |                                       |                                                              |
|                | Collection done on 11 August 2018     |                                                              |
|                |                                       |                                                              |
|                | Number of instalments = 5             |                                                              |
|                |                                       |                                                              |
|                | 2 instalments are paid                |                                                              |
|                |                                       |                                                              |
|                | User tries to Index Loan request      |                                                              |
+----------------+---------------------------------------+--------------------------------------------------------------+
| IR_23          | **[Policy Cancellation/Free look      | User will be allowed to index policy cancellation request.   |
|                | Cancellation]{.underline}**           |                                                              |
|                |                                       | Policy status will change to AL                              |
|                | Policy number XXXX00011,              |                                                              |
|                |                                       | Paid to Date will be 31 March 2017                           |
|                | Policy issuance date is 15-Mar-1993   |                                                              |
|                |                                       |                                                              |
|                | Paid to Date = 31 March 2017          |                                                              |
|                |                                       |                                                              |
|                | Product: CWLA                         |                                                              |
|                |                                       |                                                              |
|                | Policy Status = AL                    |                                                              |
|                |                                       |                                                              |
|                | Frequency = Monthly                   |                                                              |
|                |                                       |                                                              |
|                | Instalment Revival request indexed on |                                                              |
|                | 06 August 2018.                       |                                                              |
|                |                                       |                                                              |
|                | Instalment revival Request Approved   |                                                              |
|                | on 10 August 2018.                    |                                                              |
|                |                                       |                                                              |
|                | Instalment amount = 1000              |                                                              |
|                |                                       |                                                              |
|                | Taxes on unpaid                       |                                                              |
|                |                                       |                                                              |
|                | Collection done on 11 August 2018 =   |                                                              |
|                | 1300                                  |                                                              |
|                |                                       |                                                              |
|                | Number of instalments = 5             |                                                              |
|                |                                       |                                                              |
|                | 2 instalments are paid                |                                                              |
|                |                                       |                                                              |
|                | User tries to index Policy            |                                                              |
|                | Cancellation request as on 15 Sep     |                                                              |
|                | 2018                                  |                                                              |
+----------------+---------------------------------------+--------------------------------------------------------------+
| IR_25          | **[Letters]{.underline}**             | After approval below letters will be generated:              |
|                |                                       |                                                              |
|                | Policy number XXXX00011,              | - Instalment revival acceptance letter as on 06 August 2018  |
|                |                                       |                                                              |
|                | Policy issuance date is 15-Mar-1993   | - Instalment revival Letter will be generated as on 16       |
|                |                                       |   August 2018                                                |
|                | Paid to Date = 31 March 2017          |                                                              |
|                |                                       |                                                              |
|                | Product: CWLA                         |                                                              |
|                |                                       |                                                              |
|                | Policy Status = AL                    |                                                              |
|                |                                       |                                                              |
|                | Premium Frequency = Semi Annual       |                                                              |
|                |                                       |                                                              |
|                | Instalment Amount = 1000              |                                                              |
|                |                                       |                                                              |
|                | Taxes on unpaid premium = 300         |                                                              |
|                |                                       |                                                              |
|                | Customer opted 3 instalments          |                                                              |
|                |                                       |                                                              |
|                | Instalment Revival request Approved   |                                                              |
|                | on 06 August 2018.                    |                                                              |
|                |                                       |                                                              |
|                | First instalment paid = 1300 as on 16 |                                                              |
|                | August 2018                           |                                                              |
+----------------+---------------------------------------+--------------------------------------------------------------+
| IR_25          | **[Letters]{.underline}**             | Reinstatement Rejection Letter will be generated as on 06    |
|                |                                       | August 2018                                                  |
|                | Policy number XXXX00011,              |                                                              |
|                |                                       |                                                              |
|                | Policy issuance date is 15-Mar-1993   |                                                              |
|                |                                       |                                                              |
|                | Paid to Date = 31 March 2017          |                                                              |
|                |                                       |                                                              |
|                | Product: CWLA                         |                                                              |
|                |                                       |                                                              |
|                | Policy Status = AL                    |                                                              |
|                |                                       |                                                              |
|                | Premium Frequency = Semi Annual       |                                                              |
|                |                                       |                                                              |
|                | Instalment Amount = 1000              |                                                              |
|                |                                       |                                                              |
|                | Taxes on unpaid premium = 300         |                                                              |
|                |                                       |                                                              |
|                | Customer opted 3 instalments          |                                                              |
|                |                                       |                                                              |
|                | Instalment Revival request Rejected   |                                                              |
|                | on 06 August 2018.                    |                                                              |
+----------------+---------------------------------------+--------------------------------------------------------------+
| IR_26          | **[Rebate]{.underline}**              | No Rebate will be given on advance instalment payment.       |
|                |                                       |                                                              |
|                | Policy number XXXX00011,              |                                                              |
|                |                                       |                                                              |
|                | Policy issuance date is 15-Mar-1993   |                                                              |
|                |                                       |                                                              |
|                | Paid to Date = 31 March 2017          |                                                              |
|                |                                       |                                                              |
|                | Product: CWLA                         |                                                              |
|                |                                       |                                                              |
|                | Policy Status = AL                    |                                                              |
|                |                                       |                                                              |
|                | Premium Frequency = Semi Annual       |                                                              |
|                |                                       |                                                              |
|                | Instalment Amount = 1000              |                                                              |
|                |                                       |                                                              |
|                | Taxes on unpaid premium = 300         |                                                              |
|                |                                       |                                                              |
|                | Customer opted 3 instalments          |                                                              |
|                |                                       |                                                              |
|                | Instalment Revival request Approved   |                                                              |
|                | on 06 August 2018.                    |                                                              |
|                |                                       |                                                              |
|                | First Instalment amount paid = 2000   |                                                              |
+----------------+---------------------------------------+--------------------------------------------------------------+
| IR_27          | **[Receipt                            | First Instalment need to be paid within 60 days from the     |
|                | Cancellation]{.underline}**           | date of Approval of the request.                             |
|                |                                       |                                                              |
|                | Policy number XXXX00011,              |                                                              |
|                |                                       |                                                              |
|                | Policy issuance date is 15-Mar-1993   |                                                              |
|                |                                       |                                                              |
|                | Paid to Date = 31 March 2017          |                                                              |
|                |                                       |                                                              |
|                | Product: CWLA                         |                                                              |
|                |                                       |                                                              |
|                | Policy Status = AL                    |                                                              |
|                |                                       |                                                              |
|                | Premium Frequency = Semi Annual       |                                                              |
|                |                                       |                                                              |
|                | Instalment Amount = 1000              |                                                              |
|                |                                       |                                                              |
|                | Taxes on unpaid premium = 300         |                                                              |
|                |                                       |                                                              |
|                | Customer opted 3 instalments          |                                                              |
|                |                                       |                                                              |
|                | Instalment Revival request Approved   |                                                              |
|                | on 06 August 2018.                    |                                                              |
|                |                                       |                                                              |
|                | First Instalment amount paid = 2000   |                                                              |
|                | as on 16 August 2018                  |                                                              |
|                |                                       |                                                              |
|                | Receipt Cancelled on 16 August 2018   |                                                              |
+----------------+---------------------------------------+--------------------------------------------------------------+
| IR_27          | **[Receipt                            | Policy status will be AL and Paid to date will be 31 March   |
|                | Cancellation]{.underline}**           | 2017 and amount thus collected as first instalment will move |
|                |                                       | to suspense                                                  |
|                | Policy number XXXX00011,              |                                                              |
|                |                                       |                                                              |
|                | Policy issuance date is 15-Mar-1993   |                                                              |
|                |                                       |                                                              |
|                | Paid to Date = 31 March 2017          |                                                              |
|                |                                       |                                                              |
|                | Product: CWLA                         |                                                              |
|                |                                       |                                                              |
|                | Policy Status = AL                    |                                                              |
|                |                                       |                                                              |
|                | Premium Frequency = Semi Annual       |                                                              |
|                |                                       |                                                              |
|                | Instalment Amount = 1000              |                                                              |
|                |                                       |                                                              |
|                | Taxes on unpaid premium = 300         |                                                              |
|                |                                       |                                                              |
|                | Customer opted 3 instalments          |                                                              |
|                |                                       |                                                              |
|                | Instalment Revival request Approved   |                                                              |
|                | on 06 August 2018.                    |                                                              |
|                |                                       |                                                              |
|                | First Instalment amount paid = 2000   |                                                              |
|                | as on 16 August 2018                  |                                                              |
|                |                                       |                                                              |
|                | Premium amount collected as on Oct    |                                                              |
|                | 2018                                  |                                                              |
|                |                                       |                                                              |
|                | Receipt of Premium Cancelled on 1 Oct |                                                              |
|                | 2018                                  |                                                              |
|                |                                       |                                                              |
|                | Premium Collection not done till 31   |                                                              |
|                | Oct 2018                              |                                                              |
+----------------+---------------------------------------+--------------------------------------------------------------+
| IR_28          | **[Suspense Reversal]{.underline}**   | Error "Suspense reversal not allowed for first collection"   |
|                |                                       | will be thrown.                                              |
|                | Policy number XXXX00011,              |                                                              |
|                |                                       |                                                              |
|                | Policy issuance date is 15-Mar-1993   |                                                              |
|                |                                       |                                                              |
|                | Paid to Date = 31 March 2017          |                                                              |
|                |                                       |                                                              |
|                | Product: CWLA                         |                                                              |
|                |                                       |                                                              |
|                | Policy Status = AL                    |                                                              |
|                |                                       |                                                              |
|                | Premium Frequency = Semi Annual       |                                                              |
|                |                                       |                                                              |
|                | Instalment Amount = 1000              |                                                              |
|                |                                       |                                                              |
|                | Taxes on unpaid premium = 300         |                                                              |
|                |                                       |                                                              |
|                | Customer opted 3 instalments          |                                                              |
|                |                                       |                                                              |
|                | Instalment Revival request Approved   |                                                              |
|                | on 06 August 2018.                    |                                                              |
|                |                                       |                                                              |
|                | First Instalment amount paid = 2000   |                                                              |
|                | as on 16 August 2018                  |                                                              |
|                |                                       |                                                              |
|                | User tries to do suspense reversal on |                                                              |
|                | 20 August 2018                        |                                                              |
+----------------+---------------------------------------+--------------------------------------------------------------+
| IR_28          | **[Suspense Reversal]{.underline}**   | Error "Suspense reversal not allowed for first collection"   |
|                |                                       | will be thrown.                                              |
|                | Policy number XXXX00011,              |                                                              |
|                |                                       |                                                              |
|                | Policy issuance date is 15-Mar-1993   |                                                              |
|                |                                       |                                                              |
|                | Paid to Date = 31 March 2017          |                                                              |
|                |                                       |                                                              |
|                | Product: CWLA                         |                                                              |
|                |                                       |                                                              |
|                | Policy Status = AL                    |                                                              |
|                |                                       |                                                              |
|                | Premium Frequency = Semi Annual       |                                                              |
|                |                                       |                                                              |
|                | Instalment Amount = 1000              |                                                              |
|                |                                       |                                                              |
|                | Taxes on unpaid premium = 300         |                                                              |
|                |                                       |                                                              |
|                | Customer opted 3 instalments          |                                                              |
|                |                                       |                                                              |
|                | Instalment Revival request Approved   |                                                              |
|                | on 06 August 2018.                    |                                                              |
|                |                                       |                                                              |
|                | First Instalment amount paid = 2000   |                                                              |
|                | as on 16 August 2018                  |                                                              |
|                |                                       |                                                              |
|                | Suspense Reversal (due to data entry  |                                                              |
|                | error) on 20 August 2018              |                                                              |
+----------------+---------------------------------------+--------------------------------------------------------------+
| IR_28          | **[Suspense Reversal]{.underline}**   | Policy status will be AL and paid to date will be 31 March   |
|                |                                       | 2017 and amount paid towards first instalment will move to   |
|                | Policy number XXXX00011,              | suspense                                                     |
|                |                                       |                                                              |
|                | Policy issuance date is 15-Mar-1993   |                                                              |
|                |                                       |                                                              |
|                | Paid to Date = 31 March 2017          |                                                              |
|                |                                       |                                                              |
|                | Product: CWLA                         |                                                              |
|                |                                       |                                                              |
|                | Policy Status = AL                    |                                                              |
|                |                                       |                                                              |
|                | Premium Frequency = Semi Annual       |                                                              |
|                |                                       |                                                              |
|                | Instalment Amount = 1000              |                                                              |
|                |                                       |                                                              |
|                | Taxes on unpaid premium = 300         |                                                              |
|                |                                       |                                                              |
|                | Customer opted 3 instalments          |                                                              |
|                |                                       |                                                              |
|                | Instalment Revival request Approved   |                                                              |
|                | on 06 August 2018.                    |                                                              |
|                |                                       |                                                              |
|                | First Instalment amount paid = 2000   |                                                              |
|                | as on 16 August 2018                  |                                                              |
|                |                                       |                                                              |
|                | 2^nd^ instalment collected on 1 Sep   |                                                              |
|                | 2018                                  |                                                              |
|                |                                       |                                                              |
|                | Suspense Reversal of 2^nd^ instalment |                                                              |
|                | on 20 Sep 2018                        |                                                              |
+----------------+---------------------------------------+--------------------------------------------------------------+
| IR_29          | **[Maximum number of                  | User will be allowed to index the Revival request as only    |
|                | Revivals]{.underline}**               | one revival request has been completed.                      |
|                |                                       |                                                              |
|                | Policy number XXXX00011,              |                                                              |
|                |                                       |                                                              |
|                | Policy issuance date is 15-Mar-1993   |                                                              |
|                |                                       |                                                              |
|                | Paid to Date = 31 March 2017          |                                                              |
|                |                                       |                                                              |
|                | Product: CWLA                         |                                                              |
|                |                                       |                                                              |
|                | Policy Status = AL                    |                                                              |
|                |                                       |                                                              |
|                | Premium Frequency = Semi Annual       |                                                              |
|                |                                       |                                                              |
|                | Instalment Amount = 1000              |                                                              |
|                |                                       |                                                              |
|                | Taxes on unpaid premium = 300         |                                                              |
|                |                                       |                                                              |
|                | Customer opted 3 instalments          |                                                              |
|                |                                       |                                                              |
|                | Instalment Revival request Approved   |                                                              |
|                | on 06 August 2018.                    |                                                              |
|                |                                       |                                                              |
|                | First Instalment amount paid = 2000   |                                                              |
|                | as on 16 August 2018                  |                                                              |
|                |                                       |                                                              |
|                | All the instalments paid by user by 1 |                                                              |
|                | Oct 2018                              |                                                              |
|                |                                       |                                                              |
|                | User tries to index Revival request   |                                                              |
|                | on 1 Jan 2020                         |                                                              |
+----------------+---------------------------------------+--------------------------------------------------------------+
| IR_29          | **[Maximum number of                  | User will be allowed to Index Revival request                |
|                | Revivals]{.underline}**               |                                                              |
|                |                                       |                                                              |
|                | Policy number XXXX00011,              |                                                              |
|                |                                       |                                                              |
|                | Policy issuance date is 15-Mar-1993   |                                                              |
|                |                                       |                                                              |
|                | Paid to Date = 31 March 2017          |                                                              |
|                |                                       |                                                              |
|                | Product: CWLA                         |                                                              |
|                |                                       |                                                              |
|                | Policy Status = AL                    |                                                              |
|                |                                       |                                                              |
|                | Premium Frequency = Semi Annual       |                                                              |
|                |                                       |                                                              |
|                | Instalment Amount = 1000              |                                                              |
|                |                                       |                                                              |
|                | Taxes on unpaid premium = 300         |                                                              |
|                |                                       |                                                              |
|                | Customer opted 3 instalments          |                                                              |
|                |                                       |                                                              |
|                | [1^st^ Revival]{.underline}           |                                                              |
|                |                                       |                                                              |
|                | Instalment Revival request Approved   |                                                              |
|                | on 06 August 2018.                    |                                                              |
|                |                                       |                                                              |
|                | First Instalment amount paid = 2000   |                                                              |
|                | as on 16 August 2018                  |                                                              |
|                |                                       |                                                              |
|                | All the instalments paid by user by 1 |                                                              |
|                | Oct 2018 and policy status is AP      |                                                              |
|                |                                       |                                                              |
|                | [2^nd^ Revival]{.underline}           |                                                              |
|                |                                       |                                                              |
|                | User tries to index Revival request   |                                                              |
|                | on 1 Jan 2020                         |                                                              |
|                |                                       |                                                              |
|                | Collection completed on 15 Jan 2020   |                                                              |
|                | and Policy status changed to AP.      |                                                              |
|                |                                       |                                                              |
|                | [3^rd^ Reviva]{.underline}l           |                                                              |
|                |                                       |                                                              |
|                | Policy status is AL on 20 Oct 2022    |                                                              |
|                |                                       |                                                              |
|                | User tries to Index Instalment        |                                                              |
|                | revival Request on 1 Dec 2022.        |                                                              |
+----------------+---------------------------------------+--------------------------------------------------------------+
| IR_29          | **[Maximum number of                  | User will be allowed to Index Revival request                |
|                | Revivals]{.underline}**               |                                                              |
|                |                                       |                                                              |
|                | Policy number XXXX00011,              |                                                              |
|                |                                       |                                                              |
|                | Policy issuance date is 15-Mar-1993   |                                                              |
|                |                                       |                                                              |
|                | Paid to Date = 31 March 2017          |                                                              |
|                |                                       |                                                              |
|                | Product: CWLA                         |                                                              |
|                |                                       |                                                              |
|                | Policy Status = AL                    |                                                              |
|                |                                       |                                                              |
|                | Premium Frequency = Semi Annual       |                                                              |
|                |                                       |                                                              |
|                | Instalment Amount = 1000              |                                                              |
|                |                                       |                                                              |
|                | Taxes on unpaid premium = 300         |                                                              |
|                |                                       |                                                              |
|                | Customer opted 3 instalments          |                                                              |
|                |                                       |                                                              |
|                | [1^st^ Revival]{.underline}           |                                                              |
|                |                                       |                                                              |
|                | Instalment Revival request Approved   |                                                              |
|                | on 06 August 2018.                    |                                                              |
|                |                                       |                                                              |
|                | First Instalment amount paid = 2000   |                                                              |
|                | as on 16 August 2018                  |                                                              |
|                |                                       |                                                              |
|                | All the instalments paid by user by 1 |                                                              |
|                | Oct 2018 and policy status is AP      |                                                              |
|                |                                       |                                                              |
|                | [2^nd^ Revival]{.underline}           |                                                              |
|                |                                       |                                                              |
|                | User tries to index Revival request   |                                                              |
|                | on 1 Jan 2020                         |                                                              |
|                |                                       |                                                              |
|                | Collection completed on 15 Jan 2020   |                                                              |
|                | and Policy status changed to AP.      |                                                              |
|                |                                       |                                                              |
|                | All the instalment are not paid and   |                                                              |
|                | policy status changes to AL on 1      |                                                              |
|                | April 2020                            |                                                              |
|                |                                       |                                                              |
|                | [3^rd^ Reviva]{.underline}l           |                                                              |
|                |                                       |                                                              |
|                | User tries to Index Instalment        |                                                              |
|                | revival Request on 20 Oct 2020        |                                                              |
+----------------+---------------------------------------+--------------------------------------------------------------+
| IR_30          | **[Receipt]{.underline}**             | Receipt will be generated as on 16 August 2018 with amount   |
|                |                                       | 2000                                                         |
|                | Policy number XXXX00011,              |                                                              |
|                |                                       |                                                              |
|                | Policy issuance date is 15-Mar-1993   |                                                              |
|                |                                       |                                                              |
|                | Paid to Date = 31 March 2017          |                                                              |
|                |                                       |                                                              |
|                | Product: CWLA                         |                                                              |
|                |                                       |                                                              |
|                | Policy Status = AL                    |                                                              |
|                |                                       |                                                              |
|                | Premium Frequency = Semi Annual       |                                                              |
|                |                                       |                                                              |
|                | Instalment Amount = 1000              |                                                              |
|                |                                       |                                                              |
|                | Taxes on unpaid premium = 300         |                                                              |
|                |                                       |                                                              |
|                | Customer opted 3 instalments          |                                                              |
|                |                                       |                                                              |
|                | Instalment Revival request Approved   |                                                              |
|                | on 06 August 2018.                    |                                                              |
|                |                                       |                                                              |
|                | First Instalment amount paid = 2000   |                                                              |
|                | as on 16 August 2018                  |                                                              |
+----------------+---------------------------------------+--------------------------------------------------------------+
| IR_32          | **[Non-Financial]{.underline}**       | User will be allowed to index Address Change Request         |
|                |                                       |                                                              |
|                | Policy number XXXX00011,              |                                                              |
|                |                                       |                                                              |
|                | Policy issuance date is 15-Mar-1993   |                                                              |
|                |                                       |                                                              |
|                | Paid to Date = 31 March 2017          |                                                              |
|                |                                       |                                                              |
|                | Product: CWLA                         |                                                              |
|                |                                       |                                                              |
|                | Policy Status = AL                    |                                                              |
|                |                                       |                                                              |
|                | Premium Frequency = Semi Annual       |                                                              |
|                |                                       |                                                              |
|                | Instalment Amount = 1000              |                                                              |
|                |                                       |                                                              |
|                | Taxes on unpaid premium = 300         |                                                              |
|                |                                       |                                                              |
|                | Customer opted 3 instalments          |                                                              |
|                |                                       |                                                              |
|                | Instalment Revival request Approved   |                                                              |
|                | on 06 August 2018.                    |                                                              |
|                |                                       |                                                              |
|                | First Instalment amount paid = 2000   |                                                              |
|                | as on 16 August 2018                  |                                                              |
|                |                                       |                                                              |
|                | User tries to Index Address Change    |                                                              |
|                | request as on 20 August 2018          |                                                              |
+----------------+---------------------------------------+--------------------------------------------------------------+
| IR_32          | **[Non-Financial]{.underline}**       | User will be allowed to Approve Address Change Request even  |
|                |                                       | if all the instalments are not paid.                         |
|                | Policy number XXXX00011,              |                                                              |
|                |                                       |                                                              |
|                | Policy issuance date is 15-Mar-1993   |                                                              |
|                |                                       |                                                              |
|                | Paid to Date = 31 March 2017          |                                                              |
|                |                                       |                                                              |
|                | Product: CWLA                         |                                                              |
|                |                                       |                                                              |
|                | Policy Status = AL                    |                                                              |
|                |                                       |                                                              |
|                | Premium Frequency = Semi Annual       |                                                              |
|                |                                       |                                                              |
|                | Instalment Amount = 1000              |                                                              |
|                |                                       |                                                              |
|                | Taxes on unpaid premium = 300         |                                                              |
|                |                                       |                                                              |
|                | Customer opted 3 instalments          |                                                              |
|                |                                       |                                                              |
|                | Instalment Revival request Approved   |                                                              |
|                | on 06 August 2018.                    |                                                              |
|                |                                       |                                                              |
|                | First Instalment amount paid = 2000   |                                                              |
|                | as on 16 August 2018                  |                                                              |
|                |                                       |                                                              |
|                | User tries to Approve Address Change  |                                                              |
|                | request as on 20 August 2018          |                                                              |
+----------------+---------------------------------------+--------------------------------------------------------------+
| IR_32          | **[Billing Frequency                  | User will not be allowed to index billing frequency change   |
|                | Change]{.underline}**                 | request until all the instalments are paid.                  |
|                |                                       |                                                              |
|                | Policy number XXXX00011,              |                                                              |
|                |                                       |                                                              |
|                | Policy issuance date is 15-Mar-1993   |                                                              |
|                |                                       |                                                              |
|                | Paid to Date = 31 March 2017          |                                                              |
|                |                                       |                                                              |
|                | Product: CWLA                         |                                                              |
|                |                                       |                                                              |
|                | Policy Status = AL                    |                                                              |
|                |                                       |                                                              |
|                | Premium Frequency = Semi Annual       |                                                              |
|                |                                       |                                                              |
|                | Instalment Amount = 1000              |                                                              |
|                |                                       |                                                              |
|                | Taxes on unpaid premium = 300         |                                                              |
|                |                                       |                                                              |
|                | Customer opted 3 instalments          |                                                              |
|                |                                       |                                                              |
|                | Instalment Revival request Approved   |                                                              |
|                | on 06 August 2018.                    |                                                              |
|                |                                       |                                                              |
|                | First Instalment amount paid = 2000   |                                                              |
|                | as on 16 August 2018                  |                                                              |
|                |                                       |                                                              |
|                | User tries to Index Billing frequency |                                                              |
|                | Change request as on 20 August 2018   |                                                              |
+----------------+---------------------------------------+--------------------------------------------------------------+
| IR_34          | **[Collection]{.underline}**          | 1300 will be collected through Instalment Revival screen.    |
|                |                                       |                                                              |
|                | Policy number XXXX00011,              | 520 will be collected through renewal screen.                |
|                |                                       |                                                              |
|                | Policy issuance date is 15-Mar-1993   |                                                              |
|                |                                       |                                                              |
|                | Paid to Date = 31 March 2017          |                                                              |
|                |                                       |                                                              |
|                | Product: CWLA                         |                                                              |
|                |                                       |                                                              |
|                | Policy Status = AL                    |                                                              |
|                |                                       |                                                              |
|                | Premium Frequency = Semi Annual       |                                                              |
|                |                                       |                                                              |
|                | Instalment Amount = 1000              |                                                              |
|                |                                       |                                                              |
|                | Taxes on unpaid premium = 300         |                                                              |
|                |                                       |                                                              |
|                | Premium amount = 500                  |                                                              |
|                |                                       |                                                              |
|                | Taxes on Premium = 20                 |                                                              |
|                |                                       |                                                              |
|                | Customer opted 3 instalments          |                                                              |
|                |                                       |                                                              |
|                | Instalment Revival request Approved   |                                                              |
|                | on 06 August 2018.                    |                                                              |
|                |                                       |                                                              |
|                | Collection of 1820 done on 16 August  |                                                              |
|                | 2018                                  |                                                              |
+----------------+---------------------------------------+--------------------------------------------------------------+
| IR_34          | **[Collection]{.underline}**          | User will not be allowed to make instalment payment on 12    |
|                |                                       | August 2018 as required premium (incl. taxes) are not paid   |
|                | Policy number XXXX00011,              | yet which is required for completion of collection stage.    |
|                |                                       |                                                              |
|                | Policy issuance date is 15-Mar-1993   | Error "Revival request not completed-due premium not paid"   |
|                |                                       | will be thrown.                                              |
|                | Paid to Date = 31 March 2017          |                                                              |
|                |                                       |                                                              |
|                | Product: CWLA                         |                                                              |
|                |                                       |                                                              |
|                | Policy Status = AL                    |                                                              |
|                |                                       |                                                              |
|                | Premium Frequency = Semi Annual       |                                                              |
|                |                                       |                                                              |
|                | Instalment Amount = 1000              |                                                              |
|                |                                       |                                                              |
|                | Taxes on unpaid premium = 300         |                                                              |
|                |                                       |                                                              |
|                | Premium amount = 500                  |                                                              |
|                |                                       |                                                              |
|                | Taxes on Premium = 20                 |                                                              |
|                |                                       |                                                              |
|                | Customer opted 3 instalments          |                                                              |
|                |                                       |                                                              |
|                | Instalment Revival request Approved   |                                                              |
|                | on 06 August 2018.                    |                                                              |
|                |                                       |                                                              |
|                | 1300 instalment amount paid through   |                                                              |
|                | instalment revival screen on 10       |                                                              |
|                | August 2018.                          |                                                              |
|                |                                       |                                                              |
|                | Premium (incl. taxes) of 520 is not   |                                                              |
|                | paid yet.                             |                                                              |
|                |                                       |                                                              |
|                | User tries to pay instalment amount   |                                                              |
|                | again on 12 August 2018               |                                                              |
+----------------+---------------------------------------+--------------------------------------------------------------+
| IR_34          | **[Collection]{.underline}**          | User will not be allowed to make premium payment on 12       |
|                |                                       | August 2018 as first instalment amount is not paid yet which |
|                | Policy number XXXX00011,              | is required for completion of collection stage.              |
|                |                                       |                                                              |
|                | Policy issuance date is 15-Mar-1993   | Error "Revival request not completed-revival instalment      |
|                |                                       | payment is pending" will be thrown.                          |
|                | Paid to Date = 31 March 2017          |                                                              |
|                |                                       |                                                              |
|                | Product: CWLA                         |                                                              |
|                |                                       |                                                              |
|                | Policy Status = AL                    |                                                              |
|                |                                       |                                                              |
|                | Premium Frequency = Semi Annual       |                                                              |
|                |                                       |                                                              |
|                | Instalment Amount = 1000              |                                                              |
|                |                                       |                                                              |
|                | Taxes on unpaid premium = 300         |                                                              |
|                |                                       |                                                              |
|                | Premium amount = 500                  |                                                              |
|                |                                       |                                                              |
|                | Taxes on Premium = 20                 |                                                              |
|                |                                       |                                                              |
|                | Customer opted 3 instalments          |                                                              |
|                |                                       |                                                              |
|                | Instalment Revival request Approved   |                                                              |
|                | on 06 August 2018.                    |                                                              |
|                |                                       |                                                              |
|                | Premium (incl. taxes) of 520 is paid  |                                                              |
|                | on 10 August 2018.                    |                                                              |
|                |                                       |                                                              |
|                | First instalment is not yet paid.     |                                                              |
|                |                                       |                                                              |
|                | User tries to pay Premium for next    |                                                              |
|                | due on 12 August 2018                 |                                                              |
+----------------+---------------------------------------+--------------------------------------------------------------+
| IR_34          | **[Collection]{.underline}**          | User will be allowed to make premium payment for next due as |
|                |                                       | policy status is AP.                                         |
|                | Policy number XXXX00011,              |                                                              |
|                |                                       |                                                              |
|                | Policy issuance date is 15-Mar-1993   |                                                              |
|                |                                       |                                                              |
|                | Paid to Date = 31 March 2017          |                                                              |
|                |                                       |                                                              |
|                | Product: CWLA                         |                                                              |
|                |                                       |                                                              |
|                | Policy Status = AL                    |                                                              |
|                |                                       |                                                              |
|                | Premium Frequency = Semi Annual       |                                                              |
|                |                                       |                                                              |
|                | Instalment Amount = 1000              |                                                              |
|                |                                       |                                                              |
|                | Taxes on unpaid premium = 300         |                                                              |
|                |                                       |                                                              |
|                | Premium amount = 500                  |                                                              |
|                |                                       |                                                              |
|                | Taxes on Premium = 20                 |                                                              |
|                |                                       |                                                              |
|                | Customer opted 3 instalments          |                                                              |
|                |                                       |                                                              |
|                | Instalment Revival request Approved   |                                                              |
|                | on 06 August 2018.                    |                                                              |
|                |                                       |                                                              |
|                | Collection of amount 1820 is done on  |                                                              |
|                | 10 August 2018                        |                                                              |
|                |                                       |                                                              |
|                | User tries to pay Premium for next    |                                                              |
|                | due on 12 August 2018                 |                                                              |
+----------------+---------------------------------------+--------------------------------------------------------------+
| IR_34          | **[Collection]{.underline}**          | User will be allowed to make advance instalment payment.     |
|                |                                       |                                                              |
|                | Policy number XXXX00011,              |                                                              |
|                |                                       |                                                              |
|                | Policy issuance date is 15-Mar-1993   |                                                              |
|                |                                       |                                                              |
|                | Paid to Date = 31 March 2017          |                                                              |
|                |                                       |                                                              |
|                | Product: CWLA                         |                                                              |
|                |                                       |                                                              |
|                | Policy Status = AL                    |                                                              |
|                |                                       |                                                              |
|                | Premium Frequency = Semi Annual       |                                                              |
|                |                                       |                                                              |
|                | Instalment Amount = 1000              |                                                              |
|                |                                       |                                                              |
|                | Taxes on unpaid premium = 300         |                                                              |
|                |                                       |                                                              |
|                | Premium amount = 500                  |                                                              |
|                |                                       |                                                              |
|                | Taxes on Premium = 20                 |                                                              |
|                |                                       |                                                              |
|                | Customer opted 3 instalments          |                                                              |
|                |                                       |                                                              |
|                | Instalment Revival request Approved   |                                                              |
|                | on 06 August 2018.                    |                                                              |
|                |                                       |                                                              |
|                | Collection of amount 1820 is done on  |                                                              |
|                | 10 August 2018                        |                                                              |
|                |                                       |                                                              |
|                | Collection stage = completed          |                                                              |
|                |                                       |                                                              |
|                | User tries to pay advance instalment  |                                                              |
|                | on 12 August 2018                     |                                                              |
+----------------+---------------------------------------+--------------------------------------------------------------+
| IR_35          | **[Withdrawal of Instalment Revival   | Revival request will allowed to be withdrawn as collection   |
|                | request]{.underline}**                | is not done yet.                                             |
|                |                                       |                                                              |
|                | Policy number XXXX00011,              | Request status will be withdrawn                             |
|                |                                       |                                                              |
|                | Policy issuance date is 15-Mar-1993   |                                                              |
|                |                                       |                                                              |
|                | Paid to Date = 31 March 2017          |                                                              |
|                |                                       |                                                              |
|                | Product: CWLA                         |                                                              |
|                |                                       |                                                              |
|                | Policy Status = AL                    |                                                              |
|                |                                       |                                                              |
|                | Premium Frequency = Semi Annual       |                                                              |
|                |                                       |                                                              |
|                | Instalment Amount = 1000              |                                                              |
|                |                                       |                                                              |
|                | Taxes on unpaid premium = 300         |                                                              |
|                |                                       |                                                              |
|                | Premium amount = 500                  |                                                              |
|                |                                       |                                                              |
|                | Taxes on Premium = 20                 |                                                              |
|                |                                       |                                                              |
|                | Customer opted 3 instalments          |                                                              |
|                |                                       |                                                              |
|                | Instalment Revival request Approved   |                                                              |
|                | on 06 August 2018.                    |                                                              |
|                |                                       |                                                              |
|                | Collection of amount 1820 is not done |                                                              |
|                | yet                                   |                                                              |
|                |                                       |                                                              |
|                | User tries to withdraw Revival        |                                                              |
|                | request.                              |                                                              |
+----------------+---------------------------------------+--------------------------------------------------------------+
| IR_35          | **[Withdrawal of Instalment Revival   | User will not be allowed to withdraw revival request.        |
|                | request]{.underline}**                |                                                              |
|                |                                       | Error "Revival request cannot be withdrawn as collection is  |
|                | Policy number XXXX00011,              | done" will be thrown.                                        |
|                |                                       |                                                              |
|                | Policy issuance date is 15-Mar-1993   |                                                              |
|                |                                       |                                                              |
|                | Paid to Date = 31 March 2017          |                                                              |
|                |                                       |                                                              |
|                | Product: CWLA                         |                                                              |
|                |                                       |                                                              |
|                | Policy Status = AL                    |                                                              |
|                |                                       |                                                              |
|                | Premium Frequency = Semi Annual       |                                                              |
|                |                                       |                                                              |
|                | Instalment Amount = 1000              |                                                              |
|                |                                       |                                                              |
|                | Taxes on unpaid premium = 300         |                                                              |
|                |                                       |                                                              |
|                | Premium amount = 500                  |                                                              |
|                |                                       |                                                              |
|                | Taxes on Premium = 20                 |                                                              |
|                |                                       |                                                              |
|                | Customer opted 3 instalments          |                                                              |
|                |                                       |                                                              |
|                | Instalment Revival request Approved   |                                                              |
|                | on 06 August 2018.                    |                                                              |
|                |                                       |                                                              |
|                | Collection of amount 1820 is done on  |                                                              |
|                | 10 August 2018                        |                                                              |
|                |                                       |                                                              |
|                | User tries to withdraw Revival        |                                                              |
|                | request on 15 August 2018             |                                                              |
+----------------+---------------------------------------+--------------------------------------------------------------+
| [IR_36]{.mark} | **[[Collection]{.underline}]{.mark}** | [Amount thus collected will move to suspense and policy will |
|                |                                       | move to AL status for not receiving Instalment on the due    |
|                | [Policy number XXXX00011,]{.mark}     | date.]{.mark}                                                |
|                |                                       |                                                              |
|                | [Policy issuance date is              |                                                              |
|                | 15-Mar-1993]{.mark}                   |                                                              |
|                |                                       |                                                              |
|                | [Paid to Date = 31 March 2017]{.mark} |                                                              |
|                |                                       |                                                              |
|                | [Product: CWLA]{.mark}                |                                                              |
|                |                                       |                                                              |
|                | [Policy Status = AL]{.mark}           |                                                              |
|                |                                       |                                                              |
|                | [Premium Frequency = Semi             |                                                              |
|                | Annual]{.mark}                        |                                                              |
|                |                                       |                                                              |
|                | [Instalment Amount = 1000]{.mark}     |                                                              |
|                |                                       |                                                              |
|                | [Taxes on unpaid premium =            |                                                              |
|                | 300]{.mark}                           |                                                              |
|                |                                       |                                                              |
|                | [Premium amount = 500]{.mark}         |                                                              |
|                |                                       |                                                              |
|                | [Taxes on Premium = 20]{.mark}        |                                                              |
|                |                                       |                                                              |
|                | [Customer opted 3 instalments]{.mark} |                                                              |
|                |                                       |                                                              |
|                | [Instalment Revival request Approved  |                                                              |
|                | on 06 August 2018.]{.mark}            |                                                              |
|                |                                       |                                                              |
|                | [Collection of 1820 (first instalment |                                                              |
|                | and premium) done on 16 August        |                                                              |
|                | 2018]{.mark}                          |                                                              |
|                |                                       |                                                              |
|                | [Collection of second instalment on   |                                                              |
|                | 1^st^ September 2018 done through     |                                                              |
|                | 'Renewal' transaction type instead of |                                                              |
|                | 'Instalment Revival'.]{.mark}         |                                                              |
+----------------+---------------------------------------+--------------------------------------------------------------+

***\***

# Screens and Navigation

> To accommodate the business rules for Instalment revival there would
> be few changes required on UI of Collection screen and Policy Search
> Screen.

- First collection amount consist of Premium + Instalment amount + Taxes
  on unpaid Premium + Taxes on Premium (where applicable). Premium +
  Taxes on premium to be collected through 'Renewal' screen.
  Instalment + taxes on unpaid premiums to be collected through
  'Instalment Revival' screen.

- Subsequent instalments will be collected through instalment revival
  screen

- Subsequent premium will be collected through Renewal Premium screen.
  From Date and To Date on renewal screen will be current frequency date

> [Example 2:]{.underline}
>
> Paid to Date = 31 March 2018
>
> Revival Indexing Done on 15 Dec 2018
>
> Premium Frequency = Quarterly
>
> From Date = 1 Oct 2018
>
> To Date will be 31Dec 2018
>
> The screens below show the navigation and changes:-

1.  **Collection Icon on dashboard:-**

![](media/image2.png){width="6.5in" height="2.4722222222222223in"}

On Clicking Collection Icon:-

One more transaction type will be added as Instalment Revival which will
be a radio button.

![](media/image3.png){width="6.846801181102363in"
height="1.9524518810148732in"}

On clicking the radio button for instalment revival below screen will
open-

![](media/image4.png){width="6.495833333333334in"
height="1.9305555555555556in"}

1.  **[Policy Search-View History-Collection Tab:-]{.mark}**

![](media/image5.png){width="6.5in" height="0.6895833333333333in"}

**[Data Elements]{.underline}**

+---------------+--------------+----------------+----------------------------+
| **Field       | **UI         | **Properties** | **Comments**               |
| Name**        | Display**    |                |                            |
+===============+==============+================+============================+
| **Installment | > **Label**  |                |                            |
| Revival**     |              |                |                            |
+---------------+--------------+----------------+----------------------------+
| Policy Number | Text box     | > Enabled      | Search the policy to make  |
|               |              |                | payment                    |
+---------------+--------------+----------------+----------------------------+
| **Installment | > **Label**  |                |                            |
| Revival       |              |                |                            |
| Details**     |              |                |                            |
+---------------+--------------+----------------+----------------------------+
| Policy Number | Alphanumeric | > Read Only    | Policy Number will be      |
|               |              |                | displayed                  |
+---------------+--------------+----------------+----------------------------+
| Insured Name  | Text box     | > Read Only    | Full Name of the Insured   |
|               |              |                | will be displayed          |
+---------------+--------------+----------------+----------------------------+
| Agent ID      | Alphanumeric | > Read Only    | Name of the Agent as       |
|               |              |                | displayed on the Renewal   |
|               |              |                | screen.                    |
+---------------+--------------+----------------+----------------------------+
| From (Date)   | Date         | > Read Only    | It will be the due date    |
|               |              |                | from which premiums were   |
|               |              |                | due.                       |
|               |              |                |                            |
|               |              |                | Example:                   |
|               |              |                |                            |
|               |              |                | Paid to Date 31 March 2018 |
|               |              |                |                            |
|               |              |                | Revival Indexing Done on 1 |
|               |              |                | Dec 2018                   |
|               |              |                |                            |
|               |              |                | Premium Frequency =        |
|               |              |                | Monthly                    |
|               |              |                |                            |
|               |              |                | From Date will be 1 April  |
|               |              |                | 2018                       |
+---------------+--------------+----------------+----------------------------+
| To (Date)\_   | Date         | > Enabled      | It will be 1 frequency     |
|               |              |                | before current Frequency.  |
|               |              |                | User has an option to      |
|               |              |                | change the 'To date' based |
|               |              |                | on the number of           |
|               |              |                | instalments paid as        |
|               |              |                | advance                    |
|               |              |                |                            |
|               |              |                | [Example 1:]{.underline}   |
|               |              |                |                            |
|               |              |                | Paid to Date 31 March 2018 |
|               |              |                |                            |
|               |              |                | Revival Indexing Done on   |
|               |              |                | 15 Dec 2018                |
|               |              |                |                            |
|               |              |                | Premium Frequency =        |
|               |              |                | Monthly                    |
|               |              |                |                            |
|               |              |                | To Date will be 30 Nov     |
|               |              |                | 2018 will auto populate    |
|               |              |                |                            |
|               |              |                | If number of instalments   |
|               |              |                | are 5 and each instalment  |
|               |              |                | is of amount 100. User     |
|               |              |                | paid 330RS. 30 is tax on   |
|               |              |                | unpaid premium.            |
|               |              |                |                            |
|               |              |                | User has option to change  |
|               |              |                | the 'To date' based on     |
|               |              |                | current collection date.   |
|               |              |                |                            |
|               |              |                | Current date = 20 Dec      |
|               |              |                | 2018.                      |
|               |              |                |                            |
|               |              |                | To Date can be changed to  |
|               |              |                | 28 Feb 2019.               |
|               |              |                |                            |
|               |              |                | 100 will be for Dec 2018   |
|               |              |                |                            |
|               |              |                | 100 will be for Jan 2019   |
|               |              |                |                            |
|               |              |                | 100 will be for Feb 2019   |
|               |              |                |                            |
|               |              |                | [Example 2:]{.underline}   |
|               |              |                |                            |
|               |              |                | Paid to Date 31 March 2018 |
|               |              |                |                            |
|               |              |                | Revival Indexing Done on   |
|               |              |                | 15 Dec 2018                |
|               |              |                |                            |
|               |              |                | Premium Frequency =        |
|               |              |                | Quarterly                  |
|               |              |                |                            |
|               |              |                | To Date will be 30 Sep     |
|               |              |                | 2018 will be auto          |
|               |              |                | populated. User has an     |
|               |              |                | option to change the 'To   |
|               |              |                | Date'.                     |
+---------------+--------------+----------------+----------------------------+
| Installment   | Numeric      | > Read Only    | Will display the           |
| revival       |              |                | instalment Amount          |
| amount        |              |                |                            |
+---------------+--------------+----------------+----------------------------+
| Taxes on      | Numeric      | > Read Only    | This field will display    |
| unpaid        |              |                | the taxes on total unpaid  |
| premium       |              |                | premium. It will be        |
|               |              |                | displayed only at the time |
|               |              |                | of first instalment.       |
|               |              |                |                            |
|               |              |                | From subsequent            |
|               |              |                | instalments this field     |
|               |              |                | will display Zero Value.   |
+---------------+--------------+----------------+----------------------------+
| Instalment    | Numeric      | > Read Only    | This field will display    |
| Received      |              |                | the amount available in    |
|               |              |                | Policy suspense against    |
|               |              |                | instalment as an extra     |
|               |              |                | amount.                    |
+---------------+--------------+----------------+----------------------------+
| Current       | Numeric      | > Read Only    | This will denote the       |
| instalment    |              |                | number of instalments      |
|               |              |                | customer has paid + the    |
|               |              |                | instalment he is paying    |
|               |              |                | currently.                 |
|               |              |                |                            |
|               |              |                | If user is covering future |
|               |              |                | instalments also, then it  |
|               |              |                | will be counted under      |
|               |              |                | current instalment.        |
|               |              |                |                            |
|               |              |                | Example:                   |
|               |              |                |                            |
|               |              |                | Total instalments are 5.   |
|               |              |                |                            |
|               |              |                | At the time of paying      |
|               |              |                | 2^nd^ instalment, user is  |
|               |              |                | paying for 3^rd^           |
|               |              |                | instalment also.           |
|               |              |                |                            |
|               |              |                | Then current instalment    |
|               |              |                | will display as 3.         |
+---------------+--------------+----------------+----------------------------+
| Total Number  | Numeric      | > Read Only    | This will denote the       |
| of            |              |                | number of instalments      |
| Instalments   |              |                | customer has opted at the  |
| opted         |              |                | time of instalment revival |
|               |              |                | request                    |
+---------------+--------------+----------------+----------------------------+
| Total Amount  | Numeric      | > Read Only    | The amount customer needs  |
|               |              |                | to pay based on the 'From  |
|               |              |                | and To date' + taxes on    |
|               |              |                | unpaid premium (at the     |
|               |              |                | time of first instalment)  |
|               |              |                |                            |
|               |              |                | [Example:]{.underline}     |
|               |              |                |                            |
|               |              |                | Paid to Date 31 March 2018 |
|               |              |                |                            |
|               |              |                | Revival Indexing Done on   |
|               |              |                | 15 Dec 2018                |
|               |              |                |                            |
|               |              |                | Premium Frequency =        |
|               |              |                | Monthly                    |
|               |              |                |                            |
|               |              |                | To Date will be 30 Nov     |
|               |              |                | 2018 will auto populate    |
|               |              |                |                            |
|               |              |                | User has option to change  |
|               |              |                | the 'To date' based on     |
|               |              |                | current collection date.   |
|               |              |                |                            |
|               |              |                | Current collection date =  |
|               |              |                | 20 Dec 2018.               |
|               |              |                |                            |
|               |              |                | To Date changed to 28 Feb  |
|               |              |                | 2019.                      |
|               |              |                |                            |
|               |              |                | Each instalment amount =   |
|               |              |                | 100                        |
|               |              |                |                            |
|               |              |                | Taxes on unpaid premium =  |
|               |              |                | 50                         |
|               |              |                |                            |
|               |              |                | Total Amount will be       |
|               |              |                | displayed as 350.          |
+---------------+--------------+----------------+----------------------------+
| Actual Amount | Text box     | > Enabled      | Total amount customer is   |
| paid          |              |                | paying will be entered     |
|               |              |                | here                       |
+---------------+--------------+----------------+----------------------------+

Error Messages:

+----------------------+-----------------------------------------------+
| **Error and Edit     | Description                                   |
| Messages**           |                                               |
+:=====================+:==============================================+
| This field is        | If Add is clicked without entering Policy     |
| required             | Number.                                       |
+----------------------+-----------------------------------------------+
| Policy Number not    | If incorrect policy number is entered.        |
| found                |                                               |
+----------------------+-----------------------------------------------+
| Instalment revival   | - If instalment Revival request is not        |
| not in process       |   indexed.                                    |
|                      |                                               |
|                      | - When normal Revival request is indexed.     |
|                      |                                               |
|                      | - If all the installments have been paid for  |
|                      |   the policy and the user is trying to add    |
|                      |   that policy for collection                  |
+----------------------+-----------------------------------------------+
| Error "Revival       | When user tries to index the                  |
| Instalments are      | conversion/commutation/Billing Method change  |
| pending"             | request when all the instalments are not      |
|                      | paid.                                         |
+----------------------+-----------------------------------------------+
| Instalment amount    | If user tries to change the 'To date' but the |
| insufficient         | instalment amount paid is not sufficient to   |
|                      | cover the 'To Date'                           |
+----------------------+-----------------------------------------------+

# Documents/Letters

  -----------------------------------------------------------------------
  Document                Trigger                 Template
  required/Letter                                 
  generated                                       
  ----------------------- ----------------------- -----------------------
                                                  

  Revival Instalment      Approval                ![](media/image6.emf)
  Acceptance Letter                               

                                                  

                                                  
  -----------------------------------------------------------------------

# 
