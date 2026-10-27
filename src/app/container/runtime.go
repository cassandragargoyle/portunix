/*
 *  SPDX-License-Identifier: MIT
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

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

	return chooseRuntime(cfg.ContainerRuntime, systemRuntimeChecks)
}

// runtimeChecks reports whether a runtime is installed and usable; a
// parameter so the selection rules can be tested without real runtimes
type runtimeChecks struct {
	dockerAvailable func() bool
	dockerRunning   func() bool
	podmanAvailable func() bool
	podmanRunning   func() bool
}

// systemRuntimeChecks probes the runtimes installed on this host
var systemRuntimeChecks = runtimeChecks{
	dockerAvailable: IsDockerAvailable,
	dockerRunning:   IsDockerRunning,
	podmanAvailable: IsPodmanAvailable,
	podmanRunning:   IsPodmanRunning,
}

// chooseRuntime applies the selection rules: the configured runtime, or
// podman when none is configured, and only if it is installed and usable
func chooseRuntime(configured string, checks runtimeChecks) (string, error) {
	runtime := configured
	if runtime == "" {
		runtime = "podman"
	}

	// Check if selected runtime is available and running
	switch runtime {
	case "docker":
		if !checks.dockerAvailable() {
			return "", fmt.Errorf("Docker not installed")
		}
		if !checks.dockerRunning() {
			return "", fmt.Errorf("Docker installed but daemon not running")
		}
		return "docker", nil
	case "podman":
		if !checks.podmanAvailable() {
			return "", fmt.Errorf("Podman not installed")
		}
		if !checks.podmanRunning() {
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
