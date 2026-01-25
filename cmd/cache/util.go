package cache

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	internalcfg "github.com/meigma/blob-cli/internal/config"
)

// cacheTypeAll is the special cache type name for all caches.
const cacheTypeAll = "all"

// cacheType describes a cache subdirectory.
type cacheType struct {
	Name        string // Display name
	SubDir      string // Subdirectory under cache root
	Description string // Human-readable description
}

// cacheTypes lists all cache types in display order.
var cacheTypes = []cacheType{
	{"content", "content", "File content cache"},
	{"blocks", "blocks", "HTTP range block cache"},
	{"refs", "refs", "Tag to digest mappings"},
	{"manifests", "manifests", "OCI manifest cache"},
	{"indexes", "indexes", "Archive index cache"},
}

// validCacheType returns true if the given type name is valid.
func validCacheType(name string) bool {
	if name == cacheTypeAll {
		return true
	}
	for _, ct := range cacheTypes {
		if ct.Name == name {
			return true
		}
	}
	return false
}

// cacheTypeNames returns a list of valid cache type names.
func cacheTypeNames() []string {
	names := make([]string, len(cacheTypes)+1)
	for i, ct := range cacheTypes {
		names[i] = ct.Name
	}
	names[len(cacheTypes)] = cacheTypeAll
	return names
}

// resolveCacheDir returns the cache directory to use.
// Priority: config file > XDG default.
func resolveCacheDir(cfg *internalcfg.Config) (string, error) {
	if cfg.Cache.Dir != "" {
		return cfg.Cache.Dir, nil
	}
	return internalcfg.CacheDir()
}

// getDirSize calculates the total size of all files in a directory recursively.
// Returns 0 if the directory doesn't exist. Warns to stderr on permission errors.
func getDirSize(dir string) int64 {
	var size int64
	var hadError bool
	//nolint:errcheck // Walk errors are handled by tracking hadError
	filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			hadError = true
			return fs.SkipDir // Skip inaccessible directories
		}
		if !d.IsDir() {
			if info, infoErr := d.Info(); infoErr == nil {
				size += info.Size()
			}
		}
		return nil
	})
	if hadError {
		fmt.Fprintf(os.Stderr, "Warning: some files in %s could not be accessed; size may be incomplete\n", dir)
	}
	return size
}

// countFiles counts all files in a directory recursively.
// Returns 0 if the directory doesn't exist. Warns to stderr on permission errors.
func countFiles(dir string) int {
	var count int
	var hadError bool
	//nolint:errcheck // Walk errors are handled by tracking hadError
	filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			hadError = true
			return fs.SkipDir // Skip inaccessible directories
		}
		if !d.IsDir() {
			count++
		}
		return nil
	})
	if hadError {
		fmt.Fprintf(os.Stderr, "Warning: some files in %s could not be accessed; count may be incomplete\n", dir)
	}
	return count
}

// isCacheTypeEnabled returns whether a cache type is enabled in the config.
func isCacheTypeEnabled(cfg *internalcfg.Config, name string) bool {
	switch name {
	case "content":
		return cfg.Cache.ContentEnabled()
	case "blocks":
		return cfg.Cache.BlocksEnabled()
	case "refs":
		return cfg.Cache.RefsEnabled()
	case "manifests":
		return cfg.Cache.ManifestsEnabled()
	case "indexes":
		return cfg.Cache.IndexesEnabled()
	default:
		return false
	}
}
