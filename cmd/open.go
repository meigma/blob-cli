package cmd

import (
	"context"
	"errors"
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/meigma/blob"
	blobcore "github.com/meigma/blob/core"
	"github.com/spf13/cobra"

	internalcfg "github.com/meigma/blob-cli/internal/config"
	"github.com/meigma/blob-cli/internal/tui/open"
)

var openCmd = &cobra.Command{
	Use:   "open <ref>",
	Short: "Open an interactive file browser for a blob archive",
	Long: `Open an interactive TUI to explore blob archive contents.

Features a split-view layout with file tree on the left and content
preview on the right. Files load on-demand via HTTP range requests
for fast navigation.

Navigation:
  Arrow keys    Navigate file list / scroll preview
  Tab           Switch focus between tree and preview
  Enter/Right   Enter directory or preview file
  Left          Go to parent directory
  c             Copy selected file (prompts for path)
  q/Esc         Quit`,
	Example: `  blob open ghcr.io/acme/configs:v1.0.0
  blob open myalias`,
	Args: cobra.ExactArgs(1),
	RunE: runOpen,
}

func init() {
	rootCmd.AddCommand(openCmd)
}

func runOpen(cmd *cobra.Command, args []string) error {
	// 1. Get config from context
	cfg := internalcfg.FromContext(cmd.Context())
	if cfg == nil {
		return errors.New("configuration not loaded")
	}

	// 2. Parse arguments
	inputRef := args[0]

	// 3. Resolve alias
	resolvedRef := cfg.ResolveAlias(inputRef)

	// 4. Create client
	client, err := newClient(cfg)
	if err != nil {
		return fmt.Errorf("creating client: %w", err)
	}

	// 5. Create loader function for async archive loading
	ctx := cmd.Context()
	loader := makeArchiveLoader(ctx, client, resolvedRef)

	// 6. Create and run the TUI (starts with loading screen)
	model := open.New(resolvedRef, loader)
	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)

	if _, err := p.Run(); err != nil {
		return fmt.Errorf("running TUI: %w", err)
	}

	return nil
}

// makeArchiveLoader creates a LoadFunc that fetches the archive from the registry.
func makeArchiveLoader(ctx context.Context, client *blob.Client, ref string) open.LoadFunc {
	return func() (*blob.IndexView, *blob.Archive, error) {
		// Pull archive (lazy - does NOT download data blob)
		archive, err := client.Pull(ctx, ref)
		if err != nil {
			return nil, nil, fmt.Errorf("accessing archive %s: %w", ref, err)
		}

		// Create index view from the archive's index data
		index, err := blobcore.NewIndexView(archive.IndexData())
		if err != nil {
			return nil, nil, fmt.Errorf("parsing index: %w", err)
		}

		return index, archive, nil
	}
}
