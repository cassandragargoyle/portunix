package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"portunix.ai/app/virt"
)

// virtTemplateCmd represents the virt template command
var virtTemplateCmd = &cobra.Command{
	Use:   "template",
	Short: "Manage virtual machine templates",
	Long: `Manage virtual machine templates for easy VM creation.

Templates provide pre-configured settings for different operating systems:
- Recommended resource allocation (RAM, CPU, disk)
- OS-specific optimizations and features
- Required ISO files and drivers
- Post-installation scripts

Available subcommands:
  list  - List all available templates
  show  - Show detailed template information

Examples:
  portunix virt template list
  portunix virt template show ubuntu-24.04
  portunix virt template show windows11`,
}

// virtTemplateListCmd represents the template list command
var virtTemplateListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available VM templates",
	Long:  `List all available virtual machine templates with their descriptions.`,
	Run: func(cmd *cobra.Command, args []string) {
		templates, err := virt.ListTemplates()
		if err != nil {
			fmt.Printf("Error loading templates: %v\n", err)
			os.Exit(1)
		}

		if len(templates) == 0 {
			fmt.Println("No templates found.")
			return
		}

		fmt.Println("Available VM Templates:")
		fmt.Println("======================")
		fmt.Printf("%-20s %-30s %-10s %-10s\n", "NAME", "DESCRIPTION", "MIN RAM", "REC RAM")
		fmt.Printf("%-20s %-30s %-10s %-10s\n", "----", "-----------", "-------", "-------")

		for name, template := range templates {
			fmt.Printf("%-20s %-30s %-10s %-10s\n",
				name, template.Description, template.MinRAM, template.RecommendedRAM)
		}

		fmt.Printf("\nTo use a template:\n")
		fmt.Printf("  portunix virt create myvm --template <template-name>\n")
		fmt.Printf("\nFor details: portunix virt template show <template-name>\n")
	},
}

// virtTemplateShowCmd represents the template show command
var virtTemplateShowCmd = &cobra.Command{
	Use:   "show [template-name]",
	Short: "Show detailed information about a template",
	Long:  `Show detailed information about a virtual machine template including all configuration options.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		templateName := args[0]

		template, err := virt.GetTemplate(templateName)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			fmt.Println("\nAvailable templates:")
			templates, _ := virt.ListTemplates()
			for name := range templates {
				fmt.Printf("  %s\n", name)
			}
			os.Exit(1)
		}

		fmt.Printf("Template: %s\n", templateName)
		fmt.Printf("================\n\n")
		fmt.Printf("Name:         %s\n", template.Name)
		fmt.Printf("Description:  %s\n", template.Description)
		fmt.Printf("ISO File:     %s\n", template.ISO)
		fmt.Printf("OS Variant:   %s\n", template.OSVariant)
		fmt.Printf("\nResource Requirements:\n")
		fmt.Printf("  Minimum RAM:     %s\n", template.MinRAM)
		fmt.Printf("  Recommended RAM: %s\n", template.RecommendedRAM)
		fmt.Printf("  Minimum Disk:    %s\n", template.MinDisk)
		fmt.Printf("  Recommended Disk:%s\n", template.RecommendedDisk)

		if len(template.Features) > 0 {
			fmt.Printf("\nFeatures:\n")
			for _, feature := range template.Features {
				fmt.Printf("  - %s\n", feature)
			}
		}

		if len(template.RequiredFeatures) > 0 {
			fmt.Printf("\nRequired Features:\n")
			for feature, value := range template.RequiredFeatures {
				fmt.Printf("  - %s: %s\n", feature, value)
			}
		}

		if len(template.Drivers) > 0 {
			fmt.Printf("\nRequired Drivers:\n")
			for _, driver := range template.Drivers {
				fmt.Printf("  - %s\n", driver)
			}
		}

		if len(template.PostInstall) > 0 {
			fmt.Printf("\nPost-Installation Steps:\n")
			for _, step := range template.PostInstall {
				fmt.Printf("  - %s\n", step)
			}
		}

		fmt.Printf("\nTo create a VM with this template:\n")
		fmt.Printf("  portunix virt create myvm --template %s\n", templateName)

		fmt.Printf("\nTo download the required ISO:\n")
		fmt.Printf("  portunix virt iso download %s\n", getISONameFromFilename(template.ISO))
	},
}

// Helper function to map ISO filenames back to downloadable names
func getISONameFromFilename(filename string) string {
	filenameToName := map[string]string{
		"ubuntu-24.04-desktop-amd64.iso":  "ubuntu-24.04",
		"ubuntu-24.04.3-server-amd64.iso": "ubuntu-24.04-server",
		"ubuntu-22.04-desktop-amd64.iso":  "ubuntu-22.04",
		"ubuntu-22.04.3-server-amd64.iso": "ubuntu-22.04-server",
		"debian-12.4.0-amd64-netinst.iso": "debian-12",
		"Win11_24H2_English_x64v2.iso":    "windows11-eval",
		"Win10_22H2_English_x64.iso":      "windows10-eval",
	}

	if name, exists := filenameToName[filename]; exists {
		return name
	}
	return filename
}

func init() {
	// Add template commands to virt
	virtCmd.AddCommand(virtTemplateCmd)
	virtTemplateCmd.AddCommand(virtTemplateListCmd)
	virtTemplateCmd.AddCommand(virtTemplateShowCmd)
}