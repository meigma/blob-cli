package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
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
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := "/custom/config/blob"
		if dir != want {
			t.Errorf("ConfigDir() = %q, want %q", dir, want)
		}
	})

	t.Run("without XDG_CONFIG_HOME", func(t *testing.T) {
		os.Unsetenv("XDG_CONFIG_HOME")
		dir, err := ConfigDir()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		home, _ := os.UserHomeDir()
		want := filepath.Join(home, ".config", "blob")
		if dir != want {
			t.Errorf("ConfigDir() = %q, want %q", dir, want)
		}
	})
}

func TestConfigPath(t *testing.T) {
	origXDG := os.Getenv("XDG_CONFIG_HOME")
	t.Cleanup(func() {
		os.Setenv("XDG_CONFIG_HOME", origXDG)
	})

	os.Setenv("XDG_CONFIG_HOME", "/custom/config")
	path, err := ConfigPath()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "/custom/config/blob/config.yaml"
	if path != want {
		t.Errorf("ConfigPath() = %q, want %q", path, want)
	}
}

func TestCacheDir(t *testing.T) {
	origXDG := os.Getenv("XDG_CACHE_HOME")
	t.Cleanup(func() {
		os.Setenv("XDG_CACHE_HOME", origXDG)
	})

	t.Run("with XDG_CACHE_HOME set", func(t *testing.T) {
		os.Setenv("XDG_CACHE_HOME", "/custom/cache")
		dir, err := CacheDir()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := "/custom/cache/blob"
		if dir != want {
			t.Errorf("CacheDir() = %q, want %q", dir, want)
		}
	})

	t.Run("without XDG_CACHE_HOME", func(t *testing.T) {
		os.Unsetenv("XDG_CACHE_HOME")
		dir, err := CacheDir()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		home, _ := os.UserHomeDir()
		want := filepath.Join(home, ".cache", "blob")
		if dir != want {
			t.Errorf("CacheDir() = %q, want %q", dir, want)
		}
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
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := "/custom/data/blob"
		if dir != want {
			t.Errorf("DataDir() = %q, want %q", dir, want)
		}
	})

	t.Run("without XDG_DATA_HOME", func(t *testing.T) {
		os.Unsetenv("XDG_DATA_HOME")
		dir, err := DataDir()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		home, _ := os.UserHomeDir()
		want := filepath.Join(home, ".local", "share", "blob")
		if dir != want {
			t.Errorf("DataDir() = %q, want %q", dir, want)
		}
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
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := "/custom/path/config.yaml"
		if path != want {
			t.Errorf("ConfigPathUsed() = %q, want %q", path, want)
		}
	})

	t.Run("without internal config path (fallback to XDG)", func(t *testing.T) {
		viper.Reset()
		os.Setenv("XDG_CONFIG_HOME", "/xdg/config")
		path, err := ConfigPathUsed()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := "/xdg/config/blob/config.yaml"
		if path != want {
			t.Errorf("ConfigPathUsed() = %q, want %q", path, want)
		}
	})
}
