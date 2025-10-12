package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"portunix.ai/app/sandbox"

	"github.com/spf13/cobra"
)

// sandboxRunInSandboxCmd represents the run-in-sandbox command
var sandboxRunInSandboxCmd = &cobra.Command{
	Use:   "run-in-sandbox [installation-type]",
	Short: "Runs Portunix installation commands inside a Windows Sandbox with SSH enabled.",
	Long: `This command orchestrates the process of running Portunix installation commands
inside a Windows Sandbox. It generates a .wsb file with SSH enabled,
maps the current Portunix executable into the sandbox, starts the sandbox,
waits for SSH to become available, and then executes the specified installation
within the sandbox.

Available installation types:
  default - Installs Python embedded, Java, and VSCode (recommended)
  empty   - Creates empty sandbox without installing any packages
  python  - Installs Python only
  java    - Installs Java only
  vscode  - Installs VSCode only

Examples:
  portunix sandbox run-in-sandbox default
  portunix sandbox run-in-sandbox empty
  portunix sandbox run-in-sandbox python --keep-files
  portunix sandbox run-in-sandbox java --no-cleanup
  portunix sandbox run-in-sandbox vscode

Flags:
  --keep-files, --no-cleanup    Keep temporary files without prompting (default: prompt user)`,
	Args: cobra.MinimumNArgs(1), // Requires at least one command to execute
	Run: func(cmd *cobra.Command, args []string) {
		// Parse flags manually since we need to handle both flags and installation type
		var keepFiles bool
		var filteredArgs []string

		for _, arg := range args {
			if arg == "--keep-files" || arg == "--no-cleanup" {
				keepFiles = true
			} else {
				filteredArgs = append(filteredArgs, arg)
			}
		}

		// Use filtered args (without flags) for installation type
		args = filteredArgs
		if len(args) == 0 {
			fmt.Println("Error: installation type is required")
			return
		}
		// 1. Check if Windows Sandbox is already running
		running, err := sandbox.CheckSandboxRunning()
		if err != nil {
			fmt.Printf("Error checking for running sandbox: %v\n", err)
			return
		}

		if running {
			shouldClose, err := sandbox.PromptToCloseSandbox()
			if err != nil {
				fmt.Printf("Error reading user input: %v\n", err)
				return
			}

			if shouldClose {
				err = sandbox.CloseSandbox()
				if err != nil {
					fmt.Printf("Error closing sandbox: %v\n", err)
					return
				}
			} else {
				fmt.Println("Aborting - Windows Sandbox is still running")
				return
			}
		}

		// 2. Determine Portunix executable path
		executablePath, err := os.Executable()
		if err != nil {
			fmt.Printf("Error getting current executable path: %v\n", err)
			return
		}

		// Create a temporary directory for portunix.exe
		tempDir := ".tmp"
		timestamp := time.Now().Format("20060102_150405")
		tempPortunixDir := filepath.Join(tempDir, fmt.Sprintf("portunix_%s", timestamp))

		// Convert to absolute path for Windows Sandbox
		absPortunixDir, err := filepath.Abs(tempPortunixDir)
		if err != nil {
			fmt.Printf("Error getting absolute path: %v\n", err)
			return
		}
		fmt.Printf("Creating temporary directory: %s\n", tempPortunixDir)
		err = os.MkdirAll(tempPortunixDir, 0755)
		if err != nil {
			fmt.Printf("Error creating temporary directory: %v\n", err)
			return
		}

		// Variable to control cleanup
		shouldCleanup := true
		defer func() {
			if shouldCleanup {
				os.RemoveAll(tempPortunixDir)
			}
		}()

		// Copy portunix.exe to the temporary directory
		destExecutablePath := filepath.Join(tempPortunixDir, filepath.Base(executablePath))

		// Check if executable already exists and has correct size
		srcStat, err := os.Stat(executablePath)
		if err != nil {
			fmt.Printf("Error getting source executable info: %v\n", err)
			return
		}

		if destStat, err := os.Stat(destExecutablePath); err == nil {
			if destStat.Size() == srcStat.Size() {
				fmt.Printf("Executable already copied (%d bytes), skipping\n", destStat.Size())
			} else {
				fmt.Printf("Executable exists but wrong size, re-copying\n")
				os.Remove(destExecutablePath)
			}
		}

		// Copy if doesn't exist or was removed
		if _, err := os.Stat(destExecutablePath); os.IsNotExist(err) {
			fmt.Printf("Copying executable from: %s\n", executablePath)
			fmt.Printf("Copying executable to: %s\n", destExecutablePath)

			// Try using hardlink first (faster)
			err = os.Link(executablePath, destExecutablePath)
			if err != nil {
				fmt.Printf("Hardlink failed, trying copy: %v\n", err)

				// Fallback to regular copy
				srcFile, err := os.Open(executablePath)
				if err != nil {
					fmt.Printf("Error opening source executable: %v\n", err)
					return
				}
				defer srcFile.Close()

				dstFile, err := os.Create(destExecutablePath)
				if err != nil {
					fmt.Printf("Error creating destination executable: %v\n", err)
					return
				}
				defer dstFile.Close()

				bytesCopied, err := io.Copy(dstFile, srcFile)
				if err != nil {
					fmt.Printf("Error copying executable: %v\n", err)
					return
				}
				fmt.Printf("Successfully copied %d bytes\n", bytesCopied)
			} else {
				fmt.Printf("Successfully created hardlink\n")
			}
		}

		// Ensure Notepad++ is available
		err = sandbox.EnsureNotepadPlusPlus(tempPortunixDir)
		if err != nil {
			fmt.Printf("Error setting up Notepad++: %v\n", err)
			return
		}

		// Ensure Win32-OpenSSH is available
		err = sandbox.EnsureWin32OpenSSH(tempPortunixDir)
		if err != nil {
			fmt.Printf("Error setting up Win32-OpenSSH: %v\n", err)
			return
		}

		// Check if Python embeddable should be cached and copied
		err = sandbox.EnsurePythonEmbeddable(args, tempPortunixDir)
		if err != nil {
			fmt.Printf("Error setting up Python embeddable: %v\n", err)
			return
		}

		// Pre-download winget for faster sandbox installation
		if shouldPreDownloadWinget(args) {
			err = sandbox.PreDownloadWinget(tempPortunixDir)
			if err != nil {
				fmt.Printf("Warning: Failed to pre-download winget: %v\n", err)
				// Continue execution, this is not critical
			}
		}

		// Check if VSCode should use custom configuration (after sandbox check)
		useVSCodeWSB := false
		for i, arg := range args {
			if arg == "install" && i+1 < len(args) && args[i+1] == "vscode" {
				useVSCodeWSB = true
				break
			}
		}

		if useVSCodeWSB {
			err = sandbox.EnsureVSCodeScripts(tempPortunixDir)
			if err != nil {
				fmt.Printf("Error setting up VSCode scripts: %v\n", err)
				return
			}

			// Use custom VSCode WSB configuration
			err = sandbox.UseVSCodeWSBFile(tempPortunixDir)
			if err != nil {
				fmt.Printf("Error setting up VSCode WSB file: %v\n", err)
				return
			}

			// Handle cleanup for VSCode sandbox - wait for user input
			fmt.Println("\nVSCode sandbox is running!")
			fmt.Printf("Temporary files preserved at: %s\n", absPortunixDir)
			fmt.Println("VSCode will be available at: C:\\Portunix\\VSCodeInstall.cmd")
			fmt.Println()
			fmt.Println("IMPORTANT: Keep this window open while VSCode is installing!")
			fmt.Println("The sandbox needs access to the mapped files during installation.")
			fmt.Println()
			fmt.Print("Press Enter after VSCode installation completes to clean up and exit...")

			reader := bufio.NewReader(os.Stdin)
			reader.ReadString('\n')

			fmt.Println("Cleaning up temporary files...")
			// Note: cleanup happens when function exits
			return
		}

		// 3. Generate .wsb file with SSH enabled and mapped folder
		// Prepend "install" to the arguments since sandbox commands typically use install subcommand
		sandboxArgs := append([]string{"install"}, args...)

		config := sandbox.SandboxConfig{
			EnableNetworking: true,
			EnableClipboard:  true,
			EnableSSH:        true,
			SSHPort:          "22", // Default SSH port
			MappedFolders: []sandbox.MappedFolder{
				{
					HostPath:    absPortunixDir,
					SandboxPath: "C:\\Portunix",
					ReadOnly:    false,
				},
			},
			SandboxArgs: sandboxArgs, // Pass arguments with "install" prepended
		}

		wsbFilePath, batchScriptPath, err := sandbox.GenerateWsbFile(config)
		if err != nil {
			fmt.Printf("Error generating .wsb file: %v\n", err)
			return
		}

		// Copy batch script to the temp directory if it exists
		if batchScriptPath != "" {
			batchSrcFile, err := os.Open(batchScriptPath)
			if err != nil {
				fmt.Printf("Error opening batch script: %v\n", err)
				return
			}
			defer batchSrcFile.Close()

			batchDestPath := filepath.Join(tempPortunixDir, "setup.bat")
			batchDstFile, err := os.Create(batchDestPath)
			if err != nil {
				fmt.Printf("Error creating batch script destination: %v\n", err)
				return
			}
			defer batchDstFile.Close()

			_, err = io.Copy(batchDstFile, batchSrcFile)
			if err != nil {
				fmt.Printf("Error copying batch script: %v\n", err)
				return
			}
			fmt.Printf("Batch script copied to: %s\n", batchDestPath)
		}

		// Copy PowerShell script to the temp directory
		psScriptSrcPath := filepath.Join(filepath.Dir(batchScriptPath), "Install-PortableOpenSSH.ps1")
		if _, err := os.Stat(psScriptSrcPath); err == nil {
			psScriptDestPath := filepath.Join(tempPortunixDir, "Install-PortableOpenSSH.ps1")
			psScriptContent, err := os.ReadFile(psScriptSrcPath)
			if err != nil {
				fmt.Printf("Error reading PowerShell script: %v\n", err)
				return
			}

			err = os.WriteFile(psScriptDestPath, psScriptContent, 0644)
			if err != nil {
				fmt.Printf("Error writing PowerShell script: %v\n", err)
				return
			}
			fmt.Printf("PowerShell script copied to: %s\n", psScriptDestPath)
		}
		// defer os.Remove(wsbFilePath) // Clean up the generated .wsb file - REMOVED

		fmt.Printf("Generated .wsb file: %s\n", wsbFilePath)

		// 4. Start the sandbox
		fmt.Println("Starting Windows Sandbox...")
		fmt.Println("⏱️  This may take several minutes, especially on the first run")
		fmt.Println("📦 Windows needs to initialize the sandbox environment")
		fmt.Println("\nMONITORING INSTRUCTIONS:")
		fmt.Println("- Watch the sandbox window for real-time progress")
		fmt.Println("- Look for timestamped messages showing each step")
		fmt.Println("- Check sandbox_setup.log on the sandbox desktop for detailed logs")
		fmt.Println("- You can re-run setup manually: C:\\Portunix\\setup.bat")
		fmt.Println("- Notepad++ available at: C:\\Portunix\\notepadplusplus\\notepad++.exe")
		fmt.Println("- Win32-OpenSSH available at: C:\\Portunix\\openssh\\OpenSSH-Win64\\")
		fmt.Println("- The window will stay open for 30 seconds after completion")
		fmt.Println("- Commands are logged with exit codes for troubleshooting")
		fmt.Println()

		err = sandbox.StartSandbox(wsbFilePath)
		if err != nil {
			fmt.Printf("Error starting sandbox: %v\n", err)
			return
		}

		// 5. Wait for SSH to be ready
		sshPort := config.SSHPort
		fmt.Printf("Waiting for SSH on port %s to become ready...\n", sshPort)
		err = sandbox.WaitForSSH(sshPort, 5*time.Minute) // 5 minutes timeout
		if err != nil {
			fmt.Printf("Error waiting for SSH: %v\n", err)
			fmt.Println()
			fmt.Println("SSH setup failed. For debugging, you can:")
			fmt.Printf("- Check sandbox window and log file\n")
			fmt.Printf("- Manually run: %s\n", filepath.Join(absPortunixDir, "setup.bat"))
			fmt.Printf("- Inspect files in: %s\n", absPortunixDir)
			fmt.Println()

			// Ask user about cleanup
			fmt.Print("Delete temporary files and exit? (y/N): ")
			reader := bufio.NewReader(os.Stdin)
			response, _ := reader.ReadString('\n')
			response = strings.TrimSpace(strings.ToLower(response))

			if response == "y" || response == "yes" {
				shouldCleanup = true
				fmt.Println("Cleaning up and exiting...")
			} else {
				shouldCleanup = false
				fmt.Printf("Temporary files preserved in: %s\n", absPortunixDir)
				fmt.Println("You can manually run setup.bat for debugging.")
			}
			return
		}

		// 6. SSH connection info and monitoring
		fmt.Println("✓ SSH is ready - commands are now executing in sandbox")
		fmt.Println("✓ Monitor the sandbox window to see real-time progress")
		fmt.Println("✓ All operations are logged with timestamps and exit codes")
		fmt.Println()

		// Display SSH connection information
		fmt.Println("📡 SSH CONNECTION INFORMATION:")
		fmt.Println("════════════════════════════════════════════════")

		// Try to read SSH info from the shared file
		var sandboxIP, username, password string
		sshInfoPath := filepath.Join(".tmp")
		entries, err := os.ReadDir(sshInfoPath)
		if err == nil {
			var newestDir string
			var newestTime int64

			for _, entry := range entries {
				if entry.IsDir() && strings.HasPrefix(entry.Name(), "portunix_") {
					info, err := entry.Info()
					if err != nil {
						continue
					}
					if info.ModTime().Unix() > newestTime {
						newestTime = info.ModTime().Unix()
						newestDir = entry.Name()
					}
				}
			}

			if newestDir != "" {
				sshInfoFile := filepath.Join(sshInfoPath, newestDir, ".tmp", "ssh_info.txt")
				if content, err := os.ReadFile(sshInfoFile); err == nil {
					lines := strings.Split(string(content), "\n")

					for _, line := range lines {
						line = strings.TrimSpace(line)
						if strings.HasPrefix(line, "IP Address:") {
							parts := strings.Split(line, ":")
							if len(parts) >= 2 {
								sandboxIP = strings.TrimSpace(parts[1])
							}
						} else if strings.HasPrefix(line, "Username:") {
							parts := strings.Split(line, ":")
							if len(parts) >= 2 {
								username = strings.TrimSpace(parts[1])
							}
						} else if strings.HasPrefix(line, "Password:") {
							parts := strings.Split(line, ":")
							if len(parts) >= 2 {
								password = strings.TrimSpace(parts[1])
							}
						}
					}
				}
			}
		}

		if sandboxIP != "" && username != "" && password != "" {
			fmt.Printf("🔗 IP Address: %s\n", sandboxIP)
			fmt.Printf("👤 Username:   %s\n", username)
			fmt.Printf("🔐 Password:   %s\n", password)
			fmt.Printf("📄 SSH Command: ssh %s@%s\n", username, sandboxIP)
			fmt.Println()
			fmt.Println("💡 CONNECTION TIPS:")
			fmt.Println("   • Open new terminal/PowerShell window")
			fmt.Printf("   • Run: ssh %s@%s\n", username, sandboxIP)
			fmt.Printf("   • Enter password: %s\n", password)
			fmt.Println("   • You can run commands directly in the sandbox")
			fmt.Println("   • Files are shared at: C:\\Portunix\\")
		} else {
			fmt.Println("🔗 IP Address: Detecting...")
			fmt.Println("👤 Username:   WDAGUtilityAccount")
			fmt.Println("🔐 Password:   (Generated in sandbox)")
			fmt.Println("📄 SSH Command: ssh WDAGUtilityAccount@<sandbox-ip>")
			fmt.Println()
			fmt.Println("💡 CONNECTION TIPS:")
			fmt.Println("   • Check sandbox window for SSH information")
			fmt.Printf("   • SSH info file: %s\n", filepath.Join(absPortunixDir, ".tmp", "ssh_info.txt"))
			fmt.Println("   • You can connect manually once setup completes")
		}

		fmt.Println("════════════════════════════════════════════════")
		fmt.Println()

		// 7. Wait for sandbox to complete and close
		fmt.Println("Waiting for sandbox to complete execution...")
		fmt.Println("The sandbox window will close automatically when commands finish.")
		fmt.Println("You can also close it manually when you're done.")
		fmt.Println()

		// Monitor sandbox until it's closed
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			running, err := sandbox.CheckSandboxRunning()
			if err != nil {
				fmt.Printf("Error checking sandbox status: %v\n", err)
				break
			}

			if !running {
				fmt.Println("✓ Sandbox has been closed")
				break
			}

			select {
			case <-ticker.C:
				// Continue monitoring
			}
		}

		// 8. Cleanup decision
		fmt.Println()
		if keepFiles {
			shouldCleanup = false
			fmt.Printf("Temporary files preserved in: %s\n", absPortunixDir)
			fmt.Println("(--keep-files flag specified)")
		} else {
			fmt.Print("Delete temporary files? (Y/n): ")
			reader := bufio.NewReader(os.Stdin)
			response, _ := reader.ReadString('\n')
			response = strings.TrimSpace(strings.ToLower(response))

			if response == "n" || response == "no" {
				shouldCleanup = false
				fmt.Printf("Temporary files preserved in: %s\n", absPortunixDir)
				fmt.Println("You can manually run setup.bat for debugging.")
			} else {
				shouldCleanup = true
				fmt.Println("Cleaning up temporary files...")
			}
		}
	},
}

func init() {
	sandboxCmd.AddCommand(sandboxRunInSandboxCmd)
}

// shouldPreDownloadWinget checks if winget should be pre-downloaded based on args
func shouldPreDownloadWinget(args []string) bool {
	for _, arg := range args {
		if arg == "install" {
			// Check if next arg is winget
			for i, a := range args {
				if a == "install" && i+1 < len(args) && args[i+1] == "winget" {
					return true
				}
			}
		}
	}
	// Always pre-download for empty sandbox (user might install winget later)
	return len(args) == 0 || (len(args) == 1 && args[0] == "empty")
}
