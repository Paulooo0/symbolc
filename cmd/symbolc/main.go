package main

import (
	"os"

	"github.com/Paulooo0/symbolc/internal/cli"
)

func main() {
	if err := cli.RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
