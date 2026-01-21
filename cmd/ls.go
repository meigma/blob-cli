package cmd

import (
	"github.com/spf13/cobra"
)

var lsCmd = &cobra.Command{
	Use:   "ls <ref> [path]",
	Short: "List files and directories in an archive",
	Long: `List files and directories in an archive.

Lists the contents of an archive at the specified path. If no path
is provided, lists the root directory.`,
	Example: `  blob ls ghcr.io/acme/configs:v1.0.0
  blob ls -lh ghcr.io/acme/configs:v1.0.0 /etc
  blob ls --digest ghcr.io/acme/configs:v1.0.0`,
	Args: cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func init() {
	lsCmd.Flags().BoolP("long", "l", false, "long format (permissions, size, hash)")
	lsCmd.Flags().BoolP("human", "h", false, "human-readable sizes (use with -l)")
	lsCmd.Flags().Bool("digest", false, "show file digests")
}
