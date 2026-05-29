/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
)

// proxmoxCmd delegates all `portunix proxmox ...` invocations to the
// ptx-proxmox helper binary. The helper carries the actual implementation
// (see src/helpers/ptx-proxmox/). If the helper is missing, print a
// pointer to the install flow rather than silently failing.
var proxmoxCmd = &cobra.Command{
	Use:                "proxmox",
	Short:              "Manage Proxmox VE hypervisors (issue #167)",
	Long:               `Manage VMs and LXC containers on remote Proxmox VE hypervisors via the Proxmox REST API. Delegates to the ptx-proxmox helper.`,
	DisableFlagParsing: true, // let ptx-proxmox parse its own flags
	RunE: func(cmd *cobra.Command, args []string) error {
		helperPath := findProxmoxHelper()
		if helperPath == "" {
			return fmt.Errorf("ptx-proxmox helper not found next to portunix — run `make build` or re-install")
		}
		return dispatchToProxmoxHelper(helperPath, args)
	},
}

// findProxmoxHelper locates ptx-proxmox binary next to the main portunix
// executable first (the normal install layout), then falls back to PATH so
// developer builds from arbitrary locations still work.
func findProxmoxHelper() string {
	helperName := "ptx-proxmox"
	if runtime.GOOS == "windows" {
		helperName += ".exe"
	}
	if execPath, err := os.Executable(); err == nil {
		candidate := filepath.Join(filepath.Dir(execPath), helperName)
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	if p, err := exec.LookPath(helperName); err == nil {
		return p
	}
	return ""
}

// dispatchToProxmoxHelper forwards args verbatim, prefixed with the `proxmox`
// subcommand name so cobra inside the helper routes correctly.
func dispatchToProxmoxHelper(helperPath string, args []string) error {
	helperArgs := append([]string{"proxmox"}, args...)
	cmd := exec.Command(helperPath, helperArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func init() {
	rootCmd.AddCommand(proxmoxCmd)
}
