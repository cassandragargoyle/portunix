package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"portunix.ai/portunix/src/helpers/ptx-make/cmd"
)

var version = "dev"

// rootCmd represents the base command for ptx-make
var rootCmd = &cobra.Command{
	Use:   "ptx-make",
	Short: "Cross-platform Makefile utilities",
	Long: `ptx-make provides cross-platform utility functions for Makefiles.
It eliminates the need for platform-specific conditionals by providing
identical behavior on Windows, Linux, and macOS.

This binary is typically invoked by the main portunix dispatcher via:
  portunix make <command> [arguments]

Or directly as:
  ptx-make <command> [arguments]

Available commands:
  File Operations:
    copy      - Copy files/directories with wildcard support
    mkdir     - Create directory tree (like mkdir -p)
    rm        - Remove files/directories recursively
    exists    - Check path existence (exit code 0/1)
    ls        - List directory contents (cross-platform)

  Build Metadata:
    version   - Git version tag (git describe --tags --always --dirty)
    commit    - Short git commit hash
    timestamp - UTC timestamp in ISO 8601 format

  Build Tools:
    gobuild   - Cross-platform Go compilation with env vars

  Utilities:
    checksum  - Generate SHA256 checksums
    chmod     - Set file permissions (no-op on Windows)
    json      - Generate JSON from key-value pairs
    env       - Export platform variables for Makefile`,
	Version:            version,
	DisableFlagParsing: true,
	Run: func(c *cobra.Command, args []string) {
		handleCommand(args)
	},
}

func handleCommand(args []string) {
	if len(args) == 0 {
		showHelp()
		return
	}

	command := args[0]
	subArgs := args[1:]

	// Handle dispatcher prefix "make" - strip it if present
	if command == "make" {
		if len(subArgs) == 0 {
			showHelp()
			return
		}
		command = subArgs[0]
		subArgs = subArgs[1:]
	}

	// Handle special flags
	switch command {
	case "--version", "-v":
		fmt.Printf("ptx-make version %s\n", version)
		return
	case "--help", "-h":
		showHelp()
		return
	case "--description":
		fmt.Println("Cross-platform Makefile utilities")
		return
	case "--list-commands":
		fmt.Println("make")
		return
	}

	// Route to appropriate command
	var err error
	switch command {
	case "copy":
		err = cmd.RunCopy(subArgs)
	case "mkdir":
		err = cmd.RunMkdir(subArgs)
	case "rm":
		err = cmd.RunRm(subArgs)
	case "exists":
		cmd.RunExists(subArgs)
		return
	case "ls":
		err = cmd.RunLs(subArgs)
	case "gobuild":
		err = cmd.RunGoBuild(subArgs)
	case "version":
		err = cmd.RunVersion(subArgs)
	case "commit":
		err = cmd.RunCommit(subArgs)
	case "timestamp":
		cmd.RunTimestamp(subArgs)
		return
	case "checksum":
		err = cmd.RunChecksum(subArgs)
	case "chmod":
		err = cmd.RunChmod(subArgs)
	case "json":
		err = cmd.RunJson(subArgs)
	case "env":
		cmd.RunEnv(subArgs)
		return
	default:
		fmt.Printf("Unknown command: %s\n", command)
		fmt.Println("Run 'portunix make --help' for available commands")
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func showHelp() {
	fmt.Println("Usage: portunix make <command> [arguments]")
	fmt.Println()
	fmt.Println("Cross-platform Makefile utilities")
	fmt.Println()
	fmt.Println("File Operations:")
	fmt.Println("  copy <src> <dst>         - Copy files/directories with wildcard support")
	fmt.Println("  mkdir <path>             - Create directory tree (like mkdir -p)")
	fmt.Println("  rm <path>                - Remove files/directories recursively")
	fmt.Println("  exists <path>            - Check path existence (exit code 0/1)")
	fmt.Println("  ls [options] [path]      - List directory contents (cross-platform)")
	fmt.Println()
	fmt.Println("Build Metadata:")
	fmt.Println("  version                  - Git version tag (git describe)")
	fmt.Println("  commit                   - Short git commit hash")
	fmt.Println("  timestamp                - UTC timestamp in ISO 8601 format")
	fmt.Println()
	fmt.Println("Build Tools:")
	fmt.Println("  gobuild [VAR=val]... cmd - Cross-platform Go compilation with env vars")
	fmt.Println()
	fmt.Println("Utilities:")
	fmt.Println("  checksum <dir> [output]  - Generate SHA256 checksums")
	fmt.Println("  chmod <mode> <file>      - Set file permissions (no-op on Windows)")
	fmt.Println("  json <k=v>...            - Generate JSON from key-value pairs")
	fmt.Println("  env                      - Export platform variables for Makefile")
	fmt.Println()
	fmt.Println("Examples:")
	fmt.Println("  portunix make mkdir dist/bin")
	fmt.Println("  portunix make copy src/*.go dist/")
	fmt.Println("  portunix make rm build/")
	fmt.Println("  portunix make ls -lah")
	fmt.Println("  portunix make version")
	fmt.Println("  portunix make gobuild GOOS=linux GOARCH=amd64 go build -o output .")
	fmt.Println("  portunix make json version=1.0.0 platform=linux-x64")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
