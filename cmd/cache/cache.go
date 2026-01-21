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
  - manifests:  OCI manifest cache
  - indexes:    Archive index cache`,
}

func init() {
	Cmd.AddCommand(statusCmd)
	Cmd.AddCommand(clearCmd)
	Cmd.AddCommand(pathCmd)
}
