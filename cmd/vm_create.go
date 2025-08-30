package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"portunix.cz/app/install"
)

var (
	vmISO      string
	vmDiskSize string
	vmRAM      string
	vmCPUs     int
	vmName     string
	vmOS       string
	vmBridge   bool
)

// vmCreateCmd represents the vm create command
var vmCreateCmd = &cobra.Command{
	Use:   "create [name]",
	Short: "Create a new virtual machine",
	Long: `Create a new virtual machine with specified configuration.
	
Supports creating VMs for:
- Windows 11 (with TPM and Secure Boot)
- Windows 10
- Various Linux distributions`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Set VM name
		if len(args) > 0 {
			vmName = args[0]
		}
		if vmName == "" {
			fmt.Printf("Error: %v\n", fmt.Errorf("VM name is required")); os.Exit(1)
		}
		
		// Validate ISO file
		if vmISO == "" {
			fmt.Printf("Error: %v\n", fmt.Errorf("ISO file is required (use --iso flag)"))
			os.Exit(1)
		}
		
		// Check and install QEMU if needed
		if err := ensureQEMU(); err != nil {
			fmt.Printf("Error: %v\n", fmt.Errorf("failed to ensure QEMU installation: %w", err)); os.Exit(1)
		}
		
		// Check if ISO exists, download if needed
		vmISO = ensureISO(vmISO, vmOS)
		
		// Create VM directory
		vmDir := filepath.Join(os.Getenv("HOME"), "VMs", vmName)
		if err := os.MkdirAll(vmDir, 0755); err != nil {
			fmt.Printf("Error: %v\n", fmt.Errorf("failed to create VM directory: %w", err)); os.Exit(1)
		}
		
		diskPath := filepath.Join(vmDir, fmt.Sprintf("%s.qcow2", vmName))
		
		fmt.Printf("Creating VM '%s'...\n", vmName)
		fmt.Printf("  ISO: %s\n", vmISO)
		fmt.Printf("  Disk: %s (size: %s)\n", diskPath, vmDiskSize)
		fmt.Printf("  RAM: %s\n", vmRAM)
		fmt.Printf("  CPUs: %d\n", vmCPUs)
		fmt.Printf("  OS Type: %s\n", vmOS)
		
		// Create disk image
		fmt.Println("\nCreating disk image...")
		createCmd := exec.Command("qemu-img", "create", "-f", "qcow2", diskPath, vmDiskSize)
		createCmd.Stdout = os.Stdout
		createCmd.Stderr = os.Stderr
		if err := createCmd.Run(); err != nil {
			fmt.Printf("Error: %v\n", fmt.Errorf("failed to create disk image: %w", err)); os.Exit(1)
		}
		
		// Determine if we need special Windows 11 settings
		if vmOS == "windows11" {
			if err := createWindows11VM(vmName, diskPath, vmISO); err != nil {
				fmt.Printf("Error: %v\n", fmt.Errorf("failed to create Windows 11 VM: %w", err)); os.Exit(1)
			}
		} else {
			if err := createStandardVM(vmName, diskPath, vmISO); err != nil {
				fmt.Printf("Error: %v\n", fmt.Errorf("failed to create VM: %w", err)); os.Exit(1)
			}
		}
		
		fmt.Println("\n✅ VM created successfully!")
		fmt.Printf("\nTo start the VM, run:\n  portunix vm start %s\n", vmName)
		fmt.Printf("\nTo create a snapshot after installation:\n  portunix vm snapshot create %s clean-install\n", vmName)
	},
}

func createWindows11VM(name, diskPath, isoPath string) error {
	fmt.Println("Creating Windows 11 VM with TPM and Secure Boot support...")
	
	// Check for required Windows 11 components
	if err := checkWindows11Requirements(); err != nil {
		return fmt.Errorf("Windows 11 requirements check failed: %w", err)
	}
	
	// Create UEFI vars file for the VM
	vmDir := filepath.Dir(diskPath)
	uefiVarsPath := filepath.Join(vmDir, "OVMF_VARS.fd")
	
	// Copy UEFI template (prefer Secure Boot enabled)
	uefiTemplate := findUEFIVarsTemplate()
	if uefiTemplate == "" {
		return fmt.Errorf("UEFI firmware not found. Please install OVMF package (apt: ovmf, dnf/yum: edk2-ovmf)")
	}
	
	fmt.Printf("Using UEFI firmware: %s\n", uefiTemplate)
	if strings.Contains(uefiTemplate, "secboot") {
		fmt.Println("✅ Secure Boot enabled firmware detected")
	} else {
		fmt.Println("⚠️  Using standard UEFI firmware (Secure Boot may not be available)")
	}
	
	copyCmd := exec.Command("cp", uefiTemplate, uefiVarsPath)
	if err := copyCmd.Run(); err != nil {
		return fmt.Errorf("failed to copy UEFI vars: %w", err)
	}
	
	// Create TPM directory
	tpmDir := filepath.Join(vmDir, "tpm")
	if err := os.MkdirAll(tpmDir, 0755); err != nil {
		return fmt.Errorf("failed to create TPM directory: %w", err)
	}
	
	// Check if we have drivers ISO for VirtIO
	driversISO := ""
	possibleDriverPaths := []string{
		"/home/zdenek/ISOs/virtio-win-0.1.271.iso",
		filepath.Join(os.Getenv("HOME"), "ISOs", "virtio-win.iso"),
		filepath.Join(os.Getenv("HOME"), "Downloads", "virtio-win.iso"),
	}
	for _, path := range possibleDriverPaths {
		if _, err := os.Stat(path); err == nil {
			driversISO = path
			fmt.Printf("✅ Found VirtIO drivers: %s\n", path)
			break
		}
	}
	
	// Build virt-install command for Windows 11
	args := []string{
		"virt-install",
		"--name", name,
		"--memory", strings.TrimSuffix(vmRAM, "G") + "000",  // Convert G to MB
		"--vcpus", strconv.Itoa(vmCPUs),
		"--disk", fmt.Sprintf("path=%s,format=qcow2,bus=virtio", diskPath),
		"--cdrom", isoPath,
	}
	
	// Add drivers ISO if available
	if driversISO != "" {
		args = append(args, "--disk", fmt.Sprintf("path=%s,device=cdrom,bus=sata", driversISO))
	}
	
	args = append(args,
		"--os-variant", "win11",
		"--network", getNetworkConfig(),
		"--graphics", "spice",
		"--video", "qxl",
		"--boot", "uefi",
		"--features", "smm=on",
		"--tpm", "backend.type=emulator,backend.version=2.0,model=tpm-crb",
		"--noautoconsole",
		"--wait", "0",
	)
	
	cmd := exec.Command("sudo", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		// Fallback to QEMU command if virt-install fails
		fmt.Println("virt-install failed, trying direct QEMU command...")
		return runQEMUWindows11(name, diskPath, isoPath, tpmDir)
	}
	
	return nil
}

func runQEMUWindows11(name, diskPath, isoPath, tpmDir string) error {
	// Convert RAM to MB
	ramMB := strings.TrimSuffix(vmRAM, "G")
	ramMBInt, _ := strconv.Atoi(ramMB)
	ramMBInt = ramMBInt * 1024
	
	vmDir := filepath.Dir(diskPath)
	
	// Start swtpm for TPM emulation
	tpmSocket := filepath.Join(vmDir, "tpm-socket")
	fmt.Println("Starting TPM 2.0 emulator...")
	swtpmCmd := exec.Command("swtpm", "socket",
		"--tpmstate", fmt.Sprintf("dir=%s", tpmDir),
		"--ctrl", fmt.Sprintf("type=unixio,path=%s", tpmSocket),
		"--tpm2",
		"--daemon")
	if err := swtpmCmd.Run(); err != nil {
		fmt.Printf("Warning: Failed to start TPM emulator: %v\n", err)
		fmt.Println("Continuing without TPM (Windows 11 may not install properly)")
	}
	
	args := []string{
		"qemu-system-x86_64",
		"-enable-kvm",
		"-name", name,
		"-m", strconv.Itoa(ramMBInt),
		"-smp", strconv.Itoa(vmCPUs),
		"-cpu", "host",
		"-machine", "q35,smm=on,accel=kvm",
		"-global", "driver=cxl-type3,property=size,value=256M",
		"-global", "ICH9-LPC.disable_s3=1",
		"-drive", fmt.Sprintf("file=%s,format=qcow2,if=virtio", diskPath),
		"-cdrom", isoPath,
		"-boot", "menu=on",
		"-device", "virtio-net,netdev=net0",
		"-netdev", "user,id=net0",
		"-vga", "qxl",
		"-device", "virtio-tablet",
		"-device", "virtio-keyboard",
		"-spice", "port=5900,disable-ticketing=on",
		"-device", "virtio-serial",
		"-chardev", "spicevmc,id=spicechannel0,name=vdagent",
		"-device", "virtserialport,chardev=spicechannel0,name=com.redhat.spice.0",
	}
	
	// Add UEFI support (required for Windows 11)
	uefiCode := findUEFIFirmware()
	uefiVars := filepath.Join(vmDir, "OVMF_VARS.fd")
	
	if uefiCode != "" {
		args = append(args,
			"-drive", fmt.Sprintf("if=pflash,format=raw,readonly=on,file=%s", uefiCode),
			"-drive", fmt.Sprintf("if=pflash,format=raw,file=%s", uefiVars))
	} else {
		fmt.Println("⚠️  Warning: UEFI firmware not found, Windows 11 may not install")
	}
	
	// Add TPM device if socket exists
	if _, err := os.Stat(tpmSocket); err == nil {
		args = append(args,
			"-chardev", fmt.Sprintf("socket,id=chrtpm,path=%s", tpmSocket),
			"-tpmdev", "emulator,id=tpm0,chardev=chrtpm",
			"-device", "tpm-tis,tpmdev=tpm0")
		fmt.Println("✅ TPM 2.0 emulation enabled")
	}
	
	// Save the command to a script for easy re-running
	scriptPath := filepath.Join(filepath.Dir(diskPath), fmt.Sprintf("run-%s.sh", name))
	script := fmt.Sprintf("#!/bin/bash\n%s\n", strings.Join(args, " \\\n  "))
	if err := os.WriteFile(scriptPath, []byte(script), 0755); err != nil {
		fmt.Printf("Warning: Failed to save run script: %v\n", err)
	}
	
	fmt.Printf("\n📝 QEMU command saved to: %s\n", scriptPath)
	fmt.Println("\nStarting Windows 11 installation...")
	fmt.Println("Note: You can connect to the VM using a SPICE client on port 5900")
	
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	return cmd.Start()
}

func createStandardVM(name, diskPath, isoPath string) error {
	fmt.Printf("Creating standard VM '%s'...\n", name)
	
	// Determine OS variant
	osVariant := vmOS
	if osVariant == "" {
		osVariant = "generic"
	}
	
	// Use virt-install for better management
	args := []string{
		"virt-install",
		"--name", name,
		"--memory", strings.TrimSuffix(vmRAM, "G") + "000",  // Convert G to MB
		"--vcpus", strconv.Itoa(vmCPUs),
		"--disk", fmt.Sprintf("path=%s,format=qcow2", diskPath),
		"--cdrom", isoPath,
		"--os-variant", osVariant,
		"--network", getNetworkConfig(),
		"--graphics", "spice",
		"--noautoconsole",
		"--wait", "0",
	}
	
	cmd := exec.Command("sudo", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	if err := cmd.Run(); err != nil {
		// Fallback to basic QEMU command
		fmt.Println("virt-install failed, trying direct QEMU command...")
		return runBasicQEMU(name, diskPath, isoPath)
	}
	
	return nil
}

func runBasicQEMU(name, diskPath, isoPath string) error {
	// Convert RAM to MB
	ramMB := strings.TrimSuffix(vmRAM, "G")
	ramMBInt, _ := strconv.Atoi(ramMB)
	ramMBInt = ramMBInt * 1024
	
	args := []string{
		"qemu-system-x86_64",
		"-enable-kvm",
		"-name", name,
		"-m", strconv.Itoa(ramMBInt),
		"-smp", strconv.Itoa(vmCPUs),
		"-cpu", "host",
		"-drive", fmt.Sprintf("file=%s,format=qcow2", diskPath),
		"-cdrom", isoPath,
		"-boot", "d",
		"-vga", "virtio",
		"-display", "gtk",
	}
	
	if vmBridge {
		args = append(args, "-netdev", "bridge,id=net0,br=br0")
		args = append(args, "-device", "virtio-net,netdev=net0")
	} else {
		args = append(args, "-nic", "user,model=virtio")
	}
	
	// Save the command to a script
	scriptPath := filepath.Join(filepath.Dir(diskPath), fmt.Sprintf("run-%s.sh", name))
	script := fmt.Sprintf("#!/bin/bash\n%s\n", strings.Join(args, " \\\n  "))
	if err := os.WriteFile(scriptPath, []byte(script), 0755); err != nil {
		fmt.Printf("Warning: Failed to save run script: %v\n", err)
	}
	
	fmt.Printf("\n📝 QEMU command saved to: %s\n", scriptPath)
	
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	return cmd.Start()
}

func getNetworkConfig() string {
	if vmBridge {
		return "bridge=br0,model=virtio"
	}
	return "default"
}

// checkWindows11Requirements checks if all required components for Windows 11 VM are installed
func checkWindows11Requirements() error {
	var missingComponents []string
	
	// Check for TPM emulator (swtpm)
	if _, err := exec.LookPath("swtpm"); err != nil {
		missingComponents = append(missingComponents, "swtpm (TPM 2.0 emulator)")
	}
	
	// Check for UEFI firmware
	if findUEFIFirmware() == "" {
		missingComponents = append(missingComponents, "OVMF/UEFI firmware")
	}
	
	// Check for virt-install (optional but recommended)
	if _, err := exec.LookPath("virt-install"); err != nil {
		fmt.Println("⚠️  virt-install not found (optional, but recommended for better VM management)")
	}
	
	if len(missingComponents) > 0 {
		fmt.Println("\n❌ Missing required components for Windows 11 VM:")
		for _, comp := range missingComponents {
			fmt.Printf("  - %s\n", comp)
		}
		fmt.Println("\n🔧 To install missing components, run:")
		fmt.Println("  portunix install qemu")
		fmt.Println("\nThis will install QEMU/KVM with all Windows 11 requirements including:")
		fmt.Println("  - TPM 2.0 emulator (swtpm)")
		fmt.Println("  - UEFI/Secure Boot firmware (OVMF)")
		fmt.Println("  - Virtualization tools (virt-manager, libvirt)")
		return fmt.Errorf("missing required components")
	}
	
	fmt.Println("✅ All Windows 11 requirements are met:")
	fmt.Println("  - TPM 2.0 emulator: Available")
	fmt.Println("  - UEFI firmware: Available")
	
	return nil
}

func findUEFIFirmware() string {
	// For Windows 11, prefer Secure Boot enabled firmware
	possiblePaths := []string{
		// Secure Boot enabled firmware (preferred for Windows 11)
		"/usr/share/OVMF/OVMF_CODE.secboot.fd",
		"/usr/share/edk2-ovmf/x64/OVMF_CODE.secboot.fd",
		"/usr/share/qemu/OVMF_CODE.secboot.fd",
		"/usr/share/edk2/ovmf/OVMF_CODE.secboot.fd",
		// Standard UEFI firmware (fallback)
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

// findUEFIVarsTemplate finds the UEFI VARS template file
func findUEFIVarsTemplate() string {
	possiblePaths := []string{
		// Secure Boot enabled vars (preferred for Windows 11)
		"/usr/share/OVMF/OVMF_VARS.secboot.fd",
		"/usr/share/edk2-ovmf/x64/OVMF_VARS.secboot.fd",
		"/usr/share/qemu/OVMF_VARS.secboot.fd",
		"/usr/share/edk2/ovmf/OVMF_VARS.secboot.fd",
		// Standard UEFI vars (fallback)
		"/usr/share/OVMF/OVMF_VARS.fd",
		"/usr/share/edk2-ovmf/x64/OVMF_VARS.fd",
		"/usr/share/qemu/OVMF_VARS.fd",
		"/usr/share/edk2/ovmf/OVMF_VARS.fd",
	}
	
	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	
	return ""
}

// ensureISO checks if ISO exists, downloads if needed
func ensureISO(isoPath string, osType string) string {
	// If ISO is an absolute path and exists, use it
	if filepath.IsAbs(isoPath) {
		if _, err := os.Stat(isoPath); err == nil {
			return isoPath
		}
	}
	
	// Check common locations including cache
	cwd, _ := os.Getwd()
	commonPaths := []string{
		isoPath,
		filepath.Join(cwd, ".cache", "isos", isoPath),
		filepath.Join(os.Getenv("HOME"), ".cache", "isos", isoPath),
		filepath.Join(os.Getenv("HOME"), "Downloads", isoPath),
		filepath.Join(os.Getenv("HOME"), "Downloads", "ISOs", isoPath),
		filepath.Join(os.Getenv("HOME"), "VMs", "ISOs", isoPath),
	}
	
	for _, path := range commonPaths {
		if _, err := os.Stat(path); err == nil {
			fmt.Printf("Found ISO: %s\n", path)
			return path
		}
	}
	
	// If Windows ISO requested, search for any Windows ISO in cache
	if osType == "windows11" || osType == "windows10" {
		if cachedISO := findWindowsISOInCache(osType); cachedISO != "" {
			fmt.Printf("✅ Found cached Windows ISO: %s\n", cachedISO)
			return cachedISO
		}
	}
	
	// ISO not found, try to download it
	fmt.Printf("ISO not found: %s\n", isoPath)
	
	// Determine OS type from filename or flag
	downloadOS := osType
	if downloadOS == "generic" || downloadOS == "" {
		// Try to detect from ISO filename
		lowerISO := strings.ToLower(isoPath)
		if strings.Contains(lowerISO, "win11") || strings.Contains(lowerISO, "windows11") {
			downloadOS = "windows11"
		} else if strings.Contains(lowerISO, "win10") || strings.Contains(lowerISO, "windows10") {
			downloadOS = "windows10"
		} else if strings.Contains(lowerISO, "ubuntu") {
			downloadOS = "ubuntu"
		} else if strings.Contains(lowerISO, "debian") {
			downloadOS = "debian"
		} else {
			// Ask user what to download
			fmt.Printf("\n⚠️  Cannot determine OS type from filename: %s\n", isoPath)
			fmt.Println("Please specify OS type with --os flag (windows11, windows10, ubuntu, debian)")
			os.Exit(1)
		}
	}
	
	fmt.Printf("\n📥 Downloading %s ISO...\n", downloadOS)
	fmt.Println("This may take a while depending on your internet connection.")
	
	// Use the install system to download ISO to cache
	installer := &install.ISOInstaller{
		OSType:    downloadOS,
		Variant:   "latest",
		OutputDir: "", // Will use default cache dir (.cache/isos)
	}
	
	downloadedPath, err := installer.Download()
	if err != nil {
		if strings.Contains(err.Error(), "manual download required") {
			// Manual download instructions were already shown
			fmt.Println("\nOnce you've downloaded the ISO, run this command again:")
			fmt.Printf("  %s\n", strings.Join(os.Args, " "))
		} else {
			fmt.Printf("\n❌ Failed to download ISO: %v\n", err)
		}
		os.Exit(1)
	}
	
	fmt.Printf("\n✅ ISO downloaded successfully: %s\n", downloadedPath)
	return downloadedPath
}

// ensureQEMU checks if QEMU is installed and installs it if needed
func ensureQEMU() error {
	// Check if qemu-img is available in PATH or common locations
	if isQEMUInstalled() {
		// For Windows 11 VMs, also check TPM and UEFI components
		if vmOS == "windows11" {
			// Just show a note about components, don't fail if missing
			// The checkWindows11Requirements() will handle the detailed check
			fmt.Println("✅ QEMU/KVM is installed")
			fmt.Println("🔍 Checking Windows 11 specific requirements...")
		}
		return nil
	}
	
	fmt.Println("🔧 QEMU not found, installing automatically...")
	
	// Install QEMU using configuration-based installation system
	fmt.Println("📦 Installing QEMU/KVM stack...")
	if vmOS == "windows11" {
		fmt.Println("Installing with Windows 11 support (TPM 2.0, UEFI/Secure Boot)...")
	}
	if err := install.InstallPackage("qemu", "default"); err != nil {
		return fmt.Errorf("failed to install QEMU/KVM packages: %w", err)
	}
	
	fmt.Println("✅ QEMU/KVM installation completed!")
	if vmOS == "windows11" {
		fmt.Println("✅ Windows 11 support installed:")
		fmt.Println("  - TPM 2.0 emulator (swtpm)")
		fmt.Println("  - UEFI/Secure Boot firmware (OVMF)")
	}
	fmt.Println("\n💡 Note: If you encounter libvirt connection errors, run:")
	fmt.Println("   newgrp libvirt")
	fmt.Println("   (or log out and back in for permanent group changes)")
	return nil
}

// addUserToGroups adds the current user to libvirt and kvm groups
func addUserToGroups() error {
	// Get current username
	username := os.Getenv("USER")
	if username == "" {
		return fmt.Errorf("could not determine current username")
	}
	
	// Add to libvirt group
	cmd := exec.Command("sudo", "usermod", "-aG", "libvirt", username)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to add user to libvirt group: %w", err)
	}
	
	// Add to kvm group
	cmd = exec.Command("sudo", "usermod", "-aG", "kvm", username)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to add user to kvm group: %w", err)
	}
	
	return nil
}

// isQEMUInstalled checks if QEMU is installed in PATH or common locations
func isQEMUInstalled() bool {
	// Check if qemu-system-x86_64 is available (main virtualization binary)
	if _, err := exec.LookPath("qemu-system-x86_64"); err == nil {
		// Also check for qemu-img (needed for disk operations)
		if _, err := exec.LookPath("qemu-img"); err == nil {
			fmt.Println("✅ QEMU/KVM is already installed")
			return true
		}
	}
	
	return false
}


// findWindowsISOInCache searches for Windows ISO files in cache directory
func findWindowsISOInCache(osType string) string {
	cacheDir := filepath.Join(os.Getenv("HOME"), ".cache", "isos")
	cwdCache := filepath.Join(".cache", "isos")
	
	// Check both current working directory cache and home cache
	cacheDirs := []string{cwdCache, cacheDir}
	
	var foundISOs []string
	
	for _, dir := range cacheDirs {
		if _, err := os.Stat(dir); err != nil {
			continue // Directory doesn't exist
		}
		
		// Read directory entries
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}
		
		// Look for ISO files
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			
			filename := strings.ToLower(entry.Name())
			
			// Check if it's an ISO file
			if !strings.HasSuffix(filename, ".iso") {
				continue
			}
			
			// Check if it matches the requested OS type
			if osType == "windows11" {
				if strings.Contains(filename, "win11") || strings.Contains(filename, "windows11") || 
				   strings.Contains(filename, "windows_11") || strings.Contains(filename, "w11") ||
				   strings.Contains(filename, "win_11") {
					foundISOs = append(foundISOs, filepath.Join(dir, entry.Name()))
				}
			} else if osType == "windows10" {
				if strings.Contains(filename, "win10") || strings.Contains(filename, "windows10") || 
				   strings.Contains(filename, "windows_10") || strings.Contains(filename, "w10") ||
				   strings.Contains(filename, "win_10") {
					foundISOs = append(foundISOs, filepath.Join(dir, entry.Name()))
				}
			}
		}
	}
	
	// If no ISOs found, return empty
	if len(foundISOs) == 0 {
		return ""
	}
	
	// If only one ISO found, use it
	if len(foundISOs) == 1 {
		return foundISOs[0]
	}
	
	// Multiple ISOs found, let user choose
	fmt.Printf("\n🔍 Found %d Windows %s ISO files:\n", len(foundISOs), strings.TrimPrefix(osType, "windows"))
	for i, iso := range foundISOs {
		// Show file info
		info, _ := os.Stat(iso)
		sizeMB := info.Size() / (1024 * 1024)
		fmt.Printf("  %d. %s (%.0f MB)\n", i+1, filepath.Base(iso), float64(sizeMB))
	}
	
	// Ask user to select
	fmt.Printf("\nSelect ISO number (1-%d): ", len(foundISOs))
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("Error reading input: %v\n", err)
		return foundISOs[0] // Default to first
	}
	
	// Parse selection
	input = strings.TrimSpace(input)
	choice, err := strconv.Atoi(input)
	if err != nil || choice < 1 || choice > len(foundISOs) {
		fmt.Println("Invalid selection, using first ISO")
		return foundISOs[0]
	}
	
	return foundISOs[choice-1]
}

func init() {
	vmCreateCmd.Flags().StringVar(&vmISO, "iso", "", "Path to installation ISO file")
	vmCreateCmd.Flags().StringVar(&vmDiskSize, "disk-size", "60G", "Disk size (e.g., 60G)")
	vmCreateCmd.Flags().StringVar(&vmRAM, "ram", "4G", "RAM size (e.g., 4G, 8G)")
	vmCreateCmd.Flags().IntVar(&vmCPUs, "cpus", 4, "Number of CPU cores")
	vmCreateCmd.Flags().StringVar(&vmOS, "os", "generic", "OS type (windows11, windows10, ubuntu, generic)")
	vmCreateCmd.Flags().BoolVar(&vmBridge, "bridge", false, "Use bridge networking instead of NAT")
	
	vmCreateCmd.MarkFlagRequired("iso")
	
	vmCmd.AddCommand(vmCreateCmd)
}