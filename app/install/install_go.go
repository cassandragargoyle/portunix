package install

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"portunix.cz/app"
)

// Extract the Go tarball
func extractGo(src, dest string) error {
	fmt.Println("Extracting Go archive...")

	// Open the tarball
	file, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open tarball: %v", err)
	}
	defer file.Close()

	if filepath.Ext(src) == ".zip" {
		return app.UnzipFile(src, dest)
	} else {
		// Read it as a gzip file
		gzipReader, err := gzip.NewReader(file)
		if err != nil {
			return fmt.Errorf("failed to read gzip: %v", err)
		}
		defer gzipReader.Close()

		// Extract tar contents
		tarReader := tar.NewReader(gzipReader)
		for {
			header, err := tarReader.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				return fmt.Errorf("failed to read tar header: %v", err)
			}

			// Determine target file path
			targetPath := filepath.Join(dest, header.Name)

			switch header.Typeflag {
			case tar.TypeDir:
				if err := os.MkdirAll(targetPath, os.ModePerm); err != nil {
					return fmt.Errorf("failed to create directory: %v", err)
				}
			case tar.TypeReg:
				outFile, err := os.Create(targetPath)
				if err != nil {
					return fmt.Errorf("failed to create file: %v", err)
				}
				if _, err := io.Copy(outFile, tarReader); err != nil {
					outFile.Close()
					return fmt.Errorf("failed to copy file: %v", err)
				}
				outFile.Close()
			default:
				fmt.Printf("Ignoring file of type %c in tarball\n", header.Typeflag)
			}
		}
	}
	fmt.Println("Extraction complete.")
	return nil
}

// Set up the Go environment
func setupGoEnv(goPath string) error {
	fmt.Println("Setting up Go environment...")
	bashrc := filepath.Join(os.Getenv("HOME"), ".bashrc")
	envVars := fmt.Sprintf("\nexport PATH=$PATH:%s/bin\n", goPath)

	// Append Go PATH to .bashrc
	file, err := os.OpenFile(bashrc, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to write to .bashrc: %v", err)
	}
	defer file.Close()

	_, err = file.WriteString(envVars)
	if err != nil {
		return fmt.Errorf("failed to write environment variables: %v", err)
	}

	fmt.Println("Go environment set up. Run `source ~/.bashrc` to apply changes.")
	return nil
}

func InstallGo(arguments []string) error {
	// Define Go version and URLs
	goVersion := "1.23.5"
	osType := runtime.GOOS
	archType := runtime.GOARCH

	downloadURL := ""
	tarballPath := ""
	extractPath := ""
	installPath := ""

	if osType == "windows" {
		downloadURL = fmt.Sprintf("https://go.dev/dl/go%s.%s-%s.zip", goVersion, osType, archType)
		tarballPath = fmt.Sprintf("go%s.%s-%s.zip", goVersion, osType, archType)
		//installPath = filepath.Join(os.Getenv("ProgramFiles"), "Go")
		extractPath = "C:\\"
		installPath = filepath.Join("C:\\", "Go")
	} else {
		downloadURL = fmt.Sprintf("https://golang.org/dl/go%s.%s-%s.tar.gz", goVersion, osType, archType)
		tarballPath = "/tmp/go.tar.gz"
		extractPath = "/usr/local"
		installPath = "/usr/local/go"
	}

	// Check if the tarball already exists
	if !app.FileExist(tarballPath) {
		// Download Go tarball
		if err := app.DownloadFile(tarballPath, downloadURL); err != nil {
			fmt.Println("Error downloading Go:", err)
			return err
		}
	}

	// Extract the tarball
	if err := extractGo(tarballPath, extractPath); err != nil {
		fmt.Println("Error extracting Go:", err)
		return err
	}

	// Set up Go environment
	if err := setupGoEnv(installPath); err != nil {
		fmt.Println("Error setting up Go environment:", err)
		return err
	}

	// Check if Go is installed
	fmt.Println("Go installation complete. Verifying installation...")
	cmd := exec.Command("go", "version")
	output, err := cmd.Output()
	if err != nil {
		fmt.Println("Go verification failed. Please check your installation.")
	} else {
		fmt.Printf("Go installed successfully: %s\n", string(output))
	}
	return nil
}

func WinInstallGo(arguments []string) error {
	// Define the URL for the Go MSI installer
	goVersion := "1.23.5"
	installerURL := fmt.Sprintf("https://go.dev/dl/go%s.windows-amd64.msi", goVersion)
	installer := fmt.Sprintf("go%s.windows-amd64.msi", goVersion)

	// Check if the installer already exists
	if !app.FileExist(installer) {
		fmt.Println("Download Go instaler ...")
		// Download the installer
		err := app.DownloadFile(installer, installerURL)

		if err != nil {
			fmt.Println("Error during download:", err)
			return err
		}
	}
	// Run the MSI installer
	err := runMSIInstaller(installer)
	if err != nil {
		fmt.Println("Error during installation:", err)
		return err
	}
	fmt.Println("Go installation completed. Please verify using `go version`.")
	return nil
}

// Function to execute the MSI installer
func runMSIInstaller(filePath string) error {
	fmt.Println("Running MSI installer...")

	// Execute the MSI installer
	cmd := exec.Command("msiexec", "/i", filePath, "/quiet", "/norestart")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("failed to run installer: %v", err)
	}

	fmt.Println("Installation completed successfully.")
	return nil
}
