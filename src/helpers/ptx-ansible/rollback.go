package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// RollbackManager handles error recovery and rollback operations
type RollbackManager struct {
	actions         []ExecutedAction
	ptxbook         *PtxbookFile
	rollbackConfig  *PtxbookRollback
	templateEngine  *TemplateEngine
	logFile         string
	enabled         bool
}

// ExecutedAction represents a completed action that may need rollback
type ExecutedAction struct {
	Type        string    // "package_install", "ansible_playbook"
	Target      string    // Package name or playbook path
	Details     string    // Additional details for rollback
	Timestamp   time.Time
	Environment string    // "local", "container", "virt"
	Success     bool
}

// NewRollbackManager creates a new rollback manager
func NewRollbackManager(ptxbook *PtxbookFile) *RollbackManager {
	manager := &RollbackManager{
		actions:        make([]ExecutedAction, 0),
		ptxbook:        ptxbook,
		rollbackConfig: ptxbook.Spec.Rollback,
		enabled:        ptxbook.Spec.Rollback != nil && ptxbook.Spec.Rollback.Enabled,
	}

	// Create template engine for rollback processing
	if manager.enabled {
		globalVars := ptxbook.Spec.Variables
		envVars := ptxbook.Spec.Environment
		if manager.rollbackConfig.CustomVariables != nil {
			if globalVars == nil {
				globalVars = make(map[string]interface{})
			}
			for k, v := range manager.rollbackConfig.CustomVariables {
				globalVars[k] = v
			}
		}
		manager.templateEngine = NewTemplateEngine(globalVars, envVars)
	}

	// Setup logging
	if manager.enabled && manager.rollbackConfig.PreserveLogs {
		manager.setupLogging()
	}

	return manager
}

// setupLogging creates a log file for rollback operations
func (rm *RollbackManager) setupLogging() {
	logDir := "/tmp/ptx-ansible-logs"
	os.MkdirAll(logDir, 0755)

	timestamp := time.Now().Format("20060102-150405")
	rm.logFile = filepath.Join(logDir, fmt.Sprintf("rollback-%s.log", timestamp))

	// Create initial log entry
	rm.log(fmt.Sprintf("Rollback logging started for playbook: %s", rm.ptxbook.Metadata.Name))
}

// log writes a message to the rollback log file
func (rm *RollbackManager) log(message string) {
	if rm.logFile == "" {
		return
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	logEntry := fmt.Sprintf("[%s] %s\n", timestamp, message)

	if file, err := os.OpenFile(rm.logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
		defer file.Close()
		file.WriteString(logEntry)
	}
}

// RecordAction records an action that was executed
func (rm *RollbackManager) RecordAction(actionType, target, details, environment string, success bool) {
	if !rm.enabled {
		return
	}

	action := ExecutedAction{
		Type:        actionType,
		Target:      target,
		Details:     details,
		Timestamp:   time.Now(),
		Environment: environment,
		Success:     success,
	}

	rm.actions = append(rm.actions, action)
	rm.log(fmt.Sprintf("Recorded action: %s %s (%s) - Success: %t", actionType, target, environment, success))
}

// ExecuteRollback performs rollback operations when execution fails
func (rm *RollbackManager) ExecuteRollback(failureReason string) error {
	if !rm.enabled {
		return nil
	}

	rm.log(fmt.Sprintf("Starting rollback due to failure: %s", failureReason))

	fmt.Printf("🔄 Starting rollback operations...\n")
	fmt.Printf("   Reason: %s\n", failureReason)

	errors := make([]string, 0)

	// Execute custom rollback actions first
	if rm.rollbackConfig.OnFailure != nil {
		for i, action := range rm.rollbackConfig.OnFailure {
			rm.log(fmt.Sprintf("Executing custom rollback action %d: %s", i+1, action.Description))

			if err := rm.executeCustomRollbackAction(action); err != nil {
				errorMsg := fmt.Sprintf("Custom rollback action %d failed: %v", i+1, err)
				errors = append(errors, errorMsg)
				rm.log(errorMsg)
			} else {
				rm.log(fmt.Sprintf("Custom rollback action %d completed successfully", i+1))
			}
		}
	}

	// Execute automatic rollback for recorded actions (in reverse order)
	rm.log("Starting automatic rollback of recorded actions")

	for i := len(rm.actions) - 1; i >= 0; i-- {
		action := rm.actions[i]

		// Only rollback successful actions
		if !action.Success {
			continue
		}

		rm.log(fmt.Sprintf("Rolling back action: %s %s", action.Type, action.Target))

		if err := rm.rollbackAction(action); err != nil {
			errorMsg := fmt.Sprintf("Failed to rollback %s %s: %v", action.Type, action.Target, err)
			errors = append(errors, errorMsg)
			rm.log(errorMsg)
		} else {
			rm.log(fmt.Sprintf("Successfully rolled back: %s %s", action.Type, action.Target))
		}
	}

	if len(errors) > 0 {
		rm.log(fmt.Sprintf("Rollback completed with %d errors", len(errors)))
		fmt.Printf("⚠️  Rollback completed with %d errors:\n", len(errors))
		for _, err := range errors {
			fmt.Printf("   - %s\n", err)
		}
		return fmt.Errorf("rollback completed with errors: %s", strings.Join(errors, "; "))
	}

	rm.log("Rollback completed successfully")
	fmt.Printf("✅ Rollback completed successfully\n")

	if rm.logFile != "" {
		fmt.Printf("   Rollback log: %s\n", rm.logFile)
	}

	return nil
}

// executeCustomRollbackAction executes a custom rollback action
func (rm *RollbackManager) executeCustomRollbackAction(action RollbackAction) error {
	// Check conditional execution
	if action.When != "" {
		shouldExecute, err := ProcessConditionalExecution(action.When, rm.ptxbook.Spec.Variables, rm.ptxbook.Spec.Environment)
		if err != nil {
			return fmt.Errorf("failed to evaluate rollback condition: %v", err)
		}
		if !shouldExecute {
			rm.log(fmt.Sprintf("Skipping rollback action due to condition: %s", action.When))
			return nil
		}
	}

	fmt.Printf("   Executing rollback: %s\n", action.Description)

	switch action.Type {
	case "command":
		return rm.executeRollbackCommand(action)
	case "package_remove":
		return rm.executePackageRemoval(action)
	case "file_restore":
		return rm.executeFileRestore(action)
	default:
		return fmt.Errorf("unknown rollback action type: %s", action.Type)
	}
}

// executeRollbackCommand executes a command-based rollback action
func (rm *RollbackManager) executeRollbackCommand(action RollbackAction) error {
	if action.Command == "" {
		return fmt.Errorf("command is required for rollback action")
	}

	// Process command template
	command, err := rm.templateEngine.ProcessTemplate(action.Command)
	if err != nil {
		return fmt.Errorf("failed to process command template: %v", err)
	}

	// Execute command
	cmd := exec.Command("bash", "-c", command)
	output, err := cmd.CombinedOutput()

	rm.log(fmt.Sprintf("Rollback command: %s", command))
	rm.log(fmt.Sprintf("Command output: %s", string(output)))

	if err != nil {
		return fmt.Errorf("command failed: %v - output: %s", err, string(output))
	}

	return nil
}

// executePackageRemoval removes a Portunix package as rollback
func (rm *RollbackManager) executePackageRemoval(action RollbackAction) error {
	if action.Package == "" {
		return fmt.Errorf("package name is required for package removal rollback")
	}

	// Process package name template
	packageName, err := rm.templateEngine.ProcessTemplate(action.Package)
	if err != nil {
		return fmt.Errorf("failed to process package name template: %v", err)
	}

	// Get portunix binary path
	_, err = getPortunixBinaryPath()
	if err != nil {
		return fmt.Errorf("failed to find portunix binary: %v", err)
	}

	// Note: Portunix doesn't currently support package removal
	// This is a placeholder for when that functionality is added
	rm.log(fmt.Sprintf("Package removal requested for: %s (not yet implemented)", packageName))
	fmt.Printf("   Warning: Package removal not yet implemented for: %s\n", packageName)

	return nil
}

// executeFileRestore restores a file from backup
func (rm *RollbackManager) executeFileRestore(action RollbackAction) error {
	if action.Path == "" {
		return fmt.Errorf("path is required for file restore rollback")
	}

	// Process path template
	filePath, err := rm.templateEngine.ProcessTemplate(action.Path)
	if err != nil {
		return fmt.Errorf("failed to process file path template: %v", err)
	}

	// Look for backup file
	backupPath := filePath + ".ptx-backup"

	if _, err := os.Stat(backupPath); os.IsNotExist(err) {
		return fmt.Errorf("backup file not found: %s", backupPath)
	}

	// Restore file
	if err := copyFile(backupPath, filePath); err != nil {
		return fmt.Errorf("failed to restore file: %v", err)
	}

	rm.log(fmt.Sprintf("File restored: %s from %s", filePath, backupPath))

	// Remove backup
	os.Remove(backupPath)

	return nil
}

// rollbackAction performs automatic rollback for a recorded action
func (rm *RollbackManager) rollbackAction(action ExecutedAction) error {
	switch action.Type {
	case "package_install":
		return rm.rollbackPackageInstall(action)
	case "ansible_playbook":
		return rm.rollbackAnsiblePlaybook(action)
	default:
		rm.log(fmt.Sprintf("Unknown action type for rollback: %s", action.Type))
		return nil
	}
}

// rollbackPackageInstall attempts to rollback a package installation
func (rm *RollbackManager) rollbackPackageInstall(action ExecutedAction) error {
	// Note: Portunix doesn't currently support package removal
	// Log the need for manual cleanup
	rm.log(fmt.Sprintf("Package installation rollback needed for: %s (manual cleanup required)", action.Target))
	fmt.Printf("   Warning: Manual cleanup may be needed for package: %s\n", action.Target)

	return nil
}

// rollbackAnsiblePlaybook attempts to rollback an Ansible playbook execution
func (rm *RollbackManager) rollbackAnsiblePlaybook(action ExecutedAction) error {
	// Ansible playbook rollback is complex and depends on the playbook content
	// For now, we just log the need for manual inspection
	rm.log(fmt.Sprintf("Ansible playbook rollback needed for: %s (manual inspection required)", action.Target))
	fmt.Printf("   Warning: Manual inspection may be needed for playbook: %s\n", action.Target)

	return nil
}

// copyFile copies a file from source to destination
func copyFile(source, destination string) error {
	input, err := os.ReadFile(source)
	if err != nil {
		return err
	}

	return os.WriteFile(destination, input, 0644)
}

// GetLogFile returns the path to the rollback log file
func (rm *RollbackManager) GetLogFile() string {
	return rm.logFile
}

// IsEnabled returns whether rollback is enabled
func (rm *RollbackManager) IsEnabled() bool {
	return rm.enabled
}

// GetRecordedActions returns a copy of recorded actions
func (rm *RollbackManager) GetRecordedActions() []ExecutedAction {
	actions := make([]ExecutedAction, len(rm.actions))
	copy(actions, rm.actions)
	return actions
}