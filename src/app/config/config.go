package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the main configuration structure
type Config struct {
	ContainerRuntime string `yaml:"container_runtime,omitempty"` // docker, podman
	Verbose          bool   `yaml:"verbose,omitempty"`
	AutoUpdate       bool   `yaml:"auto_update,omitempty"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		ContainerRuntime: "podman", // Default to podman
		Verbose:          false,
		AutoUpdate:       true,
	}
}

// LoadConfig loads configuration from file with priority order
func LoadConfig() (*Config, error) {
	config := DefaultConfig()
	
	// Try configuration files in priority order
	configPaths := []string{
		"./portunix-config.yaml",                    // Project directory
		filepath.Join(getUserHome(), ".portunix", "config.yaml"), // User directory
		filepath.Join(getUserHome(), ".config", "portunix", "config.yaml"), // Linux standard
	}
	
	var loadedFrom string
	for _, path := range configPaths {
		if fileExists(path) {
			if err := loadConfigFromFile(config, path); err != nil {
				return nil, fmt.Errorf("failed to load config from %s: %w", path, err)
			}
			loadedFrom = path
			break
		}
	}
	
	// If no config file found, use defaults
	if loadedFrom == "" {
		// Use built-in defaults
	}
	
	return config, nil
}

// loadConfigFromFile loads configuration from a specific file
func loadConfigFromFile(config *Config, path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	
	return yaml.Unmarshal(data, config)
}

// SaveConfig saves configuration to the user config file
func SaveConfig(config *Config) error {
	userConfigDir := filepath.Join(getUserHome(), ".portunix")
	if err := os.MkdirAll(userConfigDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	
	configPath := filepath.Join(userConfigDir, "config.yaml")
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	
	return os.WriteFile(configPath, data, 0644)
}

// GetConfigValue returns a specific configuration value
func GetConfigValue(key string) (string, error) {
	config, err := LoadConfig()
	if err != nil {
		return "", err
	}
	
	switch key {
	case "container_runtime":
		return config.ContainerRuntime, nil
	case "verbose":
		if config.Verbose {
			return "true", nil
		}
		return "false", nil
	case "auto_update":
		if config.AutoUpdate {
			return "true", nil
		}
		return "false", nil
	default:
		return "", fmt.Errorf("unknown configuration key: %s", key)
	}
}

// SetConfigValue sets a specific configuration value
func SetConfigValue(key, value string) error {
	config, err := LoadConfig()
	if err != nil {
		return err
	}
	
	switch key {
	case "container_runtime":
		if value != "docker" && value != "podman" {
			return fmt.Errorf("invalid container runtime: %s (must be 'docker' or 'podman')", value)
		}
		config.ContainerRuntime = value
	case "verbose":
		config.Verbose = value == "true"
	case "auto_update":
		config.AutoUpdate = value == "true"
	default:
		return fmt.Errorf("unknown configuration key: %s", key)
	}
	
	return SaveConfig(config)
}

// fileExists checks if a file exists
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// getUserHome returns the user's home directory
func getUserHome() string {
	if home := os.Getenv("HOME"); home != "" {
		return home
	}
	return os.Getenv("USERPROFILE") // Windows
}