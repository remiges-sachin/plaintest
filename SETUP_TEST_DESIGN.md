# PlainTest Setup-Test Design

PlainTest separates setup from tests. Setup runs once. Tests iterate over data.

## Core Concept

Every PlainTest command has two phases:
- **Setup phase**: Runs once to prepare environment
- **Test phase**: Runs multiple times with data

```bash
./plaintest run --setup "auth.Login" --test "api tests" -d users.csv
#               ↑ Runs once          ↑ Runs for each CSV row
```

## What's a Link?

A link specifies what to run from a collection. Write links using dots.

```bash
auth                        # whole collection
"auth".Login               # one request
"api tests"."User Folder"  # one folder
```

Combine multiple items with commas:
```bash
"api tests"."Create User,Update User,Delete User"
```

## Setup Phase

Setup links run once, regardless of CSV data.

```bash
./plaintest run --setup "db.Init" --setup "auth.Login" --test "api tests" -d data.csv
#               ↑ Runs once       ↑ Runs once          ↑ Iterates
```

Each setup link runs as its own Newman process. Environment variables flow to the next link.

Common setup tasks:
- Database initialization
- Authentication
- Test data creation
- Service health checks

## Test Phase

Test links iterate when CSV data is provided.

```bash
./plaintest run --test "api tests.User Tests" --test "api tests.Product Tests" -d data.csv
#               ↑ Runs for each row           ↑ Runs for each row
```

Without CSV data, test links run once:
```bash
./plaintest run --test "smoke tests"
#               ↑ Runs once (no data file)
```

## How PlainTest Executes

Each link becomes a separate Newman command:

```bash
./plaintest run --setup "auth.Login" --test "api tests.User Tests" -d users.csv
```

Executes as:
```bash
# Step 1: Setup (no CSV)
newman run auth.json --folder "Login" --export-environment temp_env.json

# Step 2: Test (with CSV)
newman run api_tests.json --folder "User Tests" -d users.csv -e temp_env.json
```

## Execution Order

1. All setup links run first (left to right)
2. All test links run after setup (left to right)
3. Environment variables flow between all links

```bash
./plaintest run --test "products" --setup "auth" --test "users" --setup "db"
#               Actual order: db → auth → products → users
```

Setup always runs before tests, regardless of command order.

## Item Selection

Newman's --folder flag accepts folder names or request names. PlainTest passes these directly.

```bash
"api tests"."User Folder,Create Request,Delete Request"
```

Becomes:
```bash
newman run api_tests.json --folder "User Folder" --folder "Create Request" --folder "Delete Request"
```

Items run in collection order, not selection order. Newman treats --folder as a filter.

## Rules

1. Every link runs as a separate Newman process
2. Setup links never iterate with CSV
3. Test links iterate when -d is present
4. Same collection can appear in multiple links
5. Items within a link run in collection order

## Examples

Basic authentication setup:
```bash
./plaintest run --setup "auth.Login" --test "api tests"
```

Multiple setup steps:
```bash
./plaintest run --setup "db.Init" --setup "auth.Login" --setup "cache.Warm" \
                --test "api tests"
```

Data-driven testing:
```bash
./plaintest run --setup "auth.Login" \
                --test "api tests.Create User,Verify User" \
                -d users.csv -r 1-10
```

Multiple test collections:
```bash
./plaintest run --setup "auth" \
                --test "user tests" --test "product tests" --test "order tests" \
                -d test_data.csv
```

Selective testing:
```bash
./plaintest run --test "api tests"."Health Check,Status Endpoint"
```

## What You Can't Do

Mix setup and test for the same collection in one link:
```bash
# Can't specify some items as setup and others as test in one link
# Split into separate links instead
```

Control execution order within a link:
```bash
# Items run in collection order, not selection order
"api tests"."Cleanup,Setup,Test"  # Runs as: Setup,Test,Cleanup
```

## Why This Design?

Clear separation of concerns. Setup prepares. Tests validate.

No confusion about what iterates and what doesn't. Setup phase = once. Test phase = iteration.

Each link as a separate Newman process keeps environment handling simple. PlainTest manages the flow between processes.

## Summary

PlainTest uses --setup and --test to separate one-time preparation from iterative testing. Each link runs as its own Newman process. Simple semantics, predictable behavior.
