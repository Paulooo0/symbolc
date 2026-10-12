package cli

import (
	"fmt"
	"path/filepath"

	"github.com/Paulooo0/symbolc/internal/analyzer"
	"github.com/Paulooo0/symbolc/internal/config"
	"github.com/Paulooo0/symbolc/internal/output"
	"github.com/spf13/cobra"
)

var outputPath string
var outputFormat string

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Synchronize symbol context with source code",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(configFile)
		if err != nil {
			return err
		}

		g, err := analyzer.NewAnalyzer(*cfg).Analyze(cmd.Context())
		if err != nil {
			return fmt.Errorf("sync: analyze: %w", err)
		}

		if outputFormat != "" {
			cfg.Output.Format = outputFormat
		}
		if cfg.Output.Format != config.DefaultOutputFormat {
			return fmt.Errorf("sync: unsupported output format: %s", cfg.Output.Format)
		}

		data, err := output.MarshalYAML(g, *cfg)
		if err != nil {
			return fmt.Errorf("sync: marshal output: %w", err)
		}

		path := cfg.Output.Path
		if outputPath != "" {
			path = outputPath
		}
		if path == "" || path == "-" {
			if _, err := cmd.OutOrStdout().Write(data); err != nil {
				return fmt.Errorf("sync: write output: %w", err)
			}
			return nil
		}
		if cfg.BaseDir != "" && !filepath.IsAbs(path) {
			path = filepath.Join(cfg.BaseDir, path)
		}
		if err := output.WriteFileAtomic(path, data); err != nil {
			return fmt.Errorf("sync: write output file: %w", err)
		}

		return nil
	},
}
