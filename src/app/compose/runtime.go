package compose

import (
	"fmt"
	"os/exec"

	"portunix.ai/app/config"
)

// ComposeRuntime represents a detected compose tool
type ComposeRuntime struct {
	Name    string   // Display name (e.g., "Docker Compose V2")
	Command string   // Primary command (e.g., "docker")
	Args    []string // Arguments to invoke compose (e.g., ["compose"] for V2)
	Version string   // Detected version
}

// GetComposeRuntime detects and returns the available compose runtime
// Detection priority depends on configured container runtime:
// - If docker preferred: docker compose → docker-compose → podman-compose
// - If podman preferred: podman-compose → docker compose → docker-compose
func GetComposeRuntime() (*ComposeRuntime, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		// Default to docker-first order if config fails
		return detectComposeRuntime("docker")
	}

	preferredRuntime := cfg.ContainerRuntime
	if preferredRuntime == "" || preferredRuntime == "auto" {
		preferredRuntime = "docker"
	}

	return detectComposeRuntime(preferredRuntime)
}

// detectComposeRuntime attempts to find a compose tool based on preference
func detectComposeRuntime(preferred string) (*ComposeRuntime, error) {
	var detectors []func() *ComposeRuntime

	if preferred == "podman" {
		detectors = []func() *ComposeRuntime{
			detectPodmanCompose,
			detectDockerComposeV2,
			detectDockerComposeV1,
		}
	} else {
		// Default: docker first
		detectors = []func() *ComposeRuntime{
			detectDockerComposeV2,
			detectDockerComposeV1,
			detectPodmanCompose,
		}
	}

	for _, detect := range detectors {
		if runtime := detect(); runtime != nil {
			return runtime, nil
		}
	}

	return nil, fmt.Errorf("no compose tool detected")
}

// detectDockerComposeV2 checks for Docker Compose V2 (docker compose)
func detectDockerComposeV2() *ComposeRuntime {
	cmd := exec.Command("docker", "compose", "version", "--short")
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	version := string(output)
	if len(version) > 0 {
		// Trim newline
		if version[len(version)-1] == '\n' {
			version = version[:len(version)-1]
		}
	}

	return &ComposeRuntime{
		Name:    "Docker Compose V2",
		Command: "docker",
		Args:    []string{"compose"},
		Version: version,
	}
}

// detectDockerComposeV1 checks for Docker Compose V1 (docker-compose)
func detectDockerComposeV1() *ComposeRuntime {
	cmd := exec.Command("docker-compose", "version", "--short")
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	version := string(output)
	if len(version) > 0 {
		// Trim newline
		if version[len(version)-1] == '\n' {
			version = version[:len(version)-1]
		}
	}

	return &ComposeRuntime{
		Name:    "Docker Compose V1",
		Command: "docker-compose",
		Args:    []string{},
		Version: version,
	}
}

// detectPodmanCompose checks for podman-compose
func detectPodmanCompose() *ComposeRuntime {
	cmd := exec.Command("podman-compose", "version")
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	version := string(output)
	if len(version) > 0 {
		// Trim newline
		if version[len(version)-1] == '\n' {
			version = version[:len(version)-1]
		}
	}

	return &ComposeRuntime{
		Name:    "Podman Compose",
		Command: "podman-compose",
		Args:    []string{},
		Version: version,
	}
}

// IsDockerComposeV2Available checks if Docker Compose V2 is available
func IsDockerComposeV2Available() bool {
	return detectDockerComposeV2() != nil
}

// IsDockerComposeV1Available checks if Docker Compose V1 is available
func IsDockerComposeV1Available() bool {
	return detectDockerComposeV1() != nil
}

// IsPodmanComposeAvailable checks if podman-compose is available
func IsPodmanComposeAvailable() bool {
	return detectPodmanCompose() != nil
}

// GetComposeInfo returns information about available compose tools
func GetComposeInfo() map[string]bool {
	info := make(map[string]bool)

	info["docker-compose-v2"] = IsDockerComposeV2Available()
	info["docker-compose-v1"] = IsDockerComposeV1Available()
	info["podman-compose"] = IsPodmanComposeAvailable()

	return info
}

// GetInstallationInstructions returns help text for installing compose tools
func GetInstallationInstructions() string {
	return `No compose tool detected. Install one of the following:

For Docker:
  Docker Compose V2 (recommended): Included with Docker Desktop
  Docker Compose V1: portunix install docker-compose

For Podman:
  podman-compose: portunix install podman-compose

Tip: Use 'portunix container check' to verify container runtime availability.`
}
