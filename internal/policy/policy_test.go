package policy

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/meigma/blob-cli/internal/config"
)

func TestLoadFile(t *testing.T) {
	t.Run("keyless signature policy", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "policy.yaml")
		content := `
signature:
  keyless:
    issuer: https://token.actions.githubusercontent.com
    identity: https://github.com/acme/*/.github/workflows/*
`
		err := os.WriteFile(path, []byte(content), 0o644)
		require.NoError(t, err)

		policy, err := LoadFile(path)
		require.NoError(t, err)
		require.NotNil(t, policy)
		require.NotNil(t, policy.Signature)
		require.NotNil(t, policy.Signature.Keyless)
		assert.Equal(t, "https://token.actions.githubusercontent.com", policy.Signature.Keyless.Issuer)
		assert.Equal(t, "https://github.com/acme/*/.github/workflows/*", policy.Signature.Keyless.Identity)
	})

	t.Run("SLSA provenance policy", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "policy.yaml")
		content := `
provenance:
  slsa:
    repository: acme/configs
    branch: main
`
		err := os.WriteFile(path, []byte(content), 0o644)
		require.NoError(t, err)

		policy, err := LoadFile(path)
		require.NoError(t, err)
		require.NotNil(t, policy)
		require.NotNil(t, policy.Provenance)
		require.NotNil(t, policy.Provenance.SLSA)
		assert.Equal(t, "acme/configs", policy.Provenance.SLSA.Repository)
		assert.Equal(t, "main", policy.Provenance.SLSA.Branch)
	})

	t.Run("combined signature and provenance", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "policy.yaml")
		content := `
signature:
  keyless:
    issuer: https://token.actions.githubusercontent.com
    identity: https://github.com/acme/*/.github/workflows/*
provenance:
  slsa:
    repository: acme/configs
    branch: main
    tag: v*
`
		err := os.WriteFile(path, []byte(content), 0o644)
		require.NoError(t, err)

		policy, err := LoadFile(path)
		require.NoError(t, err)
		require.NotNil(t, policy)
		require.NotNil(t, policy.Signature)
		require.NotNil(t, policy.Provenance)
	})

	t.Run("file not found", func(t *testing.T) {
		_, err := LoadFile("/nonexistent/policy.yaml")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "reading policy file")
	})

	t.Run("invalid yaml", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "policy.yaml")
		err := os.WriteFile(path, []byte("not: valid: yaml: ["), 0o644)
		require.NoError(t, err)

		_, err = LoadFile(path)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "parsing policy file")
	})
}

func TestConvertConfigPolicy(t *testing.T) {
	t.Run("empty policy", func(t *testing.T) {
		policy, err := ConvertConfigPolicy(config.Policy{})
		require.NoError(t, err)
		assert.Nil(t, policy)
	})

	t.Run("keyless signature only", func(t *testing.T) {
		cfgPolicy := config.Policy{
			Signature: &config.SignaturePolicy{
				Keyless: &config.KeylessConfig{
					Issuer:   "https://token.actions.githubusercontent.com",
					Identity: "https://github.com/acme/*/.github/workflows/*",
				},
			},
		}
		policy, err := ConvertConfigPolicy(cfgPolicy)
		require.NoError(t, err)
		assert.NotNil(t, policy)
	})

	t.Run("SLSA provenance with repository", func(t *testing.T) {
		cfgPolicy := config.Policy{
			Provenance: &config.ProvenancePolicy{
				SLSA: &config.SLSAConfig{
					Repository: "acme/configs",
					Branch:     "main",
				},
			},
		}
		policy, err := ConvertConfigPolicy(cfgPolicy)
		require.NoError(t, err)
		assert.NotNil(t, policy)
	})

	t.Run("SLSA provenance with builder only", func(t *testing.T) {
		cfgPolicy := config.Policy{
			Provenance: &config.ProvenancePolicy{
				SLSA: &config.SLSAConfig{
					Builder: "https://github.com/slsa-framework/slsa-github-generator",
				},
			},
		}
		policy, err := ConvertConfigPolicy(cfgPolicy)
		require.NoError(t, err)
		assert.NotNil(t, policy)
	})

	t.Run("missing keyless issuer", func(t *testing.T) {
		cfgPolicy := config.Policy{
			Signature: &config.SignaturePolicy{
				Keyless: &config.KeylessConfig{
					Identity: "https://github.com/acme/*/.github/workflows/*",
				},
			},
		}
		_, err := ConvertConfigPolicy(cfgPolicy)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "keyless issuer is required")
	})

	t.Run("missing keyless identity", func(t *testing.T) {
		cfgPolicy := config.Policy{
			Signature: &config.SignaturePolicy{
				Keyless: &config.KeylessConfig{
					Issuer: "https://token.actions.githubusercontent.com",
				},
			},
		}
		_, err := ConvertConfigPolicy(cfgPolicy)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "keyless identity is required")
	})

	t.Run("key path not implemented", func(t *testing.T) {
		cfgPolicy := config.Policy{
			Signature: &config.SignaturePolicy{
				Key: &config.KeyConfig{
					Path: "/path/to/key.pub",
				},
			},
		}
		_, err := ConvertConfigPolicy(cfgPolicy)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not yet implemented")
	})

	t.Run("both keyless and key specified", func(t *testing.T) {
		cfgPolicy := config.Policy{
			Signature: &config.SignaturePolicy{
				Keyless: &config.KeylessConfig{
					Issuer:   "https://token.actions.githubusercontent.com",
					Identity: "https://github.com/acme/*/.github/workflows/*",
				},
				Key: &config.KeyConfig{
					Path: "/path/to/key.pub",
				},
			},
		}
		_, err := ConvertConfigPolicy(cfgPolicy)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot specify both keyless and key")
	})

	t.Run("SLSA missing repository and builder", func(t *testing.T) {
		cfgPolicy := config.Policy{
			Provenance: &config.ProvenancePolicy{
				SLSA: &config.SLSAConfig{
					Branch: "main",
				},
			},
		}
		_, err := ConvertConfigPolicy(cfgPolicy)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "must specify repository or builder")
	})
}

func TestBuildPolicies(t *testing.T) {
	t.Run("no policies when all disabled", func(t *testing.T) {
		cfg := &config.Config{}
		policies, err := BuildPolicies(cfg, "ghcr.io/test:v1", nil, "", true)
		require.NoError(t, err)
		assert.Empty(t, policies)
	})

	t.Run("nil config with no default policy", func(t *testing.T) {
		policies, err := BuildPolicies(nil, "ghcr.io/test:v1", nil, "", true)
		require.NoError(t, err)
		assert.Empty(t, policies)
	})

	t.Run("policy file loading", func(t *testing.T) {
		dir := t.TempDir()
		path := filepath.Join(dir, "policy.yaml")
		content := `
provenance:
  slsa:
    repository: acme/configs
`
		err := os.WriteFile(path, []byte(content), 0o644)
		require.NoError(t, err)

		policies, err := BuildPolicies(nil, "ghcr.io/test:v1", []string{path}, "", true)
		require.NoError(t, err)
		assert.Len(t, policies, 1)
	})

	t.Run("invalid policy file", func(t *testing.T) {
		_, err := BuildPolicies(nil, "ghcr.io/test:v1", []string{"/nonexistent.yaml"}, "", true)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "loading policy")
	})

	t.Run("config policies when not disabled", func(t *testing.T) {
		cfg := &config.Config{
			Policies: []config.PolicyRule{
				{
					Match: "ghcr\\.io/test/.*",
					Policy: config.Policy{
						Provenance: &config.ProvenancePolicy{
							SLSA: &config.SLSAConfig{
								Repository: "test/repo",
							},
						},
					},
				},
			},
		}
		policies, err := BuildPolicies(cfg, "ghcr.io/test/app:v1", nil, "", false)
		require.NoError(t, err)
		assert.Len(t, policies, 1)
	})

	t.Run("config policies skipped when disabled", func(t *testing.T) {
		cfg := &config.Config{
			Policies: []config.PolicyRule{
				{
					Match: "ghcr\\.io/test/.*",
					Policy: config.Policy{
						Provenance: &config.ProvenancePolicy{
							SLSA: &config.SLSAConfig{
								Repository: "test/repo",
							},
						},
					},
				},
			},
		}
		policies, err := BuildPolicies(cfg, "ghcr.io/test/app:v1", nil, "", true)
		require.NoError(t, err)
		assert.Empty(t, policies)
	})
}
