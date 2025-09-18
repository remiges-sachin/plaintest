# PlainTest CSV Testing Guide

## Mental Model: Same Test, Different Data

CSV in PlainTest like assembly line - same test steps, different inputs each time!

Use CSV when testing one endpoint with many different data combinations.

## When Use CSV

**Good for CSV:**
- Boundary testing (min/max values)
- Different user types (admin/user/guest)
- Valid/invalid input combinations
- Business rule variations
- Error message validation

**NOT good for CSV:**
- Different API endpoints
- Complex conditional logic
- One-off tests
- Tests needing different setup/teardown

## Three Column Categories

Every PlainTest CSV has three types of columns:

### 1. Meta Columns (Test Information)
Start with `test_` or no prefix
- `test_id` - unique identifier for each test case
- `test_description` - what this row testing (optional)
- `test_priority` - high/medium/low (optional)

### 2. Input Columns (Data to Send)
Always start with `input_`
- `input_email` - goes in request body
- `input_quantity` - goes in request body
- `input_user_type` - goes in request header/body

### 3. Expected Columns (What Should Return)
Always start with `expected_`
- `expected_status` - HTTP status code
- `expected_message` - error/success message
- `expected_total` - calculated values

## Example CSV Structure

```csv
test_id,test_description,input_email,input_age,input_type,expected_status,expected_message
valid_adult,Normal user,john@test.com,25,standard,200,Account created
valid_senior,Senior discount,senior@test.com,65,senior,200,Account created with discount
invalid_email,Bad email format,notanemail,25,standard,400,Invalid email format
underage,Too young,kid@test.com,16,standard,400,Must be 18 or older
edge_min_age,Exactly 18,barely@test.com,18,standard,200,Account created
edge_max_age,Very old,ancient@test.com,120,standard,200,Account created
over_max_age,Too old,tooold@test.com,121,standard,400,Invalid age
missing_email,No email provided,,25,standard,400,Email required
sql_injection,Security test,admin@test.com'; DROP TABLE;,25,standard,400,Invalid email format
```

## Using in Postman

### Request Body
```json
{
    "email": "{{input_email}}",
    "age": {{input_age}},
    "accountType": "{{input_type}}"
}
```

### Tests Tab
```javascript
// Use test_id as test name for clear reporting
pm.test(pm.variables.get("test_id"), function() {
    // Check status
    const expectedStatus = parseInt(pm.variables.get("expected_status"));
    pm.response.to.have.status(expectedStatus);

    // Check message if provided
    const expectedMessage = pm.variables.get("expected_message");
    if (expectedMessage) {
        const response = pm.response.json();
        pm.expect(response.message).to.include(expectedMessage);
    }
});
```

## CSV Best Practices

### Keep It Simple
- No nested JSON in cells
- No complex formulas
- One CSV per test scenario
- Maximum 20-30 rows per CSV (more = hard maintain)

### Naming Conventions
- CSV filename = what testing: `user_registration_tests.csv`
- Columns = lowercase with underscore: `input_user_name`
- Test IDs = descriptive but short: `valid_admin`, `expired_token`

### Organization
```
Test Data/
├── smoke_tests.csv (5-10 rows max)
├── user_registration.csv
├── order_validation.csv
└── payment_processing.csv
```

### Empty Values
- Use empty cell for null/missing
- Use `"null"` (string) if testing literal "null"
- Use `0` for zero, not empty

## Running CSV Tests

### In Postman Runner
1. Select collection/folder
2. Select CSV file
3. Preview data (check mapping!)
4. Run

### Tips for Success
- Always preview CSV first - check column mapping
- Start with 1-2 rows, then expand
- Keep Excel/CSV editor open while building
- Use Excel colors: yellow=meta, blue=input, green=expected

## Common CSV Mistakes

**Mistake 1: Magic values without meaning**
```csv
Bad:  test1,user@test.com,1,200
Good: valid_standard_user,user@test.com,standard,200
```

**Mistake 2: Mixing test types**
```csv
Bad: One CSV testing login, orders, and payments
Good: Separate CSV for each
```

**Mistake 3: No prefix convention**
```csv
Bad:  email,status,message (what is input? what expected?)
Good: input_email,expected_status,expected_message
```

**Mistake 4: Complex data in cells**
```csv
Bad:  {"user":{"email":"test@test.com","age":25}}
Good: input_email,input_age (flat columns)
```

## Mental Model Summary

Think of CSV as test cases table:
- Each row = one test execution
- Meta columns = test information
- Input columns = request data
- Expected columns = assertions

Keep simple, keep organized, keep prefixes clear!

*CSV powerful but complexity demon lurk if make too complex! Start simple!*
