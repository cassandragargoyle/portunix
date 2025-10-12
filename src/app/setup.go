package app

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

// TODO: Finish and use
func Setup() {
	// Get the directory where the program is executed
	execDir, err := os.Getwd()
	if err != nil {
		fmt.Println("Error getting the current directory:", err)
		return
	}

	newPath := execDir + ";" + os.Getenv("PATH")
	os.Setenv("PATH", newPath)

	fmt.Println("New PATH value:", newPath)

	// Open a new console window
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/k") // Open a new cmd window on Windows
	default:
		fmt.Println("This script is currently designed for Windows only.")
		return
	}

	// Pass the updated PATH value to the new process
	cmd.Env = append(os.Environ(), "PATH="+newPath)

	// Start the new shell
	err = cmd.Start()
	if err != nil {
		fmt.Println("Error launching the shell:", err)
		return
	}

	fmt.Println("A new console window has been opened.")
}
