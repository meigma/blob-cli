package config

// Config represents the complete blob-cli configuration.
type Config struct {
	// Output format: "text" or "json".
	Output string `mapstructure:"output" json:"output"`

	// Verbose level (0 = normal, 1+ = increasingly verbose).
	Verbose int `mapstructure:"verbose" json:"verbose"`

	// Quiet suppresses non-error output.
	Quiet bool `mapstructure:"quiet" json:"quiet"`

	// NoColor disables colored output.
	NoColor bool `mapstructure:"no-color" json:"no_color"`

	// PlainHTTP enables plain HTTP (no TLS) for registries.
	PlainHTTP bool `mapstructure:"plain-http" json:"plain_http"`

	// Compression type for push: "none" or "zstd".
	Compression string `mapstructure:"compression" json:"compression"`

	// Cache settings.
	Cache CacheConfig `mapstructure:"cache" json:"cache"`

	// Aliases map short names to full OCI references.
	Aliases map[string]string `mapstructure:"aliases" json:"aliases"`

	// Policies define verification requirements by reference pattern.
	Policies []PolicyRule `mapstructure:"policies" json:"policies,omitempty"`
}

// CacheConfig holds cache-related settings.
type CacheConfig struct {
	// Enabled controls whether caching is active.
	Enabled bool `mapstructure:"enabled" json:"enabled"`

	// MaxSize is the maximum cache size (e.g., "5GB", "500MB").
	MaxSize string `mapstructure:"max_size" json:"max_size"`
}

// PolicyRule maps a reference pattern to verification policies.
type PolicyRule struct {
	// Match is a regex pattern matched against fully-expanded references.
	Match string `mapstructure:"match" json:"match"`

	// Policy defines the verification requirements.
	Policy Policy `mapstructure:"policy" json:"policy"`
}

// Policy defines verification requirements for an archive.
type Policy struct {
	// Signature verification requirements.
	Signature *SignaturePolicy `mapstructure:"signature" json:"signature,omitempty"`

	// Provenance verification requirements.
	Provenance *ProvenancePolicy `mapstructure:"provenance" json:"provenance,omitempty"`
}

// SignaturePolicy defines signature verification requirements.
type SignaturePolicy struct {
	// Keyless defines Sigstore keyless verification.
	Keyless *KeylessConfig `mapstructure:"keyless" json:"keyless,omitempty"`

	// Key defines key-based signature verification.
	Key *KeyConfig `mapstructure:"key" json:"key,omitempty"`
}

// KeylessConfig defines Sigstore keyless verification requirements.
type KeylessConfig struct {
	// Issuer is the OIDC issuer URL (e.g., "https://token.actions.githubusercontent.com").
	Issuer string `mapstructure:"issuer" json:"issuer"`

	// Identity is the expected signer identity (supports wildcards with *).
	Identity string `mapstructure:"identity" json:"identity"`
}

// KeyConfig defines key-based signature verification.
type KeyConfig struct {
	// Path to a local public key file.
	Path string `mapstructure:"path" json:"path,omitempty"`

	// URL to fetch the public key from.
	URL string `mapstructure:"url" json:"url,omitempty"`
}

// ProvenancePolicy defines provenance verification requirements.
type ProvenancePolicy struct {
	// SLSA defines SLSA provenance requirements.
	SLSA *SLSAConfig `mapstructure:"slsa" json:"slsa,omitempty"`
}

// SLSAConfig defines SLSA provenance requirements.
type SLSAConfig struct {
	// Builder is the expected builder identity (supports wildcards).
	Builder string `mapstructure:"builder" json:"builder,omitempty"`

	// Repository is the expected source repository (supports wildcards).
	Repository string `mapstructure:"repository" json:"repository,omitempty"`

	// Branch restricts to builds from a specific branch.
	Branch string `mapstructure:"branch" json:"branch,omitempty"`

	// Tag restricts to builds from a specific tag pattern.
	Tag string `mapstructure:"tag" json:"tag,omitempty"`
}
