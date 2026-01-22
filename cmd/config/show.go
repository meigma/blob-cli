package config

import (
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	internalcfg "github.com/meigma/blob-cli/internal/config"
)

var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Display current configuration",
	Long: `Display current configuration.

Shows the effective configuration merged from all sources (defaults,
config file, environment variables).`,
	Example: `  blob config show
  blob config show --output json`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := internalcfg.FromContext(cmd.Context())
		if cfg == nil {
			return errors.New("configuration not loaded")
		}

		output := viper.GetString("output")
		if output == "json" {
			return showJSON(cfg)
		}
		return showText(cfg)
	},
}

func showJSON(cfg *internalcfg.Config) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(cfg)
}

func showText(cfg *internalcfg.Config) error {
	fmt.Println("Configuration")
	fmt.Println(strings.Repeat("-", 50))

	// Core settings
	fmt.Printf("output:       %s\n", cfg.Output)
	fmt.Printf("compression:  %s\n", cfg.Compression)
	fmt.Printf("verbose:      %d\n", cfg.Verbose)
	fmt.Printf("quiet:        %t\n", cfg.Quiet)
	fmt.Printf("no-color:     %t\n", cfg.NoColor)

	// Cache settings
	fmt.Println()
	fmt.Println("cache:")
	fmt.Printf("  enabled:    %t\n", cfg.Cache.Enabled)
	fmt.Printf("  max_size:   %s\n", cfg.Cache.MaxSize)

	// Aliases (sorted for deterministic output)
	fmt.Println()
	if len(cfg.Aliases) == 0 {
		fmt.Println("aliases:      (none)")
	} else {
		fmt.Println("aliases:")
		names := make([]string, 0, len(cfg.Aliases))
		for name := range cfg.Aliases {
			names = append(names, name)
		}
		slices.SortFunc(names, cmp.Compare)
		for _, name := range names {
			fmt.Printf("  %s -> %s\n", name, cfg.Aliases[name])
		}
	}

	// Policies
	fmt.Println()
	if len(cfg.Policies) == 0 {
		fmt.Println("policies:     (none)")
	} else {
		fmt.Println("policies:")
		for _, rule := range cfg.Policies {
			fmt.Printf("  match: %s\n", rule.Match)
		}
	}

	return nil
}
