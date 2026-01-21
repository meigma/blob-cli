package config

import (
	"github.com/spf13/cobra"
)

var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Display current configuration",
	Long: `Display current configuration.

Shows the effective configuration merged from all sources (defaults,
config file, environment variables). Each value is annotated with
its source.`,
	Example: `  blob config show
  blob config show --output json
  blob config show --resolved`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func init() {
	showCmd.Flags().Bool("resolved", false, "show fully resolved values (expand env vars)")
}
