package app

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func CreateVm(vmtype string, vmname string, iso string, basefolder string) ([]byte, error) {
	switch vmtype {
	case "qemu", "kvm":
		return createQemuVm(vmname, iso, basefolder)
	case "vbox", "virtualbox":
		return createVboxVm(vmname, iso, basefolder)
	default:
		return nil, fmt.Errorf("unsupported VM type: %s (supported: qemu, vbox)", vmtype)
	}
}

// createQemuVm creates a QEMU/KVM virtual machine
func createQemuVm(vmname string, iso string, basefolder string) ([]byte, error) {
	// Create VM directory
	vmDir := filepath.Join(basefolder, vmname)
	if err := os.MkdirAll(vmDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create VM directory: %w", err)
	}
	
	// Create disk image (60GB default)
	diskPath := filepath.Join(vmDir, fmt.Sprintf("%s.qcow2", vmname))
	createCmd := exec.Command("qemu-img", "create", "-f", "qcow2", diskPath, "60G")
	if output, err := createCmd.CombinedOutput(); err != nil {
		return nil, fmt.Errorf("failed to create disk image: %s", string(output))
	}
	
	// Create run script for the VM
	scriptPath := filepath.Join(vmDir, fmt.Sprintf("run-%s.sh", vmname))
	
	// Determine OS type from ISO name to apply appropriate settings
	osType := detectOSFromISO(iso)
	
	var qemuArgs []string
	if osType == "windows11" {
		qemuArgs = []string{
			"qemu-system-x86_64",
			"-enable-kvm",
			"-name", vmname,
			"-m", "4096",
			"-smp", "4",
			"-cpu", "host",
			"-machine", "q35,smm=on",
			"-drive", fmt.Sprintf("file=%s,format=qcow2,if=virtio", diskPath),
			"-cdrom", iso,
			"-boot", "d",
			"-device", "virtio-net,netdev=net0",
			"-netdev", "user,id=net0",
			"-vga", "qxl",
			"-device", "virtio-tablet",
			"-device", "virtio-keyboard",
			"-display", "gtk",
		}
		
		// Add UEFI support if available
		if uefiCode := findUEFIFirmware(); uefiCode != "" {
			uefiVarsPath := filepath.Join(vmDir, "OVMF_VARS.fd")
			// Copy UEFI vars template
			copyCmd := exec.Command("cp", "/usr/share/OVMF/OVMF_VARS.fd", uefiVarsPath)
			copyCmd.Run() // Ignore error - fallback will work
			
			qemuArgs = append(qemuArgs,
				"-drive", fmt.Sprintf("if=pflash,format=raw,readonly=on,file=%s", uefiCode),
				"-drive", fmt.Sprintf("if=pflash,format=raw,file=%s", uefiVarsPath),
			)
		}
	} else {
		// Standard VM settings for Linux/other OS
		qemuArgs = []string{
			"qemu-system-x86_64",
			"-enable-kvm",
			"-name", vmname,
			"-m", "4096",
			"-smp", "4",
			"-cpu", "host",
			"-drive", fmt.Sprintf("file=%s,format=qcow2", diskPath),
			"-cdrom", iso,
			"-boot", "d",
			"-vga", "virtio",
			"-nic", "user,model=virtio",
			"-display", "gtk",
		}
	}
	
	// Create run script
	script := fmt.Sprintf("#!/bin/bash\n%s\n", 
		fmt.Sprintf("%s", joinArgs(qemuArgs)))
	
	if err := os.WriteFile(scriptPath, []byte(script), 0755); err != nil {
		return nil, fmt.Errorf("failed to create run script: %w", err)
	}
	
	// Try to create via virt-install for better management
	virtInstallArgs := []string{
		"virt-install",
		"--name", vmname,
		"--memory", "4096",
		"--vcpus", "4",
		"--disk", fmt.Sprintf("path=%s,format=qcow2", diskPath),
		"--cdrom", iso,
		"--os-variant", "generic",
		"--network", "default",
		"--graphics", "spice",
		"--noautoconsole",
		"--wait", "0",
	}
	
	// Try virt-install first, fallback to script
	virtCmd := exec.Command("sudo", virtInstallArgs...)
	if output, err := virtCmd.CombinedOutput(); err == nil {
		return append([]byte("VM created successfully with virt-install.\n"), output...), nil
	}
	
	return []byte(fmt.Sprintf("VM '%s' created successfully!\nDisk: %s\nRun script: %s\n\nTo start the VM:\n  bash %s\n  or\n  portunix vm start %s\n", 
		vmname, diskPath, scriptPath, scriptPath, vmname)), nil
}

// createVboxVm creates a VirtualBox virtual machine using the preprocessor
func createVboxVm(vmname string, iso string, basefolder string) ([]byte, error) {
	exist, error := PreprocessorCheckExists()
	if error != nil {
		return nil, error
	}

	if exist {
		var cmd []string
		cmd = append(cmd, "vbox")
		cmd = append(cmd, "--create")
		cmd = append(cmd, "--type")
		cmd = append(cmd, "demo")
		cmd = append(cmd, "--vmname")
		cmd = append(cmd, vmname)
		cmd = append(cmd, "--iso")
		cmd = append(cmd, iso)
		cmd = append(cmd, "--basefolder")
		cmd = append(cmd, basefolder)

		output, error := PreprocessorExecute(cmd)
		return output, error
	}
	return nil, fmt.Errorf("VirtualBox preprocessor not available")
}

// Helper functions
func detectOSFromISO(isoPath string) string {
	filename := filepath.Base(isoPath)
	filename = strings.ToLower(filename)
	
	if strings.Contains(filename, "win11") || strings.Contains(filename, "windows11") {
		return "windows11"
	}
	if strings.Contains(filename, "win10") || strings.Contains(filename, "windows10") {
		return "windows10"
	}
	if strings.Contains(filename, "ubuntu") {
		return "ubuntu"
	}
	if strings.Contains(filename, "debian") {
		return "debian"
	}
	if strings.Contains(filename, "centos") {
		return "centos"
	}
	if strings.Contains(filename, "fedora") {
		return "fedora"
	}
	
	return "linux"
}

func findUEFIFirmware() string {
	possiblePaths := []string{
		"/usr/share/OVMF/OVMF_CODE.fd",
		"/usr/share/edk2-ovmf/x64/OVMF_CODE.fd",
		"/usr/share/qemu/OVMF_CODE.fd",
		"/usr/share/edk2/ovmf/OVMF_CODE.fd",
	}
	
	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	
	return ""
}

func joinArgs(args []string) string {
	return fmt.Sprintf("%s", args[0]) + " \\\n  " + 
		fmt.Sprintf("%v", args[1:len(args)])[1:len(fmt.Sprintf("%v", args[1:len(args)]))-1]
}
