package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	internalcfg "github.com/meigma/blob-cli/internal/config"
)

func TestPrepareDestination(t *testing.T) {
	t.Run("existing directory", func(t *testing.T) {
		dir := t.TempDir()
		got, err := prepareDestination(dir)
		require.NoError(t, err)
		assert.Equal(t, dir, got)
	})

	t.Run("nonexistent path creates directory", func(t *testing.T) {
		dir := t.TempDir()
		newDir := filepath.Join(dir, "subdir", "nested")

		got, err := prepareDestination(newDir)
		require.NoError(t, err)
		assert.Equal(t, newDir, got)

		// Verify directory was created
		info, err := os.Stat(newDir)
		require.NoError(t, err)
		assert.True(t, info.IsDir())
	})

	t.Run("path is a file", func(t *testing.T) {
		dir := t.TempDir()
		file := filepath.Join(dir, "file.txt")
		err := os.WriteFile(file, []byte("test"), 0o644)
		require.NoError(t, err)

		_, err = prepareDestination(file)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not a directory")
	})

	t.Run("relative path converted to absolute", func(t *testing.T) {
		// Use current directory which definitely exists
		got, err := prepareDestination(".")
		require.NoError(t, err)
		assert.True(t, filepath.IsAbs(got))
	})
}

func TestPullCmd_NilConfig(t *testing.T) {
	viper.Reset()

	// Don't set config in context
	ctx := context.Background()

	pullCmd.SetContext(ctx)
	err := pullCmd.RunE(pullCmd, []string{"ghcr.io/test:v1"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "configuration not loaded")
}

func TestPullCmd_InvalidReference(t *testing.T) {
	viper.Reset()

	dir := t.TempDir()
	cfg := &internalcfg.Config{}
	ctx := internalcfg.WithConfig(context.Background(), cfg)

	pullCmd.SetContext(ctx)
	err := pullCmd.RunE(pullCmd, []string{"ghcr.io/nonexistent/ref:v1", dir})

	// Should fail during pull (before destination handling)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "pulling archive")
}

func TestPullText(t *testing.T) {
	tests := []struct {
		name       string
		result     *pullResult
		wantOutput string
	}{
		{
			name: "basic pull",
			result: &pullResult{
				Ref:            "ghcr.io/test:v1",
				Destination:    "/tmp/output",
				FileCount:      42,
				TotalSizeHuman: "1.5M",
				Verified:       false,
			},
			wantOutput: "Pulled ghcr.io/test:v1\n  Destination: /tmp/output\n  Files: 42\n  Size: 1.5M\n",
		},
		{
			name: "pull with alias resolution",
			result: &pullResult{
				Ref:            "myalias:v1",
				ResolvedRef:    "ghcr.io/acme/repo:v1",
				Destination:    "/tmp/output",
				FileCount:      10,
				TotalSizeHuman: "512K",
				Verified:       false,
			},
			wantOutput: "Pulled myalias:v1\n  Resolved: ghcr.io/acme/repo:v1\n  Destination: /tmp/output\n  Files: 10\n  Size: 512K\n",
		},
		{
			name: "pull with verification",
			result: &pullResult{
				Ref:            "ghcr.io/test:v1",
				Destination:    "/tmp/output",
				FileCount:      5,
				TotalSizeHuman: "2.3M",
				Verified:       true,
				PoliciesCount:  2,
			},
			wantOutput: "Pulled ghcr.io/test:v1\n  Destination: /tmp/output\n  Files: 5\n  Size: 2.3M\n  Verified: 2 policies applied\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := pullText(tt.result)

			w.Close()
			os.Stdout = oldStdout

			var buf bytes.Buffer
			buf.ReadFrom(r)

			require.NoError(t, err)
			assert.Equal(t, tt.wantOutput, buf.String())
		})
	}
}

func TestPullJSON(t *testing.T) {
	tests := []struct {
		name   string
		result *pullResult
	}{
		{
			name: "basic pull",
			result: &pullResult{
				Ref:            "ghcr.io/test:v1",
				Destination:    "/tmp/output",
				FileCount:      42,
				TotalSize:      1572864,
				TotalSizeHuman: "1.5M",
				Verified:       false,
			},
		},
		{
			name: "pull with alias resolution",
			result: &pullResult{
				Ref:            "myalias:v1",
				ResolvedRef:    "ghcr.io/acme/repo:v1",
				Destination:    "/tmp/output",
				FileCount:      10,
				TotalSize:      524288,
				TotalSizeHuman: "512K",
				Verified:       false,
			},
		},
		{
			name: "pull with verification",
			result: &pullResult{
				Ref:            "ghcr.io/test:v1",
				Destination:    "/tmp/output",
				FileCount:      5,
				TotalSize:      2411724,
				TotalSizeHuman: "2.3M",
				Verified:       true,
				PoliciesCount:  2,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := pullJSON(tt.result)

			w.Close()
			os.Stdout = oldStdout

			var buf bytes.Buffer
			buf.ReadFrom(r)

			require.NoError(t, err)

			// Parse the JSON and verify fields
			var got pullResult
			err = json.Unmarshal(buf.Bytes(), &got)
			require.NoError(t, err)
			assert.Equal(t, tt.result.Ref, got.Ref)
			assert.Equal(t, tt.result.ResolvedRef, got.ResolvedRef)
			assert.Equal(t, tt.result.Destination, got.Destination)
			assert.Equal(t, tt.result.FileCount, got.FileCount)
			assert.Equal(t, tt.result.TotalSize, got.TotalSize)
			assert.Equal(t, tt.result.Verified, got.Verified)
			assert.Equal(t, tt.result.PoliciesCount, got.PoliciesCount)
		})
	}
}

func TestOutputPullResult_Quiet(t *testing.T) {
	viper.Reset()

	cfg := &internalcfg.Config{Quiet: true}
	result := &pullResult{
		Ref:         "ghcr.io/test:v1",
		Destination: "/tmp/output",
		FileCount:   10,
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := outputPullResult(cfg, result)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)

	require.NoError(t, err)
	assert.Empty(t, buf.String(), "quiet mode should produce no output")
}
