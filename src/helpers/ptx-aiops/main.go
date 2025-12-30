package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var version = "dev"

// rootCmd represents the base command for ptx-aiops
var rootCmd = &cobra.Command{
	Use:   "ptx-aiops",
	Short: "Portunix AI Operations Helper",
	Long: `ptx-aiops is a helper binary for Portunix that handles AI/ML operations tooling.
It provides unified interface for GPU monitoring, Ollama container management,
Open WebUI deployment, and AI infrastructure operations.

This binary is typically invoked by the main portunix dispatcher and should not be used directly.

Supported features:
- GPU detection and monitoring (NVIDIA)
- Ollama container management with GPU support
- Open WebUI container deployment
- AI stack management (Ollama + WebUI)
- Model management for local LLM inference`,
	Version:            version,
	DisableFlagParsing: true,
	Run: func(cmd *cobra.Command, args []string) {
		handleCommand(args)
	},
}

func handleCommand(args []string) {
	if len(args) == 0 {
		fmt.Println("No command specified")
		return
	}

	command := args[0]
	subArgs := args[1:]

	switch command {
	case "aiops":
		if len(subArgs) == 0 {
			showAIOpsHelp()
		} else {
			handleAIOpsCommand(subArgs)
		}
	case "--version":
		fmt.Printf("ptx-aiops version %s\n", version)
	case "--description":
		fmt.Println("Portunix AI Operations Helper")
	case "--list-commands":
		fmt.Println("aiops")
	default:
		fmt.Printf("Unknown command: %s\n", command)
		fmt.Println("Supported commands: aiops")
	}
}

func showAIOpsHelp() {
	fmt.Println("Usage: portunix aiops [subcommand]")
	fmt.Println()
	fmt.Println("AI Operations Commands:")
	fmt.Println()
	fmt.Println("GPU Operations:")
	fmt.Println("  gpu status               - Show GPU status and driver info")
	fmt.Println("  gpu status --watch       - Real-time GPU monitoring (default: 5s refresh)")
	fmt.Println("  gpu status --watch --interval 2  - Custom refresh interval")
	fmt.Println("  gpu usage                - Show GPU utilization summary")
	fmt.Println("  gpu processes            - List processes using GPU")
	fmt.Println("  gpu check                - Verify GPU and container toolkit readiness")
	fmt.Println()
	fmt.Println("Ollama Container Operations:")
	fmt.Println("  ollama container create  - Create Ollama container (with GPU if available)")
	fmt.Println("  ollama container create --cpu  - Force CPU-only mode")
	fmt.Println("  ollama container status  - Show Ollama container status")
	fmt.Println("  ollama container start   - Start stopped Ollama container")
	fmt.Println("  ollama container stop    - Stop running Ollama container")
	fmt.Println("  ollama container remove  - Remove Ollama container")
	fmt.Println()
	fmt.Println("Model Operations:")
	fmt.Println("  model list               - List installed models in container")
	fmt.Println("  model list --available   - List available models from Ollama registry")
	fmt.Println("  model install <name>     - Install model to container")
	fmt.Println("  model info <name>        - Show model details")
	fmt.Println("  model remove <name>      - Remove model from container")
	fmt.Println("  model run <name>         - Interactive chat with model")
	fmt.Println()
	fmt.Println("Open WebUI Operations:")
	fmt.Println("  webui container create   - Create Open WebUI container")
	fmt.Println("  webui container status   - Show WebUI container status")
	fmt.Println("  webui container start    - Start stopped WebUI container")
	fmt.Println("  webui container stop     - Stop running WebUI container")
	fmt.Println("  webui container remove   - Remove WebUI container")
	fmt.Println("  webui open               - Open WebUI in browser")
	fmt.Println()
	fmt.Println("Stack Operations:")
	fmt.Println("  stack create             - Create full stack (Ollama + WebUI)")
	fmt.Println("  stack status             - Show all containers status")
	fmt.Println("  stack start              - Start all containers")
	fmt.Println("  stack stop               - Stop all containers")
	fmt.Println("  stack remove             - Remove all containers")
}

func handleAIOpsCommand(args []string) {
	if len(args) == 0 {
		showAIOpsHelp()
		return
	}

	subcommand := args[0]
	subArgs := args[1:]

	switch subcommand {
	case "gpu":
		handleGPUCommand(subArgs)
	case "ollama":
		handleOllamaCommand(subArgs)
	case "model":
		handleModelCommand(subArgs)
	case "webui":
		handleWebUICommand(subArgs)
	case "stack":
		handleStackCommand(subArgs)
	case "--help", "-h":
		showAIOpsHelp()
	default:
		fmt.Printf("Unknown aiops subcommand: %s\n", subcommand)
		fmt.Println("Run 'portunix aiops --help' for available commands")
	}
}

// GPU command handlers
func handleGPUCommand(args []string) {
	if len(args) == 0 {
		showGPUHelp()
		return
	}

	subcommand := args[0]
	subArgs := args[1:]

	switch subcommand {
	case "status":
		handleGPUStatus(subArgs)
	case "usage":
		handleGPUUsage()
	case "processes":
		handleGPUProcesses()
	case "check":
		handleGPUCheck()
	case "--help", "-h":
		showGPUHelp()
	default:
		fmt.Printf("Unknown gpu subcommand: %s\n", subcommand)
		fmt.Println("Run 'portunix aiops gpu --help' for available commands")
	}
}

func showGPUHelp() {
	fmt.Println("Usage: portunix aiops gpu [subcommand]")
	fmt.Println()
	fmt.Println("GPU Operations:")
	fmt.Println("  status               - Show GPU status and driver info")
	fmt.Println("  status --watch       - Real-time GPU monitoring (default: 5s refresh)")
	fmt.Println("  status --watch --interval <sec> - Custom refresh interval")
	fmt.Println("  usage                - Show GPU utilization summary")
	fmt.Println("  processes            - List processes using GPU")
	fmt.Println("  check                - Verify GPU and container toolkit readiness")
}

// Ollama command handlers
func handleOllamaCommand(args []string) {
	if len(args) == 0 {
		showOllamaHelp()
		return
	}

	subcommand := args[0]
	subArgs := args[1:]

	switch subcommand {
	case "container":
		handleOllamaContainerCommand(subArgs)
	case "--help", "-h":
		showOllamaHelp()
	default:
		fmt.Printf("Unknown ollama subcommand: %s\n", subcommand)
		fmt.Println("Run 'portunix aiops ollama --help' for available commands")
	}
}

func showOllamaHelp() {
	fmt.Println("Usage: portunix aiops ollama [subcommand]")
	fmt.Println()
	fmt.Println("Ollama Container Operations:")
	fmt.Println("  container create       - Create Ollama container (with GPU if available)")
	fmt.Println("  container create --cpu - Force CPU-only mode")
	fmt.Println("  container status       - Show Ollama container status")
	fmt.Println("  container start        - Start stopped Ollama container")
	fmt.Println("  container stop         - Stop running Ollama container")
	fmt.Println("  container remove       - Remove Ollama container")
}

func handleOllamaContainerCommand(args []string) {
	if len(args) == 0 {
		fmt.Println("Ollama container subcommand required")
		fmt.Println("Available: create, status, start, stop, remove")
		return
	}

	subcommand := args[0]
	subArgs := args[1:]

	switch subcommand {
	case "create":
		handleOllamaContainerCreate(subArgs)
	case "status":
		handleOllamaContainerStatus()
	case "start":
		handleOllamaContainerStart()
	case "stop":
		handleOllamaContainerStop()
	case "remove":
		handleOllamaContainerRemove()
	default:
		fmt.Printf("Unknown ollama container subcommand: %s\n", subcommand)
	}
}

// Model command handlers
func handleModelCommand(args []string) {
	if len(args) == 0 {
		showModelHelp()
		return
	}

	subcommand := args[0]
	subArgs := args[1:]

	switch subcommand {
	case "list", "ls":
		handleModelList(subArgs)
	case "install":
		handleModelInstall(subArgs)
	case "info":
		handleModelInfo(subArgs)
	case "remove", "rm":
		handleModelRemove(subArgs)
	case "run":
		handleModelRun(subArgs)
	case "--help", "-h":
		showModelHelp()
	default:
		fmt.Printf("Unknown model subcommand: %s\n", subcommand)
		fmt.Println("Run 'portunix aiops model --help' for available commands")
	}
}

func showModelHelp() {
	fmt.Println("Usage: portunix aiops model [subcommand]")
	fmt.Println()
	fmt.Println("Model Operations:")
	fmt.Println("  list                     - List installed models in container")
	fmt.Println("  list --container <name>  - List models in specific container")
	fmt.Println("  list --available         - List available models from Ollama registry")
	fmt.Println("  install <name>           - Install model to default container")
	fmt.Println("  install <name> --container <name> - Install to specific container")
	fmt.Println("  info <name>              - Show model details")
	fmt.Println("  remove <name>            - Remove model from container")
	fmt.Println("  run <name>               - Interactive chat with model")
	fmt.Println("  run <name> --prompt \"...\" - Single prompt execution")
}

// WebUI command handlers
func handleWebUICommand(args []string) {
	if len(args) == 0 {
		showWebUIHelp()
		return
	}

	subcommand := args[0]
	subArgs := args[1:]

	switch subcommand {
	case "container":
		handleWebUIContainerCommand(subArgs)
	case "open":
		handleWebUIOpen()
	case "--help", "-h":
		showWebUIHelp()
	default:
		fmt.Printf("Unknown webui subcommand: %s\n", subcommand)
		fmt.Println("Run 'portunix aiops webui --help' for available commands")
	}
}

func showWebUIHelp() {
	fmt.Println("Usage: portunix aiops webui [subcommand]")
	fmt.Println()
	fmt.Println("Open WebUI Operations:")
	fmt.Println("  container create   - Create Open WebUI container")
	fmt.Println("  container status   - Show WebUI container status")
	fmt.Println("  container start    - Start stopped WebUI container")
	fmt.Println("  container stop     - Stop running WebUI container")
	fmt.Println("  container remove   - Remove WebUI container")
	fmt.Println("  open               - Open WebUI in browser")
}

func handleWebUIContainerCommand(args []string) {
	if len(args) == 0 {
		fmt.Println("WebUI container subcommand required")
		fmt.Println("Available: create, status, start, stop, remove")
		return
	}

	subcommand := args[0]
	subArgs := args[1:]

	switch subcommand {
	case "create":
		handleWebUIContainerCreate(subArgs)
	case "status":
		handleWebUIContainerStatus()
	case "start":
		handleWebUIContainerStart()
	case "stop":
		handleWebUIContainerStop()
	case "remove":
		handleWebUIContainerRemove()
	default:
		fmt.Printf("Unknown webui container subcommand: %s\n", subcommand)
	}
}

// Stack command handlers
func handleStackCommand(args []string) {
	if len(args) == 0 {
		showStackHelp()
		return
	}

	subcommand := args[0]
	subArgs := args[1:]

	switch subcommand {
	case "create":
		handleStackCreate(subArgs)
	case "status":
		handleStackStatus()
	case "start":
		handleStackStart()
	case "stop":
		handleStackStop()
	case "remove":
		handleStackRemove()
	case "--help", "-h":
		showStackHelp()
	default:
		fmt.Printf("Unknown stack subcommand: %s\n", subcommand)
		fmt.Println("Run 'portunix aiops stack --help' for available commands")
	}
}

func showStackHelp() {
	fmt.Println("Usage: portunix aiops stack [subcommand]")
	fmt.Println()
	fmt.Println("Stack Operations (All Containers):")
	fmt.Println("  create                   - Create full stack (Ollama + WebUI)")
	fmt.Println("  create --models <list>   - Create with pre-installed models")
	fmt.Println("  status                   - Show all containers status")
	fmt.Println("  start                    - Start all containers")
	fmt.Println("  stop                     - Stop all containers")
	fmt.Println("  remove                   - Remove all containers")
}

func init() {
	rootCmd.Flags().BoolP("version", "v", false, "Show version")
	rootCmd.Flags().Bool("description", false, "Show description")
	rootCmd.Flags().Bool("list-commands", false, "List available commands")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
