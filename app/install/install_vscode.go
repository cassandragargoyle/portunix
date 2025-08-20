package install

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"portunix.cz/app"
)

func InstallVSCode(arguments []string) error {
	isVSCodeInstalled, err := isVSCodeInstalled()
	if err != nil {
		fmt.Printf("Error determining if VS Code is installed: %v\n", err)
		return err
	}
	if !isVSCodeInstalled {
		// Determine the appropriate download URL based on the OS and architecture
		downloadURL, err := getVSCodeDownloadURL()
		if err != nil {
			fmt.Printf("Error determining the download URL: %v\n", err)
			return err
		}

		installer := "vscode_installer"

		switch runtime.GOOS {
		case "windows":
			installer += ".exe"
		}

		// Use cache directory for VSCode installer
		cacheDir := ".cache"
		vscodeCacheDir := filepath.Join(cacheDir, "vscode")
		cachedInstaller := filepath.Join(vscodeCacheDir, installer)

		// Create cache directory if it doesn't exist
		if err := os.MkdirAll(vscodeCacheDir, 0755); err != nil {
			fmt.Printf("Error creating cache directory: %v\n", err)
			return err
		}

		if !app.FileExist(cachedInstaller) {
			// Download the VS Code installer to cache
			fmt.Println("Downloading VS Code installer to cache...")
			err = app.DownloadFile(downloadURL, cachedInstaller)
			if err != nil {
				fmt.Printf("Error while downloading: %v\n", err)
				return err
			}
			fmt.Println("Download complete.")
		} else {
			fmt.Println("VS Code installer found in cache")
		}

		// Run the installer from cache
		fmt.Println("Running the VS Code installer...")
		err = runInstaller(cachedInstaller)
		if err != nil {
			fmt.Printf("Error during installation: %v\n", err)
			return err
		}
	}
	fmt.Println("Install default extension ...")
	InstallVSCodeDefaultExtension()
	fmt.Println("Installation complete.")
	return nil
}

func isVSCodeInstalled() (bool, error) {
	// Attempt to locate the `code` command
	_, err := exec.LookPath("code")
	if err != nil {
		// Return false if the command is not found
		return false, nil
	}
	return true, nil
}

// getVSCodeDownloadURL determines the appropriate URL for downloading VS Code
func getVSCodeDownloadURL() (string, error) {
	switch runtime.GOOS {
	case "windows":
		return "https://update.code.visualstudio.com/latest/win32-x64-user/stable", nil
	case "darwin":
		return "https://code.visualstudio.com/sha/download?build=stable&os=darwin-universal", nil
	case "linux":
		return "https://code.visualstudio.com/sha/download?build=stable&os=linux-deb-x64", nil
	default:
		return "", fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

// runInstaller executes the downloaded installer
func runInstaller(installerPath string) error {
	switch runtime.GOOS {
	case "windows":
		// Use absolute path if it's a cache path, otherwise use relative path
		var cmdPath string
		if filepath.IsAbs(installerPath) {
			cmdPath = installerPath
		} else {
			cmdPath = "./" + installerPath
		}
		cmd := exec.Command(cmdPath, "/silent", "/verysilent")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	case "darwin":
		cmd := exec.Command("open", installerPath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	case "linux":
		cmd := exec.Command("sudo", "dpkg", "-i", installerPath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	default:
		return fmt.Errorf("unsupported operating system: %s", runtime.GOOS)
	}
}

func InstallVSCodeDefaultExtension() error {
	err := InstallVSCodeExtension("C/C++", "ms-vscode.cpptools")
	if err != nil {
		return err
	}
	return nil
}

func InstallVSCodeExtension(label string, extension string) error {
	isInstaled, err := isExtensionInstalled(extension)
	if err != nil {
		return err
	}
	if !isInstaled {
		fmt.Printf("Installing %s extension...\n", label)
		cmd := exec.Command("code", "--install-extension", extension)
		output, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("Failed to install extension '%s': %v\nOutput: %s", extension, err, string(output))
			return err
		} else {
			fmt.Printf("%s extension installed successfully.\n", label)
		}
	} else {
		fmt.Printf("%s extension already installed.\n", label)
	}
	return nil
}

func isExtensionInstalled(extension string) (bool, error) {
	// Run the `code --list-extensions` command
	cmd := exec.Command("code", "--list-extensions")
	var output bytes.Buffer
	cmd.Stdout = &output

	if err := cmd.Run(); err != nil {
		fmt.Errorf("Failed to list extensions: %v", err)
		return false, err
	}

	// Read the output and check if it contains the searched extension ID
	scanner := bufio.NewScanner(&output)
	for scanner.Scan() {
		if strings.TrimSpace(scanner.Text()) == extension {
			return true, nil
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Errorf("Error reading extensions output: %v", err)
		return false, err
	}
	return false, nil
}