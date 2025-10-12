package datastore

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents the complete datastore configuration
type Config struct {
	DefaultPlugin string                  `yaml:"default_plugin"`
	Routes        []RouteConfig           `yaml:"routes"`
	Plugins       map[string]PluginConfig `yaml:"plugins"`
}

// RouteConfig defines routing rules for data storage
type RouteConfig struct {
	Name    string                 `yaml:"name"`
	Pattern string                 `yaml:"pattern"`
	Plugin  string                 `yaml:"plugin"`
	Config  map[string]interface{} `yaml:"config"`
}

// PluginConfig contains plugin-specific configuration
type PluginConfig struct {
	ConnectionString string                 `yaml:"connection_string,omitempty"`
	Auth             AuthConfig             `yaml:"auth,omitempty"`
	Settings         map[string]interface{} `yaml:"settings,omitempty"`
}

// AuthConfig contains authentication configuration
type AuthConfig struct {
	Username string `yaml:"username,omitempty"`
	Password string `yaml:"password,omitempty"`
	Token    string `yaml:"token,omitempty"`
}

// LoadConfig loads datastore configuration from file
func LoadConfig(configPath string) (*Config, error) {
	if configPath == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get user home directory: %w", err)
		}
		configPath = filepath.Join(homeDir, ".portunix", "datastore.yaml")
	}

	// Create default config if file doesn't exist
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		defaultConfig := getDefaultConfig()
		if err := SaveConfig(defaultConfig, configPath); err != nil {
			return nil, fmt.Errorf("failed to create default config: %w", err)
		}
		return defaultConfig, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Expand environment variables
	expandEnvVars(&config)

	return &config, nil
}

// SaveConfig saves configuration to file
func SaveConfig(config *Config, configPath string) error {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// getDefaultConfig returns the default datastore configuration
func getDefaultConfig() *Config {
	return &Config{
		DefaultPlugin: "file-plugin",
		Routes: []RouteConfig{
			{
				Name:    "user_documentation",
				Pattern: "docs/*",
				Plugin:  "file-plugin",
				Config: map[string]interface{}{
					"base_path": "~/.portunix/data/docs",
					"format":    "yaml",
				},
			},
			{
				Name:    "system_logs",
				Pattern: "logs/*",
				Plugin:  "file-plugin",
				Config: map[string]interface{}{
					"base_path": "~/.portunix/data/logs",
					"format":    "json",
				},
			},
			{
				Name:    "project_metadata",
				Pattern: "projects/*",
				Plugin:  "file-plugin",
				Config: map[string]interface{}{
					"base_path": "~/.portunix/data/projects",
					"format":    "yaml",
				},
			},
			{
				Name:    "cache_data",
				Pattern: "cache/*",
				Plugin:  "file-plugin",
				Config: map[string]interface{}{
					"base_path": "~/.portunix/data/cache",
					"format":    "json",
					"ttl":       3600,
				},
			},
			{
				Name:    "fallback",
				Pattern: "*",
				Plugin:  "file-plugin",
				Config: map[string]interface{}{
					"base_path": "~/.portunix/data",
					"format":    "yaml",
				},
			},
		},
		Plugins: map[string]PluginConfig{
			"file-plugin": {
				Settings: map[string]interface{}{
					"create_directories": true,
					"backup_enabled":     true,
				},
			},
		},
	}
}

// FindRoute finds the appropriate route for a given key
func (c *Config) FindRoute(key string) *RouteConfig {
	// Check specific routes first (more specific patterns)
	for _, route := range c.Routes {
		if route.Pattern == "*" {
			continue // Skip wildcard, check it last
		}
		if matchPattern(route.Pattern, key) {
			return &route
		}
	}

	// Check wildcard routes
	for _, route := range c.Routes {
		if route.Pattern == "*" {
			return &route
		}
	}

	return nil
}

// matchPattern checks if a key matches a pattern
func matchPattern(pattern, key string) bool {
	// Simple pattern matching for now
	// TODO: Implement more sophisticated pattern matching (regex, glob, etc.)

	if pattern == "*" {
		return true
	}

	// Handle prefix patterns like "docs/*"
	if strings.HasSuffix(pattern, "/*") {
		prefix := strings.TrimSuffix(pattern, "/*")
		return strings.HasPrefix(key, prefix+"/") || key == prefix
	}

	// Handle suffix patterns like "*.log"
	if strings.HasPrefix(pattern, "*.") {
		suffix := strings.TrimPrefix(pattern, "*")
		return strings.HasSuffix(key, suffix)
	}

	// Exact match
	return pattern == key
}

// expandEnvVars expands environment variables in the configuration
func expandEnvVars(config *Config) {
	for pluginName, pluginConfig := range config.Plugins {
		if pluginConfig.ConnectionString != "" {
			pluginConfig.ConnectionString = os.ExpandEnv(pluginConfig.ConnectionString)
		}
		if pluginConfig.Auth.Username != "" {
			pluginConfig.Auth.Username = os.ExpandEnv(pluginConfig.Auth.Username)
		}
		if pluginConfig.Auth.Password != "" {
			pluginConfig.Auth.Password = os.ExpandEnv(pluginConfig.Auth.Password)
		}
		if pluginConfig.Auth.Token != "" {
			pluginConfig.Auth.Token = os.ExpandEnv(pluginConfig.Auth.Token)
		}
		config.Plugins[pluginName] = pluginConfig
	}
}

// ValidateConfig validates the configuration
func ValidateConfig(config *Config) error {
	if config.DefaultPlugin == "" {
		return fmt.Errorf("default_plugin cannot be empty")
	}

	// Check if default plugin exists in plugins section
	if _, exists := config.Plugins[config.DefaultPlugin]; !exists {
		return fmt.Errorf("default plugin '%s' not found in plugins configuration", config.DefaultPlugin)
	}

	// Validate routes
	for i, route := range config.Routes {
		if route.Name == "" {
			return fmt.Errorf("route %d: name cannot be empty", i)
		}
		if route.Pattern == "" {
			return fmt.Errorf("route %d: pattern cannot be empty", i)
		}
		if route.Plugin == "" {
			return fmt.Errorf("route %d: plugin cannot be empty", i)
		}

		// Check if plugin exists
		if _, exists := config.Plugins[route.Plugin]; !exists {
			return fmt.Errorf("route %d: plugin '%s' not found in plugins configuration", i, route.Plugin)
		}
	}

	return nil
}
