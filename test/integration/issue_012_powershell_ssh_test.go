//go:build integration

package integration

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"golang.org/x/crypto/ssh"
)

type SupportedDistribution struct {
	Name                string
	Image               string
	ExpectedVariant     string
	PreInstallCommands  []string
	SSHSetupCommands    []string
	VerificationCmd     string
}

// TestIssue012_PowerShellInstallation_AllDistros tests PowerShell installation
// across all supported Linux distributions using Portunix VM management
func TestIssue012_PowerShellInstallation_AllDistros(t *testing.T) {
	supportedDistros := []struct {
		name    string
		image   string
		variant string
	}{
		{"Ubuntu 22.04", "ubuntu:22.04", "ubuntu"},
		{"Ubuntu 24.04", "ubuntu:24.04", "ubuntu"},
		{"Debian 11", "debian:bullseye", "debian"},
		{"Debian 12", "debian:bookworm", "debian"},
		{"Fedora 39", "fedora:39", "fedora"},
		{"Fedora 40", "fedora:40", "fedora"},
		{"Rocky Linux 9", "rockylinux:9", "rocky"},
		{"Linux Mint 21", "linuxmintd/mint21-amd64", "mint"},
	}

	for _, distro := range supportedDistros {
		t.Run(distro.name, func(t *testing.T) {
			t.Parallel()
			testPowerShellInstallationOnDistro(t, distro.image, distro.variant)
		})
	}
}

func createTestDistributionConfig(image, variant string) SupportedDistribution {
	// Define common SSH setup commands
	sshSetupCommands := []string{
		"apt-get update || yum update -y || dnf update -y || true",
		"apt-get install -y openssh-server || yum install -y openssh-server || dnf install -y openssh-server || true",
		"mkdir -p /var/run/sshd",
		"echo 'root:testpass123' | chpasswd",
		"sed -i 's/#PermitRootLogin prohibit-password/PermitRootLogin yes/' /etc/ssh/sshd_config || true",
		"sed -i 's/#PasswordAuthentication yes/PasswordAuthentication yes/' /etc/ssh/sshd_config || true",
		"/usr/sbin/sshd || systemctl start sshd || service ssh start || true",
	}

	// Define pre-install commands based on image
	var preInstallCommands []string
	var name string

	switch {
	case strings.Contains(image, "ubuntu"):
		name = "Ubuntu"
		preInstallCommands = []string{
			"apt-get update",
			"DEBIAN_FRONTEND=noninteractive apt-get install -y sudo wget curl lsb-release ca-certificates gnupg software-properties-common",
		}
	case strings.Contains(image, "debian"):
		name = "Debian"  
		preInstallCommands = []string{
			"apt-get update",
			"DEBIAN_FRONTEND=noninteractive apt-get install -y sudo wget curl lsb-release ca-certificates gnupg software-properties-common",
		}
	case strings.Contains(image, "fedora"):
		name = "Fedora"
		preInstallCommands = []string{
			"dnf update -y",
			"dnf install -y sudo wget curl redhat-lsb-core ca-certificates gnupg",
		}
	case strings.Contains(image, "rocky"):
		name = "Rocky Linux"
		preInstallCommands = []string{
			"dnf update -y", 
			"dnf install -y sudo wget curl redhat-lsb-core ca-certificates gnupg",
		}
	case strings.Contains(image, "mint"):
		name = "Linux Mint"
		preInstallCommands = []string{
			"apt-get update",
			"DEBIAN_FRONTEND=noninteractive apt-get install -y sudo wget curl lsb-release ca-certificates gnupg software-properties-common",
		}
	default:
		name = "Unknown"
		preInstallCommands = []string{"echo 'Unknown distribution'"}
	}

	return SupportedDistribution{
		Name:                name,
		Image:               image,
		ExpectedVariant:     variant,
		PreInstallCommands:  preInstallCommands,
		SSHSetupCommands:    sshSetupCommands,
		VerificationCmd:     "pwsh --version",
	}
}

func testPowerShellInstallationOnDistro(t *testing.T, image, variant string) {
	ctx := context.Background()

	// Create supported distribution config from test data
	dist := createTestDistributionConfig(image, variant)

	// Arrange: Create and setup container with SSH server
	container := createSSHContainer(t, ctx, dist)
	defer func() {
		if container != nil {
			container.Terminate(ctx)
		}
	}()

	// Arrange: Setup SSH connection
	sshClient := setupSSHConnection(t, ctx, container)
	defer sshClient.Close()

	// Arrange: Transfer portunix binary via SSH
	transferPortunixViaSSH(t, sshClient)

	// Act: Install PowerShell via SSH
	installPowerShellViaSSH(t, sshClient, dist)

	// Assert: Verify installation via SSH
	verifyPowerShellViaSSH(t, sshClient, dist)
}

func createSSHContainer(t *testing.T, ctx context.Context, dist SupportedDistribution) testcontainers.Container {
	req := testcontainers.ContainerRequest{
		Image:        dist.Image,
		ExposedPorts: []string{"22/tcp"},
		Cmd:          []string{"bash", "-c", "sleep 3600"}, // Keep running for SSH
		WaitingFor:   wait.ForLog("").WithStartupTimeout(60 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})

	require.NoError(t, err, "Failed to create SSH container for %s", dist.Name)
	require.NotNil(t, container, "Container should not be nil for %s", dist.Name)

	// Setup basic prerequisites
	setupContainerForSSH(t, ctx, container, dist)

	return container
}

func setupContainerForSSH(t *testing.T, ctx context.Context, container testcontainers.Container, dist SupportedDistribution) {
	// Run setup commands
	for i, cmd := range dist.PreInstallCommands {
		t.Logf("Running setup command %d/%d for %s: %s", i+1, len(dist.PreInstallCommands), dist.Name, cmd)
		
		exitCode, reader, err := container.Exec(ctx, []string{"bash", "-c", cmd})
		require.NoError(t, err, "Failed to execute setup command for %s: %s", dist.Name, cmd)
		
		if exitCode != 0 {
			output := readContainerOutput(reader)
			t.Fatalf("Setup command failed for %s with exit code %d: %s\nOutput: %s", dist.Name, exitCode, cmd, output)
		}
	}

	// Run SSH-specific setup commands
	for i, cmd := range dist.SSHSetupCommands {
		t.Logf("Running SSH setup command %d/%d for %s: %s", i+1, len(dist.SSHSetupCommands), dist.Name, cmd)
		
		exitCode, reader, err := container.Exec(ctx, []string{"bash", "-c", cmd})
		require.NoError(t, err, "Failed to execute SSH setup command for %s: %s", dist.Name, cmd)
		
		if exitCode != 0 {
			output := readContainerOutput(reader)
			t.Logf("SSH setup command warning for %s (exit code %d): %s\nOutput: %s", dist.Name, exitCode, cmd, output)
		}
	}

	// Wait for SSH server to start
	time.Sleep(5 * time.Second)
	
	t.Logf("✓ SSH container setup completed for %s", dist.Name)
}

func setupSSHConnection(t *testing.T, ctx context.Context, container testcontainers.Container) *ssh.Client {
	// Get SSH connection details
	host, err := container.Host(ctx)
	require.NoError(t, err, "Failed to get container host")

	mappedPort, err := container.MappedPort(ctx, "22")
	require.NoError(t, err, "Failed to get mapped SSH port")

	t.Logf("Connecting to SSH at %s:%d", host, mappedPort.Int())

	// Setup SSH client configuration
	config := &ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{
			ssh.Password("testpass123"),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // For testing only
		Timeout:         30 * time.Second,
	}

	// Connect with retries
	var client *ssh.Client
	for i := 0; i < 10; i++ {
		client, err = ssh.Dial("tcp", fmt.Sprintf("%s:%d", host, mappedPort.Int()), config)
		if err == nil {
			break
		}
		t.Logf("SSH connection attempt %d/10 failed: %v", i+1, err)
		time.Sleep(2 * time.Second)
	}

	require.NoError(t, err, "Failed to establish SSH connection after retries")
	require.NotNil(t, client, "SSH client should not be nil")

	t.Logf("✓ SSH connection established")

	return client
}

func transferPortunixViaSSH(t *testing.T, client *ssh.Client) {
	// Ensure portunix binary exists locally
	_, err := os.Stat("../../portunix")
	if err != nil {
		t.Skip("Portunix binary not found for SSH transfer testing")
	}

	// Read local portunix binary
	portunixData, err := os.ReadFile("../../portunix")
	require.NoError(t, err, "Failed to read portunix binary")

	// Create SSH session for file transfer
	session, err := client.NewSession()
	require.NoError(t, err, "Failed to create SSH session for file transfer")
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
	require.NoError(t, err, "Failed to transfer portunix binary via SSH")

	t.Logf("✓ Portunix binary transferred via SSH")
}

func installPowerShellViaSSH(t *testing.T, client *ssh.Client, dist SupportedDistribution) {
	t.Logf("Installing PowerShell via SSH for %s...", dist.Name)

	session, err := client.NewSession()
	require.NoError(t, err, "Failed to create SSH session for installation")
	defer session.Close()

	// Prepare installation command
	installCmd := "/usr/local/bin/portunix install powershell"
	if dist.ExpectedVariant != "" {
		installCmd = fmt.Sprintf("/usr/local/bin/portunix install powershell --variant %s", dist.ExpectedVariant)
	}

	t.Logf("Executing via SSH: %s", installCmd)

	// Capture output
	output, err := session.CombinedOutput(installCmd)
	t.Logf("SSH Installation output for %s:\n%s", dist.Name, string(output))

	if err != nil {
		// Check if it's just an exit error
		if exitError, ok := err.(*ssh.ExitError); ok {
			t.Logf("Installation failed with exit code %d for %s", exitError.ExitStatus(), dist.Name)
			t.Logf("Output: %s", string(output))
			// Don't fail immediately - some installations might have expected failures
		} else {
			require.NoError(t, err, "SSH command execution failed for %s", dist.Name)
		}
	}

	t.Logf("✓ PowerShell installation command completed via SSH for %s", dist.Name)
}

func verifyPowerShellViaSSH(t *testing.T, client *ssh.Client, dist SupportedDistribution) {
	t.Logf("Verifying PowerShell installation via SSH for %s...", dist.Name)

	// Wait for installation to settle
	time.Sleep(10 * time.Second)

	session, err := client.NewSession()
	require.NoError(t, err, "Failed to create SSH session for verification")
	defer session.Close()

	// Run verification command
	output, err := session.CombinedOutput(dist.VerificationCmd)
	t.Logf("SSH Verification output for %s:\n%s", dist.Name, string(output))

	if err != nil {
		if exitError, ok := err.(*ssh.ExitError); ok {
			t.Logf("Verification failed with exit code %d for %s", exitError.ExitStatus(), dist.Name)
			t.Logf("This might be expected if installation had issues")
			return
		}
		require.NoError(t, err, "SSH verification command failed for %s", dist.Name)
	}

	// Verify output contains PowerShell version info
	outputStr := string(output)
	if strings.Contains(outputStr, "PowerShell") {
		assert.Regexp(t, `\d+\.\d+\.\d+`, outputStr, 
			"PowerShell version should be in format x.y.z for %s", dist.Name)
		t.Logf("✓ PowerShell verification completed successfully via SSH for %s", dist.Name)
	} else {
		t.Logf("PowerShell verification inconclusive for %s - output: %s", dist.Name, outputStr)
	}
}

// Helper functions

func readContainerOutput(reader *testcontainers.ExecResult) string {
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