package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	// Default container settings
	OllamaContainerName = "portunix-ollama"
	OllamaImage         = "ollama/ollama:latest"
	OllamaPort          = "11434"
	OllamaAPIEndpoint   = "http://localhost:11434"
)

// OllamaContainerStatus represents the status of Ollama container
type OllamaContainerStatus struct {
	Exists     bool
	Running    bool
	Name       string
	Image      string
	GPUEnabled bool
	Port       string
	Created    string
}

// getOllamaDataDir returns the directory for Ollama model storage
func getOllamaDataDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "/tmp/portunix/aiops/ollama"
	}
	return filepath.Join(homeDir, ".portunix", "aiops", "ollama", "models")
}

// ensureOllamaDataDir creates Ollama data directory if it doesn't exist
func ensureOllamaDataDir() error {
	dataDir := getOllamaDataDir()
	return os.MkdirAll(dataDir, 0755)
}

// detectContainerRuntime returns the available container runtime
func detectContainerRuntime() string {
	if isDockerAvailable() {
		return "docker"
	}
	if isPodmanAvailable() {
		return "podman"
	}
	return ""
}

// getOllamaContainerStatus checks the status of Ollama container
func getOllamaContainerStatus() OllamaContainerStatus {
	status := OllamaContainerStatus{
		Name: OllamaContainerName,
	}

	runtime := detectContainerRuntime()
	if runtime == "" {
		return status
	}

	// Check if container exists
	cmd := exec.Command(runtime, "inspect", OllamaContainerName, "--format", "{{.State.Running}}")
	output, err := cmd.Output()
	if err != nil {
		// Container doesn't exist
		return status
	}

	status.Exists = true
	status.Running = strings.TrimSpace(string(output)) == "true"

	// Get more details
	cmd = exec.Command(runtime, "inspect", OllamaContainerName,
		"--format", "{{.Config.Image}}|{{.Created}}")
	output, err = cmd.Output()
	if err == nil {
		parts := strings.Split(strings.TrimSpace(string(output)), "|")
		if len(parts) >= 2 {
			status.Image = parts[0]
			status.Created = parts[1]
		}
	}

	// Check if GPU is enabled (look for GPU device in config)
	cmd = exec.Command(runtime, "inspect", OllamaContainerName,
		"--format", "{{.HostConfig.DeviceRequests}}")
	output, err = cmd.Output()
	if err == nil {
		status.GPUEnabled = strings.Contains(string(output), "gpu")
	}

	status.Port = OllamaPort
	return status
}

// handleOllamaContainerCreate creates a new Ollama container
func handleOllamaContainerCreate(args []string) {
	// Parse flags
	cpuOnly := false
	for _, arg := range args {
		if arg == "--cpu" {
			cpuOnly = true
		}
	}

	runtime := detectContainerRuntime()
	if runtime == "" {
		fmt.Println("❌ No container runtime available")
		fmt.Println("Install Docker or Podman first:")
		fmt.Println("  portunix install docker")
		fmt.Println("  portunix install podman")
		return
	}

	// Check if container already exists
	status := getOllamaContainerStatus()
	if status.Exists {
		fmt.Printf("❌ Container '%s' already exists\n", OllamaContainerName)
		fmt.Println()
		if status.Running {
			fmt.Println("Container is running. To recreate:")
		} else {
			fmt.Println("Container is stopped. To recreate:")
		}
		fmt.Printf("  portunix aiops ollama container remove\n")
		fmt.Printf("  portunix aiops ollama container create\n")
		return
	}

	// Ensure data directory exists
	if err := ensureOllamaDataDir(); err != nil {
		fmt.Printf("❌ Failed to create data directory: %v\n", err)
		return
	}

	dataDir := getOllamaDataDir()

	fmt.Println("Creating Ollama container...")
	fmt.Printf("  Name:   %s\n", OllamaContainerName)
	fmt.Printf("  Image:  %s\n", OllamaImage)
	fmt.Printf("  Port:   %s\n", OllamaPort)
	fmt.Printf("  Data:   %s\n", dataDir)

	// Build container command
	cmdArgs := []string{
		"run", "-d",
		"--name", OllamaContainerName,
		"-p", OllamaPort + ":" + OllamaPort,
		"-v", dataDir + ":/root/.ollama",
	}

	// Add GPU support if available and not CPU-only mode
	gpuEnabled := false
	if !cpuOnly {
		gpus, err := detectNvidiaGPUs()
		if err == nil && len(gpus) > 0 {
			toolkit := checkContainerToolkit()
			if toolkit.Installed {
				if runtime == "docker" {
					cmdArgs = append(cmdArgs, "--gpus", "all")
				} else {
					cmdArgs = append(cmdArgs, "--device", "nvidia.com/gpu=all")
				}
				gpuEnabled = true
				fmt.Println("  GPU:    ✓ Enabled")
			} else {
				fmt.Println("  GPU:    ⚠ Available but toolkit not installed")
				fmt.Println("          Run: portunix install nvidia-container-toolkit")
			}
		} else {
			fmt.Println("  GPU:    ✗ Not available (CPU mode)")
		}
	} else {
		fmt.Println("  GPU:    ✗ Disabled (CPU-only mode)")
	}

	// Add restart policy
	cmdArgs = append(cmdArgs, "--restart", "unless-stopped")

	// Add image
	cmdArgs = append(cmdArgs, OllamaImage)

	fmt.Println()
	fmt.Printf("Running: %s %s\n", runtime, strings.Join(cmdArgs, " "))

	cmd := exec.Command(runtime, cmdArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("❌ Failed to create container: %v\n", err)
		fmt.Printf("Output: %s\n", string(output))
		return
	}

	fmt.Println()
	fmt.Printf("✓ Container '%s' created successfully\n", OllamaContainerName)
	fmt.Println()

	// Wait for container to be ready
	fmt.Print("Waiting for Ollama to start...")
	ready := false
	for i := 0; i < 30; i++ {
		time.Sleep(time.Second)
		fmt.Print(".")

		resp, err := http.Get(OllamaAPIEndpoint + "/api/tags")
		if err == nil {
			resp.Body.Close()
			ready = true
			break
		}
	}

	fmt.Println()
	if ready {
		fmt.Println("✓ Ollama is ready!")
		fmt.Println()
		fmt.Println("Next steps:")
		fmt.Println("  portunix aiops model install llama3.2      - Install a model")
		fmt.Println("  portunix aiops model list                  - List installed models")
		fmt.Println("  portunix aiops model run llama3.2          - Chat with model")
		if gpuEnabled {
			fmt.Println()
			fmt.Println("GPU acceleration is enabled for faster inference.")
		}
	} else {
		fmt.Println("⚠ Ollama is taking longer than expected to start")
		fmt.Println("Check logs with: portunix container logs", OllamaContainerName)
	}
}

// handleOllamaContainerStatus shows Ollama container status
func handleOllamaContainerStatus() {
	status := getOllamaContainerStatus()

	fmt.Println()
	fmt.Println("Ollama Container Status")
	fmt.Println(strings.Repeat("━", 50))
	fmt.Println()

	if !status.Exists {
		fmt.Printf("Container '%s': Not found\n", OllamaContainerName)
		fmt.Println()
		fmt.Println("Create container with:")
		fmt.Println("  portunix aiops ollama container create")
		return
	}

	stateEmoji := "🔴"
	stateText := "Stopped"
	if status.Running {
		stateEmoji = "🟢"
		stateText = "Running"
	}

	fmt.Printf("Container:   %s\n", status.Name)
	fmt.Printf("Status:      %s %s\n", stateEmoji, stateText)
	fmt.Printf("Image:       %s\n", status.Image)

	gpuStatus := "Disabled"
	if status.GPUEnabled {
		gpuStatus = "Enabled"
	}
	fmt.Printf("GPU:         %s\n", gpuStatus)
	fmt.Printf("API Port:    %s\n", status.Port)

	if status.Created != "" {
		fmt.Printf("Created:     %s\n", status.Created[:19])
	}

	// Check API availability if running
	if status.Running {
		fmt.Println()
		fmt.Print("API Status:  ")
		resp, err := http.Get(OllamaAPIEndpoint + "/api/tags")
		if err == nil {
			resp.Body.Close()
			fmt.Println("✓ Available at", OllamaAPIEndpoint)
		} else {
			fmt.Println("⚠ Not responding")
		}
	}

	fmt.Println()
	fmt.Println(strings.Repeat("━", 50))

	if status.Running {
		fmt.Println("Commands:")
		fmt.Println("  portunix aiops model list           - List installed models")
		fmt.Println("  portunix aiops model install <name> - Install a model")
		fmt.Println("  portunix aiops ollama container stop - Stop container")
	} else {
		fmt.Println("Start container with:")
		fmt.Println("  portunix aiops ollama container start")
	}
}

// handleOllamaContainerStart starts the Ollama container
func handleOllamaContainerStart() {
	status := getOllamaContainerStatus()

	if !status.Exists {
		fmt.Printf("❌ Container '%s' does not exist\n", OllamaContainerName)
		fmt.Println("Create it with: portunix aiops ollama container create")
		return
	}

	if status.Running {
		fmt.Printf("Container '%s' is already running\n", OllamaContainerName)
		return
	}

	runtime := detectContainerRuntime()
	cmd := exec.Command(runtime, "start", OllamaContainerName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("❌ Failed to start container: %v\n", err)
		fmt.Printf("Output: %s\n", string(output))
		return
	}

	fmt.Printf("✓ Container '%s' started\n", OllamaContainerName)

	// Wait for API to be ready
	fmt.Print("Waiting for Ollama API...")
	for i := 0; i < 15; i++ {
		time.Sleep(time.Second)
		fmt.Print(".")
		resp, err := http.Get(OllamaAPIEndpoint + "/api/tags")
		if err == nil {
			resp.Body.Close()
			fmt.Println(" Ready!")
			return
		}
	}
	fmt.Println(" ⚠ API not responding yet")
}

// handleOllamaContainerStop stops the Ollama container
func handleOllamaContainerStop() {
	status := getOllamaContainerStatus()

	if !status.Exists {
		fmt.Printf("❌ Container '%s' does not exist\n", OllamaContainerName)
		return
	}

	if !status.Running {
		fmt.Printf("Container '%s' is already stopped\n", OllamaContainerName)
		return
	}

	runtime := detectContainerRuntime()
	cmd := exec.Command(runtime, "stop", OllamaContainerName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("❌ Failed to stop container: %v\n", err)
		fmt.Printf("Output: %s\n", string(output))
		return
	}

	fmt.Printf("✓ Container '%s' stopped\n", OllamaContainerName)
}

// handleOllamaContainerRemove removes the Ollama container
func handleOllamaContainerRemove() {
	status := getOllamaContainerStatus()

	if !status.Exists {
		fmt.Printf("Container '%s' does not exist\n", OllamaContainerName)
		return
	}

	runtime := detectContainerRuntime()

	// Stop first if running
	if status.Running {
		fmt.Printf("Stopping container '%s'...\n", OllamaContainerName)
		cmd := exec.Command(runtime, "stop", OllamaContainerName)
		cmd.Run()
	}

	// Remove container
	fmt.Printf("Removing container '%s'...\n", OllamaContainerName)
	cmd := exec.Command(runtime, "rm", OllamaContainerName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("❌ Failed to remove container: %v\n", err)
		fmt.Printf("Output: %s\n", string(output))
		return
	}

	fmt.Printf("✓ Container '%s' removed\n", OllamaContainerName)
	fmt.Println()
	fmt.Printf("Note: Model data preserved in %s\n", getOllamaDataDir())
	fmt.Println("To remove model data, delete this directory manually.")
}

// OllamaModel represents model info from Ollama API
type OllamaModel struct {
	Name       string    `json:"name"`
	Model      string    `json:"model"`
	ModifiedAt time.Time `json:"modified_at"`
	Size       int64     `json:"size"`
	Digest     string    `json:"digest"`
	Details    struct {
		ParentModel       string   `json:"parent_model"`
		Format            string   `json:"format"`
		Family            string   `json:"family"`
		Families          []string `json:"families"`
		ParameterSize     string   `json:"parameter_size"`
		QuantizationLevel string   `json:"quantization_level"`
	} `json:"details"`
}

// OllamaTagsResponse represents response from /api/tags
type OllamaTagsResponse struct {
	Models []OllamaModel `json:"models"`
}

// getContainerStatus checks status of a specific container
func getContainerStatus(containerName string) (exists bool, running bool) {
	runtime := detectContainerRuntime()
	if runtime == "" {
		return false, false
	}

	cmd := exec.Command(runtime, "inspect", containerName, "--format", "{{.State.Running}}")
	output, err := cmd.Output()
	if err != nil {
		return false, false
	}

	return true, strings.TrimSpace(string(output)) == "true"
}

// handleModelList lists installed models
func handleModelList(args []string) {
	// Parse flags
	showAvailable := false
	containerName := OllamaContainerName
	customContainer := false

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--available":
			showAvailable = true
		case "--container":
			if i+1 < len(args) {
				containerName = args[i+1]
				customContainer = true
				i++
			}
		}
	}

	if showAvailable {
		showAvailableModels()
		return
	}

	// Display container being used
	if customContainer {
		fmt.Printf("Using container: %s\n", containerName)
	} else {
		fmt.Printf("Using container: %s (default)\n", containerName)
	}
	fmt.Println()

	// Check if container exists and is running
	exists, running := getContainerStatus(containerName)
	if !exists {
		fmt.Printf("❌ Container '%s' not found\n", containerName)
		if customContainer {
			fmt.Println("Available Ollama containers:")
			listOllamaContainers()
		} else {
			fmt.Println("Create it with: portunix aiops ollama container create")
		}
		return
	}

	if !running {
		fmt.Printf("❌ Container '%s' is not running\n", containerName)
		fmt.Printf("Start it with: portunix container start %s\n", containerName)
		return
	}

	// Get models from API - use container exec to query Ollama inside container
	runtime := detectContainerRuntime()
	cmd := exec.Command(runtime, "exec", containerName, "ollama", "list")
	output, err := cmd.Output()
	if err != nil {
		// Fallback to API if exec fails (might be different port mapping)
		resp, apiErr := http.Get(OllamaAPIEndpoint + "/api/tags")
		if apiErr != nil {
			fmt.Printf("❌ Failed to connect to Ollama: %v\n", err)
			return
		}
		defer resp.Body.Close()

		var tagsResp OllamaTagsResponse
		if decErr := json.NewDecoder(resp.Body).Decode(&tagsResp); decErr != nil {
			fmt.Printf("❌ Failed to parse response: %v\n", decErr)
			return
		}

		displayModelsFromAPI(tagsResp.Models)
		return
	}

	// Parse ollama list output directly
	displayModelsFromExec(string(output))
}

// displayModelsFromExec parses and displays output from ollama list command
func displayModelsFromExec(output string) {
	lines := strings.Split(strings.TrimSpace(output), "\n")

	if len(lines) <= 1 {
		fmt.Println("No models installed.")
		fmt.Println()
		fmt.Println("Install a model with:")
		fmt.Println("  portunix aiops model install llama3.2")
		fmt.Println("  portunix aiops model list --available")
		return
	}

	fmt.Println("Installed Models:")
	fmt.Println("┌────────────────────────────┬────────────┬──────────────┬─────────────────────┐")
	fmt.Println("│ NAME                       │ SIZE       │ QUANTIZATION │ MODIFIED            │")
	fmt.Println("├────────────────────────────┼────────────┼──────────────┼─────────────────────┤")

	modelCount := 0
	// Skip header line
	for i := 1; i < len(lines); i++ {
		line := lines[i]
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Parse space-separated fields (NAME ID SIZE MODIFIED)
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}

		name := fields[0]
		if len(name) > 26 {
			name = name[:23] + "..."
		}

		size := fields[2]

		// Extract quantization from name if present (e.g., llama3.2:3b-q4_0)
		quant := "-"
		if strings.Contains(fields[0], "q4") || strings.Contains(fields[0], "q5") ||
			strings.Contains(fields[0], "q8") || strings.Contains(fields[0], "fp16") {
			parts := strings.Split(fields[0], "-")
			if len(parts) > 1 {
				quant = parts[len(parts)-1]
			}
		}

		// Modified time (combine remaining fields)
		modified := strings.Join(fields[3:], " ")
		if len(modified) > 19 {
			modified = modified[:19]
		}

		fmt.Printf("│ %-26s │ %10s │ %-12s │ %-19s │\n", name, size, quant, modified)
		modelCount++
	}

	fmt.Println("└────────────────────────────┴────────────┴──────────────┴─────────────────────┘")
	fmt.Printf("\nTotal: %d models\n", modelCount)
}

// displayModelsFromAPI displays models from API response
func displayModelsFromAPI(models []OllamaModel) {
	if len(models) == 0 {
		fmt.Println("No models installed.")
		fmt.Println()
		fmt.Println("Install a model with:")
		fmt.Println("  portunix aiops model install llama3.2")
		fmt.Println("  portunix aiops model list --available")
		return
	}

	fmt.Println("Installed Models:")
	fmt.Println("┌────────────────────────────┬────────────┬─────────────────┬──────────────┬─────────────────────┐")
	fmt.Println("│ NAME                       │ SIZE       │ PARAMETERS      │ QUANTIZATION │ MODIFIED            │")
	fmt.Println("├────────────────────────────┼────────────┼─────────────────┼──────────────┼─────────────────────┤")

	totalSize := int64(0)
	for _, model := range models {
		name := model.Name
		if len(name) > 26 {
			name = name[:23] + "..."
		}

		paramSize := model.Details.ParameterSize
		if paramSize == "" {
			paramSize = "-"
		}

		quant := model.Details.QuantizationLevel
		if quant == "" {
			quant = "-"
		}

		modified := model.ModifiedAt.Format("2006-01-02 15:04:05")

		fmt.Printf("│ %-26s │ %10s │ %-15s │ %-12s │ %-19s │\n",
			name, formatBytes(model.Size), paramSize, quant, modified)

		totalSize += model.Size
	}

	fmt.Println("└────────────────────────────┴────────────┴─────────────────┴──────────────┴─────────────────────┘")
	fmt.Printf("\nTotal: %d models (%s)\n", len(models), formatBytes(totalSize))
}

// listOllamaContainers lists available Ollama containers
func listOllamaContainers() {
	runtime := detectContainerRuntime()
	if runtime == "" {
		fmt.Println("  No container runtime available")
		return
	}

	// List containers with ollama image
	cmd := exec.Command(runtime, "ps", "-a", "--filter", "ancestor=ollama/ollama", "--format", "{{.Names}}")
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("  Unable to list containers")
		return
	}

	containers := strings.TrimSpace(string(output))
	if containers == "" {
		fmt.Println("  No Ollama containers found")
		fmt.Println("  Create one with: portunix aiops ollama container create")
		return
	}

	for _, name := range strings.Split(containers, "\n") {
		if name != "" {
			fmt.Printf("  • %s\n", name)
		}
	}
}

// ModelRecommendation holds model info with hardware requirements
type ModelRecommendation struct {
	Name        string
	Description string
	Sizes       string
	MinVRAM     int64 // Minimum VRAM in GB for smallest variant
	MaxVRAM     int64 // VRAM needed for largest variant
}

// getModelRecommendations returns list of models with their requirements
func getModelRecommendations() []ModelRecommendation {
	return []ModelRecommendation{
		{"llama3.2", "Meta's latest lightweight model", "1b, 3b", 2, 4},
		{"llama3.1", "General purpose, high quality", "8b, 70b, 405b", 6, 256},
		{"mistral", "Fast and efficient 7B model", "7b", 6, 6},
		{"mixtral", "Mixture of experts architecture", "8x7b, 8x22b", 24, 128},
		{"codellama", "Specialized for code generation", "7b, 13b, 34b, 70b", 6, 48},
		{"phi3", "Microsoft's efficient model", "mini, medium", 3, 8},
		{"gemma2", "Google's open model", "2b, 9b, 27b", 2, 20},
		{"qwen2.5", "Alibaba's multilingual model", "0.5b - 72b", 1, 48},
		{"deepseek-coder-v2", "Advanced coding capabilities", "16b, 236b", 12, 160},
	}
}

// calculateModelRating calculates star rating (0-5) based on GPU capabilities
func calculateModelRating(model ModelRecommendation, gpuVRAM int64, hasGPU bool) string {
	if !hasGPU {
		// CPU only - rate based on model size (smaller = more stars)
		if model.MinVRAM <= 2 {
			return "★★★☆☆" // Small models work well on CPU
		} else if model.MinVRAM <= 4 {
			return "★★☆☆☆" // Medium models are slow on CPU
		} else if model.MinVRAM <= 8 {
			return "★☆☆☆☆" // Large models very slow on CPU
		}
		return "☆☆☆☆☆" // Very large models impractical on CPU
	}

	// GPU available - rate based on VRAM fit
	if gpuVRAM >= model.MaxVRAM {
		return "★★★★★" // Can run all variants
	} else if gpuVRAM >= model.MinVRAM*2 {
		return "★★★★☆" // Can run comfortably
	} else if gpuVRAM >= model.MinVRAM {
		return "★★★☆☆" // Can run smallest variant
	} else if gpuVRAM >= model.MinVRAM/2 {
		return "★★☆☆☆" // Might work with quantization
	} else if gpuVRAM > 0 {
		return "★☆☆☆☆" // Will need significant offloading
	}
	return "☆☆☆☆☆" // Not recommended
}

// showAvailableModels displays available models from Ollama registry with GPU compatibility
func showAvailableModels() {
	fmt.Println()

	// Detect GPU capabilities
	gpuVRAM := int64(0)
	hasGPU := false
	gpuName := "No GPU detected"

	gpus, err := detectNvidiaGPUs()
	if err == nil && len(gpus) > 0 {
		hasGPU = true
		gpuVRAM = gpus[0].MemoryTotal / (1024 * 1024 * 1024) // Convert to GB
		gpuName = fmt.Sprintf("%s (%d GB VRAM)", gpus[0].Name, gpuVRAM)
	}

	fmt.Printf("Available Ollama Models (Rating based on: %s)\n", gpuName)
	fmt.Println("┌──────────────────────┬─────────────────────────────────────────┬───────────────────┬───────────┐")
	fmt.Println("│ NAME                 │ DESCRIPTION                             │ SIZES             │ RATING    │")
	fmt.Println("├──────────────────────┼─────────────────────────────────────────┼───────────────────┼───────────┤")

	models := getModelRecommendations()
	for _, model := range models {
		rating := calculateModelRating(model, gpuVRAM, hasGPU)
		fmt.Printf("│ %-20s │ %-39s │ %-17s │ %s │\n",
			model.Name, model.Description, model.Sizes, rating)
	}

	fmt.Println("└──────────────────────┴─────────────────────────────────────────┴───────────────────┴───────────┘")
	fmt.Println()
	fmt.Println("Rating Legend:")
	fmt.Println("  ★★★★★ - Excellent: Can run all variants with great performance")
	fmt.Println("  ★★★★☆ - Very Good: Runs comfortably on your hardware")
	fmt.Println("  ★★★☆☆ - Good: Can run smallest variant")
	fmt.Println("  ★★☆☆☆ - Fair: May need quantization or be slow")
	fmt.Println("  ★☆☆☆☆ - Poor: Significant limitations expected")
	fmt.Println("  ☆☆☆☆☆ - Not Recommended: Hardware insufficient")
	fmt.Println()
	fmt.Println("Use 'portunix aiops model install <name>' to download a model.")
	fmt.Println("Example: portunix aiops model install llama3.2:3b")
}

// handleModelInstall installs a model to Ollama container
func handleModelInstall(args []string) {
	if len(args) == 0 {
		fmt.Println("❌ Model name required")
		fmt.Println("Usage: portunix aiops model install <model-name>")
		fmt.Println("Example: portunix aiops model install llama3.2")
		return
	}

	modelName := args[0]
	containerName := OllamaContainerName
	customContainer := false

	// Parse additional flags
	for i := 1; i < len(args); i++ {
		if args[i] == "--container" && i+1 < len(args) {
			containerName = args[i+1]
			customContainer = true
			i++
		}
	}

	// Display container being used
	if customContainer {
		fmt.Printf("Using container: %s\n", containerName)
	} else {
		fmt.Printf("Using container: %s (default)\n", containerName)
	}
	fmt.Println()

	// Check if container exists and is running
	exists, running := getContainerStatus(containerName)
	if !exists {
		fmt.Printf("❌ Container '%s' not found\n", containerName)
		if customContainer {
			fmt.Println("Available Ollama containers:")
			listOllamaContainers()
		} else {
			fmt.Println("Create it with: portunix aiops ollama container create")
		}
		return
	}

	if !running {
		fmt.Printf("❌ Container '%s' is not running\n", containerName)
		fmt.Printf("Start it with: portunix container start %s\n", containerName)
		return
	}

	fmt.Printf("Installing model: %s\n", modelName)
	fmt.Println("This may take a while depending on model size...")
	fmt.Println()

	// Execute ollama pull inside container
	runtime := detectContainerRuntime()
	cmd := exec.Command(runtime, "exec", containerName, "ollama", "pull", modelName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("\n❌ Failed to install model: %v\n", err)
		return
	}

	fmt.Println()
	fmt.Printf("✓ Model %s installed successfully\n", modelName)
	fmt.Println()
	fmt.Printf("Run model with: portunix aiops model run %s\n", modelName)
	if customContainer {
		fmt.Printf("             or: portunix aiops model run %s --container %s\n", modelName, containerName)
	}
}

// handleModelInfo shows model information
func handleModelInfo(args []string) {
	if len(args) == 0 {
		fmt.Println("❌ Model name required")
		fmt.Println("Usage: portunix aiops model info <model-name>")
		return
	}

	modelName := args[0]

	// Show info from embedded registry
	showModelInfoFromRegistry(modelName)
}

// ModelInfoEntry holds detailed model information
type ModelInfoEntry struct {
	DisplayName   string
	Provider      string
	License       string
	Variants      []string
	Description   string
	MinRAM        string
	RecommendedHW string
	UseCases      []string
}

// getModelRegistry returns embedded model registry
func getModelRegistry() map[string]ModelInfoEntry {
	return map[string]ModelInfoEntry{
		"llama3.2": {
			DisplayName:   "Llama 3.2",
			Provider:      "Meta (Ollama Library)",
			License:       "Llama 3.2 Community License",
			Variants:      []string{"llama3.2:1b - 1.3 GB", "llama3.2:3b - 2.0 GB"},
			Description:   "Meta's latest lightweight language model optimized for fast inference on consumer hardware.",
			MinRAM:        "4 GB (1b) / 6 GB (3b)",
			RecommendedHW: "8 GB RAM or 4 GB VRAM",
			UseCases:      []string{"General text generation", "Summarization", "Simple Q&A", "Basic coding assistance"},
		},
		"llama3.1": {
			DisplayName:   "Llama 3.1",
			Provider:      "Meta (Ollama Library)",
			License:       "Llama 3.1 Community License",
			Variants:      []string{"llama3.1:8b - 4.7 GB", "llama3.1:70b - 40 GB", "llama3.1:405b - 231 GB"},
			Description:   "Meta's flagship model with excellent reasoning and instruction following capabilities.",
			MinRAM:        "16 GB (8b) / 64 GB (70b) / 256+ GB (405b)",
			RecommendedHW: "16 GB RAM or 8 GB VRAM (8b)",
			UseCases:      []string{"Complex reasoning", "Long-form content", "Multi-turn conversations", "Advanced coding"},
		},
		"mistral": {
			DisplayName:   "Mistral 7B",
			Provider:      "Mistral AI (Ollama Library)",
			License:       "Apache 2.0",
			Variants:      []string{"mistral:7b - 4.1 GB"},
			Description:   "Fast and efficient 7B parameter model with excellent performance-to-size ratio.",
			MinRAM:        "8 GB",
			RecommendedHW: "16 GB RAM or 8 GB VRAM",
			UseCases:      []string{"General purpose tasks", "Code completion", "Instruction following"},
		},
		"mixtral": {
			DisplayName:   "Mixtral (MoE)",
			Provider:      "Mistral AI (Ollama Library)",
			License:       "Apache 2.0",
			Variants:      []string{"mixtral:8x7b - 26 GB", "mixtral:8x22b - 80 GB"},
			Description:   "Mixture of Experts model with 8 experts, providing excellent quality with sparse activation.",
			MinRAM:        "48 GB (8x7b) / 128 GB (8x22b)",
			RecommendedHW: "32 GB VRAM (8x7b)",
			UseCases:      []string{"Complex tasks", "Multilingual support", "High-quality generation"},
		},
		"codellama": {
			DisplayName:   "Code Llama",
			Provider:      "Meta (Ollama Library)",
			License:       "Llama 2 Community License",
			Variants:      []string{"codellama:7b - 3.8 GB", "codellama:13b - 7.4 GB", "codellama:34b - 19 GB"},
			Description:   "Specialized model for code generation, completion, and understanding.",
			MinRAM:        "8 GB (7b) / 16 GB (13b) / 32 GB (34b)",
			RecommendedHW: "16 GB RAM or 8 GB VRAM (7b)",
			UseCases:      []string{"Code generation", "Code completion", "Code review", "Bug fixing"},
		},
		"phi3": {
			DisplayName:   "Phi-3",
			Provider:      "Microsoft (Ollama Library)",
			License:       "MIT",
			Variants:      []string{"phi3:mini - 2.2 GB", "phi3:medium - 7.9 GB"},
			Description:   "Microsoft's efficient small language model with impressive reasoning capabilities.",
			MinRAM:        "4 GB (mini) / 16 GB (medium)",
			RecommendedHW: "8 GB RAM or 4 GB VRAM",
			UseCases:      []string{"Reasoning tasks", "Math problems", "Code assistance", "Edge deployment"},
		},
		"gemma2": {
			DisplayName:   "Gemma 2",
			Provider:      "Google (Ollama Library)",
			License:       "Gemma Terms of Use",
			Variants:      []string{"gemma2:2b - 1.6 GB", "gemma2:9b - 5.4 GB", "gemma2:27b - 16 GB"},
			Description:   "Google's open model family with strong performance across various tasks.",
			MinRAM:        "4 GB (2b) / 12 GB (9b) / 32 GB (27b)",
			RecommendedHW: "8 GB RAM or 4 GB VRAM (2b)",
			UseCases:      []string{"Text generation", "Question answering", "Summarization", "Creative writing"},
		},
		"qwen2.5": {
			DisplayName:   "Qwen 2.5",
			Provider:      "Alibaba (Ollama Library)",
			License:       "Qwen License",
			Variants:      []string{"qwen2.5:0.5b - 0.4 GB", "qwen2.5:1.5b - 1.0 GB", "qwen2.5:7b - 4.4 GB", "qwen2.5:72b - 47 GB"},
			Description:   "Alibaba's multilingual model with strong Chinese and English performance.",
			MinRAM:        "2 GB (0.5b) / 4 GB (1.5b) / 16 GB (7b) / 96 GB (72b)",
			RecommendedHW: "8 GB VRAM (7b)",
			UseCases:      []string{"Multilingual tasks", "Chinese NLP", "Code generation", "Math"},
		},
		"deepseek-coder-v2": {
			DisplayName:   "DeepSeek Coder V2",
			Provider:      "DeepSeek (Ollama Library)",
			License:       "DeepSeek License",
			Variants:      []string{"deepseek-coder-v2:16b - 8.9 GB", "deepseek-coder-v2:236b - 133 GB"},
			Description:   "Advanced coding model with strong performance across multiple programming languages.",
			MinRAM:        "24 GB (16b) / 256 GB (236b)",
			RecommendedHW: "16 GB VRAM (16b)",
			UseCases:      []string{"Code generation", "Code completion", "Code explanation", "Debugging"},
		},
		"llava": {
			DisplayName:   "LLaVA",
			Provider:      "LLaVA Team (Ollama Library)",
			License:       "Apache 2.0",
			Variants:      []string{"llava:7b - 4.5 GB", "llava:13b - 8.0 GB"},
			Description:   "Large Language and Vision Assistant - multimodal model for image understanding.",
			MinRAM:        "16 GB (7b) / 24 GB (13b)",
			RecommendedHW: "12 GB VRAM (7b)",
			UseCases:      []string{"Image description", "Visual Q&A", "Image analysis", "OCR"},
		},
		"starcoder2": {
			DisplayName:   "StarCoder 2",
			Provider:      "BigCode (Ollama Library)",
			License:       "BigCode OpenRAIL-M",
			Variants:      []string{"starcoder2:3b - 1.7 GB", "starcoder2:7b - 4.0 GB", "starcoder2:15b - 9.0 GB"},
			Description:   "Open-source code LLM trained on The Stack v2 dataset.",
			MinRAM:        "8 GB (3b) / 16 GB (7b) / 32 GB (15b)",
			RecommendedHW: "8 GB VRAM (7b)",
			UseCases:      []string{"Code generation", "Code completion", "Fill-in-the-middle", "Multi-language coding"},
		},
	}
}

// showModelInfoFromRegistry displays model info from embedded data
func showModelInfoFromRegistry(modelName string) {
	models := getModelRegistry()

	baseName := strings.Split(modelName, ":")[0]

	info, exists := models[baseName]
	if !exists {
		fmt.Printf("Model: %s\n", modelName)
		fmt.Println(strings.Repeat("━", 50))
		fmt.Println()
		fmt.Println("Detailed information not available in local registry.")
		fmt.Println("Visit: https://ollama.ai/library/" + baseName)
		return
	}

	fmt.Printf("\nModel: %s\n", baseName)
	fmt.Println(strings.Repeat("━", 50))
	fmt.Println()
	fmt.Printf("Display Name:    %s\n", info.DisplayName)
	fmt.Printf("Provider:        %s\n", info.Provider)
	fmt.Printf("License:         %s\n", info.License)
	fmt.Println()
	fmt.Println("Available Variants:")
	for _, v := range info.Variants {
		fmt.Printf("  • %s\n", v)
	}
	fmt.Println()
	fmt.Println("Description:")
	fmt.Printf("  %s\n", info.Description)
	fmt.Println()
	fmt.Println("System Requirements:")
	fmt.Printf("  Minimum RAM:   %s\n", info.MinRAM)
	fmt.Printf("  Recommended:   %s\n", info.RecommendedHW)
	fmt.Println()
	fmt.Println("Use Cases:")
	for _, u := range info.UseCases {
		fmt.Printf("  ✓ %s\n", u)
	}
	fmt.Println()
	fmt.Printf("Install: portunix aiops model install %s\n", baseName)
}

// handleModelRemove removes a model from Ollama container
func handleModelRemove(args []string) {
	if len(args) == 0 {
		fmt.Println("❌ Model name required")
		fmt.Println("Usage: portunix aiops model remove <model-name>")
		fmt.Println("       portunix aiops model remove <model-name> --container <name>")
		fmt.Println("       portunix aiops model remove <model-name> --force")
		return
	}

	modelName := args[0]
	containerName := OllamaContainerName
	customContainer := false
	force := false

	// Parse flags
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--container":
			if i+1 < len(args) {
				containerName = args[i+1]
				customContainer = true
				i++
			}
		case "--force", "-f":
			force = true
		}
	}

	// Display container being used
	if customContainer {
		fmt.Printf("Using container: %s\n", containerName)
	} else {
		fmt.Printf("Using container: %s (default)\n", containerName)
	}
	fmt.Println()

	// Check if container exists and is running
	exists, running := getContainerStatus(containerName)
	if !exists {
		fmt.Printf("❌ Container '%s' not found\n", containerName)
		return
	}

	if !running {
		fmt.Printf("❌ Container '%s' is not running\n", containerName)
		fmt.Printf("Start it with: portunix container start %s\n", containerName)
		return
	}

	// Confirm deletion unless --force is used
	if !force {
		fmt.Printf("Are you sure you want to remove model '%s'? [y/N]: ", modelName)
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" && response != "yes" && response != "Yes" {
			fmt.Println("Cancelled.")
			return
		}
	}

	fmt.Printf("Removing model: %s\n", modelName)

	runtime := detectContainerRuntime()
	cmd := exec.Command(runtime, "exec", containerName, "ollama", "rm", modelName)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Printf("❌ Failed to remove model: %v\n", err)
		return
	}

	fmt.Printf("✓ Model %s removed from container %s\n", modelName, containerName)
}

// handleModelRun starts interactive chat or single prompt
func handleModelRun(args []string) {
	if len(args) == 0 {
		fmt.Println("❌ Model name required")
		fmt.Println("Usage: portunix aiops model run <model-name>")
		fmt.Println("       portunix aiops model run <model-name> --prompt \"Your question\"")
		fmt.Println("       portunix aiops model run <model-name> --container <name>")
		return
	}

	modelName := args[0]
	containerName := OllamaContainerName
	customContainer := false
	prompt := ""

	// Parse flags
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--prompt":
			if i+1 < len(args) {
				prompt = args[i+1]
				i++
			}
		case "--container":
			if i+1 < len(args) {
				containerName = args[i+1]
				customContainer = true
				i++
			}
		}
	}

	// Check if container exists and is running
	exists, running := getContainerStatus(containerName)
	if !exists {
		fmt.Printf("❌ Container '%s' not found\n", containerName)
		if customContainer {
			fmt.Println("Available Ollama containers:")
			listOllamaContainers()
		} else {
			fmt.Println("Create it with: portunix aiops ollama container create")
		}
		return
	}

	if !running {
		fmt.Printf("❌ Container '%s' is not running\n", containerName)
		fmt.Printf("Start it with: portunix container start %s\n", containerName)
		return
	}

	runtime := detectContainerRuntime()

	if prompt != "" {
		// Single prompt mode
		fmt.Printf("Using model: %s", modelName)
		if customContainer {
			fmt.Printf(" (container: %s)", containerName)
		}
		fmt.Println()
		fmt.Println(strings.Repeat("━", 50))
		fmt.Println()

		cmd := exec.Command(runtime, "exec", containerName, "ollama", "run", modelName, prompt)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()
	} else {
		// Interactive mode
		fmt.Printf("\nUsing model: %s", modelName)
		if customContainer {
			fmt.Printf(" (container: %s)", containerName)
		}
		fmt.Println()
		fmt.Println("Type 'exit' or Ctrl+C to quit.")
		fmt.Println(strings.Repeat("━", 50))

		cmd := exec.Command(runtime, "exec", "-it", containerName, "ollama", "run", modelName)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()
	}
}
