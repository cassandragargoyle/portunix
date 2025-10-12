package system

import (
	"bytes"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
)

// GetQEMUVersion returns the version of QEMU if available
func GetQEMUVersion() string {
	// Determine the QEMU binary name based on platform
	var qemuBinary string
	if runtime.GOOS == "windows" {
		qemuBinary = "qemu-system-x86_64.exe"
	} else {
		qemuBinary = "qemu-system-x86_64"
	}

	// Try to get QEMU version
	cmd := exec.Command(qemuBinary, "--version")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = nil // Ignore stderr

	if err := cmd.Run(); err != nil {
		// Try alternative name
		if runtime.GOOS != "windows" {
			cmd = exec.Command("kvm", "--version")
			out.Reset()
			cmd.Stdout = &out
			cmd.Stderr = nil
			if err := cmd.Run(); err != nil {
				return ""
			}
		} else {
			return ""
		}
	}

	// Parse version from output
	// Example: "QEMU emulator version 8.0.0 (Debian 1:8.0+dfsg-1)"
	output := strings.TrimSpace(out.String())
	re := regexp.MustCompile(`version\s+(\d+\.\d+\.\d+)`)
	matches := re.FindStringSubmatch(output)
	if len(matches) > 1 {
		return "v" + matches[1]
	}

	return ""
}

// GetVirtualBoxVersion returns the version of VirtualBox if available
func GetVirtualBoxVersion() string {
	// Determine the VBoxManage binary name based on platform
	var vboxBinary string
	if runtime.GOOS == "windows" {
		vboxBinary = "VBoxManage.exe"
	} else {
		vboxBinary = "VBoxManage"
	}

	// Try to get VirtualBox version
	cmd := exec.Command(vboxBinary, "--version")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = nil // Ignore stderr

	if err := cmd.Run(); err != nil {
		return ""
	}

	// Parse version from output
	// Example: "7.0.12r159484" or "7.0.12_Ubuntur159484"
	output := strings.TrimSpace(out.String())
	// Extract version number before 'r' or '_'
	re := regexp.MustCompile(`^(\d+\.\d+\.\d+)`)
	matches := re.FindStringSubmatch(output)
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

	// Try to get libvirt version
	cmd := exec.Command("virsh", "--version")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = nil // Ignore stderr

	if err := cmd.Run(); err != nil {
		return ""
	}

	// Parse version from output
	// Example: "9.0.0"
	version := strings.TrimSpace(out.String())
	if version != "" && !strings.HasPrefix(version, "v") {
		version = "v" + version
	}
	return version
}

// GetDockerVersion returns the version of Docker if available
func GetDockerVersion() string {
	// Try to get Docker version
	cmd := exec.Command("docker", "version", "--format", "{{.Server.Version}}")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = nil // Ignore stderr

	if err := cmd.Run(); err != nil {
		// Try alternative format
		cmd = exec.Command("docker", "--version")
		var out2 bytes.Buffer
		cmd.Stdout = &out2
		cmd.Stderr = nil

		if err := cmd.Run(); err != nil {
			return ""
		}

		// Parse from "Docker version 24.0.7, build afdd53b"
		output := strings.TrimSpace(out2.String())
		parts := strings.Fields(output)
		if len(parts) >= 3 && strings.HasPrefix(parts[2], "v") {
			return strings.TrimSuffix(parts[2], ",")
		}
		for i, part := range parts {
			if i > 0 && (strings.Contains(part, ".") || strings.HasPrefix(part, "v")) {
				return strings.TrimSuffix(strings.TrimPrefix(part, "v"), ",")
			}
		}
		return ""
	}

	version := strings.TrimSpace(out.String())
	if !strings.HasPrefix(version, "v") {
		version = "v" + version
	}
	return version
}

// GetPodmanVersion returns the version of Podman if available
func GetPodmanVersion() string {
	// Try to get Podman version with format
	cmd := exec.Command("podman", "version", "--format", "{{.Version}}")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = nil // Ignore stderr

	if err := cmd.Run(); err != nil {
		// Try alternative format
		cmd = exec.Command("podman", "--version")
		var out2 bytes.Buffer
		cmd.Stdout = &out2
		cmd.Stderr = nil

		if err := cmd.Run(); err != nil {
			return ""
		}

		// Parse from "podman version 4.6.1"
		output := strings.TrimSpace(out2.String())
		parts := strings.Fields(output)
		if len(parts) >= 3 {
			version := parts[2]
			if !strings.HasPrefix(version, "v") {
				version = "v" + version
			}
			return version
		}
		return ""
	}

	version := strings.TrimSpace(out.String())
	if !strings.HasPrefix(version, "v") {
		version = "v" + version
	}
	return version
}