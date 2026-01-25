package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/meigma/blob"
	"github.com/spf13/cobra"

	internalcfg "github.com/meigma/blob-cli/internal/config"
)

var catCmd = &cobra.Command{
	Use:   "cat <ref> <file>...",
	Short: "Print file contents to stdout",
	Long: `Print file contents to stdout.

Useful for viewing, piping, or combining files from an archive.
Uses HTTP range requests to fetch only the requested files without
downloading the entire archive.`,
	Example: `  blob cat ghcr.io/acme/configs:v1.0.0 config.json
  blob cat ghcr.io/acme/configs:v1.0.0 config.json | jq .
  blob cat ghcr.io/acme/configs:v1.0.0 header.txt body.txt footer.txt > combined.txt`,
	Args: cobra.MinimumNArgs(2),
	RunE: runCat,
}

func init() {
	catCmd.Flags().Bool("skip-cache", false, "bypass registry caches for this operation")
}

func runCat(cmd *cobra.Command, args []string) error {
	// 1. Get config from context
	cfg := internalcfg.FromContext(cmd.Context())
	if cfg == nil {
		return errors.New("configuration not loaded")
	}

	// 2. Parse arguments
	inputRef := args[0]
	filePaths := args[1:]

	// 3. Parse flags
	skipCache, flagErr := cmd.Flags().GetBool("skip-cache")
	if flagErr != nil {
		return fmt.Errorf("reading skip-cache flag: %w", flagErr)
	}

	// 4. Resolve alias
	resolvedRef := cfg.ResolveAlias(inputRef)

	// 5. Create client (lazy - only downloads manifest + index)
	var client *blob.Client
	var err error
	if skipCache {
		client, err = blob.NewClient(clientOptsNoCache(cfg)...)
	} else {
		client, err = newClient(cfg)
	}
	if err != nil {
		return fmt.Errorf("creating client: %w", err)
	}

	// 6. Pull archive (lazy - does NOT download data blob)
	ctx := cmd.Context()
	var pullOpts []blob.PullOption
	if skipCache {
		pullOpts = append(pullOpts, blob.PullWithSkipCache())
	}
	blobArchive, err := client.Pull(ctx, resolvedRef, pullOpts...)
	if err != nil {
		return fmt.Errorf("accessing archive %s: %w", resolvedRef, err)
	}

	// 7. Validate all files exist and are not directories before outputting anything
	normalizedPaths, err := blobArchive.ValidateFiles(filePaths...)
	if err != nil {
		var ve *blob.ValidationError
		if errors.As(err, &ve) {
			switch ve.Reason {
			case "is a directory":
				return fmt.Errorf("cannot cat directory: %s", ve.Path)
			case "not found":
				return fmt.Errorf("file not found: %s", ve.Path)
			default:
				return fmt.Errorf("invalid path: %s: %s", ve.Path, ve.Reason)
			}
		}
		return fmt.Errorf("validating files: %w", err)
	}

	// 8. Check quiet mode - suppress output only after validation
	if cfg.Quiet {
		return nil
	}

	// 9. Stream each file to stdout
	for _, normalizedPath := range normalizedPaths {
		if err := catFile(blobArchive, normalizedPath); err != nil {
			return err
		}
	}

	return nil
}

// catFile streams a single file from the archive to stdout.
// Each file read triggers an HTTP range request for just that file's bytes.
func catFile(archive *blob.Archive, filePath string) error {
	// Open the file (triggers HTTP range request)
	f, err := archive.Open(filePath)
	if err != nil {
		return fmt.Errorf("opening %s: %w", filePath, err)
	}
	defer f.Close()

	// Stream to stdout
	if _, err := io.Copy(os.Stdout, f); err != nil {
		return fmt.Errorf("reading %s: %w", filePath, err)
	}

	return nil
}
