package main

import (
	"os"

	"github.com/meigma/blob-cli/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
