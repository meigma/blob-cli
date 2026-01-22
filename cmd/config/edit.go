package config

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/spf13/cobra"

	internalcfg "github.com/meigma/blob-cli/internal/config"
)

var editCmd = &cobra.Command{
	Use:   "edit",
	Short: "Open configuration file in $EDITOR",
	Long: `Open configuration file in $EDITOR.

Opens the configuration file in your default editor. Uses $EDITOR,
falling back to $VISUAL, then vi (or notepad on Windows).

Creates the config file with defaults if it doesn't exist.`,
	Example: `  blob config edit`,
	Args:    cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := internalcfg.ConfigPathUsed()
		if err != nil {
			return err
		}

		// Create default config file if it doesn't exist
		if _, err := os.Stat(path); os.IsNotExist(err) {
			if err := internalcfg.SaveDefaultWithComments(path); err != nil {
				return fmt.Errorf("creating config file: %w", err)
			}
		}

		editorCmd, editorArgs := parseEditor(getEditor())
		allArgs := append(editorArgs, path)

		c := exec.Command(editorCmd, allArgs...) //nolint:gosec // editor is user-controlled via $EDITOR
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr

		return c.Run()
	},
}

// getEditor returns the user's preferred editor.
func getEditor() string {
	if editor := os.Getenv("EDITOR"); editor != "" {
		return editor
	}
	if editor := os.Getenv("VISUAL"); editor != "" {
		return editor
	}
	if runtime.GOOS == "windows" {
		return "notepad"
	}
	return "vi"
}

// parseEditor splits an editor string into command and arguments.
// Handles common cases like "code -w" or "vim -u NONE".
func parseEditor(editor string) (cmd string, args []string) {
	parts := strings.Fields(editor)
	if len(parts) == 0 {
		return "vi", nil
	}
	return parts[0], parts[1:]
}
