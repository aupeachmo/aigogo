#!/bin/bash
set -euo pipefail

# Get the latest version tag
latest=$(git describe --tags --abbrev=0 2>/dev/null || echo "")

if [ -n "$latest" ]; then
    # Strip v prefix, split into parts, increment patch
    version="${latest#v}"
    major=$(echo "$version" | cut -d. -f1)
    minor=$(echo "$version" | cut -d. -f2)
    patch=$(echo "$version" | cut -d. -f3)
    suggested="v${major}.${minor}.$((patch + 1))"
    echo "Current release: $latest"
    read -rp "New release tag [$suggested]: " tag
    tag="${tag:-$suggested}"
else
    echo "No existing tags found."
    read -rp "New release tag: " tag
    [ -z "$tag" ] && echo "No tag provided." && exit 1
fi

# Validate format
if ! echo "$tag" | grep -qE '^v[0-9]+\.[0-9]+\.[0-9]+$'; then
    echo "Tag must match vMAJOR.MINOR.PATCH (e.g. v3.1.0)"
    exit 1
fi

# Confirm
echo ""
echo "This will:"
echo "  1. Create GPG-signed tag: $tag"
echo "  2. Push tag to origin"
echo "  3. GitHub Actions will build, attest, and create the release"
echo ""
read -rp "Proceed? [y/N] " confirm
[[ "$confirm" =~ ^[Yy]$ ]] || { echo "Aborted."; exit 0; }

git tag -s "$tag" -m "Release $tag"
git push origin "$tag"

echo ""
echo "Tag $tag pushed. Release workflow started."
echo "Watch progress: https://github.com/$(git remote get-url origin | sed 's|.*github.com[:/]\(.*\)\.git|\1|')/actions"
