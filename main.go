package main

import (
	_ "embed"
	"errors"
	"fmt"
	"os"
	"os/exec"

	"portunix.ai/app/sandbox"
	"portunix.ai/app/update"
	appversion "portunix.ai/app/version"
	"portunix.ai/cmd"
	"portunix.ai/portunix/src/dispatcher"
)

//go:embed assets/scripts/windows/Install-PortableOpenSSH.ps1
var installOpenSSHScript string

//go:embed assets/scripts/windows/VSCodeInstall.cmd
var vscodeInstallScript string

//go:embed assets/scripts/windows/PortunixSystem.ps1
var portunixSystemPSScript string

// Version will be set at build time using ldflags.
var version = "dev"

func main() {
	// Set the version for update package and version package
	update.Version = version
	appversion.ProductVersion = version

	// Initialize dispatcher
	disp := dispatcher.NewDispatcher(version)

	// Check if we should dispatch to a helper binary
	args := os.Args[1:]
	if helperPath, shouldDispatch := disp.ShouldDispatch(args); shouldDispatch {
		if err := disp.Dispatch(helperPath, args); err != nil {
			// Propagate the helper's exit code instead of masking it as 1.
			// os/exec returns *ExitError for non-zero child exits; unwrap
			// and forward the status so shells observe what the helper
			// actually returned (e.g. `portunix ssh exec … "exit 42"` ⇒ 42).
			var ee *exec.ExitError
			if errors.As(err, &ee) {
				os.Exit(ee.ExitCode())
			}
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Set the version for root command (for --version flag)
	cmd.SetVersion()

	// Set the embedded scripts in sandbox package
	sandbox.InstallOpenSSHScript = installOpenSSHScript
	sandbox.VSCodeInstallScript = vscodeInstallScript
	sandbox.PortunixSystemPSScript = portunixSystemPSScript

	// Normal command execution - always show help when no arguments
	cmd.Execute()
}
