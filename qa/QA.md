# QA Checklist

Manual test checklist for every aigogo command and option.

## Docker QA Environment

Build and run a clean Ubuntu container with aigogo pre-installed:

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
mkdir /tmp/qa-aigogo && cd /tmp/qa-aigogo
```

## Author Commands

- [ ] `aigogo init` — creates aigogo.json
- [ ] `aigogo add file <path>` — adds file to manifest
- [ ] `aigogo add file <path> --force` — adds file even if ignored
- [ ] `aigogo add file <glob>` — adds multiple files via glob
- [ ] `aigogo add dep <pkg> <ver>` — adds runtime dependency
- [ ] `aigogo add dep --from-pyproject` — imports deps from pyproject.toml
- [ ] `aigogo add dev <pkg> <ver>` — adds dev dependency
- [ ] `aigogo add dev --from-pyproject` — imports dev deps from pyproject.toml
- [ ] `aigogo rm file <path>` — removes file from manifest
- [ ] `aigogo rm dep <pkg>` — removes runtime dependency
- [ ] `aigogo rm dev <pkg>` — removes dev dependency
- [ ] `aigogo scan` — auto-detects dependencies from source
- [ ] `aigogo validate` — checks declared deps match imports
- [ ] `aigogo build` — builds with auto-incremented version
- [ ] `aigogo build <name>:<tag>` — builds with explicit version
- [ ] `aigogo build --force` — rebuilds even if exists
- [ ] `aigogo build --no-validate` — skips dep validation

## Consumer Commands

- [ ] `aigogo add <name>:<tag>` — adds local package to lock file
- [ ] `aigogo add <registry>/<name>:<tag>` — adds remote package to lock file
- [ ] `aigogo install` — installs from aigogo.lock (creates symlinks)
- [ ] `aigogo install` — writes `.pth` file to Python site-packages (when Python packages present)
- [ ] `aigogo install` — creates `.aigogo/.pth-location` tracking file
- [ ] `aigogo install` — Python import works without manual PYTHONPATH
- [ ] `aigogo install` — falls back to PYTHONPATH hint when python3 unavailable
- [ ] `aigogo install` — JS packages get real directory with file symlinks (not directory symlink)
- [ ] `aigogo install` — JS packages get generated `package.json` with correct `main` entry point
- [ ] `aigogo install` — generates `.aigogo/register.js` when JS packages present
- [ ] `aigogo install` — JS `require('@aigogo/...')` works via register script

## Uninstall Command

- [ ] `aigogo uninstall` — removes `.aigogo/` directory
- [ ] `aigogo uninstall` — removes `.pth` file from Python site-packages
- [ ] `aigogo uninstall` — removes `register.js`
- [ ] `aigogo uninstall` — preserves `aigogo.lock`
- [ ] `aigogo uninstall` — prints nothing-to-uninstall when `.aigogo/` absent

## show-deps Formats

- [ ] `aigogo show-deps <path>` — text output (default)
- [ ] `aigogo show-deps <path> --format text` — explicit text
- [ ] `aigogo show-deps <path> --format requirements` — pip requirements.txt
- [ ] `aigogo show-deps <path> --format pip` — alias for requirements
- [ ] `aigogo show-deps <path> --format pyproject` — PEP 621 TOML
- [ ] `aigogo show-deps <path> --format pep621` — alias for pyproject
- [ ] `aigogo show-deps <path> --format poetry` — Poetry TOML
- [ ] `aigogo show-deps <path> --format npm` — package.json fragment
- [ ] `aigogo show-deps <path> --format package-json` — alias for npm
- [ ] `aigogo show-deps <path> --format yarn` — yarn add commands
- [ ] `aigogo show-deps <dir>` — accepts directory (finds aigogo.json)
- [ ] Python format on JS package → error
- [ ] JS format on Python package → error

## Cache Management

- [ ] `aigogo list` — shows cached packages
- [ ] `aigogo remove <name>:<tag>` — deletes from cache
- [ ] `aigogo remove-all` — prompts then deletes all
- [ ] `aigogo remove-all --force` — skips prompt

## Registry Commands

- [ ] `aigogo login <registry>` — interactive login
- [ ] `aigogo login <registry> -u <user>` — login with username
- [ ] `aigogo login <registry> -u <user> -p` — password from stdin
- [ ] `aigogo login --dockerhub` — Docker Hub shortcut
- [ ] `aigogo login ghcr.io` — GitHub Container Registry (PAT as password)
- [ ] `aigogo logout <registry>` — removes credentials
- [ ] `aigogo pull <registry>/<name>:<tag>` — pulls without installing
- [ ] `aigogo pull ghcr.io/<name>:<tag>` — pulls from ghcr.io (Basic auth)
- [ ] `aigogo push <registry>/<name>:<tag> --from <local>` — pushes to registry
- [ ] `aigogo push ghcr.io/<name>:<tag> --from <local>` — pushes to ghcr.io
- [ ] `aigogo delete <registry>/<name>:<tag>` — deletes from registry
- [ ] `aigogo delete <registry>/<name>:<tag> --all` — deletes all tags
- [ ] `aigogo search <term>` — searches registry (placeholder)

## Utilities

- [ ] `aigogo version` — prints version
- [ ] `aigogo completion bash` — bash completion script
- [ ] `aigogo completion zsh` — zsh completion script
- [ ] `aigogo completion fish` — fish completion script

## Error Cases

- [ ] No args → prints help
- [ ] Unknown command → error
- [ ] `aigogo build` with no aigogo.json → error
- [ ] `aigogo install` with no aigogo.lock → error
- [ ] `aigogo push` without `--from` → error
- [ ] `aigogo show-deps <path> --format invalid` → error listing valid formats
- [ ] `aigogo uninstall` outside any project → error

## Automated Test Harness

`qa/run.sh` exercises every command and error case listed above.

```bash
# Local-only mode (no registry tests)
bash qa/run.sh --local

# Full mode — prompts for registry, username, password, and repo namespace
bash qa/run.sh

# Use a custom binary
AIGOGO=/path/to/aigogo bash qa/run.sh --local
```

The script creates a temp workspace, runs every test, and prints a PASS/FAIL/SKIP summary. All command output is captured to a log file printed at the end. Registry tests are skipped when credentials are not provided.
