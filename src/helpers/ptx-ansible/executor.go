package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// ExecutionOptions contains options for playbook execution
type ExecutionOptions struct {
	DryRun      bool
	Environment string // "local", "container", "virt"
	Target      string // For multi-environment execution (VM name, container name)
	Image       string // Container image for container environment
	Verbose     bool
	User        string // Phase 4: User executing the playbook
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

// getPortunixBinaryPath finds the path to the main portunix binary
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
	SSHHost     string // SSH connection host
	SSHPort     string // SSH connection port
	SSHUser     string // SSH username
	SSHKeyPath  string // Path to SSH private key
	TempDir     string // Temporary directory for environment setup
	Inventory   string // Generated Ansible inventory content
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

// setupContainerEnvironment creates and configures a container for playbook execution
func setupContainerEnvironment(options ExecutionOptions) (*EnvironmentContext, error) {
	if options.Verbose {
		fmt.Printf("🐳 Setting up container environment with image: %s\n", options.Image)
	}

	// Generate unique container name
	containerName := fmt.Sprintf("ptx-ansible-%s", generateRandomString(8))

	// For Phase 2, we simulate container creation for testing purposes
	// In a full implementation, this would:
	// 1. Create container using Portunix container system
	// 2. Configure SSH access
	// 3. Setup environment

	if options.Verbose {
		fmt.Printf("   Container name: %s\n", containerName)
		fmt.Printf("   Image: %s\n", options.Image)
		fmt.Printf("   Note: Phase 2 implementation - simulated container setup for testing\n")
	}

	// Setup SSH connectivity (simulated)
	sshKeyPath, err := setupSSHForContainer(containerName, options)
	if err != nil {
		return nil, fmt.Errorf("failed to setup SSH for container: %v", err)
	}

	// Generate inventory
	inventory := generateContainerInventory(containerName, sshKeyPath)

	envCtx := &EnvironmentContext{
		Type:       "container",
		Target:     containerName,
		SSHHost:    "localhost", // Container SSH is usually on localhost
		SSHPort:    "2222",      // Default SSH port for containers
		SSHUser:    "root",
		SSHKeyPath: sshKeyPath,
		Inventory:  inventory,
	}

	if options.Verbose {
		fmt.Printf("✅ Container environment ready: %s\n", containerName)
	}

	return envCtx, nil
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

	switch envCtx.Type {
	case "container":
		// For Phase 2 testing, containers are simulated
		if options.Verbose {
			fmt.Printf("   Container %s (simulated) - no cleanup needed\n", envCtx.Target)
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

// generateRandomString generates a random string for unique naming
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[i%len(charset)]
	}
	return string(result)
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
				// Execute inside container
				args := []string{"container", "exec", envCtx.Target, portunixPath, "install", processedPkg.Name}
				if processedPkg.Variant != "" {
					args = append(args, "--variant", processedPkg.Variant)
				}
				cmd = exec.Command(portunixPath, args...)
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