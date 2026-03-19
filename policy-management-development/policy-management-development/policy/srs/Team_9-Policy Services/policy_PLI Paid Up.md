**Postal Life Insurance**

**Department of Post**

**Software Requirements Specification (SRS)**

**1. Policy -- paid up/auto paid up**

1.  **Executive Summary**

**This document outlines the business requirements for conditions under
which a policy is treated as paid up/auto paid up in the Insurance
Management System (IMS). It intends to notify under which circumstances,
the policy is declared as paid up and auto paid up .**

2.  **Business Objectives**

**The IMS project addresses the requirement of declaring a policy as
paid up /auto paid up. The insurant should be able to get the proceeds
upon declaring the policy as paid up/auto paid up**

3.  **Project Scope**

**The scope is limited to declare a policy as paid up/auto paid up under
the conditions laid down in POLI Rules 2011, to facilitate the insurant
to receive the proceeds accordingly**

+-----------+-------------------------------------------------+
| **RFP Req | **Requirement Description**                     |
| No.**     |                                                 |
|           | **(Pl refer FR 22,223,343,925& 1246)**          |
+:=========:+:================================================+
| **SR-PAID | **Must have facility to flag the first unpaid   |
| UP-1**    | premium , as well as the inception date and     |
|           | date of last instalment, as the basic           |
|           | parameters to decide whether the policy has to  |
|           | be decided as paid up.**                        |
+-----------+-------------------------------------------------+
| **SR-Paid | **System should have the facility to identify   |
| up-2**    | different kind of policies and its business     |
|           | rules to declare a policy as paid up**          |
+-----------+-------------------------------------------------+
| **SR-Paid | **Policies which are in-force for more than 3   |
| up-3**    | years only are eligible for paid up or auto     |
|           | paid up.**                                      |
+-----------+-------------------------------------------------+
| **SR-Paid | **An insured person is not to be considered as  |
| up-4**    | in arrears of premium for any month so long as  |
|           | he has not been able to draw his pay, pension   |
|           | or subsistence allowance during suspension, or, |
|           | if the insured person is on leave in India, any |
|           | leave allowance though due, for the month next  |
|           | before it is due because of circumstances       |
|           | beyond his control. Therefore system should be  |
|           | able to distinguish and calculate paid up value |
|           | accordingly.**                                  |
+-----------+-------------------------------------------------+
| **SR-Paid | **System should be able to distinguish between  |
| up-5**    | cash policy and pay recovery policies to        |
|           | determine a policy to be paid up considering    |
|           | the above parameters.**                         |
+-----------+-------------------------------------------------+
| **SR-Paid | **System should take cognizance of pay recovery |
| up-6**    | policies only in respect of the parameters laid |
|           | down under SR-PAID UP-6, as above.**            |
+-----------+-------------------------------------------------+
| **SR-Paid | **In case of pay recovery policies, no policy   |
| up-7**    | will be treated as paid until 11 months from    |
|           | the date month of the first un paid premium,    |
|           | and auto paid up or paid up value to be         |
|           | determined by the system accordingly.**         |
+-----------+-------------------------------------------------+
| **SR-     | **If a policy is intended to be treated as      |
| Paid      | forced surrender, then the system should be     |
| up-8**    | able to flag the last instalment date and treat |
|           | the policy as auto paid up.**                   |
+-----------+-------------------------------------------------+
| **SR-Paid | **In case of revival by instalment process, if  |
| up 9**    | the revival instalments are not paid on time,   |
|           | the last instalment paid before the revival     |
|           | application, is to be flagged by the system,    |
|           | and auto paid up value to be calculated         |
|           | accordingly.**                                  |
+-----------+-------------------------------------------------+
| **SR-Paid | **In case of policies incepted after            |
| up 10**   | \_\_\_\_\_\_\_\_\_\_\_\_\_\*\*, if the non      |
|           | credits crosses 60 months , but after 36 months |
|           | from the month of inception, then such policies |
|           | should be declared as auto paid up by the       |
|           | system.**                                       |
+-----------+-------------------------------------------------+
| **SR-Paid | **SMS/Email /Whatsapp messages have to be sent  |
| up-8**    | whenever a policy is declared as paid up/auto   |
|           | paid up.**                                      |
+-----------+-------------------------------------------------+

**\*\* date to be ascertained from the files concerned.**

# 

# 

# In-Scope

Following PLI products are in scope:

  ------------------------------------------------------------
   **PLI/RPLI**    **Product           **Product Type**
                     Name**     
  -------------- -------------- ------------------------------
       PLI          Suraksha         Whole Life Assurance

       PLI          Santosh          Endowment Assurance

       PLI          Suvidha         Convertible Whole Life
                                          Assurance

       PLI          Sumangal        Anticipated Endowment
                                          Assurance

       PLI        Child Policy           Child Policy

       PLI       Yugal Suraksha      Joint Life Assurance

       RPLI      Gram Suraksha       Whole Life Assurance

       RPLI       Gram Santosh       Endowment Assurance

       RPLI       Gram Suvidha      Convertible Whole Life
                                          Assurance

       RPLI      Gram Sumangal      Anticipated Endowment
                                          Assurance

       RPLI       Child policy           Child policy

       RPLI        Gram Priya       Anticipated Endowment
                                          Assurance
  ------------------------------------------------------------

Business Rules

1\. A paid-up policy is one where the policy holder stops paying regular
premiums, but continues to enjoy partial insurance coverage

2\. When a policy is converted into a paid-up policy, there is no
further obligation to pay regular premium and no premium is to be
permitted to be paid.

3\. The policyholder or beneficiaries do not receive the policy's
original coverage, and it is reduced to the extent of the paid up value
plus applicable bonus, as the case may be.

4\. The formula for paid up or auto paid up value is

Sum Assured ( No. of instalments paid (in months)/Total number of
instalments (in months)
