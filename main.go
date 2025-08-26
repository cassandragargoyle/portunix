// Portunus (jméno římského boha klíčů a dveří, symbolizujícího průchod kamkoliv).

package main

import (
	_ "embed"

	"portunix.cz/app/install"
	"portunix.cz/app/sandbox"
	"portunix.cz/app/update"
	appversion "portunix.cz/app/version"
	"portunix.cz/cmd"
)

//go:embed assets/scripts/windows/Install-PortableOpenSSH.ps1
var installOpenSSHScript string

//go:embed assets/scripts/windows/VSCodeInstall.cmd
var vscodeInstallScript string

//go:embed assets/scripts/windows/PortunixSystem.ps1
var portunixSystemPSScript string

//go:embed assets/install-packages.json
var installPackagesConfig string

const trace = true

// Version will be set at build time using ldflags
var version = "dev"

func main() {
	// Set the version for update package and version package
	update.Version = version
	appversion.ProductVersion = version

	// Set the embedded scripts in sandbox package
	sandbox.InstallOpenSSHScript = installOpenSSHScript
	sandbox.VSCodeInstallScript = vscodeInstallScript
	sandbox.PortunixSystemPSScript = portunixSystemPSScript

	// Set the embedded install config
	install.DefaultInstallConfig = installPackagesConfig

	cmd.Execute()
}
