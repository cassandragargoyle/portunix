package install

import (
	"fmt"
	"strings"
)

// PreMCPSetupHook runs before MCP server setup to ensure required AI assistants are installed
func PreMCPSetupHook(selectedAssistants []string) error {
	fmt.Println("🔍 Checking AI assistant dependencies for MCP setup...")

	missing := DetectMissingAssistants(selectedAssistants)
	if len(missing) > 0 {
		return PromptInstallMissing(missing)
	}

	fmt.Println("✅ All required AI assistants are installed")
	return nil
}

// ValidateMCPDependencies validates that all necessary AI assistants are installed for MCP
func ValidateMCPDependencies() error {
	requiredAssistants := []string{"claude-code"}
	recommendedAssistants := GetRecommendedAIAssistants()

	fmt.Println("════════════════════════════════════════")
	fmt.Println("🔍 MCP DEPENDENCY CHECK")
	fmt.Println("════════════════════════════════════════")

	// Check required assistants
	missing := DetectMissingAssistants(requiredAssistants)
	if len(missing) > 0 {
		fmt.Printf("❌ Required AI assistants missing: %s\n", strings.Join(missing, ", "))
		return PromptInstallMissing(missing)
	}

	// Check recommended assistants
	recommendedMissing := DetectMissingAssistants(recommendedAssistants)
	if len(recommendedMissing) > 0 {
		fmt.Printf("💡 Recommended AI assistants missing: %s\n", strings.Join(recommendedMissing, ", "))
		fmt.Print("Install recommended assistants for better MCP experience? [Y/n]: ")

		var response string
		fmt.Scanln(&response)

		if response == "" || strings.ToLower(response) == "y" || strings.ToLower(response) == "yes" {
			return PromptInstallMissing(recommendedMissing)
		}
	}

	fmt.Println("✅ MCP dependencies satisfied")
	return nil
}

// GetMCPCompatibleAssistants returns AI assistants that are compatible with MCP
func GetMCPCompatibleAssistants() []string {
	return []string{
		"claude-code",
		"claude-desktop",
		"gemini-cli",
	}
}

// InstallMCPReadyEnvironment installs a complete MCP-ready development environment
func InstallMCPReadyEnvironment() error {
	fmt.Println("🚀 Installing MCP-ready development environment...")

	// Install the mcp-ready preset
	if err := InstallPreset("mcp-ready"); err != nil {
		return fmt.Errorf("failed to install mcp-ready preset: %w", err)
	}

	// Validate installation
	if err := ValidateMCPDependencies(); err != nil {
		return fmt.Errorf("MCP dependency validation failed: %w", err)
	}

	fmt.Println("✅ MCP-ready environment installed successfully!")
	return nil
}

// RecommendAIAssistants analyzes current setup and recommends AI assistants
func RecommendAIAssistants() {
	fmt.Println("════════════════════════════════════════")
	fmt.Println("💡 AI ASSISTANT RECOMMENDATIONS")
	fmt.Println("════════════════════════════════════════")

	statuses := DetectAIAssistants()
	recommendations := GetRecommendedAIAssistants()

	fmt.Println("Current AI assistant status:")
	for _, status := range statuses {
		if status.Installed {
			fmt.Printf("  ✅ %s - Already installed", status.Name)
			if status.Version != "" {
				fmt.Printf(" (v%s)", status.Version)
			}
			fmt.Println()
		} else {
			fmt.Printf("  ❌ %s - Not installed", status.Name)
			if contains(recommendations, status.Name) {
				fmt.Print(" (recommended)")
			}
			fmt.Println()
		}
	}

	fmt.Println()
	fmt.Println("Recommended installation commands:")
	for _, rec := range recommendations {
		isInstalled := false
		for _, status := range statuses {
			if status.Name == rec && status.Installed {
				isInstalled = true
				break
			}
		}

		if !isInstalled {
			fmt.Printf("  portunix install %s\n", rec)
		}
	}

	fmt.Println()
	fmt.Println("Or install all recommended assistants at once:")
	fmt.Println("  portunix install ai-assistant-full")
	fmt.Println("════════════════════════════════════════")
}

// Helper function to check if slice contains string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}