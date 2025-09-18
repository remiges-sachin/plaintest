# PlainTest

A CLI framework for API testing that acts as a Newman proxy with CSV-driven test data, following PlainTest methodology.

## Features

### Implemented
- **Newman Proxy**: Pass-through any Newman flags while adding PlainTest functionality
- **Auto-Discovery for All Resources**: Collections, environments, and data files can be referenced by name instead of full paths
- **Single Environment Auto-Detection**: Automatically uses environment if only one exists in the project
- **Collection Chaining with Environment Sharing**: Run multiple collections in sequence with automatic environment variable sharing: `plaintest run get_auth api_tests`
- **CSV Row Selection**: Run specific test subsets with `-r 2`, `-r 2-5`, or `-r 1,3,5` patterns
- **Project Initialization**: `plaintest init` creates working templates with DummyJSON API examples
- **CSV-Driven Testing**: Data-driven tests using PlainTest three-column structure (META → INPUT → EXPECTED)
- **Newman Compatibility**: Use any Newman flag: `--verbose`, `--timeout`, `--bail`, etc.
- **Authentication Templates**: Working examples with DummyJSON login flow
- **Script Push/Pull**: `plaintest scripts pull` and `plaintest scripts push` edit Postman scripts in standalone `.js` files
- **Report Generation**: `--reports` generates timestamped JSON and HTML report files with request/response data

### To Be Implemented
- **Report Summary**: Test execution summaries and failure analysis

## Quick Start

```bash
# Install dependencies
npm install -g newman newman-reporter-htmlextra

# Build PlainTest
make build
# or
go build -o plaintest ./cmd/plaintest

# Initialize new project with working templates
./plaintest init

# List project resources
./plaintest list collections
./plaintest list data
./plaintest list environments
./plaintest list scripts

# Pull scripts from collections for editing
./plaintest scripts pull my-api

# Push edited scripts back to collection
./plaintest scripts push my-api

# Run smoke tests (auto-discovered)
./plaintest run smoke

# See all available collections
./plaintest run --help

# Run authentication then tests (collection chaining with environment sharing)
./plaintest run get_auth api_tests

# Run auth collection once, then test collections with CSV iterations (--once flag)
./plaintest run get_auth api_tests -d data/example.csv --once get_auth

# Run CSV-driven tests with row selection (PlainTest-specific)
./plaintest run api_tests -d data/example.csv -r 1-3

# Use any Newman flags (proxy mode)
./plaintest run smoke --verbose --timeout 10000

# Capture request/response data
./plaintest run api_tests --reports

# Mix PlainTest and Newman features (names supported: -e production -d example)
./plaintest run get_auth api_tests -d data/example.csv -r 2 --bail --verbose
```

## Documentation

- **[USER_DOCS.md](USER_DOCS.md)** - Command reference and usage guide
- **[SCRIPT_SYNC.md](SCRIPT_SYNC.md)** - Details for the script sync workflow
- **Reference implementation**: See `reference/cvl-kra/PlainTest-README.md` for a full CVL PAN validation example
- **Command reference**: Run `./plaintest --help` for command options

## PlainTest Methodology

PlainTest follows a structured approach to API testing:

### CSV Structure
Three-column pattern for test data:
- **META columns** (`test_*`): Test identification and metadata
- **INPUT columns** (`input_*`): Request parameters and data
- **EXPECTED columns** (`expected_*`): Expected response values

### Test Types
- **Smoke Tests**: Maximum 5 critical path tests, 30-second timeout
- **Full Tests**: Complete test suite with CSV data iteration
- **Authentication Flow**: Two-phase execution (auth → tests)

## Architecture

```
cmd/plaintest/          # CLI entry point and command definitions
internal/
├── core/              # Version and core utilities
├── newman/            # Newman service wrapper
├── csv/               # CSV processing and row selection
├── scriptsync/        # Raw↔Postman script↔build sync utilities
└── templates/         # Project template generation
```

## Development

```bash
# Run tests
go test ./...

# Run with coverage
make coverage-check

# Build binary
go build -o plaintest ./cmd/plaintest

# Install pre-commit hooks
make pre-commit-install
```

## Version

Current version: 0.0.1-dev
