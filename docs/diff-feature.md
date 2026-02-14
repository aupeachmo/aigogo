# `aigg diff` Command & Push `--dry-run` Guard

## Overview

Two new capabilities for comparing package versions and avoiding redundant pushes:

- **`aigg diff`** — compare package contents between working directory, local builds, and remote registry images, with git-style unified diff output.
- **`aigg push --dry-run`** — check whether a push would upload anything new, without actually pushing.

## Usage

```bash
# Compare working directory against the latest local build (uses name:version from aigogo.json)
aigg diff

# Compare working directory against a specific local build
aigg diff utils:1.0.0

# Compare two local builds
aigg diff utils:1.0.0 utils:1.1.0

# Compare latest local build against a remote registry image
aigg diff --remote docker.io/org/utils:1.0.0

# Compare a specific local build against a remote image
aigg diff --remote utils:1.0.0 docker.io/org/utils:1.0.0

# Compact summary (M/A/D listing instead of unified diff)
aigg diff --summary utils:1.0.0

# Check if a push is needed without pushing
aigg push docker.io/org/utils:1.0.0 --from utils:1.0.0 --dry-run
```

## Output Formats

**Default (unified diff):**
```
diff a/utils.py b/utils.py
--- a/utils.py
+++ b/utils.py
@@ -1,3 +1,4 @@
 import os
+import sys
 def helper():
-    return "old"
+    return "new"

Only in a: old_module.py
Only in b: new_module.py

1 modified, 1 added, 1 removed, 1 unchanged
```

**Summary (`--summary`):**
```
M utils.py
A new_module.py
D old_module.py
1 modified, 1 added, 1 removed, 1 unchanged
```

**Identical packages:** `Packages are identical.`

## Sameness Detection (Two-Tier)

The `--dry-run` flag and `--remote` comparisons use a two-tier approach:

1. **Tier 1 (cheap):** Compare the SHA256 digest of the local `layer.tar` against the remote registry manifest's layer digest. If they match, the packages are identical — no content download needed.

2. **Tier 2 (accurate):** If digests differ (or Tier 1 can't be performed), pull the remote image, extract both to temp directories, and do a file-by-file content comparison with unified diffs.

Tier 1 can have false negatives because `builder.go` sets `.aigogo-manifest.json` ModTime to `time.Now()`, making tar bytes non-deterministic across rebuilds of the same content.

## What Gets Compared

The diff compares **source files only**. These files are excluded from comparison:
- `aigogo.json` — the package manifest (version changes with every build)
- `.aigogo-metadata.json` — local build metadata
- `.aigogo-manifest.json` — embedded tar metadata

Binary files are detected (null-byte check in first 8KB) and reported as `Binary files a/... and b/... differ` rather than showing content.

## Known Limitations

1. **Large file memory usage** — The unified diff uses an O(NM) LCS algorithm. Two files with 10,000 lines each allocate ~800MB for the DP table. Typical agent source files are well under this, but large generated files could cause issues.

2. **Permission-only changes are invisible** — The diff compares file content only (`bytes.Equal`). A file whose permissions changed (e.g. `0644` to `0755`) but whose content is identical will show as unchanged.

3. **CRLF vs LF** — Byte-level comparison means `\r\n` vs `\n` is detected as a modification. The unified diff output will show lines as changed but the `\r` characters won't be visually obvious.

4. **Symlinks are followed** — `filepath.Walk` dereferences symlinks. Broken symlinks will cause errors. Circular symlinks could potentially loop (though Go's walker has some protection).

5. **Empty directories ignored** — Only files are compared. An empty directory present on one side but absent on the other will not appear in the diff.

6. **`--remote` with unprefixed ref** — Running `aigg diff --remote utils:1.0.0` (without a registry prefix like `docker.io/org/`) will attempt to reach `docker.io/utils:1.0.0`, which is probably not what was intended. No validation warns about this.

7. **No-newline-at-EOF** — If one file ends with `\n` and the other does not, `bytes.Equal` correctly flags the file as modified, but the unified diff may not clearly indicate that the only difference is a trailing newline.

8. **Tier 1 false negatives by design** — `builder.go` sets `.aigogo-manifest.json` ModTime to `time.Now()`, so rebuilding the same content produces different tar bytes. Tier 1 will report "different" and fall through to Tier 2. Tier 1 only matches when pushing the exact same build twice without rebuilding in between.
