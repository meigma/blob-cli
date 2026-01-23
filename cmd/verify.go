package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/meigma/blob"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/meigma/blob-cli/internal/archive"
	internalcfg "github.com/meigma/blob-cli/internal/config"
	"github.com/meigma/blob-cli/internal/policy"
)

const (
	// exitCodePolicyViolation is the exit code for verification failures.
	exitCodePolicyViolation = 5
)

var verifyCmd = &cobra.Command{
	Use:   "verify <ref>",
	Short: "Verify signatures and attestations on an archive",
	Long: `Verify signatures and attestations on an archive.

Checks that the archive meets the specified policy requirements
for signatures and attestations. Policies can be specified as
YAML files or OPA Rego policies.

If no policies are specified (via flags or config), verification
succeeds with a warning that no verification was performed.`,
	Example: `  blob verify ghcr.io/acme/configs:v1.0.0
  blob verify --policy policy.yaml ghcr.io/acme/configs:v1.0.0
  blob verify --policy-rego custom.rego ghcr.io/acme/configs:v1.0.0
  blob verify --no-default-policy --policy policy.yaml ghcr.io/acme/configs:v1.0.0`,
	Args: cobra.ExactArgs(1),
	RunE: runVerify,
}

func init() {
	verifyCmd.Flags().StringArray("policy", nil, "policy file for verification (repeatable)")
	verifyCmd.Flags().String("policy-rego", "", "OPA Rego policy file")
	verifyCmd.Flags().Bool("no-default-policy", false, "skip policies from config file")
}

// verifyResult contains the result of a verify operation.
type verifyResult struct {
	Ref             string         `json:"ref"`
	ResolvedRef     string         `json:"resolved_ref,omitempty"`
	Digest          string         `json:"digest"`
	Verified        bool           `json:"verified"`
	Status          string         `json:"status"` // "verified", "no_policies"
	PoliciesApplied int            `json:"policies_applied"`
	Signatures      []referrerInfo `json:"signatures,omitempty"`
	Attestations    []referrerInfo `json:"attestations,omitempty"`
}

// verifyFlags holds the parsed command flags.
type verifyFlags struct {
	policyFiles     []string
	policyRego      string
	noDefaultPolicy bool
}

func runVerify(cmd *cobra.Command, args []string) error {
	// 1. Get config from context
	cfg := internalcfg.FromContext(cmd.Context())
	if cfg == nil {
		return errors.New("configuration not loaded")
	}

	// 2. Parse arguments
	inputRef := args[0]

	// 3. Parse flags
	flags, err := parseVerifyFlags(cmd)
	if err != nil {
		return err
	}

	// 4. Resolve alias
	resolvedRef := cfg.ResolveAlias(inputRef)

	// 5. Build policies from config + flags
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

	// 6. Build result
	result := verifyResult{
		Ref:             inputRef,
		PoliciesApplied: len(policies),
	}
	if inputRef != resolvedRef {
		result.ResolvedRef = resolvedRef
	}

	// 7. Handle no-policies case
	if len(policies) == 0 {
		// No policies - vacuous success with warning
		inspectResult, inspectErr := archive.Inspect(cmd.Context(), resolvedRef)
		if inspectErr != nil {
			return fmt.Errorf("inspecting archive: %w", inspectErr)
		}

		result.Digest = inspectResult.Digest()
		result.Verified = false
		result.Status = "no_policies"

		// Fetch referrers for signatures/attestations
		populateReferrers(cmd.Context(), inspectResult, &result)

		// Output warning and result
		if !cfg.Quiet && viper.GetString("output") != internalcfg.OutputJSON {
			fmt.Fprintln(os.Stderr, "Warning: No policies applied - archive not verified")
		}

		return outputVerifyResult(cfg, &result)
	}

	// 8. Create client with policies for verification
	clientOpts := []blob.Option{blob.WithDockerConfig()}
	for _, p := range policies {
		clientOpts = append(clientOpts, blob.WithPolicy(p))
	}
	client, err := blob.NewClient(clientOpts...)
	if err != nil {
		return fmt.Errorf("creating client: %w", err)
	}

	// 9. Verify by calling Inspect (which triggers policy evaluation)
	ctx := cmd.Context()
	inspectResult, err := client.Inspect(ctx, resolvedRef)
	if err != nil {
		if errors.Is(err, blob.ErrPolicyViolation) {
			return &ExitError{
				Code: exitCodePolicyViolation,
				Err:  fmt.Errorf("verification failed: %w", err),
			}
		}
		return fmt.Errorf("verifying archive: %w", err)
	}

	// 10. Verification succeeded
	result.Digest = inspectResult.Digest()
	result.Verified = true
	result.Status = "verified"

	// Fetch referrers for signatures/attestations
	populateReferrers(ctx, inspectResult, &result)

	return outputVerifyResult(cfg, &result)
}

// parseVerifyFlags extracts and validates flags from the command.
func parseVerifyFlags(cmd *cobra.Command) (verifyFlags, error) {
	var flags verifyFlags
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

// populateReferrers fetches signatures and attestations and adds them to the result.
func populateReferrers(ctx context.Context, inspectResult *blob.InspectResult, result *verifyResult) {
	signatures, sigErr := inspectResult.Referrers(ctx, sigstoreArtifactType)
	if sigErr == nil {
		result.Signatures = convertBlobReferrers(signatures)
	} else if !errors.Is(sigErr, blob.ErrReferrersUnsupported) {
		fmt.Fprintf(os.Stderr, "Warning: failed to fetch signatures: %v\n", sigErr)
	}

	attestations, attErr := inspectResult.Referrers(ctx, inTotoArtifactType)
	if attErr == nil {
		result.Attestations = convertBlobReferrers(attestations)
	} else if !errors.Is(attErr, blob.ErrReferrersUnsupported) {
		fmt.Fprintf(os.Stderr, "Warning: failed to fetch attestations: %v\n", attErr)
	}
}

// convertBlobReferrers converts blob.Referrer slice to referrerInfo slice.
func convertBlobReferrers(refs []blob.Referrer) []referrerInfo {
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

// outputVerifyResult formats and outputs the verify result.
func outputVerifyResult(cfg *internalcfg.Config, result *verifyResult) error {
	if cfg.Quiet {
		return nil
	}
	if viper.GetString("output") == internalcfg.OutputJSON {
		return verifyJSON(result)
	}
	return verifyText(result)
}

func verifyJSON(result *verifyResult) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}

func verifyText(result *verifyResult) error {
	if result.Verified {
		fmt.Printf("Verified %s\n", result.Ref)
	} else {
		fmt.Printf("%s\n", result.Ref)
	}

	if result.ResolvedRef != "" {
		fmt.Printf("Resolved: %s\n", result.ResolvedRef)
	}
	fmt.Printf("Digest: %s\n", result.Digest)

	if result.Verified {
		fmt.Printf("Policies: %d applied\n", result.PoliciesApplied)
	}

	if len(result.Signatures) > 0 {
		fmt.Println()
		fmt.Println("Signatures:")
		for _, sig := range result.Signatures {
			fmt.Printf("  %s\n", sig.Digest)
		}
	}

	if len(result.Attestations) > 0 {
		fmt.Println()
		fmt.Println("Attestations:")
		for _, att := range result.Attestations {
			fmt.Printf("  %s\n", att.Digest)
		}
	}

	return nil
}
