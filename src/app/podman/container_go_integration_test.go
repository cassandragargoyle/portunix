//go:build integration
// +build integration

package podman

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	"portunix.ai/app/container"
)

// ContainerGoIntegrationTestSuite defines integration test suite for Podman Go container installation
type ContainerGoIntegrationTestSuite struct {
	suite.Suite
	ctx               context.Context
	testContainerName string
}

func TestContainerGoIntegrationSuite(t *testing.T) {
	// Skip if Podman is not available - using new API
	if !container.HasPodman() {
		t.Skip("Podman not available - skipping integration tests")
	}
	
	suite.Run(t, new(ContainerGoIntegrationTestSuite))
}

func (suite *ContainerGoIntegrationTestSuite) SetupTest() {
	suite.ctx = context.Background()
	suite.testContainerName = fmt.Sprintf("portunix-go-podman-integration-test-%d", time.Now().Unix())
}

func (suite *ContainerGoIntegrationTestSuite) TearDownTest() {
	// Clean up test container if it exists
	if suite.testContainerName != "" {
		cleanupPodmanContainer(suite.testContainerName)
	}
}

// Test complete Go installation workflow in Podman container
func (suite *ContainerGoIntegrationTestSuite) TestPodmanGoInstallationWorkflow() {
	// Compile current Portunix binary for testing
	suite.T().Log("Compiling Portunix binary for Podman container test...")
	err := compilePortunixBinary()
	suite.Require().NoError(err, "Failed to compile Portunix binary")

	// Configure Podman container
	config := PodmanConfig{
		Image:            "ubuntu:22.04",
		ContainerName:    suite.testContainerName,
		InstallationType: "go",
		EnableSSH:        false, // Disable SSH for faster test
		KeepRunning:      true,
		CacheShared:      false, // Disable cache for clean test
		Rootless:         true,  // Use rootless mode
	}

	// Create and start container
	suite.T().Log("Creating Podman container...")
	args := buildPodmanRunArgs(config)
	cmd := exec.Command("podman", args...)
	err = cmd.Start()
	suite.Require().NoError(err, "Failed to start Podman container")

	// Wait for container to be ready
	suite.T().Log("Waiting for container to be ready...")
	err = waitForContainer(suite.testContainerName)
	suite.Require().NoError(err, "Container failed to start properly")

	// Test that container is running
	suite.T().Log("Verifying container is running...")
	isRunning, err := isPodmanContainerRunning(suite.testContainerName)
	suite.Require().NoError(err)
	suite.Require().True(isRunning, "Container should be running")

	// Copy Portunix binary to container
	suite.T().Log("Copying Portunix binary to container...")
	err = copyPortunixToPodmanContainer(suite.testContainerName)
	suite.Require().NoError(err, "Failed to copy Portunix binary to container")

	// Verify Portunix binary exists and is executable in container
	suite.T().Log("Verifying Portunix binary in container...")
	checkCmd := []string{"ls", "-la", "/usr/local/bin/portunix"}
	err = execInPodmanContainer(suite.testContainerName, checkCmd)
	suite.Require().NoError(err, "Portunix binary not found in container")

	// Test Portunix version command inside container
	suite.T().Log("Testing Portunix version command in container...")
	versionCmd := []string{"portunix", "version"}
	err = execInPodmanContainer(suite.testContainerName, versionCmd)
	suite.Require().NoError(err, "Portunix version command failed in container")

	// Install Go using Portunix inside container
	suite.T().Log("Installing Go development environment in container...")
	err = runPortunixInstallInPodmanContainer(suite.testContainerName, "go")
	suite.Require().NoError(err, "Failed to install Go in container")

	// Verify Go installation
	suite.T().Log("Verifying Go installation in container...")
	goVersionCmd := []string{"sh", "-c", "source /etc/profile && go version"}
	err = execInPodmanContainer(suite.testContainerName, goVersionCmd)
	suite.Require().NoError(err, "Go version command failed - Go not properly installed")

	// Test Go compilation in container
	suite.T().Log("Testing Go compilation in container...")
	testGoCmd := []string{"sh", "-c", `
		source /etc/profile &&
		cd /tmp &&
		echo 'package main; import "fmt"; func main() { fmt.Println("Hello from Go in Podman!") }' > hello.go &&
		go run hello.go
	`}
	err = execInPodmanContainer(suite.testContainerName, testGoCmd)
	suite.Require().NoError(err, "Go compilation test failed")

	suite.T().Log("✅ Podman Go installation integration test completed successfully")
}

// Test installSoftwareInPodmanContainer integration
func (suite *ContainerGoIntegrationTestSuite) TestInstallSoftwareInPodmanContainer_Integration() {
	suite.testContainerName = fmt.Sprintf("portunix-podman-install-integration-test-%d", time.Now().Unix())
	
	// Compile Portunix binary
	err := compilePortunixBinary()
	suite.Require().NoError(err)

	// Start minimal Ubuntu container
	cmd := exec.Command("podman", "run", "-d", "-it", "--name", suite.testContainerName, "ubuntu:22.04", "sleep", "infinity")
	err = cmd.Run()
	suite.Require().NoError(err)

	// Wait for container
	err = waitForContainer(suite.testContainerName)
	suite.Require().NoError(err)

	// Create package manager info (would normally be detected)
	pkgManager := &PackageManagerInfo{
		Manager:      "apt-get",
		UpdateCmd:    "apt-get update",
		InstallCmd:   "apt-get install -y",
		Distribution: "ubuntu",
	}

	// Test installSoftwareInPodmanContainer with Go type
	suite.T().Log("Testing installSoftwareInPodmanContainer with Go...")
	err = installSoftwareInPodmanContainer(suite.testContainerName, "go", pkgManager)
	suite.Require().NoError(err, "installSoftwareInPodmanContainer failed")

	// Verify Go is installed
	goVersionCmd := []string{"sh", "-c", "source /etc/profile && go version"}
	err = execInPodmanContainer(suite.testContainerName, goVersionCmd)
	suite.Require().NoError(err, "Go verification failed after installSoftwareInPodmanContainer")

	suite.T().Log("✅ installSoftwareInPodmanContainer integration test completed successfully")
}

// Test Go installation with Alpine (different package manager)
func (suite *ContainerGoIntegrationTestSuite) TestGoInstallationAlpine() {
	suite.testContainerName = fmt.Sprintf("portunix-go-podman-alpine-test-%d", time.Now().Unix())
	
	// Compile Portunix binary
	err := compilePortunixBinary()
	suite.Require().NoError(err)

	config := PodmanConfig{
		Image:            "alpine:3.18",
		ContainerName:    suite.testContainerName,
		InstallationType: "go",
		EnableSSH:        false,
		KeepRunning:      true,
		CacheShared:      false,
		Rootless:         true,
	}

	// Start container
	args := buildPodmanRunArgs(config)
	cmd := exec.Command("podman", args...)
	err = cmd.Start()
	suite.Require().NoError(err)

	// Wait and copy binary
	err = waitForContainer(suite.testContainerName)
	suite.Require().NoError(err)

	err = copyPortunixToPodmanContainer(suite.testContainerName)
	suite.Require().NoError(err)

	// Install Go
	err = runPortunixInstallInPodmanContainer(suite.testContainerName, "go")
	suite.Require().NoError(err, "Go installation failed on Alpine with Podman")

	// Verify installation
	goVersionCmd := []string{"sh", "-c", "source /etc/profile && go version"}
	err = execInPodmanContainer(suite.testContainerName, goVersionCmd)
	suite.Require().NoError(err, "Go not working on Alpine with Podman")

	suite.T().Log("✅ Podman Alpine Go installation test completed successfully")
}

// Helper functions

// isPodmanAvailable is now deprecated - use container.HasPodman() instead
// Keeping for backward compatibility
func isPodmanAvailable() bool {
	return container.HasPodman()
}

func compilePortunixBinary() error {
	// Change to project root and compile
	cmd := exec.Command("go", "build", "-o", "portunix")
	cmd.Dir = "/media/zdenek/DevDisk/DEV/CassandraGargoyle/portunix/portunix"
	return cmd.Run()
}

func isPodmanContainerRunning(containerName string) (bool, error) {
	cmd := exec.Command("podman", "ps", "--filter", fmt.Sprintf("name=%s", containerName), "--format", "{{.Names}}")
	output, err := cmd.Output()
	if err != nil {
		return false, err
	}
	return strings.Contains(string(output), containerName), nil
}

func cleanupPodmanContainer(containerName string) {
	// Stop and remove test container
	exec.Command("podman", "stop", containerName).Run()
	exec.Command("podman", "rm", containerName).Run()
}