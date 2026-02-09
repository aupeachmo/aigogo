# `aigogo build` - Build Packages Locally

The `build` command creates snippet packages locally without pushing to a registry, enabling testing and offline development.

## Auto-Versioning (Recommended)

Run `aigogo build` without arguments to automatically increment the patch version:

```bash
aigogo build              # Reads aigogo.json, auto-increments, builds
```

**How it works:**
1. Reads `name` and `version` from `aigogo.json`
2. Increments patch version (e.g., `0.1.0` → `0.1.1`)
3. Builds the package
4. Updates `aigogo.json` with the new version

This is the recommended workflow for iterative development.

## Subdirectory Support

Like `git`, commands work from any subdirectory. aigogo searches up the directory tree to find `aigogo.json`:

```bash
my-project/
├── aigogo.json
└── src/
    └── utils/
        └── helper.py

$ cd src/utils
$ aigogo build            # ✓ Finds aigogo.json in ../..
```

## Usage

```bash
# Auto-increment and build (recommended)
aigogo build

# Or specify version explicitly
aigogo build <name>:<tag>

# Force rebuild
aigogo build --force

# Skip validation
aigogo build --no-validate
```

## Options

- `--force` - Force rebuild even if package already exists
- `--no-validate` - Skip dependency validation

## How It Works

Packages files into local cache directory (`~/.aigogo/cache/`) without requiring Docker or Podman.

**With auto-versioning:**

```bash
$ aigogo build

Auto-incrementing version: 0.1.0 -> 0.1.1
Building package: utils:0.1.1
Validating dependencies...
✓ Validation passed
Building to cache: /home/user/.aigogo/cache/utils_0.1.1
Packaging 3 file(s)...
  + api_client.py
  + utils.py
  + helpers.py
✓ Updated aigogo.json version to 0.1.1

✓ Successfully built utils:0.1.1

Next steps:
  Test locally:  aigogo add utils:0.1.1 && aigogo install
  Push to registry: aigogo push docker.io/myorg/utils
```

**With explicit version:**

```bash
$ aigogo build utils:2.0.0

Building package: utils:2.0.0
# ... (same output as above)
```

**Benefits:**
- ✅ No Docker/Podman required
- ✅ Fast build process
- ✅ Works offline
- ✅ Perfect for testing before pushing
- ✅ All builds in one location (`~/.aigogo/cache/`)

## Examples

### Basic Local Build (Auto-Versioning)

```bash
# 1. Create package
cd my-snippet
aigogo init

# 2. Add files and dependencies
aigogo add file api_client.py
aigogo add dep requests ">=2.31.0"

# 3. Build locally (auto-increments from 0.1.0 to 0.1.1)
aigogo build

# 4. Make changes and rebuild (0.1.1 -> 0.1.2)
aigogo build

# 5. Test it
cd ~/test-project
aigogo add mysnippet:0.1.2
aigogo install
python -c "from aigogo.mysnippet import ..."
```

### Explicit Version (Major/Minor Bumps)

```bash
# When you need to specify version explicitly
aigogo build mysnippet:2.0.0  # Major version bump
aigogo build mysnippet:1.1.0  # Minor version bump

# Note: Explicit versions don't update aigogo.json
```

### Build with Registry Prefix (Still Local!)

```bash
# This is still a LOCAL build
# The registry prefix is just part of the name
aigogo build docker.io/myorg/utils:1.0.0

⚠️  Note: Building with registry prefix - this is a local build
         To push to registry later, use: aigogo push docker.io/myorg/utils:1.0.0
```

### Force Rebuild

```bash
# First build
aigogo build utils:1.0.0
✓ Successfully built

# Try to build again
aigogo build utils:1.0.0
Error: package already exists in cache: utils:1.0.0
Use --force to rebuild

# Force rebuild
aigogo build utils:1.0.0 --force
✓ Successfully built (rebuilt)
```

### Skip Validation

```bash
# Build without validating dependencies
aigogo build utils:1.0.0 --no-validate
```

## Workflows

### 1. Test-Build-Push Workflow

```bash
# 1. Build locally
aigogo build utils:1.0.0

# 2. Test locally
cd ~/test-project
aigogo add utils:1.0.0
aigogo install

# 3. If tests pass, push to registry
aigogo push docker.io/myorg/utils:1.0.0 --from utils:1.0.0
```

### 2. Offline Development

```bash
# Work without internet
aigogo build mysnippet:dev

# Test
cd ~/test-project
aigogo add mysnippet:dev
aigogo install

# Make changes, rebuild
cd ~/my-snippet
aigogo build mysnippet:dev --force
```

### 3. CI/CD Pipeline

```yaml
# .github/workflows/test.yml
- name: Build snippet
  run: aigogo build myorg/utils:${GITHUB_SHA}

- name: Extract and test
  run: |
    aigogo add myorg/utils:${GITHUB_SHA}
    aigogo install

- name: Push if tests pass
  run: aigogo push docker.io/myorg/utils:${VERSION} --from myorg/utils:${GITHUB_SHA}
```

### 4. Multi-Version Testing

```bash
# Build multiple versions
aigogo build utils:1.0.0
aigogo build utils:2.0.0-beta
aigogo build utils:3.0.0-alpha

# List all
aigogo list

# Test each
for v in 1.0.0 2.0.0-beta 3.0.0-alpha; do
  mkdir test-$v && cd test-$v
  aigogo add utils:$v
  aigogo install
  cd ..
done
```

## Local vs Remote Naming

### Local References (No Registry)

```bash
# These are LOCAL references
aigogo build utils:1.0.0
aigogo build myorg/utils:1.0.0
aigogo build my-snippets:dev
```

**Characteristics:**
- No dots in first part (no domain)
- Stored in local cache only
- Not pushed anywhere

### Remote References (With Registry)

```bash
# These have registry prefixes
aigogo build docker.io/myorg/utils:1.0.0
aigogo build ghcr.io/myorg/utils:1.0.0
aigogo build registry.company.com/team/utils:1.0.0
```

**Note:** Even with registry prefix, `build` creates LOCAL builds!  
To push to the actual registry, use `aigogo push`.

## Integration with Other Commands

### With `add` + `install`

```bash
# Build → Add → Install
aigogo build utils:1.0.0
cd ~/test-project
aigogo add utils:1.0.0   # Adds from local cache
aigogo install            # Creates import links
```

### With `list`

```bash
# Build → List
aigogo build utils:1.0.0
aigogo list

# Output:
# 🔨 utils:1.0.0
#    Type: local build
#    Time: just now
#    Size: 15.2 KB
```

### With `push`

```bash
# Build → Push
aigogo build utils:1.0.0
aigogo push docker.io/myorg/utils:1.0.0 --from utils:1.0.0
```

### With `remove`

```bash
# Build → Remove
aigogo build utils:test
aigogo remove utils:test  # Clean up test build
```

## Error Handling

### No aigogo.json

```bash
$ aigogo build utils:1.0.0
Error: failed to load aigogo.json
Run 'aigogo init' first
```

**Solution:**
```bash
aigogo init
# Edit aigogo.json
aigogo build utils:1.0.0
```

### Already Exists

```bash
$ aigogo build utils:1.0.0
Error: package already exists in cache: utils:1.0.0
Use --force to rebuild
```

**Solution:**
```bash
aigogo build utils:1.0.0 --force
```

### No Files Found

```bash
$ aigogo build utils:1.0.0
Error: no files to package
```

**Solution:**
- Check `files.include` in `aigogo.json`
- Ensure source files exist
- Try `aigogo validate` to see what's detected


## Performance

Building to local cache is fast and requires no external dependencies:

- ⚡ **Fast**: Direct file operations, no Docker overhead
- 💾 **Efficient**: Minimal disk space usage
- 🔌 **Offline**: Works without internet or Docker daemon
- 🔄 **Consistent**: Same behavior on all platforms

## Best Practices

### 1. Use Local Builds for Development

```bash
# Fast iteration
aigogo build mysnippet:dev --force
aigogo add mysnippet:dev && aigogo install
# Test, modify, repeat
```

### 2. Use Meaningful Tags

```bash
# Good
aigogo build utils:1.0.0
aigogo build utils:2.0.0-beta
aigogo build utils:dev
aigogo build utils:${GIT_SHA}

# Avoid
aigogo build utils:1
aigogo build utils:latest  # Ambiguous
```

### 3. Clean Up Test Builds

```bash
# After testing
aigogo remove utils:test
aigogo remove utils:experiment
```

### 4. Validate Before Building

```bash
# Check first
aigogo validate
aigogo scan

# Then build
aigogo build utils:1.0.0
```

### 5. Use --force Judiciously

```bash
# During development: OK
aigogo build mysnippet:dev --force

# For releases: Avoid (build fresh)
rm -rf ~/.aigogo/cache/utils_1.0.0
aigogo build utils:1.0.0
```

## Advanced Usage

### Build Multiple Variants

```bash
# Different configurations
aigogo build utils:minimal --no-validate
aigogo build utils:full
aigogo build utils:debug
```

### Scripted Builds

```bash
#!/bin/bash
VERSIONS=("1.0.0" "1.0.1" "1.1.0")

for ver in "${VERSIONS[@]}"; do
  echo "Building $ver..."
  aigogo build utils:$ver --force

  # Test each
  mkdir -p test-$ver && cd test-$ver
  aigogo add utils:$ver
  aigogo install
  cd ..
done
```

### Build from Different Directories

```bash
# Build snippet A
cd /path/to/snippet-a
aigogo build snippet-a:1.0.0

# Build snippet B
cd /path/to/snippet-b
aigogo build snippet-b:1.0.0

# Both are in cache
aigogo list
```

## Troubleshooting

### Build is Slow

**Check:**
- Are you using `--docker` or `--podman`? (slower than local)
- Large files being packaged?
- Run `aigogo list` to see sizes

**Solution:**
- Use local build (default, no flags)
- Exclude unnecessary files in `aigogo.json`

### Wrong Files Packaged

**Check:**
```bash
aigogo build utils:test
aigogo add utils:test && aigogo install
ls -la .aigogo/imports/  # Check what got installed
```

**Solution:**
- Update `files.include` in `aigogo.json`
- Use `aigogo scan` to see what's detected

### Can't Find Built Package

```bash
$ aigogo add utils:1.0.0
Error: local image not found: utils:1.0.0
```

**Check:**
```bash
aigogo list  # See what's actually built
```

**Common issues:**
- Typo in name/tag
- Built with different name
- Removed from cache

## Summary

**Command**: `aigogo build <name>:<tag>`

**Purpose**: Build snippet packages locally for testing

**How it works**:
- Builds to local cache (`~/.aigogo/cache/`)
- No Docker or Podman required
- Fast, offline-capable, cross-platform

**Flags**:
- `--force` - rebuild if exists
- `--no-validate` - skip validation

**Perfect for**:
- ✅ Testing before pushing
- ✅ Offline development
- ✅ CI/CD pipelines
- ✅ Local iteration

**Next steps after building**:
- Test: `aigogo add <name>:<tag>` + `aigogo install`
- Push: `aigogo push <registry>/<name>:<tag> --from <name>:<tag>`
- List: `aigogo list`

**Note**: When you push to a registry, aigogo automatically creates Docker-compatible image format. You don't need Docker installed locally to use aigogo!

