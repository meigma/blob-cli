// Package config provides typed configuration management for blob-cli.
//
// This package wraps Viper to provide type-safe access to configuration
// values loaded from multiple sources (flags, environment, config file).
//
// # Configuration Precedence
//
// Configuration values are resolved in the following order (highest priority first):
//  1. Command-line flags
//  2. Environment variables (prefixed with BLOB_)
//  3. Config file ($XDG_CONFIG_HOME/blob/config.yaml)
//  4. Built-in defaults
//
// # Context Integration
//
// The configuration is passed to commands via context.Context:
//
//	cfg, err := config.LoadFromViper()
//	if err != nil {
//	    return err
//	}
//	ctx := config.WithConfig(ctx, cfg)
//
// Commands retrieve the configuration using:
//
//	cfg := config.FromContext(ctx)
//
// # Alias Resolution
//
// Aliases map short names to full OCI references:
//
//	cfg.ResolveAlias("foo")       // -> "ghcr.io/acme/foo:latest"
//	cfg.ResolveAlias("foo:v1")    // -> "ghcr.io/acme/foo:v1"
//
// # Policy Matching
//
// Policies are matched against fully-expanded references using regex patterns:
//
//	policies := cfg.GetPoliciesForRef("ghcr.io/acme/repo:v1")
package config
