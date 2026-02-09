# Version Specification Cheatsheet

Quick reference for dependency version constraints across ecosystems.

## Python (pip/uv/poetry)

### Quick Reference

| Constraint | Example | Allows | Use Case |
|------------|---------|--------|----------|
| `==` | `==2.31.0` | Only 2.31.0 | Exact reproducibility |
| `>=,<` | `>=2.31.0,<3.0.0` | 2.31.x, 2.32.x, 2.99.x | **Recommended for snippets** |
| `~=` | `~=2.31.0` | 2.31.0, 2.31.1, 2.31.2 | Patch updates only |
| `>=` | `>=2.31.0` | Any version ≥2.31.0 | Maximum flexibility (risky) |
| `!=` | `>=2.0,!=2.5.0` | Exclude specific version | Known broken version |

### Examples for aigogo Snippets

```toml
[project]
dependencies = [
    # RECOMMENDED: Constrain major version
    "requests>=2.31.0,<3.0.0",      # ✅ Best for most cases
    
    # MODERATE: Patch updates only
    "click~=8.1.0",                  # ✅ Good for stability
    
    # STRICT: Exact version
    "urllib3==2.1.0",                # ⚠️ Use for deployment only
    
    # FLEXIBLE: Minimum only
    "python-dateutil>=2.8.0",        # ✅ For very stable APIs
    
    # DANGEROUS: No constraint
    "pyyaml",                        # ❌ Avoid
]
```

## Node.js (npm/yarn/pnpm)

### Quick Reference

| Constraint | Example | Allows | Equivalent |
|------------|---------|--------|------------|
| `^` | `^1.6.0` | 1.6.0 to <2.0.0 | `>=1.6.0 <2.0.0` |
| `~` | `~4.17.21` | 4.17.x | `>=4.17.21 <4.18.0` |
| Exact | `4.18.2` | Only 4.18.2 | `==4.18.2` (Python) |
| Range | `>=1.6.0 <2.0.0` | Explicit range | Same as `^1.6.0` |
| `*` | `*` | Any version | ❌ Never use |

### Examples for aigogo Snippets

```json
{
  "dependencies": {
    "axios": "^1.6.0",          // ✅ Recommended
    "lodash": "~4.17.21",       // ✅ Stable
    "express": "4.18.2",        // ⚠️ Exact (strict)
    "react": ">=18.0.0 <19",    // ✅ Explicit range
    "next": "*"                 // ❌ Never do this
  }
}
```

## Go (go.mod)

```go
require (
    // Minimum version (MVS algorithm finds compatible set)
    github.com/pkg/errors v0.9.1
    
    // Major versions in path for v2+
    github.com/example/lib/v2 v2.3.4
    
    // Indirect dependencies
    golang.org/x/sync v0.6.0 // indirect
)
```

**Go's approach:** Minimum Version Selection (MVS)
- Specifies minimums, builds with minimum compatible set
- `go.sum` provides checksums
- Major versions in import path

## Rust (Cargo.toml)

```toml
[dependencies]
# Caret (default): ^1.0.0 = >=1.0.0 <2.0.0
serde = "1.0.195"           # ✅ Default behavior
tokio = "^1.35"             # ✅ Explicit caret
regex = "~1.10"             # Tilde: >=1.10.0 <1.11.0
clap = "=4.4.18"            # Exact version
```

## Comparison Matrix

### Same Meaning, Different Syntax

| Meaning | Python | Node.js | Rust | Go |
|---------|--------|---------|------|-----|
| Exact | `==2.31.0` | `2.31.0` | `=2.31.0` | `v2.31.0` |
| Minor updates OK | `~=2.31.0` | `~2.31.0` | `~2.31` | (use MVS) |
| Major version OK | `>=2.31,<3` | `^2.31.0` | `^2.31` or `2.31` | (use MVS) |
| Minimum | `>=2.31.0` | `>=2.31.0` | `>=2.31.0` | `v2.31.0` |

## Decision Tree

```
┌─ Need exact reproducibility?
│  └─ YES → Use exact: ==2.31.0 or 2.31.0
│  └─ NO  ↓
│
├─ Is API stable and backward compatible?
│  └─ YES → Use flexible: >=2.31.0,<3.0.0 or ^2.31.0
│  └─ NO  ↓
│
├─ Need security patches but minimal changes?
│  └─ YES → Use pessimistic: ~=2.31.0 or ~2.31.0
│  └─ NO  ↓
│
└─ Testing? Use range for flexibility
   or exact for reproducibility
```

## Real-World Examples

### Scenario 1: Web Framework

**Problem:** Need Flask, which has stable API

```toml
# pyproject.toml
dependencies = [
    "flask>=2.3.0,<3.0.0",     # Major version constraint
    "werkzeug>=3.0.0,<4.0.0",  # Flask dependency
]
```

**Why:** Flask 2.x API is stable, get bug fixes and features

### Scenario 2: HTTP Client

**Problem:** Need requests for API calls

```toml
dependencies = [
    "requests>=2.31.0,<3.0.0",  # requests 2.x is stable
]
```

**Why:** requests 2.x has been stable for years

### Scenario 3: CLI Tool

**Problem:** Need click for command-line interface

```toml
dependencies = [
    "click>=8.0.0,<9.0.0",  # click 8.x has stable API
]
```

**Why:** click follows semver, 8.x won't break

### Scenario 4: Deployment Script

**Problem:** Must work identically everywhere

```python
# requirements.txt (use pip freeze output)
boto3==1.34.34
botocore==1.34.34
click==8.1.7
requests==2.31.2
```

**Why:** Deployment = reproducibility > flexibility

## Common Mistakes

### ❌ Too Strict

```python
# Problem: Will conflict with almost everything
requests==2.31.2
click==8.1.7
```

**Fix:**
```python
requests>=2.31.0,<3.0.0
click>=8.1.0,<9.0.0
```

### ❌ Too Loose

```python
# Problem: Might get breaking changes
requests>=2.31.0
click
```

**Fix:**
```python
requests>=2.31.0,<3.0.0  # Add upper bound
click>=8.0.0,<9.0.0      # Add constraint
```

### ❌ No Python Version

```toml
[project]
dependencies = [
    "requests>=2.31.0,<3.0.0",
]
# Missing: requires-python
```

**Fix:**
```toml
[project]
requires-python = ">=3.8,<4.0"
dependencies = [
    "requests>=2.31.0,<3.0.0",
]
```

### ❌ Including Lock Files in Snippets

```json
{
  "files": [
    "code.py",
    "requirements.txt",
    "requirements-lock.txt"  // ❌ Don't include in reusable snippets
  ]
}
```

**Fix:**
```json
{
  "files": [
    "code.py",
    "requirements.txt"  // ✅ Flexible ranges only
  ]
}
```

## Testing Your Constraints

### Test Matrix

```bash
# Test minimum version
pip install requests==2.31.0
python -m pytest

# Test latest in range
pip install "requests>=2.31.0,<3.0.0"
python -m pytest

# Test with different Python versions
python3.8 -m pytest
python3.9 -m pytest
python3.12 -m pytest
```

### Automated Testing

```yaml
# .github/workflows/test.yml
strategy:
  matrix:
    python-version: ["3.8", "3.9", "3.10", "3.11", "3.12"]
    deps:
      - "requests==2.31.0"  # Minimum
      - "requests==2.31.2"  # Specific
      - "requests>=2.31"    # Latest
```

## Summary Recommendations

### For aigogo Snippet Packages (Libraries)

```toml
[project]
name = "my-snippet"
version = "1.0.0"
requires-python = ">=3.8,<4.0"

dependencies = [
    # Pattern: >=MINIMUM,<NEXT_MAJOR
    "package-name>=X.Y.0,<X+1.0.0",
]
```

**Why:** Maximum compatibility with user projects

### For Standalone Tools

```python
# requirements.txt (from pip freeze)
package-name==X.Y.Z
```

**Why:** Exact reproducibility for deployment

### General Rule

**Snippet creator → Flexible ranges**
```toml
"requests>=2.31.0,<3.0.0"  # Allow updates
```

**Snippet user → Lock after testing**
```bash
pip install -r requirements.txt
# Test everything works
pip freeze > requirements-lock.txt  # Lock for deployment
```

## Quick Copy-Paste Templates

### Python Package

```toml
[project]
name = "snippet-name"
version = "1.0.0"
requires-python = ">=3.8,<4.0"

dependencies = [
    "requests>=2.31.0,<3.0.0",
    "click>=8.0.0,<9.0.0",
]
```

### Node.js Package

```json
{
  "name": "snippet-name",
  "version": "1.0.0",
  "dependencies": {
    "axios": "^1.6.0",
    "lodash": "^4.17.21"
  }
}
```

### Go Package

```go
module github.com/user/snippet-name

go 1.22

require (
    github.com/pkg/errors v0.9.1
    golang.org/x/sync v0.6.0
)
```

---

**See DEPENDENCY_VERSIONS.md for complete details!**

