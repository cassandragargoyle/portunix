package logging

import (
	"os"
	"reflect"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.Level != "info" {
		t.Errorf("Expected default level 'info', got '%s'", config.Level)
	}
	if config.Format != "text" {
		t.Errorf("Expected default format 'text', got '%s'", config.Format)
	}
	if !reflect.DeepEqual(config.Output, []string{"console"}) {
		t.Errorf("Expected default output ['console'], got %v", config.Output)
	}
	if config.TimeFormat != "2006-01-02 15:04:05" {
		t.Errorf("Expected default time format '2006-01-02 15:04:05', got '%s'", config.TimeFormat)
	}
	if config.NoColor != false {
		t.Errorf("Expected default NoColor false, got %t", config.NoColor)
	}
}

func TestConfigClone(t *testing.T) {
	original := &Config{
		Level:      "debug",
		Format:     "json",
		Output:     []string{"console", "file"},
		FilePath:   "/var/log/test.log",
		TimeFormat: "15:04:05",
		NoColor:    true,
		Modules:    map[string]string{"test": "warn"},
		MaxSize:    50,
		MaxAge:     7,
		MaxBackups: 5,
	}

	clone := original.Clone()

	// Test that clone is not the same instance
	if original == clone {
		t.Error("Clone returned same instance")
	}

	// Test that values are copied
	if clone.Level != original.Level {
		t.Errorf("Level not cloned correctly: expected %s, got %s", original.Level, clone.Level)
	}
	if clone.Format != original.Format {
		t.Errorf("Format not cloned correctly: expected %s, got %s", original.Format, clone.Format)
	}
	if !reflect.DeepEqual(clone.Output, original.Output) {
		t.Errorf("Output not cloned correctly: expected %v, got %v", original.Output, clone.Output)
	}
	if clone.FilePath != original.FilePath {
		t.Errorf("FilePath not cloned correctly: expected %s, got %s", original.FilePath, clone.FilePath)
	}

	// Test that slice and map are deep copied
	clone.Output[0] = "modified"
	if original.Output[0] == "modified" {
		t.Error("Output slice was not deep copied")
	}

	clone.Modules["test"] = "error"
	if original.Modules["test"] == "error" {
		t.Error("Modules map was not deep copied")
	}
}

func TestConfigLoadFromEnv(t *testing.T) {
	// Save original environment
	originalEnv := map[string]string{
		"PORTUNIX_LOG_LEVEL":        os.Getenv("PORTUNIX_LOG_LEVEL"),
		"PORTUNIX_LOG_FORMAT":       os.Getenv("PORTUNIX_LOG_FORMAT"),
		"PORTUNIX_LOG_OUTPUT":       os.Getenv("PORTUNIX_LOG_OUTPUT"),
		"PORTUNIX_LOG_FILE":         os.Getenv("PORTUNIX_LOG_FILE"),
		"PORTUNIX_LOG_NO_COLOR":     os.Getenv("PORTUNIX_LOG_NO_COLOR"),
		"PORTUNIX_LOG_MODULE_TEST":  os.Getenv("PORTUNIX_LOG_MODULE_TEST"),
	}

	// Set test environment variables
	os.Setenv("PORTUNIX_LOG_LEVEL", "debug")
	os.Setenv("PORTUNIX_LOG_FORMAT", "json")
	os.Setenv("PORTUNIX_LOG_OUTPUT", "console,file")
	os.Setenv("PORTUNIX_LOG_FILE", "/tmp/test.log")
	os.Setenv("PORTUNIX_LOG_NO_COLOR", "true")
	os.Setenv("PORTUNIX_LOG_MODULE_TEST", "warn")

	defer func() {
		// Restore original environment
		for key, value := range originalEnv {
			if value == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, value)
			}
		}
	}()

	config := DefaultConfig()
	config.LoadFromEnv()

	if config.Level != "debug" {
		t.Errorf("Expected level 'debug', got '%s'", config.Level)
	}
	if config.Format != "json" {
		t.Errorf("Expected format 'json', got '%s'", config.Format)
	}
	if !reflect.DeepEqual(config.Output, []string{"console", "file"}) {
		t.Errorf("Expected output ['console', 'file'], got %v", config.Output)
	}
	if config.FilePath != "/tmp/test.log" {
		t.Errorf("Expected file path '/tmp/test.log', got '%s'", config.FilePath)
	}
	if config.NoColor != true {
		t.Errorf("Expected NoColor true, got %t", config.NoColor)
	}
	if config.Modules["test"] != "warn" {
		t.Errorf("Expected module 'test' level 'warn', got '%s'", config.Modules["test"])
	}
}

func TestConfigValidate(t *testing.T) {
	testCases := []struct {
		name           string
		config         *Config
		expectedLevel  string
		expectedFormat string
		expectedOutput []string
	}{
		{
			name: "Valid config",
			config: &Config{
				Level:  "debug",
				Format: "json",
				Output: []string{"console", "file"},
			},
			expectedLevel:  "debug",
			expectedFormat: "json",
			expectedOutput: []string{"console", "file"},
		},
		{
			name: "Invalid level",
			config: &Config{
				Level:  "invalid",
				Format: "text",
				Output: []string{"console"},
			},
			expectedLevel:  "info",
			expectedFormat: "text",
			expectedOutput: []string{"console"},
		},
		{
			name: "Invalid format",
			config: &Config{
				Level:  "info",
				Format: "invalid",
				Output: []string{"console"},
			},
			expectedLevel:  "info",
			expectedFormat: "text",
			expectedOutput: []string{"console"},
		},
		{
			name: "Invalid output",
			config: &Config{
				Level:  "info",
				Format: "text",
				Output: []string{"invalid"},
			},
			expectedLevel:  "info",
			expectedFormat: "text",
			expectedOutput: []string{"console"},
		},
		{
			name: "Empty output",
			config: &Config{
				Level:  "info",
				Format: "text",
				Output: []string{},
			},
			expectedLevel:  "info",
			expectedFormat: "text",
			expectedOutput: []string{"console"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.config.Validate()
			if err != nil {
				t.Errorf("Validate() returned error: %v", err)
			}

			if tc.config.Level != tc.expectedLevel {
				t.Errorf("Expected level '%s', got '%s'", tc.expectedLevel, tc.config.Level)
			}
			if tc.config.Format != tc.expectedFormat {
				t.Errorf("Expected format '%s', got '%s'", tc.expectedFormat, tc.config.Format)
			}
			if !reflect.DeepEqual(tc.config.Output, tc.expectedOutput) {
				t.Errorf("Expected output %v, got %v", tc.expectedOutput, tc.config.Output)
			}
		})
	}
}

func TestConfigGetModuleLevel(t *testing.T) {
	config := &Config{
		Level: "info",
		Modules: map[string]string{
			"test":    "debug",
			"another": "warn",
		},
	}

	// Test existing module
	if level := config.GetModuleLevel("test"); level != "debug" {
		t.Errorf("Expected module 'test' level 'debug', got '%s'", level)
	}

	// Test non-existing module (should return global level)
	if level := config.GetModuleLevel("nonexistent"); level != "info" {
		t.Errorf("Expected non-existent module level 'info', got '%s'", level)
	}

	// Test case insensitive
	if level := config.GetModuleLevel("TEST"); level != "debug" {
		t.Errorf("Expected case-insensitive module 'TEST' level 'debug', got '%s'", level)
	}
}

func TestConfigSetModuleLevel(t *testing.T) {
	config := DefaultConfig()

	config.SetModuleLevel("test", "debug")

	if config.Modules["test"] != "debug" {
		t.Errorf("Expected module 'test' level 'debug', got '%s'", config.Modules["test"])
	}

	// Test case insensitive
	config.SetModuleLevel("TEST", "warn")
	if config.Modules["test"] != "warn" {
		t.Errorf("Expected case-insensitive module 'test' level 'warn', got '%s'", config.Modules["test"])
	}
}

func TestConfigManager(t *testing.T) {
	config := DefaultConfig()
	manager := NewConfigManager(config)

	if manager == nil {
		t.Fatal("NewConfigManager() returned nil")
	}

	// Test getting logger
	logger1 := manager.GetLogger("test")
	if logger1 == nil {
		t.Fatal("GetLogger() returned nil")
	}

	// Test getting same logger returns same instance
	logger2 := manager.GetLogger("test")
	if logger1 != logger2 {
		t.Error("GetLogger() returned different instances for same component")
	}

	// Test updating config
	newConfig := config.Clone()
	newConfig.Level = "debug"
	manager.UpdateConfig(newConfig)

	// Test setting global level
	manager.SetLevel("warn")

	// Test setting module level
	manager.SetModuleLevel("test", "error")
}

func TestGlobalConfigFunctions(t *testing.T) {
	// Test GetLogger
	logger := GetLogger("test")
	if logger == nil {
		t.Fatal("GetLogger() returned nil")
	}

	// Test SetGlobalLogLevel
	SetGlobalLogLevel("debug")

	// Test SetModuleLogLevel
	SetModuleLogLevel("test", "warn")

	// Test UpdateGlobalConfig
	config := DefaultConfig()
	config.Level = "error"
	UpdateGlobalConfig(config)
}

func BenchmarkConfigValidate(b *testing.B) {
	config := &Config{
		Level:  "info",
		Format: "json",
		Output: []string{"console", "file"},
		Modules: map[string]string{
			"module1": "debug",
			"module2": "warn",
			"module3": "error",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		config.Validate()
	}
}

func BenchmarkConfigClone(b *testing.B) {
	config := &Config{
		Level:      "info",
		Format:     "json",
		Output:     []string{"console", "file", "syslog"},
		FilePath:   "/var/log/app.log",
		TimeFormat: "2006-01-02 15:04:05",
		NoColor:    false,
		Modules: map[string]string{
			"module1": "debug",
			"module2": "warn",
			"module3": "error",
			"module4": "info",
			"module5": "trace",
		},
		MaxSize:    100,
		MaxAge:     30,
		MaxBackups: 10,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		config.Clone()
	}
}