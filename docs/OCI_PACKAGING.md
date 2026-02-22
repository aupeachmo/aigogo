# OCI Images Are Just Tarballs

## How Container Images Actually Work

An OCI/Docker image is not magic. Strip away the tooling and what you have is:

1. **A manifest** — a small JSON file listing the pieces
2. **A config blob** — JSON metadata (entrypoint, env vars, etc.)
3. **One or more layers** — each layer is a plain tar archive

That's it. A "container image" in a registry is just these blobs stored behind an HTTP API. The registry doesn't run anything — it's a file server with content-addressable storage.

The **Docker Registry V2 API** is a straightforward REST protocol:

- `PUT /v2/<repo>/blobs/uploads/` — upload a blob (any bytes)
- `PUT /v2/<repo>/manifests/<tag>` — publish a manifest pointing to those blobs
- `GET /v2/<repo>/manifests/<tag>` — fetch a manifest
- `GET /v2/<repo>/blobs/<digest>` — download a blob

Nothing in this protocol requires the blobs to contain an operating system, binaries, or anything Docker-specific. The registry stores bytes and returns bytes.

## Why This Works for Package Distribution

aigogo exploits this by creating the **minimum viable OCI image**:

| Component | What aigogo puts there |
|-----------|----------------------|
| Config blob | `{}` (literally two bytes — empty JSON) |
| Layer | A single tar containing your source files |
| Manifest | Standard Docker v2 JSON linking the above |

The config is empty because there's nothing to configure — no entrypoint, no environment, no OS. The layer is an uncompressed tar of your project files with their original directory structure and permissions preserved.

This is a valid image by spec. Registries accept it. Clients can pull it. But it's not runnable as a container — and it doesn't need to be. We're using the registry as a **distribution network**, not a runtime platform.

No Docker daemon is involved at any point. aigogo creates tar archives with Go's standard `archive/tar` package and talks to registries over plain HTTP.

## Using Docker Registries as Package Distribution

### The Workflow

```
Author → build → push to registry → consumers pull → install locally
```

**Build** creates a local cache of your package files:
```bash
aigg build    # copies files to ~/.aigogo/cache/<name>_<version>/
```

**Push** wraps those files in a minimal OCI image and uploads:
```bash
aigg push ghcr.io/org/my-agent:1.0.0 --from my-agent:1.0.0
```

This creates a tar of your files, uploads it as a blob, uploads the empty `{}` config as another blob, then publishes a manifest tying them together.

**Pull** downloads the tar and extracts it:
```bash
aigg add ghcr.io/org/my-agent:1.0.0
```

This fetches the manifest, downloads the layer blob, extracts the tar, and stores files in a local content-addressable store (`~/.aigogo/store/sha256/`).

### Why Registries Instead of a Custom Server

Docker registries are:

- **Free** — Docker Hub, GitHub Container Registry, and others offer free tiers
- **Universal** — every cloud provider has one, every CI system supports them
- **Authenticated** — built-in access control, token-based auth, org-level permissions
- **Content-addressable** — every blob is referenced by its SHA256 digest
- **Immutable** — once pushed, a digest always returns the same bytes
- **Battle-tested** — the infrastructure that serves millions of container images daily

Building a custom package server would mean reimplementing all of this. Using registries means getting it for free.

### What Gets Distributed

A pushed package contains:

- Your source files (Python, JavaScript, etc.) in their original structure
- An `aigogo.json` manifest describing the package
- An `.aigogo-manifest.json` metadata file (stripped on extraction)

No compiled artifacts, no OS layers, no Docker-specific files. Pull it and you get back exactly the source files you pushed.

### Security Scanners

Registry security scanners may flag these artifacts since they lack an OS layer — see [SECURITY_SCANNERS.md](SECURITY_SCANNERS.md) for how to handle this.
