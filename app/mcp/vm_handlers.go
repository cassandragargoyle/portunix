package mcp

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// VM Management Handlers for MCP

func (s *Server) handleVMList() (interface{}, error) {
	// Check permissions
	if !s.hasPermission("vm:list") {
		return nil, fmt.Errorf("insufficient permissions to list VMs")
	}

	// List VMs using virsh if available
	cmd := exec.Command("virsh", "list", "--all")
	output, err := cmd.Output()
	if err != nil {
		// Fallback to checking VM directory
		vmDir := filepath.Join(os.Getenv("HOME"), "VMs")
		entries, err := os.ReadDir(vmDir)
		if err != nil {
			return map[string]interface{}{
				"vms":     []interface{}{},
				"message": "No VMs found or VM directory doesn't exist",
			}, nil
		}

		var vms []map[string]interface{}
		for _, entry := range entries {
			if entry.IsDir() {
				vms = append(vms, map[string]interface{}{
					"name":   entry.Name(),
					"status": "unknown",
					"path":   filepath.Join(vmDir, entry.Name()),
				})
			}
		}

		return map[string]interface{}{
			"vms":    vms,
			"source": "filesystem",
		}, nil
	}

	// Parse virsh output
	lines := strings.Split(string(output), "\n")
	var vms []map[string]interface{}
	
	for i, line := range lines {
		if i < 2 || strings.TrimSpace(line) == "" {
			continue // Skip header lines
		}
		
		fields := strings.Fields(line)
		if len(fields) >= 3 {
			status := "shut off"
			if len(fields) > 3 {
				status = strings.Join(fields[2:], " ")
			}
			
			vms = append(vms, map[string]interface{}{
				"id":     fields[0],
				"name":   fields[1],
				"status": status,
			})
		}
	}

	return map[string]interface{}{
		"vms":    vms,
		"source": "libvirt",
	}, nil
}

func (s *Server) handleVMCreate(args map[string]interface{}) (interface{}, error) {
	// Check permissions
	if !s.hasPermission("vm:create") {
		return nil, fmt.Errorf("insufficient permissions to create VMs")
	}

	// Extract parameters
	name, ok := args["name"].(string)
	if !ok {
		return nil, fmt.Errorf("VM name is required")
	}

	osType, ok := args["os"].(string)
	if !ok {
		return nil, fmt.Errorf("OS type is required")
	}

	// Set defaults
	ram := "4G"
	if r, ok := args["ram"].(string); ok {
		ram = r
	}

	disk := "60G"
	if d, ok := args["disk"].(string); ok {
		disk = d
	}

	cpus := 4
	if c, ok := args["cpus"].(float64); ok {
		cpus = int(c)
	}

	// Determine ISO file
	isoFile := fmt.Sprintf("%s.iso", osType)
	if iso, ok := args["iso"].(string); ok {
		isoFile = iso
	}

	// Build portunix vm create command
	cmd := exec.Command("portunix", "vm", "create", name,
		"--iso", isoFile,
		"--os", osType,
		"--ram", ram,
		"--disk-size", disk,
		"--cpus", fmt.Sprintf("%d", cpus))

	output, err := cmd.CombinedOutput()
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("Failed to create VM: %v\nOutput: %s", err, string(output)),
		}, nil
	}

	return map[string]interface{}{
		"success": true,
		"name":    name,
		"os":      osType,
		"ram":     ram,
		"disk":    disk,
		"cpus":    cpus,
		"message": string(output),
	}, nil
}

func (s *Server) handleVMStart(vmName string) (interface{}, error) {
	// Check permissions
	if !s.hasPermission("vm:start") {
		return nil, fmt.Errorf("insufficient permissions to start VMs")
	}

	// Try virsh first
	cmd := exec.Command("virsh", "start", vmName)
	output, err := cmd.CombinedOutput()
	if err == nil {
		return map[string]interface{}{
			"success": true,
			"name":    vmName,
			"message": "VM started successfully via libvirt",
			"output":  string(output),
		}, nil
	}

	// Fallback to running QEMU script
	scriptPath := filepath.Join(os.Getenv("HOME"), "VMs", vmName, fmt.Sprintf("run-%s.sh", vmName))
	if _, err := os.Stat(scriptPath); err == nil {
		cmd = exec.Command("bash", scriptPath)
		if err := cmd.Start(); err != nil {
			return nil, fmt.Errorf("failed to start VM: %w", err)
		}

		return map[string]interface{}{
			"success": true,
			"name":    vmName,
			"message": "VM started via QEMU script",
			"pid":     cmd.Process.Pid,
		}, nil
	}

	return nil, fmt.Errorf("VM '%s' not found", vmName)
}

func (s *Server) handleVMStop(vmName string, force bool) (interface{}, error) {
	// Check permissions
	if !s.hasPermission("vm:stop") {
		return nil, fmt.Errorf("insufficient permissions to stop VMs")
	}

	// Try virsh shutdown
	shutdownCmd := "shutdown"
	if force {
		shutdownCmd = "destroy"
	}

	cmd := exec.Command("virsh", shutdownCmd, vmName)
	output, err := cmd.CombinedOutput()
	
	if err == nil {
		return map[string]interface{}{
			"success": true,
			"name":    vmName,
			"forced":  force,
			"message": fmt.Sprintf("VM %s successfully", shutdownCmd),
			"output":  string(output),
		}, nil
	}

	// If not force and virsh failed, try ACPI shutdown
	if !force {
		// Send ACPI shutdown signal
		cmd = exec.Command("virsh", "shutdown", vmName, "--mode", "acpi")
		output, err = cmd.CombinedOutput()
		if err == nil {
			return map[string]interface{}{
				"success": true,
				"name":    vmName,
				"message": "ACPI shutdown signal sent",
				"output":  string(output),
			}, nil
		}
	}

	return nil, fmt.Errorf("failed to stop VM '%s': %v", vmName, err)
}

func (s *Server) handleVMSnapshot(args map[string]interface{}) (interface{}, error) {
	// Check permissions
	if !s.hasPermission("vm:snapshot") {
		return nil, fmt.Errorf("insufficient permissions to manage VM snapshots")
	}

	// Extract parameters
	vmName, ok := args["vm"].(string)
	if !ok {
		return nil, fmt.Errorf("VM name is required")
	}

	action, ok := args["action"].(string)
	if !ok {
		return nil, fmt.Errorf("action is required")
	}

	switch action {
	case "create":
		snapshotName, ok := args["snapshot"].(string)
		if !ok {
			return nil, fmt.Errorf("snapshot name is required for create action")
		}

		cmd := exec.Command("virsh", "snapshot-create-as", vmName, snapshotName,
			"--description", fmt.Sprintf("Snapshot created via MCP on %s", time.Now().Format(time.RFC3339)))
		output, err := cmd.CombinedOutput()
		if err != nil {
			// Fallback to qemu-img snapshot
			diskPath := filepath.Join(os.Getenv("HOME"), "VMs", vmName, fmt.Sprintf("%s.qcow2", vmName))
			cmd = exec.Command("qemu-img", "snapshot", "-c", snapshotName, diskPath)
			output, err = cmd.CombinedOutput()
			if err != nil {
				return nil, fmt.Errorf("failed to create snapshot: %v\nOutput: %s", err, string(output))
			}
		}

		return map[string]interface{}{
			"success":  true,
			"vm":       vmName,
			"snapshot": snapshotName,
			"action":   "created",
			"message":  string(output),
		}, nil

	case "restore":
		snapshotName, ok := args["snapshot"].(string)
		if !ok {
			return nil, fmt.Errorf("snapshot name is required for restore action")
		}

		cmd := exec.Command("virsh", "snapshot-revert", vmName, snapshotName)
		output, err := cmd.CombinedOutput()
		if err != nil {
			// Fallback to qemu-img snapshot
			diskPath := filepath.Join(os.Getenv("HOME"), "VMs", vmName, fmt.Sprintf("%s.qcow2", vmName))
			cmd = exec.Command("qemu-img", "snapshot", "-a", snapshotName, diskPath)
			output, err = cmd.CombinedOutput()
			if err != nil {
				return nil, fmt.Errorf("failed to restore snapshot: %v\nOutput: %s", err, string(output))
			}
		}

		return map[string]interface{}{
			"success":  true,
			"vm":       vmName,
			"snapshot": snapshotName,
			"action":   "restored",
			"message":  string(output),
		}, nil

	case "list":
		cmd := exec.Command("virsh", "snapshot-list", vmName)
		output, err := cmd.CombinedOutput()
		if err != nil {
			// Fallback to qemu-img snapshot list
			diskPath := filepath.Join(os.Getenv("HOME"), "VMs", vmName, fmt.Sprintf("%s.qcow2", vmName))
			cmd = exec.Command("qemu-img", "snapshot", "-l", diskPath)
			output, err = cmd.CombinedOutput()
			if err != nil {
				return map[string]interface{}{
					"success":   false,
					"vm":        vmName,
					"snapshots": []interface{}{},
					"error":     fmt.Sprintf("Failed to list snapshots: %v", err),
				}, nil
			}
		}

		// Parse snapshot list
		lines := strings.Split(string(output), "\n")
		var snapshots []map[string]interface{}
		
		for _, line := range lines {
			fields := strings.Fields(line)
			if len(fields) >= 2 && !strings.Contains(line, "Name") && !strings.Contains(line, "---") {
				snapshots = append(snapshots, map[string]interface{}{
					"name": fields[0],
					"info": strings.Join(fields[1:], " "),
				})
			}
		}

		return map[string]interface{}{
			"success":   true,
			"vm":        vmName,
			"snapshots": snapshots,
		}, nil

	case "delete":
		snapshotName, ok := args["snapshot"].(string)
		if !ok {
			return nil, fmt.Errorf("snapshot name is required for delete action")
		}

		cmd := exec.Command("virsh", "snapshot-delete", vmName, snapshotName)
		output, err := cmd.CombinedOutput()
		if err != nil {
			// Fallback to qemu-img snapshot
			diskPath := filepath.Join(os.Getenv("HOME"), "VMs", vmName, fmt.Sprintf("%s.qcow2", vmName))
			cmd = exec.Command("qemu-img", "snapshot", "-d", snapshotName, diskPath)
			output, err = cmd.CombinedOutput()
			if err != nil {
				return nil, fmt.Errorf("failed to delete snapshot: %v\nOutput: %s", err, string(output))
			}
		}

		return map[string]interface{}{
			"success":  true,
			"vm":       vmName,
			"snapshot": snapshotName,
			"action":   "deleted",
			"message":  string(output),
		}, nil

	default:
		return nil, fmt.Errorf("unknown action: %s", action)
	}
}

func (s *Server) handleVMInfo(vmName string) (interface{}, error) {
	// Check permissions
	if !s.hasPermission("vm:info") {
		return nil, fmt.Errorf("insufficient permissions to get VM info")
	}

	info := map[string]interface{}{
		"name": vmName,
	}

	// Try to get libvirt info
	cmd := exec.Command("virsh", "dominfo", vmName)
	output, err := cmd.Output()
	if err == nil {
		// Parse virsh dominfo output
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				info[strings.ToLower(strings.ReplaceAll(key, " ", "_"))] = value
			}
		}
	}

	// Check disk info
	diskPath := filepath.Join(os.Getenv("HOME"), "VMs", vmName, fmt.Sprintf("%s.qcow2", vmName))
	if diskInfo, err := os.Stat(diskPath); err == nil {
		info["disk_path"] = diskPath
		info["disk_size"] = diskInfo.Size()
		
		// Get disk details using qemu-img
		cmd = exec.Command("qemu-img", "info", diskPath, "--output=json")
		if output, err := cmd.Output(); err == nil {
			var diskDetails map[string]interface{}
			if err := json.Unmarshal(output, &diskDetails); err == nil {
				info["disk_details"] = diskDetails
			}
		}
	}

	// Check if VM is running
	cmd = exec.Command("virsh", "list", "--name", "--state-running")
	output, err = cmd.Output()
	if err == nil {
		runningVMs := strings.Split(string(output), "\n")
		for _, vm := range runningVMs {
			if strings.TrimSpace(vm) == vmName {
				info["state"] = "running"
				break
			}
		}
	}
	
	if _, ok := info["state"]; !ok {
		info["state"] = "shut off"
	}

	// Get VM configuration path
	configPath := filepath.Join(os.Getenv("HOME"), "VMs", vmName, fmt.Sprintf("run-%s.sh", vmName))
	if _, err := os.Stat(configPath); err == nil {
		info["run_script"] = configPath
	}

	return info, nil
}

// Helper function to check permissions
func (s *Server) hasPermission(permission string) bool {
	// Simple permission check based on permission level
	switch s.Permissions {
	case "full":
		return true
	case "standard":
		// Standard can do most VM operations except delete
		return !strings.Contains(permission, "delete")
	case "limited":
		// Limited can only list and get info
		return strings.Contains(permission, "list") || strings.Contains(permission, "info")
	default:
		return false
	}
}