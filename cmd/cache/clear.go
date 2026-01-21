package cache

import (
	"github.com/spf13/cobra"
)

var clearCmd = &cobra.Command{
	Use:   "clear [type]",
	Short: "Clear caches",
	Long: `Clear caches. Clears all caches by default.

Cache types:
  content     File content cache (deduplicated across archives)
  manifests   OCI manifest cache
  indexes     Archive index cache
  all         All caches (default)`,
	Example: `  blob cache clear              # Clear all caches (prompts for confirmation)
  blob cache clear --force      # Clear all without prompting
  blob cache clear content      # Clear only content cache
  blob cache clear manifests    # Clear only manifest cache`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func init() {
	clearCmd.Flags().Bool("force", false, "skip confirmation prompt")
}
