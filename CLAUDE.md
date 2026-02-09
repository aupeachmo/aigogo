# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**aigogo** is a lightweight code snippet package manager that uses Docker registries as a transport mechanism. It enables sharing and distributing reusable code snippets (not full packages) across projects. Key distinction: this manages CODE SNIPPETS, not dependencies.

## Build & Test Commands

```bash
# Build
make build              # Build for current platform (output: bin/aigogo)
make build-all          # Build for Linux (AMD64/ARM64), macOS (Intel/ARM), Windows

# Test
go test -v ./...        # Run all tests
go test -v ./pkg/manifest  # Run tests for specific package
go test -coverprofile=coverage.out ./...  # With coverage

# Lint
golangci-lint run --timeout=5m   # Same linter config as CI

# Other
go fmt ./...            # Format code
go vet ./...            # Check for issues
go mod tidy             # Manage dependencies
make install            # Install to /usr/local/bin (requires sudo)
make install-user       # Install to ~/bin (no sudo)
```

## Post-Edit Linting (IMPORTANT)

After modifying any `.go` file, **always** run the linter before considering the task complete:

```bash
# Install if not available
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

golangci-lint run --timeout=5m
```

This matches the CI lint job (`.github/workflows/test.yml` — `golangci/golangci-lint-action@v9`). Fix any errors before moving on. Common lint issues:
- `errcheck`: unchecked error return values (e.g. `.Close()`, `os.RemoveAll`) — use `_ = x.Close()` or check explicitly
- `staticcheck ST1005`: error strings should not be capitalized
- `govet`: suspicious constructs (shadowed variables, printf format mismatches)

## Architecture

### Entry Point
- `main.go` - Entry point with version injection via ldflags (`-X main.Version`)

### CLI Commands (`cmd/`)
20 commands built without external CLI framework. Key files:
- `root.go` - Command routing and argument parsing
- `add.go` - Add packages to lock file, or files/dependencies to manifest
- `install.go` - Install packages from aigogo.lock (creates symlinks)
- `uninstall.go` - Remove installed packages, .pth file, register.js, and .aigogo/ directory
- `build.go` - Local build with auto-versioning
- `push.go` - Push to registry (requires `--from` flag for local builds)

### Core Packages (`pkg/`)

**store/** - Content-Addressable Storage (CAS)
- `store.go` - Immutable package storage by SHA256 hash (~/.aigogo/store/)
- Packages stored at `~/.aigogo/store/sha256/<prefix>/<hash>/`
- Files made read-only after storage

**lockfile/** - Lock file management
- `lockfile.go` - Load/Save/Find aigogo.lock files
- Tracks package versions, integrity hashes, and sources
- `NormalizeName()` converts package names for Python (`my-utils` → `my_utils`)

**imports/** - Language-specific import setup
- `setup.go` - Creates `.aigogo/imports/` directory structure
- `pth.go` - Manages `.pth` file in Python site-packages for automatic path configuration
- `register.go` - Generates `.aigogo/register.js` for Node.js module resolution, resolves JS entry points
- Python namespace: `.aigogo/imports/aigogo/<package>/` with `__init__.py` (directory symlink to store)
- JavaScript scope: `.aigogo/imports/@aigogo/<package>/` (real dir with file symlinks + generated `package.json`)
- Auto-updates `.gitignore` to exclude `.aigogo/`

**manifest/** - Manifest (aigogo.json) handling
- `types.go` - Data structures: Manifest, Language, Dependencies, FileSpec
- `loader.go` - Load/Save/Validate manifest JSON
- `finder.go` - Find aigogo.json by walking up directory tree (like git)
- `discovery.go` - Auto-discover files by language patterns
- `ignore.go` - `.aigogoignore` file support (gitignore-compatible pattern matching)

**docker/** - Registry and local cache operations
- `local_builder.go` - Build packages to local cache (~/.aigogo/cache)
- `builder.go` - Create Docker image tar structures
- `extractor.go` - Extract files from cached packages
- `puller.go` / `pusher.go` - Registry pull/push operations
- `utils.go` - Image ref parsing, cache directory utilities, hash functions

**depgen/** - Dependency file generation
- `generator.go` - Generate requirements.txt, package.json, go.mod, Cargo.toml
- `scanner.go` - Scan source files for imports
- `validator.go` - Validate declared vs actual dependencies

**auth/** - Registry authentication
- Stores credentials in `~/.aigogo/auth.json` (mode 0600)
- Docker Hub OAuth2 token exchange support

### Key Design Patterns

1. **Content-Addressable Storage**: Packages stored by SHA256 hash for immutability
2. **Lock File Workflow**: `aigogo.lock` pins exact versions and hashes for reproducibility
3. **Namespace Imports**: Python uses `from aigogo.package_name`, JS uses `@aigogo/package-name`
4. **Local-First Workflow**: Build locally first, then push with explicit `--from` flag
5. **No Docker Daemon**: Local builds don't require Docker running
6. **Subdirectory Support**: All commands work from any subdirectory (finds aigogo.json upward)
7. **Auto-Versioning**: `aigogo build` without args increments patch version
8. **`.aigogoignore` Support**: Gitignore-compatible file exclusion
9. **AI Metadata**: Optional `ai` field in aigogo.json for agent discovery (see MACHINES.md)

### AI Agent Integration

- `pkg/manifest/types.go` defines `AISpec` struct with `summary`, `capabilities`, `usage`, `inputs`, `outputs`
- The `ai` field is optional on `Manifest` (`json:"ai,omitempty"`)
- `.claude/commands/aigogo.md` is a Claude Code slash command (`/aigogo`) for AI-assisted workflows
- See `MACHINES.md` for full documentation

### Directory Structure

```
~/.aigogo/
├── store/sha256/           # Content-addressable storage (new)
│   └── ab/abc123.../       # Package by hash
│       ├── files/          # Package files (read-only)
│       └── aigogo.json     # Package manifest
├── cache/                  # Legacy build/pull cache
│   ├── <name>_<version>/   # Local builds
│   └── images/             # Pulled images
└── auth.json               # Registry credentials

project/
├── aigogo.json             # Package manifest (for authors)
├── aigogo.lock             # Lock file (for consumers) - commit to git
├── .aigogo/                # Import links - gitignored
│   ├── .pth-location       # Tracks where aigogo.pth was installed
│   ├── register.js         # Node.js module resolution script
│   └── imports/
│       ├── aigogo/         # Python namespace
│       │   ├── __init__.py
│       │   └── my_utils/   # Symlink → store
│       └── @aigogo/        # JavaScript scope
│           └── my-utils/   # Real dir: file symlinks + package.json
└── .aigogoignore           # File exclusion patterns
```

### Lock File Format (aigogo.lock)

```json
{
  "version": 1,
  "packages": {
    "my_utils": {
      "version": "1.0.0",
      "integrity": "sha256:abc123...",
      "source": "docker.io/org/my-utils:1.0.0",
      "language": "python",
      "files": ["utils.py", "helpers.py"]
    }
  }
}
```

### Supported Languages
Python, JavaScript/TypeScript - fully supported with namespace imports.
Go, Rust - supported for package authoring (auto-discovery, dependency generation).

## Keeping Docs in Sync

When modifying `.go` files (especially `cmd/`), check and update:
- `cmd/completion.go` - Shell completions. Must list every command, subcommand, flag, and format alias.
- `qa/QA.md` - Command checklist. Must cover every command, flag, alias, and error case.
- `qa/run.sh` - Automated test harness. Must cover every command tested in QA.md.
- `README.md` - Command reference tables and usage examples.
- `CLAUDE.md` - Architecture section if commands/packages change.
- `MACHINES.md` - If `ai` metadata or examples change.
- `examples/README.md` - If examples or `show-deps` formats change.
- `.claude/commands/aigogo.md` - If author/consumer workflows change.
- `LANGUAGES.md` - If language support changes.
- `.github/README-CI.md`, `.github/RELEASE.md` - If build flags or release process changes.

## Dependencies

Minimal - only two external dependencies:
- `github.com/BurntSushi/toml` - TOML parsing for pyproject.toml
- `golang.org/x/term` - Terminal handling for password input
