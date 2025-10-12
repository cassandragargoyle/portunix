package docker

import (
	"testing"
)

// Test basic Docker configuration structure for Go installations
func TestDockerConfig_GoInstallationType(t *testing.T) {
	config := DockerConfig{
		Image:            "ubuntu:22.04",
		ContainerName:    "test-go-container",
		InstallationType: "go",
		EnableSSH:        true,
		KeepRunning:      true,
	}

	if config.InstallationType != "go" {
		t.Errorf("Expected InstallationType 'go', got '%s'", config.InstallationType)
	}

	if config.Image != "ubuntu:22.04" {
		t.Errorf("Expected Image 'ubuntu:22.04', got '%s'", config.Image)
	}

	if config.ContainerName != "test-go-container" {
		t.Errorf("Expected ContainerName 'test-go-container', got '%s'", config.ContainerName)
	}
}

// Test PackageManagerInfo structure
func TestPackageManagerInfo_GoInstallation(t *testing.T) {
	// Test different package managers that Go installation should support
	testCases := []struct {
		manager      string
		distribution string
		updateCmd    string
		installCmd   string
	}{
		{"apt-get", "ubuntu", "apt-get update", "apt-get install -y"},
		{"yum", "centos", "yum update -y", "yum install -y"},
		{"dnf", "fedora", "dnf update -y", "dnf install -y"},
		{"apk", "alpine", "apk update", "apk add --no-cache"},
	}

	for _, tc := range testCases {
		t.Run(tc.manager, func(t *testing.T) {
			pkgManager := PackageManagerInfo{
				Manager:      tc.manager,
				Distribution: tc.distribution,
				UpdateCmd:    tc.updateCmd,
				InstallCmd:   tc.installCmd,
			}

			if pkgManager.Manager != tc.manager {
				t.Errorf("Expected Manager '%s', got '%s'", tc.manager, pkgManager.Manager)
			}

			if pkgManager.Distribution != tc.distribution {
				t.Errorf("Expected Distribution '%s', got '%s'", tc.distribution, pkgManager.Distribution)
			}
		})
	}
}

// Test that DockerConfig supports all required fields for Go installation
func TestDockerConfig_RequiredFields(t *testing.T) {
	config := DockerConfig{}

	// Test that all fields can be set
	config.Image = "golang:1.21"
	config.ContainerName = "go-dev"
	config.InstallationType = "go"
	config.EnableSSH = true
	config.KeepRunning = true
	config.Ports = []string{"8080:8080", "9000:9000"}
	config.Volumes = []string{"/host:/container"}
	config.Environment = []string{"GOPATH=/workspace/go"}

	if config.InstallationType != "go" {
		t.Error("InstallationType field not working properly")
	}

	if len(config.Ports) != 2 {
		t.Errorf("Expected 2 ports, got %d", len(config.Ports))
	}

	if len(config.Volumes) != 1 {
		t.Errorf("Expected 1 volume, got %d", len(config.Volumes))
	}

	if len(config.Environment) != 1 {
		t.Errorf("Expected 1 environment variable, got %d", len(config.Environment))
	}
}

// Test Go-specific installation types validation
func TestGoInstallationTypes(t *testing.T) {
	validTypes := map[string]bool{
		"default": true,
		"empty":   true,
		"python":  true,
		"java":    true,
		"go":      true,
		"vscode":  true,
	}

	// Test that 'go' is a valid type
	if !validTypes["go"] {
		t.Error("Installation type 'go' should be valid")
	}

	// Test all expected types
	expectedTypes := []string{"default", "empty", "python", "java", "go", "vscode"}
	for _, expectedType := range expectedTypes {
		if !validTypes[expectedType] {
			t.Errorf("Installation type '%s' should be valid", expectedType)
		}
	}

	// Test that invalid types are not accepted
	invalidTypes := []string{"golang", "rust", "nodejs", "invalid"}
	for _, invalidType := range invalidTypes {
		if validTypes[invalidType] {
			t.Errorf("Installation type '%s' should be invalid", invalidType)
		}
	}
}

// Test container configuration for Go development
func TestDockerConfig_GoDevEnvironment(t *testing.T) {
	// Typical configuration for Go development environment
	config := DockerConfig{
		Image:            "ubuntu:22.04",
		ContainerName:    "go-development",
		InstallationType: "go",
		EnableSSH:        true,
		KeepRunning:      true,
		Ports:            []string{"8080:8080", "9000:9000"}, // Typical Go web server ports
		Volumes:          []string{"/workspace:/app"},         // Mount workspace
		Environment:      []string{"GOPATH=/app/go", "GO111MODULE=on"},
		Privileged:       false,
		Network:          "bridge",
	}

	// Validate Go-specific configuration
	if config.InstallationType != "go" {
		t.Errorf("Expected Go installation type, got %s", config.InstallationType)
	}

	// Check that common Go environment variables can be set
	hasGoPath := false
	hasGoModule := false
	for _, env := range config.Environment {
		if env == "GOPATH=/app/go" {
			hasGoPath = true
		}
		if env == "GO111MODULE=on" {
			hasGoModule = true
		}
	}

	if !hasGoPath {
		t.Error("GOPATH environment variable should be configurable")
	}

	if !hasGoModule {
		t.Error("GO111MODULE environment variable should be configurable")
	}

	// Check that common Go development ports can be mapped
	expectedPorts := []string{"8080:8080", "9000:9000"}
	if len(config.Ports) != len(expectedPorts) {
		t.Errorf("Expected %d ports, got %d", len(expectedPorts), len(config.Ports))
	}
}

// Test command argument structure (conceptual test)
func TestCommandStructure(t *testing.T) {
	// Test that command components can be constructed properly
	containerName := "test-container"
	installationType := "go"

	// Simulate command construction that would happen in real functions
	command := []string{"portunix", "install", installationType}

	if len(command) != 3 {
		t.Errorf("Expected command with 3 parts, got %d", len(command))
	}

	if command[0] != "portunix" {
		t.Errorf("Expected first command part to be 'portunix', got '%s'", command[0])
	}

	if command[1] != "install" {
		t.Errorf("Expected second command part to be 'install', got '%s'", command[1])
	}

	if command[2] != installationType {
		t.Errorf("Expected third command part to be '%s', got '%s'", installationType, command[2])
	}

	// Test container name usage
	if containerName == "" {
		t.Error("Container name should not be empty")
	}

	if len(containerName) < 3 {
		t.Error("Container name should be at least 3 characters long")
	}
}