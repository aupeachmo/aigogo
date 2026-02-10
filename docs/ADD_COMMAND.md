# aigg add Command

## Overview

The `add` command adds files or dependencies to your `aigogo.json` manifest file. It has three subcommands: `file`, `dep`, and `dev`.

**Subdirectory Support**: Like `git`, this command works from any subdirectory. aigogo searches up the tree to find `aigogo.json`.

## Subcommands

### `add file` - Add Files to Package

Adds files or glob patterns to the `files.include` array.

**Usage:**
```bash
aigg add file <path>...
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
aigg add file api_client.py

# Add multiple files
aigg add file utils.py helpers.py config.json

# Add with glob pattern
aigg add file "*.py"

# Add with directory glob
aigg add file "lib/**/*.js"

# Add file that's normally ignored by .aigogoignore
aigg add file generated_code.py --force
# Equivalent:
aigg add file --force generated_code.py
```

**Error Handling:**
```bash
# File doesn't exist
$ aigg add file nonexistent.py
Error: file does not exist: nonexistent.py

# Glob matches nothing
$ aigg add file "*.xyz"
Error: glob pattern '*.xyz' matches no files

# files.include is set to "auto"
$ aigg add file test.py
Error: files.include is set to 'auto'
Please edit aigogo.json to change it to an array before adding files
```

**Interactive Mode:**
```bash
$ aigg add file
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
aigg add dep <package> <version>

# Import from pyproject.toml (Python only)
aigg add dep --from-pyproject
```

**Manual Examples:**
```bash
# Python
aigg add dep requests ">=2.31.0,<3.0.0"

# JavaScript
aigg add dep axios "^1.6.0"

# Go
aigg add dep github.com/pkg/errors v0.9.1

# Rust
aigg add dep serde "1.0"
```

**Import from pyproject.toml (Python):**

For Python projects, you can automatically import dependencies from `pyproject.toml`:

```bash
# Import all runtime dependencies
$ aigg add dep --from-pyproject
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
$ aigg add dep
Package name: requests
Version constraint (e.g., >=2.31.0,<3.0.0): >=2.31.0,<3.0.0
✓ Added requests >=2.31.0,<3.0.0 to runtime dependencies

Next steps:
  1. Run 'aigg validate' to check your dependencies
  2. Run 'aigg scan' to detect any missing dependencies
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
aigg add dev <package> <version>

# Import from pyproject.toml (Python only)
aigg add dev --from-pyproject
```

**Manual Examples:**
```bash
# Python testing
aigg add dev pytest "^7.0.0"
aigg add dev black ">=23.0.0"

# JavaScript testing
aigg add dev jest "^29.0.0"

# Go testing
aigg add dev github.com/stretchr/testify v1.8.4
```

**Import from pyproject.toml (Python):**

```bash
# Import all development dependencies
$ aigg add dev --from-pyproject
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
$ aigg add dev
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
aigg init

# Write code
cat > utils.py <<'EOF'
import requests

def fetch_data(url):
    return requests.get(url).json()
EOF

# Add file
aigg add file utils.py

# Add dependencies
aigg add dep requests ">=2.31.0,<3.0.0"
aigg add dev pytest "^7.0.0"

# Validate and build
aigg validate
aigg build my-utils:1.0.0
```

### Python Project with pyproject.toml

If you already have a Poetry or uv project with `pyproject.toml`:

```bash
cd ~/my-poetry-project

# Initialize aigogo
aigg init

# Add files
aigg add file "src/**/*.py"

# Import all dependencies from pyproject.toml
aigg add dep --from-pyproject
aigg add dev --from-pyproject

# Build
aigg build my-project:1.0.0
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
aigg init

# Add multiple files
aigg add file "src/*.py" lib/helpers.py config/settings.py

# Add dependencies
aigg add dep requests ">=2.31.0"
aigg add dep pyyaml ">=6.0"
aigg add dev pytest "^7.0.0"
aigg add dev black ">=23.0.0"

# Build
aigg build my-project:1.0.0
```

### Using Glob Patterns

```bash
# Python project
aigg add file "*.py"
aigg add file "src/**/*.py"

# JavaScript project
aigg add file "lib/**/*.js"
aigg add file "index.js"

# Multiple extensions
aigg add file "*.py" "*.json" "*.yaml"
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

After `aigg add file utils.py`:

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
$ aigg add file test.py
✓ Added 1 file(s) to include list

$ aigg add file test.py
⚠ Skipping 'test.py' (already in include list)
No new files to add
```

Dependencies:
```bash
$ aigg add dep requests ">=3.0.0"
Error: package 'requests' is already declared as a runtime dependency with version '>=2.31.0,<3.0.0'
```

### Preventing "auto" Mode Issues

If `files.include` is set to `"auto"`, file commands will error:

```bash
$ aigg add file test.py
Error: files.include is set to 'auto'
Please edit aigogo.json to change it to an array before adding files
```

**Note:** As of v2.0, `aigg init` creates an empty array instead of `"auto"`, so this should only occur in manually created manifests.

---

## Error Handling

### No aigogo.json

```bash
$ aigg add file test.py
Error: failed to load aigogo.json: no such file
Run 'aigg init' first
```

### Invalid Subcommand

```bash
$ aigg add
Error: usage: aigg add <file|dep|dev> [args...]

$ aigg add invalid test.py
Error: unknown subcommand 'invalid'
Valid subcommands: file, dep, dev
```

### File Validation Errors

```bash
# Non-existent file (literal path)
$ aigg add file missing.py
Error: file does not exist: missing.py

# Invalid glob pattern
$ aigg add file "[invalid"
Error: invalid glob pattern '[invalid': syntax error

# Glob matches nothing
$ aigg add file "*.nonexistent"
Error: glob pattern '*.nonexistent' matches no files
```

### Empty Arguments

```bash
$ aigg add dep
Package name: [press enter]
Error: package name is required

$ aigg add dep requests ""
Error: version is required
```

---

## Integration with Other Commands

### With scan

```bash
# Scan finds imports
aigg scan
# Shows: "requests detected"

# Add the dependency
aigg add dep requests ">=2.31.0,<3.0.0"

# Validate it's correct
aigg validate
# ✅ All detected dependencies declared
```

### With validate

```bash
# Add files and deps
aigg add file api.py
aigg add dep requests ">=2.31.0"

# Validate configuration
aigg validate
# Checks if:
# - api.py exists
# - requests is actually imported
```

### With build

```bash
# Add everything
aigg add file "*.py"
aigg add dep requests ">=2.31.0"

# Build
aigg build utils:1.0.0
# Packages api.py with generated requirements.txt
```

---

## Tips and Best Practices

### Use Glob Patterns for Multiple Files

```bash
# Instead of:
aigg add file a.py b.py c.py d.py

# Use:
aigg add file "*.py"
```

### Add Files Before Building

Always explicitly add files instead of relying on `"auto"`:

```bash
aigg init
aigg add file src/*.py lib/utils.py
aigg build my-pkg:1.0.0
```

### Add Dev Dependencies Separately

```bash
# Runtime deps
aigg add dep requests ">=2.31.0"
aigg add dep click ">=8.0.0"

# Dev deps
aigg add dev pytest "^7.0.0"
aigg add dev black ">=23.0.0"
```

### Validate After Adding

```bash
aigg add file utils.py
aigg add dep requests ">=2.31.0"
aigg validate  # Catches issues early
```

---

## Summary

**File Management:**
```bash
aigg add file <path>...         # Add files/globs
aigg add file "*.py" lib/*.js   # Multiple patterns
```

**Runtime Dependencies:**
```bash
aigg add dep <pkg> <ver>        # Add runtime dep
aigg add dep requests ">=2.31.0,<3.0.0"
```

**Dev Dependencies:**
```bash
aigg add dev <pkg> <ver>        # Add dev dep
aigg add dev pytest "^7.0.0"
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
