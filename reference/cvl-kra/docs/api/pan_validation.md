# `pan_validation()`

{{>toc}}

**Description** : This WS allows OKRA to make KYC inquiries to CVL via an API. It can accept requests in both XML ~~and JSON~~ formats. Maximum of 25 Pan No’s allowed.

## `XML`
### Request

#### `Single PAN`
```xml
<?xml version="1.0" encoding="utf-8"?>
<soap12:Envelope xmlns:soap12="http://www.w3.org/2003/05/soap-envelope">
  <soap12:Body>
    <PANValidation xmlns="https://test.cvlkra.com/">
      <inputXML>
        <![CDATA[
            <APP_REQ_ROOT>
                <APP_PAN_INQ>
                    <APP_PAN_NO>GLHJK1232D</APP_PAN_NO>
                    <APP_PAN_DOB/>
                    <APP_IOP_FLG>RS</APP_IOP_FLG>
                    <APP_POS_CODE>H</APP_POS_CODE>
                </APP_PAN_INQ>
                <APP_SUMM_REC>
                    <APP_OTHKRA_CODE>CAMS</APP_OTHKRA_CODE>
                    <APP_OTHKRA_BATCH>123456</APP_OTHKRA_BATCH>
                    <APP_REQ_DATE>28-04-2021</APP_REQ_DATE>
                    <APP_TOTAL_REC>1</APP_TOTAL_REC>
                </APP_SUMM_REC>
            </APP_REQ_ROOT>
        ]]>
      </inputXML>
      <userName>cvl.admin</userName>
      <PosCode>cvlkra20</PosCode>
      <password>%2f%2fw0mhHSDuXD%2fWKAZRTbOw%3d%3d</password>
      <PassKey>12345</PassKey>
    </PANValidation>
  </soap12:Body>
</soap12:Envelope>
```

#### `Bulk PAN`
```xml
<?xml version="1.0" encoding="utf-8"?>
<soap12:Envelope xmlns:soap12="http://www.w3.org/2003/05/soap-envelope">
  <soap12:Body>
    <PANValidation xmlns="https://test.cvlkra.com/">
      <inputXML>
        <![CDATA[
            <APP_REQ_ROOT>
                <APP_PAN_INQ>
                    <APP_PAN_NO>GLHJK1232D</APP_PAN_NO> <!- pan 1 -->
                    <APP_PAN_DOB/>
                    <APP_IOP_FLG>RS</APP_IOP_FLG>
                    <APP_POS_CODE>H</APP_POS_CODE>
                </APP_PAN_INQ>
                 <APP_PAN_INQ>
                    <APP_PAN_NO>IJSPJ9839A</APP_PAN_NO> <!- pan 2 -->
                    <APP_PAN_DOB/>
                    <APP_IOP_FLG>RS</APP_IOP_FLG>
                    <APP_POS_CODE>H</APP_POS_CODE>
                </APP_PAN_INQ>
                <APP_SUMM_REC>
                    <APP_OTHKRA_CODE>CAMS</APP_OTHKRA_CODE>
                    <APP_OTHKRA_BATCH>123456</APP_OTHKRA_BATCH>
                    <APP_REQ_DATE>28-04-2021</APP_REQ_DATE>
                    <APP_TOTAL_REC>1</APP_TOTAL_REC>
                </APP_SUMM_REC>
            </APP_REQ_ROOT>
        ]]>
      </inputXML>
      <userName>cvl.admin</userName>
      <PosCode>cvlkra20</PosCode>
      <password>%2f%2fw0mhHSDuXD%2fWKAZRTbOw%3d%3d</password>
      <PassKey>12345</PassKey>
    </PANValidation>
  </soap12:Body>
</soap12:Envelope>
```

#### Properties list
* `<soap:Envelope>`: The root element of the SOAP message, defining the namespaces used in the document.
* `<soap:Body>`: Contains the main content of the SOAP request.
 * `<PANValidation xmlns="https://www.cvlkra.com/">`: The main SOAP method for PAN validation.
 * `<inputXML>`: CDATA section containing the XML request for PAN validation.
 * `<APP_REQ_ROOT>`: Request root start tag and can contain multiple `<APP_PAN_INQ>`.
 * `<APP_PAN_INQ>`: Contains details of the PAN inquiry.
 * `<APP_PAN_NO>`: string, mandatory, The PAN number being validated (e.g., "ACAFS9645R" or "ACAFS9645T").
 * `<APP_IOP_FLG>`: string, mandatory, Inquiry operation type flag. (e.g.`IE`: Inquiry Existing; `IN`: Inquiry New)
 * `<APP_POS_CODE>`: string, mandatory, Point of Service (POS) code for the KRA office (e.g., "ndmL1" or "NDML1").
 * `<APP_SUMM_REC>`: Contains summary details of the PAN inquiry batch.
 * `<APP_OTHKRA_BATCH>`: string, mandatory, The batch ID for the KRA inquiry (e.g., "001").
 * `<APP_OTHKRA_CODE>`: string, mandatory, The KRA code handling the inquiry (e.g., "NDML1").
 * `<APP_REQ_DATE>`: string, mandatory, The request date in **DD-MM-YYYY HH:MM:SS** format (e.g., "18-12-2024 00:00:00").
 * `<APP_TOTAL_REC>`: integer, mandatory, The total number of records in the batch (e.g., 2).
* `<userName>`: string, mandatory, The username for authenticating the request (e.g., "ind-bulluser").
* `<PosCode>`: string, mandatory, Point of Service (POS) code identifying the KRA office (e.g., "CVLKRA1").
* `<password>`: string, mandatory, The encryption key used for secure communication (e.g., "1234567891011121").
* `<PassKey>`: string, mandatory, The KRA code handling the inquiry.

### Processing

1. Perform Authorization Header Validation
  * Check Content Type is XML
  * Check Content Length matches with evaluated length
  * Check SOAPAction is equals to PANValidation
1. Bind the incoming payload with Go Struct model
1. Iterate for each PAN and set PANListCount
1. Decrypt the Encrypted password and Passkey and get the actual password
1. Perform validation with actual password and passkey present in Redis
1. Perform validation process using username and pos code from organization table
1. Perform Summary Validation
  * Check APP_IOP_FLG Flag == "IE"
  * Check the PANListCount is not more than 25
  * Check APP_OTHKRA_CODE code
  * Check the APP_REQ_DATE is equals to current date
  * Check the APP_TOTAL_REC is equals to PANListCount
1. Iterate for each PANList
  * Perform PAN validation
  * Check PAN exist in `kycrecords`
  * If the result size is 1
      * Check for operation_key if REGISTRATION
      * Set all the status parameter as per response
  * Else If the result size is greater than 1
      * Get the latest modified `kycrecords`
      * Set all the status parameter as per response
  * Else
      * Set Status as `Not Available`
1. Build final response by adding summary details
1. Send the response

### Function to get pan by status

```
print("Method Starts here...")

Query: SELECT * FROM kycrecords WHERE operation_key IN ('REGISTRATION', 'MODIFICATION') ORDER BY processed_at DESC

[] panFromDb = getPanFromDB(panNumber) {}

sizeOfPanFromDb = len(panFromDb)

if sizeOfPanFromDb == 1 {

  return getDataByReg(panFromDb[0]);

} else if (sizeOfPanFromDb > 1) {

  return getDataByRegAndMod(panFromDb)

} else {

  return "00","00", emptyData {}

}

```


```
getDataByReg(panFromDb[0]) {

  reg_status = panFromDb.status

  kyc_status = reg_status

  mod_status = "00"

  return kyc_status, mod_status, panFromDb[0]

}

```


```
getDataByRegAndMod(panFromDb) {

  curr_mod_status = panFromDb[0].status

  mod_status = curr_mod_status
  kyc_status, data = getKycStatus(panFromDb)

  return kyc_status, mod_status, data

}

getKycStatus(panFromDb) {

  curr_mod_status = panFromDb[0].status
  pre_mod_status = panFromDb[1].status

  curr_mod_status_priority: =statusVspriority[curr_mod_status]
  pre_mod_status_priority: =statusVspriority[pre_mod_status]

  if curr_mod_status_priority < pre_mod_status_priority {

    return pre_mod_status, panFromDb[1]
  } else {

    return curr_mod_status, panFromDb[0]
  }


  return curr_mod_status, panFromDb[0]
}

statusVspriority = map[string] int32 {
  "00": 0
  "01": 1,
  "02": 2,
  "03": 3,
  "04": 4,
  "05": 5,
  "06": 6,
  "07": 7,
  "08": 8,
  "09": 9,
}

//internal status codes
Not Avaliable: 00
Draft : 01,
MI Completed : 02,
Pending with KRA : 03,
CVL KRA Maker Verified : 04,
KYC Rejected : 05,
On HOLD : 06,
KYC Registered : 07,
KYC Validated : 08,
KYC Deactivated : 09,

```

### Response
#### Success

##### `Single PAN`
```xml
<?xml version="1.0" encoding="UTF-8"?>
<soap12:Envelope xmlns:soap12="http://www.w3.org/2003/05/soap-envelope" xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
   <soap12:Body>
      <PANValidationResponse xmlns="https://test.cvlkra.com/">
         <PANValidationResult>
            <APP_RES_ROOT>
               <APP_PAN_INQ>
                  <APP_PAN_NO>AAAPA1111A</APP_PAN_NO>
                  <APP_NAME>MEHR CHADDAH</APP_NAME>
                  <APP_IOP_FLG>I</APP_IOP_FLG>
                  <APP_STATUS>02</APP_STATUS>
                  <APP_PAN_DOB>25-11-1990 00:00:00</APP_PAN_DOB>
                  <APP_ENTRYDT>01-09-2017 17:52:21</APP_ENTRYDT>
                  <APP_STATUSDT>17-02-2020 14:52:14</APP_STATUSDT>
                  <APP_MODDT>04-03-2021 14:35:21</APP_MODDT>
                  <APP_POS_CODE>CDSL</APP_POS_CODE>
                  <APP_STATUS_DELTA />
                  <APP_UPDT_STATUS>01</APP_UPDT_STATUS>
                  <APP_HOLD_DEACTIVE_RMKS />
                  <APP_UPDT_RMKS />
                  <APP_KYC_MODE>0</APP_KYC_MODE>
                  <APP_IPV_FLAG>Y</APP_IPV_FLAG>
                  <APP_UBO_FLAG />
               </APP_PAN_INQ>
               <APP_PAN_SUMM>
                  <APP_OTHKRA_CODE>CDSL</APP_OTHKRA_CODE>
                  <APP_OTHKRA_BATCH>20210529</APP_OTHKRA_BATCH>
                  <APP_REQ_DATE>26/05/2021</APP_REQ_DATE>
                  <APP_RESPONSE_DATE>31-05-2021 09:52:35</APP_RESPONSE_DATE>
                  <APP_TOTAL_REC>1</APP_TOTAL_REC>
               </APP_PAN_SUMM>
            </APP_RES_ROOT>
         </PANValidationResult>
      </PANValidationResponse>
   </soap12:Body>
</soap12:Envelope>
```

##### `Bulk PAN`
```xml
<?xml version="1.0" encoding="UTF-8"?>
<soap12:Envelope xmlns:soap12="http://www.w3.org/2003/05/soap-envelope" xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
   <soap12:Body>
      <PANValidationResponse xmlns="https://test.cvlkra.com/">
         <PANValidationResult>
            <APP_RES_ROOT>
               <APP_PAN_INQ>
                  <APP_PAN_NO>AAAPA1111A</APP_PAN_NO>
                  <APP_NAME>MEHR CHADDAH</APP_NAME>
                  <APP_IOP_FLG>I</APP_IOP_FLG>
                  <APP_STATUS>02</APP_STATUS>
                  <APP_PAN_DOB>25-11-1990 00:00:00</APP_PAN_DOB>
                  <APP_ENTRYDT>01-09-2017 17:52:21</APP_ENTRYDT>
                  <APP_STATUSDT>17-02-2020 14:52:14</APP_STATUSDT>
                  <APP_MODDT>04-03-2021 14:35:21</APP_MODDT>
                  <APP_POS_CODE>CDSL</APP_POS_CODE>
                  <APP_STATUS_DELTA />
                  <APP_UPDT_STATUS>01</APP_UPDT_STATUS>
                  <APP_HOLD_DEACTIVE_RMKS />
                  <APP_UPDT_RMKS />
                  <APP_KYC_MODE>0</APP_KYC_MODE>
                  <APP_IPV_FLAG>Y</APP_IPV_FLAG>
                  <APP_UBO_FLAG />
               </APP_PAN_INQ>
               <APP_PAN_INQ>
                  <APP_PAN_NO>AAAPA1111A</APP_PAN_NO>
                  <APP_NAME>MEHR CHADDAH</APP_NAME>
                  <APP_IOP_FLG>I</APP_IOP_FLG>
                  <APP_STATUS>02</APP_STATUS>
                  <APP_PAN_DOB>25-11-1990 00:00:00</APP_PAN_DOB>
                  <APP_ENTRYDT>01-09-2017 17:52:21</APP_ENTRYDT>
                  <APP_STATUSDT>17-02-2020 14:52:14</APP_STATUSDT>
                  <APP_MODDT>04-03-2021 14:35:21</APP_MODDT>
                  <APP_POS_CODE>CDSL</APP_POS_CODE>
                  <APP_STATUS_DELTA />
                  <APP_UPDT_STATUS>01</APP_UPDT_STATUS>
                  <APP_HOLD_DEACTIVE_RMKS />
                  <APP_UPDT_RMKS />
                  <APP_KYC_MODE>0</APP_KYC_MODE>
                  <APP_IPV_FLAG>Y</APP_IPV_FLAG>
                  <APP_UBO_FLAG />
               </APP_PAN_INQ>
               <APP_PAN_SUMM>
                  <APP_OTHKRA_CODE>CDSL</APP_OTHKRA_CODE>
                  <APP_OTHKRA_BATCH>20210529</APP_OTHKRA_BATCH>
                  <APP_REQ_DATE>26/05/2021</APP_REQ_DATE>
                  <APP_RESPONSE_DATE>31-05-2021 09:52:35</APP_RESPONSE_DATE>
                  <APP_TOTAL_REC>2</APP_TOTAL_REC>
               </APP_PAN_SUMM>
            </APP_RES_ROOT>
         </PANValidationResult>
      </PANValidationResponse>
   </soap12:Body>
</soap12:Envelope>
```

#### Error

[Refer to error messages] (https://redmine.cvlkra.remiges.tech/projects/cvl-kra/wiki/Messages#Messages)

##### `Single PAN`
```xml
<?xml version="1.0" encoding="UTF-8"?>
<soap12:Envelope xmlns:soap12="http://www.w3.org/2003/05/soap-envelope" xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
   <soap12:Body>
      <PANValidationResponse xmlns="https://test.cvlkra.com/">
         <PANValidationResult>
            <APP_REQ_ROOT>
               <ERROR>
                  <ERROR_CODE>WEBERR-001</ERROR_CODE>
                  <ERROR_MSG>Invalid User ID / PosCode / Password / Access Privilege Not SET</ERROR_MSG>
               </ERROR>
            </APP_REQ_ROOT>
         </PANValidationResult>
      </PANValidationResponse>
   </soap12:Body>
</soap12:Envelope>
```

##### `Bulk PAN`

###### `All Failed`
```xml
<?xml version="1.0" encoding="UTF-8"?>
<soap12:Envelope xmlns:soap12="http://www.w3.org/2003/05/soap-envelope" xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
   <soap12:Body>
      <PANValidationResponse xmlns="https://test.cvlkra.com/">
         <PANValidationResult>
            <APP_RES_ROOT>
               <APP_PAN_INQ>
                  <APP_PAN_NO>AAAPA1111A</APP_PAN_NO>
                  <APP_NAME />
                  <APP_IOP_FLG>RS</APP_IOP_FLG>
                  <APP_STATUS>05</APP_STATUS>
                  <APP_PAN_DOB />
                  <APP_ENTRYDT />
                  <APP_STATUSDT />
                  <APP_MODDT />
                  <APP_POS_CODE>CVL</APP_POS_CODE>
                  <APP_STATUS_DELTA />
                  <APP_UPDT_STATUS>05</APP_UPDT_STATUS>
                  <APP_HOLD_DEACTIVE_RMKS />
                  <APP_UPDT_RMKS />
                  <APP_KYC_MODE>0</APP_KYC_MODE>
                  <APP_IPV_FLAG />
                  <APP_UBO_FLAG />
                  <APP_PER_ADD_PROOF />
                  <APP_COR_ADD_PROOF />
               </APP_PAN_INQ>
               <APP_PAN_INQ>
                  <APP_PAN_NO>ABWPV8985E</APP_PAN_NO>
                  <APP_NAME />
                  <APP_IOP_FLG>RS</APP_IOP_FLG>
                  <APP_STATUS>05</APP_STATUS>
                  <APP_PAN_DOB />
                  <APP_ENTRYDT />
                  <APP_STATUSDT />
                  <APP_MODDT />
                  <APP_POS_CODE>CVL</APP_POS_CODE>
                  <APP_STATUS_DELTA />
                  <APP_UPDT_STATUS>05</APP_UPDT_STATUS>
                  <APP_HOLD_DEACTIVE_RMKS />
                  <APP_UPDT_RMKS />
                  <APP_KYC_MODE>0</APP_KYC_MODE>
                  <APP_IPV_FLAG />
                  <APP_UBO_FLAG />
                  <APP_PER_ADD_PROOF />
                  <APP_COR_ADD_PROOF />
               </APP_PAN_INQ>
               <APP_PAN_SUMM>
                  <APP_OTHKRA_CODE>CDSL</APP_OTHKRA_CODE>
                  <APP_OTHKRA_BATCH>20210529</APP_OTHKRA_BATCH>
                  <APP_REQ_DATE>26/05/2021</APP_REQ_DATE>
                  <APP_RESPONSE_DATE>31-05-2021 09:52:35</APP_RESPONSE_DATE>
                  <APP_TOTAL_REC>2</APP_TOTAL_REC>
               </APP_PAN_SUMM>
            </APP_RES_ROOT>
         </PANValidationResult>
      </PANValidationResponse>
   </soap12:Body>
</soap12:Envelope>
```

###### `Partially Failed`
```xml
<?xml version="1.0" encoding="UTF-8"?>
<soap12:Envelope xmlns:soap12="http://www.w3.org/2003/05/soap-envelope" xmlns:xsd="http://www.w3.org/2001/XMLSchema" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
   <soap12:Body>
      <PANValidationResponse xmlns="https://test.cvlkra.com/">
         <PANValidationResult>
            <APP_RES_ROOT>
               <APP_PAN_INQ>
                  <APP_PAN_NO>AAAPA1111A</APP_PAN_NO>
                  <APP_NAME>MEHR CHADDAH</APP_NAME>
                  <APP_IOP_FLG>I</APP_IOP_FLG>
                  <APP_STATUS>02</APP_STATUS>
                  <APP_PAN_DOB>25-11-1990 00:00:00</APP_PAN_DOB>
                  <APP_ENTRYDT>01-09-2017 17:52:21</APP_ENTRYDT>
                  <APP_STATUSDT>17-02-2020 14:52:14</APP_STATUSDT>
                  <APP_MODDT>04-03-2021 14:35:21</APP_MODDT>
                  <APP_POS_CODE>CDSL</APP_POS_CODE>
                  <APP_STATUS_DELTA />
                  <APP_UPDT_STATUS>01</APP_UPDT_STATUS>
                  <APP_HOLD_DEACTIVE_RMKS />
                  <APP_UPDT_RMKS />
                  <APP_KYC_MODE>0</APP_KYC_MODE>
                  <APP_IPV_FLAG>Y</APP_IPV_FLAG>
                  <APP_UBO_FLAG />
               </APP_PAN_INQ>
               <APP_PAN_INQ>
                  <APP_PAN_NO>ABWPV8985E</APP_PAN_NO>
                  <APP_NAME />
                  <APP_IOP_FLG>RS</APP_IOP_FLG>
                  <APP_STATUS>05</APP_STATUS>
                  <APP_PAN_DOB />
                  <APP_ENTRYDT />
                  <APP_STATUSDT />
                  <APP_MODDT />
                  <APP_POS_CODE>CVL</APP_POS_CODE>
                  <APP_STATUS_DELTA />
                  <APP_UPDT_STATUS>05</APP_UPDT_STATUS>
                  <APP_HOLD_DEACTIVE_RMKS />
                  <APP_UPDT_RMKS />
                  <APP_KYC_MODE>0</APP_KYC_MODE>
                  <APP_IPV_FLAG />
                  <APP_UBO_FLAG />
                  <APP_PER_ADD_PROOF />
                  <APP_COR_ADD_PROOF />
               </APP_PAN_INQ>
               <APP_PAN_SUMM>
                  <APP_OTHKRA_CODE>CDSL</APP_OTHKRA_CODE>
                  <APP_OTHKRA_BATCH>20210529</APP_OTHKRA_BATCH>
                  <APP_REQ_DATE>26/05/2021</APP_REQ_DATE>
                  <APP_RESPONSE_DATE>31-05-2021 09:52:35</APP_RESPONSE_DATE>
                  <APP_TOTAL_REC>2</APP_TOTAL_REC>
               </APP_PAN_SUMM>
            </APP_RES_ROOT>
         </PANValidationResult>
      </PANValidationResponse>
   </soap12:Body>
</soap12:Envelope>
```
