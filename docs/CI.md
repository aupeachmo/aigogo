# CI/CD Guide

This guide covers using aigg in continuous integration and deployment pipelines. aigg is CI-friendly by design: it's a single static binary, requires no Docker daemon, and talks directly to registry APIs over HTTPS.

## Installing aigg in CI

aigg ships as a single binary with no runtime dependencies. Download it from GitHub Releases.

**GitHub Actions:**

```yaml
- name: Install aigg
  run: |
    curl -sL https://github.com/aupeachmo/aigogo/releases/latest/download/aigg-linux-amd64.tar.gz | tar xz
    sudo mv aigg-linux-amd64 /usr/local/bin/aigg
```

**GitLab CI:**

```yaml
before_script:
  - curl -sL https://github.com/aupeachmo/aigogo/releases/latest/download/aigg-linux-amd64.tar.gz | tar xz
  - mv aigg-linux-amd64 /usr/local/bin/aigg
```

**Pinning a version** (recommended for reproducibility):

```bash
curl -sL https://github.com/aupeachmo/aigogo/releases/download/v3.0.0/aigg-linux-amd64.tar.gz | tar xz
```

For ARM64 runners, replace `amd64` with `arm64`.

## Registry Authentication

For public packages, no authentication is needed. For private registries, use the `-u` and `-p` flags to log in non-interactively:

```yaml
- name: Registry login
  run: echo "$REGISTRY_TOKEN" | aigg login ghcr.io -u "$REGISTRY_USER" -p
```

The `-p` flag reads the password from stdin, keeping it out of shell history and process listings.

**GitHub Actions example with secrets:**

```yaml
- name: Registry login
  run: echo "${{ secrets.REGISTRY_TOKEN }}" | aigg login ghcr.io -u "${{ secrets.REGISTRY_USER }}" -p
```

**Docker Hub:**

```yaml
- name: Docker Hub login
  run: echo "${{ secrets.DOCKERHUB_TOKEN }}" | aigg login --dockerhub -u "${{ secrets.DOCKERHUB_USER }}" -p
```

## Installing Packages

The standard CI pattern: your repo has an `aigogo.lock` committed to git, and the pipeline runs `aigg install` to fetch and link packages.

```yaml
steps:
  - uses: actions/checkout@v4

  - name: Install aigg
    run: |
      curl -sL https://github.com/aupeachmo/aigogo/releases/latest/download/aigg-linux-amd64.tar.gz | tar xz
      sudo mv aigg-linux-amd64 /usr/local/bin/aigg

  - name: Install AI packages
    run: aigg install

  - name: Run tests
    run: pytest
```

After `aigg install`, packages are importable:

```python
# Python
from aigogo.my_utils import some_function
```

```javascript
// JavaScript (CommonJS)
require('./.aigogo/register');
const { fn } = require('@aigogo/my-utils');
```

## Caching the Store

aigg stores packages in `~/.aigogo/store/` by content hash (SHA256). This directory is an ideal CI cache target since its contents are immutable.

**GitHub Actions:**

```yaml
- name: Cache aigg store
  uses: actions/cache@v4
  with:
    path: ~/.aigogo/store
    key: aigg-${{ hashFiles('aigogo.lock') }}
    restore-keys: aigg-

- name: Install AI packages
  run: aigg install
```

When the cache hits, `aigg install` skips network fetches entirely and only creates the local symlinks.

**GitLab CI:**

```yaml
cache:
  key:
    files:
      - aigogo.lock
  paths:
    - .aigogo-store-cache/

variables:
  HOME_AIGOGO: .aigogo-store-cache

before_script:
  - mkdir -p "$HOME_AIGOGO"
  - ln -sfn "$(pwd)/$HOME_AIGOGO" ~/.aigogo/store
```

## Building Deployable Artifacts

When your CI builds a Docker image, Lambda zip, or other deployable artifact that needs aigg packages baked in, there are two approaches.

### Approach 1: Install in the build image

Run `aigg install` inside your Dockerfile. The resulting `.aigogo/` directory and store are self-contained within the image.

```dockerfile
FROM python:3.12-slim AS builder
# Install aigg
RUN curl -sL https://github.com/aupeachmo/aigogo/releases/latest/download/aigg-linux-amd64.tar.gz | tar xz \
    && mv aigg-linux-amd64 /usr/local/bin/aigg
# Copy lock file and install packages
WORKDIR /app
COPY aigogo.lock .
RUN aigg install

FROM python:3.12-slim
WORKDIR /app
# Copy the store and import links
COPY --from=builder /root/.aigogo/store /root/.aigogo/store
COPY --from=builder /app/.aigogo /app/.aigogo
# Copy your application
COPY . .
```

The key detail: both `~/.aigogo/store/` (where files live) and `.aigogo/` (where symlinks point) must be in the final image. The symlinks in `.aigogo/imports/` reference absolute paths into the store.

### Approach 2: Install in CI, copy into artifact

Run `aigg install` in the CI step, then include `.aigogo/` and the store in whatever artifact you produce.

```yaml
- name: Install AI packages
  run: aigg install

- name: Build artifact
  run: |
    mkdir -p dist
    cp -rL .aigogo/imports/ dist/aigogo-imports/
    # -L follows symlinks, producing real files in the artifact
    cp -r src/ dist/
    zip -r artifact.zip dist/
```

Using `cp -rL` resolves symlinks into real files, making the artifact self-contained without needing the store directory. Set `PYTHONPATH` or `NODE_PATH` to point at the copied imports directory at runtime.

## Full Example: GitHub Actions

```yaml
name: Build and Test
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Python
        uses: actions/setup-python@v5
        with:
          python-version: "3.12"

      - name: Install aigg
        run: |
          curl -sL https://github.com/aupeachmo/aigogo/releases/latest/download/aigg-linux-amd64.tar.gz | tar xz
          sudo mv aigg-linux-amd64 /usr/local/bin/aigg

      - name: Cache aigg store
        uses: actions/cache@v4
        with:
          path: ~/.aigogo/store
          key: aigg-${{ hashFiles('aigogo.lock') }}
          restore-keys: aigg-

      - name: Login to registry
        if: env.REGISTRY_TOKEN != ''
        run: echo "${{ secrets.REGISTRY_TOKEN }}" | aigg login ghcr.io -u "${{ secrets.REGISTRY_USER }}" -p
        env:
          REGISTRY_TOKEN: ${{ secrets.REGISTRY_TOKEN }}

      - name: Install AI packages
        run: aigg install

      - name: Install Python dependencies
        run: |
          # Option A: If your project has a pyproject.toml, add the aigogo group and install
          # aigg show-deps aigogo.lock --format pyproject >> pyproject.toml
          # pip install -e '.[aigogo]'
          #
          # Option B: Direct install from requirements format
          aigg show-deps aigogo.lock --format requirements > /tmp/agent-deps.txt
          pip install -r /tmp/agent-deps.txt

      - name: Run tests
        run: pytest
```

## Full Example: GitLab CI

```yaml
stages:
  - test

test:
  stage: test
  image: python:3.12-slim
  before_script:
    - curl -sL https://github.com/aupeachmo/aigogo/releases/latest/download/aigg-linux-amd64.tar.gz | tar xz
    - mv aigg-linux-amd64 /usr/local/bin/aigg
    - echo "$REGISTRY_TOKEN" | aigg login ghcr.io -u "$REGISTRY_USER" -p
  script:
    - aigg install
    - pip install -r requirements.txt
    - pytest
  cache:
    key:
      files:
        - aigogo.lock
    paths:
      - ~/.aigogo/store/
```

## CI Requirements Summary

| Requirement | Details |
|-------------|---------|
| **Docker daemon** | Not needed. aigg uses registry HTTP APIs directly. |
| **Network** | HTTPS to your Docker V2 registry (Docker Hub, ghcr.io, etc.). Not needed if store is cached and lock file hasn't changed. |
| **Python** | Only if installing Python packages. Needed to locate `site-packages` for the `.pth` file. If unavailable, set `PYTHONPATH` manually. |
| **Node.js** | Only if installing JavaScript packages. The generated `register.js` handles path setup. |
| **Filesystem** | Symlink support (standard on Linux/macOS CI runners). Write access to `~/.aigogo/`. |
| **Permissions** | No root required. The store lives under `$HOME`. |
| **Architecture** | Binaries available for Linux and macOS, both AMD64 and ARM64. |

## Python Path in CI

`aigg install` auto-configures the Python import path by writing a `.pth` file to `site-packages`. This works in most CI environments with Python installed. If it fails (e.g., read-only site-packages), aigg prints a warning and you can set the path manually:

```yaml
- name: Install AI packages
  run: |
    aigg install
    export PYTHONPATH="$(pwd)/.aigogo/imports:$PYTHONPATH"
```

## JavaScript Path in CI

For Node.js packages, `aigg install` generates `.aigogo/register.js`. Use it as a preload:

```yaml
- name: Run app
  run: node --require ./.aigogo/register.js app.js
```

Or in your test runner:

```json
{
  "scripts": {
    "test": "node --require ./.aigogo/register.js node_modules/.bin/jest"
  }
}
```

## Security Scanners

If your CI pipeline or registry has automated container scanning (Trivy, Snyk, Grype, Docker Scout, AWS ECR scanning), be aware that aigogo artifacts are **not runnable container images**. They contain only source code in a minimal Docker v2 manifest structure (empty `{}` config, single tar layer with source files). Scanners may produce false positives, scan failures, or dashboard noise.

**Recommended:** Push aigogo packages to a dedicated repository namespace (e.g., `ghcr.io/myorg/aigogo/`) and exclude that path from your scanner configuration. See [Security Scanners](SECURITY_SCANNERS.md) for detailed guidance and scanner-specific exclusion examples.

## Tips

- **Pin the aigg version** in CI for reproducible builds. Use a specific release URL instead of `latest`.
- **Cache aggressively.** The store is content-addressed, so cached entries never go stale. Only new packages require network fetches.
- **Commit `aigogo.lock` to git.** This is what makes `aigg install` reproducible across environments.
- **Don't commit `.aigogo/`.** It contains machine-specific symlinks and is regenerated by `aigg install`. The default `.gitignore` rule handles this.
- **Use `show-deps`** to generate language-native dependency files from your lock file, so your existing package manager (`pip`, `npm`) can install the agent's own dependencies.
