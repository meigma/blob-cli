package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"testing"

	"github.com/meigma/blob"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInspectCmd_NilConfig(t *testing.T) {
	// Reset viper and restore after test to avoid affecting other tests.
	viper.Reset()
	t.Cleanup(viper.Reset)

	ctx := context.Background()

	inspectCmd.SetContext(ctx)
	err := inspectCmd.RunE(inspectCmd, []string{"ghcr.io/test:v1"})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "configuration not loaded")
}

func TestConvertReferrers_Nil(t *testing.T) {
	got := convertReferrers(nil)
	assert.Nil(t, got)
}

func TestConvertReferrers_Empty(t *testing.T) {
	got := convertReferrers([]blob.Referrer{})
	assert.Nil(t, got)
}

func TestConvertReferrers_Populated(t *testing.T) {
	refs := []blob.Referrer{
		{
			Digest:       "sha256:abc123",
			ArtifactType: "application/vnd.dev.sigstore.bundle.v0.3+json",
			Annotations:  map[string]string{"key": "value"},
		},
		{
			Digest:       "sha256:def456",
			ArtifactType: "application/vnd.in-toto+json",
		},
	}

	got := convertReferrers(refs)

	require.Len(t, got, 2)
	assert.Equal(t, "sha256:abc123", got[0].Digest)
	assert.Equal(t, "application/vnd.dev.sigstore.bundle.v0.3+json", got[0].ArtifactType)
	assert.Equal(t, map[string]string{"key": "value"}, got[0].Annotations)
	assert.Equal(t, "sha256:def456", got[1].Digest)
	assert.Nil(t, got[1].Annotations)
}

func TestWarnReferrerError_Nil(t *testing.T) {
	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	warnReferrerError(nil, "signatures")

	w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer
	buf.ReadFrom(r)

	assert.Empty(t, buf.String())
}

func TestWarnReferrerError_Unsupported(t *testing.T) {
	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	warnReferrerError(blob.ErrReferrersUnsupported, "signatures")

	w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer
	buf.ReadFrom(r)

	// ErrReferrersUnsupported should be silently ignored
	assert.Empty(t, buf.String())
}

func TestWarnReferrerError_OtherError(t *testing.T) {
	// Capture stderr
	oldStderr := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	testErr := errors.New("authentication failed")
	warnReferrerError(testErr, "signatures")

	w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer
	buf.ReadFrom(r)

	// Other errors should produce a warning
	got := buf.String()
	assert.Contains(t, got, "Warning:")
	assert.Contains(t, got, "signatures")
	assert.Contains(t, got, "authentication failed")
}

func TestInspectText_Basic(t *testing.T) {
	output := &inspectOutput{
		Ref:         "ghcr.io/test:v1",
		Digest:      "sha256:abc123def456",
		Files:       42,
		Compression: "zstd",
		Size: sizeInfo{
			Compressed:   1024,
			Uncompressed: 2048,
			Ratio:        0.5,
		},
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := inspectText(output)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)

	require.NoError(t, err)
	got := buf.String()
	assert.Contains(t, got, "Reference:    ghcr.io/test:v1")
	assert.Contains(t, got, "Digest:       sha256:abc123def456")
	assert.Contains(t, got, "Files:        42")
	assert.Contains(t, got, "Compression:  zstd")
	assert.Contains(t, got, "1.0K")
	assert.Contains(t, got, "2.0K uncompressed")
	assert.NotContains(t, got, "Resolved:")
	assert.NotContains(t, got, "Signatures:")
	assert.NotContains(t, got, "Attestations:")
	assert.NotContains(t, got, "Annotations:")
}

func TestInspectText_WithResolvedRef(t *testing.T) {
	output := &inspectOutput{
		Ref:         "test",
		ResolvedRef: "ghcr.io/acme/test:latest",
		Digest:      "sha256:abc123",
		Files:       1,
		Compression: "none",
		Size:        sizeInfo{Compressed: 100, Uncompressed: 100},
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := inspectText(output)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)

	require.NoError(t, err)
	got := buf.String()
	assert.Contains(t, got, "Reference:    test")
	assert.Contains(t, got, "Resolved:     ghcr.io/acme/test:latest")
}

func TestInspectText_WithSections(t *testing.T) {
	output := &inspectOutput{
		Ref:         "ghcr.io/test:v1",
		Digest:      "sha256:abc123",
		Files:       1,
		Compression: "zstd",
		Created:     "2024-01-15T10:30:00Z",
		Size:        sizeInfo{Compressed: 100, Uncompressed: 200},
		Signatures: []referrerInfo{
			{Digest: "sha256:sig1"},
		},
		Attestations: []referrerInfo{
			{Digest: "sha256:att1"},
		},
		Annotations: map[string]string{
			"org.example.key": "value",
		},
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := inspectText(output)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)

	require.NoError(t, err)
	got := buf.String()
	assert.Contains(t, got, "Created:      2024-01-15T10:30:00Z")
	assert.Contains(t, got, "Signatures:")
	assert.Contains(t, got, "sha256:sig1")
	assert.Contains(t, got, "Attestations:")
	assert.Contains(t, got, "sha256:att1")
	assert.Contains(t, got, "Annotations:")
	assert.Contains(t, got, "org.example.key: value")
}

func TestInspectJSON(t *testing.T) {
	output := &inspectOutput{
		Ref:         "ghcr.io/test:v1",
		ResolvedRef: "ghcr.io/resolved:v1",
		Digest:      "sha256:abc123",
		Created:     "2024-01-15T10:30:00Z",
		Files:       42,
		Compression: "zstd",
		Size: sizeInfo{
			Compressed:   1024,
			Uncompressed: 2048,
			Ratio:        0.5,
		},
		Signatures: []referrerInfo{
			{Digest: "sha256:sig1", ArtifactType: "sigstore"},
		},
		Annotations: map[string]string{"key": "value"},
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := inspectJSON(output)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)

	require.NoError(t, err)

	var got inspectOutput
	err = json.Unmarshal(buf.Bytes(), &got)
	require.NoError(t, err)

	assert.Equal(t, "ghcr.io/test:v1", got.Ref)
	assert.Equal(t, "ghcr.io/resolved:v1", got.ResolvedRef)
	assert.Equal(t, "sha256:abc123", got.Digest)
	assert.Equal(t, "2024-01-15T10:30:00Z", got.Created)
	assert.Equal(t, 42, got.Files)
	assert.Equal(t, "zstd", got.Compression)
	assert.Equal(t, uint64(1024), got.Size.Compressed)
	assert.Equal(t, uint64(2048), got.Size.Uncompressed)
	assert.Equal(t, 0.5, got.Size.Ratio)
	require.Len(t, got.Signatures, 1)
	assert.Equal(t, "sha256:sig1", got.Signatures[0].Digest)
	assert.Equal(t, map[string]string{"key": "value"}, got.Annotations)
}

func TestInspectJSON_OmitsEmpty(t *testing.T) {
	output := &inspectOutput{
		Ref:         "ghcr.io/test:v1",
		Digest:      "sha256:abc123",
		Files:       1,
		Compression: "none",
		Size:        sizeInfo{Compressed: 100, Uncompressed: 100},
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	err := inspectJSON(output)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	buf.ReadFrom(r)

	require.NoError(t, err)

	jsonStr := buf.String()
	assert.NotContains(t, jsonStr, "resolved_ref")
	assert.NotContains(t, jsonStr, "created")
	assert.NotContains(t, jsonStr, "signatures")
	assert.NotContains(t, jsonStr, "attestations")
	assert.NotContains(t, jsonStr, "annotations")
}
