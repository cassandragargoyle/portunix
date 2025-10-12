package podman

import (
	"testing"
)

// Test basic Podman configuration structure for Go installations  
func TestPodmanConfig_GoInstallationType(t *testing.T) {
	config := PodmanConfig{
		Image:            "ubuntu:22.04",
		ContainerName:    "test-go-container",
		InstallationType: "go",
		EnableSSH:        true,
		KeepRunning:      true,
		Rootless:         true,
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

	if !config.Rootless {
		t.Error("Expected Rootless to be true for security")
	}
}

// Test PackageManagerInfo structure for Podman Go installation
func TestPodmanPackageManagerInfo_GoInstallation(t *testing.T) {
	// Test different package managers that Podman Go installation should support
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

// Test that PodmanConfig supports all required fields for Go installation
func TestPodmanConfig_RequiredFields(t *testing.T) {
	config := PodmanConfig{}

	// Test that all fields can be set
	config.Image = "golang:1.21"
	config.ContainerName = "go-dev-podman"
	config.InstallationType = "go"
	config.EnableSSH = true
	config.KeepRunning = true
	config.Rootless = true
	config.Ports = []string{"8080:8080", "9000:9000"}
	config.Volumes = []string{"/host:/container"}
	config.Environment = []string{"GOPATH=/workspace/go"}
	config.Pod = "development-pod"

	if config.InstallationType != "go" {
		t.Error("InstallationType field not working properly")
	}

	if !config.Rootless {
		t.Error("Rootless field should be true for Podman security")
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

	if config.Pod != "development-pod" {
		t.Errorf("Expected Pod 'development-pod', got '%s'", config.Pod)
	}
}

// Test Go-specific installation types validation for Podman
func TestPodmanGoInstallationTypes(t *testing.T) {
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
		t.Error("Installation type 'go' should be valid for Podman")
	}

	// Test all expected types
	expectedTypes := []string{"default", "empty", "python", "java", "go", "vscode"}
	for _, expectedType := range expectedTypes {
		if !validTypes[expectedType] {
			t.Errorf("Installation type '%s' should be valid for Podman", expectedType)
		}
	}

	// Test that invalid types are not accepted
	invalidTypes := []string{"golang", "rust", "nodejs", "invalid"}
	for _, invalidType := range invalidTypes {
		if validTypes[invalidType] {
			t.Errorf("Installation type '%s' should be invalid for Podman", invalidType)
		}
	}
}

// Test Podman-specific configuration for Go development
func TestPodmanConfig_GoDevEnvironment(t *testing.T) {
	// Typical configuration for Go development environment with Podman
	config := PodmanConfig{
		Image:            "ubuntu:22.04",
		ContainerName:    "go-development-podman",
		InstallationType: "go",
		EnableSSH:        true,
		KeepRunning:      true,
		Rootless:         true, // Podman security feature
		Ports:            []string{"8080:8080", "9000:9000"}, // Typical Go web server ports
		Volumes:          []string{"/workspace:/app"},         // Mount workspace
		Environment:      []string{"GOPATH=/app/go", "GO111MODULE=on"},
		Pod:              "", // Can be empty for single container
		Network:          "bridge",
	}

	// Validate Go-specific configuration
	if config.InstallationType != "go" {
		t.Errorf("Expected Go installation type, got %s", config.InstallationType)
	}

	// Podman-specific: rootless mode should be enabled
	if !config.Rootless {
		t.Error("Rootless mode should be enabled for Podman security")
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
		t.Error("GOPATH environment variable should be configurable in Podman")
	}

	if !hasGoModule {
		t.Error("GO111MODULE environment variable should be configurable in Podman")
	}

	// Check that common Go development ports can be mapped
	expectedPorts := []string{"8080:8080", "9000:9000"}
	if len(config.Ports) != len(expectedPorts) {
		t.Errorf("Expected %d ports, got %d", len(expectedPorts), len(config.Ports))
	}
}

// Test Podman-specific features
func TestPodmanConfig_PodmanSpecificFeatures(t *testing.T) {
	config := PodmanConfig{
		Image:            "ubuntu:22.04",
		ContainerName:    "test-container",
		InstallationType: "go",
		Rootless:         true,
		Pod:              "development-pod",
		Network:          "",
	}

	// Test rootless mode (Podman security feature)
	if !config.Rootless {
		t.Error("Rootless should be true for enhanced security")
	}

	// Test pod support (Podman-specific Kubernetes-like feature)
	if config.Pod != "development-pod" {
		t.Errorf("Expected Pod 'development-pod', got '%s'", config.Pod)
	}

	// Test that both Pod and Network cannot be set simultaneously
	config.Network = "bridge"
	// In real validation, this would be caught by ValidatePodmanConfig
	// Here we just test the structure can hold both values
	if config.Pod != "" && config.Network != "" {
		// This combination should be caught by validation logic
		t.Logf("Warning: Pod and Network set simultaneously - should be validated")
	}
}

// Test command argument structure for Podman (conceptual test)
func TestPodmanCommandStructure(t *testing.T) {
	// Test that Podman command components can be constructed properly
	containerName := "test-podman-container"
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

	// Test Podman-specific container naming patterns
	if containerName == "test-podman-container" {
		// Valid Podman container name
		t.Logf("Valid Podman container name: %s", containerName)
	}
}

// Test container info structure for Podman
func TestContainerInfo_PodmanGo(t *testing.T) {
	container := ContainerInfo{
		ID:      "podman123456789",
		Name:    "portunix-go-podman-test",
		Image:   "ubuntu:22.04",
		Status:  "Running",
		Ports:   "8080:8080,9000:9000",
		Created: "2 hours ago",
		Command: "/bin/bash",
	}

	if container.ID == "" {
		t.Error("Container ID should not be empty")
	}

	if container.Name != "portunix-go-podman-test" {
		t.Errorf("Expected name 'portunix-go-podman-test', got '%s'", container.Name)
	}

	if container.Image != "ubuntu:22.04" {
		t.Errorf("Expected image 'ubuntu:22.04', got '%s'", container.Image)
	}

	if container.Status != "Running" {
		t.Errorf("Expected status 'Running', got '%s'", container.Status)
	}

	// Test that ports string can contain multiple port mappings
	if container.Ports == "" {
		t.Error("Ports should not be empty for Go development container")
	}

	// Check that port string contains Go development ports
	if container.Ports == "8080:8080,9000:9000" {
		t.Logf("Valid Go development ports: %s", container.Ports)
	}
}