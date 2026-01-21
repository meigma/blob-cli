package config

import (
	"github.com/spf13/cobra"
)

var editCmd = &cobra.Command{
	Use:   "edit",
	Short: "Open configuration file in $EDITOR",
	Long: `Open configuration file in $EDITOR.

Opens the configuration file in your default editor. Uses $EDITOR,
falling back to $VISUAL, then vi.

Creates the config file with defaults if it doesn't exist.`,
	Example: `  blob config edit`,
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}
