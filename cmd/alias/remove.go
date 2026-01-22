package alias

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	internalcfg "github.com/meigma/blob-cli/internal/config"
)

var removeCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove an alias",
	Long: `Remove an alias from the configuration file.

Deletes the specified alias. This action cannot be undone.`,
	Example: `  blob alias remove foo`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]

		cfg := internalcfg.FromContext(cmd.Context())
		if cfg == nil {
			return errors.New("configuration not loaded")
		}

		// Check if alias exists
		if _, exists := cfg.Aliases[name]; !exists {
			return fmt.Errorf("alias %q not found", name)
		}

		// Create new config with alias removed
		newCfg := cfg.RemoveAlias(name)

		// Get config path and save
		path, err := internalcfg.ConfigPathUsed()
		if err != nil {
			return fmt.Errorf("determining config path: %w", err)
		}

		if err := internalcfg.Save(newCfg, path); err != nil {
			return fmt.Errorf("saving config: %w", err)
		}

		// Output result (respects --quiet for all formats)
		if cfg.Quiet {
			return nil
		}
		if viper.GetString("output") == "json" {
			return removeJSON(name)
		}
		return removeText(name)
	},
}

func removeJSON(name string) error {
	data := map[string]string{
		"action": "removed",
		"name":   name,
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}

func removeText(name string) error {
	fmt.Printf("Removed alias %q\n", name)
	return nil
}
