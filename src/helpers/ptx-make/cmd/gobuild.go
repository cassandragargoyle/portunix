package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

// envVarPattern matches environment variable assignments like GOOS=linux, CGO_ENABLED=0
var envVarPattern = regexp.MustCompile(`^[A-Z_][A-Z0-9_]*=`)

// RunGoBuild executes go build with cross-platform environment variable support
// Usage: portunix make gobuild [VAR=value]... <command> [args...]
// Example: portunix make gobuild GOOS=linux GOARCH=amd64 go build -o output .
func RunGoBuild(args []string) error {
	if len(args) == 0 {
		showGoBuildHelp()
		return nil
	}

	// Handle --help flag
	if args[0] == "--help" || args[0] == "-h" {
		showGoBuildHelp()
		return nil
	}

	// Parse environment variables and find command start
	envVars := make(map[string]string)
	cmdStart := 0

	envVars, cmdStart = ParseEnvVars(args)

	// Check if command is specified
	if cmdStart >= len(args) {
		return fmt.Errorf("no command specified after environment variables\n\nUsage: portunix make gobuild [VAR=value]... <command> [args...]\nExample: portunix make gobuild GOOS=linux go build -o output .")
	}

	// Extract command and arguments
	cmdName := args[cmdStart]
	cmdArgs := args[cmdStart+1:]

	// Check if command exists
	cmdPath, err := exec.LookPath(cmdName)
	if err != nil {
		return fmt.Errorf("command not found: %s", cmdName)
	}

	// Build the command
	cmd := exec.Command(cmdPath, cmdArgs...)

	// Set environment: inherit current environment and add/override with specified vars
	cmd.Env = os.Environ()
	for key, value := range envVars {
		cmd.Env = append(cmd.Env, key+"="+value)
	}

	// Connect stdio
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Execute
	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		return err
	}

	return nil
}

// IsEnvVar checks if a string matches the pattern for environment variable assignment
// Exported for testing
func IsEnvVar(s string) bool {
	return envVarPattern.MatchString(s)
}

// ParseEnvVars parses environment variables from args and returns env map and command start index
// Exported for testing
func ParseEnvVars(args []string) (map[string]string, int) {
	envVars := make(map[string]string)
	cmdStart := 0

	for i, arg := range args {
		if IsEnvVar(arg) {
			parts := strings.SplitN(arg, "=", 2)
			key := parts[0]
			value := ""
			if len(parts) > 1 {
				value = parts[1]
			}
			envVars[key] = value
			cmdStart = i + 1
		} else {
			break
		}
	}

	return envVars, cmdStart
}

// showGoBuildHelp displays help for the gobuild command
func showGoBuildHelp() {
	fmt.Println("Usage: portunix make gobuild [VAR=value]... <command> [args...]")
	fmt.Println()
	fmt.Println("Execute a command with specified environment variables.")
	fmt.Println("Provides cross-platform support for Unix-style inline environment variables.")
	fmt.Println()
	fmt.Println("On Windows, Unix-style syntax like 'GOOS=linux go build' doesn't work.")
	fmt.Println("This command provides identical behavior across all platforms.")
	fmt.Println()
	fmt.Println("Arguments:")
	fmt.Println("  VAR=value    Environment variable to set (can specify multiple)")
	fmt.Println("  command      Command to execute (e.g., go)")
	fmt.Println("  args         Arguments to pass to the command")
	fmt.Println()
	fmt.Println("Common Go Environment Variables:")
	fmt.Println("  GOOS         Target operating system (linux, windows, darwin)")
	fmt.Println("  GOARCH       Target architecture (amd64, arm64, 386)")
	fmt.Println("  CGO_ENABLED  Enable/disable CGO (0 or 1)")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  # Cross-compile for Linux from any platform")
	fmt.Println("  portunix make gobuild GOOS=linux GOARCH=amd64 go build -o myapp .")
	fmt.Println()
	fmt.Println("  # Build for macOS ARM64 with CGO disabled")
	fmt.Println("  portunix make gobuild CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -o myapp .")
	fmt.Println()
	fmt.Println("  # With ldflags for version embedding")
	fmt.Println("  portunix make gobuild GOOS=linux go build -ldflags \"-X main.version=1.0.0\" -o myapp .")
	fmt.Println()
	fmt.Println("  # Native build (no cross-compilation)")
	fmt.Println("  portunix make gobuild go build -o myapp .")
	fmt.Println()
	fmt.Println("Makefile Integration:")
	fmt.Println("  # Before (Unix only):")
	fmt.Println("  # GOOS=linux GOARCH=amd64 go build -o output .")
	fmt.Println()
	fmt.Println("  # After (cross-platform):")
	fmt.Println("  # portunix make gobuild GOOS=linux GOARCH=amd64 go build -o output .")
}
