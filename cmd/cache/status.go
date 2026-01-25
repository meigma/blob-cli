package cache

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/meigma/blob-cli/internal/archive"
	internalcfg "github.com/meigma/blob-cli/internal/config"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show cache sizes for all cache types",
	Long: `Show cache sizes for all cache types.

Displays the size and entry count for each cache type, as well
as the total cache size.`,
	Example: `  blob cache status
  blob cache status --output json`,
	Args: cobra.NoArgs,
	RunE: runStatus,
}

// cacheStats holds statistics for a single cache type.
type cacheStats struct {
	Name      string `json:"name"`
	Path      string `json:"path"`
	Enabled   bool   `json:"enabled"`
	Size      int64  `json:"size"`
	SizeHuman string `json:"size_human"`
	Files     int    `json:"files"`
}

// statusResult contains the status output data.
type statusResult struct {
	Root       string       `json:"root"`
	Caches     []cacheStats `json:"caches"`
	TotalSize  int64        `json:"total_size"`
	TotalHuman string       `json:"total_size_human"`
	TotalFiles int          `json:"total_files"`
}

func runStatus(cmd *cobra.Command, _ []string) error {
	cfg := internalcfg.FromContext(cmd.Context())
	if cfg == nil {
		return errors.New("configuration not loaded")
	}

	cacheDir, err := resolveCacheDir(cfg)
	if err != nil {
		return fmt.Errorf("determining cache directory: %w", err)
	}

	result := statusResult{
		Root:   cacheDir,
		Caches: make([]cacheStats, 0, len(cacheTypes)),
	}

	for _, ct := range cacheTypes {
		path := filepath.Join(cacheDir, ct.SubDir)
		enabled := isCacheTypeEnabled(cfg, ct.Name)
		size := getDirSize(path)
		files := countFiles(path)

		result.Caches = append(result.Caches, cacheStats{
			Name:      ct.Name,
			Path:      path,
			Enabled:   enabled,
			Size:      size,
			SizeHuman: archive.FormatSize(uint64(max(0, size))), //nolint:gosec // size is always non-negative
			Files:     files,
		})
		result.TotalSize += size
		result.TotalFiles += files
	}
	result.TotalHuman = archive.FormatSize(uint64(max(0, result.TotalSize))) //nolint:gosec // size is always non-negative

	if cfg.Quiet {
		return nil
	}

	if viper.GetString("output") == internalcfg.OutputJSON {
		return statusJSON(&result)
	}
	return statusText(&result)
}

func statusJSON(result *statusResult) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}

func statusText(result *statusResult) error {
	fmt.Printf("Cache directory: %s\n", result.Root)
	fmt.Println()

	// Calculate max width for alignment
	maxNameLen := 0
	for _, c := range result.Caches {
		if len(c.Name) > maxNameLen {
			maxNameLen = len(c.Name)
		}
	}

	for _, c := range result.Caches {
		status := ""
		if !c.Enabled {
			status = " (disabled)"
		}
		fmt.Printf("  %-*s  %8s  %5d files%s\n", maxNameLen, c.Name, c.SizeHuman, c.Files, status)
	}

	fmt.Println()
	fmt.Printf("Total: %s (%d files)\n", result.TotalHuman, result.TotalFiles)

	return nil
}
