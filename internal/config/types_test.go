package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func ptr[T any](v T) *T {
	return &v
}

func TestCacheConfig_ContentEnabled(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		config CacheConfig
		want   bool
	}{
		{
			name:   "global disabled",
			config: CacheConfig{Enabled: false},
			want:   false,
		},
		{
			name:   "global enabled, no per-cache config",
			config: CacheConfig{Enabled: true},
			want:   true,
		},
		{
			name:   "global enabled, per-cache nil enabled",
			config: CacheConfig{Enabled: true, Content: &IndividualCacheConfig{}},
			want:   true,
		},
		{
			name:   "global enabled, per-cache explicitly enabled",
			config: CacheConfig{Enabled: true, Content: &IndividualCacheConfig{Enabled: ptr(true)}},
			want:   true,
		},
		{
			name:   "global enabled, per-cache explicitly disabled",
			config: CacheConfig{Enabled: true, Content: &IndividualCacheConfig{Enabled: ptr(false)}},
			want:   false,
		},
		{
			name:   "global disabled overrides per-cache enabled",
			config: CacheConfig{Enabled: false, Content: &IndividualCacheConfig{Enabled: ptr(true)}},
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.config.ContentEnabled()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCacheConfig_BlocksEnabled(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		config CacheConfig
		want   bool
	}{
		{
			name:   "global disabled",
			config: CacheConfig{Enabled: false},
			want:   false,
		},
		{
			name:   "global enabled, no per-cache config",
			config: CacheConfig{Enabled: true},
			want:   true,
		},
		{
			name:   "global enabled, per-cache explicitly disabled",
			config: CacheConfig{Enabled: true, Blocks: &IndividualCacheConfig{Enabled: ptr(false)}},
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.config.BlocksEnabled()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCacheConfig_RefsEnabled(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		config CacheConfig
		want   bool
	}{
		{
			name:   "global disabled",
			config: CacheConfig{Enabled: false},
			want:   false,
		},
		{
			name:   "global enabled, no per-cache config",
			config: CacheConfig{Enabled: true},
			want:   true,
		},
		{
			name:   "global enabled, per-cache explicitly disabled",
			config: CacheConfig{Enabled: true, Refs: &IndividualCacheConfig{Enabled: ptr(false)}},
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.config.RefsEnabled()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCacheConfig_ManifestsEnabled(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		config CacheConfig
		want   bool
	}{
		{
			name:   "global disabled",
			config: CacheConfig{Enabled: false},
			want:   false,
		},
		{
			name:   "global enabled, no per-cache config",
			config: CacheConfig{Enabled: true},
			want:   true,
		},
		{
			name:   "global enabled, per-cache explicitly disabled",
			config: CacheConfig{Enabled: true, Manifests: &IndividualCacheConfig{Enabled: ptr(false)}},
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.config.ManifestsEnabled()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCacheConfig_IndexesEnabled(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		config CacheConfig
		want   bool
	}{
		{
			name:   "global disabled",
			config: CacheConfig{Enabled: false},
			want:   false,
		},
		{
			name:   "global enabled, no per-cache config",
			config: CacheConfig{Enabled: true},
			want:   true,
		},
		{
			name:   "global enabled, per-cache explicitly disabled",
			config: CacheConfig{Enabled: true, Indexes: &IndividualCacheConfig{Enabled: ptr(false)}},
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.config.IndexesEnabled()
			assert.Equal(t, tt.want, got)
		})
	}
}
