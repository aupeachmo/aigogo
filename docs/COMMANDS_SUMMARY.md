# aigg Commands Summary

Complete reference of all aigg commands and their purposes.

## Command Overview

| Command | Scope | Purpose | Destructive |
|---------|-------|---------|-------------|
| `init` | Local | Initialize new package | No |
| `add file` | Local | Add files to manifest | No |
| `add dep` | Local | Add runtime dependency to manifest | No |
| `add dev` | Local | Add dev dependency to manifest | No |
| `rm file` | Local | Remove files from manifest | No |
| `rm dep` | Local | Remove runtime dependency from manifest | No |
| `rm dev` | Local | Remove dev dependency from manifest | No |
| `validate` | Local | Check dependencies vs imports | No |
| `scan` | Local | Detect dependencies from code | No |
| `install` | Local | Install packages from aigogo.lock | No |
| `build` | Local | Build package (auto-version or explicit) | No |
| `push` | Remote | Upload package to registry | No |
| `pull` | Remote | Download package (no extract) | No |
| `list` | Local | Show cached packages | No |
| `show-deps` | Local | Display dependencies in various formats | No |
| `remove` | Local | Delete from local cache | Yes (local) |
| `remove-all` | Local | Delete all from local cache | Yes (local) |
| `delete` | Remote | Delete from registry | ⚠️ Yes (permanent) |
| `login` | Auth | Authenticate with registry | No |
| `logout` | Auth | Remove registry credentials | No |
| `search` | Remote | Search registry (placeholder) | No |
| `version` | Info | Show version | No |
| `completion` | Info | Generate shell completion | No |

## Command Categories

### 📝 Manifest Management (Local)

**`init`** - Initialize new package
```bash
aigg init
# Creates aigogo.json with empty files array
```

**`add file`** - Add files to package
```bash
aigg add file utils.py api.py
aigg add file "*.py"  # Glob patterns supported
# Updates aigogo.json files.include
```

**`add dep`** - Add runtime dependency
```bash
# Manual mode
aigg add dep requests ">=2.31.0,<3.0.0"
# Updates aigogo.json dependencies.runtime

# Import from pyproject.toml (Python only)
aigg add dep --from-pyproject
# Reads [tool.poetry.dependencies] or [project.dependencies]
# Automatically sets Python version requirement
```

**`add dev`** - Add dev dependency
```bash
# Manual mode
aigg add dev pytest "^7.0.0"
# Updates aigogo.json dependencies.dev

# Import from pyproject.toml (Python only)
aigg add dev --from-pyproject
# Reads [tool.poetry.dev-dependencies] or [project.optional-dependencies]
```

**`rm file`** - Remove files from package
```bash
aigg rm file old_utils.py
# Updates aigogo.json files.include
```

**`rm dep`** - Remove runtime dependency
```bash
aigg rm dep requests
# Updates aigogo.json dependencies.runtime
```

**`rm dev`** - Remove dev dependency
```bash
aigg rm dev pytest
# Updates aigogo.json dependencies.dev
```

### ✅ Validation (Local)

**`validate`** - Check dependencies
```bash
aigg validate
# Scans files, compares with declared deps
```

**`scan`** - Detect dependencies
```bash
aigg scan
# Analyzes imports, suggests what to add
```

### 📥 Build & Install (Local)

**`build`** - Build package locally
```bash
aigg build utils:1.0.0       # Explicit name and tag
aigg build                   # Auto-increment patch version
aigg build --force           # Rebuild even if exists
aigg build --no-validate     # Skip dependency validation
```

**`install`** - Install packages from lock file
```bash
aigg install
# Reads aigogo.lock, stores packages in CAS, creates import symlinks
# Python: from aigogo.package_name import ...
# JavaScript: import ... from '@aigogo/package-name'
```

### 📦 Distribution (Remote)

**`push`** - Upload to registry
```bash
# Push to Docker Hub
aigg push docker.io/myorg/utils:1.0.0 --from utils:1.0.0

# Push to GitHub Container Registry
aigg push ghcr.io/myorg/utils:1.0.0 --from utils:1.0.0
```

**`pull`** - Download only
```bash
aigg pull docker.io/myorg/utils:1.0.0
aigg pull ghcr.io/myorg/utils:1.0.0
# Pulls from registry, saves to cache
```

### 🗑️ Cleanup

**`remove`** - Delete from local cache
```bash
aigg remove docker.io/myorg/utils:1.0.0
# Removes from ~/.aigogo/cache/
```

**`remove-all`** - Delete all from local cache
```bash
aigg remove-all              # Prompts for confirmation
aigg remove-all --force      # Skip confirmation
# Removes everything from ~/.aigogo/cache/
```

**`delete`** - Delete from registry ⚠️
```bash
# Delete specific tag
aigg delete docker.io/myorg/utils:1.0.0

# Delete all tags
aigg delete docker.io/myorg/utils --all
# Permanently removes from remote registry
```

### 🔐 Authentication

**`login`** - Authenticate
```bash
# Interactive login (prompts for username and password)
aigg login docker.io

# Use Docker Hub shortcut (no registry argument needed)
aigg login --dockerhub

# GitHub Container Registry (use a PAT as password)
aigg login ghcr.io

# With username flag
aigg login docker.io -u myusername

# Read password from stdin (prevents password in shell history)
echo "mypassword" | aigg login docker.io -u myusername -p

# ghcr.io with PAT from stdin
echo "$GHCR_PAT" | aigg login ghcr.io -u myusername -p

# Docker Hub with stdin password
echo "mypassword" | aigg login --dockerhub -u myusername -p
# Stores credentials for registry access
```

**`logout`** - Remove credentials
```bash
aigg logout docker.io
aigg logout ghcr.io
# Removes stored credentials for the specified registry
```

### 🔍 Discovery

**`search`** - Search registry
```bash
aigg search utils
# (Placeholder - not fully implemented)
```

### ℹ️ Information

**`version`** - Show version
```bash
aigg version
# Shows aigg version, platform, Go version
```

**`list`** - List cached packages
```bash
aigg list
# Shows packages in local cache with:
#   - Package name and version
#   - Type (local build or registry pull)
#   - Build/pull time
#   - Size
#   - Language and version (if available)
#   - Dependency count (runtime and dev)
```

**`show-deps`** - Display dependencies in various formats
```bash
aigg show-deps <path>                        # Text format (default)
aigg show-deps <path> --format pyproject     # PEP 621 format (alias: pep621)
aigg show-deps <path> --format poetry        # Poetry format
aigg show-deps <path> --format requirements  # pip requirements.txt (alias: pip)
aigg show-deps <path> --format npm           # package.json fragment (alias: package-json)
aigg show-deps <path> --format yarn          # yarn add commands

# Path can be:
# - Directory containing aigogo.json
# - Direct path to aigogo.json file
# - Current directory (.)
```

## Common Workflows

### Creating and Publishing a Snippet

```bash
# 1. Initialize
aigg init

# 2. Add files
aigg add file utils.py helpers.py

# 3. Add dependencies
aigg add dep requests ">=2.31.0,<3.0.0"
aigg add dep click ">=8.0.0,<9.0.0"
aigg add dev pytest "^7.0.0"

# 4. Validate
aigg validate

# 5. Build locally
aigg build utils:1.0.0

# 6. Login
aigg login --dockerhub -u myusername
# Or: aigg login docker.io -u myusername
# Or: aigg login ghcr.io -u myusername  (use PAT as password)

# 7. Push
aigg push docker.io/myorg/utils:1.0.0 --from utils:1.0.0
# Or: aigg push ghcr.io/myorg/utils:1.0.0 --from utils:1.0.0
```

### Using a Snippet

```bash
# 1. Login (if private)
aigg login --dockerhub
# Or: aigg login docker.io
# Or: aigg login ghcr.io

# 2. Add and install snippet
aigg add docker.io/myorg/utils:1.0.0
# Or: aigg add ghcr.io/myorg/utils:1.0.0
aigg install

# 3. Install external dependencies
aigg show-deps .aigogo/imports/aigogo/utils --format requirements | pip install -r /dev/stdin

# 4. Use the code
python -c "from aigogo.utils import fetch_data; print(fetch_data('https://api.example.com'))"
```

### Maintaining Dependencies

```bash
# Check what's actually used
aigg scan

# Add missing runtime dependencies
aigg add dep <package> <version>

# Add dev dependencies
aigg add dev <package> <version>

# For Python projects with pyproject.toml
aigg add dep --from-pyproject
aigg add dev --from-pyproject

# Remove unused dependencies
aigg rm dep <package>
aigg rm dev <package>

# Validate everything matches
aigg validate
```

### Integrating Snippets into Existing Projects

```bash
# Install the snippet
aigg add utils:1.0.0
aigg install

# View dependencies in format matching your project
# For uv/PEP 621 projects:
aigg show-deps .aigogo/imports/aigogo/utils --format pyproject

# For Poetry projects:
aigg show-deps .aigogo/imports/aigogo/utils --format poetry

# For pip/requirements.txt:
aigg show-deps .aigogo/imports/aigogo/utils --format requirements >> requirements.txt

# Copy the output into your pyproject.toml or requirements.txt
# Then install:
pip install -e .
# or
poetry install
```

### Cleaning Up

```bash
# Remove old cached downloads
aigg list
aigg remove docker.io/myorg/utils:0.9.0

# Remove all cached packages
aigg remove-all              # Prompts for confirmation
aigg remove-all --force      # Skip confirmation

# Delete specific old version from registry
aigg delete docker.io/myorg/utils:0.9.0

# Delete entire deprecated repository from registry
aigg delete docker.io/myorg/deprecated-utils --all
```

## Destruction Matrix

Understanding what each destructive command affects:

| Command | Affects | Reversible | How to Reverse |
|---------|---------|------------|----------------|
| `rm file/dep/dev` | Local manifest | ✅ Yes | Re-add with `aigg add file/dep/dev` |
| `remove` | Local cache (single) | ✅ Yes | Re-add with `aigg add` + `aigg install` |
| `remove-all` | Local cache (all) | ✅ Yes | Re-add with `aigg add` + `aigg install` |
| `delete` | Remote registry | ❌ **NO** | Must re-push |

## Command Comparison

### `rm` vs `remove` vs `remove-all` vs `delete`

Often confused - here's the difference:

```bash
# rm - Edits aigogo.json
aigg rm dep requests  # Remove runtime dependency
aigg rm dev pytest    # Remove dev dependency
aigg rm file old.py   # Remove file from include list
# Effect: Removes from aigogo.json
# File: aigogo.json modified
# Reversible: Yes (aigg add dep/dev/file ...)

# remove - Cleans local cache (single package)
aigg remove docker.io/myorg/utils:1.0.0
# Effect: Deletes from ~/.aigogo/cache/
# Files: Specific cached package deleted
# Reversible: Yes (aigg add ... + aigg install)

# remove-all - Cleans local cache (all packages)
aigg remove-all              # Prompts for confirmation
aigg remove-all --force      # Skip confirmation
# Effect: Deletes everything from ~/.aigogo/cache/
# Files: All cached packages deleted
# Reversible: Yes (aigg add ... + aigg install)

# delete - Removes from registry
aigg delete docker.io/myorg/utils:1.0.0
# Effect: Permanently removes from remote
# Registry: Image deleted
# Reversible: NO (must re-push)
```

### `push` vs `pull`

```bash
# push - Upload to registry
aigg push docker.io/myorg/utils:1.0.0
# Direction: Local → Registry
# Creates: Remote image

# pull - Download to cache
aigg pull docker.io/myorg/utils:1.0.0
# Direction: Registry → Cache
# Creates: ~/.aigogo/cache/docker.io_myorg_utils_1.0.0/

# To install after pulling, use add + install:
aigg add docker.io/myorg/utils:1.0.0
aigg install
```

### `add` vs `scan`

```bash
# add - Manually declare files/dependencies
aigg add file utils.py
aigg add dep requests ">=2.31.0,<3.0.0"
aigg add dev pytest "^7.0.0"
# Action: You tell aigogo what you need
# Updates: aigogo.json

# scan - Auto-detect dependencies
aigg scan
# Action: aigogo tells you what you're using
# Updates: Nothing (just shows suggestions)
```

## Tips

### Use Tab Completion

If your shell supports it:
```bash
aigg <TAB>
# Shows all commands
```

### Check Before Destructive Operations

```bash
# Before delete
aigg pull docker.io/myorg/utils:1.0.0  # Backup locally
aigg delete docker.io/myorg/utils:1.0.0  # Then delete

# Before remove
aigg list  # See what you have
aigg remove docker.io/myorg/utils:1.0.0  # Remove specific one
```

### Chain Commands

```bash
# Add multiple dependencies
aigg add dep requests ">=2.31.0,<3.0.0" && \
aigg add dep click ">=8.0.0,<9.0.0" && \
aigg validate

# Full publish flow
aigg init && \
aigg add file utils.py && \
aigg add dep requests ">=2.31.0,<3.0.0" && \
aigg validate && \
aigg build utils:1.0.0 && \
aigg push docker.io/myorg/utils:1.0.0 --from utils:1.0.0
```

### Use Aliases

```bash
# Add to ~/.bashrc or ~/.zshrc
alias ag='aigg'
alias agaf='aigg add file'
alias agad='aigg add dep'
alias agav='aigg add dev'
alias agrf='aigg rm file'
alias agrd='aigg rm dep'
alias agv='aigg validate'
alias ags='aigg scan'

# Use them
ag init
agaf utils.py
agad requests ">=2.31.0,<3.0.0"
agav pytest "^7.0.0"
agv
```

## See Also

- [ADD_COMMAND.md](ADD_COMMAND.md) - Detailed `add` documentation
- [RM_COMMAND.md](RM_COMMAND.md) - Detailed `rm` documentation
- [SHOW_DEPS_COMMAND.md](SHOW_DEPS_COMMAND.md) - Detailed `show-deps` documentation
- [DELETE_COMMAND.md](DELETE_COMMAND.md) - Detailed `delete` documentation
- [README.md](../README.md) - Main project documentation
- [QUICKSTART_V2.md](../QUICKSTART_V2.md) - Quick start guide

