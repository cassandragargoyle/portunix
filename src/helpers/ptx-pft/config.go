package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	ConfigFileName = ".pft-config.json"
	CacheFileName  = ".pft-cache.json"
)

// SyncConfig holds synchronization settings
type SyncConfig struct {
	Auto               bool   `json:"auto"`
	Interval           string `json:"interval"`
	ConflictResolution string `json:"conflict_resolution"`
}

// StatusMappings defines how local statuses map to provider statuses
type StatusMappings struct {
	Open      string `json:"open"`
	Planned   string `json:"planned"`
	Started   string `json:"started"`
	Completed string `json:"completed"`
	Declined  string `json:"declined"`
}

// Mappings holds all mapping configurations
type Mappings struct {
	Status StatusMappings `json:"status"`
}

// SMTPConfig holds SMTP server configuration
type SMTPConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	From     string `json:"from"`
}

// AreaConfig holds configuration for a single area (voc, vos, vob, voe)
type AreaConfig struct {
	Provider  string `json:"provider,omitempty"`   // fider, clearflask, eververse, local
	URL       string `json:"url,omitempty"`        // Provider endpoint URL
	APIToken  string `json:"api_token,omitempty"`  // API token for authentication
	ProjectID string `json:"project_id,omitempty"` // For ClearFlask multi-project
	ProductID string `json:"product_id,omitempty"` // For Eververse multi-product
}

// Config represents the .pft-config.json structure
type Config struct {
	Name     string      `json:"name"`
	Path     string      `json:"path"`
	SMTP     *SMTPConfig `json:"smtp,omitempty"` // SMTP configuration for notifications
	VoC      *AreaConfig `json:"voc,omitempty"`  // Voice of Customer
	VoS      *AreaConfig `json:"vos,omitempty"`  // Voice of Stakeholder
	VoB      *AreaConfig `json:"vob,omitempty"`  // Voice of Business
	VoE      *AreaConfig `json:"voe,omitempty"`  // Voice of Engineer
	Sync     SyncConfig  `json:"sync"`
	Mappings Mappings    `json:"mappings"`
}

// NewDefaultConfig creates a new Config with default values
func NewDefaultConfig() *Config {
	return &Config{
		// Provider is empty by default (local/offline mode)
		Sync: SyncConfig{
			Auto:               false,
			Interval:           "1h",
			ConflictResolution: "timestamp",
		},
		Mappings: Mappings{
			Status: StatusMappings{
				Open:      "pending",
				Planned:   "in_progress",
				Started:   "in_progress",
				Completed: "implemented",
				Declined:  "rejected",
			},
		},
	}
}

// LoadConfig loads configuration from .pft-config.json
// It searches in the current directory and parent directories
func LoadConfig() (*Config, error) {
	configPath, err := findConfigFile()
	if err != nil {
		return nil, err
	}

	return LoadConfigFromPath(configPath)
}

// LoadConfigFromPath loads configuration from a specific path
func LoadConfigFromPath(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// Save writes the configuration to .pft-config.json
func (c *Config) Save(dir string) error {
	path := filepath.Join(dir, ConfigFileName)
	return c.SaveToPath(path)
}

// SaveToPath writes the configuration to a specific path
func (c *Config) SaveToPath(path string) error {
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("product name is required")
	}
	if c.Path == "" {
		return fmt.Errorf("document path is required")
	}

	// Check if path exists
	if _, err := os.Stat(c.Path); os.IsNotExist(err) {
		return fmt.Errorf("document path does not exist: %s", c.Path)
	}

	// Validate area configs
	areas := map[string]*AreaConfig{"voc": c.VoC, "vos": c.VoS, "vob": c.VoB, "voe": c.VoE}
	for name, area := range areas {
		if area != nil {
			if err := validateAreaConfig(name, area); err != nil {
				return err
			}
		}
	}

	return nil
}

// validateAreaConfig validates a single area configuration
func validateAreaConfig(name string, area *AreaConfig) error {
	if area.Provider == "" {
		return nil // local/unconfigured is valid
	}

	validProviders := []string{"fider", "clearflask", "eververse", "local"}
	isValid := false
	for _, p := range validProviders {
		if area.Provider == p {
			isValid = true
			break
		}
	}
	if !isValid {
		return fmt.Errorf("invalid provider '%s' for area %s", area.Provider, name)
	}

	// Provider-specific validation
	if area.Provider == "clearflask" && area.ProjectID == "" {
		return fmt.Errorf("project_id is required for ClearFlask provider in area %s", name)
	}
	if area.Provider != "local" && area.URL == "" {
		return fmt.Errorf("url is required for provider %s in area %s", area.Provider, name)
	}

	return nil
}

// findConfigFile searches for .pft-config.json in current and parent directories
func findConfigFile() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	for {
		configPath := filepath.Join(dir, ConfigFileName)
		if _, err := os.Stat(configPath); err == nil {
			return configPath, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("config file %s not found", ConfigFileName)
}

// ConfigExists checks if a configuration file exists in the given directory
func ConfigExists(dir string) bool {
	path := filepath.Join(dir, ConfigFileName)
	_, err := os.Stat(path)
	return err == nil
}

// GetConfigPath returns the path to the config file in the given directory
func GetConfigPath(dir string) string {
	return filepath.Join(dir, ConfigFileName)
}

// GetAreaConfig returns the AreaConfig for a given area name
func (c *Config) GetAreaConfig(area string) *AreaConfig {
	switch area {
	case "voc":
		return c.VoC
	case "vos":
		return c.VoS
	case "vob":
		return c.VoB
	case "voe":
		return c.VoE
	default:
		return nil
	}
}

// SetAreaConfig sets the AreaConfig for a given area name
func (c *Config) SetAreaConfig(area string, cfg *AreaConfig) {
	switch area {
	case "voc":
		c.VoC = cfg
	case "vos":
		c.VoS = cfg
	case "vob":
		c.VoB = cfg
	case "voe":
		c.VoE = cfg
	}
}

// GetAreaProviderConfig returns ProviderConfig for a specific area
func (c *Config) GetAreaProviderConfig(area string) ProviderConfig {
	areaCfg := c.GetAreaConfig(area)
	if areaCfg == nil {
		return ProviderConfig{}
	}

	options := make(map[string]string)
	if areaCfg.ProjectID != "" {
		options["project_id"] = areaCfg.ProjectID
	}
	if areaCfg.ProductID != "" {
		options["product_id"] = areaCfg.ProductID
	}

	return ProviderConfig{
		Endpoint: areaCfg.URL,
		APIToken: areaCfg.APIToken,
		Options:  options,
	}
}

// GetAreaProvider returns provider name for a specific area (defaults to "local")
func (c *Config) GetAreaProvider(area string) string {
	areaCfg := c.GetAreaConfig(area)
	if areaCfg == nil || areaCfg.Provider == "" {
		return "local"
	}
	return areaCfg.Provider
}

// Legacy compatibility methods - return first configured area's values

// GetProvider returns the first configured provider (legacy compatibility)
func (c *Config) GetProvider() string {
	for _, area := range []string{"voc", "vos", "vob", "voe"} {
		if cfg := c.GetAreaConfig(area); cfg != nil && cfg.Provider != "" {
			return cfg.Provider
		}
	}
	return "local"
}

// GetEndpoint returns the first configured endpoint URL (legacy compatibility)
func (c *Config) GetEndpoint() string {
	for _, area := range []string{"voc", "vos", "vob", "voe"} {
		if cfg := c.GetAreaConfig(area); cfg != nil && cfg.URL != "" {
			return cfg.URL
		}
	}
	return ""
}

// GetAPIToken returns the first configured API token (legacy compatibility)
func (c *Config) GetAPIToken() string {
	for _, area := range []string{"voc", "vos", "vob", "voe"} {
		if cfg := c.GetAreaConfig(area); cfg != nil && cfg.APIToken != "" {
			return cfg.APIToken
		}
	}
	return ""
}

// GetProjectID returns the first configured project ID (legacy compatibility)
func (c *Config) GetProjectID() string {
	for _, area := range []string{"voc", "vos", "vob", "voe"} {
		if cfg := c.GetAreaConfig(area); cfg != nil && cfg.ProjectID != "" {
			return cfg.ProjectID
		}
	}
	return ""
}

