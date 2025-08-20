package install

import (
	"fmt"
	"os"
	"os/exec"

	"portunix.cz/app"
)

const defaultInstallPathOpenHabian = "/opt/home/"
const defaultInstallPathDebian = "/usr/share/home/"

const daemonAppName = "daemon"

func InstallLnx(arguments []string) {
	what := arguments[0]
	switch what {
	case "all":
	case "manager":
	case "java":
		LnxInstallJava(arguments[1:])
	case "python":
		LnxInstallPython(arguments[1:])
	case "daemon":
		LnxInstallDaemon(arguments[1:])
	case "service":
		// app.InstallServiceLnx("8005")
	case "go":
		InstallGo(arguments[1:])
	}
}

func LnxInstallJava(arguments []string) error {
	procesedArgs := ProcessArgumentsInstallJava(arguments)
	return LnxInstallJavaRun(procesedArgs["version"], procesedArgs["variant"], false)
}

func LnxInstallJavaRun(version string, variant string, dryRun bool) error {
	output, err := LnxExecAptCommand("update")

	if err != nil {
		fmt.Println("Error:", err)
		return err
	}

	// Print the output of the command
	fmt.Println(string(output))

	var javaPackage string

	if variant == "" || variant == "openjdk" {
		switch version {
		case "8":
			javaPackage = variant + "-8-jdk"
		case "11":
			javaPackage = variant + "-11-jdk"
		case "17":
			javaPackage = variant + "-17-jdk"
		default:
			javaPackage = variant + "-11-jdk"
		}

		output, err = LnxExecRootCommand("apt", "install", javaPackage)

		if err != nil {
			fmt.Println("Error:", err)
			return err
		}

		// Print the output of the command
		fmt.Println(string(output))
	}
	//TODO:Other variants
	return nil
}

func LnxInstallPython(arguments []string) error {
	fmt.Println("Instaling python on Linux with apt ...")
	output, err := LnxExecAptCommand("update")
	fmt.Println(string(output))
	if err != nil {
		return err
	}
	output, err = LnxExecAptCommand("install", "-y", "python3")
	fmt.Println(string(output))
	return err
}

func LnxExecRootCommand(args ...string) ([]byte, error) {
	var cmd *exec.Cmd
	if !LnxIsRoot() {
		cmd = exec.Command("sudo", args...)
	} else {
		cmd = exec.Command(args[0], args[1:]...)
	}

	// Execute the command
	output, err := cmd.CombinedOutput()
	return output, err
}

func LnxIsRoot() bool {
	return os.Geteuid() == 0
}

func LnxExecAptCommand(args ...string) ([]byte, error) {
	if err := LnxCheckApt(); err != nil {
		return nil, err
	}
	var commandArgs []string
	// Create a slice starting with "apt"
	if len(args) > 0 && args[0] != "apt" {
		commandArgs = []string{"apt"}
	} else {
		commandArgs = args
	}

	// Append variadic arguments to the slice
	commandArgs = append(commandArgs, args...)
	return LnxExecRootCommand(commandArgs...)
}

func LnxCheckApt() error {
	// Check if the 'apt' command is available
	if _, err := exec.LookPath("apt"); err != nil {
		return fmt.Errorf("'apt' tool is not available. Make sure you're on a Debian/Ubuntu-based system.")
	}
	return nil
}

func LnxInstallDaemon(arguments []string) error {

	osName, err := app.GetOSName()
	if err != nil {
		return err
	}

	var installPath string

	switch osName {
	case "Debian":
	case "Raspbian":
		installPath = defaultInstallPathDebian
	case "openHABian":
		installPath = defaultInstallPathOpenHabian
	}
	fmt.Printf("Instalation path:" + installPath)
	return nil
}
