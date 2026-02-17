# Security Scanners and aigogo Registry Artifacts

## The Problem

aigogo uses Docker/OCI registries as a transport mechanism for distributing source code packages. The artifacts it pushes are **not runnable container images** — they contain only source files packed in a minimal Docker v2 manifest structure.

However, security scanners (Trivy, Snyk, Grype, Docker Scout, AWS ECR scanning, etc.) that monitor your registry will see these artifacts and attempt to scan them as if they were container images. This can cause:

- **False positives**: Scanners may flag the empty config or missing OS layer as anomalous.
- **Scan failures**: Some scanners expect a valid OS filesystem and fail on source-only tarballs.
- **Noise in dashboards**: aigogo artifacts mixed in with real container images clutter vulnerability reports.

## What aigogo Pushes

An aigogo artifact pushed to a registry consists of:

| Component | Value | Notes |
|-----------|-------|-------|
| **Manifest** | Docker v2 (`application/vnd.docker.distribution.manifest.v2+json`) | Standard Docker manifest |
| **Config blob** | `{}` (2 bytes) | Empty JSON — no OS, no entrypoint, no env vars |
| **Layer** | Single plain tar (`application/vnd.docker.image.rootfs.diff.tar`) | Contains source files + `.aigogo-manifest.json` |

This is a valid Docker image by spec, but it has no operating system, no packages, no binaries — just source code files. There is nothing to scan for CVEs.

## Recommended Mitigations

### 1. Use a Dedicated Repository or Namespace

Keep aigogo artifacts in a clearly named repository path so scanners can be configured to skip them:

```bash
# Good: clear namespace separation
aigg push ghcr.io/myorg/aigogo/my-agent:1.0.0
aigg push ghcr.io/myorg/aigogo-agents/my-agent:1.0.0

# Avoid: mixed with real container images
aigg push ghcr.io/myorg/my-agent:1.0.0
```

Using a prefix like `aigogo/` or `aigogo-agents/` in the repository path makes it straightforward to write exclusion rules.

### 2. Use Tag Conventions

Add a consistent tag prefix or suffix to identify aigogo artifacts:

```bash
# Tag conventions that signal "not a container"
aigg push ghcr.io/myorg/agents/my-agent:aigogo-1.0.0
aigg push ghcr.io/myorg/agents/my-agent:src-1.0.0
```

### 3. Configure Scanner Exclusions

Most scanners support repository exclusion patterns.

**Trivy:**

```bash
# Skip aigogo repositories
trivy image --skip-dirs "aigogo" ghcr.io/myorg/aigogo/my-agent:1.0.0

# Or exclude via .trivyignore in CI
echo "ghcr.io/myorg/aigogo/*" >> .trivyignore
```

In CI, filter which repositories Trivy scans:

```yaml
- name: Scan container images
  run: |
    for image in $(list-production-images); do
      # Skip aigogo source packages
      if echo "$image" | grep -q "/aigogo/"; then
        echo "Skipping aigogo artifact: $image"
        continue
      fi
      trivy image "$image"
    done
```

**Snyk:**

```bash
# Exclude aigogo repos from Snyk Container monitoring
snyk monitor --docker --exclude="aigogo/*"
```

Or in Snyk's web UI, configure repository filtering to exclude the `aigogo/` namespace.

**Grype:**

```bash
# Grype will likely report 0 vulnerabilities (no OS packages)
# but you can skip entirely:
grype ghcr.io/myorg/aigogo/my-agent:1.0.0 || true
```

**AWS ECR Scanning:**

ECR's built-in scanning triggers on push. To avoid scanning aigogo artifacts:
- Push aigogo packages to a separate ECR repository with scanning disabled
- Or use ECR scan filters to exclude repositories matching `aigogo*`

**Docker Scout:**

```bash
# Exclude repositories from Docker Scout analysis
docker scout config exclude add "myorg/aigogo/*"
```

### 4. Use a Separate Registry

For maximum separation, push aigogo artifacts to a dedicated registry instance that has no security scanning enabled:

```bash
# Dedicated aigogo registry (no scanner attached)
aigg push aigogo-registry.internal:5000/my-agent:1.0.0

# Production container registry (scanner attached)
docker push prod-registry.internal:5000/my-service:latest
```

This is the cleanest approach for organizations with strict scanning policies.

## Why Not OCI Artifacts?

You might wonder why aigogo doesn't use [OCI Artifacts](https://github.com/opencontainers/artifacts) with a custom media type, which would cleanly distinguish aigogo packages from container images. The reason is **registry compatibility**: not all registries support custom artifact types. Docker Hub, in particular, requires standard Docker v2 manifests. By using the standard Docker image format, aigogo works with every Docker V2-compatible registry out of the box.

This trade-off (universal compatibility vs. clean type distinction) is intentional. The mitigations above — namespace separation, tag conventions, and scanner exclusions — are practical solutions that work today across all registries.

## Summary

| Strategy | Effort | Effectiveness |
|----------|--------|---------------|
| Dedicated namespace (`aigogo/`) | Low | High — easy scanner exclusion rules |
| Tag conventions | Low | Medium — helps humans, some scanners |
| Scanner exclusion config | Medium | High — eliminates false positives |
| Separate registry | High | Complete — no scanner sees aigogo artifacts |

**Recommendation:** Use a dedicated repository namespace (e.g., `ghcr.io/myorg/aigogo/`) and configure your scanner to exclude that path. This is the lowest-effort approach that solves the problem cleanly.
