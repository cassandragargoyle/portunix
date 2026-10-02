/*
 *  This file is part of CassandraGargoyle Community Project
 *  Licensed under the MIT License - see LICENSE file for details
 */
package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// readInto unmarshals the JSON file at path into v, failing the test on error.
func readInto(t *testing.T, path string, v interface{}) {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if err := json.Unmarshal(b, v); err != nil {
		t.Fatalf("unmarshal %s: %v", path, err)
	}
}

// mustExist fails the test if path is not present.
func mustExist(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected file %s to exist: %v", path, err)
	}
}

// authorVenture drives the full authoring surface end-to-end (acceptance
// criteria #1-#8) and returns the venture root directory.
func authorVenture(t *testing.T) string {
	t.Helper()
	parent := t.TempDir()
	root := filepath.Join(parent, "ai-in-HR.venture")
	p := func(extra ...string) []string { return append([]string{"--path", root}, extra...) }

	// #1 container
	handleNewVenture([]string{"ai-in-HR", "--path", parent})
	// registries
	handleAddProduct(append([]string{"pilot"}, p("--name", "Pilot")...))
	handleAddTeam(append([]string{"platform"}, p("--name", "Platform", "--kind", "team")...))
	handleAddInitiative(append([]string{"ai-in-hr"}, p("--title", "AI in HR")...))
	// #2 idea
	handleAddIdea(append([]string{"ocr-cv"}, p("--title", "OCR of CVs")...))
	handleComplexity(append([]string{"ocr-cv"}, p("--set", "m")...))
	handleInitiativeLink(append([]string{"ai-in-hr"}, p("--idea", "ocr-cv")...))
	// #3 use-case
	handleAddUseCase(append([]string{"reco-ocr"}, p("--title", "Recognise CV via OCR", "--product", "pilot", "--implements", "ocr-cv")...))
	// #4 estimate
	handleEstimate(append([]string{"reco-ocr"}, p("--by", "Zdenek", "--sp", "8", "--why", "retrieval plus mapping")...))
	// #5 delivery backlog + epic + link
	handleNewBacklog(append([]string{"hr-delivery"}, p("--kind", "delivery", "--team", "platform")...))
	handleAddEpic(append([]string{"hr-ai"}, p("--backlog", "hr-delivery", "--title", "HR AI", "--initiative", "ai-in-hr")...))
	handleEpicLink(append([]string{"hr-ai"}, p("--usecase", "reco-ocr")...))

	return root
}

func TestIdeas_NewVenture(t *testing.T) {
	parent := t.TempDir()
	handleNewVenture([]string{"ai-in-HR", "--path", parent})
	root := filepath.Join(parent, "ai-in-HR.venture")

	mustExist(t, venturePath(root))
	var v Venture
	readInto(t, venturePath(root), &v)
	if v.Slug != "ai-in-HR" || v.ID != "ai-in-HR" || v.SchemaVersion != contentSchemaVersion {
		t.Fatalf("unexpected venture: %+v", v)
	}

	// #1 an (empty) discovery backlog.
	bkPath := backlogPath(root, "ai-in-HR")
	mustExist(t, bkPath)
	var bk Backlog
	readInto(t, bkPath, &bk)
	if bk.Kind != "discovery" || len(bk.MemberRefs) != 0 {
		t.Fatalf("expected empty discovery backlog, got %+v", bk)
	}
}

func TestIdeas_IdeaHasComplexityNotStoryPoints(t *testing.T) {
	root := authorVenture(t)
	path := opportunityPath(root, "ocr-cv")
	mustExist(t, path)

	raw, _ := os.ReadFile(path)
	if strings.Contains(string(raw), "storyPoints") {
		t.Fatalf("idea JSON must not carry storyPoints:\n%s", raw)
	}
	if !strings.Contains(string(raw), "complexity") {
		t.Fatalf("idea JSON must carry complexity:\n%s", raw)
	}

	var o Opportunity
	readInto(t, path, &o)
	if o.Complexity == nil || *o.Complexity != "m" {
		t.Fatalf("expected complexity m, got %v", o.Complexity)
	}

	// #2 the idea is registered as a node in the discovery backlog.
	var bk Backlog
	readInto(t, backlogPath(root, "ai-in-HR"), &bk)
	if len(bk.MemberRefs) != 1 || bk.MemberRefs[0] != "ocr-cv" {
		t.Fatalf("expected idea registered in discovery backlog, got %v", bk.MemberRefs)
	}
}

func TestIdeas_UseCaseImplementsIdea(t *testing.T) {
	root := authorVenture(t)
	path := usecasePath(root, "reco-ocr")
	mustExist(t, path)

	var u UseCase
	readInto(t, path, &u)
	if u.ProductRef != "pilot" {
		t.Fatalf("expected productRef pilot, got %q", u.ProductRef)
	}
	if len(u.ImplementsIdeaRefs) != 1 || u.ImplementsIdeaRefs[0] != "ocr-cv" {
		t.Fatalf("expected implementsIdeaRefs=[ocr-cv], got %v", u.ImplementsIdeaRefs)
	}
}

func TestIdeas_EstimateRollsUp(t *testing.T) {
	root := authorVenture(t)

	estPath := estimationPath(root, "reco-ocr")
	mustExist(t, estPath)
	var est Estimation
	readInto(t, estPath, &est)
	if est.SubjectRef != "reco-ocr" || len(est.Estimations) != 1 || est.Estimations[0].StoryPoints != 8 {
		t.Fatalf("unexpected estimation: %+v", est)
	}

	var u UseCase
	readInto(t, usecasePath(root, "reco-ocr"), &u)
	if u.StoryPoints == nil || *u.StoryPoints != 8 {
		t.Fatalf("expected use-case storyPoints=8, got %v", u.StoryPoints)
	}
}

func TestIdeas_DeliveryBacklogEpic(t *testing.T) {
	root := authorVenture(t)

	var e Epic
	readInto(t, epicPath(root, "hr-ai"), &e)
	if e.BacklogRef != "hr-delivery" {
		t.Fatalf("expected backlogRef hr-delivery, got %q", e.BacklogRef)
	}
	if len(e.InitiativeRefs) != 1 || e.InitiativeRefs[0] != "ai-in-hr" {
		t.Fatalf("expected initiativeRefs=[ai-in-hr], got %v", e.InitiativeRefs)
	}
	if len(e.UseCaseRefs) != 1 || e.UseCaseRefs[0] != "reco-ocr" {
		t.Fatalf("expected useCaseRefs=[reco-ocr], got %v", e.UseCaseRefs)
	}

	var bk Backlog
	readInto(t, backlogPath(root, "hr-delivery"), &bk)
	if bk.Kind != "delivery" || bk.TeamRef == nil || *bk.TeamRef != "platform" {
		t.Fatalf("unexpected delivery backlog: %+v", bk)
	}
	found := false
	for _, m := range bk.MemberRefs {
		if m == "hr-ai" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected epic hr-ai in backlog members, got %v", bk.MemberRefs)
	}
}

func TestIdeas_DiscoveryGraphHasImplementsEdge(t *testing.T) {
	root := authorVenture(t)
	idx, err := loadVentureIndex(root)
	if err != nil {
		t.Fatalf("index: %v", err)
	}
	doc := buildDiscoveryGraph(idx)

	if !hasNode(doc, "ocr-cv") {
		t.Fatalf("expected idea node ocr-cv, nodes=%v", doc.Nodes)
	}
	if !hasLink(doc, "reco-ocr", "ocr-cv", "implements") {
		t.Fatalf("expected implements edge reco-ocr->ocr-cv, links=%v", doc.Links)
	}
}

func TestIdeas_BacklogGraphMatchesShape(t *testing.T) {
	root := authorVenture(t)
	idx, err := loadVentureIndex(root)
	if err != nil {
		t.Fatalf("index: %v", err)
	}
	var bk Backlog
	readInto(t, backlogPath(root, "hr-delivery"), &bk)
	doc := buildBacklogGraph(idx, &bk)

	if !hasLink(doc, "hr-ai", "reco-ocr", "contains") {
		t.Fatalf("expected contains edge hr-ai->reco-ocr, links=%v", doc.Links)
	}
	if !hasLink(doc, "reco-ocr", "ocr-cv", "implements") {
		t.Fatalf("expected implements edge reco-ocr->ocr-cv, links=%v", doc.Links)
	}
}

func TestIdeas_RollupInitiative(t *testing.T) {
	root := authorVenture(t)
	idx, err := loadVentureIndex(root)
	if err != nil {
		t.Fatalf("index: %v", err)
	}
	if sp := initiativeStoryPoints(idx, "ai-in-hr"); sp != 8 {
		t.Fatalf("expected initiative Σ storyPoints=8, got %v", sp)
	}
	counts, weighted := complexityBreakdown(idx, idx.inits["ai-in-hr"])
	if counts["m"] != 1 || weighted != 3 {
		t.Fatalf("expected complexity m:1 weighted 3, got %v / %v", counts, weighted)
	}
}

func TestIdeas_AllWrittenJSONValidates(t *testing.T) {
	root := authorVenture(t)

	var v Venture
	readInto(t, venturePath(root), &v)
	assertValid(t, &v)

	for _, o := range mustLoadOpportunities(t, root) {
		assertValid(t, o)
	}
	for _, u := range mustLoadUseCases(t, root) {
		assertValid(t, u)
	}
	for _, e := range mustLoadEpics(t, root) {
		assertValid(t, e)
	}
	for _, b := range mustLoadBacklogs(t, root) {
		assertValid(t, b)
	}
	for _, i := range mustLoadInitiatives(t, root) {
		assertValid(t, i)
	}
}

func TestIdeas_Migrate(t *testing.T) {
	parent := t.TempDir()
	src := filepath.Join(parent, "legacy.discovery")
	if err := os.MkdirAll(src, 0755); err != nil {
		t.Fatal(err)
	}
	// v1 discovery.json + a v1 opportunity carrying legacy estimate.storyPoints.
	if err := os.WriteFile(filepath.Join(src, "discovery.json"),
		[]byte(`{"name":"Legacy","description":"old"}`), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "foo.opportunity.json"),
		[]byte(`{"id":"foo","slug":"foo","title":"Foo","estimate":{"storyPoints":5}}`), 0644); err != nil {
		t.Fatal(err)
	}

	handleIdeasMigrate([]string{src})

	dst := filepath.Join(parent, "legacy.venture")
	mustExist(t, venturePath(dst))
	var v Venture
	readInto(t, venturePath(dst), &v)
	if v.Name != "Legacy" || v.Slug != "legacy" {
		t.Fatalf("unexpected migrated venture: %+v", v)
	}

	moved := filepath.Join(recordsDir(dst), "foo.opportunity.json")
	mustExist(t, moved)
	raw, _ := os.ReadFile(moved)
	if strings.Contains(string(raw), "estimate") {
		t.Fatalf("legacy estimate must be dropped:\n%s", raw)
	}
}

// --- validation of rejected documents ---

func TestIdeas_ValidateRejectsBadEnum(t *testing.T) {
	b := &Backlog{ID: "x", Slug: "x", Name: "X", Kind: "bogus"}
	if err := b.Validate(); err == nil {
		t.Fatalf("expected invalid backlog kind to fail validation")
	}
	u := &UseCase{ID: "x", Slug: "x", Title: "X"} // missing productRef
	if err := u.Validate(); err == nil {
		t.Fatalf("expected use-case without productRef to fail validation")
	}
	o := &Opportunity{ID: "x", Slug: "x", Title: "X", Complexity: strp("huge")}
	if err := o.Validate(); err == nil {
		t.Fatalf("expected invalid complexity to fail validation")
	}
}

// --- helpers ---

func assertValid(t *testing.T, v validator) {
	t.Helper()
	if err := v.Validate(); err != nil {
		t.Fatalf("expected valid entity, got: %v", err)
	}
}

func hasNode(doc *GraphDoc, id string) bool {
	for _, n := range doc.Nodes {
		if n.ID == id {
			return true
		}
	}
	return false
}

func hasLink(doc *GraphDoc, source, target, rel string) bool {
	for _, l := range doc.Links {
		if l.Source == source && l.Target == target && l.Rel == rel {
			return true
		}
	}
	return false
}

func mustLoadOpportunities(t *testing.T, root string) []*Opportunity {
	t.Helper()
	v, err := loadOpportunities(root)
	if err != nil {
		t.Fatal(err)
	}
	return v
}
func mustLoadUseCases(t *testing.T, root string) []*UseCase {
	t.Helper()
	v, err := loadUseCases(root)
	if err != nil {
		t.Fatal(err)
	}
	return v
}
func mustLoadEpics(t *testing.T, root string) []*Epic {
	t.Helper()
	v, err := loadEpics(root)
	if err != nil {
		t.Fatal(err)
	}
	return v
}
func mustLoadBacklogs(t *testing.T, root string) []*Backlog {
	t.Helper()
	v, err := loadBacklogs(root)
	if err != nil {
		t.Fatal(err)
	}
	return v
}
func mustLoadInitiatives(t *testing.T, root string) []*Initiative {
	t.Helper()
	v, err := loadInitiatives(root)
	if err != nil {
		t.Fatal(err)
	}
	return v
}
