package cmd

import (
	"encoding/json"
	"fmt"
	"strings"
)

// CommandInfo represents a command with all its metadata
type CommandInfo struct {
	Name        string           `json:"name"`
	Brief       string           `json:"brief"`        // For basic help
	Description string           `json:"description"`   // For expert help
	Category    string           `json:"category"`      // For AI help categorization
	Parameters  []ParameterInfo  `json:"parameters,omitempty"`
	Examples    []string         `json:"examples,omitempty"`
	SubCommands []CommandInfo    `json:"subcommands,omitempty"`
}

// ParameterInfo represents a command parameter
type ParameterInfo struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Required    bool     `json:"required"`
	Description string   `json:"description"`
	Default     string   `json:"default,omitempty"`
	Choices     []string `json:"choices,omitempty"`
}

// CommandRegistry holds all command definitions
var CommandRegistry = []CommandInfo{
	// Core commands
	{
		Name:        "install",
		Brief:       "Install packages and tools",
		Description: "Install various development tools, programming languages, and packages. Supports multiple package managers and installation methods including Chocolatey, WinGet, direct downloads, and system package managers.",
		Category:    "core",
		Parameters: []ParameterInfo{
			{Name: "package", Type: "string", Required: true, Description: "Package name to install"},
			{Name: "variant", Type: "string", Required: false, Description: "Package variant (e.g., 'full', 'minimal', 'default')"},
			{Name: "dry-run", Type: "boolean", Required: false, Description: "Preview installation without making changes"},
			{Name: "force", Type: "boolean", Required: false, Description: "Force reinstall even if already installed"},
		},
		Examples: []string{
			"portunix install nodejs",
			"portunix install python --variant full",
			"portunix install java --dry-run",
		},
	},
	{
		Name:        "update",
		Brief:       "Update Portunix to the latest version",
		Description: "Check for and install updates to Portunix itself. Downloads the latest release from GitHub, verifies checksums, and safely replaces the current binary with automatic rollback on failure.",
		Category:    "core",
		Parameters: []ParameterInfo{
			{Name: "check", Type: "boolean", Required: false, Description: "Only check for updates without installing"},
			{Name: "force", Type: "boolean", Required: false, Description: "Force reinstall current version"},
		},
		Examples: []string{
			"portunix update",
			"portunix update --check",
			"portunix update --force",
		},
	},
	{
		Name:        "docker",
		Brief:       "Manage Docker containers",
		Description: "Complete Docker container management including installation, creation, execution, SSH access, and lifecycle management. Provides simplified interface for common Docker operations.",
		Category:    "container",
		Examples: []string{
			"portunix docker run ubuntu",
			"portunix docker ssh my-container",
			"portunix docker list",
		},
	},
	{
		Name:        "plugin",
		Brief:       "Manage plugins",
		Description: "Complete plugin lifecycle management including listing, installing, enabling/disabling, starting/stopping, creating new plugins, and health monitoring. Plugins use gRPC for communication.",
		Category:    "core",
		Examples: []string{
			"portunix plugin list",
			"portunix plugin install ./my-plugin",
			"portunix plugin enable agile",
		},
	},
	{
		Name:        "mcp",
		Brief:       "MCP server for AI assistants",
		Description: "Model Context Protocol server for integration with AI assistants like Claude. Provides structured tool access for AI agents to interact with the Portunix system.",
		Category:    "integration",
		Examples: []string{
			"portunix mcp configure",
			"portunix mcp serve",
			"portunix mcp status",
		},
	},
	{
		Name:        "container",
		Brief:       "Universal container management",
		Description: "Universal container interface that automatically selects between Docker and Podman based on availability. Provides consistent commands across both container runtimes.",
		Category:    "container",
		Examples: []string{
			"portunix container run ubuntu",
			"portunix container exec my-container bash",
			"portunix container list",
		},
	},
	{
		Name:        "virt",
		Brief:       "Virtual machine management",
		Description: "Universal virtualization management supporting multiple backends including QEMU/KVM, VirtualBox, VMware, and Hyper-V. Create, manage, and interact with virtual machines across platforms.",
		Category:    "virtualization",
		Examples: []string{
			"portunix virt create myvm --iso ubuntu.iso",
			"portunix virt start myvm",
			"portunix virt ssh myvm",
		},
	},
	{
		Name:        "system",
		Brief:       "System information",
		Description: "Display detailed system information including OS details, hardware specs, network configuration, and installed software versions. Useful for debugging and environment verification.",
		Category:    "utility",
		Examples: []string{
			"portunix system info",
			"portunix system os",
		},
	},
	{
		Name:        "make",
		Brief:       "Cross-platform Makefile utilities",
		Description: "Cross-platform build utilities for Makefiles. Provides portable implementations of common file operations, build metadata generation, and Go compilation helpers that work consistently across Windows, Linux, and macOS.",
		Category:    "utility",
		Examples: []string{
			"portunix make mkdir dist/bin",
			"portunix make copy src/*.go dist/",
			"portunix make rm build/",
			"portunix make version",
			"portunix make gobuild GOOS=linux GOARCH=amd64 go build -o output .",
		},
	},
	{
		Name:        "package",
		Brief:       "Package management and registry",
		Description: "Package management and registry operations including listing available packages, searching by name or description, and viewing detailed package information.",
		Category:    "core",
		Examples: []string{
			"portunix package list",
			"portunix package search python",
			"portunix package info nodejs",
		},
	},
	{
		Name:        "pft",
		Brief:       "Product feedback tool integration",
		Description: "Manage integration with external Product Feedback Tools (Fider.io, Canny, ProductBoard). Provides bidirectional synchronization between local project documentation and external feedback systems.",
		Category:    "integration",
		Examples: []string{
			"portunix pft example",
			"portunix pft configure --name \"My Product\" --path /path/to/docs",
			"portunix pft deploy",
			"portunix pft status",
		},
	},
	// Additional commands for expert level
	{
		Name:        "podman",
		Brief:       "Manage Podman containers",
		Description: "Podman container management with rootless container support. Alternative to Docker with enhanced security features.",
		Category:    "container",
	},
	{
		Name:        "sandbox",
		Brief:       "Windows Sandbox management",
		Description: "Create and manage isolated Windows Sandbox environments for safe testing and development.",
		Category:    "virtualization",
	},
	{
		Name:        "completion",
		Brief:       "Generate shell completions",
		Description: "Generate shell completion scripts for bash, zsh, fish, and PowerShell to enable tab completion for Portunix commands.",
		Category:    "utility",
	},
	{
		Name:        "config",
		Brief:       "Manage configuration",
		Description: "View and modify Portunix configuration settings including default behaviors, paths, and preferences.",
		Category:    "utility",
	},
	{
		Name:        "datastore",
		Brief:       "Datastore management",
		Description: "Manage pluggable datastore backends for persistent storage of Portunix data.",
		Category:    "core",
	},
	{
		Name:        "guid",
		Brief:       "Generate and validate GUIDs/UUIDs",
		Description: "Generate random or deterministic GUIDs/UUIDs and validate UUID format. Supports both random UUID v4 generation and deterministic UUID v5 generation from string inputs.",
		Category:    "utility",
		Parameters: []ParameterInfo{
			{Name: "subcommand", Type: "string", Required: true, Description: "Operation to perform", Choices: []string{"random", "from", "validate"}},
		},
		Examples: []string{
			"portunix guid random",
			"portunix guid from \"project-name\" \"environment-prod\"",
			"portunix guid validate \"550e8400-e29b-41d4-a716-446655440000\"",
		},
		SubCommands: []CommandInfo{
			{
				Name:        "random",
				Brief:       "Generate random UUID v4",
				Description: "Generate a cryptographically secure random UUID v4 following RFC 4122 standard",
				Category:    "utility",
			},
			{
				Name:        "from",
				Brief:       "Generate deterministic UUID from strings",
				Description: "Generate deterministic UUID v5 based on two input strings. Same inputs always produce same UUID",
				Category:    "utility",
				Parameters: []ParameterInfo{
					{Name: "string1", Type: "string", Required: true, Description: "First input string"},
					{Name: "string2", Type: "string", Required: true, Description: "Second input string"},
				},
			},
			{
				Name:        "validate",
				Brief:       "Validate UUID format",
				Description: "Check if provided string is valid UUID format according to RFC 4122",
				Category:    "utility",
				Parameters: []ParameterInfo{
					{Name: "uuid", Type: "string", Required: true, Description: "UUID string to validate"},
				},
			},
		},
	},
}

// GetBasicCommands returns only the essential commands for basic help
func GetBasicCommands() []CommandInfo {
	essentials := []string{"install", "update", "plugin", "mcp", "container", "virt", "system", "make", "package", "pft"}
	var basic []CommandInfo
	for _, cmd := range CommandRegistry {
		for _, name := range essentials {
			if cmd.Name == name {
				basic = append(basic, cmd)
				break
			}
		}
	}
	return basic
}

// GenerateBasicHelp generates the basic help output
func GenerateBasicHelp() string {
	var sb strings.Builder

	sb.WriteString("Portunix - Universal environment management tool\n\n")
	sb.WriteString("Usage: portunix [command] [options]\n\n")
	sb.WriteString("Common commands:\n")

	// Get basic commands and format them
	commands := GetBasicCommands()
	maxLen := 0
	for _, cmd := range commands {
		if len(cmd.Name) > maxLen {
			maxLen = len(cmd.Name)
		}
	}

	for _, cmd := range commands {
		sb.WriteString(fmt.Sprintf("  %-*s  %s\n", maxLen+2, cmd.Name, cmd.Brief))
	}

	// Add help levels section (MANDATORY)
	sb.WriteString("\nHelp levels:\n")
	sb.WriteString("  --help         This help - basic commands and usage (current)\n")
	sb.WriteString("  --help-expert  Extended help with all options, examples, and advanced features\n")
	sb.WriteString("  --help-ai      Machine-readable format optimized for AI/LLM parsing\n")

	sb.WriteString("\nUse 'portunix <command> --help' for command details\n")
	sb.WriteString("Use 'portunix --help-expert' for complete documentation\n")

	return sb.String()
}

// GenerateExpertHelp generates the expert help output with all details
func GenerateExpertHelp() string {
	var sb strings.Builder

	sb.WriteString("Portunix - Universal environment management tool\n\n")
	sb.WriteString("Portunix is a command-line interface (CLI) tool designed to simplify\n")
	sb.WriteString("the management of environments. It allows you to install software,\n")
	sb.WriteString("configure settings, create virtual machines, and more.\n\n")
	sb.WriteString("EXPERT DOCUMENTATION - Complete command reference\n")
	sb.WriteString("=" + strings.Repeat("=", 50) + "\n\n")

	sb.WriteString("Usage: portunix [command] [options]\n\n")

	// Group commands by category
	categories := make(map[string][]CommandInfo)
	for _, cmd := range CommandRegistry {
		if cmd.Category == "" {
			cmd.Category = "other"
		}
		categories[cmd.Category] = append(categories[cmd.Category], cmd)
	}

	// Display commands by category
	categoryOrder := []string{"core", "container", "virtualization", "integration", "utility", "other"}
	for _, cat := range categoryOrder {
		if cmds, ok := categories[cat]; ok {
			sb.WriteString(fmt.Sprintf("\n%s Commands:\n", strings.ToUpper(cat)))
			sb.WriteString(strings.Repeat("-", 40) + "\n")

			for _, cmd := range cmds {
				sb.WriteString(fmt.Sprintf("\n  %s\n", cmd.Name))
				sb.WriteString(fmt.Sprintf("    %s\n", cmd.Description))

				if len(cmd.Parameters) > 0 {
					sb.WriteString("    Parameters:\n")
					for _, param := range cmd.Parameters {
						required := ""
						if param.Required {
							required = " (required)"
						}
						sb.WriteString(fmt.Sprintf("      --%s (%s)%s: %s\n",
							param.Name, param.Type, required, param.Description))
						if param.Default != "" {
							sb.WriteString(fmt.Sprintf("        Default: %s\n", param.Default))
						}
						if len(param.Choices) > 0 {
							sb.WriteString(fmt.Sprintf("        Choices: %s\n", strings.Join(param.Choices, ", ")))
						}
					}
				}

				if len(cmd.Examples) > 0 {
					sb.WriteString("    Examples:\n")
					for _, ex := range cmd.Examples {
						sb.WriteString(fmt.Sprintf("      %s\n", ex))
					}
				}
			}
		}
	}

	// Add environment variables section
	sb.WriteString("\n\nENVIRONMENT VARIABLES:\n")
	sb.WriteString(strings.Repeat("-", 40) + "\n")
	sb.WriteString("  PORTUNIX_HOME       Base directory for Portunix data\n")
	sb.WriteString("  PORTUNIX_CACHE      Cache directory for downloads\n")
	sb.WriteString("  PORTUNIX_LOG_LEVEL  Logging level (debug, info, warn, error)\n")

	// Add configuration files section
	sb.WriteString("\n\nCONFIGURATION FILES:\n")
	sb.WriteString(strings.Repeat("-", 40) + "\n")
	sb.WriteString("  ~/.portunix/config.json    User configuration\n")
	sb.WriteString("  ~/.portunix/plugins/       Plugin directory\n")
	sb.WriteString("  ~/.portunix/cache/        Download cache\n")

	// Add help levels section
	sb.WriteString("\n\nHELP LEVELS:\n")
	sb.WriteString(strings.Repeat("-", 40) + "\n")
	sb.WriteString("  --help         Basic help with common commands\n")
	sb.WriteString("  --help-expert  This documentation - complete reference (current)\n")
	sb.WriteString("  --help-ai      JSON format for AI/LLM integration\n")

	sb.WriteString("\n\nFor more information, visit: https://github.com/CassandraGargoyle/portunix\n")

	return sb.String()
}

// GenerateAIHelp generates JSON formatted help for AI/LLM consumption
func GenerateAIHelp() (string, error) {
	// Structure for AI output
	type AIHelpOutput struct {
		Tool        string        `json:"tool"`
		Version     string        `json:"version"`
		Description string        `json:"description"`
		Commands    []CommandInfo `json:"commands"`
		Environment []struct {
			Name        string `json:"name"`
			Description string `json:"description"`
		} `json:"environment"`
	}

	output := AIHelpOutput{
		Tool:        "portunix",
		Version:     "latest",
		Description: "Portunix is a command-line interface (CLI) tool designed to simplify the management of environments. It allows you to install software, configure settings, create virtual machines, and more.",
		Commands:    CommandRegistry,
		Environment: []struct {
			Name        string `json:"name"`
			Description string `json:"description"`
		}{
			{Name: "PORTUNIX_HOME", Description: "Base directory for Portunix data"},
			{Name: "PORTUNIX_CACHE", Description: "Cache directory for downloads"},
			{Name: "PORTUNIX_LOG_LEVEL", Description: "Logging level (debug, info, warn, error)"},
		},
	}

	jsonBytes, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}

// GetCommandInfo returns detailed information about a specific command
func GetCommandInfo(commandName string) *CommandInfo {
	for _, cmd := range CommandRegistry {
		if cmd.Name == commandName {
			return &cmd
		}
	}
	return nil
}