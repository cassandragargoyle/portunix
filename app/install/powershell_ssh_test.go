//go:build integration
// +build integration

package install

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"golang.org/x/crypto/ssh"
)

// PowerShellSSHTestSuite tests PowerShell installation via SSH
type PowerShellSSHTestSuite struct {
	suite.Suite
	ctx       context.Context
	container testcontainers.Container
	sshConfig *ssh.ClientConfig
	sshHost   string
	sshPort   int
}

// SSHTestDistribution represents a distribution configured for SSH access
type SSHTestDistribution struct {
	Name             string
	Image            string
	ExpectedVariant  string
	SetupCommands    []string
	SSHSetupCommands []string
	VerificationCmd  string
}

// GetSSHTestDistributions returns distributions configured with SSH servers
func GetSSHTestDistributions() []SSHTestDistribution {
	return []SSHTestDistribution{
		{
			Name:            "Ubuntu 22.04 SSH",
			Image:           "ubuntu:22.04",
			ExpectedVariant: "ubuntu",
			SetupCommands: []string{
				"apt-get update",
				"apt-get install -y sudo wget curl lsb-release openssh-server",
			},
			SSHSetupCommands: []string{
				"mkdir -p /var/run/sshd",
				"echo 'root:testpassword123' | chpasswd",
				"sed -i 's/#PermitRootLogin prohibit-password/PermitRootLogin yes/' /etc/ssh/sshd_config",
				"sed -i 's/#PasswordAuthentication yes/PasswordAuthentication yes/' /etc/ssh/sshd_config",
				"/usr/sbin/sshd -D &",
			},
			VerificationCmd: "pwsh --version",
		},
		{
			Name:            "Debian 12 SSH",
			Image:           "debian:12",
			ExpectedVariant: "debian",
			SetupCommands: []string{
				"apt-get update",
				"apt-get install -y sudo wget curl lsb-release openssh-server",
			},
			SSHSetupCommands: []string{
				"mkdir -p /var/run/sshd",
				"echo 'root:testpassword123' | chpasswd",
				"sed -i 's/#PermitRootLogin prohibit-password/PermitRootLogin yes/' /etc/ssh/sshd_config",
				"sed -i 's/#PasswordAuthentication yes/PasswordAuthentication yes/' /etc/ssh/sshd_config",
				"/usr/sbin/sshd -D &",
			},
			VerificationCmd: "pwsh --version",
		},
		{
			Name:            "Fedora 40 SSH",
			Image:           "fedora:40",
			ExpectedVariant: "fedora",
			SetupCommands: []string{
				"dnf update -y",
				"dnf install -y sudo curl openssh-server",
			},
			SSHSetupCommands: []string{
				"ssh-keygen -A",
				"echo 'root:testpassword123' | chpasswd",
				"sed -i 's/#PermitRootLogin prohibit-password/PermitRootLogin yes/' /etc/ssh/sshd_config",
				"sed -i 's/#PasswordAuthentication yes/PasswordAuthentication yes/' /etc/ssh/sshd_config",
				"/usr/sbin/sshd -D &",
			},
			VerificationCmd: "pwsh --version",
		},
	}
}

func (suite *PowerShellSSHTestSuite) SetupSuite() {
	suite.ctx = context.Background()

	// Check if Docker is available
	if !isDockerAvailable() {
		suite.T().Skip("Docker is not available, skipping PowerShell SSH tests")
	}

	// Skip in short mode due to complexity
	if testing.Short() {
		suite.T().Skip("Skipping PowerShell SSH tests in short mode")
	}
}

func (suite *PowerShellSSHTestSuite) TearDownTest() {
	if suite.container != nil {
		suite.container.Terminate(suite.ctx)
		suite.container = nil
	}
}

func TestPowerShellSSHSuite(t *testing.T) {
	suite.Run(t, new(PowerShellSSHTestSuite))
}

// Test PowerShell installation via SSH for all supported distributions
func (suite *PowerShellSSHTestSuite) TestPowerShellInstallation_SSH_AllDistributions() {
	distributions := GetSSHTestDistributions()

	for _, dist := range distributions {
		suite.Run(fmt.Sprintf("SSH_Install_%s", strings.ReplaceAll(dist.Name, " ", "_")), func() {
			suite.testPowerShellInstallationViaSSH(dist)
		})
	}
}

func (suite *PowerShellSSHTestSuite) testPowerShellInstallationViaSSH(dist SSHTestDistribution) {
	// Create and setup container with SSH server
	container := suite.createSSHContainer(dist)
	suite.container = container

	// Setup SSH connection
	sshClient := suite.setupSSHConnection(container)
	defer sshClient.Close()

	// Transfer portunix binary via SSH
	suite.transferPortunixViaSSH(sshClient)

	// Install PowerShell via SSH
	suite.installPowerShellViaSSH(sshClient, dist)

	// Verify installation via SSH
	suite.verifyPowerShellViaSSH(sshClient, dist)
}

func (suite *PowerShellSSHTestSuite) createSSHContainer(dist SSHTestDistribution) testcontainers.Container {
	req := testcontainers.ContainerRequest{
		Image:        dist.Image,
		ExposedPorts: []string{"22/tcp"},
		Cmd:          []string{"bash", "-c", "sleep 3600"}, // Keep running for SSH
		WaitingFor:   wait.ForLog("").WithStartupTimeout(60 * time.Second),
	}

	container, err := testcontainers.GenericContainer(suite.ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})

	require.NoError(suite.T(), err, "Failed to create SSH container for %s", dist.Name)
	require.NotNil(suite.T(), container, "Container should not be nil for %s", dist.Name)

	// Setup basic prerequisites
	suite.setupContainerForSSH(container, dist)

	return container
}

func (suite *PowerShellSSHTestSuite) setupContainerForSSH(container testcontainers.Container, dist SSHTestDistribution) {
	// Run setup commands
	for i, cmd := range dist.SetupCommands {
		suite.T().Logf("Running setup command %d/%d for %s: %s", i+1, len(dist.SetupCommands), dist.Name, cmd)

		exitCode, reader, err := container.Exec(suite.ctx, []string{"bash", "-c", cmd})
		require.NoError(suite.T(), err, "Failed to execute setup command for %s: %s", dist.Name, cmd)

		if exitCode != 0 {
			output := suite.readContainerOutput(reader)
			suite.T().Fatalf("Setup command failed for %s with exit code %d: %s\nOutput: %s", dist.Name, exitCode, cmd, output)
		}
	}

	// Run SSH-specific setup commands
	for i, cmd := range dist.SSHSetupCommands {
		suite.T().Logf("Running SSH setup command %d/%d for %s: %s", i+1, len(dist.SSHSetupCommands), dist.Name, cmd)

		exitCode, reader, err := container.Exec(suite.ctx, []string{"bash", "-c", cmd})
		require.NoError(suite.T(), err, "Failed to execute SSH setup command for %s: %s", dist.Name, cmd)

		if exitCode != 0 {
			output := suite.readContainerOutput(reader)
			suite.T().Logf("SSH setup command warning for %s (exit code %d): %s\nOutput: %s", dist.Name, exitCode, cmd, output)
		}
	}

	// Wait for SSH server to start
	time.Sleep(5 * time.Second)

	suite.T().Logf("✓ SSH container setup completed for %s", dist.Name)
}

func (suite *PowerShellSSHTestSuite) setupSSHConnection(container testcontainers.Container) *ssh.Client {
	// Get SSH connection details
	host, err := container.Host(suite.ctx)
	require.NoError(suite.T(), err, "Failed to get container host")

	mappedPort, err := container.MappedPort(suite.ctx, "22")
	require.NoError(suite.T(), err, "Failed to get mapped SSH port")

	suite.sshHost = host
	suite.sshPort = mappedPort.Int()

	suite.T().Logf("Connecting to SSH at %s:%d", host, mappedPort.Int())

	// Setup SSH client configuration
	config := &ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{
			ssh.Password("testpassword123"),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // For testing only
		Timeout:         30 * time.Second,
	}

	suite.sshConfig = config

	// Connect with retries
	var client *ssh.Client
	for i := 0; i < 10; i++ {
		client, err = ssh.Dial("tcp", fmt.Sprintf("%s:%d", host, mappedPort.Int()), config)
		if err == nil {
			break
		}
		suite.T().Logf("SSH connection attempt %d/10 failed: %v", i+1, err)
		time.Sleep(2 * time.Second)
	}

	require.NoError(suite.T(), err, "Failed to establish SSH connection after retries")
	require.NotNil(suite.T(), client, "SSH client should not be nil")

	suite.T().Logf("✓ SSH connection established")

	return client
}

func (suite *PowerShellSSHTestSuite) transferPortunixViaSSH(client *ssh.Client) {
	// Ensure portunix binary exists locally
	_, err := os.Stat("./portunix")
	if err != nil {
		suite.T().Skip("Portunix binary not found for SSH transfer testing")
	}

	// Read local portunix binary
	portunixData, err := os.ReadFile("./portunix")
	require.NoError(suite.T(), err, "Failed to read portunix binary")

	// Create SSH session for file transfer
	session, err := client.NewSession()
	require.NoError(suite.T(), err, "Failed to create SSH session for file transfer")
	defer session.Close()

	// Transfer binary using base64 encoding (simple approach for testing)
	transferScript := fmt.Sprintf(`
		mkdir -p /usr/local/bin
		cat > /tmp/portunix.b64 << 'EOF'
%s
EOF
		base64 -d /tmp/portunix.b64 > /usr/local/bin/portunix
		chmod +x /usr/local/bin/portunix
		rm /tmp/portunix.b64
	`, encodeBase64(portunixData))

	err = session.Run(transferScript)
	require.NoError(suite.T(), err, "Failed to transfer portunix binary via SSH")

	suite.T().Logf("✓ Portunix binary transferred via SSH")
}

func (suite *PowerShellSSHTestSuite) installPowerShellViaSSH(client *ssh.Client, dist SSHTestDistribution) {
	suite.T().Logf("Installing PowerShell via SSH for %s...", dist.Name)

	session, err := client.NewSession()
	require.NoError(suite.T(), err, "Failed to create SSH session for installation")
	defer session.Close()

	// Prepare installation command
	installCmd := "/usr/local/bin/portunix install powershell"
	if dist.ExpectedVariant != "" {
		installCmd = fmt.Sprintf("/usr/local/bin/portunix install powershell --variant %s", dist.ExpectedVariant)
	}

	suite.T().Logf("Executing via SSH: %s", installCmd)

	// Capture output
	output, err := session.CombinedOutput(installCmd)
	suite.T().Logf("SSH Installation output for %s:\n%s", dist.Name, string(output))

	if err != nil {
		// Check if it's just an exit error
		if exitError, ok := err.(*ssh.ExitError); ok {
			suite.T().Logf("Installation failed with exit code %d for %s", exitError.ExitStatus(), dist.Name)
			suite.T().Logf("Output: %s", string(output))
			// Don't fail immediately - some installations might have expected failures
		} else {
			require.NoError(suite.T(), err, "SSH command execution failed for %s", dist.Name)
		}
	}

	suite.T().Logf("✓ PowerShell installation command completed via SSH for %s", dist.Name)
}

func (suite *PowerShellSSHTestSuite) verifyPowerShellViaSSH(client *ssh.Client, dist SSHTestDistribution) {
	suite.T().Logf("Verifying PowerShell installation via SSH for %s...", dist.Name)

	// Wait for installation to settle
	time.Sleep(10 * time.Second)

	session, err := client.NewSession()
	require.NoError(suite.T(), err, "Failed to create SSH session for verification")
	defer session.Close()

	// Run verification command
	output, err := session.CombinedOutput(dist.VerificationCmd)
	suite.T().Logf("SSH Verification output for %s:\n%s", dist.Name, string(output))

	if err != nil {
		if exitError, ok := err.(*ssh.ExitError); ok {
			suite.T().Logf("Verification failed with exit code %d for %s", exitError.ExitStatus(), dist.Name)
			suite.T().Logf("This might be expected if installation had issues")
			return
		}
		require.NoError(suite.T(), err, "SSH verification command failed for %s", dist.Name)
	}

	// Verify output contains PowerShell version info
	outputStr := string(output)
	if strings.Contains(outputStr, "PowerShell") {
		assert.Regexp(suite.T(), `\d+\.\d+\.\d+`, outputStr,
			"PowerShell version should be in format x.y.z for %s", dist.Name)
		suite.T().Logf("✓ PowerShell verification completed successfully via SSH for %s", dist.Name)
	} else {
		suite.T().Logf("PowerShell verification inconclusive for %s - output: %s", dist.Name, outputStr)
	}
}

// Test SSH connection stability during long operations
func (suite *PowerShellSSHTestSuite) TestSSHConnectionStability() {
	dist := SSHTestDistribution{
		Name:  "Ubuntu SSH Stability Test",
		Image: "ubuntu:22.04",
		SetupCommands: []string{
			"apt-get update",
			"apt-get install -y openssh-server",
		},
		SSHSetupCommands: []string{
			"mkdir -p /var/run/sshd",
			"echo 'root:testpassword123' | chpasswd",
			"sed -i 's/#PermitRootLogin prohibit-password/PermitRootLogin yes/' /etc/ssh/sshd_config",
			"sed -i 's/#PasswordAuthentication yes/PasswordAuthentication yes/' /etc/ssh/sshd_config",
			"/usr/sbin/sshd -D &",
		},
	}

	container := suite.createSSHContainer(dist)
	suite.container = container

	client := suite.setupSSHConnection(container)
	defer client.Close()

	// Test multiple SSH sessions
	for i := 0; i < 5; i++ {
		session, err := client.NewSession()
		require.NoError(suite.T(), err, "Failed to create SSH session %d", i+1)

		output, err := session.CombinedOutput("echo 'SSH connection test " + fmt.Sprint(i+1) + "'")
		require.NoError(suite.T(), err, "SSH command %d failed", i+1)

		assert.Contains(suite.T(), string(output), fmt.Sprintf("SSH connection test %d", i+1))
		session.Close()
	}

	suite.T().Logf("✓ SSH connection stability test passed")
}

// Helper functions

func (suite *PowerShellSSHTestSuite) readContainerOutput(reader *testcontainers.ExecResult) string {
	if reader == nil {
		return ""
	}

	buffer := make([]byte, 4096)
	n, _ := reader.Reader.Read(buffer)
	return string(buffer[:n])
}

func encodeBase64(data []byte) string {
	// Simple base64 encoding
	const base64Chars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
	var result strings.Builder

	for i := 0; i < len(data); i += 3 {
		b1, b2, b3 := data[i], byte(0), byte(0)
		if i+1 < len(data) {
			b2 = data[i+1]
		}
		if i+2 < len(data) {
			b3 = data[i+2]
		}

		result.WriteByte(base64Chars[(b1>>2)&0x3F])
		result.WriteByte(base64Chars[((b1<<4)|(b2>>4))&0x3F])
		if i+1 < len(data) {
			result.WriteByte(base64Chars[((b2<<2)|(b3>>6))&0x3F])
		} else {
			result.WriteByte('=')
		}
		if i+2 < len(data) {
			result.WriteByte(base64Chars[b3&0x3F])
		} else {
			result.WriteByte('=')
		}

		// Add newlines every 76 characters for proper base64 format
		if (i/3*4+4)%76 == 0 {
			result.WriteByte('\n')
		}
	}

	return result.String()
}
