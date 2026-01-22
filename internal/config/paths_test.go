package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigDir(t *testing.T) {
	// Save and restore env
	origXDG := os.Getenv("XDG_CONFIG_HOME")
	t.Cleanup(func() {
		os.Setenv("XDG_CONFIG_HOME", origXDG)
	})

	t.Run("with XDG_CONFIG_HOME set", func(t *testing.T) {
		os.Setenv("XDG_CONFIG_HOME", "/custom/config")
		dir, err := ConfigDir()
		require.NoError(t, err)
		assert.Equal(t, "/custom/config/blob", dir)
	})

	t.Run("without XDG_CONFIG_HOME", func(t *testing.T) {
		os.Unsetenv("XDG_CONFIG_HOME")
		dir, err := ConfigDir()
		require.NoError(t, err)
		home, _ := os.UserHomeDir()
		want := filepath.Join(home, ".config", "blob")
		assert.Equal(t, want, dir)
	})
}

func TestConfigPath(t *testing.T) {
	origXDG := os.Getenv("XDG_CONFIG_HOME")
	t.Cleanup(func() {
		os.Setenv("XDG_CONFIG_HOME", origXDG)
	})

	os.Setenv("XDG_CONFIG_HOME", "/custom/config")
	path, err := ConfigPath()
	require.NoError(t, err)
	assert.Equal(t, "/custom/config/blob/config.yaml", path)
}

func TestCacheDir(t *testing.T) {
	origXDG := os.Getenv("XDG_CACHE_HOME")
	t.Cleanup(func() {
		os.Setenv("XDG_CACHE_HOME", origXDG)
	})

	t.Run("with XDG_CACHE_HOME set", func(t *testing.T) {
		os.Setenv("XDG_CACHE_HOME", "/custom/cache")
		dir, err := CacheDir()
		require.NoError(t, err)
		assert.Equal(t, "/custom/cache/blob", dir)
	})

	t.Run("without XDG_CACHE_HOME", func(t *testing.T) {
		os.Unsetenv("XDG_CACHE_HOME")
		dir, err := CacheDir()
		require.NoError(t, err)
		home, _ := os.UserHomeDir()
		want := filepath.Join(home, ".cache", "blob")
		assert.Equal(t, want, dir)
	})
}

func TestDataDir(t *testing.T) {
	origXDG := os.Getenv("XDG_DATA_HOME")
	t.Cleanup(func() {
		os.Setenv("XDG_DATA_HOME", origXDG)
	})

	t.Run("with XDG_DATA_HOME set", func(t *testing.T) {
		os.Setenv("XDG_DATA_HOME", "/custom/data")
		dir, err := DataDir()
		require.NoError(t, err)
		assert.Equal(t, "/custom/data/blob", dir)
	})

	t.Run("without XDG_DATA_HOME", func(t *testing.T) {
		os.Unsetenv("XDG_DATA_HOME")
		dir, err := DataDir()
		require.NoError(t, err)
		home, _ := os.UserHomeDir()
		want := filepath.Join(home, ".local", "share", "blob")
		assert.Equal(t, want, dir)
	})
}

func TestConfigPathUsed(t *testing.T) {
	origXDG := os.Getenv("XDG_CONFIG_HOME")
	t.Cleanup(func() {
		os.Setenv("XDG_CONFIG_HOME", origXDG)
		viper.Reset()
	})

	t.Run("with internal config path set", func(t *testing.T) {
		viper.Reset()
		viper.Set("internal.config_path", "/custom/path/config.yaml")
		path, err := ConfigPathUsed()
		require.NoError(t, err)
		assert.Equal(t, "/custom/path/config.yaml", path)
	})

	t.Run("without internal config path (fallback to XDG)", func(t *testing.T) {
		viper.Reset()
		os.Setenv("XDG_CONFIG_HOME", "/xdg/config")
		path, err := ConfigPathUsed()
		require.NoError(t, err)
		assert.Equal(t, "/xdg/config/blob/config.yaml", path)
	})
}
