package cmd

import (
	"github.com/spf13/cobra"
)

var signCmd = &cobra.Command{
	Use:   "sign <ref>",
	Short: "Sign an archive using Sigstore keyless signing",
	Long: `Sign an archive using Sigstore keyless signing.

Signs the specified archive reference using Sigstore. By default,
uses keyless signing which authenticates via OIDC. A private key
can be specified for key-based signing instead.`,
	Example: `  blob sign ghcr.io/acme/configs:v1.0.0
  blob sign --key cosign.key ghcr.io/acme/configs:v1.0.0`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func init() {
	signCmd.Flags().String("key", "", "sign with a private key instead of keyless")
	signCmd.Flags().Bool("output-signature", false, "print signature to stdout instead of uploading")
}
