/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package main

import (
	"encoding/json"
	"fmt"
)

// showHelpAI emits a machine-readable command summary (per Issue #163).
func showHelpAI() {
	type CommandInfo struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	type AIHelp struct {
		Tool        string        `json:"tool"`
		Version     string        `json:"version"`
		Description string        `json:"description"`
		Commands    []CommandInfo `json:"commands"`
	}
	help := AIHelp{
		Tool:        "ptx-ssh",
		Version:     version,
		Description: "Unified SSH client and non-interactive password auth for Portunix",
		Commands: []CommandInfo{
			{Name: "connect", Description: "Open an interactive SSH session"},
			{Name: "exec", Description: "Run a single remote command"},
			{Name: "copy", Description: "Transfer files between local and remote via SFTP"},
			{Name: "rsync", Description: "Run rsync over an ssh tunnel"},
			{Name: "bootstrap-key", Description: "Push a public key to a password-only host"},
			{Name: "trust", Description: "Accept a remote host key explicitly"},
			{Name: "known-hosts list", Description: "List entries in the Portunix known_hosts store"},
			{Name: "known-hosts remove", Description: "Remove entries from the Portunix known_hosts store"},
			{Name: "agent start", Description: "Start an ssh-agent and print shell env"},
			{Name: "agent add", Description: "Add an identity to the ssh-agent"},
			{Name: "agent stop", Description: "Stop the active ssh-agent"},
			{Name: "sshpass-compat", Description: "Emit a one-shot password fd/fifo for ansible --ask-pass"},
		},
	}
	data, _ := json.MarshalIndent(help, "", "  ")
	fmt.Println(string(data))
}

func showHelpExpert() {
	fmt.Printf("PTX-SSH v%s - Unified SSH client and non-interactive password auth\n", version)
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println()
	fmt.Println("DESCRIPTION:")
	fmt.Println("  SSH client consolidated under 'portunix ssh'. Supports native Go crypto/ssh")
	fmt.Println("  (default) and a wrapper mode that delegates to the system ssh/scp/rsync for")
	fmt.Println("  tools that only accept passwords via a pty. Passwords are pulled from the")
	fmt.Println("  ptx-credential store or read non-interactively from FDs/files; plaintext on")
	fmt.Println("  the command line is explicitly rejected.")
	fmt.Println()
	fmt.Println("COMMANDS:")
	fmt.Println("  connect <user@host>                Interactive SSH session")
	fmt.Println("  exec <user@host> <cmd>             Run a single command remotely")
	fmt.Println("  copy <src> <dst>                   SFTP transfer (user@host:/path in either side)")
	fmt.Println("  rsync <src> <dst>                  Rsync over SSH (requires local rsync)")
	fmt.Println("  bootstrap-key <user@host>          Push ~/.ssh/id_ed25519.pub to the host")
	fmt.Println("  trust <user@host>                  Record host key without auth")
	fmt.Println("  known-hosts list|remove            Manage ~/.portunix/ssh/known_hosts")
	fmt.Println("  agent start|add|stop               ssh-agent lifecycle")
	fmt.Println("  sshpass-compat <cred-id>           Emit one-shot password fd/fifo")
	fmt.Println()
	fmt.Println("CREDENTIAL FLAGS (mutually exclusive):")
	fmt.Println("  -i, --identity <path>              Private key (preferred default)")
	fmt.Println("      --credential <id>              Fetch from ptx-credential store")
	fmt.Println("      --credential-fd <N>            Read from open file descriptor")
	fmt.Println("      --credential-file <path>       Read first line of 0400/0600 file")
	fmt.Println("      --credential-env               Read PTX_SSH_PASSWORD env var")
	fmt.Println("      --ask-user                     Prompt for username on TTY")
	fmt.Println("      --ask-pass                     Prompt for password on TTY (no echo)")
	fmt.Println("      --credential-store <name>      Optional ptx-credential store name")
	fmt.Println()
	fmt.Println("SESSION FLAGS:")
	fmt.Println("      --wrapper                      Exec system ssh/scp under pty")
	fmt.Println("      --host-key-check=strict|accept-new|off")
	fmt.Println("                                     Host-key policy (default: accept-new)")
	fmt.Println("      --use-system-known-hosts       Additionally consult ~/.ssh/known_hosts")
	fmt.Println()
	fmt.Println("ENVIRONMENT:")
	fmt.Println("  PTX_SSH_PASSWORD                   Password for --credential-env")
	fmt.Println("  PORTUNIX_CI=1                      Refuse host-key-check=off and interactive prompts")
	fmt.Println()
	fmt.Println("STORAGE:")
	fmt.Println("  ~/.portunix/ssh/known_hosts        Portunix-local known_hosts store")
	fmt.Println()
	fmt.Println("SAFETY:")
	fmt.Println("  * --password <value> on the CLI is rejected")
	fmt.Println("  * host-key-check=off is refused under PORTUNIX_CI=1")
	fmt.Println("  * passwords are wrapped in SecretString; never logged at any verbosity")
}
