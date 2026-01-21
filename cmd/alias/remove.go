package alias

import (
	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove an alias",
	Long: `Remove an alias from the configuration file.

Deletes the specified alias. This action cannot be undone.`,
	Example: `  blob alias remove foo`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}
