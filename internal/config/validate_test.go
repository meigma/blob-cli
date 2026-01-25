package config

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateOutput(t *testing.T) {
	tests := []struct {
		value   string
		wantErr bool
	}{
		{"text", false},
		{"json", false},
		{"xml", true},
		{"", true},
		{"TEXT", true}, // case sensitive
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			err := validateOutput(tt.value)
			if tt.wantErr {
				require.Error(t, err)
				assert.True(t, errors.Is(err, ErrInvalidConfig), "error should wrap ErrInvalidConfig")
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateCompression(t *testing.T) {
	tests := []struct {
		value   string
		wantErr bool
	}{
		{"none", false},
		{"zstd", false},
		{"gzip", true},
		{"", true},
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			err := validateCompression(tt.value)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateCacheSize(t *testing.T) {
	tests := []struct {
		value   string
		wantErr bool
	}{
		{"5GB", false},
		{"500MB", false},
		{"1TB", false},
		{"100KB", false},
		{"1024B", false},
		{"1024", false}, // bytes implied
		{"5.5GB", false},
		{"", false},       // empty is valid (use default)
		{"invalid", true}, // no number
		{"-5GB", true},    // negative
		{"5PB", true},     // invalid unit
		{"5 GB", false},   // space is trimmed
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			err := validateCacheSize(tt.value)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidatePolicies(t *testing.T) {
	tests := []struct {
		name     string
		policies []PolicyRule
		wantErr  bool
	}{
		{
			name:     "empty policies",
			policies: nil,
			wantErr:  false,
		},
		{
			name: "valid pattern",
			policies: []PolicyRule{
				{Match: `ghcr\.io/acme/.*`},
			},
			wantErr: false,
		},
		{
			name: "invalid regex",
			policies: []PolicyRule{
				{Match: "[invalid"},
			},
			wantErr: true,
		},
		{
			name: "empty match",
			policies: []PolicyRule{
				{Match: ""},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePolicies(tt.policies)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateCache(t *testing.T) {
	tests := []struct {
		name    string
		cache   CacheConfig
		wantErr bool
	}{
		{
			name:    "empty config",
			cache:   CacheConfig{},
			wantErr: false,
		},
		{
			name:    "valid ref_ttl",
			cache:   CacheConfig{RefTTL: "5m"},
			wantErr: false,
		},
		{
			name:    "valid ref_ttl with hours",
			cache:   CacheConfig{RefTTL: "1h30m"},
			wantErr: false,
		},
		{
			name:    "invalid ref_ttl",
			cache:   CacheConfig{RefTTL: "invalid"},
			wantErr: true,
		},
		{
			name:    "invalid ref_ttl no unit",
			cache:   CacheConfig{RefTTL: "5"},
			wantErr: true,
		},
		{
			name:    "valid max_size",
			cache:   CacheConfig{MaxSize: "5GB"},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateCache(&tt.cache)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *Config
		wantErr bool
	}{
		{
			name:    "valid default config",
			cfg:     Default(),
			wantErr: false,
		},
		{
			name: "invalid output",
			cfg: &Config{
				Output:      "xml",
				Compression: "zstd",
			},
			wantErr: true,
		},
		{
			name: "invalid compression",
			cfg: &Config{
				Output:      "text",
				Compression: "gzip",
			},
			wantErr: true,
		},
		{
			name: "invalid cache ref_ttl",
			cfg: &Config{
				Output:      "text",
				Compression: "zstd",
				Cache:       CacheConfig{RefTTL: "invalid"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validate(tt.cfg)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
