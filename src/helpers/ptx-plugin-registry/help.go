/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

// handleMetaArg processes the --help-ai / --help-expert flags before Cobra
// parses the arguments. Returning true means the flag was handled.
func handleMetaArg(arg string) bool {
	switch arg {
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

// showHelpAI emits a machine-readable command summary (per Issues #163, #244)
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
		Tool:        "ptx-plugin-registry",
		Version:     version,
		Description: "Local gRPC registry for plugin platform-capability queries",
		Commands: []CommandInfo{
			{Name: "plugin-registry serve", Description: "Start the plugin-registry gRPC server"},
		},
	}
	data, _ := json.MarshalIndent(help, "", "  ")
	fmt.Println(string(data))
}

// showHelpExpert prints detailed help with transports, flags and examples
func showHelpExpert() {
	fmt.Printf("PTX-PLUGIN-REGISTRY v%s - Plugin registry gRPC server\n", strings.TrimPrefix(version, "v"))
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println()
	fmt.Println("DESCRIPTION:")
	fmt.Println("  Serves the PluginRegistryService gRPC API so hosting platforms can")
	fmt.Println("  discover which installed Portunix plugins target them and fetch each")
	fmt.Println("  plugin's platform-specific registration payload. The caller starts the")
	fmt.Println("  server for the duration of a query and stops it (no system service).")
	fmt.Println("  Invoked as 'portunix plugin-registry ...'.")
	fmt.Println()
	fmt.Println("COMMANDS:")
	fmt.Println("  serve                              Start the gRPC server (until SIGINT/SIGTERM)")
	fmt.Println()
	fmt.Println("SERVE FLAGS:")
	fmt.Println("  -m, --mode unix|tcp                Transport (default: unix on Linux/macOS,")
	fmt.Println("                                     tcp on Windows)")
	fmt.Println("  -s, --socket <path>                Unix socket path (mode=unix)")
	fmt.Println("  -p, --port <n>                     TCP port (mode=tcp, default 9500)")
	fmt.Println("      --bind <addr>                  TCP bind address (default 127.0.0.1)")
	fmt.Println("      --plugin-dir <path>            Override the plugin install directory")
	fmt.Println()
	fmt.Println("ENVIRONMENT:")
	fmt.Println("  XDG_RUNTIME_DIR                    Base directory for the default socket")
	fmt.Println()
	fmt.Println("STORAGE:")
	fmt.Println("  $XDG_RUNTIME_DIR/portunix/plugin-registry.sock")
	fmt.Println("                                     Default socket (fallback ~/.portunix/run/)")
	fmt.Println("  ~/.portunix/plugins/registry.json  Plugin registry read by the server")
	fmt.Println()
	fmt.Println("SAFETY:")
	fmt.Println("  * unix socket is created with mode 0600")
	fmt.Println("  * TCP listens on loopback only unless --bind is changed")
	fmt.Println()
	fmt.Println("EXAMPLES:")
	fmt.Println("  portunix plugin-registry serve")
	fmt.Println("  portunix plugin-registry serve --mode unix --socket /tmp/ptx-reg.sock")
	fmt.Println("  portunix plugin-registry serve --mode tcp --port 9500")
}
