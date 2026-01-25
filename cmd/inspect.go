package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/meigma/blob"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/meigma/blob-cli/internal/archive"
	internalcfg "github.com/meigma/blob-cli/internal/config"
)

const (
	// sigstoreArtifactType is the OCI artifact type for sigstore bundles.
	sigstoreArtifactType = "application/vnd.dev.sigstore.bundle.v0.3+json"
	// inTotoArtifactType is the OCI artifact type for in-toto attestations.
	inTotoArtifactType = "application/vnd.in-toto+json"
)

var inspectCmd = &cobra.Command{
	Use:   "inspect <ref>",
	Short: "Show metadata about an archive",
	Long: `Show metadata about an archive without downloading it.

Displays information including:
  - Manifest digest
  - Total file count
  - Total size (compressed/uncompressed)
  - Compression type
  - Signatures (if any)
  - Attestations (if any)
  - Annotations`,
	Example: `  blob inspect ghcr.io/acme/configs:v1.0.0
  blob inspect --output json ghcr.io/acme/configs:v1.0.0`,
	Args: cobra.ExactArgs(1),
	RunE: runInspect,
}

func init() {
	inspectCmd.Flags().Bool("skip-cache", false, "bypass registry caches for this operation")
}

// inspectOutput contains the inspect output data for JSON format.
type inspectOutput struct {
	Ref          string            `json:"ref"`
	ResolvedRef  string            `json:"resolved_ref,omitempty"`
	Digest       string            `json:"digest"`
	Created      string            `json:"created,omitempty"`
	Files        int               `json:"files"`
	Size         sizeInfo          `json:"size"`
	Compression  string            `json:"compression"`
	Signatures   []referrerInfo    `json:"signatures,omitempty"`
	Attestations []referrerInfo    `json:"attestations,omitempty"`
	Annotations  map[string]string `json:"annotations,omitempty"`
}

// sizeInfo contains size information.
type sizeInfo struct {
	Compressed   uint64  `json:"compressed"`
	Uncompressed uint64  `json:"uncompressed"`
	Ratio        float64 `json:"ratio"`
}

// referrerInfo contains information about a signature or attestation.
type referrerInfo struct {
	Digest       string            `json:"digest"`
	ArtifactType string            `json:"artifact_type"`
	Annotations  map[string]string `json:"annotations,omitempty"`
}

func runInspect(cmd *cobra.Command, args []string) error {
	cfg := internalcfg.FromContext(cmd.Context())
	if cfg == nil {
		return errors.New("configuration not loaded")
	}

	inputRef := args[0]
	resolvedRef := cfg.ResolveAlias(inputRef)
	skipCache, err := cmd.Flags().GetBool("skip-cache")
	if err != nil {
		return fmt.Errorf("reading skip-cache flag: %w", err)
	}

	var opts archive.InspectOptions
	if skipCache {
		opts.ClientOpts = clientOptsNoCache(cfg)
		opts.InspectOpts = []blob.InspectOption{blob.InspectWithSkipCache()}
	} else {
		opts.ClientOpts = clientOpts(cfg)
	}

	result, err := archive.InspectWithOptions(cmd.Context(), resolvedRef, opts)
	if err != nil {
		return err
	}

	compression := determineCompression(result.Index())

	// Fetch referrers (signatures and attestations).
	ctx := cmd.Context()
	signatures, sigErr := result.Referrers(ctx, sigstoreArtifactType)
	attestations, attErr := result.Referrers(ctx, inTotoArtifactType)

	output := buildInspectOutput(inputRef, resolvedRef, result, compression, signatures, attestations)

	if cfg.Quiet {
		return nil
	}

	// Warn on unexpected referrer errors (ignore ErrReferrersUnsupported).
	// Placed after quiet check to respect --quiet flag.
	warnReferrerError(sigErr, "signatures")
	warnReferrerError(attErr, "attestations")

	if viper.GetString("output") == internalcfg.OutputJSON {
		return inspectJSON(&output)
	}
	return inspectText(&output)
}

// warnReferrerError prints a warning to stderr for unexpected referrer errors.
// ErrReferrersUnsupported is silently ignored since many registries don't support referrers.
func warnReferrerError(err error, kind string) {
	if err == nil || errors.Is(err, blob.ErrReferrersUnsupported) {
		return
	}
	fmt.Fprintf(os.Stderr, "Warning: failed to fetch %s: %v\n", kind, err)
}

// determineCompression checks entries for compression type.
func determineCompression(index *blob.IndexView) string {
	for entry := range index.Entries() {
		if entry.Compression() == blob.CompressionZstd {
			return "zstd"
		}
	}
	return "none"
}

// buildInspectOutput creates the output struct from inspection result.
func buildInspectOutput(
	inputRef, resolvedRef string,
	result *blob.InspectResult,
	compression string,
	signatures, attestations []blob.Referrer,
) inspectOutput {
	output := inspectOutput{
		Ref:         inputRef,
		Digest:      result.Digest(),
		Files:       result.FileCount(),
		Compression: compression,
		Annotations: result.Manifest().Annotations(),
		Size: sizeInfo{
			Compressed:   result.TotalCompressedSize(),
			Uncompressed: result.TotalUncompressedSize(),
			Ratio:        result.CompressionRatio(),
		},
	}

	if inputRef != resolvedRef {
		output.ResolvedRef = resolvedRef
	}

	if created := result.Created(); !created.IsZero() {
		output.Created = created.Format(time.RFC3339)
	}

	output.Signatures = convertReferrers(signatures)
	output.Attestations = convertReferrers(attestations)

	return output
}

// convertReferrers converts blob.Referrer slice to referrerInfo slice.
func convertReferrers(refs []blob.Referrer) []referrerInfo {
	if len(refs) == 0 {
		return nil
	}
	result := make([]referrerInfo, len(refs))
	for i, r := range refs {
		result[i] = referrerInfo{
			Digest:       r.Digest,
			ArtifactType: r.ArtifactType,
			Annotations:  r.Annotations,
		}
	}
	return result
}

func inspectJSON(output *inspectOutput) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(output)
}

func inspectText(output *inspectOutput) error {
	fmt.Printf("Reference:    %s\n", output.Ref)
	if output.ResolvedRef != "" {
		fmt.Printf("Resolved:     %s\n", output.ResolvedRef)
	}
	fmt.Printf("Digest:       %s\n", output.Digest)
	fmt.Printf("Files:        %d\n", output.Files)
	fmt.Printf("Size:         %s (%s uncompressed)\n",
		archive.FormatSize(output.Size.Compressed),
		archive.FormatSize(output.Size.Uncompressed))
	fmt.Printf("Compression:  %s\n", output.Compression)
	if output.Created != "" {
		fmt.Printf("Created:      %s\n", output.Created)
	}

	if len(output.Signatures) > 0 {
		fmt.Println()
		fmt.Println("Signatures:")
		for _, sig := range output.Signatures {
			fmt.Printf("  %s\n", sig.Digest)
		}
	}

	if len(output.Attestations) > 0 {
		fmt.Println()
		fmt.Println("Attestations:")
		for _, att := range output.Attestations {
			fmt.Printf("  %s\n", att.Digest)
		}
	}

	if len(output.Annotations) > 0 {
		fmt.Println()
		fmt.Println("Annotations:")
		for k, v := range output.Annotations {
			fmt.Printf("  %s: %s\n", k, v)
		}
	}

	return nil
}
