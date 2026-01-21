package cmd

import (
	"github.com/spf13/cobra"
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
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}
