package alias

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "alias",
	Short: "Manage reference aliases",
	Long: `Manage reference aliases.

Aliases allow you to use short names for frequently used references.
For example, you can create an alias "foo" for "ghcr.io/acme/repo/foo"
and then use "blob pull foo:v1" instead of the full reference.`,
}

func init() {
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(setCmd)
	Cmd.AddCommand(removeCmd)
}
