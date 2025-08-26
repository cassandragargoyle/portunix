package selfinstall

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
)

// Options represents installation options
type Options struct {
	SourcePath   string
	TargetPath   string
	CreateConfig bool
	AddToPath    bool
	Silent       bool
}

// InstallSilent performs silent installation with provided options
func InstallSilent(options Options) error {
	fmt.Printf("Installing Portunix to %s...\n", options.TargetPath)
	
	// Create target directory if it doesn't exist
	targetDir := filepath.Dir(options.TargetPath)
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}
	
	// Copy binary to target location
	if err := copyFile(options.SourcePath, options.TargetPath); err != nil {
		return fmt.Errorf("failed to copy binary: %w", err)
	}
	
	// Make executable on Unix systems
	if runtime.GOOS != "windows" {
		if err := os.Chmod(options.TargetPath, 0755); err != nil {
			return fmt.Errorf("failed to set executable permissions: %w", err)
		}
	}
	
	// Add to PATH if requested
	if options.AddToPath {
		if err := AddToSystemPath(targetDir); err != nil {
			fmt.Printf("⚠ Warning: Failed to add to PATH: %v\n", err)
			fmt.Printf("  Please add %s to your PATH manually\n", targetDir)
		} else {
			fmt.Println("✓ Added to system PATH")
		}
	}
	
	// Create config if requested
	if options.CreateConfig {
		if err := createDefaultConfig(); err != nil {
			fmt.Printf("⚠ Warning: Failed to create config: %v\n", err)
		} else {
			fmt.Println("✓ Created default configuration")
		}
	}
	
	// Verify installation
	if err := VerifyInstallation(options.TargetPath); err != nil {
		return fmt.Errorf("installation verification failed: %w", err)
	}
	
	fmt.Println("\n✓ Installation completed successfully!")
	fmt.Printf("  Run '%s --version' to verify\n", filepath.Base(options.TargetPath))
	
	return nil
}

// InstallInteractive performs interactive installation
func InstallInteractive(sourcePath string) error {
	fmt.Println("Welcome to Portunix Installation!")
	fmt.Printf("Version: %s\n\n", getVersion())
	
	// Prompt for installation location
	targetPath, err := PromptInstallLocation()
	if err != nil {
		return fmt.Errorf("installation cancelled: %w", err)
	}
	
	if targetPath == "" {
		fmt.Println("Installation cancelled.")
		return nil
	}
	
	// Check if target exists
	if _, err := os.Stat(targetPath); err == nil {
		backup, err := PromptBackup(targetPath)
		if err != nil {
			return err
		}
		if backup {
			backupPath := targetPath + ".backup"
			if err := os.Rename(targetPath, backupPath); err != nil {
				return fmt.Errorf("failed to create backup: %w", err)
			}
			fmt.Printf("✓ Created backup at %s\n", backupPath)
		}
	}
	
	// Copy binary
	fmt.Printf("Installing to %s...\n", targetPath)
	if err := copyFile(sourcePath, targetPath); err != nil {
		return fmt.Errorf("failed to copy binary: %w", err)
	}
	
	// Make executable on Unix systems
	if runtime.GOOS != "windows" {
		if err := os.Chmod(targetPath, 0755); err != nil {
			return fmt.Errorf("failed to set executable permissions: %w", err)
		}
	}
	
	fmt.Println("✓ Binary installed successfully")
	
	// Prompt for PATH addition
	addPath, err := PromptAddToPath()
	if err != nil {
		fmt.Printf("⚠ Warning: %v\n", err)
	} else if addPath {
		targetDir := filepath.Dir(targetPath)
		if err := AddToSystemPath(targetDir); err != nil {
			fmt.Printf("⚠ Warning: Failed to add to PATH: %v\n", err)
			fmt.Printf("  Please add %s to your PATH manually\n", targetDir)
		} else {
			fmt.Println("✓ Added to system PATH")
		}
	}
	
	// Prompt for config creation
	createConfig, err := PromptCreateConfig()
	if err != nil {
		fmt.Printf("⚠ Warning: %v\n", err)
	} else if createConfig {
		if err := createDefaultConfig(); err != nil {
			fmt.Printf("⚠ Warning: Failed to create config: %v\n", err)
		} else {
			fmt.Println("✓ Created default configuration")
		}
	}
	
	// Verify installation
	if err := VerifyInstallation(targetPath); err != nil {
		fmt.Printf("⚠ Warning: Installation verification failed: %v\n", err)
	} else {
		fmt.Println("✓ Installation verified")
	}
	
	// Show summary
	ShowInstallationSummary(targetPath)
	
	return nil
}

// VerifyInstallation verifies that the installation was successful
func VerifyInstallation(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("binary not found at %s", path)
	}
	
	if info.IsDir() {
		return fmt.Errorf("%s is a directory, not a file", path)
	}
	
	// Check if executable on Unix
	if runtime.GOOS != "windows" {
		if info.Mode()&0111 == 0 {
			return fmt.Errorf("binary at %s is not executable", path)
		}
	}
	
	return nil
}

// GetDefaultInstallPath returns the default installation path for the current OS
func GetDefaultInstallPath() string {
	switch runtime.GOOS {
	case "windows":
		// Try Program Files first, then user local
		progFiles := os.Getenv("PROGRAMFILES")
		if progFiles != "" {
			return filepath.Join(progFiles, "Portunix", "portunix.exe")
		}
		return filepath.Join(os.Getenv("LOCALAPPDATA"), "Portunix", "portunix.exe")
	case "darwin":
		return "/usr/local/bin/portunix"
	default: // linux and others
		// Check if we have write access to /usr/local/bin
		if err := checkWriteAccess("/usr/local/bin"); err == nil {
			return "/usr/local/bin/portunix"
		}
		// Fall back to user's home bin
		home, _ := os.UserHomeDir()
		return filepath.Join(home, "bin", "portunix")
	}
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()
	
	// Create destination directory if needed
	dstDir := filepath.Dir(dst)
	if err := os.MkdirAll(dstDir, 0755); err != nil {
		return err
	}
	
	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()
	
	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}
	
	// Copy file permissions
	sourceInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	
	return os.Chmod(dst, sourceInfo.Mode())
}

// checkWriteAccess checks if we have write access to a directory
func checkWriteAccess(dir string) error {
	testFile := filepath.Join(dir, ".portunix-test")
	file, err := os.Create(testFile)
	if err != nil {
		return err
	}
	file.Close()
	os.Remove(testFile)
	return nil
}

// createDefaultConfig creates default configuration files
func createDefaultConfig() error {
	configDir, err := getConfigDir()
	if err != nil {
		return err
	}
	
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}
	
	configFile := filepath.Join(configDir, "config.yaml")
	
	// Check if config already exists
	if _, err := os.Stat(configFile); err == nil {
		return nil // Config already exists
	}
	
	// Create default config
	defaultConfig := `# Portunix Configuration
# Generated during installation

# Default settings
verbose: false
auto_update: true
update_channel: stable
`
	
	return os.WriteFile(configFile, []byte(defaultConfig), 0644)
}

// getConfigDir returns the configuration directory path
func getConfigDir() (string, error) {
	switch runtime.GOOS {
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			return "", fmt.Errorf("APPDATA environment variable not set")
		}
		return filepath.Join(appData, "Portunix"), nil
	case "darwin":
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, "Library", "Application Support", "Portunix"), nil
	default: // linux and others
		configHome := os.Getenv("XDG_CONFIG_HOME")
		if configHome != "" {
			return filepath.Join(configHome, "portunix"), nil
		}
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return filepath.Join(home, ".config", "portunix"), nil
	}
}

// getVersion returns the current version
func getVersion() string {
	// This will be set by the update module
	return "v1.4.0"
}