package alias

import (
	"github.com/spf13/cobra"
)

var setCmd = &cobra.Command{
	Use:   "set <name> <ref>",
	Short: "Add or update an alias",
	Long: `Add or update an alias.

Creates a new alias or updates an existing one. The alias maps
a short name to a full registry reference. The reference may
optionally include a tag.`,
	Example: `  blob alias set foo ghcr.io/acme/repo/foo
  blob alias set prod ghcr.io/acme/repo/app:stable`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}
