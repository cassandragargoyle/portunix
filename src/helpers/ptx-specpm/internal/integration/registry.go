/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package integration

// All returns the drivers known to this build. Phase 1: only Claude Code
// ships end-to-end. Future drivers (Phase 4: Copilot, Cursor, Gemini CLI,
// Codex CLI) plug in here once their files land in this package.
func All() []Driver {
	return []Driver{
		&claudeDriver{},
	}
}

// Lookup returns the driver matching name, or nil + false when unknown.
func Lookup(name string) (Driver, bool) {
	for _, d := range All() {
		if d.Name() == name {
			return d, true
		}
	}
	return nil, false
}
