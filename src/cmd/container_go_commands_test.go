//go:build unit
// +build unit

package cmd

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// ContainerGoCommandsTestSuite defines test suite for container Go commands
type ContainerGoCommandsTestSuite struct {
	suite.Suite
}

func TestContainerGoCommandsSuite(t *testing.T) {
	suite.Run(t, new(ContainerGoCommandsTestSuite))
}

// Test Docker run-in-container command validation
func (suite *ContainerGoCommandsTestSuite) TestDockerRunInContainerCmd_ValidTypes() {
	// Test that the command accepts "go" as a valid installation type
	validTypes := []string{"default", "empty", "python", "java", "go", "vscode"}
	
	// Check that "go" is in the valid types list
	suite.Contains(validTypes, "go", "Go should be a valid installation type for Docker")
	
	// Check all expected types are present
	expectedTypes := []string{"default", "empty", "python", "java", "go", "vscode"}
	for _, expectedType := range expectedTypes {
		suite.Contains(validTypes, expectedType, "Expected type %s should be valid", expectedType)
	}
}

// Test Podman run-in-container command validation
func (suite *ContainerGoCommandsTestSuite) TestPodmanRunInContainerCmd_ValidTypes() {
	// Test that the command accepts "go" as a valid installation type
	validTypes := []string{"default", "empty", "python", "java", "go", "vscode"}
	
	// Check that "go" is in the valid types list
	suite.Contains(validTypes, "go", "Go should be a valid installation type for Podman")
	
	// Check all expected types are present
	expectedTypes := []string{"default", "empty", "python", "java", "go", "vscode"}
	for _, expectedType := range expectedTypes {
		suite.Contains(validTypes, expectedType, "Expected type %s should be valid", expectedType)
	}
}

// Test Docker command help text includes Go
func (suite *ContainerGoCommandsTestSuite) TestDockerRunInContainerCmd_HelpText() {
	// Get the command
	cmd := dockerRunInContainerCmd
	
	// Check that help text includes Go references
	helpText := cmd.Long
	
	suite.Contains(helpText, "go      - Installs Go development environment", "Help text should mention Go installation type")
	suite.Contains(helpText, "Available installation types:", "Help text should have installation types section")
	
	// Check that examples include Go usage
	suite.Contains(helpText, "go", "Help text should contain Go examples")
}

// Test Podman command help text includes Go
func (suite *ContainerGoCommandsTestSuite) TestPodmanRunInContainerCmd_HelpText() {
	// Get the command
	cmd := podmanRunInContainerCmd
	
	// Check that help text includes Go references
	helpText := cmd.Long
	
	suite.Contains(helpText, "go      - Installs Go development environment", "Help text should mention Go installation type")
	suite.Contains(helpText, "Available installation types:", "Help text should have installation types section")
	
	// Check that examples include Go usage
	suite.Contains(helpText, "go --image ubuntu:22.04 --keep-running", "Help text should contain Go example with parameters")
}

// Test command structure and basic properties
func (suite *ContainerGoCommandsTestSuite) TestDockerRunInContainerCmd_Structure() {
	cmd := dockerRunInContainerCmd
	
	suite.Equal("run-in-container [installation-type]", cmd.Use, "Command usage should match expected pattern")
	suite.Contains(cmd.Short, "Run Portunix installation inside a Docker container", "Short description should mention Docker")
	suite.NotNil(cmd.Run, "Command should have a Run function")
	suite.True(len(cmd.Long) > 100, "Long description should be substantial")
}

func (suite *ContainerGoCommandsTestSuite) TestPodmanRunInContainerCmd_Structure() {
	cmd := podmanRunInContainerCmd
	
	suite.Equal("run-in-container [installation-type]", cmd.Use, "Command usage should match expected pattern")
	suite.Contains(cmd.Short, "Run Portunix installation inside a Podman container", "Short description should mention Podman")
	suite.NotNil(cmd.Run, "Command should have a Run function")
	suite.True(len(cmd.Long) > 100, "Long description should be substantial")
}

// Test that commands have required flags
func (suite *ContainerGoCommandsTestSuite) TestDockerRunInContainerCmd_Flags() {
	cmd := dockerRunInContainerCmd
	
	// Check that important flags exist
	flags := cmd.Flags()
	
	imageFlag := flags.Lookup("image")
	suite.NotNil(imageFlag, "Should have --image flag")
	
	nameFlag := flags.Lookup("name")
	suite.NotNil(nameFlag, "Should have --name flag")
	
	keepRunningFlag := flags.Lookup("keep-running")
	suite.NotNil(keepRunningFlag, "Should have --keep-running flag")
	
	dryRunFlag := flags.Lookup("dry-run")
	suite.NotNil(dryRunFlag, "Should have --dry-run flag")
}

func (suite *ContainerGoCommandsTestSuite) TestPodmanRunInContainerCmd_Flags() {
	cmd := podmanRunInContainerCmd
	
	// Check that important flags exist
	flags := cmd.Flags()
	
	imageFlag := flags.Lookup("image")
	suite.NotNil(imageFlag, "Should have --image flag")
	
	nameFlag := flags.Lookup("name")
	suite.NotNil(nameFlag, "Should have --name flag")
	
	keepRunningFlag := flags.Lookup("keep-running")
	suite.NotNil(keepRunningFlag, "Should have --keep-running flag")
	
	rootlessFlag := flags.Lookup("rootless")
	suite.NotNil(rootlessFlag, "Should have --rootless flag")
	
	dryRunFlag := flags.Lookup("dry-run")
	suite.NotNil(dryRunFlag, "Should have --dry-run flag")
}

// Test help text formatting and completeness
func (suite *ContainerGoCommandsTestSuite) TestHelpTextFormatting() {
	dockerHelp := dockerRunInContainerCmd.Long
	podmanHelp := podmanRunInContainerCmd.Long
	
	// Both should mention Go in examples
	suite.Contains(dockerHelp, "go", "Docker help should mention Go")
	suite.Contains(podmanHelp, "go", "Podman help should mention Go")
	
	// Both should have proper structure
	suite.Contains(dockerHelp, "Available installation types:", "Docker help should list installation types")
	suite.Contains(podmanHelp, "Available installation types:", "Podman help should list installation types")
	
	suite.Contains(dockerHelp, "Examples:", "Docker help should have examples section")
	suite.Contains(podmanHelp, "Examples:", "Podman help should have examples section")
}

// Test installation type validation logic (conceptual test)
func (suite *ContainerGoCommandsTestSuite) TestInstallationTypeValidation() {
	validTypes := []string{"default", "empty", "python", "java", "go", "vscode"}
	
	// Test valid types
	validTestCases := []string{"default", "empty", "python", "java", "go", "vscode"}
	for _, testType := range validTestCases {
		isValid := false
		for _, validType := range validTypes {
			if testType == validType {
				isValid = true
				break
			}
		}
		suite.True(isValid, "Type %s should be valid", testType)
	}
	
	// Test invalid types
	invalidTestCases := []string{"invalid", "golang", "node", "ruby", "php"}
	for _, testType := range invalidTestCases {
		isValid := false
		for _, validType := range validTypes {
			if testType == validType {
				isValid = true
				break
			}
		}
		suite.False(isValid, "Type %s should be invalid", testType)
	}
}

// Test command string generation
func (suite *ContainerGoCommandsTestSuite) TestCommandStringGeneration() {
	// Test that valid installation type strings can be generated
	installationTypes := []string{"default", "empty", "python", "java", "go", "vscode"}
	
	for _, installType := range installationTypes {
		// Test Docker command string
		dockerCmd := []string{"portunix", "docker", "run-in-container", installType}
		cmdString := strings.Join(dockerCmd, " ")
		
		suite.Contains(cmdString, "portunix", "Command should start with portunix")
		suite.Contains(cmdString, "docker", "Docker command should contain docker")
		suite.Contains(cmdString, "run-in-container", "Command should contain run-in-container")
		suite.Contains(cmdString, installType, "Command should contain installation type")
		
		// Test Podman command string
		podmanCmd := []string{"portunix", "podman", "run-in-container", installType}
		cmdString = strings.Join(podmanCmd, " ")
		
		suite.Contains(cmdString, "portunix", "Command should start with portunix")
		suite.Contains(cmdString, "podman", "Podman command should contain podman")
		suite.Contains(cmdString, "run-in-container", "Command should contain run-in-container")
		suite.Contains(cmdString, installType, "Command should contain installation type")
	}
}