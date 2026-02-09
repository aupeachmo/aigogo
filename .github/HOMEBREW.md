# Homebrew Distribution

This document describes how to distribute aigogo via Homebrew.

## Overview

There are two approaches:

| Approach | Install Command | Control | Approval Needed |
|----------|----------------|---------|-----------------|
| Personal Tap | `brew tap aupeachmo/aigogo && brew install aigogo` | Full | None |
| Homebrew Core | `brew install aigogo` | Limited | Yes |

**Recommendation**: Start with a personal tap, consider homebrew-core later once the project has traction.

## Option 1: Personal Tap (Recommended)

### Repository Structure

The tap repository (`aupeachmo/homebrew-aigogo`) contains:

```
homebrew-aigogo/
├── Formula/
│   └── aigogo.rb    # Homebrew formula (auto-updated by CI)
└── README.md        # Install instructions and setup docs
```

A template for the initial repository contents is in `.github/BREW_REPO/`.

### Setup

1. Create a new repository: `aupeachmo/homebrew-aigogo`
2. Copy the contents of `.github/BREW_REPO/` into it (Formula/ and README.md)
3. Set up the GitHub App for CI automation (see below)

### Users Install With

```bash
brew tap aupeachmo/aigogo
brew install aigogo
```

### CI Automation (GitHub App)

The release workflow (`.github/workflows/release.yml`) includes an `update-homebrew` job that automatically updates the formula on each release. It uses a GitHub App for authentication, which:

- **Never expires** (unlike Personal Access Tokens)
- **Produces signed ("Verified") commits** via the GitHub API
- **Has minimal scope** (Contents: Read and write on the tap repo only)

#### How It Works

1. The `update-homebrew` job runs after the `release` job completes
2. It downloads SHA256 checksums from the new release
3. Generates an updated `Formula/aigogo.rb` with the new version and hashes
4. Uses `actions/create-github-app-token@v1` to mint a short-lived token
5. Pushes the formula via the GitHub Contents API, producing a verified signed commit

#### One-Time Setup

1. **Create a GitHub App** at https://github.com/settings/apps/new
   - **App name**: `aigogo-homebrew-updater` (or any name)
   - **Homepage URL**: `https://github.com/aupeachmo/aigogo`
   - **Permissions** (Repository permissions only):
     - **Contents**: Read and write
   - **Where can this app be installed?**: Only on this account
   - Leave webhooks disabled (uncheck "Active" under Webhook)
   - Click **Create GitHub App**

2. **Generate a private key**
   - On the App's settings page, scroll to "Private keys"
   - Click **Generate a private key**
   - Save the downloaded `.pem` file

3. **Note the App ID**
   - On the App's settings page, find **App ID** (a number like `123456`)

4. **Install the App on the tap repository**
   - Go to https://github.com/settings/apps/aigogo-homebrew-updater/installations
   - Click **Install**
   - Select **Only select repositories** and choose `homebrew-aigogo`
   - Click **Install**

5. **Add secrets to the aigogo repository**
   - Go to https://github.com/aupeachmo/aigogo/settings/secrets/actions
   - Add `BREW_APP_ID` — the App ID from step 3
   - Add `BREW_APP_PRIVATE_KEY` — the entire contents of the `.pem` file from step 2

6. **Delete the `.pem` file** from your local machine (it's now stored securely in GitHub Secrets)

#### Verification

After the next release, check the `homebrew-aigogo` repository's commit history. The formula update commit should show a "Verified" badge and be authored by your GitHub App.

#### Troubleshooting

- **"Resource not accessible by integration"**: The App isn't installed on `homebrew-aigogo`, or `Contents: Read and write` permission is missing.
- **"Bad credentials"**: The private key or App ID is incorrect. Regenerate the private key and update the secret.
- **Formula not updated**: Check the release workflow logs in `aupeachmo/aigogo` Actions tab. The `update-homebrew` job runs after the release is created.

### Manual Formula Update

If the automation fails, you can update the formula manually:

```bash
# Get checksums from the release
VERSION=X.Y.Z
BASE_URL="https://github.com/aupeachmo/aigogo/releases/download/v${VERSION}"
curl -sL "${BASE_URL}/aigogo-darwin-amd64.tar.gz.sha256"
curl -sL "${BASE_URL}/aigogo-darwin-arm64.tar.gz.sha256"
curl -sL "${BASE_URL}/aigogo-linux-amd64.tar.gz.sha256"
curl -sL "${BASE_URL}/aigogo-linux-arm64.tar.gz.sha256"

# Edit Formula/aigogo.rb with the new version and SHA256 hashes
# Commit and push
```

## Option 2: Homebrew Core

For inclusion in the official `homebrew-core` repository.

### Requirements

The Homebrew team requires:

- **Notable project**: Demonstrated user base (GitHub stars, downloads)
- **Stable releases**: Semantic versioning, no frequent breaking changes
- **Open source license**: MIT, Apache 2.0, etc.
- **Build from source preferred**: They prefer formulas that compile from source
- **No vendored dependencies issues**: Clean dependency tree

### Process

1. Fork `Homebrew/homebrew-core`

2. Create a formula (they prefer source builds):

```ruby
class Aigogo < Formula
  desc "Make packaging and distributing your AI agents a breeze"
  homepage "https://github.com/aupeachmo/aigogo"
  url "https://github.com/aupeachmo/aigogo/archive/refs/tags/v3.0.0.tar.gz"
  sha256 "SOURCE_TARBALL_SHA256"
  license "MPL-2.0"

  depends_on "go" => :build

  def install
    system "go", "build", "-trimpath", *std_go_args(ldflags: "-s -w -X main.Version=#{version}")
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/aigogo version")
  end
end
```

3. Submit PR to `Homebrew/homebrew-core`

4. Address reviewer feedback

5. Once merged, updates require new PRs (or their bot may auto-bump)

### Ongoing Maintenance

- Version bumps can be automated by Homebrew's `brew bump-formula-pr`
- Major changes require manual PRs
- Must follow Homebrew's style guidelines

## Testing Formulas Locally

```bash
# Test a local formula file
brew install --build-from-source ./aigogo.rb

# Audit the formula
brew audit --strict aigogo

# Test the formula
brew test aigogo

# Uninstall
brew uninstall aigogo
```

## Checklist for First Release

- [ ] Create `aupeachmo/homebrew-aigogo` repository (use `.github/BREW_REPO/` as template)
- [ ] Create GitHub App (`aigogo-homebrew-updater`) with Contents: Read and write permission
- [ ] Install the App on `homebrew-aigogo` repository
- [ ] Add `BREW_APP_ID` secret to `aupeachmo/aigogo`
- [ ] Add `BREW_APP_PRIVATE_KEY` secret to `aupeachmo/aigogo`
- [ ] Tag a release and verify the `update-homebrew` job runs successfully
- [ ] Verify the tap commit shows a "Verified" badge
- [ ] Test: `brew tap aupeachmo/aigogo && brew install aigogo`
- [ ] Update main README with Homebrew install instructions
