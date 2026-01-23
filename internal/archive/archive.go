// Package archive provides utilities for inspecting and listing blob archives.
package archive

import (
	"cmp"
	"context"
	"fmt"
	"io/fs"
	"path"
	"slices"
	"strings"
	"time"

	"github.com/meigma/blob"
)

// DirEntry represents a file or synthesized directory for display.
// Archives only store files; directories are synthesized from file paths.
type DirEntry struct {
	Name    string      // Base name (not full path)
	Path    string      // Full path in archive
	IsDir   bool        // True for synthesized directories
	Mode    fs.FileMode // File mode bits
	Size    uint64      // Original (uncompressed) size
	ModTime time.Time   // Modification time
	Hash    []byte      // SHA256 hash (files only)

	// Children holds nested entries for tree building.
	// Only populated by BuildTree.
	Children []*DirEntry
}

// Inspect fetches archive metadata from a registry without downloading file data.
// Additional client options can be passed to customize the client behavior.
func Inspect(ctx context.Context, ref string, opts ...blob.Option) (*blob.InspectResult, error) {
	clientOpts := append([]blob.Option{blob.WithDockerConfig()}, opts...)
	client, err := blob.NewClient(clientOpts...)
	if err != nil {
		return nil, fmt.Errorf("creating client: %w", err)
	}

	result, err := client.Inspect(ctx, ref)
	if err != nil {
		return nil, fmt.Errorf("inspecting archive %s: %w", ref, err)
	}

	return result, nil
}

// ListDir returns the immediate children of a directory path.
// If dirPath is empty or "/", lists the root directory.
// Returns entries sorted alphabetically by name.
func ListDir(index *blob.IndexView, dirPath string) ([]*DirEntry, error) {
	dirPath = normalizePath(dirPath)

	// Build prefix for filtering entries
	var prefix string
	if dirPath == "" {
		prefix = ""
	} else {
		prefix = dirPath + "/"
	}

	seen := make(map[string]*DirEntry)

	for entry := range index.EntriesWithPrefix(prefix) {
		entryPath := entry.Path()

		// Get the relative path after the prefix
		relPath := strings.TrimPrefix(entryPath, prefix)
		if relPath == "" {
			continue
		}

		// Get the first path component (immediate child)
		var name string
		slashIdx := strings.Index(relPath, "/")
		if slashIdx == -1 {
			name = relPath
		} else {
			name = relPath[:slashIdx]
		}

		// Skip if we've already seen this name
		if _, exists := seen[name]; exists {
			continue
		}

		if slashIdx == -1 {
			// This is a file (no more path components)
			hashBytes := entry.HashBytes()
			hash := make([]byte, len(hashBytes))
			copy(hash, hashBytes)

			seen[name] = &DirEntry{
				Name:    name,
				Path:    entryPath,
				IsDir:   false,
				Mode:    entry.Mode(),
				Size:    entry.OriginalSize(),
				ModTime: entry.ModTime(),
				Hash:    hash,
			}
		} else {
			// This is a directory (synthesized)
			childPath := dirPath + "/" + name
			if dirPath == "" {
				childPath = name
			}
			seen[name] = &DirEntry{
				Name:  name,
				Path:  childPath,
				IsDir: true,
				Mode:  fs.ModeDir | 0o755, // Default directory mode
			}
		}
	}

	// Convert to slice and sort
	entries := make([]*DirEntry, 0, len(seen))
	for _, e := range seen {
		entries = append(entries, e)
	}
	slices.SortFunc(entries, func(a, b *DirEntry) int {
		return cmp.Compare(a.Name, b.Name)
	})

	return entries, nil
}

// BuildTree builds a hierarchical tree structure rooted at dirPath.
// If maxDepth is 0, the tree depth is unlimited.
// If maxDepth is > 0, the tree is limited to that many levels.
func BuildTree(index *blob.IndexView, dirPath string, maxDepth int) (*DirEntry, error) {
	dirPath = normalizePath(dirPath)

	// Create the root entry
	rootName := "."
	if dirPath != "" {
		rootName = path.Base(dirPath)
	}
	root := &DirEntry{
		Name:  rootName,
		Path:  dirPath,
		IsDir: true,
		Mode:  fs.ModeDir | 0o755,
	}

	// Build tree recursively
	if err := buildTreeRecursive(index, root, dirPath, 1, maxDepth); err != nil {
		return nil, err
	}

	return root, nil
}

func buildTreeRecursive(index *blob.IndexView, parent *DirEntry, dirPath string, currentDepth, maxDepth int) error {
	// Check depth limit
	if maxDepth > 0 && currentDepth > maxDepth {
		return nil
	}

	entries, err := ListDir(index, dirPath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		parent.Children = append(parent.Children, entry)

		if entry.IsDir {
			if err := buildTreeRecursive(index, entry, entry.Path, currentDepth+1, maxDepth); err != nil {
				return err
			}
		}
	}

	return nil
}

// SortDirsFirst sorts entries with directories first, then files.
// Within each group, entries are sorted alphabetically.
func SortDirsFirst(entries []*DirEntry) {
	slices.SortFunc(entries, func(a, b *DirEntry) int {
		// Directories come before files
		if a.IsDir && !b.IsDir {
			return -1
		}
		if !a.IsDir && b.IsDir {
			return 1
		}
		// Within same type, sort alphabetically
		return cmp.Compare(a.Name, b.Name)
	})
}

// normalizePath normalizes a path for consistent handling.
// - Empty string, ".", "/" all become ""
// - Removes leading "/" and trailing "/"
// - "foo/" becomes "foo"
// - "/foo/bar/" becomes "foo/bar"
// - "foo/./bar" becomes "foo/bar" (via path.Clean)
func normalizePath(p string) string {
	// Handle special cases
	if p == "" || p == "." || p == "/" {
		return ""
	}

	// Clean the path to handle dot segments (./  ../)
	p = path.Clean(p)
	p = strings.TrimPrefix(p, "/")
	p = strings.TrimSuffix(p, "/")

	if p == "." {
		return ""
	}

	return p
}
