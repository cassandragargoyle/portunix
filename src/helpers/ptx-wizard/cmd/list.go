/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List available wizards",
	Long:  "List all built-in, user, and plugin-provided wizards discoverable in the search paths.",
	Run:   listWizards,
}

// listWizards walks all configured search paths and prints discovered wizards.
func listWizards(_ *cobra.Command, _ []string) {
	type wizardEntry struct {
		Name        string
		Description string
		Path        string
		Source      string
	}

	var entries []wizardEntry
	seen := make(map[string]bool)

	for _, sp := range searchPaths() {
		matches, err := filepath.Glob(filepath.Join(sp.dir, "*.yaml"))
		if err != nil {
			continue
		}
		for _, m := range matches {
			name := strings.TrimSuffix(filepath.Base(m), ".yaml")
			if seen[name] {
				continue
			}
			seen[name] = true

			desc := readWizardDescription(m)
			entries = append(entries, wizardEntry{
				Name:        name,
				Description: desc,
				Path:        m,
				Source:      sp.label,
			})
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		if entries[i].Source != entries[j].Source {
			return entries[i].Source < entries[j].Source
		}
		return entries[i].Name < entries[j].Name
	})

	if len(entries) == 0 {
		fmt.Println("No wizards found.")
		fmt.Println()
		printSearchPaths()
		return
	}

	fmt.Println("Available Wizards:")
	fmt.Println("==================")
	currentSource := ""
	for _, e := range entries {
		if e.Source != currentSource {
			fmt.Printf("\n[%s]\n", e.Source)
			currentSource = e.Source
		}
		desc := e.Description
		if desc == "" {
			desc = "(no description)"
		}
		fmt.Printf("  %-25s %s\n", e.Name, desc)
	}

	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  portunix wizard run <wizard-name>")
	fmt.Println("  portunix wizard run <path-to-yaml-file>")
}

// searchPath bundles a directory and a human-readable label.
type searchPath struct {
	dir   string
	label string
}

// searchPaths returns the ordered list of wizard search paths.
//
// Order:
//  1. ./examples/wizards/ (project bundle)
//  2. ~/.portunix/wizards/ (user)
//  3. $PORTUNIX_WIZARD_PATH split by OS path separator (plugins / extras)
func searchPaths() []searchPath {
	var paths []searchPath

	if abs, err := filepath.Abs("examples/wizards"); err == nil {
		paths = append(paths, searchPath{dir: abs, label: "Built-in"})
	}

	if home, err := os.UserHomeDir(); err == nil {
		paths = append(paths, searchPath{
			dir:   filepath.Join(home, ".portunix", "wizards"),
			label: "User",
		})
	}

	if extra := os.Getenv("PORTUNIX_WIZARD_PATH"); extra != "" {
		sep := string(os.PathListSeparator)
		for _, dir := range strings.Split(extra, sep) {
			dir = strings.TrimSpace(dir)
			if dir == "" {
				continue
			}
			paths = append(paths, searchPath{dir: dir, label: "Plugin"})
		}
	}

	return paths
}

// printSearchPaths prints search paths used (for empty-result diagnostics).
func printSearchPaths() {
	fmt.Println("Search paths:")
	for _, sp := range searchPaths() {
		exists := ""
		if _, err := os.Stat(sp.dir); err != nil {
			exists = " (not found)"
		}
		fmt.Printf("  [%s] %s%s\n", sp.label, sp.dir, exists)
	}
	fmt.Println()
	if runtime.GOOS == "windows" {
		fmt.Println("Tip: Set PORTUNIX_WIZARD_PATH=C:\\path1;C:\\path2 to add custom directories.")
	} else {
		fmt.Println("Tip: Set PORTUNIX_WIZARD_PATH=/path1:/path2 to add custom directories.")
	}
}

// readWizardDescription extracts only the description field from a wizard YAML
// without fully unmarshaling the page tree (cheap for list display).
func readWizardDescription(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	var meta struct {
		Wizard struct {
			Description string `yaml:"description"`
		} `yaml:"wizard"`
	}
	if err := yaml.Unmarshal(data, &meta); err != nil {
		return ""
	}
	return meta.Wizard.Description
}

// resolveWizardPath turns a user-supplied wizard reference into a file path.
// Used by both `run` and `validate`.
func resolveWizardPath(input string) (string, error) {
	if strings.Contains(input, "/") || strings.Contains(input, "\\") || strings.HasSuffix(input, ".yaml") {
		if _, err := os.Stat(input); err != nil {
			return "", fmt.Errorf("wizard file not found: %s", input)
		}
		return input, nil
	}

	for _, sp := range searchPaths() {
		candidate := filepath.Join(sp.dir, input+".yaml")
		if _, err := os.Stat(candidate); err == nil {
			return candidate, nil
		}
	}

	return "", fmt.Errorf("wizard not found: %s. Use 'portunix wizard list' to see available wizards", input)
}
