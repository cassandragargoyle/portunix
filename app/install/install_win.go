package install

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"portunix.cz/app"
)

func InstallWin(arguments []string) {
	what := arguments[0]
	switch what {
	case "daemon":
		WinInstallDaemon()
	case "java":
		WinInstallJava(arguments[1:])
	case "python":
		WinInstallPython(arguments[1:])
	case "wsl":
		WinInstallWsl(arguments[1:])
	case "msvc":
		WinInstallMSBuildtools(arguments[1:])
	case "vscode":
		InstallVSCode(arguments[1:])
	// TODO: specified as vscodeextension:default
	case "vscodeextension":
		InstallVSCodeDefaultExtension()
	case "go":
		WinInstallGo(arguments[1:])
	case "default":
		WinInstallDefault()
	case "empty":
		WinInstallEmpty()
	default:
		fmt.Println("Unknown installation type:", what)
	}
}

func WinInstallJava(arguments []string) error {

	procesedArgs := ProcessArgumentsInstallJava(arguments)
	return WinInstallJavaRun(procesedArgs["version"], procesedArgs["variant"], false)
}

func WinInstallJavaRun(version string, variant string, dryRun bool) error {
	return WinInstallJavaHttpRun(version, variant, dryRun)
}

func WinInstallJavaHttpRun(version string, variant string, dryRun bool) error {

	defaultVariant := "openjdk"
	defaultMainVersion := "17"
	variablesMapping := map[string]string{}

	mainVersion := strings.Split(app.GetValueOrDefault(version, defaultMainVersion), ".")[0]

	key := fmt.Sprintf("%s_%s", mainVersion, app.GetValueOrDefault(variant, defaultVariant))

	variablesMapping["main_version"] = mainVersion

	javaVersionMap := map[string]string{
		"8_openjdk":  "8u432b06",
		"11_openjdk": "11.0.25_9",
		"17_openjdk": "17.0.13_11",
		"21_openjdk": "21.0.5_11",
	}

	// Check if the current variant and version has a corresponding mapping
	if javaVersion, exists := javaVersionMap[key]; exists {
		variablesMapping["version"] = javaVersion
		variablesMapping["versionPlus"] = strings.Replace(javaVersion, "_", "%2B", -1)
		variablesMapping["versionDot"] = strings.Replace(javaVersion, "_", ".", -1)

	} else {
		return errors.New(fmt.Sprintf("Unsupported Java variant: %s major version: %s .", variant, version))
	}

	javaDownloadURLMapping := map[string]string{
		"8_openjdk":  "https://github.com/adoptium/temurin${main_version}-binaries/releases/download/jdk-${versionPlus}/OpenJDK${main_version}U-jdk_${arch}_windows_hotspot_${version}.msi",
		"11_openjdk": "https://github.com/adoptium/temurin${main_version}-binaries/releases/download/jdk-${versionPlus}/OpenJDK${main_version}U-jdk_${arch}_windows_hotspot_${version}.msi",
		"17_openjdk": "https://github.com/adoptium/temurin${main_version}-binaries/releases/download/jdk-${versionPlus}/OpenJDK${main_version}U-jdk_${arch}_windows_hotspot_${version}.msi",
		"21_openjdk": "https://github.com/adoptium/temurin${main_version}-binaries/releases/download/jdk-${versionPlus}/OpenJDK${main_version}U-jdk_${arch}_windows_hotspot_${version}.msi",
	}

	// URL for downloading the JDK
	//javaDownloadURL := "https://github.com/adoptium/temurin17-binaries/releases/download/jdk-17.0.13%2B11/OpenJDK17U-jdk_x64_windows_hotspot_17.0.13_11.msi"
	//javaDownloadURL := "https://download.oracle.com/java/17/latest/jdk-17_windows-x64_bin.exe"

	var javaDownloadURL string
	// Check if the current variant and version has a corresponding mapping
	if val, exists := javaDownloadURLMapping[key]; exists {
		javaDownloadURL = val
	} else {
		return errors.New(fmt.Sprintf("Unsupported Java variant: %s major version: %s for download.", variant, version))
	}

	// Path to save the installer
	//fileName := "jdk-17_windows-x64_bin.exe"
	//fileName := "OpenJDK17U-jdk_x64_windows_hotspot_17.0.13_11.msi"

	javafileNameMapping := map[string]string{
		"8_openjdk":  "OpenJDK8U-jdk_${arch}_windows_hotspot_${version}.msi",
		"11_openjdk": "OpenJDK11U-jdk_${arch}_windows_hotspot_${version}.msi",
		"17_openjdk": "OpenJDK17U-jdk_${arch}_windows_hotspot_${version}.msi",
		"21_openjdk": "OpenJDK21U-jdk_${arch}_windows_hotspot_${version}.msi",
	}

	var fileName string
	// Check if the current variant and version has a corresponding mapping
	if val, exists := javafileNameMapping[key]; exists {
		fileName = val
	} else {
		return errors.New(fmt.Sprintf("Undefined Java file for variant: %s major version: %s.", variant, version))
	}

	javaHomeMapping := map[string]string{
		"8_openjdk":  "${Program Files}/Eclipse Adoptium/jdk-${version}-hotspot",
		"11_openjdk": "${Program Files}/Eclipse Adoptium/jdk-${versionDot}-hotspot",
		"17_openjdk": "${Program Files}/Eclipse Adoptium/jdk-${version}-hotspot",
		"21_openjdk": "${Program Files}/Eclipse Adoptium/jdk-${version}-hotspot",
	}

	//TODO: x86
	variablesMapping["Program Files"] = os.Getenv("ProgramFiles")

	var javaHome string
	// Check if the current variant and version has a corresponding mapping
	if val, exists := javaHomeMapping[key]; exists {
		javaHome = val
	} else {
		javaHome = ""
		log.Printf(fmt.Sprintf("Undefined Java HOME for variant: %s major version: %s.", variant, version))
	}

	// Map Go architectures to corresponding file name architecture strings
	archMapping := map[string]string{
		"amd64":   "x64",
		"arm64":   "aarch64",
		"386":     "x86",
		"ppc64le": "ppc64le",
	}

	currentArch := runtime.GOARCH
	// Check if the current architecture has a corresponding mapping
	if mappedArch, exists := archMapping[currentArch]; exists {
		variablesMapping["arch"] = mappedArch
	} else {
		return errors.New(fmt.Sprintf("Unsupported system architecture: %s for Java installation.", currentArch))
	}

	// replaceTemplateVariables
	fileName = app.ReplaceTemplateVariables(fileName, variablesMapping)
	javaHome = app.ReplaceTemplateVariables(javaHome, variablesMapping)
	//https://github.com/adoptium/temurin11-binaries/releases/download/jdk-11.0.25%2B9/OpenJDK11U-jdk_x64_windows_hotspot_11.0.25_9.msi
	javaDownloadURL = app.ReplaceTemplateVariables(javaDownloadURL, variablesMapping)

	//info https://adoptium.net/installation/windows/

	// Get the current working directory
	cwd, err := os.Getwd()
	if err != nil {
		log.Println("Error getting current directory:", err)
		return errors.New(fmt.Sprintf("Error getting current directory:%s", err))
	}

	// Check cache directory first
	cacheDir := ".cache"
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}
	
	cachedInstallerPath := filepath.Join(cacheDir, fileName)
	installerPath := filepath.Join(cwd, fileName)

	// Check if the file exists in cache first
	if app.FileExist(cachedInstallerPath) {
		fmt.Printf("Using cached Java installer: %s\n", cachedInstallerPath)
		installerPath = cachedInstallerPath
	} else if app.FileExist(installerPath) {
		fmt.Printf("File '%s' already exists in the current directory.\n", fileName)
	} else {
		fmt.Println("Downloading Java installer...")
		err := app.DownloadFile(javaDownloadURL, cachedInstallerPath)
		if err != nil {
			log.Printf("Error downloading Java: %s", err)
			return err
		}
		fmt.Printf("Java installer cached: %s\n", cachedInstallerPath)
		installerPath = cachedInstallerPath
	}

	// Verify Java installation
	fmt.Println("Verify current Java installation")
	cmd := exec.Command("java", "-version")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err == nil {
		fmt.Println("Java is already installed.")
		//TODO: enable higher version to install
		return nil
	}
	if !dryRun {
		fmt.Println("Running the installer...")
		// "/s" enables silent installation
		//if exe
		//cmd = exec.Command(installerPath, "/s")
		//features := "ADDLOCAL=FeatureMain,FeatureEnvironment,FeatureJarFileRunWith,FeatureJavaHome"
		features := "ADDLOCAL=ALL"
		cmd = exec.Command("msiexec", "/i", installerPath, features, "/quiet")

		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		err = cmd.Run()
		if err != nil {
			log.Printf("Error during installation:%s", err)
			return err
		}

		//simulate enviroment variables setting
		if javaHome != "" {
			// Set JAVA_HOME after installation
			err = os.Setenv("JAVA_HOME", javaHome)
			if err != nil {
				fmt.Println("Error setting JAVA_HOME:", err)
				return err
			}

			// Add JAVA_HOME/bin to PATH
			currentPath := os.Getenv("PATH")
			os.Setenv("PATH", currentPath+";"+os.Getenv("JAVA_HOME")+"/bin")

			// Verify the updated environment
			fmt.Println("JAVA_HOME set to:", os.Getenv("JAVA_HOME"))
			fmt.Println("Updated PATH:", os.Getenv("PATH"))
		}
	}
	fmt.Println("Installation complete. Verifying Java version...")

	// Verify Java installation
	cmd = exec.Command("java", "-version")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		log.Printf("Java installation verification failed:%s", err)
		return err
	} else {
		fmt.Println("Java was successfully installed.")
	}
	return nil
}

func WinInstallDefault() {
	fmt.Println("Installing default packages: Python embedded, Java, and VSCode...")
	
	// Install Python embedded
	fmt.Println("1/3 Installing Python embedded...")
	WinInstallPython([]string{"--embeddable"})
	
	// Install Java
	fmt.Println("2/3 Installing Java...")
	WinInstallJava([]string{})
	
	// Install VSCode
	fmt.Println("3/3 Installing VSCode...")
	InstallVSCode([]string{})
	
	fmt.Println("Default installation completed!")
}

func WinInstallEmpty() {
	fmt.Println("Empty installation - no packages will be installed.")
	fmt.Println("Sandbox is ready for manual operations.")
}

func WinInstallDaemon() error {
	fmt.Println("TODO: Install daemon on windows.")
	return nil
}

func WinInstallWsl(arguments []string) error {
	if !isInstalledByVersion("wsl") {
		//Instal wsl
	}
	cmd := exec.Command("wsl", "--install")

	fmt.Println("Installing WSL...")
	output, err := cmd.CombinedOutput()

	if err != nil {
		fmt.Printf("Failed to install WSL: %v\n", err)
		return err
	} else {
		fmt.Printf("Failed to install WSL: %v\n", string(output))
	}
	fmt.Println("WSL installation completed successfully!")
	return nil
}

func isInstalledByVersion(program string) bool {
	// Check if the "wsl" command is available
	cmd := exec.Command(program, "--version")
	if err := cmd.Run(); err != nil {
		// program is not installed
		return false
	}
	// program is installed
	return true
}

