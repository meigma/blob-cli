package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/meigma/blob"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	internalcfg "github.com/meigma/blob-cli/internal/config"
)

var tagCmd = &cobra.Command{
	Use:   "tag <src-ref> <dst-ref>",
	Short: "Tag an existing manifest with a new reference",
	Long: `Tag an existing manifest with a new reference.

Creates a new tag pointing to the same manifest as the source
reference. This operation does not copy data, only creates a
new reference to the existing content.`,
	Example: `  blob tag ghcr.io/acme/configs:v1.0.0 ghcr.io/acme/configs:latest
  blob tag ghcr.io/acme/configs@sha256:abc... ghcr.io/acme/configs:stable`,
	Args: cobra.ExactArgs(2),
	RunE: runTag,
}

// tagResult contains the result of a tag operation.
type tagResult struct {
	SrcRef         string `json:"src_ref"`
	ResolvedSrcRef string `json:"resolved_src_ref,omitempty"`
	DstRef         string `json:"dst_ref"`
	ResolvedDstRef string `json:"resolved_dst_ref,omitempty"`
	Digest         string `json:"digest"`
	Status         string `json:"status"`
}

func runTag(cmd *cobra.Command, args []string) error {
	cfg := internalcfg.FromContext(cmd.Context())
	if cfg == nil {
		return errors.New("configuration not loaded")
	}

	srcRef := args[0]
	dstRef := args[1]

	resolvedSrcRef := cfg.ResolveAlias(srcRef)
	resolvedDstRef := cfg.ResolveAlias(dstRef)

	client, err := blob.NewClient(blob.WithDockerConfig())
	if err != nil {
		return fmt.Errorf("creating client: %w", err)
	}

	ctx := cmd.Context()

	manifest, err := client.Fetch(ctx, resolvedSrcRef)
	if err != nil {
		return fmt.Errorf("fetching source manifest: %w", err)
	}

	digest := manifest.Digest()

	if err := client.Tag(ctx, resolvedDstRef, digest); err != nil {
		return fmt.Errorf("tagging manifest: %w", err)
	}

	result := tagResult{
		SrcRef: srcRef,
		DstRef: dstRef,
		Digest: digest,
		Status: "success",
	}

	if srcRef != resolvedSrcRef {
		result.ResolvedSrcRef = resolvedSrcRef
	}
	if dstRef != resolvedDstRef {
		result.ResolvedDstRef = resolvedDstRef
	}

	return outputTagResult(cfg, &result)
}

// outputTagResult formats and outputs the tag result.
func outputTagResult(cfg *internalcfg.Config, result *tagResult) error {
	if cfg.Quiet {
		return nil
	}
	if viper.GetString("output") == internalcfg.OutputJSON {
		return tagJSON(result)
	}
	return tagText(result)
}

func tagJSON(result *tagResult) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}

func tagText(result *tagResult) error {
	fmt.Printf("Tagged %s\n", result.DstRef)
	if result.ResolvedDstRef != "" {
		fmt.Printf("  Resolved: %s\n", result.ResolvedDstRef)
	}
	fmt.Printf("Source: %s\n", result.SrcRef)
	if result.ResolvedSrcRef != "" {
		fmt.Printf("  Resolved: %s\n", result.ResolvedSrcRef)
	}
	fmt.Printf("Digest: %s\n", result.Digest)
	return nil
}
