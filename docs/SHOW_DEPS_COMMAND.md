# aigg show-deps Command

## Overview

The `show-deps` command displays dependencies from an `aigogo.json` manifest in various output formats. This is particularly useful when integrating aigogo agents into existing projects with their own dependency management systems.

## Usage

```bash
aigg show-deps <path> [--format <format>]
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
$ aigg show-deps vendor/my-snippet

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

Output ready to copy into a PEP 621 `pyproject.toml`. Dependencies are placed in aigogo-specific optional-dependency groups so they are clearly separated from your own project dependencies:

```bash
$ aigg show-deps vendor/my-snippet --format pyproject

# Add these to your pyproject.toml
# Install with: pip install -e '.[aigogo]' or pip install -e '.[aigogo,aigogo-dev]'

[project]
requires-python = ">=3.9"

[project.optional-dependencies]
aigogo = [
    "requests>=2.31.0,<3.0.0",
    "pyyaml>=6.0,<7.0",
    "click>=8.0.0,<9.0.0",
]
aigogo-dev = [
    "pytest^7.0.0",
    "black>=23.0.0,<24.0.0",
]
```

To install: `pip install -e '.[aigogo]'` (runtime only) or `pip install -e '.[aigogo,aigogo-dev]'` (with dev deps). To remove aigogo dependencies, delete the `aigogo` and `aigogo-dev` groups from `[project.optional-dependencies]`.

### Poetry Format

Output ready to copy into a Poetry `pyproject.toml`. Dependencies use dedicated aigogo groups:

```bash
$ aigg show-deps vendor/my-snippet --format poetry

# Add these to your pyproject.toml
# Install with: poetry install --with aigogo or poetry install --with aigogo,aigogo-dev

[tool.poetry.dependencies]
python = ">=3.9"

[tool.poetry.group.aigogo.dependencies]
requests = ">=2.31.0,<3.0.0"
pyyaml = ">=6.0,<7.0"
click = ">=8.0.0,<9.0.0"

[tool.poetry.group.aigogo-dev.dependencies]
pytest = "^7.0.0"
black = ">=23.0.0,<24.0.0"
```

To install: `poetry install --with aigogo` or `poetry install --with aigogo,aigogo-dev`. To remove, delete the `[tool.poetry.group.aigogo.*]` sections.

### Requirements Format

Output ready for `requirements.txt`:

```bash
$ aigg show-deps vendor/my-snippet --format requirements

# aigogo-managed runtime dependencies
# To remove: delete these entries and uninstall the packages
requests>=2.31.0,<3.0.0
pyyaml>=6.0,<7.0
click>=8.0.0,<9.0.0
```

### NPM Format (JavaScript)

Output ready to merge into `package.json`. Dependencies go into standard `dependencies`/`devDependencies` sections (for npm/yarn compatibility), plus an `aigogo` metadata key that tracks which deps are aigogo-managed:

```bash
$ aigg show-deps vendor/my-js-snippet --format npm

{
  "dependencies": {
    "express": "^4.18.0",
    "axios": "^1.6.0"
  },
  "devDependencies": {
    "jest": "^29.0.0"
  },
  "aigogo": {
    "managedDependencies": ["express", "axios"],
    "managedDevDependencies": ["jest"]
  }
}
```

To remove aigogo dependencies: delete the packages listed in `aigogo.managedDependencies` and `aigogo.managedDevDependencies` from the corresponding sections, then delete the `aigogo` key.

### Yarn Format (JavaScript)

Output as yarn add commands:

```bash
$ aigg show-deps vendor/my-js-snippet --format yarn

# aigogo-managed dependencies
# To remove: uninstall these packages and delete the "aigogo" key from package.json
yarn add "express@^4.18.0" "axios@^1.6.0"
yarn add --dev "jest@^29.0.0"
```

---

## Common Workflows

### Workflow 1: Integrate Snippet into Existing Project

```bash
# Step 1: Install the snippet
aigg add my-snippet:1.0.0
aigg install

# Step 2: View dependencies in your project's format
# For PEP 621 (uv, modern setuptools):
aigg show-deps vendor/my-snippet --format pyproject

# For Poetry:
aigg show-deps vendor/my-snippet --format poetry

# For pip/requirements.txt:
aigg show-deps .aigogo/imports/aigogo/my_snippet --format requirements > snippet-requirements.txt

# Step 3: Copy/paste the output into your pyproject.toml or requirements.txt
```

### Workflow 2: Review Dependencies Before Installing

```bash
# Check what dependencies a snippet needs before installing
aigg pull my-snippet:1.0.0
aigg show-deps ~/.aigogo/cache/my-snippet_1.0.0/

# Decide if you want to integrate it
aigg add my-snippet:1.0.0
aigg install
```

### Workflow 3: Generate Combined Requirements

```bash
# Install multiple snippets
aigg add utils:1.0.0
aigg add helpers:2.0.0
aigg install

# Generate combined requirements.txt
aigg show-deps .aigogo/imports/aigogo/utils --format requirements > requirements-snippets.txt
aigg show-deps .aigogo/imports/aigogo/helpers --format requirements >> requirements-snippets.txt

# Install all
pip install -r requirements-snippets.txt
```

### Workflow 4: Use with Current Directory

```bash
# If you're already in a directory with aigogo.json
cd vendor/my-snippet
aigg show-deps . --format pyproject

# Or just specify the file
aigg show-deps aigogo.json --format poetry
```

---

## Examples

### Example 1: Basic Text Output

```bash
$ aigg show-deps vendor/api-client

Package: api-client
Version: 1.0.0
Language: python >=3.8,<4.0

Runtime Dependencies (2):
  • requests >=2.31.0,<3.0.0
  • urllib3 >=2.0.0
```

### Example 2: Copy to pyproject.toml (PEP 621)

```bash
$ aigg show-deps vendor/api-client --format pyproject > deps.txt
$ cat deps.txt

# Add these to your pyproject.toml
# Install with: pip install -e '.[aigogo]' or pip install -e '.[aigogo,aigogo-dev]'

[project]
requires-python = ">=3.8,<4.0"

[project.optional-dependencies]
aigogo = [
    "requests>=2.31.0,<3.0.0",
    "urllib3>=2.0.0",
]
```

Then copy the `[project.optional-dependencies]` section into your `pyproject.toml` and install with `pip install -e '.[aigogo]'`. To remove aigogo deps later, just delete the `aigogo` group.

### Example 3: Poetry Project Integration

```bash
$ aigg add langchain-utils:2.0.0 && aigg install
$ aigg show-deps .aigogo/imports/aigogo/langchain_utils --format poetry

# Add these to your pyproject.toml
# Install with: poetry install --with aigogo or poetry install --with aigogo,aigogo-dev

[tool.poetry.dependencies]
python = ">=3.9"

[tool.poetry.group.aigogo.dependencies]
langchain = ">=0.1.0,<0.2.0"
openai = ">=1.0.0,<2.0.0"

# Copy these sections into your pyproject.toml, then: poetry install --with aigogo
```

### Example 4: Create Separate Requirements File

```bash
$ aigg add data-utils:1.5.0 && aigg install
$ aigg show-deps .aigogo/imports/aigogo/data_utils --format requirements > snippet-requirements.txt

# Now install with:
$ pip install -r snippet-requirements.txt
```

---

## Path Handling

The command accepts both files and directories:

```bash
# Directory path (looks for aigogo.json inside)
aigg show-deps vendor/my-snippet
aigg show-deps .

# Direct file path
aigg show-deps vendor/my-snippet/aigogo.json
aigg show-deps ./aigogo.json

# Absolute paths
aigg show-deps /home/user/projects/snippet/aigogo.json
```

---

## Language Support

The `pyproject`, `poetry`, and `requirements` formats are only available for Python packages. The `npm` and `yarn` formats are only available for JavaScript/TypeScript packages. Attempting to use a format with the wrong language will result in an error:

```bash
$ aigg show-deps go-utils --format pyproject
Error: pyproject format is only supported for Python packages (current language: go)

$ aigg show-deps python-utils --format npm
Error: npm format is only supported for JavaScript packages (current language: python)
```

The `text` format (default) works with all languages.

---

## Error Handling

### File Not Found

```bash
$ aigg show-deps nonexistent
Error: failed to access path: no such file or directory
```

### Invalid Format

```bash
$ aigg show-deps . --format invalid
Error: unsupported format: invalid
Supported formats: text, pyproject, poetry, requirements, npm, yarn
```

### No aigogo.json in Directory

```bash
$ aigg show-deps some-directory
Error: failed to load manifest: failed to read manifest: no such file or directory
```

---

## Tips

### Quick Copy to Clipboard

```bash
# Linux (X11)
aigg show-deps . --format pyproject | xclip -selection clipboard

# macOS
aigg show-deps . --format pyproject | pbcopy

# Windows (PowerShell)
aigg show-deps . --format pyproject | Set-Clipboard
```

### Compare with Existing Dependencies

```bash
# Show what the snippet needs
aigg show-deps vendor/snippet --format requirements

# Compare with your current requirements
diff <(aigg show-deps vendor/snippet --format requirements) requirements.txt
```

### Batch Processing

```bash
# Show deps for all snippets in vendor/
for dir in vendor/*/ ; do
    echo "=== $dir ==="
    aigg show-deps "$dir"
    echo
done
```

---

## Integration with Other Commands

### After `aigg install`

```bash
aigg add utils:1.0.0
aigg install
aigg show-deps .aigogo/imports/aigogo/utils --format pyproject  # View deps to add
```

### Before `aigg build`

```bash
# Check what dependencies will be packaged
aigg show-deps .
aigg build
```

### With `aigg list`

```bash
# List cached packages and show deps for one
aigg list
aigg show-deps ~/.aigogo/cache/my-package_1.0.0/
```

---

## See Also

- [ADD_COMMAND.md](ADD_COMMAND.md) - Add dependencies to aigogo.json
- [README.md](../README.md) - Main documentation

