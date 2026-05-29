/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package scaffold

import (
	"bytes"
	"os"
	"path/filepath"
)

// AugmentClaudeMD ensures CLAUDE.md at the project root contains a single
// include line referencing .specpm/memory/governance.md. Idempotent:
//   - file missing → created with the include line
//   - include line already present → file untouched
//   - file present without the line → line appended after a blank-line separator
//
// User content elsewhere in CLAUDE.md is never modified.
func AugmentClaudeMD(targetDir string) (path string, changed bool, err error) {
	path = filepath.Join(targetDir, "CLAUDE.md")
	const includeLine = "@.specpm/memory/governance.md"

	body, readErr := os.ReadFile(path)
	if os.IsNotExist(readErr) {
		header := []byte("# Project Instructions\n\n" + includeLine + "\n")
		if werr := os.WriteFile(path, header, 0o644); werr != nil {
			return path, false, werr
		}
		return path, true, nil
	}
	if readErr != nil {
		return path, false, readErr
	}
	if bytes.Contains(body, []byte(includeLine)) {
		return path, false, nil
	}
	if len(body) > 0 && body[len(body)-1] != '\n' {
		body = append(body, '\n')
	}
	body = append(body, '\n')
	body = append(body, []byte(includeLine+"\n")...)
	if werr := os.WriteFile(path, body, 0o644); werr != nil {
		return path, false, werr
	}
	return path, true, nil
}
