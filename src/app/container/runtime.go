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
	
	// Check if selected runtime is available
	switch runtime {
	case "docker":
		if IsDockerAvailable() {
			return "docker", nil
		}
		return "", fmt.Errorf("Docker not available (not installed or not running)")
	case "podman":
		if IsPodmanAvailable() {
			return "podman", nil
		}
		return "", fmt.Errorf("Podman not available (not installed)")
	default:
		return "", fmt.Errorf("unknown container runtime: %s (must be 'docker' or 'podman')", runtime)
	}
}

// IsDockerAvailable checks if Docker is available and running
func IsDockerAvailable() bool {
	cmd := exec.Command("docker", "version", "--format", "{{.Client.Version}}")
	return cmd.Run() == nil
}

// IsPodmanAvailable checks if Podman is available
func IsPodmanAvailable() bool {
	cmd := exec.Command("podman", "version", "--format", "{{.Client.Version}}")
	return cmd.Run() == nil
}

// GetRuntimeInfo returns information about available container runtimes
func GetRuntimeInfo() (map[string]bool, error) {
	info := make(map[string]bool)
	
	info["docker"] = IsDockerAvailable()
	info["podman"] = IsPodmanAvailable()
	
	return info, nil
}

// ValidateRuntime checks if a runtime name is valid
func ValidateRuntime(runtime string) error {
	if runtime != "docker" && runtime != "podman" {
		return fmt.Errorf("invalid container runtime: %s (must be 'docker' or 'podman')", runtime)
	}
	return nil
}