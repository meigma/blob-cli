package config

import (
	"context"
	"testing"
)

func TestWithConfig_FromContext(t *testing.T) {
	cfg := &Config{Output: "json"}
	ctx := context.Background()

	// Add config to context
	ctx = WithConfig(ctx, cfg)

	// Retrieve it
	got := FromContext(ctx)
	if got == nil {
		t.Fatal("FromContext returned nil")
	}
	if got.Output != "json" {
		t.Errorf("Output = %q, want %q", got.Output, "json")
	}
}

func TestFromContext_NoConfig(t *testing.T) {
	ctx := context.Background()
	got := FromContext(ctx)
	if got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

func TestMustFromContext_Success(t *testing.T) {
	cfg := &Config{Output: "text"}
	ctx := WithConfig(context.Background(), cfg)

	got := MustFromContext(ctx)
	if got == nil {
		t.Fatal("MustFromContext returned nil")
	}
	if got.Output != "text" {
		t.Errorf("Output = %q, want %q", got.Output, "text")
	}
}

func TestMustFromContext_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic, got none")
		}
	}()

	ctx := context.Background()
	MustFromContext(ctx) // should panic
}
