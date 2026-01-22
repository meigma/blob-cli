package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/meigma/blob-cli/cmd/alias"
	"github.com/meigma/blob-cli/cmd/cache"
	"github.com/meigma/blob-cli/cmd/config"
	internalcfg "github.com/meigma/blob-cli/internal/config"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "blob",
	Short: "A CLI for working with blob archives in OCI registries",
	Long: `blob is a command-line tool for pushing, pulling, and inspecting
blob archives stored in OCI-compliant container registries.

Archives support random access via HTTP range requests, enabling efficient
retrieval of individual files without downloading the entire archive.`,
	SilenceUsage:  true,
	SilenceErrors: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Load typed configuration from Viper
		cfg, err := internalcfg.LoadFromViper()
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		// Attach config to context for use by subcommands
		ctx := internalcfg.WithConfig(cmd.Context(), cfg)
		cmd.SetContext(ctx)

		return nil
	},
}

func Execute() error {
	ctx := context.Background()
	return rootCmd.ExecuteContext(ctx)
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: $XDG_CONFIG_HOME/blob/config.yaml)")
	rootCmd.PersistentFlags().String("output", "text", "output format: text, json")
	rootCmd.PersistentFlags().CountP("verbose", "v", "increase verbosity (can be repeated: -vv, -vvv)")
	rootCmd.PersistentFlags().BoolP("quiet", "q", false, "suppress non-error output")
	rootCmd.PersistentFlags().Bool("no-color", false, "disable colored output")

	// Bind flags to Viper
	// Note: "config" is NOT bound to Viper to avoid BLOB_CONFIG env var affecting
	// config path/edit commands differently than actual config loading.
	viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output"))
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("quiet", rootCmd.PersistentFlags().Lookup("quiet"))
	viper.BindPFlag("no-color", rootCmd.PersistentFlags().Lookup("no-color"))

	// Add core commands
	rootCmd.AddCommand(pushCmd)
	rootCmd.AddCommand(pullCmd)
	rootCmd.AddCommand(cpCmd)
	rootCmd.AddCommand(catCmd)
	rootCmd.AddCommand(lsCmd)
	rootCmd.AddCommand(inspectCmd)
	rootCmd.AddCommand(treeCmd)
	rootCmd.AddCommand(signCmd)
	rootCmd.AddCommand(verifyCmd)
	rootCmd.AddCommand(tagCmd)

	// Add subcommand groups
	rootCmd.AddCommand(cache.Cmd)
	rootCmd.AddCommand(alias.Cmd)
	rootCmd.AddCommand(config.Cmd)
}

func initConfig() {
	// Set defaults before loading config
	internalcfg.SetDefaults(viper.GetViper())

	// Determine and store the effective config path
	// This is stored BEFORE AutomaticEnv to avoid BLOB_CONFIG affecting it
	var effectivePath string
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
		effectivePath = cfgFile
	} else {
		configHome := os.Getenv("XDG_CONFIG_HOME")
		if configHome == "" {
			home, err := os.UserHomeDir()
			if err != nil {
				fmt.Fprintln(os.Stderr, "Warning: could not determine home directory:", err)
				return
			}
			configHome = filepath.Join(home, ".config")
		}

		viper.AddConfigPath(filepath.Join(configHome, "blob"))
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
		effectivePath = filepath.Join(configHome, "blob", "config.yaml")
	}

	// Store the effective path using an internal key (not affected by AutomaticEnv)
	viper.Set("internal.config_path", effectivePath)

	viper.SetEnvPrefix("BLOB")
	viper.AutomaticEnv()

	// Config file is optional - don't fail if missing
	viper.ReadInConfig() //nolint:errcheck // config file is optional
}
