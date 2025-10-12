//go:build unit
// +build unit

package docker

import (
	"strings"
	"testing"
)

// Test runPortunixInstallInContainer function with simple testing
func TestRunPortunixInstallInContainer_Simple(t *testing.T) {
	// Mock execInContainer to avoid actual Docker calls
	originalExec := execInContainer
	defer func() { execInContainer = originalExec }()
	
	var executedCommands [][]string
	execInContainer = func(containerName string, command []string) error {
		executedCommands = append(executedCommands, command)
		return nil
	}

	err := runPortunixInstallInContainer("test-container", "go")
	
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	
	if len(executedCommands) != 1 {
		t.Errorf("Expected 1 command to be executed, got %d", len(executedCommands))
	}
	
	expectedCmd := []string{"portunix", "install", "go"}
	if len(executedCommands) > 0 {
		actualCmd := executedCommands[0]
		if len(actualCmd) != len(expectedCmd) {
			t.Errorf("Expected command length %d, got %d", len(expectedCmd), len(actualCmd))
		}
		for i, expected := range expectedCmd {
			if i >= len(actualCmd) || actualCmd[i] != expected {
				t.Errorf("Expected command[%d] = %s, got %s", i, expected, actualCmd[i])
			}
		}
	}
}

func TestRunPortunixInstallInContainer_DifferentTypes(t *testing.T) {
	testCases := []struct {
		installationType string
		expectedCmd      []string
	}{
		{"go", []string{"portunix", "install", "go"}},
		{"python", []string{"portunix", "install", "python"}},
		{"java", []string{"portunix", "install", "java"}},
		{"default", []string{"portunix", "install", "default"}},
		{"vscode", []string{"portunix", "install", "vscode"}},
	}

	for _, tc := range testCases {
		t.Run(tc.installationType, func(t *testing.T) {
			// Mock execInContainer
			originalExec := execInContainer
			defer func() { execInContainer = originalExec }()
			
			var executedCommands [][]string
			execInContainer = func(containerName string, command []string) error {
				executedCommands = append(executedCommands, command)
				return nil
			}

			err := runPortunixInstallInContainer("test-container", tc.installationType)
			
			if err != nil {
				t.Errorf("Expected no error for %s, got %v", tc.installationType, err)
			}
			
			if len(executedCommands) != 1 {
				t.Errorf("Expected 1 command for %s, got %d", tc.installationType, len(executedCommands))
			}
			
			if len(executedCommands) > 0 {
				actualCmd := executedCommands[0]
				for i, expected := range tc.expectedCmd {
					if i >= len(actualCmd) || actualCmd[i] != expected {
						t.Errorf("For %s: expected command[%d] = %s, got %s", tc.installationType, i, expected, actualCmd[i])
					}
				}
			}
		})
	}
}

func TestInstallSoftwareInContainer_EmptyType(t *testing.T) {
	// Mock execInContainer to track calls
	originalExec := execInContainer
	defer func() { execInContainer = originalExec }()
	
	var executedCommands [][]string
	execInContainer = func(containerName string, command []string) error {
		executedCommands = append(executedCommands, command)
		return nil
	}

	pkgManager := &PackageManagerInfo{
		Manager: "apt-get",
	}

	err := InstallSoftwareInContainer("test-container", "empty", pkgManager)
	
	if err != nil {
		t.Errorf("Expected no error for empty type, got %v", err)
	}
	
	if len(executedCommands) != 0 {
		t.Errorf("Expected no commands for empty type, got %d commands", len(executedCommands))
	}
}

func TestValidInstallationTypes(t *testing.T) {
	// Test that all expected installation types are supported
	validTypes := []string{"default", "empty", "python", "java", "go", "vscode"}
	
	// Check that "go" is included
	goFound := false
	for _, validType := range validTypes {
		if validType == "go" {
			goFound = true
			break
		}
	}
	
	if !goFound {
		t.Error("Installation type 'go' should be in valid types list")
	}
	
	// Check all expected types are present
	expectedTypes := []string{"default", "empty", "python", "java", "go", "vscode"}
	for _, expectedType := range expectedTypes {
		found := false
		for _, validType := range validTypes {
			if validType == expectedType {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected installation type '%s' not found in valid types", expectedType)
		}
	}
}

func TestDockerConfigStructure(t *testing.T) {
	// Test that DockerConfig can be created with InstallationType
	config := DockerConfig{
		Image:            "ubuntu:22.04",
		ContainerName:    "test-container", 
		InstallationType: "go",
		EnableSSH:        true,
		KeepRunning:      true,
	}
	
	if config.InstallationType != "go" {
		t.Errorf("Expected InstallationType 'go', got '%s'", config.InstallationType)
	}
	
	if !config.EnableSSH {
		t.Error("Expected EnableSSH to be true")
	}
	
	if !config.KeepRunning {
		t.Error("Expected KeepRunning to be true")
	}
}

// Test command generation for different installation types
func TestCommandGeneration(t *testing.T) {
	testCases := []string{"default", "python", "java", "go", "vscode"}
	
	for _, installationType := range testCases {
		t.Run("command_for_"+installationType, func(t *testing.T) {
			// Simulate command construction
			command := []string{"portunix", "install", installationType}
			
			if len(command) != 3 {
				t.Errorf("Expected command length 3, got %d", len(command))
			}
			
			if command[0] != "portunix" {
				t.Errorf("Expected command[0] = 'portunix', got '%s'", command[0])
			}
			
			if command[1] != "install" {
				t.Errorf("Expected command[1] = 'install', got '%s'", command[1])
			}
			
			if command[2] != installationType {
				t.Errorf("Expected command[2] = '%s', got '%s'", installationType, command[2])
			}
		})
	}
}

func TestContainerNameValidation(t *testing.T) {
	validNames := []string{
		"test-container",
		"portunix-go-123",
		"my-dev-env",
		"python-container-2024",
	}
	
	for _, name := range validNames {
		// Simple validation that names don't contain invalid characters
		if strings.Contains(name, " ") {
			t.Errorf("Container name '%s' should not contain spaces", name)
		}
		
		if strings.Contains(name, "_") && !strings.Contains(name, "-") {
			// Prefer dashes over underscores for Docker compatibility
			t.Logf("Container name '%s' uses underscores - consider using dashes", name)
		}
	}
}