package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"

	"gopkg.in/yaml.v3"
)

// Config is the plugin configuration from config.yaml
type Config struct {
	Plugin   PluginConfig `yaml:"plugin"`
	LogLevel string       `yaml:"log_level"`
	Timeout  int          `yaml:"timeout"`
}

// PluginConfig is the plugin section of config.yaml
type PluginConfig struct {
	Name        string `yaml:"name"`
	Version     string `yaml:"version"`
	Description string `yaml:"description"`
}

// Load reads the configuration; a missing file falls back to the defaults
func Load(path string) (*Config, error) {
	var config Config
	data, err := os.ReadFile(path) // #nosec G304 -- path of the plugin's own config
	switch {
	case errors.Is(err, fs.ErrNotExist):
		config.Plugin = PluginConfig{Name: "my-go-plugin", Version: "1.0.0"}
	case err != nil:
		return nil, fmt.Errorf("failed to read config file: %w", err)
	default:
		if err := yaml.Unmarshal(data, &config); err != nil {
			return nil, fmt.Errorf("failed to parse config: %w", err)
		}
	}

	// Set defaults
	if config.LogLevel == "" {
		config.LogLevel = "info"
	}
	if config.Timeout == 0 {
		config.Timeout = 30
	}

	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// Validate checks the configuration values
func (c *Config) Validate() error {
	if c.Plugin.Name == "" {
		return fmt.Errorf("plugin name is required")
	}
	if c.Plugin.Version == "" {
		return fmt.Errorf("plugin version is required")
	}
	if c.Timeout <= 0 {
		return fmt.Errorf("timeout must be positive")
	}
	return nil
}
