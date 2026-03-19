Postal Life Insurance

Department of Post

**Software Requirements Specification (SRS)**

**Policy surrender**

1.  **Executive Summary**

This document outlines the business requirements for Policy Surrender
with Illustrations of Insurance Management System (IMS). It aims to
processing of Surrendering of a Policy as per the ruling provided .

2.  **Business Objectives**

The IMS project addresses the requirement of surrendering of a Policy by
the Insurant and when required . This system will index the surrender
request, process and effect the payments .

3.  **Project Scope**

**Policy surrender will be a module of the Insurance Admin system used
by Insurant/Customer as and when required and unwilling to continue the
policy .**

**4. Business Requirements**

4.  **Functional Requirements Specification**

Surrender value of a policy, is the amount that is payable to the
Insurant, when he/she foregoes the contingent benefit of his policy and
surrenders it for an immediate payment. Surrender value depends on the
surrender factor and type and term of policy and as per the prescribed
ruling mentioned in POLI

**Business Logic**

1.  Policy should be in force

2.  Policy should have been completed 36 months.

3.  If there is any defaults of Premium (for less than 3 years policy, 6
    months and for more than 3 years 12 months grace period), premium
    should be deducted with Interest amount from surrender value along
    with any other deductions as applicable

4.  For policy attained Maturity , should be treated for Maturity
    process only

5.  For WLA policies(PLI & RPLI) beyond maturity period eligible for
    Surrender

6.  AEA & GY policy is not eligible for Surrender

7.  If the defaults are other monthly mode like Quarterly, Half annually
    & Annually premium is to calculated for Monthly and proportionate
    amount is to be deducted along with Interest and other deductions as
    applicable.

8.  Necessary documents written consent , policy bond, premium receipt
    book,Pay recovery certificate . Loan receipt book,loan bond etc ,
    indeminity bond for the lost policy are to be displayed in the
    system and obtained from customer while processing.

**Calculation Logic**

**Cumulative age**: Date of application for surrender-Date of
birth(15days or more will be counted as 1 month)

**Paid_up_value**:(Number of Premiums Paid\*sum assured amount)/Total
number of Premiums Payable

**Surrender value**: Paid-up Value + Bonus)\* Surrender
Factor(Ref:Surrender factor table)-Unpaid Premium-Unpaid Loan Principle
and Interest

**Surrender Paid in Advance**

For policies in respect of which premium is paid annually in advance,
surrender value is to be calculated at the end of the year irrespective
of the date of surrender but payment of surrender value may be made when
the policy holder asks for it.

**Surrender Date calculation in monthly mode:**

Policy shall continue to be in force till the end of the month in which
the application for surrender is received.Accordingly the premium shall
also be payable for the period for which the policy continues to be in
force.

**Surrender Date calculation in Annual mode:-**

Surrender value is to be calculated at the end of the year irrespective
of the date of surrender but payment of surrender value may be made when
the policy holder asks for it.

**Bonus: Nil for less than 5 years**

For more than 5 years : - proportionate to the reduced SA

Bonus calculation on paid-up amount :- (Bonus rate\*Paid Up value)/1000
.

**Payment: Paid to Insurant or Assignee for assigned policy**

  --------------------------------------------------------------------------
  [Product            [Product Name]{.mark}                  [Surrender
  Type]{.mark}                                               Allowed
                                                             (Y/N)]{.mark}
  ------------------- -------------------------------------- ---------------
  Whole Life          Suraksha- Whole Life Assurance         Yes, after
  Assurance                                                  premium payment
                                                             of 4 yrs.

  Endowment Assurance Santosh-Endowment Assurance            Yes, after
                                                             premium payment
                                                             of 3 yrs.

  Convertible Whole   Suvidha - Convertible Whole Life       Yes, after
  Life Assurance      Assurance                              premium payment
                                                             of 4 yrs.

  Anticipated         Sumangal - Anticipated Endowment       No
  Endowment Assurance Assurance                              

  Endowment           Child Policy                           Yes, after
                                                             premium payment
                                                             of 5 yrs.

  Endowment Assurance Yugal Suraksha - Joint Life Assurance  Yes, after
                                                             premium payment
                                                             of 3 yrs.
  --------------------------------------------------------------------------

RPLI:

  -----------------------------------------------------------------------
      Product Type                  Product Name              Surrenders
                                                               Allowed
                                                                (Y/N)
  --------------------- ------------------------------------ ------------
  Whole Life Assurance  Gram Suraksha- Whole Life Assurance   Yes, after
                                                               premium
                                                             payment of 4
                                                                 yrs.

   Endowment Assurance    Gram Santosh-Endowment Assurance    Yes, after
                                                               premium
                                                             payment of 3
                                                                 yrs.

    Convertible Whole     Gram Suvidha - Convertible Whole    Yes, after
     Life Assurance                Life Assurance              premium
                                                             payment of 4
                                                                 yrs

  Anticipated Endowment     Gram Sumangal - Anticipated      no Surrender
        Assurance               Endowment Assurance          

        Endowment                   Child Policy              Yes, after
                                                               premium
                                                             payment of 5
                                                                 yrs.

   10 year Anticipated               Gram Priya              No Surrender
   Endowment Assurance                                       

                                                             
  -----------------------------------------------------------------------

**Quote screen:**

Quote Type: Surrender

Policy Number:

Current date

**Display Screen:**

Policy Number:

Name:

Policy status:

Customer Id:

Product Name:

Policy issue date:

Paid till date:

Duplicate Policy bond:

![](media/image1.png){width="6.268055555555556in"
height="0.7722222222222223in"}

**After clicking Get quote:**

![A screenshot of a document AI-generated content may be
incorrect.](media/image2.png){width="6.268055555555556in"
height="5.227083333333334in"}

Bonus details:

![A table of numbers with numbers AI-generated content may be
incorrect.](media/image3.png){width="6.268055555555556in"
height="5.5625in"}

**Surrender claim Indexing**

Policy Number :

Service request Type(drop Menu): Surrender Claim

**Display : As in Quote mentioned above**

**After submitting: Generating Service Request Number after checking all
the conditions and rulings on the policy**

**Surrender Process At CPC**

### 

### Calculation of Surrender Value(example)

Plan Type:EA

Entry Date:01-Jan-2007

Age at Entry:33

Age at Maturity: 60

Date of Birth:10-Jan-1974

SA:20000

Bonus: 10000

Premium paid upto :31-Dec-2025

Date of Application:31-Dec-2025

Premium:100

Paid Up Value:20000x(18x12)/((60-33)x12)=20000x216/324=13333

Surrender Factor Value:0.725(not exact value)

Surrender Value: (13333+10000)\*0.725

Net Value: 16916
