package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var mcpStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check status of Portunix MCP server integration with Claude Code",
	Long: `Check the current status of Portunix MCP server integration with Claude Code.

This command will:
- Check if Claude Code is installed and accessible
- Verify if Portunix MCP server is configured
- Test the MCP server connection (if running)
- Display detailed configuration information

Examples:
  portunix mcp status           # Show status overview
  portunix mcp status --verbose # Show detailed information
  portunix mcp status --json    # Output in JSON format`,
	Run: func(cmd *cobra.Command, args []string) {
		verbose, _ := cmd.Flags().GetBool("verbose")
		jsonOutput, _ := cmd.Flags().GetBool("json")

		if err := checkMCPStatus(verbose, jsonOutput); err != nil {
			log.Fatalf("Failed to check MCP status: %v", err)
		}
	},
}

type MCPStatus struct {
	ClaudeCodeInstalled bool         `json:"claude_code_installed"`
	MCPConfigured       bool         `json:"mcp_configured"`
	MCPServerRunning    bool         `json:"mcp_server_running"`
	PortunixPath        string       `json:"portunix_path"`
	ConfigurationError  string       `json:"configuration_error,omitempty"`
	MCPServerConfig     string       `json:"mcp_server_config,omitempty"`
	ConfiguredPort      int          `json:"configured_port,omitempty"`
	PortAvailable       bool         `json:"port_available"`
	PortConflict        bool         `json:"port_conflict"`
	PortProcessInfo     *ProcessInfo `json:"port_process_info,omitempty"`
	SuggestedPorts      []int        `json:"suggested_ports,omitempty"`
	ServerStartError    string       `json:"server_start_error,omitempty"`
}

type ProcessInfo struct {
	PID         int    `json:"pid"`
	ProcessName string `json:"process_name"`
	CommandLine string `json:"command_line,omitempty"`
}

func checkMCPStatus(verbose, jsonOutput bool) error {
	status := &MCPStatus{}

	// Check Claude Code installation
	status.ClaudeCodeInstalled = isClaudeCodeInstalled()

	// Get Portunix path
	if path, err := getCurrentExecutablePath(); err == nil {
		status.PortunixPath = path
	}

	// Check MCP configuration
	status.MCPConfigured = isMCPAlreadyConfigured()

	// Get configured port from MCP configuration
	status.ConfiguredPort = getConfiguredMCPPort()
	if status.ConfiguredPort == 0 {
		status.ConfiguredPort = 3001 // default port
	}

	// Check port availability and get process info if occupied
	status.PortAvailable = isPortAvailable(status.ConfiguredPort)
	status.PortConflict = !status.PortAvailable

	// Get process information if port is occupied
	if status.PortConflict {
		status.PortProcessInfo = getProcessUsingPort(status.ConfiguredPort)
		status.SuggestedPorts = findAvailablePorts(3)
	}

	// Check if MCP server is running (enhanced check)
	status.MCPServerRunning, status.ServerStartError = isMCPServerRunningEnhanced(status.ConfiguredPort)

	// Get detailed MCP configuration if available
	if status.MCPConfigured {
		if config := getMCPServerConfig(); config != "" {
			status.MCPServerConfig = config
		}
	}

	// Output results
	if jsonOutput {
		return outputStatusAsJSON(status)
	} else {
		return outputStatusAsText(status, verbose)
	}
}

func isMCPServerRunning() bool {
	running, _ := isMCPServerRunningEnhanced(3001)
	return running
}

func isMCPServerRunningEnhanced(port int) (bool, string) {
	// Find claude executable
	claudePath, err := getClaudePath()
	if err != nil {
		return false, "Claude Code not found"
	}

	// Try to get the MCP server status from Claude Code
	cmd := exec.Command(claudePath, "mcp", "get", "portunix")
	output, err := cmd.Output()
	if err == nil {
		outputStr := string(output)
		// Check for connection status indicators
		if strings.Contains(outputStr, "✓") || strings.Contains(outputStr, "✅") ||
			strings.Contains(outputStr, "Connected") || strings.Contains(outputStr, "OK") {
			return true, ""
		}
		if strings.Contains(outputStr, "✗") || strings.Contains(outputStr, "❌") ||
			strings.Contains(outputStr, "Failed to connect") {
			// Server is configured but not responding - check if process is running
			return checkPortAndProcess(port)
		}
	}

	// Fallback to port and process checking
	return checkPortAndProcess(port)
}

func checkPortAndProcess(port int) (bool, string) {
	// Check if port is occupied and by what process
	if !isPortAvailable(port) {
		processInfo := getProcessUsingPort(port)
		if processInfo != nil {
			if strings.Contains(processInfo.ProcessName, "portunix") {
				// Port is occupied by portunix - server process is running but not responding to Claude
				return true, "Running but not responding to Claude Code"
			}
			// Port is occupied by different process
			return false, fmt.Sprintf("Port %d is occupied by %s (PID: %d)", port, processInfo.ProcessName, processInfo.PID)
		}
		// Port is occupied but we couldn't identify the process
		return false, fmt.Sprintf("Port %d is occupied by unknown process", port)
	}

	// Port is available, so server is not running
	return false, "MCP server not running"
}

func getMCPServerConfig() string {
	// Find claude executable
	claudePath, err := getClaudePath()
	if err != nil {
		return ""
	}

	// Get detailed MCP server configuration
	cmd := exec.Command(claudePath, "mcp", "list", "--verbose")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	// Extract portunix-specific configuration
	lines := strings.Split(string(output), "\n")
	var configLines []string
	inPortunixSection := false

	for _, line := range lines {
		if strings.Contains(line, "portunix") {
			inPortunixSection = true
		} else if inPortunixSection && strings.TrimSpace(line) == "" {
			break
		}

		if inPortunixSection {
			configLines = append(configLines, line)
		}
	}

	return strings.Join(configLines, "\n")
}

func outputStatusAsJSON(status *MCPStatus) error {
	output, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal status to JSON: %w", err)
	}

	fmt.Println(string(output))
	return nil
}

func outputStatusAsText(status *MCPStatus, verbose bool) error {
	if !verbose {
		// Brief output
		return outputBriefStatus(status)
	}

	// Verbose output (original detailed format)
	fmt.Println("📊 Portunix MCP Integration Status")
	fmt.Println("==================================")

	// Claude Code installation
	fmt.Print("Claude Code Installation: ")
	if status.ClaudeCodeInstalled {
		fmt.Println("✅ INSTALLED")
	} else {
		fmt.Println("❌ NOT FOUND")
		fmt.Println("  Install Claude Code: curl -fsSL https://claude.ai/cli/install.sh | sh")
	}

	// Portunix path
	fmt.Print("Portunix Executable: ")
	if status.PortunixPath != "" {
		fmt.Printf("✅ %s\n", status.PortunixPath)
	} else {
		fmt.Println("❌ NOT FOUND")
	}

	// MCP configuration
	fmt.Print("MCP Configuration: ")
	if status.MCPConfigured {
		fmt.Println("✅ CONFIGURED")
	} else {
		fmt.Println("❌ NOT CONFIGURED")
		fmt.Println("  Configure: portunix mcp configure")
	}

	// MCP server status with port information
	fmt.Print("MCP Server Status: ")
	if status.MCPServerRunning {
		fmt.Printf("✅ RUNNING (port %d)\n", status.ConfiguredPort)
	} else {
		fmt.Println("❌ NOT RUNNING")
		if status.ServerStartError != "" {
			fmt.Printf("  Error: %s\n", status.ServerStartError)
		}
		if status.MCPConfigured {
			if status.PortConflict {
				fmt.Printf("  ⚠️  Port %d is occupied!\n", status.ConfiguredPort)
				if status.PortProcessInfo != nil {
					fmt.Printf("  Process: %s (PID: %d)\n", status.PortProcessInfo.ProcessName, status.PortProcessInfo.PID)
					if status.PortProcessInfo.CommandLine != "" {
						fmt.Printf("  Command: %s\n", status.PortProcessInfo.CommandLine)
					}
				}
				if len(status.SuggestedPorts) > 0 {
					fmt.Printf("  Try alternative ports: %v\n", status.SuggestedPorts)
					fmt.Printf("  Start with: portunix mcp-server --port %d\n", status.SuggestedPorts[0])
				}
			} else {
				fmt.Println("  Start server: portunix mcp-server")
			}
		}
	}

	// Overall status
	fmt.Println("\nOverall Status:")
	if status.ClaudeCodeInstalled && status.MCPConfigured && status.MCPServerRunning {
		fmt.Println("🎉 FULLY OPERATIONAL")
		fmt.Println("Your Portunix MCP integration is working correctly!")
	} else if status.ClaudeCodeInstalled && status.MCPConfigured {
		fmt.Println("⚠️  CONFIGURED BUT NOT RUNNING")
		fmt.Println("Start the MCP server to enable AI integration.")
	} else if status.ClaudeCodeInstalled {
		fmt.Println("⚠️  READY FOR CONFIGURATION")
		fmt.Println("Run 'portunix mcp configure' to set up integration.")
	} else {
		fmt.Println("❌ NOT READY")
		fmt.Println("Install Claude Code first, then configure integration.")
	}

	// Verbose output
	if status.MCPServerConfig != "" {
		fmt.Println("\nDetailed MCP Configuration:")
		fmt.Println("---------------------------")
		fmt.Println(status.MCPServerConfig)
	}

	// Quick commands
	fmt.Println("\nQuick Commands:")
	if !status.ClaudeCodeInstalled {
		fmt.Println("  curl -fsSL https://claude.ai/cli/install.sh | sh  # Install Claude Code")
	}
	if !status.MCPConfigured {
		fmt.Println("  portunix mcp configure                           # Configure integration")
	}
	if !status.MCPServerRunning && status.MCPConfigured {
		if status.PortConflict && len(status.SuggestedPorts) > 0 {
			fmt.Printf("  portunix mcp-server --port %d                    # Start on available port\n", status.SuggestedPorts[0])
			fmt.Printf("  portunix mcp remove && portunix mcp configure --port %d  # Reconfigure with new port\n", status.SuggestedPorts[0])
		} else {
			fmt.Println("  portunix mcp-server                              # Start MCP server")
		}
	}
	if status.MCPConfigured {
		fmt.Println("  portunix mcp remove                              # Remove integration")
	}

	return nil
}

func outputBriefStatus(status *MCPStatus) error {
	// Brief one-line status
	if status.ClaudeCodeInstalled && status.MCPConfigured && status.MCPServerRunning {
		fmt.Println("🎉 Portunix MCP: OPERATIONAL")
	} else if status.ClaudeCodeInstalled && status.MCPConfigured {
		fmt.Println("⚠️  Portunix MCP: CONFIGURED, NOT RUNNING")
	} else if status.ClaudeCodeInstalled {
		fmt.Println("⚠️  Portunix MCP: READY FOR CONFIGURATION")
	} else {
		fmt.Println("❌ Portunix MCP: NOT READY (install Claude Code)")
	}

	return nil
}

// Port checking functions
func isPortAvailable(port int) bool {
	address := fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		return false
	}
	defer listener.Close()
	return true
}

func findAvailablePorts(count int) []int {
	var availablePorts []int
	startPort := 3001

	for port := startPort; port < startPort+1000 && len(availablePorts) < count; port++ {
		if isPortAvailable(port) {
			availablePorts = append(availablePorts, port)
		}
	}

	return availablePorts
}

func getConfiguredMCPPort() int {
	// Find claude executable
	claudePath, err := getClaudePath()
	if err != nil {
		return 0
	}

	// Get MCP configuration and try to extract port
	cmd := exec.Command(claudePath, "mcp", "show", "portunix")
	output, err := cmd.Output()
	if err != nil {
		return 0
	}

	// Parse output to find port number
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "--port") {
			// Extract port number from line like: --port 3001
			parts := strings.Fields(line)
			for i, part := range parts {
				if part == "--port" && i+1 < len(parts) {
					if port, err := strconv.Atoi(parts[i+1]); err == nil {
						return port
					}
				}
			}
		}
		if strings.Contains(line, "port") && strings.Contains(line, ":") {
			// Extract port from URL like: ws://localhost:3001/mcp
			if idx := strings.Index(line, ":"); idx != -1 {
				portPart := line[idx+1:]
				if slashIdx := strings.Index(portPart, "/"); slashIdx != -1 {
					portPart = portPart[:slashIdx]
				}
				if port, err := strconv.Atoi(strings.TrimSpace(portPart)); err == nil {
					return port
				}
			}
		}
	}

	return 0
}

// getProcessUsingPort returns information about the process using the specified port
func getProcessUsingPort(port int) *ProcessInfo {
	// Try different methods in order of preference

	// Method 1: Try lsof first (most reliable)
	if runtime.GOOS == "linux" {
		if info := getProcessUsingPortLsof(port); info != nil {
			return info
		}
	}

	// Method 2: Try ss command (modern replacement for netstat)
	if runtime.GOOS == "linux" {
		if info := getProcessUsingPortSS(port); info != nil {
			return info
		}
	}

	// Method 3: Try netstat
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("netstat", "-tulpn")
	case "windows":
		cmd = exec.Command("netstat", "-ano")
	default:
		return nil
	}

	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, fmt.Sprintf(":%d", port)) {
			return parseProcessFromNetstat(line, runtime.GOOS)
		}
	}

	return nil
}

func getProcessUsingPortSS(port int) *ProcessInfo {
	cmd := exec.Command("ss", "-tulpn")
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, fmt.Sprintf(":%d", port)) {
			return parseProcessFromSS(line)
		}
	}

	return nil
}

func parseProcessFromSS(line string) *ProcessInfo {
	// SS output format: tcp   LISTEN 0      4096               *:3001             *:*    users:(("portunix",pid=11380,fd=3))
	if !strings.Contains(line, "users:((") {
		return nil
	}

	// Find the users:((...)) part
	usersStart := strings.Index(line, "users:((")
	if usersStart == -1 {
		return nil
	}

	usersEnd := strings.Index(line[usersStart:], "))")
	if usersEnd == -1 {
		return nil
	}

	usersPart := line[usersStart : usersStart+usersEnd+2]
	// Extract content between users:((" and "))
	start := strings.Index(usersPart, "((\"") + 3
	end := strings.Index(usersPart, "\"))")
	if start < 3 || end == -1 {
		return nil
	}

	processInfo := usersPart[start:end]
	// Format: processname",pid=12345,fd=3
	parts := strings.Split(processInfo, ",")
	if len(parts) < 2 {
		return nil
	}

	processName := parts[0]
	pidPart := parts[1]

	// Extract PID from "pid=12345"
	if !strings.HasPrefix(pidPart, "pid=") {
		return nil
	}

	pidStr := strings.TrimPrefix(pidPart, "pid=")
	if pid, err := strconv.Atoi(pidStr); err == nil {
		cmdLine := getProcessCommandLine(pid)
		return &ProcessInfo{
			PID:         pid,
			ProcessName: processName,
			CommandLine: cmdLine,
		}
	}

	return nil
}

func parseProcessFromNetstat(line, osType string) *ProcessInfo {
	fields := strings.Fields(line)

	switch osType {
	case "linux":
		// Linux netstat format: ... LISTEN pid/processname
		if len(fields) >= 7 && (strings.Contains(line, "LISTEN") || strings.Contains(line, "tcp")) {
			pidProcess := fields[len(fields)-1]
			if pidProcess != "-" && strings.Contains(pidProcess, "/") {
				parts := strings.Split(pidProcess, "/")
				if len(parts) >= 2 {
					if pid, err := strconv.Atoi(parts[0]); err == nil {
						processName := parts[1]
						cmdLine := getProcessCommandLine(pid)
						return &ProcessInfo{
							PID:         pid,
							ProcessName: processName,
							CommandLine: cmdLine,
						}
					}
				}
			}
		}
	case "windows":
		// Windows netstat format: ... LISTENING pid
		if len(fields) >= 4 && strings.Contains(line, "LISTENING") {
			pidStr := fields[len(fields)-1]
			if pid, err := strconv.Atoi(pidStr); err == nil {
				processName := getWindowsProcessName(pid)
				cmdLine := getProcessCommandLine(pid)
				return &ProcessInfo{
					PID:         pid,
					ProcessName: processName,
					CommandLine: cmdLine,
				}
			}
		}
	}

	return nil
}

func getProcessUsingPortLsof(port int) *ProcessInfo {
	cmd := exec.Command("lsof", "-i", fmt.Sprintf(":%d", port), "-t")
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	pidStr := strings.TrimSpace(string(output))
	if pidStr == "" {
		return nil
	}

	lines := strings.Split(pidStr, "\n")
	if len(lines) == 0 {
		return nil
	}

	firstPidStr := strings.TrimSpace(lines[0])
	if pid, err := strconv.Atoi(firstPidStr); err == nil {
		processName := getLinuxProcessName(pid)
		cmdLine := getProcessCommandLine(pid)
		return &ProcessInfo{
			PID:         pid,
			ProcessName: processName,
			CommandLine: cmdLine,
		}
	}

	return nil
}

func getLinuxProcessName(pid int) string {
	cmd := exec.Command("ps", "-p", strconv.Itoa(pid), "-o", "comm=")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Sprintf("PID:%d", pid)
	}
	return strings.TrimSpace(string(output))
}

func getWindowsProcessName(pid int) string {
	cmd := exec.Command("tasklist", "/FI", fmt.Sprintf("PID eq %d", pid), "/FO", "CSV", "/NH")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Sprintf("PID:%d", pid)
	}

	line := strings.TrimSpace(string(output))
	if line != "" {
		// Parse CSV format: "processname.exe","pid","sessionname","session#","memusage"
		parts := strings.Split(line, ",")
		if len(parts) > 0 {
			processName := strings.Trim(parts[0], "\"")
			return processName
		}
	}

	return fmt.Sprintf("PID:%d", pid)
}

func getProcessCommandLine(pid int) string {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "linux":
		cmd = exec.Command("ps", "-p", strconv.Itoa(pid), "-o", "args=")
	case "windows":
		cmd = exec.Command("wmic", "process", "where", fmt.Sprintf("ProcessId=%d", pid), "get", "CommandLine", "/value")
	default:
		return ""
	}

	output, err := cmd.Output()
	if err != nil {
		return ""
	}

	cmdLine := strings.TrimSpace(string(output))

	if runtime.GOOS == "windows" {
		// Parse wmic output: CommandLine=actual_command_line
		lines := strings.Split(cmdLine, "\n")
		for _, line := range lines {
			if strings.HasPrefix(line, "CommandLine=") {
				return strings.TrimPrefix(line, "CommandLine=")
			}
		}
		return ""
	}

	// Truncate very long command lines
	if len(cmdLine) > 100 {
		cmdLine = cmdLine[:97] + "..."
	}

	return cmdLine
}

func init() {
	mcpCmd.AddCommand(mcpStatusCmd)

	mcpStatusCmd.Flags().BoolP("verbose", "v", false, "Show detailed information")
	mcpStatusCmd.Flags().BoolP("json", "j", false, "Output in JSON format")
}
