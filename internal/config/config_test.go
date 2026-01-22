package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func TestLoad(t *testing.T) {
	v := viper.New()
	SetDefaults(v)

	cfg, err := Load(v)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Check defaults were applied
	if cfg.Output != "text" {
		t.Errorf("Output = %q, want %q", cfg.Output, "text")
	}
	if cfg.Compression != "zstd" {
		t.Errorf("Compression = %q, want %q", cfg.Compression, "zstd")
	}
	if !cfg.Cache.Enabled {
		t.Error("Cache.Enabled = false, want true")
	}
	if cfg.Cache.MaxSize != "5GB" {
		t.Errorf("Cache.MaxSize = %q, want %q", cfg.Cache.MaxSize, "5GB")
	}
}

func TestLoad_WithViperOverrides(t *testing.T) {
	v := viper.New()
	SetDefaults(v)

	// Override some values
	v.Set("output", "json")
	v.Set("compression", "none")

	cfg, err := Load(v)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Output != "json" {
		t.Errorf("Output = %q, want %q", cfg.Output, "json")
	}
	if cfg.Compression != "none" {
		t.Errorf("Compression = %q, want %q", cfg.Compression, "none")
	}
}

func TestLoad_ValidationError(t *testing.T) {
	v := viper.New()
	v.Set("output", "invalid")
	v.Set("compression", "zstd")

	_, err := Load(v)
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}
}

func TestLoad_AliasesInitialized(t *testing.T) {
	v := viper.New()
	SetDefaults(v)

	cfg, err := Load(v)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Aliases map should be initialized even if empty
	if cfg.Aliases == nil {
		t.Error("Aliases map is nil, should be initialized")
	}
}

func TestSave(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "subdir", "config.yaml")

	cfg := &Config{
		Output:      "json",
		Compression: "zstd",
		Cache: CacheConfig{
			Enabled: true,
			MaxSize: "10GB",
		},
		Aliases: map[string]string{
			"foo": "ghcr.io/acme/foo",
		},
	}

	err := Save(cfg, path)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Verify file was created
	if _, statErr := os.Stat(path); os.IsNotExist(statErr) {
		t.Fatal("config file was not created")
	}

	// Verify content can be read back
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading config file: %v", err)
	}

	// Check that it contains expected content
	content := string(data)
	if content == "" {
		t.Error("config file is empty")
	}
}

func TestSaveDefault(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	err := SaveDefault(path)
	if err != nil {
		t.Fatalf("SaveDefault() error = %v", err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("config file was not created")
	}
}

func TestSaveDefaultWithComments(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	err := SaveDefaultWithComments(path)
	if err != nil {
		t.Fatalf("SaveDefaultWithComments() error = %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading config file: %v", err)
	}

	content := string(data)

	// Check for comments
	if content[0] != '#' {
		t.Error("config file should start with a comment")
	}

	// Check for key configuration values
	checks := []string{"output:", "compression:", "cache:", "aliases:", "policies:"}
	for _, check := range checks {
		if !contains(content, check) {
			t.Errorf("config file missing %q", check)
		}
	}
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

func TestDefault(t *testing.T) {
	cfg := Default()

	if cfg.Output != "text" {
		t.Errorf("Output = %q, want %q", cfg.Output, "text")
	}
	if cfg.Compression != "zstd" {
		t.Errorf("Compression = %q, want %q", cfg.Compression, "zstd")
	}
	if !cfg.Cache.Enabled {
		t.Error("Cache.Enabled = false, want true")
	}
	if cfg.Aliases == nil {
		t.Error("Aliases is nil")
	}
}
