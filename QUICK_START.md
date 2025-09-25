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

## Step 1: Run Your First Test

Start with smoke tests (no authentication needed):
```bash
./plaintest run --test smoke
```

## Step 2: Run Tests with Authentication

Many tests need login first. Use `--setup` for authentication:
```bash
./plaintest run --setup get_auth --test api_tests
```

Setup runs once. Tests use the authentication token.

## Step 3: Run Tests with CSV Data

Test multiple scenarios from a CSV file:
```bash
./plaintest run --setup get_auth --test api_tests -d example
```

Setup runs once. Tests run for each CSV row.

## Step 4: Debug Specific Rows

Test only row 3 from your CSV:
```bash
./plaintest run --setup get_auth --test api_tests -d example -r 3
```

Perfect for debugging failing tests.

## Common Patterns

**Simple test (no auth):**
```bash
./plaintest run --test smoke
```

**Auth + test:**
```bash
./plaintest run --setup auth_collection --test test_collection
```

**Auth + test + CSV:**
```bash
./plaintest run --setup auth_collection --test test_collection -d data_file
```

**Debug specific row:**
```bash
./plaintest run --setup auth_collection --test test_collection -d data_file -r 5
```

## Next Steps

- See [CLI_REFERENCE.md](CLI_REFERENCE.md) for all commands and flags
- Read [FRAMEWORK.md](FRAMEWORK.md) to understand the testing approach
- Check [SCRIPT_SYNC.md](SCRIPT_SYNC.md) to edit Postman scripts
