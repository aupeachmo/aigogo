# Binary Size Optimization

## Stripping Strategy

aigogo binaries are optimized for size using multiple techniques:

### 1. Go Build Flags

```bash
go build -ldflags="-s -w" -o aigogo
```

- **`-s`**: Omit the symbol table and debug information
- **`-w`**: Omit the DWARF symbol table
- **Result**: ~30% size reduction

### 2. Additional Stripping (Unix platforms)

For Linux and macOS binaries, we run `strip` after building:

```bash
strip aigogo-linux-amd64
strip aigogo-darwin-amd64
strip aigogo-darwin-arm64
```

- **Result**: Additional 5-10% size reduction
- **Note**: macOS binaries must be stripped on macOS (native runners)

### 3. Windows Binaries

Windows binaries rely solely on `-ldflags="-s -w"` for size optimization:
- No standard `strip` command available on Windows runners
- The ldflags approach is sufficient and widely used for Go binaries
- Installing MinGW strip adds complexity with minimal benefit (~1-2% reduction)

## Size Comparison

Typical binary sizes (v3.0.0):

| Build Type | Linux AMD64 | Linux ARM64 | macOS AMD64 | macOS ARM64 | Windows AMD64 | Windows ARM64 |
|------------|-------------|-------------|-------------|-------------|---------------|---------------|
| Unstripped | ~12.5 MB | ~12.0 MB | ~12.0 MB | ~11.8 MB | ~12.5 MB | ~11.8 MB |
| With -s -w | ~9.3 MB | ~8.9 MB | ~8.9 MB | ~8.7 MB | ~9.3 MB | ~8.7 MB |
| With strip | ~8.9 MB | ~8.9 MB* | ~8.5 MB | ~8.3 MB | ~9.3 MB* | ~8.7 MB* |

*Linux ARM64 and Windows use `-s -w` only (no post-build strip on cross-compiled/Windows targets)

## Testing Stripped Binaries

Verify stripped binaries work correctly:

```bash
# Check if binary is stripped
file aigogo-linux-amd64
# Should show: "stripped"

# Verify it still works
./aigogo-linux-amd64 version
./aigogo-linux-amd64 init
```

## Further Optimization

For even smaller binaries, consider:

1. **UPX Compression** (optional, may trigger antivirus):
   ```bash
   upx --best --lzma aigogo-linux-amd64
   ```
   - Can reduce size by 50-70%
   - Increases startup time slightly
   - May be flagged by antivirus software

2. **Build with Go 1.22+ dead code elimination**:
   Already enabled by default in Go 1.22+

3. **Trim unused dependencies**:
   ```bash
   go mod tidy
   ```

## Current GitHub Actions Workflow

Our release workflow:
1. ✅ Builds for each platform (Linux ARM64 is cross-compiled)
2. ✅ Builds with `-ldflags="-s -w"` (all platforms)
3. ✅ Strips Unix binaries (Linux and macOS) with `strip` command
4. ✅ Windows relies on ldflags only (no post-build strip)
5. ✅ Creates compressed tarballs/zips (additional ~60% size reduction)

Final distribution sizes:
- `aigogo-linux-amd64.tar.gz`: ~3.2 MB
- `aigogo-linux-arm64.tar.gz`: ~3.1 MB
- `aigogo-darwin-amd64.tar.gz`: ~3.0 MB
- `aigogo-darwin-arm64.tar.gz`: ~2.9 MB
- `aigogo-windows-amd64.zip`: ~3.2 MB
- `aigogo-windows-arm64.zip`: ~3.1 MB
