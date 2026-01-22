package config

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWithConfig_FromContext(t *testing.T) {
	cfg := &Config{Output: "json"}
	ctx := context.Background()

	// Add config to context
	ctx = WithConfig(ctx, cfg)

	// Retrieve it
	got := FromContext(ctx)
	require.NotNil(t, got)
	assert.Equal(t, "json", got.Output)
}

func TestFromContext_NoConfig(t *testing.T) {
	ctx := context.Background()
	got := FromContext(ctx)
	assert.Nil(t, got)
}

func TestMustFromContext_Success(t *testing.T) {
	cfg := &Config{Output: "text"}
	ctx := WithConfig(context.Background(), cfg)

	got := MustFromContext(ctx)
	require.NotNil(t, got)
	assert.Equal(t, "text", got.Output)
}

func TestMustFromContext_Panics(t *testing.T) {
	ctx := context.Background()
	assert.Panics(t, func() {
		MustFromContext(ctx)
	})
}
