package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/meigma/blob"
	"github.com/meigma/blob/policy/sigstore"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	internalcfg "github.com/meigma/blob-cli/internal/config"
)

var pushCmd = &cobra.Command{
	Use:   "push <ref> <path>",
	Short: "Push a directory to an OCI registry as a blob archive",
	Long: `Push a directory to an OCI registry as a blob archive.

The directory contents are archived and uploaded to the specified
registry reference. Files are compressed individually using zstd
by default for optimal random access performance.`,
	Example: `  blob push ghcr.io/acme/configs:v1.0.0 ./config
  blob push --sign ghcr.io/acme/configs:latest ./config
  blob push --compression none ghcr.io/acme/data:v1 ./data`,
	Args: cobra.ExactArgs(2),
	RunE: runPush,
}

func init() {
	pushCmd.Flags().StringP("compression", "c", "zstd", "compression type: none, zstd")
	pushCmd.Flags().Bool("skip-compressed", true, "skip compressing already-compressed files")
	pushCmd.Flags().Bool("sign", false, "sign the archive after pushing")
	pushCmd.Flags().StringArray("annotation", nil, "add annotation to manifest (k=v, repeatable)")

	_ = viper.BindPFlag("compression", pushCmd.Flags().Lookup("compression"))
}

// pushResult contains the result of a push operation.
type pushResult struct {
	Ref             string `json:"ref"`
	Status          string `json:"status"`
	Signed          bool   `json:"signed,omitempty"`
	SignatureDigest string `json:"signature_digest,omitempty"`
}

// pushFlags holds the parsed command flags.
type pushFlags struct {
	compression    blob.Compression
	skipCompressed bool
	sign           bool
	annotations    map[string]string
}

func runPush(cmd *cobra.Command, args []string) error {
	ref := args[0]
	srcPath := args[1]

	cfg := internalcfg.FromContext(cmd.Context())
	if cfg == nil {
		return errors.New("configuration not loaded")
	}

	if err := validateSourcePath(srcPath); err != nil {
		return err
	}

	flags, err := parsePushFlags(cmd)
	if err != nil {
		return err
	}

	client, err := blob.NewClient(blob.WithDockerConfig())
	if err != nil {
		return fmt.Errorf("creating client: %w", err)
	}

	pushOpts := buildPushOptions(flags)

	ctx := cmd.Context()
	if err := client.Push(ctx, ref, srcPath, pushOpts...); err != nil {
		return fmt.Errorf("pushing archive: %w", err)
	}

	result := pushResult{
		Ref:    ref,
		Status: "success",
	}

	if flags.sign {
		if err := signArchive(ctx, client, ref, &result); err != nil {
			return err
		}
	}

	return outputPushResult(cfg, result)
}

// parsePushFlags extracts and validates flags from the command.
func parsePushFlags(cmd *cobra.Command) (pushFlags, error) {
	var flags pushFlags
	var err error

	compressionStr := cmd.Flags().Lookup("compression").Value.String()
	flags.compression, err = mapCompression(compressionStr)
	if err != nil {
		return flags, err
	}

	flags.skipCompressed, err = cmd.Flags().GetBool("skip-compressed")
	if err != nil {
		return flags, fmt.Errorf("reading skip-compressed flag: %w", err)
	}

	flags.sign, err = cmd.Flags().GetBool("sign")
	if err != nil {
		return flags, fmt.Errorf("reading sign flag: %w", err)
	}

	annotationStrs, err := cmd.Flags().GetStringArray("annotation")
	if err != nil {
		return flags, fmt.Errorf("reading annotation flag: %w", err)
	}

	flags.annotations, err = parseAnnotations(annotationStrs)
	if err != nil {
		return flags, err
	}

	return flags, nil
}

// buildPushOptions creates blob.PushOption slice from flags.
func buildPushOptions(flags pushFlags) []blob.PushOption {
	opts := []blob.PushOption{
		blob.PushWithCompression(flags.compression),
	}
	if flags.skipCompressed {
		opts = append(opts, blob.PushWithSkipCompression(blob.DefaultSkipCompression(1024)))
	}
	if len(flags.annotations) > 0 {
		opts = append(opts, blob.PushWithAnnotations(flags.annotations))
	}
	return opts
}

// signArchive signs the pushed archive using Sigstore keyless signing.
func signArchive(ctx context.Context, client *blob.Client, ref string, result *pushResult) error {
	signer, err := sigstore.NewSigner(
		sigstore.WithEphemeralKey(),
		sigstore.WithFulcio("https://fulcio.sigstore.dev"),
		sigstore.WithRekor("https://rekor.sigstore.dev"),
		sigstore.WithAmbientCredentials(),
	)
	if err != nil {
		return fmt.Errorf("creating signer: %w", err)
	}

	sigDigest, err := client.Sign(ctx, ref, signer)
	if err != nil {
		return fmt.Errorf("signing archive: %w", err)
	}

	result.Signed = true
	result.SignatureDigest = sigDigest
	return nil
}

// outputPushResult formats and outputs the push result.
func outputPushResult(cfg *internalcfg.Config, result pushResult) error {
	if cfg.Quiet {
		return nil
	}
	if viper.GetString("output") == internalcfg.OutputJSON {
		return pushJSON(result)
	}
	return pushText(result)
}

func pushJSON(result pushResult) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}

func pushText(result pushResult) error {
	fmt.Printf("Pushed %s\n", result.Ref)
	if result.Signed {
		fmt.Printf("Signed: %s\n", result.SignatureDigest)
	}
	return nil
}

// validateSourcePath checks that the path exists and is a directory.
func validateSourcePath(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("source path does not exist: %s", path)
		}
		return fmt.Errorf("accessing source path: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("source path is not a directory: %s", path)
	}
	return nil
}

// mapCompression converts a compression string to a blob.Compression value.
func mapCompression(s string) (blob.Compression, error) {
	switch s {
	case internalcfg.CompressionZstd, "":
		return blob.CompressionZstd, nil
	case internalcfg.CompressionNone:
		return blob.CompressionNone, nil
	default:
		return 0, fmt.Errorf("invalid compression type %q: must be 'none' or 'zstd'", s)
	}
}

// parseAnnotations parses annotation strings in key=value format.
// Returns an empty map (not nil) when annotations is empty.
func parseAnnotations(annotations []string) (map[string]string, error) {
	result := make(map[string]string, len(annotations))
	for _, ann := range annotations {
		idx := strings.Index(ann, "=")
		if idx < 1 {
			return nil, fmt.Errorf("invalid annotation %q: must be key=value", ann)
		}
		result[ann[:idx]] = ann[idx+1:]
	}
	return result, nil
}
