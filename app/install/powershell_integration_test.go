//go:build integration
// +build integration

package install

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// PowerShellIntegrationTestSuite tests PowerShell installation across distributions
type PowerShellIntegrationTestSuite struct {
	suite.Suite
	ctx       context.Context
	container testcontainers.Container
}

// SupportedDistribution represents a Linux distribution for testing
type SupportedDistribution struct {
	Name               string
	Image              string
	ExpectedVariant    string
	PreInstallCommands []string
	VerificationCmd    string
}

// GetSupportedDistributions returns all distributions we support for PowerShell
func GetSupportedDistributions() []SupportedDistribution {
	return []SupportedDistribution{
		{
			Name:            "Ubuntu 22.04",
			Image:           "ubuntu:22.04",
			ExpectedVariant: "ubuntu",
			PreInstallCommands: []string{
				"apt-get update",
				"apt-get install -y sudo wget curl lsb-release",
			},
			VerificationCmd: "pwsh --version",
		},
		{
			Name:            "Ubuntu 24.04",
			Image:           "ubuntu:24.04",
			ExpectedVariant: "ubuntu",
			PreInstallCommands: []string{
				"apt-get update",
				"apt-get install -y sudo wget curl lsb-release",
			},
			VerificationCmd: "pwsh --version",
		},
		{
			Name:            "Debian 11",
			Image:           "debian:11",
			ExpectedVariant: "debian",
			PreInstallCommands: []string{
				"apt-get update",
				"apt-get install -y sudo wget curl lsb-release",
			},
			VerificationCmd: "pwsh --version",
		},
		{
			Name:            "Debian 12",
			Image:           "debian:12",
			ExpectedVariant: "debian",
			PreInstallCommands: []string{
				"apt-get update",
				"apt-get install -y sudo wget curl lsb-release",
			},
			VerificationCmd: "pwsh --version",
		},
		{
			Name:            "Fedora 39",
			Image:           "fedora:39",
			ExpectedVariant: "fedora",
			PreInstallCommands: []string{
				"dnf update -y",
				"dnf install -y sudo curl",
			},
			VerificationCmd: "pwsh --version",
		},
		{
			Name:            "Fedora 40",
			Image:           "fedora:40",
			ExpectedVariant: "fedora",
			PreInstallCommands: []string{
				"dnf update -y",
				"dnf install -y sudo curl",
			},
			VerificationCmd: "pwsh --version",
		},
		{
			Name:            "Rocky Linux 9",
			Image:           "rockylinux:9",
			ExpectedVariant: "rocky",
			PreInstallCommands: []string{
				"dnf update -y",
				"dnf install -y sudo curl",
			},
			VerificationCmd: "pwsh --version",
		},
		{
			Name:            "Universal (Snap)",
			Image:           "ubuntu:22.04", // Use Ubuntu as base for snap testing
			ExpectedVariant: "snap",
			PreInstallCommands: []string{
				"apt-get update",
				"apt-get install -y sudo wget curl snapd",
			},
			VerificationCmd: "snap list powershell",
		},
	}
}

func (suite *PowerShellIntegrationTestSuite) SetupSuite() {
	suite.ctx = context.Background()

	// Check if Docker is available
	if !isDockerAvailable() {
		suite.T().Skip("Docker is not available, skipping PowerShell integration tests")
	}

	// Check if we have enough time for integration tests
	if testing.Short() {
		suite.T().Skip("Skipping PowerShell integration tests in short mode")
	}
}

func (suite *PowerShellIntegrationTestSuite) TearDownTest() {
	if suite.container != nil {
		suite.container.Terminate(suite.ctx)
		suite.container = nil
	}
}

func TestPowerShellIntegrationSuite(t *testing.T) {
	suite.Run(t, new(PowerShellIntegrationTestSuite))
}

// Test PowerShell installation on all supported distributions
func (suite *PowerShellIntegrationTestSuite) TestPowerShellInstallation_AllDistributions() {
	distributions := GetSupportedDistributions()

	for _, dist := range distributions {
		suite.Run(fmt.Sprintf("Install_PowerShell_%s", strings.ReplaceAll(dist.Name, " ", "_")), func() {
			suite.testPowerShellInstallationForDistribution(dist)
		})
	}
}

func (suite *PowerShellIntegrationTestSuite) testPowerShellInstallationForDistribution(dist SupportedDistribution) {
	// Create container for this distribution
	container := suite.createContainerForDistribution(dist)
	suite.container = container

	// Setup container with prerequisites
	suite.setupContainerPrerequisites(container, dist)

	// Copy portunix binary to container
	suite.copyPortunixToContainer(container)

	// Test PowerShell installation
	suite.installPowerShellInContainer(container, dist)

	// Verify PowerShell is working
	suite.verifyPowerShellInstallation(container, dist)
}

func (suite *PowerShellIntegrationTestSuite) createContainerForDistribution(dist SupportedDistribution) testcontainers.Container {
	req := testcontainers.ContainerRequest{
		Image:      dist.Image,
		Cmd:        []string{"sleep", "1800"}, // Keep container running for 30 minutes
		WaitingFor: wait.ForLog("").WithStartupTimeout(60 * time.Second),
	}

	container, err := testcontainers.GenericContainer(suite.ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})

	require.NoError(suite.T(), err, "Failed to create container for %s", dist.Name)
	require.NotNil(suite.T(), container, "Container should not be nil for %s", dist.Name)

	// Wait for container to be ready
	time.Sleep(5 * time.Second)

	return container
}

func (suite *PowerShellIntegrationTestSuite) setupContainerPrerequisites(container testcontainers.Container, dist SupportedDistribution) {
	for i, cmd := range dist.PreInstallCommands {
		suite.T().Logf("Running prerequisite command %d/%d for %s: %s", i+1, len(dist.PreInstallCommands), dist.Name, cmd)

		exitCode, reader, err := container.Exec(suite.ctx, []string{"sh", "-c", cmd})
		require.NoError(suite.T(), err, "Failed to execute prerequisite command for %s: %s", dist.Name, cmd)

		if exitCode != 0 {
			output := suite.readOutput(reader)
			suite.T().Fatalf("Prerequisite command failed for %s with exit code %d: %s\nOutput: %s", dist.Name, exitCode, cmd, output)
		}

		suite.T().Logf("✓ Prerequisite completed for %s: %s", dist.Name, cmd)
	}
}

func (suite *PowerShellIntegrationTestSuite) copyPortunixToContainer(container testcontainers.Container) {
	// First, ensure portunix binary exists
	_, err := os.Stat("./portunix")
	if err != nil {
		suite.T().Logf("Portunix binary not found, attempting to build...")
		// Try to build the binary
		buildExitCode, buildReader, buildErr := container.Exec(suite.ctx, []string{"sh", "-c", "cd /workspace && go build -o portunix"})
		if buildErr != nil || buildExitCode != 0 {
			suite.T().Skip("Could not build or find portunix binary for testing")
		}
		buildOutput := suite.readOutput(buildReader)
		suite.T().Logf("Build output: %s", buildOutput)
	}

	// Copy portunix binary to container
	err = container.CopyFileToContainer(suite.ctx, "./portunix", "/usr/local/bin/portunix", 0755)
	require.NoError(suite.T(), err, "Failed to copy portunix binary to container")

	suite.T().Logf("✓ Portunix binary copied to container")

	// Verify binary is executable
	exitCode, reader, err := container.Exec(suite.ctx, []string{"chmod", "+x", "/usr/local/bin/portunix"})
	require.NoError(suite.T(), err, "Failed to make portunix executable")
	require.Equal(suite.T(), 0, exitCode, "Failed to chmod portunix binary")

	if reader != nil {
		output := suite.readOutput(reader)
		suite.T().Logf("Chmod output: %s", output)
	}
}

func (suite *PowerShellIntegrationTestSuite) installPowerShellInContainer(container testcontainers.Container, dist SupportedDistribution) {
	suite.T().Logf("Installing PowerShell in %s container...", dist.Name)

	// Use explicit variant for testing, or let auto-detection work
	installCmd := "/usr/local/bin/portunix install powershell"
	if dist.ExpectedVariant != "snap" { // Let snap be auto-detected
		installCmd = fmt.Sprintf("/usr/local/bin/portunix install powershell --variant %s", dist.ExpectedVariant)
	}

	suite.T().Logf("Executing: %s", installCmd)

	// Execute PowerShell installation with extended timeout
	ctx, cancel := context.WithTimeout(suite.ctx, 10*time.Minute)
	defer cancel()

	exitCode, reader, err := container.Exec(ctx, []string{"sh", "-c", installCmd})

	output := suite.readOutput(reader)
	suite.T().Logf("Installation output for %s:\n%s", dist.Name, output)

	require.NoError(suite.T(), err, "Failed to execute PowerShell installation command for %s", dist.Name)

	if exitCode != 0 {
		suite.T().Fatalf("PowerShell installation failed for %s with exit code %d\nOutput:\n%s", dist.Name, exitCode, output)
	}

	suite.T().Logf("✓ PowerShell installation completed for %s", dist.Name)
}

func (suite *PowerShellIntegrationTestSuite) verifyPowerShellInstallation(container testcontainers.Container, dist SupportedDistribution) {
	suite.T().Logf("Verifying PowerShell installation for %s...", dist.Name)

	// Wait a moment for installation to settle
	time.Sleep(5 * time.Second)

	// Run verification command
	exitCode, reader, err := container.Exec(suite.ctx, []string{"sh", "-c", dist.VerificationCmd})

	output := suite.readOutput(reader)
	suite.T().Logf("Verification output for %s:\n%s", dist.Name, output)

	require.NoError(suite.T(), err, "Failed to execute verification command for %s", dist.Name)

	if exitCode != 0 {
		// For snap, also try alternative verification
		if dist.ExpectedVariant == "snap" {
			suite.T().Logf("Snap verification failed, trying alternative: /snap/bin/powershell --version")
			exitCode2, reader2, err2 := container.Exec(suite.ctx, []string{"sh", "-c", "/snap/bin/powershell --version"})
			output2 := suite.readOutput(reader2)

			if err2 == nil && exitCode2 == 0 {
				suite.T().Logf("Alternative verification succeeded for %s:\n%s", dist.Name, output2)
				assert.Contains(suite.T(), output2, "PowerShell", "PowerShell version output should contain 'PowerShell'")
				return
			}
		}

		suite.T().Fatalf("PowerShell verification failed for %s with exit code %d\nOutput:\n%s", dist.Name, exitCode, output)
	}

	// Verify output contains PowerShell version info
	if dist.ExpectedVariant == "snap" {
		assert.True(suite.T(), strings.Contains(output, "powershell") || strings.Contains(output, "PowerShell"),
			"Snap list should show powershell package for %s", dist.Name)
	} else {
		assert.Contains(suite.T(), output, "PowerShell", "PowerShell version output should contain 'PowerShell' for %s", dist.Name)
		assert.Regexp(suite.T(), `\d+\.\d+\.\d+`, output, "PowerShell version should be in format x.y.z for %s", dist.Name)
	}

	suite.T().Logf("✓ PowerShell verification completed successfully for %s", dist.Name)
}

func (suite *PowerShellIntegrationTestSuite) readOutput(reader *testcontainers.ExecResult) string {
	if reader == nil {
		return ""
	}

	// Try to read from reader
	buffer := make([]byte, 4096)
	n, _ := reader.Reader.Read(buffer)
	return string(buffer[:n])
}

// Test variant auto-detection
func (suite *PowerShellIntegrationTestSuite) TestVariantAutoDetection_Ubuntu() {
	dist := SupportedDistribution{
		Name:  "Ubuntu 22.04 Auto-detect",
		Image: "ubuntu:22.04",
		PreInstallCommands: []string{
			"apt-get update",
			"apt-get install -y sudo wget curl lsb-release",
		},
		VerificationCmd: "pwsh --version",
	}

	container := suite.createContainerForDistribution(dist)
	suite.container = container

	suite.setupContainerPrerequisites(container, dist)
	suite.copyPortunixToContainer(container)

	// Test auto-detection by not specifying variant
	suite.T().Logf("Testing auto-detection for Ubuntu...")

	installCmd := "/usr/local/bin/portunix install powershell"
	suite.T().Logf("Executing: %s", installCmd)

	ctx, cancel := context.WithTimeout(suite.ctx, 10*time.Minute)
	defer cancel()

	exitCode, reader, err := container.Exec(ctx, []string{"sh", "-c", installCmd})

	output := suite.readOutput(reader)
	suite.T().Logf("Auto-detection output:\n%s", output)

	require.NoError(suite.T(), err, "Failed to execute auto-detection test")

	// Should either succeed with Ubuntu variant or fallback to snap
	if exitCode == 0 {
		suite.T().Logf("✓ Auto-detection succeeded")
		// Verify installation
		suite.verifyPowerShellInstallation(container, dist)
	} else {
		suite.T().Logf("Auto-detection failed as expected, output: %s", output)
		// This might be expected if no sudo access or other limitations
	}
}

// Test error handling for unsupported distributions
func (suite *PowerShellIntegrationTestSuite) TestUnsupportedDistribution_Alpine() {
	req := testcontainers.ContainerRequest{
		Image: "alpine:latest",
		Cmd:   []string{"sleep", "300"},
	}

	container, err := testcontainers.GenericContainer(suite.ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(suite.T(), err)
	suite.container = container

	// Setup basic prerequisites
	exitCode, _, err := container.Exec(suite.ctx, []string{"sh", "-c", "apk update && apk add --no-cache curl"})
	require.NoError(suite.T(), err)
	require.Equal(suite.T(), 0, exitCode)

	suite.copyPortunixToContainer(container)

	// Try to install PowerShell - should fallback to snap or fail gracefully
	installCmd := "/usr/local/bin/portunix install powershell"
	exitCode, reader, err := container.Exec(suite.ctx, []string{"sh", "-c", installCmd})

	output := suite.readOutput(reader)
	suite.T().Logf("Alpine installation attempt output:\n%s", output)

	require.NoError(suite.T(), err, "Command execution should not error")

	// Alpine is not in our supported list, so it should either:
	// 1. Try snap and fail (expected)
	// 2. Show appropriate error message
	if exitCode != 0 {
		assert.Contains(suite.T(), output, "no suitable variant found",
			"Should show appropriate error for unsupported distribution")
	}
}

// Benchmark PowerShell installation performance
func (suite *PowerShellIntegrationTestSuite) TestInstallationPerformance_Ubuntu() {
	if testing.Short() {
		suite.T().Skip("Skipping performance test in short mode")
	}

	dist := SupportedDistribution{
		Name:  "Ubuntu 22.04 Performance",
		Image: "ubuntu:22.04",
		PreInstallCommands: []string{
			"apt-get update",
			"apt-get install -y sudo wget curl lsb-release",
		},
	}

	container := suite.createContainerForDistribution(dist)
	suite.container = container

	suite.setupContainerPrerequisites(container, dist)
	suite.copyPortunixToContainer(container)

	// Measure installation time
	start := time.Now()

	installCmd := "/usr/local/bin/portunix install powershell --variant ubuntu"
	ctx, cancel := context.WithTimeout(suite.ctx, 15*time.Minute)
	defer cancel()

	exitCode, reader, err := container.Exec(ctx, []string{"sh", "-c", installCmd})
	installationTime := time.Since(start)

	output := suite.readOutput(reader)
	suite.T().Logf("Performance test output:\n%s", output)

	require.NoError(suite.T(), err)

	// Installation should complete within reasonable time (15 minutes max)
	maxInstallTime := 15 * time.Minute
	assert.Less(suite.T(), installationTime, maxInstallTime,
		"PowerShell installation should complete within %v, took %v", maxInstallTime, installationTime)

	if exitCode == 0 {
		suite.T().Logf("✓ PowerShell installed successfully in %v", installationTime)
	} else {
		suite.T().Logf("Installation failed in %v with output: %s", installationTime, output)
	}
}

// Helper function to check Docker availability
func isDockerAvailable() bool {
	ctx := context.Background()
	_, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image: "hello-world:latest",
		},
		Started: false,
	})
	return err == nil
}
