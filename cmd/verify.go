package cmd

import (
	"github.com/spf13/cobra"
)

var verifyCmd = &cobra.Command{
	Use:   "verify <ref>",
	Short: "Verify signatures and attestations on an archive",
	Long: `Verify signatures and attestations on an archive.

Checks that the archive meets the specified policy requirements
for signatures and attestations. Policies can be specified as
YAML files or OPA Rego policies.`,
	Example: `  blob verify ghcr.io/acme/configs:v1.0.0
  blob verify --policy policy.yaml ghcr.io/acme/configs:v1.0.0
  blob verify --policy-rego custom.rego ghcr.io/acme/configs:v1.0.0`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func init() {
	verifyCmd.Flags().StringArray("policy", nil, "policy file for verification (repeatable)")
	verifyCmd.Flags().String("policy-rego", "", "OPA Rego policy file")
	verifyCmd.Flags().String("policy-bundle", "", "OPA bundle for policy evaluation")
}
