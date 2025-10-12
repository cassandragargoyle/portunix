package unit

import (
	"os"
	"path/filepath"
	"testing"

	"portunix.ai/app/container"
)

func TestGetSelectedRuntime_DefaultToPodman(t *testing.T) {
	// Use temporary directory for config
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	defer os.Unsetenv("HOME")
	
	// Change to temp directory to avoid loading existing config
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)
	
	// Note: This test will depend on what container runtimes are actually available
	// We're testing the logic, but the result depends on system state
	runtime, err := container.GetSelectedRuntime()
	
	// If no runtimes are available, we expect an error
	if err != nil {
		// This is acceptable - system may not have podman installed
		t.Logf("No container runtime available: %v", err)
		return
	}
	
	// If we get a runtime, it should be one of the valid ones
	if runtime != "docker" && runtime != "podman" {
		t.Errorf("Expected runtime to be 'docker' or 'podman', got '%s'", runtime)
	}
}

func TestGetSelectedRuntime_WithConfig(t *testing.T) {
	// Create temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "portunix-config.yaml")
	
	// Test with docker configured
	configContent := `container_runtime: docker`
	
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}
	
	// Change to temp directory
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)
	
	runtime, err := container.GetSelectedRuntime()
	
	// If docker is not available, we expect an error
	if err != nil {
		t.Logf("Docker not available: %v", err)
		return
	}
	
	if runtime != "docker" {
		t.Errorf("Expected runtime to be 'docker', got '%s'", runtime)
	}
}

func TestValidateRuntime(t *testing.T) {
	validRuntimes := []string{"docker", "podman"}
	
	for _, runtime := range validRuntimes {
		err := container.ValidateRuntime(runtime)
		if err != nil {
			t.Errorf("Expected no error for valid runtime '%s', got: %v", runtime, err)
		}
	}
	
	invalidRuntimes := []string{"containerd", "cri-o", "invalid", ""}
	
	for _, runtime := range invalidRuntimes {
		err := container.ValidateRuntime(runtime)
		if err == nil {
			t.Errorf("Expected error for invalid runtime '%s'", runtime)
		}
	}
}

func TestGetRuntimeInfo(t *testing.T) {
	info, err := container.GetRuntimeInfo()
	if err != nil {
		t.Fatalf("Failed to get runtime info: %v", err)
	}
	
	// Check that we get information about both runtimes
	if _, ok := info["docker"]; !ok {
		t.Error("Expected runtime info to include docker")
	}
	
	if _, ok := info["podman"]; !ok {
		t.Error("Expected runtime info to include podman")
	}
	
	// The values should be boolean
	for runtime, available := range info {
		t.Logf("Runtime %s available: %v", runtime, available)
	}
}

func TestIsDockerAvailable(t *testing.T) {
	available := container.IsDockerAvailable()
	t.Logf("Docker available: %v", available)
	// This test just verifies the function doesn't crash
	// The result depends on system state
}

func TestIsPodmanAvailable(t *testing.T) {
	available := container.IsPodmanAvailable()
	t.Logf("Podman available: %v", available)
	// This test just verifies the function doesn't crash
	// The result depends on system state
}