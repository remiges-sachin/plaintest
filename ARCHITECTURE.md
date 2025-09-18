# PlainTest Architecture

## Overview

PlainTest is a Newman proxy that adds collection chaining, CSV row selection, and request/response capture while maintaining Newman compatibility.

## Design Philosophy

**Newman Proxy Approach**: Rather than wrapping Newman's functionality, PlainTest acts as a transparent proxy that:
1. **Passes through Newman flags** - Users get Newman functionality
2. **Adds three specific features** - Collection chaining, CSV row selection, and request/response capture
3. **Auto-discovers collections** - No hard-coded collection mappings

## Project Structure

```
plaintest/
├── cmd/plaintest/           # CLI entry point
│   ├── main.go             # Newman proxy with collection discovery
│   └── main_test.go        # CLI integration tests
├── internal/
│   ├── core/               # Core functionality
│   │   └── version.go      # Version information
│   ├── newman/             # Newman service wrapper
│   │   ├── service.go      # Newman subprocess execution with flags
│   │   └── service_test.go # Newman service tests
│   ├── csv/                # CSV processing
│   │   ├── processor.go    # Row selection and filtering
│   │   └── processor_test.go # CSV processing tests
│   ├── scriptsync/         # Collection script extract and build logic
│   └── templates/          # Project templates (embedded in Go code)
│       └── templates.go    # Template generation for init
├── .pre-commit-config.yaml # Code quality hooks
├── Makefile               # Build and test automation
└── go.mod                 # Go module definition
```

## Core Components

### 1. Newman Proxy (`cmd/plaintest/main.go`)

**Role**: Intelligent command-line proxy between user and Newman

**Key Functions**:
- `discoverCollections()` - Auto-discover `*.postman_collection.json` files
- `parseArguments()` - Separate collection names from Newman flags
- `buildRunCommandLong()` - Dynamic help text with available collections

**Proxy Logic**:
```go
// 1. Discover collections from filesystem
collectionMap := discoverCollections()

// 2. Parse user input: separate collections from Newman flags
collections, newmanFlags := parseArguments(rawArgs, collectionMap)

// 3. For each collection, run Newman with processed flags
for _, collection := range collections {
    service.RunWithFlags(collectionPath, finalNewmanFlags)
}
```

### 2. Collection Discovery

**Auto-Discovery Algorithm**:
1. Scan `collections/*.postman_collection.json` files
2. Extract collection name from filename (remove `.postman_collection.json`)
3. Map collection names from filenames
4. Build collection map: `name → file_path`

**Result**:
- No hard-coded collection names
- Users can add collections without code changes
- Follows Postman export conventions
- Maintains backward compatibility

### 3. Newman Service (`internal/newman/service.go`)

**Three Execution Methods**:

1. **Legacy**: `Run(collection, options)` - Structured options (backward compatibility)
2. **Proxy**: `RunWithFlags(collection, flags)` - Direct flag pass-through
3. **Environment Export**: `RunWithEnvironmentExport(collection, flags, exportPath)` - Flag pass-through with environment export

**Environment Export Implementation**:
```go
func (s *Service) RunWithEnvironmentExport(collection string, flags []string, exportEnvPath string) (*Result, error) {
    args := []string{"run", collection}
    args = append(args, flags...)
    args = append(args, "--export-environment", exportEnvPath)  // Add environment export

    cmd := exec.Command(s.executable, args...)
    output, err := cmd.CombinedOutput()

    return &Result{
        Success:  err == nil,
        ExitCode: cmd.ProcessState.ExitCode(),
        Output:   string(output),
    }, err
}
```

### 4. CSV Processing (`internal/csv/processor.go`)

**PlainTest-Specific Feature**: Row selection from CSV data

**Integration with Newman**:
1. Extract CSV file from Newman flags (`-d`, `--iteration-data`)
2. Process row selection if `-r` flag specified
3. Create temporary filtered CSV file
4. Replace CSV path in Newman flags with filtered file

**Row Selection Patterns**:
- Single: `-r 2` → row 2 only
- Range: `-r 2-5` → rows 2 through 5
- List: `-r 1,3,5` → rows 1, 3, and 5

### 5. Script Extract and Build (`internal/scriptsync/service.go`)

**Purpose**: Extract Postman scripts to editable files and build collections from edited scripts.

**Commands**:
- `plaintest scripts pull collection-name` - Pull scripts from collection to editable `.js` files
- `plaintest scripts push collection-name` - Push updated scripts from `.js` files to collection

**Key Details**:
- Extract always overwrites script files with current collection content
- Build updates collection in-place with script content
- Scripts become the source of truth after extraction

## Command Flow

### Basic Execution
```
User: ./plaintest run smoke --verbose
  ↓
1. Discover collections → {smoke: "collections/smoke.postman_collection.json"}
2. Parse arguments → collections=[smoke], flags=[--verbose]
3. Add default environment → flags=[--verbose, -e, environments/dummyjson.postman_environment.json]
4. Execute: newman run collections/smoke.postman_collection.json --verbose -e environments/dummyjson.postman_environment.json
```

### Collection Chaining with Environment Sharing + CSV Row Selection
```
User: ./plaintest run get_auth api_tests -d data/example.csv -r 2-5 --bail
  ↓
1. Discover collections → {get_auth: "...", api_tests: "..."}
2. Parse arguments → collections=[get_auth, api_tests], flags=[-d, data/example.csv, --bail]
3. For get_auth collection:
   - Create temporary environment file for sharing
   - Execute: newman run collections/get_auth.postman_collection.json -d data/example.csv --bail -e ... --export-environment /tmp/plaintest_env_xxx.json
   - Newman merges base environment with script-generated variables (auth tokens)
4. For api_tests collection:
   - Use merged environment from previous collection (base + auth tokens)
   - Process row selection: create temp file with rows 2-5
   - Replace CSV path: [-d, /tmp/filtered.csv, --bail]
   - Replace environment: [-e, /tmp/plaintest_env_xxx.json, --bail]
   - Execute: newman run collections/api_tests.postman_collection.json -d /tmp/filtered.csv -e /tmp/plaintest_env_xxx.json --bail
5. Cleanup temporary files

Environment Evolution:
- Start: base_environment.json
- After get_auth: base_environment.json + {authToken: "xyz", userId: 123}
- api_tests receives: merged environment with all variables
```

## Design Decisions

### 1. Proxy vs Wrapper
**Decision**: Proxy approach
**Rationale**:
- Users get Newman functionality
- No need to reimplement Newman flags
- Automatic compatibility with new Newman features
- Focused codebase for PlainTest value-add

### 2. Auto-Discovery vs Hard-Coded
**Decision**: Auto-discovery with backward compatibility
**Rationale**:
- Users can add collections without code changes
- Follows Postman conventions
- Maintains existing user workflows
- Self-documenting help text

### 3. Flag Parsing Strategy
**Decision**: Raw argument parsing with collection/flag separation
**Rationale**:
- Cobra's flag parsing interferes with Newman flag pass-through
- Raw parsing gives control over argument processing
- Allows mixing PlainTest flags (-r) with Newman flags (--verbose)

## Newman Compatibility

**Compatibility**: Newman flags work unchanged:
```bash
./plaintest run api_tests --verbose --timeout 30000 --bail --reporters cli,htmlextra --reporter-htmlextra-export reports/
```

**PlainTest Additions**: Three features layered on top:
1. **Collection chaining with environment sharing**: `./plaintest run get_auth api_tests`
2. **CSV row selection**: `-r 2-5`
3. **Report generation**: `--reports` (generates timestamped JSON and HTML files)

## Future Extensibility

**Adding New Features**:
1. **PlainTest-specific**: Add to `parseArguments()` flag filtering
2. **Newman pass-through**: Requires no changes - automatic compatibility

**Collection Management**:
- Add `.postman_collection.json` files to `collections/` directory
- Auto-discovered and immediately available
- No code changes required

## Error Handling

**Collection Discovery Errors**:
- Missing collections directory → Warning, continue with empty map
- Invalid collection names → Error message with available options
- File access issues → Fallback behavior

**Newman Execution Errors**:
- Newman not installed → Installation instructions
- Collection file issues → Show Newman's native error messages
- Flag conflicts → Let Newman handle validation

This architecture provides a separation between PlainTest's value-add features and Newman's core functionality, while maintaining compatibility and extensibility.
