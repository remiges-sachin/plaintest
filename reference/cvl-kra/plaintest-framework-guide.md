# PlainTest Framework

## Introduction

PlainTest is testing framework for API using Postman. PlainTest solve big problem: many tester not programmer but still need test API good.

PlainTest philosophy simple:
- If tester not understand test, test too complex
- Test what business care about, not what computer science book say
- Start small, grow confident
- No complexity demon spirit in test code

PlainTest make regression testing possible for all grug, not just big brain developer grug.

## Smoke Tests

Smoke test is first line of defense. Like checking if fire make smoke before cooking mammoth.

### What is Smoke Test?

Smoke test answer one question: "Is API alive or dead?"

Not care if API perfect. Not care if response pretty. Just care if API respond and not on fire.

### PlainTest Smoke Rules

1. **Maximum 5 tests** - more than 5 not smoke test anymore
2. **Run under 30 seconds** - if take longer, too complex
3. **Test only critical endpoints** - login, health check, main resource
4. **Check only basics**:
   - API return 200 OK?
   - API return JSON?
   - Response time under 5 seconds?

### Example Smoke Test

```
Smoke Test Collection/
├── 1_API_is_alive
│   GET {{base_url}}/health
│   Test: Status code is 200
│
├── 2_Can_login
│   POST {{base_url}}/login
│   Test: Status code is 200
│   Test: Response has token
│
└── 3_Main_endpoint_work
    GET {{base_url}}/api/products
    Test: Status code is 200
    Test: Response is array
```

### When Run Smoke Test?

- Before deploy new version
- Every morning automatic
- Before running big test suite
- When someone say "API not working"

### Smoke Test Fail - What Do?

If smoke test fail, stop everything! No point running 500 tests if API dead.

Tell developer immediately. Smoke test fail = emergency.

Remember: smoke test like canary in coal mine. When canary stop singing, grug leave mine fast!

## Postman Organization

PlainTest use three level organization:

**Workspace** = your team testing space
- One workspace per project/team
- Contains all collections
- Shares environment variables

**Collection** = group of related tests
- Separate collection for each test type (smoke, regression, edge cases)
- Can run entire collection at once
- Export/import as JSON file

**Folder** = organize inside collection
- Group by feature or endpoint
- Just for organization
- No folder nesting - keep flat!

### PlainTest Structure
```
PlainTest Workspace/
├── SMOKE_TESTS (collection)
├── Critical_Path_Tests (collection)
├── Edge_Cases (collection)
└── Environments (shared)
    ├── Dev
    ├── Staging
    └── Production
```

Keep smoke test in own collection always! Different purpose, different schedule.

### Collection vs Folder Choice

**Use separate collections when:**
- Different purpose (smoke vs regression)
- Different run schedule (hourly vs nightly)
- Different ownership (payment team vs user team)
- Different API altogether

**Use folders within collection when:**
- Different scenarios of same flow
- Different test data sets
- Different branches of business logic

**Example - Good folder use:**
```
Collection: Order_Tests/
├── Folder: Happy_Path/
│   ├── Create Order
│   └── Payment Success
├── Folder: Payment_Decline/
│   ├── Create Order
│   └── Payment Fail
└── Folder: Out_of_Stock/
    ├── Create Order
    └── Stock Check Fail
```

All related scenarios in one collection, organized by folder. Can run all or just one scenario!

## Critical Path Tests

## Edge Case Tests

## Functional Testing Focus

PlainTest primarily do functional testing - verify API work correct!

**What PlainTest Check (Functional):**
- Correct response for valid input
- Correct error for invalid input
- Business rules applied right
- Data saved/updated/deleted properly

**Example Functional Tests:**
```
POST /order with valid data → Order created with ID
POST /order with no payment → Return 400 error
GET /order/123 → Return order 123 details
DELETE /order/123 → Order removed from system
```

**Not Main Focus (Non-Functional):**
- Response time (performance)
- Load handling (scalability)
- Security vulnerabilities

PlainTest can include basic performance check (like timeout), but main job is ensure API do what supposed to do!

### Key Understanding: Functional Testing vs Structure

**Important:** Functional testing NOT separate collection or folder!

Functional testing is WHAT you test (correctness), not HOW you organize tests!

**Every PlainTest collection does functional testing:**
```
SMOKE_TESTS collection → functional tests (does API work?)
User_API_Tests collection → functional tests (user operations correct?)
Order_API_Tests collection → functional tests (order logic correct?)
```

**Structure based on business logic, not test type:**
- Collection = API or feature area (Users, Orders, Payments)
- Folder = scenario or operation (Create, Update, Delete, Error_Cases)
- Request = specific test case (Valid_User, Invalid_Email)
- Tests tab = where functional testing happen (assertions!)

**Remember:** The assertions you write in Tests tab = the functional testing. Every test checking "does API do what supposed to do?" is functional test!

## Test Data Management

## Environment Setup

## Writing Good Assertions

## Test Organization

## Running Tests

## When Test Fail

## Reporting

## Maintenance

## Common Mistakes
