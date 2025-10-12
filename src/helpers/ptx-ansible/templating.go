package main

import (
	"fmt"
	"os"
	"regexp"
	"runtime"
	"strings"
)

// TemplateEngine handles advanced Jinja2-style variable templating
type TemplateEngine struct {
	variables   map[string]interface{}
	environment map[string]interface{}
}

// NewTemplateEngine creates a new template engine with given variables
func NewTemplateEngine(variables map[string]interface{}, environment map[string]interface{}) *TemplateEngine {
	engine := &TemplateEngine{
		variables:   make(map[string]interface{}),
		environment: make(map[string]interface{}),
	}

	// Copy variables
	for k, v := range variables {
		engine.variables[k] = v
	}

	// Copy environment variables
	for k, v := range environment {
		engine.environment[k] = v
	}

	// Add built-in environment variables
	engine.addBuiltInVariables()

	return engine
}

// addBuiltInVariables adds built-in environment variables
func (te *TemplateEngine) addBuiltInVariables() {
	// Operating system information
	te.environment["os"] = runtime.GOOS
	te.environment["arch"] = runtime.GOARCH
	te.environment["os_family"] = getOSFamily()

	// Environment variables
	te.environment["user"] = os.Getenv("USER")
	if te.environment["user"] == "" {
		te.environment["user"] = os.Getenv("USERNAME") // Windows
	}
	te.environment["home"] = os.Getenv("HOME")
	if te.environment["home"] == "" {
		te.environment["home"] = os.Getenv("USERPROFILE") // Windows
	}
	te.environment["pwd"] = getCurrentDirectory()
	te.environment["hostname"] = getHostname()

	// Container/VM detection
	te.environment["is_container"] = isRunningInContainer()
	te.environment["is_vm"] = isRunningInVM()
	te.environment["is_wsl"] = isRunningInWSL()
}

// ProcessTemplate processes a template string with variables
func (te *TemplateEngine) ProcessTemplate(template string) (string, error) {
	result := template

	// Process simple variable substitutions {{ variable_name }}
	simpleVarRegex := regexp.MustCompile(`\{\{\s*([a-zA-Z_][a-zA-Z0-9_]*)\s*\}\}`)
	result = simpleVarRegex.ReplaceAllStringFunc(result, func(match string) string {
		// Extract variable name
		varName := simpleVarRegex.FindStringSubmatch(match)[1]

		// Look in variables first, then environment
		if value, exists := te.variables[varName]; exists {
			return fmt.Sprintf("%v", value)
		}
		if value, exists := te.environment[varName]; exists {
			return fmt.Sprintf("%v", value)
		}

		// Return original if not found
		return match
	})

	// Process conditional expressions {{ variable_name if condition else default }}
	conditionalRegex := regexp.MustCompile(`\{\{\s*([^}]+)\s+if\s+([^}]+)\s+else\s+([^}]+)\s*\}\}`)
	result = conditionalRegex.ReplaceAllStringFunc(result, func(match string) string {
		matches := conditionalRegex.FindStringSubmatch(match)
		if len(matches) != 4 {
			return match
		}

		valueExpr := strings.TrimSpace(matches[1])
		condition := strings.TrimSpace(matches[2])
		defaultValue := strings.TrimSpace(matches[3])

		// Evaluate condition
		if te.evaluateCondition(condition) {
			return te.evaluateExpression(valueExpr)
		}
		return te.evaluateExpression(defaultValue)
	})

	// Process environment-specific conditionals {{ 'value' if os == 'linux' }}
	envConditionalRegex := regexp.MustCompile(`\{\{\s*'([^']+)'\s+if\s+([^}]+)\s*\}\}`)
	result = envConditionalRegex.ReplaceAllStringFunc(result, func(match string) string {
		matches := envConditionalRegex.FindStringSubmatch(match)
		if len(matches) != 3 {
			return match
		}

		value := matches[1]
		condition := strings.TrimSpace(matches[2])

		// Evaluate condition
		if te.evaluateCondition(condition) {
			return value
		}
		return ""
	})

	return result, nil
}

// evaluateExpression evaluates a simple expression
func (te *TemplateEngine) evaluateExpression(expr string) string {
	expr = strings.TrimSpace(expr)

	// Remove quotes if present
	if strings.HasPrefix(expr, "'") && strings.HasSuffix(expr, "'") {
		return strings.Trim(expr, "'")
	}
	if strings.HasPrefix(expr, "\"") && strings.HasSuffix(expr, "\"") {
		return strings.Trim(expr, "\"")
	}

	// Look up variable
	if value, exists := te.variables[expr]; exists {
		return fmt.Sprintf("%v", value)
	}
	if value, exists := te.environment[expr]; exists {
		return fmt.Sprintf("%v", value)
	}

	return expr
}

// evaluateCondition evaluates a conditional expression
func (te *TemplateEngine) evaluateCondition(condition string) bool {
	condition = strings.TrimSpace(condition)

	// Handle equality comparisons (variable == 'value')
	if strings.Contains(condition, "==") {
		parts := strings.Split(condition, "==")
		if len(parts) != 2 {
			return false
		}

		left := te.evaluateExpression(strings.TrimSpace(parts[0]))
		right := te.evaluateExpression(strings.TrimSpace(parts[1]))

		return left == right
	}

	// Handle inequality comparisons (variable != 'value')
	if strings.Contains(condition, "!=") {
		parts := strings.Split(condition, "!=")
		if len(parts) != 2 {
			return false
		}

		left := te.evaluateExpression(strings.TrimSpace(parts[0]))
		right := te.evaluateExpression(strings.TrimSpace(parts[1]))

		return left != right
	}

	// Handle simple variable existence
	varName := strings.TrimSpace(condition)
	if value, exists := te.variables[varName]; exists {
		return te.isTruthy(value)
	}
	if value, exists := te.environment[varName]; exists {
		return te.isTruthy(value)
	}

	return false
}

// isTruthy determines if a value is "truthy"
func (te *TemplateEngine) isTruthy(value interface{}) bool {
	switch v := value.(type) {
	case bool:
		return v
	case string:
		return v != "" && v != "0" && strings.ToLower(v) != "false"
	case int, int32, int64:
		return fmt.Sprintf("%v", v) != "0"
	case float32, float64:
		return fmt.Sprintf("%v", v) != "0"
	}
	return value != nil
}

// MergeVariables merges additional variables into the template engine
func (te *TemplateEngine) MergeVariables(additional map[string]interface{}) {
	for k, v := range additional {
		te.variables[k] = v
	}
}

// Utility functions for environment detection

func getOSFamily() string {
	switch runtime.GOOS {
	case "linux":
		return "unix"
	case "darwin":
		return "unix"
	case "windows":
		return "windows"
	default:
		return runtime.GOOS
	}
}

func getCurrentDirectory() string {
	if pwd, err := os.Getwd(); err == nil {
		return pwd
	}
	return ""
}

func getHostname() string {
	if hostname, err := os.Hostname(); err == nil {
		return hostname
	}
	return "unknown"
}

func isRunningInContainer() bool {
	// Check for common container indicators
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}

	// Check cgroup for container indicators
	if content, err := os.ReadFile("/proc/1/cgroup"); err == nil {
		contentStr := string(content)
		return strings.Contains(contentStr, "docker") ||
			   strings.Contains(contentStr, "containerd") ||
			   strings.Contains(contentStr, "lxc")
	}

	return false
}

func isRunningInVM() bool {
	// Check DMI/SMBIOS for VM indicators (Linux)
	if runtime.GOOS == "linux" {
		vmIndicators := []string{
			"/sys/class/dmi/id/product_name",
			"/sys/class/dmi/id/sys_vendor",
			"/sys/class/dmi/id/board_vendor",
		}

		for _, path := range vmIndicators {
			if content, err := os.ReadFile(path); err == nil {
				contentStr := strings.ToLower(string(content))
				if strings.Contains(contentStr, "vmware") ||
				   strings.Contains(contentStr, "virtualbox") ||
				   strings.Contains(contentStr, "qemu") ||
				   strings.Contains(contentStr, "kvm") ||
				   strings.Contains(contentStr, "xen") ||
				   strings.Contains(contentStr, "microsoft") {
					return true
				}
			}
		}
	}

	return false
}

func isRunningInWSL() bool {
	// Check for WSL-specific indicators
	if runtime.GOOS == "linux" {
		// Check kernel version for WSL
		if content, err := os.ReadFile("/proc/version"); err == nil {
			contentStr := strings.ToLower(string(content))
			return strings.Contains(contentStr, "microsoft") ||
				   strings.Contains(contentStr, "wsl")
		}

		// Check for WSL environment variable
		if wsl := os.Getenv("WSL_DISTRO_NAME"); wsl != "" {
			return true
		}
	}

	return false
}

// Advanced templating functions

// ProcessConditionalExecution evaluates 'when' conditions for packages/playbooks
func ProcessConditionalExecution(whenCondition string, variables map[string]interface{}, environment map[string]interface{}) (bool, error) {
	if whenCondition == "" {
		return true, nil // No condition means always execute
	}

	engine := NewTemplateEngine(variables, environment)
	return engine.evaluateCondition(whenCondition), nil
}

// ProcessPackageVariables processes variables for a specific package
func ProcessPackageVariables(pkg *PtxbookPackage, globalVars map[string]interface{}, environment map[string]interface{}) (*PtxbookPackage, error) {
	engine := NewTemplateEngine(globalVars, environment)

	// Merge package-specific variables
	if pkg.Vars != nil {
		engine.MergeVariables(pkg.Vars)
	}

	// Process name template
	name, err := engine.ProcessTemplate(pkg.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to process package name template: %v", err)
	}

	// Process variant template
	variant, err := engine.ProcessTemplate(pkg.Variant)
	if err != nil {
		return nil, fmt.Errorf("failed to process package variant template: %v", err)
	}

	// Create processed package
	processedPkg := &PtxbookPackage{
		Name:    name,
		Variant: variant,
		When:    pkg.When, // Keep original for condition evaluation
		Vars:    pkg.Vars, // Keep original vars
	}

	return processedPkg, nil
}

// ProcessPlaybookVariables processes variables for a specific playbook
func ProcessPlaybookVariables(playbook *AnsiblePlaybook, globalVars map[string]interface{}, environment map[string]interface{}) (*AnsiblePlaybook, error) {
	engine := NewTemplateEngine(globalVars, environment)

	// Merge playbook-specific variables
	if playbook.Vars != nil {
		engine.MergeVariables(playbook.Vars)
	}

	// Process path template
	path, err := engine.ProcessTemplate(playbook.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to process playbook path template: %v", err)
	}

	// Create processed playbook
	processedPlaybook := &AnsiblePlaybook{
		Path: path,
		When: playbook.When, // Keep original for condition evaluation
		Vars: playbook.Vars, // Keep original vars
	}

	return processedPlaybook, nil
}