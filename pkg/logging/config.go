package logging

import (
	"encoding/json"
	"os"
	"strings"
)

// Config holds the logging configuration
type Config struct {
	Level      string              `json:"level" yaml:"level"`           // Log level: trace, debug, info, warn, error
	Format     string              `json:"format" yaml:"format"`         // Output format: text, json
	Output     []string            `json:"output" yaml:"output"`         // Output targets: console, file, syslog
	FilePath   string              `json:"file_path" yaml:"file_path"`   // Path for file output
	TimeFormat string              `json:"time_format" yaml:"time_format"` // Time format for logs
	NoColor    bool                `json:"no_color" yaml:"no_color"`     // Disable color output
	Modules    map[string]string   `json:"modules" yaml:"modules"`       // Per-module log levels
	MaxSize    int                 `json:"max_size" yaml:"max_size"`     // Max size in MB for log rotation
	MaxAge     int                 `json:"max_age" yaml:"max_age"`       // Max age in days for log retention
	MaxBackups int                 `json:"max_backups" yaml:"max_backups"` // Max number of old log files
}

// DefaultConfig returns the default logging configuration
func DefaultConfig() *Config {
	return &Config{
		Level:      "info",
		Format:     "text",
		Output:     []string{"console"},
		FilePath:   "",
		TimeFormat: "2006-01-02 15:04:05",
		NoColor:    false,
		Modules:    make(map[string]string),
		MaxSize:    100, // 100 MB
		MaxAge:     30,  // 30 days
		MaxBackups: 10,
	}
}

// Clone creates a deep copy of the configuration
func (c *Config) Clone() *Config {
	config := &Config{
		Level:      c.Level,
		Format:     c.Format,
		Output:     make([]string, len(c.Output)),
		FilePath:   c.FilePath,
		TimeFormat: c.TimeFormat,
		NoColor:    c.NoColor,
		Modules:    make(map[string]string),
		MaxSize:    c.MaxSize,
		MaxAge:     c.MaxAge,
		MaxBackups: c.MaxBackups,
	}

	copy(config.Output, c.Output)
	for k, v := range c.Modules {
		config.Modules[k] = v
	}

	return config
}

// LoadFromEnv loads configuration from environment variables
func (c *Config) LoadFromEnv() {
	// PORTUNIX_LOG_LEVEL
	if level := os.Getenv("PORTUNIX_LOG_LEVEL"); level != "" {
		c.Level = level
	}

	// PORTUNIX_LOG_FORMAT
	if format := os.Getenv("PORTUNIX_LOG_FORMAT"); format != "" {
		c.Format = format
	}

	// PORTUNIX_LOG_OUTPUT (comma-separated)
	if output := os.Getenv("PORTUNIX_LOG_OUTPUT"); output != "" {
		c.Output = strings.Split(output, ",")
	}

	// PORTUNIX_LOG_FILE
	if filePath := os.Getenv("PORTUNIX_LOG_FILE"); filePath != "" {
		c.FilePath = filePath
	}

	// PORTUNIX_LOG_NO_COLOR
	if noColor := os.Getenv("PORTUNIX_LOG_NO_COLOR"); noColor == "true" || noColor == "1" {
		c.NoColor = true
	}

	// Module-specific levels: PORTUNIX_LOG_MODULE_<MODULE>=<LEVEL>
	for _, env := range os.Environ() {
		if strings.HasPrefix(env, "PORTUNIX_LOG_MODULE_") {
			parts := strings.SplitN(env, "=", 2)
			if len(parts) == 2 {
				module := strings.ToLower(strings.TrimPrefix(parts[0], "PORTUNIX_LOG_MODULE_"))
				level := parts[1]
				c.Modules[module] = level
			}
		}
	}
}

// LoadFromFile loads configuration from a JSON or YAML file
func (c *Config) LoadFromFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	// Try JSON first
	if err := json.Unmarshal(data, c); err != nil {
		// If JSON fails, could try YAML in future
		return err
	}

	return nil
}

// Validate checks if the configuration is valid
func (c *Config) Validate() error {
	// Validate log level
	validLevels := []string{"trace", "debug", "info", "warn", "warning", "error", "fatal", "panic"}
	if !contains(validLevels, strings.ToLower(c.Level)) {
		c.Level = "info" // Default to info if invalid
	}

	// Validate format
	if c.Format != "text" && c.Format != "json" {
		c.Format = "text"
	}

	// Validate output targets
	validOutputs := []string{"console", "stdout", "stderr", "file", "syslog"}
	var outputs []string
	for _, output := range c.Output {
		if contains(validOutputs, strings.ToLower(output)) {
			outputs = append(outputs, output)
		}
	}
	if len(outputs) == 0 {
		outputs = []string{"console"} // Default to console
	}
	c.Output = outputs

	// Validate module levels
	for module, level := range c.Modules {
		if !contains(validLevels, strings.ToLower(level)) {
			c.Modules[module] = c.Level // Use global level as fallback
		}
	}

	return nil
}

// GetModuleLevel returns the log level for a specific module
func (c *Config) GetModuleLevel(module string) string {
	if level, ok := c.Modules[strings.ToLower(module)]; ok {
		return level
	}
	return c.Level
}

// SetModuleLevel sets the log level for a specific module
func (c *Config) SetModuleLevel(module, level string) {
	c.Modules[strings.ToLower(module)] = level
}

// Helper function to check if slice contains string
func contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

// ConfigManager manages runtime configuration changes
type ConfigManager struct {
	config   *Config
	factory  *Factory
	loggers  map[string]Logger
}

// NewConfigManager creates a new configuration manager
func NewConfigManager(config *Config) *ConfigManager {
	if config == nil {
		config = DefaultConfig()
	}
	config.LoadFromEnv()
	config.Validate()

	return &ConfigManager{
		config:  config,
		factory: NewFactory(config),
		loggers: make(map[string]Logger),
	}
}

// GetLogger returns a logger for the specified component
func (cm *ConfigManager) GetLogger(component string) Logger {
	if logger, ok := cm.loggers[component]; ok {
		return logger
	}

	logger := cm.factory.CreateLogger(component)
	cm.loggers[component] = logger
	return logger
}

// UpdateConfig updates the configuration and recreates all loggers
func (cm *ConfigManager) UpdateConfig(config *Config) {
	config.Validate()
	cm.config = config
	cm.factory = NewFactory(config)

	// Recreate all existing loggers with new configuration
	for component := range cm.loggers {
		cm.loggers[component] = cm.factory.CreateLogger(component)
	}
}

// SetLevel sets the global log level
func (cm *ConfigManager) SetLevel(level string) {
	cm.config.Level = level
	cm.UpdateConfig(cm.config)
}

// SetModuleLevel sets the log level for a specific module
func (cm *ConfigManager) SetModuleLevel(module, level string) {
	cm.config.SetModuleLevel(module, level)
	cm.UpdateConfig(cm.config)
}

// Global configuration manager instance
var globalConfigManager = NewConfigManager(nil)

// GetLogger returns a logger for the specified component using global config
func GetLogger(component string) Logger {
	return globalConfigManager.GetLogger(component)
}

// UpdateGlobalConfig updates the global logging configuration
func UpdateGlobalConfig(config *Config) {
	globalConfigManager.UpdateConfig(config)
}

// SetGlobalLogLevel sets the global log level
func SetGlobalLogLevel(level string) {
	globalConfigManager.SetLevel(level)
}

// SetModuleLogLevel sets the log level for a specific module globally
func SetModuleLogLevel(module, level string) {
	globalConfigManager.SetModuleLevel(module, level)
}