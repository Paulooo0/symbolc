package cli

import (
	"github.com/Paulooo0/symbolc/internal/config"
	"github.com/spf13/cobra"
)

var configFile string

var RootCmd = &cobra.Command{
	Use:   "symbolc",
	Short: "Generate and inspect source code symbol context",
}

func init() {
	RootCmd.PersistentFlags().StringVarP(
		&configFile,
		"config",
		"c",
		config.DefaultFile,
		"path to symbolc configuration",
	)
	syncCmd.Flags().StringVarP(&outputPath, "output", "o", "", "output file path, or - for stdout")
	syncCmd.Flags().StringVarP(&outputFormat, "format", "f", "", "output format")

	RootCmd.AddCommand(syncCmd)
}
