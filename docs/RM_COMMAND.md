# aigogo rm Command

## Overview

The `rm` command removes files or dependencies from your `aigogo.json` manifest file. It has three subcommands: `file`, `dep`, and `dev`.

**Subdirectory Support**: Like `git`, this command works from any subdirectory. aigogo searches up the tree to find `aigogo.json`.

## Subcommands

### `rm file` - Remove Files from Package

Removes files or patterns from the `files.include` array.

**Usage:**
```bash
aigogo rm file <path>...
```

**Features:**
- Removes specified files from include list
- Supports removing multiple files at once
- Interactive mode shows current files
- Shows remaining files after removal

**Examples:**
```bash
# Remove single file
aigogo rm file old_utils.py

# Remove multiple files
aigogo rm file test.py helper.py config.json

# Remove glob pattern
aigogo rm file "*.pyc"
```

**Interactive Mode:**
```bash
$ aigogo rm file
Current files in include list:
  1. api_client.py
  2. utils.py
  3. helpers.py

File path(s) to remove (space-separated): utils.py helpers.py
✓ Removed 2 file(s) from include list:
  - utils.py
  - helpers.py

Remaining files (1):
  - api_client.py
```

**Error Handling:**
```bash
# File not in include list
$ aigogo rm file nonexistent.py
Error: no matching files found in include list

# files.include is set to "auto"
$ aigogo rm file test.py
Error: files.include is set to 'auto'
Cannot remove individual files when using auto-discovery

# No files in include list
$ aigogo rm file test.py
Error: no files in include list
```

---

### `rm dep` - Remove Runtime Dependency

Removes a runtime dependency from `dependencies.runtime`.

**Usage:**
```bash
aigogo rm dep <package>
```

**Examples:**
```bash
# Python
aigogo rm dep requests

# JavaScript
aigogo rm dep axios

# Go
aigogo rm dep github.com/pkg/errors

# Rust
aigogo rm dep serde
```

**Interactive Mode:**
```bash
$ aigogo rm dep
Current runtime dependencies:
  1. requests (>=2.31.0,<3.0.0)
  2. click (>=8.0.0,<9.0.0)
  3. pyyaml (>=6.0)

Package name to remove: requests
✓ Removed requests (>=2.31.0,<3.0.0) from runtime dependencies

Remaining runtime dependencies (2):
  - click (>=8.0.0,<9.0.0)
  - pyyaml (>=6.0)
```

---

### `rm dev` - Remove Development Dependency

Removes a development dependency from `dependencies.dev`.

**Usage:**
```bash
aigogo rm dev <package>
```

**Examples:**
```bash
# Python testing
aigogo rm dev pytest
aigogo rm dev black

# JavaScript testing
aigogo rm dev jest

# Go testing
aigogo rm dev github.com/stretchr/testify
```

**Interactive Mode:**
```bash
$ aigogo rm dev
Current development dependencies:
  1. pytest (^7.0.0)
  2. black (>=23.0.0)

Package name to remove: pytest
✓ Removed pytest (^7.0.0) from development dependencies

Remaining development dependencies (1):
  - black (>=23.0.0)
```

---

## Complete Workflow Examples

### Removing Unused Files

```bash
# Show what files are included
cat aigogo.json

# Remove old files
aigogo rm file old_api.py deprecated.py

# Verify changes
aigogo validate
```

### Cleaning Up Dependencies

```bash
# Remove unused runtime dependency
aigogo rm dep requests

# Remove unused dev dependency
aigogo rm dev pytest

# Validate
aigogo validate
```

### Interactive Cleanup

```bash
# See all dependencies and remove interactively
aigogo rm dep
# Shows list, prompts for package name

aigogo rm dev
# Shows dev deps, prompts for package name
```

---

## Behavior Details

### Removing Files

Before:
```json
{
  "files": {
    "include": ["utils.py", "api.py", "config.json"]
  }
}
```

After `aigogo rm file utils.py`:
```json
{
  "files": {
    "include": ["api.py", "config.json"]
  }
}
```

### Empty Include List

If all files are removed:
```bash
$ aigogo rm file api.py config.json
✓ Removed 2 file(s) from include list:
  - api.py
  - config.json

No files remaining in include list
```

Result:
```json
{
  "files": {
    "include": []
  }
}
```

### Automatic Cleanup

If both `runtime` and `dev` dependencies are empty, the entire `dependencies` section is removed:

```bash
# Last runtime dep
$ aigogo rm dep requests
✓ Removed requests (>=2.31.0) from runtime dependencies

No runtime dependencies remaining
```

If `dependencies.dev` is also empty:
```json
{
  "name": "my-package"
  // No dependencies section
}
```

---

## Error Handling

### No aigogo.json

```bash
$ aigogo rm file test.py
Error: failed to load aigogo.json: no such file
Run 'aigogo init' first
```

### Invalid Subcommand

```bash
$ aigogo rm
Error: usage: aigogo rm <file|dep|dev> [args...]

$ aigogo rm invalid test.py
Error: unknown subcommand 'invalid'
Valid subcommands: file, dep, dev
```

### File Not Found

```bash
$ aigogo rm file nonexistent.py
Error: no matching files found in include list
```

### Dependency Not Found

```bash
$ aigogo rm dep nonexistent
Error: package 'nonexistent' not found in runtime dependencies

$ aigogo rm dev pytest
Error: package 'pytest' not found in development dependencies
```

### No Dependencies

```bash
$ aigogo rm dep requests
Error: no runtime dependencies found in aigogo.json

$ aigogo rm dev pytest
Error: no development dependencies found in aigogo.json
```

### Auto Mode Issue

```bash
$ aigogo rm file test.py
Error: files.include is set to 'auto'
Cannot remove individual files when using auto-discovery
```

**Solution:** Manually edit `aigogo.json` to convert `"include": "auto"` to an array.

---

## Integration with Other Commands

### With add

```bash
# Add then remove
aigogo add file test.py
aigogo rm file test.py

# Update a dependency (remove + add)
aigogo rm dep requests
aigogo add dep requests ">=3.0.0"
```

### With validate

```bash
# Remove file
aigogo rm file old_api.py

# Validate remaining files exist
aigogo validate
```

### With build

```bash
# Remove unused files
aigogo rm file test_*.py

# Build without test files
aigogo build utils:1.0.0
```

---

## Tips and Best Practices

### Use Interactive Mode to See Current State

```bash
# Don't remember what deps you have?
aigogo rm dep
# Shows full list, then prompts

aigogo rm file
# Shows files, then prompts
```

### Remove Before Updating

To update a dependency version:
```bash
aigogo rm dep requests
aigogo add dep requests ">=3.0.0"
```

### Clean Up Unused Dev Dependencies

```bash
# Remove old test framework
aigogo rm dev unittest

# Add new one
aigogo add dev pytest "^7.0.0"
```

### Validate After Removal

```bash
aigogo rm dep requests
aigogo validate
# Checks if any code still imports requests
```

---

## Comparison with add

| Action | add Command | rm Command |
|--------|------------|-----------|
| Files | `add file <path>...` | `rm file <path>...` |
| Runtime deps | `add dep <pkg> <ver>` | `rm dep <pkg>` |
| Dev deps | `add dev <pkg> <ver>` | `rm dev <pkg>` |
| Interactive | Prompts for name & version | Shows list, prompts for name |
| Validation | Checks file exists | Checks file in list |

---

## Summary

**File Management:**
```bash
aigogo rm file <path>...          # Remove files
aigogo rm file old.py test.py     # Multiple files
```

**Runtime Dependencies:**
```bash
aigogo rm dep <pkg>               # Remove runtime dep
aigogo rm dep requests
```

**Dev Dependencies:**
```bash
aigogo rm dev <pkg>               # Remove dev dep
aigogo rm dev pytest
```

✅ **Features:**
- Interactive mode shows current items
- Removes multiple files at once
- Shows remaining items after removal
- Auto-cleans empty dependency sections

✅ **What it doesn't do:**
- Delete files from disk (only removes from manifest)
- Uninstall packages (use pip/npm/cargo)
- Remove entries with wildcards (must match exact name)

**Quick reference:**
```bash
aigogo rm file <path>...    # Remove files
aigogo rm dep <package>     # Remove runtime dep
aigogo rm dev <package>     # Remove dev dep
aigogo rm <subcommand>      # Interactive mode
```
