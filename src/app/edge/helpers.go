package edge

import (
	"fmt"
	"os/exec"
	"strings"
)

// Helper methods for container and system operations

func (m *Manager) getContainerStatus(containerName string) string {
	cmd := exec.Command("podman", "ps", "--filter", fmt.Sprintf("name=%s", containerName), "--format", "{{.Status}}")
	output, err := cmd.Output()
	if err != nil {
		// Try with docker if podman fails
		cmd = exec.Command("docker", "ps", "--filter", fmt.Sprintf("name=%s", containerName), "--format", "{{.Status}}")
		output, err = cmd.Output()
		if err != nil {
			return ""
		}
	}

	return strings.TrimSpace(string(output))
}

func (m *Manager) startContainer(containerName string) error {
	cmd := exec.Command("podman", "start", containerName)
	if err := cmd.Run(); err != nil {
		// Try with docker if podman fails
		cmd = exec.Command("docker", "start", containerName)
		return cmd.Run()
	}
	return nil
}

func (m *Manager) stopContainer(containerName string) error {
	cmd := exec.Command("podman", "stop", containerName)
	if err := cmd.Run(); err != nil {
		// Try with docker if podman fails
		cmd = exec.Command("docker", "stop", containerName)
		return cmd.Run()
	}
	return nil
}

func (m *Manager) showContainerLogs(containerName string, follow bool, tail int) error {
	args := []string{"logs"}
	if follow {
		args = append(args, "-f")
	}
	if tail > 0 {
		args = append(args, "--tail", fmt.Sprintf("%d", tail))
	}
	args = append(args, containerName)

	cmd := exec.Command("podman", args...)
	if err := cmd.Run(); err != nil {
		// Try with docker if podman fails
		cmd = exec.Command("docker", args...)
		return cmd.Run()
	}
	return nil
}

func (m *Manager) runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = nil // Don't capture output, let it print directly
	cmd.Stderr = nil
	return cmd.Run()
}

func (m *Manager) getSystemUptime() string {
	cmd := exec.Command("uptime", "-p")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}
