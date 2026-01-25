# blob

[![CI](https://github.com/meigma/blob-cli/actions/workflows/ci.yml/badge.svg)](https://github.com/meigma/blob-cli/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/v/release/meigma/blob-cli)](https://github.com/meigma/blob-cli/releases)
[![Go Version](https://img.shields.io/github/go-mod/go-version/meigma/blob-cli)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT%2FApache--2.0-blue.svg)](LICENSE-MIT)

A CLI for working with blob archives in OCI registries. Push directories to registries, pull them back, or extract individual files using HTTP range requests without downloading entire archives.

![Demo](demo/demo.gif)

## Features

- **Push/Pull directories** to OCI registries as blob archives
- **Selective extraction** of files using HTTP range requests (no full download required)
- **Sigstore signing** with keyless (Fulcio) or key-based signatures
- **Policy verification** for signatures and SLSA provenance attestations
- **Interactive TUI** file browser for exploring archives
- **Alias support** for frequently used references
- **Multi-layer caching** with content deduplication, manifest caching, and per-cache control

## Installation

### Install Script (Recommended)

```bash
curl -fsSL https://blob.meigma.dev/install.sh | bash
```

### Homebrew (macOS/Linux)

```bash
brew install meigma/tap/blob
```

### Scoop (Windows)

```powershell
scoop bucket add meigma https://github.com/meigma/scoop-bucket
scoop install blob
```

### Go Install

```bash
go install github.com/meigma/blob-cli@latest
```

### Manual Download

Download pre-built binaries from the [releases page](https://github.com/meigma/blob-cli/releases).

All release artifacts include:
- SHA256 checksums signed with Sigstore
- SBOM (SPDX format) for each archive

## Quick Start

```bash
# Push a directory to a registry
blob push ghcr.io/acme/configs:v1.0.0 ./config

# Pull an archive to a local directory
blob pull ghcr.io/acme/configs:v1.0.0 ./local

# Extract a single file (uses HTTP range requests)
blob cp ghcr.io/acme/configs:v1.0.0:/nginx.conf ./nginx.conf

# View a file without downloading
blob cat ghcr.io/acme/configs:v1.0.0 config.json

# List archive contents
blob ls ghcr.io/acme/configs:v1.0.0

# Interactive file browser
blob open ghcr.io/acme/configs:v1.0.0
```

## Commands

### Archive Operations

| Command | Description |
|---------|-------------|
| `blob push <ref> <path>` | Push a directory to an OCI registry |
| `blob pull <ref> [path]` | Pull an archive to a local directory |
| `blob cp <ref>:<path>... <dest>` | Copy files from an archive (uses range requests) |
| `blob cat <ref> <file>...` | Print file contents to stdout |

### Inspection

| Command | Description |
|---------|-------------|
| `blob ls <ref> [path]` | List files and directories |
| `blob tree <ref> [path]` | Display directory structure as a tree |
| `blob inspect <ref>` | Show archive metadata, signatures, and attestations |
| `blob open <ref>` | Interactive TUI file browser |

### Security

| Command | Description |
|---------|-------------|
| `blob sign <ref>` | Sign an archive with Sigstore |
| `blob verify <ref>` | Verify signatures and attestations |

### Management

| Command | Description |
|---------|-------------|
| `blob tag <src> <dst>` | Tag a manifest with a new reference |
| `blob alias list\|set\|remove` | Manage reference aliases |
| `blob cache status\|clear\|path` | Manage local caches |
| `blob config show\|path\|edit` | View and edit configuration |

## Configuration

Configuration is stored at `~/.config/blob/config.yaml` (XDG-compliant).

```yaml
# Default output format (text or json)
output: text

# Default compression for push (none or zstd)
compression: zstd

# Cache settings
cache:
  enabled: true
  max_size: 5GB

# Aliases for frequently used references
aliases:
  configs: ghcr.io/acme/repo/configs
  app: ghcr.io/acme/repo/app:stable

# Default verification policies by image pattern
policies:
  - match: ghcr\.io/acme/.*
    policy:
      signature:
        keyless:
          issuer: https://token.actions.githubusercontent.com
          identity: https://github.com/acme/*/.github/workflows/*
```

### Environment Variables

| Variable | Description |
|----------|-------------|
| `BLOB_CONFIG` | Config file path |
| `BLOB_OUTPUT` | Default output format |
| `BLOB_CACHE_DIR` | Cache directory |
| `BLOB_USERNAME` | Registry username |
| `BLOB_PASSWORD` | Registry password |
| `NO_COLOR` | Disable colored output |

## Caching

Blob maintains several caches to improve performance and reduce bandwidth usage:

| Cache | Description |
|-------|-------------|
| `content` | File content cache (deduplicated by hash across archives) |
| `blocks` | HTTP range block cache |
| `refs` | Tag to digest mappings |
| `manifests` | OCI manifest cache |
| `indexes` | Archive index cache |

Cache location follows XDG Base Directory Specification (`~/.cache/blob` by default).

### Cache Commands

```bash
# Show cache sizes and file counts
blob cache status

# Show cache directory paths
blob cache path

# Clear all caches
blob cache clear

# Clear a specific cache type
blob cache clear indexes
```

### Cache Configuration

```yaml
# ~/.config/blob/config.yaml
cache:
  enabled: true
  dir: /custom/cache/path  # Optional: override cache location
  ref_ttl: 5m              # TTL for tag-to-digest cache (default: 5m)

  # Per-cache control (all enabled by default when cache.enabled is true)
  content:
    enabled: true
  blocks:
    enabled: true
  refs:
    enabled: true
  manifests:
    enabled: true
  indexes:
    enabled: false  # Disable specific cache types
```

## Signing and Verification

### Sign an archive

```bash
# Keyless signing with Sigstore (opens browser for OIDC)
blob sign ghcr.io/acme/configs:v1.0.0

# Sign with a private key
blob sign --key cosign.key ghcr.io/acme/configs:v1.0.0
```

### Verify signatures

```bash
# Verify with a policy file
blob verify --policy policy.yaml ghcr.io/acme/configs:v1.0.0

# Verify with OPA Rego policy
blob verify --policy-rego custom.rego ghcr.io/acme/configs:v1.0.0
```

### Policy file format

```yaml
# policy.yaml
signature:
  keyless:
    issuer: https://token.actions.githubusercontent.com
    identity: https://github.com/acme/configs/.github/workflows/*

provenance:
  slsa:
    builder: https://github.com/slsa-framework/slsa-github-generator/.github/workflows/*
    repository: acme/configs
    branch: main
```

## JSON Output

All commands support `--output json` for machine-readable output:

```bash
blob inspect --output json ghcr.io/acme/configs:v1.0.0
blob ls --output json ghcr.io/acme/configs:v1.0.0
```

## Global Flags

```
--output <format>   Output format: text, json (default: text)
--config <file>     Config file path
--verbose, -v       Increase verbosity (repeatable: -vv, -vvv)
--quiet, -q         Suppress non-error output
--no-color          Disable colored output
--plain-http        Use HTTP instead of HTTPS for registries
```

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Usage error |
| 3 | Authentication error |
| 4 | Not found |
| 5 | Verification failed |

## License

Licensed under either of [Apache License, Version 2.0](LICENSE-APACHE) or [MIT License](LICENSE-MIT) at your option.
