package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/meigma/blob"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/meigma/blob-cli/internal/archive"
	internalcfg "github.com/meigma/blob-cli/internal/config"
	"github.com/meigma/blob-cli/internal/policy"
)

var pullCmd = &cobra.Command{
	Use:   "pull <ref> [path]",
	Short: "Pull an archive from an OCI registry to a local directory",
	Long: `Pull an archive from an OCI registry to a local directory.

Downloads and extracts the blob archive to the specified destination
directory. If no path is provided, extracts to the current directory.

Verification policies can be specified to enforce signature and
attestation requirements before extraction.`,
	Example: `  blob pull ghcr.io/acme/configs:v1.0.0 ./local
  blob pull foo:v1 ./local                          # Using alias
  blob pull --policy policy.yaml ghcr.io/acme/configs:v1.0.0
  blob pull --no-default-policy foo:v1 ./local      # Skip config policies`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runPull,
}

func init() {
	pullCmd.Flags().StringArray("policy", nil, "policy file for verification (repeatable)")
	pullCmd.Flags().String("policy-rego", "", "OPA Rego policy file")
	pullCmd.Flags().Bool("no-default-policy", false, "skip policies from config file")
}

// pullResult contains the result of a pull operation.
type pullResult struct {
	Ref            string `json:"ref"`
	ResolvedRef    string `json:"resolved_ref,omitempty"`
	Destination    string `json:"destination"`
	FileCount      int    `json:"file_count"`
	TotalSize      uint64 `json:"total_size"`
	TotalSizeHuman string `json:"total_size_human,omitempty"`
	Verified       bool   `json:"verified"`
	PoliciesCount  int    `json:"policies_applied,omitempty"`
}

// pullFlags holds the parsed command flags.
type pullFlags struct {
	policyFiles     []string
	policyRego      string
	noDefaultPolicy bool
}

func runPull(cmd *cobra.Command, args []string) error {
	// 1. Get config from context
	cfg := internalcfg.FromContext(cmd.Context())
	if cfg == nil {
		return errors.New("configuration not loaded")
	}

	// 2. Parse arguments
	inputRef := args[0]
	destDir := "."
	if len(args) > 1 {
		destDir = args[1]
	}

	// 3. Parse flags
	flags, err := parsePullFlags(cmd)
	if err != nil {
		return err
	}

	// 4. Resolve alias FIRST (before policy matching)
	resolvedRef := cfg.ResolveAlias(inputRef)

	// 5. Build policies from config + flags (before creating destination)
	policies, err := policy.BuildPolicies(
		cfg,
		resolvedRef,
		flags.policyFiles,
		flags.policyRego,
		flags.noDefaultPolicy,
	)
	if err != nil {
		return fmt.Errorf("building policies: %w", err)
	}

	// 6. Create client with policies
	clientOpts := []blob.Option{blob.WithDockerConfig()}
	for _, p := range policies {
		clientOpts = append(clientOpts, blob.WithPolicy(p))
	}
	client, err := blob.NewClient(clientOpts...)
	if err != nil {
		return fmt.Errorf("creating client: %w", err)
	}

	// 7. Pull archive (policy verification happens here)
	ctx := cmd.Context()
	blobArchive, err := client.Pull(ctx, resolvedRef)
	if err != nil {
		if errors.Is(err, blob.ErrPolicyViolation) {
			return fmt.Errorf("verification failed: %w", err)
		}
		return fmt.Errorf("pulling archive: %w", err)
	}

	// 8. Prepare destination directory (only after successful pull)
	destDir, err = prepareDestination(destDir)
	if err != nil {
		return err
	}

	// 9. Extract files
	copyOpts := []blob.CopyOption{
		blob.CopyWithOverwrite(false),
		blob.CopyWithPreserveMode(true),
		blob.CopyWithPreserveTimes(true),
	}
	if err := blobArchive.CopyDir(destDir, ".", copyOpts...); err != nil {
		return fmt.Errorf("extracting files: %w", err)
	}

	// 10. Build result
	result := pullResult{
		Ref:         inputRef,
		Destination: destDir,
		FileCount:   blobArchive.Len(),
		Verified:    len(policies) > 0,
	}

	if inputRef != resolvedRef {
		result.ResolvedRef = resolvedRef
	}

	// Compute total size
	for entry := range blobArchive.Entries() {
		result.TotalSize += entry.OriginalSize()
	}
	result.TotalSizeHuman = archive.FormatSize(result.TotalSize)

	if len(policies) > 0 {
		result.PoliciesCount = len(policies)
	}

	// 11. Output result
	return outputPullResult(cfg, &result)
}

// parsePullFlags extracts and validates flags from the command.
func parsePullFlags(cmd *cobra.Command) (pullFlags, error) {
	var flags pullFlags
	var err error

	flags.policyFiles, err = cmd.Flags().GetStringArray("policy")
	if err != nil {
		return flags, fmt.Errorf("reading policy flag: %w", err)
	}

	flags.policyRego, err = cmd.Flags().GetString("policy-rego")
	if err != nil {
		return flags, fmt.Errorf("reading policy-rego flag: %w", err)
	}

	flags.noDefaultPolicy, err = cmd.Flags().GetBool("no-default-policy")
	if err != nil {
		return flags, fmt.Errorf("reading no-default-policy flag: %w", err)
	}

	return flags, nil
}

// prepareDestination validates and prepares the destination directory.
func prepareDestination(destDir string) (string, error) {
	// Convert to absolute path
	absPath, err := filepath.Abs(destDir)
	if err != nil {
		return "", fmt.Errorf("resolving path: %w", err)
	}

	// Check if path exists
	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Create directory with restrictive permissions
			if mkdirErr := os.MkdirAll(absPath, 0o750); mkdirErr != nil {
				return "", fmt.Errorf("creating directory: %w", mkdirErr)
			}
			return absPath, nil
		}
		return "", fmt.Errorf("accessing path: %w", err)
	}

	// Path exists - must be a directory
	if !info.IsDir() {
		return "", fmt.Errorf("destination is not a directory: %s", absPath)
	}

	return absPath, nil
}

// outputPullResult formats and outputs the pull result.
func outputPullResult(cfg *internalcfg.Config, result *pullResult) error {
	if cfg.Quiet {
		return nil
	}
	if viper.GetString("output") == internalcfg.OutputJSON {
		return pullJSON(result)
	}
	return pullText(result)
}

func pullJSON(result *pullResult) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}

func pullText(result *pullResult) error {
	fmt.Printf("Pulled %s\n", result.Ref)
	if result.ResolvedRef != "" {
		fmt.Printf("  Resolved: %s\n", result.ResolvedRef)
	}
	fmt.Printf("  Destination: %s\n", result.Destination)
	fmt.Printf("  Files: %d\n", result.FileCount)
	fmt.Printf("  Size: %s\n", result.TotalSizeHuman)

	if result.Verified {
		fmt.Printf("  Verified: %d policies applied\n", result.PoliciesCount)
	}

	return nil
}
