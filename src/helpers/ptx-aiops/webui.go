package main

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	// Default WebUI container settings
	WebUIContainerName = "portunix-webui"
	WebUIImage         = "ghcr.io/open-webui/open-webui:main"
	WebUIPort          = "3000"
	WebUIInternalPort  = "8080"
	WebUIEndpoint      = "http://localhost:3000"
)

// WebUIContainerStatus represents the status of WebUI container
type WebUIContainerStatus struct {
	Exists  bool
	Running bool
	Name    string
	Image   string
	Port    string
	Created string
}

// getWebUIDataDir returns the directory for WebUI data storage
func getWebUIDataDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "/tmp/portunix/aiops/webui"
	}
	return filepath.Join(homeDir, ".portunix", "aiops", "webui", "data")
}

// ensureWebUIDataDir creates WebUI data directory if it doesn't exist
func ensureWebUIDataDir() error {
	dataDir := getWebUIDataDir()
	return os.MkdirAll(dataDir, 0755)
}

// getWebUIContainerStatus checks the status of WebUI container
func getWebUIContainerStatus() WebUIContainerStatus {
	status := WebUIContainerStatus{
		Name: WebUIContainerName,
	}

	runtimeCmd := detectContainerRuntime()
	if runtimeCmd == "" {
		return status
	}

	// Check if container exists
	cmd := exec.Command(runtimeCmd, "inspect", WebUIContainerName, "--format", "{{.State.Running}}")
	output, err := cmd.Output()
	if err != nil {
		return status
	}

	status.Exists = true
	status.Running = strings.TrimSpace(string(output)) == "true"

	// Get more details
	cmd = exec.Command(runtimeCmd, "inspect", WebUIContainerName,
		"--format", "{{.Config.Image}}|{{.Created}}")
	output, err = cmd.Output()
	if err == nil {
		parts := strings.Split(strings.TrimSpace(string(output)), "|")
		if len(parts) >= 2 {
			status.Image = parts[0]
			status.Created = parts[1]
		}
	}

	status.Port = WebUIPort
	return status
}

// handleWebUIContainerCreate creates a new Open WebUI container
func handleWebUIContainerCreate(args []string) {
	runtimeCmd := detectContainerRuntime()
	if runtimeCmd == "" {
		fmt.Println("❌ No container runtime available")
		fmt.Println("Install Docker or Podman first")
		return
	}

	// Check if container already exists
	status := getWebUIContainerStatus()
	if status.Exists {
		fmt.Printf("❌ Container '%s' already exists\n", WebUIContainerName)
		if status.Running {
			fmt.Println("Container is running.")
		} else {
			fmt.Println("Container is stopped. Start with: portunix aiops webui container start")
		}
		return
	}

	// Ensure data directory exists
	if err := ensureWebUIDataDir(); err != nil {
		fmt.Printf("❌ Failed to create data directory: %v\n", err)
		return
	}

	dataDir := getWebUIDataDir()

	fmt.Println("Creating Open WebUI container...")
	fmt.Printf("  Name:   %s\n", WebUIContainerName)
	fmt.Printf("  Image:  %s\n", WebUIImage)
	fmt.Printf("  Port:   %s\n", WebUIPort)
	fmt.Printf("  Data:   %s\n", dataDir)

	// Check if Ollama container is running for connection
	ollamaStatus := getOllamaContainerStatus()
	ollamaURL := OllamaAPIEndpoint
	if ollamaStatus.Running {
		fmt.Printf("  Ollama: ✓ Connected to %s\n", ollamaURL)
	} else {
		fmt.Println("  Ollama: ⚠ Not running (WebUI will work but can't use local models)")
	}

	// Build container command
	cmdArgs := []string{
		"run", "-d",
		"--name", WebUIContainerName,
		"-p", WebUIPort + ":" + WebUIInternalPort,
		"-v", dataDir + ":/app/backend/data",
		"-e", "OLLAMA_BASE_URL=" + ollamaURL,
		"--add-host=host.docker.internal:host-gateway",
		"--restart", "unless-stopped",
		WebUIImage,
	}

	fmt.Println()
	fmt.Printf("Running: %s %s\n", runtimeCmd, strings.Join(cmdArgs, " "))

	cmd := exec.Command(runtimeCmd, cmdArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("❌ Failed to create container: %v\n", err)
		fmt.Printf("Output: %s\n", string(output))
		return
	}

	fmt.Println()
	fmt.Printf("✓ Container '%s' created successfully\n", WebUIContainerName)
	fmt.Println()

	// Wait for container to be ready
	fmt.Print("Waiting for Open WebUI to start...")
	ready := false
	for i := 0; i < 60; i++ {
		time.Sleep(time.Second)
		fmt.Print(".")

		resp, err := http.Get(WebUIEndpoint)
		if err == nil {
			resp.Body.Close()
			ready = true
			break
		}
	}

	fmt.Println()
	if ready {
		fmt.Println("✓ Open WebUI is ready!")
		fmt.Println()
		fmt.Printf("Access WebUI at: %s\n", WebUIEndpoint)
		fmt.Println()
		fmt.Println("Next steps:")
		fmt.Println("  portunix aiops webui open       - Open WebUI in browser")
		fmt.Println("  portunix aiops model list       - List available models")
	} else {
		fmt.Println("⚠ Open WebUI is taking longer than expected to start")
		fmt.Println("Check logs with: portunix container logs", WebUIContainerName)
	}
}

// handleWebUIContainerStatus shows WebUI container status
func handleWebUIContainerStatus() {
	status := getWebUIContainerStatus()

	fmt.Println()
	fmt.Println("Open WebUI Container Status")
	fmt.Println(strings.Repeat("━", 50))
	fmt.Println()

	if !status.Exists {
		fmt.Printf("Container '%s': Not found\n", WebUIContainerName)
		fmt.Println()
		fmt.Println("Create container with:")
		fmt.Println("  portunix aiops webui container create")
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
	fmt.Printf("Web Port:    %s\n", status.Port)

	if status.Created != "" {
		fmt.Printf("Created:     %s\n", status.Created[:19])
	}

	// Check web availability if running
	if status.Running {
		fmt.Println()
		fmt.Print("Web Status:  ")
		resp, err := http.Get(WebUIEndpoint)
		if err == nil {
			resp.Body.Close()
			fmt.Println("✓ Available at", WebUIEndpoint)
		} else {
			fmt.Println("⚠ Not responding")
		}
	}

	fmt.Println()
	fmt.Println(strings.Repeat("━", 50))

	if status.Running {
		fmt.Println("Commands:")
		fmt.Println("  portunix aiops webui open             - Open in browser")
		fmt.Println("  portunix aiops webui container stop   - Stop container")
	} else {
		fmt.Println("Start container with:")
		fmt.Println("  portunix aiops webui container start")
	}
}

// handleWebUIContainerStart starts the WebUI container
func handleWebUIContainerStart() {
	status := getWebUIContainerStatus()

	if !status.Exists {
		fmt.Printf("❌ Container '%s' does not exist\n", WebUIContainerName)
		fmt.Println("Create it with: portunix aiops webui container create")
		return
	}

	if status.Running {
		fmt.Printf("Container '%s' is already running\n", WebUIContainerName)
		return
	}

	runtimeCmd := detectContainerRuntime()
	cmd := exec.Command(runtimeCmd, "start", WebUIContainerName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("❌ Failed to start container: %v\n", err)
		fmt.Printf("Output: %s\n", string(output))
		return
	}

	fmt.Printf("✓ Container '%s' started\n", WebUIContainerName)
	fmt.Printf("Access WebUI at: %s\n", WebUIEndpoint)
}

// handleWebUIContainerStop stops the WebUI container
func handleWebUIContainerStop() {
	status := getWebUIContainerStatus()

	if !status.Exists {
		fmt.Printf("❌ Container '%s' does not exist\n", WebUIContainerName)
		return
	}

	if !status.Running {
		fmt.Printf("Container '%s' is already stopped\n", WebUIContainerName)
		return
	}

	runtimeCmd := detectContainerRuntime()
	cmd := exec.Command(runtimeCmd, "stop", WebUIContainerName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("❌ Failed to stop container: %v\n", err)
		fmt.Printf("Output: %s\n", string(output))
		return
	}

	fmt.Printf("✓ Container '%s' stopped\n", WebUIContainerName)
}

// handleWebUIContainerRemove removes the WebUI container
func handleWebUIContainerRemove() {
	status := getWebUIContainerStatus()

	if !status.Exists {
		fmt.Printf("Container '%s' does not exist\n", WebUIContainerName)
		return
	}

	runtimeCmd := detectContainerRuntime()

	// Stop first if running
	if status.Running {
		fmt.Printf("Stopping container '%s'...\n", WebUIContainerName)
		cmd := exec.Command(runtimeCmd, "stop", WebUIContainerName)
		cmd.Run()
	}

	// Remove container
	fmt.Printf("Removing container '%s'...\n", WebUIContainerName)
	cmd := exec.Command(runtimeCmd, "rm", WebUIContainerName)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("❌ Failed to remove container: %v\n", err)
		fmt.Printf("Output: %s\n", string(output))
		return
	}

	fmt.Printf("✓ Container '%s' removed\n", WebUIContainerName)
	fmt.Println()
	fmt.Printf("Note: User data preserved in %s\n", getWebUIDataDir())
}

// handleWebUIOpen opens the WebUI in default browser
func handleWebUIOpen() {
	status := getWebUIContainerStatus()

	if !status.Running {
		fmt.Println("❌ Open WebUI container is not running")
		fmt.Println("Start it with: portunix aiops webui container start")
		return
	}

	url := WebUIEndpoint

	// Check if WebUI is responding
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("⚠ WebUI is not responding yet")
		fmt.Println("Check logs with: portunix container logs", WebUIContainerName)
		return
	}
	resp.Body.Close()

	fmt.Printf("Opening %s in browser...\n", url)

	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "darwin":
		cmd = exec.Command("open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		fmt.Println("Unable to open browser automatically")
		fmt.Printf("Please open %s in your browser\n", url)
		return
	}

	if err := cmd.Start(); err != nil {
		fmt.Printf("Failed to open browser: %v\n", err)
		fmt.Printf("Please open %s manually\n", url)
	}
}
