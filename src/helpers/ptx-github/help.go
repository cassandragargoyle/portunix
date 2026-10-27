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
		Tool:        "ptx-github",
		Version:     version,
		Description: "GitHub integration: pure-Go clone, releases and asset downloads, encrypted token store",
		Commands: []CommandInfo{
			{Name: "github clone", Description: "Clone a GitHub repository (pure Go, no external git binary)"},
			{Name: "github checkout", Description: "Checkout a branch, tag, or commit in a local clone"},
			{Name: "github tags", Description: "List tags from a local repo or remote URL"},
			{Name: "github releases", Description: "List releases for a GitHub repository"},
			{Name: "github download", Description: "Download a release asset"},
			{Name: "github info", Description: "Show metadata for a GitHub repository"},
			{Name: "github status", Description: "Show status of the local Git repository"},
			{Name: "github auth login", Description: "Store a GitHub personal access token"},
			{Name: "github auth logout", Description: "Remove a stored GitHub account"},
			{Name: "github auth status", Description: "Show current authentication status"},
			{Name: "github auth list", Description: "List stored GitHub accounts"},
			{Name: "github auth set-default", Description: "Change the default account"},
		},
	}
	data, _ := json.MarshalIndent(help, "", "  ")
	fmt.Println(string(data))
}

// showHelpExpert prints detailed help with flags, authentication and examples
func showHelpExpert() {
	fmt.Printf("PTX-GITHUB v%s - GitHub integration helper\n", strings.TrimPrefix(version, "v"))
	fmt.Println("═══════════════════════════════════════════════════════════")
	fmt.Println()
	fmt.Println("DESCRIPTION:")
	fmt.Println("  GitHub repositories, releases and authentication for Portunix. Clones")
	fmt.Println("  with go-git (no external git binary), downloads release assets with")
	fmt.Println("  resume and SHA-256 verification, and keeps tokens in an AES-256-GCM")
	fmt.Println("  encrypted store. Invoked as 'portunix github ...'.")
	fmt.Println()
	fmt.Println("COMMANDS:")
	fmt.Println("  clone <owner/repo|url> [path]      Clone a repository")
	fmt.Println("  checkout <ref>                     Checkout a branch, tag, or commit")
	fmt.Println("  tags [owner/repo|url]              List tags (local repo or remote)")
	fmt.Println("  releases <owner/repo|url>          List releases")
	fmt.Println("  download <owner/repo|url> [tag]    Download release assets")
	fmt.Println("  info <owner/repo|url>              Show repository metadata")
	fmt.Println("  status                             Show status of the local repository")
	fmt.Println("  auth login|logout|status|list|set-default")
	fmt.Println("                                     Manage stored accounts")
	fmt.Println()
	fmt.Println("GLOBAL FLAGS:")
	fmt.Println("      --account <name>               Named account from the encrypted store")
	fmt.Println("      --json                         Machine-readable JSON output")
	fmt.Println("      --password                     Prompt for the token store password")
	fmt.Println("      --store <path>                 Override path to the token store")
	fmt.Println()
	fmt.Println("CLONE FLAGS:")
	fmt.Println("  -b, --branch <ref>                 Branch or tag to checkout")
	fmt.Println("      --depth <n>                    Shallow clone with n commits")
	fmt.Println("      --bare | --no-checkout         Bare repository / skip checkout")
	fmt.Println("      --recurse-submodules           Clone submodules recursively")
	fmt.Println()
	fmt.Println("DOWNLOAD FLAGS:")
	fmt.Println("      --asset <name> | --all         Exact asset name / all assets")
	fmt.Println("      --platform <os-arch|auto>      Select asset by platform")
	fmt.Println("  -o, --output <path>                Output file or directory")
	fmt.Println("      --resume | --overwrite         Resume partial file / replace existing")
	fmt.Println("      --sha256 <hex>                 Verify the downloaded file")
	fmt.Println()
	fmt.Println("AUTH LOGIN FLAGS:")
	fmt.Println("      --token <value|->              Token value ('-' reads from stdin)")
	fmt.Println("      --name <account>               Account name (default \"default\")")
	fmt.Println("      --default                      Make this account the default")
	fmt.Println("      --skip-verify                  Do not validate the token against GitHub")
	fmt.Println()
	fmt.Println("AUTHENTICATION ORDER:")
	fmt.Println("  --account, $PORTUNIX_GITHUB_TOKEN, $GITHUB_TOKEN, $GH_TOKEN,")
	fmt.Println("  then the default account in the encrypted store")
	fmt.Println()
	fmt.Println("STORAGE:")
	fmt.Println("  ~/.portunix/github/auth.json       Encrypted token store (AES-256-GCM)")
	fmt.Println()
	fmt.Println("EXAMPLES:")
	fmt.Println("  portunix github clone cassandragargoyle/portunix --depth 1")
	fmt.Println("  portunix github releases cassandragargoyle/portunix --limit 5 --assets")
	fmt.Println("  portunix github download cassandragargoyle/portunix --platform auto -o ./dl")
	fmt.Println("  echo \"$TOKEN\" | portunix github auth login --token - --default")
}
