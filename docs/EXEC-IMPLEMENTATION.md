# exec & clean — Implementation Details

This document describes everything implemented for `aigg exec` and `aigg clean`, including the design decisions, file changes, and dependency isolation strategy.

## Overview

`aigg exec` brings an `npx`-like workflow to aigogo: run an agent's entrypoint script directly without manually setting up interpreters or dependencies. `aigg clean` provides disk usage visibility and cleanup for cached state.

## Files Created

| File | Purpose |
|------|---------|
| `cmd/exec.go` | The `exec` command: package resolution, interpreter discovery, dependency env setup, script execution |
| `cmd/clean.go` | The `clean` command: disk usage summary, `--envs`/`--cache`/`--store`/`--all` flags |
| `cmd/exec_test.go` | Unit tests for version parsing, version constraint checking, env path, setEnv helper |
| `cmd/clean_test.go` | Unit tests for dirStats and cleanDirectory |
| `docs/exec-quickstart.md` | User-facing quickstart guide |
| `docs/exec-implementation.md` | This file |

## Files Modified

| File | Change |
|------|--------|
| `pkg/manifest/types.go` | Added `Scripts map[string]string` field to `Manifest` struct |
| `pkg/manifest/loader.go` | Added validation for `Scripts` (no empty names or file paths) |
| `pkg/manifest/loader_test.go` | Added test cases for scripts validation |
| `pkg/docker/local_builder.go` | Added build-time validation that script files exist in packaged files |
| `cmd/root.go` | Registered `exec` and `clean` commands, added to usage order |
| `cmd/uninstall.go` | Extended to remove exec environments for packages in the lock file |
| `cmd/completion.go` | Added `exec` and `clean` to bash, zsh, and fish completions with flags |
| `README.md` | Added `exec` and `clean` to command reference |
| `CLAUDE.md` | Updated architecture section |
| `MACHINES.md` | Noted exec as a new capability |
| `qa/QA.md` | Added test checklists for exec and clean |
| `qa/run.sh` | Added automated tests for exec and clean |
| `.claude/commands/aigogo.md` | Added exec and clean workflows |
| `examples/README.md` | Mentioned exec capability |

## Design Decisions

### Manifest Schema: `scripts` Field

```json
{
  "scripts": {
    "my-agent": "run.py"
  }
}
```

- Keys are command names (typically matching the package name)
- Values are file paths relative to the package root
- File paths only — no `module:function` notation (simpler, language-agnostic)
- Build-time validation ensures referenced files are in the package
- Multiple scripts per package are supported but uncommon

Why `aigogo.json` instead of `pyproject.toml` / `package.json`? Consistency across languages. One source of truth regardless of Python, JavaScript, or future languages.

### Interpreter Resolution (Shim + Validation)

Rather than bundling interpreters or creating binaries (PyInstaller), `exec` uses whatever Python/Node is available on the system and validates it meets the version constraint.

**Python search order:**
1. `$VIRTUAL_ENV/bin/python` — respects active venv
2. `./venv/bin/python`, `./.venv/bin/python` — project-local venvs
3. `python3` on `$PATH`
4. `python` on `$PATH`

**Node:** `node` on `$PATH`.

Version validation parses constraints like `>=3.8,<4.0` and compares against the interpreter's reported version. If no interpreter satisfies the constraint, a clear error is printed.

### Dependency Isolation

Exec environments live at `~/.aigogo/envs/<hash>/`, keyed by the package's SHA256 integrity hash from the lock file.

```
~/.aigogo/envs/
└── abc123def456.../
    ├── .venv/              # Python: isolated virtual environment
    │   └── lib/python3.x/site-packages/...
    ├── requirements.txt    # Generated from aigogo.json dependencies
    ├── node_modules/       # Node: npm-installed dependencies
    └── package.json        # Generated from aigogo.json dependencies
```

**Why outside the store?** The store (`~/.aigogo/store/`) is immutable and content-addressable. Files are made read-only via `MakeReadOnly()`. Exec environments are mutable derived state (installed packages change with platform/Python version), so they must live separately.

**Why keyed by hash?** Two projects using the same package version share one environment. Updating the package (new hash) gets a fresh env automatically.

**Python env creation:**
1. Try `uv venv` + `uv pip install` (fastest, modern)
2. Fall back to `python3 -m venv` + `<venv>/bin/pip install`

The venv's own pip is used (not system pip), avoiding permission issues. If `python3 -m venv` fails (missing `ensurepip`), a clear error suggests installing `python3-venv` or `uv`.

**Node env creation:**
1. Generate `package.json` with dependencies from `aigogo.json`
2. Run `npm install --prefix <env_dir>`

### Script Execution

The script runs via `syscall.Exec`, which replaces the current process. This ensures:
- Signals (SIGTERM, SIGINT) reach the script directly
- Exit codes propagate correctly
- stdin/stdout/stderr are not proxied (zero overhead)
- Environment variables from the caller naturally inherit

**Python execution:**
```
PYTHONPATH=<store>/files <venv>/bin/python <store>/files/<script> [args...]
```
When a venv exists, the venv's Python is used (which has access to installed site-packages). `PYTHONPATH` includes the files directory so intra-package imports work.

**Node execution:**
```
NODE_PATH=<store>/files:<env>/node_modules node <store>/files/<script> [args...]
```

### Clean Command

`aigg clean` with no flags shows a disk usage summary:

```
aigogo disk usage:

  Exec environments:       340MB  (12 items)   aigg clean --envs
  Build/pull cache:        180MB  (8 items)    aigg clean --cache
  Package store:            95MB  (15 items)   aigg clean --store

  Total: 615MB

Use aigg clean --all to remove everything
```

Flags: `--envs`, `--cache`, `--store`, `--all`. Each removes the corresponding `~/.aigogo/` subdirectory.

### Uninstall Extension

`aigg uninstall` now also removes exec environments for packages listed in the project's `aigogo.lock`. This is safe because envs are derived state — they recreate lazily on next `aigg exec`.

## Version Constraint Parser

The constraint parser (`cmd/exec.go`) handles PEP 440-style and semver-style constraints:

- Operators: `>=`, `<=`, `>`, `<`, `==`, `!=`
- Compound: `>=3.8,<4.0` (comma-separated, all must match)
- Version parsing: strips `v` prefix, handles pre-release suffixes (e.g. `3.12.0rc1` parses as `3.12.0`)
- Missing segments treated as zero: `3.8` equals `3.8.0`

## Test Coverage

**Unit tests (`cmd/exec_test.go`):**
- `TestParseVersion` — version string to int array
- `TestCompareVersions` — version comparison semantics
- `TestCheckVersionConstraint` — compound constraint evaluation
- `TestEvaluateConstraint` — individual operator evaluation
- `TestSetEnv` — environment variable manipulation
- `TestEnvPath` — env directory path construction

**Unit tests (`cmd/clean_test.go`):**
- `TestDirStats` — directory size and item counting
- `TestCleanDirectory` — directory removal

**Manifest tests (`pkg/manifest/loader_test.go`):**
- `valid scripts` — scripts field accepted
- `script with empty name` — rejected
- `script with empty file` — rejected

**QA tests (`qa/run.sh`):**
- Build a package with scripts, exec it, verify output
- Clean commands with all flag variants
- Error cases for missing agents and no-scripts packages
