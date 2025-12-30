package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

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
	Version:               version,
	DisableFlagParsing:    true, // Disable automatic flag parsing to handle custom flags
	Run: func(cmd *cobra.Command, args []string) {
		// Handle the dispatched command directly
		handleCommand(args)
	},
}

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
	fmt.Println("Python Development Commands:")
	fmt.Println()
	fmt.Println("Virtual Environment Management:")
	fmt.Println("  venv create <name>           - Create a new virtual environment")
	fmt.Println("  venv list                    - List all virtual environments with Python versions")
	fmt.Println("  venv list --group-by-version - Group venvs by Python version")
	fmt.Println("  venv exists <name>           - Check if venv exists (exit code 0/1)")
	fmt.Println("  venv scan [path]             - Discover all venvs in directory")
	fmt.Println("  venv activate <name>         - Activate virtual environment")
	fmt.Println("  venv delete <name>           - Remove virtual environment")
	fmt.Println("  venv info <name>             - Show venv details (Python version, packages)")
	fmt.Println()
	fmt.Println("Package Management:")
	fmt.Println("  pip install <package>        - Install package to active/specified venv")
	fmt.Println("  pip install -r requirements.txt - Install from requirements file")
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
	fmt.Println("Options:")
	fmt.Println("  --venv <name>                - Target specific virtual environment")
	fmt.Println("  --global                     - Operate on system Python")
	fmt.Println("  --path <path>                - Custom venv location")
}

func handlePythonCommand(args []string) {
	if len(args) == 0 {
		showPythonHelp()
		return
	}

	subcommand := args[0]
	subArgs := args[1:]

	switch subcommand {
	case "venv":
		handleVenvCommand(subArgs)
	case "pip":
		handlePipCommand(subArgs)
	case "build":
		handleBuildCommand(subArgs)
	case "check":
		handleCheckCommand()
	case "--help", "-h":
		showPythonHelp()
	default:
		fmt.Printf("Unknown python subcommand: %s\n", subcommand)
		fmt.Println("Run 'portunix python --help' for available commands")
	}
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

func handleCheckCommand() {
	fmt.Println("Checking Python environment...")
	// TODO: Implement Python detection and helper status check
	fmt.Println("✅ ptx-python helper is available")
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
	if len(args) == 0 {
		fmt.Println("Error: Virtual environment name required")
		fmt.Println("Usage: portunix python venv create <name> [--python <version>]")
		os.Exit(1)
	}

	venvName := args[0]
	pythonVersion := ""

	// Parse optional --python flag
	for i := 1; i < len(args); i++ {
		if args[i] == "--python" && i+1 < len(args) {
			pythonVersion = args[i+1]
			break
		}
	}

	vm, err := NewVenvManager()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if err := vm.CreateVenv(venvName, pythonVersion); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
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
	if len(args) == 0 {
		fmt.Println("Error: Virtual environment name required")
		fmt.Println("Usage: portunix python venv info <name>")
		os.Exit(1)
	}

	venvName := args[0]

	vm, err := NewVenvManager()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	info, err := vm.GetVenvInfo(venvName)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Virtual Environment: %s\n", info.Name)
	fmt.Printf("Python Version: %s\n", info.PythonVersion)
	fmt.Printf("Location: %s\n", info.Path)
	fmt.Printf("Packages: %d installed\n", info.PackageCount)
	fmt.Printf("Size: %s\n", formatSize(info.Size))
}

func handleVenvDelete(args []string) {
	if len(args) == 0 {
		fmt.Println("Error: Virtual environment name required")
		fmt.Println("Usage: portunix python venv delete <name>")
		os.Exit(1)
	}

	venvName := args[0]

	vm, err := NewVenvManager()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if err := vm.DeleteVenv(venvName); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func handleVenvActivate(args []string) {
	if len(args) == 0 {
		fmt.Println("Error: Virtual environment name required")
		fmt.Println("Usage: portunix python venv activate <name>")
		os.Exit(1)
	}

	// TODO: Implementation for venv activation
	// This is complex because activation typically modifies shell environment
	fmt.Println("Note: Venv activation modifies shell environment")
	fmt.Println("To activate manually, run:")

	homeDir, _ := os.UserHomeDir()
	venvPath := fmt.Sprintf("%s/.portunix/python/venvs/%s", homeDir, args[0])

	if runtime.GOOS == "windows" {
		fmt.Printf("  %s\\Scripts\\activate\n", venvPath)
	} else {
		fmt.Printf("  source %s/bin/activate\n", venvPath)
	}
}

func handleVenvScan(args []string) {
	fmt.Println("Scanning for virtual environments...")
	fmt.Println("TODO: Implementation for venv scanning in custom paths")
}

// Pip command implementations
func handlePipInstall(args []string) {
	if len(args) == 0 {
		fmt.Println("Error: Package name or -r requirements.txt required")
		fmt.Println("Usage: portunix python pip install <package> [--venv <name>]")
		fmt.Println("       portunix python pip install -r requirements.txt [--venv <name>]")
		os.Exit(1)
	}

	venvName := ""
	isRequirementsFile := false
	requirementsPath := ""
	packageName := ""

	// Parse arguments
	for i := 0; i < len(args); i++ {
		if args[i] == "-r" && i+1 < len(args) {
			isRequirementsFile = true
			requirementsPath = args[i+1]
			i++
		} else if args[i] == "--venv" && i+1 < len(args) {
			venvName = args[i+1]
			i++
		} else if packageName == "" && !isRequirementsFile {
			packageName = args[i]
		}
	}

	if venvName == "" {
		fmt.Println("Error: --venv flag required to specify target virtual environment")
		fmt.Println("Usage: portunix python pip install <package> --venv <name>")
		fmt.Println("       portunix python pip install -r requirements.txt --venv <name>")
		os.Exit(1)
	}

	vm, err := NewVenvManager()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if isRequirementsFile {
		// Install from requirements.txt
		if err := vm.InstallRequirements(venvName, requirementsPath); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	} else {
		// Install single package
		if err := vm.InstallPackage(venvName, packageName); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	}
}

func handlePipUninstall(args []string) {
	fmt.Println("Uninstalling package...")
	fmt.Println("TODO: Implementation in progress")
}

func handlePipList(args []string) {
	venvName := ""

	// Parse --venv flag
	for i := 0; i < len(args); i++ {
		if args[i] == "--venv" && i+1 < len(args) {
			venvName = args[i+1]
			break
		}
	}

	if venvName == "" {
		fmt.Println("Error: --venv flag required to specify target virtual environment")
		fmt.Println("Usage: portunix python pip list --venv <name>")
		os.Exit(1)
	}

	vm, err := NewVenvManager()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if err := vm.ListPackages(venvName); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}

func handlePipFreeze(args []string) {
	venvName := ""

	// Parse --venv flag
	for i := 0; i < len(args); i++ {
		if args[i] == "--venv" && i+1 < len(args) {
			venvName = args[i+1]
			break
		}
	}

	if venvName == "" {
		fmt.Println("Error: --venv flag required to specify target virtual environment")
		fmt.Println("Usage: portunix python pip freeze --venv <name>")
		os.Exit(1)
	}

	vm, err := NewVenvManager()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	venvPath := filepath.Join(vm.venvBaseDir, venvName)
	if !vm.VenvExists(venvName) {
		fmt.Printf("Error: virtual environment '%s' does not exist\n", venvName)
		os.Exit(1)
	}

	pipExe := vm.getPipExecutable(venvPath)
	cmd := exec.Command(pipExe, "freeze")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
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
