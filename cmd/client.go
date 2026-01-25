package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/meigma/blob"

	internalcfg "github.com/meigma/blob-cli/internal/config"
)

// newClient creates a new blob client with options from config.
func newClient(cfg *internalcfg.Config, opts ...blob.Option) (*blob.Client, error) {
	baseOpts := clientOpts(cfg)
	baseOpts = append(baseOpts, opts...)
	return blob.NewClient(baseOpts...)
}

// clientOpts returns the base client options from config.
// This is useful when passing options to functions that create their own client.
// If caching is enabled but the cache directory cannot be resolved, a warning
// is written to stderr and caching is disabled for this operation.
func clientOpts(cfg *internalcfg.Config) []blob.Option {
	opts := []blob.Option{blob.WithDockerConfig()}
	if cfg.PlainHTTP {
		opts = append(opts, blob.WithPlainHTTP(true))
	}
	if cfg.Cache.Enabled {
		cacheDir, err := resolveCacheDir(cfg)
		if err != nil {
			if !cfg.Quiet {
				fmt.Fprintf(os.Stderr, "Warning: cache disabled: %v\n", err)
			}
		} else {
			opts = append(opts, buildCacheOpts(cfg, cacheDir)...)
		}
	}
	return opts
}

// buildCacheOpts returns cache options based on config.
// Each cache type is enabled individually based on the config settings.
func buildCacheOpts(cfg *internalcfg.Config, cacheDir string) []blob.Option {
	var opts []blob.Option
	cache := &cfg.Cache

	if cache.ContentEnabled() {
		opts = append(opts, blob.WithContentCacheDir(filepath.Join(cacheDir, "content")))
	}
	if cache.BlocksEnabled() {
		opts = append(opts, blob.WithBlockCacheDir(filepath.Join(cacheDir, "blocks")))
	}
	if cache.RefsEnabled() {
		opts = append(opts, blob.WithRefCacheDir(filepath.Join(cacheDir, "refs")))
	}
	if cache.ManifestsEnabled() {
		opts = append(opts, blob.WithManifestCacheDir(filepath.Join(cacheDir, "manifests")))
	}
	if cache.IndexesEnabled() {
		opts = append(opts, blob.WithIndexCacheDir(filepath.Join(cacheDir, "indexes")))
	}

	// Only set TTL if refs cache is enabled
	if cache.RefsEnabled() && cache.RefTTL != "" {
		if ttl, err := time.ParseDuration(cache.RefTTL); err == nil {
			opts = append(opts, blob.WithRefCacheTTL(ttl))
		}
	}

	return opts
}

// clientOptsNoCache returns client options without caching.
// Use this when --skip-cache flag is set.
func clientOptsNoCache(cfg *internalcfg.Config) []blob.Option {
	opts := []blob.Option{blob.WithDockerConfig()}
	if cfg.PlainHTTP {
		opts = append(opts, blob.WithPlainHTTP(true))
	}
	return opts
}

// resolveCacheDir returns the cache directory to use.
// Priority: config file > XDG default.
func resolveCacheDir(cfg *internalcfg.Config) (string, error) {
	if cfg.Cache.Dir != "" {
		return cfg.Cache.Dir, nil
	}
	return internalcfg.CacheDir()
}
