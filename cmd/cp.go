package cmd

import (
	"github.com/spf13/cobra"
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
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func init() {
	cpCmd.Flags().BoolP("recursive", "r", true, "copy directories recursively")
	cpCmd.Flags().Bool("preserve", false, "preserve file permissions from archive")
}
