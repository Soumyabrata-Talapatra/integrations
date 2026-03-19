> INTERNAL APPROVAL FORM

**Project Name:** Billing & Collection of various PLI/RPLI receipts

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

[DUPLICATE POLICY BOND
[5](#duplicate-policy-bond)](#duplicate-policy-bond)

[BILLING METHOD CHANGE
[5](#billing-method-change)](#billing-method-change)

[NOMINATION CHANGE [5](#nomination-change)](#nomination-change)

[PREMIUM RECEIPTBOOK [5](#premium-receiptbook)](#premium-receiptbook)

[ADDRESS CHANGE [6](#address-change)](#address-change)

[NAME CHANGE [6](#name-change)](#name-change)

[FREELOOK [6](#freelook)](#freelook)

[COMMUTATION [6](#commutation)](#commutation)

[MATURITY [7](#maturity)](#maturity)

[LOAN [7](#loan)](#loan)

[SURRENDER [7](#surrender)](#surrender)

[DEATH CLAIM [7](#death-claim)](#death-claim)

[SURVIVAL [8](#survival)](#survival)

[REVIVAL [8](#revival)](#revival)

[NEW PROPOSAL [8](#new-proposal)](#new-proposal)

[REDUCED PAIDUP [8](#reduced-paidup)](#reduced-paidup)

[**5. Attachments** [9](#attachments)](#attachments)

## **1. Executive Summary**

The purpose of this document is to define the functional and
non-functional requirements for generating standardized receipts for
various PLI/RPLI transactions (Loan, Maturity, Surrender, Revival,
Premium Payment, etc.) within the IMS. The goal is to ensure
consistency, accuracy, and compliance with Department of Posts
standards.

## **2. Project Scope**

This Module will:

- Display mandatory details such as Policyholder Name, Policy Number,
  Transaction Type, Amount, Date, Office Code, and User ID.

- Provide uniform templates across all transaction types.

- Include security features like QR code/barcode and digital signature.

## **3. Business Requirements**

+-------------+---------------------------------------------------------+
| **ID**      | **Requirements**                                        |
+=============+=========================================================+
| FS_RCPT_001 | System shall generate receipts for all transaction      |
|             | types (Loan, Maturity, Surrender, Revival, Premium      |
|             | Payment).                                               |
+-------------+---------------------------------------------------------+
| FS_RCPT_002 | Receipt shall include:                                  |
|             |                                                         |
|             | - Policy Number                                         |
|             |                                                         |
|             | - Insurant/Policyholder Name                            |
|             |                                                         |
|             | - Transaction Type                                      |
|             |                                                         |
|             | - Amount Paid/Received                                  |
|             |                                                         |
|             | - Date & Time                                           |
|             |                                                         |
|             | - Office Code                                           |
|             |                                                         |
|             | - User ID                                               |
|             |                                                         |
|             | - Policy Type (PLI/RPLI)                                |
|             |                                                         |
|             | - QR Code/Barcode for verification                      |
|             |                                                         |
|             | - Digital Signature for authenticity                    |
+-------------+---------------------------------------------------------+
| FS_RCPT_003 | System shall validate presence of Insurant Name before  |
|             | receipt generation.                                     |
+-------------+---------------------------------------------------------+
| FS_RCPT_004 | All receipts shall follow a uniform template with       |
|             | standardized header and footer.                         |
+-------------+---------------------------------------------------------+
| FS_RCPT_005 | Header shall display "Department of Posts -- Postal     |
|             | Life Insurance / RPLI".                                 |
+-------------+---------------------------------------------------------+
| FS_RCPT_006 | System shall fetch Insurant Name and Policy details     |
|             | from the central policy database.                       |
+-------------+---------------------------------------------------------+
| FS_RCPT_007 | QR code/barcode shall link to receipt verification      |
|             | service.                                                |
+-------------+---------------------------------------------------------+
| FS_RCPT_008 | Digital signature shall confirm authenticity.           |
+-------------+---------------------------------------------------------+

## **4. Functional Requirements Specification**

### DUPLICATE POLICY BOND

POLICY NUMBER-AP-512948-UC

![C:\\Users\\dop\\Desktop\\Capture8.PNG](media/image1.png){width="6.5in"
height="1.5063396762904637in"}

### BILLING METHOD CHANGE

POLICY NUMBER-AP-589243-CC

![C:\\Users\\dop\\Desktop\\Capture9.PNG](media/image2.png){width="6.5in"
height="1.5386745406824147in"}

### NOMINATION CHANGE

POLICY NUMBER-AP-584307-CS

![C:\\Users\\dop\\Desktop\\Capture10.PNG](media/image3.png){width="6.5002646544181975in"
height="1.3756157042869641in"}

### PREMIUM RECEIPTBOOK

POLICY NUMBER-0000012040494

![C:\\Users\\dop\\Desktop\\Capture11.PNG](media/image4.png){width="5.546772747156606in"
height="1.2240594925634296in"}

### ADDRESS CHANGE

POLICY NUMBER-AP-511774-UC

![C:\\Users\\dop\\Desktop\\Capture13.PNG](media/image5.png){width="6.5in"
height="1.5792825896762905in"}

### NAME CHANGE

POLICY NUMBER-0000000458382

![C:\\Users\\dop\\Desktop\\Capture15.PNG](media/image6.png){width="6.5in"
height="1.508173665791776in"}

### FREELOOK

POLICY NUMBER-0000012103934

![C:\\Users\\dop\\Desktop\\Capture17.PNG](media/image7.png){width="6.500620078740157in"
height="1.2227143482064742in"}

### COMMUTATION

POLICY NUMBER-0000001916590

![C:\\Users\\dop\\Desktop\\Capture18.PNG](media/image8.png){width="5.930179352580927in"
height="1.2880686789151357in"}

### MATURITY 

POLICY NUMBER- AP-439698-CS

> ![C:\\Users\\dop\\Desktop\\Capture.PNG](media/image9.png){width="6.5in"
> height="1.5068110236220473in"}

### LOAN

POLICY NUMBER-0000001940091

![C:\\Users\\dop\\Desktop\\Capture1.PNG](media/image10.png){width="6.5in"
height="1.4648162729658794in"}

### SURRENDER

POLICY NUMBER-AP-403856-UC

![C:\\Users\\dop\\Desktop\\Capture3.PNG](media/image11.png){width="6.5in"
height="1.5057808398950132in"}

### DEATH CLAIM

POLICY NUMBER-AP-419708-UC

![C:\\Users\\dop\\Desktop\\Capture4.PNG](media/image12.png){width="5.858844050743657in"
height="1.4197233158355205in"}

### SURVIVAL

POLICY NUMBER-0000003126432

![C:\\Users\\dop\\Desktop\\Capture5.PNG](media/image13.png){width="6.5in"
height="1.5051006124234472in"}

### REVIVAL

POLICY NUMBER-AP-627201-CS

![C:\\Users\\dop\\Desktop\\Capture7.PNG](media/image14.png){width="6.5in"
height="1.4939818460192476in"}

### NEW PROPOSAL

POLICY NUMBER- 0000012786412

![C:\\Users\\dop\\Desktop\\Capture12.PNG](media/image15.png){width="5.432922134733158in"
height="1.6737193788276465in"}

### REDUCED PAIDUP

POLICY NUMBER-R-AP-VP-EA-156274

![C:\\Users\\dop\\Desktop\\Capture20.PNG](media/image16.png){width="6.064825021872266in"
height="1.4066590113735784in"}

## **5. Attachments**

The following documents can be referred.
