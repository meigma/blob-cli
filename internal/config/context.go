package config

import "context"

// contextKey is a private type for context keys to avoid collisions.
type contextKey struct{}

// WithConfig returns a new context with the config attached.
func WithConfig(ctx context.Context, cfg *Config) context.Context {
	return context.WithValue(ctx, contextKey{}, cfg)
}

// FromContext retrieves the config from context.
// Returns nil if no config is present.
func FromContext(ctx context.Context) *Config {
	cfg, ok := ctx.Value(contextKey{}).(*Config)
	if !ok {
		return nil
	}
	return cfg
}

// MustFromContext retrieves the config from context.
// Panics if no config is present (indicates a programming error).
func MustFromContext(ctx context.Context) *Config {
	cfg := FromContext(ctx)
	if cfg == nil {
		panic("config: no config in context")
	}
	return cfg
}
