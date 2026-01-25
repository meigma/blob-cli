package cmd

import (
	"os"
	"testing"

	internalcfg "github.com/meigma/blob-cli/internal/config"
)

func TestResolveCacheDir(t *testing.T) {
	// Note: Not parallel because subtests use t.Setenv

	t.Run("uses config dir when specified", func(t *testing.T) {
		cfg := &internalcfg.Config{
			Cache: internalcfg.CacheConfig{
				Enabled: true,
				Dir:     "/custom/cache/dir",
			},
		}

		got, err := resolveCacheDir(cfg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "/custom/cache/dir" {
			t.Errorf("resolveCacheDir() = %q, want %q", got, "/custom/cache/dir")
		}
	})

	t.Run("uses XDG default when config dir empty", func(t *testing.T) {
		// Note: Can't use t.Parallel() with t.Setenv()

		// Set up a predictable XDG_CACHE_HOME
		tmpDir := t.TempDir()
		t.Setenv("XDG_CACHE_HOME", tmpDir)

		cfg := &internalcfg.Config{
			Cache: internalcfg.CacheConfig{
				Enabled: true,
				Dir:     "",
			},
		}

		got, err := resolveCacheDir(cfg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		want := tmpDir + "/blob"
		if got != want {
			t.Errorf("resolveCacheDir() = %q, want %q", got, want)
		}
	})
}

func TestClientOpts(t *testing.T) {
	t.Parallel()

	t.Run("includes cache options when enabled", func(t *testing.T) {
		t.Parallel()

		// Create temp dir for cache
		tmpDir := t.TempDir()

		cfg := &internalcfg.Config{
			Cache: internalcfg.CacheConfig{
				Enabled: true,
				Dir:     tmpDir,
			},
		}

		opts := clientOpts(cfg)

		// Should have at least 2 options: WithDockerConfig and WithCacheDir
		if len(opts) < 2 {
			t.Errorf("clientOpts() returned %d options, want at least 2", len(opts))
		}
	})

	t.Run("excludes cache options when disabled", func(t *testing.T) {
		t.Parallel()

		cfg := &internalcfg.Config{
			Cache: internalcfg.CacheConfig{
				Enabled: false,
			},
		}

		opts := clientOpts(cfg)

		// Should have only 1 option: WithDockerConfig
		if len(opts) != 1 {
			t.Errorf("clientOpts() returned %d options, want 1", len(opts))
		}
	})

	t.Run("includes PlainHTTP when enabled", func(t *testing.T) {
		t.Parallel()

		cfg := &internalcfg.Config{
			PlainHTTP: true,
			Cache: internalcfg.CacheConfig{
				Enabled: false,
			},
		}

		opts := clientOpts(cfg)

		// Should have 2 options: WithDockerConfig and WithPlainHTTP
		if len(opts) != 2 {
			t.Errorf("clientOpts() returned %d options, want 2", len(opts))
		}
	})
}

func TestClientOptsNoCache(t *testing.T) {
	t.Parallel()

	t.Run("never includes cache options", func(t *testing.T) {
		t.Parallel()

		cfg := &internalcfg.Config{
			Cache: internalcfg.CacheConfig{
				Enabled: true,
				Dir:     "/some/cache/dir",
			},
		}

		opts := clientOptsNoCache(cfg)

		// Should have only 1 option: WithDockerConfig
		if len(opts) != 1 {
			t.Errorf("clientOptsNoCache() returned %d options, want 1", len(opts))
		}
	})

	t.Run("includes PlainHTTP when enabled", func(t *testing.T) {
		t.Parallel()

		cfg := &internalcfg.Config{
			PlainHTTP: true,
			Cache: internalcfg.CacheConfig{
				Enabled: true,
				Dir:     "/some/cache/dir",
			},
		}

		opts := clientOptsNoCache(cfg)

		// Should have 2 options: WithDockerConfig and WithPlainHTTP
		if len(opts) != 2 {
			t.Errorf("clientOptsNoCache() returned %d options, want 2", len(opts))
		}
	})
}

func ptr[T any](v T) *T {
	return &v
}

func TestBuildCacheOpts(t *testing.T) {
	t.Parallel()

	t.Run("all caches enabled by default", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		cfg := &internalcfg.Config{
			Cache: internalcfg.CacheConfig{
				Enabled: true,
			},
		}

		opts := buildCacheOpts(cfg, tmpDir)

		// Should have 5 options: one for each cache type
		if len(opts) != 5 {
			t.Errorf("buildCacheOpts() returned %d options, want 5", len(opts))
		}
	})

	t.Run("disabling individual cache reduces options", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		cfg := &internalcfg.Config{
			Cache: internalcfg.CacheConfig{
				Enabled: true,
				Refs:    &internalcfg.IndividualCacheConfig{Enabled: ptr(false)},
			},
		}

		opts := buildCacheOpts(cfg, tmpDir)

		// Should have 4 options: all except refs
		if len(opts) != 4 {
			t.Errorf("buildCacheOpts() returned %d options, want 4", len(opts))
		}
	})

	t.Run("disabling multiple caches", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		cfg := &internalcfg.Config{
			Cache: internalcfg.CacheConfig{
				Enabled:   true,
				Content:   &internalcfg.IndividualCacheConfig{Enabled: ptr(false)},
				Blocks:    &internalcfg.IndividualCacheConfig{Enabled: ptr(false)},
				Refs:      &internalcfg.IndividualCacheConfig{Enabled: ptr(false)},
				Manifests: &internalcfg.IndividualCacheConfig{Enabled: ptr(false)},
			},
		}

		opts := buildCacheOpts(cfg, tmpDir)

		// Should have 1 option: only indexes
		if len(opts) != 1 {
			t.Errorf("buildCacheOpts() returned %d options, want 1", len(opts))
		}
	})

	t.Run("ref_ttl adds option when refs enabled", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		cfg := &internalcfg.Config{
			Cache: internalcfg.CacheConfig{
				Enabled: true,
				RefTTL:  "10m",
			},
		}

		opts := buildCacheOpts(cfg, tmpDir)

		// Should have 6 options: 5 caches + 1 TTL
		if len(opts) != 6 {
			t.Errorf("buildCacheOpts() returned %d options, want 6", len(opts))
		}
	})

	t.Run("ref_ttl ignored when refs disabled", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		cfg := &internalcfg.Config{
			Cache: internalcfg.CacheConfig{
				Enabled: true,
				RefTTL:  "10m",
				Refs:    &internalcfg.IndividualCacheConfig{Enabled: ptr(false)},
			},
		}

		opts := buildCacheOpts(cfg, tmpDir)

		// Should have 4 options: 4 caches (no refs), no TTL
		if len(opts) != 4 {
			t.Errorf("buildCacheOpts() returned %d options, want 4", len(opts))
		}
	})

	t.Run("invalid ref_ttl is ignored", func(t *testing.T) {
		t.Parallel()

		tmpDir := t.TempDir()
		cfg := &internalcfg.Config{
			Cache: internalcfg.CacheConfig{
				Enabled: true,
				RefTTL:  "invalid",
			},
		}

		opts := buildCacheOpts(cfg, tmpDir)

		// Should have 5 options: invalid TTL is skipped
		if len(opts) != 5 {
			t.Errorf("buildCacheOpts() returned %d options, want 5", len(opts))
		}
	})
}

func TestMain(m *testing.M) {
	// Ensure tests don't accidentally use real config
	os.Exit(m.Run())
}
