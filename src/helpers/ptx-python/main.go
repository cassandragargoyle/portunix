/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var version = "dev"

// rootCmd represents the base command for ptx-python
var rootCmd = &cobra.Command{
	Use:   "ptx-python",
	Short: "Portunix Python Development Helper",
	Long: `ptx-python is a helper binary for Portunix that handles all Python development operations.
It provides unified interface for virtual environment management, package installation,
code building, and quality tools.

This binary is typically invoked by the main portunix dispatcher and should not be used directly.

Supported features:
- Virtual environment management (venv)
- Package management (pip)
- Build & distribution (PyInstaller, cx_Freeze)
- Code quality tools (linting, formatting, type checking)
- Multi-version Python management`,
	Version:            version,
	DisableFlagParsing: true, // Disable automatic flag parsing to handle custom flags
	Run: func(cmd *cobra.Command, args []string) {
		// Handle the dispatched command directly
		handleCommand(args)
	},
}

// handleCommand dispatches the "python" command routed to this helper by the
// parent portunix binary (see src/dispatcher/dispatcher.go), plus the discovery
// meta-flags --version, --description, and --list-commands used by the
// dispatcher. args arrive without the binary name prefix.
func handleCommand(args []string) {
	// Handle dispatched commands: python
	if len(args) == 0 {
		fmt.Println("No command specified")
		return
	}

	command := args[0]
	subArgs := args[1:]

	switch command {
	case "python":
		if len(subArgs) == 0 {
			// Show python help
			showPythonHelp()
		} else {
			handlePythonCommand(subArgs)
		}
	case "--version":
		fmt.Printf("ptx-python version %s\n", version)
	case "--description":
		fmt.Println("Portunix Python Development Helper")
	case "--list-commands":
		fmt.Println("python")
	default:
		fmt.Printf("Unknown command: %s\n", command)
		fmt.Println("Supported commands: python")
	}
}

func showPythonHelp() {
	fmt.Println("Usage: portunix python [subcommand]")
	fmt.Println()
	fmt.Println("Compatibility: Python 3.x only (Python 2 is not supported)")
	fmt.Println()
	fmt.Println("Python Development Commands:")
	fmt.Println()
	fmt.Println("Project Setup:")
	fmt.Println("  init                         - Initialize project: create ./.venv, install deps")
	fmt.Println("  init --force                 - Recreate existing venv")
	fmt.Println("  init --python <version>      - Specify Python version (e.g., 3.11)")
	fmt.Println()
	fmt.Println("Virtual Environment Management:")
	fmt.Println("  venv create <name>           - Create centralized venv (~/.portunix/python/venvs/)")
	fmt.Println("  venv create --local          - Create project-local venv (./.venv)")
	fmt.Println("  venv create --path <dir>     - Create venv at custom location")
	fmt.Println("  venv list                    - List all virtual environments")
	fmt.Println("  venv list --group-by-version - Group venvs by Python version")
	fmt.Println("  venv exists <name>           - Check if venv exists (exit code 0/1)")
	fmt.Println("  venv info                    - Show ./.venv details (auto-detect)")
	fmt.Println("  venv info --verbose          - Include component versions (pip, setuptools)")
	fmt.Println("  venv info --json             - Output in JSON format (implies --verbose)")
	fmt.Println("  venv delete <name>           - Remove virtual environment")
	fmt.Println("  venv delete --local          - Remove ./.venv")
	fmt.Println("  venv activate <name>         - Show activation command")
	fmt.Println("  venv scan [path]             - Discover all venvs in directory")
	fmt.Println()
	fmt.Println("Package Management:")
	fmt.Println("  pip install <package>        - Install package (auto-detects ./.venv)")
	fmt.Println("  pip install -r requirements.txt - Install from requirements file")
	fmt.Println("  pip install <pkg> --local    - Install to ./.venv explicitly")
	fmt.Println("  pip install <pkg> --venv <n> - Install to centralized venv")
	fmt.Println("  pip uninstall <package>      - Remove package")
	fmt.Println("  pip list                     - List installed packages")
	fmt.Println("  pip freeze                   - Generate requirements.txt")
	fmt.Println()
	fmt.Println("Build & Distribution:")
	fmt.Println("  build exe <script.py>        - Build standalone executable with PyInstaller")
	fmt.Println("  build freeze <script.py>     - Build with cx_Freeze")
	fmt.Println("  build wheel                  - Build wheel distribution package")
	fmt.Println("  build sdist                  - Build source distribution package")
	fmt.Println()
	fmt.Println("Code Quality:")
	fmt.Println("  check syntax [paths...]      - Validate Python syntax via AST")
	fmt.Println("  lint [paths...]              - Run linter (default: ruff)")
	fmt.Println("  lint --linter pylint|flake8|ruff   - Choose linter")
	fmt.Println("  lint --format json|html|github     - Output format")
	fmt.Println("  format [paths...]            - Format code (default: black)")
	fmt.Println("  format --formatter autopep8|black  - Choose formatter")
	fmt.Println("  format --check               - Do not modify; CI mode")
	fmt.Println("  typecheck [paths...]         - Run mypy")
	fmt.Println("  test [paths...]              - Run pytest")
	fmt.Println("  test --coverage              - Run with coverage report")
	fmt.Println("  test --watch                 - Re-run on file changes")
	fmt.Println()
	fmt.Println("Version Management:")
	fmt.Println("  version list                 - List installed Python versions")
	fmt.Println("  version use <ver>            - Pin project version (.python-version)")
	fmt.Println("  version detect               - Detect project Python requirement")
	fmt.Println()
	fmt.Println("Environment & Project:")
	fmt.Println("  env show                     - Show Python env configuration")
	fmt.Println("  env set <VAR> <value>        - Print shell snippet to set env var")
	fmt.Println("  init project <name>          - Scaffold new Python project")
	fmt.Println("  init project <n> --template console|library|web|package")
	fmt.Println("  audit                        - Security scan (pip-audit/safety)")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --local                      - Use project-local venv (./.venv)")
	fmt.Println("  --path <path>                - Use venv at custom location")
	fmt.Println("  --venv <name>                - Use centralized venv by name")
	fmt.Println("  --global                     - Operate on system Python")
}

func handlePythonCommand(args []string) {
	if len(args) == 0 {
		showPythonHelp()
		return
	}

	subcommand := args[0]
	subArgs := args[1:]

	switch subcommand {
	case "init":
		handleInitCommand(subArgs)
	case "venv":
		handleVenvCommand(subArgs)
	case "pip":
		handlePipCommand(subArgs)
	case "build":
		handleBuildCommand(subArgs)
	case "check":
		handleCheckCommand(subArgs)
	case "lint":
		handleLintCommand(subArgs)
	case "format":
		handleFormatCommand(subArgs)
	case "typecheck":
		handleTypecheckCommand(subArgs)
	case "test":
		handleTestCommand(subArgs)
	case "version":
		handleVersionCommand(subArgs)
	case "env":
		handleEnvCommand(subArgs)
	case "audit":
		handleAuditCommand(subArgs)
	case "--help", "-h":
		showPythonHelp()
	default:
		fmt.Printf("Unknown python subcommand: %s\n", subcommand)
		fmt.Println("Run 'portunix python --help' for available commands")
	}
}

// handleInitCommand initializes a Python project with local venv. The
// "project" subcommand scaffolds a brand-new project directory instead.
func handleInitCommand(args []string) {
	if len(args) > 0 && args[0] == "project" {
		handleInitProject(args[1:])
		return
	}

	force := false
	pythonVersion := ""
	customPath := ""

	// Parse flags
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--force", "-f":
			force = true
		case "--python":
			if i+1 < len(args) {
				pythonVersion = args[i+1]
				i++
			}
		case "--path":
			if i+1 < len(args) {
				customPath = args[i+1]
				i++
			}
		}
	}

	vm, err := NewVenvManager()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Determine venv path
	var venvPath string
	if customPath != "" {
		venvPath, err = filepath.Abs(customPath)
		if err != nil {
			fmt.Printf("Error: invalid path: %v\n", err)
			os.Exit(1)
		}
	} else {
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Printf("Error: failed to get current directory: %v\n", err)
			os.Exit(1)
		}
		venvPath = filepath.Join(cwd, ".venv")
	}

	// Check for existing venv
	if vm.VenvExistsAtPath(venvPath) && !force {
		fmt.Printf("Virtual environment already exists at %s\n", venvPath)
		fmt.Println("Use --force to recreate it")
		os.Exit(1)
	}

	// Detect requirements file
	requirementsFile, reqErr := vm.DetectRequirementsFile()
	if reqErr != nil {
		fmt.Printf("Warning: %v\n", reqErr)
		fmt.Println("Continuing without installing dependencies...")
	}

	// Step 1: Create venv
	fmt.Println()
	fmt.Println("🐍 Initializing Python project...")
	fmt.Println()

	if err := vm.CreateLocalVenv(venvPath, force, pythonVersion); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Step 2: Upgrade pip
	if err := vm.UpgradePip(venvPath); err != nil {
		fmt.Printf("Warning: failed to upgrade pip: %v\n", err)
		// Continue anyway, pip might still work
	}

	// Step 3: Install dependencies
	if reqErr == nil {
		fmt.Println()
		if err := vm.InstallRequirementsAtPath(venvPath, requirementsFile); err != nil {
			fmt.Printf("Error installing dependencies: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✅ Dependencies installed successfully")
	}

	// Display activation command
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println("✅ Project initialized successfully!")
	fmt.Println()
	fmt.Println("To activate the virtual environment, run:")
	fmt.Printf("  %s\n", vm.GetActivationCommand(venvPath))
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
}

// Command handlers
func handleVenvCommand(args []string) {
	if len(args) == 0 {
		showVenvHelp()
		return
	}

	subcommand := args[0]
	subArgs := args[1:]

	switch subcommand {
	case "create":
		handleVenvCreate(subArgs)
	case "list", "ls":
		handleVenvList(subArgs)
	case "exists":
		handleVenvExists(subArgs)
	case "info":
		handleVenvInfo(subArgs)
	case "delete", "rm":
		handleVenvDelete(subArgs)
	case "activate":
		handleVenvActivate(subArgs)
	case "scan":
		handleVenvScan(subArgs)
	case "--help", "-h":
		showVenvHelp()
	default:
		fmt.Printf("Unknown venv subcommand: %s\n", subcommand)
		fmt.Println("Run 'portunix python venv --help' for available commands")
	}
}

func handlePipCommand(args []string) {
	if len(args) == 0 {
		showPipHelp()
		return
	}

	subcommand := args[0]
	subArgs := args[1:]

	switch subcommand {
	case "install":
		handlePipInstall(subArgs)
	case "uninstall":
		handlePipUninstall(subArgs)
	case "list", "ls":
		handlePipList(subArgs)
	case "freeze":
		handlePipFreeze(subArgs)
	case "--help", "-h":
		showPipHelp()
	default:
		fmt.Printf("Unknown pip subcommand: %s\n", subcommand)
		fmt.Println("Run 'portunix python pip --help' for available commands")
	}
}

// handleCheckCommand handles 'check' and its subcommands. With no args it
// prints helper status; with 'syntax [paths...]' it validates Python syntax
// via the AST parser (Phase 3).
func handleCheckCommand(args []string) {
	if len(args) == 0 {
		fmt.Println("Checking Python environment...")
		fmt.Println("✅ ptx-python helper is available")
		return
	}

	switch args[0] {
	case "syntax":
		target, paths, _ := parseQualityArgs(args[1:])
		qm, err := NewQualityManager()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		if err := qm.CheckSyntax(target, paths); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	case "--help", "-h":
		fmt.Println("Usage: portunix python check [syntax <paths...>]")
	default:
		fmt.Printf("Unknown check subcommand: %s\n", args[0])
		fmt.Println("Run 'portunix python check --help' for available commands")
	}
}

// parseQualityArgs extracts common venv-target flags (--local, --path, --venv)
// from a quality-tool argument list, plus the remaining positional paths and
// unknown flags (returned as extras to be passed through to the underlying
// tool). Auto-detection of ./.venv is enabled.
func parseQualityArgs(args []string) (target *VenvTarget, paths []string, extras []string) {
	venvName, pathFlag := "", ""
	localFlag := false
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch a {
		case "--local", "-l":
			localFlag = true
		case "--path":
			if i+1 < len(args) {
				pathFlag = args[i+1]
				i++
			}
		case "--venv":
			if i+1 < len(args) {
				venvName = args[i+1]
				i++
			}
		default:
			if strings.HasPrefix(a, "-") {
				extras = append(extras, a)
			} else {
				paths = append(paths, a)
			}
		}
	}

	vm, err := NewVenvManager()
	if err == nil {
		t, terr := vm.ResolveVenvPath(localFlag, pathFlag, venvName, true)
		if terr == nil {
			target = t
		}
	}
	return target, paths, extras
}

func showVenvHelp() {
	fmt.Println("Usage: portunix python venv [subcommand]")
	fmt.Println()
	fmt.Println("Virtual Environment Management:")
	fmt.Println("  create <name>           - Create a new virtual environment")
	fmt.Println("  list                    - List all virtual environments with Python versions")
	fmt.Println("  list --group-by-version - Group venvs by Python version")
	fmt.Println("  exists <name>           - Check if venv exists (exit code 0/1)")
	fmt.Println("  scan [path]             - Discover all venvs in directory")
	fmt.Println("  activate <name>         - Activate virtual environment")
	fmt.Println("  delete <name>           - Remove virtual environment")
	fmt.Println("  info <name>             - Show venv details (Python version, packages)")
}

func showPipHelp() {
	fmt.Println("Usage: portunix python pip [subcommand]")
	fmt.Println()
	fmt.Println("Package Management:")
	fmt.Println("  install <package>        - Install package to active/specified venv")
	fmt.Println("  install -r requirements.txt - Install from requirements file")
	fmt.Println("  uninstall <package>      - Remove package")
	fmt.Println("  list                     - List installed packages")
	fmt.Println("  freeze                   - Generate requirements.txt")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --venv <name>            - Target specific virtual environment")
	fmt.Println("  --global                 - Operate on system Python")
}

// Venv command implementations
func handleVenvCreate(args []string) {
	venvName := ""
	pythonVersion := ""
	localFlag := false
	pathFlag := ""
	force := false

	// Parse arguments and flags
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--local", "-l":
			localFlag = true
		case "--path":
			if i+1 < len(args) {
				pathFlag = args[i+1]
				i++
			}
		case "--python":
			if i+1 < len(args) {
				pythonVersion = args[i+1]
				i++
			}
		case "--force", "-f":
			force = true
		default:
			if !strings.HasPrefix(args[i], "-") && venvName == "" {
				venvName = args[i]
			}
		}
	}

	vm, err := NewVenvManager()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Determine if this is local or centralized venv
	if localFlag || pathFlag != "" {
		// Create local/custom path venv
		var venvPath string
		if pathFlag != "" {
			venvPath, err = filepath.Abs(pathFlag)
			if err != nil {
				fmt.Printf("Error: invalid path: %v\n", err)
				os.Exit(1)
			}
		} else {
			cwd, err := os.Getwd()
			if err != nil {
				fmt.Printf("Error: failed to get current directory: %v\n", err)
				os.Exit(1)
			}
			venvPath = filepath.Join(cwd, ".venv")
		}

		if err := vm.CreateLocalVenv(venvPath, force, pythonVersion); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Println()
		fmt.Println("To activate the virtual environment, run:")
		fmt.Printf("  %s\n", vm.GetActivationCommand(venvPath))
	} else {
		// Create centralized venv (original behavior)
		if venvName == "" {
			fmt.Println("Error: Virtual environment name required")
			fmt.Println("Usage: portunix python venv create <name> [--python <version>]")
			fmt.Println("       portunix python venv create --local [--python <version>]")
			fmt.Println("       portunix python venv create --path <dir> [--python <version>]")
			os.Exit(1)
		}

		if err := vm.CreateVenv(venvName, pythonVersion); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		venvPath := filepath.Join(vm.venvBaseDir, venvName)
		fmt.Println()
		fmt.Println("To activate the virtual environment, run:")
		fmt.Printf("  %s\n", vm.GetActivationCommand(venvPath))
	}
}

func handleVenvList(args []string) {
	vm, err := NewVenvManager()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	venvs, err := vm.ListVenvs()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if len(venvs) == 0 {
		fmt.Println("No virtual environments found.")
		fmt.Printf("Create one with: portunix python venv create <name>\n")
		return
	}

	// Check for --group-by-version flag
	groupByVersion := false
	for _, arg := range args {
		if arg == "--group-by-version" {
			groupByVersion = true
			break
		}
	}

	if groupByVersion {
		displayVenvsGrouped(venvs)
	} else {
		displayVenvsList(venvs)
	}
}

func displayVenvsList(venvs []*VenvInfo) {
	fmt.Printf("Virtual Environments in %s:\n\n", venvs[0].Path[:len(venvs[0].Path)-len(venvs[0].Name)-1])

	for _, venv := range venvs {
		sizeStr := formatSize(venv.Size)
		fmt.Printf("  %-20s (Python %s, %d packages, %s)\n",
			venv.Name, venv.PythonVersion, venv.PackageCount, sizeStr)
	}
}

func displayVenvsGrouped(venvs []*VenvInfo) {
	// Group by Python version
	grouped := make(map[string][]*VenvInfo)
	for _, venv := range venvs {
		grouped[venv.PythonVersion] = append(grouped[venv.PythonVersion], venv)
	}

	fmt.Println("Virtual Environments grouped by Python version:\n")
	for version, venvList := range grouped {
		fmt.Printf("Python %s (%d environment(s)):\n", version, len(venvList))
		for _, venv := range venvList {
			sizeStr := formatSize(venv.Size)
			fmt.Printf("  - %-20s (%d packages, %s)\n",
				venv.Name, venv.PackageCount, sizeStr)
		}
		fmt.Println()
	}
}

func formatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.0f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func handleVenvExists(args []string) {
	if len(args) == 0 {
		fmt.Println("Error: Virtual environment name required")
		fmt.Println("Usage: portunix python venv exists <name>")
		os.Exit(1)
	}

	venvName := args[0]

	vm, err := NewVenvManager()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if vm.VenvExists(venvName) {
		os.Exit(0) // Exists - exit code 0
	} else {
		os.Exit(1) // Does not exist - exit code 1
	}
}

func handleVenvInfo(args []string) {
	venvName := ""
	localFlag := false
	pathFlag := ""
	verboseFlag := false
	jsonFlag := false

	// Parse arguments and flags
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--local", "-l":
			localFlag = true
		case "--path":
			if i+1 < len(args) {
				pathFlag = args[i+1]
				i++
			}
		case "--verbose", "-v":
			verboseFlag = true
		case "--json":
			jsonFlag = true
			verboseFlag = true // JSON implies verbose
		default:
			if !strings.HasPrefix(args[i], "-") && venvName == "" {
				venvName = args[i]
			}
		}
	}

	vm, err := NewVenvManager()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Resolve venv path
	var venvPath string
	var isLocal bool

	if localFlag || pathFlag != "" {
		target, err := vm.ResolveVenvPath(localFlag, pathFlag, "", false)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		venvPath = target.Path
		isLocal = target.IsLocal
	} else if venvName != "" {
		venvPath = filepath.Join(vm.venvBaseDir, venvName)
		isLocal = false
	} else {
		// Default: auto-detect local .venv
		target, err := vm.ResolveVenvPath(false, "", "", true)
		if err != nil {
			fmt.Println("Error: No .venv found in current directory")
			fmt.Println("Create one with: portunix python init")
			os.Exit(1)
		}
		venvPath = target.Path
		isLocal = target.IsLocal
	}

	// Get info (verbose or basic)
	var info *VenvInfo
	if verboseFlag {
		info, err = vm.GetVenvInfoAtPathVerbose(venvPath)
	} else {
		info, err = vm.GetVenvInfoAtPath(venvPath)
	}
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	info.IsLocal = isLocal

	// Output
	if jsonFlag {
		outputVenvInfoJSON(info)
	} else {
		outputVenvInfoText(info, verboseFlag)
	}
}

func outputVenvInfoText(info *VenvInfo, verbose bool) {
	fmt.Printf("Virtual Environment: %s\n", info.Name)
	fmt.Printf("Python Version: %s\n", info.PythonVersion)
	fmt.Printf("Location: %s\n", info.Path)
	fmt.Printf("Packages: %d installed\n", info.PackageCount)
	fmt.Printf("Size: %s\n", formatSize(info.Size))
	if info.IsLocal {
		fmt.Println("Type: Project-local")
	} else {
		fmt.Println("Type: Centralized")
	}

	if verbose && len(info.Components) > 0 {
		fmt.Println()
		fmt.Println("Components:")
		for name, version := range info.Components {
			fmt.Printf("  %s: %s\n", name, version)
		}
	}
}

func outputVenvInfoJSON(info *VenvInfo) {
	// Manual JSON output to avoid importing encoding/json for this small case
	fmt.Println("{")
	fmt.Printf("  \"name\": \"%s\",\n", info.Name)
	fmt.Printf("  \"path\": \"%s\",\n", escapeJSON(info.Path))
	fmt.Printf("  \"python_version\": \"%s\",\n", info.PythonVersion)
	fmt.Printf("  \"package_count\": %d,\n", info.PackageCount)
	fmt.Printf("  \"size_bytes\": %d,\n", info.Size)
	fmt.Printf("  \"size_human\": \"%s\",\n", info.SizeHuman)
	fmt.Printf("  \"is_local\": %t,\n", info.IsLocal)
	fmt.Printf("  \"exists\": %t", info.Exists)

	if len(info.Components) > 0 {
		fmt.Println(",")
		fmt.Println("  \"components\": {")
		i := 0
		for name, version := range info.Components {
			if i > 0 {
				fmt.Println(",")
			}
			fmt.Printf("    \"%s\": \"%s\"", name, version)
			i++
		}
		fmt.Println()
		fmt.Println("  }")
	} else {
		fmt.Println()
	}
	fmt.Println("}")
}

func escapeJSON(s string) string {
	// Simple JSON string escaping for paths
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	return s
}

func handleVenvDelete(args []string) {
	venvName := ""
	localFlag := false
	pathFlag := ""

	// Parse arguments and flags
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--local", "-l":
			localFlag = true
		case "--path":
			if i+1 < len(args) {
				pathFlag = args[i+1]
				i++
			}
		default:
			if !strings.HasPrefix(args[i], "-") && venvName == "" {
				venvName = args[i]
			}
		}
	}

	vm, err := NewVenvManager()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if localFlag || pathFlag != "" {
		// Delete local or custom path venv
		target, err := vm.ResolveVenvPath(localFlag, pathFlag, "", false)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}

		if err := vm.DeleteVenvAtPath(target.Path); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	} else {
		// Delete centralized venv
		if venvName == "" {
			fmt.Println("Error: Virtual environment name required")
			fmt.Println("Usage: portunix python venv delete <name>")
			fmt.Println("       portunix python venv delete --local")
			fmt.Println("       portunix python venv delete --path <dir>")
			os.Exit(1)
		}

		if err := vm.DeleteVenv(venvName); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	}
}

func handleVenvActivate(args []string) {
	venvName := ""
	localFlag := false
	pathFlag := ""

	// Parse arguments and flags
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--local", "-l":
			localFlag = true
		case "--path":
			if i+1 < len(args) {
				pathFlag = args[i+1]
				i++
			}
		default:
			if !strings.HasPrefix(args[i], "-") && venvName == "" {
				venvName = args[i]
			}
		}
	}

	vm, err := NewVenvManager()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	var venvPath string

	if localFlag || pathFlag != "" {
		// Local or custom path venv
		target, err := vm.ResolveVenvPath(localFlag, pathFlag, "", false)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		venvPath = target.Path
	} else if venvName != "" {
		// Centralized venv
		venvPath = filepath.Join(vm.venvBaseDir, venvName)
	} else {
		// Try auto-detect local venv
		cwd, _ := os.Getwd()
		localVenvPath := filepath.Join(cwd, ".venv")
		if vm.VenvExistsAtPath(localVenvPath) {
			venvPath = localVenvPath
		} else {
			fmt.Println("Error: Virtual environment name required")
			fmt.Println("Usage: portunix python venv activate <name>")
			fmt.Println("       portunix python venv activate --local")
			fmt.Println("       portunix python venv activate --path <dir>")
			os.Exit(1)
		}
	}

	// Check if venv exists
	if !vm.VenvExistsAtPath(venvPath) {
		fmt.Printf("Error: Virtual environment does not exist at %s\n", venvPath)
		os.Exit(1)
	}

	fmt.Println("Note: Venv activation modifies shell environment")
	fmt.Println("To activate, run:")
	fmt.Printf("  %s\n", vm.GetActivationCommand(venvPath))
}

func handleVenvScan(args []string) {
	fmt.Println("Scanning for virtual environments...")
	fmt.Println("TODO: Implementation for venv scanning in custom paths")
}

// Pip command implementations
func handlePipInstall(args []string) {
	if len(args) == 0 {
		fmt.Println("Error: Package name or -r requirements.txt required")
		fmt.Println("Usage: portunix python pip install <package> [--local|--venv <name>]")
		fmt.Println("       portunix python pip install -r requirements.txt [--local|--venv <name>]")
		os.Exit(1)
	}

	venvName := ""
	localFlag := false
	pathFlag := ""
	isRequirementsFile := false
	requirementsPath := ""
	packages := []string{}

	// Parse arguments
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "-r":
			if i+1 < len(args) {
				isRequirementsFile = true
				requirementsPath = args[i+1]
				i++
			}
		case "--venv":
			if i+1 < len(args) {
				venvName = args[i+1]
				i++
			}
		case "--local", "-l":
			localFlag = true
		case "--path":
			if i+1 < len(args) {
				pathFlag = args[i+1]
				i++
			}
		case "--upgrade", "-U":
			packages = append(packages, "--upgrade")
		default:
			if !strings.HasPrefix(args[i], "-") {
				packages = append(packages, args[i])
			}
		}
	}

	vm, err := NewVenvManager()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Resolve venv target (with auto-detect for pip commands)
	target, err := vm.ResolveVenvPath(localFlag, pathFlag, venvName, true)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		fmt.Println("Usage: portunix python pip install <package> [--local|--venv <name>]")
		fmt.Println("       portunix python pip install -r requirements.txt [--local|--venv <name>]")
		fmt.Println()
		fmt.Println("Tip: Create a local venv first with: portunix python init")
		os.Exit(1)
	}

	if isRequirementsFile {
		// Install from requirements.txt
		if err := vm.InstallRequirementsAtPath(target.Path, requirementsPath); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	} else if len(packages) > 0 {
		// Install packages
		for _, pkg := range packages {
			if pkg == "--upgrade" {
				continue
			}
			if err := vm.InstallPackageAtPath(target.Path, pkg); err != nil {
				fmt.Printf("Error installing %s: %v\n", pkg, err)
				os.Exit(1)
			}
		}
	} else {
		fmt.Println("Error: No package specified")
		os.Exit(1)
	}
}

func handlePipUninstall(args []string) {
	fmt.Println("Uninstalling package...")
	fmt.Println("TODO: Implementation in progress")
}

func handlePipList(args []string) {
	venvName := ""
	localFlag := false
	pathFlag := ""

	// Parse flags
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--venv":
			if i+1 < len(args) {
				venvName = args[i+1]
				i++
			}
		case "--local", "-l":
			localFlag = true
		case "--path":
			if i+1 < len(args) {
				pathFlag = args[i+1]
				i++
			}
		}
	}

	vm, err := NewVenvManager()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Resolve venv target (with auto-detect)
	target, err := vm.ResolveVenvPath(localFlag, pathFlag, venvName, true)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		fmt.Println("Usage: portunix python pip list [--local|--venv <name>]")
		os.Exit(1)
	}

	if err := vm.ListPackagesAtPath(target.Path); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func handlePipFreeze(args []string) {
	venvName := ""
	localFlag := false
	pathFlag := ""

	// Parse flags
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--venv":
			if i+1 < len(args) {
				venvName = args[i+1]
				i++
			}
		case "--local", "-l":
			localFlag = true
		case "--path":
			if i+1 < len(args) {
				pathFlag = args[i+1]
				i++
			}
		}
	}

	vm, err := NewVenvManager()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	// Resolve venv target (with auto-detect)
	target, err := vm.ResolveVenvPath(localFlag, pathFlag, venvName, true)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		fmt.Println("Usage: portunix python pip freeze [--local|--venv <name>]")
		os.Exit(1)
	}

	if err := vm.FreezePackagesAtPath(target.Path); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

// Build command handlers
func handleBuildCommand(args []string) {
	if len(args) == 0 {
		showBuildHelp()
		return
	}

	subcommand := args[0]
	subArgs := args[1:]

	switch subcommand {
	case "exe":
		handleBuildExe(subArgs)
	case "freeze":
		handleBuildFreeze(subArgs)
	case "wheel":
		handleBuildWheel(subArgs)
	case "sdist":
		handleBuildSdist(subArgs)
	case "--help", "-h":
		showBuildHelp()
	default:
		fmt.Printf("Unknown build subcommand: %s\n", subcommand)
		fmt.Println("Run 'portunix python build --help' for available commands")
	}
}

func showBuildHelp() {
	fmt.Println("Usage: portunix python build [subcommand]")
	fmt.Println()
	fmt.Println("Build & Distribution Commands:")
	fmt.Println("  exe <script.py>         - Build standalone executable with PyInstaller")
	fmt.Println("  freeze <script.py>      - Build with cx_Freeze (alternative)")
	fmt.Println("  wheel                   - Build wheel distribution package")
	fmt.Println("  sdist                   - Build source distribution package")
	fmt.Println()
	fmt.Println("Build exe options:")
	fmt.Println("  --venv <name>           - Use specific virtual environment")
	fmt.Println("  --name <name>           - Set custom executable name")
	fmt.Println("  --onefile               - Create single executable file")
	fmt.Println("  --console               - Create console application (default)")
	fmt.Println("  --windowed              - Create windowed application (no console)")
	fmt.Println("  --icon <file.ico>       - Set application icon")
	fmt.Println("  --distpath <path>       - Output directory (default: dist)")
	fmt.Println()
	fmt.Println("Build freeze options:")
	fmt.Println("  --venv <name>           - Use specific virtual environment")
	fmt.Println("  --name <name>           - Set custom executable name")
	fmt.Println("  --target-version <ver>  - Target Python version")
	fmt.Println("  --distpath <path>       - Output directory")
	fmt.Println()
	fmt.Println("Build wheel/sdist options:")
	fmt.Println("  --venv <name>           - Use specific virtual environment")
	fmt.Println("  --path <path>           - Project path (default: current directory)")
}

func handleBuildExe(args []string) {
	if len(args) == 0 {
		fmt.Println("Error: Script file required")
		fmt.Println("Usage: portunix python build exe <script.py> [options]")
		os.Exit(1)
	}

	// Parse arguments
	opts := BuildExeOptions{
		Script: args[0],
	}

	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--venv":
			if i+1 < len(args) {
				opts.VenvName = args[i+1]
				i++
			}
		case "--name":
			if i+1 < len(args) {
				opts.Name = args[i+1]
				i++
			}
		case "--icon":
			if i+1 < len(args) {
				opts.Icon = args[i+1]
				i++
			}
		case "--distpath":
			if i+1 < len(args) {
				opts.OutputDir = args[i+1]
				i++
			}
		case "--onefile":
			opts.OneFile = true
		case "--console":
			opts.Console = true
		case "--windowed":
			opts.Windowed = true
		default:
			// Unknown flag, might be for PyInstaller
			opts.ExtraArgs = append(opts.ExtraArgs, args[i])
		}
	}

	bm, err := NewBuildManager()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if err := bm.BuildExe(opts); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func handleBuildFreeze(args []string) {
	if len(args) == 0 {
		fmt.Println("Error: Script file required")
		fmt.Println("Usage: portunix python build freeze <script.py> [options]")
		os.Exit(1)
	}

	opts := BuildFreezeOptions{
		Script: args[0],
	}

	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--venv":
			if i+1 < len(args) {
				opts.VenvName = args[i+1]
				i++
			}
		case "--name":
			if i+1 < len(args) {
				opts.Name = args[i+1]
				i++
			}
		case "--target-version":
			if i+1 < len(args) {
				opts.TargetVersion = args[i+1]
				i++
			}
		case "--distpath":
			if i+1 < len(args) {
				opts.OutputDir = args[i+1]
				i++
			}
		}
	}

	bm, err := NewBuildManager()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if err := bm.BuildFreeze(opts); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func handleBuildWheel(args []string) {
	venvName := ""
	projectPath := "."

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--venv":
			if i+1 < len(args) {
				venvName = args[i+1]
				i++
			}
		case "--path":
			if i+1 < len(args) {
				projectPath = args[i+1]
				i++
			}
		}
	}

	bm, err := NewBuildManager()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if err := bm.BuildWheel(venvName, projectPath); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func handleBuildSdist(args []string) {
	venvName := ""
	projectPath := "."

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--venv":
			if i+1 < len(args) {
				venvName = args[i+1]
				i++
			}
		case "--path":
			if i+1 < len(args) {
				projectPath = args[i+1]
				i++
			}
		}
	}

	bm, err := NewBuildManager()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if err := bm.BuildSdist(venvName, projectPath); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

// ===== Phase 3: Code Quality handlers =====

func handleLintCommand(args []string) {
	linter := "ruff"
	format := ""
	outputFile := ""
	extras := []string{}
	rest := []string{}

	// Pull linter-specific flags out first so they don't end up in extras.
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch a {
		case "--linter":
			if i+1 < len(args) {
				linter = args[i+1]
				i++
			}
		case "--format":
			if i+1 < len(args) {
				format = args[i+1]
				i++
			}
		case "--output":
			if i+1 < len(args) {
				outputFile = args[i+1]
				i++
			}
		default:
			rest = append(rest, a)
		}
	}

	target, paths, passthrough := parseQualityArgs(rest)
	extras = append(extras, passthrough...)

	qm, err := NewQualityManager()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	if err := qm.Lint(LintOptions{
		Target:     target,
		Paths:      paths,
		Linter:     linter,
		Format:     format,
		OutputFile: outputFile,
		ExtraArgs:  extras,
	}); err != nil {
		// Linters return non-zero on findings; surface their exit code.
		os.Exit(1)
	}
}

func handleFormatCommand(args []string) {
	formatter := "black"
	check := false
	rest := []string{}

	for i := 0; i < len(args); i++ {
		a := args[i]
		switch a {
		case "--formatter":
			if i+1 < len(args) {
				formatter = args[i+1]
				i++
			}
		case "--check":
			check = true
		default:
			rest = append(rest, a)
		}
	}

	target, paths, extras := parseQualityArgs(rest)

	qm, err := NewQualityManager()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	if err := qm.Format(FormatOptions{
		Target:    target,
		Paths:     paths,
		Formatter: formatter,
		Check:     check,
		ExtraArgs: extras,
	}); err != nil {
		os.Exit(1)
	}
}

func handleTypecheckCommand(args []string) {
	target, paths, extras := parseQualityArgs(args)
	qm, err := NewQualityManager()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	if err := qm.Typecheck(TypecheckOptions{
		Target:    target,
		Paths:     paths,
		ExtraArgs: extras,
	}); err != nil {
		os.Exit(1)
	}
}

func handleTestCommand(args []string) {
	coverage := false
	watch := false
	rest := []string{}

	for i := 0; i < len(args); i++ {
		a := args[i]
		switch a {
		case "--coverage":
			coverage = true
		case "--watch":
			watch = true
		default:
			rest = append(rest, a)
		}
	}

	target, paths, extras := parseQualityArgs(rest)

	qm, err := NewQualityManager()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	if err := qm.Test(TestOptions{
		Target:    target,
		Paths:     paths,
		Coverage:  coverage,
		Watch:     watch,
		ExtraArgs: extras,
	}); err != nil {
		os.Exit(1)
	}
}

// ===== Phase 4: Version / Env / Audit / Init Project handlers =====

func handleVersionCommand(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: portunix python version [list|use <ver>|detect]")
		return
	}
	sub := args[0]
	subArgs := args[1:]

	vm := NewVersionManager()

	switch sub {
	case "list", "ls":
		installs, err := vm.ListInstalls()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		if len(installs) == 0 {
			fmt.Println("No Python interpreters found.")
			return
		}
		fmt.Println("Installed Python interpreters:")
		for _, p := range installs {
			fmt.Printf("  Python %-8s  %-10s  %s\n", p.Version, "["+p.Source+"]", p.Executable)
		}
	case "use":
		if len(subArgs) == 0 {
			fmt.Println("Usage: portunix python version use <ver>")
			os.Exit(1)
		}
		ver := subArgs[0]
		if err := vm.WriteProjectVersion(ver); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("✅ Wrote .python-version (%s)\n", ver)
		fmt.Println()
		fmt.Println("Tools that honor .python-version: pyenv, uv, hatch.")
		fmt.Println("For Portunix venv creation, pass --python explicitly:")
		fmt.Printf("  portunix python venv create myenv --python %s\n", ver)
	case "detect":
		ver, err := vm.DetectProjectVersion()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		if ver == "" {
			fmt.Println("No project Python version found (.python-version or pyproject.toml).")
			return
		}
		fmt.Printf("Detected project Python: %s\n", ver)
	case "--help", "-h":
		fmt.Println("Usage: portunix python version [list|use <ver>|detect]")
	default:
		fmt.Printf("Unknown version subcommand: %s\n", sub)
		os.Exit(1)
	}
}

func handleEnvCommand(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: portunix python env [show|set <VAR> <value>]")
		return
	}
	em, err := NewEnvManager()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	sub := args[0]
	subArgs := args[1:]
	switch sub {
	case "show":
		if err := em.ShowEnv(); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	case "set":
		if len(subArgs) < 2 {
			fmt.Println("Usage: portunix python env set <VAR> <value>")
			os.Exit(1)
		}
		if err := em.SetEnv(subArgs[0], strings.Join(subArgs[1:], " ")); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	case "--help", "-h":
		fmt.Println("Usage: portunix python env [show|set <VAR> <value>]")
	default:
		fmt.Printf("Unknown env subcommand: %s\n", sub)
		os.Exit(1)
	}
}

func handleAuditCommand(args []string) {
	em, err := NewEnvManager()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	target, _, extras := parseQualityArgs(args)
	if err := em.Audit(target, extras); err != nil {
		os.Exit(1)
	}
}

// handleInitProject scaffolds a new Python project (called from
// handleInitCommand when the first arg is "project").
func handleInitProject(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: portunix python init project <name> [--template console|library|web|package] [--path <dir>]")
		os.Exit(1)
	}

	opts := InitProjectOptions{Template: "console"}
	for i := 0; i < len(args); i++ {
		a := args[i]
		switch a {
		case "--template":
			if i+1 < len(args) {
				opts.Template = args[i+1]
				i++
			}
		case "--path":
			if i+1 < len(args) {
				opts.Path = args[i+1]
				i++
			}
		default:
			if !strings.HasPrefix(a, "-") && opts.Name == "" {
				opts.Name = a
			}
		}
	}

	em, err := NewEnvManager()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	if err := em.InitProject(opts); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	// Add version flag
	rootCmd.Flags().BoolP("version", "v", false, "Show version")
	rootCmd.Flags().Bool("description", false, "Show description")
	rootCmd.Flags().Bool("list-commands", false, "List available commands")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
