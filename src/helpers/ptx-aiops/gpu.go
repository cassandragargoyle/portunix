package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"
)

// GPUInfo represents NVIDIA GPU information
type GPUInfo struct {
	Index            int
	Name             string
	DriverVersion    string
	CUDAVersion      string
	UtilizationGPU   int    // percent
	UtilizationMem   int    // percent
	MemoryUsed       int64  // bytes
	MemoryTotal      int64  // bytes
	Temperature      int    // celsius
	PowerDraw        int    // watts
	PowerLimit       int    // watts
	FanSpeed         int    // percent
	ComputeMode      string
	PersistenceMode  string
}

// ProcessInfo represents a process using GPU
type ProcessInfo struct {
	PID        int
	Name       string
	GPUMemory  int64  // bytes
	GPUUsage   int    // percent (if available)
}

// ContainerToolkitInfo represents NVIDIA Container Toolkit status
type ContainerToolkitInfo struct {
	Installed bool
	Version   string
	Runtime   string // docker or podman
	Verified  bool   // GPU access verified in container
}

// isNvidiaSMIAvailable checks if nvidia-smi command is available
func isNvidiaSMIAvailable() bool {
	_, err := exec.LookPath("nvidia-smi")
	return err == nil
}

// detectNvidiaGPUs detects and returns information about NVIDIA GPUs
func detectNvidiaGPUs() ([]GPUInfo, error) {
	if !isNvidiaSMIAvailable() {
		return nil, fmt.Errorf("nvidia-smi not found - NVIDIA drivers may not be installed")
	}

	// Query GPU information using nvidia-smi with CSV output
	cmd := exec.Command("nvidia-smi",
		"--query-gpu=index,name,driver_version,memory.total,memory.used,utilization.gpu,utilization.memory,temperature.gpu,power.draw,power.limit,fan.speed",
		"--format=csv,noheader,nounits")

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to query GPU info: %v", err)
	}

	var gpus []GPUInfo
	scanner := bufio.NewScanner(strings.NewReader(string(output)))

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, ", ")

		if len(parts) < 11 {
			continue
		}

		gpu := GPUInfo{}

		// Parse each field
		gpu.Index, _ = strconv.Atoi(strings.TrimSpace(parts[0]))
		gpu.Name = strings.TrimSpace(parts[1])
		gpu.DriverVersion = strings.TrimSpace(parts[2])

		// Memory total (in MiB from nvidia-smi, convert to bytes)
		memTotal, _ := strconv.ParseInt(strings.TrimSpace(parts[3]), 10, 64)
		gpu.MemoryTotal = memTotal * 1024 * 1024

		// Memory used
		memUsed, _ := strconv.ParseInt(strings.TrimSpace(parts[4]), 10, 64)
		gpu.MemoryUsed = memUsed * 1024 * 1024

		// Utilization GPU
		gpu.UtilizationGPU, _ = strconv.Atoi(strings.TrimSpace(parts[5]))

		// Utilization memory
		gpu.UtilizationMem, _ = strconv.Atoi(strings.TrimSpace(parts[6]))

		// Temperature
		gpu.Temperature, _ = strconv.Atoi(strings.TrimSpace(parts[7]))

		// Power draw
		powerDraw, _ := strconv.ParseFloat(strings.TrimSpace(parts[8]), 64)
		gpu.PowerDraw = int(powerDraw)

		// Power limit
		powerLimit, _ := strconv.ParseFloat(strings.TrimSpace(parts[9]), 64)
		gpu.PowerLimit = int(powerLimit)

		// Fan speed
		fanStr := strings.TrimSpace(parts[10])
		if fanStr != "[N/A]" && fanStr != "" {
			gpu.FanSpeed, _ = strconv.Atoi(fanStr)
		}

		gpus = append(gpus, gpu)
	}

	// Get CUDA version separately
	cudaVersion := getCUDAVersion()
	for i := range gpus {
		gpus[i].CUDAVersion = cudaVersion
	}

	return gpus, nil
}

// getCUDAVersion extracts CUDA version from nvidia-smi
func getCUDAVersion() string {
	cmd := exec.Command("nvidia-smi", "--query-gpu=driver_version", "--format=csv,noheader")
	// CUDA version is shown in the header, not in query-gpu
	// We need to parse the full output
	cmd = exec.Command("nvidia-smi")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	// Look for CUDA Version in output
	re := regexp.MustCompile(`CUDA Version:\s*(\d+\.\d+)`)
	matches := re.FindStringSubmatch(string(output))
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// getGPUProcesses returns list of processes using GPU
func getGPUProcesses() ([]ProcessInfo, error) {
	if !isNvidiaSMIAvailable() {
		return nil, fmt.Errorf("nvidia-smi not found")
	}

	cmd := exec.Command("nvidia-smi",
		"--query-compute-apps=pid,process_name,used_memory",
		"--format=csv,noheader,nounits")

	output, err := cmd.Output()
	if err != nil {
		// No processes using GPU is not an error
		return []ProcessInfo{}, nil
	}

	var processes []ProcessInfo
	scanner := bufio.NewScanner(strings.NewReader(string(output)))

	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}

		parts := strings.Split(line, ", ")
		if len(parts) < 3 {
			continue
		}

		proc := ProcessInfo{}
		proc.PID, _ = strconv.Atoi(strings.TrimSpace(parts[0]))
		proc.Name = strings.TrimSpace(parts[1])

		memMiB, _ := strconv.ParseInt(strings.TrimSpace(parts[2]), 10, 64)
		proc.GPUMemory = memMiB * 1024 * 1024

		processes = append(processes, proc)
	}

	return processes, nil
}

// checkContainerToolkit checks if NVIDIA Container Toolkit is installed
func checkContainerToolkit() ContainerToolkitInfo {
	info := ContainerToolkitInfo{}

	// Check for nvidia-ctk
	cmd := exec.Command("nvidia-ctk", "--version")
	output, err := cmd.Output()
	if err == nil {
		info.Installed = true
		// Parse version from output
		versionStr := strings.TrimSpace(string(output))
		if strings.Contains(versionStr, "version") {
			parts := strings.Fields(versionStr)
			for i, p := range parts {
				if p == "version" && i+1 < len(parts) {
					info.Version = parts[i+1]
					break
				}
			}
		}
	}

	// Detect container runtime
	if isDockerAvailable() {
		info.Runtime = "docker"
	} else if isPodmanAvailable() {
		info.Runtime = "podman"
	}

	// Verify GPU access in container (only if toolkit installed and runtime available)
	if info.Installed && info.Runtime != "" {
		info.Verified = verifyContainerGPUAccess(info.Runtime)
	}

	return info
}

// verifyContainerGPUAccess tests if GPU is accessible from container
func verifyContainerGPUAccess(runtime string) bool {
	var cmd *exec.Cmd

	if runtime == "docker" {
		cmd = exec.Command("docker", "run", "--rm", "--gpus", "all",
			"nvidia/cuda:12.0-base-ubuntu22.04", "nvidia-smi", "--query-gpu=name", "--format=csv,noheader")
	} else {
		cmd = exec.Command("podman", "run", "--rm", "--device", "nvidia.com/gpu=all",
			"nvidia/cuda:12.0-base-ubuntu22.04", "nvidia-smi", "--query-gpu=name", "--format=csv,noheader")
	}

	err := cmd.Run()
	return err == nil
}

// isDockerAvailable checks if Docker is available
func isDockerAvailable() bool {
	cmd := exec.Command("docker", "version", "--format", "{{.Client.Version}}")
	return cmd.Run() == nil
}

// isPodmanAvailable checks if Podman is available
func isPodmanAvailable() bool {
	cmd := exec.Command("podman", "version", "--format", "{{.Client.Version}}")
	return cmd.Run() == nil
}

// formatBytes formats bytes to human-readable string
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// generateProgressBar creates ASCII progress bar
func generateProgressBar(percent int, width int) string {
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}

	filled := (percent * width) / 100
	empty := width - filled

	bar := strings.Repeat("█", filled) + strings.Repeat("░", empty)
	return bar
}

// handleGPUStatus shows GPU status
func handleGPUStatus(args []string) {
	// Parse flags
	watch := false
	interval := 5 // default 5 seconds

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--watch", "-w":
			watch = true
		case "--interval":
			if i+1 < len(args) {
				interval, _ = strconv.Atoi(args[i+1])
				if interval < 1 {
					interval = 1
				}
				i++
			}
		}
	}

	if watch {
		runGPUStatusWatch(interval)
	} else {
		displayGPUStatus()
	}
}

// displayGPUStatus shows single GPU status snapshot
func displayGPUStatus() {
	gpus, err := detectNvidiaGPUs()
	if err != nil {
		fmt.Println("NVIDIA GPU Status")
		fmt.Println(strings.Repeat("━", 78))
		fmt.Println()
		fmt.Printf("❌ %v\n", err)
		fmt.Println()
		fmt.Println("To use GPU features, ensure NVIDIA drivers are installed.")
		fmt.Println("Install via: sudo apt install nvidia-driver-XXX (Ubuntu/Debian)")
		fmt.Println("         or: sudo dnf install nvidia-driver (Fedora)")
		return
	}

	if len(gpus) == 0 {
		fmt.Println("No NVIDIA GPUs detected")
		return
	}

	fmt.Println()
	fmt.Println("NVIDIA GPU Status")
	fmt.Println(strings.Repeat("━", 78))
	fmt.Println()

	for _, gpu := range gpus {
		fmt.Printf("GPU %d: %s\n", gpu.Index, gpu.Name)
		fmt.Printf("  Driver Version:    %s\n", gpu.DriverVersion)
		if gpu.CUDAVersion != "" {
			fmt.Printf("  CUDA Version:      %s\n", gpu.CUDAVersion)
		}
		fmt.Println()

		// Utilization with progress bar
		memPercent := 0
		if gpu.MemoryTotal > 0 {
			memPercent = int((gpu.MemoryUsed * 100) / gpu.MemoryTotal)
		}

		fmt.Printf("  Utilization:       %3d%%  [%s]\n", gpu.UtilizationGPU, generateProgressBar(gpu.UtilizationGPU, 29))
		fmt.Printf("  Memory:            %s / %s (%d%%)\n", formatBytes(gpu.MemoryUsed), formatBytes(gpu.MemoryTotal), memPercent)
		fmt.Printf("                     [%s]\n", generateProgressBar(memPercent, 29))
		fmt.Printf("  Temperature:       %d°C\n", gpu.Temperature)

		if gpu.PowerLimit > 0 {
			powerPercent := (gpu.PowerDraw * 100) / gpu.PowerLimit
			fmt.Printf("  Power:             %dW / %dW (%d%%)\n", gpu.PowerDraw, gpu.PowerLimit, powerPercent)
		}

		if gpu.FanSpeed > 0 {
			fmt.Printf("  Fan Speed:         %d%%\n", gpu.FanSpeed)
		}
		fmt.Println()
	}

	// Container toolkit status
	toolkit := checkContainerToolkit()
	if toolkit.Installed {
		fmt.Printf("Container Toolkit:   ✓ Installed")
		if toolkit.Version != "" {
			fmt.Printf(" (v%s)", toolkit.Version)
		}
		fmt.Println()
		if toolkit.Runtime != "" {
			status := "GPU access verified"
			if !toolkit.Verified {
				status = "GPU access NOT verified"
			}
			fmt.Printf("Container Runtime:   %s (%s)\n", strings.Title(toolkit.Runtime), status)
		}
	} else {
		fmt.Println("Container Toolkit:   ✗ Not installed")
		fmt.Println("                     Run: portunix install nvidia-container-toolkit")
	}
	fmt.Println(strings.Repeat("━", 78))
}

// runGPUStatusWatch runs GPU status in watch mode with auto-refresh
func runGPUStatusWatch(interval int) {
	// Set up signal handler for clean exit
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	defer ticker.Stop()

	// Display initial status
	clearScreen()
	displayGPUStatusWatch(interval)

	for {
		select {
		case <-ticker.C:
			clearScreen()
			displayGPUStatusWatch(interval)
		case <-sigChan:
			fmt.Println("\nExiting GPU monitor...")
			return
		}
	}
}

// clearScreen clears terminal screen
func clearScreen() {
	fmt.Print("\033[H\033[2J")
}

// displayGPUStatusWatch shows GPU status for watch mode
func displayGPUStatusWatch(interval int) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")

	gpus, err := detectNvidiaGPUs()
	if err != nil {
		fmt.Printf("NVIDIA GPU Monitor (Refreshing every %ds, press Ctrl+C to exit)\n", interval)
		fmt.Println(strings.Repeat("━", 78))
		fmt.Println()
		fmt.Printf("❌ GPU not available: %v\n", err)
		fmt.Println()
		fmt.Println("Possible causes:")
		fmt.Println("  • NVIDIA drivers not installed")
		fmt.Println("  • GPU in use by another process exclusively")
		fmt.Println("  • Driver/hardware issue")
		fmt.Println()
		fmt.Printf("Last check: %s\n", timestamp)
		fmt.Println(strings.Repeat("━", 78))
		fmt.Println("Waiting for GPU to become available...")
		return
	}

	if len(gpus) == 0 {
		fmt.Printf("NVIDIA GPU Monitor (Refreshing every %ds, press Ctrl+C to exit)\n", interval)
		fmt.Println(strings.Repeat("━", 78))
		fmt.Println()
		fmt.Println("No NVIDIA GPUs detected")
		fmt.Printf("Last check: %s\n", timestamp)
		return
	}

	fmt.Printf("NVIDIA GPU Monitor (Refreshing every %ds, press Ctrl+C to exit)\n", interval)
	fmt.Println(strings.Repeat("━", 78))
	fmt.Println()

	for _, gpu := range gpus {
		fmt.Printf("GPU %d: %-40s %s\n", gpu.Index, gpu.Name, timestamp)
		fmt.Println("┌──────────────┬────────────────────────────────────────────────────────┬───────┐")

		memPercent := 0
		if gpu.MemoryTotal > 0 {
			memPercent = int((gpu.MemoryUsed * 100) / gpu.MemoryTotal)
		}

		fmt.Printf("│ UTILIZATION  │ %-54s │ %3d%% │\n", generateProgressBar(gpu.UtilizationGPU, 54), gpu.UtilizationGPU)
		fmt.Printf("│ MEMORY       │ %-54s │ %3d%% │\n", generateProgressBar(memPercent, 54), memPercent)
		fmt.Printf("│ TEMPERATURE  │ %-54s │ %3d°C │\n", generateProgressBar(gpu.Temperature, 54), gpu.Temperature)

		if gpu.PowerLimit > 0 {
			powerPercent := (gpu.PowerDraw * 100) / gpu.PowerLimit
			fmt.Printf("│ POWER        │ %-54s │ %3dW │\n", generateProgressBar(powerPercent, 54), gpu.PowerDraw)
		}

		if gpu.FanSpeed > 0 {
			fmt.Printf("│ FAN          │ %-54s │ %3d%% │\n", generateProgressBar(gpu.FanSpeed, 54), gpu.FanSpeed)
		}

		fmt.Println("└──────────────┴────────────────────────────────────────────────────────┴───────┘")
		fmt.Println()
		fmt.Printf("Memory: %s / %s    Power Limit: %dW    CUDA: %s\n",
			formatBytes(gpu.MemoryUsed), formatBytes(gpu.MemoryTotal), gpu.PowerLimit, gpu.CUDAVersion)
	}
	fmt.Println(strings.Repeat("━", 78))
}

// handleGPUUsage shows compact GPU utilization summary
func handleGPUUsage() {
	gpus, err := detectNvidiaGPUs()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	if len(gpus) == 0 {
		fmt.Println("No NVIDIA GPUs detected")
		return
	}

	fmt.Println()
	fmt.Println("GPU Utilization Summary:")

	for _, gpu := range gpus {
		memPercent := 0
		if gpu.MemoryTotal > 0 {
			memPercent = int((gpu.MemoryUsed * 100) / gpu.MemoryTotal)
		}

		fmt.Println("┌─────────────────────────────────────────────────────────────────────────────┐")
		fmt.Printf("│ GPU %d: %-69s │\n", gpu.Index, gpu.Name)
		fmt.Println("├─────────────────────────────────────────────────────────────────────────────┤")
		fmt.Printf("│ Compute:  %-20s %3d%%   │ Memory:   %-20s %3d%% │\n",
			generateProgressBar(gpu.UtilizationGPU, 20), gpu.UtilizationGPU,
			generateProgressBar(memPercent, 20), memPercent)
		fmt.Println("├─────────────────────────────────────────────────────────────────────────────┤")

		fmt.Printf("│ Temp: %3d°C  │  Power: %3dW/%3dW  │  Fan: %3d%%  │  Memory: %s/%s │\n",
			gpu.Temperature, gpu.PowerDraw, gpu.PowerLimit, gpu.FanSpeed,
			formatBytes(gpu.MemoryUsed), formatBytes(gpu.MemoryTotal))
		fmt.Println("└─────────────────────────────────────────────────────────────────────────────┘")
	}
}

// handleGPUProcesses lists processes using GPU
func handleGPUProcesses() {
	processes, err := getGPUProcesses()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println()
	fmt.Println("GPU Processes:")

	if len(processes) == 0 {
		fmt.Println("No processes currently using GPU")
		return
	}

	fmt.Println("┌───────┬────────────────────────────────┬────────────┬─────────────┐")
	fmt.Println("│ PID   │ PROCESS                        │ GPU MEMORY │ GPU USAGE   │")
	fmt.Println("├───────┼────────────────────────────────┼────────────┼─────────────┤")

	totalMem := int64(0)
	for _, proc := range processes {
		name := proc.Name
		if len(name) > 30 {
			name = name[:27] + "..."
		}

		usageStr := "-"
		if proc.GPUUsage > 0 {
			usageStr = fmt.Sprintf("%d%%", proc.GPUUsage)
		}

		fmt.Printf("│ %5d │ %-30s │ %10s │ %11s │\n",
			proc.PID, name, formatBytes(proc.GPUMemory), usageStr)
		totalMem += proc.GPUMemory
	}

	fmt.Println("└───────┴────────────────────────────────┴────────────┴─────────────┘")

	// Get total GPU memory for percentage
	gpus, err := detectNvidiaGPUs()
	if err == nil && len(gpus) > 0 {
		totalGPUMem := gpus[0].MemoryTotal
		percent := 0
		if totalGPUMem > 0 {
			percent = int((totalMem * 100) / totalGPUMem)
		}
		fmt.Printf("\nTotal GPU Memory Used: %s / %s (%d%%)\n",
			formatBytes(totalMem), formatBytes(totalGPUMem), percent)
	}
}

// handleGPUCheck verifies GPU and toolkit readiness for containers
func handleGPUCheck() {
	fmt.Println()
	fmt.Println("GPU Container Readiness Check")
	fmt.Println(strings.Repeat("━", 50))
	fmt.Println()

	allReady := true

	// 1. Check NVIDIA GPU
	fmt.Print("1. NVIDIA GPU Detection: ")
	gpus, err := detectNvidiaGPUs()
	if err != nil {
		fmt.Printf("❌ %v\n", err)
		allReady = false
	} else if len(gpus) == 0 {
		fmt.Println("❌ No NVIDIA GPUs detected")
		allReady = false
	} else {
		fmt.Printf("✓ %d GPU(s) detected\n", len(gpus))
		for _, gpu := range gpus {
			fmt.Printf("   └─ %s (Driver: %s, CUDA: %s)\n", gpu.Name, gpu.DriverVersion, gpu.CUDAVersion)
		}
	}

	// 2. Check Container Runtime
	fmt.Print("\n2. Container Runtime: ")
	runtime := ""
	if isDockerAvailable() {
		runtime = "docker"
		fmt.Println("✓ Docker available")
	} else if isPodmanAvailable() {
		runtime = "podman"
		fmt.Println("✓ Podman available")
	} else {
		fmt.Println("❌ Neither Docker nor Podman available")
		allReady = false
	}

	// 3. Check NVIDIA Container Toolkit
	fmt.Print("\n3. NVIDIA Container Toolkit: ")
	toolkit := checkContainerToolkit()
	if toolkit.Installed {
		fmt.Printf("✓ Installed")
		if toolkit.Version != "" {
			fmt.Printf(" (v%s)", toolkit.Version)
		}
		fmt.Println()
	} else {
		fmt.Println("❌ Not installed")
		fmt.Println("   Install with: portunix install nvidia-container-toolkit")
		allReady = false
	}

	// 4. Verify GPU access in container
	if toolkit.Installed && runtime != "" && len(gpus) > 0 {
		fmt.Print("\n4. Container GPU Access: ")
		if toolkit.Verified {
			fmt.Println("✓ GPU accessible from containers")
		} else {
			fmt.Println("⚠ GPU access not verified (may require restart)")
			fmt.Println("   Try: sudo systemctl restart", runtime)
		}
	}

	// Summary
	fmt.Println()
	fmt.Println(strings.Repeat("━", 50))
	if allReady && toolkit.Verified {
		fmt.Println("✓ System ready for GPU-accelerated containers!")
		fmt.Println()
		fmt.Println("Next steps:")
		fmt.Println("  portunix aiops ollama container create  - Create Ollama with GPU")
		fmt.Println("  portunix aiops stack create             - Create full AI stack")
	} else if allReady && !toolkit.Verified {
		fmt.Println("⚠ System partially ready")
		fmt.Println()
		fmt.Println("GPU access verification failed. Try:")
		fmt.Println("  1. Restart container runtime")
		fmt.Println("  2. Reboot system")
		fmt.Println("  3. Verify nvidia-ctk runtime configure was run")
	} else {
		fmt.Println("❌ System not ready for GPU containers")
		fmt.Println()
		fmt.Println("Required actions:")
		if len(gpus) == 0 {
			fmt.Println("  • Install NVIDIA drivers")
		}
		if runtime == "" {
			fmt.Println("  • Install Docker or Podman")
		}
		if !toolkit.Installed {
			fmt.Println("  • Run: portunix install nvidia-container-toolkit")
		}
	}
}
