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
	"github.com/meigma/blob/registry/oras"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	internalcfg "github.com/meigma/blob-cli/internal/config"
)

var signCmd = &cobra.Command{
	Use:   "sign <ref>",
	Short: "Sign an archive using Sigstore keyless signing",
	Long: `Sign an archive using Sigstore keyless signing.

Signs the specified archive reference using Sigstore. By default,
uses keyless signing which authenticates via OIDC. A private key
can be specified for key-based signing instead.`,
	Example: `  blob sign ghcr.io/acme/configs:v1.0.0
  blob sign --key cosign.key ghcr.io/acme/configs:v1.0.0
  blob sign --output-signature ghcr.io/acme/configs:v1.0.0 > sig.json`,
	Args: cobra.ExactArgs(1),
	RunE: runSign,
}

func init() {
	signCmd.Flags().String("key", "", "sign with a private key instead of keyless")
	signCmd.Flags().Bool("output-signature", false, "print signature to stdout instead of uploading")
}

// signResult contains the result of a sign operation.
type signResult struct {
	Ref             string `json:"ref"`
	ResolvedRef     string `json:"resolved_ref,omitempty"`
	SignatureDigest string `json:"signature_digest,omitempty"`
	Status          string `json:"status"`
}

// signFlags holds the parsed command flags.
type signFlags struct {
	keyPath         string
	outputSignature bool
}

func runSign(cmd *cobra.Command, args []string) error {
	// 1. Get config from context
	cfg := internalcfg.FromContext(cmd.Context())
	if cfg == nil {
		return errors.New("configuration not loaded")
	}

	// 2. Parse arguments
	inputRef := args[0]

	// 3. Parse flags
	flags, err := parseSignFlags(cmd)
	if err != nil {
		return err
	}

	// 4. Resolve alias
	resolvedRef := cfg.ResolveAlias(inputRef)

	// 5. Build signer
	signer, err := buildSigner(flags)
	if err != nil {
		return fmt.Errorf("creating signer: %w", err)
	}

	// 6. Handle two output modes
	ctx := cmd.Context()
	var result signResult
	result.Ref = inputRef
	if inputRef != resolvedRef {
		result.ResolvedRef = resolvedRef
	}

	if flags.outputSignature {
		// Output mode: sign and print to stdout
		return signToStdout(ctx, resolvedRef, signer)
	}

	// Normal mode: sign and upload
	client, err := blob.NewClient(blob.WithDockerConfig())
	if err != nil {
		return fmt.Errorf("creating client: %w", err)
	}

	sigDigest, err := client.Sign(ctx, resolvedRef, signer)
	if err != nil {
		return fmt.Errorf("signing archive: %w", err)
	}

	result.SignatureDigest = sigDigest
	result.Status = "success"

	return outputSignResult(cfg, &result)
}

// parseSignFlags extracts and validates flags from the command.
func parseSignFlags(cmd *cobra.Command) (signFlags, error) {
	var flags signFlags
	var err error

	flags.keyPath, err = cmd.Flags().GetString("key")
	if err != nil {
		return flags, fmt.Errorf("reading key flag: %w", err)
	}

	flags.outputSignature, err = cmd.Flags().GetBool("output-signature")
	if err != nil {
		return flags, fmt.Errorf("reading output-signature flag: %w", err)
	}

	return flags, nil
}

// buildSigner creates a signer based on the flags.
func buildSigner(flags signFlags) (*sigstore.Signer, error) {
	if flags.keyPath != "" {
		// Key-based signing
		pemData, err := os.ReadFile(flags.keyPath)
		if err != nil {
			return nil, fmt.Errorf("reading key file: %w", err)
		}

		// Password from BLOB_KEY_PASSWORD env var (optional, for encrypted keys)
		var password []byte
		if pwd := os.Getenv("BLOB_KEY_PASSWORD"); pwd != "" {
			password = []byte(pwd)
		}

		return sigstore.NewSigner(
			sigstore.WithPrivateKeyPEM(pemData, password),
			sigstore.WithRekor("https://rekor.sigstore.dev"),
		)
	}

	// Keyless signing (default)
	return sigstore.NewSigner(
		sigstore.WithEphemeralKey(),
		sigstore.WithFulcio("https://fulcio.sigstore.dev"),
		sigstore.WithRekor("https://rekor.sigstore.dev"),
		sigstore.WithAmbientCredentials(),
	)
}

// signToStdout fetches the manifest and signs it, writing the signature bundle to stdout.
func signToStdout(ctx context.Context, ref string, signer *sigstore.Signer) error {
	// Extract and validate the reference portion (tag or digest)
	reference := extractReference(ref)
	if reference == "" {
		return fmt.Errorf("invalid reference %q: must include a tag or digest", ref)
	}

	// Create OCI client to fetch raw manifest bytes
	ociClient := oras.New(oras.WithDockerConfig())

	// Resolve the reference to get the descriptor
	desc, err := ociClient.Resolve(ctx, ref, reference)
	if err != nil {
		return fmt.Errorf("resolving reference: %w", err)
	}

	// Fetch the raw manifest bytes (not re-serialized)
	_, rawManifest, err := ociClient.FetchManifest(ctx, ref, &desc)
	if err != nil {
		return fmt.Errorf("fetching manifest: %w", err)
	}

	// Sign the raw manifest bytes
	sig, err := signer.Sign(ctx, rawManifest)
	if err != nil {
		return fmt.Errorf("signing manifest: %w", err)
	}

	// Write signature bundle to stdout
	_, err = os.Stdout.Write(sig.Data)
	if err != nil {
		return fmt.Errorf("writing signature: %w", err)
	}

	return nil
}

// extractReference extracts the tag or digest portion from a reference string.
func extractReference(ref string) string {
	// Find @ for digest references
	if idx := strings.LastIndex(ref, "@"); idx != -1 {
		return ref[idx+1:]
	}
	// Find : for tag references (after the last /)
	lastSlash := strings.LastIndex(ref, "/")
	if idx := strings.LastIndex(ref[lastSlash+1:], ":"); idx != -1 {
		return ref[lastSlash+1+idx+1:]
	}
	return ""
}

// outputSignResult formats and outputs the sign result.
func outputSignResult(cfg *internalcfg.Config, result *signResult) error {
	if cfg.Quiet {
		return nil
	}
	if viper.GetString("output") == internalcfg.OutputJSON {
		return signJSON(result)
	}
	return signText(result)
}

func signJSON(result *signResult) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}

func signText(result *signResult) error {
	fmt.Printf("Signed %s\n", result.Ref)
	if result.ResolvedRef != "" {
		fmt.Printf("  Resolved: %s\n", result.ResolvedRef)
	}
	fmt.Printf("Signature: %s\n", result.SignatureDigest)
	return nil
}
