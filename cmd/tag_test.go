package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTagCmd_NilConfig(t *testing.T) {
	// Reset viper and restore after test to avoid affecting other tests.
	viper.Reset()
	t.Cleanup(viper.Reset)

	ctx := context.Background()

	tagCmd.SetContext(ctx)
	err := tagCmd.RunE(tagCmd, []string{"ghcr.io/test:v1", "ghcr.io/test:latest"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "configuration not loaded")
}

func TestTagText_Basic(t *testing.T) {
	result := &tagResult{
		SrcRef: "ghcr.io/acme/configs:v1.0.0",
		DstRef: "ghcr.io/acme/configs:latest",
		Digest: "sha256:abc123def456",
		Status: "success",
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := tagText(result)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)

	require.NoError(t, err)
	got := buf.String()
	assert.Contains(t, got, "Tagged ghcr.io/acme/configs:latest")
	assert.Contains(t, got, "Source: ghcr.io/acme/configs:v1.0.0")
	assert.Contains(t, got, "Digest: sha256:abc123def456")
	assert.NotContains(t, got, "Resolved:")
}

func TestTagText_WithResolvedRefs(t *testing.T) {
	result := &tagResult{
		SrcRef:         "src-alias",
		ResolvedSrcRef: "ghcr.io/acme/configs:v1.0.0",
		DstRef:         "dst-alias",
		ResolvedDstRef: "ghcr.io/acme/configs:latest",
		Digest:         "sha256:abc123",
		Status:         "success",
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := tagText(result)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)

	require.NoError(t, err)
	got := buf.String()
	assert.Contains(t, got, "Tagged dst-alias")
	assert.Contains(t, got, "Resolved: ghcr.io/acme/configs:latest")
	assert.Contains(t, got, "Source: src-alias")
	assert.Contains(t, got, "Resolved: ghcr.io/acme/configs:v1.0.0")
}

func TestTagJSON(t *testing.T) {
	result := &tagResult{
		SrcRef:         "src-alias",
		ResolvedSrcRef: "ghcr.io/acme/configs:v1.0.0",
		DstRef:         "dst-alias",
		ResolvedDstRef: "ghcr.io/acme/configs:latest",
		Digest:         "sha256:abc123def456",
		Status:         "success",
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := tagJSON(result)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)

	require.NoError(t, err)

	var got tagResult
	err = json.Unmarshal(buf.Bytes(), &got)
	require.NoError(t, err)

	assert.Equal(t, "src-alias", got.SrcRef)
	assert.Equal(t, "ghcr.io/acme/configs:v1.0.0", got.ResolvedSrcRef)
	assert.Equal(t, "dst-alias", got.DstRef)
	assert.Equal(t, "ghcr.io/acme/configs:latest", got.ResolvedDstRef)
	assert.Equal(t, "sha256:abc123def456", got.Digest)
	assert.Equal(t, "success", got.Status)
}

func TestTagJSON_OmitsEmpty(t *testing.T) {
	result := &tagResult{
		SrcRef: "ghcr.io/acme/configs:v1.0.0",
		DstRef: "ghcr.io/acme/configs:latest",
		Digest: "sha256:abc123",
		Status: "success",
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := tagJSON(result)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)

	require.NoError(t, err)

	jsonStr := buf.String()
	assert.NotContains(t, jsonStr, "resolved_src_ref")
	assert.NotContains(t, jsonStr, "resolved_dst_ref")
}
