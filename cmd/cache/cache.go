package cache

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "cache",
	Short: "Manage local caches",
	Long: `Manage local caches.

Blob maintains several caches to improve performance:
  - content:    File content cache (deduplicated across archives)
  - blocks:     HTTP range block cache
  - refs:       Tag to digest mappings
  - manifests:  OCI manifest cache
  - indexes:    Archive index cache

Cache location follows XDG Base Directory Specification:
$XDG_CACHE_HOME/blob or ~/.cache/blob by default.

Override with cache.dir in config file or BLOB_CACHE_DIR environment variable.`,
}

func init() {
	Cmd.AddCommand(statusCmd)
	Cmd.AddCommand(clearCmd)
	Cmd.AddCommand(pathCmd)
}
