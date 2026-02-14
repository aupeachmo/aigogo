# QA Checklist

Manual test checklist for every aigg command and option.

## Docker QA Environment

Build and run a clean Ubuntu container with aigg pre-installed:

```bash
# From the repo root — build the Linux binary first, then the QA image
make build-linux          # or build-linux-arm for ARM hosts
docker build -f qa/Dockerfile -t aigogo-qa .

# Interactive shell
docker run --rm -it aigogo-qa

# Run the automated test harness (bind-mounted, not vendored)
docker run --rm -v "$(pwd)/qa/run.sh:/usr/local/bin/run-qa:ro" aigogo-qa bash run-qa --local

# Full mode with registry credentials
docker run --rm -it -v "$(pwd)/qa/run.sh:/usr/local/bin/run-qa:ro" aigogo-qa bash run-qa
```

## Setup (inside the container or locally)

```bash
mkdir /tmp/qa-aigg && cd /tmp/qa-aigg
```

## Author Commands

- [ ] `aigg init` — creates aigogo.json
- [ ] `aigg add file <path>` — adds file to manifest
- [ ] `aigg add file <path> --force` — adds file even if ignored
- [ ] `aigg add file <glob>` — adds multiple files via glob
- [ ] `aigg add dep <pkg> <ver>` — adds runtime dependency
- [ ] `aigg add dep --from-pyproject` — imports deps from pyproject.toml
- [ ] `aigg add dev <pkg> <ver>` — adds dev dependency
- [ ] `aigg add dev --from-pyproject` — imports dev deps from pyproject.toml
- [ ] `aigg rm file <path>` — removes file from manifest
- [ ] `aigg rm dep <pkg>` — removes runtime dependency
- [ ] `aigg rm dev <pkg>` — removes dev dependency
- [ ] `aigg scan` — auto-detects dependencies from source
- [ ] `aigg validate` — checks declared deps match imports
- [ ] `aigg build` — builds with auto-incremented version
- [ ] `aigg build <name>:<tag>` — builds with explicit version
- [ ] `aigg build --force` — rebuilds even if exists
- [ ] `aigg build --no-validate` — skips dep validation

## Diff Command

- [ ] `aigg diff` — working dir vs latest local build
- [ ] `aigg diff <name>:<tag>` — working dir vs specified build
- [ ] `aigg diff <ref-a> <ref-b>` — two local builds
- [ ] `aigg diff --remote <remote-ref>` — local build vs remote
- [ ] `aigg diff --remote <local-ref> <remote-ref>` — specific local vs remote
- [ ] `aigg diff --summary` — compact M/A/D output
- [ ] `aigg diff` with identical packages → "Packages are identical."

## Consumer Commands

- [ ] `aigg add <name>:<tag>` — adds local package to lock file
- [ ] `aigg add <registry>/<name>:<tag>` — adds remote package to lock file
- [ ] `aigg install` — installs from aigogo.lock (creates symlinks)
- [ ] `aigg install` — writes `.pth` file to Python site-packages (when Python packages present)
- [ ] `aigg install` — creates `.aigogo/.pth-location` tracking file
- [ ] `aigg install` — Python import works without manual PYTHONPATH
- [ ] `aigg install` — falls back to PYTHONPATH hint when python3 unavailable
- [ ] `aigg install` — JS packages get real directory with file symlinks (not directory symlink)
- [ ] `aigg install` — JS packages get generated `package.json` with correct `main` entry point
- [ ] `aigg install` — generates `.aigogo/register.js` when JS packages present
- [ ] `aigg install` — JS `require('@aigogo/...')` works via register script

## Uninstall Command

- [ ] `aigg uninstall` — removes `.aigogo/` directory
- [ ] `aigg uninstall` — removes `.pth` file from Python site-packages
- [ ] `aigg uninstall` — removes `register.js`
- [ ] `aigg uninstall` — preserves `aigogo.lock`
- [ ] `aigg uninstall` — prints nothing-to-uninstall when `.aigogo/` absent

## show-deps Formats

- [ ] `aigg show-deps <path>` — text output (default)
- [ ] `aigg show-deps <path> --format text` — explicit text
- [ ] `aigg show-deps <path> --format requirements` — pip requirements.txt
- [ ] `aigg show-deps <path> --format pip` — alias for requirements
- [ ] `aigg show-deps <path> --format pyproject` — PEP 621 TOML
- [ ] `aigg show-deps <path> --format pep621` — alias for pyproject
- [ ] `aigg show-deps <path> --format poetry` — Poetry TOML
- [ ] `aigg show-deps <path> --format npm` — package.json fragment
- [ ] `aigg show-deps <path> --format package-json` — alias for npm
- [ ] `aigg show-deps <path> --format yarn` — yarn add commands
- [ ] `aigg show-deps <dir>` — accepts directory (finds aigogo.json)
- [ ] Python format on JS package → error
- [ ] JS format on Python package → error

## Cache Management

- [ ] `aigg list` — shows cached packages
- [ ] `aigg remove <name>:<tag>` — deletes from cache
- [ ] `aigg remove-all` — prompts then deletes all
- [ ] `aigg remove-all --force` — skips prompt

## Registry Commands

- [ ] `aigg login <registry>` — interactive login
- [ ] `aigg login <registry> -u <user>` — login with username
- [ ] `aigg login <registry> -u <user> -p` — password from stdin
- [ ] `aigg login --dockerhub` — Docker Hub shortcut
- [ ] `aigg login ghcr.io` — GitHub Container Registry (PAT as password)
- [ ] `aigg logout <registry>` — removes credentials
- [ ] `aigg pull <registry>/<name>:<tag>` — pulls without installing
- [ ] `aigg pull ghcr.io/<name>:<tag>` — pulls from ghcr.io (Basic auth)
- [ ] `aigg push <registry>/<name>:<tag> --from <local>` — pushes to registry
- [ ] `aigg push <registry>/<name>:<tag> --from <local> --dry-run` — checks if push needed
- [ ] `aigg push ghcr.io/<name>:<tag> --from <local>` — pushes to ghcr.io
- [ ] `aigg delete <registry>/<name>:<tag>` — deletes from registry
- [ ] `aigg delete <registry>/<name>:<tag> --all` — deletes all tags
- [ ] `aigg search <term>` — searches registry (placeholder)

## Utilities

- [ ] `aigg version` — prints version
- [ ] `aigg completion bash` — bash completion script
- [ ] `aigg completion zsh` — zsh completion script
- [ ] `aigg completion fish` — fish completion script

## Error Cases

- [ ] No args → prints help
- [ ] Unknown command → error
- [ ] `aigg build` with no aigogo.json → error
- [ ] `aigg install` with no aigogo.lock → error
- [ ] `aigg push` without `--from` → error
- [ ] `aigg show-deps <path> --format invalid` → error listing valid formats
- [ ] `aigg uninstall` outside any project → error

## Automated Test Harness

`qa/run.sh` exercises every command and error case listed above.

```bash
# Local-only mode (no registry tests)
bash qa/run.sh --local

# Full mode — prompts for registry, username, password, and repo namespace
bash qa/run.sh

# Use a custom binary
AIGOGO=/path/to/aigg bash qa/run.sh --local
```

The script creates a temp workspace, runs every test, and prints a PASS/FAIL/SKIP summary. All command output is captured to a log file printed at the end. Registry tests are skipped when credentials are not provided.
