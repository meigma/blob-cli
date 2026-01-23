package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/meigma/blob/policy/sigstore"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractReference(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "tag reference",
			input: "ghcr.io/acme/configs:v1.0.0",
			want:  "v1.0.0",
		},
		{
			name:  "digest reference",
			input: "ghcr.io/acme/configs@sha256:abc123",
			want:  "sha256:abc123",
		},
		{
			name:  "tag with port",
			input: "localhost:5000/repo:latest",
			want:  "latest",
		},
		{
			name:  "no tag or digest",
			input: "ghcr.io/acme/configs",
			want:  "",
		},
		{
			name:  "nested path with tag",
			input: "ghcr.io/acme/team/configs:v2",
			want:  "v2",
		},
		{
			name:  "digest takes precedence over tag-like chars",
			input: "ghcr.io/acme/configs:v1@sha256:abc123",
			want:  "sha256:abc123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractReference(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestSignText(t *testing.T) {
	tests := []struct {
		name       string
		result     signResult
		wantOutput string
	}{
		{
			name: "basic sign",
			result: signResult{
				Ref:             "ghcr.io/test:v1",
				SignatureDigest: "sha256:abc123",
				Status:          "success",
			},
			wantOutput: "Signed ghcr.io/test:v1\nSignature: sha256:abc123\n",
		},
		{
			name: "sign with resolved ref",
			result: signResult{
				Ref:             "myalias:v1",
				ResolvedRef:     "ghcr.io/acme/configs:v1",
				SignatureDigest: "sha256:def456",
				Status:          "success",
			},
			wantOutput: "Signed myalias:v1\n  Resolved: ghcr.io/acme/configs:v1\nSignature: sha256:def456\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := signText(&tt.result)

			w.Close()
			os.Stdout = oldStdout

			var buf bytes.Buffer
			buf.ReadFrom(r)

			require.NoError(t, err)
			assert.Equal(t, tt.wantOutput, buf.String())
		})
	}
}

func TestSignToStdout_InvalidReference(t *testing.T) {
	// signToStdout should return a clear error when reference has no tag or digest
	ctx := context.Background()

	// Create a mock signer (won't be used since we fail early)
	signer, err := sigstore.NewSigner(sigstore.WithEphemeralKey())
	require.NoError(t, err)

	err = signToStdout(ctx, "ghcr.io/acme/configs", signer)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid reference")
	assert.Contains(t, err.Error(), "must include a tag or digest")
}

func TestSignJSON(t *testing.T) {
	tests := []struct {
		name   string
		result signResult
	}{
		{
			name: "basic sign",
			result: signResult{
				Ref:             "ghcr.io/test:v1",
				SignatureDigest: "sha256:abc123",
				Status:          "success",
			},
		},
		{
			name: "sign with resolved ref",
			result: signResult{
				Ref:             "myalias:v1",
				ResolvedRef:     "ghcr.io/acme/configs:v1",
				SignatureDigest: "sha256:def456",
				Status:          "success",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := signJSON(&tt.result)

			w.Close()
			os.Stdout = oldStdout

			var buf bytes.Buffer
			buf.ReadFrom(r)

			require.NoError(t, err)

			// Parse the JSON and verify fields
			var got signResult
			err = json.Unmarshal(buf.Bytes(), &got)
			require.NoError(t, err)
			assert.Equal(t, tt.result.Ref, got.Ref)
			assert.Equal(t, tt.result.ResolvedRef, got.ResolvedRef)
			assert.Equal(t, tt.result.SignatureDigest, got.SignatureDigest)
			assert.Equal(t, tt.result.Status, got.Status)
		})
	}
}
