/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package drivers

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// runCmd runs a command, streaming stdout/stderr to the user, and returns
// any error so callers can surface it with context.
func runCmd(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// runContainer routes container lifecycle through `portunix container`.
// Helper drivers MUST NOT call docker/podman directly — they delegate to
// ptx-container so the same UX works regardless of which runtime is installed.
func runContainer(verb, name string) error {
	return runCmd("portunix", "container", verb, name)
}

// containerRunning checks whether a container with the given name is running.
// Uses `portunix container inspect` (the supported subcommand — `status` does
// not exist) and looks for `"Running": true` in its JSON output.
// Returns (false, nil) if the container is missing or stopped.
func containerRunning(name string) (bool, error) {
	out, err := exec.Command("portunix", "container", "inspect", name, "--format", "json").Output()
	if err != nil {
		// Treat exit-code errors from `container inspect` as "not running" rather
		// than fatal so Status() can report a clean stopped state.
		if _, ok := err.(*exec.ExitError); ok {
			return false, nil
		}
		// If portunix itself is missing we surface that — drivers without the
		// dispatcher in PATH cannot operate in container mode anyway.
		return false, err
	}
	// `container inspect` output is the engine's native inspect JSON (Docker /
	// Podman both expose `.State.Running`). We don't fully parse the structure
	// because `Running` is uniquely named and reliable as a substring match.
	return strings.Contains(string(out), `"Running": true`) ||
		strings.Contains(string(out), `"Running":true`), nil
}

// containerName returns the conventional container name for an instance
// derived from the engine slug. Falls back to "portunix-<engine>" when no
// explicit instance is given.
func containerName(instance, engineSlug string) string {
	if instance != "" {
		return instance
	}
	return "portunix-" + engineSlug
}

// resolveBackupDir builds (and creates) the backup directory used by drivers.
// Order of precedence: explicit destination > $PORTUNIX_HOME/database/<engine>
// > $HOME/.portunix/database/backups/<engine>/<instance>.
func resolveBackupDir(destination, engine, instance string) (string, error) {
	if destination != "" {
		if err := os.MkdirAll(destination, 0o700); err != nil {
			return "", err
		}
		return destination, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home dir: %w", err)
	}
	if instance == "" {
		instance = "default"
	}
	dir := filepath.Join(home, ".portunix", "database", "backups", engine, instance)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	return dir, nil
}

// shellQuote wraps a value in single quotes for safe sh -c interpolation.
// We escape any embedded single quotes by closing/opening the string —
// `'` becomes `'\”`. Used only with values constructed by us, never raw
// user input from the network.
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// nonEmptyLines splits text on newlines and drops empty lines.
func nonEmptyLines(s string) []string {
	var out []string
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimRight(line, "\r")
		if strings.TrimSpace(line) != "" {
			out = append(out, line)
		}
	}
	return out
}

// firstNonEmptyLine returns the first non-blank line of s, trimmed.
func firstNonEmptyLine(s string) string {
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			return line
		}
	}
	return ""
}

// splitPipe splits a `psql -At` style row on "|" and trims each field.
// PostgreSQL's `-At` output uses pipe as the field separator.
func splitPipe(s string) []string {
	parts := strings.Split(s, "|")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}

// safeIndex returns parts[i] or "" if out of range.
func safeIndex(parts []string, i int) string {
	if i < 0 || i >= len(parts) {
		return ""
	}
	return parts[i]
}

// jsonifyColumns converts pipe-separated `schema.table|column|type` rows
// into a compact JSON array. Stays in this package so drivers don't need
// encoding/json for trivial output shapes.
func jsonifyColumns(s string) string {
	var b strings.Builder
	b.WriteString("[")
	first := true
	for _, line := range nonEmptyLines(s) {
		parts := splitPipe(line)
		if len(parts) < 3 {
			continue
		}
		if !first {
			b.WriteString(",")
		}
		first = false
		fmt.Fprintf(&b, "{\"table\":%q,\"column\":%q,\"type\":%q}", parts[0], parts[1], parts[2])
	}
	b.WriteString("]")
	return b.String()
}
