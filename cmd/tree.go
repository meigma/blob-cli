package cmd

import (
	"github.com/spf13/cobra"
)

var treeCmd = &cobra.Command{
	Use:   "tree <ref> [path]",
	Short: "Display directory structure as a tree",
	Long: `Display directory structure as a tree.

Shows the hierarchical structure of files and directories in an
archive, similar to the tree command.`,
	Example: `  blob tree ghcr.io/acme/configs:v1.0.0
  blob tree -L 2 ghcr.io/acme/configs:v1.0.0 /etc`,
	Args: cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func init() {
	treeCmd.Flags().IntP("level", "L", 0, "descend only n levels deep (0 = unlimited)")
	treeCmd.Flags().Bool("dirsfirst", false, "list directories before files")
}
