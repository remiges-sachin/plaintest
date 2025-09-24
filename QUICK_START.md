# PlainTest Quick Start Guide

Follow these steps to run your first tests.

## Before You Start

Check PlainTest is installed:
```bash
./plaintest version
```

See available test collections:
```bash
./plaintest list collections
```

Available collections in this project:
- `smoke` - Basic health check tests
- `get_auth` - Authentication setup
- `api_tests` - Main API testing
- `test-collection` - Additional test scenarios

## Step 1: Run Your First Test

Start with smoke tests (no authentication needed):
```bash
./plaintest run --test smoke
```

What happened:
- PlainTest ran the smoke test collection
- Tests executed against a test API
- Results showed in your terminal

## Step 2: Run Tests with Authentication

Many tests need login first. Use `--setup` for authentication:
```bash
./plaintest run --setup get_auth --test api_tests
```

What happened:
- `get_auth` ran once to login
- Login token was saved
- `api_tests` ran using that token

## Step 3: Run Tests with CSV Data

Test multiple scenarios from a CSV file:
```bash
./plaintest run --setup get_auth --test api_tests -d example
```

What happened:
- `get_auth` ran once (not for each CSV row)
- `api_tests` ran once for each row in example.csv
- Each row tested a different scenario

## Step 4: Test Specific Rows

Test only rows 1-3 from your CSV:
```bash
./plaintest run --setup get_auth --test api_tests -d example -r 1-3
```

## Step 5: Multi-Collection Testing

Run multiple test collections with authentication:
```bash
./plaintest run --setup get_auth --test api_tests --test test-collection -d example
```

What happened:
- `get_auth` ran once for authentication
- `api_tests` ran for each CSV row
- `test-collection` ran for each CSV row
- Both test collections used the same auth token

## Common Patterns

### Pattern 1: Simple Test
```bash
./plaintest run --test <collection_name>
```

### Pattern 2: Auth + Test
```bash
./plaintest run --setup <auth_collection> --test <test_collection>
```

### Pattern 3: Auth + CSV Test
```bash
./plaintest run --setup <auth_collection> --test <test_collection> -d <data_file>
```

### Pattern 4: Multiple Collections
```bash
./plaintest run --setup <auth_collection> --test <test1> --test <test2> -d <data_file>
```

## Understanding Results

**All tests passed!** - Everything worked
**Tests failed with exit code** - Some tests failed
**Reports generated in reports/** - Check HTML report for details

## Add Reports

Generate HTML reports:
```bash
./plaintest run --setup get_auth --test api_tests --reports
```

Find reports in `reports/` folder.

## Troubleshooting

**"Collection not found"**
Run `./plaintest list collections` to see available names.

**"Environment not found"**
Run `./plaintest list environments` to see available environments.

**"CSV file not found"**
Run `./plaintest list data` to see available data files. Use names without `.csv` extension.

## Next Steps

- Read [USER_DOCS.md](USER_DOCS.md) for all commands
- Check `data/` folder for CSV examples
- Try different collections from `./plaintest list collections`
