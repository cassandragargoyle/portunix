package integration

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"portunix.ai/portunix/test/testframework"
)

// TestIssue075HugoContainerInstall tests Hugo installation in container via Portunix
// This test validates Issue #075 implementation:
// 1. Hugo installation with standard and extended variants
// 2. Cross-platform architecture detection
// 3. Version 0.150.1 support
// 4. hugo-extended alias functionality
// 5. Verification of SCSS/Sass support in extended variant
func TestIssue075HugoContainerInstall(t *testing.T) {
	tf := testframework.NewTestFramework("Issue075_Hugo_Container_Install")
	tf.Start(t, "Complete Hugo installation test: Standard + Extended variants in clean container environment")

	success := true
	defer func() {
		tf.Finish(t, success)
	}()

	if testing.Short() {
		tf.Info(t, "Skipping container test in short mode")
		t.Skip("Skipping container test in short mode")
	}

	suite := &HugoContainerTestSuite{
		tf:            tf,
		t:             t,
		containerName: "portunix-hugo-test",
		projectRoot:   "../..",
	}

	// Test phases
	tf.Step(t, "Initialize test environment")
	suite.setupEnvironment()
	defer suite.cleanup()

	tf.Separator()

	tf.Step(t, "Create Ubuntu container for Hugo testing")
	suite.createTestContainer()

	tf.Separator()

	tf.Step(t, "Install Portunix in container")
	suite.installPortunixInContainer()

	tf.Separator()

	tf.Step(t, "Test Hugo Standard variant installation")
	if !suite.testHugoStandard() {
		success = false
		return
	}

	tf.Separator()

	tf.Step(t, "Clean environment for Extended test")
	suite.cleanHugoInstallation()

	tf.Separator()

	tf.Step(t, "Test Hugo Extended variant installation")
	if !suite.testHugoExtended() {
		success = false
		return
	}

	tf.Separator()

	tf.Step(t, "Clean environment for alias test")
	suite.cleanHugoInstallation()

	tf.Separator()

	tf.Step(t, "Test hugo-extended alias installation")
	if !suite.testHugoExtendedAlias() {
		success = false
		return
	}

	tf.Separator()

	tf.Step(t, "Test Hugo site creation and build")
	if !suite.testHugoSiteCreation() {
		success = false
		return
	}

	tf.Success(t, "All Hugo installation variants tested successfully")
}

type HugoContainerTestSuite struct {
	tf            *testframework.TestFramework
	t             *testing.T
	containerName string
	projectRoot   string
	binaryPath    string
}

func (suite *HugoContainerTestSuite) runContainerCommand(command string, args ...string) (string, error) {
	// Use Portunix container exec instead of direct docker/podman
	fullArgs := append([]string{"container", "exec", suite.containerName, command}, args...)

	suite.tf.Command(suite.t, suite.binaryPath, fullArgs)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, suite.binaryPath, fullArgs...)
	output, err := cmd.CombinedOutput()

	if err != nil {
		suite.tf.Error(suite.t, fmt.Sprintf("Container command failed: %s", command), err.Error())
		suite.tf.Output(suite.t, string(output), 500)
		return string(output), err
	}

	suite.tf.Output(suite.t, string(output), 300)
	return string(output), nil
}

func (suite *HugoContainerTestSuite) setupEnvironment() {
	// Build portunix binary if needed
	suite.binaryPath = filepath.Join(suite.projectRoot, "portunix")
	if _, err := os.Stat(suite.binaryPath); os.IsNotExist(err) {
		suite.tf.Info(suite.t, "Building portunix binary")
		cmd := exec.Command("go", "build", "-o", "portunix", ".")
		cmd.Dir = suite.projectRoot
		if output, err := cmd.CombinedOutput(); err != nil {
			suite.tf.Error(suite.t, "Failed to build portunix binary", err.Error())
			suite.t.Fatalf("Build failed: %v\nOutput: %s", err, string(output))
		}
		suite.tf.Success(suite.t, "Portunix binary built successfully")
	} else {
		suite.tf.Success(suite.t, "Using existing portunix binary")
	}

	// Verify container runtime is available via Portunix
	cmd := exec.Command(suite.binaryPath, "container", "--help")
	if output, err := cmd.CombinedOutput(); err != nil {
		suite.tf.Error(suite.t, "Portunix container system not available, cannot run container tests", err.Error())
		suite.tf.Output(suite.t, string(output), 300)
		suite.t.Skip("Portunix container system not available")
	}
	suite.tf.Success(suite.t, "Portunix container system available")
}

func (suite *HugoContainerTestSuite) createTestContainer() {
	// Remove existing container using Portunix
	suite.tf.Info(suite.t, "Cleaning up any existing test containers")
	cleanupCmd := exec.Command(suite.binaryPath, "container", "remove", suite.containerName)
	cleanupCmd.Run() // Ignore errors - container might not exist

	suite.tf.Info(suite.t, "Creating Ubuntu 22.04 container for Hugo testing")

	// Use Portunix container run for test container
	suite.tf.Command(suite.t, suite.binaryPath, []string{"container", "run", "--name", suite.containerName, "-d", "ubuntu:22.04", "sleep", "3600"})
	cmd := exec.Command(suite.binaryPath, "container", "run", "--name", suite.containerName, "-d", "ubuntu:22.04", "sleep", "3600")

	if output, err := cmd.CombinedOutput(); err != nil {
		suite.tf.Error(suite.t, "Failed to create test container via Portunix", err.Error())
		suite.tf.Output(suite.t, string(output), 500)
		suite.t.Fatalf("Test container creation failed: %v", err)
	}

	// Wait for container setup
	suite.tf.Info(suite.t, "Waiting for test container initialization", "10s")
	time.Sleep(10 * time.Second)

	// Verify container is ready
	if output, err := suite.runContainerCommand("echo", "Hugo test container ready"); err != nil {
		suite.tf.Error(suite.t, "Test container not responding", err.Error())
		suite.t.Fatalf("Test container setup verification failed: %v", err)
	} else if !strings.Contains(output, "Hugo test container ready") {
		suite.tf.Error(suite.t, "Test container startup verification failed")
		suite.t.Fatalf("Unexpected test container response: %s", output)
	}

	suite.tf.Success(suite.t, "Ubuntu test container ready for Hugo installation")
}

func (suite *HugoContainerTestSuite) installPortunixInContainer() {
	suite.tf.Info(suite.t, "Installing Portunix binary in container via Portunix container system")

	// Use Portunix container copy to copy binary
	suite.tf.Command(suite.t, suite.binaryPath, []string{"container", "cp", suite.binaryPath, suite.containerName + ":/usr/local/bin/portunix"})
	cmd := exec.Command(suite.binaryPath, "container", "cp", suite.binaryPath, suite.containerName+":/usr/local/bin/portunix")
	if output, err := cmd.CombinedOutput(); err != nil {
		suite.tf.Error(suite.t, "Failed to copy portunix to container via Portunix", err.Error())
		suite.tf.Output(suite.t, string(output), 500)
		suite.t.Fatalf("Copy failed: %v", err)
	}

	// Make executable
	if _, err := suite.runContainerCommand("chmod", "+x", "/usr/local/bin/portunix"); err != nil {
		suite.t.Fatalf("Failed to make portunix executable: %v", err)
	}

	// Verify installation
	output, err := suite.runContainerCommand("/usr/local/bin/portunix", "--version")
	if err != nil {
		suite.tf.Error(suite.t, "Portunix not working in container", err.Error())
		suite.t.Fatalf("Version check failed: %v", err)
	}

	suite.tf.Success(suite.t, "Portunix installed and verified in container",
		fmt.Sprintf("Version: %s", strings.TrimSpace(output)))
}

func (suite *HugoContainerTestSuite) testHugoStandard() bool {
	suite.tf.Info(suite.t, "Testing Hugo Standard variant installation")

	// Test dry-run first
	suite.tf.Info(suite.t, "Testing dry-run mode for Hugo standard installation")
	dryRunOutput, err := suite.runContainerCommand("/usr/local/bin/portunix", "install", "hugo", "--variant", "standard", "--dry-run")
	if err != nil {
		suite.tf.Error(suite.t, "Hugo standard dry-run failed", err.Error())
		return false
	}

	if !strings.Contains(dryRunOutput, "📦 INSTALLING: Hugo Static Site Generator") ||
	   !strings.Contains(dryRunOutput, "🔧 Variant: standard") ||
	   !strings.Contains(dryRunOutput, "v0.150.1") {
		suite.tf.Error(suite.t, "Dry-run output missing expected Hugo standard info")
		suite.tf.Output(suite.t, dryRunOutput, 500)
		return false
	}
	suite.tf.Success(suite.t, "Hugo standard package configuration verified")

	// Install Hugo standard via Portunix
	suite.tf.Info(suite.t, "Installing Hugo standard using Portunix")
	installOutput, err := suite.runContainerCommand("/usr/local/bin/portunix", "install", "hugo", "--variant", "standard")
	if err != nil {
		suite.tf.Error(suite.t, "Hugo standard installation via Portunix failed", err.Error())
		suite.tf.Output(suite.t, installOutput, 500)
		return false
	}

	// Verify Hugo installation
	hugoVersion, err := suite.runContainerCommand("hugo", "version")
	if err != nil {
		suite.tf.Error(suite.t, "Hugo command not available after installation", err.Error())
		return false
	}

	if !strings.Contains(hugoVersion, "0.150.1") {
		suite.tf.Error(suite.t, "Hugo version mismatch", fmt.Sprintf("Expected v0.150.1, got: %s", hugoVersion))
		return false
	}

	// Verify it's standard version (should NOT have extended features)
	if strings.Contains(strings.ToLower(hugoVersion), "extended") {
		suite.tf.Error(suite.t, "Standard variant incorrectly shows extended features")
		return false
	}

	suite.tf.Success(suite.t, "Hugo Standard variant installed successfully",
		fmt.Sprintf("Version: %s", strings.TrimSpace(hugoVersion)))

	return true
}

func (suite *HugoContainerTestSuite) testHugoExtended() bool {
	suite.tf.Info(suite.t, "Testing Hugo Extended variant installation")

	// Test dry-run first
	suite.tf.Info(suite.t, "Testing dry-run mode for Hugo extended installation")
	dryRunOutput, err := suite.runContainerCommand("/usr/local/bin/portunix", "install", "hugo", "--variant", "extended", "--dry-run")
	if err != nil {
		suite.tf.Error(suite.t, "Hugo extended dry-run failed", err.Error())
		return false
	}

	if !strings.Contains(dryRunOutput, "📦 INSTALLING: Hugo Static Site Generator") ||
	   !strings.Contains(dryRunOutput, "🔧 Variant: extended") ||
	   !strings.Contains(dryRunOutput, "v0.150.1") ||
	   !strings.Contains(dryRunOutput, "hugo_extended_0.150.1_linux-amd64.tar.gz") {
		suite.tf.Error(suite.t, "Dry-run output missing expected Hugo extended info")
		suite.tf.Output(suite.t, dryRunOutput, 500)
		return false
	}
	suite.tf.Success(suite.t, "Hugo extended package configuration verified")

	// Install Hugo extended via Portunix
	suite.tf.Info(suite.t, "Installing Hugo extended using Portunix")
	installOutput, err := suite.runContainerCommand("/usr/local/bin/portunix", "install", "hugo", "--variant", "extended")
	if err != nil {
		suite.tf.Error(suite.t, "Hugo extended installation via Portunix failed", err.Error())
		suite.tf.Output(suite.t, installOutput, 500)
		return false
	}

	// Verify Hugo installation
	hugoVersion, err := suite.runContainerCommand("hugo", "version")
	if err != nil {
		suite.tf.Error(suite.t, "Hugo command not available after extended installation", err.Error())
		return false
	}

	if !strings.Contains(hugoVersion, "0.150.1") {
		suite.tf.Error(suite.t, "Hugo extended version mismatch", fmt.Sprintf("Expected v0.150.1, got: %s", hugoVersion))
		return false
	}

	// Verify it's extended version (should have extended features)
	if !strings.Contains(strings.ToLower(hugoVersion), "extended") {
		suite.tf.Error(suite.t, "Extended variant does not show extended features")
		return false
	}

	// Test SCSS support (extended feature)
	suite.tf.Info(suite.t, "Testing SCSS support in Hugo extended")
	envOutput, err := suite.runContainerCommand("hugo", "env")
	if err != nil {
		suite.tf.Warning(suite.t, "Hugo env command failed", err.Error())
	} else {
		suite.tf.Output(suite.t, envOutput, 300)
		if strings.Contains(strings.ToLower(envOutput), "extended") {
			suite.tf.Success(suite.t, "SCSS/Sass support confirmed in Hugo extended")
		}
	}

	suite.tf.Success(suite.t, "Hugo Extended variant installed successfully",
		fmt.Sprintf("Version: %s", strings.TrimSpace(hugoVersion)))

	return true
}

func (suite *HugoContainerTestSuite) testHugoExtendedAlias() bool {
	suite.tf.Info(suite.t, "Testing hugo-extended alias functionality")

	// Test dry-run of hugo-extended alias
	suite.tf.Info(suite.t, "Testing dry-run mode for hugo-extended alias")
	dryRunOutput, err := suite.runContainerCommand("/usr/local/bin/portunix", "install", "hugo-extended", "--dry-run")
	if err != nil {
		suite.tf.Error(suite.t, "hugo-extended alias dry-run failed", err.Error())
		return false
	}

	if !strings.Contains(dryRunOutput, "📦 INSTALLING: Hugo Extended Static Site Generator") ||
	   !strings.Contains(dryRunOutput, "🏗️  Installation type: redirect") {
		suite.tf.Error(suite.t, "Dry-run output missing expected hugo-extended alias info")
		suite.tf.Output(suite.t, dryRunOutput, 500)
		return false
	}
	suite.tf.Success(suite.t, "hugo-extended alias configuration verified")

	// Install hugo-extended alias via Portunix
	suite.tf.Info(suite.t, "Installing hugo-extended alias using Portunix")
	installOutput, err := suite.runContainerCommand("/usr/local/bin/portunix", "install", "hugo-extended")
	if err != nil {
		suite.tf.Error(suite.t, "hugo-extended alias installation via Portunix failed", err.Error())
		suite.tf.Output(suite.t, installOutput, 500)
		return false
	}

	// Should contain redirect message
	if !strings.Contains(installOutput, "🔀 Redirecting to package: hugo") {
		suite.tf.Error(suite.t, "hugo-extended alias did not show expected redirect message")
		suite.tf.Output(suite.t, installOutput, 500)
		return false
	}

	// Verify Hugo installation (should be extended variant)
	hugoVersion, err := suite.runContainerCommand("hugo", "version")
	if err != nil {
		suite.tf.Error(suite.t, "Hugo command not available after hugo-extended alias installation", err.Error())
		return false
	}

	if !strings.Contains(hugoVersion, "0.150.1") || !strings.Contains(strings.ToLower(hugoVersion), "extended") {
		suite.tf.Error(suite.t, "hugo-extended alias did not install extended variant",
			fmt.Sprintf("Got: %s", hugoVersion))
		return false
	}

	suite.tf.Success(suite.t, "hugo-extended alias installed successfully",
		fmt.Sprintf("Redirected to Hugo Extended: %s", strings.TrimSpace(hugoVersion)))

	return true
}

func (suite *HugoContainerTestSuite) testHugoSiteCreation() bool {
	suite.tf.Info(suite.t, "Testing Hugo site creation and build functionality")

	// Create a test Hugo site
	suite.tf.Info(suite.t, "Creating test Hugo site")
	createOutput, err := suite.runContainerCommand("hugo", "new", "site", "/tmp/test-hugo-site")
	if err != nil {
		suite.tf.Error(suite.t, "Failed to create Hugo site", err.Error())
		suite.tf.Output(suite.t, createOutput, 500)
		return false
	}

	if !strings.Contains(createOutput, "Congratulations") {
		suite.tf.Error(suite.t, "Hugo site creation did not complete successfully")
		suite.tf.Output(suite.t, createOutput, 500)
		return false
	}

	suite.tf.Success(suite.t, "Hugo site created successfully")

	// Change to site directory and create basic content
	suite.tf.Info(suite.t, "Setting up basic site configuration")
	configCommands := []string{
		"cd /tmp/test-hugo-site",
		"echo 'baseURL = \"https://example.com\"' > hugo.toml",
		"echo 'title = \"Test Hugo Site\"' >> hugo.toml",
		"echo 'theme = \"\"' >> hugo.toml",
	}

	for _, cmd := range configCommands {
		if _, err := suite.runContainerCommand("bash", "-c", cmd); err != nil {
			suite.tf.Error(suite.t, "Failed to configure Hugo site", err.Error())
			return false
		}
	}

	// Create a simple content file
	suite.tf.Info(suite.t, "Creating test content")
	contentCmd := `cd /tmp/test-hugo-site && echo '+++
title = "Test Page"
date = "2025-09-26"
draft = false
+++

# Test Hugo Site

This is a test page to verify Hugo functionality.' > content/test.md`

	if _, err := suite.runContainerCommand("bash", "-c", contentCmd); err != nil {
		suite.tf.Error(suite.t, "Failed to create test content", err.Error())
		return false
	}

	// Build the site
	suite.tf.Info(suite.t, "Building Hugo site")
	buildOutput, err := suite.runContainerCommand("bash", "-c", "cd /tmp/test-hugo-site && hugo --quiet")
	if err != nil {
		suite.tf.Error(suite.t, "Hugo site build failed", err.Error())
		suite.tf.Output(suite.t, buildOutput, 500)
		return false
	}

	// Verify build output exists
	suite.tf.Info(suite.t, "Verifying build output")
	listOutput, err := suite.runContainerCommand("ls", "-la", "/tmp/test-hugo-site/public/")
	if err != nil {
		suite.tf.Error(suite.t, "Failed to list build output", err.Error())
		return false
	}

	if !strings.Contains(listOutput, "index.html") {
		suite.tf.Error(suite.t, "Hugo build did not generate expected files")
		suite.tf.Output(suite.t, listOutput, 500)
		return false
	}

	suite.tf.Success(suite.t, "Hugo site built successfully")
	suite.tf.Output(suite.t, listOutput, 300)

	// Test module system (if supported)
	suite.tf.Info(suite.t, "Testing Hugo module system")
	modOutput, err := suite.runContainerCommand("bash", "-c", "cd /tmp/test-hugo-site && hugo mod help")
	if err != nil {
		suite.tf.Warning(suite.t, "Hugo modules not available or failed", err.Error())
	} else {
		if strings.Contains(modOutput, "Hugo Modules") {
			suite.tf.Success(suite.t, "Hugo module system is available")
		}
	}

	return true
}

func (suite *HugoContainerTestSuite) cleanHugoInstallation() {
	suite.tf.Info(suite.t, "Cleaning Hugo installation from container")

	// Remove Hugo binary
	suite.runContainerCommand("rm", "-f", "/usr/local/bin/hugo")

	// Clean any cached files
	suite.runContainerCommand("rm", "-rf", "/tmp/test-hugo-site")

	// Verify Hugo is removed
	if _, err := suite.runContainerCommand("hugo", "version"); err == nil {
		suite.tf.Warning(suite.t, "Hugo still available after cleanup - PATH might have other installations")
	} else {
		suite.tf.Success(suite.t, "Hugo successfully removed from container")
	}
}

func (suite *HugoContainerTestSuite) cleanup() {
	suite.tf.Info(suite.t, "Cleaning up test environment")

	// Remove test container using Portunix
	suite.tf.Command(suite.t, suite.binaryPath, []string{"container", "remove", suite.containerName})
	cmd := exec.Command(suite.binaryPath, "container", "remove", suite.containerName)
	cmd.Run() // Ignore errors in cleanup

	suite.tf.Success(suite.t, "Cleanup completed using Portunix container system")
}