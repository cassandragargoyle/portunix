package cmd

import (
	"fmt"

	"portunix.ai/portunix/src/pkg/platform"
)

// RunEnv executes the env command
func RunEnv(args []string) {
	os := platform.GetOS()
	arch := platform.GetArchitecture()

	// Platform-specific values
	var exe, slash, pathsep string
	if platform.IsWindows() {
		exe = ".exe"
		slash = "\\"
		pathsep = ";"
	} else {
		exe = ""
		slash = "/"
		pathsep = ":"
	}

	fmt.Println("# Platform variables for Makefile")
	fmt.Printf("OS=%s\n", os)
	fmt.Printf("ARCH=%s\n", arch)
	fmt.Printf("EXE=%s\n", exe)
	fmt.Printf("SLASH=%s\n", slash)
	fmt.Printf("PATHSEP=%s\n", pathsep)
}
