# aigogo delete Command

## Overview

The `delete` command removes a snippet package from a remote Docker registry permanently.

**⚠️ WARNING**: This is a destructive operation that cannot be undone!

## Usage

```bash
# Delete a specific tag
aigogo delete <registry>/<name>:<tag>

# Delete all tags in a repository
aigogo delete <registry>/<name> --all
```

## Options

- `--all` - Delete all tags in the repository (requires stronger confirmation: 'DELETE ALL')

**Note**: Flags must come before the image reference:
```bash
# Correct
aigogo delete --all docker.io/myorg/utils

# Incorrect
aigogo delete docker.io/myorg/utils --all
```

### With Confirmation

The command requires explicit confirmation:

```bash
$ aigogo delete docker.io/myorg/utils:1.0.0
⚠️  WARNING: This will permanently delete docker.io/myorg/utils:1.0.0 from the registry
Are you sure? Type 'yes' to confirm: yes
Deleting docker.io/myorg/utils:1.0.0 from registry...
✓ Successfully deleted docker.io/myorg/utils:1.0.0 from registry

Note: The local cache is not affected. To remove from cache, run:
  aigogo remove docker.io/myorg/utils:1.0.0
```

### Cancellation

Type anything other than "yes" to cancel:

```bash
$ aigogo delete docker.io/myorg/utils:1.0.0
⚠️  WARNING: This will permanently delete docker.io/myorg/utils:1.0.0 from the registry
Are you sure? Type 'yes' to confirm: no
Delete cancelled
```

## Command Comparison

| Command | Scope | Reversible | Affects |
|---------|-------|------------|---------|
| `aigogo rm` | Local manifest | ✅ Yes | `aigogo.json` file |
| `aigogo remove` | Local cache | ✅ Yes | `~/.aigogo/cache/` |
| `aigogo delete` | ⚠️ Remote registry | ❌ **NO** | Docker registry |

**Key Points**:
- `rm` - edits your local manifest
- `remove` - deletes from local cache (can re-download)
- `delete` - **permanently removes from registry** (cannot undo!)

## How It Works

### Docker Registry API V2

The delete command uses the Docker Registry HTTP API v2:

1. **Get Manifest Digest**:
   ```
   HEAD /v2/<name>/manifests/<tag>
   → Returns: Docker-Content-Digest: sha256:abc123...
   ```

2. **Delete by Digest**:
   ```
   DELETE /v2/<name>/manifests/sha256:abc123...
   → Returns: 202 Accepted
   ```

**Note**: You cannot delete by tag directly, must use digest.

### Authentication Required

You must be logged in with sufficient permissions:

```bash
aigogo login docker.io
aigogo delete docker.io/myorg/utils:1.0.0
```

### Registry Support

Not all registries support deletion:

| Registry | Delete Support | Notes |
|----------|----------------|-------|
| **Docker Hub** | ⚠️ Limited | Only repo owners, may not work for free accounts |
| **GitHub Container Registry** | ✅ Yes | Full support with proper permissions |
| **GitLab Container Registry** | ✅ Yes | Full support |
| **AWS ECR** | ✅ Yes | Via image lifecycle policies |
| **Google GCR** | ✅ Yes | Full support |
| **Azure ACR** | ✅ Yes | Full support |
| **Self-hosted** | ⚠️ Depends | Must enable with `REGISTRY_STORAGE_DELETE_ENABLED=true` |

## Examples

### Delete Specific Tag

```bash
$ aigogo delete docker.io/myorg/utils:1.0.0
⚠️  WARNING: This will permanently delete docker.io/myorg/utils:1.0.0 from the registry
Are you sure? Type 'yes' to confirm: yes
✓ Successfully deleted docker.io/myorg/utils:1.0.0 from registry
```

### Delete All Tags (Entire Repository)

```bash
$ aigogo delete docker.io/myorg/utils --all

⚠️  WARNING: This will permanently delete ALL tags in docker.io/myorg/utils

Type 'DELETE ALL' to confirm: DELETE ALL

Found 4 tag(s):
  - 0.9.0
  - 1.0.0
  - 2.0.0
  - latest

Deleting 0.9.0... ✓ Deleted
Deleting 1.0.0... ✓ Deleted
Deleting 2.0.0... ✓ Deleted
Deleting latest... ✓ Deleted

✓ Successfully deleted all 4 tag(s) from docker.io/myorg/utils
```

### Delete All Tags - Partial Failure

If some tags fail to delete (e.g., due to permissions), the command continues and reports failures:

```bash
$ aigogo delete docker.io/myorg/utils --all

Found 5 tag(s):
  - 0.9.0
  - 1.0.0
  - 1.1.0
  - 2.0.0
  - latest

Deleting 0.9.0... ✓ Deleted
Deleting 1.0.0... ✓ Deleted
Deleting 1.1.0... ✗ Failed: authentication failed (insufficient permissions)
Deleting 2.0.0... ✓ Deleted
Deleting latest... ✓ Deleted

⚠️  Warning: 1 tag(s) failed to delete: [1.1.0]
Successfully deleted 4 out of 5 tags
```

### Delete from GitHub Container Registry

```bash
$ aigogo delete ghcr.io/myorg/utils:1.0.0
⚠️  WARNING: This will permanently delete ghcr.io/myorg/utils:1.0.0 from the registry
Are you sure? Type 'yes' to confirm: yes
✓ Successfully deleted ghcr.io/myorg/utils:1.0.0 from registry
```

### Delete and Remove Locally

```bash
# Delete from registry
aigogo delete docker.io/myorg/utils:1.0.0

# Remove from local cache
aigogo remove docker.io/myorg/utils:1.0.0
```

## Error Cases

### Not Logged In

```bash
$ aigogo delete docker.io/myorg/utils:1.0.0
Error: not logged in to docker.io
Run 'aigogo login docker.io' first
```

**Solution**: Login first:
```bash
aigogo login docker.io
```

### Insufficient Permissions

```bash
$ aigogo delete docker.io/myorg/utils:1.0.0
Error: authentication failed (insufficient permissions)
```

**Causes**:
- Not the repository owner
- Read-only credentials
- Organization permissions not granted

### Registry Doesn't Support Deletion

```bash
$ aigogo delete my-registry.example.com/utils:1.0.0
Error: registry does not support deletion (check registry configuration)
```

**Solution**: For self-hosted registries, enable deletion:
```yaml
# config.yml
storage:
  delete:
    enabled: true
```

### Image Not Found

```bash
$ aigogo delete docker.io/myorg/utils:1.0.0
Error: image not found: docker.io/myorg/utils:1.0.0
```

**Causes**:
- Wrong image name
- Tag already deleted
- Image never existed

### Already Deleted

```bash
$ aigogo delete docker.io/myorg/utils:1.0.0
Error: manifest not found (may have been already deleted)
```

## Use Cases

### 1. Remove Old/Deprecated Versions

```bash
# Delete old versions individually
aigogo delete docker.io/myorg/utils:0.9.0
aigogo delete docker.io/myorg/utils:0.9.1
aigogo delete docker.io/myorg/utils:0.9.2

# Or delete entire repository at once
aigogo delete docker.io/myorg/old-utils --all
```

### 2. Remove Test/Development Tags

```bash
# Clean up test releases
aigogo delete docker.io/myorg/utils:test
aigogo delete docker.io/myorg/utils:dev
aigogo delete docker.io/myorg/utils:experiment
```

### 3. Fix Mistaken Push

```bash
# Oops, pushed wrong content
aigogo delete docker.io/myorg/utils:1.0.0

# Push correct version
aigogo push docker.io/myorg/utils:1.0.0
```

### 4. Cleanup After Organization Change

```bash
# Moving from old org to new org
# Delete from old location
aigogo delete docker.io/old-org/utils:1.0.0

# Push to new location
aigogo push docker.io/new-org/utils:1.0.0
```

## Important Considerations

### 1. Cannot Undo

Once deleted, the image is gone forever:
- Cannot recover the deleted manifest
- Cannot restore without re-pushing
- Other users cannot pull it

### 2. Doesn't Delete Layers

Deleting a manifest doesn't immediately delete the layers:
- Layers may be shared by other images
- Registry garbage collection needed
- Disk space freed only after GC

### 3. Local Cache Unaffected

Deleting from registry doesn't touch local cache:
```bash
# After delete, local cache still has it
aigogo list
# docker.io/myorg/utils:1.0.0 (cached)

# Remove locally too
aigogo remove docker.io/myorg/utils:1.0.0
```

### 4. Tag vs Digest

- Deleting by tag removes that tag reference
- Other tags pointing to same digest are not affected
- To fully remove, delete all tags pointing to same digest

Example:
```bash
# If both tags point to same content:
# utils:1.0.0 → sha256:abc123
# utils:latest → sha256:abc123

# Deleting one tag doesn't affect the other
aigogo delete docker.io/myorg/utils:1.0.0
# utils:latest still exists and points to sha256:abc123
```

## Registry-Specific Notes

### Docker Hub

- Only repository owners can delete
- Free accounts may have limitations
- May require web UI for some operations
- Deletion may not free space immediately

### GitHub Container Registry (ghcr.io)

- Full delete support
- Requires `write:packages` permission
- Can use personal access token (PAT)
- Integrates with GitHub Actions

```bash
# Login with PAT
echo $GITHUB_TOKEN | aigogo login ghcr.io -u username --password-stdin

# Delete
aigogo delete ghcr.io/myorg/utils:1.0.0
```

### Self-Hosted Registry

Enable deletion in config:

```yaml
# config.yml
version: 0.1
storage:
  delete:
    enabled: true
  filesystem:
    rootdirectory: /var/lib/registry
```

Restart registry:
```bash
docker restart registry
```

Run garbage collection:
```bash
docker exec registry bin/registry garbage-collect /etc/docker/registry/config.yml
```

## Best Practices

### 1. Use with Caution

Only delete when absolutely necessary:
- ✅ Removing test/dev tags
- ✅ Cleaning up very old versions
- ✅ Fixing critical mistakes
- ❌ Don't delete actively used versions
- ❌ Don't delete latest stable releases

### 2. Communicate Before Deleting

If used by a team:
```bash
# Notify team
# "Deleting utils:0.9.x versions tomorrow"

# Then delete
aigogo delete docker.io/myorg/utils:0.9.0
```

### 3. Keep Latest Stable

```bash
# OK to delete
aigogo delete docker.io/myorg/utils:1.0.0-beta
aigogo delete docker.io/myorg/utils:1.0.0-rc1

# Keep these
# utils:1.0.0 (latest stable)
# utils:2.0.0 (current)
```

### 4. Document Deletions

Keep a log:
```bash
# deletion-log.txt
2024-01-07: Deleted utils:0.9.x (replaced by 1.0.0)
2024-01-15: Deleted utils:test (experimental tag)
```

### 5. Test First

Use a test repository:
```bash
# Test deletion flow
aigogo push docker.io/myorg/test-delete:1.0.0
aigogo delete docker.io/myorg/test-delete:1.0.0
# Verify it worked
```

## Automation

### Script to Delete Old Versions

```bash
#!/bin/bash
# delete-old-versions.sh

REPO="docker.io/myorg/utils"
VERSIONS_TO_DELETE=("0.9.0" "0.9.1" "0.9.2")

for ver in "${VERSIONS_TO_DELETE[@]}"; do
    echo "Deleting $REPO:$ver"
    echo "yes" | aigogo delete "$REPO:$ver"
done
```

**Note**: Auto-confirmation with `echo "yes" |` bypasses safety prompt - use carefully!

### Script to Delete Entire Repository

```bash
#!/bin/bash
# delete-entire-repo.sh

REPO="docker.io/myorg/deprecated-utils"

echo "This will delete ALL tags in $REPO"
echo "DELETE ALL" | aigogo delete "$REPO" --all
```

**⚠️ DANGER**: This deletes ALL tags! Use with extreme caution!

### CI/CD Cleanup

```yaml
# .github/workflows/cleanup.yml
name: Cleanup Old Versions

on:
  schedule:
    - cron: '0 0 * * 0'  # Weekly

jobs:
  cleanup:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Login
        run: aigogo login ghcr.io -u ${{ github.actor }} -p ${{ secrets.GITHUB_TOKEN }}
      
      - name: Delete old dev tags
        run: |
          for tag in dev test experiment; do
            echo "yes" | aigogo delete ghcr.io/${{ github.repository }}:$tag || true
          done
```

## Troubleshooting

### Problem: "Registry does not support deletion"

**Solution**: Enable deletion on registry:
```yaml
storage:
  delete:
    enabled: true
```

### Problem: "Insufficient permissions"

**Solution**: Use credentials with write/delete permissions:
```bash
# Use admin token
aigogo login docker.io -u admin -p $ADMIN_TOKEN
```

### Problem: "Manifest not found"

**Possible causes**:
1. Already deleted
2. Wrong tag name
3. Wrong registry

**Solution**: Verify image exists first:
```bash
aigogo pull docker.io/myorg/utils:1.0.0
# If this works, image exists
```

## Summary

**Commands**: 
- `aigogo delete <registry>/<name>:<tag>` - Delete specific tag
- `aigogo delete <registry>/<name> --all` - Delete all tags

**Purpose**: Permanently delete from remote registry

**Features**:
- ✅ Confirmation prompt (safety - 'yes' for single, 'DELETE ALL' for --all)
- ✅ Works with any Docker Registry API v2 compliant registry
- ✅ Proper authentication
- ✅ Error handling for common issues
- ✅ Partial failure handling (continues with remaining tags)
- ✅ Progress reporting for bulk deletion

**Important**:
- ⚠️ **Cannot undo**
- ⚠️ Requires proper permissions
- ⚠️ Not all registries support it
- ⚠️ Local cache unaffected

**Use cases**:
- Clean up old versions
- Remove test/dev tags
- Fix mistaken pushes
- Manage repository lifecycle

**Remember**: With great power comes great responsibility - use delete carefully! 🚨

