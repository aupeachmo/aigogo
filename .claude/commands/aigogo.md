# aigg - AI Agent Manager

You are helping the user work with aigg, an AI agent manager that uses Docker registries as transport. Follow these instructions exactly.

## Determine Intent

The user may want to:
1. **Create a new package** from code in the current project
2. **Consume a package** by adding it to their project
3. **Build and publish** a package to a registry
4. **Manage packages** (list, remove, uninstall, show deps)
5. **Troubleshoot** an issue with aigg

Ask the user what they want to do if not clear from context.

## Workflow: Create a New Package

1. Check if `aigogo.json` already exists in the current directory or parent directories
2. If not, run `aigg init` to create one
3. Read the generated `aigogo.json` and update it:
   - Set `name` to something descriptive based on the code
   - Set `description` based on what the code does
   - Detect the correct `language.name` from the source files present
   - Set appropriate `language.version` constraints
4. Add source files: `aigg add file <paths...>`
   - Use glob patterns where appropriate (e.g. `"*.py"`)
   - Use `--force` to add files that match `.aigogoignore` patterns
   - Respect any `.aigogoignore` file (gitignore-compatible syntax)
5. Scan for dependencies: `aigg scan`
   - Review the output and add any detected dependencies: `aigg add dep <pkg> <version>`
   - Add dev dependencies with: `aigg add dev <pkg> <version>`
   - For Python projects with `pyproject.toml`, use `aigg add dep --from-pyproject` (or `aigg add dev --from-pyproject` for dev deps)
6. Remove files or deps if needed: `aigg rm file|dep|dev <name>`
7. Validate: `aigg validate`
8. If the manifest has an `ai` field, populate it (see AI Metadata below)
9. Build: `aigg build` (auto-increments patch version) or `aigg build <name>:<tag>`
   - Use `--force` to rebuild even if already cached
   - Use `--no-validate` to skip dependency validation

## Workflow: Consume a Package

1. Run `aigg add <reference>` where reference is either:
   - A registry path: `docker.io/org/package:tag`
   - A local cache reference: `package:tag`
2. Run `aigg install` to create import symlinks
3. Show the user how to import the package:
   - Python: `from aigogo.package_name import ...`
   - JavaScript: `require('@aigogo/package-name')` or `import ... from '@aigogo/package-name'`
   - For JS, remind them to add `require('./.aigogo/register')` at the top of their entry point
4. Remind them to commit `aigogo.lock` to git and add `.aigogo/` to `.gitignore`

## Workflow: Build and Publish

1. Ensure `aigogo.json` is valid: `aigg validate`
2. Build locally: `aigg build <name>:<tag>` or just `aigg build` for auto-versioning
3. Test locally in another directory if needed
4. Login to registry: `aigg login <registry>`
5. Push: `aigg push <registry>/<name>:<tag> --from <name>:<tag>`

## Workflow: Execute an Agent

Run an agent's entrypoint script directly (npx-like):

```bash
aigg exec <agent_name> [args...]
ENV_VAR=value aigg exec <agent_name> arg1 arg2
```

Prerequisites:
1. The agent must be in `aigogo.lock` (run `aigg add` + `aigg install` first)
2. The agent's `aigogo.json` must have a `scripts` field mapping command names to files
3. A compatible interpreter (Python/Node) must be available

On first run, `aigg exec` installs runtime dependencies into an isolated environment at `~/.aigogo/envs/<hash>/`. Subsequent runs skip installation.

## Workflow: Manage Packages

- **List cached packages**: `aigg list`
- **Remove from local cache**: `aigg remove <name:tag>`
- **Clear entire cache**: `aigg remove-all`
- **Clean cached data**: `aigg clean [--envs|--cache|--store|--all]`
- **Uninstall from project**: `aigg uninstall` (removes .aigogo/ directory, .pth file, register.js, exec envs)
- **Pull without installing**: `aigg pull <registry/name:tag>`
- **Delete from registry**: `aigg delete <registry/name:tag>`
- **Show dependencies**: `aigg show-deps <path> [--format text|pyproject|poetry|requirements|npm|yarn]`
- **Logout from registry**: `aigg logout <registry>`

## AI Metadata

When creating or updating a package, check if the `aigogo.json` has an `ai` field. If not, offer to add one. The `ai` field helps AI agents evaluate and use the package once they already have it:

```json
{
  "ai": {
    "summary": "One sentence describing what this agent does and when to use it",
    "capabilities": ["verb phrase for each thing the code can do"],
    "usage": "from aigogo.package_name import function\nresult = function(args)",
    "inputs": "Description of expected inputs and their types",
    "outputs": "Description of return values and their types"
  }
}
```

- **summary**: Concise and action-oriented. Answer "when would I use this?" not just "what is this?"
- **capabilities**: Short verb phrases like "fetch JSON from URLs" or "format log entries as JSON"
- **usage**: Minimal, copy-pasteable import and call example
- **inputs/outputs**: Optional but helpful for agents to evaluate fit without reading source

**Important limitation**: The `ai` field ships with the package but there is no discovery or search infrastructure. `aigg search` is a stub. An agent cannot query a registry for packages by capability. The `ai` field is only useful once a package is already locally available (in the store or cache). See [MACHINES.md](../../MACHINES.md#current-limitations) for details.

## Scripts (Exec Entrypoints)

When creating a package that should be executable via `aigg exec`, add a `scripts` field:

```json
{
  "scripts": {
    "my-agent": "run.py"
  }
}
```

- Keys are command names (typically matching the package name)
- Values are file paths that must be in `files.include`
- The script file should have `if __name__ == "__main__"` (Python) or be directly runnable by Node (JS)
- `aigg build` validates that script files exist in the package

## Important Rules

- Always read existing files before modifying them
- Never push to a registry without the user explicitly confirming
- The `--from` flag is required for `aigg push`
- `aigogo.lock` should be committed to git; `.aigogo/` should be gitignored
- Package names are normalized: `my-utils` becomes `my_utils` in Python imports
- All commands work from subdirectories (aigogo.json is found by walking up)
