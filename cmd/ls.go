package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/meigma/blob-cli/internal/archive"
	internalcfg "github.com/meigma/blob-cli/internal/config"
)

var lsCmd = &cobra.Command{
	Use:   "ls <ref> [path]",
	Short: "List files and directories in an archive",
	Long: `List files and directories in an archive.

Lists the contents of an archive at the specified path. If no path
is provided, lists the root directory.`,
	Example: `  blob ls ghcr.io/acme/configs:v1.0.0
  blob ls -lh ghcr.io/acme/configs:v1.0.0 /etc
  blob ls --digest ghcr.io/acme/configs:v1.0.0`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runLs,
}

func init() {
	// Pre-define help without -h shorthand so we can use -h for human
	lsCmd.Flags().Bool("help", false, "help for ls")
	lsCmd.Flags().BoolP("human", "h", false, "human-readable sizes (use with -l)")
	lsCmd.Flags().BoolP("long", "l", false, "long format (permissions, size, hash)")
	lsCmd.Flags().Bool("digest", false, "show file digests")
}

// lsFlags holds the parsed command flags.
type lsFlags struct {
	long   bool
	human  bool
	digest bool
}

// lsResult contains the ls output data for JSON format.
type lsResult struct {
	Ref     string        `json:"ref"`
	Path    string        `json:"path"`
	Entries []lsEntryJSON `json:"entries"`
}

// lsEntryJSON represents a single entry in JSON output.
type lsEntryJSON struct {
	Name      string `json:"name"`
	Path      string `json:"path"`
	IsDir     bool   `json:"is_dir"`
	Mode      string `json:"mode,omitempty"`
	Size      uint64 `json:"size,omitempty"`
	SizeHuman string `json:"size_human,omitempty"`
	Digest    string `json:"digest,omitempty"`
	ModTime   string `json:"mod_time,omitempty"`
}

func runLs(cmd *cobra.Command, args []string) error {
	cfg := internalcfg.FromContext(cmd.Context())
	if cfg == nil {
		return errors.New("configuration not loaded")
	}

	ref := cfg.ResolveAlias(args[0])
	dirPath := "/"
	if len(args) > 1 {
		dirPath = args[1]
	}

	flags, err := parseLsFlags(cmd)
	if err != nil {
		return err
	}

	result, err := archive.Inspect(cmd.Context(), ref)
	if err != nil {
		return err
	}

	entries, err := archive.ListDir(result.Index(), dirPath)
	if err != nil {
		return err
	}

	if cfg.Quiet {
		return nil
	}

	if viper.GetString("output") == internalcfg.OutputJSON {
		return lsJSON(ref, dirPath, entries, flags)
	}
	return lsText(entries, flags)
}

func parseLsFlags(cmd *cobra.Command) (lsFlags, error) {
	var flags lsFlags
	var err error

	flags.long, err = cmd.Flags().GetBool("long")
	if err != nil {
		return flags, fmt.Errorf("reading long flag: %w", err)
	}

	flags.human, err = cmd.Flags().GetBool("human")
	if err != nil {
		return flags, fmt.Errorf("reading human flag: %w", err)
	}

	flags.digest, err = cmd.Flags().GetBool("digest")
	if err != nil {
		return flags, fmt.Errorf("reading digest flag: %w", err)
	}

	return flags, nil
}

func lsJSON(ref, dirPath string, entries []*archive.DirEntry, flags lsFlags) error {
	result := lsResult{
		Ref:     ref,
		Path:    dirPath,
		Entries: make([]lsEntryJSON, 0, len(entries)),
	}

	for _, entry := range entries {
		jsonEntry := lsEntryJSON{
			Name:  entry.Name,
			Path:  entry.Path,
			IsDir: entry.IsDir,
		}

		if flags.long {
			jsonEntry.Mode = archive.FormatMode(entry.Mode, entry.IsDir)
			if !entry.IsDir {
				jsonEntry.Size = entry.Size
				jsonEntry.ModTime = entry.ModTime.Format(time.RFC3339)
				if flags.human {
					jsonEntry.SizeHuman = archive.FormatSize(entry.Size)
				}
			}
		}

		if flags.digest && !entry.IsDir && len(entry.Hash) > 0 {
			jsonEntry.Digest = archive.FormatDigest(entry.Hash)
		}

		result.Entries = append(result.Entries, jsonEntry)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}

func lsText(entries []*archive.DirEntry, flags lsFlags) error {
	if len(entries) == 0 {
		return nil
	}

	maxSizeWidth := calculateMaxSizeWidth(entries, flags)

	for _, entry := range entries {
		printLsEntry(entry, flags, maxSizeWidth)
	}

	return nil
}

func calculateMaxSizeWidth(entries []*archive.DirEntry, flags lsFlags) int {
	if !flags.long {
		return 0
	}

	var maxWidth int
	for _, entry := range entries {
		sizeStr := formatEntrySize(entry.Size, flags.human)
		if len(sizeStr) > maxWidth {
			maxWidth = len(sizeStr)
		}
	}
	return maxWidth
}

func formatEntrySize(size uint64, human bool) string {
	if human {
		return archive.FormatSize(size)
	}
	return strconv.FormatUint(size, 10)
}

func printLsEntry(entry *archive.DirEntry, flags lsFlags, maxSizeWidth int) {
	name := entry.Name
	if entry.IsDir {
		name += "/"
	}

	switch {
	case flags.long && flags.digest:
		printLongWithDigest(entry, name, maxSizeWidth, flags.human)
	case flags.long:
		printLong(entry, name, maxSizeWidth, flags.human)
	case flags.digest:
		printDigestOnly(entry, name)
	default:
		fmt.Println(name)
	}
}

func printLongWithDigest(entry *archive.DirEntry, name string, maxSizeWidth int, human bool) {
	mode := archive.FormatMode(entry.Mode, entry.IsDir)
	sizeStr := formatEntrySize(entry.Size, human)
	digest := formatEntryDigest(entry)
	fmt.Printf("%s  %*s  %-20s  %s\n", mode, maxSizeWidth, sizeStr, digest, name)
}

func printLong(entry *archive.DirEntry, name string, maxSizeWidth int, human bool) {
	mode := archive.FormatMode(entry.Mode, entry.IsDir)
	sizeStr := formatEntrySize(entry.Size, human)
	fmt.Printf("%s  %*s  %s\n", mode, maxSizeWidth, sizeStr, name)
}

func printDigestOnly(entry *archive.DirEntry, name string) {
	digest := formatEntryDigest(entry)
	fmt.Printf("%-20s  %s\n", digest, name)
}

func formatEntryDigest(entry *archive.DirEntry) string {
	if entry.IsDir || len(entry.Hash) == 0 {
		return ""
	}
	return archive.FormatDigest(entry.Hash)
}
