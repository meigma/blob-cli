package config

import "github.com/spf13/viper"

// Default output format values.
const (
	OutputText = "text"
	OutputJSON = "json"
)

// Default compression values.
const (
	CompressionNone = "none"
	CompressionZstd = "zstd"
)

// Default returns a new Config with default values.
func Default() *Config {
	return &Config{
		Output:      OutputText,
		Verbose:     0,
		Quiet:       false,
		NoColor:     false,
		PlainHTTP:   false,
		Compression: CompressionZstd,
		Cache: CacheConfig{
			Enabled: true,
			MaxSize: "5GB",
		},
		Aliases:  make(map[string]string),
		Policies: nil,
	}
}

// SetDefaults configures Viper with default values.
// Call this before loading config to ensure defaults are set.
func SetDefaults(v *viper.Viper) {
	v.SetDefault("output", OutputText)
	v.SetDefault("verbose", 0)
	v.SetDefault("quiet", false)
	v.SetDefault("no-color", false)
	v.SetDefault("plain-http", false)
	v.SetDefault("compression", CompressionZstd)
	v.SetDefault("cache.enabled", true)
	v.SetDefault("cache.max_size", "5GB")
}
