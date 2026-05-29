/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package main

import (
	"bufio"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// initOptions bundles the flag values for the `mcp init` command.
type initOptions struct {
	assistant   string
	serverType  string
	preset      string
	force       bool
	scope       string
	env         map[string]string
	timeout     string
	fromJSON    string
	bindAddress string
	tlsCert     string
	tlsKey      string
	// installMissing auto-installs missing AI assistants via ptx-installer
	// instead of erroring/prompting (issue #035 auto-dependency hook).
	installMissing bool
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize MCP server configuration with interactive wizard",
	Long: `Initialize MCP server configuration for AI assistant integration.

This interactive wizard will guide you through:
- Detecting installed AI assistants (Claude Code, Claude Desktop, Gemini CLI)
- Configuring MCP server for each assistant
- Setting up appropriate security profiles
- Testing the integration

Examples:
  portunix mcp init                                                    # Interactive wizard
  portunix mcp init --assistant claude-code                            # Claude Code with stdio (default)
  portunix mcp init --assistant claude-code --scope user               # Claude Code with user scope
  portunix mcp init --assistant claude-code --env KEY=VAL,FOO=bar      # Pass env vars to server
  portunix mcp init --assistant claude-code --timeout 60s              # Custom startup timeout
  portunix mcp init --from-json path/to/config.json                    # Import full configuration
  portunix mcp init --assistant claude-desktop --type remote           # Claude Desktop with remote server
  portunix mcp init --assistant gemini-cli --install-missing           # Install gemini-cli first if absent
  portunix mcp init --preset development                               # Use development preset`,
	Run: func(cmd *cobra.Command, args []string) {
		opts := initOptions{}
		opts.assistant, _ = cmd.Flags().GetString("assistant")
		opts.serverType, _ = cmd.Flags().GetString("type")
		opts.preset, _ = cmd.Flags().GetString("preset")
		opts.force, _ = cmd.Flags().GetBool("force")
		opts.scope, _ = cmd.Flags().GetString("scope")
		opts.timeout, _ = cmd.Flags().GetString("timeout")
		opts.fromJSON, _ = cmd.Flags().GetString("from-json")
		opts.bindAddress, _ = cmd.Flags().GetString("bind-address")
		opts.tlsCert, _ = cmd.Flags().GetString("tls-cert")
		opts.tlsKey, _ = cmd.Flags().GetString("tls-key")
		opts.installMissing, _ = cmd.Flags().GetBool("install-missing")

		envFlag, _ := cmd.Flags().GetString("env")
		envMap, err := parseEnvFlag(envFlag)
		if err != nil {
			fmt.Printf("❌ Invalid --env: %v\n", err)
			os.Exit(1)
		}
		opts.env = envMap

		// --from-json takes precedence over all other modes.
		if opts.fromJSON != "" {
			if err := runFromJSONConfiguration(opts.fromJSON, opts.force); err != nil {
				fmt.Printf("❌ Failed to import configuration: %v\n", err)
				os.Exit(1)
			}
			return
		}

		// Preset configuration
		if opts.preset != "" {
			if err := runPresetConfiguration(opts.preset, opts.force); err != nil {
				fmt.Printf("❌ Failed to configure with preset: %v\n", err)
				os.Exit(1)
			}
			return
		}

		// Non-interactive: assistant specified on command line
		if opts.assistant != "" {
			if err := runNonInteractiveConfiguration(opts); err != nil {
				fmt.Printf("❌ Failed to configure: %v\n", err)
				os.Exit(1)
			}
			return
		}

		// Otherwise, run interactive wizard
		if err := runInteractiveWizard(opts.force, opts.installMissing); err != nil {
			fmt.Printf("❌ Wizard failed: %v\n", err)
			os.Exit(1)
		}
	},
}

func runInteractiveWizard(force, installMissing bool) error {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("🚀 Welcome to Portunix MCP Server Setup Wizard")
	fmt.Println("=" + strings.Repeat("=", 50))
	fmt.Println()

	// Step 1: Detect existing configuration
	fmt.Println("📋 Step 1: Checking existing configuration...")
	config, err := loadMCPConfiguration()
	if err == nil && len(config.Assistants) > 0 && !force {
		fmt.Println("🔧 Found existing MCP server configuration:")
		for _, assistant := range config.Assistants {
			status := "✅"
			if !isAssistantInstalled(assistant.Name) {
				status = "⚠️ (not installed)"
			}
			fmt.Printf("   %s %s\n", status, getAssistantDisplayName(assistant.Name))
		}
		fmt.Println("\nYou can:")
		fmt.Println("1. Add another AI assistant to existing configuration")
		fmt.Println("2. Reconfigure existing setup")
		fmt.Println("3. Delete configuration")
		fmt.Println("4. Cancel")
		fmt.Print("\nChoice (1-4): ")
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(response)

		switch response {
		case "1":
			fmt.Println("Continuing to add new assistant...")
		case "2":
			fmt.Println("Reconfiguring existing setup...")
		case "3":
			fmt.Print("⚠️  This will delete MCP server configuration. Continue? (y/N): ")
			confirmResponse, _ := reader.ReadString('\n')
			confirmResponse = strings.ToLower(strings.TrimSpace(confirmResponse))

			if confirmResponse == "y" || confirmResponse == "yes" {
				fmt.Println("🗑️  Deleting configuration...")

				if err := removeMCPConfiguration(); err != nil {
					fmt.Printf("⚠️  Warning: Failed to remove Portunix configuration: %v\n", err)
				}

				if err := removeMCPServerFromClaudeCode(); err != nil {
					fmt.Printf("⚠️  Warning: Failed to remove from Claude Code: %v\n", err)
				}

				fmt.Println("✅ Configuration deleted. Starting fresh setup...")
				config = nil
			} else {
				fmt.Println("Configuration deletion cancelled.")
				return nil
			}
		case "4":
			fmt.Println("Configuration cancelled.")
			return nil
		default:
			fmt.Println("Invalid choice. Configuration cancelled.")
			return nil
		}
	}

	// Step 2: Detect installed AI assistants
	fmt.Println("\n📋 Step 2: Detecting installed AI assistants...")
	detectedAssistants := detectInstalledAssistants()

	if len(detectedAssistants) == 0 {
		fmt.Println("❌ No AI assistants detected.")
	} else {
		fmt.Println("✅ Detected assistants:")

		configuredMap := make(map[string]bool)
		if config != nil && err == nil && len(config.Assistants) > 0 {
			for _, assistant := range config.Assistants {
				configuredMap[assistant.Name] = true
			}
		}

		for i, assistant := range detectedAssistants {
			status := ""
			if configuredMap[assistant] {
				status = " (already configured)"
			}
			fmt.Printf("   %d. %s%s\n", i+1, getAssistantDisplayName(assistant), status)
		}

		// Auto-dependency hook (issue #035): when at least one assistant is
		// present but others are missing, offer to install the missing ones.
		if newly := offerInstallMissingAssistants(reader, installMissing); len(newly) > 0 {
			detectedAssistants = append(detectedAssistants, newly...)
		}
	}

	if len(detectedAssistants) == 0 {
		fmt.Println("❌ No AI assistants detected.")
		fmt.Println("\nWould you like to install an AI assistant?")
		fmt.Println("1. Claude Code (recommended for CLI development)")
		fmt.Println("2. Claude Desktop (for GUI experience)")
		fmt.Println("3. Skip installation")
		fmt.Print("\nChoice (1-3): ")

		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			fmt.Println("\n🚀 Installing Claude Code...")
			if err := installClaudeCode(); err != nil {
				return fmt.Errorf("failed to install Claude Code: %w", err)
			}
			detectedAssistants = append(detectedAssistants, "claude-code")
		case "2":
			fmt.Println("\n🚀 Installing Claude Desktop...")
			fmt.Println("Please visit: https://claude.ai/download")
			fmt.Println("After installation, run this wizard again.")
			return nil
		case "3":
			fmt.Println("Skipping installation. Run wizard again after installing an AI assistant.")
			return nil
		default:
			return fmt.Errorf("invalid choice")
		}
	}

	// Step 3: Select AI assistant to configure
	fmt.Println("\n📋 Step 3: Select AI assistant to configure:")

	configuredMap := make(map[string]bool)
	if config != nil && err == nil && len(config.Assistants) > 0 {
		for _, assistant := range config.Assistants {
			configuredMap[assistant.Name] = true
		}
	}

	for i, assistant := range detectedAssistants {
		status := ""
		if configuredMap[assistant] {
			status = " (reconfigure)"
		} else {
			status = " (new)"
		}
		fmt.Printf("%d. %s%s\n", i+1, getAssistantDisplayName(assistant), status)
	}

	var selectedAssistant string
	if len(detectedAssistants) == 1 {
		selectedAssistant = detectedAssistants[0]
		fmt.Printf("\nSelected: %s\n", getAssistantDisplayName(selectedAssistant))
	} else {
		fmt.Printf("Choice (1-%d): ", len(detectedAssistants))

		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)
		idx := 0
		fmt.Sscanf(choice, "%d", &idx)

		if idx < 1 || idx > len(detectedAssistants) {
			return fmt.Errorf("invalid choice")
		}
		selectedAssistant = detectedAssistants[idx-1]
	}

	if configuredMap[selectedAssistant] {
		fmt.Printf("🔄 Reconfiguring %s...\n", getAssistantDisplayName(selectedAssistant))
	} else {
		fmt.Printf("🆕 Configuring %s for the first time...\n", getAssistantDisplayName(selectedAssistant))
	}

	// Step 4: Configure server type based on assistant
	fmt.Println("\n📋 Step 4: Configuring server type...")
	serverType := getDefaultServerType(selectedAssistant)

	opts := initOptions{
		assistant: selectedAssistant,
		force:     force,
	}

	switch selectedAssistant {
	case "claude-code":
		fmt.Println("Claude Code uses stdio communication by default.")
		serverType = "stdio"
		promptClaudeCodeOptions(reader, &opts)
	case "claude-desktop":
		fmt.Println("Claude Desktop requires remote server configuration.")
		serverType = "remote"
		if err := promptRemoteOptions(reader, &opts, 3002); err != nil {
			return err
		}
	default:
		// Other assistants — keep server type from getDefaultServerType.
	}
	opts.serverType = serverType

	// Step 5: Configure security profile
	fmt.Println("\n📋 Step 5: Select security profile:")
	fmt.Println("1. Development (CLI tools, local development)")
	fmt.Println("2. Standard (Desktop apps, full integration)")
	fmt.Println("3. Restricted (minimal permissions)")
	fmt.Print("Choice (1-3, default 1): ")

	secChoice, _ := reader.ReadString('\n')
	secChoice = strings.TrimSpace(secChoice)
	if secChoice == "" {
		secChoice = "1"
	}

	var securityProfile string
	switch secChoice {
	case "1":
		securityProfile = "development"
	case "2":
		securityProfile = "standard"
	case "3":
		securityProfile = "restricted"
	default:
		securityProfile = "development"
	}

	// Step 6: Apply configuration
	fmt.Println("\n📋 Step 6: Applying configuration...")
	if err := applyAssistantConfiguration(selectedAssistant, serverType, securityProfile, opts); err != nil {
		return fmt.Errorf("failed to apply configuration: %w", err)
	}

	// Step 7: Test connection
	fmt.Println("\n📋 Step 7: Testing connection...")
	if err := testAssistantConnection(selectedAssistant); err != nil {
		fmt.Printf("⚠️  Connection test failed: %v\n", err)
		fmt.Println("You may need to start the MCP server manually.")
	} else {
		fmt.Println("✅ Connection test successful!")
	}

	// Final instructions
	fmt.Println("\n🎉 Configuration complete!")
	fmt.Println("\nNext steps:")
	if serverType == "stdio" {
		fmt.Printf("1. The MCP server will start automatically when you use %s\n", getAssistantDisplayName(selectedAssistant))
		fmt.Printf("2. Open %s and test the integration\n", getAssistantDisplayName(selectedAssistant))
	} else {
		fmt.Println("1. Start the MCP server: portunix mcp start")
		fmt.Printf("2. Open %s and test the integration\n", getAssistantDisplayName(selectedAssistant))
	}

	return nil
}

func runNonInteractiveConfiguration(opts initOptions) error {
	assistant := opts.assistant
	fmt.Printf("🔧 Configuring MCP server for %s...\n", getAssistantDisplayName(assistant))

	if isMCPAlreadyConfigured() && !opts.force {
		return fmt.Errorf("MCP server already configured. Use --force to reconfigure")
	}

	if !isAssistantInstalled(assistant) {
		if opts.installMissing {
			fmt.Printf("📦 %s not installed; installing via ptx-installer...\n", getAssistantDisplayName(assistant))
			if err := installAssistant(assistant); err != nil {
				return fmt.Errorf("failed to install %s: %w", getAssistantDisplayName(assistant), err)
			}
			if !isAssistantInstalled(assistant) {
				return fmt.Errorf("%s still not detected after installation", getAssistantDisplayName(assistant))
			}
		} else {
			fmt.Printf("❌ %s is not installed\n", getAssistantDisplayName(assistant))
			fmt.Println("Install it first, pass --install-missing, or run the interactive wizard for installation help")
			return fmt.Errorf("assistant not installed")
		}
	}

	serverType := opts.serverType
	if serverType == "" {
		serverType = getDefaultServerType(assistant)
	}

	if opts.timeout != "" {
		if _, err := time.ParseDuration(opts.timeout); err != nil {
			return fmt.Errorf("invalid --timeout %q: %w", opts.timeout, err)
		}
	}

	securityProfile := getDefaultSecurityProfile(assistant)
	if err := applyAssistantConfiguration(assistant, serverType, securityProfile, opts); err != nil {
		return err
	}

	fmt.Println("✅ Configuration applied successfully!")
	return nil
}

// runFromJSONConfiguration loads a complete MCPConfiguration document from a
// JSON file, validates it, persists it and applies it to each configured
// assistant.
func runFromJSONConfiguration(path string, force bool) error {
	fmt.Printf("📥 Importing MCP configuration from %s...\n", path)
	config, err := loadMCPConfigurationFromFile(path)
	if err != nil {
		return err
	}

	if isMCPConfigurationExists() && !force {
		return fmt.Errorf("MCP configuration already exists; use --force to overwrite")
	}

	if err := saveMCPConfiguration(config); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}
	fmt.Println("✅ Configuration written to", getMCPConfigFile())

	opts := initOptions{
		scope:       config.Scope,
		env:         config.Env,
		timeout:     config.Timeout,
		bindAddress: config.BindAddress,
		tlsCert:     config.TLSCert,
		tlsKey:      config.TLSKey,
	}
	var failed []string
	for _, a := range config.Assistants {
		serverType := a.ServerType
		if serverType == "" {
			serverType = config.ServerType
		}
		fmt.Printf("🔧 Applying %s...\n", getAssistantDisplayName(a.Name))
		if err := applyAssistantToTool(a.Name, serverType, config.SecurityProfile, opts, config); err != nil {
			fmt.Printf("⚠️  %s: %v\n", a.Name, err)
			failed = append(failed, a.Name)
		}
	}
	if len(failed) > 0 {
		return fmt.Errorf("import finished with %d failure(s): %s", len(failed), strings.Join(failed, ", "))
	}
	fmt.Println("✅ Import complete")
	return nil
}

func runPresetConfiguration(preset string, force bool) error {
	fmt.Printf("🔧 Applying preset configuration: %s\n", preset)

	switch preset {
	case "development":
		return runNonInteractiveConfiguration(initOptions{
			assistant:  "claude-code",
			serverType: "stdio",
			force:      force,
		})
	case "standard":
		return runNonInteractiveConfiguration(initOptions{
			assistant:  "claude-desktop",
			serverType: "remote",
			force:      force,
		})
	default:
		return fmt.Errorf("unknown preset: %s", preset)
	}
}

// installClaudeCode delegates claude-code installation to the ptx-installer
// helper (preferred path) and falls back to direct npm / curl if the helper
// is not co-located on disk. Issue #186b removed the direct
// portunix.ai/app/install import so ptx-mcp is now decoupled from the legacy
// installation subsystem.
func installClaudeCode() error {
	if ptxInstaller, ok := findPtxInstaller(); ok {
		cmd := exec.Command(ptxInstaller, "install", "claude-code")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		if err := cmd.Run(); err == nil {
			return nil
		}
		fmt.Println("ptx-installer failed, falling back to npm...")
	} else {
		fmt.Println("ptx-installer not found next to ptx-mcp, falling back to npm...")
	}

	cmd := exec.Command("npm", "install", "-g", "@anthropic-ai/claude-code")
	if err := cmd.Run(); err != nil {
		fmt.Println("npm installation failed, trying curl method...")
		installScript := "curl -fsSL https://claude.ai/cli/install.sh | sh"
		cmd = exec.Command("sh", "-c", installScript)
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("installation failed")
		}
	}
	return nil
}

// findPtxInstaller locates the ptx-installer helper binary that should sit
// next to this helper on disk. Returns the absolute path and true on
// success; empty string and false otherwise.
func findPtxInstaller() (string, bool) {
	exe, err := os.Executable()
	if err != nil {
		return "", false
	}
	dir := filepath.Dir(exe)
	name := "ptx-installer"
	if filepath.Ext(exe) == ".exe" {
		name += ".exe"
	}
	candidate := filepath.Join(dir, name)
	info, err := os.Stat(candidate)
	if err != nil || info.IsDir() {
		return "", false
	}
	return candidate, true
}

// installAssistant installs the named AI assistant via the co-located
// ptx-installer helper (issue #035 auto-dependency hook). claude-code keeps its
// npm/curl fallback (installClaudeCode) for environments without ptx-installer;
// other assistants require ptx-installer to be present.
func installAssistant(name string) error {
	if name == "claude-code" {
		return installClaudeCode()
	}
	ptxInstaller, ok := findPtxInstaller()
	if !ok {
		return fmt.Errorf("ptx-installer helper not found next to ptx-mcp; install %s manually", name)
	}
	cmd := exec.Command(ptxInstaller, "install", name)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// offerInstallMissingAssistants surfaces known AI assistants that are not yet
// installed and installs the selected ones via ptx-installer (issue #035
// auto-dependency hook). When autoYes is true the prompt is skipped (the
// non-interactive --install-missing path). Returns the names that were
// successfully installed.
func offerInstallMissingAssistants(reader *bufio.Reader, autoYes bool) []string {
	missing := missingAssistants()
	if len(missing) == 0 {
		return nil
	}

	fmt.Println("\n🔎 Some AI assistants are not installed:")
	for _, name := range missing {
		fmt.Printf("   ❌ %s\n", getAssistantDisplayName(name))
	}

	if !autoYes {
		fmt.Print("\nInstall the missing assistant(s) now? [y/N]: ")
		ans, _ := reader.ReadString('\n')
		ans = strings.ToLower(strings.TrimSpace(ans))
		if ans != "y" && ans != "yes" {
			return nil
		}
	}

	var installed []string
	for _, name := range missing {
		fmt.Printf("\n🚀 Installing %s...\n", getAssistantDisplayName(name))
		if err := installAssistant(name); err != nil {
			fmt.Printf("⚠️  Failed to install %s: %v\n", getAssistantDisplayName(name), err)
			continue
		}
		if isAssistantInstalled(name) {
			installed = append(installed, name)
		}
	}
	return installed
}

func applyAssistantConfiguration(assistant, serverType, securityProfile string, opts initOptions) error {
	fmt.Printf("📝 Configuring %s with %s server and %s security profile...\n",
		getAssistantDisplayName(assistant), serverType, securityProfile)

	config, err := loadMCPConfiguration()
	if err != nil {
		config = &MCPConfiguration{
			ServerType:      serverType,
			SecurityProfile: securityProfile,
			Assistants:      []AssistantConfig{},
		}

		if serverType == "remote" {
			switch assistant {
			case "claude-code":
				config.Port = 3001
			case "claude-desktop":
				config.Port = 3002
			case "gemini-cli":
				config.Port = 3003
			default:
				config.Port = 3001
			}
			config.Protocol = "https"
		}
	}

	// Merge advanced options into the persisted configuration.
	if opts.scope != "" {
		config.Scope = opts.scope
	}
	if opts.timeout != "" {
		config.Timeout = opts.timeout
	}
	if len(opts.env) > 0 {
		if config.Env == nil {
			config.Env = make(map[string]string)
		}
		for k, v := range opts.env {
			config.Env[k] = v
		}
	}
	if opts.bindAddress != "" {
		config.BindAddress = opts.bindAddress
	}
	if opts.tlsCert != "" {
		config.TLSCert = opts.tlsCert
	}
	if opts.tlsKey != "" {
		config.TLSKey = opts.tlsKey
	}

	found := false
	for i, a := range config.Assistants {
		if a.Name == assistant {
			config.Assistants[i].ServerType = serverType
			config.Assistants[i].Configured = true
			found = true
			break
		}
	}
	if !found {
		config.Assistants = append(config.Assistants, AssistantConfig{
			Name:       assistant,
			ServerType: serverType,
			Configured: true,
		})
	}

	if err := saveMCPConfiguration(config); err != nil {
		return fmt.Errorf("failed to save configuration: %w", err)
	}

	return applyAssistantToTool(assistant, serverType, securityProfile, opts, config)
}

// applyAssistantToTool invokes the per-assistant integration step.
func applyAssistantToTool(assistant, serverType, securityProfile string, opts initOptions, config *MCPConfiguration) error {
	switch assistant {
	case "claude-code":
		return configureClaudeCode(serverType, opts)
	case "claude-desktop":
		return configureClaudeDesktop(serverType, opts, config)
	case "gemini-cli":
		return configureGeminiCLI(serverType)
	default:
		fmt.Printf("⚠️  Assistant-specific configuration for %s not implemented yet\n", assistant)
		return nil
	}
}

func configureClaudeCode(serverType string, opts initOptions) error {
	fmt.Println("   🔧 Configuring Claude Code MCP integration...")
	scope := opts.scope
	if scope == "" {
		scope = "local"
	}
	mode := serverType
	if mode == "" || mode == "remote" {
		mode = "stdio"
	}
	return configureClaudeCodeMCP(mode, scope, 3001, "standard", true, opts.env, opts.timeout)
}

// configureClaudeDesktop writes (or merges into) the platform-specific
// mcp_servers.json so that Claude Desktop picks up Portunix without manual
// editing. Existing entries are preserved.
func configureClaudeDesktop(serverType string, opts initOptions, config *MCPConfiguration) error {
	fmt.Println("   🔧 Configuring Claude Desktop MCP integration...")

	configPath := getClaudeDesktopConfigPath()
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Load existing configuration (if any).
	existing := map[string]interface{}{}
	if data, err := os.ReadFile(configPath); err == nil && len(data) > 0 {
		if err := json.Unmarshal(data, &existing); err != nil {
			return fmt.Errorf("existing %s is not valid JSON: %w", configPath, err)
		}
	} else if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to read %s: %w", configPath, err)
	}

	// The schema used by Claude Desktop nests servers under "servers".
	serversAny, ok := existing["servers"]
	servers := map[string]interface{}{}
	if ok {
		if m, ok := serversAny.(map[string]interface{}); ok {
			servers = m
		}
	}

	// Build the portunix entry.
	portunixPath, err := getPortunixExecutablePath()
	if err != nil {
		return fmt.Errorf("portunix executable not found: %w", err)
	}
	args := []string{"mcp", "serve"}
	if serverType == "remote" && config != nil {
		args = append(args, "--mode", "tcp", "--port", fmt.Sprintf("%d", config.Port))
	} else {
		args = append(args, "--mode", "stdio")
	}
	entry := map[string]interface{}{
		"command": portunixPath,
		"args":    args,
	}
	if len(opts.env) > 0 {
		entry["env"] = opts.env
	}
	servers["portunix"] = entry
	existing["servers"] = servers

	// Write back preserving indentation.
	out, err := json.MarshalIndent(existing, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal configuration: %w", err)
	}
	if err := os.WriteFile(configPath, out, 0o644); err != nil {
		return fmt.Errorf("failed to write %s: %w", configPath, err)
	}

	// Verify by re-reading.
	verify, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("verification read failed: %w", err)
	}
	var roundTrip map[string]interface{}
	if err := json.Unmarshal(verify, &roundTrip); err != nil {
		return fmt.Errorf("written file is not valid JSON: %w", err)
	}
	if rtServers, ok := roundTrip["servers"].(map[string]interface{}); !ok || rtServers["portunix"] == nil {
		return fmt.Errorf("portunix entry missing after write")
	}

	fmt.Printf("   ✅ Updated %s\n", configPath)
	return nil
}

func configureGeminiCLI(serverType string) error {
	fmt.Println("   🔧 Configuring Gemini CLI MCP integration...")
	fmt.Println("   ⚠️  Gemini CLI MCP integration is experimental")
	fmt.Println("   Configuration method needs to be researched")
	return nil
}

// generateSelfSignedCert creates a development self-signed certificate at the
// given paths. Returns paths to the generated cert and key files.
func generateSelfSignedCert(certPath, keyPath string) error {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return fmt.Errorf("generate key: %w", err)
	}

	serial, err := rand.Int(rand.Reader, big.NewInt(1<<62))
	if err != nil {
		return fmt.Errorf("generate serial: %w", err)
	}

	template := x509.Certificate{
		SerialNumber: serial,
		Subject:      pkix.Name{CommonName: "portunix-mcp-dev"},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(365 * 24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1"), net.ParseIP("::1")},
		DNSNames:     []string{"localhost"},
	}
	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &priv.PublicKey, priv)
	if err != nil {
		return fmt.Errorf("create cert: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(certPath), 0o755); err != nil {
		return fmt.Errorf("create cert dir: %w", err)
	}
	certOut, err := os.Create(certPath)
	if err != nil {
		return fmt.Errorf("open cert file: %w", err)
	}
	defer certOut.Close()
	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return fmt.Errorf("write cert: %w", err)
	}

	keyOut, err := os.OpenFile(keyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("open key file: %w", err)
	}
	defer keyOut.Close()
	privBytes, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return fmt.Errorf("marshal key: %w", err)
	}
	if err := pem.Encode(keyOut, &pem.Block{Type: "PRIVATE KEY", Bytes: privBytes}); err != nil {
		return fmt.Errorf("write key: %w", err)
	}
	return nil
}

func testAssistantConnection(assistant string) error {
	fmt.Printf("Testing connection with %s...\n", getAssistantDisplayName(assistant))
	config, err := loadMCPConfiguration()
	if err != nil {
		return fmt.Errorf("no configuration loaded: %w", err)
	}
	result, err := handshakeAssistant(assistant, config)
	if err != nil {
		return err
	}
	fmt.Printf("   protocol=%s", result.ProtocolVersion)
	if result.ServerName != "" {
		fmt.Printf(", server=%s/%s", result.ServerName, result.ServerVersion)
	}
	fmt.Println()
	return nil
}

// promptClaudeCodeOptions asks for Claude Code-specific options during the
// interactive wizard: configuration scope, env vars and startup timeout.
func promptClaudeCodeOptions(reader *bufio.Reader, opts *initOptions) {
	fmt.Print("\nClaude Code scope (local/project/user, default local): ")
	scope, _ := reader.ReadString('\n')
	scope = strings.TrimSpace(scope)
	if scope == "" {
		scope = "local"
	}
	switch scope {
	case "local", "project", "user":
		opts.scope = scope
	default:
		fmt.Printf("⚠️  Invalid scope %q, falling back to 'local'\n", scope)
		opts.scope = "local"
	}

	fmt.Print("Environment variables (KEY=VAL,KEY2=VAL2 or empty): ")
	envStr, _ := reader.ReadString('\n')
	envStr = strings.TrimSpace(envStr)
	if envStr != "" {
		envMap, err := parseEnvFlag(envStr)
		if err != nil {
			fmt.Printf("⚠️  Invalid env (%v) — skipped\n", err)
		} else {
			opts.env = envMap
		}
	}

	fmt.Print("Startup timeout (e.g. 30s, 2m, default empty): ")
	timeout, _ := reader.ReadString('\n')
	timeout = strings.TrimSpace(timeout)
	if timeout != "" {
		if _, err := time.ParseDuration(timeout); err != nil {
			fmt.Printf("⚠️  Invalid timeout %q — skipped\n", timeout)
		} else {
			opts.timeout = timeout
		}
	}
}

// promptRemoteOptions asks for remote MCP server options: protocol, port,
// bind address and (when applicable) TLS certificate and key paths.
func promptRemoteOptions(reader *bufio.Reader, opts *initOptions, defaultPort int) error {
	suggestions := findAvailablePorts(3)
	if len(suggestions) > 0 {
		fmt.Printf("\nSuggested free ports: %v\n", suggestions)
	}

	fmt.Printf("Port (default %d): ", defaultPort)
	portStr, _ := reader.ReadString('\n')
	portStr = strings.TrimSpace(portStr)
	if portStr == "" {
		portStr = fmt.Sprintf("%d", defaultPort)
	}
	// Port is also tracked in the existing config; we just validate here.
	var port int
	if _, err := fmt.Sscanf(portStr, "%d", &port); err != nil || port <= 0 || port > 65535 {
		return fmt.Errorf("invalid port %q", portStr)
	}

	fmt.Println("Protocol options: http, https, ws, wss")
	fmt.Print("Protocol (default https): ")
	protocol, _ := reader.ReadString('\n')
	protocol = strings.TrimSpace(protocol)
	if protocol == "" {
		protocol = "https"
	}
	if !isValidProtocol(protocol) {
		return fmt.Errorf("invalid protocol %q", protocol)
	}

	fmt.Print("Bind address (localhost, 0.0.0.0 or IP, default localhost): ")
	bind, _ := reader.ReadString('\n')
	bind = strings.TrimSpace(bind)
	if bind == "" {
		bind = "localhost"
	}
	if !isValidBindAddress(bind) {
		return fmt.Errorf("invalid bind address %q", bind)
	}
	opts.bindAddress = bind

	if protocol == "https" || protocol == "wss" {
		fmt.Print("TLS certificate path (empty to generate self-signed for dev): ")
		certPath, _ := reader.ReadString('\n')
		certPath = strings.TrimSpace(certPath)
		fmt.Print("TLS key path: ")
		keyPath, _ := reader.ReadString('\n')
		keyPath = strings.TrimSpace(keyPath)

		if certPath == "" && keyPath == "" {
			home, _ := os.UserHomeDir()
			certPath = filepath.Join(home, ".portunix", "tls", "mcp-dev.crt")
			keyPath = filepath.Join(home, ".portunix", "tls", "mcp-dev.key")
			fmt.Printf("Generating self-signed certificate at %s ...\n", certPath)
			if err := generateSelfSignedCert(certPath, keyPath); err != nil {
				return fmt.Errorf("self-signed cert generation: %w", err)
			}
			fmt.Println("⚠️  Self-signed certificate — not for production use.")
		} else if certPath == "" || keyPath == "" {
			return fmt.Errorf("both --tls-cert and --tls-key must be provided")
		} else {
			if _, err := os.Stat(certPath); err != nil {
				return fmt.Errorf("tls cert not readable: %w", err)
			}
			if _, err := os.Stat(keyPath); err != nil {
				return fmt.Errorf("tls key not readable: %w", err)
			}
		}
		opts.tlsCert = certPath
		opts.tlsKey = keyPath
	}

	// Protocol and port are not part of initOptions; persist them directly
	// to MCPConfiguration so applyAssistantConfiguration can pick them up.
	overrideRemoteProtocol(protocol, port)
	return nil
}

// overrideRemoteProtocol patches the persisted MCPConfiguration with the
// prompted protocol/port without losing other fields. Called from
// promptRemoteOptions because initOptions does not carry a protocol field.
func overrideRemoteProtocol(protocol string, port int) {
	config, err := loadMCPConfiguration()
	if err != nil {
		config = &MCPConfiguration{}
	}
	config.Protocol = protocol
	config.Port = port
	_ = saveMCPConfiguration(config)
}

// configureClaudeCodeMCP runs the existing claude mcp add integration with
// support for the new env / timeout passthrough. Lives here (rather than in
// configure.go) so that init.go's wizard path can call it without circular
// dependencies on opts shape.
func configureClaudeCodeMCP(mode, scope string, port int, permissions string, force bool, env map[string]string, timeout string) error {
	return configureMCPIntegrationExt(mode, scope, port, permissions, force, env, timeout)
}

func init() {
	mcpCmd.AddCommand(initCmd)

	initCmd.Flags().String("assistant", "", "AI assistant to configure (claude-code, claude-desktop, gemini-cli)")
	initCmd.Flags().String("type", "", "Server type (stdio, remote)")
	initCmd.Flags().String("preset", "", "Use preset configuration (development, standard)")
	initCmd.Flags().Bool("force", false, "Force reconfiguration even if already configured")
	initCmd.Flags().String("scope", "", "Claude Code configuration scope (local, project, user)")
	initCmd.Flags().String("env", "", "Environment variables passed to MCP server (KEY=VAL,KEY2=VAL2)")
	initCmd.Flags().String("timeout", "", "Startup timeout for MCP server (e.g. 30s, 2m)")
	initCmd.Flags().String("from-json", "", "Path to a JSON configuration file to import")
	initCmd.Flags().String("bind-address", "", "Bind address for remote MCP server (localhost, 0.0.0.0, or IP)")
	initCmd.Flags().String("tls-cert", "", "Path to TLS certificate for https/wss")
	initCmd.Flags().String("tls-key", "", "Path to TLS key for https/wss")
	initCmd.Flags().Bool("install-missing", false, "Auto-install missing AI assistants via ptx-installer")
}
