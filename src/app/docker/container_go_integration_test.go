//go:build integration
// +build integration

package docker

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// ContainerGoIntegrationTestSuite defines integration test suite for Go container installation
type ContainerGoIntegrationTestSuite struct {
	suite.Suite
	ctx               context.Context
	testContainerName string
}

func TestContainerGoIntegrationSuite(t *testing.T) {
	// Skip if Docker is not available
	if !isDockerAvailable() {
		t.Skip("Docker not available - skipping integration tests")
	}
	
	suite.Run(t, new(ContainerGoIntegrationTestSuite))
}

func (suite *ContainerGoIntegrationTestSuite) SetupTest() {
	suite.ctx = context.Background()
	suite.testContainerName = fmt.Sprintf("portunix-go-integration-test-%d", time.Now().Unix())
}

func (suite *ContainerGoIntegrationTestSuite) TearDownTest() {
	// Clean up test container if it exists
	if suite.testContainerName != "" {
		cleanupContainer(suite.testContainerName)
	}
}

// Test complete Go installation workflow in Docker container
func (suite *ContainerGoIntegrationTestSuite) TestDockerGoInstallationWorkflow() {
	// Compile current Portunix binary for testing
	suite.T().Log("Compiling Portunix binary for container test...")
	err := compilePortunixBinary()
	suite.Require().NoError(err, "Failed to compile Portunix binary")

	// Configure Docker container
	config := DockerConfig{
		Image:            "ubuntu:22.04",
		ContainerName:    suite.testContainerName,
		InstallationType: "go",
		EnableSSH:        false, // Disable SSH for faster test
		KeepRunning:      true,
		CacheShared:      false, // Disable cache for clean test
	}

	// Create and start container
	suite.T().Log("Creating Docker container...")
	args := buildDockerRunArgs(config)
	cmd := exec.Command("docker", args...)
	err = cmd.Start()
	suite.Require().NoError(err, "Failed to start Docker container")

	// Wait for container to be ready
	suite.T().Log("Waiting for container to be ready...")
	err = waitForContainer(suite.testContainerName)
	suite.Require().NoError(err, "Container failed to start properly")

	// Test that container is running
	suite.T().Log("Verifying container is running...")
	isRunning, err := isContainerRunning(suite.testContainerName)
	suite.Require().NoError(err)
	suite.Require().True(isRunning, "Container should be running")

	// Copy Portunix binary to container
	suite.T().Log("Copying Portunix binary to container...")
	err = copyPortunixToContainer(suite.testContainerName)
	suite.Require().NoError(err, "Failed to copy Portunix binary to container")

	// Verify Portunix binary exists and is executable in container
	suite.T().Log("Verifying Portunix binary in container...")
	checkCmd := []string{"ls", "-la", "/usr/local/bin/portunix"}
	err = execInContainer(suite.testContainerName, checkCmd)
	suite.Require().NoError(err, "Portunix binary not found in container")

	// Test Portunix version command inside container
	suite.T().Log("Testing Portunix version command in container...")
	versionCmd := []string{"portunix", "version"}
	err = execInContainer(suite.testContainerName, versionCmd)
	suite.Require().NoError(err, "Portunix version command failed in container")

	// Install Go using Portunix inside container
	suite.T().Log("Installing Go development environment in container...")
	err = runPortunixInstallInContainer(suite.testContainerName, "go")
	suite.Require().NoError(err, "Failed to install Go in container")

	// Verify Go installation
	suite.T().Log("Verifying Go installation in container...")
	goVersionCmd := []string{"sh", "-c", "source /etc/profile && go version"}
	err = execInContainer(suite.testContainerName, goVersionCmd)
	suite.Require().NoError(err, "Go version command failed - Go not properly installed")

	// Test Go compilation in container
	suite.T().Log("Testing Go compilation in container...")
	testGoCmd := []string{"sh", "-c", `
		source /etc/profile &&
		cd /tmp &&
		echo 'package main; import "fmt"; func main() { fmt.Println("Hello from Go!") }' > hello.go &&
		go run hello.go
	`}
	err = execInContainer(suite.testContainerName, testGoCmd)
	suite.Require().NoError(err, "Go compilation test failed")

	suite.T().Log("✅ Docker Go installation integration test completed successfully")
}

// Test Go installation with different package managers
func (suite *ContainerGoIntegrationTestSuite) TestGoInstallationAlpine() {
	// Test with Alpine Linux (apk package manager)
	suite.testContainerName = fmt.Sprintf("portunix-go-alpine-test-%d", time.Now().Unix())
	
	// Compile Portunix binary
	err := compilePortunixBinary()
	suite.Require().NoError(err)

	config := DockerConfig{
		Image:            "alpine:3.18",
		ContainerName:    suite.testContainerName,
		InstallationType: "go",
		EnableSSH:        false,
		KeepRunning:      true,
		CacheShared:      false,
	}

	// Start container
	args := buildDockerRunArgs(config)
	cmd := exec.Command("docker", args...)
	err = cmd.Start()
	suite.Require().NoError(err)

	// Wait and copy binary
	err = waitForContainer(suite.testContainerName)
	suite.Require().NoError(err)

	err = copyPortunixToContainer(suite.testContainerName)
	suite.Require().NoError(err)

	// Install Go
	err = runPortunixInstallInContainer(suite.testContainerName, "go")
	suite.Require().NoError(err, "Go installation failed on Alpine")

	// Verify installation
	goVersionCmd := []string{"sh", "-c", "source /etc/profile && go version"}
	err = execInContainer(suite.testContainerName, goVersionCmd)
	suite.Require().NoError(err, "Go not working on Alpine")

	suite.T().Log("✅ Alpine Go installation test completed successfully")
}

// Helper functions

func isDockerAvailable() bool {
	cmd := exec.Command("docker", "version")
	return cmd.Run() == nil
}

func compilePortunixBinary() error {
	// Change to project root and compile
	cmd := exec.Command("go", "build", "-o", "portunix")
	cmd.Dir = "/media/zdenek/DevDisk/DEV/CassandraGargoyle/portunix/portunix"
	return cmd.Run()
}

func isContainerRunning(containerName string) (bool, error) {
	cmd := exec.Command("docker", "ps", "--filter", fmt.Sprintf("name=%s", containerName), "--format", "{{.Names}}")
	output, err := cmd.Output()
	if err != nil {
		return false, err
	}
	return strings.Contains(string(output), containerName), nil
}

func cleanupContainer(containerName string) {
	// Stop and remove test container
	exec.Command("docker", "stop", containerName).Run()
	exec.Command("docker", "rm", containerName).Run()
}

// Test InstallSoftwareInContainer integration
func (suite *ContainerGoIntegrationTestSuite) TestInstallSoftwareInContainer_Integration() {
	suite.testContainerName = fmt.Sprintf("portunix-install-integration-test-%d", time.Now().Unix())
	
	// Compile Portunix binary
	err := compilePortunixBinary()
	suite.Require().NoError(err)

	// Start minimal Ubuntu container
	cmd := exec.Command("docker", "run", "-d", "-it", "--name", suite.testContainerName, "ubuntu:22.04", "sleep", "infinity")
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

	// Test InstallSoftwareInContainer with Go type
	suite.T().Log("Testing InstallSoftwareInContainer with Go...")
	err = InstallSoftwareInContainer(suite.testContainerName, "go", pkgManager)
	suite.Require().NoError(err, "InstallSoftwareInContainer failed")

	// Verify Go is installed
	goVersionCmd := []string{"sh", "-c", "source /etc/profile && go version"}
	err = execInContainer(suite.testContainerName, goVersionCmd)
	suite.Require().NoError(err, "Go verification failed after InstallSoftwareInContainer")

	suite.T().Log("✅ InstallSoftwareInContainer integration test completed successfully")
}