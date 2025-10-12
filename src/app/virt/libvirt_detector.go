package virt

import (
	"os/exec"
	"strings"
)

// LibvirtStatus represents libvirt daemon status
type LibvirtStatus struct {
	Installed           bool
	DaemonType          string // "monolithic" (libvirtd) or "modular" (virtqemud)
	DaemonName          string // Actual daemon name (libvirtd or virtqemud)
	SocketName          string // Socket name (libvirtd.socket or virtqemud.socket)
	Running             bool
	Enabled             bool
	Masked              bool
	SocketActivated     bool
	SocketFailed        bool     // Socket is in failed state
	SocketState         string   // Socket state details (active/failed/inactive)
	MissingDependencies []string // Dependencies that are missing or not-found
	FailedDependencies  []string // Dependencies that are in failed state
	MaskedDependencies  []string // Dependencies that are masked
	Version             string
	Issues              []string
	Recommendations     []string
}

// DetectLibvirtStatus detects libvirt daemon status
func DetectLibvirtStatus() (*LibvirtStatus, error) {
	status := &LibvirtStatus{
		Issues:              []string{},
		Recommendations:     []string{},
		MissingDependencies: []string{},
		FailedDependencies:  []string{},
		MaskedDependencies:  []string{},
	}

	// Check libvirt-daemon package (via virsh command)
	if _, err := exec.LookPath("virsh"); err == nil {
		status.Installed = true
	}

	if !status.Installed {
		status.Issues = append(status.Issues, "Libvirt is not installed")
		status.Recommendations = append(status.Recommendations, "Install libvirt: portunix install libvirt")
		return status, nil
	}

	// Get version
	if version := getLibvirtVersion(); version != "" {
		status.Version = version
	}

	// Try monolithic daemon first (older systems)
	if checkService("libvirtd.service") {
		status.DaemonType = "monolithic"
		status.DaemonName = "libvirtd"
		status.Running = isServiceActive("libvirtd")
		status.Enabled = isServiceEnabled("libvirtd")
		status.Masked = isServiceMasked("libvirtd")
	} else if checkService("virtqemud.service") {
		// Try modular daemon (newer systems)
		status.DaemonType = "modular"
		status.DaemonName = "virtqemud"
		status.Running = isServiceActive("virtqemud")
		status.Enabled = isServiceEnabled("virtqemud")
		status.Masked = isServiceMasked("virtqemud")
	}

	// Check socket activation
	if checkService("libvirtd.socket") {
		status.SocketName = "libvirtd.socket"
		status.SocketActivated = isServiceActive("libvirtd.socket")
		status.SocketState = getServiceState("libvirtd.socket")
		status.SocketFailed = isServiceFailed("libvirtd.socket")
	} else if checkService("virtqemud.socket") {
		status.SocketName = "virtqemud.socket"
		status.SocketActivated = isServiceActive("virtqemud.socket")
		status.SocketState = getServiceState("virtqemud.socket")
		status.SocketFailed = isServiceFailed("virtqemud.socket")
	}

	// Check dependencies
	status.MissingDependencies, status.FailedDependencies, status.MaskedDependencies = checkLibvirtDependencies(status.DaemonName)

	// Analyze issues (priority order)
	// Dependencies are HIGHEST priority - they cause socket/daemon failures
	if len(status.MaskedDependencies) > 0 {
		status.Issues = append(status.Issues, "Critical libvirt dependencies are masked: "+strings.Join(status.MaskedDependencies, ", "))
		status.Recommendations = append(status.Recommendations, "Unmask dependencies: sudo portunix virt check --fix-libvirt")
	}
	if len(status.MissingDependencies) > 0 {
		status.Issues = append(status.Issues, "Required libvirt dependencies missing: "+strings.Join(status.MissingDependencies, ", "))
		status.Recommendations = append(status.Recommendations, "Install missing packages or enable services: sudo portunix virt check --fix-libvirt")
	}
	if len(status.FailedDependencies) > 0 {
		status.Issues = append(status.Issues, "Libvirt dependencies in failed state: "+strings.Join(status.FailedDependencies, ", "))
		status.Recommendations = append(status.Recommendations, "Reset failed dependencies: sudo portunix virt check --fix-libvirt")
	}

	// Socket failed is CRITICAL - but check if it's due to dependencies first
	if status.SocketFailed && len(status.MaskedDependencies) == 0 && len(status.MissingDependencies) == 0 {
		status.Issues = append(status.Issues, "Libvirt socket is in failed state (trigger-limit-hit or crash loop)")
		status.Recommendations = append(status.Recommendations, "Fix socket: sudo portunix virt check --fix-libvirt")
	}

	// Check daemon issues only if no dependency or socket problems
	if len(status.MaskedDependencies) == 0 && len(status.MissingDependencies) == 0 && len(status.FailedDependencies) == 0 && !status.SocketFailed {
		if status.Masked {
			status.Issues = append(status.Issues, "Libvirt daemon is masked (cannot be started)")
			status.Recommendations = append(status.Recommendations, "Unmask and start: sudo portunix virt check --fix-libvirt")
		} else if !status.Running && !status.SocketActivated {
			status.Issues = append(status.Issues, "Libvirt daemon is not running")
			status.Recommendations = append(status.Recommendations, "Start daemon: sudo portunix virt check --fix-libvirt")
		} else if !status.Enabled && !status.SocketActivated {
			status.Issues = append(status.Issues, "Libvirt daemon not enabled on boot")
			status.Recommendations = append(status.Recommendations, "Enable daemon: sudo systemctl enable "+status.DaemonName)
		}
	}

	return status, nil
}

// Helper functions

// checkService checks if a systemd service exists
func checkService(name string) bool {
	cmd := exec.Command("systemctl", "list-unit-files", name)
	output, err := cmd.Output()
	return err == nil && strings.Contains(string(output), name)
}

// isServiceActive checks if a service is currently active
func isServiceActive(name string) bool {
	cmd := exec.Command("systemctl", "is-active", name)
	output, _ := cmd.Output()
	return strings.TrimSpace(string(output)) == "active"
}

// isServiceEnabled checks if a service is enabled to start on boot
func isServiceEnabled(name string) bool {
	cmd := exec.Command("systemctl", "is-enabled", name)
	output, _ := cmd.Output()
	enabled := strings.TrimSpace(string(output))
	return enabled == "enabled" || enabled == "static"
}

// isServiceMasked checks if a service is masked
func isServiceMasked(name string) bool {
	cmd := exec.Command("systemctl", "show", name, "-p", "UnitFileState")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.Contains(string(output), "masked")
}

// getServiceState gets the detailed state of a service
func getServiceState(name string) string {
	cmd := exec.Command("systemctl", "show", name, "-p", "ActiveState", "-p", "SubState", "-p", "Result")
	output, err := cmd.Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(output))
}

// isServiceFailed checks if a service is in failed state
func isServiceFailed(name string) bool {
	cmd := exec.Command("systemctl", "is-failed", name)
	output, _ := cmd.Output()
	state := strings.TrimSpace(string(output))
	return state == "failed"
}

// checkLibvirtDependencies checks required libvirt dependencies
// Returns: missing, failed, masked
func checkLibvirtDependencies(daemonName string) ([]string, []string, []string) {
	missing := []string{}
	failed := []string{}
	masked := []string{}

	// Dependencies based on daemon type
	var deps []string
	if daemonName == "libvirtd" {
		// Monolithic daemon dependencies
		deps = []string{
			"virtlogd.socket",
			"virtlogd.service",
			"virtlockd.socket",
		}
	} else if daemonName == "virtqemud" {
		// Modular daemon dependencies
		deps = []string{
			"virtlogd.socket",
			"virtlogd.service",
			"virtlockd.socket",
		}
	}

	for _, dep := range deps {
		// Check if unit file exists
		if !checkService(dep) {
			missing = append(missing, dep)
			continue
		}

		// Check if masked
		if isServiceMasked(dep) {
			masked = append(masked, dep)
			continue
		}

		// Check if failed
		if isServiceFailed(dep) {
			failed = append(failed, dep)
		}
	}

	return missing, failed, masked
}

// getLibvirtVersion gets libvirt version via virsh
func getLibvirtVersion() string {
	cmd := exec.Command("virsh", "--version")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

// UnmaskLibvirtDaemon unmasks and starts libvirt daemon
func UnmaskLibvirtDaemon(daemonName string) error {
	// Unmask
	cmd := exec.Command("sudo", "systemctl", "unmask", daemonName)
	if err := cmd.Run(); err != nil {
		return err
	}

	// Start (ignore error if already running)
	cmd = exec.Command("sudo", "systemctl", "start", daemonName)
	cmd.Run() // Ignore error - daemon might already be running
	return nil
}

// StartLibvirtDaemon starts libvirt daemon
func StartLibvirtDaemon(daemonName string) error {
	cmd := exec.Command("sudo", "systemctl", "start", daemonName)
	cmd.Run() // Ignore error - daemon might already be running
	return nil
}

// EnableLibvirtDaemon enables libvirt daemon on boot
func EnableLibvirtDaemon(daemonName string) error {
	cmd := exec.Command("sudo", "systemctl", "enable", daemonName)
	return cmd.Run()
}

// ResetFailedSocket resets failed socket state and restarts it
func ResetFailedSocket(socketName string) error {
	// Reset failed state for both socket and service
	cmd := exec.Command("sudo", "systemctl", "reset-failed", socketName)
	cmd.Run() // Ignore error - might not be in failed state

	// Also reset the daemon service
	daemonName := strings.Replace(socketName, ".socket", "", 1)
	cmd = exec.Command("sudo", "systemctl", "reset-failed", daemonName)
	cmd.Run() // Ignore error

	// Restart socket
	cmd = exec.Command("sudo", "systemctl", "restart", socketName)
	return cmd.Run()
}

// SwitchToDirectDaemon disables socket activation and enables direct daemon
func SwitchToDirectDaemon(daemonName, socketName string) error {
	// Stop and disable socket activation
	cmd := exec.Command("sudo", "systemctl", "stop", socketName)
	cmd.Run() // Ignore error

	cmd = exec.Command("sudo", "systemctl", "disable", socketName)
	cmd.Run() // Ignore error

	// Enable and start direct daemon
	cmd = exec.Command("sudo", "systemctl", "enable", daemonName)
	if err := cmd.Run(); err != nil {
		return err
	}

	cmd = exec.Command("sudo", "systemctl", "start", daemonName)
	cmd.Run() // Ignore error - might already be running

	return nil
}

// UnmaskDependency unmasks a masked dependency
func UnmaskDependency(depName string) error {
	cmd := exec.Command("sudo", "systemctl", "unmask", depName)
	return cmd.Run()
}

// ResetFailedDependency resets a failed dependency
func ResetFailedDependency(depName string) error {
	cmd := exec.Command("sudo", "systemctl", "reset-failed", depName)
	cmd.Run() // Ignore error if not failed

	// Try to start it
	cmd = exec.Command("sudo", "systemctl", "start", depName)
	cmd.Run() // Ignore error - might auto-start via socket

	return nil
}
