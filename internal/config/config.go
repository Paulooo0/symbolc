package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

const DefaultFile = "symbolc.yml"

type Config struct {
	Version   int             `yaml:"version"`
	Source    SourceConfig    `yaml:"source"`
	Contracts ContractsConfig `yaml:"contracts"`
}

type SourceConfig struct {
	Roots []string `yaml:"roots"`
}

type ContractsConfig struct {
	Include []string `yaml:"include"`
	Exclude []string `yaml:"exclude"`
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

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config: invalid config: %w", err)
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	if c.Version != 1 {
		return fmt.Errorf("config: unsupported config version: %d", c.Version)
	}

	if len(c.Source.Roots) == 0 {
		return fmt.Errorf("config: source.roots must contain at least one path")
	}
	return nil
}
