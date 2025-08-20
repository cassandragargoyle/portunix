package install

import (
	"archive/zip"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"portunix.cz/app"
)

func WinInstallPython(arguments []string) error {
	// Check OS
	osType := runtime.GOOS
	var downloadURL string
	var installer string
	//TODO: Read from arguments
	var version = "last"
	var last_version = "3.13.6"
	var useGUI = false
	var useEmbeddable = false
	
	// Check for parameters
	for _, arg := range arguments {
		if arg == "--gui" || arg == "-gui" {
			useGUI = true
		} else if arg == "--embeddable" || arg == "-embeddable" {
			useEmbeddable = true
		}
	}
	switch osType {
	case "windows":
		if useEmbeddable {
			// Embeddable Python (ZIP version)
			switch version {
			case "last":
				version = last_version
				installer = fmt.Sprintf("python-%s-embed-amd64.zip", version)
				downloadURL = fmt.Sprintf("https://www.python.org/ftp/python/%s/", version) + installer
			case "3.13.6":
				installer = "python-3.13.6-embed-amd64.zip"
				downloadURL = "https://www.python.org/ftp/python/3.13.6/" + installer
			case "3.11.5":
				installer = "python-3.11.5-embed-amd64.zip"
				downloadURL = "https://www.python.org/ftp/python/3.11.5/" + installer
			default:
				installer = fmt.Sprintf("python-%s-embed-amd64.zip", version)
				downloadURL = fmt.Sprintf("https://www.python.org/ftp/python/%s/", version) + installer
			}
		} else {
			// Standard Python installer (EXE version)
			switch version {
			case "last":
				version = last_version
				installer = fmt.Sprintf("python-%s-amd64.exe", version)
				downloadURL = fmt.Sprintf("https://www.python.org/ftp/python/%s/", version) + installer
			case "3.13.6":
				installer = "python-3.13.6-amd64.exe"
				downloadURL = "https://www.python.org/ftp/python/3.13.6/" + installer
			case "3.11.5":
				installer = "python-3.11.5-amd64.exe"
				downloadURL = "https://www.python.org/ftp/python/3.11.5/" + installer
			default:
				installer = fmt.Sprintf("python-%s-amd64.exe", version)
				downloadURL = fmt.Sprintf("https://www.python.org/ftp/python/%s/", version) + installer
			}
		}

	case "darwin":
		installer = "python-3.11.5-macos11.pkg"
		downloadURL = "https://www.python.org/ftp/python/3.11.5/" + installer
	case "linux":
		return LnxInstallPython(nil)
	default:
		return errors.New("Unsupported operating system!")
	}

	if useEmbeddable {
		// For embeddable, use cache system
		if err := ensurePythonEmbeddableCache(installer, downloadURL); err != nil {
			fmt.Printf("Error preparing Python embeddable: %v\n", err)
			return err
		}
	} else {
		// For regular installer, download to current directory
		if !app.FileExist(installer) {
			// download file
			fmt.Printf("Download from URL: %s\n", downloadURL)
			err := app.DownloadFile(downloadURL, installer)
			if err != nil {
				fmt.Println("Error downloading the installer file.", err)
				return err
			}
			fmt.Printf("Installer file %s downloaded.\n", installer)
		} else {
			fmt.Printf("Installer file %s exists.\n", installer)
		}
	}

	//TODO: Verify Python installed
	var err error

	switch osType {
	case "windows":
		if useEmbeddable {
			err = WinSetupPythonEmbeddable(installer)
		} else {
			err = WinSetupPython(installer, useGUI)
		}
	case "darwin":
		//TODO:
	}
	if err == nil {
		fmt.Println("Python has been successfully installed.")
	}
	return err
}

// downloadFile from URL end save file
func downloadFile(filepath string, url string) error {
	// Open URL
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, resp.Body)
	return err
}

// Function to run the Python installer
func WinSetupPython(installer string, useGUI bool) error {
	fmt.Printf("Installing Python from %s...\n", installer)
	
	var cmd *exec.Cmd
	
	if useGUI {
		// GUI installation
		fmt.Println("Starting Python installer with GUI...")
		fmt.Println("Please follow the installation wizard in the GUI window.")
		
		cmd = exec.Command("./"+installer)
	} else {
		// Silent installation
		fmt.Println("Using silent installation. This may take a few minutes. Please wait...")
		
		// Create context with timeout (10 minutes should be enough)
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()
		
		// Correct Python installer parameters for silent installation
		cmd = exec.CommandContext(ctx, "./"+installer, 
			"/quiet",                    // Silent installation
			"InstallAllUsers=1",         // Install for all users
			"PrependPath=1",            // Add to PATH
			"Include_test=0",           // Don't install test suite
			"Include_tcltk=0",          // Don't install Tcl/Tk
			"Include_launcher=1",       // Include py launcher
			"InstallLauncherAllUsers=1", // Launcher for all users
			"AssociateFiles=1",         // Associate .py files
			"CompileAll=0",             // Don't compile all .py files (faster)
			"SimpleInstall=1")          // Simple installation
	}
	
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	// Start the installation
	fmt.Printf("Starting Python installation...\n")
	start := time.Now()
	
	err := cmd.Run()
	
	duration := time.Since(start)
	fmt.Printf("Installation completed in %v\n", duration.Round(time.Second))
	
	if err != nil {
		if !useGUI {
			// Only check for timeout in silent mode (when using context)
			if strings.Contains(err.Error(), "context deadline exceeded") {
				fmt.Printf("Installation timed out after 10 minutes\n")
				return fmt.Errorf("installation timed out")
			}
		}
		fmt.Printf("Installation error: %s\n", err)
		return err
	}
	
	// Verify installation
	fmt.Println("Verifying Python installation...")
	verifyCmd := exec.Command("python", "--version")
	output, err := verifyCmd.Output()
	if err != nil {
		fmt.Println("Warning: Could not verify Python installation")
	} else {
		fmt.Printf("✓ Python installed successfully: %s", string(output))
	}
	
	return nil
}

// ensurePythonEmbeddableCache downloads Python embeddable to cache if not exists
func ensurePythonEmbeddableCache(installer, downloadURL string) error {
	// Cache directory for Python embeddable
	cacheDir := ".cache"
	pythonCacheDir := filepath.Join(cacheDir, "python-embeddable")
	cachedZipPath := filepath.Join(pythonCacheDir, installer)
	
	// Create cache directory if it doesn't exist
	if err := os.MkdirAll(pythonCacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create Python cache directory: %w", err)
	}
	
	// Check if Python ZIP is already cached
	if _, err := os.Stat(cachedZipPath); os.IsNotExist(err) {
		fmt.Printf("Downloading Python embeddable to cache: %s\n", downloadURL)
		
		// Download using the app.DownloadFile function
		if err := app.DownloadFile(downloadURL, cachedZipPath); err != nil {
			return fmt.Errorf("failed to download Python embeddable: %w", err)
		}
		fmt.Printf("Python embeddable cached: %s\n", cachedZipPath)
	} else {
		fmt.Printf("Python embeddable found in cache: %s\n", cachedZipPath)
	}
	
	return nil
}

// Function to setup Python embeddable version
func WinSetupPythonEmbeddable(zipFile string) error {
	fmt.Printf("Setting up Python embeddable...\n")
	
	// Target directory
	targetDir := "C:\\Python"
	
	// Create target directory
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return fmt.Errorf("failed to create target directory %s: %w", targetDir, err)
	}
	
	// Determine ZIP file path - prioritize sandbox cache, then local cache
	var zipPath string
	
	// First check if ZIP exists in sandbox python-embeddable directory
	sandboxZipPath := "C:\\Portunix\\python-embeddable\\" + filepath.Base(zipFile)
	if _, err := os.Stat(sandboxZipPath); err == nil {
		fmt.Printf("Using cached Python embeddable from sandbox: %s\n", sandboxZipPath)
		zipPath = sandboxZipPath
	} else {
		// Check local cache
		cacheZipPath := filepath.Join(".cache", "python-embeddable", filepath.Base(zipFile))
		if _, err := os.Stat(cacheZipPath); err == nil {
			fmt.Printf("Using cached Python embeddable from local cache: %s\n", cacheZipPath)
			zipPath = cacheZipPath
		} else {
			fmt.Printf("Using downloaded Python embeddable: %s\n", zipFile)
			zipPath = zipFile
		}
	}
	
	// Extract ZIP file
	fmt.Printf("Extracting %s to %s...\n", zipPath, targetDir)
	if err := extractZipFile(zipPath, targetDir); err != nil {
		return fmt.Errorf("failed to extract ZIP file: %w", err)
	}
	
	// Add to system PATH
	fmt.Println("Adding Python to system PATH...")
	if err := addToSystemPath(targetDir); err != nil {
		return fmt.Errorf("failed to add to system PATH: %w", err)
	}
	
	// Verify installation
	fmt.Println("Verifying Python embeddable installation...")
	verifyCmd := exec.Command("python", "--version")
	output, err := verifyCmd.Output()
	if err != nil {
		fmt.Println("Warning: Could not verify Python installation")
	} else {
		fmt.Printf("✓ Python embeddable installed successfully: %s", string(output))
	}
	
	return nil
}

// extractZipFile extracts a ZIP file to destination directory
func extractZipFile(src, dest string) error {
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	// Make destination directory
	os.MkdirAll(dest, 0755)

	// Extract files
	for _, f := range r.File {
		path := filepath.Join(dest, f.Name)

		if f.FileInfo().IsDir() {
			os.MkdirAll(path, f.FileInfo().Mode())
			continue
		}

		// Create parent directories
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}

		// Extract file
		rc, err := f.Open()
		if err != nil {
			return err
		}

		outFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.FileInfo().Mode())
		if err != nil {
			rc.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}

	return nil
}

// addToSystemPath adds a directory to the system PATH environment variable
func addToSystemPath(dir string) error {
	// Use PowerShell to add to system PATH
	psScript := fmt.Sprintf(`
		$currentPath = [Environment]::GetEnvironmentVariable("PATH", "Machine")
		if ($currentPath -notlike "*%s*") {
			$newPath = $currentPath + ";%s"
			[Environment]::SetEnvironmentVariable("PATH", $newPath, "Machine")
			Write-Host "Added %s to system PATH"
		} else {
			Write-Host "%s is already in system PATH"
		}
	`, dir, dir, dir, dir)
	
	cmd := exec.Command("powershell", "-Command", psScript)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to update PATH: %w\nOutput: %s", err, string(output))
	}
	
	fmt.Printf("PATH update result: %s\n", string(output))
	return nil
}
