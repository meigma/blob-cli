package config

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "config",
	Short: "View and manage CLI configuration",
	Long: `View and manage CLI configuration.

Configuration is read from multiple sources in order of precedence:
  1. Command-line flags
  2. Environment variables (BLOB_*)
  3. Config file
  4. Defaults`,
}

func init() {
	Cmd.AddCommand(showCmd)
	Cmd.AddCommand(pathCmd)
	Cmd.AddCommand(editCmd)
}
