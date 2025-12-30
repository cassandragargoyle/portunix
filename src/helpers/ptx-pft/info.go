package main

import (
	"encoding/json"
	"fmt"
)

// MethodologyInfo describes a project template/methodology
type MethodologyInfo struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Standard    string   `json:"standard,omitempty"`
	Directories []string `json:"directories"`
	UseCases    []string `json:"use_cases"`
}

// PFTInfo contains documentation for all PTX-PFT methodologies
type PFTInfo struct {
	Version       string            `json:"version"`
	Description   string            `json:"description"`
	Methodologies []MethodologyInfo `json:"methodologies"`
	Workflow      string            `json:"workflow"`
	Commands      []CommandInfo     `json:"commands"`
}

// CommandInfo describes a PFT command
type CommandInfo struct {
	Command     string `json:"command"`
	Description string `json:"description"`
}

// handleInfoCommand handles the info subcommand
func handleInfoCommand(args []string) {
	var outputFormat string = "text"

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--format", "-f":
			if i+1 < len(args) {
				outputFormat = args[i+1]
				i++
			}
		case "--json":
			outputFormat = "json"
		case "--help", "-h":
			showInfoHelp()
			return
		}
	}

	info := getPFTInfo()

	switch outputFormat {
	case "json":
		outputJSON(info)
	default:
		outputText(info)
	}
}

func getPFTInfo() PFTInfo {
	return PFTInfo{
		Version:     version,
		Description: "PTX-PFT (Portunix Product Feedback Tool) provides structured requirements management using ISO 16355 QFD methodology.",
		Methodologies: []MethodologyInfo{
			{
				Name:        "basic",
				Description: "Minimal project structure with Voice directories only",
				Directories: []string{"voc/", "vos/", "vob/", "voe/"},
				UseCases: []string{
					"Quick prototyping",
					"Simple feedback collection",
					"Projects without formal QFD process",
				},
			},
			{
				Name:        "qfd",
				Description: "Full ISO 16355 Quality Function Deployment structure",
				Standard:    "ISO 16355",
				Directories: []string{
					"VoC/ (Voice of Customer)",
					"  verbatims/",
					"  needs/",
					"VoB/ (Voice of Business)",
					"  verbatims/",
					"  needs/",
					"VoE/ (Voice of Engineering)",
					"  verbatims/",
					"  needs/",
					"  constraints/",
					"VoS/ (Voice of Stakeholder)",
					"  verbatims/",
					"  needs/",
					"requirements/",
					"matrices/",
				},
				UseCases: []string{
					"Systematic requirements management",
					"Customer-driven product development",
					"Traceability from customer needs to technical requirements",
					"Priority-based decision making",
					"Conflict resolution between stakeholder groups",
				},
			},
		},
		Workflow: `QFD Workflow (ISO 16355):
1. Acquisition   - Capture verbatim into VoX/verbatims/
2. Structuring   - Create need in VoX/needs/ linked to verbatim
3. Analysis      - Add context, Kano category
4. Prioritization - AHP score, update matrices/
5. Translation   - Create technical requirement in requirements/

Flow: Verbatim -> Reworded Need -> Customer Requirement -> Quality Characteristic -> Technical Requirement`,
		Commands: []CommandInfo{
			{Command: "project create <name>", Description: "Create new PFT project (default: QFD template)"},
			{Command: "project create <name> --template basic", Description: "Create minimal project"},
			{Command: "configure", Description: "Configure feedback providers"},
			{Command: "sync", Description: "Bidirectional sync with external systems"},
			{Command: "pull", Description: "Pull from external system"},
			{Command: "push", Description: "Push to external system"},
			{Command: "list", Description: "List feedback items"},
			{Command: "category", Description: "Manage categories"},
			{Command: "user", Description: "Manage user registry"},
			{Command: "notify", Description: "Send notifications"},
			{Command: "report", Description: "Generate reports"},
			{Command: "export", Description: "Export data"},
			{Command: "info", Description: "Show this documentation"},
		},
	}
}

func outputJSON(info PFTInfo) {
	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		fmt.Printf("Error encoding JSON: %v\n", err)
		return
	}
	fmt.Println(string(data))
}

func outputText(info PFTInfo) {
	fmt.Println("PTX-PFT - Portunix Product Feedback Tool")
	fmt.Printf("Version: %s\n", info.Version)
	fmt.Println()
	fmt.Println(info.Description)
	fmt.Println()
	fmt.Println("=" + repeatString("=", 60))
	fmt.Println("METHODOLOGIES")
	fmt.Println("=" + repeatString("=", 60))
	fmt.Println()

	for _, m := range info.Methodologies {
		fmt.Printf("## %s\n", m.Name)
		fmt.Println(m.Description)
		if m.Standard != "" {
			fmt.Printf("Standard: %s\n", m.Standard)
		}
		fmt.Println()
		fmt.Println("Directories:")
		for _, d := range m.Directories {
			fmt.Printf("  %s\n", d)
		}
		fmt.Println()
		fmt.Println("Use Cases:")
		for _, u := range m.UseCases {
			fmt.Printf("  - %s\n", u)
		}
		fmt.Println()
	}

	fmt.Println("=" + repeatString("=", 60))
	fmt.Println("WORKFLOW")
	fmt.Println("=" + repeatString("=", 60))
	fmt.Println()
	fmt.Println(info.Workflow)
	fmt.Println()

	fmt.Println("=" + repeatString("=", 60))
	fmt.Println("COMMANDS")
	fmt.Println("=" + repeatString("=", 60))
	fmt.Println()
	for _, c := range info.Commands {
		fmt.Printf("  %-40s %s\n", c.Command, c.Description)
	}
	fmt.Println()
	fmt.Println("For detailed help: portunix pft --help")
}

func repeatString(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}

func showInfoHelp() {
	fmt.Println("Usage: portunix pft info [options]")
	fmt.Println()
	fmt.Println("Show PTX-PFT documentation and methodology information.")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --format, -f <fmt>  Output format: text, json (default: text)")
	fmt.Println("  --json              Shorthand for --format json")
	fmt.Println("  --help, -h          Show this help")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  portunix pft info")
	fmt.Println("  portunix pft info --json")
}
