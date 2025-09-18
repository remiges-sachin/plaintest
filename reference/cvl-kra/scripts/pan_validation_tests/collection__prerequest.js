// Newman automatically sets all CSV column values as variables
// This runs once per CSV row (97 iterations total)

// Set current date in dd-mm-yyyy format
const today = new Date().toLocaleDateString('en-GB').split('/').join('-');
pm.variables.set('input_request_date', today);

const testId = pm.variables.get('test_id');
const testDescription = pm.variables.get('test_description');
const expectedHttp = pm.variables.get('expected_http_status');
const expectedApp = pm.variables.get('expected_app_status');

// Log current test details
console.log(`\n[TEST] Test ${testId}: ${testDescription}`);
console.log(`[INPUT] Input Data:`, {
    PAN: pm.variables.get('input_pan_no') || 'BLANK',
    DOB: pm.variables.get('input_dob') || 'BLANK',
    IOP: pm.variables.get('input_iop_flag') || 'BLANK',
    POS: pm.variables.get('input_pos_code') || 'BLANK',
    OKRA_Code: pm.variables.get('input_okra_code') || 'BLANK',
    OKRA_Batch: pm.variables.get('input_okra_batch') || 'BLANK',
    Req_Date: pm.variables.get('input_request_date') || 'BLANK',
    Total_Records: pm.variables.get('input_total_records') || 'BLANK'
});
console.log(`[EXPECTED] Expected: HTTP ${expectedHttp}, App: ${expectedApp}`);
