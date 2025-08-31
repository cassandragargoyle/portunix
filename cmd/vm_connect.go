package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"io/ioutil"

	"github.com/spf13/cobra"
)

var (
	connectPort   int
	connectViewer string
)

var vmConnectCmd = &cobra.Command{
	Use:   "connect [vm-name]",
	Short: "Connect to a running VM with SPICE client",
	Long: `Connect to a running QEMU VM using a SPICE client for clipboard support.

This command will:
- Detect the VM's SPICE port
- Launch the appropriate viewer (virt-viewer, remote-viewer, or spicy)
- Enable clipboard sharing if SPICE Guest Tools are installed`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		vmName := args[0]
		
		if err := connectToVM(vmName); err != nil {
			fmt.Printf("\n❌ Failed to connect to VM: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	vmCmd.AddCommand(vmConnectCmd)
	
	vmConnectCmd.Flags().IntVar(&connectPort, "port", 0, "SPICE port (default: auto-detect)")
	vmConnectCmd.Flags().StringVar(&connectViewer, "viewer", "auto", "Viewer to use: virt-viewer, remote-viewer, spicy, or auto")
}

func connectToVM(vmName string) error {
	fmt.Printf("\n🔍 Looking for VM '%s'...\n", vmName)
	
	// Auto-detect SPICE port if not specified
	port := connectPort
	if port == 0 {
		detectedPort, err := detectSPICEPort(vmName)
		if err != nil {
			// Default to 5900 if detection fails
			fmt.Println("⚠️  Could not auto-detect SPICE port, using default 5900")
			port = 5900
		} else {
			port = detectedPort
			fmt.Printf("✅ Detected SPICE port: %d\n", port)
		}
	}
	
	// Find available viewer
	viewer := connectViewer
	if viewer == "auto" {
		viewer = findAvailableViewer()
		if viewer == "" {
			return fmt.Errorf("no SPICE viewer found. Please install virt-viewer or spicy")
		}
	}
	
	fmt.Printf("🚀 Connecting with %s to port %d...\n", viewer, port)
	
	// Build connection command
	var connectCmd *exec.Cmd
	spiceURL := fmt.Sprintf("spice://localhost:%d", port)
	
	switch viewer {
	case "virt-viewer":
		connectCmd = exec.Command("virt-viewer", "--connect", spiceURL)
	case "remote-viewer":
		connectCmd = exec.Command("remote-viewer", spiceURL)
	case "spicy":
		connectCmd = exec.Command("spicy", "-h", "localhost", "-p", fmt.Sprintf("%d", port))
	default:
		return fmt.Errorf("unknown viewer: %s", viewer)
	}
	
	// Start viewer
	if err := connectCmd.Start(); err != nil {
		return fmt.Errorf("failed to start %s: %w", viewer, err)
	}
	
	fmt.Println("\n✅ SPICE viewer launched!")
	fmt.Println("\n📋 Clipboard support:")
	fmt.Println("  • If SPICE Guest Tools are installed in Windows, clipboard will work automatically")
	fmt.Println("  • To install: portunix install spice-guest-tools")
	fmt.Println("\n🎮 Viewer controls:")
	fmt.Println("  • Ctrl+Alt+F: Toggle fullscreen")
	fmt.Println("  • Ctrl+Alt+G: Release mouse grab")
	fmt.Println("  • View menu: USB device redirection")
	
	return nil
}

func detectSPICEPort(vmName string) (int, error) {
	// Try to read from VM run script
	vmDir := filepath.Join(getVMBaseDir(), vmName)
	runScript := filepath.Join(vmDir, fmt.Sprintf("run-%s.sh", vmName))
	
	if content, err := ioutil.ReadFile(runScript); err == nil {
		// Look for -spice port=XXXX
		scriptStr := string(content)
		if strings.Contains(scriptStr, "-spice") {
			// Extract port number
			if idx := strings.Index(scriptStr, "port="); idx != -1 {
				portStr := scriptStr[idx+5:]
				// Find the end of port number
				endIdx := strings.IndexAny(portStr, ",\n ")
				if endIdx != -1 {
					portStr = portStr[:endIdx]
					var port int
					if _, err := fmt.Sscanf(portStr, "%d", &port); err == nil {
						return port, nil
					}
				}
			}
		}
	}
	
	// Try to detect from running QEMU processes
	psCmd := exec.Command("ps", "aux")
	output, err := psCmd.Output()
	if err == nil {
		lines := strings.Split(string(output), "\n")
		for _, line := range lines {
			if strings.Contains(line, vmName) && strings.Contains(line, "-spice") {
				// Extract port from command line
				if idx := strings.Index(line, "port="); idx != -1 {
					portStr := line[idx+5:]
					endIdx := strings.IndexAny(portStr, ",\n ")
					if endIdx != -1 {
						portStr = portStr[:endIdx]
						var port int
						if _, err := fmt.Sscanf(portStr, "%d", &port); err == nil {
							return port, nil
						}
					}
				}
			}
		}
	}
	
	return 0, fmt.Errorf("could not detect SPICE port")
}

func findAvailableViewer() string {
	// Check for viewers in order of preference
	viewers := []string{"virt-viewer", "remote-viewer", "spicy"}
	
	for _, viewer := range viewers {
		if _, err := exec.LookPath(viewer); err == nil {
			return viewer
		}
	}
	
	return ""
}

func getVMBaseDir() string {
	// Default VM directory
	homeDir, _ := os.UserHomeDir()
	return filepath.Join(homeDir, "VMs")
}