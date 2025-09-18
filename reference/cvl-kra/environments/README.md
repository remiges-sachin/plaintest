# Environment Files Setup

## Security Notice
Environment files containing passwords are excluded from version control to protect credentials.

## Setup Instructions

### 1. Create Environment Files
```bash
# Copy templates to create your environment files
cp localhost.postman_environment.json.template localhost.postman_environment.json
cp uat.postman_environment.json.template uat.postman_environment.json
```

### 2. Password Management
- The test script automatically fetches passwords using the Get Password API
- No manual password entry required
- Environment files remain local and secure

### 3. File Structure
- `*.postman_environment.json.template` - Template files (tracked in git)
- `*.postman_environment.json` - Actual files (excluded from git)

## Configuration Details

### Localhost Environment
- **Protocol**: HTTP
- **Base URL**: localhost:8083
- **Password**: Auto-fetched via API

### UAT Environment
- **Protocol**: HTTPS
- **Base URL**: uat.cvlkra.remiges.tech/api
- **Password**: Auto-fetched via API

## Troubleshooting

If environment files are missing:
1. Run the setup commands above
2. Verify templates exist in the environments directory
3. Check that `.gitignore` excludes `*.postman_environment.json`
