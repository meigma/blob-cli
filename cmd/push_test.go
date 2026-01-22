package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/meigma/blob"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	internalcfg "github.com/meigma/blob-cli/internal/config"
)

func TestParseAnnotations(t *testing.T) {
	tests := []struct {
		name    string
		input   []string
		want    map[string]string
		wantErr string
	}{
		{
			name:  "nil input",
			input: nil,
			want:  map[string]string{},
		},
		{
			name:  "empty input",
			input: []string{},
			want:  map[string]string{},
		},
		{
			name:  "single valid annotation",
			input: []string{"key=value"},
			want:  map[string]string{"key": "value"},
		},
		{
			name:  "multiple valid annotations",
			input: []string{"a=1", "b=2", "c=3"},
			want:  map[string]string{"a": "1", "b": "2", "c": "3"},
		},
		{
			name:  "empty value",
			input: []string{"key="},
			want:  map[string]string{"key": ""},
		},
		{
			name:  "value contains equals",
			input: []string{"key=val=ue"},
			want:  map[string]string{"key": "val=ue"},
		},
		{
			name:  "value with multiple equals",
			input: []string{"key=a=b=c"},
			want:  map[string]string{"key": "a=b=c"},
		},
		{
			name:    "no equals sign",
			input:   []string{"keyvalue"},
			wantErr: "invalid annotation",
		},
		{
			name:    "empty key",
			input:   []string{"=value"},
			wantErr: "invalid annotation",
		},
		{
			name:    "only equals",
			input:   []string{"="},
			wantErr: "invalid annotation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseAnnotations(tt.input)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestMapCompression(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    blob.Compression
		wantErr string
	}{
		{
			name:  "zstd",
			input: "zstd",
			want:  blob.CompressionZstd,
		},
		{
			name:  "none",
			input: "none",
			want:  blob.CompressionNone,
		},
		{
			name:  "empty defaults to zstd",
			input: "",
			want:  blob.CompressionZstd,
		},
		{
			name:    "invalid compression",
			input:   "gzip",
			wantErr: "invalid compression type",
		},
		{
			name:    "invalid uppercase",
			input:   "ZSTD",
			wantErr: "invalid compression type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := mapCompression(tt.input)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestValidateSourcePath(t *testing.T) {
	t.Run("valid directory", func(t *testing.T) {
		dir := t.TempDir()
		err := validateSourcePath(dir)
		require.NoError(t, err)
	})

	t.Run("nonexistent path", func(t *testing.T) {
		err := validateSourcePath("/nonexistent/path/that/does/not/exist")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "does not exist")
	})

	t.Run("path is a file", func(t *testing.T) {
		dir := t.TempDir()
		file := filepath.Join(dir, "file.txt")
		err := os.WriteFile(file, []byte("test"), 0o644)
		require.NoError(t, err)

		err = validateSourcePath(file)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "not a directory")
	})
}

func TestPushCmd_NilConfig(t *testing.T) {
	viper.Reset()

	// Create a temp directory for source path
	dir := t.TempDir()

	// Don't set config in context
	ctx := context.Background()

	pushCmd.SetContext(ctx)
	err := pushCmd.RunE(pushCmd, []string{"ghcr.io/test:v1", dir})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "configuration not loaded")
}

func TestPushCmd_InvalidSourcePath(t *testing.T) {
	viper.Reset()

	cfg := &internalcfg.Config{}
	ctx := internalcfg.WithConfig(context.Background(), cfg)

	pushCmd.SetContext(ctx)
	err := pushCmd.RunE(pushCmd, []string{"ghcr.io/test:v1", "/nonexistent/path"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "does not exist")
}

func TestPushCmd_SourcePathIsFile(t *testing.T) {
	viper.Reset()

	// Create a temp file
	dir := t.TempDir()
	file := filepath.Join(dir, "file.txt")
	err := os.WriteFile(file, []byte("test"), 0o644)
	require.NoError(t, err)

	cfg := &internalcfg.Config{}
	ctx := internalcfg.WithConfig(context.Background(), cfg)

	pushCmd.SetContext(ctx)
	err = pushCmd.RunE(pushCmd, []string{"ghcr.io/test:v1", file})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "not a directory")
}

func TestPushText(t *testing.T) {
	tests := []struct {
		name       string
		result     pushResult
		wantOutput string
	}{
		{
			name: "basic push",
			result: pushResult{
				Ref:    "ghcr.io/test:v1",
				Status: "success",
			},
			wantOutput: "Pushed ghcr.io/test:v1\n",
		},
		{
			name: "push with signing",
			result: pushResult{
				Ref:             "ghcr.io/test:v1",
				Status:          "success",
				Signed:          true,
				SignatureDigest: "sha256:abc123",
			},
			wantOutput: "Pushed ghcr.io/test:v1\nSigned: sha256:abc123\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := pushText(tt.result)

			w.Close()
			os.Stdout = oldStdout

			var buf bytes.Buffer
			buf.ReadFrom(r)

			require.NoError(t, err)
			assert.Equal(t, tt.wantOutput, buf.String())
		})
	}
}

func TestPushJSON(t *testing.T) {
	tests := []struct {
		name   string
		result pushResult
	}{
		{
			name: "basic push",
			result: pushResult{
				Ref:    "ghcr.io/test:v1",
				Status: "success",
			},
		},
		{
			name: "push with signing",
			result: pushResult{
				Ref:             "ghcr.io/test:v1",
				Status:          "success",
				Signed:          true,
				SignatureDigest: "sha256:abc123",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := pushJSON(tt.result)

			w.Close()
			os.Stdout = oldStdout

			var buf bytes.Buffer
			buf.ReadFrom(r)

			require.NoError(t, err)

			// Parse the JSON and verify fields
			var got pushResult
			err = json.Unmarshal(buf.Bytes(), &got)
			require.NoError(t, err)
			assert.Equal(t, tt.result.Ref, got.Ref)
			assert.Equal(t, tt.result.Status, got.Status)
			assert.Equal(t, tt.result.Signed, got.Signed)
			assert.Equal(t, tt.result.SignatureDigest, got.SignatureDigest)
		})
	}
}
