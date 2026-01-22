package alias

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
		cfg := internalcfg.FromContext(cmd.Context())
		if cfg == nil {
			return errors.New("configuration not loaded")
		}

		if cfg.Quiet {
			return nil
		}

		if viper.GetString("output") == "json" {
			return listJSON(cfg)
		}
		return listText(cfg)
	},
}

func listJSON(cfg *internalcfg.Config) error {
	data := map[string]map[string]string{
		"aliases": cfg.Aliases,
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}

func listText(cfg *internalcfg.Config) error {
	if len(cfg.Aliases) == 0 {
		fmt.Println("No aliases configured.")
		return nil
	}

	fmt.Println("Aliases")
	fmt.Println(strings.Repeat("-", 50))

	// Sort aliases for deterministic output
	names := make([]string, 0, len(cfg.Aliases))
	for name := range cfg.Aliases {
		names = append(names, name)
	}
	slices.SortFunc(names, cmp.Compare)

	// Find max name length for alignment
	maxLen := 0
	for _, name := range names {
		if len(name) > maxLen {
			maxLen = len(name)
		}
	}

	for _, name := range names {
		fmt.Printf("%-*s  -> %s\n", maxLen, name, cfg.Aliases[name])
	}

	return nil
}
