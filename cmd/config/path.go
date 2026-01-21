package config

import (
	"github.com/spf13/cobra"
)

var pathCmd = &cobra.Command{
	Use:   "path",
	Short: "Show configuration file path",
	Long: `Show configuration file path.

Displays the path to the configuration file. The default location
follows the XDG Base Directory Specification.`,
	Example: `  blob config path`,
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}
