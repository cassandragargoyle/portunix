/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package engine

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"portunix.ai/portunix/src/pkg/archive"
)

// DockerInstaller handles Docker installation on various platforms
type DockerInstaller struct {
	storage        *StorageAnalyzer
	dryRun         bool
	dataRoot       string // Explicit --data-root path (empty when not provided)
	nonInteractive bool   // --yes: suppress prompts and use recommended defaults
}

// NewDockerInstaller creates a new Docker installer instance.
// dataRoot overrides storage selection when non-empty; nonInteractive suppresses
// the interactive directory prompt and falls back to the recommended default.
func NewDockerInstaller(dryRun bool, dataRoot string, nonInteractive bool) *DockerInstaller {
	return &DockerInstaller{
		storage:        NewStorageAnalyzer(10), // 10 GB minimum for Docker
		dryRun:         dryRun,
		dataRoot:       dataRoot,
		nonInteractive: nonInteractive,
	}
}

// Install performs Docker installation based on current platform
func (d *DockerInstaller) Install() error {
	fmt.Println("🐳 Starting Docker installation with intelligent storage detection...")

	// Check if Docker is already installed
	if d.isDockerInstalled() {
		fmt.Println("✅ Docker is already installed")
		return d.verifyInstallation()
	}

	switch runtime.GOOS {
	case "windows":
		return d.installWindows()
	case "linux":
		return d.installLinux()
	case "darwin":
		return d.installMacOS()
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

func (d *DockerInstaller) isDockerInstalled() bool {
	cmd := exec.Command("docker", "--version")
	return cmd.Run() == nil
}

func (d *DockerInstaller) verifyInstallation() error {
	fmt.Println("\n🔍 Verifying Docker installation...")

	// Check version
	cmd := exec.Command("docker", "--version")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("docker --version failed: %w", err)
	}
	fmt.Printf("✅ %s", string(output))

	// Check daemon
	cmd = exec.Command("docker", "info")
	if err := cmd.Run(); err != nil {
		fmt.Println("⚠️  Docker daemon may not be running")
		fmt.Println("   Start Docker Desktop or run: sudo systemctl start docker")
	} else {
		fmt.Println("✅ Docker daemon is running")
	}

	return nil
}

// checkWindowsAdminRequired returns an error when a real Docker Desktop
// installation would fail without Administrator privileges. Dry-run is allowed
// from a non-elevated shell so users can preview the installation plan.
//
// Used as the safety-net message when UAC auto-elevation is unavailable or
// declined. The happy path now uses decideElevation + runWithUAC instead.
func checkWindowsAdminRequired(dryRun, isAdmin bool) error {
	if dryRun || isAdmin {
		return nil
	}
	return fmt.Errorf("❌ Docker Desktop installation requires Administrator privileges.\n" +
		"   Please run Portunix from an elevated PowerShell or cmd and try again.\n" +
		"   Tip: --dry-run works from a non-elevated shell.")
}

// elevationAction tells the Docker installer how to launch the Docker Desktop
// installer EXE based on the current process's privilege state.
type elevationAction int

const (
	// elevationDirect: run the installer in-process. Either we already have
	// admin, or we're in dry-run and skip the launch entirely.
	elevationDirect elevationAction = iota
	// elevationUAC: spawn the installer through a UAC consent prompt because
	// the current process is not elevated.
	elevationUAC
)

// decideElevation picks how to launch Docker Desktop's installer. Pure
// function so the policy is unit-testable without touching the OS.
func decideElevation(dryRun, isAdmin bool) elevationAction {
	if dryRun || isAdmin {
		return elevationDirect
	}
	return elevationUAC
}

// runDockerInstaller launches Docker Desktop's installer, escalating via UAC
// when the current process is not elevated. On UAC decline it surfaces a
// clear, actionable error pointing at the elevated-shell fallback.
func (d *DockerInstaller) runDockerInstaller(installerPath string, args []string) error {
	switch decideElevation(d.dryRun, IsAdmin()) {
	case elevationDirect:
		cmd := exec.Command(installerPath, args...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("Docker Desktop installation failed: %w", err)
		}
		return nil

	case elevationUAC:
		fmt.Println("🔐 Requesting Administrator privileges for Docker Desktop installer...")
		fmt.Println("   A UAC prompt will appear — please click Yes to continue.")
		fmt.Println("   Note: installer output won't appear in this console — watch the Docker installer's own progress dialog.")
		err := runWithUAC(installerPath, args)
		if err == nil {
			return nil
		}
		if errors.Is(err, errUACDeclined) {
			return fmt.Errorf("❌ UAC elevation was declined.\n" +
				"   Docker Desktop installer requires Administrator privileges.\n" +
				"   Please re-run `portunix install docker` and approve the UAC prompt,\n" +
				"   or run Portunix from an elevated PowerShell or cmd.")
		}
		return fmt.Errorf("Docker Desktop installation failed: %w", err)
	}
	return fmt.Errorf("unexpected elevation action")
}

// prereqChoice identifies which Docker virtualization backend the user
// picked from the interactive menu.
type prereqChoice int

const (
	prereqNone prereqChoice = iota
	prereqWSL
	prereqHyperV
	prereqCancel
)

// decidePrereqChoice picks a prerequisite choice from (hasWSL, hasHyperV,
// dryRun, nonInteractive, userInput). Pure function — all I/O happens at the
// call site. `userInput` is the trimmed text the user typed ("" = accept
// default).
func decidePrereqChoice(hasWSL, hasHyperV, dryRun, nonInteractive bool, userInput string) (prereqChoice, error) {
	if hasWSL || hasHyperV {
		return prereqNone, nil
	}
	if dryRun || nonInteractive {
		return prereqWSL, nil
	}
	switch strings.TrimSpace(userInput) {
	case "", "1":
		return prereqWSL, nil
	case "2":
		return prereqHyperV, nil
	case "3":
		return prereqCancel, nil
	default:
		return prereqNone, fmt.Errorf("invalid choice: %q", userInput)
	}
}

// hasWindowsWSL reports whether WSL is installed and usable. `wsl --status`
// returns exit 0 only when the WSL runtime is present.
func hasWindowsWSL() bool {
	return exec.Command("wsl", "--status").Run() == nil
}

// hasWindowsHyperV reports whether the Hyper-V Windows feature is enabled.
// Uses PowerShell because there is no lightweight native Go query.
func hasWindowsHyperV() bool {
	cmd := exec.Command("powershell", "-NoProfile", "-NonInteractive", "-Command",
		"(Get-WindowsOptionalFeature -Online -FeatureName Microsoft-Hyper-V-All -ErrorAction SilentlyContinue).State -eq 'Enabled'")
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) == "True"
}

// enableWindowsHyperV enables the Hyper-V Windows feature without an
// immediate restart. Must be called from an elevated process.
func enableWindowsHyperV(dryRun bool) error {
	if dryRun {
		fmt.Println("🔍 DRY RUN - Would enable Windows feature Microsoft-Hyper-V-All")
		return nil
	}
	fmt.Println("\n🔧 Enabling Hyper-V Windows feature (this may take a minute)...")
	if err := enableWindowsOptionalFeature("Microsoft-Hyper-V-All"); err != nil {
		return fmt.Errorf("failed to enable Hyper-V: %w", err)
	}
	fmt.Println("\n✅ Hyper-V enabled successfully!")
	fmt.Println("⚠️  A system restart is required before Hyper-V is usable.")
	fmt.Println("   After restart, you can run: portunix install docker")
	return nil
}

// ensureWindowsVirtualizationPrereqs interactively resolves missing
// WSL2/Hyper-V prerequisites for Docker Desktop. Returns true when Docker
// install may proceed (prereqs already satisfied). When a prereq is
// installed inline it returns false + prereqInstalledError so the caller
// can exit cleanly after pointing the user at the restart step.
func (d *DockerInstaller) ensureWindowsVirtualizationPrereqs() (bool, error) {
	hasWSL := hasWindowsWSL()
	hasHV := hasWindowsHyperV()
	if hasWSL || hasHV {
		return true, nil
	}

	fmt.Println("\n❌ Docker Desktop requires WSL2 or Hyper-V, but neither is available.")

	userInput := ""
	if !d.dryRun && !d.nonInteractive {
		fmt.Println("\n📁 Select prerequisite to install:")
		fmt.Println("   1. WSL2 (recommended — requires restart)")
		fmt.Println("   2. Hyper-V (requires restart)")
		fmt.Println("   3. Cancel (install Docker later)")
		fmt.Print("Choice [1]: ")

		reader := bufio.NewReader(os.Stdin)
		line, err := reader.ReadString('\n')
		if err != nil {
			return false, fmt.Errorf("failed to read choice: %w", err)
		}
		userInput = line
	}

	choice, err := decidePrereqChoice(hasWSL, hasHV, d.dryRun, d.nonInteractive, userInput)
	if err != nil {
		return false, err
	}

	switch choice {
	case prereqWSL:
		if err := NewWSLInstaller(d.dryRun).Install(); err != nil {
			return false, err
		}
		return false, nil
	case prereqHyperV:
		if err := enableWindowsHyperV(d.dryRun); err != nil {
			return false, err
		}
		return false, nil
	case prereqCancel:
		fmt.Println("   Docker install aborted. Run `portunix install wsl` later when ready.")
		return false, nil
	}
	return false, fmt.Errorf("unexpected prereq choice: %v", choice)
}

// installWindows installs Docker Desktop on Windows.
//
// Elevation policy: the early admin fail-fast from #177 is gone — the
// installer EXE itself is now launched through a UAC prompt (issue #178)
// when the current process is not elevated. Download and prereq detection
// run non-elevated; only the installer EXE crosses the UAC boundary.
func (d *DockerInstaller) installWindows() error {
	proceed, err := d.ensureWindowsVirtualizationPrereqs()
	if err != nil {
		return err
	}
	if !proceed {
		// A prerequisite was installed inline (or user cancelled). The user
		// message is already printed by the inner installer; we skip the
		// Docker install and exit cleanly.
		return nil
	}

	fmt.Println("\n📊 Analyzing available storage...")

	drives, err := d.storage.GetWindowsDrives()
	if err != nil {
		return fmt.Errorf("failed to analyze storage: %w", err)
	}

	// Display storage options
	fmt.Println("\n💾 Available drives:")
	for _, drive := range drives {
		status := ""
		spaceBytes := parseSpaceString(drive.FreeSpace)
		if spaceBytes < d.storage.minSpace {
			status = " ⚠️ (insufficient space)"
		}
		fmt.Printf("   %s:\\ - %s free / %s total%s\n", drive.Letter, drive.FreeSpace, drive.TotalSpace, status)
	}

	// Resolve final data-root: explicit flag, interactive prompt, or recommended default.
	dataRoot, explicit, err := d.resolveWindowsDataRoot(drives)
	if err != nil {
		return err
	}

	fmt.Printf("\n✅ Selected Docker data-root: %s\n", dataRoot)

	if d.dryRun {
		fmt.Println("\n🔍 DRY RUN - Would perform the following:")
		fmt.Printf("   1. Download Docker Desktop installer\n")
		fmt.Printf("   2. Install Docker Desktop with data-root: %s\n", dataRoot)
		fmt.Printf("   3. Configure Docker settings\n")
		fmt.Printf("   4. Verify installation\n")
		return nil
	}

	// Download Docker Desktop
	fmt.Println("\n📥 Downloading Docker Desktop for Windows...")
	dockerURL := "https://desktop.docker.com/win/main/amd64/Docker%20Desktop%20Installer.exe"

	installerPath, err := archive.DownloadFileWithProperFilename(dockerURL, os.TempDir())
	if err != nil {
		return fmt.Errorf("failed to download Docker Desktop: %w", err)
	}

	// Verify the downloaded file exists
	if _, err := os.Stat(installerPath); os.IsNotExist(err) {
		return fmt.Errorf("downloaded installer not found at: %s", installerPath)
	}

	// Run installer
	fmt.Println("🔧 Installing Docker Desktop...")

	// Docker Desktop installer arguments.
	// NOTE: --quiet is intentionally omitted so Docker's built-in progress
	// dialog stays visible during the ~1-2 minute install.
	args := []string{
		"install",
		"--accept-license",
	}

	if err := d.runDockerInstaller(installerPath, args); err != nil {
		return err
	}

	// Configure data-root when the user chose explicitly, or when the automatic
	// pick landed on a non-C drive (legacy behavior: leave Docker Desktop's
	// default alone for auto-selected C:).
	if explicit || !strings.EqualFold(filepath.VolumeName(dataRoot), "C:") {
		fmt.Printf("⚙️  Configuring Docker data-root to: %s\n", dataRoot)
		if err := d.configureWindowsDataRoot(dataRoot); err != nil {
			fmt.Printf("⚠️  Could not configure data-root: %v\n", err)
			fmt.Printf("   You can manually set data-root in Docker Desktop settings\n")
		}
	}

	fmt.Println("\n✅ Docker Desktop installed successfully!")

	// Try to start Docker Desktop automatically
	fmt.Println("🚀 Starting Docker Desktop...")
	startCmd := exec.Command("cmd", "/C", "start", "", "Docker Desktop")
	startCmd.Run() // Ignore errors

	// Wait for Docker daemon to become available (up to 120 seconds)
	fmt.Println("⏳ Waiting for Docker daemon to start...")
	for i := 0; i < 24; i++ {
		time.Sleep(5 * time.Second)

		checkCmd := exec.Command("docker", "version")
		if checkCmd.Run() == nil {
			fmt.Println("✅ Docker daemon is running and ready!")
			return nil
		}
		fmt.Printf("   Waiting... (%d/120s)\n", (i+1)*5)
	}

	// Docker didn't start within timeout
	fmt.Println("\n⚠️  Docker Desktop is starting but daemon is not ready yet.")
	fmt.Println("   You may need to wait a bit longer or restart your computer.")

	return nil
}

func (d *DockerInstaller) configureWindowsDataRoot(dataRoot string) error {
	// Create data directory
	if err := os.MkdirAll(dataRoot, 0755); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	// Docker Desktop config is in %USERPROFILE%\.docker\daemon.json
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	configDir := filepath.Join(homeDir, ".docker")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	configPath := filepath.Join(configDir, "daemon.json")

	// Write daemon.json with data-root
	config := fmt.Sprintf(`{
  "data-root": "%s"
}`, strings.ReplaceAll(dataRoot, "\\", "\\\\"))

	return os.WriteFile(configPath, []byte(config), 0644)
}

// installLinux installs Docker Engine on Linux
func (d *DockerInstaller) installLinux() error {
	fmt.Println("\n📊 Analyzing available storage...")

	partitions, err := d.storage.GetLinuxPartitions()
	if err != nil {
		return fmt.Errorf("failed to analyze storage: %w", err)
	}

	// Display storage options
	fmt.Println("\n💾 Available partitions:")
	for _, part := range partitions {
		status := ""
		spaceBytes := parseSpaceString(part.FreeSpace)
		if spaceBytes < d.storage.minSpace {
			status = " ⚠️ (insufficient space)"
		}
		fmt.Printf("   %s - %s free / %s total%s\n", part.MountPoint, part.FreeSpace, part.TotalSpace, status)
	}

	// Resolve final data-root: explicit flag, interactive prompt, or recommended default.
	dataRoot, _, err := d.resolveLinuxDataRoot(partitions)
	if err != nil {
		return err
	}

	fmt.Printf("\n✅ Selected Docker data-root: %s\n", dataRoot)

	if d.dryRun {
		fmt.Println("\n🔍 DRY RUN - Would perform the following:")
		fmt.Println("   1. Install Docker prerequisites")
		fmt.Println("   2. Add Docker repository")
		fmt.Println("   3. Install Docker Engine")
		fmt.Printf("   4. Configure data-root: %s\n", dataRoot)
		fmt.Println("   5. Start Docker service")
		fmt.Println("   6. Add current user to docker group")
		return nil
	}

	// Detect distribution
	distro := d.detectLinuxDistro()
	fmt.Printf("🐧 Detected distribution: %s\n", distro)

	switch distro {
	case "ubuntu", "debian":
		return d.installDockerUbuntuDebian(dataRoot)
	case "fedora", "centos", "rhel", "rocky":
		return d.installDockerFedoraCentOS(dataRoot)
	case "arch":
		return d.installDockerArch(dataRoot)
	default:
		return d.installDockerGeneric(dataRoot)
	}
}

func (d *DockerInstaller) detectLinuxDistro() string {
	// Try /etc/os-release first
	data, err := os.ReadFile("/etc/os-release")
	if err == nil {
		content := strings.ToLower(string(data))
		if strings.Contains(content, "ubuntu") {
			return "ubuntu"
		}
		if strings.Contains(content, "debian") {
			return "debian"
		}
		if strings.Contains(content, "fedora") {
			return "fedora"
		}
		if strings.Contains(content, "centos") {
			return "centos"
		}
		if strings.Contains(content, "rhel") || strings.Contains(content, "red hat") {
			return "rhel"
		}
		if strings.Contains(content, "rocky") {
			return "rocky"
		}
		if strings.Contains(content, "arch") {
			return "arch"
		}
	}

	return "unknown"
}

func (d *DockerInstaller) installDockerUbuntuDebian(dataRoot string) error {
	commands := [][]string{
		{"sudo", "apt-get", "update"},
		{"sudo", "apt-get", "install", "-y", "ca-certificates", "curl", "gnupg"},
		{"sudo", "install", "-m", "0755", "-d", "/etc/apt/keyrings"},
	}

	// Add Docker GPG key
	fmt.Println("🔑 Adding Docker GPG key...")
	gpgCmd := exec.Command("bash", "-c", "curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /etc/apt/keyrings/docker.gpg")
	if err := gpgCmd.Run(); err != nil {
		fmt.Printf("⚠️  GPG key setup may have issues: %v\n", err)
	}

	// Add repository
	fmt.Println("📦 Adding Docker repository...")
	repoCmd := exec.Command("bash", "-c", `echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu $(. /etc/os-release && echo "$VERSION_CODENAME") stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null`)
	if err := repoCmd.Run(); err != nil {
		fmt.Printf("⚠️  Repository setup may have issues: %v\n", err)
	}

	commands = append(commands,
		[]string{"sudo", "apt-get", "update"},
		[]string{"sudo", "apt-get", "install", "-y", "docker-ce", "docker-ce-cli", "containerd.io", "docker-buildx-plugin", "docker-compose-plugin"},
	)

	for _, cmdArgs := range commands {
		fmt.Printf("🔧 Running: %s\n", strings.Join(cmdArgs, " "))
		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("command failed: %w", err)
		}
	}

	// Configure data-root if not default location
	if dataRoot != "/var/lib/docker" {
		if err := d.configureLinuxDataRoot(dataRoot); err != nil {
			fmt.Printf("⚠️  Could not configure data-root: %v\n", err)
		}
	}

	// Add user to docker group
	d.addUserToDockerGroup()

	// Start and enable Docker
	exec.Command("sudo", "systemctl", "start", "docker").Run()
	exec.Command("sudo", "systemctl", "enable", "docker").Run()

	fmt.Println("\n✅ Docker Engine installed successfully!")
	return nil
}

func (d *DockerInstaller) installDockerFedoraCentOS(dataRoot string) error {
	commands := [][]string{
		{"sudo", "dnf", "-y", "install", "dnf-plugins-core"},
		{"sudo", "dnf", "config-manager", "--add-repo", "https://download.docker.com/linux/fedora/docker-ce.repo"},
		{"sudo", "dnf", "install", "-y", "docker-ce", "docker-ce-cli", "containerd.io", "docker-buildx-plugin", "docker-compose-plugin"},
	}

	for _, cmdArgs := range commands {
		fmt.Printf("🔧 Running: %s\n", strings.Join(cmdArgs, " "))
		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			// Try yum as fallback
			if cmdArgs[1] == "dnf" {
				cmdArgs[1] = "yum"
				cmd = exec.Command(cmdArgs[0], cmdArgs[1:]...)
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				if err := cmd.Run(); err != nil {
					return fmt.Errorf("command failed: %w", err)
				}
			} else {
				return fmt.Errorf("command failed: %w", err)
			}
		}
	}

	if dataRoot != "/var/lib/docker" {
		d.configureLinuxDataRoot(dataRoot)
	}

	d.addUserToDockerGroup()
	exec.Command("sudo", "systemctl", "start", "docker").Run()
	exec.Command("sudo", "systemctl", "enable", "docker").Run()

	fmt.Println("\n✅ Docker Engine installed successfully!")
	return nil
}

func (d *DockerInstaller) installDockerArch(dataRoot string) error {
	commands := [][]string{
		{"sudo", "pacman", "-Sy", "--noconfirm", "docker"},
	}

	for _, cmdArgs := range commands {
		fmt.Printf("🔧 Running: %s\n", strings.Join(cmdArgs, " "))
		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("command failed: %w", err)
		}
	}

	if dataRoot != "/var/lib/docker" {
		d.configureLinuxDataRoot(dataRoot)
	}

	d.addUserToDockerGroup()
	exec.Command("sudo", "systemctl", "start", "docker").Run()
	exec.Command("sudo", "systemctl", "enable", "docker").Run()

	fmt.Println("\n✅ Docker Engine installed successfully!")
	return nil
}

func (d *DockerInstaller) installDockerGeneric(dataRoot string) error {
	// Use official Docker installation script
	fmt.Println("🔧 Using Docker's official installation script...")

	cmd := exec.Command("bash", "-c", "curl -fsSL https://get.docker.com | sudo sh")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Docker installation script failed: %w", err)
	}

	if dataRoot != "/var/lib/docker" {
		d.configureLinuxDataRoot(dataRoot)
	}

	d.addUserToDockerGroup()
	exec.Command("sudo", "systemctl", "start", "docker").Run()
	exec.Command("sudo", "systemctl", "enable", "docker").Run()

	fmt.Println("\n✅ Docker Engine installed successfully!")
	return nil
}

func (d *DockerInstaller) configureLinuxDataRoot(dataRoot string) error {
	// Create data directory
	if err := exec.Command("sudo", "mkdir", "-p", dataRoot).Run(); err != nil {
		return fmt.Errorf("failed to create data directory: %w", err)
	}

	// Create daemon.json
	config := fmt.Sprintf(`{
  "data-root": "%s"
}`, dataRoot)

	tmpFile := filepath.Join(os.TempDir(), "docker-daemon.json")
	if err := os.WriteFile(tmpFile, []byte(config), 0644); err != nil {
		return err
	}

	// Move to /etc/docker/daemon.json
	exec.Command("sudo", "mkdir", "-p", "/etc/docker").Run()
	return exec.Command("sudo", "mv", tmpFile, "/etc/docker/daemon.json").Run()
}

func (d *DockerInstaller) addUserToDockerGroup() {
	user := os.Getenv("USER")
	if user == "" {
		user = os.Getenv("LOGNAME")
	}

	if user != "" && user != "root" {
		fmt.Printf("👤 Adding user '%s' to docker group...\n", user)
		exec.Command("sudo", "usermod", "-aG", "docker", user).Run()
		fmt.Println("   Log out and back in for group changes to take effect")
	}
}

// installMacOS installs Docker Desktop on macOS
func (d *DockerInstaller) installMacOS() error {
	fmt.Println("\n🍎 Installing Docker Desktop for macOS...")

	if d.dryRun {
		fmt.Println("\n🔍 DRY RUN - Would perform the following:")
		fmt.Println("   1. Check for Homebrew")
		fmt.Println("   2. Install Docker Desktop via brew cask")
		return nil
	}

	// Check for Homebrew
	if _, err := exec.LookPath("brew"); err != nil {
		return fmt.Errorf("Homebrew is required to install Docker Desktop on macOS. Install from https://brew.sh")
	}

	cmd := exec.Command("brew", "install", "--cask", "docker")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("Docker Desktop installation failed: %w", err)
	}

	fmt.Println("\n✅ Docker Desktop installed successfully!")
	fmt.Println("   Open Docker Desktop from Applications to complete setup")

	return nil
}

// resolveWindowsDataRoot determines the final Docker data-root path on Windows.
// Precedence: explicit --data-root flag > interactive prompt > recommended default.
// Returns the chosen path and explicit=true when the user (or flag) actively picked it.
func (d *DockerInstaller) resolveWindowsDataRoot(drives []DriveInfo) (string, bool, error) {
	if d.dataRoot != "" {
		if err := d.validateWindowsPath(d.dataRoot, drives); err != nil {
			return "", false, err
		}
		return d.dataRoot, true, nil
	}

	recommended, err := d.storage.analyzeWindowsStorage()
	if err != nil {
		return "", false, err
	}

	if d.nonInteractive {
		return fmt.Sprintf("%s:\\docker-data", recommended), false, nil
	}

	return d.promptWindowsDataRoot(drives, recommended)
}

// resolveLinuxDataRoot determines the final Docker data-root path on Linux.
// Precedence: explicit --data-root flag > interactive prompt > recommended default.
func (d *DockerInstaller) resolveLinuxDataRoot(partitions []PartitionInfo) (string, bool, error) {
	if d.dataRoot != "" {
		return d.dataRoot, true, nil
	}

	recommended, err := d.storage.analyzeLinuxStorage()
	if err != nil {
		return "", false, err
	}

	if d.nonInteractive {
		return filepath.Join(recommended, "docker-data"), false, nil
	}

	return d.promptLinuxDataRoot(partitions, recommended)
}

// promptWindowsDataRoot shows an interactive menu of eligible drives plus a custom
// path option. Pressing Enter accepts the recommended drive.
func (d *DockerInstaller) promptWindowsDataRoot(drives []DriveInfo, recommended string) (string, bool, error) {
	var eligible []DriveInfo
	for _, dr := range drives {
		if parseSpaceString(dr.FreeSpace) >= d.storage.minSpace {
			eligible = append(eligible, dr)
		}
	}
	if len(eligible) == 0 {
		return "", false, fmt.Errorf("no drives with sufficient space (>= %d GB)", d.storage.minSpace/(1024*1024*1024))
	}

	defaultIdx := 0
	for i, dr := range eligible {
		if dr.Letter == recommended {
			defaultIdx = i
			break
		}
	}

	fmt.Println("\n📁 Select Docker data-root location:")
	for i, dr := range eligible {
		marker := ""
		if i == defaultIdx {
			marker = " (recommended)"
		}
		fmt.Printf("   %d. %s:\\docker-data - %s free%s\n", i+1, dr.Letter, dr.FreeSpace, marker)
	}
	customIdx := len(eligible) + 1
	fmt.Printf("   %d. Custom path (enter manually)\n", customIdx)
	fmt.Printf("Choice [%d]: ", defaultIdx+1)

	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", false, fmt.Errorf("failed to read choice: %w", err)
	}
	line = strings.TrimSpace(line)

	choice := defaultIdx + 1
	if line != "" {
		n, err := strconv.Atoi(line)
		if err != nil || n < 1 || n > customIdx {
			return "", false, fmt.Errorf("invalid choice: %q", line)
		}
		choice = n
	}

	if choice == customIdx {
		fmt.Print("Enter custom path (e.g. D:\\docker-data): ")
		custom, err := reader.ReadString('\n')
		if err != nil {
			return "", false, fmt.Errorf("failed to read path: %w", err)
		}
		custom = strings.TrimSpace(custom)
		if custom == "" {
			return "", false, fmt.Errorf("custom path cannot be empty")
		}
		if err := d.validateWindowsPath(custom, drives); err != nil {
			return "", false, err
		}
		return custom, true, nil
	}

	return fmt.Sprintf("%s:\\docker-data", eligible[choice-1].Letter), true, nil
}

// promptLinuxDataRoot shows an interactive menu of eligible partitions plus a
// custom path option. Pressing Enter accepts the recommended partition.
func (d *DockerInstaller) promptLinuxDataRoot(partitions []PartitionInfo, recommended string) (string, bool, error) {
	var eligible []PartitionInfo
	for _, p := range partitions {
		if parseSpaceString(p.FreeSpace) >= d.storage.minSpace {
			eligible = append(eligible, p)
		}
	}
	if len(eligible) == 0 {
		return "", false, fmt.Errorf("no partitions with sufficient space (>= %d GB)", d.storage.minSpace/(1024*1024*1024))
	}

	defaultIdx := 0
	for i, p := range eligible {
		if p.MountPoint == recommended {
			defaultIdx = i
			break
		}
	}

	fmt.Println("\n📁 Select Docker data-root location:")
	for i, p := range eligible {
		marker := ""
		if i == defaultIdx {
			marker = " (recommended)"
		}
		fmt.Printf("   %d. %s - %s free%s\n", i+1, filepath.Join(p.MountPoint, "docker-data"), p.FreeSpace, marker)
	}
	customIdx := len(eligible) + 1
	fmt.Printf("   %d. Custom path (enter manually)\n", customIdx)
	fmt.Printf("Choice [%d]: ", defaultIdx+1)

	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", false, fmt.Errorf("failed to read choice: %w", err)
	}
	line = strings.TrimSpace(line)

	choice := defaultIdx + 1
	if line != "" {
		n, err := strconv.Atoi(line)
		if err != nil || n < 1 || n > customIdx {
			return "", false, fmt.Errorf("invalid choice: %q", line)
		}
		choice = n
	}

	if choice == customIdx {
		fmt.Print("Enter custom path (e.g. /mnt/data/docker): ")
		custom, err := reader.ReadString('\n')
		if err != nil {
			return "", false, fmt.Errorf("failed to read path: %w", err)
		}
		custom = strings.TrimSpace(custom)
		if custom == "" {
			return "", false, fmt.Errorf("custom path cannot be empty")
		}
		if !filepath.IsAbs(custom) {
			return "", false, fmt.Errorf("custom path must be absolute: %q", custom)
		}
		return custom, true, nil
	}

	return filepath.Join(eligible[choice-1].MountPoint, "docker-data"), true, nil
}

// validateWindowsPath checks that the Windows path has a drive letter and the
// drive has at least the minimum required free space. Uses the pre-fetched
// drives list to avoid re-running the PowerShell query.
func (d *DockerInstaller) validateWindowsPath(path string, drives []DriveInfo) error {
	vol := filepath.VolumeName(path)
	if len(vol) < 2 || vol[1] != ':' {
		return fmt.Errorf("invalid Windows path: %q (must include drive letter, e.g. D:\\docker-data)", path)
	}
	letter := strings.TrimSuffix(vol, ":")
	for _, dr := range drives {
		if strings.EqualFold(dr.Letter, letter) {
			if parseSpaceString(dr.FreeSpace) < d.storage.minSpace {
				return fmt.Errorf("drive %s:\\ has insufficient free space (%s, need >= %d GB)",
					letter, dr.FreeSpace, d.storage.minSpace/(1024*1024*1024))
			}
			return nil
		}
	}
	return fmt.Errorf("drive %s:\\ not found", letter)
}
