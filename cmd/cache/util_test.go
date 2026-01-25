package cache

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidCacheType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		typeName string
		want     bool
	}{
		{"all", "all", true},
		{"content", "content", true},
		{"blocks", "blocks", true},
		{"refs", "refs", true},
		{"manifests", "manifests", true},
		{"indexes", "indexes", true},
		{"invalid", "invalid", false},
		{"empty", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := validCacheType(tt.typeName)
			if got != tt.want {
				t.Errorf("validCacheType(%q) = %v, want %v", tt.typeName, got, tt.want)
			}
		})
	}
}

func TestCacheTypeNames(t *testing.T) {
	t.Parallel()

	names := cacheTypeNames()

	// Should contain all cache types plus "all"
	expectedCount := len(cacheTypes) + 1
	if len(names) != expectedCount {
		t.Errorf("cacheTypeNames() returned %d names, want %d", len(names), expectedCount)
	}

	// Last element should be "all"
	if names[len(names)-1] != cacheTypeAll {
		t.Errorf("cacheTypeNames() last element = %q, want %q", names[len(names)-1], cacheTypeAll)
	}

	// All individual cache types should be present
	for _, ct := range cacheTypes {
		found := false
		for _, name := range names {
			if name == ct.Name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("cacheTypeNames() missing cache type %q", ct.Name)
		}
	}
}

func TestGetDirSize(t *testing.T) {
	t.Parallel()

	t.Run("empty directory", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		size := getDirSize(dir)
		if size != 0 {
			t.Errorf("getDirSize(empty dir) = %d, want 0", size)
		}
	})

	t.Run("directory with files", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()

		// Create test files
		content := []byte("test content")
		if err := os.WriteFile(filepath.Join(dir, "file1.txt"), content, 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "file2.txt"), content, 0o644); err != nil {
			t.Fatal(err)
		}

		size := getDirSize(dir)
		expectedSize := int64(len(content) * 2)
		if size != expectedSize {
			t.Errorf("getDirSize() = %d, want %d", size, expectedSize)
		}
	})

	t.Run("nested directories", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()

		subdir := filepath.Join(dir, "subdir")
		if err := os.MkdirAll(subdir, 0o755); err != nil {
			t.Fatal(err)
		}

		content := []byte("nested file")
		if err := os.WriteFile(filepath.Join(subdir, "nested.txt"), content, 0o644); err != nil {
			t.Fatal(err)
		}

		size := getDirSize(dir)
		expectedSize := int64(len(content))
		if size != expectedSize {
			t.Errorf("getDirSize() = %d, want %d", size, expectedSize)
		}
	})

	t.Run("nonexistent directory", func(t *testing.T) {
		t.Parallel()
		size := getDirSize("/nonexistent/path/that/does/not/exist")
		if size != 0 {
			t.Errorf("getDirSize(nonexistent) = %d, want 0", size)
		}
	})
}

func TestCountFiles(t *testing.T) {
	t.Parallel()

	t.Run("empty directory", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		count := countFiles(dir)
		if count != 0 {
			t.Errorf("countFiles(empty dir) = %d, want 0", count)
		}
	})

	t.Run("directory with files", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()

		// Create test files
		for i := range 3 {
			if err := os.WriteFile(filepath.Join(dir, "file"+string(rune('a'+i))), []byte("content"), 0o644); err != nil {
				t.Fatal(err)
			}
		}

		count := countFiles(dir)
		if count != 3 {
			t.Errorf("countFiles() = %d, want 3", count)
		}
	})

	t.Run("nested directories", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()

		// Create nested structure with files
		subdir := filepath.Join(dir, "subdir")
		if err := os.MkdirAll(subdir, 0o755); err != nil {
			t.Fatal(err)
		}

		if err := os.WriteFile(filepath.Join(dir, "root.txt"), []byte("content"), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(subdir, "nested.txt"), []byte("content"), 0o644); err != nil {
			t.Fatal(err)
		}

		count := countFiles(dir)
		if count != 2 {
			t.Errorf("countFiles() = %d, want 2", count)
		}
	})

	t.Run("nonexistent directory", func(t *testing.T) {
		t.Parallel()
		count := countFiles("/nonexistent/path/that/does/not/exist")
		if count != 0 {
			t.Errorf("countFiles(nonexistent) = %d, want 0", count)
		}
	})
}
