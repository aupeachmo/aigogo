# Delete All Tags Feature

## Overview

The `--all` flag for the `delete` command enables deletion of all tags in a repository in a single operation.

## Usage

```bash
# Delete a specific tag (existing behavior)
aigogo delete docker.io/myorg/utils:1.0.0

# Delete all tags in a repository (new)
aigogo delete --all docker.io/myorg/utils
```

**Important**: Flags must come **before** the image reference.

## Implementation Details

### Architecture

**Files Modified:**
- `pkg/docker/deleter.go` - Added `listTags()` and `DeleteAll()` methods
- `cmd/delete.go` - Added flag parsing and routing logic
- `cmd/root.go` - Already had flag support, no changes needed

### Flow

1. **User runs**: `aigogo delete --all docker.io/myorg/utils`
2. **Flag parsing**: Root command parses `--all` flag
3. **Confirmation**: Prompts for `DELETE ALL` confirmation (using `bufio.Reader`)
4. **List tags**: Calls Docker Registry API `/v2/<name>/tags/list`
5. **Delete each**: Loops through tags, calling existing `Delete()` for each
6. **Progress**: Shows real-time progress (✓ or ✗) for each tag
7. **Summary**: Reports successes/failures

### API Calls

```
GET /v2/<repository>/tags/list
→ Returns: {"name": "...", "tags": ["1.0.0", "2.0.0", "latest"]}

For each tag:
  HEAD /v2/<repository>/manifests/<tag>
  → Returns: Docker-Content-Digest header
  
  DELETE /v2/<repository>/manifests/<digest>
  → Returns: 202 Accepted
```

### Error Handling

- **Partial failures**: Continues with remaining tags if some fail
- **Authentication errors**: Clear message with login instructions
- **Registry doesn't support deletion**: Informative error message
- **Repository not found**: Clear 404 error

### Safety Features

1. **Stronger confirmation**: Requires typing `DELETE ALL` (not just `yes`)
2. **Preview**: Shows all tags before deletion
3. **Progress reporting**: See which tags succeed/fail in real-time
4. **Failure summary**: Lists all failed tags at the end

## Examples

### Success Case

```bash
$ aigogo delete --all docker.io/myorg/utils

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

### Partial Failure

```bash
$ aigogo delete --all docker.io/myorg/utils

Found 3 tag(s):
  - 1.0.0
  - 2.0.0
  - protected

Deleting 1.0.0... ✓ Deleted
Deleting 2.0.0... ✓ Deleted
Deleting protected... ✗ Failed: authentication failed (insufficient permissions)

⚠️  Warning: 1 tag(s) failed to delete: [protected]
Successfully deleted 2 out of 3 tags
```

### Cancellation

```bash
$ aigogo delete --all docker.io/myorg/utils

⚠️  WARNING: This will permanently delete ALL tags in docker.io/myorg/utils

Type 'DELETE ALL' to confirm: no
Delete cancelled
```

## Testing

```bash
# Build
go build -o aigogo .

# Test single delete (existing)
echo "yes" | ./aigogo delete test.example.com/fake:1.0.0

# Test delete all with confirmation
echo "DELETE ALL" | ./aigogo delete --all test.example.com/fake

# Test delete all with cancellation
echo "no" | ./aigogo delete --all test.example.com/fake

# Check help
./aigogo delete --help
```

## Documentation

Updated files:
- `docs/DELETE_COMMAND.md` - Added --all examples and usage
- `docs/COMMANDS_SUMMARY.md` - Added --all to command overview
- `README.md` - Added delete commands to table

## Future Enhancements

Possible improvements:
- `--dry-run` flag to preview without deleting
- `--filter` to delete tags matching a pattern (e.g., `--filter "v0.*"`)
- `--older-than` to delete tags older than a date
- Parallel deletion for faster bulk operations
- Progress bar for large numbers of tags

## Notes

### Go Flag Package Behavior

Go's standard `flag` package stops parsing when it encounters a non-flag argument, which is why flags must come before positional arguments:

```bash
# Correct
aigogo delete --all docker.io/myorg/utils

# Incorrect (--all is treated as a positional argument)
aigogo delete docker.io/myorg/utils --all
```

This is standard Go behavior and matches other Go CLI tools.

### Registry Limitations

Not all registries support deletion:
- Docker Hub: Limited support, may require web UI
- GitHub Container Registry (ghcr.io): Full support ✅
- Self-hosted: Requires `REGISTRY_STORAGE_DELETE_ENABLED=true`

### Confirmation Reading

The confirmation prompt uses `bufio.Reader.ReadString('\n')` instead of `fmt.Scanln()` to properly read multi-word confirmations like "DELETE ALL". This is important because `fmt.Scanln()` stops at the first space character.

