package cache

import (
	"github.com/spf13/cobra"
)

var pathCmd = &cobra.Command{
	Use:   "path",
	Short: "Show cache directory paths",
	Long: `Show cache directory paths.

Displays the paths for each cache type. Paths follow the XDG
Base Directory Specification.`,
	Example: `  blob cache path
  blob cache path --output json`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}
