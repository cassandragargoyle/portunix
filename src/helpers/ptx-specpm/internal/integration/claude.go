/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package integration

import (
	"embed"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// claudeAssets holds the slash-command and skill templates that are part
// of the helper itself (integration glue authored here, not upstream kit
// content). Embedding this small bundle is consistent with ADR-040 D3:
// the rejection there was of embedding the *kit*, not of embedding the
// helper's own assets.
//
//go:embed assets/claude/*.md
var claudeAssets embed.FS

// claudeVerbs are the verbs emitted by the Phase 1 Claude integration.
var claudeVerbs = []string{
	"init", "spec", "plan", "risks", "decisions", "review", "report", "check",
}

type claudeDriver struct{}

func (c *claudeDriver) Name() string { return "claude" }

func (c *claudeDriver) Description() string {
	return "Claude Code (Anthropic) — slash-commands or skills bundle"
}

func (c *claudeDriver) Detect() (bool, string) {
	if _, err := exec.LookPath("claude"); err != nil {
		return false, ""
	}
	out, err := exec.Command("claude", "--version").CombinedOutput()
	if err != nil {
		return true, "" // installed but version unreadable
	}
	v := strings.TrimSpace(string(out))
	if v == "" {
		return true, ""
	}
	return true, v
}

func (c *claudeDriver) Render(opts RenderOptions) (*RenderResult, error) {
	if opts.TargetDir == "" {
		return nil, fmt.Errorf("claude driver: empty target dir")
	}

	res := &RenderResult{AgentName: "claude"}

	if opts.Skills {
		res.Mode = "skills"
		for _, verb := range claudeVerbs {
			data, err := claudeAssets.ReadFile("assets/claude/specpm." + verb + ".md")
			if err != nil {
				return nil, fmt.Errorf("read embedded claude template %q: %w", verb, err)
			}
			skillDir := filepath.Join(opts.TargetDir, ".claude", "skills", "specpm-"+verb)
			if err := os.MkdirAll(skillDir, 0o755); err != nil {
				return nil, err
			}
			skillFile := filepath.Join(skillDir, "SKILL.md")
			if err := os.WriteFile(skillFile, data, 0o644); err != nil {
				return nil, err
			}
			res.WrittenPaths = append(res.WrittenPaths, skillFile)
		}
	} else {
		res.Mode = "slash-commands"
		cmdDir := filepath.Join(opts.TargetDir, ".claude", "commands")
		if err := os.MkdirAll(cmdDir, 0o755); err != nil {
			return nil, err
		}
		for _, verb := range claudeVerbs {
			data, err := claudeAssets.ReadFile("assets/claude/specpm." + verb + ".md")
			if err != nil {
				return nil, fmt.Errorf("read embedded claude template %q: %w", verb, err)
			}
			out := filepath.Join(cmdDir, "specpm."+verb+".md")
			if err := os.WriteFile(out, data, 0o644); err != nil {
				return nil, err
			}
			res.WrittenPaths = append(res.WrittenPaths, out)
		}
	}

	if opts.GovernanceIncludeFn != nil {
		path, changed, err := opts.GovernanceIncludeFn(opts.TargetDir)
		if err != nil {
			return nil, fmt.Errorf("augment CLAUDE.md: %w", err)
		}
		if changed {
			res.WrittenPaths = append(res.WrittenPaths, path)
		}
	}

	return res, nil
}
