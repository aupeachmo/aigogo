# homebrew-aigogo

Homebrew tap for [aigogo](https://github.com/aupeach/aigogo)

## Install

```bash
brew tap aupeach/aigogo
brew install aigogo
```

## Upgrade

```bash
brew update
brew upgrade aigogo
```

## How it works

The formula in `Formula/aigogo.rb` is automatically updated by the [aigogo release workflow](https://github.com/aupeach/aigogo/actions) each time a new version is tagged. The workflow downloads the release checksums, generates an updated formula, and commits it to this repository via the GitHub API (producing verified, signed commits).

## Manual formula update

If the automation fails, you can update the formula manually:

```bash
# Get checksums from the release
VERSION=X.Y.Z
BASE_URL="https://github.com/aupeach/aigogo/releases/download/v${VERSION}"
curl -sL "${BASE_URL}/aigogo-darwin-amd64.tar.gz.sha256"
curl -sL "${BASE_URL}/aigogo-darwin-arm64.tar.gz.sha256"
curl -sL "${BASE_URL}/aigogo-linux-amd64.tar.gz.sha256"
curl -sL "${BASE_URL}/aigogo-linux-arm64.tar.gz.sha256"

# Edit Formula/aigogo.rb with the new version and SHA256 hashes
# Commit and push
```

## Automation setup (for maintainers)

The release workflow in `aupeach/aigogo` uses a GitHub App to push formula updates to this repository. This produces signed ("Verified") commits and never expires.

### One-time setup

1. **Create a GitHub App** at https://github.com/settings/apps/new
   - **App name**: `aigogo-homebrew-updater` (or any name)
   - **Homepage URL**: `https://github.com/aupeach/aigogo`
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
   - Go to https://github.com/aupeach/aigogo/settings/secrets/actions
   - Add `BREW_APP_ID` — the App ID from step 3
   - Add `BREW_APP_PRIVATE_KEY` — the entire contents of the `.pem` file from step 2

6. **Delete the `.pem` file** from your local machine (it's now stored securely in GitHub Secrets)

### Verification

After the next release, check this repository's commit history. The formula update commit should show a "Verified" badge and be authored by your GitHub App.

### Troubleshooting

- **"Resource not accessible by integration"**: The App isn't installed on `homebrew-aigogo`, or `Contents: Read and write` permission is missing.
- **"Bad credentials"**: The private key or App ID is incorrect. Regenerate the private key and update the secret.
- **Formula not updated**: Check the release workflow logs in `aupeach/aigogo` Actions tab. The `update-homebrew` job runs after the release is created.
