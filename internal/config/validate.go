package config

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// ErrInvalidConfig is returned when configuration validation fails.
var ErrInvalidConfig = errors.New("invalid configuration")

// validate checks the config for invalid values.
func validate(cfg *Config) error {
	if err := validateOutput(cfg.Output); err != nil {
		return err
	}
	if err := validateCompression(cfg.Compression); err != nil {
		return err
	}
	if err := validateCache(&cfg.Cache); err != nil {
		return err
	}
	return validatePolicies(cfg.Policies)
}

// validateCache validates cache configuration.
func validateCache(cache *CacheConfig) error {
	if cache.MaxSize != "" {
		if err := validateCacheSize(cache.MaxSize); err != nil {
			return err
		}
	}
	if cache.RefTTL != "" {
		if _, err := time.ParseDuration(cache.RefTTL); err != nil {
			return fmt.Errorf("%w: cache.ref_ttl must be a valid duration (e.g., 5m, 1h), got %q", ErrInvalidConfig, cache.RefTTL)
		}
	}
	return nil
}

func validateOutput(v string) error {
	switch v {
	case OutputText, OutputJSON:
		return nil
	default:
		return fmt.Errorf("%w: output must be %q or %q, got %q", ErrInvalidConfig, OutputText, OutputJSON, v)
	}
}

func validateCompression(v string) error {
	switch v {
	case CompressionNone, CompressionZstd:
		return nil
	default:
		return fmt.Errorf("%w: compression must be %q or %q, got %q", ErrInvalidConfig, CompressionNone, CompressionZstd, v)
	}
}

// validateCacheSize validates a size string like "5GB", "500MB", "1TB".
func validateCacheSize(v string) error {
	if v == "" {
		return nil
	}

	// Parse the numeric portion and unit
	v = strings.TrimSpace(v)
	if v == "" {
		return fmt.Errorf("%w: cache.max_size cannot be empty", ErrInvalidConfig)
	}

	// Find where the number ends
	numEnd := 0
	for i, r := range v {
		if !unicode.IsDigit(r) && r != '.' {
			numEnd = i
			break
		}
		numEnd = i + 1
	}

	if numEnd == 0 {
		return fmt.Errorf("%w: cache.max_size must start with a number, got %q", ErrInvalidConfig, v)
	}

	numStr := v[:numEnd]
	unit := strings.ToUpper(strings.TrimSpace(v[numEnd:]))

	// Validate number
	num, err := strconv.ParseFloat(numStr, 64)
	if err != nil || num < 0 {
		return fmt.Errorf("%w: cache.max_size has invalid number %q", ErrInvalidConfig, numStr)
	}

	// Validate unit
	validUnits := map[string]bool{
		"":   true, // bytes
		"B":  true,
		"KB": true,
		"MB": true,
		"GB": true,
		"TB": true,
	}
	if !validUnits[unit] {
		return fmt.Errorf("%w: cache.max_size has invalid unit %q (valid: B, KB, MB, GB, TB)", ErrInvalidConfig, unit)
	}

	return nil
}

func validatePolicies(policies []PolicyRule) error {
	for i, rule := range policies {
		if rule.Match == "" {
			return fmt.Errorf("%w: policies[%d].match cannot be empty", ErrInvalidConfig, i)
		}

		// Validate that the pattern compiles
		if _, err := regexp.Compile(rule.Match); err != nil {
			return fmt.Errorf("%w: policies[%d].match is invalid regex %q: %v", ErrInvalidConfig, i, rule.Match, err)
		}
	}
	return nil
}
