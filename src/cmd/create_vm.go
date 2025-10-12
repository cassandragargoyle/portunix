package cmd

import (
	"fmt"

	"portunix.ai/app"

	"github.com/spf13/cobra"
)

// createVmCmd represents the vm command
var createVmCmd = &cobra.Command{
	Use:   "vm",
	Short: "Creates a new virtual machine.",
	Long: `Creates a new virtual machine using the specified parameters.

Example:
  portunix create vm --vmname my-vm --iso /path/to/ubuntu.iso --basefolder /path/to/vms`,
	Run: func(cmd *cobra.Command, args []string) {
		vmtype, _ := cmd.Flags().GetString("vmtype")
		vmname, _ := cmd.Flags().GetString("vmname")
		iso, _ := cmd.Flags().GetString("iso")
		basefolder, _ := cmd.Flags().GetString("basefolder")

		if vmname == "" || iso == "" || basefolder == "" {
			fmt.Println("Please specify vmname, iso, and basefolder.")
			return
		}

		output, err := app.CreateVm(vmtype, vmname, iso, basefolder)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		fmt.Println(string(output))
	},
}

func init() {
	createCmd.AddCommand(createVmCmd)

	createVmCmd.Flags().String("vmtype", "vbox", "Type of the virtual machine (vbox, qemu)")
	createVmCmd.Flags().String("vmname", "", "Name of the virtual machine")
	createVmCmd.Flags().String("iso", "", "Path to the ISO file")
	createVmCmd.Flags().String("basefolder", "", "Path to the base folder for the virtual machine")
}
