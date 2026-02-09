# aigogo show-deps Command

## Overview

The `show-deps` command displays dependencies from an `aigogo.json` manifest in various output formats. This is particularly useful when integrating aigogo snippets into existing projects with their own dependency management systems.

## Usage

```bash
aigogo show-deps <path> [--format <format>]
```

**Arguments:**
- `<path>` - Path to `aigogo.json` file or directory containing it

**Flags:**
- `--format <format>` - Output format (default: `text`)
  - `text` - Human-readable text format
  - `pyproject` - PEP 621 format for `pyproject.toml` (alias: `pep621`)
  - `poetry` - Poetry format for `pyproject.toml`
  - `requirements` - pip requirements.txt format (alias: `pip`)
  - `npm` - package.json dependencies fragment (alias: `package-json`)
  - `yarn` - yarn add commands

---

## Output Formats

### Text Format (Default)

Human-readable output showing package info and dependencies:

```bash
$ aigogo show-deps vendor/my-snippet

Package: my-snippet
Version: 1.0.0
Language: python >=3.9

Runtime Dependencies (3):
  • requests >=2.31.0,<3.0.0
  • pyyaml >=6.0,<7.0
  • click >=8.0.0,<9.0.0

Development Dependencies (2):
  • pytest ^7.0.0
  • black >=23.0.0,<24.0.0
```

### PEP 621 Format (pyproject)

Output ready to copy into a PEP 621 `pyproject.toml`:

```bash
$ aigogo show-deps vendor/my-snippet --format pyproject

# Add these to your pyproject.toml

[project]
requires-python = ">=3.9"

[project.dependencies]
    "requests>=2.31.0,<3.0.0",
    "pyyaml>=6.0,<7.0",
    "click>=8.0.0,<9.0.0",

[project.optional-dependencies]
dev = [
    "pytest^7.0.0",
    "black>=23.0.0,<24.0.0",
]
```

### Poetry Format

Output ready to copy into a Poetry `pyproject.toml`:

```bash
$ aigogo show-deps vendor/my-snippet --format poetry

# Add these to your pyproject.toml

[tool.poetry.dependencies]
python = ">=3.9"
requests = ">=2.31.0,<3.0.0"
pyyaml = ">=6.0,<7.0"
click = ">=8.0.0,<9.0.0"

[tool.poetry.group.dev.dependencies]
pytest = "^7.0.0"
black = ">=23.0.0,<24.0.0"
```

### Requirements Format

Output ready for `requirements.txt`:

```bash
$ aigogo show-deps vendor/my-snippet --format requirements

# Runtime dependencies for requirements.txt
requests>=2.31.0,<3.0.0
pyyaml>=6.0,<7.0
click>=8.0.0,<9.0.0
```

### NPM Format (JavaScript)

Output ready to merge into `package.json`:

```bash
$ aigogo show-deps vendor/my-js-snippet --format npm

{
  "dependencies": {
    "express": "^4.18.0",
    "axios": "^1.6.0"
  },
  "devDependencies": {
    "jest": "^29.0.0"
  }
}
```

### Yarn Format (JavaScript)

Output as yarn add commands:

```bash
$ aigogo show-deps vendor/my-js-snippet --format yarn

yarn add express@^4.18.0
yarn add axios@^1.6.0
yarn add --dev jest@^29.0.0
```

---

## Common Workflows

### Workflow 1: Integrate Snippet into Existing Project

```bash
# Step 1: Install the snippet
aigogo add my-snippet:1.0.0
aigogo install

# Step 2: View dependencies in your project's format
# For PEP 621 (uv, modern setuptools):
aigogo show-deps vendor/my-snippet --format pyproject

# For Poetry:
aigogo show-deps vendor/my-snippet --format poetry

# For pip/requirements.txt:
aigogo show-deps .aigogo/imports/aigogo/my_snippet --format requirements > snippet-requirements.txt

# Step 3: Copy/paste the output into your pyproject.toml or requirements.txt
```

### Workflow 2: Review Dependencies Before Installing

```bash
# Check what dependencies a snippet needs before installing
aigogo pull my-snippet:1.0.0
aigogo show-deps ~/.aigogo/cache/my-snippet_1.0.0/

# Decide if you want to integrate it
aigogo add my-snippet:1.0.0
aigogo install
```

### Workflow 3: Generate Combined Requirements

```bash
# Install multiple snippets
aigogo add utils:1.0.0
aigogo add helpers:2.0.0
aigogo install

# Generate combined requirements.txt
aigogo show-deps .aigogo/imports/aigogo/utils --format requirements > requirements-snippets.txt
aigogo show-deps .aigogo/imports/aigogo/helpers --format requirements >> requirements-snippets.txt

# Install all
pip install -r requirements-snippets.txt
```

### Workflow 4: Use with Current Directory

```bash
# If you're already in a directory with aigogo.json
cd vendor/my-snippet
aigogo show-deps . --format pyproject

# Or just specify the file
aigogo show-deps aigogo.json --format poetry
```

---

## Examples

### Example 1: Basic Text Output

```bash
$ aigogo show-deps vendor/api-client

Package: api-client
Version: 1.0.0
Language: python >=3.8,<4.0

Runtime Dependencies (2):
  • requests >=2.31.0,<3.0.0
  • urllib3 >=2.0.0
```

### Example 2: Copy to pyproject.toml (PEP 621)

```bash
$ aigogo show-deps vendor/api-client --format pyproject > deps.txt
$ cat deps.txt

# Add these to your pyproject.toml

[project]
requires-python = ">=3.8,<4.0"

[project.dependencies]
    "requests>=2.31.0,<3.0.0",
    "urllib3>=2.0.0",
```

Then manually copy into your `pyproject.toml`:

```toml
[project]
name = "my-project"
version = "0.1.0"
requires-python = ">=3.8"
dependencies = [
    # ... your existing dependencies ...
    "requests>=2.31.0,<3.0.0",
    "urllib3>=2.0.0",
]
```

### Example 3: Poetry Project Integration

```bash
$ aigogo add langchain-utils:2.0.0 && aigogo install
$ aigogo show-deps .aigogo/imports/aigogo/langchain_utils --format poetry

# Add these to your pyproject.toml

[tool.poetry.dependencies]
python = ">=3.9"
langchain = ">=0.1.0,<0.2.0"
openai = ">=1.0.0,<2.0.0"

# Copy these lines into your existing [tool.poetry.dependencies] section
```

### Example 4: Create Separate Requirements File

```bash
$ aigogo add data-utils:1.5.0 && aigogo install
$ aigogo show-deps .aigogo/imports/aigogo/data_utils --format requirements > snippet-requirements.txt

# Now install with:
$ pip install -r snippet-requirements.txt
```

---

## Path Handling

The command accepts both files and directories:

```bash
# Directory path (looks for aigogo.json inside)
aigogo show-deps vendor/my-snippet
aigogo show-deps .

# Direct file path
aigogo show-deps vendor/my-snippet/aigogo.json
aigogo show-deps ./aigogo.json

# Absolute paths
aigogo show-deps /home/user/projects/snippet/aigogo.json
```

---

## Language Support

The `pyproject`, `poetry`, and `requirements` formats are only available for Python packages. The `npm` and `yarn` formats are only available for JavaScript/TypeScript packages. Attempting to use a format with the wrong language will result in an error:

```bash
$ aigogo show-deps go-utils --format pyproject
Error: pyproject format is only supported for Python packages (current language: go)

$ aigogo show-deps python-utils --format npm
Error: npm format is only supported for JavaScript packages (current language: python)
```

The `text` format (default) works with all languages.

---

## Error Handling

### File Not Found

```bash
$ aigogo show-deps nonexistent
Error: failed to access path: no such file or directory
```

### Invalid Format

```bash
$ aigogo show-deps . --format invalid
Error: unsupported format: invalid
Supported formats: text, pyproject, poetry, requirements, npm, yarn
```

### No aigogo.json in Directory

```bash
$ aigogo show-deps some-directory
Error: failed to load manifest: failed to read manifest: no such file or directory
```

---

## Tips

### Quick Copy to Clipboard

```bash
# Linux (X11)
aigogo show-deps . --format pyproject | xclip -selection clipboard

# macOS
aigogo show-deps . --format pyproject | pbcopy

# Windows (PowerShell)
aigogo show-deps . --format pyproject | Set-Clipboard
```

### Compare with Existing Dependencies

```bash
# Show what the snippet needs
aigogo show-deps vendor/snippet --format requirements

# Compare with your current requirements
diff <(aigogo show-deps vendor/snippet --format requirements) requirements.txt
```

### Batch Processing

```bash
# Show deps for all snippets in vendor/
for dir in vendor/*/ ; do
    echo "=== $dir ==="
    aigogo show-deps "$dir"
    echo
done
```

---

## Integration with Other Commands

### After `aigogo install`

```bash
aigogo add utils:1.0.0
aigogo install
aigogo show-deps .aigogo/imports/aigogo/utils --format pyproject  # View deps to add
```

### Before `aigogo build`

```bash
# Check what dependencies will be packaged
aigogo show-deps .
aigogo build
```

### With `aigogo list`

```bash
# List cached packages and show deps for one
aigogo list
aigogo show-deps ~/.aigogo/cache/my-package_1.0.0/
```

---

## See Also

- [ADD_COMMAND.md](ADD_COMMAND.md) - Add dependencies to aigogo.json
- [README.md](../README.md) - Main documentation

