package main

import (
	"os"

	"github.com/uptriva/cli/internal/cmd"
)

var version = "dev"

func main() {
	if err := cmd.NewRoot(version).Execute(); err != nil {
		os.Exit(1)
	}
}
