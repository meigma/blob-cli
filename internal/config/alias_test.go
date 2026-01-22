package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfig_ResolveAlias(t *testing.T) {
	tests := []struct {
		name    string
		aliases map[string]string
		input   string
		want    string
	}{
		{
			name:    "simple alias adds latest tag",
			aliases: map[string]string{"foo": "ghcr.io/acme/foo"},
			input:   "foo",
			want:    "ghcr.io/acme/foo:latest",
		},
		{
			name:    "alias with tag override",
			aliases: map[string]string{"foo": "ghcr.io/acme/foo"},
			input:   "foo:v1",
			want:    "ghcr.io/acme/foo:v1",
		},
		{
			name:    "alias with default tag",
			aliases: map[string]string{"foo": "ghcr.io/acme/foo:stable"},
			input:   "foo",
			want:    "ghcr.io/acme/foo:stable",
		},
		{
			name:    "alias with default tag override",
			aliases: map[string]string{"foo": "ghcr.io/acme/foo:stable"},
			input:   "foo:v2",
			want:    "ghcr.io/acme/foo:v2",
		},
		{
			name:    "not an alias passthrough",
			aliases: map[string]string{"foo": "ghcr.io/acme/foo"},
			input:   "ghcr.io/other/repo:v1",
			want:    "ghcr.io/other/repo:v1",
		},
		{
			name:    "alias with digest override",
			aliases: map[string]string{"foo": "ghcr.io/acme/foo"},
			input:   "foo@sha256:abc123",
			want:    "ghcr.io/acme/foo@sha256:abc123",
		},
		{
			name:    "nil aliases map",
			aliases: nil,
			input:   "foo",
			want:    "foo",
		},
		{
			name:    "empty aliases map",
			aliases: map[string]string{},
			input:   "foo",
			want:    "foo",
		},
		{
			name:    "registry with port not confused as tag",
			aliases: map[string]string{},
			input:   "localhost:5000/repo:v1",
			want:    "localhost:5000/repo:v1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &Config{Aliases: tt.aliases}
			got := cfg.ResolveAlias(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestConfig_SetAlias(t *testing.T) {
	cfg := &Config{
		Aliases: map[string]string{"existing": "ghcr.io/acme/existing"},
	}

	newCfg := cfg.SetAlias("new", "ghcr.io/acme/new")

	// Verify new config has the alias
	assert.Equal(t, "ghcr.io/acme/new", newCfg.Aliases["new"])

	// Verify original is unchanged (immutability)
	_, ok := cfg.Aliases["new"]
	assert.False(t, ok, "original config should not be modified")

	// Verify existing alias preserved
	assert.Equal(t, "ghcr.io/acme/existing", newCfg.Aliases["existing"])
}

func TestConfig_SetAlias_NilMap(t *testing.T) {
	cfg := &Config{Aliases: nil}
	newCfg := cfg.SetAlias("foo", "ghcr.io/acme/foo")

	assert.Equal(t, "ghcr.io/acme/foo", newCfg.Aliases["foo"])
}

func TestConfig_RemoveAlias(t *testing.T) {
	cfg := &Config{
		Aliases: map[string]string{
			"foo": "ghcr.io/acme/foo",
			"bar": "ghcr.io/acme/bar",
		},
	}

	newCfg := cfg.RemoveAlias("foo")

	// Verify alias removed from new config
	_, ok := newCfg.Aliases["foo"]
	assert.False(t, ok, "alias should be removed from new config")

	// Verify other alias preserved
	assert.Equal(t, "ghcr.io/acme/bar", newCfg.Aliases["bar"])

	// Verify original unchanged
	assert.Equal(t, "ghcr.io/acme/foo", cfg.Aliases["foo"])
}

func TestConfig_RemoveAlias_NonExistent(t *testing.T) {
	cfg := &Config{
		Aliases: map[string]string{"foo": "ghcr.io/acme/foo"},
	}

	newCfg := cfg.RemoveAlias("nonexistent")

	// Should not panic and should preserve existing aliases
	assert.Equal(t, "ghcr.io/acme/foo", newCfg.Aliases["foo"])
}

func TestParseRef(t *testing.T) {
	tests := []struct {
		input         string
		wantBase      string
		wantTag       string
		wantHasTagOrD bool
	}{
		{"foo", "foo", "", false},
		{"foo:v1", "foo", ":v1", true},
		{"foo@sha256:abc", "foo", "@sha256:abc", true},
		{"ghcr.io/acme/repo", "ghcr.io/acme/repo", "", false},
		{"ghcr.io/acme/repo:v1", "ghcr.io/acme/repo", ":v1", true},
		{"ghcr.io/acme/repo@sha256:abc", "ghcr.io/acme/repo", "@sha256:abc", true},
		{"localhost:5000/repo", "localhost:5000/repo", "", false},
		{"localhost:5000/repo:v1", "localhost:5000/repo", ":v1", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			base, tag, hasTag := parseRef(tt.input)
			require.Equal(t, tt.wantBase, base, "base mismatch")
			require.Equal(t, tt.wantTag, tag, "tag mismatch")
			require.Equal(t, tt.wantHasTagOrD, hasTag, "hasTag mismatch")
		})
	}
}
