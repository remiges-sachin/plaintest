# Script Pull and Push Workflow

## Goal
Edit Postman scripts in your favorite editor while keeping collections in sync.

## Simple Two-Command Workflow

### 1. Pull Scripts (One Time)
```bash
plaintest scripts pull my-api
```
- Reads from `collections/my-api.postman_collection.json`
- Creates editable JS files in `scripts/my-api/`
- **Always overwrites** existing script files

### 2. Build Collection (After Editing)
```bash
plaintest scripts push my-api
```
- Reads scripts from `scripts/my-api/`
- Updates `collections/my-api.postman_collection.json` with your changes
- **Scripts are the source of truth**

## Directory Structure

```
plaintest-project/
├── collections/
│   ├── my-api.postman_collection.json    # Original from Postman
│   └── smoke.postman_collection.json
└── scripts/
    ├── my-api/
    │   ├── _collection__prerequest.js    # Collection-level scripts
    │   ├── _collection__test.js
    │   ├── get-user__prerequest.js       # Request scripts
    │   ├── get-user__test.js
    │   └── create-user__test.js
    └── smoke/
        └── health-check__test.js
```

## Script File Names

Collection scripts start with underscore: `_collection__prerequest.js`

Request scripts use the request name: `get-users__test.js`

Why the underscore? Prevents collision if you name a request "collection".

## Typical Workflow

```bash
# 1. Export from Postman → collections/my-api.postman_collection.json

# 2. Extract scripts (one time setup)
plaintest scripts pull my-api
# ✓ Extracted: _collection__prerequest.js
# ✓ Extracted: get-user__test.js
# ✓ Extracted: create-user__test.js
# Extraction complete

# 3. Edit scripts in VS Code, WebStorm, etc.
code scripts/my-api/get-user__test.js

# 4. Build collection with your changes
plaintest scripts push my-api
# ✓ Injected: _collection__prerequest.js
# ✓ Injected: get-user__test.js
# ✓ Injected: create-user__test.js
# Collection updated successfully

# 5. Run tests with updated collection
plaintest run my-api
```

## Key Benefits

- **Simple mental model**: Extract once, then edit scripts freely
- **No complex conflict resolution**: Scripts always win after extraction
- **Version control friendly**: Clear separation of scripts and collections
- **Editor support**: Full IDE features for JavaScript editing

## Important Notes

- **Extract overwrites**: Running extract again will overwrite your script files
- **Scripts are authoritative**: After extraction, scripts become the source of truth
- **Use version control**: Keep both scripts and collections in git
- **Collection structure**: Don't change collection structure in Postman after extraction

## Script File Naming

- **Collection scripts**: `_collection__prerequest.js`, `_collection__test.js`
- **Request scripts**: `{request-name}__{event-type}.js`
- **Nested items**: Follows directory structure

## Troubleshooting

**Q: I accidentally overwrote my scripts with extract**
A: Use `git checkout` to restore your script files

**Q: Build says "script file not found"**
A: Make sure all script files exist, or re-extract if collection structure changed

**Q: Want to update collection structure?**
A: Update in Postman, export new collection, then extract again

**Q: Scripts not working in Newman?**
A: Run `plaintest scripts push` first to update the collection
