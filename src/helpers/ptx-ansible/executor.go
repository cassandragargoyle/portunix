package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// ExecutionOptions contains options for playbook execution
type ExecutionOptions struct {
	DryRun        bool
	Environment   string   // "local", "container", "virt"
	Target        string   // For multi-environment execution (VM name, container name)
	Image         string   // Container image for container environment
	Runtime       string   // Container runtime: "docker", "podman", or "" for auto-detect
	ContainerName string   // Custom container name (optional)
	Ports         []string // Port mappings for container (e.g., "1313:1313")
	Volumes       []string // Volume mappings for container (e.g., "./workspace:/workspace")
	NamedVolumes  []string // Named volumes for container (e.g., "node_modules:/app/node_modules")
	Verbose       bool
	User          string   // Phase 4: User executing the playbook
	ScriptFilter  []string // Phase 1 #128: Filter scripts to run (empty = all)
	ListScripts   bool     // Phase 1 #128: Just list available scripts
}

// ExecutionResult contains the result of playbook execution
type ExecutionResult struct {
	Success bool
	Message string
	Errors  []string
}

// ExecutePlaybook executes a .ptxbook file with the given options
func ExecutePlaybook(filePath string, options ExecutionOptions) (*ExecutionResult, error) {
	// Phase 4: Initialize enterprise systems
	auditConfig := GetDefaultAuditConfig()
	auditMgr, err := NewAuditManager(auditConfig)
	if err != nil {
		return &ExecutionResult{
			Success: false,
			Message: "Failed to initialize audit system",
			Errors:  []string{err.Error()},
		}, err
	}

	rbacConfig := GetDefaultRBACConfig()
	rbacMgr, err := NewRBACManager(rbacConfig, auditMgr)
	if err != nil {
		return &ExecutionResult{
			Success: false,
			Message: "Failed to initialize RBAC system",
			Errors:  []string{err.Error()},
		}, err
	}

	secretMgr := NewSecretManager(auditMgr)

	// Start audit logging for this execution
	startTime := time.Now()
	auditMgr.LogPlaybookExecution(options.User, options.Environment, filePath, true, 0, nil)

	// Parse the playbook file first
	ptxbook, err := ParsePtxbookFile(filePath)
	if err != nil {
		auditMgr.LogPlaybookExecution(options.User, options.Environment, filePath, false, time.Since(startTime), err)
		return &ExecutionResult{
			Success: false,
			Message: "Failed to parse playbook",
			Errors:  []string{err.Error()},
		}, err
	}

	// Phase 4: Check RBAC permissions for playbook execution
	accessResult := rbacMgr.CheckAccess(&AccessRequest{
		User:        options.User,
		Permission:  PermissionPlaybookExecute,
		Resource:    filePath,
		Environment: options.Environment,
	})

	if !accessResult.Granted {
		auditMgr.LogPlaybookExecution(options.User, options.Environment, filePath, false, time.Since(startTime),
			fmt.Errorf("access denied: %s", accessResult.Reason))
		return &ExecutionResult{
			Success: false,
			Message: "Access denied",
			Errors:  []string{accessResult.Reason},
		}, fmt.Errorf("access denied: %s", accessResult.Reason)
	}

	result := &ExecutionResult{
		Success: true,
		Message: "Playbook execution completed",
		Errors:  []string{},
	}

	// Phase 3: Initialize rollback manager
	rollbackManager := NewRollbackManager(ptxbook)
	if rollbackManager.IsEnabled() && options.Verbose {
		fmt.Printf("🛡️  Rollback protection enabled\n")
		if rollbackManager.GetLogFile() != "" {
			fmt.Printf("   Log file: %s\n", rollbackManager.GetLogFile())
		}
	}

	// Phase 4: Process secret references in playbook
	if err := secretMgr.ProcessSecretReferences(ptxbook); err != nil {
		auditMgr.LogPlaybookExecution(options.User, options.Environment, filePath, false, time.Since(startTime), err)
		return &ExecutionResult{
			Success: false,
			Message: "Failed to process secret references",
			Errors:  []string{err.Error()},
		}, err
	}

	if options.Verbose {
		fmt.Printf("🏢 Enterprise Features Active\n")
		fmt.Printf("   🔐 Secrets Management: AES-256-GCM encryption\n")
		fmt.Printf("   📊 Audit Logging: Security audit enabled\n")
		fmt.Printf("   🔐 Role-Based Access Control: User '%s' validated\n", options.User)
		fmt.Printf("   👥 Multi-User Environment: Enterprise mode\n")
		fmt.Printf("\n")

		fmt.Printf("🚀 Executing playbook: %s\n", ptxbook.Metadata.Name)
		if ptxbook.Metadata.Description != "" {
			fmt.Printf("   Description: %s\n", ptxbook.Metadata.Description)
		}
		fmt.Printf("   Environment: %s\n", options.Environment)
		fmt.Printf("   User: %s\n", options.User)
		if options.Environment == "container" && options.Image != "" {
			fmt.Printf("   Container Image: %s\n", options.Image)
		}
		if options.Environment == "virt" && options.Target != "" {
			fmt.Printf("   VM Target: %s\n", options.Target)
		}
	}

	// Phase 2: Setup environment if not local
	var envCtx *EnvironmentContext
	if options.Environment != "local" {
		var setupErr error
		envCtx, setupErr = setupEnvironment(options)
		if setupErr != nil {
			result.Success = false
			result.Errors = append(result.Errors, fmt.Sprintf("Environment setup failed: %v", setupErr))
			return result, setupErr
		}
		defer cleanupEnvironment(envCtx, options)
	}

	// Phase 1: Execute Portunix packages
	if ptxbook.Spec.Portunix != nil && len(ptxbook.Spec.Portunix.Packages) > 0 {
		if options.Verbose {
			fmt.Printf("📦 Installing %d Portunix packages...\n", len(ptxbook.Spec.Portunix.Packages))
		}

		if err := executePortunixPackagesWithRollback(ptxbook, options, envCtx, rollbackManager); err != nil {
			result.Success = false
			result.Errors = append(result.Errors, fmt.Sprintf("Portunix package installation failed: %v", err))

			// Phase 3: Execute rollback on failure
			if rollbackManager.IsEnabled() {
				if rollbackErr := rollbackManager.ExecuteRollback(err.Error()); rollbackErr != nil {
					result.Errors = append(result.Errors, fmt.Sprintf("Rollback failed: %v", rollbackErr))
				}
			}

			return result, err
		}

		if options.Verbose {
			fmt.Println("✅ Portunix packages installed successfully")
		}
	}

	// Phase 2: Execute Ansible playbooks if present
	if ptxbook.Spec.Ansible != nil && len(ptxbook.Spec.Ansible.Playbooks) > 0 {
		if options.Verbose {
			fmt.Printf("🔧 Executing %d Ansible playbooks...\n", len(ptxbook.Spec.Ansible.Playbooks))
		}

		// Check if Ansible is available
		if !isAnsibleAvailable() {
			errMsg := "Ansible is required but not available. Install with: portunix install ansible"
			result.Success = false
			result.Errors = append(result.Errors, errMsg)

			// Phase 3: Execute rollback on failure
			if rollbackManager.IsEnabled() {
				if rollbackErr := rollbackManager.ExecuteRollback(errMsg); rollbackErr != nil {
					result.Errors = append(result.Errors, fmt.Sprintf("Rollback failed: %v", rollbackErr))
				}
			}

			return result, fmt.Errorf(errMsg)
		}

		if err := executeAnsiblePlaybooksWithRollback(ptxbook, options, envCtx, rollbackManager); err != nil {
			result.Success = false
			result.Errors = append(result.Errors, fmt.Sprintf("Ansible playbook execution failed: %v", err))

			// Phase 3: Execute rollback on failure
			if rollbackManager.IsEnabled() {
				if rollbackErr := rollbackManager.ExecuteRollback(err.Error()); rollbackErr != nil {
					result.Errors = append(result.Errors, fmt.Sprintf("Rollback failed: %v", rollbackErr))
				}
			}

			return result, err
		}

		if options.Verbose {
			fmt.Println("✅ Ansible playbooks executed successfully")
		}
	}

	// Execute custom scripts if present
	if len(ptxbook.Spec.Scripts) > 0 {
		if options.Verbose {
			fmt.Printf("📜 Executing %d custom scripts...\n", len(ptxbook.Spec.Scripts))
		}

		if err := executeScripts(ptxbook, options, envCtx); err != nil {
			result.Success = false
			result.Errors = append(result.Errors, fmt.Sprintf("Script execution failed: %v", err))

			// Execute rollback on failure
			if rollbackManager.IsEnabled() {
				if rollbackErr := rollbackManager.ExecuteRollback(err.Error()); rollbackErr != nil {
					result.Errors = append(result.Errors, fmt.Sprintf("Rollback failed: %v", rollbackErr))
				}
			}

			return result, err
		}

		if options.Verbose {
			fmt.Println("✅ Custom scripts executed successfully")
		}
	}

	if options.Verbose {
		fmt.Println("🎉 Playbook execution completed successfully")
		fmt.Printf("📊 Audit trail logged for compliance\n")
	}

	// Phase 4: Final audit logging
	auditMgr.LogPlaybookExecution(options.User, options.Environment, filePath, result.Success, time.Since(startTime), nil)

	return result, nil
}

// executePortunixPackages installs the Portunix packages specified in the playbook
func executePortunixPackages(ptxbook *PtxbookFile, options ExecutionOptions, envCtx *EnvironmentContext) error {
	// Get the path to the main portunix binary
	portunixPath, err := getPortunixBinaryPath()
	if err != nil {
		return fmt.Errorf("failed to find portunix binary: %v", err)
	}

	for _, pkg := range ptxbook.Spec.Portunix.Packages {
		if options.Verbose {
			if pkg.Variant != "" {
				fmt.Printf("   Installing %s (variant: %s)...\n", pkg.Name, pkg.Variant)
			} else {
				fmt.Printf("   Installing %s...\n", pkg.Name)
			}
		}

		if options.DryRun {
			fmt.Printf("   [DRY-RUN] Would install: %s\n", pkg.Name)
			continue
		}

		// Build install command based on environment
		var cmd *exec.Cmd
		if envCtx != nil {
			// For container/VM environments, execute install inside the environment
			switch envCtx.Type {
			case "container":
				// Execute inside container
				args := []string{"container", "exec", envCtx.Target, portunixPath, "install", pkg.Name}
				if pkg.Variant != "" {
					args = append(args, "--variant", pkg.Variant)
				}
				cmd = exec.Command(portunixPath, args...)
			case "virt":
				// Execute on VM via SSH (simplified approach)
				// In a full implementation, this would copy the binary and execute remotely
				return fmt.Errorf("portunix package installation on VMs not yet implemented in Phase 2")
			}
		} else {
			// Local execution
			args := []string{"install", pkg.Name}
			if pkg.Variant != "" {
				args = append(args, "--variant", pkg.Variant)
			}
			cmd = exec.Command(portunixPath, args...)
		}

		if options.Verbose {
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
		}

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to install package %s: %v", pkg.Name, err)
		}
	}

	return nil
}

// executeAnsiblePlaybooks executes the Ansible playbooks specified in the playbook
func executeAnsiblePlaybooks(ptxbook *PtxbookFile, options ExecutionOptions, envCtx *EnvironmentContext) error {
	playbookDir := filepath.Dir(ptxbook.Metadata.Name) // Assume playbooks are relative to .ptxbook file

	for _, playbook := range ptxbook.Spec.Ansible.Playbooks {
		if options.Verbose {
			fmt.Printf("   Executing Ansible playbook: %s\n", playbook.Path)
		}

		if options.DryRun {
			fmt.Printf("   [DRY-RUN] Would execute: %s\n", playbook.Path)
			continue
		}

		// Resolve playbook path (relative to .ptxbook file)
		playbookPath := playbook.Path
		if !filepath.IsAbs(playbookPath) {
			playbookPath = filepath.Join(playbookDir, playbookPath)
		}

		// Check if playbook file exists
		if _, err := os.Stat(playbookPath); os.IsNotExist(err) {
			return fmt.Errorf("ansible playbook not found: %s", playbookPath)
		}

		// Build ansible-playbook command
		args := []string{"ansible-playbook", playbookPath}

		// Configure inventory and connection based on environment
		if envCtx != nil {
			// Create temporary inventory file
			inventoryPath, err := createTemporaryInventory(envCtx.Inventory)
			if err != nil {
				return fmt.Errorf("failed to create inventory file: %v", err)
			}
			defer os.Remove(inventoryPath)

			args = append(args, "-i", inventoryPath)

			if options.Verbose {
				fmt.Printf("   Using inventory: %s\n", inventoryPath)
				fmt.Printf("   Target: %s (%s)\n", envCtx.Target, envCtx.Type)
			}
		} else {
			// Default to localhost execution
			args = append(args, "-i", "localhost,")
			args = append(args, "--connection", "local")
		}

		// Execute the ansible-playbook command
		cmd := exec.Command("ansible-playbook", args[1:]...)
		if options.Verbose {
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
		}

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to execute ansible playbook %s: %v", playbook.Path, err)
		}
	}

	return nil
}

// getPortunixBinaryPath finds the path to the main portunix binary for the current OS
func getPortunixBinaryPath() (string, error) {
	// Get the current executable path (ptx-ansible)
	currentExe, err := os.Executable()
	if err != nil {
		return "", err
	}

	// The portunix binary should be in the same directory
	execDir := filepath.Dir(currentExe)
	portunixPath := filepath.Join(execDir, "portunix")

	// Add .exe suffix on Windows
	if strings.Contains(strings.ToLower(os.Getenv("OS")), "windows") {
		portunixPath += ".exe"
	}

	// Check if the binary exists
	if _, err := os.Stat(portunixPath); os.IsNotExist(err) {
		return "", fmt.Errorf("portunix binary not found at %s", portunixPath)
	}

	return portunixPath, nil
}

// getLinuxBinaryPath finds the path to the Linux portunix binary (for container use)
// Uses cross-platform binary distribution (ADR-031, Issue #125)
func getLinuxBinaryPath() (string, error) {
	// Get the current executable path (ptx-ansible)
	currentExe, err := os.Executable()
	if err != nil {
		return "", err
	}
	execDir := filepath.Dir(currentExe)

	// On Linux, just return the current binary
	if runtime.GOOS == "linux" {
		linuxPath := filepath.Join(execDir, "portunix")
		if _, err := os.Stat(linuxPath); err == nil {
			return linuxPath, nil
		}
	}

	// For cross-platform (e.g., Windows host → Linux container), use platform binaries
	// First check cache directory
	platformDir, err := getPlatformBinariesDir("linux-amd64")
	if err == nil {
		linuxPath := filepath.Join(platformDir, "portunix")
		if _, err := os.Stat(linuxPath); err == nil {
			return linuxPath, nil
		}
	}

	// Fallback: try to extract from platform archive
	platformDir, err = extractPlatformArchive("linux-amd64", false)
	if err == nil {
		linuxPath := filepath.Join(platformDir, "portunix")
		if _, err := os.Stat(linuxPath); err == nil {
			return linuxPath, nil
		}
	}

	return "", fmt.Errorf("Linux portunix binary not found. Cross-platform binaries may not be installed (ADR-031)")
}

// getPlatformBinariesDir returns the directory containing extracted platform binaries
// Part of ADR-031: Cross-Platform Binary Distribution Strategy
func getPlatformBinariesDir(platform string) (string, error) {
	// Get portunix installation directory
	execPath, err := os.Executable()
	if err != nil {
		return "", err
	}
	execDir := filepath.Dir(execPath)

	// Check cache directory: <install_dir>/cache/<platform>/
	cacheDir := filepath.Join(execDir, "cache", platform)
	if info, err := os.Stat(cacheDir); err == nil && info.IsDir() {
		// Verify portunix binary exists in cache
		binaryName := "portunix"
		if strings.HasPrefix(platform, "windows") {
			binaryName = "portunix.exe"
		}
		if _, err := os.Stat(filepath.Join(cacheDir, binaryName)); err == nil {
			return cacheDir, nil
		}
	}

	return "", fmt.Errorf("platform binaries not cached for %s", platform)
}

// extractPlatformArchive extracts platform binaries from the platform archive
// Part of ADR-031: Cross-Platform Binary Distribution Strategy
func extractPlatformArchive(platform string, verbose bool) (string, error) {
	execPath, err := os.Executable()
	if err != nil {
		return "", err
	}
	execDir := filepath.Dir(execPath)

	// Look for platform archive in <install_dir>/platforms/
	platformsDir := filepath.Join(execDir, "platforms")

	// Determine archive name based on platform
	var archiveName string
	if strings.HasPrefix(platform, "windows") {
		archiveName = platform + ".zip"
	} else {
		archiveName = platform + ".tar.gz"
	}

	archivePath := filepath.Join(platformsDir, archiveName)
	if _, err := os.Stat(archivePath); os.IsNotExist(err) {
		return "", fmt.Errorf("platform archive not found: %s", archivePath)
	}

	// Create cache directory
	cacheDir := filepath.Join(execDir, "cache", platform)
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create cache directory: %v", err)
	}

	if verbose {
		fmt.Printf("   Extracting platform binaries from %s...\n", archiveName)
	}

	// Extract archive based on type
	if strings.HasSuffix(archivePath, ".zip") {
		if err := extractZip(archivePath, cacheDir); err != nil {
			return "", fmt.Errorf("failed to extract zip: %v", err)
		}
	} else {
		if err := extractTarGz(archivePath, cacheDir); err != nil {
			return "", fmt.Errorf("failed to extract tar.gz: %v", err)
		}
	}

	if verbose {
		fmt.Printf("   ✓ Platform binaries extracted to cache\n")
	}

	return cacheDir, nil
}

// extractTarGz extracts a tar.gz archive to the specified directory
func extractTarGz(archivePath, destDir string) error {
	cmd := exec.Command("tar", "-xzf", archivePath, "-C", destDir)
	return cmd.Run()
}

// extractZip extracts a zip archive to the specified directory
func extractZip(archivePath, destDir string) error {
	if runtime.GOOS == "windows" {
		// Use PowerShell on Windows
		cmd := exec.Command("powershell", "-Command",
			fmt.Sprintf("Expand-Archive -Path '%s' -DestinationPath '%s' -Force", archivePath, destDir))
		return cmd.Run()
	}
	// Use unzip on Linux/macOS
	cmd := exec.Command("unzip", "-o", archivePath, "-d", destDir)
	return cmd.Run()
}

// detectContainerPlatform detects the target platform from container image
// Part of ADR-031: Cross-Platform Binary Distribution Strategy
func detectContainerPlatform(image string) string {
	// Default to linux-amd64 for most container images
	// Container images are almost always Linux-based
	image = strings.ToLower(image)

	// Check for ARM-based images
	if strings.Contains(image, "arm64") || strings.Contains(image, "aarch64") {
		return "linux-arm64"
	}

	// Windows containers (rare but possible)
	if strings.Contains(image, "windows") || strings.Contains(image, "nanoserver") || strings.Contains(image, "servercore") {
		return "windows-amd64"
	}

	// Default to linux-amd64 (most common case)
	return "linux-amd64"
}

// getPlatformBinaries returns paths to all platform binaries for the given platform
// Part of ADR-031: Cross-Platform Binary Distribution Strategy
func getPlatformBinaries(platform string, verbose bool) (map[string]string, error) {
	// First try to get from cache
	platformDir, err := getPlatformBinariesDir(platform)
	if err != nil {
		// Not in cache, try to extract
		platformDir, err = extractPlatformArchive(platform, verbose)
		if err != nil {
			return nil, fmt.Errorf("platform binaries not available for %s: %v", platform, err)
		}
	}

	// Build map of binary name -> path
	binaries := make(map[string]string)
	ext := ""
	if strings.HasPrefix(platform, "windows") {
		ext = ".exe"
	}

	// List of expected binaries
	binaryNames := []string{
		"portunix",
		"ptx-installer",
		"ptx-ansible",
		"ptx-container",
		"ptx-virt",
		"ptx-mcp",
		"ptx-pft",
		"ptx-python",
		"ptx-prompting",
		"ptx-aiops",
		"ptx-make",
	}

	for _, name := range binaryNames {
		binaryPath := filepath.Join(platformDir, name+ext)
		if _, err := os.Stat(binaryPath); err == nil {
			binaries[name] = binaryPath
		}
	}

	if len(binaries) == 0 {
		return nil, fmt.Errorf("no binaries found for platform %s in %s", platform, platformDir)
	}

	return binaries, nil
}

// isAnsibleAvailable checks if Ansible is installed and available
func isAnsibleAvailable() bool {
	cmd := exec.Command("ansible", "--version")
	return cmd.Run() == nil
}

// substituteVariables substitutes variables in the given text using the playbook variables
func substituteVariables(text string, variables map[string]interface{}) string {
	result := text
	for key, value := range variables {
		placeholder := fmt.Sprintf("{{ %s }}", key)
		replacement := fmt.Sprintf("%v", value)
		result = strings.ReplaceAll(result, placeholder, replacement)
	}
	return result
}

// EnvironmentContext holds context information for non-local environments
type EnvironmentContext struct {
	Type        string // "container" or "virt"
	Target      string // Container ID or VM name
	Runtime     string // Container runtime: "docker" or "podman"
	SSHHost     string // SSH connection host
	SSHPort     string // SSH connection port
	SSHUser     string // SSH username
	SSHKeyPath  string // Path to SSH private key
	TempDir     string // Temporary directory for environment setup (also stores runtime for containers)
	Inventory   string // Generated Ansible inventory content
	WorkDir     string // Working directory for script execution
}

// setupEnvironment prepares the environment for playbook execution
func setupEnvironment(options ExecutionOptions) (*EnvironmentContext, error) {
	switch options.Environment {
	case "container":
		return setupContainerEnvironment(options)
	case "virt":
		return setupVirtEnvironment(options)
	default:
		return nil, fmt.Errorf("unsupported environment: %s", options.Environment)
	}
}

// copyBinariesToContainer copies Portunix binaries to a running container
// This is called by the internal _bin-update script
func copyBinariesToContainer(containerName string, options ExecutionOptions, portunixPath, runtime string, useDirectRuntime bool) error {
	if options.Verbose {
		fmt.Printf("   📦 Copying platform binaries to container...\n")
	}

	// ADR-031: Detect target platform and get appropriate binaries
	targetPlatform := detectContainerPlatform(options.Image)
	if options.Verbose {
		fmt.Printf("   Detected target platform: %s\n", targetPlatform)
	}

	// Get Linux binary path for fallback
	linuxBinaryPath, _ := getLinuxBinaryPath()

	// Get platform binaries (uses cache or extracts from archive)
	platformBinaries, err := getPlatformBinaries(targetPlatform, options.Verbose)
	if err != nil {
		// Fallback to legacy behavior if platform binaries not available
		if options.Verbose {
			fmt.Printf("   ⚠️  Platform binaries not available: %v\n", err)
			fmt.Printf("   Falling back to local binaries (may not work cross-platform)\n")
		}

		// Get directory containing local binaries
		execDir := filepath.Dir(linuxBinaryPath)

		// List of binaries to copy (legacy behavior)
		legacyBinaries := []string{"portunix", "ptx-installer", "ptx-ansible", "ptx-container", "ptx-virt", "ptx-mcp", "ptx-pft"}
		platformBinaries = make(map[string]string)
		for _, binary := range legacyBinaries {
			srcPath := filepath.Join(execDir, binary)
			if _, err := os.Stat(srcPath); err == nil {
				platformBinaries[binary] = srcPath
			}
		}
	}

	// Copy platform binaries into the container
	for binaryName, srcPath := range platformBinaries {
		destPath := containerName + ":/usr/local/bin/" + binaryName

		var copyCmd *exec.Cmd
		if useDirectRuntime {
			copyCmd = exec.Command(runtime, "cp", srcPath, destPath)
		} else {
			copyCmd = exec.Command(portunixPath, "container", "cp", srcPath, destPath)
		}

		if err := copyCmd.Run(); err != nil {
			return fmt.Errorf("failed to copy %s to container: %v", binaryName, err)
		}

		// Make executable
		var chmodCmd *exec.Cmd
		if useDirectRuntime {
			chmodCmd = exec.Command(runtime, "exec", containerName, "chmod", "+x", "/usr/local/bin/"+binaryName)
		} else {
			chmodCmd = exec.Command(portunixPath, "container", "exec", containerName, "chmod", "+x", "/usr/local/bin/"+binaryName)
		}
		chmodCmd.Run()

		if options.Verbose {
			fmt.Printf("   ✓ Copied %s\n", binaryName)
		}
	}

	if options.Verbose {
		fmt.Printf("   All binaries installed in container (%d total)\n", len(platformBinaries))
	}

	// Install ca-certificates for HTTPS downloads using portunix
	if options.Verbose {
		fmt.Printf("   Installing ca-certificates...\n")
	}

	containerPortunixPath := "/usr/local/bin/portunix"
	var installCACmd *exec.Cmd
	if useDirectRuntime {
		installCACmd = exec.Command(runtime, "exec", containerName, containerPortunixPath, "install", "ca-certificates")
	} else {
		installCACmd = exec.Command(portunixPath, "container", "exec", containerName, containerPortunixPath, "install", "ca-certificates")
	}
	if options.Verbose {
		installCACmd.Stdout = os.Stdout
		installCACmd.Stderr = os.Stderr
	}
	installCACmd.Run() // Ignore errors, some distros may not need this

	return nil
}

// setupContainerEnvironment creates and configures a container for playbook execution
func setupContainerEnvironment(options ExecutionOptions) (*EnvironmentContext, error) {
	if options.Verbose {
		fmt.Printf("🐳 Setting up container environment with image: %s\n", options.Image)
	}

	// Get portunix binary path (for running local commands)
	portunixPath, err := getPortunixBinaryPath()
	if err != nil {
		return nil, fmt.Errorf("failed to find portunix binary: %v", err)
	}

	// Determine runtime to use
	runtime := options.Runtime
	useDirectRuntime := runtime == "docker" || runtime == "podman"

	if !useDirectRuntime {
		// Check if container runtime is available via portunix, auto-install if not
		runtime, err = ensureContainerRuntime(options)
		if err != nil {
			return nil, fmt.Errorf("failed to ensure container runtime: %v", err)
		}
	}

	if options.Verbose {
		if useDirectRuntime {
			fmt.Printf("   Using explicit runtime from playbook: %s\n", runtime)
		} else {
			fmt.Printf("   Using container runtime: %s\n", runtime)
		}
	}

	// Ensure Docker daemon is running (for Docker runtime)
	if runtime == "docker" {
		if err := ensureDockerDaemonRunning(options.Verbose, portunixPath); err != nil {
			return nil, err
		}
	}

	// Use custom container name or generate one
	containerName := options.ContainerName
	if containerName == "" {
		containerName = fmt.Sprintf("ptx-ansible-%s", generateRandomString(8))
	}

	// Parse volumes to separate bind mounts and named volumes
	bindMounts, namedVolumes := parseVolumes(options.Volumes)

	if options.Verbose {
		fmt.Printf("   Container name: %s\n", containerName)
		fmt.Printf("   Image: %s\n", options.Image)
		if len(options.Ports) > 0 {
			fmt.Printf("   Port mappings: %v\n", options.Ports)
		}
		if len(bindMounts) > 0 {
			fmt.Printf("   Bind mounts: %v\n", bindMounts)
		}
		if len(namedVolumes) > 0 {
			fmt.Printf("   Named volumes: %v\n", namedVolumes)
		}
	}

	// Create named volumes before starting container
	if len(namedVolumes) > 0 {
		if err := createNamedVolumes(namedVolumes, runtime, options.Verbose); err != nil {
			return nil, fmt.Errorf("failed to create named volumes: %v", err)
		}
	}

	// Create and start the container
	if options.Verbose {
		fmt.Printf("   Creating container...\n")
	}

	var createCmd *exec.Cmd
	if useDirectRuntime {
		// Use explicit runtime directly - build args with port and volume mappings
		args := []string{"run", "-d"}
		for _, port := range options.Ports {
			args = append(args, "-p", port)
		}
		// Add bind mounts
		for _, vol := range bindMounts {
			args = append(args, "-v", vol)
		}
		// Add named volumes (without :named suffix)
		for _, vol := range namedVolumes {
			args = append(args, "-v", vol)
		}
		args = append(args, "--name", containerName, options.Image, "sleep", "infinity")
		createCmd = exec.Command(runtime, args...)
	} else {
		// Use portunix container (auto-selects runtime) - build args with port and volume mappings
		args := []string{"container", "run", "-d"}
		for _, port := range options.Ports {
			args = append(args, "-p", port)
		}
		// Add bind mounts
		for _, vol := range bindMounts {
			args = append(args, "-v", vol)
		}
		// Add named volumes (without :named suffix)
		for _, vol := range namedVolumes {
			args = append(args, "-v", vol)
		}
		args = append(args, "--name", containerName, options.Image, "sleep", "infinity")
		createCmd = exec.Command(portunixPath, args...)
	}

	if options.Verbose {
		createCmd.Stdout = os.Stdout
		createCmd.Stderr = os.Stderr
	}
	if err := createCmd.Run(); err != nil {
		return nil, fmt.Errorf("failed to create container: %v", err)
	}

	if options.Verbose {
		fmt.Printf("   Container created successfully\n")
	}

	// Note: Binary copying is now handled by _bin-update internal script
	// This allows playbooks to control when binaries are copied

	// Create workspace directory for script execution
	workDir := "/workspace"
	if options.Verbose {
		fmt.Printf("   Creating workspace directory: %s\n", workDir)
	}
	var mkdirCmd *exec.Cmd
	if useDirectRuntime {
		mkdirCmd = exec.Command(runtime, "exec", containerName, "mkdir", "-p", workDir)
	} else {
		mkdirCmd = exec.Command(portunixPath, "container", "exec", containerName, "mkdir", "-p", workDir)
	}
	mkdirCmd.Run() // Ignore errors if directory already exists

	if options.Verbose {
		fmt.Printf("   Container initialized\n")
	}

	// Setup SSH connectivity (for Ansible if needed)
	sshKeyPath, err := setupSSHForContainer(containerName, options)
	if err != nil {
		// Non-fatal, SSH is optional for direct container execution
		if options.Verbose {
			fmt.Printf("   Note: SSH setup skipped (not required for direct execution)\n")
		}
		sshKeyPath = ""
	}

	// Generate inventory
	inventory := generateContainerInventory(containerName, sshKeyPath)

	envCtx := &EnvironmentContext{
		Type:       "container",
		Target:     containerName,
		Runtime:    runtime, // Container runtime (docker/podman)
		SSHHost:    "localhost",
		SSHPort:    "2222",
		SSHUser:    "root",
		SSHKeyPath: sshKeyPath,
		Inventory:  inventory,
		TempDir:    runtime, // Store runtime for cleanup (legacy)
		WorkDir:    workDir, // Working directory for scripts
	}

	if options.Verbose {
		fmt.Printf("✅ Container environment ready: %s\n", containerName)
	}

	return envCtx, nil
}

// cleanupContainer removes a container
func cleanupContainer(portunixPath, containerName, runtime string, useDirectRuntime bool) {
	var cmd *exec.Cmd
	if useDirectRuntime {
		cmd = exec.Command(runtime, "rm", "-f", containerName)
	} else {
		cmd = exec.Command(portunixPath, "container", "rm", "-f", containerName)
	}
	cmd.Run() // Ignore errors
}

// setupVirtEnvironment configures a virtual machine for playbook execution
func setupVirtEnvironment(options ExecutionOptions) (*EnvironmentContext, error) {
	if options.Target == "" {
		return nil, fmt.Errorf("--target required for virt environment")
	}

	if options.Verbose {
		fmt.Printf("🖥️  Setting up VM environment for target: %s\n", options.Target)
	}

	// Get path to portunix binary for VM commands
	portunixPath, err := getPortunixBinaryPath()
	if err != nil {
		return nil, fmt.Errorf("failed to find portunix binary: %v", err)
	}

	// Check if VM exists and is running
	checkCmd := exec.Command(portunixPath, "virt", "list", "--name", options.Target)
	if err := checkCmd.Run(); err != nil {
		return nil, fmt.Errorf("VM '%s' not found or not accessible", options.Target)
	}

	// Setup SSH connectivity for VM
	sshKeyPath, sshHost, sshPort, err := setupSSHForVM(options.Target, options)
	if err != nil {
		return nil, fmt.Errorf("failed to setup SSH for VM: %v", err)
	}

	// Generate inventory
	inventory := generateVMInventory(options.Target, sshHost, sshPort, sshKeyPath)

	envCtx := &EnvironmentContext{
		Type:       "virt",
		Target:     options.Target,
		SSHHost:    sshHost,
		SSHPort:    sshPort,
		SSHUser:    "root", // Assuming root access for VMs
		SSHKeyPath: sshKeyPath,
		Inventory:  inventory,
	}

	if options.Verbose {
		fmt.Printf("✅ VM environment ready: %s (%s:%s)\n", options.Target, sshHost, sshPort)
	}

	return envCtx, nil
}

// setupSSHForContainer configures SSH access to a container
func setupSSHForContainer(containerName string, options ExecutionOptions) (string, error) {
	// For Phase 2, we assume the container already has SSH setup
	// In a full implementation, this would:
	// 1. Generate SSH key pair
	// 2. Install SSH server in container
	// 3. Configure SSH access
	// 4. Return path to private key

	// Placeholder implementation - return mock SSH key path
	sshKeyPath := fmt.Sprintf("/tmp/ptx-ansible-%s.key", containerName)

	if options.Verbose {
		fmt.Printf("   SSH key path: %s\n", sshKeyPath)
	}

	return sshKeyPath, nil
}

// setupSSHForVM configures SSH access to a virtual machine
func setupSSHForVM(vmName string, options ExecutionOptions) (string, string, string, error) {
	// For Phase 2, we assume the VM already has SSH setup
	// In a full implementation, this would:
	// 1. Query VM IP address
	// 2. Check SSH connectivity
	// 3. Setup SSH keys if needed
	// 4. Return connection details

	// Placeholder implementation - return mock connection details
	sshKeyPath := fmt.Sprintf("/tmp/ptx-ansible-%s.key", vmName)
	sshHost := "192.168.122.100" // Mock VM IP
	sshPort := "22"

	if options.Verbose {
		fmt.Printf("   SSH connection: %s@%s:%s\n", "root", sshHost, sshPort)
		fmt.Printf("   SSH key path: %s\n", sshKeyPath)
	}

	return sshKeyPath, sshHost, sshPort, nil
}

// generateContainerInventory creates Ansible inventory for container execution
func generateContainerInventory(containerName, sshKeyPath string) string {
	return fmt.Sprintf(`[containers]
%s ansible_host=localhost ansible_port=2222 ansible_user=root ansible_ssh_private_key_file=%s ansible_ssh_common_args='-o StrictHostKeyChecking=no'
`, containerName, sshKeyPath)
}

// generateVMInventory creates Ansible inventory for VM execution
func generateVMInventory(vmName, sshHost, sshPort, sshKeyPath string) string {
	return fmt.Sprintf(`[vms]
%s ansible_host=%s ansible_port=%s ansible_user=root ansible_ssh_private_key_file=%s ansible_ssh_common_args='-o StrictHostKeyChecking=no'
`, vmName, sshHost, sshPort, sshKeyPath)
}

// cleanupEnvironment cleans up the environment after playbook execution
func cleanupEnvironment(envCtx *EnvironmentContext, options ExecutionOptions) {
	if envCtx == nil {
		return
	}

	if options.Verbose {
		fmt.Printf("🧹 Cleaning up %s environment: %s\n", envCtx.Type, envCtx.Target)
	}

	// Get portunix path for cleanup commands
	portunixPath, _ := getPortunixBinaryPath()

	switch envCtx.Type {
	case "container":
		// Remove the container
		runtime := envCtx.TempDir // Runtime stored during setup
		useDirectRuntime := runtime == "docker" || runtime == "podman"

		if options.Verbose {
			fmt.Printf("   Removing container %s...\n", envCtx.Target)
		}

		cleanupContainer(portunixPath, envCtx.Target, runtime, useDirectRuntime)

		if options.Verbose {
			fmt.Printf("   Container removed\n")
		}
	case "virt":
		// For VMs, we don't automatically stop them as they might be persistent
		if options.Verbose {
			fmt.Printf("   VM %s left running (not automatically stopped)\n", envCtx.Target)
		}
	}

	// Clean up temporary SSH keys
	if envCtx.SSHKeyPath != "" {
		if err := os.Remove(envCtx.SSHKeyPath); err != nil && options.Verbose {
			fmt.Printf("   Warning: Failed to remove SSH key %s: %v\n", envCtx.SSHKeyPath, err)
		}
	}
}

// ensureContainerRuntime checks if a container runtime is available and installs one if not
func ensureContainerRuntime(options ExecutionOptions) (string, error) {
	// Check if Docker is available
	if _, err := exec.LookPath("docker"); err == nil {
		return "docker", nil
	}

	// Check if Podman is available
	if _, err := exec.LookPath("podman"); err == nil {
		return "podman", nil
	}

	// No container runtime found - try to auto-install
	if options.Verbose {
		fmt.Printf("   No container runtime found, attempting auto-install...\n")
	}

	// Get path to portunix binary
	portunixPath, err := getPortunixBinaryPath()
	if err != nil {
		return "", fmt.Errorf("no container runtime available (docker or podman). Install with: portunix install docker")
	}

	// Try to install Docker first (more common)
	if options.Verbose {
		fmt.Printf("   Installing Docker...\n")
	}

	cmd := exec.Command(portunixPath, "install", "docker")
	if options.Verbose {
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}

	if err := cmd.Run(); err != nil {
		// Docker install failed - ask user before trying Podman
		fmt.Printf("\n❌ Docker installation failed: %v\n", err)
		fmt.Printf("\n🤔 Would you like to try installing Podman instead? (y/N): ")

		var response string
		fmt.Scanln(&response)
		response = strings.ToLower(strings.TrimSpace(response))

		if response != "y" && response != "yes" {
			return "", fmt.Errorf("Docker installation failed and Podman installation was declined. Please install a container runtime manually")
		}

		if options.Verbose {
			fmt.Printf("   Trying Podman installation...\n")
		}

		cmd = exec.Command(portunixPath, "install", "podman")
		if options.Verbose {
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
		}

		if err := cmd.Run(); err != nil {
			return "", fmt.Errorf("failed to install Podman. Please install Docker or Podman manually")
		}

		// Verify Podman is now available
		if _, err := exec.LookPath("podman"); err == nil {
			if options.Verbose {
				fmt.Printf("   ✅ Podman installed successfully\n")
			}
			return "podman", nil
		}
	}

	// Verify Docker is now available
	if _, err := exec.LookPath("docker"); err == nil {
		if options.Verbose {
			fmt.Printf("   ✅ Docker installed successfully\n")
		}
		return "docker", nil
	}

	return "", fmt.Errorf("container runtime installation completed but not found in PATH. You may need to restart your terminal")
}

// generateRandomString generates a random string for unique naming
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[i%len(charset)]
	}
	return string(result)
}

// findDockerDesktopPath finds Docker Desktop executable path
func findDockerDesktopPath(verbose bool) string {
	// 1. First try "where docker" to check if Docker is on PATH
	if verbose {
		fmt.Println("   Checking PATH with 'where docker'...")
	}
	whereCmd := exec.Command("where", "docker")
	output, err := whereCmd.Output()
	if err == nil {
		// Parse first line - path to docker.exe
		lines := strings.Split(strings.TrimSpace(string(output)), "\n")
		if len(lines) > 0 {
			dockerExePath := strings.TrimSpace(lines[0])
			if verbose {
				fmt.Printf("   Found docker.exe at: %s\n", dockerExePath)
			}
			// docker.exe is typically in Docker\Docker\resources\bin\docker.exe
			// Docker Desktop.exe is in Docker\Docker\Docker Desktop.exe (not in resources!)
			dockerDir := filepath.Dir(dockerExePath) // .../resources/bin
			resourcesDir := filepath.Dir(dockerDir)  // .../resources
			dockerDockerDir := filepath.Dir(resourcesDir) // .../Docker\Docker
			dockerDesktop := filepath.Join(dockerDockerDir, "Docker Desktop.exe")
			if verbose {
				fmt.Printf("   Checking: %s\n", dockerDesktop)
			}
			if _, err := os.Stat(dockerDesktop); err == nil {
				return dockerDesktop
			}
		}
	} else if verbose {
		fmt.Println("   Docker not found on PATH")
	}

	// 2. Try registry: HKEY_LOCAL_MACHINE\SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\Docker Desktop
	if verbose {
		fmt.Println("   Checking registry: HKEY_LOCAL_MACHINE\\SOFTWARE\\Microsoft\\Windows\\CurrentVersion\\Uninstall\\Docker Desktop")
	}
	regCmd := exec.Command("reg", "query", `HKEY_LOCAL_MACHINE\SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall\Docker Desktop`, "/v", "InstallLocation")
	output, err = regCmd.Output()
	if err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.Contains(line, "InstallLocation") && strings.Contains(line, "REG_SZ") {
				parts := strings.SplitN(line, "REG_SZ", 2)
				if len(parts) == 2 {
					installDir := strings.TrimSpace(parts[1])
					if verbose {
						fmt.Printf("   InstallLocation: %s\n", installDir)
					}
					// Docker Desktop.exe is directly in install directory (not in resources!)
					dockerDesktop := filepath.Join(installDir, "Docker Desktop.exe")
					if verbose {
						fmt.Printf("   Checking: %s\n", dockerDesktop)
					}
					if _, err := os.Stat(dockerDesktop); err == nil {
						return dockerDesktop
					}
				}
			}
		}
	} else if verbose {
		fmt.Printf("   Registry key not found: %v\n", err)
	}

	// 3. Fallback: try common installation paths
	if verbose {
		fmt.Println("   Trying common installation paths...")
	}
	commonPaths := []string{
		`C:\Program Files\Docker\Docker\Docker Desktop.exe`,
		`C:\Program Files (x86)\Docker\Docker\Docker Desktop.exe`,
	}

	for _, path := range commonPaths {
		if verbose {
			fmt.Printf("   Checking: %s\n", path)
		}
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}

	return ""
}

// ensureDockerDaemonRunning checks if Docker daemon is running and starts it if needed
// If Docker is not installed, it offers to install it via portunix
func ensureDockerDaemonRunning(verbose bool, portunixPath string) error {
	// Check if Docker daemon is already running using "docker info"
	// When daemon is running, output contains "Server:" with info
	// When daemon is not running, output contains "Server:" followed by "failed to connect"
	fmt.Println("   Checking if Docker daemon is running...")
	checkCmd := exec.Command("docker", "info")
	output, _ := checkCmd.CombinedOutput()
	outputStr := string(output)
	if strings.Contains(outputStr, "Server:") && !strings.Contains(outputStr, "failed to connect") {
		fmt.Println("   ✅ Docker daemon is already running")
		return nil
	}

	fmt.Println("   Docker daemon is not running, attempting to start...")
	fmt.Println("   Searching for Docker Desktop...")

	// Try to start Docker based on OS
	if runtime.GOOS == "windows" {
		// Try to find Docker Desktop path from registry
		dockerDesktopPath := findDockerDesktopPath(verbose)
		if dockerDesktopPath == "" {
			// Docker not installed - offer to install it
			return offerDockerInstallation(verbose, portunixPath)
		}

		fmt.Printf("   Found Docker Desktop at: %s\n", dockerDesktopPath)
		fmt.Printf("   Starting Docker Desktop from: %s\n", dockerDesktopPath)

		// Start Docker Desktop
		startCmd := exec.Command("cmd", "/C", "start", "", dockerDesktopPath)
		startCmd.Run() // Ignore errors

		fmt.Println("   ⏳ Waiting for Docker daemon to start (up to 5 minutes)...")

		// Wait for Docker daemon to become available (up to 5 minutes)
		// Use "docker info" and check for Server info (not "failed to connect")
		for i := 0; i < 60; i++ {
			time.Sleep(5 * time.Second)

			checkCmd := exec.Command("docker", "info")
			output, _ := checkCmd.CombinedOutput()
			outputStr := string(output)
			// Docker daemon is running if output contains server info and no connection failure
			if strings.Contains(outputStr, "Server:") && !strings.Contains(outputStr, "failed to connect") {
				fmt.Println("   ✅ Docker daemon is running and ready!")
				return nil
			}
			fmt.Printf("   Waiting... (%d/300s)\n", (i+1)*5)
		}

		return fmt.Errorf("Docker daemon did not start within 5 minutes. Please start Docker Desktop manually")
	} else {
		// On Linux, try systemctl
		startCmd := exec.Command("sudo", "systemctl", "start", "docker")
		if err := startCmd.Run(); err != nil {
			return fmt.Errorf("failed to start Docker daemon: %v. Try: sudo systemctl start docker", err)
		}

		// Wait a bit for daemon to be ready
		time.Sleep(3 * time.Second)

		checkCmd := exec.Command("docker", "version")
		if checkCmd.Run() == nil {
			if verbose {
				fmt.Println("   ✅ Docker daemon started successfully!")
			}
			return nil
		}

		return fmt.Errorf("Docker daemon started but not responding. Check with: docker version")
	}
}

// offerDockerInstallation offers to install container runtime when none is found
func offerDockerInstallation(verbose bool, portunixPath string) error {
	fmt.Println()
	fmt.Println("❌ No container runtime found on this system.")
	fmt.Println()
	fmt.Println("A container runtime (Docker or Podman) is required to run container-based playbooks.")
	fmt.Println("Would you like to install one now?")
	fmt.Println()
	fmt.Println("  [1] Install Docker Desktop (recommended for Windows)")
	fmt.Println("  [2] Install Podman Desktop (alternative)")
	fmt.Println("  [3] Cancel")
	fmt.Println()
	fmt.Print("Your choice (1/2/3): ")

	var response string
	fmt.Scanln(&response)
	response = strings.TrimSpace(response)

	switch response {
	case "1", "":
		// Install Docker
		fmt.Println()
		fmt.Println("📦 Installing Docker...")
		fmt.Println()

		cmd := exec.Command(portunixPath, "install", "docker")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("Docker installation failed: %v\nYou can try installing manually or use: portunix install podman", err)
		}

		fmt.Println()
		fmt.Println("✅ Docker installed successfully!")
		fmt.Println("⚠️  Please start Docker Desktop and run this command again.")
		fmt.Println()
		return fmt.Errorf("Docker installed - please start Docker Desktop and run the playbook again")

	case "2":
		// Install Podman
		fmt.Println()
		fmt.Println("📦 Installing Podman...")
		fmt.Println()

		cmd := exec.Command(portunixPath, "install", "podman")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("Podman installation failed: %v", err)
		}

		fmt.Println()
		fmt.Println("✅ Podman installed successfully!")
		fmt.Println("ℹ️  Note: Your playbook specifies Docker. Consider updating it to use Podman,")
		fmt.Println("    or run: portunix playbook run <playbook> --runtime podman")
		fmt.Println()
		return fmt.Errorf("Podman installed - please run the playbook again with --runtime podman")

	case "3":
		return fmt.Errorf("installation cancelled by user")

	default:
		return fmt.Errorf("invalid choice. Please install Docker manually with: portunix install docker")
	}
}

// createTemporaryInventory creates a temporary inventory file and returns its path
func createTemporaryInventory(inventoryContent string) (string, error) {
	tmpFile, err := os.CreateTemp("", "ptx-ansible-inventory-*.ini")
	if err != nil {
		return "", fmt.Errorf("failed to create temporary inventory file: %v", err)
	}
	defer tmpFile.Close()

	if _, err := tmpFile.WriteString(inventoryContent); err != nil {
		os.Remove(tmpFile.Name())
		return "", fmt.Errorf("failed to write inventory content: %v", err)
	}

	return tmpFile.Name(), nil
}

// Phase 3: Enhanced execution functions with rollback support

// executePortunixPackagesWithRollback installs Portunix packages with conditional execution and rollback support
func executePortunixPackagesWithRollback(ptxbook *PtxbookFile, options ExecutionOptions, envCtx *EnvironmentContext, rollbackManager *RollbackManager) error {
	// Get the path to the main portunix binary
	portunixPath, err := getPortunixBinaryPath()
	if err != nil {
		return fmt.Errorf("failed to find portunix binary: %v", err)
	}

	// Create template engine for variable processing
	_ = NewTemplateEngine(ptxbook.Spec.Variables, ptxbook.Spec.Environment)

	for _, pkg := range ptxbook.Spec.Portunix.Packages {
		// Phase 3: Process package variables and templates
		processedPkg, err := ProcessPackageVariables(&pkg, ptxbook.Spec.Variables, ptxbook.Spec.Environment)
		if err != nil {
			return fmt.Errorf("failed to process package variables for %s: %v", pkg.Name, err)
		}

		// Phase 3: Evaluate conditional execution
		if pkg.When != "" {
			shouldExecute, err := ProcessConditionalExecution(pkg.When, ptxbook.Spec.Variables, ptxbook.Spec.Environment)
			if err != nil {
				return fmt.Errorf("failed to evaluate condition for package %s: %v", pkg.Name, err)
			}

			if !shouldExecute {
				if options.Verbose {
					fmt.Printf("   Skipping %s (condition not met: %s)\n", pkg.Name, pkg.When)
				}
				continue
			}
		}

		if options.Verbose {
			if processedPkg.Variant != "" {
				fmt.Printf("   Installing %s (variant: %s)...\n", processedPkg.Name, processedPkg.Variant)
			} else {
				fmt.Printf("   Installing %s...\n", processedPkg.Name)
			}
		}

		if options.DryRun {
			fmt.Printf("   [DRY-RUN] Would install: %s\n", processedPkg.Name)
			continue
		}

		// Build install command based on environment
		var cmd *exec.Cmd
		if envCtx != nil {
			// For container/VM environments, execute install inside the environment
			switch envCtx.Type {
			case "container":
				// Execute inside container - portunix is at /usr/local/bin/portunix in container
				containerPortunixPath := "/usr/local/bin/portunix"
				runtime := envCtx.TempDir // Runtime stored during setup
				useDirectRuntime := runtime == "docker" || runtime == "podman"

				var execArgs []string
				if useDirectRuntime {
					// Use explicit runtime directly
					execArgs = []string{"exec", envCtx.Target, containerPortunixPath, "install", processedPkg.Name}
					if processedPkg.Variant != "" {
						execArgs = append(execArgs, "--variant", processedPkg.Variant)
					}
					cmd = exec.Command(runtime, execArgs...)
				} else {
					// Use portunix container exec
					execArgs = []string{"container", "exec", envCtx.Target, containerPortunixPath, "install", processedPkg.Name}
					if processedPkg.Variant != "" {
						execArgs = append(execArgs, "--variant", processedPkg.Variant)
					}
					cmd = exec.Command(portunixPath, execArgs...)
				}
			case "virt":
				// Execute on VM via SSH (simplified approach)
				// In a full implementation, this would copy the binary and execute remotely
				return fmt.Errorf("portunix package installation on VMs not yet implemented in Phase 2")
			}
		} else {
			// Local execution
			args := []string{"install", processedPkg.Name}
			if processedPkg.Variant != "" {
				args = append(args, "--variant", processedPkg.Variant)
			}
			cmd = exec.Command(portunixPath, args...)
		}

		if options.Verbose {
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
		}

		// Execute with rollback tracking
		err = cmd.Run()
		success := err == nil

		// Record action for potential rollback
		environment := "local"
		if envCtx != nil {
			environment = envCtx.Type
		}
		rollbackManager.RecordAction("package_install", processedPkg.Name, processedPkg.Variant, environment, success)

		if err != nil {
			return fmt.Errorf("failed to install package %s: %v", processedPkg.Name, err)
		}
	}

	return nil
}

// executeAnsiblePlaybooksWithRollback executes Ansible playbooks with conditional execution and rollback support
func executeAnsiblePlaybooksWithRollback(ptxbook *PtxbookFile, options ExecutionOptions, envCtx *EnvironmentContext, rollbackManager *RollbackManager) error {
	playbookDir := filepath.Dir(ptxbook.Metadata.Name) // Assume playbooks are relative to .ptxbook file

	for _, playbook := range ptxbook.Spec.Ansible.Playbooks {
		// Phase 3: Process playbook variables and templates
		processedPlaybook, err := ProcessPlaybookVariables(&playbook, ptxbook.Spec.Variables, ptxbook.Spec.Environment)
		if err != nil {
			return fmt.Errorf("failed to process playbook variables for %s: %v", playbook.Path, err)
		}

		// Phase 3: Evaluate conditional execution
		if playbook.When != "" {
			shouldExecute, err := ProcessConditionalExecution(playbook.When, ptxbook.Spec.Variables, ptxbook.Spec.Environment)
			if err != nil {
				return fmt.Errorf("failed to evaluate condition for playbook %s: %v", playbook.Path, err)
			}

			if !shouldExecute {
				if options.Verbose {
					fmt.Printf("   Skipping %s (condition not met: %s)\n", playbook.Path, playbook.When)
				}
				continue
			}
		}

		if options.Verbose {
			fmt.Printf("   Executing Ansible playbook: %s\n", processedPlaybook.Path)
		}

		if options.DryRun {
			fmt.Printf("   [DRY-RUN] Would execute: %s\n", processedPlaybook.Path)
			continue
		}

		// Resolve playbook path (relative to .ptxbook file)
		playbookPath := processedPlaybook.Path
		if !filepath.IsAbs(playbookPath) {
			playbookPath = filepath.Join(playbookDir, playbookPath)
		}

		// Check if playbook file exists
		if _, err := os.Stat(playbookPath); os.IsNotExist(err) {
			return fmt.Errorf("ansible playbook not found: %s", playbookPath)
		}

		// Build ansible-playbook command
		args := []string{"ansible-playbook", playbookPath}

		// Configure inventory and connection based on environment
		if envCtx != nil {
			// Create temporary inventory file
			inventoryPath, err := createTemporaryInventory(envCtx.Inventory)
			if err != nil {
				return fmt.Errorf("failed to create inventory file: %v", err)
			}
			defer os.Remove(inventoryPath)

			args = append(args, "-i", inventoryPath)

			if options.Verbose {
				fmt.Printf("   Using inventory: %s\n", inventoryPath)
				fmt.Printf("   Target: %s (%s)\n", envCtx.Target, envCtx.Type)
			}
		} else {
			// Default to localhost execution
			args = append(args, "-i", "localhost,")
			args = append(args, "--connection", "local")
		}

		// Execute the ansible-playbook command
		cmd := exec.Command("ansible-playbook", args[1:]...)
		if options.Verbose {
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
		}

		// Execute with rollback tracking
		err = cmd.Run()
		success := err == nil

		// Record action for potential rollback
		environment := "local"
		if envCtx != nil {
			environment = envCtx.Type
		}
		rollbackManager.RecordAction("ansible_playbook", processedPlaybook.Path, "", environment, success)

		if err != nil {
			return fmt.Errorf("failed to execute ansible playbook %s: %v", processedPlaybook.Path, err)
		}
	}

	return nil
}

// evaluateScriptCondition evaluates a shell condition (e.g., "! -d ./site")
// Returns true if condition passes (script should run), false otherwise
func evaluateScriptCondition(condition string, envCtx *EnvironmentContext, options ExecutionOptions) (bool, error) {
	if condition == "" {
		return true, nil // No condition, always run
	}

	var cmd *exec.Cmd
	if envCtx != nil && envCtx.Type == "container" {
		// Evaluate condition inside container
		portunixPath, err := getPortunixBinaryPath()
		if err != nil {
			return false, err
		}
		runtime := envCtx.TempDir
		useDirectRuntime := runtime == "docker" || runtime == "podman"

		workDir := envCtx.WorkDir
		if workDir == "" {
			workDir = "/workspace"
		}

		// Use test command to evaluate condition
		testCmd := fmt.Sprintf("cd %s && test %s", workDir, condition)
		if useDirectRuntime {
			cmd = exec.Command(runtime, "exec", envCtx.Target, "sh", "-c", testCmd)
		} else {
			cmd = exec.Command(portunixPath, "container", "exec", envCtx.Target, "sh", "-c", testCmd)
		}
	} else {
		// Local evaluation - use appropriate shell for OS
		if runtime.GOOS == "windows" {
			// Windows: use cmd /c with if exist syntax
			cmd = exec.Command("cmd", "/c", fmt.Sprintf("if %s (exit 0) else (exit 1)", condition))
		} else {
			cmd = exec.Command("sh", "-c", fmt.Sprintf("test %s", condition))
		}
	}

	err := cmd.Run()
	return err == nil, nil // err == nil means condition passed
}

// executeScripts executes custom scripts defined in the playbook
func executeScripts(ptxbook *PtxbookFile, options ExecutionOptions, envCtx *EnvironmentContext) error {
	// Define script execution order - internal scripts first, then common scripts
	// Internal scripts (prefix "internal:") are executed before user scripts
	scriptOrder := []string{"internal:bin-update", "init", "create", "dev", "build", "test", "serve", "deploy"}

	// Helper function to check if script should run
	shouldRunScript := func(name string) bool {
		// Internal scripts (prefix "internal:") always run regardless of filter
		if strings.HasPrefix(name, "internal:") {
			return true
		}
		if len(options.ScriptFilter) == 0 {
			return true // No filter, run all
		}
		for _, allowed := range options.ScriptFilter {
			if strings.TrimSpace(allowed) == name {
				return true
			}
		}
		return false
	}

	// Collect all scripts (simple + extended)
	allScripts := make(map[string]struct {
		Command   string
		Condition string
	})

	// Add simple scripts
	for name, cmd := range ptxbook.Spec.Scripts {
		allScripts[name] = struct {
			Command   string
			Condition string
		}{Command: cmd, Condition: ""}
	}

	// Add/override with extended scripts
	for name, cfg := range ptxbook.Spec.ScriptsExt {
		allScripts[name] = struct {
			Command   string
			Condition string
		}{Command: cfg.Command, Condition: cfg.Condition}
	}

	for _, scriptName := range scriptOrder {
		script, exists := allScripts[scriptName]
		if !exists {
			continue
		}

		// Check if script should run based on filter
		if !shouldRunScript(scriptName) {
			if options.Verbose {
				fmt.Printf("   Skipping script '%s' (not in filter)\n", scriptName)
			}
			continue
		}

		// Evaluate condition if present
		if script.Condition != "" {
			conditionPassed, err := evaluateScriptCondition(script.Condition, envCtx, options)
			if err != nil {
				return fmt.Errorf("failed to evaluate condition for script '%s': %v", scriptName, err)
			}
			if !conditionPassed {
				if options.Verbose {
					fmt.Printf("   Skipping script '%s' (condition not met: %s)\n", scriptName, script.Condition)
				}
				continue
			}
		}

		// Handle built-in internal scripts
		if scriptName == "internal:bin-update" && script.Command == "builtin" {
			if envCtx == nil || envCtx.Type != "container" {
				if options.Verbose {
					fmt.Printf("   Skipping internal:bin-update (only runs in container environment)\n")
				}
				continue
			}

			fmt.Printf("   Running internal:bin-update (copying Portunix binaries to container)...\n")

			if options.DryRun {
				fmt.Printf("   [DRY-RUN] Would copy binaries to container\n")
				continue
			}

			// Get runtime info from environment context
			portunixPath, err := getPortunixBinaryPath()
			if err != nil {
				return fmt.Errorf("failed to find portunix binary: %v", err)
			}

			// Detect runtime (docker or podman)
			containerRuntime := "docker"
			if envCtx.Runtime != "" {
				containerRuntime = envCtx.Runtime
			}

			// Call the binary copy function
			if err := copyBinariesToContainer(envCtx.Target, options, portunixPath, containerRuntime, false); err != nil {
				return fmt.Errorf("internal:bin-update failed: %v", err)
			}

			fmt.Printf("   ✓ internal:bin-update completed\n")
			continue
		}

		// Always show which script is running
		fmt.Printf("   Running script '%s': %s\n", scriptName, script.Command)

		if options.DryRun {
			fmt.Printf("   [DRY-RUN] Would run: %s\n", script.Command)
			continue
		}

		var cmd *exec.Cmd
		if envCtx != nil && envCtx.Type == "container" {
			// Execute inside container with proper working directory
			portunixPath, err := getPortunixBinaryPath()
			if err != nil {
				return fmt.Errorf("failed to find portunix binary: %v", err)
			}
			// Wrap command with cd to working directory
			workDir := envCtx.WorkDir
			if workDir == "" {
				workDir = "/workspace"
			}
			wrappedCmd := fmt.Sprintf("cd %s && %s", workDir, script.Command)
			cmd = exec.Command(portunixPath, "container", "exec", envCtx.Target, "sh", "-c", wrappedCmd)
		} else {
			// Local execution - use appropriate shell for OS
			if runtime.GOOS == "windows" {
				cmd = exec.Command("cmd", "/c", script.Command)
			} else {
				cmd = exec.Command("sh", "-c", script.Command)
			}
		}

		// Always show output from scripts (not just in verbose mode)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("script '%s' failed: %v", scriptName, err)
		}

		fmt.Printf("   ✓ Script '%s' completed\n", scriptName)
	}

	// Execute any remaining scripts not in the predefined order
	for scriptName, script := range allScripts {
		// Skip if already executed (in predefined order)
		found := false
		for _, ordered := range scriptOrder {
			if scriptName == ordered {
				found = true
				break
			}
		}
		if found {
			continue
		}

		// Check if script should run based on filter
		if !shouldRunScript(scriptName) {
			if options.Verbose {
				fmt.Printf("   Skipping script '%s' (not in filter)\n", scriptName)
			}
			continue
		}

		// Evaluate condition if present
		if script.Condition != "" {
			conditionPassed, err := evaluateScriptCondition(script.Condition, envCtx, options)
			if err != nil {
				return fmt.Errorf("failed to evaluate condition for script '%s': %v", scriptName, err)
			}
			if !conditionPassed {
				if options.Verbose {
					fmt.Printf("   Skipping script '%s' (condition not met: %s)\n", scriptName, script.Condition)
				}
				continue
			}
		}

		// Always show which script is running
		fmt.Printf("   Running script '%s': %s\n", scriptName, script.Command)

		if options.DryRun {
			fmt.Printf("   [DRY-RUN] Would run: %s\n", script.Command)
			continue
		}

		var cmd *exec.Cmd
		if envCtx != nil && envCtx.Type == "container" {
			portunixPath, err := getPortunixBinaryPath()
			if err != nil {
				return fmt.Errorf("failed to find portunix binary: %v", err)
			}
			// Wrap command with cd to working directory
			workDir := envCtx.WorkDir
			if workDir == "" {
				workDir = "/workspace"
			}
			wrappedCmd := fmt.Sprintf("cd %s && %s", workDir, script.Command)
			cmd = exec.Command(portunixPath, "container", "exec", envCtx.Target, "sh", "-c", wrappedCmd)
		} else {
			// Local execution - use appropriate shell for OS
			if runtime.GOOS == "windows" {
				cmd = exec.Command("cmd", "/c", script.Command)
			} else {
				cmd = exec.Command("sh", "-c", script.Command)
			}
		}

		// Always show output from scripts (not just in verbose mode)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("script '%s' failed: %v", scriptName, err)
		}

		fmt.Printf("   ✓ Script '%s' completed\n", scriptName)
	}

	return nil
}

// getEnvironmentFromPlaybook extracts environment settings from playbook spec
func getEnvironmentFromPlaybook(ptxbook *PtxbookFile) (target, runtime, image, containerName string, ports, volumes []string) {
	if ptxbook.Spec.Environment == nil {
		return "local", "", "", "", nil, nil
	}

	if t, ok := ptxbook.Spec.Environment["target"].(string); ok {
		target = t
	} else {
		target = "local"
	}

	if r, ok := ptxbook.Spec.Environment["runtime"].(string); ok {
		runtime = r
	}

	if i, ok := ptxbook.Spec.Environment["image"].(string); ok {
		image = i
	}

	if cn, ok := ptxbook.Spec.Environment["container_name"].(string); ok {
		containerName = cn
	}

	// Parse ports - can be a single string or array of strings
	if p, ok := ptxbook.Spec.Environment["ports"].(string); ok {
		ports = []string{p}
	} else if pList, ok := ptxbook.Spec.Environment["ports"].([]interface{}); ok {
		for _, port := range pList {
			if ps, ok := port.(string); ok {
				ports = append(ports, ps)
			}
		}
	}

	// Parse volumes - can be a single string or array of strings
	// Supports :named suffix for Docker named volumes
	if v, ok := ptxbook.Spec.Environment["volumes"].(string); ok {
		volumes = []string{v}
	} else if vList, ok := ptxbook.Spec.Environment["volumes"].([]interface{}); ok {
		for _, vol := range vList {
			if vs, ok := vol.(string); ok {
				volumes = append(volumes, vs)
			}
		}
	}

	return
}

// parseVolumes separates bind mounts and named volumes
// Named volumes have :named suffix, e.g., "node_modules:/app/node_modules:named"
func parseVolumes(volumes []string) (bindMounts, namedVolumes []string) {
	for _, vol := range volumes {
		if strings.HasSuffix(vol, ":named") {
			// Remove :named suffix and add to named volumes
			namedVol := strings.TrimSuffix(vol, ":named")
			namedVolumes = append(namedVolumes, namedVol)
		} else {
			bindMounts = append(bindMounts, vol)
		}
	}
	return
}

// createNamedVolumes creates Docker/Podman named volumes if they don't exist
func createNamedVolumes(namedVolumes []string, runtime string, verbose bool) error {
	for _, vol := range namedVolumes {
		// Extract volume name (first part before :)
		parts := strings.Split(vol, ":")
		if len(parts) < 2 {
			continue
		}
		volumeName := parts[0]

		if verbose {
			fmt.Printf("   Creating named volume: %s\n", volumeName)
		}

		// Create volume using docker/podman
		cmd := exec.Command(runtime, "volume", "create", volumeName)
		if err := cmd.Run(); err != nil {
			// Volume might already exist, which is fine
			if verbose {
				fmt.Printf("   Volume %s already exists or creation skipped\n", volumeName)
			}
		}
	}
	return nil
}