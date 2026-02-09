# aigogo add Command

## Overview

The `add` command adds files or dependencies to your `aigogo.json` manifest file. It has three subcommands: `file`, `dep`, and `dev`.

**Subdirectory Support**: Like `git`, this command works from any subdirectory. aigogo searches up the tree to find `aigogo.json`.

## Subcommands

### `add file` - Add Files to Package

Adds files or glob patterns to the `files.include` array.

**Usage:**
```bash
aigogo add file <path>...
```

**Features:**
- Validates files exist before adding
- Supports glob patterns (`*.py`, `lib/**/*.js`)
- Prevents duplicate entries
- Checks for multiple files at once
- Respects `.aigogoignore` (use `--force` to override)
- Flags can appear before or after file paths

**Examples:**
```bash
# Add single file
aigogo add file api_client.py

# Add multiple files
aigogo add file utils.py helpers.py config.json

# Add with glob pattern
aigogo add file "*.py"

# Add with directory glob
aigogo add file "lib/**/*.js"

# Add file that's normally ignored by .aigogoignore
aigogo add file generated_code.py --force
# Equivalent:
aigogo add file --force generated_code.py
```

**Error Handling:**
```bash
# File doesn't exist
$ aigogo add file nonexistent.py
Error: file does not exist: nonexistent.py

# Glob matches nothing
$ aigogo add file "*.xyz"
Error: glob pattern '*.xyz' matches no files

# files.include is set to "auto"
$ aigogo add file test.py
Error: files.include is set to 'auto'
Please edit aigogo.json to change it to an array before adding files
```

**Interactive Mode:**
```bash
$ aigogo add file
File paths (space-separated): api_client.py utils.py
✓ Added 2 file(s) to include list:
  - api_client.py
  - utils.py
```

---

### `add dep` - Add Runtime Dependency

Adds a runtime dependency to `dependencies.runtime`.

**Usage:**
```bash
# Manual mode
aigogo add dep <package> <version>

# Import from pyproject.toml (Python only)
aigogo add dep --from-pyproject
```

**Manual Examples:**
```bash
# Python
aigogo add dep requests ">=2.31.0,<3.0.0"

# JavaScript
aigogo add dep axios "^1.6.0"

# Go
aigogo add dep github.com/pkg/errors v0.9.1

# Rust
aigogo add dep serde "1.0"
```

**Import from pyproject.toml (Python):**

For Python projects, you can automatically import dependencies from `pyproject.toml`:

```bash
# Import all runtime dependencies
$ aigogo add dep --from-pyproject
📦 Reading dependencies from: /path/to/pyproject.toml
✓ Detected format: poetry
✓ Set Python version requirement: ^3.9

Adding 3 runtime dependencies...

✓ Added requests >=2.31.0,<3.0.0
✓ Added pyyaml >=6.0,<7.0
✓ Added click >=8.0.0,<9.0.0

✓ Successfully added 3 dependencies
```

**Supported Formats:**
- **Poetry**: Reads from `[tool.poetry.dependencies]`
- **uv/PEP 621**: Reads from `[project.dependencies]`

Both formats will automatically extract and set the Python version from:
- Poetry: `tool.poetry.dependencies.python`
- PEP 621: `project.requires-python`

**Interactive Mode:**
```bash
$ aigogo add dep
Package name: requests
Version constraint (e.g., >=2.31.0,<3.0.0): >=2.31.0,<3.0.0
✓ Added requests >=2.31.0,<3.0.0 to runtime dependencies

Next steps:
  1. Run 'aigogo validate' to check your dependencies
  2. Run 'aigogo scan' to detect any missing dependencies
```

**Version Format Suggestions:**

| Language | Suggestion | Example |
|----------|------------|---------|
| Python | `>=2.31.0,<3.0.0` | Range with upper bound |
| JavaScript | `^1.6.0` or `~1.6.0` | Caret or tilde |
| Go | `v1.2.3` | Specific version with v prefix |
| Rust | `1.0` or `^1.0` | Major.minor or caret |

---

### `add dev` - Add Development Dependency

Adds a development dependency to `dependencies.dev`.

**Usage:**
```bash
# Manual mode
aigogo add dev <package> <version>

# Import from pyproject.toml (Python only)
aigogo add dev --from-pyproject
```

**Manual Examples:**
```bash
# Python testing
aigogo add dev pytest "^7.0.0"
aigogo add dev black ">=23.0.0"

# JavaScript testing
aigogo add dev jest "^29.0.0"

# Go testing
aigogo add dev github.com/stretchr/testify v1.8.4
```

**Import from pyproject.toml (Python):**

```bash
# Import all development dependencies
$ aigogo add dev --from-pyproject
📦 Reading dependencies from: /path/to/pyproject.toml
✓ Detected format: poetry

Adding 2 development dependencies...

✓ Added pytest ^7.0.0
✓ Added black >=23.0.0,<24.0.0

✓ Successfully added 2 dependencies
```

**Supported Dev Dependency Sources:**
- **Poetry**: `[tool.poetry.dev-dependencies]` and `[tool.poetry.group.dev.dependencies]`
- **PEP 621**: `[project.optional-dependencies]` (groups containing "dev", "test", or "doc")

**Interactive Mode:**
```bash
$ aigogo add dev
Package name: pytest
Version constraint (e.g., >=2.31.0,<3.0.0): ^7.0.0
✓ Added pytest ^7.0.0 to development dependencies
```

---

## Complete Workflow Examples

### New Project with Files and Dependencies

```bash
# Initialize
mkdir my-utils && cd my-utils
aigogo init

# Write code
cat > utils.py <<'EOF'
import requests

def fetch_data(url):
    return requests.get(url).json()
EOF

# Add file
aigogo add file utils.py

# Add dependencies
aigogo add dep requests ">=2.31.0,<3.0.0"
aigogo add dev pytest "^7.0.0"

# Validate and build
aigogo validate
aigogo build my-utils:1.0.0
```

### Python Project with pyproject.toml

If you already have a Poetry or uv project with `pyproject.toml`:

```bash
cd ~/my-poetry-project

# Initialize aigogo
aigogo init

# Add files
aigogo add file "src/**/*.py"

# Import all dependencies from pyproject.toml
aigogo add dep --from-pyproject
aigogo add dev --from-pyproject

# Build
aigogo build my-project:1.0.0
```

**Example pyproject.toml (Poetry):**
```toml
[tool.poetry]
name = "my-project"
version = "0.1.0"

[tool.poetry.dependencies]
python = "^3.9"
requests = ">=2.31.0,<3.0.0"
pyyaml = ">=6.0,<7.0"

[tool.poetry.group.dev.dependencies]
pytest = "^7.0.0"
black = ">=23.0.0"
```

**Example pyproject.toml (uv/PEP 621):**
```toml
[project]
name = "my-project"
version = "0.1.0"
requires-python = ">=3.9"
dependencies = [
    "requests>=2.31.0",
    "pyyaml>=6.0,<7.0",
]

[project.optional-dependencies]
dev = [
    "pytest>=7.0.0",
    "black>=23.0.0",
]
```

### Existing Project with Multiple Files

```bash
cd ~/my-existing-project

# Initialize
aigogo init

# Add multiple files
aigogo add file "src/*.py" lib/helpers.py config/settings.py

# Add dependencies
aigogo add dep requests ">=2.31.0"
aigogo add dep pyyaml ">=6.0"
aigogo add dev pytest "^7.0.0"
aigogo add dev black ">=23.0.0"

# Build
aigogo build my-project:1.0.0
```

### Using Glob Patterns

```bash
# Python project
aigogo add file "*.py"
aigogo add file "src/**/*.py"

# JavaScript project
aigogo add file "lib/**/*.js"
aigogo add file "index.js"

# Multiple extensions
aigogo add file "*.py" "*.json" "*.yaml"
```

---

## Behavior Details

### Adding First File

If `files.include` is empty or null:

```json
{
  "files": {
    "include": []
  }
}
```

After `aigogo add file utils.py`:

```json
{
  "files": {
    "include": ["utils.py"]
  }
}
```

### Duplicate Detection

Files:
```bash
$ aigogo add file test.py
✓ Added 1 file(s) to include list

$ aigogo add file test.py
⚠ Skipping 'test.py' (already in include list)
No new files to add
```

Dependencies:
```bash
$ aigogo add dep requests ">=3.0.0"
Error: package 'requests' is already declared as a runtime dependency with version '>=2.31.0,<3.0.0'
```

### Preventing "auto" Mode Issues

If `files.include` is set to `"auto"`, file commands will error:

```bash
$ aigogo add file test.py
Error: files.include is set to 'auto'
Please edit aigogo.json to change it to an array before adding files
```

**Note:** As of v2.0, `aigogo init` creates an empty array instead of `"auto"`, so this should only occur in manually created manifests.

---

## Error Handling

### No aigogo.json

```bash
$ aigogo add file test.py
Error: failed to load aigogo.json: no such file
Run 'aigogo init' first
```

### Invalid Subcommand

```bash
$ aigogo add
Error: usage: aigogo add <file|dep|dev> [args...]

$ aigogo add invalid test.py
Error: unknown subcommand 'invalid'
Valid subcommands: file, dep, dev
```

### File Validation Errors

```bash
# Non-existent file (literal path)
$ aigogo add file missing.py
Error: file does not exist: missing.py

# Invalid glob pattern
$ aigogo add file "[invalid"
Error: invalid glob pattern '[invalid': syntax error

# Glob matches nothing
$ aigogo add file "*.nonexistent"
Error: glob pattern '*.nonexistent' matches no files
```

### Empty Arguments

```bash
$ aigogo add dep
Package name: [press enter]
Error: package name is required

$ aigogo add dep requests ""
Error: version is required
```

---

## Integration with Other Commands

### With scan

```bash
# Scan finds imports
aigogo scan
# Shows: "requests detected"

# Add the dependency
aigogo add dep requests ">=2.31.0,<3.0.0"

# Validate it's correct
aigogo validate
# ✅ All detected dependencies declared
```

### With validate

```bash
# Add files and deps
aigogo add file api.py
aigogo add dep requests ">=2.31.0"

# Validate configuration
aigogo validate
# Checks if:
# - api.py exists
# - requests is actually imported
```

### With build

```bash
# Add everything
aigogo add file "*.py"
aigogo add dep requests ">=2.31.0"

# Build
aigogo build utils:1.0.0
# Packages api.py with generated requirements.txt
```

---

## Tips and Best Practices

### Use Glob Patterns for Multiple Files

```bash
# Instead of:
aigogo add file a.py b.py c.py d.py

# Use:
aigogo add file "*.py"
```

### Add Files Before Building

Always explicitly add files instead of relying on `"auto"`:

```bash
aigogo init
aigogo add file src/*.py lib/utils.py
aigogo build my-pkg:1.0.0
```

### Add Dev Dependencies Separately

```bash
# Runtime deps
aigogo add dep requests ">=2.31.0"
aigogo add dep click ">=8.0.0"

# Dev deps
aigogo add dev pytest "^7.0.0"
aigogo add dev black ">=23.0.0"
```

### Validate After Adding

```bash
aigogo add file utils.py
aigogo add dep requests ">=2.31.0"
aigogo validate  # Catches issues early
```

---

## Summary

**File Management:**
```bash
aigogo add file <path>...         # Add files/globs
aigogo add file "*.py" lib/*.js   # Multiple patterns
```

**Runtime Dependencies:**
```bash
aigogo add dep <pkg> <ver>        # Add runtime dep
aigogo add dep requests ">=2.31.0,<3.0.0"
```

**Dev Dependencies:**
```bash
aigogo add dev <pkg> <ver>        # Add dev dep
aigogo add dev pytest "^7.0.0"
```

✅ **Features:**
- File existence validation
- Glob pattern support
- Duplicate detection
- Language-specific version hints
- Interactive prompts

✅ **What it doesn't do:**
- Install packages (use pip/npm/cargo)
- Fetch available versions
- Update existing entries (use `rm` then `add`)
