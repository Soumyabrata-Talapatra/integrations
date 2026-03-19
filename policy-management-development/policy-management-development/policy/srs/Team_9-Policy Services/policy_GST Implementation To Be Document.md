**DEPARTMENT OF POSTS**

**MINISTRY OF COMMUNICATIONS & IT**

**GOVERNMENT OF INDIA**

**System Requirements Specification (SRS)**

Changes related to Implementation of Goods and Services Tax (GST)

**Table of Content**

# Contents {#contents .TOC-Heading}

[1 Overview [2](#section)](#section)

[2 Applicability [2](#applicability)](#applicability)

[2.1 Applicable [2](#applicable)](#applicable)

[3 GST Changes: [4](#gst-changes)](#gst-changes)

[3.1 Premium Collection computation:
[4](#premium-collection-computation)](#premium-collection-computation)

[3.2 Premium Calculation Logic incorporating GST in System:
[4](#premium-calculation-logic-incorporating-gst-in-system)](#premium-calculation-logic-incorporating-gst-in-system)

[3.2.1 Changes in New Business Screen:
[5](#changes-in-new-business-screen)](#changes-in-new-business-screen)

[3.2.2 Illustrations [6](#illustrations)](#illustrations)

[3.3 Rejection/Termination of PROPOSAL:
[18](#rejectiontermination-of-proposal)](#rejectiontermination-of-proposal)

[3.4 Collection Screen Changes:
[19](#collection-screen-changes)](#collection-screen-changes)

[3.5 Claims Processing Screen Changes:
[19](#claims-processing-screen-changes)](#claims-processing-screen-changes)

[3.6 Policy History & Summary Screen Changes:
[20](#policy-history-summary-screen-changes)](#policy-history-summary-screen-changes)

[3.7 Meghdoot Upload/bulk upload:
[20](#meghdoot-uploadbulk-upload)](#meghdoot-uploadbulk-upload)

[3.8 Report Changes: [22](#report-changes)](#report-changes)

[3.9 GST Number Mapping: [22](#gst-number-mapping)](#gst-number-mapping)

[3.10 List where UTGST/SGST is applicable:
[22](#list-where-utgstsgst-is-applicable)](#list-where-utgstsgst-is-applicable)

#  {#section .Style2}

#  {#section-1 .Style2}

# **Overview** {#overview .Style2}

#  {#section-2 .Style2}

The purpose of the document is to outline the impact of Implementation
of Goods and Services Tax (GST)

- GST @ \_\_\_\_\_% of the gross premium (First year premium) and @
  \_\_\_\_\_\_\_ % of the gross premium (in subsequent years) to be
  collected from the policy holders and consolidated (Premium + GST)
  receipt to be issued.

# **Applicability** {#applicability .Style2}

## Applicable 

- The Scope of GST taxation changes

  - is applicable across PAN India

  - The rate of GST will be different for first year & subsequent policy
    years.

  - GST rate will have two components -- CGST, SGST/UTGST

  - CGST , SGST/UTGST rate will be defined at HO, SO, BO Level

- The changes done will be restricted to below modules and are mentioned
  in the document in detail:

  - Premium Calculation changes

  - Meghdoot/Bulk Upload changes

  - Accounting entries changes

  - Reports

  - Customer Portal Changes (including Screen level changes)

# **GST Changes:** {#gst-changes .Style2}

## **Premium Collection computation:**

Following collections will be impacted ONLY:

a.  New Business Collection

b.  Renewal Collection

c.  Revival Collection

## **Premium Calculation Logic incorporating GST in System:**

GST to be collected from customer will depend ONLY on:

- For the initial premiums, GST rate applicable will be of the office
  code of user who is login in the application and doing the indexing of
  the proposal.

- For Renewal Premiums, GST rate applicable will be of the office code
  of the user who is login in the application and doing the renewal
  collection.

- For Revival Premium Collection,

  - For Revival Quote:

> For Revival Quote, GST Rate applicable will be that of Policy Issuance
> office code.

- Revival Requests Indexed & not Approved (Quote Section of
  DE/QC/Approver)

> GST Rate applicable will be of the user office code who is login in
> the application and doing the revival request Indexing.

- Revival Collection:

> GST Rate applicable will be of the user office code who is login in
> the application and doing the revival collection.

- [If the customer is making the renewal premium payment through the
  Self Service Portal, then GST rate of the state where NARO is situated
  will be applicable.]{.mark}

- [If the proposal is indexed through the customer portal, then GST rate
  will be of the state where NARO is situated will be
  applicable.]{.mark}

- Issuance Office code (office code from where the policy has been
  issued) will be checked for GST rate and other rates.

- In case the indexing of the proposal or collection of the renewal
  collection is done other than HO, SO or BO, then error message "GST
  not applicable for offices other than HO, SO or BO".

- Accounting entry of suspense to be passed on the day of collection of
  initial premium and when proposal is accepted then, accounting entries
  are to be passed for collection of premium and tax against earlier
  suspense entry; and in case of rejection/ withdrawal/ termination of
  proposal, no tax is deducted out of initial premium.

- Year for which the premium is paid (First Year or Renewal Year)

  - First year premium refers to premium payable till the first policy
    anniversary.

  - Renewal premium refers to premium payable subsequent to the first
    policy anniversary.

- Effective date is the date of actual transaction ( Transaction date )

<!-- -->

- The GST has to be applied on the Premium Due Amount paid as per the
  paid to date selected by the user.

  - The GST computation will not consider the Interest in case of
    Revival/Reinstatement.

  - [The GST computation will consider rebate if any applicable.]{.mark}

- Actual Amount Paid =\> Total Amount as derived in Collection screen

- Total Amount =

a.  [Premium Due Amount --Rebate (If any)]{.mark}

b.  [Add GST (CGS + SGST/ UTGST) (to be rounded off to nearest rupee)
    on (a) above]{.mark}

c.  [Add Interest/default (if any applicable)]{.mark}

d.  [Subtract: Amount in suspense]{.mark}

- [For cases where through one receipt, premium for First Year as well
  as Renewal Year has been paid, then Rounding off GST will be done
  after adding First Year GST and Renewal Year GST and not
  separately]{.mark}

### 3.2.1 Changes in New Business Screen:

- On the UI wherever the labels reflect inclusive of taxes will include
  value inclusive of GST.

GST calculation rules for various scenarios are mentioned below:

+------+----------------------------+-----------------------------------+
| **Sr | **Scenario:**              | **Rule**                          |
| No** |                            |                                   |
+:====:+============================+===================================+
| 1    | Indexing /Collections/     | If the policy is backdated then   |
|      | Approval in the same month | on Approval Date, Service tax/GST |
|      | and issuance and approval  | Rate taken will be as of issuance |
|      | date is same.              | date.                             |
|      |                            |                                   |
|      |                            | If the approval date and the      |
|      |                            | issuance date is same, then       |
|      |                            | Service Tax/GST rate as on        |
|      |                            | approval date can be taken.       |
+------+----------------------------+-----------------------------------+
| 2    | Indexing / Collections/    | If the policy is backdated then   |
|      | Approval in the same month | on Approval Date, Service tax/GST |
|      | but there is a change in   | Rate taken will be as of issuance |
|      | tax rate.                  | date.                             |
|      |                            |                                   |
|      | Changes.                   | If the approval date and the      |
|      |                            | issuance date is same, then       |
|      |                            | Service Tax/GST rate as on        |
|      |                            | approval date can be taken        |
+------+----------------------------+-----------------------------------+
| 3    | Indexing & Approval in the | If the policy is backdated then   |
|      | different month and there  | on Approval Date, Service tax/GST |
|      | is a change in modal       | Rate taken will be as of issuance |
|      | premium without change in  | date.                             |
|      | tax rate.                  |                                   |
|      |                            | If the approval date and the      |
|      |                            | issuance date is same, then       |
|      |                            | Service Tax/GST rate as on        |
|      |                            | approval date can be taken        |
+------+----------------------------+-----------------------------------+
| 4    | Indexing & Approval in the | Service Tax/GST Rate at the time  |
|      | different month and there  | of Proposal issuance will be      |
|      | is a change in modal       | considered for tax calculation    |
|      | premium with change in tax | (either service tax/GST) on the   |
|      | rate.                      | changed '" Premium payable" (as   |
|      |                            | derived from modal premium) as    |
|      |                            | determined at the time of         |
|      |                            | approval.                         |
+------+----------------------------+-----------------------------------+
| 5    | Advance collection done &  | Service Tax/GST Rate at the time  |
|      | there is no change in tax  | of Proposal issuance will be      |
|      | rate                       | considered for tax calculation    |
|      |                            | (either service tax/GST) on the   |
|      |                            | "Amount Paid" (as derived from    |
|      |                            | modal premium and premium         |
|      |                            | payable) as determined at the     |
|      |                            | time of approval. Rebate will not |
|      |                            | be considered for initial premium |
+------+----------------------------+-----------------------------------+
| 6    | Advance collection done &  | If the policy is backdated then   |
|      | there is a change in tax   | on Approval Date, Rate taken will |
|      | rate                       | be as of issuance date.           |
|      |                            |                                   |
|      |                            | If the approval date and the      |
|      |                            | issuance date is same, then rate  |
|      |                            | as on approval date can be taken. |
+------+----------------------------+-----------------------------------+
| 7    | Backdating done & there is | Service Tax/GST Rate at the time  |
|      | no change in tax rate.     | of Proposal issuance will be      |
|      |                            | considered for tax calculation    |
|      |                            | (either service tax/GST) on the   |
|      |                            | changed '" Premium payable" (as   |
|      |                            | derived from modal premium) as    |
|      |                            | determined at the time of         |
|      |                            | approval.                         |
+------+----------------------------+-----------------------------------+
| 8    | Backdating done & there is | Service Tax/GST Rate at the time  |
|      | change in tax              | of Proposal issuance will be      |
|      | rate.(advance collection   | considered for tax calculation    |
|      | done for the same)         | (either service tax/GST) on the   |
|      |                            | changed '" Premium payable" (as   |
|      |                            | derived from modal premium) as    |
|      |                            | determined at the time of         |
|      |                            | approval.                         |
+------+----------------------------+-----------------------------------+
| 9    | Rebate applicable on the   | Rebate is not applicable at the   |
|      | premium                    | time of issuance and hence, GST   |
|      |                            | will be applied on modal premium  |
|      |                            | and rate will be taken up as of   |
|      |                            | Policy issuance date.             |
+------+----------------------------+-----------------------------------+
| 10   | GST Rate applicable for    | In case the indexing of the       |
|      | HO, SO and BO              | proposal or collection of the     |
|      |                            | renewal collection is done other  |
|      |                            | than HO, SO or BO, then error     |
|      |                            | message "GST not applicable for   |
|      |                            | offices other than HO, SO or BO"  |
|      |                            | will appear.                      |
+------+----------------------------+-----------------------------------+

Screens in NBF & changes if any applicable:

- Quote Screen:

Disclaimer to be shown Exclusive of Taxes

- New Business Indexing:

Premium inclusive of tax will include premium inclusive of GST.

- Processing screens (DE/QC/Approval screens):

Premium inclusive of tax will include premium inclusive of GST/Service
tax as applicable.

Accounting entry of suspense is passed on the day of collection of
initial premium and when proposal is accepted then, accounting entries
are passed for collection of premium and tax against earlier suspense
entry; and- in case of rejection/ withdrawal/ termination of proposal,
no tax is deducted out of initial premium.

### 3.2.2 Illustrations 

+--------------+-----------------------+----------------------------------------------------+----------------------------------------+
| Illustration | Scenarios             | Illustrations                                      | Expected System Response               |
| Number       |                       |                                                    |                                        |
+:=============+=======================+====================================================+========================================+
| IUD_1        | Indexing/collection   | Proposal number XXXX00011,                         | GST rate applicable as on 10-July-2017 |
|              | and approval is done  |                                                    | will be applicable.                    |
|              | in same month and     | Frequency: Monthly                                 |                                        |
|              | issuance              |                                                    | GST Rate applicable will be of Noida   |
|              | date=approval date    | Modal premium: 1200                                | HO.                                    |
|              | and this date is      |                                                    |                                        |
|              | after GST is made     | Premium collected as on indexing date.             |                                        |
|              | effective.            |                                                    |                                        |
|              |                       | Indexing date: 10-July-2017                        |                                        |
|              |                       |                                                    |                                        |
|              |                       | Issuance/RCD: 10-July-2017                         |                                        |
|              |                       |                                                    |                                        |
|              |                       | Approval date: 25-July-2017                        |                                        |
|              |                       |                                                    |                                        |
|              |                       | GST made effective on 01-July-2017                 |                                        |
|              |                       |                                                    |                                        |
|              |                       | Office where indexing is done = Noida HO (Office   |                                        |
|              |                       | Code = BR1000000000)                               |                                        |
|              |                       |                                                    |                                        |
|              |                       | User Office Code who has done the Indexing=        |                                        |
|              |                       | BR1000000000                                       |                                        |
+--------------+-----------------------+----------------------------------------------------+----------------------------------------+
| IUD_2        | Indexing /            | Proposal number XXXX00011,                         | Service tax rate applicable as on      |
|              | Collections/ Approval |                                                    | 10-Jun-2017 will be applicable.        |
|              | in the same month and | Frequency: Monthly                                 |                                        |
|              | issuance              |                                                    | Service tax rate applicable is that of |
|              | date/approval date is | Modal premium: 1200                                | ROI                                    |
|              | before GST is made    |                                                    |                                        |
|              | effective.            | Premium collected as on indexing date.             |                                        |
|              |                       |                                                    |                                        |
|              |                       | GST made effective 01-July-2017                    |                                        |
|              |                       |                                                    |                                        |
|              |                       | Indexing date: 10-Jun-2017                         |                                        |
|              |                       |                                                    |                                        |
|              |                       | Issuance/ RCD: 10-Jun-2017                         |                                        |
|              |                       |                                                    |                                        |
|              |                       | Approval date: 25-Jun-2017                         |                                        |
|              |                       |                                                    |                                        |
|              |                       | Issuance state: Rest ROI.                          |                                        |
+--------------+-----------------------+----------------------------------------------------+----------------------------------------+
| IUD_3        | Indexing /            | Proposal number XXXX00011,                         | Service tax rate applicable as on      |
|              | Collections/ Approval |                                                    | 10-Jun-2017 will be applicable.        |
|              | in different months   | Frequency: Monthly                                 |                                        |
|              | and issuance date is  |                                                    | On Rs.1200 service tax will be         |
|              | when Service tax was  | Modal premium: 1200                                | collected with tax rate 3.74%.         |
|              | effective and         |                                                    |                                        |
|              | approval date is      | Premium collected as on approval date.             | Rate applicable is that of ROI         |
|              | where GST was         |                                                    |                                        |
|              | effective.            | GST made effective 01-July-2017                    |                                        |
|              |                       |                                                    |                                        |
|              |                       | Service tax rate: 3.74%                            |                                        |
|              |                       |                                                    |                                        |
|              |                       | GST rate : 4.5%                                    |                                        |
|              |                       |                                                    |                                        |
|              |                       | Indexing date: 10-Jun-2017                         |                                        |
|              |                       |                                                    |                                        |
|              |                       | Issuance/ RCD: 10-Jun-2017                         |                                        |
|              |                       |                                                    |                                        |
|              |                       | Approval date: 25-Jul-2017                         |                                        |
|              |                       |                                                    |                                        |
|              |                       | Issuance state: Rest ROI.                          |                                        |
+--------------+-----------------------+----------------------------------------------------+----------------------------------------+
| IUD_4        | Indexing /            | Proposal number XXXX00011,                         | Service tax rate applicable as on      |
|              | Collections/ Approval |                                                    | 10-Jun-2017 will be applicable.        |
|              | in the same month     | Frequency: Monthly                                 | Rs.1200 will be collected with tax     |
|              | after GST is          |                                                    | rate 1.87%.                            |
|              | effective but policy  | Modal premium: 1200                                |                                        |
|              | is backdated to a     |                                                    | Rate applicable is that of ROI         |
|              | date where service    | Premium collected as on approval date.             |                                        |
|              | tax was effective.    |                                                    |                                        |
|              |                       | GST made effective 01-July-2017                    |                                        |
|              |                       |                                                    |                                        |
|              |                       | Service tax rate: 1.87%                            |                                        |
|              |                       |                                                    |                                        |
|              |                       | GST rate : 4.5%                                    |                                        |
|              |                       |                                                    |                                        |
|              |                       | Indexing date: 10-July-2017                        |                                        |
|              |                       |                                                    |                                        |
|              |                       | Approval date: 25-July-2017                        |                                        |
|              |                       |                                                    |                                        |
|              |                       | Issuance/RCD: 10-Jun-2017                          |                                        |
|              |                       |                                                    |                                        |
|              |                       | Issuance state: J&K.                               |                                        |
+--------------+-----------------------+----------------------------------------------------+----------------------------------------+
| IUD_5        | Rounding off Logic -  | Proposal number XXXX00011,                         | GST will be calculated as follows on   |
|              | Indexing/collection   |                                                    | modal premium Rs.1200                  |
|              | and approval is done  | Frequency: Monthly                                 |                                        |
|              | in same month after   |                                                    | First Year                             |
|              | implementation of     | Modal premium: 1200                                |                                        |
|              | GST.                  |                                                    | CGST @ 2.25% = 27                      |
|              |                       | Indexing date: 10-July-2017                        |                                        |
|              |                       |                                                    | UTGST/SGST @ 2.25%= 27                 |
|              |                       | Approval date: 25-July-2017                        |                                        |
|              |                       |                                                    | GST @ 4.50% = 54                       |
|              |                       | Issuance/ RCD: 10-July-2017                        |                                        |
|              |                       |                                                    | Rounded off value for GST = 54         |
|              |                       | Issuance date: 10-July-2017                        |                                        |
|              |                       |                                                    | Rate applicable will be that of Burari |
|              |                       | Office where indexing is done = BURARI SO          | SO.                                    |
|              |                       |                                                    |                                        |
|              |                       | Office where approval is done = NOIDA HO           |                                        |
|              |                       |                                                    |                                        |
|              |                       | Office Code of Noida HO = BR100200000              |                                        |
|              |                       |                                                    |                                        |
|              |                       | User Office Name/Code who has done the             |                                        |
|              |                       | Indexing=Burari SO (BR100200001)                   |                                        |
+--------------+-----------------------+----------------------------------------------------+----------------------------------------+
| IUD_6        | Indexing/collection   | Proposal number XXXX00011,                         | Service tax rate applicable as on      |
|              | is done in same month |                                                    | 10-Jun-2017 will be applicable will be |
|              | and approval is done  | Frequency: Monthly                                 | collected with tax rate 3.74%.         |
|              | in different month    |                                                    |                                        |
|              | after implementation  | Modal premium: 1200                                | Service tax Rate applicable is that of |
|              | of GST.               |                                                    | ROI                                    |
|              |                       | Indexing date: 10-Jun-2017                         |                                        |
|              |                       |                                                    |                                        |
|              |                       | Approval date: 05-July-2017                        |                                        |
|              |                       |                                                    |                                        |
|              |                       | Issuance/ RCD: 10-Jun-2017                         |                                        |
|              |                       |                                                    |                                        |
|              |                       | Service tax rate: 3.74%                            |                                        |
|              |                       |                                                    |                                        |
|              |                       | GST rate : 4.5%                                    |                                        |
|              |                       |                                                    |                                        |
|              |                       | Issuance state: Rest ROI.                          |                                        |
+--------------+-----------------------+----------------------------------------------------+----------------------------------------+
| IUD_7        | Indexing/collection   | Proposal number XXXX00011,                         | Service tax rate applicable as on      |
|              | and approval is done  |                                                    | 10-Jun-2017 will be applicable for     |
|              | in same month after   | Frequency: Monthly                                 | premium Rs. 2400 will be collected     |
|              | implementation of GST |                                                    | with tax rate 3.74%.                   |
|              | and RCD date is       | Modal premium: 1200                                |                                        |
|              | backdated.            |                                                    | Service tax Rate applicable is that of |
|              |                       | Indexing date: 10-July-2017                        | ROI                                    |
|              |                       |                                                    |                                        |
|              |                       | Approval date: 15-July-2017                        |                                        |
|              |                       |                                                    |                                        |
|              |                       | RCD: 10-Jun-2017                                   |                                        |
|              |                       |                                                    |                                        |
|              |                       | Issuance Date: 10-Jun-2017                         |                                        |
|              |                       |                                                    |                                        |
|              |                       | Due to difference in indexing/approval months the  |                                        |
|              |                       | modal premium to be collected is 1200\*2=2400.     |                                        |
|              |                       |                                                    |                                        |
|              |                       | Modal Premium to be collected Rs.2400.             |                                        |
|              |                       |                                                    |                                        |
|              |                       | Issuance state: Rest ROI.                          |                                        |
+--------------+-----------------------+----------------------------------------------------+----------------------------------------+
| IUD_8        | Indexing/collection   | Proposal number XXXX00011,                         | GST will be calculated on modal        |
|              | is done in same month |                                                    | premium Rs.1350 as follows:            |
|              | and approval is done  | Frequency: Monthly                                 |                                        |
|              | in different month    |                                                    | First Year                             |
|              | after implementation  | Modal premium: 1200                                |                                        |
|              | of GST.               |                                                    | CGST @ 2.25% = 30.375                  |
|              |                       | Indexing date: 10-July-2017                        |                                        |
|              | There is change in    |                                                    | UTGST/SGST @ 2.25%= 30.375             |
|              | modal premium due to  | Premium collected Rs.1200 as on indexing date.     |                                        |
|              | change in DOB.        |                                                    | GST @ 4.50% = 60.75                    |
|              |                       | Approval date: 25-Aug-2017                         |                                        |
|              |                       |                                                    | Rounded off value for GST = 61         |
|              |                       | Issuance/ RCD: 10-July-2017                        |                                        |
|              |                       |                                                    | GST Rate applicable will be that of    |
|              |                       | Due to change in date of Birth modal premium       | Mysore HO.                             |
|              |                       | changed to Rs.1350 which is collected at the time  |                                        |
|              |                       | of approval.                                       |                                        |
|              |                       |                                                    |                                        |
|              |                       | Office where indexing is done = BURARI SO          |                                        |
|              |                       |                                                    |                                        |
|              |                       | Office where approval is done = NOIDA HO           |                                        |
|              |                       |                                                    |                                        |
|              |                       | User Office Code who has done Indexing = Mysore HO |                                        |
|              |                       | = SK084900000                                      |                                        |
+--------------+-----------------------+----------------------------------------------------+----------------------------------------+
| IUD_9        | Indexing/collection   | Proposal number XXXX00011,                         | Service Tax applicable as on issuance  |
|              | is done in same month |                                                    | date i.e. 10-May-2017 will be applied  |
|              | and approval is done  | Frequency: Monthly                                 | on the modal premium Rs.2500.          |
|              | in different month    |                                                    |                                        |
|              | after implementation  | Modal premium: 1200                                | Service tax Rate applicable is that of |
|              | of GST.               |                                                    | ROI                                    |
|              |                       | Indexing date: 10-Jun-2017                         |                                        |
|              | There is change in    |                                                    |                                        |
|              | modal premium due to  | Premium collected Rs.1200 as on indexing date.     |                                        |
|              | change in DOB.        |                                                    |                                        |
|              |                       | Approval date: 15-July-2017                        |                                        |
|              | RCD is backdated.     |                                                    |                                        |
|              |                       | Issuance/RCD: 10-May-2017                          |                                        |
|              |                       |                                                    |                                        |
|              |                       | Due to change in date of Birth and change in RCD   |                                        |
|              |                       | modal premium changed to Rs.2500                   |                                        |
|              |                       |                                                    |                                        |
|              |                       | Issuance state: Rest ROI.                          |                                        |
+--------------+-----------------------+----------------------------------------------------+----------------------------------------+
| IUD_10       | Rejection or          | Proposal number XXXX00011,                         | GST Rs.54 thus collected will ~~not~~  |
|              | termination of new    |                                                    | be refunded along with the premium.    |
|              | business proposal     | Frequency: Monthly                                 |                                        |
|              | after implementation  |                                                    | Premiums will be refunded after        |
|              | of GST                | Modal premium: 1200                                | deducting medical fees.                |
|              |                       |                                                    |                                        |
|              |                       | Premium collected Rs.1200 as on indexing date.     | Same will happen in case of            |
|              |                       |                                                    | termination.                           |
|              |                       | Premium includes GST Rs.54.                        |                                        |
|              |                       |                                                    |                                        |
|              |                       | Medical fees Rs.150 collected                      |                                        |
|              |                       |                                                    |                                        |
|              |                       | Proposal is rejected by the approver.              |                                        |
|              |                       |                                                    |                                        |
|              |                       | Indexing date: 10-July-2017                        |                                        |
|              |                       |                                                    |                                        |
|              |                       | Rejection date: 15-July-2017                       |                                        |
+--------------+-----------------------+----------------------------------------------------+----------------------------------------+
| IUD_11       | Rejection or          | Proposal number XXXX00011,                         | Service tax Rs.20 thus collected will  |
|              | termination of new    |                                                    | ~~not~~ be refunded along with the     |
|              | business proposal     | Frequency: Monthly                                 | premiums.                              |
|              | after implementation  |                                                    |                                        |
|              | of GST                | Modal premium: 1200                                | Same will happen in case of rejection  |
|              |                       |                                                    | or termination.                        |
|              |                       | Premium collected Rs.1220 as on indexing date.     |                                        |
|              |                       |                                                    |                                        |
|              |                       | Premium includes service tax Rs.20.                |                                        |
|              |                       |                                                    |                                        |
|              |                       | Proposal is rejected by the approver.              |                                        |
|              |                       |                                                    |                                        |
|              |                       | Indexing date: 10-Jun-2017                         |                                        |
|              |                       |                                                    |                                        |
|              |                       | Rejection date: 15-July-2017                       |                                        |
+--------------+-----------------------+----------------------------------------------------+----------------------------------------+
| IUD_12       | Indexing/collection   | Proposal number XXXX00011,                         | GST applicable as on issuance date     |
|              | and approval is done  |                                                    | i.e. 10-July-2017 will be applied on   |
|              | in same month.        | Frequency: Monthly                                 | the total premium Rs.1300.             |
|              |                       |                                                    |                                        |
|              | Advance premium is    | Modal premium: 1200                                | GST Rate applicable will be that of    |
|              | collected as on       |                                                    | Burari SO.                             |
|              | approval date.        | Indexing date: 10-July-2017                        |                                        |
|              |                       |                                                    |                                        |
|              |                       | Premium collected Rs.1200 as on indexing date.     |                                        |
|              |                       |                                                    |                                        |
|              |                       | Approval date: 15-June-2017                        |                                        |
|              |                       |                                                    |                                        |
|              |                       | Advance premium Rs.100 is collected on approval    |                                        |
|              |                       | date.                                              |                                        |
|              |                       |                                                    |                                        |
|              |                       | Issuance/ RCD: 10-July-2017                        |                                        |
|              |                       |                                                    |                                        |
|              |                       | GST made effective 01-July-2017                    |                                        |
|              |                       |                                                    |                                        |
|              |                       | Service tax rate: 3.74%                            |                                        |
|              |                       |                                                    |                                        |
|              |                       | GST rate : 4.5%                                    |                                        |
|              |                       |                                                    |                                        |
|              |                       | Office where indexing is done = BURARI SO          |                                        |
|              |                       |                                                    |                                        |
|              |                       | Office where approval is done = NOIDA HO           |                                        |
|              |                       |                                                    |                                        |
|              |                       | User Office Code who has done Indexing = BURARI SO |                                        |
|              |                       | = SK084900000                                      |                                        |
+--------------+-----------------------+----------------------------------------------------+----------------------------------------+
| IUD_13       | Indexing/collection   | Proposal number XXXX00011,                         | Service Tax will be calculated on      |
|              | is done in same month |                                                    | Rs.2400 total premium.                 |
|              | and approval is done  | Frequency: Monthly                                 |                                        |
|              | in different month.   |                                                    | Service tax Rate applicable is that of |
|              |                       | Modal premium: 1200                                | ROI                                    |
|              | Advance premium is    |                                                    |                                        |
|              | collected as on       | Indexing date: 10-Jun-2017                         |                                        |
|              | approval date.        |                                                    |                                        |
|              |                       | Premium collected Rs.1200 as on indexing date.     |                                        |
|              |                       |                                                    |                                        |
|              |                       | Approval date: 15-July-2017                        |                                        |
|              |                       |                                                    |                                        |
|              |                       | Premium Rs.2400 is collected on approval date.     |                                        |
|              |                       |                                                    |                                        |
|              |                       | Issunace/RCD: 10-Jun-2017                          |                                        |
|              |                       |                                                    |                                        |
|              |                       | GST made effective 01-July-2017                    |                                        |
|              |                       |                                                    |                                        |
|              |                       | Service tax rate: 3.74%                            |                                        |
|              |                       |                                                    |                                        |
|              |                       | GST rate : 4.5%                                    |                                        |
|              |                       |                                                    |                                        |
|              |                       | Issuance state: Rest ROI.                          |                                        |
+--------------+-----------------------+----------------------------------------------------+----------------------------------------+
| IUD_14       | Indexing/collection   | Proposal number XXXX00011,                         | Service Tax will be calculated on      |
|              | is done in same month |                                                    | Rs.3000 total premium collected.       |
|              | and approval is done  | Frequency: Monthly                                 |                                        |
|              | in different month    |                                                    | Service tax Rate applicable is that of |
|              | and issuance date is  | Modal premium: 1200                                | ROI                                    |
|              | backdated.            |                                                    |                                        |
|              |                       | Indexing date: 10-Jun-2017                         |                                        |
|              | Advance premium is    |                                                    |                                        |
|              | collected as on       | Premium collected Rs.1200 as on indexing date.     |                                        |
|              | approval date.        |                                                    |                                        |
|              |                       | Approval date: 15-July-2017                        |                                        |
|              |                       |                                                    |                                        |
|              |                       | Premium Rs.3000 is collected on approval date.     |                                        |
|              |                       |                                                    |                                        |
|              |                       | Issuance/ RCD: 10-May-2017                         |                                        |
|              |                       |                                                    |                                        |
|              |                       | Indexing date: 10-Jun-2017                         |                                        |
+--------------+-----------------------+----------------------------------------------------+----------------------------------------+
| IUD_15       | Advance premium is    | Proposal number XXXX00011,                         | GST will be calculated as follows on   |
|              | collected for first   |                                                    | premiums collected.                    |
|              | and second policy     | Frequency: Monthly                                 |                                        |
|              | year                  |                                                    | **First Year**                         |
|              |                       | Modal premium: 525                                 |                                        |
|              |                       |                                                    | From July-2017 to Jun 2018 on Rs.6300  |
|              |                       | Indexing date: 10-July-2017                        |                                        |
|              |                       |                                                    | CGST @ 2.25% = 141.75                  |
|              |                       | Premium collected for first and second policy year |                                        |
|              |                       | Rs.12600 as on indexing date.                      | UTGST/SGST @ 2.25%= 141.75             |
|              |                       |                                                    |                                        |
|              |                       | Approval date : 15-July-2017                       | GST @ 4.50% = 283.50 (A)               |
|              |                       |                                                    |                                        |
|              |                       | Issuance/ RCD: 10-July-2017                        | **Renewal Year**                       |
|              |                       |                                                    |                                        |
|              |                       | GST made effective 01-July-2017                    | From July 2018 to June 2019 on Rs.6300 |
|              |                       |                                                    |                                        |
|              |                       | Service tax rate: 3.74%                            | CGST @ 2.25% = 141.75                  |
|              |                       |                                                    |                                        |
|              |                       | GST rate : 4.5%                                    | UTGST/SGST @ 2.25%= 141.75             |
|              |                       |                                                    |                                        |
|              |                       |                                                    | GST @ 4.50% = 283.50 (A)               |
|              |                       |                                                    |                                        |
|              |                       |                                                    | Total GST (A + B) =567                 |
+--------------+-----------------------+----------------------------------------------------+----------------------------------------+
| IUD_16       | Initial premium       | Proposal number XXXX00011,                         | GST will be calculated as follows on   |
|              | collected for the     |                                                    | Premium Rs.450                         |
|              | policy wherein sum    | Frequency: Monthly                                 |                                        |
|              | assured rebate was    |                                                    | First Year                             |
|              | given                 | Modal premium (adjusted with SA Rebate): 450       |                                        |
|              |                       |                                                    | CGST @ 2.25% = 10.125                  |
|              |                       | Indexing date: 10-July-2017                        |                                        |
|              |                       |                                                    | UTGST/SGST @ 2.25%= 10.125             |
|              |                       | Approval date: 15-July-2017                        |                                        |
|              |                       |                                                    | GST @ 4.50% = 20.25                    |
|              |                       | RCD: 10-July-2017                                  |                                        |
|              |                       |                                                    | Rounded off value for GST = 20         |
|              |                       | Issuance date : 10-July-2017                       |                                        |
+--------------+-----------------------+----------------------------------------------------+----------------------------------------+
| IUD_17       | Change in Issuance    | Policy number XXXX00011,                           | GST Rate applicable will be that of    |
|              | office                |                                                    | Dhule SO.                              |
|              |                       | Policy issuance date is 15-Jul-2017                |                                        |
|              |                       |                                                    |                                        |
|              |                       | Before approval the circle has been change to      |                                        |
|              |                       | Tamil Nadu with issuance HO as Pondicherry.        |                                        |
|              |                       |                                                    |                                        |
|              |                       | User Office Code who has done Indexing = Dhule SO  |                                        |
|              |                       | = SK084900100                                      |                                        |
+--------------+-----------------------+----------------------------------------------------+----------------------------------------+
| IUD_18       | Renewal premium       | Policy number XXXX00011,                           | GST Rate will be calculated as follows |
|              | collected in first    |                                                    | on modal premium Rs.500                |
|              | policy year           | Modal Premium Rs.500                               |                                        |
|              |                       |                                                    | First Year                             |
|              |                       | Frequency: Monthly                                 |                                        |
|              |                       |                                                    | CGST @ 2.25% = 11.25                   |
|              |                       | Policy issuance date is 15-Jan-2017                |                                        |
|              |                       |                                                    | UTGST/SGST @ 2.25%= 11.25              |
|              |                       | Paid to date is 30-Jun-2017                        |                                        |
|              |                       |                                                    | GST @ 4.50% = 22.50                    |
|              |                       | System date is 12-July-2017                        |                                        |
|              |                       |                                                    | Rounded off value for GST = 23         |
|              |                       | Premium collected Rs.500 as on 12 July 2017.       |                                        |
|              |                       |                                                    | GST Rate applicable will be that of    |
|              |                       | Premium collection Office = Noida HO (Office Code  | Noida HO                               |
|              |                       | = BR100200000)                                     |                                        |
|              |                       |                                                    |                                        |
|              |                       | User Office Code who is doing the collection =     |                                        |
|              |                       | Noida HO (Office Code = BR100200000)               |                                        |
+--------------+-----------------------+----------------------------------------------------+----------------------------------------+
| IUD_19       | Reinstatement premium | Policy number XXXX00011,                           | GST will be calculated on Rs.3000 as   |
|              | collected in case of  |                                                    | follows on premiums due exclusive of   |
|              | revival wherein       | Policy issuance date is 01-May-2017                | interest.                              |
|              | premium is to be      |                                                    |                                        |
|              | collected for first   | Paid to date is 31-Mar-2018                        | **First Year**                         |
|              | and second policy     |                                                    |                                        |
|              | years. Interest       | Frequency is monthly                               | For April 2018 at Rs.500               |
|              | charged on the        |                                                    |                                        |
|              | premium due.          | Premium 500                                        | CGST @ 2.25% = 11.25                   |
|              |                       |                                                    |                                        |
|              |                       | System date is 01-July-2018                        | UTGST/SGST @ 2.25%= 11.25              |
|              |                       |                                                    |                                        |
|              |                       | Premium collected Rs.2000                          | GST @ 4.50% = 22.50 (A)                |
|              |                       |                                                    |                                        |
|              |                       | Interest Rs.200                                    | **Renewal Year**                       |
|              |                       |                                                    |                                        |
|              |                       | Premium collection Office = Noida HO               | From May 2018 to July 2018 at Rs.1500  |
|              |                       |                                                    |                                        |
|              |                       | Office Code = BR100200000                          | CGST @ 1.125% = 16.875                 |
|              |                       |                                                    |                                        |
|              |                       | User Office Code =BR100200000                      | UTGST/SGST @ 1.125%= 16.875            |
|              |                       |                                                    |                                        |
|              |                       |                                                    | GST @ 2.25% = 33.75 (B)                |
|              |                       |                                                    |                                        |
|              |                       |                                                    | Total GST (A + B) = 56.25              |
|              |                       |                                                    |                                        |
|              |                       |                                                    | Rounded off GST value = Rs.56          |
|              |                       |                                                    |                                        |
|              |                       |                                                    | GST Rate applicable will be that of    |
|              |                       |                                                    | Noida HO                               |
+--------------+-----------------------+----------------------------------------------------+----------------------------------------+
| IUD_20       | Revival premium       | Policy number XXXX00011,                           | GST will be calculated on Rs.3500 as   |
|              | collected in case of  |                                                    | follows on premiums due exclusive of   |
|              | revival wherein       | Policy issuance date is 01-May-2017                | interest.                              |
|              | premium is to be      |                                                    |                                        |
|              | collected for first   | Paid to date is 31-Jan-2018                        | **First Year**                         |
|              | and second policy     |                                                    |                                        |
|              | years. Interest       | Frequency is monthly                               | From Feb 2018 to April 2018 at Rs.1500 |
|              | charged on the        |                                                    |                                        |
|              | premium due.          | Premium 500                                        | CGST @ 2.25% = 33.75                   |
|              |                       |                                                    |                                        |
|              |                       | System date is 01-August-2018                      | UTGST/SGST @ 2.25%= 33.75              |
|              |                       |                                                    |                                        |
|              |                       | Premium collected Rs.3500                          | GST @ 4.50% = 67.50 (A)                |
|              |                       |                                                    |                                        |
|              |                       | Policy status is AL                                | **Renewal Year**                       |
|              |                       |                                                    |                                        |
|              |                       | Interest Rs.200                                    | From May 2018 to August 2018 at        |
|              |                       |                                                    | Rs.2000                                |
|              |                       | Premium collection Office = Noida HO               |                                        |
|              |                       |                                                    | CGST @ 1.125% = 22.5                   |
|              |                       | Office Code = BR100200000                          |                                        |
|              |                       |                                                    | UTGST/SGST @ 1.125%= 22.5              |
|              |                       | User Office Code =BR100200000                      |                                        |
|              |                       |                                                    | GST @ 2.25% = 45                       |
|              |                       |                                                    |                                        |
|              |                       |                                                    | Total GST (A + B) =112.50              |
|              |                       |                                                    |                                        |
|              |                       |                                                    | Rounded off GST value = 113            |
|              |                       |                                                    |                                        |
|              |                       |                                                    | GST Rate applicable will be that of    |
|              |                       |                                                    | Noida HO                               |
+--------------+-----------------------+----------------------------------------------------+----------------------------------------+
| IUD_21       | Advance premium is    | Policy number XXXX00011,                           | GST will be calculated after deducting |
|              | collected on the      |                                                    | rebate from premium on Rs.3450 as      |
|              | policy                | Policy issuance date is 01-Jan-2016                | follows                                |
|              |                       |                                                    |                                        |
|              |                       | Paid to date is 30-June-2017                       | **Renewal Year**                       |
|              |                       |                                                    |                                        |
|              |                       | System date is 12-July-2017                        | CGST @ 1.125% = 38.81.375              |
|              |                       |                                                    |                                        |
|              |                       | Frequency is monthly                               | UTGST/SGST @ 1.125%=38.81              |
|              |                       |                                                    |                                        |
|              |                       | Modal Premium is Rs.500                            | GST @ 2.25% = 77.62                    |
|              |                       |                                                    |                                        |
|              |                       | Premium is paid from 01-July-2017 to               | Rounded off value for GST = 78         |
|              |                       | 31-Jan-2018=3500                                   |                                        |
|              |                       |                                                    | GST Rate applicable will be that of    |
|              |                       | Rebate Rs.50                                       | Noida HO                               |
|              |                       |                                                    |                                        |
|              |                       | Rebate is calculated for advance premium paid for  |                                        |
|              |                       | 7 months                                           |                                        |
|              |                       |                                                    |                                        |
|              |                       | Premium collection Office = Noida HO               |                                        |
|              |                       |                                                    |                                        |
|              |                       | Office Code = BR100200000                          |                                        |
|              |                       |                                                    |                                        |
|              |                       | User Office Code =BR100200000                      |                                        |
+--------------+-----------------------+----------------------------------------------------+----------------------------------------+
| IUD_22       | ~~Renewal premium is  | ~~Policy is issued in J&K and renewal premium is   | ~~Service tax at the rate applicable   |
|              | collected in the      | collected in Maharashtra~~                         | for J & K will be applied on the       |
|              | state other than      |                                                    | renewal premiums.~~                    |
|              | issuance state~~      |                                                    |                                        |
+--------------+-----------------------+----------------------------------------------------+----------------------------------------+
| IUD_23       | Cheque dishonour      | Policy number XXXX00011,                           | GST thus collected on premiums will    |
|              | after renewal premium |                                                    | not be reversed in case of cheque      |
|              | collection            | Policy issuance date is 15-Jan-2017                | dishonour.                             |
|              |                       |                                                    |                                        |
|              |                       | Paid to date is 31-Mar-2017                        | However, the amount of GST so          |
|              |                       |                                                    | collected through dishonoured cheque   |
|              |                       | System date is 12-July-2017                        | will be reflected as minus balance in  |
|              |                       |                                                    | suspense and adjusted against          |
|              |                       | Frequency is monthly                               | subsequent collection from the         |
|              |                       |                                                    | insurant or Claim settlement, as the   |
|              |                       | Premium Rs.500 (including GST) is collected        | case may be, along with cheque         |
|              |                       | through cheque. Cheque dishonour is intimated as   | dishonour charge.                      |
|              |                       | on 20.07.2017.                                     |                                        |
+--------------+-----------------------+----------------------------------------------------+----------------------------------------+
| IUD_24       | Receipt Cancellation  | Policy number XXXX00011,                           | GST thus collected on premiums will be |
|              | after renewal premium |                                                    | reversed in case of receipt            |
|              | collection            | Policy issuance date is 15-Jan-2017                | cancellation.                          |
|              |                       |                                                    |                                        |
|              |                       | Paid to date is 31-Mar-2017                        |                                        |
|              |                       |                                                    |                                        |
|              |                       | System date is 12-July-2017                        |                                        |
|              |                       |                                                    |                                        |
|              |                       | Frequency is monthly                               |                                        |
|              |                       |                                                    |                                        |
|              |                       | Premium Rs.500 (including GST) is collected as on  |                                        |
|              |                       | 12-July-2017 in policy number XXXX00012.           |                                        |
|              |                       |                                                    |                                        |
|              |                       | Receipt is cancelled on the same day as premium is |                                        |
|              |                       | collected in the wrong policy.                     |                                        |
+--------------+-----------------------+----------------------------------------------------+----------------------------------------+
| IUD_25       | Meghdoot or Bulk      | Policy number XXXX00011,                           | GST will not be applicable on the      |
|              | upload                |                                                    | premium collection as transaction date |
|              |                       | Policy issuance date is 15-Jan-2017                | is prior to that of 01.07.2017.        |
|              |                       |                                                    |                                        |
|              |                       | Paid to date is 31-May-2017                        | In this case, service tax will be      |
|              |                       |                                                    | applicable. System will validate this  |
|              |                       | System date is 12-July-2017                        | based on transaction date.             |
|              |                       |                                                    |                                        |
|              |                       | Frequency is monthly                               |                                        |
|              |                       |                                                    |                                        |
|              |                       | Premium Rs.500                                     |                                        |
|              |                       |                                                    |                                        |
|              |                       | Meghdoot upload is done as on 12-July-2017 and     |                                        |
|              |                       | transaction date in the file is 01-June-2017.      |                                        |
+--------------+-----------------------+----------------------------------------------------+----------------------------------------+
| IUD_26       | Meghdoot or Bulk      | Policy number XXXX00011,                           | GST will be applicable on the premium  |
|              | upload                |                                                    | collection as transaction date is that |
|              |                       | Policy issuance date is 15-Jan-2017                | of 01.07.2017. System will validate    |
|              |                       |                                                    | this based on transaction date.        |
|              |                       | Paid to date is 30-June-2017                       |                                        |
|              |                       |                                                    | GST Rate applicable will be that of    |
|              |                       | System date is 12-July-2017                        | Noida HO.                              |
|              |                       |                                                    |                                        |
|              |                       | Frequency is monthly                               |                                        |
|              |                       |                                                    |                                        |
|              |                       | Premium Rs.500                                     |                                        |
|              |                       |                                                    |                                        |
|              |                       | Meghdoot upload is done as on 12-July-2017 and     |                                        |
|              |                       | transaction date in the file is 01-July-2017.      |                                        |
|              |                       |                                                    |                                        |
|              |                       | Office = NOIDA HO                                  |                                        |
|              |                       |                                                    |                                        |
|              |                       | Office Code of the user uploading the file = Noida |                                        |
|              |                       | HO=BR100200000                                     |                                        |
+--------------+-----------------------+----------------------------------------------------+----------------------------------------+
| IUD_27       | Loan repayment        | Policy number XXXX00011,                           | GST will not be charged for loan       |
|              |                       |                                                    | repayment.                             |
|              |                       | Policy issuance date is 15-Jan-2015                |                                        |
|              |                       |                                                    |                                        |
|              |                       | Loan outstanding 5000                              |                                        |
|              |                       |                                                    |                                        |
|              |                       | Loan repayment done as on 12-July-2017.            |                                        |
+--------------+-----------------------+----------------------------------------------------+----------------------------------------+
| IUD_28       | Miscellaneous         | Policy number XXXX00011,                           | GST will not be charged for            |
|              | collection-Duplicate  |                                                    | miscellaneous collection.              |
|              | Policy Bond           | Policy issuance date is 15-Jan-2015                |                                        |
|              |                       |                                                    |                                        |
|              |                       | Miscellaneous collection is done for duplicate     |                                        |
|              |                       | policy bond                                        |                                        |
+--------------+-----------------------+----------------------------------------------------+----------------------------------------+
| IUD_29       | Miscellaneous         | Policy number XXXX00011,                           | GST will not be charged for            |
|              | collection-Duplicate  |                                                    | miscellaneous collection.              |
|              | receipt book          | Policy issuance date is 15-Jan-2015                |                                        |
|              |                       |                                                    |                                        |
|              |                       | Miscellaneous collection is done for duplicate     |                                        |
|              |                       | receipt book.                                      |                                        |
+--------------+-----------------------+----------------------------------------------------+----------------------------------------+
| IUD_30       | Miscellaneous         | Policy number XXXX00011,                           | GST will not be charged for            |
|              | collection-cheque     |                                                    | miscellaneous collection.              |
|              | dishonour charges     | Policy issuance date is 15-Jan-2015                |                                        |
|              |                       |                                                    |                                        |
|              |                       | Cheque dishonour charges are collected.            |                                        |
+--------------+-----------------------+----------------------------------------------------+----------------------------------------+
| itIUD_31     | Miscellaneous         | Policy number XXXX00011,                           | GST will not be charged for            |
|              | collection-Conversion |                                                    | miscellaneous collection.              |
|              | charges               | Policy issuance date is 15-Jan-2015                |                                        |
|              |                       |                                                    |                                        |
|              |                       | Policy Converted from CWLA to EA. Conversion       |                                        |
|              |                       | charges collected.                                 |                                        |
+--------------+-----------------------+----------------------------------------------------+----------------------------------------+
| IUD_32       | GST in case of        | Policy number XXXX00011,                           | GST rate at the time of indexing of    |
|              | recovery of unpaid    |                                                    | claim will be applied on unpaid        |
|              | premiums in case of   | Policy issuance date is 01-May-2010                | premiums.                              |
|              | maturity or death     |                                                    |                                        |
|              | claim                 | Paid to date is 31-Jan-2018                        | GST Rate applicable is that of Noida   |
|              |                       |                                                    | HO.                                    |
|              |                       | Premium 500                                        |                                        |
|              |                       |                                                    |                                        |
|              |                       | Maturity date is 01-May-2018                       |                                        |
|              |                       |                                                    |                                        |
|              |                       | Maturity claim is indexed on 05-April -2018        |                                        |
|              |                       |                                                    |                                        |
|              |                       | Policy status is IL                                |                                        |
|              |                       |                                                    |                                        |
|              |                       | Claim indexing office code = Noida HO =            |                                        |
|              |                       | BR100200000                                        |                                        |
|              |                       |                                                    |                                        |
|              |                       | User office code indexing the claim = Noida HO =   |                                        |
|              |                       | BR100200000                                        |                                        |
+--------------+-----------------------+----------------------------------------------------+----------------------------------------+
| IUD_33       | Renewal premium       | Policy number XXXX00011,                           | GST will be calculated as follows on   |
|              | collected partly in   |                                                    | modal premium Rs.1000                  |
|              | cash and partly in    | Policy issuance date is 15-Jan-2017                |                                        |
|              | cheque.               |                                                    | First Year                             |
|              |                       | Modal Premium Rs.500                               |                                        |
|              |                       |                                                    | CGST @ 2.25% = 22.5                    |
|              |                       | Frequency: Monthly                                 |                                        |
|              |                       |                                                    | UTGST/SGST @ 2.25%= 22.50              |
|              |                       | Paid to date is 31-May-2017                        |                                        |
|              |                       |                                                    | GST @ 4.50% = 45                       |
|              |                       | System date is 12-July-2017                        |                                        |
|              |                       |                                                    | Rounded off value for GST = 45         |
|              |                       | Total premium to be collected is Rs.1000.          |                                        |
|              |                       |                                                    | Premium to be collected with tax is    |
|              |                       |                                                    | Rs.1045.                               |
|              |                       |                                                    |                                        |
|              |                       |                                                    | GST Rs.45 will be collected in cash.   |
|              |                       |                                                    |                                        |
|              |                       |                                                    | System will take care of bifurcation   |
|              |                       |                                                    | of GST from Cash amount collected.     |
+--------------+-----------------------+----------------------------------------------------+----------------------------------------+
| IUD_34       | If Renewal premium is | Policy number XXXX00011,                           | GST will be calculated as follows on   |
|              | paid through customer |                                                    | modal premium Rs.500                   |
|              | portal                | Policy issuance date is 15-Jan-2017                |                                        |
|              |                       |                                                    | First Year                             |
|              |                       | Modal Premium Rs.500                               |                                        |
|              |                       |                                                    | CGST @ 2.25% = 11.25                   |
|              |                       | Frequency: Monthly                                 |                                        |
|              |                       |                                                    | UTGST/SGST @ 2.25%= 11.25              |
|              |                       | Paid to date is 30-Jun-2017                        |                                        |
|              |                       |                                                    | GST @ 4.50% = 22.50                    |
|              |                       | System date is 12-July-2017                        |                                        |
|              |                       |                                                    | Rounded off value for GST = 23         |
|              |                       | Premium collected Rs.500                           |                                        |
|              |                       |                                                    | Rate will be that of the Sansad Marg   |
|              |                       |                                                    | HO.                                    |
+--------------+-----------------------+----------------------------------------------------+----------------------------------------+
| IUD_34       | If indexing is done   | Proposal number XXXX00011,                         | GST rate as on 10-July-2017 will be    |
|              | through the customer  |                                                    | applicable.                            |
|              | portal                | Indexing date: 10-July-2017                        |                                        |
|              |                       |                                                    | GST Rate applicable will be of Noida   |
|              |                       | Frequency: Monthly                                 | HO.                                    |
|              |                       |                                                    |                                        |
|              |                       | Modal premium: 1200                                |                                        |
|              |                       |                                                    |                                        |
|              |                       | Premium collected as on indexing date.             |                                        |
|              |                       |                                                    |                                        |
|              |                       | Approval date: 25-July-2017                        |                                        |
|              |                       |                                                    |                                        |
|              |                       | RCD: 10-July-2017                                  |                                        |
|              |                       |                                                    |                                        |
|              |                       | Issuance date: 10-July-2017                        |                                        |
|              |                       |                                                    |                                        |
|              |                       | Office where proposal is indexed on counter (In    |                                        |
|              |                       | McCamish) = Noida HO                               |                                        |
|              |                       |                                                    |                                        |
|              |                       | Office Code = BR100200000                          |                                        |
|              |                       |                                                    |                                        |
|              |                       | User Office Code =BR100200000                      |                                        |
+--------------+-----------------------+----------------------------------------------------+----------------------------------------+
| IUD_35       | GST Rate applicable   | Proposal number XXXX00011,                         | Error message should be displayed as   |
|              | for HO, SO and BO     |                                                    | "GST not applicable for offices other  |
|              |                       | Indexing date: 10-July-2017                        | than HO, SO or BO".                    |
|              |                       |                                                    |                                        |
|              |                       | Frequency: Monthly                                 |                                        |
|              |                       |                                                    |                                        |
|              |                       | Modal premium: 1200                                |                                        |
|              |                       |                                                    |                                        |
|              |                       | Premium collected as on indexing date.             |                                        |
|              |                       |                                                    |                                        |
|              |                       | Approval date: 25-July-2017                        |                                        |
|              |                       |                                                    |                                        |
|              |                       | RCD: 10-July-2017                                  |                                        |
|              |                       |                                                    |                                        |
|              |                       | Issuance date: 10-July-2017                        |                                        |
|              |                       |                                                    |                                        |
|              |                       | User Office Code who is doing the indexing belongs |                                        |
|              |                       | to UTTAR PRADESH CO                                |                                        |
+--------------+-----------------------+----------------------------------------------------+----------------------------------------+
| IUD_36       | Impact of GST on      | Report generated for period : 01 July to 02 July   | In the report generated the GST        |
|              | reports.              |                                                    | bifurcation between cash & cheque as   |
|              |                       | Details of Premium collection done as below:       | per the premium method is as follows   |
|              |                       |                                                    | where first GST will be recovered from |
|              |                       |   ------------------------------------------------ | cash and then from cheque.             |
|              |                       |    Details    Initial      Renewal      Revival    |                                        |
|              |                       |              Collection   Collection   Collection  |   ------------------------------------ |
|              |                       |              Proposal 1   Receipt 2    Receipt 2   |   GST Recovery  TOTAL   Cash   Cheque  |
|              |                       |   --------- ------------ ------------ ------------ |                  GST                   |
|              |                       |    Premium      1500         2000         500      |   ------------ ------- ------ -------- |
|              |                       |     Paid                                           |     Initial      68      0       68    |
|              |                       |                                                    |    Collection                          |
|              |                       |     CGST       33.75         22.5        11.25     |    Proposal 1                          |
|              |                       |                                                    |                                        |
|              |                       |     SGST       33.75         22.5        11.25     |     Renewal      45      45      0     |
|              |                       |    /UTGST                                          |    Collection                          |
|              |                       |                                                    |    Receipt 2                           |
|              |                       |   TOTAL GST      68           45           23      |                                        |
|              |                       |                                                    |     Revival      23      20      3     |
|              |                       |     Cash         0           2000          20      |    Collection                          |
|              |                       |                                                    |    Receipt 2                           |
|              |                       |    Cheque       1500          0           480      |   ------------------------------------ |
|              |                       |   ------------------------------------------------ |                                        |
|              |                       |                                                    | The summary will accordingly show the  |
|              |                       |                                                    | bifurcation as below:                  |
|              |                       |                                                    |                                        |
|              |                       |                                                    |   --------------------                 |
|              |                       |                                                    |   First Year    1500                   |
|              |                       |                                                    |     Premium                            |
|              |                       |                                                    |   ----------- --------                 |
|              |                       |                                                    |     Renewal     2000                   |
|              |                       |                                                    |     Premium                            |
|              |                       |                                                    |                                        |
|              |                       |                                                    |     Revival     500                    |
|              |                       |                                                    |     Premium                            |
|              |                       |                                                    |                                        |
|              |                       |                                                    |   First Year   33.75                   |
|              |                       |                                                    |      CGST                              |
|              |                       |                                                    |    collected                           |
|              |                       |                                                    |                                        |
|              |                       |                                                    |   First Year   33.75                   |
|              |                       |                                                    |      SGST                              |
|              |                       |                                                    |    Collected                           |
|              |                       |                                                    |                                        |
|              |                       |                                                    |   First Year     0                     |
|              |                       |                                                    |      UTGST                             |
|              |                       |                                                    |    Collected                           |
|              |                       |                                                    |                                        |
|              |                       |                                                    |     Renewal     22.5                   |
|              |                       |                                                    |    Year CGST                           |
|              |                       |                                                    |    Collected                           |
|              |                       |                                                    |                                        |
|              |                       |                                                    |     Renewal     22.5                   |
|              |                       |                                                    |    Year SGST                           |
|              |                       |                                                    |    Collected                           |
|              |                       |                                                    |                                        |
|              |                       |                                                    |     Renewal      0                     |
|              |                       |                                                    |   Year UTGST                           |
|              |                       |                                                    |    Collected                           |
|              |                       |                                                    |                                        |
|              |                       |                                                    |      Total      4000                   |
|              |                       |                                                    |     Amount                             |
|              |                       |                                                    |    Collected                           |
|              |                       |                                                    |                                        |
|              |                       |                                                    |      Total      4000                   |
|              |                       |                                                    |     Premium                            |
|              |                       |                                                    |     paid by                            |
|              |                       |                                                    |    Cash and                            |
|              |                       |                                                    |     Cheque                             |
|              |                       |                                                    |                                        |
|              |                       |                                                    |      Total       0                     |
|              |                       |                                                    |    Cancelled                           |
|              |                       |                                                    |     Amount                             |
|              |                       |                                                    |   --------------------                 |
+--------------+-----------------------+----------------------------------------------------+----------------------------------------+
| IUD_36       | GST Rate applicable   | Policy number XXXX00011,                           | GST rate applicable will be of         |
|              | for Revival Quote     |                                                    | issuance office code which is Noida HO |
|              |                       | Policy issuance date is 01-May-2017                |                                        |
|              |                       |                                                    |                                        |
|              |                       | Paid to date is 31-Jan-2018                        |                                        |
|              |                       |                                                    |                                        |
|              |                       | Frequency is monthly                               |                                        |
|              |                       |                                                    |                                        |
|              |                       | Premium 500                                        |                                        |
|              |                       |                                                    |                                        |
|              |                       | System date is 01-August-2018                      |                                        |
|              |                       |                                                    |                                        |
|              |                       | Policy status is AL                                |                                        |
|              |                       |                                                    |                                        |
|              |                       | Revival Quote 01 Aug 2018                          |                                        |
|              |                       |                                                    |                                        |
|              |                       | Issuance office: Noida HO                          |                                        |
|              |                       |                                                    |                                        |
|              |                       | User ID who is doing Revival Quote generation:     |                                        |
|              |                       | Mysore HO                                          |                                        |
+--------------+-----------------------+----------------------------------------------------+----------------------------------------+
| IUD_37       | GST Rate applicable   | Policy number XXXX00011,                           | GST rate applicable will be of Mysore  |
|              | for Revival Requests  |                                                    | HO                                     |
|              | Indexed & not         | Policy issuance date is 01-May-2017                |                                        |
|              | Approved (Quote       |                                                    |                                        |
|              | Section of            | Issuance office:                                   |                                        |
|              | DE/QC/Approver)       |                                                    |                                        |
|              |                       | Paid to date is 31-Jan-2018                        |                                        |
|              |                       |                                                    |                                        |
|              |                       | Frequency is monthly                               |                                        |
|              |                       |                                                    |                                        |
|              |                       | Premium 500                                        |                                        |
|              |                       |                                                    |                                        |
|              |                       | System date is 01-August-2018                      |                                        |
|              |                       |                                                    |                                        |
|              |                       | Policy status is AL                                |                                        |
|              |                       |                                                    |                                        |
|              |                       | Revival Indexed on 01 Aug 2018 ; Request is not    |                                        |
|              |                       | yet approved.                                      |                                        |
|              |                       |                                                    |                                        |
|              |                       | Revival Indexing office: Noida HO                  |                                        |
|              |                       |                                                    |                                        |
|              |                       | User who is doing Revival Indexing: Mysore HO      |                                        |
+--------------+-----------------------+----------------------------------------------------+----------------------------------------+
| IUD_38       | GST Rate applicable   | Policy number XXXX00011,                           | GST rate applicable will be of Mysore  |
|              | for Revival           |                                                    | HO                                     |
|              | Collection            | Policy issuance date is 01-May-2017                |                                        |
|              |                       |                                                    |                                        |
|              |                       | Issuance office:                                   |                                        |
|              |                       |                                                    |                                        |
|              |                       | Paid to date is 31-Jan-2018                        |                                        |
|              |                       |                                                    |                                        |
|              |                       | Frequency is monthly                               |                                        |
|              |                       |                                                    |                                        |
|              |                       | Premium 500                                        |                                        |
|              |                       |                                                    |                                        |
|              |                       | System date is 01-August-2018                      |                                        |
|              |                       |                                                    |                                        |
|              |                       | Policy status is AL                                |                                        |
|              |                       |                                                    |                                        |
|              |                       | Revival Indexed on 01 Aug 2018 ; Request is not    |                                        |
|              |                       | yet approved.                                      |                                        |
|              |                       |                                                    |                                        |
|              |                       | Revival Indexing office: Noida HO                  |                                        |
|              |                       |                                                    |                                        |
|              |                       | User ID who is doing Revival Indexing: Noida HO    |                                        |
|              |                       |                                                    |                                        |
|              |                       | User iD who is doing Revival Collection: Mysore HO |                                        |
+--------------+-----------------------+----------------------------------------------------+----------------------------------------+

**Note:** Defaults/ interest and rebates in the above calculations are
not as per actuals and the same are just for understanding of
calculation of Service Tax/GST. In system these components should be as
per actual.

## **Rejection/Termination of PROPOSAL:**

- Modal premium at the time of rejection will be considered for the
  rejection and Service tax/GST will be applied on the modal premium as
  per the rate applicable on issuance date.

  - If there is a change in tax (either service tax or GST) rate at
    issuance vis a vis rejection date the service tax/GST collected will
    not be reversed/refunded back to the customer.

  - If no premium amount is paid, there will be no reversal and no
    voucher will be generated.

  - In case of rejection/ termination/ withdrawal of proposal, Service
    Tax/ GST will not apply and premium will be refunded without
    deducting any tax/ GST

## **Collection Screen Changes:**

- Collection Screen:

The field Service/Sales tax will reflect GST amount.

Total Amount will be the amount including GST amount.

Tax Rate at the time of Collection will be considered for tax
calculation (either service tax/GST) on the Premium due amount as
determined by the paid to date selected by the user.

- Revival Screen:

There will be no changes in the screen.

- GST calculation is as per calculation formula/logic mentioned in the
  above section (3.2)

Various scenarios on how GST will be calculated is mentioned as below:

+------+----------------------------+-----------------------------------+
| **Sr | **Scenario:**              | **Rule**                          |
| No** |                            |                                   |
+:====:+============================+===================================+
| 1    | Revival Collection         | Tax Rate at the time of           |
|      |                            | Collection will be considered for |
|      |                            | tax calculation (either service   |
|      |                            | tax/GST) on the Premium due       |
|      |                            | amount as determined at the time  |
|      |                            | of approval.                      |
+------+----------------------------+-----------------------------------+
| 2    | Revival Collection with    | Tax Rate at the time of           |
|      | advance collection         | Collection will be considered for |
|      |                            | tax calculation (either service   |
|      |                            | tax/GST) on the Premium due       |
|      |                            | amount as determined by the paid  |
|      |                            | to date selected by the user.     |
|      |                            |                                   |
|      |                            | Office code where transaction is  |
|      |                            | happening and that CGST, SGST or  |
|      |                            | UTGST rate will be applicable.    |
|      |                            | This office code will be of the   |
|      |                            | user office code who is login in  |
|      |                            | the application.                  |
+------+----------------------------+-----------------------------------+

## **Claims Processing Screen Changes:**

- Death Claims:

No changes in the screen, however the tax on unpaid premium will reflect
be GST on unpaid premium is GST is effective as on the date of indexing
of claim.

- Maturity Claim

No changes in the screen, however the tax on unpaid premium will reflect
be GST on unpaid premium is GST is effective as on the date of indexing
of claim.

> GST Rate will be of that HO, SO or BO where the claim has been
> indexed. And this office code will be of the user office code who is
> login in the application and indexed the claim.

## **Policy History & Summary Screen Changes:**

- Policy search screen

No changes on the screen

- Policy details screen

> Total Amount will be the amount including Service tax/GST amount. GST
> Rate will be of the HO, SO or BO where the proposal has been indexed.
> Therefore issuance office code will be considered and the office code
> will be of the user office code who is login in the application and
> indexed the proposal.

- Collection Tab

  - The field Service/Sales tax will reflect GST amount.

  - Amount Collected will be the amount including GST amount.

- Income tax certificate:

No changes in the screen.

Field Properties of the new fields:

## **Meghdoot Upload/bulk upload:**

fields needs to be added in the bulk upload file as follows:

- First Year CGST

- Renewal Year CGST

- First Year UTGST/ SGST

- Renewal Year UTGST/SGST

**CSV fields and UI validations**

+-------------------+--------------+------------------------+-------------+-------------------------+
| **Field Name**    | **Acceptable | **Mandatory/Optional** | **Example** | **UI Error messages**   |
|                   | formats**    |                        |             |                         |
+===================+==============+========================+=============+=========================+
| First Year CGST   | Numbers up   | Optional               | Can be null | 1\. If transaction      |
|                   | to two       |                        | or numeric  | date\>= 1 Jul 2017 and  |
|                   | decimals.    |                        |             | FY CGST is not updated, |
|                   |              |                        |             | then error message to   |
|                   |              |                        |             | be displayed- "FY CGST  |
|                   |              |                        |             | tax is required"        |
|                   |              |                        |             |                         |
|                   |              |                        |             | 2\. On UI If data type  |
|                   |              |                        |             | is other than numeric   |
|                   |              |                        |             | or is blank -- then     |
|                   |              |                        |             | error message to be     |
|                   |              |                        |             | displayed "Only numeric |
|                   |              |                        |             | value accepted".        |
|                   |              |                        |             |                         |
|                   |              |                        |             | 3\. If the amount       |
|                   |              |                        |             | entered is incorrect as |
|                   |              |                        |             | required by system then |
|                   |              |                        |             | error message should be |
|                   |              |                        |             | " Incorrect First Year  |
|                   |              |                        |             | CGST"                   |
+-------------------+--------------+------------------------+-------------+-------------------------+
| Renewal_Year_CGST | Numbers up   | Optional               | Can be null | 1\. If transaction      |
|                   | to two       |                        | or numeric  | date\>=1 Jul 2017 and   |
|                   | decimals.    |                        |             | Renewal year CGST is    |
|                   |              |                        |             | not updated, then error |
|                   |              |                        |             | message to be           |
|                   |              |                        |             | displayed- "Renewal     |
|                   |              |                        |             | Year CGST tax is        |
|                   |              |                        |             | required                |
|                   |              |                        |             |                         |
|                   |              |                        |             | 2\. On UI If data type  |
|                   |              |                        |             | is other than numeric   |
|                   |              |                        |             | or is blank -- then     |
|                   |              |                        |             | error message to be     |
|                   |              |                        |             | displayed "Only numeric |
|                   |              |                        |             | value accepted"         |
|                   |              |                        |             |                         |
|                   |              |                        |             | 3\. If the amount       |
|                   |              |                        |             | entered is incorrect as |
|                   |              |                        |             | required by system then |
|                   |              |                        |             | error message should be |
|                   |              |                        |             | "Incorrect Renewal      |
|                   |              |                        |             | CGST"                   |
+-------------------+--------------+------------------------+-------------+-------------------------+
| First Year        | Numbers up   | Optional               | Can be null | 1\. If transaction      |
| UTGST/SGST        | to two       |                        | or numeric  | date\>= 1 Jul 2017 and  |
|                   | decimals.    |                        |             | FY UTGST/SGST           |
|                   |              |                        |             |                         |
|                   |              |                        |             | is not updated, then    |
|                   |              |                        |             | error message to be     |
|                   |              |                        |             | displayed- "FY          |
|                   |              |                        |             | SGST/UTGST tax is       |
|                   |              |                        |             | required"               |
|                   |              |                        |             |                         |
|                   |              |                        |             | 2\. On UI If data type  |
|                   |              |                        |             | is other than numeric   |
|                   |              |                        |             | or is blank -- then     |
|                   |              |                        |             | error message to be     |
|                   |              |                        |             | displayed "Only numeric |
|                   |              |                        |             | value accepted"         |
|                   |              |                        |             |                         |
|                   |              |                        |             | 3\. If the amount       |
|                   |              |                        |             | entered is incorrect as |
|                   |              |                        |             | required by system then |
|                   |              |                        |             | error message should be |
|                   |              |                        |             | "Incorrect First Year   |
|                   |              |                        |             | UTGST/SGST"             |
+-------------------+--------------+------------------------+-------------+-------------------------+
| Renewal_Year      | Numbers up   | Optional               | Can be null | 1\. If transaction date |
| UTGST/SGST        | to two       |                        | or numeric  | \>= 1 Jul 2017 and      |
|                   | decimals.    |                        |             | Renewal year UTGST/SGST |
|                   |              |                        |             |                         |
|                   |              |                        |             | is not updated, then    |
|                   |              |                        |             | error message to be     |
|                   |              |                        |             | displayed- "Renewal     |
|                   |              |                        |             | Year UTGST/SGST tax is  |
|                   |              |                        |             | required                |
|                   |              |                        |             |                         |
|                   |              |                        |             | 2\. On UI If data type  |
|                   |              |                        |             | is other than numeric   |
|                   |              |                        |             | or is blank -- then     |
|                   |              |                        |             | error message to be     |
|                   |              |                        |             | displayed "Only numeric |
|                   |              |                        |             | value accepted"         |
|                   |              |                        |             |                         |
|                   |              |                        |             | 3\. If the amount       |
|                   |              |                        |             | entered is incorrect as |
|                   |              |                        |             | required by system then |
|                   |              |                        |             | error message should be |
|                   |              |                        |             | " Incorrect Renewal     |
|                   |              |                        |             | Year UTGST/SGST"        |
+-------------------+--------------+------------------------+-------------+-------------------------+
| Tax_Type          | Numeric      | Conditional Mandatory  | 1 for first | 1.If FY CGST/UTGST/SGST |
|                   |              |                        | year tax    | is updated and Tax Type |
|                   |              | (Required if first     |             | is blank, then display  |
|                   |              | year tax/renewal year  | 2 for       | error message - "Tax    |
|                   |              | tax/both are present)  | renewal     | Type is required"       |
|                   |              |                        | year tax    |                         |
|                   |              |                        |             | 2 On UI If data type is |
|                   |              |                        | 3 for both  | other than numeric or   |
|                   |              |                        |             | is blank -- then error  |
|                   |              |                        |             | message to be displayed |
|                   |              |                        |             | "Only numeric value     |
|                   |              |                        |             | accepted"               |
+-------------------+--------------+------------------------+-------------+-------------------------+
| Total Receipt     | In Rupee     | Mandatory              |             | If the total Receipt    |
|                   | after round  |                        |             | amount is not an        |
|                   | off          |                        |             | combination of "total   |
|                   |              |                        |             | receipt will be         |
|                   |              |                        |             | inclusive of total      |
|                   |              |                        |             | premiums + Interest --  |
|                   |              |                        |             | Rebate + Taxes on the   |
|                   |              |                        |             | premiums" then error    |
|                   |              |                        |             | message should be       |
|                   |              |                        |             | displayed as "Incorrect |
|                   |              |                        |             | Total Receipt Amount"   |
+-------------------+--------------+------------------------+-------------+-------------------------+

1.  The total receipt will be inclusive of total premiums + Interest --
    Rebate +Taxes on the premiums

**Accounting Entries for GST:**

- New Accounts to be set up for CGST, SGST, UTGST and GST in Accounting

- Accounting Entries to be passed for CGST, SGST, UTGST and GST when the
  premium is applied towards the Policy premium due.

##  **Report Changes:**

- GSTN No. will be displayed as a header for reports generated at
  HO/NARO level and will not be displayed for reports generated at
  Divisional/Regional/Circle/National level.

[New Format]{.mark} New format for counter and customer portal receipt
to be devised including Tax components

Duplicate receipt printed from Collection History screen should have
Service tax/Sales tax value/GST (as applicable on the date of receipt
generated) added to it in Taxes field.

- HRMS Pay Deduction Premium file will undergo changes in the Request
  File where new fields will be added related to GST.

- **GL Integration CHANGES:** In the GL Integration new chart of
  accounts will be added for CGST, SGST, UTGST and GST

## **GST Number Mapping:**

GST Number Mapping to office code to be done in new solution.

## **List where UTGST/SGST is applicable:**

- The mapping will be done at HO Level.

- Mapping list will have identification.
