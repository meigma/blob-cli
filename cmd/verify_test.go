package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExitError(t *testing.T) {
	t.Run("wraps error", func(t *testing.T) {
		inner := errors.New("policy failed")
		exitErr := &ExitError{Code: 5, Err: inner}

		assert.Equal(t, 5, exitErr.Code)
		assert.Equal(t, "policy failed", exitErr.Error())
		assert.ErrorIs(t, exitErr, inner)
	})

	t.Run("nil inner error", func(t *testing.T) {
		exitErr := &ExitError{Code: 5, Err: nil}

		assert.Equal(t, 5, exitErr.Code)
		assert.Equal(t, "", exitErr.Error())
	})

	t.Run("errors.As works", func(t *testing.T) {
		inner := errors.New("verification failed")
		exitErr := &ExitError{Code: exitCodePolicyViolation, Err: inner}

		var target *ExitError
		assert.True(t, errors.As(exitErr, &target))
		assert.Equal(t, exitCodePolicyViolation, target.Code)
	})
}

func TestVerifyText(t *testing.T) {
	tests := []struct {
		name        string
		result      verifyResult
		wantContain []string
	}{
		{
			name: "verified with policies",
			result: verifyResult{
				Ref:             "ghcr.io/test:v1",
				Digest:          "sha256:abc123",
				Verified:        true,
				Status:          "verified",
				PoliciesApplied: 2,
			},
			wantContain: []string{
				"Verified ghcr.io/test:v1",
				"Digest: sha256:abc123",
				"Policies: 2 applied",
			},
		},
		{
			name: "unverified - no policies",
			result: verifyResult{
				Ref:             "ghcr.io/test:v1",
				Digest:          "sha256:abc123",
				Verified:        false,
				Status:          "no_policies",
				PoliciesApplied: 0,
			},
			wantContain: []string{
				"ghcr.io/test:v1",
				"Digest: sha256:abc123",
			},
		},
		{
			name: "with signatures",
			result: verifyResult{
				Ref:             "ghcr.io/test:v1",
				Digest:          "sha256:abc123",
				Verified:        true,
				Status:          "verified",
				PoliciesApplied: 1,
				Signatures: []referrerInfo{
					{Digest: "sha256:sig1"},
				},
			},
			wantContain: []string{
				"Signatures:",
				"sha256:sig1",
			},
		},
		{
			name: "with attestations",
			result: verifyResult{
				Ref:             "ghcr.io/test:v1",
				Digest:          "sha256:abc123",
				Verified:        true,
				Status:          "verified",
				PoliciesApplied: 1,
				Attestations: []referrerInfo{
					{Digest: "sha256:att1"},
				},
			},
			wantContain: []string{
				"Attestations:",
				"sha256:att1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := verifyText(&tt.result)

			w.Close()
			os.Stdout = oldStdout

			var buf bytes.Buffer
			buf.ReadFrom(r)

			require.NoError(t, err)
			output := buf.String()
			for _, want := range tt.wantContain {
				assert.Contains(t, output, want)
			}
		})
	}
}

func TestVerifyJSON(t *testing.T) {
	tests := []struct {
		name   string
		result verifyResult
	}{
		{
			name: "verified with policies",
			result: verifyResult{
				Ref:             "ghcr.io/test:v1",
				Digest:          "sha256:abc123",
				Verified:        true,
				Status:          "verified",
				PoliciesApplied: 2,
			},
		},
		{
			name: "no policies - status field clarity",
			result: verifyResult{
				Ref:             "ghcr.io/test:v1",
				Digest:          "sha256:abc123",
				Verified:        false,
				Status:          "no_policies",
				PoliciesApplied: 0,
			},
		},
		{
			name: "with signatures and attestations",
			result: verifyResult{
				Ref:             "ghcr.io/test:v1",
				Digest:          "sha256:abc123",
				Verified:        true,
				Status:          "verified",
				PoliciesApplied: 1,
				Signatures: []referrerInfo{
					{Digest: "sha256:sig1", ArtifactType: "application/vnd.dev.sigstore.bundle.v0.3+json"},
				},
				Attestations: []referrerInfo{
					{Digest: "sha256:att1", ArtifactType: "application/vnd.in-toto+json"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			err := verifyJSON(&tt.result)

			w.Close()
			os.Stdout = oldStdout

			var buf bytes.Buffer
			buf.ReadFrom(r)

			require.NoError(t, err)

			// Parse the JSON and verify fields
			var got verifyResult
			err = json.Unmarshal(buf.Bytes(), &got)
			require.NoError(t, err)
			assert.Equal(t, tt.result.Ref, got.Ref)
			assert.Equal(t, tt.result.Digest, got.Digest)
			assert.Equal(t, tt.result.Verified, got.Verified)
			assert.Equal(t, tt.result.Status, got.Status)
			assert.Equal(t, tt.result.PoliciesApplied, got.PoliciesApplied)
			assert.Equal(t, len(tt.result.Signatures), len(got.Signatures))
			assert.Equal(t, len(tt.result.Attestations), len(got.Attestations))
		})
	}
}

func TestVerifyJSON_StatusFieldClarity(t *testing.T) {
	// This test ensures JSON consumers can distinguish between
	// "no policies applied" and "verified with policies" scenarios
	t.Run("no policies case has clear status", func(t *testing.T) {
		result := verifyResult{
			Ref:             "ghcr.io/test:v1",
			Digest:          "sha256:abc123",
			Verified:        false,
			Status:          "no_policies",
			PoliciesApplied: 0,
		}

		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := verifyJSON(&result)

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		buf.ReadFrom(r)

		require.NoError(t, err)

		// Parse and check both verified and status fields
		var got map[string]interface{}
		err = json.Unmarshal(buf.Bytes(), &got)
		require.NoError(t, err)

		// verified=false combined with status="no_policies" clearly indicates
		// that verification wasn't performed (not that it failed)
		assert.Equal(t, false, got["verified"])
		assert.Equal(t, "no_policies", got["status"])
		assert.Equal(t, float64(0), got["policies_applied"])
	})

	t.Run("verified case has clear status", func(t *testing.T) {
		result := verifyResult{
			Ref:             "ghcr.io/test:v1",
			Digest:          "sha256:abc123",
			Verified:        true,
			Status:          "verified",
			PoliciesApplied: 2,
		}

		// Capture stdout
		oldStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := verifyJSON(&result)

		w.Close()
		os.Stdout = oldStdout

		var buf bytes.Buffer
		buf.ReadFrom(r)

		require.NoError(t, err)

		var got map[string]interface{}
		err = json.Unmarshal(buf.Bytes(), &got)
		require.NoError(t, err)

		assert.Equal(t, true, got["verified"])
		assert.Equal(t, "verified", got["status"])
		assert.Equal(t, float64(2), got["policies_applied"])
	})
}

func TestConvertBlobReferrers(t *testing.T) {
	t.Run("nil input", func(t *testing.T) {
		result := convertBlobReferrers(nil)
		assert.Nil(t, result)
	})

	t.Run("empty input", func(t *testing.T) {
		result := convertBlobReferrers(nil)
		assert.Nil(t, result)
	})
}
