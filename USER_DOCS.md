# PlainTest User Documentation

## Overview

PlainTest is a Newman proxy that adds CSV-driven testing and collection chaining.

## Prerequisites

Install Newman and reporter:

```bash
npm install -g newman newman-reporter-htmlextra
```

## Installation

Build PlainTest from source:

```bash
go build -o plaintest ./cmd/plaintest
```

## Commands Reference

### `plaintest version`

Shows the current version.

```bash
./plaintest version
```

### `plaintest init`

Creates project structure with templates:

```bash
./plaintest init
```

Creates:
- `collections/raw/` - Store Postman exports
- `collections/build/` - Generated collections used for execution
- `scripts/` - Extracted editable Postman scripts
- `data/` - CSV test data files
- `environments/` - Environment configuration files
- `reports/` - Test execution reports directory

Templates include examples using DummyJSON API.

### `plaintest list [resource-type]`

Lists available project resources:

```bash
# List all available collections
./plaintest list collections

# List all CSV data files
./plaintest list data

# List all environment files
./plaintest list environments

# List all extracted script directories
./plaintest list scripts
```

**Resource Types:**
- `collections` - Shows Postman collections in collections/ directory with their file paths
- `data` - Shows CSV files in data/ directory for test data
- `environments` - Shows environment files in environments/ directory
- `scripts` - Shows extracted script directories in scripts/ with file counts

**Output Example:**
```
$ ./plaintest list collections
Available collections:
  smoke (collections/smoke.postman_collection.json)
  api_tests (collections/api_tests.postman_collection.json)
  get_auth (collections/get_auth.postman_collection.json)

$ ./plaintest list scripts
Available script directories:
  my-api (3 script files)
  smoke-tests (1 script files)
```

This shows what resources are available in your project.

### `plaintest scripts pull [collection-name]`

Extracts scripts from a Postman collection to JavaScript files:

```bash
./plaintest scripts pull my-api
```

Workflow:
- Reads collection from `collections/my-api.postman_collection.json`
- Extracts all scripts to `scripts/my-api/` directory
- Always overwrites existing script files

Files you'll see after pulling:
- `_collection__*.js` - runs before all requests
- `request-name__*.js` - runs for one request

The underscore keeps collection scripts separate from request scripts.

### `plaintest scripts push [collection-name]`

Builds a collection with updated scripts from JavaScript files:

```bash
./plaintest scripts push my-api
```

Workflow:
- Reads scripts from `scripts/my-api/`
- Updates `collections/my-api.postman_collection.json` with script content
- Scripts are the source of truth after extraction

Extract once, then edit scripts and build.

### `plaintest run [collections...] [newman-flags...]`

Execute API tests using Newman as a proxy. PlainTest adds two features:

1. **Collection chaining with environment sharing** - Run multiple collections in sequence with environment variable sharing
2. **CSV row selection** - Filter CSV data to specific rows

#### Auto-Discovery System

PlainTest discovers resources in your project directories and lets you reference them by name:

```bash
# See collections, environments, and data files
./plaintest run --help
```

**Collections** (from `collections/` directory):
- `smoke` → `collections/smoke.postman_collection.json`
- `api_tests` → `collections/api_tests.postman_collection.json`
- `get_auth` → `collections/get_auth.postman_collection.json`

**Environments** (from `environments/` directory):
- `production` → `environments/production.postman_environment.json`
- `staging` → `environments/staging.postman_environment.json`
- `AWS-QA` → `environments/AWS-QA.postman_environment.json`

**Data Files** (from `data/` directory):
- `example` → `data/example.csv`
- `add_ucc` → `data/add_ucc.csv`
- `test_cases` → `data/test_cases.csv`

**Defaults**: If only one environment file exists, PlainTest uses it without requiring the `-e` flag.

#### PlainTest-Specific Flags

- `-r, --rows string` - CSV row selection (PlainTest feature)
  - Single row: `-r 2`
  - Range: `-r 2-5`
  - List: `-r 1,3,5`
  - Note: pass the value as a separate argument (use `-r 2`, not `-r2`)

- `--reports` - Generate timestamped HTML and JSON report files (PlainTest feature)
  - Creates timestamped JSON and HTML files in `reports/` directory
  - JSON file: machine-readable format with request/response data
  - HTML file: human-readable format with request/response details
  - Files named: `[collection]_YYYYMMDDTHHMMSS.json` and `[collection]_YYYYMMDDTHHMMSS.html`

- `--debug` - Print the Newman command before running (PlainTest feature)
  - Shows the exact Newman command with all flags for troubleshooting
  - Useful for understanding collection chaining and flag processing

- `--setup "collection1.item1,item2"` - Setup links that run once without CSV iteration (PlainTest feature)
  - Setup phase runs before tests, regardless of command order
  - Use dot notation for specific requests: `"collection"."item1,item2"`
  - Example: `plaintest run --setup "get_auth.Login" --test "api_tests" -d data.csv`

- `--test "collection1.item1,item2"` - Test links that iterate with CSV data (PlainTest feature)
  - Test phase runs after setup, iterating over CSV rows
  - Use dot notation for specific requests: `"collection"."item1,item2"`
  - Example: `plaintest run --setup get_auth --test "api_tests.Create User,Update User" -d data.csv`

#### Newman Flags (Passed Through)

Other flags are passed directly to Newman:

- `-e, --environment` - Environment file (supports names: `-e production` or paths: `-e environments/production.postman_environment.json`)
- `-d, --iteration-data` - CSV data file (supports names: `-d example` or paths: `-d data/example.csv`)
- `--verbose` - Verbose output with request/response details
- `--timeout` - Request timeout in milliseconds
- `--bail` - Stop on first failure
- `--reporters` - Specify reporters
- `--export-*` - Export options
- And many more - see `newman run --help`

#### Setup-Test Phase Execution

PlainTest separates setup from tests using `--setup` and `--test` flags. Setup runs once, tests iterate with CSV data.

**Setup Phase**: Runs once regardless of CSV data
```bash
# Setup runs once, even with CSV data
plaintest run --setup get_auth --test api_tests -d data.csv
```

**Test Phase**: Iterates with CSV data when present
```bash
# Tests iterate over each CSV row
plaintest run --test api_tests -d data.csv
```

**Combined Setup-Test**:
```bash
# Multiple setup links, multiple test links
plaintest run --setup db_setup --setup get_auth --test api_tests --test integration_tests -d data.csv
```

**Dot Notation for Granular Control**:
```bash
# Run specific requests from collections
plaintest run --setup "get_auth.Login" --test "api_tests.Create User,Update User" -d data.csv
```

#### Examples

**Basic collection execution:**
```bash
./plaintest run smoke
./plaintest run api_tests
```

**Setup-test execution with environment sharing (PlainTest feature):**
```bash
./plaintest run --setup get_auth --test api_tests
./plaintest run --setup "get_auth.Login" --test "api_tests.Create User,Update User"
```

Collection chaining automatically shares environment variables between collections using environment merging:

1. **Base environment preserved**: First collection starts with your specified environment file
2. **Environment enhanced**: Newman adds auth tokens and script-modified variables to the base environment
3. **Merged environment shared**: Second collection receives base environment + additions from first collection
4. **Chain continues**: Each collection builds on the accumulated environment state

**Example flow:**
```bash
./plaintest run --setup get_auth --test api_tests -e production
# Setup phase:  Uses production.json + adds authToken
# Test phase: Uses production.json + authToken (merged environment)
```

This enables authentication workflows where login tokens from the first collection automatically flow to subsequent API test collections.

**CSV data with Newman flags:**
```bash
./plaintest run api_tests -d data/example.csv
./plaintest run api_tests -e environments/localhost.postman_environment.json -d data/example.csv
```

**CSV row selection (PlainTest feature):**
```bash
./plaintest run api_tests -d data/example.csv -r 1-3
./plaintest run api_tests -d data/example.csv -r 1,3,5
```

**Newman verbose output:**
```bash
./plaintest run smoke --verbose
./plaintest run api_tests -d data/example.csv --verbose
```

**Newman timeout and bail:**
```bash
./plaintest run smoke --timeout 30000 --bail
```

**Complex combinations (using name-based syntax):**
```bash
./plaintest run --setup get_auth --test api_tests -e production -d example -r 2-5 --verbose --bail --timeout 10000
```

**Capture requests and responses:**
```bash
# Basic capture
./plaintest run smoke --reports

# Capture with CSV data
./plaintest run api_tests -d example --reports

# Capture with setup-test execution
./plaintest run --setup get_auth --test api_tests --reports
```

**Export reports:**
```bash
./plaintest run api_tests -d data/example.csv --reporters cli,htmlextra --reporter-htmlextra-export reports/ # pragma: allowlist secret
```

### `plaintest help`

Shows usage information:

```bash
./plaintest --help
./plaintest run --help
```

## Usage Workflow

1. **Initialize project**:
   ```bash
   ./plaintest init
   ```

2. **See available collections**:
   ```bash
   ./plaintest run --help
   ```

3. **Run smoke tests** (works immediately):
   ```bash
   ./plaintest run smoke
   ```

4. **Run with verbose output**:
   ```bash
   ./plaintest run smoke --verbose
   ```

5. **Run authentication workflow** (collection chaining):
   ```bash
   ./plaintest run get_auth api_tests
   ```

6. **Run data-driven tests**:
   ```bash
   ./plaintest run api_tests -d data/example.csv
   ```

7. **Add your own resources**:
   - Add `*.postman_collection.json` files to `collections/`
   - Add `*.postman_environment.json` files to `environments/`
   - Add `*.csv` files to `data/`
   - They'll be auto-discovered and available by name immediately

## Auto-Discovery Reference

PlainTest automatically discovers resources in your project directories:

### Discovery Patterns and Name Mapping

**Collections** (searched in order):
1. `collections/build/*.postman_collection.json` → name (e.g., `my_tests`)
2. `collections/*.postman_collection.json` → name (e.g., `my_tests`)

**Environments**: `environments/*.postman_environment.json` → name (e.g., `production`)
**Data Files**: `data/*.csv` → name (e.g., `test_cases`)

### Usage Examples

```bash
# Use names instead of paths (recommended)
./plaintest run my_api_tests -e production -d test_cases

# Full paths still work (backward compatible)
./plaintest run collections/my_api_tests.postman_collection.json \
  -e environments/production.postman_environment.json \
  -d data/test_cases.csv
```

### Single Environment Auto-Detection

If only one environment file exists in `environments/`, PlainTest automatically uses it:

```bash
# If only production.postman_environment.json exists
./plaintest run api_tests -d example
# Equivalent to: ./plaintest run api_tests -e production -d example
```

## CSV Data Format

PlainTest uses three-column CSV structure:

### Column Types
- **META columns** (`test_*`): Test identification and metadata
- **INPUT columns** (`input_*`): Request parameters and data
- **EXPECTED columns** (`expected_*`): Expected response values

### Example CSV
```csv
test_id,test_name,test_description,input_endpoint,input_method,expected_status,expected_username
TC_001,Valid Auth User,Get current authenticated user,/auth/me,GET,200,emmaj
TC_002,Auth Required,Get user profile via auth,/auth/me,GET,200,emmaj
```

### Row Selection (PlainTest Feature)
- Single row: `-r 2` (run only row 2)
- Range: `-r 2-5` (run rows 2 through 5)
- List: `-r 1,3,5` (run rows 1, 3, and 5)

## Newman Proxy Usage

### Newman Compatibility
Use any Newman feature:
```bash
./plaintest run api_tests --verbose --timeout 30000 --bail --reporters cli,json
```

### PlainTest Value-Add
Get setup-test execution with environment sharing and CSV row selection:
```bash
./plaintest run --setup get_auth --test api_tests -d data/example.csv -r 1-3
```

Environment variables (including authentication tokens) from the first collection are automatically shared with subsequent collections.

### Auto-Discovery
Add collections without touching code:
```bash
# Add file: collections/integration_tests.postman_collection.json
./plaintest run integration_tests  # Works immediately!
```

### Capturing Request/Response Logs
Use Newman reporters to persist request and response payloads:

```bash
# JSON report (machine-readable)
./plaintest run api_tests -d data/example.csv \
  --reporters cli,json \
  --reporter-json-export reports/api_tests.json

# HTML Extra report (human-readable)
./plaintest run api_tests -d data/example.csv \
  --reporters cli,htmlextra \
  --reporter-htmlextra-export reports/api_tests.html
```

Exports are written relative to the current working directory unless you provide absolute paths.

## Troubleshooting

### Newman not found
Install Newman globally:
```bash
npm install -g newman newman-reporter-htmlextra
```

### Collection not found
- Check file is in `collections/` directory
- Ensure filename ends with `.postman_collection.json`
- Run `./plaintest run --help` to see available collections

### CSV row selection not working
- Ensure you have `-d` flag with CSV file
- Check CSV file exists and has header row
- Row numbers start from 1 (header is not counted)

### Newman flags not working
Newman flags are supported. Check Newman documentation:
```bash
newman run --help
```

## Working with Captured Data

The `--reports` flag generates timestamped JSON and HTML files containing request/response data:

### Usage
```bash
# View request/response data
jq '.run.executions[0] | {method: .request.method, url: .request.url, status: .response.code}' reports/collection_*.json

# Convert response body from Buffer to text
jq -r '.run.executions[0].response.stream.data | implode' reports/collection_*.json

# Find failed requests
jq '.run.executions[] | select(.response.code >= 400)' reports/collection_*.json
```

HTML files can be opened in a browser for inspection.

## Examples by Use Case

### API Health Monitoring
```bash
./plaintest run smoke --timeout 5000
```

### Development Testing
```bash
./plaintest run --setup get_auth --test api_tests -d data/dev_data.csv --verbose
```

### CI/CD Pipeline
```bash
./plaintest run --test smoke --test api_tests --bail --timeout 30000 --reporters cli,json
```

### Manual Testing
```bash
./plaintest run api_tests -d data/manual_test.csv -r 5-10 --verbose
```
