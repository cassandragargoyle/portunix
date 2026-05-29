package system

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
)

// execTimeout bounds external command probes so a stuck binary cannot block
// system info collection. 1.5s is generous for "give me a version string".
const execTimeout = 1500 * time.Millisecond

// runWithTimeout executes a command with a deadline and returns stdout.
func runWithTimeout(timeout time.Duration, name string, args ...string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, name, args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = nil
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return out.String(), nil
}

// GetQEMUVersion returns the version of QEMU if available
func GetQEMUVersion() string {
	var qemuBinary string
	if runtime.GOOS == "windows" {
		qemuBinary = "qemu-system-x86_64.exe"
	} else {
		qemuBinary = "qemu-system-x86_64"
	}

	out, err := runWithTimeout(execTimeout, qemuBinary, "--version")
	if err != nil {
		if runtime.GOOS == "windows" {
			return ""
		}
		out, err = runWithTimeout(execTimeout, "kvm", "--version")
		if err != nil {
			return ""
		}
	}

	// Example: "QEMU emulator version 8.0.0 (Debian 1:8.0+dfsg-1)"
	re := regexp.MustCompile(`version\s+(\d+\.\d+\.\d+)`)
	matches := re.FindStringSubmatch(strings.TrimSpace(out))
	if len(matches) > 1 {
		return "v" + matches[1]
	}
	return ""
}

// GetVirtualBoxVersion returns the version of VirtualBox if available
func GetVirtualBoxVersion() string {
	var vboxBinary string
	if runtime.GOOS == "windows" {
		vboxBinary = "VBoxManage.exe"
	} else {
		vboxBinary = "VBoxManage"
	}

	out, err := runWithTimeout(execTimeout, vboxBinary, "--version")
	if err != nil {
		return ""
	}

	// Example: "7.0.12r159484" or "7.0.12_Ubuntur159484"
	re := regexp.MustCompile(`^(\d+\.\d+\.\d+)`)
	matches := re.FindStringSubmatch(strings.TrimSpace(out))
	if len(matches) > 1 {
		return "v" + matches[1]
	}
	return ""
}

// GetLibvirtVersion returns the version of Libvirt if available
func GetLibvirtVersion() string {
	if runtime.GOOS == "windows" {
		return "" // Libvirt is not typically available on Windows
	}

	out, err := runWithTimeout(execTimeout, "virsh", "--version")
	if err != nil {
		return ""
	}

	// Example: "9.0.0"
	version := strings.TrimSpace(out)
	if version != "" && !strings.HasPrefix(version, "v") {
		version = "v" + version
	}
	return version
}

// GetDockerVersion returns the version of Docker if available.
// Tries the lightweight `docker --version` first (no daemon round-trip);
// only falls back to `docker version --format ...` when needed.
func GetDockerVersion() string {
	if out, err := runWithTimeout(execTimeout, "docker", "--version"); err == nil {
		if v := parseDockerVersionOutput(out); v != "" {
			return v
		}
	}
	if out, err := runWithTimeout(execTimeout, "docker", "version", "--format", "{{.Server.Version}}"); err == nil {
		return strings.TrimSpace(out)
	}
	return ""
}

// parseDockerVersionOutput extracts version number from "docker --version" output
// e.g. "Docker version 27.5.1, build 9f9e405" -> "27.5.1"
func parseDockerVersionOutput(output string) string {
	output = strings.TrimSpace(output)
	re := regexp.MustCompile(`\d+\.\d+\.\d+`)
	if match := re.FindString(output); match != "" {
		return match
	}
	return ""
}

// GetPodmanVersion returns the version of Podman if available.
// `podman --version` is local-only and fast; `podman version --format ...`
// can hit the daemon and is therefore used only as fallback.
func GetPodmanVersion() string {
	if out, err := runWithTimeout(execTimeout, "podman", "--version"); err == nil {
		// Parse from "podman version 4.6.1"
		parts := strings.Fields(strings.TrimSpace(out))
		if len(parts) >= 3 {
			return parts[2]
		}
	}
	if out, err := runWithTimeout(execTimeout, "podman", "version", "--format", "{{.Version}}"); err == nil {
		return strings.TrimSpace(out)
	}
	return ""
}

// socketReachable returns true if path is a Unix socket that accepts a
// connection within the given timeout. Stale or dead sockets return false.
func socketReachable(path string, timeout time.Duration) bool {
	fi, err := os.Stat(path)
	if err != nil {
		return false
	}
	if fi.Mode()&os.ModeSocket == 0 {
		return false
	}
	conn, err := net.DialTimeout("unix", path, timeout)
	if err != nil {
		return false
	}
	conn.Close()
	return true
}

// dockerSocketCandidates returns paths where the Docker daemon socket may live.
func dockerSocketCandidates() []string {
	if dh := os.Getenv("DOCKER_HOST"); strings.HasPrefix(dh, "unix://") {
		return []string{strings.TrimPrefix(dh, "unix://")}
	}
	candidates := []string{"/var/run/docker.sock"}
	if home, err := os.UserHomeDir(); err == nil {
		candidates = append(candidates,
			filepath.Join(home, ".docker", "run", "docker.sock"),
			filepath.Join(home, ".docker", "desktop", "docker.sock"),
		)
	}
	return candidates
}

// podmanSocketCandidates returns paths where the Podman API socket may live.
func podmanSocketCandidates() []string {
	var candidates []string
	if xdg := os.Getenv("XDG_RUNTIME_DIR"); xdg != "" {
		candidates = append(candidates, filepath.Join(xdg, "podman", "podman.sock"))
	}
	if uid := os.Getuid(); uid > 0 {
		candidates = append(candidates, fmt.Sprintf("/run/user/%d/podman/podman.sock", uid))
	}
	candidates = append(candidates, "/run/podman/podman.sock")
	return candidates
}

// IsDockerDaemonRunning reports whether the Docker daemon is reachable.
// Linux/macOS: probes well-known sockets directly (microseconds).
// Windows: falls back to `docker info` with a short deadline.
func IsDockerDaemonRunning() bool {
	if runtime.GOOS != "windows" {
		for _, p := range dockerSocketCandidates() {
			if socketReachable(p, 200*time.Millisecond) {
				return true
			}
		}
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), execTimeout)
	defer cancel()
	return exec.CommandContext(ctx, "docker", "info").Run() == nil
}

// IsPodmanSocketRunning reports whether the Podman API socket is listening.
// Direct socket probe on Linux/macOS; `podman info` fallback on Windows.
func IsPodmanSocketRunning() bool {
	if runtime.GOOS != "windows" {
		for _, p := range podmanSocketCandidates() {
			if socketReachable(p, 200*time.Millisecond) {
				return true
			}
		}
		return false
	}
	ctx, cancel := context.WithTimeout(context.Background(), execTimeout)
	defer cancel()
	return exec.CommandContext(ctx, "podman", "info").Run() == nil
}

// DetectComposeInfo detects container compose tool availability
func DetectComposeInfo(dockerInstalled, dockerDaemonRunning, podmanInstalled, podmanSocketRunning bool) *ComposeInfo {
	info := &ComposeInfo{}

	// Try Docker Compose first (if Docker is installed and daemon is running)
	if dockerInstalled {
		if dockerDaemonRunning {
			// Docker Compose V2 (docker compose)
			if out, err := runWithTimeout(execTimeout, "docker", "compose", "version", "--short"); err == nil {
				info.Available = true
				info.Type = "Docker Compose"
				info.Version = strings.TrimSpace(out)
				info.DaemonReady = true
				return info
			}

			// Docker Compose V1 (docker-compose)
			if out, err := runWithTimeout(execTimeout, "docker-compose", "--version"); err == nil {
				info.Available = true
				info.Type = "Docker Compose (V1)"
				// Parse version from "docker-compose version 1.29.2, build 5becea4c"
				if parts := strings.Fields(strings.TrimSpace(out)); len(parts) >= 3 {
					info.Version = strings.TrimSuffix(parts[2], ",")
				}
				info.DaemonReady = true
				return info
			}
		} else {
			info.WarningMessage = "Docker installed but daemon not running"
		}
	}

	// Try Podman Compose (if Podman is installed)
	if podmanInstalled {
		// Built-in podman compose (Podman 3.0+)
		if out, err := runWithTimeout(execTimeout, "podman", "compose", "version"); err == nil {
			info.Available = true
			info.Type = "Podman Compose"
			output := strings.TrimSpace(out)
			if strings.Contains(output, "version") {
				parts := strings.Fields(output)
				for i, p := range parts {
					if p == "version" && i+1 < len(parts) {
						info.Version = parts[i+1]
						break
					}
				}
			}
			if podmanSocketRunning {
				info.DaemonReady = true
			} else {
				info.DaemonReady = false
				info.WarningMessage = "Podman socket not running - start with: systemctl --user start podman.socket"
			}
			return info
		}

		// Standalone podman-compose
		if out, err := runWithTimeout(execTimeout, "podman-compose", "--version"); err == nil {
			info.Available = true
			info.Type = "podman-compose"
			// Parse version from "podman-compose version: 1.0.6"
			output := strings.TrimSpace(out)
			if strings.Contains(output, ":") {
				parts := strings.Split(output, ":")
				if len(parts) >= 2 {
					info.Version = strings.TrimSpace(parts[1])
				}
			}
			if podmanSocketRunning {
				info.DaemonReady = true
			} else {
				info.DaemonReady = false
				info.WarningMessage = "Podman socket not running - start with: systemctl --user start podman.socket"
			}
			return info
		}

		// Podman installed but no compose
		if info.WarningMessage == "" {
			info.WarningMessage = "Podman installed but no compose tool found"
		}
	}

	return info
}
