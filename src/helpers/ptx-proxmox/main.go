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
	"strings"

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

// showHelpExpert prints detailed help with flags, storage and examples
// (per Issue #244)
func showHelpExpert() {
	fmt.Printf("PTX-PROXMOX v%s - Proxmox VE management helper\n", strings.TrimPrefix(version, "v"))
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println()
	fmt.Println("DESCRIPTION:")
	fmt.Println("  Remote Proxmox VE management over the REST API: connection profiles,")
	fmt.Println("  VM/CT lifecycle, creation from templates with cloud-init, snapshots,")
	fmt.Println("  and SSH/SCP with guest IP resolution via the QEMU guest agent or LXC")
	fmt.Println("  interfaces. Invoked as 'portunix proxmox ...'.")
	fmt.Println()
	fmt.Println("COMMANDS:")
	fmt.Println("  auth login|status|logout|list      Manage connection profiles")
	fmt.Println("  list                               List VMs and containers")
	fmt.Println("  info|status <name|vmid>            Show configuration / runtime status")
	fmt.Println("  start|stop|shutdown|restart <id>   Lifecycle operations")
	fmt.Println("  delete <name|vmid>                 Destroy a VM or CT")
	fmt.Println("  template list                      List templates and ISOs")
	fmt.Println("  create vm|ct                       Create a VM or CT from a template")
	fmt.Println("  cloud-init <name|vmid>             Configure cloud-init on a VM")
	fmt.Println("  snapshot create|list|revert|delete Manage snapshots")
	fmt.Println("  resolve-ip <name|vmid>             Print the guest IP address")
	fmt.Println("  ssh <name|vmid> [-- ssh-args]      Interactive SSH to the guest")
	fmt.Println("  exec <name|vmid> -- <cmd>          Run a command over SSH")
	fmt.Println("  copy <src> <dst>                   SCP to/from a guest (<name>:<path>)")
	fmt.Println()
	fmt.Println("AUTH LOGIN FLAGS:")
	fmt.Println("      --host <host>                  Proxmox VE hostname or IP (required)")
	fmt.Println("      --port <n>                     API port (default 8006)")
	fmt.Println("      --token-id <user@realm!name>   API token ID (with --token-secret)")
	fmt.Println("      --user <user@realm>            Username for password auth")
	fmt.Println("      --profile <name>               Profile name (default 'default' or host)")
	fmt.Println("      --insecure                     Skip TLS verification (self-signed certs)")
	fmt.Println()
	fmt.Println("COMMON FLAGS:")
	fmt.Println("      --profile <name>               Profile to use (defaults to current)")
	fmt.Println("      --wait / --timeout <s>         Wait for the Proxmox task to finish")
	fmt.Println("      --format table|json            Output format (list)")
	fmt.Println()
	fmt.Println("CREATE FLAGS:")
	fmt.Println("      --name, --node, --template     Required")
	fmt.Println("      --vmid <n>                     VMID (0 = next free)")
	fmt.Println("      --cores, --memory, --disk      Resources (1 core, 1024 MiB, 8 GiB)")
	fmt.Println("      --storage <pool>               Storage pool (default local-lvm)")
	fmt.Println("      --ssh-key <path> | --password  Guest access (CT needs one of them)")
	fmt.Println()
	fmt.Println("SSH FLAGS:")
	fmt.Println("  -u, --user <name>                  Guest user (default root)")
	fmt.Println("  -i, --identity <path>              SSH identity file")
	fmt.Println("  -p, --port <n>                     SSH port (default 22)")
	fmt.Println("  -o, --option KEY=VALUE             Extra ssh -o options (repeatable)")
	fmt.Println("      --ipv6                         Prefer IPv6 when resolving the guest")
	fmt.Println()
	fmt.Println("ENVIRONMENT:")
	fmt.Println("  PORTUNIX_PROXMOX_CONFIG            Override path to the profile file")
	fmt.Println()
	fmt.Println("STORAGE:")
	fmt.Println("  ~/.config/portunix/proxmox.json    Connection profiles (mode 0600)")
	fmt.Println()
	fmt.Println("EXAMPLES:")
	fmt.Println("  portunix proxmox auth login --host pve.example.com --token-id 'root@pam!ptx' --token-secret ...")
	fmt.Println("  portunix proxmox list --type vm --format json")
	fmt.Println("  portunix proxmox create ct --name web --node pve --template debian-12 --ssh-key ~/.ssh/id_ed25519.pub")
	fmt.Println("  portunix proxmox snapshot create web before-upgrade")
	fmt.Println("  portunix proxmox exec web -- uname -a")
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
// to enumerate helpers (--description, --list-commands, --help-ai,
// --help-expert). Returning true means the flag was handled and main should
// exit without invoking cobra.
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
	case "--help-expert":
		showHelpExpert()
		return true
	}
	return false
}

// handleMetaArgs handles the meta flags both for a direct call
// (ptx-<helper> --help-ai) and for a dispatcher call, where the top-level
// command name comes first (ptx-<helper> <command> --help-ai)
func handleMetaArgs(args []string) bool {
	switch len(args) {
	case 1:
		return handleMetaArg(args[0])
	case 2:
		if isTopLevelCommand(args[0]) {
			return handleMetaArg(args[1])
		}
	}
	return false
}

// isTopLevelCommand reports whether name is a direct subcommand of rootCmd
// or one of its aliases
func isTopLevelCommand(name string) bool {
	for _, c := range rootCmd.Commands() {
		if c.Name() == name || c.HasAlias(name) {
			return true
		}
	}
	return false
}

func main() {
	if handleMetaArgs(os.Args[1:]) {
		return
	}
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
