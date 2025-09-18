// Define test data for smoke tests
const smokeTestData = {
    'TC_INQ_REQ_001': {
        input_pan_no: '',
        input_dob: '',
        input_iop_flag: '',
        input_pos_code: '',
        input_okra_code: '',
        input_okra_batch: '',
        input_request_date: '',
        input_total_records: '',
        expected_http_status: '400',
        expected_app_status: ''
    },
    'TC_INQ_REQ_002': {
        input_pan_no: 'OVSWF8950H',
        input_dob: '28-08-1996',
        input_iop_flag: 'IE',
        input_pos_code: 'A1249',
        input_okra_code: 'CVL',
        input_okra_batch: '1234',
        input_request_date: new Date().toLocaleDateString('en-GB').split('/').join('-'),
        input_total_records: '1',
        expected_http_status: '200',
        expected_app_status: 'Success'
    },
    'TC_INQ_REQ_004': {
        input_pan_no: '!@#$%^&*',
        input_dob: '28-08-1996',
        input_iop_flag: 'IE',
        input_pos_code: 'A1249',
        input_okra_code: 'CVL',
        input_okra_batch: '1234',
        input_request_date: new Date().toLocaleDateString('en-GB').split('/').join('-'),
        input_total_records: '1',
        expected_http_status: '200',
        expected_app_status: 'WEBERR-999'
    }
};

// Extract test ID from request name
const requestName = pm.info.requestName;
const testIdMatch = requestName.match(/TC_INQ_REQ_\d{3}/);

if (testIdMatch) {
    const testId = testIdMatch[0];
    const testData = smokeTestData[testId];

    if (testData) {
        // Set variables from test data
        Object.keys(testData).forEach(key => {
            pm.variables.set(key, testData[key]);
        });

        console.log(`[START] Smoke Test: ${testId}`);
    }
}
