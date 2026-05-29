/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package main

import (
	"fmt"
	"os"

	"portunix.ai/portunix/src/helpers/ptx-specpm/cmd"
)

var version = "dev"

// displayName is the user-facing brand. When the binary is invoked directly
// (e.g. './ptx-specpm --version') we keep the helper name. When it is routed
// by the Portunix dispatcher (the parent prepends 'specpm' as os.Args[1]),
// we present 'portunix specpm' so the user does not see the internal helper
// name.
var displayName = "ptx-specpm"

func main() {
	// Detect dispatcher routing BEFORE we strip the command name.
	if len(os.Args) > 1 && os.Args[1] == "specpm" {
		displayName = "portunix specpm"
	}

	// Dispatcher meta-flags handled before stripping the dispatched command.
	for _, arg := range os.Args[1:] {
		if arg == "--help-ai" || arg == "--help-expert" {
			if dispatcherMetaFlags([]string{arg}) {
				return
			}
		}
	}

	// When invoked via the dispatcher as "portunix specpm ...", strip the
	// routed command name so Cobra sees the helper's own subcommand tree.
	if len(os.Args) > 1 && os.Args[1] == "specpm" {
		os.Args = append(os.Args[:1], os.Args[2:]...)
	}

	if len(os.Args) > 1 && dispatcherMetaFlags(os.Args[1:]) {
		return
	}

	cmd.SetVersion(version)
	cmd.SetDisplayName(displayName)
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, displayName+":", err)
		os.Exit(1)
	}
}

// dispatcherMetaFlags handles the standard helper meta-flags before Cobra
// gets a chance to interpret them. Mirrors the convention in ptx-ssh and
// other helpers (see Issue #163).
func dispatcherMetaFlags(args []string) bool {
	if len(args) == 0 {
		return false
	}
	switch args[0] {
	case "--description":
		fmt.Println("Initialize project-management specifications using spec-kit-pm")
		return true
	case "--list-commands":
		fmt.Println("specpm")
		return true
	case "--version":
		fmt.Printf("%s version %s\n", displayName, version)
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
