// Get expected values from CSV data (automatically set by Newman)
const expectedHttpStatus = parseInt(pm.variables.get('expected_http_status'));
const expectedAppStatus = pm.variables.get('expected_app_status');
const testId = pm.variables.get('test_id');
const testDescription = pm.variables.get('test_description');

// Test 1: HTTP Status Code
pm.test(`${testId} - HTTP Status should be ${expectedHttpStatus}`, function () {
    pm.response.to.have.status(expectedHttpStatus);
});

// Test 2: Response Time
pm.test(`${testId} - Response time (< 5000ms)`, function () {
    pm.expect(pm.response.responseTime).to.be.below(5000);
});

// Test 3: Content Type
pm.test(`${testId} - Response content type is XML`, function () {
    const contentType = pm.response.headers.get('Content-Type');
    pm.expect(contentType).to.include('xml');
});

// Parse XML and validate response fields
const responseText = pm.response.text();
const xml2js = require('xml2js');
const parseString = xml2js.parseString;

parseString(responseText, { explicitArray: false }, (err, result) => {
    if (err) {
        console.error('Error parsing XML:', err);
        pm.test(`${testId} - XML response parseable`, function () {
            pm.expect.fail('Failed to parse XML response: ' + err.message);
        });
        return;
    }

    try {
        // Extract response structure
        const envelope = result['soap12:Envelope'] || result['soap:Envelope'];
        const body = envelope['soap12:Body'] || envelope['soap:Body'];
        const panValidationResponse = body['PANValidationResponse'];
        const panValidationResult = panValidationResponse['PANValidationResult'];
        const appResRoot = panValidationResult['APP_RES_ROOT'];

        if (expectedHttpStatus === 200) {
            // Validate response structure
            pm.test(`${testId} - Response has correct SOAP structure`, function () {
                pm.expect(appResRoot).to.be.an('object');
                pm.expect(appResRoot['APP_PAN_INQ']).to.be.an('object');
                pm.expect(appResRoot['APP_PAN_SUMM']).to.be.an('object');
            });

            const panInq = appResRoot['APP_PAN_INQ'];
            const panSumm = appResRoot['APP_PAN_SUMM'];

            // Test APP_STATUS field specifically
            pm.test(`${testId} - APP_STATUS should be ${expectedAppStatus}`, function () {
                pm.expect(panInq['APP_STATUS']).to.equal(expectedAppStatus);
            });

            // Validate PAN number is returned
            const inputPan = pm.variables.get('input_pan_no');
            if (inputPan) {
                pm.test(`${testId} - PAN number echoed correctly`, function () {
                    pm.expect(panInq['APP_PAN_NO']).to.equal(inputPan);
                });
            }

            // Validate OKRA code is returned
            const inputOkraCode = pm.variables.get('input_okra_code');
            if (inputOkraCode) {
                pm.test(`${testId} - OKRA code preserved`, function () {
                    pm.expect(panSumm['APP_OTHKRA_CODE']).to.equal(inputOkraCode);
                });
            }

            // Validate OKRA batch is returned
            const inputOkraBatch = pm.variables.get('input_okra_batch');
            if (inputOkraBatch) {
                pm.test(`${testId} - OKRA batch preserved`, function () {
                    pm.expect(panSumm['APP_OTHKRA_BATCH']).to.equal(String(inputOkraBatch));
                });
            }

            // Validate total records is returned
            const inputTotalRecords = pm.variables.get('input_total_records');
            if (inputTotalRecords) {
                pm.test(`${testId} - Total records preserved`, function () {
                    pm.expect(panSumm['APP_TOTAL_REC']).to.equal(String(inputTotalRecords));
                });
            }

            // Validate response date format
            pm.test(`${testId} - Response date is present`, function () {
                const responseDate = panSumm['APP_RESPONSE_DATE'];
                pm.expect(responseDate).to.be.a('string');
                pm.expect(responseDate.length).to.be.greaterThan(0);
            });

        } else {
            // For non-200 responses, still try to parse if XML structure exists
            console.log(`[${testId}] Non-200 response - checking for error structure`);
        }

    } catch (parseError) {
        console.error('Error extracting fields from parsed XML:', parseError);
        pm.test(`${testId} - XML structure should be valid`, function () {
            pm.expect.fail('Failed to extract fields from XML: ' + parseError.message);
        });
    }
});

// Test 5: SOAP Envelope Structure
pm.test(`${testId} - Valid SOAP response structure`, function () {
    pm.expect(responseText).to.match(/<soap.*:Envelope|<.*:Envelope/);
    pm.expect(responseText).to.include('PANValidation');
});

// Log test completion
if (pm.test && pm.test.results && pm.test.results.length > 0) {
    const testResults = pm.test.results;
    const passed = testResults.filter(result => result.pass).length;
    const total = testResults.length;
    const status = passed === total ? '[PASS] PASS' : '[FAIL] FAIL';

    console.log(`${status} ${testId}: ${passed}/${total} tests passed`);

    if (passed < total) {
        const failures = testResults.filter(result => !result.pass);
        console.log('[FAIL] Failed assertions:');
        failures.forEach(failure => {
            console.log(`   - ${failure.name}`);
        });
    }
} else {
    console.log(`[COMPLETE] ${testId} - Tests completed`);
}
