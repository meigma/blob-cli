package cmd

import (
	"github.com/spf13/cobra"
)

var catCmd = &cobra.Command{
	Use:   "cat <ref> <file>...",
	Short: "Print file contents to stdout",
	Long: `Print file contents to stdout.

Useful for viewing, piping, or combining files from an archive.
Uses HTTP range requests to fetch only the requested files without
downloading the entire archive.`,
	Example: `  blob cat ghcr.io/acme/configs:v1.0.0 config.json
  blob cat ghcr.io/acme/configs:v1.0.0 config.json | jq .
  blob cat ghcr.io/acme/configs:v1.0.0 header.txt body.txt footer.txt > combined.txt`,
	Args: cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}
