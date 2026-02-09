# aigogo

**aigogo** is a package manager for AI agents that uses Docker registries as a transport mechanism. Share, distribute, and reuse code packages across projects and agents without the overhead of full package ecosystems.

## Features

- 🚀 **Simple**: Use Docker registries to distribute packages
- 📦 **Lightweight**: No runtime dependencies, just pure file packaging
- 🔒 **Secure**: Leverage Docker registry authentication and encryption
- 🌐 **Universal**: Works with any Docker-compatible registry
- 🔄 **Version Control**: Use Docker tags for versioning
- ✅ **Dependency Management**: Automatically generate and display language-specific dependencies
- 🔍 **Validation**: Scan source files and validate dependencies
- 🔨 **Local-First**: Build and test locally before pushing to registries
- 📋 **Lock Files**: Reproducible installs with `aigogo.lock`
- 🔗 **Namespace Imports**: Use `from aigogo.package_name` (Python) or `@aigogo/package-name` (JS)

---

## Table of Contents

- [Quick Start](#quick-start)
- [Installation](#installation)
- [Workflows](#workflows)
  - [Using Packages (Consumer)](#using-packages-consumer)
  - [Creating Packages (Author)](#creating-packages-author)
  - [Local-Only Usage](#local-only-usage)
- [Command Reference](#command-reference)
- [aigogo.json Format](#aigogojson-format)
- [Development Setup](#development-setup)
- [License](#license)

---

## Quick Start

### 1. Install aigogo

```bash
# From source
git clone https://github.com/aupeach/aigogo.git
cd aigogo
make build
sudo make install  # Installs to /usr/local/bin
```

### 2. Use a Package (Consumer Workflow)

```bash
cd ~/my-project

# Add a package from Docker Hub
aigogo add docker.io/org/my-utils:1.0.0

# Or from GitHub Container Registry
aigogo add ghcr.io/org/my-utils:1.0.0

# Install packages (creates import symlinks)
aigogo install

# Use in Python
python -c "from aigogo.my_utils import helper; print(helper())"

# Use in JavaScript
node -e "const utils = require('@aigogo/my-utils'); console.log(utils)"
```

### 3. Create a Package (Author Workflow)

```bash
# Create a directory
mkdir my-api-utils && cd my-api-utils

# Create a Python file
cat > api_client.py <<'EOF'
import requests

def fetch_json(url):
    response = requests.get(url)
    response.raise_for_status()
    return response.json()
EOF

# Initialize, add files, build
aigogo init
aigogo add file api_client.py
aigogo add dep requests ">=2.31.0,<3.0.0"
aigogo build

# Push to Docker Hub (optional)
aigogo login docker.io
aigogo push docker.io/yourusername/api-utils:1.0.0 --from api-utils:1.0.0

# Or push to GitHub Container Registry
aigogo login ghcr.io
aigogo push ghcr.io/yourusername/api-utils:1.0.0 --from api-utils:1.0.0
```

---

## Installation

### From Binary Release

```bash
# Download latest release for your platform
# Linux (AMD64)
wget https://github.com/aupeach/aigogo/releases/latest/download/aigogo-linux-amd64.tar.gz
tar -xzf aigogo-linux-amd64.tar.gz
sudo mv aigogo-linux-amd64 /usr/local/bin/aigogo

# Linux (ARM64)
wget https://github.com/aupeach/aigogo/releases/latest/download/aigogo-linux-arm64.tar.gz
tar -xzf aigogo-linux-arm64.tar.gz
sudo mv aigogo-linux-arm64 /usr/local/bin/aigogo

# macOS (Intel)
wget https://github.com/aupeach/aigogo/releases/latest/download/aigogo-darwin-amd64.tar.gz
tar -xzf aigogo-darwin-amd64.tar.gz
sudo mv aigogo-darwin-amd64 /usr/local/bin/aigogo

# macOS (Apple Silicon)
wget https://github.com/aupeach/aigogo/releases/latest/download/aigogo-darwin-arm64.tar.gz
tar -xzf aigogo-darwin-arm64.tar.gz
sudo mv aigogo-darwin-arm64 /usr/local/bin/aigogo

# Verify
aigogo version
```

### From Source

```bash
git clone https://github.com/aupeach/aigogo.git
cd aigogo
make build
sudo make install  # Installs to /usr/local/bin + shell completion

# Or install to ~/bin (no sudo)
make install-user  # Also installs shell completion

# Installation automatically configures tab completion for bash/zsh
```

### Build for Multiple Platforms

```bash
make build-all
ls -lh bin/
```

---

## Workflows

### Using Packages (Consumer)

The recommended workflow for using aigogo packages in your project:

```bash
cd ~/my-project

# Step 1: Add packages to your project (Docker Hub or ghcr.io)
aigogo add docker.io/org/string-utils:1.0.0
aigogo add ghcr.io/org/api-client:2.0.0

# Step 2: Install packages (creates symlinks in .aigogo/)
aigogo install

# Step 3: Commit the lock file to version control
git add aigogo.lock
git commit -m "Add aigogo packages"

# Step 4: Use the packages in your code
```

**Python usage:**
```python
# aigogo install auto-configures your Python path via .pth file
from aigogo.string_utils import titlecase, reverse
from aigogo.api_client import fetch_json

print(titlecase("hello world"))
data = fetch_json("https://api.example.com/data")
```

**JavaScript usage:**
```javascript
require('./.aigogo/register'); // auto-configures module resolution

const stringUtils = require('@aigogo/string-utils');
const apiClient = require('@aigogo/api-client');

console.log(stringUtils.titlecase("hello world"));
```

**Team workflow:**
```bash
# When a teammate clones the project
git clone <repo>
cd <project>
aigogo install  # Recreates .aigogo/ from aigogo.lock
```

### Creating Packages (Author)

```bash
# Step 1: Create and initialize
mkdir my-utils && cd my-utils
aigogo init

# Step 2: Add your code
aigogo add file "*.py"              # Add Python files
aigogo add dep requests ">=2.31.0"  # Add dependencies

# Step 3: Validate and build
aigogo validate                     # Check dependencies match code
aigogo build                        # Build locally (auto-versions)

# Step 4: Test locally
cd ~/test-project
aigogo add ../my-utils/my-utils:0.1.1  # Add from local cache
aigogo install
python -c "from aigogo.my_utils import ..."

# Step 5: Share (optional - Docker Hub or ghcr.io)
aigogo login docker.io          # or: aigogo login ghcr.io
aigogo push docker.io/you/my-utils:0.1.1 --from my-utils:0.1.1
```

### Local-Only Usage

You don't need any Docker registry to use aigogo! Perfect for personal tools or private development.

```bash
# Machine 1: Create a utility package
mkdir ~/string-utils && cd ~/string-utils

cat > string_helpers.py <<'EOF'
def titlecase(text):
    return text.title()

def reverse(text):
    return text[::-1]
EOF

aigogo init
aigogo add file string_helpers.py
aigogo build string-utils:1.0.0

# Machine 1: Use in another project
cd ~/my-project
aigogo add string-utils:1.0.0  # Local reference (no registry)
aigogo install

python -c "from aigogo.string_utils import titlecase; print(titlecase('hello'))"
```

**Sharing locally without a registry:**
```bash
# Machine 1: Package the local build
cd ~/.aigogo/cache
tar -czf string-utils-1.0.0.tar.gz string-utils_1.0.0/

# Transfer via USB, network share, or scp
scp string-utils-1.0.0.tar.gz colleague@machine2:~

# Machine 2: Import the build
mkdir -p ~/.aigogo/cache && cd ~/.aigogo/cache
tar -xzf ~/string-utils-1.0.0.tar.gz

# Machine 2: Use it
cd ~/my-project
aigogo add string-utils:1.0.0
aigogo install
```

---

## Tab Completion

`aigogo` supports tab completion for bash, zsh, and fish shells.

### Bash

```bash
# Persistent (recommended)
aigogo completion bash | sudo tee /etc/bash_completion.d/aigogo > /dev/null
source ~/.bashrc

# Or add to ~/.bashrc
echo 'source <(aigogo completion bash)' >> ~/.bashrc
```

### Zsh

```bash
mkdir -p ~/.zsh/completions
aigogo completion zsh > ~/.zsh/completions/_aigogo
echo 'fpath=(~/.zsh/completions $fpath)' >> ~/.zshrc
echo 'autoload -Uz compinit && compinit' >> ~/.zshrc
exec zsh
```

### Fish

```bash
mkdir -p ~/.config/fish/completions
aigogo completion fish > ~/.config/fish/completions/aigogo.fish
```

---

## Command Reference

### Consumer Commands

```bash
# Add packages to aigogo.lock
aigogo add <registry/repo:tag>     # Add from registry
aigogo add <name:tag>              # Add from local cache

# Install packages from lock file
aigogo install                     # Creates .aigogo/imports/ with symlinks

# Remove installed packages and import configuration
aigogo uninstall                   # Removes .aigogo/, .pth file, register.js
```

### Author Commands

```bash
# Initialize and manage manifest
aigogo init                        # Create aigogo.json
aigogo add file <path>...          # Add files to package
aigogo add dep <pkg> <ver>         # Add runtime dependency
aigogo add dep --from-pyproject    # Import deps from pyproject.toml
aigogo add dev <pkg> <ver>         # Add dev dependency
aigogo rm file <path>...           # Remove files
aigogo rm dep <pkg>                # Remove dependency
aigogo scan                        # Auto-detect dependencies
aigogo validate                    # Verify dependencies match code

# Build and share
aigogo build                       # Build locally (auto-increments version)
aigogo build <name>:<tag>          # Build with explicit version
aigogo push <registry>/<name>:<tag> --from <local-name>:<tag>
```

### Cache Management

```bash
aigogo list                        # Show cached packages
aigogo remove <name>:<tag>         # Delete from local cache
aigogo remove-all                  # Delete all cached packages
```

### Registry Commands

```bash
aigogo login <registry>            # Authenticate (interactive)
aigogo login --dockerhub           # Login to Docker Hub
aigogo login ghcr.io               # Login to GitHub Container Registry
aigogo pull <registry/repo:tag>    # Pull without installing
aigogo delete <registry/repo:tag>  # Delete from registry
aigogo search <query>              # Search registry
```

**Supported registries:** Docker Hub (`docker.io`), GitHub Container Registry (`ghcr.io`), and any Docker V2-compatible registry.

> **ghcr.io tip:** Use a [Personal Access Token](https://github.com/settings/tokens) with `read:packages` / `write:packages` scope as your password when logging in.

### Utilities

```bash
aigogo show-deps <path>            # Show dependencies from aigogo.json
aigogo show-deps <path> --format pyproject  # Output in pyproject.toml format
aigogo show-deps <path> --format poetry     # Output in Poetry format
aigogo show-deps <path> --format requirements  # Output as requirements.txt
aigogo show-deps <path> --format npm        # Output as package.json fragment
aigogo show-deps <path> --format yarn       # Output as yarn add commands
aigogo version                     # Show version
aigogo completion <bash|zsh|fish>  # Generate shell completion
```

### Command Summary

| Command | Purpose | Example |
|---------|---------|---------|
| `add` | Add package or file/dep | `aigogo add docker.io/org/utils:1.0.0` |
| `install` | Install from lock file | `aigogo install` |
| `uninstall` | Remove imports & config | `aigogo uninstall` |
| `init` | Create aigogo.json | `aigogo init` |
| `build` | Package locally | `aigogo build` |
| `push` | Upload to registry | `aigogo push ghcr.io/me/utils:1.0.0 --from utils:1.0.0` |
| `list` | Show cached packages | `aigogo list` |
| `remove` | Delete from cache | `aigogo remove utils:1.0.0` |
| `validate` | Check dependencies | `aigogo validate` |
| `scan` | Find dependencies | `aigogo scan` |

---

## Project Structure

After running `aigogo install`, your project will have:

```
my-project/
├── aigogo.lock              # Lock file - COMMIT THIS
├── .aigogo/                 # Import links - GITIGNORED
│   ├── register.js          # Node.js path registration script
│   ├── .pth-location        # Tracks where aigogo.pth was installed
│   └── imports/
│       ├── aigogo/          # Python namespace
│       │   ├── __init__.py  # Namespace marker
│       │   ├── string_utils/  → ~/.aigogo/store/sha256/ab/abc.../files/
│       │   └── api_client/    → ~/.aigogo/store/sha256/cd/cde.../files/
│       └── @aigogo/         # JavaScript scope
│           ├── string-utils/  # Real dir with file symlinks + package.json
│           └── api-client/    # Real dir with file symlinks + package.json
├── .gitignore               # Contains: .aigogo/
└── your-code.py
```

**Global store structure:**
```
~/.aigogo/
├── store/sha256/            # Content-addressable storage
│   └── ab/abc123.../        # Package by hash
│       ├── files/           # Package files (read-only)
│       └── aigogo.json      # Package manifest
├── cache/                   # Build cache
└── auth.json                # Registry credentials
```

---

## aigogo.json Format

### Unified Manifest

```json
{
  "$schema": "https://aigg.sh/schema/v2.json",
  "name": "my-snippet",
  "version": "1.0.0",
  "description": "Description of your package",
  "author": "Your Name",
  "language": {
    "name": "python",
    "version": ">=3.8,<4.0"
  },
  "dependencies": {
    "runtime": [
      {"package": "requests", "version": ">=2.31.0,<3.0.0"}
    ],
    "dev": [
      {"package": "pytest", "version": ">=7.0.0"}
    ]
  },
  "files": {
    "include": "auto",
    "exclude": ["*.pyc", "__pycache__"]
  },
  "metadata": {
    "license": "MPL-2.0",
    "tags": ["http", "api", "client"]
  },
  "ai": {
    "summary": "Decorate Python functions to auto-generate OpenAI-compatible tool-calling schemas.",
    "capabilities": ["Generate tool schemas from type hints", "Dispatch tool calls by name"],
    "usage": "from aigogo.my_snippet import tool, get_tools\n\n@tool\ndef my_func(arg: str) -> str: ..."
  }
}
```

### aigogo.lock Format

```json
{
  "version": 1,
  "packages": {
    "my_utils": {
      "version": "1.0.0",
      "integrity": "sha256:abc123def456...",
      "source": "docker.io/org/my-utils:1.0.0",
      "language": "python",
      "files": ["utils.py", "helpers.py"]
    }
  }
}
```

### Auto-Discovery

Set `"files": {"include": "auto"}` and aigogo will automatically discover source files based on your language:

- **Python**: Finds all `.py` files
- **JavaScript/TypeScript**: Finds `.js`, `.ts`, `.jsx`, `.tsx`, `.mjs`, `.cjs`
- **Go**: Finds all `.go` files
- **Rust**: Finds all `.rs` files

### .aigogoignore File

Create a `.aigogoignore` file (like `.gitignore`) to exclude files:

```gitignore
# Build artifacts
dist/
build/

# Test files
tests/
*_test.py

# IDE files
.vscode/
.idea/
```

---

## Python Setup

`aigogo install` automatically configures your Python environment by writing an `aigogo.pth` file to your active Python's `site-packages` directory. This works with system Python, venv, Poetry, and uv virtualenvs — no manual setup required.

```python
# Just works after aigogo install
from aigogo.my_utils import helper
```

**Manual fallback** (if auto-configuration fails, e.g. python3 not installed):
```bash
export PYTHONPATH="$(pwd)/.aigogo/imports:$PYTHONPATH"
python your_script.py
```

## JavaScript Setup

`aigogo install` generates a register script at `.aigogo/register.js` that configures Node.js module resolution automatically.

**Option 1: Require in your entry point (CommonJS)**
```javascript
require('./.aigogo/register');

const { countTokens } = require('@aigogo/token-budget-js');
```

**Option 2: Preload flag (CommonJS and ESM)**
```bash
node --require ./.aigogo/register.js app.js
```

**Entry point resolution:** aigogo generates a `package.json` with a `main` field so that `require('@aigogo/pkg')` works. The entry point is resolved from top-level files only (priority: `index.js` > `index.mjs` > `index.cjs` > single file > first file alphabetically). Packages with JS files only in subdirectories will need explicit paths, e.g. `require('@aigogo/pkg/sub/file')`.

**Manual fallback** (if the register script approach doesn't fit your setup):
```bash
export NODE_PATH="$(pwd)/.aigogo/imports:$NODE_PATH"
node your_script.js
```

---

## Development Setup

### Prerequisites

- **Go 1.24+**: [Install Go](https://go.dev/doc/install)
- **Git**: For version control
- **Make**: For build automation

### Setup

```bash
git clone https://github.com/aupeach/aigogo.git
cd aigogo
go mod download
make build
./bin/aigogo version
```

### Project Structure

```
aigogo/
├── cmd/              # CLI commands
│   ├── root.go       # Command routing
│   ├── add.go        # Add packages/files/deps
│   ├── install.go    # Install from lock file
│   ├── build.go      # Build command
│   └── ...
├── pkg/              # Core packages
│   ├── store/        # Content-addressable storage
│   ├── lockfile/     # Lock file management
│   ├── imports/      # Import namespace setup
│   ├── docker/       # Registry interaction
│   ├── manifest/     # Manifest parsing
│   ├── depgen/       # Dependency generation
│   └── auth/         # Authentication
├── main.go           # Entry point
├── Makefile          # Build automation
└── go.mod            # Go module
```

### Running Tests

```bash
go test -v ./...
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

---

## License

MPL-2.0 - See [LICENSE](LICENSE) for details.

---

## AI Agent Integration

aigogo packages can include an optional `ai` field in `aigogo.json` that describes the package in terms AI agents can parse -- summary, capabilities, usage examples, and input/output descriptions. This enables agents to discover, evaluate, and use packages without reading source code.

A Claude Code skill (`/aigogo`) is also included for AI-assisted package creation and consumption.

See [MACHINES.md](MACHINES.md) for full documentation.

## Links

- **GitHub**: https://github.com/aupeach/aigogo
- **Issues**: https://github.com/aupeach/aigogo/issues
- **Releases**: https://github.com/aupeach/aigogo/releases
