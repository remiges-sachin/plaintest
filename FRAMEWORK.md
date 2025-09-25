# PlainTest Framework

PlainTest makes API testing efficient by separating concerns.

PlainTest is a functional and regression testing framework. It validates API behavior and correctness, not performance or load capacity.

Three things to remember:
1. Smoke tests run first (5 tests, 30 seconds)
2. CSV data drives iteration
3. Setup runs once, tests run many times

## The PlainTest Framework

You run 100 API tests. Twenty fail. Which layer broke?

PlainTest organizes tests into layers:
- **Smoke** - Is the API alive? (5 tests, 30 seconds)
- **Setup** - Get credentials once
- **Tests** - Run your CSV rows
- **Reports** - See what happened

Example: You're testing user registration with 100 different email addresses.

Without layers, you authenticate 100 times (once per CSV row). When registration test 47 fails, you don't know if authentication failed or the email validation failed.

With layers, authentication runs once in setup. Test 47 uses that authentication. When it fails, you know it's the email data, not auth.

The PlainTest CLI implements these layers. The `--setup` flag runs once. The `--test` flag iterates with CSV data.

You could apply these layers manually with Newman. The CLI just makes it automatic.

## Project Structure

```
plaintest-project/
├── collections/     # Test logic
├── scripts/        # Extracted JavaScript files
├── data/           # Test cases
├── environments/   # Configurations
└── reports/        # Results
```

Collections contain your Postman exports. Data holds CSV files with test cases. Environments store URLs and settings. Reports capture what happened.

This structure separates test logic from test scripts from test data from test configuration.

## Practical Problems PlainTest Solves

PlainTest was built to solve real testing problems:

**Debugging during development** - Use `-r 3` to run only CSV row 3 while fixing that specific test case.

**Avoiding redundant operations** - Setup phase runs once (authentication, database setup). Test phase iterates only the actual tests you want to repeat. This prevents 100 authentication calls when you have 100 CSV rows.

**Collection chaining** - Run authentication, then API tests. Tokens automatically flow between collections.

**CI/CD integration** - Runs Postman collections from command line for CI. Exit codes indicate success/failure. JSON reports contain full request/response data for automated analysis.

**Detailed reporting** - HTML reports for humans to review. JSON reports for machines to process.

Each feature addresses a specific pain point in API testing workflows.

## Start with Smoke Tests

What fails first when your API breaks?

Write 5 tests that answer this question:

```
1. Health endpoint responds
2. Authentication works
3. Main resource loads
4. Database connects
5. Third-party service responds
```

Run these in under 30 seconds. If any fail, stop. Don't run 200 tests when your API is down.

Example smoke collection:

```
GET {{base_url}}/health → 200 OK
POST {{base_url}}/auth → returns token
GET {{base_url}}/users → returns array
GET {{base_url}}/accounts → returns array
GET {{base_url}}/external/status → 200 OK
```

Your smoke tests are your canary. They tell you if it's safe to proceed.

## CSV Structure: Three Columns

CSV files have three types of columns:

**META** - Test information
- `test_id` - unique identifier
- `test_name` - what you're testing

**INPUT** - Request data
- `input_email` - goes in request body
- `input_age` - goes in request body
- `input_type` - goes in request headers

**EXPECTED** - Response validation
- `expected_status` - HTTP status code
- `expected_message` - error message
- `expected_total` - calculated values

Example CSV structure:

| test_id       | test_name              | input_email     | input_age | expected_status | expected_message              |
|---------------|------------------------|-----------------|-----------|-----------------|-------------------------------|
| valid_user    | Standard registration  | user@test.com   | 25        | 200            | Account created               |
| invalid_email | Bad email format       | notanemail      | 25        | 400            | Invalid email format          |
| underage      | Too young              | kid@test.com    | 16        | 400            | Must be 18 or older          |
| edge_case     | Exactly minimum age    | teen@test.com   | 18        | 200            | Account created               |
| senior        | Senior discount        | senior@test.com | 65        | 200            | Account created with discount |

Your actual CSV file contains these same rows as comma-separated values:
```
test_id,test_name,input_email,input_age,expected_status,expected_message
valid_user,Standard registration,user@test.com,25,200,Account created
```

One endpoint, five test cases. Same request logic, different data each time.

## Authentication Pattern

Use service accounts for consistent authentication.

Service accounts are designed for automation, not humans. They don't have human account restrictions:
- No password expiry
- No MFA requirements
- No account lockouts from failed attempts
- No "change password every 90 days" policies

This makes them reliable for automated testing.

Create a setup collection that gets your token:

```javascript
// In setup collection, after login request
pm.environment.set("auth_token", pm.response.json().access_token);
```

Then use that token in your test collections:

```javascript
pm.request.headers.add({
    key: "Authorization",
    value: "Bearer {{auth_token}}"
});
```

Setup runs once and sets the token. Tests use that token for all iterations.

## Service Account Setup

You need three things:

**Keycloak Client**: Create a client with service account enabled. Set client authentication to ON. Turn off everything else except "Service Account Roles".

**Service User**: Keycloak creates this automatically. It's named `service-account-[client-id]`. Since this isn't a human user, human account rules don't apply.

**Capabilities Mapper**: Add a user attribute mapper that puts capabilities in the token. Map user attribute `caps` to token claim `caps`.

This gives you a service account that can authenticate and carry the right permissions for testing without human account complications.

## Test Organization

**Collections** = business flows
- `user_registration` - complete user signup flow
- `order_processing` - complete order lifecycle
- `payment_integration` - complete payment flow

**Folders** = scenarios within flows
- `happy_path` - normal operation
- `error_cases` - validation failures
- `edge_cases` - boundary conditions

**Requests** = individual API calls
- `create_user` - POST to /users
- `validate_email` - GET to /users/validate
- `send_welcome` - POST to /notifications

One collection tests one business capability. Folders organize different scenarios of that capability.

## Collection Organization Strategies

**Use one collection with folders when:**
- Same API and base URL
- Same authentication method
- Want single file to manage

Example structure:
```
user_api/
├── Auth/
│   └── Login
├── Users/
│   ├── Create
│   └── Update
└── Admin/
    └── Delete
```

**Use separate collections when:**
- Different APIs or services
- Different run schedules (smoke vs regression)
- Different team ownership
- Need clean separation between setup and tests

Example structure:
```
collections/
├── smoke.json         # Quick health checks
├── auth.json          # Just authentication
└── user_tests.json    # User operations
```

Choose based on how you want to organize and run your tests.

## Script Management

Postman stores scripts as strings in JSON files. You can't edit strings in your IDE. You can't see diffs properly. You can't run linters.

PlainTest extracts scripts to JavaScript files.

Why? Scripts are code. Code needs proper tools.

Structure after extraction:
```
scripts/
├── user_api/
│   ├── _collection__test.js      # Collection-level tests
│   └── create_user__test.js      # Request-specific tests
```

Benefits:
- Edit in VS Code with syntax highlighting
- Git shows real code changes
- Use ESLint and prettier

Extract once: `plaintest scripts pull user_api`
Edit as JavaScript.
Push back: `plaintest scripts push user_api`

The principle: Separate test structure (collection) from test logic (scripts). Each lives in its natural format.

## Running Tests

Three-phase execution pattern:

**Phase 1: Smoke** (30 seconds)
```bash
plaintest run smoke
```

If smoke fails, stop. Your API isn't ready for testing.

**Phase 2: Setup** (runs once)
```bash
plaintest run --setup get_auth --test user_tests -d user_data.csv
```

Setup gets authentication token. Stores it in environment.

**Phase 3: Test** (iterates with CSV)
Test phase runs once per CSV row. Each row gets the same token from setup.

This pattern gives you predictable authentication and repeatable test data.

## CSV Data Drives Iteration

One CSV file = one test scenario with multiple cases.

```csv
test_id,input_pan,input_name,expected_status
valid_pan,ABCDE1234F,John Doe,200
invalid_format,INVALID,John Doe,400
missing_name,ABCDE1234F,,400
numeric_pan,1234567890,John Doe,400
special_chars,ABCDE@234F,John Doe,400
```

Five rows = five test executions of the same endpoint.

Each row tests a different validation rule. Same collection logic, different input data.

Debug specific tests during development:

```bash
plaintest run user_tests -d user_data.csv -r 3
```

Test 3 failing? Run just that row while you fix it. No need to run all 100 tests to debug one.

```bash
plaintest run user_tests -d user_data.csv -r 2-5  # Debug rows 2 through 5
plaintest run user_tests -d user_data.csv -r 1,3,5  # Debug specific rows
```

The `-r` flag lets you focus on the exact test that's broken.

## Directory Structure in Practice

Example project structure:

```
user-api-tests/
├── collections/
│   ├── smoke.postman_collection.json
│   ├── auth.postman_collection.json
│   ├── user_registration.postman_collection.json
│   └── user_registration_csv.postman_collection.json
├── data/
│   ├── user_registration_data.csv
│   ├── user_validation_data.csv
│   └── email_validation_data.csv
├── environments/
│   ├── ci.postman_environment.json
│   ├── dev.postman_environment.json
│   └── staging.postman_environment.json
└── reports/
    ├── smoke_20240924T103045.html
    └── user_tests_20240924T103045.json
```

Collections pair up: `user_registration` generates dynamic data, `user_registration_csv` uses predefined data.

Data files match collection names. Environment files match deployment stages.

## Reports Tell You What Happened

Two report types:

**JSON** - Machine readable. Contains full request/response data. Use for debugging and automation.

**HTML** - Human readable. Contains formatted test results. Use for manual review and sharing.

Example JSON query:

```bash
jq '.run.executions[] | select(.response.code >= 400)' reports/user_tests_*.json
```

This shows all failed requests with their full context.

## When Tests Fail

Smoke test failure = environment problem. Check if your API is running.

Setup test failure = authentication problem. Check service account credentials.

Test iteration failure = data problem. Check your CSV values and expected results.

Each failure type has a different root cause. The three-phase pattern helps you identify which type you're dealing with.

## Test Data Principles

**Use explicit data, not generated data** for regression tests.

Generated data changes every run. This makes debugging hard. Use CSV files with known values for tests you run repeatedly.

```csv
# Good - explicit values
test_id,input_pan,input_name,expected_status
known_valid,ABCDE1234F,Test User,200
known_invalid,INVALID123,Test User,400

# Bad - formulas and generation
test_id,input_pan,input_name,expected_status
random,=RANDBETWEEN(),=CONCATENATE(),200
```

Save generated data for exploratory testing. Use explicit data for automated testing.

## Collection Types

**Standard collections** generate fresh data each run. Use for development testing where you need unique values.

**CSV collections** read predefined data from files. Use for regression testing where you need consistent results.

Both test the same endpoints. The difference is data source.

Example commands:

```bash
# Fresh data each time
plaintest run auth user_registration -e ci

# Same data each time
plaintest run auth user_registration_csv -e ci -d user_data.csv
```

Choose based on whether you need consistency or uniqueness.

## Response Time Validation

PlainTest validates response times, not system performance.

PlainTest runs tests sequentially through Newman. It cannot simulate concurrent users or measure system load. For load testing, use k6, JMeter, or Locust.

What PlainTest can do:

**Assert response times** in test scripts:

```javascript
pm.test("Response time under 500ms", function () {
    pm.expect(pm.response.responseTime).to.be.below(500);
});
```

**Set timeout limits** with Newman flags:

```bash
plaintest run user_tests --timeout 5000  # 5 second timeout
```

This validates that your API responds within acceptable time limits during functional testing. For actual performance testing with load simulation, use dedicated tools.

## The Framework in Action

Complete workflow:

1. **Start**: `plaintest run smoke` (is API alive?)
2. **Auth**: `plaintest run --setup auth --test user_tests -d user_data.csv`
3. **Results**: Check `reports/` directory for HTML summary and JSON details

This pattern works for any API. Change the collection names and CSV data, but keep the three-phase structure.

The framework separates concerns: authentication from testing, test logic from test data, success from failure analysis.

That separation makes your API testing predictable.
