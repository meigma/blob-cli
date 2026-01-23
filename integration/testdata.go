//go:build integration

package integration

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/meigma/blob-cli/cmd"
)

//go:embed testdata/sample-project
var sampleProject embed.FS

// run is the main entry point for the CLI, returning an exit code.
func run() int {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)

		var exitErr *cmd.ExitError
		if errors.As(err, &exitErr) {
			return exitErr.Code
		}

		return 1
	}
	return 0
}

// copyTestData copies the embedded sample-project to the given work directory.
func copyTestData(workDir string) error {
	return fs.WalkDir(sampleProject, "testdata/sample-project", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Calculate relative path from "testdata/sample-project"
		relPath, err := filepath.Rel("testdata/sample-project", path)
		if err != nil {
			return err
		}

		// Skip the root directory itself
		if relPath == "." {
			return nil
		}

		destPath := filepath.Join(workDir, "sample-project", relPath)

		if d.IsDir() {
			return os.MkdirAll(destPath, 0755)
		}

		data, err := sampleProject.ReadFile(path)
		if err != nil {
			return err
		}

		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			return err
		}

		return os.WriteFile(destPath, data, 0644)
	})
}
