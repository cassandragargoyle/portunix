package unit

import (
	"os"
	"path/filepath"
	"testing"

	"portunix.ai/app/config"
)

func TestDefaultConfig(t *testing.T) {
	cfg := config.DefaultConfig()
	
	if cfg.ContainerRuntime != "podman" {
		t.Errorf("Expected default container runtime to be 'podman', got '%s'", cfg.ContainerRuntime)
	}
	
	if cfg.Verbose != false {
		t.Errorf("Expected default verbose to be false, got %v", cfg.Verbose)
	}
	
	if cfg.AutoUpdate != true {
		t.Errorf("Expected default auto_update to be true, got %v", cfg.AutoUpdate)
	}
}

func TestLoadConfigWithDefaults(t *testing.T) {
	// Test loading config when no config file exists
	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load default config: %v", err)
	}
	
	if cfg.ContainerRuntime != "podman" {
		t.Errorf("Expected default container runtime to be 'podman', got '%s'", cfg.ContainerRuntime)
	}
}

func TestLoadConfigFromFile(t *testing.T) {
	// Create temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "portunix-config.yaml")
	
	configContent := `container_runtime: docker
verbose: true
auto_update: false`
	
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}
	
	// Change to temp directory so the config file is found
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)
	
	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("Failed to load config from file: %v", err)
	}
	
	if cfg.ContainerRuntime != "docker" {
		t.Errorf("Expected container runtime to be 'docker', got '%s'", cfg.ContainerRuntime)
	}
	
	if cfg.Verbose != true {
		t.Errorf("Expected verbose to be true, got %v", cfg.Verbose)
	}
	
	if cfg.AutoUpdate != false {
		t.Errorf("Expected auto_update to be false, got %v", cfg.AutoUpdate)
	}
}

func TestGetConfigValue(t *testing.T) {
	// Create temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "portunix-config.yaml")
	
	configContent := `container_runtime: docker
verbose: true`
	
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}
	
	// Change to temp directory
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)
	
	tests := []struct {
		key      string
		expected string
	}{
		{"container_runtime", "docker"},
		{"verbose", "true"},
		{"auto_update", "true"}, // default value
	}
	
	for _, test := range tests {
		value, err := config.GetConfigValue(test.key)
		if err != nil {
			t.Errorf("Failed to get config value for %s: %v", test.key, err)
		}
		if value != test.expected {
			t.Errorf("Expected %s to be '%s', got '%s'", test.key, test.expected, value)
		}
	}
}

func TestSetConfigValue(t *testing.T) {
	// Use temporary directory for config
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	defer os.Unsetenv("HOME")
	
	// Test setting container runtime
	err := config.SetConfigValue("container_runtime", "docker")
	if err != nil {
		t.Fatalf("Failed to set container_runtime: %v", err)
	}
	
	// Verify the value was set
	value, err := config.GetConfigValue("container_runtime")
	if err != nil {
		t.Fatalf("Failed to get container_runtime: %v", err)
	}
	
	if value != "docker" {
		t.Errorf("Expected container_runtime to be 'docker', got '%s'", value)
	}
	
	// Test setting invalid container runtime
	err = config.SetConfigValue("container_runtime", "invalid")
	if err == nil {
		t.Error("Expected error when setting invalid container runtime")
	}
}

func TestValidateContainerRuntime(t *testing.T) {
	// Test setting valid container runtimes
	validRuntimes := []string{"docker", "podman"}
	
	for _, runtime := range validRuntimes {
		err := config.SetConfigValue("container_runtime", runtime)
		if err != nil {
			t.Errorf("Expected no error for valid runtime '%s', got: %v", runtime, err)
		}
	}
	
	// Test setting invalid container runtime
	err := config.SetConfigValue("container_runtime", "containerd")
	if err == nil {
		t.Error("Expected error when setting invalid container runtime 'containerd'")
	}
}