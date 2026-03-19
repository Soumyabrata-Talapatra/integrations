Postal Life Insurance

Department of Post

**Software Requirements Specification (SRS)**

**Policy Commutation**

1.  **Executive Summary**

> This document is detailing the functionality of Commutation of a
> Policy and its impact on other functionalities.

2.  **Business Objectives**

The IMS project addresses the requirement of commuting a Policy by the
Insurant as and when required for reducing the premium. This system will
index the policy as Commutation request, process and effect the change
of renewals from the effective period as per the guidelines provided by
Business Division

3.  **Project Scope**

**Policy Commutation will be a module of the Insurance Admin system used
by Insurant/Customer as and when required for reducing the premium
amount paid to continue the policy .**

4.  **Business Requirements**

<!-- -->

1.  **Functional Requirements Specification**

Surrender value of a policy, is the amount that is payable to the
Insurant, when he/she foregoes the contingent benefit of his policy and
surrenders it for an immediate payment. Surrender value depends on the
surrender factor and type and term of policy and as per the prescribed
ruling mentioned in POLI

# 

## Brief Description {#brief-description .ABC}

> This document details the functionality of Commutation and its impact
> on other functionalities.

# System Flows

## Standard Flow

> Indexing of Commutation Request → Data Entry → Quality Checker →
> Approver → Letter Generation (Decrease in sum assured accepted letter)
> → Request status changed to approved

## Sub-Standard Flows

> [Rejected by Approver]{.underline}
>
> Indexing of Commutation Request → Data Entry → Quality Checker →
> Approver rejected the request → Letter Generation (Decrease in sum
> assured rejection letter) → Request status changed to Rejected.
>
> [Redirected by Approver]{.underline}
>
> Indexing of Commutation Request → Data Entry → Quality Checker →
> Approver redirected the request → Data Entry → Quality Checker →
> Approver → Letter Generation (Decrease in sum assured accepted letter)
> → Request status changed to approved
>
> [Missing document by Quality Checker]{.underline}
>
> Indexing of Commutation Request → Data Entry → Quality Checker raises
> missing document → Missing document received → Quality checker →
> Approver → Letter Generation (Decrease in sum assured accepted letter)
> → Request status changed to approved
>
> [Missing document by Approver]{.underline}
>
> Indexing of Commutation Request → Data Entry → Quality Checker →
> Approver raises missing document → Missing document received →
> Approver → Letter Generation (Decrease in sum assured accepted letter)
> → Request status changed to approved
>
> [Withdrawal of Request]{.underline}
>
> Indexing of Commutation Request → Data Entry → Quality Checker →
> Approver → withdrawal of request → request status changed to
> withdrawn.
>
> Indexing of Commutation Request → Data Entry → Quality Checker →
> withdrawal of request → request status changed to withdrawn.

# Commutation Flow

![](media/image1.emf)

# Business Rules and Logics

# Calcuation

[Step 1:]{.underline}

New Premium **needs** to be selected

PUV as on current date= Total Premium Paid as on current date / Total
Premium Payable\*Old SA

[Step 2 :]{.underline}

Difference Amount = New Premium x ( 1000 (for RPLI) or 5000 (for PLI) or
10000 (for Yugal Suraksha)) / Rate of Premium from Premium Table as on
the current date)

[Step 3 :]{.underline}

New SA = PUV + Difference

Difference amount calculated will be rounded off to nearest thousand and
added to paid up value.

[Step 4 :]{.underline}

New difference amount = New sum assured -- Paid up value

[Step 5 :]{.underline}

New Premium after rounding off = (Premium Rate from premium rate table
as on current date \* New Difference amount) / 5000 (PLI) or 1000 (RPLI)
or 10000 (Yugal Suraksha)

**[Decrease in Sum Assured - Rules]{.underline}**

- Paid up value has to be calculated

- Attained age as on commutation effective date has to be calculated

- Difference Amount = New Sum Assured-PUV

- New Sum assured cannot be less than the Paid up value.

- Premium will be calculated on Difference Amount

- Premium rate will be based on attained age as on current Date and
  Remaining Term of the Policy/premium ceasing age.

- Rebate will be calculated based on Difference Amount:

- Rebate of Re. 1 per Rs.20000 and for Yugal Suraksha Re. 1 per Rs.40000
  of Sum Assured will be calculated on Difference Amount i.e. New Sum
  Assured -- Paid up Value.

- 

- There is no limit on the minimum and maximum counts of Commutations
  during a policy term.

Example:

  -------------------------------------------
  Policy                              xxxxxxx
  ----------------------------- -------------
  Issue date                           Mar-20

  Date of Birth                    07-07-1983

  Attained age at the time of              43
  conversion                    

  Commutation date                  17-Oct-25

  Sum Assured                          500000

  Premium frequency                   Monthly

  Premiums paid before                     68
  conversion                    

  Total Premium Payable Months            216

  Paid up value                   157407.4074

  Opted SA                             300000

  Remaining Sum Assured:               142593

   Rounded to multiples of             143000
  10000                         

  Premium Rate for remaining               36
  sum assured                   

  Premium amount before                  1030
  rebate(Monthly)               

  Rebate                                    7

  Premium amount After               **1023**
  rebate(Monthly)               

                                
  -------------------------------------------

> **Eligibility Criteria-**

1.  Policy Status should be paid till date.

2.  Any PLI/RPLI product except Sumangal, Gram Sumangal and Gram Priya.

<!-- -->

A.  **[Quote]{.underline}**

Login by indexer-\>Quote icon-\>Enter quote type as commutation-\>Enter
policy number-\>select date-\>Next button-\> select any option either
change in premium amount or change in Sum assured" -\> Enter revised
premium amount or sum assured in respective field-\> Click on get quote
button.

Login by indexer-\>Quote icon

![A screenshot of a computer AI-generated content may be
incorrect.](media/image2.png){width="6.5in"
height="2.3047626859142607in"}

Enter quote type as commutation-\>Enter policy number-\>select
date-\>Next button

![A screenshot of a computer AI-generated content may be
incorrect.](media/image3.png){width="6.5in"
height="1.0523545494313211in"}

+------------------+---------------------------------------------------+
| **Error          | **Description**                                   |
| Message**        |                                                   |
+==================+===================================================+
| Invalid policy   | When no policy number is entered on the Quote and |
| number           | 'Next' button is clicked.                         |
+------------------+---------------------------------------------------+
| Policy Details   | When incorrect policy number is entered on the    |
| not found        | Quote and 'Next' button is clicked.               |
+------------------+---------------------------------------------------+
| Commutation is   | User tries to generate quote when already         |
| allowed only     | commutation request is effective on the policy or |
| once during the  | is approver or in pending status.                 |
| term of the      |                                                   |
| policy.          | Error will be thrown on Quote screen on the click |
|                  | of the 'Next' button.                             |
+------------------+---------------------------------------------------+

Select any option either 'change in premium amount' or 'change in Sum
assured'. If User selected "Change in Premium Amount", then user should
enter the revised premium amount.

![A screenshot of a computer AI-generated content may be
incorrect.](media/image4.png){width="6.5in"
height="2.2523447069116362in"}

On entrant the revised premium amount, user will get 2 options for the
selection of premium amount based on the sum assured.

![A screenshot of a computer AI-generated content may be
incorrect.](media/image5.png){width="6.5in"
height="2.687159886264217in"}

If nothing is selected and user clicks submit button

![A screenshot of a computer AI-generated content may be
incorrect.](media/image6.png){width="6.002389545056868in"
height="1.69581583552056in"}

On the click of get quote, user will get an option to open or save the
quote

![A screenshot of a computer AI-generated content may be
incorrect.](media/image7.png){width="6.5in"
height="2.922189413823272in"}

If user chose 'change in sum assured'

![A screenshot of a computer AI-generated content may be
incorrect.](media/image8.png){width="6.087123797025372in"
height="2.278380358705162in"}

If sum assured selected is less than the allowed limit

![A computer screen shot of a computer AI-generated content may be
incorrect.](media/image9.png){width="5.844998906386702in"
height="1.4785520559930008in"}

If paid up value id greater than the entered sum assured value, then
below error will be thrown

![A computer screen shot of a computer AI-generated content may be
incorrect.](media/image10.png){width="5.436306867891513in"
height="1.4003390201224848in"}

If sum assured entered is not in multiples of ~~1000 for PLI policies~~
10000 for PLI policies and 5000 for RPLI policies then below error will
be thrown.

![A computer screen shot of a computer AI-generated content may be
incorrect.](media/image11.png){width="6.060856299212598in"
height="2.7052099737532807in"}

After selecting the sum assured, user should click on the get quote.
User will get an option to open or save the quote PDF.

![A computer screen with a message box AI-generated content may be
incorrect.](media/image12.png){width="5.928488626421697in"
height="2.0695844269466317in"}

> ![A screenshot of a computer AI-generated content may be
> incorrect.](media/image13.png){width="3.49621719160105in"
> height="3.0783169291338583in"}

+------------------+---------------------------------------------------+
| **Error          | **Description**                                   |
| Message**        |                                                   |
+==================+===================================================+
| Product not      | If user tries to index commutation request on     |
| eligible         | products - Sumangal (AEA), Gram Sumangal (RAEA)   |
|                  | and Gram Priya (GY).                              |
|                  |                                                   |
|                  | Error will be thrown on Quote screen on the click |
|                  | of the Get Quote button.                          |
+------------------+---------------------------------------------------+
| Invalid policy   | If policy status is other than AP at the time of  |
| status           | indexing or approval.                             |
|                  |                                                   |
|                  | Error will be thrown on Quote screen on the click |
|                  | of the Get Quote button.                          |
+------------------+---------------------------------------------------+
| Please select an | When 'change in premium amount' is selected by    |
| option           | the user but form the given option nothing is     |
|                  | selected and submit button is clicked.            |
|                  |                                                   |
|                  | Error will be thrown on Quote screen on the click |
|                  | of the Get Quote button.                          |
+------------------+---------------------------------------------------+
| Commutation -    | When 'change in sum assured' is selected and the  |
| Applied sum      | amount entered is less than the allowed limit.    |
| assured is less  |                                                   |
| than allowed     | This error will be thrown on the Quote screen on  |
| limit for this   | entering the sum assured amount.                  |
| product          |                                                   |
+------------------+---------------------------------------------------+
| Commutation -    | When user enter the sum assured amount less than  |
| Current Paid up  | the current paid up value of the policy.          |
| value is greater |                                                   |
| than reduced sum | This error will be thrown on the Quote screen on  |
| assured          | entering the sum assured amount.                  |
+------------------+---------------------------------------------------+
| Commutation -    | When 'change in sum assured' is selected and the  |
| Sum assured      | amount entered is not in multiple of 1000 for PLI |
| should be in     | product and 5000 for RPLI product.                |
| ~~multiple of    |                                                   |
| 1000 for PLI~~   | This error will be thrown on the Quote screen on  |
| multiple of      | entering the sum assured amount.                  |
| 10000 for PLI    |                                                   |
| product and 5000 |                                                   |
| for RPLI         |                                                   |
| product.         |                                                   |
+------------------+---------------------------------------------------+

B.  **Service request indexing**

Login by indexer→ Service request indexing icon→ Select request type as
Commutation→ Enter policy number→ select date→ Click on next button

![A screenshot of a computer AI-generated content may be
incorrect.](media/image14.png){width="6.305932852143482in"
height="1.339225721784777in"}

Click on submit button

![A screenshot of a computer AI-generated content may be
incorrect.](media/image15.png){width="6.261873359580052in"
height="1.7044400699912512in"}

On the click of the submit button, request ID will be generated. Request
nomenclature "PSCOMXXXXXXXXXX"

![A screenshot of a computer AI-generated content may be
incorrect.](media/image16.png){width="6.5in"
height="1.2434951881014873in"}

+------------------+---------------------------------------------------+
| **Error          | **Description**                                   |
| Message**        |                                                   |
+==================+===================================================+
| Invalid policy   | When no policy number is entered on the Indexing  |
| number           | screen and 'Next' button is clicked.              |
+------------------+---------------------------------------------------+
| Policy Details   | When incorrect policy number is entered on the    |
| not found        | Indexing screen and 'Next' button is clicked.     |
+------------------+---------------------------------------------------+
|                  |                                                   |
+------------------+---------------------------------------------------+
| Financial        | If any financial request is pending, commutation  |
| request is       | cannot be indexed.                                |
| pending          |                                                   |
|                  | Loan, Conversion, Death Claim, Maturity Claim,    |
|                  | Billing frequency change, Reduced paid up, Policy |
|                  | cancellation, Free look cancellation, DOB change, |
|                  | Surrender and Commutation.                        |
|                  |                                                   |
|                  | Error will be thrown on Indexing screen on the    |
|                  | click of the submit button.                       |
+------------------+---------------------------------------------------+

C.  **Data entry screen-**

> Login by data entry operator→ Click on inbox → Select the reserved
> request id

![A screenshot of a computer AI-generated content may be
incorrect.](media/image17.png){width="6.5in"
height="1.5208333333333333in"}

Select any option either change in premium amount or change in sum
assured → Click on continue button

- Change in Sum Assured

![A screenshot of a computer AI-generated content may be
incorrect.](media/image18.png){width="6.5in"
height="1.9479166666666667in"}

- Change in Premium

![A screenshot of a computer AI-generated content may be
incorrect.](media/image19.png){width="6.5in"
height="1.9895833333333333in"}

On the click on the Submit button, below pop up will appear. On clicking
'Yes', request will be submitted to the Quality Checker.

![A screenshot of a computer AI-generated content may be
incorrect.](media/image20.png){width="6.538664698162729in"
height="1.7236975065616797in"}

+------------------+---------------------------------------------------+
| **Error          | **Description**                                   |
| Message**        |                                                   |
+==================+===================================================+
| Please select an | When 'change in premium amount' is selected by    |
| option           | the user but form the given option nothing is     |
|                  | selected and submit button is clicked.            |
|                  |                                                   |
|                  | This error is applicable for Data Entry and       |
|                  | Quality checker screen.                           |
+------------------+---------------------------------------------------+
| Commutation -    | When 'change in sum assured' is selected and the  |
| Applied sum      | amount entered is less than the allowed limit.    |
| assured is less  |                                                   |
| than allowed     | This error will be thrown on the Data Entry and   |
| limit for this   | Quality checker screen on entering the sum        |
| product          | assured amount.                                   |
+------------------+---------------------------------------------------+
| Commutation -    | When user enter the sum assured amount less than  |
| Current Paid up  | the current paid up value of the policy.          |
| value is greater |                                                   |
| than reduced sum | This error will be thrown on the g, Data Entry    |
| assured          | and Quality checker screen on entering the sum    |
|                  | assured amount.                                   |
+------------------+---------------------------------------------------+
| Commutation -    | When 'change in sum assured' is selected and the  |
| Sum assured      | amount entered is not in multiple of 10000 for    |
| should be in     | PLI product and 5000 for RPLI product.            |
| multiple of      |                                                   |
| ~~1000~~ 10000   | This error will be thrown on Data Entry and       |
| for PLI product  | Quality checker screen on entering the sum        |
| and 5000 for     | assured amount.                                   |
| RPLI product.    |                                                   |
+------------------+---------------------------------------------------+

D.  **Quality checker screen:**

Login by Quality checker → Click on inbox → Select the reserved request
id

![A screenshot of a computer AI-generated content may be
incorrect.](media/image21.png){width="6.875in"
height="1.7395833333333333in"}

![A screenshot of a computer AI-generated content may be
incorrect.](media/image22.png){width="6.885416666666667in"
height="3.8645833333333335in"}

On the click of the Submit button, pop up will appear. On clicking
'Yes', request will be submitted to Approver.

![A screenshot of a computer AI-generated content may be
incorrect.](media/image23.png){width="6.551484033245845in"
height="1.6466054243219597in"}

E.  **Approval Screen:**

Login by Approver→ Click on inbox→ Select the reserved request id

![A screenshot of a computer AI-generated content may be
incorrect.](media/image24.png){width="6.5in" height="2.0in"}

![A screenshot of a computer AI-generated content may be
incorrect.](media/image25.png){width="6.5in"
height="2.746578083989501in"}

On the click of submit button, pop up will appear. On clicking 'Yes'
button, request will be approved.

![A screenshot of a computer AI-generated content may be
incorrect.](media/image26.png){width="5.992012248468941in"
height="1.5540288713910761in"}
