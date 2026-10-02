/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// writeFixtureFile creates a Markdown file under dir/sub, creating dirs as needed.
func writeFixtureFile(t *testing.T, dir, rel, content string) {
	t.Helper()
	full := filepath.Join(dir, rel)
	if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
		t.Fatalf("mkdir %s: %v", full, err)
	}
	if err := os.WriteFile(full, []byte(content), 0644); err != nil {
		t.Fatalf("write %s: %v", full, err)
	}
}

// newFixtureProject builds a small PFT project with VoC/VoS voices, tags,
// author, domains, related cross-references, and one malformed file.
func newFixtureProject(t *testing.T) string {
	t.Helper()
	root := t.TempDir()

	writeFixtureFile(t, root, "VoC/verbatims/VC-V001-login.md", `---
id: VC-V001
title: Fast login
area: voc
author: Alice
tags:
  - auth
  - ux
domains:
  - security
related:
  - VC-V002
status: pending
---

# Fast login

Users want a faster login flow with fewer steps.
`)

	writeFixtureFile(t, root, "VoC/needs/VC-V002-sso.md", `---
id: VC-V002
title: Single sign-on
area: voc
author: Bob
tags:
  - auth
domains:
  - security
status: pending
---

# Single sign-on

See [[VC-V001]] for the related fast-login request.
`)

	writeFixtureFile(t, root, "VoS/needs/VS-S01-audit.md", `---
id: VS-S01
title: Audit trail
area: vos
author: Alice
tags:
  - compliance
domains:
  - governance
status: pending
---

# Audit trail

Stakeholders require an audit trail. Related to [[VC-V002]] and a missing [[VC-X999]].
`)

	// README files must be ignored
	writeFixtureFile(t, root, "VoC/README.md", "# VoC\n\nignore me\n")

	// Malformed frontmatter (unterminated / invalid YAML) must be skipped
	writeFixtureFile(t, root, "VoC/needs/VC-BAD.md", `---
id: VC-BAD
title: Broken
tags: [unclosed
---

# Broken
`)

	return root
}

func TestBuildPFTGraph_Tags(t *testing.T) {
	root := newFixtureProject(t)

	doc, err := BuildPFTGraph(root, "tags")
	if err != nil {
		t.Fatalf("BuildPFTGraph: %v", err)
	}

	// 3 valid voice nodes (VC-V001, VC-V002, VS-S01); VC-BAD skipped, README ignored
	voices := countNodesByType(doc, "voice")
	if voices != 3 {
		t.Errorf("expected 3 voice nodes, got %d", voices)
	}

	// Tag hubs: auth, ux, compliance = 3 distinct
	hubs := countNodesByType(doc, "tag")
	if hubs != 3 {
		t.Errorf("expected 3 tag hub nodes, got %d", hubs)
	}
	if got := doc.Meta["hubs"]; got != 3 {
		t.Errorf("meta.hubs = %v, want 3", got)
	}
	if got := doc.Meta["voices"]; got != 3 {
		t.Errorf("meta.voices = %v, want 3", got)
	}
	if got := doc.Meta["skipped"]; got != 1 {
		t.Errorf("meta.skipped = %v, want 1", got)
	}

	// Hub links use no rel; voice node carries voice label + description
	vc1 := findNode(doc, "VC-V001")
	if vc1 == nil {
		t.Fatal("VC-V001 node missing")
	}
	if vc1.Voice != "VoC" {
		t.Errorf("VC-V001 voice = %q, want VoC", vc1.Voice)
	}
	if vc1.Description == "" {
		t.Error("VC-V001 description should be populated from first paragraph")
	}
	if vc1.Meta["author"] != "Alice" {
		t.Errorf("VC-V001 meta.author = %v, want Alice", vc1.Meta["author"])
	}
}

func TestBuildPFTGraph_HubDimensions(t *testing.T) {
	root := newFixtureProject(t)

	cases := map[string]int{
		"author":  2, // Alice, Bob
		"domains": 2, // security, governance
	}
	for hub, wantHubs := range cases {
		doc, err := BuildPFTGraph(root, hub)
		if err != nil {
			t.Fatalf("hub %s: %v", hub, err)
		}
		// Voice count is stable across hub dimensions
		if v := countNodesByType(doc, "voice"); v != 3 {
			t.Errorf("hub %s: voice nodes = %d, want 3", hub, v)
		}
		// Hub nodes always use the generic "tag" type
		if h := countNodesByType(doc, "tag"); h != wantHubs {
			t.Errorf("hub %s: hub nodes = %d, want %d", hub, h, wantHubs)
		}
		if got := doc.Meta["hub"]; got != hub {
			t.Errorf("meta.hub = %v, want %s", got, hub)
		}
	}
}

func TestBuildPFTGraph_RelatedLinks(t *testing.T) {
	root := newFixtureProject(t)

	doc, err := BuildPFTGraph(root, "tags")
	if err != nil {
		t.Fatalf("BuildPFTGraph: %v", err)
	}

	related := relatedLinks(doc)
	// Expected undirected pairs (dedup):
	//   VC-V001 <-> VC-V002 (from related frontmatter + [[VC-V001]] wikilink) => 1
	//   VS-S01  <-> VC-V002 (from [[VC-V002]] wikilink)                       => 1
	//   [[VC-X999]] points to a non-existent node => dropped
	if len(related) != 2 {
		t.Fatalf("expected 2 related links, got %d: %+v", len(related), related)
	}
	if got := doc.Meta["related"]; got != 2 {
		t.Errorf("meta.related = %v, want 2", got)
	}

	for _, l := range related {
		if l.Rel != "related" {
			t.Errorf("related link missing rel=related: %+v", l)
		}
		if l.Target == "VC-X999" || l.Source == "VC-X999" {
			t.Errorf("link to non-existent node should be dropped: %+v", l)
		}
	}
}

func TestBuildPFTGraph_UnknownHub(t *testing.T) {
	root := newFixtureProject(t)
	if _, err := BuildPFTGraph(root, "bogus"); err == nil {
		t.Error("expected error for unknown hub, got nil")
	}
}

func TestWriteGraphJSON_NoBOM(t *testing.T) {
	root := newFixtureProject(t)
	doc, err := BuildPFTGraph(root, "tags")
	if err != nil {
		t.Fatalf("BuildPFTGraph: %v", err)
	}

	out := filepath.Join(t.TempDir(), "graph.json")
	if err := WriteGraphJSON(doc, out); err != nil {
		t.Fatalf("WriteGraphJSON: %v", err)
	}

	data, err := os.ReadFile(out)
	if err != nil {
		t.Fatalf("read output: %v", err)
	}
	if len(data) >= 3 && data[0] == 0xEF && data[1] == 0xBB && data[2] == 0xBF {
		t.Error("output JSON must not start with a UTF-8 BOM")
	}

	// Must be valid JSON matching the contract shape
	var parsed GraphDoc
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if len(parsed.Nodes) == 0 || len(parsed.Links) == 0 {
		t.Error("parsed graph should have nodes and links")
	}
}

// --- helpers ---

func countNodesByType(doc *GraphDoc, typ string) int {
	n := 0
	for _, node := range doc.Nodes {
		if node.Type == typ {
			n++
		}
	}
	return n
}

func findNode(doc *GraphDoc, id string) *GraphNode {
	for i := range doc.Nodes {
		if doc.Nodes[i].ID == id {
			return &doc.Nodes[i]
		}
	}
	return nil
}

func relatedLinks(doc *GraphDoc) []GraphLink {
	var out []GraphLink
	for _, l := range doc.Links {
		if l.Rel == "related" {
			out = append(out, l)
		}
	}
	return out
}
