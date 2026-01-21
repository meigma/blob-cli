package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var pushCmd = &cobra.Command{
	Use:   "push <ref> <path>",
	Short: "Push a directory to an OCI registry as a blob archive",
	Long: `Push a directory to an OCI registry as a blob archive.

The directory contents are archived and uploaded to the specified
registry reference. Files are compressed individually using zstd
by default for optimal random access performance.`,
	Example: `  blob push ghcr.io/acme/configs:v1.0.0 ./config
  blob push --sign ghcr.io/acme/configs:latest ./config
  blob push --compression none ghcr.io/acme/data:v1 ./data`,
	Args: cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		return nil
	},
}

func init() {
	pushCmd.Flags().StringP("compression", "c", "zstd", "compression type: none, zstd")
	pushCmd.Flags().Bool("skip-compressed", true, "skip compressing already-compressed files")
	pushCmd.Flags().Bool("sign", false, "sign the archive after pushing")
	pushCmd.Flags().StringArray("annotation", nil, "add annotation to manifest (k=v, repeatable)")

	viper.BindPFlag("compression", pushCmd.Flags().Lookup("compression"))
}
