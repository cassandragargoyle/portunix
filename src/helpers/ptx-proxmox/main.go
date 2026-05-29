/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

// ptx-proxmox is the Portunix helper binary for Proxmox VE management
// (issue #167). Phases 1–4 implemented:
//   - Phase 1: auth (login / status / logout / list)
//   - Phase 2: VM/CT lifecycle (list, info, status, start, stop, restart,
//     shutdown, delete, create vm/ct, template list, cloud-init)
//   - Phase 3: snapshot create/list/revert/delete, resolve-ip, ssh, exec, copy
//   - Phase 4: .ptxbook `environment.type: proxmox` with auto-create
//     (integrated in ptx-ansible; this helper supplies the underlying CLI)
package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var version = "dev"

// proxmoxCmd is the subcommand surface invoked via the dispatcher as
// `portunix proxmox ...`. The root helper command (ptx-proxmox) sits above it
// so cobra's help routing matches the main binary.
var proxmoxCmd = &cobra.Command{
	Use:   "proxmox",
	Short: "Manage Proxmox VE hypervisors (issue #167)",
	Long: `Manage VMs and LXC containers on remote Proxmox VE hypervisors via
the Proxmox REST API. Supports connection profiles (auth login/status/logout),
VM and CT lifecycle (list, info, status, start, stop, restart, delete),
creation from templates with cloud-init, snapshots, and SSH/SCP with automatic
IP resolution via QEMU guest agent or LXC interfaces. .ptxbook playbooks can
target this helper via 'environment.type: proxmox' (handled by ptx-ansible).`,
}

var rootCmd = &cobra.Command{
	Use:     "ptx-proxmox",
	Short:   "Portunix Proxmox VE Management Helper",
	Long:    `ptx-proxmox is a Portunix helper binary for Proxmox VE management. Invoke it through the main dispatcher as 'portunix proxmox ...'.`,
	Version: version,
}

// showHelpAI emits a machine-readable command inventory for the --help-ai
// convention used by other helpers (ptx-virt, ptx-trace). It matches the
// structure the main dispatcher expects when building AI documentation.
func showHelpAI() {
	type cmdInfo struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	type aiHelp struct {
		Tool        string    `json:"tool"`
		Version     string    `json:"version"`
		Description string    `json:"description"`
		Commands    []cmdInfo `json:"commands"`
	}
	help := aiHelp{
		Tool:        "ptx-proxmox",
		Version:     version,
		Description: "Remote Proxmox VE management: auth, VM/CT lifecycle, snapshots, SSH",
		Commands: []cmdInfo{
			{Name: "proxmox auth login", Description: "Save a Proxmox VE connection profile"},
			{Name: "proxmox auth status", Description: "Show active profile and validate credentials"},
			{Name: "proxmox auth logout", Description: "Remove a stored Proxmox profile"},
			{Name: "proxmox auth list", Description: "List stored Proxmox profiles"},
			{Name: "proxmox list", Description: "List VMs and LXC containers"},
			{Name: "proxmox info", Description: "Show VM/CT configuration"},
			{Name: "proxmox status", Description: "Show VM/CT runtime status"},
			{Name: "proxmox start", Description: "Start a VM or CT"},
			{Name: "proxmox stop", Description: "Hard-stop a VM or CT"},
			{Name: "proxmox shutdown", Description: "Send graceful shutdown signal"},
			{Name: "proxmox restart", Description: "Reboot a VM or CT"},
			{Name: "proxmox delete", Description: "Destroy a VM or CT"},
			{Name: "proxmox template list", Description: "List available templates and ISOs"},
			{Name: "proxmox create vm", Description: "Create a QEMU VM (optionally from template)"},
			{Name: "proxmox create ct", Description: "Create an LXC container from template"},
			{Name: "proxmox cloud-init", Description: "Configure cloud-init on a VM"},
			{Name: "proxmox snapshot create", Description: "Create a snapshot"},
			{Name: "proxmox snapshot list", Description: "List snapshots"},
			{Name: "proxmox snapshot revert", Description: "Revert a VM/CT to a snapshot"},
			{Name: "proxmox snapshot delete", Description: "Delete a snapshot"},
			{Name: "proxmox ssh", Description: "Open interactive SSH to a VM/CT (IP resolved via agent)"},
			{Name: "proxmox exec", Description: "Run a command over SSH"},
			{Name: "proxmox copy", Description: "SCP a file to/from a VM/CT"},
		},
	}
	data, _ := json.MarshalIndent(help, "", "  ")
	fmt.Println(string(data))
}

func init() {
	rootCmd.AddCommand(proxmoxCmd)
	proxmoxCmd.AddCommand(newAuthCmd())

	// Phase 2a: read-only inspection
	proxmoxCmd.AddCommand(newListCmd())
	proxmoxCmd.AddCommand(newInfoCmd())
	proxmoxCmd.AddCommand(newStatusCmd())

	// Phase 2b: lifecycle
	proxmoxCmd.AddCommand(newLifecycleCmd("start", "Start a VM or CT",
		func(c *Client, r *ResourceRef) (string, error) { return c.Start(r) }))
	proxmoxCmd.AddCommand(newLifecycleCmd("stop", "Hard-stop a VM or CT",
		func(c *Client, r *ResourceRef) (string, error) { return c.Stop(r) }))
	proxmoxCmd.AddCommand(newLifecycleCmd("shutdown", "Send a graceful shutdown signal",
		func(c *Client, r *ResourceRef) (string, error) { return c.Shutdown(r) }))
	proxmoxCmd.AddCommand(newLifecycleCmd("restart", "Reboot a VM or CT",
		func(c *Client, r *ResourceRef) (string, error) { return c.Reboot(r) }))
	proxmoxCmd.AddCommand(newDeleteCmd())

	// Phase 2c: creation + templates + cloud-init
	proxmoxCmd.AddCommand(newTemplateCmd())
	proxmoxCmd.AddCommand(newCreateCmd())
	proxmoxCmd.AddCommand(newCloudInitCmd())

	// Phase 3: snapshots + SSH
	proxmoxCmd.AddCommand(newSnapshotCmd())
	proxmoxCmd.AddCommand(newResolveIPCmd())
	proxmoxCmd.AddCommand(newSSHCmd())
	proxmoxCmd.AddCommand(newExecCmd())
	proxmoxCmd.AddCommand(newCopyCmd())

	rootCmd.SetVersionTemplate("ptx-proxmox version {{.Version}}\n")
}

// handleMetaArg processes dispatcher-introspection flags the main binary uses
// to enumerate helpers (--description, --list-commands, --help-ai). Returning
// true means the flag was handled and main should exit without invoking cobra.
func handleMetaArg(arg string) bool {
	switch arg {
	case "--description":
		fmt.Println("Portunix Proxmox VE Management Helper")
		return true
	case "--list-commands":
		fmt.Println("proxmox")
		return true
	case "--help-ai":
		showHelpAI()
		return true
	}
	return false
}

func main() {
	if len(os.Args) == 2 {
		if handleMetaArg(os.Args[1]) {
			return
		}
	}
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
