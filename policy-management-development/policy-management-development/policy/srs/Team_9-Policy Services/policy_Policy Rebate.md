**DEPARTMENT OF POSTS**

**MINISTRY OF COMMUNICATIONS & IT**

**GOVERNMENT OF INDIA**

**[Rebate Calculation]{.underline}**

## Brief Description

This document describes the calculation of Rebate through illustrations
for RPLI and PLI policies.

Rebate is applicable on policy when advance payment of premium is paid
by the insured. Rebate is applicable when frequency is set as monthly.
The rebate eligible criteria is different for PLI and RPLI policy as
mentioned in the document.

# Business Rules for Rebate:-

+-----------+---------------+-------------------------+----------------------+
| **BR No** | **Scenario**  | **Business Rule**       | **Illustration**     |
+===========+===============+=========================+======================+
| REBATE_01 | RPLI Products | All RPLI products       | ** **N/A             |
|           | eligible for  | eligible for rebate :-  |                      |
|           | rebate        |                         |                      |
|           |               +-------------------------+                      |
|           |               | 1.Gram Suraksha (RWLA)  |                      |
|           |               +-------------------------+                      |
|           |               | 2.Gram Santosh (REA)    |                      |
|           |               +-------------------------+                      |
|           |               | 3.Gram Suvidha ( RCWA)  |                      |
|           |               +-------------------------+                      |
|           |               | 4.Gram Sumangal (RAEA ) |                      |
|           |               +-------------------------+                      |
|           |               | 5.Gram Priya (GY)       |                      |
|           |               +-------------------------+                      |
|           |               | 6.Children Policy (RCP) |                      |
+-----------+---------------+-------------------------+----------------------+
| REBATE_02 | PLI Products  | All PLI products        | N/A                  |
|           | eligible for  | eligible for rebate :-  |                      |
|           | PLI products  |                         |                      |
|           |               +-------------------------+                      |
|           |               | 1.Suraksha (WLA)        |                      |
|           |               +-------------------------+                      |
|           |               | 2.Santosh (EA)          |                      |
|           |               +-------------------------+                      |
|           |               | 3.Suvidha(CWA)          |                      |
|           |               +-------------------------+                      |
|           |               | 4.Sumangal(AEA)         |                      |
|           |               +-------------------------+                      |
|           |               | 5.Children Policy(CP)   |                      |
|           |               +-------------------------+                      |
|           |               | 6.Yugal Suraksha(YS)    |                      |
+-----------+---------------+-------------------------+----------------------+
| REBATE_03 | Frequency     | Rebate is applicable    | 25,26,27             |
|           |               | only when frequency is  |                      |
|           |               | monthly (for both RPLI  |                      |
|           |               | & PLI )                 |                      |
+-----------+               +-------------------------+----------------------+
| REBATE_04 |               | Rebate is not           | 13 ,14 , 15          |
|           |               | applicable when         |                      |
|           |               | frequency is quarterly  |                      |
|           |               | /semi-annual /annual    |                      |
|           |               | (for both RPLI & PLI )  |                      |
+-----------+---------------+-------------------------+----------------------+
| REBATE_05 | Rebate %      | Rebate will not be      |                      |
|           | eligibility   | applicable when advance |                      |
|           | on advance    | renewal premium less    |                      |
|           | renewable     | than 3 months for RPLI  |                      |
|           | premium on    | and less than 6 months  |                      |
|           | PLI/RPLI      | for PLI policy for any  |                      |
|           | policy        | frequency               |                      |
|           |               | monthly/quarterly       |                      |
|           |               | /Semi-Annually /Annual  |                      |
+-----------+               +-------------------------+----------------------+
| REBATE_06 |               | Rebate will be provided | 25,26,27             |
|           |               | @ 0.5% when advance     |                      |
|           |               | renewable premium       |                      |
|           |               | received equals to or   |                      |
|           |               | more than 3 months but  |                      |
|           |               | less than 6 months for  |                      |
|           |               | monthly frequency only  |                      |
|           |               | for RPLI policy (not    |                      |
|           |               | applicable for PLI)     |                      |
+-----------+               +-------------------------+----------------------+
| REBATE_07 |               | Rebate will be provided | 9                    |
|           |               | @ 1% when advance       |                      |
|           |               | renewable premium       |                      |
|           |               | received equals to or   |                      |
|           |               | more than 6 months but  |                      |
|           |               | less than 12 months for |                      |
|           |               | monthly frequency (for  |                      |
|           |               | both RPLI & PLI )       |                      |
+-----------+               +-------------------------+----------------------+
| REBATE_08 |               | Rebate will be provided |                      |
|           |               | @ 2% when advance       |                      |
|           |               | renewable premium       |                      |
|           |               | received equals to or   |                      |
|           |               | more than 12 months for |                      |
|           |               | monthly frequency (for  |                      |
|           |               | both RPLI & PLI )       |                      |
+-----------+---------------+-------------------------+----------------------+
| REBATE_9  | Rebate        | Any advance initial     | 29                   |
|           | eligibility   | premium received for 3  |                      |
|           | on Initial    | months or more in RPLI  |                      |
|           | Premium on    | and 6 months or more in |                      |
|           | RPLI /PLI     | PLI then rebate will    |                      |
|           | policy        | not be provided for     |                      |
|           |               | monthly frequency only  |                      |
|           |               | and rebate will not be  |                      |
|           |               | applicable for          |                      |
|           |               | quarterly               |                      |
|           |               | /semi-annually/annually |                      |
|           |               | frequency (for both     |                      |
|           |               | RPLI & PLI )            |                      |
+-----------+---------------+-------------------------+----------------------+
| REBATE_10 | Rebate        | **For all PLI /RPLI     | 17,18,19,20,21,      |
|           | Calculation   | (except Yugal           | 22,23,25,26,27       |
|           |               | Suraksha)**             |                      |
|           |               +-------------------------+                      |
|           |               | Rebate = Number of      |                      |
|           |               | advance premiums paid   |                      |
|           |               | from current month till |                      |
|           |               | To date \* Modal        |                      |
|           |               | Premium Amount  \*      |                      |
|           |               | Rebate %                |                      |
|           |               +-------------------------+                      |
|           |               |                         |                      |
|           |               +-------------------------+                      |
|           |               | **Yugal Suraksha :-**   |                      |
|           |               +-------------------------+                      |
|           |               | Rebate is calculated    |                      |
|           |               | basis the number of     |                      |
|           |               | premium paid in advance |                      |
|           |               | from current month till |                      |
|           |               | to date                 |                      |
|           |               +-------------------------+                      |
|           |               | Rebate for Yugal        |                      |
|           |               | Suraksha = One month    |                      |
|           |               | premium amount\*Rebate  |                      |
|           |               | %                       |                      |
|           |               +-------------------------+                      |
|           |               |                         |                      |
|           |               +-------------------------+                      |
|           |               | **Note :**              |                      |
|           |               +-------------------------+                      |
|           |               | Rebate % mentioned in   |                      |
|           |               | the Rebate % grid       |                      |
+-----------+---------------+-------------------------+----------------------+
| REBATE_11 | Rebate        | Rebate and Interest     |  32                  |
|           | eligible on   | together will be        |                      |
|           | revival of    | applicable on revival   |                      |
|           | lapse policy  | of lapse policy; when   |                      |
|           |               | advance premium for     |                      |
|           |               | 3months (RPLI)/ 6months |                      |
|           |               | (PLI) or more is        |                      |
|           |               | received on lapse       |                      |
|           |               | policy from current     |                      |
|           |               | month for monthly       |                      |
|           |               | frequency.              |                      |
+-----------+---------------+-------------------------+----------------------+
| REBATE_12 | Pay-recovery  | Rebate will be          |                      |
|           |               | applicable for pay      |                      |
|           |               | recovery policies if    |                      |
|           |               | done through            |                      |
|           |               | collections modules per |                      |
|           |               | rebate PLI/RPLI policy  |                      |
|           |               | rules                   |                      |
+-----------+---------------+-------------------------+----------------------+
| REBATE_13 | Meghdoot      | Rebate validations will | 28                   |
|           | Upload        | be not be applicable    |                      |
|           |               | for meghdoot upload     |                      |
|           |               | (for both RPLI & PLI )  |                      |
+-----------+---------------+-------------------------+----------------------+
| REBATE_14 | Multiple      | Rebate will not be      | 30,31                |
|           | receipts      | applicable in case if   |                      |
|           |               | advance premium is paid |                      |
|           |               | in multiple             |                      |
|           |               | receipts(for both RPLI  |                      |
|           |               | & PLI )                 |                      |
+-----------+---------------+-------------------------+----------------------+
| REBATE_15 | Rebate %      | Rebate percentages with | Refer Rebate % grid  |
|           | should be     | respect to frequency    |                      |
|           | configurable  | for PLI/RPLI need to be |                      |
|           |               | configured and should   |                      |
|           |               | not be hardcoded as     |                      |
|           |               | mentioned in rebate %   |                      |
|           |               | grid                    |                      |
+-----------+---------------+-------------------------+----------------------+
| REBATE_16 | Taxes         | Taxes not applicable on |                      |
|           |               | rebate (for both RPLI & |                      |
|           |               | PLI )                   |                      |
+-----------+---------------+-------------------------+----------------------+
| REBATE_17 | Receipt       | Rebate will get         |                      |
|           | Cancellation  | reversed at the time of |                      |
|           |               | receipt cancellation    |                      |
|           |               | (for both RPLI & PLI )  |                      |
+-----------+---------------+-------------------------+----------------------+
| REBATE_18 | Rounding off  | Rounding off logic will |                      |
|           |               | continue as per         |                      |
|           |               | existing process ,if    |                      |
|           |               | value contains less     |                      |
|           |               | than .5 than nearest    |                      |
|           |               | lowest value will be    |                      |
|           |               | considered and if value |                      |
|           |               | contains equals to or   |                      |
|           |               | more than .5 than       |                      |
|           |               | nearest highest value   |                      |
|           |               | will be considered(for  |                      |
|           |               | both RPLI & PLI )       |                      |
+-----------+---------------+-------------------------+----------------------+
| REBATE-19 | Configuration | All parameters related  |                      |
|           |               | to Rebate based on sum  |                      |
|           |               | assured and premium     |                      |
|           |               | should be configurable  |                      |
|           |               | in UIs                  |                      |
+-----------+---------------+-------------------------+----------------------+

## Rebate Percentage Grid:

- Rebate percentage grid basis the PLI and RPLI policies and frequency

RPLI /PLI rebate percentage need to be configured and should not be
hardcoded basis the below grid:-

+----------+-------------+------------+-----------------------------------+------------+----------+
| **S.No** | **Carrier** | **Product  | **Period                          | **Advance  | **Rebate |
|          |             | Name**     | (Monthly/Quarterly/Semi-annually/ | Premium    | %**      |
|          |             |            | Annually )**                      | Duration** |          |
|          +-------------+            |                                   |            |          |
|          | **(         |            |                                   |            |          |
|          | PLI/RPLI    |            |                                   |            |          |
|          | )**         |            |                                   |            |          |
+==========+=============+============+===================================+============+==========+
| 1        | PLI         | Suraksha\  | Monthly                           | 0 to 5     | Nil      |
|          |             | Santosh\   |                                   |            |          |
|          |             | Suvidha\   |                                   |            |          |
|          |             | Children   |                                   |            |          |
|          |             | Policy\    |                                   |            |          |
|          |             | Sumangal   |                                   |            |          |
+----------+-------------+            +-----------------------------------+------------+----------+
| 2        | PLI         |            | Monthly                           | 6 to 11    | 1%       |
+----------+-------------+            +-----------------------------------+------------+----------+
| 3        | PLI         |            | Monthly                           | 12 and     | 2%       |
|          |             |            |                                   | above      |          |
+----------+-------------+------------+-----------------------------------+------------+----------+
| 4        | PLI         | Santosh    | Quarterly/Semi-annually/annually  | 0 to 5     | Nil      |
+----------+-------------+            |                                   +------------+----------+
| 5        | PLI         |            |                                   | 6 to 11    | Nil      |
+----------+-------------+            |                                   +------------+----------+
| 6        | PLI         |            |                                   | 12 and     | Nil      |
|          |             |            |                                   | above      |          |
+----------+-------------+------------+-----------------------------------+------------+----------+
| 7        | PLI         | Yugal      | Monthly                           | 0 to 2     | Nil      |
|          |             | Suraksha   |                                   |            |          |
|          |             |            |                                   |            |          |
|          |             | (Rebate    |                                   |            |          |
|          |             | calculated |                                   |            |          |
|          |             | on one     |                                   |            |          |
|          |             | month      |                                   |            |          |
|          |             | premium    |                                   |            |          |
|          |             | amount)    |                                   |            |          |
+----------+-------------+            +-----------------------------------+------------+----------+
| 8        | PLI         |            | Monthly                           | 3 to 5     | 2%       |
+----------+-------------+            +-----------------------------------+------------+----------+
| 9        | PLI         |            | Monthly                           | 6 to 11    | 10%      |
+----------+-------------+            +-----------------------------------+------------+----------+
| 10       | PLI         |            | Monthly                           | 12 and     | 50%      |
|          |             |            |                                   | above      |          |
+----------+-------------+------------+-----------------------------------+------------+----------+
| 11       | PLI         | Yugal      | Quarterly/Semi-annually/annually  | 3 to 5     | Nil      |
|          |             | Suraksha   |                                   |            |          |
+----------+-------------+            |                                   +------------+----------+
| 12       | PLI         |            |                                   | 6 to 11    | Nil      |
+----------+-------------+            |                                   +------------+----------+
| 13       | PLI         |            |                                   | 12 and     | Nil      |
|          |             |            |                                   | above      |          |
+----------+-------------+------------+-----------------------------------+------------+----------+
| 14       | RPLI        | Gram       | Monthly                           | 0 to 2     | Nil      |
|          |             | Suraksha\  |                                   |            |          |
|          |             | Gram       |                                   |            |          |
|          |             | Santosh\   |                                   |            |          |
|          |             | Gram       |                                   |            |          |
|          |             | Suvidha\   |                                   |            |          |
|          |             | Gram       |                                   |            |          |
|          |             | Sumangal\  |                                   |            |          |
|          |             | Gram       |                                   |            |          |
|          |             | Priya\     |                                   |            |          |
|          |             | Children   |                                   |            |          |
|          |             | Policy     |                                   |            |          |
+----------+-------------+            +-----------------------------------+------------+----------+
| 15       | RPLI        |            | Monthly                           | 3 to 5     | 0.5%     |
+----------+-------------+            +-----------------------------------+------------+----------+
| 16       | RPLI        |            | Monthly                           | 6 to 11    | 1%       |
+----------+-------------+            +-----------------------------------+------------+----------+
| 17       | RPLI        |            | Monthly                           | 12 and     | 2%       |
|          |             |            |                                   | above      |          |
+----------+-------------+------------+-----------------------------------+------------+----------+
| 18       | RPLI        | Gram       | Quarterly/Semi-annually/Annually  | 0 to 2     | Nil      |
|          |             | Suraksha\  |                                   |            |          |
|          |             | Gram       |                                   |            |          |
|          |             | Santosh\   |                                   |            |          |
|          |             | Gram       |                                   |            |          |
|          |             | Suvidha\   |                                   |            |          |
|          |             | Gram       |                                   |            |          |
|          |             | Sumangal\  |                                   |            |          |
|          |             | Gram Priya |                                   |            |          |
+----------+-------------+            |                                   +------------+----------+
| 19       | RPLI        |            |                                   | 3 to 5     | Nil      |
+----------+-------------+            |                                   +------------+----------+
| 20       | RPLI        |            |                                   | 6 to 11    | Nil      |
+----------+-------------+            |                                   +------------+----------+
| 21       | RPLI        |            |                                   | 12 and     | Nil      |
|          |             |            |                                   | above      |          |
+----------+-------------+------------+-----------------------------------+------------+----------+

# Illustrations

1.  System will provide Rebate only when advance premium for 6 months or
    more s done for PLI policy and Policy status is AP and frequency is
    monthly

2.  System will provide Rebate and Interest together when advance
    premium for 6 months or more is done for PLI policy and Policy
    status is Lapse and frequency is monthly

3.  If insurant paying premium for six month or more than six month in
    r/o quarterly/ half yearly/ annually frequency then system should
    not provide Rebate. Rebate is allowed for monthly frequency only for
    both PLI /RPLI policy

4.  **Scenario --**Rebate on advance Premium of 3 months or more and
    less than 6 months for RPLI Products is 0.5 % , this is not
    applicable for PLI policies

5.  **Scenario --**Rebate on advance Premium of 6 months or more and
    less than 12 months for PLI and RPLI Products is 1 % for policies
    other than Yugal Suraksha.

6.  **Scenario --** Rebate on advance Premium of 12 months for PLI and
    RPLI Products is 2 % for policies other than Yugal Suraksha

7.  The rebate percentages with respect to frequency for PLI / RPLI
    policy need to be configured and should not be hardcoded

8.  System will not provide Rebate when advance premium for 5 months is
    given by customer and frequency is monthly.

**Result-** No Rebate is provided to customer. System should display
Premium Rebate as Rs 0.00

  ------------------------------------------------------------------------------------
  Policy Number   From Date     To Date    Action
  -------------- ------------ ------------ -------------------------------------------
   TN-500110-CS   01/01/2025   31/05/2025  No Rebate is provided to customer. System
                                           should display Premium Rebate as Rs0.00

  ------------------------------------------------------------------------------------

9.  System will provide Rebate when advance premium for 6 months is
    given by customer and frequency is monthly.

**Result-** Rebate is provided to customer.

  -------------------------------------------------------------------------
  Policy Number   From Date     To Date    Action
  -------------- ------------ ------------ --------------------------------
   TN-500110-CS   01/06/2025   30/11/2025  Rebate is provided to customer.

  -------------------------------------------------------------------------

10. System will provide Rebate when advance premium for more than 6
    months is given by customer and Advance Premium is considered from
    the current month.

**Result-** Rebate is provided to customer when the Advance Premium is
considered from the current month and calculation will be taken as
(Monthly Premium \* No of monthly advance premiums from current month
till the To date) chosen by user.

  -------------------------------------------------------------------------------------------
  Policy Number  Current      From Date    To Date      No. of      Monthly     Action
                 Month                                  monthly     Premium     
                                                        advance                 
                                                        premiums                
  -------------- ------------ ------------ ------------ ----------- ----------- -------------
   TN-500110-CS  01/09/2025   01/09/2025   28/02/2026   6           130         Rebate is
                                                                                provided to
                                                                                customer when
                                                                                the Advance
                                                                                Premium is
                                                                                considered
                                                                                from the
                                                                                current month
                                                                                and
                                                                                calculation
                                                                                will be taken
                                                                                as Monthly
                                                                                Premium \* No
                                                                                of monthly
                                                                                advance
                                                                                premiums from
                                                                                current month
                                                                                till the To
                                                                                date chosen
                                                                                by user.

  -------------------------------------------------------------------------------------------

11. System will not provide Rebate when advance premium of 5 months is
    given by customer and Advance Premium is considered from the current
    month.

**Result-** Rebate is not provided to customer.

12. System will not provide Rebate when advance premium for more than 6
    months is given by customer and Advance Premium is considered from
    the due month.

**Result-** Rebate is not provided to customer when the Advance Premium
is considered from the due month.

  ------------------------------------------------------------------------------------
  Policy Number  Current Date  From Date     To Date    Action
  -------------- ------------ ------------ ------------ ------------------------------
   TN-500110-CS   06/10/2025   01/09/2025   28/02/2026  Rebate is not provided to
                                                        customer when the Advance
                                                        Premium is considered from the
                                                        due month.

  ------------------------------------------------------------------------------------

13. System will not provide Rebate when advance premium for 11 months is
    given by customer and frequency is Quarterly.

**Result-** No Rebate is provided to customer. System should display
Premium Rebate as Rs 0.00

  -----------------------------------------------------------------------------------
  Policy Number   Frequency   From Date     To Date   Action
  -------------- ----------- ------------ ----------- -------------------------------
   TN-500110-CS   Quarterly   01/06/2025   30/4/2026  No Rebate is provided to
                                                      customer. System should display
                                                      Premium Rebate as Rs 0

  -----------------------------------------------------------------------------------

14. System will not provide Rebate when advance premium for 12 months is
    given by customer and frequency is annually.

**Result-** No Rebate is provided to customer. System should display
Premium Rebate as Rs 0.00.

  --------------------------------------------------------------------------------
  Policy Number   Frequency   From Date     To Date    Action
  -------------- ----------- ------------ ------------ ---------------------------
   TN-500110-CS    Annual     01/06/2025   30/06/2026  No Rebate is provided to
                                                       customer. System should
                                                       display Premium Rebate as
                                                       Rs 0

  --------------------------------------------------------------------------------

15. System will not provide Rebate when advance premium for 6 months is
    given by customer and frequency is Half-yearly

**Result-** No Rebate is provided to customer. System should display
Premium Rebate as Rs 0.00.

  --------------------------------------------------------------------------------
  Policy Number   Frequency   From Date     To Date    Action
  -------------- ----------- ------------ ------------ ---------------------------
   TN-500110-CS  Half Yearly  01/06/2025   31/12/2025  No Rebate is provided to
                                                       customer. System should
                                                       display Premium Rebate as
                                                       Rs 0

  --------------------------------------------------------------------------------

16. System will not consider calculation of Rebate for the policies for
    which collection is done via Meghdoot Upload.

**Result**- If any of the collection done for 6 months from Meghdoot
Upload then that policy will not be eligible for Rebate.

17. Scenario -- Rebate on advance premiums for Yugal Suraksha Policies
    --

Result-

Premium month is greater or equal to 3 month and less than 6 month   =
2% rebate on one month premium amount

Premium month is greater or equal to 6 month and less than 12 month =
10% rebate on one month premium amount

Premium month is greater or equal to 12 month                           
                = 50% rebate on one month premium amount

18. Scenario -- Rebate on advance premiums for the months greater than
    or equal to 3 months and less than 6 months for YS Policies.

Result-

  ----------------------------------------------------------------------------
    Policy Number    From Date     To Date    Action
  ----------------- ------------ ------------ --------------------------------
   TN-YS-500110-CS   01/10/2025   31/12/2025  Rebate is provided to customer.
                                              It is calculated as 2% rebate on
                                              one month premium amount.

  ----------------------------------------------------------------------------

19. Scenario -- Rebate on advance premiums for the months greater than
    or equal to 3 months and less than 6 months for YS Policies.

Result-

  ----------------------------------------------------------------------------
    Policy Number    From Date     To Date    Action
  ----------------- ------------ ------------ --------------------------------
   TN-YS-500110-CS   01/06/2025   31/10/2025  Rebate is provided to customer.
                                              It is calculated as 2% rebate on
                                              one month premium amount.

  ----------------------------------------------------------------------------

20. Scenario -- Rebate on advance premiums for the months less than 3
    months for YS Policies.

Result-

  ----------------------------------------------------------------------------
    Policy Number    From Date     To Date    Action
  ----------------- ------------ ------------ --------------------------------
   TN-YS-500110-CS   01/10/2025   30/11/2025  No Rebate is provided to
                                              customer

  ----------------------------------------------------------------------------

21. Scenario -- Rebate on advance premiums for the months greater than
    or equal to 6 months and less than 12 months for YS Policies.

Result-

  ----------------------------------------------------------------------------
    Policy Number    From Date     To Date    Action
  ----------------- ------------ ------------ --------------------------------
   TN-YS-500110-CS   01/06/2025   30/11/2025  Rebate is provided to customer.
                                              Calculated as 10% rebate on one
                                              month premium amount.

  ----------------------------------------------------------------------------

22. Scenario -- Rebate on advance premiums for the months greater than
    or equal to 6 months and less than 12 months for YS Policies.

Result-

  ----------------------------------------------------------------------------
    Policy Number    From Date     To Date    Action
  ----------------- ------------ ------------ --------------------------------
   TN-YS-500110-CS   01/06/2025   30/04/2026  Rebate is provided to
                                              customer.Calculated as 10%
                                              rebate on one month premium
                                              amount.

  ----------------------------------------------------------------------------

23. Scenario -- Rebate on advance premiums for the months greater than
    and equal to 12 months for YS Policies.

Result-

  ----------------------------------------------------------------------------
    Policy Number    From Date     To Date    Action
  ----------------- ------------ ------------ --------------------------------
   TN-YS-500110-CS   01/06/2025   31/05/2026  Rebate is provided to customer.
                                              ICalculated as 50% rebate on one
                                              month premium amount.

  ----------------------------------------------------------------------------

## IUD's For RPLI (Rebate for RPLI for 3 months as 0.5 %)

24. For RPLI policy**,** System will not provide Rebate when advance
    premium for \<=2 months is given by customer and frequency is
    monthly.

**Result-** No Rebate is provided to customer. System should display
Premium Rebate as Rs 0

  ----------------------------------------------------------------------------------
   Policy Number   Current Date  From Date     To Date    Result
  ---------------- ------------ ------------ ------------ --------------------------
   R-TN-EA-123456   01/11/2025   01/10/2025   31/12/2025  No Rebate is provided to
                                                          customer. System should
                                                          display Premium Rebate as
                                                          Rs0 and Interest to be
                                                          collected

  ----------------------------------------------------------------------------------

25. For RPLI policy ,System will provide Rebate when advance premium for
    \>=3 and \<6 months is given by customer and frequency is monthly
    and premium amount is Rs 1000

**Result-** Rebate is provided to the customer \@0.5% on 3 months
advance premium received.

+---------------+------------+------------+------------+-------------------------------+
| Policy Number | Current    | From Date  | To Date    | Result                        |
|               | Date       |            |            |                               |
+:=============:+:==========:+:==========:+:==========:+===============================+
| 0000000123456 | 01/11/2025 | 01/10/2025 | 31/01/2026 | Interest to be collected and  |
|               |            |            |            |                               |
|               |            |            |            | Rebate is provided to         |
|               |            |            |            | customer \@0.5%.              |
|               |            |            |            |                               |
|               |            |            |            | 3 months advance Premium paid |
|               |            |            |            | from Nov -25 till Jan -26     |
|               |            |            |            | ,Current Month Collection is  |
|               |            |            |            | also included during Rebate   |
|               |            |            |            | Calculation                   |
|               |            |            |            |                               |
|               |            |            |            | **Computation logic :-**      |
|               |            |            |            |                               |
|               |            |            |            | Rebate calculated on premium  |
|               |            |            |            | = 1000\*3=3000                |
|               |            |            |            |                               |
|               |            |            |            | Rebate @ 0.5% = 3000\*0.5% =  |
|               |            |            |            | 15                            |
+---------------+------------+------------+------------+-------------------------------+

26. For RPLI policy ,System will provide Rebate when advance premium for
    \>=3 and \<6 (4 months) is given by customer and frequency is
    monthly and premium amount is Rs 1000

**Result-** Rebate is provided to the customer \@0.5% on 4 months
advance premium received.

+---------------+------------+----------+----------+----------------------------+
| Policy Number | Current    | From     | To Date  | Result                     |
|               | Date       | Date     |          |                            |
+:=============:+:==========:+:========:+:========:+============================+
| 0000001234567 | 01/11/2025 | 01/10/25 | 28/02/26 | Interest to be collected   |
|               |            |          |          | and                        |
|               |            |          |          |                            |
|               |            |          |          | Rebate is provided to      |
|               |            |          |          | customer \@0.5%.           |
|               |            |          |          |                            |
|               |            |          |          | 4 months advance Premium   |
|               |            |          |          | paid from Nov -25 till Feb |
|               |            |          |          | -26                        |
|               |            |          |          |                            |
|               |            |          |          | **Computation logic :-**   |
|               |            |          |          |                            |
|               |            |          |          | Rebate calculated on       |
|               |            |          |          | premium = 1000\*4=4000     |
|               |            |          |          |                            |
|               |            |          |          | Rebate @ 0.5% = 4000\*0.5% |
|               |            |          |          | = 20                       |
+---------------+------------+----------+----------+----------------------------+

27. For RPLI policy ,System will provide Rebate when advance premium for
    \>=3 and \<6 (5 months) is given by customer and frequency is
    monthly and premium amount is Rs 1000

**Result-** Rebate is provided to the customer \@0.5% on 5 months
advance premium received.

+---------------+------------+------------+------------+----------------------------+
| Policy Number | Current    | From Date  | To Date    | Result                     |
|               | Date       |            |            |                            |
+:=============:+============+:==========:+:==========:+============================+
| 0000001213178 | 01/11/2025 | 01/10/2025 | 31/03/2026 | Interest to be collected   |
|               |            |            |            | and                        |
|               |            |            |            |                            |
|               |            |            |            | Rebate is provided to      |
|               |            |            |            | customer \@0.5%.           |
|               |            |            |            |                            |
|               |            |            |            | 5 months advance Premium   |
|               |            |            |            | paid from Nov-25till       |
|               |            |            |            | Mar-26                     |
|               |            |            |            |                            |
|               |            |            |            | **Computation logic :-**   |
|               |            |            |            |                            |
|               |            |            |            | Rebate calculated on       |
|               |            |            |            | premium = 1000\*5=5000     |
|               |            |            |            |                            |
|               |            |            |            | Rebate @ 0.5% = 5000\*0.5% |
|               |            |            |            | = 25                       |
+---------------+------------+------------+------------+----------------------------+

28. System will not consider calculation of Rebate for the policies for
    which collection is done via Meghdoot Upload.

**Result**- If any of the collection done from Meghdoot Upload then that
policy will be eligible for Rebate entered by user in the upload file
.There will be no system validation on the amount entered.

29. System will not provide Rebate when initial premium collected in
    advance for \>=3 and \<6 (for 3 months )is given by customer for
    monthly frequency for RPLI policy

**Result-** Rebate is not provided to customer when initial premium
received in advance for RPLI policy.

Modal Premium: 1000

Amount Collected on current date: 3000

30. System will not provide Rebate when renewal premium collected
    through multiple receipt for modal premium 1000

**Result-** Rebate is not provided to customer when renewal premium
received in advance through multiple receipt.

+--------------+-----------+-----------+------------+---------+--------+------------------+
| Policy       | Current   | From Date | To Date    | Receipt | Amount | Action           |
| Number       | Date      |           |            | Number  |        |                  |
+:============:+:=========:+:=========:+:==========:+:=======:+:======:+==================+
| TN-500110-CS | 1/11/2025 | 1/10/2025 | 30/11/2025 | 123456  | 2000   | No Rebate will   |
|              |           |           |            |         |        | be given when    |
|              |           |           |            |         |        | premium received |
|              |           |           |            |         |        | through multiple |
|              |           |           |            |         |        | receipt          |
+--------------+-----------+-----------+------------+---------+--------+                  |
| TN-500110-CS | 1/11/2025 | 1/12/2025 | 31/3/2026  | 2345678 | 2000   |                  |
+--------------+-----------+-----------+------------+---------+--------+------------------+

31. System will provide Rebate when renewal premium collected through
    single receipt for modal premium 1000

**Result-** Rebate is provided to customer when renewal premium received
in advance through single receipt.

  -----------------------------------------------------------------------------------------
  Policy Number    Current    From Date    To Date    Receipt   Amount  Action
                    Date                              Number            
  -------------- ----------- ----------- ----------- --------- -------- -------------------
   TN-500110-CS   1/11/2025   1/10/2025   31/1/2026   2114603    4000   Rebate will be
                                                                        given when premium
                                                                        received through
                                                                        single receipt on
                                                                        Rs 3000i.e from
                                                                        Nov'25 till Jan '26

  -----------------------------------------------------------------------------------------

32. System will provide Rebate and interest when advance premium for 3
    months is given by customer for lapse policy

**Result-** Rebate and interest both will be provided to customer.

+--------------+----------+---------+------------+------------+---------------------+
| Policy       | Current  | Policy  | From Date  | To Date    | Action              |
| Number       | Date     | Status  |            |            |                     |
+:============:+==========+:=======:+:==========:+:==========:+=====================+
| TN-500110-CS | 01/11/25 | IL/VL   | 01/04/2025 | 31/03/2026 | Rebate and interest |
|              |          |         |            |            | both will be        |
|              |          |         |            |            | applicable          |
|              |          |         |            |            |                     |
|              |          |         |            |            | Interest applicable |
|              |          |         |            |            | : From April'25     |
|              |          |         |            |            | till Oct '25        |
|              |          |         |            |            |                     |
|              |          |         |            |            | Rebate Applicable : |
|              |          |         |            |            | From Nov-25 till    |
|              |          |         |            |            | Mar26 equals to 5   |
|              |          |         |            |            | months , customer   |
|              |          |         |            |            | eligible for rebate |
|              |          |         |            |            | for this period     |
+--------------+----------+---------+------------+------------+---------------------+

**Rebate on Sum assured**

Premium Rebates on Sum assured for evaluating Premium at the time New
Business

  -----------------------------------------------------------------------
  **Band**              **Rebate**        **Applicable Products**
  --------------------- ----------------- -------------------------------
  Upto 20,000           Nil               All PLI products

  Above 20,000          Re. 1 per 20,000  All PLI products except Yugal
                        of SA             Suraksha

  Upto 40,000           Nil               PLI Yugal Suraksha

  Above 40,000          Re. 1 per 10,000  PLI Yugal Suraksha
                        of SA             
  -----------------------------------------------------------------------

In respect of Quarterly ,Half yearly, Annual policies, Rebate on the sum
assured has to be given based on the frequency and number of months
involved

Example

Term 10 Years

Husband Age = 30 Years

Wife Age = 25 Years

Difference Age = 5 Years

Addition to be made to lower age as per table = 3

Age at entry = 25+3 = 28 Years

Tabular premium = 93

Monthly per 10,000 SA

Premium for 90,000 SA = ((90000/10000)\*93) = 837

Sum Assured Rebate = 40000: Re. 1

90000 -- 40000 = 50000

(50000/10000) = 5

Total SA Rebate = (5+1) = 6

Monthly premium = 837 -- 6 = 831

Quarterly premium = (831\*3) -- (831\*(2/100))

= 2493 -- 16.62 = 2476.38

Half yearly premium = (831\*6) -- (831\*(10/100))

= 4986 -- 83.1 = 4902.90

Yearly = (831\*12) -- (831\*(50/100))

= 9972 -- 415.50 = 9556.50
