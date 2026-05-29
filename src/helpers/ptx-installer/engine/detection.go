/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package engine

import (
	"os/exec"
	"regexp"
	"runtime"
	"sort"
	"strings"

	"portunix.ai/portunix/src/helpers/ptx-installer/registry"
)

// aiAssistantCategory is the registry category that groups installable AI
// assistants (claude-code, claude-desktop, gemini-cli, ...)
const aiAssistantCategory = "development/ai-tools"

// AssistantStatus is the detection result for a single AI assistant package
type AssistantStatus struct {
	Name        string `json:"name"`
	DisplayName string `json:"displayName"`
	Category    string `json:"category"`
	Installed   bool   `json:"installed"`
	Version     string `json:"version,omitempty"`
	// VerifyCommand is the platform verification command used for detection
	VerifyCommand string `json:"verifyCommand,omitempty"`
}

// versionPattern matches a dotted version token such as 1.2 or 10.4.3
var versionPattern = regexp.MustCompile(`\d+\.\d+(?:\.\d+)?`)

// selectAIAssistants returns the installable AI assistant packages (Kind
// "Package", category development/ai-tools), excluding bundles, sorted by name
func selectAIAssistants(pkgs map[string]*registry.Package) []*registry.Package {
	out := make([]*registry.Package, 0)
	for _, pkg := range pkgs {
		if pkg.Kind != "Package" {
			continue // skip bundles and other kinds
		}
		if pkg.Metadata.Category != aiAssistantCategory {
			continue
		}
		out = append(out, pkg)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].Metadata.Name < out[j].Metadata.Name
	})
	return out
}

// verifyCommandForOS returns the platform verification command for the given
// OS, with the windows_sandbox -> windows fallback used elsewhere in the
// installer. Returns "" when no command is defined for the platform
func verifyCommandForOS(pkg *registry.Package, osName string) string {
	plat, ok := pkg.Spec.Platforms[osName]
	if !ok && osName == "windows_sandbox" {
		plat, ok = pkg.Spec.Platforms["windows"]
	}
	if !ok {
		return ""
	}
	if plat.Verification != nil {
		return plat.Verification.Command
	}
	return ""
}

// parseVersionOutput extracts the first dotted version token from a command's
// output. Returns "" when the output carries no version
func parseVersionOutput(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	return versionPattern.FindString(raw)
}

// isPathLookup reports whether the verification command is a bare path lookup
// (where/which) whose output is a file path, not a version string
func isPathLookup(command string) bool {
	c := strings.ToLower(strings.TrimSpace(command))
	return strings.HasPrefix(c, "where ") || strings.HasPrefix(c, "which ")
}

// runVerification executes a verification command and reports whether it
// succeeded (exit code 0) plus any version parsed from its output
func runVerification(command string) (installed bool, version string) {
	if strings.TrimSpace(command) == "" {
		return false, ""
	}

	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", command)
	} else {
		cmd = exec.Command("sh", "-c", command)
	}

	out, err := cmd.CombinedOutput()
	if err != nil {
		return false, ""
	}

	// where/which only confirm presence; their output is a path, not a version
	if isPathLookup(command) {
		return true, ""
	}
	return true, parseVersionOutput(string(out))
}

// DetectAIAssistants detects which AI assistant packages are installed on the
// given OS by running each package's platform verification command. Packages
// without a verification command for the OS are reported as not installed
func DetectAIAssistants(reg *registry.PackageRegistry, osName string) []AssistantStatus {
	assistants := selectAIAssistants(reg.GetAllPackages())
	results := make([]AssistantStatus, 0, len(assistants))

	for _, pkg := range assistants {
		command := verifyCommandForOS(pkg, osName)
		status := AssistantStatus{
			Name:          pkg.Metadata.Name,
			DisplayName:   pkg.Metadata.DisplayName,
			Category:      pkg.Metadata.Category,
			VerifyCommand: command,
		}
		if command != "" {
			status.Installed, status.Version = runVerification(command)
		}
		results = append(results, status)
	}

	return results
}
