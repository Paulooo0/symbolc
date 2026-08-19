package main

import (
	"fmt"
	"os"

	"github.com/Paulooo0/symbolc/internal/config"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "symbolc",
	Short: "Generate and inspect source code symbol context",
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

var configFile string

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Synchronize symbol context with source code",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(configFile)
		if err != nil {
			return err
		}

		fmt.Printf(
			"configuration loaded: %d source root(s)\n",
			len(cfg.Source.Roots),
		)

		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(
		nil,
		"config",
		"symbolc.yml",
		"path to symbolc configuration",
	)

	rootCmd.AddCommand(syncCmd)
}
