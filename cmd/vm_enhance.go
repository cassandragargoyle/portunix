package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"io/ioutil"
	"bufio"

	"github.com/spf13/cobra"
)

var (
	enableClipboard bool
	enableUSB       bool
	enableAudio     bool
)

var vmEnhanceCmd = &cobra.Command{
	Use:   "enhance [vm-name]",
	Short: "Enhance existing VM with additional features",
	Long: `Enhance an existing QEMU VM with additional features like clipboard support, 
USB redirection, and improved graphics.

This command will:
- Detect current VM configuration
- Add SPICE support for clipboard sharing
- Upgrade to QXL graphics driver
- Enable USB redirection
- Provide instructions for guest tools installation`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		vmName := args[0]
		
		if err := enhanceVM(vmName); err != nil {
			fmt.Printf("\n❌ Failed to enhance VM: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	vmCmd.AddCommand(vmEnhanceCmd)
	
	vmEnhanceCmd.Flags().BoolVar(&enableClipboard, "clipboard", true, "Enable clipboard sharing via SPICE")
	vmEnhanceCmd.Flags().BoolVar(&enableUSB, "usb", true, "Enable USB redirection")
	vmEnhanceCmd.Flags().BoolVar(&enableAudio, "audio", true, "Enable audio support")
}

func enhanceVM(vmName string) error {
	fmt.Printf("\n🔍 Analyzing VM '%s'...\n", vmName)
	
	// Find VM directory
	vmDir := filepath.Join(getVMBaseDir(), vmName)
	if _, err := os.Stat(vmDir); os.IsNotExist(err) {
		return fmt.Errorf("VM '%s' not found in %s", vmName, vmDir)
	}
	
	// Look for run script
	runScript := filepath.Join(vmDir, fmt.Sprintf("run-%s.sh", vmName))
	if _, err := os.Stat(runScript); os.IsNotExist(err) {
		return fmt.Errorf("VM run script not found: %s", runScript)
	}
	
	// Read current configuration
	scriptContent, err := ioutil.ReadFile(runScript)
	if err != nil {
		return fmt.Errorf("failed to read VM script: %w", err)
	}
	
	scriptStr := string(scriptContent)
	
	// Check if already has SPICE
	if strings.Contains(scriptStr, "-spice") {
		fmt.Println("✅ VM already has SPICE support enabled")
		
		// Check for specific features
		if !strings.Contains(scriptStr, "qxl-vga") {
			fmt.Println("⚠️  VM uses basic SPICE, upgrading to QXL graphics...")
			scriptStr = upgradeToQXL(scriptStr)
		}
		
		if enableUSB && !strings.Contains(scriptStr, "usb-redir") {
			fmt.Println("➕ Adding USB redirection support...")
			scriptStr = addUSBRedirection(scriptStr)
		}
	} else {
		fmt.Println("🔄 VM currently uses VNC or SDL display")
		fmt.Println("🚀 Migrating to SPICE for clipboard support...")
		
		// Backup original script
		backupPath := runScript + ".backup"
		if err := ioutil.WriteFile(backupPath, scriptContent, 0755); err != nil {
			return fmt.Errorf("failed to create backup: %w", err)
		}
		fmt.Printf("📁 Original configuration backed up to: %s\n", backupPath)
		
		// Convert to SPICE
		scriptStr = convertToSPICE(scriptStr)
	}
	
	// Write updated script
	if err := ioutil.WriteFile(runScript, []byte(scriptStr), 0755); err != nil {
		return fmt.Errorf("failed to update VM script: %w", err)
	}
	
	fmt.Println("\n✅ VM enhancement completed!")
	fmt.Println("\n📋 Next steps:")
	fmt.Println("1. Stop the VM if it's currently running")
	fmt.Printf("2. Start the VM using: %s\n", runScript)
	fmt.Println("3. Connect with SPICE client: virt-viewer --connect spice://localhost:5900")
	fmt.Println("4. Install SPICE Guest Tools in Windows:")
	fmt.Println("   portunix install spice-guest-tools")
	fmt.Println("\n💡 After installing guest tools, clipboard sharing will be enabled!")
	
	// Offer to install SPICE tools ISO
	fmt.Println("\n❓ Would you like to download SPICE Guest Tools ISO now? (y/n)")
	reader := bufio.NewReader(os.Stdin)
	response, _ := reader.ReadString('\n')
	response = strings.TrimSpace(strings.ToLower(response))
	
	if response == "y" || response == "yes" {
		downloadSPICETools(vmDir)
	}
	
	return nil
}

func upgradeToQXL(script string) string {
	// Replace standard VGA with QXL
	script = strings.ReplaceAll(script, "-vga qxl", "")
	script = strings.ReplaceAll(script, "-vga std", "")
	script = strings.ReplaceAll(script, "-vga cirrus", "")
	script = strings.ReplaceAll(script, "-vga vmware", "")
	
	// Add enhanced QXL device before -spice
	qxlDevice := "  -device qxl-vga,ram_size=67108864,vram_size=67108864,vgamem_mb=16 \\\n"
	
	if !strings.Contains(script, "qxl-vga") {
		script = strings.Replace(script, "  -spice", qxlDevice+"  -spice", 1)
	}
	
	return script
}

func addUSBRedirection(script string) string {
	usbDevices := `  -device ich9-usb-ehci1,id=usb \
  -device ich9-usb-uhci1,masterbus=usb.0,firstport=0,multifunction=on \
  -device ich9-usb-uhci2,masterbus=usb.0,firstport=2 \
  -device ich9-usb-uhci3,masterbus=usb.0,firstport=4 \
  -chardev spicevmc,name=usbredir,id=usbredirchardev1 \
  -device usb-redir,chardev=usbredirchardev1,id=usbredirdev1 \
`
	
	// Add USB devices after SPICE configuration
	if strings.Contains(script, "virtserialport") {
		insertPoint := strings.Index(script, "virtserialport")
		endOfLine := strings.Index(script[insertPoint:], "\n")
		if endOfLine != -1 {
			insertPos := insertPoint + endOfLine + 1
			script = script[:insertPos] + usbDevices + script[insertPos:]
		}
	}
	
	return script
}

func convertToSPICE(script string) string {
	// Remove existing display options
	script = strings.ReplaceAll(script, "-vnc :0", "")
	script = strings.ReplaceAll(script, "-vnc :1", "")
	script = strings.ReplaceAll(script, "-display sdl", "")
	script = strings.ReplaceAll(script, "-display gtk", "")
	script = strings.ReplaceAll(script, "-nographic", "")
	
	// Remove old VGA options
	script = strings.ReplaceAll(script, "-vga qxl", "")
	script = strings.ReplaceAll(script, "-vga std", "")
	script = strings.ReplaceAll(script, "-vga cirrus", "")
	
	// Add SPICE configuration
	spiceConfig := `  -device qxl-vga,ram_size=67108864,vram_size=67108864,vgamem_mb=16 \
  -spice port=5900,addr=127.0.0.1,disable-ticketing=on,image-compression=auto_glz,streaming-video=filter \
  -device virtio-serial-pci \
  -chardev spicevmc,id=spicechannel0,name=vdagent \
  -device virtserialport,chardev=spicechannel0,name=com.redhat.spice.0 \
`
	
	if enableUSB {
		spiceConfig += `  -device ich9-usb-ehci1,id=usb \
  -device ich9-usb-uhci1,masterbus=usb.0,firstport=0,multifunction=on \
  -device ich9-usb-uhci2,masterbus=usb.0,firstport=2 \
  -device ich9-usb-uhci3,masterbus=usb.0,firstport=4 \
  -chardev spicevmc,name=usbredir,id=usbredirchardev1 \
  -device usb-redir,chardev=usbredirchardev1,id=usbredirdev1 \
`
	}
	
	// Find a good insertion point (after network configuration)
	if strings.Contains(script, "-netdev") {
		insertPoint := strings.Index(script, "-netdev")
		endOfLine := strings.Index(script[insertPoint:], "\n")
		if endOfLine != -1 {
			insertPos := insertPoint + endOfLine + 1
			script = script[:insertPos] + spiceConfig + script[insertPos:]
		}
	} else {
		// Add before the last line (usually the closing of the script)
		lines := strings.Split(script, "\n")
		if len(lines) > 1 {
			lines[len(lines)-1] = spiceConfig + lines[len(lines)-1]
			script = strings.Join(lines, "\n")
		}
	}
	
	return script
}

func downloadSPICETools(vmDir string) {
	fmt.Println("\n📥 Downloading SPICE Guest Tools...")
	
	// Create downloads directory in VM folder
	downloadsDir := filepath.Join(vmDir, "downloads")
	os.MkdirAll(downloadsDir, 0755)
	
	spiceToolsPath := filepath.Join(downloadsDir, "spice-guest-tools.exe")
	
	// Download using wget or curl
	downloadCmd := exec.Command("wget", 
		"-O", spiceToolsPath,
		"https://www.spice-space.org/download/windows/spice-guest-tools/spice-guest-tools-latest.exe")
	
	if err := downloadCmd.Run(); err != nil {
		// Try curl as fallback
		downloadCmd = exec.Command("curl",
			"-L", "-o", spiceToolsPath,
			"https://www.spice-space.org/download/windows/spice-guest-tools/spice-guest-tools-latest.exe")
		
		if err := downloadCmd.Run(); err != nil {
			fmt.Printf("⚠️  Failed to download SPICE tools: %v\n", err)
			fmt.Println("You can manually download from:")
			fmt.Println("https://www.spice-space.org/download/windows/spice-guest-tools/")
			return
		}
	}
	
	fmt.Printf("\n✅ SPICE Guest Tools downloaded to: %s\n", spiceToolsPath)
	fmt.Println("\n📋 Installation instructions:")
	fmt.Println("1. Copy the file to your Windows VM")
	fmt.Println("2. Run the installer as Administrator")
	fmt.Println("3. Restart Windows after installation")
	fmt.Println("4. Clipboard sharing will be enabled automatically!")
}