package config

import (
	"testing"
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
			if got != tt.want {
				t.Errorf("ResolveAlias(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestConfig_SetAlias(t *testing.T) {
	cfg := &Config{
		Aliases: map[string]string{"existing": "ghcr.io/acme/existing"},
	}

	// Set a new alias
	newCfg := cfg.SetAlias("new", "ghcr.io/acme/new")

	// Verify new config has the alias
	if newCfg.Aliases["new"] != "ghcr.io/acme/new" {
		t.Errorf("new config missing alias: got %v", newCfg.Aliases)
	}

	// Verify original is unchanged (immutability)
	if _, ok := cfg.Aliases["new"]; ok {
		t.Error("original config was modified")
	}

	// Verify existing alias preserved
	if newCfg.Aliases["existing"] != "ghcr.io/acme/existing" {
		t.Error("existing alias was lost")
	}
}

func TestConfig_SetAlias_NilMap(t *testing.T) {
	cfg := &Config{Aliases: nil}
	newCfg := cfg.SetAlias("foo", "ghcr.io/acme/foo")

	if newCfg.Aliases["foo"] != "ghcr.io/acme/foo" {
		t.Errorf("alias not set: got %v", newCfg.Aliases)
	}
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
	if _, ok := newCfg.Aliases["foo"]; ok {
		t.Error("alias not removed from new config")
	}

	// Verify other alias preserved
	if newCfg.Aliases["bar"] != "ghcr.io/acme/bar" {
		t.Error("other alias was lost")
	}

	// Verify original unchanged
	if cfg.Aliases["foo"] != "ghcr.io/acme/foo" {
		t.Error("original config was modified")
	}
}

func TestConfig_RemoveAlias_NonExistent(t *testing.T) {
	cfg := &Config{
		Aliases: map[string]string{"foo": "ghcr.io/acme/foo"},
	}

	newCfg := cfg.RemoveAlias("nonexistent")

	// Should not panic and should preserve existing aliases
	if newCfg.Aliases["foo"] != "ghcr.io/acme/foo" {
		t.Error("existing alias was lost")
	}
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
			if base != tt.wantBase {
				t.Errorf("parseRef(%q) base = %q, want %q", tt.input, base, tt.wantBase)
			}
			if tag != tt.wantTag {
				t.Errorf("parseRef(%q) tag = %q, want %q", tt.input, tag, tt.wantTag)
			}
			if hasTag != tt.wantHasTagOrD {
				t.Errorf("parseRef(%q) hasTag = %v, want %v", tt.input, hasTag, tt.wantHasTagOrD)
			}
		})
	}
}
