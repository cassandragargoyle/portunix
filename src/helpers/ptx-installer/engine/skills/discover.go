/*
 *  SPDX-License-Identifier: MIT
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package skills

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// SkillFile is the marker file of a skill directory
const SkillFile = "SKILL.md"

// skillNamePattern restricts skill names to safe single path segments
var skillNamePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]*$`)

// skippedDirs are never descended into during discovery
var skippedDirs = map[string]bool{".git": true, "node_modules": true}

// Discover finds every directory containing SKILL.md below root. A skill
// directory is not searched further for nested skills. Warnings describe
// skipped candidates (invalid or duplicate names).
func Discover(root string) (skills []ResolvedSkill, warnings []string, err error) {
	seen := map[string]string{}
	walkErr := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}
		if path != root && skippedDirs[d.Name()] {
			return filepath.SkipDir
		}
		info, statErr := os.Lstat(filepath.Join(path, SkillFile))
		if statErr != nil || !info.Mode().IsRegular() {
			return nil
		}
		skill, perr := loadSkill(path)
		if perr != nil {
			warnings = append(warnings, perr.Error())
			return filepath.SkipDir
		}
		if prev, dup := seen[skill.Name]; dup {
			warnings = append(warnings, fmt.Sprintf("duplicate skill name %q in %s (already found in %s), skipped", skill.Name, path, prev))
			return filepath.SkipDir
		}
		seen[skill.Name] = path
		skills = append(skills, skill)
		return filepath.SkipDir
	})
	if walkErr != nil {
		return nil, warnings, fmt.Errorf("scan %s: %w", root, walkErr)
	}
	sort.Slice(skills, func(i, j int) bool { return skills[i].Name < skills[j].Name })
	return skills, warnings, nil
}

// loadSkill reads the front-matter and file list of one skill directory
func loadSkill(dir string) (ResolvedSkill, error) {
	meta, err := parseFrontMatter(filepath.Join(dir, SkillFile))
	if err != nil {
		return ResolvedSkill{}, err
	}
	name := meta["name"]
	if name == "" {
		name = filepath.Base(dir)
	}
	if !skillNamePattern.MatchString(name) {
		return ResolvedSkill{}, fmt.Errorf("skill in %s has an invalid name %q, skipped", dir, name)
	}
	files, err := listSkillFiles(dir)
	if err != nil {
		return ResolvedSkill{}, err
	}
	return ResolvedSkill{Name: name, Description: meta["description"], SourceDir: dir, Files: files}, nil
}

// listSkillFiles returns the regular files of a skill as sorted slash paths.
// Symlinks are ignored so a skill cannot reference files outside its tree.
func listSkillFiles(dir string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if path != dir && skippedDirs[d.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		if !d.Type().IsRegular() {
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		files = append(files, filepath.ToSlash(rel))
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("list files of %s: %w", dir, err)
	}
	sort.Strings(files)
	return files, nil
}

// parseFrontMatter reads the simple "key: value" pairs of a YAML front-matter
// block. Folded/literal block scalars (">" / "|") are joined into one line.
func parseFrontMatter(path string) (map[string]string, error) {
	f, err := os.Open(path) // #nosec G304 -- SKILL.md inside the discovered source tree
	if err != nil {
		return nil, err
	}
	defer f.Close()

	meta := map[string]string{}
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)
	if !scanner.Scan() || strings.TrimSpace(strings.TrimPrefix(scanner.Text(), "\ufeff")) != "---" {
		return meta, nil
	}
	var blockKey string
	var block []string
	flush := func() {
		if blockKey != "" {
			meta[blockKey] = strings.Join(block, " ")
		}
		blockKey, block = "", nil
	}
	for scanner.Scan() {
		line := strings.TrimRight(scanner.Text(), "\r")
		if strings.TrimSpace(line) == "---" {
			break
		}
		if blockKey != "" && (strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t")) {
			if t := strings.TrimSpace(line); t != "" {
				block = append(block, t)
			}
			continue
		}
		flush()
		key, value, ok := strings.Cut(line, ":")
		if !ok || strings.HasPrefix(line, " ") || strings.HasPrefix(line, "#") {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if strings.HasPrefix(value, ">") || strings.HasPrefix(value, "|") {
			blockKey = key
			continue
		}
		meta[key] = unquote(value)
	}
	flush()
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	return meta, nil
}

func unquote(v string) string {
	if len(v) >= 2 && (v[0] == '"' && v[len(v)-1] == '"' || v[0] == '\'' && v[len(v)-1] == '\'') {
		return v[1 : len(v)-1]
	}
	return v
}

// SelectSkills picks the requested skills by name; all selects everything
func SelectSkills(available []ResolvedSkill, names []string, all bool) ([]ResolvedSkill, error) {
	if all {
		return available, nil
	}
	byName := make(map[string]ResolvedSkill, len(available))
	for _, s := range available {
		byName[s.Name] = s
	}
	var selected []ResolvedSkill
	picked := map[string]bool{}
	for _, n := range names {
		s, ok := byName[n]
		if !ok {
			return nil, fmt.Errorf("skill %q not found in source (available: %s)", n, strings.Join(skillNames(available), ", "))
		}
		if !picked[n] {
			picked[n] = true
			selected = append(selected, s)
		}
	}
	return selected, nil
}

func skillNames(list []ResolvedSkill) []string {
	names := make([]string, len(list))
	for i, s := range list {
		names[i] = s.Name
	}
	return names
}
