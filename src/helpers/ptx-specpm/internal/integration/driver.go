/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
// Package integration owns the per-AI-agent integration drivers. Phase 1
// ships only the Claude Code driver; the Driver interface is stable so
// Copilot, Cursor, Gemini CLI, etc. can plug in during Phase 4 without
// touching the core helper code.
package integration

// Driver is implemented by each agent integration target.
type Driver interface {
	Name() string
	Description() string
	// Detect returns whether the agent CLI is installed and (if available)
	// its version string.
	Detect() (installed bool, version string)
	// Render writes the integration's bundle (slash-commands or skills) into
	// the project at opts.TargetDir.
	Render(opts RenderOptions) (*RenderResult, error)
}

// RenderOptions is the input contract for Driver.Render.
type RenderOptions struct {
	TargetDir        string
	Skills           bool
	IgnoreAgentTools bool
	// GovernanceIncludeFn is an injected hook that augments the host CLAUDE.md
	// (or equivalent) with an include line pointing at the governance memory
	// file. Drivers that don't support such an augmentation simply ignore it.
	GovernanceIncludeFn func(targetDir string) (path string, changed bool, err error)
}

// RenderResult reports what the driver wrote so the init summary can show it.
type RenderResult struct {
	AgentName    string
	Mode         string // "slash-commands" or "skills"
	WrittenPaths []string
}
