# aigg rm Command

## Overview

The `rm` command removes files or dependencies from your `aigogo.json` manifest file. It has three subcommands: `file`, `dep`, and `dev`.

**Subdirectory Support**: Like `git`, this command works from any subdirectory. aigogo searches up the tree to find `aigogo.json`.

## Subcommands

### `rm file` - Remove Files from Package

Removes files or patterns from the `files.include` array.

**Usage:**
```bash
aigg rm file <path>...
```

**Features:**
- Removes specified files from include list
- Supports removing multiple files at once
- Interactive mode shows current files
- Shows remaining files after removal

**Examples:**
```bash
# Remove single file
aigg rm file old_utils.py

# Remove multiple files
aigg rm file test.py helper.py config.json

# Remove glob pattern
aigg rm file "*.pyc"
```

**Interactive Mode:**
```bash
$ aigg rm file
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
$ aigg rm file nonexistent.py
Error: no matching files found in include list

# files.include is set to "auto"
$ aigg rm file test.py
Error: files.include is set to 'auto'
Cannot remove individual files when using auto-discovery

# No files in include list
$ aigg rm file test.py
Error: no files in include list
```

---

### `rm dep` - Remove Runtime Dependency

Removes a runtime dependency from `dependencies.runtime`.

**Usage:**
```bash
aigg rm dep <package>
```

**Examples:**
```bash
# Python
aigg rm dep requests

# JavaScript
aigg rm dep axios

# Go
aigg rm dep github.com/pkg/errors

# Rust
aigg rm dep serde
```

**Interactive Mode:**
```bash
$ aigg rm dep
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
aigg rm dev <package>
```

**Examples:**
```bash
# Python testing
aigg rm dev pytest
aigg rm dev black

# JavaScript testing
aigg rm dev jest

# Go testing
aigg rm dev github.com/stretchr/testify
```

**Interactive Mode:**
```bash
$ aigg rm dev
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
aigg rm file old_api.py deprecated.py

# Verify changes
aigg validate
```

### Cleaning Up Dependencies

```bash
# Remove unused runtime dependency
aigg rm dep requests

# Remove unused dev dependency
aigg rm dev pytest

# Validate
aigg validate
```

### Interactive Cleanup

```bash
# See all dependencies and remove interactively
aigg rm dep
# Shows list, prompts for package name

aigg rm dev
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

After `aigg rm file utils.py`:
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
$ aigg rm file api.py config.json
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
$ aigg rm dep requests
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
$ aigg rm file test.py
Error: failed to load aigogo.json: no such file
Run 'aigg init' first
```

### Invalid Subcommand

```bash
$ aigg rm
Error: usage: aigg rm <file|dep|dev> [args...]

$ aigg rm invalid test.py
Error: unknown subcommand 'invalid'
Valid subcommands: file, dep, dev
```

### File Not Found

```bash
$ aigg rm file nonexistent.py
Error: no matching files found in include list
```

### Dependency Not Found

```bash
$ aigg rm dep nonexistent
Error: package 'nonexistent' not found in runtime dependencies

$ aigg rm dev pytest
Error: package 'pytest' not found in development dependencies
```

### No Dependencies

```bash
$ aigg rm dep requests
Error: no runtime dependencies found in aigogo.json

$ aigg rm dev pytest
Error: no development dependencies found in aigogo.json
```

### Auto Mode Issue

```bash
$ aigg rm file test.py
Error: files.include is set to 'auto'
Cannot remove individual files when using auto-discovery
```

**Solution:** Manually edit `aigogo.json` to convert `"include": "auto"` to an array.

---

## Integration with Other Commands

### With add

```bash
# Add then remove
aigg add file test.py
aigg rm file test.py

# Update a dependency (remove + add)
aigg rm dep requests
aigg add dep requests ">=3.0.0"
```

### With validate

```bash
# Remove file
aigg rm file old_api.py

# Validate remaining files exist
aigg validate
```

### With build

```bash
# Remove unused files
aigg rm file test_*.py

# Build without test files
aigg build utils:1.0.0
```

---

## Tips and Best Practices

### Use Interactive Mode to See Current State

```bash
# Don't remember what deps you have?
aigg rm dep
# Shows full list, then prompts

aigg rm file
# Shows files, then prompts
```

### Remove Before Updating

To update a dependency version:
```bash
aigg rm dep requests
aigg add dep requests ">=3.0.0"
```

### Clean Up Unused Dev Dependencies

```bash
# Remove old test framework
aigg rm dev unittest

# Add new one
aigg add dev pytest "^7.0.0"
```

### Validate After Removal

```bash
aigg rm dep requests
aigg validate
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
aigg rm file <path>...          # Remove files
aigg rm file old.py test.py     # Multiple files
```

**Runtime Dependencies:**
```bash
aigg rm dep <pkg>               # Remove runtime dep
aigg rm dep requests
```

**Dev Dependencies:**
```bash
aigg rm dev <pkg>               # Remove dev dep
aigg rm dev pytest
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
aigg rm file <path>...    # Remove files
aigg rm dep <package>     # Remove runtime dep
aigg rm dev <package>     # Remove dev dep
aigg rm <subcommand>      # Interactive mode
```
