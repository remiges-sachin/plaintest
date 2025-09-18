// Smoke test validation - only run for PAN validation requests
// Skip collection-level tests for Get Password request
if (!pm.info.requestName.includes('Get Password')) {
    const expectedHttpStatus = parseInt(pm.variables.get('expected_http_status'));
    const expectedAppStatus = pm.variables.get('expected_app_status');

    // Validate HTTP Status
    if (!isNaN(expectedHttpStatus)) {
        pm.test(`[PASS] HTTP Status: ${expectedHttpStatus}`, function () {
            pm.response.to.have.status(expectedHttpStatus);
        });
    }

    // Validate Response Time
    pm.test('[PASS] Response Time (< 3000ms)', function () {
        pm.expect(pm.response.responseTime).to.be.below(3000);
    });

    // Validate XML response structure
    pm.test('[PASS] Valid XML Response', function () {
        const responseText = pm.response.text();
        pm.expect(responseText).to.include('PANValidation');
        pm.expect(responseText).to.match(/<\?xml|<soap/);
    });

    // Validate Application Status
    if (expectedAppStatus) {
        const responseText = pm.response.text();

        if (expectedAppStatus === 'Success') {
            pm.test('[PASS] Success Response', function () {
                pm.expect(responseText).to.include('APP_RES_ROOT');
            });
        } else {
            pm.test(`[PASS] Error Code: ${expectedAppStatus}`, function () {
                pm.expect(responseText).to.include(expectedAppStatus);
            });
        }
    }
}

// Log test execution result
if (pm.test.results && pm.test.results.length > 0) {
    const allPassed = pm.test.results.every(result => result.pass);
    const status = allPassed ? '[SUCCESS] SMOKE PASS' : '[FAILURE] SMOKE FAIL';
    console.log(`${status} - ${pm.info.requestName}`);
}
