package cmd

import (
	"github.com/spf13/cobra"
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
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}
