/*
 *  SPDX-License-Identifier: MIT
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package skills

import (
	"fmt"
	"sort"
	"strings"
)

// sharedProjectDir is the project-level directory read by every agent
// following the shared .agents convention
const sharedProjectDir = ".agents/skills"

// universalTarget is the shared .agents/skills directory, always written
// unless --no-universal is given
var universalTarget AgentTarget = &dirTarget{
	id:          UniversalID,
	displayName: "Universal (.agents/skills)",
	projectDir:  sharedProjectDir,
	userDir:     userPath{rootHome, ".agents/skills"},
}

// agentRegistry holds the per-agent drivers. Adding an agent is an additive
// entry here (ADR-043 D3). Directories and detection follow the agent table
// of the `skills` CLI (skills@1.5.9); most agents read the shared
// .agents/skills in a project and differ only in their user-scope directory.
var agentRegistry = []AgentTarget{
	&dirTarget{
		id:          "amp",
		displayName: "Amp",
		projectDir:  sharedProjectDir,
		userDir:     userPath{rootConfig, "agents/skills"},
		detectPaths: []userPath{{rootConfig, "amp"}},
	},
	&dirTarget{
		id:            "claude",
		displayName:   "Claude Code",
		projectDir:    ".claude/skills",
		userDir:       userPath{rootHome, ".claude/skills"},
		userEnv:       "CLAUDE_CONFIG_DIR",
		detectPackage: "claude-code",
		detectPaths:   []userPath{{rootHome, ".claude"}},
	},
	&dirTarget{
		id:          "copilot",
		displayName: "GitHub Copilot",
		projectDir:  sharedProjectDir,
		userDir:     userPath{rootHome, ".copilot/skills"},
		detectPaths: []userPath{{rootHome, ".copilot"}},
	},
	&dirTarget{
		id:          "cursor",
		displayName: "Cursor",
		projectDir:  sharedProjectDir,
		userDir:     userPath{rootHome, ".cursor/skills"},
		detectPaths: []userPath{{rootHome, ".cursor"}},
	},
	&dirTarget{
		id:            "gemini",
		displayName:   "Gemini CLI",
		projectDir:    sharedProjectDir,
		userDir:       userPath{rootHome, ".gemini/skills"},
		detectPackage: "gemini-cli",
		detectPaths:   []userPath{{rootHome, ".gemini"}},
	},
	&dirTarget{
		id:          "opencode",
		displayName: "OpenCode",
		projectDir:  sharedProjectDir,
		userDir:     userPath{rootConfig, "opencode/skills"},
		detectPaths: []userPath{{rootConfig, "opencode"}},
	},
	&dirTarget{
		id:          "zed",
		displayName: "Zed",
		projectDir:  sharedProjectDir,
		userDir:     userPath{rootHome, ".agents/skills"},
		detectPaths: []userPath{{rootConfig, "zed"}},
	},
}

// Universal returns the universal target
func Universal() AgentTarget { return universalTarget }

// Agents returns the registered per-agent drivers sorted by ID
func Agents() []AgentTarget {
	out := append([]AgentTarget(nil), agentRegistry...)
	sort.Slice(out, func(i, j int) bool { return out[i].ID() < out[j].ID() })
	return out
}

// AllTargets returns the universal target followed by every agent driver
func AllTargets() []AgentTarget {
	return append([]AgentTarget{universalTarget}, Agents()...)
}

// agentAliases maps the agent names used by the `skills` CLI to driver IDs
var agentAliases = map[string]string{
	"claude-code":    "claude",
	"gemini-cli":     "gemini",
	"github-copilot": "copilot",
}

// LookupAgent returns the agent driver with the given ID or alias
func LookupAgent(id string) (AgentTarget, error) {
	if canonical, ok := agentAliases[id]; ok {
		id = canonical
	}
	for _, a := range agentRegistry {
		if a.ID() == id {
			return a, nil
		}
	}
	ids := make([]string, 0, len(agentRegistry))
	for _, a := range Agents() {
		ids = append(ids, a.ID())
	}
	return nil, fmt.Errorf("unknown agent %q (supported: %s)", id, strings.Join(ids, ", "))
}

// ResolveTargets builds the target list for an install: the universal target
// (unless disabled) plus all agents, the given agents, or the detected ones.
// A non-nil empty agentIDs means "no agents" (e.g. an interactive selection
// of none) and disables detection.
func ResolveTargets(agentIDs []string, allAgents, noUniversal bool) ([]AgentTarget, error) {
	var targets []AgentTarget
	if !noUniversal {
		targets = append(targets, universalTarget)
	}
	switch {
	case allAgents:
		targets = append(targets, Agents()...)
	case agentIDs != nil:
		seen := map[string]bool{}
		for _, id := range agentIDs {
			a, err := LookupAgent(id)
			if err != nil {
				return nil, err
			}
			if seen[a.ID()] {
				continue
			}
			seen[a.ID()] = true
			targets = append(targets, a)
		}
	default:
		for _, a := range Agents() {
			if ok, _ := a.Detect(); ok {
				targets = append(targets, a)
			}
		}
	}
	if len(targets) == 0 {
		return nil, fmt.Errorf("no install target selected (--no-universal requires --agent or --all-agents)")
	}
	return targets, nil
}
