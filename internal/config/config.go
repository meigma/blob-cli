package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"
)

// Load reads configuration from the provided Viper instance and returns a typed Config.
// This should be called after Viper has been initialized and all sources
// (flags, env, file) have been loaded.
func Load(v *viper.Viper) (*Config, error) {
	cfg := &Config{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("unmarshaling config: %w", err)
	}

	// Ensure aliases map is initialized
	if cfg.Aliases == nil {
		cfg.Aliases = make(map[string]string)
	}

	if err := validate(cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

// LoadFromViper loads config using the global Viper instance.
func LoadFromViper() (*Config, error) {
	return Load(viper.GetViper())
}

// Save writes the config to the specified path as YAML.
// Creates parent directories if they don't exist.
func Save(cfg *Config, path string) error {
	// Ensure parent directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	// Marshal to YAML
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	// Write file with appropriate permissions
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("writing config file: %w", err)
	}

	return nil
}

// SaveDefault creates a config file at path with default values.
// Creates parent directories if they don't exist.
func SaveDefault(path string) error {
	return Save(Default(), path)
}

// defaultConfigTemplate is the template for a new config file with comments.
const defaultConfigTemplate = `# blob-cli configuration file
# See: https://github.com/meigma/blob-cli

# Default output format: text, json
output: text

# Default compression for push: none, zstd
compression: zstd

# Cache settings
cache:
  enabled: true
  max_size: 5GB

# Aliases for frequently used references
# Usage: blob pull foo:v1 → ghcr.io/acme/repo/foo:v1
aliases: {}
  # foo: ghcr.io/acme/repo/foo
  # bar: ghcr.io/acme/repo/bar

# Default policies applied by image pattern (regex)
# Matched against fully-expanded reference (after alias resolution)
# Multiple patterns can match; all matching policies are combined (AND)
policies: []
  # - match: ghcr\.io/acme/.*
  #   policy:
  #     signature:
  #       keyless:
  #         issuer: https://token.actions.githubusercontent.com
  #         identity: https://github.com/acme/*/.github/workflows/*
`

// SaveDefaultWithComments creates a config file at path with default values
// and helpful comments. Creates parent directories if they don't exist.
func SaveDefaultWithComments(path string) error {
	// Ensure parent directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	// Write the template with comments
	if err := os.WriteFile(path, []byte(defaultConfigTemplate), 0o600); err != nil {
		return fmt.Errorf("writing config file: %w", err)
	}

	return nil
}
