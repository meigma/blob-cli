package cache

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/meigma/blob-cli/internal/archive"
	internalcfg "github.com/meigma/blob-cli/internal/config"
)

var clearCmd = &cobra.Command{
	Use:   "clear [type]",
	Short: "Clear caches",
	Long: `Clear caches. Clears all caches by default.

Cache types:
  content     File content cache (deduplicated across archives)
  blocks      HTTP range block cache
  refs        Tag to digest mappings
  manifests   OCI manifest cache
  indexes     Archive index cache
  all         All caches (default)`,
	Example: `  blob cache clear              # Clear all caches (prompts for confirmation)
  blob cache clear --force      # Clear all without prompting
  blob cache clear content      # Clear only content cache
  blob cache clear manifests    # Clear only manifest cache`,
	Args: cobra.MaximumNArgs(1),
	RunE: runClear,
}

func init() {
	clearCmd.Flags().Bool("force", false, "skip confirmation prompt")
}

// clearResult contains the clear output data.
type clearResult struct {
	Cleared    []string `json:"cleared"`
	TotalSize  int64    `json:"total_size_cleared"`
	TotalHuman string   `json:"total_size_human"`
	TotalFiles int      `json:"total_files_cleared"`
}

func runClear(cmd *cobra.Command, args []string) error {
	cfg := internalcfg.FromContext(cmd.Context())
	if cfg == nil {
		return errors.New("configuration not loaded")
	}

	targetType, typesToClear, err := parseClearArgs(args)
	if err != nil {
		return err
	}

	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		return fmt.Errorf("reading force flag: %w", err)
	}

	cacheDir, err := resolveCacheDir(cfg)
	if err != nil {
		return fmt.Errorf("determining cache directory: %w", err)
	}

	totalSize, totalFiles := calculateCacheSizes(cacheDir, typesToClear)

	// Require --force for non-interactive (JSON) output
	if viper.GetString("output") == internalcfg.OutputJSON && !force {
		return errors.New("--force required when using --output json")
	}

	if !force && !cfg.Quiet {
		confirmed, promptErr := promptClearConfirmation(targetType, totalSize, totalFiles)
		if promptErr != nil {
			return promptErr
		}
		if !confirmed {
			fmt.Fprintln(os.Stderr, "Canceled.")
			return nil
		}
	}

	result, err := executeClear(cacheDir, typesToClear, totalSize, totalFiles)
	if err != nil {
		return err
	}

	return outputClearResult(cfg, result)
}

// parseClearArgs parses and validates the cache type argument.
func parseClearArgs(args []string) (string, []cacheType, error) {
	targetType := cacheTypeAll
	if len(args) > 0 {
		targetType = args[0]
	}

	if !validCacheType(targetType) {
		return "", nil, fmt.Errorf("invalid cache type %q, valid types: %s", targetType, strings.Join(cacheTypeNames(), ", "))
	}

	var typesToClear []cacheType
	if targetType == cacheTypeAll {
		typesToClear = cacheTypes
	} else {
		for _, ct := range cacheTypes {
			if ct.Name == targetType {
				typesToClear = []cacheType{ct}
				break
			}
		}
	}

	return targetType, typesToClear, nil
}

// calculateCacheSizes calculates total size and file count for the given cache types.
func calculateCacheSizes(cacheDir string, types []cacheType) (totalSize int64, totalFiles int) {
	for _, ct := range types {
		path := filepath.Join(cacheDir, ct.SubDir)
		totalSize += getDirSize(path)
		totalFiles += countFiles(path)
	}
	return totalSize, totalFiles
}

// promptClearConfirmation prompts the user for confirmation.
// Returns false (not confirmed) on EOF or non-interactive stdin.
func promptClearConfirmation(targetType string, totalSize int64, totalFiles int) (bool, error) {
	typeDesc := targetType + " cache"
	if targetType == cacheTypeAll {
		typeDesc = "all caches"
	}

	fmt.Printf("Clear %s? (%s, %d files) [y/N]: ",
		typeDesc,
		archive.FormatSize(uint64(max(0, totalSize))), //nolint:gosec // size is always non-negative
		totalFiles)

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		// Treat EOF (non-interactive, piped stdin) as "no"
		if errors.Is(err, io.EOF) {
			fmt.Println() // newline since user didn't press enter
			return false, nil
		}
		return false, fmt.Errorf("reading response: %w", err)
	}
	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes", nil
}

// executeClear clears the specified cache types.
func executeClear(cacheDir string, types []cacheType, totalSize int64, totalFiles int) (*clearResult, error) {
	result := &clearResult{
		Cleared:    make([]string, 0, len(types)),
		TotalSize:  totalSize,
		TotalHuman: archive.FormatSize(uint64(max(0, totalSize))), //nolint:gosec // size is always non-negative
		TotalFiles: totalFiles,
	}

	for _, ct := range types {
		path := filepath.Join(cacheDir, ct.SubDir)
		if err := clearDirectory(path); err != nil {
			return nil, fmt.Errorf("clearing %s cache: %w", ct.Name, err)
		}
		result.Cleared = append(result.Cleared, ct.Name)
	}

	return result, nil
}

// outputClearResult outputs the clear result in the appropriate format.
func outputClearResult(cfg *internalcfg.Config, result *clearResult) error {
	if cfg.Quiet {
		return nil
	}
	if viper.GetString("output") == internalcfg.OutputJSON {
		return clearJSON(result)
	}
	return clearText(result)
}

// clearDirectory removes all contents of a directory but keeps the directory itself.
func clearDirectory(dir string) error {
	// Check if directory exists
	info, err := os.Stat(dir)
	if os.IsNotExist(err) {
		return nil // Nothing to clear
	}
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", dir)
	}

	// Read directory entries
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}

	// Remove each entry
	for _, entry := range entries {
		path := filepath.Join(dir, entry.Name())
		if err := os.RemoveAll(path); err != nil {
			return err
		}
	}

	return nil
}

func clearJSON(result *clearResult) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}

func clearText(result *clearResult) error {
	if len(result.Cleared) == 0 {
		fmt.Println("No caches to clear.")
		return nil
	}

	fmt.Printf("Cleared %s (%d files)\n", result.TotalHuman, result.TotalFiles)
	for _, name := range result.Cleared {
		fmt.Printf("  - %s\n", name)
	}
	return nil
}
