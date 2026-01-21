# blob-cli Design Document

> **Status:** Draft
> **Purpose:** Temporary design document to flesh out CLI design before implementation

---

## Principles

1. **User-friendly + automation-friendly** — Clean, minimal commands with friendly coloring/messaging that falls back to plain text in non-TTY environments
2. **Structured output** — All commands support `--output json` for machine consumption
3. **Standard configuration** — Cobra/Viper pattern for flags, environment variables, and config files
4. **XDG compliance** — All paths follow XDG Base Directory Specification

---

## Commands

### Core Commands

| Command | Description |
|---------|-------------|
| `blob push <ref> <path>` | Push a directory to an OCI registry as a blob archive |
| `blob pull <ref> [path]` | Pull an entire archive to a local directory |
| `blob cp <ref>:<path> <dest>` | Copy specific file(s) or directories from an archive to local filesystem |
| `blob cat <ref> <file>...` | Print file contents to stdout (for viewing/piping) |
| `blob ls <ref> [path]` | List files/directories in an archive |

### Inspection & Metadata

| Command | Description |
|---------|-------------|
| `blob inspect <ref>` | Show archive metadata (file count, size, signatures, attestations) |
| `blob tree <ref> [path]` | Display directory structure as a tree |

### Security & Provenance

| Command | Description |
|---------|-------------|
| `blob sign <ref>` | Sign an archive with Sigstore |
| `blob verify <ref>` | Verify signatures and attestations against policies |

### Registry Management

| Command | Description |
|---------|-------------|
| `blob tag <src-ref> <dst-ref>` | Tag an existing manifest with a new reference |

### Cache Management

| Command | Description |
|---------|-------------|
| `blob cache status` | Show cache sizes (individual and total) |
| `blob cache clear [type]` | Clear caches (all or specific type) |
| `blob cache path` | Show cache directory paths |

### Aliases

| Command | Description |
|---------|-------------|
| `blob alias list` | List all configured aliases |
| `blob alias set <name> <ref>` | Add or update an alias |
| `blob alias remove <name>` | Remove an alias |

### Configuration

| Command | Description |
|---------|-------------|
| `blob config show` | Display current configuration |
| `blob config path` | Show configuration file path |
| `blob config edit` | Open configuration in $EDITOR |

---

## Command Details

### `blob push`

```
blob push <ref> <path>

Push a directory to an OCI registry as a blob archive.

Arguments:
  <ref>     Target reference (e.g., ghcr.io/org/repo:tag)
  <path>    Source directory to archive

Flags:
  -c, --compression <type>    Compression type: none, zstd (default: zstd)
      --skip-compressed       Skip compressing already-compressed files (default: true)
      --sign                  Sign the archive after pushing
      --annotation <k=v>      Add annotation to manifest (repeatable)

Examples:
  blob push ghcr.io/acme/configs:v1.0.0 ./config
  blob push --sign ghcr.io/acme/configs:latest ./config
```

### `blob pull`

```
blob pull <ref> [path]

Pull an archive from an OCI registry to a local directory.

Arguments:
  <ref>     Source reference, alias, or alias:tag
  [path]    Destination directory (default: current directory)

Flags:
      --policy <file>         Policy file for verification (can be repeated)
      --policy-rego <file>    OPA Rego policy file
      --policy-bundle <file>  OPA bundle for policy evaluation
      --no-default-policy     Skip policies from config file

Examples:
  blob pull ghcr.io/acme/configs:v1.0.0 ./local
  blob pull foo:v1 ./local                          # Using alias
  blob pull --policy policy.yaml ghcr.io/acme/configs:v1.0.0
  blob pull --no-default-policy foo:v1 ./local      # Skip config policies
```

### `blob cp`

```
blob cp <ref>:<path>... <dest>

Copy files or directories from an archive to the local filesystem.
Uses HTTP range requests — does not download the full archive.

Arguments:
  <ref>:<path>   Source (reference or alias) and path within archive (can be repeated)
  <dest>         Local destination (file or directory)

Flags:
  -r, --recursive   Copy directories recursively (default: true for directories)
      --preserve    Preserve file permissions from archive

Behavior:
  - Single file to file:      blob cp reg/repo:v1:/config.json ./config.json
  - Single file to dir:       blob cp reg/repo:v1:/config.json ./output/
  - Multiple files to dir:    blob cp reg/repo:v1:/a.json reg/repo:v1:/b.json ./output/
  - Directory to directory:   blob cp reg/repo:v1:/etc/nginx ./nginx-config

Examples:
  blob cp ghcr.io/acme/configs:v1.0.0:/config.json ./config.json
  blob cp ghcr.io/acme/configs:v1.0.0:/etc/nginx/ ./nginx/
  blob cp ghcr.io/acme/configs:v1.0.0:/a.json ghcr.io/acme/configs:v1.0.0:/b.json ./
```

### `blob cat`

```
blob cat <ref> <file>...

Print file contents to stdout. Useful for viewing, piping, or combining files.
Uses HTTP range requests — does not download the full archive.

Arguments:
  <ref>       Source reference or alias
  <file>...   File path(s) within the archive

Flags:
  (none specific)

Examples:
  blob cat ghcr.io/acme/configs:v1.0.0 config.json
  blob cat ghcr.io/acme/configs:v1.0.0 config.json | jq .
  blob cat ghcr.io/acme/configs:v1.0.0 header.txt body.txt footer.txt > combined.txt
```

### `blob ls`

```
blob ls <ref> [path]

List files and directories in an archive.

Arguments:
  <ref>    Source reference
  [path]   Path within archive (default: root)

Flags:
  -l, --long      Long format (permissions, size, hash)
  -h, --human     Human-readable sizes (use with -l)
      --digest    Show file digests

Examples:
  blob ls ghcr.io/acme/configs:v1.0.0
  blob ls -lh ghcr.io/acme/configs:v1.0.0 /etc
```

### `blob inspect`

```
blob inspect <ref>

Show metadata about an archive without downloading it.

Arguments:
  <ref>    Source reference

Flags:
  (none specific)

Output includes:
  - Manifest digest
  - Total file count
  - Total size (compressed/uncompressed)
  - Compression type
  - Signatures (if any)
  - Attestations (if any)
  - Annotations

Examples:
  blob inspect ghcr.io/acme/configs:v1.0.0
  blob inspect --output json ghcr.io/acme/configs:v1.0.0
```

### `blob tree`

```
blob tree <ref> [path]

Display directory structure as a tree.

Arguments:
  <ref>    Source reference
  [path]   Path within archive (default: root)

Flags:
  -L, --level <n>    Descend only n levels deep
      --dirsfirst    List directories before files

Examples:
  blob tree ghcr.io/acme/configs:v1.0.0
  blob tree -L 2 ghcr.io/acme/configs:v1.0.0 /etc
```

### `blob sign`

```
blob sign <ref>

Sign an archive using Sigstore keyless signing.

Arguments:
  <ref>    Reference to sign

Flags:
      --key <path>           Sign with a private key instead of keyless
      --output-signature     Print signature to stdout instead of uploading

Examples:
  blob sign ghcr.io/acme/configs:v1.0.0
  blob sign --key cosign.key ghcr.io/acme/configs:v1.0.0
```

### `blob verify`

```
blob verify <ref>

Verify signatures and attestations on an archive.

Arguments:
  <ref>    Reference to verify

Flags:
      --policy <file>         Policy file for verification (can be repeated)
      --policy-rego <file>    OPA Rego policy file
      --policy-bundle <file>  OPA bundle for policy evaluation

Examples:
  blob verify ghcr.io/acme/configs:v1.0.0
  blob verify --policy policy.yaml ghcr.io/acme/configs:v1.0.0
  blob verify --policy-rego custom.rego ghcr.io/acme/configs:v1.0.0
```

### `blob tag`

```
blob tag <src-ref> <dst-ref>

Tag an existing manifest with a new reference.

Arguments:
  <src-ref>   Source reference (must exist)
  <dst-ref>   Destination reference (new tag)

Examples:
  blob tag ghcr.io/acme/configs:v1.0.0 ghcr.io/acme/configs:latest
  blob tag ghcr.io/acme/configs@sha256:abc... ghcr.io/acme/configs:stable
```

### `blob cache`

```
blob cache <subcommand>

Manage local caches.

Subcommands:
  status          Show cache sizes
  clear [type]    Clear caches
  path            Show cache directory paths

Cache types:
  content         File content cache (deduplicated across archives)
  manifests       OCI manifest cache
  indexes         Archive index cache
  all             All caches (default for clear)
```

### `blob cache status`

```
blob cache status

Show cache sizes for all cache types.

Example output:
  Cache Status
  ────────────────────────────────
  Content:     1.2 GB   (4,231 files)
  Manifests:   12 MB    (89 entries)
  Indexes:     8.4 MB   (89 entries)
  ────────────────────────────────
  Total:       1.2 GB

Example (JSON):
  blob cache status --output json
  {"content": {"size": 1288490188, "files": 4231}, ...}
```

### `blob cache clear`

```
blob cache clear [type]

Clear caches. Clears all caches by default.

Arguments:
  [type]    Cache type to clear: content, manifests, indexes, all (default: all)

Flags:
      --force    Skip confirmation prompt

Examples:
  blob cache clear              # Clear all caches (prompts for confirmation)
  blob cache clear --force      # Clear all without prompting
  blob cache clear content      # Clear only content cache
  blob cache clear manifests    # Clear only manifest cache
```

### `blob cache path`

```
blob cache path

Show cache directory paths.

Example output:
  Cache Paths
  ─────────────────────────────────────────────
  Root:       ~/.cache/blob/
  Content:    ~/.cache/blob/content/
  Manifests:  ~/.cache/blob/manifests/
  Indexes:    ~/.cache/blob/indexes/
```

### `blob alias`

```
blob alias <subcommand>

Manage reference aliases.

Subcommands:
  list              List all aliases
  set <name> <ref>  Add or update an alias
  remove <name>     Remove an alias
```

### `blob alias list`

```
blob alias list

List all configured aliases.

Example output:
  Aliases
  ───────────────────────────────────────
  foo    → ghcr.io/acme/repo/foo
  bar    → ghcr.io/acme/repo/bar
  baz    → ghcr.io/acme/repo/baz:stable
```

### `blob alias set`

```
blob alias set <name> <ref>

Add or update an alias. Writes to the config file.

Arguments:
  <name>   Alias name (short identifier)
  <ref>    Full reference (may include tag)

Examples:
  blob alias set foo ghcr.io/acme/repo/foo
  blob alias set prod ghcr.io/acme/repo/app:stable
```

### `blob alias remove`

```
blob alias remove <name>

Remove an alias from the config file.

Arguments:
  <name>   Alias name to remove

Examples:
  blob alias remove foo
```

### `blob config`

```
blob config <subcommand>

View and manage CLI configuration.

Subcommands:
  show    Display current configuration (merged from all sources)
  path    Show configuration file path
  edit    Open configuration file in $EDITOR

Examples:
  blob config show                 # Display current config
  blob config show --output json   # As JSON
  blob config path                 # Show config file location
  blob config edit                 # Open in editor
```

### `blob config show`

```
blob config show

Display the current configuration, merged from defaults, config file,
and environment variables. Shows effective values and their sources.

Flags:
      --resolved    Show fully resolved values (expand env vars)

Example output:
  Configuration
  ─────────────────────────────────────────────
  output:       text                 (default)
  compression:  zstd                 (config)
  cache:
    enabled:    true                 (env: BLOB_CACHE_ENABLED)
    max_size:   5GB                  (config)

  Aliases:
    foo → ghcr.io/acme/repo/foo

  Policies:
    ghcr\.io/acme/.* → signature (keyless), provenance (slsa)
```

### `blob config edit`

```
blob config edit

Open the configuration file in your default editor.
Uses $EDITOR, falling back to $VISUAL, then vi.

Creates the config file with defaults if it doesn't exist.
```

---

## Global Flags

All commands support these flags:

```
      --output <format>   Output format: text, json (default: text)
  -v, --verbose           Increase verbosity (can be repeated: -vv, -vvv)
  -q, --quiet             Suppress non-error output
      --no-color          Disable colored output
      --config <file>     Path to config file
```

## Reference Arguments

All commands that accept a `<ref>` argument support:

- **Full reference:** `ghcr.io/acme/repo:v1.0.0`
- **Aliases:** `foo` or `foo:v1` (expanded via config)
- **Digest references:** `ghcr.io/acme/repo@sha256:abc...`

See [Alias Resolution](#alias-resolution) for details.

---

## Configuration

### Precedence (highest to lowest)

1. Command-line flags
2. Environment variables (`BLOB_*`)
3. Config file
4. Defaults

### Environment Variables

| Variable | Description |
|----------|-------------|
| `BLOB_CONFIG` | Path to config file |
| `BLOB_OUTPUT` | Default output format |
| `BLOB_NO_COLOR` | Disable colors (also: `NO_COLOR`) |
| `BLOB_CACHE_DIR` | Cache directory |
| `BLOB_USERNAME` | Registry username |
| `BLOB_PASSWORD` | Registry password |

### XDG Paths

| Purpose | Path |
|---------|------|
| Config | `$XDG_CONFIG_HOME/blob/config.yaml` (default: `~/.config/blob/config.yaml`) |
| Cache | `$XDG_CACHE_HOME/blob/` (default: `~/.cache/blob/`) |
| Data | `$XDG_DATA_HOME/blob/` (default: `~/.local/share/blob/`) |

### Config File Format

```yaml
# ~/.config/blob/config.yaml

# Default output format
output: text

# Cache settings
cache:
  enabled: true
  max_size: 5GB

# Default compression for push
compression: zstd

# Aliases for frequently used references
# Usage: blob pull foo:v1 → ghcr.io/acme/repo/foo:v1
aliases:
  foo: ghcr.io/acme/repo/foo
  bar: ghcr.io/acme/repo/bar
  # Can include tag (blob pull baz → ghcr.io/acme/repo/baz:stable)
  baz: ghcr.io/acme/repo/baz:stable

# Default policies applied by image pattern (regex)
# Matched against fully-expanded reference (after alias resolution)
# Multiple patterns can match; all matching policies are combined (AND)
policies:
  - match: ghcr\.io/acme/.*
    policy:
      signature:
        keyless:
          issuer: https://token.actions.githubusercontent.com
          identity: https://github.com/acme/*/.github/workflows/*
      provenance:
        slsa:
          builder: https://github.com/slsa-framework/slsa-github-generator/.github/workflows/*
          repository: acme/*

  - match: ghcr\.io/acme/repo/prod-.*
    policy:
      # Additional requirements for prod images
      provenance:
        slsa:
          branch: main  # Prod must come from main branch
```

**Authentication:** Registry credentials are read from Docker's config (`~/.docker/config.json`) and credential helpers. For CI environments, use `BLOB_USERNAME` and `BLOB_PASSWORD` environment variables.

### Alias Resolution

Aliases expand short names to full references:

```bash
blob pull foo           # → ghcr.io/acme/repo/foo:latest
blob pull foo:v1        # → ghcr.io/acme/repo/foo:v1
blob pull foo@sha256:…  # → ghcr.io/acme/repo/foo@sha256:…
```

If the alias already includes a tag (e.g., `baz: .../baz:stable`), it's used as the default:

```bash
blob pull baz           # → ghcr.io/acme/repo/baz:stable
blob pull baz:v2        # → ghcr.io/acme/repo/baz:v2 (override)
```

### Policy Pattern Matching

Policies are matched against the **fully-expanded reference** (after alias resolution):

```bash
blob pull foo:v1
# 1. Expand alias: ghcr.io/acme/repo/foo:v1
# 2. Match against policy patterns
# 3. ghcr.io/acme/repo/foo:v1 matches "ghcr\.io/acme/.*" → policy applied
```

Multiple matching policies are combined with AND logic. Explicit `--policy` flags are added to (not replaced by) config policies.

To skip config policies for a single command:

```bash
blob pull --no-default-policy ghcr.io/acme/repo/foo:v1
```

---

## Policy Files

Blob supports two policy formats: a simple YAML format for common verification patterns, and OPA/Rego for complex custom logic.

### YAML Policy Format (recommended for most cases)

```yaml
# policy.yaml

# Require a Sigstore keyless signature
signature:
  keyless:
    issuer: https://token.actions.githubusercontent.com
    # identity supports wildcards
    identity: https://github.com/acme/configs/.github/workflows/*

# Or require a key-based signature
# signature:
#   key:
#     path: /path/to/cosign.pub
#   # or
#   key:
#     url: https://example.com/cosign.pub

# Require SLSA provenance
provenance:
  slsa:
    builder: https://github.com/slsa-framework/slsa-github-generator/.github/workflows/*
    repository: acme/configs
    # Optional: restrict to specific branches/tags
    branch: main
    # tag: v*
```

### Combining Multiple Policies

Multiple `--policy` flags are combined with AND logic:

```bash
# Both policies must pass
blob pull --policy sig.yaml --policy provenance.yaml ghcr.io/acme/configs:v1
```

For OR logic or more complex combinations, use OPA.

### OPA/Rego Policies (for complex cases)

For advanced use cases requiring custom logic:

```bash
# Single Rego file
blob verify --policy-rego policy.rego ghcr.io/acme/configs:v1

# OPA bundle (for larger policy sets)
blob verify --policy-bundle bundle.tar.gz ghcr.io/acme/configs:v1
```

Example Rego policy:

```rego
# policy.rego
package blob.verify

default allow = false

# Allow if signed by either release or security team
allow {
    some sig in input.signatures
    sig.issuer == "https://token.actions.githubusercontent.com"
    allowed_identities[sig.identity]
}

allowed_identities := {
    "https://github.com/acme/configs/.github/workflows/release.yaml@refs/heads/main",
    "https://github.com/acme/security/.github/workflows/sign.yaml@refs/heads/main",
}
```

### Policy Input Schema (for OPA)

OPA policies receive the following input:

```json
{
  "reference": "ghcr.io/acme/configs:v1.0.0",
  "digest": "sha256:abc123...",
  "signatures": [
    {
      "type": "sigstore-keyless",
      "issuer": "https://token.actions.githubusercontent.com",
      "identity": "https://github.com/acme/configs/.github/workflows/release.yaml@refs/heads/main",
      "timestamp": "2024-01-15T10:30:00Z"
    }
  ],
  "attestations": [
    {
      "type": "slsa-provenance-v1",
      "builder": "https://github.com/slsa-framework/slsa-github-generator/.github/workflows/generator_generic_slsa3.yml@refs/tags/v1.9.0",
      "repository": "acme/configs",
      "ref": "refs/tags/v1.0.0",
      "digest": "sha256:abc123..."
    }
  ],
  "annotations": {
    "org.opencontainers.image.source": "https://github.com/acme/configs"
  }
}
```

---

## Output Formats

### Text (default, TTY)

Human-friendly output with colors and formatting:

```
$ blob inspect ghcr.io/acme/configs:v1.0.0

Reference:    ghcr.io/acme/configs:v1.0.0
Digest:       sha256:abc123...
Files:        142
Size:         2.4 MB (8.1 MB uncompressed)
Compression:  zstd
Created:      2024-01-15T10:30:00Z

Signatures:
  ✓ Sigstore keyless (GitHub Actions)
    Identity: https://github.com/acme/configs/.github/workflows/release.yaml@refs/tags/v1.0.0
    Issuer:   https://token.actions.githubusercontent.com

Attestations:
  ✓ SLSA Provenance v1.0
    Builder: https://github.com/slsa-framework/slsa-github-generator
```

### Text (non-TTY)

Plain text without colors:

```
$ blob inspect ghcr.io/acme/configs:v1.0.0 | cat

Reference: ghcr.io/acme/configs:v1.0.0
Digest: sha256:abc123...
Files: 142
...
```

### JSON

Machine-readable output:

```
$ blob inspect --output json ghcr.io/acme/configs:v1.0.0

{
  "reference": "ghcr.io/acme/configs:v1.0.0",
  "digest": "sha256:abc123...",
  "files": 142,
  "size": {
    "compressed": 2400000,
    "uncompressed": 8100000
  },
  "compression": "zstd",
  "created": "2024-01-15T10:30:00Z",
  "signatures": [...],
  "attestations": [...]
}
```

---

## Error Handling

### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Usage error (bad arguments/flags) |
| 3 | Authentication error |
| 4 | Not found (reference doesn't exist) |
| 5 | Verification failed (policy violation) |

### Error Output

Errors go to stderr. In JSON mode, errors are also JSON:

```
$ blob pull ghcr.io/acme/nonexistent:v1
Error: reference not found: ghcr.io/acme/nonexistent:v1

$ blob pull --output json ghcr.io/acme/nonexistent:v1
{"error": "reference not found", "reference": "ghcr.io/acme/nonexistent:v1", "code": 4}
```

---

## Future Enhancements (not in v1)

- `blob login` / `blob logout` — Registry authentication management
- `blob copy` — Copy archives between registries (registry-to-registry)
- `blob diff` — Compare two archives
- Shell completions (`blob completion bash/zsh/fish`)
- Gittuf policy support — Source integrity verification (pending gittuf maturity)

---

## Open Questions

1. **Progress indicators** — Spinners? Progress bars? Both?
2. **Caching behavior** — On by default? Off by default? Per-command control?
