package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

const (
	// appName is the application directory name used in XDG paths.
	appName = "blob"

	// configFileName is the config file name (without extension).
	configFileName = "config"

	// configFileExt is the config file extension.
	configFileExt = "yaml"
)

// ConfigDir returns the configuration directory path following XDG Base Directory Specification.
// Uses $XDG_CONFIG_HOME/blob or ~/.config/blob.
//
//nolint:revive // stuttering is acceptable for clarity
func ConfigDir() (string, error) {
	if dir := os.Getenv("XDG_CONFIG_HOME"); dir != "" {
		return filepath.Join(dir, appName), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("determining home directory: %w", err)
	}

	return filepath.Join(home, ".config", appName), nil
}

// ConfigPath returns the full path to the configuration file.
//
//nolint:revive // stuttering is acceptable for clarity
func ConfigPath() (string, error) {
	dir, err := ConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(dir, configFileName+"."+configFileExt), nil
}

// ConfigPathUsed returns the config file path that is actually in use.
// If --config flag was specified, returns that path.
// Otherwise returns the default XDG path.
//
//nolint:revive // stuttering is acceptable for clarity
func ConfigPathUsed() (string, error) {
	// Check if the effective path was set by initConfig()
	// This is stored in an internal key to avoid BLOB_CONFIG env var interference
	if path := viper.GetString("internal.config_path"); path != "" {
		return path, nil
	}

	// Fall back to default XDG path (for cases where initConfig wasn't called)
	return ConfigPath()
}

// CacheDir returns the cache directory path following XDG Base Directory Specification.
// Uses $XDG_CACHE_HOME/blob or ~/.cache/blob.
func CacheDir() (string, error) {
	if dir := os.Getenv("XDG_CACHE_HOME"); dir != "" {
		return filepath.Join(dir, appName), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("determining home directory: %w", err)
	}

	return filepath.Join(home, ".cache", appName), nil
}

// DataDir returns the data directory path following XDG Base Directory Specification.
// Uses $XDG_DATA_HOME/blob or ~/.local/share/blob.
func DataDir() (string, error) {
	if dir := os.Getenv("XDG_DATA_HOME"); dir != "" {
		return filepath.Join(dir, appName), nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("determining home directory: %w", err)
	}

	return filepath.Join(home, ".local", "share", appName), nil
}
