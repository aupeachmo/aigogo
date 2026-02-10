# Language Support

aigg fully supports Python and JavaScript/TypeScript for both authoring and consuming packages. Go and Rust have partial authoring support (manifest creation, file discovery, dependency scanning) but no consumer import infrastructure.

This document covers the two fully supported languages.

## Python

### Authoring

**File discovery**: `**/*.py`

**Dependency files**: `requirements.txt`, `pyproject.toml`

**Import scanning**: Detects `import x` and `from x import y`. Filters out relative imports and standard library modules (os, sys, re, json, time, datetime, collections, itertools, functools, pathlib, typing, abc, io, math, random, string, subprocess, threading, multiprocessing, logging, argparse, configparser, unittest, sqlite3, csv, xml, html, urllib, http, email, base64, hashlib).

**Dependency import**: `aigg add dep --from-pyproject` reads dependencies from an existing `pyproject.toml` and adds them to `aigogo.json`.

**Version constraints**: `==1.0.0` (exact), `>=1.0.0,<2.0.0` (range), `~=1.0.0` (compatible).

### Consumer

**Namespace**: `aigogo` — packages install to `.aigogo/imports/aigogo/<package_name>/`

**Name normalization**: Hyphens convert to underscores (`my-utils` becomes `my_utils`).

**Import syntax**:
```python
from aigogo.my_utils import some_function
```

**Path configuration**: `aigg install` writes an `aigogo.pth` file into the active Python environment's `site-packages` directory. This is the same mechanism `pip install -e` uses and works with system Python, venv, Poetry, and uv virtualenvs. The `.pth` file contains the absolute path to `.aigogo/imports/`, which Python adds to `sys.path` at startup.

Detection order:
1. `$VIRTUAL_ENV` environment variable (if set, finds `lib/pythonX.Y/site-packages/` within it)
2. `python3 -c "import sysconfig; print(sysconfig.get_path('purelib'))"` fallback

The `.pth` file location is tracked in `.aigogo/.pth-location` for cleanup by `aigg uninstall`.

**show-deps formats**:

| Format | Alias | Output |
|--------|-------|--------|
| `text` | | Human-readable summary |
| `requirements` | `pip` | `package>=1.0.0` (one per line) |
| `pyproject` | `pep621` | PEP 621 `[project.dependencies]` TOML |
| `poetry` | | `[tool.poetry.dependencies]` TOML |

## JavaScript / TypeScript

### Authoring

**File discovery**: `**/*.js`, `**/*.ts`, `**/*.jsx`, `**/*.tsx`, `**/*.mjs`, `**/*.cjs`

**Dependency file**: `package.json`

**Import scanning**: Detects `import ... from '...'` and `require('...')`. Filters out local imports (starting with `.` or `/`). No built-in stdlib exclusion list.

**Version constraints**: `1.0.0` (exact), `^1.0.0` (compatible), `~1.0.0` (patch-level), `>=1.0.0` (minimum).

### Consumer

**Scope**: `@aigogo` — packages install to `.aigogo/imports/@aigogo/<package-name>/`

**Import syntax**:
```javascript
const { fn } = require('@aigogo/my-utils');
// or
import { fn } from '@aigogo/my-utils';
```

**Entry point resolution**: When creating the package directory, aigg resolves the main entry point from top-level files only:

1. `index.js` (highest priority)
2. `index.mjs`
3. `index.cjs`
4. Single JS file (if only one exists)
5. First JS file alphabetically

If a package has no top-level `.js`/`.mjs`/`.cjs` files but has JS files in subdirectories, aigg prints a warning. Consumers can still use explicit paths (e.g. `require('@aigogo/pkg/sub/file')`).

**Path configuration**: `aigg install` creates `.aigogo/register.js`, which adds `.aigogo/imports/` to Node.js `NODE_PATH` and forces a path reload. Use it one of two ways:

```javascript
// At the top of your entry point
require('./.aigogo/register');
```

```bash
# Or as a preload flag
node --require ./.aigogo/register.js app.js
```

The register script is removed by `aigg uninstall`.

**show-deps formats**:

| Format | Alias | Output |
|--------|-------|--------|
| `text` | | Human-readable summary |
| `npm` | `package-json` | `{"dependencies": {...}}` JSON |
| `yarn` | | `yarn add "pkg@version"` commands |

## Implementation Checklist

When adding a new language:

1. **pkg/manifest/types.go** — Add to `SupportedLanguages()` list
2. **pkg/manifest/discovery.go** — Add file extension glob patterns
3. **pkg/depgen/generator.go** — Add dependency file generation
4. **pkg/depgen/scanner.go** — Add import scanning with a builtin/stdlib exclusion list. Follow the pattern of `pythonStdlib()` and `nodeBuiltins()`: create a `<lang>Builtins()` function returning `map[string]bool` of standard library module names, and filter them out in the scanner so they aren't reported as external dependencies. Add tests that verify builtins are excluded (see `TestScanJavaScriptBuiltins`)
5. **pkg/depgen/validator.go** — Add version constraint detection
6. **pkg/imports/setup.go** — Add namespace setup (directory structure, symlinks)
7. **cmd/install.go** — Add language detection and path configuration
8. **cmd/show_deps.go** — Add output format(s) for the language's package manager
9. **Tests** — Add tests for all new functions
10. **Documentation** — Update this file, README.md, CLAUDE.md, QA.md, qa/run.sh
