package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/Paulooo0/symbolc/pkg/ptr"
	"gopkg.in/yaml.v3"
)

const DefaultFile = "symbolc.yml"
const DefaultOutputFormat = "yaml"

type Config struct {
	Version   *int8           `yaml:"version"`
	Source    SourceConfig    `yaml:"source"`
	Contracts ContractsConfig `yaml:"contracts"`
	Output    OutputConfig    `yaml:"output"`
	BaseDir   string          `yaml:"-"`
}

type SourceConfig struct {
	Roots []string `yaml:"roots"`
}

type ContractsConfig struct {
	Include []string `yaml:"include"`
	Exclude []string `yaml:"exclude"`
}

type OutputConfig struct {
	Format string `yaml:"format"`
	Path   string `yaml:"path"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config: read file: %w", err)
	}

	cfg := &Config{}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("config: parse yaml data: %w", err)
	}

	cfg.BaseDir, err = filepath.Abs(filepath.Dir(path))
	if err != nil {
		return nil, fmt.Errorf("config: resolve base directory: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config: invalid config: %w", err)
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	if c.Version == nil {
		c.Version = ptr.ToPtr(int8(1))
	}

	if *c.Version != 1 {
		return fmt.Errorf("config: unsupported config version: %d", *c.Version)
	}

	if len(c.Source.Roots) == 0 {
		return fmt.Errorf("config: source.roots must contain at least one path")
	}

	if c.Output.Format == "" {
		c.Output.Format = DefaultOutputFormat
	}
	if c.Output.Format != DefaultOutputFormat {
		return fmt.Errorf("config: unsupported output format: %s", c.Output.Format)
	}
	return nil
}
