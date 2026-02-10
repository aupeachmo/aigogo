# aigg - Code Snippet Package Manager

You are helping the user work with aigg, a code snippet package manager that uses Docker registries as transport. Follow these instructions exactly.

## Determine Intent

The user may want to:
1. **Create a new package** from code in the current project
2. **Consume a package** by adding it to their project
3. **Build and publish** a package to a registry
4. **Troubleshoot** an issue with aigg

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
   - Respect any `.aigogoignore` file
5. Scan for dependencies: `aigg scan`
   - Review the output and add any detected dependencies: `aigg add dep <pkg> <version>`
   - For Python projects with `pyproject.toml`, use `aigg add dep --from-pyproject`
6. Validate: `aigg validate`
7. If the manifest has an `ai` field, populate it:
   - Write a clear `summary` describing what the snippet does
   - List `capabilities` as short action phrases
   - Add a `usage` example showing the primary import and function call
8. Build: `aigg build`

## Workflow: Consume a Package

1. Run `aigg add <reference>` where reference is either:
   - A registry path: `docker.io/org/package:tag`
   - A local cache reference: `package:tag`
2. Run `aigg install` to create import symlinks
3. Show the user how to import the package:
   - Python: `from aigogo.package_name import ...`
   - JavaScript: `require('@aigogo/package-name')` or `import ... from '@aigogo/package-name'`
4. Remind them to commit `aigogo.lock` to git

## Workflow: Build and Publish

1. Ensure `aigogo.json` is valid: `aigg validate`
2. Build locally: `aigg build <name>:<tag>` or just `aigg build` for auto-versioning
3. Test locally in another directory if needed
4. Login to registry: `aigg login <registry>`
5. Push: `aigg push <registry>/<name>:<tag> --from <name>:<tag>`

## AI Metadata

When creating or updating a package, check if the `aigogo.json` has an `ai` field. If not, offer to add one. The `ai` field helps AI agents discover and use the snippet:

```json
{
  "ai": {
    "summary": "One sentence describing what this snippet does and when to use it",
    "capabilities": ["verb phrase for each thing the code can do"],
    "usage": "from aigogo.package_name import function\nresult = function(args)"
  }
}
```

Keep the summary concise and action-oriented. Capabilities should be short phrases like "fetch JSON from URLs" or "format log entries as JSON". Usage should show the minimal import and call.

## Important Rules

- Always read existing files before modifying them
- Never push to a registry without the user explicitly confirming
- The `--from` flag is required for `aigg push`
- `aigogo.lock` should be committed to git; `.aigogo/` should be gitignored
- Package names are normalized: `my-utils` becomes `my_utils` in Python imports
