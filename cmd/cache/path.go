package cache

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	internalcfg "github.com/meigma/blob-cli/internal/config"
)

var pathCmd = &cobra.Command{
	Use:   "path",
	Short: "Show cache directory paths",
	Long: `Show cache directory paths.

Displays the paths for each cache type. Paths follow the XDG
Base Directory Specification.`,
	Example: `  blob cache path
  blob cache path --output json`,
	Args: cobra.NoArgs,
	RunE: runPath,
}

// pathResult contains the path output data.
type pathResult struct {
	Root  string            `json:"root"`
	Paths map[string]string `json:"paths"`
}

func runPath(cmd *cobra.Command, _ []string) error {
	cfg := internalcfg.FromContext(cmd.Context())
	if cfg == nil {
		return errors.New("configuration not loaded")
	}

	cacheDir, err := resolveCacheDir(cfg)
	if err != nil {
		return fmt.Errorf("determining cache directory: %w", err)
	}

	result := pathResult{
		Root:  cacheDir,
		Paths: make(map[string]string, len(cacheTypes)),
	}
	for _, ct := range cacheTypes {
		result.Paths[ct.Name] = filepath.Join(cacheDir, ct.SubDir)
	}

	if cfg.Quiet {
		return nil
	}

	if viper.GetString("output") == internalcfg.OutputJSON {
		return pathJSON(&result)
	}
	return pathText(&result)
}

func pathJSON(result *pathResult) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}

func pathText(result *pathResult) error {
	fmt.Printf("Cache directory: %s\n", result.Root)
	fmt.Println()
	for _, ct := range cacheTypes {
		fmt.Printf("  %-12s %s\n", ct.Name+":", result.Paths[ct.Name])
	}
	return nil
}
