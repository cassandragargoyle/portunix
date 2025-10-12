package container

import (
	"bytes"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"sync"
	"time"
)

// RuntimeType represents the type of container runtime
type RuntimeType string

const (
	RuntimeDocker RuntimeType = "docker"
	RuntimePodman RuntimeType = "podman"
	RuntimeBoth   RuntimeType = "both"
	RuntimeNone   RuntimeType = "none"
)

// RuntimeInfo contains information about a specific container runtime
type RuntimeInfo struct {
	Available bool
	Version   string
	Path      string
	Features  map[string]bool
}

// ContainerCapabilities represents the container runtime capabilities of the system
type ContainerCapabilities struct {
	Available     bool
	Runtime       RuntimeType
	DockerInfo    *RuntimeInfo
	PodmanInfo    *RuntimeInfo
	Preferred     RuntimeType
	Features      map[string]bool
	LastChecked   time.Time
	CheckDuration time.Duration
}

var (
	cachedCapabilities *ContainerCapabilities
	cacheMutex         sync.RWMutex
	cacheExpiry        = 5 * time.Minute
)

// GetContainerCapabilities detects and returns container runtime capabilities
func GetContainerCapabilities() (*ContainerCapabilities, error) {
	return getCapabilities(false)
}

// RefreshContainerCapabilities forces a refresh of container capabilities
func RefreshContainerCapabilities() (*ContainerCapabilities, error) {
	return getCapabilities(true)
}

// HasContainerRuntime provides a simple check for test integration
func HasContainerRuntime() bool {
	caps, err := GetContainerCapabilities()
	if err != nil {
		return false
	}
	return caps.Available
}

// HasDocker checks if Docker is available
func HasDocker() bool {
	caps, err := GetContainerCapabilities()
	if err != nil {
		return false
	}
	return caps.DockerInfo != nil && caps.DockerInfo.Available
}

// HasPodman checks if Podman is available
func HasPodman() bool {
	caps, err := GetContainerCapabilities()
	if err != nil {
		return false
	}
	return caps.PodmanInfo != nil && caps.PodmanInfo.Available
}

// GetPreferredRuntime returns the preferred container runtime
func GetPreferredRuntime() RuntimeType {
	caps, err := GetContainerCapabilities()
	if err != nil {
		return RuntimeNone
	}
	return caps.Preferred
}

func getCapabilities(forceRefresh bool) (*ContainerCapabilities, error) {
	cacheMutex.RLock()
	if !forceRefresh && cachedCapabilities != nil {
		if time.Since(cachedCapabilities.LastChecked) < cacheExpiry {
			defer cacheMutex.RUnlock()
			return cachedCapabilities, nil
		}
	}
	cacheMutex.RUnlock()

	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	// Double-check after acquiring write lock
	if !forceRefresh && cachedCapabilities != nil {
		if time.Since(cachedCapabilities.LastChecked) < cacheExpiry {
			return cachedCapabilities, nil
		}
	}

	startTime := time.Now()
	caps := &ContainerCapabilities{
		Features:    make(map[string]bool),
		LastChecked: startTime,
	}

	// Detect Docker
	caps.DockerInfo = detectDocker()
	
	// Detect Podman
	caps.PodmanInfo = detectPodman()

	// Determine overall runtime status
	if caps.DockerInfo.Available && caps.PodmanInfo.Available {
		caps.Runtime = RuntimeBoth
		caps.Available = true
		// Prefer Docker if both are available (can be configured later)
		caps.Preferred = RuntimeDocker
	} else if caps.DockerInfo.Available {
		caps.Runtime = RuntimeDocker
		caps.Available = true
		caps.Preferred = RuntimeDocker
	} else if caps.PodmanInfo.Available {
		caps.Runtime = RuntimePodman
		caps.Available = true
		caps.Preferred = RuntimePodman
	} else {
		caps.Runtime = RuntimeNone
		caps.Available = false
		caps.Preferred = RuntimeNone
	}

	// Aggregate features
	if caps.DockerInfo.Available {
		for feature, supported := range caps.DockerInfo.Features {
			if supported {
				caps.Features[feature] = true
			}
		}
	}
	if caps.PodmanInfo.Available {
		for feature, supported := range caps.PodmanInfo.Features {
			if supported {
				caps.Features[feature] = true
			}
		}
	}

	caps.CheckDuration = time.Since(startTime)
	cachedCapabilities = caps

	return caps, nil
}

func detectDocker() *RuntimeInfo {
	info := &RuntimeInfo{
		Features: make(map[string]bool),
	}

	// Find Docker executable
	path, err := exec.LookPath("docker")
	if err != nil {
		return info
	}
	info.Path = path

	// Get Docker version
	cmd := exec.Command("docker", "version", "--format", "{{.Server.Version}}")
	output, err := cmd.Output()
	if err != nil {
		// Fallback to regular version command
		cmd = exec.Command("docker", "--version")
		output, err = cmd.Output()
		if err != nil {
			return info
		}
		// Parse version from "Docker version 24.0.5, build ced0996"
		versionRegex := regexp.MustCompile(`Docker version ([0-9]+\.[0-9]+\.[0-9]+)`)
		matches := versionRegex.FindSubmatch(output)
		if len(matches) > 1 {
			info.Version = string(matches[1])
			info.Available = true
		}
	} else {
		info.Version = strings.TrimSpace(string(output))
		info.Available = true
	}

	if !info.Available {
		return info
	}

	// Check for Docker Compose
	info.Features["compose"] = checkDockerCompose()
	
	// Check for BuildKit/Buildx
	info.Features["buildx"] = checkDockerBuildx()
	
	// Check for volume support (always true if Docker is available)
	info.Features["volumes"] = true
	
	// Check for network support
	info.Features["networks"] = true
	
	// Check if Docker daemon is running
	cmd = exec.Command("docker", "info")
	if err := cmd.Run(); err == nil {
		info.Features["daemon_running"] = true
	}

	return info
}

func detectPodman() *RuntimeInfo {
	info := &RuntimeInfo{
		Features: make(map[string]bool),
	}

	// Find Podman executable
	path, err := exec.LookPath("podman")
	if err != nil {
		return info
	}
	info.Path = path

	// Get Podman version
	cmd := exec.Command("podman", "version", "--format", "{{.Version}}")
	output, err := cmd.Output()
	if err != nil {
		// Fallback to regular version command
		cmd = exec.Command("podman", "--version")
		output, err = cmd.Output()
		if err != nil {
			return info
		}
		// Parse version from "podman version 4.6.1"
		versionRegex := regexp.MustCompile(`podman version ([0-9]+\.[0-9]+\.[0-9]+)`)
		matches := versionRegex.FindSubmatch(output)
		if len(matches) > 1 {
			info.Version = string(matches[1])
			info.Available = true
		}
	} else {
		info.Version = strings.TrimSpace(string(output))
		info.Available = true
	}

	if !info.Available {
		return info
	}

	// Check for Podman Compose
	info.Features["compose"] = checkPodmanCompose()
	
	// Check for BuildKit support (Podman uses buildah)
	info.Features["buildx"] = false // Podman doesn't have buildx
	
	// Check for volume support
	info.Features["volumes"] = true
	
	// Check for network support
	info.Features["networks"] = true
	
	// Check if Podman is functional
	cmd = exec.Command("podman", "info")
	if err := cmd.Run(); err == nil {
		info.Features["daemon_running"] = true // Podman doesn't require daemon but this indicates it's functional
	}

	// Check for Docker compatibility mode
	cmd = exec.Command("podman", "version", "--format", "{{.Client.APIVersion}}")
	if output, err := cmd.Output(); err == nil && len(output) > 0 {
		info.Features["docker_compat"] = true
	}

	return info
}

func checkDockerCompose() bool {
	// Check for docker compose (v2)
	cmd := exec.Command("docker", "compose", "version")
	if err := cmd.Run(); err == nil {
		return true
	}
	
	// Check for docker-compose (v1)
	if path, err := exec.LookPath("docker-compose"); err == nil && path != "" {
		return true
	}
	
	return false
}

func checkDockerBuildx() bool {
	cmd := exec.Command("docker", "buildx", "version")
	return cmd.Run() == nil
}

func checkPodmanCompose() bool {
	// Check for podman-compose
	if path, err := exec.LookPath("podman-compose"); err == nil && path != "" {
		return true
	}
	
	// Check for podman compose (integrated)
	cmd := exec.Command("podman", "compose", "version")
	return cmd.Run() == nil
}

// FormatCapabilities returns a formatted string representation of capabilities
func (c *ContainerCapabilities) FormatCapabilities() string {
	var buf bytes.Buffer
	
	buf.WriteString("Container Runtime Status:\n")
	
	// Docker status
	if c.DockerInfo != nil {
		if c.DockerInfo.Available {
			buf.WriteString(fmt.Sprintf("  Docker: ✓ Available (version %s)\n", c.DockerInfo.Version))
		} else {
			buf.WriteString("  Docker: ✗ Not available\n")
		}
	} else {
		buf.WriteString("  Docker: ✗ Not detected\n")
	}
	
	// Podman status
	if c.PodmanInfo != nil {
		if c.PodmanInfo.Available {
			buf.WriteString(fmt.Sprintf("  Podman: ✓ Available (version %s)\n", c.PodmanInfo.Version))
		} else {
			buf.WriteString("  Podman: ✗ Not available\n")
		}
	} else {
		buf.WriteString("  Podman: ✗ Not detected\n")
	}
	
	// Preferred runtime
	if c.Available {
		buf.WriteString(fmt.Sprintf("  Preferred: %s\n", c.Preferred))
	}
	
	// Capabilities
	if len(c.Features) > 0 {
		buf.WriteString("\nCapabilities:\n")
		features := []string{
			"compose", "buildx", "volumes", "networks", "daemon_running", "docker_compat",
		}
		for _, feature := range features {
			if supported, exists := c.Features[feature]; exists && supported {
				buf.WriteString(fmt.Sprintf("  - %s: ✓\n", formatFeatureName(feature)))
			}
		}
	}
	
	return buf.String()
}

func formatFeatureName(feature string) string {
	switch feature {
	case "compose":
		return "Compose support"
	case "buildx":
		return "BuildKit/Buildx"
	case "volumes":
		return "Volume mounting"
	case "networks":
		return "Network creation"
	case "daemon_running":
		return "Runtime active"
	case "docker_compat":
		return "Docker compatibility"
	default:
		return feature
	}
}