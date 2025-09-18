# Testing Requirements

- The test framework automatically generates today's date in the correct format for each test run.
- Each test case gets a descriptive name from the CSV data that shows what is being tested.
- The framework parses XML responses and extracts specific field values for validation.
- Tests verify individual XML fields rather than checking the entire response as a string.
- Every test automatically checks that the API responds within 5 seconds.
