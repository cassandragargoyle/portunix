package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"portunix.ai/app/sandbox"

	"github.com/spf13/cobra"
)

// sandboxGenerateCmd represents the generate command
var sandboxGenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generates a .wsb configuration file for Windows Sandbox.",
	Long: `Generates a .wsb configuration file based on the provided flags.
This file can then be used to start a customized Windows Sandbox instance.

If the --default flag is used, it will enable networking and clipboard, and set a default logon command.
Specific flags will override default settings.

Example:
  portunix sandbox generate --networking --clipboard --logon-command "notepad.exe"
  portunix sandbox generate --default
  portunix sandbox generate --map-folder "C:\\HostFolder:C:\\SandboxFolder:true"
  portunix sandbox generate --enable-ssh --ssh-port 2222`,
	Run: func(cmd *cobra.Command, args []string) {
		useDefault, _ := cmd.Flags().GetBool("default")

		enableNetworking, _ := cmd.Flags().GetBool("networking")
		enableClipboard, _ := cmd.Flags().GetBool("clipboard")
		enablePrinter, _ := cmd.Flags().GetBool("printer")
		enableMicrophone, _ := cmd.Flags().GetBool("microphone")
		enableGPU, _ := cmd.Flags().GetBool("gpu")
		logonCommand, _ := cmd.Flags().GetString("logon-command")
		mapFolders, _ := cmd.Flags().GetStringArray("map-folder")
		enableSSH, _ := cmd.Flags().GetBool("enable-ssh")
		sshPort, _ := cmd.Flags().GetString("ssh-port")

		// Apply default settings if --default flag is used and no specific flag is set
		if useDefault {
			if !cmd.Flags().Changed("networking") {
				enableNetworking = true
			}
			if !cmd.Flags().Changed("clipboard") {
				enableClipboard = true
			}
			if !cmd.Flags().Changed("logon-command") {
				logonCommand = "notepad.exe"
			}
			if !cmd.Flags().Changed("enable-ssh") {
				enableSSH = true
			}
			if !cmd.Flags().Changed("ssh-port") {
				sshPort = "22"
			}
		}

		var mappedFolders []sandbox.MappedFolder
		for _, mf := range mapFolders {
			parts := strings.Split(mf, ":")
			if len(parts) != 3 {
				fmt.Printf("Invalid --map-folder format: %s. Expected HostPath:SandboxPath:ReadOnly\n", mf)
				return
			}
			hostPath := parts[0]
			sandboxPath := parts[1]
			readOnly, err := strconv.ParseBool(parts[2])
			if err != nil {
				fmt.Printf("Invalid ReadOnly value for --map-folder: %s. Expected true or false\n", parts[2])
				return
			}
			mappedFolders = append(mappedFolders, sandbox.MappedFolder{
				HostPath:    hostPath,
				SandboxPath: sandboxPath,
				ReadOnly:    readOnly,
			})
		}

		config := sandbox.SandboxConfig{
			EnableNetworking: enableNetworking,
			EnableClipboard:  enableClipboard,
			EnablePrinter:    enablePrinter,
			EnableMicrophone: enableMicrophone,
			EnableGPU:        enableGPU,
			MappedFolders:    mappedFolders,
			LogonCommand:     logonCommand,
			EnableSSH:        enableSSH,
			SSHPort:          sshPort,
		}

		wsbFilePath, _, err := sandbox.GenerateWsbFile(config)
		if err != nil {
			fmt.Printf("Error generating .wsb file: %v\n", err)
			return
		}

		fmt.Printf("Generated .wsb file: %s\n", wsbFilePath)
	},
}

func init() {
	sandboxCmd.AddCommand(sandboxGenerateCmd)

	sandboxGenerateCmd.Flags().Bool("networking", false, "Enable networking in the sandbox")
	sandboxGenerateCmd.Flags().Bool("clipboard", true, "Enable clipboard sharing with the host")
	sandboxGenerateCmd.Flags().Bool("printer", false, "Enable printer sharing with the host")
	sandboxGenerateCmd.Flags().Bool("microphone", false, "Enable microphone sharing with the host")
	sandboxGenerateCmd.Flags().Bool("gpu", false, "Enable GPU virtualization in the sandbox")
	sandboxGenerateCmd.Flags().String("logon-command", "", "Command to execute automatically after logon in the sandbox")
	sandboxGenerateCmd.Flags().Bool("default", false, "Use default sandbox settings (enables networking, clipboard, and sets notepad.exe as logon command)")
	sandboxGenerateCmd.Flags().StringArray("map-folder", []string{}, "Map a folder into the sandbox (format: HostPath:SandboxPath:ReadOnly)")
	sandboxGenerateCmd.Flags().Bool("enable-ssh", false, "Enable OpenSSH server in the sandbox")
	sandboxGenerateCmd.Flags().String("ssh-port", "22", "Port for the OpenSSH server in the sandbox")
}
