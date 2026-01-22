package config

import (
	"maps"
	"strings"
)

// ResolveAlias expands an alias to a full reference.
// If name is not an alias, returns it unchanged.
//
// Alias resolution handles tag overrides:
//   - "alias" with alias "foo: ghcr.io/acme/foo" → "ghcr.io/acme/foo:latest"
//   - "alias:v1" with alias "foo: ghcr.io/acme/foo" → "ghcr.io/acme/foo:v1"
//   - "alias" with alias "foo: ghcr.io/acme/foo:stable" → "ghcr.io/acme/foo:stable"
//   - "alias:v1" with alias "foo: ghcr.io/acme/foo:stable" → "ghcr.io/acme/foo:v1" (override)
func (c *Config) ResolveAlias(name string) string {
	if c.Aliases == nil {
		return name
	}

	aliasName, tagOrDigest, hasTagOrDigest := parseRef(name)

	// Look up the alias
	ref, ok := c.Aliases[aliasName]
	if !ok {
		// Not an alias, return unchanged
		return name
	}

	// If the user provided a tag/digest, use it (override alias default)
	if hasTagOrDigest {
		// Strip any existing tag from the alias ref and use the provided one
		baseRef, _, _ := parseRef(ref)
		return baseRef + tagOrDigest
	}

	// No tag provided by user
	// If alias has a tag, use it; otherwise default to :latest
	_, _, hasAliasTag := parseRef(ref)
	if hasAliasTag {
		return ref
	}

	return ref + ":latest"
}

// SetAlias returns a new Config with the alias added or updated.
// The original Config is not modified.
func (c *Config) SetAlias(name, ref string) *Config {
	newCfg := c.clone()
	if newCfg.Aliases == nil {
		newCfg.Aliases = make(map[string]string)
	}
	newCfg.Aliases[name] = ref
	return newCfg
}

// RemoveAlias returns a new Config without the specified alias.
// The original Config is not modified.
// If the alias doesn't exist, returns a copy of the original.
func (c *Config) RemoveAlias(name string) *Config {
	newCfg := c.clone()
	delete(newCfg.Aliases, name)
	return newCfg
}

// clone creates a shallow copy of the Config with a deep copy of maps/slices.
func (c *Config) clone() *Config {
	newCfg := *c

	// Deep copy aliases map
	if c.Aliases != nil {
		newCfg.Aliases = maps.Clone(c.Aliases)
	}

	// Deep copy policies slice
	if c.Policies != nil {
		newCfg.Policies = make([]PolicyRule, len(c.Policies))
		copy(newCfg.Policies, c.Policies)
	}

	return &newCfg
}

// parseRef splits a reference into base and tag/digest components.
// Returns: (base, tagOrDigest including separator, hasTagOrDigest)
//
// Examples:
//   - "foo" → ("foo", "", false)
//   - "foo:v1" → ("foo", ":v1", true)
//   - "foo@sha256:abc" → ("foo", "@sha256:abc", true)
//   - "ghcr.io/acme/repo:v1" → ("ghcr.io/acme/repo", ":v1", true)
func parseRef(ref string) (base, tagOrDigest string, hasTagOrDigest bool) {
	// Check for digest first (@ takes precedence in OCI refs)
	if idx := strings.LastIndex(ref, "@"); idx != -1 {
		return ref[:idx], ref[idx:], true
	}

	// Check for tag
	// Need to handle registry:port/path:tag correctly
	// The tag is after the last colon that comes after any slash
	lastSlash := strings.LastIndex(ref, "/")
	lastColon := strings.LastIndex(ref, ":")

	// If there's a colon after the last slash (or no slash), it's a tag
	if lastColon > lastSlash {
		return ref[:lastColon], ref[lastColon:], true
	}

	return ref, "", false
}
