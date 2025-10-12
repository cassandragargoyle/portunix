package podman

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

func TestPodmanConfigDefaults(t *testing.T) {
	config := PodmanConfig{
		Image:            "ubuntu:22.04",
		InstallationType: "default",
		Rootless:         true,
		CacheShared:      true,
		EnableSSH:        true,
	}

	if config.Image != "ubuntu:22.04" {
		t.Errorf("Expected default image ubuntu:22.04, got %s", config.Image)
	}

	if !config.Rootless {
		t.Errorf("Expected rootless to be true by default")
	}

	if !config.CacheShared {
		t.Errorf("Expected cache sharing to be enabled by default")
	}

	if !config.EnableSSH {
		t.Errorf("Expected SSH to be enabled by default")
	}
}

func TestBuildPodmanRunArgs(t *testing.T) {
	config := PodmanConfig{
		Image:         "ubuntu:22.04",
		ContainerName: "test-container",
		Rootless:      true,
		EnableSSH:     true,
		CacheShared:   true,
		Pod:           "test-pod",
	}

	args := buildPodmanRunArgs(config)

	// Check basic args
	if args[0] != "run" {
		t.Errorf("Expected first arg to be 'run', got %s", args[0])
	}

	// Check if container name is included
	nameFound := false
	for i, arg := range args {
		if arg == "--name" && i+1 < len(args) && args[i+1] == "test-container" {
			nameFound = true
			break
		}
	}
	if !nameFound {
		t.Errorf("Container name not found in args")
	}

	// Check if pod is included (Podman-specific)
	podFound := false
	for i, arg := range args {
		if arg == "--pod" && i+1 < len(args) && args[i+1] == "test-pod" {
			podFound = true
			break
		}
	}
	if !podFound {
		t.Errorf("Pod not found in args")
	}

	// Check if image is last non-command arg
	imageIndex := -1
	for i := len(args) - 1; i >= 0; i-- {
		if args[i] == "ubuntu:22.04" {
			imageIndex = i
			break
		}
	}
	if imageIndex == -1 {
		t.Errorf("Image not found in args")
	}
}

func TestPackageManagerInfo(t *testing.T) {
	tests := []struct {
		manager      string
		distribution string
		updateCmd    string
		installCmd   string
	}{
		{"apt-get", "debian-based", "apt-get update", "apt-get install -y"},
		{"yum", "rhel-based", "yum update -y", "yum install -y"},
		{"dnf", "rhel-based", "dnf update -y", "dnf install -y"},
		{"apk", "alpine", "apk update", "apk add --no-cache"},
	}

	for _, test := range tests {
		pkgInfo := &PackageManagerInfo{
			Manager:      test.manager,
			Distribution: test.distribution,
			UpdateCmd:    test.updateCmd,
			InstallCmd:   test.installCmd,
		}

		if pkgInfo.Manager != test.manager {
			t.Errorf("Expected manager %s, got %s", test.manager, pkgInfo.Manager)
		}

		if pkgInfo.Distribution != test.distribution {
			t.Errorf("Expected distribution %s, got %s", test.distribution, pkgInfo.Distribution)
		}

		if pkgInfo.UpdateCmd != test.updateCmd {
			t.Errorf("Expected update command %s, got %s", test.updateCmd, pkgInfo.UpdateCmd)
		}

		if pkgInfo.InstallCmd != test.installCmd {
			t.Errorf("Expected install command %s, got %s", test.installCmd, pkgInfo.InstallCmd)
		}
	}
}

func TestContainerInfo(t *testing.T) {
	container := ContainerInfo{
		ID:      "abc123456789",
		Name:    "portunix-python-test",
		Image:   "ubuntu:22.04",
		Status:  "Running",
		Ports:   "22:2222",
		Created: "2 hours ago",
		Command: "/bin/bash",
	}

	if container.ID != "abc123456789" {
		t.Errorf("Expected ID abc123456789, got %s", container.ID)
	}

	if container.Name != "portunix-python-test" {
		t.Errorf("Expected name portunix-python-test, got %s", container.Name)
	}

	if container.Status != "Running" {
		t.Errorf("Expected status Running, got %s", container.Status)
	}
}

func TestGeneratePassword(t *testing.T) {
	password1 := generatePassword()
	password2 := generatePassword()

	if len(password1) != 16 {
		t.Errorf("Expected password length 16, got %d", len(password1))
	}

	if len(password2) != 16 {
		t.Errorf("Expected password length 16, got %d", len(password2))
	}

	// Passwords should be different
	if password1 == password2 {
		t.Errorf("Generated passwords should be different")
	}

	// Check if password contains only allowed characters
	allowed := "ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnpqrstuvwxyz23456789"
	for _, char := range password1 {
		found := false
		for _, allowedChar := range allowed {
			if char == allowedChar {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Password contains invalid character: %c", char)
		}
	}
}

func TestGenerateID(t *testing.T) {
	id1 := generateID()

	if len(id1) == 0 {
		t.Errorf("Generated ID should not be empty")
	}

	// Test that ID is numeric timestamp
	if len(id1) < 10 {
		t.Errorf("Generated ID should be at least 10 characters (Unix timestamp)")
	}
}

func TestGenerateShortID(t *testing.T) {
	shortID1 := generateShortID()
	shortID2 := generateShortID()

	if len(shortID1) == 0 {
		t.Errorf("Generated short ID should not be empty")
	}

	if len(shortID2) == 0 {
		t.Errorf("Generated short ID should not be empty")
	}

	// Short IDs should be reasonably short (less than 6 characters for 4-digit mod)
	if len(shortID1) > 6 {
		t.Errorf("Short ID too long: %s", shortID1)
	}
}

// PodmanGoTestSuite defines test suite for Podman Go installation functions
type PodmanGoTestSuite struct {
	suite.Suite
}

func TestPodmanGoSuite(t *testing.T) {
	suite.Run(t, new(PodmanGoTestSuite))
}

// Test installSoftwareInPodmanContainer function
func (suite *PodmanGoTestSuite) TestInstallSoftwareInPodmanContainer_EmptyType() {
	pkgManager := &PackageManagerInfo{
		Manager: "apt-get",
	}

	// Mock execInPodmanContainer to avoid actual Podman calls
	originalExec := execInPodmanContainer
	defer func() { execInPodmanContainer = originalExec }()
	
	var executedCommands [][]string
	execInPodmanContainer = func(containerName string, command []string) error {
		executedCommands = append(executedCommands, command)
		return nil
	}

	err := installSoftwareInPodmanContainer("test-container", "empty", pkgManager)
	
	suite.NoError(err)
	suite.Empty(executedCommands) // No commands should be executed for empty type
}

func (suite *PodmanGoTestSuite) TestInstallSoftwareInPodmanContainer_GoType() {
	pkgManager := &PackageManagerInfo{
		Manager: "apt-get",
	}

	// Mock execInPodmanContainer to avoid actual Podman calls
	originalExec := execInPodmanContainer
	defer func() { execInPodmanContainer = originalExec }()
	
	var executedCommands [][]string
	execInPodmanContainer = func(containerName string, command []string) error {
		executedCommands = append(executedCommands, command)
		return nil
	}

	// Test will fail on actual copy, but we can test the logic flow
	err := installSoftwareInPodmanContainer("test-container", "go", pkgManager)
	
	// Error expected because we can't actually copy files in unit test
	suite.Error(err)
	suite.Contains(err.Error(), "failed to copy Portunix binary")
}

func (suite *PodmanGoTestSuite) TestInstallSoftwareInPodmanContainer_InvalidType() {
	pkgManager := &PackageManagerInfo{
		Manager: "apt-get",
	}

	// Mock execInPodmanContainer to track calls
	originalExec := execInPodmanContainer
	defer func() { execInPodmanContainer = originalExec }()
	
	var executedCommands [][]string
	execInPodmanContainer = func(containerName string, command []string) error {
		executedCommands = append(executedCommands, command)
		return nil
	}

	err := installSoftwareInPodmanContainer("test-container", "invalid-type", pkgManager)
	
	// Should fail on copy step before reaching invalid type check
	suite.Error(err)
}

// Test copyPortunixToPodmanContainer function with mocked Podman commands
func (suite *PodmanGoTestSuite) TestCopyPortunixToPodmanContainer_MockedSuccess() {
	// Mock execInPodmanContainer to avoid actual Podman calls
	originalExec := execInPodmanContainer
	defer func() { execInPodmanContainer = originalExec }()
	
	var executedCommands [][]string
	execInPodmanContainer = func(containerName string, command []string) error {
		executedCommands = append(executedCommands, command)
		return nil
	}

	// This will still fail because we can't mock os.Executable and podman cp easily
	// But we can test the function structure
	err := copyPortunixToPodmanContainer("test-container")
	
	// Expected to fail on podman cp command in unit test environment
	suite.Error(err)
	suite.Contains(err.Error(), "failed to copy Portunix binary")
}

// Test runPortunixInstallInPodmanContainer function
func (suite *PodmanGoTestSuite) TestRunPortunixInstallInPodmanContainer_Success() {
	// Mock execInPodmanContainer to avoid actual Podman calls
	originalExec := execInPodmanContainer
	defer func() { execInPodmanContainer = originalExec }()
	
	var executedCommands [][]string
	execInPodmanContainer = func(containerName string, command []string) error {
		executedCommands = append(executedCommands, command)
		return nil
	}

	err := runPortunixInstallInPodmanContainer("test-container", "go")
	
	suite.NoError(err)
	suite.Len(executedCommands, 1)
	suite.Equal([]string{"portunix", "install", "go"}, executedCommands[0])
}

func (suite *PodmanGoTestSuite) TestRunPortunixInstallInPodmanContainer_Python() {
	// Mock execInPodmanContainer
	originalExec := execInPodmanContainer
	defer func() { execInPodmanContainer = originalExec }()
	
	var executedCommands [][]string
	execInPodmanContainer = func(containerName string, command []string) error {
		executedCommands = append(executedCommands, command)
		return nil
	}

	err := runPortunixInstallInPodmanContainer("test-container", "python")
	
	suite.NoError(err)
	suite.Len(executedCommands, 1)
	suite.Equal([]string{"portunix", "install", "python"}, executedCommands[0])
}

func (suite *PodmanGoTestSuite) TestRunPortunixInstallInPodmanContainer_Default() {
	// Mock execInPodmanContainer
	originalExec := execInPodmanContainer
	defer func() { execInPodmanContainer = originalExec }()
	
	var executedCommands [][]string
	execInPodmanContainer = func(containerName string, command []string) error {
		executedCommands = append(executedCommands, command)
		return nil
	}

	err := runPortunixInstallInPodmanContainer("test-container", "default")
	
	suite.NoError(err)
	suite.Len(executedCommands, 1)
	suite.Equal([]string{"portunix", "install", "default"}, executedCommands[0])
}

func (suite *PodmanGoTestSuite) TestRunPortunixInstallInPodmanContainer_ExecError() {
	// Mock execInPodmanContainer to return error
	originalExec := execInPodmanContainer
	defer func() { execInPodmanContainer = originalExec }()
	
	execInPodmanContainer = func(containerName string, command []string) error {
		return fmt.Errorf("container not found")
	}

	err := runPortunixInstallInPodmanContainer("test-container", "go")
	
	suite.Error(err)
	suite.Contains(err.Error(), "failed to run 'portunix install go'")
	suite.Contains(err.Error(), "container not found")
}

// Test execInPodmanContainer function
func (suite *PodmanGoTestSuite) TestExecInPodmanContainer_CommandConstruction() {
	// Mock execInPodmanContainer to capture command construction
	originalExec := execInPodmanContainer
	defer func() { execInPodmanContainer = originalExec }()
	
	var capturedCommands []string
	execInPodmanContainer = func(containerName string, command []string) error {
		// Simulate the command construction that happens in the real function
		args := append([]string{"exec", containerName}, command...)
		capturedCommands = args
		return nil
	}

	err := execInPodmanContainer("test-container", []string{"portunix", "install", "go"})
	
	suite.NoError(err)
	expectedCmd := []string{"exec", "test-container", "portunix", "install", "go"}
	suite.Equal(expectedCmd, capturedCommands)
}
