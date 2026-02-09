# Release Process

## How Version Injection Works

aigogo uses **linker flags** to inject the version at build time - no manual version updates needed.

### Build-Time Injection

```bash
# Makefile uses git describe
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "0.0.1")
go build -ldflags="-s -w -X main.Version=$(VERSION)" .

# GitHub Actions extracts from tag
go build -ldflags="-s -w -X main.Version=v3.0.0" .
```

### Version Flow

```
Git Tag (v3.0.0)
    ↓
Build: -X main.Version=v3.0.0
    ↓
main.go: Version = "v3.0.0"
    ↓
User runs: ./aigogo version
    ↓
Output: "aigogo version v3.0.0"
```

### Local Development Builds

```bash
$ make build
$ ./aigogo version
aigogo version v3.0.0-5-g4f8a2c1-dirty
```

The version shows:
- `v3.0.0` - Last git tag
- `-5` - 5 commits since tag
- `-g4f8a2c1` - Git commit hash
- `-dirty` - Uncommitted changes

### Fallback

If no git tags exist or building without Makefile:
```bash
$ go build .
$ ./aigogo version
aigogo version 0.0.1  # Fallback default
```

## Creating a New Release

### 1. Create and Push Tag

```bash
# Using the release script (recommended)
./release.sh

# Or manually (creates a GPG-signed tag)
git tag -s v3.0.0 -m "Release v3.0.0"
git push origin v3.0.0
```

### 3. GitHub Actions

The release workflow will automatically:
- Build binaries for Linux (AMD64/ARM64), macOS (AMD64/ARM64), and Windows (AMD64/ARM64)
- Strip binaries for smaller size (Unix platforms)
- Create tarballs/zips with checksums
- Generate build provenance attestations for each artifact
- Create a GitHub Release
- Upload all artifacts
- Update the Homebrew tap formula (`aupeachmo/homebrew-aigogo`) with new version and checksums

### 4. Verify Release

Check https://github.com/aupeachmo/aigogo/releases

## Release Checklist

- [ ] All tests passing locally (`go test ./...`)
- [ ] Documentation reviewed
- [ ] GPG-signed tag created and pushed (`./release.sh`)
- [ ] GitHub Actions workflow succeeded
- [ ] Release appears on GitHub with "Verified" tag badge
- [ ] Binaries downloadable and working
- [ ] `./aigogo version` shows correct version
- [ ] `gh attestation verify <artifact> --repo aupeachmo/aigogo` passes
- [ ] Homebrew tap formula updated (check `aupeachmo/homebrew-aigogo` for verified commit)

## Manual Release (if needed)

If GitHub Actions fails, build manually. Note: binaries should be built and stripped on their native platforms for best results.

```bash
# Linux AMD64 (build on Linux)
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o aigogo-linux-amd64
strip aigogo-linux-amd64
tar -czf aigogo-linux-amd64.tar.gz aigogo-linux-amd64
sha256sum aigogo-linux-amd64.tar.gz > aigogo-linux-amd64.tar.gz.sha256

# Linux ARM64 (cross-compile on Linux)
GOOS=linux GOARCH=arm64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o aigogo-linux-arm64
tar -czf aigogo-linux-arm64.tar.gz aigogo-linux-arm64
sha256sum aigogo-linux-arm64.tar.gz > aigogo-linux-arm64.tar.gz.sha256

# macOS AMD64 (build on macOS Intel or with Rosetta)
GOOS=darwin GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o aigogo-darwin-amd64
strip aigogo-darwin-amd64
tar -czf aigogo-darwin-amd64.tar.gz aigogo-darwin-amd64
shasum -a 256 aigogo-darwin-amd64.tar.gz > aigogo-darwin-amd64.tar.gz.sha256

# macOS ARM64 (build on Apple Silicon)
GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o aigogo-darwin-arm64
strip aigogo-darwin-arm64
tar -czf aigogo-darwin-arm64.tar.gz aigogo-darwin-arm64
shasum -a 256 aigogo-darwin-arm64.tar.gz > aigogo-darwin-arm64.tar.gz.sha256

# Windows AMD64
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o aigogo-windows-amd64.exe
zip aigogo-windows-amd64.zip aigogo-windows-amd64.exe
sha256sum aigogo-windows-amd64.zip > aigogo-windows-amd64.zip.sha256

# Windows ARM64
GOOS=windows GOARCH=arm64 CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o aigogo-windows-arm64.exe
zip aigogo-windows-arm64.zip aigogo-windows-arm64.exe
sha256sum aigogo-windows-arm64.zip > aigogo-windows-arm64.zip.sha256
```

Then create release manually on GitHub and upload all archives.

## Testing Releases

```bash
# Download and test Linux binary
wget https://github.com/aupeachmo/aigogo/releases/download/v3.0.0/aigogo-linux-amd64.tar.gz
tar -xzf aigogo-linux-amd64.tar.gz
./aigogo-linux-amd64 version
./aigogo-linux-amd64 --help

# Download and test macOS Intel binary
wget https://github.com/aupeachmo/aigogo/releases/download/v3.0.0/aigogo-darwin-amd64.tar.gz
tar -xzf aigogo-darwin-amd64.tar.gz
./aigogo-darwin-amd64 version

# Download and test macOS Apple Silicon binary
wget https://github.com/aupeachmo/aigogo/releases/download/v3.0.0/aigogo-darwin-arm64.tar.gz
tar -xzf aigogo-darwin-arm64.tar.gz
./aigogo-darwin-arm64 version

# Download and test Windows binary (PowerShell)
Invoke-WebRequest -Uri "https://github.com/aupeachmo/aigogo/releases/download/v3.0.0/aigogo-windows-amd64.zip" -OutFile aigogo.zip
Expand-Archive aigogo.zip -DestinationPath .
.\aigogo-windows-amd64.exe version
```

## Version Numbering

Follow Semantic Versioning (semver):

- **MAJOR** (v2.0.0 → v3.0.0): Breaking changes
- **MINOR** (v3.0.0 → v3.1.0): New features, backward compatible
- **PATCH** (v3.0.0 → v3.0.1): Bug fixes, backward compatible

### Version Formats

| Format | Example | Use Case |
|--------|---------|----------|
| Release | `v3.0.1` | Official releases |
| Pre-release | `v3.1.0-beta.1` | Beta/RC versions |
| Development | `v3.0.0-5-g3f4a2c1` | Local dev builds |
| Dirty | `v3.0.0-5-g3f4a2c1-dirty` | Uncommitted changes |
| Fallback | `0.0.1` | No git/no tags |

## Release Signing & Verification

Releases are secured at two levels: GPG-signed git tags and GitHub build provenance attestations.

### GPG-Signed Tags

All release tags are created with `git tag -s`, producing a GPG signature. This shows a "Verified" badge on the tag in GitHub.

**Setup (one-time):**

1. Generate a GPG key (if you don't have one):
   ```bash
   gpg --full-generate-key
   ```
2. Configure git to use it:
   ```bash
   gpg --list-secret-keys --keyid-format=long
   git config --global user.signingkey <YOUR_KEY_ID>
   ```
3. Upload your public key to GitHub:
   ```bash
   gpg --armor --export <YOUR_KEY_ID>
   ```
   Paste the output at https://github.com/settings/gpg/new

**Verifying a tag:**
```bash
git verify-tag v3.0.0
```

### Build Provenance Attestations

Every release artifact (`.tar.gz` and `.zip`) is attested using `actions/attest-build-provenance` during the CI build. This creates a cryptographic record that the artifact was built by the release workflow in this repository.

**Verifying an artifact:**
```bash
# Download an artifact, then verify it
gh attestation verify aigogo-linux-amd64.tar.gz --repo aupeachmo/aigogo
```

This confirms:
- The artifact was built by a GitHub Actions workflow
- The workflow ran in the `aupeachmo/aigogo` repository
- The artifact has not been tampered with since it was built

### Future Option: Sigstore cosign

For vendor-neutral signing that doesn't depend on GitHub infrastructure, [Sigstore cosign](https://docs.sigstore.dev/) is a well-established alternative. It uses keyless signing via OIDC identity tokens from GitHub Actions.

**How it works:**
1. The CI workflow requests an OIDC token from GitHub Actions
2. `cosign sign-blob` sends the token to Fulcio (Sigstore's certificate authority), which issues a short-lived signing certificate (~20 minutes)
3. cosign signs the artifact with the certificate's ephemeral private key
4. The signature and certificate are recorded in Rekor, Sigstore's public immutable transparency log
5. The private key is discarded — it only existed in memory during signing
6. Two files are produced per artifact: `*.sig` (signature) and `*.cert` (certificate)

**How users would verify:**
```bash
cosign verify-blob aigogo-linux-amd64.tar.gz \
  --signature aigogo-linux-amd64.tar.gz.sig \
  --certificate aigogo-linux-amd64.tar.gz.cert \
  --certificate-identity-regexp "github.com/aupeachmo/aigogo" \
  --certificate-oidc-issuer "https://token.actions.githubusercontent.com"
```

**Advantages over GitHub Attestations:**
- Vendor-neutral — works without the `gh` CLI, doesn't depend on GitHub's infrastructure
- Public transparency log — anyone can audit the Rekor log
- Broad ecosystem adoption — used by Kubernetes, Homebrew, and many Go projects
- Cross-platform distribution — verification works regardless of where artifacts are hosted

**Trade-offs:**
- Users need `cosign` installed (less ubiquitous than `gh`)
- Produces additional `.sig` and `.cert` files per artifact
- Slightly more CI complexity

This could be added alongside the existing attestations if the project gains users outside the GitHub ecosystem.

## Pre-releases

For beta/RC versions:

```bash
git tag -s v3.1.0-beta.1 -m "Release v3.1.0-beta.1"
git push origin v3.1.0-beta.1
```

The workflow will automatically mark it as a pre-release.

## Troubleshooting

### Version shows "0.0.1"

**Cause**: Building without Makefile or git not available.

**Solution**:
```bash
# Use Makefile
make build

# Or set VERSION explicitly
VERSION=v3.0.0 make build
```

### Version shows "dirty"

**Cause**: Uncommitted changes in working directory.

**Solution**:
```bash
git status        # Check what's dirty
git stash         # Or commit changes
make build
```

### No git tags found

**Cause**: Repository has no tags yet.

**Solution**:
```bash
git tag -s v3.0.0 -m "Initial release"
git push origin v3.0.0
```

### Verify version was injected

```bash
# Check binary
./aigogo version

# Check binary metadata
go version -m ./aigogo | grep main.Version
```

