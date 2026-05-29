/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package main

import (
	"encoding/json"
	"fmt"

	"portunix.ai/portunix/src/helpers/ptx-specpm/internal/kit"
)

func showHelpAI() {
	type CommandInfo struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	type AIHelp struct {
		Tool        string        `json:"tool"`
		Version     string        `json:"version"`
		Description string        `json:"description"`
		KitRef      string        `json:"kit_default_ref"`
		Commands    []CommandInfo `json:"commands"`
	}
	help := AIHelp{
		Tool:        displayName,
		Version:     version,
		Description: "Initialize project-management specifications using spec-kit-pm",
		KitRef:      kit.DefaultPinnedRef,
		Commands: []CommandInfo{
			{Name: "init", Description: "Bootstrap a project with the spec-kit-pm scaffold"},
			{Name: "upgrade", Description: "Refresh .specpm/ kit cache; --ref bumps pin, --to-default tracks Portunix release"},
			{Name: "check", Description: "Inspect detected agents, active profile, and bundled kit version"},
			{Name: "version", Description: "Print helper and bundled kit version"},
			{Name: "integration list", Description: "List supported AI-agent integration targets"},
		},
	}
	data, _ := json.MarshalIndent(help, "", "  ")
	fmt.Println(string(data))
}

func showHelpExpert() {
	fmt.Printf("%s v%s - Project-management specifications via spec-kit-pm\n", displayName, version)
	fmt.Println("================================================================")
	fmt.Println()
	fmt.Println("DESCRIPTION:")
	fmt.Println("  Bootstraps the spec-kit-pm scaffold inside a project so AI agents")
	fmt.Println("  (Claude Code in Phase 1) can drive Charter, plan, risks, decisions,")
	fmt.Println("  KPIs, and reports through /specpm.* slash-commands or skills.")
	fmt.Println()
	fmt.Println("  The kit lives upstream at CassandraGargoyle/spec-kit-pm. The helper")
	fmt.Println("  contains only logic; kit content is fetched at runtime via shallow")
	fmt.Println("  clone and cached at ~/.cache/portunix/specpm/<ref>/.")
	fmt.Println()
	fmt.Println("COMMANDS:")
	fmt.Println("  init <name|.|--here>             Bootstrap project scaffold")
	fmt.Println("  upgrade                          Refresh .specpm/ kit cache (Phase 2)")
	fmt.Println("  check                            Show installed agents, profile, kit ref")
	fmt.Println("  version                          Print helper version + kit ref")
	fmt.Println("  integration list                 List supported AI-agent targets")
	fmt.Println()
	fmt.Println("INIT FLAGS:")
	fmt.Println("  --here                           Initialise in current directory")
	fmt.Println("  --force                          Overwrite scaffold files (NEVER artifacts)")
	fmt.Println("  --integration <agent>            Agent target (Phase 1: claude)")
	fmt.Println("  --integration-options=\"--skills\" Emit Claude Code skills instead of slash-commands")
	fmt.Println("  --profile <name>                 Apply a profile from .specpm/profiles/")
	fmt.Println("  --source <git|path>              Kit source (default: git)")
	fmt.Println("  --ref <tag-or-sha>               Override pinned ref (with --source git)")
	fmt.Println("  --ignore-agent-tools             Skip detection of installed agent CLIs")
	fmt.Println()
	fmt.Println("UPGRADE FLAGS:")
	fmt.Println("  --ref <X>                        Bump project pin to git ref X")
	fmt.Println("  --to-default                     Shortcut for --ref <DefaultPinnedRef>")
	fmt.Println("  --source <path|git>              Switch provider (incompatible with --ref+<path>)")
	fmt.Println()
	fmt.Println("LAYOUT:")
	fmt.Println("  .specpm/                         Kit cache (templates, agents, workflows)")
	fmt.Println("    kit.json / extensions/ / presets/   Carve-outs preserved across upgrade")
	fmt.Println("  specs/project/<id>/              User artifacts (Charter, plan, risks, ...) — sacred")
	fmt.Println("  ~/.cache/portunix/specpm/<ref>/  Runtime cache (offline reuse)")
	fmt.Println()
	fmt.Println("REQUIREMENTS:")
	fmt.Println("  --source git: local 'git' binary on PATH + network on first run")
	fmt.Println("  Air-gapped: --source <path> against a vendored fork")
	fmt.Println()
	fmt.Println("REFERENCES:")
	fmt.Println("  ADR-040 (Active)   Phase 1 Architecture Decision Record")
	fmt.Println("  ADR-041 (Active)   Phase 2 — upgrade UX, drift policy, air-gapped workflow")
	fmt.Println("  Issue #183         Phase 1 implementation tracker")
	fmt.Println("  Issue #184         Phase 2 implementation tracker")
}
