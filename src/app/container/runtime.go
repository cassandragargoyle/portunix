package container

import (
	"fmt"
	"os/exec"

	"portunix.ai/app/config"
)

// GetSelectedRuntime returns the configured container runtime
func GetSelectedRuntime() (string, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return "", fmt.Errorf("failed to load configuration: %w", err)
	}

	// Use configured runtime or default to podman
	runtime := cfg.ContainerRuntime
	if runtime == "" {
		runtime = "podman"
	}

	// Check if selected runtime is available and running
	switch runtime {
	case "docker":
		if !IsDockerAvailable() {
			return "", fmt.Errorf("Docker not installed")
		}
		if !IsDockerRunning() {
			return "", fmt.Errorf("Docker installed but daemon not running")
		}
		return "docker", nil
	case "podman":
		if !IsPodmanAvailable() {
			return "", fmt.Errorf("Podman not installed")
		}
		if !IsPodmanRunning() {
			return "", fmt.Errorf("Podman installed but not functional")
		}
		return "podman", nil
	default:
		return "", fmt.Errorf("unknown container runtime: %s (must be 'docker' or 'podman')", runtime)
	}
}

// IsDockerAvailable checks if Docker is installed (binary exists in PATH)
func IsDockerAvailable() bool {
	_, err := exec.LookPath("docker")
	return err == nil
}

// IsDockerRunning checks if Docker daemon is running
func IsDockerRunning() bool {
	cmd := exec.Command("docker", "info")
	return cmd.Run() == nil
}

// IsPodmanAvailable checks if Podman is installed (binary exists in PATH)
func IsPodmanAvailable() bool {
	_, err := exec.LookPath("podman")
	return err == nil
}

// IsPodmanRunning checks if Podman is functional
func IsPodmanRunning() bool {
	cmd := exec.Command("podman", "info")
	return cmd.Run() == nil
}

// RuntimeStatus represents the status of a container runtime
type RuntimeStatus struct {
	Installed bool
	Running   bool
}

// GetRuntimeInfo returns information about available container runtimes
// Returns map with "installed" status (true if binary exists in PATH)
func GetRuntimeInfo() (map[string]bool, error) {
	info := make(map[string]bool)

	// For backward compatibility, return true only if both installed AND running
	info["docker"] = IsDockerAvailable() && IsDockerRunning()
	info["podman"] = IsPodmanAvailable() && IsPodmanRunning()

	return info, nil
}

// GetRuntimeDetailedInfo returns detailed information about container runtimes
func GetRuntimeDetailedInfo() map[string]RuntimeStatus {
	info := make(map[string]RuntimeStatus)

	info["docker"] = RuntimeStatus{
		Installed: IsDockerAvailable(),
		Running:   IsDockerRunning(),
	}
	info["podman"] = RuntimeStatus{
		Installed: IsPodmanAvailable(),
		Running:   IsPodmanRunning(),
	}

	return info
}

// ValidateRuntime checks if a runtime name is valid
func ValidateRuntime(runtime string) error {
	if runtime != "docker" && runtime != "podman" {
		return fmt.Errorf("invalid container runtime: %s (must be 'docker' or 'podman')", runtime)
	}
	return nil
}
