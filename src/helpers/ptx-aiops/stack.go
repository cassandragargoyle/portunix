package main

import (
	"fmt"
	"strings"
	"time"
)

// handleStackCreate creates full AI stack (Ollama + WebUI)
func handleStackCreate(args []string) {
	// Parse flags
	models := []string{}
	cpuOnly := false

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--models":
			if i+1 < len(args) {
				models = strings.Split(args[i+1], ",")
				i++
			}
		case "--cpu":
			cpuOnly = true
		}
	}

	fmt.Println()
	fmt.Println("Creating AI Stack (Ollama + Open WebUI)")
	fmt.Println(strings.Repeat("━", 60))
	fmt.Println()

	// Step 1: Check prerequisites
	fmt.Println("📋 Step 1: Checking prerequisites...")

	runtime := detectContainerRuntime()
	if runtime == "" {
		fmt.Println("❌ No container runtime available")
		fmt.Println("Install Docker or Podman first:")
		fmt.Println("  portunix install docker")
		return
	}
	fmt.Printf("  ✓ Container runtime: %s\n", runtime)

	// Check GPU if not CPU-only
	gpuAvailable := false
	if !cpuOnly {
		gpus, err := detectNvidiaGPUs()
		if err == nil && len(gpus) > 0 {
			toolkit := checkContainerToolkit()
			if toolkit.Installed {
				gpuAvailable = true
				fmt.Printf("  ✓ GPU: %s (toolkit installed)\n", gpus[0].Name)
			} else {
				fmt.Printf("  ⚠ GPU detected but toolkit not installed\n")
				fmt.Println("    Install with: portunix install nvidia-container-toolkit")
			}
		} else {
			fmt.Println("  ℹ GPU: Not available (using CPU mode)")
		}
	} else {
		fmt.Println("  ℹ GPU: Disabled (CPU-only mode)")
	}

	// Step 2: Create Ollama container
	fmt.Println()
	fmt.Println("📦 Step 2: Creating Ollama container...")

	ollamaStatus := getOllamaContainerStatus()
	if ollamaStatus.Exists {
		if ollamaStatus.Running {
			fmt.Printf("  ✓ Ollama container already exists and running\n")
		} else {
			fmt.Printf("  ⚠ Ollama container exists but stopped, starting...\n")
			handleOllamaContainerStart()
		}
	} else {
		// Create Ollama container
		createArgs := []string{}
		if cpuOnly || !gpuAvailable {
			createArgs = append(createArgs, "--cpu")
		}
		handleOllamaContainerCreate(createArgs)
	}

	// Wait a bit for Ollama to be ready
	time.Sleep(2 * time.Second)

	// Step 3: Install models if specified
	if len(models) > 0 {
		fmt.Println()
		fmt.Println("📥 Step 3: Installing models...")

		for _, model := range models {
			model = strings.TrimSpace(model)
			if model == "" {
				continue
			}
			fmt.Printf("  Installing %s...\n", model)
			handleModelInstall([]string{model})
		}
	}

	// Step 4: Create WebUI container
	fmt.Println()
	fmt.Println("🌐 Step 4: Creating Open WebUI container...")

	webuiStatus := getWebUIContainerStatus()
	if webuiStatus.Exists {
		if webuiStatus.Running {
			fmt.Printf("  ✓ WebUI container already exists and running\n")
		} else {
			fmt.Printf("  ⚠ WebUI container exists but stopped, starting...\n")
			handleWebUIContainerStart()
		}
	} else {
		handleWebUIContainerCreate([]string{})
	}

	// Summary
	fmt.Println()
	fmt.Println(strings.Repeat("━", 60))
	fmt.Println("✅ AI Stack created successfully!")
	fmt.Println()
	fmt.Println("Access points:")
	fmt.Printf("  🌐 Open WebUI:  %s\n", WebUIEndpoint)
	fmt.Printf("  🤖 Ollama API:  %s\n", OllamaAPIEndpoint)
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  portunix aiops webui open           - Open WebUI in browser")
	fmt.Println("  portunix aiops model list           - List installed models")
	fmt.Println("  portunix aiops stack status         - Check stack status")
	fmt.Println("  portunix aiops stack stop           - Stop all containers")

	if len(models) == 0 {
		fmt.Println()
		fmt.Println("💡 No models installed. Install one with:")
		fmt.Println("  portunix aiops model install llama3.2")
	}
}

// handleStackStatus shows status of all stack containers
func handleStackStatus() {
	fmt.Println()
	fmt.Println("AI Stack Status")
	fmt.Println(strings.Repeat("━", 60))
	fmt.Println()

	// Ollama status
	ollamaStatus := getOllamaContainerStatus()
	fmt.Print("Ollama Container:   ")
	if !ollamaStatus.Exists {
		fmt.Println("❌ Not created")
	} else if ollamaStatus.Running {
		fmt.Println("🟢 Running")
		fmt.Printf("  └─ API: %s\n", OllamaAPIEndpoint)
	} else {
		fmt.Println("🔴 Stopped")
	}

	// WebUI status
	webuiStatus := getWebUIContainerStatus()
	fmt.Print("Open WebUI:         ")
	if !webuiStatus.Exists {
		fmt.Println("❌ Not created")
	} else if webuiStatus.Running {
		fmt.Println("🟢 Running")
		fmt.Printf("  └─ Web: %s\n", WebUIEndpoint)
	} else {
		fmt.Println("🔴 Stopped")
	}

	// GPU status
	fmt.Println()
	fmt.Print("GPU Acceleration:   ")
	gpus, err := detectNvidiaGPUs()
	if err != nil || len(gpus) == 0 {
		fmt.Println("❌ Not available (CPU mode)")
	} else {
		if ollamaStatus.GPUEnabled {
			fmt.Printf("✓ Enabled (%s)\n", gpus[0].Name)
		} else {
			fmt.Println("⚠ Available but not enabled")
		}
	}

	fmt.Println()
	fmt.Println(strings.Repeat("━", 60))

	// Show available commands based on state
	if !ollamaStatus.Exists || !webuiStatus.Exists {
		fmt.Println("Create stack with: portunix aiops stack create")
	} else if !ollamaStatus.Running || !webuiStatus.Running {
		fmt.Println("Start stack with: portunix aiops stack start")
	} else {
		fmt.Println("Stack is ready! Commands:")
		fmt.Println("  portunix aiops webui open   - Open in browser")
		fmt.Println("  portunix aiops model list   - List models")
		fmt.Println("  portunix aiops stack stop   - Stop stack")
	}
}

// handleStackStart starts all stack containers
func handleStackStart() {
	fmt.Println()
	fmt.Println("Starting AI Stack...")
	fmt.Println()

	started := 0

	// Start Ollama
	ollamaStatus := getOllamaContainerStatus()
	if ollamaStatus.Exists {
		if !ollamaStatus.Running {
			fmt.Print("Starting Ollama... ")
			handleOllamaContainerStart()
			started++
		} else {
			fmt.Println("Ollama already running")
		}
	} else {
		fmt.Println("⚠ Ollama container not found. Create with: portunix aiops stack create")
	}

	// Start WebUI
	webuiStatus := getWebUIContainerStatus()
	if webuiStatus.Exists {
		if !webuiStatus.Running {
			fmt.Print("Starting Open WebUI... ")
			handleWebUIContainerStart()
			started++
		} else {
			fmt.Println("Open WebUI already running")
		}
	} else {
		fmt.Println("⚠ WebUI container not found. Create with: portunix aiops stack create")
	}

	fmt.Println()
	if started > 0 {
		fmt.Println("✓ Stack started")
		fmt.Printf("  🌐 Open WebUI:  %s\n", WebUIEndpoint)
		fmt.Printf("  🤖 Ollama API:  %s\n", OllamaAPIEndpoint)
	}
}

// handleStackStop stops all stack containers
func handleStackStop() {
	fmt.Println()
	fmt.Println("Stopping AI Stack...")
	fmt.Println()

	stopped := 0

	// Stop WebUI first (depends on Ollama)
	webuiStatus := getWebUIContainerStatus()
	if webuiStatus.Running {
		fmt.Print("Stopping Open WebUI... ")
		handleWebUIContainerStop()
		stopped++
	}

	// Stop Ollama
	ollamaStatus := getOllamaContainerStatus()
	if ollamaStatus.Running {
		fmt.Print("Stopping Ollama... ")
		handleOllamaContainerStop()
		stopped++
	}

	fmt.Println()
	if stopped > 0 {
		fmt.Println("✓ Stack stopped")
	} else {
		fmt.Println("No containers were running")
	}
}

// handleStackRemove removes all stack containers
func handleStackRemove() {
	fmt.Println()
	fmt.Println("Removing AI Stack...")
	fmt.Println()

	// Remove WebUI first
	webuiStatus := getWebUIContainerStatus()
	if webuiStatus.Exists {
		fmt.Println("Removing Open WebUI container...")
		handleWebUIContainerRemove()
	}

	// Remove Ollama
	ollamaStatus := getOllamaContainerStatus()
	if ollamaStatus.Exists {
		fmt.Println()
		fmt.Println("Removing Ollama container...")
		handleOllamaContainerRemove()
	}

	fmt.Println()
	fmt.Println("✓ AI Stack removed")
	fmt.Println()
	fmt.Println("Data directories preserved:")
	fmt.Printf("  Ollama models: %s\n", getOllamaDataDir())
	fmt.Printf("  WebUI data:    %s\n", getWebUIDataDir())
	fmt.Println()
	fmt.Println("To completely remove data, delete these directories manually.")
}
