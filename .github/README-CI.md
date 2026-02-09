# GitHub Actions Workflows

This directory contains GitHub Actions workflows for aigogo.

## Workflows

### 1. `release.yml` - Release Builds

**Trigger**: When a version tag is pushed (e.g., `v3.0.0`)

**Platforms**:
- Linux AMD64
- Linux ARM64
- macOS AMD64 (Intel)
- macOS ARM64 (Apple Silicon - M1/M2/M3)
- Windows AMD64
- Windows ARM64

**Process**:
1. Builds binaries for each platform (Linux ARM64 is cross-compiled)
2. Strips binaries for smaller size (Unix platforms)
3. Creates tarballs (Unix) or zips (Windows) with SHA256 checksums
4. Attests build provenance for each artifact (SLSA)
5. Creates a GitHub Release with all artifacts
6. Updates the Homebrew tap formula (`aupeach/homebrew-aigogo`) with new version and checksums via signed commit

**Usage**:
```bash
git tag -s v3.0.0 -m "Release v3.0.0"
git push origin v3.0.0
```

### 2. `build.yml` - Build Matrix Test

**Trigger**: Push to main/master, or pull requests

**Purpose**: Test that the code builds on multiple platforms

**Platforms tested**:
- Linux AMD64 (ubuntu-latest)
- Linux ARM64 (ubuntu-latest, cross-compiled)
- macOS AMD64 (macos-15)
- macOS ARM64 (macos-latest)
- Windows AMD64 (windows-latest)
- Windows ARM64 (windows-latest)

### 3. `test.yml` - Tests & Linting

**Trigger**: Push to main/master, or pull requests

**Actions**:
- Run Go tests with race detection
- Generate coverage report
- Build binary
- Run golangci-lint

### 4. `dependabot.yml` - Dependency Updates

**Schedule**: Daily

**Monitors**:
- Go modules (`go.mod`)
- GitHub Actions versions

## Release Process

See [RELEASE.md](RELEASE.md) for detailed release instructions.

**Quick release**:
```bash
# 1. Create and push tag (version is injected automatically via ldflags)
git tag -s v3.0.0 -m "Release v3.0.0"
git push origin v3.0.0

# 4. GitHub Actions automatically creates release
```

## Artifacts

Release artifacts include:
- `aigogo-linux-amd64.tar.gz` - Linux AMD64 binary (stripped)
- `aigogo-linux-arm64.tar.gz` - Linux ARM64 binary
- `aigogo-darwin-amd64.tar.gz` - macOS Intel binary (stripped)
- `aigogo-darwin-arm64.tar.gz` - macOS ARM64 binary (stripped)
- `aigogo-windows-amd64.zip` - Windows AMD64 binary
- `aigogo-windows-arm64.zip` - Windows ARM64 binary
- `*.sha256` - Checksums for verification

## Requirements

- `GITHUB_TOKEN` — Automatically provided. Used for creating releases and build attestations.
- `BREW_APP_ID` — GitHub App ID for the Homebrew tap updater. See [HOMEBREW.md](HOMEBREW.md) for setup.
- `BREW_APP_PRIVATE_KEY` — GitHub App private key for the Homebrew tap updater.

## Testing Locally

Build for all platforms locally:

```bash
# Linux AMD64
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o aigogo-linux-amd64
strip aigogo-linux-amd64

# Linux ARM64
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o aigogo-linux-arm64

# macOS AMD64 (Intel)
GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o aigogo-darwin-amd64
strip aigogo-darwin-amd64  # Only works on macOS

# macOS ARM64 (Apple Silicon)
GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o aigogo-darwin-arm64
strip aigogo-darwin-arm64  # Only works on macOS

# Windows AMD64
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o aigogo-windows-amd64.exe

# Windows ARM64
GOOS=windows GOARCH=arm64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o aigogo-windows-arm64.exe
```

## Badges

Add these to your README.md:

```markdown
[![Release](https://github.com/aupeach/aigogo/actions/workflows/release.yml/badge.svg)](https://github.com/aupeach/aigogo/actions/workflows/release.yml)
[![Build](https://github.com/aupeach/aigogo/actions/workflows/build.yml/badge.svg)](https://github.com/aupeach/aigogo/actions/workflows/build.yml)
[![Test](https://github.com/aupeach/aigogo/actions/workflows/test.yml/badge.svg)](https://github.com/aupeach/aigogo/actions/workflows/test.yml)
```

## Architecture Notes

aigogo v3.0+ uses:
- **Content-Addressable Storage (CAS)**: Packages stored by SHA256 hash at `~/.aigogo/store/`
- **Lock Files**: `aigogo.lock` for reproducible installs
- **Namespace Imports**: `from aigogo.package_name` (Python), `@aigogo/package-name` (JS)

Key commands:
- `aigogo add <package>` - Add package to lock file
- `aigogo install` - Install packages from lock file
- `aigogo build` - Build package locally
- `aigogo push` - Push to registry
