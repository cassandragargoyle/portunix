/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Graph data contract shared with the graphlens plugin (portunix-plugins
// ADR-022): a single JSON document with meta + nodes + links. The viewer is
// domain-agnostic; the PFT-specific mapping (folder layout, frontmatter fields,
// ISO 16355 prefixes) lives here.

// GraphNode is a single node in the graphlens contract.
type GraphNode struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Voice       string                 `json:"voice,omitempty"`
	Label       string                 `json:"label"`
	Title       string                 `json:"title,omitempty"`
	Description string                 `json:"description,omitempty"`
	Meta        map[string]interface{} `json:"meta,omitempty"`
}

// GraphLink is a single edge in the graphlens contract.
type GraphLink struct {
	Source      string `json:"source"`
	Target      string `json:"target"`
	Rel         string `json:"rel,omitempty"`
	Description string `json:"description,omitempty"`
}

// GraphDoc is the full graphlens document (meta + nodes + links).
type GraphDoc struct {
	Meta  map[string]interface{} `json:"meta"`
	Nodes []GraphNode            `json:"nodes"`
	Links []GraphLink            `json:"links"`
}

// graphVoiceDirs lists the voice areas to scan (resolved via getVoiceDir, which
// handles QFD PascalCase and basic lowercase variants).
var graphVoiceDirs = []string{"voc", "vob", "vos", "voe"}

// areaToVoice maps the frontmatter area value to the canonical voice label.
var areaToVoice = map[string]string{"voc": "VoC", "vob": "VoB", "vos": "VoS", "voe": "VoE"}

// prefixToVoice derives the voice label from the id prefix (VC-…, VB-…, …) when
// the area field is absent.
var prefixToVoice = map[string]string{"VC": "VoC", "VB": "VoB", "VS": "VoS", "VE": "VoE"}

// graphDescriptionMax caps the detail-panel description length.
const graphDescriptionMax = 320

// graphMetaExclude lists frontmatter keys not surfaced as node meta (shown
// elsewhere or internal).
var graphMetaExclude = map[string]bool{"id": true, "title": true, "publication": true}

// graphHubs enumerates the supported hub dimensions.
var graphHubs = map[string]bool{"tags": true, "author": true, "domains": true}

// relatedFields lists frontmatter fields carrying cross-references to other
// voices (parent/child, sources, addressed needs, conflicts, related).
var relatedFields = []string{"related", "parent", "parents", "children", "child", "sources", "source", "addresses", "conflicts"}

var (
	wikilinkRe   = regexp.MustCompile(`\[\[([^\]]+)\]\]`)
	paraSplitRe  = regexp.MustCompile(`\n\s*\n`)
	whitespaceRe = regexp.MustCompile(`\s+`)
)

// BuildPFTGraph builds the graphlens graph document from a PFT project directory.
// hub selects what the central hub nodes represent: "tags" (default), "author",
// or "domains". A single malformed-frontmatter file is skipped with a logged
// warning; the build never aborts on it.
func BuildPFTGraph(projectDir, hub string) (*GraphDoc, error) {
	if hub == "" {
		hub = "tags"
	}
	if !graphHubs[hub] {
		return nil, fmt.Errorf("unknown hub %q; expected tags, author, or domains", hub)
	}

	info, err := os.Stat(projectDir)
	if err != nil || !info.IsDir() {
		return nil, fmt.Errorf("PFT project directory not found: %s", projectDir)
	}

	var nodes []GraphNode
	var links []GraphLink
	hubIDs := map[string]bool{}
	voiceIDs := map[string]bool{}
	relatedMap := map[string][]string{}
	skipped := 0

	for _, voice := range graphVoiceDirs {
		base := getVoiceDir(projectDir, voice)
		for _, path := range collectVoiceFiles(base) {
			meta, body, err := parseGraphFrontmatter(path)
			if err != nil {
				// Skip a single file with malformed frontmatter; do not abort
				fmt.Fprintf(os.Stderr, "[graph] skipping %s (frontmatter error): %v\n", path, err)
				skipped++
				continue
			}

			id := asString(meta["id"])
			if id == "" {
				continue
			}
			voiceIDs[id] = true
			relatedMap[id] = collectRelated(meta, body)

			nodes = append(nodes, GraphNode{
				ID:          id,
				Type:        "voice",
				Voice:       voiceLabel(meta, id),
				Label:       id,
				Title:       firstNonEmpty(asString(meta["title"]), id),
				Description: firstParagraph(body),
				Meta:        nodeMeta(meta),
			})

			switch hub {
			case "tags":
				for _, tag := range asStringSlice(meta["tags"]) {
					hubID := "tag:" + tag
					if !hubIDs[hubID] {
						hubIDs[hubID] = true
						nodes = append(nodes, GraphNode{ID: hubID, Type: "tag", Label: tag})
					}
					links = append(links, GraphLink{Source: hubID, Target: id,
						Description: fmt.Sprintf("%s is tagged %q", id, tag)})
				}
			case "author":
				if author := asString(meta["author"]); author != "" {
					// Hub nodes use the viewer's generic hub type ("tag") so the
					// unchanged, domain-agnostic viewer renders them as hubs.
					hubID := "author:" + author
					if !hubIDs[hubID] {
						hubIDs[hubID] = true
						nodes = append(nodes, GraphNode{ID: hubID, Type: "tag", Label: author})
					}
					links = append(links, GraphLink{Source: hubID, Target: id,
						Description: fmt.Sprintf("%s proposed by %s", id, author)})
				}
			case "domains":
				for _, domain := range asStringSlice(meta["domains"]) {
					hubID := "domain:" + domain
					if !hubIDs[hubID] {
						hubIDs[hubID] = true
						nodes = append(nodes, GraphNode{ID: hubID, Type: "tag", Label: domain})
					}
					links = append(links, GraphLink{Source: hubID, Target: id,
						Description: fmt.Sprintf("%s in domain %s", id, domain)})
				}
			}
		}
	}

	// Voice-to-voice cross-reference links, independent of the hub dimension.
	// Deduplicated as undirected pairs and tagged rel="related".
	seenPairs := map[string]bool{}
	relatedCount := 0
	for _, src := range sortedKeys(relatedMap) {
		for _, rel := range relatedMap[src] {
			if rel == src || !voiceIDs[rel] {
				continue
			}
			a, b := src, rel
			if a > b {
				a, b = b, a
			}
			pair := a + "\x00" + b
			if seenPairs[pair] {
				continue
			}
			seenPairs[pair] = true
			links = append(links, GraphLink{Source: src, Target: rel, Rel: "related",
				Description: fmt.Sprintf("%s related to %s", src, rel)})
			relatedCount++
		}
	}

	voiceCount := 0
	for _, n := range nodes {
		if n.Type == "voice" {
			voiceCount++
		}
	}

	metaBlock := map[string]interface{}{
		"title":   fmt.Sprintf("PFT — Voices network (hub: %s)", hub),
		"hub":     hub,
		"voices":  voiceCount,
		"hubs":    len(hubIDs),
		"related": relatedCount,
	}
	if skipped > 0 {
		metaBlock["skipped"] = skipped
		fmt.Fprintf(os.Stderr, "[graph] %d file(s) skipped due to frontmatter errors\n", skipped)
	}

	return &GraphDoc{Meta: metaBlock, Nodes: nodes, Links: links}, nil
}

// WriteGraphJSON writes the graph document as UTF-8 JSON (no BOM) to outPath.
func WriteGraphJSON(doc *GraphDoc, outPath string) error {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(doc); err != nil {
		return err
	}
	return os.WriteFile(outPath, buf.Bytes(), 0644)
}

// collectVoiceFiles walks base recursively and returns sorted *.md paths,
// skipping README files. Missing directories yield an empty slice.
func collectVoiceFiles(base string) []string {
	var files []string
	if info, err := os.Stat(base); err != nil || !info.IsDir() {
		return files
	}
	_ = filepath.WalkDir(base, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		name := strings.ToLower(d.Name())
		if !strings.HasSuffix(name, ".md") || strings.HasPrefix(name, "readme") {
			return nil
		}
		files = append(files, path)
		return nil
	})
	sort.Strings(files)
	return files
}

// parseGraphFrontmatter returns (frontmatter, body) for a Markdown file. The
// frontmatter is parsed as YAML; a file without frontmatter yields an empty map
// and the full content as body. A YAML error is returned so the caller can skip
// the file.
func parseGraphFrontmatter(path string) (map[string]interface{}, string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, "", err
	}
	text := strings.TrimPrefix(string(content), "\uFEFF")
	meta := map[string]interface{}{}

	lines := strings.Split(text, "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return meta, text, nil
	}
	end := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			end = i
			break
		}
	}
	if end == -1 {
		// No closing delimiter — treat as plain content, no frontmatter
		return meta, text, nil
	}

	fmBlock := strings.Join(lines[1:end], "\n")
	body := strings.Join(lines[end+1:], "\n")
	if err := yaml.Unmarshal([]byte(fmBlock), &meta); err != nil {
		return nil, "", err
	}
	if meta == nil {
		meta = map[string]interface{}{}
	}
	return meta, body, nil
}

// voiceLabel derives the VoC/VoB/VoS/VoE label from the area field, falling back
// to the id prefix.
func voiceLabel(meta map[string]interface{}, id string) string {
	if v, ok := areaToVoice[strings.ToLower(asString(meta["area"]))]; ok {
		return v
	}
	if len(id) >= 2 {
		if v, ok := prefixToVoice[strings.ToUpper(id[:2])]; ok {
			return v
		}
	}
	return ""
}

// firstParagraph extracts the first real content paragraph, stripping headings
// and blockquote markers, capped at graphDescriptionMax characters.
func firstParagraph(body string) string {
	for _, para := range paraSplitRe.Split(strings.TrimSpace(body), -1) {
		rawLines := strings.Split(para, "\n")

		headingOnly := true
		for _, ln := range rawLines {
			t := strings.TrimSpace(ln)
			if t != "" && !strings.HasPrefix(t, "#") {
				headingOnly = false
				break
			}
		}
		if headingOnly {
			continue
		}

		var cleaned []string
		for _, ln := range rawLines {
			t := strings.TrimSpace(ln)
			if strings.HasPrefix(t, "#") {
				continue
			}
			t = strings.TrimSpace(strings.TrimLeft(t, "> "))
			if t != "" {
				cleaned = append(cleaned, t)
			}
		}
		text := whitespaceRe.ReplaceAllString(strings.Join(cleaned, " "), " ")
		text = strings.Trim(text, ` "„“`)
		if text == "" {
			continue
		}
		if runes := []rune(text); len(runes) > graphDescriptionMax {
			text = strings.TrimRight(string(runes[:graphDescriptionMax]), " ") + "…"
		}
		return text
	}
	return ""
}

// nodeMeta selects frontmatter fields to surface in the detail panel: drops
// internal/duplicated keys, empty values, and nested maps; keeps scalars and
// lists. Returns nil when nothing is left.
func nodeMeta(meta map[string]interface{}) map[string]interface{} {
	out := map[string]interface{}{}
	for k, v := range meta {
		if graphMetaExclude[k] || v == nil {
			continue
		}
		switch t := v.(type) {
		case string:
			if strings.TrimSpace(t) == "" {
				continue
			}
		case []interface{}:
			if len(t) == 0 {
				continue
			}
		case map[string]interface{}:
			continue
		}
		out[k] = v
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// collectRelated gathers referenced voice ids from the cross-reference
// frontmatter fields and from [[wikilink]] references in the body, normalized
// and deduplicated.
func collectRelated(meta map[string]interface{}, body string) []string {
	seen := map[string]bool{}
	var out []string
	add := func(raw string) {
		s := normalizeWikilink(raw)
		if s == "" || seen[s] {
			return
		}
		seen[s] = true
		out = append(out, s)
	}
	for _, field := range relatedFields {
		for _, v := range asStringSlice(meta[field]) {
			add(v)
		}
	}
	for _, m := range wikilinkRe.FindAllStringSubmatch(body, -1) {
		add(m[1])
	}
	return out
}

// normalizeWikilink strips [[ ]] brackets and any |alias / #section suffix.
func normalizeWikilink(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimSuffix(strings.TrimPrefix(s, "[["), "]]")
	if i := strings.IndexAny(s, "|#"); i >= 0 {
		s = s[:i]
	}
	return strings.TrimSpace(s)
}

// asString renders a scalar frontmatter value as a string.
func asString(v interface{}) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return t
	case int:
		return strconv.Itoa(t)
	case int64:
		return strconv.FormatInt(t, 10)
	case float64:
		return strconv.FormatFloat(t, 'g', -1, 64)
	case bool:
		return strconv.FormatBool(t)
	default:
		return fmt.Sprintf("%v", t)
	}
}

// asStringSlice renders a frontmatter value as a slice of trimmed non-empty
// strings, accepting a scalar or a list.
func asStringSlice(v interface{}) []string {
	var out []string
	appendStr := func(s string) {
		if s = strings.TrimSpace(s); s != "" {
			out = append(out, s)
		}
	}
	switch t := v.(type) {
	case []interface{}:
		for _, e := range t {
			appendStr(asString(e))
		}
	case []string:
		for _, e := range t {
			appendStr(e)
		}
	case nil:
		// nothing
	default:
		appendStr(asString(t))
	}
	return out
}

// firstNonEmpty returns a if non-blank, otherwise b.
func firstNonEmpty(a, b string) string {
	if strings.TrimSpace(a) != "" {
		return a
	}
	return b
}

// sortedKeys returns the map keys sorted for deterministic iteration.
func sortedKeys(m map[string][]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// graphlensAvailable probes whether the graphlens plugin is installed by
// invoking `portunix graphlens --version`. The probe is bounded by a timeout
// and detached from stdin so a missing/misbehaving plugin cannot hang the build.
func graphlensAvailable() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, portunixBinary(), "graphlens", "--version")
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run() == nil
}

// runGraphlensView hands the built graph JSON to the graphlens plugin viewer.
func runGraphlensView(file string, port int, noOpen bool) error {
	args := []string{"graphlens", "view", file}
	if port > 0 {
		args = append(args, "--port", strconv.Itoa(port))
	}
	if noOpen {
		args = append(args, "--no-open")
	}
	cmd := exec.Command(portunixBinary(), args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// portunixBinary resolves the main portunix binary: prefer the sibling next to
// this helper, fall back to PATH lookup.
func portunixBinary() string {
	if exe, err := os.Executable(); err == nil {
		candidate := filepath.Join(filepath.Dir(exe), "portunix")
		if strings.Contains(strings.ToLower(os.Getenv("OS")), "windows") {
			candidate += ".exe"
		}
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	return "portunix"
}
