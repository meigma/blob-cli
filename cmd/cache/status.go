package cache

import (
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show cache sizes for all cache types",
	Long: `Show cache sizes for all cache types.

Displays the size and entry count for each cache type, as well
as the total cache size.`,
	Example: `  blob cache status
  blob cache status --output json`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}
