package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoad(t *testing.T) {
	v := viper.New()
	SetDefaults(v)

	cfg, err := Load(v)
	require.NoError(t, err)

	// Check defaults were applied
	assert.Equal(t, "text", cfg.Output)
	assert.Equal(t, "zstd", cfg.Compression)
	assert.True(t, cfg.Cache.Enabled)
	assert.Equal(t, "5GB", cfg.Cache.MaxSize)
}

func TestLoad_WithViperOverrides(t *testing.T) {
	v := viper.New()
	SetDefaults(v)

	// Override some values
	v.Set("output", "json")
	v.Set("compression", "none")

	cfg, err := Load(v)
	require.NoError(t, err)

	assert.Equal(t, "json", cfg.Output)
	assert.Equal(t, "none", cfg.Compression)
}

func TestLoad_ValidationError(t *testing.T) {
	v := viper.New()
	v.Set("output", "invalid")
	v.Set("compression", "zstd")

	_, err := Load(v)
	require.Error(t, err)
}

func TestLoad_AliasesInitialized(t *testing.T) {
	v := viper.New()
	SetDefaults(v)

	cfg, err := Load(v)
	require.NoError(t, err)

	// Aliases map should be initialized even if empty
	assert.NotNil(t, cfg.Aliases)
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
	require.NoError(t, err)

	// Verify file was created
	_, err = os.Stat(path)
	require.NoError(t, err, "config file should exist")

	// Verify content can be read back
	data, err := os.ReadFile(path)
	require.NoError(t, err)

	// Check that it contains expected content
	assert.NotEmpty(t, string(data))
}

func TestSaveDefault(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	err := SaveDefault(path)
	require.NoError(t, err)

	_, err = os.Stat(path)
	require.NoError(t, err, "config file should exist")
}

func TestSaveDefaultWithComments(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")

	err := SaveDefaultWithComments(path)
	require.NoError(t, err)

	data, err := os.ReadFile(path)
	require.NoError(t, err)

	content := string(data)

	// Check for comments
	assert.Equal(t, '#', rune(content[0]), "config file should start with a comment")

	// Check for key configuration values
	checks := []string{"output:", "compression:", "cache:", "aliases:", "policies:"}
	for _, check := range checks {
		assert.Contains(t, content, check)
	}
}

func TestDefault(t *testing.T) {
	cfg := Default()

	assert.Equal(t, "text", cfg.Output)
	assert.Equal(t, "zstd", cfg.Compression)
	assert.True(t, cfg.Cache.Enabled)
	assert.NotNil(t, cfg.Aliases)
}
