package selfinstall

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// PromptInstallLocation prompts the user to select installation location
func PromptInstallLocation() (string, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Select installation location:")

	options := getInstallOptions()
	for i, opt := range options {
		fmt.Printf("[%d] %s", i+1, opt.Display)
		if opt.Recommended {
			fmt.Print(" (recommended)")
		}
		fmt.Println()
	}
	fmt.Printf("[%d] Custom location\n", len(options)+1)
	fmt.Printf("[%d] Cancel installation\n", len(options)+2)

	fmt.Print("\nPlease select [1-" + fmt.Sprint(len(options)+2) + "]: ")

	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	input = strings.TrimSpace(input)

	var choice int
	if _, err := fmt.Sscanf(input, "%d", &choice); err != nil {
		return "", fmt.Errorf("invalid selection")
	}

	if choice < 1 || choice > len(options)+2 {
		return "", fmt.Errorf("invalid selection")
	}

	// Cancel
	if choice == len(options)+2 {
		return "", nil
	}

	// Custom location
	if choice == len(options)+1 {
		fmt.Print("Enter custom installation path: ")
		customPath, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}
		customPath = strings.TrimSpace(customPath)

		// Add binary name if only directory provided
		if !strings.HasSuffix(customPath, "portunix") && !strings.HasSuffix(customPath, "portunix.exe") {
			if runtime.GOOS == "windows" {
				customPath = filepath.Join(customPath, "portunix.exe")
			} else {
				customPath = filepath.Join(customPath, "portunix")
			}
		}

		return customPath, nil
	}

	// Predefined location
	return options[choice-1].Path, nil
}

// PromptAddToPath prompts the user to add to PATH
func PromptAddToPath() (bool, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("\nAdd Portunix to system PATH? [Y/n]: ")
	input, err := reader.ReadString('\n')
	if err != nil {
		return false, err
	}

	input = strings.TrimSpace(strings.ToLower(input))
	return input == "" || input == "y" || input == "yes", nil
}

// PromptCreateConfig prompts the user to create configuration
func PromptCreateConfig() (bool, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("\nCreate default configuration? [Y/n]: ")
	input, err := reader.ReadString('\n')
	if err != nil {
		return false, err
	}

	input = strings.TrimSpace(strings.ToLower(input))
	return input == "" || input == "y" || input == "yes", nil
}

// PromptBackup prompts the user to backup existing installation
func PromptBackup(existingPath string) (bool, error) {
	reader := bufio.NewReader(os.Stdin)

	fmt.Printf("\n⚠ Warning: %s already exists.\n", existingPath)
	fmt.Print("Create backup? [Y/n]: ")

	input, err := reader.ReadString('\n')
	if err != nil {
		return false, err
	}

	input = strings.TrimSpace(strings.ToLower(input))
	return input == "" || input == "y" || input == "yes", nil
}

// ShowInstallationSummary shows installation summary
func ShowInstallationSummary(path string) error {
	fmt.Println("\n" + strings.Repeat("=", 50))
	fmt.Println("Installation completed successfully!")
	fmt.Println(strings.Repeat("=", 50))
	fmt.Printf("Installed to: %s\n", path)

	// Check if in PATH
	dir := filepath.Dir(path)
	if IsInPath(dir) {
		fmt.Println("✓ Directory is in PATH")
		fmt.Println("\nYou can now run 'portunix' from anywhere")
	} else {
		fmt.Println("⚠ Directory is NOT in PATH")
		fmt.Printf("\nTo use Portunix from anywhere, add this to your PATH:\n")
		fmt.Printf("  %s\n", dir)

		if runtime.GOOS == "windows" {
			fmt.Println("\nWindows: System Properties → Environment Variables → Path")
		} else {
			shell := os.Getenv("SHELL")
			if strings.Contains(shell, "bash") {
				fmt.Printf("\nAdd to ~/.bashrc or ~/.bash_profile:\n")
				fmt.Printf("  export PATH=\"%s:$PATH\"\n", dir)
			} else if strings.Contains(shell, "zsh") {
				fmt.Printf("\nAdd to ~/.zshrc:\n")
				fmt.Printf("  export PATH=\"%s:$PATH\"\n", dir)
			} else {
				fmt.Printf("\nAdd to your shell configuration:\n")
				fmt.Printf("  export PATH=\"%s:$PATH\"\n", dir)
			}
		}
	}

	fmt.Printf("\nRun '%s --version' to verify installation\n", filepath.Base(path))

	return nil
}

type installOption struct {
	Path        string
	Display     string
	Recommended bool
}

func getInstallOptions() []installOption {
	switch runtime.GOOS {
	case "windows":
		progFiles := os.Getenv("PROGRAMFILES")
		localAppData := os.Getenv("LOCALAPPDATA")

		options := []installOption{}

		if progFiles != "" {
			options = append(options, installOption{
				Path:        filepath.Join(progFiles, "Portunix", "portunix.exe"),
				Display:     filepath.Join(progFiles, "Portunix"),
				Recommended: true,
			})
		}

		if localAppData != "" {
			options = append(options, installOption{
				Path:    filepath.Join(localAppData, "Portunix", "portunix.exe"),
				Display: filepath.Join(localAppData, "Portunix"),
			})
		}

		home, _ := os.UserHomeDir()
		if home != "" {
			options = append(options, installOption{
				Path:    filepath.Join(home, "bin", "portunix.exe"),
				Display: filepath.Join(home, "bin"),
			})
		}

		return options

	case "darwin":
		home, _ := os.UserHomeDir()
		return []installOption{
			{
				Path:        "/usr/local/bin/portunix",
				Display:     "/usr/local/bin",
				Recommended: checkWriteAccess("/usr/local/bin") == nil,
			},
			{
				Path:    filepath.Join(home, "bin", "portunix"),
				Display: filepath.Join(home, "bin"),
			},
			{
				Path:    "/opt/portunix/bin/portunix",
				Display: "/opt/portunix/bin",
			},
		}

	default: // linux and others
		home, _ := os.UserHomeDir()
		options := []installOption{
			{
				Path:        "/usr/local/bin/portunix",
				Display:     "/usr/local/bin",
				Recommended: checkWriteAccess("/usr/local/bin") == nil,
			},
			{
				Path:    filepath.Join(home, "bin", "portunix"),
				Display: filepath.Join(home, "bin"),
			},
			{
				Path:    filepath.Join(home, ".local", "bin", "portunix"),
				Display: filepath.Join(home, ".local", "bin"),
			},
		}

		// Add /opt if writable
		if checkWriteAccess("/opt") == nil {
			options = append(options, installOption{
				Path:    "/opt/portunix/bin/portunix",
				Display: "/opt/portunix/bin",
			})
		}

		return options
	}
}
