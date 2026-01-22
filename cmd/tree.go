package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/meigma/blob-cli/internal/archive"
	internalcfg "github.com/meigma/blob-cli/internal/config"
)

var treeCmd = &cobra.Command{
	Use:   "tree <ref> [path]",
	Short: "Display directory structure as a tree",
	Long: `Display directory structure as a tree.

Shows the hierarchical structure of files and directories in an
archive, similar to the tree command.`,
	Example: `  blob tree ghcr.io/acme/configs:v1.0.0
  blob tree -L 2 ghcr.io/acme/configs:v1.0.0 /etc`,
	Args: cobra.RangeArgs(1, 2),
	RunE: runTree,
}

func init() {
	treeCmd.Flags().IntP("level", "L", 0, "descend only n levels deep (0 = unlimited)")
	treeCmd.Flags().Bool("dirsfirst", false, "list directories before files")
}

// treeFlags holds the parsed command flags.
type treeFlags struct {
	level     int
	dirsFirst bool
}

// treeResult contains the tree output data for JSON format.
type treeResult struct {
	Ref       string    `json:"ref"`
	Path      string    `json:"path"`
	Root      *treeNode `json:"root"`
	DirCount  int       `json:"directory_count"`
	FileCount int       `json:"file_count"`
}

// treeNode represents a single node in the JSON tree.
type treeNode struct {
	Name     string      `json:"name"`
	Path     string      `json:"path"`
	IsDir    bool        `json:"is_dir"`
	Children []*treeNode `json:"children,omitempty"`
}

func runTree(cmd *cobra.Command, args []string) error {
	cfg := internalcfg.FromContext(cmd.Context())
	if cfg == nil {
		return errors.New("configuration not loaded")
	}

	ref := cfg.ResolveAlias(args[0])
	dirPath := "/"
	if len(args) > 1 {
		dirPath = args[1]
	}

	flags, err := parseTreeFlags(cmd)
	if err != nil {
		return err
	}

	result, err := archive.Inspect(cmd.Context(), ref)
	if err != nil {
		return err
	}

	root, err := archive.BuildTree(result.Index(), dirPath, flags.level)
	if err != nil {
		return err
	}

	if cfg.Quiet {
		return nil
	}

	if viper.GetString("output") == internalcfg.OutputJSON {
		return treeJSON(ref, dirPath, root, flags)
	}
	return treeText(root, flags)
}

func parseTreeFlags(cmd *cobra.Command) (treeFlags, error) {
	var flags treeFlags
	var err error

	flags.level, err = cmd.Flags().GetInt("level")
	if err != nil {
		return flags, fmt.Errorf("reading level flag: %w", err)
	}

	flags.dirsFirst, err = cmd.Flags().GetBool("dirsfirst")
	if err != nil {
		return flags, fmt.Errorf("reading dirsfirst flag: %w", err)
	}

	return flags, nil
}

func treeJSON(ref, dirPath string, root *archive.DirEntry, flags treeFlags) error {
	dirs, files := archive.Counts(root)

	result := treeResult{
		Ref:       ref,
		Path:      dirPath,
		Root:      convertToTreeNode(root, flags.dirsFirst),
		DirCount:  dirs,
		FileCount: files,
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}

func convertToTreeNode(entry *archive.DirEntry, dirsFirst bool) *treeNode {
	node := &treeNode{
		Name:  entry.Name,
		Path:  entry.Path,
		IsDir: entry.IsDir,
	}

	if len(entry.Children) > 0 {
		children := make([]*archive.DirEntry, len(entry.Children))
		copy(children, entry.Children)

		if dirsFirst {
			archive.SortDirsFirst(children)
		}

		node.Children = make([]*treeNode, 0, len(children))
		for _, child := range children {
			node.Children = append(node.Children, convertToTreeNode(child, dirsFirst))
		}
	}

	return node
}

func treeText(root *archive.DirEntry, flags treeFlags) error {
	printer := &archive.TreePrinter{
		DirsFirst: flags.dirsFirst,
		Writer:    os.Stdout,
	}

	printer.Print(root)

	// Print summary line
	dirs, files := archive.Counts(root)
	fmt.Println()
	fmt.Printf("%s, %s\n", pluralize(dirs, "directory", "directories"), pluralize(files, "file", "files"))

	return nil
}

func pluralize(n int, singular, plural string) string {
	if n == 1 {
		return fmt.Sprintf("%d %s", n, singular)
	}
	return fmt.Sprintf("%d %s", n, plural)
}
