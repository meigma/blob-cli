package alias

import (
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all configured aliases",
	Long: `List all configured aliases.

Displays all aliases defined in the configuration file along with
their target references.`,
	Example: `  blob alias list
  blob alias list --output json`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}
