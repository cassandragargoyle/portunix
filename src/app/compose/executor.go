package compose

import (
	"fmt"
	"os"
	"os/exec"
)

// Execute runs a compose command with the detected runtime
// All arguments are passed through to the underlying compose tool
func Execute(args []string) error {
	runtime, err := GetComposeRuntime()
	if err != nil {
		return fmt.Errorf("%v\n\n%s", err, GetInstallationInstructions())
	}

	return ExecuteWithRuntime(runtime, args)
}

// ExecuteWithRuntime runs a compose command using the specified runtime
func ExecuteWithRuntime(runtime *ComposeRuntime, args []string) error {
	// Build the full command arguments
	// For docker compose V2: docker compose <args>
	// For docker-compose V1: docker-compose <args>
	// For podman-compose: podman-compose <args>
	fullArgs := append(runtime.Args, args...)

	cmd := exec.Command(runtime.Command, fullArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}

// ExecuteWithOutput runs a compose command and returns the output
func ExecuteWithOutput(args []string) (string, error) {
	runtime, err := GetComposeRuntime()
	if err != nil {
		return "", fmt.Errorf("%v\n\n%s", err, GetInstallationInstructions())
	}

	fullArgs := append(runtime.Args, args...)

	cmd := exec.Command(runtime.Command, fullArgs...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), err
	}

	return string(output), nil
}

// GetRuntimeDescription returns a formatted description of the detected runtime
func GetRuntimeDescription() string {
	runtime, err := GetComposeRuntime()
	if err != nil {
		return "No compose runtime available"
	}

	return fmt.Sprintf("%s (version %s)", runtime.Name, runtime.Version)
}
