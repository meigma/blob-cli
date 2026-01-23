package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/meigma/blob"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/meigma/blob-cli/internal/archive"
	internalcfg "github.com/meigma/blob-cli/internal/config"
)

var cpCmd = &cobra.Command{
	Use:   "cp <ref>:<path>... <dest>",
	Short: "Copy files or directories from an archive to the local filesystem",
	Long: `Copy files or directories from an archive to the local filesystem.

Uses HTTP range requests to fetch only the requested files without
downloading the entire archive. Multiple source paths can be specified.

Behavior:
  - Single file to file:      blob cp reg/repo:v1:/config.json ./config.json
  - Single file to dir:       blob cp reg/repo:v1:/config.json ./output/
  - Multiple files to dir:    blob cp reg/repo:v1:/a.json reg/repo:v1:/b.json ./output/
  - Directory to directory:   blob cp reg/repo:v1:/etc/nginx ./nginx-config`,
	Example: `  blob cp ghcr.io/acme/configs:v1.0.0:/config.json ./config.json
  blob cp ghcr.io/acme/configs:v1.0.0:/etc/nginx/ ./nginx/
  blob cp ghcr.io/acme/configs:v1.0.0:/a.json ghcr.io/acme/configs:v1.0.0:/b.json ./`,
	Args: cobra.MinimumNArgs(2),
	RunE: runCp,
}

func init() {
	cpCmd.Flags().BoolP("recursive", "r", true, "copy directories recursively")
	cpCmd.Flags().Bool("preserve", false, "preserve file permissions and timestamps from archive")
	cpCmd.Flags().BoolP("force", "f", false, "overwrite existing files")
}

// cpFlags holds the parsed command flags.
type cpFlags struct {
	recursive bool
	preserve  bool
	force     bool
}

// cpSource represents a parsed source argument (ref:/path).
type cpSource struct {
	inputRef string // Original input reference (for display)
	ref      string // Resolved reference
	path     string // Path within archive (with leading /)
}

// cpResolvedSource represents a source with its archive and detected type.
type cpResolvedSource struct {
	cpSource
	archive *blob.Archive
	isDir   bool
}

// cpResult contains the result of a copy operation.
type cpResult struct {
	Sources     []cpSourceResult `json:"sources"`
	Destination string           `json:"destination"`
	FileCount   int              `json:"file_count"`
	TotalSize   uint64           `json:"total_size"`
	SizeHuman   string           `json:"size_human,omitempty"`
}

// cpSourceResult represents a single source in the result.
type cpSourceResult struct {
	Ref  string `json:"ref"`
	Path string `json:"path"`
}

func runCp(cmd *cobra.Command, args []string) error {
	// 1. Get config from context
	cfg := internalcfg.FromContext(cmd.Context())
	if cfg == nil {
		return errors.New("configuration not loaded")
	}

	// 2. Parse flags
	flags, err := parseCpFlags(cmd)
	if err != nil {
		return err
	}

	// 3. Parse source arguments (all but last)
	sourceArgs := args[:len(args)-1]
	dest := args[len(args)-1]

	sources, err := parseSourceArgs(sourceArgs, cfg)
	if err != nil {
		return err
	}

	// 4. Pull archives and resolve source types
	ctx := cmd.Context()
	archiveCache := make(map[string]*blob.Archive)
	resolvedSources := make([]cpResolvedSource, 0, len(sources))

	for _, src := range sources {
		rsrc, resolveErr := resolveSource(ctx, src, archiveCache)
		if resolveErr != nil {
			return resolveErr
		}
		resolvedSources = append(resolvedSources, rsrc)
	}

	// 5. Validate destination and determine overall copy mode
	destPath, err := validateAndPrepareDestination(resolvedSources, dest, flags)
	if err != nil {
		return err
	}

	// 6. Execute copy operations
	result := &cpResult{
		Sources:     make([]cpSourceResult, 0, len(sources)),
		Destination: destPath,
	}

	copyOpts := buildCopyOpts(flags)

	for _, rsrc := range resolvedSources {
		count, size, err := copyResolvedSource(rsrc, destPath, flags, copyOpts, len(resolvedSources) > 1)
		if err != nil {
			return err
		}
		result.FileCount += count
		result.TotalSize += size
		result.Sources = append(result.Sources, cpSourceResult{
			Ref:  rsrc.inputRef,
			Path: rsrc.path,
		})
	}

	result.SizeHuman = archive.FormatSize(result.TotalSize)

	// 7. Output result
	return outputCpResult(cfg, result)
}

// resolveSource pulls the archive (if not cached) and detects if the source is a file or directory.
func resolveSource(ctx context.Context, src cpSource, cache map[string]*blob.Archive) (cpResolvedSource, error) {
	// Get or create archive for this ref
	blobArchive, ok := cache[src.ref]
	if !ok {
		client, clientErr := blob.NewClient(blob.WithDockerConfig())
		if clientErr != nil {
			return cpResolvedSource{}, fmt.Errorf("creating client: %w", clientErr)
		}
		var pullErr error
		blobArchive, pullErr = client.Pull(ctx, src.ref)
		if pullErr != nil {
			return cpResolvedSource{}, fmt.Errorf("accessing archive %s: %w", src.ref, pullErr)
		}
		cache[src.ref] = blobArchive
	}

	// Detect if source is a file or directory
	srcPath := blob.NormalizePath(src.path)
	if !blobArchive.Exists(srcPath) {
		return cpResolvedSource{}, fmt.Errorf("path not found in archive: %s", src.path)
	}
	isDir := blobArchive.IsDir(srcPath)

	return cpResolvedSource{
		cpSource: src,
		archive:  blobArchive,
		isDir:    isDir,
	}, nil
}

// destInfo holds information about the destination path.
type destInfo struct {
	absPath       string
	exists        bool
	isDir         bool
	endsWithSlash bool
}

// getDestInfo gathers information about the destination path.
func getDestInfo(dest string) (destInfo, error) {
	absPath, err := filepath.Abs(dest)
	if err != nil {
		return destInfo{}, fmt.Errorf("resolving destination path: %w", err)
	}

	info, statErr := os.Stat(absPath)
	exists := statErr == nil
	isDir := exists && info.IsDir()
	endsSlash := strings.HasSuffix(dest, "/") || strings.HasSuffix(dest, string(os.PathSeparator))

	return destInfo{
		absPath:       absPath,
		exists:        exists,
		isDir:         isDir,
		endsWithSlash: endsSlash,
	}, nil
}

// ensureDir creates the directory if it doesn't exist.
func ensureDir(path string) error {
	if err := os.MkdirAll(path, 0o750); err != nil {
		return fmt.Errorf("creating directory: %w", err)
	}
	return nil
}

// validateAndPrepareDestination validates the destination against sources and prepares it.
func validateAndPrepareDestination(sources []cpResolvedSource, dest string, flags cpFlags) (string, error) {
	di, err := getDestInfo(dest)
	if err != nil {
		return "", err
	}

	// Check for directory sources without -r flag
	for _, src := range sources {
		if src.isDir && !flags.recursive {
			return "", fmt.Errorf("cannot copy directory %s without -r flag", src.path)
		}
	}

	// Multiple sources always require a directory destination
	if len(sources) > 1 {
		return prepareMultiSourceDest(di)
	}

	// Single source
	return prepareSingleSourceDest(sources[0], di)
}

// prepareMultiSourceDest prepares destination for multiple sources.
func prepareMultiSourceDest(di destInfo) (string, error) {
	if di.exists && !di.isDir {
		return "", errors.New("destination must be a directory when copying multiple sources")
	}
	if !di.exists {
		if err := ensureDir(di.absPath); err != nil {
			return "", err
		}
	}
	return di.absPath, nil
}

// prepareSingleSourceDest prepares destination for a single source.
func prepareSingleSourceDest(src cpResolvedSource, di destInfo) (string, error) {
	// Directory source requires directory destination
	if src.isDir {
		if di.exists && !di.isDir {
			return "", fmt.Errorf("cannot copy directory %s to file %s", src.path, di.absPath)
		}
		if !di.exists {
			if err := ensureDir(di.absPath); err != nil {
				return "", err
			}
		}
		return di.absPath, nil
	}

	// File source to directory
	if di.isDir || di.endsWithSlash {
		if !di.exists {
			if err := ensureDir(di.absPath); err != nil {
				return "", err
			}
		}
		return di.absPath, nil
	}

	// File to file copy - ensure parent directory exists
	parentDir := filepath.Dir(di.absPath)
	if err := ensureDir(parentDir); err != nil {
		return "", fmt.Errorf("creating parent directory: %w", err)
	}

	return di.absPath, nil
}

// copyResolvedSource copies a resolved source to the destination.
func copyResolvedSource(rsrc cpResolvedSource, destPath string, flags cpFlags, opts []blob.CopyOption, multiSource bool) (fileCount int, totalSize uint64, err error) {
	srcPath := blob.NormalizePath(rsrc.path)

	if rsrc.isDir {
		return copyDirectory(rsrc.archive, srcPath, rsrc.path, destPath, opts)
	}

	// File copy - determine if copying to directory or specific file
	destInfo, statErr := os.Stat(destPath)
	destIsDir := statErr == nil && destInfo.IsDir()

	if destIsDir || multiSource {
		return copyFileToDir(rsrc.archive, srcPath, rsrc.path, destPath, opts)
	}

	return copyFileToFile(rsrc.archive, srcPath, rsrc.path, destPath, flags)
}

// copyDirectory copies a directory recursively.
func copyDirectory(blobArchive *blob.Archive, srcPath, displayPath, destPath string, opts []blob.CopyOption) (fileCount int, totalSize uint64, err error) {
	normalizedPath := blob.NormalizePath(srcPath)
	stats, err := blobArchive.CopyDir(destPath, normalizedPath, opts...)
	if err != nil {
		return 0, 0, fmt.Errorf("copying directory %s: %w", displayPath, err)
	}
	return stats.FileCount, stats.TotalBytes, nil
}

// copyFileToDir copies a file into a directory.
func copyFileToDir(blobArchive *blob.Archive, srcPath, displayPath, destPath string, opts []blob.CopyOption) (fileCount int, totalSize uint64, err error) {
	// Verify source exists and is a file
	if !blobArchive.IsFile(srcPath) {
		if blobArchive.IsDir(srcPath) {
			return 0, 0, fmt.Errorf("expected file but got directory: %s", displayPath)
		}
		return 0, 0, fmt.Errorf("file not found: %s", displayPath)
	}

	stats, err := blobArchive.CopyToWithOptions(destPath, []string{srcPath}, opts...)
	if err != nil {
		return 0, 0, fmt.Errorf("copying %s: %w", displayPath, err)
	}

	return stats.FileCount, stats.TotalBytes, nil
}

// copyFileToFile copies a single file to a specific destination path.
// Uses manual implementation to control permissions (0644 default vs CopyFile's 0600).
func copyFileToFile(blobArchive *blob.Archive, srcPath, displayPath, destPath string, flags cpFlags) (fileCount int, totalSize uint64, err error) {
	entry, ok := blobArchive.Entry(srcPath)
	if !ok {
		if blobArchive.IsDir(srcPath) {
			return 0, 0, fmt.Errorf("expected file but got directory: %s", displayPath)
		}
		return 0, 0, fmt.Errorf("file not found: %s", displayPath)
	}
	if entry.Mode().IsDir() {
		return 0, 0, fmt.Errorf("expected file but got directory: %s", displayPath)
	}

	// Check if destination exists BEFORE reading file content (avoid unnecessary downloads)
	if _, statErr := os.Stat(destPath); statErr == nil {
		if !flags.force {
			// File exists and force not set - skip without error
			return 0, 0, nil
		}
	}

	// Now read the file content (triggers HTTP range request)
	content, err := blobArchive.ReadFile(srcPath)
	if err != nil {
		return 0, 0, fmt.Errorf("reading %s: %w", displayPath, err)
	}

	perm := os.FileMode(0o644)
	if flags.preserve {
		perm = entry.Mode()
	}
	if err := os.WriteFile(destPath, content, perm); err != nil {
		return 0, 0, fmt.Errorf("writing %s: %w", destPath, err)
	}

	// Preserve modification time if requested
	if flags.preserve {
		if err := os.Chtimes(destPath, entry.ModTime(), entry.ModTime()); err != nil {
			// Non-fatal error - log but continue
			_ = err
		}
	}

	return 1, entry.OriginalSize(), nil
}

// parseCpFlags extracts and validates flags from the command.
func parseCpFlags(cmd *cobra.Command) (cpFlags, error) {
	var flags cpFlags
	var err error

	flags.recursive, err = cmd.Flags().GetBool("recursive")
	if err != nil {
		return flags, fmt.Errorf("reading recursive flag: %w", err)
	}

	flags.preserve, err = cmd.Flags().GetBool("preserve")
	if err != nil {
		return flags, fmt.Errorf("reading preserve flag: %w", err)
	}

	flags.force, err = cmd.Flags().GetBool("force")
	if err != nil {
		return flags, fmt.Errorf("reading force flag: %w", err)
	}

	return flags, nil
}

// parseSourceArg parses a single source argument in "ref:/path" format.
func parseSourceArg(arg string, cfg *internalcfg.Config) (cpSource, error) {
	// Find ":/" which separates ref from archive path
	// Archive paths always start with "/"
	idx := strings.Index(arg, ":/")
	if idx == -1 {
		return cpSource{}, fmt.Errorf("invalid source format %q: expected <ref>:<path> (path must start with /)", arg)
	}

	inputRef := arg[:idx]
	archivePath := arg[idx+1:] // Include the leading /

	if inputRef == "" {
		return cpSource{}, fmt.Errorf("invalid source format %q: reference cannot be empty", arg)
	}

	resolvedRef := cfg.ResolveAlias(inputRef)

	return cpSource{
		inputRef: inputRef,
		ref:      resolvedRef,
		path:     archivePath,
	}, nil
}

// parseSourceArgs parses all source arguments.
func parseSourceArgs(args []string, cfg *internalcfg.Config) ([]cpSource, error) {
	sources := make([]cpSource, 0, len(args))
	for _, arg := range args {
		src, err := parseSourceArg(arg, cfg)
		if err != nil {
			return nil, err
		}
		sources = append(sources, src)
	}
	return sources, nil
}

// buildCopyOpts creates copy options based on flags.
func buildCopyOpts(flags cpFlags) []blob.CopyOption {
	opts := []blob.CopyOption{blob.CopyWithOverwrite(flags.force)}
	if flags.preserve {
		opts = append(opts, blob.CopyWithPreserveMode(true), blob.CopyWithPreserveTimes(true))
	}
	return opts
}

// outputCpResult formats and outputs the copy result.
func outputCpResult(cfg *internalcfg.Config, result *cpResult) error {
	if cfg.Quiet {
		return nil
	}
	if viper.GetString("output") == internalcfg.OutputJSON {
		return cpJSON(result)
	}
	return cpText(result)
}

func cpJSON(result *cpResult) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}

func cpText(result *cpResult) error {
	fmt.Printf("Copied %d file(s) (%s)\n", result.FileCount, result.SizeHuman)
	for _, src := range result.Sources {
		fmt.Printf("  %s:%s\n", src.Ref, src.Path)
	}
	fmt.Printf("  → %s\n", result.Destination)
	return nil
}
