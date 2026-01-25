package cache

import (
	"os"
	"path/filepath"
	"testing"
)

func TestParseClearArgs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		args        []string
		wantType    string
		wantCount   int
		wantErr     bool
		errContains string
	}{
		{
			name:      "no args defaults to all",
			args:      []string{},
			wantType:  cacheTypeAll,
			wantCount: len(cacheTypes),
		},
		{
			name:      "explicit all",
			args:      []string{"all"},
			wantType:  cacheTypeAll,
			wantCount: len(cacheTypes),
		},
		{
			name:      "content type",
			args:      []string{"content"},
			wantType:  "content",
			wantCount: 1,
		},
		{
			name:      "blocks type",
			args:      []string{"blocks"},
			wantType:  "blocks",
			wantCount: 1,
		},
		{
			name:        "invalid type",
			args:        []string{"invalid"},
			wantErr:     true,
			errContains: "invalid cache type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gotType, gotTypes, err := parseClearArgs(tt.args)

			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("error %q should contain %q", err.Error(), tt.errContains)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if gotType != tt.wantType {
				t.Errorf("parseClearArgs() type = %q, want %q", gotType, tt.wantType)
			}

			if len(gotTypes) != tt.wantCount {
				t.Errorf("parseClearArgs() types count = %d, want %d", len(gotTypes), tt.wantCount)
			}
		})
	}
}

func TestClearDirectory(t *testing.T) {
	t.Parallel()

	t.Run("clears files but keeps directory", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()

		// Create test files
		for i := range 3 {
			if err := os.WriteFile(filepath.Join(dir, "file"+string(rune('a'+i))), []byte("content"), 0o644); err != nil {
				t.Fatal(err)
			}
		}

		// Clear the directory
		if err := clearDirectory(dir); err != nil {
			t.Fatalf("clearDirectory() error: %v", err)
		}

		// Directory should still exist
		info, err := os.Stat(dir)
		if err != nil {
			t.Fatalf("directory should exist: %v", err)
		}
		if !info.IsDir() {
			t.Error("should still be a directory")
		}

		// Should be empty
		entries, err := os.ReadDir(dir)
		if err != nil {
			t.Fatalf("ReadDir() error: %v", err)
		}
		if len(entries) != 0 {
			t.Errorf("directory should be empty, has %d entries", len(entries))
		}
	})

	t.Run("clears nested directories", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()

		// Create nested structure
		subdir := filepath.Join(dir, "subdir", "nested")
		if err := os.MkdirAll(subdir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(subdir, "file.txt"), []byte("content"), 0o644); err != nil {
			t.Fatal(err)
		}

		// Clear the directory
		if err := clearDirectory(dir); err != nil {
			t.Fatalf("clearDirectory() error: %v", err)
		}

		// Should be empty
		entries, err := os.ReadDir(dir)
		if err != nil {
			t.Fatalf("ReadDir() error: %v", err)
		}
		if len(entries) != 0 {
			t.Errorf("directory should be empty, has %d entries", len(entries))
		}
	})

	t.Run("nonexistent directory returns nil", func(t *testing.T) {
		t.Parallel()
		err := clearDirectory("/nonexistent/path/that/does/not/exist")
		if err != nil {
			t.Errorf("clearDirectory(nonexistent) should return nil, got: %v", err)
		}
	})

	t.Run("empty directory returns nil", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()

		err := clearDirectory(dir)
		if err != nil {
			t.Errorf("clearDirectory(empty) should return nil, got: %v", err)
		}
	})
}

func TestCalculateCacheSizes(t *testing.T) {
	t.Parallel()

	t.Run("empty cache directories", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()

		types := []cacheType{{Name: "test", SubDir: "test"}}
		size, files := calculateCacheSizes(dir, types)

		if size != 0 {
			t.Errorf("size = %d, want 0", size)
		}
		if files != 0 {
			t.Errorf("files = %d, want 0", files)
		}
	})

	t.Run("cache with files", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()

		// Create test subdir with files
		testDir := filepath.Join(dir, "test")
		if err := os.MkdirAll(testDir, 0o755); err != nil {
			t.Fatal(err)
		}
		content := []byte("test content")
		if err := os.WriteFile(filepath.Join(testDir, "file.txt"), content, 0o644); err != nil {
			t.Fatal(err)
		}

		types := []cacheType{{Name: "test", SubDir: "test"}}
		size, files := calculateCacheSizes(dir, types)

		if size != int64(len(content)) {
			t.Errorf("size = %d, want %d", size, len(content))
		}
		if files != 1 {
			t.Errorf("files = %d, want 1", files)
		}
	})
}

// contains checks if s contains substr (simple helper).
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || s != "" && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := range len(s) - len(substr) + 1 {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
