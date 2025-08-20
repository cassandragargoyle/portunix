package install

import (
	"runtime"
	"strconv"
	"strings"

	"portunix.cz/app"
	"portunix.cz/app/docker"
	"portunix.cz/app/podman"
)

func ToArguments(what string) []string {
	arguments := make([]string, 1)
	arguments[0] = what
	return arguments
}

func Install(arguments []string) {
	if len(arguments) == 0 {
		return
	}

	// Special handling for Docker and Podman installation
	packageName := arguments[0]
	if packageName == "docker" {
		// Check for -y flag
		autoAccept := false
		for _, arg := range arguments[1:] {
			if arg == "-y" {
				autoAccept = true
				break
			}
		}

		if err := docker.InstallDocker(autoAccept); err != nil {
			return
		}
		return
	}

	if packageName == "podman" {
		// Check for -y flag
		autoAccept := false
		for _, arg := range arguments[1:] {
			if arg == "-y" {
				autoAccept = true
				break
			}
		}

		if err := podman.InstallPodman(autoAccept); err != nil {
			return
		}
		return
	}

	// Try new JSON-based installer first
	variant := ""

	// Check if variant is specified
	if len(arguments) > 1 {
		for _, arg := range arguments[1:] {
			if arg != "--gui" && arg != "--embeddable" && !strings.HasPrefix(arg, "--") {
				variant = arg
				break
			}
		}
	}

	// Try to install using new system
	if err := InstallPackage(packageName, variant); err == nil {
		return // Success with new system
	}

	// Fall back to old system
	os := runtime.GOOS
	if os == "linux" {
		InstallLnx(arguments)
	} else if os == "windows" {
		InstallWin(arguments)
	} else {
		//TODO:
	}
}

func ProcessArgumentsInstall(arguments []string) (map[string]string, []string) {
	//TODO: use list
	enabledArguments := []string{"version", "variant"}
	return app.ProcessArguments(arguments, enabledArguments)
}

func ProcessArgumentsInstallJava(arguments []string) map[string]string {
	argsMap, other := ProcessArgumentsInstall(arguments)
	// Check if the first 'other' argument is a version
	if len(other) > 0 {
		if _, err := strconv.Atoi(other[0]); err == nil {
			argsMap["version"] = other[0]
			other = other[1:] // Remove the version from 'other'
		}
	}
	for _, str := range other {
		if str == "openjdk" {
			argsMap["variant"] = str
		}
	}
	return argsMap
}
