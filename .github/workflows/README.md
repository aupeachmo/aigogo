# Workflows Summary

| Workflow | Trigger | Purpose |
|----------|---------|---------|
| **release.yml** | Tag push (`v*.*.*`) | Build & release binaries for Linux, macOS & Windows |
| **build.yml** | Push/PR to main | Test builds on multiple platforms |
| **test.yml** | Push/PR to main | Run tests and linting |

## Quick Start

**Create a release:**
```bash
git tag -s v3.0.0 -m "Release v3.0.0"
git push origin v3.0.0
```

**Check status:**
- Visit: https://github.com/aupeachmo/aigogo/actions
- Check releases: https://github.com/aupeachmo/aigogo/releases

See [RELEASE.md](../RELEASE.md) for detailed instructions.
