package cmd

import (
	"github.com/spf13/cobra"
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
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func init() {
	pullCmd.Flags().StringArray("policy", nil, "policy file for verification (repeatable)")
	pullCmd.Flags().String("policy-rego", "", "OPA Rego policy file")
	pullCmd.Flags().String("policy-bundle", "", "OPA bundle for policy evaluation")
	pullCmd.Flags().Bool("no-default-policy", false, "skip policies from config file")
}
