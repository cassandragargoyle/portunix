package install

import (
	"fmt"
	"os"
	"os/exec"

	"portunix.cz/app"
)

func WinInstallMSBuildtools(arguments []string) error {
	// URL for downloading the Visual Studio Build Tools installer
	msvcURL := "https://aka.ms/vs/17/release/vs_buildtools.exe"
	installer := "vs_buildtools.exe"

	if !app.FileExist(installer) {
		// Download the installer
		fmt.Println("Downloading vs_BuildTools installer...")
		err := app.DownloadFile(installer, msvcURL)
		if err != nil {
			fmt.Printf("Error while downloading: %v\n", err)
			return err
		}
		fmt.Println("Download complete.")
	}else {
		fmt.Printf("Installer file %s exists.\n", installer)
	}

	// Run the installer
	fmt.Println("Running the vs_BuildTools installer...")
	var err = WinSetupMSBuildtools(installer)
	if err != nil {
		fmt.Println("Installation error.")
	} else {
		fmt.Println("Installation complete.")
	}

	return nil
}

func WinSetupMSBuildtools(installer string) error {
	fmt.Printf("Install file %s.\n", installer)
	cmd := exec.Command("./" + installer, "--quiet", "--wait", "--add", "Microsoft.VisualStudio.Workload.VCTools", "--includeRecommended", "--norestart")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		fmt.Printf("Installation error %s.\n", err)
	}
	return err
}