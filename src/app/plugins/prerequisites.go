/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package plugins

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"portunix.ai/app/version"
)

// CheckStatus represents the result status of a single prerequisite check
type CheckStatus string

const (
	CheckStatusOK      CheckStatus = "ok"
	CheckStatusWarning CheckStatus = "warning"
	CheckStatusError   CheckStatus = "error"
)

// RuntimeCheckResult holds the result of a runtime availability check
type RuntimeCheckResult struct {
	Status   CheckStatus `json:"status"`
	Required string      `json:"required"`
	Found    string      `json:"found,omitempty"`
	Message  string      `json:"message,omitempty"`
	Fix      string      `json:"fix,omitempty"`
}

// VersionCheckResult holds the result of a version check
type VersionCheckResult struct {
	Status   CheckStatus `json:"status"`
	Required string      `json:"required"`
	Found    string      `json:"found"`
	Message  string      `json:"message,omitempty"`
}

// OSCheckResult holds the result of an OS compatibility check
type OSCheckResult struct {
	Status   CheckStatus `json:"status"`
	Required []string    `json:"required"`
	Current  string      `json:"current"`
	Message  string      `json:"message,omitempty"`
}

// ToolCheckResult holds the result of an optional tool check
type ToolCheckResult struct {
	Name    string      `json:"name"`
	Status  CheckStatus `json:"status"`
	Reason  string      `json:"reason,omitempty"`
	Message string      `json:"message,omitempty"`
	Fix     string      `json:"fix,omitempty"`
}

// PrerequisiteCheckResult holds the complete result of all prerequisite checks
type PrerequisiteCheckResult struct {
	PluginName      string              `json:"plugin"`
	PluginVersion   string              `json:"version"`
	Runtime         *RuntimeCheckResult `json:"runtime,omitempty"`
	PortunixVersion *VersionCheckResult `json:"portunix_version,omitempty"`
	OSSupport       *OSCheckResult      `json:"os_support,omitempty"`
	OptionalTools   []ToolCheckResult   `json:"optional_tools,omitempty"`
	Overall         CheckStatus         `json:"overall"`
}

// PrerequisitesSummary holds aggregated results for multiple plugins
type PrerequisitesSummary struct {
	Results []PrerequisiteCheckResult `json:"results"`
	Summary struct {
		Total   int `json:"total"`
		OK      int `json:"ok"`
		Warning int `json:"warning"`
		Error   int `json:"error"`
	} `json:"summary"`
}

// CheckPrerequisites validates all prerequisites for a plugin manifest
func CheckPrerequisites(manifest *PluginManifest) *PrerequisiteCheckResult {
	result := &PrerequisiteCheckResult{
		PluginName:    manifest.Name,
		PluginVersion: manifest.Version,
		Overall:       CheckStatusOK,
	}

	// Check runtime
	result.Runtime = checkRuntime(manifest.Plugin.Runtime, manifest.Plugin.RuntimeVersion)
	if result.Runtime.Status == CheckStatusError {
		result.Overall = CheckStatusError
	}

	// Check Portunix version
	result.PortunixVersion = checkPortunixVersion(manifest.Dependencies.PortunixMinVersion)
	if result.PortunixVersion.Status == CheckStatusError {
		result.Overall = CheckStatusError
	}

	// Check OS support
	result.OSSupport = checkOSSupport(manifest.Dependencies.OSSupport)
	if result.OSSupport.Status == CheckStatusError {
		result.Overall = CheckStatusError
	}

	// Check optional tools
	for _, tool := range manifest.Dependencies.OptionalTools {
		toolResult := checkOptionalTool(tool)
		result.OptionalTools = append(result.OptionalTools, toolResult)
		if toolResult.Status == CheckStatusWarning && result.Overall == CheckStatusOK {
			result.Overall = CheckStatusWarning
		}
	}

	return result
}

// CheckPrerequisitesFromRegistry validates prerequisites using registry data
func CheckPrerequisitesFromRegistry(name, ver, runtimeType, runtimeVer, portunixMinVer string, osSupport []string, optionalTools []OptionalTool) *PrerequisiteCheckResult {
	result := &PrerequisiteCheckResult{
		PluginName:    name,
		PluginVersion: ver,
		Overall:       CheckStatusOK,
	}

	result.Runtime = checkRuntime(runtimeType, runtimeVer)
	if result.Runtime.Status == CheckStatusError {
		result.Overall = CheckStatusError
	}

	result.PortunixVersion = checkPortunixVersion(portunixMinVer)
	if result.PortunixVersion.Status == CheckStatusError {
		result.Overall = CheckStatusError
	}

	result.OSSupport = checkOSSupport(osSupport)
	if result.OSSupport.Status == CheckStatusError {
		result.Overall = CheckStatusError
	}

	for _, tool := range optionalTools {
		toolResult := checkOptionalTool(tool)
		result.OptionalTools = append(result.OptionalTools, toolResult)
		if toolResult.Status == CheckStatusWarning && result.Overall == CheckStatusOK {
			result.Overall = CheckStatusWarning
		}
	}

	return result
}

// HasErrors returns true if the check result contains any errors
func (r *PrerequisiteCheckResult) HasErrors() bool {
	return r.Overall == CheckStatusError
}

// HasWarnings returns true if the check result contains warnings (but no errors)
func (r *PrerequisiteCheckResult) HasWarnings() bool {
	return r.Overall == CheckStatusWarning
}

// FormatHuman returns a human-readable string representation of the check result
func (r *PrerequisiteCheckResult) FormatHuman() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Plugin: %s v%s\n", r.PluginName, r.PluginVersion))

	if r.Runtime != nil {
		icon := statusIcon(r.Runtime.Status)
		sb.WriteString(fmt.Sprintf("  Runtime:            %-20s %s", r.Runtime.Required, icon))
		if r.Runtime.Found != "" {
			sb.WriteString(fmt.Sprintf(" Found: %s", r.Runtime.Found))
		}
		if r.Runtime.Message != "" {
			sb.WriteString(fmt.Sprintf(" %s", r.Runtime.Message))
		}
		sb.WriteString("\n")
		if r.Runtime.Fix != "" {
			sb.WriteString(fmt.Sprintf("                      Fix: %s\n", r.Runtime.Fix))
		}
	}

	if r.PortunixVersion != nil {
		icon := statusIcon(r.PortunixVersion.Status)
		sb.WriteString(fmt.Sprintf("  Portunix version:   %-20s %s Current: %s\n", r.PortunixVersion.Required, icon, r.PortunixVersion.Found))
	}

	if r.OSSupport != nil {
		icon := statusIcon(r.OSSupport.Status)
		sb.WriteString(fmt.Sprintf("  OS support:         %-20s %s Current: %s\n", strings.Join(r.OSSupport.Required, ","), icon, r.OSSupport.Current))
	}

	if len(r.OptionalTools) > 0 {
		sb.WriteString("  Optional tools:\n")
		for _, tool := range r.OptionalTools {
			icon := statusIcon(tool.Status)
			sb.WriteString(fmt.Sprintf("    - %-30s %s", tool.Name, icon))
			if tool.Message != "" {
				sb.WriteString(fmt.Sprintf(" %s", tool.Message))
			}
			sb.WriteString("\n")
			if tool.Fix != "" {
				sb.WriteString(fmt.Sprintf("                                    Fix: %s\n", tool.Fix))
			}
		}
	}

	return sb.String()
}

// FormatJSON returns JSON representation of the check result
func (r *PrerequisiteCheckResult) FormatJSON() (string, error) {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// checkRuntime validates that the required runtime is available
func checkRuntime(runtimeType, runtimeVersion string) *RuntimeCheckResult {
	if runtimeType == "" || runtimeType == "native" {
		return &RuntimeCheckResult{
			Status:   CheckStatusOK,
			Required: "native (self-contained)",
			Found:    "n/a",
		}
	}

	required := runtimeType
	if runtimeVersion != "" {
		required = fmt.Sprintf("%s %s", runtimeType, runtimeVersion)
	}

	switch runtimeType {
	case "java":
		return checkJavaRuntime(required, runtimeVersion)
	case "python":
		return CheckPythonRuntime(required, runtimeVersion)
	default:
		return &RuntimeCheckResult{
			Status:   CheckStatusWarning,
			Required: required,
			Message:  fmt.Sprintf("Unknown runtime type: %s", runtimeType),
		}
	}
}

// checkJavaRuntime checks if Java is installed and meets version requirements
func checkJavaRuntime(required, versionConstraint string) *RuntimeCheckResult {
	cmd := exec.Command("java", "-version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return &RuntimeCheckResult{
			Status:   CheckStatusError,
			Required: required,
			Message:  "Java runtime not found",
			Fix:      "portunix install java-21",
		}
	}

	// Parse Java version from output (e.g., 'openjdk version "21.0.3"' or 'java version "1.8.0_392"')
	outputStr := string(output)
	javaVersion := parseJavaVersion(outputStr)
	if javaVersion == "" {
		return &RuntimeCheckResult{
			Status:   CheckStatusWarning,
			Required: required,
			Found:    outputStr,
			Message:  "Could not parse Java version",
		}
	}

	// Check version constraint if specified
	if versionConstraint != "" {
		if !satisfiesConstraint(javaVersion, versionConstraint) {
			return &RuntimeCheckResult{
				Status:   CheckStatusError,
				Required: required,
				Found:    javaVersion,
				Message:  fmt.Sprintf("Java version %s does not satisfy %s", javaVersion, versionConstraint),
				Fix:      suggestJavaInstall(versionConstraint),
			}
		}
	}

	return &RuntimeCheckResult{
		Status:   CheckStatusOK,
		Required: required,
		Found:    javaVersion,
	}
}

// CheckPythonRuntime checks if Python is installed and meets version requirements
func CheckPythonRuntime(required, versionConstraint string) *RuntimeCheckResult {
	// Try python3 first, then python
	var cmd *exec.Cmd
	var output []byte
	var err error

	for _, pythonCmd := range []string{"python3", "python"} {
		cmd = exec.Command(pythonCmd, "--version")
		output, err = cmd.CombinedOutput()
		if err == nil {
			break
		}
	}

	if err != nil {
		return &RuntimeCheckResult{
			Status:   CheckStatusError,
			Required: required,
			Message:  "Python runtime not found",
			Fix:      "portunix install python",
		}
	}

	// Parse Python version from output (e.g., 'Python 3.12.1')
	outputStr := strings.TrimSpace(string(output))
	pythonVersion := parsePythonVersion(outputStr)
	if pythonVersion == "" {
		return &RuntimeCheckResult{
			Status:   CheckStatusWarning,
			Required: required,
			Found:    outputStr,
			Message:  "Could not parse Python version",
		}
	}

	if versionConstraint != "" {
		if !satisfiesConstraint(pythonVersion, versionConstraint) {
			return &RuntimeCheckResult{
				Status:   CheckStatusError,
				Required: required,
				Found:    pythonVersion,
				Message:  fmt.Sprintf("Python version %s does not satisfy %s", pythonVersion, versionConstraint),
				Fix:      "portunix install python",
			}
		}
	}

	return &RuntimeCheckResult{
		Status:   CheckStatusOK,
		Required: required,
		Found:    pythonVersion,
	}
}

// checkPortunixVersion validates the current Portunix version against requirement
func checkPortunixVersion(minVersion string) *VersionCheckResult {
	currentVersion := version.ProductVersion

	if minVersion == "" {
		return &VersionCheckResult{
			Status:   CheckStatusOK,
			Required: "any",
			Found:    currentVersion,
		}
	}

	// Dev versions are always compatible
	cleanVersion := strings.TrimPrefix(currentVersion, "v")
	if cleanVersion == "dev" || strings.Contains(cleanVersion, "dev") {
		return &VersionCheckResult{
			Status:   CheckStatusOK,
			Required: minVersion,
			Found:    currentVersion,
			Message:  "Development version (always compatible)",
		}
	}

	if !satisfiesConstraint(currentVersion, minVersion) {
		return &VersionCheckResult{
			Status:   CheckStatusError,
			Required: minVersion,
			Found:    currentVersion,
			Message:  fmt.Sprintf("Portunix %s does not satisfy %s", currentVersion, minVersion),
		}
	}

	return &VersionCheckResult{
		Status:   CheckStatusOK,
		Required: minVersion,
		Found:    currentVersion,
	}
}

// checkOSSupport validates the current OS against supported OS list
func checkOSSupport(supportedOS []string) *OSCheckResult {
	currentOS := runtime.GOOS

	if len(supportedOS) == 0 {
		return &OSCheckResult{
			Status:   CheckStatusOK,
			Required: []string{"any"},
			Current:  currentOS,
		}
	}

	for _, os := range supportedOS {
		if os == currentOS {
			return &OSCheckResult{
				Status:   CheckStatusOK,
				Required: supportedOS,
				Current:  currentOS,
			}
		}
	}

	return &OSCheckResult{
		Status:   CheckStatusError,
		Required: supportedOS,
		Current:  currentOS,
		Message:  fmt.Sprintf("Current OS '%s' is not in supported list: %s", currentOS, strings.Join(supportedOS, ", ")),
	}
}

// checkOptionalTool checks if an optional tool is available in PATH
func checkOptionalTool(tool OptionalTool) ToolCheckResult {
	_, err := exec.LookPath(tool.Name)
	if err != nil {
		result := ToolCheckResult{
			Name:    tool.Name,
			Status:  CheckStatusWarning,
			Reason:  tool.Reason,
			Message: fmt.Sprintf("Not found (%s)", tool.Reason),
		}
		// Suggest install command if available
		fix := suggestToolInstall(tool.Name)
		if fix != "" {
			result.Fix = fix
		}
		return result
	}

	return ToolCheckResult{
		Name:   tool.Name,
		Status: CheckStatusOK,
		Reason: tool.Reason,
	}
}

// parseJavaVersion extracts the version number from java -version output
func parseJavaVersion(output string) string {
	// Match patterns like: "21.0.3", "17.0.8", "1.8.0_392"
	re := regexp.MustCompile(`(?:version\s+"?)(\d+[\d._]*)`)
	matches := re.FindStringSubmatch(output)
	if len(matches) < 2 {
		return ""
	}
	ver := matches[1]
	// Normalize old-style versions: 1.8.0_392 -> 8
	if strings.HasPrefix(ver, "1.") {
		parts := strings.Split(ver, ".")
		if len(parts) >= 2 {
			return parts[1]
		}
	}
	// For modern versions like 21.0.3, return major version string
	return ver
}

// parsePythonVersion extracts version from "Python X.Y.Z" output
func parsePythonVersion(output string) string {
	re := regexp.MustCompile(`Python\s+(\d+\.\d+(?:\.\d+)?)`)
	matches := re.FindStringSubmatch(output)
	if len(matches) < 2 {
		return ""
	}
	return matches[1]
}

// satisfiesConstraint checks if a version satisfies a constraint like ">=21", ">=3.10", ">=1.5.0"
func satisfiesConstraint(versionStr, constraint string) bool {
	constraint = strings.TrimSpace(constraint)

	// Parse operator and version from constraint
	var op string
	var constraintVer string

	if strings.HasPrefix(constraint, ">=") {
		op = ">="
		constraintVer = strings.TrimPrefix(constraint, ">=")
	} else if strings.HasPrefix(constraint, ">") {
		op = ">"
		constraintVer = strings.TrimPrefix(constraint, ">")
	} else if strings.HasPrefix(constraint, "<=") {
		op = "<="
		constraintVer = strings.TrimPrefix(constraint, "<=")
	} else if strings.HasPrefix(constraint, "<") {
		op = "<"
		constraintVer = strings.TrimPrefix(constraint, "<")
	} else if strings.HasPrefix(constraint, "=") {
		op = "="
		constraintVer = strings.TrimPrefix(constraint, "=")
	} else {
		// No operator means exact match or >= (treat as >=)
		op = ">="
		constraintVer = constraint
	}

	constraintVer = strings.TrimSpace(constraintVer)

	// Strip 'v' prefix for comparison
	versionStr = strings.TrimPrefix(versionStr, "v")
	constraintVer = strings.TrimPrefix(constraintVer, "v")

	cmp := compareVersionStrings(versionStr, constraintVer)

	switch op {
	case ">=":
		return cmp >= 0
	case ">":
		return cmp > 0
	case "<=":
		return cmp <= 0
	case "<":
		return cmp < 0
	case "=":
		return cmp == 0
	default:
		return cmp >= 0
	}
}

// compareVersionStrings compares two version strings numerically
// Returns -1 if a < b, 0 if a == b, 1 if a > b
func compareVersionStrings(a, b string) int {
	// Normalize: replace underscores with dots, strip non-version suffixes
	a = strings.ReplaceAll(a, "_", ".")
	b = strings.ReplaceAll(b, "_", ".")

	partsA := strings.Split(a, ".")
	partsB := strings.Split(b, ".")

	maxLen := len(partsA)
	if len(partsB) > maxLen {
		maxLen = len(partsB)
	}

	for i := 0; i < maxLen; i++ {
		var numA, numB int
		if i < len(partsA) {
			numA, _ = strconv.Atoi(partsA[i])
		}
		if i < len(partsB) {
			numB, _ = strconv.Atoi(partsB[i])
		}

		if numA < numB {
			return -1
		}
		if numA > numB {
			return 1
		}
	}

	return 0
}

// suggestJavaInstall returns the portunix install command for the required Java version
func suggestJavaInstall(constraint string) string {
	// Extract major version from constraint
	constraint = strings.TrimPrefix(constraint, ">=")
	constraint = strings.TrimSpace(constraint)
	parts := strings.Split(constraint, ".")
	if len(parts) > 0 {
		major, err := strconv.Atoi(parts[0])
		if err == nil {
			return fmt.Sprintf("portunix install java-%d", major)
		}
	}
	return "portunix install java-21"
}

// suggestToolInstall returns a portunix install command for a tool if available
func suggestToolInstall(toolName string) string {
	// Known tool-to-package mappings
	knownTools := map[string]string{
		"protoc":    "portunix install protoc",
		"tesseract": "portunix install tesseract",
		"docker":    "portunix container setup",
		"podman":    "portunix podman setup",
		"go":        "portunix install go",
		"node":      "portunix install nodejs",
		"npm":       "portunix install nodejs",
		"python3":   "portunix install python",
		"python":    "portunix install python",
		"java":      "portunix install java-21",
		"mvn":       "portunix install maven",
		"gradle":    "portunix install gradle",
		"git":       "portunix install git",
		"curl":      "portunix install curl",
		"wget":      "portunix install wget",
	}

	if fix, ok := knownTools[toolName]; ok {
		return fix
	}
	return ""
}

// statusIcon returns the emoji icon for a check status
func statusIcon(status CheckStatus) string {
	switch status {
	case CheckStatusOK:
		return "\u2705" // green checkmark
	case CheckStatusWarning:
		return "\u26a0\ufe0f" // warning
	case CheckStatusError:
		return "\u274c" // red X
	default:
		return "?"
	}
}
