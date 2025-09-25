# Script Pull and Push Workflow

Edit Postman scripts in your favorite editor while keeping collections in sync.

## The Problem

Postman stores scripts as strings in JSON. You can't edit strings in your IDE.

## The Solution

Extract scripts to JavaScript files. Edit them properly. Push changes back.

## Workflow

### 1. Pull Scripts (One Time)
```bash
plaintest scripts pull my-api
```

Extracts scripts from `collections/my-api.postman_collection.json` to `scripts/my-api/` directory.

### 2. Edit Scripts
Scripts become JavaScript files:

```
scripts/my-api/
├── _collection__prerequest.js    # Collection-level scripts
├── _collection__test.js
├── get-user__prerequest.js       # Request scripts
├── get-user__test.js
└── create-user__test.js
```

Edit in your IDE with syntax highlighting, linting, debugging.

### 3. Push Changes
```bash
plaintest scripts push my-api
```

Updates `collections/my-api.postman_collection.json` with your script changes.

## Key Rules

- **Pull once, then edit scripts freely**
- **Scripts are the source of truth after extraction**
- **Don't edit in Postman after pulling**

## Benefits

- Edit in VS Code, WebStorm, etc.
- Version control shows real code changes
- Use linters and formatters
- Debug with proper tools

## File Naming

- Collection scripts start with `_collection__`
- Request scripts use request name: `get-users__test.js`
- Underscore prevents collisions

See [CLI_REFERENCE.md](CLI_REFERENCE.md) for complete script command documentation.
