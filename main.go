package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/meigma/blob-cli/cmd"
)

func main() {
	os.Exit(run())
}

func run() int {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)

		// Check for specific exit codes
		var exitErr *cmd.ExitError
		if errors.As(err, &exitErr) {
			return exitErr.Code
		}

		return 1
	}
	return 0
}
