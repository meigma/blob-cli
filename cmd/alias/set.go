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
		name := args[0]
		ref := args[1]

		cfg := internalcfg.FromContext(cmd.Context())
		if cfg == nil {
			return errors.New("configuration not loaded")
		}

		// Check if this is an update or new alias
		_, isUpdate := cfg.Aliases[name]

		// Create new config with alias set
		newCfg := cfg.SetAlias(name, ref)

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
			return setJSON(name, ref, isUpdate)
		}
		return setText(name, ref, isUpdate)
	},
}

func setJSON(name, ref string, isUpdate bool) error {
	action := "created"
	if isUpdate {
		action = "updated"
	}
	data := map[string]string{
		"action": action,
		"name":   name,
		"ref":    ref,
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(data)
}

func setText(name, ref string, isUpdate bool) error {
	if isUpdate {
		fmt.Printf("Updated alias %q -> %s\n", name, ref)
	} else {
		fmt.Printf("Created alias %q -> %s\n", name, ref)
	}
	return nil
}
